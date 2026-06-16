package mdsend

import "time"

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
	next := s.After.Add(s.Delay)
	return next, s
}

func (l Letter) GetSchedule() (s Schedule, err error) {
	m, ok := l.Frontmatter[FieldNameSchedule]
	if !ok {
		return s, nil
	}
	s = m.(Schedule)
	return s, nil
}
