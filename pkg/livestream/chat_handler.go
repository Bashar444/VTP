package livestream

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Bashar444/VTP/pkg/auth"
	"github.com/Bashar444/VTP/pkg/models"
)

// ChatHandler handles HTTP requests for stream chat
type ChatHandler struct {
	chatRepo     *ChatRepository
	sessionRepo  *Repository
	logger       *log.Logger
	tokenService *auth.TokenService
}

// NewChatHandler creates a new chat handler
func NewChatHandler(chatRepo *ChatRepository, sessionRepo *Repository, logger *log.Logger, tokenService *auth.TokenService) *ChatHandler {
	return &ChatHandler{
		chatRepo:     chatRepo,
		sessionRepo:  sessionRepo,
		logger:       logger,
		tokenService: tokenService,
	}
}

// getUserIDFromRequest extracts user ID from JWT token in Authorization header
func (h *ChatHandler) getUserIDFromRequest(r *http.Request) string {
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

// CreateMessageRequest represents the request to send a chat message
type CreateMessageRequest struct {
	Content     string  `json:"content"`
	MessageType string  `json:"message_type"` // text, question, announcement
	ReplyToID   *string `json:"reply_to_id"`
}

// ChatMessageResponse represents a chat message response
type ChatMessageResponse struct {
	ID          string  `json:"id"`
	SessionID   string  `json:"session_id"`
	UserID      string  `json:"user_id"`
	UserName    string  `json:"user_name,omitempty"`
	MessageType string  `json:"message_type"`
	Content     string  `json:"content"`
	IsPinned    bool    `json:"is_pinned"`
	IsAnswered  bool    `json:"is_answered"`
	ReplyToID   *string `json:"reply_to_id,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

// RegisterChatRoutes is deprecated - chat routes are now handled via Handler.SetChatHandler()
// This method is kept for backward compatibility but does nothing
func (h *ChatHandler) RegisterChatRoutes(mux *http.ServeMux) {
	// No-op: chat routes are now integrated into the main handler
}

// HandleChatRequest handles chat-related requests routed from the main handler
func (h *ChatHandler) HandleChatRequest(w http.ResponseWriter, r *http.Request, sessionID string, parts []string) {
	if len(parts) == 0 {
		// Direct /chat endpoint
		if r.Method == http.MethodGet {
			h.GetMessages(w, r, sessionID)
		} else if r.Method == http.MethodPost {
			h.SendMessage(w, r, sessionID)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	subAction := parts[0]
	switch subAction {
	case "questions":
		if r.Method == http.MethodGet {
			h.GetQuestions(w, r, sessionID)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	case "pinned":
		if r.Method == http.MethodGet {
			h.GetPinnedMessages(w, r, sessionID)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		// Message ID action (pin, answer)
		messageID := subAction
		if len(parts) > 1 {
			messageAction := parts[1]
			switch messageAction {
			case "pin":
				h.PinMessage(w, r, sessionID, messageID)
			case "answer":
				h.MarkAsAnswered(w, r, sessionID, messageID)
			default:
				http.NotFound(w, r)
			}
		} else {
			http.NotFound(w, r)
		}
	}
}

// handleChatRoutes routes chat-related requests (deprecated - kept for reference)
func (h *ChatHandler) handleChatRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/streaming/sessions/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 {
		return // Let other handlers process
	}

	sessionID := parts[0]
	action := parts[1]

	switch action {
	case "chat":
		if len(parts) > 2 {
			// Sub-actions like /chat/questions, /chat/pinned
			subAction := parts[2]
			switch subAction {
			case "questions":
				if r.Method == http.MethodGet {
					h.GetQuestions(w, r, sessionID)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			case "pinned":
				if r.Method == http.MethodGet {
					h.GetPinnedMessages(w, r, sessionID)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			default:
				// Message ID action (pin, answer)
				messageID := subAction
				if len(parts) > 3 {
					messageAction := parts[3]
					switch messageAction {
					case "pin":
						h.PinMessage(w, r, sessionID, messageID)
					case "answer":
						h.MarkAsAnswered(w, r, sessionID, messageID)
					default:
						http.NotFound(w, r)
					}
				} else {
					http.NotFound(w, r)
				}
			}
		} else {
			// Base /chat endpoint
			switch r.Method {
			case http.MethodGet:
				h.GetMessages(w, r, sessionID)
			case http.MethodPost:
				h.SendMessage(w, r, sessionID)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		}
	}
}

// SendMessage handles POST /api/v1/streaming/sessions/{id}/chat
func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request, sessionID string) {
	userID := h.getUserIDFromRequest(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Content == "" {
		respondError(w, http.StatusBadRequest, "Content is required")
		return
	}

	// Verify session exists
	session, err := h.sessionRepo.GetSessionByID(r.Context(), sessionID)
	if err != nil {
		if err == ErrSessionNotFound {
			respondError(w, http.StatusNotFound, "Session not found")
			return
		}
		h.logger.Printf("Failed to get session: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get session")
		return
	}

	// Only allow chat in live sessions
	if session.Status != "live" {
		respondError(w, http.StatusConflict, "Chat is only available during live sessions")
		return
	}

	if req.MessageType == "" {
		req.MessageType = "text"
	}

	// Only instructors can send announcements
	if req.MessageType == "announcement" && session.InstructorID != userID {
		respondError(w, http.StatusForbidden, "Only the instructor can send announcements")
		return
	}

	msg := &models.StreamChatMessage{
		SessionID:   sessionID,
		UserID:      userID,
		MessageType: req.MessageType,
		Content:     req.Content,
		ReplyToID:   req.ReplyToID,
	}

	if err := h.chatRepo.CreateMessage(r.Context(), msg); err != nil {
		h.logger.Printf("Failed to create message: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to send message")
		return
	}

	respondJSON(w, http.StatusCreated, toChatMessageResponse(msg))
}

// GetMessages handles GET /api/v1/streaming/sessions/{id}/chat
func (h *ChatHandler) GetMessages(w http.ResponseWriter, r *http.Request, sessionID string) {
	messageType := r.URL.Query().Get("type")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 100
	}

	// Check for recent messages query
	if r.URL.Query().Get("recent") == "true" {
		messages, err := h.chatRepo.GetRecentMessages(r.Context(), sessionID, limit)
		if err != nil {
			h.logger.Printf("Failed to get recent messages: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to get messages")
			return
		}

		resp := make([]*ChatMessageResponse, len(messages))
		for i, m := range messages {
			resp[i] = toChatMessageResponse(m)
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"messages": resp,
			"count":    len(resp),
		})
		return
	}

	messages, err := h.chatRepo.GetMessages(r.Context(), sessionID, messageType, limit, offset)
	if err != nil {
		h.logger.Printf("Failed to get messages: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get messages")
		return
	}

	resp := make([]*ChatMessageResponse, len(messages))
	for i, m := range messages {
		resp[i] = toChatMessageResponse(m)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"messages": resp,
		"count":    len(resp),
		"limit":    limit,
		"offset":   offset,
	})
}

// GetQuestions handles GET /api/v1/streaming/sessions/{id}/chat/questions
func (h *ChatHandler) GetQuestions(w http.ResponseWriter, r *http.Request, sessionID string) {
	answeredOnly := r.URL.Query().Get("answered") == "true"

	questions, err := h.chatRepo.GetQuestions(r.Context(), sessionID, answeredOnly)
	if err != nil {
		h.logger.Printf("Failed to get questions: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get questions")
		return
	}

	resp := make([]*ChatMessageResponse, len(questions))
	for i, q := range questions {
		resp[i] = toChatMessageResponse(q)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"questions": resp,
		"count":     len(resp),
	})
}

// GetPinnedMessages handles GET /api/v1/streaming/sessions/{id}/chat/pinned
func (h *ChatHandler) GetPinnedMessages(w http.ResponseWriter, r *http.Request, sessionID string) {
	messages, err := h.chatRepo.GetPinnedMessages(r.Context(), sessionID)
	if err != nil {
		h.logger.Printf("Failed to get pinned messages: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to get pinned messages")
		return
	}

	resp := make([]*ChatMessageResponse, len(messages))
	for i, m := range messages {
		resp[i] = toChatMessageResponse(m)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"pinned_messages": resp,
		"count":           len(resp),
	})
}

// PinMessage handles POST /api/v1/streaming/sessions/{id}/chat/{messageId}/pin
func (h *ChatHandler) PinMessage(w http.ResponseWriter, r *http.Request, sessionID, messageID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := h.getUserIDFromRequest(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Only instructor can pin messages
	session, err := h.sessionRepo.GetSessionByID(r.Context(), sessionID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Session not found")
		return
	}

	if session.InstructorID != userID {
		respondError(w, http.StatusForbidden, "Only the instructor can pin messages")
		return
	}

	var req struct {
		Pinned bool `json:"pinned"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if err := h.chatRepo.PinMessage(r.Context(), messageID, req.Pinned); err != nil {
		h.logger.Printf("Failed to pin message: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to pin message")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"message_id": messageID,
		"pinned":     req.Pinned,
	})
}

// MarkAsAnswered handles POST /api/v1/streaming/sessions/{id}/chat/{messageId}/answer
func (h *ChatHandler) MarkAsAnswered(w http.ResponseWriter, r *http.Request, sessionID, messageID string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := h.getUserIDFromRequest(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Only instructor can mark as answered
	session, err := h.sessionRepo.GetSessionByID(r.Context(), sessionID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Session not found")
		return
	}

	if session.InstructorID != userID {
		respondError(w, http.StatusForbidden, "Only the instructor can mark questions as answered")
		return
	}

	if err := h.chatRepo.MarkAsAnswered(r.Context(), messageID); err != nil {
		h.logger.Printf("Failed to mark as answered: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to mark as answered")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"message_id": messageID,
		"answered":   true,
	})
}

func toChatMessageResponse(m *models.StreamChatMessage) *ChatMessageResponse {
	return &ChatMessageResponse{
		ID:          m.ID,
		SessionID:   m.SessionID,
		UserID:      m.UserID,
		MessageType: m.MessageType,
		Content:     m.Content,
		IsPinned:    m.IsPinned,
		IsAnswered:  m.IsAnswered,
		ReplyToID:   m.ReplyToID,
		CreatedAt:   m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
