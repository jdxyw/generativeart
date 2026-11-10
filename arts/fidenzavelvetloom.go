package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type fidenzaVelvetLoom struct {
	beams      int
	octaves    int
	turbulence float64
}

// NewFidenzaVelvetLoom returns a FidenzaVelvetLoom generator that weaves a collision-free
// curved lacquer & cloisonné flow-beam tapestry over a deep midnight-obsidian velvet ground
// with 3D drop shadows, internal silk striation ribs, and burnished 24K gold wire rims.
func NewFidenzaVelvetLoom(beams, octaves int, turbulence float64) *fidenzaVelvetLoom {
	if beams <= 0 {
		beams = 18
	}
	if octaves <= 0 {
		octaves = 5
	}
	if turbulence <= 0 {
		turbulence = 2.2
	}
	return &fidenzaVelvetLoom{
		beams:      common.ConstrainInt(beams, 8, 32),
		octaves:    common.ConstrainInt(octaves, 2, 8),
		turbulence: common.Constrain(turbulence, 0.8, 4.0),
	}
}

type velvetSegmentPoly struct {
	poly      [][2]float64
	upperEdge [][2]float64
	lowerEdge [][2]float64
	ribs      [][][2]float64
	col       color.RGBA
	colorIdx  int
	width     float64
}

// Generative renders a static FidenzaVelvetLoom composition.
func (fvl *fidenzaVelvetLoom) Generative(c *generativeart.Canva) {
	fvl.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless looping FidenzaVelvetLoom animation.
func (fvl *fidenzaVelvetLoom) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
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
		seed = 20262601
	}
	rng := rand.New(rand.NewSource(seed))
	ne := common.NewNoiseEngine(seed + 2601)

	isAnim := totalFrames > 1
	twoPiPhase := 2.0 * math.Pi * phase
	phaseCos, phaseSin := 0.0, 0.0
	if isAnim {
		phaseCos = 0.15 * math.Cos(twoPiPhase)
		phaseSin = 0.15 * math.Sin(twoPiPhase)
	}

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	// 1. Deep Midnight-Obsidian & Prussian Velvet Background (#070b16 -> #101b36)
	velvetCenter := color.RGBA{R: 0x10, G: 0x1b, B: 0x36, A: 255}
	velvetEdge := color.RGBA{R: 0x06, G: 0x09, B: 0x13, A: 255}
	for y := 0; y < h; y++ {
		ny := (float64(y)+0.5)/fh - 0.5
		for x := 0; x < w; x++ {
			nx := (float64(x)+0.5)/fw - 0.5
			rDist := common.Constrain(math.Hypot(nx, ny)*1.38, 0.0, 1.0)
			bgCol := render.LerpOklab(velvetCenter, velvetEdge, math.Pow(rDist, 1.15))
			lr, lg, lb, _ := render.ColorToLinearRGB(bgCol)
			rc.SetPixelLinear(x, y, lr, lg, lb, 1.0)
		}
	}

	margin := 30.0 * scale
	mortarGap := 3.2 * scale
	segmentGap := 2.4 * scale
	stepSize := 3.0 * scale
	if stepSize < 1.0 {
		stepSize = 1.0
	}

	// 2. Asymmetric Twin-Vortex + Perlin Derivative Flow Angle Field
	baseAngle := 0.18*math.Pi + (rng.Float64()-0.5)*0.30
	turbScale := common.Constrain(fvl.turbulence/2.2, 0.60, 1.65)
	octBonus := 0.04 * float64(fvl.octaves-5)

	flowAngle := func(px, py float64) float64 {
		nx := px / fw
		ny := py / fh

		val0, dx0, dy0 := ne.Perlin2DDeriv(nx*1.02+phaseCos, ny*1.02+phaseSin)
		val1 := 0.36 * ne.Perlin2D(nx*2.15-phaseSin*0.65, ny*2.15+phaseCos*0.65)

		// C^infinity smooth S-curve harmonic + analytical derivative deflection (zero branch-cut kinks!)
		sCurve := (0.94 + octBonus) * turbScale * (val0 + val1)
		gradSwing := 0.28 * math.Atan2(dy0, dx0+1.65)
		gentleArch := 0.44 * math.Sin((nx*1.25-ny*0.85)*math.Pi)
		return baseAngle + sCurve*math.Pi*0.72 + gradSwing + gentleArch
	}

	// 3. Collision-Free Occupancy Grid Packing
	cellSize := 2.0
	gridW := int(math.Ceil(fw/cellSize)) + 2
	gridH := int(math.Ceil(fh/cellSize)) + 2
	occupied := make([]bool, gridW*gridH)

	isDiscClear := func(cx, cy, radius float64) bool {
		if cx-radius < margin || cx+radius > fw-margin || cy-radius < margin || cy+radius > fh-margin {
			return false
		}
		checkR := radius + 0.5*mortarGap
		minGX := common.MaxInt(0, int(math.Floor((cx-checkR)/cellSize)))
		maxGX := common.MinInt(gridW-1, int(math.Ceil((cx+checkR)/cellSize)))
		minGY := common.MaxInt(0, int(math.Floor((cy-checkR)/cellSize)))
		maxGY := common.MinInt(gridH-1, int(math.Ceil((cy+checkR)/cellSize)))
		r2 := checkR * checkR
		for gy := minGY; gy <= maxGY; gy++ {
			py := (float64(gy) + 0.5) * cellSize
			dy := py - cy
			dy2 := dy * dy
			rowOff := gy * gridW
			for gx := minGX; gx <= maxGX; gx++ {
				px := (float64(gx) + 0.5) * cellSize
				dx := px - cx
				if dx*dx+dy2 <= r2 && occupied[rowOff+gx] {
					return false
				}
			}
		}
		return true
	}

	markDiscOccupied := func(cx, cy, radius float64) {
		stampR := radius + 0.5*mortarGap
		minGX := common.MaxInt(0, int(math.Floor((cx-stampR)/cellSize)))
		maxGX := common.MinInt(gridW-1, int(math.Ceil((cx+stampR)/cellSize)))
		minGY := common.MaxInt(0, int(math.Floor((cy-stampR)/cellSize)))
		maxGY := common.MinInt(gridH-1, int(math.Ceil((cy+stampR)/cellSize)))
		r2 := stampR * stampR
		for gy := minGY; gy <= maxGY; gy++ {
			py := (float64(gy) + 0.5) * cellSize
			dy := py - cy
			dy2 := dy * dy
			rowOff := gy * gridW
			for gx := minGX; gx <= maxGX; gx++ {
				px := (float64(gx) + 0.5) * cellSize
				dx := px - cx
				if dx*dx+dy2 <= r2 {
					occupied[rowOff+gx] = true
				}
			}
		}
	}

	type beamCandidate struct {
		x, y     float64
		width    float64
		maxSteps int
		tier     int
	}

	halton := func(index, base int) float64 {
		f := 1.0
		r := 0.0
		i := index
		for i > 0 {
			f /= float64(base)
			r += f * float64(i%base)
			i /= base
		}
		return r
	}

	usableW := math.Max(fw-2.0*margin, 40.0)
	usableH := math.Max(fh-2.0*margin, 40.0)
	totalCandidates := 1500
	candidates := make([]beamCandidate, 0, totalCandidates)

	for i := 0; i < totalCandidates; i++ {
		u := math.Mod(halton(i+1, 2)+rng.Float64()*0.06, 1.0)
		v := math.Mod(halton(i+1, 3)+rng.Float64()*0.06, 1.0)
		x0 := margin + (0.04+0.92*u)*usableW
		y0 := margin + (0.04+0.92*v)*usableH

		var width float64
		var maxSteps, tier int
		if i < 95 {
			width = (26.0 + rng.Float64()*11.0) * scale
			maxSteps = 34 + rng.Intn(28)
			tier = 0
		} else if i < 520 {
			width = (12.0 + rng.Float64()*9.0) * scale
			maxSteps = 24 + rng.Intn(28)
			tier = 1
		} else {
			width = (5.0 + rng.Float64()*5.0) * scale
			maxSteps = 15 + rng.Intn(28)
			tier = 2
		}
		candidates = append(candidates, beamCandidate{
			x: x0, y: y0, width: width, maxSteps: maxSteps, tier: tier,
		})
	}

	beams := make([]flowBeam, 0, 200)
	for idx, cand := range candidates {
		halfW := 0.5 * cand.width
		if !isDiscClear(cand.x, cand.y, halfW) {
			continue
		}

		fwdPts := make([][2]float64, 0, cand.maxSteps+1)
		fwdPts = append(fwdPts, [2]float64{cand.x, cand.y})
		cx, cy := cand.x, cand.y
		for s := 0; s < cand.maxSteps; s++ {
			th1 := flowAngle(cx, cy)
			mx := cx + 0.5*stepSize*math.Cos(th1)
			my := cy + 0.5*stepSize*math.Sin(th1)
			th2 := flowAngle(mx, my)
			nx := cx + stepSize*math.Cos(th2)
			ny := cy + stepSize*math.Sin(th2)
			if !isDiscClear(nx, ny, halfW) {
				break
			}
			fwdPts = append(fwdPts, [2]float64{nx, ny})
			cx, cy = nx, ny
		}

		remSteps := cand.maxSteps - (len(fwdPts) - 1)
		bwdPts := make([][2]float64, 0, remSteps)
		if remSteps > 0 {
			bx, by := cand.x, cand.y
			for s := 0; s < remSteps; s++ {
				th1 := flowAngle(bx, by)
				mx := bx - 0.5*stepSize*math.Cos(th1)
				my := by - 0.5*stepSize*math.Sin(th1)
				th2 := flowAngle(mx, my)
				nx := bx - stepSize*math.Cos(th2)
				ny := by - stepSize*math.Sin(th2)
				if !isDiscClear(nx, ny, halfW) {
					break
				}
				bwdPts = append(bwdPts, [2]float64{nx, ny})
				bx, by = nx, ny
			}
		}

		totalSteps := len(bwdPts) + len(fwdPts) - 1
		if totalSteps < 6 {
			continue
		}

		rawPts := make([][2]float64, 0, totalSteps+1)
		for i := len(bwdPts) - 1; i >= 0; i-- {
			rawPts = append(rawPts, bwdPts[i])
		}
		rawPts = append(rawPts, fwdPts...)

		for i := 0; i < len(rawPts); i++ {
			markDiscOccupied(rawPts[i][0], rawPts[i][1], halfW)
			if i+1 < len(rawPts) {
				mx := 0.5 * (rawPts[i][0] + rawPts[i+1][0])
				my := 0.5 * (rawPts[i][1] + rawPts[i+1][1])
				markDiscOccupied(mx, my, halfW)
			}
		}

		bPts := make([]flowBeamPoint, len(rawPts))
		cumArc := 0.0
		for i := 0; i < len(rawPts); i++ {
			if i > 0 {
				cumArc += math.Hypot(rawPts[i][0]-rawPts[i-1][0], rawPts[i][1]-rawPts[i-1][1])
			}
			var tx, ty float64
			if i == 0 {
				tx = rawPts[1][0] - rawPts[0][0]
				ty = rawPts[1][1] - rawPts[0][1]
			} else if i == len(rawPts)-1 {
				tx = rawPts[i][0] - rawPts[i-1][0]
				ty = rawPts[i][1] - rawPts[i-1][1]
			} else {
				tx = rawPts[i+1][0] - rawPts[i-1][0]
				ty = rawPts[i+1][1] - rawPts[i-1][1]
			}
			tLen := math.Hypot(tx, ty)
			if tLen < 1e-6 {
				tx, ty = 1.0, 0.0
			} else {
				tx /= tLen
				ty /= tLen
			}
			bPts[i] = flowBeamPoint{
				x: rawPts[i][0], y: rawPts[i][1], nx: -ty, ny: tx, s: cumArc,
			}
		}

		beams = append(beams, flowBeam{
			pts: bPts, width: cand.width, totalArc: cumArc, tier: cand.tier, seedOrder: idx,
		})
	}

	// 4. Imperial Cloisonné & Gilded Lacquer Palette blended with c.Opts().ColorSchema()
	basePalette := [7]color.RGBA{
		{R: 0xe6, G: 0x3e, B: 0x30, A: 255}, // 0: Cinnabar Vermilion (#e63e30)
		{R: 0xf6, G: 0xbe, B: 0x3c, A: 255}, // 1: Imperial 24K Gold (#f6be3c)
		{R: 0x20, G: 0xb6, B: 0xb2, A: 255}, // 2: Kingfisher Turquoise (#20b6b2)
		{R: 0x1c, G: 0x52, B: 0xa8, A: 255}, // 3: Lapis Lazuli (#1c52a8)
		{R: 0xf2, G: 0x7c, B: 0x38, A: 255}, // 4: Amber Coral (#f27c38)
		{R: 0x2a, G: 0x8c, B: 0x6a, A: 255}, // 5: Malachite Jade (#2a8c6a)
		{R: 0xfc, G: 0xf3, B: 0xe2, A: 255}, // 6: Mother-of-Pearl Ivory (#fcf3e2)
	}
	schema := c.Opts().ColorSchema()
	var palette [7]color.RGBA
	if len(schema) > 0 {
		uGrad := render.NewMultiStopGradient(schema...)
		for i := 0; i < 7; i++ {
			palette[i] = render.LerpOklab(basePalette[i], uGrad.SampleOklab(float64(i)/6.0), 0.22)
		}
	} else {
		palette = basePalette
	}

	paletteWeights := [7]float64{0.18, 0.20, 0.18, 0.15, 0.11, 0.09, 0.09}

	pickSegmentColor := func(prevIdx int, uPos, vPos float64) int {
		spatialWave := 0.5 + 0.5*math.Sin(uPos*math.Pi*1.8-vPos*math.Pi*1.4)
		var weights [7]float64
		totalW := 0.0
		for i := 0; i < 7; i++ {
			if i == prevIdx {
				continue
			}
			wVal := paletteWeights[i]
			if (i == 0 || i == 1 || i == 4) && spatialWave > 0.50 {
				wVal *= 1.30
			} else if (i == 2 || i == 3 || i == 5) && spatialWave <= 0.50 {
				wVal *= 1.30
			}
			weights[i] = wVal
			totalW += wVal
		}
		r := rng.Float64() * totalW
		acc := 0.0
		for i := 0; i < 7; i++ {
			acc += weights[i]
			if r <= acc && weights[i] > 0 {
				return i
			}
		}
		return (prevIdx + 1) % 7
	}

	sampleBeamAtArc := func(b *flowBeam, s float64) flowBeamPoint {
		pts := b.pts
		if s <= 0 || len(pts) == 1 {
			return pts[0]
		}
		last := len(pts) - 1
		if s >= pts[last].s {
			return pts[last]
		}
		lo, hi := 0, last
		for hi-lo > 1 {
			mid := (lo + hi) >> 1
			if pts[mid].s <= s {
				lo = mid
			} else {
				hi = mid
			}
		}
		ds := pts[hi].s - pts[lo].s
		t := 0.0
		if ds > 1e-6 {
			t = (s - pts[lo].s) / ds
		}
		x := pts[lo].x + t*(pts[hi].x-pts[lo].x)
		y := pts[lo].y + t*(pts[hi].y-pts[lo].y)
		nx := pts[lo].nx + t*(pts[hi].nx-pts[lo].nx)
		ny := pts[lo].ny + t*(pts[hi].ny-pts[lo].ny)
		nLen := math.Hypot(nx, ny)
		if nLen > 1e-6 {
			nx /= nLen
			ny /= nLen
		}
		return flowBeamPoint{x: x, y: y, nx: nx, ny: ny, s: s}
	}

	bandScale := common.Constrain(float64(fvl.beams)/18.0, 0.55, 1.75)
	segments := make([]velvetSegmentPoly, 0, len(beams)*5)

	for bIdx := range beams {
		b := &beams[bIdx]
		if b.totalArc < 10.0*scale {
			continue
		}

		meanSegLen := (20.0 * scale) / bandScale
		if b.tier == 0 {
			meanSegLen = (24.0 * scale) / bandScale
		} else if b.tier == 2 {
			meanSegLen = (16.0 * scale) / bandScale
		}
		numSegs := common.ConstrainInt(int(math.Round(b.totalArc/meanSegLen)), 1, 9)
		if b.totalArc >= 38.0*scale && numSegs < 3 {
			numSegs = 3
		}

		segWeights := make([]float64, numSegs)
		weightSum := 0.0
		for k := 0; k < numSegs; k++ {
			wVal := 0.55 + rng.Float64()*1.35
			if rng.Float64() < 0.24 {
				wVal *= 0.56
			}
			segWeights[k] = wVal
			weightSum += wVal
		}

		halfW := 0.5 * b.width
		prevColor := -1
		curArc := 0.0

		for k := 0; k < numSegs; k++ {
			segSpan := b.totalArc * (segWeights[k] / weightSum)
			sStart := curArc
			sEnd := curArc + segSpan
			curArc = sEnd

			if k > 0 {
				sStart += 0.5 * segmentGap
			}
			if k < numSegs-1 {
				sEnd -= 0.5 * segmentGap
			}
			if sEnd-sStart < 3.2*scale {
				continue
			}

			subSteps := common.MaxInt(3, int(math.Ceil((sEnd-sStart)/(1.8*scale))))
			leftCurve := make([][2]float64, subSteps+1)
			rightCurve := make([][2]float64, subSteps+1)

			// Internal silk striation ribs along wider hero/medium segments
			numRibs := 0
			if b.width >= 18.0*scale {
				numRibs = 3
			} else if b.width >= 11.0*scale {
				numRibs = 2
			}
			ribs := make([][][2]float64, numRibs)
			for rIdx := 0; rIdx < numRibs; rIdx++ {
				ribs[rIdx] = make([][2]float64, subSteps+1)
			}

			for j := 0; j <= subSteps; j++ {
				sPos := sStart + (sEnd-sStart)*float64(j)/float64(subSteps)
				pt := sampleBeamAtArc(b, sPos)
				leftCurve[j] = [2]float64{pt.x + halfW*pt.nx, pt.y + halfW*pt.ny}
				rightCurve[j] = [2]float64{pt.x - halfW*pt.nx, pt.y - halfW*pt.ny}
				for rIdx := 0; rIdx < numRibs; rIdx++ {
					uLat := -0.55 + 1.10*float64(rIdx+1)/float64(numRibs+1)
					ribs[rIdx][j] = [2]float64{pt.x + uLat*halfW*pt.nx, pt.y + uLat*halfW*pt.ny}
				}
			}

			poly := make([][2]float64, 0, 2*(subSteps+1)+1)
			for j := 0; j <= subSteps; j++ {
				poly = append(poly, leftCurve[j])
			}
			for j := subSteps; j >= 0; j-- {
				poly = append(poly, rightCurve[j])
			}
			poly = append(poly, leftCurve[0])

			midPt := sampleBeamAtArc(b, 0.5*(sStart+sEnd))
			colIdx := pickSegmentColor(prevColor, midPt.x/fw, midPt.y/fh)
			prevColor = colIdx

			var upperEdge, lowerEdge [][2]float64
			if midPt.nx+midPt.ny < 0 {
				upperEdge, lowerEdge = leftCurve, rightCurve
			} else {
				upperEdge, lowerEdge = rightCurve, leftCurve
			}

			segments = append(segments, velvetSegmentPoly{
				poly:      poly,
				upperEdge: upperEdge,
				lowerEdge: lowerEdge,
				ribs:      ribs,
				col:       palette[colIdx],
				colorIdx:  colIdx,
				width:     b.width,
			})
		}
	}

	// 5. Render 3D Drop Shadows, Satin Lacquer Tiles, Silk Striation Ribs & 24K Gold Wire Rims
	shadowCol := color.RGBA{R: 0x02, G: 0x03, B: 0x08, A: 255}
	dxShadow := 3.2 * scale
	dyShadow := 3.8 * scale

	toRenderPts := func(src [][2]float64) []render.Point {
		out := make([]render.Point, len(src))
		for i, p := range src {
			out[i] = render.Pt(p[0], p[1])
		}
		return out
	}

	for i := range segments {
		seg := &segments[i]
		nPts := len(seg.poly)
		softPoly := make([]render.Point, nPts)
		mainPoly := make([]render.Point, nPts)
		for j := 0; j < nPts; j++ {
			softPoly[j] = render.Pt(seg.poly[j][0]+dxShadow*1.45, seg.poly[j][1]+dyShadow*1.45)
			mainPoly[j] = render.Pt(seg.poly[j][0]+dxShadow, seg.poly[j][1]+dyShadow)
		}
		rc.FillPolygon(softPoly, shadowCol, 0.32)
		rc.FillPolygon(mainPoly, shadowCol, 0.55)
	}

	bevelHighlightCol := color.RGBA{R: 0xff, G: 0xf8, B: 0xe6, A: 255}
	bevelShadeCol := color.RGBA{R: 0x08, G: 0x0c, B: 0x18, A: 255}
	goldWireCol := color.RGBA{R: 0xf8, G: 0xcc, B: 0x62, A: 255}

	for i := range segments {
		seg := &segments[i]
		polyPts := toRenderPts(seg.poly)

		rc.FillPolygon(polyPts, seg.col, 1.0)

		// Subtle internal silk striation ribs on wide blocks
		for _, rib := range seg.ribs {
			rc.StrokePolyline(toRenderPts(rib), 0.75*scale, bevelHighlightCol, 0.16)
		}

		rc.StrokePolyline(toRenderPts(seg.upperEdge), 1.15*scale, bevelHighlightCol, 0.48)
		rc.StrokePolyline(toRenderPts(seg.lowerEdge), 0.95*scale, bevelShadeCol, 0.35)
		rc.StrokePolyline(polyPts, 0.95*scale, goldWireCol, 0.78)
	}

	rc.ApplyPaperTexture(0.08, seed)
	c.MergeRenderCanvas(rc)
}
