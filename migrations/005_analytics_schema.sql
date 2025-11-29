-- Phase 4: Analytics Event Collection & Storage Schema
-- Tables for event collection, storage, and aggregation

-- Table: analytics_events
-- Stores raw analytics events from streaming system
CREATE TABLE IF NOT EXISTS analytics_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(50) NOT NULL,
    recording_id UUID NOT NULL,
    user_id UUID NOT NULL,
    session_id VARCHAR(255) NOT NULL,
    event_timestamp TIMESTAMP NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_analytics_events_recording_id ON analytics_events(recording_id);
CREATE INDEX IF NOT EXISTS idx_analytics_events_user_id ON analytics_events(user_id);
CREATE INDEX IF NOT EXISTS idx_analytics_events_event_type ON analytics_events(event_type);
CREATE INDEX IF NOT EXISTS idx_analytics_events_timestamp ON analytics_events(event_timestamp DESC);

-- Table: playback_sessions
-- Stores user playback session data (one record per user per recording)
CREATE TABLE IF NOT EXISTS playback_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recording_id UUID NOT NULL,
    user_id UUID NOT NULL,
    session_id VARCHAR(255) NOT NULL,
    duration_seconds INTEGER NOT NULL,
    watched_seconds INTEGER NOT NULL,
    completion_rate DECIMAL(5, 2) NOT NULL,
    quality VARCHAR(50),
    buffer_events INTEGER DEFAULT 0,
    started_at TIMESTAMP NOT NULL,
    ended_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_playback_sessions_recording_id ON playback_sessions(recording_id);
CREATE INDEX IF NOT EXISTS idx_playback_sessions_user_id ON playback_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_playback_sessions_session_id ON playback_sessions(session_id);

-- Table: quality_events
-- Tracks quality/bitrate changes during playback
CREATE TABLE IF NOT EXISTS quality_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id VARCHAR(255) NOT NULL,
    recording_id UUID NOT NULL,
    user_id UUID NOT NULL,
    old_quality VARCHAR(50),
    new_quality VARCHAR(50) NOT NULL,
    old_bitrate INTEGER,
    new_bitrate INTEGER,
    reason VARCHAR(100),
    quality_timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_quality_events_session_id ON quality_events(session_id);
CREATE INDEX IF NOT EXISTS idx_quality_events_recording_id ON quality_events(recording_id);

-- Table: engagement_metrics
-- Aggregate engagement metrics per user per lecture
CREATE TABLE IF NOT EXISTS engagement_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recording_id UUID NOT NULL,
    user_id UUID NOT NULL,
    total_views INTEGER NOT NULL DEFAULT 0,
    total_watch_time_seconds INTEGER NOT NULL DEFAULT 0,
    average_quality VARCHAR(50),
    buffer_events_total INTEGER NOT NULL DEFAULT 0,
    engagement_score DECIMAL(5, 2) NOT NULL DEFAULT 0.0,
    last_watched TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_engagement_metrics_user_recording ON engagement_metrics(user_id, recording_id);

-- Table: lecture_statistics
-- Aggregate statistics per lecture
CREATE TABLE IF NOT EXISTS lecture_statistics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recording_id UUID NOT NULL,
    course_id UUID NOT NULL,
    total_viewers INTEGER NOT NULL DEFAULT 0,
    unique_viewers INTEGER NOT NULL DEFAULT 0,
    average_watch_time DECIMAL(10, 2) NOT NULL DEFAULT 0.0,
    average_completion_rate DECIMAL(5, 2) NOT NULL DEFAULT 0.0,
    quality_distribution JSONB,
    buffer_events_total INTEGER NOT NULL DEFAULT 0,
    avg_engagement_score DECIMAL(5, 2) NOT NULL DEFAULT 0.0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_lecture_statistics_recording_id ON lecture_statistics(recording_id);
CREATE INDEX IF NOT EXISTS idx_lecture_statistics_course_id ON lecture_statistics(course_id);

-- Table: course_statistics
-- Aggregate statistics per course
CREATE TABLE IF NOT EXISTS course_statistics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID NOT NULL,
    total_lectures INTEGER NOT NULL DEFAULT 0,
    total_unique_students INTEGER NOT NULL DEFAULT 0,
    average_attendance_rate DECIMAL(5, 2) NOT NULL DEFAULT 0.0,
    average_completion_rate DECIMAL(5, 2) NOT NULL DEFAULT 0.0,
    engagement_trend VARCHAR(20),
    most_watched_lectures JSONB,
    struggle_lectures JSONB,
    avg_engagement_score DECIMAL(5, 2) NOT NULL DEFAULT 0.0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_course_statistics_course_id ON course_statistics(course_id);

-- Permissions for analytics tables
GRANT SELECT, INSERT, UPDATE ON analytics_events TO analytics_user;
GRANT SELECT, INSERT, UPDATE ON playback_sessions TO analytics_user;
GRANT SELECT, INSERT, UPDATE ON quality_events TO analytics_user;
GRANT SELECT, INSERT, UPDATE ON engagement_metrics TO analytics_user;
GRANT SELECT, INSERT, UPDATE ON lecture_statistics TO analytics_user;
GRANT SELECT, INSERT, UPDATE ON course_statistics TO analytics_user;
