package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type escherImpossibleBelvedere struct {
	cols, rows    int
	morphStrength float64
}

// NewEscherImpossibleBelvedere returns an escherImpossibleBelvedere engine that renders an
// axis-aligned Square Grid (symmetry = 4) M.C. Escher Metamorphosis woodcut print:
// across the orthogonal square lattice, 3D recessed 4-chambered square architectural coffers
// at the outer perimeter continuously morph with a 0.00px zero-gap guarantee into interlocking
// winged Koi-Rays / Pegasus creatures along the central diagonal wave.
func NewEscherImpossibleBelvedere(cols, rows int, morphStrength float64) *escherImpossibleBelvedere {
	if cols <= 0 {
		cols = 8
	}
	if rows <= 0 {
		rows = 8
	}
	if morphStrength <= 0 {
		morphStrength = 1.0
	}
	return &escherImpossibleBelvedere{
		cols:          cols,
		rows:          rows,
		morphStrength: morphStrength,
	}
}

// Generative renders a static frame of the Square-Grid Escher Metamorphosis print.
func (eib *escherImpossibleBelvedere) Generative(c *generativeart.Canva) {
	eib.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360-degree phase loop (phase in [0, 1)).
func (eib *escherImpossibleBelvedere) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
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

	seed := c.Seed()
	if seed == 0 {
		seed = 20262903
	}

	cols := common.ConstrainInt(eib.cols, 5, 14)
	rows := common.ConstrainInt(eib.rows, 5, 14)
	if cols >= 10 {
		cols = 8
	}
	if rows >= 10 {
		rows = 8
	}
	strength := common.Constrain(eib.morphStrength, 0.4, 1.4)
	isAnim := totalFrames > 1
	phi := 2.0 * math.Pi * phase

	// Curated 3-Chromatic Square-Coffer Woodcut Palette:
	// Family 0: Warm Museum Alabaster Ivory (#faf3e3) + Dark Ink (#181c26)
	// Family 1: Venetian Terracotta-Vermilion (#d44a32) + Burnished Gold (#f6be44)
	// Family 2: Imperial Cobalt-Prussian Blue (#163464) + Celadon-Turquoise (#32b4a8)
	ivoryParchment := color.RGBA{R: 250, G: 243, B: 227, A: 255}
	brightIvory := color.RGBA{R: 255, G: 252, B: 244, A: 255}
	darkInk := color.RGBA{R: 22, G: 26, B: 36, A: 255}
	terracottaRed := color.RGBA{R: 212, G: 74, B: 50, A: 255}
	burnishedGold := color.RGBA{R: 246, G: 190, B: 68, A: 255}
	cobaltBlue := color.RGBA{R: 22, G: 52, B: 100, A: 255}
	celadonAccent := color.RGBA{R: 50, G: 180, B: 168, A: 255}
	antiqueBrass := color.RGBA{R: 186, G: 142, B: 64, A: 255}

	type sqTheme struct {
		baseCol       color.RGBA
		cofferTop     color.RGBA
		cofferLeft    color.RGBA
		cofferBottom  color.RGBA
		cofferRight   color.RGBA
		creatureTop   color.RGBA
		creatureLeft  color.RGBA
		creatureBot   color.RGBA
		creatureRight color.RGBA
		lineCol       color.RGBA
		accentCol     color.RGBA
		eyeRingCol    color.RGBA
	}

	themes := [3]sqTheme{
		// Family 0: Warm Museum Ivory + Dark Ink & Terracotta
		{
			baseCol:       ivoryParchment,
			cofferTop:     brightIvory,
			cofferLeft:    render.LerpOklab(ivoryParchment, darkInk, 0.14),
			cofferBottom:  render.LerpOklab(ivoryParchment, darkInk, 0.28),
			cofferRight:   render.LerpOklab(ivoryParchment, darkInk, 0.44),
			creatureTop:   brightIvory,
			creatureLeft:  ivoryParchment,
			creatureBot:   render.LerpOklab(ivoryParchment, darkInk, 0.10),
			creatureRight: render.LerpOklab(ivoryParchment, darkInk, 0.18),
			lineCol:       darkInk,
			accentCol:     terracottaRed,
			eyeRingCol:    burnishedGold,
		},
		// Family 1: Venetian Terracotta-Vermilion + Burnished Gold
		{
			baseCol:       terracottaRed,
			cofferTop:     render.LerpOklab(terracottaRed, ivoryParchment, 0.40),
			cofferLeft:    render.LerpOklab(terracottaRed, burnishedGold, 0.22),
			cofferBottom:  render.LerpOklab(terracottaRed, darkInk, 0.24),
			cofferRight:   render.LerpOklab(terracottaRed, darkInk, 0.46),
			creatureTop:   render.LerpOklab(terracottaRed, burnishedGold, 0.28),
			creatureLeft:  terracottaRed,
			creatureBot:   render.LerpOklab(terracottaRed, darkInk, 0.12),
			creatureRight: render.LerpOklab(terracottaRed, darkInk, 0.22),
			lineCol:       darkInk,
			accentCol:     burnishedGold,
			eyeRingCol:    ivoryParchment,
		},
		// Family 2: Imperial Cobalt-Prussian Blue + Celadon-Turquoise
		{
			baseCol:       cobaltBlue,
			cofferTop:     render.LerpOklab(cobaltBlue, celadonAccent, 0.46),
			cofferLeft:    render.LerpOklab(cobaltBlue, celadonAccent, 0.24),
			cofferBottom:  render.LerpOklab(cobaltBlue, darkInk, 0.24),
			cofferRight:   render.LerpOklab(cobaltBlue, darkInk, 0.48),
			creatureTop:   render.LerpOklab(cobaltBlue, celadonAccent, 0.28),
			creatureLeft:  cobaltBlue,
			creatureBot:   render.LerpOklab(cobaltBlue, darkInk, 0.14),
			creatureRight: render.LerpOklab(cobaltBlue, darkInk, 0.24),
			lineCol:       celadonAccent,
			accentCol:     ivoryParchment,
			eyeRingCol:    burnishedGold,
		},
	}

	rc := c.NewRenderCanvas()
	rc.FillBackground(ivoryParchment)
	rc.SetBlendMode(render.BlendOver)

	margin := 24.0 * scale
	innerW := math.Max(64.0, fw-2.0*margin)
	innerH := math.Max(64.0, fh-2.0*margin)
	cx, cy := 0.5*fw, 0.5*fh

	cellW := innerW / float64(cols)
	cellH := innerH / float64(rows)
	cellSize := math.Min(cellW, cellH)
	originX := cx - 0.5*float64(cols)*cellSize
	originY := cy - 0.5*float64(rows)*cellSize

	framePoly := []render.Point{
		render.Pt(margin, margin),
		render.Pt(fw-margin, margin),
		render.Pt(fw-margin, fh-margin),
		render.Pt(margin, fh-margin),
	}
	rc.PushClipPolygon(framePoly)

	// Spatial Metamorphosis Field evalMorphT(pt):
	// Outer corners & perimeter are pure straight-edged 3D Recessed Square Coffers (t = 0.0);
	// Diagonal S-corridor across the square grid smoothly morphs into interlocking 4-winged creatures (t -> 1.0)!
	evalMorphT := func(pt render.Point) float64 {
		nx := (pt.X - cx) / (0.5 * innerW)
		ny := (pt.Y - cy) / (0.5 * innerH)

		const invSqrt2 = 0.7071067811865475
		u := (nx + ny) * invSqrt2
		v := (nx - ny) * invSqrt2

		if isAnim {
			wavePhase := phi - u*2.3 + 0.40*math.Sin(1.6*v-phi)
			rawWave := 0.5 + 0.5*math.Cos(wavePhase)
			fullCycle := common.Smoothstep(0.10, 0.90, rawWave)
			radGuard := common.Smoothstep(1.32, 0.80, math.Hypot(nx*0.74, ny*0.74))
			return strength * 1.15 * fullCycle * (0.25 + 0.75*radGuard)
		}

		curvedU := u - 0.18*math.Sin(1.4*v)
		bandDist := math.Abs(curvedU) / 1.15
		bandRamp := common.Constrain(1.0-bandDist, 0.0, 1.0)
		bandEnv := 0.50*bandRamp + 0.50*(0.5-0.5*math.Cos(math.Pi*bandRamp))

		radDist := math.Hypot(nx*0.78, ny*0.78)
		radRamp := common.Constrain((1.20-radDist)/1.06, 0.0, 1.0)
		radEnv := 0.48*radRamp + 0.52*(0.5-0.5*math.Cos(math.Pi*radRamp))

		combined := bandEnv * (0.28 + 0.72*radEnv)
		norm := common.Constrain((combined-0.06)/0.84, 0.0, 1.0)
		gentle := 0.55*norm + 0.45*(0.5-0.5*math.Cos(math.Pi*norm))
		return strength * 1.08 * gentle
	}

	// Axis-Aligned Square Grid Vertex G(col, row):
	// During GIF animation (isAnim), shifting (col, row) by (-phase, -phase) translates the
	// entire square lattice diagonally along (-1, -1) while preserving ((col - row) % 3) color parity!
	sqVertex := func(col, row int) render.Point {
		fc := float64(col)
		fr := float64(row)
		if isAnim {
			fc -= phase
			fr -= phase
		}
		return render.Pt(originX+fc*cellSize, originY+fr*cellSize)
	}

	const samplesPerEdge = 22

	// Canonical Horizontal Grid Edge H(col, row): G(col, row) -> G(col+1, row)
	// Shared between Top edge of square (col, row) and reversed Bottom edge of square (col, row-1)!
	evalHEdge := func(col, row int) ([]render.Point, float64) {
		a := sqVertex(col, row)
		b := sqVertex(col+1, row)
		mid := a.Lerp(b, 0.5)
		mT := evalMorphT(mid)

		pts := make([]render.Point, samplesPerEdge+1)
		pts[0] = a
		for i := 1; i < samplesPerEdge; i++ {
			t := float64(i) / float64(samplesPerEdge)
			win := math.Sin(math.Pi * t)

			// Interlocking Escher Dorsal Fin / Wing S-Scroll along Horizontal Edges
			dy := -mT * cellSize * (0.25*math.Sin(2.0*math.Pi*t-0.28) + 0.10*math.Sin(4.0*math.Pi*t-0.4) + 0.06*math.Sin(math.Pi*t)) * win
			dx := mT * cellSize * (0.08 * math.Sin(2.0*math.Pi*t-0.5) * win * win)

			pts[i] = render.Pt(a.X+(b.X-a.X)*t+dx, a.Y+dy)
		}
		pts[samplesPerEdge] = b
		return pts, mT
	}

	// Canonical Vertical Grid Edge V(col, row): G(col, row) -> G(col, row+1)
	// Shared between Right edge of square (col-1, row) and reversed Left edge of square (col, row)!
	evalVEdge := func(col, row int) ([]render.Point, float64) {
		a := sqVertex(col, row)
		b := sqVertex(col, row+1)
		mid := a.Lerp(b, 0.5)
		mT := evalMorphT(mid)

		pts := make([]render.Point, samplesPerEdge+1)
		pts[0] = a
		for i := 1; i < samplesPerEdge; i++ {
			t := float64(i) / float64(samplesPerEdge)
			win := math.Sin(math.Pi * t)

			// Complementary Interlocking Head-Snout & Bifurcated Tail S-Scroll along Vertical Edges
			dx := -mT * cellSize * (0.24*math.Sin(2.0*math.Pi*t+0.28) - 0.10*math.Sin(4.0*math.Pi*t) + 0.06*math.Sin(math.Pi*t)) * win
			dy := -mT * cellSize * (0.08 * math.Sin(2.0*math.Pi*t) * win * win)

			pts[i] = render.Pt(a.X+dx, a.Y+(b.Y-a.Y)*t+dy)
		}
		pts[samplesPerEdge] = b
		return pts, mT
	}

	reversePts := func(in []render.Point) []render.Point {
		out := make([]render.Point, len(in))
		for i := range in {
			out[i] = in[len(in)-1-i]
		}
		return out
	}

	evalCurve := func(pts []render.Point, s float64) render.Point {
		s = common.Constrain(s, 0.0, 1.0)
		fIdx := s * float64(len(pts)-1)
		i0 := int(math.Floor(fIdx))
		i1 := i0 + 1
		if i1 >= len(pts) {
			return pts[len(pts)-1]
		}
		return pts[i0].Lerp(pts[i1], fIdx-float64(i0))
	}

	// Coons-Patch inside a 3-sided / 4-sided chamber bounded by seamA (hub -> cornerA), outer edgeE (cornerA -> cornerB), and seamB (hub -> cornerB):
	// Parameter u in [0, 1] goes radially from hub (u=0) to outer edgeE (u=1);
	// Parameter v in [0, 1] goes angularly from seamA (v=0) to seamB (v=1)!
	triChamberPoint := func(seamA, edgeE, seamB []render.Point, u, v float64) render.Point {
		u = common.Constrain(u, 0.0, 1.0)
		v = common.Constrain(v, 0.0, 1.0)
		pA := evalCurve(seamA, u)
		pB := evalCurve(seamB, u)
		chordAB := pA.Lerp(pB, v)

		// Blend outer edge curve deviation proportionally to u^1.4 so u=1 lands 100% on edgeE(v) and u=0 lands at hub!
		ePt := evalCurve(edgeE, v)
		eChord := seamA[len(seamA)-1].Lerp(seamB[len(seamB)-1], v)
		edgeDev := ePt.Sub(eChord).Mul(math.Pow(u, 1.35))
		return chordAB.Add(edgeDev)
	}

	type sqCell struct {
		col, row  int
		colorIdx  int
		center    render.Point
		morphT    float64
		normT     float64
		verts     [4]render.Point
		edges     [4][]render.Point // 0: Top (V0->V1), 1: Right (V1->V2), 2: Bottom (V2->V3), 3: Left (V3->V0)
		edgeMorph [4]float64
		boundary  []render.Point
	}

	var cells []sqCell
	for r := -2; r <= rows+2; r++ {
		for col := -2; col <= cols+2; col++ {
			v0 := sqVertex(col, r)
			v1 := sqVertex(col+1, r)
			v2 := sqVertex(col+1, r+1)
			v3 := sqVertex(col, r+1)
			cPt := render.Pt(0.5*(v0.X+v2.X), 0.5*(v0.Y+v2.Y))

			if cPt.X < margin-cellSize*1.5 || cPt.X > fw-margin+cellSize*1.5 ||
				cPt.Y < margin-cellSize*1.5 || cPt.Y > fh-margin+cellSize*1.5 {
				continue
			}

			e0, m0 := evalHEdge(col, r)
			e1, m1 := evalVEdge(col+1, r)
			e2Rev, m2 := evalHEdge(col, r+1)
			e3Rev, m3 := evalVEdge(col, r)
			e2 := reversePts(e2Rev)
			e3 := reversePts(e3Rev)

			boundary := make([]render.Point, 0, 4*samplesPerEdge)
			boundary = append(boundary, e0[:samplesPerEdge]...)
			boundary = append(boundary, e1[:samplesPerEdge]...)
			boundary = append(boundary, e2[:samplesPerEdge]...)
			boundary = append(boundary, e3[:samplesPerEdge]...)

			mT := evalMorphT(cPt)
			nT := common.Constrain(mT/math.Max(strength, 0.01), 0.0, 1.0)

			// 3-Chromatic Diagonal Square Coloring: ((col - r) % 3 + 3) % 3
			cIdx := ((col - r) % 3)
			if cIdx < 0 {
				cIdx += 3
			}

			cells = append(cells, sqCell{
				col:       col,
				row:       r,
				colorIdx:  cIdx,
				center:    cPt,
				morphT:    mT,
				normT:     nT,
				verts:     [4]render.Point{v0, v1, v2, v3},
				edges:     [4][]render.Point{e0, e1, e2, e3},
				edgeMorph: [4]float64{m0, m1, m2, m3},
				boundary:  boundary,
			})
		}
	}

	// Pass 1: 4-Chamber Pyramidal Square Coffer -> Origami Pinwheel -> Interlocking Winged Koi-Ray
	for idx := range cells {
		cell := &cells[idx]
		th := themes[cell.colorIdx]
		normT := cell.normT

		smoothT := 0.45*normT + 0.55*escherCubicSmoothstep(0.0, 1.0, normT)
		wCoffer := 1.0 - escherCubicSmoothstep(0.12, 0.76, normT)
		wPinwheel := math.Sin(math.Pi * common.Constrain((normT-0.08)/0.86, 0.0, 1.0))
		wCreature := escherCubicSmoothstep(0.22, 0.92, normT)
		wEye := escherCubicSmoothstep(0.38, 0.94, normT)

		rc.FillPolygon(cell.boundary, th.baseCol, 1.0)
		closedB := append(append([]render.Point(nil), cell.boundary...), cell.boundary[0])
		rc.StrokePolyline(closedB, 0.85*scale, th.baseCol, 1.0)

		rc.PushClipPolygon(cell.boundary)

		// Central hub shifts slightly toward the upper-left swim/flight axis (-135 deg) as smoothT grows
		flightDir := render.Pt(-0.7071, -0.7071)
		hub := cell.center.Add(flightDir.Mul(0.10 * cellSize * smoothT))

		// Build the 4 internal diagonal seams from hub to the 4 square corners V0, V1, V2, V3.
		// At morphT = 0, these are the 4 straight 45-deg miter diagonals of a 3D recessed square pyramid coffer!
		// As morphT -> 1, they curve into a 4-bladed pinwheel / anatomical wing-and-tail divider!
		buildDiagSeam := func(targetV render.Point, bendSign, amp float64) []render.Point {
			chord := targetV.Sub(hub)
			cLen := chord.Len()
			tan := chord.Mul(1.0 / math.Max(cLen, 1e-6))
			norm := tan.Perp()
			pts := make([]render.Point, samplesPerEdge+1)
			for i := 0; i <= samplesPerEdge; i++ {
				t := float64(i) / float64(samplesPerEdge)
				env := math.Sin(math.Pi * t)
				dn := bendSign * amp * cell.morphT * (0.20*math.Sin(2.0*math.Pi*t-0.3) + 0.08*env)
				du := 0.04 * cell.morphT * math.Sin(2.0*math.Pi*t) * env
				pts[i] = hub.Lerp(targetV, t).Add(tan.Mul(du * cLen)).Add(norm.Mul(dn * cLen))
			}
			return pts
		}

		seam0 := buildDiagSeam(cell.verts[0], +1.0, 0.78) // hub -> Top-Left V0
		seam1 := buildDiagSeam(cell.verts[1], +1.0, 0.82) // hub -> Top-Right V1
		seam2 := buildDiagSeam(cell.verts[2], -1.0, 0.82) // hub -> Bottom-Right V2
		seam3 := buildDiagSeam(cell.verts[3], -1.0, 0.78) // hub -> Bottom-Left V3

		buildTriChamber := func(sA, eOuter, sB []render.Point) []render.Point {
			poly := make([]render.Point, 0, len(sA)+len(eOuter)+len(sB))
			poly = append(poly, sA...)
			poly = append(poly, eOuter[1:samplesPerEdge]...)
			for i := len(sB) - 1; i >= 1; i-- {
				poly = append(poly, sB[i])
			}
			return poly
		}

		// 4 Chambers of the Square:
		// - topChamber    (seam0 -> e0 -> seam1): Sunlit Top Coffer Wall -> Upper Dorsal Wing
		// - rightChamber  (seam1 -> e1 -> seam2): Deep-Shaded Right Coffer Wall -> Fanned Caudal Tail
		// - bottomChamber (seam2 -> e2 -> seam3): Medium-Shaded Bottom Coffer Wall -> Lower Ventral Wing
		// - leftChamber   (seam3 -> e3 -> seam0): Light-Shaded Left Coffer Wall -> Head, Gill & Eye
		topChamber := buildTriChamber(seam0, cell.edges[0], seam1)
		rightChamber := buildTriChamber(seam1, cell.edges[1], seam2)
		bottomChamber := buildTriChamber(seam2, cell.edges[2], seam3)
		leftChamber := buildTriChamber(seam3, cell.edges[3], seam0)

		topCol := render.LerpOklab(th.cofferTop, th.creatureTop, smoothT)
		rightCol := render.LerpOklab(th.cofferRight, th.creatureRight, smoothT)
		botCol := render.LerpOklab(th.cofferBottom, th.creatureBot, smoothT)
		leftCol := render.LerpOklab(th.cofferLeft, th.creatureLeft, smoothT)

		rc.FillPolygon(topChamber, topCol, 1.0)
		rc.FillPolygon(rightChamber, rightCol, 1.0)
		rc.FillPolygon(bottomChamber, botCol, 1.0)
		rc.FillPolygon(leftChamber, leftCol, 1.0)

		// Stage 1: 3D Recessed Square Coffer Step Rings & Woodcut Hatching (wCoffer > 0.01)
		if wCoffer > 0.01 {
			// 3 concentric recessed square step rings around the 4 chambers!
			for rIdx, uRing := range [3]float64{0.34, 0.58, 0.80} {
				var sqRing []render.Point
				chambers := [4][3][]render.Point{
					{seam0, cell.edges[0], seam1},
					{seam1, cell.edges[1], seam2},
					{seam2, cell.edges[2], seam3},
					{seam3, cell.edges[3], seam0},
				}
				for _, ch := range chambers {
					for s := 0; s < 8; s++ {
						v := float64(s) / 8.0
						sqRing = append(sqRing, triChamberPoint(ch[0], ch[1], ch[2], uRing, v))
					}
				}
				sqRing = append(sqRing, sqRing[0])
				rCol := th.accentCol
				if rIdx == 1 {
					rCol = th.lineCol
				}
				rc.StrokePolyline(sqRing, (0.76-0.12*float64(rIdx))*scale, rCol, (0.68-0.12*float64(rIdx))*wCoffer)
			}

			// Parallel woodcut engraving hatch lines on the shaded Bottom, Left, and Right coffer walls
			drawCofferHatch := func(sA, eOuter, sB []render.Point, shadeDepth float64, cross bool) {
				nH := common.ConstrainInt(int(math.Round(cellSize/(5.5*scale))), 4, 9)
				hW := (0.46 + 0.48*shadeDepth) * scale
				hAlpha := common.Constrain(0.24+0.52*shadeDepth, 0.16, 0.78) * wCoffer
				for i := 1; i < nH; i++ {
					u := 0.14 + 0.78*(float64(i)/float64(nH))
					pts := make([]render.Point, 9)
					for j := 0; j <= 8; j++ {
						pts[j] = triChamberPoint(sA, eOuter, sB, u, 0.06+0.88*float64(j)/8.0)
					}
					rc.StrokePolyline(pts, hW, th.lineCol, hAlpha)
				}
				if cross {
					for j := 1; j < 6; j++ {
						v := float64(j) / 6.0
						pts := make([]render.Point, 7)
						for i := 0; i <= 6; i++ {
							pts[i] = triChamberPoint(sA, eOuter, sB, 0.16+0.76*float64(i)/6.0, v)
						}
						rc.StrokePolyline(pts, 0.44*scale, th.lineCol, 0.44*wCoffer)
					}
				}
			}

			drawCofferHatch(seam3, cell.edges[3], seam0, 0.24, false)
			drawCofferHatch(seam2, cell.edges[2], seam3, 0.46, false)
			drawCofferHatch(seam1, cell.edges[1], seam2, 0.74, true)
		}

		// Stage 2: Transitional Origami Pinwheel Chevrons (wPinwheel > 0.02)
		if wPinwheel > 0.02 {
			drawPinwheelRibs := func(sA, eOuter, sB []render.Point) {
				for cIdx := 1; cIdx <= 3; cIdx++ {
					u := 0.25 + 0.22*float64(cIdx)
					pL := triChamberPoint(sA, eOuter, sB, u, 0.14)
					pM := triChamberPoint(sA, eOuter, sB, common.Clamp01(u+0.12*smoothT), 0.50)
					pR := triChamberPoint(sA, eOuter, sB, u, 0.86)
					rib := render.CatmullRomSpline([]render.Point{pL, pM, pR}, 6)
					rc.StrokePolyline(rib, 0.62*scale, th.lineCol, 0.52*wPinwheel)
				}
			}
			drawPinwheelRibs(seam0, cell.edges[0], seam1)
			drawPinwheelRibs(seam2, cell.edges[2], seam3)
		}

		// Stage 3: Full Winged Koi-Ray / Pegasus Fin Pinions, Caudal Tail Fan & Eye (wCreature > 0.01)
		if wCreature > 0.01 {
			// Primary Wing / Pectoral Fin Pinions in Top Chamber (seam0 -> e0 -> seam1) & Bottom Chamber (seam2 -> e2 -> seam3)
			drawWingChamberPinions := func(sA, eOuter, sB []render.Point, flip float64) {
				const nPinions = 5
				for f := 1; f <= nPinions; f++ {
					vf := float64(f) / float64(nPinions+1)
					pts := make([]render.Point, 8)
					for s := 0; s <= 7; s++ {
						ts := float64(s) / 7.0
						u := 0.16 + 0.72*ts*(0.45+0.55*wCreature)
						v := common.Constrain(vf+flip*0.10*math.Sin(ts*math.Pi)*wCreature, 0.10, 0.90)
						pts[s] = triChamberPoint(sA, eOuter, sB, u, v)
					}
					rc.StrokePolyline(pts, (0.68+0.32*(1.0-vf))*scale, th.lineCol, 0.76*wCreature)
					if f%2 == 1 {
						rc.StrokePolyline(pts, 0.46*scale, th.accentCol, 0.52*wCreature)
					}
				}
			}
			drawWingChamberPinions(seam0, cell.edges[0], seam1, +1.0)
			drawWingChamberPinions(seam2, cell.edges[2], seam3, -1.0)

			// Fanned Caudal Tail Volutes in Right Chamber (seam1 -> e1 -> seam2)
			for tr := 1; tr <= 5; tr++ {
				vf := float64(tr) / 6.0
				pts := make([]render.Point, 8)
				for s := 0; s <= 7; s++ {
					ts := float64(s) / 7.0
					u := 0.14 + 0.74*ts*(0.48+0.52*wCreature)
					v := common.Constrain(vf+(vf-0.5)*0.18*math.Sin(ts*math.Pi)*wCreature, 0.10, 0.90)
					pts[s] = triChamberPoint(seam1, cell.edges[1], seam2, u, v)
				}
				rc.StrokePolyline(pts, 0.66*scale, th.lineCol, 0.68*wCreature)
			}

			// Sculpted Gill Collar & Eye in Left Head Chamber (seam3 -> e3 -> seam0)
			for gIdx := 1; gIdx <= 2; gIdx++ {
				ug := 0.34 + 0.16*float64(gIdx)
				gL := triChamberPoint(seam3, cell.edges[3], seam0, ug, 0.18)
				gM := triChamberPoint(seam3, cell.edges[3], seam0, ug+0.08, 0.50)
				gR := triChamberPoint(seam3, cell.edges[3], seam0, ug, 0.82)
				gArc := render.CatmullRomSpline([]render.Point{gL, gM, gR}, 6)
				rc.StrokePolyline(gArc, 0.74*scale, th.accentCol, 0.62*wCreature)
			}
		}

		// Continuous 4 Internal Diagonal Seams (at t=0, exact 45-deg square coffer miter lines!)
		seamW := (1.02 + 0.34*smoothT) * scale
		rc.StrokePolyline(seam0, seamW, th.lineCol, 0.84)
		rc.StrokePolyline(seam1, seamW, th.lineCol, 0.84)
		rc.StrokePolyline(seam2, seamW, th.lineCol, 0.84)
		rc.StrokePolyline(seam3, seamW, th.lineCol, 0.84)

		// Progressive Diamond-to-Circular Golden Eye in the Left Head Chamber (wEye > 0.01)
		if wEye > 0.01 {
			headAnchor := triChamberPoint(seam3, cell.edges[3], seam0, 0.64, 0.52)
			eyeR := (1.35 + 1.95*wEye) * scale
			rc.DrawGlowDot(headAnchor.X, headAnchor.Y, eyeR*1.42, th.lineCol, 0.88*wEye, 0.92)
			rc.DrawGlowDot(headAnchor.X, headAnchor.Y, eyeR*1.00, th.eyeRingCol, 0.95*wEye, 0.94)
			rc.DrawGlowDot(headAnchor.X, headAnchor.Y, eyeR*0.48, darkInk, 0.96*wEye, 0.95)
			rc.DrawGlowDot(headAnchor.X-0.26*eyeR, headAnchor.Y-0.26*eyeR, eyeR*0.24, brightIvory, 0.92*wEye, 0.90)
		}

		rc.PopClip()
	}

	// Pass 2: Crisp Master Woodcut Engraving Outlines along all shared square boundaries
	for idx := range cells {
		cell := &cells[idx]
		for k := 0; k < 4; k++ {
			ePts := cell.edges[k]
			eNormT := common.Constrain(cell.edgeMorph[k]/math.Max(strength, 0.01), 0.0, 1.0)
			masterW := (1.20 + 0.50*escherCubicSmoothstep(0.0, 1.0, eNormT)) * scale
			if k < 2 {
				rc.StrokePolyline(ePts, masterW, darkInk, 0.92)
			} else {
				rc.StrokePolyline(ePts, masterW*0.62, darkInk, 0.52)
			}
		}
	}

	rc.PopClip()

	// Step 3: Crisp Double-Ruled Museum Print Mat & Border Frame
	topMat := []render.Point{render.Pt(0, 0), render.Pt(fw, 0), render.Pt(fw, margin), render.Pt(0, margin)}
	botMat := []render.Point{render.Pt(0, fh-margin), render.Pt(fw, fh-margin), render.Pt(fw, fh), render.Pt(0, fh)}
	leftMat := []render.Point{render.Pt(0, margin), render.Pt(margin, margin), render.Pt(margin, fh-margin), render.Pt(0, fh-margin)}
	rightMat := []render.Point{render.Pt(fw-margin, margin), render.Pt(fw, margin), render.Pt(fw, fh-margin), render.Pt(fw-margin, fh-margin)}
	rc.FillPolygon(topMat, ivoryParchment, 1.0)
	rc.FillPolygon(botMat, ivoryParchment, 1.0)
	rc.FillPolygon(leftMat, ivoryParchment, 1.0)
	rc.FillPolygon(rightMat, ivoryParchment, 1.0)

	outerMargin := margin - 4.5*scale
	outerRule := []render.Point{
		render.Pt(outerMargin, outerMargin),
		render.Pt(fw-outerMargin, outerMargin),
		render.Pt(fw-outerMargin, fh-outerMargin),
		render.Pt(outerMargin, fh-outerMargin),
		render.Pt(outerMargin, outerMargin),
	}
	innerRule := []render.Point{
		render.Pt(margin, margin),
		render.Pt(fw-margin, margin),
		render.Pt(fw-margin, fh-margin),
		render.Pt(margin, fh-margin),
		render.Pt(margin, margin),
	}
	rc.StrokePolyline(outerRule, 2.1*scale, darkInk, 0.92)
	rc.StrokePolyline(innerRule, 1.2*scale, antiqueBrass, 0.85)
	rc.StrokePolyline(innerRule, 0.75*scale, darkInk, 0.88)

	corners := [4]render.Point{
		render.Pt(outerMargin, outerMargin),
		render.Pt(fw-outerMargin, outerMargin),
		render.Pt(fw-outerMargin, fh-outerMargin),
		render.Pt(outerMargin, fh-outerMargin),
	}
	dRad := 2.6 * scale
	for _, cp := range corners {
		diamond := []render.Point{
			render.Pt(cp.X, cp.Y-dRad),
			render.Pt(cp.X+dRad, cp.Y),
			render.Pt(cp.X, cp.Y+dRad),
			render.Pt(cp.X-dRad, cp.Y),
		}
		rc.FillPolygon(diamond, terracottaRed, 0.95)
		rc.StrokePolyline(append(diamond, diamond[0]), 0.75*scale, darkInk, 0.92)
	}

	rc.ApplyPaperTexture(0.12, seed)
	c.MergeRenderCanvas(rc)
}
