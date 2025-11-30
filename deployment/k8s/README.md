# Kubernetes Deployment Configuration for VTP Platform

This directory contains Kubernetes manifests for deploying the VTP platform with horizontal scaling capabilities.

## Architecture

```
                         External Traffic
                               |
                         LoadBalancer
                               |
                         Ingress Controller
                               |
              +----------------+----------------+
              |                |                |
         Backend-1        Backend-2        Backend-3
         (Deployment: 3+ replicas)
              |                |                |
              +----------------+----------------+
                      |                |
                 PostgreSQL        Redis
              (StatefulSet)    (StatefulSet)
```

## Prerequisites

1. **Kubernetes Cluster**: Version 1.20+
2. **kubectl**: Configured with cluster access
3. **Docker Registry**: For storing container images
4. **Storage**: PersistentVolumes for PostgreSQL and Redis

## Quick Start

### 1. Build and Push Docker Image

```bash
# Build the backend image
docker build -t your-registry.com/vtp-backend:latest .

# Push to registry
docker push your-registry.com/vtp-backend:latest
```

### 2. Configure Secrets

```bash
# Create namespace
kubectl create namespace vtp

# Create secrets
kubectl create secret generic vtp-secrets \
  --from-literal=jwt-secret=your-jwt-secret-here \
  --from-literal=postgres-password=your-postgres-password \
  --namespace=vtp

# Create TLS certificate secret (for ingress)
kubectl create secret tls vtp-tls \
  --cert=path/to/cert.pem \
  --key=path/to/key.pem \
  --namespace=vtp
```

### 3. Deploy PostgreSQL and Redis

```bash
kubectl apply -f postgres-statefulset.yaml
kubectl apply -f redis-statefulset.yaml
```

Wait for databases to be ready:

```bash
kubectl wait --for=condition=ready pod -l app=postgres --timeout=300s -n vtp
kubectl wait --for=condition=ready pod -l app=redis --timeout=300s -n vtp
```

### 4. Deploy Backend Application

```bash
kubectl apply -f backend-deployment.yaml
kubectl apply -f backend-service.yaml
```

### 5. Deploy Ingress Controller

```bash
# Install Nginx Ingress Controller (if not already installed)
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.1/deploy/static/provider/cloud/deploy.yaml

# Deploy VTP ingress
kubectl apply -f ingress.yaml
```

### 6. Deploy Horizontal Pod Autoscaler

```bash
kubectl apply -f hpa.yaml
```

## Verify Deployment

```bash
# Check all resources
kubectl get all -n vtp

# Check pod status
kubectl get pods -n vtp

# Check services
kubectl get svc -n vtp

# Check ingress
kubectl get ingress -n vtp

# Check HPA status
kubectl get hpa -n vtp

# View logs from a backend pod
kubectl logs -f deployment/vtp-backend -n vtp

# Check health endpoint
kubectl port-forward svc/vtp-backend 8080:8080 -n vtp
curl http://localhost:8080/health
```

## Scaling

### Manual Scaling

```bash
# Scale to 5 replicas
kubectl scale deployment vtp-backend --replicas=5 -n vtp

# Check scaling progress
kubectl rollout status deployment/vtp-backend -n vtp
```

### Auto-Scaling (HPA)

The HPA automatically scales between 3-10 replicas based on:
- CPU utilization: target 70%
- Memory utilization: target 80%

View HPA status:

```bash
kubectl get hpa vtp-backend -n vtp --watch
```

## Monitoring

### View Resource Usage

```bash
# Pod resource usage
kubectl top pods -n vtp

# Node resource usage
kubectl top nodes
```

### View Logs

```bash
# All backend pods
kubectl logs -l app=vtp-backend -n vtp --tail=100 -f

# Specific pod
kubectl logs vtp-backend-<pod-id> -n vtp
```

### Describe Resources

```bash
# Deployment details
kubectl describe deployment vtp-backend -n vtp

# Service details
kubectl describe service vtp-backend -n vtp

# Ingress details
kubectl describe ingress vtp-ingress -n vtp
```

## Configuration

### Environment Variables

Edit `backend-deployment.yaml` to configure:

- `DATABASE_URL`: PostgreSQL connection string
- `REDIS_URL`: Redis connection string
- `JWT_SECRET`: JWT signing key (from secret)
- `CORS_ORIGINS`: Allowed CORS origins
- `LOG_LEVEL`: Logging level (debug, info, warn, error)

### Resource Limits

Adjust resources in `backend-deployment.yaml`:

```yaml
resources:
  requests:
    cpu: 500m      # 0.5 CPU cores
    memory: 256Mi  # 256 MB RAM
  limits:
    cpu: 1000m     # 1 CPU core
    memory: 512Mi  # 512 MB RAM
```

### Ingress Configuration

Edit `ingress.yaml` to configure:

- `host`: Your domain name
- `tls`: TLS certificate configuration
- Annotations for rate limiting, authentication, etc.

## Troubleshooting

### Pods Not Starting

```bash
# Check pod events
kubectl describe pod <pod-name> -n vtp

# Check logs
kubectl logs <pod-name> -n vtp

# Common issues:
# - Image pull errors: Check registry credentials
# - CrashLoopBackOff: Check logs for application errors
# - Pending: Check node resources and PersistentVolume availability
```

### Database Connection Issues

```bash
# Test database connectivity from backend pod
kubectl exec -it deployment/vtp-backend -n vtp -- /bin/sh
wget -qO- http://postgres:5432

# Check PostgreSQL logs
kubectl logs statefulset/postgres -n vtp
```

### Ingress Not Working

```bash
# Check ingress controller logs
kubectl logs -n ingress-nginx deployment/ingress-nginx-controller

# Check ingress status
kubectl describe ingress vtp-ingress -n vtp

# Common issues:
# - DNS not pointing to LoadBalancer IP
# - TLS certificate issues
# - Backend service not reachable
```

### HPA Not Scaling

```bash
# Check HPA status
kubectl describe hpa vtp-backend -n vtp

# Check metrics server
kubectl top nodes
kubectl top pods -n vtp

# Install metrics-server if not available:
# kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

## Updates and Rollbacks

### Rolling Update

```bash
# Update image
kubectl set image deployment/vtp-backend vtp-backend=your-registry.com/vtp-backend:v2 -n vtp

# Watch rollout
kubectl rollout status deployment/vtp-backend -n vtp
```

### Rollback

```bash
# View rollout history
kubectl rollout history deployment/vtp-backend -n vtp

# Rollback to previous version
kubectl rollout undo deployment/vtp-backend -n vtp

# Rollback to specific revision
kubectl rollout undo deployment/vtp-backend --to-revision=2 -n vtp
```

## Cleanup

```bash
# Delete all VTP resources
kubectl delete namespace vtp

# Or delete individually
kubectl delete -f .
```

## Production Considerations

1. **Database**: Use managed PostgreSQL (AWS RDS, Google Cloud SQL, Azure Database)
2. **Redis**: Use managed Redis (AWS ElastiCache, Google Memorystore, Azure Cache)
3. **Storage**: Use persistent volumes with appropriate storage class
4. **Monitoring**: Deploy Prometheus + Grafana for metrics
5. **Logging**: Use ELK stack or cloud logging (CloudWatch, Stackdriver)
6. **Secrets**: Use external secret management (Vault, AWS Secrets Manager)
7. **Backups**: Automate database backups with CronJobs
8. **CDN**: Use CloudFront/Cloudflare for static content
9. **Security**: Implement NetworkPolicies and PodSecurityPolicies

## Files

- `backend-deployment.yaml`: Backend application deployment
- `backend-service.yaml`: Backend service (ClusterIP)
- `postgres-statefulset.yaml`: PostgreSQL database
- `redis-statefulset.yaml`: Redis cache
- `ingress.yaml`: Ingress controller configuration
- `hpa.yaml`: Horizontal Pod Autoscaler
- `configmap.yaml`: Application configuration
- `secrets.yaml.example`: Example secrets file

## Support

For issues, see the main project documentation or contact the development team.
