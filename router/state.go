package router

import (
	"log"
	"slices"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const logState = true

type Router interface {
	RoutePacket(int) (int, error)
	Broadcast() map[int]int
	Recieve(int, map[int]int)
	Info() map[int]int
	AddRouter(int, int)
	RemoveRouter(int)
	Copy() Router
}

type State struct {
	currentRouter int
	selected      map[int]bool
	routers       map[int]Router
}

func NewState(routerID int) *State {
	s := &State{currentRouter: routerID}

	s.selected = make(map[int]bool, 30)
	s.routers = make(map[int]Router, 30)

	return s
}

func (s *State) StoreRouterState(rTree *RouterTree) {
	for id, r := range rTree.Routers {
		s.routers[id] = r.Router.Copy()
	}
}

func (s *State) DetectAdjacent(pTree *PipeTree) {
	for id1, router1 := range s.routers {
		for i := range pTree.Router1 {
			if pTree.Router1[i] == id1 {
				router1.AddRouter(pTree.Router2[i], pTree.Weight[i])
			}

			if pTree.Router2[i] == id1 {
				router1.AddRouter(pTree.Router1[i], pTree.Weight[i])
			}
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
	INFO_NAME = iota
	INFO_IP
	INFO_DIST
)

type RouterState struct {
	state      []*State
	current    int
	destID     int
	RouterInfo map[int]*gtk.ListStore
}

func NewRouterState() *RouterState {
	rs := &RouterState{}

	rs.state = make([]*State, 0, 30)
	rs.RouterInfo = make(map[int]*gtk.ListStore, 30)
	return rs
}

func (rs *RouterState) Start(source, dest int, rTree *RouterTree) {
	rs.destID = dest

	s := NewState(source)
	s.selected[source] = true
	s.StoreRouterState(rTree)
	rs.state = append(rs.state, s)

	if logState {
		log.Printf("Sending packet from %d to %d", source, dest)
	}

	rs.loadState(rTree)
}

func (rs *RouterState) PrevState(rTree *RouterTree) {
	if rs.current == 0 {
		return
	}

	rs.state = slices.Delete(rs.state, rs.current, rs.current+1)
	rs.current--

	rs.loadState(rTree)
}

func (rs *RouterState) RoutePacket(pTree *PipeTree) {
	s := rs.state[rs.current]
	if s == nil {
		return
	}

	r := pTree.Routers.Routers[s.currentRouter]
	s.DetectAdjacent(pTree)

	nextHop, err := r.Router.RoutePacket(rs.destID)
	if err != nil {
		log.Print(err)
		return
	}

	if logState {
		log.Printf("Sending packet from %d to %d", r.id, nextHop)
	}

	nextState := NewState(nextHop)

	for i := range s.selected {
		nextState.selected[i] = true
	}

	nextState.selected[nextHop] = true
	nextState.StoreRouterState(pTree.Routers)
	rs.state = append(rs.state, nextState)
	rs.current++

	rs.loadState(pTree.Routers)
}

func (rs *RouterState) loadState(rTree *RouterTree) {
	s := rs.state[rs.current]

	for id, r := range rTree.Routers {
		if r == nil {
			continue
		}

		r.Router = s.routers[id]
		r.Selected = s.selected[id]
	}
}

func (rs *RouterState) Broadcast() {
	s := rs.state[rs.current]

	for id := range s.routers {
		s.Broadcast(id)
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
		if rs.RouterInfo[r1] == nil {
			rs.RouterInfo[r1], _ = gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_INT)
		}

		rs.RouterInfo[r1].Clear()
		info := r.Info()

		for r2, dist := range info {
			err := rs.addInfo(r1, r2, dist, rTree)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}

func (rs *RouterState) addInfo(r1, r2, dist int, rTree *RouterTree) (err error) {
	model := rTree.Model.ToTreeModel()
	routerModel := rs.RouterInfo[r1]
	iter := rTree.RouterIter[r2]

	adjName, err := ModelGetValue[string](model, iter, ROUTER_NAME)
	if err != nil {
		return
	}

	adjIP, err := ModelGetValue[string](model, iter, ROUTER_IP)
	if err != nil {
		return
	}

	row := routerModel.Append()
	routerModel.SetValue(row, INFO_NAME, adjName)
	routerModel.SetValue(row, INFO_IP, adjIP)
	routerModel.SetValue(row, INFO_DIST, dist)

	return
}
