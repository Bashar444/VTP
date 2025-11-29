package recording

import (
	"time"

	"github.com/google/uuid"
)

// Recording status constants
const (
	StatusPending    = "pending"
	StatusRecording  = "recording"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
	StatusArchived   = "archived"
	StatusDeleted    = "deleted"
)

// Recording access levels
const (
	AccessLevelView     = "view"
	AccessLevelDownload = "download"
	AccessLevelShare    = "share"
	AccessLevelDelete   = "delete"
)

// Recording share types
const (
	ShareTypeUser   = "user"
	ShareTypeRole   = "role"
	ShareTypePublic = "public"
	ShareTypeLink   = "link"
)

// Recording represents a room recording
type Recording struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	RoomID          uuid.UUID              `json:"room_id" db:"room_id"`
	Title           string                 `json:"title" db:"title"`
	Description     *string                `json:"description" db:"description"`
	StartedBy       uuid.UUID              `json:"started_by" db:"started_by"`
	StartedAt       time.Time              `json:"started_at" db:"started_at"`
	StoppedAt       *time.Time             `json:"stopped_at" db:"stopped_at"`
	DurationSeconds *int                   `json:"duration_seconds" db:"duration_seconds"`
	Status          string                 `json:"status" db:"status"`
	Format          string                 `json:"format" db:"format"`
	FilePath        *string                `json:"file_path" db:"file_path"`
	FileSizeBytes   *int64                 `json:"file_size_bytes" db:"file_size_bytes"`
	MimeType        *string                `json:"mime_type" db:"mime_type"`
	BitrateKbps     *int                   `json:"bitrate_kbps" db:"bitrate_kbps"`
	FrameRateFps    *int                   `json:"frame_rate_fps" db:"frame_rate_fps"`
	Resolution      *string                `json:"resolution" db:"resolution"`
	Codecs          *string                `json:"codecs" db:"codecs"`
	ErrorMessage    *string                `json:"error_message" db:"error_message"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
	DeletedAt       *time.Time             `json:"deleted_at" db:"deleted_at"`
}

// RecordingParticipant tracks a user in a recording
type RecordingParticipant struct {
	ID            uuid.UUID              `json:"id" db:"id"`
	RecordingID   uuid.UUID              `json:"recording_id" db:"recording_id"`
	UserID        uuid.UUID              `json:"user_id" db:"user_id"`
	PeerID        string                 `json:"peer_id" db:"peer_id"`
	JoinedAt      time.Time              `json:"joined_at" db:"joined_at"`
	LeftAt        *time.Time             `json:"left_at" db:"left_at"`
	ProducerCount int                    `json:"producer_count" db:"producer_count"`
	ConsumerCount int                    `json:"consumer_count" db:"consumer_count"`
	BytesSent     int64                  `json:"bytes_sent" db:"bytes_sent"`
	BytesReceived int64                  `json:"bytes_received" db:"bytes_received"`
	PacketsSent   int64                  `json:"packets_sent" db:"packets_sent"`
	PacketsLost   int64                  `json:"packets_lost" db:"packets_lost"`
	Metadata      map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

// RecordingSharing represents sharing permissions
type RecordingSharing struct {
	ID             uuid.UUID              `json:"id" db:"id"`
	RecordingID    uuid.UUID              `json:"recording_id" db:"recording_id"`
	SharedBy       uuid.UUID              `json:"shared_by" db:"shared_by"`
	SharedWith     *uuid.UUID             `json:"shared_with" db:"shared_with"`
	ShareType      string                 `json:"share_type" db:"share_type"`
	AccessLevel    string                 `json:"access_level" db:"access_level"`
	ExpiryAt       *time.Time             `json:"expiry_at" db:"expiry_at"`
	ShareLinkToken *string                `json:"share_link_token" db:"share_link_token"`
	IsRevoked      bool                   `json:"is_revoked" db:"is_revoked"`
	RevokedAt      *time.Time             `json:"revoked_at" db:"revoked_at"`
	Metadata       map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" db:"updated_at"`
}

// RecordingAccessLog tracks access to recordings
type RecordingAccessLog struct {
	ID               uuid.UUID              `json:"id" db:"id"`
	RecordingID      uuid.UUID              `json:"recording_id" db:"recording_id"`
	UserID           uuid.UUID              `json:"user_id" db:"user_id"`
	Action           string                 `json:"action" db:"action"`
	IPAddress        *string                `json:"ip_address" db:"ip_address"`
	UserAgent        *string                `json:"user_agent" db:"user_agent"`
	BytesTransferred *int64                 `json:"bytes_transferred" db:"bytes_transferred"`
	DurationSeconds  *int                   `json:"duration_seconds" db:"duration_seconds"`
	Status           string                 `json:"status" db:"status"`
	ErrorMessage     *string                `json:"error_message" db:"error_message"`
	Metadata         map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
}

// StartRecordingRequest is the request to start a recording
type StartRecordingRequest struct {
	RoomID      uuid.UUID `json:"room_id" binding:"required"`
	Title       string    `json:"title" binding:"required"`
	Description *string   `json:"description"`
	Format      string    `json:"format,omitempty"` // webm, mp4, etc. Default: webm
	BitrateKbps *int      `json:"bitrate_kbps,omitempty"`
	FrameRate   *int      `json:"frame_rate,omitempty"`
}

// StartRecordingResponse is the response after starting a recording
type StartRecordingResponse struct {
	RecordingID uuid.UUID `json:"recording_id"`
	Status      string    `json:"status"`
	StartedAt   time.Time `json:"started_at"`
	RoomID      uuid.UUID `json:"room_id"`
	Message     string    `json:"message,omitempty"`
}

// StopRecordingRequest is the request to stop a recording
type StopRecordingRequest struct {
	RecordingID uuid.UUID `json:"recording_id" binding:"required"`
}

// StopRecordingResponse is the response after stopping a recording
type StopRecordingResponse struct {
	RecordingID     uuid.UUID `json:"recording_id"`
	Status          string    `json:"status"`
	StoppedAt       time.Time `json:"stopped_at"`
	DurationSeconds int       `json:"duration_seconds,omitempty"`
	FilePath        *string   `json:"file_path,omitempty"`
	Message         string    `json:"message,omitempty"`
}

// GetRecordingResponse represents recording details
type GetRecordingResponse struct {
	ID              uuid.UUID              `json:"id"`
	RoomID          uuid.UUID              `json:"room_id"`
	Title           string                 `json:"title"`
	Description     *string                `json:"description"`
	StartedBy       uuid.UUID              `json:"started_by"`
	StartedAt       time.Time              `json:"started_at"`
	StoppedAt       *time.Time             `json:"stopped_at"`
	DurationSeconds *int                   `json:"duration_seconds"`
	Status          string                 `json:"status"`
	Format          string                 `json:"format"`
	FileSizeBytes   *int64                 `json:"file_size_bytes"`
	MimeType        *string                `json:"mime_type"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// ListRecordingsResponse is the response for listing recordings
type ListRecordingsResponse struct {
	Recordings []GetRecordingResponse `json:"recordings"`
	Total      int                    `json:"total"`
	Limit      int                    `json:"limit"`
	Offset     int                    `json:"offset"`
}

// DeleteRecordingRequest is the request to delete a recording
type DeleteRecordingRequest struct {
	RecordingID uuid.UUID `json:"recording_id" binding:"required"`
	Permanent   *bool     `json:"permanent,omitempty"` // If true, hard delete; else soft delete
}

// DeleteRecordingResponse is the response after deleting a recording
type DeleteRecordingResponse struct {
	RecordingID uuid.UUID `json:"recording_id"`
	Status      string    `json:"status"`
	DeletedAt   time.Time `json:"deleted_at"`
	Message     string    `json:"message,omitempty"`
}

// RecordingListQuery is the query parameters for listing recordings
type RecordingListQuery struct {
	RoomID uuid.UUID
	UserID uuid.UUID
	Status string
	Limit  int
	Offset int
}

// ValidateStatus checks if a status is valid
func ValidateStatus(status string) bool {
	validStatuses := []string{
		StatusPending,
		StatusRecording,
		StatusProcessing,
		StatusCompleted,
		StatusFailed,
		StatusArchived,
		StatusDeleted,
	}
	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// ValidateAccessLevel checks if an access level is valid
func ValidateAccessLevel(level string) bool {
	validLevels := []string{
		AccessLevelView,
		AccessLevelDownload,
		AccessLevelShare,
		AccessLevelDelete,
	}
	for _, l := range validLevels {
		if l == level {
			return true
		}
	}
	return false
}

// ValidateShareType checks if a share type is valid
func ValidateShareType(shareType string) bool {
	validTypes := []string{
		ShareTypeUser,
		ShareTypeRole,
		ShareTypePublic,
		ShareTypeLink,
	}
	for _, t := range validTypes {
		if t == shareType {
			return true
		}
	}
	return false
}
