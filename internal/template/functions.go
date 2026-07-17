package template

import (
	"encoding/base64"
	"html/template"
	"net/url"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/dkotik/mdsend/queue"
)

func functions() template.FuncMap {
	return template.FuncMap{
		"RFC3339": func() string {
			return time.Now().Format(time.RFC3339)
		},
		"base64":    base64.RawStdEncoding.EncodeToString,
		"base64URL": base64.RawURLEncoding.EncodeToString,
		"urlQuery":  url.QueryEscape,
		"urlPath":   url.PathEscape,
		"base58": func(in string) string {
			return string(base58.Encode([]byte(in)))
		},
		// "base64": func(in string) string {
		// 	return base64.RawURLEncoding.EncodeToString([]byte(in))
		// },
		"skipMessageIfTrue": func(condition any) (any, error) {
			switch v := condition.(type) {
			case nil:
				return condition, nil
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, complex64, complex128:
				if v == 0 {
					return condition, nil
				}
				return condition, queue.NewSkipSentinelError()
			case string:
				if strings.TrimSpace(v) == "" {
					return condition, nil
				}
				return condition, queue.NewSkipSentinelError()
			case map[any]any:
				if len(v) == 0 {
					return condition, nil
				}
				return condition, queue.NewSkipSentinelError()
			case []any:
				if len(v) == 0 {
					return condition, nil
				}
				return condition, queue.NewSkipSentinelError()
			default:
				return condition, queue.NewSkipSentinelError()
			}
		},
		"reverse": func(in string) string {
			n := 0
			rune := make([]rune, len(in))
			for _, r := range in {
				rune[n] = r
				n++
			}
			rune = rune[0:n]

			for i := 0; i < n/2; i++ { // reverse
				rune[i], rune[n-1-i] = rune[n-1-i], rune[i]
			}
			return string(rune)
		},
	}
}
