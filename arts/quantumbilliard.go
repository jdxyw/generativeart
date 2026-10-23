package arts

import (
	"image/color"
	"math"
	"math/rand"
	"runtime"
	"sync"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type quantumBilliard struct {
	modesN     int
	waveNumber float64
}

// NewQuantumBilliard returns a quantumBilliard engine that renders Berry-Heller
// quantum chaos random-wave eigenfunctions and scarred probability densities
// Psi(x,y) = sum c_m exp(i k_m . r) alongside Bohmian quantum probability-current
// streamlines J = Im(Psi^* grad Psi) and Marching Squares nodal lines.
func NewQuantumBilliard(modesN int, waveNumber float64) *quantumBilliard {
	if modesN <= 0 {
		modesN = 48
	}
	if waveNumber <= 0 {
		waveNumber = 1.0
	}
	return &quantumBilliard{
		modesN:     modesN,
		waveNumber: waveNumber,
	}
}

// Generative renders the Berry-Heller quantum chaos eigenfunction scars and Bohmian streamlines.
func (qb *quantumBilliard) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := c.Width(), c.Height()
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0
	cx, cy := 0.5*fw, 0.5*fh

	// Quantum Probability Density & Bohmian Current Palette
	abyssVacuum := color.RGBA{R: 6, G: 9, B: 20, A: 255}
	deepIndigo := color.RGBA{R: 18, G: 38, B: 96, A: 255}
	quantumTeal := color.RGBA{R: 34, G: 168, B: 182, A: 255}
	plasmaCyan := color.RGBA{R: 112, G: 238, B: 246, A: 255}
	scarCoral := color.RGBA{R: 228, G: 68, B: 52, A: 255}
	scarGold := color.RGBA{R: 252, G: 196, B: 62, A: 255}
	ivoryPeak := color.RGBA{R: 255, G: 252, B: 236, A: 255}

	rc.FillBackground(abyssVacuum)
	rc.SetBlendMode(render.BlendOver)

	rng := rand.New(rand.NewSource(c.Seed() + 1926))
	numModes := qb.modesN
	if numModes < 24 {
		numModes = 48
	}

	// Wave-vector magnitude k0 on the energy shell |k| = k0
	k0 := (0.044 * qb.waveNumber) / scale

	type planeWaveMode struct {
		kx, ky     float64
		amp        float64
		phaseRe    float64
		phaseIm    float64
	}

	modes := make([]planeWaveMode, 0, numModes)
	// Superpose isotropic Berry random plane waves + 3 dominant scarred periodic-orbit beams (diamond/bow-tie billiard scars)
	scarAngles := []float64{math.Pi / 6.0, -math.Pi / 6.0, math.Pi / 2.0}
	for i := 0; i < numModes; i++ {
		ang := (float64(i)/float64(numModes))*2.0*math.Pi + (rng.Float64()-0.5)*0.08
		amp := 1.0 / math.Sqrt(float64(numModes))
		// Enhance amplitude near the 3 unstable periodic orbit scar angles (Heller quantum scars)
		for _, sa := range scarAngles {
			dAng := math.Abs(math.Sin(ang - sa))
			if dAng < 0.18 {
				amp *= 1.85
			}
		}
		modes = append(modes, planeWaveMode{
			kx:      k0 * math.Cos(ang),
			ky:      k0 * math.Sin(ang),
			amp:     amp,
			phaseRe: rng.Float64() * 2.0 * math.Pi,
			phaseIm: rng.Float64() * 2.0 * math.Pi,
		})
	}

	// Evaluate complex wavefunction Psi(x, y) = psiR + i*psiI and analytical gradients
	evalWaveFunction := func(x, y float64) (psiR, psiI, dRdx, dRdy, dIdx, dIdy float64) {
		dx := x - cx
		dy := y - cy
		for _, m := range modes {
			argRe := m.kx*dx + m.ky*dy + m.phaseRe
			argIm := m.kx*dx + m.ky*dy + m.phaseIm
			cR, sR := math.Cos(argRe)*m.amp, math.Sin(argRe)*m.amp
			cI, sI := math.Cos(argIm)*m.amp, math.Sin(argIm)*m.amp

			psiR += cR
			psiI += cI
			dRdx -= sR * m.kx
			dRdy -= sR * m.ky
			dIdx -= sI * m.kx
			dIdy -= sI * m.ky
		}
		return
	}

	// 1. Parallel Evaluation of Quantum Probability Density |Psi|^2 & 3D Wave-Packet Relief
	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
	probGrid := make([]float64, w*h)

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
					pR, pI, dRdx, dRdy, dIdx, dIdy := evalWaveFunction(fx, fy)

					prob := pR*pR + pI*pI
					probGrid[rowOff+x] = prob

					// Gradient of |Psi|^2 = 2*(psiR * grad(psiR) + psiI * grad(psiI))
					gPx := 2.0 * (pR*dRdx + pI*dIdx)
					gPy := 2.0 * (pR*dRdy + pI*dIdy)

					nx, ny, nz := -gPx*18.0*scale, -gPy*18.0*scale, 1.0
					nLen := math.Sqrt(nx*nx + ny*ny + nz*nz)
					nx, ny, nz = nx/nLen, ny/nLen, nz/nLen

					ndotl := common.Constrain(-0.60*nx-0.64*ny+0.48*nz, 0, 1)
					spec := math.Pow(common.Constrain(-0.42*nx-0.48*ny+0.77*nz, 0, 1), 24.0)

					// Normalize probability density |Psi|^2 into [0, 1]
					u := common.Constrain(math.Sqrt(prob)*0.52, 0, 1)
					phase := math.Atan2(pI, pR)
					warmMix := 0.5 + 0.5*math.Cos(phase)

					var baseCol color.RGBA
					switch {
					case u < 0.25:
						baseCol = render.LerpOklab(abyssVacuum, deepIndigo, u/0.25)
					case u < 0.55:
						cMid := render.LerpOklab(quantumTeal, scarCoral, warmMix*0.65)
						baseCol = render.LerpOklab(deepIndigo, cMid, (u-0.25)/0.30)
					case u < 0.82:
						cHigh := render.LerpOklab(plasmaCyan, scarGold, warmMix*0.82)
						baseCol = render.LerpOklab(quantumTeal, cHigh, (u-0.55)/0.27)
					default:
						baseCol = render.LerpOklab(scarGold, ivoryPeak, (u-0.82)/0.18)
					}

					col := render.LerpOklab(abyssVacuum, baseCol, common.Constrain(0.22+0.85*ndotl, 0, 1))
					if spec > 0.04 && u > 0.28 {
						col = render.LerpOklab(col, ivoryPeak, common.Constrain(spec*0.82, 0, 1))
					}

					rc.SetPixelLinear(x, y, render.SRGB8ToLinear(col.R), render.SRGB8ToLinear(col.G), render.SRGB8ToLinear(col.B), 1.0)
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	// 2. Extract Marching Squares Quantum Nodal & Scar Contours
	gridLookup := func(x, y float64) float64 {
		ix := common.ConstrainInt(int(x), 0, w-1)
		iy := common.ConstrainInt(int(y), 0, h-1)
		return math.Sqrt(probGrid[iy*w+ix]) * 0.52
	}
	for _, iso := range []float64{0.36, 0.58, 0.78} {
		contours := render.ExtractMarchingSquares(96, 96, fw, fh, iso, gridLookup)
		cPath := &canvas.Path{}
		for _, poly := range contours {
			if len(poly) < 4 {
				continue
			}
			smooth := render.CatmullRomSpline(poly, 3)
			cPath.MoveTo(smooth[0].X, smooth[0].Y)
			for k := 1; k < len(smooth); k++ {
				cPath.LineTo(smooth[k].X, smooth[k].Y)
			}
		}
		col := plasmaCyan
		if iso > 0.5 {
			col = scarGold
		}
		rc.StrokePath(cPath, 0.85*scale, canvas.RoundCap, canvas.RoundJoin, col, 0.48)
	}

	// 3. Trace Jobard-Lefer Bohmian Probability-Current Streamlines J = Im(Psi^* grad Psi)
	bohmianVel := func(x, y float64) render.Point {
		pR, pI, dRdx, dRdy, dIdx, dIdy := evalWaveFunction(x, y)
		jx := pR*dIdx - pI*dRdx
		jy := pR*dIdy - pI*dRdy
		return render.Pt(jx, jy)
	}

	streamlines := render.TraceEvenStreamlines(render.StreamlineConfig{
		Width:    fw,
		Height:   fh,
		DSep:     9.0 * scale,
		DTest:    4.6 * scale,
		StepSize: 2.0 * scale,
		MaxSteps: 130,
		Margin:   8.0 * scale,
	}, bohmianVel)

	for idx, line := range streamlines {
		if len(line) < 6 {
			continue
		}
		lineW := 1.15 * scale
		alpha := 0.68
		if idx%4 == 0 {
			lineW = 1.75 * scale
			alpha = 0.90
		}
		rc.DrawTaperedSpline(line, lineW, render.TaperBothEnds, func(t float64) color.RGBA {
			if idx%3 == 0 {
				return render.LerpOklab(plasmaCyan, ivoryPeak, t*0.6)
			}
			return render.LerpOklab(scarGold, ivoryPeak, math.Sin(t*math.Pi)*0.65)
		}, alpha)
	}

	rc.ApplyBloom(0.80, 6, 0.26)
	c.MergeRenderCanvas(rc)
}
