package arts

import (
	"image/color"
	"math"
	"runtime"
	"sync"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type kuramotoWave struct {
	defectsN    int
	streamlines int
}

// NewKuramotoWave returns a kuramotoWave engine that simulates a 2D coupled
// phase-oscillator chimera field with topological vortex/antivortex defects,
// branch-cut-free cyclic Oklab shading, 3D Blinn-Phong phase ridges, and
// Jobard-Lefer evenly-spaced phase-gradient streamlines.
func NewKuramotoWave(defectsN, streamlines int) *kuramotoWave {
	if defectsN <= 0 {
		defectsN = 8
	}
	if streamlines <= 0 {
		streamlines = 180
	}
	return &kuramotoWave{
		defectsN:    defectsN,
		streamlines: streamlines,
	}
}

// Generative renders the Kuramoto coupled phase-oscillator topological defect field.
func (kw *kuramotoWave) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := c.Width(), c.Height()
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0

	// Cyclic Branch-Cut-Free Oklab Phase Wheel Palette
	abyssCore := color.RGBA{R: 8, G: 11, B: 22, A: 255}
	cyclicRing := []color.RGBA{
		{R: 20, G: 58, B: 142, A: 255},  // 0 deg: Deep Lapis Cobalt
		{R: 38, G: 176, B: 182, A: 255}, // 60 deg: Jingtai Turquoise
		{R: 246, G: 192, B: 58, A: 255}, // 120 deg: Imperial Gold
		{R: 222, G: 64, B: 46, A: 255},  // 180 deg: Cinnabar Coral
		{R: 162, G: 28, B: 98, A: 255},  // 240 deg: Byzantium Magenta
		{R: 62, G: 26, B: 118, A: 255}, // 300 deg: Royal Indigo Violet
	}
	ivorySpec := color.RGBA{R: 255, G: 250, B: 228, A: 255}
	goldStream := color.RGBA{R: 255, G: 222, B: 118, A: 255}

	rc.FillBackground(abyssCore)
	rc.SetBlendMode(render.BlendOver)

	type phaseDefect struct {
		x, y   float64
		charge float64 // +1 or -1 topological winding number
		spin   float64 // spiral twist factor
	}

	numD := kw.defectsN
	if numD < 4 {
		numD = 8
	}
	defects := make([]phaseDefect, 0, numD)
	for i := 0; i < numD; i++ {
		ang := (float64(i)/float64(numD))*2.0*math.Pi + 0.28
		rad := (0.24 + 0.08*math.Sin(float64(i)*2.1)) * math.Min(fw, fh)
		q := 1.0
		if i%2 == 1 {
			q = -1.0
		}
		defects = append(defects, phaseDefect{
			x:      0.5*fw + rad*math.Cos(ang),
			y:      0.5*fh + rad*math.Sin(ang),
			charge: q,
			spin:   q * 0.032 / scale,
		})
	}

	ne := common.NewNoiseEngine(c.Seed() + 1907)
	coreXi := 22.0 * scale

	// Evaluate complex order parameter Psi(x, y) = A(x, y) * exp(i * Theta(x, y))
	evalOrderParam := func(x, y float64) (amp, theta float64) {
		// Smooth background Kuramoto wave packet
		thetaSum := 1.8 * ne.FBM2D(x/(210.0*scale), y/(210.0*scale), 3, 2.0, 0.5)
		ampProd := 1.0

		for _, d := range defects {
			dx := x - d.x
			dy := y - d.y
			r := math.Hypot(dx, dy)
			ang := math.Atan2(dy, dx)
			// Spiral phase contribution: q * atan2(dy, dx) + spin * r * exp(-r/L)
			spiral := d.spin * r * math.Exp(-r/(180.0*scale))
			thetaSum += d.charge*ang + spiral
			// Ginzburg-Landau vortex core amplitude suppression tanh(r / xi)
			ampProd *= math.Tanh(r / coreXi)
		}
		return ampProd, thetaSum
	}

	sampleCyclicColor := func(theta float64) color.RGBA {
		// Normalize theta into [0, 1) periodically with zero branch-cut seam
		u := math.Mod(theta/(2.0*math.Pi), 1.0)
		if u < 0 {
			u += 1.0
		}
		pos := u * float64(len(cyclicRing))
		idx0 := int(math.Floor(pos)) % len(cyclicRing)
		idx1 := (idx0 + 1) % len(cyclicRing)
		frac := pos - math.Floor(pos)
		return render.LerpOklab(cyclicRing[idx0], cyclicRing[idx1], frac)
	}

	// 1. Parallel Evaluation of the Continuous Phase Field & 3D Harmonic Ridge Lighting
	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
	eps := 1.5 * scale
	var wg sync.WaitGroup
	for wIdx := 0; wIdx < workers; wIdx++ {
		y0 := (wIdx * h) / workers
		y1 := ((wIdx + 1) * h) / workers
		wg.Add(1)
		go func(rowStart, rowEnd int) {
			defer wg.Done()
			for y := rowStart; y < rowEnd; y++ {
				fy := float64(y) + 0.5
				for x := 0; x < w; x++ {
					fx := float64(x) + 0.5
					amp, th := evalOrderParam(fx, fy)
					_, thX := evalOrderParam(fx+eps, fy)
					_, thY := evalOrderParam(fx, fy+eps)

					// Periodic height function h = amp * sin(2 * theta) for 3D wave-front relief
					h0 := amp * math.Sin(2.0*th)
					hX := amp * math.Sin(2.0*thX)
					hY := amp * math.Sin(2.0*thY)
					gx := (hX - h0) / eps
					gy := (hY - h0) / eps

					nx, ny, nz := -gx*12.0, -gy*12.0, 1.0
					nLen := math.Sqrt(nx*nx + ny*ny + nz*nz)
					nx, ny, nz = nx/nLen, ny/nLen, nz/nLen

					ndotl := common.Constrain(-0.60*nx-0.65*ny+0.46*nz, 0, 1)
					spec := math.Pow(common.Constrain(-0.42*nx-0.48*ny+0.77*nz, 0, 1), 22.0)

					baseCol := sampleCyclicColor(th)
					// Darken smoothly inside topological vortex cores where amplitude A(x, y) -> 0
					col := render.LerpOklab(abyssCore, baseCol, common.Constrain(amp*(0.42+0.68*ndotl), 0, 1))
					if spec > 0.04 {
						col = render.LerpOklab(col, ivorySpec, common.Constrain(spec*0.78*amp, 0, 1))
					}

					rc.SetPixelLinear(x, y, render.SRGB8ToLinear(col.R), render.SRGB8ToLinear(col.G), render.SRGB8ToLinear(col.B), 1.0)
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	// 2. Trace Jobard-Lefer Evenly-Spaced Phase-Gradient Streamlines (grad Theta)
	phaseVelFn := func(x, y float64) render.Point {
		amp, _ := evalOrderParam(x, y)
		c0, s0 := 0.0, 0.0
		{
			_, th := evalOrderParam(x, y)
			c0, s0 = math.Cos(th), math.Sin(th)
		}
		_, thX := evalOrderParam(x+eps, y)
		_, thY := evalOrderParam(x, y+eps)
		// Unwrapped phase gradient: grad(theta) = cos(theta)*grad(sin(theta)) - sin(theta)*grad(cos(theta))
		dcdx := (math.Cos(thX) - c0) / eps
		dsdx := (math.Sin(thX) - s0) / eps
		dcdy := (math.Cos(thY) - c0) / eps
		dsdy := (math.Sin(thY) - s0) / eps
		vx := (c0*dsdx - s0*dcdx) * (0.2 + 0.8*amp)
		vy := (c0*dsdy - s0*dcdy) * (0.2 + 0.8*amp)
		return render.Pt(vx, vy)
	}

	streamlines := render.TraceEvenStreamlines(render.StreamlineConfig{
		Width:    fw,
		Height:   fh,
		DSep:     8.5 * scale,
		DTest:    4.4 * scale,
		StepSize: 2.0 * scale,
		MaxSteps: 140,
		Margin:   10.0 * scale,
	}, phaseVelFn)

	for idx, line := range streamlines {
		if len(line) < 6 {
			continue
		}
		lineW := 1.15 * scale
		alpha := 0.62
		if idx%5 == 0 {
			lineW = 1.85 * scale
			alpha = 0.88
		}
		rc.DrawTaperedSpline(line, lineW, render.TaperBothEnds, func(t float64) color.RGBA {
			pt := line[common.ConstrainInt(int(t*float64(len(line)-1)), 0, len(line)-1)]
			_, th := evalOrderParam(pt.X, pt.Y)
			cPhase := sampleCyclicColor(th + 0.6)
			return render.LerpOklab(cPhase, goldStream, 0.65)
		}, alpha)
	}

	// 3. Highlight the Topological Vortex & Antivortex Singularity Cores
	for _, d := range defects {
		coreCol := goldStream
		if d.charge < 0 {
			coreCol = color.RGBA{R: 72, G: 238, B: 248, A: 255}
		}
		rc.DrawSoftRing(d.x, d.y, 10.5*scale, 1.3*scale, coreCol, 0.82)
		rc.DrawGlowDot(d.x, d.y, 3.8*scale, ivorySpec, 0.96, 0.85)
	}

	rc.ApplyBloom(0.80, 6, 0.24)
	c.MergeRenderCanvas(rc)
}
