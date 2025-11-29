package models

import "time"

// User represents a platform user (student, teacher, admin)
type User struct {
	ID           string    `db:"id" json:"id"`
	Email        string    `db:"email" json:"email"`
	Phone        string    `db:"phone" json:"phone"`
	FullName     string    `db:"full_name" json:"full_name"`
	Role         string    `db:"role" json:"role"` // student, teacher, admin
	PasswordHash string    `db:"password_hash" json:"-"`
	Locale       string    `db:"locale" json:"locale"` // ar_SY
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// Course represents an educational course
type Course struct {
	ID            string    `db:"id" json:"id"`
	TeacherID     string    `db:"teacher_id" json:"teacher_id"`
	TitleAr       string    `db:"title_ar" json:"title_ar"`
	DescriptionAr string    `db:"description_ar" json:"description_ar"`
	Syllabus      string    `db:"syllabus" json:"syllabus"`
	Visibility    string    `db:"visibility" json:"visibility"` // public, private
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

// Lesson represents a lesson within a course
type Lesson struct {
	ID        string    `db:"id" json:"id"`
	CourseID  string    `db:"course_id" json:"course_id"`
	Type      string    `db:"type" json:"type"` // video, live, doc
	TitleAr   string    `db:"title_ar" json:"title_ar"`
	MediaRefs string    `db:"media_refs" json:"media_refs"` // JSON array of file references
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// LiveSession represents a live classroom session
type LiveSession struct {
	ID            string     `db:"id" json:"id"`
	LessonID      string     `db:"lesson_id" json:"lesson_id"`
	SFURoomID     string     `db:"sfu_room_id" json:"sfu_room_id"`
	StartTime     time.Time  `db:"start_time" json:"start_time"`
	EndTime       *time.Time `db:"end_time" json:"end_time"`
	RecordingRefs string     `db:"recording_refs" json:"recording_refs"` // JSON array
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

// Recording represents a recorded video file
type Recording struct {
	ID           string    `db:"id" json:"id"`
	SessionID    string    `db:"session_id" json:"session_id"`
	S3URL        string    `db:"s3_url" json:"s3_url"`
	Duration     int       `db:"duration" json:"duration"`           // seconds
	FileSize     int64     `db:"file_size" json:"file_size"`         // bytes
	Format       string    `db:"format" json:"format"`               // mp4, hls
	TranscriptAr string    `db:"transcript_ar" json:"transcript_ar"` // Arabic transcript
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// Assignment represents a course assignment
type Assignment struct {
	ID            string    `db:"id" json:"id"`
	LessonID      string    `db:"lesson_id" json:"lesson_id"`
	TitleAr       string    `db:"title_ar" json:"title_ar"`
	DescriptionAr string    `db:"description_ar" json:"description_ar"`
	DueDate       time.Time `db:"due_date" json:"due_date"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

// Submission represents a student's assignment submission
type Submission struct {
	ID           string    `db:"id" json:"id"`
	AssignmentID string    `db:"assignment_id" json:"assignment_id"`
	StudentID    string    `db:"student_id" json:"student_id"`
	FileRefs     string    `db:"file_refs" json:"file_refs"` // JSON array
	Grade        *float32  `db:"grade" json:"grade"`
	Feedback     string    `db:"feedback" json:"feedback"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// Chat represents a message in a class chat
type Chat struct {
	ID        string    `db:"id" json:"id"`
	RoomID    string    `db:"room_id" json:"room_id"` // session or course room
	SenderID  string    `db:"sender_id" json:"sender_id"`
	Content   string    `db:"content" json:"content"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// Instructor represents a certified instructor/teacher
type Instructor struct {
	ID               string    `db:"id" json:"id"`
	UserID           string    `db:"user_id" json:"user_id"` // Link to User table
	NameAr           string    `db:"name_ar" json:"name_ar"`
	BioAr            string    `db:"bio_ar" json:"bio_ar"`
	Specialization   string    `db:"specialization" json:"specialization"` // JSON array of subject IDs
	HourlyRate       float64   `db:"hourly_rate" json:"hourly_rate"`       // USD
	Rating           float32   `db:"rating" json:"rating"`
	TotalReviews     int       `db:"total_reviews" json:"total_reviews"`
	YearsExperience  int       `db:"years_experience" json:"years_experience"`
	CertificationsAr string    `db:"certifications_ar" json:"certifications_ar"` // JSON array
	Availability     string    `db:"availability" json:"availability"`           // JSON object {day: [time slots]}
	IsVerified       bool      `db:"is_verified" json:"is_verified"`
	IsActive         bool      `db:"is_active" json:"is_active"`
	ProfileImageURL  string    `db:"profile_image_url" json:"profile_image_url"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

// Subject represents an academic subject
type Subject struct {
	ID        string    `db:"id" json:"id"`
	NameAr    string    `db:"name_ar" json:"name_ar"`
	Level     string    `db:"level" json:"level"` // elementary, intermediate, advanced
	Category  string    `db:"category" json:"category"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// Meeting represents a scheduled one-on-one or group session
type Meeting struct {
	ID           string     `db:"id" json:"id"`
	InstructorID string     `db:"instructor_id" json:"instructor_id"`
	StudentID    string     `db:"student_id" json:"student_id"` // null for group meetings
	SubjectID    string     `db:"subject_id" json:"subject_id"`
	TitleAr      string     `db:"title_ar" json:"title_ar"`
	ScheduledAt  time.Time  `db:"scheduled_at" json:"scheduled_at"`
	Duration     int        `db:"duration" json:"duration"` // minutes
	MeetingURL   string     `db:"meeting_url" json:"meeting_url"`
	RoomID       string     `db:"room_id" json:"room_id"`
	Status       string     `db:"status" json:"status"` // scheduled, completed, cancelled
	EndTime      *time.Time `db:"end_time" json:"end_time"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

// StudyMaterial represents supplementary learning materials
type StudyMaterial struct {
	ID           string    `db:"id" json:"id"`
	CourseID     string    `db:"course_id" json:"course_id"`
	InstructorID string    `db:"instructor_id" json:"instructor_id"`
	TitleAr      string    `db:"title_ar" json:"title_ar"`
	Type         string    `db:"type" json:"type"` // pdf, slides, notes, worksheet
	FileURL      string    `db:"file_url" json:"file_url"`
	FileSize     int64     `db:"file_size" json:"file_size"`
	Downloads    int       `db:"downloads" json:"downloads"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}
