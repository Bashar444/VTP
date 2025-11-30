package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	ErrTOTPNotEnabled     = errors.New("TOTP is not enabled for this user")
	ErrInvalidTOTPCode    = errors.New("invalid TOTP code")
	ErrTOTPAlreadyEnabled = errors.New("TOTP is already enabled")
)

// TwoFactorService handles 2FA operations
type TwoFactorService struct {
	db          *sql.DB
	totpService *TOTPService
}

// NewTwoFactorService creates a new 2FA service
func NewTwoFactorService(db *sql.DB, issuer string) *TwoFactorService {
	return &TwoFactorService{
		db:          db,
		totpService: NewTOTPService(issuer),
	}
}

// SetupTOTPRequest contains data for TOTP setup
type SetupTOTPRequest struct {
	UserID string
	Email  string
}

// SetupTOTPResponse contains TOTP setup information
type SetupTOTPResponse struct {
	Secret      string   `json:"secret"`
	QRCodeURL   string   `json:"qr_code_url"`
	BackupCodes []string `json:"backup_codes"`
}

// Setup2FA initializes 2FA for a user
func (s *TwoFactorService) Setup2FA(ctx context.Context, userID, email string) (*SetupTOTPResponse, error) {
	// Check if 2FA is already enabled
	var totpEnabled bool
	err := s.db.QueryRowContext(ctx, "SELECT totp_enabled FROM users WHERE id = $1", userID).Scan(&totpEnabled)
	if err != nil {
		return nil, fmt.Errorf("failed to check 2FA status: %w", err)
	}
	if totpEnabled {
		return nil, ErrTOTPAlreadyEnabled
	}

	// Generate TOTP secret
	secret, qrURL, err := s.totpService.GenerateSecret(email)
	if err != nil {
		return nil, err
	}

	// Generate backup codes
	backupCodes, err := s.totpService.GenerateBackupCodes(8)
	if err != nil {
		return nil, fmt.Errorf("failed to generate backup codes: %w", err)
	}

	// Store secret and backup codes (not yet enabled)
	backupCodesJSON, _ := json.Marshal(backupCodes)
	_, err = s.db.ExecContext(ctx,
		`UPDATE users SET totp_secret = $1, backup_codes = $2, updated_at = $3 WHERE id = $4`,
		secret, backupCodesJSON, time.Now(), userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to store TOTP secret: %w", err)
	}

	return &SetupTOTPResponse{
		Secret:      secret,
		QRCodeURL:   qrURL,
		BackupCodes: backupCodes,
	}, nil
}

// Enable2FA enables 2FA after verifying the initial code
func (s *TwoFactorService) Enable2FA(ctx context.Context, userID, code string) error {
	// Get user's TOTP secret
	var secret string
	var totpEnabled bool
	err := s.db.QueryRowContext(ctx,
		"SELECT totp_secret, totp_enabled FROM users WHERE id = $1",
		userID,
	).Scan(&secret, &totpEnabled)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if totpEnabled {
		return ErrTOTPAlreadyEnabled
	}

	if secret == "" {
		return errors.New("TOTP not set up - call Setup2FA first")
	}

	// Validate the code
	if !s.totpService.ValidateCodeWithWindow(secret, code, 1, 1) {
		return ErrInvalidTOTPCode
	}

	// Enable 2FA
	now := time.Now()
	_, err = s.db.ExecContext(ctx,
		`UPDATE users SET totp_enabled = $1, totp_verified_at = $2, last_totp_verified = $3, updated_at = $4 WHERE id = $5`,
		true, now, now, now, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to enable 2FA: %w", err)
	}

	return nil
}

// Verify2FA verifies a TOTP code for an enabled user
func (s *TwoFactorService) Verify2FA(ctx context.Context, userID, code string) error {
	// Get user's TOTP secret and status
	var secret string
	var totpEnabled bool
	var backupCodesJSON []byte

	err := s.db.QueryRowContext(ctx,
		"SELECT totp_secret, totp_enabled, backup_codes FROM users WHERE id = $1",
		userID,
	).Scan(&secret, &totpEnabled, &backupCodesJSON)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if !totpEnabled {
		return ErrTOTPNotEnabled
	}

	// Try TOTP code first
	if s.totpService.ValidateCodeWithWindow(secret, code, 1, 1) {
		// Update last verified time
		_, err = s.db.ExecContext(ctx,
			"UPDATE users SET last_totp_verified = $1 WHERE id = $2",
			time.Now(), userID,
		)
		return err
	}

	// Try backup codes
	var backupCodes []string
	if err := json.Unmarshal(backupCodesJSON, &backupCodes); err == nil {
		for i, backupCode := range backupCodes {
			if backupCode == code {
				// Remove used backup code
				backupCodes = append(backupCodes[:i], backupCodes[i+1:]...)
				newBackupCodesJSON, _ := json.Marshal(backupCodes)
				_, err = s.db.ExecContext(ctx,
					"UPDATE users SET backup_codes = $1, last_totp_verified = $2 WHERE id = $3",
					newBackupCodesJSON, time.Now(), userID,
				)
				return err
			}
		}
	}

	return ErrInvalidTOTPCode
}

// Disable2FA disables 2FA for a user
func (s *TwoFactorService) Disable2FA(ctx context.Context, userID, password string) error {
	// Verify password before disabling (security measure)
	// This would require password verification logic from auth service

	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET totp_enabled = $1, totp_secret = $2, totp_verified_at = $3, backup_codes = $4, updated_at = $5 WHERE id = $6`,
		false, "", nil, "[]", time.Now(), userID,
	)
	if err != nil {
		return fmt.Errorf("failed to disable 2FA: %w", err)
	}

	return nil
}

// GetBackupCodes retrieves remaining backup codes
func (s *TwoFactorService) GetBackupCodes(ctx context.Context, userID string) ([]string, error) {
	var backupCodesJSON []byte
	err := s.db.QueryRowContext(ctx,
		"SELECT backup_codes FROM users WHERE id = $1",
		userID,
	).Scan(&backupCodesJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup codes: %w", err)
	}

	var backupCodes []string
	if err := json.Unmarshal(backupCodesJSON, &backupCodes); err != nil {
		return nil, err
	}

	return backupCodes, nil
}

// RegenerateBackupCodes creates new backup codes
func (s *TwoFactorService) RegenerateBackupCodes(ctx context.Context, userID string) ([]string, error) {
	backupCodes, err := s.totpService.GenerateBackupCodes(8)
	if err != nil {
		return nil, err
	}

	backupCodesJSON, _ := json.Marshal(backupCodes)
	_, err = s.db.ExecContext(ctx,
		"UPDATE users SET backup_codes = $1, updated_at = $2 WHERE id = $3",
		backupCodesJSON, time.Now(), userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to store backup codes: %w", err)
	}

	return backupCodes, nil
}
