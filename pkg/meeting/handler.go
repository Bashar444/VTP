package meeting

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Bashar444/VTP/pkg/models"
	"github.com/gorilla/mux"
)

// Handler handles HTTP requests for meetings
type Handler struct {
	service *Service
}

// NewHandler creates a new meeting handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateMeetingRequest represents the request to create a meeting
type CreateMeetingRequest struct {
	InstructorID string    `json:"instructor_id"`
	StudentID    string    `json:"student_id"`
	SubjectID    string    `json:"subject_id"`
	TitleAr      string    `json:"title_ar"`
	ScheduledAt  time.Time `json:"scheduled_at"`
	Duration     int       `json:"duration"`
	MeetingURL   string    `json:"meeting_url"`
	RoomID       string    `json:"room_id"`
}

// UpdateMeetingRequest represents the request to update a meeting
type UpdateMeetingRequest struct {
	TitleAr     string    `json:"title_ar"`
	ScheduledAt time.Time `json:"scheduled_at"`
	Duration    int       `json:"duration"`
	MeetingURL  string    `json:"meeting_url"`
	RoomID      string    `json:"room_id"`
	Status      string    `json:"status"`
}

// MeetingResponse represents the meeting response
type MeetingResponse struct {
	ID           string     `json:"id"`
	InstructorID string     `json:"instructor_id"`
	StudentID    string     `json:"student_id"`
	SubjectID    string     `json:"subject_id"`
	TitleAr      string     `json:"title_ar"`
	ScheduledAt  time.Time  `json:"scheduled_at"`
	Duration     int        `json:"duration"`
	MeetingURL   string     `json:"meeting_url"`
	RoomID       string     `json:"room_id"`
	Status       string     `json:"status"`
	EndTime      *time.Time `json:"end_time"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CreateMeeting handles POST /api/v1/meetings
func (h *Handler) CreateMeeting(w http.ResponseWriter, r *http.Request) {
	var req CreateMeetingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	meeting := &models.Meeting{
		InstructorID: req.InstructorID,
		StudentID:    req.StudentID,
		SubjectID:    req.SubjectID,
		TitleAr:      req.TitleAr,
		ScheduledAt:  req.ScheduledAt,
		Duration:     req.Duration,
		MeetingURL:   req.MeetingURL,
		RoomID:       req.RoomID,
		Status:       "scheduled",
	}

	if err := h.service.CreateMeeting(r.Context(), meeting); err != nil {
		if err == ErrTimeConflict {
			respondError(w, http.StatusConflict, "Meeting time conflicts with existing meeting")
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, toMeetingResponse(meeting))
}

// GetMeeting handles GET /api/v1/meetings/{id}
func (h *Handler) GetMeeting(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	meeting, err := h.service.GetMeeting(r.Context(), id)
	if err != nil {
		if err == ErrMeetingNotFound {
			respondError(w, http.StatusNotFound, "Meeting not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, toMeetingResponse(meeting))
}

// ListMeetings handles GET /api/v1/meetings
func (h *Handler) ListMeetings(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	pageSize, _ := strconv.Atoi(query.Get("page_size"))
	instructorID := query.Get("instructor_id")
	studentID := query.Get("student_id")
	subjectID := query.Get("subject_id")
	status := query.Get("status")

	filters := map[string]interface{}{}
	if instructorID != "" {
		filters["instructor_id"] = instructorID
	}
	if studentID != "" {
		filters["student_id"] = studentID
	}
	if subjectID != "" {
		filters["subject_id"] = subjectID
	}
	if status != "" {
		filters["status"] = status
	}

	meetings, err := h.service.ListMeetings(r.Context(), filters, page, pageSize)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses := make([]MeetingResponse, len(meetings))
	for i, meeting := range meetings {
		responses[i] = toMeetingResponse(meeting)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"meetings":  responses,
		"page":      page,
		"page_size": pageSize,
		"total":     len(responses),
	})
}

// UpdateMeeting handles PUT /api/v1/meetings/{id}
func (h *Handler) UpdateMeeting(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req UpdateMeetingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	meeting := &models.Meeting{
		ID:          id,
		TitleAr:     req.TitleAr,
		ScheduledAt: req.ScheduledAt,
		Duration:    req.Duration,
		MeetingURL:  req.MeetingURL,
		RoomID:      req.RoomID,
		Status:      req.Status,
	}

	if err := h.service.UpdateMeeting(r.Context(), meeting); err != nil {
		if err == ErrMeetingNotFound {
			respondError(w, http.StatusNotFound, "Meeting not found")
			return
		}
		if err == ErrTimeConflict {
			respondError(w, http.StatusConflict, "Meeting time conflicts with existing meeting")
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, _ := h.service.GetMeeting(r.Context(), id)
	respondJSON(w, http.StatusOK, toMeetingResponse(updated))
}

// DeleteMeeting handles DELETE /api/v1/meetings/{id}
func (h *Handler) DeleteMeeting(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.DeleteMeeting(r.Context(), id); err != nil {
		if err == ErrMeetingNotFound {
			respondError(w, http.StatusNotFound, "Meeting not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Meeting deleted successfully"})
}

// CancelMeeting handles POST /api/v1/meetings/{id}/cancel
func (h *Handler) CancelMeeting(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.CancelMeeting(r.Context(), id); err != nil {
		if err == ErrMeetingNotFound {
			respondError(w, http.StatusNotFound, "Meeting not found")
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	meeting, _ := h.service.GetMeeting(r.Context(), id)
	respondJSON(w, http.StatusOK, toMeetingResponse(meeting))
}

// CompleteMeeting handles POST /api/v1/meetings/{id}/complete
func (h *Handler) CompleteMeeting(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.CompleteMeeting(r.Context(), id); err != nil {
		if err == ErrMeetingNotFound {
			respondError(w, http.StatusNotFound, "Meeting not found")
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	meeting, _ := h.service.GetMeeting(r.Context(), id)
	respondJSON(w, http.StatusOK, toMeetingResponse(meeting))
}

// Helper functions
func toMeetingResponse(meeting *models.Meeting) MeetingResponse {
	return MeetingResponse{
		ID:           meeting.ID,
		InstructorID: meeting.InstructorID,
		StudentID:    meeting.StudentID,
		SubjectID:    meeting.SubjectID,
		TitleAr:      meeting.TitleAr,
		ScheduledAt:  meeting.ScheduledAt,
		Duration:     meeting.Duration,
		MeetingURL:   meeting.MeetingURL,
		RoomID:       meeting.RoomID,
		Status:       meeting.Status,
		EndTime:      meeting.EndTime,
		CreatedAt:    meeting.CreatedAt,
		UpdatedAt:    meeting.UpdatedAt,
	}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
