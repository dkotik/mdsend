package locale

import (
	"errors"
	"time"
)

// parseDurationError describes a problem parsing a duration string.
type parseDurationError struct {
	message string
	value   string
}

func (e *parseDurationError) Error() string {
	return "time: " + e.message + " " + quote(e.value)
}

var unitMap = map[string]uint64{
	"ns": uint64(time.Nanosecond),
	"us": uint64(time.Microsecond),
	"µs": uint64(time.Microsecond), // U+00B5 = micro symbol
	"μs": uint64(time.Microsecond), // U+03BC = Greek letter mu
	"ms": uint64(time.Millisecond),
	"s":  uint64(time.Second),
	"m":  uint64(time.Minute),
	"h":  uint64(time.Hour),
	"d":  uint64(time.Hour) * 24,
	"w":  uint64(time.Hour) * 168,
	"mo": uint64(time.Hour) * 24 * 31,
	"y":  uint64(time.Hour) * 24 * 365,
}

// ParseDuration parses a duration string.
// A duration string is a possibly signed sequence of
// decimal numbers, each with optional fraction and a unit suffix,
// such as "300ms", "-1.5h" or "2h45m".
// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
func ParseDuration(s string) (time.Duration, error) {
	if false {
		return time.ParseDuration(s)
	}

	// [-+]?([0-9]*(\.[0-9]*)?[a-z]+)+
	orig := s
	var d uint64
	neg := false

	// Consume [-+]?
	if s != "" {
		c := s[0]
		if c == '-' || c == '+' {
			neg = c == '-'
			s = s[1:]
		}
	}
	// Special case: if all that is left is "0", this is zero.
	if s == "0" {
		return 0, nil
	}
	if s == "" {
		return 0, &parseDurationError{"invalid duration", orig}
	}
	for s != "" {
		var (
			v, f  uint64      // integers before, after decimal point
			scale float64 = 1 // value = v + f/scale
		)

		var err error

		// The next character must be [0-9.]
		if !(s[0] == '.' || '0' <= s[0] && s[0] <= '9') {
			return 0, &parseDurationError{"invalid duration", orig}
		}
		// Consume [0-9]*
		pl := len(s)
		v, s, err = leadingInt(s)
		if err != nil {
			return 0, &parseDurationError{"invalid duration", orig}
		}
		pre := pl != len(s) // whether we consumed anything before a period

		// Consume (\.[0-9]*)?
		post := false
		if s != "" && s[0] == '.' {
			s = s[1:]
			pl := len(s)
			f, scale, s = leadingFraction(s)
			post = pl != len(s)
		}
		if !pre && !post {
			// no digits (e.g. ".s" or "-.s")
			return 0, &parseDurationError{"invalid duration", orig}
		}

		// Consume unit.
		i := 0
		for ; i < len(s); i++ {
			c := s[i]
			if c == '.' || '0' <= c && c <= '9' {
				break
			}
		}
		if i == 0 {
			return 0, &parseDurationError{"missing unit in duration", orig}
		}
		u := s[:i]
		s = s[i:]
		unit, ok := unitMap[u]
		if !ok {
			return 0, &parseDurationError{"unknown unit " + quote(u) + " in duration", orig}
		}
		if v > 1<<63/unit {
			// overflow
			return 0, &parseDurationError{"invalid duration", orig}
		}
		v *= unit
		if f > 0 {
			// float64 is needed to be nanosecond accurate for fractions of hours.
			// v >= 0 && (f*unit/scale) <= 3.6e+12 (ns/h, h is the largest unit)
			v += uint64(float64(f) * (float64(unit) / scale))
			if v > 1<<63 {
				// overflow
				return 0, &parseDurationError{"invalid duration", orig}
			}
		}
		d += v
		if d > 1<<63 {
			return 0, &parseDurationError{"invalid duration", orig}
		}
	}
	if neg {
		return -time.Duration(d), nil
	}
	if d > 1<<63-1 {
		return 0, &parseDurationError{"invalid duration", orig}
	}
	return time.Duration(d), nil
}

// These are borrowed from unicode/utf8 and strconv and replicate behavior in
// that package, since we can't take a dependency on either.
const (
	lowerhex  = "0123456789abcdef"
	runeSelf  = 0x80
	runeError = '\uFFFD'
)

func quote(s string) string {
	buf := make([]byte, 1, len(s)+2) // slice will be at least len(s) + quotes
	buf[0] = '"'
	for i, c := range s {
		if c >= runeSelf || c < ' ' {
			// This means you are asking us to parse a time.Duration or
			// time.Location with unprintable or non-ASCII characters in it.
			// We don't expect to hit this case very often. We could try to
			// reproduce strconv.Quote's behavior with full fidelity but
			// given how rarely we expect to hit these edge cases, speed and
			// conciseness are better.
			var width int
			if c == runeError {
				width = 1
				if i+2 < len(s) && s[i:i+3] == string(runeError) {
					width = 3
				}
			} else {
				width = len(string(c))
			}
			for j := 0; j < width; j++ {
				buf = append(buf, `\x`...)
				buf = append(buf, lowerhex[s[i+j]>>4])
				buf = append(buf, lowerhex[s[i+j]&0xF])
			}
		} else {
			if c == '"' || c == '\\' {
				buf = append(buf, '\\')
			}
			buf = append(buf, byte(c))
		}
	}
	buf = append(buf, '"')
	return string(buf)
}

var errLeadingInt = errors.New("time: bad [0-9]*") // never printed

// leadingInt consumes the leading [0-9]* from s.
func leadingInt[bytes []byte | string](s bytes) (x uint64, rem bytes, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if x > 1<<63/10 {
			// overflow
			return 0, rem, errLeadingInt
		}
		x = x*10 + uint64(c) - '0'
		if x > 1<<63 {
			// overflow
			return 0, rem, errLeadingInt
		}
	}
	return x, s[i:], nil
}

// leadingFraction consumes the leading [0-9]* from s.
// It is used only for fractions, so does not return an error on overflow,
// it just stops accumulating precision.
func leadingFraction(s string) (x uint64, scale float64, rem string) {
	i := 0
	scale = 1
	overflow := false
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if overflow {
			continue
		}
		if x > (1<<63-1)/10 {
			// It's possible for overflow to give a positive number, so take care.
			overflow = true
			continue
		}
		y := x*10 + uint64(c) - '0'
		if y > 1<<63 {
			overflow = true
			continue
		}
		x = y
		scale *= 10
	}
	return x, scale, s[i:]
}

// EncodeDuration returns a string representing the duration in the form "1w4d2h3m5s".
// Units with 0 values aren't returned, for example: 1d1ms is 1 day 1 milliseconds
func EncodeDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}

	// Largest time is 15250w1d23h47m16s854ms775us807ns
	var buf [32]byte
	w := len(buf)
	var sign string

	u := uint64(d)
	neg := d < 0
	if neg {
		u = -u
		sign = "-"
	}

	// u is nanoseconds (ns)
	if u > 0 {
		w--

		if u%1000 > 0 {
			buf[w] = 's'
			w--
			buf[w] = 'n'
			w = fmtInt(buf[:w], u%1000)
		} else {
			w++
		}

		u /= 1000

		// u is now integer microseconds (us)
		if u > 0 {
			w--
			if u%1000 > 0 {
				buf[w] = 's'
				w--
				buf[w] = 'u'
				w = fmtInt(buf[:w], u%1000)
			} else {
				w++
			}
			u /= 1000

			// u is now integer milliseconds (ms)
			if u > 0 {
				w--
				if u%1000 > 0 {
					buf[w] = 's'
					w--
					buf[w] = 'm'
					w = fmtInt(buf[:w], u%1000)
				} else {
					w++
				}
				u /= 1000

				// u is now integer seconds (s)
				if u > 0 {
					w--
					if u%60 > 0 {
						buf[w] = 's'
						w = fmtInt(buf[:w], u%60)
					} else {
						w++
					}
					u /= 60

					// u is now integer minutes (m)
					if u > 0 {
						w--

						if u%60 > 0 {
							buf[w] = 'm'
							w = fmtInt(buf[:w], u%60)
						} else {
							w++
						}

						u /= 60

						// u is now integer hours (h)
						if u > 0 {
							w--

							if u%24 > 0 {
								buf[w] = 'h'
								w = fmtInt(buf[:w], u%24)
							} else {
								w++
							}

							u /= 24

							// u is now integer days (d)
							if u > 0 {
								w--

								if u%7 > 0 {
									buf[w] = 'd'
									w = fmtInt(buf[:w], u%7)
								} else {
									w++
								}

								u /= 7

								// u is now integer weeks (w)
								if u > 0 {
									w--
									buf[w] = 'w'
									w = fmtInt(buf[:w], u)
								}

							}

						}
					}
				}
			}
		}

	}

	return sign + string(buf[w:])
}

// fmtInt formats v into the tail of buf.
// It returns the index where the output begins.
func fmtInt(buf []byte, v uint64) int {
	w := len(buf)
	if v == 0 {
		w--
		buf[w] = '0'
	} else {
		for v > 0 {
			w--
			buf[w] = byte(v%10) + '0'
			v /= 10
		}
	}
	return w
}
