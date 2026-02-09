CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,                -- Firebase UID
    email TEXT NOT NULL UNIQUE,         -- User's email
    role TEXT DEFAULT 'user',           -- 'admin' or 'user'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for faster email lookups
CREATE INDEX idx_users_email ON users(email);
