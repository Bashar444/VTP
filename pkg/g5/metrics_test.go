package g5

import (
	"context"
	"testing"
	"time"
)

func TestNewMetricsCollector(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	collector := NewMetricsCollector(client)

	if collector == nil {
		t.Fatal("expected collector instance, got nil")
	}

	if collector.client != client {
		t.Errorf("expected client to be set")
	}

	if collector.reportInterval != 10*time.Second {
		t.Errorf("expected default report interval 10s")
	}
}

func TestMetricsCollectorStart(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	collector := NewMetricsCollector(client)
	ctx := context.Background()

	err := collector.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	err = collector.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

func TestMetricsCollectorStartSession(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	collector := NewMetricsCollector(client)

	collector.StartSession("session-1", "node-1")

	metrics := collector.GetSessionMetrics("session-1")
	if metrics == nil {
		t.Errorf("expected session metrics, got nil")
	}

	if metrics.SessionID != "session-1" {
		t.Errorf("expected session-1, got %s", metrics.SessionID)
	}

	if metrics.Status != "active" {
		t.Errorf("expected status active, got %s", metrics.Status)
	}
}

func TestMetricsCollectorEndSession(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	collector := NewMetricsCollector(client)

	collector.StartSession("session-1", "node-1")
	collector.EndSession("session-1")

	metrics := collector.GetSessionMetrics("session-1")
	if metrics == nil {
		t.Errorf("expected session metrics, got nil")
	}

	if metrics.Status != "completed" {
		t.Errorf("expected status completed, got %s", metrics.Status)
	}

	if metrics.Duration == 0 {
		t.Errorf("expected non-zero duration")
	}
}

func TestMetricsCollectorRecordSample(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	collector := NewMetricsCollector(client)

	collector.StartSession("session-1", "node-1")

	sample := MetricsSample{
		Timestamp:  time.Now(),
		SessionID:  "session-1",
		Latency:    25,
		Bandwidth:  45,
		PacketLoss: 0.1,
		Jitter:     5,
	}

	collector.RecordSample(sample)

	metrics := collector.GetSessionMetrics("session-1")
	if metrics == nil {
		t.Errorf("expected metrics, got nil")
	}

	if metrics.SampleCount != 1 {
		t.Errorf("expected 1 sample, got %d", metrics.SampleCount)
	}
}

func TestMetricsCollectorRecordLatency(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	collector := NewMetricsCollector(client)

	collector.StartSession("session-1", "node-1")
	collector.RecordLatency("session-1", 25)

	metrics := collector.GetSessionMetrics("session-1")
	if metrics == nil {
		t.Errorf("expected metrics, got nil")
	}

	if metrics.SampleCount != 1 {
		t.Errorf("expected 1 sample, got %d", metrics.SampleCount)
	}
}

func TestMetricsCollectorRecordBandwidth(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	collector := NewMetricsCollector(client)

	collector.StartSession("session-1", "node-1")
	collector.RecordBandwidth("session-1", 45)

	metrics := collector.GetSessionMetrics("session-1")
	if metrics == nil {
		t.Errorf("expected metrics, got nil")
	}

	if metrics.SampleCount != 1 {
		t.Errorf("expected 1 sample, got %d", metrics.SampleCount)
	}
}

func TestMetricsCollectorRecordVideoQuality(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	collector := NewMetricsCollector(client)

	collector.StartSession("session-1", "node-1")
	collector.RecordVideoQuality("session-1", "1080p", 30, "1920x1080")

	metrics := collector.GetSessionMetrics("session-1")
	if metrics == nil {
		t.Errorf("expected metrics, got nil")
	}

	if metrics.VideoQuality != "1080p" {
		t.Errorf("expected 1080p, got %s", metrics.VideoQuality)
	}

	if metrics.FrameRate != 30 {
		t.Errorf("expected frame rate 30, got %d", metrics.FrameRate)
	}

	if metrics.Resolution != "1920x1080" {
		t.Errorf("expected 1920x1080, got %s", metrics.Resolution)
	}
}

func TestMetricsCollectorRecordAudioCodec(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	collector := NewMetricsCollector(client)

	collector.StartSession("session-1", "node-1")
	collector.RecordAudioCodec("session-1", "AAC")

	metrics := collector.GetSessionMetrics("session-1")
	if metrics == nil {
		t.Errorf("expected metrics, got nil")
	}

	if metrics.AudioCodec != "AAC" {
		t.Errorf("expected AAC, got %s", metrics.AudioCodec)
	}
}

func TestMetricsCollectorRecordFrameDropped(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	collector := NewMetricsCollector(client)

	collector.StartSession("session-1", "node-1")
	collector.RecordFrameDropped("session-1")
	collector.RecordFrameDropped("session-1")

	metrics := collector.GetSessionMetrics("session-1")
	if metrics == nil {
		t.Errorf("expected metrics, got nil")
	}

	if metrics.FramesDropped != 2 {
		t.Errorf("expected 2 dropped frames, got %d", metrics.FramesDropped)
	}
}

func TestMetricsCollectorRecordPacketLost(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	collector := NewMetricsCollector(client)

	collector.StartSession("session-1", "node-1")
	collector.RecordPacketLost("session-1", 5)

	metrics := collector.GetSessionMetrics("session-1")
	if metrics == nil {
		t.Errorf("expected metrics, got nil")
	}

	if metrics.PacketsLost != 5 {
		t.Errorf("expected 5 packets lost, got %d", metrics.PacketsLost)
	}
}

func TestMetricsCollectorRecordPacketSent(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	collector := NewMetricsCollector(client)

	collector.StartSession("session-1", "node-1")
	collector.RecordPacketSent("session-1", 100)

	metrics := collector.GetSessionMetrics("session-1")
	if metrics == nil {
		t.Errorf("expected metrics, got nil")
	}

	if metrics.PacketsSent != 100 {
		t.Errorf("expected 100 packets sent, got %d", metrics.PacketsSent)
	}
}

func TestMetricsCollectorGetSessionMetrics(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	collector := NewMetricsCollector(client)

	collector.StartSession("session-1", "node-1")

	metrics := collector.GetSessionMetrics("session-1")
	if metrics == nil {
		t.Errorf("expected metrics, got nil")
	}

	// Test nonexistent session
	metrics = collector.GetSessionMetrics("nonexistent")
	if metrics != nil {
		t.Errorf("expected nil for nonexistent session")
	}
}

func TestMetricsCollectorGetAllSessionMetrics(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	collector := NewMetricsCollector(client)

	collector.StartSession("session-1", "node-1")
	collector.StartSession("session-2", "node-2")

	allMetrics := collector.GetAllSessionMetrics()
	if len(allMetrics) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(allMetrics))
	}
}

func TestMetricsCollectorGetGlobalMetrics(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	collector := NewMetricsCollector(client)

	collector.StartSession("session-1", "node-1")

	globalMetrics := collector.GetGlobalMetrics()
	if globalMetrics == nil {
		t.Errorf("expected global metrics, got nil")
	}
}

func TestMetricsCollectorGetSessionCount(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	collector := NewMetricsCollector(client)

	collector.StartSession("session-1", "node-1")
	collector.StartSession("session-2", "node-2")

	count := collector.GetSessionCount()
	if count != 2 {
		t.Errorf("expected 2 sessions, got %d", count)
	}
}

func TestMetricsCollectorGetActiveSessionCount(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	collector := NewMetricsCollector(client)

	collector.StartSession("session-1", "node-1")
	collector.StartSession("session-2", "node-2")
	collector.EndSession("session-2")

	count := collector.GetActiveSessionCount()
	if count != 1 {
		t.Errorf("expected 1 active session, got %d", count)
	}
}

func TestMetricsCollectorClearSession(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	collector := NewMetricsCollector(client)

	collector.StartSession("session-1", "node-1")
	collector.ClearSession("session-1")

	metrics := collector.GetSessionMetrics("session-1")
	if metrics != nil {
		t.Errorf("expected nil after clearing session")
	}
}

func TestMetricsCollectorRegisterSessionCallback(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	collector := NewMetricsCollector(client)

	callback := func(metrics *SessionMetrics) {
		// Callback will be invoked when metrics are recorded
	}

	collector.RegisterSessionCallback(callback)

	if len(collector.metricsCallbacks) == 0 {
		t.Errorf("expected callback to be registered")
	}
}

func TestMetricsCollectorRegisterAggregationCallback(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	collector := NewMetricsCollector(client)

	callback := func(metrics *GlobalMetrics) {
		// Callback will be invoked when aggregation completes
	}

	collector.RegisterAggregationCallback(callback)

	if len(collector.aggregationCallbacks) == 0 {
		t.Errorf("expected callback to be registered")
	}
}

func TestMetricsCollectorMultipleSamples(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	collector := NewMetricsCollector(client)

	collector.StartSession("session-1", "node-1")

	// Record multiple samples
	for i := 0; i < 10; i++ {
		collector.RecordLatency("session-1", int64(20+i))
		collector.RecordBandwidth("session-1", int64(40+i))
	}

	metrics := collector.GetSessionMetrics("session-1")
	if metrics == nil {
		t.Errorf("expected metrics, got nil")
	}

	if metrics.SampleCount != 10 {
		t.Errorf("expected 10 samples, got %d", metrics.SampleCount)
	}

	if metrics.AvgLatency <= 0 {
		t.Errorf("expected positive average latency")
	}
}

func TestMetricsCollectorConcurrency(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	collector := NewMetricsCollector(client)

	collector.StartSession("session-1", "node-1")
	collector.StartSession("session-2", "node-2")

	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func(sessionID string) {
			for j := 0; j < 10; j++ {
				collector.RecordLatency(sessionID, int64(20+j))
				collector.RecordBandwidth(sessionID, int64(40+j))
				time.Sleep(time.Millisecond)
			}
			done <- true
		}("session-" + string(rune(i%2+1)))
	}

	for i := 0; i < 5; i++ {
		<-done
	}
}
