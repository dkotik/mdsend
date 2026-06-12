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

func encodeName(name string) string {
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

// WriteAddressHeader writes a list of MIME-mail encoded address header to the given writer.
// It respects [LineLengthLimit] by wrapping long lines.
// If no addresses are provided, it does nothing.
//
// TODO: if the name is too long, it will not be wrapped.
// TODO: if the email address is too long, it will not be wrapped.
func WriteAddressHeader(w io.Writer, name string, addresses ...mail.Address) (err error) {
	if len(addresses) == 0 {
		return nil
	}
	n, err := w.Write([]byte(name))
	if err != nil {
		return err
	}
	if n != len(name) {
		return io.ErrShortWrite
	}
	if _, err = w.Write([]byte(":")); err != nil {
		return err
	}

	remaining := LineLengthLimit - n - 1 // for the colon
	cutIfNeededThenWrite := func(word string) (err error) {
		length := len(word)
		if remaining > length+1 { // +1 for the prefixing space
			_, err = w.Write([]byte(" "))
			if err != nil {
				return err
			}
			n, err := w.Write([]byte(word))
			if err != nil {
				return err
			}
			if n != length {
				return io.ErrShortWrite
			}
			remaining = remaining - n - 1 // for the space
			return err
		}
		n, err := w.Write([]byte(CRNL + " "))
		if err != nil {
			return err
		}
		if n != 3 {
			return io.ErrShortWrite
		}
		n, err = w.Write([]byte(word))
		if err != nil {
			return err
		}
		if n != length {
			return io.ErrShortWrite
		}
		remaining = LineLengthLimit - n - 1 // for the space
		return nil
	}

	isNamePresent := false
	finalIndex := len(addresses) - 1
	for n, address := range addresses {
		isNamePresent = address.Name != ""
		if !isNamePresent {
			if err = cutIfNeededThenWrite(address.Address + ","); err != nil {
				return err
			}
			continue
		}
		if err = cutIfNeededThenWrite(encodeName(address.Name)); err != nil {
			return err
		}
		if n == finalIndex {
			if err = cutIfNeededThenWrite("<" + address.Address + ">"); err != nil {
				return err
			}
			break
		}
		if err = cutIfNeededThenWrite("<" + address.Address + ">,"); err != nil {
			return err
		}
	}

	i, err := w.Write([]byte(CRNL))
	n += i
	if err != nil {
		return err
	}
	if i != 2 {
		return io.ErrShortWrite
	}
	return nil
}
