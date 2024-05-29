package protocol

import (
	"maps"
	"routing-gui/router"
)

type LinkStateRouter struct {
	id          int
	graph       Graph
	graphUpdate bool
	adjacent    map[int]int
	nextHop     map[int]int
	dist        map[int]int
}

func NewLinkStateRouter(routerID int) *LinkStateRouter {
	ls := &LinkStateRouter{id: routerID}

	ls.graph = *NewGraph()
	ls.adjacent = make(map[int]int)

	return ls
}

func (ls *LinkStateRouter) RoutePacket(dest int) (nextHop int, err error) {
	nextHop = ls.nextHop[dest]

	return
}

func (ls *LinkStateRouter) updateRouting() {
	if ls.graphUpdate {
		ls.nextHop, ls.dist = ls.graph.Dijkstra(ls.id)
	}

	ls.graphUpdate = false
}

func (ls *LinkStateRouter) Broadcast() map[int]int {
	return ls.adjacent
}

func (ls *LinkStateRouter) Recieve(routerID int, distList map[int]int) {
	ls.graph.adjList[routerID] = make([]node, 0)

	for adjID, w := range distList {
		ls.graph.AddEdge(routerID, adjID, w)
		ls.graphUpdate = true
	}

	ls.updateRouting()
}

func (ls *LinkStateRouter) Info() (info []router.Path) {
	info = make([]router.Path, 0, len(ls.nextHop))

	for id := range ls.nextHop {
		p := router.Path{
			DestID:    id,
			Dist:      ls.dist[id],
			NextHopID: ls.nextHop[id],
		}

		info = append(info, p)
	}

	return info
}

func (ls *LinkStateRouter) AddRouter(routerID, dist int) {
	ls.adjacent[routerID] = dist
	ls.graph.AddEdge(ls.id, routerID, dist)
	ls.graphUpdate = true

	ls.updateRouting()
}

func (ls *LinkStateRouter) RemoveRouter(routerID int) {
	delete(ls.adjacent, routerID)
	ls.graph.RemoveNode(routerID)

	ls.updateRouting()
}

func (ls *LinkStateRouter) Copy() router.Router {
	newLS := NewLinkStateRouter(ls.id)

	newLS.adjacent = maps.Clone(ls.adjacent)
	newLS.nextHop = maps.Clone(ls.nextHop)
	newLS.dist = maps.Clone(ls.dist)

	for x, adj := range ls.graph.adjList {
		for _, node := range adj {
			newLS.graph.AddEdge(x, node.n, node.weight)
		}
	}

	return newLS
}
