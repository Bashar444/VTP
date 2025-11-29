# Phase 3 - Course Management System
## Initialization & Planning Document

**Status:** ✅ Ready to Begin (Phase 2A Production Ready)  
**Start Date:** November 24, 2025  
**Estimated Duration:** 5-6 days  
**Total Estimated Lines of Code:** 3,500+ lines  

---

## 📋 Executive Summary

Phase 3 integrates the Phase 2A Recording System with Course Management, enabling:
- **Instructors** to link recordings to courses
- **Students** to access course-specific recordings
- **Administrators** to manage course content and permissions
- **Analytics** to track student engagement per course

**Status Before Phase 3:**
- ✅ Phase 1a: Authentication (Complete)
- ✅ Phase 1b: WebRTC Signaling (Complete)
- ✅ Phase 2a: Recording System (Complete - Days 1-4)
- 🎯 Phase 3: Course Management (Starting Now)

---

## 🎯 Phase 3 Objectives

### Primary Goals
1. **Course CRUD Operations** - Create, read, update, delete courses
2. **Recording-to-Course Linking** - Associate recordings with courses
3. **Access Control** - Role-based permissions (Instructor, Student, Admin)
4. **Course Content Management** - Organize lectures by course
5. **Student Enrollment** - Manage course enrollment and access

### Secondary Goals
1. **Course Dashboard** - View course statistics and metrics
2. **Recording Management** - Manage recordings within courses
3. **Audit Logging** - Track course and recording access
4. **Bulk Operations** - Import/export course data
5. **Search & Filtering** - Find courses and recordings efficiently

---

## 📊 Phase 3 Architecture

### Database Schema (Phase 3)

**New Tables:**
```sql
CREATE TABLE courses (
    id UUID PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    instructor_id UUID NOT NULL REFERENCES users(id),
    department VARCHAR(100),
    semester VARCHAR(20),
    year INT,
    status VARCHAR(20) DEFAULT 'active',
    max_students INT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE course_recordings (
    id UUID PRIMARY KEY,
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    recording_id UUID NOT NULL REFERENCES recordings(id) ON DELETE CASCADE,
    lecture_number INT,
    lecture_title VARCHAR(255),
    sequence_order INT,
    is_published BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(course_id, recording_id)
);

CREATE TABLE course_enrollments (
    id UUID PRIMARY KEY,
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    enrollment_date TIMESTAMP DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'active',
    UNIQUE(course_id, student_id)
);

CREATE TABLE course_permissions (
    id UUID PRIMARY KEY,
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(course_id, user_id)
);

CREATE TABLE course_activity (
    id UUID PRIMARY KEY,
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    details JSONB,
    timestamp TIMESTAMP DEFAULT NOW()
);
```

**Indexes for Performance:**
```sql
CREATE INDEX idx_courses_instructor ON courses(instructor_id);
CREATE INDEX idx_courses_semester ON courses(semester, year);
CREATE INDEX idx_course_recordings_course ON course_recordings(course_id);
CREATE INDEX idx_course_recordings_recording ON course_recordings(recording_id);
CREATE INDEX idx_course_enrollments_course ON course_enrollments(course_id);
CREATE INDEX idx_course_enrollments_student ON course_enrollments(student_id);
CREATE INDEX idx_course_permissions_course ON course_permissions(course_id);
CREATE INDEX idx_course_permissions_user ON course_permissions(user_id);
CREATE INDEX idx_course_activity_course ON course_activity(course_id);
CREATE INDEX idx_course_activity_timestamp ON course_activity(timestamp);
```

### API Endpoints (Phase 3)

**Course Management (12 endpoints):**
```
POST   /api/v1/courses                    - Create course
GET    /api/v1/courses                    - List courses (filtered by user)
GET    /api/v1/courses/{id}               - Get course details
PUT    /api/v1/courses/{id}               - Update course
DELETE /api/v1/courses/{id}               - Delete course
GET    /api/v1/courses/{id}/recordings    - List course recordings
POST   /api/v1/courses/{id}/recordings    - Add recording to course
DELETE /api/v1/courses/{id}/recordings/{rid} - Remove recording from course
POST   /api/v1/courses/{id}/publish       - Publish course
GET    /api/v1/courses/{id}/stats         - Get course statistics
POST   /api/v1/courses/bulk/import        - Import courses (CSV/JSON)
GET    /api/v1/courses/bulk/export        - Export courses
```

**Enrollment Management (6 endpoints):**
```
POST   /api/v1/courses/{id}/enroll       - Enroll student in course
DELETE /api/v1/courses/{id}/enroll/{sid} - Remove student enrollment
GET    /api/v1/courses/{id}/students     - List enrolled students
GET    /api/v1/courses/{id}/enrollment   - Get user's enrollment status
POST   /api/v1/courses/{id}/invite       - Send course invitation
GET    /api/v1/enrollments               - List user's enrolled courses
```

**Permissions Management (4 endpoints):**
```
POST   /api/v1/courses/{id}/permissions     - Set user permissions
GET    /api/v1/courses/{id}/permissions     - Get course permissions
DELETE /api/v1/courses/{id}/permissions/{uid} - Remove user permissions
GET    /api/v1/courses/{id}/access-check    - Check user access
```

**Analytics (4 endpoints):**
```
GET    /api/v1/courses/{id}/analytics       - Course analytics
GET    /api/v1/courses/{id}/engagement      - Student engagement
GET    /api/v1/recordings/{id}/access-log   - Recording access log
GET    /api/v1/courses/{id}/reports/export  - Export course report
```

---

## 📁 File Structure (Phase 3)

**New Files to Create:**

```
pkg/course/
├── types.go                  (100 lines) - Type definitions
├── service.go               (350 lines) - Business logic
├── service_test.go          (200 lines) - Unit tests
├── handlers.go              (500 lines) - HTTP handlers
├── permissions.go           (200 lines) - Permission checking
├── permissions_test.go      (150 lines) - Permission tests
├── enrollment.go            (250 lines) - Enrollment logic
├── enrollment_test.go       (150 lines) - Enrollment tests
├── analytics.go             (300 lines) - Analytics & reporting
├── analytics_test.go        (150 lines) - Analytics tests
└── activity_log.go          (200 lines) - Activity tracking

migrations/
└── 003_courses_schema.sql   (150 lines) - Database schema

cmd/
└── main.go                  (updated)   - Add course service initialization

frontend/
└── course-dashboard.html    (600 lines) - Course UI (if needed)
```

**Files to Modify:**

```
cmd/main.go                 - Add course service initialization
pkg/auth/middleware.go      - Add course permission checks
```

---

## 🔄 Implementation Plan (5 Days)

### Day 1: Database & Core Types (2-3 hours)
- [ ] Create migration file `003_courses_schema.sql`
- [ ] Define Course types (Course, Enrollment, Permission)
- [ ] Define request/response types
- [ ] Create database initialization function
- [ ] Write schema validation tests

**Deliverables:**
- `pkg/course/types.go` (100 lines)
- `migrations/003_courses_schema.sql` (150 lines)
- Schema validation tests

### Day 2: Service Layer (3-4 hours)
- [ ] Implement CourseService
- [ ] Implement CRUD operations
- [ ] Implement course-recording linking
- [ ] Implement enrollment logic
- [ ] Add permission checking
- [ ] Write unit tests

**Deliverables:**
- `pkg/course/service.go` (350+ lines)
- `pkg/course/service_test.go` (200+ lines)
- All core business logic

### Day 3: HTTP Handlers (3-4 hours)
- [ ] Create course handlers (CRUD)
- [ ] Create enrollment handlers
- [ ] Create permission handlers
- [ ] Create analytics handlers
- [ ] Add request validation
- [ ] Add error handling

**Deliverables:**
- `pkg/course/handlers.go` (500+ lines)
- All 26 endpoints implemented
- Request validation

### Day 4: Advanced Features (3-4 hours)
- [ ] Implement permission system
- [ ] Implement activity logging
- [ ] Implement analytics & reporting
- [ ] Add bulk import/export
- [ ] Add access logging

**Deliverables:**
- `pkg/course/permissions.go` (200+ lines)
- `pkg/course/activity_log.go` (200+ lines)
- `pkg/course/analytics.go` (300+ lines)
- Bulk operations support

### Day 5: Integration & Testing (3-4 hours)
- [ ] Integrate into main.go
- [ ] End-to-end testing
- [ ] Performance optimization
- [ ] Documentation
- [ ] Deployment verification

**Deliverables:**
- Updated `cmd/main.go`
- E2E tests
- Complete documentation
- Production binary

---

## 🚀 Quick Start (Next 30 Minutes)

### Step 1: Database Migration
```sql
-- Create courses table
psql -U postgres -d vtp_db -c "
CREATE TABLE courses (
    id UUID PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    instructor_id UUID NOT NULL REFERENCES users(id),
    department VARCHAR(100),
    semester VARCHAR(20),
    year INT,
    status VARCHAR(20) DEFAULT 'active',
    max_students INT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);"

CREATE INDEX idx_courses_instructor ON courses(instructor_id);
```

### Step 2: Start Day 1 Implementation
```bash
# Create new file
touch pkg/course/types.go

# Begin with type definitions
# See below for starter code
```

### Step 3: Verify Schema
```bash
# Verify tables created
psql -U postgres -d vtp_db -dt

# Expected: courses, course_recordings, course_enrollments, etc.
```

---

## 💻 Starter Code (Day 1)

### `pkg/course/types.go` Template

```go
package course

import (
	"time"

	"github.com/google/uuid"
)

// Course represents an educational course
type Course struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Code        string    `db:"code" json:"code"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	InstructorID uuid.UUID `db:"instructor_id" json:"instructor_id"`
	Department  string    `db:"department" json:"department"`
	Semester    string    `db:"semester" json:"semester"`
	Year        int       `db:"year" json:"year"`
	Status      string    `db:"status" json:"status"`
	MaxStudents int       `db:"max_students" json:"max_students"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// CourseEnrollment represents a student's enrollment in a course
type CourseEnrollment struct {
	ID             uuid.UUID `db:"id" json:"id"`
	CourseID       uuid.UUID `db:"course_id" json:"course_id"`
	StudentID      uuid.UUID `db:"student_id" json:"student_id"`
	EnrollmentDate time.Time `db:"enrollment_date" json:"enrollment_date"`
	Status         string    `db:"status" json:"status"`
}

// CoursePermission represents user permissions for a course
type CoursePermission struct {
	ID       uuid.UUID `db:"id" json:"id"`
	CourseID uuid.UUID `db:"course_id" json:"course_id"`
	UserID   uuid.UUID `db:"user_id" json:"user_id"`
	Role     string    `db:"role" json:"role"` // instructor, ta, viewer
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// CourseRecording links a recording to a course
type CourseRecording struct {
	ID            uuid.UUID `db:"id" json:"id"`
	CourseID      uuid.UUID `db:"course_id" json:"course_id"`
	RecordingID   uuid.UUID `db:"recording_id" json:"recording_id"`
	LectureNumber int       `db:"lecture_number" json:"lecture_number"`
	LectureTitle  string    `db:"lecture_title" json:"lecture_title"`
	SequenceOrder int       `db:"sequence_order" json:"sequence_order"`
	IsPublished   bool      `db:"is_published" json:"is_published"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

// Request/Response Types
type CreateCourseRequest struct {
	Code        string `json:"code" validate:"required,max=50"`
	Name        string `json:"name" validate:"required,max=255"`
	Description string `json:"description"`
	Department  string `json:"department"`
	Semester    string `json:"semester"`
	Year        int    `json:"year"`
	MaxStudents int    `json:"max_students"`
}

type CreateCourseResponse struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
```

---

## 🔐 Permission Model

**Roles & Permissions:**

| Role | Course CRUD | Manage Students | View Recordings | Edit Recordings | Publish |
|------|------------|------------------|-----------------|-----------------|---------|
| Admin | ✅ | ✅ | ✅ | ✅ | ✅ |
| Instructor | Own Only | ✅ | ✅ | ✅ | ✅ |
| TA | Read Only | ✅ | ✅ | Limited | ❌ |
| Student | Read Only | ❌ | Enrolled | ❌ | ❌ |
| Viewer | ❌ | ❌ | Public | ❌ | ❌ |

---

## 📊 Data Flow Diagram

```
User (Student/Instructor/Admin)
         ↓
    Authentication (Phase 1a) ✅
         ↓
    Authorization Check (Phase 3 NEW)
         ↓
    Course Service (Phase 3 NEW)
         ├─→ Course CRUD
         ├─→ Enrollment Management
         ├─→ Permission System
         └─→ Activity Logging
         ↓
    Recording Service (Phase 2a) ✅
         ├─→ Recording Control
         ├─→ Storage Management
         └─→ Streaming/Playback
         ↓
    PostgreSQL Database
```

---

## 🧪 Testing Strategy

**Unit Tests (Per Component):**
- Service layer tests (business logic)
- Handler tests (HTTP logic)
- Permission tests (access control)
- Integration tests (service interactions)

**E2E Tests (Full Workflow):**
```
1. Create course
2. Enroll students
3. Upload recording
4. Link to course
5. Students access recording
6. Verify audit log
```

**Load Tests:**
- 100 courses
- 1,000 students
- 500 recordings

---

## 📈 Success Criteria

### Functional Requirements
- ✅ All CRUD operations working
- ✅ Permission system enforced
- ✅ Enrollment management operational
- ✅ Analytics capturing data
- ✅ Audit logging comprehensive

### Non-Functional Requirements
- ✅ Response time < 200ms
- ✅ Memory usage < 100MB
- ✅ Database queries < 50ms
- ✅ Supports 1,000+ concurrent users
- ✅ Build time < 5 seconds

### Documentation Requirements
- ✅ API documentation (26 endpoints)
- ✅ Database schema documented
- ✅ Permission model documented
- ✅ Integration guide
- ✅ Troubleshooting guide

---

## ⚠️ Important Considerations

### Security
- Implement strict permission checks on every endpoint
- Validate user role before course operations
- Log all administrative actions
- Sanitize input (SQL injection prevention)
- Use parameterized queries

### Performance
- Index frequently queried columns
- Use database connection pooling
- Cache permission lookups (5-min TTL)
- Batch student enrollment operations
- Consider read replicas for reporting

### Scalability
- Design for multi-tenant scenarios
- Prepare for horizontal scaling
- Use event-based activity logging
- Design for archive/cleanup operations
- Consider message queues for bulk operations

---

## 🎯 Next Actions

### Immediate (Next 30 minutes)
1. ✅ Run database migration for Phase 3
2. ✅ Create `pkg/course/` directory
3. ✅ Create `pkg/course/types.go` with starter code
4. ✅ Create `migrations/003_courses_schema.sql`

### Today (Next 2-3 hours)
1. ✅ Complete Day 1: Types and schema
2. ⏳ Begin Day 2: Service layer
3. ⏳ Run initial tests

### This Week
1. ✅ Days 1-2: Core implementation
2. ⏳ Days 3-4: Handlers and features
3. ⏳ Day 5: Integration and testing
4. ⏳ Production deployment

---

## 📚 Reference Documents

**Phase 3 Documentation (To Be Created):**
- `PHASE_3_DAY_1_SCHEMA.md` - Database design details
- `PHASE_3_DAY_1_TYPES.md` - Type system documentation
- `PHASE_3_API_REFERENCE.md` - All 26 endpoints with examples
- `PHASE_3_PERMISSIONS_GUIDE.md` - Permission system explained
- `PHASE_3_INTEGRATION_GUIDE.md` - Integration into main.go

**Related Phase 2A Documents (Already Completed):**
- `PHASE_2A_MASTER_SUMMARY.md` - Recording system overview
- `PHASE_2A_DAY_4_API_REFERENCE.md` - Recording endpoints

---

## 📞 Support

**Questions About:**
- **Database Design** → Check `migrations/003_courses_schema.sql`
- **API Usage** → See endpoint reference section above
- **Permission System** → See permission model section
- **Integration** → Check integration guide (coming Day 5)
- **Types & Structure** → See starter code above

---

## ✅ Deployment Checklist (End of Phase 3)

- [ ] All 26 endpoints implemented
- [ ] All tests passing (unit + E2E)
- [ ] Permission system enforced
- [ ] Audit logging complete
- [ ] Analytics working
- [ ] Documentation complete
- [ ] Production binary created
- [ ] Integration verified
- [ ] Performance tested
- [ ] Security audit passed
- [ ] Deployment guide created

---

## 📊 Phase Progress

```
Phase 1a: Authentication      ✅ COMPLETE (100%)
Phase 1b: WebRTC Signaling    ✅ COMPLETE (100%)
Phase 2a: Recording System    ✅ COMPLETE (100%)
Phase 3: Course Management   🎯 STARTING NOW (0%)

Overall VTP Platform: 75% COMPLETE (After Phase 3: 100%)
```

---

**Phase 3 is ready to begin!**

Start with Day 1: Database Schema & Types

Estimated time to Phase 3 completion: 5-6 days  
Expected total code additions: 3,500+ lines

Ready to proceed? 🚀
