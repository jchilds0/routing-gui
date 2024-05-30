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
	RouterInfo  *gtk.Box
	routerList  *gtk.TreeView
	routerLabel *gtk.HeaderBar
}

func NewRouterTree() *RouterTree {
	rTree := &RouterTree{MaxRouterID: 1}
	rTree.Routers = make(map[int]*RouterIcon, 100)
	rTree.RouterIter = make(map[int]*gtk.TreeIter, 100)

	rTree.Model, _ = gtk.ListStoreNew(glib.TYPE_INT, glib.TYPE_STRING, glib.TYPE_STRING)

	rTree.RouterInfo, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)

	rTree.routerLabel, _ = gtk.HeaderBarNew()
	rTree.RouterInfo.PackStart(rTree.routerLabel, false, false, 0)

	rTree.routerList, _ = gtk.TreeViewNew()
	rTree.RouterInfo.PackStart(rTree.routerList, false, false, 0)

	cols := []string{"Destination Name", "Destination IP", "Next Hop Name", "Next Hop IP", "Distance"}
	cols_index := []int{INFO_DEST_NAME, INFO_DEST_IP, INFO_NEXT_NAME, INFO_NEXT_IP, INFO_DIST}
	stateCol, _ := gtk.CellRendererTextNew()

	for i := range cols {
		col, _ := gtk.TreeViewColumnNewWithAttribute(cols[i], stateCol, "text", cols_index[i])
		rTree.routerList.AppendColumn(col)
	}

	return rTree
}

func (rTree *RouterTree) AddColumns(tree *gtk.TreeView, getRouterList func(int) gtk.ITreeModel) {
	name, _ := gtk.CellRendererTextNew()
	name.SetProperty("editable", true)
	name.Connect("edited",
		func(cell *gtk.CellRendererText, path, text string) {
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
		})

	col, _ := gtk.TreeViewColumnNewWithAttribute("Name", name, "text", ROUTER_NAME)
	tree.AppendColumn(col)

	ip, _ := gtk.CellRendererTextNew()
	ip.SetProperty("editable", true)
	ip.Connect("edited",
		func(cell *gtk.CellRendererText, path, text string) {
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
		})
	col, _ = gtk.TreeViewColumnNewWithAttribute("IP Address", ip, "text", ROUTER_IP)
	tree.AppendColumn(col)

	tree.Connect("row-activated",
		func(tree *gtk.TreeView, path *gtk.TreePath, column *gtk.TreeViewColumn) {
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

			routerList := getRouterList(routerID)
			rTree.routerList.SetModel(routerList)
			rTree.routerLabel.SetTitle(fmt.Sprintf("Router %s State", name))
		})
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
