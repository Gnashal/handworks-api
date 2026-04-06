package utils

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"
)

func DerefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func DerefFloat32(f *float32) float32 {
	if f == nil {
		return 0
	}
	return *f
}

func DerefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func DerefRaw(r *json.RawMessage) json.RawMessage {
	if r == nil {
		return json.RawMessage("{}")
	}
	return *r
}

func GenerateOrderNumber(quoteID string, createdAt time.Time) string {
	// Combine salts
	input := fmt.Sprintf("%s-%d", quoteID, createdAt.UnixNano())

	hash := sha256.Sum256([]byte(input))

	num := binary.BigEndian.Uint64(hash[:8])

	orderNumber := num % 100000000000

	return fmt.Sprintf("%011d", orderNumber)
}

func DetermineAttendanceStatus(timeIn time.Time) string {
	startOfShift := time.Date(
		timeIn.Year(),
		timeIn.Month(),
		timeIn.Day(),
		9, 0, 0, 0,
		timeIn.Location(),
	)

	if timeIn.After(startOfShift) {
		return "LATE"
	}

	return "ON_TIME"
}

func GetCurrentCalendarMonth(month string) (time.Time, time.Time, error) {
	if month == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("month is required")
	}

	start, err := time.ParseInLocation("2006-01", month, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid month format, expected YYYY-MM")
	}

	end := start.AddDate(0, 1, 0)
	return start, end, nil
}
