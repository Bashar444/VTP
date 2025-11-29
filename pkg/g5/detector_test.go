package g5

import (
	"context"
	"testing"
	"time"
)

func TestNewNetworkDetector(t *testing.T) {
	cfg := &Config{
		Enabled:           true,
		DetectionInterval: 1000,
		MaxLatencyTarget:  50,
	}

	detector := NewNetworkDetector(cfg)

	if detector == nil {
		t.Fatal("expected detector instance, got nil")
	}

	if detector.config != cfg {
		t.Errorf("expected config to be set")
	}

	if detector.currentNetwork == nil {
		t.Errorf("expected currentNetwork initialized, got nil")
	}

	if detector.isRunning {
		t.Errorf("expected isRunning=false, got true")
	}
}

func TestDetectorStart(t *testing.T) {
	cfg := &Config{
		Enabled:           true,
		DetectionInterval: 1000,
		MaxLatencyTarget:  50,
	}

	detector := NewNetworkDetector(cfg)
	ctx := context.Background()

	err := detector.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !detector.isRunning {
		t.Errorf("expected isRunning=true after Start")
	}

	err = detector.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

func TestDetectorStartAlreadyRunning(t *testing.T) {
	cfg := &Config{
		Enabled:           true,
		DetectionInterval: 1000,
		MaxLatencyTarget:  50,
	}

	detector := NewNetworkDetector(cfg)
	ctx := context.Background()

	err := detector.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Try to start again
	err = detector.Start(ctx)
	if err == nil {
		t.Errorf("expected error when starting twice, got nil")
	}

	detector.Stop()
}

func TestDetectorStartNoConfig(t *testing.T) {
	detector := NewNetworkDetector(nil)
	ctx := context.Background()

	err := detector.Start(ctx)
	if err == nil {
		t.Errorf("expected error with no config, got nil")
	}
}

func TestDetectorStop(t *testing.T) {
	cfg := &Config{
		Enabled:           true,
		DetectionInterval: 1000,
		MaxLatencyTarget:  50,
	}

	detector := NewNetworkDetector(cfg)
	ctx := context.Background()

	err := detector.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	err = detector.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if detector.isRunning {
		t.Errorf("expected isRunning=false after Stop")
	}
}

func TestDetectorStopNotRunning(t *testing.T) {
	cfg := &Config{
		Enabled:           true,
		DetectionInterval: 1000,
		MaxLatencyTarget:  50,
	}

	detector := NewNetworkDetector(cfg)

	err := detector.Stop()
	if err == nil {
		t.Errorf("expected error when stopping non-running detector, got nil")
	}
}

func TestDetectNetwork(t *testing.T) {
	cfg := &Config{
		Enabled:           true,
		DetectionInterval: 1000,
		MaxLatencyTarget:  50,
	}

	detector := NewNetworkDetector(cfg)
	ctx := context.Background()

	result, err := detector.DetectNetwork(ctx)
	if err != nil {
		t.Fatalf("DetectNetwork failed: %v", err)
	}

	if result == nil {
		t.Errorf("expected detection result, got nil")
	}

	if result.Timestamp.IsZero() {
		t.Errorf("expected non-zero timestamp")
	}
}

func TestGetCurrentNetwork(t *testing.T) {
	cfg := &Config{
		Enabled:           true,
		DetectionInterval: 1000,
		MaxLatencyTarget:  50,
	}

	detector := NewNetworkDetector(cfg)

	network := detector.GetCurrentNetwork()
	if network == nil {
		t.Errorf("expected network, got nil")
	}

	if network.Type != NetworkUnknown {
		t.Errorf("expected NetworkUnknown initially, got %v", network.Type)
	}

	if network.Connected {
		t.Errorf("expected Connected=false initially, got true")
	}
}

func TestIs5GAvailable(t *testing.T) {
	cfg := &Config{
		Enabled:           true,
		DetectionInterval: 1000,
		MaxLatencyTarget:  50,
	}

	detector := NewNetworkDetector(cfg)

	// Initially not available
	if detector.Is5GAvailable() {
		t.Errorf("expected 5G not available initially")
	}

	// Set up a 5G network
	detector.currentNetwork = &Network5GStatus{
		Type:      Network5G,
		Latency:   25,
		Bandwidth: 45,
		Connected: true,
	}

	if !detector.Is5GAvailable() {
		t.Errorf("expected 5G to be available")
	}

	// Test with high latency
	detector.currentNetwork.Latency = 100
	if detector.Is5GAvailable() {
		t.Errorf("expected 5G not available with high latency")
	}
}

func TestGetNetworkQuality(t *testing.T) {
	cfg := &Config{
		Enabled:           true,
		DetectionInterval: 1000,
		MaxLatencyTarget:  50,
	}

	detector := NewNetworkDetector(cfg)

	// Initially should be 0
	quality := detector.GetNetworkQuality()
	if quality != 0 {
		t.Errorf("expected quality 0 initially, got %d", quality)
	}

	// Set up a good network
	detector.currentNetwork = &Network5GStatus{
		Type:      NetworkType(Network5G),
		Latency:   25,
		Bandwidth: 45,
		Connected: true,
	}

	quality = detector.GetNetworkQuality()
	if quality <= 0 {
		t.Errorf("expected positive quality, got %d", quality)
	}

	if quality > 100 {
		t.Errorf("expected quality <= 100, got %d", quality)
	}
}

func TestSetMetricsCallback(t *testing.T) {
	cfg := &Config{
		Enabled:           true,
		DetectionInterval: 1000,
		MaxLatencyTarget:  50,
	}

	detector := NewNetworkDetector(cfg)

	callbackCalled := false
	callback := func(network *Network5GStatus) {
		callbackCalled = true
	}

	detector.SetMetricsCallback(callback)

	// Give callback time to fire
	time.Sleep(100 * time.Millisecond)

	if !callbackCalled {
		t.Logf("Callback not called yet (may need detector to be running)")
	}
	if detector.metricsCallback == nil {
		t.Errorf("expected callback to be set")
	}
}

func TestDetermineNetworkType(t *testing.T) {
	cfg := &Config{
		Enabled:           true,
		DetectionInterval: 1000,
		MaxLatencyTarget:  50,
	}

	detector := NewNetworkDetector(cfg)

	tests := []struct {
		name            string
		latency         int
		bandwidth       int
		expectedNetType NetworkType
	}{
		{"5G network", 25, 45, NetworkType(Network5G)},
		{"4G network", 100, 20, NetworkType(Network4G)},
		{"WiFi network", 75, 10, NetworkType(NetworkWiFi)},
		{"High latency", 200, 50, NetworkType(NetworkWiFi)},
		{"Unknown", 0, 0, NetworkType(NetworkUnknown)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			netType := detector.determineNetworkType(tt.latency, tt.bandwidth)
			if netType != tt.expectedNetType {
				t.Errorf("expected %v, got %v", tt.expectedNetType, netType)
			}
		})
	}
}

func TestGetSignalStrength(t *testing.T) {
	cfg := &Config{
		Enabled:           true,
		DetectionInterval: 1000,
		MaxLatencyTarget:  50,
	}

	detector := NewNetworkDetector(cfg)

	tests := []struct {
		name           string
		networkType    NetworkType
		expectNegative bool
	}{
		{"5G signal", NetworkType(Network5G), true},
		{"4G signal", NetworkType(Network4G), true},
		{"WiFi signal", NetworkType(NetworkWiFi), true},
		{"Unknown signal", NetworkType(NetworkUnknown), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strength := detector.getSignalStrength(tt.networkType)
			if tt.expectNegative && strength >= 0 {
				t.Errorf("expected negative signal strength, got %d", strength)
			}
		})
	}
}

func TestCalculateLatencyScore(t *testing.T) {
	cfg := &Config{
		Enabled:           true,
		DetectionInterval: 1000,
		MaxLatencyTarget:  50,
	}

	detector := NewNetworkDetector(cfg)

	tests := []struct {
		name     string
		latency  int
		minScore int
		maxScore int
	}{
		{"0ms latency", 0, 100, 100},
		{"50ms latency", 50, 70, 80},
		{"100ms latency", 100, 40, 60},
		{"200ms+ latency", 200, 0, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := detector.calculateLatencyScore(tt.latency)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("expected score between %d-%d, got %d", tt.minScore, tt.maxScore, score)
			}
		})
	}
}

func TestCalculateBandwidthScore(t *testing.T) {
	cfg := &Config{
		Enabled:           true,
		DetectionInterval: 1000,
		MaxLatencyTarget:  50,
	}

	detector := NewNetworkDetector(cfg)

	tests := []struct {
		name      string
		bandwidth int
		expected  int
	}{
		{"50Mbps+", 50, 100},
		{"30Mbps", 30, 80},
		{"10Mbps", 10, 40},
		{"1Mbps", 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := detector.calculateBandwidthScore(tt.bandwidth)
			if score != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, score)
			}
		})
	}
}

func TestMeasureLatency(t *testing.T) {
	cfg := &Config{
		Enabled:           true,
		DetectionInterval: 1000,
		MaxLatencyTarget:  50,
	}

	detector := NewNetworkDetector(cfg)
	ctx := context.Background()

	latency, err := detector.measureLatency(ctx)
	if err != nil {
		t.Fatalf("measureLatency failed: %v", err)
	}

	if latency < 0 {
		t.Errorf("expected non-negative latency, got %d", latency)
	}
}

func TestMeasureBandwidth(t *testing.T) {
	cfg := &Config{
		Enabled:           true,
		DetectionInterval: 1000,
		MaxLatencyTarget:  50,
	}

	detector := NewNetworkDetector(cfg)
	ctx := context.Background()

	bandwidth, err := detector.measureBandwidth(ctx)
	if err != nil {
		t.Fatalf("measureBandwidth failed: %v", err)
	}

	if bandwidth < 0 {
		t.Errorf("expected non-negative bandwidth, got %d", bandwidth)
	}
}

func TestGetStatistics(t *testing.T) {
	cfg := &Config{
		Enabled:           true,
		DetectionInterval: 1000,
		MaxLatencyTarget:  50,
	}

	detector := NewNetworkDetector(cfg)

	stats := detector.GetStatistics()
	if stats == nil {
		t.Errorf("expected statistics, got nil")
	}

	if _, ok := stats["currentNetwork"]; !ok {
		t.Errorf("expected currentNetwork in stats")
	}

	if _, ok := stats["isRunning"]; !ok {
		t.Errorf("expected isRunning in stats")
	}

	if _, ok := stats["networkQuality"]; !ok {
		t.Errorf("expected networkQuality in stats")
	}
}

func TestDetectorConcurrency(t *testing.T) {
	cfg := &Config{
		Enabled:           true,
		DetectionInterval: 100,
		MaxLatencyTarget:  50,
	}

	detector := NewNetworkDetector(cfg)
	ctx := context.Background()

	err := detector.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Read from multiple goroutines
	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				detector.GetCurrentNetwork()
				detector.Is5GAvailable()
				detector.GetNetworkQuality()
				time.Sleep(10 * time.Millisecond)
			}
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}

	err = detector.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}
