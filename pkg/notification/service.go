package notification

import (
	"context"
	"errors"
	"log"

	"github.com/Bashar444/VTP/pkg/models"
)

var (
	ErrUserIDRequired    = errors.New("user ID is required")
	ErrInvalidType       = errors.New("invalid notification type")
	ErrInvalidChannel    = errors.New("invalid notification channel")
	ErrSenderUnavailable = errors.New("notification sender unavailable")
)

// ValidTypes are the allowed notification types
var ValidTypes = map[string]bool{
	"info":       true,
	"warning":    true,
	"success":    true,
	"error":      true,
	"assignment": true,
	"meeting":    true,
	"grade":      true,
	"attendance": true,
}

// ValidChannels are the allowed notification channels
var ValidChannels = map[string]bool{
	"in_app": true,
	"email":  true,
	"sms":    true,
	"push":   true,
}

// EmailSender interface for sending emails
type EmailSender interface {
	SendEmail(ctx context.Context, to, subject, body string) error
}

// SMSSender interface for sending SMS
type SMSSender interface {
	SendSMS(ctx context.Context, phone, message string) error
}

// PushSender interface for sending push notifications
type PushSender interface {
	SendPush(ctx context.Context, userID, title, body string, data map[string]string) error
}

// Service handles notification business logic
type Service struct {
	repo        *Repository
	emailSender EmailSender
	smsSender   SMSSender
	pushSender  PushSender
	logger      *log.Logger
}

// NewService creates a new notification service
func NewService(repo *Repository, logger *log.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// WithEmailSender sets the email sender
func (s *Service) WithEmailSender(sender EmailSender) *Service {
	s.emailSender = sender
	return s
}

// WithSMSSender sets the SMS sender
func (s *Service) WithSMSSender(sender SMSSender) *Service {
	s.smsSender = sender
	return s
}

// WithPushSender sets the push notification sender
func (s *Service) WithPushSender(sender PushSender) *Service {
	s.pushSender = sender
	return s
}

// CreateNotification creates a new notification
func (s *Service) CreateNotification(ctx context.Context, n *models.Notification) error {
	if n.UserID == "" {
		return ErrUserIDRequired
	}
	if !ValidTypes[n.Type] {
		return ErrInvalidType
	}
	if !ValidChannels[n.Channel] {
		return ErrInvalidChannel
	}
	if n.DeliveryStatus == "" {
		n.DeliveryStatus = "pending"
	}

	if err := s.repo.Create(ctx, n); err != nil {
		return err
	}

	// Async send based on channel
	go s.deliverNotification(context.Background(), n)

	return nil
}

// deliverNotification handles the actual delivery of the notification
func (s *Service) deliverNotification(ctx context.Context, n *models.Notification) {
	var err error

	switch n.Channel {
	case "email":
		if s.emailSender != nil {
			// In production, you'd fetch the user's email
			err = s.emailSender.SendEmail(ctx, "", n.TitleAr, n.MessageAr)
		}
	case "sms":
		if s.smsSender != nil {
			// In production, you'd fetch the user's phone
			err = s.smsSender.SendSMS(ctx, "", n.MessageAr)
		}
	case "push":
		if s.pushSender != nil {
			err = s.pushSender.SendPush(ctx, n.UserID, n.TitleAr, n.MessageAr, nil)
		}
	case "in_app":
		// In-app notifications are already stored, just mark as delivered
		err = nil
	}

	status := "delivered"
	if err != nil {
		status = "failed"
		if s.logger != nil {
			s.logger.Printf("Failed to deliver notification %s: %v", n.ID, err)
		}
	}

	if updateErr := s.repo.UpdateDeliveryStatus(ctx, n.ID, status); updateErr != nil && s.logger != nil {
		s.logger.Printf("Failed to update notification status %s: %v", n.ID, updateErr)
	}
}

// GetNotification retrieves a notification by ID
func (s *Service) GetNotification(ctx context.Context, id string) (*models.Notification, error) {
	return s.repo.GetByID(ctx, id)
}

// GetUserNotifications retrieves notifications for a user
func (s *Service) GetUserNotifications(ctx context.Context, userID string, unreadOnly bool, page, pageSize int) ([]models.Notification, error) {
	if userID == "" {
		return nil, ErrUserIDRequired
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.ListByUser(ctx, userID, unreadOnly, pageSize, offset)
}

// MarkAsRead marks a notification as read
func (s *Service) MarkAsRead(ctx context.Context, id string) error {
	return s.repo.MarkAsRead(ctx, id)
}

// MarkAllAsRead marks all notifications for a user as read
func (s *Service) MarkAllAsRead(ctx context.Context, userID string) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}

// GetUnreadCount returns the count of unread notifications
func (s *Service) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	return s.repo.CountUnread(ctx, userID)
}

// DeleteNotification removes a notification
func (s *Service) DeleteNotification(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// NotifyAssignment sends assignment notifications to students
func (s *Service) NotifyAssignment(ctx context.Context, assignmentID, titleAr string, studentIDs []string) error {
	var notifications []models.Notification
	refType := "assignment"
	for _, studentID := range studentIDs {
		n := models.Notification{
			UserID:        studentID,
			TitleAr:       "واجب جديد",
			TitleEn:       "New Assignment",
			MessageAr:     "تم إضافة واجب جديد: " + titleAr,
			MessageEn:     "New assignment added: " + titleAr,
			Type:          "assignment",
			Channel:       "in_app",
			ReferenceType: &refType,
			ReferenceID:   &assignmentID,
		}
		notifications = append(notifications, n)
	}
	return s.repo.BulkCreate(ctx, notifications)
}

// NotifyMeeting sends meeting notifications to participants
func (s *Service) NotifyMeeting(ctx context.Context, meetingID, titleAr string, participantIDs []string) error {
	var notifications []models.Notification
	refType := "meeting"
	for _, participantID := range participantIDs {
		n := models.Notification{
			UserID:        participantID,
			TitleAr:       "حصة قادمة",
			TitleEn:       "Upcoming Class",
			MessageAr:     "لديك حصة قادمة: " + titleAr,
			MessageEn:     "You have an upcoming class: " + titleAr,
			Type:          "meeting",
			Channel:       "in_app",
			ReferenceType: &refType,
			ReferenceID:   &meetingID,
		}
		notifications = append(notifications, n)
	}
	return s.repo.BulkCreate(ctx, notifications)
}

// NotifyGrade sends grade notification to a student
func (s *Service) NotifyGrade(ctx context.Context, studentID, subjectName string, grade int, maxGrade int) error {
	gradeType := "grade"
	n := &models.Notification{
		UserID:        studentID,
		TitleAr:       "علامة جديدة",
		TitleEn:       "New Grade",
		MessageAr:     "تم تسجيل علامتك في " + subjectName,
		MessageEn:     "Your grade has been recorded in " + subjectName,
		Type:          "grade",
		Channel:       "in_app",
		ReferenceType: &gradeType,
	}
	return s.CreateNotification(ctx, n)
}

// NotifyAttendance sends attendance notification to a student/parent
func (s *Service) NotifyAttendance(ctx context.Context, studentID, status, date string) error {
	attType := "attendance"
	statusAr := map[string]string{
		"present": "حاضر",
		"absent":  "غائب",
		"late":    "متأخر",
		"excused": "غياب مبرر",
	}
	n := &models.Notification{
		UserID:        studentID,
		TitleAr:       "سجل الحضور",
		TitleEn:       "Attendance Record",
		MessageAr:     "تم تسجيل حضورك: " + statusAr[status] + " - " + date,
		MessageEn:     "Attendance recorded: " + status + " - " + date,
		Type:          "attendance",
		Channel:       "in_app",
		ReferenceType: &attType,
	}
	return s.CreateNotification(ctx, n)
}
