package loaders

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"

	"github.com/spf13/viper"
)

// ViperLoader constructs message using viper config parser from a markdown file.
type ViperLoader struct{}

// Load processes the markdown file into a message structure.
func (l *ViperLoader) Load(source string, r io.Reader) (*Message, error) {
	m := &Message{}
	s := NewDocumentChunkReader(r, DocumentBoundary)
	context := viper.New()
	context.SetConfigType(`yaml`)
	err := context.ReadConfig(s)
	if err == nil {
		err = context.Unmarshal(&m.Data)
		if err == nil {
			b := bytes.NewBuffer(nil)
			for s.Next() {
				io.Copy(b, s)
				b.Write([]byte("\n\n"))
			}
			m.Body = b.Bytes()
		}
	}
	m.Source = source
	dir := filepath.Dir(source)
	if val, ok := m.Data[`from`].(string); ok {
		m.From = val
	}
	if val, ok := m.Data[`reply-to`].(string); ok {
		m.ReplyTo = val
	}
	if val, ok := m.Data[`subject`].(string); ok {
		m.Subject = val
	}
	if val, ok := m.Data[`comments`].(string); ok {
		m.Comments = val
	}
	if val, ok := m.Data[`keywords`].(string); ok {
		m.Keywords = val
	}
	if val, ok := m.Data[`attachments`].([]interface{}); ok {
		for _, a := range val {
			m.Attachments = append(m.Attachments, PathAutoJoin(dir, fmt.Sprintf(`%v`, a)))
		}
	}
	m.Date = fmt.Sprintf(`%v`, m.Data[`date`])
	m.To = participantsFromInterface(dir, m.Data[`to`])
	m.CC = participantsFromInterface(dir, m.Data[`cc`])
	m.BCC = participantsFromInterface(dir, m.Data[`bcc`])
	// if val, ok := m.Data[`to`].(string); ok {
	// 	*m.To = append(&m.To, Participant{Email: val})
	// }
	return m, err
}
