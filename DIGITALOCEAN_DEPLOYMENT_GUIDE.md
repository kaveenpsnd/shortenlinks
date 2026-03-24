# DigitalOcean Deployment Guide - Complete Setup

## Overview

This guide walks you through deploying your URL Shortener application from Azure AKS to a DigitalOcean Ubuntu Droplet using Docker Compose and automated CI/CD with GitHub Container Registry (GHCR).

**Timeline:** ~2-3 hours for first-time setup, then ~5 minutes per deployment afterward.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [1. Create DigitalOcean Account & Droplet](#1-create-digitalocean-account--droplet)
3. [2. Configure SSH Access](#2-configure-ssh-access)
4. [3. Set Up Droplet Environment](#3-set-up-droplet-environment)
5. [4. Configure GitHub Repository](#4-configure-github-repository)
6. [5. Configure DNS & SSL/TLS](#5-configure-dns--ssltls)
7. [6. Deploy Application](#6-deploy-application)
8. [7. Verify Deployment](#7-verify-deployment)
9. [Maintenance & Troubleshooting](#maintenance--troubleshooting)

---

## Prerequisites

Before starting, ensure you have:

- [ ] GitHub repository with this project
- [ ] DigitalOcean account (create one at https://www.digitalocean.com)
- [ ] Domain name (e.g., shrten.link)
- [ ] Docker images building successfully (backend in `cmd/api/`, frontend in `frontend/`)
- [ ] SSH key pair (or follow Generation steps below)

---

## 1. Create DigitalOcean Account & Droplet

### Step 1.1: Create Account
1. Go to https://www.digitalocean.com
2. Sign up for a new account
3. Add payment method (credit/debit card)

### Step 1.2: Create a Droplet
1. Click **Create** → **Droplets**
2. Choose image: **Ubuntu 22.04 x64**
3. Choose size: **Basic** → **Regular with SSD** → **$6/month** (2GB RAM, 1 CPU, 50GB SSD)
4. Choose datacenter: Pick closest to your users (e.g., New York, San Francisco)
5. Authentication: **New SSH Key** (or use existing)
   - Follow the prompt to generate and download your SSH key
   - Or paste your public SSH key
6. Hostname: `url-shortener`
7. Click **Create Droplet**

Wait ~1-2 minutes for the droplet to boot up. You'll see the public IP address displayed (e.g., `192.168.1.100`).

---

## 2. Configure SSH Access

### Step 2.1: Generate SSH Key (if you don't have one)

On your **local machine**:

```powershell
# Windows PowerShell
ssh-keygen -t rsa -b 4096 -f $env:USERPROFILE\.ssh\digitalocean_key -C "github-actions"
```

Or on **macOS/Linux**:

```bash
ssh-keygen -t rsa -b 4096 -f ~/.ssh/digitalocean_key -C "github-actions"
```

This creates:
- Private key: `digitalocean_key` (keep secure!)
- Public key: `digitalocean_key.pub` (goes on droplet)

### Step 2.2: SSH into Droplet

Using the SSH key downloaded from DigitalOcean:

```bash
# Replace YOUR_DROPLET_IP with actual IP
ssh -i /path/to/private/key root@YOUR_DROPLET_IP
```

Or if you used password authentication, DigitalOcean emails you the password:

```bash
ssh root@YOUR_DROPLET_IP
```

### Step 2.3: Add Your SSH Key to Authorized Keys

Once SSH'd into the droplet:

```bash
# Create .ssh directory if it doesn't exist
mkdir -p ~/.ssh
chmod 700 ~/.ssh

# Add your public key
cat >> ~/.ssh/authorized_keys << 'EOF'
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDx... (contents of digitalocean_key.pub)
EOF

# Set correct permissions
chmod 600 ~/.ssh/authorized_keys

# Verify
cat ~/.ssh/authorized_keys
```

Test the new SSH key:

```bash
ssh -i digitalocean_key root@YOUR_DROPLET_IP
```

---

## 3. Set Up Droplet Environment

SSH into your droplet and execute the following:

### Step 3.1: Update System

```bash
apt update && apt upgrade -y
apt install -y curl git nano
```

### Step 3.2: Install Docker & Docker Compose

```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
rm get-docker.sh

# Add your user to docker group (optional, for non-root docker)
usermod -aG docker root

# Verify Docker installation
docker --version
docker compose version
```

### Step 3.3: Create Project Directory

```bash
mkdir -p /root/url-shortener
cd /root/url-shortener
chmod 755 /root/url-shortener
```

### Step 3.4: Create Certificate Directory

```bash
mkdir -p /root/url-shortener/certs
chmod 755 /root/url-shortener/certs
```

### Step 3.5: Clone Repository (Initial Setup)

```bash
cd /root/url-shortener
git clone https://github.com/yourusername/url-shortener.git .
git checkout main
```

### Step 3.6: Create Initial .env File

```bash
cat > /root/url-shortener/.env << 'EOF'
GHCR_USERNAME=yourgithubusername
DB_USER=postgres
DB_PASSWORD=GenerateStrongPassword123!
DB_NAME=urlshortener
FIREBASE_PROJECT_ID=your-firebase-project-id
VITE_API_URL=https://api.shrten.link
VITE_FIREBASE_CONFIG={"apiKey":"..."}
GIN_MODE=release
EOF

# Secure the .env file
chmod 600 /root/url-shortener/.env
```

---

## 4. Configure GitHub Repository

### Step 4.1: Add Repository Secrets

Go to your GitHub repository:

1. **Settings** → **Secrets and variables** → **Actions**
2. Click **New repository secret**
3. Add the following secrets:

```
Name: DO_DROPLET_IP
Value: [Your DigitalOcean droplet IP]

Name: DO_SSH_USER
Value: root

Name: DO_SSH_KEY
Value: [Contents of your private SSH key - include -----BEGIN and -----END lines]

Name: DO_PROJECT_PATH
Value: /root/url-shortener

Name: DB_USER
Value: postgres

Name: DB_PASSWORD
Value: [Strong password - same as in .env]

Name: DB_NAME
Value: urlshortener

Name: FIREBASE_PROJECT_ID
Value: [Your Firebase project ID]

Name: VITE_API_URL
Value: https://api.shrten.link

Name: VITE_FIREBASE_CONFIG
Value: [Your Firebase config JSON]

Name: SLACK_WEBHOOK_URL (optional)
Value: [Your Slack webhook for notifications]
```

### Step 4.2: Verify Dockerfiles

Ensure you have:
- `cmd/api/Dockerfile` for backend
- `frontend/Dockerfile` for frontend

Example backend Dockerfile (`cmd/api/Dockerfile`):

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/api

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates curl

WORKDIR /app
COPY --from=builder /app/api .

EXPOSE 8080
CMD ["./api"]
```

Example frontend Dockerfile (`frontend/Dockerfile`):

```dockerfile
# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

# Runtime stage
FROM node:18-alpine
WORKDIR /app
RUN npm install -g http-server

COPY --from=builder /app/dist ./dist

EXPOSE 3000
CMD ["http-server", "dist", "-p", "3000", "--cors"]
```

### Step 4.3: Test GitHub Actions Workflow

1. Go to **Actions** tab in GitHub
2. You should see the `Deploy to DigitalOcean` workflow
3. Manually trigger it: **Run workflow** (optional for testing)

---

## 5. Configure DNS & SSL/TLS

### Step 5.1: Update DNS Records

Go to your domain registrar (e.g., GoDaddy, Namecheap) and add/update:

| Type | Host | Value |
|------|------|-------|
| A | @ | YOUR_DROPLET_IP |
| A | www | YOUR_DROPLET_IP |
| A | api | YOUR_DROPLET_IP |

Example for `shrten.link`:
- `shrten.link` → YOUR_DROPLET_IP
- `www.shrten.link` → YOUR_DROPLET_IP
- `api.shrten.link` → YOUR_DROPLET_IP

Wait 2-5 minutes for DNS propagation:

```bash
# Check DNS resolution
nslookup shrten.link
nslookup api.shrten.link
```

### Step 5.2: Generate SSL/TLS Certificates

On your **local machine**, generate certificates using Let's Encrypt (requires `certbot`):

```bash
# Install certbot
# macOS: brew install certbot
# Windows: choco install certbot
# Ubuntu: sudo apt install certbot

# Generate certificate (replace with your domain)
certbot certonly --manual \
  -d shrten.link \
  -d www.shrten.link \
  -d api.shrten.link \
  --agree-tos \
  --email your-email@example.com

# This will generate files in:
# /etc/letsencrypt/live/shrten.link/
# - fullchain.pem (certificate)
# - privkey.pem (private key)
```

### Step 5.3: Upload Certificates to Droplet

```bash
# From your local machine, copy certificates to droplet
scp -i digitalocean_key \
  /etc/letsencrypt/live/shrten.link/fullchain.pem \
  root@YOUR_DROPLET_IP:/root/url-shortener/certs/shrten.link.crt

scp -i digitalocean_key \
  /etc/letsencrypt/live/shrten.link/privkey.pem \
  root@YOUR_DROPLET_IP:/root/url-shortener/certs/shrten.link.key

# Set permissions on droplet
ssh -i digitalocean_key root@YOUR_DROPLET_IP
chmod 600 /root/url-shortener/certs/*
```

---

## 6. Deploy Application

### Automatic Deployment (via GitHub Actions)

Simply push to the `main` branch:

```bash
git add .
git commit -m "Deploy to DigitalOcean"
git push origin main
```

GitHub Actions will automatically:
1. Build backend image
2. Build frontend image
3. Push to GHCR
4. SSH into droplet
5. Pull latest images
6. Run `docker compose up -d`

Monitor the deployment:
1. Go to GitHub → **Actions** tab
2. Click the latest workflow run
3. Watch the deployment steps

### Manual Deployment (if needed)

SSH into your droplet and run:

```bash
cd /root/url-shortener

# Pull latest code
git pull origin main

# Update .env with latest secrets if needed
# (Usually done by GitHub Actions)

# Log in to GHCR
echo "YOUR_GITHUB_TOKEN" | docker login ghcr.io -u yourusername --password-stdin

# Pull latest images
docker pull ghcr.io/yourusername/url-shortener-backend:latest
docker pull ghcr.io/yourusername/url-shortener-frontend:latest

# Deploy
docker compose -f docker-compose.prod.yml up -d --force-recreate

# Check status
docker compose -f docker-compose.prod.yml ps

# View logs
docker compose -f docker-compose.prod.yml logs -f
```

---

## 7. Verify Deployment

### Health Checks

```bash
# From your local machine
curl -I https://api.shrten.link/health
curl -I https://shrten.link

# From the droplet
curl http://localhost:8080/health
curl http://localhost:3000
```

### SSH into Droplet for Manual Verification

```bash
ssh -i digitalocean_key root@YOUR_DROPLET_IP

# View running containers
docker ps

# Check logs
docker compose -f docker-compose.prod.yml logs

# Test backend
docker exec url-shortener-backend curl -f http://localhost:8080/health

# Test database connection
docker exec url-shortener-postgres psql -U postgres -d urlshortener -c "SELECT 1;"

# Check available disk space
df -h

# Check memory usage
free -h
```

### Browser Tests

1. Open https://shrten.link in your browser
2. Try creating a shortened URL
3. Try accessing an API endpoint: https://api.shrten.link/api/links

---

## Maintenance & Troubleshooting

### View Application Logs

```bash
ssh -i digitalocean_key root@YOUR_DROPLET_IP

# All services
docker compose -f docker-compose.prod.yml logs -f

# Specific service
docker compose -f docker-compose.prod.yml logs -f backend
docker compose -f docker-compose.prod.yml logs -f frontend
docker compose -f docker-compose.prod.yml logs -f postgres
docker compose -f docker-compose.prod.yml logs -f redis
```

### Restart Services

```bash
# Restart all services
docker compose -f docker-compose.prod.yml restart

# Restart specific service
docker compose -f docker-compose.prod.yml restart backend
```

### Stop/Start Services

```bash
# Stop all
docker compose -f docker-compose.prod.yml down

# Start all
docker compose -f docker-compose.prod.yml up -d
```

### Database Backup & Restore

```bash
# Backup database
docker exec url-shortener-postgres pg_dump -U postgres -d urlshortener > backup.sql

# Restore database
docker exec -i url-shortener-postgres psql -U postgres -d urlshortener < backup.sql
```

### SSL Certificate Renewal

Let's Encrypt certificates expire every 90 days. To renew:

```bash
# From local machine
certbot renew

# Upload new certificates to droplet
scp -i digitalocean_key \
  /etc/letsencrypt/live/shrten.link/fullchain.pem \
  root@YOUR_DROPLET_IP:/root/url-shortener/certs/shrten.link.crt

scp -i digitalocean_key \
  /etc/letsencrypt/live/shrten.link/privkey.pem \
  root@YOUR_DROPLET_IP:/root/url-shortener/certs/shrten.link.key

# Reload Nginx on droplet
docker exec url-shortener-nginx nginx -s reload
```

### Common Issues

#### Port 80/443 Already in Use
```bash
# Check what's using the ports
sudo lsof -i :80
sudo lsof -i :443

# Stop the service using them
docker compose -f docker-compose.prod.yml down
```

#### Backend Health Check Failing
```bash
# Check backend logs
docker compose -f docker-compose.prod.yml logs backend

# Verify database connection
docker exec url-shortener-backend env | grep DB_
```

#### Domain Not Resolving
```bash
# Check DNS
nslookup shrten.link
dig api.shrten.link

# Wait for DNS propagation (can take 5+ minutes)
```

#### Out of Disk Space
```bash
# Check usage
df -h

# Remove old Docker images
docker image prune -a

# Remove unused volumes
docker volume prune
```

---

## Storage & Scaling

### Current Configuration
- **Droplet:** 2GB RAM, 1 CPU, 50GB SSD
- **Database:** PostgreSQL with persistent volume
- **Cache:** Redis with persistent volume
- **Expected Load:** 1,000-5,000 requests/minute

### If You Need to Scale

1. **Upgrade Droplet:** DigitalOcean → **Resize** (change CPU/RAM)
2. **Increase Storage:** Add block storage volume
3. **Load Balancer:** DigitalOcean LoadBalancer to distribute traffic
4. **Managed Databases:** Use DigitalOcean's managed PostgreSQL instead

---

## Security Checklist

- [ ] SSH key authentication enabled (no password auth)
- [ ] Firewall configured (only 80, 443, 22 open to 0.0.0.0)
- [ ] Database password is strong (16+ chars, mixed case, numbers, symbols)
- [ ] `.env` file is not in Git repository
- [ ] Secrets are secured in GitHub (all marked as secrets)
- [ ] SSL/TLS certificates valid and renewed before expiry
- [ ] Regular backups of database
- [ ] Monitoring and alerting configured

---

## Support & Resources

- **DigitalOcean Docs:** https://docs.digitalocean.com/
- **Docker Docs:** https://docs.docker.com/
- **Let's Encrypt:** https://letsencrypt.org/
- **GitHub Actions:** https://docs.github.com/actions

---

## Quick Reference Commands

```bash
# SSH into droplet
ssh -i digitalocean_key root@YOUR_DROPLET_IP

# Restart services
docker compose -f docker-compose.prod.yml restart

# View logs
docker compose -f docker-compose.prod.yml logs -f

# Check status
docker compose -f docker-compose.prod.yml ps

# Pull latest code
git pull origin main

# Manual deployment
docker compose -f docker-compose.prod.yml up -d --force-recreate

# Check disk space
df -h

# Check memory
free -h
```

