# Caching Strategy Guide

## Overview
The VTP platform uses Redis for caching and session management to improve performance and reduce database load.

## Architecture

### Cache Layers
1. **Application Cache** - In-memory Go maps for hot data
2. **Redis Cache** - Distributed cache for shared data
3. **Database** - Source of truth

### Cache Pattern: Cache-Aside
```
1. Check Redis cache
2. If miss, query database
3. Store result in Redis
4. Return data
```

## Redis Setup

### Local Development
```bash
# Using Docker
docker run -d --name redis -p 6379:6379 redis:7-alpine

# Using Windows (with Chocolatey)
choco install redis-64

# Using Ubuntu/Debian
sudo apt-get install redis-server
```

### Environment Variables
```bash
export REDIS_ADDR="localhost:6379"
export REDIS_PASSWORD=""
export REDIS_DB=0
```

## Cache Keys

### Naming Convention
```
<entity>:<id>
<entity>:list:<filters>
<prefix>:<entity>:<id>
```

### Examples
```go
user:123e4567-e89b-12d3-a456-426614174000
course:abc123
instructors:list:verified=true&page=1
session:sess_xyz789
stats:course:abc123
ratelimit:api:192.168.1.1
```

## Expiration Times

| Data Type | TTL | Reason |
|-----------|-----|--------|
| User Profile | 30min | Moderately stable |
| Course List | 5min | Frequently updated |
| Instructor Details | 2h | Stable content |
| Session Data | 7 days | Login persistence |
| Stats/Analytics | 24h | Computed daily |
| Rate Limit | 1min-1h | Per-endpoint basis |

## Usage Examples

### Basic Caching
```go
import "github.com/Bashar444/VTP/pkg/cache"

// Initialize Redis
redisCache, err := cache.NewRedisCache("localhost:6379", "", 0)
if err != nil {
    log.Fatal(err)
}
defer redisCache.Close()

// Store data
ctx := context.Background()
err = redisCache.SetJSON(ctx, "user:123", userData, cache.ExpirationMedium)

// Retrieve data
var user User
err = redisCache.GetJSON(ctx, "user:123", &user)
```

### Cache-Aside Pattern
```go
func GetCourse(ctx context.Context, courseID string) (*Course, error) {
    cacheKey := cache.CacheKey{}.Course(courseID)
    
    // Try cache first
    var course Course
    if err := redisCache.GetJSON(ctx, cacheKey, &course); err == nil {
        return &course, nil
    }
    
    // Cache miss - fetch from DB
    course, err := db.QueryCourse(ctx, courseID)
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    redisCache.SetJSON(ctx, cacheKey, course, cache.ExpirationMedium)
    
    return course, nil
}
```

### Cache Invalidation
```go
// On update
func UpdateCourse(ctx context.Context, courseID string, data *CourseData) error {
    // Update database
    if err := db.UpdateCourse(ctx, courseID, data); err != nil {
        return err
    }
    
    // Invalidate cache
    cacheKey := cache.CacheKey{}.Course(courseID)
    redisCache.Delete(ctx, cacheKey)
    
    return nil
}
```

### Session Management
```go
sessionStore := cache.NewSessionStore(redisCache)

// Create session
sessionData := &cache.SessionData{
    UserID: "user123",
    Email:  "user@example.com",
    Role:   "student",
    Data:   map[string]interface{}{"locale": "ar_SY"},
}
err = sessionStore.CreateSession(ctx, sessionID, sessionData)

// Get session
session, err := sessionStore.GetSession(ctx, sessionID)

// Extend session
err = sessionStore.ExtendSession(ctx, sessionID)

// Delete session (logout)
err = sessionStore.DeleteSession(ctx, sessionID)
```

### Rate Limiting
```go
rateLimiter := cache.NewRateLimiter(redisCache)

// Check rate limit (100 requests per minute)
allowed, err := rateLimiter.Allow(ctx, "api:"+userID, 100, time.Minute)
if !allowed {
    return ErrRateLimitExceeded
}

// Get remaining requests
remaining, err := rateLimiter.GetRemainingRequests(ctx, "api:"+userID, 100)
```

## Caching Best Practices

### 1. Cache What's Expensive
- Database queries
- API calls
- Complex computations
- Frequently accessed data

### 2. Don't Cache Everything
Avoid caching:
- Rapidly changing data
- Large objects (> 1MB)
- User-specific sensitive data
- Real-time data

### 3. Set Appropriate TTLs
```go
// Short TTL for frequently updated data
cache.SetJSON(ctx, key, data, 5*time.Minute)

// Long TTL for stable data
cache.SetJSON(ctx, key, data, 24*time.Hour)

// No expiration for static data
cache.Set(ctx, key, data, 0)
```

### 4. Handle Cache Failures Gracefully
```go
func GetData(ctx context.Context, key string) (*Data, error) {
    // Try cache
    var data Data
    if err := cache.GetJSON(ctx, key, &data); err == nil {
        return &data, nil
    }
    
    // Cache miss or error - fetch from DB
    data, err := db.Query(ctx, key)
    if err != nil {
        return nil, err
    }
    
    // Try to cache (ignore errors)
    _ = cache.SetJSON(ctx, key, data, cache.ExpirationMedium)
    
    return data, nil
}
```

### 5. Invalidate Strategically
```go
// Invalidate on write
func UpdateUser(ctx context.Context, userID string, data *UserData) error {
    if err := db.Update(ctx, userID, data); err != nil {
        return err
    }
    
    // Invalidate related caches
    cache.Delete(ctx, 
        cache.CacheKey{}.User(userID),
        cache.CacheKey{}.UserStats(userID),
    )
    
    return nil
}
```

## Cache Warming

### Preload Hot Data
```go
func WarmCache(ctx context.Context) error {
    // Load popular courses
    courses, err := db.GetPopularCourses(ctx, 50)
    if err != nil {
        return err
    }
    
    for _, course := range courses {
        key := cache.CacheKey{}.Course(course.ID)
        cache.SetJSON(ctx, key, course, cache.ExpirationLong)
    }
    
    return nil
}
```

### Scheduled Warming
```go
// Run every hour
ticker := time.NewTicker(1 * time.Hour)
go func() {
    for range ticker.C {
        if err := WarmCache(context.Background()); err != nil {
            log.Printf("Cache warming failed: %v", err)
        }
    }
}()
```

## Monitoring

### Key Metrics
- **Hit Rate**: Cache hits / Total requests
- **Miss Rate**: Cache misses / Total requests
- **Eviction Rate**: Keys evicted / Total keys
- **Memory Usage**: Current / Max memory
- **Connection Count**: Active connections

### Redis Commands
```bash
# Monitor cache performance
redis-cli INFO stats

# Check memory usage
redis-cli INFO memory

# View all keys (dev only!)
redis-cli KEYS "*"

# Check specific key TTL
redis-cli TTL user:123

# Flush cache (DANGEROUS!)
redis-cli FLUSHDB
```

### Application Metrics
```go
import "github.com/prometheus/client_golang/prometheus"

var (
    cacheHits = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cache_hits_total",
            Help: "Total number of cache hits",
        },
        []string{"type"},
    )
    
    cacheMisses = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cache_misses_total",
            Help: "Total number of cache misses",
        },
        []string{"type"},
    )
)
```

## Performance Optimization

### 1. Pipeline Multiple Operations
```go
pipe := redisClient.Pipeline()
pipe.Set(ctx, "key1", "value1", 0)
pipe.Set(ctx, "key2", "value2", 0)
pipe.Set(ctx, "key3", "value3", 0)
_, err := pipe.Exec(ctx)
```

### 2. Use Compression for Large Objects
```go
import "compress/gzip"

func SetCompressed(ctx context.Context, key string, data interface{}) error {
    // Marshal to JSON
    jsonData, err := json.Marshal(data)
    if err != nil {
        return err
    }
    
    // Compress
    var buf bytes.Buffer
    writer := gzip.NewWriter(&buf)
    writer.Write(jsonData)
    writer.Close()
    
    // Store compressed
    return cache.Set(ctx, key, buf.Bytes(), cache.ExpirationMedium)
}
```

### 3. Batch Key Operations
```go
// Get multiple keys at once
keys := []string{"course:1", "course:2", "course:3"}
results, err := redisClient.MGet(ctx, keys...).Result()
```

## Troubleshooting

### High Memory Usage
1. Check key count: `redis-cli DBSIZE`
2. Find large keys: `redis-cli --bigkeys`
3. Review TTL settings
4. Implement eviction policy

### Cache Stampede
Problem: Many requests fetch same data when cache expires

Solution: Use mutex or probabilistic early expiration
```go
func GetWithStampedeProtection(ctx context.Context, key string) (*Data, error) {
    // Try cache
    data, err := cache.Get(ctx, key)
    if err == nil {
        return data, nil
    }
    
    // Use mutex to prevent stampede
    mutex := sync.Mutex{}
    mutex.Lock()
    defer mutex.Unlock()
    
    // Check cache again (another goroutine may have loaded it)
    data, err = cache.Get(ctx, key)
    if err == nil {
        return data, nil
    }
    
    // Fetch and cache
    data, err = fetchFromDB(ctx, key)
    if err != nil {
        return nil, err
    }
    
    cache.Set(ctx, key, data, cache.ExpirationMedium)
    return data, nil
}
```

### Connection Pool Exhaustion
```go
// Configure connection pool
redisClient := redis.NewClient(&redis.Options{
    Addr:         "localhost:6379",
    PoolSize:     100,           // Max connections
    MinIdleConns: 10,            // Min idle connections
    PoolTimeout:  4 * time.Second,
    IdleTimeout:  5 * time.Minute,
})
```

## Production Considerations

### 1. Redis Persistence
```conf
# redis.conf
save 900 1      # Save after 900s if 1 key changed
save 300 10     # Save after 300s if 10 keys changed
save 60 10000   # Save after 60s if 10000 keys changed

appendonly yes  # Enable AOF
```

### 2. High Availability
- Use Redis Sentinel for automatic failover
- Or Redis Cluster for horizontal scaling
- Configure replicas for read scaling

### 3. Security
```bash
# Set password
requirepass your_strong_password

# Disable dangerous commands
rename-command FLUSHDB ""
rename-command FLUSHALL ""
rename-command CONFIG ""
```

### 4. Monitoring & Alerts
- Monitor memory usage (< 80%)
- Track hit/miss ratio (> 80% hits)
- Alert on connection failures
- Monitor latency (< 1ms p99)

## Integration with VTP Platform

### Cached Endpoints
- `GET /api/v1/courses` - Course listings
- `GET /api/v1/instructors` - Instructor profiles
- `GET /api/v1/recordings/:id` - Recording metadata
- `GET /api/v1/analytics/stats` - Analytics data

### Session-Based Endpoints
- All authenticated routes use Redis sessions
- JWT tokens cached for validation
- User permissions cached per session

### Rate-Limited Endpoints
- `/api/v1/auth/login` - 5 attempts/minute
- `/api/v1/auth/register` - 3 attempts/hour
- `/api/v1/auth/forgot-password` - 3 attempts/hour

## Resources

- [Redis Documentation](https://redis.io/docs/)
- [go-redis Guide](https://redis.uptrace.dev/)
- [Caching Best Practices](https://redis.io/docs/manual/patterns/)
