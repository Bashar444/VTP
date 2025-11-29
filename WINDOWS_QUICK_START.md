# VTP Platform - Windows Quick Start (3-Terminal Setup)

## Prerequisites Ready ✅
- Go v1.25.3 installed
- Node.js v24.11.1 installed
- npm 11.6.2 installed
- FFmpeg status: ❌ NOT INSTALLED (see instructions below)
- Recordings directory: ✅ Created at `C:\Users\basha\Desktop\VTP\recordings`
- .env file: ✅ Updated with Windows paths

---

## 🚨 IMPORTANT: Install FFmpeg First

Run this in PowerShell (as Administrator):

```powershell
# Option 1: Via Chocolatey (Fastest)
choco install ffmpeg -y
ffmpeg -version

# Option 2: Manual - Download from https://ffmpeg.org/download.html
# Extract to C:\ffmpeg\, add C:\ffmpeg\bin to Windows PATH, restart PowerShell
```

**Verify FFmpeg is installed:**
```powershell
ffmpeg -version
```

---

## Ensure Database Services are Running

**Via Docker (Recommended):**
```powershell
cd "C:\Users\basha\Desktop\VTP"
docker-compose up -d
docker-compose ps
```

**Or manually ensure:**
- PostgreSQL running on `localhost:5432`
- Redis running on `localhost:6379`  
- MinIO running on `localhost:9000`

---

## Deployment: Open 3 PowerShell Terminals

### 💻 Terminal 1: Mediasoup SFU (WebRTC Server)

```powershell
cd "C:\Users\basha\Desktop\VTP\mediasoup-sfu"
npm install
npm start
```

Wait for: `Server ready at http://127.0.0.1:3000`

---

### 💻 Terminal 2: Go Backend (API Server)

```powershell
cd "C:\Users\basha\Desktop\VTP"
go build -o app.exe ./cmd
.\app.exe
```

Wait for all initialization steps to complete, should see:
```
✓ Database connected
✓ Migrations completed
✓ Recording directory: C:\Users\basha\Desktop\VTP\recordings
✓ Found ffmpeg on PATH: ...
Server running on http://localhost:8080
```

---

### 💻 Terminal 3: Verification & Health Checks

Wait 5 seconds for servers to fully start, then run:

```powershell
# Quick health check
Write-Host "Checking services..."
$mediasoup = Invoke-WebRequest -Uri "http://localhost:3000/health" -UseBasicParsing -ErrorAction SilentlyContinue
$backend = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing -ErrorAction SilentlyContinue

if ($mediasoup.StatusCode -eq 200) { Write-Host "✅ Mediasoup SFU: Running on port 3000" } else { Write-Host "❌ Mediasoup SFU: Not responding" }
if ($backend.StatusCode -eq 200) { Write-Host "✅ Go Backend: Running on port 8080" } else { Write-Host "❌ Go Backend: Not responding" }

Write-Host "`n✅ Deployment Complete!"
Write-Host "`nAccess the platform:"
Write-Host "  - API: http://localhost:8080"
Write-Host "  - Mediasoup SFU: http://localhost:3000"
Write-Host "  - Frontend: http://localhost:3001 (if running)"
```

---

## Current Environment Configuration

| Setting | Value | Status |
|---------|-------|--------|
| Project Root | `C:\Users\basha\Desktop\VTP` | ✅ Ready |
| Recordings Dir | `C:\Users\basha\Desktop\VTP\recordings` | ✅ Created |
| RECORDINGS_DIR (env) | Updated in .env | ✅ Ready |
| FFMPEG_PATH (env) | `ffmpeg` (use PATH lookup) | ⚠️ Requires installation |
| DATABASE_URL | `postgres://postgres:postgres@localhost:5432/vtp_db` | ⚠️ Requires DB running |
| MEDIASOUP_URL | `http://localhost:3000` | ⏳ Ready after Terminal 1 starts |
| JWT_SECRET | `vtp-super-secret-key-2025-change-in-production` | ⚠️ Change in production |

---

## Troubleshooting Quick Reference

**Port already in use?**
```powershell
# Check what's using port 3000 or 8080
Get-NetTCPConnection -LocalPort 3000 -ErrorAction SilentlyContinue | Select-Object LocalPort, OwningProcess
# Kill process if needed:
Stop-Process -Id <PID> -Force
```

**Database won't connect?**
```powershell
# Start with Docker
docker-compose up -d vtp-db
docker-compose logs vtp-db
```

**Go build fails?**
```powershell
go clean -cache
go mod tidy
go build -o app.exe ./cmd
```

**FFmpeg not found?**
```powershell
# Install via Chocolatey (easiest)
choco install ffmpeg -y

# Or manually: Download from ffmpeg.org and add to PATH
```

---

## Summary

| Component | Port | Status | Action |
|-----------|------|--------|--------|
| Mediasoup SFU | 3000 | Ready | Terminal 1: `npm start` |
| Go Backend | 8080 | Ready | Terminal 2: `go run ./cmd` |
| PostgreSQL | 5432 | Needs setup | `docker-compose up -d` or manual install |
| Redis | 6379 | Needs setup | `docker-compose up -d` or manual install |
| MinIO | 9000 | Needs setup | `docker-compose up -d` or manual install |
| **FFmpeg** | N/A | ⚠️ NOT INSTALLED | `choco install ffmpeg -y` |

---

## Next Steps

1. **Install FFmpeg** (run as Administrator):
   ```powershell
   choco install ffmpeg -y
   ```

2. **Start infrastructure services**:
   ```powershell
   cd "C:\Users\basha\Desktop\VTP"
   docker-compose up -d
   ```

3. **Open 3 terminals** and follow the deployment commands above

4. **Verify all services** are responding (Terminal 3)

5. **Check the detailed guide** at `WINDOWS_DEPLOYMENT_GUIDE.md` for more info

