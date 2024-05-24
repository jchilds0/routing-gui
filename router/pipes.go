package router

import (
	"log"
	"math"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const (
	ROUTER1_ID = iota
	ROUTER1_NAME
	ROUTER1_IP
	ROUTER2_ID
	ROUTER2_NAME
	ROUTER2_IP
	WEIGHT
)

type PipeTree struct {
	Router1 []int
	Router2 []int
	Routers *RouterTree
	Model   *gtk.ListStore
	List    *gtk.TreeView
	Box     *gtk.Box
}

func NewPipeTree(rs *RouterTree) *PipeTree {
	pTree := &PipeTree{
		Routers: rs,
	}

	pTree.Router1 = make([]int, 0, 100)
	pTree.Router2 = make([]int, 0, 100)

	pTree.Box, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)

	{
		// Buttons
		buttonBox, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
		pTree.Box.PackStart(buttonBox, false, false, 15)

		label1, _ := gtk.LabelNewWithMnemonic("Router 1: ")
		label1.SetWidthChars(10)
		buttonBox.PackStart(label1, false, false, 0)

		cell, _ := gtk.CellRendererTextNew()

		router1, _ := gtk.ComboBoxNewWithModel(pTree.Routers.Model)
		router1.SetActive(ROUTER_NAME)
		router1.CellLayout.PackStart(cell, true)
		router1.CellLayout.AddAttribute(cell, "text", ROUTER_NAME)

		buttonBox.PackStart(router1, false, false, 0)

		label2, _ := gtk.LabelNewWithMnemonic("Router 2: ")
		label2.SetWidthChars(10)
		buttonBox.PackStart(label2, false, false, 0)

		router2, _ := gtk.ComboBoxNewWithModel(pTree.Routers.Model.ToTreeModel())
		router2.SetActive(ROUTER_NAME)
		router2.CellLayout.PackStart(cell, true)
		router2.CellLayout.AddAttribute(cell, "text", ROUTER_NAME)

		buttonBox.PackStart(router2, false, false, 0)

		addPipe, _ := gtk.ButtonNewWithLabel("Add Connection")
		buttonBox.PackStart(addPipe, false, false, 15)

		addPipe.Connect("clicked", func() {
			iter1, err := router1.GetActiveIter()
			if err != nil {
				log.Printf("Error getting router 1: %s", err)
				return
			}

			iter2, err := router2.GetActiveIter()
			if err != nil {
				log.Printf("Error getting router 2: %s", err)
				return
			}

			iter, err := pTree.addConnection(iter1, iter2)
			if err != nil {
				log.Printf("Error adding connection: %s", err)
				return
			}

			err = pTree.updateRow(iter)
			if err != nil {
				log.Printf("Error updating row: %s", err)
				return
			}
		})

		removePipe, _ := gtk.ButtonNewWithLabel("Remove Connection")
		buttonBox.PackStart(removePipe, false, false, 15)
	}

	{
		// List View
		pTree.Model, _ = gtk.ListStoreNew(
			glib.TYPE_INT,
			glib.TYPE_STRING,
			glib.TYPE_STRING,
			glib.TYPE_INT,
			glib.TYPE_STRING,
			glib.TYPE_STRING,
			glib.TYPE_STRING,
		)

		pTree.List, _ = gtk.TreeViewNew()
		pTree.List.SetModel(pTree.Model.ToTreeModel())

		pTree.Box.PackStart(pTree.List, true, true, 0)

		cell, _ := gtk.CellRendererTextNew()

		col, _ := gtk.TreeViewColumnNewWithAttribute("Router 1 Name", cell, "text", ROUTER1_NAME)
		pTree.List.AppendColumn(col)

		col, _ = gtk.TreeViewColumnNewWithAttribute("Router 1 IP", cell, "text", ROUTER1_IP)
		pTree.List.AppendColumn(col)

		col, _ = gtk.TreeViewColumnNewWithAttribute("Router 2 Name", cell, "text", ROUTER2_NAME)
		pTree.List.AppendColumn(col)

		col, _ = gtk.TreeViewColumnNewWithAttribute("Router 2 IP", cell, "text", ROUTER2_IP)
		pTree.List.AppendColumn(col)

		col, _ = gtk.TreeViewColumnNewWithAttribute("Weight", cell, "text", WEIGHT)
		pTree.List.AppendColumn(col)
	}

	return pTree
}

func (pTree *PipeTree) addConnection(r1, r2 *gtk.TreeIter) (iter *gtk.TreeIter, err error) {
	model := pTree.Routers.Model.ToTreeModel()
	iter = pTree.Model.Append()

	routerID, err := ModelGetValue[int](model, r1, ROUTER_ID)
	if err != nil {
		return
	}

	pTree.Model.SetValue(iter, ROUTER1_ID, routerID)
	pTree.Router1 = append(pTree.Router1, routerID)

	routerID, err = ModelGetValue[int](model, r2, ROUTER_ID)
	if err != nil {
		return
	}

	pTree.Model.SetValue(iter, ROUTER2_ID, routerID)
	pTree.Router2 = append(pTree.Router2, routerID)

	return
}

func (pTree *PipeTree) updateRow(iter *gtk.TreeIter) (err error) {
	model := pTree.Model.ToTreeModel()
	ids := [4]int{ROUTER1_ID, ROUTER1_ID, ROUTER2_ID, ROUTER2_ID}
	valueCols := [4]int{ROUTER_NAME, ROUTER_IP, ROUTER_NAME, ROUTER_IP}
	destCols := [4]int{ROUTER1_NAME, ROUTER1_IP, ROUTER2_NAME, ROUTER2_IP}

	var s string
	var id int
	for i := range 4 {
		id, err = ModelGetValue[int](model, iter, ids[i])
		if err != nil {
			return
		}

		s, err = pTree.Routers.GetValue(id, valueCols[i])
		if err != nil {
			return
		}

		pTree.Model.SetValue(iter, destCols[i], s)
	}

	return
}

func (rTree *RouterTree) clipRouters(cr *cairo.Context) {
	cr.SetFillRule(cairo.FILL_RULE_EVEN_ODD)

	for _, r := range rTree.Routers {
		if r == nil {
			continue
		}

		x := r.X + r.W/2
		y := r.Y + r.H/2

		cr.Arc(x, y, 30, 0, 2*math.Pi)
		cr.Fill()
	}
}

func (pTree *PipeTree) Draw(cr *cairo.Context) {
	cr.SetSourceRGB(0, 1, 1)

	pTree.Routers.clipRouters(cr)
	cr.SetFillRule(cairo.FILL_RULE_WINDING)

	for conn := range pTree.Router1 {
		r1Id := pTree.Router1[conn]
		r2Id := pTree.Router2[conn]

		r1 := pTree.Routers.Routers[r1Id]
		r2 := pTree.Routers.Routers[r2Id]

		x1 := r1.X + r1.W/2
		y1 := r1.Y + r1.H/2

		x2 := r2.X + r2.W/2
		y2 := r2.Y + r2.H/2

		cr.MoveTo(x1, y1)
		cr.LineTo(x2, y2)
		cr.SetLineWidth(5)
		cr.Stroke()
	}

	cr.Fill()

}
