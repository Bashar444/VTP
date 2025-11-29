# Mediasoup SFU Deployment Guide

## Quick Start with Docker (Recommended)

### Prerequisites
- Docker (https://www.docker.com/products/docker-desktop)
- Docker Compose (included with Docker Desktop)

### Deploy Full Stack

```bash
# Clone/navigate to VTP project
cd c:\Users\Admin\Desktop\VTP

# Start all services (PostgreSQL, Redis, Mediasoup SFU, Go API)
docker-compose up -d

# Verify services are running
docker-compose ps

# Check logs
docker-compose logs -f mediasoup-sfu
docker-compose logs -f api
```

### Service URLs
- **Mediasoup SFU**: http://localhost:3000
- **Go API (Signalling)**: http://localhost:8080/socket.io/
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379
- **MinIO S3**: http://localhost:9000
- **pgAdmin**: http://localhost:5050

---

## Manual Installation (If Docker Not Available)

### Requirements
- Node.js 16+ (https://nodejs.org/)
- npm 8+ (included with Node.js)

### Installation Steps

1. **Install Dependencies**
```bash
cd mediasoup-sfu
npm install
```

2. **Configure Environment**
```bash
# Edit .env file with your settings
# Default settings should work for local development
```

3. **Start Service**
```bash
# Production mode
npm start

# Development mode (with auto-reload)
npm run dev
```

4. **Verify Service is Running**
```bash
curl http://localhost:3000/health
# Expected response: {"status":"ok","timestamp":"...","worker":"active"}
```

---

## Docker Deployment Details

### Build Mediasoup Image

```bash
# Build from Dockerfile
docker build -t vtp-mediasoup-sfu:latest ./mediasoup-sfu

# Run individual container
docker run -d \
  --name vtp-mediasoup-sfu \
  -p 3000:3000 \
  -p 40000-49999:40000-49999/udp \
  -p 40000-49999:40000-49999/tcp \
  -e MEDIASOUP_ANNOUNCED_IP=127.0.0.1 \
  vtp-mediasoup-sfu:latest
```

### Docker Compose Services

The `docker-compose.yml` includes:

1. **PostgreSQL** - Database (port 5432)
2. **Redis** - Cache & sessions (port 6379)
3. **Mediasoup SFU** - WebRTC media routing (port 3000, RTC ports 40000-49999)
4. **Go API** - Authentication & signalling (port 8080)
5. **MinIO** - S3-compatible storage (port 9000)
6. **pgAdmin** - Database UI (port 5050)

### Common Docker Commands

```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f mediasoup-sfu

# Rebuild service
docker-compose up -d --build mediasoup-sfu

# Check service status
docker-compose ps

# Execute command in container
docker exec vtp-mediasoup-sfu npm list

# Remove everything (careful!)
docker-compose down -v
```

---

## Port Configuration

### Mediasoup Ports
- **HTTP**: 3000 (REST API, health check)
- **RTC**: 40000-49999 (WebRTC media, UDP/TCP)

### Network Ports
- PostgreSQL: 5432
- Redis: 6379
- MinIO: 9000 (S3 API), 9001 (Console)
- pgAdmin: 5050

### Firewall Rules (For Production)
```bash
# Allow Mediasoup RTC ports
ufw allow 3000/tcp
ufw allow 3000/udp
ufw allow 40000:49999/udp
ufw allow 40000:49999/tcp

# Allow Go API
ufw allow 8080/tcp

# Database (internal only)
ufw allow from any to 127.0.0.1 port 5432
ufw allow from any to 127.0.0.1 port 6379
```

---

## Health Checks

### Mediasoup Health Check
```bash
curl http://localhost:3000/health

# Response:
# {
#   "status": "ok",
#   "timestamp": "2025-01-01T00:00:00Z",
#   "worker": "active"
# }
```

### Go API Health Check
```bash
curl http://localhost:8080/health
```

### Check All Services
```bash
docker-compose ps

# Expected output:
# NAME                  STATUS
# vtp-postgres         Up (healthy)
# vtp-redis            Up (healthy)
# vtp-mediasoup-sfu    Up (healthy)
# vtp-api              Up (healthy)
```

---

## Configuration

### Environment Variables (.env)

```env
# Mediasoup Configuration
MEDIASOUP_PORT=3000
MEDIASOUP_LISTEN_IP=127.0.0.1
MEDIASOUP_ANNOUNCED_IP=127.0.0.1
MEDIASOUP_RTC_MIN_PORT=40000
MEDIASOUP_RTC_MAX_PORT=49999

# Logging
LOG_LEVEL=info

# Environment
NODE_ENV=development
```

### For Production Deployment

Update `docker-compose.yml`:
```yaml
environment:
  NODE_ENV: production
  MEDIASOUP_LISTEN_IP: "0.0.0.0"
  MEDIASOUP_ANNOUNCED_IP: "your.domain.com"
  LOG_LEVEL: warn
```

---

## Testing the Deployment

### 1. Check Service Health
```bash
curl http://localhost:3000/health
```

### 2. Get All Rooms
```bash
curl http://localhost:3000/rooms
# Expected: {"rooms":[],"totalRooms":0,"totalPeers":0}
```

### 3. Test Integration with Go Client
```bash
go test ./pkg/mediasoup -v -run TestHealthCheck
```

### 4. Run Full Test Suite
```bash
go test ./pkg/... -v
# Expected: 19/19 tests passing
```

---

## Monitoring and Debugging

### View Real-time Logs
```bash
# Mediasoup logs
docker-compose logs -f mediasoup-sfu

# Go API logs
docker-compose logs -f api

# All services
docker-compose logs -f
```

### Inspect Container
```bash
# Get container ID
docker ps | grep mediasoup-sfu

# Access container shell
docker exec -it vtp-mediasoup-sfu sh

# Check running processes
docker top vtp-mediasoup-sfu
```

### Performance Monitoring
```bash
# CPU and memory usage
docker stats vtp-mediasoup-sfu

# Detailed container info
docker inspect vtp-mediasoup-sfu
```

---

## Troubleshooting

### Service Won't Start

1. **Port Already in Use**
```bash
# Check what's using port 3000
netstat -ano | findstr :3000

# Kill the process (Windows)
taskkill /PID <pid> /F

# Or use different port in .env
MEDIASOUP_PORT=3001
```

2. **Memory Issues**
```bash
# Increase Docker memory
docker update --memory=2g vtp-mediasoup-sfu
```

3. **Network Issues**
```bash
# Verify network connectivity
docker network inspect vtp-network

# Rebuild network
docker-compose down
docker-compose up -d --build
```

### Service Crashes

1. **Check Logs**
```bash
docker-compose logs mediasoup-sfu | tail -50
```

2. **Verify Dependencies**
```bash
docker-compose ps
# All services should be "Up (healthy)"
```

3. **Restart Service**
```bash
docker-compose restart mediasoup-sfu
```

### RTC Port Issues

1. **Firewall Blocking**
```bash
# Test RTC port accessibility
telnet 127.0.0.1 40000

# Or in PowerShell
Test-NetConnection -ComputerName 127.0.0.1 -Port 40000
```

2. **NAT/Router Configuration**
- For production, update `MEDIASOUP_ANNOUNCED_IP` to your public IP
- Configure port forwarding on router: 40000-49999 UDP/TCP

---

## Production Deployment

### Pre-Deployment Checklist
- [ ] Test on local machine
- [ ] Update environment variables for production
- [ ] Configure firewall rules
- [ ] Set up monitoring (Prometheus/Grafana)
- [ ] Configure logging (ELK stack)
- [ ] Set up backups (database, recordings)
- [ ] Configure SSL/TLS for API
- [ ] Update DNS records
- [ ] Load test (see Performance Testing below)

### Deploy to Server

```bash
# SSH into server
ssh user@your-server.com

# Clone repository
git clone https://github.com/your-org/vtp-platform.git
cd vtp-platform

# Create .env file with production settings
cat > .env << EOF
NODE_ENV=production
MEDIASOUP_ANNOUNCED_IP=your-server-ip.com
DATABASE_URL=postgresql://...
REDIS_URL=redis://...
EOF

# Start services
docker-compose up -d

# Verify deployment
curl https://your-server.com/health
```

### Monitoring Stack (Optional)

```bash
# Add to docker-compose.yml for monitoring
# Prometheus (metrics)
# Grafana (dashboards)
# ELK Stack (logging)
```

---

## Performance Testing

### Load Testing with Apache Bench
```bash
# Test health endpoint
ab -n 1000 -c 10 http://localhost:3000/health

# Expected: All requests should succeed with <100ms latency
```

### Concurrent Room Testing
```bash
# Create script to test multiple concurrent rooms
# See phase-1c-load-test.js for implementation
```

---

## Next Steps

1. ✅ **Deploy Mediasoup Service** - DONE
2. ⏳ **Client-side WebRTC Implementation** - Coming next
3. ⏳ **End-to-end Testing** - After client implementation
4. ⏳ **Performance Testing** - Load and stress testing
5. ⏳ **Production Hardening** - Security and monitoring

---

## Support

For issues or questions:
1. Check logs: `docker-compose logs mediasoup-sfu`
2. Review PHASE_1C_INTEGRATION.md for API details
3. Check test files for usage examples
4. Consult Mediasoup documentation: https://mediasoup.org/

---

**Status: Mediasoup Service Ready for Deployment** ✅
