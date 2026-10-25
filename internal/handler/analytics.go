package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	"github.com/suiten/suiten-server/internal/middleware"
	"github.com/suiten/suiten-server/internal/repository"
)

// AnalyticsHandler handles analytics endpoints
type AnalyticsHandler struct {
	db            *sqlx.DB
	analyticsRepo *repository.AnalyticsRepository
}

// NewAnalyticsHandler creates a new AnalyticsHandler
func NewAnalyticsHandler(db *sqlx.DB) *AnalyticsHandler {
	return &AnalyticsHandler{
		db:            db,
		analyticsRepo: repository.NewAnalyticsRepository(db),
	}
}

// parsePeriod converts period string to time.Time
func parsePeriod(period string) time.Time {
	now := time.Now()
	switch period {
	case "7d":
		return now.AddDate(0, 0, -7)
	case "30d":
		return now.AddDate(0, 0, -30)
	case "90d":
		return now.AddDate(0, 0, -90)
	case "all":
		return time.Time{}
	default:
		return now.AddDate(0, 0, -30)
	}
}

var axisLabels = map[string]string{
	"risk":         "リスク",
	"success_rate": "成功確率",
	"rationale":    "根拠",
	"concept":      "概念",
}

var actionLabels = map[string]string{
	"dialog": "AI対話",
	"books":  "関連書籍",
	"videos": "関連動画",
}

// PersonalAnalytics returns personal analytics for the authenticated user
func (h *AnalyticsHandler) PersonalAnalytics(c echo.Context) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errorResponse(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "Invalid token")
	}

	period := c.QueryParam("period")
	if period == "" {
		period = "30d"
	}
	since := parsePeriod(period)

	ctx := c.Request().Context()

	// Get axis distribution
	axisDist, err := h.analyticsRepo.GetPersonalAxisDistribution(ctx, userID, since)
	if err != nil {
		c.Logger().Errorf("Failed to get axis distribution: %v", err)
	}

	axisDistribution := map[string]int{
		"risk":         0,
		"success_rate": 0,
		"rationale":    0,
		"concept":      0,
		"total":        0,
	}
	for _, ad := range axisDist {
		axisDistribution[ad.AxisSelected] = ad.Count
		axisDistribution["total"] += ad.Count
	}

	// Get action stats
	actionStatsData, err := h.analyticsRepo.GetPersonalActionStats(ctx, userID, since)
	if err != nil {
		c.Logger().Errorf("Failed to get action stats: %v", err)
	}

	actionStats := make(map[string]map[string]interface{})
	for _, actionType := range []string{"dialog", "books", "videos"} {
		actionStats[actionType] = map[string]interface{}{
			"shown":           0,
			"opened":          0,
			"completed":       0,
			"open_rate":       0.0,
			"completion_rate": 0.0,
		}
	}
	for _, as := range actionStatsData {
		stats := actionStats[as.Type]
		stats["shown"] = as.Shown
		stats["opened"] = as.Opened
		stats["completed"] = as.Completed
		if as.Shown > 0 {
			stats["open_rate"] = float64(as.Opened) / float64(as.Shown)
		}
		if as.Opened > 0 {
			stats["completion_rate"] = float64(as.Completed) / float64(as.Opened)
		}
	}

	// Get feedback stats
	feedbackStatsData, err := h.analyticsRepo.GetPersonalFeedbackStats(ctx, userID, since)
	if err != nil {
		c.Logger().Errorf("Failed to get feedback stats: %v", err)
		feedbackStatsData = &repository.FeedbackStats{}
	}

	feedbackStats := map[string]interface{}{
		"total_count":         feedbackStatsData.Total,
		"average_helpfulness": feedbackStatsData.Average,
		"distribution": map[string]int{
			"1": feedbackStatsData.Count1,
			"2": feedbackStatsData.Count2,
			"3": feedbackStatsData.Count3,
		},
	}

	// Generate insights
	insights := generateInsights(axisDistribution, actionStats)

	northStar, err := h.buildNorthStar(ctx, userID, time.Now(), since)
	if err != nil {
		c.Logger().Errorf("Failed to get north star metric: %v", err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"north_star":        northStar,
			"period":            period,
			"axis_distribution": axisDistribution,
			"action_stats":      actionStats,
			"feedback_stats":    feedbackStats,
			"insights":          insights,
		},
	})
}

// GlobalAnalytics returns anonymized global analytics
func (h *AnalyticsHandler) GlobalAnalytics(c echo.Context) error {
	_, err := middleware.GetUserID(c)
	if err != nil {
		return errorResponse(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "Invalid token")
	}

	period := "30d"
	since := parsePeriod(period)

	ctx := c.Request().Context()

	// Get user stats
	userStats, err := h.analyticsRepo.GetGlobalUserStats(ctx, since)
	if err != nil {
		c.Logger().Errorf("Failed to get user stats: %v", err)
		userStats = &repository.GlobalStats{}
	}

	// Get axis distribution
	axisDist, err := h.analyticsRepo.GetGlobalAxisDistribution(ctx, since)
	if err != nil {
		c.Logger().Errorf("Failed to get axis distribution: %v", err)
	}

	total := 0
	for _, ad := range axisDist {
		total += ad.Count
	}

	axisDistribution := map[string]float64{
		"risk":         0.0,
		"success_rate": 0.0,
		"rationale":    0.0,
		"concept":      0.0,
	}
	if total > 0 {
		for _, ad := range axisDist {
			axisDistribution[ad.AxisSelected] = float64(ad.Count) / float64(total)
		}
	}

	// Get action stats
	actionStatsData, err := h.analyticsRepo.GetGlobalActionStats(ctx, since)
	if err != nil {
		c.Logger().Errorf("Failed to get action stats: %v", err)
	}

	actionOpenRates := map[string]float64{
		"dialog": 0.0,
		"books":  0.0,
		"videos": 0.0,
	}
	actionCompletionRates := map[string]float64{
		"dialog": 0.0,
		"books":  0.0,
		"videos": 0.0,
	}
	for _, as := range actionStatsData {
		if as.Shown > 0 {
			actionOpenRates[as.Type] = float64(as.Opened) / float64(as.Shown)
		}
		if as.Opened > 0 {
			actionCompletionRates[as.Type] = float64(as.Completed) / float64(as.Opened)
		}
	}

	// Get feedback stats
	feedbackStatsData, err := h.analyticsRepo.GetGlobalFeedbackStats(ctx, since)
	if err != nil {
		c.Logger().Errorf("Failed to get feedback stats: %v", err)
		feedbackStatsData = &repository.FeedbackStats{}
	}

	feedbackDistribution := map[string]float64{
		"1": 0.0,
		"2": 0.0,
		"3": 0.0,
	}
	if feedbackStatsData.Total > 0 {
		feedbackDistribution["1"] = float64(feedbackStatsData.Count1) / float64(feedbackStatsData.Total)
		feedbackDistribution["2"] = float64(feedbackStatsData.Count2) / float64(feedbackStatsData.Total)
		feedbackDistribution["3"] = float64(feedbackStatsData.Count3) / float64(feedbackStatsData.Total)
	}

	northStar, err := h.buildNorthStar(ctx, uuid.Nil, time.Now(), since)
	if err != nil {
		c.Logger().Errorf("Failed to get north star metric: %v", err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"north_star":              northStar,
			"period":                  period,
			"total_users":             userStats.TotalUsers,
			"active_users":            userStats.ActiveUsers,
			"axis_distribution":       axisDistribution,
			"action_open_rates":       actionOpenRates,
			"action_completion_rates": actionCompletionRates,
			"feedback_average":        feedbackStatsData.Average,
			"feedback_distribution":   feedbackDistribution,
		},
	})
}

// Trends returns trend data
func (h *AnalyticsHandler) Trends(c echo.Context) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return errorResponse(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "Invalid token")
	}

	metric := c.QueryParam("metric")
	if metric == "" {
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "metric parameter is required")
	}

	period := c.QueryParam("period")
	if period == "" {
		period = "30d"
	}
	since := parsePeriod(period)

	scope := c.QueryParam("scope")
	if scope == "" {
		scope = "personal"
	}

	ctx := c.Request().Context()

	var dataPoints []map[string]interface{}

	switch metric {
	case "axis":
		if scope == "personal" {
			rows, err := h.analyticsRepo.GetPersonalAxisTrend(ctx, userID, since)
			if err != nil {
				c.Logger().Errorf("Failed to get axis trend: %v", err)
			}
			dataPoints = aggregateAxisTrend(rows)
		} else {
			rows, err := h.analyticsRepo.GetGlobalAxisTrend(ctx, since)
			if err != nil {
				c.Logger().Errorf("Failed to get axis trend: %v", err)
			}
			dataPoints = aggregateAxisTrend(rows)
		}

	case "action":
		if scope == "personal" {
			rows, err := h.analyticsRepo.GetPersonalActionTrend(ctx, userID, since)
			if err != nil {
				c.Logger().Errorf("Failed to get action trend: %v", err)
			}
			dataPoints = aggregateActionTrend(rows)
		} else {
			rows, err := h.analyticsRepo.GetGlobalActionTrend(ctx, since)
			if err != nil {
				c.Logger().Errorf("Failed to get action trend: %v", err)
			}
			dataPoints = aggregateActionTrend(rows)
		}

	case "feedback":
		if scope == "personal" {
			rows, err := h.analyticsRepo.GetPersonalFeedbackTrend(ctx, userID, since)
			if err != nil {
				c.Logger().Errorf("Failed to get feedback trend: %v", err)
			}
			dataPoints = aggregateFeedbackTrend(rows)
		} else {
			rows, err := h.analyticsRepo.GetGlobalFeedbackTrend(ctx, since)
			if err != nil {
				c.Logger().Errorf("Failed to get feedback trend: %v", err)
			}
			dataPoints = aggregateFeedbackTrend(rows)
		}

	default:
		return errorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid metric. Use: axis, action, feedback")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"metric":      metric,
			"period":      period,
			"data_points": dataPoints,
		},
	})
}

// aggregateAxisTrend aggregates axis trend data by date
func aggregateAxisTrend(rows []repository.AxisTrendRow) []map[string]interface{} {
	dateMap := make(map[string]map[string]int)

	for _, row := range rows {
		if _, ok := dateMap[row.Date]; !ok {
			dateMap[row.Date] = map[string]int{
				"risk":         0,
				"success_rate": 0,
				"rationale":    0,
				"concept":      0,
			}
		}
		dateMap[row.Date][row.AxisSelected] = row.Count
	}

	result := make([]map[string]interface{}, 0, len(dateMap))
	for date, values := range dateMap {
		result = append(result, map[string]interface{}{
			"date":   date,
			"values": values,
		})
	}

	return result
}

// aggregateActionTrend aggregates action trend data by date
func aggregateActionTrend(rows []repository.ActionTrendRow) []map[string]interface{} {
	dateMap := make(map[string]map[string]map[string]int)

	for _, row := range rows {
		if _, ok := dateMap[row.Date]; !ok {
			dateMap[row.Date] = make(map[string]map[string]int)
		}
		dateMap[row.Date][row.Type] = map[string]int{
			"opened":    row.Opened,
			"completed": row.Completed,
		}
	}

	result := make([]map[string]interface{}, 0, len(dateMap))
	for date, values := range dateMap {
		result = append(result, map[string]interface{}{
			"date":   date,
			"values": values,
		})
	}

	return result
}

// aggregateFeedbackTrend aggregates feedback trend data
func aggregateFeedbackTrend(rows []repository.FeedbackTrendRow) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(rows))

	for _, row := range rows {
		result = append(result, map[string]interface{}{
			"date": row.Date,
			"values": map[string]interface{}{
				"average": row.Average,
				"count":   row.Count,
			},
		})
	}

	return result
}

// generateInsights generates insights based on personal stats
func generateInsights(axisDistribution map[string]int, actionStats map[string]map[string]interface{}) []map[string]interface{} {
	insights := make([]map[string]interface{}, 0)

	total := axisDistribution["total"]
	if total > 0 {
		// Find top axis
		topAxis := ""
		topCount := 0
		for axis, count := range axisDistribution {
			if axis != "total" && count > topCount {
				topAxis = axis
				topCount = count
			}
		}

		// If top axis is more than 40% of total
		if topAxis != "" && float64(topCount)/float64(total) > 0.4 {
			label := axisLabels[topAxis]
			if label == "" {
				label = topAxis
			}
			insights = append(insights, map[string]interface{}{
				"type":    "axis_preference",
				"message": "「" + label + "」の観点で考えることが多いようです",
			})
		}
	}

	// Check action usage
	for actionType, stats := range actionStats {
		shown, _ := stats["shown"].(int)
		openRate, _ := stats["open_rate"].(float64)

		label := actionLabels[actionType]
		if label == "" {
			label = actionType
		}

		if openRate > 0.6 {
			insights = append(insights, map[string]interface{}{
				"type":    "action_suggestion",
				"message": label + "の活用度が高いです。さらに深掘りしてみましょう",
			})
		} else if openRate < 0.2 && shown > 10 {
			insights = append(insights, map[string]interface{}{
				"type":    "action_suggestion",
				"message": label + "も試してみましょう",
			})
		}
	}

	return insights
}
