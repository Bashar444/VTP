package g5

import "time"

// NetworkType represents the type of network connection
type NetworkType string

const (
	Network5G      NetworkType = "5G"
	Network4G      NetworkType = "4G"
	NetworkWiFi    NetworkType = "WiFi"
	NetworkLTE     NetworkType = "LTE"
	NetworkUnknown NetworkType = "Unknown"
)

// Network5GStatus represents the current 5G network status
type Network5GStatus struct {
	Type           NetworkType `json:"type"`
	Latency        int         `json:"latency"`        // milliseconds
	Bandwidth      int         `json:"bandwidth"`      // Mbps
	SignalStrength int         `json:"signalStrength"` // dBm
	Connected      bool        `json:"connected"`
	Timestamp      time.Time   `json:"timestamp"`
	RSRP           int         `json:"rsrp,omitempty"` // Reference Signal Received Power
	RSRQ           int         `json:"rsrq,omitempty"` // Reference Signal Received Quality
	SINR           int         `json:"sinr,omitempty"` // Signal-to-Interference-plus-Noise Ratio
}

// EdgeNode represents a 5G edge computing node
type EdgeNode struct {
	ID          string     `json:"id"`
	Region      string     `json:"region"`
	Country     string     `json:"country"`
	Endpoint    string     `json:"endpoint"`
	Latency     int        `json:"latency"`  // milliseconds
	Capacity    int        `json:"capacity"` // connections
	Load        float64    `json:"load"`     // percentage 0-100
	Status      NodeStatus `json:"status"`
	LastChecked time.Time  `json:"lastChecked"`
	Distance    int        `json:"distance,omitempty"` // km from user
	Available   int        `json:"available"`          // available capacity
}

// NodeStatus represents the status of an edge node
type NodeStatus string

const (
	NodeOnline      NodeStatus = "online"
	NodeOffline     NodeStatus = "offline"
	NodeDegraded    NodeStatus = "degraded"
	NodeMaintenance NodeStatus = "maintenance"
)

// QualityProfile represents a streaming quality preset
type QualityProfile struct {
	Name         string `json:"name"`
	Bitrate      int    `json:"bitrate"`    // Kbps
	Resolution   string `json:"resolution"` // e.g., "4K", "1440p", "1080p", "720p"
	FPS          int    `json:"fps"`
	Codec        string `json:"codec"`        // e.g., "VP9", "H.265", "H.264"
	MaxLatency   int    `json:"maxLatency"`   // milliseconds
	MinBandwidth int    `json:"minBandwidth"` // Kbps
}

// QualityLevel represents adaptive streaming quality levels
type QualityLevel string

const (
	QualityUltraHD  QualityLevel = "ultra_hd" // 4K
	QualityHighDef  QualityLevel = "high_def" // 1440p
	QualityStandard QualityLevel = "standard" // 1080p
	QualityMedium   QualityLevel = "medium"   // 720p
	QualityLow      QualityLevel = "low"      // 480p
	QualityAuto     QualityLevel = "auto"     // Automatic selection
)

// NetworkMetrics represents real-time network performance metrics
type NetworkMetrics struct {
	SessionID           string      `json:"sessionId"`
	Timestamp           time.Time   `json:"timestamp"`
	NetworkType         NetworkType `json:"networkType"`
	Latency             int         `json:"latency"`             // ms
	AvgLatency          int         `json:"avgLatency"`          // ms (rolling average)
	MaxLatency          int         `json:"maxLatency"`          // ms
	MinLatency          int         `json:"minLatency"`          // ms
	Bandwidth           int         `json:"bandwidth"`           // Mbps
	AvailableBandwidth  int         `json:"availableBandwidth"`  // Mbps
	PacketLoss          float64     `json:"packetLoss"`          // percentage
	Jitter              int         `json:"jitter"`              // ms
	BufferHealth        float64     `json:"bufferHealth"`        // percentage
	ConnectionStability float64     `json:"connectionStability"` // percentage
}

// EdgeMetrics represents edge node performance metrics
type EdgeMetrics struct {
	NodeID      string    `json:"nodeId"`
	Timestamp   time.Time `json:"timestamp"`
	Latency     int       `json:"latency"`     // ms
	Load        float64   `json:"load"`        // percentage
	CPUUsage    float64   `json:"cpuUsage"`    // percentage
	MemoryUsage float64   `json:"memoryUsage"` // percentage
	Connections int       `json:"connections"`
	Throughput  int       `json:"throughput"` // Mbps
}

// AdaptiveStrategy defines quality adaptation strategy
type AdaptiveStrategy struct {
	Enabled         bool    `json:"enabled"`
	TargetLatency   int     `json:"targetLatency"`   // ms
	TargetBandwidth int     `json:"targetBandwidth"` // Mbps
	MinBitrate      int     `json:"minBitrate"`      // Kbps
	MaxBitrate      int     `json:"maxBitrate"`      // Kbps
	SwitchThreshold float64 `json:"switchThreshold"` // percentage change to trigger switch
	CheckInterval   int     `json:"checkInterval"`   // milliseconds
}

// 5GConfig represents 5G module configuration
type Config struct {
	Enabled            bool             `json:"enabled"`
	DetectionInterval  int              `json:"detectionInterval"` // milliseconds
	MetricsInterval    int              `json:"metricsInterval"`   // milliseconds
	EdgeCheckInterval  int              `json:"edgeCheckInterval"` // milliseconds
	AdaptiveQuality    AdaptiveStrategy `json:"adaptiveQuality"`
	PreferredEdgeNode  string           `json:"preferredEdgeNode,omitempty"`
	AllowNetworkSwitch bool             `json:"allowNetworkSwitch"`
	MaxLatencyTarget   int              `json:"maxLatencyTarget"`   // ms
	MinBandwidthTarget int              `json:"minBandwidthTarget"` // Kbps
}

// DetectionResult represents the result of network detection
type DetectionResult struct {
	Detected       bool        `json:"detected"`
	NetworkType    NetworkType `json:"networkType"`
	Latency        int         `json:"latency"`
	Bandwidth      int         `json:"bandwidth"`
	SignalStrength int         `json:"signalStrength"`
	Timestamp      time.Time   `json:"timestamp"`
	Error          string      `json:"error,omitempty"`
}

// QualityAdjustment represents a quality profile change
type QualityAdjustment struct {
	FromProfile QualityLevel `json:"fromProfile"`
	ToProfile   QualityLevel `json:"toProfile"`
	Reason      string       `json:"reason"`
	Timestamp   time.Time    `json:"timestamp"`
	Duration    int          `json:"duration"` // milliseconds
}

// HealthCheck represents edge node health status
type HealthCheck struct {
	NodeID    string     `json:"nodeId"`
	Status    NodeStatus `json:"status"`
	Latency   int        `json:"latency"` // ms
	LastCheck time.Time  `json:"lastCheck"`
	Errors    int        `json:"errors"`
	Success   bool       `json:"success"`
	Message   string     `json:"message,omitempty"`
}

// Statistics represents aggregated network statistics
type Statistics struct {
	Period              time.Duration `json:"period"`
	TotalSessions       int           `json:"totalSessions"`
	AvgLatency          int           `json:"avgLatency"`
	MaxLatency          int           `json:"maxLatency"`
	MinLatency          int           `json:"minLatency"`
	AvgBandwidth        int           `json:"avgBandwidth"`
	TotalPacketsLost    int           `json:"totalPacketsLost"`
	AveragePacketLoss   float64       `json:"averagePacketLoss"`
	BufferUnderflows    int           `json:"bufferUnderflows"`
	QualitySwitches     int           `json:"qualitySwitches"`
	EdgeNodeFailovers   int           `json:"edgeNodeFailovers"`
	ConnectionStability float64       `json:"connectionStability"` // percentage
	AvgPlaybackTime     int           `json:"avgPlaybackTime"`     // seconds
}

// RequestOptions represents options for 5G requests
type RequestOptions struct {
	Timeout    time.Duration
	RetryCount int
	RetryDelay time.Duration
	UseEdge    bool
	EdgeNodeID string
}

// ResponseError represents a 5G API error response
type ResponseError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}
