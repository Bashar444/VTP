package analytics

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EventCollectorImpl implements event collection and batching
type EventCollectorImpl struct {
	events       []AnalyticsEvent
	eventsMu     sync.RWMutex
	maxBatchSize int
	batchTimeout time.Duration
	ticker       *time.Ticker
	doneCh       chan struct{}
	onBatchFull  func(events []AnalyticsEvent) error
	logger       *log.Logger
}

// NewEventCollector creates a new event collector
func NewEventCollector(maxBatchSize int, batchTimeout time.Duration, logger *log.Logger) *EventCollectorImpl {
	ec := &EventCollectorImpl{
		events:       make([]AnalyticsEvent, 0, maxBatchSize),
		maxBatchSize: maxBatchSize,
		batchTimeout: batchTimeout,
		doneCh:       make(chan struct{}),
		logger:       logger,
	}

	// Start background batch processor
	go ec.processBatches()

	return ec
}

// RecordEvent records a new analytics event
func (ec *EventCollectorImpl) RecordEvent(eventType EventType, recordingID, userID uuid.UUID, sessionID string, metadata map[string]interface{}) error {
	event := AnalyticsEvent{
		EventID:     uuid.New(),
		EventType:   eventType,
		RecordingID: recordingID,
		UserID:      userID,
		SessionID:   sessionID,
		Timestamp:   time.Now(),
		Metadata:    metadata,
		CreatedAt:   time.Now(),
	}

	return ec.AddEvent(event)
}

// AddEvent adds an event to the collection
func (ec *EventCollectorImpl) AddEvent(event AnalyticsEvent) error {
	ec.eventsMu.Lock()
	defer ec.eventsMu.Unlock()

	ec.events = append(ec.events, event)

	if len(ec.events) >= ec.maxBatchSize {
		return ec.flushLocked()
	}

	return nil
}

// GetPendingEvents returns events awaiting batch processing
func (ec *EventCollectorImpl) GetPendingEvents() []AnalyticsEvent {
	ec.eventsMu.RLock()
	defer ec.eventsMu.RUnlock()

	result := make([]AnalyticsEvent, len(ec.events))
	copy(result, ec.events)
	return result
}

// SetBatchCallback sets the callback for full batches
func (ec *EventCollectorImpl) SetBatchCallback(callback func(events []AnalyticsEvent) error) {
	ec.onBatchFull = callback
}

// processBatches handles periodic batch processing
func (ec *EventCollectorImpl) processBatches() {
	ec.ticker = time.NewTicker(ec.batchTimeout)
	defer ec.ticker.Stop()

	for {
		select {
		case <-ec.ticker.C:
			ec.flushBatch()
		case <-ec.doneCh:
			return
		}
	}
}

// flushBatch flushes pending events
func (ec *EventCollectorImpl) flushBatch() {
	ec.eventsMu.Lock()
	err := ec.flushLocked()
	ec.eventsMu.Unlock()

	if err != nil && ec.logger != nil {
		ec.logger.Printf("[Analytics] Error flushing batch: %v\n", err)
	}
}

// flushLocked flushes events (must be called with lock held)
func (ec *EventCollectorImpl) flushLocked() error {
	if len(ec.events) == 0 {
		return nil
	}

	if ec.onBatchFull == nil {
		ec.events = ec.events[:0]
		return nil
	}

	eventsCopy := make([]AnalyticsEvent, len(ec.events))
	copy(eventsCopy, ec.events)
	ec.events = ec.events[:0]

	// Call callback outside of lock to avoid deadlock
	// This is not ideal but necessary for callback to do I/O
	go func() {
		if err := ec.onBatchFull(eventsCopy); err != nil && ec.logger != nil {
			ec.logger.Printf("[Analytics] Error processing batch: %v\n", err)
		}
	}()

	return nil
}

// Stop stops the event collector
func (ec *EventCollectorImpl) Stop() {
	// Flush remaining events
	ec.flushBatch()
	close(ec.doneCh)
}

// EventSerializer serializes events to JSON
type EventSerializer struct {
	logger *log.Logger
}

// NewEventSerializer creates a new event serializer
func NewEventSerializer(logger *log.Logger) *EventSerializer {
	return &EventSerializer{logger: logger}
}

// SerializeEvent serializes a single event to JSON
func (es *EventSerializer) SerializeEvent(event AnalyticsEvent) ([]byte, error) {
	return json.Marshal(event)
}

// SerializeEvents serializes multiple events to JSON
func (es *EventSerializer) SerializeEvents(events []AnalyticsEvent) ([]byte, error) {
	return json.Marshal(events)
}

// DeserializeEvent deserializes a JSON event
func (es *EventSerializer) DeserializeEvent(data []byte) (*AnalyticsEvent, error) {
	var event AnalyticsEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("failed to deserialize event: %w", err)
	}
	return &event, nil
}

// DeserializeEvents deserializes JSON events
func (es *EventSerializer) DeserializeEvents(data []byte) ([]AnalyticsEvent, error) {
	var events []AnalyticsEvent
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, fmt.Errorf("failed to deserialize events: %w", err)
	}
	return events, nil
}

// EventValidator validates events
type EventValidator struct {
	logger *log.Logger
}

// NewEventValidator creates a new event validator
func NewEventValidator(logger *log.Logger) *EventValidator {
	return &EventValidator{logger: logger}
}

// ValidateEvent validates a single event
func (ev *EventValidator) ValidateEvent(event AnalyticsEvent) error {
	if event.EventID == uuid.Nil {
		return fmt.Errorf("event ID is required")
	}
	if event.RecordingID == uuid.Nil {
		return fmt.Errorf("recording ID is required")
	}
	if event.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}
	if event.SessionID == "" {
		return fmt.Errorf("session ID is required")
	}
	if event.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}

	// Validate event type
	validTypes := map[EventType]bool{
		EventRecordingStarted: true,
		EventRecordingStopped: true,
		EventPlaybackStarted:  true,
		EventPlaybackStopped:  true,
		EventPlaybackPaused:   true,
		EventPlaybackResumed:  true,
		EventSegmentRequested: true,
		EventQualitySelected:  true,
		EventQualityChanged:   true,
		EventBufferEvent:      true,
		EventSessionEnded:     true,
	}

	if !validTypes[event.EventType] {
		return fmt.Errorf("invalid event type: %s", event.EventType)
	}

	return nil
}

// ValidateEvents validates multiple events
func (ev *EventValidator) ValidateEvents(events []AnalyticsEvent) []error {
	var errors []error
	for i, event := range events {
		if err := ev.ValidateEvent(event); err != nil {
			errors = append(errors, fmt.Errorf("event %d: %w", i, err))
		}
	}
	return errors
}

// EventBuilder helps construct events
type EventBuilder struct {
	event AnalyticsEvent
}

// NewEventBuilder creates a new event builder
func NewEventBuilder() *EventBuilder {
	return &EventBuilder{
		event: AnalyticsEvent{
			EventID:   uuid.New(),
			Timestamp: time.Now(),
			CreatedAt: time.Now(),
			Metadata:  make(map[string]interface{}),
		},
	}
}

// WithType sets the event type
func (eb *EventBuilder) WithType(eventType EventType) *EventBuilder {
	eb.event.EventType = eventType
	return eb
}

// WithRecordingID sets the recording ID
func (eb *EventBuilder) WithRecordingID(recordingID uuid.UUID) *EventBuilder {
	eb.event.RecordingID = recordingID
	return eb
}

// WithUserID sets the user ID
func (eb *EventBuilder) WithUserID(userID uuid.UUID) *EventBuilder {
	eb.event.UserID = userID
	return eb
}

// WithSessionID sets the session ID
func (eb *EventBuilder) WithSessionID(sessionID string) *EventBuilder {
	eb.event.SessionID = sessionID
	return eb
}

// WithMetadata adds metadata
func (eb *EventBuilder) WithMetadata(key string, value interface{}) *EventBuilder {
	eb.event.Metadata[key] = value
	return eb
}

// WithTimestamp sets the timestamp
func (eb *EventBuilder) WithTimestamp(timestamp time.Time) *EventBuilder {
	eb.event.Timestamp = timestamp
	return eb
}

// Build returns the constructed event
func (eb *EventBuilder) Build() AnalyticsEvent {
	return eb.event
}

// PlaybackSessionBuilder helps construct playback sessions
type PlaybackSessionBuilder struct {
	session PlaybackSession
}

// NewPlaybackSessionBuilder creates a new session builder
func NewPlaybackSessionBuilder() *PlaybackSessionBuilder {
	return &PlaybackSessionBuilder{
		session: PlaybackSession{
			ID:           uuid.New(),
			SessionStart: time.Now(),
			CreatedAt:    time.Now(),
		},
	}
}

// WithRecordingID sets the recording ID
func (psb *PlaybackSessionBuilder) WithRecordingID(recordingID uuid.UUID) *PlaybackSessionBuilder {
	psb.session.RecordingID = recordingID
	return psb
}

// WithUserID sets the user ID
func (psb *PlaybackSessionBuilder) WithUserID(userID uuid.UUID) *PlaybackSessionBuilder {
	psb.session.UserID = userID
	return psb
}

// WithDuration sets the total duration
func (psb *PlaybackSessionBuilder) WithDuration(seconds int) *PlaybackSessionBuilder {
	psb.session.TotalDurationSeconds = seconds
	return psb
}

// WithWatchedDuration sets the watched duration
func (psb *PlaybackSessionBuilder) WithWatchedDuration(seconds int) *PlaybackSessionBuilder {
	psb.session.WatchedDurationSeconds = seconds
	return psb
}

// WithQuality sets the quality
func (psb *PlaybackSessionBuilder) WithQuality(quality string) *PlaybackSessionBuilder {
	psb.session.QualitySelected = quality
	return psb
}

// WithBufferEvents sets the buffer event count
func (psb *PlaybackSessionBuilder) WithBufferEvents(count int) *PlaybackSessionBuilder {
	psb.session.BufferEvents = count
	return psb
}

// Build returns the constructed session
func (psb *PlaybackSessionBuilder) Build() PlaybackSession {
	if psb.session.TotalDurationSeconds > 0 {
		psb.session.CompletionRate = (float64(psb.session.WatchedDurationSeconds) / float64(psb.session.TotalDurationSeconds)) * 100
	}
	return psb.session
}
