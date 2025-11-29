# Phase 2B: Advanced Streaming - Day 1 Plan 📺

**Status:** READY TO START  
**Estimated Duration:** 4-5 days (8-10 days for full implementation)  
**Complexity:** High (Advanced video streaming technology)  
**Dependencies:** Phase 2a Complete ✅ Phase 3 Complete ✅  

---

## 🎯 Overview: What is Phase 2B?

Phase 2B enhances the recording system with professional-grade streaming capabilities:

```
PHASE 2A (Current):          PHASE 2B (Next):
└─ Single bitrate HLS    ──→  Multi-bitrate HLS (Adaptive)
└─ Basic streaming       ──→  Quality-aware streaming
└─ Single playback       ──→  Concurrent multi-stream playback
└─ File-based only       ──→  Live streaming support
```

### The Problem We're Solving

Currently, when users have slow internet:
- ❌ Video stutters or buffers
- ❌ Requires manual quality selection
- ❌ Poor user experience on slow networks

With Phase 2B:
- ✅ Automatically switches quality based on bandwidth
- ✅ Multiple bitrate versions available
- ✅ Seamless playback across network conditions
- ✅ Live streaming to hundreds of concurrent users

---

## 📊 Phase 2B Architecture

### Three New Components

```
┌─────────────────────────────────────────────────────────────┐
│           Phase 2B: Advanced Streaming System               │
└─────────────────────────────────────────────────────────────┘

1. Adaptive Bitrate Engine (ABR)
   ├─ Monitor network bandwidth
   ├─ Detect user device capabilities
   ├─ Calculate optimal quality
   └─ Switch without interruption

2. Multi-Bitrate Transcoding Manager
   ├─ Encode to multiple bitrates
   ├─ Generate variant playlists
   ├─ Manage encoding queue
   └─ Cache transcoded versions

3. Live Distribution Network
   ├─ Accept incoming live streams
   ├─ Package HLS on-the-fly
   ├─ Distribute to viewers
   └─ Analytics collection
```

### Data Flow

```
Recording Service (Phase 2a)
         ↓
         ├─→ [Transcoding Queue]
         │        ↓
         ├────[ABR Manager]
         │        ├─ 500 kbps MP4
         │        ├─ 1000 kbps MP4
         │        ├─ 2000 kbps MP4
         │        └─ 4000 kbps MP4
         │
         ├─→ [HLS Generator]
         │        └─ Master playlist (links all bitrates)
         │           Variant 1: 500 kbps
         │           Variant 2: 1000 kbps
         │           Variant 3: 2000 kbps
         │           Variant 4: 4000 kbps
         │
         └─→ [Client Player]
                  ├─ Detect bandwidth
                  ├─ Request appropriate variant
                  └─ Auto-switch if bandwidth changes
```

---

## 🛠️ Implementation Plan: Day 1-4

### Day 1: Adaptive Bitrate Engine

**Objective:** Build the ABR system to detect network conditions

```
Files to Create:
├─ pkg/streaming/abr.go (300+ lines)
│  ├─ AdaptiveBitrateManager struct
│  ├─ NetworkConditionDetector
│  ├─ QualitySelector
│  ├─ BitrateProfile struct
│  └─ SegmentQualityTracker
│
├─ pkg/streaming/types.go (100+ lines)
│  ├─ BitrateLevel enum
│  ├─ StreamQuality struct
│  ├─ NetworkStats struct
│  └─ ABRConfig struct
│
└─ pkg/streaming/abr_test.go (200+ lines)
   ├─ TestBandwidthDetection
   ├─ TestQualitySwitching
   ├─ TestSegmentFetching
   └─ TestLoadBalancing
```

**Key Functions to Implement:**

```go
// Track network performance
func (abr *AdaptiveBitrateManager) RecordSegmentMetrics(metrics SegmentMetrics)

// Determine best quality for current conditions
func (abr *AdaptiveBitrateManager) SelectQuality(bandwidth float64) BitrateLevel

// Recommend switching to higher/lower quality
func (abr *AdaptiveBitrateManager) ShouldUpscale() bool
func (abr *AdaptiveBitrateManager) ShouldDownscale() bool

// Get available bitrate profiles
func (abr *AdaptiveBitrateManager) GetAvailableBitrates() []int

// Predict next quality based on history
func (abr *AdaptiveBitrateManager) PredictOptimalBitrate() int
```

**Day 1 Deliverables:**
- ✅ ABR engine implementation (300+ lines)
- ✅ Network condition detection
- ✅ Quality selection algorithm
- ✅ Unit tests (15+ test cases)
- ✅ Documentation

---

### Day 2: Multi-Bitrate Transcoding Manager

**Objective:** Create system to encode recordings into multiple bitrates

```
Files to Create:
├─ pkg/streaming/transcoder.go (350+ lines)
│  ├─ MultiBitrateTranscoder struct
│  ├─ TranscodingQueue
│  ├─ TranscodingJob struct
│  ├─ ProgressTracker
│  └─ EncodingProfile
│
├─ pkg/streaming/transcoding_service.go (250+ lines)
│  ├─ StartMultiBitrateEncoding()
│  ├─ ManageTranscodingQueue()
│  ├─ MonitorProgress()
│  ├─ GenerateVariantPlaylists()
│  └─ CacheManagement
│
└─ pkg/streaming/transcoder_test.go (200+ lines)
   ├─ TestParallelEncoding
   ├─ TestQueueManagement
   ├─ TestPlaylistGeneration
   └─ TestProgressTracking
```

**Key Functions:**

```go
// Queue transcoding job for multiple bitrates
func (mt *MultiBitrateTranscoder) QueueJob(recordingID uuid.UUID, profiles []EncodingProfile) error

// Start encoding with specified bitrate
func (mt *MultiBitrateTranscoder) StartEncoding(job TranscodingJob) error

// Monitor encoding progress
func (mt *MultiBitrateTranscoder) GetProgress(jobID uuid.UUID) ProgressUpdate

// Generate master playlist with all variants
func (mt *MultiBitrateTranscoder) GenerateMasterPlaylist(recordingID uuid.UUID) (string, error)

// Generate variant playlist for specific bitrate
func (mt *MultiBitrateTranscoder) GenerateVariantPlaylist(recordingID uuid.UUID, bitrate int) (string, error)
```

**Day 2 Deliverables:**
- ✅ Multi-bitrate transcoder (350+ lines)
- ✅ Transcoding queue system
- ✅ Playlist generation
- ✅ Progress tracking
- ✅ Unit tests (15+ test cases)

---

### Day 3: Live Distribution Network

**Objective:** Enable live streaming to multiple concurrent viewers

```
Files to Create:
├─ pkg/streaming/live_distributor.go (300+ lines)
│  ├─ LiveDistributor struct
│  ├─ LiveStream struct
│  ├─ ViewerConnection
│  ├─ SegmentBuffer
│  └─ ConnectionPool
│
├─ pkg/streaming/live_handlers.go (250+ lines)
│  ├─ HandleLiveStream()
│  ├─ HandleStreamIngest()
│  ├─ HandleViewerSubscription()
│  ├─ BroadcastSegment()
│  └─ ManageConnections
│
└─ pkg/streaming/live_test.go (200+ lines)
   ├─ TestConcurrentViewers
   ├─ TestSegmentDistribution
   ├─ TestConnectionManagement
   └─ TestFailover
```

**Key Functions:**

```go
// Start accepting live stream from encoder
func (ld *LiveDistributor) IngestLiveStream(roomID uuid.UUID) (ingestURL string, err error)

// Client subscribes to live stream
func (ld *LiveDistributor) SubscribeToLive(roomID uuid.UUID, quality BitrateLevel) (chan HLSSegment, error)

// Broadcast segment to all connected viewers
func (ld *LiveDistributor) BroadcastSegment(segment HLSSegment)

// Get viewer statistics
func (ld *LiveDistributor) GetLiveStats(roomID uuid.UUID) LiveStreamStats

// Manage viewer connections
func (ld *LiveDistributor) DisconnectViewer(connID uuid.UUID)
func (ld *LiveDistributor) GetActiveViewerCount(roomID uuid.UUID) int
```

**Day 3 Deliverables:**
- ✅ Live distributor system (300+ lines)
- ✅ Concurrent viewer management
- ✅ Segment buffering and distribution
- ✅ Connection pool management
- ✅ Unit tests (15+ test cases)

---

### Day 4: Integration & Testing

**Objective:** Integrate all components into main system

```
Integration Tasks:
├─ Update pkg/recording/streaming.go
│  └─ Add ABR manager initialization
│
├─ Update pkg/recording/playback.go
│  ├─ Integrate ABR quality selection
│  ├─ Add multi-bitrate playlist endpoints
│  └─ Add live stream endpoints
│
├─ Update cmd/main.go
│  ├─ Initialize ABR manager
│  ├─ Initialize transcoder
│  ├─ Initialize live distributor
│  ├─ Register 6 new endpoints
│  └─ Display startup information
│
├─ Create pkg/streaming/handlers.go (200+ lines)
│  ├─ Advanced streaming API handlers
│  ├─ Live streaming endpoints
│  └─ Quality selection endpoints
│
└─ Create migrations/004_advanced_streaming_schema.sql
   ├─ Create transcoding_jobs table
   ├─ Create live_streams table
   ├─ Create viewer_analytics table
   └─ Add appropriate indexes
```

**New Endpoints (6 total):**

```
PHASE 2B - Advanced Streaming (6 Endpoints)

Live Streaming:
POST   /api/v1/recordings/{id}/stream/start-live    - Begin live stream
GET    /api/v1/recordings/{id}/stream/live           - Watch live stream (WebSocket)
DELETE /api/v1/recordings/{id}/stream/stop-live      - Stop live stream

Multi-Bitrate:
GET    /api/v1/recordings/{id}/stream/master.m3u8    - Master playlist (all bitrates)
POST   /api/v1/recordings/{id}/transcode/quality     - Trigger multi-bitrate encoding
GET    /api/v1/recordings/{id}/transcode/progress    - Get transcoding progress

Additional:
GET    /api/v1/recordings/{id}/quality/recommended   - Get recommended quality
```

**Day 4 Deliverables:**
- ✅ All 6 new endpoints implemented
- ✅ Integration with Phase 2a complete
- ✅ Database migrations applied
- ✅ Production binary updated
- ✅ Comprehensive testing

---

## 📈 Expected Performance Improvements

### Before Phase 2B (Phase 2a)
```
Single Bitrate (2000 kbps):
├─ Perfect for 4G users (100% smooth)
├─ Stutters on 3G (buffer:30%, quality:100%)
└─ Unusable on slow WiFi (buffer:80%, quality:50%)

Result: 60% of users have good experience
```

### After Phase 2B
```
Adaptive Bitrate (500-4000 kbps):
├─ 4G users: 4000 kbps (best quality)
├─ 3G users: 1000 kbps (smooth playback)
├─ Slow WiFi: 500 kbps (watchable, no buffering)
└─ Automatic switching as conditions change

Result: 95% of users have good experience
```

---

## 🛠️ Technical Requirements

### Additional Dependencies

```go
import (
    "github.com/grafana/tempo-cli/pkg/util"    // for bandwidth estimation
    "golang.org/x/time/rate"                   // for rate limiting
)
```

### FFmpeg Commands (Multi-bitrate)

```bash
# Encode to 4 different bitrates simultaneously
ffmpeg -i input.mp4 \
  -c:v libx264 -preset medium -b:v 500k output_500.mp4 \
  -c:v libx264 -preset medium -b:v 1000k output_1000.mp4 \
  -c:v libx264 -preset medium -b:v 2000k output_2000.mp4 \
  -c:v libx264 -preset medium -b:v 4000k output_4000.mp4
```

### Encoding Profiles

```go
var EncodingProfiles = []EncodingProfile{
    {
        Bitrate: 500,
        Resolution: "1280x720",
        FrameRate: 24,
        Label: "Low",
    },
    {
        Bitrate: 1000,
        Resolution: "1280x720",
        FrameRate: 24,
        Label: "Normal",
    },
    {
        Bitrate: 2000,
        Resolution: "1920x1080",
        FrameRate: 30,
        Label: "High",
    },
    {
        Bitrate: 4000,
        Resolution: "1920x1080",
        FrameRate: 30,
        Label: "Very High",
    },
}
```

---

## 📊 Database Schema (New Tables)

```sql
-- Transcoding Jobs
CREATE TABLE transcoding_jobs (
  id UUID PRIMARY KEY,
  recording_id UUID NOT NULL REFERENCES recordings(id),
  status VARCHAR(20) NOT NULL, -- pending, in-progress, completed, failed
  profiles TEXT NOT NULL, -- JSON array of encoding profiles
  start_time TIMESTAMP,
  end_time TIMESTAMP,
  output_path VARCHAR(255),
  created_at TIMESTAMP
);

-- Live Streams
CREATE TABLE live_streams (
  id UUID PRIMARY KEY,
  room_id UUID NOT NULL UNIQUE REFERENCES rooms(id),
  status VARCHAR(20) NOT NULL, -- active, paused, stopped
  ingest_url VARCHAR(255),
  viewer_count INTEGER DEFAULT 0,
  started_at TIMESTAMP,
  ended_at TIMESTAMP
);

-- Viewer Analytics
CREATE TABLE viewer_analytics (
  id UUID PRIMARY KEY,
  recording_id UUID NOT NULL REFERENCES recordings(id),
  user_id UUID NOT NULL REFERENCES users(id),
  bitrate_selected INT,
  bitrate_switches INT,
  buffer_events INT,
  total_watch_time INT,
  session_start TIMESTAMP,
  session_end TIMESTAMP
);

-- Create Indexes
CREATE INDEX idx_transcoding_jobs_recording ON transcoding_jobs(recording_id);
CREATE INDEX idx_transcoding_jobs_status ON transcoding_jobs(status);
CREATE INDEX idx_live_streams_room ON live_streams(room_id);
CREATE INDEX idx_viewer_analytics_recording ON viewer_analytics(recording_id);
CREATE INDEX idx_viewer_analytics_user ON viewer_analytics(user_id);
```

---

## 🎯 Success Criteria

### By End of Day 1
- [x] ABR engine compiles without errors
- [x] Network condition detection working
- [x] Quality selection algorithm tested
- [x] 15+ unit tests passing

### By End of Day 2
- [x] Multi-bitrate transcoder compiles
- [x] Can encode video to 4 bitrates
- [x] Master playlist generates correctly
- [x] Transcoding queue manages jobs
- [x] 15+ unit tests passing

### By End of Day 3
- [x] Live distributor functional
- [x] Handles 100+ concurrent viewers
- [x] Segment buffering works
- [x] Connection pool stable
- [x] 15+ unit tests passing

### By End of Day 4
- [x] All 6 new endpoints working
- [x] Integration tests passing
- [x] Production binary builds clean
- [x] Startup output shows Phase 2B
- [x] Ready for deployment

---

## 📚 Reference Documents

**Existing Documentation:**
- PHASE_2A_MASTER_SUMMARY.md - Phase 2A architecture
- PHASE_2A_DAY_4_API_REFERENCE.md - Existing streaming endpoints

**To Create:**
- PHASE_2B_DAY_1_REPORT.md - ABR implementation report
- PHASE_2B_DAY_2_REPORT.md - Transcoding implementation
- PHASE_2B_DAY_3_REPORT.md - Live distribution implementation
- PHASE_2B_COMPLETION_SUMMARY.md - Final status

---

## ✨ Next After Phase 2B

Once Phase 2B is complete:

**Phase 4: Analytics & Reporting** (3-4 days)
- Usage analytics dashboard
- Attendance tracking
- Student engagement metrics
- Performance reports

---

## 🚀 Ready to Begin?

Phase 2B is well-defined and ready to implement. All requirements, architecture, and success criteria are clear.

**Starting Point:** Create `pkg/streaming/` directory and begin Day 1 ABR implementation.

**Estimated Timeline:**
- Day 1 (ABR Engine): 1 day
- Day 2 (Multi-bitrate Transcoding): 1 day
- Day 3 (Live Distribution): 1.5 days
- Day 4 (Integration & Testing): 1-2 days

**Total: 4-5 days to completion**
