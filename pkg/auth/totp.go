package auth

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTPService handles two-factor authentication operations
type TOTPService struct {
	issuer string
}

// NewTOTPService creates a new TOTP service
func NewTOTPService(issuer string) *TOTPService {
	return &TOTPService{
		issuer: issuer,
	}
}

// GenerateSecret creates a new TOTP secret for a user
func (s *TOTPService) GenerateSecret(email string) (string, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: email,
		Period:      30,
		SecretSize:  20,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	return key.Secret(), key.URL(), nil
}

// ValidateCode verifies a TOTP code against a secret
func (s *TOTPService) ValidateCode(secret, code string) bool {
	return totp.Validate(code, secret)
}

// ValidateCodeWithWindow validates with a time window for clock skew
func (s *TOTPService) ValidateCodeWithWindow(secret, code string, windowBefore, windowAfter int) bool {
	// Check current time
	if totp.Validate(code, secret) {
		return true
	}

	// Check previous periods
	for i := 1; i <= windowBefore; i++ {
		pastTime := time.Now().Add(-time.Duration(i*30) * time.Second)
		if validateAtTime(secret, code, pastTime) {
			return true
		}
	}

	// Check future periods
	for i := 1; i <= windowAfter; i++ {
		futureTime := time.Now().Add(time.Duration(i*30) * time.Second)
		if validateAtTime(secret, code, futureTime) {
			return true
		}
	}

	return false
}

// validateAtTime validates code at a specific time
func validateAtTime(secret, code string, t time.Time) bool {
	validCode, err := totp.GenerateCode(secret, t)
	if err != nil {
		return false
	}
	return validCode == code
}

// GenerateBackupCodes creates backup codes for account recovery
func (s *TOTPService) GenerateBackupCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		code, err := generateRandomCode(8)
		if err != nil {
			return nil, err
		}
		codes[i] = code
	}
	return codes, nil
}

// generateRandomCode creates a random alphanumeric code
func generateRandomCode(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes)[:length], nil
}
