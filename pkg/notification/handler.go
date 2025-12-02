package notification

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Bashar444/VTP/pkg/models"
	"github.com/Bashar444/VTP/pkg/utils"
)

// Handler handles HTTP requests for notifications
type Handler struct {
	service *Service
}

// NewHandler creates a new notification handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateNotificationRequest represents the request to create a notification
type CreateNotificationRequest struct {
	UserID        string  `json:"user_id"`
	TitleAr       string  `json:"title_ar"`
	TitleEn       string  `json:"title_en"`
	MessageAr     string  `json:"message_ar"`
	MessageEn     string  `json:"message_en"`
	Type          string  `json:"type"`
	Channel       string  `json:"channel"`
	ReferenceType *string `json:"reference_type,omitempty"`
	ReferenceID   *string `json:"reference_id,omitempty"`
}

// CreateNotification handles POST /api/v1/notifications
func (h *Handler) CreateNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteErr(w, http.StatusMethodNotAllowed, nil)
		return
	}

	var req CreateNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErr(w, http.StatusBadRequest, err)
		return
	}

	n := &models.Notification{
		UserID:        req.UserID,
		TitleAr:       req.TitleAr,
		TitleEn:       req.TitleEn,
		MessageAr:     req.MessageAr,
		MessageEn:     req.MessageEn,
		Type:          req.Type,
		Channel:       req.Channel,
		ReferenceType: req.ReferenceType,
		ReferenceID:   req.ReferenceID,
	}

	if err := h.service.CreateNotification(r.Context(), n); err != nil {
		status := http.StatusBadRequest
		utils.WriteErr(w, status, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, n)
}

// GetUserNotifications handles GET /api/v1/notifications
func (h *Handler) GetUserNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErr(w, http.StatusMethodNotAllowed, nil)
		return
	}

	// Get user ID from auth context (in production)
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		utils.WriteErr(w, http.StatusBadRequest, ErrUserIDRequired)
		return
	}

	unreadOnly := r.URL.Query().Get("unread_only") == "true"
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	notifications, err := h.service.GetUserNotifications(r.Context(), userID, unreadOnly, page, pageSize)
	if err != nil {
		utils.WriteErr(w, http.StatusInternalServerError, err)
		return
	}

	unreadCount, _ := h.service.GetUnreadCount(r.Context(), userID)

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"notifications": notifications,
		"unread_count":  unreadCount,
		"page":          page,
		"page_size":     pageSize,
	})
}

// GetNotification handles GET /api/v1/notifications/{id}
func (h *Handler) GetNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErr(w, http.StatusMethodNotAllowed, nil)
		return
	}

	id := utils.Param(r, "id")
	notification, err := h.service.GetNotification(r.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if err == ErrNotificationNotFound {
			status = http.StatusNotFound
		}
		utils.WriteErr(w, status, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, notification)
}

// MarkAsRead handles POST /api/v1/notifications/{id}/read
func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteErr(w, http.StatusMethodNotAllowed, nil)
		return
	}

	id := utils.Param(r, "id")
	if err := h.service.MarkAsRead(r.Context(), id); err != nil {
		status := http.StatusInternalServerError
		if err == ErrNotificationNotFound {
			status = http.StatusNotFound
		}
		utils.WriteErr(w, status, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Notification marked as read"})
}

// MarkAllAsRead handles POST /api/v1/notifications/read-all
func (h *Handler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteErr(w, http.StatusMethodNotAllowed, nil)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		utils.WriteErr(w, http.StatusBadRequest, ErrUserIDRequired)
		return
	}

	if err := h.service.MarkAllAsRead(r.Context(), userID); err != nil {
		utils.WriteErr(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "All notifications marked as read"})
}

// GetUnreadCount handles GET /api/v1/notifications/unread-count
func (h *Handler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErr(w, http.StatusMethodNotAllowed, nil)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		utils.WriteErr(w, http.StatusBadRequest, ErrUserIDRequired)
		return
	}

	count, err := h.service.GetUnreadCount(r.Context(), userID)
	if err != nil {
		utils.WriteErr(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]int{"unread_count": count})
}

// DeleteNotification handles DELETE /api/v1/notifications/{id}
func (h *Handler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteErr(w, http.StatusMethodNotAllowed, nil)
		return
	}

	id := utils.Param(r, "id")
	if err := h.service.DeleteNotification(r.Context(), id); err != nil {
		status := http.StatusInternalServerError
		if err == ErrNotificationNotFound {
			status = http.StatusNotFound
		}
		utils.WriteErr(w, status, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "Notification deleted"})
}

// BroadcastNotification handles POST /api/v1/notifications/broadcast
func (h *Handler) BroadcastNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteErr(w, http.StatusMethodNotAllowed, nil)
		return
	}

	var req struct {
		UserIDs   []string `json:"user_ids"`
		TitleAr   string   `json:"title_ar"`
		TitleEn   string   `json:"title_en"`
		MessageAr string   `json:"message_ar"`
		MessageEn string   `json:"message_en"`
		Type      string   `json:"type"`
		Channel   string   `json:"channel"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErr(w, http.StatusBadRequest, err)
		return
	}

	if len(req.UserIDs) == 0 {
		utils.WriteErr(w, http.StatusBadRequest, ErrUserIDRequired)
		return
	}

	var notifications []models.Notification
	for _, userID := range req.UserIDs {
		n := models.Notification{
			UserID:         userID,
			TitleAr:        req.TitleAr,
			TitleEn:        req.TitleEn,
			MessageAr:      req.MessageAr,
			MessageEn:      req.MessageEn,
			Type:           req.Type,
			Channel:        req.Channel,
			DeliveryStatus: "pending",
		}
		notifications = append(notifications, n)
	}

	// Use bulk create for efficiency
	// Note: This would need to be added to the service
	for _, n := range notifications {
		_ = h.service.CreateNotification(r.Context(), &n)
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Broadcast sent",
		"count":   len(req.UserIDs),
	})
}
