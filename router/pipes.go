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
	Router1 []int
	Router2 []int
	Weight  []int
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
	pTree.Weight = make([]int, 0, 100)

	pTree.Box, _ = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)

	{
		// Buttons
		buttonBox, _ := gtk.FlowBoxNew()
		buttonBox.SetSelectionMode(gtk.SELECTION_NONE)
		buttonBox.SetColumnSpacing(10)
		buttonBox.SetRowSpacing(10)

		pTree.Box.PackStart(buttonBox, false, false, 15)

		/* Router 1 */
		box1, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
		buttonBox.Add(box1)

		label1, _ := gtk.LabelNewWithMnemonic("Router 1:")
		label1.SetWidthChars(10)
		box1.PackStart(label1, false, false, 0)

		cell, _ := gtk.CellRendererTextNew()

		router1, _ := gtk.ComboBoxNewWithModel(pTree.Routers.Model)
		router1.SetActive(ROUTER_NAME)
		router1.CellLayout.PackStart(cell, true)
		router1.CellLayout.AddAttribute(cell, "text", ROUTER_NAME)
		box1.PackStart(router1, false, false, 0)

		/* Router 2 */
		box2, _ := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
		buttonBox.Add(box2)

		label2, _ := gtk.LabelNewWithMnemonic("Router 2: ")
		label2.SetWidthChars(10)
		box2.PackStart(label2, false, false, 0)

		router2, _ := gtk.ComboBoxNewWithModel(pTree.Routers.Model)
		router2.SetActive(ROUTER_NAME)
		router2.CellLayout.PackStart(cell, true)
		router2.CellLayout.AddAttribute(cell, "text", ROUTER_NAME)
		box2.PackStart(router2, false, false, 0)

		addPipe, _ := gtk.ButtonNewWithLabel("Add Connection")
		buttonBox.Add(addPipe)

		addPipe.Connect("clicked", func() {
			model := pTree.Routers.Model.ToTreeModel()

			iter1, err := router1.GetActiveIter()
			if err != nil {
				log.Printf("Error getting router 1: %s", err)
				return
			}

			router1ID, err := gtk_utils.ModelGetValue[int](model, iter1, ROUTER_ID)
			if err != nil {
				return
			}

			iter2, err := router2.GetActiveIter()
			if err != nil {
				log.Printf("Error getting router 2: %s", err)
				return
			}

			router2ID, err := gtk_utils.ModelGetValue[int](model, iter2, ROUTER_ID)
			if err != nil {
				return
			}

			err = pTree.AddConnection(router1ID, router2ID, 1)
			if err != nil {
				log.Printf("Error adding connection: %s", err)
				return
			}
		})

		removePipe, _ := gtk.ButtonNewWithLabel("Remove Connection")
		buttonBox.Add(removePipe)
	}

	{
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
		pTree.List.AppendColumn(col)
	}

	return pTree
}

func (pTree *PipeTree) AddConnection(r1, r2, w int) (err error) {
	iter := pTree.Model.Append()

	pTree.Model.SetValue(iter, ROUTER1_ID, r1)
	pTree.Router1 = append(pTree.Router1, r1)

	pTree.Model.SetValue(iter, ROUTER2_ID, r2)
	pTree.Router2 = append(pTree.Router2, r2)

	pTree.Weight = append(pTree.Weight, w)
	pTree.Model.SetValue(iter, WEIGHT, w)

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

		radius := float64(75)

		startX := x1 + radius*unitX
		startY := y1 + radius*unitY
		endX := x2 - radius*unitX
		endY := y2 - radius*unitY
		lineWidth := float64(8)

		cr.SetSourceRGB(0, 0, 0)
		cr.MoveTo(startX, startY)
		cr.LineTo(endX, endY)
		cr.SetLineWidth(lineWidth)
		cr.Stroke()

		cr.Arc(startX, startY, lineWidth/2, 0, 2*math.Pi)
		cr.Arc(endX, endY, lineWidth/2, 0, 2*math.Pi)
		cr.Fill()

		cr.SetSourceRGB(0.5, 0.5, 0.5)
		cr.MoveTo(startX, startY)
		cr.LineTo(endX, endY)
		cr.SetLineWidth(lineWidth / 4)
		cr.Stroke()
		cr.Fill()

		cr.SetSourceRGB(0, 0, 0)
		cr.SelectFontFace("Georgia", cairo.FONT_SLANT_NORMAL, cairo.FONT_WEIGHT_NORMAL)
		cr.SetFontSize(16)

		centerX := (endX + startX) / 2
		centerY := (endY + startY) / 2

		normalX, normalY := normalVector(unitX, unitY)
		dist := float64(20)

		w := strconv.Itoa(pTree.Weight[conn])
		cr.MoveTo(centerX+normalX*dist, centerY+normalY*dist)
		cr.ShowText(w)
	}
}
