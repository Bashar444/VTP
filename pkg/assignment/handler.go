package assignment

import (
    "encoding/json"
    "net/http"
    "vtp/pkg/utils"
    m "vtp/pkg/models"
)

type Handler struct{ svc *Service }

func NewHandler(s *Service) *Handler { return &Handler{svc: s} }

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    var a m.Assignment
    if err := json.NewDecoder(r.Body).Decode(&a); err != nil { utils.WriteErr(w, http.StatusBadRequest, err); return }
    res, err := h.svc.Create(r.Context(), &a)
    if err != nil { utils.WriteErr(w, http.StatusBadRequest, err); return }
    utils.WriteJSON(w, http.StatusCreated, res)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query()
    var instructorID, subjectID *string
    if v := q.Get("instructor_id"); v != "" { instructorID = &v }
    if v := q.Get("subject_id"); v != "" { subjectID = &v }
    res, err := h.svc.List(r.Context(), instructorID, subjectID)
    if err != nil { utils.WriteErr(w, http.StatusInternalServerError, err); return }
    utils.WriteJSON(w, http.StatusOK, map[string]interface{}{ "assignments": res })
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
    id := utils.Param(r, "id")
    res, err := h.svc.Get(r.Context(), id)
    if err != nil { utils.WriteErr(w, http.StatusNotFound, err); return }
    utils.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
    var sub m.AssignmentSubmission
    if err := json.NewDecoder(r.Body).Decode(&sub); err != nil { utils.WriteErr(w, http.StatusBadRequest, err); return }
    res, err := h.svc.Submit(r.Context(), &sub)
    if err != nil { utils.WriteErr(w, http.StatusBadRequest, err); return }
    utils.WriteJSON(w, http.StatusCreated, res)
}

func (h *Handler) Grade(w http.ResponseWriter, r *http.Request) {
    id := utils.Param(r, "submissionId")
    var payload struct { Grade int `json:"grade"`; FeedbackAR *string `json:"feedback_ar"` }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil { utils.WriteErr(w, http.StatusBadRequest, err); return }
    res, err := h.svc.Grade(r.Context(), id, payload.Grade, payload.FeedbackAR)
    if err != nil { utils.WriteErr(w, http.StatusBadRequest, err); return }
    utils.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) ListSubmissions(w http.ResponseWriter, r *http.Request) {
    assignmentID := utils.Param(r, "id")
    res, err := h.svc.ListSubmissions(r.Context(), assignmentID)
    if err != nil { utils.WriteErr(w, http.StatusInternalServerError, err); return }
    utils.WriteJSON(w, http.StatusOK, map[string]interface{}{ "submissions": res })
}
