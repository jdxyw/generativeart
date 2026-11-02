package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type escherMetamorphosis struct {
	cols, rows    int
	morphStrength float64
}

// NewEscherMetamorphosis returns an EscherMetamorphosis generator that renders an
// M.C. Escher Metamorphosis / Sky and Water / Day and Night isohedral hexagonal
// tessellation. Across the honeycomb lattice, tiles continuously transform with a
// 0.00px zero-gap guarantee and an ultra-smooth multi-stage progression from 3D
// isometric tumbling Necker cubes at the outer perimeter through geometric origami
// proto-feathers into interlocking winged Escher birds in the central diagonal band.
func NewEscherMetamorphosis(cols, rows int, morphStrength float64) *escherMetamorphosis {
	if cols <= 0 {
		cols = 10
	}
	if rows <= 0 {
		rows = 10
	}
	if morphStrength <= 0 {
		morphStrength = 1.0
	}
	return &escherMetamorphosis{
		cols:          cols,
		rows:          rows,
		morphStrength: morphStrength,
	}
}

// Generative renders a static EscherMetamorphosis woodcut/lithograph print.
func (em *escherMetamorphosis) Generative(c *generativeart.Canva) {
	em.GenerativeFrame(c, 0, 1, 0)
}

// GenerativeFrame renders frame frameIdx of a seamless looping EscherMetamorphosis animation.
func (em *escherMetamorphosis) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
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
		seed = 20262004
	}

	cols := em.cols
	if cols < 4 {
		cols = 10
	}
	rows := em.rows
	if rows < 4 {
		rows = 10
	}
	// Use a slightly finer effective grid (e.g., 12x12 for default 10x10) so there are
	// more intermediate transitional tiles across the metamorphosis progression.
	effCols := cols
	effRows := rows
	if cols <= 10 && rows <= 10 {
		effCols = cols + 2
		effRows = rows + 2
	}

	strength := em.morphStrength
	if strength <= 0 {
		strength = 1.0
	}

	isAnim := totalFrames > 1
	phi := 2.0 * math.Pi * phase

	// Curated 3-Chromatic Escher Woodcut & Lithograph Palette:
	// Family 0: Warm Museum Ivory Parchment (#faf3e3) with dark ink engraving lines (#181c26)
	// Family 1: Cinnabar Vermilion / Terracotta (#d4422e) with burnished gold highlights (#f6be44)
	// Family 2: Deep Prussian Midnight Blue (#142c58) with celadon-turquoise accents (#32b4a8)
	ivoryParchment := color.RGBA{R: 250, G: 243, B: 227, A: 255} // #faf3e3
	brightIvory := color.RGBA{R: 255, G: 252, B: 244, A: 255}    // #fffcf4
	darkInk := color.RGBA{R: 24, G: 28, B: 38, A: 255}           // #181c26
	cinnabarRed := color.RGBA{R: 212, G: 66, B: 46, A: 255}      // #d4422e
	burnishedGold := color.RGBA{R: 246, G: 190, B: 68, A: 255}   // #f6be44
	prussianBlue := color.RGBA{R: 20, G: 44, B: 88, A: 255}      // #142c58
	celadonAccent := color.RGBA{R: 50, G: 180, B: 168, A: 255}   // #32b4a8
	antiqueBrass := color.RGBA{R: 186, G: 142, B: 64, A: 255}    // #ba8e40

	type tileTheme struct {
		baseCol      color.RGBA
		cubeTopCol   color.RGBA
		cubeLeftCol  color.RGBA
		cubeRightCol color.RGBA
		birdTopCol   color.RGBA
		birdLeftCol  color.RGBA
		birdRightCol color.RGBA
		lineCol      color.RGBA
		accentCol    color.RGBA
		bevelCol     color.RGBA
		eyeRingCol   color.RGBA
	}

	themes := [3]tileTheme{
		// Family 0: Warm Museum Ivory Parchment (#faf3e3) + Dark Ink (#181c26)
		{
			baseCol:      ivoryParchment,
			cubeTopCol:   brightIvory,
			cubeLeftCol:  render.LerpOklab(ivoryParchment, darkInk, 0.22),
			cubeRightCol: render.LerpOklab(ivoryParchment, darkInk, 0.46),
			birdTopCol:   render.LerpOklab(ivoryParchment, brightIvory, 0.45),
			birdLeftCol:  ivoryParchment,
			birdRightCol: render.LerpOklab(ivoryParchment, darkInk, 0.11),
			lineCol:      darkInk,
			accentCol:    render.LerpOklab(darkInk, cinnabarRed, 0.36),
			bevelCol:     brightIvory,
			eyeRingCol:   burnishedGold,
		},
		// Family 1: Cinnabar Vermilion / Terracotta (#d4422e) + Burnished Gold (#f6be44)
		{
			baseCol:      cinnabarRed,
			cubeTopCol:   render.LerpOklab(cinnabarRed, ivoryParchment, 0.42),
			cubeLeftCol:  render.LerpOklab(cinnabarRed, darkInk, 0.22),
			cubeRightCol: render.LerpOklab(cinnabarRed, darkInk, 0.48),
			birdTopCol:   render.LerpOklab(cinnabarRed, burnishedGold, 0.18),
			birdLeftCol:  cinnabarRed,
			birdRightCol: render.LerpOklab(cinnabarRed, darkInk, 0.12),
			lineCol:      darkInk,
			accentCol:    burnishedGold,
			bevelCol:     render.LerpOklab(cinnabarRed, burnishedGold, 0.62),
			eyeRingCol:   ivoryParchment,
		},
		// Family 2: Deep Prussian Midnight Blue (#142c58) + Celadon-Turquoise (#32b4a8)
		{
			baseCol:      prussianBlue,
			cubeTopCol:   render.LerpOklab(prussianBlue, celadonAccent, 0.44),
			cubeLeftCol:  render.LerpOklab(prussianBlue, celadonAccent, 0.18),
			cubeRightCol: render.LerpOklab(prussianBlue, darkInk, 0.48),
			birdTopCol:   render.LerpOklab(prussianBlue, celadonAccent, 0.18),
			birdLeftCol:  prussianBlue,
			birdRightCol: render.LerpOklab(prussianBlue, darkInk, 0.14),
			lineCol:      celadonAccent,
			accentCol:    ivoryParchment,
			bevelCol:     render.LerpOklab(prussianBlue, celadonAccent, 0.65),
			eyeRingCol:   burnishedGold,
		},
	}

	rc := c.NewRenderCanvas()
	rc.FillBackground(ivoryParchment)
	rc.SetBlendMode(render.BlendOver)

	// Museum print border geometry
	margin := 24.0 * scale
	innerW := math.Max(64.0, fw-2.0*margin)
	innerH := math.Max(64.0, fh-2.0*margin)
	cx, cy := 0.5*fw, 0.5*fh

	const sqrt3 = 1.7320508075688772
	// Choose hexagon circumradius hexR so (effCols x effRows) honeycomb cells span the print viewport
	hexRX := innerW / (1.5*float64(effCols-1) + 1.35)
	hexRY := innerH / (sqrt3*float64(effRows-1) + 1.35)
	hexR := math.Min(hexRX, hexRY)
	if hexR < 8.0 {
		hexR = 8.0
	}

	// Clip all tessellation drawing cleanly to the inner museum print frame
	framePoly := []render.Point{
		render.Pt(margin, margin),
		render.Pt(fw-margin, margin),
		render.Pt(fw-margin, fh-margin),
		render.Pt(margin, fh-margin),
	}
	rc.PushClipPolygon(framePoly)

	// Ultra-smooth, gradual spatial metamorphosis field evalMorphT(pt) in [0, 1.0 * morphStrength]:
	// Uses a gentle linear-cosine ramp across the entire viewport so ~60% of the canvas lies in
	// the intermediate transitional range t in [0.15, 0.85], and adjacent tiles differ by at most ~0.10-0.14.
	evalMorphT := func(pt render.Point) float64 {
		nx := (pt.X - cx) / (0.5 * innerW)
		ny := (pt.Y - cy) / (0.5 * innerH)

		// Diagonal wave coordinate u along (-1, 1)/sqrt(2) and cross-diagonal v along (1, 1)/sqrt(2)
		const invSqrt2 = 0.7071067811865475
		u := (nx - ny) * invSqrt2
		v := (nx + ny) * invSqrt2

		waveShift := 0.0
		breath := 0.0
		if isAnim {
			waveShift = 0.24 * math.Sin(phi)
			breath = 0.06 * math.Cos(phi)
		}

		// Wide S-curved diagonal metamorphosis corridor spanning the entire canvas
		curvedU := u - 0.18*math.Sin(1.35*v+waveShift)
		bandDist := math.Abs(curvedU) / (1.28 + breath)
		bandRamp := common.Constrain(1.0-bandDist, 0.0, 1.0)
		// Gentle blend of linear and cosine ramp (no steep quintic cliff!)
		bandEnv := 0.55*bandRamp + 0.45*(0.5-0.5*math.Cos(math.Pi*bandRamp))

		// Broad radial envelope so outer perimeter and corners reach pure 3D cube state (t -> 0.0)
		radDist := math.Hypot(nx*0.76, ny*0.76)
		radRamp := common.Constrain((1.24-radDist)/1.12, 0.0, 1.0)
		radEnv := 0.50*radRamp + 0.50*(0.5-0.5*math.Cos(math.Pi*radRamp))

		// Combine corridor and radial envelopes smoothly so:
		// - Outer corners (|u| ~ 1.2, radDist ~ 1.05): t = 0.00 (pure 3D isometric Necker cubes)
		// - Wide 5-to-6-tile transition zone: t = 0.12 -> 0.25 -> 0.38 -> 0.52 -> 0.66 -> 0.80
		// - Central diagonal wave axis (u ~ 0, radDist < 0.35): t = 1.00 * strength (full Escher birds)
		combined := bandEnv * (0.30 + 0.70*radEnv)
		norm := common.Constrain((combined-0.05)/0.86, 0.0, 1.0)
		gentle := 0.62*norm + 0.38*(0.5-0.5*math.Cos(math.Pi*norm))
		return strength * gentle
	}

	// Axial honeycomb basis for flat-top regular hexagons (vertex k at angle k * pi / 3):
	// Center(q, r) = (cx + q * 1.5 * hexR, cy + (r + 0.5 * q) * sqrt(3) * hexR)
	axialCenter := func(q, r int) render.Point {
		fq := float64(q)
		fr := float64(r)
		return render.Pt(
			cx+fq*1.5*hexR,
			cy+(fr+0.5*fq)*sqrt3*hexR,
		)
	}

	// Exact axial neighbor offsets across the 6 hexagon edges k = 0..5:
	// Edge 0 (v0->v1) is opposite Edge 3 (v3->v4)
	// Edge 1 (v1->v2) is opposite Edge 4 (v4->v5)
	// Edge 2 (v2->v3) is opposite Edge 5 (v5->v0)
	neighborDQ := [6]int{+1, 0, -1, -1, 0, +1}
	neighborDR := [6]int{0, +1, +1, 0, -1, -1}

	const samplesPerEdge = 20

	// buildSharedEdge constructs the deformed curve from vertex k to vertex (k+1)%6 of cell (q, r)
	// using the symmetric edgeMorphT = 0.5 * (morphT(A) + morphT(B)) and render.BuildIsohedralMorphTile.
	// Windowing the displacement by sin(pi * t) pins t=0 and t=1 to the exact lattice vertices while
	// preserving 100% identical world coordinates between Edge k of A and Edge (k+3)%6 of B (0.00px gap!).
	buildSharedEdge := func(center render.Point, verts [6]render.Point, k int, edgeMorphT float64) []render.Point {
		refPoly := render.BuildIsohedralMorphTile(center, hexR, 0.0, edgeMorphT, 6, samplesPerEdge)
		a := verts[k]
		b := verts[(k+1)%6]
		edgePts := make([]render.Point, samplesPerEdge+1)
		edgePts[0] = a
		baseIdx := k * samplesPerEdge
		for i := 1; i < samplesPerEdge; i++ {
			t := float64(i) / float64(samplesPerEdge)
			chordPt := a.Lerp(b, t)
			rawPt := refPoly[baseIdx+i]
			// Symmetric endpoint pin window: sin(pi*(1-t)) == sin(pi*t)
			win := math.Sin(math.Pi * t)
			edgePts[i] = chordPt.Add(rawPt.Sub(chordPt).Mul(win))
		}
		edgePts[samplesPerEdge] = b
		return edgePts
	}

	// Determine axial bounds covering the entire viewport plus margin
	maxQ := effCols/2 + 3
	maxR := effRows/2 + maxQ/2 + 3

	type cellData struct {
		q, r      int
		colorIdx  int
		center    render.Point
		morphT    float64
		normT     float64
		verts     [6]render.Point
		edges     [6][]render.Point
		boundary  []render.Point
		edgeMorph [6]float64
	}

	var cells []cellData
	for q := -maxQ; q <= maxQ; q++ {
		for r := -maxR; r <= maxR; r++ {
			cPt := axialCenter(q, r)
			if cPt.X < margin-hexR*1.8 || cPt.X > fw-margin+hexR*1.8 ||
				cPt.Y < margin-hexR*1.8 || cPt.Y > fh-margin+hexR*1.8 {
				continue
			}

			mT := evalMorphT(cPt)
			nT := common.Constrain(mT/math.Max(strength, 0.01), 0.0, 1.0)

			var verts [6]render.Point
			for k := 0; k < 6; k++ {
				ang := float64(k) * math.Pi / 3.0
				verts[k] = render.Pt(cPt.X+hexR*math.Cos(ang), cPt.Y+hexR*math.Sin(ang))
			}

			var edges [6][]render.Point
			var edgeMorph [6]float64
			boundary := make([]render.Point, 0, 6*samplesPerEdge)

			for k := 0; k < 6; k++ {
				nbPt := axialCenter(q+neighborDQ[k], r+neighborDR[k])
				nbMorphT := evalMorphT(nbPt)
				// CRITICAL ZERO-GAP FORMULA: exact symmetric average morphT on the shared edge
				eMorph := 0.5 * (mT + nbMorphT)
				edgeMorph[k] = eMorph
				ePts := buildSharedEdge(cPt, verts, k, eMorph)
				edges[k] = ePts
				boundary = append(boundary, ePts[:samplesPerEdge]...)
			}

			// Proper 3-chromatic hexagonal coloring: (q - r) mod 3 in {0, 1, 2}
			colIdx := ((q - r) % 3)
			if colIdx < 0 {
				colIdx += 3
			}

			cells = append(cells, cellData{
				q:         q,
				r:         r,
				colorIdx:  colIdx,
				center:    cPt,
				morphT:    mT,
				normT:     nT,
				verts:     verts,
				edges:     edges,
				boundary:  boundary,
				edgeMorph: edgeMorph,
			})
		}
	}

	// Pass 1: Unified 3-Chamber Partition, Continuous Oklab Shading & Multi-Stage Coons-Patch Line Morphing
	for idx := range cells {
		cell := &cells[idx]
		th := themes[cell.colorIdx]
		normT := cell.normT

		// Continuous interpolation weights across all stages in [0, 1]:
		// - smoothT: continuous color & geometry progression from 3D Necker cube (0.0) to Escher bird (1.0)
		// - wCube:   recessed isometric facet rings & parallel woodcut hatching (1.0 at t=0 -> fades out by t=0.78)
		// - wMid:    transitional origami chevron & proto-feather ribs peaking across t in [0.15, 0.85]
		// - wBird:   organic wing pinions, scalloped breast plumage & tail volutes emerging over t in [0.22, 0.94]
		// - wEye:    progressive diamond-to-circular golden bird eye emerging over t in [0.40, 0.95]
		smoothT := 0.45*normT + 0.55*escherCubicSmoothstep(0.0, 1.0, normT)
		wCube := 1.0 - escherCubicSmoothstep(0.14, 0.78, normT)
		wMid := math.Sin(math.Pi * common.Constrain((normT-0.08)/0.86, 0.0, 1.0))
		wBird := escherCubicSmoothstep(0.22, 0.92, normT)
		wEye := escherCubicSmoothstep(0.40, 0.95, normT)

		// 1A. Base anti-seam fill of the full deformed hexagon
		rc.FillPolygon(cell.boundary, th.baseCol, 1.0)
		closedBoundary := append(append([]render.Point(nil), cell.boundary...), cell.boundary[0])
		rc.StrokePolyline(closedBoundary, 0.85*scale, th.baseCol, 1.0)

		// Central hub shifts smoothly from exact hexagon center (at t=0) toward the upper-right flight axis
		flightDir := render.Pt(math.Cos(-math.Pi/3.0), math.Sin(-math.Pi/3.0))
		hub := cell.center.Add(flightDir.Mul(0.11 * hexR * smoothT))

		// Build the 3 internal partition curves from hub to alternating vertices v[1], v[3], v[5].
		// At t = 0, these are exact straight radial segments forming the 3 rhombic faces of an isometric 3D cube.
		// As t increases from 0 to 1, they bend smoothly into the bird's wing shoulders and tail spine!
		buildInternalSeam := func(targetVert render.Point, bendSign, amp float64) []render.Point {
			chord := targetVert.Sub(hub)
			cLen := chord.Len()
			tan := chord.Mul(1.0 / math.Max(cLen, 1e-6))
			norm := tan.Perp()
			pts := make([]render.Point, samplesPerEdge+1)
			for i := 0; i <= samplesPerEdge; i++ {
				t := float64(i) / float64(samplesPerEdge)
				env := math.Sin(math.Pi * t)
				dn := bendSign * amp * cell.morphT * (0.22*math.Sin(2.0*math.Pi*t-0.35) + 0.08*env)
				du := 0.05 * cell.morphT * math.Sin(2.0*math.Pi*t) * env
				pts[i] = hub.Lerp(targetVert, t).
					Add(tan.Mul(du * cLen)).
					Add(norm.Mul(dn * cLen))
			}
			return pts
		}

		seam1 := buildInternalSeam(cell.verts[1], +1.0, 0.85) // Lower-right wing/tail divider (hub -> v1)
		seam3 := buildInternalSeam(cell.verts[3], -1.0, 0.85) // Left wing/tail divider (hub -> v3)
		seam5 := buildInternalSeam(cell.verts[5], +0.7, 0.65) // Upper head/crest divider (hub -> v5)

		// 1B. Unified 3-Chamber Topology across ALL t in [0, 1] (zero sudden jumps!):
		// - topChamber   (bounded by seam3, edge3, edge4, seam5): sunlit top cube facet -> upper head & left-upper wing
		// - leftChamber  (bounded by seam1, edge1, edge2, seam3): medium-shaded left cube facet -> lower tail & breast fan
		// - rightChamber (bounded by seam5, edge5, edge0, seam1): deep-shaded right cube facet -> right wing lobe
		buildChamberPoly := func(sA, eA, eB, sB []render.Point) []render.Point {
			poly := make([]render.Point, 0, len(sA)+len(eA)+len(eB)+len(sB))
			poly = append(poly, sA...)
			poly = append(poly, eA[1:samplesPerEdge]...)
			poly = append(poly, eB[:samplesPerEdge]...)
			for i := len(sB) - 1; i >= 1; i-- {
				poly = append(poly, sB[i])
			}
			return poly
		}

		topChamber := buildChamberPoly(seam3, cell.edges[3], cell.edges[4], seam5)
		leftChamber := buildChamberPoly(seam1, cell.edges[1], cell.edges[2], seam3)
		rightChamber := buildChamberPoly(seam5, cell.edges[5], cell.edges[0], seam1)

		// Continuous Oklab shading evolution from crisp 3D isometric cube lighting (t=0) to sculpted bird relief (t=1)
		topCol := render.LerpOklab(th.cubeTopCol, th.birdTopCol, smoothT)
		leftCol := render.LerpOklab(th.cubeLeftCol, th.birdLeftCol, smoothT)
		rightCol := render.LerpOklab(th.cubeRightCol, th.birdRightCol, smoothT)

		rc.FillPolygon(topChamber, topCol, 1.0)
		rc.FillPolygon(leftChamber, leftCol, 1.0)
		rc.FillPolygon(rightChamber, rightCol, 1.0)

		// Anti-seam hairline along the 3 internal seams so adjacent chambers fuse cleanly
		rc.StrokePolyline(seam3, 0.65*scale, topCol, 0.55)
		rc.StrokePolyline(seam1, 0.65*scale, leftCol, 0.55)
		rc.StrokePolyline(seam5, 0.65*scale, rightCol, 0.55)

		// Coons-Patch / Transfinite Bilinear Interpolation helper inside any curved 4-sided chamber:
		// Maps unit square (u, v) in [0, 1]^2 smoothly onto the curved chamber bounded by:
		//   v = 0: seamA(u)   (from hub at u=0 to cornerA at u=1)
		//   u = 0: seamB(v)   (from hub at v=0 to cornerB at v=1)
		//   u = 1: edgeA(v)   (from cornerA at v=0 to outerCorner at v=1)
		//   v = 1: edgeB(1-u) (from cornerB at u=0 to outerCorner at u=1)
		// At t = 0 (straight edges), this is exact bilinear interpolation on the isometric cube rhombus!
		// As t increases, any geometric line or curve in (u, v) bends and warps continuously with the chamber!
		evalCurve := func(pts []render.Point, s float64) render.Point {
			s = common.Constrain(s, 0.0, 1.0)
			fIdx := s * float64(len(pts)-1)
			i0 := int(math.Floor(fIdx))
			i1 := i0 + 1
			if i1 >= len(pts) {
				return pts[len(pts)-1]
			}
			frac := fIdx - float64(i0)
			return pts[i0].Lerp(pts[i1], frac)
		}

		chamberPoint := func(sA, eA, eB, sB []render.Point, u, v float64) render.Point {
			u = common.Constrain(u, 0.0, 1.0)
			v = common.Constrain(v, 0.0, 1.0)
			c00 := sA[0]                  // hub (0, 0)
			c10 := sA[len(sA)-1]          // cornerA (1, 0)
			c11 := eA[len(eA)-1]          // outerCorner (1, 1)
			c01 := sB[len(sB)-1]          // cornerB (0, 1)
			cU0 := evalCurve(sA, u)       // bottom boundary (u, 0)
			cU1 := evalCurve(eB, 1.0-u)   // top boundary (u, 1)
			c0V := evalCurve(sB, v)       // left boundary (0, v)
			c1V := evalCurve(eA, v)       // right boundary (1, v)

			// Coons patch = ruled(u) + ruled(v) - bilinear(corners)
			ruledU := cU0.Mul(1.0 - v).Add(cU1.Mul(v))
			ruledV := c0V.Mul(1.0 - u).Add(c1V.Mul(u))
			bilin := c00.Mul((1.0 - u) * (1.0 - v)).
				Add(c10.Mul(u * (1.0 - v))).
				Add(c01.Mul((1.0 - u) * v)).
				Add(c11.Mul(u * v))
			return ruledU.Add(ruledV).Sub(bilin)
		}

		traceChamberSegment := func(sA, eA, eB, sB []render.Point, u0, v0, u1, v1 float64, steps int) []render.Point {
			pts := make([]render.Point, steps+1)
			for i := 0; i <= steps; i++ {
				f := float64(i) / float64(steps)
				pts[i] = chamberPoint(sA, eA, eB, sB, u0+(u1-u0)*f, v0+(v1-v0)*f)
			}
			return pts
		}

		// 1C. Stage 1: Warped Isometric Cube Facet Steps & Woodcut Engraving Hatching (wCube > 0.01)
		// Because every line is mapped through chamberPoint(u, v), the cube's recessed architectural steps
		// and parallel hatch lines physically bend along with the curved edges as the cube unfolds!
		if wCube > 0.01 {
			drawWarpedCubeChamber := func(sA, eA, eB, sB []render.Point, shadeDepth float64, crossHatch bool, ringAlpha float64) {
				// Two concentric recessed sub-rhombus rings that warp smoothly with the chamber
				rings := [2][2]float64{{0.18, 0.82}, {0.34, 0.66}}
				for rIdx, rg := range rings {
					lo, hi := rg[0], rg[1]
					ringAlphaR := ringAlpha * wCube * (1.0 - 0.28*float64(rIdx))
					ringCol := th.accentCol
					if rIdx == 1 {
						ringCol = th.lineCol
					}
					ringPts := make([]render.Point, 0, 33)
					ringPts = append(ringPts, traceChamberSegment(sA, eA, eB, sB, lo, lo, hi, lo, 8)...)
					ringPts = append(ringPts, traceChamberSegment(sA, eA, eB, sB, hi, lo, hi, hi, 8)[1:]...)
					ringPts = append(ringPts, traceChamberSegment(sA, eA, eB, sB, hi, hi, lo, hi, 8)[1:]...)
					ringPts = append(ringPts, traceChamberSegment(sA, eA, eB, sB, lo, hi, lo, lo, 8)[1:]...)
					rc.StrokePolyline(ringPts, (0.72-0.14*float64(rIdx))*scale, ringCol, ringAlphaR)
				}

				// Parallel woodcut hatch lines mapped through the Coons patch
				if shadeDepth > 0.12 {
					nHatch := int(math.Round(hexR / math.Max(2.4, 4.0*scale)))
					if nHatch < 4 {
						nHatch = 4
					}
					if nHatch > 11 {
						nHatch = 11
					}
					hatchW := (0.48 + 0.52*shadeDepth) * scale
					hatchAlpha := common.Constrain(0.24+0.54*shadeDepth, 0.15, 0.82) * wCube
					for i := 1; i < nHatch; i++ {
						u := float64(i) / float64(nHatch)
						// Gradually bow the hatch line toward a proto-barb curve as normT increases
						bow := 0.08 * smoothT
						hPts := make([]render.Point, 9)
						for j := 0; j <= 8; j++ {
							v := float64(j) / 8.0
							uWarped := common.Constrain(u+bow*math.Sin(math.Pi*v), 0.04, 0.96)
							hPts[j] = chamberPoint(sA, eA, eB, sB, uWarped, v)
						}
						rc.StrokePolyline(hPts, hatchW, th.lineCol, hatchAlpha)
					}

					// Cross-hatching on the deepest-shaded right cube facet
					if crossHatch && shadeDepth > 0.42 {
						nCross := int(math.Round(float64(nHatch) * 0.78))
						if nCross < 3 {
							nCross = 3
						}
						crossW := (0.42 + 0.38*shadeDepth) * scale
						crossAlpha := common.Constrain((shadeDepth-0.30)*0.68, 0.12, 0.58) * wCube * (1.0 - 0.35*smoothT)
						for j := 1; j < nCross; j++ {
							v := float64(j) / float64(nCross)
							cPts := traceChamberSegment(sA, eA, eB, sB, 0.05, v, 0.95, v, 8)
							rc.StrokePolyline(cPts, crossW, th.lineCol, crossAlpha)
						}
					}
				}
			}

			// Top sunlit facet (light warm engraving), Left medium-shaded facet, Right deep-shaded cross-hatched facet
			drawWarpedCubeChamber(seam3, cell.edges[3], cell.edges[4], seam5, 0.18, false, 0.62)
			drawWarpedCubeChamber(seam1, cell.edges[1], cell.edges[2], seam3, 0.46, false, 0.54)
			drawWarpedCubeChamber(seam5, cell.edges[5], cell.edges[0], seam1, 0.74, true, 0.58)
		}

		// 1D. Stage 2: Transitional Origami & Geometric Chevron Ribs (wMid > 0.02, peaking across normT in [0.15, 0.85])
		// Bridges the visual evolution from parallel isometric cube hatching to V-shaped proto-feathers!
		if wMid > 0.02 {
			drawChevronProtoFeathers := func(sA, eA, eB, sB []render.Point, numChevrons int, alphaScale float64) {
				for cIdx := 1; cIdx <= numChevrons; cIdx++ {
					f := float64(cIdx) / float64(numChevrons+1)
					// Chevron V-rib anchored along the diagonal (f, f) inside the curved chamber
					span := 0.20 + 0.55*f
					pLeft := chamberPoint(sA, eA, eB, sB, common.Constrain(f+span*0.48, 0.08, 0.92), common.Constrain(f-span*0.28, 0.06, 0.92))
					pApex := chamberPoint(sA, eA, eB, sB, common.Constrain(f+0.09*smoothT, 0.08, 0.94), common.Constrain(f+0.09*smoothT, 0.08, 0.94))
					pRight := chamberPoint(sA, eA, eB, sB, common.Constrain(f-span*0.28, 0.06, 0.92), common.Constrain(f+span*0.48, 0.08, 0.92))

					// Blend between crisp origami 3-point折线 (early transition) and smooth spline (late transition)
					var ribPts []render.Point
					if normT < 0.45 {
						midL := pLeft.Lerp(pApex, 0.5)
						midR := pApex.Lerp(pRight, 0.5)
						ribPts = []render.Point{pLeft, midL, pApex, midR, pRight}
					} else {
						ribPts = render.CatmullRomSpline([]render.Point{pLeft, pApex, pRight}, 6)
					}
					ribW := (0.62 + 0.28*(1.0-f)) * scale
					rc.StrokePolyline(ribPts, ribW, th.lineCol, 0.52*wMid*alphaScale)
					if cIdx%2 == 1 {
						rc.StrokePolyline(ribPts, 0.48*scale, th.accentCol, 0.38*wMid*alphaScale)
					}
				}
			}

			// Draw origami chevron ribs in the two wing chambers and the tail chamber
			drawChevronProtoFeathers(seam3, cell.edges[3], cell.edges[4], seam5, 4, 0.92)
			drawChevronProtoFeathers(seam5, cell.edges[5], cell.edges[0], seam1, 4, 1.00)
			drawChevronProtoFeathers(seam1, cell.edges[1], cell.edges[2], seam3, 3, 0.85)
		}

		// 1E. Stage 3: Organic Escher Wing Pinions, Scalloped Plumage, Tail Volutes & Progressive Eye (wBird > 0.01)
		if wBird > 0.01 {
			// 3D Inner Bevel Highlight along the curved winged boundary
			bevelPts := make([]render.Point, 0, len(cell.boundary)+1)
			insetBevel := 0.082
			lightVec := render.Pt(-0.55, -0.83)
			for i := 0; i < len(cell.boundary); i++ {
				p := cell.boundary[i]
				inward := cell.center.Sub(p)
				bevelPts = append(bevelPts, p.Add(inward.Mul(insetBevel)))
			}
			bevelPts = append(bevelPts, bevelPts[0])
			rc.StrokePolyline(bevelPts, (1.05+0.65*wBird)*scale, th.bevelCol, 0.48*wBird)

			// Directional highlight arc along the sun-facing upper wing edges (edges 3, 4, 5)
			for _, ek := range []int{3, 4, 5} {
				eCurve := cell.edges[ek]
				litRibbon := make([]render.Point, len(eCurve))
				for i, p := range eCurve {
					inward := cell.center.Sub(p).Norm()
					dot := common.Constrain(-(inward.X*lightVec.X + inward.Y*lightVec.Y), 0.0, 1.0)
					offset := (1.35 + 1.15*dot) * scale
					litRibbon[i] = p.Add(inward.Mul(offset))
				}
				rc.StrokePolyline(litRibbon, 1.18*scale, th.accentCol, 0.58*wBird)
			}

			// Engraved Wing Feathers (Primary & Secondary Pinions) on the two lateral wing chambers.
			// Both their span (growing outward from the spine toward the wing tip) and curvature scale with wBird!
			drawWingPinions := func(sA, eA, eB, sB []render.Point, maxFeathers int, flip float64) {
				for f := 1; f <= maxFeathers; f++ {
					u := float64(f) / float64(maxFeathers+1)
					// Feather staggering: primary feathers emerge first, outer pinions join smoothly as wBird grows
					featherThreshold := 0.14 * float64(f-1) / float64(maxFeathers)
					fWeight := escherCubicSmoothstep(featherThreshold, math.Min(1.0, featherThreshold+0.65), wBird)
					if fWeight <= 0.01 {
						continue
					}

					rootPt := evalCurve(sA, u)
					var edgeTarget render.Point
					if u < 0.52 {
						edgeTarget = evalCurve(eA, u/0.52)
					} else {
						edgeTarget = evalCurve(eB, (u-0.52)/0.48)
					}

					// Feather length grows progressively from 45% to 90% of the chamber span as fWeight increases
					reach := 0.45 + 0.45*fWeight
					tipIn := rootPt.Lerp(edgeTarget.Lerp(cell.center, 0.10), reach)
					mid := rootPt.Lerp(tipIn, 0.52)
					chord := tipIn.Sub(rootPt)
					norm := chord.Perp().Norm()
					archPt := mid.Add(norm.Mul(flip * 0.15 * chord.Len() * fWeight))

					featherCurve := render.CatmullRomSpline([]render.Point{rootPt, archPt, tipIn}, 8)
					wFeather := (0.64 + 0.42*(1.0-u)) * scale
					rc.StrokePolyline(featherCurve, wFeather, th.lineCol, (0.54+0.30*(1.0-u))*fWeight)

					// Secondary woodcut hatch barb & metallic accent along alternating pinions
					if f%2 == 1 {
						accentCurve := render.CatmullRomSpline([]render.Point{
							rootPt.Lerp(archPt, 0.25),
							archPt.Add(norm.Mul(flip * 0.04 * chord.Len())),
							tipIn.Lerp(rootPt, 0.15),
						}, 6)
						rc.StrokePolyline(accentCurve, 0.56*scale, th.accentCol, 0.48*fWeight)
					}
				}
			}

			drawWingPinions(seam5, cell.edges[5], cell.edges[0], seam1, 6, +1.0)
			drawWingPinions(seam3, cell.edges[3], cell.edges[4], seam5, 6, -1.0)

			// Tail Volutes & Scalloped Plumage Fan in the lower chamber (seam1, edge1, edge2, seam3)
			const tailRays = 5
			for tr := 1; tr <= tailRays; tr++ {
				v := float64(tr) / float64(tailRays+1)
				var edgeTarget render.Point
				if v <= 0.5 {
					edgeTarget = evalCurve(cell.edges[1], v*2.0)
				} else {
					edgeTarget = evalCurve(cell.edges[2], (v-0.5)*2.0)
				}
				reach := 0.48 + 0.38*wBird
				tailEnd := hub.Lerp(edgeTarget, reach)
				chord := tailEnd.Sub(hub)
				perp := chord.Perp().Norm()
				swirl := (v - 0.5) * 0.32 * chord.Len() * wBird
				ctrl1 := hub.Lerp(tailEnd, 0.36).Add(perp.Mul(swirl))
				ctrl2 := hub.Lerp(tailEnd, 0.74).Sub(perp.Mul(swirl * 0.65))
				volute := render.CatmullRomSpline([]render.Point{hub, ctrl1, ctrl2, tailEnd}, 8)
				rc.StrokePolyline(volute, 0.70*scale, th.lineCol, 0.65*wBird)
			}

			// Scalloped woodcut cross-chevron bands across the bird's body/breast
			for ringIdx := 1; ringIdx <= 3; ringIdx++ {
				rf := 0.26 + 0.19*float64(ringIdx)
				pL := evalCurve(seam3, rf)
				pM := chamberPoint(seam1, cell.edges[1], cell.edges[2], seam3, rf*0.76, rf*0.76)
				pR := evalCurve(seam1, rf)
				scallop := render.CatmullRomSpline([]render.Point{pL, pM, pR}, 8)
				rc.StrokePolyline(scallop, 0.64*scale, th.accentCol, 0.52*wBird)
			}
		}

		// 1F. Continuous Internal Spine / Cube Ridge Seams across all t in [0, 1]
		// At t = 0, these are the 3 crisp interior edges of the isometric Necker cube meeting at hub.
		// At t = 1, they are the sculpted anatomical wing/tail seams of the Escher bird!
		seamW := (1.05 + 0.35*smoothT) * scale
		seamAlpha := 0.82 + 0.10*smoothT
		rc.StrokePolyline(seam1, seamW, th.lineCol, seamAlpha)
		rc.StrokePolyline(seam3, seamW, th.lineCol, seamAlpha)
		rc.StrokePolyline(seam5, seamW*0.92, th.lineCol, seamAlpha*0.92)

		// 1G. Progressive Eye & Crest Emergence (wEye > 0.01, over normT in [0.40, 0.95]):
		// Smoothly morphs from a geometric isometric lozenge/diamond into the full golden Escher bird eye!
		if wEye > 0.01 {
			headAnchor := chamberPoint(seam3, cell.edges[3], cell.edges[4], seam5, 0.36, 0.62)
			eyeR := (1.35 + 2.05*wEye) * scale

			// Geometric diamond proto-eye in the mid-transition zone (fades out as circular iris takes over)
			diamondAlpha := math.Sin(math.Pi*wEye) * 0.78
			if diamondAlpha > 0.04 {
				dR := eyeR * 1.25
				lozenge := []render.Point{
					render.Pt(headAnchor.X, headAnchor.Y-dR*0.85),
					render.Pt(headAnchor.X+dR*1.15, headAnchor.Y),
					render.Pt(headAnchor.X, headAnchor.Y+dR*0.85),
					render.Pt(headAnchor.X-dR*1.15, headAnchor.Y),
				}
				rc.FillPolygon(lozenge, th.eyeRingCol, diamondAlpha*0.75)
				rc.StrokePolyline(append(lozenge, lozenge[0]), 0.72*scale, th.lineCol, diamondAlpha)
			}

			// Full circular golden Escher bird eye with engraved halo, pupil, and specular glint
			roundEyeWeight := escherCubicSmoothstep(0.18, 0.95, wEye)
			if roundEyeWeight > 0.02 {
				rc.DrawGlowDot(headAnchor.X, headAnchor.Y, eyeR*1.42, th.lineCol, 0.86*roundEyeWeight, 0.92)
				rc.DrawGlowDot(headAnchor.X, headAnchor.Y, eyeR*1.00, th.eyeRingCol, 0.94*roundEyeWeight, 0.94)
				rc.DrawGlowDot(headAnchor.X, headAnchor.Y, eyeR*0.50, darkInk, 0.96*roundEyeWeight, 0.95)
				rc.DrawGlowDot(headAnchor.X-0.28*eyeR, headAnchor.Y-0.28*eyeR, eyeR*0.24, ivoryParchment, 0.92*roundEyeWeight, 0.90)
			}
		}
	}

	// Pass 2: Crisp Master Woodcut Engraving Outlines along all shared isohedral boundaries
	for idx := range cells {
		cell := &cells[idx]
		// Draw each of the 6 edges; to avoid double-stroking interior shared edges excessively,
		// stroke edges 0, 1, 2 at full master weight and edges 3, 4, 5 as a fine reinforcing pass.
		for k := 0; k < 6; k++ {
			ePts := cell.edges[k]
			eNormT := common.Constrain(cell.edgeMorph[k]/math.Max(strength, 0.01), 0.0, 1.0)
			masterW := (1.22 + 0.52*escherCubicSmoothstep(0.0, 1.0, eNormT)) * scale
			if k < 3 {
				rc.StrokePolyline(ePts, masterW, darkInk, 0.92)
			} else {
				rc.StrokePolyline(ePts, masterW*0.62, darkInk, 0.52)
			}
		}
	}

	// Pop the inner viewport clip mask before drawing the double-ruled museum print border
	rc.PopClip()

	// Step 4: Crisp Double-Ruled Museum Print Mat & Border Frame
	// Mask the outer 24px*scale margin cleanly in warm museum parchment (#faf3e3) so any edge antialiasing is razor-sharp
	topMat := []render.Point{render.Pt(0, 0), render.Pt(fw, 0), render.Pt(fw, margin), render.Pt(0, margin)}
	botMat := []render.Point{render.Pt(0, fh-margin), render.Pt(fw, fh-margin), render.Pt(fw, fh), render.Pt(0, fh)}
	leftMat := []render.Point{render.Pt(0, margin), render.Pt(margin, margin), render.Pt(margin, fh-margin), render.Pt(0, fh-margin)}
	rightMat := []render.Point{render.Pt(fw-margin, margin), render.Pt(fw, margin), render.Pt(fw, fh-margin), render.Pt(fw-margin, fh-margin)}
	rc.FillPolygon(topMat, ivoryParchment, 1.0)
	rc.FillPolygon(botMat, ivoryParchment, 1.0)
	rc.FillPolygon(leftMat, ivoryParchment, 1.0)
	rc.FillPolygon(rightMat, ivoryParchment, 1.0)

	// Outer heavy woodcut rule and inner fine gold/ink rule
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

	// Corner museum print registration diamonds
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
		rc.FillPolygon(diamond, cinnabarRed, 0.95)
		rc.StrokePolyline(append(diamond, diamond[0]), 0.75*scale, darkInk, 0.92)
	}

	// Post-processing: authentic cotton paper fiber modulation
	rc.ApplyPaperTexture(0.12, seed)
	c.MergeRenderCanvas(rc)
}

func escherCubicSmoothstep(edge0, edge1, x float64) float64 {
	if edge1 <= edge0 {
		if x >= edge1 {
			return 1.0
		}
		return 0.0
	}
	t := common.Constrain((x-edge0)/(edge1-edge0), 0.0, 1.0)
	return t * t * (3.0 - 2.0*t)
}
