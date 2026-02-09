package ports

import (
	"context"
	"time"

	"github.com/kaveenpsnd/url-shortener/internal/core/domain"
)

// LinkRepository defines how we interact with the database.
// Any database implementation (Postgres, Mongo, Redis) must follow this rule.
type LinkRepository interface {
	// Save takes a Link and stores it. Returns error if it fails.
	Save(ctx context.Context, link domain.Link) error

	// FindByCode takes a short code (e.g. "h7K9") and returns the full Link.
	FindByCode(ctx context.Context, code string) (domain.Link, error)
	FindByUserID(ctx context.Context, userID string) ([]domain.Link, error)
	IncrementClicks(ctx context.Context, code string) error
	UpdateExpiration(ctx context.Context, shortCode string, newExp time.Time) error
}
