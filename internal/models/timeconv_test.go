package models

import (
	"testing"
	"time"
)

func TestUnixToTime(t *testing.T) {
	want := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		in   int64
	}{
		{"seconds", want.Unix()},
		{"milliseconds from older clients", want.UnixMilli()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UnixToTime(tt.in); !got.Equal(want) {
				t.Errorf("UnixToTime(%d) = %v, want %v", tt.in, got, want)
			}
		})
	}
}

func TestUnixPtrToTimeNil(t *testing.T) {
	if UnixPtrToTime(nil) != nil {
		t.Error("UnixPtrToTime(nil) should be nil")
	}
	if TimePtrToUnix(nil) != nil {
		t.Error("TimePtrToUnix(nil) should be nil")
	}
}
