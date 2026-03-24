# 🚀 Quick Start - Deploy in 10 Minutes

## What You Need Before Starting

- ✅ GitHub secrets added (7 total)
- ✅ DigitalOcean droplet IP
- ✅ Domain DNS updated
- ✅ SSH key configured
- ⏱️ Time: ~10 minutes

---

## Let's Deploy!

### Step 1: Verify Droplet (2 min)

SSH into your droplet:
```bash
ssh -i your-ssh-key root@DROPLET_IP
```

Check Docker is running:
```bash
docker --version
docker compose version
```

### Step 2: Clone & Setup (3 min)

```bash
cd /root/url-shortener
git clone https://github.com/YOUR-USERNAME/url-shortener.git .
git checkout main

# Create .env file with secrets from GitHub
cat > .env << 'EOF'
GHCR_USERNAME=your-github-username
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=shortener
FIREBASE_PROJECT_ID=urlshortner-138d6
VITE_API_URL=https://api.shrten.link
VITE_FIREBASE_CONFIG={"apiKey":"AIza...","authDomain":"..."}
GIN_MODE=release
EOF

chmod 600 .env
```

### Step 3: Deploy (3 min)

```bash
# Login to GHCR
echo "YOUR_GITHUB_TOKEN" | docker login ghcr.io -u YOUR-USERNAME --password-stdin

# Pull latest images
docker pull ghcr.io/YOUR-USERNAME/url-shortener-backend:latest
docker pull ghcr.io/YOUR-USERNAME/url-shortener-frontend:latest

# Start services
docker compose -f docker-compose.prod.yml up -d

# Wait for health checks
sleep 10

# Check status
docker compose -f docker-compose.prod.yml ps
```

### Step 4: Test (2 min)

From your local machine:
```bash
# Test API
curl -I https://api.shrten.link/health

# Test frontend
curl -I https://shrten.link

# Create test link
curl -X POST https://api.shrten.link/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"original_url": "https://example.com"}'
```

---

## ✅ Done!

Your application is now deployed at:
- 🌐 Frontend: https://shrten.link
- 🔌 API: https://api.shrten.link

---

## 🔄 Future Deployments (1 minute)

Just push to main branch and GitHub Actions handles everything:

```bash
git add .
git commit -m "Your changes"
git push origin main
```

That's it! 🎉

---

## 🆘 Troubleshooting

**Health check failing?**
```bash
docker compose -f docker-compose.prod.yml logs backend
```

**Port already in use?**
```bash
docker compose -f docker-compose.prod.yml down
docker compose -f docker-compose.prod.yml up -d
```

**Out of space?**
```bash
docker system prune -a
```

See `DIGITALOCEAN_DEPLOYMENT_GUIDE.md` for more help!
