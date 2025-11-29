# Phase 4: Analytics & Reporting - Overview & Planning 📊

**Status:** PLANNED (After Phase 2B)  
**Estimated Duration:** 3-4 days  
**Complexity:** Medium  
**Dependencies:** Phase 1a ✅ Phase 2a ✅ Phase 3 ✅ Phase 2B ⏳  

---

## 🎯 Overview: Why Analytics Matter

Currently, instructors can:
- ✅ Create courses
- ✅ Record lectures
- ✅ Stream content
- ❌ But they DON'T know:
  - Who watched what?
  - For how long?
  - What was engagement?
  - Which lectures were helpful?

Phase 4 solves this with comprehensive analytics.

---

## 📊 Three Analytics Systems

### 1. Engagement Analytics (Student Level)

**Questions Answered:**
- Which students watched which lectures?
- How much did they watch (time)?
- Did they watch multiple times?
- When did they watch (real-time)?
- What quality did they watch at?

**Example Dashboard:**
```
Student: John Smith (john@university.edu)
┌─────────────────────────────────────────┐
│ CS101 Course Analytics                  │
├─────────────────────────────────────────┤
│ Lectures Watched: 12 of 15             │
│ Total Watch Time: 18 hours 45 min      │
│ Average Per Lecture: 93 minutes        │
│ Last Watched: 2025-11-24 14:30         │
│ Quality Preference: 1080p (High)       │
│ Engagement Score: 85/100               │
│ Buffers Experienced: 2                 │
│ Completion Rate: 80%                   │
└─────────────────────────────────────────┘
```

### 2. Lecture Analytics (Content Level)

**Questions Answered:**
- How many students watched this lecture?
- Total watch time across all students?
- Which parts were rewatched?
- Where did students drop off?
- Quality distribution?

**Example Dashboard:**
```
Lecture: "Introduction to Variables"
┌─────────────────────────────────────────┐
│ Lecture Analytics                       │
├─────────────────────────────────────────┤
│ Recording Duration: 60 minutes          │
│ Unique Viewers: 42 of 50 students      │
│ Total Views: 65 (some watched twice)   │
│ Average Watch Time: 58 minutes         │
│ Completion Rate: 96%                   │
│ Most Replayed Section: 25:30-27:45    │
│ Quality Distribution:                  │
│   • 500kbps:  5%  (poor connection)   │
│   • 1000kbps: 25% (normal)            │
│   • 2000kbps: 40% (good)              │
│   • 4000kbps: 30% (excellent)         │
│ Peak Viewers: 35 concurrent           │
│ Buffer Events: 3                       │
└─────────────────────────────────────────┘
```

### 3. Course Analytics (Program Level)

**Questions Answered:**
- Overall course engagement?
- Attendance patterns?
- Performance by topic?
- Which lectures are most valuable?
- Trends over semester?

**Example Dashboard:**
```
Course: CS101 - Introduction to Computer Science
┌─────────────────────────────────────────────┐
│ Course Analytics                            │
├─────────────────────────────────────────────┤
│ Total Students: 50                         │
│ Attending Students: 47 (94%)              │
│ Total Lectures: 15                        │
│ Total View Sessions: 425                  │
│ Total Unique Viewers: 49                  │
│ Average Attendance: 94%                   │
│ Total Watch Time: 185 hours               │
│ Average Watch Time Per Student: 3.7 hrs   │
│ Course Engagement Score: 88/100           │
│ Most Watched Lecture: #8 (65 views)      │
│ Least Watched Lecture: #15 (31 views)    │
│ Trend: ↗ Increasing engagement           │
│                                           │
│ Performance Metrics:                      │
│ • Above 85% Engagement: 38 students      │
│ • 70-85% Engagement: 9 students          │
│ • Below 70% Engagement: 2 students       │
│ • At Risk (no activity): 1 student       │
└─────────────────────────────────────────────┘
```

---

## 🏗️ Architecture

### Data Collection Points

```
Event Stream:
├─ Recording Started          → Analytics DB
├─ Recording Stopped          → Analytics DB
├─ Playback Session Started   → Analytics DB
├─ Segment Requested          → Analytics DB
├─ Quality Selected           → Analytics DB
├─ Quality Changed            → Analytics DB
├─ Buffer Event Occurred      → Analytics DB
├─ Playback Paused            → Analytics DB
├─ Playback Resumed           → Analytics DB
├─ Playback Stopped           → Analytics DB
└─ Session Ended              → Analytics DB
```

### Processing Pipeline

```
Raw Events
    ↓
Event Parser (validate, normalize)
    ↓
Session Aggregator (group related events)
    ↓
Metrics Calculator (derive statistics)
    ↓
Analytics DB Storage
    ↓
Query Service
    ↓
API Endpoints
    ↓
Dashboard UI / Reports
```

### Database Schema (New Tables)

```sql
-- Playback Sessions (one record per viewer per lecture)
CREATE TABLE playback_sessions (
  id UUID PRIMARY KEY,
  recording_id UUID NOT NULL REFERENCES recordings(id),
  user_id UUID NOT NULL REFERENCES users(id),
  session_start TIMESTAMP NOT NULL,
  session_end TIMESTAMP,
  total_duration_seconds INT,
  watched_duration_seconds INT,
  pause_count INT DEFAULT 0,
  resume_count INT DEFAULT 0,
  quality_selected VARCHAR(20),
  buffer_events INT DEFAULT 0,
  completion_rate DECIMAL(5,2),
  created_at TIMESTAMP
);

-- Quality Events (track quality changes)
CREATE TABLE quality_events (
  id UUID PRIMARY KEY,
  session_id UUID NOT NULL REFERENCES playback_sessions(id),
  timestamp TIMESTAMP NOT NULL,
  bitrate INT,
  resolution VARCHAR(20),
  reason VARCHAR(50), -- user_selected, auto_downgrade, auto_upgrade
  created_at TIMESTAMP
);

-- Engagement Metrics (aggregate stats)
CREATE TABLE engagement_metrics (
  id UUID PRIMARY KEY,
  recording_id UUID NOT NULL REFERENCES recordings(id),
  user_id UUID NOT NULL REFERENCES users(id),
  total_watch_time_seconds INT,
  completion_percentage INT,
  rewatch_count INT,
  avg_quality VARCHAR(20),
  last_watched TIMESTAMP,
  engagement_score INT, -- 0-100
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

-- Lecture Statistics (aggregate per lecture)
CREATE TABLE lecture_statistics (
  id UUID PRIMARY KEY,
  recording_id UUID NOT NULL REFERENCES recordings(id),
  unique_viewers INT,
  total_views INT,
  avg_watch_time_seconds INT,
  completion_rate DECIMAL(5,2),
  peak_concurrent_viewers INT,
  total_buffer_events INT,
  quality_distribution JSON, -- {500: 5, 1000: 25, 2000: 40, 4000: 30}
  most_replayed_section VARCHAR(50),
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

-- Course Statistics (aggregate per course)
CREATE TABLE course_statistics (
  id UUID PRIMARY KEY,
  course_id UUID NOT NULL REFERENCES courses(id),
  total_students INT,
  attending_students INT,
  total_lectures INT,
  total_view_sessions INT,
  avg_attendance_rate DECIMAL(5,2),
  total_watch_time_seconds BIGINT,
  course_engagement_score INT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

-- Create Indexes
CREATE INDEX idx_playback_sessions_recording ON playback_sessions(recording_id);
CREATE INDEX idx_playback_sessions_user ON playback_sessions(user_id);
CREATE INDEX idx_playback_sessions_start ON playback_sessions(session_start);
CREATE INDEX idx_engagement_metrics_recording ON engagement_metrics(recording_id);
CREATE INDEX idx_engagement_metrics_user ON engagement_metrics(user_id);
CREATE INDEX idx_lecture_statistics_recording ON lecture_statistics(recording_id);
CREATE INDEX idx_course_statistics_course ON course_statistics(course_id);
```

---

## 🎯 Implementation Plan (4 Days)

### Day 1: Event Collection & Storage

**Objective:** Capture all user actions

```
Files to Create:
├─ pkg/analytics/events.go (150+ lines)
│  ├─ Event types
│  ├─ EventCollector struct
│  ├─ Event serialization
│  └─ Batch event processing
│
├─ pkg/analytics/storage.go (200+ lines)
│  ├─ Store playback sessions
│  ├─ Store quality events
│  ├─ Store buffer events
│  └─ Batch write optimization
│
└─ pkg/analytics/events_test.go (150+ lines)
   ├─ Event parsing tests
   ├─ Storage tests
   └─ Batch processing tests
```

### Day 2: Metrics Calculation

**Objective:** Calculate analytics from raw events

```
Files to Create:
├─ pkg/analytics/calculator.go (250+ lines)
│  ├─ EngagementCalculator struct
│  ├─ Calculate completion rate
│  ├─ Calculate engagement score
│  ├─ Detect quality patterns
│  └─ Identify drop-off points
│
├─ pkg/analytics/aggregation.go (200+ lines)
│  ├─ Aggregate by lecture
│  ├─ Aggregate by course
│  ├─ Aggregate by student
│  └─ Generate summary stats
│
└─ pkg/analytics/calculator_test.go (150+ lines)
   ├─ Calculation accuracy tests
   ├─ Aggregation tests
   └─ Edge case handling
```

### Day 3: Query Service & API

**Objective:** Expose analytics through REST API

```
Files to Create:
├─ pkg/analytics/queries.go (200+ lines)
│  ├─ Query student engagement
│  ├─ Query lecture statistics
│  ├─ Query course statistics
│  ├─ Date range filtering
│  └─ Sorting and pagination
│
├─ pkg/analytics/handlers.go (250+ lines)
│  ├─ GET /api/v1/analytics/students/{id}
│  ├─ GET /api/v1/analytics/lectures/{id}
│  ├─ GET /api/v1/analytics/courses/{id}
│  ├─ GET /api/v1/analytics/reports/attendance
│  ├─ GET /api/v1/analytics/reports/engagement
│  └─ GET /api/v1/analytics/reports/performance
│
└─ pkg/analytics/handlers_test.go (150+ lines)
   ├─ API endpoint tests
   ├─ Response format validation
   └─ Authorization tests
```

### Day 4: Integration & Reporting

**Objective:** Integrate with existing system, create reports

```
Integration Tasks:
├─ Update pkg/recording/playback.go
│  ├─ Add event collection calls
│  ├─ Track playback events
│  └─ Send to analytics
│
├─ Update cmd/main.go
│  ├─ Initialize analytics engine
│  ├─ Register 6 analytics endpoints
│  └─ Display startup info
│
├─ Create pkg/analytics/reports.go (200+ lines)
│  ├─ Generate attendance reports
│  ├─ Generate engagement reports
│  ├─ Generate performance reports
│  └─ Export to CSV/PDF
│
└─ Create migrations/005_analytics_schema.sql
   ├─ Create all analytics tables
   └─ Create all indexes
```

---

## 📈 New API Endpoints (6 Total)

```
PHASE 4 - Analytics & Reporting (6 Endpoints)

Student Analytics:
GET    /api/v1/analytics/students/{student_id}
       Response: Student engagement metrics, watched lectures, completion rates

Lecture Analytics:
GET    /api/v1/analytics/lectures/{recording_id}
       Response: Viewership stats, engagement metrics, quality distribution

Course Analytics:
GET    /api/v1/analytics/courses/{course_id}
       Response: Overall course stats, attendance, engagement trends

Attendance Reports:
GET    /api/v1/analytics/reports/attendance?course_id={id}&start_date={date}&end_date={date}
       Response: Student attendance records, patterns, alerts

Engagement Reports:
GET    /api/v1/analytics/reports/engagement?course_id={id}
       Response: Student engagement scores, trends, recommendations

Performance Reports:
GET    /api/v1/analytics/reports/performance?course_id={id}
       Response: Course performance metrics, problem areas, success areas
```

---

## 📊 Analytics Data Examples

### Student Engagement Metric
```json
{
  "student_id": "uuid-...",
  "name": "John Smith",
  "email": "john@university.edu",
  "course_id": "uuid-...",
  "metrics": {
    "lectures_watched": 12,
    "lectures_total": 15,
    "watch_rate": 80,
    "total_watch_time_minutes": 1125,
    "average_completion_percent": 95,
    "engagement_score": 88,
    "quality_preference": "1080p",
    "buffer_events": 2,
    "last_watched_at": "2025-11-24T14:30:00Z"
  },
  "watch_history": [
    {
      "lecture_number": 1,
      "recording_id": "uuid-...",
      "watch_sessions": 1,
      "total_watched_minutes": 58,
      "completion_percent": 97,
      "last_watched_at": "2025-11-20T10:00:00Z"
    }
  ],
  "alerts": [
    "Engagement declining: -15% from last week"
  ]
}
```

### Lecture Statistics
```json
{
  "recording_id": "uuid-...",
  "lecture_number": 5,
  "title": "Functions and Scope",
  "duration_minutes": 60,
  "statistics": {
    "unique_viewers": 47,
    "total_views": 51,
    "average_watch_time_minutes": 57,
    "completion_rate": 95,
    "engagement_score": 92,
    "peak_concurrent_viewers": 35,
    "buffer_events": 3,
    "published_at": "2025-11-18T08:00:00Z"
  },
  "quality_distribution": {
    "500kbps": 5,
    "1000kbps": 25,
    "2000kbps": 40,
    "4000kbps": 30
  },
  "replay_analysis": {
    "most_replayed_section": "25:30-27:45",
    "rewatch_reason": "Difficult concept - loops explanation"
  }
}
```

### Course Statistics
```json
{
  "course_id": "uuid-...",
  "course_name": "CS101 - Intro to CS",
  "semester": "Fall",
  "year": 2025,
  "statistics": {
    "total_students": 50,
    "attending_students": 47,
    "attendance_rate": 94,
    "total_lectures": 15,
    "total_view_sessions": 425,
    "unique_viewers": 49,
    "total_watch_hours": 185,
    "average_watch_hours_per_student": 3.7,
    "course_engagement_score": 88,
    "trend": "increasing"
  },
  "engagement_distribution": {
    "excellent": 38,
    "good": 9,
    "poor": 2,
    "at_risk": 1
  },
  "top_lectures": [
    {"number": 8, "title": "Arrays", "views": 65},
    {"number": 5, "title": "Functions", "views": 62}
  ],
  "needs_attention": [
    {"student": "jane@university.edu", "engagement": 45},
    {"student": "bob@university.edu", "engagement": 38}
  ]
}
```

---

## 🎯 Success Criteria

### Day 1
- [x] Event types defined
- [x] Event collection implemented
- [x] Events stored in database
- [x] Batch processing working
- [x] 20+ unit tests passing

### Day 2
- [x] Metrics calculated correctly
- [x] Aggregation working
- [x] Accuracy validated
- [x] Performance optimized
- [x] 15+ unit tests passing

### Day 3
- [x] All 6 API endpoints working
- [x] Query performance good (<200ms)
- [x] Response formats correct
- [x] Authorization checks in place
- [x] 20+ integration tests passing

### Day 4
- [x] Full integration complete
- [x] Reports generating correctly
- [x] Production binary builds
- [x] All systems working together
- [x] Deployment ready

---

## 📚 Reports Available

### Attendance Report
```
Shows which students attended which lectures and when
Identifies patterns (e.g., students never attending Tuesday lectures)
Alerts on chronic absences
```

### Engagement Report
```
Student engagement scores (0-100)
Trends over time
Comparison to class average
Recommendations for at-risk students
```

### Performance Report
```
Course performance metrics
Which lectures are most effective
Which topics students struggle with
Recommendations for curriculum improvement
```

---

## 🚀 Timeline Summary

```
PHASE COMPLETION SEQUENCE:

✅ Phase 1a: Authentication (COMPLETE)
✅ Phase 1b: WebRTC Signalling (COMPLETE)
✅ Phase 2a: Recording System (COMPLETE)
✅ Phase 3: Course Management (COMPLETE)
🚀 Phase 2B: Advanced Streaming (NEXT - 4-5 days)
   ├─ Day 1: ABR Engine
   ├─ Day 2: Multi-bitrate Transcoding
   ├─ Day 3: Live Distribution
   └─ Day 4: Integration

📊 Phase 4: Analytics & Reporting (AFTER 2B - 3-4 days)
   ├─ Day 1: Event Collection
   ├─ Day 2: Metrics Calculation
   ├─ Day 3: API & Queries
   └─ Day 4: Integration & Reports

Total Implementation: ~8-9 days
Total Endpoints: 40 + 6 + 6 = 52 endpoints
Production Ready: YES
```

---

## 💡 Benefits After Phase 4

With all four phases complete:

**For Students:**
- ✅ Watch lectures on any network (adaptive streaming)
- ✅ Personalized learning experience (ABR)
- ✅ Track own progress
- ✅ See engagement metrics

**For Instructors:**
- ✅ Organize courses and enrollments
- ✅ Know who's watching what
- ✅ Identify struggling students
- ✅ Optimize curriculum
- ✅ Generate attendance reports

**For Administrators:**
- ✅ Overall platform usage
- ✅ Course performance
- ✅ Technology impact
- ✅ Data-driven decisions

---

## ✨ Next Steps

1. ✅ Complete Phase 3 (TODAY)
2. 🚀 Start Phase 2B (NEXT 4-5 days)
3. 📊 Then Phase 4 (3-4 days after Phase 2B)
4. 🎓 Full system ready for production deployment

**Recommended sequence provides maximum value:**
- Course Management first (organize content)
- Advanced Streaming second (improve user experience)
- Analytics last (measure impact)
