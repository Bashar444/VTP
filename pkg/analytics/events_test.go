package analytics

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewEventCollector(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	collector := NewEventCollector(100, 5*time.Second, logger)

	if collector == nil {
		t.Error("Expected event collector to be created")
	}
	if collector.maxBatchSize != 100 {
		t.Errorf("Expected max batch size 100, got %d", collector.maxBatchSize)
	}
	if collector.batchTimeout != 5*time.Second {
		t.Errorf("Expected batch timeout 5s, got %v", collector.batchTimeout)
	}
}

func TestRecordEvent(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	collector := NewEventCollector(100, 5*time.Second, logger)
	defer collector.Stop()

	recordingID := uuid.New()
	userID := uuid.New()
	sessionID := "session-001"
	metadata := map[string]interface{}{"bitrate": 2000, "quality": "High"}

	err := collector.RecordEvent(EventPlaybackStarted, recordingID, userID, sessionID, metadata)
	if err != nil {
		t.Errorf("Failed to record event: %v", err)
	}

	events := collector.GetPendingEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 pending event, got %d", len(events))
	}

	if events[0].EventType != EventPlaybackStarted {
		t.Errorf("Expected EventPlaybackStarted, got %v", events[0].EventType)
	}
	if events[0].RecordingID != recordingID {
		t.Errorf("Expected recording ID %v, got %v", recordingID, events[0].RecordingID)
	}
	if events[0].UserID != userID {
		t.Errorf("Expected user ID %v, got %v", userID, events[0].UserID)
	}
	if events[0].SessionID != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, events[0].SessionID)
	}
}

func TestEventBuilder(t *testing.T) {
	recordingID := uuid.New()
	userID := uuid.New()
	sessionID := "session-001"

	event := NewEventBuilder().
		WithType(EventQualityChanged).
		WithRecordingID(recordingID).
		WithUserID(userID).
		WithSessionID(sessionID).
		WithMetadata("old_bitrate", 1000).
		WithMetadata("new_bitrate", 2000).
		Build()

	if event.EventType != EventQualityChanged {
		t.Errorf("Expected EventQualityChanged, got %v", event.EventType)
	}
	if event.RecordingID != recordingID {
		t.Errorf("Expected recording ID %v, got %v", recordingID, event.RecordingID)
	}
	if event.UserID != userID {
		t.Errorf("Expected user ID %v, got %v", userID, event.UserID)
	}
	if event.SessionID != sessionID {
		t.Errorf("Expected session ID %s, got %s", sessionID, event.SessionID)
	}

	oldBitrate, ok := event.Metadata["old_bitrate"]
	if !ok || oldBitrate != 1000 {
		t.Error("Expected old_bitrate metadata")
	}
	newBitrate, ok := event.Metadata["new_bitrate"]
	if !ok || newBitrate != 2000 {
		t.Error("Expected new_bitrate metadata")
	}
}

func TestEventValidator(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	validator := NewEventValidator(logger)

	// Valid event
	validEvent := NewEventBuilder().
		WithType(EventPlaybackStarted).
		WithRecordingID(uuid.New()).
		WithUserID(uuid.New()).
		WithSessionID("session-001").
		Build()

	if err := validator.ValidateEvent(validEvent); err != nil {
		t.Errorf("Expected valid event, got error: %v", err)
	}

	// Invalid event - missing event type
	invalidEvent := AnalyticsEvent{
		EventID:     uuid.New(),
		RecordingID: uuid.New(),
		UserID:      uuid.New(),
		SessionID:   "session-001",
		Timestamp:   time.Now(),
		CreatedAt:   time.Now(),
	}

	if err := validator.ValidateEvent(invalidEvent); err == nil {
		t.Error("Expected invalid event error")
	}

	// Invalid event - missing recording ID
	invalidEvent2 := NewEventBuilder().
		WithType(EventPlaybackStarted).
		WithUserID(uuid.New()).
		WithSessionID("session-001").
		Build()

	if err := validator.ValidateEvent(invalidEvent2); err == nil {
		t.Error("Expected invalid event error for missing recording ID")
	}
}

func TestEventSerializer(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	serializer := NewEventSerializer(logger)

	event := NewEventBuilder().
		WithType(EventPlaybackStarted).
		WithRecordingID(uuid.New()).
		WithUserID(uuid.New()).
		WithSessionID("session-001").
		WithMetadata("bitrate", 2000).
		Build()

	// Serialize
	data, err := serializer.SerializeEvent(event)
	if err != nil {
		t.Errorf("Failed to serialize event: %v", err)
	}

	// Deserialize
	deserializedEvent, err := serializer.DeserializeEvent(data)
	if err != nil {
		t.Errorf("Failed to deserialize event: %v", err)
	}

	if deserializedEvent.EventType != event.EventType {
		t.Errorf("Expected event type %v, got %v", event.EventType, deserializedEvent.EventType)
	}
	if deserializedEvent.RecordingID != event.RecordingID {
		t.Errorf("Expected recording ID %v, got %v", event.RecordingID, deserializedEvent.RecordingID)
	}
}

func TestPlaybackSessionBuilder(t *testing.T) {
	recordingID := uuid.New()
	userID := uuid.New()

	session := NewPlaybackSessionBuilder().
		WithRecordingID(recordingID).
		WithUserID(userID).
		WithDuration(3600).        // 60 minutes
		WithWatchedDuration(3420). // 57 minutes
		WithQuality("1080p").
		WithBufferEvents(2).
		Build()

	if session.RecordingID != recordingID {
		t.Errorf("Expected recording ID %v, got %v", recordingID, session.RecordingID)
	}
	if session.UserID != userID {
		t.Errorf("Expected user ID %v, got %v", userID, session.UserID)
	}
	if session.TotalDurationSeconds != 3600 {
		t.Errorf("Expected duration 3600, got %d", session.TotalDurationSeconds)
	}
	if session.WatchedDurationSeconds != 3420 {
		t.Errorf("Expected watched duration 3420, got %d", session.WatchedDurationSeconds)
	}
	if session.QualitySelected != "1080p" {
		t.Errorf("Expected quality 1080p, got %s", session.QualitySelected)
	}
	if session.BufferEvents != 2 {
		t.Errorf("Expected buffer events 2, got %d", session.BufferEvents)
	}

	// Check completion rate (3420/3600 = 0.95 = 95%)
	expectedCompletion := (float64(3420) / float64(3600)) * 100
	if session.CompletionRate != expectedCompletion {
		t.Errorf("Expected completion rate %.2f, got %.2f", expectedCompletion, session.CompletionRate)
	}
}

func TestBatchProcessing(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	collector := NewEventCollector(3, 100*time.Millisecond, logger)
	defer collector.Stop()

	batchCount := 0
	collector.SetBatchCallback(func(events []AnalyticsEvent) error {
		batchCount++
		if len(events) != 3 {
			t.Errorf("Expected batch size 3, got %d", len(events))
		}
		return nil
	})

	// Record 3 events - should trigger batch immediately
	for i := 0; i < 3; i++ {
		err := collector.RecordEvent(EventPlaybackStarted, uuid.New(), uuid.New(), "session-001", nil)
		if err != nil {
			t.Errorf("Failed to record event: %v", err)
		}
	}

	// Wait for batch processing
	time.Sleep(200 * time.Millisecond)

	if batchCount == 0 {
		t.Error("Expected batch to be processed")
	}

	// Check that pending events are now empty
	events := collector.GetPendingEvents()
	if len(events) != 0 {
		t.Errorf("Expected 0 pending events after batch, got %d", len(events))
	}
}

func TestMultipleEventTypes(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	collector := NewEventCollector(100, 5*time.Second, logger)
	defer collector.Stop()

	recordingID := uuid.New()
	userID := uuid.New()
	sessionID := "session-001"

	eventTypes := []EventType{
		EventPlaybackStarted,
		EventSegmentRequested,
		EventQualitySelected,
		EventBufferEvent,
		EventQualityChanged,
		EventPlaybackPaused,
		EventPlaybackResumed,
		EventPlaybackStopped,
	}

	for _, eventType := range eventTypes {
		err := collector.RecordEvent(eventType, recordingID, userID, sessionID, nil)
		if err != nil {
			t.Errorf("Failed to record event type %v: %v", eventType, err)
		}
	}

	events := collector.GetPendingEvents()
	if len(events) != len(eventTypes) {
		t.Errorf("Expected %d pending events, got %d", len(eventTypes), len(events))
	}

	for i, eventType := range eventTypes {
		if events[i].EventType != eventType {
			t.Errorf("Expected event type %v at index %d, got %v", eventType, i, events[i].EventType)
		}
	}
}

func TestEventMetadataVariations(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	collector := NewEventCollector(100, 5*time.Second, logger)
	defer collector.Stop()

	recordingID := uuid.New()
	userID := uuid.New()

	testCases := []struct {
		name     string
		metadata map[string]interface{}
	}{
		{
			name: "Quality metadata",
			metadata: map[string]interface{}{
				"bitrate":    2000,
				"resolution": "1080p",
				"source":     "adaptive",
			},
		},
		{
			name: "Buffer metadata",
			metadata: map[string]interface{}{
				"duration_ms": 500,
				"position_ms": 30000,
			},
		},
		{
			name:     "Empty metadata",
			metadata: map[string]interface{}{},
		},
	}

	for _, tc := range testCases {
		err := collector.RecordEvent(EventPlaybackStarted, recordingID, userID, "session-001", tc.metadata)
		if err != nil {
			t.Errorf("%s: Failed to record event: %v", tc.name, err)
		}
	}

	events := collector.GetPendingEvents()
	if len(events) != len(testCases) {
		t.Errorf("Expected %d events, got %d", len(testCases), len(events))
	}
}

func TestEventTimestamps(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	collector := NewEventCollector(100, 5*time.Second, logger)
	defer collector.Stop()

	now := time.Now()

	event := NewEventBuilder().
		WithType(EventPlaybackStarted).
		WithRecordingID(uuid.New()).
		WithUserID(uuid.New()).
		WithSessionID("session-001").
		WithTimestamp(now).
		Build()

	err := collector.AddEvent(event)
	if err != nil {
		t.Errorf("Failed to add event: %v", err)
	}

	events := collector.GetPendingEvents()
	if len(events) == 0 {
		t.Error("Expected event in pending queue")
	}

	if !events[0].Timestamp.Equal(now) {
		t.Errorf("Expected timestamp %v, got %v", now, events[0].Timestamp)
	}
}

func TestValidateMultipleEvents(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	validator := NewEventValidator(logger)

	events := []AnalyticsEvent{
		NewEventBuilder().
			WithType(EventPlaybackStarted).
			WithRecordingID(uuid.New()).
			WithUserID(uuid.New()).
			WithSessionID("session-001").
			Build(),
		NewEventBuilder().
			WithType(EventQualityChanged).
			WithRecordingID(uuid.New()).
			WithUserID(uuid.New()).
			WithSessionID("session-001").
			Build(),
		NewEventBuilder().
			WithType(EventBufferEvent).
			WithRecordingID(uuid.New()).
			WithUserID(uuid.New()).
			WithSessionID("session-001").
			Build(),
	}

	errors := validator.ValidateEvents(events)
	if len(errors) != 0 {
		t.Errorf("Expected no validation errors, got %d", len(errors))
	}
}

func TestSerializeMultipleEvents(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	serializer := NewEventSerializer(logger)

	events := make([]AnalyticsEvent, 5)
	for i := 0; i < 5; i++ {
		events[i] = NewEventBuilder().
			WithType(EventPlaybackStarted).
			WithRecordingID(uuid.New()).
			WithUserID(uuid.New()).
			WithSessionID("session-001").
			Build()
	}

	data, err := serializer.SerializeEvents(events)
	if err != nil {
		t.Errorf("Failed to serialize events: %v", err)
	}

	deserializedEvents, err := serializer.DeserializeEvents(data)
	if err != nil {
		t.Errorf("Failed to deserialize events: %v", err)
	}

	if len(deserializedEvents) != len(events) {
		t.Errorf("Expected %d events, got %d", len(events), len(deserializedEvents))
	}
}
