# Phase 1C Deployment Guide - Mediasoup SFU Integration

**Date:** November 21, 2025  
**Status:** Deployment & Integration Testing Phase  
**Objective:** Deploy Mediasoup service and validate Phase 1C integration

---

## Quick Deployment Checklist

### Step 1: Verify Prerequisites ✓
- [x] Go 1.24.0+ installed
- [x] Node.js 16+ (for Mediasoup service)
- [x] PostgreSQL/Redis running (from Phase 1a)
- [x] All code compiled and tested locally
- [x] 19/19 unit tests passing

### Step 2: Start Mediasoup Node.js Service
```powershell
cd mediasoup-sfu
npm install
npm start
```

**Expected Output:**
```
✓ Mediasoup worker created
✓ Mediasoup SFU server listening on port 3000
═══════════════════════════════════════════════════════════
  VTP Mediasoup SFU Service Started
═══════════════════════════════════════════════════════════
  Endpoint: http://localhost:3000
  Health: http://localhost:3000/health
  Listen IP: 127.0.0.1
  Announced IP: 127.0.0.1
  RTC Port Range: 40000 - 49999
═══════════════════════════════════════════════════════════
```

**Port Allocation:**
- Mediasoup HTTP API: `3000`
- WebRTC RTC Ports: `40000-49999` (10,000 ports)

### Step 3: Build Go Backend with Mediasoup Integration
```powershell
cd ..
go mod tidy
go build -o bin/vtp cmd/main.go
```

**Expected:**
- Build succeeds without errors
- Binary: `bin/vtp`

### Step 4: Start Go Backend
```powershell
# Set environment variable for Mediasoup
$env:MEDIASOUP_URL="http://localhost:3000"

# Run with database from Phase 1a
go run cmd/main.go
```

**Expected Output:**
```
═══════════════════════════════════════════════════════════════
  VTP Platform - Educational Live Video Streaming System
═══════════════════════════════════════════════════════════════

[1/5] Initializing database connection...
      ✓ Database connected

[2/5] Running database migrations...
      ✓ Migrations completed

[3/5] Initializing authentication services...
      ✓ Token service (access: 24h, refresh: 168h)
      ✓ Password service (bcrypt cost: 12)
      ✓ User store
      ✓ Auth handlers
      ✓ Auth middleware

[3b/5] Initializing WebRTC signalling server...
      ✓ Socket.IO server initialized
      ✓ Room manager initialized
      ✓ Signalling handlers registered

[4/5] Registering HTTP routes...
      ✓ GET /health
      ... (auth and signalling routes)

[5/5] Starting HTTP server...
      ✓ Listening on http://localhost:8080

═══════════════════════════════════════════════════════════════
  Available Endpoints
═══════════════════════════════════════════════════════════════
```

---

## Architecture Verification Checklist

### Service Ports
| Service | Port | Status |
|---------|------|--------|
| PostgreSQL | 5432 | ✓ Phase 1a |
| Redis | 6379 | ✓ Phase 1a |
| Go Backend | 8080 | ✓ Running |
| Mediasoup SFU | 3000 | ✓ Running |
| WebRTC RTC | 40000-49999 | ✓ Available |

### Architecture Flow
```
┌─────────────────────────────────────────────┐
│          Web Client                         │
│    (Browser with WebRTC + Socket.IO)       │
└─────────────┬───────────────────────────────┘
              │
    ┌─────────┴──────────────┐
    │                        │
    v (Port 8080)            v (Port 3000)
    │                        │
┌────────────────────┐   ┌──────────────────┐
│   Go Backend       │   │  Node.js SFU     │
│                    │   │                  │
│ Phase 1a: Auth     │   │ Mediasoup        │
│ Phase 1b: Signal   │   │ WebRTC Routing   │
│ Phase 1c: Mediasoup│   │ Codec Config     │
│ Integration        │   │                  │
└────────────────────┘   └──────────────────┘
     │         ▲               │         ▲
     └─HTTP───┘               └─HTTP───┘
       REST API                  REST API
     (11 endpoints)           (from client)
```

---

## Testing Plan

### 1. Service Health Checks
```bash
# Check Mediasoup health
curl http://localhost:3000/health

# Expected response:
# {"status":"ok","timestamp":"2025-11-21T...","worker":"ready"}

# Check Go backend health
curl http://localhost:8080/health

# Expected response:
# {"status":"ok","service":"vtp-platform","version":"1.0.0"}
```

### 2. Room Operations (using curl)
```bash
# Create/join room
curl -X POST http://localhost:3000/rooms/room-123/peers \
  -H "Content-Type: application/json" \
  -d '{
    "peerId": "peer-1",
    "userId": "user-123",
    "email": "user@example.com",
    "fullName": "Test User",
    "role": "student",
    "isProducer": true,
    "roomName": "Test Room"
  }'

# Expected response includes:
# - roomId, peerId, rtpCapabilities, peers[]

# Get room info
curl http://localhost:3000/rooms/room-123

# Get all rooms
curl http://localhost:3000/rooms
```

### 3. Transport Operations
```bash
# Create transport
curl -X POST http://localhost:3000/rooms/room-123/transports \
  -H "Content-Type: application/json" \
  -d '{
    "peerId": "peer-1",
    "direction": "send"
  }'

# Expected response includes:
# - transportId, iceParameters, iceCandidates, dtlsParameters

# Connect transport
curl -X POST http://localhost:3000/rooms/room-123/transports/transport-123/connect \
  -H "Content-Type: application/json" \
  -d '{
    "dtlsParameters": {...}
  }'
```

### 4. Producer/Consumer Operations
```bash
# Create producer
curl -X POST http://localhost:3000/rooms/room-123/producers \
  -H "Content-Type: application/json" \
  -d '{
    "peerId": "peer-1",
    "kind": "video",
    "rtpParameters": {...}
  }'

# Create consumer
curl -X POST http://localhost:3000/rooms/room-123/consumers \
  -H "Content-Type: application/json" \
  -d '{
    "peerId": "peer-2",
    "producerId": "producer-123",
    "rtpCapabilities": {...}
  }'
```

### 5. Go Client Integration Tests
```bash
# Run all tests
go test ./pkg/... -v

# Expected output:
# === RUN   TestHealthCheck
# --- PASS: TestHealthCheck (0.00s)
# === RUN   TestGetRooms
# --- PASS: TestGetRooms (0.00s)
# ... (19 tests total)
# PASS: 19/19 tests
```

---

## Environment Variables

### Mediasoup Service (.env)
```dotenv
# Service
MEDIASOUP_PORT=3000
NODE_ENV=development
LOG_LEVEL=info

# Mediasoup Worker
MEDIASOUP_LOG_LEVEL=error
MEDIASOUP_LISTEN_IP=127.0.0.1
MEDIASOUP_ANNOUNCED_IP=127.0.0.1
MEDIASOUP_RTC_MIN_PORT=40000
MEDIASOUP_RTC_MAX_PORT=49999
```

### Go Backend (.env)
```dotenv
# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/vtp_db?sslmode=disable

# Authentication
JWT_SECRET=your-super-secret-key-here
JWT_EXPIRY_HOURS=24
JWT_REFRESH_EXPIRY_HOURS=168

# Server
PORT=8080

# Mediasoup Integration (Phase 1c)
MEDIASOUP_URL=http://localhost:3000
```

---

## Key Integration Points

### 1. Phase 1a ↔ Phase 1b
- Authentication tokens still required for protected endpoints
- User information passed through signalling
- Role-based access control maintained

### 2. Phase 1b ↔ Phase 1c
- Socket.IO events enhanced with Mediasoup callbacks
- RoomManager still functional
- Participant tracking integrated with peer management
- Backward compatible (all Phase 1b features still work)

### 3. Phase 1c Integration Handler
- Mediasoup client initialized on backend startup
- Room operations synchronized
- Transport lifecycle managed
- Producer/consumer negotiation handled
- Automatic cleanup on peer disconnect

### 4. New Mediasoup Go Client
- RESTful HTTP interface to Mediasoup
- 11 public API methods
- Proper error handling
- Type-safe parameter passing
- Request/response marshalling

---

## Deployment Readiness Assessment

### Code Level
- [x] All 19 unit tests passing
- [x] Code compiles without errors
- [x] Error handling complete
- [x] Logging configured
- [x] Documentation comprehensive

### Infrastructure Level
- [x] Ports available (3000, 8080, 40000-49999)
- [x] Database running (Phase 1a)
- [x] Dependencies installed
- [x] Configuration files in place

### Integration Level
- [x] Go client implements all endpoints
- [x] Mediasoup handlers registered
- [x] Socket.IO event handlers working
- [x] Type definitions complete
- [x] Backward compatibility maintained

### Operational Level
- [x] Health check endpoints
- [x] Logging and monitoring
- [x] Error recovery
- [x] Resource cleanup
- [x] Graceful degradation

---

## Monitoring and Debugging

### View Mediasoup Logs
```bash
# Terminal where Mediasoup is running
# Should show:
# - Worker creation
# - Room creation
# - Peer join/leave
# - Transport/producer/consumer lifecycle
```

### View Go Backend Logs
```bash
# Terminal where Go backend is running
# Should show:
# - Connection events
# - Room operations
# - Signalling messages
# - Mediasoup API calls
```

### Test Endpoints

**Health Checks:**
```bash
curl http://localhost:3000/health
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/signalling/health
```

**Room Statistics:**
```bash
curl http://localhost:8080/api/v1/signalling/rooms/stats
```

---

## Troubleshooting

### Mediasoup Service Won't Start
1. Check Node.js version: `node --version` (should be 16+)
2. Check npm dependencies: `npm install` in mediasoup-sfu/
3. Check ports: `netstat -ano | findstr 3000`
4. Check logs for errors

### Go Backend Can't Connect to Mediasoup
1. Verify Mediasoup is running on port 3000
2. Check MEDIASOUP_URL environment variable
3. Check firewall/network connectivity
4. Test endpoint: `curl http://localhost:3000/health`

### RTC Ports Not Available
1. Check if ports 40000-49999 are in use: `netstat -ano | findstr 40000`
2. Check firewall rules
3. Check for other services using UDP ports
4. Modify MEDIASOUP_RTC_MIN/MAX_PORT if needed

### Unit Tests Failing
1. Run with verbose output: `go test ./pkg/... -v`
2. Check Go version: `go version` (should be 1.24+)
3. Check dependencies: `go mod tidy && go mod verify`
4. Individual test: `go test -run TestHealthCheck -v`

---

## Next Steps

### Immediate (Today)
1. ✓ Deploy Mediasoup service
2. ✓ Start Go backend
3. ✓ Verify service health
4. ⏳ Integration testing with cURL
5. ⏳ Multi-peer scenario testing

### Short-term (This Week)
1. Client-side WebRTC implementation
2. Browser compatibility testing
3. End-to-end testing with multiple clients
4. Performance testing under load

### Medium-term (Next Week)
1. Security audit
2. Production deployment
3. Phase 2A: Recording system
4. Monitoring and alerting setup

---

## Performance Expectations

### Response Times (Local Testing)
- Room join: <100ms
- Transport creation: 50-150ms
- Producer creation: 50-150ms
- Consumer creation: 50-150ms
- Health check: <5ms

### Capacity
- Signalling server: 1000+ concurrent connections
- Mediasoup router: 1000+ peers per worker
- Available RTC ports: 10,000 (40000-49999)
- Theoretical max peers: Limited by CPU/memory

### Resource Usage
- Go process: ~50-100MB RAM
- Node.js process: ~150-200MB RAM
- Per connection: ~100-200KB RAM
- Network overhead: Minimal (HTTP/1.1)

---

## Success Criteria

✅ **Phase 1C is successfully deployed when:**

1. Mediasoup service running on port 3000
2. Go backend running on port 8080 with Mediasoup integration
3. Health checks passing for both services
4. All 19 unit tests passing
5. Room creation/deletion working
6. Peer join/leave working
7. Transport creation working
8. Producer/consumer negotiation working
9. Multi-peer scenario passing
10. Error handling and cleanup working

---

## Sign-off

**Deployment Prepared:** November 21, 2025  
**Status:** Ready for Deployment  
**Last Verified:** Pre-deployment checklist completed  
**Documentation:** Complete

**Next Phase:** Phase 1C Integration Testing

---

**For detailed technical information, refer to:**
- `PHASE_1C_README.md` - Overview and setup
- `PHASE_1C_INTEGRATION.md` - Technical integration details
- `PHASE_1C_COMPLETE_SUMMARY.md` - Implementation summary
- `PHASE_1C_VALIDATION_CHECKLIST.md` - QA validation

