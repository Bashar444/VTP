# Vercel Production Fix Guide

## Critical Issue
Vercel Production environment is **missing the API URL**, causing all API calls to fail.

## Fix Steps (5 minutes)

### 1. Add Environment Variable to Vercel

1. Go to: https://vercel.com/bashar444s-projects/vtp/settings/environment-variables
2. Click **"Add New"**
3. Fill in:
   - **Key:** `NEXT_PUBLIC_API_URL`
   - **Value:** `https://vtp-production.up.railway.app`
   - **Environment:** ✅ **Production** (check ONLY Production)
4. Click **"Save"**

### 2. Trigger Redeploy

After adding the variable:
1. Go to: https://vercel.com/bashar444s-projects/vtp
2. Click **"Deployments"** tab
3. Find the latest deployment
4. Click **"..."** → **"Redeploy"**
5. ✅ Check **"Use existing Build Cache"** (faster)
6. Click **"Redeploy"**

**Build time:** ~2-3 minutes

### 3. Verify After Deploy

Once deployment shows "Ready":

```powershell
# Test from PowerShell
$body = @{
    email = "test@example.com"
    password = "Test123!@#"
    full_name = "Test User"
    role = "student"
} | ConvertTo-Json

Invoke-RestMethod -Uri 'https://vtp-production.up.railway.app/api/v1/auth/register' -Method Post -Headers @{'Content-Type'='application/json'} -Body $body
```

Then test from Vercel UI:
1. Open: https://vtp-mu.vercel.app/register
2. Fill form and submit
3. Check browser Console (F12) for errors
4. Check Network tab for API calls

## Alternative: Add Railway CORS Variable

If you haven't added this yet to Railway:

1. Go to Railway Dashboard → VTP Backend → Variables
2. Add:
   ```
   CORS_ORIGINS=http://localhost:3000,https://vtp-mu.vercel.app,https://vtp-git-main-bashar-al-aghas-projects.vercel.app
   ```
3. Railway will auto-redeploy (~2 minutes)

## What's Working Now

✅ Backend: Railway deployed with CORS middleware
✅ Database: PostgreSQL connected
✅ CORS: Allows Vercel frontend requests
✅ API: All endpoints responding correctly

## What's Broken

❌ Vercel: Missing `NEXT_PUBLIC_API_URL` in Production environment
❌ Result: Frontend makes API calls to `http://localhost:8080` (invalid in browser)

## Cost Concerns: Free Alternatives

If Railway costs are too high, we can migrate to **Render.com**:

### Render.com Free Tier
- ✅ Free PostgreSQL (90 days, renewable)
- ✅ Free Web Service (spins down after 15min inactivity)
- ✅ No credit card required
- ✅ Auto-deploy from GitHub
- ⚠️ Cold start: ~30 seconds on first request

### Migration to Render (if needed)
1. Export Railway PostgreSQL data
2. Create Render account
3. Create PostgreSQL service (free)
4. Create Web Service from GitHub repo
5. Add same environment variables
6. Update Vercel `NEXT_PUBLIC_API_URL` to new Render URL

**Let me know if you want to migrate to Render to save costs!**
