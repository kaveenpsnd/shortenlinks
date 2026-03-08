package repository

import (
	"database/sql"
	"log"
)

// RunMigrations creates all required tables if they do not already exist.
// Must be called after a successful db.Ping() and before the server starts.
func RunMigrations(db *sql.DB) error {
	// users must be created first because short_links references it
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			role TEXT DEFAULT 'user',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,

		`CREATE TABLE IF NOT EXISTS short_links (
			id BIGINT PRIMARY KEY,
			original_url TEXT NOT NULL,
			short_code VARCHAR(10) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP,
			clicks BIGINT DEFAULT 0,
			user_id TEXT REFERENCES users(id)
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_short_code ON short_links(short_code)`,
		`CREATE INDEX IF NOT EXISTS idx_short_links_user_id ON short_links(user_id)`,
	}

	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return err
		}
	}

	log.Println("Database migrations completed successfully!")
	return nil
}
