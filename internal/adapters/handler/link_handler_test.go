package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kaveenpsnd/url-shortener/internal/core/domain"
	"github.com/kaveenpsnd/url-shortener/internal/core/service"
	"github.com/kaveenpsnd/url-shortener/pkg/snowflake"
)

// ==================== Mocks ====================

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
	return nil
}

func (m *mockLinkRepo) UpdateExpiration(ctx context.Context, code string, exp time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if link, ok := m.links[code]; ok {
		link.ExpiresAt = &exp
		m.links[code] = link
		return nil
	}
	return errors.New("link not found")
}

type mockCache struct {
	mu   sync.Mutex
	data map[string]string
}

func newMockCache() *mockCache {
	return &mockCache{data: make(map[string]string)}
}

func (m *mockCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

func (m *mockCache) Get(ctx context.Context, key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.data[key], nil
}

func (m *mockCache) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

// ==================== Helpers ====================

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestHandler(t *testing.T) (*LinkHandler, *service.LinkService) {
	t.Helper()
	repo := newMockLinkRepo()
	cache := newMockCache()
	node, _ := snowflake.NewNode(1)
	svc := service.NewLinkService(repo, cache, node)
	h := NewLinkHandler(svc)
	return h, svc
}

// ==================== Tests: Shorten Handler ====================

func TestShortenHandler_Success(t *testing.T) {
	h, _ := setupTestHandler(t)

	router := gin.New()
	router.POST("/api/shorten", h.Shorten)

	body, _ := json.Marshal(map[string]interface{}{
		"original_url": "https://example.com",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["short_code"] == nil || resp["short_code"] == "" {
		t.Error("Response should contain short_code")
	}
}

func TestShortenHandler_WithExpiration(t *testing.T) {
	h, _ := setupTestHandler(t)

	router := gin.New()
	router.POST("/api/shorten", h.Shorten)

	body, _ := json.Marshal(map[string]interface{}{
		"original_url":       "https://example.com",
		"expires_in_minutes": 30,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["expires_at"] == nil {
		t.Error("Response should contain expires_at")
	}
}

func TestShortenHandler_WithUserID(t *testing.T) {
	h, _ := setupTestHandler(t)

	router := gin.New()
	// Simulate auth middleware setting user_id
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user-123")
		c.Next()
	})
	router.POST("/api/shorten", h.Shorten)

	body, _ := json.Marshal(map[string]interface{}{
		"original_url": "https://example.com",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["user_id"] == nil {
		t.Error("Response should contain user_id for authenticated user")
	}
}

func TestShortenHandler_MissingURL(t *testing.T) {
	h, _ := setupTestHandler(t)

	router := gin.New()
	router.POST("/api/shorten", h.Shorten)

	body, _ := json.Marshal(map[string]interface{}{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/shorten", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestShortenHandler_InvalidJSON(t *testing.T) {
	h, _ := setupTestHandler(t)

	router := gin.New()
	router.POST("/api/shorten", h.Shorten)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/shorten", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// ==================== Tests: Redirect Handler ====================

func TestRedirectHandler_Success(t *testing.T) {
	h, svc := setupTestHandler(t)
	ctx := context.Background()

	// Create a link
	link, _ := svc.Shorten(ctx, "https://redirect-target.com", 0, nil)

	router := gin.New()
	router.GET("/:code", h.Redirect)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/"+link.ShortCode, nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "https://redirect-target.com" {
		t.Errorf("Location header = %q, want %q", location, "https://redirect-target.com")
	}
}

func TestRedirectHandler_NotFound(t *testing.T) {
	h, _ := setupTestHandler(t)

	router := gin.New()
	router.GET("/:code", h.Redirect)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/nonexistent", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

// ==================== Tests: GetMyLinks Handler ====================

func TestGetMyLinksHandler_Success(t *testing.T) {
	h, svc := setupTestHandler(t)
	ctx := context.Background()

	userID := "user-links-test"
	svc.Shorten(ctx, "https://link1.com", 0, &userID)
	svc.Shorten(ctx, "https://link2.com", 0, &userID)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	router.GET("/api/user/links", h.GetMyLinks)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/user/links", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	count := int(resp["count"].(float64))
	if count != 2 {
		t.Errorf("Expected 2 links, got %d", count)
	}
}

func TestGetMyLinksHandler_NoUserID(t *testing.T) {
	h, _ := setupTestHandler(t)

	router := gin.New()
	router.GET("/api/user/links", h.GetMyLinks)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/user/links", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// ==================== Tests: GetStats Handler ====================

func TestGetStatsHandler_Success(t *testing.T) {
	h, svc := setupTestHandler(t)
	ctx := context.Background()

	link, _ := svc.Shorten(ctx, "https://stats-test.com", 0, nil)

	router := gin.New()
	router.GET("/api/links/:code/stats", h.GetStats)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/links/"+link.ShortCode+"/stats", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["short_code"] != link.ShortCode {
		t.Errorf("short_code = %v, want %v", resp["short_code"], link.ShortCode)
	}
	if resp["original_url"] != "https://stats-test.com" {
		t.Errorf("original_url = %v, want %v", resp["original_url"], "https://stats-test.com")
	}
}

func TestGetStatsHandler_NotFound(t *testing.T) {
	h, _ := setupTestHandler(t)

	router := gin.New()
	router.GET("/api/links/:code/stats", h.GetStats)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/links/nonexistent/stats", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
