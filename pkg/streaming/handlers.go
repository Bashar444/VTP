package streaming

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ABRHandlers handles HTTP requests for adaptive bitrate streaming
type ABRHandlers struct {
	manager *AdaptiveBitrateManager
	logger  *log.Logger
}

// NewABRHandlers creates a new ABR handlers instance
func NewABRHandlers(manager *AdaptiveBitrateManager, logger *log.Logger) *ABRHandlers {
	return &ABRHandlers{
		manager: manager,
		logger:  logger,
	}
}

// RegisterABRRoutes registers all ABR HTTP routes
func (h *ABRHandlers) RegisterABRRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/recordings/", func(w http.ResponseWriter, r *http.Request) {
		// Extract recording ID and route to appropriate handler
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/recordings/"), "/")
		if len(parts) < 3 {
			http.NotFound(w, r)
			return
		}

		recordingID := parts[0]

		// Route to ABR endpoints
		if len(parts) >= 3 && parts[1] == "abr" {
			switch parts[2] {
			case "quality":
				if r.Method == http.MethodPost {
					h.SelectQualityHandler(w, r, recordingID)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			case "stats":
				if r.Method == http.MethodGet {
					h.GetABRStatsHandler(w, r, recordingID)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			case "metrics":
				if r.Method == http.MethodPost {
					h.RecordMetricsHandler(w, r, recordingID)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			default:
				http.NotFound(w, r)
			}
		}
	})
}

// SelectQualityRequest represents a request to select video quality
type SelectQualityRequest struct {
	Bandwidth int `json:"bandwidth"` // Bandwidth in kbps
}

// SelectQualityResponse represents the quality selection response
type SelectQualityResponse struct {
	BitrateLevel    int    `json:"bitrate_level"`
	BitrateLabel    string `json:"bitrate_label"`
	RecommendedBits int    `json:"recommended_bitrate_kbps"`
	Resolution      string `json:"resolution"`
	FrameRate       int    `json:"frame_rate"`
	Message         string `json:"message"`
	Timestamp       int64  `json:"timestamp"`
}

// SelectQualityHandler handles POST /api/v1/recordings/{id}/abr/quality
func (h *ABRHandlers) SelectQualityHandler(w http.ResponseWriter, r *http.Request, recordingID string) {
	var req SelectQualityRequest

	// Parse request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		h.logger.Printf("SelectQuality - invalid request: %v", err)
		return
	}

	// Validate bandwidth
	if req.Bandwidth <= 0 {
		http.Error(w, `{"error":"Bandwidth must be positive"}`, http.StatusBadRequest)
		h.logger.Printf("SelectQuality - invalid bandwidth: %d", req.Bandwidth)
		return
	}

	// Select quality
	level := h.manager.SelectQuality(float64(req.Bandwidth))

	// Convert level to bitrate and label
	bitrate := h.bitrateFromLevel(level)
	label := h.labelFromLevel(level)
	resolution := h.resolutionFromLevel(level)
	frameRate := h.frameRateFromLevel(level)

	// Create response
	resp := SelectQualityResponse{
		BitrateLevel:    int(level),
		BitrateLabel:    label,
		RecommendedBits: bitrate,
		Resolution:      resolution,
		FrameRate:       frameRate,
		Message:         fmt.Sprintf("Selected %s (%d kbps) for bandwidth %d kbps", label, bitrate, req.Bandwidth),
		Timestamp:       int64(getUnixMillis()),
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	h.logger.Printf("SelectQuality - recording=%s bandwidth=%d kbps -> %s (%d kbps)", recordingID, req.Bandwidth, label, bitrate)
}

// SegmentMetricsRequest represents segment delivery metrics
type SegmentMetricsRequest struct {
	SegmentNumber   int `json:"segment_number"`
	RequestTime     int `json:"request_time_ms"`
	DownloadTime    int `json:"download_time_ms"`
	BytesDownloaded int `json:"bytes_downloaded"`
	BitrateSent     int `json:"bitrate_kbps"`
	BufferOccupancy int `json:"buffer_occupancy_percent"`
}

// RecordMetricsResponse represents the metrics recording response
type RecordMetricsResponse struct {
	Accepted        bool   `json:"accepted"`
	Message         string `json:"message"`
	CurrentBitrate  int    `json:"current_bitrate_kbps"`
	ShouldUpscale   bool   `json:"should_upscale"`
	ShouldDownscale bool   `json:"should_downscale"`
	Timestamp       int64  `json:"timestamp"`
}

// RecordMetricsHandler handles POST /api/v1/recordings/{id}/abr/metrics
func (h *ABRHandlers) RecordMetricsHandler(w http.ResponseWriter, r *http.Request, recordingID string) {
	var req SegmentMetricsRequest

	// Parse request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		h.logger.Printf("RecordMetrics - invalid request: %v", err)
		return
	}

	// Create segment metrics
	metrics := SegmentMetrics{
		SegmentNumber:   req.SegmentNumber,
		RequestTime:     time.Now(),
		DownloadTime:    req.DownloadTime,
		BytesDownloaded: req.BytesDownloaded,
		Bitrate:         req.BitrateSent,
		BufferOccupancy: float64(req.BufferOccupancy),
	}

	// Record metrics
	h.manager.RecordSegmentMetrics(metrics)

	// Check if we should adapt quality
	shouldUpscale := h.manager.ShouldUpscale()
	shouldDownscale := h.manager.ShouldDownscale()

	// Create response
	resp := RecordMetricsResponse{
		Accepted:        true,
		Message:         "Metrics recorded successfully",
		CurrentBitrate:  h.manager.GetCurrentBitrate(),
		ShouldUpscale:   shouldUpscale,
		ShouldDownscale: shouldDownscale,
		Timestamp:       int64(getUnixMillis()),
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	h.logger.Printf("RecordMetrics - recording=%s segment=%d download_time=%dms buffer=%d%% -> upscale=%v downscale=%v",
		recordingID, req.SegmentNumber, req.DownloadTime, req.BufferOccupancy, shouldUpscale, shouldDownscale)
}

// ABRStatsResponse represents comprehensive ABR statistics
type ABRStatsResponse struct {
	CurrentBitrate      int                    `json:"current_bitrate_kbps"`
	AvailableBitrates   []int                  `json:"available_bitrates_kbps"`
	OptimalBitrate      int                    `json:"optimal_bitrate_kbps"`
	RecentSegments      int                    `json:"recent_segments_count"`
	AverageDownloadTime int64                  `json:"average_download_time_ms"`
	AverageBufferHealth float64                `json:"average_buffer_health_percent"`
	NetworkStats        map[string]interface{} `json:"network_stats"`
	Timestamp           int64                  `json:"timestamp"`
}

// GetABRStatsHandler handles GET /api/v1/recordings/{id}/abr/stats
func (h *ABRHandlers) GetABRStatsHandler(w http.ResponseWriter, r *http.Request, recordingID string) {
	stats := h.manager.GetStatistics()

	// Extract detailed statistics
	avgDownloadTime := int64(0)
	avgBufferHealth := 0.0
	if recentCount, ok := stats["recent_segments"].(int); ok && recentCount > 0 {
		if avgDL, ok := stats["avg_download_time"].(float64); ok {
			avgDownloadTime = int64(avgDL)
		}
		if avgBuf, ok := stats["avg_buffer_health"].(float64); ok {
			avgBufferHealth = avgBuf
		}
	}

	// Create response
	resp := ABRStatsResponse{
		CurrentBitrate:      h.manager.GetCurrentBitrate(),
		AvailableBitrates:   h.manager.GetAvailableBitrates(),
		OptimalBitrate:      h.manager.PredictOptimalBitrate(),
		RecentSegments:      stats["recent_segments"].(int),
		AverageDownloadTime: avgDownloadTime,
		AverageBufferHealth: avgBufferHealth,
		NetworkStats:        stats,
		Timestamp:           int64(getUnixMillis()),
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	h.logger.Printf("GetABRStats - recording=%s current=%d optimal=%d segments=%d",
		recordingID, resp.CurrentBitrate, resp.OptimalBitrate, resp.RecentSegments)
}

// Helper functions to convert bitrate level to properties

func (h *ABRHandlers) bitrateFromLevel(level BitrateLevel) int {
	switch level {
	case BitrateVeryLow:
		return 500
	case BitrateLow:
		return 1000
	case BitrateMedium:
		return 2000
	case BitrateHigh:
		return 4000
	default:
		return 500
	}
}

func (h *ABRHandlers) labelFromLevel(level BitrateLevel) string {
	switch level {
	case BitrateVeryLow:
		return "VeryLow"
	case BitrateLow:
		return "Low"
	case BitrateMedium:
		return "Medium"
	case BitrateHigh:
		return "High"
	default:
		return "Unknown"
	}
}

func (h *ABRHandlers) resolutionFromLevel(level BitrateLevel) string {
	switch level {
	case BitrateVeryLow:
		return "1280x720"
	case BitrateLow:
		return "1280x720"
	case BitrateMedium:
		return "1920x1080"
	case BitrateHigh:
		return "1920x1080"
	default:
		return "1280x720"
	}
}

func (h *ABRHandlers) frameRateFromLevel(level BitrateLevel) int {
	switch level {
	case BitrateVeryLow:
		return 24
	case BitrateLow:
		return 24
	case BitrateMedium:
		return 30
	case BitrateHigh:
		return 30
	default:
		return 24
	}
}

// Helper to get current unix time in milliseconds
func getUnixMillis() int64 {
	return time.Now().UnixNano() / 1e6
}

// ParseRecordingID extracts recording ID from URL path
func ParseRecordingID(path string) (string, error) {
	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/recordings/"), "/")
	if len(parts) < 1 || parts[0] == "" {
		return "", fmt.Errorf("invalid recording ID in path")
	}

	// Parse as integer to validate it's a valid ID
	if _, err := strconv.ParseInt(parts[0], 10, 64); err != nil {
		return "", fmt.Errorf("invalid recording ID format")
	}

	return parts[0], nil
}
