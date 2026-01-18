package livestream

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Bashar444/VTP/pkg/models"
	"github.com/google/uuid"
)

var (
	ErrSessionNotFound    = errors.New("streaming session not found")
	ErrInvalidSessionData = errors.New("invalid session data")
	ErrAlreadyLive        = errors.New("session is already live")
	ErrNotLive            = errors.New("session is not live")
	ErrUnauthorized       = errors.New("unauthorized to perform this action")
)

// Repository handles database operations for live streaming
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new live streaming repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateSession creates a new streaming session
func (r *Repository) CreateSession(ctx context.Context, session *models.StreamingSession) error {
	if session.InstructorID == "" || session.Title == "" || session.RoomID == "" {
		return ErrInvalidSessionData
	}

	session.ID = uuid.New().String()
	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()

	if session.Status == "" {
		session.Status = "scheduled"
	}
	if session.MaxParticipants == 0 {
		session.MaxParticipants = 100
	}
	if session.Settings == "" {
		session.Settings = `{"allowChat": true, "allowQuestions": true, "allowRaiseHand": true}`
	}

	query := `
		INSERT INTO streaming_sessions (
			id, room_id, instructor_id, course_id, title, description,
			status, scheduled_at, started_at, ended_at, max_participants,
			is_recording, recording_id, settings, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.RoomID,
		session.InstructorID,
		nullString(session.CourseID),
		session.Title,
		session.Description,
		session.Status,
		nullTime(session.ScheduledAt),
		nullTime(session.StartedAt),
		nullTime(session.EndedAt),
		session.MaxParticipants,
		session.IsRecording,
		nullString(session.RecordingID),
		session.Settings,
		session.CreatedAt,
		session.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetSessionByID retrieves a session by ID
func (r *Repository) GetSessionByID(ctx context.Context, id string) (*models.StreamingSession, error) {
	session := &models.StreamingSession{}
	var courseID, recordingID sql.NullString
	var scheduledAt, startedAt, endedAt sql.NullTime

	query := `
		SELECT id, room_id, instructor_id, course_id, title, description,
			   status, scheduled_at, started_at, ended_at, max_participants,
			   is_recording, recording_id, settings, created_at, updated_at
		FROM streaming_sessions
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID,
		&session.RoomID,
		&session.InstructorID,
		&courseID,
		&session.Title,
		&session.Description,
		&session.Status,
		&scheduledAt,
		&startedAt,
		&endedAt,
		&session.MaxParticipants,
		&session.IsRecording,
		&recordingID,
		&session.Settings,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if courseID.Valid {
		session.CourseID = &courseID.String
	}
	if recordingID.Valid {
		session.RecordingID = &recordingID.String
	}
	if scheduledAt.Valid {
		session.ScheduledAt = &scheduledAt.Time
	}
	if startedAt.Valid {
		session.StartedAt = &startedAt.Time
	}
	if endedAt.Valid {
		session.EndedAt = &endedAt.Time
	}

	return session, nil
}

// GetSessionByRoomID retrieves a session by room ID
func (r *Repository) GetSessionByRoomID(ctx context.Context, roomID string) (*models.StreamingSession, error) {
	session := &models.StreamingSession{}
	var courseID, recordingID sql.NullString
	var scheduledAt, startedAt, endedAt sql.NullTime

	query := `
		SELECT id, room_id, instructor_id, course_id, title, description,
			   status, scheduled_at, started_at, ended_at, max_participants,
			   is_recording, recording_id, settings, created_at, updated_at
		FROM streaming_sessions
		WHERE room_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	err := r.db.QueryRowContext(ctx, query, roomID).Scan(
		&session.ID,
		&session.RoomID,
		&session.InstructorID,
		&courseID,
		&session.Title,
		&session.Description,
		&session.Status,
		&scheduledAt,
		&startedAt,
		&endedAt,
		&session.MaxParticipants,
		&session.IsRecording,
		&recordingID,
		&session.Settings,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if courseID.Valid {
		session.CourseID = &courseID.String
	}
	if recordingID.Valid {
		session.RecordingID = &recordingID.String
	}
	if scheduledAt.Valid {
		session.ScheduledAt = &scheduledAt.Time
	}
	if startedAt.Valid {
		session.StartedAt = &startedAt.Time
	}
	if endedAt.Valid {
		session.EndedAt = &endedAt.Time
	}

	return session, nil
}

// UpdateSession updates a streaming session
func (r *Repository) UpdateSession(ctx context.Context, session *models.StreamingSession) error {
	session.UpdatedAt = time.Now()

	query := `
		UPDATE streaming_sessions
		SET title = $2, description = $3, status = $4, scheduled_at = $5,
			started_at = $6, ended_at = $7, is_recording = $8, recording_id = $9,
			settings = $10, updated_at = $11
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.Title,
		session.Description,
		session.Status,
		nullTime(session.ScheduledAt),
		nullTime(session.StartedAt),
		nullTime(session.EndedAt),
		session.IsRecording,
		nullString(session.RecordingID),
		session.Settings,
		session.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrSessionNotFound
	}

	return nil
}

// ListSessions lists streaming sessions with filters
func (r *Repository) ListSessions(ctx context.Context, instructorID string, status string, liveOnly bool, limit, offset int) ([]*models.StreamingSession, int, error) {
	var sessions []*models.StreamingSession
	var total int

	// Build query with filters
	baseQuery := `FROM streaming_sessions WHERE 1=1`
	args := []interface{}{}
	argCount := 0

	if instructorID != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND instructor_id = $%d", argCount)
		args = append(args, instructorID)
	}

	if status != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}

	if liveOnly {
		baseQuery += " AND status = 'live'"
	}

	// Get total count
	countQuery := "SELECT COUNT(*) " + baseQuery
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count sessions: %w", err)
	}

	// Get sessions
	argCount++
	limitArg := argCount
	argCount++
	offsetArg := argCount

	query := fmt.Sprintf(`
		SELECT id, room_id, instructor_id, course_id, title, description,
			   status, scheduled_at, started_at, ended_at, max_participants,
			   is_recording, recording_id, settings, created_at, updated_at
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, limitArg, offsetArg)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		session := &models.StreamingSession{}
		var courseID, recordingID sql.NullString
		var scheduledAt, startedAt, endedAt sql.NullTime

		err := rows.Scan(
			&session.ID,
			&session.RoomID,
			&session.InstructorID,
			&courseID,
			&session.Title,
			&session.Description,
			&session.Status,
			&scheduledAt,
			&startedAt,
			&endedAt,
			&session.MaxParticipants,
			&session.IsRecording,
			&recordingID,
			&session.Settings,
			&session.CreatedAt,
			&session.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan session: %w", err)
		}

		if courseID.Valid {
			session.CourseID = &courseID.String
		}
		if recordingID.Valid {
			session.RecordingID = &recordingID.String
		}
		if scheduledAt.Valid {
			session.ScheduledAt = &scheduledAt.Time
		}
		if startedAt.Valid {
			session.StartedAt = &startedAt.Time
		}
		if endedAt.Valid {
			session.EndedAt = &endedAt.Time
		}

		sessions = append(sessions, session)
	}

	return sessions, total, nil
}

// AddParticipant adds a participant to a session
func (r *Repository) AddParticipant(ctx context.Context, participant *models.StreamingParticipant) error {
	participant.ID = uuid.New().String()
	participant.JoinedAt = time.Now()
	participant.CreatedAt = time.Now()

	if participant.Role == "" {
		participant.Role = "viewer"
	}

	query := `
		INSERT INTO streaming_participants (
			id, session_id, user_id, role, joined_at, left_at,
			is_muted, is_video_enabled, is_hand_raised, watch_duration, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (session_id, user_id) DO UPDATE SET
			joined_at = EXCLUDED.joined_at,
			left_at = NULL,
			is_muted = EXCLUDED.is_muted,
			is_video_enabled = EXCLUDED.is_video_enabled
	`

	_, err := r.db.ExecContext(ctx, query,
		participant.ID,
		participant.SessionID,
		participant.UserID,
		participant.Role,
		participant.JoinedAt,
		nullTime(participant.LeftAt),
		participant.IsMuted,
		participant.IsVideoEnabled,
		participant.IsHandRaised,
		participant.WatchDuration,
		participant.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}

	return nil
}

// RemoveParticipant marks a participant as left
func (r *Repository) RemoveParticipant(ctx context.Context, sessionID, userID string) error {
	now := time.Now()

	query := `
		UPDATE streaming_participants
		SET left_at = $3
		WHERE session_id = $1 AND user_id = $2 AND left_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, sessionID, userID, now)
	if err != nil {
		return fmt.Errorf("failed to remove participant: %w", err)
	}

	return nil
}

// GetParticipants gets all active participants in a session
func (r *Repository) GetParticipants(ctx context.Context, sessionID string) ([]*models.StreamingParticipant, error) {
	query := `
		SELECT p.id, p.session_id, p.user_id, p.role, p.joined_at, p.left_at,
			   p.is_muted, p.is_video_enabled, p.is_hand_raised, p.watch_duration, p.created_at
		FROM streaming_participants p
		WHERE p.session_id = $1 AND p.left_at IS NULL
		ORDER BY p.joined_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}
	defer rows.Close()

	var participants []*models.StreamingParticipant
	for rows.Next() {
		p := &models.StreamingParticipant{}
		var leftAt sql.NullTime

		err := rows.Scan(
			&p.ID,
			&p.SessionID,
			&p.UserID,
			&p.Role,
			&p.JoinedAt,
			&leftAt,
			&p.IsMuted,
			&p.IsVideoEnabled,
			&p.IsHandRaised,
			&p.WatchDuration,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan participant: %w", err)
		}

		if leftAt.Valid {
			p.LeftAt = &leftAt.Time
		}

		participants = append(participants, p)
	}

	return participants, nil
}

// GetParticipantCount gets the count of active participants
func (r *Repository) GetParticipantCount(ctx context.Context, sessionID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM streaming_participants WHERE session_id = $1 AND left_at IS NULL`

	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count participants: %w", err)
	}

	return count, nil
}

// Helper functions
func nullString(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
