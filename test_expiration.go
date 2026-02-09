package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	_ = godotenv.Load()

	// Connect to DB
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("DB unreachable:", err)
	}

	ctx := context.Background()

	// Insert a link that expires in 1 second
	now := time.Now().UTC()
	expiresAt := now.Add(1 * time.Second)

	query := `INSERT INTO short_links (id, original_url, short_code, created_at, expires_at)
	          VALUES ($1, $2, $3, $4, $5)`

	_, err = db.ExecContext(ctx, query,
		9999999,
		"https://www.google.com",
		"testcode99",
		now,
		&expiresAt,
	)
	if err != nil {
		log.Fatal("Insert failed:", err)
	}

	fmt.Println("✓ Link inserted successfully")
	fmt.Printf("  - Short Code: testcode99\n")
	fmt.Printf("  - Created At: %v\n", now)
	fmt.Printf("  - Expires At: %v\n", expiresAt)

	// Query it back
	selectQuery := `SELECT id, original_url, short_code, created_at, expires_at FROM short_links WHERE short_code = $1`
	var link struct {
		id, originalURL, shortCode string
		createdAt                  time.Time
		expiresAt                  *time.Time
	}

	err = db.QueryRowContext(ctx, selectQuery, "testcode99").Scan(
		&link.id, &link.originalURL, &link.shortCode, &link.createdAt, &link.expiresAt,
	)
	if err != nil {
		log.Fatal("Select failed:", err)
	}

	fmt.Println("\n✓ Link retrieved from database:")
	fmt.Printf("  - Short Code: %s\n", link.shortCode)
	fmt.Printf("  - Original URL: %s\n", link.originalURL)
	fmt.Printf("  - Created At: %v\n", link.createdAt)
	fmt.Printf("  - Expires At: %v\n", link.expiresAt)

	// Check if it's expired NOW
	now2 := time.Now().UTC()
	if link.expiresAt != nil {
		isExpired := now2.After(*link.expiresAt)
		fmt.Printf("\n✓ Expiration check (NOW):\n")
		fmt.Printf("  - Now: %v\n", now2)
		fmt.Printf("  - Expires At: %v\n", link.expiresAt)
		fmt.Printf("  - Is Expired: %v\n", isExpired)
	}

	// Wait 2 seconds and check again
	fmt.Println("\n⏳ Waiting 2 seconds...")
	time.Sleep(2 * time.Second)

	now3 := time.Now().UTC()
	if link.expiresAt != nil {
		isExpired := now3.After(*link.expiresAt)
		fmt.Printf("\n✓ Expiration check (AFTER 2 seconds):\n")
		fmt.Printf("  - Now: %v\n", now3)
		fmt.Printf("  - Expires At: %v\n", link.expiresAt)
		fmt.Printf("  - Is Expired: %v\n", isExpired)
	}
}
