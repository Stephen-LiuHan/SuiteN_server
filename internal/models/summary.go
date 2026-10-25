package models

import "time"

// SummaryMinCaptures is the number of captures a book needs before its summary
// can count as complete. One capture is a single reaction; the summary only shows
// the reader's way of thinking when it cuts across several of them.
const SummaryMinCaptures = 3

// SummaryMinHelpfulness is the lowest helpfulness rating (1-3) for a complete summary.
const SummaryMinHelpfulness = 2

// SummaryMetrics holds the book-level summary fields used for the North Star Metric
// (monthly count of completed "自分のまとめ").
type SummaryMetrics struct {
	GeneratedAt  *time.Time
	SavedAt      *time.Time
	EditedAt     *time.Time
	Helpfulness  *int
	CaptureCount *int
}

// IsComplete reports whether the summary meets the completion definition:
// built from at least SummaryMinCaptures captures, saved or edited by the reader,
// and rated at least SummaryMinHelpfulness.
func (m SummaryMetrics) IsComplete() bool {
	if m.GeneratedAt == nil || m.CaptureCount == nil || m.Helpfulness == nil {
		return false
	}
	if *m.CaptureCount < SummaryMinCaptures {
		return false
	}
	if m.SavedAt == nil && m.EditedAt == nil {
		return false
	}
	return *m.Helpfulness >= SummaryMinHelpfulness
}

// CompletedAt returns the time the summary became complete, or nil if it is not.
// The latest of the saved, edited, and generated times is used, since completion
// happens on the last action that satisfied the definition.
func (m SummaryMetrics) CompletedAt() *time.Time {
	if !m.IsComplete() {
		return nil
	}
	var latest time.Time
	for _, t := range []*time.Time{m.GeneratedAt, m.SavedAt, m.EditedAt} {
		if t != nil && t.After(latest) {
			latest = *t
		}
	}
	return &latest
}
