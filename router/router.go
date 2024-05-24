package router

import "github.com/gotk3/gotk3/cairo"

type Router struct {
	X float64
	Y float64
	W float64
	H float64
}

func NewRouter() *Router {
	r := &Router{X: 10, Y: 10, W: 100, H: 100}

	return r
}

func (r *Router) Draw(cr *cairo.Context) {
	cr.Rectangle(r.X, r.Y, r.W, r.H)
	cr.SetSourceRGB(0.5, 0.5, 1)
	cr.Fill()
}

func (r *Router) Contains(x float64, y float64) bool {
	return (r.X < x && x < r.X+r.W) && (r.Y < y && y < r.Y+r.H)
}

func (r *Router) UpdatePos(delX float64, delY float64) {
	r.X += delX
	r.Y += delY
}
