package models

import (
	"time"

	"github.com/google/uuid"
)

// Book represents a book entity
type Book struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	UserID    uuid.UUID  `db:"user_id" json:"user_id"`
	Title     string     `db:"title" json:"title"`
	Author    *string    `db:"author" json:"author,omitempty"`
	Summary   *string    `db:"summary" json:"summary,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`

	// Book-level summary metrics (North Star Metric)
	SummaryGeneratedAt  *time.Time `db:"summary_generated_at" json:"summary_generated_at,omitempty"`
	SummarySavedAt      *time.Time `db:"summary_saved_at" json:"summary_saved_at,omitempty"`
	SummaryEditedAt     *time.Time `db:"summary_edited_at" json:"summary_edited_at,omitempty"`
	SummaryHelpfulness  *int       `db:"summary_helpfulness" json:"summary_helpfulness,omitempty"`
	SummaryCaptureCount *int       `db:"summary_capture_count" json:"summary_capture_count,omitempty"`
	SummaryCompletedAt  *time.Time `db:"summary_completed_at" json:"summary_completed_at,omitempty"`
}

// BookSyncRequest represents a book in sync request
type BookSyncRequest struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Author    *string   `json:"author,omitempty"`
	Summary   *string   `json:"summary,omitempty"`
	CreatedAt int64     `json:"created_at"`
	UpdatedAt int64     `json:"updated_at"`
}

// ToModel converts BookSyncRequest to Book
func (r *BookSyncRequest) ToModel(userID uuid.UUID) *Book {
	return &Book{
		ID:        r.ID,
		UserID:    userID,
		Title:     r.Title,
		Author:    r.Author,
		Summary:   r.Summary,
		CreatedAt: UnixToTime(r.CreatedAt),
		UpdatedAt: UnixToTime(r.UpdatedAt),
	}
}

// ToSyncResponse converts Book to sync response format
func (b *Book) ToSyncResponse() map[string]interface{} {
	resp := map[string]interface{}{
		"id":         b.ID,
		"title":      b.Title,
		"created_at": b.CreatedAt.Unix(),
		"updated_at": b.UpdatedAt.Unix(),
	}
	if b.Author != nil {
		resp["author"] = *b.Author
	}
	if b.Summary != nil {
		resp["summary"] = *b.Summary
	}
	if b.DeletedAt != nil {
		resp["deleted_at"] = b.DeletedAt.Unix()
	}
	return resp
}
