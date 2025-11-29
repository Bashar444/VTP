package analytics

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMetricsCalculator(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	mc := NewMetricsCalculator(nil, logger)

	recordingID := uuid.New()
	userID := uuid.New()

	session := PlaybackSession{
		ID:                     uuid.New(),
		RecordingID:            recordingID,
		UserID:                 userID,
		SessionStart:           time.Now(),
		TotalDurationSeconds:   3600,
		WatchedDurationSeconds: 3060, // 85% watched
		PauseCount:             3,
		ResumeCount:            2,
		QualitySelected:        "1080p",
		BufferEvents:           2,
		CompletionRate:         85,
		CreatedAt:              time.Now(),
	}

	metrics, err := mc.CalculateEngagementMetrics(recordingID, userID, session)
	if err != nil {
		t.Fatalf("Failed to calculate engagement metrics: %v", err)
	}

	if metrics.EngagementScore < 0 || metrics.EngagementScore > 100 {
		t.Errorf("Engagement score out of range: %d", metrics.EngagementScore)
	}

	if metrics.CompletionPercentage != 85 {
		t.Errorf("Expected completion 85, got %d", metrics.CompletionPercentage)
	}

	t.Logf("Calculated engagement score: %d", metrics.EngagementScore)
}

func TestEngagementScoreCalculation(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	mc := NewMetricsCalculator(nil, logger)

	tests := []struct {
		name       string
		completion int
		buffers    int
		pauses     int
		duration   int
		minScore   int
		maxScore   int
	}{
		{"Perfect watch", 100, 0, 0, 3600, 90, 100},
		{"Good watch", 80, 1, 2, 3600, 70, 85},
		{"Fair watch", 50, 3, 5, 1800, 30, 60},
		{"Poor watch", 20, 10, 10, 600, 0, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := mc.calculateEngagementScore(tt.completion, tt.buffers, tt.pauses, tt.duration)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("Score %d out of expected range [%d, %d]", score, tt.minScore, tt.maxScore)
			}
			t.Logf("Score: %d", score)
		})
	}
}

func TestLectureStatisticsCalculation(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	mc := NewMetricsCalculator(nil, logger)

	recordingID := uuid.New()
	now := time.Now()

	sessions := []PlaybackSession{
		{
			ID:                     uuid.New(),
			RecordingID:            recordingID,
			UserID:                 uuid.New(),
			SessionStart:           now,
			TotalDurationSeconds:   3600,
			WatchedDurationSeconds: 3060, // 85%
			BufferEvents:           1,
			QualitySelected:        "1080p",
			CompletionRate:         85,
			CreatedAt:              now,
		},
		{
			ID:                     uuid.New(),
			RecordingID:            recordingID,
			UserID:                 uuid.New(),
			SessionStart:           now,
			TotalDurationSeconds:   3600,
			WatchedDurationSeconds: 2160, // 60%
			BufferEvents:           3,
			QualitySelected:        "720p",
			CompletionRate:         60,
			CreatedAt:              now,
		},
		{
			ID:                     uuid.New(),
			RecordingID:            recordingID,
			UserID:                 uuid.New(),
			SessionStart:           now,
			TotalDurationSeconds:   3600,
			WatchedDurationSeconds: 1800, // 50%
			BufferEvents:           2,
			QualitySelected:        "1080p",
			CompletionRate:         50,
			CreatedAt:              now,
		},
	}

	stats, err := mc.CalculateLectureStatistics(recordingID, sessions)
	if err != nil {
		t.Fatalf("Failed to calculate lecture statistics: %v", err)
	}

	if stats.UniqueViewers != 3 {
		t.Errorf("Expected 3 unique viewers, got %d", stats.UniqueViewers)
	}

	if stats.AvgWatchTimeSeconds < 1500 || stats.AvgWatchTimeSeconds > 1800 {
		t.Errorf("Unexpected average watch time: %d", stats.AvgWatchTimeSeconds)
	}

	expectedCompletion := (85.0 + 60.0 + 50.0) / 3.0
	if stats.CompletionRate < expectedCompletion-1 || stats.CompletionRate > expectedCompletion+1 {
		t.Errorf("Expected completion %.1f, got %.1f", expectedCompletion, stats.CompletionRate)
	}

	t.Logf("Lecture stats: viewers=%d, avgWatch=%ds, completion=%.1f%%",
		stats.UniqueViewers, stats.AvgWatchTimeSeconds, stats.CompletionRate)
}

func TestCourseStatisticsCalculation(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	mc := NewMetricsCalculator(nil, logger)

	courseID := uuid.New()
	now := time.Now()

	lectureStats := []LectureStatistics{
		{
			ID:                  uuid.New(),
			RecordingID:         uuid.New(),
			UniqueViewers:       30,
			TotalViews:          40,
			AvgWatchTimeSeconds: 2400,
			CompletionRate:      75,
			TotalBufferEvents:   5,
			CreatedAt:           now,
			UpdatedAt:           now,
		},
		{
			ID:                  uuid.New(),
			RecordingID:         uuid.New(),
			UniqueViewers:       28,
			TotalViews:          35,
			AvgWatchTimeSeconds: 2100,
			CompletionRate:      70,
			TotalBufferEvents:   3,
			CreatedAt:           now,
			UpdatedAt:           now,
		},
		{
			ID:                  uuid.New(),
			RecordingID:         uuid.New(),
			UniqueViewers:       25,
			TotalViews:          30,
			AvgWatchTimeSeconds: 1800,
			CompletionRate:      65,
			TotalBufferEvents:   4,
			CreatedAt:           now,
			UpdatedAt:           now,
		},
	}

	stats, err := mc.CalculateCourseStatistics(courseID, lectureStats)
	if err != nil {
		t.Fatalf("Failed to calculate course statistics: %v", err)
	}

	if stats.TotalLectures != 3 {
		t.Errorf("Expected 3 lectures, got %d", stats.TotalLectures)
	}

	expectedAvgCompletion := (75.0 + 70.0 + 65.0) / 3.0
	if stats.AverageCompletion < expectedAvgCompletion-1 || stats.AverageCompletion > expectedAvgCompletion+1 {
		t.Errorf("Expected completion %.1f, got %.1f", expectedAvgCompletion, stats.AverageCompletion)
	}

	t.Logf("Course stats: lectures=%d, students=%d, avgCompletion=%.1f%%",
		stats.TotalLectures, stats.TotalStudents, stats.AverageCompletion)
}

func TestEngagementScorer(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	mc := NewMetricsCalculator(nil, logger)
	scorer := NewEngagementScorer(mc, logger)

	session := PlaybackSession{
		ID:                     uuid.New(),
		RecordingID:            uuid.New(),
		UserID:                 uuid.New(),
		SessionStart:           time.Now(),
		TotalDurationSeconds:   3600,
		WatchedDurationSeconds: 3000, // 83%
		PauseCount:             3,
		ResumeCount:            2,
		QualitySelected:        "1080p",
		BufferEvents:           2,
		CompletionRate:         83,
		CreatedAt:              time.Now(),
	}

	breakdown := scorer.ScoreEngagement(session, []AnalyticsEvent{})
	if breakdown.TotalScore < 0 || breakdown.TotalScore > 100 {
		t.Errorf("Score out of range: %d", breakdown.TotalScore)
	}

	if breakdown.CompletionScore < 0 || breakdown.CompletionScore > 40 {
		t.Errorf("Completion score out of range: %d", breakdown.CompletionScore)
	}

	if breakdown.DurationScore < 0 || breakdown.DurationScore > 20 {
		t.Errorf("Duration score out of range: %d", breakdown.DurationScore)
	}

	if breakdown.QualityScore < 0 || breakdown.QualityScore > 20 {
		t.Errorf("Quality score out of range: %d", breakdown.QualityScore)
	}

	if breakdown.InteractionScore < 0 || breakdown.InteractionScore > 20 {
		t.Errorf("Interaction score out of range: %d", breakdown.InteractionScore)
	}

	t.Logf("Engagement breakdown: total=%d, completion=%d, duration=%d, quality=%d, interaction=%d",
		breakdown.TotalScore, breakdown.CompletionScore, breakdown.DurationScore,
		breakdown.QualityScore, breakdown.InteractionScore)
}

func TestQualityRecommendation(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	mc := NewMetricsCalculator(nil, logger)
	scorer := NewEngagementScorer(mc, logger)

	tests := []struct {
		name           string
		bufferEvents   int
		expectedQuality string
	}{
		{"Good connection", 0, "4K"},
		{"Slight buffering", 2, "1080p"},
		{"Heavy buffering", 5, "720p"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := PlaybackSession{
				ID:           uuid.New(),
				BufferEvents: tt.bufferEvents,
			}
			recommended := scorer.recommendQuality(session)
			if recommended != tt.expectedQuality {
				t.Errorf("Expected %s, got %s", tt.expectedQuality, recommended)
			}
		})
	}
}

func TestAggregationService(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	as := NewAggregationService(nil, logger)

	courseID := uuid.New()
	now := time.Now()

	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	metrics, err := as.AggregateWeeklyMetrics(courseID, weekStart)
	if err != nil {
		t.Fatalf("Failed to aggregate weekly metrics: %v", err)
	}

	if metrics.PeriodType != "weekly" {
		t.Errorf("Expected period type 'weekly', got %s", metrics.PeriodType)
	}

	if metrics.PeriodEnd.Before(metrics.PeriodStart) {
		t.Errorf("Period end before start")
	}

	t.Logf("Weekly metrics: %s to %s", metrics.PeriodStart.Format("2006-01-02"), metrics.PeriodEnd.Format("2006-01-02"))
}

func TestMonthlyAggregation(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	as := NewAggregationService(nil, logger)

	courseID := uuid.New()
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	metrics, err := as.AggregateMonthlyMetrics(courseID, monthStart)
	if err != nil {
		t.Fatalf("Failed to aggregate monthly metrics: %v", err)
	}

	if metrics.PeriodType != "monthly" {
		t.Errorf("Expected period type 'monthly', got %s", metrics.PeriodType)
	}

	t.Logf("Monthly metrics: %s to %s", metrics.PeriodStart.Format("2006-01-02"), metrics.PeriodEnd.Format("2006-01-02"))
}

func TestTrendCalculation(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	as := NewAggregationService(nil, logger)

	tests := []struct {
		name     string
		current  float64
		previous float64
	}{
		{"Positive trend", 75.0, 60.0},   // 25% increase
		{"Negative trend", 50.0, 75.0},   // -33% decrease
		{"No change", 70.0, 70.0},        // 0% change
		{"Zero previous", 80.0, 0.0},     // Edge case
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trend := as.CalculateTrendScore(tt.current, tt.previous)
			t.Logf("Trend from %.1f to %.1f: %.1f%%", tt.previous, tt.current, trend)
		})
	}
}

func TestAlertGeneration(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	thresholds := DefaultPerformanceThreshold()
	alertGen := NewAlertGenerator(thresholds, logger)

	// Low engagement alert
	lowEngagementMetrics := &EngagementMetrics{
		ID:                   uuid.New(),
		UserID:               uuid.New(),
		RecordingID:          uuid.New(),
		EngagementScore:      20,
		CompletionPercentage: 80,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	alert, err := alertGen.GenerateAlert(lowEngagementMetrics)
	if err != nil {
		t.Fatalf("Failed to generate alert: %v", err)
	}

	if alert == nil {
		t.Errorf("Expected alert for low engagement, got nil")
	} else if alert.AlertType != "low_engagement" {
		t.Errorf("Expected low_engagement alert, got %s", alert.AlertType)
	}

	t.Logf("Generated alert: %s - %s", alert.AlertType, alert.Message)
}

func TestLowCompletionAlert(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	thresholds := DefaultPerformanceThreshold()
	alertGen := NewAlertGenerator(thresholds, logger)

	lowCompletionMetrics := &EngagementMetrics{
		ID:                   uuid.New(),
		UserID:               uuid.New(),
		RecordingID:          uuid.New(),
		EngagementScore:      50,
		CompletionPercentage: 20, // Below 30% threshold
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	alert, err := alertGen.GenerateAlert(lowCompletionMetrics)
	if err != nil {
		t.Fatalf("Failed to generate alert: %v", err)
	}

	if alert == nil {
		t.Errorf("Expected alert for low completion, got nil")
	} else if alert.AlertType != "low_completion" {
		t.Errorf("Expected low_completion alert, got %s", alert.AlertType)
	}

	t.Logf("Generated alert: %s - %s", alert.AlertType, alert.Message)
}

func TestNoAlertForGoodMetrics(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	thresholds := DefaultPerformanceThreshold()
	alertGen := NewAlertGenerator(thresholds, logger)

	goodMetrics := &EngagementMetrics{
		ID:                   uuid.New(),
		UserID:               uuid.New(),
		RecordingID:          uuid.New(),
		EngagementScore:      85,
		CompletionPercentage: 90,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	alert, err := alertGen.GenerateAlert(goodMetrics)
	if err != nil {
		t.Fatalf("Failed to check metrics: %v", err)
	}

	if alert != nil {
		t.Errorf("Unexpected alert for good metrics: %s", alert.AlertType)
	}

	t.Log("No alert generated for good metrics")
}

func TestDefaultPerformanceThreshold(t *testing.T) {
	threshold := DefaultPerformanceThreshold()

	if threshold.LowEngagementScore != 30 {
		t.Errorf("Expected low engagement score 30, got %d", threshold.LowEngagementScore)
	}

	if threshold.HighBufferEventCount != 5 {
		t.Errorf("Expected high buffer count 5, got %d", threshold.HighBufferEventCount)
	}

	if threshold.LowCompletionRate != 0.3 {
		t.Errorf("Expected low completion 0.3, got %v", threshold.LowCompletionRate)
	}

	t.Log("Default thresholds validated")
}

func TestEngagementScoreEdgeCases(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	mc := NewMetricsCalculator(nil, logger)

	tests := []struct {
		name       string
		completion int
		buffers    int
		pauses     int
		duration   int
	}{
		{"Zero duration", 0, 0, 0, 0},
		{"Extreme buffers", 100, 100, 0, 3600},
		{"Extreme pauses", 100, 0, 100, 3600},
		{"Long duration", 0, 0, 0, 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := mc.calculateEngagementScore(tt.completion, tt.buffers, tt.pauses, tt.duration)
			if score < 0 || score > 100 {
				t.Errorf("Score %d out of valid range", score)
			}
			t.Logf("Score: %d", score)
		})
	}
}

func TestEmptyLectureStatistics(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	mc := NewMetricsCalculator(nil, logger)

	recordingID := uuid.New()
	stats, err := mc.CalculateLectureStatistics(recordingID, []PlaybackSession{})
	if err != nil {
		t.Fatalf("Failed to handle empty sessions: %v", err)
	}

	if stats.RecordingID != recordingID {
		t.Errorf("Recording ID mismatch")
	}

	if stats.UniqueViewers != 0 {
		t.Errorf("Expected 0 viewers for empty sessions")
	}

	t.Log("Empty lecture statistics handled correctly")
}

func TestEmptyCourseStatistics(t *testing.T) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	mc := NewMetricsCalculator(nil, logger)

	courseID := uuid.New()
	stats, err := mc.CalculateCourseStatistics(courseID, []LectureStatistics{})
	if err != nil {
		t.Fatalf("Failed to handle empty lecture stats: %v", err)
	}

	if stats.CourseID != courseID {
		t.Errorf("Course ID mismatch")
	}

	if stats.TotalLectures != 0 {
		t.Errorf("Expected 0 lectures for empty stats")
	}

	t.Log("Empty course statistics handled correctly")
}

func BenchmarkEngagementScoreCalculation(b *testing.B) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	mc := NewMetricsCalculator(nil, logger)

	for i := 0; i < b.N; i++ {
		mc.calculateEngagementScore(75, 2, 3, 3600)
	}
}

func BenchmarkLectureStatisticsCalculation(b *testing.B) {
	logger := log.New(os.Stderr, "test: ", log.LstdFlags)
	mc := NewMetricsCalculator(nil, logger)

	recordingID := uuid.New()
	sessions := make([]PlaybackSession, 50)
	for i := 0; i < 50; i++ {
		sessions[i] = PlaybackSession{
			ID:                     uuid.New(),
			RecordingID:            recordingID,
			UserID:                 uuid.New(),
			TotalDurationSeconds:   3600,
			WatchedDurationSeconds: 2700,
			BufferEvents:           2,
			CompletionRate:         75,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mc.CalculateLectureStatistics(recordingID, sessions)
	}
}
