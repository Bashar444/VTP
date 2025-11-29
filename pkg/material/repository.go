package material

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Bashar444/VTP/pkg/models"
	"github.com/google/uuid"
)

var (
	ErrMaterialNotFound    = errors.New("study material not found")
	ErrInvalidMaterialData = errors.New("invalid study material data")
)

// Repository handles study material database operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new material repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new study material
func (r *Repository) Create(ctx context.Context, material *models.StudyMaterial) error {
	if material.InstructorID == "" || material.TitleAr == "" || material.FileURL == "" {
		return ErrInvalidMaterialData
	}

	material.ID = uuid.New().String()
	material.CreatedAt = time.Now()
	material.UpdatedAt = time.Now()

	query := `
		INSERT INTO study_materials (
			id, course_id, instructor_id, title_ar, type,
			file_url, file_size, downloads, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		material.ID,
		nullString(material.CourseID),
		material.InstructorID,
		material.TitleAr,
		material.Type,
		material.FileURL,
		material.FileSize,
		material.Downloads,
		material.CreatedAt,
		material.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create study material: %w", err)
	}

	return nil
}

// GetByID retrieves a study material by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*models.StudyMaterial, error) {
	material := &models.StudyMaterial{}
	var courseID sql.NullString

	query := `
		SELECT id, course_id, instructor_id, title_ar, type,
			   file_url, file_size, downloads, created_at, updated_at
		FROM study_materials
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&material.ID,
		&courseID,
		&material.InstructorID,
		&material.TitleAr,
		&material.Type,
		&material.FileURL,
		&material.FileSize,
		&material.Downloads,
		&material.CreatedAt,
		&material.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrMaterialNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get study material: %w", err)
	}

	material.CourseID = courseID.String

	return material, nil
}

// List retrieves study materials with filters
func (r *Repository) List(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.StudyMaterial, error) {
	query := `
		SELECT id, course_id, instructor_id, title_ar, type,
			   file_url, file_size, downloads, created_at, updated_at
		FROM study_materials
		WHERE 1=1
	`
	args := []interface{}{}
	argPos := 1

	// Add filters
	if courseID, ok := filters["course_id"].(string); ok && courseID != "" {
		query += fmt.Sprintf(" AND course_id = $%d", argPos)
		args = append(args, courseID)
		argPos++
	}

	if instructorID, ok := filters["instructor_id"].(string); ok && instructorID != "" {
		query += fmt.Sprintf(" AND instructor_id = $%d", argPos)
		args = append(args, instructorID)
		argPos++
	}

	if materialType, ok := filters["type"].(string); ok && materialType != "" {
		query += fmt.Sprintf(" AND type = $%d", argPos)
		args = append(args, materialType)
		argPos++
	}

	// Add ordering
	query += " ORDER BY created_at DESC"

	// Add pagination
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, limit)
		argPos++
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list study materials: %w", err)
	}
	defer rows.Close()

	materials := []*models.StudyMaterial{}
	for rows.Next() {
		material := &models.StudyMaterial{}
		var courseID sql.NullString

		err := rows.Scan(
			&material.ID,
			&courseID,
			&material.InstructorID,
			&material.TitleAr,
			&material.Type,
			&material.FileURL,
			&material.FileSize,
			&material.Downloads,
			&material.CreatedAt,
			&material.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan study material: %w", err)
		}

		material.CourseID = courseID.String
		materials = append(materials, material)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return materials, nil
}

// Update updates a study material
func (r *Repository) Update(ctx context.Context, material *models.StudyMaterial) error {
	if material.ID == "" {
		return ErrInvalidMaterialData
	}

	material.UpdatedAt = time.Now()

	query := `
		UPDATE study_materials
		SET title_ar = $1, type = $2, file_url = $3,
			file_size = $4, updated_at = $5
		WHERE id = $6
	`

	result, err := r.db.ExecContext(ctx, query,
		material.TitleAr,
		material.Type,
		material.FileURL,
		material.FileSize,
		material.UpdatedAt,
		material.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update study material: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrMaterialNotFound
	}

	return nil
}

// Delete deletes a study material
func (r *Repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM study_materials WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete study material: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrMaterialNotFound
	}

	return nil
}

// IncrementDownloads increments the download counter
func (r *Repository) IncrementDownloads(ctx context.Context, id string) error {
	query := `
		UPDATE study_materials
		SET downloads = downloads + 1
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment downloads: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrMaterialNotFound
	}

	return nil
}

// Helper functions
func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
