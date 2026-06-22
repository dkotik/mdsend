package markdown

import (
	"bytes"
	"strings"
	"unicode"
)

func IsPathLocal(p string) bool {
	if len(p) == 0 {
		return false
	}
	switch p[0] {
	case '.':
		return true
	case '/', '\\':
		return false
	case ' ', '\t', '\n', '\r':
		for i, c := range p[1:] {
			if unicode.IsSpace(c) {
				continue
			}
			return IsPathLocal(p[i:])
		}
		return false
	}
	return !strings.Contains(p, "://")
}

func IsPathLocalBytes(p []byte) bool {
	if len(p) == 0 {
		return false
	}
	switch p[0] {
	case '.':
		return true
	case '/', '\\':
		return false
	case ' ', '\t', '\n', '\r':
		for i, c := range p {
			if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
				return IsPathLocalBytes(p[i-1:])
			}
		}
		return false
	}
	return !bytes.Contains(p, []byte("://"))
}
