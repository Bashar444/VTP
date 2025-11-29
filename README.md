# VTP Platform - Arabic Educational Platform with Live Video

A comprehensive, scalable educational platform built for Arabic-speaking users (Syrian focus) with live video streaming, recording, chat, and assignment management.

## Project Status: Phase 0 Complete ✓

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Browser / Client                         │
└────────────────────┬────────────────────────────────────────┘
                     │ WebRTC + Socket.IO
┌────────────────────▼────────────────────────────────────────┐
│         Go API Gateway & Signalling Server                   │
│              (port 8080, gorilla/websocket)                  │
└────────────────────┬────────────────────────────────────────┘
                     │ gRPC / HTTP
┌────────────────────▼────────────────────────────────────────┐
│  Mediasoup SFU (Node.js, port 3000) - RTP Media Router     │
│  - Handle WebRTC transports & producers/consumers          │
│  - Pipe streams to recorder                                 │
└────────────────────┬────────────────────────────────────────┘
                     │
    ┌────────────────┴────────────────┬──────────────┐
    │                                 │              │
┌───▼──────────┐          ┌───────────▼──────┐   ┌──▼─────┐
│  PostgreSQL  │          │  Redis Cache     │   │ MinIO  │
│  (users,     │          │  (sessions,      │   │ (S3    │
│  courses,    │          │   presence)      │   │ storage│
│  sessions)   │          │                  │   │ local) │
└──────────────┘          └──────────────────┘   └────────┘
```

## Tech Stack

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| **Backend** | Go 1.21 + Gorilla WebSocket | Fast, stateless, concurrency, low latency |
| **SFU** | Mediasoup (Node.js) | Best-in-class WebRTC SFU, excellent docs |
| **Database** | PostgreSQL 15 | ACID compliance, scales to 100k+ users |
| **Cache** | Redis | Session state, presence, real-time chat |
| **Storage** | MinIO (S3-compatible) | Local development, easy cloud migration |
| **Frontend** | React/Next.js (coming next) | RTL support, SSR, component-driven |
| **Deployment** | Docker Compose (MVP) → K8s | Rapid iteration now, scalability later |

## Project Structure

```
vtp-platform/
├── cmd/
│   └── main.go                      # Application entry point
├── pkg/
│   ├── auth/                        # JWT, user registration/login (TODO)
│   ├── signalling/                  # WebSocket signalling for WebRTC (TODO)
│   ├── db/                          # Database connection & migration runner
│   ├── recorder/                    # FFmpeg integration (TODO)
│   ├── models/                      # Data models (User, Course, Lesson, etc.)
│   └── ...
├── mediasoup/                       # Node.js SFU server
│   ├── index.js                     # Mediasoup initialization & routing
│   └── package.json
├── migrations/
│   └── 001_initial_schema.sql       # PostgreSQL schema
├── frontend/                        # React/Next.js (coming in Phase 2b)
├── docker-compose.yml               # Local development stack
├── Dockerfile                       # Go binary build
├── Makefile                         # Development commands
├── .env.example                     # Configuration template
├── go.mod / go.sum                  # Go dependencies
└── README.md
```

## Quick Start

### Prerequisites
- Docker & Docker Compose (recommended)
- Go 1.21+ (for local development)
- Node.js 18+ (for Mediasoup)

### Start All Services (Docker Compose)

```bash
# Clone/navigate to project
cd /path/to/vtp-platform

# Start services
make docker-up

# Expected output:
# ✓ Services started:
#   - API: http://localhost:8080
#   - Mediasoup: http://localhost:3000
#   - PostgreSQL: localhost:5432
#   - Redis: localhost:6379
#   - MinIO Console: http://localhost:9001
#   - pgAdmin: http://localhost:5050
```

### Verify Setup

```bash
# Check API health
curl http://localhost:8080/health
# Response: {"status":"ok","service":"vtp-platform"}

# Check Mediasoup health
curl http://localhost:3000/health
# Response: {"status":"ok","service":"mediasoup-sfu"}

# Access pgAdmin (manage PostgreSQL)
# URL: http://localhost:5050
# Email: admin@example.com
# Password: admin
```

### Stop Services

```bash
make docker-down
```

## Database Schema

The schema supports the full data model for the MVP:

- **users** - Student, teacher, admin profiles
- **courses** - Course metadata and visibility
- **lessons** - Individual lessons with type (video, live, doc)
- **live_sessions** - Active or completed live sessions
- **recordings** - Recorded video metadata
- **assignments** - Assignment definitions per lesson
- **submissions** - Student submissions with grades
- **chats** - Real-time chat messages per room
- **session_participants** - Track attendance

All tables use UUID primary keys and timestamps for auditability.

## Development Workflow

### Running Tests (when available)
```bash
make test
```

### Building Locally
```bash
make build
# Output: ./bin/vtp
```

### Code Formatting
```bash
make fmt    # Format Go code
make lint   # Run go vet
```

## Current Phase Progress

### ✓ Phase 0: Project Setup & Foundations (Weeks 1–2)
- ✓ Go module initialized
- ✓ Folder structure created
- ✓ Git initialized with .gitignore
- ✓ PostgreSQL schema designed with all MVP tables
- ✓ Docker Compose stack configured
  - PostgreSQL 15
  - Redis 7
  - Mediasoup SFU (Node.js)
  - MinIO (S3 storage)
  - pgAdmin (DB management)
- ✓ Database connection & migration runner implemented
- ✓ Data models defined (User, Course, Lesson, etc.)
- ✓ Configuration template (.env.example)
- ✓ Makefile for common tasks

### Next: Phase 1a (Weeks 2–3)
- [ ] Implement JWT auth service
- [ ] User registration & login endpoints
- [ ] Password hashing (bcrypt)
- [ ] Token validation middleware

## Environment Variables

Copy `.env.example` to `.env` and customize:

```bash
cp .env.example .env
```

Key variables:
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string
- `JWT_SECRET` - Change this in production!
- `S3_*` - MinIO credentials
- `MEDIASOUP_URL` - SFU server location

## Roadmap

| Phase | Timeline | Deliverables |
|-------|----------|--------------|
| **0** | Weeks 1–2 | ✓ Project setup, DB schema, Docker stack |
| **1a** | Weeks 2–3 | Auth service (JWT, registration, login) |
| **1b** | Weeks 3–4 | Signalling API (Socket.IO, room logic) |
| **1c** | Weeks 4–5 | Mediasoup integration (WebRTC media) |
| **2a** | Weeks 5–6 | Recording pipeline (FFmpeg → S3) |
| **2b** | Weeks 6–7 | Frontend (React, video player, HLS playback) |
| **2c** | Weeks 7–8 | Live UI, chat, end-to-end demo |
| **3** | Weeks 8–10 | MVP polish, load testing, deployment |
| **5E** | Weeks 10–11 | ✓ Analytics frontend module & testing |
| **5G** | Weeks 11–12 | ✓ 5G Network Optimization (Nov 28, 2025) |

## Security Notes (MVP)

⚠️ **This is MVP code. Before production:**
- [ ] Change `JWT_SECRET` to a strong value
- [ ] Enable HTTPS/TLS for all services
- [ ] Implement rate limiting on APIs
- [ ] Add CORS middleware with allowed origins
- [ ] Validate and sanitize all user inputs
- [ ] Encrypt sensitive data at rest (S3, DB)
- [ ] Set up WAF rules for the API gateway
- [ ] Regular security audits

## Contribution Guidelines

When code is generated by agents:
1. Agent proposes code with clear context (input/output, error handling)
2. You review and approve before merging
3. Code includes tests where applicable
4. Commits follow conventional commits (feat:, fix:, test:, etc.)

## Useful Commands

```bash
# View all available commands
make help

# Monitor all container logs
make logs

# Access PostgreSQL directly
psql -U postgres -d vtp_db -h localhost

# View MinIO console
open http://localhost:9001  # macOS
xdg-open http://localhost:9001  # Linux
start http://localhost:9001  # Windows
```

## Troubleshooting

### PostgreSQL connection fails
```bash
docker-compose ps  # Check if postgres container is running
docker-compose logs postgres  # View postgres logs
```

### Mediasoup won't start
```bash
docker-compose logs mediasoup
# Ensure node_modules are installed in mediasoup/
cd mediasoup && npm install
```

### Port already in use
```bash
# Change ports in docker-compose.yml or kill process:
lsof -i :8080  # macOS/Linux
Get-Process -Name "*" | Where-Object {$_.Handles -gt 100}  # Windows
```

## References

- [Mediasoup Documentation](https://mediasoup.org/documentation/)
- [PostgreSQL Performance Tips](https://wiki.postgresql.org/wiki/Performance_Optimization)
- [Docker Compose Docs](https://docs.docker.com/compose/)
- [Go Best Practices](https://golang.org/doc/effective_go)

## License

MIT (update as needed)

## Contact & Support

For questions or issues, refer to the approved implementation plan or contact the development team.

---

**Next Step:** Move to Phase 1a implementation (Auth Service) when ready. Code will be generated with agent assistance and your approval on each component.
