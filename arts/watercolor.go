package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type watercolor struct {
	shapeNum     int
	layers       int
	subdivisions int
	roughness    float64
}

// NewWatercolor returns a watercolor object that renders Tyler Hobbs-style
// wet-on-wet watercolor washes with Multiply pigment blending and cotton paper grain.
func NewWatercolor(shapeNum, layers, subdivisions int, roughness float64) *watercolor {
	if shapeNum <= 0 {
		shapeNum = 9
	}
	if layers <= 0 {
		layers = 22
	}
	if subdivisions <= 0 {
		subdivisions = 5
	}
	if roughness <= 0 {
		roughness = 0.32
	}
	return &watercolor{
		shapeNum:     shapeNum,
		layers:       layers,
		subdivisions: subdivisions,
		roughness:    roughness,
	}
}

// Generative draws wet-on-wet watercolor polygons and fine tapered ink arcs on textured paper.
func (wc *watercolor) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := w / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	seed := rand.Int63()
	rng := rand.New(rand.NewSource(seed))

	palette := c.Opts().ColorSchema()
	if len(palette) == 0 {
		palette = common.Dunhuang
	}

	// Multiply blend mode so overlapping washes darken organically like real watercolor pigments.
	rc.SetBlendMode(render.BlendMultiply)

	cx, cy := w*0.5, h*0.5
	for i := 0; i < wc.shapeNum; i++ {
		col := palette[i%len(palette)]
		ang0 := float64(i) * (2.0 * math.Pi / float64(wc.shapeNum))
		dist := (55.0 + rng.Float64()*115.0) * scale
		px := cx + dist*math.Cos(ang0)
		py := cy + dist*math.Sin(ang0)
		rad := (95.0 + rng.Float64()*65.0) * scale

		poly := make([]render.Point, 8)
		for v := 0; v < 8; v++ {
			a := float64(v)*(2.0*math.Pi/8.0) + rng.Float64()*0.2
			r := rad * (0.85 + 0.3*rng.Float64())
			poly[v] = render.Pt(px+r*math.Cos(a), py+r*math.Sin(a))
		}
		rc.DrawWatercolorPolygon(poly, col, wc.layers, wc.subdivisions, wc.roughness, 0.042, rng)
	}

	// Fine tapered ink accent arcs on top.
	rc.SetBlendMode(render.BlendOver)
	inkCol := c.Opts().LineColor()
	if inkCol.A == 0 || inkCol == common.Tomato {
		inkCol = color.RGBA{R: 32, G: 28, B: 38, A: 255}
	}
	for i := 0; i < 12; i++ {
		r := (50.0 + float64(i)*18.0) * scale
		startA := rng.Float64() * math.Pi * 2.0
		span := 0.8 + rng.Float64()*1.8
		pts := make([]render.Point, 10)
		for k := 0; k < 10; k++ {
			a := startA + span*(float64(k)/9.0)
			pts[k] = render.Pt(cx+r*math.Cos(a), cy+r*math.Sin(a))
		}
		rc.DrawTaperedSpline(
			pts,
			2.2*scale,
			render.TaperBothEnds,
			func(float64) color.RGBA { return inkCol },
			0.55,
		)
	}

	rc.ApplyPaperTexture(0.55, rand.Int63())
	rc.ApplyVignette(0.18)
	c.MergeRenderCanvas(rc)
}
