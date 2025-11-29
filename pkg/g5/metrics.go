package g5

import (
	"context"
	"sort"
	"sync"
	"time"
)

// MetricsCollector collects and tracks 5G performance metrics
type MetricsCollector struct {
	mu                   sync.RWMutex
	sessionMetrics       map[string]*SessionMetrics
	globalMetrics        *GlobalMetrics
	client               *Client
	reportInterval       time.Duration
	stopChan             chan struct{}
	wg                   sync.WaitGroup
	metricsCallbacks     []func(*SessionMetrics)
	aggregationCallbacks []func(*GlobalMetrics)
}

// SessionMetrics tracks metrics for a specific session
type SessionMetrics struct {
	SessionID         string
	StartTime         time.Time
	EndTime           time.Time
	Duration          int64 // milliseconds
	SampleCount       int
	LatencySamples    []int64
	BandwidthSamples  []int64
	PacketLossSamples []float32
	JitterSamples     []int64
	AvgLatency        int64
	MinLatency        int64
	MaxLatency        int64
	AvgBandwidth      int64
	MinBandwidth      int64
	MaxBandwidth      int64
	AvgPacketLoss     float32
	AvgJitter         int64
	EdgeNodeID        string
	VideoQuality      string
	AudioCodec        string
	FrameRate         int
	Resolution        string
	BytesSent         int64
	BytesReceived     int64
	FramesDropped     int
	PacketsLost       int
	PacketsSent       int
	LastUpdate        time.Time
	Status            string
}

// GlobalMetrics tracks aggregate metrics across all sessions
type GlobalMetrics struct {
	TotalSessions       int
	ActiveSessions      int
	AvgSessionDuration  int64
	TotalBytesTransfer  int64
	GlobalAvgLatency    int64
	GlobalAvgBandwidth  int64
	GlobalAvgPacketLoss float32
	PeakConcurrent      int
	TimestampFrom       time.Time
	TimestampTo         time.Time
	CollectionDuration  int64
}

// MetricsSample is a single data point
type MetricsSample struct {
	Timestamp     time.Time
	SessionID     string
	Latency       int64
	Bandwidth     int64
	PacketLoss    float32
	Jitter        int64
	Quality       string
	BytesSent     int64
	BytesReceived int64
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(client *Client) *MetricsCollector {
	return &MetricsCollector{
		sessionMetrics:       make(map[string]*SessionMetrics),
		globalMetrics:        &GlobalMetrics{},
		client:               client,
		reportInterval:       10 * time.Second,
		stopChan:             make(chan struct{}),
		metricsCallbacks:     make([]func(*SessionMetrics), 0),
		aggregationCallbacks: make([]func(*GlobalMetrics), 0),
	}
}

// Start begins metrics collection
func (mc *MetricsCollector) Start(ctx context.Context) error {
	mc.wg.Add(1)
	go mc.collectionLoop(ctx)
	return nil
}

// Stop gracefully stops the metrics collector
func (mc *MetricsCollector) Stop() error {
	close(mc.stopChan)
	mc.wg.Wait()
	return nil
}

// StartSession begins tracking metrics for a session
func (mc *MetricsCollector) StartSession(sessionID string, edgeNodeID string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.sessionMetrics[sessionID] = &SessionMetrics{
		SessionID:         sessionID,
		StartTime:         time.Now(),
		EdgeNodeID:        edgeNodeID,
		Status:            "active",
		LatencySamples:    make([]int64, 0, 100),
		BandwidthSamples:  make([]int64, 0, 100),
		PacketLossSamples: make([]float32, 0, 100),
		JitterSamples:     make([]int64, 0, 100),
	}
}

// EndSession marks the end of a session
func (mc *MetricsCollector) EndSession(sessionID string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if metrics, exists := mc.sessionMetrics[sessionID]; exists {
		metrics.EndTime = time.Now()
		metrics.Duration = metrics.EndTime.Sub(metrics.StartTime).Milliseconds()
		metrics.Status = "completed"
		metrics.LastUpdate = time.Now()
	}
}

// RecordSample records a single metrics sample
func (mc *MetricsCollector) RecordSample(sample MetricsSample) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if metrics, exists := mc.sessionMetrics[sample.SessionID]; exists {
		metrics.LatencySamples = append(metrics.LatencySamples, sample.Latency)
		metrics.BandwidthSamples = append(metrics.BandwidthSamples, sample.Bandwidth)
		metrics.PacketLossSamples = append(metrics.PacketLossSamples, sample.PacketLoss)
		metrics.JitterSamples = append(metrics.JitterSamples, sample.Jitter)
		metrics.BytesSent += sample.BytesSent
		metrics.BytesReceived += sample.BytesReceived
		metrics.SampleCount++
		metrics.LastUpdate = time.Now()

		// Update aggregate values
		mc.recalculateSessionMetrics(metrics)
	}
}

// RecordLatency records a latency sample
func (mc *MetricsCollector) RecordLatency(sessionID string, latency int64) {
	mc.RecordSample(MetricsSample{
		Timestamp: time.Now(),
		SessionID: sessionID,
		Latency:   latency,
	})
}

// RecordBandwidth records a bandwidth sample
func (mc *MetricsCollector) RecordBandwidth(sessionID string, bandwidth int64) {
	mc.RecordSample(MetricsSample{
		Timestamp: time.Now(),
		SessionID: sessionID,
		Bandwidth: bandwidth,
	})
}

// RecordPacketLoss records a packet loss sample
func (mc *MetricsCollector) RecordPacketLoss(sessionID string, loss float32) {
	mc.RecordSample(MetricsSample{
		Timestamp:  time.Now(),
		SessionID:  sessionID,
		PacketLoss: loss,
	})
}

// RecordJitter records a jitter sample
func (mc *MetricsCollector) RecordJitter(sessionID string, jitter int64) {
	mc.RecordSample(MetricsSample{
		Timestamp: time.Now(),
		SessionID: sessionID,
		Jitter:    jitter,
	})
}

// RecordVideoQuality records video quality for a session
func (mc *MetricsCollector) RecordVideoQuality(sessionID string, quality string, frameRate int, resolution string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if metrics, exists := mc.sessionMetrics[sessionID]; exists {
		metrics.VideoQuality = quality
		metrics.FrameRate = frameRate
		metrics.Resolution = resolution
	}
}

// RecordAudioCodec records audio codec for a session
func (mc *MetricsCollector) RecordAudioCodec(sessionID string, codec string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if metrics, exists := mc.sessionMetrics[sessionID]; exists {
		metrics.AudioCodec = codec
	}
}

// RecordFrameDropped increments frame drop counter
func (mc *MetricsCollector) RecordFrameDropped(sessionID string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if metrics, exists := mc.sessionMetrics[sessionID]; exists {
		metrics.FramesDropped++
	}
}

// RecordPacketLost increments packet loss counter
func (mc *MetricsCollector) RecordPacketLost(sessionID string, count int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if metrics, exists := mc.sessionMetrics[sessionID]; exists {
		metrics.PacketsLost += count
	}
}

// RecordPacketSent increments packet sent counter
func (mc *MetricsCollector) RecordPacketSent(sessionID string, count int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if metrics, exists := mc.sessionMetrics[sessionID]; exists {
		metrics.PacketsSent += count
	}
}

// recalculateSessionMetrics updates aggregate values for a session
func (mc *MetricsCollector) recalculateSessionMetrics(metrics *SessionMetrics) {
	if metrics.SampleCount == 0 {
		return
	}

	// Calculate average latency
	var latencySum int64
	metrics.MinLatency = metrics.LatencySamples[0]
	metrics.MaxLatency = metrics.LatencySamples[0]
	for _, sample := range metrics.LatencySamples {
		latencySum += sample
		if sample < metrics.MinLatency {
			metrics.MinLatency = sample
		}
		if sample > metrics.MaxLatency {
			metrics.MaxLatency = sample
		}
	}
	if len(metrics.LatencySamples) > 0 {
		metrics.AvgLatency = latencySum / int64(len(metrics.LatencySamples))
	}

	// Calculate average bandwidth
	var bandwidthSum int64
	metrics.MinBandwidth = int64(^uint64(0) >> 1)
	metrics.MaxBandwidth = 0
	for _, sample := range metrics.BandwidthSamples {
		bandwidthSum += sample
		if sample < metrics.MinBandwidth {
			metrics.MinBandwidth = sample
		}
		if sample > metrics.MaxBandwidth {
			metrics.MaxBandwidth = sample
		}
	}
	if len(metrics.BandwidthSamples) > 0 {
		metrics.AvgBandwidth = bandwidthSum / int64(len(metrics.BandwidthSamples))
	}

	// Calculate average packet loss
	var lossSum float32
	for _, sample := range metrics.PacketLossSamples {
		lossSum += sample
	}
	if len(metrics.PacketLossSamples) > 0 {
		metrics.AvgPacketLoss = lossSum / float32(len(metrics.PacketLossSamples))
	}

	// Calculate average jitter
	var jitterSum int64
	for _, sample := range metrics.JitterSamples {
		jitterSum += sample
	}
	if len(metrics.JitterSamples) > 0 {
		metrics.AvgJitter = jitterSum / int64(len(metrics.JitterSamples))
	}
}

// GetSessionMetrics returns metrics for a specific session
func (mc *MetricsCollector) GetSessionMetrics(sessionID string) *SessionMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if metrics, exists := mc.sessionMetrics[sessionID]; exists {
		// Return a copy to avoid external modification
		copy := *metrics
		return &copy
	}
	return nil
}

// GetAllSessionMetrics returns metrics for all sessions
func (mc *MetricsCollector) GetAllSessionMetrics() map[string]*SessionMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make(map[string]*SessionMetrics)
	for sessionID, metrics := range mc.sessionMetrics {
		copy := *metrics
		result[sessionID] = &copy
	}
	return result
}

// GetGlobalMetrics returns aggregate metrics across all sessions
func (mc *MetricsCollector) GetGlobalMetrics() *GlobalMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	copy := *mc.globalMetrics
	return &copy
}

// collectionLoop runs the metrics collection and reporting loop
func (mc *MetricsCollector) collectionLoop(ctx context.Context) {
	defer mc.wg.Done()

	ticker := time.NewTicker(mc.reportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-mc.stopChan:
			return
		case <-ticker.C:
			mc.aggregateMetrics()
			mc.reportMetrics(ctx)
		}
	}
}

// aggregateMetrics calculates global metrics from all sessions
func (mc *MetricsCollector) aggregateMetrics() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	globalMetrics := &GlobalMetrics{
		TimestampFrom: time.Now().Add(-mc.reportInterval),
		TimestampTo:   time.Now(),
	}

	globalMetrics.TotalSessions = len(mc.sessionMetrics)
	globalMetrics.ActiveSessions = 0

	var totalLatency, totalBandwidth int64
	var totalPacketLoss float32
	latencyCount, bandwidthCount := 0, 0

	for _, metrics := range mc.sessionMetrics {
		if metrics.Status == "active" {
			globalMetrics.ActiveSessions++
			globalMetrics.PeakConcurrent = max(globalMetrics.PeakConcurrent, 1)
		}

		if metrics.Duration > 0 {
			if globalMetrics.AvgSessionDuration == 0 {
				globalMetrics.AvgSessionDuration = metrics.Duration
			} else {
				globalMetrics.AvgSessionDuration = (globalMetrics.AvgSessionDuration + metrics.Duration) / 2
			}
		}

		globalMetrics.TotalBytesTransfer += metrics.BytesSent + metrics.BytesReceived

		if metrics.AvgLatency > 0 {
			totalLatency += metrics.AvgLatency
			latencyCount++
		}
		if metrics.AvgBandwidth > 0 {
			totalBandwidth += metrics.AvgBandwidth
			bandwidthCount++
		}
		totalPacketLoss += metrics.AvgPacketLoss
	}

	if latencyCount > 0 {
		globalMetrics.GlobalAvgLatency = totalLatency / int64(latencyCount)
	}
	if bandwidthCount > 0 {
		globalMetrics.GlobalAvgBandwidth = totalBandwidth / int64(bandwidthCount)
	}
	if len(mc.sessionMetrics) > 0 {
		globalMetrics.GlobalAvgPacketLoss = totalPacketLoss / float32(len(mc.sessionMetrics))
	}

	globalMetrics.CollectionDuration = mc.reportInterval.Milliseconds()
	mc.globalMetrics = globalMetrics

	// Call aggregation callbacks
	for _, callback := range mc.aggregationCallbacks {
		go callback(globalMetrics)
	}
}

// reportMetrics sends metrics to the backend
func (mc *MetricsCollector) reportMetrics(ctx context.Context) {
	mc.mu.RLock()
	activeMetrics := make([]*SessionMetrics, 0)
	for _, metrics := range mc.sessionMetrics {
		if metrics.Status == "active" {
			activeMetrics = append(activeMetrics, metrics)
		}
	}
	mc.mu.RUnlock()

	for _, metrics := range activeMetrics {
		// Report to API
		_ = mc.client.ReportMetrics(ctx, &NetworkMetrics{
			SessionID:  metrics.SessionID,
			Timestamp:  time.Now(),
			AvgLatency: int(metrics.AvgLatency),
			Bandwidth:  int(metrics.AvgBandwidth),
			PacketLoss: float64(metrics.AvgPacketLoss),
		})

		// Call session callbacks
		mc.mu.RLock()
		callbacks := mc.metricsCallbacks
		mc.mu.RUnlock()

		for _, callback := range callbacks {
			go callback(metrics)
		}
	}
}

// RegisterSessionCallback registers a callback for session metric updates
func (mc *MetricsCollector) RegisterSessionCallback(callback func(*SessionMetrics)) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.metricsCallbacks = append(mc.metricsCallbacks, callback)
}

// RegisterAggregationCallback registers a callback for global metric aggregation
func (mc *MetricsCollector) RegisterAggregationCallback(callback func(*GlobalMetrics)) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.aggregationCallbacks = append(mc.aggregationCallbacks, callback)
}

// ClearSession removes a session from tracking (after saving)
func (mc *MetricsCollector) ClearSession(sessionID string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	delete(mc.sessionMetrics, sessionID)
}

// ClearOldSessions removes sessions older than the specified duration
func (mc *MetricsCollector) ClearOldSessions(olderThan time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	now := time.Now()
	for sessionID, metrics := range mc.sessionMetrics {
		if metrics.Status == "completed" && now.Sub(metrics.EndTime) > olderThan {
			delete(mc.sessionMetrics, sessionID)
		}
	}
}

// GetTopSessions returns sessions sorted by a metric (e.g., "latency", "bandwidth")
func (mc *MetricsCollector) GetTopSessions(sortBy string, limit int) []*SessionMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	sessions := make([]*SessionMetrics, 0, len(mc.sessionMetrics))
	for _, metrics := range mc.sessionMetrics {
		copy := *metrics
		sessions = append(sessions, &copy)
	}

	// Sort based on sortBy parameter
	if sortBy == "latency" {
		mc.sortSessionsByLatency(sessions)
	} else if sortBy == "bandwidth" {
		mc.sortSessionsByBandwidth(sessions)
	} else if sortBy == "duration" {
		mc.sortSessionsByDuration(sessions)
	}

	// Return top N
	if limit < len(sessions) {
		sessions = sessions[:limit]
	}

	return sessions
}

// sortSessionsByLatency sorts sessions by average latency (descending)
func (mc *MetricsCollector) sortSessionsByLatency(sessions []*SessionMetrics) {
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].AvgLatency > sessions[j].AvgLatency
	})
}

// sortSessionsByBandwidth sorts sessions by average bandwidth (descending)
func (mc *MetricsCollector) sortSessionsByBandwidth(sessions []*SessionMetrics) {
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].AvgBandwidth > sessions[j].AvgBandwidth
	})
}

// sortSessionsByDuration sorts sessions by duration (descending)
func (mc *MetricsCollector) sortSessionsByDuration(sessions []*SessionMetrics) {
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Duration > sessions[j].Duration
	})
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GetSessionCount returns the total number of tracked sessions
func (mc *MetricsCollector) GetSessionCount() int {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return len(mc.sessionMetrics)
}

// GetActiveSessionCount returns the number of active sessions
func (mc *MetricsCollector) GetActiveSessionCount() int {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	count := 0
	for _, metrics := range mc.sessionMetrics {
		if metrics.Status == "active" {
			count++
		}
	}
	return count
}
