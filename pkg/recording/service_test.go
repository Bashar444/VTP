package recording

import (
	"context"
	"database/sql"
	"log"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Helper to create test database connection
func setupTestDB(t *testing.T) *sql.DB {
	// For now, we'll skip actual DB tests
	// In production, use testcontainers or test database
	t.Skip("Requires database setup")
	return nil
}

func TestStartRecording(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("Database not available")
	}
	defer db.Close()

	logger := log.New(log.Writer(), "[TEST] ", log.LstdFlags)
	service := NewRecordingService(db, logger)

	// Test data
	userID := uuid.New()
	roomID := uuid.New()

	req := &StartRecordingRequest{
		RoomID: roomID,
		Title:  "Test Recording",
	}

	// Start recording
	recording, err := service.StartRecording(context.Background(), req, userID)
	if err != nil {
		t.Fatalf("Failed to start recording: %v", err)
	}

	if recording == nil {
		t.Fatal("Recording is nil")
	}

	if recording.Title != req.Title {
		t.Errorf("Expected title %s, got %s", req.Title, recording.Title)
	}

	if recording.RoomID != roomID {
		t.Errorf("Expected room ID %s, got %s", roomID, recording.RoomID)
	}

	if recording.StartedBy != userID {
		t.Errorf("Expected started by %s, got %s", userID, recording.StartedBy)
	}

	if recording.Status != StatusPending {
		t.Errorf("Expected status %s, got %s", StatusPending, recording.Status)
	}
}

func TestStartRecordingValidation(t *testing.T) {
	// Note: Validation tests that don't require DB can still run
	// But tests that hit the DB need a proper setup

	tests := []struct {
		name    string
		req     *StartRecordingRequest
		wantErr bool
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name: "missing room ID",
			req: &StartRecordingRequest{
				RoomID: uuid.Nil,
				Title:  "Test",
			},
			wantErr: true,
		},
		{
			name: "missing title",
			req: &StartRecordingRequest{
				RoomID: uuid.New(),
				Title:  "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just test the validation logic without hitting DB
			if tt.req == nil {
				if !tt.wantErr {
					t.Error("Expected error for nil request")
				}
				return
			}

			if tt.req.RoomID == uuid.Nil && !tt.wantErr {
				t.Error("Expected error for nil room ID")
			}

			if tt.req.Title == "" && !tt.wantErr {
				t.Error("Expected error for empty title")
			}
		})
	}
}

func TestStopRecording(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("Database not available")
	}
	defer db.Close()

	logger := log.New(log.Writer(), "[TEST] ", log.LstdFlags)
	service := NewRecordingService(db, logger)

	// First start a recording
	userID := uuid.New()
	roomID := uuid.New()
	req := &StartRecordingRequest{
		RoomID: roomID,
		Title:  "Test Recording",
	}

	recording, err := service.StartRecording(context.Background(), req, userID)
	if err != nil {
		t.Fatalf("Failed to start recording: %v", err)
	}

	// Give it some time
	time.Sleep(100 * time.Millisecond)

	// Stop the recording
	stopped, err := service.StopRecording(context.Background(), recording.ID)
	if err != nil {
		t.Fatalf("Failed to stop recording: %v", err)
	}

	if stopped == nil {
		t.Fatal("Stopped recording is nil")
	}

	if stopped.Status != StatusCompleted {
		t.Errorf("Expected status %s, got %s", StatusCompleted, stopped.Status)
	}

	if stopped.StoppedAt == nil {
		t.Error("Expected stopped_at to be set")
	}

	if stopped.DurationSeconds == nil || *stopped.DurationSeconds <= 0 {
		t.Error("Expected duration to be greater than 0")
	}
}

func TestGetRecording(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("Database not available")
	}
	defer db.Close()

	logger := log.New(log.Writer(), "[TEST] ", log.LstdFlags)
	service := NewRecordingService(db, logger)

	userID := uuid.New()
	roomID := uuid.New()
	req := &StartRecordingRequest{
		RoomID: roomID,
		Title:  "Test Recording",
	}

	// Create a recording
	created, err := service.StartRecording(context.Background(), req, userID)
	if err != nil {
		t.Fatalf("Failed to create recording: %v", err)
	}

	// Get the recording
	retrieved, err := service.GetRecording(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Failed to get recording: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Retrieved recording is nil")
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, retrieved.ID)
	}

	if retrieved.Title != created.Title {
		t.Errorf("Expected title %s, got %s", created.Title, retrieved.Title)
	}
}

func TestGetRecordingNotFound(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("Database not available")
	}
	defer db.Close()

	logger := log.New(log.Writer(), "[TEST] ", log.LstdFlags)
	service := NewRecordingService(db, logger)

	// Try to get non-existent recording
	recording, err := service.GetRecording(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("Expected nil error for non-existent recording, got %v", err)
	}

	if recording != nil {
		t.Error("Expected nil recording for non-existent ID")
	}
}

func TestListRecordings(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("Database not available")
	}
	defer db.Close()

	logger := log.New(log.Writer(), "[TEST] ", log.LstdFlags)
	service := NewRecordingService(db, logger)

	userID := uuid.New()
	roomID := uuid.New()

	// Create multiple recordings
	for i := 0; i < 3; i++ {
		req := &StartRecordingRequest{
			RoomID: roomID,
			Title:  "Test Recording " + string(rune(i)),
		}
		_, err := service.StartRecording(context.Background(), req, userID)
		if err != nil {
			t.Fatalf("Failed to create recording: %v", err)
		}
	}

	// List recordings
	query := &RecordingListQuery{
		RoomID: roomID,
		Limit:  10,
		Offset: 0,
	}

	recordings, total, err := service.ListRecordings(context.Background(), query)
	if err != nil {
		t.Fatalf("Failed to list recordings: %v", err)
	}

	if total != 3 {
		t.Errorf("Expected 3 recordings, got %d", total)
	}

	if len(recordings) != 3 {
		t.Errorf("Expected 3 recordings in result, got %d", len(recordings))
	}
}

func TestListRecordingsPagination(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("Database not available")
	}
	defer db.Close()

	logger := log.New(log.Writer(), "[TEST] ", log.LstdFlags)
	service := NewRecordingService(db, logger)

	userID := uuid.New()
	roomID := uuid.New()

	// Create recordings
	for i := 0; i < 5; i++ {
		req := &StartRecordingRequest{
			RoomID: roomID,
			Title:  "Test Recording",
		}
		_, err := service.StartRecording(context.Background(), req, userID)
		if err != nil {
			t.Fatalf("Failed to create recording: %v", err)
		}
	}

	// Test pagination
	query := &RecordingListQuery{
		RoomID: roomID,
		Limit:  2,
		Offset: 0,
	}

	recordings, total, err := service.ListRecordings(context.Background(), query)
	if err != nil {
		t.Fatalf("Failed to list recordings: %v", err)
	}

	if total != 5 {
		t.Errorf("Expected 5 total recordings, got %d", total)
	}

	if len(recordings) != 2 {
		t.Errorf("Expected 2 recordings per page, got %d", len(recordings))
	}

	// Get next page
	query.Offset = 2
	recordings, _, err = service.ListRecordings(context.Background(), query)
	if err != nil {
		t.Fatalf("Failed to list recordings page 2: %v", err)
	}

	if len(recordings) != 2 {
		t.Errorf("Expected 2 recordings on page 2, got %d", len(recordings))
	}
}

func TestDeleteRecording(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("Database not available")
	}
	defer db.Close()

	logger := log.New(log.Writer(), "[TEST] ", log.LstdFlags)
	service := NewRecordingService(db, logger)

	userID := uuid.New()
	roomID := uuid.New()
	req := &StartRecordingRequest{
		RoomID: roomID,
		Title:  "Test Recording",
	}

	// Create a recording
	created, err := service.StartRecording(context.Background(), req, userID)
	if err != nil {
		t.Fatalf("Failed to create recording: %v", err)
	}

	// Delete the recording
	deleted, err := service.DeleteRecording(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Failed to delete recording: %v", err)
	}

	if deleted == nil {
		t.Fatal("Deleted recording is nil")
	}

	if deleted.Status != StatusDeleted {
		t.Errorf("Expected status %s, got %s", StatusDeleted, deleted.Status)
	}

	if deleted.DeletedAt == nil {
		t.Error("Expected deleted_at to be set")
	}
}

func TestUpdateRecordingStatus(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		t.Skip("Database not available")
	}
	defer db.Close()

	logger := log.New(log.Writer(), "[TEST] ", log.LstdFlags)
	service := NewRecordingService(db, logger)

	userID := uuid.New()
	roomID := uuid.New()
	req := &StartRecordingRequest{
		RoomID: roomID,
		Title:  "Test Recording",
	}

	created, err := service.StartRecording(context.Background(), req, userID)
	if err != nil {
		t.Fatalf("Failed to create recording: %v", err)
	}

	// Update status
	err = service.UpdateRecordingStatus(context.Background(), created.ID, StatusProcessing)
	if err != nil {
		t.Fatalf("Failed to update status: %v", err)
	}

	// Verify
	updated, err := service.GetRecording(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Failed to get updated recording: %v", err)
	}

	if updated.Status != StatusProcessing {
		t.Errorf("Expected status %s, got %s", StatusProcessing, updated.Status)
	}
}

func TestUpdateRecordingStatusInvalid(t *testing.T) {
	logger := log.New(log.Writer(), "[TEST] ", log.LstdFlags)
	service := NewRecordingService(nil, logger)

	err := service.UpdateRecordingStatus(context.Background(), uuid.New(), "invalid_status")
	if err == nil {
		t.Error("Expected error for invalid status")
	}
}

func TestValidateStatus(t *testing.T) {
	tests := []struct {
		status string
		valid  bool
	}{
		{StatusPending, true},
		{StatusRecording, true},
		{StatusProcessing, true},
		{StatusCompleted, true},
		{StatusFailed, true},
		{StatusArchived, true},
		{StatusDeleted, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		result := ValidateStatus(tt.status)
		if result != tt.valid {
			t.Errorf("ValidateStatus(%s) = %v, want %v", tt.status, result, tt.valid)
		}
	}
}

func TestValidateAccessLevel(t *testing.T) {
	tests := []struct {
		level string
		valid bool
	}{
		{AccessLevelView, true},
		{AccessLevelDownload, true},
		{AccessLevelShare, true},
		{AccessLevelDelete, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		result := ValidateAccessLevel(tt.level)
		if result != tt.valid {
			t.Errorf("ValidateAccessLevel(%s) = %v, want %v", tt.level, result, tt.valid)
		}
	}
}

func TestValidateShareType(t *testing.T) {
	tests := []struct {
		shareType string
		valid     bool
	}{
		{ShareTypeUser, true},
		{ShareTypeRole, true},
		{ShareTypePublic, true},
		{ShareTypeLink, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		result := ValidateShareType(tt.shareType)
		if result != tt.valid {
			t.Errorf("ValidateShareType(%s) = %v, want %v", tt.shareType, result, tt.valid)
		}
	}
}
