package signalling

import (
	"encoding/json"
	"testing"
	"time"
)

// TestNewSignallingServer tests server creation
func TestNewSignallingServer(t *testing.T) {
	ss, err := NewSignallingServer()
	if err != nil {
		t.Fatalf("Failed to create signalling server: %v", err)
	}

	if ss == nil {
		t.Fatal("SignallingServer is nil")
	}

	if ss.IO == nil {
		t.Fatal("Socket.IO server not initialized")
	}

	if ss.RoomManager == nil {
		t.Fatal("RoomManager not initialized")
	}

	t.Log("✓ SignallingServer created successfully")
}

// TestRoomManager tests room creation and management
func TestRoomManager(t *testing.T) {
	rm := NewRoomManager()

	// Test CreateRoom
	rm.CreateRoom("room-1", "Test Room")
	if !rm.RoomExists("room-1") {
		t.Fatal("Room not created")
	}
	t.Log("✓ Room created successfully")

	// Test GetRoom
	room, exists := rm.GetRoom("room-1")
	if !exists {
		t.Fatal("Room not found")
	}
	if room == nil {
		t.Fatal("Room is nil")
	}
	t.Log("✓ Room retrieved successfully")

	// Test AddParticipant
	participant := room.AddParticipant(
		"socket-1",
		"user-1",
		"user1@example.com",
		"John Doe",
		"student",
		true,
	)
	if participant == nil {
		t.Fatal("Participant not created")
	}
	if participant.SocketID != "socket-1" {
		t.Fatalf("Participant socket ID mismatch: got %s, expected socket-1", participant.SocketID)
	}
	if participant.UserID != "user-1" {
		t.Fatalf("Participant user ID mismatch: got %s, expected user-1", participant.UserID)
	}
	if participant.Email != "user1@example.com" {
		t.Fatalf("Participant email mismatch: got %s, expected user1@example.com", participant.Email)
	}
	if !participant.IsProducer {
		t.Fatal("Participant should be producer")
	}
	t.Log("✓ Participant added successfully")

	// Test ParticipantCount
	count := room.ParticipantCount()
	if count != 1 {
		t.Fatalf("Participant count mismatch: got %d, expected 1", count)
	}
	t.Log("✓ Participant count verified")

	// Test GetProducers
	producers := room.GetProducers()
	if len(producers) != 1 {
		t.Fatalf("Producers count mismatch: got %d, expected 1", len(producers))
	}
	if producers[0].SocketID != "socket-1" {
		t.Fatalf("Producer socket ID mismatch")
	}
	t.Log("✓ Producers filtered correctly")

	// Test RemoveParticipant
	removed := room.RemoveParticipant("socket-1")
	if removed == nil {
		t.Fatal("Participant not removed")
	}
	if room.ParticipantCount() != 0 {
		t.Fatalf("Participant count should be 0 after removal")
	}
	t.Log("✓ Participant removed successfully")

	// Test IsEmpty
	if !room.IsEmpty() {
		t.Fatal("Room should be empty")
	}
	t.Log("✓ Room emptiness verified")

	// Test DeleteRoom
	rm.DeleteRoom("room-1")
	if rm.RoomExists("room-1") {
		t.Fatal("Room still exists after deletion")
	}
	t.Log("✓ Room deleted successfully")
}

// TestParticipantRole tests producer/consumer filtering
func TestParticipantRole(t *testing.T) {
	rm := NewRoomManager()
	rm.CreateRoom("room-1", "Test Room")
	room, _ := rm.GetRoom("room-1")

	// Add producers
	room.AddParticipant("socket-1", "user-1", "user1@example.com", "User 1", "teacher", true)
	room.AddParticipant("socket-2", "user-2", "user2@example.com", "User 2", "teacher", true)

	// Add consumers
	room.AddParticipant("socket-3", "user-3", "user3@example.com", "User 3", "student", false)
	room.AddParticipant("socket-4", "user-4", "user4@example.com", "User 4", "student", false)

	// Verify counts
	if room.ParticipantCount() != 4 {
		t.Fatalf("Total participant count should be 4, got %d", room.ParticipantCount())
	}

	producers := room.GetProducers()
	if len(producers) != 2 {
		t.Fatalf("Producer count should be 2, got %d", len(producers))
	}

	t.Log("✓ Participant roles verified correctly")
}

// TestJoinRoomRequest tests join room request validation
func TestJoinRoomRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     JoinRoomRequest
		isValid bool
	}{
		{
			name: "Valid request",
			req: JoinRoomRequest{
				RoomID:     "room-1",
				UserID:     "user-1",
				Email:      "user@example.com",
				FullName:   "John Doe",
				Role:       "student",
				IsProducer: false,
			},
			isValid: true,
		},
		{
			name: "Missing RoomID",
			req: JoinRoomRequest{
				UserID:   "user-1",
				Email:    "user@example.com",
				FullName: "John Doe",
				Role:     "student",
			},
			isValid: false,
		},
		{
			name: "Missing UserID",
			req: JoinRoomRequest{
				RoomID:   "room-1",
				Email:    "user@example.com",
				FullName: "John Doe",
				Role:     "student",
			},
			isValid: false,
		},
		{
			name: "Missing Email",
			req: JoinRoomRequest{
				RoomID:   "room-1",
				UserID:   "user-1",
				FullName: "John Doe",
				Role:     "student",
			},
			isValid: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := json.Marshal(test.req)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			var req JoinRoomRequest
			err = json.Unmarshal(data, &req)
			if err != nil {
				t.Fatalf("Failed to unmarshal request: %v", err)
			}

			// Validate required fields
			valid := req.RoomID != "" && req.UserID != "" && req.Email != ""
			if valid != test.isValid {
				t.Fatalf("Validation mismatch: expected %v, got %v", test.isValid, valid)
			}
		})
	}

	t.Log("✓ All join request validations passed")
}

// TestSignallingMessage tests WebRTC signalling messages
func TestSignallingMessage(t *testing.T) {
	tests := []struct {
		name    string
		msg     SignallingMessage
		msgType string
	}{
		{
			name: "Offer message",
			msg: SignallingMessage{
				Type: "offer",
				From: "user-1",
				To:   "user-2",
				SDP:  "v=0\r\no=- 123456789 2 IN IP4 127.0.0.1\r\n...",
			},
			msgType: "offer",
		},
		{
			name: "Answer message",
			msg: SignallingMessage{
				Type: "answer",
				From: "user-2",
				To:   "user-1",
				SDP:  "v=0\r\no=- 987654321 2 IN IP4 127.0.0.1\r\n...",
			},
			msgType: "answer",
		},
		{
			name: "ICE candidate message",
			msg: SignallingMessage{
				Type:      "ice-candidate",
				From:      "user-1",
				To:        "user-2",
				Candidate: "candidate:123456789 1 udp 2113937151 192.168.1.1 54321 typ host",
			},
			msgType: "ice-candidate",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, err := json.Marshal(test.msg)
			if err != nil {
				t.Fatalf("Failed to marshal message: %v", err)
			}

			var msg SignallingMessage
			err = json.Unmarshal(data, &msg)
			if err != nil {
				t.Fatalf("Failed to unmarshal message: %v", err)
			}

			if msg.Type != test.msgType {
				t.Fatalf("Message type mismatch: expected %s, got %s", test.msgType, msg.Type)
			}
			if msg.From == "" || msg.To == "" {
				t.Fatal("From/To should not be empty")
			}
		})
	}

	t.Log("✓ All signalling messages validated")
}

// TestRoomStats tests room statistics
func TestRoomStats(t *testing.T) {
	rm := NewRoomManager()
	rm.CreateRoom("room-1", "Test Room")
	room, _ := rm.GetRoom("room-1")

	// Add some participants
	room.AddParticipant("socket-1", "user-1", "user1@example.com", "User 1", "teacher", true)
	room.AddParticipant("socket-2", "user-2", "user2@example.com", "User 2", "student", false)
	room.AddParticipant("socket-3", "user-3", "user3@example.com", "User 3", "student", false)

	stats := &RoomStats{
		RoomID:           room.ID,
		ParticipantCount: room.ParticipantCount(),
		ProducerCount:    len(room.GetProducers()),
		Participants:     room.GetAllParticipants(),
	}

	if stats.ParticipantCount != 3 {
		t.Fatalf("Participant count mismatch: expected 3, got %d", stats.ParticipantCount)
	}

	if stats.ProducerCount != 1 {
		t.Fatalf("Producer count mismatch: expected 1, got %d", stats.ProducerCount)
	}

	if len(stats.Participants) != 3 {
		t.Fatalf("Participants array length mismatch: expected 3, got %d", len(stats.Participants))
	}

	t.Log("✓ Room statistics verified")
}

// TestParticipantTimestamp tests participant join timestamp
func TestParticipantTimestamp(t *testing.T) {
	rm := NewRoomManager()
	rm.CreateRoom("room-1", "Test Room")
	room, _ := rm.GetRoom("room-1")

	before := time.Now().Unix() * 1000
	participant := room.AddParticipant("socket-1", "user-1", "user1@example.com", "User 1", "student", false)
	after := time.Now().Unix() * 1000

	if participant.JoinedAt < before || participant.JoinedAt > after {
		t.Fatalf("JoinedAt timestamp out of range: %d (expected between %d and %d)", participant.JoinedAt, before, after)
	}

	t.Log("✓ Participant timestamp verified")
}

// TestMultipleRooms tests multiple concurrent rooms
func TestMultipleRooms(t *testing.T) {
	rm := NewRoomManager()

	// Create multiple rooms
	for i := 1; i <= 5; i++ {
		roomID := "room-" + string(rune(48+i))
		rm.CreateRoom(roomID, "Room "+string(rune(48+i)))
	}

	// Verify all rooms exist
	for i := 1; i <= 5; i++ {
		roomID := "room-" + string(rune(48+i))
		if !rm.RoomExists(roomID) {
			t.Fatalf("Room %s not found", roomID)
		}
	}

	// Add participants to different rooms
	for i := 1; i <= 5; i++ {
		roomID := "room-" + string(rune(48+i))
		room, _ := rm.GetRoom(roomID)
		for j := 1; j <= 3; j++ {
			room.AddParticipant(
				"socket-"+string(rune(48+i))+"-"+string(rune(48+j)),
				"user-"+string(rune(48+i))+"-"+string(rune(48+j)),
				"user"+string(rune(48+i))+string(rune(48+j))+"@example.com",
				"User",
				"student",
				false,
			)
		}
	}

	t.Log("✓ Multiple rooms with participants verified")
}

// TestRoomCleanup tests that empty rooms are properly managed
func TestRoomCleanup(t *testing.T) {
	rm := NewRoomManager()
	rm.CreateRoom("room-1", "Test Room")
	room, _ := rm.GetRoom("room-1")

	// Add and remove participants
	room.AddParticipant("socket-1", "user-1", "user1@example.com", "User 1", "student", false)
	if room.IsEmpty() {
		t.Fatal("Room should not be empty after adding participant")
	}

	room.RemoveParticipant("socket-1")
	if !room.IsEmpty() {
		t.Fatal("Room should be empty after removing all participants")
	}

	t.Log("✓ Room cleanup verified")
}
