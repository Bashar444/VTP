package recording

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"os/exec"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// StreamingFormat defines output format for streaming
type StreamingFormat string

const (
	FormatHLS  StreamingFormat = "hls"
	FormatDASH StreamingFormat = "dash"
	FormatMP4  StreamingFormat = "mp4"
)

// StreamingProfile contains streaming configuration
type StreamingProfile struct {
	Format          StreamingFormat
	VideoCodec      string
	AudioCodec      string
	Bitrates        []int // kbps
	ResolutionHigh  string
	ResolutionLow   string
	SegmentDuration int
}

// DefaultHLSProfile provides HLS streaming configuration
var DefaultHLSProfile = StreamingProfile{
	Format:          FormatHLS,
	VideoCodec:      "libx264",
	AudioCodec:      "aac",
	Bitrates:        []int{500, 1000, 2000, 4000},
	ResolutionHigh:  "1920x1080",
	ResolutionLow:   "1280x720",
	SegmentDuration: 10,
}

// StreamingManager manages recording streaming and playback
type StreamingManager struct {
	storageManager *StorageManager
	db             *sql.DB
	logger         *log.Logger
	outputPath     string
}

// NewStreamingManager creates new streaming manager
func NewStreamingManager(storageManager *StorageManager, db *sql.DB, logger *log.Logger, outputPath string) *StreamingManager {
	return &StreamingManager{
		storageManager: storageManager,
		db:             db,
		logger:         logger,
		outputPath:     outputPath,
	}
}

// TranscodeToHLS converts recording to HLS format for streaming
func (m *StreamingManager) TranscodeToHLS(ctx context.Context, recordingID uuid.UUID, inputPath string, profile StreamingProfile) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	outputDir := fmt.Sprintf("%s/%s", m.outputPath, recordingID.String())
	playlistPath := fmt.Sprintf("%s/playlist.m3u8", outputDir)

	// Build FFmpeg command for HLS transcoding
	cmd := exec.CommandContext(ctx,
		"ffmpeg",
		"-i", inputPath,
		"-c:v", profile.VideoCodec,
		"-c:a", profile.AudioCodec,
		"-hls_time", strconv.Itoa(profile.SegmentDuration),
		"-hls_playlist_type", "vod",
		"-hls_segment_filename", fmt.Sprintf("%s/segment-%%03d.ts", outputDir),
		playlistPath,
	)

	if err := cmd.Run(); err != nil {
		m.logger.Printf("FFmpeg HLS transcoding failed: %v", err)
		return err
	}

	// Update database with streaming status
	query := `
		UPDATE recordings
		SET streaming_format = $1, streaming_ready = true, streaming_generated_at = $2, updated_at = $3
		WHERE id = $4
	`

	now := time.Now().UTC()
	_, err := m.db.ExecContext(ctx, query, string(FormatHLS), now, now, recordingID)
	if err != nil {
		m.logger.Printf("Failed to update streaming status: %v", err)
	}

	m.logger.Printf("HLS transcoding complete: %s", recordingID)
	return nil
}

// TranscodeToMP4 converts recording to MP4 format
func (m *StreamingManager) TranscodeToMP4(ctx context.Context, recordingID uuid.UUID, inputPath string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	outputPath := fmt.Sprintf("%s/%s.mp4", m.outputPath, recordingID.String())

	cmd := exec.CommandContext(ctx,
		"ffmpeg",
		"-i", inputPath,
		"-c:v", "libx264",
		"-crf", "23",
		"-c:a", "aac",
		"-b:a", "128k",
		outputPath,
	)

	if err := cmd.Run(); err != nil {
		m.logger.Printf("FFmpeg MP4 transcoding failed: %v", err)
		return err
	}

	return nil
}

// GenerateThumbnail creates thumbnail from recording
func (m *StreamingManager) GenerateThumbnail(ctx context.Context, recordingID uuid.UUID, inputPath string, timestampSeconds int) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	thumbPath := fmt.Sprintf("%s/%s_thumb.jpg", m.outputPath, recordingID.String())

	cmd := exec.CommandContext(ctx,
		"ffmpeg",
		"-ss", strconv.Itoa(timestampSeconds),
		"-i", inputPath,
		"-vf", "scale=320:180",
		"-vframes", "1",
		thumbPath,
	)

	if err := cmd.Run(); err != nil {
		m.logger.Printf("Failed to generate thumbnail: %v", err)
		return "", err
	}

	// Save thumbnail path to database
	query := `
		UPDATE recordings
		SET thumbnail_path = $1, updated_at = $2
		WHERE id = $3
	`

	now := time.Now().UTC()
	_, err := m.db.ExecContext(ctx, query, thumbPath, now, recordingID)
	if err != nil {
		m.logger.Printf("Failed to save thumbnail path: %v", err)
	}

	m.logger.Printf("Thumbnail generated: %s", recordingID)
	return thumbPath, nil
}

// ExtractMetadata extracts recording metadata using FFprobe
func (m *StreamingManager) ExtractMetadata(ctx context.Context, recordingID uuid.UUID, inputPath string) (*RecordingMetadata, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Use FFprobe to get metadata (simplified - real implementation would parse JSON output)
	cmd := exec.CommandContext(ctx,
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration,bit_rate",
		"-of", "default=noprint_wrappers=1:nokey=1:nokey=0",
		inputPath,
	)

	output, err := cmd.Output()
	if err != nil {
		m.logger.Printf("Failed to extract metadata: %v", err)
		return nil, err
	}

	metadata := &RecordingMetadata{
		RecordingID: recordingID,
		ExtractedAt: time.Now().UTC(),
		RawData:     string(output),
	}

	return metadata, nil
}

// PlaybackSession represents an active playback session
type PlaybackSession struct {
	ID            uuid.UUID
	RecordingID   uuid.UUID
	UserID        uuid.UUID
	StartedAt     time.Time
	LastHeartbeat time.Time
	Position      int64 // seconds
	Quality       string
	UserAgent     string
}

// PlaybackAnalytics tracks recording playback analytics
type PlaybackAnalytics struct {
	RecordingID     uuid.UUID
	TotalSessions   int
	TotalPlayTime   int64 // seconds
	AveragePlayTime int64 // seconds
	UniqueViewers   int
	CompletionRate  float64 // percentage
	LastAccessedAt  time.Time
	LastAccessedBy  uuid.UUID
}

// RecordingMetadata contains metadata about a recording
type RecordingMetadata struct {
	RecordingID   uuid.UUID
	Duration      int64 // seconds
	Bitrate       int64 // kbps
	VideoCodec    string
	AudioCodec    string
	Resolution    string
	FrameRate     string
	ThumbnailPath *string
	ExtractedAt   time.Time
	RawData       string
}

// GetPlaybackAnalytics retrieves playback analytics for a recording
func (m *StreamingManager) GetPlaybackAnalytics(ctx context.Context, recordingID uuid.UUID) (*PlaybackAnalytics, error) {
	query := `
		SELECT 
			COUNT(DISTINCT session_id) as total_sessions,
			COUNT(DISTINCT user_id) as unique_viewers,
			SUM(CAST(metadata->>'duration' AS BIGINT)) as total_playtime,
			MAX(accessed_at) as last_accessed_at
		FROM recording_access_log
		WHERE recording_id = $1 AND action = 'playback'
	`

	var totalSessions int
	var uniqueViewers int
	var totalPlayTime sql.NullInt64
	var lastAccessedAt sql.NullTime

	err := m.db.QueryRowContext(ctx, query, recordingID).Scan(
		&totalSessions,
		&uniqueViewers,
		&totalPlayTime,
		&lastAccessedAt,
	)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	analytics := &PlaybackAnalytics{
		RecordingID:   recordingID,
		TotalSessions: totalSessions,
		UniqueViewers: uniqueViewers,
	}

	if totalPlayTime.Valid {
		analytics.TotalPlayTime = totalPlayTime.Int64
		if totalSessions > 0 {
			analytics.AveragePlayTime = analytics.TotalPlayTime / int64(totalSessions)
		}
	}

	if lastAccessedAt.Valid {
		analytics.LastAccessedAt = lastAccessedAt.Time
	}

	return analytics, nil
}

// LogPlaybackEvent logs a playback event for analytics
func (m *StreamingManager) LogPlaybackEvent(ctx context.Context, recordingID uuid.UUID, userID uuid.UUID, eventType string, metadata map[string]interface{}) error {
	query := `
		INSERT INTO recording_access_log (id, recording_id, user_id, action, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	// Convert metadata to JSON string (simplified)
	metadataStr := "{}"
	if metadata != nil {
		if pos, ok := metadata["position"]; ok {
			metadataStr = fmt.Sprintf(`{"position":%d}`, pos)
		}
	}

	_, err := m.db.ExecContext(ctx, query, uuid.New(), recordingID, userID, eventType, metadataStr, time.Now().UTC())
	return err
}

// GeneratePlaylistURL creates a URL for HLS playlist
func (m *StreamingManager) GeneratePlaylistURL(ctx context.Context, recordingID uuid.UUID) (string, error) {
	// Check if streaming is ready
	query := `SELECT streaming_ready FROM recordings WHERE id = $1`

	var ready bool
	err := m.db.QueryRowContext(ctx, query, recordingID).Scan(&ready)
	if err != nil {
		return "", err
	}

	if !ready {
		return "", fmt.Errorf("streaming not ready for recording: %s", recordingID)
	}

	// Return playlist URL
	url := fmt.Sprintf("/api/v1/recordings/%s/stream/playlist.m3u8", recordingID.String())
	return url, nil
}

// CalculateQuality determines quality based on bitrate
func (m *StreamingManager) CalculateQuality(bitrateMbps float64) string {
	switch {
	case bitrateMbps < 1:
		return "low"
	case bitrateMbps < 2.5:
		return "medium"
	case bitrateMbps < 5:
		return "high"
	default:
		return "ultra"
	}
}

// CalculateBufferHealth calculates buffer health percentage (0-100)
func (m *StreamingManager) CalculateBufferHealth(bufferedSeconds int, totalDuration int) float64 {
	if totalDuration == 0 {
		return 0
	}
	return math.Min(100, (float64(bufferedSeconds)/float64(totalDuration))*100)
}

// EstimateBandwidth estimates required bandwidth for quality
func (m *StreamingManager) EstimateBandwidth(quality string) int {
	bandwidths := map[string]int{
		"low":    500,  // kbps
		"medium": 1500, // kbps
		"high":   3000, // kbps
		"ultra":  6000, // kbps
	}

	if bw, ok := bandwidths[quality]; ok {
		return bw
	}
	return 1500 // default
}
