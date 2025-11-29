# 🚀 Phase 3 - Next Steps & Day 2 Quick Start

**Current Status:** Phase 3 Day 1 Complete ✅  
**Date:** November 24, 2025  
**Next Action:** Phase 3 Day 2 Implementation  
**Estimated Time:** 3-4 hours  
**Expected Output:** 400+ lines of code, 26 API endpoints  

---

## 📋 Quick Decision Matrix

### Option A: Continue Immediately (Recommended) ⭐
**Start Phase 3 Day 2 Now**
- Implement HTTP handlers
- Create 26 API endpoints
- Add validation and error handling
- Time: 3-4 hours
- Output: Working API layer

**Next Command:**
```bash
# Run database migration first
psql -U postgres -d vtp_db -f migrations/003_courses_schema.sql

# Create handlers file
New-Item -Path pkg\course\handlers.go
```

### Option B: Production Deployment
**Deploy Phase 2A First**
- Start vtp-platform.exe binary
- Test all 15 recording endpoints
- Verify production readiness
- Time: 30 minutes

**Command:**
```bash
./vtp-platform.exe
```

### Option C: Testing & Verification
**Verify Phase 3 Foundation**
- Test database connection
- Verify schema created
- Check service methods
- Time: 20 minutes

---

## ✅ Pre-Day 2 Verification Checklist

Before starting Day 2 implementation, verify:

```bash
# 1. Check database tables exist
psql -U postgres -d vtp_db -dt
# Expected: courses, course_recordings, course_enrollments, etc.

# 2. Verify indexes created
psql -U postgres -d vtp_db -di
# Expected: 15+ course-related indexes

# 3. Check views exist
psql -U postgres -d vtp_db -dv
# Expected: active_courses, course_dashboard_stats

# 4. Verify code compiles
go build ./pkg/course

# 5. Check service package loads
go test ./pkg/course -v
```

---

## 📁 Phase 3 Day 2: Implementation Plan

### What You'll Create Today

**1. HTTP Handlers (pkg/course/handlers.go)**
- 26 endpoint handlers
- Request validation
- Error handling
- Response formatting

**2. Permission Middleware (pkg/course/permissions.go)**
- Role-based access control
- Permission checking
- Authorization logic

**3. Analytics (pkg/course/analytics.go)**
- Statistics calculations
- Engagement metrics
- Report generation

**4. Unit Tests**
- Service tests
- Handler tests
- Permission tests

---

## 🎯 Phase 3 Day 2 Detailed Steps

### Step 1: Create Handlers File (1-1.5 hours)

**Create:** `pkg/course/handlers.go`

**Structure:**
```go
package course

import (
    "net/http"
    "github.com/google/uuid"
    "log"
    "database/sql"
)

type CourseHandler struct {
    service *CourseService
    logger  *log.Logger
}

func NewCourseHandler(service *CourseService, logger *log.Logger) *CourseHandler {
    return &CourseHandler{
        service: service,
        logger:  logger,
    }
}

// 26 handlers to implement...
```

**Handlers to Implement:**
1. CreateCourseHandler - POST /api/v1/courses
2. ListCoursesHandler - GET /api/v1/courses
3. GetCourseHandler - GET /api/v1/courses/{id}
4. UpdateCourseHandler - PUT /api/v1/courses/{id}
5. DeleteCourseHandler - DELETE /api/v1/courses/{id}
6. ListCourseRecordingsHandler - GET /api/v1/courses/{id}/recordings
7. AddRecordingHandler - POST /api/v1/courses/{id}/recordings
8. RemoveRecordingHandler - DELETE /api/v1/courses/{id}/recordings/{rid}
9. PublishCourseHandler - POST /api/v1/courses/{id}/publish
10. GetCourseStatsHandler - GET /api/v1/courses/{id}/stats
11. EnrollStudentHandler - POST /api/v1/courses/{id}/enroll
12. RemoveEnrollmentHandler - DELETE /api/v1/courses/{id}/enroll/{sid}
13. ListEnrollmentsHandler - GET /api/v1/courses/{id}/students
14. CheckEnrollmentHandler - GET /api/v1/courses/{id}/enrollment
15. ListUserCoursesHandler - GET /api/v1/enrollments
16. SetPermissionHandler - POST /api/v1/courses/{id}/permissions
17. ListPermissionsHandler - GET /api/v1/courses/{id}/permissions
18. RemovePermissionHandler - DELETE /api/v1/courses/{id}/permissions/{uid}
19. CheckAccessHandler - GET /api/v1/courses/{id}/access-check
20. GetAnalyticsHandler - GET /api/v1/courses/{id}/analytics
21. GetEngagementHandler - GET /api/v1/courses/{id}/engagement
22. GetAccessLogHandler - GET /api/v1/recordings/{id}/access-log
23. ExportReportHandler - GET /api/v1/courses/{id}/reports/export
24. BulkImportHandler - POST /api/v1/courses/bulk/import
25. BulkExportHandler - GET /api/v1/courses/bulk/export
26. InviteStudentHandler - POST /api/v1/courses/{id}/invite

**Code per Handler:** ~15-20 lines average  
**Total for All Handlers:** 400+ lines

### Step 2: Add to main.go (15-30 minutes)

**Modify:** `cmd/main.go`

**Add:**
```go
// Initialize course service
courseService := course.NewCourseService(db, logger)
courseHandler := course.NewCourseHandler(courseService, logger)

// Register course routes
mux.HandleFunc("POST /api/v1/courses", courseHandler.CreateCourseHandler)
mux.HandleFunc("GET /api/v1/courses", courseHandler.ListCoursesHandler)
// ... 24 more routes
```

### Step 3: Test Endpoints (30-45 minutes)

**Start Server:**
```bash
go run cmd/main.go
```

**Test Endpoints:**
```bash
# Test create course
curl -X POST http://localhost:8080/api/v1/courses \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "code":"CS101",
    "name":"Introduction to Computer Science",
    "semester":"Fall",
    "year":2025,
    "description":"Learn programming fundamentals"
  }'

# Test list courses
curl -X GET http://localhost:8080/api/v1/courses \
  -H "Authorization: Bearer TOKEN"

# Test enroll student
curl -X POST http://localhost:8080/api/v1/courses/{courseId}/enroll \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"student_id":"550e8400-e29b-41d4-a716-446655440000"}'
```

---

## 💻 Starter Template for handlers.go

### Basic Handler Structure

```go
package course

import (
    "encoding/json"
    "net/http"
    "github.com/google/uuid"
)

func (ch *CourseHandler) CreateCourseHandler(w http.ResponseWriter, r *http.Request) {
    // Extract user ID from auth context
    userID, ok := r.Context().Value("user_id").(uuid.UUID)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Parse request
    var req CreateCourseRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Call service
    course, err := ch.service.CreateCourse(r.Context(), req, userID)
    if err != nil {
        ch.logger.Printf("Error creating course: %v", err)
        http.Error(w, "Failed to create course", http.StatusInternalServerError)
        return
    }

    // Return response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(course)
}

// ... Similar pattern for other 25 handlers
```

---

## 🧪 Testing Strategy for Day 2

### Unit Tests
```go
func TestCreateCourse(t *testing.T) {
    // Setup database
    // Create service
    // Call CreateCourse
    // Verify result
    // Cleanup
}
```

### Integration Tests
```go
func TestCreateCourseWithEnrollment(t *testing.T) {
    // Create course
    // Enroll student
    // Verify enrollment
}
```

### E2E Tests
```bash
# Full workflow test
1. Create course
2. Enroll students
3. Add recordings
4. Publish
5. Access as student
6. Verify audit log
```

---

## 📊 Expected Output (End of Day 2)

**Files Created/Modified:**
- ✅ `pkg/course/handlers.go` (NEW) - 400+ lines
- ✅ `cmd/main.go` (MODIFIED) - Add course service init + routes

**Functionality Delivered:**
- ✅ All 26 endpoints implemented
- ✅ Request validation
- ✅ Error handling
- ✅ Logging
- ✅ Testable

**API Status:**
- Course CRUD: ✅
- Enrollment management: ✅
- Permission system: ✅
- Analytics ready: ✅

**Build Status:**
- Expected: ✅ CLEAN (0 errors)
- Tests: ✅ PASSING

---

## 🔧 Common Development Setup

### VS Code Extensions (Recommended)
- Go (golang.go)
- REST Client (humao.rest-client)
- Thunder Client (rangav.vscode-thunder-client)

### Debug Commands
```bash
# Build and run with logging
go run -v cmd/main.go

# Run specific handler test
go test ./pkg/course -run TestCreateCourse -v

# Profile memory
go tool pprof http://localhost:6060/debug/pprof/heap
```

---

## ⚠️ Common Issues & Solutions

### Issue: "Unauthorized" on endpoints
**Solution:** Include Authorization header with JWT token
```bash
curl -H "Authorization: Bearer TOKEN"
```

### Issue: Database connection error
**Solution:** Verify PostgreSQL is running and migration applied
```bash
psql -U postgres -d vtp_db -c "SELECT COUNT(*) FROM courses;"
```

### Issue: Handler not found (404)
**Solution:** Verify route registered in main.go
```bash
# Check mux registration
grep "api/v1/courses" cmd/main.go
```

### Issue: Type mismatch errors
**Solution:** Ensure types match between handler and service
- Check request type in handler
- Check response type in service
- Verify struct tags

---

## 📈 Progress Tracking

**Phase 3 Progress (By Day):**
```
Day 1 (Today):   Schema & Types     ✅ COMPLETE (20%)
Day 2 (Next):    Handlers & API     🎯 STARTING  (20%)
Day 3 (Future):  Permissions & Logs  ⏳ PLANNED   (20%)
Day 4 (Future):  Analytics & Bulk    ⏳ PLANNED   (20%)
Day 5 (Future):  Integration & Test  ⏳ PLANNED   (20%)

Total Phase 3:   Course Management  20% COMPLETE
```

---

## 🎯 Day 2 Success Criteria

✅ All 26 endpoints created  
✅ All handlers working  
✅ Request validation implemented  
✅ Error handling in place  
✅ Logging configured  
✅ Build succeeds  
✅ Tests passing  
✅ API documented  

---

## 🚀 Ready to Start?

### Immediate Actions (Next 10 minutes)

1. **Run database migration:**
```bash
psql -U postgres -d vtp_db -f migrations/003_courses_schema.sql
```

2. **Verify tables created:**
```bash
psql -U postgres -d vtp_db -dt
```

3. **Create handlers file:**
```bash
# Create empty file
New-Item -Path pkg\course\handlers.go -Force
```

4. **Start coding:**
```bash
# Open in VS Code
code pkg/course/handlers.go
```

---

## 📚 Reference Materials

**Use These Documents:**
- `PHASE_3_INITIALIZATION_PLAN.md` - Full API spec
- `PHASE_2A_DAY_4_API_REFERENCE.md` - Similar endpoint patterns
- `pkg/course/types.go` - Request/response types
- `pkg/course/service.go` - Service methods

---

## 💡 Pro Tips

1. **Copy patterns from Phase 2A** - Handler structure is similar
2. **Use validation tags** - Already defined in types
3. **Test each endpoint** - Use curl or REST Client
4. **Commit frequently** - Save progress often
5. **Use logging** - Log all operations for debugging

---

## 📞 Quick Reference

| Need | File | Location |
|------|------|----------|
| Handler template | PHASE_3_INITIALIZATION_PLAN.md | Starter Code section |
| API spec | PHASE_3_INITIALIZATION_PLAN.md | API Endpoints section |
| Type definitions | pkg/course/types.go | Request Types section |
| Service methods | pkg/course/service.go | All methods |

---

## ✨ Phase 3 Day 2 Summary

**What You'll Accomplish:**
- Build complete HTTP API layer
- 26 working endpoints
- Full CRUD operations
- Permission checks
- Error handling
- Production-ready handlers

**Time Required:** 3-4 hours  
**Difficulty:** Moderate (many endpoints, same pattern)  
**Outcome:** Working API for course management  

---

## 🎉 Phase 3 Day 2 Is Ready!

### Ready to Begin?

**Option 1: Start Immediately** (Recommended)
```bash
cd c:\Users\Admin\OneDrive\Desktop\VTP
psql -U postgres -d vtp_db -f migrations/003_courses_schema.sql
New-Item -Path pkg\course\handlers.go -Force
code pkg/course/handlers.go
```

**Option 2: Continue Tomorrow**
- Take a break now
- Review PHASE_3_INITIALIZATION_PLAN.md
- Start fresh in morning

---

**Current Phase 3 Status: Day 1 ✅ | Ready for Day 2 🚀**

Next delivery: 26 working API endpoints  
Expected time: 3-4 hours  
Expected output: 400+ lines of code  

Let's build the API layer! 🎯
