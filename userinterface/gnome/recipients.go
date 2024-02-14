package gnome

import (
	"fmt"
	"log"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

var recipients uint16

func (g *GUI) addRecipient(p string) error {
	recipients++
	r, err := gtk.ListBoxRowNew()
	if err != nil {
		return err
	}
	b, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 10)
	if err != nil {
		return err
	}
	l, err := gtk.LabelNew(p)
	if err != nil {
		return err
	}
	l2, err := gtk.LabelNew(fmt.Sprintf(`%d.`, recipients))
	if err != nil {
		return err
	}
	b.Add(l2)
	b.Add(l)
	r.Add(b)
	r.ShowAll()
	g.recipients.Prepend(r)
	return err
}

func (g *GUI) AddRecipient(p string) {
	glib.IdleAdd(func() {
		if err := g.addRecipient(p); err != nil {
			log.Println("Interface error:", err)
		}
	})
}
