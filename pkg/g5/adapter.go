package g5

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Adapter is the main 5G network adapter coordinating all subcomponents
type Adapter struct {
	mu               sync.RWMutex
	detector         *NetworkDetector
	client           *Client
	edgeManager      *EdgeNodeManager
	qualitySelector  *QualitySelector
	metricsCollector *MetricsCollector
	config           *AdapterConfig
	currentSession   *SessionContext
	started          bool
	stopChan         chan struct{}
	wg               sync.WaitGroup
	statusCallbacks  []func(*AdapterStatus)
	warningCallbacks []func(*AdapterWarning)
}

// AdapterConfig defines configuration for the 5G adapter
type AdapterConfig struct {
	APIBaseURL              string
	DetectionInterval       time.Duration
	HealthCheckInterval     time.Duration
	MetricsReportInterval   time.Duration
	MaxEdgeConnections      int
	EnableMetricsCollection bool
	EnableAutoQualityAdapt  bool
	TargetLatency           int64 // milliseconds
	TargetBandwidth         int64 // kilobits per second
	QualitySwitchThreshold  int   // percentage change before switching
}

// SessionContext tracks the current active session
type SessionContext struct {
	ID             string
	StartedAt      time.Time
	EdgeNodeID     string
	CurrentQuality string
	IsActive       bool
	BytesSent      int64
	BytesReceived  int64
}

// AdapterStatus provides current status of the adapter
type AdapterStatus struct {
	IsStarted           bool
	IsHealthy           bool
	CurrentNetwork      *Network5GStatus
	CurrentQuality      string
	ActiveEdgeNode      *EdgeNode
	ActiveSessionID     string
	TotalActiveSessions int
	GlobalMetrics       *GlobalMetrics
	DetectorStatus      string
	EdgeManagerStatus   string
	LastStatusUpdate    time.Time
}

// AdapterWarning represents a warning from the adapter
type AdapterWarning struct {
	Level      string // "info", "warning", "error"
	Code       string
	Message    string
	Component  string
	Timestamp  time.Time
	IsResolved bool
}

// DefaultAdapterConfig returns default adapter configuration
func DefaultAdapterConfig() *AdapterConfig {
	return &AdapterConfig{
		APIBaseURL:              "https://api.5g.vtp.local",
		DetectionInterval:       2 * time.Second,
		HealthCheckInterval:     30 * time.Second,
		MetricsReportInterval:   10 * time.Second,
		MaxEdgeConnections:      10,
		EnableMetricsCollection: true,
		EnableAutoQualityAdapt:  true,
		TargetLatency:           50,
		TargetBandwidth:         20000,
		QualitySwitchThreshold:  10,
	}
}

// NewAdapter creates a new 5G adapter with provided configuration
func NewAdapter(config *AdapterConfig) (*Adapter, error) {
	if config == nil {
		config = DefaultAdapterConfig()
	}

	// Initialize client
	client := NewClient(config.APIBaseURL, &Config{
		Enabled:           true,
		DetectionInterval: int(config.DetectionInterval.Milliseconds()),
		MaxLatencyTarget:  int(config.TargetLatency),
	})

	// Initialize subcomponents
	detector := NewNetworkDetector(&Config{
		Enabled:           true,
		DetectionInterval: int(config.DetectionInterval.Milliseconds()),
		MaxLatencyTarget:  int(config.TargetLatency),
	})
	edgeManager := NewEdgeNodeManager(client, config.MaxEdgeConnections)
	qualitySelector := NewQualitySelector(&AdaptiveStrategy{
		Enabled:         config.EnableAutoQualityAdapt,
		TargetLatency:   int(config.TargetLatency),
		TargetBandwidth: int(config.TargetBandwidth),
	})
	metricsCollector := NewMetricsCollector(client)

	return &Adapter{
		detector:         detector,
		client:           client,
		edgeManager:      edgeManager,
		qualitySelector:  qualitySelector,
		metricsCollector: metricsCollector,
		config:           config,
		stopChan:         make(chan struct{}),
		statusCallbacks:  make([]func(*AdapterStatus), 0),
		warningCallbacks: make([]func(*AdapterWarning), 0),
	}, nil
}

// Start initializes and starts the 5G adapter
func (a *Adapter) Start(ctx context.Context) error {
	a.mu.Lock()

	if a.started {
		a.mu.Unlock()
		return errors.New("adapter already started")
	}

	// Start detector
	if err := a.detector.Start(ctx); err != nil {
		a.mu.Unlock()
		return err
	}

	// Start edge manager
	if err := a.edgeManager.Start(ctx); err != nil {
		a.detector.Stop()
		a.mu.Unlock()
		return err
	}

	// Start metrics collector if enabled
	if a.config.EnableMetricsCollection {
		if err := a.metricsCollector.Start(ctx); err != nil {
			a.detector.Stop()
			a.edgeManager.Stop()
			a.mu.Unlock()
			return err
		}
	}

	a.started = true
	a.mu.Unlock()

	// Start status monitoring loop
	a.wg.Add(1)
	go a.statusMonitorLoop(ctx)

	// Emit started event
	a.emitStatus()

	return nil
}

// Stop gracefully stops the 5G adapter
func (a *Adapter) Stop() error {
	a.mu.Lock()

	if !a.started {
		a.mu.Unlock()
		return errors.New("adapter not started")
	}

	a.started = false
	a.mu.Unlock()

	close(a.stopChan)
	a.wg.Wait()

	// Stop all subcomponents
	a.detector.Stop()
	a.edgeManager.Stop()
	if a.config.EnableMetricsCollection {
		a.metricsCollector.Stop()
	}

	// End current session if active
	if a.currentSession != nil && a.currentSession.IsActive {
		a.EndSession()
	}

	return nil
}

// StartSession initiates a new 5G streaming session
func (a *Adapter) StartSession(sessionID string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.started {
		return errors.New("adapter not started")
	}

	if a.currentSession != nil && a.currentSession.IsActive {
		return errors.New("session already active")
	}

	// Select edge node
	criteria := SelectionCriteria{
		ExcludeOffline: true,
		MaxLatency:     100,
	}
	selection, err := a.edgeManager.SelectNode(criteria)
	if err != nil {
		a.emitWarning(&AdapterWarning{
			Level:     "warning",
			Code:      "NO_EDGE_NODE",
			Message:   "No suitable edge node found: " + err.Error(),
			Component: "EdgeManager",
			Timestamp: time.Now(),
		})
		return err
	}

	// Start metrics tracking
	if a.config.EnableMetricsCollection {
		a.metricsCollector.StartSession(sessionID, selection.SelectedNode.ID)
	}

	// Create session context
	a.currentSession = &SessionContext{
		ID:         sessionID,
		StartedAt:  time.Now(),
		EdgeNodeID: selection.SelectedNode.ID,
		IsActive:   true,
	}

	a.emitStatus()
	return nil
}

// EndSession terminates the current session
func (a *Adapter) EndSession() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.currentSession != nil {
		if a.config.EnableMetricsCollection {
			a.metricsCollector.EndSession(a.currentSession.ID)
		}
		a.currentSession.IsActive = false
	}
}

// GetStatus returns current adapter status
func (a *Adapter) GetStatus() *AdapterStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()

	status := &AdapterStatus{
		IsStarted:        a.started,
		IsHealthy:        a.isHealthy(),
		LastStatusUpdate: time.Now(),
	}

	if a.started {
		status.CurrentNetwork = a.detector.GetCurrentNetwork()
		profile := a.qualitySelector.GetCurrentProfile()
		if profile != nil {
			status.CurrentQuality = profile.Name
		}
		status.DetectorStatus = "running"
		status.EdgeManagerStatus = "running"

		if a.currentSession != nil {
			status.ActiveSessionID = a.currentSession.ID
			nodeMetrics := a.edgeManager.GetNodeStatus(a.currentSession.EdgeNodeID)
			if nodeMetrics != nil {
				// Find the actual node to get the full EdgeNode
				nodes := a.edgeManager.GetAllNodes()
				for _, node := range nodes {
					if node.ID == a.currentSession.EdgeNodeID {
						status.ActiveEdgeNode = node
						break
					}
				}
			}
		}

		status.TotalActiveSessions = a.metricsCollector.GetActiveSessionCount()

		if a.config.EnableMetricsCollection {
			status.GlobalMetrics = a.metricsCollector.GetGlobalMetrics()
		}
	}

	return status
}

// AdaptQuality adjusts streaming quality based on network conditions
func (a *Adapter) AdaptQuality() (string, error) {
	a.mu.RLock()
	detector := a.detector
	selector := a.qualitySelector
	a.mu.RUnlock()

	if !a.IsStarted() {
		return "", errors.New("adapter not started")
	}

	network := detector.GetCurrentNetwork()
	if network == nil {
		return "", errors.New("network not available")
	}

	quality, err := selector.SelectQuality(network.Latency, network.Bandwidth)
	if err != nil {
		return "", err
	}

	a.mu.Lock()
	if a.currentSession != nil {
		a.currentSession.CurrentQuality = string(quality)
	}
	a.mu.Unlock()

	profile := selector.GetCurrentProfile()
	if profile != nil {
		return profile.Name, nil
	}
	return string(quality), nil
}

// RecordMetric records a performance metric for the current session
func (a *Adapter) RecordMetric(metric string, value interface{}) {
	a.mu.RLock()
	collector := a.metricsCollector
	session := a.currentSession
	a.mu.RUnlock()

	if !a.config.EnableMetricsCollection || session == nil {
		return
	}

	switch metric {
	case "latency":
		if latency, ok := value.(int64); ok {
			collector.RecordLatency(session.ID, latency)
		}
	case "bandwidth":
		if bandwidth, ok := value.(int64); ok {
			collector.RecordBandwidth(session.ID, bandwidth)
		}
	case "packetLoss":
		if loss, ok := value.(float32); ok {
			collector.RecordPacketLoss(session.ID, loss)
		}
	case "jitter":
		if jitter, ok := value.(int64); ok {
			collector.RecordJitter(session.ID, jitter)
		}
	case "frameDropped":
		collector.RecordFrameDropped(session.ID)
	case "packetLost":
		if count, ok := value.(int); ok {
			collector.RecordPacketLost(session.ID, count)
		}
	case "packetSent":
		if count, ok := value.(int); ok {
			collector.RecordPacketSent(session.ID, count)
		}
	}
}

// GetCurrentNetwork returns the current detected network status
func (a *Adapter) GetCurrentNetwork() *Network5GStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.started {
		return nil
	}
	return a.detector.GetCurrentNetwork()
}

// GetNetworkQuality returns the current network quality score (0-100)
func (a *Adapter) GetNetworkQuality() int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.started {
		return 0
	}
	return a.detector.GetNetworkQuality()
}

// Is5GAvailable checks if 5G is currently available
func (a *Adapter) Is5GAvailable() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.started {
		return false
	}
	return a.detector.Is5GAvailable()
}

// GetAvailableEdgeNodes returns list of available edge nodes
func (a *Adapter) GetAvailableEdgeNodes() []*EdgeNode {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.started {
		return []*EdgeNode{}
	}
	return a.edgeManager.GetAllNodes()
}

// GetClosestEdgeNode returns the edge node with lowest latency
func (a *Adapter) GetClosestEdgeNode(ctx context.Context) (*EdgeNode, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.started {
		return nil, errors.New("adapter not started")
	}
	return a.edgeManager.GetClosestNode(ctx)
}

// IsStarted returns whether the adapter is running
func (a *Adapter) IsStarted() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.started
}

// IsHealthy checks if the adapter is in a healthy state
func (a *Adapter) isHealthy() bool {
	// Check detector health
	if a.detector == nil {
		return false
	}

	// Check edge manager health
	nodes := a.edgeManager.GetAllNodes()
	if len(nodes) == 0 {
		return false
	}

	// Check if any nodes are online
	healthySummary := a.edgeManager.GetHealthySummary()
	return len(healthySummary) > 0
}

// statusMonitorLoop continuously monitors adapter status and emits updates
func (a *Adapter) statusMonitorLoop(ctx context.Context) {
	defer a.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-a.stopChan:
			return
		case <-ticker.C:
			a.checkAndEmitStatus()
		}
	}
}

// checkAndEmitStatus checks conditions and emits status/warnings
func (a *Adapter) checkAndEmitStatus() {
	a.mu.RLock()
	started := a.started
	currentSession := a.currentSession
	detector := a.detector
	a.mu.RUnlock()

	if !started {
		return
	}

	// Check network quality
	network := detector.GetCurrentNetwork()
	if network != nil {
		if int64(network.Latency) > a.config.TargetLatency {
			a.emitWarning(&AdapterWarning{
				Level:     "warning",
				Code:      "HIGH_LATENCY",
				Message:   "Network latency exceeds target",
				Component: "Detector",
				Timestamp: time.Now(),
			})
		}

		if int64(network.Bandwidth) < a.config.TargetBandwidth {
			a.emitWarning(&AdapterWarning{
				Level:     "warning",
				Code:      "LOW_BANDWIDTH",
				Message:   "Network bandwidth below target",
				Component: "Detector",
				Timestamp: time.Now(),
			})
		}
	}

	// Check session health
	if currentSession != nil && currentSession.IsActive {
		metrics := a.metricsCollector.GetSessionMetrics(currentSession.ID)
		if metrics != nil {
			if metrics.AvgPacketLoss > 0.05 {
				a.emitWarning(&AdapterWarning{
					Level:     "warning",
					Code:      "HIGH_PACKET_LOSS",
					Message:   "Packet loss exceeds threshold",
					Component: "MetricsCollector",
					Timestamp: time.Now(),
				})
			}
		}
	}

	a.emitStatus()
}

// emitStatus broadcasts current adapter status to registered callbacks
func (a *Adapter) emitStatus() {
	a.mu.RLock()
	callbacks := make([]func(*AdapterStatus), len(a.statusCallbacks))
	copy(callbacks, a.statusCallbacks)
	a.mu.RUnlock()

	status := a.GetStatus()
	for _, callback := range callbacks {
		go callback(status)
	}
}

// emitWarning broadcasts a warning to registered callbacks
func (a *Adapter) emitWarning(warning *AdapterWarning) {
	a.mu.RLock()
	callbacks := make([]func(*AdapterWarning), len(a.warningCallbacks))
	copy(callbacks, a.warningCallbacks)
	a.mu.RUnlock()

	for _, callback := range callbacks {
		go callback(warning)
	}
}

// RegisterStatusCallback registers a callback for status updates
func (a *Adapter) RegisterStatusCallback(callback func(*AdapterStatus)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.statusCallbacks = append(a.statusCallbacks, callback)
}

// RegisterWarningCallback registers a callback for warnings
func (a *Adapter) RegisterWarningCallback(callback func(*AdapterWarning)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.warningCallbacks = append(a.warningCallbacks, callback)
}

// RegisterMetricsCallback registers a callback for session metrics
func (a *Adapter) RegisterMetricsCallback(callback func(*SessionMetrics)) {
	a.mu.RLock()
	collector := a.metricsCollector
	a.mu.RUnlock()

	if collector != nil {
		collector.RegisterSessionCallback(callback)
	}
}

// GetSessionMetrics returns metrics for the current session
func (a *Adapter) GetSessionMetrics() *SessionMetrics {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.currentSession == nil {
		return nil
	}
	return a.metricsCollector.GetSessionMetrics(a.currentSession.ID)
}

// GetGlobalMetrics returns aggregate metrics across all sessions
func (a *Adapter) GetGlobalMetrics() *GlobalMetrics {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.config.EnableMetricsCollection {
		return nil
	}
	return a.metricsCollector.GetGlobalMetrics()
}

// DetectNetworkType checks current 5G availability and network type
func (a *Adapter) DetectNetworkType(ctx context.Context) (*DetectionResult, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.started {
		return nil, errors.New("adapter not started")
	}

	return a.detector.DetectNetwork(ctx)
}

// GetConfig returns the current adapter configuration
func (a *Adapter) GetConfig() *AdapterConfig {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.config
}

// UpdateConfig updates adapter configuration
func (a *Adapter) UpdateConfig(newConfig *AdapterConfig) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if newConfig != nil {
		a.config = newConfig
	}
}
