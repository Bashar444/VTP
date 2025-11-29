package signalling

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/yourusername/vtp-platform/pkg/auth"
	"github.com/yourusername/vtp-platform/pkg/g5"
)

// APIHandler provides HTTP endpoints for signalling management
type APIHandler struct {
	SignallingServer *SignallingServer
	AuthMiddleware   *auth.AuthMiddleware
	G5Adapter        *g5.Adapter
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

// ============= 5G Network Endpoints =============

// GetNetworkStatusHandler returns current 5G network status
func (h *APIHandler) GetNetworkStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if h.G5Adapter == nil || !h.G5Adapter.IsStarted() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"error":"5G adapter not available"}`)
		return
	}

	network := h.G5Adapter.GetCurrentNetwork()
	if network == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"type":            network.Type,
		"latency_ms":      network.Latency,
		"bandwidth_kbps":  network.Bandwidth,
		"signal_strength": network.SignalStrength,
		"is_5g_available": h.G5Adapter.Is5GAvailable(),
		"quality_score":   h.G5Adapter.GetNetworkQuality(),
		"timestamp":       getCurrentTime(),
	})
	log.Printf("✓ Network status retrieved")
}

// GetNetworkMetricsHandler returns current network metrics
func (h *APIHandler) GetNetworkMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if h.G5Adapter == nil || !h.G5Adapter.IsStarted() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"error":"5G adapter not available"}`)
		return
	}

	status := h.G5Adapter.GetStatus()
	metrics := h.G5Adapter.GetGlobalMetrics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"adapter_status":          status.IsHealthy,
		"active_sessions":         status.TotalActiveSessions,
		"current_quality":         status.CurrentQuality,
		"active_edge_node":        status.ActiveEdgeNode,
		"global_avg_latency_ms":   metrics.GlobalAvgLatency,
		"global_avg_bandwidth":    metrics.GlobalAvgBandwidth,
		"global_packet_loss":      metrics.GlobalAvgPacketLoss,
		"peak_concurrent":         metrics.PeakConcurrent,
		"total_bytes_transferred": metrics.TotalBytesTransfer,
		"timestamp":               getCurrentTime(),
	})
	log.Printf("✓ Network metrics retrieved")
}

// DetectNetworkHandler triggers network detection
func (h *APIHandler) DetectNetworkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if h.G5Adapter == nil || !h.G5Adapter.IsStarted() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"error":"5G adapter not available"}`)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*1000000000) // 10 seconds
	defer cancel()

	network, err := h.G5Adapter.DetectNetworkType(ctx)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error":"Network detection failed: %s"}`, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"type":            network.NetworkType,
		"latency_ms":      network.Latency,
		"bandwidth_kbps":  network.Bandwidth,
		"signal_strength": network.SignalStrength,
		"detected_at":     network.Timestamp,
	})
	log.Printf("✓ Network detection completed: %v", network.NetworkType)
}

// GetEdgeNodesHandler returns available edge nodes
func (h *APIHandler) GetEdgeNodesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if h.G5Adapter == nil || !h.G5Adapter.IsStarted() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"error":"5G adapter not available"}`)
		return
	}

	nodes := h.G5Adapter.GetAvailableEdgeNodes()

	type nodeResponse struct {
		ID        string  `json:"id"`
		Region    string  `json:"region"`
		Country   string  `json:"country"`
		Endpoint  string  `json:"endpoint"`
		Latency   int     `json:"latency_ms"`
		Status    string  `json:"status"`
		Load      float64 `json:"load_percentage"`
		Available int     `json:"available_capacity"`
		Capacity  int     `json:"max_capacity"`
	}

	response := make([]nodeResponse, 0, len(nodes))
	for _, node := range nodes {
		response = append(response, nodeResponse{
			ID:        node.ID,
			Region:    node.Region,
			Country:   node.Country,
			Endpoint:  node.Endpoint,
			Latency:   node.Latency,
			Status:    string(node.Status),
			Load:      node.Load,
			Available: node.Available,
			Capacity:  node.Capacity,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"nodes": response,
		"count": len(response),
	})
	log.Printf("✓ Edge nodes retrieved: %d available", len(response))
}

// ConnectToEdgeHandler initiates connection to an edge node
func (h *APIHandler) ConnectToEdgeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if h.G5Adapter == nil || !h.G5Adapter.IsStarted() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"error":"5G adapter not available"}`)
		return
	}

	var req struct {
		SessionID string `json:"session_id"`
		NodeID    string `json:"node_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error":"Invalid request body"}`)
		return
	}

	if req.SessionID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error":"session_id is required"}`)
		return
	}

	// Start session if node_id not provided (auto-select best node)
	if req.NodeID == "" {
		err := h.G5Adapter.StartSession(req.SessionID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error":"Failed to start session: %s"}`, err.Error())
			return
		}
	} else {
		// Manual node selection would be handled by application
		err := h.G5Adapter.StartSession(req.SessionID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error":"Failed to start session: %s"}`, err.Error())
			return
		}
	}

	status := h.G5Adapter.GetStatus()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":         true,
		"session_id":      req.SessionID,
		"edge_node_id":    status.ActiveEdgeNode,
		"current_quality": status.CurrentQuality,
		"message":         "Connected to edge node",
	})
	log.Printf("✓ Session started: %s on edge node", req.SessionID)
}

// GetSessionMetricsHandler returns metrics for current session
func (h *APIHandler) GetSessionMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if h.G5Adapter == nil || !h.G5Adapter.IsStarted() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"error":"5G adapter not available"}`)
		return
	}

	metrics := h.G5Adapter.GetSessionMetrics()
	if metrics == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id":         metrics.SessionID,
		"duration_ms":        metrics.Duration,
		"avg_latency_ms":     metrics.AvgLatency,
		"min_latency_ms":     metrics.MinLatency,
		"max_latency_ms":     metrics.MaxLatency,
		"avg_bandwidth_kbps": metrics.AvgBandwidth,
		"avg_packet_loss":    metrics.AvgPacketLoss,
		"avg_jitter_ms":      metrics.AvgJitter,
		"video_quality":      metrics.VideoQuality,
		"frames_dropped":     metrics.FramesDropped,
		"bytes_sent":         metrics.BytesSent,
		"bytes_received":     metrics.BytesReceived,
		"sample_count":       metrics.SampleCount,
	})
	log.Printf("✓ Session metrics retrieved: %s", metrics.SessionID)
}

// AdaptQualityHandler triggers quality adaptation
func (h *APIHandler) AdaptQualityHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if h.G5Adapter == nil || !h.G5Adapter.IsStarted() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"error":"5G adapter not available"}`)
		return
	}

	quality, err := h.G5Adapter.AdaptQuality()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error":"Quality adaptation failed: %s"}`, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"new_quality": quality,
		"adapted_at":  getCurrentTime(),
	})
	log.Printf("✓ Quality adapted to: %s", quality)
}

// GetAdapterStatusHandler returns full adapter status
func (h *APIHandler) GetAdapterStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if h.G5Adapter == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"error":"5G adapter not available"}`)
		return
	}

	status := h.G5Adapter.GetStatus()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"is_started":            status.IsStarted,
		"is_healthy":            status.IsHealthy,
		"current_quality":       status.CurrentQuality,
		"active_session_id":     status.ActiveSessionID,
		"total_active_sessions": status.TotalActiveSessions,
		"detector_status":       status.DetectorStatus,
		"edge_manager_status":   status.EdgeManagerStatus,
		"last_update":           getCurrentTime(),
	})
	log.Printf("✓ Adapter status retrieved")
}
