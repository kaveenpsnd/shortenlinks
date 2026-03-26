package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/kaveenpsnd/url-shortener/internal/core/domain"
	"github.com/kaveenpsnd/url-shortener/pkg/snowflake"
)

// ==================== Mock Repository ====================

type mockLinkRepo struct {
	mu    sync.Mutex
	links map[string]domain.Link
}

func newMockLinkRepo() *mockLinkRepo {
	return &mockLinkRepo{links: make(map[string]domain.Link)}
}

func (m *mockLinkRepo) Save(ctx context.Context, link domain.Link) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.links[link.ShortCode] = link
	return nil
}

func (m *mockLinkRepo) FindByCode(ctx context.Context, code string) (domain.Link, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	link, ok := m.links[code]
	if !ok {
		return domain.Link{}, errors.New("link not found")
	}
	return link, nil
}

func (m *mockLinkRepo) FindByUserID(ctx context.Context, userID string) ([]domain.Link, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []domain.Link
	for _, link := range m.links {
		if link.UserID != nil && *link.UserID == userID {
			result = append(result, link)
		}
	}
	return result, nil
}

func (m *mockLinkRepo) IncrementClicks(ctx context.Context, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if link, ok := m.links[code]; ok {
		link.Clicks++
		m.links[code] = link
		return nil
	}
	return errors.New("link not found")
}

func (m *mockLinkRepo) UpdateExpiration(ctx context.Context, shortCode string, newExp time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if link, ok := m.links[shortCode]; ok {
		link.ExpiresAt = &newExp
		m.links[shortCode] = link
		return nil
	}
	return errors.New("link not found")
}

// ==================== Mock Cache ====================

type mockCache struct {
	mu   sync.Mutex
	data map[string]string
}

func newMockCache() *mockCache {
	return &mockCache{data: make(map[string]string)}
}

func (m *mockCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

func (m *mockCache) Get(ctx context.Context, key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	val, ok := m.data[key]
	if !ok {
		return "", nil
	}
	return val, nil
}

func (m *mockCache) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

// ==================== Helper ====================

func setupService(t *testing.T) (*LinkService, *mockLinkRepo, *mockCache) {
	t.Helper()
	repo := newMockLinkRepo()
	cache := newMockCache()
	node, err := snowflake.NewNode(1)
	if err != nil {
		t.Fatal(err)
	}
	svc := NewLinkService(repo, cache, node)
	return svc, repo, cache
}

// ==================== Tests: Shorten ====================

func TestShorten_Success(t *testing.T) {
	svc, repo, cache := setupService(t)
	ctx := context.Background()

	link, err := svc.Shorten(ctx, "https://example.com", 0, nil)
	if err != nil {
		t.Fatalf("Shorten failed: %v", err)
	}

	if link.OriginalURL != "https://example.com" {
		t.Errorf("OriginalURL = %q, want %q", link.OriginalURL, "https://example.com")
	}
	if link.ShortCode == "" {
		t.Error("ShortCode is empty")
	}
	if link.ExpiresAt != nil {
		t.Error("ExpiresAt should be nil for no expiration")
	}
	if link.UserID != nil {
		t.Error("UserID should be nil for anonymous")
	}

	// Verify saved to repo
	if _, err := repo.FindByCode(ctx, link.ShortCode); err != nil {
		t.Errorf("Link not found in repo: %v", err)
	}

	// Verify cached
	cached, _ := cache.Get(ctx, link.ShortCode)
	if cached != "https://example.com" {
		t.Errorf("Cache value = %q, want %q", cached, "https://example.com")
	}
}

func TestShorten_WithExpiration(t *testing.T) {
	svc, _, _ := setupService(t)
	ctx := context.Background()

	link, err := svc.Shorten(ctx, "https://example.com", 30, nil)
	if err != nil {
		t.Fatalf("Shorten failed: %v", err)
	}

	if link.ExpiresAt == nil {
		t.Fatal("ExpiresAt should not be nil")
	}

	// The expiration should be approximately 30 minutes from now
	expectedExp := time.Now().UTC().Add(30 * time.Minute)
	diff := link.ExpiresAt.Sub(expectedExp)
	if diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("ExpiresAt is too far from expected: diff=%v", diff)
	}
}

func TestShorten_WithUserID(t *testing.T) {
	svc, _, _ := setupService(t)
	ctx := context.Background()

	userID := "firebase-uid-123"
	link, err := svc.Shorten(ctx, "https://example.com", 0, &userID)
	if err != nil {
		t.Fatalf("Shorten failed: %v", err)
	}

	if link.UserID == nil {
		t.Fatal("UserID should not be nil")
	}
	if *link.UserID != userID {
		t.Errorf("UserID = %q, want %q", *link.UserID, userID)
	}
}

func TestShorten_UniqueShortCodes(t *testing.T) {
	svc, _, _ := setupService(t)
	ctx := context.Background()

	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		link, err := svc.Shorten(ctx, "https://example.com", 0, nil)
		if err != nil {
			t.Fatalf("Shorten failed at iteration %d: %v", i, err)
		}
		if codes[link.ShortCode] {
			t.Fatalf("Duplicate short code at iteration %d: %s", i, link.ShortCode)
		}
		codes[link.ShortCode] = true
	}
}

// ==================== Tests: Resolve ====================

func TestResolve_FromCache(t *testing.T) {
	svc, _, cache := setupService(t)
	ctx := context.Background()

	// Pre-populate cache
	cache.Set(ctx, "abc123", "https://cached.com", time.Hour)

	url, err := svc.Resolve(ctx, "abc123")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if url != "https://cached.com" {
		t.Errorf("Resolved URL = %q, want %q", url, "https://cached.com")
	}
}

func TestResolve_FromDB(t *testing.T) {
	svc, _, _ := setupService(t)
	ctx := context.Background()

	// Create via Shorten (saves to both repo and cache)
	link, _ := svc.Shorten(ctx, "https://fromdb.com", 0, nil)

	// Clear cache to force DB lookup
	svc.cache.Delete(ctx, link.ShortCode)

	url, err := svc.Resolve(ctx, link.ShortCode)
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if url != "https://fromdb.com" {
		t.Errorf("Resolved URL = %q, want %q", url, "https://fromdb.com")
	}
}

func TestResolve_NotFound(t *testing.T) {
	svc, _, _ := setupService(t)
	ctx := context.Background()

	_, err := svc.Resolve(ctx, "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent code")
	}
}

func TestResolve_ExpiredLink(t *testing.T) {
	svc, repo, _ := setupService(t)
	ctx := context.Background()

	// Insert an expired link directly into the repo
	expiredTime := time.Now().UTC().Add(-1 * time.Hour)
	repo.Save(ctx, domain.Link{
		ID:          999,
		OriginalURL: "https://expired.com",
		ShortCode:   "expired1",
		CreatedAt:   time.Now().UTC().Add(-2 * time.Hour),
		ExpiresAt:   &expiredTime,
	})

	_, err := svc.Resolve(ctx, "expired1")
	if err == nil {
		t.Error("Expected error for expired link")
	}
	if err != nil && err.Error() != "link has expired" {
		t.Errorf("Expected 'link has expired' error, got: %v", err)
	}
}

// ==================== Tests: GetLinkInfo ====================

func TestGetLinkInfo_Success(t *testing.T) {
	svc, _, _ := setupService(t)
	ctx := context.Background()

	created, _ := svc.Shorten(ctx, "https://info.com", 60, nil)

	link, err := svc.GetLinkInfo(ctx, created.ShortCode)
	if err != nil {
		t.Fatalf("GetLinkInfo failed: %v", err)
	}
	if link.OriginalURL != "https://info.com" {
		t.Errorf("OriginalURL = %q, want %q", link.OriginalURL, "https://info.com")
	}
}

func TestGetLinkInfo_NotFound(t *testing.T) {
	svc, _, _ := setupService(t)
	ctx := context.Background()

	_, err := svc.GetLinkInfo(ctx, "doesnotexist")
	if err == nil {
		t.Error("Expected error for non-existent link")
	}
}

// ==================== Tests: GetUserLinks ====================

func TestGetUserLinks_ReturnsUserLinks(t *testing.T) {
	svc, _, _ := setupService(t)
	ctx := context.Background()

	userID := "user-1"
	svc.Shorten(ctx, "https://link1.com", 0, &userID)
	svc.Shorten(ctx, "https://link2.com", 0, &userID)
	svc.Shorten(ctx, "https://link3.com", 0, nil) // anonymous

	links, err := svc.GetUserLinks(ctx, userID)
	if err != nil {
		t.Fatalf("GetUserLinks failed: %v", err)
	}
	if len(links) != 2 {
		t.Errorf("Expected 2 links, got %d", len(links))
	}
}

func TestGetUserLinks_NoLinks(t *testing.T) {
	svc, _, _ := setupService(t)
	ctx := context.Background()

	links, err := svc.GetUserLinks(ctx, "no-links-user")
	if err != nil {
		t.Fatalf("GetUserLinks failed: %v", err)
	}
	if len(links) != 0 {
		t.Errorf("Expected 0 links, got %d", len(links))
	}
}

// ==================== Tests: UpdateExpiration ====================

func TestUpdateExpiration_Success(t *testing.T) {
	svc, repo, _ := setupService(t)
	ctx := context.Background()

	link, _ := svc.Shorten(ctx, "https://update-exp.com", 0, nil)

	err := svc.UpdateExpiration(ctx, link.ShortCode, 120)
	if err != nil {
		t.Fatalf("UpdateExpiration failed: %v", err)
	}

	// Verify in repo
	updated, _ := repo.FindByCode(ctx, link.ShortCode)
	if updated.ExpiresAt == nil {
		t.Fatal("ExpiresAt should be set after update")
	}

	expectedExp := time.Now().UTC().Add(120 * time.Minute)
	diff := updated.ExpiresAt.Sub(expectedExp)
	if diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("Updated ExpiresAt is too far from expected: diff=%v", diff)
	}
}

func TestUpdateExpiration_NotFound(t *testing.T) {
	svc, _, _ := setupService(t)
	ctx := context.Background()

	err := svc.UpdateExpiration(ctx, "nonexistent", 60)
	if err == nil {
		t.Error("Expected error for non-existent link")
	}
}

func TestUpdateExpiration_ClearsCache(t *testing.T) {
	svc, _, cache := setupService(t)
	ctx := context.Background()

	link, _ := svc.Shorten(ctx, "https://cache-clear.com", 0, nil)

	// Verify it's in cache
	val, _ := cache.Get(ctx, link.ShortCode)
	if val == "" {
		t.Fatal("Link should be in cache before update")
	}

	svc.UpdateExpiration(ctx, link.ShortCode, 60)

	// Cache should be cleared
	val, _ = cache.Get(ctx, link.ShortCode)
	if val != "" {
		t.Error("Cache should be cleared after UpdateExpiration")
	}
}
