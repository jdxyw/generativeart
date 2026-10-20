package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type nautilusSpiral struct {
	turns           int
	linesPerChamber int
}

// NewNautilusSpiral returns a nautilusSpiral engine that renders a metallic
// striated logarithmic nautilus spiral engraving with 3D cylindrical relief shading.
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

// Generative renders the metallic brushed bronze/gold and slate-blue logarithmic spiral engraving.
func (ns *nautilusSpiral) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	// Metallic engraving palette
	antiqueGold := color.RGBA{R: 218, G: 168, B: 92, A: 255}
	brightBrass := color.RGBA{R: 246, G: 214, B: 146, A: 255}
	copperRust := color.RGBA{R: 196, G: 102, B: 58, A: 255}
	slateBlue := color.RGBA{R: 72, G: 104, B: 128, A: 255}
	deepSlate := color.RGBA{R: 42, G: 62, B: 82, A: 255}
	creviceDark := color.RGBA{R: 16, G: 22, B: 30, A: 255}

	// Dark slate-bronze radial backdrop centered at the spiral apex
	apexX := 0.67 * w
	apexY := 0.29 * h
	maxCornerDist := math.Hypot(apexX, h-apexY)
	for py := 0; py < c.Height(); py++ {
		fy := float64(py) + 0.5
		for px := 0; px < c.Width(); px++ {
			fx := float64(px) + 0.5
			d := math.Hypot(fx-apexX, fy-apexY) / math.Max(1.0, maxCornerDist)
			ang := math.Atan2(fy-apexY, fx-apexX)
			swirl := 0.5 + 0.5*math.Sin(ang*2.0-math.Log(math.Max(2.0, d*300.0))*3.2)
			bgBase := render.LerpOklab(deepSlate, creviceDark, common.Constrain(d*0.85, 0, 1))
			bgWarm := render.LerpOklab(bgBase, copperRust, 0.16*swirl*(1.0-0.5*d))
			rc.BlendPixel(px, py, bgWarm, 1.0)
		}
	}

	// Logarithmic spiral parameters: r(theta) = a * exp(b * theta)
	// With b = 0.195 and 5 turns (10*pi radians), exp(0.195 * 10 * pi) ~= 457.8,
	// expanding from r = 2px at the inner eye out to ~915px at the far corners.
	const b = 0.195
	a := 2.0 * scale
	totalTurns := float64(ns.turns)
	if totalTurns < 3.0 {
		totalTurns = 5.0
	}
	// Extend extra 0.65 turn so outer chambers completely cover the far canvas corners
	maxTheta := (totalTurns + 0.65) * 2.0 * math.Pi

	spiralPt := func(theta float64) render.Point {
		if theta < 0 {
			return render.Pt(apexX, apexY)
		}
		r := a * math.Exp(b*theta)
		return render.Pt(apexX+r*math.Cos(theta), apexY+r*math.Sin(theta))
	}

	spiralRadius := func(theta float64) float64 {
		if theta < 0 {
			return 0
		}
		return a * math.Exp(b*theta)
	}

	// Build 3D cylindrical relief color function for a chamber
	makeChamberColorFn := func(baseMetal, secondaryMetal, coolTint color.RGBA, phase float64) func(s, t float64) color.RGBA {
		return func(s, t float64) color.RGBA {
			s = common.Constrain(s, 0, 1)
			t = common.Constrain(t, 0, 1)

			// Base chromatic shift along the chamber length t and width s
			along := 0.5 + 0.5*math.Sin(t*math.Pi*1.3+phase)
			body := render.LerpOklab(baseMetal, secondaryMetal, along*0.68)
			coolMix := 0.5 + 0.5*math.Cos((s*1.4-t*0.8)*math.Pi+phase*0.7)
			body = render.LerpOklab(body, coolTint, coolMix*0.42)

			// 3D cylindrical specular ridge around s = 0.35 + 0.15*sin(t*pi)
			specCenter := 0.35 + 0.15*math.Sin(t*math.Pi)
			distSpec := math.Abs(s - specCenter)
			spec := math.Exp(-distSpec * distSpec / (2.0 * 0.016))
			body = render.LerpOklab(body, brightBrass, spec*0.76)

			// Secondary cool metallic sheen on the opposite flank
			counterCenter := 0.76 - 0.10*math.Sin(t*math.Pi)
			distCounter := math.Abs(s - counterCenter)
			counterSheen := math.Exp(-distCounter * distCounter / (2.0 * 0.022))
			body = render.LerpOklab(body, slateBlue, counterSheen*0.45)

			// Deep ambient-occlusion crevice shadows near s = 0 and s = 1
			edgeDist := math.Min(s, 1.0-s)
			if edgeDist < 0.18 {
				ao := 1.0 - (edgeDist / 0.18)
				aoSmooth := ao * ao * (3.0 - 2.0*ao)
				body = render.LerpOklab(body, creviceDark, aoSmooth*0.84)
			}

			// Subtle septal crease darkening near t = 0 and t = 1
			tEdge := math.Min(t, 1.0-t)
			if tEdge < 0.12 {
				tAO := 1.0 - (tEdge / 0.12)
				body = render.LerpOklab(body, deepSlate, tAO*tAO*0.52)
			}

			return body
		}
	}

	rc.SetBlendMode(render.BlendOver)

	// 1. Primary Nautilus Chambers bounded by consecutive spiral turns r(theta) and r(theta + 2*pi)
	// Subdivide each turn into golden-proportioned angular sectors with curved septa ribs.
	const sectorsPerTurn = 18
	dTheta := (2.0 * math.Pi) / float64(sectorsPerTurn)
	totalSectors := int(math.Ceil(maxTheta / dTheta))

	for sec := 0; sec < totalSectors; sec++ {
		theta0 := float64(sec) * dTheta
		theta1 := theta0 + dTheta

		// Sample inner curve r(theta) and outer curve r(theta + 2*pi)
		// For the innermost turn, curveA starts at the apex
		const curveSamples = 20
		curveInner := make([]render.Point, curveSamples)
		curveOuter := make([]render.Point, curveSamples)

		for k := 0; k < curveSamples; k++ {
			u := float64(k) / float64(curveSamples-1)
			th := theta0 + u*(theta1-theta0)
			if th < 2.0*math.Pi {
				// Innermost eye: interpolate from apex to first turn
				pOuter := spiralPt(th)
				curveInner[k] = render.Pt(apexX, apexY).Lerp(pOuter, 0.08)
				curveOuter[k] = pOuter
			} else {
				curveInner[k] = spiralPt(th - 2.0*math.Pi)
				curveOuter[k] = spiralPt(th)
			}
		}

		// Alternate metallic tones across sectors and turns for rich bronze/copper/slate-blue rhythm
		turnIdx := sec / sectorsPerTurn
		secInTurn := sec % sectorsPerTurn
		phase := float64(sec)*0.45 + float64(turnIdx)*1.1

		var primaryMetal, secondaryMetal, coolMetal color.RGBA
		switch (secInTurn + turnIdx*2) % 4 {
		case 0:
			primaryMetal = antiqueGold
			secondaryMetal = copperRust
			coolMetal = slateBlue
		case 1:
			primaryMetal = copperRust
			secondaryMetal = antiqueGold
			coolMetal = deepSlate
		case 2:
			primaryMetal = slateBlue
			secondaryMetal = antiqueGold
			coolMetal = deepSlate
		default:
			primaryMetal = antiqueGold
			secondaryMetal = slateBlue
			coolMetal = copperRust
		}

		// Adjust line density slightly for tiny inner chambers vs expansive outer chambers
		rMid := spiralRadius(0.5 * (theta0 + theta1))
		nLines := ns.linesPerChamber
		if rMid < 28.0*scale {
			nLines = int(math.Max(18, float64(ns.linesPerChamber)*0.42))
		} else if rMid < 85.0*scale {
			nLines = int(math.Max(32, float64(ns.linesPerChamber)*0.72))
		}

		// Longitudinal spiral striations (curveA = inner spiral arc, curveB = outer spiral arc)
		colorFnLong := makeChamberColorFn(primaryMetal, secondaryMetal, coolMetal, phase)
		rc.DrawHatchPatch(curveInner, curveOuter, nLines, 1.05*scale, 0.04, colorFnLong, 0.92)

		// 2. Curved Nautilus Septa Chambers (transverse curved ribs between theta0 and theta1)
		if rMid > 12.0*scale {
			septumA := make([]render.Point, curveSamples)
			septumB := make([]render.Point, curveSamples)
			for k := 0; k < curveSamples; k++ {
				v := float64(k) / float64(curveSamples-1)
				// Concave nautilus septum curve bowing backward along the spiral
				lag0 := 0.24 * math.Sin(v*math.Pi)
				lag1 := 0.24 * math.Sin(v*math.Pi)

				var pA0, pA1, pB0, pB1 render.Point
				if theta0 < 2.0*math.Pi {
					pA0 = render.Pt(apexX, apexY)
					pA1 = spiralPt(theta0 - lag0)
					pB0 = render.Pt(apexX, apexY)
					pB1 = spiralPt(theta1 - lag1)
				} else {
					pA0 = spiralPt(theta0 - 2.0*math.Pi - lag0*0.6)
					pA1 = spiralPt(theta0 - lag0)
					pB0 = spiralPt(theta1 - 2.0*math.Pi - lag1*0.6)
					pB1 = spiralPt(theta1 - lag1)
				}
				baseA := pA0.Lerp(pA1, v)
				baseB := pB0.Lerp(pB1, v)
				chordA := pA1.Sub(pA0)
				chordB := pB1.Sub(pB0)
				bowAmt := -0.14 * math.Sin(v*math.Pi)
				septumA[k] = baseA.Add(chordA.Perp().Mul(bowAmt))
				septumB[k] = baseB.Add(chordB.Perp().Mul(bowAmt))
			}

			transLines := int(math.Max(22, float64(nLines)*0.62))
			colorFnTrans := makeChamberColorFn(secondaryMetal, coolMetal, antiqueGold, phase+1.57)
			rc.DrawHatchPatch(septumA, septumB, transLines, 0.92*scale, -0.08, colorFnTrans, 0.56)

			// Engraved metallic bevel along each septum rib
			rc.StrokePolyline(septumA, 2.2*scale, creviceDark, 0.72)
			rc.StrokePolyline(septumA, 1.1*scale, render.LerpOklab(antiqueGold, brightBrass, 0.45), 0.78)
		}
	}

	// 3. Multi-layered 3D Beveled Logarithmic Spiral Spine coiling into the tight center eye
	const spineSteps = 720
	for i := 0; i < spineSteps; i++ {
		u0 := float64(i) / float64(spineSteps)
		u1 := float64(i+1) / float64(spineSteps)
		th0 := u0 * maxTheta
		th1 := u1 * maxTheta
		thMid := 0.5 * (th0 + th1)

		p0 := spiralPt(th0)
		p1 := spiralPt(th1)
		r := spiralRadius(thMid)

		// Spine thickness scales organically with logarithmic radius
		ribWidth := common.Constrain(1.2*scale+r*0.042, 1.1*scale, 16.5*scale)
		seg := p1.Sub(p0)
		if seg.Len() < 1e-6 {
			continue
		}
		n := seg.Norm().Perp()

		// Outer ambient-occlusion shadow groove
		shadowPath := &canvas.Path{}
		shadowPath.MoveTo(p0.X-n.X*ribWidth*0.28, p0.Y-n.Y*ribWidth*0.28)
		shadowPath.LineTo(p1.X-n.X*ribWidth*0.28, p1.Y-n.Y*ribWidth*0.28)
		rc.StrokePath(shadowPath, ribWidth*1.45, canvas.RoundCap, canvas.RoundJoin, creviceDark, 0.85)

		// Cool slate-steel outer bevel
		slatePath := &canvas.Path{}
		slatePath.MoveTo(p0.X-n.X*ribWidth*0.14, p0.Y-n.Y*ribWidth*0.14)
		slatePath.LineTo(p1.X-n.X*ribWidth*0.14, p1.Y-n.Y*ribWidth*0.14)
		rc.StrokePath(slatePath, ribWidth*1.05, canvas.RoundCap, canvas.RoundJoin, deepSlate, 0.90)

		// Warm copper-bronze main spine body
		wave := 0.5 + 0.5*math.Sin(thMid*3.0)
		bodyCol := render.LerpOklab(copperRust, antiqueGold, wave)
		bodyPath := &canvas.Path{}
		bodyPath.MoveTo(p0.X, p0.Y)
		bodyPath.LineTo(p1.X, p1.Y)
		rc.StrokePath(bodyPath, ribWidth*0.72, canvas.RoundCap, canvas.RoundJoin, bodyCol, 0.95)

		// Specular antique-gold / brass ridge along the lit crest of the spiral rib
		crestPath := &canvas.Path{}
		crestPath.MoveTo(p0.X+n.X*ribWidth*0.16, p0.Y+n.Y*ribWidth*0.16)
		crestPath.LineTo(p1.X+n.X*ribWidth*0.16, p1.Y+n.Y*ribWidth*0.16)
		rc.StrokePath(crestPath, math.Max(0.8*scale, ribWidth*0.28), canvas.RoundCap, canvas.RoundJoin, brightBrass, 0.92)
	}

	// Tight luminous bronze/gold pinhead at the logarithmic spiral apex eye
	rc.DrawGlowDot(apexX, apexY, 6.5*scale, creviceDark, 0.92, 0.6)
	rc.DrawGlowDot(apexX, apexY, 3.2*scale, brightBrass, 0.95, 0.4)

	rc.ApplyVignette(0.22)
	c.MergeRenderCanvasToneMapped(rc, 1.06, 0)
}
