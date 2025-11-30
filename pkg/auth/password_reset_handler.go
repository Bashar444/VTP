package auth

import (
	"encoding/json"
	"net/http"
)

// PasswordResetHandler handles password reset HTTP requests
type PasswordResetHandler struct {
	service *PasswordResetService
}

// NewPasswordResetHandler creates a new password reset handler
func NewPasswordResetHandler(service *PasswordResetService) *PasswordResetHandler {
	return &PasswordResetHandler{service: service}
}

// RequestResetRequest is the request body for requesting a password reset
type RequestResetRequest struct {
	Email string `json:"email"`
}

// VerifyTokenRequest is the request body for verifying a reset token
type VerifyTokenRequest struct {
	Token string `json:"token"`
}

// ResetPasswordRequest is the request body for resetting password
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// RequestPasswordReset handles POST /api/v1/auth/forgot-password
func (h *PasswordResetHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req RequestResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" {
		respondError(w, http.StatusBadRequest, "Email is required")
		return
	}

	// Get client info for logging
	ipAddress := getClientIP(r)
	userAgent := r.UserAgent()

	token, err := h.service.RequestPasswordReset(r.Context(), req.Email, ipAddress, userAgent)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to process request")
		return
	}

	// In production, send email with token
	// For now, return token in response (remove in production!)
	response := map[string]string{
		"message": "If the email exists, a password reset link has been sent",
	}

	// TODO: Remove this in production - only for testing
	if token != "" {
		response["token"] = token // Only for development
		response["reset_url"] = "/reset-password?token=" + token
	}

	respondJSON(w, http.StatusOK, response)
}

// VerifyResetToken handles POST /api/v1/auth/verify-reset-token
func (h *PasswordResetHandler) VerifyResetToken(w http.ResponseWriter, r *http.Request) {
	var req VerifyTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Token == "" {
		respondError(w, http.StatusBadRequest, "Token is required")
		return
	}

	userID, err := h.service.VerifyResetToken(r.Context(), req.Token)
	if err != nil {
		if err == ErrInvalidResetToken || err == ErrTokenAlreadyUsed {
			respondError(w, http.StatusUnauthorized, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to verify token")
		return
	}

	// Get email for confirmation
	email, _ := h.service.GetUserEmail(r.Context(), req.Token)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"valid":   true,
		"user_id": userID,
		"email":   email,
	})
}

// ResetPassword handles POST /api/v1/auth/reset-password
func (h *PasswordResetHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		respondError(w, http.StatusBadRequest, "Token and new password are required")
		return
	}

	// Validate password strength
	if len(req.NewPassword) < 8 {
		respondError(w, http.StatusBadRequest, "Password must be at least 8 characters")
		return
	}

	err := h.service.ResetPassword(r.Context(), req.Token, req.NewPassword)
	if err != nil {
		if err == ErrInvalidResetToken || err == ErrTokenAlreadyUsed {
			respondError(w, http.StatusUnauthorized, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to reset password")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Password reset successfully",
	})
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies/load balancers)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
