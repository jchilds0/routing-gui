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
var routers *router.RouterTree
var pipes *router.PipeTree
var state *router.RouterState

func buildWindow(win *gtk.Window) {
	box, _ := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)
	win.Add(box)

	routers = router.NewRouterTree()
	pipes = router.NewPipeTree(routers)
	state = router.NewRouterState()

	{
		// left box
		left, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
		box.Pack1(left, false, false)

		label, _ := gtk.LabelNew("Routers")
		left.PackStart(label, false, false, 0)

		nb, _ := gtk.NotebookNew()
		left.PackStart(nb, true, true, 0)

		labelRouter, _ := gtk.LabelNewWithMnemonic("Routers")
		nb.AppendPage(routers.List, labelRouter)

		labelPipes, _ := gtk.LabelNewWithMnemonic("Connections")
		nb.AppendPage(pipes.Box, labelPipes)
	}

	{
		// right box
		padding := uint(10)

		right, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
		box.Pack2(right, true, true)

		buttons, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
		right.PackStart(buttons, false, false, padding)

		addRouter, _ := gtk.ButtonNewWithLabel("Add Router")
		buttons.PackStart(addRouter, false, false, padding)

		prevState, _ := gtk.ButtonNewWithLabel("Prev State")
		buttons.PackStart(prevState, false, false, padding)

		nextState, _ := gtk.ButtonNewWithLabel("Next State")
		buttons.PackStart(nextState, false, false, padding)

		draw, _ := gtk.DrawingAreaNew()
		right.PackStart(draw, true, true, 0)

		draw.AddEvents(gdk.BUTTON1_MASK)
		draw.AddEvents(int(gdk.POINTER_MOTION_MASK))

		addRouter.Connect("clicked", func() {
			newRouter := router.NewRouter()
			routers.AddRouter(newRouter)
			draw.QueueDraw()
		})

		prevState.Connect("clicked", func() {
			state.UpdateState(state.CurrentState-1, routers)
			draw.QueueDraw()
		})

		nextState.Connect("clicked", func() {
			state.UpdateState(state.CurrentState+1, routers)
			draw.QueueDraw()
		})

		draw.Connect("draw", func(d *gtk.DrawingArea, cr *cairo.Context) {
			routers.Draw(cr)
			pipes.Draw(cr)
		})

		draw.Connect("motion-notify-event", drawLoop)
	}
}

func drawLoop(d *gtk.DrawingArea, event *gdk.Event) {
	b := gdk.EventButtonNewFromEvent(event)
	defer func() {
		mouseX = b.X()
		mouseY = b.Y()
	}()

	if b.State()&uint(gdk.BUTTON_PRESS_MASK) == 0 {
		selectRouter = nil
		return
	}

	if selectRouter != nil {
		selectRouter.UpdatePos(b.X()-mouseX, b.Y()-mouseY)
		d.QueueDraw()

		return
	}

	for _, r := range routers.Routers {
		if r == nil {
			continue
		}

		if r.Contains(b.X(), b.Y()) {
			selectRouter = r
		}
	}
}

func pressLoop(d *gtk.DrawingArea, event *gdk.Event) {
	b := gdk.EventButtonNewFromEvent(event)

	for _, r := range routers.Routers {
		if r == nil {
			continue
		}

		r.Selected = r.Contains(b.X(), b.Y()) && !r.Selected
	}

	d.QueueDraw()
}
