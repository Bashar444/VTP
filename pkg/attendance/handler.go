package attendance

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Bashar444/VTP/pkg/models"
	"github.com/Bashar444/VTP/pkg/utils"
)

// Handler handles HTTP requests for attendance
type Handler struct {
	service *Service
}

// NewHandler creates a new attendance handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RecordAttendanceRequest represents the request to record attendance
type RecordAttendanceRequest struct {
	StudentID      string `json:"student_id"`
	ClassSectionID string `json:"class_section_id"`
	MeetingID      string `json:"meeting_id,omitempty"`
	SubjectID      string `json:"subject_id,omitempty"`
	Date           string `json:"date"` // YYYY-MM-DD
	Status         string `json:"status"`
	Notes          string `json:"notes,omitempty"`
}

// BulkAttendanceRequest represents bulk attendance recording
type BulkAttendanceRequest struct {
	ClassSectionID string `json:"class_section_id"`
	SubjectID      string `json:"subject_id,omitempty"`
	MeetingID      string `json:"meeting_id,omitempty"`
	Date           string `json:"date"`
	Records        []struct {
		StudentID string `json:"student_id"`
		Status    string `json:"status"`
		Notes     string `json:"notes,omitempty"`
	} `json:"records"`
}

// RecordAttendance handles POST /api/v1/attendance
func (h *Handler) RecordAttendance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteErr(w, http.StatusMethodNotAllowed, nil)
		return
	}

	var req RecordAttendanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErr(w, http.StatusBadRequest, err)
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		date = time.Now()
	}

	a := &models.Attendance{
		StudentID:      req.StudentID,
		ClassSectionID: req.ClassSectionID,
		Status:         req.Status,
		Date:           date,
		Notes:          req.Notes,
	}
	if req.MeetingID != "" {
		a.MeetingID = &req.MeetingID
	}
	if req.SubjectID != "" {
		a.SubjectID = &req.SubjectID
	}

	if err := h.service.RecordAttendance(r.Context(), a); err != nil {
		status := http.StatusBadRequest
		if err == ErrAttendanceNotFound {
			status = http.StatusNotFound
		}
		utils.WriteErr(w, status, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, a)
}

// BulkRecordAttendance handles POST /api/v1/attendance/bulk
func (h *Handler) BulkRecordAttendance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteErr(w, http.StatusMethodNotAllowed, nil)
		return
	}

	var req BulkAttendanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErr(w, http.StatusBadRequest, err)
		return
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		date = time.Now()
	}

	var records []models.Attendance
	for _, rec := range req.Records {
		a := models.Attendance{
			StudentID:      rec.StudentID,
			ClassSectionID: req.ClassSectionID,
			Status:         rec.Status,
			Date:           date,
			Notes:          rec.Notes,
		}
		if req.MeetingID != "" {
			a.MeetingID = &req.MeetingID
		}
		if req.SubjectID != "" {
			a.SubjectID = &req.SubjectID
		}
		records = append(records, a)
	}

	if err := h.service.BulkRecordAttendance(r.Context(), records); err != nil {
		utils.WriteErr(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Attendance recorded successfully",
		"count":   len(records),
	})
}

// GetStudentAttendance handles GET /api/v1/attendance/student/{id}
func (h *Handler) GetStudentAttendance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErr(w, http.StatusMethodNotAllowed, nil)
		return
	}

	studentID := utils.Param(r, "id")
	if studentID == "" {
		utils.WriteErr(w, http.StatusBadRequest, ErrStudentIDRequired)
		return
	}

	// Parse date range from query params
	startStr := r.URL.Query().Get("start_date")
	endStr := r.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	var err error

	if startStr != "" {
		startDate, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			utils.WriteErr(w, http.StatusBadRequest, ErrInvalidDateRange)
			return
		}
	} else {
		startDate = time.Now().AddDate(0, -1, 0) // Default: last month
	}

	if endStr != "" {
		endDate, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			utils.WriteErr(w, http.StatusBadRequest, ErrInvalidDateRange)
			return
		}
	} else {
		endDate = time.Now()
	}

	records, err := h.service.GetStudentAttendance(r.Context(), studentID, startDate, endDate)
	if err != nil {
		utils.WriteErr(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"student_id": studentID,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
		"records":    records,
	})
}

// GetStudentStats handles GET /api/v1/attendance/student/{id}/stats
func (h *Handler) GetStudentStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErr(w, http.StatusMethodNotAllowed, nil)
		return
	}

	studentID := utils.Param(r, "id")
	if studentID == "" {
		utils.WriteErr(w, http.StatusBadRequest, ErrStudentIDRequired)
		return
	}

	startStr := r.URL.Query().Get("start_date")
	endStr := r.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	var err error

	if startStr != "" {
		startDate, _ = time.Parse("2006-01-02", startStr)
	} else {
		startDate = time.Now().AddDate(0, -3, 0) // Default: last 3 months
	}

	if endStr != "" {
		endDate, _ = time.Parse("2006-01-02", endStr)
	} else {
		endDate = time.Now()
	}

	stats, err := h.service.GetStudentStats(r.Context(), studentID, startDate, endDate)
	if err != nil {
		utils.WriteErr(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, stats)
}

// GetClassAttendance handles GET /api/v1/attendance/class/{id}
func (h *Handler) GetClassAttendance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErr(w, http.StatusMethodNotAllowed, nil)
		return
	}

	classSectionID := utils.Param(r, "id")
	dateStr := r.URL.Query().Get("date")

	var date time.Time
	if dateStr != "" {
		date, _ = time.Parse("2006-01-02", dateStr)
	} else {
		date = time.Now()
	}

	records, err := h.service.GetClassAttendance(r.Context(), classSectionID, date)
	if err != nil {
		utils.WriteErr(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"class_section_id": classSectionID,
		"date":             date.Format("2006-01-02"),
		"records":          records,
	})
}

// GetMeetingAttendance handles GET /api/v1/attendance/meeting/{id}
func (h *Handler) GetMeetingAttendance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErr(w, http.StatusMethodNotAllowed, nil)
		return
	}

	meetingID := utils.Param(r, "id")
	records, err := h.service.GetMeetingAttendance(r.Context(), meetingID)
	if err != nil {
		utils.WriteErr(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"meeting_id": meetingID,
		"records":    records,
	})
}

// UpdateAttendance handles PUT /api/v1/attendance/{id}
func (h *Handler) UpdateAttendance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.WriteErr(w, http.StatusMethodNotAllowed, nil)
		return
	}

	id := utils.Param(r, "id")
	var req struct {
		Status string `json:"status"`
		Notes  string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErr(w, http.StatusBadRequest, err)
		return
	}

	existing, err := h.service.GetAttendance(r.Context(), id)
	if err != nil {
		utils.WriteErr(w, http.StatusNotFound, err)
		return
	}

	existing.Status = req.Status
	existing.Notes = req.Notes

	if err := h.service.UpdateAttendance(r.Context(), existing); err != nil {
		utils.WriteErr(w, http.StatusBadRequest, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, existing)
}

// GenerateReport handles GET /api/v1/attendance/report
func (h *Handler) GenerateReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteErr(w, http.StatusMethodNotAllowed, nil)
		return
	}

	studentID := r.URL.Query().Get("student_id")
	if studentID == "" {
		utils.WriteErr(w, http.StatusBadRequest, ErrStudentIDRequired)
		return
	}

	startStr := r.URL.Query().Get("start_date")
	endStr := r.URL.Query().Get("end_date")
	includeRecords := r.URL.Query().Get("include_records") == "true"

	var startDate, endDate time.Time
	if startStr != "" {
		startDate, _ = time.Parse("2006-01-02", startStr)
	} else {
		startDate = time.Now().AddDate(0, -3, 0)
	}
	if endStr != "" {
		endDate, _ = time.Parse("2006-01-02", endStr)
	} else {
		endDate = time.Now()
	}

	report, err := h.service.GenerateReport(r.Context(), studentID, startDate, endDate, includeRecords)
	if err != nil {
		utils.WriteErr(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, report)
}
