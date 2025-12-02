package attendance

import (
	"context"
	"errors"
	"time"

	"github.com/Bashar444/VTP/pkg/models"
)

var (
	ErrInvalidStatus     = errors.New("invalid attendance status")
	ErrInvalidDateRange  = errors.New("invalid date range")
	ErrStudentIDRequired = errors.New("student ID is required")
)

// Service handles attendance business logic
type Service struct {
	repo *Repository
}

// NewService creates a new attendance service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ValidStatuses are the allowed attendance statuses
var ValidStatuses = map[string]bool{
	"present": true,
	"absent":  true,
	"late":    true,
	"excused": true,
}

// RecordAttendance records a student's attendance
func (s *Service) RecordAttendance(ctx context.Context, a *models.Attendance) error {
	if a.StudentID == "" {
		return ErrStudentIDRequired
	}
	if !ValidStatuses[a.Status] {
		return ErrInvalidStatus
	}
	if a.Date.IsZero() {
		a.Date = time.Now()
	}
	return s.repo.Create(ctx, a)
}

// UpdateAttendance updates an existing attendance record
func (s *Service) UpdateAttendance(ctx context.Context, a *models.Attendance) error {
	if !ValidStatuses[a.Status] {
		return ErrInvalidStatus
	}
	return s.repo.Update(ctx, a)
}

// GetAttendance retrieves an attendance record by ID
func (s *Service) GetAttendance(ctx context.Context, id string) (*models.Attendance, error) {
	return s.repo.GetByID(ctx, id)
}

// GetStudentAttendance retrieves attendance records for a student within a date range
func (s *Service) GetStudentAttendance(ctx context.Context, studentID string, startDate, endDate time.Time) ([]models.Attendance, error) {
	if studentID == "" {
		return nil, ErrStudentIDRequired
	}
	if endDate.Before(startDate) {
		return nil, ErrInvalidDateRange
	}
	return s.repo.ListByStudent(ctx, studentID, startDate, endDate)
}

// GetClassAttendance retrieves attendance for a class section on a specific date
func (s *Service) GetClassAttendance(ctx context.Context, classSectionID string, date time.Time) ([]models.Attendance, error) {
	return s.repo.ListByClassSection(ctx, classSectionID, date)
}

// GetMeetingAttendance retrieves attendance for a meeting
func (s *Service) GetMeetingAttendance(ctx context.Context, meetingID string) ([]models.Attendance, error) {
	return s.repo.ListByMeeting(ctx, meetingID)
}

// GetStudentStats retrieves attendance statistics for a student
func (s *Service) GetStudentStats(ctx context.Context, studentID string, startDate, endDate time.Time) (*AttendanceStats, error) {
	if studentID == "" {
		return nil, ErrStudentIDRequired
	}
	if endDate.Before(startDate) {
		return nil, ErrInvalidDateRange
	}
	return s.repo.GetStudentStats(ctx, studentID, startDate, endDate)
}

// BulkRecordAttendance records attendance for multiple students
func (s *Service) BulkRecordAttendance(ctx context.Context, records []models.Attendance) error {
	for _, a := range records {
		if a.StudentID == "" {
			return ErrStudentIDRequired
		}
		if !ValidStatuses[a.Status] {
			return ErrInvalidStatus
		}
	}
	return s.repo.BulkCreate(ctx, records)
}

// AutoMarkPresent marks a student as present when they join a meeting
func (s *Service) AutoMarkPresent(ctx context.Context, studentID, meetingID string) error {
	return s.repo.MarkPresent(ctx, studentID, meetingID)
}

// AttendanceReport holds attendance report data
type AttendanceReport struct {
	StudentID   string              `json:"student_id"`
	StudentName string              `json:"student_name"`
	Stats       *AttendanceStats    `json:"stats"`
	Records     []models.Attendance `json:"records,omitempty"`
}

// GenerateReport generates an attendance report for a student
func (s *Service) GenerateReport(ctx context.Context, studentID string, startDate, endDate time.Time, includeRecords bool) (*AttendanceReport, error) {
	stats, err := s.GetStudentStats(ctx, studentID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	report := &AttendanceReport{
		StudentID: studentID,
		Stats:     stats,
	}

	if includeRecords {
		records, err := s.GetStudentAttendance(ctx, studentID, startDate, endDate)
		if err != nil {
			return nil, err
		}
		report.Records = records
	}

	return report, nil
}
