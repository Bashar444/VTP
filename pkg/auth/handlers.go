package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest represents the request body for token refresh
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// ChangePasswordRequest represents the request body for password change
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// RegisterResponse represents the response after successful registration
type RegisterResponse struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
	Message  string `json:"message"`
}

// LoginResponse represents the response after successful login
type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
	TokenType    string       `json:"token_type"`
	User         UserResponse `json:"user"`
}

// UserResponse represents user data in API responses (without sensitive info)
type UserResponse struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
}

// RefreshResponse represents the response after token refresh
type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	userStore       *UserStore
	tokenService    *TokenService
	passwordService *PasswordService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userStore *UserStore, tokenService *TokenService, passwordService *PasswordService) *AuthHandler {
	return &AuthHandler{
		userStore:       userStore,
		tokenService:    tokenService,
		passwordService: passwordService,
	}
}

// RegisterHandler handles POST /api/v1/auth/register
func (ah *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Validate request method
	if r.Method != http.MethodPost {
		ah.respondError(w, http.StatusMethodNotAllowed, "INVALID_METHOD", "Only POST requests are allowed")
		return
	}

	// Parse request body
	var req RegisterRequest
	if err := ah.parseJSONBody(r, &req); err != nil {
		ah.respondError(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" || req.FullName == "" || req.Role == "" {
		ah.respondError(w, http.StatusBadRequest, "MISSING_FIELDS", "email, password, full_name, and role are required")
		return
	}

	// Validate email format (basic)
	if !isValidEmail(req.Email) {
		ah.respondError(w, http.StatusBadRequest, "INVALID_EMAIL", "Invalid email format")
		return
	}

	// Validate password strength
	if err := ah.passwordService.ValidatePassword(req.Password); err != nil {
		ah.respondError(w, http.StatusBadRequest, "WEAK_PASSWORD", err.Error())
		return
	}

	// Create user
	user, err := ah.userStore.CreateUser(r.Context(), req.Email, req.Phone, req.FullName, req.Role, req.Password)
	if err != nil {
		if strings.Contains(err.Error(), "email already registered") {
			ah.respondError(w, http.StatusConflict, "EMAIL_EXISTS", "Email already registered")
		} else {
			ah.respondError(w, http.StatusInternalServerError, "REGISTRATION_FAILED", "Failed to register user")
		}
		return
	}

	// Prepare response
	resp := RegisterResponse{
		UserID:   user.ID,
		Email:    user.Email,
		FullName: user.FullName,
		Role:     user.Role,
		Message:  "User registered successfully",
	}

	ah.respondSuccess(w, http.StatusCreated, resp)
}

// LoginHandler handles POST /api/v1/auth/login
func (ah *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Validate request method
	if r.Method != http.MethodPost {
		ah.respondError(w, http.StatusMethodNotAllowed, "INVALID_METHOD", "Only POST requests are allowed")
		return
	}

	// Parse request body
	var req LoginRequest
	if err := ah.parseJSONBody(r, &req); err != nil {
		ah.respondError(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		ah.respondError(w, http.StatusBadRequest, "MISSING_FIELDS", "email and password are required")
		return
	}

	// Authenticate user
	user, err := ah.userStore.AuthenticateUser(r.Context(), req.Email, req.Password)
	if err != nil {
		ah.respondError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
		return
	}

	// Generate tokens
	tokenPair, err := ah.tokenService.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		ah.respondError(w, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", "Failed to generate tokens")
		return
	}

	// Update last login
	_ = ah.userStore.UpdateLastLogin(r.Context(), user.ID)

	// Prepare response
	resp := LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    tokenPair.TokenType,
		User: UserResponse{
			UserID:   user.ID,
			Email:    user.Email,
			FullName: user.FullName,
			Phone:    user.Phone,
			Role:     user.Role,
		},
	}

	ah.respondSuccess(w, http.StatusOK, resp)
}

// RefreshHandler handles POST /api/v1/auth/refresh
func (ah *AuthHandler) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	// Validate request method
	if r.Method != http.MethodPost {
		ah.respondError(w, http.StatusMethodNotAllowed, "INVALID_METHOD", "Only POST requests are allowed")
		return
	}

	// Parse request body
	var req RefreshRequest
	if err := ah.parseJSONBody(r, &req); err != nil {
		ah.respondError(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	// Validate refresh token
	if req.RefreshToken == "" {
		ah.respondError(w, http.StatusBadRequest, "MISSING_FIELDS", "refresh_token is required")
		return
	}

	// Generate new access token
	newAccessToken, err := ah.tokenService.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		ah.respondError(w, http.StatusUnauthorized, "TOKEN_INVALID", "Invalid or expired refresh token")
		return
	}

	// Prepare response
	resp := RefreshResponse{
		AccessToken: newAccessToken,
		ExpiresIn:   int64(ah.tokenService.accessDuration.Seconds()),
		TokenType:   "Bearer",
	}

	ah.respondSuccess(w, http.StatusOK, resp)
}

// ChangePasswordHandler handles POST /api/v1/auth/change-password
func (ah *AuthHandler) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	// Validate request method
	if r.Method != http.MethodPost {
		ah.respondError(w, http.StatusMethodNotAllowed, "INVALID_METHOD", "Only POST requests are allowed")
		return
	}

	// Extract user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		ah.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Parse request body
	var req ChangePasswordRequest
	if err := ah.parseJSONBody(r, &req); err != nil {
		ah.respondError(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}

	// Validate required fields
	if req.CurrentPassword == "" || req.NewPassword == "" {
		ah.respondError(w, http.StatusBadRequest, "MISSING_FIELDS", "current_password and new_password are required")
		return
	}

	// Change password
	err := ah.userStore.ChangePassword(r.Context(), userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		if strings.Contains(err.Error(), "invalid current password") {
			ah.respondError(w, http.StatusUnauthorized, "INVALID_PASSWORD", "Current password is incorrect")
		} else if strings.Contains(err.Error(), "validation failed") {
			ah.respondError(w, http.StatusBadRequest, "WEAK_PASSWORD", err.Error())
		} else {
			ah.respondError(w, http.StatusInternalServerError, "PASSWORD_CHANGE_FAILED", "Failed to change password")
		}
		return
	}

	ah.respondSuccess(w, http.StatusOK, map[string]string{"message": "Password changed successfully"})
}

// GetProfileHandler handles GET /api/v1/auth/profile
func (ah *AuthHandler) GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	// Validate request method
	if r.Method != http.MethodGet {
		ah.respondError(w, http.StatusMethodNotAllowed, "INVALID_METHOD", "Only GET requests are allowed")
		return
	}

	// Extract user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		ah.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Get user
	user, err := ah.userStore.GetUserByID(r.Context(), userID)
	if err != nil {
		ah.respondError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		return
	}

	// Prepare response
	resp := UserResponse{
		UserID:   user.ID,
		Email:    user.Email,
		FullName: user.FullName,
		Phone:    user.Phone,
		Role:     user.Role,
	}

	ah.respondSuccess(w, http.StatusOK, resp)
}

// Helper methods

// parseJSONBody parses JSON from request body
func (ah *AuthHandler) parseJSONBody(r *http.Request, v interface{}) error {
	// Limit request body size to 1MB
	r.Body = http.MaxBytesReader(nil, r.Body, 1024*1024)
	defer r.Body.Close()

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	// Unmarshal JSON
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return nil
}

// respondSuccess sends a successful response
func (ah *AuthHandler) respondSuccess(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// respondError sends an error response
func (ah *AuthHandler) respondError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error: message,
		Code:  code,
	})
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	// Simple validation: must contain @ and .
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}
	if !strings.Contains(parts[1], ".") {
		return false
	}
	return true
}
