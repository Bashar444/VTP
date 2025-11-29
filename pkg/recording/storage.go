package recording

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// StorageBackend defines the interface for recording storage
type StorageBackend interface {
	// Upload uploads recording file to storage
	Upload(ctx context.Context, recordingID uuid.UUID, filePath string) error
	// Download downloads recording file from storage
	Download(ctx context.Context, recordingID uuid.UUID, writer io.Writer) error
	// Delete deletes recording file from storage
	Delete(ctx context.Context, recordingID uuid.UUID) error
	// Exists checks if recording exists in storage
	Exists(ctx context.Context, recordingID uuid.UUID) (bool, error)
	// GetSize gets file size in bytes
	GetSize(ctx context.Context, recordingID uuid.UUID) (int64, error)
	// GetURL gets public URL for recording (if supported)
	GetURL(ctx context.Context, recordingID uuid.UUID, expiresIn time.Duration) (string, error)
}

// LocalStorageBackend implements StorageBackend using local filesystem
type LocalStorageBackend struct {
	basePath string
	logger   *log.Logger
}

// NewLocalStorageBackend creates a new local storage backend
func NewLocalStorageBackend(basePath string, logger *log.Logger) (*LocalStorageBackend, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorageBackend{
		basePath: basePath,
		logger:   logger,
	}, nil
}

// getFilePath returns the full file path for a recording
func (b *LocalStorageBackend) getFilePath(recordingID uuid.UUID) string {
	return filepath.Join(b.basePath, recordingID.String()+".webm")
}

// Upload copies file from local path to storage
func (b *LocalStorageBackend) Upload(ctx context.Context, recordingID uuid.UUID, filePath string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Open source file
	src, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dstPath := b.getFilePath(recordingID)
	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(dstPath) // Clean up on error
		return fmt.Errorf("failed to copy file: %w", err)
	}

	b.logger.Printf("Recording uploaded: %s (%s)", recordingID, dstPath)
	return nil
}

// Download copies file from storage to writer
func (b *LocalStorageBackend) Download(ctx context.Context, recordingID uuid.UUID, writer io.Writer) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	filePath := b.getFilePath(recordingID)
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("recording not found: %s", recordingID)
		}
		return fmt.Errorf("failed to open recording: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(writer, file); err != nil {
		return fmt.Errorf("failed to download recording: %w", err)
	}

	b.logger.Printf("Recording downloaded: %s", recordingID)
	return nil
}

// Delete removes recording file from storage
func (b *LocalStorageBackend) Delete(ctx context.Context, recordingID uuid.UUID) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	filePath := b.getFilePath(recordingID)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete recording: %w", err)
	}

	b.logger.Printf("Recording deleted: %s", recordingID)
	return nil
}

// Exists checks if recording file exists
func (b *LocalStorageBackend) Exists(ctx context.Context, recordingID uuid.UUID) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	filePath := b.getFilePath(recordingID)
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// GetSize returns file size in bytes
func (b *LocalStorageBackend) GetSize(ctx context.Context, recordingID uuid.UUID) (int64, error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	filePath := b.getFilePath(recordingID)
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("recording not found: %s", recordingID)
		}
		return 0, fmt.Errorf("failed to get file size: %w", err)
	}

	return fileInfo.Size(), nil
}

// GetURL returns local file path as URL (not truly public)
func (b *LocalStorageBackend) GetURL(ctx context.Context, recordingID uuid.UUID, expiresIn time.Duration) (string, error) {
	filePath := b.getFilePath(recordingID)
	// For local storage, just return the file path
	// In production with S3/Azure, this would return a pre-signed URL
	return filePath, nil
}

// StorageManager manages recording storage operations
type StorageManager struct {
	backend StorageBackend
	db      *sql.DB
	logger  *log.Logger
}

// NewStorageManager creates a new storage manager
func NewStorageManager(backend StorageBackend, db *sql.DB, logger *log.Logger) *StorageManager {
	return &StorageManager{
		backend: backend,
		db:      db,
		logger:  logger,
	}
}

// UploadRecording uploads completed recording to storage
func (m *StorageManager) UploadRecording(ctx context.Context, recordingID uuid.UUID, filePath string) error {
	if err := m.backend.Upload(ctx, recordingID, filePath); err != nil {
		m.logger.Printf("Failed to upload recording %s: %v", recordingID, err)
		return err
	}

	// Update database with storage status
	query := `
		UPDATE recordings
		SET storage_status = $1, storage_uploaded_at = $2, updated_at = $3
		WHERE id = $4
	`

	now := time.Now().UTC()
	_, err := m.db.ExecContext(ctx, query, "uploaded", now, now, recordingID)
	if err != nil {
		m.logger.Printf("Failed to update storage status: %v", err)
		// Don't return error - upload succeeded even if DB update failed
	}

	m.logger.Printf("Recording storage complete: %s", recordingID)
	return nil
}

// DownloadRecording downloads recording from storage
func (m *StorageManager) DownloadRecording(ctx context.Context, recordingID uuid.UUID, writer io.Writer) error {
	// Check if recording exists in storage
	exists, err := m.backend.Exists(ctx, recordingID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("recording not found in storage: %s", recordingID)
	}

	// Log download access
	query := `
		INSERT INTO recording_access_log (id, recording_id, user_id, action, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	// User ID would come from context in real implementation
	userID := uuid.Nil
	_, _ = m.db.ExecContext(ctx, query, uuid.New(), recordingID, userID, "download", time.Now().UTC())

	return m.backend.Download(ctx, recordingID, writer)
}

// DeleteRecording removes recording from storage
func (m *StorageManager) DeleteRecording(ctx context.Context, recordingID uuid.UUID) error {
	return m.backend.Delete(ctx, recordingID)
}

// CleanupExpiredRecordings removes recordings older than retention period
func (m *StorageManager) CleanupExpiredRecordings(ctx context.Context, retentionDays int) error {
	query := `
		SELECT id FROM recordings
		WHERE status = 'archived' 
		AND stopped_at < NOW() - INTERVAL '1 day' * $1
		AND deleted_at IS NULL
		LIMIT 100
	`

	rows, err := m.db.QueryContext(ctx, query, retentionDays)
	if err != nil {
		m.logger.Printf("Failed to query expired recordings: %v", err)
		return err
	}
	defer rows.Close()

	deletedCount := 0
	for rows.Next() {
		var recordingID uuid.UUID
		if err := rows.Scan(&recordingID); err != nil {
			m.logger.Printf("Failed to scan recording ID: %v", err)
			continue
		}

		// Delete from storage
		if err := m.backend.Delete(ctx, recordingID); err != nil {
			m.logger.Printf("Failed to delete recording from storage: %v", err)
			continue
		}

		// Mark as deleted in database
		updateQuery := `
			UPDATE recordings
			SET deleted_at = $1, updated_at = $2
			WHERE id = $3
		`

		now := time.Now().UTC()
		if _, err := m.db.ExecContext(ctx, updateQuery, now, now, recordingID); err != nil {
			m.logger.Printf("Failed to mark recording as deleted: %v", err)
		}

		deletedCount++
	}

	m.logger.Printf("Cleanup complete: %d recordings deleted", deletedCount)
	return rows.Err()
}

// GetRecordingURL returns a download URL for the recording
func (m *StorageManager) GetRecordingURL(ctx context.Context, recordingID uuid.UUID, expiresIn time.Duration) (string, error) {
	return m.backend.GetURL(ctx, recordingID, expiresIn)
}

// GetRecordingSize returns the size of a recording
func (m *StorageManager) GetRecordingSize(ctx context.Context, recordingID uuid.UUID) (int64, error) {
	return m.backend.GetSize(ctx, recordingID)
}
