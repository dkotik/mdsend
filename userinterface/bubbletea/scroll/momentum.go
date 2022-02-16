package scroll

import (
	"sync"
	"time"
)

const momentumMaximum = 5

type Momentum struct {
	quantity     int
	blockedUntil time.Time
	frequency    time.Duration
	mu           sync.Mutex
}

func (m *Momentum) Up() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	t := time.Now()
	if t.Before(m.blockedUntil) {
		return 0 // too many messages
	}

	if m.blockedUntil.Add(m.frequency * momentumMaximum).Before(t) {
		m.quantity = 0 // reset old momentum values
	}
	m.blockedUntil = t.Add(m.frequency)

	if m.quantity > 0 {
		m.quantity = -1 // change direction
	} else if m.quantity > -momentumMaximum {
		m.quantity--
	}
	return m.quantity
}

func (m *Momentum) Down() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	t := time.Now()
	if t.Before(m.blockedUntil) {
		return 0 // too many messages
	}

	if m.blockedUntil.Add(m.frequency * momentumMaximum).Before(t) {
		m.quantity = 0 // reset old momentum values
	}
	m.blockedUntil = t.Add(m.frequency)

	if m.quantity < 0 {
		m.quantity = 1 // change direction
	} else if m.quantity < momentumMaximum {
		m.quantity++
	}
	return m.quantity
}
