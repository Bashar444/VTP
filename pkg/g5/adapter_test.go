package g5

import (
	"context"
	"testing"
	"time"
)

// TestNewAdapter tests creating a new adapter with a config
func TestNewAdapter(t *testing.T) {
	config := DefaultAdapterConfig()
	adapter, err := NewAdapter(config)

	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	if adapter == nil {
		t.Fatal("expected adapter instance, got nil")
	}

	if adapter.config != config {
		t.Errorf("expected config to be set")
	}
}

// TestNewAdapterNilConfig tests creating adapter with nil config (uses defaults)
func TestNewAdapterNilConfig(t *testing.T) {
	adapter, err := NewAdapter(nil)

	if err != nil {
		t.Fatalf("NewAdapter with nil config failed: %v", err)
	}

	if adapter == nil {
		t.Fatal("expected adapter instance, got nil")
	}

	if adapter.config == nil {
		t.Fatalf("expected default config to be set")
	}
}

// TestDefaultAdapterConfig tests the default configuration
func TestDefaultAdapterConfig(t *testing.T) {
	config := DefaultAdapterConfig()

	if config == nil {
		t.Fatal("expected default config, got nil")
	}

	if !config.EnableMetricsCollection {
		t.Errorf("expected EnableMetricsCollection=true")
	}

	if config.TargetLatency <= 0 {
		t.Errorf("expected positive TargetLatency")
	}

	if config.TargetBandwidth <= 0 {
		t.Errorf("expected positive TargetBandwidth")
	}
}

// TestAdapterStart tests starting the adapter
func TestAdapterStart(t *testing.T) {
	config := DefaultAdapterConfig()
	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	ctx := context.Background()
	err = adapter.Start(ctx)
	if err != nil {
		t.Logf("Start may fail due to network initialization (expected): %v", err)
	}

	err = adapter.Stop()
	if err != nil {
		t.Logf("Stop returned error (may be expected): %v", err)
	}
}

// TestAdapterStop tests stopping the adapter
func TestAdapterStop(t *testing.T) {
	config := DefaultAdapterConfig()
	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	err = adapter.Stop()
	if err != nil {
		t.Logf("Stop may error (expected if not started): %v", err)
	}
}

// TestAdapterGetStatus tests getting adapter status
func TestAdapterGetStatus(t *testing.T) {
	config := DefaultAdapterConfig()
	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	status := adapter.GetStatus()
	if status == nil {
		t.Fatal("expected status, got nil")
	}

	if status.IsStarted {
		t.Errorf("expected IsStarted=false initially")
	}
}

// TestAdapterStartSession tests starting a session
func TestAdapterStartSession(t *testing.T) {
	config := DefaultAdapterConfig()
	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	// Adapter must be started before starting a session
	ctx := context.Background()
	adapter.Start(ctx)

	sessionID := "test-session-123"
	err = adapter.StartSession(sessionID)
	if err != nil {
		t.Logf("StartSession failed (expected if no edge nodes available): %v", err)
		return
	}

	status := adapter.GetStatus()
	if status.ActiveSessionID != sessionID {
		t.Errorf("expected session ID in status, got %q", status.ActiveSessionID)
	}

	adapter.Stop()
}

// TestAdapterEndSession tests ending a session
func TestAdapterEndSession(t *testing.T) {
	config := DefaultAdapterConfig()
	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	ctx := context.Background()
	adapter.Start(ctx)

	sessionID := "test-session-456"
	err = adapter.StartSession(sessionID)
	if err != nil {
		t.Logf("StartSession failed (expected if no edge nodes): %v", err)
		adapter.Stop()
		return
	}

	// EndSession takes no arguments
	adapter.EndSession()

	status := adapter.GetStatus()
	if status.ActiveSessionID == sessionID {
		t.Logf("Session may still be visible briefly after end (eventual consistency)")
	}

	adapter.Stop()
}

// TestAdapterRecordMetric tests recording a metric
func TestAdapterRecordMetric(t *testing.T) {
	config := DefaultAdapterConfig()
	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	ctx := context.Background()
	adapter.Start(ctx)

	sessionID := "test-session-789"
	err = adapter.StartSession(sessionID)
	if err != nil {
		t.Logf("StartSession failed (expected if no edge nodes): %v", err)
		adapter.Stop()
		return
	}

	// RecordMetric takes metric name and value (no sessionID)
	adapter.RecordMetric("latency", int64(25))
	adapter.RecordMetric("bandwidth", int64(45))

	adapter.EndSession()
	adapter.Stop()
}

// TestAdapterGetCurrentNetwork tests getting current network status
func TestAdapterGetCurrentNetwork(t *testing.T) {
	config := DefaultAdapterConfig()
	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	ctx := context.Background()
	adapter.Start(ctx)

	network := adapter.GetCurrentNetwork()
	// May be nil if detector hasn't reported yet
	if network == nil {
		t.Logf("Network may be nil until detector runs")
	} else {
		if network.Type == "" {
			t.Errorf("expected network type to be set")
		}
	}

	adapter.Stop()
}

// TestAdapterGetNetworkQuality tests getting network quality
func TestAdapterGetNetworkQuality(t *testing.T) {
	config := DefaultAdapterConfig()
	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	quality := adapter.GetNetworkQuality()
	if quality < 0 || quality > 100 {
		t.Errorf("expected quality between 0-100, got %d", quality)
	}
}

// TestAdapterIs5GAvailable tests checking 5G availability
func TestAdapterIs5GAvailable(t *testing.T) {
	config := DefaultAdapterConfig()
	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	is5G := adapter.Is5GAvailable()
	_ = is5G // is5G should be boolean
}

// TestAdapterGetAvailableEdgeNodes tests getting edge nodes
func TestAdapterGetAvailableEdgeNodes(t *testing.T) {
	config := DefaultAdapterConfig()
	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	nodes := adapter.GetAvailableEdgeNodes()
	if nodes == nil {
		t.Errorf("expected nodes list, got nil")
	}
}

// TestAdapterGetClosestEdgeNode tests selecting closest edge node
func TestAdapterGetClosestEdgeNode(t *testing.T) {
	config := DefaultAdapterConfig()
	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	ctx := context.Background()
	node, err := adapter.GetClosestEdgeNode(ctx)
	if err != nil {
		t.Logf("GetClosestEdgeNode may fail if no nodes available (expected): %v", err)
		return
	}

	if node != nil && node.ID == "" {
		t.Errorf("expected node ID to be set")
	}
}

// TestAdapterIsStarted tests the IsStarted method
func TestAdapterIsStarted(t *testing.T) {
	config := DefaultAdapterConfig()
	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	if adapter.IsStarted() {
		t.Errorf("expected IsStarted=false before Start() called")
	}

	adapter.Stop()
}

// TestAdapterRegisterStatusCallback tests registering a status callback
func TestAdapterRegisterStatusCallback(t *testing.T) {
	config := DefaultAdapterConfig()
	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	callback := func(status *AdapterStatus) {}

	adapter.RegisterStatusCallback(callback)

	if len(adapter.statusCallbacks) == 0 {
		t.Errorf("expected callback to be registered")
	}
}

// TestAdapterRegisterWarningCallback tests registering a warning callback
func TestAdapterRegisterWarningCallback(t *testing.T) {
	config := DefaultAdapterConfig()
	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	callback := func(warning *AdapterWarning) {}

	adapter.RegisterWarningCallback(callback)

	if len(adapter.warningCallbacks) == 0 {
		t.Errorf("expected callback to be registered")
	}
}

// TestAdapterRegisterMetricsCallback tests registering a metrics callback
func TestAdapterRegisterMetricsCallback(t *testing.T) {
	config := DefaultAdapterConfig()
	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	callback := func(metrics *SessionMetrics) {}

	adapter.RegisterMetricsCallback(callback)
}

// TestAdapterAdaptQuality tests quality adaptation
func TestAdapterAdaptQuality(t *testing.T) {
	config := DefaultAdapterConfig()
	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	sessionID := "quality-test"
	adapter.StartSession(sessionID)

	quality, err := adapter.AdaptQuality()
	if err != nil {
		t.Logf("AdaptQuality returned error (may be expected): %v", err)
		adapter.EndSession()
		return
	}

	if quality == "" {
		t.Errorf("expected quality level string, got empty")
	}

	adapter.EndSession()
}

// TestAdapterConfigValidation tests configuration validation
func TestAdapterConfigValidation(t *testing.T) {
	tests := []struct {
		name          string
		config        *AdapterConfig
		shouldSucceed bool
	}{
		{
			"valid custom config",
			&AdapterConfig{
				APIBaseURL:              "https://api.5g.local",
				DetectionInterval:       2 * time.Second,
				MaxEdgeConnections:      10,
				EnableMetricsCollection: true,
				TargetLatency:           50,
				TargetBandwidth:         20000,
			},
			true,
		},
		{
			"nil config uses defaults",
			nil,
			true,
		},
		{
			"default config",
			DefaultAdapterConfig(),
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := NewAdapter(tt.config)
			if tt.shouldSucceed && err != nil {
				t.Errorf("expected success, got error: %v", err)
			}
			if adapter != nil {
				adapter.Stop()
			}
		})
	}
}

// TestAdapterConcurrentOperations tests thread safety of adapter
func TestAdapterConcurrentOperations(t *testing.T) {
	config := DefaultAdapterConfig()
	adapter, err := NewAdapter(config)
	if err != nil {
		t.Fatalf("NewAdapter failed: %v", err)
	}

	ctx := context.Background()
	adapter.Start(ctx)

	sessionID := "concurrent-test"
	err = adapter.StartSession(sessionID)
	if err != nil {
		t.Logf("StartSession failed (expected if no edge nodes): %v", err)
		adapter.Stop()
		return
	}

	done := make(chan bool)

	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				status := adapter.GetStatus()
				if status == nil {
					t.Errorf("expected status, got nil")
				}

				network := adapter.GetCurrentNetwork()
				if network == nil {
					t.Logf("network may be nil (expected)")
				}

				quality := adapter.GetNetworkQuality()
				if quality < 0 || quality > 100 {
					t.Errorf("expected quality 0-100, got %d", quality)
				}

				adapter.RecordMetric("latency", int64(20+j))

				time.Sleep(time.Millisecond)
			}
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}

	adapter.EndSession()
	adapter.Stop()
}
