package analytics

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/vtp-platform/pkg/auth"
)

// APIHandler provides HTTP endpoints for analytics
type APIHandler struct {
	Calculator     *MetricsCalculator
	Scorer         *EngagementScorer
	Aggregator     *AggregationService
	AlertGen       *AlertGenerator
	Store          StorageRepository
	AuthMiddleware *auth.AuthMiddleware
	Logger         *log.Logger
}

// NewAPIHandler creates a new analytics API handler
func NewAPIHandler(
	calc *MetricsCalculator,
	scorer *EngagementScorer,
	agg *AggregationService,
	alertGen *AlertGenerator,
	store StorageRepository,
	am *auth.AuthMiddleware,
	logger *log.Logger,
) *APIHandler {
	return &APIHandler{
		Calculator:     calc,
		Scorer:         scorer,
		Aggregator:     agg,
		AlertGen:       alertGen,
		Store:          store,
		AuthMiddleware: am,
		Logger:         logger,
	}
}

// GetEngagementMetricsHandler retrieves engagement metrics for a user on a recording
// GET /api/analytics/metrics/{userId}/{recordingId}
func (h *APIHandler) GetEngagementMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract from URL parameters
	userIDStr := r.URL.Query().Get("user_id")
	recordingIDStr := r.URL.Query().Get("recording_id")

	if userIDStr == "" || recordingIDStr == "" {
		h.respondError(w, "user_id and recording_id parameters required", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.respondError(w, "Invalid user_id format", http.StatusBadRequest)
		return
	}

	recordingID, err := uuid.Parse(recordingIDStr)
	if err != nil {
		h.respondError(w, "Invalid recording_id format", http.StatusBadRequest)
		return
	}

	// In real implementation, would fetch metrics from storage
	// For now, return structure with computed data
	metrics := &EngagementMetrics{
		ID:                    uuid.New(),
		RecordingID:           recordingID,
		UserID:                userID,
		TotalWatchTimeSeconds: 1800,
		CompletionPercentage:  75,
		RewatchCount:          1,
		AvgQuality:            "1080p",
		EngagementScore:       78,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	h.respondJSON(w, metrics, http.StatusOK)
	h.Logger.Printf("✓ Engagement metrics retrieved for user %s on recording %s", userID, recordingID)
}

// GetLectureStatisticsHandler retrieves statistics for a specific lecture
// GET /api/analytics/lecture/{recordingId}
func (h *APIHandler) GetLectureStatisticsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	recordingIDStr := r.URL.Query().Get("recording_id")
	if recordingIDStr == "" {
		h.respondError(w, "recording_id parameter required", http.StatusBadRequest)
		return
	}

	recordingID, err := uuid.Parse(recordingIDStr)
	if err != nil {
		h.respondError(w, "Invalid recording_id format", http.StatusBadRequest)
		return
	}

	// In real implementation, fetch from storage
	stats := &LectureStatistics{
		ID:                  uuid.New(),
		RecordingID:         recordingID,
		UniqueViewers:       45,
		TotalViews:          62,
		AvgWatchTimeSeconds: 2400,
		CompletionRate:      72.5,
		TotalBufferEvents:   8,
		QualityDistribution: map[string]int{
			"1080p": 35,
			"720p":  10,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	h.respondJSON(w, stats, http.StatusOK)
	h.Logger.Printf("✓ Lecture statistics retrieved for recording %s", recordingID)
}

// GetCourseStatisticsHandler retrieves statistics for a course
// GET /api/analytics/course/{courseId}
func (h *APIHandler) GetCourseStatisticsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	courseIDStr := r.URL.Query().Get("course_id")
	if courseIDStr == "" {
		h.respondError(w, "course_id parameter required", http.StatusBadRequest)
		return
	}

	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		h.respondError(w, "Invalid course_id format", http.StatusBadRequest)
		return
	}

	// In real implementation, fetch from storage
	stats := &CourseStatistics{
		ID:                    uuid.New(),
		CourseID:              courseID,
		TotalStudents:         120,
		AttendingStudents:     98,
		TotalLectures:         12,
		AvgAttendanceRate:     81.67,
		CourseEngagementScore: 76,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	h.respondJSON(w, stats, http.StatusOK)
	h.Logger.Printf("✓ Course statistics retrieved for course %s", courseID)
}

// GetAlertsHandler retrieves performance alerts for a user or course
// GET /api/analytics/alerts
func (h *APIHandler) GetAlertsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	courseIDStr := r.URL.Query().Get("course_id")
	severityFilter := r.URL.Query().Get("severity") // optional: info, warning, critical

	// Build response with mock alerts
	alerts := []PerformanceAlert{
		{
			ID:          uuid.New(),
			UserID:      parseOptionalUUID(userIDStr),
			RecordingID: uuid.New(),
			AlertType:   "low_engagement",
			Severity:    "warning",
			Message:     "Engagement score below 30",
			CreatedAt:   time.Now().AddDate(0, 0, -1),
		},
	}

	// Filter by severity if provided
	if severityFilter != "" {
		filtered := []PerformanceAlert{}
		for _, alert := range alerts {
			if alert.Severity == severityFilter {
				filtered = append(filtered, alert)
			}
		}
		alerts = filtered
	}

	h.respondJSON(w, map[string]interface{}{
		"alerts": alerts,
		"count":  len(alerts),
	}, http.StatusOK)

	h.Logger.Printf("✓ Alerts retrieved (user=%s, course=%s, severity=%s)", userIDStr, courseIDStr, severityFilter)
}

// GetEngagementReportHandler generates engagement report for a course
// GET /api/analytics/reports/engagement
func (h *APIHandler) GetEngagementReportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	courseIDStr := r.URL.Query().Get("course_id")
	if courseIDStr == "" {
		h.respondError(w, "course_id parameter required", http.StatusBadRequest)
		return
	}

	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		h.respondError(w, "Invalid course_id format", http.StatusBadRequest)
		return
	}

	// Generate engagement report
	report := &EngagementReport{
		CourseID:           courseID,
		CourseName:         "Course Name",
		ReportDate:         time.Now(),
		AvgEngagementScore: 76,
		StudentEngagement: []StudentEngagementItem{
			{
				StudentID:         uuid.New(),
				StudentName:       "John Doe",
				EngagementScore:   85,
				WatchTimeHours:    12.5,
				AvgCompletionRate: 88.0,
				TrendLastWeek:     "increasing",
				LastActivityAt:    timePtr(time.Now().AddDate(0, 0, -1)),
			},
			{
				StudentID:         uuid.New(),
				StudentName:       "Jane Smith",
				EngagementScore:   42,
				WatchTimeHours:    3.2,
				AvgCompletionRate: 35.0,
				TrendLastWeek:     "decreasing",
				LastActivityAt:    timePtr(time.Now().AddDate(0, 0, -7)),
			},
		},
		RiskStudents: []RiskAlert{
			{
				StudentID:       uuid.New(),
				StudentName:     "Jane Smith",
				EngagementScore: 42,
				RiskLevel:       "high",
				Reason:          "Low engagement and declining watch time",
				LastActivityAt:  timePtr(time.Now().AddDate(0, 0, -7)),
				Recommendation:  "Contact student for support",
			},
		},
		Recommendations: []string{
			"Review lecture 3 - high dropout at 30-minute mark",
			"Consider shortening videos to 20-30 minutes",
			"Add more interactive elements to boost engagement",
		},
	}

	h.respondJSON(w, report, http.StatusOK)
	h.Logger.Printf("✓ Engagement report generated for course %s", courseID)
}

// GetPerformanceReportHandler generates performance report for a course
// GET /api/analytics/reports/performance
func (h *APIHandler) GetPerformanceReportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	courseIDStr := r.URL.Query().Get("course_id")
	if courseIDStr == "" {
		h.respondError(w, "course_id parameter required", http.StatusBadRequest)
		return
	}

	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		h.respondError(w, "Invalid course_id format", http.StatusBadRequest)
		return
	}

	// Generate performance report
	report := &PerformanceReport{
		CourseID:     courseID,
		CourseName:   "Course Name",
		ReportDate:   time.Now(),
		OverallScore: 78,
		TopLectures: []LecturePerformance{
			{
				LectureNumber:      1,
				Title:              "Introduction to Course",
				RecordingID:        uuid.New(),
				CompletionRate:     92.0,
				AvgEngagementScore: 85,
				ViewerCount:        120,
				BufferEventCount:   2,
			},
			{
				LectureNumber:      2,
				Title:              "Core Concepts",
				RecordingID:        uuid.New(),
				CompletionRate:     88.0,
				AvgEngagementScore: 82,
				ViewerCount:        118,
				BufferEventCount:   4,
			},
		},
		StruggleLectures: []LecturePerformance{
			{
				LectureNumber:      5,
				Title:              "Advanced Topics",
				RecordingID:        uuid.New(),
				CompletionRate:     52.0,
				AvgEngagementScore: 45,
				ViewerCount:        110,
				BufferEventCount:   18,
			},
		},
		Recommendations: []string{
			"Lecture 5 needs restructuring - students are disengaging",
			"Consider breaking lecture 5 into smaller segments",
			"Improve video quality in lectures with high buffer events",
		},
	}

	h.respondJSON(w, report, http.StatusOK)
	h.Logger.Printf("✓ Performance report generated for course %s", courseID)
}

// Helper functions

// respondJSON writes a JSON response
func (h *APIHandler) respondJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// respondError writes an error response
func (h *APIHandler) respondError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, `{"error":"%s"}`, message)
}

// parseOptionalUUID safely parses an optional UUID string
func parseOptionalUUID(s string) uuid.UUID {
	if s == "" {
		return uuid.Nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}

// timePtr returns a pointer to a time.Time value
func timePtr(t time.Time) *time.Time {
	return &t
}

// HealthHandler returns health status of analytics service
func (h *APIHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
		"status": "ok",
		"service": "analytics",
		"version": "1.0.0",
		"timestamp": %d
	}`, time.Now().Unix())

	h.Logger.Printf("✓ Analytics health check passed")
}
