package main

import (
	"log"
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
var selectRouter *router.RouterIcon
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

		addRouterButton, _ := gtk.ButtonNewWithLabel("Add Router")
		buttons.PackStart(addRouterButton, false, false, padding)

		label, _ := gtk.LabelNew("Send Message ")
		label.SetWidthChars(10)
		buttons.PackStart(label, false, false, padding)

		cell, _ := gtk.CellRendererTextNew()

		source, _ := gtk.LabelNew("Source")
		source.SetWidthChars(10)
		buttons.PackStart(source, false, false, padding)

		sourceSelect, _ := gtk.ComboBoxNewWithModel(routers.Model)
		sourceSelect.SetActive(router.ROUTER_NAME)
		sourceSelect.CellLayout.PackStart(cell, true)
		sourceSelect.CellLayout.AddAttribute(cell, "text", router.ROUTER_NAME)

		buttons.PackStart(sourceSelect, false, false, padding)

		dest, _ := gtk.LabelNew("Destination")
		dest.SetWidthChars(10)
		buttons.PackStart(dest, false, false, padding)

		destSelect, _ := gtk.ComboBoxNewWithModel(routers.Model)
		destSelect.SetActive(router.ROUTER_NAME)
		destSelect.CellLayout.PackStart(cell, true)
		destSelect.CellLayout.AddAttribute(cell, "text", router.ROUTER_NAME)

		buttons.PackStart(destSelect, false, false, padding)

		send, _ := gtk.ButtonNewWithLabel("Send")
		buttons.PackStart(send, false, false, padding)

		prevState, _ := gtk.ButtonNewWithLabel("Prev State")
		prevState.SetSensitive(false)
		buttons.PackStart(prevState, false, false, padding)

		nextState, _ := gtk.ButtonNewWithLabel("Next State")
		nextState.SetSensitive(false)
		buttons.PackStart(nextState, false, false, padding)

		broadcast, _ := gtk.ButtonNewWithLabel("Broadcast")
		broadcast.SetSensitive(false)
		buttons.PackStart(broadcast, false, false, padding)

		draw, _ := gtk.DrawingAreaNew()
		right.PackStart(draw, true, true, 0)

		draw.AddEvents(gdk.BUTTON1_MASK)
		draw.AddEvents(int(gdk.POINTER_MOTION_MASK))

		addRouterButton.Connect("clicked", func() {
			addRouter(draw)
		})

		send.Connect("clicked", func() {
			sourceIter, err := sourceSelect.GetActiveIter()
			if err != nil {
				log.Print(err)
				return
			}

			sourceID, err := router.ModelGetValue[int](routers.Model.ToTreeModel(), sourceIter, router.ROUTER_ID)
			if err != nil {
				log.Print(err)
				return
			}

			destIter, err := destSelect.GetActiveIter()
			if err != nil {
				log.Print(err)
				return
			}

			destID, err := router.ModelGetValue[int](routers.Model.ToTreeModel(), destIter, router.ROUTER_ID)
			if err != nil {
				log.Print(err)
				return
			}

			state.Start(sourceID, destID, routers)
			state.LoadState(routers)
			draw.QueueDraw()
			nextState.SetSensitive(true)
			broadcast.SetSensitive(true)

			sourceSelect.SetSensitive(false)
			destSelect.SetSensitive(false)
			send.SetSensitive(false)
		})

		prevState.Connect("clicked", func() {
			state.PrevState()
			state.LoadState(routers)
			draw.QueueDraw()

			prevState.SetSensitive(state.IsPrevState())
			nextState.SetSensitive(state.IsNextState())
		})

		nextState.Connect("clicked", func() {
			state.NextState(pipes)
			state.LoadState(routers)
			draw.QueueDraw()

			prevState.SetSensitive(state.IsPrevState())
			nextState.SetSensitive(state.IsNextState())
		})

		broadcast.Connect("clicked", func() {
			state.Broadcast(pipes)
		})

		draw.Connect("draw", func(d *gtk.DrawingArea, cr *cairo.Context) {
			routers.Draw(cr)
			pipes.Draw(cr)
		})

		draw.Connect("motion-notify-event", drawLoop)

		// testing layout
		addRouter(draw)
		addRouter(draw)
		addRouter(draw)
		addRouter(draw)

		pipes.AddConnection(1, 2)
		pipes.AddConnection(2, 3)
		pipes.AddConnection(1, 3)
		pipes.AddConnection(3, 4)
	}
}

func addRouter(draw *gtk.DrawingArea) {
	ls := router.NewLinkStateRouter(routers.MaxRouterID)
	newRouter := router.NewRouter(ls)
	routers.AddRouter(newRouter)
	draw.QueueDraw()
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
