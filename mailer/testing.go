package mailer

import (
	"github.com/dkotik/mdsend"
	"github.com/dkotik/mdsend/internal"
	"github.com/dkotik/mdsend/media"
)

var mockMediaContraints = media.Constraints{
	Width:   100,
	Height:  100,
	Quality: 20,
}

func NewMockAttachment(name string) mdsend.Attachment {
	var b []byte
	switch name {
	case "cat":
		b = internal.Cat
	case "panda":
		b = internal.Panda
	case "chamillion":
		b = internal.Chamillion
	default:
		panic("unknown mock attachment name")
	}

	attachment, err := mdsend.NewAttachment(b, mockMediaContraints)
	if err != nil {
		panic(err)
	}
	attachment.Name = name + ".jpg"
	return attachment
}

func NewInlineMockAttachment(name string) mdsend.Attachment {
	attachment := NewMockAttachment(name)
	attachment.ContentID = attachment.Hash + "@test.com"
	return attachment
}
