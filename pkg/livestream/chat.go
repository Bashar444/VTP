package livestream

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Bashar444/VTP/pkg/models"
	"github.com/google/uuid"
)

// ChatRepository handles chat database operations
type ChatRepository struct {
	db *sql.DB
}

// NewChatRepository creates a new chat repository
func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// CreateMessage creates a new chat message
func (r *ChatRepository) CreateMessage(ctx context.Context, msg *models.StreamChatMessage) error {
	msg.ID = uuid.New().String()
	msg.CreatedAt = time.Now()
	msg.UpdatedAt = time.Now()

	if msg.MessageType == "" {
		msg.MessageType = "text"
	}

	query := `
		INSERT INTO stream_chat_messages (
			id, session_id, user_id, message_type, content,
			is_pinned, is_answered, reply_to_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		msg.ID,
		msg.SessionID,
		msg.UserID,
		msg.MessageType,
		msg.Content,
		msg.IsPinned,
		msg.IsAnswered,
		nullString(msg.ReplyToID),
		msg.CreatedAt,
		msg.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

// GetMessages retrieves messages for a session
func (r *ChatRepository) GetMessages(ctx context.Context, sessionID string, messageType string, limit, offset int) ([]*models.StreamChatMessage, error) {
	baseQuery := `
		SELECT id, session_id, user_id, message_type, content,
			   is_pinned, is_answered, reply_to_id, created_at, updated_at
		FROM stream_chat_messages
		WHERE session_id = $1
	`

	args := []interface{}{sessionID}
	argCount := 1

	if messageType != "" {
		argCount++
		baseQuery += fmt.Sprintf(" AND message_type = $%d", argCount)
		args = append(args, messageType)
	}

	argCount++
	limitArg := argCount
	argCount++
	offsetArg := argCount

	query := fmt.Sprintf("%s ORDER BY created_at ASC LIMIT $%d OFFSET $%d", baseQuery, limitArg, offsetArg)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.StreamChatMessage
	for rows.Next() {
		msg := &models.StreamChatMessage{}
		var replyToID sql.NullString

		err := rows.Scan(
			&msg.ID,
			&msg.SessionID,
			&msg.UserID,
			&msg.MessageType,
			&msg.Content,
			&msg.IsPinned,
			&msg.IsAnswered,
			&replyToID,
			&msg.CreatedAt,
			&msg.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		if replyToID.Valid {
			msg.ReplyToID = &replyToID.String
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

// GetRecentMessages retrieves recent messages (for new joiners)
func (r *ChatRepository) GetRecentMessages(ctx context.Context, sessionID string, limit int) ([]*models.StreamChatMessage, error) {
	query := `
		SELECT id, session_id, user_id, message_type, content,
			   is_pinned, is_answered, reply_to_id, created_at, updated_at
		FROM stream_chat_messages
		WHERE session_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, sessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.StreamChatMessage
	for rows.Next() {
		msg := &models.StreamChatMessage{}
		var replyToID sql.NullString

		err := rows.Scan(
			&msg.ID,
			&msg.SessionID,
			&msg.UserID,
			&msg.MessageType,
			&msg.Content,
			&msg.IsPinned,
			&msg.IsAnswered,
			&replyToID,
			&msg.CreatedAt,
			&msg.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		if replyToID.Valid {
			msg.ReplyToID = &replyToID.String
		}

		messages = append(messages, msg)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// GetQuestions retrieves questions for a session
func (r *ChatRepository) GetQuestions(ctx context.Context, sessionID string, answeredOnly bool) ([]*models.StreamChatMessage, error) {
	query := `
		SELECT id, session_id, user_id, message_type, content,
			   is_pinned, is_answered, reply_to_id, created_at, updated_at
		FROM stream_chat_messages
		WHERE session_id = $1 AND message_type = 'question'
	`

	args := []interface{}{sessionID}

	if answeredOnly {
		query += " AND is_answered = TRUE"
	}

	query += " ORDER BY created_at ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get questions: %w", err)
	}
	defer rows.Close()

	var messages []*models.StreamChatMessage
	for rows.Next() {
		msg := &models.StreamChatMessage{}
		var replyToID sql.NullString

		err := rows.Scan(
			&msg.ID,
			&msg.SessionID,
			&msg.UserID,
			&msg.MessageType,
			&msg.Content,
			&msg.IsPinned,
			&msg.IsAnswered,
			&replyToID,
			&msg.CreatedAt,
			&msg.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan question: %w", err)
		}

		if replyToID.Valid {
			msg.ReplyToID = &replyToID.String
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

// MarkAsAnswered marks a question as answered
func (r *ChatRepository) MarkAsAnswered(ctx context.Context, messageID string) error {
	query := `
		UPDATE stream_chat_messages
		SET is_answered = TRUE, updated_at = $2
		WHERE id = $1 AND message_type = 'question'
	`

	_, err := r.db.ExecContext(ctx, query, messageID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to mark as answered: %w", err)
	}

	return nil
}

// PinMessage pins or unpins a message
func (r *ChatRepository) PinMessage(ctx context.Context, messageID string, pinned bool) error {
	query := `
		UPDATE stream_chat_messages
		SET is_pinned = $2, updated_at = $3
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, messageID, pinned, time.Now())
	if err != nil {
		return fmt.Errorf("failed to pin message: %w", err)
	}

	return nil
}

// GetPinnedMessages retrieves pinned messages for a session
func (r *ChatRepository) GetPinnedMessages(ctx context.Context, sessionID string) ([]*models.StreamChatMessage, error) {
	query := `
		SELECT id, session_id, user_id, message_type, content,
			   is_pinned, is_answered, reply_to_id, created_at, updated_at
		FROM stream_chat_messages
		WHERE session_id = $1 AND is_pinned = TRUE
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pinned messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.StreamChatMessage
	for rows.Next() {
		msg := &models.StreamChatMessage{}
		var replyToID sql.NullString

		err := rows.Scan(
			&msg.ID,
			&msg.SessionID,
			&msg.UserID,
			&msg.MessageType,
			&msg.Content,
			&msg.IsPinned,
			&msg.IsAnswered,
			&replyToID,
			&msg.CreatedAt,
			&msg.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pinned message: %w", err)
		}

		if replyToID.Valid {
			msg.ReplyToID = &replyToID.String
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

// DeleteMessage deletes a chat message
func (r *ChatRepository) DeleteMessage(ctx context.Context, messageID string) error {
	query := `DELETE FROM stream_chat_messages WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, messageID)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}
	return nil
}
