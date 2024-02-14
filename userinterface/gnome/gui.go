package gnome

import (
	"fmt"

	_ "embed" // for static assets

	"github.com/dkotik/mdsend/loggers"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

var (
	_ loggers.Logger = &GUI{}

	//go:embed interface.glade
	interfaceGlade []byte

	//go:embed style.css
	styleCSS []byte
)

// GUI presents a graphical user interface that satisfies the Logger interface.
type GUI struct {
	*gtk.ApplicationWindow
	Send *gtk.Button

	recipients  *gtk.ListBox
	progress    *gtk.ProgressBar
	letter      *gtk.TextBuffer
	deliveries  *gtk.TextBuffer
	errors      *gtk.TextBuffer
	total, done uint
}

func (g *GUI) Open(p string) error {
	return nil
}

func (g *GUI) Close() error {
	return nil
}

func (g *GUI) SetTotal(total uint) {
	g.done = 0
	g.total = total
	g.progress.SetFraction(0)
}

func (g *GUI) SetLetter(s string) {
	g.letter.SetText(s)
}

func (g *GUI) Progress(s string) {
	glib.IdleAdd(func() {
		if g.done >= g.total {
			g.progress.SetText(fmt.Sprintf(`100%% %s`, s))
			return
		}
		g.done++
		done := float64(g.done) / float64(g.total)
		g.progress.SetText(fmt.Sprintf(`%d%% %s`, int(done*100), s))
		g.progress.SetFraction(done)
	})
}

func (g *GUI) LogSkip(s string, v ...interface{}) {
	g.LogFail(s, v...)
}

func (g *GUI) LogInfo(s string, v ...interface{}) {
	message := fmt.Sprintf(s, v...)
	glib.IdleAdd(func() {
		g.deliveries.Insert(g.deliveries.GetEndIter(), "\n"+message)
	})
	g.Progress(message)
}

func (g *GUI) LogSent(s string, v ...interface{}) {
	g.LogInfo(s, v...)
}

func (g *GUI) LogWarn(s string, v ...interface{}) {
	g.LogInfo(s, v...)
}

func (g *GUI) LogFail(s string, v ...interface{}) {
	message := "Error: " + fmt.Sprintf(s, v...)
	glib.IdleAdd(func() {
		g.errors.Insert(g.errors.GetEndIter(), "\n"+message)
	})
	g.Progress(message)
}

func (g *GUI) LogTest(s string, v ...interface{}) {
	g.LogInfo(s, v...)
}

// Load creates a Window from stored glade definitions.
func Load() (*GUI, error) {
	gtk.Init(nil)
	b, err := gtk.BuilderNew()
	if err != nil {
		return nil, err
	}

	style, _ := gtk.CssProviderNew()
	screen, _ := gdk.ScreenGetDefault()
	gtk.AddProviderForScreen(screen, style, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)

	b.AddFromString(string(interfaceGlade))
	style.LoadFromData(string(styleCSS))

	obj, err := b.GetObject(`root`)
	if err != nil {
		return nil, err
	}
	gui := &GUI{ApplicationWindow: obj.(*gtk.ApplicationWindow)}
	obj, err = b.GetObject(`recipients`)
	if err != nil {
		return gui, err
	}
	gui.recipients = obj.(*gtk.ListBox)
	obj, err = b.GetObject(`letter`)
	if err != nil {
		return gui, err
	}
	gui.letter = obj.(*gtk.TextBuffer)
	obj, err = b.GetObject(`deliveries`)
	if err != nil {
		return gui, err
	}
	gui.deliveries = obj.(*gtk.TextBuffer)
	obj, err = b.GetObject(`errors`)
	if err != nil {
		return gui, err
	}
	gui.errors = obj.(*gtk.TextBuffer)
	obj, err = b.GetObject(`progress`)
	if err != nil {
		return gui, err
	}
	gui.progress = obj.(*gtk.ProgressBar)

	obj, err = b.GetObject(`send`)
	if err != nil {
		return gui, err
	}
	gui.Send = obj.(*gtk.Button)

	// gui.SetTotal(15)
	// go func() {
	// 	for {
	// 		gui.LogInfo(fmt.Sprintf(`TEst %d`, time.Now().Unix()))
	// 		gui.AddRecipient(`p string`)
	// 		time.Sleep(time.Millisecond * 30)
	// 		if gui.total == gui.done {
	// 			break
	// 		}
	// 	}
	// }()
	return gui, err
}
