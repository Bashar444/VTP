package instructor

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Bashar444/VTP/pkg/models"
	"github.com/gorilla/mux"
)

// Handler handles HTTP requests for instructors
type Handler struct {
	service *Service
}

// NewHandler creates a new instructor handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateInstructorRequest represents the request to create an instructor
type CreateInstructorRequest struct {
	UserID           string              `json:"user_id"`
	NameAr           string              `json:"name_ar"`
	BioAr            string              `json:"bio_ar"`
	Specialization   []string            `json:"specialization"`
	HourlyRate       float64             `json:"hourly_rate"`
	YearsExperience  int                 `json:"years_experience"`
	CertificationsAr []string            `json:"certifications_ar"`
	Availability     map[string][]string `json:"availability"`
	ProfileImageURL  string              `json:"profile_image_url"`
}

// UpdateInstructorRequest represents the request to update an instructor
type UpdateInstructorRequest struct {
	NameAr           string              `json:"name_ar"`
	BioAr            string              `json:"bio_ar"`
	Specialization   []string            `json:"specialization"`
	HourlyRate       float64             `json:"hourly_rate"`
	YearsExperience  int                 `json:"years_experience"`
	CertificationsAr []string            `json:"certifications_ar"`
	Availability     map[string][]string `json:"availability"`
	ProfileImageURL  string              `json:"profile_image_url"`
}

// InstructorResponse represents the instructor response
type InstructorResponse struct {
	ID               string              `json:"id"`
	UserID           string              `json:"user_id"`
	NameAr           string              `json:"name_ar"`
	BioAr            string              `json:"bio_ar"`
	Specialization   []string            `json:"specialization"`
	HourlyRate       float64             `json:"hourly_rate"`
	Rating           float32             `json:"rating"`
	TotalReviews     int                 `json:"total_reviews"`
	YearsExperience  int                 `json:"years_experience"`
	CertificationsAr []string            `json:"certifications_ar"`
	Availability     map[string][]string `json:"availability"`
	IsVerified       bool                `json:"is_verified"`
	IsActive         bool                `json:"is_active"`
	ProfileImageURL  string              `json:"profile_image_url"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

// CreateInstructor handles POST /api/instructors
func (h *Handler) CreateInstructor(w http.ResponseWriter, r *http.Request) {
	var req CreateInstructorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Convert arrays to JSON strings
	specializationJSON, _ := json.Marshal(req.Specialization)
	certificationsJSON, _ := json.Marshal(req.CertificationsAr)
	availabilityJSON, _ := json.Marshal(req.Availability)

	instructor := &models.Instructor{
		UserID:           req.UserID,
		NameAr:           req.NameAr,
		BioAr:            req.BioAr,
		Specialization:   string(specializationJSON),
		HourlyRate:       req.HourlyRate,
		YearsExperience:  req.YearsExperience,
		CertificationsAr: string(certificationsJSON),
		Availability:     string(availabilityJSON),
		ProfileImageURL:  req.ProfileImageURL,
		IsActive:         true,
		IsVerified:       false,
		Rating:           0,
		TotalReviews:     0,
	}

	if err := h.service.CreateInstructor(r.Context(), instructor); err != nil {
		if err == ErrInstructorAlreadyExists {
			respondError(w, http.StatusConflict, "Instructor already exists for this user")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, toInstructorResponse(instructor))
}

// GetInstructor handles GET /api/instructors/{id}
func (h *Handler) GetInstructor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	instructor, err := h.service.GetInstructor(r.Context(), id)
	if err != nil {
		if err == ErrInstructorNotFound {
			respondError(w, http.StatusNotFound, "Instructor not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, toInstructorResponse(instructor))
}

// ListInstructors handles GET /api/instructors
func (h *Handler) ListInstructors(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	pageSize, _ := strconv.Atoi(query.Get("page_size"))
	subjectID := query.Get("subject_id")
	minRating, _ := strconv.ParseFloat(query.Get("min_rating"), 64)
	isVerified := query.Get("is_verified") == "true"

	// Build filters
	filters := map[string]interface{}{
		"is_active": true,
	}

	if isVerified {
		filters["is_verified"] = true
	}

	if minRating > 0 {
		filters["min_rating"] = minRating
	}

	if subjectID != "" {
		filters["subject_id"] = subjectID
	}

	instructors, err := h.service.ListInstructors(r.Context(), filters, page, pageSize)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses := make([]InstructorResponse, len(instructors))
	for i, instructor := range instructors {
		responses[i] = toInstructorResponse(instructor)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"instructors": responses,
		"page":        page,
		"page_size":   pageSize,
		"total":       len(responses),
	})
}

// UpdateInstructor handles PUT /api/instructors/{id}
func (h *Handler) UpdateInstructor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req UpdateInstructorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Convert arrays to JSON strings
	specializationJSON, _ := json.Marshal(req.Specialization)
	certificationsJSON, _ := json.Marshal(req.CertificationsAr)
	availabilityJSON, _ := json.Marshal(req.Availability)

	instructor := &models.Instructor{
		ID:               id,
		NameAr:           req.NameAr,
		BioAr:            req.BioAr,
		Specialization:   string(specializationJSON),
		HourlyRate:       req.HourlyRate,
		YearsExperience:  req.YearsExperience,
		CertificationsAr: string(certificationsJSON),
		Availability:     string(availabilityJSON),
		ProfileImageURL:  req.ProfileImageURL,
	}

	if err := h.service.UpdateInstructor(r.Context(), instructor, userID); err != nil {
		if err == ErrUnauthorized {
			respondError(w, http.StatusForbidden, "Forbidden")
			return
		}
		if err == ErrInstructorNotFound {
			respondError(w, http.StatusNotFound, "Instructor not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Fetch updated instructor
	updated, _ := h.service.GetInstructor(r.Context(), id)
	respondJSON(w, http.StatusOK, toInstructorResponse(updated))
}

// DeleteInstructor handles DELETE /api/instructors/{id}
func (h *Handler) DeleteInstructor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get user ID from context
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := h.service.DeleteInstructor(r.Context(), id, userID); err != nil {
		if err == ErrUnauthorized {
			respondError(w, http.StatusForbidden, "Forbidden")
			return
		}
		if err == ErrInstructorNotFound {
			respondError(w, http.StatusNotFound, "Instructor not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Instructor deleted successfully"})
}

// GetAvailableSlots handles GET /api/instructors/{id}/availability
func (h *Handler) GetAvailableSlots(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	dateStr := r.URL.Query().Get("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	slots, err := h.service.GetAvailableSlots(r.Context(), id, date)
	if err != nil {
		if err == ErrInstructorNotFound {
			respondError(w, http.StatusNotFound, "Instructor not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"date":  dateStr,
		"slots": slots,
	})
}

// Helper functions
func toInstructorResponse(instructor *models.Instructor) InstructorResponse {
	var specialization []string
	var certifications []string
	var availability map[string][]string

	json.Unmarshal([]byte(instructor.Specialization), &specialization)
	json.Unmarshal([]byte(instructor.CertificationsAr), &certifications)
	json.Unmarshal([]byte(instructor.Availability), &availability)

	return InstructorResponse{
		ID:               instructor.ID,
		UserID:           instructor.UserID,
		NameAr:           instructor.NameAr,
		BioAr:            instructor.BioAr,
		Specialization:   specialization,
		HourlyRate:       instructor.HourlyRate,
		Rating:           instructor.Rating,
		TotalReviews:     instructor.TotalReviews,
		YearsExperience:  instructor.YearsExperience,
		CertificationsAr: certifications,
		Availability:     availability,
		IsVerified:       instructor.IsVerified,
		IsActive:         instructor.IsActive,
		ProfileImageURL:  instructor.ProfileImageURL,
		CreatedAt:        instructor.CreatedAt,
		UpdatedAt:        instructor.UpdatedAt,
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
