package course

import (
	"time"

	"github.com/google/uuid"
)

// Course Status constants
const (
	StatusDraft     = "draft"
	StatusActive    = "active"
	StatusArchived  = "archived"
	StatusCompleted = "completed"
)

// Enrollment Status constants
const (
	EnrollmentActive    = "active"
	EnrollmentCompleted = "completed"
	EnrollmentDropped   = "dropped"
	EnrollmentSuspended = "suspended"
)

// Permission Role constants
const (
	RoleAdmin      = "admin"
	RoleInstructor = "instructor"
	RoleTA         = "ta"
	RoleStudent    = "student"
	RoleViewer     = "viewer"
)

// Course represents an educational course
type Course struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Code        string    `db:"code" json:"code"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	InstructorID uuid.UUID `db:"instructor_id" json:"instructor_id"`
	Department  string    `db:"department" json:"department"`
	Semester    string    `db:"semester" json:"semester"`
	Year        int       `db:"year" json:"year"`
	Status      string    `db:"status" json:"status"`
	MaxStudents int       `db:"max_students" json:"max_students"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// CourseEnrollment represents a student's enrollment in a course
type CourseEnrollment struct {
	ID             uuid.UUID `db:"id" json:"id"`
	CourseID       uuid.UUID `db:"course_id" json:"course_id"`
	StudentID      uuid.UUID `db:"student_id" json:"student_id"`
	EnrollmentDate time.Time `db:"enrollment_date" json:"enrollment_date"`
	Status         string    `db:"status" json:"status"`
}

// CoursePermission represents user permissions for a course
type CoursePermission struct {
	ID        uuid.UUID `db:"id" json:"id"`
	CourseID  uuid.UUID `db:"course_id" json:"course_id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	Role      string    `db:"role" json:"role"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// CourseRecording links a recording to a course
type CourseRecording struct {
	ID            uuid.UUID `db:"id" json:"id"`
	CourseID      uuid.UUID `db:"course_id" json:"course_id"`
	RecordingID   uuid.UUID `db:"recording_id" json:"recording_id"`
	LectureNumber int       `db:"lecture_number" json:"lecture_number"`
	LectureTitle  string    `db:"lecture_title" json:"lecture_title"`
	SequenceOrder int       `db:"sequence_order" json:"sequence_order"`
	IsPublished   bool      `db:"is_published" json:"is_published"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

// CourseActivity represents an audit log entry
type CourseActivity struct {
	ID           uuid.UUID            `db:"id" json:"id"`
	CourseID     uuid.UUID            `db:"course_id" json:"course_id"`
	UserID       *uuid.UUID           `db:"user_id" json:"user_id"`
	Action       string               `db:"action" json:"action"`
	ResourceType string               `db:"resource_type" json:"resource_type"`
	ResourceID   *uuid.UUID           `db:"resource_id" json:"resource_id"`
	Details      map[string]interface{} `db:"details" json:"details"`
	IPAddress    string               `db:"ip_address" json:"ip_address"`
	UserAgent    string               `db:"user_agent" json:"user_agent"`
	Timestamp    time.Time            `db:"timestamp" json:"timestamp"`
}

// RecordingAccessLog represents recording access tracking
type RecordingAccessLog struct {
	ID          uuid.UUID `db:"id" json:"id"`
	RecordingID uuid.UUID `db:"recording_id" json:"recording_id"`
	CourseID    *uuid.UUID `db:"course_id" json:"course_id"`
	UserID      *uuid.UUID `db:"user_id" json:"user_id"`
	Action      string    `db:"action" json:"action"`
	IPAddress   string    `db:"ip_address" json:"ip_address"`
	UserAgent   string    `db:"user_agent" json:"user_agent"`
	Timestamp   time.Time `db:"timestamp" json:"timestamp"`
}

// Request Types

// CreateCourseRequest for creating a new course
type CreateCourseRequest struct {
	Code        string `json:"code" validate:"required,max=50"`
	Name        string `json:"name" validate:"required,max=255"`
	Description string `json:"description" validate:"max=1000"`
	Department  string `json:"department" validate:"max=100"`
	Semester    string `json:"semester" validate:"required,max=20"`
	Year        int    `json:"year" validate:"required,min=2020,max=2100"`
	MaxStudents int    `json:"max_students" validate:"min=0"`
}

// UpdateCourseRequest for updating an existing course
type UpdateCourseRequest struct {
	Name        *string `json:"name" validate:"omitempty,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	Department  *string `json:"department" validate:"omitempty,max=100"`
	Semester    *string `json:"semester" validate:"omitempty,max=20"`
	Year        *int    `json:"year" validate:"omitempty,min=2020,max=2100"`
	MaxStudents *int    `json:"max_students" validate:"omitempty,min=0"`
	Status      *string `json:"status" validate:"omitempty,oneof=draft active archived completed"`
}

// EnrollStudentRequest for enrolling a student in a course
type EnrollStudentRequest struct {
	StudentID uuid.UUID `json:"student_id" validate:"required"`
}

// AddRecordingRequest for linking a recording to a course
type AddRecordingRequest struct {
	RecordingID  uuid.UUID `json:"recording_id" validate:"required"`
	LectureNumber *int     `json:"lecture_number" validate:"omitempty,min=1"`
	LectureTitle  *string  `json:"lecture_title" validate:"omitempty,max=255"`
	SequenceOrder int      `json:"sequence_order" validate:"required,min=0"`
}

// SetPermissionRequest for setting user permissions
type SetPermissionRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	Role   string    `json:"role" validate:"required,oneof=admin instructor ta student viewer"`
}

// Response Types

// CreateCourseResponse returns after creating a course
type CreateCourseResponse struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// CourseDetailResponse contains full course information
type CourseDetailResponse struct {
	ID              uuid.UUID `json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	InstructorID    uuid.UUID `json:"instructor_id"`
	Department      string    `json:"department"`
	Semester        string    `json:"semester"`
	Year            int       `json:"year"`
	Status          string    `json:"status"`
	MaxStudents     int       `json:"max_students"`
	EnrolledCount   int       `json:"enrolled_count"`
	RecordingCount  int       `json:"recording_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CourseListResponse for listing courses
type CourseListResponse struct {
	ID            uuid.UUID `json:"id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Semester      string    `json:"semester"`
	Year          int       `json:"year"`
	Status        string    `json:"status"`
	EnrolledCount int       `json:"enrolled_count"`
	CreatedAt     time.Time `json:"created_at"`
}

// EnrollmentListResponse for listing enrollments
type EnrollmentListResponse struct {
	ID             uuid.UUID `json:"id"`
	StudentID      uuid.UUID `json:"student_id"`
	EnrollmentDate time.Time `json:"enrollment_date"`
	Status         string    `json:"status"`
}

// CourseStatsResponse contains course statistics
type CourseStatsResponse struct {
	CourseID              uuid.UUID `json:"course_id"`
	CourseName            string    `json:"course_name"`
	TotalStudents         int       `json:"total_students"`
	TotalRecordings       int       `json:"total_recordings"`
	PublishedRecordings   int       `json:"published_recordings"`
	UniqueViewers         int       `json:"unique_viewers"`
	TotalViews            int       `json:"total_views"`
	LastAccessed          *time.Time `json:"last_accessed"`
	AverageEngagement     float64   `json:"average_engagement"`
}

// PermissionCheckResponse for checking user access
type PermissionCheckResponse struct {
	Allowed       bool   `json:"allowed"`
	Role          string `json:"role"`
	Message       string `json:"message"`
}

// ErrorResponse for API errors
type ErrorResponse struct {
	Error      string `json:"error"`
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Timestamp  time.Time `json:"timestamp"`
}
