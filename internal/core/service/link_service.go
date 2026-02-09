package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/kaveenpsnd/url-shortener/internal/core/domain"
	"github.com/kaveenpsnd/url-shortener/internal/core/ports"
	"github.com/kaveenpsnd/url-shortener/pkg/base62"
	"github.com/kaveenpsnd/url-shortener/pkg/snowflake"
)

type LinkService struct {
	repo  ports.LinkRepository
	cache ports.LinkCache
	node  *snowflake.Node
}

func NewLinkService(repo ports.LinkRepository, cache ports.LinkCache, node *snowflake.Node) *LinkService {
	return &LinkService{repo: repo, cache: cache, node: node}
}

// Update: Accept 'expiresIn' (minutes)
func (s *LinkService) Shorten(ctx context.Context, originalURL string, expiresInMinutes int, userID *string) (domain.Link, error) {
	id := s.node.Generate()
	code := base62.Encode(id)
	now := time.Now().UTC() // Consider using .UTC() later if timezones are messy

	link := domain.Link{
		ID:          id,
		OriginalURL: originalURL,
		ShortCode:   code,
		CreatedAt:   now,
		UserID:      userID,
	}

	// Calculate Expiration Date
	if expiresInMinutes > 0 {
		exp := now.Add(time.Duration(expiresInMinutes) * time.Minute)
		link.ExpiresAt = &exp
		log.Printf("[SHORTEN] Creating link with expiration: %v (expires in %d minutes)", exp, expiresInMinutes)
	} else {
		log.Printf("[SHORTEN] Creating link WITHOUT expiration")
	}

	// 1. Save to DB
	err := s.repo.Save(ctx, link)
	if err != nil {
		return domain.Link{}, err
	}

	// 2. Save to Redis
	ttl := 24 * time.Hour
	if expiresInMinutes > 0 {
		ttl = time.Duration(expiresInMinutes) * time.Minute
	}

	_ = s.cache.Set(ctx, code, originalURL, ttl)

	return link, nil
}

// GetLinkInfo fetches details without incrementing the click counter
func (s *LinkService) GetLinkInfo(ctx context.Context, code string) (domain.Link, error) {
	// We MUST go to the DB because Redis usually only stores the URL string, not the stats.
	link, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return domain.Link{}, err
	}
	return link, nil
}
func (s *LinkService) GetUserLinks(ctx context.Context, userID string) ([]domain.Link, error) {
	return s.repo.FindByUserID(ctx, userID)
}
func (s *LinkService) UpdateExpiration(ctx context.Context, code string, newMinutes int) error {
	// 1. Calculate new time
	newExp := time.Now().UTC().Add(time.Duration(newMinutes) * time.Minute)

	// 2. Update Database
	err := s.repo.UpdateExpiration(ctx, code, newExp)
	if err != nil {
		return err
	}

	// 3. Update Redis (We just set the new TTL)
	// We construct the key usually as just the code, or however your cache adapter handles it.
	// Assuming your cache adapter has a generic 'Set' or we add an 'UpdateTTL' method.
	// For simplicity, we will just INVALIDATE the cache so it fetches fresh from DB next time.
	_ = s.cache.Delete(ctx, code)

	return nil
}
func (s *LinkService) Resolve(ctx context.Context, code string) (string, error) {
	var originalURL string

	// 1. Try Redis
	cachedURL, _ := s.cache.Get(ctx, code)

	if cachedURL != "" {
		originalURL = cachedURL
		// Even if we hit cache, we still need to count the click!
		// We do this in the background (Async) so we don't slow down the user.
		go func() {
			_ = s.repo.IncrementClicks(context.Background(), code)
		}()
		return originalURL, nil
	}

	// 2. Try Database
	link, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return "", err
	}

	// 3. Check Expiration
	if link.ExpiresAt != nil && time.Now().UTC().After(*link.ExpiresAt) {
		return "", errors.New("link has expired")
	}

	// 4. Update Click Count (Background Async)
	go func() {
		_ = s.repo.IncrementClicks(context.Background(), code)
	}()

	// 5. Refill Redis Cache
	originalURL = link.OriginalURL
	ttl := 24 * time.Hour
	if link.ExpiresAt != nil {
		ttl = link.ExpiresAt.Sub(time.Now().UTC())
	}
	if ttl > 0 {
		_ = s.cache.Set(ctx, code, originalURL, ttl)
	}

	return originalURL, nil
}
