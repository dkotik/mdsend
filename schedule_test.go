package mdsend

import (
	"testing"
	"time"
)

func TestScheduleParsing(t *testing.T) {
	s, err := Letter{
		Frontmatter: map[string]any{
			FieldNameSchedule: map[string]any{
				FieldNameScheduleAfter:     "1999-01-01",
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
	if s.After.Equal(time.Date(1999, 1, 1, 0, 0, 0, 0, nil)) {
		t.Fatalf("expected schedule to be 1999-1-1, got %v", s.After)
	}
	if s.Expire != 1800 {
		t.Fatalf("expected expire to be 1800, got %d", s.Expire)
	}
	if s.Fluctuate != 600 {
		t.Fatalf("expected fluctuate to be 600, got %d", s.Fluctuate)
	}

}
