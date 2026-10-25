package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// AnalyticsRepository handles analytics data operations
type AnalyticsRepository struct {
	db *sqlx.DB
}

// NewAnalyticsRepository creates a new AnalyticsRepository
func NewAnalyticsRepository(db *sqlx.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

// AxisDistribution represents axis selection distribution
type AxisDistribution struct {
	AxisSelected string `db:"axis_selected"`
	Count        int    `db:"count"`
}

// ActionStats represents action statistics
type ActionStats struct {
	Type      string `db:"type"`
	Shown     int    `db:"shown"`
	Opened    int    `db:"opened"`
	Completed int    `db:"completed"`
}

// FeedbackStats represents feedback statistics
type FeedbackStats struct {
	Average float64 `db:"average"`
	Total   int     `db:"total"`
	Count1  int     `db:"count_1"`
	Count2  int     `db:"count_2"`
	Count3  int     `db:"count_3"`
}

// GetPersonalAxisDistribution returns axis distribution for a user
func (r *AnalyticsRepository) GetPersonalAxisDistribution(ctx context.Context, userID uuid.UUID, since time.Time) ([]AxisDistribution, error) {
	var results []AxisDistribution
	query := `
		SELECT axis_selected, COUNT(*) as count
		FROM ai_results
		WHERE user_id = $1
		  AND axis_selected IS NOT NULL
		  AND created_at >= $2
		  AND deleted_at IS NULL
		GROUP BY axis_selected
	`
	err := r.db.SelectContext(ctx, &results, query, userID, since)
	return results, err
}

// GetPersonalActionStats returns action statistics for a user
func (r *AnalyticsRepository) GetPersonalActionStats(ctx context.Context, userID uuid.UUID, since time.Time) ([]ActionStats, error) {
	var results []ActionStats
	query := `
		SELECT
			na.type,
			COUNT(DISTINCT na.id) as shown,
			COUNT(DISTINCT CASE WHEN al.opened THEN al.id END) as opened,
			COUNT(DISTINCT CASE WHEN al.completed THEN al.id END) as completed
		FROM next_actions na
		LEFT JOIN action_logs al ON na.ai_result_id = al.ai_result_id
		                         AND na.type = al.action_type
		WHERE na.user_id = $1
		  AND na.created_at >= $2
		  AND na.deleted_at IS NULL
		GROUP BY na.type
	`
	err := r.db.SelectContext(ctx, &results, query, userID, since)
	return results, err
}

// GetPersonalFeedbackStats returns feedback statistics for a user
func (r *AnalyticsRepository) GetPersonalFeedbackStats(ctx context.Context, userID uuid.UUID, since time.Time) (*FeedbackStats, error) {
	var result FeedbackStats
	query := `
		SELECT
			COALESCE(AVG(helpfulness), 0) as average,
			COUNT(*) as total,
			COUNT(CASE WHEN helpfulness = 1 THEN 1 END) as count_1,
			COUNT(CASE WHEN helpfulness = 2 THEN 1 END) as count_2,
			COUNT(CASE WHEN helpfulness = 3 THEN 1 END) as count_3
		FROM feedback_logs
		WHERE user_id = $1
		  AND created_at >= $2
	`
	err := r.db.GetContext(ctx, &result, query, userID, since)
	return &result, err
}

// GlobalStats represents global statistics
type GlobalStats struct {
	TotalUsers  int `db:"total_users"`
	ActiveUsers int `db:"active_users"`
}

// GetGlobalUserStats returns global user statistics
func (r *AnalyticsRepository) GetGlobalUserStats(ctx context.Context, since time.Time) (*GlobalStats, error) {
	var result GlobalStats
	query := `
		SELECT
			COUNT(DISTINCT id) as total_users,
			COUNT(DISTINCT CASE WHEN updated_at >= $1 THEN id END) as active_users
		FROM users
		WHERE deleted_at IS NULL
	`
	err := r.db.GetContext(ctx, &result, query, since)
	return &result, err
}

// GetGlobalAxisDistribution returns global axis distribution
func (r *AnalyticsRepository) GetGlobalAxisDistribution(ctx context.Context, since time.Time) ([]AxisDistribution, error) {
	var results []AxisDistribution
	query := `
		SELECT axis_selected, COUNT(*) as count
		FROM ai_results
		WHERE axis_selected IS NOT NULL
		  AND created_at >= $1
		  AND deleted_at IS NULL
		GROUP BY axis_selected
	`
	err := r.db.SelectContext(ctx, &results, query, since)
	return results, err
}

// GetGlobalActionStats returns global action statistics
func (r *AnalyticsRepository) GetGlobalActionStats(ctx context.Context, since time.Time) ([]ActionStats, error) {
	var results []ActionStats
	query := `
		SELECT
			na.type,
			COUNT(DISTINCT na.id) as shown,
			COUNT(DISTINCT CASE WHEN al.opened THEN al.id END) as opened,
			COUNT(DISTINCT CASE WHEN al.completed THEN al.id END) as completed
		FROM next_actions na
		LEFT JOIN action_logs al ON na.ai_result_id = al.ai_result_id
		                         AND na.type = al.action_type
		WHERE na.created_at >= $1
		  AND na.deleted_at IS NULL
		GROUP BY na.type
	`
	err := r.db.SelectContext(ctx, &results, query, since)
	return results, err
}

// GetGlobalFeedbackStats returns global feedback statistics
func (r *AnalyticsRepository) GetGlobalFeedbackStats(ctx context.Context, since time.Time) (*FeedbackStats, error) {
	var result FeedbackStats
	query := `
		SELECT
			COALESCE(AVG(helpfulness), 0) as average,
			COUNT(*) as total,
			COUNT(CASE WHEN helpfulness = 1 THEN 1 END) as count_1,
			COUNT(CASE WHEN helpfulness = 2 THEN 1 END) as count_2,
			COUNT(CASE WHEN helpfulness = 3 THEN 1 END) as count_3
		FROM feedback_logs
		WHERE created_at >= $1
	`
	err := r.db.GetContext(ctx, &result, query, since)
	return &result, err
}

// TrendDataPoint represents a data point in trend data
type TrendDataPoint struct {
	Date       string         `db:"date"`
	AxisValues map[string]int `db:"-"`
}

// AxisTrendRow represents a row from axis trend query
type AxisTrendRow struct {
	Date         string `db:"date"`
	AxisSelected string `db:"axis_selected"`
	Count        int    `db:"count"`
}

// GetPersonalAxisTrend returns axis trend for a user
func (r *AnalyticsRepository) GetPersonalAxisTrend(ctx context.Context, userID uuid.UUID, since time.Time) ([]AxisTrendRow, error) {
	var results []AxisTrendRow
	query := `
		SELECT
			DATE(created_at) as date,
			axis_selected,
			COUNT(*) as count
		FROM ai_results
		WHERE user_id = $1
		  AND axis_selected IS NOT NULL
		  AND created_at >= $2
		  AND deleted_at IS NULL
		GROUP BY DATE(created_at), axis_selected
		ORDER BY date
	`
	err := r.db.SelectContext(ctx, &results, query, userID, since)
	return results, err
}

// GetGlobalAxisTrend returns global axis trend
func (r *AnalyticsRepository) GetGlobalAxisTrend(ctx context.Context, since time.Time) ([]AxisTrendRow, error) {
	var results []AxisTrendRow
	query := `
		SELECT
			DATE(created_at) as date,
			axis_selected,
			COUNT(*) as count
		FROM ai_results
		WHERE axis_selected IS NOT NULL
		  AND created_at >= $1
		  AND deleted_at IS NULL
		GROUP BY DATE(created_at), axis_selected
		ORDER BY date
	`
	err := r.db.SelectContext(ctx, &results, query, since)
	return results, err
}

// ActionTrendRow represents a row from action trend query
type ActionTrendRow struct {
	Date      string `db:"date"`
	Type      string `db:"type"`
	Opened    int    `db:"opened"`
	Completed int    `db:"completed"`
}

// GetPersonalActionTrend returns action trend for a user
func (r *AnalyticsRepository) GetPersonalActionTrend(ctx context.Context, userID uuid.UUID, since time.Time) ([]ActionTrendRow, error) {
	var results []ActionTrendRow
	query := `
		SELECT
			DATE(al.created_at) as date,
			al.action_type as type,
			COUNT(CASE WHEN al.opened THEN 1 END) as opened,
			COUNT(CASE WHEN al.completed THEN 1 END) as completed
		FROM action_logs al
		WHERE al.user_id = $1
		  AND al.created_at >= $2
		GROUP BY DATE(al.created_at), al.action_type
		ORDER BY date
	`
	err := r.db.SelectContext(ctx, &results, query, userID, since)
	return results, err
}

// GetGlobalActionTrend returns global action trend
func (r *AnalyticsRepository) GetGlobalActionTrend(ctx context.Context, since time.Time) ([]ActionTrendRow, error) {
	var results []ActionTrendRow
	query := `
		SELECT
			DATE(al.created_at) as date,
			al.action_type as type,
			COUNT(CASE WHEN al.opened THEN 1 END) as opened,
			COUNT(CASE WHEN al.completed THEN 1 END) as completed
		FROM action_logs al
		WHERE al.created_at >= $1
		GROUP BY DATE(al.created_at), al.action_type
		ORDER BY date
	`
	err := r.db.SelectContext(ctx, &results, query, since)
	return results, err
}

// FeedbackTrendRow represents a row from feedback trend query
type FeedbackTrendRow struct {
	Date    string  `db:"date"`
	Average float64 `db:"average"`
	Count   int     `db:"count"`
}

// GetPersonalFeedbackTrend returns feedback trend for a user
func (r *AnalyticsRepository) GetPersonalFeedbackTrend(ctx context.Context, userID uuid.UUID, since time.Time) ([]FeedbackTrendRow, error) {
	var results []FeedbackTrendRow
	query := `
		SELECT
			DATE(created_at) as date,
			AVG(helpfulness) as average,
			COUNT(*) as count
		FROM feedback_logs
		WHERE user_id = $1
		  AND created_at >= $2
		GROUP BY DATE(created_at)
		ORDER BY date
	`
	err := r.db.SelectContext(ctx, &results, query, userID, since)
	return results, err
}

// GetGlobalFeedbackTrend returns global feedback trend
func (r *AnalyticsRepository) GetGlobalFeedbackTrend(ctx context.Context, since time.Time) ([]FeedbackTrendRow, error) {
	var results []FeedbackTrendRow
	query := `
		SELECT
			DATE(created_at) as date,
			AVG(helpfulness) as average,
			COUNT(*) as count
		FROM feedback_logs
		WHERE created_at >= $1
		GROUP BY DATE(created_at)
		ORDER BY date
	`
	err := r.db.SelectContext(ctx, &results, query, since)
	return results, err
}

// NorthStarMonth is the North Star Metric for one month:
// the number of books whose "自分のまとめ" was first completed in that month.
type NorthStarMonth struct {
	Month              time.Time `db:"month"`
	CompletedSummaries int       `db:"completed_summaries"`
}

// GetNorthStarByMonth returns completed summaries per month since the given time.
// Pass uuid.Nil as userID for all users.
func (r *AnalyticsRepository) GetNorthStarByMonth(ctx context.Context, userID uuid.UUID, since time.Time) ([]NorthStarMonth, error) {
	var rows []NorthStarMonth
	query := `
		SELECT date_trunc('month', summary_completed_at AT TIME ZONE 'UTC') AS month,
		       COUNT(*) AS completed_summaries
		FROM books
		WHERE summary_completed_at IS NOT NULL
		  AND summary_completed_at >= $1
		  AND deleted_at IS NULL
		  AND ($2 = '00000000-0000-0000-0000-000000000000'::uuid OR user_id = $2)
		GROUP BY month
		ORDER BY month
	`
	err := r.db.SelectContext(ctx, &rows, query, since, userID)
	return rows, err
}

// NorthStarInputs are the input metrics that drive the North Star Metric.
type NorthStarInputs struct {
	// Books with at least one capture in the period
	ActiveBooks int `db:"active_books"`
	// Average captures per active book
	CapturesPerBook float64 `db:"captures_per_book"`
	// Active books that have both a doubt (🤔) and an empathy (❤️) capture
	BooksWithDoubtAndEmpathy int `db:"books_with_doubt_and_empathy"`
	// Active books with at least SummaryMinCaptures captures
	BooksReadyForSummary int `db:"books_ready_for_summary"`
	// Of those, books that have a generated summary
	BooksWithSummary int `db:"books_with_summary"`
	// Of those, books whose summary the reader edited
	BooksWithEditedSummary int `db:"books_with_edited_summary"`
	// Average helpfulness of rated summaries (0 when none)
	SummaryHelpfulnessAvg float64 `db:"summary_helpfulness_avg"`
}

// GetNorthStarInputs returns the input metrics for books with captures since the given time.
// Pass uuid.Nil as userID for all users.
func (r *AnalyticsRepository) GetNorthStarInputs(ctx context.Context, userID uuid.UUID, since time.Time, minCaptures int) (*NorthStarInputs, error) {
	var inputs NorthStarInputs
	query := `
		WITH book_captures AS (
			SELECT b.id AS book_id,
			       COUNT(c.id) AS captures,
			       BOOL_OR(c.emotion LIKE '%🤔%') AS has_doubt,
			       BOOL_OR(c.emotion LIKE '%❤️%') AS has_empathy,
			       MAX(b.summary_generated_at) AS generated_at,
			       MAX(b.summary_edited_at) AS edited_at,
			       MAX(b.summary_helpfulness) AS helpfulness
			FROM books b
			JOIN reading_sessions s ON s.book_id = b.id AND s.deleted_at IS NULL
			JOIN captures c ON c.session_id = s.id AND c.deleted_at IS NULL
			WHERE b.deleted_at IS NULL
			  AND c.created_at >= $1
			  AND ($2 = '00000000-0000-0000-0000-000000000000'::uuid OR b.user_id = $2)
			GROUP BY b.id
		)
		SELECT COUNT(*) AS active_books,
		       COALESCE(AVG(captures), 0) AS captures_per_book,
		       COUNT(*) FILTER (WHERE has_doubt AND has_empathy) AS books_with_doubt_and_empathy,
		       COUNT(*) FILTER (WHERE captures >= $3) AS books_ready_for_summary,
		       COUNT(*) FILTER (WHERE captures >= $3 AND generated_at IS NOT NULL) AS books_with_summary,
		       COUNT(*) FILTER (WHERE captures >= $3 AND edited_at IS NOT NULL) AS books_with_edited_summary,
		       COALESCE(AVG(helpfulness) FILTER (WHERE helpfulness IS NOT NULL), 0) AS summary_helpfulness_avg
		FROM book_captures
	`
	err := r.db.GetContext(ctx, &inputs, query, since, userID, minCaptures)
	return &inputs, err
}
