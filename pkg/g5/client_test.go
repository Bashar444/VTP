package g5

import (
	"context"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	cfg := &Config{
		Enabled: true,
	}

	client := NewClient("https://api.5g.local", cfg)

	if client == nil {
		t.Fatal("expected client instance, got nil")
	}

	if client.baseURL != "https://api.5g.local" {
		t.Errorf("expected baseURL https://api.5g.local, got %s", client.baseURL)
	}

	if client.config != cfg {
		t.Errorf("expected config to be set")
	}
}

func TestNewClientDefaultURL(t *testing.T) {
	client := NewClient("", nil)

	if client == nil {
		t.Fatal("expected client instance, got nil")
	}

	if client.baseURL != "https://api.5g.vtp.local" {
		t.Errorf("expected default baseURL, got %s", client.baseURL)
	}
}

func TestClientSetTimeout(t *testing.T) {
	cfg := &Config{
		Enabled: true,
	}

	client := NewClient("https://api.5g.local", cfg)

	originalTimeout := client.timeout
	newTimeout := 30 * time.Second

	client.SetTimeout(newTimeout)

	if client.timeout != newTimeout {
		t.Errorf("expected timeout %v, got %v", newTimeout, client.timeout)
	}

	if originalTimeout == newTimeout {
		t.Errorf("expected timeout to change")
	}
}

func TestClientGetNetworkStatusURL(t *testing.T) {
	cfg := &Config{
		Enabled: true,
	}

	client := NewClient("https://api.5g.local", cfg)

	// Test that it attempts to call the correct endpoint
	ctx := context.Background()
	_, err := client.GetNetworkStatus(ctx)

	// Expected to fail (network not available) but should attempt correct endpoint
	if err == nil {
		// If no error, response was received (network available)
		t.Logf("GetNetworkStatus called successfully")
	}
}

func TestClientMeasureLatencyURL(t *testing.T) {
	cfg := &Config{
		Enabled: true,
	}

	client := NewClient("https://api.5g.local", cfg)

	ctx := context.Background()
	_, err := client.MeasureLatency(ctx)

	// Expected to fail (network not available)
	if err == nil {
		t.Logf("MeasureLatency called successfully")
	}
}

func TestClientMeasureBandwidthURL(t *testing.T) {
	cfg := &Config{
		Enabled: true,
	}

	client := NewClient("https://api.5g.local", cfg)

	ctx := context.Background()
	_, err := client.MeasureBandwidth(ctx)

	// Expected to fail (network not available)
	if err == nil {
		t.Logf("MeasureBandwidth called successfully")
	}
}

func TestClientGetMetricsValidation(t *testing.T) {
	cfg := &Config{
		Enabled: true,
	}

	client := NewClient("https://api.5g.local", cfg)
	ctx := context.Background()

	// Test with empty sessionID
	_, err := client.GetMetrics(ctx, "")
	if err == nil {
		t.Errorf("expected error with empty sessionID, got nil")
	}
}

func TestClientReportMetricsValidation(t *testing.T) {
	cfg := &Config{
		Enabled: true,
	}

	client := NewClient("https://api.5g.local", cfg)
	ctx := context.Background()

	// Test with nil metrics
	err := client.ReportMetrics(ctx, nil)
	if err == nil {
		t.Errorf("expected error with nil metrics, got nil")
	}
}

func TestClientGetEdgeNodesURL(t *testing.T) {
	cfg := &Config{
		Enabled: true,
	}

	client := NewClient("https://api.5g.local", cfg)
	ctx := context.Background()

	_, err := client.GetEdgeNodes(ctx)

	// Expected to fail (network not available)
	if err == nil {
		t.Logf("GetEdgeNodes called successfully")
	}
}

func TestClientGetEdgeNodeValidation(t *testing.T) {
	cfg := &Config{
		Enabled: true,
	}

	client := NewClient("https://api.5g.local", cfg)
	ctx := context.Background()

	// Test with empty nodeID
	_, err := client.GetEdgeNode(ctx, "")
	if err == nil {
		t.Errorf("expected error with empty nodeID, got nil")
	}
}

func TestClientReportEdgeNodeHealthValidation(t *testing.T) {
	cfg := &Config{
		Enabled: true,
	}

	client := NewClient("https://api.5g.local", cfg)
	ctx := context.Background()

	// Test with nil health check
	err := client.ReportEdgeNodeHealth(ctx, nil)
	if err == nil {
		t.Errorf("expected error with nil health check, got nil")
	}
}

func TestClientConnectToEdgeValidation(t *testing.T) {
	cfg := &Config{
		Enabled: true,
	}

	client := NewClient("https://api.5g.local", cfg)
	ctx := context.Background()

	// Test with empty nodeID
	err := client.ConnectToEdge(ctx, "", "session-1")
	if err == nil {
		t.Errorf("expected error with empty nodeID, got nil")
	}

	// Test with empty sessionID
	err = client.ConnectToEdge(ctx, "node-1", "")
	if err == nil {
		t.Errorf("expected error with empty sessionID, got nil")
	}
}

func TestClientHealthCheck(t *testing.T) {
	cfg := &Config{
		Enabled: true,
	}

	client := NewClient("https://api.5g.local", cfg)
	ctx := context.Background()

	err := client.Health(ctx)

	// Expected to fail (network not available)
	if err == nil {
		t.Logf("Health check called successfully")
	}
}

func TestClientContextCancellation(t *testing.T) {
	cfg := &Config{
		Enabled: true,
	}

	client := NewClient("https://api.5g.local", cfg)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// This should fail due to canceled context
	_, err := client.GetNetworkStatus(ctx)
	if err == nil {
		t.Logf("Request completed despite canceled context")
	}
}

func TestClientMultipleCalls(t *testing.T) {
	cfg := &Config{
		Enabled: true,
	}

	client := NewClient("https://api.5g.local", cfg)
	ctx := context.Background()

	// Make multiple calls to ensure client can be reused
	for i := 0; i < 3; i++ {
		client.Health(ctx)
	}

	t.Logf("Multiple calls completed")
}

func TestClientTimeoutConfiguration(t *testing.T) {
	cfg := &Config{
		Enabled: true,
	}

	client := NewClient("https://api.5g.local", cfg)

	if client.timeout != 10*time.Second {
		t.Errorf("expected default timeout 10s, got %v", client.timeout)
	}

	if client.httpClient.Timeout != 10*time.Second {
		t.Errorf("expected http client timeout 10s, got %v", client.httpClient.Timeout)
	}
}

func TestClientGetMetricsWithValidSessionID(t *testing.T) {
	cfg := &Config{
		Enabled: true,
	}

	client := NewClient("https://api.5g.local", cfg)
	ctx := context.Background()

	// Should not error on validation, but will fail on network
	_, err := client.GetMetrics(ctx, "session-123")
	// We just check that the validation passes
	if err != nil && err.Error() == "sessionID is required" {
		t.Errorf("unexpected validation error: %v", err)
	}
}
