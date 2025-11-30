package recording

import (
	"context"
	"database/sql"
	"time"

	"github.com/Bashar444/VTP/pkg/models"
)

type Repository struct {
	DB *sql.DB
}

type ListParams struct {
	CourseID     *string
	InstructorID *string
	SubjectID    *string
	Limit        int
	Offset       int
}

// Count returns the total number of recordings matching the provided filters (ignores limit/offset)
func (r *Repository) Count(ctx context.Context, p ListParams) (int, error) {
	q := `SELECT COUNT(*) FROM recordings`
	where := ""
	args := []any{}
	if p.CourseID != nil {
		where += " course_id = ?"
		args = append(args, *p.CourseID)
	}
	if p.InstructorID != nil {
		if where != "" {
			where += " AND"
		}
		where += " instructor_id = ?"
		args = append(args, *p.InstructorID)
	}
	if p.SubjectID != nil {
		if where != "" {
			where += " AND"
		}
		where += " subject_id = ?"
		args = append(args, *p.SubjectID)
	}
	if where != "" {
		q += " WHERE" + where
	}
	var total int
	if err := r.DB.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) List(ctx context.Context, p ListParams) ([]models.RecordingMeta, error) {
	q := `SELECT id, course_id, instructor_id, title_ar, description_ar, subject_id, file_url, duration_seconds, created_at, updated_at FROM recordings`
	where := ""
	args := []any{}
	if p.CourseID != nil {
		where += " course_id = ?"
		args = append(args, *p.CourseID)
	}
	if p.InstructorID != nil {
		if where != "" {
			where += " AND"
		}
		where += " instructor_id = ?"
		args = append(args, *p.InstructorID)
	}
	if p.SubjectID != nil {
		if where != "" {
			where += " AND"
		}
		where += " subject_id = ?"
		args = append(args, *p.SubjectID)
	}
	if where != "" {
		q += " WHERE" + where
	}
	q += " ORDER BY created_at DESC"
	if p.Limit > 0 {
		q += " LIMIT ?"
		args = append(args, p.Limit)
	}
	if p.Offset > 0 {
		q += " OFFSET ?"
		args = append(args, p.Offset)
	}
	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := []models.RecordingMeta{}
	for rows.Next() {
		var rec models.RecordingMeta
		var courseID, subjectID sql.NullString
		var desc sql.NullString
		if err := rows.Scan(&rec.ID, &courseID, &rec.InstructorID, &rec.TitleAr, &desc, &subjectID, &rec.FileURL, &rec.DurationSeconds, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
			return nil, err
		}
		if courseID.Valid {
			v := courseID.String
			rec.CourseID = &v
		}
		if subjectID.Valid {
			v := subjectID.String
			rec.SubjectID = &v
		}
		if desc.Valid {
			v := desc.String
			rec.DescriptionAr = &v
		}
		res = append(res, rec)
	}
	return res, nil
}

// RecordWithNames is a denormalized view for UI convenience
type RecordWithNames struct {
	ID               string
	CourseID         *string
	InstructorID     string
	TitleAr          string
	DescriptionAr    *string
	SubjectID        *string
	FileURL          string
	DurationSeconds  int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	CourseTitleAr    *string
	SubjectNameAr    *string
	InstructorNameAr *string
}

// ListWithNames returns recordings with denormalized names via LEFT JOINs
func (r *Repository) ListWithNames(ctx context.Context, p ListParams) ([]RecordWithNames, error) {
	q := `SELECT r.id, r.course_id, r.instructor_id, r.title_ar, r.description_ar, r.subject_id, r.file_url, r.duration_seconds, r.created_at, r.updated_at,
				 c.title_ar AS course_title_ar,
				 s.name_ar AS subject_name_ar,
				 i.name_ar AS instructor_name_ar
		  FROM recordings r
		  LEFT JOIN courses c ON c.id = r.course_id
		  LEFT JOIN subjects s ON s.id = r.subject_id
		  LEFT JOIN instructors i ON i.id = r.instructor_id`
	where := ""
	args := []any{}
	if p.CourseID != nil {
		where += " r.course_id = ?"
		args = append(args, *p.CourseID)
	}
	if p.InstructorID != nil {
		if where != "" {
			where += " AND"
		}
		where += " r.instructor_id = ?"
		args = append(args, *p.InstructorID)
	}
	if p.SubjectID != nil {
		if where != "" {
			where += " AND"
		}
		where += " r.subject_id = ?"
		args = append(args, *p.SubjectID)
	}
	if where != "" {
		q += " WHERE" + where
	}
	q += " ORDER BY r.created_at DESC"
	if p.Limit > 0 {
		q += " LIMIT ?"
		args = append(args, p.Limit)
	}
	if p.Offset > 0 {
		q += " OFFSET ?"
		args = append(args, p.Offset)
	}

	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := []RecordWithNames{}
	for rows.Next() {
		var rec RecordWithNames
		var courseID, subjectID sql.NullString
		var desc, courseTitle, subjectName, instructorName sql.NullString
		if err := rows.Scan(&rec.ID, &courseID, &rec.InstructorID, &rec.TitleAr, &desc, &subjectID, &rec.FileURL, &rec.DurationSeconds, &rec.CreatedAt, &rec.UpdatedAt,
			&courseTitle, &subjectName, &instructorName); err != nil {
			return nil, err
		}
		if courseID.Valid {
			v := courseID.String
			rec.CourseID = &v
		}
		if subjectID.Valid {
			v := subjectID.String
			rec.SubjectID = &v
		}
		if desc.Valid {
			v := desc.String
			rec.DescriptionAr = &v
		}
		if courseTitle.Valid {
			v := courseTitle.String
			rec.CourseTitleAr = &v
		}
		if subjectName.Valid {
			v := subjectName.String
			rec.SubjectNameAr = &v
		}
		if instructorName.Valid {
			v := instructorName.String
			rec.InstructorNameAr = &v
		}
		res = append(res, rec)
	}
	return res, nil
}

func (r *Repository) Create(ctx context.Context, in RepositoryCreateInput) (models.RecordingMeta, error) {
	rec := models.RecordingMeta{
		ID:              "",
		CourseID:        in.CourseID,
		InstructorID:    in.InstructorID,
		TitleAr:         in.TitleAr,
		DescriptionAr:   in.DescriptionAr,
		SubjectID:       in.SubjectID,
		FileURL:         in.FileURL,
		DurationSeconds: in.DurationSeconds,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	q := `INSERT INTO recordings(course_id, instructor_id, title_ar, description_ar, subject_id, file_url, duration_seconds) VALUES(?,?,?,?,?,?,?) RETURNING id, created_at, updated_at`
	var createdAt, updatedAt time.Time
	var id string
	var desc any
	if in.DescriptionAr == nil {
		desc = nil
	} else {
		desc = *in.DescriptionAr
	}
	var subj any
	if in.SubjectID == nil {
		subj = nil
	} else {
		subj = *in.SubjectID
	}
	var course any
	if in.CourseID == nil {
		course = nil
	} else {
		course = *in.CourseID
	}
	err := r.DB.QueryRowContext(ctx, q, course, in.InstructorID, in.TitleAr, desc, subj, in.FileURL, in.DurationSeconds).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return models.RecordingMeta{}, err
	}
	rec.ID = id
	rec.CreatedAt = createdAt
	rec.UpdatedAt = updatedAt
	return rec, nil
}

func (r *Repository) Get(ctx context.Context, id string) (models.RecordingMeta, error) {
	q := `SELECT id, course_id, instructor_id, title_ar, description_ar, subject_id, file_url, duration_seconds, created_at, updated_at FROM recordings WHERE id = ?`
	var rec models.RecordingMeta
	var courseID, subjectID sql.NullString
	var desc sql.NullString
	err := r.DB.QueryRowContext(ctx, q, id).Scan(&rec.ID, &courseID, &rec.InstructorID, &rec.TitleAr, &desc, &subjectID, &rec.FileURL, &rec.DurationSeconds, &rec.CreatedAt, &rec.UpdatedAt)
	if err != nil {
		return models.RecordingMeta{}, err
	}
	if courseID.Valid {
		v := courseID.String
		rec.CourseID = &v
	}
	if subjectID.Valid {
		v := subjectID.String
		rec.SubjectID = &v
	}
	if desc.Valid {
		v := desc.String
		rec.DescriptionAr = &v
	}
	return rec, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM recordings WHERE id = ?`, id)
	return err
}
