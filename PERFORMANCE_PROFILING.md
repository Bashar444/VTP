# Performance Profiling Guide

## Overview
The VTP platform includes built-in profiling capabilities using Go's `pprof` package. This allows real-time performance analysis of the running application.

## Available Endpoints

All profiling endpoints are available at `/debug/pprof/`:

### 1. **Heap Profile** - Memory Usage
```bash
# View in browser
http://localhost:8080/debug/pprof/heap

# Download for analysis
curl http://localhost:8080/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```

**What it shows:**
- Current memory allocation
- Objects in use
- Memory leaks

### 2. **CPU Profile** - CPU Usage
```bash
# Download 30-second CPU profile
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof

# Interactive web view
go tool pprof -http=:8081 cpu.prof
```

**What it shows:**
- CPU-intensive functions
- Hot paths in code
- Performance bottlenecks

### 3. **Goroutine Profile** - Concurrency
```bash
# View goroutines
curl http://localhost:8080/debug/pprof/goroutine > goroutine.prof
go tool pprof goroutine.prof

# Check count
curl http://localhost:8080/debug/pprof/goroutine?debug=1
```

**What it shows:**
- Active goroutines
- Goroutine leaks
- Blocking operations

### 4. **Block Profile** - Contention
```bash
# Enable block profiling (add to code)
runtime.SetBlockProfileRate(1)

# Download profile
curl http://localhost:8080/debug/pprof/block > block.prof
go tool pprof block.prof
```

**What it shows:**
- Lock contention
- Channel blocking
- Mutex waits

### 5. **Mutex Profile** - Lock Contention
```bash
# Enable mutex profiling (add to code)
runtime.SetMutexProfileFraction(1)

# Download profile
curl http://localhost:8080/debug/pprof/mutex > mutex.prof
go tool pprof mutex.prof
```

**What it shows:**
- Mutex contention
- Lock holding duration

### 6. **Trace** - Execution Trace
```bash
# Download 5-second trace
curl http://localhost:8080/debug/pprof/trace?seconds=5 > trace.out
go tool trace trace.out
```

**What it shows:**
- Goroutine execution timeline
- Network/syscall blocking
- GC pauses

## Common Profiling Workflows

### Identify Memory Leaks
```bash
# Take initial heap snapshot
curl http://localhost:8080/debug/pprof/heap > heap1.prof

# Run workload / wait some time

# Take second snapshot
curl http://localhost:8080/debug/pprof/heap > heap2.prof

# Compare
go tool pprof -base=heap1.prof heap2.prof
```

### Find CPU Bottlenecks
```bash
# Record during load test
curl "http://localhost:8080/debug/pprof/profile?seconds=60" > cpu.prof

# Analyze top functions
go tool pprof -top cpu.prof

# Interactive web UI
go tool pprof -http=:8081 cpu.prof
```

### Check for Goroutine Leaks
```bash
# Check goroutine count over time
curl http://localhost:8080/debug/pprof/goroutine?debug=1 | grep "goroutine profile:"

# Should remain stable - growing count indicates leak
```

### Analyze Database Query Performance
```bash
# Start profiling
curl "http://localhost:8080/debug/pprof/profile?seconds=30" > cpu.prof

# Run database-heavy workload during this time

# Look for database-related functions
go tool pprof -list="database/sql" cpu.prof
```

## Performance Benchmarks

### Key Metrics to Monitor

1. **Memory Usage**
   - Target: < 500MB for typical load
   - Heap allocations: minimize in hot paths
   - GC pause time: < 10ms

2. **CPU Usage**
   - Target: < 70% during normal operation
   - Identify functions taking > 5% CPU time

3. **Goroutines**
   - Expected: 50-200 for normal operation
   - > 1000 may indicate leaks

4. **Response Times**
   - API endpoints: < 100ms (p95)
   - Database queries: < 50ms (p95)
   - WebSocket messages: < 10ms

### Load Testing with Profiling
```bash
# Terminal 1: Start profiling
curl "http://localhost:8080/debug/pprof/profile?seconds=120" > load-test.prof &

# Terminal 2: Run load test
ab -n 10000 -c 100 http://localhost:8080/api/v1/courses

# Analyze results
go tool pprof -http=:8081 load-test.prof
```

## pprof Analysis Commands

### In Interactive Mode
```bash
go tool pprof cpu.prof

# Commands:
(pprof) top10          # Show top 10 functions
(pprof) list main      # Show annotated source for main package
(pprof) web            # Open graphviz visualization
(pprof) pdf            # Export to PDF
(pprof) help           # Show all commands
```

### Command-Line Analysis
```bash
# Top 20 functions by CPU time
go tool pprof -top20 cpu.prof

# Call graph as SVG
go tool pprof -svg cpu.prof > graph.svg

# Focus on specific package
go tool pprof -focus="recording" cpu.prof

# Ignore certain packages
go tool pprof -ignore="runtime" cpu.prof
```

## Security Considerations

⚠️ **WARNING**: pprof endpoints expose internal application state.

**Production Recommendations:**
1. Disable pprof in production by default
2. Use authentication/authorization for pprof endpoints
3. Only enable on-demand for debugging
4. Restrict access via firewall rules
5. Use separate debug port

**To secure pprof:**
```go
// Add authentication middleware
debugMux := http.NewServeMux()
debugMux.Handle("/debug/pprof/", authMiddleware.Middleware(http.DefaultServeMux))

// Listen on separate port
go http.ListenAndServe("localhost:6060", debugMux)
```

## Automated Profiling

### Continuous Profiling
For production monitoring, consider:
- [Pyroscope](https://pyroscope.io/) - continuous profiling
- [Datadog Continuous Profiler](https://www.datadoghq.com/product/code-profiling/)
- [Google Cloud Profiler](https://cloud.google.com/profiler)

### Scheduled Snapshots
```bash
# Cron job to capture daily profiles
0 2 * * * curl http://localhost:8080/debug/pprof/heap > /var/log/vtp/heap-$(date +\%Y\%m\%d).prof
```

## Common Issues & Solutions

### High Memory Usage
- Check for goroutine leaks (`/debug/pprof/goroutine`)
- Look for retained references in heap profile
- Review caching strategies
- Check for unbounded slices/maps

### High CPU Usage
- Profile hot paths with CPU profiler
- Check for expensive operations in loops
- Review database query efficiency
- Consider caching frequently computed values

### Slow Response Times
- Use trace to identify blocking operations
- Check for lock contention with mutex profiler
- Profile database queries
- Review network I/O patterns

## Example Analysis Session

```bash
# 1. Start application
./vtp-platform

# 2. Generate load
ab -n 5000 -c 50 http://localhost:8080/api/v1/recordings

# 3. Capture profiles
curl http://localhost:8080/debug/pprof/heap > heap.prof
curl "http://localhost:8080/debug/pprof/profile?seconds=30" > cpu.prof
curl http://localhost:8080/debug/pprof/goroutine > goroutine.prof

# 4. Analyze
go tool pprof -http=:8081 cpu.prof

# 5. Look for:
#    - Functions with high cumulative time
#    - Unexpected allocations
#    - Growing goroutine count
```

## Integration with Monitoring

The profiling data can be integrated with:
- **Prometheus**: Export runtime metrics
- **Grafana**: Visualize performance trends
- **Alerting**: Set thresholds on key metrics

See monitoring documentation for setup details.

## Resources

- [Official Go pprof docs](https://pkg.go.dev/net/http/pprof)
- [Profiling Go Programs](https://go.dev/blog/pprof)
- [High Performance Go Workshop](https://dave.cheney.net/high-performance-go-workshop/gopherchina-2019.html)
