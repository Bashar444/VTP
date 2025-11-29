package g5

import (
	"context"
	"testing"
	"time"
)

func TestNewEdgeNodeManager(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	manager := NewEdgeNodeManager(client, 1000)

	if manager == nil {
		t.Fatal("expected manager instance, got nil")
	}

	if manager.client != client {
		t.Errorf("expected client to be set")
	}

	if manager.maxConnections != 1000 {
		t.Errorf("expected maxConnections 1000, got %d", manager.maxConnections)
	}
}

func TestEdgeNodeManagerStart(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	manager := NewEdgeNodeManager(client, 1000)
	ctx := context.Background()

	err := manager.Start(ctx)
	if err != nil {
		t.Logf("Start may fail due to network (expected): %v", err)
	}

	err = manager.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

func TestEdgeNodeManagerSelectNode(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	manager := NewEdgeNodeManager(client, 1000)

	// Add some test nodes manually
	manager.mu.Lock()
	manager.nodes["node-1"] = &EdgeNode{
		ID:        "node-1",
		Region:    "us-west",
		Status:    NodeOnline,
		Capacity:  1000,
		Available: 500,
		Latency:   15,
	}
	manager.nodes["node-2"] = &EdgeNode{
		ID:        "node-2",
		Region:    "us-east",
		Status:    NodeOnline,
		Capacity:  1000,
		Available: 600,
		Latency:   25,
	}
	manager.nodeMetrics["node-1"] = &EdgeNodeMetrics{
		NodeID:         "node-1",
		AverageLatency: 15,
	}
	manager.nodeMetrics["node-2"] = &EdgeNodeMetrics{
		NodeID:         "node-2",
		AverageLatency: 25,
	}
	manager.mu.Unlock()

	criteria := SelectionCriteria{
		ExcludeOffline: true,
		MaxLatency:     50,
	}

	selection, err := manager.SelectNode(criteria)
	if err != nil {
		t.Fatalf("SelectNode failed: %v", err)
	}

	if selection == nil {
		t.Errorf("expected selection, got nil")
	}

	if selection.SelectedNode == nil {
		t.Errorf("expected selected node, got nil")
	}
}

func TestEdgeNodeManagerSelectNodeNoNodes(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	manager := NewEdgeNodeManager(client, 1000)

	criteria := SelectionCriteria{
		ExcludeOffline: true,
	}

	_, err := manager.SelectNode(criteria)
	if err == nil {
		t.Errorf("expected error with no nodes, got nil")
	}
}

func TestEdgeNodeManagerGetClosestNode(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	manager := NewEdgeNodeManager(client, 1000)

	manager.mu.Lock()
	manager.nodes["node-1"] = &EdgeNode{
		ID:        "node-1",
		Region:    "us-west",
		Status:    NodeOnline,
		Capacity:  1000,
		Available: 500,
		Latency:   15,
	}
	manager.nodeMetrics["node-1"] = &EdgeNodeMetrics{
		NodeID:         "node-1",
		AverageLatency: 15,
	}
	manager.mu.Unlock()

	ctx := context.Background()
	node, err := manager.GetClosestNode(ctx)

	if err != nil {
		t.Fatalf("GetClosestNode failed: %v", err)
	}

	if node == nil {
		t.Errorf("expected node, got nil")
	}

	if node.ID != "node-1" {
		t.Errorf("expected node-1, got %s", node.ID)
	}
}

func TestEdgeNodeManagerGetNodesInRegion(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	manager := NewEdgeNodeManager(client, 1000)

	manager.mu.Lock()
	manager.nodes["node-1"] = &EdgeNode{
		ID:     "node-1",
		Region: "us-west",
		Status: NodeOnline,
	}
	manager.nodes["node-2"] = &EdgeNode{
		ID:     "node-2",
		Region: "us-west",
		Status: NodeOnline,
	}
	manager.nodes["node-3"] = &EdgeNode{
		ID:     "node-3",
		Region: "eu-west",
		Status: NodeOnline,
	}
	manager.mu.Unlock()

	nodes := manager.GetNodesInRegion("us-west")
	if len(nodes) != 2 {
		t.Errorf("expected 2 nodes in us-west, got %d", len(nodes))
	}
}

func TestEdgeNodeManagerGetNodeStatus(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	manager := NewEdgeNodeManager(client, 1000)

	manager.mu.Lock()
	manager.nodeMetrics["node-1"] = &EdgeNodeMetrics{
		NodeID:       "node-1",
		HealthStatus: "healthy",
	}
	manager.mu.Unlock()

	status := manager.GetNodeStatus("node-1")
	if status == nil {
		t.Errorf("expected status, got nil")
	}

	if status.HealthStatus != "healthy" {
		t.Errorf("expected healthy status, got %s", status.HealthStatus)
	}
}

func TestEdgeNodeManagerGetAllNodes(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	manager := NewEdgeNodeManager(client, 1000)

	manager.mu.Lock()
	manager.nodes["node-1"] = &EdgeNode{ID: "node-1"}
	manager.nodes["node-2"] = &EdgeNode{ID: "node-2"}
	manager.mu.Unlock()

	nodes := manager.GetAllNodes()
	if len(nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(nodes))
	}
}

func TestEdgeNodeManagerGetNodeMetrics(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	manager := NewEdgeNodeManager(client, 1000)

	manager.mu.Lock()
	manager.nodeMetrics["node-1"] = &EdgeNodeMetrics{
		NodeID:         "node-1",
		AverageLatency: 25,
	}
	manager.mu.Unlock()

	metrics := manager.GetNodeMetrics("node-1")
	if metrics == nil {
		t.Errorf("expected metrics, got nil")
	}

	if metrics.AverageLatency != 25 {
		t.Errorf("expected latency 25, got %d", metrics.AverageLatency)
	}
}

func TestEdgeNodeManagerReportNodeLoad(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	manager := NewEdgeNodeManager(client, 1000)

	manager.mu.Lock()
	manager.nodes["node-1"] = &EdgeNode{
		ID:        "node-1",
		Capacity:  1000,
		Available: 500,
	}
	manager.mu.Unlock()

	err := manager.ReportNodeLoad("node-1", 600, 1000)
	if err != nil {
		t.Errorf("ReportNodeLoad failed: %v", err)
	}

	manager.mu.RLock()
	node := manager.nodes["node-1"]
	manager.mu.RUnlock()

	if node.Available != 400 {
		t.Errorf("expected available 400, got %d", node.Available)
	}
}

func TestEdgeNodeManagerReportNodeLoadNotFound(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	manager := NewEdgeNodeManager(client, 1000)

	err := manager.ReportNodeLoad("nonexistent", 100, 1000)
	if err == nil {
		t.Errorf("expected error for nonexistent node, got nil")
	}
}

func TestEdgeNodeManagerGetLoadBalancedNodes(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	manager := NewEdgeNodeManager(client, 1000)

	manager.mu.Lock()
	manager.nodes["node-1"] = &EdgeNode{
		ID:        "node-1",
		Status:    NodeOnline,
		Capacity:  1000,
		Available: 200,
	}
	manager.nodes["node-2"] = &EdgeNode{
		ID:        "node-2",
		Status:    NodeOnline,
		Capacity:  1000,
		Available: 500,
	}
	manager.mu.Unlock()

	nodes := manager.GetLoadBalancedNodes()
	if len(nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(nodes))
	}

	// node-2 should be first (lower load)
	if nodes[0].ID != "node-2" {
		t.Errorf("expected node-2 first (lower load), got %s", nodes[0].ID)
	}
}

func TestEdgeNodeManagerGetNodesByLatency(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	manager := NewEdgeNodeManager(client, 1000)

	manager.mu.Lock()
	manager.nodes["node-1"] = &EdgeNode{
		ID:     "node-1",
		Status: NodeOnline,
	}
	manager.nodes["node-2"] = &EdgeNode{
		ID:     "node-2",
		Status: NodeOnline,
	}
	manager.nodeMetrics["node-1"] = &EdgeNodeMetrics{
		NodeID:         "node-1",
		AverageLatency: 50,
	}
	manager.nodeMetrics["node-2"] = &EdgeNodeMetrics{
		NodeID:         "node-2",
		AverageLatency: 20,
	}
	manager.mu.Unlock()

	nodes := manager.GetNodesByLatency()
	if len(nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(nodes))
	}

	// node-2 should be first (lower latency)
	if nodes[0].ID != "node-2" {
		t.Errorf("expected node-2 first (lower latency), got %s", nodes[0].ID)
	}
}

func TestEdgeNodeManagerGetHotNodes(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	manager := NewEdgeNodeManager(client, 1000)

	manager.mu.Lock()
	manager.nodes["node-1"] = &EdgeNode{
		ID:        "node-1",
		Capacity:  1000,
		Available: 100, // 90% load - hot
	}
	manager.nodes["node-2"] = &EdgeNode{
		ID:        "node-2",
		Capacity:  1000,
		Available: 500, // 50% load - not hot
	}
	manager.mu.Unlock()

	hotNodes := manager.GetHotNodes()
	if len(hotNodes) != 1 {
		t.Errorf("expected 1 hot node, got %d", len(hotNodes))
	}

	if hotNodes[0].ID != "node-1" {
		t.Errorf("expected node-1, got %s", hotNodes[0].ID)
	}
}

func TestEdgeNodeManagerGetColdNodes(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	manager := NewEdgeNodeManager(client, 1000)

	manager.mu.Lock()
	manager.nodes["node-1"] = &EdgeNode{
		ID:        "node-1",
		Capacity:  1000,
		Available: 900, // 10% load - cold
	}
	manager.nodes["node-2"] = &EdgeNode{
		ID:        "node-2",
		Capacity:  1000,
		Available: 500, // 50% load - not cold
	}
	manager.mu.Unlock()

	coldNodes := manager.GetColdNodes()
	if len(coldNodes) != 1 {
		t.Errorf("expected 1 cold node, got %d", len(coldNodes))
	}

	if coldNodes[0].ID != "node-1" {
		t.Errorf("expected node-1, got %s", coldNodes[0].ID)
	}
}

func TestEdgeNodeManagerRegisterMetricsCallback(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	manager := NewEdgeNodeManager(client, 1000)

	callbackCalled := false
	callback := func(metrics *EdgeNodeMetrics) {
		callbackCalled = true
	}

	manager.RegisterMetricsCallback(callback)

	// Callback registered - log for now
	if !callbackCalled {
		t.Logf("Metrics callback registered but not called yet (expected)")
	}
	if manager.metricsCallback == nil {
		t.Errorf("expected callback to be set")
	}
}

func TestEdgeNodeManagerConcurrency(t *testing.T) {
	client := NewClient("https://api.5g.local", nil)
	manager := NewEdgeNodeManager(client, 1000)

	manager.mu.Lock()
	manager.nodes["node-1"] = &EdgeNode{
		ID:        "node-1",
		Capacity:  1000,
		Available: 500,
	}
	manager.mu.Unlock()

	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				manager.GetAllNodes()
				manager.GetLoadBalancedNodes()
				manager.GetHotNodes()
				time.Sleep(time.Millisecond)
			}
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}
}
