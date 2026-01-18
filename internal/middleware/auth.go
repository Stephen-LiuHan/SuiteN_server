package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey     string
	SigningMethod jwt.SigningMethod
}

// DefaultJWTConfig returns default JWT configuration
func DefaultJWTConfig() *JWTConfig {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "development-secret-key-change-in-production"
	}
	return &JWTConfig{
		SecretKey:     secret,
		SigningMethod: jwt.SigningMethodHS256,
	}
}

// Claims represents JWT claims
type Claims struct {
	UserID   uuid.UUID `json:"sub"`
	DeviceID uuid.UUID `json:"device_id"`
	jwt.RegisteredClaims
}

// JWTAuth returns JWT authentication middleware
func JWTAuth(cfg *JWTConfig) echo.MiddlewareFunc {
	if cfg == nil {
		cfg = DefaultJWTConfig()
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, map[string]interface{}{
					"success": false,
					"error": map[string]interface{}{
						"code":    "AUTH_TOKEN_MISSING",
						"message": "Authorization header is required",
					},
				})
			}

			// Extract token from "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, map[string]interface{}{
					"success": false,
					"error": map[string]interface{}{
						"code":    "AUTH_TOKEN_INVALID",
						"message": "Invalid authorization header format",
					},
				})
			}

			tokenString := parts[1]

			// Parse and validate token
			token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
				if token.Method != cfg.SigningMethod {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "Invalid signing method")
				}
				return []byte(cfg.SecretKey), nil
			})

			if err != nil {
				code := "AUTH_TOKEN_INVALID"
				message := "Invalid token"
				if err == jwt.ErrTokenExpired {
					code = "AUTH_TOKEN_EXPIRED"
					message = "Token has expired"
				}
				return echo.NewHTTPError(http.StatusUnauthorized, map[string]interface{}{
					"success": false,
					"error": map[string]interface{}{
						"code":    code,
						"message": message,
					},
				})
			}

			claims, ok := token.Claims.(*Claims)
			if !ok || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, map[string]interface{}{
					"success": false,
					"error": map[string]interface{}{
						"code":    "AUTH_TOKEN_INVALID",
						"message": "Invalid token claims",
					},
				})
			}

			// Store user info in context
			c.Set("user_id", claims.UserID)
			c.Set("device_id", claims.DeviceID)

			return next(c)
		}
	}
}

// GetUserID extracts user ID from context
func GetUserID(c echo.Context) (uuid.UUID, error) {
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}
	return userID, nil
}

// GetDeviceID extracts device ID from context
func GetDeviceID(c echo.Context) (uuid.UUID, error) {
	deviceID, ok := c.Get("device_id").(uuid.UUID)
	if !ok {
		return uuid.Nil, echo.NewHTTPError(http.StatusUnauthorized, "Device ID not found in context")
	}
	return deviceID, nil
}
