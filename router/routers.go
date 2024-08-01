package router

import (
	"fmt"
	"log"
	"routing-gui/gtk_utils"

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
	Routers     map[int]*RouterIcon
	RouterIter  map[int]*gtk.TreeIter
	MaxRouterID int
	Model       *gtk.ListStore

	ActiveRouterID   int
	getRouterInfo    func(int) *gtk.TreeModel
	routerInfoList   *gtk.TreeView
	routerInfoHeader *gtk.HeaderBar
}

func NewRouterTree(header *gtk.HeaderBar, tree *gtk.TreeView, getRouterInfo func(int) *gtk.TreeModel) *RouterTree {
	rTree := &RouterTree{
		MaxRouterID:      1,
		routerInfoHeader: header,
		routerInfoList:   tree,
	}
	rTree.Routers = make(map[int]*RouterIcon, 100)
	rTree.RouterIter = make(map[int]*gtk.TreeIter, 100)

	rTree.Model, _ = gtk.ListStoreNew(glib.TYPE_INT, glib.TYPE_STRING, glib.TYPE_STRING)

	cols := []string{"Destination Name", "Destination IP", "Next Hop Name", "Next Hop IP", "Distance"}
	cols_index := []int{INFO_DEST_NAME, INFO_DEST_IP, INFO_NEXT_NAME, INFO_NEXT_IP, INFO_DIST}
	stateCol, _ := gtk.CellRendererTextNew()

	for i := range cols {
		col, _ := gtk.TreeViewColumnNewWithAttribute(cols[i], stateCol, "text", cols_index[i])
		rTree.routerInfoList.AppendColumn(col)
	}

	return rTree
}

func (rTree *RouterTree) SetupTreeColumns(routerList *gtk.TreeView) {
	routerList.SetModel(rTree.Model)
	routerList.SetActivateOnSingleClick(true)

	name, _ := gtk.CellRendererTextNew()
	name.SetProperty("editable", true)
	name.Connect("edited", rTree.UpdateName)

	col, _ := gtk.TreeViewColumnNewWithAttribute("Name", name, "text", ROUTER_NAME)
	routerList.AppendColumn(col)

	ip, _ := gtk.CellRendererTextNew()
	ip.SetProperty("editable", true)
	ip.Connect("edited", rTree.UpdateIP)

	col, _ = gtk.TreeViewColumnNewWithAttribute("IP Address", ip, "text", ROUTER_IP)
	routerList.AppendColumn(col)
	routerList.Connect("row-activated", rTree.SelectRouter)
}

func (rTree *RouterTree) UpdateName(cell *gtk.CellRendererText, path, text string) {
	iter, err := rTree.Model.GetIterFromString(path)
	if err != nil {
		log.Printf("Error editing name: %s", err)
		return
	}

	id, err := gtk_utils.ModelGetValue[int](rTree.Model.ToTreeModel(), iter, ROUTER_ID)
	if err != nil {
		log.Printf("Error getting id: %s", err)
		return
	}

	r := rTree.Routers[id]
	if r != nil {
		r.Name = text
	}

	rTree.Model.SetValue(iter, ROUTER_NAME, text)
}

func (rTree *RouterTree) UpdateIP(cell *gtk.CellRendererText, path, text string) {
	iter, err := rTree.Model.GetIterFromString(path)
	if err != nil {
		log.Printf("Error editing name: %s", err)
		return
	}

	id, err := gtk_utils.ModelGetValue[int](rTree.Model.ToTreeModel(), iter, ROUTER_ID)
	if err != nil {
		log.Printf("Error getting id: %s", err)
		return
	}

	r := rTree.Routers[id]
	if r != nil {
		r.IP = text
	}

	rTree.Model.SetValue(iter, ROUTER_IP, text)
}

func (rTree *RouterTree) SelectRouter(tree *gtk.TreeView, path *gtk.TreePath, column *gtk.TreeViewColumn) {
	iter, err := rTree.Model.GetIter(path)
	if err != nil {
		log.Print(err)
		return
	}

	model := rTree.Model.ToTreeModel()
	routerID, err := gtk_utils.ModelGetValue[int](model, iter, ROUTER_ID)
	if err != nil {
		log.Print(err)
		return
	}

	name, err := gtk_utils.ModelGetValue[string](model, iter, ROUTER_NAME)
	if err != nil {
		log.Print(err)
		return
	}

	routerList := rTree.getRouterInfo(routerID)
	if routerList == nil {
		return
	}

	rTree.SetRouterState(routerID, name, routerList)
}

func (rTree *RouterTree) SetRouterState(routerID int, name string, model *gtk.TreeModel) {
	rTree.ActiveRouterID = routerID
	rTree.routerInfoHeader.SetTitle(fmt.Sprintf("Router %s State", name))
	rTree.routerInfoList.SetModel(model)
}

func (rTree *RouterTree) AddRouter(r *RouterIcon) {
	r.RouterID = rTree.MaxRouterID

	iter := rTree.Model.Append()
	rTree.Model.SetValue(iter, ROUTER_ID, r.RouterID)
	rTree.Model.SetValue(iter, ROUTER_NAME, r.Name)
	rTree.Model.SetValue(iter, ROUTER_IP, r.IP)

	rTree.Routers[r.RouterID] = r
	rTree.RouterIter[r.RouterID] = iter
	rTree.MaxRouterID++
}

func (rTree *RouterTree) GetValue(routerID, col int) (string, error) {
	iter := rTree.RouterIter[routerID]
	model := rTree.Model.ToTreeModel()

	return gtk_utils.ModelGetValue[string](model, iter, col)
}

func (rTree *RouterTree) GetRouter(iter *gtk.TreeIter) (r *RouterIcon, err error) {
	model := rTree.Model.ToTreeModel()

	routerID, err := gtk_utils.ModelGetValue[int](model, iter, ROUTER_ID)
	if err != nil {
		return
	}

	r, ok := rTree.Routers[routerID]
	if !ok {
		err = fmt.Errorf("Router %d does not exist", routerID)
		return
	}

	return
}

func (rTree *RouterTree) GetRouterIcon(name string) *RouterIcon {
	for _, r := range rTree.Routers {
		if r.Name == name {
			return r
		}
	}

	return nil
}

func (rTree *RouterTree) Draw(cr *cairo.Context) {
	for _, r := range rTree.Routers {
		if r == nil {
			continue
		}

		if r.Router == nil {
			continue
		}

		r.Draw(cr)
	}
}
