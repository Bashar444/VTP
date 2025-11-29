package analytics

import (
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Test fixtures
func createTestDB() *sql.DB {
	// For testing, use in-memory database or mock
	// In real implementation, would use test database
	return nil
}

// TestStreamingEventListener tests the streaming event listener
func TestStreamingEventListener(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	collector := NewEventCollector(100, 5*time.Second, logger)
	calculator := NewMetricsCalculator(nil, logger)
	listener := NewStreamingEventListener(collector, calculator, nil, logger)

	recordingID := uuid.New()
	userID := uuid.New()
	sessionID := "session-001"

	// Test playback started
	err := listener.OnPlaybackStarted(recordingID, userID, sessionID)
	if err != nil {
		t.Errorf("Failed to handle playback start: %v", err)
	}

	// Test quality changed
	err = listener.OnQualityChanged(sessionID, "auto", "1080p", "user_selected")
	if err != nil {
		t.Errorf("Failed to handle quality change: %v", err)
	}

	// Test buffer event
	err = listener.OnBufferEvent(sessionID)
	if err != nil {
		t.Errorf("Failed to handle buffer event: %v", err)
	}

	t.Log("✓ Streaming event listener test passed")
}

// TestPlaybackSessionTracking tests session tracking
func TestPlaybackSessionTracking(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	collector := NewEventCollector(100, 5*time.Second, logger)
	calculator := NewMetricsCalculator(nil, logger)
	listener := NewStreamingEventListener(collector, calculator, nil, logger)

	recordingID := uuid.New()
	userID := uuid.New()
	sessionID := "session-002"

	// Start playback
	listener.OnPlaybackStarted(recordingID, userID, sessionID)

	// Verify session was created
	if len(listener.activeSessions) != 1 {
		t.Errorf("Expected 1 active session, got %d", len(listener.activeSessions))
	}

	t.Log("✓ Session tracking test passed")
}

// TestReportGenerator tests report generation
func TestReportGenerator(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	calculator := NewMetricsCalculator(nil, logger)
	reporter := NewReportGenerator(nil, calculator, logger)

	courseID := uuid.New()

	// Generate engagement report
	report, err := reporter.GenerateCourseEngagementReport(courseID)
	if err != nil {
		t.Errorf("Failed to generate engagement report: %v", err)
	}

	if report.CourseID != courseID {
		t.Errorf("Course ID mismatch in report")
	}

	if report.AvgEngagementScore < 0 || report.AvgEngagementScore > 100 {
		t.Errorf("Invalid engagement score: %d", report.AvgEngagementScore)
	}

	// Generate performance report
	perfReport, err := reporter.GenerateCoursePerformanceReport(courseID)
	if err != nil {
		t.Errorf("Failed to generate performance report: %v", err)
	}

	if perfReport.CourseID != courseID {
		t.Errorf("Course ID mismatch in performance report")
	}

	t.Log("✓ Report generation test passed")
}

// TestReportInterval tests report scheduling
func TestReportInterval(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	reporter := NewReportGenerator(nil, nil, logger)

	if reporter.reportInterval != 24*time.Hour {
		t.Errorf("Expected 24 hour interval, got %v", reporter.reportInterval)
	}

	t.Log("✓ Report interval test passed")
}

// TestAlertService tests alert service
func TestAlertService(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	alertGen := NewAlertGenerator(DefaultPerformanceThreshold(), logger)
	alertSvc := NewAlertService(alertGen, nil, logger)

	// Create subscriber
	emailSub := NewEmailAlertSubscriber(logger)
	alertSvc.Subscribe("email", emailSub)

	// Verify subscription
	alertSvc.mu.RLock()
	if len(alertSvc.subscribers) != 1 {
		t.Errorf("Expected 1 subscriber, got %d", len(alertSvc.subscribers))
	}
	alertSvc.mu.RUnlock()

	// Unsubscribe
	alertSvc.Unsubscribe("email")
	alertSvc.mu.RLock()
	if len(alertSvc.subscribers) != 0 {
		t.Errorf("Expected 0 subscribers after unsubscribe")
	}
	alertSvc.mu.RUnlock()

	t.Log("✓ Alert service test passed")
}

// TestDashboardAlertSubscriber tests dashboard alert storage
func TestDashboardAlertSubscriber(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	dashboard := NewDashboardAlertSubscriber(logger)

	// Add alerts
	for i := 0; i < 5; i++ {
		alert := &PerformanceAlert{
			ID:          uuid.New(),
			UserID:      uuid.New(),
			RecordingID: uuid.New(),
			AlertType:   "low_engagement",
			Severity:    "warning",
			Message:     "Test alert",
			CreatedAt:   time.Now(),
		}

		err := dashboard.OnAlert(alert)
		if err != nil {
			t.Errorf("Failed to add alert: %v", err)
		}
	}

	// Retrieve alerts
	alerts := dashboard.GetRecentAlerts(3)
	if len(alerts) != 3 {
		t.Errorf("Expected 3 alerts, got %d", len(alerts))
	}

	t.Log("✓ Dashboard alert subscriber test passed")
}

// TestEmailAlertSubscriber tests email alerts
func TestEmailAlertSubscriber(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	emailSub := NewEmailAlertSubscriber(logger)

	alert := &PerformanceAlert{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		AlertType: "low_engagement",
		Severity:  "warning",
		Message:   "Test email alert",
		CreatedAt: time.Now(),
	}

	err := emailSub.OnAlert(alert)
	if err != nil {
		t.Errorf("Failed to send email alert: %v", err)
	}

	t.Log("✓ Email alert subscriber test passed")
}

// TestStreamingIntegration tests full streaming to analytics flow
func TestStreamingIntegration(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	collector := NewEventCollector(100, 5*time.Second, logger)
	calculator := NewMetricsCalculator(nil, logger)
	listener := NewStreamingEventListener(collector, calculator, nil, logger)

	recordingID := uuid.New()
	userID := uuid.New()
	sessionID := "session-integration"

	// Simulate full playback lifecycle
	listener.OnPlaybackStarted(recordingID, userID, sessionID)
	listener.OnQualityChanged(sessionID, "auto", "1080p", "user_selected")
	listener.OnBufferEvent(sessionID)
	listener.OnQualityChanged(sessionID, "1080p", "720p", "auto_downgrade")

	// Collect pending events
	if len(listener.activeSessions) != 1 {
		t.Errorf("Expected 1 active session")
	}

	t.Log("✓ Streaming integration test passed")
}

// TestMetricsFlowWithAlerts tests full metrics to alert flow
func TestMetricsFlowWithAlerts(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	alertGen := NewAlertGenerator(DefaultPerformanceThreshold(), logger)
	alertSvc := NewAlertService(alertGen, nil, logger)

	// Register alert subscribers
	dashboard := NewDashboardAlertSubscriber(logger)
	alertSvc.Subscribe("dashboard", dashboard)

	// Create low engagement metrics
	lowMetrics := &EngagementMetrics{
		ID:                   uuid.New(),
		UserID:               uuid.New(),
		RecordingID:          uuid.New(),
		EngagementScore:      20, // Below threshold of 30
		CompletionPercentage: 25, // Below threshold of 30%
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Process metrics
	err := alertSvc.ProcessMetricsForAlerts(lowMetrics)
	if err != nil {
		t.Errorf("Failed to process metrics: %v", err)
	}

	// Check dashboard alerts
	dashAlerts := dashboard.GetRecentAlerts(10)
	if len(dashAlerts) == 0 {
		t.Errorf("Expected alerts to be generated")
	} else if dashAlerts[0].AlertType != "low_engagement" {
		t.Errorf("Expected low_engagement alert")
	}

	t.Log("✓ Metrics to alert flow test passed")
}

// TestMultipleSubscribers tests alert delivery to multiple subscribers
func TestMultipleSubscribers(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	alertGen := NewAlertGenerator(DefaultPerformanceThreshold(), logger)
	alertSvc := NewAlertService(alertGen, nil, logger)

	// Register multiple subscribers
	alertSvc.Subscribe("email", NewEmailAlertSubscriber(logger))
	alertSvc.Subscribe("dashboard", NewDashboardAlertSubscriber(logger))

	alertSvc.mu.RLock()
	subCount := len(alertSvc.subscribers)
	alertSvc.mu.RUnlock()

	if subCount != 2 {
		t.Errorf("Expected 2 subscribers, got %d", subCount)
	}

	t.Log("✓ Multiple subscribers test passed")
}

// TestEventCollectorIntegration tests event collector with batch processing
func TestEventCollectorIntegration(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	collector := NewEventCollector(10, 1*time.Second, logger)

	recordingID := uuid.New()
	userID := uuid.New()

	// Record multiple events
	for i := 0; i < 5; i++ {
		collector.RecordEvent(
			EventPlaybackStarted,
			recordingID,
			userID,
			"session-001",
			map[string]interface{}{"index": i},
		)
	}

	// Get pending events
	events := collector.GetPendingEvents()
	if len(events) != 5 {
		t.Errorf("Expected 5 events, got %d", len(events))
	}

	// Stop collector
	collector.Stop()

	t.Log("✓ Event collector integration test passed")
}

// TestReportGeneratorWithCourses tests report generation for multiple courses
func TestReportGeneratorWithCourses(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	reporter := NewReportGenerator(nil, nil, logger)

	// Generate reports for multiple courses
	courses := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	for _, courseID := range courses {
		report, err := reporter.GenerateCourseEngagementReport(courseID)
		if err != nil {
			t.Errorf("Failed to generate report: %v", err)
		}

		if report.CourseID != courseID {
			t.Errorf("Course ID mismatch")
		}
	}

	t.Log("✓ Multi-course report generation test passed")
}

// TestAlertThresholds tests alert generation with threshold checking
func TestAlertThresholds(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	thresholds := DefaultPerformanceThreshold()
	alertGen := NewAlertGenerator(thresholds, logger)

	tests := []struct {
		name            string
		engagementScore int
		completion      int
		expectAlert     bool
	}{
		{"Good metrics", 85, 90, false},
		{"Low engagement", 20, 90, true},
		{"Low completion", 85, 20, true},
		{"Both low", 20, 20, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &EngagementMetrics{
				ID:                   uuid.New(),
				UserID:               uuid.New(),
				RecordingID:          uuid.New(),
				EngagementScore:      tt.engagementScore,
				CompletionPercentage: tt.completion,
				CreatedAt:            time.Now(),
				UpdatedAt:            time.Now(),
			}

			alert, _ := alertGen.GenerateAlert(metrics)

			if tt.expectAlert && alert == nil {
				t.Errorf("Expected alert but got none")
			} else if !tt.expectAlert && alert != nil {
				t.Errorf("Unexpected alert generated")
			}
		})
	}
}

// TestStreamingListenerBufferTracking tests buffer event counting
func TestStreamingListenerBufferTracking(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	listener := NewStreamingEventListener(
		NewEventCollector(100, 5*time.Second, logger),
		NewMetricsCalculator(nil, logger),
		nil,
		logger,
	)

	recordingID := uuid.New()
	userID := uuid.New()
	sessionID := "buffer-test"

	// Start session
	listener.OnPlaybackStarted(recordingID, userID, sessionID)

	// Trigger buffer events
	for i := 0; i < 5; i++ {
		listener.OnBufferEvent(sessionID)
	}

	// Check buffer count
	listener.mu.RLock()
	session := listener.activeSessions[sessionID]
	listener.mu.RUnlock()

	if session.BufferEvents != 5 {
		t.Errorf("Expected 5 buffer events, got %d", session.BufferEvents)
	}

	t.Log("✓ Buffer tracking test passed")
}

// BenchmarkStreamingEventProcessing benchmarks event processing
func BenchmarkStreamingEventProcessing(b *testing.B) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	listener := NewStreamingEventListener(
		NewEventCollector(100, 5*time.Second, logger),
		NewMetricsCalculator(nil, logger),
		nil,
		logger,
	)

	recordingID := uuid.New()
	userID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sessionID := uuid.New().String()
		listener.OnPlaybackStarted(recordingID, userID, sessionID)
		listener.OnQualityChanged(sessionID, "auto", "1080p", "user")
		listener.OnBufferEvent(sessionID)
	}
}

// BenchmarkAlertGeneration benchmarks alert generation
func BenchmarkAlertGeneration(b *testing.B) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	alertGen := NewAlertGenerator(DefaultPerformanceThreshold(), logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics := &EngagementMetrics{
			ID:                   uuid.New(),
			UserID:               uuid.New(),
			RecordingID:          uuid.New(),
			EngagementScore:      20,
			CompletionPercentage: 25,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		}

		alertGen.GenerateAlert(metrics)
	}
}
