package main

import (
	"routing-gui/router"

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

var mouseX, mouseY float64
var selectRouter *router.Router
var routers = make([]*router.Router, 0, 100)

func buildWindow(win *gtk.Window) {
	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	win.Add(box)

	grid, _ := gtk.GridNew()
	box.PackStart(grid, false, false, 0)

	addRouter, _ := gtk.ButtonNewWithLabel("Add Router")
	grid.Attach(addRouter, 10, 10, 10, 10)

	addRouter.Connect("clicked", func() {
		newRouter := router.NewRouter()
		routers = append(routers, newRouter)
	})

	draw, _ := gtk.DrawingAreaNew()
	box.PackStart(draw, true, true, 0)

	draw.AddEvents(gdk.BUTTON1_MASK)
	draw.AddEvents(int(gdk.POINTER_MOTION_MASK))

	draw.Connect("draw", func(d *gtk.DrawingArea, cr *cairo.Context) {
		for _, r := range routers {
			r.Draw(cr)
		}
	})

	draw.Connect("motion-notify-event", drawLoop)
}

func drawLoop(d *gtk.DrawingArea, event *gdk.Event) {
	b := gdk.EventButtonNewFromEvent(event)
	defer func() {
		mouseX = b.X()
		mouseY = b.Y()
	}()

	if b.State() != uint(gdk.BUTTON_PRESS_MASK) {
		selectRouter = nil
		return
	}

	if selectRouter != nil {
		selectRouter.UpdatePos(b.X()-mouseX, b.Y()-mouseY)
		d.QueueDraw()

		return
	}

	for _, r := range routers {
		if r.Contains(b.X(), b.Y()) {
			selectRouter = r
		}
	}
}
