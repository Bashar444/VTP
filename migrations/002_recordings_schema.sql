-- Phase 2A: Recording System Database Schema
-- Migration: 002_recordings_schema.sql
-- Description: Create tables for recording system with full audit trail

-- Enable UUID extension if not already present
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Main recordings table
CREATE TABLE IF NOT EXISTS recordings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    started_by UUID NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    stopped_at TIMESTAMP WITH TIME ZONE,
    duration_seconds INTEGER,
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'recording', 'processing', 'completed', 'failed', 'archived', 'deleted')),
    format VARCHAR(50) NOT NULL DEFAULT 'webm',
    file_path VARCHAR(1024),
    file_size_bytes BIGINT,
    mime_type VARCHAR(100),
    bitrate_kbps INTEGER,
    frame_rate_fps INTEGER,
    resolution VARCHAR(20),
    codecs VARCHAR(255),
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT fk_room FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE,
    CONSTRAINT fk_user FOREIGN KEY (started_by) REFERENCES users(id) ON DELETE SET NULL
);

-- Recording participants table (tracks who was in the recording)
CREATE TABLE IF NOT EXISTS recording_participants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    recording_id UUID NOT NULL,
    user_id UUID NOT NULL,
    peer_id VARCHAR(255) NOT NULL,
    joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP WITH TIME ZONE,
    producer_count INTEGER DEFAULT 0,
    consumer_count INTEGER DEFAULT 0,
    bytes_sent BIGINT DEFAULT 0,
    bytes_received BIGINT DEFAULT 0,
    packets_sent BIGINT DEFAULT 0,
    packets_lost BIGINT DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_recording FOREIGN KEY (recording_id) REFERENCES recordings(id) ON DELETE CASCADE,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Recording sharing table (who can access the recording)
CREATE TABLE IF NOT EXISTS recording_sharing (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    recording_id UUID NOT NULL,
    shared_by UUID NOT NULL,
    shared_with UUID,
    share_type VARCHAR(50) NOT NULL CHECK (share_type IN ('user', 'role', 'public', 'link')),
    access_level VARCHAR(50) NOT NULL CHECK (access_level IN ('view', 'download', 'share', 'delete')),
    expiry_at TIMESTAMP WITH TIME ZONE,
    share_link_token VARCHAR(255),
    is_revoked BOOLEAN DEFAULT FALSE,
    revoked_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_recording FOREIGN KEY (recording_id) REFERENCES recordings(id) ON DELETE CASCADE,
    CONSTRAINT fk_shared_by FOREIGN KEY (shared_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT fk_shared_with FOREIGN KEY (shared_with) REFERENCES users(id) ON DELETE CASCADE
);

-- Recording access audit log
CREATE TABLE IF NOT EXISTS recording_access_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    recording_id UUID NOT NULL,
    user_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL CHECK (action IN ('view', 'download', 'stream', 'delete', 'share', 'unshare')),
    ip_address INET,
    user_agent VARCHAR(1024),
    bytes_transferred BIGINT,
    duration_seconds INTEGER,
    status VARCHAR(50) NOT NULL CHECK (status IN ('success', 'failed', 'partial')),
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_recording FOREIGN KEY (recording_id) REFERENCES recordings(id) ON DELETE CASCADE,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_recordings_room_id ON recordings(room_id);
CREATE INDEX IF NOT EXISTS idx_recordings_started_by ON recordings(started_by);
CREATE INDEX IF NOT EXISTS idx_recordings_status ON recordings(status);
CREATE INDEX IF NOT EXISTS idx_recordings_created_at ON recordings(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_recordings_started_at ON recordings(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_recordings_deleted_at ON recordings(deleted_at);

CREATE INDEX IF NOT EXISTS idx_recording_participants_recording_id ON recording_participants(recording_id);
CREATE INDEX IF NOT EXISTS idx_recording_participants_user_id ON recording_participants(user_id);
CREATE INDEX IF NOT EXISTS idx_recording_participants_peer_id ON recording_participants(peer_id);
CREATE INDEX IF NOT EXISTS idx_recording_participants_joined_at ON recording_participants(joined_at DESC);

CREATE INDEX IF NOT EXISTS idx_recording_sharing_recording_id ON recording_sharing(recording_id);
CREATE INDEX IF NOT EXISTS idx_recording_sharing_shared_with ON recording_sharing(shared_with);
CREATE INDEX IF NOT EXISTS idx_recording_sharing_share_link_token ON recording_sharing(share_link_token);
CREATE INDEX IF NOT EXISTS idx_recording_sharing_expiry_at ON recording_sharing(expiry_at);

CREATE INDEX IF NOT EXISTS idx_recording_access_log_recording_id ON recording_access_log(recording_id);
CREATE INDEX IF NOT EXISTS idx_recording_access_log_user_id ON recording_access_log(user_id);
CREATE INDEX IF NOT EXISTS idx_recording_access_log_created_at ON recording_access_log(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_recording_access_log_action ON recording_access_log(action);

-- Add comments for documentation
COMMENT ON TABLE recordings IS 'Stores recording metadata for all room recordings';
COMMENT ON TABLE recording_participants IS 'Tracks participants who were in each recording session';
COMMENT ON TABLE recording_sharing IS 'Manages access permissions and sharing of recordings';
COMMENT ON TABLE recording_access_log IS 'Audit log for all recording access and operations';

COMMENT ON COLUMN recordings.status IS 'Recording lifecycle status: pending → recording → processing → completed/failed';
COMMENT ON COLUMN recordings.metadata IS 'JSON field for extensible recording properties';
COMMENT ON COLUMN recording_sharing.share_type IS 'Type of sharing: user (specific user), role (all members with role), public (everyone), link (anonymous link)';
COMMENT ON COLUMN recording_sharing.access_level IS 'Granular permission: view (stream), download (get file), share (share with others), delete (remove recording)';

-- Create materialized view for recording statistics (optional, for future use)
CREATE TABLE IF NOT EXISTS recording_stats_cache (
    recording_id UUID PRIMARY KEY,
    total_participants INTEGER,
    total_duration_seconds INTEGER,
    average_bitrate_kbps INTEGER,
    storage_bytes BIGINT,
    access_count INTEGER,
    download_count INTEGER,
    share_count INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_recording FOREIGN KEY (recording_id) REFERENCES recordings(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_recording_stats_cache_updated_at ON recording_stats_cache(updated_at DESC);

-- Migration metadata
CREATE TABLE IF NOT EXISTS schema_migrations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    version INT NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Record this migration
INSERT INTO schema_migrations (version, name) VALUES (2, '002_recordings_schema') 
ON CONFLICT (version) DO NOTHING;
