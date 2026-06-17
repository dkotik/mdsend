package renderers

import (
	"net/textproto"

	"github.com/dkotik/mdsend/loaders"
)

type UnsubscribeLinkFields struct {
	Name    string
	Address string
	ListID  string
}

func NewHeader(m *loaders.Message, to loaders.Participant) textproto.MIMEHeader {
	return textproto.MIMEHeader{}
}
