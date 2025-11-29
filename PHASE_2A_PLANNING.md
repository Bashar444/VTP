# Phase 2A Planning - Recording System Implementation

**Date:** November 21, 2025  
**Status:** Planning Phase  
**Estimated Duration:** 3-5 days  
**Complexity:** High

---

## Executive Summary

Phase 2A will implement a complete recording system for the VTP platform, enabling instructors to record live video sessions and store them for playback. This phase builds on the complete Phase 1C Mediasoup SFU integration.

**Key Objectives:**
1. Record audio/video from all peers in a room
2. Store recordings securely (local or S3)
3. Manage recording lifecycle (start/stop/delete)
4. Handle multiple concurrent recordings
5. Efficient file storage and cleanup

---

## Phase 1 Context (Completed)

### Phase 1a: Authentication ✅
- User registration and login
- JWT token management
- Password encryption (bcrypt)
- Role-based access control (RBAC)
- User profiles and settings

### Phase 1b: Signalling ✅
- Socket.IO WebRTC signaling server
- Room management
- Peer tracking and messaging
- Basic peer-to-peer WebRTC setup

### Phase 1c: Mediasoup SFU Integration ✅
- Selective Forwarding Unit for media routing
- Transport management (DTLS)
- Producer/consumer lifecycle
- Codec negotiation
- RTC port management

---

## Phase 2A: Recording System Architecture

### High-Level Design

```
┌─────────────────────────────────────────────────────────────┐
│                    Web Client                               │
│               (Start/Stop Recording UI)                     │
└──────────────┬──────────────────────────────────────────────┘
               │
       ┌───────v────────────────────┐
       │                            │
┌──────v─────────────────────┐  ┌──v──────────────────────────┐
│   Go Backend               │  │  Node.js Mediasoup SFU      │
│   (Port 8080)              │  │  (Port 3000)                │
├────────────────────────────┤  ├─────────────────────────────┤
│ Recording Manager          │  │ Recording Participant       │
│  - Start/stop recording    │  │  - Hook into producer       │
│  - File management         │  │  - Write to pipe/file       │
│  - Metadata tracking       │  │  - Real-time encoding       │
│  - Database storage        │  │                             │
│                            │  │ FFmpeg Integration          │
│ Recording Storage          │  │  - WebM/MP4 encoding        │
│  - Local filesystem        │  │  - Audio/video muxing       │
│  - S3 (optional)           │  │  - Codec selection          │
│  - Cleanup policies        │  │                             │
└────────────────────────────┘  └─────────────────────────────┘
       │                                    │
       └────────────────────┬───────────────┘
                            │
                ┌───────────v────────────┐
                │ Storage Layer          │
                ├───────────────────────┤
                │ Local: /recordings/   │
                │   - Raw files         │
                │   - Encoded files     │
                │   - Thumbnails        │
                │                       │
                │ Database:             │
                │   - Recording metadata│
                │   - Permissions       │
                │   - Access logs       │
                └───────────────────────┘
```

### Recording Flow

```
1. Instructor starts recording
   ├─ API: POST /api/v1/recordings/start
   ├─ Params: roomId, title, description
   └─ Response: recordingId, status

2. Mediasoup hooks recording participant
   ├─ Special role: RECORDER
   ├─ Subscribes to all producers
   ├─ Receives A/V streams

3. Recording writes to file
   ├─ Format selection (WebM/MP4)
   ├─ Codec selection (VP8/H264 for video, Opus for audio)
   ├─ Real-time muxing
   └─ File growth monitoring

4. File storage
   ├─ Save to /recordings/{roomId}/{recordingId}/
   ├─ Update database metadata
   ├─ Generate thumbnails
   └─ Index for search

5. Recording stops
   ├─ Flush buffers
   ├─ Finalize file
   ├─ Update status
   ├─ Calculate duration
   └─ Trigger post-processing

6. Post-processing (optional)
   ├─ Generate HLS streams
   ├─ Create low-res version
   ├─ Generate thumbnails
   └─ S3 upload if configured
```

---

## Technical Implementation Plan

### 2A.1: Database Schema Extensions

**New Tables:**
```sql
-- Recordings table
CREATE TABLE recordings (
    id UUID PRIMARY KEY,
    room_id VARCHAR(255) NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    started_by UUID NOT NULL REFERENCES users(id),
    status VARCHAR(50), -- recording, stopped, processing, done, failed
    file_path VARCHAR(1000),
    file_size BIGINT,
    duration_seconds INT,
    format VARCHAR(50), -- webm, mp4, mkv
    video_codec VARCHAR(50), -- vp8, h264, vp9
    audio_codec VARCHAR(50), -- opus, aac
    resolution_width INT,
    resolution_height INT,
    fps INT,
    bitrate INT,
    created_at TIMESTAMP,
    completed_at TIMESTAMP,
    deleted_at TIMESTAMP,
    metadata JSONB,
    INDEX idx_room_id (room_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
);

-- Recording participants (who was in recording)
CREATE TABLE recording_participants (
    id UUID PRIMARY KEY,
    recording_id UUID NOT NULL REFERENCES recordings(id),
    peer_id VARCHAR(255),
    user_id UUID REFERENCES users(id),
    full_name VARCHAR(255),
    role VARCHAR(50),
    joined_at TIMESTAMP,
    left_at TIMESTAMP
);

-- Recording permissions/sharing
CREATE TABLE recording_sharing (
    id UUID PRIMARY KEY,
    recording_id UUID NOT NULL REFERENCES recordings(id),
    shared_with_user_id UUID REFERENCES users(id),
    shared_with_role VARCHAR(50),
    permission VARCHAR(50), -- view, download, share
    shared_at TIMESTAMP
);
```

### 2A.2: Go Backend Implementation

**New Packages:**
```
pkg/recording/
  ├── recording.go          -- Recording types and core logic
  ├── recording_handler.go  -- HTTP handlers
  ├── recorder.go           -- Recording process management
  ├── storage.go            -- File storage operations
  ├── s3.go                 -- S3 integration (optional)
  ├── ffmpeg.go             -- FFmpeg wrapper
  ├── cleanup.go            -- Cleanup and archival
  └── recording_test.go     -- Unit tests
```

**Key Components:**

```go
// Recording types
type Recording struct {
    ID            string    `db:"id"`
    RoomID        string    `db:"room_id"`
    Title         string    `db:"title"`
    Description   string    `db:"description"`
    StartedBy     string    `db:"started_by"`
    Status        string    `db:"status"` // recording, stopped, processing
    FilePath      string    `db:"file_path"`
    FileSize      int64     `db:"file_size"`
    DurationSecs  int       `db:"duration_seconds"`
    Format        string    `db:"format"` // webm, mp4
    CreatedAt     time.Time `db:"created_at"`
    CompletedAt   *time.Time `db:"completed_at"`
}

// Recording service
type RecordingService struct {
    db             *sql.DB
    mediaURL       string
    storageDir     string
    s3Client       *s3.Client // optional
}

// API endpoints
func (rs *RecordingService) StartRecording(ctx context.Context, req StartRecordingRequest) (*Recording, error)
func (rs *RecordingService) StopRecording(ctx context.Context, recordingID string) error
func (rs *RecordingService) GetRecording(ctx context.Context, recordingID string) (*Recording, error)
func (rs *RecordingService) ListRecordings(ctx context.Context, roomID string) ([]*Recording, error)
func (rs *RecordingService) DeleteRecording(ctx context.Context, recordingID string) error
```

### 2A.3: Mediasoup Service Enhancement

**Recording Participant:**
```javascript
// Special peer for recording
class RecordingParticipant {
  - Joins room with special role
  - Subscribes to ALL producers
  - No producers (receive-only)
  - Streams to Go backend via pipe
  - Transparent to other peers
}
```

**FFmpeg Integration:**
```bash
# Example pipeline
ffmpeg -f libvpx \
  -pixel_format yuv420p \
  -framerate 30 \
  -i pipe:0 \
  -f opus \
  -aac_adtstoasc \
  -i pipe:1 \
  -c:v libvpx \
  -c:a libopus \
  -b:v 2000k \
  -b:a 128k \
  -f webm \
  output.webm
```

### 2A.4: API Endpoints

```
POST   /api/v1/recordings/start
  - Request: { roomId, title, description }
  - Response: { recordingId, status, startedAt }
  - Auth: Required (instructor role)

POST   /api/v1/recordings/{recordingId}/stop
  - Response: { recordingId, status, stoppedAt, duration }
  - Auth: Required (started_by or admin)

GET    /api/v1/recordings/{recordingId}
  - Response: Recording details with metadata
  - Auth: Required (with permission check)

GET    /api/v1/recordings?roomId=X&limit=10&offset=0
  - Response: List of recordings
  - Auth: Required

DELETE /api/v1/recordings/{recordingId}
  - Response: { success, deletedAt }
  - Auth: Required (started_by or admin)

GET    /api/v1/recordings/{recordingId}/download
  - Returns: Raw file download
  - Auth: Required (with permission)

GET    /api/v1/recordings/{recordingId}/stream
  - Returns: HLS stream (chunked video)
  - Auth: Required (with permission)
```

### 2A.5: Socket.IO Events

```javascript
// Events to emit
'recording-started' -> { recordingId, title, startedBy, timestamp }
'recording-stopped' -> { recordingId, duration, fileSize, timestamp }
'recording-error' -> { recordingId, error, timestamp }
'recording-ready' -> { recordingId, status } // ready for download/playback

// Broadcast to room
io.to(roomId).emit('recording-started', {...})
io.to(roomId).emit('recording-stopped', {...})
```

---

## Implementation Phases

### Phase 2A.1: Core Recording Framework (Day 1)
**Duration:** 4 hours

- [ ] Database schema creation
- [ ] Recording model and types
- [ ] Recording service initialization
- [ ] File storage setup
- [ ] Basic unit tests

**Deliverables:**
- Database tables
- Go types and models
- Storage directory structure
- Initial tests (unit)

### Phase 2A.2: Recording Lifecycle (Day 2)
**Duration:** 4-5 hours

- [ ] Start recording implementation
- [ ] Stop recording implementation
- [ ] Recording participant lifecycle
- [ ] FFmpeg process management
- [ ] File validation

**Deliverables:**
- Start/stop APIs
- Recording process
- FFmpeg integration
- Error handling

### Phase 2A.3: Storage & Retrieval (Day 3)
**Duration:** 4 hours

- [ ] File storage operations
- [ ] Database metadata tracking
- [ ] List recordings API
- [ ] Get recording API
- [ ] Permission checks

**Deliverables:**
- Storage management
- Metadata tracking
- Query APIs
- Access control

### Phase 2A.4: Advanced Features (Day 4)
**Duration:** 4 hours

- [ ] Download endpoint
- [ ] HLS streaming
- [ ] Thumbnail generation
- [ ] S3 integration (optional)
- [ ] Cleanup policies

**Deliverables:**
- Download API
- HLS support
- Thumbnails
- Cloud storage (optional)

### Phase 2A.5: Testing & Documentation (Day 5)
**Duration:** 3-4 hours

- [ ] Integration tests
- [ ] Performance testing
- [ ] End-to-end testing
- [ ] Documentation
- [ ] Validation checklist

**Deliverables:**
- Test suite
- Performance metrics
- Documentation
- Validation report

---

## Technical Stack

### Required Packages (Go)
```go
// Media Processing
- github.com/go-echarts/go-echarts (thumbnails)
- ffmpeg-go (FFmpeg wrapper)

// File Storage
- github.com/aws/aws-sdk-go-v2 (S3 optional)

// Database
- github.com/lib/pq (PostgreSQL)

// Existing
- github.com/googollee/go-socket.io (Socket.IO)
- github.com/golang-jwt/jwt/v5 (Auth)
```

### Required Tools
```
- FFmpeg (for encoding)
- ImageMagick (for thumbnails)
- PostgreSQL (existing)
```

### Configuration
```env
# Recording Service
RECORDING_STORAGE_DIR=/recordings
RECORDING_FORMAT=webm # webm, mp4
RECORDING_VIDEO_CODEC=vp8 # vp8, h264
RECORDING_AUDIO_CODEC=opus # opus, aac
RECORDING_BITRATE=2000k
RECORDING_FPS=30

# S3 (optional)
AWS_ENABLED=false
AWS_REGION=us-east-1
AWS_BUCKET=vtp-recordings
AWS_ACCESS_KEY_ID=xxx
AWS_SECRET_ACCESS_KEY=xxx

# Cleanup
RECORDING_RETENTION_DAYS=90
RECORDING_AUTO_DELETE=true
```

---

## Data Model

### Recording Lifecycle States
```
1. pending -> 2. recording -> 3. stopped -> 4. processing -> 5. ready
                                               |
                                               v
                                        ERROR (cleanup)
```

### Storage Structure
```
/recordings/
  ├── {roomId1}/
  │   ├── {recordingId1}/
  │   │   ├── recording.webm (main file)
  │   │   ├── metadata.json
  │   │   ├── thumbnail.jpg
  │   │   └── hls/ (if enabled)
  │   │       ├── stream.m3u8
  │   │       ├── segment0.ts
  │   │       └── segment1.ts
  │   └── {recordingId2}/
  │       └── ...
  └── {roomId2}/
      └── ...
```

---

## Security Considerations

### Access Control
```
Owner (instructor): Full access
  - Download
  - Share
  - Delete
  - View analytics

Shared user: Limited access
  - Download (if permission)
  - View only (read-only)

Admin: Full access
  - All recordings
  - Delete any
  - Storage management
```

### File Security
```
- Store in secure directory
- Restrict file permissions (0600)
- Hash verification (SHA256)
- Encryption at rest (optional)
- Encryption in transit (HTTPS)
```

### Audit Trail
```
- Log all access
- Log downloads
- Log deletions
- Log shares
- Track storage usage
```

---

## Performance Targets

### Recording Quality
```
Video:
  - Codec: VP8 or H264
  - Bitrate: 2000-5000 kbps (adjustable)
  - FPS: 30
  - Resolution: 720p-1080p

Audio:
  - Codec: Opus
  - Bitrate: 128 kbps
  - Sample rate: 48kHz
  - Channels: Stereo
```

### File Sizes
```
1-hour recording:
  - 720p/2000kbps: ~900MB
  - 1080p/3000kbps: ~1.3GB

Compression ratio: ~10-15% with storage optimization
```

### Processing Performance
```
Real-time transcoding: Supported (FFmpeg)
HLS generation: Post-process (30min → 5min encode)
Thumbnail generation: < 1 second
Database ops: < 100ms
```

---

## Testing Strategy

### Unit Tests
```
- Recording lifecycle tests
- Storage operation tests
- Permission tests
- File validation tests
- Cleanup tests
```

### Integration Tests
```
- Full recording workflow
- Multiple concurrent recordings
- Room with recording participant
- Download/playback
- Permission checks
```

### Performance Tests
```
- Encoding quality
- File size verification
- Database query performance
- Concurrent recordings
- Storage capacity
```

### End-to-End Tests
```
- Record real session
- Download file
- Verify playback
- Share with user
- Delete recording
```

---

## Risk Assessment

### Technical Risks
| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| FFmpeg crashes | Medium | High | Process restart, monitoring |
| Disk space full | Low | High | Quota, auto-cleanup |
| Large file handling | Low | Medium | Chunked uploads |
| Concurrent encoding | Medium | High | Resource pooling |

### Security Risks
| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| Unauthorized access | Low | High | Permission checks |
| File corruption | Very Low | High | Checksums, backups |
| Data leakage | Very Low | Very High | Encryption, audit logs |

---

## Dependencies on Previous Phases

✅ **Phase 1a:** User authentication and roles  
✅ **Phase 1b:** Room management and Socket.IO  
✅ **Phase 1c:** Mediasoup SFU (recording participant joins)

**Critical Path:**
1. Phase 1c must be complete
2. Mediasoup service must be running
3. Database must be initialized

---

## Success Criteria

✅ **Phase 2A is complete when:**

1. Recording can be started from instructor UI
2. All audio/video from room is captured
3. Recording quality meets specifications
4. Files stored securely and efficiently
5. List/download/delete endpoints working
6. Permissions enforced correctly
7. All 50+ unit tests passing
8. Integration tests passing
9. Performance meets targets
10. Documentation complete
11. Security audit passed

---

## Estimated Effort Breakdown

| Task | Duration | Resource |
|------|----------|----------|
| Design & Architecture | 2 hours | Team |
| Database Schema | 1 hour | Developer |
| Core Recording | 8 hours | Developer |
| Storage & Retrieval | 6 hours | Developer |
| Advanced Features | 6 hours | Developer |
| Testing | 6 hours | Developer + QA |
| Documentation | 3 hours | Developer + Tech Writer |
| Code Review | 2 hours | Team |
| **Total** | **34 hours** | **~5 days** |

---

## Conclusion

Phase 2A will implement a production-ready recording system that integrates seamlessly with the existing authentication, signalling, and media infrastructure. The implementation will follow the established patterns from phases 1a-1c.

**Ready to start after Phase 1C deployment is verified.**

---

## Appendix: Quick Reference Commands

```bash
# Database setup
psql -U postgres -d vtp_db -f phase_2a_schema.sql

# FFmpeg test recording
ffmpeg -f lavfi -i testsrc=duration=10:size=1280x720 \
  -f lavfi -i sine=frequency=1000 \
  -c:v libvpx -c:a libopus -f webm test.webm

# Directory setup
mkdir -p /recordings
chmod 755 /recordings

# Start backend with recording support
go run cmd/main.go

# Test recording API
curl -X POST http://localhost:8080/api/v1/recordings/start \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"roomId":"test","title":"Test"}'
```

---

**Phase 2A Planning:** COMPLETE ✅  
**Ready to implement:** After Phase 1C deployment  
**Target start date:** November 22, 2025

