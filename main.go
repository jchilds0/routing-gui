package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"routing-gui/gtk_utils"
	"routing-gui/protocol"
	"routing-gui/router"
	"strconv"
	"strings"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

var layout = flag.String("f", "", "load network from csv file")

func main() {
	flag.Parse()
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

	stateHeader, err := gtk_utils.BuilderGetObject[*gtk.HeaderBar](builder, "state-title")
	if err != nil {
		log.Fatal(err)
	}

	stateTree, err := gtk_utils.BuilderGetObject[*gtk.TreeView](builder, "router-state")
	if err != nil {
		log.Fatal(err)
	}

	routers = router.NewRouterTree(stateHeader, stateTree, func(routerID int) *gtk.TreeModel {
		if state == nil {
			return nil
		}

		tree := state.RouterInfo[routerID]
		if tree == nil {
			return nil
		}

		return tree.ToTreeModel()
	})
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

	routers.SetupTreeColumns(routerList)

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

	/* State */
	stateList, err := gtk_utils.BuilderGetObject[*gtk.TreeView](builder, "state-list")
	if err != nil {
		log.Fatal(err)
	}

	cell, _ = gtk.CellRendererTextNew()
	col, _ := gtk.TreeViewColumnNewWithAttribute("State Description", cell, "text", router.STATE_DESC)
	stateList.AppendColumn(col)
	stateList.SetModel(state.Model)

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
		newRouter("Router", "127.0.0.1")

		draw.QueueDraw()
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

	/* State */
	stateList.Connect("row-activated",
		func(tree *gtk.TreeView, path *gtk.TreePath, column *gtk.TreeViewColumn) {
			iter, err := state.Model.GetIter(path)
			if err != nil {
				log.Printf("Error selecting state: %s", err)
				return
			}

			model := state.Model.ToTreeModel()
			id, err := gtk_utils.ModelGetValue[int](model, iter, router.STATE_ID)
			if err != nil {
				log.Printf("Error selecting state: %s", err)
				return
			}

			state.LoadState(id)
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

		state.UpdateRouterInfo()
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
	draw.Connect("button-press-event", pressLoop)

	setSensitive(true, sourceSelect, destSelect)
	setSensitive(false,
		prevState, nextHopButton,
		broadcast, broadcastRouterSelect,
		broadcastRouterButton, detectButton,
		detectRouterSelect, detectRouterButton,
	)

	// load layout
	if *layout != "" {
		loadLayout(*layout)

		draw.QueueDraw()
	}
}

type sensitive interface {
	SetSensitive(bool)
}

func setSensitive(val bool, args ...sensitive) {
	for _, w := range args {
		w.SetSensitive(val)
	}
}

func newRouter(name, ip string) *router.RouterIcon {
	if routers == nil {
		log.Fatal("Routers not initialised")
	}

	r := protocol.NewLinkStateRouter(routers.MaxRouterID)

	rIcon := router.NewRouter(r)
	rIcon.Name = name
	rIcon.IP = ip
	routers.AddRouter(rIcon)

	return rIcon
}

func loadLayout(fileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Printf("Error opening layout file: %s", err)
		return
	}

	if routers == nil {
		log.Fatal("Routers not initialised")
	}

	if pipes == nil {
		log.Fatal("Pipes not initialised")
	}

	i := 1
	s := bufio.NewScanner(file)
	for s.Scan() {
		line := s.Text()
		words := strings.Split(line, ",")

		switch len(words) {
		case 2:
			newRouter(words[0], words[1])
		case 3:
			r1Name := strings.TrimSpace(words[0])
			r1 := routers.GetRouterIcon(r1Name)
			if r1 == nil {
				log.Printf("Line %d: %s", i, line)
				log.Printf("Error: router name %s does not exist", r1Name)
				continue
			}

			r2Name := strings.TrimSpace(words[1])
			r2 := routers.GetRouterIcon(r2Name)
			if r2 == nil {
				log.Printf("Line %d: %s", i, line)
				log.Printf("Error: router name %s does not exist", r2Name)
				continue
			}

			str := strings.TrimSpace(words[2])
			weight, err := strconv.Atoi(str)
			if err != nil {
				log.Printf("Line %d: %s", i, line)
				log.Printf("Error converting weight: %s", err)
				continue
			}

			pipes.AddConnection(r1.RouterID, r2.RouterID, weight)
		}

		i++
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

	if routers == nil {
		return
	}

	if state == nil {
		return
	}

	for _, r := range routers.Routers {
		if r == nil {
			continue
		}

		if !r.Contains(b.X(), b.Y()) {
			continue
		}

		if state.RouterInfo[r.RouterID] == nil {
			continue
		}

		routers.SetRouterState(r.RouterID, r.Name, state.RouterInfo[r.RouterID].ToTreeModel())
	}

	d.QueueDraw()
}
