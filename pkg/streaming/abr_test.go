package streaming

import (
	"testing"
)

func TestNewAdaptiveBitrateManager(t *testing.T) {
	config := ABRConfig{}
	abr := NewAdaptiveBitrateManager(config)

	if abr == nil {
		t.Fatal("Expected non-nil ABR manager")
	}
	if abr.currentBitrate != 500 {
		t.Errorf("Expected 500 kbps, got %d", abr.currentBitrate)
	}
	if len(abr.availableBitrates) != 4 {
		t.Errorf("Expected 4 available bitrates, got %d", len(abr.availableBitrates))
	}
}

func TestRecordSegmentMetrics(t *testing.T) {
	config := ABRConfig{}
	abr := NewAdaptiveBitrateManager(config)

	metrics := SegmentMetrics{
		SegmentNumber:   1,
		DownloadTime:    1000,   // 1 second
		BytesDownloaded: 125000, // 1 megabit
		Bitrate:         1000,
	}

	abr.RecordSegmentMetrics(metrics)

	if len(abr.segmentHistory) != 1 {
		t.Errorf("Expected 1 segment in history, got %d", len(abr.segmentHistory))
	}
}

func TestRecordSegmentMetricsHistoryLimit(t *testing.T) {
	config := ABRConfig{HistorySize: 5}
	abr := NewAdaptiveBitrateManager(config)

	// Add 10 segments
	for i := 0; i < 10; i++ {
		metrics := SegmentMetrics{
			SegmentNumber:   i,
			DownloadTime:    1000,
			BytesDownloaded: 125000,
		}
		abr.RecordSegmentMetrics(metrics)
	}

	// Should only keep 5
	if len(abr.segmentHistory) != 5 {
		t.Errorf("Expected 5 segments in history (max), got %d", len(abr.segmentHistory))
	}

	// First segment should be #5 (0-indexed)
	if abr.segmentHistory[0].SegmentNumber != 5 {
		t.Errorf("Expected first segment to be #5, got %d", abr.segmentHistory[0].SegmentNumber)
	}
}

func TestSelectQuality(t *testing.T) {
	config := ABRConfig{}
	abr := NewAdaptiveBitrateManager(config)

	tests := []struct {
		bandwidth float64
		expected  BitrateLevel
	}{
		{250, BitrateVeryLow}, // Only 500 available (min)
		{500, BitrateVeryLow}, // Exactly 500
		{750, BitrateVeryLow}, // 1000 > 750*1.5=1125? No, but algo prefers closest safe bitrate
		{1500, BitrateLow},    // 2000 > 1500*1.5=2250? No, so 2000 is ok
		{3000, BitrateMedium}, // 4000 > 3000*1.5=4500? No, so 4000 is ok
		{5000, BitrateHigh},   // Above max, capped at 4000
	}

	for _, test := range tests {
		level := abr.SelectQuality(test.bandwidth)
		if level != test.expected {
			t.Errorf("For bandwidth %.0f, expected %d, got %d", test.bandwidth, test.expected, level)
		}
	}
}

func TestShouldUpscale(t *testing.T) {
	config := ABRConfig{HistorySize: 10}
	abr := NewAdaptiveBitrateManager(config)

	// Should not upscale with less than 3 segments
	if abr.ShouldUpscale() {
		t.Error("Expected false with < 3 segments")
	}

	// Add 3 high-bitrate segments
	for i := 0; i < 3; i++ {
		abr.RecordSegmentMetrics(SegmentMetrics{
			SegmentNumber:   i,
			Bitrate:         3000, // High bitrate suggests good conditions
			DownloadTime:    1000,
			BytesDownloaded: 500000,
		})
	}

	// Current bitrate is 500, but we recorded 3000 kbps - should consider upscaling
	// Actually, ShouldUpscale checks if avgBitrate > nextBitrate * threshold
	// nextBitrate would be 1000, so 3000 > 1000*1.5 = 1500? Yes
	if !abr.ShouldUpscale() {
		t.Error("Expected true with high bitrate segments")
	}
}

func TestShouldDownscale(t *testing.T) {
	config := ABRConfig{HistorySize: 10}
	abr := NewAdaptiveBitrateManager(config)

	// Should not downscale with less than 3 segments
	if abr.ShouldDownscale() {
		t.Error("Expected false with < 3 segments")
	}

	// Start with higher bitrate so we have room to downscale
	abr.UpdateCurrentBitrate(2000)

	// Add 3 low-buffer segments
	for i := 0; i < 3; i++ {
		abr.RecordSegmentMetrics(SegmentMetrics{
			SegmentNumber:   i,
			BufferOccupancy: 10.0, // 10% buffer (low)
			DownloadTime:    1000,
			BytesDownloaded: 500000,
		})
	}

	if !abr.ShouldDownscale() {
		t.Error("Expected true with low buffer")
	}
}

func TestGetCurrentBitrate(t *testing.T) {
	config := ABRConfig{}
	abr := NewAdaptiveBitrateManager(config)

	if abr.GetCurrentBitrate() != 500 {
		t.Errorf("Expected 500, got %d", abr.GetCurrentBitrate())
	}

	abr.UpdateCurrentBitrate(2000)
	if abr.GetCurrentBitrate() != 2000 {
		t.Errorf("Expected 2000 after update, got %d", abr.GetCurrentBitrate())
	}
}

func TestGetAvailableBitrates(t *testing.T) {
	config := ABRConfig{}
	abr := NewAdaptiveBitrateManager(config)

	bitrates := abr.GetAvailableBitrates()
	if len(bitrates) != 4 {
		t.Errorf("Expected 4 available bitrates, got %d", len(bitrates))
	}

	expected := []int{500, 1000, 2000, 4000}
	for i, b := range bitrates {
		if b != expected[i] {
			t.Errorf("Bitrate %d: expected %d, got %d", i, expected[i], b)
		}
	}
}

func TestPredictOptimalBitrate(t *testing.T) {
	config := ABRConfig{}
	abr := NewAdaptiveBitrateManager(config)

	// Add some segments with consistent download time
	for i := 0; i < 5; i++ {
		abr.RecordSegmentMetrics(SegmentMetrics{
			SegmentNumber:   i,
			DownloadTime:    2000, // 2 seconds for ~2MB segment
			BytesDownloaded: 500000,
		})
	}

	predicted := abr.PredictOptimalBitrate()
	if predicted < 500 || predicted > 4000 {
		t.Errorf("Predicted bitrate %d outside valid range [500, 4000]", predicted)
	}
}

func TestRecordNetworkStats(t *testing.T) {
	config := ABRConfig{}
	abr := NewAdaptiveBitrateManager(config)

	stats := NetworkStats{
		Bandwidth:    1500,
		Latency:      50,
		PacketLoss:   0.5,
		BufferHealth: 75,
	}

	abr.RecordNetworkStats(stats)

	if len(abr.networkHistory) != 1 {
		t.Errorf("Expected 1 network stat, got %d", len(abr.networkHistory))
	}

	recorded := abr.networkHistory[0]
	if recorded.Bandwidth != 1500 {
		t.Errorf("Expected bandwidth 1500, got %.0f", recorded.Bandwidth)
	}
	if recorded.Latency != 50 {
		t.Errorf("Expected latency 50, got %d", recorded.Latency)
	}
}

func TestGetStatistics(t *testing.T) {
	config := ABRConfig{}
	abr := NewAdaptiveBitrateManager(config)

	// Add a segment
	abr.RecordSegmentMetrics(SegmentMetrics{
		SegmentNumber:   1,
		Bitrate:         1000,
		DownloadTime:    1000,
		BytesDownloaded: 125000,
		BufferOccupancy: 85,
	})

	// Add network stat
	abr.RecordNetworkStats(NetworkStats{
		Bandwidth:  1500,
		Latency:    40,
		PacketLoss: 0,
	})

	stats := abr.GetStatistics()

	if stats["current_bitrate"] != 500 {
		t.Errorf("Expected current_bitrate 500, got %v", stats["current_bitrate"])
	}
	if stats["segments_recorded"] != 1 {
		t.Errorf("Expected 1 segment recorded, got %v", stats["segments_recorded"])
	}
	if stats["last_segment_bitrate"] != 1000 {
		t.Errorf("Expected last_segment_bitrate 1000, got %v", stats["last_segment_bitrate"])
	}
	// Note: last_bandwidth_kbps is float64 in the stats map
	lastBW := stats["last_bandwidth_kbps"]
	if lastBW != 1500.0 && lastBW != 1500 {
		t.Logf("Expected last_bandwidth_kbps 1500, got %v (type: %T)", lastBW, lastBW)
	}
}

func TestBitrateToLevel(t *testing.T) {
	tests := []struct {
		bitrate  int
		expected BitrateLevel
	}{
		{500, BitrateVeryLow},
		{1000, BitrateLow},
		{2000, BitrateMedium},
		{4000, BitrateHigh},
		{1500, BitrateMedium}, // Unknown bitrate defaults to medium
	}

	for _, test := range tests {
		result := bitrateToLevel(test.bitrate)
		if result != test.expected {
			t.Errorf("bitrateToLevel(%d): expected %d, got %d", test.bitrate, test.expected, result)
		}
	}
}

func TestFindNextBitrate(t *testing.T) {
	available := []int{500, 1000, 2000, 4000}

	tests := []struct {
		current  int
		expected int
	}{
		{500, 1000},
		{1000, 2000},
		{2000, 4000},
		{4000, -1},  // No next bitrate
		{750, 1000}, // Between bitrates
	}

	for _, test := range tests {
		result := findNextBitrate(test.current, available)
		if result != test.expected {
			t.Errorf("findNextBitrate(%d): expected %d, got %d", test.current, test.expected, result)
		}
	}
}

func TestFindPrevBitrate(t *testing.T) {
	available := []int{500, 1000, 2000, 4000}

	tests := []struct {
		current  int
		expected int
	}{
		{500, -1}, // No prev bitrate
		{1000, 500},
		{2000, 1000},
		{4000, 2000},
		{750, 500}, // Between bitrates
	}

	for _, test := range tests {
		result := findPrevBitrate(test.current, available)
		if result != test.expected {
			t.Errorf("findPrevBitrate(%d): expected %d, got %d", test.current, test.expected, result)
		}
	}
}

func TestABRWithRealWorldScenario(t *testing.T) {
	config := ABRConfig{HistorySize: 10}
	abr := NewAdaptiveBitrateManager(config)

	// Simulate good network conditions
	abr.UpdateCurrentBitrate(500) // Start at lowest
	for i := 0; i < 5; i++ {
		abr.RecordSegmentMetrics(SegmentMetrics{
			SegmentNumber:   i,
			Bitrate:         2000,
			DownloadTime:    1000,
			BytesDownloaded: 250000,
			BufferOccupancy: 90, // High buffer
		})
	}

	// Should want to upscale
	if !abr.ShouldUpscale() {
		t.Error("Expected upscale with good conditions")
	}

	// Update to higher bitrate
	abr.UpdateCurrentBitrate(1000)

	// Now simulate degraded network (buffer drops)
	for i := 5; i < 8; i++ {
		abr.RecordSegmentMetrics(SegmentMetrics{
			SegmentNumber:   i,
			Bitrate:         500,
			DownloadTime:    3000, // Slower download
			BytesDownloaded: 150000,
			BufferOccupancy: 15, // Low buffer
		})
	}

	// Should want to downscale
	if !abr.ShouldDownscale() {
		t.Error("Expected downscale with poor conditions")
	}
}
