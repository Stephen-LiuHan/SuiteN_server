package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/suiten/suiten-server/internal/models"
)

// UserRepository handles user data operations
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt)
	return err
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM users WHERE id = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ExistsByEmail checks if a user with the given email exists
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE email = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &count, query, email)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Delete soft deletes a user
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET deleted_at = $1, updated_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

// HardDelete permanently deletes a user and all related data
func (r *UserRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// DeviceRepository handles device data operations
type DeviceRepository struct {
	db *sqlx.DB
}

// NewDeviceRepository creates a new DeviceRepository
func NewDeviceRepository(db *sqlx.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

// Upsert creates or updates a device
func (r *DeviceRepository) Upsert(ctx context.Context, device *models.Device) error {
	query := `
		INSERT INTO devices (id, user_id, device_id, device_name, last_active_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, device_id) DO UPDATE SET
			device_name = EXCLUDED.device_name,
			last_active_at = EXCLUDED.last_active_at
		RETURNING id
	`
	return r.db.QueryRowContext(ctx, query,
		device.ID, device.UserID, device.DeviceID, device.DeviceName,
		device.LastActiveAt, device.CreatedAt).Scan(&device.ID)
}

// FindByUserAndDeviceID finds a device by user ID and device ID
func (r *DeviceRepository) FindByUserAndDeviceID(ctx context.Context, userID uuid.UUID, deviceID string) (*models.Device, error) {
	var device models.Device
	query := `SELECT * FROM devices WHERE user_id = $1 AND device_id = $2`
	err := r.db.GetContext(ctx, &device, query, userID, deviceID)
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// FindByID finds a device by ID
func (r *DeviceRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Device, error) {
	var device models.Device
	query := `SELECT * FROM devices WHERE id = $1`
	err := r.db.GetContext(ctx, &device, query, id)
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// UpdateLastSyncAt updates the last sync time
func (r *DeviceRepository) UpdateLastSyncAt(ctx context.Context, id uuid.UUID, syncAt time.Time) error {
	query := `UPDATE devices SET last_sync_at = $1, last_active_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, syncAt, id)
	return err
}

// RefreshTokenRepository handles refresh token operations
type RefreshTokenRepository struct {
	db *sqlx.DB
}

// NewRefreshTokenRepository creates a new RefreshTokenRepository
func NewRefreshTokenRepository(db *sqlx.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create creates a new refresh token
func (r *RefreshTokenRepository) Create(ctx context.Context, token *models.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, device_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		token.ID, token.UserID, token.DeviceID, token.TokenHash,
		token.ExpiresAt, token.CreatedAt)
	return err
}

// FindByTokenHash finds a refresh token by its hash
func (r *RefreshTokenRepository) FindByTokenHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	query := `SELECT * FROM refresh_tokens WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()`
	err := r.db.GetContext(ctx, &token, query, hash)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// Revoke revokes a refresh token
func (r *RefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

// RevokeAllForUser revokes all refresh tokens for a user
func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	return err
}

// RevokeAllForDevice revokes all refresh tokens for a device
func (r *RefreshTokenRepository) RevokeAllForDevice(ctx context.Context, deviceID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked_at = $1 WHERE device_id = $2 AND revoked_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, time.Now(), deviceID)
	return err
}

// UserProfileRepository handles user profile operations
type UserProfileRepository struct {
	db *sqlx.DB
}

// NewUserProfileRepository creates a new UserProfileRepository
func NewUserProfileRepository(db *sqlx.DB) *UserProfileRepository {
	return &UserProfileRepository{db: db}
}

// Create creates a new user profile
func (r *UserProfileRepository) Create(ctx context.Context, profile *models.UserProfile) error {
	query := `
		INSERT INTO user_profiles (id, user_id, axis_preferences, action_preferences, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query,
		profile.ID, profile.UserID, profile.AxisPreferences,
		profile.ActionPreferences, profile.UpdatedAt)
	return err
}

// FindByUserID finds a user profile by user ID
func (r *UserProfileRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*models.UserProfile, error) {
	var profile models.UserProfile
	query := `SELECT * FROM user_profiles WHERE user_id = $1`
	err := r.db.GetContext(ctx, &profile, query, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// Upsert creates or updates a user profile
func (r *UserProfileRepository) Upsert(ctx context.Context, profile *models.UserProfile) error {
	query := `
		INSERT INTO user_profiles (id, user_id, axis_preferences, action_preferences, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			axis_preferences = EXCLUDED.axis_preferences,
			action_preferences = EXCLUDED.action_preferences,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.ExecContext(ctx, query,
		profile.ID, profile.UserID, profile.AxisPreferences,
		profile.ActionPreferences, profile.UpdatedAt)
	return err
}
