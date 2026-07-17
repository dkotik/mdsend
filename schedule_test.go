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

	if s.After.Year() != 1999 {
		t.Fatal("unexpected year:", s.After.Year())
	}
	if s.After.Month() != time.January {
		t.Fatal("unexpected month:", s.After.Month())
	}
	if s.After.Day() != 1 {
		t.Fatal("unexpected day:", s.After.Day())
	}
	if s.After.Hour() != 0 {
		t.Fatal("unexpected hour:", s.After.Hour())
	}
	if s.After.Minute() != 11 {
		// minute is increased, because the delay is applied to the After
		t.Fatal("unexpected minute:", s.After.Minute())
	}

	if s.Expire != 30*time.Minute {
		t.Fatalf("expected expire to be 30m, got %d", s.Expire)
	}
	if s.Fluctuate != 10*time.Minute {
		t.Fatalf("expected fluctuate to be 10m, got %d", s.Fluctuate)
	}
}
