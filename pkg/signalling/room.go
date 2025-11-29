package signalling

import (
	"log"
	"sync"
	"time"
)

// Room represents a live session/classroom
type Room struct {
	ID          string
	Name        string
	Participants map[string]*Participant // key: socket ID
	mu          sync.RWMutex
}

// Participant represents a user in a room
type Participant struct {
	SocketID  string
	UserID    string
	Email     string
	FullName  string
	Role      string // student, teacher, admin
	IsProducer bool   // Can send media
	IsConsumer bool   // Receiving media
	JoinedAt  int64  // Timestamp
}

// Message types for WebRTC signalling
type SignallingMessage struct {
	Type      string      `json:"type"` // offer, answer, ice-candidate
	From      string      `json:"from"`
	To        string      `json:"to"`
	SDP       string      `json:"sdp,omitempty"`       // For offer/answer
	Candidate string      `json:"candidate,omitempty"` // For ICE
	Data      interface{} `json:"data,omitempty"`
}

// RoomManager manages all active rooms
type RoomManager struct {
	Rooms map[string]*Room
	mu    sync.RWMutex
}

// NewRoomManager creates a new room manager
func NewRoomManager() *RoomManager {
	return &RoomManager{
		Rooms: make(map[string]*Room),
	}
}

// CreateRoom creates a new room
func (rm *RoomManager) CreateRoom(roomID, roomName string) *Room {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room := &Room{
		ID:           roomID,
		Name:         roomName,
		Participants: make(map[string]*Participant),
	}

	rm.Rooms[roomID] = room
	log.Printf("✓ Room created: %s (%s)", roomID, roomName)
	return room
}

// GetRoom retrieves a room by ID
func (rm *RoomManager) GetRoom(roomID string) (*Room, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, exists := rm.Rooms[roomID]
	return room, exists
}

// DeleteRoom removes a room
func (rm *RoomManager) DeleteRoom(roomID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	delete(rm.Rooms, roomID)
	log.Printf("✓ Room deleted: %s", roomID)
}

// GetAllRooms returns all active rooms
func (rm *RoomManager) GetAllRooms() []*Room {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rooms := make([]*Room, 0, len(rm.Rooms))
	for _, room := range rm.Rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// RoomExists checks if a room exists
func (rm *RoomManager) RoomExists(roomID string) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	_, exists := rm.Rooms[roomID]
	return exists
}

// AddParticipant adds a user to a room
func (r *Room) AddParticipant(socketID, userID, email, fullName, role string, isProducer bool) *Participant {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().Unix() * 1000 // milliseconds

	participant := &Participant{
		SocketID:   socketID,
		UserID:     userID,
		Email:      email,
		FullName:   fullName,
		Role:       role,
		IsProducer: isProducer,
		IsConsumer: true, // All participants can receive
		JoinedAt:   now,
	}

	r.Participants[socketID] = participant
	log.Printf("✓ Participant joined room %s: %s (%s) - Producer: %v", r.ID, fullName, email, isProducer)

	return participant
}

// RemoveParticipant removes a user from a room
func (r *Room) RemoveParticipant(socketID string) *Participant {
	r.mu.Lock()
	defer r.mu.Unlock()

	participant, exists := r.Participants[socketID]
	if !exists {
		return nil
	}

	delete(r.Participants, socketID)
	log.Printf("✓ Participant left room %s: %s (%s)", r.ID, participant.FullName, participant.Email)

	return participant
}

// GetParticipant retrieves a participant by socket ID
func (r *Room) GetParticipant(socketID string) (*Participant, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	participant, exists := r.Participants[socketID]
	return participant, exists
}

// GetAllParticipants returns all participants in the room
func (r *Room) GetAllParticipants() []*Participant {
	r.mu.RLock()
	defer r.mu.RUnlock()

	participants := make([]*Participant, 0, len(r.Participants))
	for _, p := range r.Participants {
		participants = append(participants, p)
	}
	return participants
}

// GetProducers returns all producers (media senders) in the room
func (r *Room) GetProducers() []*Participant {
	r.mu.RLock()
	defer r.mu.RUnlock()

	producers := make([]*Participant, 0)
	for _, p := range r.Participants {
		if p.IsProducer {
			producers = append(producers, p)
		}
	}
	return producers
}

// ParticipantCount returns the number of participants in the room
func (r *Room) ParticipantCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.Participants)
}

// IsEmpty checks if room has no participants
func (r *Room) IsEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.Participants) == 0
}
