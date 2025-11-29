package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordService handles password hashing and verification
type PasswordService struct {
	cost int
}

// NewPasswordService creates a new password service with specified bcrypt cost
func NewPasswordService(cost int) *PasswordService {
	// Validate cost is within bcrypt acceptable range (4-31)
	if cost < 4 {
		cost = 12 // Default cost
	}
	if cost > 31 {
		cost = 31
	}

	return &PasswordService{
		cost: cost,
	}
}

// HashPassword hashes a plaintext password using bcrypt
func (ps *PasswordService) HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	// Check minimum password length
	if len(password) < 8 {
		return "", errors.New("password must be at least 8 characters long")
	}

	// Hash the password with bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(password), ps.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

// VerifyPassword compares a hashed password with a plaintext password
func (ps *PasswordService) VerifyPassword(hashedPassword, plainPassword string) error {
	if hashedPassword == "" {
		return errors.New("hashed password cannot be empty")
	}

	if plainPassword == "" {
		return errors.New("plain password cannot be empty")
	}

	// Compare the hashed password with the plaintext password
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return errors.New("invalid password")
		}
		return fmt.Errorf("password verification failed: %w", err)
	}

	return nil
}

// ValidatePassword checks if a password meets security requirements
func (ps *PasswordService) ValidatePassword(password string) error {
	if password == "" {
		return errors.New("password cannot be empty")
	}

	// Minimum length requirement
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	// Maximum length requirement (bcrypt max is 72 bytes)
	if len(password) > 72 {
		return errors.New("password must not exceed 72 characters")
	}

	// Check for at least one uppercase letter
	hasUpper := false
	for _, r := range password {
		if r >= 'A' && r <= 'Z' {
			hasUpper = true
			break
		}
	}
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}

	// Check for at least one lowercase letter
	hasLower := false
	for _, r := range password {
		if r >= 'a' && r <= 'z' {
			hasLower = true
			break
		}
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}

	// Check for at least one digit
	hasDigit := false
	for _, r := range password {
		if r >= '0' && r <= '9' {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}

	return nil
}

// IsPasswordSecure returns true if password meets all security requirements
func (ps *PasswordService) IsPasswordSecure(password string) bool {
	return ps.ValidatePassword(password) == nil
}
