# VTP - منصة التعليم الافتراضية | Virtual Teaching Platform

**Arabic Educational SaaS Platform** - Like Google Classroom + Zoom + SWEEDU  
مشابه لـ Google Classroom + Zoom + SWEEDU

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Next.js](https://img.shields.io/badge/Next.js-14-black?style=flat&logo=next.js)](https://nextjs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?style=flat&logo=postgresql)](https://postgresql.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## 🎯 Overview | نظرة عامة

VTP is a comprehensive educational platform designed for Arabic-speaking students (Syrian curriculum focus) with:

- ✅ **Online Classes** - Live video meetings via Jitsi/Google Meet/Zoom
- ✅ **Subject Management** - 10+ subjects for 12th grade (Baccalaureate)
- ✅ **Assignments & Submissions** - Upload, grade, and track homework
- ✅ **Attendance Tracking** - Automatic & manual attendance with reports
- ✅ **Multi-Dashboard** - Student, Teacher, and Admin dashboards
- ✅ **Video Integration** - Google Meet / Zoom / Jitsi support
- ✅ **Notifications** - SMS / Email / In-App alerts
- ✅ **Mobile-Friendly** - PWA for Android & iOS
- ✅ **RTL Arabic Support** - Full right-to-left interface

---

## 🏗️ Architecture | البنية التقنية

```
┌─────────────────────────────────────────────────────────────────┐
│                     Browser / Mobile App                         │
│                   (Next.js + Tailwind CSS)                      │
└────────────────────────┬────────────────────────────────────────┘
                         │ REST API + WebSocket
┌────────────────────────▼────────────────────────────────────────┐
│                   Go API Gateway (port 8080)                     │
│     - Authentication (JWT + 2FA)                                │
│     - Course Management                                          │
│     - Assignment System                                          │
│     - Attendance Tracking                                        │
│     - Notifications                                              │
│     - Video Integration (Jitsi/Google Meet/Zoom)               │
└────────────────────────┬────────────────────────────────────────┘
                         │
    ┌────────────────────┼────────────────────┐
    │                    │                    │
┌───▼──────────┐   ┌─────▼─────────┐   ┌──────▼──────┐
│  PostgreSQL  │   │  Redis Cache  │   │  File Store │
│   Database   │   │   Sessions    │   │  (S3/MinIO) │
└──────────────┘   └───────────────┘   └─────────────┘
```

---

## 🚀 Quick Start | بداية سريعة

### Prerequisites

- Go 1.24+
- Node.js 18+
- PostgreSQL 15+
- Docker (optional)

### 1. Clone & Setup

```bash
git clone https://github.com/Bashar444/VTP.git
cd VTP

# Copy environment file
cp .env.example .env
```

### 2. Configure Environment

Edit `.env` file:

```env
# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/vtp_db?sslmode=disable

# JWT Authentication
JWT_SECRET=your-super-secret-key-change-in-production
JWT_EXPIRY_HOURS=24
JWT_REFRESH_EXPIRY_HOURS=168

# Video Integration (choose one)
JITSI_SERVER_URL=https://meet.jit.si
# GOOGLE_MEET_API_KEY=your-google-api-key
# ZOOM_API_KEY=your-zoom-api-key

# Email (optional)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-app-password
```

### 3. Run with Docker (Recommended)

```bash
docker-compose up -d
```

### 4. Run Locally

```bash
# Backend
go run cmd/main.go

# Frontend (in another terminal)
cd vtp-frontend
npm install
npm run dev
```

### 5. Access the Platform

- **Frontend**: http://localhost:3000
- **API**: http://localhost:8080
- **API Health**: http://localhost:8080/health

---

## 📚 Features | المميزات

### 🎓 For Students | للطلاب

- View enrolled courses and subjects
- Join live video classes
- Submit assignments online
- Track grades and attendance
- Receive notifications

### 👨‍🏫 For Teachers | للمعلمين

- Create and manage courses
- Schedule live video classes (Jitsi/Meet/Zoom)
- Create assignments with due dates
- Grade submissions
- Track student attendance
- Upload study materials (PDF, videos)

### 👔 For Admins | للإدارة

- User management (students, teachers)
- School year/term configuration
- Grade levels and class sections
- Attendance reports
- Platform announcements
- System dashboard

---

## 📖 API Documentation | توثيق الـ API

### Authentication

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/auth/register` | POST | Register new user |
| `/api/v1/auth/login` | POST | User login |
| `/api/v1/auth/refresh` | POST | Refresh JWT token |
| `/api/v1/auth/profile` | GET | Get user profile |
| `/api/v1/auth/2fa/setup` | POST | Setup 2FA |

### Subjects & Courses

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/subjects` | GET | List all subjects |
| `/api/v1/subjects/{id}` | GET | Get subject details |
| `/api/v1/courses` | GET/POST | List/Create courses |
| `/api/v1/courses/{id}` | GET/PUT | Get/Update course |

### Meetings & Video

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/meetings` | GET/POST | List/Create meetings |
| `/api/v1/meetings/{id}` | GET/PUT | Get/Update meeting |
| `/api/v1/meetings/{id}/join` | GET | Get meeting join link |
| `/api/v1/meetings/{id}/video` | POST | Create video integration |
| `/api/v1/video/providers` | GET | List video providers |

### Assignments

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/assignments` | GET/POST | List/Create assignments |
| `/api/v1/assignments/{id}` | GET | Get assignment details |
| `/api/v1/assignments/{id}/submit` | POST | Submit assignment |
| `/api/v1/assignments/{id}/submissions` | GET | List submissions |
| `/api/v1/submissions/{id}/grade` | POST | Grade submission |

### Attendance

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/attendance` | POST | Record attendance |
| `/api/v1/attendance/bulk` | POST | Bulk attendance |
| `/api/v1/attendance/student/{id}` | GET | Student attendance |
| `/api/v1/attendance/student/{id}/stats` | GET | Attendance stats |
| `/api/v1/attendance/class/{id}` | GET | Class attendance |
| `/api/v1/attendance/report` | GET | Generate report |

### Notifications

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/notifications` | GET/POST | List/Create notifications |
| `/api/v1/notifications/{id}/read` | POST | Mark as read |
| `/api/v1/notifications/read-all` | POST | Mark all as read |
| `/api/v1/notifications/unread-count` | GET | Get unread count |

### Admin

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/admin/dashboard` | GET | Admin dashboard stats |
| `/api/v1/admin/users` | GET | List all users |
| `/api/v1/admin/school-terms` | GET/POST | Manage school terms |
| `/api/v1/admin/grade-levels` | GET | List grade levels |
| `/api/v1/admin/class-sections` | GET/POST | Manage class sections |
| `/api/v1/admin/reports/attendance` | GET | Attendance report |
| `/api/v1/admin/announce` | POST | Create announcement |

---

## 🎥 Video Integration | دمج الفيديو

### Jitsi Meet (Recommended - Free)

No configuration needed! Works out of the box.

```json
POST /api/v1/meetings/{id}/video
{
  "provider": "jitsi",
  "title": "Math Class"
}
```

### Google Meet

Requires Google Workspace API setup:

1. Create project in Google Cloud Console
2. Enable Google Calendar API
3. Create OAuth credentials
4. Set `GOOGLE_MEET_API_KEY` in `.env`

### Zoom

Requires Zoom Developer account:

1. Create app at marketplace.zoom.us
2. Get API Key and Secret
3. Set `ZOOM_API_KEY` and `ZOOM_API_SECRET` in `.env`

---

## 🗃️ Database Schema | قاعدة البيانات

### Core Tables

- `users` - Students, teachers, admins
- `courses` - Educational courses
- `subjects` - Academic subjects (Math, Physics, etc.)
- `meetings` - Scheduled live classes
- `assignments` - Homework and projects
- `submissions` - Student assignment submissions

### Educational Management

- `school_terms` - Academic semesters
- `grade_levels` - 1st-12th grade
- `class_sections` - Class divisions (12-A, 12-B)
- `attendance` - Attendance records
- `student_grades` - Grade records

### Communication

- `notifications` - User notifications
- `meeting_integrations` - Video provider links

---

## 🇸🇾 Syrian Curriculum | المنهج السوري

Pre-configured 12th grade (Baccalaureate) subjects:

| Arabic | English | Category |
|--------|---------|----------|
| الرياضيات | Mathematics | العلوم |
| الفيزياء | Physics | العلوم |
| الكيمياء | Chemistry | العلوم |
| علم الأحياء | Biology | العلوم |
| اللغة العربية | Arabic Language | اللغات |
| اللغة الإنجليزية | English Language | اللغات |
| اللغة الفرنسية | French Language | اللغات |
| التاريخ | History | الاجتماعيات |
| الجغرافيا | Geography | الاجتماعيات |
| الفلسفة والمنطق | Philosophy | الإنسانيات |

---

## 📱 Frontend | الواجهة الأمامية

Built with Next.js 14 + Tailwind CSS:

- **RTL Support** - Full Arabic interface
- **Responsive Design** - Mobile-first approach
- **Dark Mode** - Light/dark theme support
- **PWA Ready** - Install as mobile app

### Pages

- `/` - Landing page
- `/login` - User login
- `/register` - New user registration
- `/dashboard` - Role-based dashboard redirect
- `/my-courses` - Student courses
- `/assignments` - Assignment list
- `/stream` - Live video streaming
- `/profile` - User profile

---

## 🔧 Development | التطوير

### Project Structure

```
VTP/
├── cmd/
│   └── main.go              # Application entry
├── pkg/
│   ├── auth/                # Authentication
│   ├── admin/               # Admin handlers
│   ├── assignment/          # Assignment system
│   ├── attendance/          # Attendance tracking
│   ├── course/              # Course management
│   ├── meeting/             # Meeting scheduling
│   ├── notification/        # Notifications
│   ├── subject/             # Subject management
│   ├── videointegration/    # Jitsi/Meet/Zoom
│   ├── models/              # Data models
│   └── db/                  # Database
├── migrations/              # SQL migrations
├── vtp-frontend/            # Next.js frontend
├── docker-compose.yml
├── Dockerfile
└── README.md
```

### Running Tests

```bash
# Backend tests
go test ./...

# Frontend tests
cd vtp-frontend
npm test
```

### Building for Production

```bash
# Backend
go build -o vtp cmd/main.go

# Frontend
cd vtp-frontend
npm run build
```

---

## 🚢 Deployment | النشر

### DigitalOcean (Recommended for Syria)

Low-cost deployment (~$12-24/month):

```bash
# 1. Create droplet (2-4GB RAM)
# 2. Install Docker
# 3. Clone repository
# 4. Run docker-compose up -d
```

### Railway

One-click deployment available.

### Vercel (Frontend Only)

```bash
cd vtp-frontend
vercel deploy
```

---

## 🤝 Contributing | المساهمة

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing`)
5. Open Pull Request

---

## 📄 License | الرخصة

MIT License - See [LICENSE](LICENSE) file

---

## 📞 Support | الدعم

- **Issues**: [GitHub Issues](https://github.com/Bashar444/VTP/issues)

---

**Built with ❤️ for Syrian Students**  
صُنع بـ ❤️ للطلاب السوريين
