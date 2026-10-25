package handler

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/suiten/suiten-server/internal/models"
	"github.com/suiten/suiten-server/internal/repository"
)

// northStarMonths is how many months of the North Star Metric are returned.
const northStarMonths = 6

// buildNorthStar returns the North Star Metric (monthly completed "自分のまとめ")
// and its input metrics. Pass uuid.Nil for all users.
func (h *AnalyticsHandler) buildNorthStar(ctx context.Context, userID uuid.UUID, now time.Time, inputsSince time.Time) (map[string]interface{}, error) {
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	since := monthStart.AddDate(0, -(northStarMonths - 1), 0)

	rows, err := h.analyticsRepo.GetNorthStarByMonth(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	inputs, err := h.analyticsRepo.GetNorthStarInputs(ctx, userID, inputsSince, models.SummaryMinCaptures)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"metric":     "completed_summaries_per_month",
		"definition": "captures >= 3, saved or edited by the reader, helpfulness >= 2",
		"monthly":    fillMonths(rows, since, northStarMonths),
		"inputs":     northStarInputsResponse(inputs),
	}, nil
}

// fillMonths returns one entry per month (oldest first), with zero for months without completions.
func fillMonths(rows []repository.NorthStarMonth, since time.Time, months int) []map[string]interface{} {
	counts := make(map[string]int, len(rows))
	for _, r := range rows {
		counts[r.Month.UTC().Format("2006-01")] = r.CompletedSummaries
	}
	result := make([]map[string]interface{}, 0, months)
	for i := 0; i < months; i++ {
		key := since.AddDate(0, i, 0).Format("2006-01")
		result = append(result, map[string]interface{}{
			"month":               key,
			"completed_summaries": counts[key],
		})
	}
	return result
}

func northStarInputsResponse(in *repository.NorthStarInputs) map[string]interface{} {
	return map[string]interface{}{
		"active_books":                 in.ActiveBooks,
		"captures_per_book":            in.CapturesPerBook,
		"doubt_and_empathy_rate":       ratio(in.BooksWithDoubtAndEmpathy, in.ActiveBooks),
		"summary_creation_rate":        ratio(in.BooksWithSummary, in.BooksReadyForSummary),
		"summary_edit_rate":            ratio(in.BooksWithEditedSummary, in.BooksWithSummary),
		"summary_helpfulness_average":  in.SummaryHelpfulnessAvg,
		"books_ready_for_summary":      in.BooksReadyForSummary,
		"books_with_summary":           in.BooksWithSummary,
		"books_with_doubt_and_empathy": in.BooksWithDoubtAndEmpathy,
		"books_with_edited_summary":    in.BooksWithEditedSummary,
	}
}

func ratio(part, whole int) float64 {
	if whole == 0 {
		return 0
	}
	return float64(part) / float64(whole)
}
