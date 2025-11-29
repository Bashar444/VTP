package g5

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

// EdgeNodeManager manages edge node discovery, health checking, and selection
type EdgeNodeManager struct {
	mu              sync.RWMutex
	nodes           map[string]*EdgeNode
	client          *Client
	healthCheckTick time.Duration
	stopChan        chan struct{}
	wg              sync.WaitGroup
	lastHealthCheck map[string]time.Time
	nodeMetrics     map[string]*EdgeNodeMetrics
	maxConnections  int
	metricsCallback func(*EdgeNodeMetrics)
}

// EdgeNodeMetrics tracks metrics for an edge node
type EdgeNodeMetrics struct {
	NodeID           string
	Region           string
	HealthStatus     string
	LastHealthCheck  time.Time
	ConsecutiveFails int
	AverageLatency   int64
	MaxCapacity      int
	CurrentLoad      int
	ConnectionCount  int
	LastError        string
	UpdatedAt        time.Time
}

// SelectionCriteria defines criteria for node selection
type SelectionCriteria struct {
	PreferredRegion string
	MaxLatency      int64
	MinCapacity     int
	ExcludeOffline  bool
	BalanceLoad     bool
}

// NodeSelection result contains selected node and alternates
type NodeSelection struct {
	SelectedNode *EdgeNode
	Alternates   []*EdgeNode
	SelectedAt   time.Time
	Reason       string
}

// NewEdgeNodeManager creates a new edge node manager
func NewEdgeNodeManager(client *Client, maxConnections int) *EdgeNodeManager {
	return &EdgeNodeManager{
		nodes:           make(map[string]*EdgeNode),
		client:          client,
		healthCheckTick: 30 * time.Second,
		stopChan:        make(chan struct{}),
		lastHealthCheck: make(map[string]time.Time),
		nodeMetrics:     make(map[string]*EdgeNodeMetrics),
		maxConnections:  maxConnections,
	}
}

// Start begins edge node manager operations
func (enm *EdgeNodeManager) Start(ctx context.Context) error {
	// Initial discovery
	err := enm.discoverNodes(ctx)
	if err != nil {
		return err
	}

	// Start health check loop
	enm.wg.Add(1)
	go enm.healthCheckLoop(ctx)

	return nil
}

// Stop gracefully stops the edge node manager
func (enm *EdgeNodeManager) Stop() error {
	close(enm.stopChan)
	enm.wg.Wait()
	return nil
}

// discoverNodes fetches available edge nodes from API
func (enm *EdgeNodeManager) discoverNodes(ctx context.Context) error {
	nodes, err := enm.client.GetEdgeNodes(ctx)
	if err != nil {
		return err
	}

	enm.mu.Lock()
	defer enm.mu.Unlock()

	enm.nodes = make(map[string]*EdgeNode)
	for i := range nodes {
		node := &nodes[i]
		enm.nodes[node.ID] = node
		if _, exists := enm.nodeMetrics[node.ID]; !exists {
			enm.nodeMetrics[node.ID] = &EdgeNodeMetrics{
				NodeID:      node.ID,
				Region:      node.Region,
				MaxCapacity: node.Capacity,
			}
		}
	}

	return nil
}

// healthCheckLoop runs periodic health checks on edge nodes
func (enm *EdgeNodeManager) healthCheckLoop(ctx context.Context) {
	defer enm.wg.Done()

	ticker := time.NewTicker(enm.healthCheckTick)
	defer ticker.Stop()

	// Discover nodes periodically too (every 5 minutes)
	discoverTicker := time.NewTicker(5 * time.Minute)
	defer discoverTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-enm.stopChan:
			return
		case <-ticker.C:
			enm.performHealthChecks(ctx)
		case <-discoverTicker.C:
			enm.discoverNodes(ctx)
		}
	}
}

// performHealthChecks checks health of all known nodes
func (enm *EdgeNodeManager) performHealthChecks(ctx context.Context) {
	enm.mu.RLock()
	nodeIDs := make([]string, 0, len(enm.nodes))
	for id := range enm.nodes {
		nodeIDs = append(nodeIDs, id)
	}
	enm.mu.RUnlock()

	for _, nodeID := range nodeIDs {
		go enm.checkNodeHealth(ctx, nodeID)
	}
}

// checkNodeHealth checks health of a single node
func (enm *EdgeNodeManager) checkNodeHealth(ctx context.Context, nodeID string) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	enm.mu.RLock()
	node, exists := enm.nodes[nodeID]
	enm.mu.RUnlock()

	if !exists {
		return
	}

	// Measure latency to node
	startTime := time.Now()
	_, err := enm.client.GetEdgeNode(ctx, nodeID)
	latency := time.Since(startTime).Milliseconds()

	enm.mu.Lock()
	defer enm.mu.Unlock()

	metrics := enm.nodeMetrics[nodeID]
	if metrics == nil {
		metrics = &EdgeNodeMetrics{
			NodeID:      nodeID,
			Region:      node.Region,
			MaxCapacity: node.Capacity,
		}
		enm.nodeMetrics[nodeID] = metrics
	}

	metrics.LastHealthCheck = time.Now()
	metrics.AverageLatency = latency

	if err != nil {
		metrics.HealthStatus = "unhealthy"
		metrics.LastError = err.Error()
		metrics.ConsecutiveFails++

		// Mark node offline if too many failures
		if metrics.ConsecutiveFails >= 3 {
			node.Status = NodeOffline
		}
	} else {
		metrics.HealthStatus = "healthy"
		metrics.ConsecutiveFails = 0
		if node.Status == NodeOffline || node.Status == NodeDegraded {
			node.Status = NodeOnline
		}
		metrics.LastError = ""
	}

	metrics.UpdatedAt = time.Now()

	// Call metrics callback if registered
	if enm.metricsCallback != nil {
		enm.metricsCallback(metrics)
	}
}

// SelectNode selects the best edge node based on criteria
func (enm *EdgeNodeManager) SelectNode(criteria SelectionCriteria) (*NodeSelection, error) {
	enm.mu.RLock()
	defer enm.mu.RUnlock()

	if len(enm.nodes) == 0 {
		return nil, errors.New("no edge nodes available")
	}

	candidates := enm.filterNodes(criteria)
	if len(candidates) == 0 {
		return nil, errors.New("no suitable edge nodes match criteria")
	}

	// Sort candidates by selection priority
	selected := enm.selectBestNode(candidates, criteria)
	if selected == nil {
		return nil, errors.New("failed to select node")
	}

	// Get alternates (up to 2 backups)
	alternates := make([]*EdgeNode, 0)
	for _, node := range candidates {
		if node.ID != selected.ID && len(alternates) < 2 {
			alternates = append(alternates, node)
		}
	}

	return &NodeSelection{
		SelectedNode: selected,
		Alternates:   alternates,
		SelectedAt:   time.Now(),
		Reason:       "Selected based on latency and load",
	}, nil
}

// filterNodes returns nodes matching the criteria
func (enm *EdgeNodeManager) filterNodes(criteria SelectionCriteria) []*EdgeNode {
	var candidates []*EdgeNode

	for _, node := range enm.nodes {
		// Check status
		if criteria.ExcludeOffline && node.Status != NodeOnline {
			continue
		}

		// Check latency
		if metrics, exists := enm.nodeMetrics[node.ID]; exists {
			if metrics.AverageLatency > criteria.MaxLatency {
				continue
			}
		}

		// Check capacity
		if node.Capacity-node.Available < criteria.MinCapacity {
			continue
		}

		// Check region preference
		if criteria.PreferredRegion != "" && node.Region != criteria.PreferredRegion {
			continue
		}

		candidates = append(candidates, node)
	}

	return candidates
}

// selectBestNode selects the best node from candidates
func (enm *EdgeNodeManager) selectBestNode(candidates []*EdgeNode, criteria SelectionCriteria) *EdgeNode {
	if len(candidates) == 0 {
		return nil
	}

	// Sort by priority: region match > latency > load
	sort.Slice(candidates, func(i, j int) bool {
		ni, nj := candidates[i], candidates[j]

		// Prefer region match
		if criteria.PreferredRegion != "" {
			iMatch := ni.Region == criteria.PreferredRegion
			jMatch := nj.Region == criteria.PreferredRegion
			if iMatch != jMatch {
				return iMatch
			}
		}

		// Prefer lower latency
		metricsI := enm.nodeMetrics[ni.ID]
		metricsJ := enm.nodeMetrics[nj.ID]
		if metricsI != nil && metricsJ != nil {
			if metricsI.AverageLatency != metricsJ.AverageLatency {
				return metricsI.AverageLatency < metricsJ.AverageLatency
			}
		}

		// Prefer lower load
		loadI := float64(ni.Capacity-ni.Available) / float64(ni.Capacity)
		loadJ := float64(nj.Capacity-nj.Available) / float64(nj.Capacity)
		return loadI < loadJ
	})

	return candidates[0]
}

// GetClosestNode returns the node with lowest latency
func (enm *EdgeNodeManager) GetClosestNode(ctx context.Context) (*EdgeNode, error) {
	selection, err := enm.SelectNode(SelectionCriteria{
		ExcludeOffline: true,
		MaxLatency:     500,
	})
	if err != nil {
		return nil, err
	}
	return selection.SelectedNode, nil
}

// GetNodesInRegion returns all healthy nodes in a region
func (enm *EdgeNodeManager) GetNodesInRegion(region string) []*EdgeNode {
	enm.mu.RLock()
	defer enm.mu.RUnlock()

	var result []*EdgeNode
	for _, node := range enm.nodes {
		if node.Region == region && node.Status == NodeOnline {
			result = append(result, node)
		}
	}
	return result
}

// GetNodeStatus returns current status of a specific node
func (enm *EdgeNodeManager) GetNodeStatus(nodeID string) *EdgeNodeMetrics {
	enm.mu.RLock()
	defer enm.mu.RUnlock()

	return enm.nodeMetrics[nodeID]
}

// GetAllNodes returns all known nodes
func (enm *EdgeNodeManager) GetAllNodes() []*EdgeNode {
	enm.mu.RLock()
	defer enm.mu.RUnlock()

	result := make([]*EdgeNode, 0, len(enm.nodes))
	for _, node := range enm.nodes {
		result = append(result, node)
	}
	return result
}

// GetNodeMetrics returns metrics for a specific node
func (enm *EdgeNodeManager) GetNodeMetrics(nodeID string) *EdgeNodeMetrics {
	enm.mu.RLock()
	defer enm.mu.RUnlock()

	if metrics, exists := enm.nodeMetrics[nodeID]; exists {
		// Return a copy to avoid external modification
		copy := *metrics
		return &copy
	}
	return nil
}

// GetHealthySummary returns a summary of healthy nodes by region
func (enm *EdgeNodeManager) GetHealthySummary() map[string]int {
	enm.mu.RLock()
	defer enm.mu.RUnlock()

	summary := make(map[string]int)
	for _, node := range enm.nodes {
		if node.Status == NodeOnline {
			summary[node.Region]++
		}
	}
	return summary
}

// ReportNodeLoad updates the current load on a node
func (enm *EdgeNodeManager) ReportNodeLoad(nodeID string, currentLoad, capacity int) error {
	enm.mu.Lock()
	defer enm.mu.Unlock()

	node, exists := enm.nodes[nodeID]
	if !exists {
		return errors.New("node not found")
	}

	node.Available = capacity - currentLoad
	node.Capacity = capacity

	if metrics, exists := enm.nodeMetrics[nodeID]; exists {
		metrics.CurrentLoad = currentLoad
		metrics.MaxCapacity = capacity
		metrics.UpdatedAt = time.Now()
	}

	return nil
}

// RegisterMetricsCallback registers a function to be called on metrics updates
func (enm *EdgeNodeManager) RegisterMetricsCallback(callback func(*EdgeNodeMetrics)) {
	enm.mu.Lock()
	defer enm.mu.Unlock()
	enm.metricsCallback = callback
}

// RefreshNode refreshes the status of a specific node
func (enm *EdgeNodeManager) RefreshNode(ctx context.Context, nodeID string) error {
	enm.mu.RLock()
	_, exists := enm.nodes[nodeID]
	enm.mu.RUnlock()

	if !exists {
		return errors.New("node not found")
	}

	// Trigger immediate health check
	enm.checkNodeHealth(ctx, nodeID)

	// Fetch fresh node data
	updated, err := enm.client.GetEdgeNode(ctx, nodeID)
	if err != nil {
		return err
	}

	enm.mu.Lock()
	enm.nodes[nodeID] = updated
	enm.mu.Unlock()

	return nil
}

// GetLoadBalancedNodes returns nodes sorted by current load (ascending)
func (enm *EdgeNodeManager) GetLoadBalancedNodes() []*EdgeNode {
	enm.mu.RLock()
	defer enm.mu.RUnlock()

	nodes := make([]*EdgeNode, 0, len(enm.nodes))
	for _, node := range enm.nodes {
		if node.Status == NodeOnline {
			nodes = append(nodes, node)
		}
	}

	// Sort by load percentage (ascending)
	sort.Slice(nodes, func(i, j int) bool {
		loadI := float64(nodes[i].Capacity-nodes[i].Available) / float64(nodes[i].Capacity)
		loadJ := float64(nodes[j].Capacity-nodes[j].Available) / float64(nodes[j].Capacity)
		return loadI < loadJ
	})

	return nodes
}

// GetNodesByLatency returns nodes sorted by latency (ascending)
func (enm *EdgeNodeManager) GetNodesByLatency() []*EdgeNode {
	enm.mu.RLock()
	defer enm.mu.RUnlock()

	nodes := make([]*EdgeNode, 0, len(enm.nodes))
	nodeMetrics := make(map[string]int64)

	for _, node := range enm.nodes {
		if node.Status == NodeOnline {
			nodes = append(nodes, node)
			if metrics, exists := enm.nodeMetrics[node.ID]; exists {
				nodeMetrics[node.ID] = metrics.AverageLatency
			}
		}
	}

	// Sort by latency (ascending)
	sort.Slice(nodes, func(i, j int) bool {
		return nodeMetrics[nodes[i].ID] < nodeMetrics[nodes[j].ID]
	})

	return nodes
}

// GetHotNodes returns nodes with high load (>80%)
func (enm *EdgeNodeManager) GetHotNodes() []*EdgeNode {
	enm.mu.RLock()
	defer enm.mu.RUnlock()

	var hot []*EdgeNode
	for _, node := range enm.nodes {
		load := float64(node.Capacity-node.Available) / float64(node.Capacity)
		if load > 0.8 {
			hot = append(hot, node)
		}
	}
	return hot
}

// GetColdNodes returns nodes with low load (<20%)
func (enm *EdgeNodeManager) GetColdNodes() []*EdgeNode {
	enm.mu.RLock()
	defer enm.mu.RUnlock()

	var cold []*EdgeNode
	for _, node := range enm.nodes {
		load := float64(node.Capacity-node.Available) / float64(node.Capacity)
		if load < 0.2 {
			cold = append(cold, node)
		}
	}
	return cold
}
