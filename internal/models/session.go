package models

import (
	"time"

	"github.com/google/uuid"
)

// ReadingSession represents a reading session
type ReadingSession struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	UserID    uuid.UUID  `db:"user_id" json:"user_id"`
	BookID    uuid.UUID  `db:"book_id" json:"book_id"`
	StartedAt time.Time  `db:"started_at" json:"started_at"`
	EndedAt   *time.Time `db:"ended_at" json:"ended_at,omitempty"`
	Note      *string    `db:"note" json:"note,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// ReadingSessionSyncRequest represents a reading session in sync request
type ReadingSessionSyncRequest struct {
	ID        uuid.UUID `json:"id"`
	BookID    uuid.UUID `json:"book_id"`
	StartedAt int64     `json:"started_at"`
	EndedAt   *int64    `json:"ended_at,omitempty"`
	Note      *string   `json:"note,omitempty"`
	CreatedAt int64     `json:"created_at"`
	UpdatedAt int64     `json:"updated_at"`
}

// ToModel converts ReadingSessionSyncRequest to ReadingSession
func (r *ReadingSessionSyncRequest) ToModel(userID uuid.UUID) *ReadingSession {
	session := &ReadingSession{
		ID:        r.ID,
		UserID:    userID,
		BookID:    r.BookID,
		StartedAt: UnixToTime(r.StartedAt),
		Note:      r.Note,
		CreatedAt: UnixToTime(r.CreatedAt),
		UpdatedAt: UnixToTime(r.UpdatedAt),
	}
	if r.EndedAt != nil {
		t := UnixToTime(*r.EndedAt)
		session.EndedAt = &t
	}
	return session
}

// ToSyncResponse converts ReadingSession to sync response format
func (s *ReadingSession) ToSyncResponse() map[string]interface{} {
	resp := map[string]interface{}{
		"id":         s.ID,
		"book_id":    s.BookID,
		"started_at": s.StartedAt.Unix(),
		"created_at": s.CreatedAt.Unix(),
		"updated_at": s.UpdatedAt.Unix(),
	}
	if s.EndedAt != nil {
		resp["ended_at"] = s.EndedAt.Unix()
	}
	if s.Note != nil {
		resp["note"] = *s.Note
	}
	if s.DeletedAt != nil {
		resp["deleted_at"] = s.DeletedAt.Unix()
	}
	return resp
}
