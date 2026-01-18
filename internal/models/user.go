package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user account
type User struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	Email        string     `db:"email" json:"email"`
	PasswordHash string     `db:"password_hash" json:"-"` // Never expose in JSON
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// UserResponse is the API response for user data
type UserResponse struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt int64     `json:"created_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		UserID:    u.ID,
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Unix(),
	}
}

// Device represents a user's device
type Device struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	UserID       uuid.UUID  `db:"user_id" json:"user_id"`
	DeviceID     string     `db:"device_id" json:"device_id"`   // Client-generated UUID
	DeviceName   *string    `db:"device_name" json:"device_name,omitempty"`
	LastSyncAt   *time.Time `db:"last_sync_at" json:"last_sync_at,omitempty"`
	LastActiveAt time.Time  `db:"last_active_at" json:"last_active_at"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
}

// RefreshToken represents a refresh token for authentication
type RefreshToken struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	UserID    uuid.UUID  `db:"user_id" json:"user_id"`
	DeviceID  uuid.UUID  `db:"device_id" json:"device_id"`
	TokenHash string     `db:"token_hash" json:"-"` // Never expose
	ExpiresAt time.Time  `db:"expires_at" json:"expires_at"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	RevokedAt *time.Time `db:"revoked_at" json:"revoked_at,omitempty"`
}

// UserProfile represents user preferences
type UserProfile struct {
	ID                uuid.UUID        `db:"id" json:"id"`
	UserID            uuid.UUID        `db:"user_id" json:"user_id"`
	AxisPreferences   JSONMap          `db:"axis_preferences" json:"axis_preferences"`
	ActionPreferences JSONMap          `db:"action_preferences" json:"action_preferences"`
	UpdatedAt         time.Time        `db:"updated_at" json:"updated_at"`
}

// DefaultAxisPreferences returns default axis preferences
func DefaultAxisPreferences() JSONMap {
	return JSONMap{
		"risk":         0,
		"success_rate": 0,
		"rationale":    0,
		"concept":      0,
	}
}

// DefaultActionPreferences returns default action preferences
func DefaultActionPreferences() JSONMap {
	return JSONMap{
		"dialog": 0,
		"books":  0,
		"videos": 0,
	}
}
