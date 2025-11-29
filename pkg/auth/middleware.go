package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

// AuthMiddleware validates JWT tokens and adds user context to requests
type AuthMiddleware struct {
	tokenService *TokenService
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(tokenService *TokenService) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
	}
}

// Middleware wraps an HTTP handler with JWT validation
func (am *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		token, err := am.extractToken(r)
		if err != nil {
			am.respondError(w, http.StatusUnauthorized, "MISSING_TOKEN", err.Error())
			return
		}

		// Validate token
		claims, err := am.tokenService.ValidateToken(token)
		if err != nil {
			am.respondError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Token validation failed: "+err.Error())
			return
		}

		// Add user context to request
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_email", claims.Email)
		ctx = context.WithValue(ctx, "user_role", claims.Role)

		// Call next handler with new context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuthMiddleware validates JWT tokens but doesn't require them
// Useful for endpoints that work with or without authentication
func (am *AuthMiddleware) OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to extract token from Authorization header
		token, err := am.extractToken(r)
		ctx := r.Context()

		if err == nil && token != "" {
			// Token exists, validate it
			claims, err := am.tokenService.ValidateToken(token)
			if err == nil {
				// Token is valid, add user context
				ctx = context.WithValue(ctx, "user_id", claims.UserID)
				ctx = context.WithValue(ctx, "user_email", claims.Email)
				ctx = context.WithValue(ctx, "user_role", claims.Role)
				ctx = context.WithValue(ctx, "authenticated", true)
			}
		}

		// Call next handler (with or without user context)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RoleMiddleware checks if user has required role(s)
func (am *AuthMiddleware) RoleMiddleware(requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user role from context
			userRole, ok := r.Context().Value("user_role").(string)
			if !ok || userRole == "" {
				am.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
				return
			}

			// Check if user role is in required roles
			hasRequiredRole := false
			for _, role := range requiredRoles {
				if userRole == role {
					hasRequiredRole = true
					break
				}
			}

			if !hasRequiredRole {
				am.respondError(w, http.StatusForbidden, "INSUFFICIENT_PERMISSIONS", "User role does not have permission for this action")
				return
			}

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}

// extractToken extracts the JWT token from Authorization header
func (am *AuthMiddleware) extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header is missing")
	}

	// Expected format: "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format, expected 'Bearer <token>'")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errors.New("token is empty")
	}

	return token, nil
}

// respondError sends an error response
func (am *AuthMiddleware) respondError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(`{"error":"` + message + `","code":"` + code + `"}`))
}

// GetUserID extracts user ID from request context
func GetUserID(r *http.Request) (string, error) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		return "", errors.New("user_id not found in context")
	}
	return userID, nil
}

// GetUserEmail extracts user email from request context
func GetUserEmail(r *http.Request) (string, error) {
	email, ok := r.Context().Value("user_email").(string)
	if !ok || email == "" {
		return "", errors.New("user_email not found in context")
	}
	return email, nil
}

// GetUserRole extracts user role from request context
func GetUserRole(r *http.Request) (string, error) {
	role, ok := r.Context().Value("user_role").(string)
	if !ok || role == "" {
		return "", errors.New("user_role not found in context")
	}
	return role, nil
}

// IsAuthenticated checks if user is authenticated
func IsAuthenticated(r *http.Request) bool {
	_, ok := r.Context().Value("user_id").(string)
	return ok
}

// HasRole checks if user has a specific role
func HasRole(r *http.Request, requiredRole string) bool {
	role, err := GetUserRole(r)
	return err == nil && role == requiredRole
}

// HasAnyRole checks if user has any of the specified roles
func HasAnyRole(r *http.Request, roles ...string) bool {
	userRole, err := GetUserRole(r)
	if err != nil {
		return false
	}
	for _, role := range roles {
		if userRole == role {
			return true
		}
	}
	return false
}
