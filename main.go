package main

import (
	"fmt"
	"log"
	"routing-gui/gtk_utils"
	"routing-gui/protocol"
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
	win.SetDecorated(true)
	win.ShowAll()
	gtk.Main()
}

var mouseX, mouseY float64
var selectRouter *router.RouterIcon
var routers *router.RouterTree
var pipes *router.PipeTree
var state *router.RouterState

func buildWindow(win *gtk.Window) {
	cell, _ := gtk.CellRendererTextNew()

	builder, err := gtk.BuilderNewFromFile("./gui.ui")
	if err != nil {
		log.Fatal(err)
	}

	box, err := gtk_utils.BuilderGetObject[*gtk.Paned](builder, "body")
	win.Add(box)

	routers = router.NewRouterTree()
	pipes = router.NewPipeTree(routers)
	state = router.NewRouterState(pipes)

	/* Routers */

	addRouterButton, err := gtk_utils.BuilderGetObject[*gtk.Button](builder, "add-router")
	if err != nil {
		log.Fatal(err)
	}

	routerSelect, err := gtk_utils.BuilderGetObject[*gtk.ComboBox](builder, "router-select")
	if err != nil {
		log.Fatal(err)
	}

	routerSelect.SetModel(routers.Model)
	routerSelect.CellLayout.PackStart(cell, true)
	routerSelect.CellLayout.AddAttribute(cell, "text", router.ROUTER_NAME)
	routerSelect.SetActive(router.ROUTER_NAME)

	removeRouterButton, err := gtk_utils.BuilderGetObject[*gtk.Button](builder, "remove-router")
	if err != nil {
		log.Fatal(err)
	}

	routerList, err := gtk_utils.BuilderGetObject[*gtk.TreeView](builder, "router-list")
	if err != nil {
		log.Fatal(err)
	}

	routerList.SetModel(routers.Model)
	routerList.SetActivateOnSingleClick(true)
	routers.AddColumns(routerList, func(routerID int) gtk.ITreeModel {
		return state.RouterInfo[routerID]
	})

	/* Connections */

	pipeSelect1, err := gtk_utils.BuilderGetObject[*gtk.ComboBox](builder, "pipe-router-1")
	if err != nil {
		log.Fatal(err)
	}

	pipeSelect1.SetModel(routers.Model)
	pipeSelect1.CellLayout.PackStart(cell, true)
	pipeSelect1.CellLayout.AddAttribute(cell, "text", router.ROUTER_NAME)
	pipeSelect1.SetActive(router.ROUTER_NAME)

	pipeSelect2, err := gtk_utils.BuilderGetObject[*gtk.ComboBox](builder, "pipe-router-2")
	if err != nil {
		log.Fatal(err)
	}

	pipeSelect2.SetModel(routers.Model)
	pipeSelect2.CellLayout.PackStart(cell, true)
	pipeSelect2.CellLayout.AddAttribute(cell, "text", router.ROUTER_NAME)
	pipeSelect2.SetActive(router.ROUTER_NAME)

	addPipe, err := gtk_utils.BuilderGetObject[*gtk.Button](builder, "add-pipe")
	if err != nil {
		log.Fatal(err)
	}

	removePipe, err := gtk_utils.BuilderGetObject[*gtk.Button](builder, "remove-pipe")
	if err != nil {
		log.Fatal(err)
	}

	pipeList, err := gtk_utils.BuilderGetObject[*gtk.TreeView](builder, "pipe-list")
	if err != nil {
		log.Fatal(err)
	}

	pipeList.SetModel(pipes.Model)
	pipes.AddColumns(pipeList)

	/* Prepare Message */

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

	broadcastRouterButton, err := gtk_utils.BuilderGetObject[*gtk.Button](builder, "broadcast-router")
	if err != nil {
		log.Fatal(err)
	}

	detectButton, err := gtk_utils.BuilderGetObject[*gtk.Button](builder, "detect")
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

	detectRouterButton, err := gtk_utils.BuilderGetObject[*gtk.Button](builder, "detect-router")
	if err != nil {
		log.Fatal(err)
	}

	prevState, err := gtk_utils.BuilderGetObject[*gtk.Button](builder, "prev-state")
	if err != nil {
		log.Fatal(err)
	}

	nextHopButton, err := gtk_utils.BuilderGetObject[*gtk.Button](builder, "next-hop")
	if err != nil {
		log.Fatal(err)
	}

	draw, err := gtk_utils.BuilderGetObject[*gtk.DrawingArea](builder, "draw")
	if err != nil {
		log.Fatal(err)
	}

	draw.AddEvents(gdk.BUTTON1_MASK)
	draw.AddEvents(int(gdk.POINTER_MOTION_MASK))

	/* Routers */
	addRouterButton.Connect("clicked", func() {
		addRouter(draw, "Router", "127.0.0.1")
	})

	removeRouterButton.Connect("clicked", func() {
		iter, err := routerSelect.GetActiveIter()
		if err != nil {
			log.Printf("Error getting router 1: %s", err)
			return
		}

		r, err := routers.GetRouter(iter)
		if err != nil {
			log.Printf("Error getting router 1: %s", err)
			return
		}

		routers.Model.Remove(iter)
		delete(routers.Routers, r.RouterID)

		// remove connections
		for id := range pipes.PipeIter {
			r1 := pipes.Router1[id]
			r2 := pipes.Router2[id]

			if r1 == r.RouterID || r2 == r.RouterID {
				pipes.RemoveConnection(pipes.PipeIter[id])
			}
		}
	})

	/* Connections */
	addPipe.Connect("clicked", func() {
		iter1, err := pipeSelect1.GetActiveIter()
		if err != nil {
			log.Printf("Error getting router 1: %s", err)
			return
		}

		r1, err := routers.GetRouter(iter1)
		if err != nil {
			log.Printf("Error getting router 1: %s", err)
			return
		}

		iter2, err := pipeSelect2.GetActiveIter()
		if err != nil {
			log.Printf("Error getting router 2: %s", err)
			return
		}

		r2, err := routers.GetRouter(iter2)
		if err != nil {
			log.Printf("Error getting router 2: %s", err)
			return
		}

		if r1.RouterID == r2.RouterID {
			log.Print("Routers are the same, not adding connection")
			return
		}

		err = pipes.AddConnection(r1.RouterID, r2.RouterID, 1)
		if err != nil {
			log.Printf("Error adding connection: %s", err)
			return
		}
	})

	removePipe.Connect("clicked", func() {
		iter1, err := pipeSelect1.GetActiveIter()
		if err != nil {
			log.Printf("Error getting router 1: %s", err)
			return
		}

		r1, err := routers.GetRouter(iter1)
		if err != nil {
			log.Printf("Error getting router 1: %s", err)
			return
		}

		iter2, err := pipeSelect2.GetActiveIter()
		if err != nil {
			log.Printf("Error getting router 2: %s", err)
			return
		}

		r2, err := routers.GetRouter(iter2)
		if err != nil {
			log.Printf("Error getting router 2: %s", err)
			return
		}

		for id := range pipes.PipeIter {
			router1 := pipes.Router1[id]
			router2 := pipes.Router2[id]

			if router1 == r1.RouterID && router2 == r2.RouterID {
				err := pipes.RemoveConnection(pipes.PipeIter[id])
				if err != nil {
					log.Print(err)
				}
			}

			if router2 == r1.RouterID && router1 == r2.RouterID {
				err := pipes.RemoveConnection(pipes.PipeIter[id])
				if err != nil {
					log.Print(err)
				}
			}
		}
	})

	/* Message */

	send.Connect("clicked", func() {
		sourceIter, err := sourceSelect.GetActiveIter()
		if err != nil {
			log.Print(err)
			return
		}

		source, err := routers.GetRouter(sourceIter)
		if err != nil {
			log.Print(err)
			return
		}

		destIter, err := destSelect.GetActiveIter()
		if err != nil {
			log.Print(err)
			return
		}

		dest, err := routers.GetRouter(destIter)
		if err != nil {
			log.Print(err)
			return
		}

		state.Start(source.RouterID, dest.RouterID)
		draw.QueueDraw()

		setSensitive(false, sourceSelect, destSelect)
		setSensitive(true, prevState, nextHopButton,
			broadcast, broadcastRouterSelect,
			broadcastRouterButton, detectButton,
			detectRouterSelect, detectRouterButton,
		)
	})

	prevState.Connect("clicked", func() {
		state.PrevState()
		state.UpdateRouterInfo()
		draw.QueueDraw()

		prevState.SetSensitive(state.IsPrevState())
		nextHopButton.SetSensitive(state.IsNextState())
	})

	nextHopButton.Connect("clicked", func() {
		state.NewState("Next Hop")

		err := state.RoutePacket()
		if err != nil {
			log.Print(err)
		}

		state.UpdateRouterInfo()
		draw.QueueDraw()

	})

	broadcast.Connect("clicked", func() {
		state.NewState("Broadcast All Routers")

		for id, r := range routers.Routers {
			if r == nil {
				continue
			}

			state.BroadcastRouter(id)
		}

		state.UpdateRouterInfo()
	})

	broadcastRouterButton.Connect("clicked", func() {
		iter, err := broadcastRouterSelect.GetActiveIter()
		if err != nil {
			log.Print(err)
			return
		}

		r, err := routers.GetRouter(iter)
		if err != nil {
			log.Print(err)
			return
		}

		state.NewState(fmt.Sprintf("Broadcast Router %s", r.Name))
		state.BroadcastRouter(r.RouterID)
		state.UpdateRouterInfo()
	})

	detectButton.Connect("clicked", func() {
		state.NewState("Detect All Routers")

		for id, r := range routers.Routers {
			if r == nil {
				continue
			}

			state.DetectAdjacent(id)
		}

		state.UpdateRouterInfo()
	})

	detectRouterButton.Connect("clicked", func() {
		iter, err := detectRouterSelect.GetActiveIter()
		if err != nil {
			log.Print(err)
			return
		}

		r, err := routers.GetRouter(iter)
		if err != nil {
			log.Print(err)
			return
		}

		state.NewState(fmt.Sprintf("Detect Router %s", r.Name))
		state.DetectAdjacent(r.RouterID)
		state.UpdateRouterInfo()
	})

	draw.Connect("draw", func(d *gtk.DrawingArea, cr *cairo.Context) {
		routers.Draw(cr)
		pipes.Draw(cr)
	})

	draw.Connect("motion-notify-event", drawLoop)

	setSensitive(true, sourceSelect, destSelect)
	setSensitive(false,
		prevState, nextHopButton,
		broadcast, broadcastRouterSelect,
		broadcastRouterButton, detectButton,
		detectRouterSelect, detectRouterButton,
	)

	// testing layout
	addRouter(draw, "A", "127.0.0.1")
	addRouter(draw, "B", "127.0.0.1")
	addRouter(draw, "C", "127.0.0.1")
	addRouter(draw, "D", "127.0.0.1")
	addRouter(draw, "E", "127.0.0.1")
	addRouter(draw, "F", "127.0.0.1")
	addRouter(draw, "G", "127.0.0.1")
	addRouter(draw, "H", "127.0.0.1")

	pipes.AddConnection(1, 2, 2) // A -- B
	pipes.AddConnection(2, 3, 7) // B -- C
	pipes.AddConnection(3, 4, 3) // C -- D
	pipes.AddConnection(2, 5, 2) // B -- E
	pipes.AddConnection(1, 7, 6) // A -- G
	pipes.AddConnection(7, 5, 1) // G -- E
	pipes.AddConnection(5, 6, 2) // E -- F
	pipes.AddConnection(6, 8, 2) // F -- H
	pipes.AddConnection(7, 8, 4) // G -- H
	pipes.AddConnection(6, 3, 3) // F -- C
	pipes.AddConnection(8, 4, 2) // H -- D

}

type sensitive interface {
	SetSensitive(bool)
}

func setSensitive(val bool, args ...sensitive) {
	for _, w := range args {
		w.SetSensitive(val)
	}
}

func addRouter(draw *gtk.DrawingArea, name string, ip string) {
	ls := protocol.NewLinkStateRouter(routers.MaxRouterID)
	newRouter := router.NewRouter(ls)
	newRouter.Name = name
	newRouter.IP = ip

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
