package models

import (
	"time"

	"github.com/google/uuid"
)

// AIResult represents an AI analysis result
type AIResult struct {
	ID                uuid.UUID       `db:"id" json:"id"`
	UserID            uuid.UUID       `db:"user_id" json:"user_id"`
	CaptureID         uuid.UUID       `db:"capture_id" json:"capture_id"`
	DraftSummary      string          `db:"draft_summary" json:"draft_summary"`
	DeepeningQuestion string          `db:"deepening_question" json:"deepening_question"`
	AxisSelected      *string         `db:"axis_selected" json:"axis_selected,omitempty"`
	TopicTags         JSONStringArray `db:"topic_tags" json:"topic_tags"`
	CreatedAt         time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time       `db:"updated_at" json:"updated_at"`
	DeletedAt         *time.Time      `db:"deleted_at" json:"deleted_at,omitempty"`
}

// AIResultSyncRequest represents an AI result in sync request
type AIResultSyncRequest struct {
	ID                uuid.UUID `json:"id"`
	CaptureID         uuid.UUID `json:"capture_id"`
	DraftSummary      string    `json:"draft_summary"`
	DeepeningQuestion string    `json:"deepening_question"`
	AxisSelected      *string   `json:"axis_selected,omitempty"`
	TopicTags         []string  `json:"topic_tags"`
	CreatedAt         int64     `json:"created_at"`
	UpdatedAt         int64     `json:"updated_at"`
}

// ToModel converts AIResultSyncRequest to AIResult
func (r *AIResultSyncRequest) ToModel(userID uuid.UUID) *AIResult {
	return &AIResult{
		ID:                r.ID,
		UserID:            userID,
		CaptureID:         r.CaptureID,
		DraftSummary:      r.DraftSummary,
		DeepeningQuestion: r.DeepeningQuestion,
		AxisSelected:      r.AxisSelected,
		TopicTags:         r.TopicTags,
		CreatedAt:         UnixToTime(r.CreatedAt),
		UpdatedAt:         UnixToTime(r.UpdatedAt),
	}
}

// ToSyncResponse converts AIResult to sync response format
func (a *AIResult) ToSyncResponse() map[string]interface{} {
	resp := map[string]interface{}{
		"id":                 a.ID,
		"capture_id":         a.CaptureID,
		"draft_summary":      a.DraftSummary,
		"deepening_question": a.DeepeningQuestion,
		"topic_tags":         a.TopicTags,
		"created_at":         a.CreatedAt.Unix(),
		"updated_at":         a.UpdatedAt.Unix(),
	}
	if a.AxisSelected != nil {
		resp["axis_selected"] = *a.AxisSelected
	}
	if a.DeletedAt != nil {
		resp["deleted_at"] = a.DeletedAt.Unix()
	}
	return resp
}

// NextAction represents a suggested next action
type NextAction struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	UserID      uuid.UUID  `db:"user_id" json:"user_id"`
	AIResultID  uuid.UUID  `db:"ai_result_id" json:"ai_result_id"`
	Type        ActionType `db:"type" json:"type"`
	Title       string     `db:"title" json:"title"`
	Description string     `db:"description" json:"description"`
	Query       *string    `db:"query" json:"query,omitempty"`
	Link        *string    `db:"link" json:"link,omitempty"`
	Rank        int        `db:"rank" json:"rank"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	DeletedAt   *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// NextActionSyncRequest represents a next action in sync request
type NextActionSyncRequest struct {
	ID          uuid.UUID  `json:"id"`
	AIResultID  uuid.UUID  `json:"ai_result_id"`
	Type        ActionType `json:"type"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Query       *string    `json:"query,omitempty"`
	Link        *string    `json:"link,omitempty"`
	Rank        int        `json:"rank"`
	CreatedAt   int64      `json:"created_at"`
}

// ToModel converts NextActionSyncRequest to NextAction
func (r *NextActionSyncRequest) ToModel(userID uuid.UUID) *NextAction {
	return &NextAction{
		ID:          r.ID,
		UserID:      userID,
		AIResultID:  r.AIResultID,
		Type:        r.Type,
		Title:       r.Title,
		Description: r.Description,
		Query:       r.Query,
		Link:        r.Link,
		Rank:        r.Rank,
		CreatedAt:   UnixToTime(r.CreatedAt),
	}
}

// ToSyncResponse converts NextAction to sync response format
func (n *NextAction) ToSyncResponse() map[string]interface{} {
	resp := map[string]interface{}{
		"id":           n.ID,
		"ai_result_id": n.AIResultID,
		"type":         n.Type,
		"title":        n.Title,
		"description":  n.Description,
		"rank":         n.Rank,
		"created_at":   n.CreatedAt.Unix(),
	}
	if n.Query != nil {
		resp["query"] = *n.Query
	}
	if n.Link != nil {
		resp["link"] = *n.Link
	}
	if n.DeletedAt != nil {
		resp["deleted_at"] = n.DeletedAt.Unix()
	}
	return resp
}
