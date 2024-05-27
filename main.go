package main

import (
	"log"
	"routing-gui/gtk_utils"
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
	builder, err := gtk.BuilderNewFromFile("./gui.ui")
	if err != nil {
		log.Fatal(err)
	}

	box, err := gtk_utils.BuilderGetObject[*gtk.Paned](builder, "body")
	win.Add(box)

	routers = router.NewRouterTree(func(routerID int) *gtk.ListStore {
		return state.RouterInfo[routerID]
	})
	pipes = router.NewPipeTree(routers)
	state = router.NewRouterState()

	nb, err := gtk_utils.BuilderGetObject[*gtk.Notebook](builder, "nb")

	labelRouter, _ := gtk.LabelNewWithMnemonic("Routers")
	nb.AppendPage(routers.List, labelRouter)

	labelPipes, _ := gtk.LabelNewWithMnemonic("Connections")
	nb.AppendPage(pipes.Box, labelPipes)

	left, err := gtk_utils.BuilderGetObject[*gtk.Box](builder, "left")
	if err != nil {
		log.Fatal(err)
	}

	left.PackStart(routers.RouterInfo, false, false, 0)

	/* Prepare Message */
	cell, _ := gtk.CellRendererTextNew()

	addRouterButton, err := gtk_utils.BuilderGetObject[*gtk.Button](builder, "add-router")
	if err != nil {
		log.Fatal(err)
	}

	sourceSelect, err := gtk_utils.BuilderGetObject[*gtk.ComboBox](builder, "source-select")
	if err != nil {
		log.Fatal(err)
	}

	sourceSelect.SetModel(routers.Model)
	sourceSelect.SetActive(router.ROUTER_NAME)
	sourceSelect.CellLayout.PackStart(cell, true)
	sourceSelect.CellLayout.AddAttribute(cell, "text", router.ROUTER_NAME)

	destSelect, err := gtk_utils.BuilderGetObject[*gtk.ComboBox](builder, "dest-select")
	if err != nil {
		log.Fatal(err)
	}

	destSelect.SetModel(routers.Model)
	destSelect.SetActive(router.ROUTER_NAME)
	destSelect.CellLayout.PackStart(cell, true)
	destSelect.CellLayout.AddAttribute(cell, "text", router.ROUTER_NAME)

	send, err := gtk_utils.BuilderGetObject[*gtk.Button](builder, "send")
	if err != nil {
		log.Fatal(err)
	}

	/* Send Message */

	broadcast, err := gtk_utils.BuilderGetObject[*gtk.Button](builder, "broadcast")
	if err != nil {
		log.Fatal(err)
	}

	broadcastRouterSelect, err := gtk_utils.BuilderGetObject[*gtk.ComboBox](builder, "broadcast-select")
	if err != nil {
		log.Fatal(err)
	}

	broadcastRouterSelect.SetModel(routers.Model)
	broadcastRouterSelect.SetActive(router.ROUTER_NAME)
	broadcastRouterSelect.CellLayout.PackStart(cell, true)
	broadcastRouterSelect.CellLayout.AddAttribute(cell, "text", router.ROUTER_NAME)

	broadcastRouter, err := gtk_utils.BuilderGetObject[*gtk.Button](builder, "broadcast-router")
	if err != nil {
		log.Fatal(err)
	}

	detect, err := gtk_utils.BuilderGetObject[*gtk.Button](builder, "detect")
	if err != nil {
		log.Fatal(err)
	}

	detectRouterSelect, err := gtk_utils.BuilderGetObject[*gtk.ComboBox](builder, "detect-select")
	if err != nil {
		log.Fatal(err)
	}

	detectRouterSelect.SetModel(routers.Model)
	detectRouterSelect.SetActive(router.ROUTER_NAME)
	detectRouterSelect.CellLayout.PackStart(cell, true)
	detectRouterSelect.CellLayout.AddAttribute(cell, "text", router.ROUTER_NAME)

	detectRouter, err := gtk_utils.BuilderGetObject[*gtk.Button](builder, "detect-router")
	if err != nil {
		log.Fatal(err)
	}

	prevState, err := gtk_utils.BuilderGetObject[*gtk.Button](builder, "prev-state")
	if err != nil {
		log.Fatal(err)
	}

	nextHop, err := gtk_utils.BuilderGetObject[*gtk.Button](builder, "next-hop")
	if err != nil {
		log.Fatal(err)
	}

	draw, err := gtk_utils.BuilderGetObject[*gtk.DrawingArea](builder, "draw")
	if err != nil {
		log.Fatal(err)
	}

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

		sourceID, err := gtk_utils.ModelGetValue[int](model, sourceIter, router.ROUTER_ID)
		if err != nil {
			log.Print(err)
			return
		}

		destIter, err := destSelect.GetActiveIter()
		if err != nil {
			log.Print(err)
			return
		}

		destID, err := gtk_utils.ModelGetValue[int](model, destIter, router.ROUTER_ID)
		if err != nil {
			log.Print(err)
			return
		}

		state.Start(sourceID, destID, routers)
		draw.QueueDraw()

		setSensitive(false, sourceSelect, destSelect)
		setSensitive(true,
			prevState, nextHop,
			broadcast, broadcastRouterSelect,
			broadcastRouter, detect,
			detectRouterSelect, detectRouter,
		)
	})

	prevState.Connect("clicked", func() {
		state.PrevState(routers)
		state.UpdateRouterInfo(routers)
		draw.QueueDraw()

		prevState.SetSensitive(state.IsPrevState())
		nextHop.SetSensitive(state.IsNextState())
	})

	nextHop.Connect("clicked", func() {
		state.NewState()

		err := state.RoutePacket(pipes)
		if err != nil {
			log.Print(err)
		}

		state.UpdateRouterInfo(routers)
		draw.QueueDraw()

		prevState.SetSensitive(state.IsPrevState())
		nextHop.SetSensitive(state.IsNextState())
	})

	broadcast.Connect("clicked", func() {
		state.NewState()

		for id, r := range routers.Routers {
			if r == nil {
				continue
			}

			state.BroadcastRouter(id)
		}

		state.UpdateRouterInfo(routers)
		prevState.SetSensitive(state.IsPrevState())
	})

	broadcastRouter.Connect("clicked", func() {
		state.NewState()

		model := routers.Model.ToTreeModel()
		iter, err := broadcastRouterSelect.GetActiveIter()
		if err != nil {
			log.Print(err)
			return
		}

		routerID, err := gtk_utils.ModelGetValue[int](model, iter, router.ROUTER_ID)
		if err != nil {
			log.Print(err)
			return
		}

		state.BroadcastRouter(routerID)
		state.UpdateRouterInfo(routers)
		prevState.SetSensitive(state.IsPrevState())
	})

	detect.Connect("clicked", func() {
		state.NewState()

		for id, r := range routers.Routers {
			if r == nil {
				continue
			}

			state.DetectAdjacent(id, pipes)
		}

		state.UpdateRouterInfo(routers)
		prevState.SetSensitive(state.IsPrevState())
	})

	detectRouter.Connect("clicked", func() {
		state.NewState()

		model := routers.Model.ToTreeModel()
		iter, err := detectRouterSelect.GetActiveIter()
		if err != nil {
			log.Print(err)
			return
		}

		routerID, err := gtk_utils.ModelGetValue[int](model, iter, router.ROUTER_ID)
		if err != nil {
			log.Print(err)
			return
		}

		state.DetectAdjacent(routerID, pipes)
		state.UpdateRouterInfo(routers)
		prevState.SetSensitive(state.IsPrevState())
	})

	draw.Connect("draw", func(d *gtk.DrawingArea, cr *cairo.Context) {
		routers.Draw(cr)
		pipes.Draw(cr)
	})

	draw.Connect("motion-notify-event", drawLoop)

	setSensitive(true, sourceSelect, destSelect)
	setSensitive(false,
		prevState, nextHop,
		broadcast, broadcastRouterSelect,
		broadcastRouter, detect,
		detectRouterSelect, detectRouter,
	)

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

func addRouter(draw *gtk.DrawingArea) {
	ls := router.NewLinkStateRouter(routers.MaxRouterID)
	newRouter := router.NewRouter(ls)
	routers.AddRouter(newRouter)
	draw.QueueDraw()
}

type sensitive interface {
	SetSensitive(bool)
}

func setSensitive(val bool, args ...sensitive) {
	for _, w := range args {
		w.SetSensitive(val)
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
