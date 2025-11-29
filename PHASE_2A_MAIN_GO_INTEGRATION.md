# Phase 2A - Day 2 Main.go Integration Guide

## Overview
This guide shows how to integrate the recording service into your main.go application. All the code is ready to use.

## Integration Steps

### 1. Update Imports
Add these imports to your main.go:

```go
package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/yourusername/vtp-platform/pkg/recording"
)
```

### 2. Initialize Recording Service in main()

```go
func main() {
	// ... existing code ...

	// Get database connection (already initialized elsewhere)
	var db *sql.DB  // Your existing database connection

	// Create logger
	logger := log.New(os.Stdout, "[Recording] ", log.LstdFlags)

	// Initialize recording service
	recordingService := recording.NewRecordingService(db, logger)

	// Initialize handlers
	recordingHandlers := recording.NewRecordingHandlers(recordingService, logger)

	// Register routes
	mux := http.NewServeMux()
	recordingHandlers.RegisterRoutes(mux)

	// ... register other routes ...

	// Start server
	port := ":8080"
	logger.Printf("Starting server on %s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		logger.Fatalf("Server error: %v", err)
	}
}
```

### 3. Full Example

Here's a complete main.go example with recording service integration:

```go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"github.com/yourusername/vtp-platform/pkg/recording"
)

func main() {
	logger := log.New(os.Stdout, "[Main] ", log.LstdFlags)

	// Connect to database
	dbURL := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		"postgres",     // user
		"postgres",     // password
		"vtp_platform", // dbname
		"localhost",    // host
		"5432",         // port
	)

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Fatalf("Database connection failed: %v", err)
	}

	logger.Println("Database connected successfully")

	// Initialize recording service
	recordingService := recording.NewRecordingService(db, logger)

	// Initialize handlers
	recordingHandlers := recording.NewRecordingHandlers(recordingService, logger)

	// Setup HTTP routes
	mux := http.NewServeMux()

	// Record routes
	recordingHandlers.RegisterRoutes(mux)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	// Start server
	port := ":8080"
	logger.Printf("Starting server on %s", port)
	logger.Printf("Recording service ready")
	logger.Printf("API endpoints:")
	logger.Printf("  POST   /api/v1/recordings/start")
	logger.Printf("  POST   /api/v1/recordings/{id}/stop")
	logger.Printf("  GET    /api/v1/recordings")
	logger.Printf("  GET    /api/v1/recordings/{id}")
	logger.Printf("  DELETE /api/v1/recordings/{id}")

	if err := http.ListenAndServe(port, mux); err != nil {
		logger.Fatalf("Server error: %v", err)
	}
}
```

## API Endpoints

### Start Recording
```bash
POST /api/v1/recordings/start
Content-Type: application/json
X-User-ID: <user-uuid>

{
  "room_id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Team Meeting",
  "description": "Q4 Planning Session"
}
```

**Response (200 OK):**
```json
{
  "recording_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "status": "recording",
  "started_at": "2025-11-21T10:30:00Z",
  "room_id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "Recording started successfully"
}
```

### Stop Recording
```bash
POST /api/v1/recordings/a1b2c3d4-e5f6-7890-abcd-ef1234567890/stop
Content-Type: application/json
X-User-ID: <user-uuid>
```

**Response (200 OK):**
```json
{
  "recording_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "status": "stopped",
  "stopped_at": "2025-11-21T10:45:00Z",
  "duration_seconds": 900,
  "file_path": "/recordings/a1b2c3d4-e5f6-7890-abcd-ef1234567890.webm",
  "message": "Recording stopped successfully"
}
```

### List Recordings
```bash
GET /api/v1/recordings?room_id=550e8400-e29b-41d4-a716-446655440000&limit=10&offset=0
Content-Type: application/json
X-User-ID: <user-uuid>
```

**Response (200 OK):**
```json
{
  "recordings": [
    {
      "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
      "room_id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Team Meeting",
      "status": "stopped",
      "started_at": "2025-11-21T10:30:00Z",
      "stopped_at": "2025-11-21T10:45:00Z",
      "duration_seconds": 900
    }
  ],
  "total": 1,
  "limit": 10,
  "offset": 0
}
```

### Get Recording
```bash
GET /api/v1/recordings/a1b2c3d4-e5f6-7890-abcd-ef1234567890
Content-Type: application/json
X-User-ID: <user-uuid>
```

**Response (200 OK):**
```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "room_id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Team Meeting",
  "description": "Q4 Planning Session",
  "started_by": "550e8400-e29b-41d4-a716-446655440111",
  "started_at": "2025-11-21T10:30:00Z",
  "stopped_at": "2025-11-21T10:45:00Z",
  "duration_seconds": 900,
  "status": "stopped",
  "format": "webm",
  "file_size_bytes": 25600000,
  "mime_type": "video/webm"
}
```

### Delete Recording
```bash
DELETE /api/v1/recordings/a1b2c3d4-e5f6-7890-abcd-ef1234567890
Content-Type: application/json
X-User-ID: <user-uuid>
```

**Response (200 OK):**
```json
{
  "recording_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "status": "deleted",
  "deleted_at": "2025-11-21T11:00:00Z",
  "message": "Recording deleted successfully"
}
```

## Testing the Integration

### Using curl

```bash
# Start a recording
curl -X POST http://localhost:8080/api/v1/recordings/start \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440111" \
  -d '{
    "room_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Test Meeting"
  }'

# List recordings
curl -X GET http://localhost:8080/api/v1/recordings \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440111"

# Stop a recording
curl -X POST http://localhost:8080/api/v1/recordings/{recording_id}/stop \
  -H "X-User-ID: 550e8400-e29b-41d4-a716-446655440111"
```

## Database Setup

Before running the application, execute the migration:

```sql
-- Run the migration to create recording tables
psql -U postgres -d vtp_platform -f migrations/002_recordings_schema.sql
```

## Troubleshooting

### "user ID not found in request"
The handler couldn't find the user ID. Make sure you're sending `X-User-ID` header:
```bash
-H "X-User-ID: 550e8400-e29b-41d4-a716-446655440111"
```

### "Failed to persist participant"
Database connection issue. Verify PostgreSQL is running and the database exists.

### "Invalid recording ID"
UUID parsing failed. Ensure the ID is a valid UUID format.

## Next Steps

1. **Integrate with Mediasoup**: Connect FFmpeg capture to Mediasoup consumers
2. **Add file storage**: Handle recording file uploads to cloud storage
3. **Add authentication middleware**: Enhance security with proper auth
4. **Add monitoring**: Track CPU/memory usage during recordings
5. **Add streaming**: Enable live stream during recording

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         HTTP Clients                        │
└────────────┬────────────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────────────┐
│                    RecordingHandlers                        │
│  ├─ StartRecordingHandler                                   │
│  ├─ StopRecordingHandler                                    │
│  ├─ GetRecordingHandler                                     │
│  ├─ ListRecordingsHandler                                   │
│  └─ DeleteRecordingHandler                                  │
└────────────┬────────────────────────────────────────────────┘
             │
             ▼
┌─────────────────────────────────────────────────────────────┐
│                    RecordingService                         │
│  ├─ StartRecording()                                        │
│  ├─ StopRecording()                                         │
│  ├─ GetRecording()                                          │
│  ├─ ListRecordings()                                        │
│  ├─ DeleteRecording()                                       │
│  └─ UpdateRecordingStatus()                                 │
└────────────┬────────────────────────────────────────────────┘
             │
     ┌───────┼───────┐
     │               │
     ▼               ▼
┌─────────────┐ ┌──────────────────────┐
│ PostgreSQL  │ │ FFmpegProcess        │
│ ├─ Recording│ │ ├─ VideoInput Pipe   │
│ └─ Parti... │ │ ├─ AudioInput Pipe   │
└─────────────┘ │ └─ OutputFile        │
                └──────────────────────┘
                      ▼
                ┌──────────────────┐
                │ WebM File Output │
                └──────────────────┘
```

## Files Structure

```
pkg/recording/
├── handlers.go          # HTTP endpoint handlers (NEW)
├── participant.go       # Participant tracking (NEW)
├── ffmpeg.go           # FFmpeg subprocess management
├── service.go          # Core business logic
├── types.go            # Type definitions
└── service_test.go     # Unit tests
```

All ready to integrate!
