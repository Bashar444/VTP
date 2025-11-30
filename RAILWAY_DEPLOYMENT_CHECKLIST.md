# Railway Deployment Checklist ✅

Use this checklist to ensure everything is configured correctly.

## Pre-Deployment

- [ ] Backend code committed to GitHub
- [ ] Dockerfile exists and builds successfully locally
- [ ] `railway.toml` committed to repository
- [ ] Frontend deployed separately (Vercel recommended)

## Railway Services Setup

### PostgreSQL Database
- [ ] PostgreSQL service created in Railway
- [ ] Status shows green checkmark
- [ ] Can see `DATABASE_URL` in service variables
- [ ] Database has > 100MB storage allocated

### Redis Cache
- [ ] Redis service created in Railway
- [ ] Status shows green checkmark  
- [ ] Can see `REDIS_URL` in service variables
- [ ] Redis memory limit set (256MB minimum)

### Backend Service (VTP)
- [ ] Service created from GitHub repository
- [ ] Build completes successfully
- [ ] Deployment shows green checkmark
- [ ] Port 8080 is exposed
- [ ] Healthcheck endpoint configured

## Environment Variables (Backend)

### Critical Variables (MUST SET)
- [ ] `APP_ENV=production`
- [ ] `PORT=8080`
- [ ] `JWT_SECRET=<32+ character random string>` ⚠️ NOT default!
- [ ] `DATABASE_URL=${{Postgres.DATABASE_URL}}`
- [ ] `REDIS_URL=${{Redis.REDIS_URL}}`
- [ ] `CORS_ORIGINS=<your-frontend-url>` (exact match, no trailing slash)

### Optional but Recommended
- [ ] `JWT_EXPIRY_HOURS=24`
- [ ] `JWT_REFRESH_EXPIRY_HOURS=168`
- [ ] `LOG_LEVEL=info`

### Email Configuration (if using password reset)
- [ ] `SMTP_HOST=<smtp-server>`
- [ ] `SMTP_PORT=587`
- [ ] `SMTP_USER=<username>`
- [ ] `SMTP_PASS=<password>`
- [ ] `FROM_EMAIL=<sender-email>`
- [ ] `RESET_BASE_URL=<frontend-url>/reset-password`

### Media Configuration (if using recordings)
- [ ] `RECORDINGS_DIR=/app/recordings`
- [ ] `FFMPEG_PATH=/usr/bin/ffmpeg`

## Environment Variables (Frontend)

- [ ] `NEXT_PUBLIC_API_URL=<railway-backend-url>`
- [ ] `NEXT_PUBLIC_APP_NAME=VTP`
- [ ] No trailing slash in API_URL

## Testing & Verification

### Health Check
```bash
curl https://<your-railway-url>/health
```
- [ ] Returns HTTP 200
- [ ] Response includes `"status":"healthy"`
- [ ] Response includes `"database":true`
- [ ] Response includes instance ID

### Authentication Endpoints
```bash
# Test registration
curl -X POST https://<your-url>/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"Test123!@#","first_name":"Test","last_name":"User","role":"student"}'
```
- [ ] Returns HTTP 200 or 201
- [ ] Returns access_token and refresh_token
- [ ] No CORS errors in response

```bash
# Test login
curl -X POST https://<your-url>/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"Test123!@#"}'
```
- [ ] Returns HTTP 200
- [ ] Returns valid JWT tokens
- [ ] Tokens can be decoded

### Database Connectivity
- [ ] Backend logs show "✓ Database connected"
- [ ] Backend logs show "✓ Migrations completed"
- [ ] No "database connection failed" errors in logs

### Redis Connectivity
- [ ] No "Redis connection refused" in logs
- [ ] Sessions persist across requests
- [ ] Rate limiting works

### CORS Configuration
- [ ] Frontend can make API calls
- [ ] No CORS errors in browser console
- [ ] Preflight OPTIONS requests succeed
- [ ] Credentials (cookies/tokens) work

### WebSocket/Socket.IO
- [ ] `/socket.io/` endpoint accessible
- [ ] WebSocket upgrade succeeds
- [ ] Real-time features work (if applicable)

## Frontend Integration

- [ ] Frontend loads without errors
- [ ] Login page accessible
- [ ] Registration page accessible
- [ ] Can register new user
- [ ] Can login with credentials
- [ ] Dashboard loads after login
- [ ] Protected routes require auth
- [ ] Logout works correctly

## Logs Verification

Check Railway deployment logs for:
- [ ] "VTP Platform - Educational Live Video Streaming System"
- [ ] "[1/5] Initializing database connection..."
- [ ] "✓ Database connected"
- [ ] "[2/5] Running database migrations..."
- [ ] "✓ Migrations completed"
- [ ] "[3/5] Initializing authentication services..."
- [ ] "✓ User store"
- [ ] "✓ 2FA service (TOTP)"
- [ ] "✓ Password reset service (24h expiry)"
- [ ] "✓ WebSocket /socket.io/"
- [ ] "✓ GET /health"
- [ ] No error messages or panics

## Security Checklist

- [ ] `JWT_SECRET` is NOT the default value
- [ ] `JWT_SECRET` is at least 32 characters
- [ ] `JWT_SECRET` is random and unguessable
- [ ] HTTPS enabled (Railway provides this automatically)
- [ ] CORS restricted to specific origins (not `*`)
- [ ] Database password is strong
- [ ] Redis password is set (Railway default)
- [ ] No secrets committed to Git

## Performance Checklist

- [ ] Backend starts in < 30 seconds
- [ ] Health check responds in < 5 seconds
- [ ] Login request completes in < 2 seconds
- [ ] No memory leaks in logs
- [ ] CPU usage stable
- [ ] Database queries optimized

## Monitoring & Alerts

- [ ] Railway health checks enabled
- [ ] Email alerts configured for downtime
- [ ] Resource usage monitored
- [ ] Error logs reviewed regularly

## Post-Deployment

- [ ] Test all critical user flows
- [ ] Verify email delivery (if configured)
- [ ] Test 2FA setup and verification
- [ ] Test password reset flow
- [ ] Create sample data (courses, users)
- [ ] Document any configuration changes
- [ ] Update team with deployment URL
- [ ] Set up monitoring dashboards

## Common Issues Checklist

If something doesn't work, check:
- [ ] All services are running (green checkmarks)
- [ ] Environment variables are set correctly
- [ ] No typos in variable names
- [ ] CORS_ORIGINS matches exactly (case-sensitive)
- [ ] Frontend API_URL has no trailing slash
- [ ] JWT_SECRET is set and not default
- [ ] DATABASE_URL is accessible
- [ ] REDIS_URL is accessible
- [ ] No CORS errors in browser console
- [ ] Backend logs show no errors

## Support Resources

- **Complete Setup Guide**: [`RAILWAY_COMPLETE_SETUP.md`](./RAILWAY_COMPLETE_SETUP.md)
- **Quick Reference**: [`README_railway.md`](./README_railway.md)
- **Deployment Guide**: [`DEPLOYMENT_SUMMARY.md`](./DEPLOYMENT_SUMMARY.md)
- **Caching Strategy**: [`CACHING_STRATEGY.md`](./CACHING_STRATEGY.md)
- **Performance**: [`PERFORMANCE_PROFILING.md`](./PERFORMANCE_PROFILING.md)
- **Scaling**: [`HORIZONTAL_SCALING_GUIDE.md`](./HORIZONTAL_SCALING_GUIDE.md)

---

## Quick Status Check

Run this command to verify everything:

```bash
# Check backend health
curl https://<your-url>/health

# Expected output:
# {
#   "status": "healthy",
#   "service": "vtp-platform",
#   "version": "1.0.0",
#   "instance": "<instance-id>",
#   "database": true
# }
```

✅ = Everything working  
⚠️ = Need attention  
❌ = Critical issue
