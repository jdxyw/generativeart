package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type cyclicAutomata struct {
	states int
	steps  int
}

// NewCyclicAutomata returns a cyclicAutomata engine that renders Belousov-Zhabotinsky
// chemical reaction spiral waves and excitable media phase fields with branch-cut-free
// complex phasor superposition, 3D Blinn-Phong shading, and Marching Squares contours.
func NewCyclicAutomata(states, steps int) *cyclicAutomata {
	if states <= 2 {
		states = 14
	}
	if steps <= 0 {
		steps = 120
	}
	return &cyclicAutomata{
		states: states,
		steps:  steps,
	}
}

// Generative renders the Belousov-Zhabotinsky chemical spiral wave relief.
func (ca *cyclicAutomata) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := c.Width(), c.Height()
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0

	// Belousov-Zhabotinsky Malachite, Turquoise, Coral & Gold Palette
	abyssTrough := color.RGBA{R: 10, G: 16, B: 34, A: 255}
	cobaltWave := color.RGBA{R: 26, G: 78, B: 152, A: 255}
	turquoiseCrest := color.RGBA{R: 42, G: 176, B: 178, A: 255}
	mintFoam := color.RGBA{R: 164, G: 238, B: 214, A: 255}
	amberFront := color.RGBA{R: 248, G: 186, B: 58, A: 255}
	coralPulse := color.RGBA{R: 226, G: 72, B: 54, A: 255}
	magentaCore := color.RGBA{R: 168, G: 32, B: 98, A: 255}
	ivorySpec := color.RGBA{R: 255, G: 248, B: 224, A: 255}

	rc.FillBackground(abyssTrough)
	rc.SetBlendMode(render.BlendOver)

	type spiralSource struct {
		x, y       float64
		charge     float64 // integer topological charge (+1, -1)
		freq       float64
		phaseShift float64
	}

	sources := []spiralSource{
		{0.30 * fw, 0.32 * fh, 1.0, 0.105 / scale, 0.0},
		{0.72 * fw, 0.28 * fh, -1.0, 0.110 / scale, 1.2},
		{0.68 * fw, 0.72 * fh, 1.0, 0.102 / scale, 2.4},
		{0.26 * fw, 0.70 * fh, -1.0, 0.108 / scale, 3.8},
		{0.50 * fw, 0.50 * fh, 1.0, 0.115 / scale, 0.7},
	}

	ne := common.NewNoiseEngine(202618)

	// Branch-cut-free complex Ginzburg-Landau / BZ spiral wave evaluation:
	// Because charge is an integer (+1 or -1), cos(k*r - m*theta + phi0) and sin(k*r - m*theta + phi0)
	// are 100% continuous across theta = +/- pi with zero branch-cut seams!
	evalBZField := func(x, y float64) float64 {
		wx := ne.FBM2D(x/(190.0*scale), y/(190.0*scale), 3, 2.0, 0.5) * 16.0 * scale
		wy := ne.FBM2D((x+500)/(190.0*scale), (y+500)/(190.0*scale), 3, 2.0, 0.5) * 16.0 * scale
		px := x + wx
		py := y + wy

		var sumCos, sumSin float64
		for _, src := range sources {
			dx := px - src.x
			dy := py - src.y
			r := math.Hypot(dx, dy)
			ang := math.Atan2(dy, dx)
			phase := r*src.freq - src.charge*ang + src.phaseShift
			// Gaussian territory envelope so each spiral pacemaker dominates its basin cleanly
			wgt := math.Exp(-(r * r) / ((145.0 * scale) * (145.0 * scale)))
			// Core amplitude suppression r / sqrt(r^2 + r0^2) eliminates singularity at r=0
			coreEnv := r / math.Hypot(r, 8.0*scale)
			sumCos += wgt * coreEnv * math.Cos(phase)
			sumSin += wgt * coreEnv * math.Sin(phase)
		}

		// Smoothly map the continuous complex phasor (sumCos, sumSin) to a [0, 1] wave height
		wavePhase := math.Atan2(sumSin, sumCos)
		s := math.Sin(wavePhase) + 0.25*math.Cos(2.0*wavePhase)
		return common.Constrain(0.5+0.42*s, 0, 1)
	}

	fieldGrid := make([]float64, w*h)
	for y := 0; y < h; y++ {
		fy := float64(y) + 0.5
		rowOff := y * w
		for x := 0; x < w; x++ {
			fx := float64(x) + 0.5
			fieldGrid[rowOff+x] = evalBZField(fx, fy)
		}
	}

	grad := render.NewMultiStopGradient(
		abyssTrough,
		cobaltWave,
		turquoiseCrest,
		mintFoam,
		amberFront,
		coralPulse,
		magentaCore,
		abyssTrough,
	)

	numStates := ca.states
	if numStates < 8 {
		numStates = 14
	}

	// 1. Shade the 3D Excitable Media Wave Surface with Blinn-Phong Specular Lighting
	for y := 0; y < h; y++ {
		rowOff := y * w
		yPrev := y - 1
		if yPrev < 0 {
			yPrev = 0
		}
		yNext := y + 1
		if yNext >= h {
			yNext = h - 1
		}
		for x := 0; x < w; x++ {
			xPrev := x - 1
			if xPrev < 0 {
				xPrev = 0
			}
			xNext := x + 1
			if xNext >= w {
				xNext = w - 1
			}

			u := fieldGrid[rowOff+x]
			uQuant := math.Floor(u*float64(numStates)) / float64(numStates)
			col := grad.SampleOklab(common.Constrain(0.62*u+0.38*uQuant, 0, 1))

			dzdx := (fieldGrid[rowOff+xNext] - fieldGrid[rowOff+xPrev]) * 14.0
			dzdy := (fieldGrid[yNext*w+x] - fieldGrid[yPrev*w+x]) * 14.0
			nx, ny, nz := -dzdx, -dzdy, 1.0
			invLen := 1.0 / math.Sqrt(nx*nx+ny*ny+nz*nz)
			nx *= invLen
			ny *= invLen
			nz *= invLen

			ndotl := common.Constrain(nx*(-0.55)+ny*(-0.60)+nz*0.58, 0, 1)
			spec := math.Pow(ndotl, 14.0)

			col = render.LerpOklab(abyssTrough, col, 0.34+0.66*ndotl)
			if spec > 0.08 {
				col = render.LerpOklab(col, ivorySpec, common.Constrain(spec*0.75, 0, 1))
			}

			rc.SetPixelLinear(
				x, y,
				render.SRGB8ToLinear(col.R),
				render.SRGB8ToLinear(col.G),
				render.SRGB8ToLinear(col.B),
				1.0,
			)
		}
	}

	// 2. Extract Marching Squares Iso-Phase Wavefront Rings
	gridLookup := func(x, y float64) float64 {
		ix := int(common.Constrain(math.Floor(x), 0, fw-1))
		iy := int(common.Constrain(math.Floor(y), 0, fh-1))
		return fieldGrid[iy*w+ix]
	}

	for _, iso := range []float64{0.26, 0.50, 0.74} {
		contours := render.ExtractMarchingSquares(110, 110, fw, fh, iso, gridLookup)
		wavePath := &canvas.Path{}
		for _, poly := range contours {
			if len(poly) < 3 {
				continue
			}
			smooth := render.CatmullRomSpline(poly, 3)
			wavePath.MoveTo(smooth[0].X, smooth[0].Y)
			for k := 1; k < len(smooth); k++ {
				wavePath.LineTo(smooth[k].X, smooth[k].Y)
			}
		}
		lineCol := ivorySpec
		if iso == 0.50 {
			lineCol = amberFront
		}
		rc.StrokePath(wavePath, 1.1*scale, canvas.RoundCap, canvas.RoundJoin, lineCol, 0.76)
	}

	c.MergeRenderCanvas(rc)
}
