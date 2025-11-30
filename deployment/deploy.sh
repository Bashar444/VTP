#!/bin/bash

# VTP Platform - Horizontal Scaling Deployment Script
# This script automates the deployment of VTP platform with horizontal scaling

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="vtp"
DOCKER_REGISTRY="${DOCKER_REGISTRY:-your-registry.com}"
IMAGE_TAG="${IMAGE_TAG:-latest}"

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi
    log_info "✓ Docker found"
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_warn "Docker Compose not found, some features may not work"
    else
        log_info "✓ Docker Compose found"
    fi
    
    # Check kubectl (for Kubernetes deployments)
    if command -v kubectl &> /dev/null; then
        log_info "✓ kubectl found"
    else
        log_warn "kubectl not found, Kubernetes deployment unavailable"
    fi
}

build_docker_image() {
    log_info "Building Docker image..."
    
    cd ..
    docker build -t ${DOCKER_REGISTRY}/vtp-backend:${IMAGE_TAG} .
    
    if [ $? -eq 0 ]; then
        log_info "✓ Docker image built successfully"
    else
        log_error "Failed to build Docker image"
        exit 1
    fi
}

push_docker_image() {
    log_info "Pushing Docker image to registry..."
    
    docker push ${DOCKER_REGISTRY}/vtp-backend:${IMAGE_TAG}
    
    if [ $? -eq 0 ]; then
        log_info "✓ Docker image pushed successfully"
    else
        log_error "Failed to push Docker image"
        exit 1
    fi
}

deploy_docker_compose() {
    log_info "Deploying with Docker Compose..."
    
    # Check if .env exists
    if [ ! -f .env ]; then
        log_error ".env file not found. Please create it from .env.example"
        exit 1
    fi
    
    # Start services
    docker-compose -f docker-compose.scaled.yml up -d
    
    if [ $? -eq 0 ]; then
        log_info "✓ Docker Compose deployment successful"
        log_info "Services are starting..."
        sleep 5
        docker-compose -f docker-compose.scaled.yml ps
    else
        log_error "Failed to deploy with Docker Compose"
        exit 1
    fi
}

deploy_kubernetes() {
    log_info "Deploying to Kubernetes..."
    
    # Create namespace
    kubectl create namespace ${NAMESPACE} --dry-run=client -o yaml | kubectl apply -f -
    log_info "✓ Namespace created/verified"
    
    # Apply ConfigMaps
    kubectl apply -f k8s/configmap.yaml
    log_info "✓ ConfigMaps applied"
    
    # Check for secrets
    if ! kubectl get secret vtp-secrets -n ${NAMESPACE} &> /dev/null; then
        log_warn "Secret 'vtp-secrets' not found. Please create it manually:"
        log_warn "kubectl create secret generic vtp-secrets \\"
        log_warn "  --from-literal=jwt-secret=your-jwt-secret \\"
        log_warn "  --from-literal=postgres-password=your-postgres-password \\"
        log_warn "  --namespace=${NAMESPACE}"
        exit 1
    fi
    log_info "✓ Secrets verified"
    
    # Deploy PostgreSQL
    kubectl apply -f k8s/postgres-statefulset.yaml
    log_info "✓ PostgreSQL StatefulSet applied"
    
    # Deploy Redis
    kubectl apply -f k8s/redis-statefulset.yaml
    log_info "✓ Redis StatefulSet applied"
    
    # Wait for databases
    log_info "Waiting for databases to be ready..."
    kubectl wait --for=condition=ready pod -l app=postgres --timeout=300s -n ${NAMESPACE} || log_warn "PostgreSQL timeout"
    kubectl wait --for=condition=ready pod -l app=redis --timeout=300s -n ${NAMESPACE} || log_warn "Redis timeout"
    
    # Deploy backend
    kubectl apply -f k8s/backend-deployment.yaml
    kubectl apply -f k8s/backend-service.yaml
    log_info "✓ Backend Deployment and Service applied"
    
    # Deploy HPA
    kubectl apply -f k8s/hpa.yaml
    log_info "✓ HorizontalPodAutoscaler applied"
    
    # Deploy Ingress
    kubectl apply -f k8s/ingress.yaml
    log_info "✓ Ingress applied"
    
    log_info "Deployment complete! Waiting for pods to be ready..."
    kubectl wait --for=condition=ready pod -l app=vtp-backend --timeout=300s -n ${NAMESPACE}
    
    log_info "✓ All pods are ready"
}

verify_deployment() {
    log_info "Verifying deployment..."
    
    if [ "$DEPLOYMENT_TYPE" == "docker-compose" ]; then
        # Test health endpoint
        sleep 5
        HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost/health || echo "000")
        
        if [ "$HTTP_CODE" == "200" ]; then
            log_info "✓ Health check passed"
        else
            log_error "Health check failed (HTTP $HTTP_CODE)"
        fi
        
        # Show running containers
        docker-compose -f docker-compose.scaled.yml ps
        
    elif [ "$DEPLOYMENT_TYPE" == "kubernetes" ]; then
        # Show pod status
        kubectl get pods -n ${NAMESPACE}
        
        # Show service status
        kubectl get svc -n ${NAMESPACE}
        
        # Show HPA status
        kubectl get hpa -n ${NAMESPACE}
        
        # Test health endpoint
        log_info "Testing health endpoint via port-forward..."
        kubectl port-forward -n ${NAMESPACE} svc/vtp-backend 8080:8080 &
        PF_PID=$!
        sleep 3
        
        HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health || echo "000")
        kill $PF_PID 2>/dev/null || true
        
        if [ "$HTTP_CODE" == "200" ]; then
            log_info "✓ Health check passed"
        else
            log_warn "Health check failed (HTTP $HTTP_CODE)"
        fi
    fi
}

show_urls() {
    log_info "============================================"
    log_info "VTP Platform Deployment Complete!"
    log_info "============================================"
    
    if [ "$DEPLOYMENT_TYPE" == "docker-compose" ]; then
        log_info ""
        log_info "Access URLs:"
        log_info "  - API: http://localhost"
        log_info "  - Health: http://localhost/health"
        log_info "  - Grafana: http://localhost:3001"
        log_info "  - Prometheus: http://localhost:9090"
        log_info ""
        log_info "View logs:"
        log_info "  docker-compose -f docker-compose.scaled.yml logs -f"
        log_info ""
        log_info "Scale backends:"
        log_info "  Edit docker-compose.scaled.yml to add backend4, backend5, etc."
        
    elif [ "$DEPLOYMENT_TYPE" == "kubernetes" ]; then
        INGRESS_IP=$(kubectl get ingress vtp-ingress -n ${NAMESPACE} -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "pending")
        
        log_info ""
        log_info "Cluster resources:"
        log_info "  - Namespace: ${NAMESPACE}"
        log_info "  - Ingress IP: ${INGRESS_IP}"
        log_info ""
        log_info "View resources:"
        log_info "  kubectl get all -n ${NAMESPACE}"
        log_info ""
        log_info "View logs:"
        log_info "  kubectl logs -f deployment/vtp-backend -n ${NAMESPACE}"
        log_info ""
        log_info "Scale manually:"
        log_info "  kubectl scale deployment vtp-backend --replicas=5 -n ${NAMESPACE}"
        log_info ""
        log_info "View HPA:"
        log_info "  kubectl get hpa -n ${NAMESPACE} --watch"
    fi
    
    log_info "============================================"
}

# Main script
main() {
    echo "VTP Platform - Horizontal Scaling Deployment"
    echo "============================================"
    echo ""
    
    # Parse arguments
    DEPLOYMENT_TYPE="${1:-docker-compose}"
    BUILD_IMAGE="${2:-no}"
    
    if [ "$DEPLOYMENT_TYPE" != "docker-compose" ] && [ "$DEPLOYMENT_TYPE" != "kubernetes" ]; then
        log_error "Usage: $0 [docker-compose|kubernetes] [build]"
        log_error "  docker-compose: Deploy with Docker Compose (default)"
        log_error "  kubernetes: Deploy to Kubernetes cluster"
        log_error "  build: Optional flag to build and push Docker image"
        exit 1
    fi
    
    check_prerequisites
    
    # Build and push image if requested
    if [ "$BUILD_IMAGE" == "build" ]; then
        build_docker_image
        
        if [ "$DEPLOYMENT_TYPE" == "kubernetes" ]; then
            push_docker_image
        fi
    fi
    
    # Deploy based on type
    if [ "$DEPLOYMENT_TYPE" == "docker-compose" ]; then
        deploy_docker_compose
    elif [ "$DEPLOYMENT_TYPE" == "kubernetes" ]; then
        deploy_kubernetes
    fi
    
    # Verify deployment
    verify_deployment
    
    # Show access information
    show_urls
}

# Run main function
main "$@"
