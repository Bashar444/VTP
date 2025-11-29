package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenClaims represents the JWT claims for a user
type TokenClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// TokenPair represents the access and refresh token pair
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// TokenService handles JWT token operations
type TokenService struct {
	secret          string
	accessDuration  time.Duration
	refreshDuration time.Duration
}

// NewTokenService creates a new token service
func NewTokenService(secret string, accessHours, refreshHours int) *TokenService {
	return &TokenService{
		secret:          secret,
		accessDuration:  time.Duration(accessHours) * time.Hour,
		refreshDuration: time.Duration(refreshHours) * time.Hour,
	}
}

// GenerateTokenPair generates both access and refresh tokens
func (ts *TokenService) GenerateTokenPair(userID, email, role string) (*TokenPair, error) {
	if userID == "" || email == "" || role == "" {
		return nil, errors.New("userID, email, and role are required")
	}

	// Generate access token
	accessToken, err := ts.generateToken(userID, email, role, ts.accessDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := ts.generateToken(userID, email, role, ts.refreshDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(ts.accessDuration.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// generateToken generates a JWT token with specified duration
func (ts *TokenService) generateToken(userID, email, role string, duration time.Duration) (string, error) {
	now := time.Now()
	expiresAt := now.Add(duration)

	claims := &TokenClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "vtp-platform",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(ts.secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns its claims
func (ts *TokenService) ValidateToken(tokenString string) (*TokenClaims, error) {
	if tokenString == "" {
		return nil, errors.New("token is empty")
	}

	claims := &TokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(ts.secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token using a refresh token
func (ts *TokenService) RefreshAccessToken(refreshTokenString string) (string, error) {
	claims, err := ts.ValidateToken(refreshTokenString)
	if err != nil {
		return "", fmt.Errorf("refresh token validation failed: %w", err)
	}

	// Generate new access token with same user info
	newAccessToken, err := ts.generateToken(claims.UserID, claims.Email, claims.Role, ts.accessDuration)
	if err != nil {
		return "", fmt.Errorf("failed to generate new access token: %w", err)
	}

	return newAccessToken, nil
}

// GetRemainingTime returns the time until token expiration in seconds
func (ts *TokenService) GetRemainingTime(tokenString string) (int64, error) {
	claims, err := ts.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}

	if claims.ExpiresAt == nil {
		return 0, errors.New("token has no expiration time")
	}

	remaining := time.Until(claims.ExpiresAt.Time).Seconds()
	if remaining < 0 {
		return 0, errors.New("token already expired")
	}

	return int64(remaining), nil
}
