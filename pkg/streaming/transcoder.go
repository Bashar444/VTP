package streaming

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// EncodingProfile represents a bitrate encoding profile
type EncodingProfile struct {
	Bitrate    int    // kbps
	Resolution string // 1280x720, 1920x1080, etc.
	FrameRate  int    // 24, 30, etc.
	Label      string // "Low", "Medium", "High", etc.
}

// TranscodingJob represents a single transcoding job
type TranscodingJob struct {
	JobID       string          // Unique job identifier
	RecordingID string          // Which recording to transcode
	Profile     EncodingProfile // Target encoding profile
	Status      JobStatus       // Current status
	Progress    float64         // 0-100%
	StartTime   time.Time       // When job started
	EndTime     time.Time       // When job completed
	InputPath   string          // Path to source video
	OutputPath  string          // Path to output video
	Error       error           // Any error that occurred
}

// JobStatus represents the status of a transcoding job
type JobStatus string

const (
	JobQueued    JobStatus = "queued"
	JobRunning   JobStatus = "running"
	JobCompleted JobStatus = "completed"
	JobFailed    JobStatus = "failed"
	JobCancelled JobStatus = "cancelled"
)

// ProgressUpdate represents progress info for a job
type ProgressUpdate struct {
	JobID    string    `json:"job_id"`
	Status   JobStatus `json:"status"`
	Progress float64   `json:"progress"` // 0-100
	Speed    float64   `json:"speed"`    // FPS
	Duration int64     `json:"duration_ms"`
	ETA      int64     `json:"eta_ms"`
	Error    string    `json:"error,omitempty"`
}

// PlaylistInfo represents metadata for an HLS playlist
type PlaylistInfo struct {
	RecordingID   string
	Bitrate       int
	Resolution    string
	FrameRate     int
	SegmentLength int // seconds
	Segments      []SegmentInfo
}

// SegmentInfo represents an HLS segment
type SegmentInfo struct {
	SegmentID int
	Duration  float64 // seconds
	Bitrate   int     // kbps
	File      string  // filename
}

// TranscodingQueue manages pending encoding jobs
type TranscodingQueue struct {
	mu      sync.RWMutex
	queue   []*TranscodingJob
	maxSize int // max concurrent jobs
	running int // currently running jobs
}

// NewTranscodingQueue creates a new transcoding queue
func NewTranscodingQueue(maxSize int) *TranscodingQueue {
	return &TranscodingQueue{
		queue:   make([]*TranscodingJob, 0),
		maxSize: maxSize,
		running: 0,
	}
}

// Add adds a job to the queue
func (tq *TranscodingQueue) Add(job *TranscodingJob) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	if len(tq.queue) >= tq.maxSize {
		return fmt.Errorf("queue full (%d jobs pending)", len(tq.queue))
	}

	job.Status = JobQueued
	tq.queue = append(tq.queue, job)
	return nil
}

// Next gets the next job from the queue
func (tq *TranscodingQueue) Next() *TranscodingJob {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	if len(tq.queue) == 0 || tq.running >= tq.maxSize {
		return nil
	}

	job := tq.queue[0]
	tq.queue = tq.queue[1:]
	tq.running++
	return job
}

// Complete marks a job as complete
func (tq *TranscodingQueue) Complete(jobID string) {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	if tq.running > 0 {
		tq.running--
	}
}

// Length returns the number of pending jobs
func (tq *TranscodingQueue) Length() int {
	tq.mu.RLock()
	defer tq.mu.RUnlock()

	return len(tq.queue)
}

// Running returns the number of currently running jobs
func (tq *TranscodingQueue) Running() int {
	tq.mu.RLock()
	defer tq.mu.RUnlock()

	return tq.running
}

// MultiBitrateTranscoder manages multi-bitrate video transcoding
type MultiBitrateTranscoder struct {
	logger              *log.Logger
	queue               *TranscodingQueue
	jobs                map[string]*TranscodingJob // Track all jobs
	jobsMu              sync.RWMutex
	storageDir          string
	ffmpegPath          string
	defaultProfiles     []EncodingProfile
	maxConcurrentJobs   int
	progressCallbacks   map[string]func(ProgressUpdate)
	progressCallbacksMu sync.RWMutex
}

// NewMultiBitrateTranscoder creates a new multi-bitrate transcoder
func NewMultiBitrateTranscoder(storageDir, ffmpegPath string, maxConcurrent int, logger *log.Logger) *MultiBitrateTranscoder {
	mt := &MultiBitrateTranscoder{
		logger:            logger,
		queue:             NewTranscodingQueue(maxConcurrent),
		jobs:              make(map[string]*TranscodingJob),
		storageDir:        storageDir,
		ffmpegPath:        ffmpegPath,
		maxConcurrentJobs: maxConcurrent,
		progressCallbacks: make(map[string]func(ProgressUpdate)),
	}

	// Initialize default encoding profiles
	mt.defaultProfiles = []EncodingProfile{
		{Bitrate: 500, Resolution: "1280x720", FrameRate: 24, Label: "VeryLow"},
		{Bitrate: 1000, Resolution: "1280x720", FrameRate: 24, Label: "Low"},
		{Bitrate: 2000, Resolution: "1920x1080", FrameRate: 30, Label: "Medium"},
		{Bitrate: 4000, Resolution: "1920x1080", FrameRate: 30, Label: "High"},
	}

	mt.logger.Println("[Transcoder] Multi-bitrate transcoder initialized")
	return mt
}

// GetDefaultProfiles returns the default encoding profiles
func (mt *MultiBitrateTranscoder) GetDefaultProfiles() []EncodingProfile {
	return mt.defaultProfiles
}

// QueueMultiBitrateJob queues a job to encode multiple bitrates
func (mt *MultiBitrateTranscoder) QueueMultiBitrateJob(recordingID, inputPath string) ([]string, error) {
	if recordingID == "" || inputPath == "" {
		return nil, fmt.Errorf("invalid recording ID or input path")
	}

	jobIDs := make([]string, 0)

	// Queue a job for each default profile
	for _, profile := range mt.defaultProfiles {
		jobID := fmt.Sprintf("%s-%d", recordingID, profile.Bitrate)

		job := &TranscodingJob{
			JobID:       jobID,
			RecordingID: recordingID,
			Profile:     profile,
			Status:      JobQueued,
			Progress:    0,
			InputPath:   inputPath,
			OutputPath:  fmt.Sprintf("%s/%s_%d.mp4", mt.storageDir, recordingID, profile.Bitrate),
			StartTime:   time.Now(),
		}

		if err := mt.queue.Add(job); err != nil {
			mt.logger.Printf("[Transcoder] Error queuing job %s: %v", jobID, err)
			return nil, err
		}

		// Store job in jobs map
		mt.jobsMu.Lock()
		mt.jobs[jobID] = job
		mt.jobsMu.Unlock()

		jobIDs = append(jobIDs, jobID)
		mt.logger.Printf("[Transcoder] Queued job %s (%d kbps)", jobID, profile.Bitrate)
	}

	return jobIDs, nil
}

// GetJob retrieves a job by ID
func (mt *MultiBitrateTranscoder) GetJob(jobID string) *TranscodingJob {
	mt.jobsMu.RLock()
	defer mt.jobsMu.RUnlock()

	return mt.jobs[jobID]
}

// UpdateJobProgress updates the progress of a job
func (mt *MultiBitrateTranscoder) UpdateJobProgress(jobID string, progress float64, speed float64) {
	mt.jobsMu.Lock()
	if job, ok := mt.jobs[jobID]; ok {
		job.Progress = progress
	}
	mt.jobsMu.Unlock()

	// Call progress callback if registered
	mt.progressCallbacksMu.RLock()
	callback, ok := mt.progressCallbacks[jobID]
	mt.progressCallbacksMu.RUnlock()

	if ok && callback != nil {
		update := ProgressUpdate{
			JobID:    jobID,
			Status:   JobRunning,
			Progress: progress,
			Speed:    speed,
		}
		callback(update)
	}
}

// CompleteJob marks a job as completed
func (mt *MultiBitrateTranscoder) CompleteJob(jobID string, err error) {
	mt.jobsMu.Lock()
	if job, ok := mt.jobs[jobID]; ok {
		if err != nil {
			job.Status = JobFailed
			job.Error = err
		} else {
			job.Status = JobCompleted
			job.Progress = 100.0
		}
		job.EndTime = time.Now()
	}
	mt.jobsMu.Unlock()

	mt.queue.Complete(jobID)

	// Call progress callback
	mt.progressCallbacksMu.RLock()
	callback, ok := mt.progressCallbacks[jobID]
	mt.progressCallbacksMu.RUnlock()

	if ok && callback != nil {
		status := JobCompleted
		errStr := ""
		if err != nil {
			status = JobFailed
			errStr = err.Error()
		}
		update := ProgressUpdate{
			JobID:    jobID,
			Status:   status,
			Progress: 100.0,
			Error:    errStr,
		}
		callback(update)
	}

	if err != nil {
		mt.logger.Printf("[Transcoder] Job %s failed: %v", jobID, err)
	} else {
		mt.logger.Printf("[Transcoder] Job %s completed successfully", jobID)
	}
}

// RegisterProgressCallback registers a callback for job progress updates
func (mt *MultiBitrateTranscoder) RegisterProgressCallback(jobID string, callback func(ProgressUpdate)) {
	mt.progressCallbacksMu.Lock()
	defer mt.progressCallbacksMu.Unlock()

	mt.progressCallbacks[jobID] = callback
}

// GetQueueStats returns statistics about the transcoding queue
func (mt *MultiBitrateTranscoder) GetQueueStats() map[string]interface{} {
	return map[string]interface{}{
		"pending_jobs":    mt.queue.Length(),
		"running_jobs":    mt.queue.Running(),
		"total_jobs":      len(mt.jobs),
		"max_concurrent":  mt.maxConcurrentJobs,
		"utilization_pct": float64(mt.queue.Running()) / float64(mt.maxConcurrentJobs) * 100,
	}
}

// GetJobStats returns detailed stats for a specific job
func (mt *MultiBitrateTranscoder) GetJobStats(jobID string) ProgressUpdate {
	mt.jobsMu.RLock()
	job, ok := mt.jobs[jobID]
	mt.jobsMu.RUnlock()

	if !ok {
		return ProgressUpdate{
			JobID:  jobID,
			Status: JobFailed,
			Error:  "job not found",
		}
	}

	errStr := ""
	if job.Error != nil {
		errStr = job.Error.Error()
	}

	duration := job.EndTime.Sub(job.StartTime).Milliseconds()
	if duration == 0 && job.Status == JobRunning {
		duration = time.Since(job.StartTime).Milliseconds()
	}

	// Estimate ETA
	eta := int64(0)
	if job.Progress > 0 && job.Progress < 100 {
		totalTime := duration / int64(job.Progress)
		eta = totalTime - duration
	}

	return ProgressUpdate{
		JobID:    jobID,
		Status:   job.Status,
		Progress: job.Progress,
		Duration: duration,
		ETA:      eta,
		Error:    errStr,
	}
}

// GenerateMasterPlaylist creates a master M3U8 playlist linking all bitrates
func (mt *MultiBitrateTranscoder) GenerateMasterPlaylist(recordingID string) (string, error) {
	if recordingID == "" {
		return "", fmt.Errorf("invalid recording ID")
	}

	// Build master playlist content
	playlist := "#EXTM3U\n"
	playlist += "#EXT-X-VERSION:3\n"
	playlist += "#EXT-X-STREAM-INF:BANDWIDTH=500000,RESOLUTION=1280x720,FRAME-RATE=24\n"
	playlist += fmt.Sprintf("%s_500.m3u8\n", recordingID)

	playlist += "#EXT-X-STREAM-INF:BANDWIDTH=1000000,RESOLUTION=1280x720,FRAME-RATE=24\n"
	playlist += fmt.Sprintf("%s_1000.m3u8\n", recordingID)

	playlist += "#EXT-X-STREAM-INF:BANDWIDTH=2000000,RESOLUTION=1920x1080,FRAME-RATE=30\n"
	playlist += fmt.Sprintf("%s_2000.m3u8\n", recordingID)

	playlist += "#EXT-X-STREAM-INF:BANDWIDTH=4000000,RESOLUTION=1920x1080,FRAME-RATE=30\n"
	playlist += fmt.Sprintf("%s_4000.m3u8\n", recordingID)

	mt.logger.Printf("[Transcoder] Generated master playlist for recording %s", recordingID)
	return playlist, nil
}

// GenerateVariantPlaylist creates an HLS variant playlist for a specific bitrate
func (mt *MultiBitrateTranscoder) GenerateVariantPlaylist(recordingID string, bitrate int) (string, error) {
	if recordingID == "" || bitrate <= 0 {
		return "", fmt.Errorf("invalid recording ID or bitrate")
	}

	// Build variant playlist content
	// In a real implementation, this would parse the actual encoded file
	playlist := "#EXTM3U\n"
	playlist += "#EXT-X-VERSION:3\n"
	playlist += "#EXT-X-TARGETDURATION:10\n"
	playlist += "#EXT-X-MEDIA-SEQUENCE:0\n"

	// Add sample segments
	for i := 0; i < 10; i++ {
		playlist += fmt.Sprintf("#EXTINF:10.0,\n%s_%d_segment_%d.ts\n", recordingID, bitrate, i)
	}

	playlist += "#EXT-X-ENDLIST\n"

	mt.logger.Printf("[Transcoder] Generated variant playlist for recording %s at %d kbps", recordingID, bitrate)
	return playlist, nil
}

// CancelJob cancels a pending or running job
func (mt *MultiBitrateTranscoder) CancelJob(jobID string) error {
	mt.jobsMu.Lock()
	if job, ok := mt.jobs[jobID]; ok {
		if job.Status == JobQueued || job.Status == JobRunning {
			job.Status = JobCancelled
			job.EndTime = time.Now()
		}
		mt.jobsMu.Unlock()

		mt.logger.Printf("[Transcoder] Job %s cancelled", jobID)
		return nil
	}
	mt.jobsMu.Unlock()

	return fmt.Errorf("job not found")
}

// GetAllJobs returns all jobs for a recording
func (mt *MultiBitrateTranscoder) GetAllJobs(recordingID string) []*TranscodingJob {
	mt.jobsMu.RLock()
	defer mt.jobsMu.RUnlock()

	var jobs []*TranscodingJob
	for _, job := range mt.jobs {
		if job.RecordingID == recordingID {
			jobs = append(jobs, job)
		}
	}
	return jobs
}

// IsRecordingCompleted checks if all encoding jobs for a recording are done
func (mt *MultiBitrateTranscoder) IsRecordingCompleted(recordingID string) bool {
	mt.jobsMu.RLock()
	defer mt.jobsMu.RUnlock()

	recordingJobs := 0
	completedJobs := 0

	for _, job := range mt.jobs {
		if job.RecordingID == recordingID {
			recordingJobs++
			if job.Status == JobCompleted {
				completedJobs++
			}
		}
	}

	return recordingJobs > 0 && recordingJobs == completedJobs
}
