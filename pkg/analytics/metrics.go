package analytics

import (
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MetricsCalculator computes engagement and quality metrics from analytics events
type MetricsCalculator struct {
	store  StorageRepository
	logger *log.Logger
	mu     sync.RWMutex
}

// NewMetricsCalculator creates a new metrics calculator
func NewMetricsCalculator(store StorageRepository, logger *log.Logger) *MetricsCalculator {
	return &MetricsCalculator{
		store:  store,
		logger: logger,
	}
}

// CalculateEngagementMetrics computes engagement scores for a user on a recording
func (mc *MetricsCalculator) CalculateEngagementMetrics(recordingID, userID uuid.UUID, session PlaybackSession) (*EngagementMetrics, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Calculate watch ratio
	watchRatio := float64(session.WatchedDurationSeconds) / float64(session.TotalDurationSeconds)
	if watchRatio < 0 {
		watchRatio = 0
	}
	if watchRatio > 1 {
		watchRatio = 1
	}

	completionPercentage := int(watchRatio * 100)

	// Calculate engagement score (0-100)
	engagementScore := mc.calculateEngagementScore(
		completionPercentage,
		session.BufferEvents,
		session.PauseCount,
		session.TotalDurationSeconds,
	)

	metrics := &EngagementMetrics{
		ID:                    uuid.New(),
		RecordingID:           recordingID,
		UserID:                userID,
		TotalWatchTimeSeconds: session.WatchedDurationSeconds,
		CompletionPercentage:  completionPercentage,
		AvgQuality:            session.QualitySelected,
		RewatchCount:          0, // Will be calculated from event history
		EngagementScore:       engagementScore,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	mc.logger.Printf("Calculated engagement metrics for user %s on recording %s: score=%d, completion=%d%%",
		userID, recordingID, engagementScore, completionPercentage)

	return metrics, nil
}

// calculateEngagementScore computes engagement score (0-100) based on multiple factors
func (mc *MetricsCalculator) calculateEngagementScore(completionPercentage, bufferEvents, pauseCount, durationSeconds int) int {
	score := 0.0

	// Completion rate (40% weight)
	completionScore := float64(completionPercentage) * 0.4 / 100.0
	score += completionScore

	// Watch duration (30% weight) - longer is better
	durationFactor := math.Min(float64(durationSeconds)/3600.0, 1.0) // Max 1 hour
	durationScore := durationFactor * 30.0
	score += durationScore

	// Buffer penalty (20% weight)
	bufferPenalty := math.Min(float64(bufferEvents)*2.0, 20.0) // Max 20 point penalty
	score -= bufferPenalty

	// Engagement penalty for too many pauses (10% weight)
	pausePenalty := math.Min(float64(pauseCount)*0.5, 10.0) // Max 10 point penalty
	score -= pausePenalty

	// Clamp score to 0-100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return int(math.Round(score))
}

// CalculateLectureStatistics computes aggregate statistics for a lecture
func (mc *MetricsCalculator) CalculateLectureStatistics(recordingID uuid.UUID, sessions []PlaybackSession) (*LectureStatistics, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if len(sessions) == 0 {
		return &LectureStatistics{
			ID:          uuid.New(),
			RecordingID: recordingID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}, nil
	}

	uniqueViewers := len(sessions)
	totalViews := len(sessions) // Could count rewatches separately
	var totalWatchTime int
	var totalCompletionRate float64
	totalBufferEvents := 0
	qualityDist := make(map[string]int)

	for _, session := range sessions {
		totalWatchTime += session.WatchedDurationSeconds
		totalCompletionRate += session.CompletionRate
		totalBufferEvents += session.BufferEvents
		if session.QualitySelected != "" {
			qualityDist[session.QualitySelected]++
		}
	}

	avgWatchTime := totalWatchTime / uniqueViewers
	avgCompletionRate := totalCompletionRate / float64(uniqueViewers)

	stats := &LectureStatistics{
		ID:                  uuid.New(),
		RecordingID:         recordingID,
		UniqueViewers:       uniqueViewers,
		TotalViews:          totalViews,
		AvgWatchTimeSeconds: avgWatchTime,
		CompletionRate:      avgCompletionRate,
		TotalBufferEvents:   totalBufferEvents,
		QualityDistribution: qualityDist,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	mc.logger.Printf("Calculated lecture statistics for recording %s: viewers=%d, avgWatchTime=%ds, completionRate=%.1f%%",
		recordingID, uniqueViewers, avgWatchTime, avgCompletionRate)

	return stats, nil
}

// CalculateCourseStatistics computes aggregate statistics for a course
func (mc *MetricsCalculator) CalculateCourseStatistics(courseID uuid.UUID, lectureStats []LectureStatistics) (*CourseStatistics, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if len(lectureStats) == 0 {
		return &CourseStatistics{
			ID:       uuid.New(),
			CourseID: courseID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil
	}

	totalLectures := len(lectureStats)
	uniqueStudents := 0
	var totalCompletionRate float64

	// Count unique students (simplified - assumes stats contain unique student count)
	for _, stats := range lectureStats {
		uniqueStudents += stats.UniqueViewers
		totalCompletionRate += stats.CompletionRate
	}

	// Average completion rate
	avgCompletionRate := totalCompletionRate / float64(totalLectures)
	attendanceRate := float64(totalLectures) * 100.0 / float64(totalLectures) // Placeholder

	courseStats := &CourseStatistics{
		ID:                uuid.New(),
		CourseID:          courseID,
		TotalLectures:     totalLectures,
		TotalStudents:     uniqueStudents,
		AvgAttendanceRate: attendanceRate,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	mc.logger.Printf("Calculated course statistics for course %s: lectures=%d, students=%d, avgCompletion=%.1f%%",
		courseID, totalLectures, uniqueStudents, avgCompletionRate)

	return courseStats, nil
}

// EngagementScorer provides detailed engagement analysis
type EngagementScorer struct {
	metrics *MetricsCalculator
	logger  *log.Logger
}

// NewEngagementScorer creates a new engagement scorer
func NewEngagementScorer(metrics *MetricsCalculator, logger *log.Logger) *EngagementScorer {
	return &EngagementScorer{
		metrics: metrics,
		logger:  logger,
	}
}

// ScoreEngagement computes comprehensive engagement score with breakdown
func (es *EngagementScorer) ScoreEngagement(session PlaybackSession, events []AnalyticsEvent) *EngagementBreakdown {
	breakdown := &EngagementBreakdown{
		SessionID:           session.ID,
		TotalScore:          0,
		CompletionScore:     0,
		DurationScore:       0,
		QualityScore:        0,
		InteractionScore:    0,
		QualityIndex:        session.QualitySelected,
		RecommendedQuality:  es.recommendQuality(session),
		CreatedAt:           time.Now(),
	}

	// Completion scoring (40 points max)
	completionRatio := float64(session.WatchedDurationSeconds) / float64(session.TotalDurationSeconds)
	breakdown.CompletionScore = int(completionRatio * 40)

	// Duration scoring (20 points max)
	durationMinutes := session.WatchedDurationSeconds / 60
	breakdown.DurationScore = 0
	if durationMinutes >= 30 {
		breakdown.DurationScore = 20
	} else if durationMinutes >= 15 {
		breakdown.DurationScore = 15
	} else if durationMinutes >= 5 {
		breakdown.DurationScore = 10
	} else {
		breakdown.DurationScore = 5
	}

	// Quality scoring (20 points max)
	breakdownPenalty := 0
	if session.BufferEvents > 5 {
		breakdownPenalty = 15
	} else if session.BufferEvents > 2 {
		breakdownPenalty = 10
	} else if session.BufferEvents > 0 {
		breakdownPenalty = 5
	}
	breakdown.QualityScore = 20 - breakdownPenalty

	// Interaction scoring (20 points max)
	interactionBonus := 0
	if session.ResumeCount > 0 {
		interactionBonus += 10 // Indicates rewatching
	}
	if session.PauseCount > 0 && session.PauseCount <= 5 {
		interactionBonus += 10 // Indicates note-taking or thinking
	}
	breakdown.InteractionScore = interactionBonus

	// Total score
	breakdown.TotalScore = breakdown.CompletionScore + breakdown.DurationScore + breakdown.QualityScore + breakdown.InteractionScore

	es.logger.Printf("Scored engagement for session %s: total=%d (completion=%d, duration=%d, quality=%d, interaction=%d)",
		session.ID, breakdown.TotalScore, breakdown.CompletionScore, breakdown.DurationScore, breakdown.QualityScore, breakdown.InteractionScore)

	return breakdown
}

// recommendQuality suggests optimal quality based on session characteristics
func (es *EngagementScorer) recommendQuality(session PlaybackSession) string {
	// If experiencing buffer events, recommend lower quality
	if session.BufferEvents > 3 {
		return "720p" // Standard definition
	}
	if session.BufferEvents > 1 {
		return "1080p" // HD
	}
	// Good experience, can handle high quality
	return "4K" // Ultra HD
}

// AggregationService aggregates metrics across time periods
type AggregationService struct {
	store  StorageRepository
	logger *log.Logger
}

// NewAggregationService creates a new aggregation service
func NewAggregationService(store StorageRepository, logger *log.Logger) *AggregationService {
	return &AggregationService{
		store:  store,
		logger: logger,
	}
}

// AggregateWeeklyMetrics aggregates metrics for a specific week
func (as *AggregationService) AggregateWeeklyMetrics(courseID uuid.UUID, weekStart time.Time) (*TimeSeriesMetrics, error) {
	weekEnd := weekStart.AddDate(0, 0, 7)

	metrics := &TimeSeriesMetrics{
		ID:           uuid.New(),
		CourseID:     courseID,
		PeriodStart:  weekStart,
		PeriodEnd:    weekEnd,
		PeriodType:   "weekly",
		AverageScore: 0,
		TrendScore:   0,
		CreatedAt:    time.Now(),
	}

	as.logger.Printf("Aggregated weekly metrics for course %s: period=%s to %s",
		courseID, weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02"))

	return metrics, nil
}

// AggregateMonthlyMetrics aggregates metrics for a specific month
func (as *AggregationService) AggregateMonthlyMetrics(courseID uuid.UUID, monthStart time.Time) (*TimeSeriesMetrics, error) {
	monthEnd := monthStart.AddDate(0, 1, 0)

	metrics := &TimeSeriesMetrics{
		ID:           uuid.New(),
		CourseID:     courseID,
		PeriodStart:  monthStart,
		PeriodEnd:    monthEnd,
		PeriodType:   "monthly",
		AverageScore: 0,
		TrendScore:   0,
		CreatedAt:    time.Now(),
	}

	as.logger.Printf("Aggregated monthly metrics for course %s: period=%s to %s",
		courseID, monthStart.Format("2006-01-02"), monthEnd.Format("2006-01-02"))

	return metrics, nil
}

// CalculateTrendScore computes trend direction and magnitude
func (as *AggregationService) CalculateTrendScore(currentMetrics, previousMetrics float64) float64 {
	if previousMetrics == 0 {
		return 0
	}
	trend := (currentMetrics - previousMetrics) / previousMetrics
	return trend * 100 // Return as percentage change
}

// EngagementBreakdown provides detailed engagement scoring
type EngagementBreakdown struct {
	SessionID           uuid.UUID `json:"session_id"`
	TotalScore          int       `json:"total_score"`          // 0-100
	CompletionScore     int       `json:"completion_score"`     // 0-40
	DurationScore       int       `json:"duration_score"`       // 0-20
	QualityScore        int       `json:"quality_score"`        // 0-20
	InteractionScore    int       `json:"interaction_score"`    // 0-20
	QualityIndex        string    `json:"quality_index"`
	RecommendedQuality  string    `json:"recommended_quality"`
	CreatedAt           time.Time `json:"created_at"`
}

// TimeSeriesMetrics represents aggregated metrics over a time period
type TimeSeriesMetrics struct {
	ID           uuid.UUID `json:"id"`
	CourseID     uuid.UUID `json:"course_id"`
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
	PeriodType   string    `json:"period_type"` // weekly, monthly, etc
	AverageScore float64   `json:"average_score"`
	TrendScore   float64   `json:"trend_score"` // percentage change
	CreatedAt    time.Time `json:"created_at"`
}

// PerformanceThreshold defines thresholds for performance alerts
type PerformanceThreshold struct {
	LowEngagementScore   int
	HighBufferEventCount int
	LowCompletionRate    float64
	HighDropoutRate      float64
}

// DefaultPerformanceThreshold returns standard thresholds
func DefaultPerformanceThreshold() PerformanceThreshold {
	return PerformanceThreshold{
		LowEngagementScore:   30, // Below 30 is concerning
		HighBufferEventCount: 5,  // More than 5 buffer events
		LowCompletionRate:    0.3, // Below 30% completion
		HighDropoutRate:      0.2, // More than 20% dropout
	}
}

// AlertGenerator creates alerts based on performance metrics
type AlertGenerator struct {
	thresholds PerformanceThreshold
	logger     *log.Logger
}

// NewAlertGenerator creates a new alert generator
func NewAlertGenerator(thresholds PerformanceThreshold, logger *log.Logger) *AlertGenerator {
	return &AlertGenerator{
		thresholds: thresholds,
		logger:     logger,
	}
}

// GenerateAlert creates an alert if metrics exceed thresholds
func (ag *AlertGenerator) GenerateAlert(metrics *EngagementMetrics) (*PerformanceAlert, error) {
	if metrics.EngagementScore < ag.thresholds.LowEngagementScore {
		alert := &PerformanceAlert{
			ID:          uuid.New(),
			UserID:      metrics.UserID,
			RecordingID: metrics.RecordingID,
			AlertType:   "low_engagement",
			Severity:    "warning",
			Message:     fmt.Sprintf("Low engagement score: %d", metrics.EngagementScore),
			CreatedAt:   time.Now(),
		}
		ag.logger.Printf("Generated alert for user %s: %s", metrics.UserID, alert.Message)
		return alert, nil
	}

	if int(metrics.CompletionPercentage) < int(ag.thresholds.LowCompletionRate*100) {
		alert := &PerformanceAlert{
			ID:          uuid.New(),
			UserID:      metrics.UserID,
			RecordingID: metrics.RecordingID,
			AlertType:   "low_completion",
			Severity:    "info",
			Message:     fmt.Sprintf("Low completion rate: %d%%", metrics.CompletionPercentage),
			CreatedAt:   time.Now(),
		}
		ag.logger.Printf("Generated alert for user %s: %s", metrics.UserID, alert.Message)
		return alert, nil
	}

	return nil, nil
}

// PerformanceAlert represents a performance-based alert
type PerformanceAlert struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	RecordingID uuid.UUID `json:"recording_id"`
	AlertType   string    `json:"alert_type"` // low_engagement, low_completion, etc
	Severity    string    `json:"severity"`   // info, warning, critical
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}
