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

type voronoiCrystal struct {
	siteNum    int
	relaxSteps int
}

// NewVoronoiCrystal returns a voronoiCrystal engine that computes Lloyd-relaxed
// Voronoi diagrams via exact half-plane polygon clipping and renders banded agate/geode cells.
func NewVoronoiCrystal(siteNum, relaxSteps int) *voronoiCrystal {
	if siteNum <= 0 {
		siteNum = 85
	}
	if relaxSteps <= 0 {
		relaxSteps = 4
	}
	return &voronoiCrystal{
		siteNum:    siteNum,
		relaxSteps: relaxSteps,
	}
}

// clipPolygonBisector clips a convex polygon against the half-plane closer to siteA than siteB:
// (P - M) . (siteB - siteA) <= 0, where M = 0.5 * (siteA + siteB).
func clipPolygonBisector(poly []render.Point, siteA, siteB render.Point) []render.Point {
	n := len(poly)
	if n == 0 {
		return nil
	}
	mid := siteA.Lerp(siteB, 0.5)
	normal := siteB.Sub(siteA)
	if normal.Len() < 1e-12 {
		return poly
	}

	inside := func(p render.Point) (float64, bool) {
		v := p.Sub(mid)
		d := v.X*normal.X + v.Y*normal.Y
		return d, d <= 1e-9
	}

	out := make([]render.Point, 0, n+2)
	prev := poly[n-1]
	prevD, prevIn := inside(prev)

	for i := 0; i < n; i++ {
		curr := poly[i]
		currD, currIn := inside(curr)

		if currIn {
			if !prevIn {
				denom := prevD - currD
				if math.Abs(denom) > 1e-12 {
					t := prevD / denom
					out = append(out, prev.Lerp(curr, common.Constrain(t, 0, 1)))
				}
			}
			out = append(out, curr)
		} else if prevIn {
			denom := prevD - currD
			if math.Abs(denom) > 1e-12 {
				t := prevD / denom
				out = append(out, prev.Lerp(curr, common.Constrain(t, 0, 1)))
			}
		}

		prev = curr
		prevD = currD
		prevIn = currIn
	}
	return out
}

// computeVoronoiCell computes the exact Voronoi polygon for sites[idx] clipped to [0, w] x [0, h].
func computeVoronoiCell(sites []render.Point, idx int, w, h float64) []render.Point {
	poly := []render.Point{
		{X: 0, Y: 0},
		{X: w, Y: 0},
		{X: w, Y: h},
		{X: 0, Y: h},
	}
	siteA := sites[idx]
	for j, siteB := range sites {
		if j == idx {
			continue
		}
		poly = clipPolygonBisector(poly, siteA, siteB)
		if len(poly) < 3 {
			return nil
		}
	}
	return poly
}

// polygonAreaCentroid computes the signed area and area-weighted centroid of a polygon.
func polygonAreaCentroid(poly []render.Point) (area float64, centroid render.Point) {
	n := len(poly)
	if n < 3 {
		if n > 0 {
			return 0, poly[0]
		}
		return 0, render.Point{}
	}
	var twiceArea, cx, cy float64
	for i := 0; i < n; i++ {
		p0 := poly[i]
		p1 := poly[(i+1)%n]
		cross := p0.X*p1.Y - p1.X*p0.Y
		twiceArea += cross
		cx += (p0.X + p1.X) * cross
		cy += (p0.Y + p1.Y) * cross
	}
	if math.Abs(twiceArea) < 1e-9 {
		var sx, sy float64
		for _, p := range poly {
			sx += p.X
			sy += p.Y
		}
		return 0, render.Pt(sx/float64(n), sy/float64(n))
	}
	inv3A := 1.0 / (3.0 * twiceArea)
	return 0.5 * math.Abs(twiceArea), render.Pt(cx*inv3A, cy*inv3A)
}

// shrinkPolygon scales polygon vertices toward center by factor s in [0, 1].
func shrinkPolygon(poly []render.Point, center render.Point, s float64) []render.Point {
	out := make([]render.Point, len(poly))
	for i, p := range poly {
		out[i] = center.Lerp(p, s)
	}
	return out
}

// Generative renders Lloyd-relaxed Voronoi agate/crystal cells onto c.
func (vc *voronoiCrystal) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	if w <= 0 || h <= 0 {
		return
	}
	scale := math.Min(w, h) / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	seed := rand.Int63()
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	schema := c.Opts().ColorSchema()
	if len(schema) == 0 {
		schema = common.Dunhuang
	}
	grad := render.NewMultiStopGradient(schema...)
	dreamGrad := render.NewMultiStopGradient(common.PastelDream...)

	// Soft mineral parchment backdrop
	bgCol := render.LerpOklab(color.RGBA{R: 244, G: 239, B: 230, A: 255}, grad.SampleOklab(0.5), 0.12)
	rc.FillBackground(bgCol)

	// 1. Generate initial 2D sites in [0, w] x [0, h].
	sites := make([]render.Point, vc.siteNum)
	margin := 12.0 * scale
	for i := 0; i < vc.siteNum; i++ {
		sites[i] = render.Pt(
			margin+rng.Float64()*math.Max(1.0, w-2.0*margin),
			margin+rng.Float64()*math.Max(1.0, h-2.0*margin),
		)
	}

	// 2. Perform Lloyd's Relaxation by moving each site to its exact Voronoi polygon centroid.
	cells := make([][]render.Point, vc.siteNum)
	for step := 0; step < vc.relaxSteps; step++ {
		for i := range sites {
			cell := computeVoronoiCell(sites, i, w, h)
			cells[i] = cell
			if len(cell) >= 3 {
				_, centroid := polygonAreaCentroid(cell)
				sites[i] = render.Pt(
					common.Constrain(centroid.X, 1.0, w-1.0),
					common.Constrain(centroid.Y, 1.0, h-1.0),
				)
			}
		}
	}
	for i := range sites {
		cells[i] = computeVoronoiCell(sites, i, w, h)
	}

	cx, cy := w*0.5, h*0.5
	maxRad := math.Hypot(cx, cy)
	if maxRad < 1.0 {
		maxRad = 1.0
	}
	lightDir := render.Pt(-0.65, -0.76).Norm()

	// 3. Render each Voronoi crystal cell.
	for i, poly := range cells {
		if len(poly) < 3 {
			continue
		}
		_, centroid := polygonAreaCentroid(poly)

		dx := centroid.X - cx
		dy := centroid.Y - cy
		radNorm := common.Constrain(math.Hypot(dx, dy)/maxRad, 0, 1)
		angNorm := (math.Atan2(dy, dx) + math.Pi) / (2.0 * math.Pi)
		nVal := 0.5 * (noise.FBM2D(centroid.X*0.005, centroid.Y*0.005, 3, 2.0, 0.5) + 1.0)

		colorT := math.Mod(0.52*radNorm+0.28*angNorm+0.35*nVal+float64(i%7)*0.03, 1.0)
		baseCol := render.LerpOklab(grad.SampleOklab(colorT), dreamGrad.SampleOklab(1.0-colorT), 0.25)
		baseLab := render.RGBAToOklab(baseCol)

		darkRimCol := render.OklabToRGBA(render.Oklab{
			L: common.Constrain(baseLab.L*0.62, 0.08, 0.92),
			A: baseLab.A * 1.08,
			B: baseLab.B * 1.08,
		}, 255)
		lightCoreCol := render.OklabToRGBA(render.Oklab{
			L: common.Constrain(baseLab.L+0.22, 0.15, 0.97),
			A: baseLab.A * 0.75,
			B: baseLab.B * 0.75,
		}, 255)

		// Base polygon fill
		rc.SetBlendMode(render.BlendOver)
		rc.FillPolygon(poly, darkRimCol, 0.88)

		// Directional bevel facets from centroid to adjacent vertices
		for v := 0; v < len(poly); v++ {
			p0 := poly[v]
			p1 := poly[(v+1)%len(poly)]
			edgeMid := p0.Lerp(p1, 0.5)
			facetNorm := edgeMid.Sub(centroid).Norm()
			illum := facetNorm.X*lightDir.X + facetNorm.Y*lightDir.Y

			facetCol := render.OklabToRGBA(render.Oklab{
				L: common.Constrain(baseLab.L+illum*0.14, 0.06, 0.98),
				A: baseLab.A,
				B: baseLab.B,
			}, 255)
			rc.FillPolygon([]render.Point{centroid, p0, p1}, facetCol, 0.62)

			// Crisp facet ridge line from centroid to vertex
			ridgeCol := lightCoreCol
			if illum < 0 {
				ridgeCol = darkRimCol
			}
			rc.StrokePolyline([]render.Point{centroid, p0}, 0.65*scale, ridgeCol, 0.28)
		}

		// Concentric agate/geode crystal growth bands inside the cell
		const bandCount = 6
		for b := 1; b <= bandCount; b++ {
			u := float64(b) / float64(bandCount+1)
			shrink := 1.0 - u*0.78
			bandPoly := shrinkPolygon(poly, centroid, shrink)

			var bandCol color.RGBA
			if b%2 == 0 {
				bandCol = render.LerpOklab(baseCol, lightCoreCol, u*0.85)
			} else {
				bandCol = render.LerpOklab(darkRimCol, lightCoreCol, u*0.55)
			}
			rc.FillPolygon(bandPoly, bandCol, 0.46)
			rc.StrokePolyline(append(bandPoly, bandPoly[0]), 0.55*scale, lightCoreCol, 0.22)
		}

		// Soft wet-on-wet watercolor pigment wash inside the cell
		rc.SetBlendMode(render.BlendMultiply)
		washPoly := shrinkPolygon(poly, centroid, 0.86)
		rc.DrawWatercolorPolygon(washPoly, baseCol, 6, 3, 0.24, 0.055, rng)

		// Crystalline core highlight at cell centroid
		rc.SetBlendMode(render.BlendScreen)
		rc.DrawGlowDot(centroid.X, centroid.Y, (4.5+rng.Float64()*5.5)*scale, lightCoreCol, 0.42, 0.25)

		// Crisp mineral vein boundary stroked with tdewolff/canvas Path + rc.StrokePath
		rc.SetBlendMode(render.BlendOver)
		cellPath := &canvas.Path{}
		cellPath.MoveTo(poly[0].X, poly[0].Y)
		for v := 1; v < len(poly); v++ {
			cellPath.LineTo(poly[v].X, poly[v].Y)
		}
		cellPath.Close()

		veinCol := render.OklabToRGBA(render.Oklab{
			L: common.Constrain(baseLab.L*0.38, 0.05, 0.35),
			A: baseLab.A * 0.6,
			B: baseLab.B * 0.6,
		}, 255)
		rc.StrokePath(cellPath, 1.35*scale, canvas.RoundCap, canvas.RoundJoin, veinCol, 0.82)
	}

	rc.ApplyPaperTexture(0.35, 42)
	rc.ApplyVignette(0.22)
	c.MergeRenderCanvas(rc)
}
