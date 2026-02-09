package ports

import (
	"context"
	"time"
)

// LinkCache defines how we talk to our fast memory storage.
type LinkCache interface {
	// Set saves a key (short_code) and value (original_url) for a specific time.
	Set(ctx context.Context, key string, value string, ttl time.Duration) error

	// Get retrieves the original_url.
	Get(ctx context.Context, key string) (string, error)

	// Delete removes a key from cache.
	Delete(ctx context.Context, key string) error
}
