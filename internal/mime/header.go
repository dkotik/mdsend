package mime

import (
	"fmt"
	"io"
	"net/textproto"
	"sort"
)

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
