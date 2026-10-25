package models

import (
	"time"

	"github.com/google/uuid"
)

// Capture represents a capture (screenshot metadata)
// Note: image_path is NOT stored on server - client local only
type Capture struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	UserID    uuid.UUID  `db:"user_id" json:"user_id"`
	SessionID uuid.UUID  `db:"session_id" json:"session_id"`
	Emotion   *string    `db:"emotion" json:"emotion,omitempty"`
	UserMemo  *string    `db:"user_memo" json:"user_memo,omitempty"`
	AIStatus  AIStatus   `db:"ai_status" json:"ai_status"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// CaptureSyncRequest represents a capture in sync request
type CaptureSyncRequest struct {
	ID        uuid.UUID `json:"id"`
	SessionID uuid.UUID `json:"session_id"`
	Emotion   *string   `json:"emotion,omitempty"`
	UserMemo  *string   `json:"user_memo,omitempty"`
	AIStatus  AIStatus  `json:"ai_status"`
	CreatedAt int64     `json:"created_at"`
	UpdatedAt int64     `json:"updated_at"`
}

// ToModel converts CaptureSyncRequest to Capture
func (r *CaptureSyncRequest) ToModel(userID uuid.UUID) *Capture {
	return &Capture{
		ID:        r.ID,
		UserID:    userID,
		SessionID: r.SessionID,
		Emotion:   r.Emotion,
		UserMemo:  r.UserMemo,
		AIStatus:  r.AIStatus,
		CreatedAt: UnixToTime(r.CreatedAt),
		UpdatedAt: UnixToTime(r.UpdatedAt),
	}
}

// ToSyncResponse converts Capture to sync response format
func (c *Capture) ToSyncResponse() map[string]interface{} {
	resp := map[string]interface{}{
		"id":         c.ID,
		"session_id": c.SessionID,
		"ai_status":  c.AIStatus,
		"created_at": c.CreatedAt.Unix(),
		"updated_at": c.UpdatedAt.Unix(),
	}
	if c.Emotion != nil {
		resp["emotion"] = *c.Emotion
	}
	if c.UserMemo != nil {
		resp["user_memo"] = *c.UserMemo
	}
	if c.DeletedAt != nil {
		resp["deleted_at"] = c.DeletedAt.Unix()
	}
	return resp
}
