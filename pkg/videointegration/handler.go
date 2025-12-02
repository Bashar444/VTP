package videointegration

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Bashar444/VTP/pkg/utils"
)

// Handler handles HTTP requests for video integrations
type Handler struct {
	service *Service
}

// NewHandler creates a new handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateIntegrationRequest represents the request to create a video integration
type CreateIntegrationRequest struct {
	MeetingID   string `json:"meeting_id"`
	Provider    string `json:"provider"` // google_meet, zoom, jitsi, internal
	Title       string `json:"title"`
	ScheduledAt string `json:"scheduled_at"` // ISO 8601
	Duration    int    `json:"duration"`     // minutes
}

// CreateIntegration handles POST /api/v1/meetings/{id}/video
func (h *Handler) CreateIntegration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteErr(w, http.StatusMethodNotAllowed, nil)
		return
	}

	var req CreateIntegrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErr(w, http.StatusBadRequest, err)
		return
	}

	// Parse scheduled time
	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		scheduledAt = time.Now().Add(1 * time.Hour)
	}

	provider := Provider(req.Provider)
	if provider == "" {
		provider = ProviderJitsi // Default to Jitsi
	}

	mi, err := h.service.CreateMeetingIntegration(
		r.Context(),
		req.MeetingID,
		provider,
		req.Title,
		scheduledAt,
		req.Duration,
	)
	if err != nil {
		status := http.StatusBadRequest
		if err == ErrInvalidProvider {
			status = http.StatusBadRequest
		}
		utils.WriteErr(w, status, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, mi)
}

// GetIntegration handles GET /api/v1/meetings/{id}/video
func (h *Handler) GetIntegration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErr(w, http.StatusMethodNotAllowed, nil)
		return
	}

	meetingID := utils.Param(r, "id")
	mi, err := h.service.GetMeetingIntegration(r.Context(), meetingID)
	if err != nil {
		status := http.StatusInternalServerError
		if err == ErrIntegrationNotFound {
			status = http.StatusNotFound
		}
		utils.WriteErr(w, status, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, mi)
}

// GetJoinLink handles GET /api/v1/meetings/{id}/join
func (h *Handler) GetJoinLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErr(w, http.StatusMethodNotAllowed, nil)
		return
	}

	meetingID := utils.Param(r, "id")
	link, err := h.service.GetJoinLink(r.Context(), meetingID)
	if err != nil {
		status := http.StatusInternalServerError
		if err == ErrIntegrationNotFound {
			status = http.StatusNotFound
		}
		utils.WriteErr(w, status, err)
		return
	}

	// Check if redirect requested
	if r.URL.Query().Get("redirect") == "true" {
		http.Redirect(w, r, link, http.StatusFound)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"meeting_id": meetingID,
		"join_link":  link,
	})
}

// GetHostLink handles GET /api/v1/meetings/{id}/host
func (h *Handler) GetHostLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErr(w, http.StatusMethodNotAllowed, nil)
		return
	}

	meetingID := utils.Param(r, "id")
	link, err := h.service.GetHostLink(r.Context(), meetingID)
	if err != nil {
		status := http.StatusInternalServerError
		if err == ErrIntegrationNotFound {
			status = http.StatusNotFound
		}
		utils.WriteErr(w, status, err)
		return
	}

	// Check if redirect requested
	if r.URL.Query().Get("redirect") == "true" {
		http.Redirect(w, r, link, http.StatusFound)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"meeting_id": meetingID,
		"host_link":  link,
	})
}

// DeleteIntegration handles DELETE /api/v1/meetings/{id}/video
func (h *Handler) DeleteIntegration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteErr(w, http.StatusMethodNotAllowed, nil)
		return
	}

	meetingID := utils.Param(r, "id")
	if err := h.service.DeleteMeetingIntegration(r.Context(), meetingID); err != nil {
		status := http.StatusInternalServerError
		if err == ErrIntegrationNotFound {
			status = http.StatusNotFound
		}
		utils.WriteErr(w, status, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Integration deleted"})
}

// ListProviders handles GET /api/v1/video/providers
func (h *Handler) ListProviders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErr(w, http.StatusMethodNotAllowed, nil)
		return
	}

	providers := []map[string]interface{}{
		{
			"id":          "jitsi",
			"name":        "Jitsi Meet",
			"name_ar":     "جيتسي ميت",
			"description": "Free, open-source video conferencing",
			"free":        true,
			"recommended": true,
		},
		{
			"id":          "google_meet",
			"name":        "Google Meet",
			"name_ar":     "جوجل ميت",
			"description": "Google's video conferencing (requires Google Workspace)",
			"free":        false,
			"recommended": false,
		},
		{
			"id":          "zoom",
			"name":        "Zoom",
			"name_ar":     "زووم",
			"description": "Popular video conferencing platform",
			"free":        false,
			"recommended": false,
		},
		{
			"id":          "internal",
			"name":        "Built-in Video",
			"name_ar":     "فيديو مدمج",
			"description": "Platform's built-in WebRTC video",
			"free":        true,
			"recommended": true,
		},
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"providers": providers,
	})
}
