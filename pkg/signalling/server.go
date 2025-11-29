package signalling

import (
	"encoding/json"
	"log"
	"net/http"

	socketio "github.com/googollee/go-socket.io"
	"github.com/yourusername/vtp-platform/pkg/mediasoup"
)

// SignallingServer manages WebRTC signalling via Socket.IO
type SignallingServer struct {
	IO          *socketio.Server
	RoomManager *RoomManager
	Mediasoup   *MediasoupIntegration
}

// NewSignallingServer creates a new signalling server
func NewSignallingServer() (*SignallingServer, error) {
	server := socketio.NewServer(nil)

	ss := &SignallingServer{
		IO:          server,
		RoomManager: NewRoomManager(),
		Mediasoup:   NewMediasoupIntegration("http://localhost:3000"),
	}

	ss.registerEventHandlers()

	return ss, nil
}

// NewSignallingServerWithMediasoup creates a signalling server with custom Mediasoup URL
func NewSignallingServerWithMediasoup(mediasoupURL string) (*SignallingServer, error) {
	server := socketio.NewServer(nil)

	ss := &SignallingServer{
		IO:          server,
		RoomManager: NewRoomManager(),
		Mediasoup:   NewMediasoupIntegration(mediasoupURL),
	}

	ss.registerEventHandlers()
	ss.RegisterMediasoupHandlers(ss.Mediasoup)

	return ss, nil
}

// registerEventHandlers sets up all Socket.IO event handlers
func (ss *SignallingServer) registerEventHandlers() {
	ss.IO.OnConnect("", func(s socketio.Conn) error {
		log.Printf("✓ Socket connected: %s", s.ID())
		return nil
	})

	ss.IO.OnDisconnect("", func(s socketio.Conn, reason string) {
		log.Printf("✗ Socket disconnected: %s (reason: %s)", s.ID(), reason)
	})

	ss.IO.OnError("", func(s socketio.Conn, e error) {
		log.Printf("❌ Socket error: %v", e)
	})

	ss.IO.OnEvent("", "join-room", func(s socketio.Conn, payload string) {
		var req JoinRoomRequest
		if err := json.Unmarshal([]byte(payload), &req); err != nil {
			s.Emit("error", map[string]string{"error": "Invalid payload"})
			return
		}

		if req.RoomID == "" || req.UserID == "" || req.Email == "" {
			s.Emit("error", map[string]string{"error": "Missing required fields"})
			return
		}

		if !ss.RoomManager.RoomExists(req.RoomID) {
			ss.RoomManager.CreateRoom(req.RoomID, req.RoomName)
		}

		room, _ := ss.RoomManager.GetRoom(req.RoomID)
		_ = room.AddParticipant(
			s.ID(),
			req.UserID,
			req.Email,
			req.FullName,
			req.Role,
			req.IsProducer,
		)

		s.Join(req.RoomID)

		// Integrate with Mediasoup if available
		var mediasoupResp *MediasoupJoinResponse
		if ss.Mediasoup != nil {
			resp, err := ss.Mediasoup.OnJoinRoom(
				req.RoomID,
				s.ID(),
				req.UserID,
				req.Email,
				req.FullName,
				req.Role,
				req.IsProducer,
			)
			if err != nil {
				log.Printf("⚠ Mediasoup join failed: %v", err)
			} else {
				mediasoupResp = &MediasoupJoinResponse{
					RtpCapabilities: resp.RtpCapabilities,
					Peers:           convertMediasoupPeers(resp.Peers),
				}
			}
		}

		response := JoinRoomResponse{
			Success:       true,
			RoomID:        req.RoomID,
			ParticipantID: s.ID(),
			Participants:  room.GetAllParticipants(),
			Mediasoup:     mediasoupResp,
		}
		s.Emit("joined-room", response)

		log.Printf("✓ User %s joined room %s", req.Email, req.RoomID)
	})

	ss.IO.OnEvent("", "leave-room", func(s socketio.Conn, payload string) {
		var req LeaveRoomRequest
		if err := json.Unmarshal([]byte(payload), &req); err != nil {
			s.Emit("error", map[string]string{"error": "Invalid payload"})
			return
		}

		room, exists := ss.RoomManager.GetRoom(req.RoomID)
		if !exists {
			s.Emit("error", map[string]string{"error": "Room not found"})
			return
		}

		room.RemoveParticipant(s.ID())
		s.Leave(req.RoomID)

		// Integrate with Mediasoup if available
		if ss.Mediasoup != nil {
			if err := ss.Mediasoup.OnLeaveRoom(req.RoomID, s.ID()); err != nil {
				log.Printf("⚠ Mediasoup leave failed: %v", err)
			}
		}

		if room.IsEmpty() {
			ss.RoomManager.DeleteRoom(req.RoomID)
		}

		log.Printf("✓ User left room %s", req.RoomID)
	})

	ss.IO.OnEvent("", "webrtc-offer", func(s socketio.Conn, payload string) {
		var msg SignallingMessage
		if err := json.Unmarshal([]byte(payload), &msg); err != nil {
			s.Emit("error", map[string]string{"error": "Invalid payload"})
			return
		}

		log.Printf("✓ Offer from %s to %s (SDP length: %d)", msg.From, msg.To, len(msg.SDP))
		s.Emit("webrtc-offer", msg)
	})

	ss.IO.OnEvent("", "webrtc-answer", func(s socketio.Conn, payload string) {
		var msg SignallingMessage
		if err := json.Unmarshal([]byte(payload), &msg); err != nil {
			s.Emit("error", map[string]string{"error": "Invalid payload"})
			return
		}

		log.Printf("✓ Answer from %s to %s (SDP length: %d)", msg.From, msg.To, len(msg.SDP))
		s.Emit("webrtc-answer", msg)
	})

	ss.IO.OnEvent("", "ice-candidate", func(s socketio.Conn, payload string) {
		var msg SignallingMessage
		if err := json.Unmarshal([]byte(payload), &msg); err != nil {
			s.Emit("error", map[string]string{"error": "Invalid payload"})
			return
		}

		log.Printf("✓ ICE candidate from %s to %s", msg.From, msg.To)
		s.Emit("ice-candidate", msg)
	})

	ss.IO.OnEvent("", "get-participants", func(s socketio.Conn, payload string) {
		var req GetParticipantsRequest
		if err := json.Unmarshal([]byte(payload), &req); err != nil {
			s.Emit("error", map[string]string{"error": "Invalid payload"})
			return
		}

		room, exists := ss.RoomManager.GetRoom(req.RoomID)
		if !exists {
			s.Emit("error", map[string]string{"error": "Room not found"})
			return
		}

		response := GetParticipantsResponse{
			RoomID:       req.RoomID,
			Participants: room.GetAllParticipants(),
			Count:        room.ParticipantCount(),
		}

		s.Emit("participants-list", response)
	})
}

// convertMediasoupPeers converts Mediasoup peers to response format
func convertMediasoupPeers(peers []mediasoup.Peer) []*MediasoupPeerInfo {
	result := make([]*MediasoupPeerInfo, len(peers))
	for i, p := range peers {
		result[i] = &MediasoupPeerInfo{
			PeerID:     p.ID,
			UserID:     p.UserID,
			Email:      p.Email,
			FullName:   p.FullName,
			Role:       p.Role,
			IsProducer: p.IsProducer,
		}
	}
	return result
}

// ServeHTTP implements http.Handler for Socket.IO
func (ss *SignallingServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ss.IO.ServeHTTP(w, r)
}

// GetRoomStats returns statistics about a room
func (ss *SignallingServer) GetRoomStats(roomID string) *RoomStats {
	room, exists := ss.RoomManager.GetRoom(roomID)
	if !exists {
		return nil
	}

	return &RoomStats{
		RoomID:           roomID,
		ParticipantCount: room.ParticipantCount(),
		ProducerCount:    len(room.GetProducers()),
		Participants:     room.GetAllParticipants(),
	}
}

// GetAllRoomStats returns statistics for all rooms
func (ss *SignallingServer) GetAllRoomStats() []*RoomStats {
	rooms := ss.RoomManager.GetAllRooms()
	stats := make([]*RoomStats, 0, len(rooms))

	for _, room := range rooms {
		stats = append(stats, &RoomStats{
			RoomID:           room.ID,
			ParticipantCount: room.ParticipantCount(),
			ProducerCount:    len(room.GetProducers()),
			Participants:     room.GetAllParticipants(),
		})
	}

	return stats
}
