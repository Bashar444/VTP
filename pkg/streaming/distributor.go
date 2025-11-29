package streaming

import (
	"fmt"
	"sync"
	"time"
)

// SegmentDeliveryProfile defines segment delivery characteristics
type SegmentDeliveryProfile struct {
	Name              string
	SegmentDuration   time.Duration
	BufferSize        int // Number of segments to buffer
	MaxConnections    int // Max concurrent viewers
	RetryAttempts     int // Retry failed deliveries
	TimeoutDuration   time.Duration
	CompressionLevel  int  // 0-9, 0=none, 9=max
	AdaptiveBuffering bool // Enable buffer adaptation
}

// ViewerSession represents a live streaming viewer connection
type ViewerSession struct {
	ViewerID          string
	SessionID         string
	RecordingID       string
	StartTime         time.Time
	LastSegmentTime   time.Time
	SegmentsReceived  int64
	BytesReceived     int64
	CurrentBitrate    string  // Current profile being delivered
	BufferHealth      float64 // 0-100, percentage buffer fullness
	ConnectionQuality string  // excellent, good, fair, poor
	Active            bool
	Timestamp         time.Time
}

// SegmentQueue holds video segments ready for delivery
type SegmentQueue struct {
	segments    map[string]*VideoSegment
	queue       []*VideoSegment
	maxSize     int
	mu          sync.RWMutex
	lastCleanup time.Time
}

// VideoSegment represents an encoded video segment
type VideoSegment struct {
	SegmentID   string
	RecordingID string
	Bitrate     string // 500k, 1k, 2k, 4k
	SequenceNum int64
	DurationMs  int64
	FilePath    string
	FileSize    int64
	CreatedTime time.Time
	ExpiresTime time.Time
	ChecksumMD5 string
	Delivered   int64 // Count of successful deliveries
	FailedCount int64 // Count of failed deliveries
	IsKeyFrame  bool
}

// DistributionStats tracks distribution network statistics
type DistributionStats struct {
	TotalSegmentsQueued int64
	TotalSegmentsServed int64
	TotalBytesServed    int64
	ActiveViewers       int
	PeakViewers         int
	AverageBitrate      float64
	SegmentLatency      time.Duration
	QueueUtilization    float64 // Percentage
	FailureRate         float64 // Percentage
	Timestamp           time.Time
}

// LiveDistributor manages video segment distribution to multiple viewers
type LiveDistributor struct {
	RecordingID          string
	DistributionProfiles map[string]*SegmentDeliveryProfile
	segmentQueue         *SegmentQueue
	activeSessions       map[string]*ViewerSession
	sessionsMu           sync.RWMutex
	stats                DistributionStats
	statsMu              sync.RWMutex
	progressCallbacks    map[string]func(SegmentDeliveryEvent)
	callbacksMu          sync.RWMutex
	mu                   sync.RWMutex
	closed               bool
	maxViewers           int
	segmentRetentionTime time.Duration
	bandwidthLimitMBps   float64
}

// SegmentDeliveryEvent reports distribution events
type SegmentDeliveryEvent struct {
	EventType        string // "segment_delivered", "viewer_joined", "viewer_left", "quality_switched"
	RecordingID      string
	SegmentID        string
	ViewerID         string
	CurrentBitrate   string
	BytesTransferred int64
	Latency          time.Duration
	Success          bool
	ErrorMessage     string
	Timestamp        time.Time
}

// NewLiveDistributor creates a new distribution manager
func NewLiveDistributor(recordingID string, maxViewers int, segmentRetentionTime time.Duration) *LiveDistributor {
	profiles := map[string]*SegmentDeliveryProfile{
		"VeryLow": {
			Name:              "VeryLow",
			SegmentDuration:   2 * time.Second,
			BufferSize:        3,
			MaxConnections:    100,
			RetryAttempts:     3,
			TimeoutDuration:   5 * time.Second,
			CompressionLevel:  6,
			AdaptiveBuffering: true,
		},
		"Low": {
			Name:              "Low",
			SegmentDuration:   2 * time.Second,
			BufferSize:        4,
			MaxConnections:    75,
			RetryAttempts:     3,
			TimeoutDuration:   5 * time.Second,
			CompressionLevel:  5,
			AdaptiveBuffering: true,
		},
		"Medium": {
			Name:              "Medium",
			SegmentDuration:   2 * time.Second,
			BufferSize:        5,
			MaxConnections:    50,
			RetryAttempts:     3,
			TimeoutDuration:   6 * time.Second,
			CompressionLevel:  4,
			AdaptiveBuffering: true,
		},
		"High": {
			Name:              "High",
			SegmentDuration:   2 * time.Second,
			BufferSize:        6,
			MaxConnections:    25,
			RetryAttempts:     2,
			TimeoutDuration:   7 * time.Second,
			CompressionLevel:  3,
			AdaptiveBuffering: true,
		},
	}

	return &LiveDistributor{
		RecordingID:          recordingID,
		DistributionProfiles: profiles,
		segmentQueue:         NewSegmentQueue(1000),
		activeSessions:       make(map[string]*ViewerSession),
		progressCallbacks:    make(map[string]func(SegmentDeliveryEvent)),
		maxViewers:           maxViewers,
		segmentRetentionTime: segmentRetentionTime,
		bandwidthLimitMBps:   100.0, // Default 100 Mbps
	}
}

// NewSegmentQueue creates a new segment queue
func NewSegmentQueue(maxSize int) *SegmentQueue {
	return &SegmentQueue{
		segments:    make(map[string]*VideoSegment),
		queue:       make([]*VideoSegment, 0, maxSize),
		maxSize:     maxSize,
		lastCleanup: time.Now(),
	}
}

// EnqueueSegment adds a segment to the distribution queue
func (ld *LiveDistributor) EnqueueSegment(segment *VideoSegment) error {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	if ld.closed {
		return fmt.Errorf("distributor is closed")
	}

	if segment == nil {
		return fmt.Errorf("segment cannot be nil")
	}

	if !segment.IsKeyFrame && ld.segmentQueue.Length() == 0 {
		return fmt.Errorf("first segment must be keyframe")
	}

	err := ld.segmentQueue.Add(segment)
	if err != nil {
		return err
	}

	ld.statsMu.Lock()
	ld.stats.TotalSegmentsQueued++
	ld.statsMu.Unlock()

	return nil
}

// JoinViewer creates a new viewer session
func (ld *LiveDistributor) JoinViewer(viewerID, bitrate string) (*ViewerSession, error) {
	ld.sessionsMu.Lock()
	defer ld.sessionsMu.Unlock()

	if ld.closed {
		return nil, fmt.Errorf("distributor is closed")
	}

	if len(ld.activeSessions) >= ld.maxViewers {
		return nil, fmt.Errorf("max viewers (%d) reached", ld.maxViewers)
	}

	if _, exists := ld.activeSessions[viewerID]; exists {
		return nil, fmt.Errorf("viewer already connected: %s", viewerID)
	}

	if _, validBitrate := ld.DistributionProfiles[bitrate]; !validBitrate {
		return nil, fmt.Errorf("invalid bitrate: %s", bitrate)
	}

	session := &ViewerSession{
		ViewerID:          viewerID,
		SessionID:         fmt.Sprintf("%s-viewer-%s-%d", ld.RecordingID, viewerID, time.Now().Unix()),
		RecordingID:       ld.RecordingID,
		StartTime:         time.Now(),
		CurrentBitrate:    bitrate,
		BufferHealth:      50.0,
		ConnectionQuality: "excellent",
		Active:            true,
		Timestamp:         time.Now(),
	}

	ld.activeSessions[viewerID] = session

	// Update stats
	ld.statsMu.Lock()
	ld.stats.ActiveViewers = len(ld.activeSessions)
	if ld.stats.ActiveViewers > ld.stats.PeakViewers {
		ld.stats.PeakViewers = ld.stats.ActiveViewers
	}
	ld.statsMu.Unlock()

	// Fire callback
	ld.fireDeliveryEvent(SegmentDeliveryEvent{
		EventType:      "viewer_joined",
		RecordingID:    ld.RecordingID,
		ViewerID:       viewerID,
		CurrentBitrate: bitrate,
		Success:        true,
		Timestamp:      time.Now(),
	})

	return session, nil
}

// LeaveViewer removes a viewer session
func (ld *LiveDistributor) LeaveViewer(viewerID string) error {
	ld.sessionsMu.Lock()
	defer ld.sessionsMu.Unlock()

	session, exists := ld.activeSessions[viewerID]
	if !exists {
		return fmt.Errorf("viewer not found: %s", viewerID)
	}

	delete(ld.activeSessions, viewerID)

	// Update stats
	ld.statsMu.Lock()
	ld.stats.ActiveViewers = len(ld.activeSessions)
	ld.statsMu.Unlock()

	// Fire callback
	ld.fireDeliveryEvent(SegmentDeliveryEvent{
		EventType:      "viewer_left",
		RecordingID:    ld.RecordingID,
		ViewerID:       viewerID,
		CurrentBitrate: session.CurrentBitrate,
		Success:        true,
		Timestamp:      time.Now(),
	})

	return nil
}

// DeliverSegment delivers a segment to a specific viewer
func (ld *LiveDistributor) DeliverSegment(viewerID, segmentID string) error {
	ld.sessionsMu.RLock()
	session, exists := ld.activeSessions[viewerID]
	ld.sessionsMu.RUnlock()

	if !exists {
		return fmt.Errorf("viewer not found: %s", viewerID)
	}

	segment := ld.segmentQueue.Get(segmentID)
	if segment == nil {
		return fmt.Errorf("segment not found: %s", segmentID)
	}

	// Check if segment bitrate matches session bitrate
	if segment.Bitrate != session.CurrentBitrate {
		return fmt.Errorf("segment bitrate mismatch: expected %s, got %s", session.CurrentBitrate, segment.Bitrate)
	}

	// Simulate delivery
	session.LastSegmentTime = time.Now()
	session.SegmentsReceived++
	session.BytesReceived += segment.FileSize

	ld.statsMu.Lock()
	ld.stats.TotalSegmentsServed++
	ld.stats.TotalBytesServed += segment.FileSize
	ld.statsMu.Unlock()

	segment.Delivered++

	// Fire callback
	ld.fireDeliveryEvent(SegmentDeliveryEvent{
		EventType:        "segment_delivered",
		RecordingID:      ld.RecordingID,
		SegmentID:        segmentID,
		ViewerID:         viewerID,
		CurrentBitrate:   session.CurrentBitrate,
		BytesTransferred: segment.FileSize,
		Success:          true,
		Timestamp:        time.Now(),
	})

	return nil
}

// SwitchBitrate adapts viewer bitrate based on conditions
func (ld *LiveDistributor) SwitchBitrate(viewerID, newBitrate string) error {
	ld.sessionsMu.Lock()
	session, exists := ld.activeSessions[viewerID]
	ld.sessionsMu.Unlock()

	if !exists {
		return fmt.Errorf("viewer not found: %s", viewerID)
	}

	if _, validBitrate := ld.DistributionProfiles[newBitrate]; !validBitrate {
		return fmt.Errorf("invalid bitrate: %s", newBitrate)
	}

	oldBitrate := session.CurrentBitrate
	session.CurrentBitrate = newBitrate
	session.Timestamp = time.Now()

	// Fire callback
	ld.fireDeliveryEvent(SegmentDeliveryEvent{
		EventType:      "quality_switched",
		RecordingID:    ld.RecordingID,
		ViewerID:       viewerID,
		CurrentBitrate: newBitrate,
		Success:        true,
		Timestamp:      time.Now(),
	})

	fmt.Printf("[Distribution] Viewer %s: %s → %s\n", viewerID, oldBitrate, newBitrate)

	return nil
}

// UpdateViewerBuffer updates viewer buffer health
func (ld *LiveDistributor) UpdateViewerBuffer(viewerID string, bufferHealth float64) error {
	ld.sessionsMu.Lock()
	session, exists := ld.activeSessions[viewerID]
	ld.sessionsMu.Unlock()

	if !exists {
		return fmt.Errorf("viewer not found: %s", viewerID)
	}

	if bufferHealth < 0 || bufferHealth > 100 {
		return fmt.Errorf("buffer health must be 0-100: got %.1f", bufferHealth)
	}

	session.BufferHealth = bufferHealth

	// Update connection quality based on buffer
	if bufferHealth > 80 {
		session.ConnectionQuality = "excellent"
	} else if bufferHealth > 60 {
		session.ConnectionQuality = "good"
	} else if bufferHealth > 40 {
		session.ConnectionQuality = "fair"
	} else {
		session.ConnectionQuality = "poor"
	}

	return nil
}

// GetNextSegment gets the next segment for a viewer
func (ld *LiveDistributor) GetNextSegment(viewerID string) (*VideoSegment, error) {
	ld.sessionsMu.RLock()
	session, exists := ld.activeSessions[viewerID]
	ld.sessionsMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("viewer not found: %s", viewerID)
	}

	// Get segment matching viewer's current bitrate
	segments := ld.segmentQueue.GetSegmentsByBitrate(session.CurrentBitrate)
	if len(segments) == 0 {
		return nil, fmt.Errorf("no segments available for bitrate: %s", session.CurrentBitrate)
	}

	// Return most recent segment
	return segments[len(segments)-1], nil
}

// GetViewerSession retrieves viewer session information
func (ld *LiveDistributor) GetViewerSession(viewerID string) (*ViewerSession, error) {
	ld.sessionsMu.RLock()
	session, exists := ld.activeSessions[viewerID]
	ld.sessionsMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("viewer not found: %s", viewerID)
	}

	return session, nil
}

// GetAllViewerSessions returns all active viewer sessions
func (ld *LiveDistributor) GetAllViewerSessions() map[string]*ViewerSession {
	ld.sessionsMu.RLock()
	defer ld.sessionsMu.RUnlock()

	// Return copy to avoid external modification
	sessionsCopy := make(map[string]*ViewerSession)
	for k, v := range ld.activeSessions {
		sessionsCopy[k] = v
	}

	return sessionsCopy
}

// GetDistributionStats returns current distribution statistics
func (ld *LiveDistributor) GetDistributionStats() DistributionStats {
	ld.statsMu.RLock()
	defer ld.statsMu.RUnlock()

	stats := ld.stats
	stats.Timestamp = time.Now()

	ld.sessionsMu.RLock()
	stats.ActiveViewers = len(ld.activeSessions)
	ld.sessionsMu.RUnlock()

	if ld.segmentQueue.Length() > 0 {
		stats.QueueUtilization = float64(ld.segmentQueue.Length()) / float64(ld.segmentQueue.maxSize) * 100.0
	}

	return stats
}

// RegisterDeliveryCallback registers a callback for distribution events
func (ld *LiveDistributor) RegisterDeliveryCallback(recordingID string, callback func(SegmentDeliveryEvent)) {
	ld.callbacksMu.Lock()
	defer ld.callbacksMu.Unlock()

	ld.progressCallbacks[recordingID] = callback
}

// fireDeliveryEvent fires a delivery event to registered callbacks
func (ld *LiveDistributor) fireDeliveryEvent(event SegmentDeliveryEvent) {
	ld.callbacksMu.RLock()
	callback, exists := ld.progressCallbacks[ld.RecordingID]
	ld.callbacksMu.RUnlock()

	if exists && callback != nil {
		go callback(event)
	}
}

// SegmentQueue helper methods

// Add adds a segment to the queue
func (sq *SegmentQueue) Add(segment *VideoSegment) error {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	if len(sq.queue) >= sq.maxSize {
		return fmt.Errorf("queue is full")
	}

	sq.queue = append(sq.queue, segment)
	sq.segments[segment.SegmentID] = segment

	return nil
}

// Get retrieves a segment from the queue
func (sq *SegmentQueue) Get(segmentID string) *VideoSegment {
	sq.mu.RLock()
	defer sq.mu.RUnlock()

	return sq.segments[segmentID]
}

// Length returns the number of segments in queue
func (sq *SegmentQueue) Length() int {
	sq.mu.RLock()
	defer sq.mu.RUnlock()

	return len(sq.queue)
}

// GetSegmentsByBitrate returns all segments for a specific bitrate
func (sq *SegmentQueue) GetSegmentsByBitrate(bitrate string) []*VideoSegment {
	sq.mu.RLock()
	defer sq.mu.RUnlock()

	var matching []*VideoSegment
	for _, segment := range sq.queue {
		if segment.Bitrate == bitrate {
			matching = append(matching, segment)
		}
	}

	return matching
}

// Remove removes a segment from the queue
func (sq *SegmentQueue) Remove(segmentID string) {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	if segment, exists := sq.segments[segmentID]; exists {
		// Remove from queue slice
		for i, s := range sq.queue {
			if s.SegmentID == segmentID {
				sq.queue = append(sq.queue[:i], sq.queue[i+1:]...)
				break
			}
		}
		delete(sq.segments, segmentID)
		_ = segment
	}
}

// CleanupExpiredSegments removes expired segments
func (sq *SegmentQueue) CleanupExpiredSegments() int {
	sq.mu.Lock()
	defer sq.mu.Unlock()

	now := time.Now()
	removed := 0

	var newQueue []*VideoSegment
	for _, segment := range sq.queue {
		if now.Before(segment.ExpiresTime) {
			newQueue = append(newQueue, segment)
		} else {
			delete(sq.segments, segment.SegmentID)
			removed++
		}
	}

	sq.queue = newQueue
	sq.lastCleanup = now

	return removed
}

// Close closes the distributor and cleans up resources
func (ld *LiveDistributor) Close() error {
	ld.mu.Lock()
	defer ld.mu.Unlock()

	if ld.closed {
		return fmt.Errorf("distributor already closed")
	}

	ld.closed = true

	// Close all viewer sessions
	ld.sessionsMu.Lock()
	ld.activeSessions = make(map[string]*ViewerSession)
	ld.sessionsMu.Unlock()

	// Clear segment queue
	ld.segmentQueue.mu.Lock()
	ld.segmentQueue.queue = make([]*VideoSegment, 0)
	ld.segmentQueue.segments = make(map[string]*VideoSegment)
	ld.segmentQueue.mu.Unlock()

	return nil
}
