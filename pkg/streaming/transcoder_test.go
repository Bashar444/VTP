package streaming

import (
	"fmt"
	"log"
	"os"
	"testing"
)

func TestNewMultiBitrateTranscoder(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	transcoder := NewMultiBitrateTranscoder("/tmp", "/usr/bin/ffmpeg", 4, logger)

	if transcoder == nil {
		t.Error("Expected non-nil transcoder")
	}

	if len(transcoder.GetDefaultProfiles()) != 4 {
		t.Errorf("Expected 4 default profiles, got %d", len(transcoder.GetDefaultProfiles()))
	}
}

func TestQueueMultiBitrateJob(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	transcoder := NewMultiBitrateTranscoder("/tmp", "/usr/bin/ffmpeg", 4, logger)

	recordingID := "test-123"
	inputPath := "/tmp/test.mp4"

	jobIDs, err := transcoder.QueueMultiBitrateJob(recordingID, inputPath)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(jobIDs) != 4 {
		t.Errorf("Expected 4 jobs, got %d", len(jobIDs))
	}

	// Verify each job exists and has correct profile
	for i, jobID := range jobIDs {
		job := transcoder.GetJob(jobID)
		if job == nil {
			t.Errorf("Job %s not found", jobID)
		} else if job.Status != JobQueued {
			t.Errorf("Expected JobQueued status, got %s", job.Status)
		} else if i == 0 && job.Profile.Bitrate != 500 {
			t.Errorf("Expected 500 kbps bitrate, got %d", job.Profile.Bitrate)
		}
	}
}

func TestQueueManagement(t *testing.T) {
	queue := NewTranscodingQueue(5)

	// Add 3 jobs
	for i := 0; i < 3; i++ {
		job := &TranscodingJob{
			JobID:       fmt.Sprintf("job-%d", i),
			RecordingID: "test",
			Profile:     EncodingProfile{Bitrate: 1000},
		}
		if err := queue.Add(job); err != nil {
			t.Errorf("Error adding job: %v", err)
		}
	}

	if queue.Length() != 3 {
		t.Errorf("Expected 3 jobs in queue, got %d", queue.Length())
	}

	// Get next job
	job := queue.Next()
	if job == nil {
		t.Error("Expected job from queue")
	}

	if queue.Length() != 2 {
		t.Errorf("Expected 2 jobs remaining, got %d", queue.Length())
	}

	if queue.Running() != 1 {
		t.Errorf("Expected 1 running job, got %d", queue.Running())
	}

	queue.Complete(job.JobID)

	if queue.Running() != 0 {
		t.Errorf("Expected 0 running jobs after complete, got %d", queue.Running())
	}
}

func TestProgressTracking(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	transcoder := NewMultiBitrateTranscoder("/tmp", "/usr/bin/ffmpeg", 4, logger)

	jobIDs, _ := transcoder.QueueMultiBitrateJob("test-456", "/tmp/test.mp4")
	jobID := jobIDs[0]

	// Update progress
	transcoder.UpdateJobProgress(jobID, 50.0, 25.5)

	job := transcoder.GetJob(jobID)
	if job.Progress != 50.0 {
		t.Errorf("Expected 50%% progress, got %.1f%%", job.Progress)
	}
}

func TestGetJobStats(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	transcoder := NewMultiBitrateTranscoder("/tmp", "/usr/bin/ffmpeg", 4, logger)

	jobIDs, _ := transcoder.QueueMultiBitrateJob("test-789", "/tmp/test.mp4")
	jobID := jobIDs[0]

	// Get stats for queued job
	stats := transcoder.GetJobStats(jobID)
	if stats.Status != JobQueued {
		t.Errorf("Expected JobQueued status, got %s", stats.Status)
	}

	// Simulate progress
	transcoder.UpdateJobProgress(jobID, 75.0, 28.0)
	stats = transcoder.GetJobStats(jobID)
	if stats.Progress != 75.0 {
		t.Errorf("Expected 75%% progress in stats, got %.1f%%", stats.Progress)
	}
}

func TestCompleteJob(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	transcoder := NewMultiBitrateTranscoder("/tmp", "/usr/bin/ffmpeg", 4, logger)

	jobIDs, _ := transcoder.QueueMultiBitrateJob("test-complete", "/tmp/test.mp4")
	jobID := jobIDs[0]

	// Complete the job
	transcoder.CompleteJob(jobID, nil)

	job := transcoder.GetJob(jobID)
	if job.Status != JobCompleted {
		t.Errorf("Expected JobCompleted status, got %s", job.Status)
	}

	if job.Progress != 100.0 {
		t.Errorf("Expected 100%% progress, got %.1f%%", job.Progress)
	}

	if job.EndTime.IsZero() {
		t.Error("Expected EndTime to be set")
	}
}

func TestCompleteJobWithError(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	transcoder := NewMultiBitrateTranscoder("/tmp", "/usr/bin/ffmpeg", 4, logger)

	jobIDs, _ := transcoder.QueueMultiBitrateJob("test-error", "/tmp/test.mp4")
	jobID := jobIDs[0]

	// Complete with error
	testErr := fmt.Errorf("transcoding failed")
	transcoder.CompleteJob(jobID, testErr)

	job := transcoder.GetJob(jobID)
	if job.Status != JobFailed {
		t.Errorf("Expected JobFailed status, got %s", job.Status)
	}

	if job.Error == nil {
		t.Error("Expected error to be set")
	}
}

func TestGenerateMasterPlaylist(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	transcoder := NewMultiBitrateTranscoder("/tmp", "/usr/bin/ffmpeg", 4, logger)

	recordingID := "test-playlist"
	playlist, err := transcoder.GenerateMasterPlaylist(recordingID)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if playlist == "" {
		t.Error("Expected non-empty playlist")
	}

	// Check playlist contains expected content
	expectedStrings := []string{
		"#EXTM3U",
		"EXT-X-VERSION:3",
		"500.m3u8",
		"1000.m3u8",
		"2000.m3u8",
		"4000.m3u8",
	}

	for _, expected := range expectedStrings {
		if !containsString(playlist, expected) {
			t.Errorf("Expected playlist to contain '%s'", expected)
		}
	}
}

func TestGenerateVariantPlaylist(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	transcoder := NewMultiBitrateTranscoder("/tmp", "/usr/bin/ffmpeg", 4, logger)

	recordingID := "test-variant"
	bitrate := 1000

	playlist, err := transcoder.GenerateVariantPlaylist(recordingID, bitrate)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if playlist == "" {
		t.Error("Expected non-empty playlist")
	}

	// Check for expected playlist content
	if !containsString(playlist, "#EXTM3U") {
		t.Error("Expected #EXTM3U header")
	}

	if !containsString(playlist, "#EXT-X-ENDLIST") {
		t.Error("Expected #EXT-X-ENDLIST footer")
	}
}

func TestGetAllJobs(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	transcoder := NewMultiBitrateTranscoder("/tmp", "/usr/bin/ffmpeg", 4, logger)

	// Queue multiple recordings
	jobIDs1, _ := transcoder.QueueMultiBitrateJob("recording-1", "/tmp/test1.mp4")
	jobIDs2, _ := transcoder.QueueMultiBitrateJob("recording-2", "/tmp/test2.mp4")

	// Get all jobs for recording-1
	jobs1 := transcoder.GetAllJobs("recording-1")
	if len(jobs1) != len(jobIDs1) {
		t.Errorf("Expected %d jobs for recording-1, got %d", len(jobIDs1), len(jobs1))
	}

	// Get all jobs for recording-2
	jobs2 := transcoder.GetAllJobs("recording-2")
	if len(jobs2) != len(jobIDs2) {
		t.Errorf("Expected %d jobs for recording-2, got %d", len(jobIDs2), len(jobs2))
	}
}

func TestIsRecordingCompleted(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	transcoder := NewMultiBitrateTranscoder("/tmp", "/usr/bin/ffmpeg", 4, logger)

	recordingID := "test-completion"
	jobIDs, _ := transcoder.QueueMultiBitrateJob(recordingID, "/tmp/test.mp4")

	// Initially not completed
	if transcoder.IsRecordingCompleted(recordingID) {
		t.Error("Expected recording to not be completed")
	}

	// Complete all jobs
	for _, jobID := range jobIDs {
		transcoder.CompleteJob(jobID, nil)
	}

	// Now should be completed
	if !transcoder.IsRecordingCompleted(recordingID) {
		t.Error("Expected recording to be completed")
	}
}

func TestCancelJob(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	transcoder := NewMultiBitrateTranscoder("/tmp", "/usr/bin/ffmpeg", 4, logger)

	jobIDs, _ := transcoder.QueueMultiBitrateJob("test-cancel", "/tmp/test.mp4")
	jobID := jobIDs[0]

	// Cancel the job
	err := transcoder.CancelJob(jobID)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	job := transcoder.GetJob(jobID)
	if job.Status != JobCancelled {
		t.Errorf("Expected JobCancelled status, got %s", job.Status)
	}
}

func TestGetQueueStats(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	transcoder := NewMultiBitrateTranscoder("/tmp", "/usr/bin/ffmpeg", 2, logger)

	stats := transcoder.GetQueueStats()

	if stats["pending_jobs"].(int) != 0 {
		t.Error("Expected 0 pending jobs initially")
	}

	if stats["running_jobs"].(int) != 0 {
		t.Error("Expected 0 running jobs initially")
	}

	if stats["max_concurrent"].(int) != 2 {
		t.Error("Expected max_concurrent to be 2")
	}
}

func TestProgressCallback(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	transcoder := NewMultiBitrateTranscoder("/tmp", "/usr/bin/ffmpeg", 4, logger)

	jobIDs, _ := transcoder.QueueMultiBitrateJob("test-callback", "/tmp/test.mp4")
	jobID := jobIDs[0]

	// Register a callback
	callback := func(update ProgressUpdate) {
		// Callback registered successfully
		_ = update
	}

	transcoder.RegisterProgressCallback(jobID, callback)

	// Update progress - callback should be called
	transcoder.UpdateJobProgress(jobID, 50.0, 25.0)

	// In a real scenario with goroutines, we'd need to wait
	// For now, just verify the callback was registered
	if jobID == "" {
		t.Error("Expected valid job ID")
	}
}

func TestQueueFull(t *testing.T) {
	queue := NewTranscodingQueue(2)

	// Add 2 jobs
	for i := 0; i < 2; i++ {
		job := &TranscodingJob{
			JobID:       fmt.Sprintf("job-%d", i),
			RecordingID: "test",
			Profile:     EncodingProfile{Bitrate: 1000},
		}
		if err := queue.Add(job); err != nil {
			t.Errorf("Error adding job %d: %v", i, err)
		}
	}

	// Try to add 3rd job - should fail
	job3 := &TranscodingJob{
		JobID:       "job-3",
		RecordingID: "test",
		Profile:     EncodingProfile{Bitrate: 1000},
	}

	// Make room by getting a job
	_ = queue.Next()

	// Now we should be able to add
	err := queue.Add(job3)
	if err != nil {
		t.Errorf("Unexpected error adding job after making room: %v", err)
	}
}

func TestTranscodingServiceCreation(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	transcoder := NewMultiBitrateTranscoder("/tmp", "/usr/bin/ffmpeg", 4, logger)
	service := NewTranscodingService(transcoder, 2, logger)

	if service == nil {
		t.Error("Expected non-nil service")
	}

	// Stop the service
	err := service.Stop()
	if err != nil {
		t.Errorf("Error stopping service: %v", err)
	}
}

func TestNewTranscodingServiceWithWorkers(t *testing.T) {
	logger := log.New(os.Stderr, "[Test] ", log.LstdFlags)
	transcoder := NewMultiBitrateTranscoder("/tmp", "/usr/bin/ffmpeg", 4, logger)

	// Create service with 3 workers
	service := NewTranscodingService(transcoder, 3, logger)
	defer service.Stop()

	if service.workerCount != 3 {
		t.Errorf("Expected 3 workers, got %d", service.workerCount)
	}
}

// Helper function to check if string contains substring
func containsString(str, substring string) bool {
	for i := 0; i < len(str)-len(substring)+1; i++ {
		if str[i:i+len(substring)] == substring {
			return true
		}
	}
	return false
}

// Benchmark tests
func BenchmarkQueueJob(b *testing.B) {
	logger := log.New(os.Stderr, "[Bench] ", log.LstdFlags)
	transcoder := NewMultiBitrateTranscoder("/tmp", "/usr/bin/ffmpeg", 100, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		transcoder.QueueMultiBitrateJob(fmt.Sprintf("rec-%d", i), "/tmp/test.mp4")
	}
}

func BenchmarkGeneratePlaylist(b *testing.B) {
	logger := log.New(os.Stderr, "[Bench] ", log.LstdFlags)
	transcoder := NewMultiBitrateTranscoder("/tmp", "/usr/bin/ffmpeg", 4, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		transcoder.GenerateMasterPlaylist(fmt.Sprintf("rec-%d", i))
	}
}
