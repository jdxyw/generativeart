package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type sandRing struct {
	ringNum       int
	grainsPerStep int
	noiseScale    float64
}

// NewSandRing returns a sandRing object that renders concentric FBM-deformed
// stochastic sand-spline rings and sub-pixel bokeh particles in linear HDR light.
func NewSandRing(ringNum, grainsPerStep int, noiseScale float64) *sandRing {
	if ringNum <= 0 {
		ringNum = 22
	}
	if grainsPerStep <= 0 {
		grainsPerStep = 22
	}
	if noiseScale <= 0 {
		noiseScale = 1.3
	}
	return &sandRing{
		ringNum:       ringNum,
		grainsPerStep: grainsPerStep,
		noiseScale:    noiseScale,
	}
}

// Generative draws concentric sand-painted rings and glowing bokeh particles.
func (sr *sandRing) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := w / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	seed := rand.Int63()
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	schema := c.Opts().ColorSchema()
	if len(schema) == 0 {
		schema = common.Plasma
	}
	grad := render.NewMultiStopGradient(schema...)

	rc.SetBlendMode(render.BlendAdd)
	cx, cy := w*0.5, h*0.5

	// Concentric FBM-deformed sand splines around canvas center.
	denom := float64(sr.ringNum - 1)
	if denom < 1 {
		denom = 1
	}
	for rIdx := 0; rIdx < sr.ringNum; rIdx++ {
		u := float64(rIdx) / denom
		baseR := 35.0*scale + u*w*0.38
		nCtrl := 24
		pts := make([]render.Point, 0, nCtrl+1)
		for k := 0; k <= nCtrl; k++ {
			ang := (float64(k) / float64(nCtrl)) * 2.0 * math.Pi
			nVal := noise.FBM2D(math.Cos(ang)*sr.noiseScale+u*2.0, math.Sin(ang)*sr.noiseScale+u*2.0, 4, 2.0, 0.5)
			rad := baseR * (1.0 + 0.22*nVal)
			pts = append(pts, render.Pt(cx+rad*math.Cos(ang), cy+rad*math.Sin(ang)))
		}
		rc.DrawSandSpline(
			pts,
			sr.grainsPerStep,
			(3.5+u*8.5)*scale,
			func(t float64) color.RGBA {
				return grad.SampleOklab(math.Mod(u*0.7+t*0.3, 1.0))
			},
			0.032,
			rng,
		)
	}

	// Luminous bokeh particles and soft rings.
	for i := 0; i < 95; i++ {
		ang := rng.Float64() * 2.0 * math.Pi
		rad := math.Sqrt(rng.Float64()) * w * 0.43
		px := cx + rad*math.Cos(ang)
		py := cy + rad*math.Sin(ang)
		col := grad.SampleOklab(rng.Float64())
		if i%6 == 0 {
			rc.DrawSoftRing(px, py, (6.0+rng.Float64()*18.0)*scale, 1.8*scale, col, 0.35)
		} else {
			rc.DrawGlowDot(px, py, (2.0+rng.Float64()*10.0)*scale, col, 0.42, 0.15)
		}
	}

	rc.ApplyBloom(0.45, 10.0, 0.65)
	rc.ApplyVignette(0.3)
	c.MergeRenderCanvasToneMapped(rc, 1.15, 0)
}
