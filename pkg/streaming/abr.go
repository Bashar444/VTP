package streaming

import (
	"log"
	"math"
	"time"
)

// NewAdaptiveBitrateManager creates new ABR manager with default configuration
func NewAdaptiveBitrateManager(config ABRConfig) *AdaptiveBitrateManager {
	if config.MinBitrate == 0 {
		config.MinBitrate = 500
	}
	if config.MaxBitrate == 0 {
		config.MaxBitrate = 4000
	}
	if config.ThresholdUp == 0 {
		config.ThresholdUp = 1.5
	}
	if config.ThresholdDown == 0 {
		config.ThresholdDown = 0.5
	}
	if config.HistorySize == 0 {
		config.HistorySize = 10
	}

	return &AdaptiveBitrateManager{
		config:            config,
		currentBitrate:    config.MinBitrate,
		segmentHistory:    make([]SegmentMetrics, 0),
		networkHistory:    make([]NetworkStats, 0),
		availableBitrates: []int{500, 1000, 2000, 4000},
	}
}

// RecordSegmentMetrics records metrics for downloaded segment
func (abr *AdaptiveBitrateManager) RecordSegmentMetrics(metrics SegmentMetrics) {
	abr.segmentHistory = append(abr.segmentHistory, metrics)

	// Keep only recent history
	if len(abr.segmentHistory) > abr.config.HistorySize {
		abr.segmentHistory = abr.segmentHistory[1:]
	}

	// Estimate bandwidth from this segment
	if metrics.DownloadTime > 0 {
		segmentBitrate := (metrics.BytesDownloaded * 8 * 1000) / metrics.DownloadTime
		log.Printf("[ABR] Segment %d: Downloaded %d bytes in %dms (%.0f kbps)",
			metrics.SegmentNumber, metrics.BytesDownloaded, metrics.DownloadTime, float64(segmentBitrate))
	}
}

// RecordNetworkStats records network condition snapshot
func (abr *AdaptiveBitrateManager) RecordNetworkStats(stats NetworkStats) {
	stats.Timestamp = time.Now()
	abr.networkHistory = append(abr.networkHistory, stats)

	// Keep only recent history
	if len(abr.networkHistory) > abr.config.HistorySize {
		abr.networkHistory = abr.networkHistory[1:]
	}

	log.Printf("[ABR] Network: %.0f kbps, %dms latency, %.1f%% loss, buffer: %.0f%%",
		stats.Bandwidth, stats.Latency, stats.PacketLoss, stats.BufferHealth)
}

// SelectQuality determines best quality for current bandwidth
func (abr *AdaptiveBitrateManager) SelectQuality(bandwidth float64) BitrateLevel {
	selectedBitrate := abr.currentBitrate

	// Safety bounds
	if selectedBitrate < abr.config.MinBitrate {
		selectedBitrate = abr.config.MinBitrate
	}
	if selectedBitrate > abr.config.MaxBitrate {
		selectedBitrate = abr.config.MaxBitrate
	}

	// Try to match available bitrate to bandwidth
	bestBitrate := selectedBitrate
	bestDiff := math.Abs(float64(selectedBitrate) - bandwidth)

	for _, bitrate := range abr.availableBitrates {
		// Don't select bitrate higher than available bandwidth with safety margin
		if float64(bitrate) > bandwidth*abr.config.ThresholdUp {
			continue
		}
		diff := math.Abs(float64(bitrate) - bandwidth)
		if diff < bestDiff {
			bestDiff = diff
			bestBitrate = bitrate
		}
	}

	return bitrateToLevel(bestBitrate)
}

// ShouldUpscale checks if should switch to higher quality
func (abr *AdaptiveBitrateManager) ShouldUpscale() bool {
	if len(abr.segmentHistory) < 3 {
		return false
	}

	// Get average bitrate of last 3 segments
	avgBitrate := 0.0
	for _, seg := range abr.segmentHistory[len(abr.segmentHistory)-3:] {
		avgBitrate += float64(seg.Bitrate)
	}
	avgBitrate /= 3.0

	// If avg bitrate is significantly higher than current, upscale
	nextBitrate := findNextBitrate(abr.currentBitrate, abr.availableBitrates)
	return nextBitrate > 0 && avgBitrate > float64(nextBitrate)*abr.config.ThresholdUp
}

// ShouldDownscale checks if should switch to lower quality
func (abr *AdaptiveBitrateManager) ShouldDownscale() bool {
	if len(abr.segmentHistory) < 3 {
		return false
	}

	// Check buffer health
	var bufferSum float64
	for _, seg := range abr.segmentHistory[len(abr.segmentHistory)-3:] {
		bufferSum += seg.BufferOccupancy
	}
	avgBuffer := bufferSum / 3.0

	// If buffer is low, downscale
	prevBitrate := findPrevBitrate(abr.currentBitrate, abr.availableBitrates)
	return prevBitrate > 0 && avgBuffer < 20.0
}

// GetCurrentBitrate returns current selected bitrate in kbps
func (abr *AdaptiveBitrateManager) GetCurrentBitrate() int {
	return abr.currentBitrate
}

// UpdateCurrentBitrate sets the current bitrate
func (abr *AdaptiveBitrateManager) UpdateCurrentBitrate(bitrate int) {
	abr.currentBitrate = bitrate
	log.Printf("[ABR] Bitrate updated to %d kbps", bitrate)
}

// GetAvailableBitrates returns list of available bitrates
func (abr *AdaptiveBitrateManager) GetAvailableBitrates() []int {
	return abr.availableBitrates
}

// PredictOptimalBitrate predicts best bitrate for next segment
func (abr *AdaptiveBitrateManager) PredictOptimalBitrate() int {
	if len(abr.segmentHistory) < 2 {
		return abr.currentBitrate
	}

	// Calculate average download time of recent segments
	var totalTime int64
	count := len(abr.segmentHistory)
	if count > 5 {
		count = 5
	}

	for i := len(abr.segmentHistory) - count; i < len(abr.segmentHistory); i++ {
		totalTime += int64(abr.segmentHistory[i].DownloadTime)
	}
	avgDownloadTime := totalTime / int64(count)

	// Assume segment is ~2MB
	const segmentSize = 2000000                                             // 2MB in bytes
	predictedBitrate := (segmentSize * 8 * 1000) / (avgDownloadTime * 1000) // kbps

	// Clamp to available bitrates
	if predictedBitrate < int64(abr.config.MinBitrate) {
		predictedBitrate = int64(abr.config.MinBitrate)
	}
	if predictedBitrate > int64(abr.config.MaxBitrate) {
		predictedBitrate = int64(abr.config.MaxBitrate)
	}

	return int(predictedBitrate)
}

// GetStatistics returns current ABR statistics
func (abr *AdaptiveBitrateManager) GetStatistics() map[string]interface{} {
	stats := make(map[string]interface{})

	stats["current_bitrate"] = abr.currentBitrate
	stats["segments_recorded"] = len(abr.segmentHistory)
	stats["available_bitrates"] = abr.availableBitrates

	if len(abr.segmentHistory) > 0 {
		lastSegment := abr.segmentHistory[len(abr.segmentHistory)-1]
		stats["last_segment_bitrate"] = lastSegment.Bitrate
		stats["last_buffer_occupancy"] = lastSegment.BufferOccupancy
		stats["last_download_time_ms"] = lastSegment.DownloadTime
	}

	if len(abr.networkHistory) > 0 {
		lastNetwork := abr.networkHistory[len(abr.networkHistory)-1]
		stats["last_bandwidth_kbps"] = lastNetwork.Bandwidth
		stats["last_latency_ms"] = lastNetwork.Latency
		stats["last_packet_loss"] = lastNetwork.PacketLoss
	}

	return stats
}

// Helper functions

func bitrateToLevel(bitrate int) BitrateLevel {
	switch bitrate {
	case 500:
		return BitrateVeryLow
	case 1000:
		return BitrateLow
	case 2000:
		return BitrateMedium
	case 4000:
		return BitrateHigh
	default:
		return BitrateMedium
	}
}

func findNextBitrate(current int, available []int) int {
	for _, b := range available {
		if b > current {
			return b
		}
	}
	return -1
}

func findPrevBitrate(current int, available []int) int {
	var prev int
	for _, b := range available {
		if b >= current {
			break
		}
		prev = b
	}
	if prev == 0 {
		return -1
	}
	return prev
}
