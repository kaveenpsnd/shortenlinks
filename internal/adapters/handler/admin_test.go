package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kaveenpsnd/url-shortener/internal/core/domain"
	"github.com/kaveenpsnd/url-shortener/internal/core/service"
	"github.com/kaveenpsnd/url-shortener/pkg/snowflake"
)

// ==================== Mock UserRepo ====================

type mockUserRepo struct {
	users map[string]domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]domain.User)}
}

func (m *mockUserRepo) Upsert(ctx context.Context, user domain.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (domain.User, error) {
	u, ok := m.users[id]
	if !ok {
		return domain.User{}, nil
	}
	return u, nil
}

func (m *mockUserRepo) GetAll(ctx context.Context) ([]domain.User, error) {
	var all []domain.User
	for _, u := range m.users {
		all = append(all, u)
	}
	return all, nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id string) error {
	if _, ok := m.users[id]; !ok {
		return nil
	}
	delete(m.users, id)
	return nil
}

// ==================== Helper ====================

func setupAdminHandler(t *testing.T) (*AdminHandler, *mockUserRepo, *service.LinkService) {
	t.Helper()
	userRepo := newMockUserRepo()
	linkRepo := newMockLinkRepo()
	cache := newMockCache()
	node, _ := snowflake.NewNode(1)
	svc := service.NewLinkService(linkRepo, cache, node)
	ah := NewAdminHandler(userRepo, svc)
	return ah, userRepo, svc
}

// ==================== Tests: GetAllUsers ====================

func TestGetAllUsersHandler_Success(t *testing.T) {
	ah, userRepo, _ := setupAdminHandler(t)

	userRepo.users["u1"] = domain.User{ID: "u1", Email: "a@b.com", Role: "user"}
	userRepo.users["u2"] = domain.User{ID: "u2", Email: "c@d.com", Role: "admin"}

	router := gin.New()
	router.GET("/api/admin/users", ah.GetAllUsers)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/admin/users", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp) != 2 {
		t.Errorf("Expected 2 users, got %d", len(resp))
	}
}

func TestGetAllUsersHandler_Empty(t *testing.T) {
	ah, _, _ := setupAdminHandler(t)

	router := gin.New()
	router.GET("/api/admin/users", ah.GetAllUsers)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/admin/users", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// ==================== Tests: UpdateLinkExpiry ====================

func TestUpdateLinkExpiryHandler_Success(t *testing.T) {
	ah, _, svc := setupAdminHandler(t)
	ctx := context.Background()

	link, _ := svc.Shorten(ctx, "https://admin-test.com", 0, nil)

	router := gin.New()
	router.PUT("/api/admin/links/:code/expiry", ah.UpdateLinkExpiry)

	body, _ := json.Marshal(map[string]interface{}{
		"new_minutes": 60,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/admin/links/"+link.ShortCode+"/expiry", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestUpdateLinkExpiryHandler_BadRequest(t *testing.T) {
	ah, _, _ := setupAdminHandler(t)

	router := gin.New()
	router.PUT("/api/admin/links/:code/expiry", ah.UpdateLinkExpiry)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/admin/links/abc/expiry", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// ==================== Tests: DeleteUser ====================

func TestDeleteUserHandler_Success(t *testing.T) {
	ah, userRepo, _ := setupAdminHandler(t)

	userRepo.users["delete-me"] = domain.User{ID: "delete-me", Email: "del@test.com", Role: "user"}

	router := gin.New()
	router.DELETE("/api/admin/users/:id", ah.DeleteUser)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/admin/users/delete-me", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// ==================== Tests: GetUserLinks (Admin) ====================

func TestGetUserLinksAdminHandler_Success(t *testing.T) {
	ah, _, svc := setupAdminHandler(t)
	ctx := context.Background()

	userID := "admin-view-user"
	svc.Shorten(ctx, "https://link1.com", 0, &userID)
	svc.Shorten(ctx, "https://link2.com", 0, &userID)

	router := gin.New()
	router.GET("/api/admin/users/:id/links", ah.GetUserLinks)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/admin/users/admin-view-user/links", nil)
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
