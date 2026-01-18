package models

import (
	"time"

	"github.com/google/uuid"
)

// FeedbackLog represents user feedback on AI results
type FeedbackLog struct {
	ID          uuid.UUID `db:"id" json:"id"`
	UserID      uuid.UUID `db:"user_id" json:"user_id"`
	CaptureID   uuid.UUID `db:"capture_id" json:"capture_id"`
	Helpfulness int       `db:"helpfulness" json:"helpfulness"` // 1-3
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// FeedbackLogSyncRequest represents a feedback log in sync request
type FeedbackLogSyncRequest struct {
	ID          uuid.UUID `json:"id"`
	CaptureID   uuid.UUID `json:"capture_id"`
	Helpfulness int       `json:"helpfulness"`
	CreatedAt   int64     `json:"created_at"`
}

// ToModel converts FeedbackLogSyncRequest to FeedbackLog
func (r *FeedbackLogSyncRequest) ToModel(userID uuid.UUID) *FeedbackLog {
	return &FeedbackLog{
		ID:          r.ID,
		UserID:      userID,
		CaptureID:   r.CaptureID,
		Helpfulness: r.Helpfulness,
		CreatedAt:   time.Unix(r.CreatedAt, 0),
	}
}

// ToSyncResponse converts FeedbackLog to sync response format
func (f *FeedbackLog) ToSyncResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":          f.ID,
		"capture_id":  f.CaptureID,
		"helpfulness": f.Helpfulness,
		"created_at":  f.CreatedAt.Unix(),
	}
}

// ActionLog represents user action tracking
type ActionLog struct {
	ID         uuid.UUID `db:"id" json:"id"`
	UserID     uuid.UUID `db:"user_id" json:"user_id"`
	AIResultID uuid.UUID `db:"ai_result_id" json:"ai_result_id"`
	ActionType string    `db:"action_type" json:"action_type"`
	Opened     bool      `db:"opened" json:"opened"`
	Completed  bool      `db:"completed" json:"completed"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

// ActionLogSyncRequest represents an action log in sync request
type ActionLogSyncRequest struct {
	ID         uuid.UUID `json:"id"`
	AIResultID uuid.UUID `json:"ai_result_id"`
	ActionType string    `json:"action_type"`
	Opened     bool      `json:"opened"`
	Completed  bool      `json:"completed"`
	CreatedAt  int64     `json:"created_at"`
}

// ToModel converts ActionLogSyncRequest to ActionLog
func (r *ActionLogSyncRequest) ToModel(userID uuid.UUID) *ActionLog {
	return &ActionLog{
		ID:         r.ID,
		UserID:     userID,
		AIResultID: r.AIResultID,
		ActionType: r.ActionType,
		Opened:     r.Opened,
		Completed:  r.Completed,
		CreatedAt:  time.Unix(r.CreatedAt, 0),
	}
}

// ToSyncResponse converts ActionLog to sync response format
func (a *ActionLog) ToSyncResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":           a.ID,
		"ai_result_id": a.AIResultID,
		"action_type":  a.ActionType,
		"opened":       a.Opened,
		"completed":    a.Completed,
		"created_at":   a.CreatedAt.Unix(),
	}
}
