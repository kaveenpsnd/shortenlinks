package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"

	"github.com/kaveenpsnd/url-shortener/internal/adapters/cache"
	"github.com/kaveenpsnd/url-shortener/internal/adapters/handler"
	"github.com/kaveenpsnd/url-shortener/internal/adapters/repository"
	"github.com/kaveenpsnd/url-shortener/internal/core/middleware" // <--- Import Middleware
	"github.com/kaveenpsnd/url-shortener/internal/core/service"
	"github.com/kaveenpsnd/url-shortener/pkg/snowflake"
)

func main() {
	// 1. Load .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, relying on environment variables")
	}

	// 2. Init Snowflake (ID Generator)
	node, err := snowflake.NewNode(1)
	if err != nil {
		log.Fatal("Failed to initialize Snowflake:", err)
	}

	// 3. Connect to Database (Postgres)
	// Fallback to localhost if env vars are missing (useful for local dev)
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5433"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		dbHost,
		dbPort,
		os.Getenv("DB_NAME"),
	)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Check DB connection
	if err := db.Ping(); err != nil {
		log.Fatal("Database unreachable:", err)
	}
	log.Println("Successfully connected to Database!")

	// Run migrations to ensure tables exist
	if err := repository.RunMigrations(db); err != nil {
		log.Fatal("Failed to run database migrations:", err)
	}

	// 4. Connect to Cache (Redis)
	redisCache := cache.NewRedisCache(os.Getenv("REDIS_ADDR"), os.Getenv("REDIS_PASSWORD"))
	log.Println("Successfully connected to Redis!")

	// 5. Dependency Injection
	// Initialize Repositories
	// Note: Ensure your constructor names match exactly what is in your files
	linkRepo := repository.NewPostgresRepository(db)
	userRepo := repository.NewPostgresUserRepository(db) // <--- New User Repo

	// Initialize Service
	svc := service.NewLinkService(linkRepo, redisCache, node)

	// Initialize Handler
	h := handler.NewLinkHandler(svc)
	adminHandler := handler.NewAdminHandler(userRepo, svc)

	// 6. Setup Router
	router := gin.Default()

	// CORS configuration - allow dev frontend origins and required methods
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:5174", "http://localhost:5175", "http://localhost:5176", "http://localhost:3000", "http://13.71.55.98", "http://20.204.185.86", "https://shrtner.link", "https://www.shrtner.link", "https://shrten.link", "https://www.shrten.link"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// --- PUBLIC ROUTES (No Login Required) ---
	router.GET("/:code", h.Redirect)

	// --- OPTIONAL AUTH ROUTES ---
	// Allows anonymous shortening, but tracks UserID if token provided
	router.POST("/api/shorten", middleware.OptionalAuthMiddleware(userRepo), h.Shorten)

	// --- PROTECTED ROUTES (Require Firebase Login) ---
	authGroup := router.Group("/api")

	// Inject the middleware with userRepo so it can sync users to DB
	authGroup.Use(middleware.AuthMiddleware(userRepo))
	{
		// These now require "Authorization: Bearer <TOKEN>" header
		authGroup.GET("/user/links", h.GetMyLinks)
		authGroup.GET("/links/:code/stats", h.GetStats)
		authGroup.GET("/links/:code/qr", h.GetQRCode)
	}

	// --- ADMIN ROUTES ---
	adminGroup := router.Group("/api/admin")
	// Use BOTH middlewares: 1. Must be Logged In, 2. Must be Admin
	adminGroup.Use(middleware.AuthMiddleware(userRepo), middleware.AdminOnly())
	{
		adminGroup.GET("/users", adminHandler.GetAllUsers)
		adminGroup.PUT("/links/:code/expiry", adminHandler.UpdateLinkExpiry)
		adminGroup.DELETE("/users/:id", adminHandler.DeleteUser)      // Delete User
		adminGroup.GET("/users/:id/links", adminHandler.GetUserLinks) // View User Links
	}

	log.Println("Server starting on :8080...")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
