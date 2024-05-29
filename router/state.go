package router

import (
	"fmt"
	"log"
	"routing-gui/gtk_utils"
	"slices"

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

type RouterState struct {
	state      []*State
	current    int
	destID     int
	RouterInfo map[int]*gtk.TreeModelSort
}

func NewRouterState() *RouterState {
	rs := &RouterState{}

	rs.state = make([]*State, 0, 30)
	rs.RouterInfo = make(map[int]*gtk.TreeModelSort, 30)
	return rs
}

func (rs *RouterState) Start(source, dest int, rTree *RouterTree) {
	rs.destID = dest

	s := NewStateFromTree(rTree)
	s.selected[source] = true
	s.currentRouter = source

	rs.state = append(rs.state, s)

	if logState {
		log.Printf("Sending packet from %d to %d", source, dest)
	}

	rs.loadState(0, rTree)
}

func (rs *RouterState) PrevState(rTree *RouterTree) {
	if rs.current == 0 {
		return
	}

	rs.state = slices.Delete(rs.state, rs.current, rs.current+1)
	rs.current--

	rs.loadState(rs.current, rTree)
}

func (rs *RouterState) DetectAdjacent(routerID int, pTree *PipeTree) {
	s := rs.state[rs.current]
	if s == nil {
		return
	}

	s.DetectAdjacent(pTree, routerID)
}

func (rs *RouterState) RoutePacket(pTree *PipeTree) error {
	s := rs.state[rs.current]
	if s == nil {
		return fmt.Errorf("Current State %d does not exist", rs.current)
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
	rs.loadState(rs.current, pTree.Routers)

	return nil
}

func (rs *RouterState) NewState() {
	s := rs.state[rs.current]
	state := NewStateFromState(s)

	rs.current = len(rs.state)
	rs.state = append(rs.state, state)
}

func (rs *RouterState) loadState(stateID int, rTree *RouterTree) {
	if stateID < 0 || stateID >= len(rs.state) {
		log.Printf("State %d out of range", stateID)
		return
	}

	s := rs.state[stateID]
	if s == nil {
		log.Printf("State %d does not exist", stateID)
		return
	}

	for id, r := range rTree.Routers {
		if r == nil {
			continue
		}

		r.Router = s.routers[id]
		r.Selected = s.selected[id]
	}
}

func (rs *RouterState) BroadcastRouter(routerID int) {
	s := rs.state[rs.current]
	s.Broadcast(routerID)
}

func (rs *RouterState) IsPrevState() bool {
	return rs.current > 0
}

func (rs *RouterState) IsNextState() bool {
	s := rs.state[rs.current]
	return s.currentRouter != rs.destID
}

func (rs *RouterState) UpdateRouterInfo(rTree *RouterTree) {
	s := rs.state[rs.current]

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
			err := rs.addInfo(model, p, rTree)
			if err != nil {
				log.Printf("Error adding info for router %d: %s", r1, err)
				continue
			}
		}
	}
}

func (rs *RouterState) addInfo(routerModel *gtk.ListStore, p Path, rTree *RouterTree) (err error) {
	model := rTree.Model.ToTreeModel()
	iter := rTree.RouterIter[p.DestID]

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

	iter = rTree.RouterIter[p.NextHopID]
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
