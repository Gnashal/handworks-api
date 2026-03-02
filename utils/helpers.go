package utils

import (
	"encoding/json"
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
