package mdsend

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type Schedule struct {
	After     time.Time     `json:"after"`
	Delay     time.Duration `json:"delay"`
	Step      time.Duration `json:"step"`
	Expire    time.Duration `json:"expire"`
	Fluctuate time.Duration `json:"fluctuate"`
}

func (s Schedule) Next() (time.Time, Schedule) {
	next := s.After
	if next.IsZero() {
		next = time.Now()
	}
	if s.Delay != 0 {
		next = next.Add(s.Delay)
		s.Delay = 0
	}
	s.After = next.Add(s.Step)
	if s.Fluctuate != 0 {
		noise := rand.NormFloat64() * float64(s.Fluctuate)
		next = next.Add(time.Duration(noise))
	}
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

func parseHoursMinutes(v string) (hours int, minutes int, ok bool) {
	h, m, _ := strings.Cut(v, ":")
	hours, err := strconv.Atoi(h)
	if err != nil || hours < 0 || hours > 23 {
		return 0, 0, false
	}

minuteLoop:
	for i, c := range m {
		if c >= '0' && c <= '9' {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(m[i+1:])) {
		case "", "am":
			m = m[:i]
			break minuteLoop
		case "pm":
			m = m[:i]
			hours += 12
			break minuteLoop
		default:
			return 0, 0, false
		}
	}
	minutes, err = strconv.Atoi(m)
	if err != nil || minutes < 0 || minutes > 59 {
		return 0, 0, false
	}
	return hours, minutes, true
}

func parseTime(v any) (time.Time, error) {
	switch v := v.(type) {
	case string:
		fields := strings.Fields(v)
		switch len(fields) {
		case 1:
			return time.Parse(time.DateOnly, fields[0])
		case 2:
			hours, minutes, ok := parseHoursMinutes(fields[1])
			if !ok {
				return time.Parse(time.DateOnly+" MST", fields[0]+" "+fields[1])
			}
			return time.Parse(time.DateOnly+" 15:04", fmt.Sprintf("%s %02d:%02d", fields[0], hours, minutes))
		case 3:
			hours, minutes, ok := parseHoursMinutes(fields[1])
			if !ok {
				return time.Time{}, fmt.Errorf("invalid time value: %s", v)
			}
			return time.Parse(time.DateOnly+" 15:04", fmt.Sprintf("%s %02d:%02d %s", fields[0], hours, minutes, fields[2]))
		default:
			return time.Time{}, fmt.Errorf("invalid time value: %s", v)
		}
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
