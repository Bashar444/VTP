# Horizontal Scaling Guide for VTP Platform

## Overview

This guide covers horizontal scaling setup for the VTP platform, enabling high availability, load distribution, and improved performance through multiple backend instances.

## Architecture

```
                                   Internet
                                      |
                                   Nginx (Load Balancer)
                                      |
                    +----------------+----------------+
                    |                |                |
                Backend-1        Backend-2        Backend-3
                (Go App)         (Go App)         (Go App)
                    |                |                |
                    +----------------+----------------+
                              |            |
                          PostgreSQL    Redis
                         (Shared DB)  (Shared Cache/Sessions)
```

## Components

### 1. Load Balancer (Nginx)

**Purpose**: Distributes incoming traffic across multiple backend instances.

**Features**:
- Load balancing algorithms (least_conn, ip_hash, round_robin)
- Health checks with automatic failover
- SSL/TLS termination
- WebSocket support for real-time signaling
- Rate limiting per endpoint
- Static content caching

**Configuration**: `deployment/nginx.conf`

### 2. Backend Instances (Go Application)

**Scaling**: 3 replicas by default, can scale to N instances.

**Key Considerations**:
- Stateless design (sessions in Redis)
- Each instance has unique INSTANCE_ID
- Shared database (PostgreSQL)
- Shared cache (Redis)
- Independent health checks

### 3. Shared Data Layer

**PostgreSQL**: Single source of truth for persistent data.
- Connection pooling (max 25 connections per instance)
- Read replicas (optional for read-heavy workloads)

**Redis**: Shared cache and session store.
- Session management (7-day expiry)
- Cache layer (5min to 7-day TTL)
- Rate limiting counters
- Real-time data

## Load Balancing Strategies

### 1. Least Connections (Default)

Routes traffic to the backend with the fewest active connections.

```nginx
upstream vtp_backend {
    least_conn;
    server backend1:8080;
    server backend2:8080;
    server backend3:8080;
}
```

**Best for**: Mixed workload with varying request durations.

### 2. IP Hash (Sticky Sessions)

Routes requests from the same client IP to the same backend.

```nginx
upstream vtp_backend {
    ip_hash;
    server backend1:8080;
    server backend2:8080;
    server backend3:8080;
}
```

**Best for**: WebSocket connections, stateful operations.

**Note**: Required for `/socket.io/` endpoints.

### 3. Round Robin

Distributes requests evenly in rotation.

```nginx
upstream vtp_backend {
    server backend1:8080;
    server backend2:8080;
    server backend3:8080;
}
```

**Best for**: Uniform workloads with similar request durations.

## Session Management

### Redis-Based Sessions

All backend instances share session data via Redis:

```go
// Session stored in Redis with 7-day expiry
sessionStore := cache.NewSessionStore(redisClient)
session, _ := sessionStore.GetSession(ctx, sessionID)
```

**Benefits**:
- No sticky sessions required for API endpoints
- Seamless failover between instances
- Centralized session management
- Automatic expiry

### WebSocket Handling

WebSocket connections require sticky sessions (ip_hash):

```nginx
upstream vtp_websocket {
    ip_hash;  # Same client always routes to same backend
    server backend1:8080;
    server backend2:8080;
    server backend3:8080;
}
```

## Health Checks

### Backend Health Endpoint

Each backend instance exposes `/health`:

```go
// Add to cmd/main.go
mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    // Check database connection
    if err := db.Ping(); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "unhealthy",
            "reason": "database unreachable",
        })
        return
    }
    
    // Check Redis connection
    if err := redisClient.Ping(r.Context()).Err(); err != nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "unhealthy",
            "reason": "redis unreachable",
        })
        return
    }
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "healthy",
        "instance": os.Getenv("INSTANCE_ID"),
        "timestamp": time.Now().Unix(),
    })
})
```

### Nginx Health Check Configuration

```nginx
upstream vtp_backend {
    server backend1:8080 max_fails=3 fail_timeout=30s;
    server backend2:8080 max_fails=3 fail_timeout=30s;
    server backend3:8080 max_fails=3 fail_timeout=30s;
}
```

**Parameters**:
- `max_fails=3`: Mark backend unhealthy after 3 consecutive failures
- `fail_timeout=30s`: Retry unhealthy backend after 30 seconds

### Docker Health Checks

```yaml
healthcheck:
  test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
  interval: 10s
  timeout: 5s
  retries: 3
  start_period: 30s
```

## Deployment

### Prerequisites

1. **Environment Variables** (create `.env` file):

```env
# Security
JWT_SECRET=your-super-secret-jwt-key-change-this

# Database
DATABASE_URL=postgresql://postgres:postgres@postgres:5432/vtp?sslmode=disable

# Redis
REDIS_URL=redis://redis:6379

# Public IP for mediasoup
PUBLIC_IP=your.public.ip.address

# Monitoring
GRAFANA_PASSWORD=secure-grafana-password
```

2. **SSL Certificates** (for production):

```bash
mkdir -p deployment/ssl
# Place cert.pem and key.pem in deployment/ssl/
```

### Start Scaled Deployment

```bash
cd deployment
docker-compose -f docker-compose.scaled.yml up -d
```

### Scale Backend Instances

Add more replicas by editing `docker-compose.scaled.yml` or using Docker Swarm/Kubernetes.

**Manual scaling**:

```yaml
# Add backend4, backend5, etc.
backend4:
  build:
    context: ..
    dockerfile: Dockerfile
  environment:
    - INSTANCE_ID=backend-4
    # ... other config
```

**Docker Swarm scaling**:

```bash
docker service scale vtp-backend=5
```

### Verify Deployment

1. **Check container status**:

```bash
docker-compose -f docker-compose.scaled.yml ps
```

2. **Test load balancer**:

```bash
curl http://localhost/health
```

3. **Check logs**:

```bash
docker-compose -f docker-compose.scaled.yml logs -f backend1
```

4. **Monitor traffic distribution**:

```bash
docker-compose -f docker-compose.scaled.yml logs nginx | grep "upstream"
```

## Scaling Strategies

### Vertical Scaling (Per Instance)

Increase resources for each backend instance:

```yaml
deploy:
  resources:
    limits:
      cpus: '2.0'      # Increase from 1.0
      memory: 1G       # Increase from 512M
```

**When to use**: High CPU/memory usage per instance.

### Horizontal Scaling (More Instances)

Add more backend replicas:

```bash
# Add backend4, backend5 in docker-compose.scaled.yml
# Update nginx.conf upstream block
```

**When to use**: High request volume, need redundancy.

### Auto-Scaling (Kubernetes)

Deploy with HPA (Horizontal Pod Autoscaler):

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: vtp-backend
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: vtp-backend
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## Performance Optimization

### 1. Connection Pooling

**PostgreSQL** (in each backend):

```go
db.SetMaxOpenConns(25)  // Max connections per instance
db.SetMaxIdleConns(5)   // Idle connections
db.SetConnMaxLifetime(5 * time.Minute)
```

**Redis** (in each backend):

```go
redisClient := redis.NewClient(&redis.Options{
    PoolSize:     10,  // Connection pool size
    MinIdleConns: 2,   // Minimum idle connections
})
```

### 2. Nginx Optimization

```nginx
# Increase worker connections
events {
    worker_connections 2048;
}

# Enable keepalive to backends
upstream vtp_backend {
    keepalive 32;  # Keep 32 connections alive
}

# Proxy keepalive
proxy_http_version 1.1;
proxy_set_header Connection "";
```

### 3. Caching Strategy

**Static content** (Nginx):

```nginx
location /static/ {
    expires 1y;
    add_header Cache-Control "public, immutable";
}
```

**API responses** (Redis):

```go
// Cache frequently accessed data
cachedCourseService := cache.NewCachedCourseService(courseService, redisClient)
course, err := cachedCourseService.GetCourse(ctx, courseID, func() (*models.Course, error) {
    return courseService.GetCourse(ctx, courseID)
})
```

### 4. Rate Limiting

**Nginx layer**:

```nginx
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=100r/m;

location /api/ {
    limit_req zone=api_limit burst=20 nodelay;
}
```

**Application layer** (Redis):

```go
rateLimiter := cache.NewRateLimiter(redisClient, "api", 100, time.Minute)
if !rateLimiter.Allow(ctx, userID) {
    return errors.New("rate limit exceeded")
}
```

## Monitoring

### Metrics to Track

1. **Request Distribution**:
   - Requests per backend instance
   - Response times per instance
   - Error rates per instance

2. **Resource Usage**:
   - CPU usage per instance
   - Memory usage per instance
   - Network I/O

3. **Backend Health**:
   - Health check pass/fail rates
   - Uptime per instance
   - Connection pool usage

4. **Load Balancer**:
   - Total requests
   - Upstream connection errors
   - Rate limiting triggers

### Prometheus Configuration

Create `deployment/prometheus.yml`:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'vtp-backend'
    static_configs:
      - targets:
        - backend1:8080
        - backend2:8080
        - backend3:8080
    metrics_path: '/metrics'

  - job_name: 'nginx'
    static_configs:
      - targets:
        - nginx:9113  # nginx-prometheus-exporter
```

### Grafana Dashboards

Access Grafana at `http://localhost:3001`:

1. **Backend Performance**:
   - Request rate per instance
   - Response time percentiles (p50, p95, p99)
   - Error rate

2. **Database Connections**:
   - Active connections per backend
   - Connection pool saturation

3. **Redis Performance**:
   - Cache hit/miss rates
   - Memory usage
   - Operations per second

## Troubleshooting

### Uneven Load Distribution

**Symptom**: One backend receives more traffic than others.

**Solutions**:
1. Check load balancing algorithm (use `least_conn`)
2. Verify all backends are healthy
3. Check for long-polling connections (use ip_hash for WebSocket)

### Session Loss

**Symptom**: Users logged out unexpectedly.

**Solutions**:
1. Verify Redis is running and accessible
2. Check Redis memory (ensure not evicting sessions)
3. Verify session TTL is sufficient (7 days default)

### WebSocket Connection Issues

**Symptom**: Real-time features not working.

**Solutions**:
1. Ensure `ip_hash` for `/socket.io/` upstream
2. Check WebSocket upgrade headers
3. Verify long timeouts (7 days) for WebSocket proxying

### Database Connection Pool Exhausted

**Symptom**: "too many connections" errors.

**Solutions**:
1. Reduce `MaxOpenConns` per backend instance
2. Increase PostgreSQL `max_connections`
3. Implement connection pooling middleware (PgBouncer)

### High Memory Usage

**Symptom**: Backends consuming excessive memory.

**Solutions**:
1. Check for memory leaks (use pprof: `/debug/pprof/heap`)
2. Reduce cache size
3. Implement memory limits in Docker

## Production Checklist

- [ ] SSL/TLS certificates installed in `deployment/ssl/`
- [ ] Environment variables configured in `.env`
- [ ] Database migrations applied
- [ ] Redis persistence enabled (`appendonly yes`)
- [ ] Health checks verified on all backends
- [ ] Load balancer tested with `curl` requests
- [ ] WebSocket connections tested
- [ ] Rate limiting configured and tested
- [ ] Monitoring dashboards configured (Prometheus/Grafana)
- [ ] Logging centralized (ELK stack or CloudWatch)
- [ ] Backup strategy implemented (PostgreSQL, Redis)
- [ ] Firewall rules configured (ports 80, 443, internal only)
- [ ] Resource limits set (CPU, memory per container)
- [ ] Auto-restart policies configured (`restart: unless-stopped`)
- [ ] Documentation updated with production URLs

## Kubernetes Deployment (Advanced)

For Kubernetes deployment, see `deployment/k8s/README.md` with:

- Deployment manifests for backend (with replicas)
- Service definitions (LoadBalancer, ClusterIP)
- Ingress controller (Nginx or Traefik)
- HorizontalPodAutoscaler configuration
- ConfigMaps and Secrets management
- StatefulSet for PostgreSQL/Redis (or use managed services)

## Cost Optimization

1. **Right-size instances**: Start with 3 backends, scale based on metrics
2. **Use managed services**: AWS RDS (PostgreSQL), ElastiCache (Redis)
3. **Auto-scaling**: Scale down during low-traffic hours
4. **CDN for static content**: Offload static files to CloudFront/Cloudflare
5. **Connection pooling**: Reduce database connection overhead (PgBouncer)

## Summary

- ✅ **Load Balancer**: Nginx with least_conn algorithm
- ✅ **Backend Scaling**: 3 replicas by default, scale to N
- ✅ **Session Management**: Redis-based, shared across instances
- ✅ **Health Checks**: Automatic failover with 30s retry
- ✅ **WebSocket Support**: IP hash for sticky sessions
- ✅ **Monitoring**: Prometheus + Grafana dashboards
- ✅ **Rate Limiting**: Nginx and application-layer protection
- ✅ **Deployment**: Docker Compose for easy scaling

Next steps: Deploy to production, monitor metrics, tune based on load.
