package assignment

import (
	"context"

	m "github.com/Bashar444/VTP/pkg/models"
)

type Service struct{ repo *Repository }

func NewService(r *Repository) *Service { return &Service{repo: r} }

func (s *Service) Create(ctx context.Context, a *m.Assignment) (*m.Assignment, error) {
	if a.TitleAR == "" {
		return nil, Err("title required")
	}
	if a.DueAt.IsZero() {
		return nil, Err("due_at required")
	}
	if a.MaxPoints <= 0 {
		a.MaxPoints = 100
	}
	return s.repo.Create(ctx, a)
}

func (s *Service) List(ctx context.Context, instructorID *string, subjectID *string) ([]m.Assignment, error) {
	return s.repo.List(ctx, instructorID, subjectID)
}

func (s *Service) Get(ctx context.Context, id string) (*m.Assignment, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Submit(ctx context.Context, sub *m.AssignmentSubmission) (*m.AssignmentSubmission, error) {
	if sub.AssignmentID == "" || sub.StudentID == "" {
		return nil, Err("assignment_id and student_id required")
	}
	return s.repo.CreateSubmission(ctx, sub)
}

func (s *Service) Grade(ctx context.Context, submissionID string, grade int, feedback *string) (*m.AssignmentSubmission, error) {
	if grade < 0 {
		grade = 0
	}
	return s.repo.GradeSubmission(ctx, submissionID, grade, feedback)
}

func (s *Service) ListSubmissions(ctx context.Context, assignmentID string) ([]m.AssignmentSubmission, error) {
	return s.repo.ListSubmissions(ctx, assignmentID)
}

type simpleErr string

func (e simpleErr) Error() string { return string(e) }
func Err(msg string) error        { return simpleErr(msg) }
