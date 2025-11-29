package mediasoup

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("Expected path /health, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","timestamp":"2024-01-01T00:00:00Z","worker":"active"}`)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	health, err := client.Health()
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}
	if health.Status != "ok" {
		t.Errorf("Expected status 'ok', got %s", health.Status)
	}
}

func TestGetRooms(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rooms" {
			t.Errorf("Expected path /rooms, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"rooms":[],"totalRooms":0,"totalPeers":0}`)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	rooms, err := client.GetRooms()
	if err != nil {
		t.Errorf("GetRooms failed: %v", err)
	}
	if rooms.TotalRooms != 0 {
		t.Errorf("Expected 0 rooms, got %d", rooms.TotalRooms)
	}
}

func TestGetRoom(t *testing.T) {
	roomID := "test-room-123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := fmt.Sprintf("/rooms/%s", roomID)
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"roomId":"%s","name":"Test Room","peerCount":0,"peers":[],"createdAt":1234567890}`, roomID)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	room, err := client.GetRoom(roomID)
	if err != nil {
		t.Errorf("GetRoom failed: %v", err)
	}
	if room.RoomID != roomID {
		t.Errorf("Expected roomID %s, got %s", roomID, room.RoomID)
	}
	if room.PeerCount != 0 {
		t.Errorf("Expected 0 peers, got %d", room.PeerCount)
	}
}

func TestCreateTransport(t *testing.T) {
	roomID := "test-room"
	peerId := "peer-123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"transportId":"transport-123",
			"iceParameters":{"usernameFrag":"abc","password":"xyz"},
			"iceCandidates":[],
			"dtlsParameters":{"role":"auto","fingerprints":[]}
		}`)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	transport, err := client.CreateTransport(roomID, peerId, "send")
	if err != nil {
		t.Errorf("CreateTransport failed: %v", err)
	}
	if transport.TransportID != "transport-123" {
		t.Errorf("Expected transportID 'transport-123', got %s", transport.TransportID)
	}
}

func TestJoinRoom(t *testing.T) {
	roomID := "test-room"
	peerId := "peer-123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"roomId":"%s",
			"peerId":"%s",
			"rtpCapabilities":{},
			"peers":[]
		}`, roomID, peerId)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	req := &JoinRoomRequest{
		PeerID:     peerId,
		UserID:     "user-123",
		Email:      "user@example.com",
		FullName:   "Test User",
		Role:       "student",
		IsProducer: true,
	}
	resp, err := client.JoinRoom(roomID, req)
	if err != nil {
		t.Errorf("JoinRoom failed: %v", err)
	}
	if resp.PeerID != peerId {
		t.Errorf("Expected peerId %s, got %s", peerId, resp.PeerID)
	}
	if resp.RoomID != roomID {
		t.Errorf("Expected roomID %s, got %s", roomID, resp.RoomID)
	}
}

func TestLeaveRoom(t *testing.T) {
	roomID := "test-room"
	peerId := "peer-123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message":"peer left"}`)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.LeaveRoom(roomID, peerId)
	if err != nil {
		t.Errorf("LeaveRoom failed: %v", err)
	}
}

func TestCreateProducer(t *testing.T) {
	roomID := "test-room"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"id":"producer-123",
			"kind":"video",
			"rtpParameters":{}
		}`)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	req := &ProducerRequest{
		PeerID: "peer-123",
		Kind:   "video",
	}
	producer, err := client.CreateProducer(roomID, req)
	if err != nil {
		t.Errorf("CreateProducer failed: %v", err)
	}
	if producer.ID != "producer-123" {
		t.Errorf("Expected producerID 'producer-123', got %s", producer.ID)
	}
	if producer.Kind != "video" {
		t.Errorf("Expected kind 'video', got %s", producer.Kind)
	}
}

func TestCreateConsumer(t *testing.T) {
	roomID := "test-room"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"id":"consumer-123",
			"producerId":"producer-123",
			"kind":"video",
			"rtpParameters":{}
		}`)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	req := &ConsumerRequest{
		PeerID:     "peer-456",
		ProducerID: "producer-123",
	}
	consumer, err := client.CreateConsumer(roomID, req)
	if err != nil {
		t.Errorf("CreateConsumer failed: %v", err)
	}
	if consumer.ID != "consumer-123" {
		t.Errorf("Expected consumerID 'consumer-123', got %s", consumer.ID)
	}
	if consumer.ProducerID != "producer-123" {
		t.Errorf("Expected producerID 'producer-123', got %s", consumer.ProducerID)
	}
}

func TestCloseProducer(t *testing.T) {
	roomID := "test-room"
	producerID := "producer-123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message":"producer closed"}`)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.CloseProducer(roomID, producerID)
	if err != nil {
		t.Errorf("CloseProducer failed: %v", err)
	}
}

func TestCloseConsumer(t *testing.T) {
	roomID := "test-room"
	consumerID := "consumer-123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message":"consumer closed"}`)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.CloseConsumer(roomID, consumerID)
	if err != nil {
		t.Errorf("CloseConsumer failed: %v", err)
	}
}
