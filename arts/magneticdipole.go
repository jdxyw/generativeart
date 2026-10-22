package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type magneticDipole struct {
	polesN       int
	linesPerPole int
}

// NewMagneticDipole returns a magneticDipole engine that renders an orthogonal
// electromagnetic field chart combining Marching Squares equipotential rings,
// Jobard-Lefer / RK4 dipole field lines, and 3D beveled plasma pole cores.
func NewMagneticDipole(polesN, linesPerPole int) *magneticDipole {
	if polesN <= 0 {
		polesN = 6
	}
	if linesPerPole <= 0 {
		linesPerPole = 28
	}
	return &magneticDipole{
		polesN:       polesN,
		linesPerPole: linesPerPole,
	}
}

// Generative renders the multi-pole electromagnetic field lines and orthogonal equipotential contours.
func (md *magneticDipole) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0
	cx, cy := 0.5*w, 0.5*h

	// Cosmic Electromagnetic & Technical Astrolabe Palette
	abyssSpace := color.RGBA{R: 8, G: 11, B: 22, A: 255}
	deepCobalt := color.RGBA{R: 18, G: 44, B: 96, A: 255}
	electricCyan := color.RGBA{R: 48, G: 204, B: 238, A: 255}
	paleIce := color.RGBA{R: 214, G: 248, B: 255, A: 255}

	crimsonMagma := color.RGBA{R: 196, G: 38, B: 58, A: 255}
	solarOrange := color.RGBA{R: 246, G: 128, B: 42, A: 255}
	plasmaGold := color.RGBA{R: 255, G: 218, B: 88, A: 255}

	rc.FillBackground(abyssSpace)
	rc.SetBlendMode(render.BlendOver)

	type magPole struct {
		pos    render.Point
		charge float64
		radius float64
	}

	numPoles := md.polesN
	if numPoles < 2 {
		numPoles = 6
	}
	poles := make([]magPole, 0, numPoles)
	// Place alternating + / - poles in a balanced asymmetric constellation
	ringR := 0.255 * math.Min(w, h)
	for i := 0; i < numPoles; i++ {
		ang := (float64(i)/float64(numPoles))*2.0*math.Pi - 0.38
		rMod := ringR * (0.88 + 0.18*math.Sin(float64(i)*1.7))
		charge := 1.0
		if i%2 == 1 {
			charge = -1.0
		}
		poles = append(poles, magPole{
			pos:    render.Pt(cx+rMod*math.Cos(ang), cy+rMod*math.Sin(ang)),
			charge: charge,
			radius: 21.0 * scale,
		})
	}

	softEpsSq := (16.0 * scale) * (16.0 * scale)
	// Scalar electrostatic/magnetic potential V(x, y) in approximately [-2.5, +2.5]
	potentialFn := func(x, y float64) float64 {
		var v float64
		for _, p := range poles {
			dx := x - p.pos.X
			dy := y - p.pos.Y
			r := math.Sqrt(dx*dx + dy*dy + softEpsSq)
			v += p.charge * (88.0 * scale) / r
		}
		return v
	}

	// Electric/Magnetic vector field E = -grad(V)
	fieldFn := func(x, y float64) render.Point {
		var ex, ey float64
		for _, p := range poles {
			dx := x - p.pos.X
			dy := y - p.pos.Y
			r2 := dx*dx + dy*dy + softEpsSq
			r3 := r2 * math.Sqrt(r2)
			factor := p.charge / r3
			ex += dx * factor
			ey += dy * factor
		}
		return render.Pt(ex, ey)
	}

	// 1. Render Soft Potential Glow Halos & Marching Squares Equipotential Contours
	isoValues := []float64{
		-1.85, -1.35, -0.95, -0.64, -0.40, -0.22, -0.09,
		0.09, 0.22, 0.40, 0.64, 0.95, 1.35, 1.85,
	}

	for _, iso := range isoValues {
		var bandCol color.RGBA
		if iso < 0 {
			u := common.Constrain((-iso)/1.85, 0, 1)
			bandCol = render.LerpOklab(deepCobalt, electricCyan, u)
		} else {
			u := common.Constrain(iso/1.85, 0, 1)
			bandCol = render.LerpOklab(crimsonMagma, plasmaGold, u)
		}

		// Translucent potential band glow
		rc.RenderIsoBand(iso-0.035, iso+0.035, 0.025, bandCol, 0.22, potentialFn)

		// Extract exact sub-pixel Marching Squares equipotential curves
		contours := render.ExtractMarchingSquares(96, 96, w, h, iso, potentialFn)
		eqPath := &canvas.Path{}
		for _, poly := range contours {
			if len(poly) < 3 {
				continue
			}
			smooth := render.CatmullRomSpline(poly, 4)
			eqPath.MoveTo(smooth[0].X, smooth[0].Y)
			for k := 1; k < len(smooth); k++ {
				eqPath.LineTo(smooth[k].X, smooth[k].Y)
			}
		}
		rc.StrokePath(eqPath, 1.05*scale, canvas.RoundCap, canvas.RoundJoin, bandCol, 0.68)
	}

	// 2. Build Combined Pole SDF to Clip Streamlines Cleanly Outside Pole Cores
	var polesSDF render.SDF2D
	for i, p := range poles {
		s := render.SDFCircle(p.pos, p.radius*0.92)
		if i == 0 {
			polesSDF = s
		} else {
			polesSDF = render.SDFUnion(polesSDF, s)
		}
	}

	rc.PushClipSDF(polesSDF, true)

	// 3. Trace Orthogonal Field Lines Using Jobard-Lefer Evenly-Spaced Streamlines + Pole-Seeded Rays
	cfg := render.DefaultStreamlineConfig(w, h, 8.2*scale)
	cfg.DTest = 4.4 * scale
	cfg.StepSize = 2.2 * scale
	cfg.MaxSteps = 220
	streamlines := render.TraceEvenStreamlines(cfg, fieldFn)

	grad := render.NewMultiStopGradient(
		paleIce,
		electricCyan,
		deepCobalt,
		crimsonMagma,
		solarOrange,
		plasmaGold,
	)

	for idx, line := range streamlines {
		if len(line) < 4 {
			continue
		}
		strokeW := 2.1 * scale
		if idx%4 == 0 {
			strokeW = 2.85 * scale
		}

		rc.StrokePolyline(line, strokeW*1.3, abyssSpace, 0.85)
		rc.DrawTaperedSpline(
			line,
			strokeW,
			render.TaperBothEnds,
			func(t float64) color.RGBA {
				pt := line[int(common.Constrain(t*float64(len(line)-1), 0, float64(len(line)-1)))]
				v := potentialFn(pt.X, pt.Y)
				u := common.Constrain(0.5+0.42*math.Tanh(v*1.1), 0, 1)
				return grad.SampleOklab(u)
			},
			0.95,
		)
	}

	rc.PopClip()

	// 4. Render 3D Beveled Pole Cores (+ Solar Gold/Coral, - Electric Cyan/Cobalt)
	for _, p := range poles {
		pSDF := render.SDFCircle(p.pos, p.radius)
		rc.RenderSDFDropShadow(pSDF, render.Pt(4*scale, 5*scale), 12*scale, 2*scale, abyssSpace, 0.88)

		if p.charge > 0 {
			rc.RenderSDFBevel(pSDF, solarOrange, crimsonMagma, plasmaGold, p.radius*0.45, render.Pt(-0.7, -0.7), 1.0)
			rc.DrawSoftRing(p.pos.X, p.pos.Y, p.radius, 2.0*scale, plasmaGold, 0.95)
			// Plus (+) symbol in core
			sym := &canvas.Path{}
			rSym := p.radius * 0.42
			sym.MoveTo(p.pos.X-rSym, p.pos.Y)
			sym.LineTo(p.pos.X+rSym, p.pos.Y)
			sym.MoveTo(p.pos.X, p.pos.Y-rSym)
			sym.LineTo(p.pos.X, p.pos.Y+rSym)
			rc.StrokePath(sym, 2.2*scale, canvas.RoundCap, canvas.RoundJoin, paleIce, 0.95)
		} else {
			rc.RenderSDFBevel(pSDF, electricCyan, deepCobalt, paleIce, p.radius*0.45, render.Pt(-0.7, -0.7), 1.0)
			rc.DrawSoftRing(p.pos.X, p.pos.Y, p.radius, 2.0*scale, paleIce, 0.95)
			// Minus (-) symbol in core
			sym := &canvas.Path{}
			rSym := p.radius * 0.42
			sym.MoveTo(p.pos.X-rSym, p.pos.Y)
			sym.LineTo(p.pos.X+rSym, p.pos.Y)
			rc.StrokePath(sym, 2.2*scale, canvas.RoundCap, canvas.RoundJoin, paleIce, 0.95)
		}
	}

	rc.ApplyBloom(0.78, 5.0*scale, 0.32)

	c.MergeRenderCanvas(rc)
}
