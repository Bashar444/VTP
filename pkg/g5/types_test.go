package g5

import (
	"testing"
	"time"
)

func TestNetworkType(t *testing.T) {
	tests := []struct {
		name     string
		netType  NetworkType
		expected string
	}{
		{"5G network", Network5G, "5G"},
		{"4G network", Network4G, "4G"},
		{"WiFi network", NetworkWiFi, "WiFi"},
		{"LTE network", NetworkLTE, "LTE"},
		{"Unknown network", NetworkUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.netType) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.netType)
			}
		})
	}
}

func TestNetwork5GStatusStruct(t *testing.T) {
	network := Network5GStatus{
		Type:           NetworkType(Network5G),
		Latency:        25,
		Bandwidth:      45,
		SignalStrength: -90,
		Connected:      true,
		Timestamp:      time.Now(),
		RSRP:           -90,
		RSRQ:           -15,
		SINR:           20,
	}

	if network.Type != NetworkType(Network5G) {
		t.Errorf("expected Network5G, got %v", network.Type)
	}

	if !network.Connected {
		t.Errorf("expected Connected=true, got false")
	}

	if network.Latency != 25 {
		t.Errorf("expected latency 25, got %d", network.Latency)
	}

	if network.Bandwidth != 45 {
		t.Errorf("expected bandwidth 45, got %d", network.Bandwidth)
	}
}

func TestEdgeNodeStruct(t *testing.T) {
	node := EdgeNode{
		ID:          "node-1",
		Region:      "us-west",
		Country:     "USA",
		Endpoint:    "node1.edge.vtp.local",
		Latency:     15,
		Capacity:    1000,
		Load:        45.5,
		Status:      NodeOnline,
		LastChecked: time.Now(),
		Distance:    50,
		Available:   550,
	}

	if node.ID != "node-1" {
		t.Errorf("expected ID node-1, got %s", node.ID)
	}

	if node.Status != NodeOnline {
		t.Errorf("expected status NodeOnline, got %v", node.Status)
	}

	if node.Capacity != 1000 {
		t.Errorf("expected capacity 1000, got %d", node.Capacity)
	}

	if node.Available != 550 {
		t.Errorf("expected available 550, got %d", node.Available)
	}
}

func TestNodeStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   NodeStatus
		expected string
	}{
		{"online", NodeOnline, "online"},
		{"offline", NodeOffline, "offline"},
		{"degraded", NodeDegraded, "degraded"},
		{"maintenance", NodeMaintenance, "maintenance"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.status)
			}
		})
	}
}

func TestQualityProfile(t *testing.T) {
	profile := QualityProfile{
		Name:         "Ultra HD (4K)",
		Bitrate:      15000,
		Resolution:   "3840x2160",
		FPS:          60,
		Codec:        "VP9/H.265",
		MaxLatency:   20,
		MinBandwidth: 12000,
	}

	if profile.Name != "Ultra HD (4K)" {
		t.Errorf("expected name Ultra HD (4K), got %s", profile.Name)
	}

	if profile.Bitrate != 15000 {
		t.Errorf("expected bitrate 15000, got %d", profile.Bitrate)
	}

	if profile.MaxLatency != 20 {
		t.Errorf("expected max latency 20, got %d", profile.MaxLatency)
	}
}

func TestQualityLevelConstants(t *testing.T) {
	tests := []struct {
		name     string
		level    QualityLevel
		expected string
	}{
		{"ultra_hd", QualityUltraHD, "ultra_hd"},
		{"high_def", QualityHighDef, "high_def"},
		{"standard", QualityStandard, "standard"},
		{"medium", QualityMedium, "medium"},
		{"low", QualityLow, "low"},
		{"auto", QualityAuto, "auto"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.level) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.level)
			}
		})
	}
}

func TestNetworkMetricsStruct(t *testing.T) {
	metrics := NetworkMetrics{
		SessionID:           "session-123",
		Timestamp:           time.Now(),
		NetworkType:         Network5G,
		Latency:             25,
		AvgLatency:          28,
		MaxLatency:          50,
		MinLatency:          20,
		Bandwidth:           45,
		AvailableBandwidth:  40,
		PacketLoss:          0.1,
		Jitter:              5,
		BufferHealth:        95.0,
		ConnectionStability: 99.8,
	}

	if metrics.SessionID != "session-123" {
		t.Errorf("expected session-123, got %s", metrics.SessionID)
	}

	if metrics.NetworkType != Network5G {
		t.Errorf("expected Network5G, got %v", metrics.NetworkType)
	}

	if metrics.PacketLoss != 0.1 {
		t.Errorf("expected packet loss 0.1, got %f", metrics.PacketLoss)
	}
}

func TestEdgeMetricsStruct(t *testing.T) {
	edgeMetrics := EdgeMetrics{
		NodeID:      "node-1",
		Timestamp:   time.Now(),
		Latency:     10,
		Load:        65.5,
		CPUUsage:    42.3,
		MemoryUsage: 58.2,
		Connections: 234,
		Throughput:  500,
	}

	if edgeMetrics.NodeID != "node-1" {
		t.Errorf("expected node-1, got %s", edgeMetrics.NodeID)
	}

	if edgeMetrics.Latency != 10 {
		t.Errorf("expected latency 10, got %d", edgeMetrics.Latency)
	}

	if edgeMetrics.Connections != 234 {
		t.Errorf("expected connections 234, got %d", edgeMetrics.Connections)
	}
}

func TestAdaptiveStrategyStruct(t *testing.T) {
	strategy := AdaptiveStrategy{
		Enabled:         true,
		TargetLatency:   50,
		TargetBandwidth: 25,
		MinBitrate:      800,
		MaxBitrate:      15000,
		SwitchThreshold: 20.0,
		CheckInterval:   5000,
	}

	if !strategy.Enabled {
		t.Errorf("expected Enabled=true, got false")
	}

	if strategy.TargetLatency != 50 {
		t.Errorf("expected target latency 50, got %d", strategy.TargetLatency)
	}

	if strategy.SwitchThreshold != 20.0 {
		t.Errorf("expected switch threshold 20.0, got %f", strategy.SwitchThreshold)
	}
}

func TestConfigStruct(t *testing.T) {
	cfg := Config{
		Enabled:            true,
		DetectionInterval:  1000,
		MetricsInterval:    10000,
		EdgeCheckInterval:  30000,
		PreferredEdgeNode:  "node-1",
		AllowNetworkSwitch: true,
		MaxLatencyTarget:   50,
		MinBandwidthTarget: 20,
	}

	if !cfg.Enabled {
		t.Errorf("expected Enabled=true, got false")
	}

	if cfg.DetectionInterval != 1000 {
		t.Errorf("expected detection interval 1000, got %d", cfg.DetectionInterval)
	}

	if cfg.PreferredEdgeNode != "node-1" {
		t.Errorf("expected preferred node node-1, got %s", cfg.PreferredEdgeNode)
	}
}

func TestDetectionResultStruct(t *testing.T) {
	result := DetectionResult{
		Detected:       true,
		NetworkType:    Network5G,
		Latency:        25,
		Bandwidth:      45,
		SignalStrength: -90,
		Timestamp:      time.Now(),
		Error:          "",
	}

	if !result.Detected {
		t.Errorf("expected Detected=true, got false")
	}

	if result.NetworkType != Network5G {
		t.Errorf("expected Network5G, got %v", result.NetworkType)
	}

	if result.Error != "" {
		t.Errorf("expected no error, got %s", result.Error)
	}
}

func TestQualityAdjustmentStruct(t *testing.T) {
	adjustment := QualityAdjustment{
		FromProfile: QualityStandard,
		ToProfile:   QualityHighDef,
		Reason:      "network improved",
		Timestamp:   time.Now(),
		Duration:    100,
	}

	if adjustment.FromProfile != QualityStandard {
		t.Errorf("expected QualityStandard, got %v", adjustment.FromProfile)
	}

	if adjustment.ToProfile != QualityHighDef {
		t.Errorf("expected QualityHighDef, got %v", adjustment.ToProfile)
	}

	if adjustment.Duration != 100 {
		t.Errorf("expected duration 100, got %d", adjustment.Duration)
	}
}

func TestHealthCheckStruct(t *testing.T) {
	healthCheck := HealthCheck{
		NodeID:    "node-1",
		Status:    NodeOnline,
		Latency:   15,
		LastCheck: time.Now(),
		Errors:    0,
		Success:   true,
		Message:   "node is healthy",
	}

	if healthCheck.NodeID != "node-1" {
		t.Errorf("expected node-1, got %s", healthCheck.NodeID)
	}

	if healthCheck.Status != NodeOnline {
		t.Errorf("expected NodeOnline, got %v", healthCheck.Status)
	}

	if !healthCheck.Success {
		t.Errorf("expected Success=true, got false")
	}
}

func TestStatisticsStruct(t *testing.T) {
	stats := Statistics{
		Period:              time.Hour,
		TotalSessions:       100,
		AvgLatency:          30,
		MaxLatency:          150,
		MinLatency:          15,
		AvgBandwidth:        35,
		TotalPacketsLost:    50,
		AveragePacketLoss:   0.5,
		BufferUnderflows:    2,
		QualitySwitches:     15,
		EdgeNodeFailovers:   1,
		ConnectionStability: 99.5,
		AvgPlaybackTime:     3600,
	}

	if stats.TotalSessions != 100 {
		t.Errorf("expected 100 sessions, got %d", stats.TotalSessions)
	}

	if stats.AvgLatency != 30 {
		t.Errorf("expected avg latency 30, got %d", stats.AvgLatency)
	}

	if stats.ConnectionStability != 99.5 {
		t.Errorf("expected connection stability 99.5, got %f", stats.ConnectionStability)
	}
}

func TestRequestOptionsStruct(t *testing.T) {
	opts := RequestOptions{
		Timeout:    10 * time.Second,
		RetryCount: 3,
		RetryDelay: 1 * time.Second,
		UseEdge:    true,
		EdgeNodeID: "node-1",
	}

	if opts.Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", opts.Timeout)
	}

	if opts.RetryCount != 3 {
		t.Errorf("expected retry count 3, got %d", opts.RetryCount)
	}

	if !opts.UseEdge {
		t.Errorf("expected UseEdge=true, got false")
	}
}

func TestResponseErrorStruct(t *testing.T) {
	respErr := ResponseError{
		Code:      "NETWORK_UNAVAILABLE",
		Message:   "5G network not available",
		Details:   map[string]interface{}{"region": "us-west"},
		Timestamp: time.Now(),
	}

	if respErr.Code != "NETWORK_UNAVAILABLE" {
		t.Errorf("expected NETWORK_UNAVAILABLE, got %s", respErr.Code)
	}

	if respErr.Message != "5G network not available" {
		t.Errorf("expected message, got %s", respErr.Message)
	}

	if respErr.Details["region"] != "us-west" {
		t.Errorf("expected region us-west, got %v", respErr.Details["region"])
	}
}
