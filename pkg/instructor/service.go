package instructor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Bashar444/VTP/pkg/models"
)

var (
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrInvalidAvailability = errors.New("invalid availability format")
)

// Service handles business logic for instructors
type Service struct {
	repo *Repository
}

// NewService creates a new instructor service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateInstructor creates a new instructor profile
func (s *Service) CreateInstructor(ctx context.Context, instructor *models.Instructor) error {
	// Validate required fields
	if instructor.NameAr == "" {
		return fmt.Errorf("name is required")
	}
	if instructor.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	// Validate hourly rate
	if instructor.HourlyRate < 0 {
		return fmt.Errorf("hourly rate cannot be negative")
	}

	// Validate rating
	if instructor.Rating < 0 || instructor.Rating > 5 {
		instructor.Rating = 0
	}

	// Initialize defaults
	if instructor.Specialization == "" {
		instructor.Specialization = "[]"
	}
	if instructor.CertificationsAr == "" {
		instructor.CertificationsAr = "[]"
	}
	if instructor.Availability == "" {
		instructor.Availability = "{}"
	}

	// Validate JSON fields
	if !isValidJSON(instructor.Specialization) {
		return fmt.Errorf("specialization must be valid JSON array")
	}
	if !isValidJSON(instructor.CertificationsAr) {
		return fmt.Errorf("certifications must be valid JSON array")
	}
	if !isValidJSON(instructor.Availability) {
		return ErrInvalidAvailability
	}

	return s.repo.Create(ctx, instructor)
}

// GetInstructor retrieves an instructor by ID
func (s *Service) GetInstructor(ctx context.Context, id string) (*models.Instructor, error) {
	return s.repo.GetByID(ctx, id)
}

// GetInstructorByUserID retrieves an instructor by user ID
func (s *Service) GetInstructorByUserID(ctx context.Context, userID string) (*models.Instructor, error) {
	return s.repo.GetByUserID(ctx, userID)
}

// ListInstructors retrieves instructors with filters
func (s *Service) ListInstructors(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*models.Instructor, error) {
	// Default pagination
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filters, pageSize, offset)
}

// UpdateInstructor updates an instructor profile
func (s *Service) UpdateInstructor(ctx context.Context, instructor *models.Instructor, userID string) error {
	// Authorization check: only the instructor owner can update
	existing, err := s.repo.GetByID(ctx, instructor.ID)
	if err != nil {
		return err
	}

	if existing.UserID != userID {
		return ErrUnauthorized
	}

	// Validate fields
	if instructor.NameAr == "" {
		return fmt.Errorf("name is required")
	}
	if instructor.HourlyRate < 0 {
		return fmt.Errorf("hourly rate cannot be negative")
	}
	if instructor.Rating < 0 || instructor.Rating > 5 {
		return fmt.Errorf("rating must be between 0 and 5")
	}

	// Validate JSON fields
	if instructor.Specialization != "" && !isValidJSON(instructor.Specialization) {
		return fmt.Errorf("specialization must be valid JSON array")
	}
	if instructor.CertificationsAr != "" && !isValidJSON(instructor.CertificationsAr) {
		return fmt.Errorf("certifications must be valid JSON array")
	}
	if instructor.Availability != "" && !isValidJSON(instructor.Availability) {
		return ErrInvalidAvailability
	}

	return s.repo.Update(ctx, instructor)
}

// DeleteInstructor soft deletes an instructor
func (s *Service) DeleteInstructor(ctx context.Context, id string, userID string) error {
	// Authorization check
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if existing.UserID != userID {
		return ErrUnauthorized
	}

	return s.repo.Delete(ctx, id)
}

// VerifyInstructor verifies an instructor (admin only)
func (s *Service) VerifyInstructor(ctx context.Context, id string) error {
	instructor, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	instructor.IsVerified = true
	return s.repo.Update(ctx, instructor)
}

// UpdateRating updates instructor rating
func (s *Service) UpdateRating(ctx context.Context, instructorID string, rating float32) error {
	if rating < 0 || rating > 5 {
		return fmt.Errorf("rating must be between 0 and 5")
	}

	return s.repo.UpdateRating(ctx, instructorID, rating)
}

// GetAvailableSlots returns available time slots for booking
func (s *Service) GetAvailableSlots(ctx context.Context, instructorID string, date time.Time) ([]string, error) {
	return s.repo.GetAvailableSlots(ctx, instructorID, date)
}

// SearchInstructorsBySubject searches instructors by subject
func (s *Service) SearchInstructorsBySubject(ctx context.Context, subjectID string, minRating float64, page, pageSize int) ([]*models.Instructor, error) {
	filters := map[string]interface{}{
		"subject_id":  subjectID,
		"is_active":   true,
		"is_verified": true,
	}

	if minRating > 0 {
		filters["min_rating"] = minRating
	}

	return s.ListInstructors(ctx, filters, page, pageSize)
}

// UpdateAvailability updates instructor availability schedule
func (s *Service) UpdateAvailability(ctx context.Context, instructorID string, userID string, availability map[string][]string) error {
	instructor, err := s.repo.GetByID(ctx, instructorID)
	if err != nil {
		return err
	}

	// Authorization check
	if instructor.UserID != userID {
		return ErrUnauthorized
	}

	// Convert to JSON
	availabilityJSON, err := json.Marshal(availability)
	if err != nil {
		return fmt.Errorf("failed to marshal availability: %w", err)
	}

	instructor.Availability = string(availabilityJSON)
	return s.repo.Update(ctx, instructor)
}

// AddSpecialization adds a subject to instructor's specialization
func (s *Service) AddSpecialization(ctx context.Context, instructorID string, userID string, subjectID string) error {
	instructor, err := s.repo.GetByID(ctx, instructorID)
	if err != nil {
		return err
	}

	// Authorization check
	if instructor.UserID != userID {
		return ErrUnauthorized
	}

	// Parse existing specializations
	var specializations []string
	if err := json.Unmarshal([]byte(instructor.Specialization), &specializations); err != nil {
		return fmt.Errorf("failed to parse specialization: %w", err)
	}

	// Check if already exists
	for _, s := range specializations {
		if s == subjectID {
			return nil // Already exists
		}
	}

	// Add new specialization
	specializations = append(specializations, subjectID)

	// Convert back to JSON
	specializationJSON, err := json.Marshal(specializations)
	if err != nil {
		return fmt.Errorf("failed to marshal specialization: %w", err)
	}

	instructor.Specialization = string(specializationJSON)
	return s.repo.Update(ctx, instructor)
}

// RemoveSpecialization removes a subject from instructor's specialization
func (s *Service) RemoveSpecialization(ctx context.Context, instructorID string, userID string, subjectID string) error {
	instructor, err := s.repo.GetByID(ctx, instructorID)
	if err != nil {
		return err
	}

	// Authorization check
	if instructor.UserID != userID {
		return ErrUnauthorized
	}

	// Parse existing specializations
	var specializations []string
	if err := json.Unmarshal([]byte(instructor.Specialization), &specializations); err != nil {
		return fmt.Errorf("failed to parse specialization: %w", err)
	}

	// Remove specialization
	newSpecializations := []string{}
	for _, s := range specializations {
		if s != subjectID {
			newSpecializations = append(newSpecializations, s)
		}
	}

	// Convert back to JSON
	specializationJSON, err := json.Marshal(newSpecializations)
	if err != nil {
		return fmt.Errorf("failed to marshal specialization: %w", err)
	}

	instructor.Specialization = string(specializationJSON)
	return s.repo.Update(ctx, instructor)
}

// Helper function to validate JSON
func isValidJSON(s string) bool {
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}
