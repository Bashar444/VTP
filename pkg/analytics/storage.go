package analytics

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// PostgresAnalyticsStore implements the analytics storage in PostgreSQL
type PostgresAnalyticsStore struct {
	db     *sql.DB
	logger *log.Logger
}

// NewPostgresAnalyticsStore creates a new PostgreSQL analytics store
func NewPostgresAnalyticsStore(db *sql.DB, logger *log.Logger) *PostgresAnalyticsStore {
	return &PostgresAnalyticsStore{
		db:     db,
		logger: logger,
	}
}

// StoreEvent stores a single event
func (ps *PostgresAnalyticsStore) StoreEvent(event AnalyticsEvent) error {
	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO analytics_events (event_id, event_type, recording_id, user_id, session_id, timestamp, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = ps.db.Exec(
		query,
		event.EventID,
		string(event.EventType),
		event.RecordingID,
		event.UserID,
		event.SessionID,
		event.Timestamp,
		metadataJSON,
		event.CreatedAt,
	)

	if err != nil {
		ps.logger.Printf("[Analytics] Error storing event: %v\n", err)
		return fmt.Errorf("failed to store event: %w", err)
	}

	return nil
}

// StoreEvents stores multiple events in a batch
func (ps *PostgresAnalyticsStore) StoreEvents(events []AnalyticsEvent) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := ps.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO analytics_events (event_id, event_type, recording_id, user_id, session_id, timestamp, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	for _, event := range events {
		metadataJSON, err := json.Marshal(event.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}

		_, err = tx.Exec(
			query,
			event.EventID,
			string(event.EventType),
			event.RecordingID,
			event.UserID,
			event.SessionID,
			event.Timestamp,
			metadataJSON,
			event.CreatedAt,
		)

		if err != nil {
			return fmt.Errorf("failed to store event: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	ps.logger.Printf("[Analytics] Stored %d events\n", len(events))
	return nil
}

// StorePlaybackSession stores a playback session
func (ps *PostgresAnalyticsStore) StorePlaybackSession(session PlaybackSession) error {
	query := `
		INSERT INTO playback_sessions 
		(id, recording_id, user_id, session_start, session_end, total_duration_seconds, 
		 watched_duration_seconds, pause_count, resume_count, quality_selected, buffer_events, 
		 completion_rate, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := ps.db.Exec(
		query,
		session.ID,
		session.RecordingID,
		session.UserID,
		session.SessionStart,
		session.SessionEnd,
		session.TotalDurationSeconds,
		session.WatchedDurationSeconds,
		session.PauseCount,
		session.ResumeCount,
		session.QualitySelected,
		session.BufferEvents,
		session.CompletionRate,
		session.CreatedAt,
	)

	if err != nil {
		ps.logger.Printf("[Analytics] Error storing playback session: %v\n", err)
		return fmt.Errorf("failed to store playback session: %w", err)
	}

	return nil
}

// UpdatePlaybackSession updates an existing playback session
func (ps *PostgresAnalyticsStore) UpdatePlaybackSession(session PlaybackSession) error {
	query := `
		UPDATE playback_sessions
		SET session_end = $1, watched_duration_seconds = $2, pause_count = $3, 
		    resume_count = $4, buffer_events = $5, completion_rate = $6
		WHERE id = $7
	`

	_, err := ps.db.Exec(
		query,
		session.SessionEnd,
		session.WatchedDurationSeconds,
		session.PauseCount,
		session.ResumeCount,
		session.BufferEvents,
		session.CompletionRate,
		session.ID,
	)

	if err != nil {
		ps.logger.Printf("[Analytics] Error updating playback session: %v\n", err)
		return fmt.Errorf("failed to update playback session: %w", err)
	}

	return nil
}

// StoreQualityEvent stores a quality change event
func (ps *PostgresAnalyticsStore) StoreQualityEvent(event QualityEvent) error {
	query := `
		INSERT INTO quality_events (id, session_id, timestamp, bitrate, resolution, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := ps.db.Exec(
		query,
		event.ID,
		event.SessionID,
		event.Timestamp,
		event.Bitrate,
		event.Resolution,
		event.Reason,
		event.CreatedAt,
	)

	if err != nil {
		ps.logger.Printf("[Analytics] Error storing quality event: %v\n", err)
		return fmt.Errorf("failed to store quality event: %w", err)
	}

	return nil
}

// StoreEngagementMetrics stores engagement metrics
func (ps *PostgresAnalyticsStore) StoreEngagementMetrics(metrics EngagementMetrics) error {
	query := `
		INSERT INTO engagement_metrics 
		(id, recording_id, user_id, total_watch_time_seconds, completion_percentage, 
		 rewatch_count, avg_quality, last_watched, engagement_score, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := ps.db.Exec(
		query,
		metrics.ID,
		metrics.RecordingID,
		metrics.UserID,
		metrics.TotalWatchTimeSeconds,
		metrics.CompletionPercentage,
		metrics.RewatchCount,
		metrics.AvgQuality,
		metrics.LastWatched,
		metrics.EngagementScore,
		metrics.CreatedAt,
		metrics.UpdatedAt,
	)

	if err != nil {
		ps.logger.Printf("[Analytics] Error storing engagement metrics: %v\n", err)
		return fmt.Errorf("failed to store engagement metrics: %w", err)
	}

	return nil
}

// UpdateEngagementMetrics updates engagement metrics
func (ps *PostgresAnalyticsStore) UpdateEngagementMetrics(metrics EngagementMetrics) error {
	query := `
		UPDATE engagement_metrics
		SET total_watch_time_seconds = $1, completion_percentage = $2, rewatch_count = $3,
		    avg_quality = $4, last_watched = $5, engagement_score = $6, updated_at = $7
		WHERE id = $8
	`

	_, err := ps.db.Exec(
		query,
		metrics.TotalWatchTimeSeconds,
		metrics.CompletionPercentage,
		metrics.RewatchCount,
		metrics.AvgQuality,
		metrics.LastWatched,
		metrics.EngagementScore,
		time.Now(),
		metrics.ID,
	)

	if err != nil {
		ps.logger.Printf("[Analytics] Error updating engagement metrics: %v\n", err)
		return fmt.Errorf("failed to update engagement metrics: %w", err)
	}

	return nil
}

// StoreLectureStatistics stores lecture statistics
func (ps *PostgresAnalyticsStore) StoreLectureStatistics(stats LectureStatistics) error {
	qualityDistJSON, err := json.Marshal(stats.QualityDistribution)
	if err != nil {
		return fmt.Errorf("failed to marshal quality distribution: %w", err)
	}

	query := `
		INSERT INTO lecture_statistics 
		(id, recording_id, unique_viewers, total_views, avg_watch_time_seconds, 
		 completion_rate, peak_concurrent_viewers, total_buffer_events, quality_distribution, 
		 most_replayed_section, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err = ps.db.Exec(
		query,
		stats.ID,
		stats.RecordingID,
		stats.UniqueViewers,
		stats.TotalViews,
		stats.AvgWatchTimeSeconds,
		stats.CompletionRate,
		stats.PeakConcurrentViewers,
		stats.TotalBufferEvents,
		qualityDistJSON,
		stats.MostReplayedSection,
		stats.CreatedAt,
		stats.UpdatedAt,
	)

	if err != nil {
		ps.logger.Printf("[Analytics] Error storing lecture statistics: %v\n", err)
		return fmt.Errorf("failed to store lecture statistics: %w", err)
	}

	return nil
}

// UpdateLectureStatistics updates lecture statistics
func (ps *PostgresAnalyticsStore) UpdateLectureStatistics(stats LectureStatistics) error {
	qualityDistJSON, err := json.Marshal(stats.QualityDistribution)
	if err != nil {
		return fmt.Errorf("failed to marshal quality distribution: %w", err)
	}

	query := `
		UPDATE lecture_statistics
		SET unique_viewers = $1, total_views = $2, avg_watch_time_seconds = $3,
		    completion_rate = $4, peak_concurrent_viewers = $5, total_buffer_events = $6,
		    quality_distribution = $7, most_replayed_section = $8, updated_at = $9
		WHERE id = $10
	`

	_, err = ps.db.Exec(
		query,
		stats.UniqueViewers,
		stats.TotalViews,
		stats.AvgWatchTimeSeconds,
		stats.CompletionRate,
		stats.PeakConcurrentViewers,
		stats.TotalBufferEvents,
		qualityDistJSON,
		stats.MostReplayedSection,
		time.Now(),
		stats.ID,
	)

	if err != nil {
		ps.logger.Printf("[Analytics] Error updating lecture statistics: %v\n", err)
		return fmt.Errorf("failed to update lecture statistics: %w", err)
	}

	return nil
}

// StoreCourseStatistics stores course statistics
func (ps *PostgresAnalyticsStore) StoreCourseStatistics(stats CourseStatistics) error {
	query := `
		INSERT INTO course_statistics 
		(id, course_id, total_students, attending_students, total_lectures, total_view_sessions,
		 avg_attendance_rate, total_watch_time_seconds, course_engagement_score, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := ps.db.Exec(
		query,
		stats.ID,
		stats.CourseID,
		stats.TotalStudents,
		stats.AttendingStudents,
		stats.TotalLectures,
		stats.TotalViewSessions,
		stats.AvgAttendanceRate,
		stats.TotalWatchTimeSeconds,
		stats.CourseEngagementScore,
		stats.CreatedAt,
		stats.UpdatedAt,
	)

	if err != nil {
		ps.logger.Printf("[Analytics] Error storing course statistics: %v\n", err)
		return fmt.Errorf("failed to store course statistics: %w", err)
	}

	return nil
}

// UpdateCourseStatistics updates course statistics
func (ps *PostgresAnalyticsStore) UpdateCourseStatistics(stats CourseStatistics) error {
	query := `
		UPDATE course_statistics
		SET total_students = $1, attending_students = $2, total_lectures = $3,
		    total_view_sessions = $4, avg_attendance_rate = $5, total_watch_time_seconds = $6,
		    course_engagement_score = $7, updated_at = $8
		WHERE id = $9
	`

	_, err := ps.db.Exec(
		query,
		stats.TotalStudents,
		stats.AttendingStudents,
		stats.TotalLectures,
		stats.TotalViewSessions,
		stats.AvgAttendanceRate,
		stats.TotalWatchTimeSeconds,
		stats.CourseEngagementScore,
		time.Now(),
		stats.ID,
	)

	if err != nil {
		ps.logger.Printf("[Analytics] Error updating course statistics: %v\n", err)
		return fmt.Errorf("failed to update course statistics: %w", err)
	}

	return nil
}
