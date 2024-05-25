package router

type RouterState struct {
	selected     []map[int]bool
	CurrentState int
}

func NewRouterState() *RouterState {
	rs := &RouterState{}

	rs.selected = make([]map[int]bool, 10)

	rs.selected[0] = make(map[int]bool, 10)
	rs.selected[0][0] = true

	rs.selected[1] = make(map[int]bool, 10)
	rs.selected[1][1] = true

	rs.selected[2] = make(map[int]bool, 10)
	rs.selected[2][2] = true

	rs.selected[3] = make(map[int]bool, 10)
	rs.selected[3][3] = true

	return rs
}

func (rs *RouterState) UpdateState(index int, rTree *RouterTree) {
	if index < 0 || index >= len(rs.selected) {
		return
	}

	if rs.selected[index] == nil {
		return
	}

	for i, r := range rTree.Routers {
		if r == nil {
			continue
		}

		r.Selected = rs.selected[index][i]
	}

	rs.CurrentState = index
	return
}

func (rs *RouterState) IsPrevState() bool {
	return rs.CurrentState > 0
}

func (rs *RouterState) IsNextState() bool {
	return rs.CurrentState < len(rs.selected)-1
}
