package analytics

import (
	"time"

	"github.com/google/uuid"
)

// Event Types
type EventType string

const (
	EventRecordingStarted EventType = "recording_started"
	EventRecordingStopped EventType = "recording_stopped"
	EventPlaybackStarted  EventType = "playback_started"
	EventPlaybackStopped  EventType = "playback_stopped"
	EventPlaybackPaused   EventType = "playback_paused"
	EventPlaybackResumed  EventType = "playback_resumed"
	EventSegmentRequested EventType = "segment_requested"
	EventQualitySelected  EventType = "quality_selected"
	EventQualityChanged   EventType = "quality_changed"
	EventBufferEvent      EventType = "buffer_event"
	EventSessionEnded     EventType = "session_ended"
)

// AnalyticsEvent represents a raw event from the streaming system
type AnalyticsEvent struct {
	EventID     uuid.UUID              `json:"event_id"`
	EventType   EventType              `json:"event_type"`
	RecordingID uuid.UUID              `json:"recording_id"`
	UserID      uuid.UUID              `json:"user_id"`
	SessionID   string                 `json:"session_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// PlaybackSession represents a user watching a lecture
type PlaybackSession struct {
	ID                     uuid.UUID  `json:"id"`
	RecordingID            uuid.UUID  `json:"recording_id"`
	UserID                 uuid.UUID  `json:"user_id"`
	SessionStart           time.Time  `json:"session_start"`
	SessionEnd             *time.Time `json:"session_end,omitempty"`
	TotalDurationSeconds   int        `json:"total_duration_seconds"`
	WatchedDurationSeconds int        `json:"watched_duration_seconds"`
	PauseCount             int        `json:"pause_count"`
	ResumeCount            int        `json:"resume_count"`
	QualitySelected        string     `json:"quality_selected"`
	BufferEvents           int        `json:"buffer_events"`
	CompletionRate         float64    `json:"completion_rate"` // 0-100
	CreatedAt              time.Time  `json:"created_at"`
}

// QualityEvent represents a quality change during playback
type QualityEvent struct {
	ID         uuid.UUID `json:"id"`
	SessionID  uuid.UUID `json:"session_id"`
	Timestamp  time.Time `json:"timestamp"`
	Bitrate    int       `json:"bitrate"`    // kbps
	Resolution string    `json:"resolution"` // 720p, 1080p, etc
	Reason     string    `json:"reason"`     // user_selected, auto_downgrade, auto_upgrade
	CreatedAt  time.Time `json:"created_at"`
}

// EngagementMetrics represents aggregate engagement for a user on a lecture
type EngagementMetrics struct {
	ID                    uuid.UUID  `json:"id"`
	RecordingID           uuid.UUID  `json:"recording_id"`
	UserID                uuid.UUID  `json:"user_id"`
	TotalWatchTimeSeconds int        `json:"total_watch_time_seconds"`
	CompletionPercentage  int        `json:"completion_percentage"` // 0-100
	RewatchCount          int        `json:"rewatch_count"`
	AvgQuality            string     `json:"avg_quality"`
	LastWatched           *time.Time `json:"last_watched,omitempty"`
	EngagementScore       int        `json:"engagement_score"` // 0-100
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// LectureStatistics represents aggregate stats for a lecture
type LectureStatistics struct {
	ID                    uuid.UUID      `json:"id"`
	RecordingID           uuid.UUID      `json:"recording_id"`
	UniqueViewers         int            `json:"unique_viewers"`
	TotalViews            int            `json:"total_views"`
	AvgWatchTimeSeconds   int            `json:"avg_watch_time_seconds"`
	CompletionRate        float64        `json:"completion_rate"` // 0-100
	PeakConcurrentViewers int            `json:"peak_concurrent_viewers"`
	TotalBufferEvents     int            `json:"total_buffer_events"`
	QualityDistribution   map[string]int `json:"quality_distribution"` // bitrate -> count
	MostReplayedSection   string         `json:"most_replayed_section,omitempty"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
}

// CourseStatistics represents aggregate stats for a course
type CourseStatistics struct {
	ID                    uuid.UUID `json:"id"`
	CourseID              uuid.UUID `json:"course_id"`
	TotalStudents         int       `json:"total_students"`
	AttendingStudents     int       `json:"attending_students"`
	TotalLectures         int       `json:"total_lectures"`
	TotalViewSessions     int       `json:"total_view_sessions"`
	AvgAttendanceRate     float64   `json:"avg_attendance_rate"` // 0-100
	TotalWatchTimeSeconds int64     `json:"total_watch_time_seconds"`
	CourseEngagementScore int       `json:"course_engagement_score"` // 0-100
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// StudentEngagementReport represents comprehensive student analytics
type StudentEngagementReport struct {
	StudentID             uuid.UUID          `json:"student_id"`
	StudentName           string             `json:"student_name"`
	StudentEmail          string             `json:"student_email"`
	CourseID              uuid.UUID          `json:"course_id"`
	LecturesWatched       int                `json:"lectures_watched"`
	LecturesTotal         int                `json:"lectures_total"`
	WatchRate             float64            `json:"watch_rate"` // percentage
	TotalWatchTimeMinutes int                `json:"total_watch_time_minutes"`
	AvgCompletionPercent  float64            `json:"avg_completion_percent"`
	EngagementScore       int                `json:"engagement_score"` // 0-100
	QualityPreference     string             `json:"quality_preference"`
	BufferEvents          int                `json:"buffer_events"`
	LastWatchedAt         *time.Time         `json:"last_watched_at,omitempty"`
	WatchHistory          []WatchHistoryItem `json:"watch_history,omitempty"`
	Alerts                []string           `json:"alerts,omitempty"`
}

// WatchHistoryItem represents a single lecture watched
type WatchHistoryItem struct {
	LectureNumber       int       `json:"lecture_number"`
	RecordingID         uuid.UUID `json:"recording_id"`
	WatchSessions       int       `json:"watch_sessions"`
	TotalWatchedMinutes int       `json:"total_watched_minutes"`
	CompletionPercent   float64   `json:"completion_percent"`
	LastWatchedAt       time.Time `json:"last_watched_at"`
	QualityPreference   string    `json:"quality_preference"`
}

// AttendanceReport shows attendance patterns
type AttendanceReport struct {
	CourseID          uuid.UUID               `json:"course_id"`
	CourseName        string                  `json:"course_name"`
	ReportDate        time.Time               `json:"report_date"`
	TotalStudents     int                     `json:"total_students"`
	AttendingStudents int                     `json:"attending_students"`
	AttendanceRate    float64                 `json:"attendance_rate"` // percentage
	StudentAttendance []StudentAttendanceItem `json:"student_attendance,omitempty"`
	AbsenceAlerts     []string                `json:"absence_alerts,omitempty"`
}

// StudentAttendanceItem represents attendance for one student
type StudentAttendanceItem struct {
	StudentID         uuid.UUID  `json:"student_id"`
	StudentName       string     `json:"student_name"`
	StudentEmail      string     `json:"student_email"`
	LecturesAttended  int        `json:"lectures_attended"`
	LecturesTotal     int        `json:"lectures_total"`
	AttendanceRate    float64    `json:"attendance_rate"`    // percentage
	ConsecutiveAbsent int        `json:"consecutive_absent"` // number of lectures missed
	LastAttendedAt    *time.Time `json:"last_attended_at,omitempty"`
}

// EngagementReport shows engagement metrics for a course
type EngagementReport struct {
	CourseID           uuid.UUID               `json:"course_id"`
	CourseName         string                  `json:"course_name"`
	ReportDate         time.Time               `json:"report_date"`
	AvgEngagementScore int                     `json:"avg_engagement_score"` // 0-100
	StudentEngagement  []StudentEngagementItem `json:"student_engagement,omitempty"`
	RiskStudents       []RiskAlert             `json:"risk_students,omitempty"`
	Recommendations    []string                `json:"recommendations,omitempty"`
}

// StudentEngagementItem represents engagement metrics for one student
type StudentEngagementItem struct {
	StudentID         uuid.UUID  `json:"student_id"`
	StudentName       string     `json:"student_name"`
	EngagementScore   int        `json:"engagement_score"` // 0-100
	WatchTimeHours    float64    `json:"watch_time_hours"`
	AvgCompletionRate float64    `json:"avg_completion_rate"` // percentage
	TrendLastWeek     string     `json:"trend_last_week"`     // "increasing", "decreasing", "stable"
	LastActivityAt    *time.Time `json:"last_activity_at,omitempty"`
}

// RiskAlert represents a student at risk
type RiskAlert struct {
	StudentID       uuid.UUID  `json:"student_id"`
	StudentName     string     `json:"student_name"`
	EngagementScore int        `json:"engagement_score"` // 0-100
	RiskLevel       string     `json:"risk_level"`       // "low", "medium", "high", "critical"
	Reason          string     `json:"reason"`           // why they're at risk
	LastActivityAt  *time.Time `json:"last_activity_at,omitempty"`
	Recommendation  string     `json:"recommendation"` // what to do
}

// PerformanceReport shows course performance metrics
type PerformanceReport struct {
	CourseID           uuid.UUID            `json:"course_id"`
	CourseName         string               `json:"course_name"`
	ReportDate         time.Time            `json:"report_date"`
	OverallScore       int                  `json:"overall_score"` // 0-100
	TopLectures        []LecturePerformance `json:"top_lectures,omitempty"`
	StruggleLectures   []LecturePerformance `json:"struggle_lectures,omitempty"`
	Recommendations    []string             `json:"recommendations,omitempty"`
	CurriculumInsights string               `json:"curriculum_insights,omitempty"`
}

// LecturePerformance represents performance metrics for a lecture
type LecturePerformance struct {
	LectureNumber     int       `json:"lecture_number"`
	Title             string    `json:"title"`
	RecordingID       uuid.UUID `json:"recording_id"`
	ViewCount         int       `json:"view_count"`
	AvgCompletionRate float64   `json:"avg_completion_rate"` // percentage
	EngagementScore   int       `json:"engagement_score"`    // 0-100
	BufferEvents      int       `json:"buffer_events"`
	ReplayCount       int       `json:"replay_count"`
	PerformanceRating string    `json:"performance_rating"` // "excellent", "good", "average", "poor"
}

// EventCollector collects and batches events
type EventCollector struct {
	events    []AnalyticsEvent
	maxSize   int
	batchTime time.Duration
}

// StorageRepository handles persistence of analytics data
type StorageRepository interface {
	StoreEvent(event AnalyticsEvent) error
	StoreEvents(events []AnalyticsEvent) error
	StorePlaybackSession(session PlaybackSession) error
	UpdatePlaybackSession(session PlaybackSession) error
	StoreQualityEvent(event QualityEvent) error
	StoreEngagementMetrics(metrics EngagementMetrics) error
	UpdateEngagementMetrics(metrics EngagementMetrics) error
	StoreLectureStatistics(stats LectureStatistics) error
	UpdateLectureStatistics(stats LectureStatistics) error
	StoreCourseStatistics(stats CourseStatistics) error
	UpdateCourseStatistics(stats CourseStatistics) error
}

// QueryRepository handles retrieving analytics data
type QueryRepository interface {
	GetPlaybackSession(sessionID uuid.UUID) (*PlaybackSession, error)
	GetUserPlaybackSessions(userID, recordingID uuid.UUID) ([]PlaybackSession, error)
	GetEngagementMetrics(recordingID, userID uuid.UUID) (*EngagementMetrics, error)
	GetLectureStatistics(recordingID uuid.UUID) (*LectureStatistics, error)
	GetCourseStatistics(courseID uuid.UUID) (*CourseStatistics, error)
	QueryPlaybackSessions(filters map[string]interface{}) ([]PlaybackSession, error)
}

// AnalyticsService combines storage and retrieval
type AnalyticsService interface {
	StorageRepository
	QueryRepository
	CollectEvent(event AnalyticsEvent) error
	GetStudentEngagementReport(studentID, courseID uuid.UUID) (*StudentEngagementReport, error)
	GetAttendanceReport(courseID uuid.UUID, startDate, endDate time.Time) (*AttendanceReport, error)
	GetEngagementReportByCourse(courseID uuid.UUID) (*EngagementReport, error)
	GetPerformanceReport(courseID uuid.UUID) (*PerformanceReport, error)
}
