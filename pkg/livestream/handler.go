package livestream

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Bashar444/VTP/pkg/auth"
	"github.com/Bashar444/VTP/pkg/models"
)

// Handler handles HTTP requests for live streaming
type Handler struct {
	service      *Service
	logger       *log.Logger
	chatHandler  *ChatHandler
	tokenService *auth.TokenService
}

// NewHandler creates a new live streaming handler
func NewHandler(service *Service, logger *log.Logger, tokenService *auth.TokenService) *Handler {
	return &Handler{
		service:      service,
		logger:       logger,
		tokenService: tokenService,
	}
}

// SetChatHandler sets the chat handler for routing chat requests
func (h *Handler) SetChatHandler(ch *ChatHandler) {
	h.chatHandler = ch
}

// getUserIDFromRequest extracts user ID from JWT token in Authorization header
func (h *Handler) getUserIDFromRequest(r *http.Request) string {
	// First try context (if auth middleware was used)
	if ctx := r.Context().Value("user_id"); ctx != nil {
		if userID, ok := ctx.(string); ok && userID != "" {
			return userID
		}
	}

	// Try X-User-ID header
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return userID
	}

	// Extract from JWT token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if h.tokenService != nil {
		claims, err := h.tokenService.ValidateToken(token)
		if err == nil {
			return claims.UserID
		}
	}

	return ""
}

// CreateSessionRequest represents the request to create a streaming session
type CreateSessionRequest struct {
	RoomID          string  `json:"room_id"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	CourseID        *string `json:"course_id"`
	ScheduledAt     *string `json:"scheduled_at"` // ISO 8601
	MaxParticipants int     `json:"max_participants"`
}

// SessionResponse represents the streaming session response
type SessionResponse struct {
	ID              string  `json:"id"`
	RoomID          string  `json:"room_id"`
	InstructorID    string  `json:"instructor_id"`
	InstructorName  string  `json:"instructor_name,omitempty"`
	CourseID        *string `json:"course_id,omitempty"`
	CourseTitle     string  `json:"course_title,omitempty"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	Status          string  `json:"status"`
	ScheduledAt     *string `json:"scheduled_at,omitempty"`
	StartedAt       *string `json:"started_at,omitempty"`
	EndedAt         *string `json:"ended_at,omitempty"`
	MaxParticipants int     `json:"max_participants"`
	IsRecording     bool    `json:"is_recording"`
	Settings        string  `json:"settings"`
	StreamURL       string  `json:"stream_url,omitempty"`
	CreatedAt       string  `json:"created_at"`
}

// ListSessionsResponse represents the list sessions response
type ListSessionsResponse struct {
	Sessions []*SessionResponse `json:"sessions"`
	Total    int                `json:"total"`
	Limit    int                `json:"limit"`
	Offset   int                `json:"offset"`
}

// RegisterRoutes registers all live streaming HTTP routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Session management
	mux.HandleFunc("/api/v1/streaming/sessions", h.handleSessions)
	mux.HandleFunc("/api/v1/streaming/sessions/", h.handleSessionByID)
	mux.HandleFunc("/api/v1/streaming/sessions/start", h.handleStartSession)
	mux.HandleFunc("/api/v1/streaming/sessions/live", h.handleLiveSessions)
}

// handleSessions handles GET (list) and POST (create) for sessions
func (h *Handler) handleSessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListSessions(w, r)
	case http.MethodPost:
		h.CreateSession(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleSessionByID handles individual session operations
func (h *Handler) handleSessionByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/streaming/sessions/")
	parts := strings.Split(path, "/")

	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	sessionID := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch action {
	case "start":
		if r.Method == http.MethodPost {
			h.StartStream(w, r, sessionID)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "stop":
		if r.Method == http.MethodPost {
			h.StopStream(w, r, sessionID)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "recording":
		if len(parts) > 2 && parts[2] == "start" {
			h.StartRecording(w, r, sessionID)
		} else if len(parts) > 2 && parts[2] == "stop" {
			h.StopRecording(w, r, sessionID)
		} else {
			http.Error(w, "Invalid recording action", http.StatusBadRequest)
		}
	case "join":
		if r.Method == http.MethodPost {
			h.JoinSession(w, r, sessionID)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "leave":
		if r.Method == http.MethodPost {
			h.LeaveSession(w, r, sessionID)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "participants":
		if r.Method == http.MethodGet {
			h.GetParticipants(w, r, sessionID)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "chat":
		// Route to chat handler
		if h.chatHandler != nil {
			h.chatHandler.HandleChatRequest(w, r, sessionID, parts[2:])
		} else {
			http.Error(w, "Chat service unavailable", http.StatusServiceUnavailable)
		}
	case "":
		if r.Method == http.MethodGet {
			h.GetSession(w, r, sessionID)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.NotFound(w, r)
	}
}

// handleStartSession handles POST /api/v1/streaming/sessions/start (create and start)
func (h *Handler) handleStartSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	h.CreateAndStartSession(w, r)
}

// handleLiveSessions handles GET /api/v1/streaming/sessions/live
func (h *Handler) handleLiveSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	h.ListLiveSessions(w, r)
}

// CreateSession handles POST /api/v1/streaming/sessions
func (h *Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromRequest(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Title == "" {
		respondError(w, http.StatusBadRequest, "Title is required")
		return
	}

	if req.RoomID == "" {
		// Generate room ID if not provided
		req.RoomID = generateRoomID()
	}

	session := &models.StreamingSession{
		RoomID:          req.RoomID,
		InstructorID:    userID,
		CourseID:        req.CourseID,
		Title:           req.Title,
		Description:     req.Description,
		Status:          "scheduled",
		MaxParticipants: req.MaxParticipants,
	}

	if req.MaxParticipants == 0 {
		session.MaxParticipants = 100
	}

	if err := h.service.CreateSession(r.Context(), session); err != nil {
		h.logger.Printf("Failed to create session: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	respondJSON(w, http.StatusCreated, toSessionResponse(session))
}

// CreateAndStartSession creates a new session and immediately starts it
func (h *Handler) CreateAndStartSession(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromRequest(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Title == "" {
		respondError(w, http.StatusBadRequest, "Title is required")
		return
	}

	if req.RoomID == "" {
		req.RoomID = generateRoomID()
	}

	session := &models.StreamingSession{
		RoomID:          req.RoomID,
		InstructorID:    userID,
		CourseID:        req.CourseID,
		Title:           req.Title,
		Description:     req.Description,
		Status:          "scheduled",
		MaxParticipants: req.MaxParticipants,
	}

	if req.MaxParticipants == 0 {
		session.MaxParticipants = 100
	}

	// Create session
	if err := h.service.CreateSession(r.Context(), session); err != nil {
		h.logger.Printf("Failed to create session: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	// Start session immediately
	session, err := h.service.StartStream(r.Context(), session.ID, userID)
	if err != nil {
		h.logger.Printf("Failed to start session: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to start session")
		return
	}

	resp := toSessionResponse(session)
	resp.StreamURL = "/stream/" + session.RoomID

	respondJSON(w, http.StatusCreated, resp)
}

// GetSession handles GET /api/v1/streaming/sessions/{id}
func (h *Handler) GetSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	session, err := h.service.GetSession(r.Context(), sessionID)
	if err != nil {
		if err == ErrSessionNotFound {
			respondError(w, http.StatusNotFound, "Session not found")
			return
		}
		h.logger.Printf("Failed to get session: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get session")
		return
	}

	respondJSON(w, http.StatusOK, toSessionResponse(session))
}

// ListSessions handles GET /api/v1/streaming/sessions
func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	// Parse query params
	instructorID := r.URL.Query().Get("instructor_id")
	if instructorID == "me" {
		instructorID = h.getUserIDFromRequest(r)
	}

	status := r.URL.Query().Get("status")
	liveOnly := r.URL.Query().Get("live") == "true"

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 20
	}

	sessions, total, err := h.service.ListSessions(r.Context(), instructorID, status, liveOnly, limit, offset)
	if err != nil {
		h.logger.Printf("Failed to list sessions: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list sessions")
		return
	}

	resp := ListSessionsResponse{
		Sessions: make([]*SessionResponse, len(sessions)),
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}

	for i, s := range sessions {
		resp.Sessions[i] = toSessionResponse(s)
	}

	respondJSON(w, http.StatusOK, resp)
}

// ListLiveSessions handles GET /api/v1/streaming/sessions/live
func (h *Handler) ListLiveSessions(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 50
	}

	sessions, total, err := h.service.ListLiveSessions(r.Context(), limit, offset)
	if err != nil {
		h.logger.Printf("Failed to list live sessions: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list live sessions")
		return
	}

	resp := ListSessionsResponse{
		Sessions: make([]*SessionResponse, len(sessions)),
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}

	for i, s := range sessions {
		resp.Sessions[i] = toSessionResponse(s)
		resp.Sessions[i].StreamURL = "/stream/" + s.RoomID
	}

	respondJSON(w, http.StatusOK, resp)
}

// StartStream handles POST /api/v1/streaming/sessions/{id}/start
func (h *Handler) StartStream(w http.ResponseWriter, r *http.Request, sessionID string) {
	userID := h.getUserIDFromRequest(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	session, err := h.service.StartStream(r.Context(), sessionID, userID)
	if err != nil {
		switch err {
		case ErrSessionNotFound:
			respondError(w, http.StatusNotFound, "Session not found")
		case ErrUnauthorized:
			respondError(w, http.StatusForbidden, "Not authorized to start this stream")
		case ErrAlreadyLive:
			respondError(w, http.StatusConflict, "Stream is already live")
		default:
			h.logger.Printf("Failed to start stream: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to start stream")
		}
		return
	}

	resp := toSessionResponse(session)
	resp.StreamURL = "/stream/" + session.RoomID

	respondJSON(w, http.StatusOK, resp)
}

// StopStream handles POST /api/v1/streaming/sessions/{id}/stop
func (h *Handler) StopStream(w http.ResponseWriter, r *http.Request, sessionID string) {
	userID := h.getUserIDFromRequest(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	session, err := h.service.StopStream(r.Context(), sessionID, userID)
	if err != nil {
		switch err {
		case ErrSessionNotFound:
			respondError(w, http.StatusNotFound, "Session not found")
		case ErrUnauthorized:
			respondError(w, http.StatusForbidden, "Not authorized to stop this stream")
		case ErrNotLive:
			respondError(w, http.StatusConflict, "Stream is not live")
		default:
			h.logger.Printf("Failed to stop stream: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to stop stream")
		}
		return
	}

	respondJSON(w, http.StatusOK, toSessionResponse(session))
}

// StartRecording handles POST /api/v1/streaming/sessions/{id}/recording/start
func (h *Handler) StartRecording(w http.ResponseWriter, r *http.Request, sessionID string) {
	userID := h.getUserIDFromRequest(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	session, err := h.service.StartRecording(r.Context(), sessionID, userID)
	if err != nil {
		h.logger.Printf("Failed to start recording: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to start recording")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"session_id":   session.ID,
		"is_recording": session.IsRecording,
		"message":      "Recording started",
	})
}

// StopRecording handles POST /api/v1/streaming/sessions/{id}/recording/stop
func (h *Handler) StopRecording(w http.ResponseWriter, r *http.Request, sessionID string) {
	userID := h.getUserIDFromRequest(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	session, err := h.service.StopRecording(r.Context(), sessionID, userID)
	if err != nil {
		h.logger.Printf("Failed to stop recording: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to stop recording")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"session_id":   session.ID,
		"is_recording": session.IsRecording,
		"message":      "Recording stopped",
	})
}

// JoinSession handles POST /api/v1/streaming/sessions/{id}/join
func (h *Handler) JoinSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	userID := h.getUserIDFromRequest(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Get role from request body (optional)
	var req struct {
		Role string `json:"role"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Role == "" {
		req.Role = "viewer"
	}

	if err := h.service.JoinSession(r.Context(), sessionID, userID, req.Role); err != nil {
		if err == ErrSessionNotFound {
			respondError(w, http.StatusNotFound, "Session not found")
			return
		}
		if err == ErrNotLive {
			respondError(w, http.StatusConflict, "Session is not live")
			return
		}
		h.logger.Printf("Failed to join session: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to join session")
		return
	}

	session, _ := h.service.GetSession(r.Context(), sessionID)
	resp := toSessionResponse(session)
	resp.StreamURL = "/stream/" + session.RoomID

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"session":    resp,
		"stream_url": resp.StreamURL,
	})
}

// LeaveSession handles POST /api/v1/streaming/sessions/{id}/leave
func (h *Handler) LeaveSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	userID := h.getUserIDFromRequest(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	if err := h.service.LeaveSession(r.Context(), sessionID, userID); err != nil {
		h.logger.Printf("Failed to leave session: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to leave session")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Left session successfully",
	})
}

// GetParticipants handles GET /api/v1/streaming/sessions/{id}/participants
func (h *Handler) GetParticipants(w http.ResponseWriter, r *http.Request, sessionID string) {
	participants, err := h.service.GetParticipants(r.Context(), sessionID)
	if err != nil {
		h.logger.Printf("Failed to get participants: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get participants")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"session_id":   sessionID,
		"participants": participants,
		"count":        len(participants),
	})
}

// Helper functions

func toSessionResponse(s *models.StreamingSession) *SessionResponse {
	resp := &SessionResponse{
		ID:              s.ID,
		RoomID:          s.RoomID,
		InstructorID:    s.InstructorID,
		CourseID:        s.CourseID,
		Title:           s.Title,
		Description:     s.Description,
		Status:          s.Status,
		MaxParticipants: s.MaxParticipants,
		IsRecording:     s.IsRecording,
		Settings:        s.Settings,
		CreatedAt:       s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if s.ScheduledAt != nil {
		t := s.ScheduledAt.Format("2006-01-02T15:04:05Z07:00")
		resp.ScheduledAt = &t
	}
	if s.StartedAt != nil {
		t := s.StartedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.StartedAt = &t
	}
	if s.EndedAt != nil {
		t := s.EndedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.EndedAt = &t
	}

	return resp
}

func generateRoomID() string {
	return "stream-" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
