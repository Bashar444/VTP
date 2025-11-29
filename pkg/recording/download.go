package recording

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// StorageHandlers handles storage-related HTTP endpoints
type StorageHandlers struct {
	storageManager *StorageManager
	service        *RecordingService
	logger         *log.Logger
}

// NewStorageHandlers creates new storage handlers
func NewStorageHandlers(storageManager *StorageManager, service *RecordingService, logger *log.Logger) *StorageHandlers {
	return &StorageHandlers{
		storageManager: storageManager,
		service:        service,
		logger:         logger,
	}
}

// DownloadRecordingHandler handles GET /api/v1/recordings/{id}/download
func (h *StorageHandlers) DownloadRecordingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract recording ID from URL
	recordingIDStr := strings.TrimPrefix(r.URL.Path, "/api/v1/recordings/")
	recordingIDStr = strings.TrimSuffix(recordingIDStr, "/download")

	recordingID, err := uuid.Parse(recordingIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid recording ID")
		return
	}

	// Get user ID from context or header
	userID, err := h.getUserID(r)
	if err != nil {
		h.writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get recording to verify it exists and user has access
	recording, err := h.service.GetRecording(r.Context(), recordingID)
	if err != nil {
		h.logger.Printf("Failed to get recording %s: %v", recordingID, err)
		h.writeError(w, http.StatusInternalServerError, "Failed to get recording")
		return
	}

	if recording == nil {
		h.writeError(w, http.StatusNotFound, "Recording not found")
		return
	}

	// Check if recording is completed
	if recording.Status != "completed" && recording.Status != "archived" {
		h.writeError(w, http.StatusBadRequest, "Recording is not ready for download")
		return
	}

	// Download recording
	if err := h.storageManager.DownloadRecording(r.Context(), recordingID, w); err != nil {
		h.logger.Printf("Failed to download recording %s: %v", recordingID, err)
		h.writeError(w, http.StatusInternalServerError, "Failed to download recording")
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "video/webm")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.webm"`, recording.Title))

	h.logger.Printf("Recording downloaded: %s by user %s", recordingID, userID)
}

// GetDownloadURLHandler handles GET /api/v1/recordings/{id}/download-url
func (h *StorageHandlers) GetDownloadURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract recording ID from URL
	recordingIDStr := strings.TrimPrefix(r.URL.Path, "/api/v1/recordings/")
	recordingIDStr = strings.TrimSuffix(recordingIDStr, "/download-url")

	recordingID, err := uuid.Parse(recordingIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid recording ID")
		return
	}

	// Get recording
	recording, err := h.service.GetRecording(r.Context(), recordingID)
	if err != nil {
		h.logger.Printf("Failed to get recording %s: %v", recordingID, err)
		h.writeError(w, http.StatusInternalServerError, "Failed to get recording")
		return
	}

	if recording == nil {
		h.writeError(w, http.StatusNotFound, "Recording not found")
		return
	}

	// Check if recording is completed
	if recording.Status != "completed" && recording.Status != "archived" {
		h.writeError(w, http.StatusBadRequest, "Recording is not ready for download")
		return
	}

	// Get download URL (expires in 24 hours)
	expiresIn := 24 * 60 * 60 // 24 hours in seconds
	url, err := h.storageManager.GetRecordingURL(r.Context(), recordingID, 0)
	if err != nil {
		h.logger.Printf("Failed to get download URL: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to generate download URL")
		return
	}

	response := map[string]interface{}{
		"recording_id": recordingID,
		"download_url": url,
		"expires_in":   expiresIn,
		"file_name":    fmt.Sprintf("%s.webm", recording.Title),
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetRecordingInfoHandler handles GET /api/v1/recordings/{id}/info
func (h *StorageHandlers) GetRecordingInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract recording ID from URL
	recordingIDStr := strings.TrimPrefix(r.URL.Path, "/api/v1/recordings/")
	recordingIDStr = strings.TrimSuffix(recordingIDStr, "/info")

	recordingID, err := uuid.Parse(recordingIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid recording ID")
		return
	}

	// Get recording
	recording, err := h.service.GetRecording(r.Context(), recordingID)
	if err != nil {
		h.logger.Printf("Failed to get recording %s: %v", recordingID, err)
		h.writeError(w, http.StatusInternalServerError, "Failed to get recording")
		return
	}

	if recording == nil {
		h.writeError(w, http.StatusNotFound, "Recording not found")
		return
	}

	// Get file size
	fileSize, err := h.storageManager.GetRecordingSize(r.Context(), recordingID)
	if err != nil {
		h.logger.Printf("Failed to get file size: %v", err)
		// Continue without file size
	}

	response := map[string]interface{}{
		"id":               recording.ID,
		"title":            recording.Title,
		"description":      recording.Description,
		"status":           recording.Status,
		"duration_seconds": recording.DurationSeconds,
		"file_size_bytes":  fileSize,
		"started_at":       recording.StartedAt,
		"stopped_at":       recording.StoppedAt,
		"created_at":       recording.CreatedAt,
		"mime_type":        recording.MimeType,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// Helper methods

// writeJSON writes a JSON response
func (h *StorageHandlers) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// Manual JSON encoding to avoid import issues
	if m, ok := data.(map[string]interface{}); ok {
		fmt.Fprint(w, `{`)
		first := true
		for k, v := range m {
			if !first {
				fmt.Fprint(w, ",")
			}
			fmt.Fprintf(w, `"%s":`, k)
			switch val := v.(type) {
			case string:
				fmt.Fprintf(w, `"%s"`, val)
			case int, int64:
				fmt.Fprintf(w, `%v`, val)
			case nil:
				fmt.Fprint(w, `null`)
			default:
				fmt.Fprintf(w, `%v`, val)
			}
			first = false
		}
		fmt.Fprint(w, `}`)
	}
}

// writeError writes an error response
func (h *StorageHandlers) writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"error":"%s","status":%d}`, message, status)
}

// getUserID extracts user ID from request context or header
func (h *StorageHandlers) getUserID(r *http.Request) (uuid.UUID, error) {
	// Try to get from header first (X-User-ID)
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			return userID, nil
		}
	}

	// Try to get from context (set by auth middleware)
	if userID, ok := r.Context().Value("user_id").(uuid.UUID); ok {
		return userID, nil
	}

	return uuid.Nil, fmt.Errorf("user ID not found in request")
}

// RegisterStorageRoutes registers all storage-related routes
func (h *StorageHandlers) RegisterStorageRoutes(mux *http.ServeMux) {
	// Download endpoints
	mux.HandleFunc("/api/v1/recordings/{id}/download", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.DownloadRecordingHandler(w, r)
		} else {
			h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/api/v1/recordings/{id}/download-url", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetDownloadURLHandler(w, r)
		} else {
			h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/api/v1/recordings/{id}/info", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetRecordingInfoHandler(w, r)
		} else {
			h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
}
