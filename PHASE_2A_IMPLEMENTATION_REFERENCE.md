# Phase 2A Recording System - Implementation Reference

**Quick Reference Guide for Recording Package**

---

## Using the Recording Service

### Initialize Service
```go
import "github.com/yourusername/vtp-platform/pkg/recording"

// Create service instance
recordingService := recording.NewRecordingService(db, logger)
```

### Start Recording
```go
req := &recording.StartRecordingRequest{
    RoomID:      roomUUID,
    Title:       "Lecture on Go Concurrency",
    Description: ptr("Recorded on Nov 21, 2025"),
    Format:      "webm",
    BitrateKbps: ptr(2500),
    FrameRate:   ptr(30),
}

recording, err := recordingService.StartRecording(ctx, req, userID)
if err != nil {
    log.Printf("Failed to start recording: %v", err)
}

// Access recording ID
recordingID := recording.ID
// Status will be: StatusPending
```

### Stop Recording
```go
stopped, err := recordingService.StopRecording(ctx, recordingID)
if err != nil {
    log.Printf("Failed to stop recording: %v", err)
}

// Access recording details
duration := *stopped.DurationSeconds  // in seconds
status := stopped.Status              // StatusCompleted
```

### Get Recording Details
```go
recording, err := recordingService.GetRecording(ctx, recordingID)
if err != nil {
    log.Printf("Error: %v", err)
}

if recording != nil {
    fmt.Printf("Recording: %s\n", recording.Title)
    fmt.Printf("Duration: %d seconds\n", *recording.DurationSeconds)
    fmt.Printf("Status: %s\n", recording.Status)
    fmt.Printf("File: %s\n", *recording.FilePath)
}
```

### List Recordings
```go
query := &recording.RecordingListQuery{
    RoomID: roomUUID,
    Status: recording.StatusCompleted,
    Limit:  10,
    Offset: 0,
}

recordings, total, err := recordingService.ListRecordings(ctx, query)
if err != nil {
    log.Printf("Error: %v", err)
}

fmt.Printf("Found %d recordings (total: %d)\n", len(recordings), total)
for _, rec := range recordings {
    fmt.Printf("- %s (%s)\n", rec.Title, rec.Status)
}
```

### Delete Recording
```go
deleted, err := recordingService.DeleteRecording(ctx, recordingID)
if err != nil {
    log.Printf("Error: %v", err)
}

fmt.Printf("Recording deleted at: %v\n", deleted.DeletedAt)
```

### Update Status
```go
err := recordingService.UpdateRecordingStatus(ctx, recordingID, recording.StatusProcessing)
if err != nil {
    log.Printf("Error: %v", err)
}
```

### Update Metadata
```go
metadata := map[string]interface{}{
    "processor": "ffmpeg",
    "quality": "1080p",
    "compression": 0.85,
}

err := recordingService.UpdateRecordingMetadata(ctx, recordingID, metadata)
if err != nil {
    log.Printf("Error: %v", err)
}
```

### Get Statistics
```go
stats, err := recordingService.GetRecordingStats(ctx, recordingID)
if err != nil {
    log.Printf("Error: %v", err)
}

fmt.Printf("Participants: %v\n", stats["participant_count"])
fmt.Printf("Access Count: %v\n", stats["access_count"])
fmt.Printf("Duration: %v seconds\n", stats["duration_seconds"])
fmt.Printf("File Size: %v bytes\n", stats["file_size_bytes"])
```

---

## Status Lifecycle

```
StartRecording() → Recording Status: StatusPending
       ↓
StopRecording()  → Recording Status: StatusCompleted
       ↓
UpdateRecordingStatus(StatusProcessing) → Processing file
       ↓
UpdateRecordingStatus(StatusCompleted) → Ready for playback
       ↓
DeleteRecording() → Status: StatusDeleted (soft delete)
```

## Status Constants

```go
recording.StatusPending      // = "pending"      - Just created
recording.StatusRecording    // = "recording"    - Active recording
recording.StatusProcessing   // = "processing"   - Being processed
recording.StatusCompleted    // = "completed"    - Ready for playback
recording.StatusFailed       // = "failed"       - Recording failed
recording.StatusArchived     // = "archived"     - Archived/moved
recording.StatusDeleted      // = "deleted"      - Soft deleted
```

---

## Request/Response Examples

### StartRecordingRequest
```go
{
    "room_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Lecture on Distributed Systems",
    "description": "Prof. Johnson's lecture",
    "format": "webm",
    "bitrate_kbps": 2500,
    "frame_rate": 30
}
```

### StartRecordingResponse
```go
{
    "recording_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
    "status": "pending",
    "started_at": "2025-11-21T14:30:00Z",
    "room_id": "550e8400-e29b-41d4-a716-446655440000",
    "message": "Recording initialized"
}
```

### StopRecordingResponse
```go
{
    "recording_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
    "status": "completed",
    "stopped_at": "2025-11-21T15:30:00Z",
    "duration_seconds": 3600,
    "file_path": "/recordings/6ba7b810-9dad-11d1-80b4-00c04fd430c8.webm"
}
```

### GetRecordingResponse
```go
{
    "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
    "room_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Lecture on Distributed Systems",
    "description": "Prof. Johnson's lecture",
    "started_by": "7ba7b810-9dad-11d1-80b4-00c04fd430c9",
    "started_at": "2025-11-21T14:30:00Z",
    "stopped_at": "2025-11-21T15:30:00Z",
    "duration_seconds": 3600,
    "status": "completed",
    "format": "webm",
    "file_size_bytes": 947483648,
    "mime_type": "video/webm",
    "bitrate_kbps": 2500,
    "frame_rate_fps": 30,
    "resolution": "1920x1080",
    "codecs": "vp9,opus",
    "metadata": {
        "processor": "ffmpeg",
        "quality": "1080p"
    },
    "created_at": "2025-11-21T14:30:00Z",
    "updated_at": "2025-11-21T15:30:00Z"
}
```

### ListRecordingsResponse
```go
{
    "recordings": [
        {
            "id": "...",
            "title": "Lecture 1",
            "status": "completed",
            ...
        },
        {
            "id": "...",
            "title": "Lecture 2",
            "status": "completed",
            ...
        }
    ],
    "total": 25,
    "limit": 10,
    "offset": 0
}
```

---

## Error Handling

All service methods follow consistent error patterns:

```go
// Error example
_, err := recordingService.GetRecording(ctx, uuid.Nil)
if err != nil {
    if strings.Contains(err.Error(), "recording_id is required") {
        // Handle validation error
    } else {
        // Handle database error
    }
}
```

Common errors:
- `"recording_id is required"` - UUID is nil
- `"recording not found"` - No matching record
- `"cannot stop a deleted recording"` - Logic error
- `"invalid status: ..."` - Status validation failed
- `"failed to start recording: <db error>"` - Database error

---

## Database Operations

### Example: Raw SQL (rarely needed)
```go
// Get recording count by room
var count int
err := db.QueryRowContext(ctx, 
    "SELECT COUNT(*) FROM recordings WHERE room_id = $1 AND deleted_at IS NULL",
    roomID).Scan(&count)
```

### Indexes Available
```
idx_recordings_room_id         - Fast room lookups
idx_recordings_started_by      - Fast user lookups
idx_recordings_status          - Fast status filters
idx_recordings_created_at      - Fast sorting by creation
idx_recordings_started_at      - Fast sorting by start time
```

---

## Validation Functions

```go
// Validate recording status
if !recording.ValidateStatus(statusString) {
    log.Printf("Invalid status: %s", statusString)
}

// Validate access level
if !recording.ValidateAccessLevel(levelString) {
    log.Printf("Invalid access level: %s", levelString)
}

// Validate share type
if !recording.ValidateShareType(typeString) {
    log.Printf("Invalid share type: %s", typeString)
}
```

---

## Helper Function

```go
// Helper to create pointers for optional fields
func ptr[T any](v T) *T {
    return &v
}

// Usage
description := ptr("Recording of today's class")
bitrate := ptr(2500)
```

---

## Next Implementation (Day 2)

The following will be implemented in Day 2:

1. **FFmpeg Integration** - Actual recording process
2. **HTTP Handlers** - API endpoints for service methods
3. **Participant Tracking** - Recording who was in the session
4. **File Management** - Store and manage recorded files

---

## Testing Your Implementation

```bash
# Run all tests
go test ./pkg/recording -v

# Run specific test
go test ./pkg/recording -v -run TestValidateStatus

# Run with coverage
go test ./pkg/recording -cover

# Build the package
go build ./pkg/recording
```

---

**For complete examples, see:**
- `PHASE_2A_QUICK_START.md` - Implementation code templates
- `PHASE_2A_DAY_1_COMPLETE.md` - Day 1 completion details
- `PHASE_2A_PLANNING.md` - Full architecture design
