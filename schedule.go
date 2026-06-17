package mdsend

import (
	"fmt"
	"strings"
	"time"
)

type Schedule struct {
	After     time.Time     `json:"after"`
	Delay     time.Duration `json:"delay"`
	Step      time.Duration `json:"step"`
	Expire    time.Duration `json:"expire"`
	Fluctuate time.Duration `json:"fluctuate"`
	// Hour      int           `json:"hour"`
	// Timezone string `json:"timezone"`
}

func (s Schedule) Next() (time.Time, Schedule) {
	next := s.After.Add(s.Step)
	return next, s
}

func (l Letter) GetSchedule() (s Schedule, err error) {
	m, ok := l.Frontmatter[FieldNameSchedule]
	if !ok {
		return s, nil
	}
	switch m := m.(type) {
	case nil:
		return s, nil
	case map[string]any:
		after, ok := m[FieldNameScheduleAfter]
		if ok {
			s.After, err = parseTime(after)
			if err != nil {
				return s, err
			}
		}
		delay, ok := m[FieldNameScheduleDelay]
		if ok {
			s.Delay, err = parseDuration(delay)
			if err != nil {
				return s, err
			}
		}
		step, ok := m[FieldNameScheduleStep]
		if ok {
			s.Step, err = parseDuration(step)
			if err != nil {
				return s, err
			}
		}
		expire, ok := m[FieldNameScheduleExpire]
		if ok {
			s.Expire, err = parseDuration(expire)
			if err != nil {
				return s, err
			}
		}
		fluctuate, ok := m[FieldNameScheduleFluctuate]
		if ok {
			s.Fluctuate, err = parseDuration(fluctuate)
			if err != nil {
				return s, err
			}
		}
	default:
		return s, fmt.Errorf("invalid schedule value: %v (%T)", m, m)
	}
	return s, nil
}

func parseMonth(v string) (time.Month, error) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "january", "jan":
		return time.January, nil
	case "february", "feb":
		return time.February, nil
	case "march", "mar":
		return time.March, nil
	case "april", "apr":
		return time.April, nil
	case "may":
		return time.May, nil
	case "june", "jun":
		return time.June, nil
	case "july", "jul":
		return time.July, nil
	case "august", "aug":
		return time.August, nil
	case "september", "sep":
		return time.September, nil
	case "october", "oct":
		return time.October, nil
	case "november", "nov":
		return time.November, nil
	case "december", "dec":
		return time.December, nil
	default:
		return 0, fmt.Errorf("invalid month value: %s", v)
	}
}

func parseTime(v any) (time.Time, error) {
	switch v := v.(type) {
	case string:
		// fields := strings.Fields(v)
		return time.Parse(time.RFC3339, v)
	default:
		return time.Time{}, fmt.Errorf("invalid time value: %v (%T)", v, v)
	}
}

func parseDuration(v any) (time.Duration, error) {
	switch v := v.(type) {
	case string:
		return time.ParseDuration(v)
	default:
		return 0, fmt.Errorf("invalid duration value: %v (%T)", v, v)
	}
}
