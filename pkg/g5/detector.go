package g5

import (
	"context"
	"errors"
	"sync"
	"time"
)

// NetworkDetector handles 5G network detection and monitoring
type NetworkDetector struct {
	mu              sync.RWMutex
	currentNetwork  *Network5GStatus
	lastDetection   time.Time
	detectionTicker *time.Ticker
	stopChan        chan struct{}
	metricsCallback func(*Network5GStatus)
	config          *Config
	isRunning       bool
}

// NewNetworkDetector creates a new network detector instance
func NewNetworkDetector(cfg *Config) *NetworkDetector {
	return &NetworkDetector{
		config:   cfg,
		stopChan: make(chan struct{}),
		currentNetwork: &Network5GStatus{
			Type:      NetworkUnknown,
			Connected: false,
			Timestamp: time.Now(),
		},
	}
}

// Start begins network detection in background
func (nd *NetworkDetector) Start(ctx context.Context) error {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	if nd.isRunning {
		return errors.New("detector already running")
	}

	if nd.config == nil {
		return errors.New("configuration not set")
	}

	// Perform initial detection
	result, err := nd.DetectNetwork(ctx)
	if err != nil {
		return err
	}

	nd.currentNetwork = &Network5GStatus{
		Type:           result.NetworkType,
		Latency:        result.Latency,
		Bandwidth:      result.Bandwidth,
		SignalStrength: result.SignalStrength,
		Connected:      result.Detected,
		Timestamp:      result.Timestamp,
	}

	nd.isRunning = true

	// Start periodic detection
	interval := time.Duration(nd.config.DetectionInterval) * time.Millisecond
	nd.detectionTicker = time.NewTicker(interval)

	go nd.detectionLoop()

	return nil
}

// Stop stops network detection
func (nd *NetworkDetector) Stop() error {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	if !nd.isRunning {
		return errors.New("detector not running")
	}

	nd.isRunning = false
	if nd.detectionTicker != nil {
		nd.detectionTicker.Stop()
	}
	close(nd.stopChan)

	return nil
}

// DetectNetwork performs a single network detection
func (nd *NetworkDetector) DetectNetwork(ctx context.Context) (*DetectionResult, error) {
	result := &DetectionResult{
		Timestamp: time.Now(),
	}

	// Simulate network detection (would use actual 5G APIs in production)
	// This is where you would call actual 5G network detection APIs

	// For now, implement basic detection logic
	latency, err := nd.measureLatency(ctx)
	if err != nil {
		result.Error = err.Error()
		result.Detected = false
		return result, nil
	}

	bandwidth, err := nd.measureBandwidth(ctx)
	if err != nil {
		result.Error = err.Error()
		result.Detected = false
		return result, nil
	}

	// Determine network type based on latency
	networkType := nd.determineNetworkType(latency, bandwidth)

	result.Detected = networkType != NetworkUnknown
	result.NetworkType = networkType
	result.Latency = latency
	result.Bandwidth = bandwidth
	result.SignalStrength = nd.getSignalStrength(networkType)

	nd.lastDetection = time.Now()

	return result, nil
}

// GetCurrentNetwork returns the current network status
func (nd *NetworkDetector) GetCurrentNetwork() *Network5GStatus {
	nd.mu.RLock()
	defer nd.mu.RUnlock()

	if nd.currentNetwork == nil {
		return &Network5GStatus{
			Type:      NetworkUnknown,
			Connected: false,
			Timestamp: time.Now(),
		}
	}

	// Return a copy to prevent external modifications
	network := *nd.currentNetwork
	return &network
}

// Is5GAvailable checks if 5G network is available and connected
func (nd *NetworkDetector) Is5GAvailable() bool {
	nd.mu.RLock()
	defer nd.mu.RUnlock()

	return nd.currentNetwork != nil &&
		nd.currentNetwork.Type == Network5G &&
		nd.currentNetwork.Connected &&
		nd.currentNetwork.Latency < nd.config.MaxLatencyTarget
}

// GetNetworkQuality returns a quality score (0-100) for current network
func (nd *NetworkDetector) GetNetworkQuality() int {
	nd.mu.RLock()
	defer nd.mu.RUnlock()

	if nd.currentNetwork == nil || !nd.currentNetwork.Connected {
		return 0
	}

	// Calculate quality based on latency and bandwidth
	latencyScore := nd.calculateLatencyScore(nd.currentNetwork.Latency)
	bandwidthScore := nd.calculateBandwidthScore(nd.currentNetwork.Bandwidth)

	// Weighted average
	quality := (latencyScore*40 + bandwidthScore*60) / 100
	if quality < 0 {
		return 0
	}
	if quality > 100 {
		return 100
	}
	return quality
}

// SetMetricsCallback sets a callback function for metrics updates
func (nd *NetworkDetector) SetMetricsCallback(callback func(*Network5GStatus)) {
	nd.mu.Lock()
	defer nd.mu.Unlock()
	nd.metricsCallback = callback
}

// detectionLoop runs periodic network detection
func (nd *NetworkDetector) detectionLoop() {
	ctx := context.Background()

	for {
		select {
		case <-nd.stopChan:
			return
		case <-nd.detectionTicker.C:
			result, err := nd.DetectNetwork(ctx)
			if err == nil && result.Detected {
				nd.mu.Lock()
				nd.currentNetwork = &Network5GStatus{
					Type:           result.NetworkType,
					Latency:        result.Latency,
					Bandwidth:      result.Bandwidth,
					SignalStrength: result.SignalStrength,
					Connected:      result.Detected,
					Timestamp:      result.Timestamp,
				}

				if nd.metricsCallback != nil {
					nd.metricsCallback(nd.currentNetwork)
				}
				nd.mu.Unlock()
			}
		}
	}
}

// measureLatency measures network latency
func (nd *NetworkDetector) measureLatency(ctx context.Context) (int, error) {
	// Simulate latency measurement
	// In production, this would ping 5G edge nodes or use proper measurement APIs

	// For demonstration: 5G targets 20-50ms, 4G targets 50-150ms
	// These values would come from actual measurements
	return 25, nil
}

// measureBandwidth measures available bandwidth
func (nd *NetworkDetector) measureBandwidth(ctx context.Context) (int, error) {
	// Simulate bandwidth measurement
	// In production, this would test actual data transfer speeds

	// For demonstration: return a typical 5G bandwidth
	return 45, nil
}

// determineNetworkType determines the network type based on characteristics
func (nd *NetworkDetector) determineNetworkType(latency, bandwidth int) NetworkType {
	// 5G characteristics: latency < 50ms, bandwidth > 20Mbps
	// 4G characteristics: latency 50-150ms, bandwidth > 5Mbps
	// WiFi characteristics: highly variable

	if latency < 50 && bandwidth > 20 {
		return Network5G
	} else if latency < 150 && bandwidth > 5 {
		return Network4G
	} else if latency > 0 {
		return NetworkWiFi
	}

	return NetworkUnknown
}

// getSignalStrength returns signal strength for network type
func (nd *NetworkDetector) getSignalStrength(networkType NetworkType) int {
	// RSRP (Reference Signal Received Power) values in dBm
	// Typical ranges: -44dBm (excellent) to -140dBm (no signal)

	switch networkType {
	case Network5G:
		return -90 // Good 5G signal
	case Network4G:
		return -100 // Good 4G signal
	case NetworkWiFi:
		return -50 // Good WiFi signal
	default:
		return -140 // No signal
	}
}

// calculateLatencyScore calculates quality score based on latency
func (nd *NetworkDetector) calculateLatencyScore(latency int) int {
	// 0ms = 100, 100ms = 50, 200ms+ = 0
	if latency <= 0 {
		return 100
	}
	score := 100 - (latency / 2)
	if score < 0 {
		return 0
	}
	return score
}

// calculateBandwidthScore calculates quality score based on bandwidth
func (nd *NetworkDetector) calculateBandwidthScore(bandwidth int) int {
	// 50Mbps+ = 100, 20Mbps = 80, 5Mbps = 40, <1Mbps = 0
	if bandwidth >= 50 {
		return 100
	} else if bandwidth >= 20 {
		return 80
	} else if bandwidth >= 5 {
		return 40
	}
	return 0
}

// GetStatistics returns network detection statistics
func (nd *NetworkDetector) GetStatistics() map[string]interface{} {
	nd.mu.RLock()
	defer nd.mu.RUnlock()

	return map[string]interface{}{
		"currentNetwork": nd.currentNetwork,
		"lastDetection":  nd.lastDetection,
		"isRunning":      nd.isRunning,
		"networkQuality": nd.GetNetworkQuality(),
	}
}
