package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// User represents a user in the system
type User struct {
	ID           string    `db:"id" json:"id"`
	Email        string    `db:"email" json:"email"`
	Phone        string    `db:"phone" json:"phone"`
	FullName     string    `db:"full_name" json:"full_name"`
	Role         string    `db:"role" json:"role"`
	PasswordHash string    `db:"password_hash" json:"-"` // Never expose hash
	Locale       string    `db:"locale" json:"locale"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// UserStore handles database operations for users
type UserStore struct {
	db *sql.DB
	ps *PasswordService
}

// NewUserStore creates a new user store
func NewUserStore(db *sql.DB, ps *PasswordService) *UserStore {
	return &UserStore{
		db: db,
		ps: ps,
	}
}

// CreateUser creates a new user in the database
func (us *UserStore) CreateUser(ctx context.Context, email, phone, fullName, role, password string) (*User, error) {
	// Validate input
	if email == "" {
		return nil, errors.New("email is required")
	}
	if fullName == "" {
		return nil, errors.New("full name is required")
	}
	if role == "" {
		return nil, errors.New("role is required")
	}
	if password == "" {
		return nil, errors.New("password is required")
	}

	// Validate role
	if !isValidRole(role) {
		return nil, fmt.Errorf("invalid role: %s (must be student, teacher, or admin)", role)
	}

	// Validate password security
	if err := us.ps.ValidatePassword(password); err != nil {
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Hash password
	passwordHash, err := us.ps.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Default locale to Arabic (Syrian)
	locale := "ar_SY"

	// Insert user into database
	query := `
		INSERT INTO users (email, phone, full_name, role, password_hash, locale, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, email, phone, full_name, role, locale, created_at, updated_at
	`

	user := &User{}
	err = us.db.QueryRowContext(ctx, query, email, phone, fullName, role, passwordHash, locale).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.FullName,
		&user.Role,
		&user.Locale,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		// Check if it's a unique constraint violation (email already exists)
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return nil, errors.New("email already registered")
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email address
func (us *UserStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}

	query := `
		SELECT id, email, phone, full_name, role, password_hash, locale, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &User{}
	err := us.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.FullName,
		&user.Role,
		&user.PasswordHash,
		&user.Locale,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (us *UserStore) GetUserByID(ctx context.Context, userID string) (*User, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	query := `
		SELECT id, email, phone, full_name, role, password_hash, locale, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &User{}
	err := us.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.FullName,
		&user.Role,
		&user.PasswordHash,
		&user.Locale,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// AuthenticateUser validates email and password, returns user if valid
func (us *UserStore) AuthenticateUser(ctx context.Context, email, password string) (*User, error) {
	if email == "" || password == "" {
		return nil, errors.New("email and password are required")
	}

	// Get user from database
	user, err := us.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Verify password
	if err := us.ps.VerifyPassword(user.PasswordHash, password); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Clear sensitive data before returning
	user.PasswordHash = ""

	return user, nil
}

// UpdateLastLogin updates the user's last login timestamp
func (us *UserStore) UpdateLastLogin(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}

	query := `
		UPDATE users
		SET updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := us.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// UpdateUserProfile updates user profile information
func (us *UserStore) UpdateUserProfile(ctx context.Context, userID, phone, fullName string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}

	query := `
		UPDATE users
		SET phone = $1, full_name = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`

	result, err := us.db.ExecContext(ctx, query, phone, fullName, userID)
	if err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// ChangePassword changes a user's password
func (us *UserStore) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}
	if currentPassword == "" {
		return errors.New("current password is required")
	}
	if newPassword == "" {
		return errors.New("new password is required")
	}

	// Validate new password
	if err := us.ps.ValidatePassword(newPassword); err != nil {
		return fmt.Errorf("new password validation failed: %w", err)
	}

	// Get current password hash
	query := `SELECT password_hash FROM users WHERE id = $1`
	var currentHash string
	err := us.db.QueryRowContext(ctx, query, userID).Scan(&currentHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify current password
	if err := us.ps.VerifyPassword(currentHash, currentPassword); err != nil {
		return errors.New("invalid current password")
	}

	// Prevent using the same password
	if err := us.ps.VerifyPassword(currentHash, newPassword); err == nil {
		return errors.New("new password must be different from current password")
	}

	// Hash new password
	newHash, err := us.ps.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password in database
	updateQuery := `
		UPDATE users
		SET password_hash = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	result, err := us.db.ExecContext(ctx, updateQuery, newHash, userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// DeleteUser deletes a user from the database (soft or hard delete)
func (us *UserStore) DeleteUser(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}

	query := `DELETE FROM users WHERE id = $1`
	result, err := us.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// isValidRole checks if a role is valid
func isValidRole(role string) bool {
	switch role {
	case "student", "teacher", "admin":
		return true
	default:
		return false
	}
}

// GetUsersByRole retrieves all users with a specific role
func (us *UserStore) GetUsersByRole(ctx context.Context, role string) ([]*User, error) {
	if role == "" {
		return nil, errors.New("role is required")
	}

	if !isValidRole(role) {
		return nil, fmt.Errorf("invalid role: %s", role)
	}

	query := `
		SELECT id, email, phone, full_name, role, locale, created_at, updated_at
		FROM users
		WHERE role = $1
		ORDER BY created_at DESC
	`

	rows, err := us.db.QueryContext(ctx, query, role)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Phone,
			&user.FullName,
			&user.Role,
			&user.Locale,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}
