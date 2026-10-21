package render

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart/common"
	"github.com/tdewolff/canvas"
)

// DrawImpastoStroke renders a 3D-shaded thick impasto oil-paint bristle along pts.
// It extrudes layered cross-sectional ribbons (under-shadow, dark flank, saturated pigment core,
// lit flank, and specular ridge) modulated by lightDir in Oklab space.
func (c *Canvas) DrawImpastoStroke(
	pts []Point,
	width float64,
	baseCol color.RGBA,
	lightDir Point,
	alpha float64,
) {
	if c == nil || len(pts) < 2 || width <= 0 || alpha <= 0 || baseCol.A == 0 {
		return
	}
	curve := CatmullRomSpline(pts, 14)
	m := len(curve)
	if m < 2 {
		return
	}

	if lightDir.Len() < 1e-6 {
		lightDir = Pt(-0.6, -0.8).Norm()
	} else {
		lightDir = lightDir.Norm()
	}

	arcLen := make([]float64, m)
	totalLen := 0.0
	for i := 1; i < m; i++ {
		totalLen += curve[i].Dist(curve[i-1])
		arcLen[i] = totalLen
	}
	if totalLen < 1e-5 {
		return
	}

	normals := make([]Point, m)
	for i := 0; i < m; i++ {
		var tan Point
		switch {
		case i == 0:
			tan = curve[1].Sub(curve[0]).Norm()
		case i == m-1:
			tan = curve[m-1].Sub(curve[m-2]).Norm()
		default:
			tan = curve[i+1].Sub(curve[i-1]).Norm()
		}
		normals[i] = tan.Perp()
	}

	// Impasto profile: rounded brush belly with soft tapered root and rounded tip
	impastoProfile := func(t float64) float64 {
		t = common.Constrain(t, 0, 1)
		if t < 0.15 {
			u := t / 0.15
			return math.Sqrt(u)
		}
		if t > 0.82 {
			u := (1.0 - t) / 0.18
			return math.Sqrt(common.Constrain(u, 0, 1))
		}
		return 1.0 - 0.08*math.Sin((t-0.15)*math.Pi/0.67)
	}

	// Average normal alignment with lightDir to orient the highlight ridge on the lit side
	avgDot := 0.0
	for i := 0; i < m; i++ {
		avgDot += normals[i].X*lightDir.X + normals[i].Y*lightDir.Y
	}
	litSign := 1.0
	if avgDot < 0 {
		litSign = -1.0
	}

	baseLab := RGBAToOklab(baseCol)
	shadowCol := OklabToRGBA(Oklab{
		L: common.Constrain(baseLab.L*0.28, 0.01, 0.98),
		A: baseLab.A * 0.65,
		B: baseLab.B * 0.65,
	}, baseCol.A)
	darkRimCol := OklabToRGBA(Oklab{
		L: common.Constrain(baseLab.L*0.58, 0.02, 0.98),
		A: baseLab.A * 0.9,
		B: baseLab.B * 0.9,
	}, baseCol.A)
	litBodyCol := OklabToRGBA(Oklab{
		L: common.Constrain(baseLab.L*1.12+0.04, 0.05, 0.98),
		A: baseLab.A * 1.04,
		B: baseLab.B * 1.04,
	}, baseCol.A)
	specCol := OklabToRGBA(Oklab{
		L: common.Constrain(baseLab.L*1.28+0.14, 0.15, 0.99),
		A: baseLab.A * 0.72,
		B: baseLab.B * 0.72,
	}, baseCol.A)

	buildOffsetRibbon := func(latOffsetFrac, widthFrac float64) *canvas.Path {
		left := make([]Point, m)
		right := make([]Point, m)
		for i := 0; i < m; i++ {
			t := arcLen[i] / totalLen
			prof := impastoProfile(t)
			n := normals[i].Mul(litSign)
			center := curve[i].Add(n.Mul(width * latOffsetFrac * prof))
			halfW := width * widthFrac * 0.5 * prof
			left[i] = center.Add(n.Mul(halfW))
			right[i] = center.Sub(n.Mul(halfW))
		}
		p := &canvas.Path{}
		p.MoveTo(left[0].X, left[0].Y)
		for i := 1; i < m; i++ {
			p.LineTo(left[i].X, left[i].Y)
		}
		for i := m - 1; i >= 0; i-- {
			p.LineTo(right[i].X, right[i].Y)
		}
		p.Close()
		return p
	}

	// 1. Cast under-shadow on the unlit side
	c.RasterizePath(buildOffsetRibbon(-0.18, 1.05), shadowCol, alpha*0.68)
	// 2. Dark outer pigment body
	c.RasterizePath(buildOffsetRibbon(-0.06, 0.96), darkRimCol, alpha*0.92)
	// 3. Main saturated pigment core
	c.RasterizePath(buildOffsetRibbon(0.02, 0.78), baseCol, alpha)
	// 4. Lit upper flank
	c.RasterizePath(buildOffsetRibbon(0.14, 0.48), litBodyCol, alpha*0.92)
	// 5. Crisp impasto specular ridge
	if width >= 1.8 {
		c.RasterizePath(buildOffsetRibbon(0.20, 0.22), specCol, alpha*0.78)
	}
}

// SpatialGrid accelerates 2D minimum-distance collision queries for non-overlapping
// streamline & capsule packing (Jobard-Lefer style).
type SpatialGrid struct {
	Width    float64
	Height   float64
	CellSize float64
	Cols     int
	Rows     int
	cells    [][]Point
}

// NewSpatialGrid creates a new 2D spatial acceleration grid.
func NewSpatialGrid(w, h, cellSize float64) *SpatialGrid {
	if cellSize < 1.0 {
		cellSize = 4.0
	}
	cols := int(math.Ceil(w/cellSize)) + 2
	rows := int(math.Ceil(h/cellSize)) + 2
	return &SpatialGrid{
		Width:    w,
		Height:   h,
		CellSize: cellSize,
		Cols:     cols,
		Rows:     rows,
		cells:    make([][]Point, cols*rows),
	}
}

func (sg *SpatialGrid) cellCoords(p Point) (int, int) {
	cx := int(math.Floor(p.X / sg.CellSize))
	cy := int(math.Floor(p.Y / sg.CellSize))
	if cx < 0 {
		cx = 0
	} else if cx >= sg.Cols {
		cx = sg.Cols - 1
	}
	if cy < 0 {
		cy = 0
	} else if cy >= sg.Rows {
		cy = sg.Rows - 1
	}
	return cx, cy
}

// CanPlace returns true if p is at least minDist away from all points already in the grid.
func (sg *SpatialGrid) CanPlace(p Point, minDist float64) bool {
	cx, cy := sg.cellCoords(p)
	rCells := int(math.Ceil(minDist/sg.CellSize)) + 1
	minDistSq := minDist * minDist

	for dy := -rCells; dy <= rCells; dy++ {
		ny := cy + dy
		if ny < 0 || ny >= sg.Rows {
			continue
		}
		for dx := -rCells; dx <= rCells; dx++ {
			nx := cx + dx
			if nx < 0 || nx >= sg.Cols {
				continue
			}
			bucket := sg.cells[ny*sg.Cols+nx]
			for _, q := range bucket {
				ddx := p.X - q.X
				ddy := p.Y - q.Y
				if ddx*ddx+ddy*ddy < minDistSq {
					return false
				}
			}
		}
	}
	return true
}

// Insert adds point p to the spatial grid.
func (sg *SpatialGrid) Insert(p Point) {
	cx, cy := sg.cellCoords(p)
	idx := cy*sg.Cols + cx
	sg.cells[idx] = append(sg.cells[idx], p)
}

// InsertPolyline inserts dense sample points along pts into the spatial grid so future streamlines respect it.
func (sg *SpatialGrid) InsertPolyline(pts []Point, sampleSpacing float64) {
	if len(pts) == 0 {
		return
	}
	if sampleSpacing <= 0.5 {
		sampleSpacing = 1.5
	}
	sg.Insert(pts[0])
	for i := 0; i < len(pts)-1; i++ {
		p0, p1 := pts[i], pts[i+1]
		d := p0.Dist(p1)
		steps := int(math.Ceil(d / sampleSpacing))
		for s := 1; s <= steps; s++ {
			u := float64(s) / float64(steps)
			sg.Insert(p0.Lerp(p1, u))
		}
	}
}

// DrawCapsule strokes a smooth round-capped polyline/capsule with exact geometric offsetting.
func (c *Canvas) DrawCapsule(pts []Point, width float64, col color.RGBA, alpha float64) {
	if c == nil || len(pts) < 2 || width <= 0 || alpha <= 0 || col.A == 0 {
		return
	}
	p := &canvas.Path{}
	p.MoveTo(pts[0].X, pts[0].Y)
	for i := 1; i < len(pts); i++ {
		p.LineTo(pts[i].X, pts[i].Y)
	}
	c.StrokePath(p, width, canvas.RoundCap, canvas.RoundJoin, col, alpha)
}

// DrawAnnularSector renders an anti-aliased annular wedge [rInner, rOuter] x [theta1, theta2]
// centered at (cx, cy) using the active BlendMode.
func (c *Canvas) DrawAnnularSector(
	cx, cy, rInner, rOuter, theta1, theta2 float64,
	col color.RGBA,
	alpha float64,
) {
	if c == nil || rOuter <= rInner || rOuter <= 0 || alpha <= 0 || col.A == 0 {
		return
	}
	span := math.Abs(theta2 - theta1)
	if span < 1e-5 {
		return
	}
	steps := int(math.Ceil(span * math.Max(12.0, rOuter*0.35)))
	if steps < 4 {
		steps = 4
	}
	if steps > 96 {
		steps = 96
	}

	p := &canvas.Path{}
	for i := 0; i <= steps; i++ {
		u := float64(i) / float64(steps)
		ang := theta1 + (theta2-theta1)*u
		x := cx + rOuter*math.Cos(ang)
		y := cy + rOuter*math.Sin(ang)
		if i == 0 {
			p.MoveTo(x, y)
		} else {
			p.LineTo(x, y)
		}
	}
	if rInner <= 0.25 {
		p.LineTo(cx, cy)
	} else {
		for i := steps; i >= 0; i-- {
			u := float64(i) / float64(steps)
			ang := theta1 + (theta2-theta1)*u
			x := cx + rInner*math.Cos(ang)
			y := cy + rInner*math.Sin(ang)
			p.LineTo(x, y)
		}
	}
	p.Close()
	c.RasterizePath(p, col, alpha)
}

// ResamplePolylineUniform resamples a polyline into count uniformly arc-length-spaced points.
func ResamplePolylineUniform(pts []Point, count int) []Point {
	if len(pts) == 0 || count <= 0 {
		return nil
	}
	if len(pts) == 1 || count == 1 {
		return []Point{pts[0]}
	}
	cum := make([]float64, len(pts))
	total := 0.0
	for i := 1; i < len(pts); i++ {
		total += pts[i].Dist(pts[i-1])
		cum[i] = total
	}
	if total < 1e-6 {
		out := make([]Point, count)
		for i := range out {
			out[i] = pts[0]
		}
		return out
	}

	out := make([]Point, count)
	out[0] = pts[0]
	out[count-1] = pts[len(pts)-1]
	segIdx := 0
	for k := 1; k < count-1; k++ {
		target := (float64(k) / float64(count-1)) * total
		for segIdx < len(pts)-2 && cum[segIdx+1] < target {
			segIdx++
		}
		segLen := cum[segIdx+1] - cum[segIdx]
		if segLen < 1e-9 {
			out[k] = pts[segIdx]
		} else {
			u := (target - cum[segIdx]) / segLen
			out[k] = pts[segIdx].Lerp(pts[segIdx+1], u)
		}
	}
	return out
}

// DrawHatchPatch fills the chamber between two bounding curves curveA and curveB with
// numLines ultra-fine striated engraving curves C_s(t) = (1-s)*A(t) + s*B(t) + bow,
// modulated by 3D relief shading across s in [0, 1] and t in [0, 1].
func (c *Canvas) DrawHatchPatch(
	curveA, curveB []Point,
	numLines int,
	lineWidth float64,
	bow float64,
	colorFn func(s, t float64) color.RGBA,
	alpha float64,
) {
	if c == nil || len(curveA) < 2 || len(curveB) < 2 || numLines <= 0 || lineWidth <= 0 || alpha <= 0 {
		return
	}
	const samples = 48
	aPts := ResamplePolylineUniform(curveA, samples)
	bPts := ResamplePolylineUniform(curveB, samples)

	for i := 0; i < numLines; i++ {
		s := 0.5
		if numLines > 1 {
			s = float64(i) / float64(numLines-1)
		}
		bowWeight := 4.0 * s * (1.0 - s) * bow
		linePts := make([]Point, samples)
		for k := 0; k < samples; k++ {
			pa := aPts[k]
			pb := bPts[k]
			base := pa.Lerp(pb, s)
			if math.Abs(bowWeight) > 1e-6 {
				chord := pb.Sub(pa)
				n := chord.Perp()
				linePts[k] = base.Add(n.Mul(bowWeight))
			} else {
				linePts[k] = base
			}
		}

		// Draw in 4 sub-segments along t so color can shift smoothly along the hatched curve
		const chunks = 4
		step := (samples - 1) / chunks
		for ch := 0; ch < chunks; ch++ {
			i0 := ch * step
			i1 := (ch + 1) * step
			if ch == chunks-1 {
				i1 = samples - 1
			}
			tMid := (float64(i0) + float64(i1)) / (2.0 * float64(samples-1))
			col := colorFn(s, tMid)
			p := &canvas.Path{}
			p.MoveTo(linePts[i0].X, linePts[i0].Y)
			for k := i0 + 1; k <= i1; k++ {
				p.LineTo(linePts[k].X, linePts[k].Y)
			}
			c.StrokePath(p, lineWidth, canvas.RoundCap, canvas.RoundJoin, col, alpha)
		}
	}
}

// DrawTangentCone computes the exact external common tangents between circle (c1, r1) and circle (c2, r2),
// fills the conical sleeve with translucent axial shading + fine longitudinal striations, and strokes the tangent edges.
func (c *Canvas) DrawTangentCone(
	c1 Point, r1 float64,
	c2 Point, r2 float64,
	col1, col2, edgeCol color.RGBA,
	alpha float64,
) {
	if c == nil || r1 <= 0 || r2 <= 0 || alpha <= 0 {
		return
	}
	dVec := c2.Sub(c1)
	d := dVec.Len()
	if d <= math.Abs(r1-r2)+1e-3 {
		return
	}

	baseAng := math.Atan2(dVec.Y, dVec.X)
	cosGamma := common.Constrain((r1-r2)/d, -1.0, 1.0)
	gamma := math.Acos(cosGamma)

	angL := baseAng + gamma
	angR := baseAng - gamma

	// 1. Translucent conical hull fill in axial slices
	const axialSlices = 28
	for s := 0; s < axialSlices; s++ {
		t0 := float64(s) / float64(axialSlices)
		t1 := float64(s+1) / float64(axialSlices)
		tMid := 0.5 * (t0 + t1)

		ca := c1.Lerp(c2, t0)
		cb := c1.Lerp(c2, t1)
		ra := r1 + (r2-r1)*t0
		rb := r1 + (r2-r1)*t1

		p1L := Pt(ca.X+ra*math.Cos(angL), ca.Y+ra*math.Sin(angL))
		p1R := Pt(ca.X+ra*math.Cos(angR), ca.Y+ra*math.Sin(angR))
		p2R := Pt(cb.X+rb*math.Cos(angR), cb.Y+rb*math.Sin(angR))
		p2L := Pt(cb.X+rb*math.Cos(angL), cb.Y+rb*math.Sin(angL))

		sliceCol := LerpOklab(col1, col2, tMid)
		c.FillPolygon([]Point{p1L, p2L, p2R, p1R}, sliceCol, alpha*0.42)
	}

	// 2. Fine longitudinal generator striations across the cone surface
	const fibers = 34
	for f := 0; f <= fibers; f++ {
		u := float64(f) / float64(fibers)
		// Cosine spacing clusters fibers near the two outer tangent edges for 3D cylindrical rim density
		v := 0.5 * (1.0 - math.Cos(u*math.Pi))
		pStart := Pt(c1.X+r1*math.Cos(angL), c1.Y+r1*math.Sin(angL)).Lerp(
			Pt(c1.X+r1*math.Cos(angR), c1.Y+r1*math.Sin(angR)), v,
		)
		pEnd := Pt(c2.X+r2*math.Cos(angL), c2.Y+r2*math.Sin(angL)).Lerp(
			Pt(c2.X+r2*math.Cos(angR), c2.Y+r2*math.Sin(angR)), v,
		)

		edgeWeight := 1.0 - 0.55*math.Sin(v*math.Pi)
		fiberCol := LerpOklab(col1, col2, 0.5)
		if v < 0.12 || v > 0.88 {
			fiberCol = LerpOklab(fiberCol, edgeCol, 0.55)
		}
		path := &canvas.Path{}
		path.MoveTo(pStart.X, pStart.Y)
		path.LineTo(pEnd.X, pEnd.Y)
		c.StrokePath(path, 0.75, canvas.RoundCap, canvas.RoundJoin, fiberCol, alpha*0.48*edgeWeight)
	}

	// 3. Crisp outer tangent lines
	t1L := Pt(c1.X+r1*math.Cos(angL), c1.Y+r1*math.Sin(angL))
	t2L := Pt(c2.X+r2*math.Cos(angL), c2.Y+r2*math.Sin(angL))
	t1R := Pt(c1.X+r1*math.Cos(angR), c1.Y+r1*math.Sin(angR))
	t2R := Pt(c2.X+r2*math.Cos(angR), c2.Y+r2*math.Sin(angR))

	edgePath := &canvas.Path{}
	edgePath.MoveTo(t1L.X, t1L.Y)
	edgePath.LineTo(t2L.X, t2L.Y)
	edgePath.MoveTo(t1R.X, t1R.Y)
	edgePath.LineTo(t2R.X, t2R.Y)
	c.StrokePath(edgePath, 1.35, canvas.RoundCap, canvas.RoundJoin, edgeCol, math.Min(1.0, alpha*1.25))
}

// DrawShadedOrb renders a Bauhaus/Kandinsky translucent gradient sphere with an off-center highlight core and crisp rim.
func (c *Canvas) DrawShadedOrb(
	center Point,
	radius float64,
	coreCol, rimCol, strokeCol color.RGBA,
	lightOffset Point,
	alpha float64,
) {
	if c == nil || radius <= 0 || alpha <= 0 {
		return
	}
	minX := int(math.Max(0, math.Floor(center.X-radius-1.5)))
	maxX := int(math.Min(float64(c.Width-1), math.Ceil(center.X+radius+1.5)))
	minY := int(math.Max(0, math.Floor(center.Y-radius-1.5)))
	maxY := int(math.Min(float64(c.Height-1), math.Ceil(center.Y+radius+1.5)))

	focal := center.Add(lightOffset.Mul(radius * 0.38))
	maxFocusDist := radius * 1.38

	for py := minY; py <= maxY; py++ {
		fy := float64(py) + 0.5
		for px := minX; px <= maxX; px++ {
			fx := float64(px) + 0.5
			distCenter := math.Hypot(fx-center.X, fy-center.Y)
			if distCenter > radius+0.75 {
				continue
			}
			// Anti-aliased outer boundary coverage
			cov := 1.0
			if distCenter > radius-0.75 {
				cov = common.Constrain((radius+0.75-distCenter)/1.5, 0, 1)
			}
			distFocal := math.Hypot(fx-focal.X, fy-focal.Y)
			t := common.Constrain(distFocal/maxFocusDist, 0, 1)
			tSmooth := t * t * (3.0 - 2.0*t)
			col := LerpOklab(coreCol, rimCol, tSmooth)
			c.BlendPixel(px, py, col, alpha*cov)
		}
	}

	if strokeCol.A > 0 {
		c.DrawSoftRing(center.X, center.Y, radius, 1.35, strokeCol, math.Min(1.0, alpha*1.15))
	}
}

// InsetPolygon shrinks each vertex of pts toward its centroid by shrink pixels,
// preserving crisp dark grout lines between adjacent low-poly crystal facets.
func InsetPolygon(pts []Point, shrink float64) []Point {
	if len(pts) == 0 {
		return nil
	}
	var cx, cy float64
	for _, p := range pts {
		cx += p.X
		cy += p.Y
	}
	center := Pt(cx/float64(len(pts)), cy/float64(len(pts)))
	out := make([]Point, len(pts))
	for i, p := range pts {
		d := p.Sub(center)
		l := d.Len()
		if l > shrink*2.5 {
			out[i] = p.Sub(d.Mul(shrink / l))
		} else {
			out[i] = p.Lerp(center, 0.08)
		}
	}
	return out
}

// DrawMosaicBrick renders a crisp, flow-aligned, sheared parallelogram mosaic brick
// along segment p0 -> p1 with automatic grout inset scaled by scale.
func (c *Canvas) DrawMosaicBrick(
	p0, p1 Point,
	halfThick, shear, scale float64,
	col color.RGBA,
	alpha float64,
) {
	if c == nil || halfThick <= 0 || alpha <= 0 || col.A == 0 {
		return
	}
	if scale <= 0 {
		scale = 1.0
	}
	dir := p1.Sub(p0)
	length := dir.Len()
	if length < 1e-4 {
		return
	}
	tVec := dir.Mul(1.0 / length)
	nVec := tVec.Perp()

	insetL := math.Min(0.45*scale, length*0.04)
	insetT := math.Min(0.35*scale, halfThick*0.06)
	a := p0.Add(tVec.Mul(insetL))
	b := p1.Sub(tVec.Mul(insetL))
	ht := math.Max(1.1, halfThick-insetT)

	shearVec := tVec.Mul(shear * ht)
	v0 := a.Add(nVec.Mul(ht)).Add(shearVec)
	v1 := b.Add(nVec.Mul(ht)).Add(shearVec)
	v2 := b.Sub(nVec.Mul(ht)).Sub(shearVec)
	v3 := a.Sub(nVec.Mul(ht)).Sub(shearVec)

	c.FillPolygon([]Point{v0, v1, v2, v3}, col, alpha)
}

// DrawCurvedTessera renders a curved, sheared ceramic mosaic tile along a short
// polyline centerline (2+ points), preserving crisp grout gaps along both longitudinal
// and transverse joints.
func (c *Canvas) DrawCurvedTessera(
	pts []Point,
	halfThick, shear, scale float64,
	col color.RGBA,
	alpha float64,
) {
	if c == nil || len(pts) < 2 || halfThick <= 0 || alpha <= 0 || col.A == 0 {
		return
	}
	if len(pts) == 2 {
		c.DrawMosaicBrick(pts[0], pts[1], halfThick, shear, scale, col, alpha)
		return
	}
	if scale <= 0 {
		scale = 1.0
	}
	insetT := math.Min(0.35*scale, halfThick*0.06)
	ht := math.Max(1.1, halfThick-insetT)

	n := len(pts)
	upper := make([]Point, n)
	lower := make([]Point, n)
	for i := 0; i < n; i++ {
		var tan Point
		switch {
		case i == 0:
			tan = pts[1].Sub(pts[0]).Norm()
		case i == n-1:
			tan = pts[n-1].Sub(pts[n-2]).Norm()
		default:
			tan = pts[i+1].Sub(pts[i-1]).Norm()
		}
		norm := tan.Perp()
		base := pts[i]
		if i == 0 {
			segLen := pts[1].Dist(pts[0])
			base = base.Add(tan.Mul(math.Min(0.45*scale, segLen*0.08)))
		} else if i == n-1 {
			segLen := pts[n-1].Dist(pts[n-2])
			base = base.Sub(tan.Mul(math.Min(0.45*scale, segLen*0.08)))
		}
		shearVec := tan.Mul(shear * ht)
		upper[i] = base.Add(norm.Mul(ht)).Add(shearVec)
		lower[i] = base.Sub(norm.Mul(ht)).Sub(shearVec)
	}

	poly := make([]Point, 0, 2*n)
	poly = append(poly, upper...)
	for i := n - 1; i >= 0; i-- {
		poly = append(poly, lower[i])
	}
	c.FillPolygon(poly, col, alpha)
}

// DrawCloisonneCell renders a vitreous cloisonné enamel (景泰蓝) compartment bounded by pts:
// pooled dark meniscus rim near the brass wire walls, saturated jade/lapis/coral core,
// off-center dome specular glaze sheen, and a double-stroked gilded brass cloisonné wire.
func (c *Canvas) DrawCloisonneCell(
	pts []Point,
	baseCol, rimCol, highlightCol, wireShadow, wireGold, wireSpec color.RGBA,
	wireWidth float64,
) {
	if c == nil || len(pts) < 3 {
		return
	}
	if wireWidth <= 0 {
		wireWidth = 1.6
	}

	// 1. Pooled deep enamel rim filling the entire cell
	c.FillPolygon(pts, rimCol, 1.0)

	// Compute cell centroid for meniscus progression and off-center dome sheen
	var cx, cy float64
	for _, p := range pts {
		cx += p.X
		cy += p.Y
	}
	center := Pt(cx/float64(len(pts)), cy/float64(len(pts)))

	// 2. Layered vitreous enamel meniscus (from pooled rim to luminous mid-body)
	const layers = 4
	for l := 1; l <= layers; l++ {
		u := float64(l) / float64(layers+1)
		shrinkFrac := u * 0.44
		ringPoly := make([]Point, len(pts))
		for i, p := range pts {
			ringPoly[i] = p.Lerp(center, shrinkFrac)
		}
		col := LerpOklab(rimCol, baseCol, math.Pow(u, 0.72))
		c.FillPolygon(ringPoly, col, 1.0)
	}

	// 3. Off-center vitreous dome highlight shifted toward upper-left light source
	domeCenter := center
	if len(pts) > 0 {
		domeCenter = center.Lerp(pts[0], 0.16)
	}
	domePoly := make([]Point, len(pts))
	for i, p := range pts {
		domePoly[i] = p.Lerp(domeCenter, 0.64)
	}
	c.FillPolygon(domePoly, LerpOklab(baseCol, highlightCol, 0.58), 0.92)

	specPoly := make([]Point, len(pts))
	for i, p := range pts {
		specPoly[i] = p.Lerp(domeCenter, 0.80)
	}
	c.FillPolygon(specPoly, highlightCol, 0.85)

	// 4. Double-stroked Gilded Brass Cloisonné Wire (shadow bevel + gold core + specular filament)
	path := &canvas.Path{}
	path.MoveTo(pts[0].X, pts[0].Y)
	for i := 1; i < len(pts); i++ {
		path.LineTo(pts[i].X, pts[i].Y)
	}
	path.Close()

	c.StrokePath(path, wireWidth*1.45, canvas.RoundCap, canvas.RoundJoin, wireShadow, 0.95)
	c.StrokePath(path, wireWidth*0.92, canvas.RoundCap, canvas.RoundJoin, wireGold, 1.0)
	c.StrokePath(path, wireWidth*0.36, canvas.RoundCap, canvas.RoundJoin, wireSpec, 0.88)
}

// DrawStainedGlassCell renders a backlit Gothic stained-glass compartment with dark lead-came
// solder borders, ruby/cobalt edge patina, glowing jewel-toned core, and a crystalline bevel.
func (c *Canvas) DrawStainedGlassCell(
	pts []Point,
	edgeCol, coreCol, glowCol, leadCol color.RGBA,
	leadWidth float64,
) {
	if c == nil || len(pts) < 3 {
		return
	}
	if leadWidth <= 0 {
		leadWidth = 1.8
	}

	var cx, cy float64
	for _, p := range pts {
		cx += p.X
		cy += p.Y
	}
	center := Pt(cx/float64(len(pts)), cy/float64(len(pts)))

	// 1. Outer lead-came shadow patina
	c.FillPolygon(pts, edgeCol, 1.0)

	// 2. Backlit stained-glass transmission layers
	midPoly := make([]Point, len(pts))
	corePoly := make([]Point, len(pts))
	glowPoly := make([]Point, len(pts))
	for i, p := range pts {
		midPoly[i] = p.Lerp(center, 0.16)
		corePoly[i] = p.Lerp(center, 0.36)
		glowPoly[i] = p.Lerp(center, 0.62)
	}
	c.FillPolygon(midPoly, LerpOklab(edgeCol, coreCol, 0.65), 1.0)
	c.FillPolygon(corePoly, coreCol, 1.0)
	c.FillPolygon(glowPoly, glowCol, 0.95)

	// 3. Subtle diagonal crystalline refraction facet inside the glass pane
	if len(pts) >= 3 {
		bevelTri := []Point{midPoly[0], midPoly[1], center}
		c.FillPolygon(bevelTri, glowCol, 0.35)
	}

	// 4. Heavy Gothic lead-came outline
	path := &canvas.Path{}
	path.MoveTo(pts[0].X, pts[0].Y)
	for i := 1; i < len(pts); i++ {
		path.LineTo(pts[i].X, pts[i].Y)
	}
	path.Close()
	c.StrokePath(path, leadWidth, canvas.RoundCap, canvas.RoundJoin, leadCol, 1.0)
}


