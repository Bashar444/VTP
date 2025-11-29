package streaming

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// DistributionService manages live distribution across multiple viewers
type DistributionService struct {
	distributors   map[string]*LiveDistributor
	distributorsMu sync.RWMutex
	workerCount    int
	workers        []*DistributionWorker
	workQueue      chan *SegmentDeliveryTask
	stopChan       chan bool
	wg             sync.WaitGroup
	logger         *log.Logger
	closed         bool
	metrics        DistributionMetrics
	metricsMu      sync.RWMutex
	cdnEnabled     bool
	cdnEndpoint    string
}

// DistributionWorker processes segment delivery tasks
type DistributionWorker struct {
	workerID int
	taskChan chan *SegmentDeliveryTask
	stopChan chan bool
	logger   *log.Logger
}

// SegmentDeliveryTask represents a segment delivery operation
type SegmentDeliveryTask struct {
	RecordingID string
	ViewerID    string
	SegmentID   string
	Priority    int // 0=low, 10=high
	CreatedTime time.Time
	Attempts    int
	MaxAttempts int
}

// DistributionMetrics tracks overall distribution metrics
type DistributionMetrics struct {
	TotalDistributors     int
	TotalActiveViewers    int
	TotalSegmentsQueued   int64
	TotalSegmentsServed   int64
	TotalBytesServed      int64
	TotalDeliveryFailures int64
	AverageLatency        time.Duration
	CDNCacheHitRate       float64
	Timestamp             time.Time
}

// NewDistributionService creates a new distribution service
func NewDistributionService(workerCount int, logger *log.Logger) *DistributionService {
	if workerCount < 1 {
		workerCount = 4
	}

	service := &DistributionService{
		distributors: make(map[string]*LiveDistributor),
		workerCount:  workerCount,
		workers:      make([]*DistributionWorker, workerCount),
		workQueue:    make(chan *SegmentDeliveryTask, 1000),
		stopChan:     make(chan bool),
		logger:       logger,
		cdnEnabled:   false,
		cdnEndpoint:  "https://cdn.example.com",
	}

	// Start worker goroutines
	for i := 0; i < workerCount; i++ {
		worker := &DistributionWorker{
			workerID: i,
			taskChan: service.workQueue,
			stopChan: make(chan bool),
			logger:   logger,
		}
		service.workers[i] = worker
		service.wg.Add(1)
		go service.distributionWorker(worker)
	}

	// Start cleanup goroutine
	service.wg.Add(1)
	go service.cleanupWorker()

	return service
}

// CreateDistributor creates a new live distributor for a recording
func (ds *DistributionService) CreateDistributor(recordingID string, maxViewers int, segmentRetention time.Duration) (*LiveDistributor, error) {
	ds.distributorsMu.Lock()
	defer ds.distributorsMu.Unlock()

	if _, exists := ds.distributors[recordingID]; exists {
		return nil, fmt.Errorf("distributor already exists for recording: %s", recordingID)
	}

	distributor := NewLiveDistributor(recordingID, maxViewers, segmentRetention)
	ds.distributors[recordingID] = distributor

	return distributor, nil
}

// GetDistributor retrieves a distributor by recording ID
func (ds *DistributionService) GetDistributor(recordingID string) (*LiveDistributor, error) {
	ds.distributorsMu.RLock()
	defer ds.distributorsMu.RUnlock()

	distributor, exists := ds.distributors[recordingID]
	if !exists {
		return nil, fmt.Errorf("distributor not found for recording: %s", recordingID)
	}

	return distributor, nil
}

// StartLiveStream initializes a new live stream
func (ds *DistributionService) StartLiveStream(recordingID string, maxViewers int) (*LiveDistributor, error) {
	return ds.CreateDistributor(recordingID, maxViewers, 30*time.Second)
}

// AddSegmentToStream adds a segment to the distribution stream
func (ds *DistributionService) AddSegmentToStream(recordingID, segmentID string, segment *VideoSegment) error {
	distributor, err := ds.GetDistributor(recordingID)
	if err != nil {
		return err
	}

	return distributor.EnqueueSegment(segment)
}

// JoinStream creates a viewer session for a stream
func (ds *DistributionService) JoinStream(recordingID, viewerID, initialBitrate string) (*ViewerSession, error) {
	distributor, err := ds.GetDistributor(recordingID)
	if err != nil {
		return nil, err
	}

	return distributor.JoinViewer(viewerID, initialBitrate)
}

// LeaveStream removes a viewer from a stream
func (ds *DistributionService) LeaveStream(recordingID, viewerID string) error {
	distributor, err := ds.GetDistributor(recordingID)
	if err != nil {
		return err
	}

	return distributor.LeaveViewer(viewerID)
}

// DeliverSegmentToViewer delivers a specific segment to a viewer
func (ds *DistributionService) DeliverSegmentToViewer(recordingID, viewerID, segmentID string) error {
	task := &SegmentDeliveryTask{
		RecordingID: recordingID,
		ViewerID:    viewerID,
		SegmentID:   segmentID,
		Priority:    5,
		CreatedTime: time.Now(),
		MaxAttempts: 3,
	}

	select {
	case ds.workQueue <- task:
		return nil
	case <-ds.stopChan:
		return fmt.Errorf("service is stopped")
	default:
		return fmt.Errorf("task queue full")
	}
}

// DeliverNextSegmentToViewer delivers the next available segment to a viewer
func (ds *DistributionService) DeliverNextSegmentToViewer(recordingID, viewerID string) (*VideoSegment, error) {
	distributor, err := ds.GetDistributor(recordingID)
	if err != nil {
		return nil, err
	}

	return distributor.GetNextSegment(viewerID)
}

// AdaptViewerQuality adapts bitrate based on network conditions
func (ds *DistributionService) AdaptViewerQuality(recordingID, viewerID string, bufferHealth float64) (string, error) {
	distributor, err := ds.GetDistributor(recordingID)
	if err != nil {
		return "", err
	}

	session, err := distributor.GetViewerSession(viewerID)
	if err != nil {
		return "", err
	}

	distributor.UpdateViewerBuffer(viewerID, bufferHealth)

	// Determine best bitrate based on buffer
	newBitrate := session.CurrentBitrate

	if bufferHealth < 20 {
		// Severe congestion - drop to lowest bitrate
		newBitrate = "VeryLow"
	} else if bufferHealth < 40 {
		// Moderate congestion - drop one level
		if session.CurrentBitrate == "High" {
			newBitrate = "Medium"
		} else if session.CurrentBitrate == "Medium" {
			newBitrate = "Low"
		}
	} else if bufferHealth > 85 {
		// Good conditions - upgrade bitrate
		if session.CurrentBitrate == "VeryLow" {
			newBitrate = "Low"
		} else if session.CurrentBitrate == "Low" {
			newBitrate = "Medium"
		} else if session.CurrentBitrate == "Medium" {
			newBitrate = "High"
		}
	}

	if newBitrate != session.CurrentBitrate {
		distributor.SwitchBitrate(viewerID, newBitrate)
	}

	return newBitrate, nil
}

// GetStreamViewers returns all active viewers for a stream
func (ds *DistributionService) GetStreamViewers(recordingID string) (map[string]*ViewerSession, error) {
	distributor, err := ds.GetDistributor(recordingID)
	if err != nil {
		return nil, err
	}

	return distributor.GetAllViewerSessions(), nil
}

// GetStreamStatistics returns distribution statistics for a stream
func (ds *DistributionService) GetStreamStatistics(recordingID string) (DistributionStats, error) {
	distributor, err := ds.GetDistributor(recordingID)
	if err != nil {
		return DistributionStats{}, err
	}

	return distributor.GetDistributionStats(), nil
}

// GetAllStreamStatistics returns statistics for all active streams
func (ds *DistributionService) GetAllStreamStatistics() []DistributionStats {
	ds.distributorsMu.RLock()
	defer ds.distributorsMu.RUnlock()

	var allStats []DistributionStats
	for _, distributor := range ds.distributors {
		allStats = append(allStats, distributor.GetDistributionStats())
	}

	return allStats
}

// EndLiveStream closes a live stream and cleans up resources
func (ds *DistributionService) EndLiveStream(recordingID string) error {
	ds.distributorsMu.Lock()
	defer ds.distributorsMu.Unlock()

	distributor, exists := ds.distributors[recordingID]
	if !exists {
		return fmt.Errorf("distributor not found: %s", recordingID)
	}

	err := distributor.Close()
	if err != nil {
		return err
	}

	delete(ds.distributors, recordingID)

	return nil
}

// EnableCDN enables CDN integration for segment delivery
func (ds *DistributionService) EnableCDN(endpoint string) {
	ds.cdnEnabled = true
	ds.cdnEndpoint = endpoint
	if ds.logger != nil {
		ds.logger.Printf("[Distribution] CDN enabled: %s\n", endpoint)
	}
}

// DisableCDN disables CDN integration
func (ds *DistributionService) DisableCDN() {
	ds.cdnEnabled = false
	if ds.logger != nil {
		ds.logger.Println("[Distribution] CDN disabled")
	}
}

// GetMetrics returns overall distribution metrics
func (ds *DistributionService) GetMetrics() DistributionMetrics {
	ds.metricsMu.RLock()
	defer ds.metricsMu.RUnlock()

	metrics := ds.metrics
	metrics.Timestamp = time.Now()

	ds.distributorsMu.RLock()
	defer ds.distributorsMu.RUnlock()

	metrics.TotalDistributors = len(ds.distributors)
	totalViewers := 0
	for _, distributor := range ds.distributors {
		stats := distributor.GetDistributionStats()
		totalViewers += stats.ActiveViewers
	}
	metrics.TotalActiveViewers = totalViewers

	return metrics
}

// Worker goroutines

// distributionWorker processes segment delivery tasks
func (ds *DistributionService) distributionWorker(worker *DistributionWorker) {
	defer ds.wg.Done()

	for {
		select {
		case task := <-worker.taskChan:
			if task != nil {
				ds.processDeliveryTask(task)
			}
		case <-worker.stopChan:
			return
		case <-ds.stopChan:
			return
		}
	}
}

// processDeliveryTask handles a single delivery task
func (ds *DistributionService) processDeliveryTask(task *SegmentDeliveryTask) {
	distributor, err := ds.GetDistributor(task.RecordingID)
	if err != nil {
		ds.recordDeliveryFailure()
		return
	}

	err = distributor.DeliverSegment(task.ViewerID, task.SegmentID)
	if err != nil {
		task.Attempts++
		if task.Attempts < task.MaxAttempts {
			// Retry
			select {
			case ds.workQueue <- task:
			case <-ds.stopChan:
			}
		}
		ds.recordDeliveryFailure()
	}
}

// cleanupWorker periodically cleans up expired segments
func (ds *DistributionService) cleanupWorker() {
	defer ds.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ds.distributorsMu.RLock()
			for _, distributor := range ds.distributors {
				distributor.segmentQueue.CleanupExpiredSegments()
			}
			ds.distributorsMu.RUnlock()
		case <-ds.stopChan:
			return
		}
	}
}

// recordDeliveryFailure updates failure metrics
func (ds *DistributionService) recordDeliveryFailure() {
	ds.metricsMu.Lock()
	defer ds.metricsMu.Unlock()

	ds.metrics.TotalDeliveryFailures++
}

// Stop gracefully shuts down the distribution service
func (ds *DistributionService) Stop() error {
	if ds.closed {
		return fmt.Errorf("service already stopped")
	}

	ds.closed = true

	// Signal all workers to stop
	close(ds.stopChan)

	// Wait for all workers to finish
	ds.wg.Wait()

	// Close all distributors
	ds.distributorsMu.Lock()
	for _, distributor := range ds.distributors {
		_ = distributor.Close()
	}
	ds.distributors = make(map[string]*LiveDistributor)
	ds.distributorsMu.Unlock()

	if ds.logger != nil {
		ds.logger.Println("[Distribution] Service stopped")
	}

	return nil
}
