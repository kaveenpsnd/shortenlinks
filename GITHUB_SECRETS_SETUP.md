# GitHub Repository Secrets - Required Setup

## Overview
To enable the CI/CD pipeline to deploy your URL Shortener application to DigitalOcean, you need to configure the following secrets in your GitHub repository. Navigate to **Settings → Secrets and variables → Actions** to add these.

---

## Required Secrets

### DigitalOcean Droplet Configuration
These secrets authenticate your GitHub Actions with your DigitalOcean droplet.

| Secret Name | Description | Example Value |
|---|---|---|
| `DO_DROPLET_IP` | Public IP address of your DigitalOcean Droplet | `192.168.1.100` |
| `DO_SSH_USER` | SSH username (usually `root` or your created user) | `root` |
| `DO_SSH_KEY` | Private SSH key for authentication | (See below for setup) |
| `DO_PROJECT_PATH` | Full path to the project directory on the droplet | `/root/url-shortener` |

### Database Configuration
Connect to PostgreSQL running in your Docker container.

| Secret Name | Description | Example Value |
|---|---|---|
| `DB_USER` | PostgreSQL database username | `postgres` |
| `DB_PASSWORD` | PostgreSQL database password (strong password!) | `SecureP@ssw0rd123!xyz` |
| `DB_NAME` | Database name | `urlshortener` |

### Frontend Configuration
Environment variables for your React/Vite application.

| Secret Name | Description | Example Value |
|---|---|---|
| `VITE_API_URL` | API endpoint for frontend to connect to | `https://api.shrten.link` |
| `VITE_FIREBASE_CONFIG` | Firebase configuration JSON (if using Firebase) | `{"apiKey":"...","projectId":"..."}` |

### Backend Configuration
Environment variables for your Go backend.

| Secret Name | Description | Example Value |
|---|---|---|
| `FIREBASE_PROJECT_ID` | Firebase project ID (if using Firebase) | `my-firebase-project` |

### Notifications (Optional)
For deployment status notifications.

| Secret Name | Description | Example Value |
|---|---|---|
| `SLACK_WEBHOOK_URL` | Slack webhook URL for notifications (optional) | `https://hooks.slack.com/services/YOUR/WEBHOOK/HERE` |

---

## Step-by-Step Secret Setup Guide

### 1. Generate SSH Key Pair (if you don't have one)

On your local machine:

```bash
ssh-keygen -t rsa -b 4096 -f digitalocean_key -C "github-actions"
```

This creates two files:
- `digitalocean_key` (private key - keep secret!)
- `digitalocean_key.pub` (public key - goes on droplet)

### 2. Add Public Key to DigitalOcean Droplet

SSH into your droplet:

```bash
ssh root@YOUR_DROPLET_IP
```

Add your public key:

```bash
mkdir -p ~/.ssh
cat >> ~/.ssh/authorized_keys << EOF
<contents of digitalocean_key.pub>
EOF

chmod 600 ~/.ssh/authorized_keys
chmod 700 ~/.ssh
```

### 3. Create Project Directory on Droplet

```bash
mkdir -p /root/url-shortener
cd /root/url-shortener
```

### 4. Add Secrets to GitHub

Go to your GitHub repository:

1. Click **Settings**
2. Select **Secrets and variables → Actions**
3. Click **New repository secret**
4. Add each secret from the table below:

#### Droplet Secrets:
```
DO_DROPLET_IP: <your_droplet_ip>
DO_SSH_USER: root
DO_SSH_KEY: <contents_of_digitalocean_key_private_key>
DO_PROJECT_PATH: /root/url-shortener
```

#### Database Secrets:
```
DB_USER: postgres
DB_PASSWORD: <generate_strong_password>
DB_NAME: urlshortener
```

#### Frontend Secrets:
```
VITE_API_URL: https://api.shrten.link
VITE_FIREBASE_CONFIG: {"apiKey":"...","projectId":"..."}  (if using Firebase)
```

#### Backend Secrets:
```
FIREBASE_PROJECT_ID: <your_firebase_project_id>  (if using Firebase)
```

#### Optional Notifications:
```
SLACK_WEBHOOK_URL: <your_slack_webhook_url>  (if using Slack)
```

---

## Pre-Deployment Checklist

- [ ] DigitalOcean droplet created and running (Ubuntu 22.04 LTS)
- [ ] Docker and Docker Compose installed on droplet
- [ ] SSH key pair generated
- [ ] Public SSH key added to droplet's `~/.ssh/authorized_keys`
- [ ] Project directory created at `/root/url-shortener`
- [ ] All secrets added to GitHub repository
- [ ] Backend Dockerfile exists at `cmd/api/Dockerfile`
- [ ] Frontend Dockerfile exists at `frontend/Dockerfile`
- [ ] Domain DNS records updated to point to droplet IP
- [ ] SSL certificates ready (or Let's Encrypt configured)

---

## Verification

### Test SSH Connection from Local Machine

```bash
ssh -i digitalocean_key root@YOUR_DROPLET_IP
```

If this works, your GitHub Actions will be able to SSH into the droplet.

### Test GitHub Actions Secret Access

Add a test workflow to verify secrets are accessible:

```yaml
name: Test Secrets
on: [workflow_dispatch]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Check secrets
        run: |
          echo "DO_DROPLET_IP is set: ${{ secrets.DO_DROPLET_IP != '' }}"
          echo "DO_SSH_KEY is set: ${{ secrets.DO_SSH_KEY != '' }}"
          echo "DB_PASSWORD is set: ${{ secrets.DB_PASSWORD != '' }}"
```

---

## Security Best Practices

1. **SSH Key**: Keep your private key secure. Never commit it to version control.
2. **Database Password**: Use a strong, randomly generated password (min 16 characters, include uppercase, lowercase, numbers, symbols).
3. **Token Rotation**: Periodically rotate your SSH keys and database passwords.
4. **Least Privilege**: Use a dedicated SSH user instead of `root` if possible.
5. **Firewall**: Restrict SSH access to your droplet IP ranges when possible.
6. **Audit**: Monitor GitHub Actions logs for unauthorized access attempts.

---

## Troubleshooting

### Authentication Failed
- Verify `DO_SSH_KEY` contains the entire private key including `BEGIN RSA PRIVATE KEY` and `END RSA PRIVATE KEY` lines
- Check that public key is in droplet's `~/.ssh/authorized_keys`
- Ensure SSH permissions are correct: 600 for key file, 700 for `.ssh` directory

### Docker Push Failed
- Verify you're using the correct GHCR URL format
- Check that GitHub token has `packages:write` permission
- Ensure the `docker login` command uses correct username and token

### Health Check Failed
- SSH into droplet and check logs: `docker compose -f docker-compose.prod.yml logs`
- Verify environment variables are set correctly in `.env` file
- Check that backend is listening on port 8080
- Verify database migrations have completed

### Timeout during SSH
- Increase SSH timeout in the GitHub Actions workflow
- Check droplet firewall rules allow port 22
- Verify droplet is running and accessible

