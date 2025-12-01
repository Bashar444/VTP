package course

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// CourseService manages course operations
type CourseService struct {
	db     *sql.DB
	logger *log.Logger
}

// NewCourseService creates a new course service
func NewCourseService(db *sql.DB, logger *log.Logger) *CourseService {
	if logger == nil {
		logger = log.New(log.Writer(), "[CourseService] ", log.LstdFlags)
	}
	return &CourseService{
		db:     db,
		logger: logger,
	}
}

// CreateCourse creates a new course
func (cs *CourseService) CreateCourse(ctx context.Context, req CreateCourseRequest, instructorID uuid.UUID) (*Course, error) {
	if req.Code == "" || req.Name == "" {
		return nil, errors.New("code and name are required")
	}

	id := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO courses (id, code, name, description, instructor_id, department, semester, year, status, max_students, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, code, name, description, instructor_id, department, semester, year, status, max_students, created_at, updated_at
	`

	course := &Course{}
	err := cs.db.QueryRowContext(ctx, query,
		id, req.Code, req.Name, req.Description, instructorID, req.Department,
		req.Semester, req.Year, StatusDraft, req.MaxStudents, now, now,
	).Scan(
		&course.ID, &course.Code, &course.Name, &course.Description, &course.InstructorID,
		&course.Department, &course.Semester, &course.Year, &course.Status, &course.MaxStudents,
		&course.CreatedAt, &course.UpdatedAt,
	)

	if err != nil {
		cs.logger.Printf("Error creating course: %v", err)
		return nil, fmt.Errorf("failed to create course: %w", err)
	}

	cs.logger.Printf("Course created: %s (ID: %s)", course.Name, course.ID)
	return course, nil
}

// GetCourse retrieves a course by ID
func (cs *CourseService) GetCourse(ctx context.Context, courseID uuid.UUID) (*Course, error) {
	query := `
		SELECT id, code, name, description, instructor_id, department, semester, year, status, max_students, created_at, updated_at
		FROM courses
		WHERE id = $1
	`

	course := &Course{}
	err := cs.db.QueryRowContext(ctx, query, courseID).Scan(
		&course.ID, &course.Code, &course.Name, &course.Description, &course.InstructorID,
		&course.Department, &course.Semester, &course.Year, &course.Status, &course.MaxStudents,
		&course.CreatedAt, &course.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("course not found: %s", courseID)
		}
		return nil, fmt.Errorf("failed to get course: %w", err)
	}

	return course, nil
}

// ListCourses retrieves courses with optional filtering
func (cs *CourseService) ListCourses(ctx context.Context, filter map[string]interface{}) ([]*Course, error) {
	query := `
		SELECT id, code, name, description, instructor_id, department, semester, year, status, max_students, created_at, updated_at
		FROM courses
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	// Add filters
	if semester, ok := filter["semester"].(string); ok && semester != "" {
		query += fmt.Sprintf(" AND semester = $%d", argIdx)
		args = append(args, semester)
		argIdx++
	}
	if year, ok := filter["year"].(int); ok && year > 0 {
		query += fmt.Sprintf(" AND year = $%d", argIdx)
		args = append(args, year)
		argIdx++
	}
	if instructorID, ok := filter["instructor_id"].(uuid.UUID); ok && instructorID != uuid.Nil {
		query += fmt.Sprintf(" AND instructor_id = $%d", argIdx)
		args = append(args, instructorID)
		argIdx++
	}
	if status, ok := filter["status"].(string); ok && status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	query += " ORDER BY created_at DESC"

	rows, err := cs.db.QueryContext(ctx, query, args...)
	if err != nil {
		cs.logger.Printf("Error listing courses: %v", err)
		return nil, fmt.Errorf("failed to list courses: %w", err)
	}
	defer rows.Close()

	var courses []*Course
	for rows.Next() {
		course := &Course{}
		if err := rows.Scan(
			&course.ID, &course.Code, &course.Name, &course.Description, &course.InstructorID,
			&course.Department, &course.Semester, &course.Year, &course.Status, &course.MaxStudents,
			&course.CreatedAt, &course.UpdatedAt,
		); err != nil {
			cs.logger.Printf("Error scanning course: %v", err)
			return nil, fmt.Errorf("failed to scan course: %w", err)
		}
		courses = append(courses, course)
	}

	return courses, nil
}

// UpdateCourse updates a course
func (cs *CourseService) UpdateCourse(ctx context.Context, courseID uuid.UUID, req UpdateCourseRequest) (*Course, error) {
	query := `
		UPDATE courses
		SET name = COALESCE($1, name),
		    description = COALESCE($2, description),
		    department = COALESCE($3, department),
		    semester = COALESCE($4, semester),
		    year = COALESCE($5, year),
		    max_students = COALESCE($6, max_students),
		    status = COALESCE($7, status),
		    updated_at = NOW()
		WHERE id = $8
		RETURNING id, code, name, description, instructor_id, department, semester, year, status, max_students, created_at, updated_at
	`

	course := &Course{}
	err := cs.db.QueryRowContext(ctx, query,
		req.Name, req.Description, req.Department, req.Semester,
		req.Year, req.MaxStudents, req.Status, courseID,
	).Scan(
		&course.ID, &course.Code, &course.Name, &course.Description, &course.InstructorID,
		&course.Department, &course.Semester, &course.Year, &course.Status, &course.MaxStudents,
		&course.CreatedAt, &course.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("course not found: %s", courseID)
		}
		return nil, fmt.Errorf("failed to update course: %w", err)
	}

	cs.logger.Printf("Course updated: %s (ID: %s)", course.Name, course.ID)
	return course, nil
}

// DeleteCourse deletes a course
func (cs *CourseService) DeleteCourse(ctx context.Context, courseID uuid.UUID) error {
	query := `DELETE FROM courses WHERE id = $1`
	result, err := cs.db.ExecContext(ctx, query, courseID)
	if err != nil {
		cs.logger.Printf("Error deleting course: %v", err)
		return fmt.Errorf("failed to delete course: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("course not found: %s", courseID)
	}

	cs.logger.Printf("Course deleted: ID %s", courseID)
	return nil
}

// EnrollStudent enrolls a student in a course
func (cs *CourseService) EnrollStudent(ctx context.Context, courseID, studentID uuid.UUID) (*CourseEnrollment, error) {
	id := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO course_enrollments (id, course_id, student_id, enrollment_date, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, course_id, student_id, enrollment_date, status
	`

	enrollment := &CourseEnrollment{}
	err := cs.db.QueryRowContext(ctx, query,
		id, courseID, studentID, now, EnrollmentActive,
	).Scan(
		&enrollment.ID, &enrollment.CourseID, &enrollment.StudentID,
		&enrollment.EnrollmentDate, &enrollment.Status,
	)

	if err != nil {
		cs.logger.Printf("Error enrolling student: %v", err)
		return nil, fmt.Errorf("failed to enroll student: %w", err)
	}

	cs.logger.Printf("Student enrolled: %s in course %s", studentID, courseID)
	return enrollment, nil
}

// RemoveStudent removes a student enrollment
func (cs *CourseService) RemoveStudent(ctx context.Context, courseID, studentID uuid.UUID) error {
	query := `DELETE FROM course_enrollments WHERE course_id = $1 AND student_id = $2`
	result, err := cs.db.ExecContext(ctx, query, courseID, studentID)
	if err != nil {
		return fmt.Errorf("failed to remove student: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("enrollment not found")
	}

	return nil
}

// ListEnrollments lists students in a course
func (cs *CourseService) ListEnrollments(ctx context.Context, courseID uuid.UUID) ([]*CourseEnrollment, error) {
	query := `
		SELECT id, course_id, student_id, enrollment_date, status
		FROM course_enrollments
		WHERE course_id = $1
		ORDER BY enrollment_date DESC
	`

	rows, err := cs.db.QueryContext(ctx, query, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to list enrollments: %w", err)
	}
	defer rows.Close()

	var enrollments []*CourseEnrollment
	for rows.Next() {
		enrollment := &CourseEnrollment{}
		if err := rows.Scan(
			&enrollment.ID, &enrollment.CourseID, &enrollment.StudentID,
			&enrollment.EnrollmentDate, &enrollment.Status,
		); err != nil {
			return nil, fmt.Errorf("failed to scan enrollment: %w", err)
		}
		enrollments = append(enrollments, enrollment)
	}

	return enrollments, nil
}

// GetStudentEnrollments lists courses a student is enrolled in
func (cs *CourseService) GetStudentEnrollments(ctx context.Context, studentID uuid.UUID) ([]*CourseEnrollment, error) {
	query := `
		SELECT e.id, e.course_id, e.student_id, e.enrollment_date, e.status,
		       c.code, c.name, c.description, c.instructor_id
		FROM course_enrollments e
		JOIN courses c ON e.course_id = c.id
		WHERE e.student_id = $1 AND e.status = $2
		ORDER BY e.enrollment_date DESC
	`

	rows, err := cs.db.QueryContext(ctx, query, studentID, EnrollmentActive)
	if err != nil {
		return nil, fmt.Errorf("failed to list student enrollments: %w", err)
	}
	defer rows.Close()

	var enrollments []*CourseEnrollment
	for rows.Next() {
		enrollment := &CourseEnrollment{}
		var code, name, description string
		var instructorID uuid.UUID
		if err := rows.Scan(
			&enrollment.ID, &enrollment.CourseID, &enrollment.StudentID,
			&enrollment.EnrollmentDate, &enrollment.Status,
			&code, &name, &description, &instructorID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan enrollment: %w", err)
		}
		enrollments = append(enrollments, enrollment)
	}

	return enrollments, nil
}

// SetPermission sets user permission for a course
func (cs *CourseService) SetPermission(ctx context.Context, courseID, userID uuid.UUID, role string) (*CoursePermission, error) {
	id := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO course_permissions (id, course_id, user_id, role, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (course_id, user_id) DO UPDATE
		SET role = $4
		RETURNING id, course_id, user_id, role, created_at
	`

	permission := &CoursePermission{}
	err := cs.db.QueryRowContext(ctx, query,
		id, courseID, userID, role, now,
	).Scan(
		&permission.ID, &permission.CourseID, &permission.UserID,
		&permission.Role, &permission.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to set permission: %w", err)
	}

	cs.logger.Printf("Permission set: %s for user %s in course %s", role, userID, courseID)
	return permission, nil
}

// GetPermission gets user permission for a course
func (cs *CourseService) GetPermission(ctx context.Context, courseID, userID uuid.UUID) (*CoursePermission, error) {
	query := `
		SELECT id, course_id, user_id, role, created_at
		FROM course_permissions
		WHERE course_id = $1 AND user_id = $2
	`

	permission := &CoursePermission{}
	err := cs.db.QueryRowContext(ctx, query, courseID, userID).Scan(
		&permission.ID, &permission.CourseID, &permission.UserID,
		&permission.Role, &permission.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No permission found (not an error)
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return permission, nil
}

// AddRecordingToCourse links a recording to a course
func (cs *CourseService) AddRecordingToCourse(ctx context.Context, courseID, recordingID uuid.UUID, lectureNumber int, lectureTitle string, sequence int) (*CourseRecording, error) {
	id := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO course_recordings (id, course_id, recording_id, lecture_number, lecture_title, sequence_order, is_published, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, course_id, recording_id, lecture_number, lecture_title, sequence_order, is_published, created_at
	`

	recording := &CourseRecording{}
	err := cs.db.QueryRowContext(ctx, query,
		id, courseID, recordingID, lectureNumber, lectureTitle, sequence, false, now,
	).Scan(
		&recording.ID, &recording.CourseID, &recording.RecordingID,
		&recording.LectureNumber, &recording.LectureTitle, &recording.SequenceOrder,
		&recording.IsPublished, &recording.CreatedAt,
	)

	if err != nil {
		cs.logger.Printf("Error adding recording to course: %v", err)
		return nil, fmt.Errorf("failed to add recording: %w", err)
	}

	return recording, nil
}

// PublishCourseRecording publishes a recording in a course
func (cs *CourseService) PublishCourseRecording(ctx context.Context, courseID, recordingID uuid.UUID) error {
	query := `
		UPDATE course_recordings
		SET is_published = true
		WHERE course_id = $1 AND recording_id = $2
	`

	result, err := cs.db.ExecContext(ctx, query, courseID, recordingID)
	if err != nil {
		return fmt.Errorf("failed to publish recording: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("recording not found in course")
	}

	return nil
}

// LogCourseActivity logs a course activity
func (cs *CourseService) LogCourseActivity(ctx context.Context, courseID uuid.UUID, userID *uuid.UUID, action string, details map[string]interface{}, ipAddress string, userAgent string) error {
	id := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO course_activity (id, course_id, user_id, action, details, ip_address, user_agent, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := cs.db.ExecContext(ctx, query,
		id, courseID, userID, action, details, ipAddress, userAgent, now,
	)

	if err != nil {
		cs.logger.Printf("Error logging activity: %v", err)
		return fmt.Errorf("failed to log activity: %w", err)
	}

	return nil
}

// LogRecordingAccess logs recording access
func (cs *CourseService) LogRecordingAccess(ctx context.Context, recordingID uuid.UUID, courseID *uuid.UUID, userID *uuid.UUID, action string, ipAddress string, userAgent string) error {
	id := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO recording_access_logs (id, recording_id, course_id, user_id, action, ip_address, user_agent, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := cs.db.ExecContext(ctx, query,
		id, recordingID, courseID, userID, action, ipAddress, userAgent, now,
	)

	if err != nil {
		cs.logger.Printf("Error logging recording access: %v", err)
		return fmt.Errorf("failed to log access: %w", err)
	}

	return nil
}
