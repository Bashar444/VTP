package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// CachedCourseService wraps course service with caching
type CachedCourseService struct {
	cache *RedisCache
}

// NewCachedCourseService creates a new cached course service
func NewCachedCourseService(cache *RedisCache) *CachedCourseService {
	return &CachedCourseService{cache: cache}
}

// GetCourse retrieves a course from cache or DB
func (s *CachedCourseService) GetCourse(ctx context.Context, courseID string, fetchFunc func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	cacheKey := CacheKey{}.Course(courseID)

	// Try cache first
	var course interface{}
	if err := s.cache.GetJSON(ctx, cacheKey, &course); err == nil && course != nil {
		return course, nil
	}

	// Cache miss - fetch from DB
	course, err := fetchFunc(ctx)
	if err != nil {
		return nil, err
	}

	// Store in cache
	s.cache.SetJSON(ctx, cacheKey, course, ExpirationMedium)

	return course, nil
}

// InvalidateCourse removes course from cache
func (s *CachedCourseService) InvalidateCourse(ctx context.Context, courseID string) error {
	cacheKey := CacheKey{}.Course(courseID)
	return s.cache.Delete(ctx, cacheKey)
}

// CachedInstructorService wraps instructor service with caching
type CachedInstructorService struct {
	cache *RedisCache
}

// NewCachedInstructorService creates a new cached instructor service
func NewCachedInstructorService(cache *RedisCache) *CachedInstructorService {
	return &CachedInstructorService{cache: cache}
}

// GetInstructor retrieves an instructor from cache or DB
func (s *CachedInstructorService) GetInstructor(ctx context.Context, instructorID string, fetchFunc func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	cacheKey := CacheKey{}.Instructor(instructorID)

	// Try cache first
	var instructor interface{}
	if err := s.cache.GetJSON(ctx, cacheKey, &instructor); err == nil && instructor != nil {
		return instructor, nil
	}

	// Cache miss - fetch from DB
	instructor, err := fetchFunc(ctx)
	if err != nil {
		return nil, err
	}

	// Store in cache
	s.cache.SetJSON(ctx, cacheKey, instructor, ExpirationMedium)

	return instructor, nil
}

// InvalidateInstructor removes instructor from cache
func (s *CachedInstructorService) InvalidateInstructor(ctx context.Context, instructorID string) error {
	cacheKey := CacheKey{}.Instructor(instructorID)
	return s.cache.Delete(ctx, cacheKey)
}

// SessionStore manages user sessions in Redis
type SessionStore struct {
	cache *RedisCache
}

// NewSessionStore creates a new session store
func NewSessionStore(cache *RedisCache) *SessionStore {
	return &SessionStore{cache: cache}
}

// SessionData represents session information
type SessionData struct {
	UserID    string                 `json:"user_id"`
	Email     string                 `json:"email"`
	Role      string                 `json:"role"`
	CreatedAt time.Time              `json:"created_at"`
	LastSeen  time.Time              `json:"last_seen"`
	Data      map[string]interface{} `json:"data"`
}

// CreateSession creates a new session
func (s *SessionStore) CreateSession(ctx context.Context, sessionID string, data *SessionData) error {
	data.CreatedAt = time.Now()
	data.LastSeen = time.Now()
	cacheKey := CacheKey{}.Session(sessionID)
	return s.cache.SetJSON(ctx, cacheKey, data, ExpirationSession)
}

// GetSession retrieves a session
func (s *SessionStore) GetSession(ctx context.Context, sessionID string) (*SessionData, error) {
	cacheKey := CacheKey{}.Session(sessionID)
	var data SessionData
	if err := s.cache.GetJSON(ctx, cacheKey, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// UpdateSession updates session data and extends expiration
func (s *SessionStore) UpdateSession(ctx context.Context, sessionID string, data *SessionData) error {
	data.LastSeen = time.Now()
	cacheKey := CacheKey{}.Session(sessionID)
	return s.cache.SetJSON(ctx, cacheKey, data, ExpirationSession)
}

// DeleteSession removes a session
func (s *SessionStore) DeleteSession(ctx context.Context, sessionID string) error {
	cacheKey := CacheKey{}.Session(sessionID)
	return s.cache.Delete(ctx, cacheKey)
}

// ExtendSession extends session expiration
func (s *SessionStore) ExtendSession(ctx context.Context, sessionID string) error {
	cacheKey := CacheKey{}.Session(sessionID)
	return s.cache.Expire(ctx, cacheKey, ExpirationSession)
}

// RateLimiter implements rate limiting using Redis
type RateLimiter struct {
	cache *RedisCache
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(cache *RedisCache) *RateLimiter {
	return &RateLimiter{cache: cache}
}

// Allow checks if an action is allowed based on rate limit
func (r *RateLimiter) Allow(ctx context.Context, key string, maxRequests int64, window time.Duration) (bool, error) {
	rateKey := fmt.Sprintf("ratelimit:%s", key)

	// Increment counter
	count, err := r.cache.IncrBy(ctx, rateKey, 1)
	if err != nil {
		return false, err
	}

	// Set expiration on first request
	if count == 1 {
		if err := r.cache.Expire(ctx, rateKey, window); err != nil {
			return false, err
		}
	}

	return count <= maxRequests, nil
}

// GetRemainingRequests returns remaining requests in current window
func (r *RateLimiter) GetRemainingRequests(ctx context.Context, key string, maxRequests int64) (int64, error) {
	rateKey := fmt.Sprintf("ratelimit:%s", key)
	count, err := r.cache.IncrBy(ctx, rateKey, 0)
	if err != nil {
		return 0, err
	}
	remaining := maxRequests - count
	if remaining < 0 {
		remaining = 0
	}
	return remaining, nil
}

// GetJSON is a helper for cache-aside pattern
func GetJSON[T any](ctx context.Context, cache *RedisCache, key string, fetchFunc func(ctx context.Context) (*T, error), expiration time.Duration) (*T, error) {
	// Try cache
	var result T
	err := cache.GetJSON(ctx, key, &result)
	if err == nil {
		// Check if result is not zero value
		data, _ := json.Marshal(result)
		if string(data) != "{}" && string(data) != "null" {
			return &result, nil
		}
	}

	// Cache miss - fetch from source
	data, err := fetchFunc(ctx)
	if err != nil {
		return nil, err
	}

	// Store in cache
	cache.SetJSON(ctx, key, data, expiration)

	return data, nil
}
