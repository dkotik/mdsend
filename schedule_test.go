package mdsend

import (
	"testing"
	"time"
)

func TestScheduleParsing(t *testing.T) {
	s, err := Letter{
		Frontmatter: map[string]any{
			FieldNameSchedule: map[string]any{
				FieldNameScheduleAfter:     "1999-01-01 00:00",
				FieldNameScheduleDelay:     "11m",
				FieldNameScheduleStep:      "1m",
				FieldNameScheduleExpire:    "30m",
				FieldNameScheduleFluctuate: "10m",
			},
		},
	}.GetSchedule()
	if err != nil {
		t.Fatal(err)
	}
	if s.After.Truncate(time.Hour).Truncate(time.Minute).Truncate(time.Second).Equal(time.Date(1999, 1, 1, 0, 0, 0, 0, time.Local)) {
		t.Fatalf("expected schedule to be 1999-1-1, got %v", s.After)
	}
	if s.Expire != 30*time.Minute {
		t.Fatalf("expected expire to be 30m, got %d", s.Expire)
	}
	if s.Fluctuate != 10*time.Minute {
		t.Fatalf("expected fluctuate to be 10m, got %d", s.Fluctuate)
	}
}
