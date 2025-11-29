package recording

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ParticipantManager manages recording participants and tracks statistics
type ParticipantManager struct {
	recordingID uuid.UUID
	db          *sql.DB
	logger      *log.Logger

	mu           sync.RWMutex
	participants map[uuid.UUID]*ParticipantStats
}

// ParticipantStats tracks statistics for a recording participant
type ParticipantStats struct {
	UserID           uuid.UUID
	JoinedAt         time.Time
	LeftAt           *time.Time
	VideoFrames      int64
	AudioFrames      int64
	VideoBytesIn     int64
	AudioBytesIn     int64
	VideoPacketsIn   int64
	AudioPacketsIn   int64
	VideoPacketsLost int64
	AudioPacketsLost int64
	LastUpdate       time.Time
}

// NewParticipantManager creates a new participant manager
func NewParticipantManager(recordingID uuid.UUID, db *sql.DB, logger *log.Logger) *ParticipantManager {
	return &ParticipantManager{
		recordingID:  recordingID,
		db:           db,
		logger:       logger,
		participants: make(map[uuid.UUID]*ParticipantStats),
	}
}

// AddParticipant adds a participant to the recording
func (pm *ParticipantManager) AddParticipant(ctx context.Context, userID uuid.UUID) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if participant already exists
	if _, exists := pm.participants[userID]; exists {
		return fmt.Errorf("participant %s already in recording", userID)
	}

	// Initialize participant stats
	stats := &ParticipantStats{
		UserID:     userID,
		JoinedAt:   time.Now().UTC(),
		LastUpdate: time.Now().UTC(),
	}

	pm.participants[userID] = stats

	// Persist to database
	if err := pm.persistParticipant(ctx, userID); err != nil {
		pm.logger.Printf("Failed to persist participant %s: %v", userID, err)
		// Don't return error - continue with in-memory tracking
	}

	pm.logger.Printf("Participant added: %s to recording %s", userID, pm.recordingID)
	return nil
}

// RemoveParticipant removes a participant from the recording
func (pm *ParticipantManager) RemoveParticipant(ctx context.Context, userID uuid.UUID) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Find and update participant
	stats, exists := pm.participants[userID]
	if !exists {
		return fmt.Errorf("participant %s not found in recording", userID)
	}

	now := time.Now().UTC()
	stats.LeftAt = &now

	// Persist removal to database
	if err := pm.persistParticipantRemoval(ctx, userID, now); err != nil {
		pm.logger.Printf("Failed to persist participant removal: %v", err)
	}

	// Keep in map for stats tracking but mark as left
	pm.logger.Printf("Participant removed: %s from recording %s", userID, pm.recordingID)
	return nil
}

// UpdateParticipantStats updates statistics for a participant
func (pm *ParticipantManager) UpdateParticipantStats(ctx context.Context, userID uuid.UUID, stats *ParticipantStatsUpdate) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	participant, exists := pm.participants[userID]
	if !exists {
		return fmt.Errorf("participant %s not found in recording", userID)
	}

	// Update stats
	if stats.VideoFrames > 0 {
		participant.VideoFrames = stats.VideoFrames
	}
	if stats.AudioFrames > 0 {
		participant.AudioFrames = stats.AudioFrames
	}
	if stats.VideoBytesIn > 0 {
		participant.VideoBytesIn = stats.VideoBytesIn
	}
	if stats.AudioBytesIn > 0 {
		participant.AudioBytesIn = stats.AudioBytesIn
	}
	if stats.VideoPacketsIn > 0 {
		participant.VideoPacketsIn = stats.VideoPacketsIn
	}
	if stats.AudioPacketsIn > 0 {
		participant.AudioPacketsIn = stats.AudioPacketsIn
	}
	if stats.VideoPacketsLost >= 0 {
		participant.VideoPacketsLost = stats.VideoPacketsLost
	}
	if stats.AudioPacketsLost >= 0 {
		participant.AudioPacketsLost = stats.AudioPacketsLost
	}

	participant.LastUpdate = time.Now().UTC()

	// Persist to database (periodic updates)
	if err := pm.persistParticipantStats(ctx, userID, participant); err != nil {
		pm.logger.Printf("Failed to persist participant stats: %v", err)
	}

	return nil
}

// GetParticipants returns all participants in the recording
func (pm *ParticipantManager) GetParticipants(ctx context.Context) ([]*ParticipantStats, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	participants := make([]*ParticipantStats, 0, len(pm.participants))
	for _, stats := range pm.participants {
		participants = append(participants, stats)
	}

	return participants, nil
}

// GetParticipant returns stats for a specific participant
func (pm *ParticipantManager) GetParticipant(ctx context.Context, userID uuid.UUID) (*ParticipantStats, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	stats, exists := pm.participants[userID]
	if !exists {
		return nil, fmt.Errorf("participant %s not found", userID)
	}

	return stats, nil
}

// GetParticipantCount returns the number of participants in the recording
func (pm *ParticipantManager) GetParticipantCount(ctx context.Context) int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return len(pm.participants)
}

// GetActiveParticipantCount returns the number of active (not left) participants
func (pm *ParticipantManager) GetActiveParticipantCount(ctx context.Context) int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	count := 0
	for _, stats := range pm.participants {
		if stats.LeftAt == nil {
			count++
		}
	}

	return count
}

// GetRecordingStats returns aggregate statistics for the recording
func (pm *ParticipantManager) GetRecordingStats(ctx context.Context) *RecordingAggregateStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	stats := &RecordingAggregateStats{
		TotalParticipants:  len(pm.participants),
		ActiveParticipants: 0,
	}

	for _, p := range pm.participants {
		if p.LeftAt == nil {
			stats.ActiveParticipants++
		}

		stats.TotalVideoFrames += p.VideoFrames
		stats.TotalAudioFrames += p.AudioFrames
		stats.TotalVideoBytesIn += p.VideoBytesIn
		stats.TotalAudioBytesIn += p.AudioBytesIn
		stats.TotalVideoPacketsIn += p.VideoPacketsIn
		stats.TotalAudioPacketsIn += p.AudioPacketsIn
		stats.TotalVideoPacketsLost += p.VideoPacketsLost
		stats.TotalAudioPacketsLost += p.AudioPacketsLost
	}

	stats.UpdatedAt = time.Now().UTC()

	return stats
}

// Database persistence helpers

// persistParticipant saves a new participant to the database
func (pm *ParticipantManager) persistParticipant(ctx context.Context, userID uuid.UUID) error {
	if pm.db == nil {
		return nil // Database not available, skip persistence
	}

	query := `
		INSERT INTO recording_participants (
			id, recording_id, user_id, joined_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
		ON CONFLICT (recording_id, user_id) DO NOTHING
	`

	now := time.Now().UTC()
	participantID := uuid.New()

	_, err := pm.db.ExecContext(ctx, query,
		participantID,
		pm.recordingID,
		userID,
		now,
		now,
		now,
	)

	return err
}

// persistParticipantRemoval updates participant left_at timestamp
func (pm *ParticipantManager) persistParticipantRemoval(ctx context.Context, userID uuid.UUID, leftAt time.Time) error {
	if pm.db == nil {
		return nil
	}

	query := `
		UPDATE recording_participants
		SET left_at = $1, updated_at = $2
		WHERE recording_id = $3 AND user_id = $4
	`

	now := time.Now().UTC()
	_, err := pm.db.ExecContext(ctx, query, leftAt, now, pm.recordingID, userID)

	return err
}

// persistParticipantStats updates participant statistics
func (pm *ParticipantManager) persistParticipantStats(ctx context.Context, userID uuid.UUID, stats *ParticipantStats) error {
	if pm.db == nil {
		return nil
	}

	// Only persist if database is available - this is periodic so failures are acceptable
	query := `
		UPDATE recording_participants
		SET 
			video_frames = $1,
			audio_frames = $2,
			video_bytes_in = $3,
			audio_bytes_in = $4,
			video_packets_in = $5,
			audio_packets_in = $6,
			video_packets_lost = $7,
			audio_packets_lost = $8,
			updated_at = $9
		WHERE recording_id = $10 AND user_id = $11
	`

	now := time.Now().UTC()
	_, err := pm.db.ExecContext(ctx, query,
		stats.VideoFrames,
		stats.AudioFrames,
		stats.VideoBytesIn,
		stats.AudioBytesIn,
		stats.VideoPacketsIn,
		stats.AudioPacketsIn,
		stats.VideoPacketsLost,
		stats.AudioPacketsLost,
		now,
		pm.recordingID,
		userID,
	)

	return err
}

// ParticipantStatsUpdate is used to update participant statistics
type ParticipantStatsUpdate struct {
	VideoFrames      int64
	AudioFrames      int64
	VideoBytesIn     int64
	AudioBytesIn     int64
	VideoPacketsIn   int64
	AudioPacketsIn   int64
	VideoPacketsLost int64
	AudioPacketsLost int64
}

// RecordingAggregateStats contains aggregate statistics for a recording
type RecordingAggregateStats struct {
	TotalParticipants     int
	ActiveParticipants    int
	TotalVideoFrames      int64
	TotalAudioFrames      int64
	TotalVideoBytesIn     int64
	TotalAudioBytesIn     int64
	TotalVideoPacketsIn   int64
	TotalAudioPacketsIn   int64
	TotalVideoPacketsLost int64
	TotalAudioPacketsLost int64
	UpdatedAt             time.Time
}
