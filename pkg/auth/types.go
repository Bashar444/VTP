package auth

import (
	"fmt"
	"strings"
)

// Role represents a user role
type Role string

const (
	RoleStudent Role = "student"
	RoleTeacher Role = "teacher"
	RoleAdmin   Role = "admin"
)

// String returns the string representation of a role
func (r Role) String() string {
	return string(r)
}

// IsValid checks if a role is valid
func (r Role) IsValid() bool {
	switch r {
	case RoleStudent, RoleTeacher, RoleAdmin:
		return true
	default:
		return false
	}
}

// ParseRole parses a string to a Role
func ParseRole(s string) (Role, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	role := Role(s)
	if !role.IsValid() {
		return "", fmt.Errorf("invalid role: %s", s)
	}
	return role, nil
}

// AllRoles returns all valid roles
func AllRoles() []Role {
	return []Role{RoleStudent, RoleTeacher, RoleAdmin}
}

// UserProfile represents public user profile information
type UserProfile struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
}

// AuthContext represents authenticated user context
type AuthContext struct {
	UserID string
	Email  string
	Role   Role
}

// IsTeacherOrAdmin checks if user is teacher or admin
func (ac *AuthContext) IsTeacherOrAdmin() bool {
	return ac.Role == RoleTeacher || ac.Role == RoleAdmin
}

// IsAdmin checks if user is admin
func (ac *AuthContext) IsAdmin() bool {
	return ac.Role == RoleAdmin
}

// IsStudent checks if user is student
func (ac *AuthContext) IsStudent() bool {
	return ac.Role == RoleStudent
}

// PermissionChecker provides role-based permission checks
type PermissionChecker struct {
	userRole Role
}

// NewPermissionChecker creates a new permission checker
func NewPermissionChecker(userRole Role) *PermissionChecker {
	return &PermissionChecker{
		userRole: userRole,
	}
}

// CanCreateCourse checks if user can create courses (teachers and admins)
func (pc *PermissionChecker) CanCreateCourse() bool {
	return pc.userRole == RoleTeacher || pc.userRole == RoleAdmin
}

// CanCreateAssignment checks if user can create assignments (teachers and admins)
func (pc *PermissionChecker) CanCreateAssignment() bool {
	return pc.userRole == RoleTeacher || pc.userRole == RoleAdmin
}

// CanGradeAssignment checks if user can grade assignments (teachers and admins)
func (pc *PermissionChecker) CanGradeAssignment() bool {
	return pc.userRole == RoleTeacher || pc.userRole == RoleAdmin
}

// CanStartLiveClass checks if user can start live classes (teachers and admins)
func (pc *PermissionChecker) CanStartLiveClass() bool {
	return pc.userRole == RoleTeacher || pc.userRole == RoleAdmin
}

// CanManageUsers checks if user can manage other users (admins only)
func (pc *PermissionChecker) CanManageUsers() bool {
	return pc.userRole == RoleAdmin
}

// CanAccessAdminPanel checks if user can access admin panel (admins only)
func (pc *PermissionChecker) CanAccessAdminPanel() bool {
	return pc.userRole == RoleAdmin
}

// CanSubmitAssignment checks if user can submit assignments (students only)
func (pc *PermissionChecker) CanSubmitAssignment() bool {
	return pc.userRole == RoleStudent
}

// CanJoinLiveClass checks if user can join live classes (all roles)
func (pc *PermissionChecker) CanJoinLiveClass() bool {
	return true // All roles can join live classes
}

// CanViewCourse checks if user can view a course (all roles)
func (pc *PermissionChecker) CanViewCourse() bool {
	return true // All roles can view courses (if enrolled)
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

// Error returns the error message
func (ve *ValidationError) Error() string {
	return fmt.Sprintf("validation error on %s: %s", ve.Field, ve.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []*ValidationError

// Error returns all error messages
func (ves ValidationErrors) Error() string {
	if len(ves) == 0 {
		return "no validation errors"
	}
	var msgs []string
	for _, ve := range ves {
		msgs = append(msgs, ve.Error())
	}
	return strings.Join(msgs, "; ")
}

// Has checks if a field has a validation error
func (ves ValidationErrors) Has(field string) bool {
	for _, ve := range ves {
		if ve.Field == field {
			return true
		}
	}
	return false
}

// Get returns validation error for a field
func (ves ValidationErrors) Get(field string) *ValidationError {
	for _, ve := range ves {
		if ve.Field == field {
			return ve
		}
	}
	return nil
}

// Validator provides validation utilities
type Validator struct{}

// ValidateEmail validates email format
func (v *Validator) ValidateEmail(email string) error {
	if email == "" {
		return &ValidationError{Field: "email", Message: "email is required"}
	}
	if len(email) > 255 {
		return &ValidationError{Field: "email", Message: "email must not exceed 255 characters"}
	}
	// Basic email validation
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return &ValidationError{Field: "email", Message: "invalid email format"}
	}
	if !strings.Contains(parts[1], ".") {
		return &ValidationError{Field: "email", Message: "invalid email format"}
	}
	return nil
}

// ValidatePassword validates password format
func (v *Validator) ValidatePassword(password string) error {
	if password == "" {
		return &ValidationError{Field: "password", Message: "password is required"}
	}
	if len(password) < 8 {
		return &ValidationError{Field: "password", Message: "password must be at least 8 characters long"}
	}
	if len(password) > 72 {
		return &ValidationError{Field: "password", Message: "password must not exceed 72 characters"}
	}
	return nil
}

// ValidateFullName validates full name format
func (v *Validator) ValidateFullName(fullName string) error {
	if fullName == "" {
		return &ValidationError{Field: "full_name", Message: "full name is required"}
	}
	if len(fullName) > 255 {
		return &ValidationError{Field: "full_name", Message: "full name must not exceed 255 characters"}
	}
	return nil
}

// ValidatePhone validates phone format (optional but if provided, must be valid)
func (v *Validator) ValidatePhone(phone string) error {
	if phone == "" {
		// Phone is optional
		return nil
	}
	if len(phone) > 20 {
		return &ValidationError{Field: "phone", Message: "phone must not exceed 20 characters"}
	}
	// Very basic validation - just check it contains digits
	hasDigit := false
	for _, ch := range phone {
		if ch >= '0' && ch <= '9' {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		return &ValidationError{Field: "phone", Message: "phone must contain at least one digit"}
	}
	return nil
}

// ValidateRole validates role format
func (v *Validator) ValidateRole(role string) error {
	if role == "" {
		return &ValidationError{Field: "role", Message: "role is required"}
	}
	_, err := ParseRole(role)
	if err != nil {
		return &ValidationError{Field: "role", Message: err.Error()}
	}
	return nil
}

// ValidateRegistration validates all registration fields
func (v *Validator) ValidateRegistration(email, password, fullName, phone, role string) ValidationErrors {
	var errs ValidationErrors

	if err := v.ValidateEmail(email); err != nil {
		errs = append(errs, err.(*ValidationError))
	}
	if err := v.ValidatePassword(password); err != nil {
		errs = append(errs, err.(*ValidationError))
	}
	if err := v.ValidateFullName(fullName); err != nil {
		errs = append(errs, err.(*ValidationError))
	}
	if err := v.ValidatePhone(phone); err != nil {
		errs = append(errs, err.(*ValidationError))
	}
	if err := v.ValidateRole(role); err != nil {
		errs = append(errs, err.(*ValidationError))
	}

	return errs
}

// PasswordRequirements represents password security requirements
type PasswordRequirements struct {
	MinLength        int
	MaxLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireDigit     bool
	RequireSpecial   bool
}

// DefaultPasswordRequirements returns default password requirements
func DefaultPasswordRequirements() PasswordRequirements {
	return PasswordRequirements{
		MinLength:        8,
		MaxLength:        72,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
		RequireSpecial:   false,
	}
}

// TokenMetadata represents metadata about a token
type TokenMetadata struct {
	UserID    string
	Email     string
	Role      string
	IssuedAt  int64
	ExpiresAt int64
	Remaining int64 // Seconds until expiration
}

// LoginAuditLog represents a login audit entry
type LoginAuditLog struct {
	UserID    string
	Email     string
	Timestamp int64
	Success   bool
	IPAddress string
	UserAgent string
	ErrorMsg  string
}

// SessionInfo represents user session information
type SessionInfo struct {
	UserID       string
	Email        string
	Role         string
	LastLogin    int64
	SessionStart int64
	TokenExpires int64
	IsActive     bool
}

// Constants for auth configuration
const (
	// DefaultAccessTokenDuration in hours
	DefaultAccessTokenDuration = 24

	// DefaultRefreshTokenDuration in hours
	DefaultRefreshTokenDuration = 168 // 7 days

	// DefaultBcryptCost for password hashing
	DefaultBcryptCost = 12

	// MaxLoginAttempts before account lockout (for future implementation)
	MaxLoginAttempts = 5

	// LockoutDuration in seconds (for future implementation)
	LockoutDuration = 900 // 15 minutes
)
