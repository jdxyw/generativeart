package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type evenStreamlines struct {
	dSep  float64
	dTest float64
}

// NewEvenStreamlines returns an evenStreamlines engine that renders a celestial
// astrolabe and multi-vortex dipole chart using Jobard-Lefer evenly-spaced streamlines,
// SDF clip masks, and 3D beveled planetary orbs.
func NewEvenStreamlines(dSep, dTest float64) *evenStreamlines {
	if dSep <= 1.5 {
		dSep = 7.5
	}
	if dTest <= 0.5 || dTest > dSep {
		dTest = dSep * 0.52
	}
	return &evenStreamlines{
		dSep:  dSep,
		dTest: dTest,
	}
}

// Generative renders the Jobard-Lefer evenly-spaced streamline celestial engraving.
func (es *evenStreamlines) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0

	// Celestial Astrolabe & Currency Engraving Palette
	abyssNavy := color.RGBA{R: 12, G: 16, B: 28, A: 255}
	midnightIndigo := color.RGBA{R: 20, G: 32, B: 58, A: 255}
	cobaltBlue := color.RGBA{R: 44, G: 108, B: 172, A: 255}
	ceruleanTeal := color.RGBA{R: 68, G: 168, B: 176, A: 255}
	mintIce := color.RGBA{R: 168, G: 228, B: 214, A: 255}

	champagneGold := color.RGBA{R: 248, G: 214, B: 108, A: 255}
	imperialAmber := color.RGBA{R: 236, G: 162, B: 52, A: 255}
	vermilionCoral := color.RGBA{R: 218, G: 76, B: 54, A: 255}
	antiqueBronze := color.RGBA{R: 92, G: 58, B: 26, A: 255}
	ivoryCream := color.RGBA{R: 252, G: 244, B: 222, A: 255}

	rc.FillBackground(abyssNavy)
	rc.SetBlendMode(render.BlendOver)

	// 1. Subtle Celestial Grid & Concentric Astrolabe Degree Rings in the Background
	gridPath := &canvas.Path{}
	cx, cy := 0.50*w, 0.50*h
	for _, rFrac := range []float64{0.22, 0.36, 0.46} {
		r := rFrac * w
		const steps = 96
		for i := 0; i <= steps; i++ {
			a := (float64(i) / float64(steps)) * 2.0 * math.Pi
			px, py := cx+r*math.Cos(a), cy+r*math.Sin(a)
			if i == 0 {
				gridPath.MoveTo(px, py)
			} else {
				gridPath.LineTo(px, py)
			}
		}
		gridPath.Close()
	}
	rc.StrokePath(gridPath, 0.85*scale, canvas.RoundCap, canvas.RoundJoin, midnightIndigo, 0.75)

	// 2. Define 3 Celestial Vortex/Dipole Orbs
	type celestialOrb struct {
		center     render.Point
		radius     float64
		vortexSign float64
		baseCol    color.RGBA
		rimCol     color.RGBA
		specCol    color.RGBA
	}
	orbs := []celestialOrb{
		{
			center:     render.Pt(0.32*w, 0.34*h),
			radius:     0.125 * w,
			vortexSign: 1.0,
			baseCol:    imperialAmber,
			rimCol:     antiqueBronze,
			specCol:    ivoryCream,
		},
		{
			center:     render.Pt(0.70*w, 0.67*h),
			radius:     0.105 * w,
			vortexSign: -1.0,
			baseCol:    ceruleanTeal,
			rimCol:     midnightIndigo,
			specCol:    mintIce,
		},
		{
			center:     render.Pt(0.71*w, 0.25*h),
			radius:     0.072 * w,
			vortexSign: 0.8,
			baseCol:    vermilionCoral,
			rimCol:     antiqueBronze,
			specCol:    champagneGold,
		},
	}

	// Build a combined 2D SDF for the 3 orbs with a small clearance collar
	var orbsSDF render.SDF2D
	for i, ob := range orbs {
		s := render.SDFCircle(ob.center, ob.radius+3.5*scale)
		if i == 0 {
			orbsSDF = s
		} else {
			orbsSDF = render.SDFUnion(orbsSDF, s)
		}
	}

	// 3. Multi-Vortex + Dipole + Mild Curl Velocity Field
	ne := common.NewNoiseEngine(202612)
	fieldFn := func(x, y float64) render.Point {
		var vx, vy float64
		for _, ob := range orbs {
			dx := x - ob.center.X
			dy := y - ob.center.Y
			r2 := dx*dx + dy*dy + (24.0*scale)*(24.0*scale)
			invR := 1.0 / math.Sqrt(r2)
			weight := (ob.radius * 1.8) * invR

			// Swirling tangential vortex + radial dipole deflection around the orb
			tx := -dy * invR * ob.vortexSign
			ty := dx * invR * ob.vortexSign
			rx := dx * invR * 0.28
			ry := dy * invR * 0.28

			vx += (tx + rx) * weight
			vy += (ty + ry) * weight
		}
		// Add smooth large-scale curl wave
		curl := ne.Curl2D(x/(260.0*scale), y/(260.0*scale), 1e-3)
		vx += curl.X * 0.38
		vy += curl.Y * 0.38
		return render.Pt(vx, vy)
	}

	// 4. Trace Jobard-Lefer Evenly-Spaced Streamlines & Render with Inverted SDF Clip Mask
	cfg := render.DefaultStreamlineConfig(w, h, es.dSep*scale)
	cfg.DTest = es.dTest * scale
	cfg.StepSize = math.Max(1.6, es.dSep*scale*0.26)
	cfg.MaxSteps = 210
	cfg.MinPoints = 7

	streamlines := render.TraceEvenStreamlines(cfg, fieldFn)

	// Push inverted SDF clip mask so streamlines are cleanly clipped outside the 3 orbs
	rc.PushClipSDF(orbsSDF, true)

	grad := render.NewMultiStopGradient(
		cobaltBlue,
		ceruleanTeal,
		mintIce,
		champagneGold,
		imperialAmber,
		vermilionCoral,
	)

	rng := rand.New(rand.NewSource(202612))
	for idx, line := range streamlines {
		if len(line) < 4 {
			continue
		}
		midPt := line[len(line)/2]
		// Compute distance to nearest orb to modulate color & stroke thickness
		minOrbDist := 1e9
		for _, ob := range orbs {
			d := midPt.Dist(ob.center) - ob.radius
			if d < minOrbDist {
				minOrbDist = d
			}
		}

		ang := math.Atan2(line[len(line)-1].Y-line[0].Y, line[len(line)-1].X-line[0].X)
		uCol := common.Constrain(math.Exp(-minOrbDist/(165.0*scale))*0.78+0.22*(0.5+0.5*math.Sin(ang*2.0+float64(idx)*0.17)), 0, 1)

		maxW := (2.6 + 1.4*math.Exp(-minOrbDist/(120.0*scale))) * scale
		if idx%5 == 0 {
			maxW *= 1.22
		}

		// Dark shadow under-stroke for 2.5D relief separation
		rc.StrokePolyline(line, maxW*1.25, abyssNavy, 0.85)

		// Variable-width calligraphic streamline ribbon
		rc.DrawTaperedSpline(
			line,
			maxW,
			render.TaperBothEnds,
			func(t float64) color.RGBA {
				shift := common.Constrain(uCol+(t-0.5)*0.24, 0, 1)
				return grad.SampleOklab(shift)
			},
			0.96,
		)

		// Occasional golden astrolabe bead along select streamlines
		if idx%6 == 0 && len(line) > 14 {
			beadPt := line[len(line)/2+rng.Intn(5)-2]
			rc.DrawGlowDot(beadPt.X, beadPt.Y, maxW*1.15, champagneGold, 0.95, 0.82)
		}
	}

	rc.PopClip()

	// 5. Render the 3 Celestial Orbs with SDF Soft Shadows, Beveled Spheres & Astrolabe Filigree
	for _, ob := range orbs {
		orbSDF := render.SDFCircle(ob.center, ob.radius)
		rc.RenderSDFDropShadow(
			orbSDF,
			render.Pt(6.0*scale, 8.0*scale),
			14.0*scale,
			2.0*scale,
			color.RGBA{R: 4, G: 6, B: 14, A: 255},
			0.85,
		)
		rc.RenderSDFBevel(
			orbSDF,
			ob.baseCol,
			ob.rimCol,
			ob.specCol,
			ob.radius*0.42,
			render.Pt(-0.65, -0.75),
			1.0,
		)

		// Clip inside the orb and use the 2D Transform Stack to draw rotated astrolabe rings & star points
		rc.PushClipCircle(ob.center.X, ob.center.Y, ob.radius*0.96)
		rc.PushMatrix()
		rc.Translate(ob.center.X, ob.center.Y)
		rc.Rotate(0.35 * ob.vortexSign)

		filigree := &canvas.Path{}
		for _, rf := range []float64{0.48, 0.74, 0.88} {
			r := ob.radius * rf
			for k := 0; k <= 64; k++ {
				a := (float64(k) / 64.0) * 2.0 * math.Pi
				px, py := r*math.Cos(a), r*math.Sin(a)
				if k == 0 {
					filigree.MoveTo(px, py)
				} else {
					filigree.LineTo(px, py)
				}
			}
			filigree.Close()
		}
		// 8-point compass star needles inside the orb
		const compassPts = 8
		for k := 0; k < compassPts; k++ {
			a := (float64(k) / float64(compassPts)) * 2.0 * math.Pi
			rTip := ob.radius * 0.84
			if k%2 == 1 {
				rTip = ob.radius * 0.54
			}
			filigree.MoveTo(0, 0)
			filigree.LineTo(rTip*math.Cos(a), rTip*math.Sin(a))
		}
		rc.StrokePath(filigree, 1.35*scale, canvas.RoundCap, canvas.RoundJoin, champagneGold, 0.92)
		rc.DrawGlowDot(0, 0, ob.radius*0.14, ivoryCream, 1.0, 0.85)

		rc.PopMatrix()
		rc.PopClip()

		// Outer gilded bezel ring around the orb
		rc.DrawSoftRing(ob.center.X, ob.center.Y, ob.radius, 2.2*scale, champagneGold, 0.96)
	}

	c.MergeRenderCanvas(rc)
}
