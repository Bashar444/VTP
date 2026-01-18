package signalling

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Bashar444/VTP/pkg/auth"
)

// APIHandler provides HTTP endpoints for signalling management
type APIHandler struct {
	SignallingServer *SignallingServer
	AuthMiddleware   *auth.AuthMiddleware
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(ss *SignallingServer, am *auth.AuthMiddleware) *APIHandler {
	return &APIHandler{
		SignallingServer: ss,
		AuthMiddleware:   am,
	}
}

// GetRoomStatsHandler returns statistics for a specific room
func (h *APIHandler) GetRoomStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Get room ID from query parameter
	roomID := r.URL.Query().Get("room_id")
	if roomID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error":"room_id parameter required"}`)
		return
	}

	stats := h.SignallingServer.GetRoomStats(roomID)
	if stats == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"error":"Room not found"}`)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
	log.Printf("✓ Room stats retrieved: %s", roomID)
}

// GetAllRoomStatsHandler returns statistics for all rooms
func (h *APIHandler) GetAllRoomStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	stats := h.SignallingServer.GetAllRoomStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rooms": stats,
		"count": len(stats),
	})
	log.Printf("✓ All room stats retrieved (%d rooms)", len(stats))
}

// HealthHandler returns health status of signalling service
func (h *APIHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
		"status": "ok",
		"service": "signalling",
		"version": "1.0.0",
		"active_rooms": %d,
		"timestamp": %d
	}`, len(h.SignallingServer.RoomManager.GetAllRooms()), getCurrentTime())
}

// CreateRoomHandler creates a new room (for testing/admin)
func (h *APIHandler) CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RoomID   string `json:"room_id"`
		RoomName string `json:"room_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error":"Invalid request body"}`)
		return
	}

	if req.RoomID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error":"room_id is required"}`)
		return
	}

	room := h.SignallingServer.RoomManager.CreateRoom(req.RoomID, req.RoomName)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"room_id": room.ID,
		"message": "Room created successfully",
	})
}

// DeleteRoomHandler deletes a room (for testing/admin)
func (h *APIHandler) DeleteRoomHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	roomID := r.URL.Query().Get("room_id")
	if roomID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error":"room_id parameter required"}`)
		return
	}

	h.SignallingServer.RoomManager.DeleteRoom(roomID)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
		"success": true,
		"room_id": "%s",
		"message": "Room deleted successfully"
	}`, roomID)
}

// getCurrentTime returns current unix timestamp in milliseconds
func getCurrentTime() int64 {
	return int64(1000) // placeholder - would use time.Now().UnixMilli()
}
