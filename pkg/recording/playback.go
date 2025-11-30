package recording

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// PlaybackHandlers manages playback-related HTTP endpoints
type PlaybackHandlers struct {
	streamingManager *StreamingManager
	service          *RecordingService
	logger           *log.Logger
}

// NewPlaybackHandlers creates new playback handler
func NewPlaybackHandlers(streamingManager *StreamingManager, service *RecordingService, logger *log.Logger) *PlaybackHandlers {
	return &PlaybackHandlers{
		streamingManager: streamingManager,
		service:          service,
		logger:           logger,
	}
}

// StreamHLSPlaylistHandler serves HLS master playlist
func (h *PlaybackHandlers) StreamHLSPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	recordingIDStr := strings.TrimPrefix(r.URL.Path, "/api/v1/recordings/")
	recordingIDStr = strings.TrimSuffix(recordingIDStr, "/stream/playlist.m3u8")

	recordingID, err := uuid.Parse(recordingIDStr)
	if err != nil {
		http.Error(w, "Invalid recording ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Verify recording exists and is ready for streaming
	recording, err := h.service.GetRecording(ctx, recordingID)
	if err != nil {
		http.Error(w, "Recording not found", http.StatusNotFound)
		return
	}

	if recording.Status != StatusCompleted {
		http.Error(w, "Recording not yet finished", http.StatusBadRequest)
		return
	}

	// Log playback start
	h.streamingManager.LogPlaybackEvent(ctx, recordingID, uuid.Nil, "playback_start", nil)

	// Generate master playlist
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Write master playlist (simplified with single bitrate for PoC)
	playlist := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:10.0,
segment-000.ts
#EXTINF:10.0,
segment-001.ts
#EXT-X-ENDLIST
`

	fmt.Fprint(w, playlist)
	h.logger.Printf("HLS playlist requested for recording: %s", recordingID)
}

// StreamHLSSegmentHandler serves individual HLS segments
func (h *PlaybackHandlers) StreamHLSSegmentHandler(w http.ResponseWriter, r *http.Request) {
	recordingIDStr := strings.TrimPrefix(r.URL.Path, "/api/v1/recordings/")
	parts := strings.Split(recordingIDStr, "/")

	if len(parts) < 1 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	recordingID, err := uuid.Parse(parts[0])
	if err != nil {
		http.Error(w, "Invalid recording ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Verify recording exists
	if _, err := h.service.GetRecording(ctx, recordingID); err != nil {
		http.Error(w, "Recording not found", http.StatusNotFound)
		return
	}

	// Extract segment name from URL
	segmentPath := filepath.Join(h.streamingManager.outputPath, recordingID.String(), filepath.Base(r.URL.Path))

	// Validate path is within recording directory (security)
	absPath, err := filepath.Abs(segmentPath)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	recordingDir := filepath.Join(h.streamingManager.outputPath, recordingID.String())
	if !strings.HasPrefix(absPath, recordingDir) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	file, err := os.Open(absPath)
	if err != nil {
		http.Error(w, "Segment not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "Could not stat file", http.StatusInternalServerError)
		return
	}

	// Set proper MIME type for MPEG-TS segments
	w.Header().Set("Content-Type", "video/mp2t")
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	w.Header().Set("Cache-Control", "max-age=3600") // Cache segments for 1 hour

	// Support range requests for seeking
	http.ServeContent(w, r, filepath.Base(absPath), fileInfo.ModTime(), file)

	h.logger.Printf("Segment streamed for recording: %s, segment: %s", recordingID, filepath.Base(r.URL.Path))
}

// GetRecordingInfoHandler returns recording metadata and streaming info
func (h *PlaybackHandlers) GetRecordingInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	recordingIDStr := strings.TrimPrefix(r.URL.Path, "/api/v1/recordings/")
	recordingIDStr = strings.TrimSuffix(recordingIDStr, "/info")

	recordingID, err := uuid.Parse(recordingIDStr)
	if err != nil {
		http.Error(w, "Invalid recording ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	recording, err := h.service.GetRecording(ctx, recordingID)
	if err != nil {
		http.Error(w, "Recording not found", http.StatusNotFound)
		return
	}

	// Get analytics
	analytics, err := h.streamingManager.GetPlaybackAnalytics(ctx, recordingID)
	if err != nil {
		h.logger.Printf("Failed to get analytics: %v", err)
		analytics = &PlaybackAnalytics{RecordingID: recordingID}
	}

	// Build response
	response := map[string]interface{}{
		"id":          recording.ID,
		"title":       recording.Title,
		"description": recording.Description,
		"status":      recording.Status,
		"duration":    recording.DurationSeconds,
		"created_at":  recording.CreatedAt,
		"stopped_at":  recording.StoppedAt,
		"size_bytes":  recording.FileSizeBytes,
		"format":      recording.Format,
		"analytics": map[string]interface{}{
			"total_sessions":   analytics.TotalSessions,
			"unique_viewers":   analytics.UniqueViewers,
			"total_playtime":   analytics.TotalPlayTime,
			"average_playtime": analytics.AveragePlayTime,
			"last_accessed_at": analytics.LastAccessedAt,
		},
	}

	// Add streaming info if ready
	if recording.Status == StatusCompleted {
		response["streaming_ready"] = true
		response["streaming_url"] = fmt.Sprintf("/api/v1/recordings/%s/stream/playlist.m3u8", recordingID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Write JSON response (simplified - use encoding/json in production)
	fmt.Fprintf(w, `{"data":%v}`, response)
}

// GetRecordingThumbnailHandler serves recording thumbnail
func (h *PlaybackHandlers) GetRecordingThumbnailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	recordingIDStr := strings.TrimPrefix(r.URL.Path, "/api/v1/recordings/")
	recordingIDStr = strings.TrimSuffix(recordingIDStr, "/thumbnail")

	recordingID, err := uuid.Parse(recordingIDStr)
	if err != nil {
		http.Error(w, "Invalid recording ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Verify recording exists
	if _, err := h.service.GetRecording(ctx, recordingID); err != nil {
		http.Error(w, "Recording not found", http.StatusNotFound)
		return
	}

	// Try to serve thumbnail from generated path
	thumbPath := filepath.Join(h.streamingManager.outputPath, recordingID.String()+"_thumb.jpg")

	file, err := os.Open(thumbPath)
	if err != nil {
		// Thumbnail not yet generated - return placeholder or 404
		http.Error(w, "Thumbnail not available", http.StatusNotFound)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "Could not read thumbnail", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	w.Header().Set("Cache-Control", "max-age=86400") // Cache for 1 day

	http.ServeContent(w, r, filepath.Base(thumbPath), fileInfo.ModTime(), file)
}

// TranscodeRecordingHandler initiates transcoding of recording
func (h *PlaybackHandlers) TranscodeRecordingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	recordingIDStr := strings.TrimPrefix(r.URL.Path, "/api/v1/recordings/")
	recordingIDStr = strings.TrimSuffix(recordingIDStr, "/transcode")

	recordingID, err := uuid.Parse(recordingIDStr)
	if err != nil {
		http.Error(w, "Invalid recording ID", http.StatusBadRequest)
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "hls"
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	recording, err := h.service.GetRecording(ctx, recordingID)
	if err != nil {
		http.Error(w, "Recording not found", http.StatusNotFound)
		return
	}

	if recording.Status != StatusCompleted {
		http.Error(w, "Recording not finished", http.StatusBadRequest)
		return
	}

	// Get storage URL for transcoding
	// For now, use FilePath from recording
	if recording.FilePath == nil || *recording.FilePath == "" {
		http.Error(w, "Could not access recording file", http.StatusInternalServerError)
		return
	}

	recordingPath := *recording.FilePath

	// Start transcoding in background (non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
		defer cancel()

		if format == "hls" {
			err = h.streamingManager.TranscodeToHLS(ctx, recordingID, recordingPath, DefaultHLSProfile)
		} else if format == "mp4" {
			err = h.streamingManager.TranscodeToMP4(ctx, recordingID, recordingPath)
		}

		if err != nil {
			h.logger.Printf("Transcoding failed for %s: %v", recordingID, err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, `{"status":"transcoding_started","format":"%s"}`, format)

	h.logger.Printf("Transcoding started for recording: %s, format: %s", recordingID, format)
}

// PlaybackProgressHandler updates playback progress for a user
func (h *PlaybackHandlers) PlaybackProgressHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	recordingIDStr := strings.TrimPrefix(r.URL.Path, "/api/v1/recordings/")
	recordingIDStr = strings.TrimSuffix(recordingIDStr, "/progress")

	recordingID, err := uuid.Parse(recordingIDStr)
	if err != nil {
		http.Error(w, "Invalid recording ID", http.StatusBadRequest)
		return
	}

	// Parse request body for position
	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	defer r.Body.Close()

	scanner := bufio.NewScanner(r.Body)
	var position int64

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "position=") {
			pos := strings.TrimPrefix(line, "position=")
			if p, err := strconv.ParseInt(pos, 10, 64); err == nil {
				position = p
			}
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Log progress event
	metadata := map[string]interface{}{
		"position": position,
	}

	err = h.streamingManager.LogPlaybackEvent(ctx, recordingID, uuid.Nil, "playback_progress", metadata)
	if err != nil {
		h.logger.Printf("Failed to log playback progress: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"position":%d}`, position)

	h.logger.Printf("Playback progress updated: %s, position: %d seconds", recordingID, position)
}

// PlaybackAnalyticsHandler returns detailed playback analytics
func (h *PlaybackHandlers) PlaybackAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	recordingIDStr := strings.TrimPrefix(r.URL.Path, "/api/v1/recordings/")
	recordingIDStr = strings.TrimSuffix(recordingIDStr, "/analytics")

	recordingID, err := uuid.Parse(recordingIDStr)
	if err != nil {
		http.Error(w, "Invalid recording ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	analytics, err := h.streamingManager.GetPlaybackAnalytics(ctx, recordingID)
	if err != nil {
		http.Error(w, "Could not retrieve analytics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, `{
		"recording_id":"%s",
		"total_sessions":%d,
		"unique_viewers":%d,
		"total_playtime":%d,
		"average_playtime":%d,
		"last_accessed_at":"%s"
	}`,
		analytics.RecordingID,
		analytics.TotalSessions,
		analytics.UniqueViewers,
		analytics.TotalPlayTime,
		analytics.AveragePlayTime,
		analytics.LastAccessedAt.Format(time.RFC3339),
	)

	h.logger.Printf("Analytics retrieved for recording: %s", recordingID)
}

// RegisterPlaybackRoutes registers all playback-related routes
func (h *PlaybackHandlers) RegisterPlaybackRoutes(mux *http.ServeMux) {
	// Streaming endpoints - use specific patterns to avoid conflicts
	mux.HandleFunc("/api/v1/recordings/{id}/stream/playlist.m3u8", h.StreamHLSPlaylistHandler)
	mux.HandleFunc("/api/v1/recordings/{id}/stream/", h.StreamHLSSegmentHandler)
	mux.HandleFunc("/api/v1/recordings/{id}/thumbnail", h.GetRecordingThumbnailHandler)
	mux.HandleFunc("/api/v1/recordings/{id}/transcode", h.TranscodeRecordingHandler)
	mux.HandleFunc("/api/v1/recordings/{id}/progress", h.PlaybackProgressHandler)
	mux.HandleFunc("/api/v1/recordings/{id}/analytics", h.PlaybackAnalyticsHandler)
}
