package notification

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Bashar444/VTP/pkg/models"
	"github.com/google/uuid"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
)

// Repository handles notification data persistence
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new notification repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create inserts a new notification
func (r *Repository) Create(ctx context.Context, n *models.Notification) error {
	if n.ID == "" {
		n.ID = uuid.New().String()
	}
	n.CreatedAt = time.Now()

	query := `
		INSERT INTO notifications (
			id, user_id, title_ar, title_en, message_ar, message_en,
			type, channel, reference_type, reference_id,
			is_read, delivery_status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.ExecContext(ctx, query,
		n.ID, n.UserID, n.TitleAr, n.TitleEn, n.MessageAr, n.MessageEn,
		n.Type, n.Channel, n.ReferenceType, n.ReferenceID,
		n.IsRead, n.DeliveryStatus, n.CreatedAt,
	)
	return err
}

// GetByID retrieves a notification by ID
func (r *Repository) GetByID(ctx context.Context, id string) (*models.Notification, error) {
	query := `
		SELECT id, user_id, title_ar, title_en, message_ar, message_en,
			type, channel, reference_type, reference_id,
			is_read, read_at, sent_at, delivery_status, created_at
		FROM notifications WHERE id = $1
	`
	var n models.Notification
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&n.ID, &n.UserID, &n.TitleAr, &n.TitleEn, &n.MessageAr, &n.MessageEn,
		&n.Type, &n.Channel, &n.ReferenceType, &n.ReferenceID,
		&n.IsRead, &n.ReadAt, &n.SentAt, &n.DeliveryStatus, &n.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotificationNotFound
	}
	return &n, err
}

// ListByUser retrieves notifications for a user
func (r *Repository) ListByUser(ctx context.Context, userID string, unreadOnly bool, limit, offset int) ([]models.Notification, error) {
	query := `
		SELECT id, user_id, title_ar, title_en, message_ar, message_en,
			type, channel, reference_type, reference_id,
			is_read, read_at, sent_at, delivery_status, created_at
		FROM notifications
		WHERE user_id = $1
	`
	args := []interface{}{userID}

	if unreadOnly {
		query += " AND is_read = false"
	}

	query += " ORDER BY created_at DESC LIMIT $2 OFFSET $3"
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var n models.Notification
		if err := rows.Scan(
			&n.ID, &n.UserID, &n.TitleAr, &n.TitleEn, &n.MessageAr, &n.MessageEn,
			&n.Type, &n.Channel, &n.ReferenceType, &n.ReferenceID,
			&n.IsRead, &n.ReadAt, &n.SentAt, &n.DeliveryStatus, &n.CreatedAt,
		); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}

// MarkAsRead marks a notification as read
func (r *Repository) MarkAsRead(ctx context.Context, id string) error {
	now := time.Now()
	query := `UPDATE notifications SET is_read = true, read_at = $2 WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id, now)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotificationNotFound
	}
	return nil
}

// MarkAllAsRead marks all notifications for a user as read
func (r *Repository) MarkAllAsRead(ctx context.Context, userID string) error {
	now := time.Now()
	query := `UPDATE notifications SET is_read = true, read_at = $2 WHERE user_id = $1 AND is_read = false`
	_, err := r.db.ExecContext(ctx, query, userID, now)
	return err
}

// UpdateDeliveryStatus updates the delivery status of a notification
func (r *Repository) UpdateDeliveryStatus(ctx context.Context, id, status string) error {
	now := time.Now()
	query := `UPDATE notifications SET delivery_status = $2, sent_at = $3 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, status, now)
	return err
}

// CountUnread counts unread notifications for a user
func (r *Repository) CountUnread(ctx context.Context, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	return count, err
}

// Delete removes a notification
func (r *Repository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM notifications WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotificationNotFound
	}
	return nil
}

// BulkCreate creates multiple notifications
func (r *Repository) BulkCreate(ctx context.Context, notifications []models.Notification) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO notifications (
			id, user_id, title_ar, title_en, message_ar, message_en,
			type, channel, reference_type, reference_id,
			is_read, delivery_status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()
	for _, n := range notifications {
		if n.ID == "" {
			n.ID = uuid.New().String()
		}
		_, err = stmt.ExecContext(ctx,
			n.ID, n.UserID, n.TitleAr, n.TitleEn, n.MessageAr, n.MessageEn,
			n.Type, n.Channel, n.ReferenceType, n.ReferenceID,
			false, "pending", now,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
