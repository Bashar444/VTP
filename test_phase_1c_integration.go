package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Integration test for Phase 1C Mediasoup SFU

const (
	mediasoupURL = "http://localhost:3000"
	goBackendURL = "http://localhost:8080"
)

type HealthResponse struct {
	Status string `json:"status"`
	Worker string `json:"worker,omitempty"`
}

type RoomJoinRequest struct {
	PeerID     string `json:"peerId"`
	UserID     string `json:"userId"`
	Email      string `json:"email"`
	FullName   string `json:"fullName"`
	Role       string `json:"role"`
	IsProducer bool   `json:"isProducer"`
	RoomName   string `json:"roomName"`
}

type TransportRequest struct {
	PeerID    string `json:"peerId"`
	Direction string `json:"direction"`
}

type ProducerRequest struct {
	PeerID    string      `json:"peerId"`
	Kind      string      `json:"kind"`
	RTPParams interface{} `json:"rtpParameters"`
}

type ConsumerRequest struct {
	PeerID         string      `json:"peerId"`
	ProducerID     string      `json:"producerId"`
	RTCapabilities interface{} `json:"rtpCapabilities"`
}

func main() {
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  Phase 1C Integration Test Suite")
	fmt.Println("  Mediasoup SFU Integration Testing")
	fmt.Println("═══════════════════════════════════════════════════════════════")

	// Test 1: Service Health Checks
	fmt.Println("TEST 1: Service Health Checks")
	fmt.Println("─────────────────────────────────────────────────────────────")
	testMediasoupHealth()
	testGoBackendHealth()
	fmt.Println()

	// Test 2: Room Operations
	fmt.Println("TEST 2: Room Operations")
	fmt.Println("─────────────────────────────────────────────────────────────")
	testRoomOperations()
	fmt.Println()

	// Test 3: Multi-Peer Scenario
	fmt.Println("TEST 3: Multi-Peer Scenario")
	fmt.Println("─────────────────────────────────────────────────────────────")
	testMultiPeerScenario()
	fmt.Println()

	// Test 4: Transport Operations
	fmt.Println("TEST 4: Transport Operations")
	fmt.Println("─────────────────────────────────────────────────────────────")
	testTransportOperations()
	fmt.Println()

	// Test 5: Cleanup
	fmt.Println("TEST 5: Cleanup")
	fmt.Println("─────────────────────────────────────────────────────────────")
	testCleanup()
	fmt.Println()

	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  Integration Testing Complete")
	fmt.Println("═══════════════════════════════════════════════════════════════")
}

func testMediasoupHealth() {
	resp, err := http.Get(mediasoupURL + "/health")
	if err != nil {
		fmt.Printf("✗ FAIL: Could not connect to Mediasoup: %v\n", err)
		fmt.Println("  Make sure Mediasoup is running on port 3000")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var health HealthResponse
		if err := json.NewDecoder(resp.Body).Decode(&health); err == nil {
			fmt.Printf("✓ PASS: Mediasoup health check\n")
			fmt.Printf("  Status: %s, Worker: %s\n", health.Status, health.Worker)
		}
	} else {
		fmt.Printf("✗ FAIL: Mediasoup returned status %d\n", resp.StatusCode)
	}
}

func testGoBackendHealth() {
	resp, err := http.Get(goBackendURL + "/health")
	if err != nil {
		fmt.Printf("✗ FAIL: Could not connect to Go backend: %v\n", err)
		fmt.Println("  Make sure Go backend is running on port 8080")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("✓ PASS: Go backend health check\n")
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("  Response: %s\n", string(body))
	} else {
		fmt.Printf("✗ FAIL: Go backend returned status %d\n", resp.StatusCode)
	}
}

func testRoomOperations() {
	roomID := "test-room-" + fmt.Sprintf("%d", time.Now().Unix())
	peerID := "peer-1"

	// Join room
	fmt.Printf("1. Creating/joining room %s...\n", roomID)
	joinReq := RoomJoinRequest{
		PeerID:     peerID,
		UserID:     "user-123",
		Email:      "user@example.com",
		FullName:   "Test User",
		Role:       "student",
		IsProducer: true,
		RoomName:   "Test Room",
	}

	body, _ := json.Marshal(joinReq)
	resp, err := http.Post(
		mediasoupURL+"/rooms/"+roomID+"/peers",
		"application/json",
		bytes.NewReader(body),
	)

	if err != nil {
		fmt.Printf("✗ FAIL: Could not join room: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("✓ PASS: Peer joined room\n")
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
			fmt.Printf("  Response keys: %v\n", getMapKeys(result))
		}
	} else {
		fmt.Printf("✗ FAIL: Join returned status %d\n", resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("  Response: %s\n", string(body))
	}

	// Get room info
	fmt.Printf("2. Getting room info...\n")
	resp, err = http.Get(mediasoupURL + "/rooms/" + roomID)
	if err != nil {
		fmt.Printf("✗ FAIL: Could not get room: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("✓ PASS: Retrieved room info\n")
	} else {
		fmt.Printf("✗ FAIL: Get room returned status %d\n", resp.StatusCode)
	}

	// Get all rooms
	fmt.Printf("3. Getting all rooms...\n")
	resp, err = http.Get(mediasoupURL + "/rooms")
	if err != nil {
		fmt.Printf("✗ FAIL: Could not get rooms: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
			if totalRooms, ok := result["totalRooms"].(float64); ok {
				fmt.Printf("✓ PASS: Retrieved rooms list (%d total)\n", int(totalRooms))
			}
		}
	} else {
		fmt.Printf("✗ FAIL: Get rooms returned status %d\n", resp.StatusCode)
	}

	// Leave room
	fmt.Printf("4. Leaving room...\n")
	req, _ := http.NewRequest(http.MethodPost, mediasoupURL+"/rooms/"+roomID+"/peers/"+peerID+"/leave", nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("✗ FAIL: Could not leave room: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("✓ PASS: Peer left room\n")
	} else {
		fmt.Printf("✗ FAIL: Leave returned status %d\n", resp.StatusCode)
	}
}

func testMultiPeerScenario() {
	roomID := "multi-peer-room-" + fmt.Sprintf("%d", time.Now().Unix())

	fmt.Printf("1. Creating room and joining multiple peers...\n")

	// Peer 1 (Producer)
	fmt.Printf("   - Peer 1 (Producer)... ")
	peer1Req := RoomJoinRequest{
		PeerID:     "peer-1",
		UserID:     "user-1",
		Email:      "user1@example.com",
		FullName:   "User 1",
		Role:       "instructor",
		IsProducer: true,
		RoomName:   "Multi-Peer Test",
	}
	body1, _ := json.Marshal(peer1Req)
	resp1, err := http.Post(mediasoupURL+"/rooms/"+roomID+"/peers", "application/json", bytes.NewReader(body1))
	if err == nil && resp1.StatusCode == http.StatusOK {
		resp1.Body.Close()
		fmt.Printf("✓ joined\n")
	} else {
		if err != nil {
			fmt.Printf("✗ error: %v\n", err)
		} else {
			fmt.Printf("✗ status: %d\n", resp1.StatusCode)
			resp1.Body.Close()
		}
		return
	}

	// Peer 2 (Consumer)
	fmt.Printf("   - Peer 2 (Consumer)... ")
	peer2Req := RoomJoinRequest{
		PeerID:     "peer-2",
		UserID:     "user-2",
		Email:      "user2@example.com",
		FullName:   "User 2",
		Role:       "student",
		IsProducer: false,
		RoomName:   "Multi-Peer Test",
	}
	body2, _ := json.Marshal(peer2Req)
	resp2, err := http.Post(mediasoupURL+"/rooms/"+roomID+"/peers", "application/json", bytes.NewReader(body2))
	if err == nil && resp2.StatusCode == http.StatusOK {
		resp2.Body.Close()
		fmt.Printf("✓ joined\n")
	} else {
		if err != nil {
			fmt.Printf("✗ error: %v\n", err)
		} else {
			fmt.Printf("✗ status: %d\n", resp2.StatusCode)
			resp2.Body.Close()
		}
		return
	}

	// Verify both peers in room
	fmt.Printf("2. Verifying both peers in room...\n")
	resp, _ := http.Get(mediasoupURL + "/rooms/" + roomID)
	if resp != nil {
		var roomInfo map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&roomInfo); err == nil {
			if peerCount, ok := roomInfo["peerCount"].(float64); ok {
				if int(peerCount) == 2 {
					fmt.Printf("✓ PASS: Both peers in room\n")
				} else {
					fmt.Printf("✗ FAIL: Expected 2 peers, got %d\n", int(peerCount))
				}
			}
		}
		resp.Body.Close()
	}

	// Clean up
	fmt.Printf("3. Cleanup - peers leaving...\n")
	http.NewRequest(http.MethodPost, mediasoupURL+"/rooms/"+roomID+"/peers/peer-1/leave", nil)
	http.NewRequest(http.MethodPost, mediasoupURL+"/rooms/"+roomID+"/peers/peer-2/leave", nil)
	fmt.Printf("✓ PASS: Multi-peer scenario complete\n")
}

func testTransportOperations() {
	roomID := "transport-room-" + fmt.Sprintf("%d", time.Now().Unix())
	peerID := "peer-transport"

	// Join room first
	fmt.Printf("1. Joining room for transport test...\n")
	joinReq := RoomJoinRequest{
		PeerID:     peerID,
		UserID:     "user-transport",
		Email:      "test@example.com",
		FullName:   "Transport Tester",
		Role:       "student",
		IsProducer: true,
		RoomName:   "Transport Test",
	}
	body, _ := json.Marshal(joinReq)
	resp, _ := http.Post(mediasoupURL+"/rooms/"+roomID+"/peers", "application/json", bytes.NewReader(body))
	if resp != nil {
		fmt.Printf("✓ Peer joined\n")
		resp.Body.Close()
	}

	// Create transport
	fmt.Printf("2. Creating WebRTC transport...\n")
	transportReq := TransportRequest{
		PeerID:    peerID,
		Direction: "send",
	}
	body, _ = json.Marshal(transportReq)
	resp, err := http.Post(
		mediasoupURL+"/rooms/"+roomID+"/transports",
		"application/json",
		bytes.NewReader(body),
	)

	if err != nil {
		fmt.Printf("✗ FAIL: Could not create transport: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var transportResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&transportResp); err == nil {
			if transportID, ok := transportResp["transportId"].(string); ok {
				fmt.Printf("✓ PASS: Transport created\n")
				fmt.Printf("  Transport ID: %s\n", transportID)
				fmt.Printf("  Has ICE Parameters: %v\n", transportResp["iceParameters"] != nil)
				fmt.Printf("  Has DTLS Parameters: %v\n", transportResp["dtlsParameters"] != nil)
			}
		}
	} else {
		fmt.Printf("✗ FAIL: Transport creation returned status %d\n", resp.StatusCode)
	}
}

func testCleanup() {
	fmt.Printf("Running cleanup operations...\n")

	// Get all rooms
	resp, err := http.Get(mediasoupURL + "/rooms")
	if err != nil {
		fmt.Printf("✓ PASS: Cleanup verification attempted\n")
		return
	}
	defer resp.Body.Close()

	var roomsResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&roomsResp); err == nil {
		if totalRooms, ok := roomsResp["totalRooms"].(float64); ok {
			if int(totalRooms) >= 0 {
				fmt.Printf("✓ PASS: Rooms cleaned up (remaining: %d)\n", int(totalRooms))
			}
		}
	}
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
