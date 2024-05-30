package router

import (
	"log"
	"math"

	"github.com/gotk3/gotk3/cairo"
)

type RouterIcon struct {
	Router   Router
	RouterID int
	Name     string
	IP       string
	X        float64
	Y        float64
	W        float64
	H        float64
	Selected bool
}

func NewRouter(r Router) *RouterIcon {
	ri := &RouterIcon{Router: r, X: 10, Y: 10, W: 100, H: 100}

	return ri
}

func (r *RouterIcon) Draw(cr *cairo.Context) {
	if r.Selected {
		centerX := r.X + r.W/2
		centerY := r.Y + r.H/2

		cr.SetSourceRGBA(0, 1, 1, 0.5)
		cr.Arc(centerX, centerY, math.Max(r.W, r.H)/2, 0, 2*math.Pi)
		cr.Fill()
	}

	img, err := cairo.NewSurfaceFromPNG("./router/server-icon.png")
	if err != nil {
		log.Printf("Error loading server icon: %s", err)
		cr.SetSourceRGB(0.5, 0.5, 1)
		cr.Rectangle(r.X, r.Y, r.W, r.H)
		cr.Fill()
	} else {
		scale := float64(0.15)

		width := int(float64(img.GetWidth()) * scale)
		height := int(float64(img.GetHeight()) * scale)

		paddingX := (r.W - float64(width)) / 2
		paddingY := (r.H - float64(height)) / 2

		surface := img.CreateSimilar(cairo.CONTENT_COLOR_ALPHA, width, height)
		crIMG := cairo.Create(surface)

		crIMG.Scale(scale, scale)
		crIMG.SetSourceSurface(img, 0, 0)
		crIMG.SetOperator(cairo.OPERATOR_SOURCE)
		crIMG.Paint()

		cr.SetSourceSurface(surface, r.X+paddingX, r.Y+paddingY)
		cr.Paint()
	}

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

func (r *RouterIcon) Contains(x float64, y float64) bool {
	return (r.X < x && x < r.X+r.W) && (r.Y < y && y < r.Y+r.H)
}

func (r *RouterIcon) UpdatePos(delX float64, delY float64) {
	r.X += delX
	r.Y += delY
}
