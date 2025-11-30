package recording

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// RecordingService handles operational recording lifecycle (start/stop/update)
type RecordingService struct {
	db  *sql.DB
	log *log.Logger
}

// NewRecordingService creates a new recording service
func NewRecordingService(db *sql.DB, logger *log.Logger) *RecordingService {
	return &RecordingService{
		db:  db,
		log: logger,
	}
}

// StartRecording creates a new recording entry in the database
func (s *RecordingService) StartRecording(ctx context.Context, req *StartRecordingRequest, userID uuid.UUID) (*Recording, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	if req.RoomID == uuid.Nil {
		return nil, errors.New("room_id is required")
	}

	if req.Title == "" {
		return nil, errors.New("title is required")
	}

	recording := &Recording{
		ID:          uuid.New(),
		RoomID:      req.RoomID,
		Title:       req.Title,
		Description: req.Description,
		StartedBy:   userID,
		StartedAt:   time.Now().UTC(),
		Status:      StatusPending,
		Format:      "webm",
		Metadata:    make(map[string]interface{}),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	// Set optional parameters
	if req.Format != "" {
		recording.Format = req.Format
	}
	if req.BitrateKbps != nil && *req.BitrateKbps > 0 {
		recording.BitrateKbps = req.BitrateKbps
	}
	if req.FrameRate != nil && *req.FrameRate > 0 {
		recording.FrameRateFps = req.FrameRate
	}

	// Store metadata
	if req.Description != nil {
		recording.Metadata["description"] = *req.Description
	}

	query := `
		INSERT INTO recordings 
		(id, room_id, title, description, started_by, started_at, status, format, bitrate_kbps, frame_rate_fps, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, room_id, title, description, started_by, started_at, status, format, metadata, created_at, updated_at
	`

	err := s.db.QueryRowContext(
		ctx,
		query,
		recording.ID,
		recording.RoomID,
		recording.Title,
		recording.Description,
		recording.StartedBy,
		recording.StartedAt,
		recording.Status,
		recording.Format,
		recording.BitrateKbps,
		recording.FrameRateFps,
		recording.Metadata,
		recording.CreatedAt,
		recording.UpdatedAt,
	).Scan(
		&recording.ID,
		&recording.RoomID,
		&recording.Title,
		&recording.Description,
		&recording.StartedBy,
		&recording.StartedAt,
		&recording.Status,
		&recording.Format,
		&recording.Metadata,
		&recording.CreatedAt,
		&recording.UpdatedAt,
	)

	if err != nil {
		s.log.Printf("Error starting recording: %v", err)
		return nil, fmt.Errorf("failed to start recording: %w", err)
	}

	s.log.Printf("Recording %s started for room %s by user %s", recording.ID, recording.RoomID, userID)
	return recording, nil
}

// StopRecording marks a recording as stopped and updates duration
func (s *RecordingService) StopRecording(ctx context.Context, recordingID uuid.UUID) (*Recording, error) {
	if recordingID == uuid.Nil {
		return nil, errors.New("recording_id is required")
	}

	now := time.Now().UTC()

	// Get the recording first
	recording, err := s.GetRecording(ctx, recordingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recording: %w", err)
	}

	if recording == nil {
		return nil, errors.New("recording not found")
	}

	if recording.Status == StatusDeleted {
		return nil, errors.New("cannot stop a deleted recording")
	}

	// Calculate duration
	duration := int(now.Sub(recording.StartedAt).Seconds())

	// Update recording status
	query := `
		UPDATE recordings
		SET status = $1, stopped_at = $2, duration_seconds = $3, updated_at = $4
		WHERE id = $5
		RETURNING id, room_id, title, description, started_by, started_at, stopped_at, 
		          duration_seconds, status, format, file_path, file_size_bytes, mime_type, 
		          bitrate_kbps, frame_rate_fps, resolution, codecs, metadata, created_at, updated_at
	`

	err = s.db.QueryRowContext(ctx, query, StatusCompleted, now, duration, now, recordingID).Scan(
		&recording.ID,
		&recording.RoomID,
		&recording.Title,
		&recording.Description,
		&recording.StartedBy,
		&recording.StartedAt,
		&recording.StoppedAt,
		&recording.DurationSeconds,
		&recording.Status,
		&recording.Format,
		&recording.FilePath,
		&recording.FileSizeBytes,
		&recording.MimeType,
		&recording.BitrateKbps,
		&recording.FrameRateFps,
		&recording.Resolution,
		&recording.Codecs,
		&recording.Metadata,
		&recording.CreatedAt,
		&recording.UpdatedAt,
	)

	if err != nil {
		s.log.Printf("Error stopping recording: %v", err)
		return nil, fmt.Errorf("failed to stop recording: %w", err)
	}

	s.log.Printf("Recording %s stopped. Duration: %d seconds", recordingID, duration)
	return recording, nil
}

// GetRecording retrieves a recording by ID
func (s *RecordingService) GetRecording(ctx context.Context, recordingID uuid.UUID) (*Recording, error) {
	if recordingID == uuid.Nil {
		return nil, errors.New("recording_id is required")
	}

	recording := &Recording{}
	query := `
		SELECT id, room_id, title, description, started_by, started_at, stopped_at,
		       duration_seconds, status, format, file_path, file_size_bytes, mime_type,
		       bitrate_kbps, frame_rate_fps, resolution, codecs, error_message, metadata, 
		       created_at, updated_at, deleted_at
		FROM recordings
		WHERE id = $1
	`

	err := s.db.QueryRowContext(ctx, query, recordingID).Scan(
		&recording.ID,
		&recording.RoomID,
		&recording.Title,
		&recording.Description,
		&recording.StartedBy,
		&recording.StartedAt,
		&recording.StoppedAt,
		&recording.DurationSeconds,
		&recording.Status,
		&recording.Format,
		&recording.FilePath,
		&recording.FileSizeBytes,
		&recording.MimeType,
		&recording.BitrateKbps,
		&recording.FrameRateFps,
		&recording.Resolution,
		&recording.Codecs,
		&recording.ErrorMessage,
		&recording.Metadata,
		&recording.CreatedAt,
		&recording.UpdatedAt,
		&recording.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		s.log.Printf("Error getting recording: %v", err)
		return nil, fmt.Errorf("failed to get recording: %w", err)
	}

	return recording, nil
}

// ListRecordings retrieves recordings with optional filters
func (s *RecordingService) ListRecordings(ctx context.Context, query *RecordingListQuery) ([]Recording, int, error) {
	if query == nil {
		query = &RecordingListQuery{
			Limit:  10,
			Offset: 0,
		}
	}

	// Default limit/offset
	if query.Limit <= 0 || query.Limit > 100 {
		query.Limit = 10
	}
	if query.Offset < 0 {
		query.Offset = 0
	}

	// Build dynamic query
	whereConditions := []string{"deleted_at IS NULL"}
	args := []interface{}{}
	argCounter := 1

	if query.RoomID != uuid.Nil {
		whereConditions = append(whereConditions, fmt.Sprintf("room_id = $%d", argCounter))
		args = append(args, query.RoomID)
		argCounter++
	}

	if query.UserID != uuid.Nil {
		whereConditions = append(whereConditions, fmt.Sprintf("started_by = $%d", argCounter))
		args = append(args, query.UserID)
		argCounter++
	}

	if query.Status != "" && ValidateStatus(query.Status) {
		whereConditions = append(whereConditions, fmt.Sprintf("status = $%d", argCounter))
		args = append(args, query.Status)
		argCounter++
	}

	whereClause := ""
	for i, cond := range whereConditions {
		if i == 0 {
			whereClause = cond
		} else {
			whereClause += " AND " + cond
		}
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM recordings WHERE %s", whereClause)
	var total int
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		s.log.Printf("Error counting recordings: %v", err)
		return nil, 0, fmt.Errorf("failed to count recordings: %w", err)
	}

	// Get paginated results
	args = append(args, query.Limit)
	args = append(args, query.Offset)

	dataQuery := fmt.Sprintf(`
		SELECT id, room_id, title, description, started_by, started_at, stopped_at,
		       duration_seconds, status, format, file_path, file_size_bytes, mime_type,
		       bitrate_kbps, frame_rate_fps, resolution, codecs, error_message, metadata,
		       created_at, updated_at, deleted_at
		FROM recordings
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCounter, argCounter+1)

	rows, err := s.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		s.log.Printf("Error querying recordings: %v", err)
		return nil, 0, fmt.Errorf("failed to query recordings: %w", err)
	}
	defer rows.Close()

	recordings := []Recording{}
	for rows.Next() {
		rec := Recording{}
		err := rows.Scan(
			&rec.ID,
			&rec.RoomID,
			&rec.Title,
			&rec.Description,
			&rec.StartedBy,
			&rec.StartedAt,
			&rec.StoppedAt,
			&rec.DurationSeconds,
			&rec.Status,
			&rec.Format,
			&rec.FilePath,
			&rec.FileSizeBytes,
			&rec.MimeType,
			&rec.BitrateKbps,
			&rec.FrameRateFps,
			&rec.Resolution,
			&rec.Codecs,
			&rec.ErrorMessage,
			&rec.Metadata,
			&rec.CreatedAt,
			&rec.UpdatedAt,
			&rec.DeletedAt,
		)
		if err != nil {
			s.log.Printf("Error scanning recording: %v", err)
			continue
		}
		recordings = append(recordings, rec)
	}

	if err = rows.Err(); err != nil {
		s.log.Printf("Error iterating recordings: %v", err)
		return nil, 0, fmt.Errorf("failed to iterate recordings: %w", err)
	}

	return recordings, total, nil
}

// DeleteRecording performs a soft delete of a recording
func (s *RecordingService) DeleteRecording(ctx context.Context, recordingID uuid.UUID) (*Recording, error) {
	if recordingID == uuid.Nil {
		return nil, errors.New("recording_id is required")
	}

	now := time.Now().UTC()

	recording := &Recording{}
	query := `
		UPDATE recordings
		SET deleted_at = $1, status = $2, updated_at = $3
		WHERE id = $4
		RETURNING id, room_id, title, description, started_by, started_at, stopped_at,
		          duration_seconds, status, format, file_path, file_size_bytes, mime_type,
		          bitrate_kbps, frame_rate_fps, resolution, codecs, error_message, metadata,
		          created_at, updated_at, deleted_at
	`

	err := s.db.QueryRowContext(ctx, query, now, StatusDeleted, now, recordingID).Scan(
		&recording.ID,
		&recording.RoomID,
		&recording.Title,
		&recording.Description,
		&recording.StartedBy,
		&recording.StartedAt,
		&recording.StoppedAt,
		&recording.DurationSeconds,
		&recording.Status,
		&recording.Format,
		&recording.FilePath,
		&recording.FileSizeBytes,
		&recording.MimeType,
		&recording.BitrateKbps,
		&recording.FrameRateFps,
		&recording.Resolution,
		&recording.Codecs,
		&recording.ErrorMessage,
		&recording.Metadata,
		&recording.CreatedAt,
		&recording.UpdatedAt,
		&recording.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("recording not found")
	}

	if err != nil {
		s.log.Printf("Error deleting recording: %v", err)
		return nil, fmt.Errorf("failed to delete recording: %w", err)
	}

	s.log.Printf("Recording %s deleted", recordingID)
	return recording, nil
}

// UpdateRecordingStatus updates the status of a recording
func (s *RecordingService) UpdateRecordingStatus(ctx context.Context, recordingID uuid.UUID, status string) error {
	if recordingID == uuid.Nil {
		return errors.New("recording_id is required")
	}

	if !ValidateStatus(status) {
		return fmt.Errorf("invalid status: %s", status)
	}

	query := `UPDATE recordings SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := s.db.ExecContext(ctx, query, status, time.Now().UTC(), recordingID)
	if err != nil {
		s.log.Printf("Error updating recording status: %v", err)
		return fmt.Errorf("failed to update recording status: %w", err)
	}

	s.log.Printf("Recording %s status updated to %s", recordingID, status)
	return nil
}

// UpdateRecordingMetadata updates metadata for a recording
func (s *RecordingService) UpdateRecordingMetadata(ctx context.Context, recordingID uuid.UUID, metadata map[string]interface{}) error {
	if recordingID == uuid.Nil {
		return errors.New("recording_id is required")
	}

	query := `UPDATE recordings SET metadata = $1, updated_at = $2 WHERE id = $3`
	_, err := s.db.ExecContext(ctx, query, metadata, time.Now().UTC(), recordingID)
	if err != nil {
		s.log.Printf("Error updating recording metadata: %v", err)
		return fmt.Errorf("failed to update recording metadata: %w", err)
	}

	return nil
}

// GetRecordingStats retrieves statistics for a recording
func (s *RecordingService) GetRecordingStats(ctx context.Context, recordingID uuid.UUID) (map[string]interface{}, error) {
	if recordingID == uuid.Nil {
		return nil, errors.New("recording_id is required")
	}

	stats := make(map[string]interface{})

	// Get recording info
	recording, err := s.GetRecording(ctx, recordingID)
	if err != nil {
		return nil, err
	}

	if recording == nil {
		return nil, errors.New("recording not found")
	}

	stats["recording_id"] = recording.ID
	stats["duration_seconds"] = recording.DurationSeconds
	stats["file_size_bytes"] = recording.FileSizeBytes
	stats["status"] = recording.Status

	// Get participant count
	var participantCount int
	query := `SELECT COUNT(DISTINCT user_id) FROM recording_participants WHERE recording_id = $1`
	err = s.db.QueryRowContext(ctx, query, recordingID).Scan(&participantCount)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get participant count: %w", err)
	}
	stats["participant_count"] = participantCount

	// Get access count
	var accessCount int
	query = `SELECT COUNT(*) FROM recording_access_log WHERE recording_id = $1`
	err = s.db.QueryRowContext(ctx, query, recordingID).Scan(&accessCount)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get access count: %w", err)
	}
	stats["access_count"] = accessCount

	return stats, nil
}
