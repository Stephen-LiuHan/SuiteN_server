package handler

import (
	"testing"
	"time"

	"github.com/suiten/suiten-server/internal/repository"
)

func TestFillMonthsAddsZeroMonths(t *testing.T) {
	since := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	rows := []repository.NorthStarMonth{
		{Month: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), CompletedSummaries: 4},
		{Month: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), CompletedSummaries: 9},
	}

	got := fillMonths(rows, since, 6)

	want := []struct {
		month string
		count int
	}{
		{"2026-05", 0}, {"2026-06", 0}, {"2026-07", 4}, {"2026-08", 0}, {"2026-09", 0}, {"2026-10", 9},
	}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i]["month"] != w.month || got[i]["completed_summaries"] != w.count {
			t.Errorf("month %d = %v, want %s/%d", i, got[i], w.month, w.count)
		}
	}
}

func TestRatio(t *testing.T) {
	if ratio(1, 0) != 0 {
		t.Error("ratio with zero denominator should be 0")
	}
	if ratio(1, 4) != 0.25 {
		t.Errorf("ratio(1, 4) = %v, want 0.25", ratio(1, 4))
	}
}
