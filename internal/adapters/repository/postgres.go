package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // The driver (we use _ because we don't call it directly)
	"github.com/kaveenpsnd/url-shortener/internal/core/domain"
)

// PostgresRepository implements the LinkRepository interface
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository is a constructor
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Save inserts the link into the database
func (r *PostgresRepository) Save(ctx context.Context, link domain.Link) error {
	query := `
		INSERT INTO short_links (id, original_url, short_code, created_at, expires_at, user_id)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	log.Printf("[REPO-SAVE] Code: %s, ExpiresAt: %v", link.ShortCode, link.ExpiresAt)
	_, err := r.db.ExecContext(ctx, query, link.ID, link.OriginalURL, link.ShortCode, link.CreatedAt, link.ExpiresAt, link.UserID)
	if err != nil {
		log.Printf("[REPO-SAVE] ERROR: %v", err)
	}
	return err
}

// IncrementClicks adds +1 to the click count for a specific short code
func (r *PostgresRepository) IncrementClicks(ctx context.Context, code string) error {
	query := `UPDATE short_links SET clicks = clicks + 1 WHERE short_code = $1`

	_, err := r.db.ExecContext(ctx, query, code)
	if err != nil {
		return err
	}
	return nil
}
func (r *PostgresRepository) FindByUserID(ctx context.Context, userID string) ([]domain.Link, error) {
	query := `
        SELECT id, original_url, short_code, created_at, expires_at, clicks, user_id 
        FROM short_links 
        WHERE user_id = $1 
        ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []domain.Link
	for rows.Next() {
		var link domain.Link
		// Scan the columns into the struct
		if err := rows.Scan(&link.ID, &link.OriginalURL, &link.ShortCode, &link.CreatedAt, &link.ExpiresAt, &link.Clicks, &link.UserID); err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, nil
}
func (r *PostgresRepository) UpdateExpiration(ctx context.Context, shortCode string, newExp time.Time) error {
	query := `UPDATE short_links SET expires_at = $1 WHERE short_code = $2`
	result, err := r.db.ExecContext(ctx, query, newExp, shortCode)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("link not found")
	}
	return nil
}

// FindByCode fetches a link by its short string (e.g. "h7K9")
func (r *PostgresRepository) FindByCode(ctx context.Context, code string) (domain.Link, error) {
	query := `
		SELECT id, original_url, short_code, created_at, expires_at, clicks, user_id 
		FROM short_links 
		WHERE short_code = $1
	`

	var link domain.Link

	// We use QueryRowContext because we expect exactly ONE result
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&link.ID,
		&link.OriginalURL,
		&link.ShortCode,
		&link.CreatedAt,
		&link.ExpiresAt,
		&link.Clicks,
		&link.UserID,
	)

	log.Printf("[REPO-FIND] Code: %s, Retrieved ExpiresAt: %v", code, link.ExpiresAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Link{}, errors.New("link not found")
		}
		log.Printf("[REPO-FIND] ERROR: %v", err)
		return domain.Link{}, err
	}

	return link, nil
}
