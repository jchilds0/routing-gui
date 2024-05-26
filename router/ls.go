package router

type LinkStateRouter struct {
	id       int
	dist     map[int]map[int]int
	adjacent map[int]int
}

func NewLinkStateRouter(routerID int) *LinkStateRouter {
	ls := &LinkStateRouter{id: routerID}

	ls.dist = make(map[int]map[int]int)
	ls.adjacent = make(map[int]int)
	return ls
}

func (ls *LinkStateRouter) RoutePacket(dest int) (nextHop int, err error) {
	for id := range ls.adjacent {
		if id == ls.id {
			continue
		}

		nextHop = id
	}

	return
}

func (ls *LinkStateRouter) Broadcast() map[int]int {
	return ls.adjacent
}

func (ls *LinkStateRouter) Recieve(routerID int, distList map[int]int) {
	ls.dist[routerID] = make(map[int]int)

	for adjID, w := range distList {
		ls.dist[routerID][adjID] = w
	}
}

func (ls *LinkStateRouter) Info() map[int]int {
	return ls.adjacent
}

func (ls *LinkStateRouter) AddRouter(routerID, dist int) {
	ls.adjacent[routerID] = dist
}

func (ls *LinkStateRouter) RemoveRouter(routerID int) {
	delete(ls.adjacent, routerID)
	delete(ls.dist, routerID)
}

func (ls *LinkStateRouter) Copy() Router {
	return ls
}
