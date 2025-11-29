package material

import (
	"context"
	"fmt"

	"github.com/Bashar444/VTP/pkg/models"
)

// Service handles business logic for study materials
type Service struct {
	repo *Repository
}

// NewService creates a new material service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateMaterial creates a new study material
func (s *Service) CreateMaterial(ctx context.Context, material *models.StudyMaterial) error {
	// Validate required fields
	if material.InstructorID == "" {
		return fmt.Errorf("instructor ID is required")
	}
	if material.TitleAr == "" {
		return fmt.Errorf("title is required")
	}
	if material.FileURL == "" {
		return fmt.Errorf("file URL is required")
	}
	if material.Type == "" {
		return fmt.Errorf("type is required")
	}

	// Validate type
	validTypes := map[string]bool{
		"pdf":       true,
		"slides":    true,
		"notes":     true,
		"worksheet": true,
		"video":     true,
		"audio":     true,
	}
	if !validTypes[material.Type] {
		return fmt.Errorf("invalid type: must be pdf, slides, notes, worksheet, video, or audio")
	}

	return s.repo.Create(ctx, material)
}

// GetMaterial retrieves a study material by ID
func (s *Service) GetMaterial(ctx context.Context, id string) (*models.StudyMaterial, error) {
	return s.repo.GetByID(ctx, id)
}

// ListMaterials retrieves study materials with filters
func (s *Service) ListMaterials(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*models.StudyMaterial, error) {
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

// UpdateMaterial updates a study material
func (s *Service) UpdateMaterial(ctx context.Context, material *models.StudyMaterial) error {
	if material.ID == "" {
		return fmt.Errorf("material ID is required")
	}
	if material.TitleAr == "" {
		return fmt.Errorf("title is required")
	}
	if material.FileURL == "" {
		return fmt.Errorf("file URL is required")
	}
	if material.Type == "" {
		return fmt.Errorf("type is required")
	}

	// Validate type
	validTypes := map[string]bool{
		"pdf":       true,
		"slides":    true,
		"notes":     true,
		"worksheet": true,
		"video":     true,
		"audio":     true,
	}
	if !validTypes[material.Type] {
		return fmt.Errorf("invalid type: must be pdf, slides, notes, worksheet, video, or audio")
	}

	return s.repo.Update(ctx, material)
}

// DeleteMaterial deletes a study material
func (s *Service) DeleteMaterial(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// GetCourseMaterials retrieves materials for a course
func (s *Service) GetCourseMaterials(ctx context.Context, courseID string, materialType string, page, pageSize int) ([]*models.StudyMaterial, error) {
	filters := map[string]interface{}{
		"course_id": courseID,
	}
	if materialType != "" {
		filters["type"] = materialType
	}
	return s.ListMaterials(ctx, filters, page, pageSize)
}

// GetInstructorMaterials retrieves materials by an instructor
func (s *Service) GetInstructorMaterials(ctx context.Context, instructorID string, materialType string, page, pageSize int) ([]*models.StudyMaterial, error) {
	filters := map[string]interface{}{
		"instructor_id": instructorID,
	}
	if materialType != "" {
		filters["type"] = materialType
	}
	return s.ListMaterials(ctx, filters, page, pageSize)
}

// TrackDownload increments download counter
func (s *Service) TrackDownload(ctx context.Context, id string) error {
	return s.repo.IncrementDownloads(ctx, id)
}
