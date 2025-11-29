package analytics

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// StreamingEventListener listens for streaming events and records analytics
type StreamingEventListener struct {
	collector      *EventCollectorImpl
	calculator     *MetricsCalculator
	store          StorageRepository
	logger         *log.Logger
	mu             sync.RWMutex
	activeSessions map[string]*PlaybackSession
}

// NewStreamingEventListener creates a listener for streaming events
func NewStreamingEventListener(
	collector *EventCollectorImpl,
	calculator *MetricsCalculator,
	store StorageRepository,
	logger *log.Logger,
) *StreamingEventListener {
	return &StreamingEventListener{
		collector:      collector,
		calculator:     calculator,
		store:          store,
		logger:         logger,
		activeSessions: make(map[string]*PlaybackSession),
	}
}

// OnPlaybackStarted handles playback start events
func (l *StreamingEventListener) OnPlaybackStarted(recordingID, userID uuid.UUID, sessionID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Record event
	l.collector.RecordEvent(
		EventPlaybackStarted,
		recordingID,
		userID,
		sessionID,
		map[string]interface{}{
			"quality": "auto",
			"bitrate": 2000,
		},
	)

	// Create new session
	session := &PlaybackSession{
		ID:                   uuid.New(),
		RecordingID:          recordingID,
		UserID:               userID,
		SessionStart:         time.Now(),
		TotalDurationSeconds: 3600, // Will be updated
		QualitySelected:      "auto",
		CreatedAt:            time.Now(),
	}

	l.activeSessions[sessionID] = session
	l.logger.Printf("Playback started: user=%s, recording=%s, session=%s", userID, recordingID, sessionID)

	return nil
}

// OnPlaybackStopped handles playback stop events
func (l *StreamingEventListener) OnPlaybackStopped(sessionID string, watchedSeconds int) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	session, ok := l.activeSessions[sessionID]
	if !ok {
		l.logger.Printf("Session not found: %s", sessionID)
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Update session
	now := time.Now()
	session.SessionEnd = &now
	session.WatchedDurationSeconds = watchedSeconds
	session.CompletionRate = float64(watchedSeconds) / float64(session.TotalDurationSeconds) * 100

	// Calculate engagement metrics
	metrics, err := l.calculator.CalculateEngagementMetrics(session.RecordingID, session.UserID, *session)
	if err != nil {
		l.logger.Printf("Failed to calculate metrics: %v", err)
		return err
	}

	// Store metrics
	if err := l.store.StoreEngagementMetrics(metrics); err != nil {
		l.logger.Printf("Failed to store metrics: %v", err)
		return err
	}

	// Record stop event
	l.collector.RecordEvent(
		EventPlaybackStopped,
		session.RecordingID,
		session.UserID,
		sessionID,
		map[string]interface{}{
			"watched_seconds": watchedSeconds,
			"completion":      session.CompletionRate,
		},
	)

	delete(l.activeSessions, sessionID)
	l.logger.Printf("Playback stopped: user=%s, watched=%ds, completion=%.1f%%",
		session.UserID, watchedSeconds, session.CompletionRate)

	return nil
}

// OnQualityChanged handles quality change events
func (l *StreamingEventListener) OnQualityChanged(sessionID string, oldQuality, newQuality string, reason string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	session, ok := l.activeSessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.QualitySelected = newQuality

	// Record event
	l.collector.RecordEvent(
		EventQualityChanged,
		session.RecordingID,
		session.UserID,
		sessionID,
		map[string]interface{}{
			"old_quality": oldQuality,
			"new_quality": newQuality,
			"reason":      reason,
		},
	)

	l.logger.Printf("Quality changed: %s -> %s (reason: %s)", oldQuality, newQuality, reason)
	return nil
}

// OnBufferEvent handles buffer events
func (l *StreamingEventListener) OnBufferEvent(sessionID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	session, ok := l.activeSessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.BufferEvents++

	// Record event
	l.collector.RecordEvent(
		EventBufferEvent,
		session.RecordingID,
		session.UserID,
		sessionID,
		map[string]interface{}{
			"buffer_count": session.BufferEvents,
		},
	)

	return nil
}

// ReportGenerator generates automated reports at scheduled intervals
type ReportGenerator struct {
	store          StorageRepository
	calculator     *MetricsCalculator
	logger         *log.Logger
	reportInterval time.Duration
	stopChan       chan bool
	mu             sync.RWMutex
	lastReportTime map[uuid.UUID]time.Time
}

// NewReportGenerator creates a new report generator
func NewReportGenerator(
	store StorageRepository,
	calculator *MetricsCalculator,
	logger *log.Logger,
) *ReportGenerator {
	return &ReportGenerator{
		store:          store,
		calculator:     calculator,
		logger:         logger,
		reportInterval: 24 * time.Hour, // Daily reports
		stopChan:       make(chan bool),
		lastReportTime: make(map[uuid.UUID]time.Time),
	}
}

// Start begins automated report generation
func (rg *ReportGenerator) Start() {
	ticker := time.NewTicker(rg.reportInterval)
	defer ticker.Stop()

	rg.logger.Printf("Report generator started (interval: %v)", rg.reportInterval)

	for {
		select {
		case <-rg.stopChan:
			rg.logger.Printf("Report generator stopped")
			return
		case <-ticker.C:
			rg.generateDailyReports()
		}
	}
}

// Stop halts report generation
func (rg *ReportGenerator) Stop() {
	rg.stopChan <- true
}

// generateDailyReports generates reports for all courses
func (rg *ReportGenerator) generateDailyReports() {
	rg.mu.Lock()
	defer rg.mu.Unlock()

	rg.logger.Printf("Generating daily reports...")

	// In real implementation, would:
	// 1. Query all active courses
	// 2. Gather metrics for each course
	// 3. Generate engagement and performance reports
	// 4. Store reports
	// 5. Trigger notifications

	rg.logger.Printf("Daily reports generation complete")
}

// GenerateCourseEngagementReport generates a report for specific course
func (rg *ReportGenerator) GenerateCourseEngagementReport(courseID uuid.UUID) (*EngagementReport, error) {
	rg.logger.Printf("Generating engagement report for course %s", courseID)

	report := &EngagementReport{
		CourseID:           courseID,
		ReportDate:         time.Now(),
		AvgEngagementScore: 76, // Would be calculated from metrics
	}

	return report, nil
}

// GenerateCoursePerformanceReport generates a performance report for specific course
func (rg *ReportGenerator) GenerateCoursePerformanceReport(courseID uuid.UUID) (*PerformanceReport, error) {
	rg.logger.Printf("Generating performance report for course %s", courseID)

	report := &PerformanceReport{
		CourseID:     courseID,
		ReportDate:   time.Now(),
		OverallScore: 78, // Would be calculated from metrics
	}

	return report, nil
}

// AlertService manages alert delivery
type AlertService struct {
	alertGen    *AlertGenerator
	store       StorageRepository
	logger      *log.Logger
	mu          sync.RWMutex
	subscribers map[string]AlertSubscriber
}

// AlertSubscriber receives alert notifications
type AlertSubscriber interface {
	OnAlert(alert *PerformanceAlert) error
}

// NewAlertService creates a new alert service
func NewAlertService(
	alertGen *AlertGenerator,
	store StorageRepository,
	logger *log.Logger,
) *AlertService {
	return &AlertService{
		alertGen:    alertGen,
		store:       store,
		logger:      logger,
		subscribers: make(map[string]AlertSubscriber),
	}
}

// Subscribe registers an alert subscriber
func (as *AlertService) Subscribe(name string, subscriber AlertSubscriber) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.subscribers[name] = subscriber
	as.logger.Printf("Alert subscriber registered: %s", name)
}

// Unsubscribe removes an alert subscriber
func (as *AlertService) Unsubscribe(name string) {
	as.mu.Lock()
	defer as.mu.Unlock()

	delete(as.subscribers, name)
	as.logger.Printf("Alert subscriber unregistered: %s", name)
}

// ProcessMetricsForAlerts checks metrics and generates alerts
func (as *AlertService) ProcessMetricsForAlerts(metrics *EngagementMetrics) error {
	as.mu.RLock()
	defer as.mu.RUnlock()

	// Generate alert if needed
	alert, err := as.alertGen.GenerateAlert(metrics)
	if err != nil {
		return err
	}

	if alert != nil {
		// Notify all subscribers
		for name, subscriber := range as.subscribers {
			if err := subscriber.OnAlert(alert); err != nil {
				as.logger.Printf("Subscriber %s failed to process alert: %v", name, err)
			}
		}
		as.logger.Printf("Alert generated and distributed: %s", alert.AlertType)
	}

	return nil
}

// EmailAlertSubscriber sends alerts via email
type EmailAlertSubscriber struct {
	logger *log.Logger
}

// NewEmailAlertSubscriber creates email alert subscriber
func NewEmailAlertSubscriber(logger *log.Logger) *EmailAlertSubscriber {
	return &EmailAlertSubscriber{logger: logger}
}

// OnAlert sends alert via email
func (eas *EmailAlertSubscriber) OnAlert(alert *PerformanceAlert) error {
	eas.logger.Printf("Sending email alert: user=%s, type=%s, severity=%s", alert.UserID, alert.AlertType, alert.Severity)
	// Would send actual email here
	return nil
}

// DashboardAlertSubscriber sends alerts to dashboard
type DashboardAlertSubscriber struct {
	logger *log.Logger
	alerts []*PerformanceAlert
	mu     sync.RWMutex
}

// NewDashboardAlertSubscriber creates dashboard alert subscriber
func NewDashboardAlertSubscriber(logger *log.Logger) *DashboardAlertSubscriber {
	return &DashboardAlertSubscriber{
		logger: logger,
		alerts: make([]*PerformanceAlert, 0),
	}
}

// OnAlert records alert to dashboard
func (das *DashboardAlertSubscriber) OnAlert(alert *PerformanceAlert) error {
	das.mu.Lock()
	defer das.mu.Unlock()

	das.alerts = append(das.alerts, alert)
	das.logger.Printf("Alert added to dashboard: %s", alert.AlertType)

	// Keep only recent alerts (last 100)
	if len(das.alerts) > 100 {
		das.alerts = das.alerts[len(das.alerts)-100:]
	}

	return nil
}

// GetRecentAlerts returns recent dashboard alerts
func (das *DashboardAlertSubscriber) GetRecentAlerts(limit int) []*PerformanceAlert {
	das.mu.RLock()
	defer das.mu.RUnlock()

	if limit > len(das.alerts) {
		limit = len(das.alerts)
	}

	result := make([]*PerformanceAlert, limit)
	copy(result, das.alerts[len(das.alerts)-limit:])
	return result
}

// AnalyticsService orchestrates all analytics components
type AnalyticsService struct {
	collector  *EventCollectorImpl
	calculator *MetricsCalculator
	store      StorageRepository
	listener   *StreamingEventListener
	reporter   *ReportGenerator
	alertSvc   *AlertService
	logger     *log.Logger
	db         *sql.DB
}

// NewAnalyticsService creates a complete analytics service
func NewAnalyticsService(
	db *sql.DB,
	logger *log.Logger,
) (*AnalyticsService, error) {
	// Initialize components
	collector := NewEventCollector(100, 5*time.Second, logger)
	store := NewPostgresAnalyticsStore(db, logger)
	calculator := NewMetricsCalculator(store, logger)
	scorer := NewEngagementScorer(calculator, logger)
	aggregator := NewAggregationService(store, logger)
	alertGen := NewAlertGenerator(DefaultPerformanceThreshold(), logger)

	// Create listener
	listener := NewStreamingEventListener(collector, calculator, store, logger)

	// Create report generator
	reporter := NewReportGenerator(store, calculator, logger)

	// Create alert service
	alertSvc := NewAlertService(alertGen, store, logger)

	// Register alert subscribers
	alertSvc.Subscribe("email", NewEmailAlertSubscriber(logger))
	alertSvc.Subscribe("dashboard", NewDashboardAlertSubscriber(logger))

	service := &AnalyticsService{
		collector:  collector,
		calculator: calculator,
		store:      store,
		listener:   listener,
		reporter:   reporter,
		alertSvc:   alertSvc,
		logger:     logger,
		db:         db,
	}

	// Set batch callback
	collector.SetBatchCallback(func(events []AnalyticsEvent) error {
		return service.processBatchEvents(events)
	})

	logger.Printf("✓ Analytics service initialized")
	return service, nil
}

// processBatchEvents processes a batch of analytics events
func (as *AnalyticsService) processBatchEvents(events []AnalyticsEvent) error {
	as.logger.Printf("Processing batch of %d events", len(events))

	// Store events
	if err := as.store.StoreEvents(events); err != nil {
		as.logger.Printf("Failed to store events: %v", err)
		return err
	}

	return nil
}

// Start begins analytics service
func (as *AnalyticsService) Start() {
	as.logger.Printf("Starting analytics service...")

	// Start collector
	as.collector.SetBatchCallback(func(events []AnalyticsEvent) error {
		return as.processBatchEvents(events)
	})

	// Start report generator in background
	go as.reporter.Start()

	as.logger.Printf("✓ Analytics service started")
}

// Stop gracefully shuts down analytics service
func (as *AnalyticsService) Stop() error {
	as.logger.Printf("Stopping analytics service...")

	// Stop collector
	as.collector.Stop()

	// Stop reporter
	as.reporter.Stop()

	as.logger.Printf("✓ Analytics service stopped")
	return nil
}

// GetEventCollector returns the event collector
func (as *AnalyticsService) GetEventCollector() *EventCollector {
	return as.collector
}

// GetStreamingListener returns the streaming event listener
func (as *AnalyticsService) GetStreamingListener() *StreamingEventListener {
	return as.listener
}

// GetReportGenerator returns the report generator
func (as *AnalyticsService) GetReportGenerator() *ReportGenerator {
	return as.reporter
}

// GetAlertService returns the alert service
func (as *AnalyticsService) GetAlertService() *AlertService {
	return as.alertSvc
}

// ProcessUserMetrics processes metrics and generates alerts
func (as *AnalyticsService) ProcessUserMetrics(metrics *EngagementMetrics) error {
	// Process for alerts
	if err := as.alertSvc.ProcessMetricsForAlerts(metrics); err != nil {
		as.logger.Printf("Failed to process alerts: %v", err)
		return err
	}

	return nil
}
