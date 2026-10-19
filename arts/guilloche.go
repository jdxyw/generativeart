package arts

import (
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type guilloche struct {
	curveNum  int
	majorR    float64
	minorR    float64
	penOffset float64
}

// NewGuilloche returns a guilloche object that renders precision hypotrochoid
// vector rosette curves using tdewolff/canvas stroke offsetting and Oklab gradients.
func NewGuilloche(curveNum int, majorR, minorR, penOffset float64) *guilloche {
	if curveNum <= 0 {
		curveNum = 28
	}
	if majorR <= 0 {
		majorR = 175.0
	}
	if minorR <= 0 {
		minorR = 65.0
	}
	if penOffset <= 0 {
		penOffset = 115.0
	}
	return &guilloche{
		curveNum:  curveNum,
		majorR:    majorR,
		minorR:    minorR,
		penOffset: penOffset,
	}
}

// Generative draws precision guilloche hypotrochoid curves and soft concentric rings.
func (g *guilloche) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	schema := c.Opts().ColorSchema()
	if len(schema) == 0 {
		schema = common.Oceanic
	}
	grad := render.NewMultiStopGradient(schema...)
	accentGrad := render.NewMultiStopGradient(common.Sunset...)

	cx, cy := w*0.5, h*0.5

	// Soft radial glow backdrop at center.
	rc.SetBlendMode(render.BlendAdd)
	rc.DrawGlowDot(cx, cy, math.Min(w, h)*0.45, grad.SampleOklab(0.3), 0.22, 0.0)

	// Guilloche hypotrochoid curves stroked via tdewolff/canvas exact stroke offsetting.
	rc.SetBlendMode(render.BlendScreen)
	denom := float64(g.curveNum - 1)
	if denom < 1 {
		denom = 1
	}
	for i := 0; i < g.curveNum; i++ {
		u := float64(i) / denom
		p := &canvas.Path{}
		R := (g.majorR + u*55.0) * scale
		r := g.minorR * scale
		d := (g.penOffset - u*45.0) * scale
		rot := u * 0.45

		steps := 360
		for s := 0; s <= steps; s++ {
			t := (float64(s) / float64(steps)) * 6.0 * math.Pi
			x0 := (R-r)*math.Cos(t) + d*math.Cos(((R-r)/r)*t)
			y0 := (R-r)*math.Sin(t) - d*math.Sin(((R-r)/r)*t)
			xr := x0*math.Cos(rot) - y0*math.Sin(rot)
			yr := x0*math.Sin(rot) + y0*math.Cos(rot)
			px, py := cx+xr*0.92, cy+yr*0.92
			if s == 0 {
				p.MoveTo(px, py)
			} else {
				p.LineTo(px, py)
			}
		}
		col := grad.SampleOklab(u)
		if i%4 == 0 {
			col = accentGrad.SampleOklab(u)
		}
		rc.StrokePath(p, 1.15*scale, canvas.RoundCap, canvas.RoundJoin, col, 0.42)
	}

	// Concentric soft rings in additive mode.
	rc.SetBlendMode(render.BlendAdd)
	for i := 1; i <= 5; i++ {
		rad := float64(i) * 48.0 * scale
		rc.DrawSoftRing(cx, cy, rad, 2.0*scale, accentGrad.SampleOklab(float64(i)/5.0), 0.38)
	}

	rc.ApplyBloom(0.5, 9.0, 0.5)
	rc.ApplyVignette(0.3)
	c.MergeRenderCanvasToneMapped(rc, 1.1, 0)
}
