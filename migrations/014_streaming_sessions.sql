-- Migration: Live Streaming Sessions
-- Description: Tables for managing live streaming sessions and chat

-- Live Streaming Sessions table
CREATE TABLE IF NOT EXISTS streaming_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id VARCHAR(100) UNIQUE NOT NULL,
    instructor_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_id UUID REFERENCES courses(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'live', 'ended', 'cancelled')),
    scheduled_at TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    max_participants INTEGER DEFAULT 100,
    is_recording BOOLEAN DEFAULT FALSE,
    recording_id UUID REFERENCES recordings(id) ON DELETE SET NULL,
    settings JSONB DEFAULT '{}'::jsonb, -- {allowChat: true, allowQuestions: true, etc.}
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_streaming_sessions_instructor ON streaming_sessions(instructor_id);
CREATE INDEX idx_streaming_sessions_course ON streaming_sessions(course_id);
CREATE INDEX idx_streaming_sessions_status ON streaming_sessions(status);
CREATE INDEX idx_streaming_sessions_room ON streaming_sessions(room_id);

-- Streaming Session Participants tracking
CREATE TABLE IF NOT EXISTS streaming_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES streaming_sessions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'viewer' CHECK (role IN ('host', 'co-host', 'viewer')),
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP WITH TIME ZONE,
    is_muted BOOLEAN DEFAULT FALSE,
    is_video_enabled BOOLEAN DEFAULT FALSE,
    is_hand_raised BOOLEAN DEFAULT FALSE,
    watch_duration INTEGER DEFAULT 0, -- seconds
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(session_id, user_id)
);

CREATE INDEX idx_streaming_participants_session ON streaming_participants(session_id);
CREATE INDEX idx_streaming_participants_user ON streaming_participants(user_id);

-- Stream Chat Messages
CREATE TABLE IF NOT EXISTS stream_chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES streaming_sessions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message_type VARCHAR(20) NOT NULL DEFAULT 'text' CHECK (message_type IN ('text', 'question', 'announcement', 'system')),
    content TEXT NOT NULL,
    is_pinned BOOLEAN DEFAULT FALSE,
    is_answered BOOLEAN DEFAULT FALSE, -- for questions
    reply_to_id UUID REFERENCES stream_chat_messages(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_stream_chat_session ON stream_chat_messages(session_id);
CREATE INDEX idx_stream_chat_user ON stream_chat_messages(user_id);
CREATE INDEX idx_stream_chat_type ON stream_chat_messages(message_type);
CREATE INDEX idx_stream_chat_created ON stream_chat_messages(created_at);

-- Stream Recordings (link sessions to recordings)
CREATE TABLE IF NOT EXISTS stream_recordings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES streaming_sessions(id) ON DELETE CASCADE,
    file_url TEXT NOT NULL,
    file_size BIGINT DEFAULT 0,
    duration INTEGER DEFAULT 0, -- seconds
    format VARCHAR(20) DEFAULT 'webm',
    thumbnail_url TEXT,
    is_processed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_stream_recordings_session ON stream_recordings(session_id);

-- View for active streams with participant counts
CREATE OR REPLACE VIEW active_streams AS
SELECT 
    s.id,
    s.room_id,
    s.title,
    s.instructor_id,
    u.full_name as instructor_name,
    s.course_id,
    c.title_ar as course_title,
    s.status,
    s.started_at,
    s.is_recording,
    s.settings,
    COUNT(DISTINCT p.user_id) as participant_count
FROM streaming_sessions s
JOIN users u ON s.instructor_id = u.id
LEFT JOIN courses c ON s.course_id = c.id
LEFT JOIN streaming_participants p ON s.id = p.session_id AND p.left_at IS NULL
WHERE s.status = 'live'
GROUP BY s.id, s.room_id, s.title, s.instructor_id, u.full_name, s.course_id, c.title_ar, s.status, s.started_at, s.is_recording, s.settings;
