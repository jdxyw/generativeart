package arts

import (
	"image/color"
	"math"
	"math/rand"
	"sort"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type cliffordNebula struct {
	pointsN    int
	a, b, c, d float64
}

// NewCliffordNebula returns a cliffordNebula engine that renders a 2D Clifford strange
// attractor using sub-pixel bilinear histogram splatting, velocity-phase Oklab color
// accumulation, logarithmic HDR tone mapping, and star-cluster glow highlights.
func NewCliffordNebula(pointsN int, a, b, c, d float64) *cliffordNebula {
	if pointsN <= 0 {
		pointsN = 1500000
	}
	if a == 0 && b == 0 && c == 0 && d == 0 {
		a, b, c, d = -1.7, 1.8, -1.9, -0.4
	}
	return &cliffordNebula{
		pointsN: pointsN,
		a:       a,
		b:       b,
		c:       c,
		d:       d,
	}
}

// Generative renders the Clifford strange attractor nebula onto the canvas.
func (cn *cliffordNebula) Generative(c *generativeart.Canva) {
	if cn.pointsN <= 0 {
		cn.pointsN = 1500000
	}
	if cn.a == 0 && cn.b == 0 && cn.c == 0 && cn.d == 0 {
		cn.a, cn.b, cn.c, cn.d = -1.7, 1.8, -1.9, -0.4
	}

	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	minDim := math.Min(fw, fh)
	scale := minDim / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	nPix := w * h
	density := make([]float64, nPix)
	colorAcc := make([]float64, nPix)
	phaseAcc := make([]float64, nPix)

	x, y := 0.1, 0.1
	// Warm up attractor trajectory to settle onto the invariant manifold.
	for i := 0; i < 64; i++ {
		nx := math.Sin(cn.a*y) + cn.c*math.Cos(cn.a*x)
		ny := math.Sin(cn.b*x) + cn.d*math.Cos(cn.b*y)
		x, y = nx, ny
	}

	cx := fw * 0.5
	cy := fh * 0.5
	// Map attractor range [-2.9, 2.9] to centered canvas coordinates.
	coordScale := (minDim * 0.46) / 2.9

	for i := 0; i < cn.pointsN; i++ {
		nx := math.Sin(cn.a*y) + cn.c*math.Cos(cn.a*x)
		ny := math.Sin(cn.b*x) + cn.d*math.Cos(cn.b*y)

		dx := nx - x
		dy := ny - y
		vel := math.Hypot(dx, dy)
		phase := (math.Atan2(dy, dx) + math.Pi) / (2.0 * math.Pi)

		velNorm := common.Clamp01(vel / 4.2)
		colorIdx := common.Clamp01(0.65*velNorm + 0.35*phase)

		px := cx + nx*coordScale
		py := cy + ny*coordScale
		x, y = nx, ny

		x0 := int(math.Floor(px))
		y0 := int(math.Floor(py))
		if x0 < 0 || x0 >= w-1 || y0 < 0 || y0 >= h-1 {
			continue
		}

		fx := px - float64(x0)
		fy := py - float64(y0)
		w00 := (1.0 - fx) * (1.0 - fy)
		w10 := fx * (1.0 - fy)
		w01 := (1.0 - fx) * fy
		w11 := fx * fy

		idx00 := y0*w + x0
		idx10 := idx00 + 1
		idx01 := idx00 + w
		idx11 := idx01 + 1

		density[idx00] += w00
		density[idx10] += w10
		density[idx01] += w01
		density[idx11] += w11

		colorAcc[idx00] += w00 * colorIdx
		colorAcc[idx10] += w10 * colorIdx
		colorAcc[idx01] += w01 * colorIdx
		colorAcc[idx11] += w11 * colorIdx

		phaseAcc[idx00] += w00 * phase
		phaseAcc[idx10] += w10 * phase
		phaseAcc[idx01] += w01 * phase
		phaseAcc[idx11] += w11 * phase
	}

	// Collect non-zero densities to compute the 99.8th percentile reference density.
	nonZero := make([]float64, 0, nPix/4)
	maxDensity := 0.0
	for _, d := range density {
		if d > 0 {
			nonZero = append(nonZero, d)
			if d > maxDensity {
				maxDensity = d
			}
		}
	}
	if len(nonZero) > 100 {
		sort.Float64s(nonZero)
		pIdx := int(float64(len(nonZero)-1) * 0.998)
		refD := nonZero[pIdx]
		if refD > 0 {
			maxDensity = refD
		}
	}
	if maxDensity <= 0 {
		maxDensity = 1.0
	}

	schema := c.Opts().ColorSchema()
	if len(schema) == 0 {
		schema = common.Inferno
	}
	grad := render.NewMultiStopGradient(schema...)
	coreGold := color.RGBA{R: 255, G: 248, B: 224, A: 255}

	rc := c.NewRenderCanvas()
	rc.SetBlendMode(render.BlendAdd)

	logDenom := math.Log(1.0 + maxDensity*18.0)
	if logDenom <= 1e-9 {
		logDenom = 1.0
	}

	type peakCandidate struct {
		x, y int
		val  float64
		col  color.RGBA
	}
	var peaks []peakCandidate

	for py := 0; py < h; py++ {
		rowOff := py * w
		for px := 0; px < w; px++ {
			idx := rowOff + px
			d := density[idx]
			if d <= 0 {
				continue
			}

			dNorm := math.Log(1.0+d*18.0) / logDenom
			dNorm = common.Clamp01(dNorm)
			mapped := math.Pow(dNorm, 0.72)

			avgColIdx := common.Clamp01(colorAcc[idx] / d)
			avgPhase := common.Clamp01(phaseAcc[idx] / d)

			// Combine velocity/phase index with density progression for rich nebula coloring.
			tGrad := common.Clamp01(0.55*mapped + 0.45*avgColIdx)
			baseCol := grad.SampleOklab(tGrad)
			iridCol := common.CosineIridescent.Sample(avgPhase*0.8 + mapped*0.35)

			nebulaCol := render.LerpOklab(baseCol, iridCol, 0.28*(1.0-mapped*0.5))
			if mapped > 0.68 {
				coreT := common.Smoothstep(0.68, 1.0, mapped)
				nebulaCol = render.LerpOklab(nebulaCol, coreGold, coreT*0.75)
			}

			rLin, gLin, bLin, _ := render.ColorToLinearRGB(nebulaCol)
			hdrBoost := 1.0 + 1.45*math.Pow(mapped, 2.2)
			rc.BlendPixelLinear(px, py, rLin*hdrBoost, gLin*hdrBoost, bLin*hdrBoost, mapped, render.BlendAdd)

			// Record local maxima candidates on a strided grid for star-cluster highlights.
			if dNorm > 0.78 && px > 4 && px < w-4 && py > 4 && py < h-4 {
				if d >= density[idx-1] && d >= density[idx+1] && d >= density[idx-w] && d >= density[idx+w] {
					peaks = append(peaks, peakCandidate{x: px, y: py, val: d, col: nebulaCol})
				}
			}
		}
	}

	// Sort peaks by density descending and pick up to 35 well-separated star-cluster highlights.
	sort.Slice(peaks, func(i, j int) bool {
		return peaks[i].val > peaks[j].val
	})

	rng := rand.New(rand.NewSource(42))
	selected := make([]peakCandidate, 0, 35)
	minSep := 18.0 * scale
	for _, pk := range peaks {
		if len(selected) >= 35 {
			break
		}
		tooClose := false
		for _, sel := range selected {
			if math.Hypot(float64(pk.x-sel.x), float64(pk.y-sel.y)) < minSep {
				tooClose = true
				break
			}
		}
		if !tooClose {
			selected = append(selected, pk)
		}
	}

	rc.SetBlendMode(render.BlendAdd)
	for i, pk := range selected {
		u := 1.0 - float64(i)/float64(common.MaxInt(len(selected), 1))
		rad := (2.2 + 5.0*u + rng.Float64()*2.5) * scale
		starCol := render.LerpOklab(pk.col, coreGold, 0.65)
		rc.DrawGlowDot(float64(pk.x)+0.5, float64(pk.y)+0.5, rad, starCol, 0.28+0.25*u, 0.18)
	}

	rc.ApplyBloom(0.45, 9.0, 0.65)
	rc.ApplyVignette(0.3)
	c.MergeRenderCanvasToneMapped(rc, 1.15, 0)
}
