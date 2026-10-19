package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type curlFlow struct {
	lineNum    int
	stepNum    int
	noiseScale float64
}

// NewCurlFlow returns a curlFlow object that renders divergence-free curl-noise
// streamlines with calligraphic tapered splines, multi-filament bristle strokes, and HDR bloom.
func NewCurlFlow(lineNum, stepNum int, noiseScale float64) *curlFlow {
	if lineNum <= 0 {
		lineNum = 110
	}
	if stepNum <= 0 {
		stepNum = 36
	}
	if noiseScale <= 0 {
		noiseScale = 0.0045
	}
	return &curlFlow{
		lineNum:    lineNum,
		stepNum:    stepNum,
		noiseScale: noiseScale,
	}
}

// Generative draws a curl-noise flow field image using the HDR render canvas.
func (cf *curlFlow) Generative(c *generativeart.Canva) {
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
		schema = common.Cyberpunk
	}
	grad := render.NewMultiStopGradient(schema...)
	warmGrad := render.NewMultiStopGradient(common.Sunset...)

	// Soft background glow nebula in additive HDR mode.
	rc.SetBlendMode(render.BlendAdd)
	for i := 0; i < 18; i++ {
		gx := rng.Float64() * w
		gy := rng.Float64() * h
		rad := (80.0 + rng.Float64()*140.0) * scale
		col := grad.SampleOklab(rng.Float64())
		rc.DrawGlowDot(gx, gy, rad, col, 0.08, 0.0)
	}

	// Primary curl-noise ribbons and bristle brush strokes.
	rc.SetBlendMode(render.BlendOver)
	ns := cf.noiseScale / scale
	stepLen := 7.5 * scale
	margin := 30.0 * scale

	for i := 0; i < cf.lineNum; i++ {
		x := rng.Float64() * w
		y := rng.Float64() * h
		pts := make([]render.Point, 0, cf.stepNum)
		for step := 0; step < cf.stepNum; step++ {
			pts = append(pts, render.Pt(x, y))
			v := noise.Curl2D(x*ns, y*ns, 0.01)
			x += v.X * stepLen
			y += v.Y * stepLen
			if x < -margin || x > w+margin || y < -margin || y > h+margin {
				break
			}
		}
		if len(pts) < 4 {
			continue
		}
		baseT := rng.Float64()
		maxW := (1.8 + rng.Float64()*7.5) * scale
		if i%5 == 0 {
			rc.DrawBristleStroke(pts, maxW*1.6, 14, warmGrad.SampleOklab(baseT), 0.38, rng)
		} else {
			rc.DrawTaperedSpline(
				pts,
				maxW,
				render.TaperCalligraphic,
				func(t float64) color.RGBA {
					return grad.SampleOklab(math.Mod(baseT+t*0.35, 1.0))
				},
				0.78,
			)
		}
	}

	// Luminous highlight filaments in additive mode.
	rc.SetBlendMode(render.BlendAdd)
	highlightCount := cf.lineNum / 3
	if highlightCount < 12 {
		highlightCount = 12
	}
	highlightSteps := (cf.stepNum * 3) / 4
	if highlightSteps < 8 {
		highlightSteps = 8
	}
	for i := 0; i < highlightCount; i++ {
		x := rng.Float64() * w
		y := rng.Float64() * h
		pts := make([]render.Point, 0, highlightSteps)
		for step := 0; step < highlightSteps; step++ {
			pts = append(pts, render.Pt(x, y))
			v := noise.Curl2D(x*ns, y*ns, 0.01)
			x += v.X * (6.0 * scale)
			y += v.Y * (6.0 * scale)
		}
		if len(pts) < 2 {
			continue
		}
		t0 := rng.Float64()
		rc.DrawTaperedSpline(
			pts,
			1.6*scale,
			render.TaperBothEnds,
			func(t float64) color.RGBA {
				return warmGrad.SampleOklab(math.Mod(t0+t*0.4, 1.0))
			},
			0.45,
		)
	}

	rc.ApplyBloom(0.55, 8.0, 0.55)
	rc.ApplyVignette(0.35)
	c.MergeRenderCanvasToneMapped(rc, 1.05, 0)
}
