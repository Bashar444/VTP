# Phase 5G - Quick Usage Guide

## Basic Setup

### 1. Initialize the Adapter

```go
import "github.com/yourusername/vtp-platform/pkg/g5"

// Use default configuration
adapter, err := g5.NewAdapter(nil)
if err != nil {
    log.Fatal(err)
}

// Or with custom configuration
config := &g5.AdapterConfig{
    APIBaseURL:              "https://api.5g.vtp.local",
    DetectionInterval:       2 * time.Second,
    EnableMetricsCollection: true,
    TargetLatency:           50,
}
adapter, err := g5.NewAdapter(config)
```

### 2. Start the Adapter

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

err := adapter.Start(ctx)
if err != nil {
    log.Fatal("Failed to start adapter:", err)
}
defer adapter.Stop()
```

### 3. Register Callbacks (Optional)

```go
// Status updates
adapter.RegisterStatusCallback(func(status *g5.AdapterStatus) {
    log.Printf("Adapter Status: Healthy=%v, Quality=%s", 
        status.IsHealthy, status.CurrentQuality)
})

// Warning events
adapter.RegisterWarningCallback(func(warning *g5.AdapterWarning) {
    log.Printf("[%s] %s: %s", 
        warning.Level, warning.Code, warning.Message)
})

// Metrics updates
adapter.RegisterMetricsCallback(func(metrics *g5.SessionMetrics) {
    log.Printf("Session %s: Latency=%dms, Quality=%s", 
        metrics.SessionID, metrics.AvgLatency, metrics.VideoQuality)
})
```

---

## Usage Examples

### Example 1: Check Network Status

```go
// Get current network
network := adapter.GetCurrentNetwork()
if network != nil {
    log.Printf("Type: %s, Latency: %dms, Bandwidth: %d Kbps",
        network.Type, network.Latency, network.Bandwidth)
}

// Check 5G availability
if adapter.Is5GAvailable() {
    log.Println("5G is available!")
}

// Get quality score (0-100)
quality := adapter.GetNetworkQuality()
log.Printf("Network Quality: %d%%", quality)
```

### Example 2: Start a Session

```go
// Start new streaming session
sessionID := "session-123"
err := adapter.StartSession(sessionID)
if err != nil {
    log.Printf("Failed to start session: %v", err)
}

// Do something with session...
time.Sleep(10 * time.Second)

// End session
adapter.EndSession()

// Get session metrics
metrics := adapter.GetSessionMetrics()
if metrics != nil {
    log.Printf("Session Duration: %dms", metrics.Duration)
    log.Printf("Avg Latency: %dms", metrics.AvgLatency)
    log.Printf("Avg Bandwidth: %d Kbps", metrics.AvgBandwidth)
    log.Printf("Packet Loss: %.2f%%", metrics.AvgPacketLoss*100)
}
```

### Example 3: Record Metrics During Session

```go
// During an active session, record metrics
adapter.RecordMetric("latency", int64(25))
adapter.RecordMetric("bandwidth", int64(15000))
adapter.RecordMetric("packetLoss", float32(0.02))
adapter.RecordMetric("jitter", int64(5))
adapter.RecordMetric("frameDropped")
```

### Example 4: Adapt Quality

```go
// Manually trigger quality adaptation
newQuality, err := adapter.AdaptQuality()
if err != nil {
    log.Printf("Quality adaptation failed: %v", err)
}

log.Printf("New Quality Level: %s", newQuality)

// Get current quality profile
profile := adapter.GetStatus().CurrentQuality
log.Printf("Current Quality: %s", profile)
```

### Example 5: Edge Node Selection

```go
// Get all available edge nodes
nodes := adapter.GetAvailableEdgeNodes()
log.Printf("Available edge nodes: %d", len(nodes))

// Get closest edge node
ctx := context.Background()
closest, err := adapter.GetClosestEdgeNode(ctx)
if err == nil {
    log.Printf("Closest node: %s (%dms latency)", 
        closest.ID, closest.Latency)
}
```

### Example 6: Get Global Metrics

```go
// Get aggregate metrics across all sessions
globalMetrics := adapter.GetGlobalMetrics()
if globalMetrics != nil {
    log.Printf("Active Sessions: %d", globalMetrics.ActiveSessions)
    log.Printf("Global Avg Latency: %dms", globalMetrics.GlobalAvgLatency)
    log.Printf("Global Avg Bandwidth: %d Kbps", globalMetrics.GlobalAvgBandwidth)
    log.Printf("Total Data Transferred: %d bytes", globalMetrics.TotalBytesTransfer)
}
```

### Example 7: Check Adapter Status

```go
// Get full status
status := adapter.GetStatus()

log.Printf("Adapter Running: %v", status.IsStarted)
log.Printf("Adapter Healthy: %v", status.IsHealthy)
log.Printf("Current Quality: %s", status.CurrentQuality)
log.Printf("Active Sessions: %d", status.TotalActiveSessions)

if status.ActiveSessionID != "" {
    log.Printf("Current Session: %s on %s", 
        status.ActiveSessionID, status.ActiveEdgeNode)
}
```

---

## REST API Examples

### Get Network Status

```bash
curl -X GET http://localhost:8000/api/network/status
```

**Response**:
```json
{
  "type": "5g",
  "latency_ms": 25,
  "bandwidth_kbps": 50000,
  "signal_strength": -95,
  "is_5g_available": true,
  "quality_score": 92,
  "timestamp": 1733241600000
}
```

### Detect Network

```bash
curl -X POST http://localhost:8000/api/network/detect
```

### Get Edge Nodes

```bash
curl -X GET http://localhost:8000/api/edge/nodes
```

**Response**:
```json
{
  "nodes": [
    {
      "id": "edge-us-east-1",
      "region": "us-east",
      "country": "USA",
      "endpoint": "https://edge-us-east-1.vtp.local",
      "latency_ms": 15,
      "status": "online",
      "load_percentage": 45,
      "available_capacity": 550,
      "max_capacity": 1000
    }
  ],
  "count": 3
}
```

### Connect to Edge Node

```bash
curl -X POST http://localhost:8000/api/edge/connect \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "sess-abc123",
    "node_id": "edge-us-east-1"
  }'
```

### Get Session Metrics

```bash
curl -X GET http://localhost:8000/api/session/metrics
```

### Adapt Quality

```bash
curl -X POST http://localhost:8000/api/quality/adapt
```

### Get Adapter Status

```bash
curl -X GET http://localhost:8000/api/adapter/status
```

---

## Configuration Examples

### High Performance Configuration

```go
config := &g5.AdapterConfig{
    APIBaseURL:              "https://api.5g.vtp.local",
    DetectionInterval:       1 * time.Second,      // More frequent
    HealthCheckInterval:     15 * time.Second,     // More frequent
    MetricsReportInterval:   5 * time.Second,      // More frequent
    MaxEdgeConnections:      20,                   // Higher capacity
    EnableMetricsCollection: true,
    EnableAutoQualityAdapt:  true,
    TargetLatency:           30,                   // Stricter target
    TargetBandwidth:         30000,                // Higher bandwidth
    QualitySwitchThreshold:  5,                    // More sensitive
}
adapter, _ := g5.NewAdapter(config)
```

### Power-Saving Configuration

```go
config := &g5.AdapterConfig{
    APIBaseURL:              "https://api.5g.vtp.local",
    DetectionInterval:       5 * time.Second,      // Less frequent
    HealthCheckInterval:     60 * time.Second,     // Less frequent
    MetricsReportInterval:   30 * time.Second,     // Less frequent
    MaxEdgeConnections:      5,                    // Lower capacity
    EnableMetricsCollection: false,                // Disable metrics
    EnableAutoQualityAdapt:  true,
    TargetLatency:           100,                  // Relaxed target
    TargetBandwidth:         10000,                // Lower bandwidth
    QualitySwitchThreshold:  20,                   // Less sensitive
}
adapter, _ := g5.NewAdapter(config)
```

---

## Common Workflows

### Workflow 1: Detect Network and Start Session

```go
// Initialize adapter
adapter, _ := g5.NewAdapter(nil)
ctx := context.Background()
adapter.Start(ctx)

// Detect network
network, _ := adapter.DetectNetworkType(ctx)
log.Printf("Network: %s (Latency: %dms)", network.Type, network.Latency)

// Start session
adapter.StartSession("my-session")

// Do work...

// End session and get metrics
adapter.EndSession()
metrics := adapter.GetSessionMetrics()
log.Printf("Total Duration: %dms", metrics.Duration)

adapter.Stop()
```

### Workflow 2: Monitor Network Quality

```go
adapter, _ := g5.NewAdapter(nil)
adapter.Start(context.Background())

// Monitor quality periodically
ticker := time.NewTicker(5 * time.Second)
for range ticker.C {
    quality := adapter.GetNetworkQuality()
    if quality < 50 {
        log.Println("Network quality degraded!")
        // Trigger quality adaptation
        newQuality, _ := adapter.AdaptQuality()
        log.Printf("Adapted to: %s", newQuality)
    }
}

adapter.Stop()
```

### Workflow 3: Select Best Edge Node

```go
adapter, _ := g5.NewAdapter(nil)
adapter.Start(context.Background())

// Get closest edge node
ctx := context.Background()
closest, _ := adapter.GetClosestEdgeNode(ctx)
log.Printf("Connecting to: %s (%dms)", closest.ID, closest.Latency)

// Start session on that node
adapter.StartSession("session-1")

adapter.Stop()
```

---

## Error Handling

```go
// Start session with error handling
err := adapter.StartSession("my-session")
if err != nil {
    if err.Error() == "adapter not started" {
        log.Println("Adapter must be started first")
    } else if err.Error() == "session already active" {
        log.Println("Already have an active session")
    } else {
        log.Printf("Failed to start session: %v", err)
    }
}

// Adapt quality with error handling
quality, err := adapter.AdaptQuality()
if err != nil {
    if err.Error() == "adapter not started" {
        log.Println("Adapter is not running")
    } else {
        log.Printf("Quality adaptation failed: %v", err)
    }
}
```

---

## Performance Tips

1. **Reuse Adapter**: Create one adapter per application, not per session
2. **Batch Metrics**: Record metrics in batches rather than individual calls
3. **Limit Callbacks**: Avoid heavy processing in callbacks
4. **Cleanup Sessions**: Use `ClearOldSessions()` to manage memory
5. **Monitor Metrics**: Check `GlobalMetrics` periodically to detect trends
6. **Configure Intervals**: Adjust detection/health-check intervals based on needs

---

## Debugging

### Enable Logging

```go
// Check adapter status frequently
go func() {
    ticker := time.NewTicker(10 * time.Second)
    for range ticker.C {
        status := adapter.GetStatus()
        log.Printf("Adapter: %+v", status)
    }
}()
```

### Collect Metrics

```go
// Get all session metrics
allMetrics := metricsCollector.GetAllSessionMetrics()
for sessionID, metrics := range allMetrics {
    log.Printf("Session %s: %+v", sessionID, metrics)
}
```

### Monitor Edge Nodes

```go
// Check node health
nodes := adapter.GetAvailableEdgeNodes()
for _, node := range nodes {
    log.Printf("Node %s: Status=%s, Load=%.1f%%", 
        node.ID, node.Status, node.Load)
}
```

---

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Adapter won't start | Check API endpoint is reachable |
| No edge nodes available | Verify edge discovery service is running |
| Sessions not tracked | Enable `EnableMetricsCollection` in config |
| Quality not adapting | Set `EnableAutoQualityAdapt: true` |
| High latency detected | Use edge nodes with lower latency |
| Metrics not updating | Register callbacks to receive updates |

---

*For complete implementation details, see PHASE_5G_CORE_COMPLETION.md*
