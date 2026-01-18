package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/suiten/suiten-server/internal/models"
)

// SyncRepository handles sync-related data operations
type SyncRepository struct {
	db *sqlx.DB
}

// NewSyncRepository creates a new SyncRepository
func NewSyncRepository(db *sqlx.DB) *SyncRepository {
	return &SyncRepository{db: db}
}

// Book operations

func (r *SyncRepository) UpsertBook(ctx context.Context, book *models.Book) error {
	query := `
		INSERT INTO books (id, user_id, title, author, summary, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			author = EXCLUDED.author,
			summary = EXCLUDED.summary,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
		WHERE books.user_id = EXCLUDED.user_id
	`
	_, err := r.db.ExecContext(ctx, query,
		book.ID, book.UserID, book.Title, book.Author, book.Summary,
		book.CreatedAt, book.UpdatedAt, book.DeletedAt)
	return err
}

func (r *SyncRepository) DeleteBook(ctx context.Context, userID, bookID uuid.UUID) error {
	query := `UPDATE books SET deleted_at = $1, updated_at = $1 WHERE id = $2 AND user_id = $3`
	_, err := r.db.ExecContext(ctx, query, time.Now(), bookID, userID)
	return err
}

func (r *SyncRepository) GetBooksUpdatedSince(ctx context.Context, userID uuid.UUID, since int64) ([]models.Book, error) {
	var books []models.Book
	query := `SELECT * FROM books WHERE user_id = $1 AND EXTRACT(EPOCH FROM updated_at) > $2 ORDER BY updated_at`
	err := r.db.SelectContext(ctx, &books, query, userID, since)
	return books, err
}

// ReadingSession operations

func (r *SyncRepository) UpsertReadingSession(ctx context.Context, session *models.ReadingSession) error {
	query := `
		INSERT INTO reading_sessions (id, user_id, book_id, started_at, ended_at, note, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			started_at = EXCLUDED.started_at,
			ended_at = EXCLUDED.ended_at,
			note = EXCLUDED.note,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
		WHERE reading_sessions.user_id = EXCLUDED.user_id
	`
	_, err := r.db.ExecContext(ctx, query,
		session.ID, session.UserID, session.BookID, session.StartedAt,
		session.EndedAt, session.Note, session.CreatedAt, session.UpdatedAt, session.DeletedAt)
	return err
}

func (r *SyncRepository) DeleteReadingSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	query := `UPDATE reading_sessions SET deleted_at = $1, updated_at = $1 WHERE id = $2 AND user_id = $3`
	_, err := r.db.ExecContext(ctx, query, time.Now(), sessionID, userID)
	return err
}

func (r *SyncRepository) GetSessionsUpdatedSince(ctx context.Context, userID uuid.UUID, since int64) ([]models.ReadingSession, error) {
	var sessions []models.ReadingSession
	query := `SELECT * FROM reading_sessions WHERE user_id = $1 AND EXTRACT(EPOCH FROM updated_at) > $2 ORDER BY updated_at`
	err := r.db.SelectContext(ctx, &sessions, query, userID, since)
	return sessions, err
}

// Capture operations

func (r *SyncRepository) UpsertCapture(ctx context.Context, capture *models.Capture) error {
	query := `
		INSERT INTO captures (id, user_id, session_id, emotion, user_memo, ai_status, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			emotion = EXCLUDED.emotion,
			user_memo = EXCLUDED.user_memo,
			ai_status = EXCLUDED.ai_status,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
		WHERE captures.user_id = EXCLUDED.user_id
	`
	_, err := r.db.ExecContext(ctx, query,
		capture.ID, capture.UserID, capture.SessionID, capture.Emotion,
		capture.UserMemo, capture.AIStatus, capture.CreatedAt, capture.UpdatedAt, capture.DeletedAt)
	return err
}

func (r *SyncRepository) DeleteCapture(ctx context.Context, userID, captureID uuid.UUID) error {
	query := `UPDATE captures SET deleted_at = $1, updated_at = $1 WHERE id = $2 AND user_id = $3`
	_, err := r.db.ExecContext(ctx, query, time.Now(), captureID, userID)
	return err
}

func (r *SyncRepository) GetCapturesUpdatedSince(ctx context.Context, userID uuid.UUID, since int64) ([]models.Capture, error) {
	var captures []models.Capture
	query := `SELECT * FROM captures WHERE user_id = $1 AND EXTRACT(EPOCH FROM updated_at) > $2 ORDER BY updated_at`
	err := r.db.SelectContext(ctx, &captures, query, userID, since)
	return captures, err
}

// AIResult operations

func (r *SyncRepository) UpsertAIResult(ctx context.Context, result *models.AIResult) error {
	query := `
		INSERT INTO ai_results (id, user_id, capture_id, draft_summary, deepening_question, axis_selected, topic_tags, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			draft_summary = EXCLUDED.draft_summary,
			deepening_question = EXCLUDED.deepening_question,
			axis_selected = EXCLUDED.axis_selected,
			topic_tags = EXCLUDED.topic_tags,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
		WHERE ai_results.user_id = EXCLUDED.user_id
	`
	_, err := r.db.ExecContext(ctx, query,
		result.ID, result.UserID, result.CaptureID, result.DraftSummary,
		result.DeepeningQuestion, result.AxisSelected, result.TopicTags,
		result.CreatedAt, result.UpdatedAt, result.DeletedAt)
	return err
}

func (r *SyncRepository) DeleteAIResult(ctx context.Context, userID, resultID uuid.UUID) error {
	query := `UPDATE ai_results SET deleted_at = $1, updated_at = $1 WHERE id = $2 AND user_id = $3`
	_, err := r.db.ExecContext(ctx, query, time.Now(), resultID, userID)
	return err
}

func (r *SyncRepository) GetAIResultsUpdatedSince(ctx context.Context, userID uuid.UUID, since int64) ([]models.AIResult, error) {
	var results []models.AIResult
	query := `SELECT * FROM ai_results WHERE user_id = $1 AND EXTRACT(EPOCH FROM updated_at) > $2 ORDER BY updated_at`
	err := r.db.SelectContext(ctx, &results, query, userID, since)
	return results, err
}

// NextAction operations

func (r *SyncRepository) UpsertNextAction(ctx context.Context, action *models.NextAction) error {
	query := `
		INSERT INTO next_actions (id, user_id, ai_result_id, type, title, description, query, link, rank, created_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			type = EXCLUDED.type,
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			query = EXCLUDED.query,
			link = EXCLUDED.link,
			rank = EXCLUDED.rank,
			deleted_at = EXCLUDED.deleted_at
		WHERE next_actions.user_id = EXCLUDED.user_id
	`
	_, err := r.db.ExecContext(ctx, query,
		action.ID, action.UserID, action.AIResultID, action.Type,
		action.Title, action.Description, action.Query, action.Link,
		action.Rank, action.CreatedAt, action.DeletedAt)
	return err
}

func (r *SyncRepository) DeleteNextAction(ctx context.Context, userID, actionID uuid.UUID) error {
	query := `UPDATE next_actions SET deleted_at = $1 WHERE id = $2 AND user_id = $3`
	_, err := r.db.ExecContext(ctx, query, time.Now(), actionID, userID)
	return err
}

func (r *SyncRepository) GetNextActionsUpdatedSince(ctx context.Context, userID uuid.UUID, since int64) ([]models.NextAction, error) {
	var actions []models.NextAction
	query := `SELECT * FROM next_actions WHERE user_id = $1 AND EXTRACT(EPOCH FROM created_at) > $2 ORDER BY created_at`
	err := r.db.SelectContext(ctx, &actions, query, userID, since)
	return actions, err
}

// FeedbackLog operations

func (r *SyncRepository) UpsertFeedbackLog(ctx context.Context, log *models.FeedbackLog) error {
	query := `
		INSERT INTO feedback_logs (id, user_id, capture_id, helpfulness, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			helpfulness = EXCLUDED.helpfulness
		WHERE feedback_logs.user_id = EXCLUDED.user_id
	`
	_, err := r.db.ExecContext(ctx, query,
		log.ID, log.UserID, log.CaptureID, log.Helpfulness, log.CreatedAt)
	return err
}

func (r *SyncRepository) GetFeedbackLogsCreatedSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]models.FeedbackLog, error) {
	var logs []models.FeedbackLog
	query := `SELECT * FROM feedback_logs WHERE user_id = $1 AND created_at >= $2 ORDER BY created_at`
	err := r.db.SelectContext(ctx, &logs, query, userID, since)
	return logs, err
}

// ActionLog operations

func (r *SyncRepository) UpsertActionLog(ctx context.Context, log *models.ActionLog) error {
	query := `
		INSERT INTO action_logs (id, user_id, ai_result_id, action_type, opened, completed, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			opened = EXCLUDED.opened,
			completed = EXCLUDED.completed
		WHERE action_logs.user_id = EXCLUDED.user_id
	`
	_, err := r.db.ExecContext(ctx, query,
		log.ID, log.UserID, log.AIResultID, log.ActionType,
		log.Opened, log.Completed, log.CreatedAt)
	return err
}

func (r *SyncRepository) GetActionLogsCreatedSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]models.ActionLog, error) {
	var logs []models.ActionLog
	query := `SELECT * FROM action_logs WHERE user_id = $1 AND created_at >= $2 ORDER BY created_at`
	err := r.db.SelectContext(ctx, &logs, query, userID, since)
	return logs, err
}

// GetSyncStats returns sync statistics for a user
func (r *SyncRepository) GetSyncStats(ctx context.Context, userID uuid.UUID) (int64, error) {
	var totalSize int64
	query := `
		SELECT COALESCE(SUM(pg_column_size(t.*)), 0)
		FROM (
			SELECT * FROM books WHERE user_id = $1
			UNION ALL SELECT * FROM reading_sessions WHERE user_id = $1
			UNION ALL SELECT * FROM captures WHERE user_id = $1
			UNION ALL SELECT * FROM ai_results WHERE user_id = $1
		) t
	`
	err := r.db.GetContext(ctx, &totalSize, query, userID)
	return totalSize, err
}
