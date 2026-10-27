package arts

import (
	"image/color"
	"math"
	"runtime"
	"sort"
	"sync"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type lagrangianCoherent struct {
	advSteps int
	ftlScale float64
}

// NewLagrangianCoherent returns a lagrangianCoherent engine that computes the
// forward and backward Finite-Time Lyapunov Exponent (FTLE) fields from the
// Cauchy-Green strain tensor C = (grad Phi)^T (grad Phi) of a chaotic vortex flow,
// revealing intersecting repelling (gold/coral) and attracting (cyan/sapphire)
// Lagrangian Coherent Structures (homoclinic/heteroclinic tangles).
func NewLagrangianCoherent(advSteps int, ftlScale float64) *lagrangianCoherent {
	if advSteps <= 0 {
		advSteps = 18
	}
	if ftlScale <= 0 {
		ftlScale = 1.0
	}
	return &lagrangianCoherent{
		advSteps: advSteps,
		ftlScale: ftlScale,
	}
}

// Generative renders the forward and backward FTLE Lagrangian Coherent Structures.
func (lc *lagrangianCoherent) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := c.Width(), c.Height()
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0

	// Dual-Manifold LCS Palette (Attracting Unstable Manifold = Cyan/Cobalt, Repelling Stable Manifold = Gold/Coral)
	abyssFluid := color.RGBA{R: 5, G: 8, B: 18, A: 255}
	cobaltDeep := color.RGBA{R: 18, G: 52, B: 132, A: 255}
	unstableCyan := color.RGBA{R: 46, G: 208, B: 232, A: 255}
	stableCoral := color.RGBA{R: 224, G: 62, B: 48, A: 255}
	stableGold := color.RGBA{R: 252, G: 196, B: 58, A: 255}
	tangleIvory := color.RGBA{R: 255, G: 252, B: 236, A: 255}

	rc.FillBackground(abyssFluid)
	rc.SetBlendMode(render.BlendOver)

	ne := common.NewNoiseEngine(c.Seed() + 2001)

	// 7-vortex chaotic advection constellation + Kelvin-Helmholtz shear
	vortexCores := [][4]float64{
		{0.32 * fw, 0.34 * fh, 54.0 * scale, 1.45},
		{0.68 * fw, 0.30 * fh, 50.0 * scale, -1.35},
		{0.66 * fw, 0.68 * fh, 54.0 * scale, 1.45},
		{0.30 * fw, 0.66 * fh, 50.0 * scale, -1.35},
		{0.50 * fw, 0.50 * fh, 46.0 * scale, 1.60},
		{0.20 * fw, 0.50 * fh, 44.0 * scale, -1.10},
		{0.80 * fw, 0.50 * fh, 44.0 * scale, 1.10},
	}

	velocityAt := func(x, y, t float64) (vx, vy float64) {
		nx := x / fw
		ny := y / fh
		// Smooth large-scale sinusoidal shear
		vx = 9.0 * scale * math.Sin(ny*2.0*math.Pi+0.35*math.Sin(t*1.5))
		vy = -9.0 * scale * math.Sin(nx*2.0*math.Pi+0.35*math.Cos(t*1.5))

		for i, vc := range vortexCores {
			cx := vc[0] + 18.0*scale*math.Cos(t*1.4+float64(i)*1.2)
			cy := vc[1] + 18.0*scale*math.Sin(t*1.4+float64(i)*1.2)
			dx := x - cx
			dy := y - cy
			r2 := dx*dx + dy*dy + vc[2]*vc[2]
			omega := vc[3] * (1650.0 * scale) / r2
			vx += -dy * omega
			vy += dx * omega
		}
		_ = ne
		return vx * lc.ftlScale, vy * lc.ftlScale
	}

	steps := lc.advSteps
	if steps < 10 || steps > 28 {
		steps = 18
	}
	dt := 0.22

	// Integrate trajectory Phi_0^T(x0, y0) using RK4 (dir = +1 for forward FTLE, -1 for backward FTLE)
	advectPoint := func(x0, y0, dir float64) (xEnd, yEnd float64) {
		px, py := x0, y0
		hStep := dt * dir
		for s := 0; s < steps; s++ {
			t := float64(s) * dt
			v1x, v1y := velocityAt(px, py, t)
			v2x, v2y := velocityAt(px+0.5*hStep*v1x, py+0.5*hStep*v1y, t+0.5*dt)
			v3x, v3y := velocityAt(px+0.5*hStep*v2x, py+0.5*hStep*v2y, t+0.5*dt)
			v4x, v4y := velocityAt(px+hStep*v3x, py+hStep*v3y, t+dt)
			px += (hStep / 6.0) * (v1x + 2.0*v2x + 2.0*v3x + v4x)
			py += (hStep / 6.0) * (v1y + 2.0*v2y + 2.0*v3y + v4y)
		}
		return px, py
	}

	// 1. Parallel Advection of Forward & Backward Flow Maps on Full Pixel Grid
	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
	fwdX := make([]float64, w*h)
	fwdY := make([]float64, w*h)
	bwdX := make([]float64, w*h)
	bwdY := make([]float64, w*h)

	var wg sync.WaitGroup
	for wIdx := 0; wIdx < workers; wIdx++ {
		y0 := (wIdx * h) / workers
		y1 := ((wIdx + 1) * h) / workers
		wg.Add(1)
		go func(rowStart, rowEnd int) {
			defer wg.Done()
			for y := rowStart; y < rowEnd; y++ {
				fy := float64(y) + 0.5
				rowOff := y * w
				for x := 0; x < w; x++ {
					fx := float64(x) + 0.5
					idx := rowOff + x
					fwdX[idx], fwdY[idx] = advectPoint(fx, fy, 1.0)
					bwdX[idx], bwdY[idx] = advectPoint(fx, fy, -1.0)
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	// 2. Compute Cauchy-Green Strain Tensor Eigenvalues lambda_max(F^T F) & FTLE Ridges
	ftleFwd := make([]float64, w*h)
	ftleBwd := make([]float64, w*h)

	computeFTLE := func(mapX, mapY []float64, x, y int) float64 {
		xM := common.MaxInt(0, x-1)
		xP := common.MinInt(w-1, x+1)
		yM := common.MaxInt(0, y-1)
		yP := common.MinInt(h-1, y+1)
		dxSpan := float64(common.MaxInt(1, xP-xM))
		dySpan := float64(common.MaxInt(1, yP-yM))

		f11 := (mapX[y*w+xP] - mapX[y*w+xM]) / dxSpan
		f21 := (mapY[y*w+xP] - mapY[y*w+xM]) / dxSpan
		f12 := (mapX[yP*w+x] - mapX[yM*w+x]) / dySpan
		f22 := (mapY[yP*w+x] - mapY[yM*w+x]) / dySpan

		// Cauchy-Green strain tensor C = F^T * F = [[c11, c12], [c12, c22]]
		c11 := f11*f11 + f21*f21
		c22 := f12*f12 + f22*f22
		c12 := f11*f12 + f21*f22

		trace := c11 + c22
		det := c11*c22 - c12*c12
		disc := math.Sqrt(math.Max(0, 0.25*trace*trace-det))
		lambdaMax := math.Max(1.0, 0.5*trace+disc)
		return 0.5 * math.Log(lambdaMax)
	}

	for wIdx := 0; wIdx < workers; wIdx++ {
		y0 := (wIdx * h) / workers
		y1 := ((wIdx + 1) * h) / workers
		wg.Add(1)
		go func(rowStart, rowEnd int) {
			defer wg.Done()
			for y := rowStart; y < rowEnd; y++ {
				rowOff := y * w
				for x := 0; x < w; x++ {
					ftleFwd[rowOff+x] = computeFTLE(fwdX, fwdY, x, y)
					ftleBwd[rowOff+x] = computeFTLE(bwdX, bwdY, x, y)
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	// Normalize FTLE fields by their 96th percentile (ignoring isolated single-pixel spikes)
	sampFwd := make([]float64, 0, w*h/4)
	sampBwd := make([]float64, 0, w*h/4)
	for i := 0; i < w*h; i += 4 {
		sampFwd = append(sampFwd, ftleFwd[i])
		sampBwd = append(sampBwd, ftleBwd[i])
	}
	sort.Float64s(sampFwd)
	sort.Float64s(sampBwd)
	p96Fwd := math.Max(0.1, sampFwd[int(float64(len(sampFwd))*0.96)])
	p96Bwd := math.Max(0.1, sampBwd[int(float64(len(sampBwd))*0.96)])
	invFwd := 1.0 / p96Fwd
	invBwd := 1.0 / p96Bwd

	// 3. Shade Dual-Manifold FTLE Ridges with 3D Normal Specular Lighting
	for wIdx := 0; wIdx < workers; wIdx++ {
		y0 := (wIdx * h) / workers
		y1 := ((wIdx + 1) * h) / workers
		wg.Add(1)
		go func(rowStart, rowEnd int) {
			defer wg.Done()
			for y := rowStart; y < rowEnd; y++ {
				rowOff := y * w
				yP := common.MinInt(h-1, y+1)
				for x := 0; x < w; x++ {
					xP := common.MinInt(w-1, x+1)
					idx := rowOff + x

					sFwd := common.Constrain(ftleFwd[idx]*invFwd, 0, 1)
					sBwd := common.Constrain(ftleBwd[idx]*invBwd, 0, 1)

					h0 := math.Max(sFwd, sBwd)
					hX := math.Max(common.Constrain(ftleFwd[rowOff+xP]*invFwd, 0, 1), common.Constrain(ftleBwd[rowOff+xP]*invBwd, 0, 1))
					hY := math.Max(common.Constrain(ftleFwd[yP*w+x]*invFwd, 0, 1), common.Constrain(ftleBwd[yP*w+x]*invBwd, 0, 1))

					gx := (hX - h0) * 14.0
					gy := (hY - h0) * 14.0
					nx, ny, nz := -gx, -gy, 1.0
					nLen := math.Sqrt(nx*nx + ny*ny + nz*nz)
					nx, ny, nz = nx/nLen, ny/nLen, nz/nLen

					ndotl := common.Constrain(-0.58*nx-0.64*ny+0.50*nz, 0, 1)
					spec := math.Pow(common.Constrain(-0.42*nx-0.48*ny+0.77*nz, 0, 1), 24.0)

					wFwd := math.Pow(sFwd, 1.55)
					wBwd := math.Pow(sBwd, 1.55)

					colFwd := render.LerpOklab(stableCoral, stableGold, common.Constrain(sFwd*1.1, 0, 1))
					colBwd := render.LerpOklab(cobaltDeep, unstableCyan, common.Constrain(sBwd*1.1, 0, 1))

					totW := wFwd + wBwd + 1e-5
					mixCol := render.LerpOklab(colBwd, colFwd, wFwd/totW)
					intensity := common.Constrain(math.Max(wFwd, wBwd)*1.18, 0, 1)

					col := render.LerpOklab(abyssFluid, mixCol, intensity*(0.38+0.62*ndotl))
					tangle := wFwd * wBwd
					if tangle > 0.35 {
						col = render.LerpOklab(col, tangleIvory, common.Constrain((tangle-0.35)*1.2, 0, 0.75))
					}
					if spec > 0.05 && intensity > 0.32 {
						col = render.LerpOklab(col, tangleIvory, common.Constrain(spec*0.65, 0, 1))
					}

					rc.SetPixelLinear(x, y, render.SRGB8ToLinear(col.R), render.SRGB8ToLinear(col.G), render.SRGB8ToLinear(col.B), 1.0)
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	// 4. Trace Jobard-Lefer Evenly-Spaced Lagrangian Transport Streamlines along the Instantaneous Hamiltonian Field
	flowVel := func(x, y float64) render.Point {
		vx, vy := velocityAt(x, y, 0.6)
		return render.Pt(vx, vy)
	}
	streamlines := render.TraceEvenStreamlines(render.StreamlineConfig{
		Width:    fw,
		Height:   fh,
		DSep:     8.5 * scale,
		DTest:    4.4 * scale,
		StepSize: 2.0 * scale,
		MaxSteps: 145,
		Margin:   8.0 * scale,
	}, flowVel)

	for idx, line := range streamlines {
		if len(line) < 6 {
			continue
		}
		lineW := 1.15 * scale
		alpha := 0.65
		if idx%4 == 0 {
			lineW = 1.75 * scale
			alpha = 0.88
		}
		rc.DrawTaperedSpline(line, lineW, render.TaperBothEnds, func(t float64) color.RGBA {
			pt := line[common.ConstrainInt(int(t*float64(len(line)-1)), 0, len(line)-1)]
			ix := common.ConstrainInt(int(pt.X), 0, w-1)
			iy := common.ConstrainInt(int(pt.Y), 0, h-1)
			sf := common.Constrain(ftleFwd[iy*w+ix]*invFwd, 0, 1)
			sb := common.Constrain(ftleBwd[iy*w+ix]*invBwd, 0, 1)
			if sf > sb {
				return render.LerpOklab(stableGold, tangleIvory, sf*0.55)
			}
			return render.LerpOklab(unstableCyan, tangleIvory, sb*0.55)
		}, alpha)
	}

	rc.ApplyBloom(0.80, 6, 0.26)
	c.MergeRenderCanvas(rc)
}
