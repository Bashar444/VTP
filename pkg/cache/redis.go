package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache handles caching operations using Redis
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache client
func NewRedisCache(addr, password string, db int) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

// Get retrieves a value from cache
func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Key doesn't exist
	}
	return val, err
}

// GetJSON retrieves and unmarshals JSON data from cache
func (c *RedisCache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := c.Get(ctx, key)
	if err != nil {
		return err
	}
	if val == "" {
		return nil
	}
	return json.Unmarshal([]byte(val), dest)
}

// Set stores a value in cache with expiration
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// SetJSON marshals and stores JSON data in cache
func (c *RedisCache) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.Set(ctx, key, data, expiration)
}

// Delete removes a key from cache
func (c *RedisCache) Delete(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, key).Result()
	return n > 0, err
}

// Expire sets a timeout on a key
func (c *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}

// IncrBy increments a counter
func (c *RedisCache) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, key, value).Result()
}

// GetTTL returns remaining time-to-live for a key
func (c *RedisCache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

// FlushDB clears all keys (use with caution!)
func (c *RedisCache) FlushDB(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// CacheKey generators for different data types
type CacheKey struct{}

// User cache keys
func (CacheKey) User(userID string) string {
	return fmt.Sprintf("user:%s", userID)
}

// Course cache keys
func (CacheKey) Course(courseID string) string {
	return fmt.Sprintf("course:%s", courseID)
}

func (CacheKey) CourseList(page, pageSize int) string {
	return fmt.Sprintf("courses:list:%d:%d", page, pageSize)
}

// Instructor cache keys
func (CacheKey) Instructor(instructorID string) string {
	return fmt.Sprintf("instructor:%s", instructorID)
}

func (CacheKey) InstructorList(filters string) string {
	return fmt.Sprintf("instructors:list:%s", filters)
}

// Session cache keys
func (CacheKey) Session(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}

// Recording cache keys
func (CacheKey) Recording(recordingID string) string {
	return fmt.Sprintf("recording:%s", recordingID)
}

func (CacheKey) RecordingList(filters string) string {
	return fmt.Sprintf("recordings:list:%s", filters)
}

// Stats cache keys
func (CacheKey) Stats(entity, id string) string {
	return fmt.Sprintf("stats:%s:%s", entity, id)
}

// Default cache expiration times
const (
	ExpirationShort   = 5 * time.Minute    // For frequently changing data
	ExpirationMedium  = 30 * time.Minute   // For moderately stable data
	ExpirationLong    = 2 * time.Hour      // For stable data
	ExpirationDay     = 24 * time.Hour     // For very stable data
	ExpirationSession = 7 * 24 * time.Hour // For user sessions
)
