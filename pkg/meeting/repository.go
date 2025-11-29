package meeting

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
	ErrMeetingNotFound    = errors.New("meeting not found")
	ErrInvalidMeetingData = errors.New("invalid meeting data")
	ErrTimeConflict       = errors.New("meeting time conflicts with existing meeting")
)

// Repository handles meeting database operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new meeting repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new meeting
func (r *Repository) Create(ctx context.Context, meeting *models.Meeting) error {
	if meeting.InstructorID == "" || meeting.SubjectID == "" || meeting.TitleAr == "" {
		return ErrInvalidMeetingData
	}

	meeting.ID = uuid.New().String()
	meeting.CreatedAt = time.Now()
	meeting.UpdatedAt = time.Now()

	if meeting.Status == "" {
		meeting.Status = "scheduled"
	}

	query := `
		INSERT INTO meetings (
			id, instructor_id, student_id, subject_id, title_ar,
			scheduled_at, duration, meeting_url, room_id, status,
			end_time, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.ExecContext(ctx, query,
		meeting.ID,
		meeting.InstructorID,
		nullString(meeting.StudentID),
		meeting.SubjectID,
		meeting.TitleAr,
		meeting.ScheduledAt,
		meeting.Duration,
		nullString(meeting.MeetingURL),
		nullString(meeting.RoomID),
		meeting.Status,
		nullTime(meeting.EndTime),
		meeting.CreatedAt,
		meeting.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create meeting: %w", err)
	}

	return nil
}

// GetByID retrieves a meeting by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*models.Meeting, error) {
	meeting := &models.Meeting{}
	var studentID, meetingURL, roomID sql.NullString
	var endTime sql.NullTime

	query := `
		SELECT id, instructor_id, student_id, subject_id, title_ar,
			   scheduled_at, duration, meeting_url, room_id, status,
			   end_time, created_at, updated_at
		FROM meetings
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&meeting.ID,
		&meeting.InstructorID,
		&studentID,
		&meeting.SubjectID,
		&meeting.TitleAr,
		&meeting.ScheduledAt,
		&meeting.Duration,
		&meetingURL,
		&roomID,
		&meeting.Status,
		&endTime,
		&meeting.CreatedAt,
		&meeting.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrMeetingNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get meeting: %w", err)
	}

	meeting.StudentID = studentID.String
	meeting.MeetingURL = meetingURL.String
	meeting.RoomID = roomID.String
	if endTime.Valid {
		meeting.EndTime = &endTime.Time
	}

	return meeting, nil
}

// List retrieves meetings with filters
func (r *Repository) List(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.Meeting, error) {
	query := `
		SELECT id, instructor_id, student_id, subject_id, title_ar,
			   scheduled_at, duration, meeting_url, room_id, status,
			   end_time, created_at, updated_at
		FROM meetings
		WHERE 1=1
	`
	args := []interface{}{}
	argPos := 1

	// Add filters
	if instructorID, ok := filters["instructor_id"].(string); ok && instructorID != "" {
		query += fmt.Sprintf(" AND instructor_id = $%d", argPos)
		args = append(args, instructorID)
		argPos++
	}

	if studentID, ok := filters["student_id"].(string); ok && studentID != "" {
		query += fmt.Sprintf(" AND student_id = $%d", argPos)
		args = append(args, studentID)
		argPos++
	}

	if subjectID, ok := filters["subject_id"].(string); ok && subjectID != "" {
		query += fmt.Sprintf(" AND subject_id = $%d", argPos)
		args = append(args, subjectID)
		argPos++
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, status)
		argPos++
	}

	if fromDate, ok := filters["from_date"].(time.Time); ok {
		query += fmt.Sprintf(" AND scheduled_at >= $%d", argPos)
		args = append(args, fromDate)
		argPos++
	}

	if toDate, ok := filters["to_date"].(time.Time); ok {
		query += fmt.Sprintf(" AND scheduled_at <= $%d", argPos)
		args = append(args, toDate)
		argPos++
	}

	// Add ordering
	query += " ORDER BY scheduled_at DESC"

	// Add pagination
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, limit)
		argPos++
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list meetings: %w", err)
	}
	defer rows.Close()

	meetings := []*models.Meeting{}
	for rows.Next() {
		meeting := &models.Meeting{}
		var studentID, meetingURL, roomID sql.NullString
		var endTime sql.NullTime

		err := rows.Scan(
			&meeting.ID,
			&meeting.InstructorID,
			&studentID,
			&meeting.SubjectID,
			&meeting.TitleAr,
			&meeting.ScheduledAt,
			&meeting.Duration,
			&meetingURL,
			&roomID,
			&meeting.Status,
			&endTime,
			&meeting.CreatedAt,
			&meeting.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan meeting: %w", err)
		}

		meeting.StudentID = studentID.String
		meeting.MeetingURL = meetingURL.String
		meeting.RoomID = roomID.String
		if endTime.Valid {
			meeting.EndTime = &endTime.Time
		}

		meetings = append(meetings, meeting)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return meetings, nil
}

// Update updates a meeting
func (r *Repository) Update(ctx context.Context, meeting *models.Meeting) error {
	if meeting.ID == "" {
		return ErrInvalidMeetingData
	}

	meeting.UpdatedAt = time.Now()

	query := `
		UPDATE meetings
		SET title_ar = $1, scheduled_at = $2, duration = $3,
			meeting_url = $4, room_id = $5, status = $6,
			end_time = $7, updated_at = $8
		WHERE id = $9
	`

	result, err := r.db.ExecContext(ctx, query,
		meeting.TitleAr,
		meeting.ScheduledAt,
		meeting.Duration,
		nullString(meeting.MeetingURL),
		nullString(meeting.RoomID),
		meeting.Status,
		nullTime(meeting.EndTime),
		meeting.UpdatedAt,
		meeting.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update meeting: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrMeetingNotFound
	}

	return nil
}

// Delete deletes a meeting
func (r *Repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM meetings WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete meeting: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrMeetingNotFound
	}

	return nil
}

// CheckTimeConflict checks if meeting time conflicts with existing meetings
func (r *Repository) CheckTimeConflict(ctx context.Context, instructorID string, scheduledAt time.Time, duration int, excludeID string) (bool, error) {
	endTime := scheduledAt.Add(time.Duration(duration) * time.Minute)

	query := `
		SELECT COUNT(*)
		FROM meetings
		WHERE instructor_id = $1
			AND status IN ('scheduled', 'in-progress')
			AND id != $2
			AND (
				(scheduled_at <= $3 AND (scheduled_at + (duration || ' minutes')::interval) > $3)
				OR
				(scheduled_at < $4 AND (scheduled_at + (duration || ' minutes')::interval) >= $4)
				OR
				(scheduled_at >= $3 AND scheduled_at < $4)
			)
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, instructorID, excludeID, scheduledAt, endTime).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check time conflict: %w", err)
	}

	return count > 0, nil
}

// Helper functions
func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
