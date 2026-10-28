package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type nautilusSpiral struct {
	turns           int
	linesPerChamber int
}

// NewNautilusSpiral returns a nautilusSpiral engine that renders a 3D fluted
// metallic logarithmic nautilus spiral with radially fluted chambers and sweeping
// S-curved longitudinal copper/gold/steel-blue ribbon plumes.
func NewNautilusSpiral(turns, linesPerChamber int) *nautilusSpiral {
	if turns <= 0 {
		turns = 5
	}
	if linesPerChamber <= 0 {
		linesPerChamber = 68
	}
	return &nautilusSpiral{
		turns:           turns,
		linesPerChamber: linesPerChamber,
	}
}

// Generative renders the 3D fluted metallic nautilus spiral and sweeping ribbon plumes.
func (ns *nautilusSpiral) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	// Core 3D metallic palette (Reference Image 8)
	obsidianGroove := color.RGBA{R: 8, G: 12, B: 18, A: 255}    // #080c12
	aoBaseDark := color.RGBA{R: 9, G: 13, B: 20, A: 255}        // #090d14
	deepSteelBlue := color.RGBA{R: 36, G: 55, B: 72, A: 255}    // #243748
	midSlateSilver := color.RGBA{R: 76, G: 110, B: 136, A: 255} // #4c6e88
	coolSteelFold := color.RGBA{R: 88, G: 124, B: 152, A: 255}  // #587c98
	specIceSilver := color.RGBA{R: 158, G: 192, B: 216, A: 255} // #9ec0d8

	deepBronze := color.RGBA{R: 140, G: 74, B: 39, A: 255}     // #8c4a27
	deepCopper := color.RGBA{R: 168, G: 68, B: 32, A: 255}     // #a84420
	fieryCopper := color.RGBA{R: 184, G: 82, B: 40, A: 255}    // #b85228
	antiqueGold := color.RGBA{R: 216, G: 156, B: 74, A: 255}   // #d89c4a
	burnishedGold := color.RGBA{R: 224, G: 159, B: 70, A: 255} // #e09f46
	spineGold := color.RGBA{R: 226, G: 158, B: 68, A: 255}     // #e29e44
	specBrass := color.RGBA{R: 250, G: 224, B: 158, A: 255}    // #fae09e
	spineHighlight := color.RGBA{R: 250, G: 226, B: 156, A: 255}
	champagneWhite := color.RGBA{R: 252, G: 240, B: 194, A: 255} // #fcf0c2

	// Spiral eye at (0.675*w, 0.275*h)
	apexX := 0.675 * w
	apexY := 0.275 * h

	// Deep obsidian-slate ambient backdrop
	maxCornerDist := math.Hypot(apexX, h-apexY)
	for py := 0; py < c.Height(); py++ {
		fy := float64(py) + 0.5
		for px := 0; px < c.Width(); px++ {
			fx := float64(px) + 0.5
			d := math.Hypot(fx-apexX, fy-apexY) / math.Max(1.0, maxCornerDist)
			bg := render.LerpOklab(deepSteelBlue, obsidianGroove, common.Constrain(0.55+0.45*d, 0, 1))
			rc.BlendPixel(px, py, bg, 1.0)
		}
	}

	// Logarithmic spiral r(theta) = a * exp(b * theta)
	const b = 0.205
	a := 1.6 * scale
	totalTurns := float64(ns.turns)
	if totalTurns < 3.0 {
		totalTurns = 5.0
	}
	maxTheta := (totalTurns + 0.35) * 2.0 * math.Pi

	spiralRadius := func(theta float64) float64 {
		if theta <= 0 {
			return 0
		}
		return a * math.Exp(b*theta)
	}

	spiralPt := func(theta float64) render.Point {
		if theta <= 0 {
			return render.Pt(apexX, apexY)
		}
		r := a * math.Exp(b*theta)
		return render.Pt(apexX+r*math.Cos(theta), apexY+r*math.Sin(theta))
	}

	rc.SetBlendMode(render.BlendOver)
	lightOffset := render.Pt(-0.38*scale, -0.38*scale)

	// Helper to render a single 3-pass 3D cylindrical metallic rib along pts
	drawFlutedRib := func(pts []render.Point, bodyCol, specCol color.RGBA, widthScale, alpha float64) {
		if len(pts) < 2 {
			return
		}
		// Pass A: Dark obsidian crevice groove
		rc.StrokePolyline(pts, 2.9*scale*widthScale, obsidianGroove, 0.96*alpha)
		// Pass B: Shaded metallic body (leaves crisp dark groove on both flanks)
		rc.StrokePolyline(pts, 1.25*scale*widthScale, bodyCol, 0.96*alpha)
		// Pass C: Specular ridge highlight offset slightly toward the light
		specPts := make([]render.Point, len(pts))
		for i, p := range pts {
			specPts[i] = p.Add(lightOffset)
		}
		rc.StrokePolyline(specPts, 0.55*scale*widthScale, specCol, 0.90*alpha)
	}

	// =========================================================================
	// LAYER 1: Background & Upper/Inner Nautilus Shell Chambers (Radially Fluted)
	// =========================================================================
	// Divide each turn into 6 broad, sweeping chambers (ZERO cross-hatching!)
	const chambersPerTurn = 6
	dTheta := (2.0 * math.Pi) / float64(chambersPerTurn)
	totalChambers := int(math.Ceil(maxTheta / dTheta))

	for ch := 0; ch < totalChambers; ch++ {
		theta0 := float64(ch) * dTheta
		theta1 := theta0 + dTheta
		thetaMid := 0.5 * (theta0 + theta1)
		rOuterMid := spiralRadius(thetaMid)
		if rOuterMid < 2.2*scale {
			continue
		}

		// Number of transverse fluted wires scales with chamber arc length so wire spacing is ~3.3*scale px
		arcLen := rOuterMid * dTheta * 0.92
		numWires := int(math.Round(arcLen / (3.3 * scale)))
		maxWires := 96
		numWires = common.ConstrainInt(numWires, 6, maxWires)

		// Helper to evaluate a transverse gently-curved radial wire at normalized chamber position u in [0, 1]
		const wireSamples = 24
		evalSeptumWire := func(u float64) ([]render.Point, render.Point) {
			th := theta0 + u*(theta1-theta0)
			pts := make([]render.Point, wireSamples)
			for k := 0; k < wireSamples; k++ {
				s := float64(k) / float64(wireSamples-1)
				// Subtle radial curve matching the taut radial flutes of Reference Image 8
				lag := 0.06 * math.Sin(s*math.Pi)
				var pIn, pOut render.Point
				if th < 2.0*math.Pi {
					pOut = spiralPt(th - lag*0.5)
					pIn = render.Pt(apexX, apexY).Lerp(pOut, 0.04)
				} else {
					pIn = spiralPt(th - 2.0*math.Pi - lag*0.35)
					pOut = spiralPt(th - lag)
				}
				base := pIn.Lerp(pOut, s)
				chord := pOut.Sub(pIn)
				bow := -0.045 * math.Sin(s*math.Pi)
				pts[k] = base.Add(chord.Perp().Mul(bow))
			}
			return pts, pts[wireSamples/2]
		}

		for wIdx := 0; wIdx < numWires; wIdx++ {
			u := (float64(wIdx) + 0.5) / float64(numWires)
			wirePts, midPt := evalSeptumWire(u)

			// Chamber partition seam ambient occlusion near u -> 0 and u -> 1
			seamDist := math.Min(u, 1.0-u)
			seamAO := common.Smoothstep(0.0, 0.14, seamDist)

			// Cylindrical 3D chamber dome shading along u in [0, 1]
			dome := math.Sin(u * math.Pi)
			specBell := math.Exp(-math.Pow((u-0.42)/0.26, 2.0))

			// Left side of canvas (x < 0.48*w) swells into warm antique gold, burnished bronze, and copper;
			// Upper-right, top, and inner-bottom-right of the spiral eye are cool steel-blue & slate-silver.
			warmFactor := common.Smoothstep(0.54*w, 0.32*w, midPt.X)

			coolBody := render.LerpOklab(deepSteelBlue, midSlateSilver, common.Clamp01(0.28+0.72*dome))
			coolSpec := render.LerpOklab(midSlateSilver, specIceSilver, common.Clamp01(0.35+0.65*specBell))

			warmBody := render.LerpOklab(deepBronze, antiqueGold, common.Clamp01(0.28+0.72*dome))
			warmSpec := render.LerpOklab(antiqueGold, specBrass, common.Clamp01(0.35+0.65*specBell))

			bodyCol := render.LerpOklab(coolBody, warmBody, warmFactor)
			specCol := render.LerpOklab(coolSpec, warmSpec, warmFactor)

			// Darken smoothly toward chamber partition seams
			bodyCol = render.LerpOklab(obsidianGroove, bodyCol, 0.25+0.75*seamAO)
			specCol = render.LerpOklab(deepSteelBlue, specCol, 0.20+0.80*seamAO)

			// Draw continuous full-length radial wire (no 3-piece step band!)
			drawFlutedRib(wirePts, bodyCol, specCol, 0.88, 0.96)
		}

		// Subtle dark recessed crease at u = 0 separating chambers (no bright white tentacle)
		seamPts, _ := evalSeptumWire(0.0)
		rc.StrokePolyline(seamPts, 2.6*scale, obsidianGroove, 0.88)
	}

	// =========================================================================
	// LAYER 2: Foreground Sweeping S-Curved Metallic Ribbon Plumes
	// =========================================================================
	// Helper to render a sweeping longitudinal multi-wire ribbon plume between railA and railB
	// using continuous full-length polylines (zero chunk joint ticks!).
	drawRibbonPlume := func(
		ctrlA, ctrlB [][2]float64,
		numWires int,
		colorFn func(s, t float64, pt render.Point) (bodyCol, specCol color.RGBA),
	) {
		ptsA := make([]render.Point, len(ctrlA))
		for i, p := range ctrlA {
			ptsA[i] = render.Pt(p[0]*w, p[1]*h)
		}
		ptsB := make([]render.Point, len(ctrlB))
		for i, p := range ctrlB {
			ptsB[i] = render.Pt(p[0]*w, p[1]*h)
		}

		railA := render.CatmullRomSpline(ptsA, 32)
		railB := render.CatmullRomSpline(ptsB, 32)
		m := len(railA)
		if len(railB) < m {
			m = len(railB)
		}
		if m < 2 {
			return
		}

		// Build closed polygon footprint of the ribbon plume
		poly := make([]render.Point, 0, 2*m)
		for i := 0; i < m; i++ {
			poly = append(poly, railA[i])
		}
		for i := m - 1; i >= 0; i-- {
			poly = append(poly, railB[i])
		}

		// Soft exterior drop shadow onto adjacent radial chambers
		for pass := 5; pass >= 1; pass-- {
			spread := float64(pass) * 3.5 * scale
			rc.StrokePolyline(railA[:m], spread*2.0, obsidianGroove, 0.25)
			rc.StrokePolyline(railB[:m], spread*2.0, obsidianGroove, 0.25)
		}

		// Fill polygon footprint with dark ambient-occlusion base (#090d14)
		// to cleanly occlude underlying radial chambers (zero grid cross-hatching!)
		rc.FillPolygon(poly, aoBaseDark, 0.97)

		// Draw continuous full-length parallel longitudinal 3D fluted wires across the plume
		for wIdx := 0; wIdx < numWires; wIdx++ {
			s := float64(wIdx) / float64(common.MaxInt(1, numWires-1))
			wire := make([]render.Point, m)
			for i := 0; i < m; i++ {
				t := float64(i) / float64(m-1)
				base := railA[i].Lerp(railB[i], s)
				span := railB[i].Sub(railA[i])
				perp := span.Perp().Norm()
				bulge := math.Sin(s*math.Pi) * math.Sin(t*math.Pi*2.0) * 2.5 * scale
				wire[i] = base.Add(perp.Mul(bulge))
			}

			bodyCol, specCol := colorFn(s, 0.5, wire[m/2])
			drawFlutedRib(wire, bodyCol, specCol, 0.92, 0.98)
		}

		// Crisp dark & bronze bevel framing rails on both borders of the ribbon plume
		rc.StrokePolyline(railA[:m], 3.8*scale, obsidianGroove, 0.92)
		rc.StrokePolyline(railB[:m], 3.8*scale, obsidianGroove, 0.92)
		rc.StrokePolyline(railA[:m], 1.5*scale, fieryCopper, 0.85)
		rc.StrokePolyline(railB[:m], 1.4*scale, antiqueGold, 0.82)
	}

	// Left-Flank Longitudinal Copper-Gold Spiral Sweep (covering the left half of Reference Image 8)
	drawRibbonPlume(
		[][2]float64{
			{0.46, 0.00},
			{0.28, 0.14},
			{0.18, 0.36},
			{0.22, 0.58},
			{0.36, 0.72},
		},
		[][2]float64{
			{0.12, -0.02},
			{0.03, 0.22},
			{0.02, 0.50},
			{0.08, 0.74},
			{0.20, 0.86},
		},
		38,
		func(s, t float64, pt render.Point) (color.RGBA, color.RGBA) {
			edgeAO := common.Smoothstep(0.0, 0.14, math.Min(s, 1.0-s))
			warmDome := math.Sin(s * math.Pi)
			coolInner := common.Smoothstep(0.45, 0.0, s)
			base := render.LerpOklab(fieryCopper, burnishedGold, warmDome)
			base = render.LerpOklab(base, coolSteelFold, coolInner*0.55)
			body := render.LerpOklab(obsidianGroove, base, 0.28+0.72*edgeAO)
			spec := render.LerpOklab(specBrass, champagneWhite, warmDome*0.7)
			if coolInner > 0.5 {
				spec = render.LerpOklab(spec, specIceSilver, (coolInner-0.5)*1.6)
			}
			return body, spec
		},
	)

	// Secondary Plume A (Upper-Right Sweeping Fan from (0.84*w, 0.25*h) to (0.99*w, 0.02*h))
	drawRibbonPlume(
		[][2]float64{
			{0.81, 0.26},
			{0.85, 0.18},
			{0.90, 0.09},
			{0.96, -0.02},
		},
		[][2]float64{
			{0.88, 0.29},
			{0.94, 0.21},
			{0.99, 0.11},
			{1.04, 0.02},
		},
		28,
		func(s, t float64, pt render.Point) (color.RGBA, color.RGBA) {
			edgeAO := common.Smoothstep(0.0, 0.16, math.Min(s, 1.0-s))
			ridge := math.Exp(-math.Pow((s-0.42)/0.26, 2.0))
			warmMix := common.Smoothstep(0.15, 0.85, 1.0-s)
			base := render.LerpOklab(coolSteelFold, fieryCopper, warmMix*0.75)
			base = render.LerpOklab(base, burnishedGold, ridge*0.65)
			body := render.LerpOklab(obsidianGroove, base, 0.30+0.70*edgeAO)
			spec := render.LerpOklab(specIceSilver, champagneWhite, warmMix)
			spec = render.LerpOklab(body, spec, ridge*(0.45+0.55*edgeAO))
			return body, spec
		},
	)

	// Secondary Plume B (Lower-Right & Bottom Interlocking Fold from (0.10*w, 0.82*h) to (0.98*w, 0.88*h))
	drawRibbonPlume(
		[][2]float64{
			{0.02, 0.80},
			{0.26, 0.79},
			{0.54, 0.82},
			{0.78, 0.80},
			{1.02, 0.74},
		},
		[][2]float64{
			{0.02, 1.03},
			{0.28, 0.99},
			{0.56, 1.02},
			{0.82, 0.98},
			{1.03, 0.94},
		},
		44,
		func(s, t float64, pt render.Point) (color.RGBA, color.RGBA) {
			edgeAO := common.Smoothstep(0.0, 0.15, math.Min(s, 1.0-s))
			// Cool steel-blue in the lower folds, warming to burnished bronze/gold on the right crest
			warmRight := common.Smoothstep(0.35, 0.85, t) * common.Smoothstep(0.8, 0.1, s)
			specCrest := math.Exp(-math.Pow((s-0.35)/0.24, 2.0)) * (0.5 + 0.5*math.Sin(t*math.Pi*1.4))
			base := render.LerpOklab(deepSteelBlue, coolSteelFold, 0.35+0.65*math.Sin(s*math.Pi))
			base = render.LerpOklab(base, burnishedGold, warmRight*0.82)
			body := render.LerpOklab(obsidianGroove, base, 0.28+0.72*edgeAO)
			spec := render.LerpOklab(specIceSilver, specBrass, warmRight)
			spec = render.LerpOklab(body, spec, 0.35+0.65*specCrest)
			return body, spec
		},
	)

	// Monumental Primary Sweeping S-Curved Ribbon Plume (76 parallel 3D copper/gold/champagne-brass wires):
	// Erupts tangentially from the right side of the spiral around (0.83*w, 0.26*h),
	// sweeps down the right flank (0.92*w, 0.65*h), curves across the entire lower half
	// ((0.55*w, 0.72*h) -> (0.14*w, 0.78*h)), and coils into the bottom-left (0.16*w, 0.96*h).
	drawRibbonPlume(
		// Inner/upper rail of the monumental S-plume
		[][2]float64{
			{0.77, 0.23},
			{0.84, 0.36},
			{0.82, 0.53},
			{0.66, 0.63},
			{0.42, 0.65},
			{0.16, 0.67},
			{0.02, 0.76},
			{0.05, 0.96},
		},
		// Outer/lower rail of the monumental S-plume
		[][2]float64{
			{0.85, 0.26},
			{0.97, 0.43},
			{0.98, 0.68},
			{0.84, 0.83},
			{0.56, 0.84},
			{0.28, 0.83},
			{0.15, 0.89},
			{0.24, 1.02},
		},
		46,
		func(s, t float64, pt render.Point) (color.RGBA, color.RGBA) {
			edgeAO := common.Smoothstep(0.0, 0.14, math.Min(s, 1.0-s))

			// Multi-strand copper / gold / steel-blue / champagne variation across s in [0, 1]
			strandWave := 0.5 + 0.5*math.Sin(s*math.Pi*14.0+1.2*math.Sin(s*math.Pi*4.0))
			coolBand := common.Smoothstep(0.60, 0.92, s) * (0.45 + 0.55*strandWave)
			champagneCrest := math.Exp(-math.Pow((s-0.38)/0.22, 2.0)) * (0.6 + 0.4*strandWave)

			base := render.LerpOklab(fieryCopper, burnishedGold, strandWave)
			base = render.LerpOklab(base, coolSteelFold, coolBand*0.78)
			base = render.LerpOklab(base, champagneWhite, champagneCrest*0.58)

			body := render.LerpOklab(obsidianGroove, base, 0.28+0.72*edgeAO)
			spec := render.LerpOklab(specBrass, champagneWhite, champagneCrest)
			if coolBand > 0.35 {
				spec = render.LerpOklab(spec, specIceSilver, (coolBand-0.35)/0.65)
			}
			return body, spec
		},
	)

	// =========================================================================
	// LAYER 3: Continuous Multi-Strand Spiral Spine (Zero Beading!)
	// =========================================================================
	// Sample the main logarithmic spiral from theta = 0 to theta = 5.2 * 2*pi at 600 points
	// and stroke continuous full-length polylines with rc.StrokePolyline.
	const spineSamples = 600
	spineMaxTheta := 5.2 * 2.0 * math.Pi
	spineOuter := make([]render.Point, spineSamples)
	spineCenter := make([]render.Point, spineSamples)
	spineInner := make([]render.Point, spineSamples)
	spineCrest := make([]render.Point, spineSamples)

	for i := 0; i < spineSamples; i++ {
		u := float64(i) / float64(spineSamples-1)
		th := u * spineMaxTheta
		p := spiralPt(th)
		r := spiralRadius(th)

		// Outward radial normal along the logarithmic spiral
		radial := render.Pt(math.Cos(th), math.Sin(th))
		offsetMag := common.Constrain(r*0.018, 0.25*scale, 4.2*scale)

		spineOuter[i] = p.Add(radial.Mul(offsetMag))
		spineCenter[i] = p
		spineInner[i] = p.Sub(radial.Mul(offsetMag * 0.85))
		spineCrest[i] = p.Add(lightOffset)
	}

	// Pass 1: Wide dark ambient-occlusion shadow band (#080c12)
	rc.StrokePolyline(spineCenter, 9.5*scale, obsidianGroove, 0.90)
	rc.StrokePolyline(spineOuter, 6.0*scale, obsidianGroove, 0.88)
	rc.StrokePolyline(spineInner, 5.5*scale, obsidianGroove, 0.88)

	// Pass 2: 3 parallel tightly-bundled continuous metallic rib strands
	// Strand 1 (Outer): Deep copper (#a84420)
	rc.StrokePolyline(spineOuter, 2.6*scale, deepCopper, 0.95)
	// Strand 2 (Inner): Cool slate-bronze counter-strand
	rc.StrokePolyline(spineInner, 2.2*scale, render.LerpOklab(deepSteelBlue, deepCopper, 0.55), 0.94)
	// Strand 3 (Center main): Antique gold (#e29e44)
	rc.StrokePolyline(spineCenter, 3.0*scale, spineGold, 0.98)
	// Strand 4 (Lit crest): Bright brass highlight (#fae29c)
	rc.StrokePolyline(spineCrest, 1.35*scale, spineHighlight, 0.95)

	// Pinpoint metallic eye at (apexX, apexY)
	rc.DrawGlowDot(apexX, apexY, 5.5*scale, obsidianGroove, 0.95, 0.65)
	rc.DrawGlowDot(apexX, apexY, 2.6*scale, spineGold, 0.98, 0.55)
	rc.DrawGlowDot(apexX-0.5*scale, apexY-0.5*scale, 1.3*scale, champagneWhite, 0.98, 0.75)

	rc.ApplyVignette(0.18)
	c.MergeRenderCanvasToneMapped(rc, 1.06, 0)
}
