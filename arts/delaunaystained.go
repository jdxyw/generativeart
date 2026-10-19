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

type delaunayStained struct {
	pointNum int
	jitter   float64
}

// NewDelaunayStained returns a delaunayStained engine that computes an exact Bowyer-Watson
// Delaunay triangulation and renders a prismatic 3D faceted stained-glass & caustic crystal mosaic.
func NewDelaunayStained(pointNum int, jitter float64) *delaunayStained {
	if pointNum <= 0 {
		pointNum = 220
	}
	if jitter <= 0 {
		jitter = 0.65
	}
	return &delaunayStained{
		pointNum: pointNum,
		jitter:   jitter,
	}
}

type delaunayTri struct {
	i, j, k    int
	cx, cy     float64
	rSq        float64
	degenerate bool
}

type delaunayEdge struct {
	u, v int
}

func canonicalEdgeKey(u, v int) [2]int {
	if u < v {
		return [2]int{u, v}
	}
	return [2]int{v, u}
}

// makeDelaunayTri constructs a CCW-oriented triangle (i, j, k) and precomputes its exact circumcircle.
func makeDelaunayTri(pts []render.Point, i, j, k int) delaunayTri {
	a, b, c := pts[i], pts[j], pts[k]
	cross := (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
	if cross < 0 {
		j, k = k, j
		b, c = c, b
	}
	d := 2.0 * (a.X*(b.Y-c.Y) + b.X*(c.Y-a.Y) + c.X*(a.Y-b.Y))
	if math.Abs(d) < 1e-9 {
		return delaunayTri{i: i, j: j, k: k, degenerate: true}
	}
	aSq := a.X*a.X + a.Y*a.Y
	bSq := b.X*b.X + b.Y*b.Y
	cSq := c.X*c.X + c.Y*c.Y

	ux := (aSq*(b.Y-c.Y) + bSq*(c.Y-a.Y) + cSq*(a.Y-b.Y)) / d
	uy := (aSq*(c.X-b.X) + bSq*(a.X-c.X) + cSq*(b.X-a.X)) / d
	dx := a.X - ux
	dy := a.Y - uy
	return delaunayTri{
		i:   i,
		j:   j,
		k:   k,
		cx:  ux,
		cy:  uy,
		rSq: dx*dx + dy*dy,
	}
}

// bowyerWatsonTriangulate computes the exact 2D Delaunay triangulation of sites inside [0, w] x [0, h].
func bowyerWatsonTriangulate(sites []render.Point, w, h float64) [][3]render.Point {
	// Deduplicate sites that are closer than 1.2 pixels to guarantee numerical stability
	cleanSites := make([]render.Point, 0, len(sites))
	for _, p := range sites {
		dup := false
		for _, q := range cleanSites {
			if math.Hypot(p.X-q.X, p.Y-q.Y) < 1.2 {
				dup = true
				break
			}
		}
		if !dup {
			cleanSites = append(cleanSites, p)
		}
	}

	// Prepend 3 super-triangle vertices enclosing [-3w, 4w] x [-3h, 4.5h]
	pts := make([]render.Point, 0, len(cleanSites)+3)
	pts = append(pts,
		render.Pt(-3.0*w, -2.5*h),
		render.Pt(4.0*w, -2.5*h),
		render.Pt(0.5*w, 4.5*h),
	)
	pts = append(pts, cleanSites...)

	tris := []delaunayTri{makeDelaunayTri(pts, 0, 1, 2)}

	for pIdx := 3; pIdx < len(pts); pIdx++ {
		p := pts[pIdx]
		var badEdges []delaunayEdge
		edgeCount := make(map[[2]int]int)
		kept := make([]delaunayTri, 0, len(tris)+8)

		for _, t := range tris {
			if t.degenerate {
				continue
			}
			dx := p.X - t.cx
			dy := p.Y - t.cy
			if dx*dx+dy*dy <= t.rSq+1e-6 {
				e1 := delaunayEdge{u: t.i, v: t.j}
				e2 := delaunayEdge{u: t.j, v: t.k}
				e3 := delaunayEdge{u: t.k, v: t.i}
				badEdges = append(badEdges, e1, e2, e3)
				edgeCount[canonicalEdgeKey(e1.u, e1.v)]++
				edgeCount[canonicalEdgeKey(e2.u, e2.v)]++
				edgeCount[canonicalEdgeKey(e3.u, e3.v)]++
			} else {
				kept = append(kept, t)
			}
		}

		// Boundary of the polygonal cavity consists of edges shared by exactly 1 bad triangle
		for _, e := range badEdges {
			if edgeCount[canonicalEdgeKey(e.u, e.v)] == 1 {
				nt := makeDelaunayTri(pts, e.u, e.v, pIdx)
				if !nt.degenerate {
					kept = append(kept, nt)
				}
			}
		}
		tris = kept
	}

	out := make([][3]render.Point, 0, len(tris))
	for _, t := range tris {
		if t.i < 3 || t.j < 3 || t.k < 3 || t.degenerate {
			continue
		}
		a, b, c := pts[t.i], pts[t.j], pts[t.k]
		area := 0.5 * math.Abs((b.X-a.X)*(c.Y-a.Y)-(b.Y-a.Y)*(c.X-a.X))
		if area > 1.0 {
			out = append(out, [3]render.Point{a, b, c})
		}
	}
	return out
}

// generateStainedGlassSites creates boundary anchor points along all 4 canvas edges plus
// interior stratified + concentric-rosette jittered sites around the focal light center.
func generateStainedGlassSites(pointNum int, jitter, w, h float64, focal render.Point, rng *rand.Rand) []render.Point {
	sites := make([]render.Point, 0, pointNum+64)

	// 1. Exact 4 canvas corners
	sites = append(sites,
		render.Pt(0, 0),
		render.Pt(w, 0),
		render.Pt(w, h),
		render.Pt(0, h),
	)

	// 2. Boundary anchor points along all 4 edges so the triangulation covers 100% of the canvas
	edgeDivs := common.MaxInt(10, int(math.Round(math.Sqrt(float64(pointNum))*0.85)))
	for i := 1; i < edgeDivs; i++ {
		u := (float64(i) + (rng.Float64()-0.5)*0.32*jitter) / float64(edgeDivs)
		u = common.Constrain(u, 0.02, 0.98)
		sites = append(sites,
			render.Pt(u*w, 0),
			render.Pt(u*w, h),
			render.Pt(0, u*h),
			render.Pt(w, u*h),
		)
	}

	// 3. Concentric rose-window / caustic rings around focal light source (28% of pointNum)
	ringBudget := common.MaxInt(18, int(float64(pointNum)*0.28))
	minDim := math.Min(w, h)
	rings := [3]struct {
		rFrac float64
		count int
	}{
		{rFrac: 0.085, count: common.MaxInt(6, ringBudget/5)},
		{rFrac: 0.185, count: common.MaxInt(10, ringBudget/3)},
		{rFrac: 0.310, count: common.MaxInt(12, ringBudget-ringBudget/5-ringBudget/3)},
	}
	for _, rg := range rings {
		rad := rg.rFrac * minDim
		phase := rng.Float64() * 2.0 * math.Pi
		for k := 0; k < rg.count; k++ {
			ang := phase + (float64(k)+rng.NormFloat64()*0.18*jitter)*(2.0*math.Pi/float64(rg.count))
			rr := rad * (1.0 + rng.NormFloat64()*0.12*jitter)
			px := focal.X + rr*math.Cos(ang)
			py := focal.Y + rr*math.Sin(ang)
			if px > 6 && px < w-6 && py > 6 && py < h-6 {
				sites = append(sites, render.Pt(px, py))
			}
		}
	}

	// 4. Stratified jittered grid sites for the remaining interior points
	rem := common.MaxInt(16, pointNum-ringBudget)
	aspect := w / math.Max(1.0, h)
	cols := common.MaxInt(4, int(math.Round(math.Sqrt(float64(rem)*aspect))))
	rows := common.MaxInt(4, int(math.Ceil(float64(rem)/float64(cols))))
	cellW := w / float64(cols)
	cellH := h / float64(rows)
	jit := common.Constrain(jitter, 0.1, 0.92)

	count := 0
	for r := 0; r < rows && count < rem; r++ {
		for c := 0; c < cols && count < rem; c++ {
			jx := (rng.Float64() - 0.5) * jit
			jy := (rng.Float64() - 0.5) * jit
			px := (float64(c) + 0.5 + jx) * cellW
			py := (float64(r) + 0.5 + jy) * cellH
			px = common.Constrain(px, 5.0, w-5.0)
			py = common.Constrain(py, 5.0, h-5.0)
			sites = append(sites, render.Pt(px, py))
			count++
		}
	}

	return sites
}

// Generative renders the Prismatic Delaunay Stained-Glass & Caustic Crystal Mosaic onto c.
func (ds *delaunayStained) Generative(c *generativeart.Canva) {
	pointNum := ds.pointNum
	jitter := ds.jitter
	if pointNum <= 0 {
		pointNum = 220
	}
	if jitter <= 0 {
		jitter = 0.65
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

	seed := int64(20260719)
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	schema := c.Opts().ColorSchema()
	if len(schema) == 0 {
		schema = common.Plasma
	}
	grad := render.NewMultiStopGradient(schema...)
	sunsetGrad := render.NewMultiStopGradient(common.Sunset...)

	rc := c.NewRenderCanvas()

	// Deep cathedral midnight underlayer
	rc.FillBackground(color.RGBA{R: 10, G: 8, B: 18, A: 255})

	// Focal light source at (0.32*w, 0.28*h)
	focal := render.Pt(0.32*fw, 0.28*fh)
	maxDist := math.Hypot(fw, fh) * 0.85
	if maxDist < 1.0 {
		maxDist = 1.0
	}

	// 1. Generate sites and compute exact Bowyer-Watson Delaunay Triangulation
	sites := generateStainedGlassSites(pointNum, jitter, fw, fh, focal, rng)
	triangles := bowyerWatsonTriangulate(sites, fw, fh)

	leadDark := color.RGBA{R: 12, G: 10, B: 18, A: 255}
	goldBevel := color.RGBA{R: 245, G: 210, B: 140, A: 255}
	specularWhite := color.RGBA{R: 255, G: 250, B: 235, A: 255}

	leadPath := &canvas.Path{}
	bevelPath := &canvas.Path{}
	ridgePath := &canvas.Path{}

	// Deduplicated vertex tracker for specular glints
	vertexSet := make(map[[2]int]render.Point)

	// 2. Render each Delaunay crystal facet and its 3 inner gemstone sub-triangles
	for tIdx, tri := range triangles {
		a, b, cPt := tri[0], tri[1], tri[2]
		centroid := render.Pt((a.X+b.X+cPt.X)/3.0, (a.Y+b.Y+cPt.Y)/3.0)
		area := 0.5 * math.Abs((b.X-a.X)*(cPt.Y-a.Y)-(b.Y-a.Y)*(cPt.X-a.X))

		for _, v := range tri {
			key := [2]int{int(math.Round(v.X * 2.0)), int(math.Round(v.Y * 2.0))}
			vertexSet[key] = v
		}

		// Vector and distance from focal light source to triangle centroid
		toCentroid := centroid.Sub(focal)
		distFromLight := toCentroid.Len()
		radNorm := common.Clamp01(distFromLight / maxDist)
		polarAng := math.Atan2(toCentroid.Y, toCentroid.X)
		angNorm := (polarAng + math.Pi) / (2.0 * math.Pi)

		// Longest edge angle gives each crystal facet an intrinsic orientation
		ab := b.Sub(a)
		facetAngle := math.Atan2(ab.Y, ab.X)
		facetMod := 0.5 * (math.Sin(facetAngle*3.0+polarAng) + 1.0)

		// FBM noise at centroid
		fbm := 0.5 * (noise.FBM2D(centroid.X*0.0042, centroid.Y*0.0042, 4, 2.0, 0.5) + 1.0)

		// Sample base jewel color in Oklab space
		colorT := common.Clamp01(0.46*(1.0-radNorm) + 0.34*fbm + 0.14*facetMod + 0.06*math.Sin(angNorm*4.0*math.Pi))
		baseCol := render.LerpOklab(
			grad.SampleOklab(colorT),
			sunsetGrad.SampleOklab(common.Clamp01(1.0-radNorm*0.85)),
			0.22,
		)
		baseLab := render.RGBAToOklab(baseCol)

		// Boost lightness near the focal light source for stained-glass backlight transmission
		backlightBoost := 0.16 * math.Pow(1.0-radNorm, 1.8)
		baseLab.L = common.Constrain(baseLab.L+backlightBoost, 0.14, 0.94)

		// Compute slightly offset interior pyramid apex using safe barycentric weights in [0.23, 0.44]
		lightDir := toCentroid.Norm()
		if lightDir.Len() < 1e-9 {
			lightDir = render.Pt(0.7, 0.7).Norm()
		}
		curl := noise.Curl2D(centroid.X*0.006, centroid.Y*0.006, 0.01)
		wA := 0.333 + 0.08*math.Sin(float64(tIdx)*1.7+curl.X)
		wB := 0.333 + 0.08*math.Cos(float64(tIdx)*2.3+curl.Y)
		wC := 1.0 - wA - wB
		apex := render.Pt(
			wA*a.X+wB*b.X+wC*cPt.X-lightDir.X*2.2*scale,
			wA*a.Y+wB*b.Y+wC*cPt.Y-lightDir.Y*2.2*scale,
		)

		// Subdivide (A, B, C) into 3 inner sub-triangles meeting at apex:
		// (A, B, apex), (B, C, apex), (C, A, apex)
		subTris := [3][3]render.Point{
			{a, b, apex},
			{b, cPt, apex},
			{cPt, a, apex},
		}

		rc.SetBlendMode(render.BlendOver)
		for sIdx, st := range subTris {
			p0, p1, p2 := st[0], st[1], st[2]
			edgeMid := p0.Lerp(p1, 0.5)
			// Outward facet slope normal in 2D (from apex toward outer edge midpoint)
			slopeDir := edgeMid.Sub(p2).Norm()
			// Illumination is highest when facet faces toward the focal light source (-lightDir)
			illum := -(slopeDir.X*lightDir.X + slopeDir.Y*lightDir.Y)

			subL := common.Constrain(baseLab.L+illum*0.17+float64(sIdx-1)*0.03, 0.08, 0.98)
			subChromaScale := 1.0
			if illum > 0.2 {
				subChromaScale = 0.88 // specular desaturation on sunlit facets
			} else if illum < -0.2 {
				subChromaScale = 1.12 // deep saturated jewel shadow
			}
			subCol := render.OklabToRGBA(render.Oklab{
				L: subL,
				A: baseLab.A * subChromaScale,
				B: baseLab.B * subChromaScale,
			}, 255)

			subPoly := []render.Point{p0, p1, p2}
			rc.FillPolygon(subPoly, subCol, 0.95)

			// Inner bevel core triangle inside each sub-facet
			subCent := render.Pt((p0.X+p1.X+p2.X)/3.0, (p0.Y+p1.Y+p2.Y)/3.0)
			innerSub := scalePolyFromCenter(subPoly, subCent, 0.64)
			innerCol := render.LerpOklab(subCol, specularWhite, common.Constrain(0.18+0.25*math.Max(0, illum), 0.08, 0.45))
			rc.FillPolygon(innerSub, innerCol, 0.38)

			// Fine internal crystal striations using rc.DrawSandLine across the sub-facet
			striationCol := render.LerpOklab(subCol, specularWhite, 0.48)
			if illum < 0 {
				striationCol = render.LerpOklab(subCol, goldBevel, 0.35)
			}
			const striationLines = 4
			for sl := 1; sl <= striationLines; sl++ {
				u := float64(sl) / float64(striationLines+1)
				q0 := p2.Lerp(p0, u)
				q1 := p2.Lerp(p1, u)
				segLen := q0.Dist(q1)
				grains := common.MaxInt(12, int(segLen*1.4))
				rc.DrawSandLine(q0, q1, grains, 0.65*scale, striationCol, 0.085, rng)
			}

			// Record pyramid ridge lines from vertices to apex
			ridgePath.MoveTo(p0.X, p0.Y)
			ridgePath.LineTo(p2.X, p2.Y)
		}

		// Wet-on-wet stained glass mottling inside larger triangles
		if area > 140.0*scale*scale {
			rc.SetBlendMode(render.BlendMultiply)
			washPoly := scalePolyFromCenter([]render.Point{a, b, cPt}, centroid, 0.85)
			rc.DrawWatercolorPolygon(washPoly, baseCol, 3, 2, 0.16, 0.05, rng)
		}

		// Internal caustic glow at the gemstone pyramid apex
		rc.SetBlendMode(render.BlendAdd)
		apexGlowR := common.Constrain(math.Sqrt(area)*0.22, 2.2*scale, 11.0*scale)
		apexCol := render.LerpOklab(baseCol, specularWhite, 0.45)
		rc.DrawGlowDot(apex.X, apex.Y, apexGlowR, apexCol, 0.26*(1.1-0.5*radNorm), 0.22)
		rc.SetBlendMode(render.BlendOver)

		// Add outer triangle edges to lead-came path and inner shrunken contour to golden bevel path
		leadPath.MoveTo(a.X, a.Y)
		leadPath.LineTo(b.X, b.Y)
		leadPath.LineTo(cPt.X, cPt.Y)
		leadPath.Close()

		bevelTri := scalePolyFromCenter([]render.Point{a, b, cPt}, centroid, 0.90)
		bevelPath.MoveTo(bevelTri[0].X, bevelTri[0].Y)
		bevelPath.LineTo(bevelTri[1].X, bevelTri[1].Y)
		bevelPath.LineTo(bevelTri[2].X, bevelTri[2].Y)
		bevelPath.Close()
	}

	// 3. Stroke internal pyramid ridges, golden bevel rims, and dark stained-glass lead cames
	rc.SetBlendMode(render.BlendOver)
	rc.StrokePath(ridgePath, 0.72*scale, canvas.RoundCap, canvas.RoundJoin, specularWhite, 0.34)
	rc.StrokePath(bevelPath, 0.68*scale, canvas.RoundCap, canvas.RoundJoin, goldBevel, 0.36)
	rc.StrokePath(leadPath, 2.15*scale, canvas.RoundCap, canvas.RoundJoin, leadDark, 0.92)
	rc.StrokePath(leadPath, 0.58*scale, canvas.RoundCap, canvas.RoundJoin, goldBevel, 0.42)

	// 4. Tiny specular vertex glints at Delaunay vertices + focal caustic halo
	rc.SetBlendMode(render.BlendAdd)
	rc.DrawGlowDot(focal.X, focal.Y, 48.0*scale, specularWhite, 0.14, 0.05)
	for _, v := range vertexSet {
		dNorm := common.Clamp01(v.Dist(focal) / maxDist)
		glintCol := render.LerpOklab(goldBevel, specularWhite, 1.0-0.6*dNorm)
		glintR := (2.2 + 2.4*(1.0-dNorm)) * scale
		rc.DrawGlowDot(v.X, v.Y, glintR, glintCol, 0.48*(1.0-0.45*dNorm), 0.42)
	}
	rc.SetBlendMode(render.BlendOver)

	rc.ApplyBloom(0.55, 7.0, 0.4)
	rc.ApplyVignette(0.25)
	c.MergeRenderCanvasToneMapped(rc, 1.08, 0)
}
