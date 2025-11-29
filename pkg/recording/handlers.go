package recording

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// RecordingHandlers handles all HTTP endpoints for recordings
type RecordingHandlers struct {
	service *RecordingService
	logger  *log.Logger
}

// NewRecordingHandlers creates a new handlers instance
func NewRecordingHandlers(service *RecordingService, logger *log.Logger) *RecordingHandlers {
	if logger == nil {
		logger = log.New(io.Discard, "[RecordingHandlers] ", log.LstdFlags)
	}

	return &RecordingHandlers{
		service: service,
		logger:  logger,
	}
}

// StartRecordingHandler handles POST /api/v1/recordings/start
func (h *RecordingHandlers) StartRecordingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse request body
	var req StartRecordingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Printf("Failed to parse request: %v", err)
		h.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get user ID from context or header
	userID, err := h.getUserID(r)
	if err != nil {
		h.logger.Printf("Failed to get user ID: %v", err)
		h.writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Start recording
	recording, err := h.service.StartRecording(r.Context(), &req, userID)
	if err != nil {
		h.logger.Printf("Failed to start recording: %v", err)
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to start recording: %v", err))
		return
	}

	// Return response
	response := StartRecordingResponse{
		RecordingID: recording.ID,
		Status:      recording.Status,
		StartedAt:   recording.StartedAt,
		RoomID:      recording.RoomID,
		Message:     "Recording started successfully",
	}

	h.writeJSON(w, http.StatusOK, response)
	h.logger.Printf("Recording started: %s in room %s", recording.ID, recording.RoomID)
}

// StopRecordingHandler handles POST /api/v1/recordings/{id}/stop
func (h *RecordingHandlers) StopRecordingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract recording ID from URL
	recordingIDStr := strings.TrimPrefix(r.URL.Path, "/api/v1/recordings/")
	recordingIDStr = strings.TrimSuffix(recordingIDStr, "/stop")

	recordingID, err := uuid.Parse(recordingIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid recording ID")
		return
	}

	// Stop recording
	recording, err := h.service.StopRecording(r.Context(), recordingID)
	if err != nil {
		h.logger.Printf("Failed to stop recording %s: %v", recordingID, err)
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to stop recording: %v", err))
		return
	}

	if recording == nil {
		h.writeError(w, http.StatusNotFound, "Recording not found")
		return
	}

	// Return response
	response := StopRecordingResponse{
		RecordingID:     recording.ID,
		Status:          recording.Status,
		StoppedAt:       *recording.StoppedAt,
		DurationSeconds: *recording.DurationSeconds,
		FilePath:        recording.FilePath,
		Message:         "Recording stopped successfully",
	}

	h.writeJSON(w, http.StatusOK, response)
	h.logger.Printf("Recording stopped: %s (duration: %d seconds)", recordingID, *recording.DurationSeconds)
}

// GetRecordingHandler handles GET /api/v1/recordings/{id}
func (h *RecordingHandlers) GetRecordingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract recording ID from URL
	recordingIDStr := strings.TrimPrefix(r.URL.Path, "/api/v1/recordings/")

	recordingID, err := uuid.Parse(recordingIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid recording ID")
		return
	}

	// Get recording
	recording, err := h.service.GetRecording(r.Context(), recordingID)
	if err != nil {
		h.logger.Printf("Failed to get recording %s: %v", recordingID, err)
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get recording: %v", err))
		return
	}

	if recording == nil {
		h.writeError(w, http.StatusNotFound, "Recording not found")
		return
	}

	// Convert to response
	response := GetRecordingResponse{
		ID:              recording.ID,
		RoomID:          recording.RoomID,
		Title:           recording.Title,
		Description:     recording.Description,
		StartedBy:       recording.StartedBy,
		StartedAt:       recording.StartedAt,
		StoppedAt:       recording.StoppedAt,
		DurationSeconds: recording.DurationSeconds,
		Status:          recording.Status,
		Format:          recording.Format,
		FileSizeBytes:   recording.FileSizeBytes,
		MimeType:        recording.MimeType,
		Metadata:        recording.Metadata,
		CreatedAt:       recording.CreatedAt,
		UpdatedAt:       recording.UpdatedAt,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// ListRecordingsHandler handles GET /api/v1/recordings
func (h *RecordingHandlers) ListRecordingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse query parameters
	roomIDStr := r.URL.Query().Get("room_id")
	userIDStr := r.URL.Query().Get("user_id")
	status := r.URL.Query().Get("status")
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

	// Default pagination
	limitInt := 10
	if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 100 {
		limitInt = l
	}

	offsetInt := 0
	if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
		offsetInt = o
	}

	// Build query
	query := &RecordingListQuery{
		Limit:  limitInt,
		Offset: offsetInt,
		Status: status,
	}

	// Parse UUIDs if provided
	if roomIDStr != "" {
		if roomID, err := uuid.Parse(roomIDStr); err == nil {
			query.RoomID = roomID
		}
	}

	if userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			query.UserID = userID
		}
	}

	// List recordings
	recordings, total, err := h.service.ListRecordings(r.Context(), query)
	if err != nil {
		h.logger.Printf("Failed to list recordings: %v", err)
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list recordings: %v", err))
		return
	}

	// Convert to response
	recordingResponses := make([]GetRecordingResponse, len(recordings))
	for i, rec := range recordings {
		recordingResponses[i] = GetRecordingResponse{
			ID:              rec.ID,
			RoomID:          rec.RoomID,
			Title:           rec.Title,
			Description:     rec.Description,
			StartedBy:       rec.StartedBy,
			StartedAt:       rec.StartedAt,
			StoppedAt:       rec.StoppedAt,
			DurationSeconds: rec.DurationSeconds,
			Status:          rec.Status,
			Format:          rec.Format,
			FileSizeBytes:   rec.FileSizeBytes,
			MimeType:        rec.MimeType,
			Metadata:        rec.Metadata,
			CreatedAt:       rec.CreatedAt,
			UpdatedAt:       rec.UpdatedAt,
		}
	}

	response := ListRecordingsResponse{
		Recordings: recordingResponses,
		Total:      total,
		Limit:      limitInt,
		Offset:     offsetInt,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// DeleteRecordingHandler handles DELETE /api/v1/recordings/{id}
func (h *RecordingHandlers) DeleteRecordingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract recording ID from URL
	recordingIDStr := strings.TrimPrefix(r.URL.Path, "/api/v1/recordings/")

	recordingID, err := uuid.Parse(recordingIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid recording ID")
		return
	}

	// Delete recording
	recording, err := h.service.DeleteRecording(r.Context(), recordingID)
	if err != nil {
		h.logger.Printf("Failed to delete recording %s: %v", recordingID, err)
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete recording: %v", err))
		return
	}

	if recording == nil {
		h.writeError(w, http.StatusNotFound, "Recording not found")
		return
	}

	// Return response
	response := DeleteRecordingResponse{
		RecordingID: recording.ID,
		Status:      recording.Status,
		DeletedAt:   *recording.DeletedAt,
		Message:     "Recording deleted successfully",
	}

	h.writeJSON(w, http.StatusOK, response)
	h.logger.Printf("Recording deleted: %s", recordingID)
}

// Helper methods

// writeJSON writes a JSON response
func (h *RecordingHandlers) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Printf("Failed to write JSON response: %v", err)
	}
}

// writeError writes an error response
func (h *RecordingHandlers) writeError(w http.ResponseWriter, status int, message string) {
	errorResponse := map[string]interface{}{
		"error":  message,
		"status": status,
	}
	h.writeJSON(w, status, errorResponse)
}

// getUserID extracts user ID from request context or header
func (h *RecordingHandlers) getUserID(r *http.Request) (uuid.UUID, error) {
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

// RegisterRoutes registers all recording routes with a router
// This assumes you have a router like chi or gorilla/mux
func (h *RecordingHandlers) RegisterRoutes(mux *http.ServeMux) {
	// POST /api/v1/recordings/start
	mux.HandleFunc("/api/v1/recordings/start", h.StartRecordingHandler)

	// POST /api/v1/recordings/{id}/stop
	mux.HandleFunc("/api/v1/recordings/{id}/stop", h.StopRecordingHandler)

	// GET /api/v1/recordings
	mux.HandleFunc("/api/v1/recordings", h.ListRecordingsHandler)

	// GET /api/v1/recordings/{id}
	// DELETE /api/v1/recordings/{id}
	mux.HandleFunc("/api/v1/recordings/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetRecordingHandler(w, r)
		} else if r.Method == http.MethodDelete {
			h.DeleteRecordingHandler(w, r)
		} else {
			h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})
}
