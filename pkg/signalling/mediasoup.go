package signalling

import (
	"encoding/json"
	"fmt"
	"log"

	socketio "github.com/googollee/go-socket.io"
	"github.com/yourusername/vtp-platform/pkg/mediasoup"
)

// MediasoupIntegration handles integration between signalling and Mediasoup
type MediasoupIntegration struct {
	MediasoupClient *mediasoup.Client
	RoomTransports  map[string]map[string]*mediasoup.Transport // roomID -> peerId -> transport
	RoomProducers   map[string]map[string]*mediasoup.Producer  // roomID -> peerId -> producer
	RoomConsumers   map[string]map[string]*mediasoup.Consumer  // roomID -> peerId -> consumer
}

// NewMediasoupIntegration creates a new Mediasoup integration handler
func NewMediasoupIntegration(mediasoupURL string) *MediasoupIntegration {
	return &MediasoupIntegration{
		MediasoupClient: mediasoup.NewClient(mediasoupURL),
		RoomTransports:  make(map[string]map[string]*mediasoup.Transport),
		RoomProducers:   make(map[string]map[string]*mediasoup.Producer),
		RoomConsumers:   make(map[string]map[string]*mediasoup.Consumer),
	}
}

// OnJoinRoom handles Mediasoup integration when a peer joins
func (mi *MediasoupIntegration) OnJoinRoom(roomID, peerId, userId, email, fullName, role string, isProducer bool) (*mediasoup.JoinRoomResponse, error) {
	// Create room in Mediasoup
	joinReq := &mediasoup.JoinRoomRequest{
		PeerID:     peerId,
		UserID:     userId,
		Email:      email,
		FullName:   fullName,
		Role:       role,
		IsProducer: isProducer,
		RoomName:   roomID,
	}

	resp, err := mi.MediasoupClient.JoinRoom(roomID, joinReq)
	if err != nil {
		log.Printf("❌ Failed to join room in Mediasoup: %v", err)
		return nil, err
	}

	// Initialize storage for this room if needed
	if _, exists := mi.RoomTransports[roomID]; !exists {
		mi.RoomTransports[roomID] = make(map[string]*mediasoup.Transport)
		mi.RoomProducers[roomID] = make(map[string]*mediasoup.Producer)
		mi.RoomConsumers[roomID] = make(map[string]*mediasoup.Consumer)
	}

	log.Printf("✓ Peer %s joined room %s in Mediasoup", peerId, roomID)
	return resp, nil
}

// OnLeaveRoom handles cleanup when a peer leaves
func (mi *MediasoupIntegration) OnLeaveRoom(roomID, peerId string) error {
	// Close all producers for this peer
	if producers, exists := mi.RoomProducers[roomID]; exists {
		if producer, hasProducer := producers[peerId]; hasProducer {
			if err := mi.MediasoupClient.CloseProducer(roomID, producer.ID); err != nil {
				log.Printf("⚠ Failed to close producer: %v", err)
			}
			delete(producers, peerId)
		}
	}

	// Close all consumers for this peer
	if consumers, exists := mi.RoomConsumers[roomID]; exists {
		if consumer, hasConsumer := consumers[peerId]; hasConsumer {
			if err := mi.MediasoupClient.CloseConsumer(roomID, consumer.ID); err != nil {
				log.Printf("⚠ Failed to close consumer: %v", err)
			}
			delete(consumers, peerId)
		}
	}

	// Close transport
	if transports, exists := mi.RoomTransports[roomID]; exists {
		delete(transports, peerId)
	}

	// Remove peer from room in Mediasoup
	err := mi.MediasoupClient.LeaveRoom(roomID, peerId)
	if err != nil {
		log.Printf("❌ Failed to leave room in Mediasoup: %v", err)
		return err
	}

	// Clean up room if empty
	room, err := mi.MediasoupClient.GetRoom(roomID)
	if err == nil && room.PeerCount == 0 {
		delete(mi.RoomTransports, roomID)
		delete(mi.RoomProducers, roomID)
		delete(mi.RoomConsumers, roomID)
	}

	log.Printf("✓ Peer %s left room %s in Mediasoup", peerId, roomID)
	return nil
}

// CreateTransport creates a WebRTC transport in Mediasoup
func (mi *MediasoupIntegration) CreateTransport(roomID, peerId, direction string) (*mediasoup.Transport, error) {
	transport, err := mi.MediasoupClient.CreateTransport(roomID, peerId, direction)
	if err != nil {
		log.Printf("❌ Failed to create transport: %v", err)
		return nil, err
	}

	// Store transport reference
	if _, exists := mi.RoomTransports[roomID]; !exists {
		mi.RoomTransports[roomID] = make(map[string]*mediasoup.Transport)
	}
	mi.RoomTransports[roomID][peerId] = transport

	log.Printf("✓ Created transport %s for peer %s", transport.TransportID, peerId)
	return transport, nil
}

// ConnectTransport connects a transport with DTLS parameters
func (mi *MediasoupIntegration) ConnectTransport(roomID, peerId, transportID string, dtlsParams mediasoup.DtlsParameters) error {
	err := mi.MediasoupClient.ConnectTransport(roomID, transportID, dtlsParams)
	if err != nil {
		log.Printf("❌ Failed to connect transport: %v", err)
		return err
	}

	log.Printf("✓ Connected transport %s", transportID)
	return nil
}

// CreateProducer creates a media producer in Mediasoup
func (mi *MediasoupIntegration) CreateProducer(roomID, peerId string, kind string, rtpParams interface{}) (*mediasoup.Producer, error) {
	req := &mediasoup.ProducerRequest{
		PeerID:        peerId,
		Kind:          kind,
		RtpParameters: rtpParams,
	}

	producer, err := mi.MediasoupClient.CreateProducer(roomID, req)
	if err != nil {
		log.Printf("❌ Failed to create producer: %v", err)
		return nil, err
	}

	// Store producer reference
	if _, exists := mi.RoomProducers[roomID]; !exists {
		mi.RoomProducers[roomID] = make(map[string]*mediasoup.Producer)
	}
	mi.RoomProducers[roomID][peerId] = producer

	log.Printf("✓ Created %s producer %s for peer %s", kind, producer.ID, peerId)
	return producer, nil
}

// CreateConsumer creates a media consumer in Mediasoup
func (mi *MediasoupIntegration) CreateConsumer(roomID, peerId, producerID string, rtpCaps interface{}) (*mediasoup.Consumer, error) {
	req := &mediasoup.ConsumerRequest{
		PeerID:          peerId,
		ProducerID:      producerID,
		RtpCapabilities: rtpCaps,
	}

	consumer, err := mi.MediasoupClient.CreateConsumer(roomID, req)
	if err != nil {
		log.Printf("❌ Failed to create consumer: %v", err)
		return nil, err
	}

	// Store consumer reference
	if _, exists := mi.RoomConsumers[roomID]; !exists {
		mi.RoomConsumers[roomID] = make(map[string]*mediasoup.Consumer)
	}
	mi.RoomConsumers[roomID][peerId] = consumer

	log.Printf("✓ Created consumer %s for peer %s from producer %s", consumer.ID, peerId, producerID)
	return consumer, nil
}

// CloseProducer closes a producer
func (mi *MediasoupIntegration) CloseProducer(roomID, peerId, producerID string) error {
	err := mi.MediasoupClient.CloseProducer(roomID, producerID)
	if err != nil {
		log.Printf("❌ Failed to close producer: %v", err)
		return err
	}

	if producers, exists := mi.RoomProducers[roomID]; exists {
		delete(producers, peerId)
	}

	log.Printf("✓ Closed producer %s", producerID)
	return nil
}

// CloseConsumer closes a consumer
func (mi *MediasoupIntegration) CloseConsumer(roomID, peerId, consumerID string) error {
	err := mi.MediasoupClient.CloseConsumer(roomID, consumerID)
	if err != nil {
		log.Printf("❌ Failed to close consumer: %v", err)
		return err
	}

	if consumers, exists := mi.RoomConsumers[roomID]; exists {
		delete(consumers, peerId)
	}

	log.Printf("✓ Closed consumer %s", consumerID)
	return nil
}

// GetRoomInfo retrieves room information from Mediasoup
func (mi *MediasoupIntegration) GetRoomInfo(roomID string) (*mediasoup.RoomInfo, error) {
	return mi.MediasoupClient.GetRoom(roomID)
}

// RegisterMediasoupHandlers registers Mediasoup-related event handlers
func (ss *SignallingServer) RegisterMediasoupHandlers(mi *MediasoupIntegration) {
	// Transport events
	ss.IO.OnEvent("", "create-transport", func(s socketio.Conn, payload string) {
		var req struct {
			RoomID    string `json:"roomId"`
			Direction string `json:"direction"`
		}

		if err := json.Unmarshal([]byte(payload), &req); err != nil {
			s.Emit("error", map[string]string{"error": "Invalid payload"})
			return
		}

		transport, err := mi.CreateTransport(req.RoomID, s.ID(), req.Direction)
		if err != nil {
			s.Emit("error", map[string]string{"error": fmt.Sprintf("Failed to create transport: %v", err)})
			return
		}

		s.Emit("transport-created", transport)
		log.Printf("✓ Transport created for client %s", s.ID())
	})

	// Connect transport with DTLS
	ss.IO.OnEvent("", "connect-transport", func(s socketio.Conn, payload string) {
		var req struct {
			RoomID         string                   `json:"roomId"`
			TransportID    string                   `json:"transportId"`
			DtlsParameters mediasoup.DtlsParameters `json:"dtlsParameters"`
		}

		if err := json.Unmarshal([]byte(payload), &req); err != nil {
			s.Emit("error", map[string]string{"error": "Invalid payload"})
			return
		}

		err := mi.ConnectTransport(req.RoomID, s.ID(), req.TransportID, req.DtlsParameters)
		if err != nil {
			s.Emit("error", map[string]string{"error": fmt.Sprintf("Failed to connect transport: %v", err)})
			return
		}

		s.Emit("transport-connected", map[string]string{"transportId": req.TransportID})
		log.Printf("✓ Transport connected for client %s", s.ID())
	})

	// Produce media
	ss.IO.OnEvent("", "produce", func(s socketio.Conn, payload string) {
		var req struct {
			RoomID        string      `json:"roomId"`
			Kind          string      `json:"kind"`
			RtpParameters interface{} `json:"rtpParameters"`
		}

		if err := json.Unmarshal([]byte(payload), &req); err != nil {
			s.Emit("error", map[string]string{"error": "Invalid payload"})
			return
		}

		producer, err := mi.CreateProducer(req.RoomID, s.ID(), req.Kind, req.RtpParameters)
		if err != nil {
			s.Emit("error", map[string]string{"error": fmt.Sprintf("Failed to produce: %v", err)})
			return
		}

		s.Emit("producer-created", producer)
		log.Printf("✓ Producer created for client %s (kind: %s)", s.ID(), req.Kind)
	})

	// Consume media
	ss.IO.OnEvent("", "consume", func(s socketio.Conn, payload string) {
		var req struct {
			RoomID          string      `json:"roomId"`
			ProducerID      string      `json:"producerId"`
			RtpCapabilities interface{} `json:"rtpCapabilities"`
		}

		if err := json.Unmarshal([]byte(payload), &req); err != nil {
			s.Emit("error", map[string]string{"error": "Invalid payload"})
			return
		}

		consumer, err := mi.CreateConsumer(req.RoomID, s.ID(), req.ProducerID, req.RtpCapabilities)
		if err != nil {
			s.Emit("error", map[string]string{"error": fmt.Sprintf("Failed to consume: %v", err)})
			return
		}

		s.Emit("consumer-created", consumer)
		log.Printf("✓ Consumer created for client %s (producer: %s)", s.ID(), req.ProducerID)
	})

	// Close producer
	ss.IO.OnEvent("", "close-producer", func(s socketio.Conn, payload string) {
		var req struct {
			RoomID     string `json:"roomId"`
			ProducerID string `json:"producerId"`
		}

		if err := json.Unmarshal([]byte(payload), &req); err != nil {
			s.Emit("error", map[string]string{"error": "Invalid payload"})
			return
		}

		err := mi.CloseProducer(req.RoomID, s.ID(), req.ProducerID)
		if err != nil {
			s.Emit("error", map[string]string{"error": fmt.Sprintf("Failed to close producer: %v", err)})
			return
		}

		s.Emit("producer-closed", map[string]string{"producerId": req.ProducerID})
		log.Printf("✓ Producer closed for client %s", s.ID())
	})

	// Close consumer
	ss.IO.OnEvent("", "close-consumer", func(s socketio.Conn, payload string) {
		var req struct {
			RoomID     string `json:"roomId"`
			ConsumerID string `json:"consumerId"`
		}

		if err := json.Unmarshal([]byte(payload), &req); err != nil {
			s.Emit("error", map[string]string{"error": "Invalid payload"})
			return
		}

		err := mi.CloseConsumer(req.RoomID, s.ID(), req.ConsumerID)
		if err != nil {
			s.Emit("error", map[string]string{"error": fmt.Sprintf("Failed to close consumer: %v", err)})
			return
		}

		s.Emit("consumer-closed", map[string]string{"consumerId": req.ConsumerID})
		log.Printf("✓ Consumer closed for client %s", s.ID())
	})

	// Get room info from Mediasoup
	ss.IO.OnEvent("", "get-room-info", func(s socketio.Conn, payload string) {
		var req struct {
			RoomID string `json:"roomId"`
		}

		if err := json.Unmarshal([]byte(payload), &req); err != nil {
			s.Emit("error", map[string]string{"error": "Invalid payload"})
			return
		}

		roomInfo, err := mi.GetRoomInfo(req.RoomID)
		if err != nil {
			s.Emit("error", map[string]string{"error": fmt.Sprintf("Failed to get room info: %v", err)})
			return
		}

		s.Emit("room-info", roomInfo)
		log.Printf("✓ Room info retrieved for %s", req.RoomID)
	})
}
