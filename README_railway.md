# VTP Backend on Railway - Quick Reference

> **⚠️ IMPORTANT**: For complete setup instructions, troubleshooting, and detailed configuration, see [`RAILWAY_COMPLETE_SETUP.md`](./RAILWAY_COMPLETE_SETUP.md)

This is a quick reference guide. The complete guide includes:
- Step-by-step service setup with screenshots
- Troubleshooting decision tree
- Testing commands and expected outputs
- Common error fixes
- Full environment variable explanations

## Quick Start

### 1. Add Required Services
In Railway dashboard:
1. **Add PostgreSQL**: Click "+ Create" → Database → PostgreSQL
2. **Add Redis**: Click "+ Create" → Database → Redis

### 2. Configure Backend Variables
Click VTP service → Variables → Add these:

```bash
# Required Core
APP_ENV=production
PORT=8080
JWT_SECRET=<generate-32-char-random-string>
DATABASE_URL=${{Postgres.DATABASE_URL}}
REDIS_URL=${{Redis.REDIS_URL}}
CORS_ORIGINS=<your-frontend-url>
```

### 3. Verify Deployment
```bash
curl https://<your-railway-url>/health
# Should return: {"status":"healthy","database":true}
```

## Complete Environment Variables Reference
Set these in Railway Project Settings → Variables:

Authentication & App
- `APP_ENV=production`
- `PORT=8080`
- `JWT_SECRET=<generate-strong-secret>`
- `JWT_EXPIRY_HOURS=24`
- `JWT_REFRESH_EXPIRY_HOURS=168`
- `CORS_ORIGINS=https://your-frontend.example.com`

Database & Cache
- `DATABASE_URL=<Railway Postgres DSN>`
- `REDIS_URL=<Railway Redis URL>`

Email (SMTP)
- `SMTP_HOST=smtp.sendgrid.net` (example)
- `SMTP_PORT=587`
- `SMTP_USER=<smtp-username>`
- `SMTP_PASS=<smtp-password>`
- `FROM_EMAIL=no-reply@your-domain.com`
- `RESET_BASE_URL=https://your-frontend.example.com/reset-password` (frontend route)

Recording / Media (optional)
- `RECORDINGS_DIR=/app/recordings`
- `FFMPEG_PATH=/usr/bin/ffmpeg`

Signaling (optional if mediasoup used)
- `PUBLIC_IP=<public-ip>`

## Start Command
In Railway service settings:
- Build with Dockerfile (recommended). Ensure "Start Command" runs the binary:
  - `./main`

If using Go without Docker:
- Start Command: `go run ./cmd/main.go`

## WebSockets & Routes
- Ensure Railway allows WebSocket upgrades. `/socket.io/` must support Upgrade headers.
- Public routes:
  - `GET /health` (Railway health checks)
  - `GET /debug/pprof/*` (restrict via IP or disable in prod if desired)

## Password Reset Email Flow
- Backend now sends reset emails containing a secure token link:
  - Link format: `${RESET_BASE_URL}?token=<token>`
- Configure SMTP variables above.

## Migrations
- Migrations run automatically on startup.
- Verify logs: "✓ Database connected" then "✓ Migrations completed".

## Quick Test
```powershell
# Test health
curl https://<railway-app-domain>/health

# Register
curl -X POST https://<domain>/api/v1/auth/register -d '{"email":"user@example.com","password":"P@ssw0rd"}' -H "Content-Type: application/json"

# Request password reset (check email)
curl -X POST https://<domain>/api/v1/auth/forgot-password -d '{"email":"user@example.com"}' -H "Content-Type: application/json"
```

## Troubleshooting
- 401/403: Check `JWT_SECRET` and `CORS_ORIGINS`.
- Sessions not persisting: Verify `REDIS_URL` connectivity.
- Reset emails not received: Check SMTP creds and provider logs.
- WebSocket errors: Ensure Railway supports Upgrade and frontend uses correct URL.
- FFmpeg not found: Confirm Dockerfile installs `ffmpeg` and `FFMPEG_PATH=/usr/bin/ffmpeg`.
