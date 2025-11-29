package course

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// CourseHandlers manages HTTP handlers for course operations
type CourseHandlers struct {
	service *CourseService
	logger  *log.Logger
}

// NewCourseHandlers creates a new course handlers instance
func NewCourseHandlers(service *CourseService, logger *log.Logger) *CourseHandlers {
	if logger == nil {
		logger = log.New(log.Writer(), "[CourseAPI] ", log.LstdFlags)
	}
	return &CourseHandlers{
		service: service,
		logger:  logger,
	}
}

// RegisterCourseRoutes registers all course routes on the default HTTP mux
func (ch *CourseHandlers) RegisterCourseRoutes(mux *http.ServeMux) {
	// Course Management
	mux.HandleFunc("/api/v1/courses", ch.handleCourses)
	mux.HandleFunc("/api/v1/courses/", ch.handleCourseDetail)

	ch.logger.Println("Course routes registered")
}

// handleCourses routes to appropriate handler based on method
func (ch *CourseHandlers) handleCourses(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		ch.CreateCourse(w, r)
	case http.MethodGet:
		ch.ListCourses(w, r)
	default:
		ch.respondError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
	}
}

// handleCourseDetail routes course detail requests
func (ch *CourseHandlers) handleCourseDetail(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Extract course ID from path: /api/v1/courses/{id} or related
	parts := strings.Split(path, "/")
	if len(parts) < 5 {
		ch.respondError(w, http.StatusNotFound, "Not found", nil)
		return
	}

	courseIDStr := parts[4]
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		ch.respondError(w, http.StatusBadRequest, "Invalid course ID", err)
		return
	}

	// Route based on remaining path and method
	if len(parts) == 5 {
		// /api/v1/courses/{id}
		switch r.Method {
		case http.MethodGet:
			ch.GetCourse(w, r, courseID)
		case http.MethodPut:
			ch.UpdateCourse(w, r, courseID)
		case http.MethodDelete:
			ch.DeleteCourse(w, r, courseID)
		default:
			ch.respondError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		}
	} else if len(parts) >= 6 {
		// /api/v1/courses/{id}/{action}
		action := parts[5]
		switch action {
		case "enroll":
			if len(parts) == 6 {
				switch r.Method {
				case http.MethodPost:
					ch.EnrollStudent(w, r, courseID)
				case http.MethodGet:
					ch.ListEnrollments(w, r, courseID)
				default:
					ch.respondError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
				}
			} else if len(parts) == 7 && r.Method == http.MethodDelete {
				// /api/v1/courses/{id}/enroll/{student_id}
				studentID, err := uuid.Parse(parts[6])
				if err != nil {
					ch.respondError(w, http.StatusBadRequest, "Invalid student ID", err)
					return
				}
				ch.RemoveStudent(w, r, courseID, studentID)
			}
		case "recordings":
			if r.Method == http.MethodPost {
				ch.AddRecording(w, r, courseID)
			} else if len(parts) == 7 && parts[6] == "publish" && r.Method == http.MethodPost {
				recordingID, err := uuid.Parse(parts[6])
				if err != nil {
					ch.respondError(w, http.StatusBadRequest, "Invalid recording ID", err)
					return
				}
				ch.PublishRecording(w, r, courseID, recordingID)
			}
		case "enrollments":
			if r.Method == http.MethodGet {
				ch.ListEnrollments(w, r, courseID)
			}
		case "permissions":
			if r.Method == http.MethodPost {
				ch.SetPermission(w, r, courseID)
			} else if len(parts) == 7 && r.Method == http.MethodGet {
				userID, err := uuid.Parse(parts[6])
				if err != nil {
					ch.respondError(w, http.StatusBadRequest, "Invalid user ID", err)
					return
				}
				ch.GetPermission(w, r, courseID, userID)
			}
		case "stats":
			if r.Method == http.MethodGet {
				ch.GetCourseStats(w, r, courseID)
			}
		default:
			ch.respondError(w, http.StatusNotFound, "Not found", nil)
		}
	}
}

// CreateCourse creates a new course (POST /api/v1/courses)
func (ch *CourseHandlers) CreateCourse(w http.ResponseWriter, r *http.Request) {
	var req CreateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ch.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get instructor ID from JWT token context
	instructorID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		ch.respondError(w, http.StatusUnauthorized, "User ID not found in context", nil)
		return
	}

	course, err := ch.service.CreateCourse(r.Context(), req, instructorID)
	if err != nil {
		ch.respondError(w, http.StatusInternalServerError, "Failed to create course", err)
		return
	}

	response := CreateCourseResponse{
		ID:        course.ID,
		Code:      course.Code,
		Name:      course.Name,
		Status:    course.Status,
		CreatedAt: course.CreatedAt,
	}

	ch.respondJSON(w, http.StatusCreated, response)
}

// ListCourses lists courses with optional filtering (GET /api/v1/courses)
func (ch *CourseHandlers) ListCourses(w http.ResponseWriter, r *http.Request) {
	filter := make(map[string]interface{})

	// Parse query parameters
	if semester := r.URL.Query().Get("semester"); semester != "" {
		filter["semester"] = semester
	}
	if year := r.URL.Query().Get("year"); year != "" {
		if yearInt, err := strconv.Atoi(year); err == nil {
			filter["year"] = yearInt
		}
	}
	if instructorID := r.URL.Query().Get("instructor_id"); instructorID != "" {
		if id, err := uuid.Parse(instructorID); err == nil {
			filter["instructor_id"] = id
		}
	}
	if status := r.URL.Query().Get("status"); status != "" {
		filter["status"] = status
	}

	courses, err := ch.service.ListCourses(r.Context(), filter)
	if err != nil {
		ch.respondError(w, http.StatusInternalServerError, "Failed to list courses", err)
		return
	}

	responses := make([]CourseListResponse, len(courses))
	for i, course := range courses {
		enrollments, _ := ch.service.ListEnrollments(r.Context(), course.ID)
		responses[i] = CourseListResponse{
			ID:            course.ID,
			Code:          course.Code,
			Name:          course.Name,
			Semester:      course.Semester,
			Year:          course.Year,
			Status:        course.Status,
			EnrolledCount: len(enrollments),
			CreatedAt:     course.CreatedAt,
		}
	}

	ch.respondJSON(w, http.StatusOK, responses)
}

// GetCourse gets a specific course (GET /api/v1/courses/{id})
func (ch *CourseHandlers) GetCourse(w http.ResponseWriter, r *http.Request, courseID uuid.UUID) {
	course, err := ch.service.GetCourse(r.Context(), courseID)
	if err != nil {
		ch.respondError(w, http.StatusNotFound, "Course not found", err)
		return
	}

	enrollments, _ := ch.service.ListEnrollments(r.Context(), courseID)

	response := CourseDetailResponse{
		ID:             course.ID,
		Code:           course.Code,
		Name:           course.Name,
		Description:    course.Description,
		InstructorID:   course.InstructorID,
		Department:     course.Department,
		Semester:       course.Semester,
		Year:           course.Year,
		Status:         course.Status,
		MaxStudents:    course.MaxStudents,
		EnrolledCount:  len(enrollments),
		RecordingCount: 0,
		CreatedAt:      course.CreatedAt,
		UpdatedAt:      course.UpdatedAt,
	}

	ch.respondJSON(w, http.StatusOK, response)
}

// UpdateCourse updates a course (PUT /api/v1/courses/{id})
func (ch *CourseHandlers) UpdateCourse(w http.ResponseWriter, r *http.Request, courseID uuid.UUID) {
	var req UpdateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ch.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	course, err := ch.service.UpdateCourse(r.Context(), courseID, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			ch.respondError(w, http.StatusNotFound, "Course not found", err)
		} else {
			ch.respondError(w, http.StatusInternalServerError, "Failed to update course", err)
		}
		return
	}

	ch.respondJSON(w, http.StatusOK, course)
}

// DeleteCourse deletes a course (DELETE /api/v1/courses/{id})
func (ch *CourseHandlers) DeleteCourse(w http.ResponseWriter, r *http.Request, courseID uuid.UUID) {
	err := ch.service.DeleteCourse(r.Context(), courseID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			ch.respondError(w, http.StatusNotFound, "Course not found", err)
		} else {
			ch.respondError(w, http.StatusInternalServerError, "Failed to delete course", err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// EnrollStudent enrolls a student in a course (POST /api/v1/courses/{id}/enroll)
func (ch *CourseHandlers) EnrollStudent(w http.ResponseWriter, r *http.Request, courseID uuid.UUID) {
	var req EnrollStudentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ch.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	enrollment, err := ch.service.EnrollStudent(r.Context(), courseID, req.StudentID)
	if err != nil {
		ch.respondError(w, http.StatusInternalServerError, "Failed to enroll student", err)
		return
	}

	ch.respondJSON(w, http.StatusCreated, enrollment)
}

// ListEnrollments lists course enrollments (GET /api/v1/courses/{id}/enrollments)
func (ch *CourseHandlers) ListEnrollments(w http.ResponseWriter, r *http.Request, courseID uuid.UUID) {
	enrollments, err := ch.service.ListEnrollments(r.Context(), courseID)
	if err != nil {
		ch.respondError(w, http.StatusInternalServerError, "Failed to list enrollments", err)
		return
	}

	responses := make([]EnrollmentListResponse, len(enrollments))
	for i, enrollment := range enrollments {
		responses[i] = EnrollmentListResponse{
			ID:             enrollment.ID,
			StudentID:      enrollment.StudentID,
			EnrollmentDate: enrollment.EnrollmentDate,
			Status:         enrollment.Status,
		}
	}

	ch.respondJSON(w, http.StatusOK, responses)
}

// RemoveStudent removes a student from a course (DELETE /api/v1/courses/{id}/enroll/{student_id})
func (ch *CourseHandlers) RemoveStudent(w http.ResponseWriter, r *http.Request, courseID uuid.UUID, studentID uuid.UUID) {
	err := ch.service.RemoveStudent(r.Context(), courseID, studentID)
	if err != nil {
		ch.respondError(w, http.StatusInternalServerError, "Failed to remove student", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddRecording adds a recording to a course (POST /api/v1/courses/{id}/recordings)
func (ch *CourseHandlers) AddRecording(w http.ResponseWriter, r *http.Request, courseID uuid.UUID) {
	var req AddRecordingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ch.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	lectureNumber := 0
	if req.LectureNumber != nil {
		lectureNumber = *req.LectureNumber
	}

	lectureTitle := ""
	if req.LectureTitle != nil {
		lectureTitle = *req.LectureTitle
	}

	recording, err := ch.service.AddRecordingToCourse(r.Context(), courseID, req.RecordingID,
		lectureNumber, lectureTitle, req.SequenceOrder)
	if err != nil {
		ch.respondError(w, http.StatusInternalServerError, "Failed to add recording", err)
		return
	}

	ch.respondJSON(w, http.StatusCreated, recording)
}

// PublishRecording publishes a recording in a course (POST /api/v1/courses/{id}/recordings/{recording_id}/publish)
func (ch *CourseHandlers) PublishRecording(w http.ResponseWriter, r *http.Request, courseID uuid.UUID, recordingID uuid.UUID) {
	err := ch.service.PublishCourseRecording(r.Context(), courseID, recordingID)
	if err != nil {
		ch.respondError(w, http.StatusInternalServerError, "Failed to publish recording", err)
		return
	}

	ch.respondJSON(w, http.StatusOK, map[string]string{"message": "Recording published"})
}

// SetPermission sets user permission for a course (POST /api/v1/courses/{id}/permissions)
func (ch *CourseHandlers) SetPermission(w http.ResponseWriter, r *http.Request, courseID uuid.UUID) {
	var req SetPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ch.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	permission, err := ch.service.SetPermission(r.Context(), courseID, req.UserID, req.Role)
	if err != nil {
		ch.respondError(w, http.StatusInternalServerError, "Failed to set permission", err)
		return
	}

	ch.respondJSON(w, http.StatusCreated, permission)
}

// GetPermission gets user permission for a course (GET /api/v1/courses/{id}/permissions/{user_id})
func (ch *CourseHandlers) GetPermission(w http.ResponseWriter, r *http.Request, courseID uuid.UUID, userID uuid.UUID) {
	permission, err := ch.service.GetPermission(r.Context(), courseID, userID)
	if err != nil {
		ch.respondError(w, http.StatusInternalServerError, "Failed to get permission", err)
		return
	}

	if permission == nil {
		ch.respondJSON(w, http.StatusOK, PermissionCheckResponse{
			Allowed: false,
			Message: "No permission found",
		})
		return
	}

	ch.respondJSON(w, http.StatusOK, PermissionCheckResponse{
		Allowed: true,
		Role:    permission.Role,
		Message: "Permission granted",
	})
}

// GetCourseStats gets course statistics (GET /api/v1/courses/{id}/stats)
func (ch *CourseHandlers) GetCourseStats(w http.ResponseWriter, r *http.Request, courseID uuid.UUID) {
	course, err := ch.service.GetCourse(r.Context(), courseID)
	if err != nil {
		ch.respondError(w, http.StatusNotFound, "Course not found", err)
		return
	}

	enrollments, _ := ch.service.ListEnrollments(r.Context(), courseID)

	// Get stats from database
	query := `
		SELECT 
			COUNT(DISTINCT cr.recording_id) as total_recordings,
			COUNT(DISTINCT CASE WHEN cr.is_published = true THEN cr.recording_id END) as published_recordings,
			COUNT(DISTINCT ral.user_id) as unique_viewers,
			COUNT(DISTINCT ral.id) as total_views
		FROM courses c
		LEFT JOIN course_recordings cr ON c.id = cr.course_id
		LEFT JOIN recording_access_logs ral ON cr.recording_id = ral.recording_id
		WHERE c.id = $1
	`

	var totalRecordings, publishedRecordings, uniqueViewers, totalViews sql.NullInt64

	err = ch.service.db.QueryRowContext(r.Context(), query, courseID).Scan(
		&totalRecordings,
		&publishedRecordings,
		&uniqueViewers,
		&totalViews,
	)

	if err != nil && err != sql.ErrNoRows {
		ch.respondError(w, http.StatusInternalServerError, "Failed to get stats", err)
		return
	}

	totalRec := int64(0)
	if totalRecordings.Valid {
		totalRec = totalRecordings.Int64
	}

	pubRec := int64(0)
	if publishedRecordings.Valid {
		pubRec = publishedRecordings.Int64
	}

	viewers := int64(0)
	if uniqueViewers.Valid {
		viewers = uniqueViewers.Int64
	}

	views := int64(0)
	if totalViews.Valid {
		views = totalViews.Int64
	}

	response := CourseStatsResponse{
		CourseID:            course.ID,
		CourseName:          course.Name,
		TotalStudents:       len(enrollments),
		TotalRecordings:     int(totalRec),
		PublishedRecordings: int(pubRec),
		UniqueViewers:       int(viewers),
		TotalViews:          int(views),
		AverageEngagement:   0,
	}

	ch.respondJSON(w, http.StatusOK, response)
}

// Helper functions

func (ch *CourseHandlers) respondJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (ch *CourseHandlers) respondError(w http.ResponseWriter, code int, message string, err error) {
	if err != nil {
		ch.logger.Printf("API Error: %s - %v", message, err)
	}

	response := ErrorResponse{
		Error:      message,
		StatusCode: code,
		Message:    fmt.Sprintf("%v", err),
		Timestamp:  sql.NullTime{}.Time,
	}

	ch.respondJSON(w, code, response)
}
