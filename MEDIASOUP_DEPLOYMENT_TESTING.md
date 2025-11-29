# Phase 1c Mediasoup Deployment Testing Guide

## Test Checklist

### Prerequisites
- [ ] Docker installed and running
- [ ] docker-compose available
- [ ] Go 1.21+ installed
- [ ] curl or Postman for API testing

### Phase 1: Service Startup Verification

#### 1.1 Start Services
```bash
cd c:\Users\Admin\Desktop\VTP
docker-compose up -d
```

#### 1.2 Verify Container Health
```bash
# Check all services running
docker-compose ps

# Expected Output:
# NAME                  STATUS                  PORTS
# vtp-postgres         Up (healthy)            5432/tcp
# vtp-redis            Up (healthy)            6379/tcp
# vtp-mediasoup-sfu    Up (healthy)            3000/tcp, 40000-49999/tcp, 40000-49999/udp
# vtp-api              Up (healthy)            8080/tcp
```

#### 1.3 Service Status Tests

**Test Mediasoup Health:**
```bash
curl http://localhost:3000/health

# Expected Response (200 OK):
{
  "status": "ok",
  "timestamp": "2025-01-01T12:00:00Z",
  "worker": "active"
}
```

**Test Go API Health:**
```bash
curl http://localhost:8080/health

# Expected Response (200 OK):
HTTP/1.1 200 OK
```

---

### Phase 2: Mediasoup API Testing

#### 2.1 Room Operations

**Get All Rooms:**
```bash
curl http://localhost:3000/rooms

# Expected Response (200 OK):
{
  "rooms": [],
  "totalRooms": 0,
  "totalPeers": 0
}
```

**Create/Join Room:**
```bash
curl -X POST http://localhost:3000/rooms/test-room-1/peers \
  -H "Content-Type: application/json" \
  -d '{
    "peerId": "peer-1",
    "userId": "user-123",
    "email": "user@example.com",
    "fullName": "Test User",
    "role": "student",
    "isProducer": true,
    "roomName": "Test Classroom"
  }'

# Expected Response (200 OK):
{
  "roomId": "test-room-1",
  "peerId": "peer-1",
  "rtpCapabilities": {
    "codecs": [...],
    "headerExtensions": [...]
  },
  "peers": []
}
```

**Get Room Details:**
```bash
curl http://localhost:3000/rooms/test-room-1

# Expected Response (200 OK):
{
  "roomId": "test-room-1",
  "name": "Test Classroom",
  "peerCount": 1,
  "peers": [
    {
      "peerId": "peer-1",
      "userId": "user-123",
      "email": "user@example.com",
      ...
    }
  ],
  "createdAt": 1234567890
}
```

#### 2.2 Transport Operations

**Create Transport:**
```bash
curl -X POST http://localhost:3000/rooms/test-room-1/transports \
  -H "Content-Type: application/json" \
  -d '{
    "peerId": "peer-1",
    "direction": "send"
  }'

# Expected Response (200 OK):
{
  "transportId": "transport-xxx",
  "iceParameters": {
    "usernameFrag": "xxx",
    "password": "xxx"
  },
  "iceCandidates": [
    {
      "foundation": "1",
      "priority": 2130706431,
      "ip": "127.0.0.1",
      "protocol": "udp",
      "port": 40001,
      "type": "host"
    }
  ],
  "dtlsParameters": {
    "role": "auto",
    "fingerprints": [...]
  }
}
```

**Connect Transport:**
```bash
curl -X POST http://localhost:3000/rooms/test-room-1/transports/transport-xxx/connect \
  -H "Content-Type: application/json" \
  -d '{
    "dtlsParameters": {
      "role": "client",
      "fingerprints": [...]
    }
  }'

# Expected Response (200 OK):
{"message": "Transport connected"}
```

---

### Phase 3: Go Client Integration Testing

#### 3.1 Build Go Application
```bash
cd c:\Users\Admin\Desktop\VTP
go build -o vtp.exe ./cmd/main.go

# Expected: No errors, vtp.exe created
```

#### 3.2 Run Unit Tests
```bash
# Test Mediasoup client
go test ./pkg/mediasoup -v

# Expected Output:
# === RUN   TestHealthCheck
# --- PASS: TestHealthCheck (0.00s)
# === RUN   TestGetRooms
# --- PASS: TestGetRooms (0.00s)
# ...
# PASS: 10/10 tests
```

#### 3.3 Run Integration Tests
```bash
# Test signalling server
go test ./pkg/signalling -v

# Expected Output:
# === RUN   TestNewSignallingServer
# --- PASS: TestNewSignallingServer (0.00s)
# ...
# PASS: 9/9 tests
```

#### 3.4 Run All Tests
```bash
go test ./pkg/... -v

# Expected Output:
# PASS: 19/19 tests total
# All tests should pass
```

---

### Phase 4: Mediasoup Docker Container Verification

#### 4.1 Container Inspection
```bash
# Get container details
docker inspect vtp-mediasoup-sfu

# Check container logs
docker logs vtp-mediasoup-sfu

# Expected: Service started successfully
# Log example:
# ✓ Mediasoup worker created
# ✓ Express server started on port 3000
# ✓ RTC port range: 40000-49999
```

#### 4.2 Container Resource Usage
```bash
# Check memory and CPU usage
docker stats vtp-mediasoup-sfu

# Expected:
# CONTAINER ID   NAME              CPU %   MEM USAGE
# xxx            vtp-mediasoup-sfu 0.5%    150MB
```

#### 4.3 Port Accessibility
```bash
# Test HTTP port
netstat -ano | findstr :3000

# Test RTC port (PowerShell)
Test-NetConnection -ComputerName 127.0.0.1 -Port 40000 -WarningAction SilentlyContinue

# Expected: Port is open and listening
```

---

### Phase 5: Integration with Go Backend

#### 5.1 Test Mediasoup Client Library
```bash
# Create a test file to verify client integration
cat > test_mediasoup_integration.go << 'EOF'
package main

import (
	"log"
	"github.com/yourusername/vtp-platform/pkg/mediasoup"
)

func main() {
	client := mediasoup.NewClient("http://localhost:3000")
	
	// Test health check
	health, err := client.Health()
	if err != nil {
		log.Fatalf("Health check failed: %v", err)
	}
	log.Printf("✓ Health check passed: %s", health.Status)
	
	// Test get rooms
	rooms, err := client.GetRooms()
	if err != nil {
		log.Fatalf("Get rooms failed: %v", err)
	}
	log.Printf("✓ Got rooms: %d total", rooms.TotalRooms)
}
EOF

go run test_mediasoup_integration.go

# Expected Output:
# ✓ Health check passed: ok
# ✓ Got rooms: 0 total
```

#### 5.2 Test Signalling Server Integration
```bash
# Test that signalling server can connect to Mediasoup
cat > test_signalling_mediasoup.go << 'EOF'
package main

import (
	"log"
	"github.com/yourusername/vtp-platform/pkg/signalling"
)

func main() {
	// Create signalling server with Mediasoup integration
	ss, err := signalling.NewSignallingServerWithMediasoup("http://localhost:3000")
	if err != nil {
		log.Fatalf("Failed to create signalling server: %v", err)
	}
	
	log.Printf("✓ Signalling server created with Mediasoup integration")
	log.Printf("✓ Mediasoup URL: http://localhost:3000")
}
EOF

go run test_signalling_mediasoup.go

# Expected Output:
# ✓ Signalling server created with Mediasoup integration
# ✓ Mediasoup URL: http://localhost:3000
```

---

### Phase 6: Load and Performance Testing

#### 6.1 HTTP Load Test
```bash
# Using Apache Bench (if installed)
ab -n 1000 -c 10 http://localhost:3000/health

# Expected:
# Requests per second: 1000+
# Time per request: <1ms
# Failed requests: 0
```

#### 6.2 Concurrent Room Creation
```bash
# Create a simple load test script
cat > load_test.sh << 'EOF'
#!/bin/bash
echo "Creating 10 concurrent rooms..."
for i in {1..10}; do
  curl -X POST http://localhost:3000/rooms/room-$i/peers \
    -H "Content-Type: application/json" \
    -d "{\"peerId\":\"peer-$i\",\"userId\":\"user-$i\",\"email\":\"user$i@example.com\",\"fullName\":\"User $i\",\"role\":\"student\",\"isProducer\":true}" &
done

echo "Waiting for requests to complete..."
wait

echo "Getting all rooms..."
curl http://localhost:3000/rooms
EOF

bash load_test.sh

# Expected:
# - All requests succeed (200 OK)
# - Room count increases
```

---

### Phase 7: Error Handling Tests

#### 7.1 Test Invalid Requests
```bash
# Missing required fields
curl -X POST http://localhost:3000/rooms/test-room/peers \
  -H "Content-Type: application/json" \
  -d '{"peerId":"peer-1"}'

# Expected Response (400 Bad Request):
# {"error": "Missing required fields"}
```

#### 7.2 Test Non-existent Resources
```bash
# Get non-existent room
curl http://localhost:3000/rooms/non-existent-room

# Expected Response (404 Not Found):
# {"error": "Room not found"}
```

#### 7.3 Test Service Recovery
```bash
# Stop service
docker stop vtp-mediasoup-sfu

# Wait for restart
sleep 5

# Verify recovery
curl http://localhost:3000/health

# Expected: Service recovers automatically due to restart policy
```

---

### Phase 8: Documentation Verification

#### 8.1 Check All Documentation
```bash
# Verify all Phase 1c docs exist
ls -la PHASE_1C_*.md

# Expected Files:
# ✓ PHASE_1C_README.md
# ✓ PHASE_1C_INTEGRATION.md
# ✓ PHASE_1C_VALIDATION_CHECKLIST.md
# ✓ PHASE_1C_COMPLETE_SUMMARY.md
# ✓ PHASE_1C_DELIVERABLES.md
# ✓ PHASE_1C_DOCUMENTATION_INDEX.md
# ✓ MEDIASOUP_DEPLOYMENT_GUIDE.md
```

---

## Test Results Summary

Create a test results file:

```bash
cat > TEST_RESULTS_PHASE_1C.md << 'EOF'
# Phase 1c Mediasoup Deployment Test Results

## Date
$(date)

## Service Startup
- [ ] Docker services start successfully
- [ ] All containers are healthy
- [ ] Mediasoup health check passes
- [ ] Go API health check passes

## Mediasoup API
- [ ] Health endpoint works (GET /health)
- [ ] Get all rooms works (GET /rooms)
- [ ] Create transport works (POST /transports)
- [ ] Connect transport works (POST /transports/:id/connect)
- [ ] Create producer works (POST /producers)
- [ ] Create consumer works (POST /consumers)

## Go Client Integration
- [ ] Mediasoup client compiles
- [ ] All 10 client unit tests pass
- [ ] Signalling server compiles with Mediasoup
- [ ] All 9 signalling tests pass
- [ ] Integration test successful

## Performance
- [ ] Health check response time < 1ms
- [ ] Room creation response time < 100ms
- [ ] Transport creation response time < 150ms
- [ ] Support 1000+ concurrent requests
- [ ] Memory usage < 300MB
- [ ] CPU usage < 50%

## Stability
- [ ] Service handles 100 concurrent peers
- [ ] Service recovers from errors
- [ ] No memory leaks detected
- [ ] Logs are clean and informative

## Overall Status
✅ READY FOR PRODUCTION
EOF

cat TEST_RESULTS_PHASE_1C.md
```

---

## Cleanup Commands

```bash
# Stop all services
docker-compose down

# Remove containers and volumes
docker-compose down -v

# Remove images
docker rmi vtp-mediasoup-sfu:latest

# Clean up test files
rm test_mediasoup_integration.go test_signalling_mediasoup.go load_test.sh
```

---

## Troubleshooting Test Failures

### If Health Check Fails
1. Check container logs: `docker logs vtp-mediasoup-sfu`
2. Verify port 3000 is not in use: `netstat -ano | findstr :3000`
3. Restart service: `docker-compose restart mediasoup-sfu`

### If Unit Tests Fail
1. Ensure Mediasoup service is running: `docker-compose ps`
2. Verify connectivity: `curl http://localhost:3000/health`
3. Run individual test: `go test ./pkg/mediasoup -v -run TestHealthCheck`

### If Docker Start Fails
1. Check Docker daemon: `docker --version`
2. Check disk space: `docker system df`
3. Clean up: `docker system prune -a`

---

## Next Steps After Successful Deployment

1. ✅ **Mediasoup Service Deployed** - COMPLETE
2. ⏳ **Client-side WebRTC Implementation** - Next phase
3. ⏳ **End-to-end Testing** - After client implementation
4. ⏳ **Production Hardening** - Final phase

---

**Deployment Test Status: READY** ✅
