package router

import "github.com/gotk3/gotk3/cairo"

type Router struct {
	id   int
	Name string
	IP   string
	X    float64
	Y    float64
	W    float64
	H    float64
}

func NewRouter() *Router {
	r := &Router{X: 10, Y: 10, W: 100, H: 100}

	return r
}

func (r *Router) Draw(cr *cairo.Context) {
	cr.Rectangle(r.X, r.Y, r.W, r.H)
	cr.SetSourceRGB(0.5, 0.5, 1)
	cr.Fill()

	centerX := r.X + r.W/2

	cr.SetSourceRGB(0, 0, 0)
	cr.SelectFontFace("Georgia", cairo.FONT_SLANT_NORMAL, cairo.FONT_WEIGHT_NORMAL)
	cr.SetFontSize(16)

	te := cr.TextExtents(r.Name)
	cr.MoveTo(centerX-te.Width/2, r.Y+r.H+20)
	cr.ShowText(r.Name)

	te = cr.TextExtents(r.IP)
	cr.MoveTo(centerX-te.Width/2, r.Y+r.H+40)
	cr.ShowText(r.IP)
}

func (r *Router) Contains(x float64, y float64) bool {
	return (r.X < x && x < r.X+r.W) && (r.Y < y && y < r.Y+r.H)
}

func (r *Router) UpdatePos(delX float64, delY float64) {
	r.X += delX
	r.Y += delY
}
