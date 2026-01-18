package livestream

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Bashar444/VTP/pkg/models"
)

// Service handles live streaming business logic
type Service struct {
	repo   *Repository
	logger *log.Logger
}

// NewService creates a new live streaming service
func NewService(repo *Repository, logger *log.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// CreateSession creates a new streaming session
func (s *Service) CreateSession(ctx context.Context, session *models.StreamingSession) error {
	s.logger.Printf("Creating streaming session: %s (room: %s)", session.Title, session.RoomID)
	return s.repo.CreateSession(ctx, session)
}

// GetSession gets a session by ID
func (s *Service) GetSession(ctx context.Context, id string) (*models.StreamingSession, error) {
	return s.repo.GetSessionByID(ctx, id)
}

// GetSessionByRoomID gets a session by room ID
func (s *Service) GetSessionByRoomID(ctx context.Context, roomID string) (*models.StreamingSession, error) {
	return s.repo.GetSessionByRoomID(ctx, roomID)
}

// StartStream starts a streaming session
func (s *Service) StartStream(ctx context.Context, sessionID, userID string) (*models.StreamingSession, error) {
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if session.InstructorID != userID {
		return nil, ErrUnauthorized
	}

	// Check if already live
	if session.Status == "live" {
		return nil, ErrAlreadyLive
	}

	// Update status
	now := time.Now()
	session.Status = "live"
	session.StartedAt = &now

	if err := s.repo.UpdateSession(ctx, session); err != nil {
		return nil, err
	}

	s.logger.Printf("Stream started: %s (room: %s)", session.Title, session.RoomID)
	return session, nil
}

// StopStream stops a streaming session
func (s *Service) StopStream(ctx context.Context, sessionID, userID string) (*models.StreamingSession, error) {
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if session.InstructorID != userID {
		return nil, ErrUnauthorized
	}

	// Check if not live
	if session.Status != "live" {
		return nil, ErrNotLive
	}

	// Update status
	now := time.Now()
	session.Status = "ended"
	session.EndedAt = &now

	if err := s.repo.UpdateSession(ctx, session); err != nil {
		return nil, err
	}

	s.logger.Printf("Stream ended: %s (room: %s)", session.Title, session.RoomID)
	return session, nil
}

// StartRecording starts recording a stream
func (s *Service) StartRecording(ctx context.Context, sessionID, userID string) (*models.StreamingSession, error) {
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.InstructorID != userID {
		return nil, ErrUnauthorized
	}

	session.IsRecording = true
	if err := s.repo.UpdateSession(ctx, session); err != nil {
		return nil, err
	}

	s.logger.Printf("Recording started for stream: %s", session.Title)
	return session, nil
}

// StopRecording stops recording a stream
func (s *Service) StopRecording(ctx context.Context, sessionID, userID string) (*models.StreamingSession, error) {
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.InstructorID != userID {
		return nil, ErrUnauthorized
	}

	session.IsRecording = false
	if err := s.repo.UpdateSession(ctx, session); err != nil {
		return nil, err
	}

	s.logger.Printf("Recording stopped for stream: %s", session.Title)
	return session, nil
}

// ListSessions lists streaming sessions
func (s *Service) ListSessions(ctx context.Context, instructorID, status string, liveOnly bool, limit, offset int) ([]*models.StreamingSession, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return s.repo.ListSessions(ctx, instructorID, status, liveOnly, limit, offset)
}

// ListLiveSessions lists only live sessions
func (s *Service) ListLiveSessions(ctx context.Context, limit, offset int) ([]*models.StreamingSession, int, error) {
	return s.repo.ListSessions(ctx, "", "live", true, limit, offset)
}

// JoinSession adds a participant to a session
func (s *Service) JoinSession(ctx context.Context, sessionID, userID, role string) error {
	// Verify session exists and is live
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return err
	}

	if session.Status != "live" && session.Status != "scheduled" {
		return ErrNotLive
	}

	// Check participant count
	count, err := s.repo.GetParticipantCount(ctx, sessionID)
	if err != nil {
		return err
	}

	if count >= session.MaxParticipants {
		return errors.New("session is full")
	}

	participant := &models.StreamingParticipant{
		SessionID:      sessionID,
		UserID:         userID,
		Role:           role,
		IsMuted:        true,
		IsVideoEnabled: false,
	}

	return s.repo.AddParticipant(ctx, participant)
}

// LeaveSession removes a participant from a session
func (s *Service) LeaveSession(ctx context.Context, sessionID, userID string) error {
	return s.repo.RemoveParticipant(ctx, sessionID, userID)
}

// GetParticipants gets all participants in a session
func (s *Service) GetParticipants(ctx context.Context, sessionID string) ([]*models.StreamingParticipant, error) {
	return s.repo.GetParticipants(ctx, sessionID)
}

// GetParticipantCount gets the count of participants
func (s *Service) GetParticipantCount(ctx context.Context, sessionID string) (int, error) {
	return s.repo.GetParticipantCount(ctx, sessionID)
}

// UpdateSession updates a session
func (s *Service) UpdateSession(ctx context.Context, session *models.StreamingSession) error {
	return s.repo.UpdateSession(ctx, session)
}
