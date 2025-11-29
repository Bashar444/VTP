package mediasoup

// RtpCapabilities represents WebRTC RTP capabilities
type RtpCapabilities struct {
	Codecs           []RtpCodecCapability `json:"codecs"`
	HeaderExtensions []RtpHeaderExtension `json:"headerExtensions"`
	FecMechanisms    []string             `json:"fecMechanisms"`
}

// RtpCodecCapability represents an RTP codec capability
type RtpCodecCapability struct {
	Kind             string                 `json:"kind"`
	MimeType         string                 `json:"mimeType"`
	PreferredPayload int                    `json:"preferredPayload,omitempty"`
	ClockRate        int                    `json:"clockRate"`
	Channels         int                    `json:"channels,omitempty"`
	Parameters       map[string]interface{} `json:"parameters,omitempty"`
	RtcpFeedback     []RtcpFeedback         `json:"rtcpFeedback,omitempty"`
}

// RtcpFeedback represents RTCP feedback capability
type RtcpFeedback struct {
	Type      string `json:"type"`
	Parameter string `json:"parameter,omitempty"`
}

// RtpHeaderExtension represents an RTP header extension
type RtpHeaderExtension struct {
	Kind      string `json:"kind"`
	URI       string `json:"uri"`
	SendID    int    `json:"sendId"`
	ReceiveID int    `json:"receiveId"`
	Encrypt   bool   `json:"encrypt"`
	Preferred bool   `json:"preferred"`
}

// RtpParameters represents WebRTC RTP parameters
type RtpParameters struct {
	Mid              string                     `json:"mid,omitempty"`
	Codecs           []RtpCodecParameters       `json:"codecs"`
	HeaderExtensions []RtpHeaderExtensionParams `json:"headerExtensions,omitempty"`
	Encodings        []RtpEncodingParameters    `json:"encodings,omitempty"`
	RTCP             RTCP                       `json:"rtcp,omitempty"`
}

// RtpCodecParameters represents codec parameters
type RtpCodecParameters struct {
	MimeType     string                 `json:"mimeType"`
	PayloadType  int                    `json:"payloadType"`
	ClockRate    int                    `json:"clockRate"`
	Channels     int                    `json:"channels,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	RtcpFeedback []RtcpFeedback         `json:"rtcpFeedback,omitempty"`
}

// RtpHeaderExtensionParams represents header extension parameters
type RtpHeaderExtensionParams struct {
	URI       string `json:"uri"`
	ID        int    `json:"id"`
	Encrypt   bool   `json:"encrypt"`
	Preferred bool   `json:"preferred,omitempty"`
}

// RtpEncodingParameters represents encoding parameters
type RtpEncodingParameters struct {
	SSRC             int    `json:"ssrc,omitempty"`
	RID              string `json:"rid,omitempty"`
	CodecPayloadType int    `json:"codecPayloadType,omitempty"`
	RTX              *RTX   `json:"rtx,omitempty"`
	DTX              bool   `json:"dtx,omitempty"`
	Scalability      string `json:"scalabilityMode,omitempty"`
	MaxBitrate       int    `json:"maxBitrate,omitempty"`
	MaxFramerate     int    `json:"maxFramerate,omitempty"`
}

// RTX represents retransmission parameters
type RTX struct {
	SSRC int `json:"ssrc"`
}

// RTCP represents RTCP parameters
type RTCP struct {
	CNAME       string `json:"cname,omitempty"`
	ReducedSize bool   `json:"reducedSize,omitempty"`
	Mux         bool   `json:"mux,omitempty"`
}

// TransportOptions represents transport creation options
type TransportOptions struct {
	PeerID    string `json:"peerId"`
	Direction string `json:"direction"`
}

// TransportConnectOptions for connecting a transport
type TransportConnectOptions struct {
	DtlsParameters DtlsParameters `json:"dtlsParameters"`
}

// ProducerOptions for creating a producer
type ProducerOptions struct {
	PeerID        string        `json:"peerId"`
	Kind          string        `json:"kind"`
	RtpParameters RtpParameters `json:"rtpParameters"`
	Paused        bool          `json:"paused,omitempty"`
}

// ConsumerOptions for creating a consumer
type ConsumerOptions struct {
	PeerID          string          `json:"peerId"`
	ProducerID      string          `json:"producerId"`
	RtpCapabilities RtpCapabilities `json:"rtpCapabilities"`
	Paused          bool            `json:"paused,omitempty"`
}

// RoomStats contains room statistics
type RoomStats struct {
	RoomID         string `json:"roomId"`
	Name           string `json:"name"`
	PeerCount      int    `json:"peerCount"`
	ProducerCount  int    `json:"producerCount"`
	ConsumerCount  int    `json:"consumerCount"`
	TransportCount int    `json:"transportCount"`
	CreatedAt      int64  `json:"createdAt"`
	UpdatedAt      int64  `json:"updatedAt"`
}

// PeerStats contains peer statistics
type PeerStats struct {
	PeerID        string `json:"peerId"`
	UserID        string `json:"userId"`
	Email         string `json:"email"`
	FullName      string `json:"fullName"`
	Role          string `json:"role"`
	IsProducer    bool   `json:"isProducer"`
	ProducerCount int    `json:"producerCount"`
	ConsumerCount int    `json:"consumerCount"`
	JoinedAt      int64  `json:"joinedAt"`
	LastActivity  int64  `json:"lastActivity"`
}

// TransportStats contains transport statistics
type TransportStats struct {
	TransportID        string `json:"transportId"`
	PeerID             string `json:"peerId"`
	Direction          string `json:"direction"`
	State              string `json:"state"`
	BytesSent          int64  `json:"bytesSent"`
	BytesReceived      int64  `json:"bytesReceived"`
	IceConnectionState string `json:"iceConnectionState"`
	DtlsState          string `json:"dtlsState"`
	CreatedAt          int64  `json:"createdAt"`
}

// ProducerStats contains producer statistics
type ProducerStats struct {
	ProducerID  string `json:"producerId"`
	PeerID      string `json:"peerId"`
	Kind        string `json:"kind"`
	Paused      bool   `json:"paused"`
	Type        string `json:"type"`
	Score       int    `json:"score"`
	BytesSent   int64  `json:"bytesSent"`
	PacketsSent int64  `json:"packetsSent"`
	CreatedAt   int64  `json:"createdAt"`
}

// ConsumerStats contains consumer statistics
type ConsumerStats struct {
	ConsumerID    string `json:"consumerId"`
	ProducerID    string `json:"producerId"`
	PeerID        string `json:"peerId"`
	Kind          string `json:"kind"`
	Paused        bool   `json:"paused"`
	Score         int    `json:"score"`
	BytesReceived int64  `json:"bytesReceived"`
	PacketsLost   int64  `json:"packetsLost"`
	CreatedAt     int64  `json:"createdAt"`
}
