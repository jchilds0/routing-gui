package router

import (
	"fmt"
	"log"
	"routing-gui/gtk_utils"
	"strconv"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const logState = true

type Path struct {
	DestID    int
	NextHopID int
	Dist      int
}

type Router interface {
	RoutePacket(int) (int, error)
	Broadcast() map[int]int
	Recieve(int, map[int]int)
	Info() []Path
	AddRouter(int, int)
	RemoveRouter(int)
	Copy() Router
}

type State struct {
	currentRouter int
	selected      map[int]bool
	routers       map[int]Router
}

func NewState() *State {
	s := &State{}

	s.selected = make(map[int]bool, 30)
	s.routers = make(map[int]Router, 30)

	return s
}

func NewStateFromTree(rTree *RouterTree) *State {
	s := NewState()

	for id, r := range rTree.Routers {
		s.routers[id] = r.Router.Copy()
	}

	return s
}

func NewStateFromState(s *State) *State {
	newState := NewState()
	newState.currentRouter = s.currentRouter

	for id, r := range s.routers {
		if r == nil {
			continue
		}

		newState.routers[id] = r.Copy()
	}

	for id := range s.selected {
		newState.selected[id] = s.selected[id]
	}

	return newState
}

func (s *State) DetectAdjacent(pTree *PipeTree, routerID int) {
	router := s.routers[routerID]

	for i := range pTree.Router1 {
		if pTree.Router1[i] == routerID {
			router.AddRouter(pTree.Router2[i], pTree.Weight[i])
		}

		if pTree.Router2[i] == routerID {
			router.AddRouter(pTree.Router1[i], pTree.Weight[i])
		}
	}
}

func (s *State) Broadcast(id1 int) {
	router1 := s.routers[id1]
	if router1 == nil {
		return
	}
	msg := router1.Broadcast()

	for id2, router2 := range s.routers {
		if router2 == nil {
			return
		}
		if id1 == id2 {
			continue
		}

		router2.Recieve(id1, msg)
	}
}

const (
	INFO_DEST_NAME = iota
	INFO_DEST_IP
	INFO_NEXT_NAME
	INFO_NEXT_IP
	INFO_DIST
)

const (
	STATE_ID = iota
	STATE_DESC
)

type RouterState struct {
	rTree      *RouterTree
	pTree      *PipeTree
	state      map[int]*State
	stateIter  map[int]*gtk.TreeIter
	currentID  int
	nextID     int
	destID     int
	RouterInfo map[int]*gtk.TreeModelSort
	StateInfo  *gtk.TreeView
	infoModel  *gtk.TreeStore
}

func NewRouterState(pTree *PipeTree) *RouterState {
	rs := &RouterState{
		rTree:     pTree.Routers,
		pTree:     pTree,
		currentID: 0,
		nextID:    1,
	}

	rs.state = make(map[int]*State, 30)
	rs.stateIter = make(map[int]*gtk.TreeIter, 30)
	rs.RouterInfo = make(map[int]*gtk.TreeModelSort, 30)

	{
		rs.StateInfo, _ = gtk.TreeViewNew()
		rs.infoModel, _ = gtk.TreeStoreNew(
			glib.TYPE_INT,
			glib.TYPE_STRING,
		)

		rs.StateInfo.SetModel(rs.infoModel)

		cell, _ := gtk.CellRendererTextNew()
		col, _ := gtk.TreeViewColumnNewWithAttribute("State Description", cell, "text", STATE_DESC)
		rs.StateInfo.AppendColumn(col)

		rs.StateInfo.Connect("row-activated",
			func(tree *gtk.TreeView, path *gtk.TreePath, column *gtk.TreeViewColumn) {
				iter, err := rs.infoModel.GetIter(path)
				if err != nil {
					log.Printf("Error selecting state: %s", err)
					return
				}

				model := rs.infoModel.ToTreeModel()
				id, err := gtk_utils.ModelGetValue[int](model, iter, STATE_ID)
				if err != nil {
					log.Printf("Error selecting state: %s", err)
					return
				}

				rs.currentID = id
				rs.loadState()
			})
	}
	return rs
}

func (rs *RouterState) Start(source, dest int) {
	rs.destID = dest

	s := NewStateFromTree(rs.rTree)
	s.selected[source] = true
	s.currentRouter = source

	rs.state[rs.currentID] = s

	model := rs.rTree.Model.ToTreeModel()
	sourceName, err := gtk_utils.ModelGetValue[string](model, rs.rTree.RouterIter[source], ROUTER_NAME)
	if err != nil {
		log.Print(err)
		sourceName = strconv.Itoa(source)
	}

	destName, err := gtk_utils.ModelGetValue[string](model, rs.rTree.RouterIter[dest], ROUTER_NAME)
	if err != nil {
		log.Print(err)
		destName = strconv.Itoa(dest)
	}

	iter := rs.infoModel.Append(nil)
	rs.stateIter[rs.currentID] = iter
	rs.infoModel.SetValue(iter, STATE_ID, rs.currentID)

	desc := fmt.Sprintf("Send Message Router %s to %s", sourceName, destName)
	rs.infoModel.SetValue(iter, STATE_DESC, desc)

	if logState {
		log.Printf("Sending packet from %d to %d", source, dest)
	}

	rs.loadState()
}

func (rs *RouterState) PrevState() {
	currentIter := rs.stateIter[rs.currentID]
	if currentIter == nil {
		return
	}

	var prevIter gtk.TreeIter
	rs.infoModel.IterParent(&prevIter, currentIter)

	prevID, err := gtk_utils.ModelGetValue[int](rs.infoModel.ToTreeModel(), &prevIter, STATE_ID)
	if err != nil {
		log.Printf("Error getting prev state: %s", err)
		return
	}

	rs.currentID = prevID
	rs.loadState()
}

func (rs *RouterState) DetectAdjacent(routerID int) {
	s := rs.state[rs.currentID]
	if s == nil {
		return
	}

	s.DetectAdjacent(rs.pTree, routerID)
}

func (rs *RouterState) RoutePacket() error {
	s := rs.state[rs.currentID]
	if s == nil {
		return fmt.Errorf("Current State %d does not exist", rs.currentID)
	}

	r := s.routers[s.currentRouter]
	if r == nil {
		return fmt.Errorf("Router %d does not exist in current state", s.currentRouter)
	}

	nextHop, err := r.RoutePacket(rs.destID)
	if err != nil {
		return err
	}

	if logState {
		log.Printf("Sending packet from %d to %d", s.currentRouter, nextHop)
	}

	s.selected[nextHop] = true
	s.currentRouter = nextHop
	rs.loadState()

	return nil
}

func (rs *RouterState) NewState(desc string) {
	s := rs.state[rs.currentID]
	iter := rs.stateIter[rs.currentID]

	rs.currentID = rs.nextID
	rs.nextID++

	rs.state[rs.currentID] = NewStateFromState(s)

	newIter := rs.infoModel.Append(iter)
	rs.stateIter[rs.currentID] = newIter

	path, _ := rs.infoModel.GetPath(iter)
	rs.StateInfo.ExpandRow(path, false)

	rs.infoModel.SetValue(rs.stateIter[rs.currentID], STATE_ID, rs.currentID)
	rs.infoModel.SetValue(rs.stateIter[rs.currentID], STATE_DESC, desc)
}

func (rs *RouterState) loadState() {
	s := rs.state[rs.currentID]
	if s == nil {
		log.Printf("State %d does not exist", rs.currentID)
		return
	}

	for id, r := range rs.rTree.Routers {
		if r == nil {
			continue
		}

		r.Router = s.routers[id]
		r.Selected = s.selected[id]
	}
}

func (rs *RouterState) BroadcastRouter(routerID int) {
	s := rs.state[rs.currentID]
	s.Broadcast(routerID)
}

func (rs *RouterState) IsPrevState() bool {
	return rs.currentID > 0
}

func (rs *RouterState) IsNextState() bool {
	s := rs.state[rs.currentID]
	return s.currentRouter != rs.destID
}

func (rs *RouterState) UpdateRouterInfo() {
	s := rs.state[rs.currentID]

	for r1, r := range s.routers {
		model, _ := gtk.ListStoreNew(
			glib.TYPE_STRING,
			glib.TYPE_STRING,
			glib.TYPE_STRING,
			glib.TYPE_STRING,
			glib.TYPE_INT,
		)

		rs.RouterInfo[r1], _ = gtk.TreeModelSortNew(model)
		rs.RouterInfo[r1].SetSortColumnId(INFO_DEST_NAME, gtk.SORT_ASCENDING)

		info := r.Info()

		for _, p := range info {
			err := rs.addInfo(model, p)
			if err != nil {
				log.Printf("Error adding info for router %d: %s", r1, err)
				continue
			}
		}
	}
}

func (rs *RouterState) addInfo(routerModel *gtk.ListStore, p Path) (err error) {
	model := rs.rTree.Model.ToTreeModel()
	iter := rs.rTree.RouterIter[p.DestID]

	destName, err := gtk_utils.ModelGetValue[string](model, iter, ROUTER_NAME)
	if err != nil {
		err = fmt.Errorf("Col %d: %s", ROUTER_NAME, err)
		return
	}

	destIP, err := gtk_utils.ModelGetValue[string](model, iter, ROUTER_IP)
	if err != nil {
		err = fmt.Errorf("Col %d: %s", ROUTER_IP, err)
		return
	}

	iter = rs.rTree.RouterIter[p.NextHopID]
	nextName := "-"
	nextIP := "-"
	if iter != nil {
		nextName, err = gtk_utils.ModelGetValue[string](model, iter, ROUTER_NAME)
		if err != nil {
			nextName = "-"
		}

		nextIP, err = gtk_utils.ModelGetValue[string](model, iter, ROUTER_IP)
		if err != nil {
			nextIP = "-"
		}

		err = nil
	}

	row := routerModel.Append()
	routerModel.SetValue(row, INFO_DEST_NAME, destName)
	routerModel.SetValue(row, INFO_DEST_IP, destIP)
	routerModel.SetValue(row, INFO_NEXT_NAME, nextName)
	routerModel.SetValue(row, INFO_NEXT_IP, nextIP)
	routerModel.SetValue(row, INFO_DIST, p.Dist)

	return
}
