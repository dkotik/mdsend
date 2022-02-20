package scroll

import (
	"time"
)

func NewMomentum(depth int, frequency time.Duration) *Momentum {
	return &Momentum{
		maximum:        depth,
		frequency:      frequency,
		resetFrequency: frequency * time.Duration(depth),
	}
}

type Momentum struct {
	quantity       int
	maximum        int
	blockedUntil   time.Time
	frequency      time.Duration
	resetFrequency time.Duration
	// mu           sync.Mutex
}

func (m *Momentum) Up() int {
	// m.mu.Lock()
	// defer m.mu.Unlock()

	t := time.Now()
	if t.Before(m.blockedUntil) {
		return 0 // too many messages
	}

	if m.blockedUntil.Add(m.resetFrequency).Before(t) {
		m.quantity = 0 // reset old momentum values
	}
	m.blockedUntil = t.Add(m.frequency)

	if m.quantity > 0 {
		m.quantity = -1 // change direction
	} else if m.quantity > -m.maximum {
		m.quantity--
	}
	return -m.quantity
}

func (m *Momentum) Down() int {
	// m.mu.Lock()
	// defer m.mu.Unlock()

	t := time.Now()
	if t.Before(m.blockedUntil) {
		return 0 // too many messages
	}

	if m.blockedUntil.Add(m.resetFrequency).Before(t) {
		m.quantity = 0 // reset old momentum values
	}
	m.blockedUntil = t.Add(m.frequency)

	if m.quantity < 0 {
		m.quantity = 1 // change direction
	} else if m.quantity < m.maximum {
		m.quantity++
	}
	return m.quantity
}
