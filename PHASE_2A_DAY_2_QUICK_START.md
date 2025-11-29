# Phase 2A Day 2 - Quick Start Guide

**Ready to Continue?** Follow this guide to begin Day 2 implementation.

---

## 🎯 Day 2 Objectives

**FFmpeg Integration & HTTP Handlers** (4-5 hours)

| Task | Time | Status |
|------|------|--------|
| Create FFmpeg wrapper | 1.5h | ⏳ TODO |
| Create HTTP handlers | 1.5h | ⏳ TODO |
| Create participant tracking | 1h | ⏳ TODO |
| Update main.go integration | 0.5h | ⏳ TODO |
| Testing & validation | 1h | ⏳ TODO |

---

## 📝 Day 2 Files to Create

### 1. pkg/recording/ffmpeg.go (250+ lines)

```go
package recording

import (
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"time"
)

// FFmpegProcess represents an active FFmpeg recording process
type FFmpegProcess struct {
	mu              sync.Mutex
	cmd             *exec.Cmd
	stdin           io.WriteCloser
	stderr          io.ReadCloser
	recordingID     string
	status          string
	startedAt       time.Time
	bytesProcessed  int64
	lastError       string
	logger          *log.Logger
}

// StartFFmpeg begins recording with FFmpeg
func (fp *FFmpegProcess) StartFFmpeg(ctx context.Context, inputFormat, outputPath string) error {
	// Implementation follows

	// FFmpeg command structure:
	// ffmpeg -f format -i pipe:0 -c:v codec -c:a codec -b:v bitrate output.webm
	
	return nil
}

// StopFFmpeg gracefully stops the FFmpeg process
func (fp *FFmpegProcess) StopFFmpeg(ctx context.Context) error {
	// Implementation follows
	return nil
}

// GetStatus returns current process status
func (fp *FFmpegProcess) GetStatus() map[string]interface{} {
	// Implementation follows
	return map[string]interface{}{}
}

// WriteAudioFrame writes audio data to FFmpeg
func (fp *FFmpegProcess) WriteAudioFrame(data []byte) error {
	// Implementation follows
	return nil
}

// WriteVideoFrame writes video data to FFmpeg
func (fp *FFmpegProcess) WriteVideoFrame(data []byte) error {
	// Implementation follows
	return nil
}
```

**Key Methods:**
- NewFFmpegProcess() - Create process
- StartFFmpeg() - Launch with proper arguments
- StopFFmpeg() - Graceful shutdown
- GetStatus() - Check health
- WriteAudioFrame() - Audio stream
- WriteVideoFrame() - Video stream

**Features:**
- Process management (start/stop/monitor)
- Stream handling (audio/video pipes)
- Error recovery and logging
- Graceful shutdown
- Status tracking

### 2. pkg/recording/handlers.go (300+ lines)

```go
package recording

import (
	"encoding/json"
	"log"
	"net/http"
	"github.com/google/uuid"
)

// RecordingHandlers wraps all HTTP handlers
type RecordingHandlers struct {
	service *RecordingService
	logger  *log.Logger
}

// NewRecordingHandlers creates handler instance
func NewRecordingHandlers(service *RecordingService, logger *log.Logger) *RecordingHandlers {
	return &RecordingHandlers{
		service: service,
		logger:  logger,
	}
}

// StartRecordingHandler handles POST /api/v1/recordings/start
func (h *RecordingHandlers) StartRecordingHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request
	// Validate user
	// Call service.StartRecording()
	// Return response
}

// StopRecordingHandler handles POST /api/v1/recordings/{id}/stop
func (h *RecordingHandlers) StopRecordingHandler(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL
	// Call service.StopRecording()
	// Return response
}

// GetRecordingHandler handles GET /api/v1/recordings/{id}
func (h *RecordingHandlers) GetRecordingHandler(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL
	// Call service.GetRecording()
	// Return recording details
}

// ListRecordingsHandler handles GET /api/v1/recordings?...
func (h *RecordingHandlers) ListRecordingsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	// Call service.ListRecordings()
	// Return paginated results
}

// DeleteRecordingHandler handles DELETE /api/v1/recordings/{id}
func (h *RecordingHandlers) DeleteRecordingHandler(w http.ResponseWriter, r *http.Request) {
	// Parse ID from URL
	// Call service.DeleteRecording()
	// Return confirmation
}

// Helper methods
func (h *RecordingHandlers) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *RecordingHandlers) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}
```

**Key Handlers:**
- StartRecordingHandler() - Begin recording
- StopRecordingHandler() - End recording
- GetRecordingHandler() - Get details
- ListRecordingsHandler() - List with filters
- DeleteRecordingHandler() - Delete recording

**Features:**
- Request parsing and validation
- User authentication checks
- Proper HTTP status codes
- JSON marshalling
- Error responses

### 3. pkg/recording/participant.go (200+ lines)

```go
package recording

import (
	"context"
	"fmt"
	"log"
	"github.com/google/uuid"
)

// ParticipantManager tracks recording participants
type ParticipantManager struct {
	db          interface{} // *sql.DB
	recordingID uuid.UUID
	peers       map[string]*RecordingParticipant
	logger      *log.Logger
}

// AddParticipant adds a user to recording
func (pm *ParticipantManager) AddParticipant(ctx context.Context, userID, peerID uuid.UUID) error {
	// Create participant record
	// Track in memory
	// Store in database
	return nil
}

// RemoveParticipant removes a user from recording
func (pm *ParticipantManager) RemoveParticipant(ctx context.Context, peerID string) error {
	// Update left_at timestamp
	// Remove from memory
	// Update database
	return nil
}

// UpdateParticipantStats updates participant statistics
func (pm *ParticipantManager) UpdateParticipantStats(ctx context.Context, peerID string, stats map[string]interface{}) error {
	// Update bytes sent/received
	// Update packet counts
	// Update database
	return nil
}

// GetParticipants returns all participants in recording
func (pm *ParticipantManager) GetParticipants(ctx context.Context) ([]RecordingParticipant, error) {
	// Query database
	// Return list
	return []RecordingParticipant{}, nil
}
```

**Key Methods:**
- AddParticipant() - Add user to recording
- RemoveParticipant() - Remove user
- UpdateParticipantStats() - Update statistics
- GetParticipants() - List all participants

**Features:**
- Mediasoup peer integration
- Real-time tracking
- Statistics collection
- Database persistence

---

## 🔧 Integration with main.go

Add to your main.go:

```go
import (
	"github.com/yourusername/vtp-platform/pkg/recording"
)

func main() {
	// ... existing code ...

	// Initialize recording service
	recordingService := recording.NewRecordingService(db, logger)
	recordingHandlers := recording.NewRecordingHandlers(recordingService, logger)

	// Register HTTP routes
	router.POST("/api/v1/recordings/start", recordingHandlers.StartRecordingHandler)
	router.POST("/api/v1/recordings/:id/stop", recordingHandlers.StopRecordingHandler)
	router.GET("/api/v1/recordings/:id", recordingHandlers.GetRecordingHandler)
	router.GET("/api/v1/recordings", recordingHandlers.ListRecordingsHandler)
	router.DELETE("/api/v1/recordings/:id", recordingHandlers.DeleteRecordingHandler)

	// ... rest of main.go ...
}
```

---

## 🧪 Testing Strategy

### Unit Tests for Handlers

```go
func TestStartRecordingHandler(t *testing.T) {
	// Create mock service
	// Create request
	// Call handler
	// Verify response
}

func TestStopRecordingHandler(t *testing.T) {
	// Similar to above
}
```

### Integration Tests

```bash
# Start backend
go run cmd/main.go

# Test start recording
curl -X POST http://localhost:8080/api/v1/recordings/start \
  -H "Content-Type: application/json" \
  -d '{"room_id":"...","title":"Test"}'

# Test stop recording
curl -X POST http://localhost:8080/api/v1/recordings/{id}/stop

# Test list
curl http://localhost:8080/api/v1/recordings

# Test get
curl http://localhost:8080/api/v1/recordings/{id}

# Test delete
curl -X DELETE http://localhost:8080/api/v1/recordings/{id}
```

---

## 📋 Day 2 Checklist

### FFmpeg Implementation
- [ ] Create pkg/recording/ffmpeg.go
- [ ] Implement FFmpegProcess struct
- [ ] Implement StartFFmpeg() method
- [ ] Implement StopFFmpeg() method
- [ ] Implement GetStatus() method
- [ ] Implement WriteAudioFrame() method
- [ ] Implement WriteVideoFrame() method
- [ ] Add error handling
- [ ] Add logging
- [ ] Write unit tests

### HTTP Handlers
- [ ] Create pkg/recording/handlers.go
- [ ] Implement StartRecordingHandler
- [ ] Implement StopRecordingHandler
- [ ] Implement GetRecordingHandler
- [ ] Implement ListRecordingsHandler
- [ ] Implement DeleteRecordingHandler
- [ ] Add request validation
- [ ] Add response marshalling
- [ ] Add error handling
- [ ] Write handler tests

### Participant Tracking
- [ ] Create pkg/recording/participant.go
- [ ] Implement ParticipantManager
- [ ] Implement AddParticipant()
- [ ] Implement RemoveParticipant()
- [ ] Implement UpdateParticipantStats()
- [ ] Implement GetParticipants()
- [ ] Add Mediasoup integration
- [ ] Write tests

### Integration
- [ ] Update cmd/main.go
- [ ] Register HTTP routes
- [ ] Initialize recording service
- [ ] Initialize handlers
- [ ] Verify compilation
- [ ] Test endpoints with curl
- [ ] Test with actual backend

### Testing & Validation
- [ ] Unit tests pass
- [ ] Handler tests pass
- [ ] HTTP endpoints respond
- [ ] Database operations work
- [ ] Error handling works
- [ ] Logging shows progress
- [ ] No compilation errors

---

## 🚀 Getting Started

### Step 1: Create FFmpeg Handler
```bash
# Create the file
touch pkg/recording/ffmpeg.go

# Add the skeleton from above
# Start implementing StartFFmpeg() and related methods
```

### Step 2: Create HTTP Handlers
```bash
# Create the file
touch pkg/recording/handlers.go

# Add all handler functions
# Connect to service methods
# Add JSON marshalling
```

### Step 3: Create Participant Manager
```bash
# Create the file
touch pkg/recording/participant.go

# Implement participant tracking
# Connect to database
# Add Mediasoup integration
```

### Step 4: Update main.go
```bash
# Add imports
# Initialize service and handlers
# Register routes
```

### Step 5: Test
```bash
# Compile
go build ./...

# Run tests
go test ./pkg/recording -v

# Test HTTP endpoints
go run cmd/main.go
curl -X POST http://localhost:8080/api/v1/recordings/start ...
```

---

## 📚 Reference Documentation

- **FFmpeg Command:** See PHASE_2A_QUICK_START.md
- **HTTP Handlers:** See PHASE_2A_QUICK_START.md Day 2
- **Service Methods:** See PHASE_2A_IMPLEMENTATION_REFERENCE.md
- **Database Schema:** See PHASE_2A_DAY_1_COMPLETE.md

---

## ⏱️ Time Estimates

| Task | Easy | Normal | Hard |
|------|------|--------|------|
| FFmpeg wrapper | 1.5h | 2h | 2.5h |
| HTTP handlers | 1h | 1.5h | 2h |
| Participant tracking | 0.75h | 1.5h | 2h |
| Integration & testing | 0.75h | 1.5h | 2h |
| **Total Day 2** | **4h** | **6.5h** | **8.5h** |

---

## ✅ Success Criteria

- [x] All 3 new files created
- [x] All methods implemented
- [x] Code compiles without errors
- [x] All tests pass
- [x] HTTP endpoints respond
- [x] Database operations work
- [x] Zero runtime errors
- [x] Ready for Day 3

---

**Ready to Start?** Create pkg/recording/ffmpeg.go and begin with the FFmpeg implementation!

Estimated completion: 4-6 hours from now

Good luck! 🚀
