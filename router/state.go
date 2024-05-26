package router

import (
	"log"
	"maps"
	"slices"
)

const logState = true

type RouterState struct {
	selected     []map[int]bool
	routers      []map[int]Router
	currentID    []int
	CurrentState int
	destID       int
}

func NewRouterState() *RouterState {
	rs := &RouterState{}

	rs.selected = make([]map[int]bool, 0, 30)
	rs.currentID = make([]int, 0, 30)
	rs.routers = make([]map[int]Router, 0, 30)

	return rs
}

func (rs *RouterState) Start(source, dest int, rTree *RouterTree) {
	rs.CurrentState = 0
	rs.destID = dest
	rs.currentID = append(rs.currentID, source)

	selection := make(map[int]bool)
	selection[source] = true
	rs.selected = append(rs.selected, selection)

	if logState {
		log.Printf("Sending packet from %d to %d", source, dest)
	}

	rs.StoreRouterState(rTree)
}

func (rs *RouterState) StoreRouterState(rTree *RouterTree) {
	routers := make(map[int]Router)

	for id, r := range rTree.Routers {
		routers[id] = r.Router.Copy()
	}

	rs.routers = append(rs.routers, routers)
}

func (rs *RouterState) PrevState() {
	if rs.CurrentState == 0 {
		return
	}

	rs.selected = slices.Delete(rs.selected, rs.CurrentState, rs.CurrentState+1)
	rs.routers = slices.Delete(rs.routers, rs.CurrentState, rs.CurrentState+1)
	rs.currentID = slices.Delete(rs.currentID, rs.CurrentState, rs.CurrentState+1)

	rs.CurrentState--
	return
}

func (rs *RouterState) NextState(pTree *PipeTree) {
	r := pTree.Routers.Routers[rs.currentID[rs.CurrentState]]

	nextHop, err := r.Router.RoutePacket(rs.destID)
	if err != nil {
		log.Print(err)
		return
	}

	if logState {
		log.Printf("Sending packet from %d to %d", r.id, nextHop)
	}

	rs.currentID = append(rs.currentID, nextHop)

	current := rs.selected[rs.CurrentState]
	selected := maps.Clone(current)
	selected[nextHop] = true
	rs.selected = append(rs.selected, selected)

	routers := make(map[int]Router, len(pTree.Routers.Routers))

	for id, r := range pTree.Routers.Routers {
		routers[id] = r.Router.Copy()
	}

	rs.routers = append(rs.routers, routers)
	rs.CurrentState++
}

func (rs *RouterState) LoadState(rTree *RouterTree) {
	routers := rs.routers[rs.CurrentState]
	selection := rs.selected[rs.CurrentState]

	for id, r := range rTree.Routers {
		if r == nil {
			continue
		}

		r.Router = routers[id]
		r.Selected = selection[id]
	}
}

func (rs *RouterState) IsPrevState() bool {
	return rs.CurrentState > 0
}

func (rs *RouterState) IsNextState() bool {
	return rs.currentID[rs.CurrentState] != rs.destID
}

func (rs *RouterState) Broadcast(pTree *PipeTree) {
	routers := rs.routers[rs.CurrentState]

	for id1, router1 := range routers {
		for i := range pTree.Router1 {
			if pTree.Router1[i] == id1 {
				router1.AddRouter(pTree.Router2[i], pTree.Weight[i])
			}

			if pTree.Router2[i] == id1 {
				router1.AddRouter(pTree.Router1[i], pTree.Weight[i])
			}
		}

		msg := router1.Broadcast()

		for id2, router2 := range routers {
			if id1 == id2 {
				continue
			}

			router2.Recieve(id1, msg)
		}
	}
}
