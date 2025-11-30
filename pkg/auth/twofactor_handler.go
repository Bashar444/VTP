package auth

import (
	"encoding/json"
	"net/http"
)

// TwoFactorHandler handles 2FA HTTP requests
type TwoFactorHandler struct {
	service *TwoFactorService
}

// NewTwoFactorHandler creates a new 2FA handler
func NewTwoFactorHandler(service *TwoFactorService) *TwoFactorHandler {
	return &TwoFactorHandler{service: service}
}

// Setup2FARequest is the request body for setting up 2FA
type Setup2FARequest struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// Enable2FARequest is the request body for enabling 2FA
type Enable2FARequest struct {
	UserID string `json:"user_id"`
	Code   string `json:"code"`
}

// Verify2FARequest is the request body for verifying 2FA
type Verify2FARequest struct {
	UserID string `json:"user_id"`
	Code   string `json:"code"`
}

// Disable2FARequest is the request body for disabling 2FA
type Disable2FARequest struct {
	UserID   string `json:"user_id"`
	Password string `json:"password"`
}

// Setup2FA handles POST /api/v1/auth/2fa/setup
func (h *TwoFactorHandler) Setup2FA(w http.ResponseWriter, r *http.Request) {
	var req Setup2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserID == "" || req.Email == "" {
		respondError(w, http.StatusBadRequest, "user_id and email are required")
		return
	}

	resp, err := h.service.Setup2FA(r.Context(), req.UserID, req.Email)
	if err != nil {
		if err == ErrTOTPAlreadyEnabled {
			respondError(w, http.StatusConflict, "2FA is already enabled")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to setup 2FA")
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// Enable2FA handles POST /api/v1/auth/2fa/enable
func (h *TwoFactorHandler) Enable2FA(w http.ResponseWriter, r *http.Request) {
	var req Enable2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserID == "" || req.Code == "" {
		respondError(w, http.StatusBadRequest, "user_id and code are required")
		return
	}

	err := h.service.Enable2FA(r.Context(), req.UserID, req.Code)
	if err != nil {
		if err == ErrInvalidTOTPCode {
			respondError(w, http.StatusUnauthorized, "Invalid verification code")
			return
		}
		if err == ErrTOTPAlreadyEnabled {
			respondError(w, http.StatusConflict, "2FA is already enabled")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to enable 2FA")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "2FA enabled successfully"})
}

// Verify2FA handles POST /api/v1/auth/2fa/verify
func (h *TwoFactorHandler) Verify2FA(w http.ResponseWriter, r *http.Request) {
	var req Verify2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserID == "" || req.Code == "" {
		respondError(w, http.StatusBadRequest, "user_id and code are required")
		return
	}

	err := h.service.Verify2FA(r.Context(), req.UserID, req.Code)
	if err != nil {
		if err == ErrInvalidTOTPCode {
			respondError(w, http.StatusUnauthorized, "Invalid verification code")
			return
		}
		if err == ErrTOTPNotEnabled {
			respondError(w, http.StatusBadRequest, "2FA is not enabled for this user")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to verify 2FA")
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"verified": true})
}

// Disable2FA handles POST /api/v1/auth/2fa/disable
func (h *TwoFactorHandler) Disable2FA(w http.ResponseWriter, r *http.Request) {
	var req Disable2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserID == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "user_id and password are required")
		return
	}

	err := h.service.Disable2FA(r.Context(), req.UserID, req.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to disable 2FA")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "2FA disabled successfully"})
}

// GetBackupCodes handles GET /api/v1/auth/2fa/backup-codes
func (h *TwoFactorHandler) GetBackupCodes(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		respondError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	codes, err := h.service.GetBackupCodes(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get backup codes")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"backup_codes": codes,
		"count":        len(codes),
	})
}

// RegenerateBackupCodes handles POST /api/v1/auth/2fa/backup-codes/regenerate
func (h *TwoFactorHandler) RegenerateBackupCodes(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserID == "" {
		respondError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	codes, err := h.service.RegenerateBackupCodes(r.Context(), req.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to regenerate backup codes")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"backup_codes": codes,
		"message":      "Backup codes regenerated successfully",
	})
}

// Helper functions
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
