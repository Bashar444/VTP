package videointegration

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Bashar444/VTP/pkg/models"
	"github.com/google/uuid"
)

var (
	ErrIntegrationNotFound  = errors.New("meeting integration not found")
	ErrProviderNotSupported = errors.New("video provider not supported")
)

// Provider represents a video conferencing provider
type Provider string

const (
	ProviderGoogleMeet Provider = "google_meet"
	ProviderZoom       Provider = "zoom"
	ProviderJitsi      Provider = "jitsi"
	ProviderInternal   Provider = "internal"
)

// ValidProviders are the supported video providers
var ValidProviders = map[string]bool{
	string(ProviderGoogleMeet): true,
	string(ProviderZoom):       true,
	string(ProviderJitsi):      true,
	string(ProviderInternal):   true,
}

// Repository handles meeting integration data persistence
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts a new meeting integration
func (r *Repository) Create(ctx context.Context, mi *models.MeetingIntegration) error {
	if mi.ID == "" {
		mi.ID = uuid.New().String()
	}
	mi.CreatedAt = time.Now()
	mi.UpdatedAt = time.Now()

	query := `
		INSERT INTO meeting_integrations (
			id, meeting_id, provider, external_meeting_id, meeting_link,
			host_link, password, settings, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		mi.ID, mi.MeetingID, mi.Provider, mi.ExternalMeetingID, mi.MeetingLink,
		mi.HostLink, mi.Password, mi.Settings, mi.CreatedAt, mi.UpdatedAt,
	)
	return err
}

// GetByMeetingID retrieves integration by meeting ID
func (r *Repository) GetByMeetingID(ctx context.Context, meetingID string) (*models.MeetingIntegration, error) {
	query := `
		SELECT id, meeting_id, provider, external_meeting_id, meeting_link,
			host_link, password, settings, created_at, updated_at
		FROM meeting_integrations WHERE meeting_id = $1
	`
	var mi models.MeetingIntegration
	err := r.db.QueryRowContext(ctx, query, meetingID).Scan(
		&mi.ID, &mi.MeetingID, &mi.Provider, &mi.ExternalMeetingID, &mi.MeetingLink,
		&mi.HostLink, &mi.Password, &mi.Settings, &mi.CreatedAt, &mi.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrIntegrationNotFound
	}
	return &mi, err
}

// Update modifies an existing integration
func (r *Repository) Update(ctx context.Context, mi *models.MeetingIntegration) error {
	mi.UpdatedAt = time.Now()
	query := `
		UPDATE meeting_integrations SET
			meeting_link = $2, host_link = $3, password = $4,
			settings = $5, updated_at = $6
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		mi.ID, mi.MeetingLink, mi.HostLink, mi.Password, mi.Settings, mi.UpdatedAt,
	)
	return err
}

// Delete removes an integration
func (r *Repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM meeting_integrations WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
