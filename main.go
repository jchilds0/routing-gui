package main

import (
	"github.com/gotk3/gotk3/gtk"
)

func main() {
	gtk.Init(nil)

	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle("Routing GUI")
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	l, _ := gtk.LabelNew("Hello, gotk3")
	win.Add(l)
	win.SetDefaultSize(800, 600)
	win.ShowAll()
	gtk.Main()
}
