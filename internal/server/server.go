// Package server builds the HTTP server (middleware and routes).
package server

import (
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"

	"github.com/suiten/suiten-server/internal/handler"
	"github.com/suiten/suiten-server/internal/middleware"
)

// New returns an Echo instance with all middleware and routes registered.
func New(database *sqlx.DB) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.RequestID())
	e.Use(echomiddleware.CORS())

	// Rate limiting
	e.Use(echomiddleware.RateLimiter(echomiddleware.NewRateLimiterMemoryStore(100)))

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

	return e
}

// Health check handler
func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().Unix(),
	})
}
