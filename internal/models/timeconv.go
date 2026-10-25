package models

import "time"

// millisThreshold separates Unix seconds from milliseconds.
// 1e11 seconds is in the year 5138, so any larger value must be milliseconds.
const millisThreshold = 100_000_000_000

// UnixToTime converts a Unix timestamp to time.Time.
// The API uses seconds, but older clients sent milliseconds, so both are accepted.
func UnixToTime(v int64) time.Time {
	if v > millisThreshold || v < -millisThreshold {
		return time.UnixMilli(v)
	}
	return time.Unix(v, 0)
}

// UnixPtrToTime converts an optional Unix timestamp to an optional time.Time.
func UnixPtrToTime(v *int64) *time.Time {
	if v == nil {
		return nil
	}
	t := UnixToTime(*v)
	return &t
}

// TimePtrToUnix converts an optional time.Time to optional Unix seconds.
func TimePtrToUnix(t *time.Time) *int64 {
	if t == nil {
		return nil
	}
	v := t.Unix()
	return &v
}
