package mime

import (
	"fmt"
	"strings"
)

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"", "\r", "%0D", "\n", "%0A")

// escapeQuotes escapes special characters in field parameter values.
//
// For historical reasons, this uses \ escaping for " and \ characters,
// and percent encoding for CR and LF.
//
// The WhatWG specification for form data encoding suggests that we should
// use percent encoding for " (%22), and should not escape \.
// https://html.spec.whatwg.org/multipage/form-control-infrastructure.html#multipart/form-data-encoding-algorithm
//
// Empirically, as of the time this comment was written, it is necessary
// to escape \ characters or else Chrome (and possibly other browsers) will
// interpet the unescaped \ as an escape.
func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// FileContentDisposition returns the value of a Content-Disposition header
// with the provided field name and file name.
func FileContentDisposition(fieldname, filename string) string {
	return fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
		escapeQuotes(fieldname), escapeQuotes(filename))
}
