package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/suiten/suiten-server/internal/middleware"
	"github.com/suiten/suiten-server/internal/models"
	"github.com/suiten/suiten-server/internal/repository"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	db              *sqlx.DB
	userRepo        *repository.UserRepository
	deviceRepo      *repository.DeviceRepository
	tokenRepo       *repository.RefreshTokenRepository
	profileRepo     *repository.UserProfileRepository
	jwtSecret       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(db *sqlx.DB) *AuthHandler {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "development-secret-key-change-in-production"
	}
	return &AuthHandler{
		db:              db,
		userRepo:        repository.NewUserRepository(db),
		deviceRepo:      repository.NewDeviceRepository(db),
		tokenRepo:       repository.NewRefreshTokenRepository(db),
		profileRepo:     repository.NewUserProfileRepository(db),
		jwtSecret:       []byte(secret),
		accessTokenTTL:  1 * time.Hour,
		refreshTokenTTL: 30 * 24 * time.Hour,
	}
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=128"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required"`
	DeviceID   string `json:"device_id" validate:"required"`
	DeviceName string `json:"device_name"`
}

// RefreshRequest represents the token refresh request body
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Register handles user registration
func (h *AuthHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	if req.Email == "" || req.Password == "" {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Email and password are required")
	}

	if len(req.Password) < 8 {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Password must be at least 8 characters")
	}

	ctx := c.Request().Context()

	// Check if email already exists
	exists, err := h.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check email")
	}
	if exists {
		return errorResponse(c, http.StatusConflict, "AUTH_EMAIL_EXISTS", "Email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to hash password")
	}

	// Create user
	now := time.Now()
	user := &models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := h.userRepo.Create(ctx, user); err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create user")
	}

	// Create default user profile
	profile := &models.UserProfile{
		ID:                uuid.New(),
		UserID:            user.ID,
		AxisPreferences:   models.DefaultAxisPreferences(),
		ActionPreferences: models.DefaultActionPreferences(),
		UpdatedAt:         now,
	}
	if err := h.profileRepo.Create(ctx, profile); err != nil {
		// Log but don't fail registration
		c.Logger().Warnf("Failed to create user profile: %v", err)
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"user_id":    user.ID,
			"email":      user.Email,
			"created_at": user.CreatedAt.Unix(),
		},
	})
}

// Login handles user login
func (h *AuthHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	if req.Email == "" || req.Password == "" || req.DeviceID == "" {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Email, password, and device_id are required")
	}

	ctx := c.Request().Context()

	// Find user by email
	user, err := h.userRepo.FindByEmail(ctx, req.Email)
	if err == sql.ErrNoRows {
		return errorResponse(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Invalid email or password")
	}
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to find user")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return errorResponse(c, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "Invalid email or password")
	}

	// Upsert device
	now := time.Now()
	device := &models.Device{
		ID:           uuid.New(),
		UserID:       user.ID,
		DeviceID:     req.DeviceID,
		DeviceName:   &req.DeviceName,
		LastActiveAt: now,
		CreatedAt:    now,
	}
	if err := h.deviceRepo.Upsert(ctx, device); err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to register device")
	}

	// Generate tokens
	accessToken, err := h.generateAccessToken(user.ID, device.ID)
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate access token")
	}

	refreshToken, refreshTokenHash, err := h.generateRefreshToken()
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate refresh token")
	}

	// Save refresh token
	tokenRecord := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		DeviceID:  device.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: now.Add(h.refreshTokenTTL),
		CreatedAt: now,
	}
	if err := h.tokenRepo.Create(ctx, tokenRecord); err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save refresh token")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"expires_in":    int(h.accessTokenTTL.Seconds()),
			"user": map[string]interface{}{
				"user_id": user.ID,
				"email":   user.Email,
			},
		},
	})
}

// Refresh handles token refresh
func (h *AuthHandler) Refresh(c echo.Context) error {
	var req RefreshRequest
	if err := c.Bind(&req); err != nil {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	if req.RefreshToken == "" {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Refresh token is required")
	}

	ctx := c.Request().Context()

	// Hash the provided token
	tokenHash := hashToken(req.RefreshToken)

	// Find token record
	tokenRecord, err := h.tokenRepo.FindByTokenHash(ctx, tokenHash)
	if err == sql.ErrNoRows {
		return errorResponse(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "Invalid refresh token")
	}
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to find token")
	}

	// Revoke old token
	if err := h.tokenRepo.Revoke(ctx, tokenRecord.ID); err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to revoke old token")
	}

	// Generate new tokens
	now := time.Now()
	accessToken, err := h.generateAccessToken(tokenRecord.UserID, tokenRecord.DeviceID)
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate access token")
	}

	newRefreshToken, newRefreshTokenHash, err := h.generateRefreshToken()
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate refresh token")
	}

	// Save new refresh token
	newTokenRecord := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    tokenRecord.UserID,
		DeviceID:  tokenRecord.DeviceID,
		TokenHash: newRefreshTokenHash,
		ExpiresAt: now.Add(h.refreshTokenTTL),
		CreatedAt: now,
	}
	if err := h.tokenRepo.Create(ctx, newTokenRecord); err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to save refresh token")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"access_token":  accessToken,
			"refresh_token": newRefreshToken,
			"expires_in":    int(h.accessTokenTTL.Seconds()),
		},
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c echo.Context) error {
	deviceID, err := middleware.GetDeviceID(c)
	if err != nil {
		return c.NoContent(http.StatusNoContent)
	}

	ctx := c.Request().Context()

	// Revoke all tokens for this device
	if err := h.tokenRepo.RevokeAllForDevice(ctx, deviceID); err != nil {
		c.Logger().Warnf("Failed to revoke tokens: %v", err)
	}

	return c.NoContent(http.StatusNoContent)
}

// DeleteAccount handles account deletion
func (h *AuthHandler) DeleteAccount(c echo.Context) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errorResponse(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "Invalid token")
	}

	ctx := c.Request().Context()

	// Revoke all tokens
	if err := h.tokenRepo.RevokeAllForUser(ctx, userID); err != nil {
		c.Logger().Warnf("Failed to revoke tokens: %v", err)
	}

	// Hard delete user (cascades to all related data)
	if err := h.userRepo.HardDelete(ctx, userID); err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete account")
	}

	return c.NoContent(http.StatusNoContent)
}

// generateAccessToken generates a JWT access token
func (h *AuthHandler) generateAccessToken(userID, deviceID uuid.UUID) (string, error) {
	now := time.Now()
	claims := &middleware.Claims{
		UserID:   userID,
		DeviceID: deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(h.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtSecret)
}

// generateRefreshToken generates a random refresh token
func (h *AuthHandler) generateRefreshToken() (token string, hash string, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}
	token = hex.EncodeToString(bytes)
	hash = hashToken(token)
	return token, hash, nil
}

// hashToken hashes a token using SHA-256
func hashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}

// errorResponse returns a standardized error response
func errorResponse(c echo.Context, status int, code, message string) error {
	return c.JSON(status, map[string]interface{}{
		"success": false,
		"data":    nil,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	})
}
