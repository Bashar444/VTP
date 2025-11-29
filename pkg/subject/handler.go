package subject

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Bashar444/VTP/pkg/models"
	"github.com/gorilla/mux"
)

// Handler handles HTTP requests for subjects
type Handler struct {
	service *Service
}

// NewHandler creates a new subject handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// SubjectRequest represents the request to create/update a subject
type SubjectRequest struct {
	NameAr   string `json:"name_ar"`
	Level    string `json:"level"`
	Category string `json:"category"`
}

// SubjectResponse represents the subject response
type SubjectResponse struct {
	ID        string    `json:"id"`
	NameAr    string    `json:"name_ar"`
	Level     string    `json:"level"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateSubject handles POST /api/v1/subjects
func (h *Handler) CreateSubject(w http.ResponseWriter, r *http.Request) {
	var req SubjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	subject := &models.Subject{
		NameAr:   req.NameAr,
		Level:    req.Level,
		Category: req.Category,
	}

	if err := h.service.CreateSubject(r.Context(), subject); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, toSubjectResponse(subject))
}

// GetSubject handles GET /api/v1/subjects/{id}
func (h *Handler) GetSubject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	subject, err := h.service.GetSubject(r.Context(), id)
	if err != nil {
		if err == ErrSubjectNotFound {
			respondError(w, http.StatusNotFound, "Subject not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, toSubjectResponse(subject))
}

// ListSubjects handles GET /api/v1/subjects
func (h *Handler) ListSubjects(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	pageSize, _ := strconv.Atoi(query.Get("page_size"))
	level := query.Get("level")
	category := query.Get("category")

	filters := map[string]interface{}{}
	if level != "" {
		filters["level"] = level
	}
	if category != "" {
		filters["category"] = category
	}

	subjects, err := h.service.ListSubjects(r.Context(), filters, page, pageSize)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses := make([]SubjectResponse, len(subjects))
	for i, subject := range subjects {
		responses[i] = toSubjectResponse(subject)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"subjects":  responses,
		"page":      page,
		"page_size": pageSize,
		"total":     len(responses),
	})
}

// UpdateSubject handles PUT /api/v1/subjects/{id}
func (h *Handler) UpdateSubject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req SubjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	subject := &models.Subject{
		ID:       id,
		NameAr:   req.NameAr,
		Level:    req.Level,
		Category: req.Category,
	}

	if err := h.service.UpdateSubject(r.Context(), subject); err != nil {
		if err == ErrSubjectNotFound {
			respondError(w, http.StatusNotFound, "Subject not found")
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, _ := h.service.GetSubject(r.Context(), id)
	respondJSON(w, http.StatusOK, toSubjectResponse(updated))
}

// DeleteSubject handles DELETE /api/v1/subjects/{id}
func (h *Handler) DeleteSubject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.DeleteSubject(r.Context(), id); err != nil {
		if err == ErrSubjectNotFound {
			respondError(w, http.StatusNotFound, "Subject not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Subject deleted successfully"})
}

// Helper functions
func toSubjectResponse(subject *models.Subject) SubjectResponse {
	return SubjectResponse{
		ID:        subject.ID,
		NameAr:    subject.NameAr,
		Level:     subject.Level,
		Category:  subject.Category,
		CreatedAt: subject.CreatedAt,
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
