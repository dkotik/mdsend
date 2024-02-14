package mdsend

import (
	"fmt"

	"github.com/dkotik/mdsend/loaders"
)

func previewRecepients(prefix string, r *[]loaders.Participant) {
	l := len(*r)
	if l == 0 {
		return
	}
	fmt.Printf("\n| %s: (last five)", prefix)
	limit := l - 5
	if limit < 0 {
		limit = 0
	}
	for i := l - 1; i >= limit; i-- {
		fmt.Printf("\n|     №%d: %s", i+1, (*r)[i])
	}
}

func PreviewMessage(m *loaders.Message) {
	fmt.Printf("%s", m.Body)
	fmt.Printf(".----------------------------------------------\n")
	fmt.Printf("| Subject (%s): %s", m.Date, m.Subject)
	previewRecepients(`To`, m.To)
	previewRecepients(`CC`, m.CC)
	previewRecepients(`BCC`, m.BCC)
	fmt.Printf("\n| Attachments: %v\n", m.Attachments)
	fmt.Printf("`----------------------------------------------\n")
}
