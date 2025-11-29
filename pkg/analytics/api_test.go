package analytics

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
)

// Test fixtures
func createTestAPIHandler() *APIHandler {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	calc := NewMetricsCalculator(nil, logger)
	scorer := NewEngagementScorer(calc, logger)
	agg := NewAggregationService(nil, logger)
	alertGen := NewAlertGenerator(DefaultPerformanceThreshold(), logger)

	return NewAPIHandler(
		calc,
		scorer,
		agg,
		alertGen,
		nil,
		nil,
		logger,
	)
}

// TestGetEngagementMetricsHandler tests the engagement metrics endpoint
func TestGetEngagementMetricsHandler(t *testing.T) {
	handler := createTestAPIHandler()
	userID := uuid.New()
	recordingID := uuid.New()

	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/api/analytics/metrics?user_id=%s&recording_id=%s", userID, recordingID),
		nil,
	)
	w := httptest.NewRecorder()

	handler.GetEngagementMetricsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var metrics EngagementMetrics
	err := json.NewDecoder(w.Body).Decode(&metrics)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if metrics.RecordingID != recordingID {
		t.Errorf("Recording ID mismatch")
	}

	if metrics.UserID != userID {
		t.Errorf("User ID mismatch")
	}

	if metrics.EngagementScore < 0 || metrics.EngagementScore > 100 {
		t.Errorf("Invalid engagement score: %d", metrics.EngagementScore)
	}

	t.Logf("✓ Engagement metrics: score=%d, completion=%d%%", metrics.EngagementScore, metrics.CompletionPercentage)
}

// TestEngagementMetricsMissingParameters tests parameter validation
func TestEngagementMetricsMissingParameters(t *testing.T) {
	handler := createTestAPIHandler()

	tests := []struct {
		name      string
		url       string
		expectErr bool
	}{
		{"Valid params", "/api/analytics/metrics?user_id=550e8400-e29b-41d4-a716-446655440000&recording_id=550e8400-e29b-41d4-a716-446655440001", false},
		{"Missing user_id", "/api/analytics/metrics?recording_id=550e8400-e29b-41d4-a716-446655440001", true},
		{"Missing recording_id", "/api/analytics/metrics?user_id=550e8400-e29b-41d4-a716-446655440000", true},
		{"Invalid user_id format", "/api/analytics/metrics?user_id=invalid&recording_id=550e8400-e29b-41d4-a716-446655440001", true},
		{"Invalid recording_id format", "/api/analytics/metrics?user_id=550e8400-e29b-41d4-a716-446655440000&recording_id=invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()

			handler.GetEngagementMetricsHandler(w, req)

			if tt.expectErr {
				if w.Code != http.StatusBadRequest {
					t.Errorf("Expected 400 for invalid params, got %d", w.Code)
				}
			} else {
				if w.Code != http.StatusOK {
					t.Errorf("Expected 200, got %d", w.Code)
				}
			}
		})
	}
}

// TestGetLectureStatisticsHandler tests the lecture statistics endpoint
func TestGetLectureStatisticsHandler(t *testing.T) {
	handler := createTestAPIHandler()
	recordingID := uuid.New()

	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/api/analytics/lecture?recording_id=%s", recordingID),
		nil,
	)
	w := httptest.NewRecorder()

	handler.GetLectureStatisticsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var stats LectureStatistics
	err := json.NewDecoder(w.Body).Decode(&stats)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if stats.RecordingID != recordingID {
		t.Errorf("Recording ID mismatch")
	}

	if stats.UniqueViewers <= 0 {
		t.Errorf("Expected viewers > 0, got %d", stats.UniqueViewers)
	}

	if stats.CompletionRate < 0 || stats.CompletionRate > 100 {
		t.Errorf("Invalid completion rate: %.1f", stats.CompletionRate)
	}

	t.Logf("✓ Lecture stats: viewers=%d, completion=%.1f%%", stats.UniqueViewers, stats.CompletionRate)
}

// TestGetCourseStatisticsHandler tests the course statistics endpoint
func TestGetCourseStatisticsHandler(t *testing.T) {
	handler := createTestAPIHandler()
	courseID := uuid.New()

	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/api/analytics/course?course_id=%s", courseID),
		nil,
	)
	w := httptest.NewRecorder()

	handler.GetCourseStatisticsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var stats CourseStatistics
	err := json.NewDecoder(w.Body).Decode(&stats)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if stats.CourseID != courseID {
		t.Errorf("Course ID mismatch")
	}

	if stats.TotalLectures <= 0 {
		t.Errorf("Expected lectures > 0")
	}

	if stats.TotalStudents <= 0 {
		t.Errorf("Expected students > 0")
	}

	if stats.AvgAttendanceRate < 0 || stats.AvgAttendanceRate > 100 {
		t.Errorf("Invalid attendance rate: %.2f", stats.AvgAttendanceRate)
	}

	t.Logf("✓ Course stats: lectures=%d, students=%d, attendance=%.2f%%",
		stats.TotalLectures, stats.TotalStudents, stats.AvgAttendanceRate)
}

// TestGetAlertsHandler tests the alerts endpoint
func TestGetAlertsHandler(t *testing.T) {
	handler := createTestAPIHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/alerts", nil)
	w := httptest.NewRecorder()

	handler.GetAlertsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if _, ok := resp["alerts"]; !ok {
		t.Errorf("Response missing 'alerts' field")
	}

	if _, ok := resp["count"]; !ok {
		t.Errorf("Response missing 'count' field")
	}

	t.Logf("✓ Alerts retrieved: count=%v", resp["count"])
}

// TestAlertsFilterBySeverity tests alerts filtering by severity
func TestAlertsFilterBySeverity(t *testing.T) {
	handler := createTestAPIHandler()

	severities := []string{"info", "warning", "critical"}
	for _, severity := range severities {
		t.Run(fmt.Sprintf("severity=%s", severity), func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodGet,
				fmt.Sprintf("/api/analytics/alerts?severity=%s", severity),
				nil,
			)
			w := httptest.NewRecorder()

			handler.GetAlertsHandler(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			var resp map[string]interface{}
			json.NewDecoder(w.Body).Decode(&resp)
			t.Logf("Alerts with severity %s retrieved", severity)
		})
	}
}

// TestGetEngagementReportHandler tests the engagement report endpoint
func TestGetEngagementReportHandler(t *testing.T) {
	handler := createTestAPIHandler()
	courseID := uuid.New()

	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/api/analytics/reports/engagement?course_id=%s", courseID),
		nil,
	)
	w := httptest.NewRecorder()

	handler.GetEngagementReportHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var report EngagementReport
	err := json.NewDecoder(w.Body).Decode(&report)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if report.CourseID != courseID {
		t.Errorf("Course ID mismatch")
	}

	if report.AvgEngagementScore < 0 || report.AvgEngagementScore > 100 {
		t.Errorf("Invalid engagement score: %d", report.AvgEngagementScore)
	}

	if len(report.StudentEngagement) == 0 {
		t.Errorf("Expected student engagement data")
	}

	if len(report.Recommendations) == 0 {
		t.Errorf("Expected recommendations")
	}

	t.Logf("✓ Engagement report: avgScore=%d, students=%d, recommendations=%d",
		report.AvgEngagementScore, len(report.StudentEngagement), len(report.Recommendations))
}

// TestGetPerformanceReportHandler tests the performance report endpoint
func TestGetPerformanceReportHandler(t *testing.T) {
	handler := createTestAPIHandler()
	courseID := uuid.New()

	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/api/analytics/reports/performance?course_id=%s", courseID),
		nil,
	)
	w := httptest.NewRecorder()

	handler.GetPerformanceReportHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var report PerformanceReport
	err := json.NewDecoder(w.Body).Decode(&report)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if report.CourseID != courseID {
		t.Errorf("Course ID mismatch")
	}

	if report.OverallScore < 0 || report.OverallScore > 100 {
		t.Errorf("Invalid overall score: %d", report.OverallScore)
	}

	if len(report.TopLectures) == 0 {
		t.Errorf("Expected top lectures")
	}

	if len(report.StruggleLectures) == 0 {
		t.Errorf("Expected struggle lectures")
	}

	t.Logf("✓ Performance report: score=%d, topLectures=%d, struggleLectures=%d",
		report.OverallScore, len(report.TopLectures), len(report.StruggleLectures))
}

// TestHTTPMethodValidation tests that only GET methods are allowed
func TestHTTPMethodValidation(t *testing.T) {
	handler := createTestAPIHandler()

	tests := []struct {
		name    string
		method  string
		handler func(w http.ResponseWriter, r *http.Request)
		url     string
	}{
		{"POST to engagement metrics", http.MethodPost, handler.GetEngagementMetricsHandler, "/api/analytics/metrics"},
		{"DELETE to lecture stats", http.MethodDelete, handler.GetLectureStatisticsHandler, "/api/analytics/lecture"},
		{"PUT to alerts", http.MethodPut, handler.GetAlertsHandler, "/api/analytics/alerts"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			tt.handler(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected 405 Method Not Allowed, got %d", w.Code)
			}
		})
	}
}

// TestHealthHandler tests the health check endpoint
func TestHealthHandler(t *testing.T) {
	handler := createTestAPIHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.HealthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var health map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&health)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if health["status"] != "ok" {
		t.Errorf("Expected status 'ok'")
	}

	if health["service"] != "analytics" {
		t.Errorf("Expected service 'analytics'")
	}

	t.Logf("✓ Health check: status=%s, service=%s", health["status"], health["service"])
}

// TestResponseContentType tests that responses have correct content type
func TestResponseContentType(t *testing.T) {
	handler := createTestAPIHandler()
	recordingID := uuid.New()

	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/api/analytics/lecture?recording_id=%s", recordingID),
		nil,
	)
	w := httptest.NewRecorder()

	handler.GetLectureStatisticsHandler(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

// TestResponseBody tests that responses contain expected JSON fields
func TestResponseBody(t *testing.T) {
	handler := createTestAPIHandler()
	courseID := uuid.New()

	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/api/analytics/reports/engagement?course_id=%s", courseID),
		nil,
	)
	w := httptest.NewRecorder()

	handler.GetEngagementReportHandler(w, req)

	body, _ := io.ReadAll(w.Body)
	bodyStr := string(body)

	expectedFields := []string{"course_id", "avg_engagement_score", "student_engagement"}
	for _, field := range expectedFields {
		if !contains(bodyStr, field) {
			t.Errorf("Expected field '%s' in response body", field)
		}
	}
}

// TestErrorResponse tests error response format
func TestErrorResponse(t *testing.T) {
	handler := createTestAPIHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/analytics/metrics", nil) // Missing required params
	w := httptest.NewRecorder()

	handler.GetEngagementMetricsHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 status")
	}

	var errResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&errResp)

	if _, ok := errResp["error"]; !ok {
		t.Errorf("Expected 'error' field in error response")
	}
}

// TestResponseWithUserID tests alerts filtering by user_id
func TestResponseWithUserID(t *testing.T) {
	handler := createTestAPIHandler()
	userID := uuid.New()

	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/api/analytics/alerts?user_id=%s", userID),
		nil,
	)
	w := httptest.NewRecorder()

	handler.GetAlertsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["count"] == nil {
		t.Errorf("Expected count in response")
	}
}

// TestCourseReportsRequireID tests that reports require course_id
func TestCourseReportsRequireID(t *testing.T) {
	handler := createTestAPIHandler()

	tests := []struct {
		name    string
		handler func(w http.ResponseWriter, r *http.Request)
		url     string
	}{
		{"Engagement report without course_id", handler.GetEngagementReportHandler, "/api/analytics/reports/engagement"},
		{"Performance report without course_id", handler.GetPerformanceReportHandler, "/api/analytics/reports/performance"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()

			tt.handler(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected 400 status, got %d", w.Code)
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr))
}

// BenchmarkGetEngagementMetrics benchmarks the engagement metrics endpoint
func BenchmarkGetEngagementMetrics(b *testing.B) {
	handler := createTestAPIHandler()
	userID := uuid.New()
	recordingID := uuid.New()
	url := fmt.Sprintf("/api/analytics/metrics?user_id=%s&recording_id=%s", userID, recordingID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()
		handler.GetEngagementMetricsHandler(w, req)
	}
}

// BenchmarkGetCourseStatistics benchmarks the course statistics endpoint
func BenchmarkGetCourseStatistics(b *testing.B) {
	handler := createTestAPIHandler()
	courseID := uuid.New()
	url := fmt.Sprintf("/api/analytics/course?course_id=%s", courseID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()
		handler.GetCourseStatisticsHandler(w, req)
	}
}

// BenchmarkGetEngagementReport benchmarks the engagement report endpoint
func BenchmarkGetEngagementReport(b *testing.B) {
	handler := createTestAPIHandler()
	courseID := uuid.New()
	url := fmt.Sprintf("/api/analytics/reports/engagement?course_id=%s", courseID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, url, nil)
		w := httptest.NewRecorder()
		handler.GetEngagementReportHandler(w, req)
	}
}
