package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"
)

var (
	ErrInvalidResetToken = errors.New("invalid or expired reset token")
	ErrTokenAlreadyUsed  = errors.New("reset token already used")
)

// PasswordResetService handles password reset operations
type PasswordResetService struct {
	db              *sql.DB
	passwordService *PasswordService
	tokenExpiry     time.Duration
	emailSender     interface {
		Send(to, subject, bodyHTML, bodyText string) error
	}
}

// NewPasswordResetService creates a new password reset service
func NewPasswordResetService(db *sql.DB, passwordService *PasswordService, tokenExpiryHours int) *PasswordResetService {
	return &PasswordResetService{
		db:              db,
		passwordService: passwordService,
		tokenExpiry:     time.Duration(tokenExpiryHours) * time.Hour,
	}
}

// WithEmailSender attaches an email sender to the service
func (s *PasswordResetService) WithEmailSender(sender interface {
	Send(to, subject, bodyHTML, bodyText string) error
}) *PasswordResetService {
	s.emailSender = sender
	return s
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
	IPAddress string
	UserAgent string
}

// RequestPasswordReset creates a reset token for a user
func (s *PasswordResetService) RequestPasswordReset(ctx context.Context, email, ipAddress, userAgent string) (string, error) {
	// Check if user exists
	var userID string
	err := s.db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Don't reveal if email exists - return success anyway for security
			return "", nil
		}
		return "", fmt.Errorf("failed to query user: %w", err)
	}

	// Generate secure random token
	token, err := generateResetToken(32)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Store token in database
	expiresAt := time.Now().Add(s.tokenExpiry)
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO password_reset_tokens (user_id, token, expires_at, ip_address, user_agent)
		 VALUES ($1, $2, $3, $4, $5)`,
		userID, token, expiresAt, ipAddress, userAgent,
	)
	if err != nil {
		return "", fmt.Errorf("failed to store reset token: %w", err)
	}

	// Send email if sender configured
	if s.emailSender != nil {
		resetBase := os.Getenv("RESET_BASE_URL")
		if resetBase == "" {
			resetBase = "https://example.com/reset-password"
		}
		link := fmt.Sprintf("%s?token=%s", resetBase, token)
		subject := "Reset your VTP account password"
		bodyText := fmt.Sprintf("We received a request to reset your password.\n\nUse this link: %s\n\nThe link expires in %d hours. If you did not request this, you can ignore this email.", link, int(s.tokenExpiry.Hours()))
		bodyHTML := fmt.Sprintf("<p>We received a request to reset your password.</p><p><a href=\"%s\">Click here to reset</a></p><p>This link expires in %d hours. If you did not request this, you can ignore this email.</p>", link, int(s.tokenExpiry.Hours()))
		_ = s.emailSender.Send(email, subject, bodyHTML, bodyText)
	}

	return token, nil
}

// VerifyResetToken checks if a reset token is valid
func (s *PasswordResetService) VerifyResetToken(ctx context.Context, token string) (string, error) {
	var userID string
	var expiresAt time.Time
	var usedAt *time.Time

	err := s.db.QueryRowContext(ctx,
		`SELECT user_id, expires_at, used_at FROM password_reset_tokens WHERE token = $1`,
		token,
	).Scan(&userID, &expiresAt, &usedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrInvalidResetToken
		}
		return "", fmt.Errorf("failed to query token: %w", err)
	}

	// Check if already used
	if usedAt != nil {
		return "", ErrTokenAlreadyUsed
	}

	// Check if expired
	if time.Now().After(expiresAt) {
		return "", ErrInvalidResetToken
	}

	return userID, nil
}

// ResetPassword resets a user's password using a valid token
func (s *PasswordResetService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Verify token
	userID, err := s.VerifyResetToken(ctx, token)
	if err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := s.passwordService.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update password
	now := time.Now()
	_, err = tx.ExecContext(ctx,
		`UPDATE users SET password_hash = $1, password_changed_at = $2, updated_at = $3 WHERE id = $4`,
		hashedPassword, now, now, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	_, err = tx.ExecContext(ctx,
		`UPDATE password_reset_tokens SET used_at = $1 WHERE token = $2`,
		now, token,
	)
	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CleanupExpiredTokens removes expired reset tokens
func (s *PasswordResetService) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM password_reset_tokens WHERE expires_at < $1`,
		time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup tokens: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}

// GetUserEmail retrieves email for a reset token (for confirmation)
func (s *PasswordResetService) GetUserEmail(ctx context.Context, token string) (string, error) {
	var email string
	err := s.db.QueryRowContext(ctx,
		`SELECT u.email FROM users u 
		 JOIN password_reset_tokens prt ON u.id = prt.user_id 
		 WHERE prt.token = $1`,
		token,
	).Scan(&email)

	if err != nil {
		return "", fmt.Errorf("failed to get email: %w", err)
	}

	return email, nil
}

// generateResetToken creates a cryptographically secure random token
func generateResetToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
