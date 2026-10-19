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

type islamicStar struct {
	gridN      int
	starAngle  float64
	strapWidth float64
}

// NewIslamicStar returns an islamicStar engine implementing Hankin's Polygons-in-Contact
// (Girih star rosette & cloisonne strapwork interlace) over a semi-regular 4.8.8 octagon-square tiling.
func NewIslamicStar(gridN int, starAngle, strapWidth float64) *islamicStar {
	if gridN <= 0 {
		gridN = 4
	}
	if starAngle <= 0 {
		starAngle = 67.5
	}
	if strapWidth <= 0 {
		strapWidth = 4.5
	}
	return &islamicStar{
		gridN:      gridN,
		starAngle:  starAngle,
		strapWidth: strapWidth,
	}
}

type girihSegment struct {
	a, b render.Point
}

// scalePolyFromCenter scales a polygon's vertices radially relative to center.
func scalePolyFromCenter(poly []render.Point, center render.Point, s float64) []render.Point {
	out := make([]render.Point, len(poly))
	for i, p := range poly {
		out[i] = center.Lerp(p, s)
	}
	return out
}

// polyCenter computes the arithmetic mean of a polygon's vertices.
func polyCenter(poly []render.Point) render.Point {
	if len(poly) == 0 {
		return render.Point{}
	}
	var sx, sy float64
	for _, p := range poly {
		sx += p.X
		sy += p.Y
	}
	inv := 1.0 / float64(len(poly))
	return render.Pt(sx*inv, sy*inv)
}

// buildOctagonHankin computes the exact Hankin Polygons-in-Contact geometry inside a regular octagon
// centered at center with apothem a (distance from center to each of the 8 edge midpoints).
// It returns:
//   - star16: central 16-vertex star rosette [P_0, Q_0, P_1, Q_1, ..., P_7, Q_7]
//   - innerStar16: nested inner star rosette
//   - kites: 8 radial petal kites (M_{k+1}, P_{k+1}, Q_k, P_k)
//   - corners: 8 corner quadrilaterals (M_k, V_{k+1}, M_{k+1}, P_k)
//   - segments: all Hankin contact-ray segments for interlaced strapwork rendering
func buildOctagonHankin(
	center render.Point,
	apothem float64,
	starAngleDeg float64,
) (
	octVerts []render.Point,
	star16 []render.Point,
	innerStar16 []render.Point,
	kites [][]render.Point,
	corners [][]render.Point,
	segments []girihSegment,
) {
	const n = 8
	step := 2.0 * math.Pi / float64(n) // pi / 4
	halfStep := step * 0.5             // pi / 8

	// Clamp contact angle to a geometrically valid range (32..80 degrees)
	alpha := common.Constrain(starAngleDeg, 32.0, 80.0) * math.Pi / 180.0
	// Angle between inward edge normal (M_k -> center) and the Hankin ray
	psi := math.Pi*0.5 - alpha
	if psi < 0.14 {
		psi = 0.14
	}

	// Circumradius R of the regular octagon from apothem a = R * cos(pi/8)
	rOct := apothem / math.Cos(halfStep)

	// Law of Sines in triangle (center, M_k, intersection):
	// First ray-ray intersection P_k lies at angle theta_k + pi/8 (along vertex V_{k+1} ray)
	r1 := apothem * math.Sin(psi) / math.Sin(halfStep+psi)
	// Second ray-ray intersection Q_k lies at angle theta_k + pi/4 (along midpoint M_{k+1} ray)
	r2 := apothem * math.Sin(psi) / math.Sin(step+psi)

	octVerts = make([]render.Point, n)
	mids := make([]render.Point, n)
	pPts := make([]render.Point, n)
	qPts := make([]render.Point, n)

	for k := 0; k < n; k++ {
		// Edge midpoint M_k is at angle k * pi/4; vertex V_k is at (k - 0.5) * pi/4
		angV := float64(k)*step - halfStep
		angM := float64(k) * step
		angP := angM + halfStep
		angQ := angM + step

		octVerts[k] = render.Pt(center.X+rOct*math.Cos(angV), center.Y+rOct*math.Sin(angV))
		mids[k] = render.Pt(center.X+apothem*math.Cos(angM), center.Y+apothem*math.Sin(angM))
		pPts[k] = render.Pt(center.X+r1*math.Cos(angP), center.Y+r1*math.Sin(angP))
		qPts[k] = render.Pt(center.X+r2*math.Cos(angQ), center.Y+r2*math.Sin(angQ))
	}

	star16 = make([]render.Point, 0, 2*n)
	innerStar16 = make([]render.Point, 0, 2*n)
	for k := 0; k < n; k++ {
		star16 = append(star16, pPts[k], qPts[k])
		// Inner filigree star alternates between scaled Q_k tips and deeper inner valleys
		angP := float64(k)*step + halfStep
		angQ := float64(k)*step + step
		rInnerOuter := r2 * 0.88
		rInnerValley := r2 * 0.46
		innerStar16 = append(innerStar16,
			render.Pt(center.X+rInnerValley*math.Cos(angP), center.Y+rInnerValley*math.Sin(angP)),
			render.Pt(center.X+rInnerOuter*math.Cos(angQ), center.Y+rInnerOuter*math.Sin(angQ)),
		)
	}

	kites = make([][]render.Point, n)
	corners = make([][]render.Point, n)
	segments = make([]girihSegment, 0, n*4)

	for k := 0; k < n; k++ {
		k1 := (k + 1) % n
		k2 := (k + 2) % n

		// Kite k is bounded by M_{k+1}, P_{k+1}, Q_k, P_k
		kites[k] = []render.Point{mids[k1], pPts[k1], qPts[k], pPts[k]}

		// Corner cell around vertex V_{k+1} is bounded by M_k, V_{k+1}, M_{k+1}, P_k
		corners[k] = []render.Point{mids[k], octVerts[k1], mids[k1], pPts[k]}

		// Two Hankin contact rays per edge midpoint M_k:
		// Ray (+) goes M_k -> P_k -> Q_k
		// Ray (-) goes M_{k+2} -> P_{k+1} -> Q_k
		segments = append(segments,
			girihSegment{a: mids[k], b: pPts[k]},
			girihSegment{a: pPts[k], b: qPts[k]},
			girihSegment{a: mids[k2], b: pPts[k1]},
			girihSegment{a: pPts[k1], b: qPts[k]},
		)
	}

	return octVerts, star16, innerStar16, kites, corners, segments
}

// buildSquareHankin computes the exact Hankin Polygons-in-Contact geometry inside an interstitial
// tilted square centered at center with vertices along the 4 cardinal directions (radius rSq).
func buildSquareHankin(
	center render.Point,
	rSq float64,
	starAngleDeg float64,
) (
	sqVerts []render.Point,
	star8 []render.Point,
	corners [][]render.Point,
	segments []girihSegment,
) {
	const n = 4
	step := 2.0 * math.Pi / float64(n) // pi / 2
	halfStep := step * 0.5             // pi / 4

	// Square apothem is rSq * cos(pi/4)
	apothem := rSq * math.Cos(halfStep)

	alpha := common.Constrain(starAngleDeg, 32.0, 80.0) * math.Pi / 180.0
	psi := math.Pi*0.5 - alpha
	if psi < 0.18 {
		psi = 0.18
	}
	if psi > 0.62 {
		psi = 0.62
	}

	// Intersection P_k along the vertex ray (theta_M + pi/4)
	r1 := apothem * math.Sin(psi) / math.Sin(halfStep+psi)

	sqVerts = make([]render.Point, n)
	mids := make([]render.Point, n)
	pPts := make([]render.Point, n)

	for k := 0; k < n; k++ {
		angV := float64(k) * step
		angM := angV + halfStep
		angP := angV + step // vertex V_{k+1} direction

		sqVerts[k] = render.Pt(center.X+rSq*math.Cos(angV), center.Y+rSq*math.Sin(angV))
		mids[k] = render.Pt(center.X+apothem*math.Cos(angM), center.Y+apothem*math.Sin(angM))
		pPts[k] = render.Pt(center.X+r1*math.Cos(angP), center.Y+r1*math.Sin(angP))
	}

	star8 = make([]render.Point, 0, 2*n)
	corners = make([][]render.Point, n)
	segments = make([]girihSegment, 0, 2*n)

	for k := 0; k < n; k++ {
		k1 := (k + 1) % n
		star8 = append(star8, mids[k], pPts[k])
		corners[k] = []render.Point{mids[k], sqVerts[k1], mids[k1], pPts[k]}
		segments = append(segments,
			girihSegment{a: mids[k], b: pPts[k]},
			girihSegment{a: mids[k1], b: pPts[k]},
		)
	}

	return sqVerts, star8, corners, segments
}

// fillCloisonneCell renders a jewel-like cloisonne enamel cell with a solid Oklab base,
// wet-on-wet watercolor pigment pooling, inner bevel ring, and radial specular glow.
func fillCloisonneCell(
	rc *render.Canvas,
	poly []render.Point,
	baseCol, accentCol, goldCol color.RGBA,
	scale float64,
	rng *rand.Rand,
	drawGlow bool,
) {
	if len(poly) < 3 {
		return
	}
	c := polyCenter(poly)
	lab := render.RGBAToOklab(baseCol)

	deepEdgeCol := render.OklabToRGBA(render.Oklab{
		L: common.Constrain(lab.L*0.68, 0.06, 0.90),
		A: lab.A * 1.08,
		B: lab.B * 1.08,
	}, 255)
	luminousCoreCol := render.LerpOklab(
		render.OklabToRGBA(render.Oklab{
			L: common.Constrain(lab.L+0.18, 0.15, 0.96),
			A: lab.A * 0.85,
			B: lab.B * 0.85,
		}, 255),
		accentCol,
		0.28,
	)

	// 1. Deep enamel base fill
	rc.SetBlendMode(render.BlendOver)
	rc.FillPolygon(poly, deepEdgeCol, 0.94)

	// 2. Concentric inner enamel body
	inner1 := scalePolyFromCenter(poly, c, 0.84)
	rc.FillPolygon(inner1, baseCol, 0.86)

	// 3. Wet-on-wet watercolor glaze wash
	rc.SetBlendMode(render.BlendMultiply)
	rc.DrawWatercolorPolygon(inner1, baseCol, 4, 3, 0.18, 0.06, rng)

	// 4. Inner jewel highlight polygon + delicate gold cloisonne wire
	rc.SetBlendMode(render.BlendOver)
	inner2 := scalePolyFromCenter(poly, c, 0.58)
	rc.FillPolygon(inner2, luminousCoreCol, 0.55)
	rc.StrokePolyline(append(inner1, inner1[0]), 0.65*scale, goldCol, 0.32)

	// 5. Radial inner jewel glow
	if drawGlow {
		rad := c.Dist(poly[0]) * 0.72
		if rad > 1.5 {
			rc.SetBlendMode(render.BlendAdd)
			rc.DrawGlowDot(c.X, c.Y, rad, luminousCoreCol, 0.24, 0.18)
			rc.SetBlendMode(render.BlendOver)
		}
	}
}

// Generative renders Hankin's Polygons-in-Contact Islamic star rosette and cloisonne strapwork.
func (is *islamicStar) Generative(c *generativeart.Canva) {
	gridN := is.gridN
	starAngle := is.starAngle
	strapWidth := is.strapWidth
	if gridN <= 0 {
		gridN = 4
	}
	if starAngle <= 0 {
		starAngle = 67.5
	}
	if strapWidth <= 0 {
		strapWidth = 4.5
	}

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

	seed := int64(5551337)
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	schema := c.Opts().ColorSchema()
	if len(schema) == 0 {
		schema = common.Dunhuang
	}
	grad := render.NewMultiStopGradient(schema...)
	ukiyoGrad := render.NewMultiStopGradient(common.UkiyoE...)

	rc := c.NewRenderCanvas()

	// Rich cobalt-lapis ceramic glaze backdrop
	bgCol := render.LerpOklab(color.RGBA{R: 14, G: 24, B: 42, A: 255}, grad.SampleOklab(0.82), 0.22)
	rc.FillBackground(bgCol)

	// Warm gold/ivory strapwork and dark casing palette
	casingCol := color.RGBA{R: 12, G: 15, B: 24, A: 255}
	strapGold := color.RGBA{R: 238, G: 204, B: 136, A: 255}
	strapIvory := color.RGBA{R: 252, G: 243, B: 220, A: 255}
	filigreeGold := color.RGBA{R: 218, G: 172, B: 92, A: 255}

	// Exact 4.8.8 Archimedes tiling geometry:
	// Octagon centers lie on a Cartesian grid with spacing cellStep.
	// Apothem of each octagon is cellStep / 2.
	// Interstitial tilted squares lie at ((i+0.5)*cellStep, (j+0.5)*cellStep)
	// with circumradius rSq = (cellStep / 2) * (2 - sqrt(2)).
	cellStep := math.Max(fw, fh) / float64(gridN)
	apothemOct := cellStep * 0.5
	rSq := apothemOct * (2.0 - math.Sqrt2)

	nx := int(math.Ceil(fw/cellStep)) + 1
	ny := int(math.Ceil(fh/cellStep)) + 1
	offsetX := (fw - float64(gridN)*cellStep) * 0.5
	offsetY := (fh - float64(gridN)*cellStep) * 0.5

	cxCanvas, cyCanvas := fw*0.5, fh*0.5
	maxDiag := math.Hypot(cxCanvas, cyCanvas)
	if maxDiag < 1.0 {
		maxDiag = 1.0
	}

	var allStrapSegments []girihSegment
	var allTilingEdges []girihSegment
	var overBridgeSegments []girihSegment
	var starCenters []render.Point

	// 1. Construct and fill all Octagon rosette cells across the 4.8.8 tiling
	for iy := -1; iy <= ny; iy++ {
		for ix := -1; ix <= nx; ix++ {
			octCenter := render.Pt(
				offsetX+(float64(ix)+0.5)*cellStep,
				offsetY+(float64(iy)+0.5)*cellStep,
			)
			if octCenter.X < -cellStep || octCenter.X > fw+cellStep ||
				octCenter.Y < -cellStep || octCenter.Y > fh+cellStep {
				continue
			}

			radNorm := common.Clamp01(octCenter.Dist(render.Pt(cxCanvas, cyCanvas)) / maxDiag)
			nVal := 0.5 * (noise.FBM2D(octCenter.X*0.003, octCenter.Y*0.003, 3, 2.0, 0.5) + 1.0)
			baseT := math.Mod(0.45*radNorm+0.35*nVal+float64((ix+iy+16)%5)*0.14, 1.0)

			octVerts, star16, innerStar16, kites, corners, segs := buildOctagonHankin(octCenter, apothemOct, starAngle)
			allStrapSegments = append(allStrapSegments, segs...)
			starCenters = append(starCenters, octCenter)

			// Record subtle underlying girih construction polygon edges
			for k := 0; k < len(octVerts); k++ {
				allTilingEdges = append(allTilingEdges, girihSegment{
					a: octVerts[k],
					b: octVerts[(k+1)%len(octVerts)],
				})
			}

			// Collect alternating over-pass bridge segments near P_k for authentic over-under weave
			for k := 0; k < len(segs); k += 4 {
				// segs[k] is M_k -> P_k and segs[k+1] is P_k -> Q_k (collinear ray R_k^+)
				// segs[k+2] is M_{k+2} -> P_{k+1} and segs[k+3] is P_{k+1} -> Q_k (collinear ray R_{k+2}^-)
				p1 := segs[k].a.Lerp(segs[k+1].b, 0.22)
				p2 := segs[k].a.Lerp(segs[k+1].b, 0.78)
				overBridgeSegments = append(overBridgeSegments, girihSegment{a: p1, b: p2})
			}

			// Fill 8 corner hexagon/chevron cells
			cornerCol := render.LerpOklab(grad.SampleOklab(math.Mod(baseT+0.62, 1.0)), ukiyoGrad.SampleOklab(0.18), 0.35)
			for _, cPoly := range corners {
				fillCloisonneCell(rc, cPoly, cornerCol, strapGold, filigreeGold, scale, rng, false)
			}

			// Fill 8 radial petal kites with alternating complementary jewel tones
			kiteColA := grad.SampleOklab(math.Mod(baseT+0.18, 1.0))
			kiteColB := render.LerpOklab(grad.SampleOklab(math.Mod(baseT+0.42, 1.0)), ukiyoGrad.SampleOklab(0.68), 0.32)
			for k, kPoly := range kites {
				kc := kiteColA
				if k%2 == 1 {
					kc = kiteColB
				}
				fillCloisonneCell(rc, kPoly, kc, strapIvory, filigreeGold, scale, rng, true)

				// Delicate inner kite filigree motif
				kCenter := polyCenter(kPoly)
				innerKite := scalePolyFromCenter(kPoly, kCenter, 0.52)
				rc.StrokePolyline(append(innerKite, innerKite[0]), 0.75*scale, strapGold, 0.48)
				rc.SetBlendMode(render.BlendAdd)
				rc.DrawGlowDot(kCenter.X, kCenter.Y, 3.2*scale, strapIvory, 0.38, 0.45)
				rc.SetBlendMode(render.BlendOver)
			}

			// Fill central 16-vertex Girih star rosette
			starCol := grad.SampleOklab(baseT)
			fillCloisonneCell(rc, star16, starCol, strapGold, filigreeGold, scale, rng, true)

			// Fill nested inner 16-vertex star & central 8-pointed gold medallion
			innerCol := render.LerpOklab(grad.SampleOklab(math.Mod(baseT+0.5, 1.0)), strapGold, 0.38)
			fillCloisonneCell(rc, innerStar16, innerCol, strapIvory, strapGold, scale, rng, true)

			coreStar := scalePolyFromCenter(innerStar16, octCenter, 0.52)
			rc.FillPolygon(coreStar, strapGold, 0.85)
			rc.StrokePolyline(append(innerStar16, innerStar16[0]), 1.1*scale, strapIvory, 0.75)
		}
	}

	// 2. Construct and fill all interstitial Tilted Square cells
	for iy := -1; iy <= ny; iy++ {
		for ix := -1; ix <= nx; ix++ {
			sqCenter := render.Pt(
				offsetX+float64(ix)*cellStep,
				offsetY+float64(iy)*cellStep,
			)
			if sqCenter.X < -cellStep || sqCenter.X > fw+cellStep ||
				sqCenter.Y < -cellStep || sqCenter.Y > fh+cellStep {
				continue
			}

			radNorm := common.Clamp01(sqCenter.Dist(render.Pt(cxCanvas, cyCanvas)) / maxDiag)
			baseT := math.Mod(0.35*radNorm+float64((ix*3+iy+20)%6)*0.16, 1.0)

			sqVerts, star8, corners, segs := buildSquareHankin(sqCenter, rSq, starAngle)
			allStrapSegments = append(allStrapSegments, segs...)
			starCenters = append(starCenters, sqCenter)

			for k := 0; k < len(sqVerts); k++ {
				allTilingEdges = append(allTilingEdges, girihSegment{
					a: sqVerts[k],
					b: sqVerts[(k+1)%len(sqVerts)],
				})
			}

			for k := 0; k < len(segs); k += 2 {
				p1 := segs[k].a.Lerp(segs[k].b, 0.18)
				p2 := segs[k].a.Lerp(segs[k].b, 0.82)
				overBridgeSegments = append(overBridgeSegments, girihSegment{a: p1, b: p2})
			}

			sqCornerCol := render.LerpOklab(ukiyoGrad.SampleOklab(math.Mod(baseT+0.25, 1.0)), grad.SampleOklab(baseT), 0.5)
			for _, cPoly := range corners {
				fillCloisonneCell(rc, cPoly, sqCornerCol, strapGold, filigreeGold, scale, rng, false)
			}

			sqStarCol := render.LerpOklab(grad.SampleOklab(math.Mod(baseT+0.35, 1.0)), strapGold, 0.25)
			fillCloisonneCell(rc, star8, sqStarCol, strapIvory, strapGold, scale, rng, true)

			innerSqStar := scalePolyFromCenter(star8, sqCenter, 0.54)
			rc.FillPolygon(innerSqStar, strapGold, 0.72)
			rc.StrokePolyline(append(innerSqStar, innerSqStar[0]), 0.85*scale, strapIvory, 0.68)
		}
	}

	// 3. Render subtle Girih construction tessellation lines underneath the strapwork
	rc.SetBlendMode(render.BlendOver)
	tilingPath := &canvas.Path{}
	for _, seg := range allTilingEdges {
		tilingPath.MoveTo(seg.a.X, seg.a.Y)
		tilingPath.LineTo(seg.b.X, seg.b.Y)
	}
	rc.StrokePath(tilingPath, 0.9*scale, canvas.RoundCap, canvas.RoundJoin, filigreeGold, 0.28)

	// 4. Render the interlaced Hankin strapwork ribbons using tdewolff/canvas Path + rc.StrokePath:
	// Pass A: Full strapwork network (wider dark shadow/border casing, warm-gold ribbon, ivory core)
	strapPath := &canvas.Path{}
	for _, seg := range allStrapSegments {
		strapPath.MoveTo(seg.a.X, seg.a.Y)
		strapPath.LineTo(seg.b.X, seg.b.Y)
	}

	sw := strapWidth * scale
	rc.StrokePath(strapPath, sw*1.72, canvas.RoundCap, canvas.RoundJoin, casingCol, 0.92)
	rc.StrokePath(strapPath, sw*1.04, canvas.RoundCap, canvas.RoundJoin, strapGold, 0.96)
	rc.StrokePath(strapPath, sw*0.42, canvas.RoundCap, canvas.RoundJoin, strapIvory, 0.82)

	// Pass B: Over-under interlace crossings (re-stroke alternating over-strands with casing + ribbon)
	bridgePath := &canvas.Path{}
	for _, seg := range overBridgeSegments {
		bridgePath.MoveTo(seg.a.X, seg.a.Y)
		bridgePath.LineTo(seg.b.X, seg.b.Y)
	}
	rc.StrokePath(bridgePath, sw*1.68, canvas.ButtCap, canvas.RoundJoin, casingCol, 0.90)
	rc.StrokePath(bridgePath, sw*1.02, canvas.RoundCap, canvas.RoundJoin, strapGold, 0.98)
	rc.StrokePath(bridgePath, sw*0.40, canvas.RoundCap, canvas.RoundJoin, strapIvory, 0.88)

	// 5. Luminous central rosette pearls and radial glow accents
	for _, sc := range starCenters {
		rc.SetBlendMode(render.BlendAdd)
		rc.DrawGlowDot(sc.X, sc.Y, 9.5*scale, strapGold, 0.42, 0.25)
		rc.DrawGlowDot(sc.X, sc.Y, 3.2*scale, strapIvory, 0.85, 0.65)
	}
	rc.SetBlendMode(render.BlendOver)

	rc.ApplyPaperTexture(0.38, 555)
	rc.ApplyVignette(0.25)
	c.MergeRenderCanvas(rc)
}
