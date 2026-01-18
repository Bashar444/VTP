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
	"os"
	"sync"
	"time"
)

var (
	ErrZoomCredentialsMissing = errors.New("Zoom S2S OAuth credentials not configured")
	ErrZoomAPIError           = errors.New("Zoom API error")
	ErrZoomTokenRefresh       = errors.New("failed to get Zoom access token")
)

// ZoomConfig holds Zoom Server-to-Server OAuth configuration
type ZoomConfig struct {
	AccountID    string
	ClientID     string
	ClientSecret string
}

// ZoomProvider implements VideoProvider for Zoom with Server-to-Server OAuth
type ZoomProvider struct {
	config      ZoomConfig
	accessToken string
	tokenExpiry time.Time
	tokenMutex  sync.RWMutex
	httpClient  *http.Client
	logger      *log.Logger
}

// Zoom OAuth token response
type zoomTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

// Zoom meeting request
type zoomMeetingRequest struct {
	Topic     string              `json:"topic"`
	Type      int                 `json:"type"` // 2 = scheduled
	StartTime string              `json:"start_time"`
	Duration  int                 `json:"duration"`
	Timezone  string              `json:"timezone"`
	Settings  zoomMeetingSettings `json:"settings"`
}

type zoomMeetingSettings struct {
	HostVideo        bool   `json:"host_video"`
	ParticipantVideo bool   `json:"participant_video"`
	JoinBeforeHost   bool   `json:"join_before_host"`
	MuteUponEntry    bool   `json:"mute_upon_entry"`
	WaitingRoom      bool   `json:"waiting_room"`
	ApprovalType     int    `json:"approval_type"` // 2 = no registration required
	Audio            string `json:"audio"`         // "both", "telephony", "voip"
}

// Zoom meeting response
type zoomMeetingResponse struct {
	ID           int64  `json:"id"`
	UUID         string `json:"uuid"`
	HostID       string `json:"host_id"`
	Topic        string `json:"topic"`
	Type         int    `json:"type"`
	StartTime    string `json:"start_time"`
	Duration     int    `json:"duration"`
	Timezone     string `json:"timezone"`
	JoinURL      string `json:"join_url"`
	StartURL     string `json:"start_url"`
	Password     string `json:"password"`
	H323Password string `json:"h323_password"`
}

// Zoom user response
type zoomUserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// NewZoomProvider creates a new Zoom provider with Server-to-Server OAuth
func NewZoomProvider(logger *log.Logger) (*ZoomProvider, error) {
	accountID := os.Getenv("ZOOM_ACCOUNT_ID")
	clientID := os.Getenv("ZOOM_CLIENT_ID")
	clientSecret := os.Getenv("ZOOM_CLIENT_SECRET")

	if accountID == "" || clientID == "" || clientSecret == "" {
		return nil, ErrZoomCredentialsMissing
	}

	provider := &ZoomProvider{
		config: ZoomConfig{
			AccountID:    accountID,
			ClientID:     clientID,
			ClientSecret: clientSecret,
		},
		httpClient: &http.Client{Timeout: 30 * time.Second},
		logger:     logger,
	}

	// Get initial access token
	if err := provider.getNewAccessToken(); err != nil {
		return nil, fmt.Errorf("failed to get initial Zoom access token: %w", err)
	}

	return provider, nil
}

// getNewAccessToken gets a new access token using Server-to-Server OAuth
func (z *ZoomProvider) getNewAccessToken() error {
	tokenURL := fmt.Sprintf("https://zoom.us/oauth/token?grant_type=account_credentials&account_id=%s", z.config.AccountID)

	req, err := http.NewRequest("POST", tokenURL, nil)
	if err != nil {
		return err
	}

	// S2S OAuth uses Basic auth with client credentials
	req.SetBasicAuth(z.config.ClientID, z.config.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := z.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		z.logger.Printf("Zoom token request failed: %s", string(body))
		return ErrZoomTokenRefresh
	}

	var tokenResp zoomTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	z.tokenMutex.Lock()
	z.accessToken = tokenResp.AccessToken
	z.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second) // Refresh 1 min early
	z.tokenMutex.Unlock()

	z.logger.Printf("Zoom access token obtained, expires in %d seconds", tokenResp.ExpiresIn)
	return nil
}

// ensureValidToken checks if token is valid and refreshes if needed
func (z *ZoomProvider) ensureValidToken() error {
	z.tokenMutex.RLock()
	expired := time.Now().After(z.tokenExpiry)
	z.tokenMutex.RUnlock()

	if expired {
		return z.getNewAccessToken()
	}
	return nil
}

// getAccessToken returns current access token, refreshing if needed
func (z *ZoomProvider) getAccessToken() (string, error) {
	if err := z.ensureValidToken(); err != nil {
		return "", err
	}
	z.tokenMutex.RLock()
	defer z.tokenMutex.RUnlock()
	return z.accessToken, nil
}

// getUserID gets the current user's ID (needed to create meetings)
func (z *ZoomProvider) getUserID(ctx context.Context) (string, error) {
	token, err := z.getAccessToken()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.zoom.us/v2/users/me", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := z.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		z.logger.Printf("Zoom get user failed: %s", string(body))
		return "", ErrZoomAPIError
	}

	var user zoomUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", err
	}

	return user.ID, nil
}

// CreateMeeting creates a Zoom meeting
func (z *ZoomProvider) CreateMeeting(ctx context.Context, title string, scheduledAt time.Time, duration int) (*MeetingDetails, error) {
	token, err := z.getAccessToken()
	if err != nil {
		return nil, err
	}

	// Get user ID first
	userID, err := z.getUserID(ctx)
	if err != nil {
		// Fallback to "me" if we can't get user ID
		userID = "me"
	}

	meeting := zoomMeetingRequest{
		Topic:     title,
		Type:      2, // Scheduled meeting
		StartTime: scheduledAt.Format("2006-01-02T15:04:05Z"),
		Duration:  duration,
		Timezone:  "UTC",
		Settings: zoomMeetingSettings{
			HostVideo:        true,
			ParticipantVideo: true,
			JoinBeforeHost:   true,
			MuteUponEntry:    false,
			WaitingRoom:      false,
			ApprovalType:     2, // No registration required
			Audio:            "both",
		},
	}

	body, err := json.Marshal(meeting)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.zoom.us/v2/users/%s/meetings", userID)
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := z.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		z.logger.Printf("Zoom create meeting failed: %s", string(respBody))
		return nil, fmt.Errorf("%w: %s", ErrZoomAPIError, string(respBody))
	}

	var meetingResp zoomMeetingResponse
	if err := json.NewDecoder(resp.Body).Decode(&meetingResp); err != nil {
		return nil, err
	}

	z.logger.Printf("Created Zoom meeting: %d - %s", meetingResp.ID, meetingResp.JoinURL)

	return &MeetingDetails{
		ExternalID:  fmt.Sprintf("%d", meetingResp.ID),
		MeetingLink: meetingResp.JoinURL,
		HostLink:    meetingResp.StartURL,
		Password:    meetingResp.Password,
		JoinURL:     meetingResp.JoinURL,
	}, nil
}

// UpdateMeeting updates a Zoom meeting
func (z *ZoomProvider) UpdateMeeting(ctx context.Context, externalID string, title string, scheduledAt time.Time) error {
	token, err := z.getAccessToken()
	if err != nil {
		return err
	}

	update := map[string]interface{}{
		"topic":      title,
		"start_time": scheduledAt.Format("2006-01-02T15:04:05Z"),
	}

	body, err := json.Marshal(update)
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://api.zoom.us/v2/meetings/%s", externalID)
	req, err := http.NewRequestWithContext(ctx, "PATCH", apiURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := z.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 204 No Content is success for PATCH
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		z.logger.Printf("Zoom update meeting failed: %s", string(respBody))
		return ErrZoomAPIError
	}

	return nil
}

// DeleteMeeting deletes a Zoom meeting
func (z *ZoomProvider) DeleteMeeting(ctx context.Context, externalID string) error {
	token, err := z.getAccessToken()
	if err != nil {
		return err
	}

	apiURL := fmt.Sprintf("https://api.zoom.us/v2/meetings/%s", externalID)
	req, err := http.NewRequestWithContext(ctx, "DELETE", apiURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := z.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 204 No Content is success for DELETE
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return ErrZoomAPIError
	}

	return nil
}

// GetMeetingInfo gets info about a Zoom meeting
func (z *ZoomProvider) GetMeetingInfo(ctx context.Context, externalID string) (*MeetingDetails, error) {
	token, err := z.getAccessToken()
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.zoom.us/v2/meetings/%s", externalID)
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := z.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrZoomAPIError
	}

	var meetingResp zoomMeetingResponse
	if err := json.NewDecoder(resp.Body).Decode(&meetingResp); err != nil {
		return nil, err
	}

	return &MeetingDetails{
		ExternalID:  fmt.Sprintf("%d", meetingResp.ID),
		MeetingLink: meetingResp.JoinURL,
		HostLink:    meetingResp.StartURL,
		Password:    meetingResp.Password,
		JoinURL:     meetingResp.JoinURL,
	}, nil
}
