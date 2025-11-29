package recording

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// FFmpegProcess represents an active FFmpeg recording process
type FFmpegProcess struct {
	mu             sync.Mutex
	cmd            *exec.Cmd
	stdin          io.WriteCloser
	stderr         io.ReadCloser
	recordingID    string
	outputPath     string
	status         string
	startedAt      time.Time
	bytesProcessed int64
	frameCount     int64
	lastError      string
	errorChan      chan error
	stopChan       chan bool
	logger         *log.Logger
	isRunning      bool
}

// FFmpegConfig contains FFmpeg configuration
type FFmpegConfig struct {
	OutputDir   string
	BitrateKbps int
	FrameRate   int
	Resolution  string
	VideoCodec  string
	AudioCodec  string
	Format      string
}

// DefaultFFmpegConfig returns default FFmpeg configuration
func DefaultFFmpegConfig() *FFmpegConfig {
	return &FFmpegConfig{
		OutputDir:   "/tmp/recordings",
		BitrateKbps: 2500,
		FrameRate:   30,
		Resolution:  "1920x1080",
		VideoCodec:  "libvpx",
		AudioCodec:  "libopus",
		Format:      "webm",
	}
}

// NewFFmpegProcess creates a new FFmpeg process
func NewFFmpegProcess(recordingID string, config *FFmpegConfig, logger *log.Logger) *FFmpegProcess {
	if config == nil {
		config = DefaultFFmpegConfig()
	}

	if logger == nil {
		logger = log.New(os.Stderr, "[FFmpeg] ", log.LstdFlags)
	}

	// Ensure output directory exists
	os.MkdirAll(config.OutputDir, 0755)

	outputPath := filepath.Join(
		config.OutputDir,
		fmt.Sprintf("%s.%s", recordingID, config.Format),
	)

	return &FFmpegProcess{
		recordingID: recordingID,
		outputPath:  outputPath,
		status:      "initializing",
		errorChan:   make(chan error, 10),
		stopChan:    make(chan bool),
		logger:      logger,
		isRunning:   false,
	}
}

// StartFFmpeg launches the FFmpeg process
func (fp *FFmpegProcess) StartFFmpeg(ctx context.Context, config *FFmpegConfig) error {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	if fp.isRunning {
		return fmt.Errorf("FFmpeg process already running for %s", fp.recordingID)
	}

	if config == nil {
		config = DefaultFFmpegConfig()
	}

	// Build FFmpeg command
	args := []string{
		"-f", "rawvideo",
		"-pixel_format", "yuv420p",
		"-video_size", config.Resolution,
		"-framerate", fmt.Sprintf("%d", config.FrameRate),
		"-i", "pipe:0", // Video from stdin
		"-f", "s16le", // Audio format: signed 16-bit little-endian
		"-sample_rate", "48000", // Standard audio sample rate
		"-channels", "2", // Stereo
		"-i", "pipe:1", // Audio from stdin
		"-c:v", config.VideoCodec,
		"-b:v", fmt.Sprintf("%dk", config.BitrateKbps),
		"-c:a", config.AudioCodec,
		"-b:a", "128k",
		"-f", config.Format,
		fp.outputPath,
	}

	fp.cmd = exec.CommandContext(ctx, "ffmpeg", args...)

	// Set up pipes
	var err error
	fp.stdin, err = fp.cmd.StdinPipe()
	if err != nil {
		fp.lastError = err.Error()
		fp.status = "failed"
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	fp.stderr, err = fp.cmd.StderrPipe()
	if err != nil {
		fp.lastError = err.Error()
		fp.status = "failed"
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the process
	err = fp.cmd.Start()
	if err != nil {
		fp.lastError = err.Error()
		fp.status = "failed"
		return fmt.Errorf("failed to start FFmpeg: %w", err)
	}

	fp.startedAt = time.Now()
	fp.status = "recording"
	fp.isRunning = true

	// Start monitoring goroutine
	go fp.monitorProcess()

	fp.logger.Printf("FFmpeg process started for recording %s, output: %s", fp.recordingID, fp.outputPath)
	return nil
}

// StopFFmpeg gracefully stops the FFmpeg process
func (fp *FFmpegProcess) StopFFmpeg(ctx context.Context) error {
	fp.mu.Lock()

	if !fp.isRunning {
		fp.mu.Unlock()
		return fmt.Errorf("FFmpeg process not running for %s", fp.recordingID)
	}

	isRunning := fp.isRunning
	fp.mu.Unlock()

	if !isRunning {
		return fmt.Errorf("FFmpeg process not running")
	}

	// Close stdin to signal end of input to FFmpeg
	if fp.stdin != nil {
		fp.stdin.Close()
	}

	// Wait for process to finish with timeout
	done := make(chan error, 1)
	go func() {
		done <- fp.cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil && err.Error() != "exit status 0" {
			fp.logger.Printf("FFmpeg exited with error: %v", err)
		}
	case <-ctx.Done():
		fp.cmd.Process.Kill()
		return fmt.Errorf("timeout waiting for FFmpeg to stop")
	}

	fp.mu.Lock()
	fp.status = "stopped"
	fp.isRunning = false
	fp.mu.Unlock()

	fp.logger.Printf("FFmpeg process stopped for recording %s", fp.recordingID)
	return nil
}

// WriteVideoFrame writes video frame data to FFmpeg
func (fp *FFmpegProcess) WriteVideoFrame(data []byte) error {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	if !fp.isRunning {
		return fmt.Errorf("FFmpeg process not running")
	}

	if fp.stdin == nil {
		return fmt.Errorf("stdin not available")
	}

	// Write video frame (file descriptor 0)
	_, err := fp.stdin.Write(data)
	if err != nil {
		fp.lastError = err.Error()
		fp.status = "error"
		return fmt.Errorf("failed to write video frame: %w", err)
	}

	fp.bytesProcessed += int64(len(data))
	fp.frameCount++

	return nil
}

// WriteAudioFrame writes audio frame data to FFmpeg
func (fp *FFmpegProcess) WriteAudioFrame(data []byte) error {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	if !fp.isRunning {
		return fmt.Errorf("FFmpeg process not running")
	}

	if fp.stdin == nil {
		return fmt.Errorf("stdin not available")
	}

	// Write audio frame (file descriptor 1)
	_, err := fp.stdin.Write(data)
	if err != nil {
		fp.lastError = err.Error()
		fp.status = "error"
		return fmt.Errorf("failed to write audio frame: %w", err)
	}

	fp.bytesProcessed += int64(len(data))

	return nil
}

// GetStatus returns current process status and statistics
func (fp *FFmpegProcess) GetStatus() map[string]interface{} {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	status := map[string]interface{}{
		"recording_id":    fp.recordingID,
		"status":          fp.status,
		"is_running":      fp.isRunning,
		"started_at":      fp.startedAt,
		"bytes_processed": fp.bytesProcessed,
		"frame_count":     fp.frameCount,
		"output_path":     fp.outputPath,
	}

	if fp.isRunning {
		duration := time.Since(fp.startedAt)
		status["duration_seconds"] = int(duration.Seconds())
	}

	if fp.lastError != "" {
		status["last_error"] = fp.lastError
	}

	// Try to get file size if file exists
	if fileInfo, err := os.Stat(fp.outputPath); err == nil {
		status["file_size_bytes"] = fileInfo.Size()
	}

	return status
}

// GetOutputPath returns the output file path
func (fp *FFmpegProcess) GetOutputPath() string {
	fp.mu.Lock()
	defer fp.mu.Unlock()
	return fp.outputPath
}

// monitorProcess monitors the FFmpeg process and logs errors
func (fp *FFmpegProcess) monitorProcess() {
	if fp.stderr == nil {
		return
	}

	// Read stderr output from FFmpeg
	scanner := io.ReadCloser(fp.stderr)
	data := make([]byte, 4096)

	for {
		n, err := scanner.Read(data)
		if err != nil {
			if err != io.EOF {
				fp.mu.Lock()
				fp.lastError = err.Error()
				fp.status = "error"
				fp.mu.Unlock()
			}
			break
		}

		// Log FFmpeg output for debugging
		if n > 0 {
			fp.logger.Printf("FFmpeg: %s", string(data[:n]))
		}
	}
}

// GetRecordingID returns the recording ID
func (fp *FFmpegProcess) GetRecordingID() string {
	fp.mu.Lock()
	defer fp.mu.Unlock()
	return fp.recordingID
}

// IsRunning checks if the process is currently running
func (fp *FFmpegProcess) IsRunning() bool {
	fp.mu.Lock()
	defer fp.mu.Unlock()
	return fp.isRunning
}

// GetLastError returns the last error message
func (fp *FFmpegProcess) GetLastError() string {
	fp.mu.Lock()
	defer fp.mu.Unlock()
	return fp.lastError
}

// Cleanup closes any open resources
func (fp *FFmpegProcess) Cleanup() error {
	fp.mu.Lock()
	defer fp.mu.Unlock()

	if fp.stdin != nil {
		fp.stdin.Close()
	}

	if fp.stderr != nil {
		fp.stderr.Close()
	}

	if fp.isRunning && fp.cmd != nil && fp.cmd.Process != nil {
		fp.cmd.Process.Kill()
	}

	return nil
}
