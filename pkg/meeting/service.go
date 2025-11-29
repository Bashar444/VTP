package meeting

import (
	"context"
	"fmt"
	"time"

	"github.com/Bashar444/VTP/pkg/models"
)

// Service handles business logic for meetings
type Service struct {
	repo *Repository
}

// NewService creates a new meeting service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateMeeting creates a new meeting with conflict detection
func (s *Service) CreateMeeting(ctx context.Context, meeting *models.Meeting) error {
	// Validate required fields
	if meeting.InstructorID == "" {
		return fmt.Errorf("instructor ID is required")
	}
	if meeting.SubjectID == "" {
		return fmt.Errorf("subject ID is required")
	}
	if meeting.TitleAr == "" {
		return fmt.Errorf("title is required")
	}
	if meeting.Duration <= 0 {
		return fmt.Errorf("duration must be positive")
	}
	if meeting.ScheduledAt.Before(time.Now()) {
		return fmt.Errorf("meeting cannot be scheduled in the past")
	}

	// Check for time conflicts
	hasConflict, err := s.repo.CheckTimeConflict(ctx, meeting.InstructorID, meeting.ScheduledAt, meeting.Duration, "")
	if err != nil {
		return fmt.Errorf("failed to check time conflict: %w", err)
	}
	if hasConflict {
		return ErrTimeConflict
	}

	return s.repo.Create(ctx, meeting)
}

// GetMeeting retrieves a meeting by ID
func (s *Service) GetMeeting(ctx context.Context, id string) (*models.Meeting, error) {
	return s.repo.GetByID(ctx, id)
}

// ListMeetings retrieves meetings with filters
func (s *Service) ListMeetings(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*models.Meeting, error) {
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

// UpdateMeeting updates a meeting with conflict detection
func (s *Service) UpdateMeeting(ctx context.Context, meeting *models.Meeting) error {
	if meeting.ID == "" {
		return fmt.Errorf("meeting ID is required")
	}
	if meeting.TitleAr == "" {
		return fmt.Errorf("title is required")
	}
	if meeting.Duration <= 0 {
		return fmt.Errorf("duration must be positive")
	}

	// Get existing meeting
	existing, err := s.repo.GetByID(ctx, meeting.ID)
	if err != nil {
		return err
	}

	// Check for time conflicts if time changed
	if !meeting.ScheduledAt.Equal(existing.ScheduledAt) || meeting.Duration != existing.Duration {
		hasConflict, err := s.repo.CheckTimeConflict(ctx, existing.InstructorID, meeting.ScheduledAt, meeting.Duration, meeting.ID)
		if err != nil {
			return fmt.Errorf("failed to check time conflict: %w", err)
		}
		if hasConflict {
			return ErrTimeConflict
		}
	}

	return s.repo.Update(ctx, meeting)
}

// DeleteMeeting deletes a meeting
func (s *Service) DeleteMeeting(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// CancelMeeting cancels a meeting
func (s *Service) CancelMeeting(ctx context.Context, id string) error {
	meeting, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if meeting.Status == "completed" {
		return fmt.Errorf("cannot cancel completed meeting")
	}
	if meeting.Status == "cancelled" {
		return fmt.Errorf("meeting is already cancelled")
	}

	meeting.Status = "cancelled"
	return s.repo.Update(ctx, meeting)
}

// CompleteMeeting marks a meeting as completed
func (s *Service) CompleteMeeting(ctx context.Context, id string) error {
	meeting, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if meeting.Status == "cancelled" {
		return fmt.Errorf("cannot complete cancelled meeting")
	}
	if meeting.Status == "completed" {
		return fmt.Errorf("meeting is already completed")
	}

	now := time.Now()
	meeting.Status = "completed"
	meeting.EndTime = &now
	return s.repo.Update(ctx, meeting)
}

// GetInstructorMeetings retrieves meetings for an instructor
func (s *Service) GetInstructorMeetings(ctx context.Context, instructorID string, status string, page, pageSize int) ([]*models.Meeting, error) {
	filters := map[string]interface{}{
		"instructor_id": instructorID,
	}
	if status != "" {
		filters["status"] = status
	}
	return s.ListMeetings(ctx, filters, page, pageSize)
}

// GetStudentMeetings retrieves meetings for a student
func (s *Service) GetStudentMeetings(ctx context.Context, studentID string, status string, page, pageSize int) ([]*models.Meeting, error) {
	filters := map[string]interface{}{
		"student_id": studentID,
	}
	if status != "" {
		filters["status"] = status
	}
	return s.ListMeetings(ctx, filters, page, pageSize)
}

// GetUpcomingMeetings retrieves upcoming meetings
func (s *Service) GetUpcomingMeetings(ctx context.Context, instructorID string, page, pageSize int) ([]*models.Meeting, error) {
	filters := map[string]interface{}{
		"instructor_id": instructorID,
		"status":        "scheduled",
		"from_date":     time.Now(),
	}
	return s.ListMeetings(ctx, filters, page, pageSize)
}

// RescheduleMeeting reschedules a meeting
func (s *Service) RescheduleMeeting(ctx context.Context, id string, newScheduledAt time.Time, newDuration int) error {
	meeting, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if meeting.Status == "completed" {
		return fmt.Errorf("cannot reschedule completed meeting")
	}
	if meeting.Status == "cancelled" {
		return fmt.Errorf("cannot reschedule cancelled meeting")
	}

	if newScheduledAt.Before(time.Now()) {
		return fmt.Errorf("meeting cannot be scheduled in the past")
	}

	// Check for conflicts
	hasConflict, err := s.repo.CheckTimeConflict(ctx, meeting.InstructorID, newScheduledAt, newDuration, meeting.ID)
	if err != nil {
		return fmt.Errorf("failed to check time conflict: %w", err)
	}
	if hasConflict {
		return ErrTimeConflict
	}

	meeting.ScheduledAt = newScheduledAt
	meeting.Duration = newDuration
	return s.repo.Update(ctx, meeting)
}
