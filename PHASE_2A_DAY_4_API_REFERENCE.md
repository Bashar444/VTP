# Phase 2A Day 4 - Streaming & Playback API Reference

## Overview

Complete REST API reference for streaming and playback endpoints. All endpoints require authentication via Bearer JWT token in Authorization header.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

All endpoints require Bearer token:

```
Authorization: Bearer <JWT_TOKEN>
```

Obtain token via `/api/v1/auth/login`:

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"pass"}'
```

---

## Endpoints

### 1. Stream HLS Playlist

**GET** `/recordings/{id}/stream/playlist.m3u8`

Stream a recording as HLS (HTTP Live Streaming). Returns M3U8 master playlist.

#### Parameters

| Name | Type | Location | Required | Description |
|------|------|----------|----------|-------------|
| id | UUID | Path | Yes | Recording ID |

#### Response

**Status:** 200 OK

**Content-Type:** application/vnd.apple.mpegurl

**Body:**
```m3u8
#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:10.0,
segment-000.ts
#EXTINF:10.0,
segment-001.ts
#EXT-X-ENDLIST
```

#### Headers

```
Cache-Control: no-cache, no-store, must-revalidate
Pragma: no-cache
Expires: 0
```

#### Example

```bash
# Get HLS playlist
curl -H "Authorization: Bearer <TOKEN>" \
  http://localhost:8080/api/v1/recordings/550e8400-e29b-41d4-a716-446655440000/stream/playlist.m3u8

# Use with VLC
vlc http://localhost:8080/api/v1/recordings/550e8400-e29b-41d4-a716-446655440000/stream/playlist.m3u8
```

#### Notes

- Recording must have status `completed`
- Streaming must be ready (transcoding complete)
- Segments cached for 1 hour
- Supports adaptive bitrate selection

---

### 2. Stream HLS Segment

**GET** `/recordings/{id}/stream/{segment}.ts`

Download individual HLS TS (transport stream) segment.

#### Parameters

| Name | Type | Location | Required | Description |
|------|------|----------|----------|-------------|
| id | UUID | Path | Yes | Recording ID |
| segment | String | Path | Yes | Segment filename (e.g., segment-000.ts) |

#### Response

**Status:** 200 OK

**Content-Type:** video/mp2t

**Body:** Binary TS segment data

#### Headers

```
Content-Length: <size_in_bytes>
Cache-Control: max-age=3600
Accept-Ranges: bytes
```

#### Example

```bash
# Download segment
curl -H "Authorization: Bearer <TOKEN>" \
  -o segment-000.ts \
  http://localhost:8080/api/v1/recordings/550e8400-e29b-41d4-a716-446655440000/stream/segment-000.ts

# Range request (for seeking)
curl -H "Authorization: Bearer <TOKEN>" \
  -H "Range: bytes=0-1048575" \
  http://localhost:8080/api/v1/recordings/550e8400-e29b-41d4-a716-446655440000/stream/segment-000.ts
```

#### Notes

- Supports HTTP Range requests for seeking
- ~500KB per 10-second segment typical
- Direct browser playback supported

---

### 3. Initiate Transcoding

**POST** `/recordings/{id}/transcode`

Start background transcoding of recording to streaming format.

#### Parameters

| Name | Type | Location | Required | Description |
|------|------|----------|----------|-------------|
| id | UUID | Path | Yes | Recording ID |
| format | String | Query | No | Format: hls, mp4 (default: hls) |

#### Request

```bash
curl -X POST \
  -H "Authorization: Bearer <TOKEN>" \
  http://localhost:8080/api/v1/recordings/550e8400-e29b-41d4-a716-446655440000/transcode?format=hls
```

#### Response

**Status:** 202 Accepted

**Content-Type:** application/json

**Body:**
```json
{
  "status": "transcoding_started",
  "format": "hls"
}
```

#### Errors

| Status | Code | Message | Cause |
|--------|------|---------|-------|
| 400 | BadRequest | Invalid recording ID | Malformed UUID |
| 404 | NotFound | Recording not found | ID doesn't exist |
| 400 | BadRequest | Recording not finished | Status != completed |

#### Example

```bash
# Start HLS transcoding
curl -X POST \
  -H "Authorization: Bearer <TOKEN>" \
  http://localhost:8080/api/v1/recordings/550e8400-e29b-41d4-a716-446655440000/transcode?format=hls

# Start MP4 transcoding
curl -X POST \
  -H "Authorization: Bearer <TOKEN>" \
  http://localhost:8080/api/v1/recordings/550e8400-e29b-41d4-a716-446655440000/transcode?format=mp4

# Poll for completion
curl -H "Authorization: Bearer <TOKEN>" \
  http://localhost:8080/api/v1/recordings/550e8400-e29b-41d4-a716-446655440000/info
```

#### Notes

- Non-blocking: returns immediately while transcoding in background
- Typical duration: 1-10 minutes depending on recording length
- Max timeout: 1 hour
- Can check status via `GET /recordings/{id}/info`

---

### 4. Track Playback Progress

**POST** `/recordings/{id}/progress`

Update playback progress and analytics for a viewing session.

#### Parameters

| Name | Type | Location | Required | Description |
|------|------|----------|----------|-------------|
| id | UUID | Path | Yes | Recording ID |

#### Request

**Content-Type:** application/x-www-form-urlencoded

**Body:**
```
position=45
```

Or as JSON:
```json
{
  "position": 45
}
```

#### Response

**Status:** 200 OK

**Content-Type:** application/json

**Body:**
```json
{
  "position": 45
}
```

#### Example

```bash
# Send playback progress (45 seconds into video)
curl -X POST \
  -H "Authorization: Bearer <TOKEN>" \
  -d "position=45" \
  http://localhost:8080/api/v1/recordings/550e8400-e29b-41d4-a716-446655440000/progress

# Send with JSON
curl -X POST \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"position":45}' \
  http://localhost:8080/api/v1/recordings/550e8400-e29b-41d4-a716-446655440000/progress
```

#### Notes

- Send periodically during playback (every 5-10 seconds recommended)
- Position in seconds
- Updates `recording_access_log` table
- User ID extracted from JWT token
- Tracks session duration and completion rate

---

### 5. Get Recording Thumbnail

**GET** `/recordings/{id}/thumbnail`

Download thumbnail preview image for recording.

#### Parameters

| Name | Type | Location | Required | Description |
|------|------|----------|----------|-------------|
| id | UUID | Path | Yes | Recording ID |

#### Response

**Status:** 200 OK

**Content-Type:** image/jpeg

**Body:** JPEG image binary data

#### Headers

```
Content-Length: <size_in_bytes>
Cache-Control: max-age=86400
```

#### Example

```bash
# Download thumbnail
curl -H "Authorization: Bearer <TOKEN>" \
  -o thumbnail.jpg \
  http://localhost:8080/api/v1/recordings/550e8400-e29b-41d4-a716-446655440000/thumbnail

# Use in HTML
<img src="http://localhost:8080/api/v1/recordings/550e8400-e29b-41d4-a716-446655440000/thumbnail" />
```

#### Errors

| Status | Code | Message | Cause |
|--------|------|---------|-------|
| 404 | NotFound | Thumbnail not available | Not yet generated |

#### Notes

- Dimensions: 320x180 pixels
- Generated automatically on first request
- Cached for 1 day
- File size typical: 15-30 KB

---

### 6. Get Playback Analytics

**GET** `/recordings/{id}/analytics`

Retrieve detailed playback analytics for a recording.

#### Parameters

| Name | Type | Location | Required | Description |
|------|------|----------|----------|-------------|
| id | UUID | Path | Yes | Recording ID |

#### Response

**Status:** 200 OK

**Content-Type:** application/json

**Body:**
```json
{
  "recording_id": "550e8400-e29b-41d4-a716-446655440000",
  "total_sessions": 42,
  "unique_viewers": 38,
  "total_playtime": 3600,
  "average_playtime": 85,
  "last_accessed_at": "2025-11-24T15:30:45Z"
}
```

#### Fields

| Field | Type | Description |
|-------|------|-------------|
| recording_id | UUID | Recording identifier |
| total_sessions | Int | Total playback sessions |
| unique_viewers | Int | Number of unique users who viewed |
| total_playtime | Int | Total seconds watched across all sessions |
| average_playtime | Int | Average seconds per session |
| last_accessed_at | ISO 8601 | Last playback timestamp |

#### Example

```bash
# Get analytics
curl -H "Authorization: Bearer <TOKEN>" \
  http://localhost:8080/api/v1/recordings/550e8400-e29b-41d4-a716-446655440000/analytics

# Parse with jq
curl -s -H "Authorization: Bearer <TOKEN>" \
  http://localhost:8080/api/v1/recordings/550e8400-e29b-41d4-a716-446655440000/analytics | \
  jq '.total_sessions, .average_playtime'
```

#### Notes

- Data aggregated from `recording_access_log` table
- Updates in real-time as playback progress is sent
- Completion rate calculated from total_playtime / (total_sessions * duration)
- Can be used for instructor dashboards

---

### 7. Get Recording Info

**GET** `/recordings/{id}/info`

Get complete recording metadata and streaming status.

#### Parameters

| Name | Type | Location | Required | Description |
|------|------|----------|----------|-------------|
| id | UUID | Path | Yes | Recording ID |

#### Response

**Status:** 200 OK

**Content-Type:** application/json

**Body:**
```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "CS101 Lecture 5",
    "description": "Introduction to Streaming",
    "status": "completed",
    "duration": 3600,
    "created_at": "2025-11-24T10:00:00Z",
    "stopped_at": "2025-11-24T11:00:00Z",
    "size_bytes": 512000000,
    "format": "webm",
    "streaming_ready": true,
    "streaming_url": "/api/v1/recordings/550e8400-e29b-41d4-a716-446655440000/stream/playlist.m3u8",
    "analytics": {
      "total_sessions": 42,
      "unique_viewers": 38,
      "total_playtime": 3600,
      "average_playtime": 85,
      "last_accessed_at": "2025-11-24T15:30:45Z"
    }
  }
}
```

#### Fields

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Recording ID |
| title | String | Display name |
| description | String | Optional description |
| status | Enum | pending, recording, processing, completed, failed |
| duration | Int | Duration in seconds |
| created_at | ISO 8601 | When recording started |
| stopped_at | ISO 8601 | When recording ended |
| size_bytes | Int64 | File size in bytes |
| format | String | Original format (webm) |
| streaming_ready | Boolean | If HLS ready |
| streaming_url | String | HLS master playlist URL |
| analytics | Object | Playback statistics |

#### Example

```bash
# Get info
curl -H "Authorization: Bearer <TOKEN>" \
  http://localhost:8080/api/v1/recordings/550e8400-e29b-41d4-a716-446655440000/info

# Get status only
curl -s -H "Authorization: Bearer <TOKEN>" \
  http://localhost:8080/api/v1/recordings/550e8400-e29b-41d4-a716-446655440000/info | \
  jq '.data | {status, streaming_ready, size_bytes}'
```

#### Errors

| Status | Code | Message | Cause |
|--------|------|---------|-------|
| 404 | NotFound | Recording not found | ID doesn't exist |

#### Notes

- Most complete recording info endpoint
- Use to check streaming_ready status before embedding player
- Size in bytes useful for storage estimates
- Shows full analytics dashboard

---

## Status Codes

| Code | Meaning | Use Case |
|------|---------|----------|
| 200 | OK | Successful request completed |
| 202 | Accepted | Transcoding started in background |
| 400 | Bad Request | Invalid parameters or state |
| 401 | Unauthorized | Missing/invalid JWT token |
| 404 | Not Found | Recording doesn't exist |
| 500 | Internal Server Error | Transcoding failed or storage error |

## Recording Status Enum

```
pending      - Just created, not yet recording
recording    - Currently being recorded (not finished)
processing   - Recording finished, post-processing in progress
completed    - Ready for playback
failed       - Error occurred during processing
archived     - Stored in long-term archival
deleted      - Soft-deleted, can be restored
```

## Streaming Quality Options

| Quality | Bitrate | Resolution | Use Case |
|---------|---------|-----------|----------|
| low | 500 kbps | 1280x720 | Mobile, low bandwidth |
| medium | 1500 kbps | 1280x720 | Standard streaming |
| high | 3000 kbps | 1920x1080 | High quality |
| ultra | 6000 kbps | 1920x1080 | 4K/archival |

## Error Handling

### Standard Error Response

```json
{
  "error": "Recording not found",
  "status": 404,
  "message": "The requested recording does not exist"
}
```

### Common Errors

**Invalid UUID Format**
```
Status: 400
Message: Invalid recording ID
```

**Recording Not Yet Finished**
```
Status: 400
Message: Recording not yet finished
```

**Streaming Not Ready**
```
Status: 400
Message: Streaming not ready for recording
```

**Thumbnail Not Available**
```
Status: 404
Message: Thumbnail not available
```

## Rate Limiting

- **Transcoding:** 1 per recording at a time
- **Segment Streaming:** No limit (CDN recommended for production)
- **Playlist Requests:** No limit
- **Analytics Queries:** 100/minute recommended

## Security Considerations

1. **JWT Validation:** All endpoints validate Bearer token
2. **Path Traversal:** Segment paths validated within recording directory
3. **Access Control:** User_id extracted from JWT for audit trails
4. **File Permissions:** Storage directory 0755, files 0644
5. **Timeouts:** All operations timeout (10s - 1h depending on operation)

## Performance Tips

1. **Caching:** Cache playlist M3U8 with `max-age=0, must-revalidate`
2. **Segments:** Cache TS segments with `max-age=3600`
3. **CDN:** Serve segments from CDN in production
4. **Compression:** Don't compress video segments (already compressed)
5. **Parallel Downloads:** Player can fetch multiple segments in parallel

## Example Workflow

### Complete Playback Flow

```bash
# 1. Get recording info
RECORDING_ID="550e8400-e29b-41d4-a716-446655440000"
INFO=$(curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/recordings/$RECORDING_ID/info)

# 2. Check if streaming ready
STREAMING_READY=$(echo $INFO | jq '.data.streaming_ready')
if [ "$STREAMING_READY" != "true" ]; then
  # 3. Start transcoding if needed
  curl -X POST \
    -H "Authorization: Bearer $TOKEN" \
    http://localhost:8080/api/v1/recordings/$RECORDING_ID/transcode?format=hls
  
  # Wait for transcoding (poll every 10 seconds)
  sleep 10
fi

# 4. Get HLS playlist URL
PLAYLIST_URL=$(echo $INFO | jq -r '.data.streaming_url')

# 5. Stream with VLC
vlc $PLAYLIST_URL

# 6. Track progress (in player)
for position in 10 20 30 40 50; do
  curl -X POST \
    -H "Authorization: Bearer $TOKEN" \
    -d "position=$position" \
    http://localhost:8080/api/v1/recordings/$RECORDING_ID/progress
  sleep 10
done

# 7. Get analytics
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/recordings/$RECORDING_ID/analytics | jq .
```

## Troubleshooting

### Streaming Returns 404

**Cause:** Recording not finished or streaming not ready
**Solution:** POST `/recordings/{id}/transcode` first

### Segments Not Found

**Cause:** Transcoding incomplete
**Solution:** Check `/recordings/{id}/info` for `streaming_ready: true`

### Slow Transcoding

**Cause:** Large file or slow CPU
**Solution:** Use lower resolution in StreamingProfile

### Storage Full

**Cause:** /tmp/recordings directory out of space
**Solution:** Implement CleanupExpiredRecordings with retention policy

## Version History

| Date | Change |
|------|--------|
| 2025-11-24 | Day 4 implementation - Streaming, Playback, Analytics |
| 2025-11-23 | Day 3 implementation - Storage & Download |
| 2025-11-22 | Day 2 implementation - FFmpeg & Handlers |
| 2025-11-21 | Day 1 implementation - Database & Types |
