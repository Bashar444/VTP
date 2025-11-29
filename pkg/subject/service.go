package subject

import (
	"context"
	"fmt"

	"github.com/Bashar444/VTP/pkg/models"
)

// Service handles business logic for subjects
type Service struct {
	repo *Repository
}

// NewService creates a new subject service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateSubject creates a new subject
func (s *Service) CreateSubject(ctx context.Context, subject *models.Subject) error {
	// Validate required fields
	if subject.NameAr == "" {
		return fmt.Errorf("name is required")
	}
	if subject.Level == "" {
		return fmt.Errorf("level is required")
	}
	if subject.Category == "" {
		return fmt.Errorf("category is required")
	}

	// Validate level
	validLevels := map[string]bool{
		"elementary":   true,
		"intermediate": true,
		"advanced":     true,
	}
	if !validLevels[subject.Level] {
		return fmt.Errorf("invalid level: must be elementary, intermediate, or advanced")
	}

	return s.repo.Create(ctx, subject)
}

// GetSubject retrieves a subject by ID
func (s *Service) GetSubject(ctx context.Context, id string) (*models.Subject, error) {
	return s.repo.GetByID(ctx, id)
}

// ListSubjects retrieves subjects with filters
func (s *Service) ListSubjects(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*models.Subject, error) {
	// Default pagination
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 50
	}
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * pageSize
	return s.repo.List(ctx, filters, pageSize, offset)
}

// UpdateSubject updates a subject
func (s *Service) UpdateSubject(ctx context.Context, subject *models.Subject) error {
	if subject.NameAr == "" {
		return fmt.Errorf("name is required")
	}
	if subject.Level == "" {
		return fmt.Errorf("level is required")
	}
	if subject.Category == "" {
		return fmt.Errorf("category is required")
	}

	// Validate level
	validLevels := map[string]bool{
		"elementary":   true,
		"intermediate": true,
		"advanced":     true,
	}
	if !validLevels[subject.Level] {
		return fmt.Errorf("invalid level: must be elementary, intermediate, or advanced")
	}

	return s.repo.Update(ctx, subject)
}

// DeleteSubject deletes a subject
func (s *Service) DeleteSubject(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// GetSubjectsByCategory retrieves subjects by category
func (s *Service) GetSubjectsByCategory(ctx context.Context, category string) ([]*models.Subject, error) {
	return s.repo.GetByCategory(ctx, category)
}

// GetSubjectsByLevel retrieves subjects by level
func (s *Service) GetSubjectsByLevel(ctx context.Context, level string) ([]*models.Subject, error) {
	return s.repo.GetByLevel(ctx, level)
}
