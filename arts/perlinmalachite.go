package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type perlinMalachite struct {
	bands        int
	octaves      int
	warpStrength float64
}

// NewPerlinMalachite returns a PerlinMalachite generator that composes a museum-grade
// Fidenza-inspired Perlin flow-field curved color-block tapestry on cotton parchment.
func NewPerlinMalachite(bands, octaves int, warpStrength float64) *perlinMalachite {
	if bands <= 0 {
		bands = 16
	}
	if octaves <= 0 {
		octaves = 5
	}
	if warpStrength <= 0 {
		warpStrength = 2.4
	}
	return &perlinMalachite{
		bands:        bands,
		octaves:      octaves,
		warpStrength: warpStrength,
	}
}

type flowBeamPoint struct {
	x, y   float64
	nx, ny float64 // unit left-normal (-sin(theta), cos(theta))
	s      float64 // cumulative arc length from start of beam
}

type flowBeam struct {
	pts       []flowBeamPoint
	width     float64
	totalArc  float64
	tier      int // 0: hero wide, 1: medium, 2: fine accent
	seedOrder int
}

type flowSegmentPoly struct {
	poly      [][2]float64
	upperEdge [][2]float64
	lowerEdge [][2]float64
	col       color.RGBA
	colorIdx  int
	width     float64
}

// Generative renders a static Fidenza-style Perlin flow-field curved color-block tapestry.
func (pm *perlinMalachite) Generative(c *generativeart.Canva) {
	pm.GenerativeFrame(c, 0, 1, 0)
}

// GenerativeFrame renders frame frameIdx of a seamless looping PerlinMalachite animation.
func (pm *perlinMalachite) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
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
		seed = 20261901
	}
	rng := rand.New(rand.NewSource(seed))
	ne := common.NewNoiseEngine(seed)

	isAnim := totalFrames > 1
	twoPiPhase := 2.0 * math.Pi * phase
	phaseCos := 0.0
	phaseSin := 0.0
	if isAnim {
		phaseCos = 0.16 * math.Cos(twoPiPhase)
		phaseSin = 0.16 * math.Sin(twoPiPhase)
	}

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	// 1. Warm museum cotton parchment background (#f4ede0) with subtle warm radial tone
	parchmentBase := color.RGBA{R: 0xf4, G: 0xed, B: 0xe0, A: 255}
	parchmentEdge := color.RGBA{R: 0xec, G: 0xe2, B: 0xd0, A: 255}
	for y := 0; y < h; y++ {
		ny := (float64(y)+0.5)/fh - 0.5
		for x := 0; x < w; x++ {
			nx := (float64(x)+0.5)/fw - 0.5
			rDist := common.Constrain(math.Hypot(nx, ny)*1.35, 0.0, 1.0)
			bgCol := render.LerpOklab(parchmentBase, parchmentEdge, rDist*rDist*0.45)
			lr, lg, lb, _ := render.ColorToLinearRGB(bgCol)
			rc.SetPixelLinear(x, y, lr, lg, lb, 1.0)
		}
	}

	margin := 36.0 * scale
	mortarGap := 3.5 * scale
	segmentGap := 2.5 * scale
	stepSize := 3.0 * scale
	if stepSize < 1.0 {
		stepSize = 1.0
	}

	// 2. Smooth Low-Frequency Perlin Angle Field:
	// Uses ne.Perlin2DDeriv(nx*0.95, ny*0.95) + 0.35*ne.Perlin2D(nx*2.1, ny*2.1)
	// to sculpt a sweeping S-curve flow angle theta(x, y) across the canvas.
	baseAngle := -0.22*math.Pi + (rng.Float64()-0.5)*0.36
	warpScale := common.Constrain(pm.warpStrength/2.4, 0.65, 1.45)
	octBonus := 0.04 * float64(common.ConstrainInt(pm.octaves, 1, 8)-5)

	flowAngle := func(px, py float64) float64 {
		nx := px / fw
		ny := py / fh

		val0, dx0, dy0 := ne.Perlin2DDeriv(nx*0.95+phaseCos, ny*0.95+phaseSin)
		val1 := 0.35 * ne.Perlin2D(nx*2.1-phaseSin*0.65, ny*2.1+phaseCos*0.65)

		// Sweeping S-curve harmonic + smooth analytical derivative curl deflection
		sCurve := (0.92 + octBonus) * warpScale * (val0 + val1)
		gradSwing := 0.28 * math.Atan2(dy0, dx0+1.65)
		gentleArch := 0.42 * math.Sin((nx*1.15-ny*0.75)*math.Pi)

		return baseAngle + sCurve*math.Pi*0.72 + gradSwing + gentleArch
	}

	// 3. Non-Overlapping Curved Color-Block Beams (Fidenza Packing Algorithm):
	// Fine 2px-cell occupancy grid where each accepted beam marks its half-width + 0.5*mortarGap,
	// and each candidate tests its half-width + 0.5*mortarGap so any two beams stay >= mortarGap apart.
	cellSize := 2.0
	gridW := int(math.Ceil(fw/cellSize)) + 2
	gridH := int(math.Ceil(fh/cellSize)) + 2
	occupied := make([]bool, gridW*gridH)

	isDiscClear := func(cx, cy, radius float64) bool {
		if cx-radius < margin || cx+radius > fw-margin || cy-radius < margin || cy+radius > fh-margin {
			return false
		}
		checkR := radius + 0.5*mortarGap
		minGX := int(math.Floor((cx - checkR) / cellSize))
		maxGX := int(math.Ceil((cx + checkR) / cellSize))
		minGY := int(math.Floor((cy - checkR) / cellSize))
		maxGY := int(math.Ceil((cy + checkR) / cellSize))
		if minGX < 0 {
			minGX = 0
		}
		if maxGX >= gridW {
			maxGX = gridW - 1
		}
		if minGY < 0 {
			minGY = 0
		}
		if maxGY >= gridH {
			maxGY = gridH - 1
		}
		r2 := checkR * checkR
		for gy := minGY; gy <= maxGY; gy++ {
			py := (float64(gy) + 0.5) * cellSize
			dy := py - cy
			dy2 := dy * dy
			rowOff := gy * gridW
			for gx := minGX; gx <= maxGX; gx++ {
				px := (float64(gx) + 0.5) * cellSize
				dx := px - cx
				if dx*dx+dy2 <= r2 {
					if occupied[rowOff+gx] {
						return false
					}
				}
			}
		}
		return true
	}

	markDiscOccupied := func(cx, cy, radius float64) {
		stampR := radius + 0.5*mortarGap
		minGX := int(math.Floor((cx - stampR) / cellSize))
		maxGX := int(math.Ceil((cx + stampR) / cellSize))
		minGY := int(math.Floor((cy - stampR) / cellSize))
		maxGY := int(math.Ceil((cy + stampR) / cellSize))
		if minGX < 0 {
			minGX = 0
		}
		if maxGX >= gridW {
			maxGX = gridW - 1
		}
		if minGY < 0 {
			minGY = 0
		}
		if maxGY >= gridH {
			maxGY = gridH - 1
		}
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

	// Generate ~360 stratified + jittered candidate beams ordered from:
	// - Wide hero beams (width = 28..38 px * scale)
	// - Medium beams (width = 14..22 px * scale)
	// - Fine accent beams (width = 5..9 px * scale)
	type beamCandidate struct {
		x, y     float64
		width    float64
		maxSteps int
		tier     int
	}

	totalCandidates := 1400
	candidates := make([]beamCandidate, 0, totalCandidates)

	// Helper using Halton low-discrepancy sequence + jitter for uniform canvas coverage
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

	for i := 0; i < totalCandidates; i++ {
		u := math.Mod(halton(i+1, 2)+rng.Float64()*0.06, 1.0)
		v := math.Mod(halton(i+1, 3)+rng.Float64()*0.06, 1.0)
		x0 := margin + (0.04+0.92*u)*usableW
		y0 := margin + (0.04+0.92*v)*usableH

		var width float64
		var maxSteps int
		var tier int
		if i < 90 {
			// Hero wide beams: 28..38 px * scale
			width = (28.0 + rng.Float64()*10.0) * scale
			maxSteps = 32 + rng.Intn(26)
			tier = 0
		} else if i < 480 {
			// Medium structural beams: 13..22 px * scale
			width = (13.0 + rng.Float64()*9.0) * scale
			maxSteps = 22 + rng.Intn(28)
			tier = 1
		} else {
			// Fine accent beams: 5..10 px * scale
			width = (5.0 + rng.Float64()*5.0) * scale
			maxSteps = 14 + rng.Intn(28)
			tier = 2
		}

		candidates = append(candidates, beamCandidate{
			x:        x0,
			y:        y0,
			width:    width,
			maxSteps: maxSteps,
			tier:     tier,
		})
	}

	beams := make([]flowBeam, 0, 180)
	for idx, cand := range candidates {
		halfW := 0.5 * cand.width
		if !isDiscClear(cand.x, cand.y, halfW) {
			continue
		}

		// Trace forward along the smooth Perlin angle field with RK2 midpoint integration
		fwdPts := make([][2]float64, 0, cand.maxSteps+1)
		fwdPts = append(fwdPts, [2]float64{cand.x, cand.y})
		cx, cy := cand.x, cand.y
		fwdSteps := cand.maxSteps
		for s := 0; s < fwdSteps; s++ {
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

		// Also trace backward from seed if room remains in cand.maxSteps so beams nest tightly into pockets
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

		// Assemble ordered centerline points from backward tail to forward head
		rawPts := make([][2]float64, 0, totalSteps+1)
		for i := len(bwdPts) - 1; i >= 0; i-- {
			rawPts = append(rawPts, bwdPts[i])
		}
		rawPts = append(rawPts, fwdPts...)

		// Mark beam footprint in the occupancy grid (including dense half-step midpoints)
		for i := 0; i < len(rawPts); i++ {
			markDiscOccupied(rawPts[i][0], rawPts[i][1], halfW)
			if i+1 < len(rawPts) {
				mx := 0.5 * (rawPts[i][0] + rawPts[i+1][0])
				my := 0.5 * (rawPts[i][1] + rawPts[i+1][1])
				markDiscOccupied(mx, my, halfW)
			}
		}

		// Compute smooth tangents, unit left-normals, and cumulative arc length
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
				x:  rawPts[i][0],
				y:  rawPts[i][1],
				nx: -ty,
				ny: tx,
				s:  cumArc,
			}
		}

		beams = append(beams, flowBeam{
			pts:       bPts,
			width:     cand.width,
			totalArc:  cumArc,
			tier:      cand.tier,
			seedOrder: idx,
		})
	}

	// 4. Curated Mid-Century / Dunhuang Palette
	palette := [7]color.RGBA{
		{R: 0xd8, G: 0x3b, B: 0x2d, A: 255}, // 0: Vermilion Red (#d83b2d)
		{R: 0xf0, G: 0xa8, B: 0x28, A: 255}, // 1: Ochre Mustard Gold (#f0a828)
		{R: 0x16, G: 0x38, B: 0x68, A: 255}, // 2: Deep Prussian Blue (#163868)
		{R: 0x22, G: 0x96, B: 0x8c, A: 255}, // 3: Celadon Turquoise (#22968c)
		{R: 0xe6, G: 0x6c, B: 0x40, A: 255}, // 4: Warm Terracotta Coral (#e66c40)
		{R: 0x1c, G: 0x22, B: 0x2b, A: 255}, // 5: Obsidian Charcoal (#1c222b)
		{R: 0xfb, G: 0xf6, B: 0xec, A: 255}, // 6: Warm Ivory Cream (#fbf6ec)
	}

	// Balanced architectural color weights for Fidenza-style rhythm
	paletteWeights := [7]float64{
		0.17, // Vermilion Red
		0.18, // Ochre Mustard Gold
		0.19, // Deep Prussian Blue
		0.17, // Celadon Turquoise
		0.12, // Warm Terracotta Coral
		0.08, // Obsidian Charcoal
		0.09, // Warm Ivory Cream
	}

	pickSegmentColor := func(prevIdx int, uPos, vPos float64) int {
		// Subtle spatial bias so warm and cool clusters echo across the tapestry
		spatialWave := 0.5 + 0.5*math.Sin(uPos*math.Pi*1.6-vPos*math.Pi*1.2)
		var weights [7]float64
		totalW := 0.0
		for i := 0; i < 7; i++ {
			if i == prevIdx {
				weights[i] = 0.0
				continue
			}
			wVal := paletteWeights[i]
			if (i == 0 || i == 1 || i == 4) && spatialWave > 0.52 {
				wVal *= 1.25
			} else if (i == 2 || i == 3) && spatialWave <= 0.52 {
				wVal *= 1.25
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

	// Interpolate point & normal along a beam at exact arc length s
	sampleBeamAtArc := func(b *flowBeam, s float64) flowBeamPoint {
		pts := b.pts
		if s <= 0 || len(pts) == 1 {
			return pts[0]
		}
		last := len(pts) - 1
		if s >= pts[last].s {
			return pts[last]
		}
		// Binary search for interval containing s
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

	// Slice each accepted beam into proportional color segments (3..9 segments modulated by pm.bands)
	// with a crisp 2.5px*scale gap between consecutive segments.
	bandScale := common.Constrain(float64(pm.bands)/16.0, 0.55, 1.65)
	segments := make([]flowSegmentPoly, 0, len(beams)*5)

	for bIdx := range beams {
		b := &beams[bIdx]
		if b.totalArc < 10.0*scale {
			continue
		}

		// Target number of segments proportional to beam arc length and pm.bands (clamped to 1..9)
		meanSegLen := (19.0 * scale) / bandScale
		if b.tier == 0 {
			meanSegLen = (23.0 * scale) / bandScale
		} else if b.tier == 2 {
			meanSegLen = (16.0 * scale) / bandScale
		}
		numSegs := int(math.Round(b.totalArc / meanSegLen))
		if b.totalArc >= 38.0*scale && numSegs < 3 {
			numSegs = 3
		}
		if numSegs < 1 {
			numSegs = 1
		}
		if numSegs > 9 {
			numSegs = 9
		}

		// Generate proportional segment weights (mix of long hero blocks and crisp short blocks)
		segWeights := make([]float64, numSegs)
		weightSum := 0.0
		for k := 0; k < numSegs; k++ {
			wVal := 0.55 + rng.Float64()*1.35
			if rng.Float64() < 0.22 {
				wVal *= 0.58 // occasional crisp narrow stripe accent
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

			// Leave a crisp 2.5px*scale gap between adjacent segments along the beam
			if k > 0 {
				sStart += 0.5 * segmentGap
			}
			if k < numSegs-1 {
				sEnd -= 0.5 * segmentGap
			}
			if sEnd-sStart < 3.2*scale {
				continue
			}

			// Resample centerline along [sStart, sEnd] at fine intervals for ultra-crisp curved edges
			subSteps := int(math.Ceil((sEnd - sStart) / (1.8 * scale)))
			if subSteps < 3 {
				subSteps = 3
			}

			leftCurve := make([][2]float64, subSteps+1)
			rightCurve := make([][2]float64, subSteps+1)
			for j := 0; j <= subSteps; j++ {
				sPos := sStart + (sEnd-sStart)*float64(j)/float64(subSteps)
				pt := sampleBeamAtArc(b, sPos)
				leftCurve[j] = [2]float64{
					pt.x + halfW*pt.nx,
					pt.y + halfW*pt.ny,
				}
				rightCurve[j] = [2]float64{
					pt.x - halfW*pt.nx,
					pt.y - halfW*pt.ny,
				}
			}

			// Closed polygon with crisp butt caps: leftCurve forward + rightCurve backward + close
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

			// Choose which lateral curve faces top-left light (-1, -1) for the bevel highlight
			var upperEdge, lowerEdge [][2]float64
			if midPt.nx+midPt.ny < 0 {
				upperEdge = leftCurve
				lowerEdge = rightCurve
			} else {
				upperEdge = rightCurve
				lowerEdge = leftCurve
			}

			segments = append(segments, flowSegmentPoly{
				poly:      poly,
				upperEdge: upperEdge,
				lowerEdge: lowerEdge,
				col:       palette[colIdx],
				colorIdx:  colIdx,
				width:     b.width,
			})
		}
	}

	// 5. Crisp Vector Polygon Rendering + 3D Drop Shadows:
	// Pass 1: Soft dark drop-shadow polygons offset by (+2.5*scale, +3.0*scale) with alpha 0.18
	shadowMainCol := color.RGBA{R: 0x16, G: 0x14, B: 0x12, A: 255}
	dxShadow := 2.5 * scale
	dyShadow := 3.0 * scale

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
			softPoly[j] = render.Pt(
				seg.poly[j][0]+dxShadow*1.32,
				seg.poly[j][1]+dyShadow*1.32,
			)
			mainPoly[j] = render.Pt(
				seg.poly[j][0]+dxShadow,
				seg.poly[j][1]+dyShadow,
			)
		}
		rc.FillPolygon(softPoly, shadowMainCol, 0.08)
		rc.FillPolygon(mainPoly, shadowMainCol, 0.18)
	}

	// Pass 2: Fill crisp curved segment polygons, top-left bevel highlights, and razor-thin dark ink/gold borders
	bevelHighlightCol := color.RGBA{R: 0xff, G: 0xfb, B: 0xf0, A: 255}
	bevelShadeCol := color.RGBA{R: 0x12, G: 0x14, B: 0x18, A: 255}
	inkBorderCol := color.RGBA{R: 0x1c, G: 0x1a, B: 0x17, A: 255}
	goldBorderCol := color.RGBA{R: 0xd8, G: 0xa4, B: 0x38, A: 255}

	for i := range segments {
		seg := &segments[i]
		polyPts := toRenderPts(seg.poly)

		// Fill the crisp curved segment polygon with its chosen palette color
		rc.FillPolygon(polyPts, seg.col, 1.0)

		// Subtle top-left bevel highlight along the upper edge and subtle shade on lower edge
		rc.StrokePolyline(toRenderPts(seg.upperEdge), 1.05*scale, bevelHighlightCol, 0.40)
		rc.StrokePolyline(toRenderPts(seg.lowerEdge), 0.85*scale, bevelShadeCol, 0.22)

		// Razor-thin dark ink/gold border (width 0.9*scale, alpha 0.55)
		borderCol := inkBorderCol
		if seg.colorIdx == 2 || seg.colorIdx == 5 {
			// Deep Prussian Blue and Obsidian Charcoal get a subtle gold-ink border accent
			borderCol = goldBorderCol
		}
		rc.StrokePolyline(polyPts, 0.9*scale, borderCol, 0.55)
	}

	// 6. Finish with tactile cotton paper texture and merge into target canvas
	rc.ApplyPaperTexture(0.14, seed)
	c.MergeRenderCanvas(rc)
}
