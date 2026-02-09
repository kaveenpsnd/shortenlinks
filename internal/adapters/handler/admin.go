package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kaveenpsnd/url-shortener/internal/core/ports"
	"github.com/kaveenpsnd/url-shortener/internal/core/service"
)

type AdminHandler struct {
	userRepo    ports.UserRepository
	linkService *service.LinkService // <--- Added this dependency
}

// Update Constructor to accept both dependencies
func NewAdminHandler(uRepo ports.UserRepository, lService *service.LinkService) *AdminHandler {
	return &AdminHandler{
		userRepo:    uRepo,
		linkService: lService,
	}
}

// 1. Get All Users
func (h *AdminHandler) GetAllUsers(c *gin.Context) {
	users, err := h.userRepo.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// 2. Update Link Expiry (New Feature)
type UpdateExpiryRequest struct {
	NewMinutes int `json:"new_minutes" binding:"required"`
}

func (h *AdminHandler) UpdateLinkExpiry(c *gin.Context) {
	code := c.Param("code")
	var req UpdateExpiryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call the service to update DB and clear Cache
	err := h.linkService.UpdateExpiration(c.Request.Context(), code, req.NewMinutes)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "link not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Expiration updated successfully", "code": code})
}

// 3. Delete User
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	// Call Repository directly (simple logic)
	err := h.userRepo.Delete(c.Request.Context(), userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "user not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User and their links deleted successfully"})
}

// 4. View A Specific User's Links (Admin View)
func (h *AdminHandler) GetUserLinks(c *gin.Context) {
	userID := c.Param("id")

	// We reuse the LinkService method we made for the user!
	links, err := h.linkService.GetUserLinks(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch links"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"count":   len(links),
		"links":   links,
	})
}
