package assignment

import (
	"context"
	"database/sql"
	"time"

	m "github.com/Bashar444/VTP/pkg/models"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Create(ctx context.Context, a *m.Assignment) (*m.Assignment, error) {
	q := `INSERT INTO assignments (course_id,instructor_id,title_ar,description_ar,subject_id,due_at,max_points)
          VALUES ($1,$2,$3,$4,$5,$6,$7)
          RETURNING id, created_at, updated_at`
	var id string
	var created, updated time.Time
	err := r.db.QueryRowContext(ctx, q, a.CourseID, a.InstructorID, a.TitleAR, a.DescriptionAR, a.SubjectID, a.DueAt, a.MaxPoints).Scan(&id, &created, &updated)
	if err != nil {
		return nil, err
	}
	a.ID = id
	a.CreatedAt = created
	a.UpdatedAt = updated
	return a, nil
}

func (r *Repository) List(ctx context.Context, instructorID *string, subjectID *string) ([]m.Assignment, error) {
	base := `SELECT id,course_id,instructor_id,title_ar,description_ar,subject_id,due_at,max_points,created_at,updated_at FROM assignments`
	var rows *sql.Rows
	var err error
	if instructorID != nil && subjectID != nil {
		rows, err = r.db.QueryContext(ctx, base+" WHERE instructor_id=$1 AND subject_id=$2 ORDER BY due_at ASC", *instructorID, *subjectID)
	} else if instructorID != nil {
		rows, err = r.db.QueryContext(ctx, base+" WHERE instructor_id=$1 ORDER BY due_at ASC", *instructorID)
	} else if subjectID != nil {
		rows, err = r.db.QueryContext(ctx, base+" WHERE subject_id=$1 ORDER BY due_at ASC", *subjectID)
	} else {
		rows, err = r.db.QueryContext(ctx, base+" ORDER BY due_at ASC")
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []m.Assignment
	for rows.Next() {
		var a m.Assignment
		var courseID, subjectIDNull sql.NullString
		err := rows.Scan(&a.ID, &courseID, &a.InstructorID, &a.TitleAR, &a.DescriptionAR, &subjectIDNull, &a.DueAt, &a.MaxPoints, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if courseID.Valid {
			cid := courseID.String
			a.CourseID = &cid
		}
		if subjectIDNull.Valid {
			sid := subjectIDNull.String
			a.SubjectID = &sid
		}
		res = append(res, a)
	}
	return res, nil
}

func (r *Repository) Get(ctx context.Context, id string) (*m.Assignment, error) {
	q := `SELECT id,course_id,instructor_id,title_ar,description_ar,subject_id,due_at,max_points,created_at,updated_at FROM assignments WHERE id=$1`
	var a m.Assignment
	var courseID, subjectID sql.NullString
	err := r.db.QueryRowContext(ctx, q, id).Scan(&a.ID, &courseID, &a.InstructorID, &a.TitleAR, &a.DescriptionAR, &subjectID, &a.DueAt, &a.MaxPoints, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if courseID.Valid {
		cid := courseID.String
		a.CourseID = &cid
	}
	if subjectID.Valid {
		sid := subjectID.String
		a.SubjectID = &sid
	}
	return &a, nil
}

func (r *Repository) CreateSubmission(ctx context.Context, s *m.AssignmentSubmission) (*m.AssignmentSubmission, error) {
	q := `INSERT INTO assignment_submissions (assignment_id,student_id,file_url,notes)
          VALUES ($1,$2,$3,$4)
          RETURNING id, submitted_at, created_at, updated_at`
	var id string
	var submitted, created, updated time.Time
	err := r.db.QueryRowContext(ctx, q, s.AssignmentID, s.StudentID, s.FileURL, s.Notes).
		Scan(&id, &submitted, &created, &updated)
	if err != nil {
		return nil, err
	}
	s.ID = id
	s.SubmittedAt = submitted
	s.CreatedAt = created
	s.UpdatedAt = updated
	return s, nil
}

func (r *Repository) GradeSubmission(ctx context.Context, submissionID string, grade int, feedback *string) (*m.AssignmentSubmission, error) {
	q := `UPDATE assignment_submissions SET grade=$1, graded_at=NOW(), feedback_ar=$2 WHERE id=$3 RETURNING assignment_id, student_id, submitted_at, file_url, notes, grade, graded_at, feedback_ar, created_at, updated_at`
	var s m.AssignmentSubmission
	s.ID = submissionID
	err := r.db.QueryRowContext(ctx, q, grade, feedback, submissionID).Scan(&s.AssignmentID, &s.StudentID, &s.SubmittedAt, &s.FileURL, &s.Notes, &s.Grade, &s.GradedAt, &s.FeedbackAR, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) ListSubmissions(ctx context.Context, assignmentID string) ([]m.AssignmentSubmission, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,assignment_id,student_id,submitted_at,file_url,notes,grade,graded_at,feedback_ar,created_at,updated_at FROM assignment_submissions WHERE assignment_id=$1 ORDER BY submitted_at DESC`, assignmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []m.AssignmentSubmission
	for rows.Next() {
		var s m.AssignmentSubmission
		err := rows.Scan(&s.ID, &s.AssignmentID, &s.StudentID, &s.SubmittedAt, &s.FileURL, &s.Notes, &s.Grade, &s.GradedAt, &s.FeedbackAR, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, err
		}
		res = append(res, s)
	}
	return res, nil
}
