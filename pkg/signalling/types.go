package signalling

// JoinRoomRequest is the payload for joining a room
type JoinRoomRequest struct {
	RoomID     string `json:"room_id"`
	RoomName   string `json:"room_name,omitempty"`
	UserID     string `json:"user_id"`
	Email      string `json:"email"`
	FullName   string `json:"full_name"`
	Role       string `json:"role"` // student, teacher, admin
	IsProducer bool   `json:"is_producer"`
}

// JoinRoomResponse is returned when user joins a room
type JoinRoomResponse struct {
	Success       bool                   `json:"success"`
	RoomID        string                 `json:"room_id"`
	ParticipantID string                 `json:"participant_id"`
	Participants  []*Participant         `json:"participants"`
	Message       string                 `json:"message,omitempty"`
	Mediasoup     *MediasoupJoinResponse `json:"mediasoup,omitempty"`
}

// LeaveRoomRequest is the payload for leaving a room
type LeaveRoomRequest struct {
	RoomID string `json:"room_id"`
	UserID string `json:"user_id"`
}

// GetParticipantsRequest is the payload for getting participants
type GetParticipantsRequest struct {
	RoomID string `json:"room_id"`
}

// GetParticipantsResponse returns list of participants
type GetParticipantsResponse struct {
	RoomID       string         `json:"room_id"`
	Participants []*Participant `json:"participants"`
	Count        int            `json:"count"`
}

// RoomStats contains statistics about a room
type RoomStats struct {
	RoomID           string         `json:"room_id"`
	ParticipantCount int            `json:"participant_count"`
	ProducerCount    int            `json:"producer_count"`
	Participants     []*Participant `json:"participants"`
}

// OfferRequest represents a WebRTC offer
type OfferRequest struct {
	To  string `json:"to"`
	SDP string `json:"sdp"`
}

// AnswerRequest represents a WebRTC answer
type AnswerRequest struct {
	To  string `json:"to"`
	SDP string `json:"sdp"`
}

// ICECandidateRequest represents an ICE candidate
type ICECandidateRequest struct {
	To        string `json:"to"`
	Candidate string `json:"candidate"`
}

// MediasoupPeerInfo represents a peer in Mediasoup response
type MediasoupPeerInfo struct {
	PeerID     string `json:"peerId"`
	UserID     string `json:"userId"`
	Email      string `json:"email"`
	FullName   string `json:"fullName"`
	Role       string `json:"role"`
	IsProducer bool   `json:"isProducer"`
}

// MediasoupJoinResponse contains Mediasoup-specific data for join response
type MediasoupJoinResponse struct {
	RtpCapabilities interface{}          `json:"rtpCapabilities"`
	Peers           []*MediasoupPeerInfo `json:"peers,omitempty"`
}
