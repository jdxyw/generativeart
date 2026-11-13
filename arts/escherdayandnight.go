package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type escherDayAndNight struct {
	cols, rows    int
	morphStrength float64
}

// NewEscherDayAndNight returns an escherDayAndNight engine inspired by M.C. Escher's
// Day and Night (1938) and Sky and Water: across a diamond checkerboard lattice,
// lower terraced farmland fields continuously morph with a 0.00px figure-ground dual
// shared-edge guarantee into counter-flying White/Gold Day-Swans (soaring rightward)
// and Prussian-Blue/Cinnabar Night-Swans (soaring leftward).
func NewEscherDayAndNight(cols, rows int, morphStrength float64) *escherDayAndNight {
	if cols <= 0 {
		cols = 10
	}
	if rows <= 0 {
		rows = 10
	}
	if morphStrength <= 0 {
		morphStrength = 1.0
	}
	return &escherDayAndNight{
		cols:          cols,
		rows:          rows,
		morphStrength: morphStrength,
	}
}

// Generative renders a static frame of the Escher Day and Night woodcut print.
func (edn *escherDayAndNight) Generative(c *generativeart.Canva) {
	edn.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360-degree phase loop (phase in [0, 1)).
func (edn *escherDayAndNight) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
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
		seed = 20262902
	}

	cols := common.MaxInt(6, edn.cols)
	rows := common.MaxInt(6, edn.rows)
	if cols < 10 {
		cols = 10
	}
	if rows < 10 {
		rows = 10
	}
	strength := common.Constrain(edn.morphStrength, 0.4, 1.4)
	isAnim := totalFrames > 1
	phi := 2.0 * math.Pi * phase

	// Harmonious M.C. Escher Woodcut Palette:
	// Warm Cream Ivory (#faf3e3), Deep Pine-Teal (#164c58) & Burnished Amber-Ochre (#d67428)
	ivoryParchment := color.RGBA{R: 250, G: 243, B: 227, A: 255}
	brightIvory := color.RGBA{R: 255, G: 252, B: 244, A: 255}
	warmWheat := color.RGBA{R: 240, G: 218, B: 174, A: 255}
	darkInk := color.RGBA{R: 22, G: 26, B: 32, A: 255}
	amberOchre := color.RGBA{R: 214, G: 116, B: 40, A: 255}
	sunGold := color.RGBA{R: 248, G: 198, B: 74, A: 255}
	pineTeal := color.RGBA{R: 22, G: 76, B: 88, A: 255}
	aquaCeladon := color.RGBA{R: 72, G: 196, B: 184, A: 255}
	antiqueBrass := color.RGBA{R: 186, G: 142, B: 64, A: 255}

	rc := c.NewRenderCanvas()
	rc.FillBackground(ivoryParchment)
	rc.SetBlendMode(render.BlendOver)

	margin := 24.0 * scale
	innerW := math.Max(64.0, fw-2.0*margin)
	innerH := math.Max(64.0, fh-2.0*margin)
	cx, cy := 0.5*fw, 0.54*fh

	// Classic 10x10 perspective diamond proportions (rx/ry ~ 1.48)
	rx := innerW / float64(cols+1)
	ry := innerH / float64(rows+2)

	framePoly := []render.Point{
		render.Pt(margin, margin),
		render.Pt(fw-margin, margin),
		render.Pt(fw-margin, fh-margin),
		render.Pt(margin, fh-margin),
	}
	rc.PushClipPolygon(framePoly)

	// 1. Grid vertex G(p, q):
	gridVertex := func(p, q int) render.Point {
		fp, fq := float64(p), float64(q)
		if isAnim {
			fp -= phase
			fq -= phase
		}
		return render.Pt(
			cx+(fp-fq)*rx,
			cy+(fp+fq)*ry,
		)
	}

	// Spatial Metamorphosis Field evalMorphT(pt) strictly in [0.0, 1.0 * strength]:
	evalMorphT := func(pt render.Point) float64 {
		nx := (pt.X - cx) / (0.5 * innerW)
		ny := (pt.Y - margin) / innerH

		waveShift := 0.0
		vSweep := 0.0
		if isAnim {
			waveShift = -phi
			vSweep = 0.34 * math.Sin(phi)
		}

		archNy := (ny + vSweep) - 0.10*math.Cos(nx*math.Pi*0.85+waveShift)
		vRamp := common.Constrain((0.88-archNy)/0.66, 0.0, 1.0)
		vSmooth := 0.55*vRamp + 0.45*(0.5-0.5*math.Cos(math.Pi*vRamp))

		edgeTaper := 1.0 - 0.35*math.Pow(math.Abs(nx), 2.2)*common.Clamp01((ny-0.45)/0.55)
		return strength * common.Clamp01(vSmooth*edgeTaper)
	}

	const samplesPerEdge = 24

	// Canonical P-Segment G(p, q) -> G(p+1, q):
	evalPSegment := func(p, q int) ([]render.Point, float64) {
		a := gridVertex(p, q)
		b := gridVertex(p+1, q)
		mid := a.Lerp(b, 0.5)
		mT := evalMorphT(mid)

		chord := b.Sub(a)
		cLen := chord.Len()
		tan := chord.Mul(1.0 / math.Max(cLen, 1e-6))
		upRightNorm := render.Pt(tan.Y, -tan.X)

		pts := make([]render.Point, samplesPerEdge+1)
		pts[0] = a
		for i := 1; i < samplesPerEdge; i++ {
			t := float64(i) / float64(samplesPerEdge)
			win := math.Sin(math.Pi * t)

			wingS := 0.26*math.Sin(2.0*math.Pi*t-0.25) + 0.11*math.Sin(4.0*math.Pi*t-0.5) + 0.08*math.Sin(math.Pi*t)
			du := mT * 0.09 * math.Sin(2.0*math.Pi*t-0.6) * win * win
			dn := mT * wingS * win

			pts[i] = a.Lerp(b, t).Add(upRightNorm.Mul(dn * cLen)).Add(tan.Mul(du * cLen))
		}
		pts[samplesPerEdge] = b
		return pts, mT
	}

	// Canonical Q-Segment G(p, q) -> G(p, q+1):
	evalQSegment := func(p, q int) ([]render.Point, float64) {
		a := gridVertex(p, q)
		b := gridVertex(p, q+1)
		mid := a.Lerp(b, 0.5)
		mT := evalMorphT(mid)

		chord := b.Sub(a)
		cLen := chord.Len()
		tan := chord.Mul(1.0 / math.Max(cLen, 1e-6))
		upLeftNorm := render.Pt(-tan.Y, tan.X)

		pts := make([]render.Point, samplesPerEdge+1)
		pts[0] = a
		for i := 1; i < samplesPerEdge; i++ {
			t := float64(i) / float64(samplesPerEdge)
			win := math.Sin(math.Pi * t)

			neckS := -0.24*math.Sin(2.0*math.Pi*t+0.30) + 0.10*math.Sin(4.0*math.Pi*t) - 0.07*math.Sin(math.Pi*t)
			du := -mT * 0.08 * math.Sin(2.0*math.Pi*t) * win * win
			dn := mT * neckS * win

			pts[i] = a.Lerp(b, t).Add(upLeftNorm.Mul(dn * cLen)).Add(tan.Mul(du * cLen))
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

	// Bilinear Coons-patch inside the 4-sided diamond cell bounded by:
	// e0: V0 -> V1 (top-right), e1: V1 -> V2 (bottom-right), e2: V2 -> V3 (bottom-left), e3: V3 -> V0 (top-left)
	cellCoonsPoint := func(e0, e1, e2, e3 []render.Point, u, v float64) render.Point {
		u = common.Constrain(u, 0.0, 1.0)
		v = common.Constrain(v, 0.0, 1.0)
		c00 := e0[0]
		c10 := e1[0]
		c11 := e2[0]
		c01 := e3[0]
		cU0 := evalCurve(e0, u)
		cU1 := evalCurve(e2, 1.0-u)
		c0V := evalCurve(e3, 1.0-v)
		c1V := evalCurve(e1, v)

		ruledU := cU0.Mul(1.0 - v).Add(cU1.Mul(v))
		ruledV := c0V.Mul(1.0 - u).Add(c1V.Mul(u))
		bilin := c00.Mul((1.0 - u) * (1.0 - v)).
			Add(c10.Mul(u * (1.0 - v))).
			Add(c01.Mul((1.0 - u) * v)).
			Add(c11.Mul(u * v))
		return ruledU.Add(ruledV).Sub(bilin)
	}

	traceCellSegment := func(e0, e1, e2, e3 []render.Point, u0, v0, u1, v1 float64, steps int) []render.Point {
		pts := make([]render.Point, steps+1)
		for i := 0; i <= steps; i++ {
			f := float64(i) / float64(steps)
			pts[i] = cellCoonsPoint(e0, e1, e2, e3, u0+(u1-u0)*f, v0+(v1-v0)*f)
		}
		return pts
	}

	// 2. Build and Render All Diamond / Swan Cells (p, q) Across the Viewport
	maxIdx := cols + rows + 4
	type dayNightCell struct {
		p, q      int
		isDaySwan bool // true: (p+q) even -> White/Gold Right-Flying Swan; false: (p+q) odd -> Prussian/Cinnabar Left-Flying Swan
		colorFam  int
		center    render.Point
		normT     float64
		edges     [4][]render.Point
		edgeMorph [4]float64
		boundary  []render.Point
	}

	var cells []dayNightCell
	for p := -maxIdx; p <= maxIdx; p++ {
		for q := -maxIdx; q <= maxIdx; q++ {
			v0 := gridVertex(p, q)
			v1 := gridVertex(p+1, q)
			v2 := gridVertex(p+1, q+1)
			v3 := gridVertex(p, q+1)
			cPt := render.Pt(0.25*(v0.X+v1.X+v2.X+v3.X), 0.25*(v0.Y+v1.Y+v2.Y+v3.Y))

			if cPt.X < margin-rx*2.2 || cPt.X > fw-margin+rx*2.2 ||
				cPt.Y < margin-ry*2.2 || cPt.Y > fh-margin+ry*2.2 {
				continue
			}

			isEven := ((p + q) & 1) == 0

			e0, m0 := evalPSegment(p, q)
			e1, m1 := evalQSegment(p+1, q)
			e2Rev, m2 := evalPSegment(p, q+1)
			e3Rev, m3 := evalQSegment(p, q)
			e2 := reversePts(e2Rev)
			e3 := reversePts(e3Rev)

			boundary := make([]render.Point, 0, 4*samplesPerEdge)
			boundary = append(boundary, e0[:samplesPerEdge]...)
			boundary = append(boundary, e1[:samplesPerEdge]...)
			boundary = append(boundary, e2[:samplesPerEdge]...)
			boundary = append(boundary, e3[:samplesPerEdge]...)

			mT := evalMorphT(cPt)
			nT := common.Constrain(mT/math.Max(strength, 0.01), 0.0, 1.0)

			// 3-chromatic sub-tint: Even cells are Warm Ivory / Champagne Gold; Odd cells alternate Prussian Blue & Cinnabar Red!
			cFam := 0
			if !isEven {
				if ((p - q) & 2) == 0 {
					cFam = 1 // Deep Prussian Blue + Celadon
				} else {
					cFam = 2 // Cinnabar Vermilion + Gold
				}
			}

			cells = append(cells, dayNightCell{
				p:         p,
				q:         q,
				isDaySwan: isEven,
				colorFam:  cFam,
				center:    cPt,
				normT:     nT,
				edges:     [4][]render.Point{e0, e1, e2, e3},
				edgeMorph: [4]float64{m0, m1, m2, m3},
				boundary:  boundary,
			})
		}
	}

	// Pass 1: Fill & Engrave Every Farmland-to-Swan Cell
	for idx := range cells {
		cell := &cells[idx]
		normT := cell.normT
		smoothT := 0.45*normT + 0.55*escherCubicSmoothstep(0.0, 1.0, normT)
		wField := 1.0 - escherCubicSmoothstep(0.10, 0.72, normT)
		wMid := math.Sin(math.Pi * common.Constrain((normT-0.08)/0.84, 0.0, 1.0))
		wSwan := escherCubicSmoothstep(0.22, 0.92, normT)

		var baseCol, wingHiCol, bellyShadowCol, lineCol, accentCol, eyeRingCol color.RGBA
		switch cell.colorFam {
		case 0: // Cream Ivory Day-Swan & Sunlit Wheat Field
			baseCol = render.LerpOklab(warmWheat, ivoryParchment, smoothT*0.65)
			wingHiCol = brightIvory
			bellyShadowCol = render.LerpOklab(ivoryParchment, darkInk, 0.16)
			lineCol = darkInk
			accentCol = sunGold
			eyeRingCol = amberOchre
		case 1: // Deep Pine-Teal Night-Swan & Irrigated Polder Field
			baseCol = pineTeal
			wingHiCol = render.LerpOklab(pineTeal, aquaCeladon, 0.48)
			bellyShadowCol = render.LerpOklab(pineTeal, darkInk, 0.42)
			lineCol = aquaCeladon
			accentCol = ivoryParchment
			eyeRingCol = sunGold
		default: // Burnished Amber-Ochre Twilight-Swan & Harvest Field
			baseCol = amberOchre
			wingHiCol = render.LerpOklab(amberOchre, sunGold, 0.42)
			bellyShadowCol = render.LerpOklab(amberOchre, darkInk, 0.34)
			lineCol = darkInk
			accentCol = sunGold
			eyeRingCol = ivoryParchment
		}

		// Base anti-seam fill
		rc.FillPolygon(cell.boundary, baseCol, 1.0)
		closedB := append(append([]render.Point(nil), cell.boundary...), cell.boundary[0])
		rc.StrokePolyline(closedB, 0.85*scale, baseCol, 1.0)

		rc.PushClipPolygon(cell.boundary)

		e0, e1, e2, e3 := cell.edges[0], cell.edges[1], cell.edges[2], cell.edges[3]

		// Shade upper wing vs lower belly half of the cell
		diagSpine := traceCellSegment(e0, e1, e2, e3, 0.0, 1.0, 1.0, 0.0, 12)
		upperHalf := make([]render.Point, 0, samplesPerEdge*2+14)
		upperHalf = append(upperHalf, e3...)
		upperHalf = append(upperHalf, e0[1:]...)
		for i := len(diagSpine) - 1; i >= 0; i-- {
			upperHalf = append(upperHalf, diagSpine[i])
		}
		rc.FillPolygon(upperHalf, wingHiCol, 0.42+0.28*smoothT)

		lowerHalf := make([]render.Point, 0, samplesPerEdge*2+14)
		lowerHalf = append(lowerHalf, diagSpine...)
		lowerHalf = append(lowerHalf, e1[1:]...)
		lowerHalf = append(lowerHalf, e2[1:]...)
		rc.FillPolygon(lowerHalf, bellyShadowCol, 0.34*(1.0-0.3*smoothT))

		// Stage 1: Dutch Polder Farmland Furrows, Irrigation Ditches & Terraced Field Borders (wField > 0.01)
		if wField > 0.01 {
			// Concentric recessed field ditch border
			for rIdx, pad := range [2]float64{0.16, 0.30} {
				lo, hi := pad, 1.0-pad
				var fRing []render.Point
				fRing = append(fRing, traceCellSegment(e0, e1, e2, e3, lo, lo, hi, lo, 8)...)
				fRing = append(fRing, traceCellSegment(e0, e1, e2, e3, hi, lo, hi, hi, 8)[1:]...)
				fRing = append(fRing, traceCellSegment(e0, e1, e2, e3, hi, hi, lo, hi, 8)[1:]...)
				fRing = append(fRing, traceCellSegment(e0, e1, e2, e3, lo, hi, lo, lo, 8)[1:]...)
				rCol := accentCol
				if rIdx == 1 {
					rCol = lineCol
				}
				rc.StrokePolyline(fRing, (0.72-0.15*float64(rIdx))*scale, rCol, 0.58*wField)
			}

			// Parallel plowed field furrows (horizontal/diagonal alternating by parity!)
			const nFurrows = 7
			for f := 1; f < nFurrows; f++ {
				t := float64(f) / float64(nFurrows)
				var furrow []render.Point
				if cell.isDaySwan {
					furrow = traceCellSegment(e0, e1, e2, e3, 0.06, t, 0.94, t, 10)
				} else {
					furrow = traceCellSegment(e0, e1, e2, e3, t, 0.06, t, 0.94, 10)
				}
				rc.StrokePolyline(furrow, 0.54*scale, lineCol, 0.48*wField)
			}
		}

		// Stage 2: Transitional Chevron Feather-Rows Mapped Strictly Inside the Coons Patch (wMid > 0.02)
		if wMid > 0.02 {
			for cIdx := 1; cIdx <= 4; cIdx++ {
				f := 0.20 + 0.15*float64(cIdx)
				chev := make([]render.Point, 9)
				for s := 0; s <= 8; s++ {
					ts := (float64(s)/8.0)*2.0 - 1.0 // [-1, 1]
					bend := (1.0 - ts*ts) * 0.10 * smoothT
					uP := common.Constrain(0.50+ts*0.32, 0.14, 0.86)
					vP := common.Constrain(f+bend, 0.14, 0.86)
					chev[s] = cellCoonsPoint(e0, e1, e2, e3, uP, vP)
				}
				rc.StrokePolyline(chev, 0.62*scale, lineCol, 0.50*wMid)
			}
		}

		// Stage 3: Full Escher Day/Night Swan Anatomy — Coons-Patch Wing Pinions, Breast Plumage & Eye (wSwan > 0.01)
		if wSwan > 0.01 {
			const numPinions = 5
			for f := 1; f <= numPinions; f++ {
				uf := float64(f) / float64(numPinions+1)
				pinion := make([]render.Point, 9)
				for s := 0; s <= 8; s++ {
					ts := float64(s) / 8.0
					arch := math.Sin(ts*math.Pi) * 0.06 * wSwan
					uP := common.Constrain(0.56*(1.0-ts)+0.22*ts, 0.20, 0.78)
					vP := common.Constrain((0.26+0.48*uf)*(1.0-0.22*ts)+0.20*ts-arch, 0.20, 0.78)
					pinion[s] = cellCoonsPoint(e0, e1, e2, e3, uP, vP)
				}
				rc.StrokePolyline(pinion, (0.70+0.28*(1.0-uf))*scale, lineCol, 0.76*wSwan)
				if f%2 == 1 {
					rc.StrokePolyline(pinion, 0.46*scale, accentCol, 0.52*wSwan)
				}
			}

			// Anatomical Wing-Shoulder & Breast Contour Seam mapped through cellCoonsPoint
			shoulderSeam := make([]render.Point, 11)
			for s := 0; s <= 10; s++ {
				ts := float64(s) / 10.0
				uP := 0.76 - 0.52*ts - 0.06*math.Sin(ts*math.Pi)
				vP := 0.22 + 0.54*ts
				shoulderSeam[s] = cellCoonsPoint(e0, e1, e2, e3, uP, vP)
			}
			rc.StrokePolyline(shoulderSeam, 1.08*scale, lineCol, 0.82*wSwan)

			// Swan Eye positioned cleanly inside the head lobe
			headPt := cellCoonsPoint(e0, e1, e2, e3, 0.20, 0.76)
			eyeR := (1.35 + 1.85*wSwan) * scale
			rc.DrawGlowDot(headPt.X, headPt.Y, eyeR*1.42, lineCol, 0.88*wSwan, 0.92)
			rc.DrawGlowDot(headPt.X, headPt.Y, eyeR*1.00, eyeRingCol, 0.95*wSwan, 0.94)
			rc.DrawGlowDot(headPt.X, headPt.Y, eyeR*0.48, darkInk, 0.96*wSwan, 0.95)
			rc.DrawGlowDot(headPt.X-0.25*eyeR, headPt.Y-0.25*eyeR, eyeR*0.24, brightIvory, 0.92*wSwan, 0.90)
		}

		rc.PopClip()
	}

	// Pass 2: Crisp Master Woodcut Engraving Outlines along all shared figure-ground boundaries
	for idx := range cells {
		cell := &cells[idx]
		for k := 0; k < 4; k++ {
			ePts := cell.edges[k]
			eNormT := common.Constrain(cell.edgeMorph[k]/math.Max(strength, 0.01), 0.0, 1.0)
			masterW := (1.18 + 0.52*escherCubicSmoothstep(0.0, 1.0, eNormT)) * scale
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
		rc.FillPolygon(diamond, amberOchre, 0.95)
		rc.StrokePolyline(append(diamond, diamond[0]), 0.75*scale, darkInk, 0.92)
	}

	rc.ApplyPaperTexture(0.12, seed)
	c.MergeRenderCanvas(rc)
}
