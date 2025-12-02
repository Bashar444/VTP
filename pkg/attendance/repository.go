package attendance

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Bashar444/VTP/pkg/models"
	"github.com/google/uuid"
)

var (
	ErrAttendanceNotFound = errors.New("attendance record not found")
	ErrDuplicateRecord    = errors.New("attendance record already exists")
)

// Repository handles attendance data persistence
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new attendance repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts a new attendance record
func (r *Repository) Create(ctx context.Context, a *models.Attendance) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()

	query := `
		INSERT INTO attendance (
			id, student_id, class_section_id, meeting_id, subject_id,
			date, status, check_in_time, check_out_time, recorded_by, notes,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.ExecContext(ctx, query,
		a.ID, a.StudentID, a.ClassSectionID, a.MeetingID, a.SubjectID,
		a.Date, a.Status, a.CheckInTime, a.CheckOutTime, a.RecordedBy, a.Notes,
		a.CreatedAt, a.UpdatedAt,
	)
	return err
}

// GetByID retrieves an attendance record by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*models.Attendance, error) {
	query := `
		SELECT id, student_id, class_section_id, meeting_id, subject_id,
			date, status, check_in_time, check_out_time, recorded_by, notes,
			created_at, updated_at
		FROM attendance WHERE id = $1
	`
	var a models.Attendance
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.StudentID, &a.ClassSectionID, &a.MeetingID, &a.SubjectID,
		&a.Date, &a.Status, &a.CheckInTime, &a.CheckOutTime, &a.RecordedBy, &a.Notes,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrAttendanceNotFound
	}
	return &a, err
}

// Update modifies an existing attendance record
func (r *Repository) Update(ctx context.Context, a *models.Attendance) error {
	a.UpdatedAt = time.Now()
	query := `
		UPDATE attendance SET
			status = $2, check_in_time = $3, check_out_time = $4,
			notes = $5, updated_at = $6
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query,
		a.ID, a.Status, a.CheckInTime, a.CheckOutTime, a.Notes, a.UpdatedAt,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrAttendanceNotFound
	}
	return nil
}

// ListByStudent retrieves attendance records for a student
func (r *Repository) ListByStudent(ctx context.Context, studentID string, startDate, endDate time.Time) ([]models.Attendance, error) {
	query := `
		SELECT id, student_id, class_section_id, meeting_id, subject_id,
			date, status, check_in_time, check_out_time, recorded_by, notes,
			created_at, updated_at
		FROM attendance
		WHERE student_id = $1 AND date >= $2 AND date <= $3
		ORDER BY date DESC
	`
	rows, err := r.db.QueryContext(ctx, query, studentID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.Attendance
	for rows.Next() {
		var a models.Attendance
		if err := rows.Scan(
			&a.ID, &a.StudentID, &a.ClassSectionID, &a.MeetingID, &a.SubjectID,
			&a.Date, &a.Status, &a.CheckInTime, &a.CheckOutTime, &a.RecordedBy, &a.Notes,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, a)
	}
	return records, nil
}

// ListByClassSection retrieves attendance records for a class section on a date
func (r *Repository) ListByClassSection(ctx context.Context, classSectionID string, date time.Time) ([]models.Attendance, error) {
	query := `
		SELECT id, student_id, class_section_id, meeting_id, subject_id,
			date, status, check_in_time, check_out_time, recorded_by, notes,
			created_at, updated_at
		FROM attendance
		WHERE class_section_id = $1 AND date = $2
		ORDER BY student_id
	`
	rows, err := r.db.QueryContext(ctx, query, classSectionID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.Attendance
	for rows.Next() {
		var a models.Attendance
		if err := rows.Scan(
			&a.ID, &a.StudentID, &a.ClassSectionID, &a.MeetingID, &a.SubjectID,
			&a.Date, &a.Status, &a.CheckInTime, &a.CheckOutTime, &a.RecordedBy, &a.Notes,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, a)
	}
	return records, nil
}

// ListByMeeting retrieves attendance records for a meeting
func (r *Repository) ListByMeeting(ctx context.Context, meetingID string) ([]models.Attendance, error) {
	query := `
		SELECT id, student_id, class_section_id, meeting_id, subject_id,
			date, status, check_in_time, check_out_time, recorded_by, notes,
			created_at, updated_at
		FROM attendance
		WHERE meeting_id = $1
		ORDER BY student_id
	`
	rows, err := r.db.QueryContext(ctx, query, meetingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.Attendance
	for rows.Next() {
		var a models.Attendance
		if err := rows.Scan(
			&a.ID, &a.StudentID, &a.ClassSectionID, &a.MeetingID, &a.SubjectID,
			&a.Date, &a.Status, &a.CheckInTime, &a.CheckOutTime, &a.RecordedBy, &a.Notes,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, a)
	}
	return records, nil
}

// GetStudentStats retrieves attendance statistics for a student
func (r *Repository) GetStudentStats(ctx context.Context, studentID string, startDate, endDate time.Time) (*AttendanceStats, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'present') as present,
			COUNT(*) FILTER (WHERE status = 'absent') as absent,
			COUNT(*) FILTER (WHERE status = 'late') as late,
			COUNT(*) FILTER (WHERE status = 'excused') as excused
		FROM attendance
		WHERE student_id = $1 AND date >= $2 AND date <= $3
	`
	var stats AttendanceStats
	err := r.db.QueryRowContext(ctx, query, studentID, startDate, endDate).Scan(
		&stats.Total, &stats.Present, &stats.Absent, &stats.Late, &stats.Excused,
	)
	if err != nil {
		return nil, err
	}
	if stats.Total > 0 {
		stats.AttendanceRate = float64(stats.Present+stats.Late) / float64(stats.Total) * 100
	}
	return &stats, nil
}

// AttendanceStats holds attendance statistics
type AttendanceStats struct {
	Total          int     `json:"total"`
	Present        int     `json:"present"`
	Absent         int     `json:"absent"`
	Late           int     `json:"late"`
	Excused        int     `json:"excused"`
	AttendanceRate float64 `json:"attendance_rate"`
}

// BulkCreate inserts multiple attendance records
func (r *Repository) BulkCreate(ctx context.Context, records []models.Attendance) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO attendance (
			id, student_id, class_section_id, meeting_id, subject_id,
			date, status, check_in_time, check_out_time, recorded_by, notes,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (student_id, class_section_id, date, subject_id) DO UPDATE SET
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at
	`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()
	for _, a := range records {
		if a.ID == "" {
			a.ID = uuid.New().String()
		}
		_, err = stmt.ExecContext(ctx,
			a.ID, a.StudentID, a.ClassSectionID, a.MeetingID, a.SubjectID,
			a.Date, a.Status, a.CheckInTime, a.CheckOutTime, a.RecordedBy, a.Notes,
			now, now,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// MarkPresent marks a student as present for a meeting
func (r *Repository) MarkPresent(ctx context.Context, studentID, meetingID string) error {
	now := time.Now()
	query := `
		INSERT INTO attendance (id, student_id, meeting_id, date, status, check_in_time, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'present', $5, $5, $5)
		ON CONFLICT (student_id, meeting_id) DO UPDATE SET
			status = 'present',
			check_in_time = COALESCE(attendance.check_in_time, EXCLUDED.check_in_time),
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.ExecContext(ctx, query, uuid.New().String(), studentID, meetingID, now.Format("2006-01-02"), now)
	return err
}
