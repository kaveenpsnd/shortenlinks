# Codebase Review: Critical Issues Found

## 🔴 CRITICAL ISSUE #1: Redis Connection Mismatch

**Location:** `cmd/api/main.go` line 72  
**Problem:** Backend code reads wrong environment variables for Redis

**Current Code:**
```go
redisCache := cache.NewRedisCache(os.Getenv("REDIS_ADDR"), os.Getenv("REDIS_PASSWORD"))
```

**Docker Compose Sets:**
```yaml
REDIS_HOST: redis
REDIS_PORT: 6379
```

**Issue:** Backend expects `REDIS_ADDR` (single string like "redis:6379") but Docker Compose sets `REDIS_HOST` and `REDIS_PORT` separately. Redis connection will fail with empty address.

**Impact:** 🔴 **DEPLOYMENT WILL FAIL** - Backend cannot connect to Redis cache
- Services will crash
- Application won't start
- Health check will fail

**Fix Required:** Update main.go to construct REDIS_ADDR from REDIS_HOST and REDIS_PORT

---

## 🔴 CRITICAL ISSUE #2: Firebase serviceAccountKey.json Path

**Location:** `internal/core/middleware/auth.go` line 16  
**Problem:** Firebase auth looks for credentials file at relative path

**Current Code:**
```go
opt := option.WithCredentialsFile("serviceAccountKey.json")
```

**Issue:** Relative path won't work when running in Docker container. File needs to be:
1. Mounted as a volume, OR
2. Path needs to be absolute, OR
3. Use environment variable to pass credentials

**Impact:** 🔴 **AUTHENTICATION WILL FAIL**
- User login won't work
- Protected routes will crash
- Firebase token verification will fail

**Fix Required:** Update to read from absolute path or use FIREBASE_SERVICE_ACCOUNT environment variable

---

## 🟡 ISSUE #3: Redis Password Not Configured

**Location:** Docker Compose vs. Redis Connection  
**Problem:** Redis runs without authentication, but code expects password

**Current State:**
- Redis started with no password (default: empty)
- Go code tries to fetch `REDIS_PASSWORD` env var (not set)
- Passes nil/empty password to Redis client

**Impact:** 🟢 **Might work** - Redis client handles empty password gracefully, but not secure

**Recommendation:** Either set Redis password or explicitly handle empty password

---

## ✅ VERIFIED WORKING

- ✅ Health endpoint `/health` added correctly
- ✅ Go 1.25 version correct
- ✅ All Dockerfiles multi-stage builds optimized
- ✅ Database migrations valid (3 files)
- ✅ Frontend React structure correct
- ✅ All dependencies declared in go.mod
- ✅ CORS configured for cross-origin requests

---

## 🛠️ REQUIRED FIXES (Priority Order)

### Fix #1: Redis Connection (CRITICAL)
File: `cmd/api/main.go` line 72

Replace:
```go
redisCache := cache.NewRedisCache(os.Getenv("REDIS_ADDR"), os.Getenv("REDIS_PASSWORD"))
```

With:
```go
// Get Redis connection details from environment
redisHost := os.Getenv("REDIS_HOST")
redisPort := os.Getenv("REDIS_PORT")
if redisHost == "" {
	redisHost = "localhost"
}
if redisPort == "" {
	redisPort = "6379"
}
redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)
redisPassword := os.Getenv("REDIS_PASSWORD")

redisCache := cache.NewRedisCache(redisAddr, redisPassword)
```

### Fix #2: Firebase Credentials Path (CRITICAL)
File: `internal/core/middleware/auth.go` line 16

**Option A: Use Environment Variable (Recommended)**
```go
credPath := os.Getenv("FIREBASE_CREDENTIALS_PATH")
if credPath == "" {
	credPath = "serviceAccountKey.json" // fallback for local dev
}
opt := option.WithCredentialsFile(credPath)
```

Update docker-compose.prod.yml to add:
```yaml
FIREBASE_CREDENTIALS_PATH: /app/serviceAccountKey.json
```

Mount the file:
```yaml
volumes:
  - ./serviceAccountKey.json:/app/serviceAccountKey.json:ro
```

**Option B: Read from Environment JSON (More Secure)**
```go
credJson := os.Getenv("FIREBASE_SERVICE_ACCOUNT_JSON")
if credJson == "" {
	// Fallback to file
	opt := option.WithCredentialsFile("serviceAccountKey.json")
} else {
	opt := option.WithCredentialsJSON([]byte(credJson))
}
```

---

## 📋 Full Summary

| Issue | Severity | Status | Impact |
|-------|----------|--------|--------|
| Redis connection mismatch | 🔴 CRITICAL | ❌ NOT FIXED | Backend crash |
| Firebase credentials path | 🔴 CRITICAL | ❌ NOT FIXED | Auth failure |
| Redis password handling | 🟡 MEDIUM | ⚠️ WORKAROUND | Security risk |
| All other checks | ✅ PASS | ✅ OK | No impact |

---

## ⏸️ Deployment Status

**Status:** ❌ **NOT READY FOR DEPLOYMENT**

Reason: Two critical issues will cause immediate failure when Docker services try to start.

**Next Action:** Apply the two fixes above before re-running the GitHub Actions workflow.
