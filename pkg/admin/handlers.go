package admin

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Bashar444/VTP/pkg/auth"
)

type Announcement struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	TitleAr   string    `json:"title_ar"`
	Message   string    `json:"message"`
	MessageAr string    `json:"message_ar"`
	Audience  string    `json:"audience"` // all | students | instructors | admins
	CreatedAt time.Time `json:"created_at"`
}

type Store struct {
	mu            sync.RWMutex
	announcements []Announcement
	nextID        int
	db            *sql.DB
}

func NewStore() *Store {
	return &Store{nextID: 1}
}

func NewStoreWithDB(db *sql.DB) *Store {
	return &Store{nextID: 1, db: db}
}

type Handler struct {
	store *Store
	am    *auth.AuthMiddleware
}

func NewHandler(store *Store, am *auth.AuthMiddleware) *Handler {
	return &Handler{store: store, am: am}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Admin-only endpoints
	mux.Handle("/api/v1/admin/announce", h.am.Middleware(h.am.RoleMiddleware("admin")(http.HandlerFunc(h.createAnnouncement))))
	mux.Handle("/api/v1/admin/announcements", h.am.Middleware(h.am.RoleMiddleware("admin")(http.HandlerFunc(h.listAnnouncements))))

	// User management
	mux.Handle("/api/v1/admin/users", h.am.Middleware(h.am.RoleMiddleware("admin")(http.HandlerFunc(h.listUsers))))
	mux.Handle("/api/v1/admin/users/create", h.am.Middleware(h.am.RoleMiddleware("admin")(http.HandlerFunc(h.createUser))))
	mux.Handle("/api/v1/admin/users/update", h.am.Middleware(h.am.RoleMiddleware("admin")(http.HandlerFunc(h.updateUser))))
	mux.Handle("/api/v1/admin/users/toggle-active", h.am.Middleware(h.am.RoleMiddleware("admin")(http.HandlerFunc(h.toggleUserActive))))

	// School management
	mux.Handle("/api/v1/admin/school-terms", h.am.Middleware(h.am.RoleMiddleware("admin")(http.HandlerFunc(h.handleSchoolTerms))))
	mux.Handle("/api/v1/admin/grade-levels", h.am.Middleware(h.am.RoleMiddleware("admin")(http.HandlerFunc(h.handleGradeLevels))))
	mux.Handle("/api/v1/admin/class-sections", h.am.Middleware(h.am.RoleMiddleware("admin")(http.HandlerFunc(h.handleClassSections))))

	// Reports
	mux.Handle("/api/v1/admin/reports/attendance", h.am.Middleware(h.am.RoleMiddleware("admin")(http.HandlerFunc(h.getAttendanceReport))))
	mux.Handle("/api/v1/admin/reports/overview", h.am.Middleware(h.am.RoleMiddleware("admin")(http.HandlerFunc(h.getOverviewReport))))

	// Dashboard stats
	mux.Handle("/api/v1/admin/dashboard", h.am.Middleware(h.am.RoleMiddleware("admin")(http.HandlerFunc(h.getDashboard))))
}

func (h *Handler) createAnnouncement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Title     string `json:"title"`
		TitleAr   string `json:"title_ar"`
		Message   string `json:"message"`
		MessageAr string `json:"message_ar"`
		Audience  string `json:"audience"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	if req.Title == "" && req.TitleAr == "" {
		http.Error(w, `{"error":"title required"}`, http.StatusBadRequest)
		return
	}
	if req.Audience == "" {
		req.Audience = "all"
	}

	h.store.mu.Lock()
	id := h.store.nextID
	h.store.nextID++
	ann := Announcement{
		ID:        fmt.Sprintf("a-%d", id),
		Title:     req.Title,
		TitleAr:   req.TitleAr,
		Message:   req.Message,
		MessageAr: req.MessageAr,
		Audience:  req.Audience,
		CreatedAt: time.Now(),
	}
	h.store.announcements = append(h.store.announcements, ann)
	h.store.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ann)
}

func (h *Handler) listAnnouncements(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	h.store.mu.RLock()
	list := make([]Announcement, len(h.store.announcements))
	copy(list, h.store.announcements)
	h.store.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"announcements": list})
}

// User Management Handlers
func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if h.store.db == nil {
		h.respondJSON(w, http.StatusOK, map[string]interface{}{
			"users": []interface{}{},
			"total": 0,
		})
		return
	}

	role := r.URL.Query().Get("role")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := `SELECT id, email, full_name, role, locale, is_active, created_at FROM users`
	args := []interface{}{}
	if role != "" {
		query += ` WHERE role = $1`
		args = append(args, role)
	}
	query += ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(len(args)+1) + ` OFFSET $` + strconv.Itoa(len(args)+2)
	args = append(args, pageSize, offset)

	rows, err := h.store.db.Query(query, args...)
	if err != nil {
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id, email, fullName, userRole, locale string
		var isActive sql.NullBool
		var createdAt time.Time
		if err := rows.Scan(&id, &email, &fullName, &userRole, &locale, &isActive, &createdAt); err != nil {
			continue
		}
		users = append(users, map[string]interface{}{
			"id":         id,
			"email":      email,
			"full_name":  fullName,
			"role":       userRole,
			"locale":     locale,
			"is_active":  isActive.Bool,
			"created_at": createdAt,
		})
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"users":     users,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email          string `json:"email"`
		FullName       string `json:"full_name"`
		Role           string `json:"role"`
		Password       string `json:"password"`
		Phone          string `json:"phone"`
		GradeLevelID   string `json:"grade_level_id"`
		ClassSectionID string `json:"class_section_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	// In production, this would create the user via auth service
	h.respondJSON(w, http.StatusCreated, map[string]string{
		"message": "User creation requires auth service integration",
		"email":   req.Email,
	})
}

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID             string `json:"id"`
		FullName       string `json:"full_name"`
		Phone          string `json:"phone"`
		Role           string `json:"role"`
		GradeLevelID   string `json:"grade_level_id"`
		ClassSectionID string `json:"class_section_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if h.store.db == nil {
		http.Error(w, `{"error":"database not configured"}`, http.StatusInternalServerError)
		return
	}

	_, err := h.store.db.Exec(`UPDATE users SET full_name = $2, phone = $3 WHERE id = $1`, req.ID, req.FullName, req.Phone)
	if err != nil {
		http.Error(w, `{"error":"update failed"}`, http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{"message": "User updated"})
}

func (h *Handler) toggleUserActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID       string `json:"id"`
		IsActive bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	if h.store.db == nil {
		http.Error(w, `{"error":"database not configured"}`, http.StatusInternalServerError)
		return
	}

	_, err := h.store.db.Exec(`UPDATE users SET is_active = $2 WHERE id = $1`, req.ID, req.IsActive)
	if err != nil {
		http.Error(w, `{"error":"update failed"}`, http.StatusInternalServerError)
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":   "User status updated",
		"is_active": req.IsActive,
	})
}

// School Management Handlers
func (h *Handler) handleSchoolTerms(w http.ResponseWriter, r *http.Request) {
	if h.store.db == nil {
		h.respondJSON(w, http.StatusOK, map[string]interface{}{"terms": []interface{}{}})
		return
	}

	switch r.Method {
	case http.MethodGet:
		rows, err := h.store.db.Query(`SELECT id, name_ar, name_en, start_date, end_date, is_active, academic_year FROM school_terms ORDER BY start_date DESC`)
		if err != nil {
			http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var terms []map[string]interface{}
		for rows.Next() {
			var id, nameAr, academicYear string
			var nameEn sql.NullString
			var startDate, endDate time.Time
			var isActive bool
			if err := rows.Scan(&id, &nameAr, &nameEn, &startDate, &endDate, &isActive, &academicYear); err != nil {
				continue
			}
			terms = append(terms, map[string]interface{}{
				"id":            id,
				"name_ar":       nameAr,
				"name_en":       nameEn.String,
				"start_date":    startDate.Format("2006-01-02"),
				"end_date":      endDate.Format("2006-01-02"),
				"is_active":     isActive,
				"academic_year": academicYear,
			})
		}
		h.respondJSON(w, http.StatusOK, map[string]interface{}{"terms": terms})

	case http.MethodPost:
		var req struct {
			NameAr       string `json:"name_ar"`
			NameEn       string `json:"name_en"`
			StartDate    string `json:"start_date"`
			EndDate      string `json:"end_date"`
			AcademicYear string `json:"academic_year"`
			IsActive     bool   `json:"is_active"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
			return
		}

		_, err := h.store.db.Exec(`
			INSERT INTO school_terms (name_ar, name_en, start_date, end_date, academic_year, is_active)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			req.NameAr, req.NameEn, req.StartDate, req.EndDate, req.AcademicYear, req.IsActive)
		if err != nil {
			http.Error(w, `{"error":"create failed"}`, http.StatusInternalServerError)
			return
		}
		h.respondJSON(w, http.StatusCreated, map[string]string{"message": "School term created"})

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleGradeLevels(w http.ResponseWriter, r *http.Request) {
	if h.store.db == nil {
		h.respondJSON(w, http.StatusOK, map[string]interface{}{"grade_levels": []interface{}{}})
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	rows, err := h.store.db.Query(`SELECT id, name_ar, name_en, level_number, education_stage FROM grade_levels ORDER BY level_number`)
	if err != nil {
		http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var levels []map[string]interface{}
	for rows.Next() {
		var id, nameAr, educationStage string
		var nameEn sql.NullString
		var levelNumber int
		if err := rows.Scan(&id, &nameAr, &nameEn, &levelNumber, &educationStage); err != nil {
			continue
		}
		levels = append(levels, map[string]interface{}{
			"id":              id,
			"name_ar":         nameAr,
			"name_en":         nameEn.String,
			"level_number":    levelNumber,
			"education_stage": educationStage,
		})
	}
	h.respondJSON(w, http.StatusOK, map[string]interface{}{"grade_levels": levels})
}

func (h *Handler) handleClassSections(w http.ResponseWriter, r *http.Request) {
	if h.store.db == nil {
		h.respondJSON(w, http.StatusOK, map[string]interface{}{"class_sections": []interface{}{}})
		return
	}

	switch r.Method {
	case http.MethodGet:
		gradeLevelID := r.URL.Query().Get("grade_level_id")
		termID := r.URL.Query().Get("term_id")

		query := `SELECT cs.id, cs.section_name, cs.max_students, gl.name_ar as grade_name, st.name_ar as term_name
			FROM class_sections cs
			JOIN grade_levels gl ON cs.grade_level_id = gl.id
			JOIN school_terms st ON cs.school_term_id = st.id
			WHERE 1=1`
		args := []interface{}{}

		if gradeLevelID != "" {
			query += ` AND cs.grade_level_id = $` + strconv.Itoa(len(args)+1)
			args = append(args, gradeLevelID)
		}
		if termID != "" {
			query += ` AND cs.school_term_id = $` + strconv.Itoa(len(args)+1)
			args = append(args, termID)
		}
		query += ` ORDER BY gl.level_number, cs.section_name`

		rows, err := h.store.db.Query(query, args...)
		if err != nil {
			http.Error(w, `{"error":"database error"}`, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var sections []map[string]interface{}
		for rows.Next() {
			var id, sectionName, gradeName, termName string
			var maxStudents int
			if err := rows.Scan(&id, &sectionName, &maxStudents, &gradeName, &termName); err != nil {
				continue
			}
			sections = append(sections, map[string]interface{}{
				"id":           id,
				"section_name": sectionName,
				"max_students": maxStudents,
				"grade_name":   gradeName,
				"term_name":    termName,
			})
		}
		h.respondJSON(w, http.StatusOK, map[string]interface{}{"class_sections": sections})

	case http.MethodPost:
		var req struct {
			GradeLevelID string `json:"grade_level_id"`
			SchoolTermID string `json:"school_term_id"`
			SectionName  string `json:"section_name"`
			MaxStudents  int    `json:"max_students"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
			return
		}
		if req.MaxStudents == 0 {
			req.MaxStudents = 40
		}

		_, err := h.store.db.Exec(`
			INSERT INTO class_sections (grade_level_id, school_term_id, section_name, max_students)
			VALUES ($1, $2, $3, $4)`,
			req.GradeLevelID, req.SchoolTermID, req.SectionName, req.MaxStudents)
		if err != nil {
			http.Error(w, `{"error":"create failed"}`, http.StatusInternalServerError)
			return
		}
		h.respondJSON(w, http.StatusCreated, map[string]string{"message": "Class section created"})

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

// Reports Handlers
func (h *Handler) getAttendanceReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if h.store.db == nil {
		h.respondJSON(w, http.StatusOK, map[string]interface{}{"report": map[string]int{}})
		return
	}

	// Date range from query params
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	if startDate == "" {
		startDate = time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	var total, present, absent, late, excused int
	row := h.store.db.QueryRow(`
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'present') as present,
			COUNT(*) FILTER (WHERE status = 'absent') as absent,
			COUNT(*) FILTER (WHERE status = 'late') as late,
			COUNT(*) FILTER (WHERE status = 'excused') as excused
		FROM attendance
		WHERE date >= $1 AND date <= $2`, startDate, endDate)
	row.Scan(&total, &present, &absent, &late, &excused)

	attendanceRate := 0.0
	if total > 0 {
		attendanceRate = float64(present+late) / float64(total) * 100
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"report": map[string]interface{}{
			"start_date":      startDate,
			"end_date":        endDate,
			"total_records":   total,
			"present":         present,
			"absent":          absent,
			"late":            late,
			"excused":         excused,
			"attendance_rate": attendanceRate,
		},
	})
}

func (h *Handler) getOverviewReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if h.store.db == nil {
		h.respondJSON(w, http.StatusOK, map[string]interface{}{"report": map[string]int{}})
		return
	}

	// Get counts for overview
	var totalStudents, totalTeachers, totalCourses, totalMeetings int
	h.store.db.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'student'`).Scan(&totalStudents)
	h.store.db.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'teacher'`).Scan(&totalTeachers)
	h.store.db.QueryRow(`SELECT COUNT(*) FROM courses`).Scan(&totalCourses)
	h.store.db.QueryRow(`SELECT COUNT(*) FROM meetings WHERE status = 'scheduled'`).Scan(&totalMeetings)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"report": map[string]int{
			"total_students":    totalStudents,
			"total_teachers":    totalTeachers,
			"total_courses":     totalCourses,
			"upcoming_meetings": totalMeetings,
		},
	})
}

func (h *Handler) getDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	dashboard := map[string]interface{}{
		"stats": map[string]interface{}{
			"total_students":      0,
			"total_teachers":      0,
			"total_courses":       0,
			"active_meetings":     0,
			"pending_assignments": 0,
		},
		"recent_activity":   []interface{}{},
		"upcoming_meetings": []interface{}{},
	}

	if h.store.db != nil {
		var students, teachers, courses, meetings, assignments int
		h.store.db.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'student'`).Scan(&students)
		h.store.db.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'teacher'`).Scan(&teachers)
		h.store.db.QueryRow(`SELECT COUNT(*) FROM courses`).Scan(&courses)
		h.store.db.QueryRow(`SELECT COUNT(*) FROM meetings WHERE status = 'scheduled' AND scheduled_at > NOW()`).Scan(&meetings)
		h.store.db.QueryRow(`SELECT COUNT(*) FROM assignments WHERE due_at > NOW()`).Scan(&assignments)

		dashboard["stats"] = map[string]interface{}{
			"total_students":      students,
			"total_teachers":      teachers,
			"total_courses":       courses,
			"active_meetings":     meetings,
			"pending_assignments": assignments,
		}
	}

	h.respondJSON(w, http.StatusOK, dashboard)
}

func (h *Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
