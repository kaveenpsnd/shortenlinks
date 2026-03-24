# SSH Key Setup for GitHub Actions Deployment

## 🔴 What Failed

The GitHub Actions deployment failed because:
- SSH authentication to DigitalOcean droplet (64.227.166.226) failed
- Error: `ssh: handshake failed: ssh: unable to authenticate`
- Root cause: `DO_SSH_KEY` secret is either missing, invalid, or malformed

## ✅ What Was Fixed

Updated `.github/workflows/deploy.yml`:
- SSH action upgraded from v1.0.0 to master (more stable)
- Added timeouts: 60s for connection, 30m for commands  
- Slack notifications now use proper webhook action (rtCamp/action-slack)

## 🔧 Required: Configure SSH Key Secret

### Step 1: Verify SSH Key Exists Locally

On your Windows machine, check if the key exists:
```powershell
Test-Path "$HOME\.ssh\do_github_actions"
```

If it exists, view it:
```powershell
Get-Content "$HOME\.ssh\do_github_actions"
```

**Expected format:** Should start with one of:
```
-----BEGIN OPENSSH PRIVATE KEY-----
```
or
```
-----BEGIN RSA PRIVATE KEY-----
```
or
```
-----BEGIN ED25519 PRIVATE KEY-----
```

### Step 2: If Key Doesn't Exist, Generate New SSH Key

```powershell
# Generate new SSH key (replace email with your own)
ssh-keygen -t ed25519 -C "your-email@example.com" -f "$HOME\.ssh\do_github_actions" -N ""

# Verify it was created
Get-Content "$HOME\.ssh\do_github_actions"
```

### Step 3: Add Public Key to DigitalOcean Droplet

**Option A: Via SSH (if you have existing access)**
```bash
ssh -i "$HOME\.ssh\do_github_actions" root@64.227.166.226
cat ~/.ssh/authorized_keys
# Add the public key if not present
```

**Option B: Manually paste the public key**

First, get the public key:
```powershell
Get-Content "$HOME\.ssh\do_github_actions.pub"
```

Then SSH to droplet and add it:
```bash
echo "PASTE_PUBLIC_KEY_HERE" >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
```

### Step 4: Add Secret to GitHub

1. Go to: https://github.com/kaveenpsnd/shortenlinks/settings/secrets/actions
2. Click "New repository secret"
3. Name: `DO_SSH_KEY`
4. Value: Paste the **ENTIRE private key content** (from Step 1)
   - Include the `-----BEGIN...` and `-----END...` lines
   - Include all lines between them
   - Preserve all newlines exactly
5. Click "Add secret"

### Step 5: Test SSH Connection Locally

```powershell
# Test the connection
ssh -i "$HOME\.ssh\do_github_actions" root@64.227.166.226 "echo 'SSH works!'"

# Expected output: SSH works!
```

If this works, your GitHub Actions deployment should also work.

## 📋 GitHub Secrets Checklist

Verify all deployment secrets are configured:

| Secret Name | Required | Status | Value |
|-------------|----------|--------|-------|
| `DO_DROPLET_IP` | ✅ YES | Check | `64.227.166.226` |
| `DO_SSH_USER` | ✅ YES | Check | `root` |
| `DO_SSH_KEY` | ✅ YES | **FIX THIS** | Your private SSH key |
| `DO_PROJECT_PATH` | ✅ YES | Check | `/root/url-shortener-new` |
| `DB_USER` | ✅ YES | ✅ Set | `postgres` |
| `DB_PASSWORD` | ✅ YES | ✅ Set | (your secure password) |
| `DB_NAME` | ✅ YES | ✅ Set | `shortener` |
| `FIREBASE_PROJECT_ID` | ✅ YES | ✅ Set | `urlshortner-138d6` |
| `VITE_API_URL` | ✅ YES | ✅ Set | `https://api.shrten.link` |
| `VITE_FIREBASE_CONFIG` | ✅ YES | ✅ Set | (JSON config) |
| `SLACK_WEBHOOK_URL` | ⚠️ NO | Optional | Only if using Slack |

**Critical:** Verify `DO_SSH_KEY` is set correctly (this is the blocker).

## 🔐 SSH Key Security Best Practices

- Never commit private keys to Git
- Keep private keys secure locally
- Rotate keys periodically
- Use ed25519 (more secure than RSA)
- Use passphrase-protected keys if possible (but for CI/CD, use passphrases in CI/CD without passphrase)

## ⚡ Quick Verification

After setting up the secret, trigger a deployment:

1. Go to: https://github.com/kaveenpsnd/shortenlinks/actions
2. Select: "Deploy to DigitalOcean" workflow
3. Click: "Run workflow" → "Run workflow"
4. Watch the logs for:
   - ✅ "SSH Action succeeded" (or similar)
   - ✅ "Backend is healthy"
   - ✅ "Frontend is healthy"

If you see these messages, deployment is working!

## 🆘 Troubleshooting

### Error: "No such file or directory"
- The SSH key secret is missing or malformed
- Solution: Re-copy the entire private key content (including BEGIN/END lines)

### Error: "Permission denied"
- SSH key exists but isn't in droplet's authorized_keys
- Solution: Add public key to `~/.ssh/authorized_keys` on droplet

### Error: "Connection refused"
- Can't reach droplet at IP address
- Solution: Verify DO_DROPLET_IP is correct (should be 64.227.166.226)

### Error: "timeout waiting for SSH"
- Network issue or droplet is down
- Solution: Test local SSH connection first

## Next Steps

1. ✅ Generate/verify SSH key locally
2. ✅ Add public key to droplet
3. ✅ Add `DO_SSH_KEY` secret to GitHub
4. ✅ Test locally: `ssh -i ~/.ssh/do_github_actions root@64.227.166.226`
5. ✅ Re-trigger GitHub Actions deployment
6. ✅ Monitor logs for success

