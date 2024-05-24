package router

import (
	"strconv"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const (
	ROUTER_ID = iota
	ROUTER_NAME
	ROUTER_IP
)

type RouterTree struct {
	Routers     map[int]*Router
	RouterIter  map[int]*gtk.TreeIter
	maxRouterID int
	Model       *gtk.ListStore
	List        *gtk.TreeView
}

func NewRouterTree() *RouterTree {
	rTree := &RouterTree{maxRouterID: 0}
	rTree.Routers = make(map[int]*Router, 100)
	rTree.RouterIter = make(map[int]*gtk.TreeIter, 100)

	rTree.Model, _ = gtk.ListStoreNew(glib.TYPE_INT, glib.TYPE_STRING, glib.TYPE_STRING)
	rTree.List, _ = gtk.TreeViewNew()
	rTree.List.SetModel(rTree.Model.ToTreeModel())

	cell, _ := gtk.CellRendererTextNew()

	col, _ := gtk.TreeViewColumnNewWithAttribute("Name", cell, "text", ROUTER_NAME)
	rTree.List.AppendColumn(col)

	col, _ = gtk.TreeViewColumnNewWithAttribute("IP Address", cell, "text", ROUTER_IP)
	rTree.List.AppendColumn(col)

	return rTree
}

func (rTree *RouterTree) AddRouter(r *Router) {
	r.id = rTree.maxRouterID
	r.Name = "Router " + strconv.Itoa(r.id)
	r.IP = "127.0.0.1"

	iter := rTree.Model.Append()
	rTree.Model.SetValue(iter, ROUTER_ID, r.id)
	rTree.Model.SetValue(iter, ROUTER_NAME, r.Name)
	rTree.Model.SetValue(iter, ROUTER_IP, r.IP)

	rTree.Routers[r.id] = r
	rTree.RouterIter[r.id] = iter
	rTree.maxRouterID++
}

func (rTree *RouterTree) GetValue(routerID, col int) (string, error) {
	iter := rTree.RouterIter[routerID]
	model := rTree.Model.ToTreeModel()

	return ModelGetValue[string](model, iter, col)
}

func (rTree *RouterTree) Draw(cr *cairo.Context) {
	for _, r := range rTree.Routers {
		if r == nil {
			continue
		}

		r.Draw(cr)
	}
}
