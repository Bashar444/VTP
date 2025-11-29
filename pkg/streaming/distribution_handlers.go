package streaming

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// DistributionHandlers handles HTTP requests for live distribution
type DistributionHandlers struct {
	service *DistributionService
}

// Request/Response Types

type StartLiveStreamRequest struct {
	RecordingID string `json:"recording_id"`
	MaxViewers  int    `json:"max_viewers"`
}

type StartLiveStreamResponse struct {
	RecordingID string   `json:"recording_id"`
	StreamID    string   `json:"stream_id"`
	Status      string   `json:"status"`
	MaxViewers  int      `json:"max_viewers"`
	Profiles    []string `json:"profiles"`
	Message     string   `json:"message"`
	Timestamp   string   `json:"timestamp"`
}

type JoinStreamRequest struct {
	ViewerID       string `json:"viewer_id"`
	InitialBitrate string `json:"initial_bitrate"`
}

type JoinStreamResponse struct {
	RecordingID    string  `json:"recording_id"`
	ViewerID       string  `json:"viewer_id"`
	SessionID      string  `json:"session_id"`
	CurrentBitrate string  `json:"current_bitrate"`
	BufferHealth   float64 `json:"buffer_health"`
	Message        string  `json:"message"`
	Timestamp      string  `json:"timestamp"`
}

type DeliverSegmentRequest struct {
	SegmentID string `json:"segment_id"`
	ViewerID  string `json:"viewer_id"`
}

type DeliverSegmentResponse struct {
	RecordingID string `json:"recording_id"`
	SegmentID   string `json:"segment_id"`
	ViewerID    string `json:"viewer_id"`
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	Timestamp   string `json:"timestamp"`
}

type AdaptQualityRequest struct {
	ViewerID     string  `json:"viewer_id"`
	BufferHealth float64 `json:"buffer_health"`
}

type AdaptQualityResponse struct {
	RecordingID  string  `json:"recording_id"`
	ViewerID     string  `json:"viewer_id"`
	OldBitrate   string  `json:"old_bitrate"`
	NewBitrate   string  `json:"new_bitrate"`
	BufferHealth float64 `json:"buffer_health"`
	Message      string  `json:"message"`
	Timestamp    string  `json:"timestamp"`
}

type StreamStatisticsResponse struct {
	RecordingID         string       `json:"recording_id"`
	ActiveViewers       int          `json:"active_viewers"`
	PeakViewers         int          `json:"peak_viewers"`
	TotalSegmentsServed int64        `json:"total_segments_served"`
	TotalBytesServed    int64        `json:"total_bytes_served"`
	Viewers             []ViewerInfo `json:"viewers"`
	Message             string       `json:"message"`
	Timestamp           string       `json:"timestamp"`
}

type ViewerInfo struct {
	ViewerID          string  `json:"viewer_id"`
	CurrentBitrate    string  `json:"current_bitrate"`
	SegmentsReceived  int64   `json:"segments_received"`
	BytesReceived     int64   `json:"bytes_received"`
	BufferHealth      float64 `json:"buffer_health"`
	ConnectionQuality string  `json:"connection_quality"`
	Duration          string  `json:"duration"`
}

type EndStreamRequest struct {
	RecordingID string `json:"recording_id"`
}

type EndStreamResponse struct {
	RecordingID string `json:"recording_id"`
	Status      string `json:"status"`
	Message     string `json:"message"`
	Timestamp   string `json:"timestamp"`
}

type DistributionMetricsResponse struct {
	TotalDistributors     int     `json:"total_distributors"`
	TotalActiveViewers    int     `json:"total_active_viewers"`
	TotalSegmentsServed   int64   `json:"total_segments_served"`
	TotalBytesServed      int64   `json:"total_bytes_served"`
	TotalDeliveryFailures int64   `json:"total_delivery_failures"`
	CDNCacheHitRate       float64 `json:"cdn_cache_hit_rate"`
	Message               string  `json:"message"`
	Timestamp             string  `json:"timestamp"`
}

// NewDistributionHandlers creates new distribution handlers
func NewDistributionHandlers(service *DistributionService) *DistributionHandlers {
	return &DistributionHandlers{
		service: service,
	}
}

// RegisterDistributionRoutes registers all distribution routes
func (dh *DistributionHandlers) RegisterDistributionRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/streams/start", dh.StartStreamHandler)
	mux.HandleFunc("/api/v1/streams/", dh.StreamOperationHandler) // POST for join/leave, GET for stats
	mux.HandleFunc("/api/v1/segments/deliver", dh.DeliverSegmentHandler)
	mux.HandleFunc("/api/v1/viewers/adapt-quality", dh.AdaptQualityHandler)
	mux.HandleFunc("/api/v1/distribution/metrics", dh.GetMetricsHandler)
	mux.HandleFunc("/api/v1/distribution/health", dh.HealthCheckHandler)
}

// StartStreamHandler starts a new live stream
func (dh *DistributionHandlers) StartStreamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StartLiveStreamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.RecordingID == "" {
		http.Error(w, "recording_id is required", http.StatusBadRequest)
		return
	}

	if req.MaxViewers < 1 {
		req.MaxViewers = 100
	}

	_, err := dh.service.StartLiveStream(req.RecordingID, req.MaxViewers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	profiles := []string{"VeryLow", "Low", "Medium", "High"}

	response := StartLiveStreamResponse{
		RecordingID: req.RecordingID,
		StreamID:    fmt.Sprintf("stream-%s", req.RecordingID),
		Status:      "started",
		MaxViewers:  req.MaxViewers,
		Profiles:    profiles,
		Message:     "Live stream started successfully",
		Timestamp:   timeString(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StreamOperationHandler handles join/leave and statistics operations
func (dh *DistributionHandlers) StreamOperationHandler(w http.ResponseWriter, r *http.Request) {
	recordingID := extractRecordingID(r.URL.Path, "/api/v1/streams/")

	if recordingID == "" {
		http.Error(w, "recording_id required in path", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodPost {
		// Join viewer operation
		var req JoinStreamRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.ViewerID == "" {
			http.Error(w, "viewer_id is required", http.StatusBadRequest)
			return
		}

		if req.InitialBitrate == "" {
			req.InitialBitrate = "Low"
		}

		session, err := dh.service.JoinStream(recordingID, req.ViewerID, req.InitialBitrate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		response := JoinStreamResponse{
			RecordingID:    recordingID,
			ViewerID:       req.ViewerID,
			SessionID:      session.SessionID,
			CurrentBitrate: req.InitialBitrate,
			BufferHealth:   session.BufferHealth,
			Message:        "Viewer joined successfully",
			Timestamp:      timeString(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	} else if r.Method == http.MethodDelete {
		// Leave viewer operation
		viewerID := r.URL.Query().Get("viewer_id")
		if viewerID == "" {
			http.Error(w, "viewer_id required in query", http.StatusBadRequest)
			return
		}

		err := dh.service.LeaveStream(recordingID, viewerID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":   "Viewer left successfully",
			"timestamp": timeString(),
		})

	} else if r.Method == http.MethodGet {
		// Get stream statistics
		stats, err := dh.service.GetStreamStatistics(recordingID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		viewers, _ := dh.service.GetStreamViewers(recordingID)
		viewerInfos := make([]ViewerInfo, 0)

		for _, viewer := range viewers {
			duration := time.Since(viewer.StartTime).String()
			viewerInfos = append(viewerInfos, ViewerInfo{
				ViewerID:          viewer.ViewerID,
				CurrentBitrate:    viewer.CurrentBitrate,
				SegmentsReceived:  viewer.SegmentsReceived,
				BytesReceived:     viewer.BytesReceived,
				BufferHealth:      viewer.BufferHealth,
				ConnectionQuality: viewer.ConnectionQuality,
				Duration:          duration,
			})
		}

		response := StreamStatisticsResponse{
			RecordingID:         recordingID,
			ActiveViewers:       stats.ActiveViewers,
			PeakViewers:         stats.PeakViewers,
			TotalSegmentsServed: stats.TotalSegmentsServed,
			TotalBytesServed:    stats.TotalBytesServed,
			Viewers:             viewerInfos,
			Message:             "Statistics retrieved successfully",
			Timestamp:           timeString(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// DeliverSegmentHandler delivers a segment to a viewer
func (dh *DistributionHandlers) DeliverSegmentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DeliverSegmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	recordingID := r.URL.Query().Get("recording_id")
	if recordingID == "" {
		http.Error(w, "recording_id required in query", http.StatusBadRequest)
		return
	}

	if req.ViewerID == "" || req.SegmentID == "" {
		http.Error(w, "viewer_id and segment_id are required", http.StatusBadRequest)
		return
	}

	err := dh.service.DeliverSegmentToViewer(recordingID, req.ViewerID, req.SegmentID)

	response := DeliverSegmentResponse{
		RecordingID: recordingID,
		SegmentID:   req.SegmentID,
		ViewerID:    req.ViewerID,
		Success:     err == nil,
		Message:     "Segment delivered successfully",
		Timestamp:   timeString(),
	}

	if err != nil {
		response.Success = false
		response.Message = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AdaptQualityHandler adapts viewer quality based on buffer health
func (dh *DistributionHandlers) AdaptQualityHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AdaptQualityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	recordingID := r.URL.Query().Get("recording_id")
	if recordingID == "" {
		http.Error(w, "recording_id required in query", http.StatusBadRequest)
		return
	}

	if req.ViewerID == "" {
		http.Error(w, "viewer_id is required", http.StatusBadRequest)
		return
	}

	// Get old bitrate
	distributor, _ := dh.service.GetDistributor(recordingID)
	session, _ := distributor.GetViewerSession(req.ViewerID)
	oldBitrate := session.CurrentBitrate

	// Adapt quality
	newBitrate, err := dh.service.AdaptViewerQuality(recordingID, req.ViewerID, req.BufferHealth)

	response := AdaptQualityResponse{
		RecordingID:  recordingID,
		ViewerID:     req.ViewerID,
		OldBitrate:   oldBitrate,
		NewBitrate:   newBitrate,
		BufferHealth: req.BufferHealth,
		Message:      "Quality adapted successfully",
		Timestamp:    timeString(),
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMetricsHandler returns distribution system metrics
func (dh *DistributionHandlers) GetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics := dh.service.GetMetrics()

	response := DistributionMetricsResponse{
		TotalDistributors:     metrics.TotalDistributors,
		TotalActiveViewers:    metrics.TotalActiveViewers,
		TotalSegmentsServed:   metrics.TotalSegmentsServed,
		TotalBytesServed:      metrics.TotalBytesServed,
		TotalDeliveryFailures: metrics.TotalDeliveryFailures,
		CDNCacheHitRate:       metrics.CDNCacheHitRate,
		Message:               "Metrics retrieved successfully",
		Timestamp:             timeString(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HealthCheckHandler checks distribution service health
func (dh *DistributionHandlers) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics := dh.service.GetMetrics()
	status := "healthy"

	if metrics.TotalDeliveryFailures > 100 {
		status = "degraded"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         status,
		"distributors":   metrics.TotalDistributors,
		"active_viewers": metrics.TotalActiveViewers,
		"failure_rate":   float64(metrics.TotalDeliveryFailures) / float64(metrics.TotalSegmentsServed+1),
		"timestamp":      timeString(),
	})
}

// Helper functions

func extractRecordingID(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	remainder := strings.TrimPrefix(path, prefix)
	parts := strings.Split(remainder, "/")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return ""
}

func timeString() string {
	return time.Now().Format("2006-01-02T15:04:05Z")
}
