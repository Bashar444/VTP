package streaming

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// TranscodingHandlers handles HTTP requests for transcoding
type TranscodingHandlers struct {
	service *TranscodingService
	logger  *log.Logger
}

// NewTranscodingHandlers creates a new transcoding handlers instance
func NewTranscodingHandlers(service *TranscodingService, logger *log.Logger) *TranscodingHandlers {
	return &TranscodingHandlers{
		service: service,
		logger:  logger,
	}
}

// RegisterTranscodingRoutes registers all transcoding HTTP routes
func (h *TranscodingHandlers) RegisterTranscodingRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/recordings/", func(w http.ResponseWriter, r *http.Request) {
		// Extract recording ID and route to appropriate handler
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/recordings/"), "/")
		if len(parts) < 3 {
			http.NotFound(w, r)
			return
		}

		recordingID := parts[0]

		// Route to transcoding endpoints
		if len(parts) >= 3 && parts[1] == "transcode" {
			switch parts[2] {
			case "quality":
				if r.Method == http.MethodPost {
					h.StartTranscodingHandler(w, r, recordingID)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			case "progress":
				if r.Method == http.MethodGet {
					h.GetTranscodingProgressHandler(w, r, recordingID)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			case "cancel":
				if r.Method == http.MethodPost {
					h.CancelTranscodingHandler(w, r, recordingID)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			default:
				http.NotFound(w, r)
			}
		} else if len(parts) >= 3 && parts[1] == "stream" && parts[2] == "master.m3u8" {
			// Master playlist endpoint
			if r.Method == http.MethodGet {
				h.GetMasterPlaylistHandler(w, r, recordingID)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}
	})
}

// StartTranscodingRequest represents a request to start transcoding
type StartTranscodingRequest struct {
	InputPath string `json:"input_path"` // Path to source video
}

// StartTranscodingResponse represents the response from starting transcoding
type StartTranscodingResponse struct {
	RecordingID string   `json:"recording_id"`
	JobIDs      []string `json:"job_ids"`
	Profiles    []string `json:"profiles"`
	Message     string   `json:"message"`
	Timestamp   int64    `json:"timestamp"`
}

// StartTranscodingHandler handles POST /api/v1/recordings/{id}/transcode/quality
func (h *TranscodingHandlers) StartTranscodingHandler(w http.ResponseWriter, r *http.Request, recordingID string) {
	var req StartTranscodingRequest

	// Parse request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		h.logger.Printf("StartTranscoding - invalid request: %v", err)
		return
	}

	// Validate input
	if req.InputPath == "" {
		http.Error(w, `{"error":"input_path is required"}`, http.StatusBadRequest)
		return
	}

	// Start transcoding
	jobIDs, err := h.service.StartMultiBitrateEncoding(recordingID, req.InputPath)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		h.logger.Printf("StartTranscoding - error: %v", err)
		return
	}

	// Build profile names
	profiles := []string{"VeryLow (500 kbps)", "Low (1000 kbps)", "Medium (2000 kbps)", "High (4000 kbps)"}

	// Create response
	resp := StartTranscodingResponse{
		RecordingID: recordingID,
		JobIDs:      jobIDs,
		Profiles:    profiles,
		Message:     fmt.Sprintf("Queued %d transcoding jobs for multi-bitrate encoding", len(jobIDs)),
		Timestamp:   getUnixMillis(),
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	h.logger.Printf("StartTranscoding - recording=%s queued %d jobs", recordingID, len(jobIDs))
}

// TranscodingProgressResponse represents the progress response
type TranscodingProgressResponse struct {
	RecordingID     string                 `json:"recording_id"`
	TotalJobs       int                    `json:"total_jobs"`
	CompletedJobs   int                    `json:"completed_jobs"`
	FailedJobs      int                    `json:"failed_jobs"`
	AverageProgress float64                `json:"average_progress_percent"`
	IsComplete      bool                   `json:"is_complete"`
	Jobs            map[string]interface{} `json:"jobs"`
	QueueStats      map[string]interface{} `json:"queue_stats"`
	Timestamp       int64                  `json:"timestamp"`
}

// GetTranscodingProgressHandler handles GET /api/v1/recordings/{id}/transcode/progress
func (h *TranscodingHandlers) GetTranscodingProgressHandler(w http.ResponseWriter, r *http.Request, recordingID string) {
	// Get transcoding status
	status := h.service.GetRecordingTranscodingStatus(recordingID)

	// Get queue statistics
	queueStats := h.service.GetQueueStats()

	// Build response
	resp := TranscodingProgressResponse{
		RecordingID:     recordingID,
		TotalJobs:       status["total_jobs"].(int),
		CompletedJobs:   status["completed_count"].(int),
		FailedJobs:      status["failed_count"].(int),
		AverageProgress: status["average_progress"].(float64),
		IsComplete:      status["is_complete"].(bool),
		Jobs:            status["jobs"].(map[string]interface{}),
		QueueStats:      queueStats,
		Timestamp:       getUnixMillis(),
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	h.logger.Printf("GetTranscodingProgress - recording=%s completed=%d/%d",
		recordingID, resp.CompletedJobs, resp.TotalJobs)
}

// CancelTranscodingHandler handles POST /api/v1/recordings/{id}/transcode/cancel
func (h *TranscodingHandlers) CancelTranscodingHandler(w http.ResponseWriter, r *http.Request, recordingID string) {
	// Cancel all encoding jobs for this recording
	err := h.service.CancelRecordingEncoding(recordingID)

	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		h.logger.Printf("CancelTranscoding - error: %v", err)
		return
	}

	// Create response
	resp := map[string]interface{}{
		"recording_id": recordingID,
		"message":      "All transcoding jobs cancelled",
		"timestamp":    getUnixMillis(),
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	h.logger.Printf("CancelTranscoding - recording=%s cancelled", recordingID)
}

// GetMasterPlaylistHandler handles GET /api/v1/recordings/{id}/stream/master.m3u8
func (h *TranscodingHandlers) GetMasterPlaylistHandler(w http.ResponseWriter, r *http.Request, recordingID string) {
	// Check if recording is finished transcoding
	if !h.service.transcoder.IsRecordingCompleted(recordingID) {
		http.Error(w, `{"error":"Recording is still transcoding"}`, http.StatusBadRequest)
		h.logger.Printf("GetMasterPlaylist - recording %s not yet completed", recordingID)
		return
	}

	// Generate master playlist
	playlist, err := h.service.transcoder.GenerateMasterPlaylist(recordingID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		h.logger.Printf("GetMasterPlaylist - error: %v", err)
		return
	}

	// Send playlist
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	fmt.Fprint(w, playlist)

	h.logger.Printf("GetMasterPlaylist - recording=%s returned", recordingID)
}

// GetVariantPlaylistHandler handles GET /api/v1/recordings/{id}/stream/{bitrate}.m3u8
func (h *TranscodingHandlers) GetVariantPlaylistHandler(w http.ResponseWriter, r *http.Request, recordingID, bitrateStr string) {
	// Parse bitrate
	bitrate, err := strconv.Atoi(bitrateStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid bitrate"}`, http.StatusBadRequest)
		return
	}

	// Generate variant playlist
	playlist, err := h.service.transcoder.GenerateVariantPlaylist(recordingID, bitrate)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		h.logger.Printf("GetVariantPlaylist - error: %v", err)
		return
	}

	// Send playlist
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	fmt.Fprint(w, playlist)

	h.logger.Printf("GetVariantPlaylist - recording=%s bitrate=%d returned", recordingID, bitrate)
}

// GeneratePlaylistsHandler handles POST /api/v1/recordings/{id}/stream/generate-playlists
func (h *TranscodingHandlers) GeneratePlaylistsHandler(w http.ResponseWriter, r *http.Request, recordingID string) {
	// Generate all playlists
	playlists, err := h.service.GeneratePlaylistsForRecording(recordingID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		h.logger.Printf("GeneratePlaylists - error: %v", err)
		return
	}

	// Create response
	resp := map[string]interface{}{
		"recording_id": recordingID,
		"playlists":    playlists,
		"message":      "Generated master and variant playlists",
		"timestamp":    getUnixMillis(),
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	h.logger.Printf("GeneratePlaylists - recording=%s playlists generated", recordingID)
}
