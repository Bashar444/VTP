package mediasoup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client for communicating with Mediasoup SFU service
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new Mediasoup client
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Room represents a room in Mediasoup
type Room struct {
	ID        string `json:"roomId"`
	Name      string `json:"name"`
	PeerCount int    `json:"peerCount"`
	CreatedAt int64  `json:"createdAt"`
}

// Peer represents a participant in a room
type Peer struct {
	ID            string `json:"peerId"`
	UserID        string `json:"userId"`
	Email         string `json:"email"`
	FullName      string `json:"fullName"`
	Role          string `json:"role"`
	IsProducer    bool   `json:"isProducer"`
	ProducerCount int    `json:"producerCount"`
	ConsumerCount int    `json:"consumerCount"`
	JoinedAt      int64  `json:"joinedAt"`
}

// Transport represents a WebRTC transport
type Transport struct {
	TransportID    string         `json:"transportId"`
	IceParameters  IceParameters  `json:"iceParameters"`
	IceCandidates  []IceCandidate `json:"iceCandidates"`
	DtlsParameters DtlsParameters `json:"dtlsParameters"`
	SctpParameters interface{}    `json:"sctpParameters"`
}

// IceParameters for WebRTC ICE
type IceParameters struct {
	UsernameFrag string `json:"usernameFrag"`
	Password     string `json:"password"`
}

// IceCandidate for WebRTC ICE
type IceCandidate struct {
	Foundation     string `json:"foundation"`
	Priority       int    `json:"priority"`
	IP             string `json:"ip"`
	Protocol       string `json:"protocol"`
	Port           int    `json:"port"`
	Type           string `json:"type"`
	TcpType        string `json:"tcpType,omitempty"`
	RelatedAddress string `json:"relatedAddress,omitempty"`
	RelatedPort    int    `json:"relatedPort,omitempty"`
}

// DtlsParameters for WebRTC DTLS
type DtlsParameters struct {
	Role         string            `json:"role"`
	Fingerprints []DtlsFingerprint `json:"fingerprints"`
}

// DtlsFingerprint represents a DTLS certificate fingerprint
type DtlsFingerprint struct {
	Algorithm string `json:"algorithm"`
	Value     string `json:"value"`
}

// Producer represents a media producer
type Producer struct {
	ID            string      `json:"id"`
	Kind          string      `json:"kind"`
	RtpParameters interface{} `json:"rtpParameters"`
}

// Consumer represents a media consumer
type Consumer struct {
	ID            string      `json:"id"`
	ProducerID    string      `json:"producerId"`
	Kind          string      `json:"kind"`
	RtpParameters interface{} `json:"rtpParameters"`
}

// RoomInfo contains information about a room
type RoomInfo struct {
	RoomID    string `json:"roomId"`
	Name      string `json:"name"`
	PeerCount int    `json:"peerCount"`
	Peers     []Peer `json:"peers"`
	CreatedAt int64  `json:"createdAt"`
}

// RoomsList contains a list of all rooms
type RoomsList struct {
	Rooms      []Room `json:"rooms"`
	TotalRooms int    `json:"totalRooms"`
	TotalPeers int    `json:"totalPeers"`
}

// JoinRoomRequest is used when joining a room
type JoinRoomRequest struct {
	PeerID     string `json:"peerId"`
	UserID     string `json:"userId"`
	Email      string `json:"email"`
	FullName   string `json:"fullName"`
	Role       string `json:"role"`
	IsProducer bool   `json:"isProducer"`
	RoomName   string `json:"roomName,omitempty"`
}

// JoinRoomResponse contains response after joining a room
type JoinRoomResponse struct {
	RoomID          string      `json:"roomId"`
	PeerID          string      `json:"peerId"`
	RtpCapabilities interface{} `json:"rtpCapabilities"`
	Peers           []Peer      `json:"peers"`
}

// TransportConnectRequest for connecting a transport
type TransportConnectRequest struct {
	DtlsParameters DtlsParameters `json:"dtlsParameters"`
}

// ProducerRequest for creating a producer
type ProducerRequest struct {
	PeerID        string      `json:"peerId"`
	Kind          string      `json:"kind"`
	RtpParameters interface{} `json:"rtpParameters"`
}

// ConsumerRequest for creating a consumer
type ConsumerRequest struct {
	PeerID          string      `json:"peerId"`
	ProducerID      string      `json:"producerId"`
	RtpCapabilities interface{} `json:"rtpCapabilities"`
}

// Health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Worker    string `json:"worker"`
}

// Helper function to make HTTP requests
func (c *Client) request(method, path string, body interface{}) ([]byte, error) {
	url := c.BaseURL + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Health checks the SFU service
func (c *Client) Health() (*HealthResponse, error) {
	data, err := c.request("GET", "/health", nil)
	if err != nil {
		return nil, err
	}

	var health HealthResponse
	if err := json.Unmarshal(data, &health); err != nil {
		return nil, err
	}

	return &health, nil
}

// GetRoom gets information about a specific room
func (c *Client) GetRoom(roomID string) (*RoomInfo, error) {
	data, err := c.request("GET", fmt.Sprintf("/rooms/%s", roomID), nil)
	if err != nil {
		return nil, err
	}

	var room RoomInfo
	if err := json.Unmarshal(data, &room); err != nil {
		return nil, err
	}

	return &room, nil
}

// GetRooms gets all rooms
func (c *Client) GetRooms() (*RoomsList, error) {
	data, err := c.request("GET", "/rooms", nil)
	if err != nil {
		return nil, err
	}

	var rooms RoomsList
	if err := json.Unmarshal(data, &rooms); err != nil {
		return nil, err
	}

	return &rooms, nil
}

// CreateTransport creates a WebRTC transport
func (c *Client) CreateTransport(roomID, peerId, direction string) (*Transport, error) {
	req := map[string]string{
		"peerId":    peerId,
		"direction": direction,
	}

	data, err := c.request("POST", fmt.Sprintf("/rooms/%s/transports", roomID), req)
	if err != nil {
		return nil, err
	}

	var transport Transport
	if err := json.Unmarshal(data, &transport); err != nil {
		return nil, err
	}

	return &transport, nil
}

// ConnectTransport connects a WebRTC transport
func (c *Client) ConnectTransport(roomID, transportID string, dtlsParams DtlsParameters) error {
	req := TransportConnectRequest{
		DtlsParameters: dtlsParams,
	}

	_, err := c.request("POST", fmt.Sprintf("/rooms/%s/transports/%s/connect", roomID, transportID), req)
	return err
}

// CreateProducer creates a media producer
func (c *Client) CreateProducer(roomID string, req *ProducerRequest) (*Producer, error) {
	data, err := c.request("POST", fmt.Sprintf("/rooms/%s/producers", roomID), req)
	if err != nil {
		return nil, err
	}

	var producer Producer
	if err := json.Unmarshal(data, &producer); err != nil {
		return nil, err
	}

	return &producer, nil
}

// CreateConsumer creates a media consumer
func (c *Client) CreateConsumer(roomID string, req *ConsumerRequest) (*Consumer, error) {
	data, err := c.request("POST", fmt.Sprintf("/rooms/%s/consumers", roomID), req)
	if err != nil {
		return nil, err
	}

	var consumer Consumer
	if err := json.Unmarshal(data, &consumer); err != nil {
		return nil, err
	}

	return &consumer, nil
}

// CloseProducer closes a producer
func (c *Client) CloseProducer(roomID, producerID string) error {
	_, err := c.request("POST", fmt.Sprintf("/rooms/%s/producers/%s/close", roomID, producerID), nil)
	return err
}

// CloseConsumer closes a consumer
func (c *Client) CloseConsumer(roomID, consumerID string) error {
	_, err := c.request("POST", fmt.Sprintf("/rooms/%s/consumers/%s/close", roomID, consumerID), nil)
	return err
}

// JoinRoom joins a participant to a room
func (c *Client) JoinRoom(roomID string, req *JoinRoomRequest) (*JoinRoomResponse, error) {
	data, err := c.request("POST", fmt.Sprintf("/rooms/%s/peers", roomID), req)
	if err != nil {
		return nil, err
	}

	var resp JoinRoomResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// LeaveRoom removes a participant from a room
func (c *Client) LeaveRoom(roomID, peerId string) error {
	_, err := c.request("POST", fmt.Sprintf("/rooms/%s/peers/%s/leave", roomID, peerId), nil)
	return err
}
