package material

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Bashar444/VTP/pkg/models"
	"github.com/gorilla/mux"
)

// Handler handles HTTP requests for study materials
type Handler struct {
	service *Service
}

// NewHandler creates a new material handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateMaterialRequest represents the request to create a study material
type CreateMaterialRequest struct {
	CourseID     string `json:"course_id"`
	InstructorID string `json:"instructor_id"`
	TitleAr      string `json:"title_ar"`
	Type         string `json:"type"`
	FileURL      string `json:"file_url"`
	FileSize     int64  `json:"file_size"`
}

// UpdateMaterialRequest represents the request to update a study material
type UpdateMaterialRequest struct {
	TitleAr  string `json:"title_ar"`
	Type     string `json:"type"`
	FileURL  string `json:"file_url"`
	FileSize int64  `json:"file_size"`
}

// MaterialResponse represents the material response
type MaterialResponse struct {
	ID           string    `json:"id"`
	CourseID     string    `json:"course_id"`
	InstructorID string    `json:"instructor_id"`
	TitleAr      string    `json:"title_ar"`
	Type         string    `json:"type"`
	FileURL      string    `json:"file_url"`
	FileSize     int64     `json:"file_size"`
	Downloads    int       `json:"downloads"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateMaterial handles POST /api/v1/materials
func (h *Handler) CreateMaterial(w http.ResponseWriter, r *http.Request) {
	var req CreateMaterialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	material := &models.StudyMaterial{
		CourseID:     req.CourseID,
		InstructorID: req.InstructorID,
		TitleAr:      req.TitleAr,
		Type:         req.Type,
		FileURL:      req.FileURL,
		FileSize:     req.FileSize,
		Downloads:    0,
	}

	if err := h.service.CreateMaterial(r.Context(), material); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, toMaterialResponse(material))
}

// GetMaterial handles GET /api/v1/materials/{id}
func (h *Handler) GetMaterial(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	material, err := h.service.GetMaterial(r.Context(), id)
	if err != nil {
		if err == ErrMaterialNotFound {
			respondError(w, http.StatusNotFound, "Study material not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, toMaterialResponse(material))
}

// ListMaterials handles GET /api/v1/materials
func (h *Handler) ListMaterials(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	pageSize, _ := strconv.Atoi(query.Get("page_size"))
	courseID := query.Get("course_id")
	instructorID := query.Get("instructor_id")
	materialType := query.Get("type")

	filters := map[string]interface{}{}
	if courseID != "" {
		filters["course_id"] = courseID
	}
	if instructorID != "" {
		filters["instructor_id"] = instructorID
	}
	if materialType != "" {
		filters["type"] = materialType
	}

	materials, err := h.service.ListMaterials(r.Context(), filters, page, pageSize)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responses := make([]MaterialResponse, len(materials))
	for i, material := range materials {
		responses[i] = toMaterialResponse(material)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"materials": responses,
		"page":      page,
		"page_size": pageSize,
		"total":     len(responses),
	})
}

// UpdateMaterial handles PUT /api/v1/materials/{id}
func (h *Handler) UpdateMaterial(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req UpdateMaterialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	material := &models.StudyMaterial{
		ID:       id,
		TitleAr:  req.TitleAr,
		Type:     req.Type,
		FileURL:  req.FileURL,
		FileSize: req.FileSize,
	}

	if err := h.service.UpdateMaterial(r.Context(), material); err != nil {
		if err == ErrMaterialNotFound {
			respondError(w, http.StatusNotFound, "Study material not found")
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, _ := h.service.GetMaterial(r.Context(), id)
	respondJSON(w, http.StatusOK, toMaterialResponse(updated))
}

// DeleteMaterial handles DELETE /api/v1/materials/{id}
func (h *Handler) DeleteMaterial(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.DeleteMaterial(r.Context(), id); err != nil {
		if err == ErrMaterialNotFound {
			respondError(w, http.StatusNotFound, "Study material not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Study material deleted successfully"})
}

// DownloadMaterial handles GET /api/v1/materials/{id}/download
func (h *Handler) DownloadMaterial(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	material, err := h.service.GetMaterial(r.Context(), id)
	if err != nil {
		if err == ErrMaterialNotFound {
			respondError(w, http.StatusNotFound, "Study material not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Track download
	_ = h.service.TrackDownload(r.Context(), id)

	// Redirect to file URL or serve file
	http.Redirect(w, r, material.FileURL, http.StatusFound)
}

// Helper functions
func toMaterialResponse(material *models.StudyMaterial) MaterialResponse {
	return MaterialResponse{
		ID:           material.ID,
		CourseID:     material.CourseID,
		InstructorID: material.InstructorID,
		TitleAr:      material.TitleAr,
		Type:         material.Type,
		FileURL:      material.FileURL,
		FileSize:     material.FileSize,
		Downloads:    material.Downloads,
		CreatedAt:    material.CreatedAt,
		UpdatedAt:    material.UpdatedAt,
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
