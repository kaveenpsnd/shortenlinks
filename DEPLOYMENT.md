# Deployment Guide - shrten.live

## Prerequisites

1. **GitHub Repository** - Push this code to GitHub
2. **Azure Container Registry** - Already configured at `kaveenazuredev.azurecr.io`
3. **Kubernetes Cluster** (AKS) - Already running
4. **Domain DNS** - Point shrten.live domains to your cluster

## GitHub Secrets Required

Configure these secrets in your GitHub repository (Settings → Secrets and variables → Actions):

```
ACR_USERNAME          - Your Azure Container Registry username
ACR_PASSWORD          - Your Azure Container Registry password/token
KUBE_CONFIG           - Your kubectl config (base64 encoded)
FIREBASE_SERVICE_ACCOUNT - Firebase service account JSON (base64 encoded)
```

### Getting Your Secrets

#### KUBE_CONFIG
```bash
cat ~/.kube/config | base64 -w 0
```

#### FIREBASE_SERVICE_ACCOUNT
```bash
cat serviceAccountKey.json | base64 -w 0
```

## DNS Configuration

Point these DNS records to your Kubernetes cluster's LoadBalancer IP:

```
shrten.live         A     <CLUSTER_IP>
api.shrten.live     A     <CLUSTER_IP>
```

Get your cluster IP:
```bash
kubectl get ingress url-shortener-ingress -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
```

## Manual Deployment

If you need to deploy manually:

```bash
# Build and push images
docker build -t kaveenazuredev.azurecr.io/backend:latest .
docker build -t kaveenazuredev.azurecr.io/frontend:latest ./frontend

docker push kaveenazuredev.azurecr.io/backend:latest
docker push kaveenazuredev.azurecr.io/frontend:latest

# Apply Kubernetes manifests
kubectl apply -f k8s/cluster-issuer.yaml
kubectl apply -f k8s/postgres.yaml
kubectl apply -f k8s/redis.yaml
kubectl apply -f k8s/backend.yaml
kubectl apply -f k8s/frontend.yaml
kubectl apply -f k8s/ingress.yaml

# Create Firebase secret (if not exists)
kubectl create secret generic firebase-key --from-file=serviceAccountKey.json
```

## Automated Deployment (GitHub Actions)

Every push to `main` branch will automatically:

1. Build backend and frontend Docker images
2. Push to Azure Container Registry with commit SHA and "latest" tags
3. Update Kubernetes deployments
4. Apply all manifests
5. Wait for rollout completion

Trigger manual deployment:
- Go to Actions tab in GitHub
- Select "Deploy to Kubernetes" workflow
- Click "Run workflow"

## Monitoring

```bash
# Check pods
kubectl get pods

# Check services
kubectl get services

# Check ingress
kubectl get ingress

# View logs
kubectl logs -f deployment/backend-deployment
kubectl logs -f deployment/frontend-deployment

# Check certificate status
kubectl get certificate
kubectl describe certificate url-shortener-tls
```

## Rollback

```bash
# Rollback backend
kubectl rollout undo deployment/backend-deployment

# Rollback frontend
kubectl rollout undo deployment/frontend-deployment
```

## Architecture

- **Frontend**: React/Vite → Nginx (shrten.live)
- **Backend**: Go/Gin API (api.shrten.live)
- **Database**: PostgreSQL 15
- **Cache**: Redis 7
- **Ingress**: NGINX with Let's Encrypt TLS
- **Registry**: Azure Container Registry
