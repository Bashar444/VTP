package recording

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type Handler struct {
	Service *Service
}

type listResponseItem struct {
	ID               string  `json:"id"`
	CourseID         *string `json:"course_id,omitempty"`
	InstructorID     string  `json:"instructor_id"`
	TitleAr          string  `json:"title_ar"`
	DescriptionAr    *string `json:"description_ar,omitempty"`
	SubjectID        *string `json:"subject_id,omitempty"`
	FileURL          string  `json:"file_url"`
	DurationSeconds  int     `json:"duration_seconds"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
	CourseTitleAr    *string `json:"course_title_ar,omitempty"`
	SubjectNameAr    *string `json:"subject_name_ar,omitempty"`
	InstructorNameAr *string `json:"instructor_name_ar,omitempty"`
}

type listResponse struct {
	Items []listResponseItem `json:"items"`
	Total int                `json:"total"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	p := ListParams{}
	if v := q.Get("course_id"); v != "" {
		p.CourseID = &v
	}
	if v := q.Get("instructor_id"); v != "" {
		p.InstructorID = &v
	}
	if v := q.Get("subject_id"); v != "" {
		p.SubjectID = &v
	}
	// optional pagination params
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			p.Limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			p.Offset = n
		}
	}
	items, err := h.Service.ListWithNames(r.Context(), p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	total, err := h.Service.Count(r.Context(), p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	out := make([]listResponseItem, 0, len(items))
	for _, it := range items {
		lr := listResponseItem{
			ID:               it.ID,
			CourseID:         it.CourseID,
			InstructorID:     it.InstructorID,
			TitleAr:          it.TitleAr,
			DescriptionAr:    it.DescriptionAr,
			SubjectID:        it.SubjectID,
			FileURL:          it.FileURL,
			DurationSeconds:  it.DurationSeconds,
			CreatedAt:        it.CreatedAt.Format(time.RFC3339),
			UpdatedAt:        it.UpdatedAt.Format(time.RFC3339),
			CourseTitleAr:    it.CourseTitleAr,
			SubjectNameAr:    it.SubjectNameAr,
			InstructorNameAr: it.InstructorNameAr,
		}
		out = append(out, lr)
	}
	json.NewEncoder(w).Encode(listResponse{Items: out, Total: total})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var in CreateInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	rec, err := h.Service.Create(r.Context(), in)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rec)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	rec, err := h.Service.Get(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(rec)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.Service.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
