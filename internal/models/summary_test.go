package models

import (
	"testing"
	"time"
)

func ptrInt(v int) *int { return &v }

func ptrTime(t time.Time) *time.Time { return &t }

func TestSummaryMetricsIsComplete(t *testing.T) {
	gen := time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)
	saved := gen.Add(time.Hour)

	complete := SummaryMetrics{
		GeneratedAt:  ptrTime(gen),
		SavedAt:      ptrTime(saved),
		Helpfulness:  ptrInt(2),
		CaptureCount: ptrInt(3),
	}

	tests := []struct {
		name   string
		modify func(m *SummaryMetrics)
		want   bool
	}{
		{"saved, 3 captures, helpfulness 2", func(m *SummaryMetrics) {}, true},
		{"edited instead of saved", func(m *SummaryMetrics) { m.SavedAt = nil; m.EditedAt = ptrTime(saved) }, true},
		{"only 2 captures", func(m *SummaryMetrics) { m.CaptureCount = ptrInt(2) }, false},
		{"neither saved nor edited", func(m *SummaryMetrics) { m.SavedAt = nil }, false},
		{"helpfulness 1", func(m *SummaryMetrics) { m.Helpfulness = ptrInt(1) }, false},
		{"not rated", func(m *SummaryMetrics) { m.Helpfulness = nil }, false},
		{"never generated", func(m *SummaryMetrics) { m.GeneratedAt = nil }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := complete
			tt.modify(&m)
			if got := m.IsComplete(); got != tt.want {
				t.Errorf("IsComplete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSummaryMetricsCompletedAtUsesLatestAction(t *testing.T) {
	gen := time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)
	saved := gen.Add(time.Hour)
	edited := gen.Add(2 * time.Hour)

	m := SummaryMetrics{
		GeneratedAt:  ptrTime(gen),
		SavedAt:      ptrTime(saved),
		EditedAt:     ptrTime(edited),
		Helpfulness:  ptrInt(3),
		CaptureCount: ptrInt(5),
	}
	got := m.CompletedAt()
	if got == nil || !got.Equal(edited) {
		t.Fatalf("CompletedAt() = %v, want %v", got, edited)
	}

	m.Helpfulness = ptrInt(1)
	if m.CompletedAt() != nil {
		t.Error("CompletedAt() should be nil for an incomplete summary")
	}
}
