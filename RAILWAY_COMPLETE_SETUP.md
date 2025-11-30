# Complete Railway Setup Guide for VTP Platform

## Problem Analysis
Based on your screenshots showing 404 errors and "health.down" status, here's what's missing:

1. ❌ **PostgreSQL database not connected** → Backend can't start properly
2. ❌ **Redis not connected** → Sessions/caching failing
3. ❌ **CORS misconfigured** → Frontend can't communicate with backend
4. ❌ **Environment variables incomplete** → Features disabled
5. ❌ **Frontend not pointing to Railway backend** → API calls going nowhere

## Railway Services Required

You need **THREE** separate services in Railway:

### Service 1: Backend (Go API)
### Service 2: Frontend (Next.js)
### Service 3: PostgreSQL Database
### Service 4: Redis Cache

---

## Step-by-Step Setup

### 1. Create PostgreSQL Database

In Railway dashboard:
1. Click **"+ Create"** → **"Database"** → **"Add PostgreSQL"**
2. Wait for provisioning (~30 seconds)
3. Copy the `DATABASE_URL` from the PostgreSQL service (it auto-generates)

### 2. Create Redis Cache

In Railway dashboard:
1. Click **"+ Create"** → **"Database"** → **"Add Redis"**
2. Wait for provisioning (~30 seconds)
3. Copy the `REDIS_URL` from the Redis service (format: `redis://default:password@host:port`)

### 3. Configure Backend Service (VTP)

Click on your VTP service → **"Variables"** tab → Add ALL these variables:

#### Required Core Variables
```bash
# Application
APP_ENV=production
PORT=8080

# Authentication (CRITICAL - generate a strong secret!)
JWT_SECRET=your-super-secret-jwt-key-minimum-32-characters-long-change-this-now
JWT_EXPIRY_HOURS=24
JWT_REFRESH_EXPIRY_HOURS=168

# Database (from PostgreSQL service - Railway auto-provides this)
DATABASE_URL=${{Postgres.DATABASE_URL}}

# Redis (from Redis service - Railway auto-provides this)
REDIS_URL=${{Redis.REDIS_URL}}

# CORS (CRITICAL - set to your frontend Railway URL)
CORS_ORIGINS=https://vtp-mu.vercel.app,https://vtp-production.up.railway.app
```

#### Optional Email Variables (for password reset)
```bash
# SMTP (example with Gmail - use app-specific password)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-app-specific-password
FROM_EMAIL=noreply@vtp-platform.com
RESET_BASE_URL=https://vtp-mu.vercel.app/reset-password
```

#### Optional Recording Variables
```bash
RECORDINGS_DIR=/app/recordings
FFMPEG_PATH=/usr/bin/ffmpeg
```

### 4. Configure Frontend Service (Vercel/Railway)

If hosting frontend on **Vercel**:
1. Go to Vercel dashboard → Your project → **"Settings"** → **"Environment Variables"**
2. Add:
```bash
NEXT_PUBLIC_API_URL=https://vtp-production.up.railway.app
NEXT_PUBLIC_APP_NAME=VTP
NEXT_PUBLIC_APP_VERSION=0.1.0
```

If hosting frontend on **Railway**:
1. Create new service from GitHub (vtp-frontend folder)
2. Set same environment variables
3. Build command: `npm run build`
4. Start command: `npm start`

### 5. Link Services in Railway

Railway should auto-link PostgreSQL and Redis, but verify:
1. Click on VTP backend service
2. Go to **"Settings"** → **"Service Variables"**
3. You should see `${{Postgres.DATABASE_URL}}` and `${{Redis.REDIS_URL}}` available
4. If not, manually reference them in Variables tab

---

## Critical Configuration Checklist

### ✅ Backend Must Have:
- [ ] `JWT_SECRET` set (NOT the default!)
- [ ] `DATABASE_URL` pointing to Railway Postgres
- [ ] `REDIS_URL` pointing to Railway Redis
- [ ] `CORS_ORIGINS` including your frontend domain
- [ ] `PORT=8080`
- [ ] Dockerfile properly configured (already done ✅)

### ✅ Frontend Must Have:
- [ ] `NEXT_PUBLIC_API_URL` pointing to Railway backend URL
- [ ] CORS origin matches exactly (no trailing slash)

### ✅ Services Running:
- [ ] PostgreSQL: Green checkmark
- [ ] Redis: Green checkmark
- [ ] Backend: Green checkmark + `/health` returns 200
- [ ] Frontend: Green checkmark + can load pages

---

## Testing Your Deployment

### 1. Test Backend Health
```bash
curl https://vtp-production.up.railway.app/health
```

**Expected response:**
```json
{
  "status": "healthy",
  "service": "vtp-platform",
  "version": "1.0.0",
  "instance": "backend-xyz",
  "database": true
}
```

**If you see `"database": false`** → Database not connected

### 2. Test API Endpoints
```bash
# Test registration
curl -X POST https://vtp-production.up.railway.app/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!@#","first_name":"Test","last_name":"User","role":"student"}'

# Test login
curl -X POST https://vtp-production.up.railway.app/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!@#"}'
```

### 3. Check Backend Logs
In Railway:
1. Click VTP service → **"Deployments"** → Latest deployment
2. Click **"View Logs"**
3. Look for:
```
✓ Database connected
✓ Migrations completed
✓ User store
✓ 2FA service (TOTP)
✓ Password reset service (24h expiry)
✓ WebSocket /socket.io/ (WebRTC signalling)
```

**Common errors:**
- `"Database connection failed"` → Check DATABASE_URL
- `"Redis connection refused"` → Check REDIS_URL
- `"CORS error"` → Check CORS_ORIGINS matches frontend
- `"JWT secret not set"` → Set JWT_SECRET

---

## Fixing Your Current Issues

### Issue 1: 404 on `/auth/login`
**Cause:** Frontend calling wrong URL or backend not running
**Fix:**
1. Verify `NEXT_PUBLIC_API_URL` in frontend env = `https://vtp-production.up.railway.app`
2. Rebuild frontend with correct env
3. Clear browser cache

### Issue 2: "status.health.down"
**Cause:** Backend `/health` endpoint returning 503 (database not connected)
**Fix:**
1. Add PostgreSQL service to Railway
2. Set `DATABASE_URL` in backend variables
3. Redeploy backend
4. Check logs for "✓ Database connected"

### Issue 3: "stream.loading.redirect" stuck
**Cause:** WebSocket connection failing or CORS blocking
**Fix:**
1. Ensure `CORS_ORIGINS` includes frontend URL
2. Check browser console for CORS errors
3. Verify `/socket.io/` endpoint accessible

---

## Full Environment Variable Template

Copy this to Railway backend **"Variables"** tab (replace values):

```bash
# === REQUIRED ===
APP_ENV=production
PORT=8080
JWT_SECRET=CHANGE-THIS-TO-RANDOM-STRING-MIN-32-CHARS
JWT_EXPIRY_HOURS=24
JWT_REFRESH_EXPIRY_HOURS=168
DATABASE_URL=${{Postgres.DATABASE_URL}}
REDIS_URL=${{Redis.REDIS_URL}}
CORS_ORIGINS=https://vtp-mu.vercel.app

# === OPTIONAL: Email ===
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-app-password
FROM_EMAIL=noreply@vtp.com
RESET_BASE_URL=https://vtp-mu.vercel.app/reset-password

# === OPTIONAL: Media ===
RECORDINGS_DIR=/app/recordings
FFMPEG_PATH=/usr/bin/ffmpeg

# === OPTIONAL: Logging ===
LOG_LEVEL=info
```

---

## What You Get After Setup

### ✅ Authentication
- User registration with email validation
- Login with JWT tokens (24h access, 7-day refresh)
- 2FA TOTP setup and verification
- Password reset via email
- Change password
- Profile management

### ✅ Courses & Content
- Create/read/update/delete courses
- Manage instructors, subjects, meetings
- Upload study materials
- Create assignments
- Student enrollment

### ✅ Live Streaming
- WebRTC signaling via Socket.IO
- Real-time video streaming
- Recording capture and playback
- HLS/DASH streaming

### ✅ Performance
- Redis caching (fast course/instructor lookups)
- Session management (7-day persistence)
- Rate limiting (prevent abuse)
- Health monitoring

---

## Quick Start Commands

### Generate JWT Secret (run locally):
```bash
# PowerShell
-join ((48..57) + (65..90) + (97..122) | Get-Random -Count 32 | ForEach-Object {[char]$_})
```

### Test Local → Railway Connection:
```bash
# Test from your machine to Railway
curl https://vtp-production.up.railway.app/health

# Expected: {"status":"healthy","database":true}
```

### View Railway Logs:
```bash
railway logs --service vtp
```

---

## Troubleshooting Decision Tree

```
Is deployment GREEN in Railway?
├─ NO → Check build logs for errors
│       Fix: Ensure Dockerfile builds successfully
│
└─ YES → Does /health return 200?
    ├─ NO → Check DATABASE_URL and REDIS_URL
    │       Fix: Add PostgreSQL/Redis services
    │
    └─ YES → Does /api/v1/auth/login work?
        ├─ NO → Check CORS_ORIGINS
        │       Fix: Must match frontend URL exactly
        │
        └─ YES → Does frontend connect?
            ├─ NO → Check NEXT_PUBLIC_API_URL
            │       Fix: Must point to Railway backend
            │
            └─ YES → ✅ Everything working!
```

---

## Next Steps After Deployment

1. **Seed Database** (optional):
   - Create initial admin user
   - Add sample courses
   - Import instructors

2. **Enable Monitoring**:
   - Check `/debug/pprof/` for profiling
   - Set up alerts in Railway
   - Monitor logs for errors

3. **Configure Email**:
   - Set up SMTP provider (Gmail, SendGrid, Mailgun)
   - Test password reset flow
   - Customize email templates

4. **Add Sample Data**:
   - Register test student/instructor accounts
   - Create demo courses
   - Test live streaming

---

## Support Checklist

Before asking for help, verify:
- [ ] PostgreSQL service exists and is running
- [ ] Redis service exists and is running
- [ ] Backend `DATABASE_URL` is set correctly
- [ ] Backend `REDIS_URL` is set correctly
- [ ] Backend `JWT_SECRET` is NOT default value
- [ ] Backend `CORS_ORIGINS` includes frontend URL
- [ ] Frontend `NEXT_PUBLIC_API_URL` points to backend
- [ ] `/health` endpoint returns 200 with `"database": true`
- [ ] Backend logs show "✓ Database connected"
- [ ] No CORS errors in browser console

---

## Summary: What to Add to Railway RIGHT NOW

1. **Add PostgreSQL**: Create → Database → PostgreSQL
2. **Add Redis**: Create → Database → Redis
3. **Set Backend Variables**:
   ```
   JWT_SECRET=<generate-strong-secret>
   DATABASE_URL=${{Postgres.DATABASE_URL}}
   REDIS_URL=${{Redis.REDIS_URL}}
   CORS_ORIGINS=https://vtp-mu.vercel.app
   PORT=8080
   APP_ENV=production
   ```
4. **Redeploy Backend**: Wait for green checkmark
5. **Test**: `curl https://vtp-production.up.railway.app/health`
6. **Update Frontend**: Set `NEXT_PUBLIC_API_URL=https://vtp-production.up.railway.app`
7. **Redeploy Frontend**: Wait for green checkmark
8. **Test**: Login should work!

Your deployment will be fully functional once these 8 steps are complete. 🚀
