# DigitalOcean Deployment Guide

## Overview

VTP Platform runs on DigitalOcean with the following architecture:

```
┌─────────────────────────────────────────────────────────────────┐
│                    VERCEL (FREE)                                │
│                    vtp-mu.vercel.app                            │
│                    Next.js Frontend                             │
└─────────────────────┬───────────────────────────────────────────┘
                      │ API Calls
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│              DIGITALOCEAN APP PLATFORM                          │
│  ┌─────────────────┐     ┌─────────────────┐                   │
│  │   Go API        │     │   Signaling     │                   │
│  │   $5/month      │     │   $5/month      │                   │
│  └────────┬────────┘     └────────┬────────┘                   │
│           │                       │                             │
│           ▼                       ▼                             │
│  ┌─────────────────┐     ┌─────────────────┐                   │
│  │   PostgreSQL    │     │     Redis       │                   │
│  │   $15/month     │     │   $15/month     │                   │
│  └─────────────────┘     └─────────────────┘                   │
└─────────────────────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────────┐
│              DIGITALOCEAN DROPLET                               │
│  ┌─────────────────┐                                           │
│  │  MediaSoup SFU  │  ← WebRTC Video Streaming                 │
│  │   $12/month     │                                           │
│  └─────────────────┘                                           │
└─────────────────────────────────────────────────────────────────┘

Total: ~$52/month (GitHub Education: $100 credit = ~2 months free)
```

## Quick Start

### 1. Deploy App Platform

```bash
# Install DigitalOcean CLI
brew install doctl  # macOS
# or download from https://github.com/digitalocean/doctl/releases

# Authenticate
doctl auth init

# Create the app
doctl apps create --spec deployment/digitalocean/app.yaml
```

Or use the DigitalOcean Dashboard:
1. Go to https://cloud.digitalocean.com/apps
2. Click "Create App"
3. Connect GitHub: Bashar444/VTP
4. Upload `app.yaml` or configure manually

### 2. Set Environment Secrets

In App Platform → Your App → Settings → App-Level Environment Variables:

```
JWT_SECRET = (generate with: openssl rand -base64 32)
```

### 3. Deploy MediaSoup Droplet

```bash
# Create droplet via dashboard or CLI
doctl compute droplet create vtp-mediasoup \
  --image ubuntu-22-04-x64 \
  --size s-1vcpu-2gb \
  --region nyc1

# SSH into droplet
ssh root@YOUR_DROPLET_IP

# Run setup script
curl -sSL https://raw.githubusercontent.com/Bashar444/VTP/main/deployment/digitalocean/setup-mediasoup.sh | bash
```

### 4. Run Database Migrations

```bash
# Get connection string from App Platform dashboard
# Database → Connection Details → Connection String

# Connect and run migrations
psql "YOUR_CONNECTION_STRING" -f migrations/001_initial_schema.sql
psql "YOUR_CONNECTION_STRING" -f migrations/002_recordings_schema.sql
# ... run all migration files in order
```

### 5. Update Vercel Frontend

In Vercel Dashboard → Project Settings → Environment Variables:

```
NEXT_PUBLIC_API_URL = https://your-app.ondigitalocean.app
NEXT_PUBLIC_WS_URL = wss://your-app.ondigitalocean.app
NEXT_PUBLIC_MEDIASOUP_URL = wss://YOUR_DROPLET_IP:3000
NEXT_PUBLIC_JITSI_SERVER = https://meet.jit.si
```

## Cost Breakdown

| Service | Size | Monthly Cost |
|---------|------|--------------|
| App Platform - API | basic-xxs | $5 |
| App Platform - Signaling | basic-xxs | $5 |
| Managed PostgreSQL | db-s-1vcpu-1gb | $15 |
| Managed Redis | db-s-1vcpu-1gb | $15 |
| Droplet - MediaSoup | s-1vcpu-2gb | $12 |
| **Total** | | **$52/month** |

With $100 GitHub Education credit: **~2 months free**

## Scaling (Future)

When you need more capacity:

```yaml
# In app.yaml, change:
instance_count: 1  →  instance_count: 3
instance_size_slug: basic-xxs  →  basic-xs  # $10/month per instance
```

## Troubleshooting

### API not responding
```bash
doctl apps logs YOUR_APP_ID --type=run
```

### MediaSoup not connecting
```bash
ssh root@YOUR_DROPLET_IP
journalctl -u mediasoup -f
```

### Database connection issues
```bash
# Check if migrations ran
psql "YOUR_CONNECTION_STRING" -c "\dt"
```
