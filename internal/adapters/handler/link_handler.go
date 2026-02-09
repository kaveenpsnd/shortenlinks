package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kaveenpsnd/url-shortener/internal/core/service"
	"github.com/skip2/go-qrcode"
)

type LinkHandler struct {
	svc *service.LinkService
}

// NewLinkHandler is the constructor
func NewLinkHandler(svc *service.LinkService) *LinkHandler {
	return &LinkHandler{svc: svc}
}

// Request Struct
type CreateLinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required"`
	ExpiresIn   int    `json:"expires_in_minutes"` // Optional field
}

// Shorten creates a new short link, optionally attaching a user ID if logged in
func (h *LinkHandler) Shorten(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[HANDLER-SHORTEN] PANIC: %v", r)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
	}()

	var req CreateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Check for User ID (Authenticated Request)
	// The AuthMiddleware sets "user_id" in the context if the token is valid.
	var userID *string
	if val, exists := c.Get("user_id"); exists {
		// Convert the interface{} to string safely
		if uidStr, ok := val.(string); ok {
			userID = &uidStr
		}
	}

	log.Printf("[HANDLER-SHORTEN] URL: %s, Expires In: %d min, UserID: %v", req.OriginalURL, req.ExpiresIn, userID)

	// 2. Pass userID to Service
	link, err := h.svc.Shorten(c.Request.Context(), req.OriginalURL, req.ExpiresIn, userID)
	if err != nil {
		log.Printf("[HANDLER-SHORTEN] ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to shorten URL"})
		return
	}

	log.Printf("[HANDLER-SHORTEN] SUCCESS: Code=%s, UserID=%v", link.ShortCode, link.UserID)

	c.JSON(http.StatusCreated, gin.H{
		"short_code": link.ShortCode,
		"short_url":  "https://shrten.link/" + link.ShortCode,
		"expires_at": link.ExpiresAt,
		"user_id":    link.UserID, // Useful for verifying the user was attached
	})
}

// Redirect handles the GET /:code request
func (h *LinkHandler) Redirect(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[HANDLER-REDIRECT] PANIC: %v", r)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
	}()

	code := c.Param("code")

	// Call the Service to find the original URL
	originalURL, err := h.svc.Resolve(c.Request.Context(), code)
	if err != nil {
		// Differentiate between "expired" and "not found" if your service returns specific errors,
		// otherwise 404 is generally safe.
		log.Printf("[HANDLER-REDIRECT] ERROR: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found or expired"})
		return
	}

	// Redirect using 302 Found (Temporary) so analytics keeps working
	c.Redirect(http.StatusFound, originalURL)
}
func (h *LinkHandler) GetMyLinks(c *gin.Context) {
	// 1. Get User ID from Context (Set by Middleware)
	val, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := val.(string)
	log.Printf("[GET-MY-LINKS] Fetching links for user_id: %s", userID)

	// 2. Call Service
	links, err := h.svc.GetUserLinks(c.Request.Context(), userID)
	if err != nil {
		log.Printf("[GET-MY-LINKS] ERROR fetching links: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch links"})
		return
	}

	log.Printf("[GET-MY-LINKS] Found %d links for user %s", len(links), userID)

	// 3. Return Data
	c.JSON(http.StatusOK, gin.H{
		"count": len(links),
		"links": links,
	})
}

// GetStats returns details about a specific link (clicks, expiration, etc.)
func (h *LinkHandler) GetStats(c *gin.Context) {
	code := c.Param("code")

	link, err := h.svc.GetLinkInfo(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"short_code":   link.ShortCode,
		"original_url": link.OriginalURL,
		"created_at":   link.CreatedAt,
		"expires_at":   link.ExpiresAt,
		"clicks":       link.Clicks,
		"user_id":      link.UserID, // Now visible in stats too
	})
}
func (h *LinkHandler) GetQRCode(c *gin.Context) {
	code := c.Param("code")

	// 1. Construct the Full URL
	// Production backend URL
	shortURL := "https://shrten.link/" + code

	// 2. Generate QR Code
	// qrcode.Encode(content, level, size_in_pixels)
	// LevelMedium is a good balance. 256 is a standard size (256x256 pixels).
	png, err := qrcode.Encode(shortURL, qrcode.Medium, 256)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate QR code"})
		return
	}

	// 3. Return the Image
	// We tell the browser: "This is an image, not text!"
	c.Header("Content-Type", "image/png")
	c.Header("Content-Length", fmt.Sprintf("%d", len(png)))

	// Write the bytes directly
	c.Writer.Write(png)
}
