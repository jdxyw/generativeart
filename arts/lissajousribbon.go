package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type lissajousRibbon struct {
	curvesN             int
	freqX, freqY, freqZ float64
}

// NewLissajousRibbon returns a lissajousRibbon engine that renders a 3D iridescent
// harmonic silk caustic sculpture using perspective-projected Lissajous manifolds.
func NewLissajousRibbon(curvesN int, freqX, freqY, freqZ float64) *lissajousRibbon {
	if curvesN <= 0 {
		curvesN = 140
	}
	if freqX <= 0 {
		freqX = 3.0
	}
	if freqY <= 0 {
		freqY = 4.0
	}
	if freqZ <= 0 {
		freqZ = 5.0
	}
	return &lissajousRibbon{
		curvesN: curvesN,
		freqX:   freqX,
		freqY:   freqY,
		freqZ:   freqZ,
	}
}

type ribbonSample struct {
	pt       render.Point
	x3, y3   float64
	z3       float64
	caustic  float64
	depthFac float64
	col      color.RGBA
}

// evalHarmonograph3D evaluates the 3D phase-shifted harmonograph/Lissajous surface at (u, s)
// and applies a smooth 3D rotation across the ribbon family.
func (lr *lissajousRibbon) evalHarmonograph3D(u, s float64) (rx, ry, rz float64) {
	x := math.Sin(lr.freqX*u+s*1.4)*math.Cos(u*0.5+s*0.3) + 0.35*math.Sin(7.0*u-s*2.1)
	y := math.Sin(lr.freqY*u+s*0.9)*math.Sin(u*0.5+s*0.5) + 0.35*math.Cos(5.0*u+s*1.7)
	z := math.Cos(lr.freqZ*u + s*1.2)

	// Slow 3D rotation (yaw + pitch) across s in [0, 1]
	yaw := -0.32 + s*0.64
	pitch := 0.26 + 0.16*math.Sin(s*math.Pi)

	cy, sy := math.Cos(yaw), math.Sin(yaw)
	cp, sp := math.Cos(pitch), math.Sin(pitch)

	x1 := x*cy + z*sy
	z1 := -x*sy + z*cy
	y1 := y*cp - z1*sp
	z2 := y*sp + z1*cp

	return x1, y1, z2
}

// Generative renders the 3D Iridescent Harmonic Silk Caustic Sculpture into c.
func (lr *lissajousRibbon) Generative(c *generativeart.Canva) {
	curvesN := lr.curvesN
	freqX, freqY, freqZ := lr.freqX, lr.freqY, lr.freqZ
	if curvesN <= 0 {
		curvesN = 140
	}
	if freqX <= 0 {
		freqX = 3.0
	}
	if freqY <= 0 {
		freqY = 4.0
	}
	if freqZ <= 0 {
		freqZ = 5.0
	}
	lr.curvesN = curvesN
	lr.freqX, lr.freqY, lr.freqZ = freqX, freqY, freqZ

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
	cx, cy := fw*0.5, fh*0.5

	schema := c.Opts().ColorSchema()
	if len(schema) == 0 {
		schema = common.Plasma
	}
	grad := render.NewMultiStopGradient(schema...)
	cyberGrad := render.NewMultiStopGradient(common.Cyberpunk...)

	rc := c.NewRenderCanvas()

	// Deep velvet cosmic background with subtle dual-lobe radial nebula glow
	bgBase := color.RGBA{R: 5, G: 4, B: 14, A: 255}
	rc.FillBackground(bgBase)
	rc.SetBlendMode(render.BlendAdd)
	rc.DrawGlowDot(cx-minDim*0.08, cy-minDim*0.06, minDim*0.48, grad.SampleOklab(0.22), 0.14, 0.0)
	rc.DrawGlowDot(cx+minDim*0.09, cy+minDim*0.07, minDim*0.44, cyberGrad.SampleOklab(0.55), 0.11, 0.0)

	const (
		uSteps = 200
		fov    = 3.15
	)
	projScale := minDim * 0.33

	denomK := float64(common.MaxInt(curvesN-1, 1))
	grid := make([][]ribbonSample, curvesN)

	// 1. Pre-evaluate all 3D curves and project to 2D with perspective fov / (fov + z).
	for k := 0; k < curvesN; k++ {
		s := float64(k) / denomK
		row := make([]ribbonSample, uSteps+1)
		for i := 0; i <= uSteps; i++ {
			u := (float64(i) / float64(uSteps)) * 2.0 * math.Pi
			x3, y3, z3 := lr.evalHarmonograph3D(u, s)
			persp := fov / math.Max(0.8, fov+z3)
			px := cx + x3*projScale*persp
			py := cy + y3*projScale*persp

			// Depth factor: closer to camera (smaller z3) is brighter and crisper
			depthFac := common.Clamp01((1.25 - z3) / 2.5)

			// Base Oklab palette blended with CosineIridescent
			uNorm := float64(i) / float64(uSteps)
			palT := common.Clamp01(0.55*s + 0.45*(0.5+0.5*math.Sin(u+s*2.2)))
			baseCol := grad.SampleOklab(palT)
			iridCol := common.CosineIridescent.Sample(uNorm*1.5 + s*0.85 + 0.25*z3)
			col := render.LerpOklab(baseCol, iridCol, 0.42+0.25*depthFac)

			row[i] = ribbonSample{
				pt:       render.Pt(px, py),
				x3:       x3,
				y3:       y3,
				z3:       z3,
				depthFac: depthFac,
				col:      col,
			}
		}
		grid[k] = row
	}

	// 2. Compute surface tangent angle & 3D ribbon normal to detect optical caustic folds
	// where the ribbon turns edge-on to the camera (|N.z| -> 0).
	for k := 0; k < curvesN; k++ {
		kNext := common.MinInt(k+1, curvesN-1)
		kPrev := common.MaxInt(k-1, 0)
		for i := 0; i <= uSteps; i++ {
			iNext := (i + 1) % uSteps
			iPrev := (i + uSteps - 1) % uSteps

			// Tangent along u
			txU := grid[k][iNext].x3 - grid[k][iPrev].x3
			tyU := grid[k][iNext].y3 - grid[k][iPrev].y3
			tzU := grid[k][iNext].z3 - grid[k][iPrev].z3

			// Tangent along s (across ribbon)
			txS := grid[kNext][i].x3 - grid[kPrev][i].x3
			tyS := grid[kNext][i].y3 - grid[kPrev][i].y3
			tzS := grid[kNext][i].z3 - grid[kPrev][i].z3

			// Cross product N = T_u x T_s
			nx := tyU*tzS - tzU*tyS
			ny := tzU*txS - txU*tzS
			nz := txU*tyS - tyU*txS
			nLen := math.Sqrt(nx*nx + ny*ny + nz*nz)
			absNz := 1.0
			if nLen > 1e-9 {
				absNz = math.Abs(nz) / nLen
			}

			// Edge-on caustic intensity peaks when absNz is near 0
			caustic := math.Pow(1.0-common.Clamp01(absNz), 2.4)
			grid[k][i].caustic = caustic
			if caustic > 0.45 {
				// Shift caustic fold highlights toward luminous champagne-white
				glowTint := color.RGBA{R: 255, G: 245, B: 225, A: 255}
				grid[k][i].col = render.LerpOklab(grid[k][i].col, glowTint, (caustic-0.45)*0.75)
			}
		}
	}

	rng := rand.New(rand.NewSource(rand.Int63()))
	rc.SetBlendMode(render.BlendAdd)

	// 3. Draw delicate sub-pixel sand filaments (18 grains) connecting corresponding points
	// between curve k and curve k+1, plus cross-sheet silk caustic struts.
	for k := 0; k < curvesN-1; k++ {
		row0 := grid[k]
		row1 := grid[k+1]
		for i := 0; i < uSteps; i++ {
			p0 := row0[i]
			p1 := row1[i]
			caustic := 0.5 * (p0.caustic + p1.caustic)
			depth := 0.5 * (p0.depthFac + p1.depthFac)

			sandAlpha := (0.012 + 0.032*caustic) * (0.45 + 0.55*depth)
			jitter := (0.55 + 0.85*(1.0-caustic)) * scale
			rc.DrawSandLine(p0.pt, p1.pt, 18, jitter, p0.col, sandAlpha, rng)
		}
	}

	// 4. Draw fine spline curves along each filament with depth & caustic modulated thickness and opacity.
	for k := 0; k < curvesN; k++ {
		row := grid[k]
		pts := make([]render.Point, len(row))
		for i := range row {
			pts[i] = row[i].pt
		}

		// Sample average depth & caustic along curve k
		var avgDepth, avgCaustic float64
		for i := 0; i < uSteps; i++ {
			avgDepth += row[i].depthFac
			avgCaustic += row[i].caustic
		}
		avgDepth /= float64(uSteps)
		avgCaustic /= float64(uSteps)

		maxWidth := (0.55 + 0.85*avgDepth + 0.65*avgCaustic) * scale
		strokeAlpha := 0.038 + 0.055*avgDepth + 0.045*avgCaustic
		if k%7 == 0 {
			// Highlighted structural harmonic ribs
			maxWidth *= 1.35
			strokeAlpha *= 1.45
		}

		rowCapture := row
		rc.DrawTaperedSpline(
			pts,
			maxWidth,
			func(t float64) float64 {
				idx := common.ConstrainInt(int(t*float64(uSteps)), 0, uSteps)
				sample := rowCapture[idx]
				// Thinner, razor-sharp caustic edge where ribbon turns edge-on, fuller in mid-sweep
				return 0.45 + 0.55*sample.depthFac + 0.40*sample.caustic
			},
			func(t float64) color.RGBA {
				idx := common.ConstrainInt(int(t*float64(uSteps)), 0, uSteps)
				return rowCapture[idx].col
			},
			strokeAlpha,
		)
	}

	// 5. Add subtle bokeh glow dots at optical caustic cusps where ribbon folds turn sharply.
	for k := 0; k < curvesN; k += 3 {
		row := grid[k]
		for i := 0; i < uSteps; i += 2 {
			sample := row[i]
			if sample.caustic > 0.72 && rng.Float64() < 0.28 {
				bokehRad := (1.8 + 4.5*sample.caustic*sample.depthFac) * scale
				bokehAlpha := (0.04 + 0.10*sample.caustic) * (0.5 + 0.5*sample.depthFac)
				rc.DrawGlowDot(sample.pt.X, sample.pt.Y, bokehRad, sample.col, bokehAlpha, 0.12)
				if sample.caustic > 0.84 && bokehRad > 3.2*scale {
					rc.DrawSoftRing(sample.pt.X, sample.pt.Y, bokehRad*1.35, 0.75*scale, sample.col, bokehAlpha*0.55)
				}
			}
		}
	}

	rc.ApplyBloom(0.45, 9.0, 0.6)
	rc.ApplyVignette(0.3)
	c.MergeRenderCanvasToneMapped(rc, 1.15, 0)
}
