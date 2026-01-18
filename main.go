package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/suiten/suiten-server/internal/db"
	"github.com/suiten/suiten-server/internal/handler"
	"github.com/suiten/suiten-server/internal/middleware"
)

func main() {
	// Initialize Echo
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.RequestID())
	e.Use(echomiddleware.CORS())

	// Rate limiting
	e.Use(echomiddleware.RateLimiter(echomiddleware.NewRateLimiterMemoryStore(100)))

	// Connect to database
	dbConfig := db.DefaultConfig()
	database, err := db.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Connected to database")

	// Initialize handlers
	authHandler := handler.NewAuthHandler(database)
	syncHandler := handler.NewSyncHandler(database)
	analyticsHandler := handler.NewAnalyticsHandler(database)
	profileHandler := handler.NewProfileHandler(database)

	// Health check endpoint
	e.GET("/health", healthCheck)

	// API v1 routes
	v1 := e.Group("/v1")

	// Public routes (no auth required)
	auth := v1.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.Refresh)

	// Protected routes (auth required)
	jwtConfig := middleware.DefaultJWTConfig()
	protected := v1.Group("")
	protected.Use(middleware.JWTAuth(jwtConfig))

	// Auth routes
	protected.POST("/auth/logout", authHandler.Logout)
	protected.DELETE("/auth/account", authHandler.DeleteAccount)

	// Sync routes
	protected.POST("/sync", syncHandler.Sync)
	protected.GET("/sync/status", syncHandler.SyncStatus)

	// Backup routes
	protected.POST("/backup/export", syncHandler.BackupExport)
	protected.POST("/backup/import", syncHandler.BackupImport)

	// Resource routes
	protected.GET("/books", syncHandler.GetBooks)
	protected.GET("/books/:id", syncHandler.GetBook)
	protected.GET("/books/:book_id/sessions", syncHandler.GetBookSessions)
	protected.GET("/sessions/:session_id/captures", syncHandler.GetSessionCaptures)
	protected.GET("/captures/:capture_id/ai-result", syncHandler.GetCaptureAIResult)
	protected.GET("/profile", profileHandler.GetProfile)
	protected.PUT("/profile", profileHandler.UpdateProfile)

	// Analytics routes
	protected.GET("/analytics/personal", analyticsHandler.PersonalAnalytics)
	protected.GET("/analytics/global", analyticsHandler.GlobalAnalytics)
	protected.GET("/analytics/trends", analyticsHandler.Trends)

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	go func() {
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server started on port %s", port)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// Health check handler
func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().Unix(),
	})
}
