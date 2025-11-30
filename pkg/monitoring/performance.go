package monitoring

import (
	"context"
	"log"
	"runtime"
	"time"
)

// PerformanceMonitor tracks and logs performance metrics
type PerformanceMonitor struct {
	interval time.Duration
	logger   *log.Logger
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(interval time.Duration, logger *log.Logger) *PerformanceMonitor {
	if logger == nil {
		logger = log.Default()
	}
	return &PerformanceMonitor{
		interval: interval,
		logger:   logger,
	}
}

// PerformanceMetrics holds current performance metrics
type PerformanceMetrics struct {
	Timestamp      time.Time
	NumGoroutines  int
	MemAllocMB     float64
	MemTotalMB     float64
	MemSysMB       float64
	NumGC          uint32
	GCPauseMs      float64
	HeapObjectsCount uint64
}

// Start begins monitoring in a background goroutine
func (pm *PerformanceMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(pm.interval)
	defer ticker.Stop()

	pm.logger.Println("Performance monitoring started")

	for {
		select {
		case <-ctx.Done():
			pm.logger.Println("Performance monitoring stopped")
			return
		case <-ticker.C:
			metrics := pm.CollectMetrics()
			pm.logMetrics(metrics)
		}
	}
}

// CollectMetrics gathers current performance metrics
func (pm *PerformanceMonitor) CollectMetrics() PerformanceMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return PerformanceMetrics{
		Timestamp:        time.Now(),
		NumGoroutines:    runtime.NumGoroutine(),
		MemAllocMB:       float64(m.Alloc) / 1024 / 1024,
		MemTotalMB:       float64(m.TotalAlloc) / 1024 / 1024,
		MemSysMB:         float64(m.Sys) / 1024 / 1024,
		NumGC:            m.NumGC,
		GCPauseMs:        float64(m.PauseNs[(m.NumGC+255)%256]) / 1000000,
		HeapObjectsCount: m.HeapObjects,
	}
}

// logMetrics logs the performance metrics
func (pm *PerformanceMonitor) logMetrics(m PerformanceMetrics) {
	pm.logger.Printf("[PERF] Goroutines: %d | Mem: %.2f MB | Sys: %.2f MB | GC: %d (%.2f ms) | Heap Objects: %d",
		m.NumGoroutines,
		m.MemAllocMB,
		m.MemSysMB,
		m.NumGC,
		m.GCPauseMs,
		m.HeapObjectsCount,
	)
}

// CheckThresholds validates metrics against thresholds and returns warnings
func (pm *PerformanceMonitor) CheckThresholds(m PerformanceMetrics) []string {
	var warnings []string

	// Check goroutine count
	if m.NumGoroutines > 1000 {
		warnings = append(warnings, "High goroutine count - possible leak")
	}

	// Check memory usage
	if m.MemAllocMB > 500 {
		warnings = append(warnings, "High memory allocation")
	}

	// Check heap objects
	if m.HeapObjectsCount > 1000000 {
		warnings = append(warnings, "High number of heap objects")
	}

	return warnings
}
