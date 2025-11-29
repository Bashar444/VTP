# VTP Platform - Windows Local Deployment Guide

## Status Summary

| Component | Status | Details |
|-----------|--------|---------|
| Go | ✅ Installed | v1.25.3 windows/amd64 |
| Node.js | ✅ Installed | v24.11.1 |
| npm | ✅ Installed | 11.6.2 |
| PostgreSQL | ⚠️ Not in PATH | Needs to be running separately or via Docker |
| Redis | ⚠️ Not checked | Needs to be running separately or via Docker |
| MinIO (S3) | ⚠️ Not checked | Needs to be running separately or via Docker |
| FFmpeg | ❌ Not installed | Needs installation for video recording features |

---

## Updated .env Configuration for Windows

The `.env` file has been updated with Windows paths:

```env
# Database Configuration
DATABASE_URL=postgres://postgres:postgres@localhost:5432/vtp_db?sslmode=disable

# Redis Configuration
REDIS_URL=redis://localhost:6379

# Server Configuration
PORT=8080
NODE_ENV=development

# Mediasoup SFU Configuration
MEDIASOUP_URL=http://localhost:3000
MEDIASOUP_LISTEN_IP=127.0.0.1
MEDIASOUP_ANNOUNCED_IP=127.0.0.1

# JWT Configuration
JWT_SECRET=vtp-super-secret-key-2025-change-in-production
JWT_EXPIRY_HOURS=24
JWT_REFRESH_EXPIRY_HOURS=168

# S3 Configuration (MinIO for local development)
S3_ENDPOINT=http://localhost:9000
S3_REGION=us-east-1
S3_BUCKET=vtp-recordings
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_USE_SSL=false

# Email Configuration (optional)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=your-email@example.com
SMTP_PASSWORD=your-password

# Frontend URL
FRONTEND_URL=http://localhost:3001

# Recording Configuration (WINDOWS PATHS)
RECORDINGS_DIR=C:\Users\basha\Desktop\VTP\recordings
FFMPEG_PATH=ffmpeg
```

### Key Changes:
- **RECORDING_DIR** → **RECORDINGS_DIR** (matches Go code variable name)
- **Old path:** `/app/recordings` → **New path:** `C:\Users\basha\Desktop\VTP\recordings` ✅ Created
- **FFMPEG_PATH:** Set to `ffmpeg` (will use PATH lookup) — **Requires installation**

---

## Prerequisites: Start Database & Infrastructure Services

Before deploying the application, you need PostgreSQL, Redis, and MinIO running. The easiest approach is **Docker Desktop**.

### Option A: Using Docker Desktop (Recommended)

#### 1. Install Docker Desktop
Download from: https://www.docker.com/products/docker-desktop

#### 2. Start Docker Services
```powershell
# Navigate to project root
cd "C:\Users\basha\Desktop\VTP"

# Start all services (PostgreSQL, Redis, MinIO)
docker-compose up -d
```

**Verify services are running:**
```powershell
docker-compose ps
```

Expected output:
```
NAME                COMMAND             STATUS
vtp-db             postgres ...        Up
vtp-redis          redis ...           Up
vtp-minio          minio ...           Up
```

### Option B: Manual Installation (Windows)

#### PostgreSQL
1. Download installer: https://www.postgresql.org/download/windows/
2. Run installer, remember password for `postgres` user
3. Verify connection:
```powershell
psql -U postgres -c "SELECT version();"
```

#### Redis
1. Download from: https://github.com/microsoftarchive/redis/releases
2. Install and run as Windows Service
3. Verify: `redis-cli ping` (should return `PONG`)

#### MinIO
1. Download: https://min.io/download#/windows
2. Create data directory: `mkdir C:\minio-data`
3. Run MinIO:
```powershell
minio.exe server C:\minio-data
```

---

## FFmpeg Installation (Required for Recording)

### Install FFmpeg via Chocolatey (Fastest)

```powershell
# Open PowerShell as Administrator

# Install Chocolatey (if not already installed)
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
Invoke-WebRequest -Uri https://community.chocolatey.org/install.ps1 -UseBasicParsing | Invoke-Expression

# Install FFmpeg
choco install ffmpeg -y

# Verify installation
ffmpeg -version
```

### Alternative: Manual Download

1. Download from: https://ffmpeg.org/download.html#build-windows
2. Extract to `C:\ffmpeg\`
3. Add `C:\ffmpeg\bin` to Windows PATH:
   - Settings → Environment Variables
   - Edit `Path` → Add `C:\ffmpeg\bin`
   - Restart PowerShell

---

## Step-by-Step Deployment Commands

### Terminal 1: Start Mediasoup SFU (WebRTC Server)

```powershell
# Open PowerShell Terminal 1
cd "C:\Users\basha\Desktop\VTP\mediasoup-sfu"

# Install dependencies (first time only)
npm install

# Start the server
npm start
```

**Expected Output:**
```
Mediasoup SFU started on port 3000
Server ready at http://127.0.0.1:3000
```

**Health Check:**
```powershell
Invoke-WebRequest -Uri "http://localhost:3000/health" -UseBasicParsing
```

---

### Terminal 2: Start Go Backend (API & Signalling)

```powershell
# Open PowerShell Terminal 2
cd "C:\Users\basha\Desktop\VTP"

# Build the Go application
go build -o app.exe ./cmd

# Run the application
.\app.exe
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
[3c/5] Initializing recording service...
      ✓ Recording directory: C:\Users\basha\Desktop\VTP\recordings
      ✓ Found ffmpeg on PATH: C:\path\to\ffmpeg.exe
      ✓ Local storage backend initialized
      ✓ Storage manager initialized
      ✓ Streaming manager initialized
      ✓ Playback handlers registered
[4/5] Initializing frontend static files...
      ✓ Frontend assets served
[5/5] Registering HTTP routes...
      ✓ All handlers registered

Server running on http://localhost:8080
```

**Health Check:**
```powershell
Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing
```

---

### Terminal 3: Run Health Checks & Integration Tests

```powershell
# Open PowerShell Terminal 3 (wait 5 seconds for both servers to fully start)

# 1. Check Mediasoup SFU Health
Write-Host "Checking Mediasoup SFU..." -ForegroundColor Green
Invoke-WebRequest -Uri "http://localhost:3000/health" -UseBasicParsing -ErrorAction SilentlyContinue | Select-Object StatusCode, StatusDescription

# 2. Check Go Backend Health
Write-Host "Checking Go Backend API..." -ForegroundColor Green
Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing -ErrorAction SilentlyContinue | Select-Object StatusCode, StatusDescription

# 3. Check Database Connection
Write-Host "Checking Database..." -ForegroundColor Green
$env:DATABASE_URL = "postgres://postgres:postgres@localhost:5432/vtp_db?sslmode=disable"
# If PostgreSQL tools are installed:
# psql -U postgres -d vtp_db -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public';"

# 4. Check Redis Connection
Write-Host "Checking Redis..." -ForegroundColor Green
# redis-cli ping  # If redis-cli is available

# 5. Check MinIO Connection
Write-Host "Checking MinIO..." -ForegroundColor Green
Invoke-WebRequest -Uri "http://localhost:9000/minio/health/live" -UseBasicParsing -ErrorAction SilentlyContinue | Select-Object StatusCode

# 6. List Recording Directory
Write-Host "Checking Recording Directory..." -ForegroundColor Green
Get-Item -Path "C:\Users\basha\Desktop\VTP\recordings" | Format-List

# 7. Verify Environment Variables
Write-Host "Environment Configuration:" -ForegroundColor Yellow
$env:RECORDINGS_DIR
$env:FFMPEG_PATH
```

---

## Complete Deployment Script (All-in-One)

Create a file named `deploy.ps1`:

```powershell
# VTP Platform - Windows Deployment Script

$VtpRoot = "C:\Users\basha\Desktop\VTP"
$RecordingsDir = Join-Path $VtpRoot "recordings"

# Colors for output
$GreenCheck = "✅"
$RedX = "❌"
$Yellow = "⚠️"

Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "  VTP Platform - Windows Deployment Script" -ForegroundColor Cyan
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan

# 1. Verify Prerequisites
Write-Host "`n[1] Checking Prerequisites..." -ForegroundColor Yellow

$prerequisites = @{
    "Go" = "go.exe"
    "Node.js" = "node.exe"
    "npm" = "npm.cmd"
}

foreach ($tool in $prerequisites.GetEnumerator()) {
    try {
        $version = & $tool.Value --version 2>&1
        Write-Host "$($GreenCheck) $($tool.Key): $($version[0])" -ForegroundColor Green
    } catch {
        Write-Host "$($RedX) $($tool.Key): Not found" -ForegroundColor Red
    }
}

# 2. Create Recordings Directory
Write-Host "`n[2] Setting up directories..." -ForegroundColor Yellow
if (!(Test-Path $RecordingsDir)) {
    New-Item -ItemType Directory -Path $RecordingsDir -Force | Out-Null
}
Write-Host "$($GreenCheck) Recordings directory: $RecordingsDir" -ForegroundColor Green

# 3. Load Environment Variables
Write-Host "`n[3] Loading environment configuration..." -ForegroundColor Yellow
cd $VtpRoot
$env:DATABASE_URL = "postgres://postgres:postgres@localhost:5432/vtp_db?sslmode=disable"
$env:REDIS_URL = "redis://localhost:6379"
$env:RECORDINGS_DIR = $RecordingsDir
$env:FFMPEG_PATH = "ffmpeg"
Write-Host "$($GreenCheck) Environment variables loaded" -ForegroundColor Green

# 4. Display Next Steps
Write-Host "`n[4] Ready for deployment!" -ForegroundColor Cyan
Write-Host "`nOpen three separate PowerShell terminals and run:" -ForegroundColor Yellow
Write-Host "`nTerminal 1 - Mediasoup SFU:" -ForegroundColor Magenta
Write-Host "  cd '$VtpRoot\mediasoup-sfu'" -ForegroundColor Gray
Write-Host "  npm install" -ForegroundColor Gray
Write-Host "  npm start" -ForegroundColor Gray

Write-Host "`nTerminal 2 - Go Backend:" -ForegroundColor Magenta
Write-Host "  cd '$VtpRoot'" -ForegroundColor Gray
Write-Host "  go build -o app.exe ./cmd" -ForegroundColor Gray
Write-Host "  .\app.exe" -ForegroundColor Gray

Write-Host "`nTerminal 3 - Health Checks:" -ForegroundColor Magenta
Write-Host "  curl http://localhost:3000/health" -ForegroundColor Gray
Write-Host "  curl http://localhost:8080/health" -ForegroundColor Gray

Write-Host "`n═══════════════════════════════════════════════════════════════" -ForegroundColor Cyan
```

**Run the script:**
```powershell
cd "C:\Users\basha\Desktop\VTP"
powershell -ExecutionPolicy Bypass -File deploy.ps1
```

---

## Troubleshooting

### Issue: FFmpeg not found
**Solution:** Install FFmpeg (see [FFmpeg Installation](#ffmpeg-installation-required-for-recording) section)

### Issue: Port 3000 or 8080 already in use
**Solution:**
```powershell
# Find what's using the port (e.g., 3000)
Get-NetTCPConnection -LocalPort 3000 -ErrorAction SilentlyContinue | 
  Format-Table LocalPort, OwningProcess -AutoSize

# Kill the process (replace PID with the OwningProcess value)
Stop-Process -Id <PID> -Force
```

### Issue: PostgreSQL connection refused
**Solution:** Ensure PostgreSQL is running
```powershell
# If using Docker:
docker-compose up -d vtp-db

# Check logs:
docker-compose logs vtp-db
```

### Issue: Redis connection error
**Solution:** Start Redis
```powershell
# If using Docker:
docker-compose up -d vtp-redis

# Or manually on Windows:
redis-server.exe
```

### Issue: "Cannot find psql" when running database checks
**Solution:** Install PostgreSQL CLI tools or use Docker

### Issue: Go build fails
**Solution:**
```powershell
# Clean and rebuild
go clean -cache
go clean -modcache
go mod tidy
go build -o app.exe ./cmd
```

---

## Directory Structure for Reference

```
C:\Users\basha\Desktop\VTP\
├── .env                          # Configuration file ✅ Updated
├── cmd/
│   └── main.go                   # Go backend entry point
├── mediasoup-sfu/
│   ├── package.json
│   ├── src/
│   │   └── index.js              # Mediasoup SFU entry point
│   └── node_modules/
├── recordings/                   # ✅ Created - stores video files
├── pkg/
│   ├── auth/                     # Authentication logic
│   ├── db/                       # Database operations
│   ├── recording/                # Recording service
│   ├── signalling/               # WebRTC signalling
│   └── streaming/                # Streaming management
└── frontend/                      # Frontend assets
```

---

## What to Verify After Startup

✅ **Mediasoup SFU (Port 3000)**
- Logs show "Server ready"
- Responds to health checks

✅ **Go Backend (Port 8080)**
- All initialization steps complete without errors
- Database migrations ran successfully
- Recording directory created
- FFmpeg path resolved

✅ **Infrastructure Services**
- PostgreSQL accessible at `localhost:5432`
- Redis accessible at `localhost:6379`
- MinIO accessible at `localhost:9000` (dashboard at `localhost:9001`)

✅ **Recording Capability**
- `C:\Users\basha\Desktop\VTP\recordings` directory exists and is writable
- FFmpeg is accessible via PATH or `FFMPEG_PATH`

---

## Next Steps After Deployment

1. **Frontend Development**: Navigate to `frontend/` or `vtp-frontend/`
2. **Run Integration Tests**: Create test endpoints to verify all services work together
3. **Create Test Users**: Use the auth API to create test accounts
4. **Test WebRTC Session**: Verify peer connections work through Mediasoup
5. **Test Recording**: Start a session and verify recordings are created in `recordings/`

---

## Production Deployment Considerations

🚨 **Before going to production:**
- Change `JWT_SECRET` to a strong random value
- Update database credentials
- Configure proper MinIO buckets and access keys
- Set `NODE_ENV=production`
- Use environment-specific `.env.production` file
- Enable HTTPS/TLS
- Configure proper CORS origins
- Set up log aggregation and monitoring
- Use container orchestration (Docker Swarm / Kubernetes)

