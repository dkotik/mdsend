package mdsend

import "io"

type Message interface {
	To() []string
	CC() []string
	BCC() []string
	Subject() string
	MIMEBody() io.ReadCloser
}
