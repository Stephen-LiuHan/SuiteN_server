package models

import "github.com/google/uuid"

// SyncRequest represents the sync API request body
type SyncRequest struct {
	LastSyncedAt int64              `json:"last_sync_at"`
	DeviceID     string             `json:"device_id"`
	Changes      SyncChangesRequest `json:"changes"`
}

// Generic SyncItem to wrap upsert/delete actions
type SyncItem[T any] struct {
	ID     uuid.UUID `json:"id"`
	Action string    `json:"action"` // "upsert" or "delete"
	Data   T         `json:"data,omitempty"`
}

// SyncChangesRequest represents the changes in a sync request
type SyncChangesRequest struct {
	Books        []SyncItem[BookSyncData]        `json:"books,omitempty"`
	Sessions     []SyncItem[SessionSyncData]     `json:"reading_sessions,omitempty"`
	Captures     []SyncItem[CaptureSyncData]     `json:"captures,omitempty"`
	AIResults    []SyncItem[AIResultSyncData]    `json:"ai_results,omitempty"`
	NextActions  []SyncItem[NextActionSyncData]  `json:"next_actions,omitempty"`
	FeedbackLogs []SyncItem[FeedbackLogSyncData] `json:"feedback_logs,omitempty"`
	ActionLogs   []SyncItem[ActionLogSyncData]   `json:"action_logs,omitempty"`
}

// BookSyncData represents the data for a book in a sync request
type BookSyncData struct {
	Title     string  `json:"title"`
	Author    *string `json:"author,omitempty"`
	Summary   *string `json:"summary,omitempty"`
	CreatedAt int64   `json:"created_at"`
	UpdatedAt int64   `json:"updated_at"`
}

// SessionSyncData represents the data for a reading session in a sync request
type SessionSyncData struct {
	BookID    uuid.UUID `json:"book_id"`
	StartedAt int64     `json:"started_at"`
	EndedAt   *int64    `json:"ended_at,omitempty"`
	Note      *string   `json:"note,omitempty"`
	CreatedAt int64     `json:"created_at"`
	UpdatedAt int64     `json:"updated_at"`
}

// CaptureSyncData represents the data for a capture in a sync request
type CaptureSyncData struct {
	SessionID uuid.UUID `json:"session_id"`
	Emotion   *string   `json:"emotion,omitempty"`
	UserMemo  *string   `json:"user_memo,omitempty"`
	AIStatus  string    `json:"ai_status"`
	CreatedAt int64     `json:"created_at"`
	UpdatedAt int64     `json:"updated_at"`
}

// AIResultSyncData represents the data for an AI result in a sync request
type AIResultSyncData struct {
	CaptureID         uuid.UUID       `json:"capture_id"`
	DraftSummary      string          `json:"draft_summary"`
	DeepeningQuestion string          `json:"deepening_question"`
	AxisSelected      *string         `json:"axis_selected,omitempty"`
	TopicTags         JSONStringArray `json:"topic_tags,omitempty"`
	CreatedAt         int64           `json:"created_at"`
	UpdatedAt         int64           `json:"updated_at"`
}

// NextActionSyncData represents the data for a next action in a sync request
type NextActionSyncData struct {
	AIResultID  uuid.UUID  `json:"ai_result_id"`
	Type        ActionType `json:"type"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Query       *string    `json:"query,omitempty"`
	Link        *string    `json:"link,omitempty"`
	Rank        int        `json:"rank"`
}

// FeedbackLogSyncData represents the data for a feedback log in a sync request
type FeedbackLogSyncData struct {
	CaptureID   uuid.UUID `json:"capture_id"`
	Helpfulness int       `json:"helpfulness"`
	CreatedAt   int64     `json:"created_at"`
}

// ActionLogSyncData represents the data for an action log in a sync request
type ActionLogSyncData struct {
	AIResultID uuid.UUID `json:"ai_result_id"`
	ActionType string    `json:"action_type"`
	Opened     bool      `json:"opened"`
	Completed  bool      `json:"completed"`
	CreatedAt  int64     `json:"created_at"`
}

// SyncResponse represents the sync API response
type SyncResponse struct {
	ServerTimestamp int64               `json:"server_timestamp"`
	Changes         SyncChangesResponse `json:"changes"`
}

// SyncChangesResponse represents the changes in a sync response
type SyncChangesResponse struct {
	Books       []BookResponse       `json:"books"`
	Sessions    []SessionResponse    `json:"sessions"`
	Captures    []CaptureResponse    `json:"captures"`
	AIResults   []AIResultResponse   `json:"ai_results"`
	NextActions []NextActionResponse `json:"next_actions"`
}

// BookResponse represents a book in sync response
type BookResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Author    *string   `json:"author,omitempty"`
	Summary   *string   `json:"summary,omitempty"`
	CreatedAt int64     `json:"created_at"`
	UpdatedAt int64     `json:"updated_at"`
	DeletedAt *int64    `json:"deleted_at,omitempty"`
}

// SessionResponse represents a reading session in sync response
type SessionResponse struct {
	ID        uuid.UUID `json:"id"`
	BookID    uuid.UUID `json:"book_id"`
	StartedAt int64     `json:"started_at"`
	EndedAt   *int64    `json:"ended_at,omitempty"`
	Note      *string   `json:"note,omitempty"`
	CreatedAt int64     `json:"created_at"`
	UpdatedAt int64     `json:"updated_at"`
	DeletedAt *int64    `json:"deleted_at,omitempty"`
}

// CaptureResponse represents a capture in sync response
type CaptureResponse struct {
	ID        uuid.UUID `json:"id"`
	SessionID uuid.UUID `json:"session_id"`
	Emotion   *string   `json:"emotion,omitempty"`
	UserMemo  *string   `json:"user_memo,omitempty"`
	AIStatus  string    `json:"ai_status"`
	CreatedAt int64     `json:"created_at"`
	UpdatedAt int64     `json:"updated_at"`
	DeletedAt *int64    `json:"deleted_at,omitempty"`
}

// AIResultResponse represents an AI result in sync response
type AIResultResponse struct {
	ID                uuid.UUID       `json:"id"`
	CaptureID         uuid.UUID       `json:"capture_id"`
	DraftSummary      string          `json:"draft_summary"`
	DeepeningQuestion string          `json:"deepening_question"`
	AxisSelected      *string         `json:"axis_selected,omitempty"`
	TopicTags         JSONStringArray `json:"topic_tags,omitempty"`
	CreatedAt         int64           `json:"created_at"`
	UpdatedAt         int64           `json:"updated_at"`
	DeletedAt         *int64          `json:"deleted_at,omitempty"`
}

// NextActionResponse represents a next action in sync response
type NextActionResponse struct {
	ID          uuid.UUID  `json:"id"`
	AIResultID  uuid.UUID  `json:"ai_result_id"`
	Type        ActionType `json:"type"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Query       *string    `json:"query,omitempty"`
	Link        *string    `json:"link,omitempty"`
	Rank        int        `json:"rank"`
	CreatedAt   int64      `json:"created_at"`
	DeletedAt   *int64     `json:"deleted_at,omitempty"`
}
