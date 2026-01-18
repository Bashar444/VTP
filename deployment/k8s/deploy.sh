#!/bin/bash
# VTP Kubernetes Deployment Script
# Deploys all VTP services to a Kubernetes cluster

set -e

NAMESPACE=${NAMESPACE:-vtp}
REGISTRY=${REGISTRY:-localhost:5000}
VERSION=${VERSION:-latest}

echo "🚀 VTP Kubernetes Deployment"
echo "   Namespace: $NAMESPACE"
echo "   Registry: $REGISTRY"
echo "   Version: $VERSION"
echo ""

# Create namespace if not exists
echo "📦 Creating namespace..."
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# Build and push images
echo "🔨 Building Docker images..."
docker build -t $REGISTRY/vtp-api:$VERSION -f Dockerfile .
docker build -t $REGISTRY/vtp-frontend:$VERSION -f vtp-frontend/Dockerfile ./vtp-frontend
docker build -t $REGISTRY/vtp-mediasoup:$VERSION -f mediasoup-sfu/Dockerfile ./mediasoup-sfu

echo "📤 Pushing images to registry..."
docker push $REGISTRY/vtp-api:$VERSION
docker push $REGISTRY/vtp-frontend:$VERSION
docker push $REGISTRY/vtp-mediasoup:$VERSION

# Apply ConfigMaps and Secrets
echo "⚙️ Applying configurations..."
kubectl apply -f deployment/k8s/configmap.yaml -n $NAMESPACE
kubectl apply -f deployment/k8s/streaming-configmap.yaml -n $NAMESPACE

# Deploy databases (StatefulSets)
echo "💾 Deploying databases..."
kubectl apply -f deployment/k8s/postgres-statefulset.yaml -n $NAMESPACE
kubectl apply -f deployment/k8s/redis-statefulset.yaml -n $NAMESPACE

# Wait for databases
echo "⏳ Waiting for databases to be ready..."
kubectl rollout status statefulset/postgres -n $NAMESPACE --timeout=120s || true
kubectl rollout status statefulset/redis -n $NAMESPACE --timeout=60s || true

# Deploy application services
echo "🌐 Deploying application services..."
kubectl apply -f deployment/k8s/backend-deployment.yaml -n $NAMESPACE
kubectl apply -f deployment/k8s/backend-service.yaml -n $NAMESPACE

# Deploy MediaSoup SFU
echo "📹 Deploying MediaSoup SFU..."
kubectl apply -f deployment/k8s/mediasoup-deployment.yaml -n $NAMESPACE

# Apply HPA for auto-scaling
echo "📈 Configuring auto-scaling..."
kubectl apply -f deployment/k8s/streaming-hpa.yaml -n $NAMESPACE
kubectl apply -f deployment/k8s/hpa.yaml -n $NAMESPACE

# Apply Ingress
echo "🔀 Configuring ingress..."
kubectl apply -f deployment/k8s/streaming-ingress.yaml -n $NAMESPACE
kubectl apply -f deployment/k8s/ingress.yaml -n $NAMESPACE

# Wait for deployments
echo "⏳ Waiting for deployments..."
kubectl rollout status deployment/vtp-backend -n $NAMESPACE --timeout=120s
kubectl rollout status deployment/mediasoup-sfu -n $NAMESPACE --timeout=120s

echo ""
echo "✅ Deployment complete!"
echo ""
echo "📊 Cluster Status:"
kubectl get pods -n $NAMESPACE
echo ""
kubectl get services -n $NAMESPACE
echo ""
kubectl get hpa -n $NAMESPACE
echo ""
echo "🔗 Access your application:"
echo "   kubectl port-forward svc/vtp-backend 8080:8080 -n $NAMESPACE"
echo "   kubectl port-forward svc/mediasoup-sfu 3002:3000 -n $NAMESPACE"
