package main

import (
	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

func main() {
	gtk.Init(nil)

	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle("Routing GUI")
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	buildWindow(win)

	win.SetDefaultSize(800, 600)
	win.ShowAll()
	gtk.Main()
}

func buildWindow(win *gtk.Window) {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	win.Add(box)

	draw, _ := gtk.DrawingAreaNew()
	box.PackStart(draw, true, true, 0)

	router := &Router{
		x: 10, y: 10, w: 100, h: 100,
	}

	draw.AddEvents(gdk.BUTTON1_MASK)
	draw.AddEvents(int(gdk.POINTER_MOTION_MASK))

	draw.Connect("draw", func(d *gtk.DrawingArea, cr *cairo.Context) {
		router.Draw(cr)
	})

	draw.Connect("motion-notify-event", func(d *gtk.DrawingArea, event *gdk.Event) {
		b := gdk.EventButtonNewFromEvent(event)

		if b.State() != uint(gdk.BUTTON_PRESS_MASK) {
			return
		}

		router.x = b.X()
		router.y = b.Y()
		d.QueueDraw()
	})
}

type Router struct {
	x float64
	y float64
	w float64
	h float64
}

func (r *Router) Draw(cr *cairo.Context) {
	cr.Rectangle(r.x, r.y, r.w, r.h)
	cr.SetSourceRGB(0.5, 0.5, 1)
	cr.Fill()
}
