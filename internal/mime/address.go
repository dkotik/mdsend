package mime

import (
	"encoding/base64"
	"io"
	"net/mail"
	"strings"
	"unicode/utf8"
)

// isVchar reports whether r is an RFC 5322 VCHAR character.
// from net/mail/message.go
func isVchar(r rune) bool {
	// Visible (printing) characters.
	return '!' <= r && r <= '~' || isMultibyte(r)
}

// isMultibyte reports whether r is a multi-byte UTF-8 character
// as supported by RFC 6532.
// from net/mail/message.go
func isMultibyte(r rune) bool {
	return r >= utf8.RuneSelf
}

// isWSP reports whether r is a WSP (white space).
// WSP is a space or horizontal tab (RFC 5234 Appendix B).
// from net/mail/message.go
func isWSP(r rune) bool {
	return r == ' ' || r == '\t'
}

// isQtext reports whether r is an RFC 5322 qtext character.
// from net/mail/message.go
func isQtext(r rune) bool {
	// Printable US-ASCII, excluding backslash or quote.
	if r == '\\' || r == '"' {
		return false
	}
	return isVchar(r)
}

// quoteString renders a string as an RFC 5322 quoted-string.
// from net/mail/message.go
func quoteString(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		if isQtext(r) || isWSP(r) {
			b.WriteRune(r)
		} else if isVchar(r) {
			b.WriteByte('\\')
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

func isAllPrintable(s string) bool {
	// loop from net/mail/message.go
	for _, r := range s {
		// isWSP here should actually be isFWS,
		// but we don't support folding yet.
		if !isVchar(r) && !isWSP(r) || isMultibyte(r) {
			return false
		}
	}
	return true
}

func EncodeName(name string) string {
	if isAllPrintable(name) {
		return quoteString(name)
	}
	b := &strings.Builder{}
	_, _ = b.WriteString(BEncodingPrefix)
	w64 := base64.NewEncoder(base64.StdEncoding, b)
	_, _ = w64.Write([]byte(name))
	_ = w64.Close()
	_, _ = b.WriteString(BEncodingSuffix)
	return b.String()
}

func WriteAddressHeader(w io.Writer, name string, addresses ...mail.Address) (err error) {
	if len(addresses) == 0 {
		return nil
	}
	n, err := w.Write([]byte(name + ": "))
	if err != nil {
		return err
	}
	if n-2 != len(name) {
		return io.ErrShortWrite
	}

	remaining := LineLengthLimit - n
	name = addresses[0].Name
	namePreceedsAddress := name != ""
	if namePreceedsAddress {
		name = EncodeName(name)
		if len(name) > remaining {
			n, err = w.Write([]byte(CRNL + " "))
			if err != nil {
				return err
			}
			if n != 3 {
				return io.ErrShortWrite
			}
			remaining = LineLengthLimit - n
		}
		n, err = w.Write([]byte(name))
		if err != nil {
			return err
		}
		if n != len(name) {
			return io.ErrShortWrite
		}
		remaining -= n
	}

	if len(addresses) > 1 {
		remaining = remaining - 2 // one for first comma and space
	}
	length := len(addresses[0].Address)
	if remaining < length-3 { // three for space < >
		n, err = w.Write([]byte(CRNL + " "))
		if err != nil {
			return err
		}
		if n != 3 {
			return io.ErrShortWrite
		}
		remaining = LineLengthLimit - n
	}
	if namePreceedsAddress {
		n, err = w.Write([]byte(" <"))
		if err != nil {
			return err
		}
		if n != 2 {
			return io.ErrShortWrite
		}
		remaining -= 2 // space and angle bracket
	}
	n, err = w.Write([]byte(addresses[0].Address))
	if err != nil {
		return err
	}
	if n != length {
		return io.ErrShortWrite
	}
	if namePreceedsAddress {
		_, err = w.Write([]byte(">"))
		if err != nil {
			return err
		}
		remaining-- // for the angle bracket
	}
	remaining -= length

	for _, address := range addresses[1:] {
		n, err = w.Write([]byte(", "))
		if err != nil {
			return err
		}
		if n != 2 {
			return io.ErrShortWrite
		}
		remaining = remaining - 2 // for the comma space
		namePreceedsAddress = address.Name != ""
		if !namePreceedsAddress {
			length = len(address.Address)
			if remaining < length {
				n, err = w.Write([]byte(CRNL + " "))
				if err != nil {
					return err
				}
				if n != 3 {
					return io.ErrShortWrite
				}
				remaining = LineLengthLimit - n
			}
			n, err = w.Write([]byte(address.Address))
			if err != nil {
				return err
			}
			if n != length {
				return io.ErrShortWrite
			}
			remaining -= length
			continue
		}

		name = EncodeName(address.Name)
		if len(name) > remaining {
			n, err = w.Write([]byte(CRNL + " "))
			if err != nil {
				return err
			}
			if n != 3 {
				return io.ErrShortWrite
			}
			remaining = LineLengthLimit - n
		}
		n, err = w.Write([]byte(name))
		if err != nil {
			return err
		}
		if n != len(name) {
			return io.ErrShortWrite
		}
		remaining -= n

		length = len(address.Address)
		if remaining < length-4 { // three for space < > and comma between addresses
			n, err = w.Write([]byte(CRNL + " "))
			if err != nil {
				return err
			}
			if n != 3 {
				return io.ErrShortWrite
			}
			remaining = LineLengthLimit - n
		}
		if namePreceedsAddress {
			n, err = w.Write([]byte(" <"))
			if err != nil {
				return err
			}
			if n != 2 {
				return io.ErrShortWrite
			}
			remaining -= 2 // for the spance and angle bracket
		}
		n, err = w.Write([]byte(address.Address))
		if err != nil {
			return err
		}
		if n != length {
			return io.ErrShortWrite
		}
		remaining -= length
		if namePreceedsAddress {
			_, err = w.Write([]byte(">"))
			if err != nil {
				return err
			}
			remaining -= 1 // for the angle bracket
		}
	}

	return err
}
