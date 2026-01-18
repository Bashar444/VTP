package videointegration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	ErrGoogleCredentialsMissing = errors.New("Google OAuth credentials not configured")
	ErrGoogleAPIError           = errors.New("Google API error")
	ErrGoogleTokenRefresh       = errors.New("failed to refresh Google access token")
)

// GoogleMeetConfig holds Google OAuth configuration
type GoogleMeetConfig struct {
	ClientID     string
	ClientSecret string
	RefreshToken string
}

// GoogleMeetProviderFull implements VideoProvider for Google Meet with real API calls
type GoogleMeetProviderFull struct {
	config      GoogleMeetConfig
	accessToken string
	tokenExpiry time.Time
	tokenMutex  sync.RWMutex
	httpClient  *http.Client
	logger      *log.Logger
}

// Google OAuth token response
type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// Google Calendar Event for creating Meet meetings
type googleCalendarEvent struct {
	Summary        string                `json:"summary"`
	Description    string                `json:"description,omitempty"`
	Start          googleEventDateTime   `json:"start"`
	End            googleEventDateTime   `json:"end"`
	ConferenceData *googleConferenceData `json:"conferenceData,omitempty"`
	Attendees      []googleAttendee      `json:"attendees,omitempty"`
}

type googleEventDateTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

type googleConferenceData struct {
	CreateRequest *googleCreateRequest `json:"createRequest,omitempty"`
	EntryPoints   []googleEntryPoint   `json:"entryPoints,omitempty"`
	ConferenceID  string               `json:"conferenceId,omitempty"`
}

type googleCreateRequest struct {
	RequestID             string                   `json:"requestId"`
	ConferenceSolutionKey googleConferenceSolution `json:"conferenceSolutionKey"`
}

type googleConferenceSolution struct {
	Type string `json:"type"`
}

type googleEntryPoint struct {
	EntryPointType string `json:"entryPointType"`
	URI            string `json:"uri"`
	Label          string `json:"label,omitempty"`
}

type googleAttendee struct {
	Email string `json:"email"`
}

type googleEventResponse struct {
	ID             string                `json:"id"`
	HtmlLink       string                `json:"htmlLink"`
	Summary        string                `json:"summary"`
	Start          googleEventDateTime   `json:"start"`
	End            googleEventDateTime   `json:"end"`
	HangoutLink    string                `json:"hangoutLink"`
	ConferenceData *googleConferenceData `json:"conferenceData,omitempty"`
}

// NewGoogleMeetProviderFull creates a new Google Meet provider with full API support
func NewGoogleMeetProviderFull(logger *log.Logger) (*GoogleMeetProviderFull, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	refreshToken := os.Getenv("GOOGLE_REFRESH_TOKEN")

	if clientID == "" || clientSecret == "" || refreshToken == "" {
		return nil, ErrGoogleCredentialsMissing
	}

	provider := &GoogleMeetProviderFull{
		config: GoogleMeetConfig{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RefreshToken: refreshToken,
		},
		httpClient: &http.Client{Timeout: 30 * time.Second},
		logger:     logger,
	}

	// Get initial access token
	if err := provider.refreshAccessToken(); err != nil {
		return nil, fmt.Errorf("failed to get initial access token: %w", err)
	}

	return provider, nil
}

// refreshAccessToken refreshes the OAuth access token using the refresh token
func (g *GoogleMeetProviderFull) refreshAccessToken() error {
	data := url.Values{}
	data.Set("client_id", g.config.ClientID)
	data.Set("client_secret", g.config.ClientSecret)
	data.Set("refresh_token", g.config.RefreshToken)
	data.Set("grant_type", "refresh_token")

	req, err := http.NewRequest("POST", "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		g.logger.Printf("Google token refresh failed: %s", string(body))
		return ErrGoogleTokenRefresh
	}

	var tokenResp googleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	g.tokenMutex.Lock()
	g.accessToken = tokenResp.AccessToken
	g.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second) // Refresh 1 min early
	g.tokenMutex.Unlock()

	g.logger.Printf("Google access token refreshed, expires in %d seconds", tokenResp.ExpiresIn)
	return nil
}

// ensureValidToken checks if token is valid and refreshes if needed
func (g *GoogleMeetProviderFull) ensureValidToken() error {
	g.tokenMutex.RLock()
	expired := time.Now().After(g.tokenExpiry)
	g.tokenMutex.RUnlock()

	if expired {
		return g.refreshAccessToken()
	}
	return nil
}

// getAccessToken returns current access token, refreshing if needed
func (g *GoogleMeetProviderFull) getAccessToken() (string, error) {
	if err := g.ensureValidToken(); err != nil {
		return "", err
	}
	g.tokenMutex.RLock()
	defer g.tokenMutex.RUnlock()
	return g.accessToken, nil
}

// CreateMeeting creates a Google Calendar event with Google Meet
func (g *GoogleMeetProviderFull) CreateMeeting(ctx context.Context, title string, scheduledAt time.Time, duration int) (*MeetingDetails, error) {
	token, err := g.getAccessToken()
	if err != nil {
		return nil, err
	}

	// Create calendar event with Meet
	endTime := scheduledAt.Add(time.Duration(duration) * time.Minute)

	event := googleCalendarEvent{
		Summary: title,
		Start: googleEventDateTime{
			DateTime: scheduledAt.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		End: googleEventDateTime{
			DateTime: endTime.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		ConferenceData: &googleConferenceData{
			CreateRequest: &googleCreateRequest{
				RequestID: fmt.Sprintf("vtp-%d", time.Now().UnixNano()),
				ConferenceSolutionKey: googleConferenceSolution{
					Type: "hangoutsMeet",
				},
			},
		},
	}

	body, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	// Create event with conferenceDataVersion=1 to enable Meet
	apiURL := "https://www.googleapis.com/calendar/v3/calendars/primary/events?conferenceDataVersion=1"
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		g.logger.Printf("Google Calendar API error: %s", string(respBody))
		return nil, fmt.Errorf("%w: %s", ErrGoogleAPIError, string(respBody))
	}

	var eventResp googleEventResponse
	if err := json.NewDecoder(resp.Body).Decode(&eventResp); err != nil {
		return nil, err
	}

	// Extract Meet link
	meetLink := eventResp.HangoutLink
	if meetLink == "" && eventResp.ConferenceData != nil {
		for _, ep := range eventResp.ConferenceData.EntryPoints {
			if ep.EntryPointType == "video" {
				meetLink = ep.URI
				break
			}
		}
	}

	g.logger.Printf("Created Google Meet: %s (Event ID: %s)", meetLink, eventResp.ID)

	return &MeetingDetails{
		ExternalID:  eventResp.ID,
		MeetingLink: meetLink,
		HostLink:    meetLink,
		JoinURL:     meetLink,
	}, nil
}

// UpdateMeeting updates a Google Calendar event
func (g *GoogleMeetProviderFull) UpdateMeeting(ctx context.Context, externalID string, title string, scheduledAt time.Time) error {
	token, err := g.getAccessToken()
	if err != nil {
		return err
	}

	// Get existing event first
	apiURL := fmt.Sprintf("https://www.googleapis.com/calendar/v3/calendars/primary/events/%s", externalID)
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ErrGoogleAPIError
	}

	var existingEvent googleEventResponse
	if err := json.NewDecoder(resp.Body).Decode(&existingEvent); err != nil {
		return err
	}

	// Update event
	updateEvent := googleCalendarEvent{
		Summary: title,
		Start: googleEventDateTime{
			DateTime: scheduledAt.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		End: existingEvent.End,
	}

	body, err := json.Marshal(updateEvent)
	if err != nil {
		return err
	}

	req, err = http.NewRequestWithContext(ctx, "PATCH", apiURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err = g.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ErrGoogleAPIError
	}

	return nil
}

// DeleteMeeting deletes a Google Calendar event
func (g *GoogleMeetProviderFull) DeleteMeeting(ctx context.Context, externalID string) error {
	token, err := g.getAccessToken()
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://www.googleapis.com/calendar/v3/calendars/primary/events/%s", externalID)
	req, err := http.NewRequestWithContext(ctx, "DELETE", apiURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 204 No Content or 410 Gone are both acceptable
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusGone && resp.StatusCode != http.StatusOK {
		return ErrGoogleAPIError
	}

	return nil
}

// GetMeetingInfo gets info about a Google Calendar event
func (g *GoogleMeetProviderFull) GetMeetingInfo(ctx context.Context, externalID string) (*MeetingDetails, error) {
	token, err := g.getAccessToken()
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://www.googleapis.com/calendar/v3/calendars/primary/events/%s", externalID)
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrGoogleAPIError
	}

	var eventResp googleEventResponse
	if err := json.NewDecoder(resp.Body).Decode(&eventResp); err != nil {
		return nil, err
	}

	meetLink := eventResp.HangoutLink
	if meetLink == "" && eventResp.ConferenceData != nil {
		for _, ep := range eventResp.ConferenceData.EntryPoints {
			if ep.EntryPointType == "video" {
				meetLink = ep.URI
				break
			}
		}
	}

	return &MeetingDetails{
		ExternalID:  eventResp.ID,
		MeetingLink: meetLink,
		HostLink:    meetLink,
		JoinURL:     meetLink,
	}, nil
}
