package videointegration

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Bashar444/VTP/pkg/models"
	"github.com/google/uuid"
)

var (
	ErrMeetingIDRequired = errors.New("meeting ID is required")
	ErrInvalidProvider   = errors.New("invalid video provider")
)

// VideoProvider interface for video conferencing providers
type VideoProvider interface {
	CreateMeeting(ctx context.Context, title string, scheduledAt time.Time, duration int) (*MeetingDetails, error)
	UpdateMeeting(ctx context.Context, externalID string, title string, scheduledAt time.Time) error
	DeleteMeeting(ctx context.Context, externalID string) error
	GetMeetingInfo(ctx context.Context, externalID string) (*MeetingDetails, error)
}

// MeetingDetails holds provider meeting details
type MeetingDetails struct {
	ExternalID  string
	MeetingLink string
	HostLink    string
	Password    string
	JoinURL     string
}

// JitsiProvider implements VideoProvider for Jitsi Meet
type JitsiProvider struct {
	serverURL string
	logger    *log.Logger
}

// NewJitsiProvider creates a new Jitsi provider
func NewJitsiProvider(serverURL string, logger *log.Logger) *JitsiProvider {
	if serverURL == "" {
		serverURL = "https://meet.jit.si" // Public Jitsi server
	}
	return &JitsiProvider{
		serverURL: serverURL,
		logger:    logger,
	}
}

// CreateMeeting creates a Jitsi meeting (Jitsi creates rooms on-demand)
func (j *JitsiProvider) CreateMeeting(ctx context.Context, title string, scheduledAt time.Time, duration int) (*MeetingDetails, error) {
	// Jitsi uses room names - we generate a unique one
	roomID := fmt.Sprintf("vtp-%s", uuid.New().String()[:8])
	meetingLink := fmt.Sprintf("%s/%s", j.serverURL, roomID)

	return &MeetingDetails{
		ExternalID:  roomID,
		MeetingLink: meetingLink,
		HostLink:    meetingLink, // Same for Jitsi
		JoinURL:     meetingLink,
	}, nil
}

// UpdateMeeting updates a Jitsi meeting (no-op for Jitsi)
func (j *JitsiProvider) UpdateMeeting(ctx context.Context, externalID string, title string, scheduledAt time.Time) error {
	// Jitsi rooms are ephemeral, no update needed
	return nil
}

// DeleteMeeting deletes a Jitsi meeting (no-op for Jitsi)
func (j *JitsiProvider) DeleteMeeting(ctx context.Context, externalID string) error {
	// Jitsi rooms are ephemeral
	return nil
}

// GetMeetingInfo gets Jitsi meeting info
func (j *JitsiProvider) GetMeetingInfo(ctx context.Context, externalID string) (*MeetingDetails, error) {
	meetingLink := fmt.Sprintf("%s/%s", j.serverURL, externalID)
	return &MeetingDetails{
		ExternalID:  externalID,
		MeetingLink: meetingLink,
		HostLink:    meetingLink,
		JoinURL:     meetingLink,
	}, nil
}

// GoogleMeetProvider implements VideoProvider for Google Meet
// Note: Full implementation requires Google Calendar API credentials
type GoogleMeetProvider struct {
	// In production, add Google API client
	logger *log.Logger
}

// NewGoogleMeetProvider creates a new Google Meet provider
func NewGoogleMeetProvider(logger *log.Logger) *GoogleMeetProvider {
	return &GoogleMeetProvider{logger: logger}
}

// CreateMeeting creates a Google Meet meeting
func (g *GoogleMeetProvider) CreateMeeting(ctx context.Context, title string, scheduledAt time.Time, duration int) (*MeetingDetails, error) {
	// TODO: Implement with Google Calendar API
	// This requires OAuth2 setup and Google Workspace integration
	// For now, return a placeholder
	meetID := fmt.Sprintf("meet.google.com/%s", uuid.New().String()[:12])
	return &MeetingDetails{
		ExternalID:  meetID,
		MeetingLink: "https://" + meetID,
		HostLink:    "https://" + meetID,
		JoinURL:     "https://" + meetID,
	}, nil
}

// UpdateMeeting updates a Google Meet meeting
func (g *GoogleMeetProvider) UpdateMeeting(ctx context.Context, externalID string, title string, scheduledAt time.Time) error {
	// TODO: Implement with Google Calendar API
	return nil
}

// DeleteMeeting deletes a Google Meet meeting
func (g *GoogleMeetProvider) DeleteMeeting(ctx context.Context, externalID string) error {
	// TODO: Implement with Google Calendar API
	return nil
}

// GetMeetingInfo gets Google Meet info
func (g *GoogleMeetProvider) GetMeetingInfo(ctx context.Context, externalID string) (*MeetingDetails, error) {
	return &MeetingDetails{
		ExternalID:  externalID,
		MeetingLink: "https://" + externalID,
		HostLink:    "https://" + externalID,
		JoinURL:     "https://" + externalID,
	}, nil
}

// Service handles video integration business logic
type Service struct {
	repo      *Repository
	providers map[Provider]VideoProvider
	logger    *log.Logger
}

// NewService creates a new video integration service
func NewService(repo *Repository, logger *log.Logger) *Service {
	return &Service{
		repo:      repo,
		providers: make(map[Provider]VideoProvider),
		logger:    logger,
	}
}

// RegisterProvider registers a video provider
func (s *Service) RegisterProvider(provider Provider, impl VideoProvider) {
	s.providers[provider] = impl
}

// CreateMeetingIntegration creates a meeting with the specified provider
func (s *Service) CreateMeetingIntegration(ctx context.Context, meetingID string, provider Provider, title string, scheduledAt time.Time, duration int) (*models.MeetingIntegration, error) {
	if meetingID == "" {
		return nil, ErrMeetingIDRequired
	}
	if !ValidProviders[string(provider)] {
		return nil, ErrInvalidProvider
	}

	// Get provider implementation
	providerImpl, ok := s.providers[provider]
	if !ok {
		// If no provider registered, create a generic link
		return s.createGenericIntegration(ctx, meetingID, provider, title)
	}

	// Create meeting with provider
	details, err := providerImpl.CreateMeeting(ctx, title, scheduledAt, duration)
	if err != nil {
		return nil, err
	}

	mi := &models.MeetingIntegration{
		MeetingID:         meetingID,
		Provider:          string(provider),
		ExternalMeetingID: details.ExternalID,
		MeetingLink:       details.MeetingLink,
		HostLink:          details.HostLink,
		Password:          details.Password,
		Settings:          "{}",
	}

	if err := s.repo.Create(ctx, mi); err != nil {
		return nil, err
	}

	return mi, nil
}

// createGenericIntegration creates a generic meeting link (internal/fallback)
func (s *Service) createGenericIntegration(ctx context.Context, meetingID string, provider Provider, title string) (*models.MeetingIntegration, error) {
	var meetingLink, hostLink string

	switch provider {
	case ProviderJitsi:
		roomID := fmt.Sprintf("vtp-%s", uuid.New().String()[:8])
		meetingLink = fmt.Sprintf("https://meet.jit.si/%s", roomID)
		hostLink = meetingLink
	case ProviderInternal:
		meetingLink = fmt.Sprintf("/meeting/%s", meetingID)
		hostLink = meetingLink
	default:
		meetingLink = fmt.Sprintf("/meeting/%s", meetingID)
		hostLink = meetingLink
	}

	mi := &models.MeetingIntegration{
		MeetingID:         meetingID,
		Provider:          string(provider),
		ExternalMeetingID: meetingID,
		MeetingLink:       meetingLink,
		HostLink:          hostLink,
		Settings:          "{}",
	}

	if err := s.repo.Create(ctx, mi); err != nil {
		return nil, err
	}

	return mi, nil
}

// GetMeetingIntegration retrieves integration by meeting ID
func (s *Service) GetMeetingIntegration(ctx context.Context, meetingID string) (*models.MeetingIntegration, error) {
	return s.repo.GetByMeetingID(ctx, meetingID)
}

// UpdateMeetingIntegration updates a meeting integration
func (s *Service) UpdateMeetingIntegration(ctx context.Context, meetingID string, title string, scheduledAt time.Time) error {
	mi, err := s.repo.GetByMeetingID(ctx, meetingID)
	if err != nil {
		return err
	}

	provider := Provider(mi.Provider)
	providerImpl, ok := s.providers[provider]
	if ok {
		if err := providerImpl.UpdateMeeting(ctx, mi.ExternalMeetingID, title, scheduledAt); err != nil {
			s.logger.Printf("Failed to update external meeting: %v", err)
		}
	}

	return s.repo.Update(ctx, mi)
}

// DeleteMeetingIntegration removes a meeting integration
func (s *Service) DeleteMeetingIntegration(ctx context.Context, meetingID string) error {
	mi, err := s.repo.GetByMeetingID(ctx, meetingID)
	if err != nil {
		return err
	}

	provider := Provider(mi.Provider)
	providerImpl, ok := s.providers[provider]
	if ok {
		if err := providerImpl.DeleteMeeting(ctx, mi.ExternalMeetingID); err != nil {
			s.logger.Printf("Failed to delete external meeting: %v", err)
		}
	}

	return s.repo.Delete(ctx, mi.ID)
}

// GetJoinLink returns the participant join link
func (s *Service) GetJoinLink(ctx context.Context, meetingID string) (string, error) {
	mi, err := s.repo.GetByMeetingID(ctx, meetingID)
	if err != nil {
		return "", err
	}
	return mi.MeetingLink, nil
}

// GetHostLink returns the host/teacher join link
func (s *Service) GetHostLink(ctx context.Context, meetingID string) (string, error) {
	mi, err := s.repo.GetByMeetingID(ctx, meetingID)
	if err != nil {
		return "", err
	}
	if mi.HostLink != "" {
		return mi.HostLink, nil
	}
	return mi.MeetingLink, nil
}
