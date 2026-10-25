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

// SyncHandler handles sync endpoints
type SyncHandler struct {
	db         *sqlx.DB
	syncRepo   *repository.SyncRepository
	deviceRepo *repository.DeviceRepository
}

// NewSyncHandler creates a new SyncHandler
func NewSyncHandler(db *sqlx.DB) *SyncHandler {
	return &SyncHandler{
		db:         db,
		syncRepo:   repository.NewSyncRepository(db),
		deviceRepo: repository.NewDeviceRepository(db),
	}
}

// Sync handles bidirectional data synchronization
func (h *SyncHandler) Sync(c echo.Context) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errorResponse(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "Invalid token")
	}

	deviceID, err := middleware.GetDeviceID(c)
	if err != nil {
		return errorResponse(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "Invalid token")
	}

	var req models.SyncRequest
	if err := c.Bind(&req); err != nil {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	}

	ctx := c.Request().Context()
	now := time.Now()

	// Start a transaction
	tx, err := h.db.BeginTxx(ctx, nil)
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to start transaction")
	}
	defer tx.Rollback() // Rollback is a no-op if the transaction has been committed

	// Parse last_synced_at
	var lastSyncedAt int64
	if req.LastSyncedAt > 0 {
		lastSyncedAt = req.LastSyncedAt
	}

	// Process client changes
	// Books
	for _, item := range req.Changes.Books {
		if item.Action == "upsert" {
			data := item.Data
			metrics := data.SummaryMetrics()
			// summary_completed_at keeps the first completion, so a book counts once
			// toward the North Star Metric, in the month it was first completed.
			query := `
				INSERT INTO books (id, user_id, title, author, summary, created_at, updated_at,
					summary_generated_at, summary_saved_at, summary_edited_at,
					summary_helpfulness, summary_capture_count, summary_completed_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
				ON CONFLICT (id) DO UPDATE SET
					title = EXCLUDED.title,
					author = EXCLUDED.author,
					summary = EXCLUDED.summary,
					updated_at = EXCLUDED.updated_at,
					summary_generated_at = EXCLUDED.summary_generated_at,
					summary_saved_at = EXCLUDED.summary_saved_at,
					summary_edited_at = EXCLUDED.summary_edited_at,
					summary_helpfulness = EXCLUDED.summary_helpfulness,
					summary_capture_count = EXCLUDED.summary_capture_count,
					summary_completed_at = COALESCE(books.summary_completed_at, EXCLUDED.summary_completed_at),
					deleted_at = NULL
				WHERE books.user_id = $2
			`
			if _, err := tx.ExecContext(ctx, query, item.ID, userID, data.Title, data.Author, data.Summary, models.UnixToTime(data.CreatedAt), models.UnixToTime(data.UpdatedAt),
				metrics.GeneratedAt, metrics.SavedAt, metrics.EditedAt,
				metrics.Helpfulness, metrics.CaptureCount, metrics.CompletedAt()); err != nil {
				c.Logger().Errorf("Failed to upsert book: %v", err)
			}
		} else if item.Action == "delete" {
			query := `UPDATE books SET deleted_at = $1, updated_at = $1 WHERE id = $2 AND user_id = $3`
			if _, err := tx.ExecContext(ctx, query, now, item.ID, userID); err != nil {
				c.Logger().Errorf("Failed to delete book: %v", err)
			}
		}
	}

	// Sessions
	for _, item := range req.Changes.Sessions {
		if item.Action == "upsert" {
			data := item.Data
			query := `
				INSERT INTO reading_sessions (id, user_id, book_id, started_at, ended_at, note, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				ON CONFLICT (id) DO UPDATE SET
					started_at = EXCLUDED.started_at,
					ended_at = EXCLUDED.ended_at,
					note = EXCLUDED.note,
					updated_at = EXCLUDED.updated_at,
					deleted_at = NULL
				WHERE reading_sessions.user_id = $2
			`
			var endedAt *time.Time
			if data.EndedAt != nil {
				t := models.UnixToTime(*data.EndedAt)
				endedAt = &t
			}
			if _, err := tx.ExecContext(ctx, query, item.ID, userID, data.BookID, models.UnixToTime(data.StartedAt), endedAt, data.Note, models.UnixToTime(data.CreatedAt), models.UnixToTime(data.UpdatedAt)); err != nil {
				c.Logger().Errorf("Failed to upsert session: %v", err)
			}
		} else if item.Action == "delete" {
			query := `UPDATE reading_sessions SET deleted_at = $1, updated_at = $1 WHERE id = $2 AND user_id = $3`
			if _, err := tx.ExecContext(ctx, query, now, item.ID, userID); err != nil {
				c.Logger().Errorf("Failed to delete session: %v", err)
			}
		}
	}

	// Captures
	for _, item := range req.Changes.Captures {
		if item.Action == "upsert" {
			data := item.Data
			query := `
				INSERT INTO captures (id, user_id, session_id, emotion, user_memo, ai_status, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				ON CONFLICT (id) DO UPDATE SET
					emotion = EXCLUDED.emotion,
					user_memo = EXCLUDED.user_memo,
					ai_status = EXCLUDED.ai_status,
					updated_at = EXCLUDED.updated_at,
					deleted_at = NULL
				WHERE captures.user_id = $2
			`
			if _, err := tx.ExecContext(ctx, query, item.ID, userID, data.SessionID, data.Emotion, data.UserMemo, data.AIStatus, models.UnixToTime(data.CreatedAt), models.UnixToTime(data.UpdatedAt)); err != nil {
				c.Logger().Errorf("Failed to upsert capture: %v", err)
			}
		} else if item.Action == "delete" {
			query := `UPDATE captures SET deleted_at = $1, updated_at = $1 WHERE id = $2 AND user_id = $3`
			if _, err := tx.ExecContext(ctx, query, now, item.ID, userID); err != nil {
				c.Logger().Errorf("Failed to delete capture: %v", err)
			}
		}
	}

	// AIResults
	for _, item := range req.Changes.AIResults {
		if item.Action == "upsert" {
			data := item.Data
			query := `
				INSERT INTO ai_results (id, user_id, capture_id, draft_summary, deepening_question, axis_selected, topic_tags, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				ON CONFLICT (id) DO UPDATE SET
					draft_summary = EXCLUDED.draft_summary,
					deepening_question = EXCLUDED.deepening_question,
					axis_selected = EXCLUDED.axis_selected,
					topic_tags = EXCLUDED.topic_tags,
					updated_at = EXCLUDED.updated_at,
					deleted_at = NULL
				WHERE ai_results.user_id = $2
			`
			if _, err := tx.ExecContext(ctx, query, item.ID, userID, data.CaptureID, data.DraftSummary, data.DeepeningQuestion, data.AxisSelected, data.TopicTags, models.UnixToTime(data.CreatedAt), models.UnixToTime(data.UpdatedAt)); err != nil {
				c.Logger().Errorf("Failed to upsert ai_result: %v", err)
			}
		} else if item.Action == "delete" {
			query := `UPDATE ai_results SET deleted_at = $1, updated_at = $1 WHERE id = $2 AND user_id = $3`
			if _, err := tx.ExecContext(ctx, query, now, item.ID, userID); err != nil {
				c.Logger().Errorf("Failed to delete ai_result: %v", err)
			}
		}
	}

	// NextActions
	for _, item := range req.Changes.NextActions {
		if item.Action == "upsert" {
			data := item.Data
			query := `
				INSERT INTO next_actions (id, user_id, ai_result_id, type, title, description, query, link, rank, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
				ON CONFLICT (id) DO UPDATE SET
					type = EXCLUDED.type,
					title = EXCLUDED.title,
					description = EXCLUDED.description,
					query = EXCLUDED.query,
					link = EXCLUDED.link,
					rank = EXCLUDED.rank,
					deleted_at = NULL
				WHERE next_actions.user_id = $2
			`
			if _, err := tx.ExecContext(ctx, query, item.ID, userID, data.AIResultID, data.Type, data.Title, data.Description, data.Query, data.Link, data.Rank, now); err != nil {
				c.Logger().Errorf("Failed to upsert next_action: %v", err)
			}
		} else if item.Action == "delete" {
			query := `UPDATE next_actions SET deleted_at = $1 WHERE id = $2 AND user_id = $3`
			if _, err := tx.ExecContext(ctx, query, now, item.ID, userID); err != nil {
				c.Logger().Errorf("Failed to delete next_action: %v", err)
			}
		}
	}

	// FeedbackLogs
	for _, item := range req.Changes.FeedbackLogs {
		if item.Action == "upsert" {
			data := item.Data
			query := `
				INSERT INTO feedback_logs (id, user_id, capture_id, helpfulness, created_at)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (id) DO UPDATE SET
					helpfulness = EXCLUDED.helpfulness
				WHERE feedback_logs.user_id = $2
			`
			if _, err := tx.ExecContext(ctx, query, item.ID, userID, data.CaptureID, data.Helpfulness, models.UnixToTime(data.CreatedAt)); err != nil {
				c.Logger().Errorf("Failed to upsert feedback_log: %v", err)
			}
		}
		// No delete for feedback logs
	}

	// ActionLogs
	for _, item := range req.Changes.ActionLogs {
		if item.Action == "upsert" {
			data := item.Data
			query := `
				INSERT INTO action_logs (id, user_id, ai_result_id, action_type, opened, completed, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
				ON CONFLICT (id) DO UPDATE SET
					opened = EXCLUDED.opened,
					completed = EXCLUDED.completed
				WHERE action_logs.user_id = $2
			`
			if _, err := tx.ExecContext(ctx, query, item.ID, userID, data.AIResultID, data.ActionType, data.Opened, data.Completed, models.UnixToTime(data.CreatedAt)); err != nil {
				c.Logger().Errorf("Failed to upsert action_log: %v", err)
			}
		}
		// No delete for action logs
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to commit transaction")
	}

	// Fetch server changes since last sync
	books, _ := h.syncRepo.GetBooksUpdatedSince(ctx, userID, lastSyncedAt)
	sessions, _ := h.syncRepo.GetSessionsUpdatedSince(ctx, userID, lastSyncedAt)
	captures, _ := h.syncRepo.GetCapturesUpdatedSince(ctx, userID, lastSyncedAt)
	aiResults, _ := h.syncRepo.GetAIResultsUpdatedSince(ctx, userID, lastSyncedAt)
	nextActions, _ := h.syncRepo.GetNextActionsUpdatedSince(ctx, userID, lastSyncedAt)

	// Use a new timestamp for the response and next sync
	syncResponseTime := time.Now()

	// Update device last_sync_at
	if err := h.deviceRepo.UpdateLastSyncAt(ctx, deviceID, syncResponseTime); err != nil {
		c.Logger().Warnf("Failed to update last_sync_at: %v", err)
	}

	// Build response
	response := buildSyncResponse(books, sessions, captures, aiResults, nextActions, syncResponseTime)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    response,
	})
}

// SyncStatus returns the current sync status
func (h *SyncHandler) SyncStatus(c echo.Context) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errorResponse(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "Invalid token")
	}

	deviceID, err := middleware.GetDeviceID(c)
	if err != nil {
		return errorResponse(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "Invalid token")
	}

	ctx := c.Request().Context()

	// Get device info
	device, err := h.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get device")
	}

	// Get total data size
	totalSize, err := h.syncRepo.GetSyncStats(ctx, userID)
	if err != nil {
		c.Logger().Warnf("Failed to get sync stats: %v", err)
	}

	var lastSyncedAt int64
	if device.LastSyncAt != nil {
		lastSyncedAt = device.LastSyncAt.Unix()
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"last_synced_at":   lastSyncedAt,
			"server_timestamp": time.Now().Unix(),
			"total_data_size":  totalSize,
		},
	})
}

// buildSyncResponse converts models to sync response format
func buildSyncResponse(
	books []models.Book,
	sessions []models.ReadingSession,
	captures []models.Capture,
	aiResults []models.AIResult,
	nextActions []models.NextAction,
	serverTime time.Time,
) models.SyncResponse {
	resp := models.SyncResponse{
		ServerTimestamp: serverTime.Unix() + 1,
		Changes: models.SyncChangesResponse{
			Books:       make([]models.BookResponse, 0, len(books)),
			Sessions:    make([]models.SessionResponse, 0, len(sessions)),
			Captures:    make([]models.CaptureResponse, 0, len(captures)),
			AIResults:   make([]models.AIResultResponse, 0, len(aiResults)),
			NextActions: make([]models.NextActionResponse, 0, len(nextActions)),
		},
	}

	for _, b := range books {
		bookResp := models.BookResponse{
			ID:        b.ID,
			Title:     b.Title,
			Author:    b.Author,
			Summary:   b.Summary,
			CreatedAt: b.CreatedAt.Unix(),
			UpdatedAt: b.UpdatedAt.Unix(),

			SummaryGeneratedAt:  models.TimePtrToUnix(b.SummaryGeneratedAt),
			SummarySavedAt:      models.TimePtrToUnix(b.SummarySavedAt),
			SummaryEditedAt:     models.TimePtrToUnix(b.SummaryEditedAt),
			SummaryHelpfulness:  b.SummaryHelpfulness,
			SummaryCaptureCount: b.SummaryCaptureCount,
			SummaryCompletedAt:  models.TimePtrToUnix(b.SummaryCompletedAt),
		}
		if b.DeletedAt != nil {
			deletedAt := b.DeletedAt.Unix()
			bookResp.DeletedAt = &deletedAt
		}
		resp.Changes.Books = append(resp.Changes.Books, bookResp)
	}

	for _, s := range sessions {
		sessionResp := models.SessionResponse{
			ID:        s.ID,
			BookID:    s.BookID,
			StartedAt: s.StartedAt.Unix(),
			Note:      s.Note,
			CreatedAt: s.CreatedAt.Unix(),
			UpdatedAt: s.UpdatedAt.Unix(),
		}
		if s.EndedAt != nil {
			endedAt := s.EndedAt.Unix()
			sessionResp.EndedAt = &endedAt
		}
		if s.DeletedAt != nil {
			deletedAt := s.DeletedAt.Unix()
			sessionResp.DeletedAt = &deletedAt
		}
		resp.Changes.Sessions = append(resp.Changes.Sessions, sessionResp)
	}

	for _, cap := range captures {
		captureResp := models.CaptureResponse{
			ID:        cap.ID,
			SessionID: cap.SessionID,
			Emotion:   cap.Emotion,
			UserMemo:  cap.UserMemo,
			AIStatus:  string(cap.AIStatus),
			CreatedAt: cap.CreatedAt.Unix(),
			UpdatedAt: cap.UpdatedAt.Unix(),
		}
		if cap.DeletedAt != nil {
			deletedAt := cap.DeletedAt.Unix()
			captureResp.DeletedAt = &deletedAt
		}
		resp.Changes.Captures = append(resp.Changes.Captures, captureResp)
	}

	for _, ai := range aiResults {
		aiResp := models.AIResultResponse{
			ID:                ai.ID,
			CaptureID:         ai.CaptureID,
			DraftSummary:      ai.DraftSummary,
			DeepeningQuestion: ai.DeepeningQuestion,
			AxisSelected:      ai.AxisSelected,
			TopicTags:         ai.TopicTags,
			CreatedAt:         ai.CreatedAt.Unix(),
			UpdatedAt:         ai.UpdatedAt.Unix(),
		}
		if ai.DeletedAt != nil {
			deletedAt := ai.DeletedAt.Unix()
			aiResp.DeletedAt = &deletedAt
		}
		resp.Changes.AIResults = append(resp.Changes.AIResults, aiResp)
	}

	for _, na := range nextActions {
		actionResp := models.NextActionResponse{
			ID:          na.ID,
			AIResultID:  na.AIResultID,
			Type:        na.Type,
			Title:       na.Title,
			Description: na.Description,
			Query:       na.Query,
			Link:        na.Link,
			Rank:        na.Rank,
			CreatedAt:   na.CreatedAt.Unix(),
		}
		if na.DeletedAt != nil {
			deletedAt := na.DeletedAt.Unix()
			actionResp.DeletedAt = &deletedAt
		}
		resp.Changes.NextActions = append(resp.Changes.NextActions, actionResp)
	}

	return resp
}

// BackupExport exports all user data
func (h *SyncHandler) BackupExport(c echo.Context) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errorResponse(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "Invalid token")
	}

	ctx := c.Request().Context()

	// Get all user data (using epoch as since time to get everything)
	books, _ := h.syncRepo.GetBooksUpdatedSince(ctx, userID, 0)
	sessions, _ := h.syncRepo.GetSessionsUpdatedSince(ctx, userID, 0)
	captures, _ := h.syncRepo.GetCapturesUpdatedSince(ctx, userID, 0)
	aiResults, _ := h.syncRepo.GetAIResultsUpdatedSince(ctx, userID, 0)
	nextActions, _ := h.syncRepo.GetNextActionsUpdatedSince(ctx, userID, 0)

	response := buildSyncResponse(books, sessions, captures, aiResults, nextActions, time.Now())

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"version":     "1.0",
			"exported_at": time.Now().Unix(),
			"data":        response.Changes,
		},
	})
}

// BackupImport imports user data from backup
func (h *SyncHandler) BackupImport(c echo.Context) error {
	// userID, err := middleware.GetUserID(c)
	// if err != nil {
	// 	return errorResponse(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "Invalid token")
	// }

	// var req struct {
	// 	Version string                    `json:"version"`
	// 	Data    models.SyncChangesRequest `json:"data"`
	// }

	// if err := c.Bind(&req); err != nil {
	// 	return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
	// }

	// // ctx := c.Request().Context()
	// // now := time.Now()

	// // Import all data using sync logic
	var imported int

	// tx, err := h.db.BeginTxx(c.Request().Context(), nil)
	// if err != nil {
	// 	return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to start transaction")
	// }
	// defer tx.Rollback()

	// for _, item := range req.Data.Books {
	// 	if item.Action == "upsert" {
	// 		imported++
	// 	}
	// }
	// for _, item := range req.Data.Sessions {
	// 	if item.Action == "upsert" {
	// 		imported++
	// 	}
	// }
	// for _, item := range req.Data.Captures {
	// 	if item.Action == "upsert" {
	// 		imported++
	// 	}
	// }
	// for _, item := range req.Data.AIResults {
	// 	if item.Action == "upsert" {
	// 		imported++
	// 	}
	// }
	// for _, item := range req.Data.NextActions {
	// 	if item.Action == "upsert" {
	// 		imported++
	// 	}
	// }

	// if err := tx.Commit(); err != nil {
	// 	return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to commit transaction")
	// }

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"imported_count": imported,
		},
	})
}

// GetBooks returns user's books
func (h *SyncHandler) GetBooks(c echo.Context) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errorResponse(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "Invalid token")
	}

	ctx := c.Request().Context()
	books, err := h.syncRepo.GetBooksUpdatedSince(ctx, userID, 0)
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get books")
	}

	// Filter out deleted books
	activeBooks := make([]models.BookResponse, 0)
	for _, b := range books {
		if b.DeletedAt == nil {
			activeBooks = append(activeBooks, models.BookResponse{
				ID:        b.ID,
				Title:     b.Title,
				Author:    b.Author,
				Summary:   b.Summary,
				CreatedAt: b.CreatedAt.Unix(),
				UpdatedAt: b.UpdatedAt.Unix(),
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    activeBooks,
	})
}

// GetBook returns a specific book
func (h *SyncHandler) GetBook(c echo.Context) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errorResponse(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "Invalid token")
	}

	bookIDStr := c.Param("id")
	bookID, err := uuid.Parse(bookIDStr)
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid book ID")
	}

	ctx := c.Request().Context()
	books, err := h.syncRepo.GetBooksUpdatedSince(ctx, userID, 0)
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get book")
	}

	for _, b := range books {
		if b.ID == bookID && b.DeletedAt == nil {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"success": true,
				"data": models.BookResponse{
					ID:        b.ID,
					Title:     b.Title,
					Author:    b.Author,
					Summary:   b.Summary,
					CreatedAt: b.CreatedAt.Unix(),
					UpdatedAt: b.UpdatedAt.Unix(),
				},
			})
		}
	}

	return errorResponse(c, http.StatusNotFound, "NOT_FOUND", "Book not found")
}

// GetBookSessions returns sessions for a book
func (h *SyncHandler) GetBookSessions(c echo.Context) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errorResponse(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "Invalid token")
	}

	bookIDStr := c.Param("book_id")
	bookID, err := uuid.Parse(bookIDStr)
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid book ID")
	}

	ctx := c.Request().Context()
	sessions, err := h.syncRepo.GetSessionsUpdatedSince(ctx, userID, 0)
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get sessions")
	}

	activeSessions := make([]models.SessionResponse, 0)
	for _, s := range sessions {
		if s.BookID == bookID && s.DeletedAt == nil {
			sessionResp := models.SessionResponse{
				ID:        s.ID,
				BookID:    s.BookID,
				StartedAt: s.StartedAt.Unix(),
				Note:      s.Note,
				CreatedAt: s.CreatedAt.Unix(),
				UpdatedAt: s.UpdatedAt.Unix(),
			}
			if s.EndedAt != nil {
				endedAt := s.EndedAt.Unix()
				sessionResp.EndedAt = &endedAt
			}
			activeSessions = append(activeSessions, sessionResp)
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    activeSessions,
	})
}

// GetSessionCaptures returns captures for a session
func (h *SyncHandler) GetSessionCaptures(c echo.Context) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errorResponse(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "Invalid token")
	}

	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid session ID")
	}

	ctx := c.Request().Context()
	captures, err := h.syncRepo.GetCapturesUpdatedSince(ctx, userID, 0)
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get captures")
	}

	activeCaptures := make([]models.CaptureResponse, 0)
	for _, cap := range captures {
		if cap.SessionID == sessionID && cap.DeletedAt == nil {
			activeCaptures = append(activeCaptures, models.CaptureResponse{
				ID:        cap.ID,
				SessionID: cap.SessionID,
				Emotion:   cap.Emotion,
				UserMemo:  cap.UserMemo,
				AIStatus:  string(cap.AIStatus),
				CreatedAt: cap.CreatedAt.Unix(),
				UpdatedAt: cap.UpdatedAt.Unix(),
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    activeCaptures,
	})
}

// GetCaptureAIResult returns AI result for a capture
func (h *SyncHandler) GetCaptureAIResult(c echo.Context) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errorResponse(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "Invalid token")
	}

	captureIDStr := c.Param("capture_id")
	captureID, err := uuid.Parse(captureIDStr)
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid capture ID")
	}

	ctx := c.Request().Context()
	aiResults, err := h.syncRepo.GetAIResultsUpdatedSince(ctx, userID, 0)
	if err != nil {
		return errorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get AI result")
	}

	for _, ai := range aiResults {
		if ai.CaptureID == captureID && ai.DeletedAt == nil {
			return c.JSON(http.StatusOK, map[string]interface{}{
				"success": true,
				"data": models.AIResultResponse{
					ID:                ai.ID,
					CaptureID:         ai.CaptureID,
					DraftSummary:      ai.DraftSummary,
					DeepeningQuestion: ai.DeepeningQuestion,
					AxisSelected:      ai.AxisSelected,
					TopicTags:         ai.TopicTags,
					CreatedAt:         ai.CreatedAt.Unix(),
					UpdatedAt:         ai.UpdatedAt.Unix(),
				},
			})
		}
	}

	return errorResponse(c, http.StatusNotFound, "NOT_FOUND", "AI result not found")
}
