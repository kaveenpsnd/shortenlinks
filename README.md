# ShrtenLink — URL Shortener

A modern, full-stack URL shortening platform with analytics, QR code generation, user authentication, and admin controls. Built with **Go** and **React**, following clean hexagonal architecture principles.

**Live:** [shrten.link](https://shrten.link) &nbsp;|&nbsp; **API:** [api.shrten.link](https://api.shrten.link)

---

## Features

| Feature | Description |
|---|---|
| **URL Shortening** | Shorten any URL — works publicly or with an account |
| **Link Expiration** | Set optional expiration time in minutes |
| **Click Analytics** | Real-time click tracking with stats dashboard |
| **QR Code Generation** | Generate 256×256 PNG QR codes for any short link |
| **User Accounts** | Firebase authentication (Google OAuth + Email/Password) |
| **User Dashboard** | Manage all your links, view stats, copy/share |
| **Admin Panel** | User management, link expiry updates, user deletion |
| **Redis Caching** | Sub-millisecond redirects via Redis with smart TTL |
| **Snowflake IDs** | Distributed unique ID generation with Base62 encoding |

---

## Tech Stack

### Backend
- **Go** (Gin framework)
- **PostgreSQL 15** — persistent storage
- **Redis 7** — caching layer
- **Firebase Admin SDK** — token verification & user sync

### Frontend
- **React 19** (Vite)
- **React Router v7**
- **Firebase SDK** — authentication
- **Axios** — HTTP client
- **Lucide React** — icons
- **qrcode.react** — inline QR rendering

### Infrastructure
- **Docker** — multi-stage builds for backend & frontend
- **Kubernetes** — production orchestration
- **NGINX** — frontend serving & reverse proxy
- **Let's Encrypt** — TLS via cert-manager
- **Azure Container Registry** — image hosting

---

## Architecture

The project follows **hexagonal (ports & adapters) architecture**:

```
cmd/api/
  main.go                    ← Entry point, dependency injection, routing

internal/
  core/
    domain/                  ← Data models (Link, User)
    service/                 ← Business logic
    ports/                   ← Interfaces (repository, cache)
    middleware/              ← Auth & authorization middleware
  adapters/
    handler/                 ← HTTP handlers (Gin)
    repository/              ← PostgreSQL implementations
    cache/                   ← Redis implementation

pkg/
  base62/                    ← Base62 encoding utility
  snowflake/                 ← Distributed ID generator

frontend/
  src/
    pages/                   ← Home, Dashboard, Analytics
    components/              ← Navbar, AuthModal, ResultCard
    context/                 ← AuthContext (Firebase)
    config/                  ← API & Firebase configuration
```

---

## API Endpoints

### Public

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/shorten` | Shorten a URL (anonymous or authenticated) |
| `GET` | `/:code` | Redirect to the original URL |

### Authenticated (Bearer token required)

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/user/links` | List authenticated user's links |
| `GET` | `/api/links/:code/stats` | Get link statistics |
| `GET` | `/api/links/:code/qr` | Generate QR code (PNG) |

### Admin (Bearer token + admin role)

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/admin/users` | List all users |
| `GET` | `/api/admin/users/:id/links` | View a user's links |
| `PUT` | `/api/admin/links/:code/expiry` | Update link expiration |
| `DELETE` | `/api/admin/users/:id` | Delete user and their links |

### Example: Shorten a URL

```bash
curl -X POST https://api.shrten.link/api/shorten \
  -H "Content-Type: application/json" \
  -d '{"original_url": "https://example.com/very/long/url", "expires_in_minutes": 1440}'
```

Response:
```json
{
  "short_code": "h7K9",
  "short_url": "https://shrten.link/h7K9",
  "expires_at": "2026-03-09T15:30:00Z"
}
```

---

## Database Schema

```sql
-- Users
CREATE TABLE users (
    id TEXT PRIMARY KEY,                -- Firebase UID
    email TEXT NOT NULL UNIQUE,
    role TEXT DEFAULT 'user',           -- 'admin' or 'user'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Short Links
CREATE TABLE short_links (
    id BIGINT PRIMARY KEY,              -- Snowflake ID
    original_url TEXT NOT NULL,
    short_code VARCHAR(10) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,               -- NULL = no expiration
    clicks BIGINT DEFAULT 0,
    user_id TEXT REFERENCES users(id)
);
```

---

## Getting Started

### Prerequisites

- **Go** 1.21+
- **Node.js** 18+
- **Docker** & **Docker Compose**
- **Firebase project** with Authentication enabled

### 1. Clone the repository

```bash
git clone https://github.com/kaveenpsnd/shortenlinks.git
cd shortenlinks
```

### 2. Start PostgreSQL & Redis

```bash
docker-compose up -d
```

This starts:
- PostgreSQL on port `5433`
- Redis on port `6379`

### 3. Configure environment

Create a `.env` file in the project root:

```env
DB_HOST=127.0.0.1
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=shortener
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
ADMIN_EMAILS=your-admin@email.com
```

Place your Firebase Admin SDK credentials as `serviceAccountKey.json` in the project root (download from Firebase Console → Project Settings → Service Accounts).

### 4. Run the backend

```bash
go mod download
go run cmd/api/main.go
```

Backend starts at `http://localhost:8080`.

### 5. Run the frontend

```bash
cd frontend
npm install
npm run dev
```

Create `frontend/.env.local`:

```env
VITE_FIREBASE_API_KEY=your_api_key
VITE_FIREBASE_AUTH_DOMAIN=your_project.firebaseapp.com
VITE_FIREBASE_PROJECT_ID=your_project_id
VITE_FIREBASE_STORAGE_BUCKET=your_project.appspot.com
VITE_FIREBASE_MESSAGING_SENDER_ID=123456789
VITE_FIREBASE_APP_ID=1:123456789:web:abc123
```

Frontend starts at `http://localhost:5173`.

---

## Docker Deployment

### Build & run with Docker Compose (local)

```bash
docker-compose up -d          # Start Postgres + Redis
go run cmd/api/main.go        # Start backend
cd frontend && npm run dev    # Start frontend dev server
```

### Production builds

```bash
# Backend
docker build -t shortenlinks-backend:latest .

# Frontend
docker build -t shortenlinks-frontend:latest ./frontend
```

---

## Kubernetes Deployment

```bash
# Apply all manifests
kubectl apply -f k8s/postgres.yaml
kubectl apply -f k8s/redis.yaml
kubectl apply -f k8s/backend.yaml
kubectl apply -f k8s/frontend.yaml
kubectl apply -f k8s/ingress.yaml
kubectl apply -f k8s/cluster-issuer.yaml

# Create Firebase credentials secret
kubectl create secret generic firebase-key \
  --from-file=serviceAccountKey.json
```

---

## Authentication Flow

```
Browser → Firebase SDK (Google/Email login)
       → Firebase issues ID token
       → Frontend sends token in Authorization header
       → Backend verifies with Firebase Admin SDK
       → User auto-synced to PostgreSQL
       → Role-based access enforced (user / admin)
```

**Middleware layers:**
- `OptionalAuthMiddleware` — allows anonymous access, attaches user if token present
- `AuthMiddleware` — requires valid token (401 if missing/invalid)
- `AdminOnly` — requires `role = 'admin'` (403 if not admin)

---

## Project Structure

```
shortenlinks/
├── cmd/api/main.go              # Entry point
├── internal/
│   ├── adapters/
│   │   ├── cache/redis.go       # Redis caching
│   │   ├── handler/             # HTTP handlers
│   │   └── repository/          # PostgreSQL repos
│   └── core/
│       ├── domain/              # Link & User models
│       ├── middleware/           # Auth middleware
│       ├── ports/               # Interfaces
│       └── service/             # Business logic
├── pkg/
│   ├── base62/                  # Base62 encoding
│   └── snowflake/               # ID generation
├── frontend/                    # React SPA
│   └── src/
│       ├── pages/               # Home, Dashboard, Analytics
│       ├── components/          # Navbar, AuthModal, ResultCard
│       ├── context/             # AuthContext
│       └── config/              # API & Firebase config
├── k8s/                         # Kubernetes manifests
├── migrations/                  # SQL migrations
├── docker-compose.yml           # Local dev (Postgres + Redis)
├── dockerfile                   # Backend Docker build
└── go.mod                       # Go dependencies
```

---

## License

This project is for educational and portfolio purposes.
