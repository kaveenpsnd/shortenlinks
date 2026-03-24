package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	"github.com/kaveenpsnd/url-shortener/internal/core/domain"
	"github.com/kaveenpsnd/url-shortener/internal/core/ports"
	"google.golang.org/api/option"
)

// AuthMiddleware now accepts the UserRepo to sync users automatically
func AuthMiddleware(userRepo ports.UserRepository) gin.HandlerFunc {
	// Initialize Firebase (Keep credentials safe!)
	// Try reading from environment variable first, then fall back to file
	credPath := os.Getenv("FIREBASE_CREDENTIALS_PATH")
	if credPath == "" {
		credPath = "serviceAccountKey.json" // fallback for local development
	}

	opt := option.WithCredentialsFile(credPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		panic("Failed to init Firebase: " + err.Error())
	}
	authClient, err := app.Auth(context.Background())
	if err != nil {
		panic("Failed to create Auth client: " + err.Error())
	}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		idToken := strings.TrimSpace(strings.Replace(authHeader, "Bearer", "", 1))
		token, err := authClient.VerifyIDToken(context.Background(), idToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// --- SYNC USER TO POSTGRES ---
		// Extract email from Firebase token claims
		email, _ := token.Claims["email"].(string)

		user := domain.User{
			ID:        token.UID,
			Email:     email,
			Role:      "user", // Default role
			CreatedAt: time.Now().UTC(),
		}

		// Save/Update user in DB
		// We use a separate context so DB operations don't fail if the request cancels
		_ = userRepo.Upsert(context.Background(), user)
		savedUser, err := userRepo.GetByID(context.Background(), token.UID)
		if err == nil {
			c.Set("role", savedUser.Role) // Set 'admin' or 'user'
		} else {
			c.Set("role", "user") // Fallback
		}

		c.Set("user_id", token.UID)
		c.Next()
	}
}

// OptionalAuthMiddleware identifies the user if a token is present, but doesn't block if missing.
func OptionalAuthMiddleware(userRepo ports.UserRepository) gin.HandlerFunc {
	// Re-using the same initialization logic (in a real app, inject the auth client)
	// Try reading from environment variable first, then fall back to file
	credPath := os.Getenv("FIREBASE_CREDENTIALS_PATH")
	if credPath == "" {
		credPath = "serviceAccountKey.json" // fallback for local development
	}

	opt := option.WithCredentialsFile(credPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		panic("Failed to init Firebase: " + err.Error())
	}
	authClient, err := app.Auth(context.Background())
	if err != nil {
		panic("Failed to create Auth client: " + err.Error())
	}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token? Proceed as anonymous.
			c.Next()
			return
		}

		idToken := strings.TrimSpace(strings.Replace(authHeader, "Bearer", "", 1))
		token, err := authClient.VerifyIDToken(context.Background(), idToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// --- SYNC USER TO POSTGRES ---
		email, _ := token.Claims["email"].(string)
		user := domain.User{
			ID:        token.UID,
			Email:     email,
			Role:      "user",
			CreatedAt: time.Now().UTC(),
		}

		_ = userRepo.Upsert(context.Background(), user)
		savedUser, err := userRepo.GetByID(context.Background(), token.UID)
		if err == nil {
			c.Set("role", savedUser.Role)
		} else {
			c.Set("role", "user")
		}

		c.Set("user_id", token.UID)
		c.Next()
	}
}

// AdminOnly ensures the authenticated user has the 'admin' role
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get role from context (set by AuthMiddleware)
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// 2. Check if role is admin
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
