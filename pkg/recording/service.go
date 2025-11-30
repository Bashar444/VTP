package recording

import (
	"context"
	"errors"

	"github.com/Bashar444/VTP/pkg/models"
)

type Service struct {
	Repo *Repository
}

type CreateInput struct {
	CourseID        *string `json:"course_id"`
	InstructorID    string  `json:"instructor_id"`
	TitleAr         string  `json:"title_ar"`
	DescriptionAr   *string `json:"description_ar"`
	SubjectID       *string `json:"subject_id"`
	FileURL         string  `json:"file_url"`
	DurationSeconds int     `json:"duration_seconds"`
}

func (s *Service) List(ctx context.Context, p ListParams) ([]models.RecordingMeta, error) {
	return s.Repo.List(ctx, p)
}

func (s *Service) ListWithNames(ctx context.Context, p ListParams) ([]RecordWithNames, error) {
	return s.Repo.ListWithNames(ctx, p)
}

func (s *Service) Create(ctx context.Context, in CreateInput) (models.RecordingMeta, error) {
	if in.InstructorID == "" || in.TitleAr == "" || in.FileURL == "" {
		return models.RecordingMeta{}, errors.New("missing required fields")
	}
	return s.Repo.Create(ctx, RepositoryCreateToRepo(in))
}

func RepositoryCreateToRepo(in CreateInput) RepositoryCreateInput {
	return RepositoryCreateInput{
		CourseID:        in.CourseID,
		InstructorID:    in.InstructorID,
		TitleAr:         in.TitleAr,
		DescriptionAr:   in.DescriptionAr,
		SubjectID:       in.SubjectID,
		FileURL:         in.FileURL,
		DurationSeconds: in.DurationSeconds,
	}
}

type RepositoryCreateInput struct {
	CourseID        *string
	InstructorID    string
	TitleAr         string
	DescriptionAr   *string
	SubjectID       *string
	FileURL         string
	DurationSeconds int
}

func (s *Service) Get(ctx context.Context, id string) (models.RecordingMeta, error) {
	return s.Repo.Get(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.Repo.Delete(ctx, id)
}

func (s *Service) Count(ctx context.Context, p ListParams) (int, error) {
	// Count should ignore pagination and only use filters
	p.Limit = 0
	p.Offset = 0
	return s.Repo.Count(ctx, p)
}
