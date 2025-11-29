package subject

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
	ErrSubjectNotFound = errors.New("subject not found")
	ErrInvalidSubjectData = errors.New("invalid subject data")
)

// Repository handles subject database operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new subject repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new subject
func (r *Repository) Create(ctx context.Context, subject *models.Subject) error {
	if subject.NameAr == "" || subject.Level == "" || subject.Category == "" {
		return ErrInvalidSubjectData
	}

	subject.ID = uuid.New().String()
	subject.CreatedAt = time.Now()

	query := `
		INSERT INTO subjects (id, name_ar, level, category, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.ExecContext(ctx, query,
		subject.ID,
		subject.NameAr,
		subject.Level,
		subject.Category,
		subject.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create subject: %w", err)
	}

	return nil
}

// GetByID retrieves a subject by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*models.Subject, error) {
	subject := &models.Subject{}

	query := `
		SELECT id, name_ar, level, category, created_at
		FROM subjects
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&subject.ID,
		&subject.NameAr,
		&subject.Level,
		&subject.Category,
		&subject.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrSubjectNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subject: %w", err)
	}

	return subject, nil
}

// List retrieves subjects with optional filters
func (r *Repository) List(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.Subject, error) {
	query := `
		SELECT id, name_ar, level, category, created_at
		FROM subjects
		WHERE 1=1
	`
	args := []interface{}{}
	argPos := 1

	// Add filters
	if level, ok := filters["level"].(string); ok && level != "" {
		query += fmt.Sprintf(" AND level = $%d", argPos)
		args = append(args, level)
		argPos++
	}

	if category, ok := filters["category"].(string); ok && category != "" {
		query += fmt.Sprintf(" AND category = $%d", argPos)
		args = append(args, category)
		argPos++
	}

	// Add ordering
	query += " ORDER BY category, level, name_ar"

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
		return nil, fmt.Errorf("failed to list subjects: %w", err)
	}
	defer rows.Close()

	subjects := []*models.Subject{}
	for rows.Next() {
		subject := &models.Subject{}
		err := rows.Scan(
			&subject.ID,
			&subject.NameAr,
			&subject.Level,
			&subject.Category,
			&subject.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subject: %w", err)
		}
		subjects = append(subjects, subject)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return subjects, nil
}

// Update updates a subject
func (r *Repository) Update(ctx context.Context, subject *models.Subject) error {
	if subject.ID == "" {
		return ErrInvalidSubjectData
	}

	query := `
		UPDATE subjects
		SET name_ar = $1, level = $2, category = $3
		WHERE id = $4
	`

	result, err := r.db.ExecContext(ctx, query,
		subject.NameAr,
		subject.Level,
		subject.Category,
		subject.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update subject: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrSubjectNotFound
	}

	return nil
}

// Delete deletes a subject
func (r *Repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM subjects WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete subject: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrSubjectNotFound
	}

	return nil
}

// GetByCategory retrieves subjects by category
func (r *Repository) GetByCategory(ctx context.Context, category string) ([]*models.Subject, error) {
	return r.List(ctx, map[string]interface{}{"category": category}, 0, 0)
}

// GetByLevel retrieves subjects by level
func (r *Repository) GetByLevel(ctx context.Context, level string) ([]*models.Subject, error) {
	return r.List(ctx, map[string]interface{}{"level": level}, 0, 0)
}
