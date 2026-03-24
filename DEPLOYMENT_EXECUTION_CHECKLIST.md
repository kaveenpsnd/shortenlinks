# DEPLOYMENT EXECUTION CHECKLIST

**Status:** ✅ Application Ready for Deployment  
**Date:** March 24, 2026  
**Project:** URL Shortener (Azure AKS → DigitalOcean)  

---

## ✅ PHASE 1: CODE FIXES (COMPLETED)

- [x] Fixed Go version (1.21 → 1.25)
- [x] Removed broken Alpine user creation from Dockerfiles
- [x] Added health check endpoint `/health` to backend
- [x] Updated .gitignore to allow workflow files
- [x] Committed all changes to GitHub
- [x] Generated validation report

**Commit:** `4618c25` - feat: add health check endpoint and pre-deployment validation report

---

## 🔄 PHASE 2: GITHUB ACTIONS WORKFLOW (AUTOMATIC)

### Timeline: When you commit to main
- [ ] Push triggers workflow automatically (deploy.yml)
- [ ] **Step 1:** Checkout code (1 min)
- [ ] **Step 2:** Build backend Docker image (3-5 min)
  - Compile Go binary
  - Push to ghcr.io/kaveenpsnd/url-shortener-backend:latest
- [ ] **Step 3:** Build frontend Docker image (2-3 min)
  - Run npm build (Vite)
  - Push to ghcr.io/kaveenpsnd/url-shortener-frontend:latest
- [ ] **Step 4:** SSH deploy step (requires secrets)
  - Connects to droplet: 64.227.166.226
  - Pulls latest images from GHCR
  - Runs docker-compose up

**Total Time:** 8-12 minutes

**Monitor at:** https://github.com/kaveenpsnd/shortenlinks/actions

---

## ⏳ PHASE 3: DROPLET PREPARATION (Before Workflow)

### 3.1 SSH Access Test
```bash
ssh -i ~/.ssh/do_github_actions root@64.227.166.226
```
- [ ] Connection successful
- [ ] Docker installed and running
- [ ] Git installed and working

### 3.2 Repository Setup
```bash
cd /root
git clone https://github.com/kaveenpsnd/shortenlinks.git url-shortener-new
cd url-shortener-new
```
- [ ] Repository cloned successfully
- [ ] All files present (migrations/, frontend/, etc.)
- [ ] Go modules downloaded (go.mod/go.sum exist)

### 3.3 Directory Structure
```bash
mkdir -p /root/url-shortener-new/certs
chmod 755 /root/url-shortener-new/certs
```
- [ ] Certs directory created
- [ ] Permissions set correctly

### 3.4 Environment Configuration
The workflow will create `.env` automatically, but verify:
```bash
ls -la /root/url-shortener-new/.env
cat /root/url-shortener-new/.env
```
Expected variables:
- [ ] GHCR_USERNAME=kaveenpsnd
- [ ] DB_USER=postgres
- [ ] DB_PASSWORD=(secret from GitHub)
- [ ] DB_NAME=shortener
- [ ] FIREBASE_PROJECT_ID=urlshortner-138d6
- [ ] VITE_API_URL=https://api.shrten.link
- [ ] VITE_FIREBASE_CONFIG=(JSON config)
- [ ] GIN_MODE=release

---

## 🔐 PHASE 4: SSL CERTIFICATE GENERATION (After Workflow)

### Prerequisites
- Droplet is running and services are up
- DNS records updated (see Phase 5)
- HTTP traffic reaching NGINX on port 80

### Generate Certificates (Run on Droplet)
```bash
# Install certbot
sudo apt update
sudo apt install -y certbot python3-certbot-nginx

# Generate certificates (auto-renewal configured)
sudo certbot certonly \
  --standalone \
  -d shrten.link \
  -d www.shrten.link \
  -d api.shrten.link \
  --non-interactive \
  --agree-tos \
  --email your-email@example.com

# Copy to Docker volume
sudo cp /etc/letsencrypt/live/shrten.link/fullchain.pem /root/url-shortener-new/certs/shrten.link.crt
sudo cp /etc/letsencrypt/live/shrten.link/privkey.pem /root/url-shortener-new/certs/shrten.link.key

# Fix permissions
sudo chown 1000:1000 /root/url-shortener-new/certs/*
```

- [ ] Certificates generated successfully
- [ ] Files in `/root/url-shortener-new/certs/shrten.link.{crt,key}`
- [ ] Restart NGINX to load certificates:
  ```bash
  docker compose -f docker-compose.prod.yml restart nginx
  ```

---

## 🌐 PHASE 5: DNS CONFIGURATION (External)

Update DNS records at your domain registrar (Namecheap, GoDaddy, etc.):

### Records to Add/Update
| Type | Name | Value |
|------|------|-------|
| A | @ (or shrten.link) | 64.227.166.226 |
| A | www | 64.227.166.226 |
| A | api | 64.227.166.226 |

### Verification (run after DNS propagates ~5-30 min)
```bash
# Check DNS resolution
nslookup shrten.link
nslookup api.shrten.link

# Verify HTTP is redirected to HTTPS
curl -i http://shrten.link
# Should see: 301 Moved Permanently (Location: https://...)

# Verify HTTPS works
curl -i https://shrten.link
# Should see: 200 OK or redirect to frontend
```

- [ ] DNS @ record points to 64.227.166.226
- [ ] DNS www record points to 64.227.166.226
- [ ] DNS api record points to 64.227.166.226
- [ ] DNS has propagated (wait 5-30 min)

---

## 🚀 PHASE 6: DEPLOYMENT EXECUTION

### 6.1 Trigger GitHub Actions
**Option A: Automatic (Already Done)**
```bash
git push origin main
# Workflow automatically triggers
```

**Option B: Manual Trigger**
1. Go to: https://github.com/kaveenpsnd/shortenlinks/actions
2. Select "Deploy to DigitalOcean" workflow
3. Click "Run workflow" button

- [ ] Workflow started
- [ ] All build steps passing
- [ ] GHCR images pushed successfully

### 6.2 Monitor Workflow Execution
```bash
# On your local machine, watch the workflow:
# https://github.com/kaveenpsnd/shortenlinks/actions

# Or SSH to droplet and check services:
ssh -i ~/.ssh/do_github_actions root@64.227.166.226
docker compose -f docker-compose.prod.yml ps
```

Expected output after workflow completes:
```
NAME                    STATUS                      PORTS
url-shortener-nginx     Up (healthy)                0.0.0.0:80->80/tcp, 0.0.0.0:443->443/tcp
url-shortener-frontend  Up (healthy)                3000/tcp
url-shortener-backend   Up (healthy)                8080/tcp
url-shortener-redis     Up (healthy)                6379/tcp
url-shortener-postgres  Up (healthy)                5432/tcp
```

- [ ] All 5 services are running
- [ ] All services show "healthy" status
- [ ] Logs show no errors: `docker compose -f docker-compose.prod.yml logs`

### 6.3 Verify Database
```bash
# Connect to database inside container
docker compose -f docker-compose.prod.yml exec postgres psql -U postgres -d shortener

# Check tables were created by migrations
\dt
# Should see: short_links, users

# Exit
\q
```

- [ ] Database connection successful
- [ ] Tables created: short_links, users
- [ ] Migrations ran successfully

---

## ✔️ PHASE 7: FUNCTIONAL TESTING

### 7.1 Backend Health Check
```bash
curl https://api.shrten.link/health
# Expected: {"status":"healthy"}
```

- [ ] Backend health endpoint responds
- [ ] Returns 200 OK

### 7.2 Create Short Link (Test)
```bash
curl -X POST https://api.shrten.link/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"original_url":"https://www.github.com/kaveenpsnd"}'

# Expected: {"short_code":"xyz123","original_url":"https://www.github.com/kaveenpsnd",...}
```

- [ ] Short link creation works
- [ ] Returns valid response with short_code

### 7.3 Redirect Test
```bash
curl -L https://shrten.link/xyz123
# Should redirect to the original URL
```

- [ ] Redirect works correctly

### 7.4 Frontend Access
```bash
# Open in browser:
https://shrten.link
```

- [ ] Frontend loads successfully
- [ ] No console errors in browser DevTools
- [ ] Firebase authentication works
- [ ] Can create links through UI

### 7.5 API Subdomain Access
```bash
curl https://api.shrten.link/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"original_url":"https://example.com"}'
```

- [ ] API subdomain works
- [ ] CORS headers present

---

## 📊 POST-DEPLOYMENT CHECKLIST

- [ ] Application accessible at https://shrten.link
- [ ] API accessible at https://api.shrten.link  
- [ ] SSL certificates valid (no browser warnings)
- [ ] All 5 Docker services healthy
- [ ] Database persisting data (no data loss)
- [ ] Redis caching working
- [ ] Firebase authentication functional
- [ ] Admin routes accessible (with auth)
- [ ] Health checks passing
- [ ] Logs show no errors

---

## 🛑 TROUBLESHOOTING

### If GitHub Actions workflow fails:
1. Check build logs: https://github.com/kaveenpsnd/shortenlinks/actions
2. Verify Git commits were pushed
3. Confirm all GitHub secrets are set (Settings → Secrets)
4. Retry workflow manually

### If services don't start on droplet:
1. SSH into droplet
2. Check Docker logs: `docker compose -f docker-compose.prod.yml logs`
3. Verify .env file has all required variables
4. Restart services: `docker compose -f docker-compose.prod.yml down && docker compose -f docker-compose.prod.yml up -d`

### If SSL certificates fail:
1. Ensure DNS records are updated and propagated
2. Ensure ports 80/443 are open
3. Check certbot logs: `sudo certbot logs`
4. Manually generate: See Phase 4 above

### If frontend/backend images fail to build:
1. Check Dockerfile syntax
2. Verify Go version: Should be 1.25
3. Verify Node version: Should be 18
4. Check package.json build script exists
5. Review GitHub Actions logs for specific errors

---

## 📝 FINAL VERIFICATION

After all phases complete, run this test:

```bash
#!/bin/bash

# Test health endpoint
echo "Testing health endpoint..."
HEALTH=$(curl -s https://api.shrten.link/health)
[ "$HEALTH" = '{"status":"healthy"}' ] && echo "✅ Health OK" || echo "❌ Health FAILED"

# Test short link creation
echo "Creating test short link..."
CREATE=$(curl -s -X POST https://api.shrten.link/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"original_url":"https://github.com/kaveenpsnd"}')
echo "$CREATE"

# Test redirect
echo "Testing redirect..."
REDIRECT=$(curl -s -L -o /dev/null -w "%{http_code}" https://shrten.link/abc123)
[ "$REDIRECT" = "200" ] && echo "✅ Redirect OK" || echo "❌ Redirect FAILED"

# Test frontend
echo "Testing frontend..."
FRONTEND=$(curl -s -o /dev/null -w "%{http_code}" https://shrten.link)
[ "$FRONTEND" = "200" ] && echo "✅ Frontend OK" || echo "❌ Frontend FAILED"

echo ""
echo "All tests completed!"
```

---

## 🎯 SUCCESS CRITERIA

✅ **Deployment is successful when:**
- All GitHub Actions workflow steps pass
- All 5 Docker services running and healthy
- Frontend accessible at https://shrten.link
- API accessible at https://api.shrten.link
- Database connections working
- Short links creating and redirecting correctly
- No SSL certificate warnings
- No error logs in docker-compose logs
- Firebase authentication functional

**Estimated total time to deployment: 30-45 minutes**

