package router

import (
	"log"
	"math"
	"routing-gui/gtk_utils"
	"strconv"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const (
	PIPE_ID = iota
	ROUTER1_ID
	ROUTER1_NAME
	ROUTER1_IP
	ROUTER2_ID
	ROUTER2_NAME
	ROUTER2_IP
	WEIGHT
)

type PipeTree struct {
	Router1    map[int]int
	Router2    map[int]int
	Weight     map[int]int
	PipeIter   map[int]*gtk.TreeIter
	nextPipeID int
	Routers    *RouterTree
	Model      *gtk.ListStore
}

func NewPipeTree(rs *RouterTree) *PipeTree {
	pTree := &PipeTree{
		Routers:    rs,
		nextPipeID: 0,
	}

	pTree.Router1 = make(map[int]int)
	pTree.Router2 = make(map[int]int)
	pTree.Weight = make(map[int]int)
	pTree.PipeIter = make(map[int]*gtk.TreeIter)

	// List View
	pTree.Model, _ = gtk.ListStoreNew(
		glib.TYPE_INT,
		glib.TYPE_INT,
		glib.TYPE_STRING,
		glib.TYPE_STRING,
		glib.TYPE_INT,
		glib.TYPE_STRING,
		glib.TYPE_STRING,
		glib.TYPE_STRING,
	)

	return pTree
}

func (pTree *PipeTree) AddColumns(tree *gtk.TreeView) {
	cell, _ := gtk.CellRendererTextNew()

	col, _ := gtk.TreeViewColumnNewWithAttribute("Router 1 Name", cell, "text", ROUTER1_NAME)
	tree.AppendColumn(col)

	col, _ = gtk.TreeViewColumnNewWithAttribute("Router 1 IP", cell, "text", ROUTER1_IP)
	tree.AppendColumn(col)

	col, _ = gtk.TreeViewColumnNewWithAttribute("Router 2 Name", cell, "text", ROUTER2_NAME)
	tree.AppendColumn(col)

	col, _ = gtk.TreeViewColumnNewWithAttribute("Router 2 IP", cell, "text", ROUTER2_IP)
	tree.AppendColumn(col)

	w, _ := gtk.CellRendererTextNew()
	w.SetProperty("editable", true)
	w.Connect("edited",
		func(cell *gtk.CellRendererText, path, text string) {
			iter, err := pTree.Model.GetIterFromString(path)
			if err != nil {
				log.Printf("Error editing name: %s", err)
				return
			}

			id, err := gtk_utils.ModelGetValue[int](pTree.Model.ToTreeModel(), iter, PIPE_ID)
			if err != nil {
				log.Printf("Error getting id: %s", err)
				return
			}

			pTree.Weight[id], err = strconv.Atoi(text)
			if err != nil {
				log.Print(err)
				return
			}

			pTree.Model.SetValue(iter, WEIGHT, text)
		})

	col, _ = gtk.TreeViewColumnNewWithAttribute("Weight", w, "text", WEIGHT)
	tree.AppendColumn(col)
}

func (pTree *PipeTree) AddConnection(r1, r2, w int) (err error) {
	iter := pTree.Model.Append()

	pTree.Model.SetValue(iter, PIPE_ID, pTree.nextPipeID)
	pTree.PipeIter[pTree.nextPipeID] = iter

	pTree.Model.SetValue(iter, ROUTER1_ID, r1)
	pTree.Router1[pTree.nextPipeID] = r1

	pTree.Model.SetValue(iter, ROUTER2_ID, r2)
	pTree.Router2[pTree.nextPipeID] = r2

	pTree.Model.SetValue(iter, WEIGHT, w)
	pTree.Weight[pTree.nextPipeID] = w

	pTree.nextPipeID++
	err = pTree.updateRow(iter)
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
		id, err = gtk_utils.ModelGetValue[int](model, iter, ids[i])
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

func unitVector(x, y float64) (float64, float64) {
	norm := math.Sqrt(math.Pow(x, 2) + math.Pow(y, 2))

	return x / norm, y / norm
}

func normalVector(x, y float64) (float64, float64) {
	if x == 0 {
		return y, 0
	}

	return -y, x
}

func (pTree *PipeTree) Draw(cr *cairo.Context) {
	for conn := range pTree.Router1 {
		r1Id := pTree.Router1[conn]
		r2Id := pTree.Router2[conn]

		r1 := pTree.Routers.Routers[r1Id]
		r2 := pTree.Routers.Routers[r2Id]

		x1 := r1.X + r1.W/2
		y1 := r1.Y + r1.H/2

		x2 := r2.X + r2.W/2
		y2 := r2.Y + r2.H/2

		unitX, unitY := unitVector(x2-x1, y2-y1)
		routerRadius := float64(75)
		weightRadius := float64(20)
		lineWidth := float64(8)

		centerX := (x1 + x2) / 2
		centerY := (y1 + y2) / 2

		l1StartX, l1StartY := x1+routerRadius*unitX, y1+routerRadius*unitY
		l1EndX, l1EndY := centerX-weightRadius*unitX, centerY-weightRadius*unitY
		l2StartX, l2StartY := centerX+weightRadius*unitX, centerY+weightRadius*unitY
		l2EndX, l2EndY := x2-routerRadius*unitX, y2-routerRadius*unitY

		// first line segment
		cr.SetSourceRGB(0, 0, 0)
		cr.MoveTo(l1StartX, l1StartY)
		cr.LineTo(l1EndX, l1EndY)
		cr.SetLineWidth(lineWidth)
		cr.Stroke()

		cr.Arc(l1StartX, l1StartY, lineWidth/2, 0, 2*math.Pi)
		cr.Arc(l1EndX, l1EndY, lineWidth/2, 0, 2*math.Pi)
		cr.Fill()

		cr.SetSourceRGB(0.5, 0.5, 0.5)
		cr.MoveTo(l1StartX, l1StartY)
		cr.LineTo(l1EndX, l1EndY)
		cr.SetLineWidth(lineWidth / 4)
		cr.Stroke()
		cr.Fill()

		// second line segment
		cr.SetSourceRGB(0, 0, 0)
		cr.MoveTo(l2StartX, l2StartY)
		cr.LineTo(l2EndX, l2EndY)
		cr.SetLineWidth(lineWidth)
		cr.Stroke()

		cr.Arc(l2StartX, l2StartY, lineWidth/2, 0, 2*math.Pi)
		cr.Arc(l2EndX, l2EndY, lineWidth/2, 0, 2*math.Pi)
		cr.Fill()

		cr.SetSourceRGB(0.5, 0.5, 0.5)
		cr.MoveTo(l2StartX, l2StartY)
		cr.LineTo(l2EndX, l2EndY)
		cr.SetLineWidth(lineWidth / 4)
		cr.Stroke()
		cr.Fill()

		// pipe weight
		cr.SetSourceRGB(0, 0, 0)
		cr.SelectFontFace("Georgia", cairo.FONT_SLANT_NORMAL, cairo.FONT_WEIGHT_NORMAL)
		cr.SetFontSize(20)

		w := strconv.Itoa(pTree.Weight[conn])
		cr.MoveTo(centerX-5, centerY+8)
		cr.ShowText(w)
	}
}

func (pTree *PipeTree) RemoveConnection(iter *gtk.TreeIter) (err error) {
	model := pTree.Model.ToTreeModel()
	pipeID, err := gtk_utils.ModelGetValue[int](model, iter, PIPE_ID)
	if err != nil {
		return
	}

	delete(pTree.Router1, pipeID)
	delete(pTree.Router2, pipeID)
	delete(pTree.Weight, pipeID)
	delete(pTree.PipeIter, pipeID)
	pTree.Model.Remove(iter)

	return
}
