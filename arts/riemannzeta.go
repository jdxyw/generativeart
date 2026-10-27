package arts

import (
	"image/color"
	"math"
	"math/cmplx"
	"runtime"
	"sync"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type riemannZeta struct {
	termsN    int
	phaseRays int
}

// NewRiemannZeta returns a riemannZeta engine that renders a complex phase portrait
// of the Riemann Zeta function zeta(s) across the critical strip Re(s) = 1/2 using
// Borwein's Chebyshev-accelerated Dirichlet series, 3D log-modulus contour relief,
// and Jobard-Lefer Newton-flow streamlines converging onto the non-trivial zeros.
func NewRiemannZeta(termsN, phaseRays int) *riemannZeta {
	if termsN <= 0 {
		termsN = 26
	}
	if phaseRays <= 0 {
		phaseRays = 16
	}
	return &riemannZeta{
		termsN:    termsN,
		phaseRays: phaseRays,
	}
}

// Generative renders the Riemann Zeta critical-strip complex phase portrait and Newton flow.
func (rz *riemannZeta) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := c.Width(), c.Height()
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0

	// Cyclic Oklab Phase Wheel Palette (zero branch-cut seam across +/- pi)
	abyssZero := color.RGBA{R: 6, G: 9, B: 18, A: 255}
	cyclicWheel := []color.RGBA{
		{R: 22, G: 62, B: 152, A: 255},  // 0 deg: Royal Lapis
		{R: 42, G: 184, B: 192, A: 255}, // 60 deg: Jingtai Turquoise
		{R: 248, G: 194, B: 58, A: 255}, // 120 deg: Imperial Gold
		{R: 224, G: 62, B: 46, A: 255},  // 180 deg: Cinnabar Vermilion
		{R: 168, G: 32, B: 108, A: 255}, // 240 deg: Byzantium Magenta
		{R: 64, G: 28, B: 126, A: 255}, // 300 deg: Deep Indigo Violet
	}
	ivorySpec := color.RGBA{R: 255, G: 252, B: 234, A: 255}
	goldLine := color.RGBA{R: 255, G: 224, B: 118, A: 255}
	cyanLine := color.RGBA{R: 96, G: 238, B: 252, A: 255}

	rc.FillBackground(abyssZero)
	rc.SetBlendMode(render.BlendOver)

	// Precompute Borwein's Chebyshev-type alternating series weights for Dirichlet Eta(s)
	nTerms := rz.termsN
	if nTerms < 18 || nTerms > 32 {
		nTerms = 26
	}
	dWeights := make([]float64, nTerms+1)
	curTerm := float64(nTerms)
	sumD := 0.0
	for j := 0; j <= nTerms; j++ {
		if j > 0 {
			num := float64(nTerms+j-1) * float64(nTerms-j+1) * 4.0
			den := float64(2*j-1) * float64(2*j)
			curTerm *= num / den
		}
		sumD += curTerm
		dWeights[j] = sumD
	}
	dN := dWeights[nTerms]
	borweinCoeff := make([]float64, nTerms)
	logK := make([]float64, nTerms)
	for k := 0; k < nTerms; k++ {
		sign := 1.0
		if k%2 == 1 {
			sign = -1.0
		}
		borweinCoeff[k] = -sign * (dWeights[k] - dN) / dN
		logK[k] = math.Log(float64(k + 1))
	}
	ln2 := math.Ln2

	// Map canvas pixel (px, py) to complex plane s = sigma + i*t:
	// Orient the critical line sigma = 0.5 along a graceful diagonal from lower-left (t=11.5) to upper-right (t=32.0),
	// spanning the first 4 non-trivial Riemann zeros (14.135, 21.022, 25.011, 30.425)!
	cosDiag, sinDiag := math.Cos(-0.68), math.Sin(-0.68)
	pixelToComplex := func(px, py float64) complex128 {
		ux := (px/fw - 0.5) * 2.0
		uy := (0.5 - py/fh) * 2.0
		// Rotate so the diagonal is aligned with the imaginary axis t
		rx := ux*cosDiag - uy*sinDiag
		ry := ux*sinDiag + uy*cosDiag
		sigma := 0.5 + rx*5.4
		t := 22.2 + ry*10.8
		return complex(sigma, t)
	}

	// Fast Borwein evaluation of zeta(s) = eta(s) / (1 - 2^(1-s))
	evalZeta := func(s complex128) complex128 {
		sig, t := real(s), imag(s)
		var etaR, etaI float64
		for k := 0; k < nTerms; k++ {
			lk := logK[k]
			mag := borweinCoeff[k] * math.Exp(-sig*lk)
			ang := -t * lk
			etaR += mag * math.Cos(ang)
			etaI += mag * math.Sin(ang)
		}
		// Divide by (1 - 2^(1-s))
		mag2 := math.Exp((1.0 - sig) * ln2)
		ang2 := -t * ln2
		denom := complex(1.0-mag2*math.Cos(ang2), -mag2*math.Sin(ang2))
		if cmplx.Abs(denom) < 1e-12 {
			return complex(1e6, 0)
		}
		return complex(etaR, etaI) / denom
	}

	sampleCyclicPhase := func(phase float64) color.RGBA {
		u := math.Mod(phase/(2.0*math.Pi), 1.0)
		if u < 0 {
			u += 1.0
		}
		pos := u * float64(len(cyclicWheel))
		i0 := int(math.Floor(pos)) % len(cyclicWheel)
		i1 := (i0 + 1) % len(cyclicWheel)
		return render.LerpOklab(cyclicWheel[i0], cyclicWheel[i1], pos-math.Floor(pos))
	}

	numRays := float64(rz.phaseRays)
	if numRays < 8 {
		numRays = 16
	}

	// 1. Parallel 2x2 SSAA Evaluation of Wegert Complex Phase Portrait & 3D Log-Modulus Relief
	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
	reGrid := make([]float64, w*h)
	imGrid := make([]float64, w*h)

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
					s0 := pixelToComplex(fx, fy)
					z0 := evalZeta(s0)
					zX := evalZeta(pixelToComplex(fx+1.2, fy))
					zY := evalZeta(pixelToComplex(fx, fy+1.2))

					reGrid[rowOff+x] = real(z0)
					imGrid[rowOff+x] = imag(z0)

					mag0 := cmplx.Abs(z0)
					phase0 := cmplx.Phase(z0)

					// Logarithmic modulus rings + iso-phase rays height function for 3D silk relief
					heightFn := func(z complex128) float64 {
						m := math.Log2(cmplx.Abs(z) + 1e-6)
						ph := cmplx.Phase(z)
						return 0.65*math.Sin(m*2.0*math.Pi) + 0.35*math.Cos(ph*numRays)
					}

					h0 := heightFn(z0)
					hX := heightFn(zX)
					hY := heightFn(zY)
					gx := hX - h0
					gy := hY - h0

					nx, ny, nz := -gx*1.8, -gy*1.8, 1.0
					nLen := math.Sqrt(nx*nx + ny*ny + nz*nz)
					nx, ny, nz = nx/nLen, ny/nLen, nz/nLen

					ndotl := common.Constrain(-0.58*nx-0.64*ny+0.50*nz, 0, 1)
					spec := math.Pow(common.Constrain(-0.42*nx-0.48*ny+0.77*nz, 0, 1), 22.0)

					baseCol := sampleCyclicPhase(phase0)
					// Sawtooth log-modulus contour bands
					logM := math.Log2(mag0 + 1e-6)
					modSaw := logM - math.Floor(logM)
					contourShade := 0.72 + 0.28*modSaw

					// Darken smoothly right at the non-trivial zeros (|zeta(s)| -> 0) to highlight the vortex eye
					zeroCore := 1.0 - math.Exp(-mag0*6.5)
					illum := common.Constrain((0.22+0.78*ndotl)*contourShade*zeroCore, 0.04, 1.0)

					col := render.LerpOklab(abyssZero, baseCol, illum)
					if spec > 0.04 {
						col = render.LerpOklab(col, ivorySpec, common.Constrain(spec*0.80*zeroCore, 0, 1))
					}

					rc.SetPixelLinear(x, y, render.SRGB8ToLinear(col.R), render.SRGB8ToLinear(col.G), render.SRGB8ToLinear(col.B), 1.0)
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	// 2. Extract Real & Imaginary Zero-Locus Contours (Re zeta(s) == 0 in Gold, Im zeta(s) == 0 in Cyan)
	// Their orthogonal intersections on the critical line Re(s) = 1/2 are the exact Non-Trivial Riemann Zeros!
	reLookup := func(x, y float64) float64 {
		ix := common.ConstrainInt(int(x), 0, w-1)
		iy := common.ConstrainInt(int(y), 0, h-1)
		return math.Atan(reGrid[iy*w+ix])
	}
	imLookup := func(x, y float64) float64 {
		ix := common.ConstrainInt(int(x), 0, w-1)
		iy := common.ConstrainInt(int(y), 0, h-1)
		return math.Atan(imGrid[iy*w+ix])
	}

	reContours := render.ExtractMarchingSquares(120, 120, fw, fh, 0.0, reLookup)
	imContours := render.ExtractMarchingSquares(120, 120, fw, fh, 0.0, imLookup)

	strokeContourSet := func(polys [][]render.Point, col color.RGBA, width, alpha float64) {
		p := &canvas.Path{}
		for _, poly := range polys {
			if len(poly) < 3 {
				continue
			}
			smooth := render.CatmullRomSpline(poly, 3)
			p.MoveTo(smooth[0].X, smooth[0].Y)
			for i := 1; i < len(smooth); i++ {
				p.LineTo(smooth[i].X, smooth[i].Y)
			}
		}
		rc.StrokePath(p, width, canvas.RoundCap, canvas.RoundJoin, col, alpha)
	}
	strokeContourSet(imContours, cyanLine, 1.35*scale, 0.68)
	strokeContourSet(reContours, goldLine, 1.55*scale, 0.82)

	// 3. Trace Jobard-Lefer Evenly-Spaced Newton-Flow Streamlines (-zeta(s) / zeta'(s))
	newtonVel := func(x, y float64) render.Point {
		s0 := pixelToComplex(x, y)
		z0 := evalZeta(s0)
		zX := evalZeta(pixelToComplex(x+1.5, y))
		zY := evalZeta(pixelToComplex(x, y+1.5))
		// Directional gradient of log|zeta(s)| in screen coordinates
		m0 := math.Log(cmplx.Abs(z0) + 1e-5)
		mX := math.Log(cmplx.Abs(zX) + 1e-5)
		mY := math.Log(cmplx.Abs(zY) + 1e-5)
		return render.Pt(-(mX - m0), -(mY - m0))
	}

	streamlines := render.TraceEvenStreamlines(render.StreamlineConfig{
		Width:    fw,
		Height:   fh,
		DSep:     10.5 * scale,
		DTest:    5.2 * scale,
		StepSize: 2.0 * scale,
		MaxSteps: 110,
		Margin:   10.0 * scale,
	}, newtonVel)

	for idx, line := range streamlines {
		if len(line) < 6 {
			continue
		}
		rc.DrawTaperedSpline(line, 1.05*scale, render.TaperBothEnds, func(t float64) color.RGBA {
			if idx%3 == 0 {
				return ivorySpec
			}
			return goldLine
		}, 0.55)
	}

	// 4. Highlight the Non-Trivial Zeros on the Critical Line Re(s) = 1/2
	zeroGammas := []float64{14.134725, 21.022040, 25.010858, 30.424876}
	for _, gamma := range zeroGammas {
		ry := (gamma - 22.2) / 10.8
		// Inverse rotation from (rx=0, ry) back to screen (px, py)
		ux := ry * sinDiag
		uy := ry * cosDiag
		zx := (ux*0.5 + 0.5) * fw
		zy := (0.5 - uy*0.5) * fh
		if zx > 15 && zx < fw-15 && zy > 15 && zy < fh-15 {
			rc.DrawSoftRing(zx, zy, 9.5*scale, 1.4*scale, goldLine, 0.92)
			rc.DrawGlowDot(zx, zy, 3.8*scale, ivorySpec, 0.98, 0.88)
		}
	}

	rc.ApplyBloom(0.80, 6, 0.25)
	c.MergeRenderCanvas(rc)
}
