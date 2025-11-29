package streaming

import (
	"time"
)

// BitrateLevel represents streaming quality level
type BitrateLevel int

const (
	BitrateVeryLow BitrateLevel = iota // 500 kbps
	BitrateLow                         // 1000 kbps
	BitrateMedium                      // 2000 kbps
	BitrateHigh                        // 4000 kbps
)

// StreamQuality represents available stream quality
type StreamQuality struct {
	Bitrate    int
	Resolution string
	FrameRate  int
	Label      string
}

// NetworkStats tracks network condition metrics
type NetworkStats struct {
	Bandwidth    float64 // kbps
	Latency      int     // milliseconds
	PacketLoss   float64 // percentage
	BufferHealth float64 // 0-100
	Timestamp    time.Time
}

// SegmentMetrics tracks segment delivery metrics
type SegmentMetrics struct {
	SegmentNumber   int
	RequestTime     time.Time
	DownloadTime    int // milliseconds
	BytesDownloaded int
	Bitrate         int
	BufferOccupancy float64
}

// ABRConfig holds ABR configuration
type ABRConfig struct {
	MinBitrate    int
	MaxBitrate    int
	ThresholdUp   float64 // bandwidth factor to upscale
	ThresholdDown float64 // bandwidth factor to downscale
	HistorySize   int     // number of segments to track
}

// AdaptiveBitrateManager manages adaptive bitrate streaming
type AdaptiveBitrateManager struct {
	config            ABRConfig
	currentBitrate    int
	segmentHistory    []SegmentMetrics
	networkHistory    []NetworkStats
	availableBitrates []int
}

// BitrateProfile represents an encoding profile
type BitrateProfile struct {
	Bitrate    int
	Resolution string
	FrameRate  int
	Label      string
}

// StreamingQuality represents quality level details
type StreamingQuality struct {
	Level      BitrateLevel
	Bitrate    int
	Resolution string
	FrameRate  int
	Label      string
}
