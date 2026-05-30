package mime

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/textproto"
	"sort"
)

func newHeaderTemplate(header textproto.MIMEHeader) (t *template.Template, err error) {
	b := &bytes.Buffer{}
	if err = writeHeader(b, header); err != nil {
		return nil, err
	}
	t, err = template.New("").Parse(b.String())
	if err != nil {
		return nil, fmt.Errorf("invalid header template: %w", err)
	}
	return t, nil
}

func writeHeader(w io.Writer, header textproto.MIMEHeader) (err error) {
	keys := make([]string, 0, len(header))
	for k := range header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range header[k] {
			if _, err = fmt.Fprintf(w, "%s: %s\r\n", k, v); err != nil {
				return err
			}
		}
	}
	_, err = fmt.Fprintf(w, "\r\n")
	return err
}
