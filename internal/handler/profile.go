package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	"github.com/suiten/suiten-server/internal/middleware"
	"github.com/suiten/suiten-server/internal/models"
	"github.com/suiten/suiten-server/internal/repository"
)

// ProfileHandler handles profile endpoints
type ProfileHandler struct {
	db          *sqlx.DB
	profileRepo *repository.UserProfileRepository
}

// NewProfileHandler creates a new ProfileHandler
func NewProfileHandler(db *sqlx.DB) *ProfileHandler {
	return &ProfileHandler{
		db:          db,
		profileRepo: repository.NewUserProfileRepository(db),
	}
}

// GetProfile returns the user's profile
func (h *ProfileHandler) GetProfile(c echo.Context) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errorResponse(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "Invalid token")
	}

	ctx := c.Request().Context()

	profile, err := h.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get profile")
	}

	if profile == nil {
		// Create default profile if it doesn't exist
		profile = &models.UserProfile{
			ID:                uuid.New(),
			UserID:            userID,
			AxisPreferences:   models.DefaultAxisPreferences(),
			ActionPreferences: models.DefaultActionPreferences(),
			UpdatedAt:         time.Now(),
		}
		if err := h.profileRepo.Create(ctx, profile); err != nil {
			return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create profile")
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"user_id":            profile.UserID,
			"axis_preferences":   profile.AxisPreferences,
			"action_preferences": profile.ActionPreferences,
			"updated_at":         profile.UpdatedAt.Unix(),
		},
	})
}

// UpdateProfileRequest represents the profile update request
type UpdateProfileRequest struct {
	AxisPreferences   models.JSONMap `json:"axis_preferences"`
	ActionPreferences models.JSONMap `json:"action_preferences"`
}

// UpdateProfile updates the user's profile
func (h *ProfileHandler) UpdateProfile(c echo.Context) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errorResponse(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "Invalid token")
	}

	var req UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	ctx := c.Request().Context()

	profile, err := h.profileRepo.FindByUserID(ctx, userID)
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get profile")
	}

	if profile == nil {
		profile = &models.UserProfile{
			ID:     uuid.New(),
			UserID: userID,
		}
	}

	// Update preferences if provided
	if req.AxisPreferences != nil {
		profile.AxisPreferences = req.AxisPreferences
	} else if profile.AxisPreferences == nil {
		profile.AxisPreferences = models.DefaultAxisPreferences()
	}

	if req.ActionPreferences != nil {
		profile.ActionPreferences = req.ActionPreferences
	} else if profile.ActionPreferences == nil {
		profile.ActionPreferences = models.DefaultActionPreferences()
	}

	profile.UpdatedAt = time.Now()

	if err := h.profileRepo.Upsert(ctx, profile); err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update profile")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"user_id":            profile.UserID,
			"axis_preferences":   profile.AxisPreferences,
			"action_preferences": profile.ActionPreferences,
			"updated_at":         profile.UpdatedAt.Unix(),
		},
	})
}
