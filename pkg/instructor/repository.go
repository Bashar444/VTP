package instructor

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Bashar444/VTP/pkg/models"
	"github.com/google/uuid"
)

var (
	ErrInstructorNotFound     = errors.New("instructor not found")
	ErrInstructorAlreadyExists = errors.New("instructor already exists for this user")
	ErrInvalidInstructorData  = errors.New("invalid instructor data")
)

// Repository handles instructor database operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new instructor repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new instructor
func (r *Repository) Create(ctx context.Context, instructor *models.Instructor) error {
	if instructor.UserID == "" || instructor.NameAr == "" {
		return ErrInvalidInstructorData
	}

	instructor.ID = uuid.New().String()
	instructor.CreatedAt = time.Now()
	instructor.UpdatedAt = time.Now()

	query := `
		INSERT INTO instructors (
			id, user_id, name_ar, bio_ar, specialization, hourly_rate,
			rating, total_reviews, years_experience, certifications_ar,
			availability, is_verified, is_active, profile_image_url,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err := r.db.ExecContext(ctx, query,
		instructor.ID,
		instructor.UserID,
		instructor.NameAr,
		instructor.BioAr,
		instructor.Specialization,
		instructor.HourlyRate,
		instructor.Rating,
		instructor.TotalReviews,
		instructor.YearsExperience,
		instructor.CertificationsAr,
		instructor.Availability,
		instructor.IsVerified,
		instructor.IsActive,
		instructor.ProfileImageURL,
		instructor.CreatedAt,
		instructor.UpdatedAt,
	)

	if err != nil {
		// Check for unique constraint violation
		if err.Error() == "pq: duplicate key value violates unique constraint \"instructors_user_id_key\"" {
			return ErrInstructorAlreadyExists
		}
		return fmt.Errorf("failed to create instructor: %w", err)
	}

	return nil
}

// GetByID retrieves an instructor by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*models.Instructor, error) {
	instructor := &models.Instructor{}

	query := `
		SELECT id, user_id, name_ar, bio_ar, specialization, hourly_rate,
			   rating, total_reviews, years_experience, certifications_ar,
			   availability, is_verified, is_active, profile_image_url,
			   created_at, updated_at
		FROM instructors
		WHERE id = $1
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&instructor.ID,
		&instructor.UserID,
		&instructor.NameAr,
		&instructor.BioAr,
		&instructor.Specialization,
		&instructor.HourlyRate,
		&instructor.Rating,
		&instructor.TotalReviews,
		&instructor.YearsExperience,
		&instructor.CertificationsAr,
		&instructor.Availability,
		&instructor.IsVerified,
		&instructor.IsActive,
		&instructor.ProfileImageURL,
		&instructor.CreatedAt,
		&instructor.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrInstructorNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get instructor: %w", err)
	}

	return instructor, nil
}

// GetByUserID retrieves an instructor by user ID
func (r *Repository) GetByUserID(ctx context.Context, userID string) (*models.Instructor, error) {
	instructor := &models.Instructor{}

	query := `
		SELECT id, user_id, name_ar, bio_ar, specialization, hourly_rate,
			   rating, total_reviews, years_experience, certifications_ar,
			   availability, is_verified, is_active, profile_image_url,
			   created_at, updated_at
		FROM instructors
		WHERE user_id = $1
	`

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&instructor.ID,
		&instructor.UserID,
		&instructor.NameAr,
		&instructor.BioAr,
		&instructor.Specialization,
		&instructor.HourlyRate,
		&instructor.Rating,
		&instructor.TotalReviews,
		&instructor.YearsExperience,
		&instructor.CertificationsAr,
		&instructor.Availability,
		&instructor.IsVerified,
		&instructor.IsActive,
		&instructor.ProfileImageURL,
		&instructor.CreatedAt,
		&instructor.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrInstructorNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get instructor by user ID: %w", err)
	}

	return instructor, nil
}

// List retrieves all instructors with optional filters
func (r *Repository) List(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.Instructor, error) {
	query := `
		SELECT id, user_id, name_ar, bio_ar, specialization, hourly_rate,
			   rating, total_reviews, years_experience, certifications_ar,
			   availability, is_verified, is_active, profile_image_url,
			   created_at, updated_at
		FROM instructors
		WHERE 1=1
	`
	args := []interface{}{}
	argPos := 1

	// Add filters
	if isActive, ok := filters["is_active"].(bool); ok {
		query += fmt.Sprintf(" AND is_active = $%d", argPos)
		args = append(args, isActive)
		argPos++
	}

	if isVerified, ok := filters["is_verified"].(bool); ok {
		query += fmt.Sprintf(" AND is_verified = $%d", argPos)
		args = append(args, isVerified)
		argPos++
	}

	if minRating, ok := filters["min_rating"].(float64); ok {
		query += fmt.Sprintf(" AND rating >= $%d", argPos)
		args = append(args, minRating)
		argPos++
	}

	if subjectID, ok := filters["subject_id"].(string); ok {
		query += fmt.Sprintf(" AND specialization @> $%d::jsonb", argPos)
		args = append(args, fmt.Sprintf(`["%s"]`, subjectID))
		argPos++
	}

	// Add ordering
	query += " ORDER BY rating DESC, total_reviews DESC"

	// Add pagination
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, limit)
		argPos++
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, offset)
		argPos++
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list instructors: %w", err)
	}
	defer rows.Close()

	instructors := []*models.Instructor{}
	for rows.Next() {
		instructor := &models.Instructor{}
		err := rows.Scan(
			&instructor.ID,
			&instructor.UserID,
			&instructor.NameAr,
			&instructor.BioAr,
			&instructor.Specialization,
			&instructor.HourlyRate,
			&instructor.Rating,
			&instructor.TotalReviews,
			&instructor.YearsExperience,
			&instructor.CertificationsAr,
			&instructor.Availability,
			&instructor.IsVerified,
			&instructor.IsActive,
			&instructor.ProfileImageURL,
			&instructor.CreatedAt,
			&instructor.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan instructor: %w", err)
		}
		instructors = append(instructors, instructor)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return instructors, nil
}

// Update updates an instructor
func (r *Repository) Update(ctx context.Context, instructor *models.Instructor) error {
	if instructor.ID == "" {
		return ErrInvalidInstructorData
	}

	instructor.UpdatedAt = time.Now()

	query := `
		UPDATE instructors
		SET name_ar = $1, bio_ar = $2, specialization = $3, hourly_rate = $4,
			rating = $5, total_reviews = $6, years_experience = $7,
			certifications_ar = $8, availability = $9, is_verified = $10,
			is_active = $11, profile_image_url = $12, updated_at = $13
		WHERE id = $14
	`

	result, err := r.db.ExecContext(ctx, query,
		instructor.NameAr,
		instructor.BioAr,
		instructor.Specialization,
		instructor.HourlyRate,
		instructor.Rating,
		instructor.TotalReviews,
		instructor.YearsExperience,
		instructor.CertificationsAr,
		instructor.Availability,
		instructor.IsVerified,
		instructor.IsActive,
		instructor.ProfileImageURL,
		instructor.UpdatedAt,
		instructor.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update instructor: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrInstructorNotFound
	}

	return nil
}

// Delete soft deletes an instructor by setting is_active to false
func (r *Repository) Delete(ctx context.Context, id string) error {
	query := `UPDATE instructors SET is_active = false, updated_at = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete instructor: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrInstructorNotFound
	}

	return nil
}

// UpdateRating updates instructor rating based on new review
func (r *Repository) UpdateRating(ctx context.Context, instructorID string, newRating float32) error {
	query := `
		UPDATE instructors
		SET rating = (rating * total_reviews + $1) / (total_reviews + 1),
			total_reviews = total_reviews + 1,
			updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, newRating, time.Now(), instructorID)
	if err != nil {
		return fmt.Errorf("failed to update rating: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrInstructorNotFound
	}

	return nil
}

// GetAvailableSlots returns available time slots for an instructor
func (r *Repository) GetAvailableSlots(ctx context.Context, instructorID string, date time.Time) ([]string, error) {
	instructor, err := r.GetByID(ctx, instructorID)
	if err != nil {
		return nil, err
	}

	// Parse availability JSON
	var availability map[string][]string
	if err := json.Unmarshal([]byte(instructor.Availability), &availability); err != nil {
		return nil, fmt.Errorf("failed to parse availability: %w", err)
	}

	// Get day of week
	dayOfWeek := date.Weekday().String()

	slots, ok := availability[dayOfWeek]
	if !ok {
		return []string{}, nil
	}

	// TODO: Filter out already booked slots by checking meetings table
	return slots, nil
}
