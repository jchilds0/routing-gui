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

	routers = router.NewRouterTree(func(routerID int) *gtk.ListStore {
		return state.RouterInfo[routerID]
	})
	pipes = router.NewPipeTree(routers)
	state = router.NewRouterState()

	{
		// left box
		left, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
		box.Pack1(left, false, false)

		label, _ := gtk.HeaderBarNew()
		label.SetTitle("Routers")
		left.PackStart(label, false, false, 0)

		nb, _ := gtk.NotebookNew()
		left.PackStart(nb, true, true, 0)

		labelRouter, _ := gtk.LabelNewWithMnemonic("Routers")
		nb.AppendPage(routers.List, labelRouter)

		labelPipes, _ := gtk.LabelNewWithMnemonic("Connections")
		nb.AppendPage(pipes.Box, labelPipes)

		left.PackEnd(routers.RouterInfo, false, false, 0)
	}

	{
		// right box
		padding := uint(10)

		right, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
		box.Pack2(right, true, true)

		header, _ := gtk.HeaderBarNew()
		header.SetTitle("Network Layout")
		right.PackStart(header, false, false, 0)

		routerButtons, _ := gtk.FlowBoxNew()
		routerButtons.SetSelectionMode(gtk.SELECTION_NONE)
		routerButtons.SetColumnSpacing(10)
		routerButtons.SetRowSpacing(10)

		right.PackStart(routerButtons, false, false, padding)

		addRouterButton, _ := gtk.ButtonNewWithLabel("Add Router")
		routerButtons.Add(addRouterButton)

		label, _ := gtk.LabelNew("Send Message ")
		label.SetWidthChars(10)
		routerButtons.Add(label)

		cell, _ := gtk.CellRendererTextNew()

		box1, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
		routerButtons.Add(box1)

		source, _ := gtk.LabelNew("Source")
		source.SetWidthChars(15)
		box1.PackStart(source, false, false, 0)

		sourceSelect, _ := gtk.ComboBoxNewWithModel(routers.Model)
		sourceSelect.SetActive(router.ROUTER_NAME)
		sourceSelect.CellLayout.PackStart(cell, true)
		sourceSelect.CellLayout.AddAttribute(cell, "text", router.ROUTER_NAME)
		box1.PackStart(sourceSelect, true, true, 0)

		box2, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
		routerButtons.Add(box2)

		dest, _ := gtk.LabelNew("Destination")
		dest.SetWidthChars(15)
		box2.PackStart(dest, false, false, 0)

		destSelect, _ := gtk.ComboBoxNewWithModel(routers.Model)
		destSelect.SetActive(router.ROUTER_NAME)
		destSelect.CellLayout.PackStart(cell, true)
		destSelect.CellLayout.AddAttribute(cell, "text", router.ROUTER_NAME)
		box2.PackStart(destSelect, true, true, 0)

		send, _ := gtk.ButtonNewWithLabel("Send")
		routerButtons.Add(send)

		split, _ := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
		right.PackStart(split, false, false, 0)

		stateButtons, _ := gtk.FlowBoxNew()
		stateButtons.SetSelectionMode(gtk.SELECTION_NONE)
		stateButtons.SetColumnSpacing(10)
		stateButtons.SetRowSpacing(10)

		right.PackStart(stateButtons, false, false, padding)

		prevState, _ := gtk.ButtonNewWithLabel("Prev State")
		prevState.SetSensitive(false)
		stateButtons.Add(prevState)

		nextState, _ := gtk.ButtonNewWithLabel("Next State")
		nextState.SetSensitive(false)
		stateButtons.Add(nextState)

		broadcast, _ := gtk.ButtonNewWithLabel("Broadcast All")
		broadcast.SetSensitive(false)
		stateButtons.Add(broadcast)

		box3, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
		stateButtons.Add(box3)

		broadcastLabel, _ := gtk.LabelNew("Broadcast Router:")
		broadcastLabel.SetWidthChars(15)
		box3.PackStart(broadcastLabel, false, false, 0)

		broadcastRouterSelect, _ := gtk.ComboBoxNewWithModel(routers.Model)
		broadcastRouterSelect.SetActive(router.ROUTER_NAME)
		broadcastRouterSelect.SetSensitive(false)
		broadcastRouterSelect.CellLayout.PackStart(cell, true)
		broadcastRouterSelect.CellLayout.AddAttribute(cell, "text", router.ROUTER_NAME)
		box3.PackStart(broadcastRouterSelect, false, false, padding)

		broadcastRouter, _ := gtk.ButtonNewWithLabel("Broadcast Router")
		broadcastRouter.SetSensitive(false)
		box3.PackStart(broadcastRouter, false, false, padding)

		split, _ = gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
		right.PackStart(split, false, false, 0)

		draw, _ := gtk.DrawingAreaNew()
		right.PackStart(draw, true, true, 0)

		draw.AddEvents(gdk.BUTTON1_MASK)
		draw.AddEvents(int(gdk.POINTER_MOTION_MASK))

		addRouterButton.Connect("clicked", func() {
			addRouter(draw)
		})

		send.Connect("clicked", func() {
			model := routers.Model.ToTreeModel()
			sourceIter, err := sourceSelect.GetActiveIter()
			if err != nil {
				log.Print(err)
				return
			}

			sourceID, err := router.ModelGetValue[int](model, sourceIter, router.ROUTER_ID)
			if err != nil {
				log.Print(err)
				return
			}

			destIter, err := destSelect.GetActiveIter()
			if err != nil {
				log.Print(err)
				return
			}

			destID, err := router.ModelGetValue[int](model, destIter, router.ROUTER_ID)
			if err != nil {
				log.Print(err)
				return
			}

			state.Start(sourceID, destID, routers)
			draw.QueueDraw()

			nextState.SetSensitive(true)
			broadcast.SetSensitive(true)
			broadcastRouter.SetSensitive(true)
			broadcastRouterSelect.SetSensitive(true)

			sourceSelect.SetSensitive(false)
			destSelect.SetSensitive(false)
			send.SetSensitive(false)
		})

		prevState.Connect("clicked", func() {
			state.PrevState(routers)
			state.UpdateRouterInfo(routers)
			draw.QueueDraw()

			prevState.SetSensitive(state.IsPrevState())
			nextState.SetSensitive(state.IsNextState())
		})

		nextState.Connect("clicked", func() {
			state.RoutePacket(pipes)
			state.UpdateRouterInfo(routers)
			draw.QueueDraw()

			prevState.SetSensitive(state.IsPrevState())
			nextState.SetSensitive(state.IsNextState())
		})

		broadcast.Connect("clicked", func() {
			state.Broadcast()
			state.UpdateRouterInfo(routers)
		})

		broadcastRouter.Connect("clicked", func() {
			model := routers.Model.ToTreeModel()
			iter, err := broadcastRouterSelect.GetActiveIter()
			if err != nil {
				log.Print(err)
				return
			}

			routerID, err := router.ModelGetValue[int](model, iter, router.ROUTER_NAME)
			if err != nil {
				log.Print(err)
				return
			}

			state.BroadcastRouter(routerID)
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
