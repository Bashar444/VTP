package streaming

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
)

// TranscodingService manages the transcoding pipeline
type TranscodingService struct {
	transcoder   *MultiBitrateTranscoder
	logger       *log.Logger
	activeJobs   map[string]bool
	activeJobsMu sync.RWMutex
	workerCount  int
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

// NewTranscodingService creates a new transcoding service
func NewTranscodingService(transcoder *MultiBitrateTranscoder, workerCount int, logger *log.Logger) *TranscodingService {
	ts := &TranscodingService{
		transcoder:  transcoder,
		logger:      logger,
		activeJobs:  make(map[string]bool),
		workerCount: workerCount,
		stopChan:    make(chan struct{}),
	}

	// Start worker goroutines
	for i := 0; i < workerCount; i++ {
		ts.wg.Add(1)
		go ts.transcodingWorker(i)
	}

	ts.logger.Printf("[TranscodingService] Service started with %d workers", workerCount)
	return ts
}

// transcodingWorker processes transcoding jobs from the queue
func (ts *TranscodingService) transcodingWorker(workerID int) {
	defer ts.wg.Done()

	for {
		select {
		case <-ts.stopChan:
			ts.logger.Printf("[TranscodingService] Worker %d stopping", workerID)
			return
		default:
			// Try to get a job from the queue
			job := ts.transcoder.queue.Next()
			if job == nil {
				// No jobs available, sleep briefly
				select {
				case <-ts.stopChan:
					return
				default:
					// Sleep for a short duration
					// In real implementation, use time.Sleep(100 * time.Millisecond)
				}
				continue
			}

			// Process the job
			ts.processTranscodingJob(job, workerID)
		}
	}
}

// processTranscodingJob executes an actual transcoding job
func (ts *TranscodingService) processTranscodingJob(job *TranscodingJob, workerID int) {
	job.Status = JobRunning
	ts.logger.Printf("[TranscodingService:W%d] Starting job %s (%d kbps)", workerID, job.JobID, job.Profile.Bitrate)

	// Mark as active
	ts.activeJobsMu.Lock()
	ts.activeJobs[job.JobID] = true
	ts.activeJobsMu.Unlock()

	// In a real implementation, this would call FFmpeg
	// For now, we'll simulate the transcoding process
	err := ts.simulateTranscoding(job)

	// Update job status
	ts.transcoder.CompleteJob(job.JobID, err)

	// Mark as inactive
	ts.activeJobsMu.Lock()
	delete(ts.activeJobs, job.JobID)
	ts.activeJobsMu.Unlock()
}

// simulateTranscoding simulates the transcoding process for testing
func (ts *TranscodingService) simulateTranscoding(job *TranscodingJob) error {
	// In a real implementation, this would call:
	// ffmpeg -i input.mp4 -b:v {bitrate}k -r {fps} -s {resolution} output.mp4

	// Simulate encoding progress
	for progress := float64(0); progress < 100; progress += 10 {
		ts.transcoder.UpdateJobProgress(job.JobID, progress, 30.0) // 30 FPS
		// Sleep to simulate encoding time
		// In real: time.Sleep(100 * time.Millisecond)
	}

	ts.transcoder.UpdateJobProgress(job.JobID, 100.0, 30.0)

	// Create dummy output file (in production, this would be the actual video)
	if err := os.WriteFile(job.OutputPath, []byte("dummy video data"), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	ts.logger.Printf("[TranscodingService] Transcoded to %s", job.OutputPath)
	return nil
}

// StartMultiBitrateEncoding queues a recording for multi-bitrate transcoding
func (ts *TranscodingService) StartMultiBitrateEncoding(recordingID, inputPath string) ([]string, error) {
	if recordingID == "" || inputPath == "" {
		return nil, fmt.Errorf("invalid recording ID or input path")
	}

	// Queue jobs for all profiles
	jobIDs, err := ts.transcoder.QueueMultiBitrateJob(recordingID, inputPath)
	if err != nil {
		ts.logger.Printf("[TranscodingService] Error starting encoding for %s: %v", recordingID, err)
		return nil, err
	}

	ts.logger.Printf("[TranscodingService] Queued %d encoding jobs for recording %s", len(jobIDs), recordingID)
	return jobIDs, nil
}

// GetTranscodingProgress returns progress for a specific job
func (ts *TranscodingService) GetTranscodingProgress(jobID string) ProgressUpdate {
	return ts.transcoder.GetJobStats(jobID)
}

// GetRecordingTranscodingStatus returns status for all jobs of a recording
func (ts *TranscodingService) GetRecordingTranscodingStatus(recordingID string) map[string]interface{} {
	jobs := ts.transcoder.GetAllJobs(recordingID)

	status := map[string]interface{}{
		"recording_id": recordingID,
		"total_jobs":   len(jobs),
		"jobs":         make([]map[string]interface{}, 0),
	}

	completedCount := 0
	failedCount := 0
	avgProgress := float64(0)

	for _, job := range jobs {
		jobInfo := map[string]interface{}{
			"job_id":   job.JobID,
			"bitrate":  job.Profile.Bitrate,
			"status":   job.Status,
			"progress": job.Progress,
			"output":   job.OutputPath,
		}

		if job.Status == JobCompleted {
			completedCount++
		} else if job.Status == JobFailed {
			failedCount++
		}
		avgProgress += job.Progress

		jobs_list := status["jobs"].([]map[string]interface{})
		jobs_list = append(jobs_list, jobInfo)
		status["jobs"] = jobs_list
	}

	if len(jobs) > 0 {
		avgProgress /= float64(len(jobs))
		status["average_progress"] = avgProgress
		status["completed_count"] = completedCount
		status["failed_count"] = failedCount
		status["is_complete"] = completedCount == len(jobs) && failedCount == 0
	}

	return status
}

// GeneratePlaylistsForRecording generates all HLS playlists for a completed encoding
func (ts *TranscodingService) GeneratePlaylistsForRecording(recordingID string) (map[string]string, error) {
	playlists := make(map[string]string)

	// Generate master playlist
	masterPlaylist, err := ts.transcoder.GenerateMasterPlaylist(recordingID)
	if err != nil {
		ts.logger.Printf("[TranscodingService] Error generating master playlist: %v", err)
		return nil, err
	}
	playlists["master"] = masterPlaylist

	// Generate variant playlists for each bitrate
	for _, profile := range ts.transcoder.GetDefaultProfiles() {
		variantPlaylist, err := ts.transcoder.GenerateVariantPlaylist(recordingID, profile.Bitrate)
		if err != nil {
			ts.logger.Printf("[TranscodingService] Error generating variant playlist: %v", err)
			return nil, err
		}

		playlistKey := fmt.Sprintf("variant_%d", profile.Bitrate)
		playlists[playlistKey] = variantPlaylist
	}

	ts.logger.Printf("[TranscodingService] Generated playlists for recording %s", recordingID)
	return playlists, nil
}

// CancelRecordingEncoding cancels all jobs for a recording
func (ts *TranscodingService) CancelRecordingEncoding(recordingID string) error {
	jobs := ts.transcoder.GetAllJobs(recordingID)

	for _, job := range jobs {
		if err := ts.transcoder.CancelJob(job.JobID); err != nil {
			ts.logger.Printf("[TranscodingService] Error cancelling job %s: %v", job.JobID, err)
		}
	}

	ts.logger.Printf("[TranscodingService] Cancelled all encoding jobs for recording %s", recordingID)
	return nil
}

// GetQueueStats returns statistics about the transcoding queue
func (ts *TranscodingService) GetQueueStats() map[string]interface{} {
	stats := ts.transcoder.GetQueueStats()
	stats["active_jobs"] = len(ts.activeJobs)
	stats["workers"] = ts.workerCount
	return stats
}

// Stop gracefully stops the transcoding service
func (ts *TranscodingService) Stop() error {
	ts.logger.Println("[TranscodingService] Stopping service...")
	close(ts.stopChan)
	ts.wg.Wait()
	ts.logger.Println("[TranscodingService] Service stopped")
	return nil
}

// ExecuteFFmpegTranscoding executes actual FFmpeg transcoding
// This is a helper function that would be called by processTranscodingJob
func ExecuteFFmpegTranscoding(ffmpegPath, inputPath, outputPath string, bitrate int, resolution string, fps int) error {
	// Build FFmpeg command
	args := []string{
		"-i", inputPath,
		"-b:v", fmt.Sprintf("%dk", bitrate),
		"-s", resolution,
		"-r", fmt.Sprintf("%d", fps),
		"-c:v", "libx264",
		"-preset", "medium",
		"-y", // Overwrite output file
		outputPath,
	}

	cmd := exec.Command(ffmpegPath, args...)

	// Execute command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg transcoding failed: %w", err)
	}

	return nil
}
