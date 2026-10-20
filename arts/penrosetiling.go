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

type penroseTiling struct {
	subdivisions int
}

type penroseTri struct {
	triType int // 0: acute half-thick rhombus (36-72-72), 1: obtuse half-thin rhombus (108-36-36)
	a, b, c render.Point
}

// NewPenroseTiling returns a penroseTiling engine that generates an exact 5-fold symmetric
// Penrose P2/P3 quasicrystal tiling via Robinson golden triangle deflation with Oklab
// 3D-beveled jewel facets and continuous Ammann/Truchet circular rosette arcs.
func NewPenroseTiling(subdivisions int) *penroseTiling {
	if subdivisions <= 0 {
		subdivisions = 6
	}
	subdivisions = common.ConstrainInt(subdivisions, 3, 8)
	return &penroseTiling{
		subdivisions: subdivisions,
	}
}

// subdividePenroseP3 performs one iteration of exact Penrose P3 Robinson triangle deflation.
//   - Type 0 (acute isosceles, sides phi:phi:1, apex A=36°, base B=72°, C=72°)
//     subdivides into 2 Type 0 triangles and 1 Type 1 triangle.
//   - Type 1 (obtuse isosceles, sides 1:1:phi, apex A=108°, base B=36°, C=36°)
//     subdivides into 1 Type 0 triangle and 1 Type 1 triangle.
func subdividePenroseP3(tris []penroseTri, invPhi, invPhiSq float64) []penroseTri {
	out := make([]penroseTri, 0, len(tris)*3)
	for _, t := range tris {
		if t.triType == 0 {
			// Split AB at P (1/phi from A) and AC at Q (1/phi^2 from A)
			p := t.a.Lerp(t.b, invPhi)
			q := t.a.Lerp(t.c, invPhiSq)
			out = append(out,
				penroseTri{triType: 0, a: t.c, b: p, c: q},
				penroseTri{triType: 0, a: t.c, b: p, c: t.b},
				penroseTri{triType: 1, a: q, b: p, c: t.a},
			)
		} else {
			// Split base BC at R (1/phi from B)
			r := t.b.Lerp(t.c, invPhi)
			out = append(out,
				penroseTri{triType: 0, a: t.b, b: r, c: t.a},
				penroseTri{triType: 1, a: r, b: t.c, c: t.a},
			)
		}
	}
	return out
}

// penroseVertexArcPath builds an exact circular arc centered at vertex v that connects
// rays v->p1 and v->p2 at the given radius using cubic Bézier segments.
func penroseVertexArcPath(v, p1, p2 render.Point, radius float64) *canvas.Path {
	if radius <= 0.25 {
		return &canvas.Path{}
	}
	a0 := math.Atan2(p1.Y-v.Y, p1.X-v.X)
	a1 := math.Atan2(p2.Y-v.Y, p2.X-v.X)
	diff := a1 - a0
	for diff > math.Pi {
		diff -= 2.0 * math.Pi
	}
	for diff < -math.Pi {
		diff += 2.0 * math.Pi
	}
	return buildArcPath(v.X, v.Y, radius, a0, a0+diff)
}

// Generative renders the Penrose quasicrystal tiling with 3D-beveled Oklab jewel facets
// and luminous screen/additive Truchet-Ammann circular rosettes.
func (pt *penroseTiling) Generative(c *generativeart.Canva) {
	subs := pt.subdivisions
	if subs <= 0 {
		subs = 6
	}
	subs = common.ConstrainInt(subs, 3, 8)

	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	cx, cy := fw*0.5, fh*0.5
	center := render.Pt(cx, cy)
	scale := math.Min(fw, fh) / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	phi := (1.0 + math.Sqrt(5.0)) * 0.5
	invPhi := 1.0 / phi
	invPhiSq := invPhi * invPhi

	schema := c.Opts().ColorSchema()
	if len(schema) < 2 {
		schema = common.Oceanic
	}
	grad := render.NewMultiStopGradient(schema...)
	dunhuangGrad := render.NewMultiStopGradient(common.Dunhuang...)

	rc := c.NewRenderCanvas()

	// 1. Deep cosmic obsidian-sapphire background before tiling
	bgCore := render.LerpOklab(grad.SampleOklab(0.08), color.RGBA{R: 10, G: 14, B: 26, A: 255}, 0.72)
	rc.FillBackground(bgCore)

	// 2. Construct the initial decagonal wheel of 10 acute Robinson triangles (Type 0)
	// with outer radius = 0.72 * hypot(w, h) and alternating mirror symmetry.
	wheelRadius := 0.72 * math.Hypot(fw, fh)
	tris := make([]penroseTri, 0, 10)
	sectorStep := 2.0 * math.Pi / 10.0

	for i := 0; i < 10; i++ {
		ang0 := float64(i)*sectorStep - math.Pi*0.5
		ang1 := float64(i+1)*sectorStep - math.Pi*0.5
		pB := render.Pt(cx+wheelRadius*math.Cos(ang0), cy+wheelRadius*math.Sin(ang0))
		pC := render.Pt(cx+wheelRadius*math.Cos(ang1), cy+wheelRadius*math.Sin(ang1))
		if i%2 == 1 {
			pB, pC = pC, pB
		}
		tris = append(tris, penroseTri{
			triType: 0,
			a:       center,
			b:       pB,
			c:       pC,
		})
	}

	// 3. Deflate Robinson triangles for pt.subdivisions iterations
	for step := 0; step < subs; step++ {
		tris = subdividePenroseP3(tris, invPhi, invPhiSq)
	}

	maxCanvasR := math.Hypot(cx, cy)
	lightDir := render.Pt(-0.62, -0.78).Norm()
	margin := 48.0 * scale

	// Paths accumulated for batch stroking of crisp tile edges and golden Truchet-Ammann arcs
	edgePath := &canvas.Path{}
	innerFacetPath := &canvas.Path{}
	primaryArcPath := &canvas.Path{}
	secondaryArcPath := &canvas.Path{}

	var accentCenters []render.Point
	var accentColors []color.RGBA

	// 4. Render each Robinson triangle with 3D-beveled Oklab jewel facets
	rc.SetBlendMode(render.BlendOver)
	for idx, t := range tris {
		centroid := render.Pt(
			(t.a.X+t.b.X+t.c.X)/3.0,
			(t.a.Y+t.b.Y+t.c.Y)/3.0,
		)
		// Cull triangles completely outside the visible canvas + margin
		if centroid.X < -margin && t.a.X < -margin && t.b.X < -margin && t.c.X < -margin {
			continue
		}
		if centroid.X > fw+margin && t.a.X > fw+margin && t.b.X > fw+margin && t.c.X > fw+margin {
			continue
		}
		if centroid.Y < -margin && t.a.Y < -margin && t.b.Y < -margin && t.c.Y < -margin {
			continue
		}
		if centroid.Y > fh+margin && t.a.Y > fh+margin && t.b.Y > fh+margin && t.c.Y > fh+margin {
			continue
		}

		radDist := centroid.Dist(center)
		radNorm := common.Clamp01(radDist / math.Max(1.0, maxCanvasR))

		// Orientation of apex -> base midpoint vector gives 5-fold / 10-fold directional iridescence
		baseMid := t.b.Lerp(t.c, 0.5)
		axisVec := baseMid.Sub(t.a).Norm()
		orientAngle := math.Atan2(axisVec.Y, axisVec.X)
		fiveFoldHarmonic := 0.5 + 0.5*math.Cos(5.0*orientAngle)
		tenFoldRing := 0.5 + 0.5*math.Cos(10.0*radNorm*math.Pi-2.0*fiveFoldHarmonic)

		var baseCol color.RGBA
		if t.triType == 0 {
			// Half-thick rhombus (acute): rich sapphire/teal/gold jewel tones
			colorT := common.Clamp01(0.14 + 0.58*radNorm + 0.22*fiveFoldHarmonic - 0.08*tenFoldRing)
			cMain := grad.SampleOklab(colorT)
			cWarm := dunhuangGrad.SampleOklab(common.Clamp01(0.25 + 0.55*fiveFoldHarmonic))
			baseCol = render.LerpOklab(cMain, cWarm, 0.24*(1.0-radNorm*0.5))
		} else {
			// Half-thin rhombus (obtuse): contrasting luminous amber/jade/cyan facets
			colorT := common.Clamp01(0.86 - 0.52*radNorm + 0.16*fiveFoldHarmonic)
			cMain := grad.SampleOklab(colorT)
			cWarm := dunhuangGrad.SampleOklab(common.Clamp01(0.15 + 0.65*(1.0-radNorm)))
			baseCol = render.LerpOklab(cMain, cWarm, 0.42)
		}

		// Directional 3D bevel shading in Oklab perceptual space
		dotLight := axisVec.X*lightDir.X + axisVec.Y*lightDir.Y
		lab := render.RGBAToOklab(baseCol)

		// Outer base bevel fill
		outerLab := render.Oklab{
			L: common.Constrain(lab.L*0.86+0.07*dotLight, 0.08, 0.94),
			A: lab.A * 1.04,
			B: lab.B * 1.04,
		}
		outerCol := render.OklabToRGBA(outerLab, 255)
		rc.FillPolygon([]render.Point{t.a, t.b, t.c}, outerCol, 0.96)

		// Split triangle into two 3D-beveled sub-facets along the symmetry axis (a -> baseMid)
		leftNorm := t.b.Sub(centroid).Norm()
		rightNorm := t.c.Sub(centroid).Norm()
		leftSpec := leftNorm.X*lightDir.X + leftNorm.Y*lightDir.Y
		rightSpec := rightNorm.X*lightDir.X + rightNorm.Y*lightDir.Y

		leftCol := render.OklabToRGBA(render.Oklab{
			L: common.Constrain(lab.L+0.11*leftSpec+0.04, 0.10, 0.97),
			A: lab.A,
			B: lab.B,
		}, 255)
		rightCol := render.OklabToRGBA(render.Oklab{
			L: common.Constrain(lab.L+0.11*rightSpec-0.03, 0.08, 0.95),
			A: lab.A,
			B: lab.B,
		}, 255)

		rc.FillPolygon([]render.Point{t.a, t.b, centroid}, leftCol, 0.82)
		rc.FillPolygon([]render.Point{t.a, centroid, t.c}, rightCol, 0.82)

		// Inner jewel table facet (scaled toward centroid)
		inA := centroid.Lerp(t.a, 0.56)
		inB := centroid.Lerp(t.b, 0.56)
		inC := centroid.Lerp(t.c, 0.56)
		gemLab := render.Oklab{
			L: common.Constrain(lab.L+0.14*(1.0-0.45*radNorm)+0.05*fiveFoldHarmonic, 0.18, 0.98),
			A: lab.A * 0.92,
			B: lab.B * 0.92,
		}
		rc.FillPolygon([]render.Point{inA, inB, inC}, render.OklabToRGBA(gemLab, 255), 0.76)

		// Accumulate outer tile edges (skip the shared rhombus diagonal b-c half the time or stroke full lattice cleanly)
		edgePath.MoveTo(t.a.X, t.a.Y)
		edgePath.LineTo(t.b.X, t.b.Y)
		edgePath.LineTo(t.c.X, t.c.Y)
		edgePath.Close()

		innerFacetPath.MoveTo(inA.X, inA.Y)
		innerFacetPath.LineTo(inB.X, inB.Y)
		innerFacetPath.LineTo(inC.X, inC.Y)
		innerFacetPath.Close()

		// 5. Build continuous circular Truchet / Ammann arcs connecting golden-ratio edge split points (1/phi and 1/phi^2).
		// Because each arc is centered at a triangle vertex and has radius equal to the exact golden split distance along both
		// incident edges, every arc meets the triangle edges at 90 degrees and links seamlessly across shared edges!
		lenAB := t.a.Dist(t.b)
		if t.triType == 0 {
			// Acute triangle: |AB| = |AC| = L, |BC| = L / phi.
			// Primary arc centered at apex A with radius L / phi (links AB and AC at 1/phi from A).
			arcA1 := penroseVertexArcPath(t.a, t.b, t.c, lenAB*invPhi)
			primaryArcPath = primaryArcPath.Join(arcA1)

			// Complementary inner arc centered at apex A with radius L / phi^2 (links AB and AC at 1/phi^2 from A).
			arcA2 := penroseVertexArcPath(t.a, t.b, t.c, lenAB*invPhiSq)
			secondaryArcPath = secondaryArcPath.Join(arcA2)

			// Base arcs centered at B and C with radius L / phi^2 (links AB<->BC and AC<->BC at 1/phi^2 from B and C).
			arcB := penroseVertexArcPath(t.b, t.a, t.c, lenAB*invPhiSq)
			arcC := penroseVertexArcPath(t.c, t.a, t.b, lenAB*invPhiSq)
			primaryArcPath = primaryArcPath.Join(arcB)
			primaryArcPath = primaryArcPath.Join(arcC)
		} else {
			// Obtuse triangle: |AB| = |AC| = L_leg, |BC| = phi * L_leg.
			// Primary arc centered at obtuse apex A with radius L_leg / phi (links AB and AC at 1/phi from A).
			arcA := penroseVertexArcPath(t.a, t.b, t.c, lenAB*invPhi)
			primaryArcPath = primaryArcPath.Join(arcA)

			// Secondary arcs centered at acute base vertices B and C with radius L_leg / phi^2.
			arcB := penroseVertexArcPath(t.b, t.a, t.c, lenAB*invPhiSq)
			arcC := penroseVertexArcPath(t.c, t.a, t.b, lenAB*invPhiSq)
			secondaryArcPath = secondaryArcPath.Join(arcB)
			secondaryArcPath = secondaryArcPath.Join(arcC)
		}

		// Collect quasicrystal star-vertex accent nodes on inner rings
		if idx%5 == 0 && radNorm < 0.88 {
			accentCenters = append(accentCenters, t.a)
			accentColors = append(accentColors, render.OklabToRGBA(render.Oklab{
				L: 0.92,
				A: lab.A * 0.6,
				B: lab.B * 0.6,
			}, 255))
		}
	}

	// 6. Stroke the quasicrystal tile lattice and inner bevel facets
	lineW := math.Max(0.65, (2.4-0.22*float64(subs))*scale)
	darkBorderCol := color.RGBA{R: 12, G: 18, B: 30, A: 255}
	goldRimCol := render.LerpOklab(dunhuangGrad.SampleOklab(0.45), color.RGBA{R: 245, G: 225, B: 175, A: 255}, 0.55)

	rc.SetBlendMode(render.BlendMultiply)
	rc.StrokePath(edgePath, lineW*1.85, canvas.RoundCap, canvas.RoundJoin, darkBorderCol, 0.58)

	rc.SetBlendMode(render.BlendOver)
	rc.StrokePath(edgePath, lineW*0.85, canvas.RoundCap, canvas.RoundJoin, darkBorderCol, 0.82)
	rc.StrokePath(innerFacetPath, math.Max(0.45, lineW*0.52), canvas.RoundCap, canvas.RoundJoin, goldRimCol, 0.34)

	// 7. Render the unbroken 5-fold Truchet / Ammann circular rosettes and ribbons in Screen & Add blend modes
	arcColorOuter := render.LerpOklab(grad.SampleOklab(0.78), color.RGBA{R: 110, G: 230, B: 255, A: 255}, 0.45)
	arcColorCore := render.LerpOklab(dunhuangGrad.SampleOklab(0.50), color.RGBA{R: 255, G: 242, B: 205, A: 255}, 0.65)
	secondaryArcCol := render.LerpOklab(grad.SampleOklab(0.62), color.RGBA{R: 255, G: 195, B: 125, A: 255}, 0.50)

	rc.SetBlendMode(render.BlendScreen)
	rc.StrokePath(secondaryArcPath, lineW*1.55, canvas.RoundCap, canvas.RoundJoin, secondaryArcCol, 0.46)
	rc.StrokePath(primaryArcPath, lineW*2.45, canvas.RoundCap, canvas.RoundJoin, arcColorOuter, 0.52)

	rc.SetBlendMode(render.BlendAdd)
	rc.StrokePath(primaryArcPath, lineW*1.05, canvas.RoundCap, canvas.RoundJoin, arcColorCore, 0.72)
	rc.StrokePath(secondaryArcPath, lineW*0.65, canvas.RoundCap, canvas.RoundJoin, color.RGBA{R: 255, G: 250, B: 235, A: 255}, 0.42)

	// 8. Central decagonal rosette halo and vertex star glints
	rc.SetBlendMode(render.BlendAdd)
	ringRadii := [3]float64{wheelRadius * math.Pow(invPhi, float64(subs-1)), wheelRadius * math.Pow(invPhi, float64(subs-2)), wheelRadius * math.Pow(invPhi, float64(subs-3))}
	for i, rr := range ringRadii {
		if rr > 6.0*scale && rr < maxCanvasR*0.92 {
			rc.DrawSoftRing(cx, cy, rr, (1.8+float64(i)*0.6)*scale, arcColorCore, 0.22)
		}
	}

	rng := rand.New(rand.NewSource(int64(subs)*9173 + 42))
	maxGlints := common.MinInt(len(accentCenters), 180)
	stepStride := common.MaxInt(1, len(accentCenters)/maxGlints)
	for i := 0; i < len(accentCenters); i += stepStride {
		p := accentCenters[i]
		if p.X >= 0 && p.X <= fw && p.Y >= 0 && p.Y <= fh {
			rad := (1.8 + 1.6*rng.Float64()) * scale
			rc.DrawGlowDot(p.X, p.Y, rad, accentColors[i], 0.32, 0.25)
		}
	}
	// Central quasicrystal sun-star core
	rc.DrawGlowDot(cx, cy, 9.5*scale, arcColorCore, 0.65, 0.35)

	// 9. Post-processing & ACES filmic tone-mapped resolve
	rc.ApplyBloom(0.52, 8.0, 0.45)
	rc.ApplyVignette(0.28)
	c.MergeRenderCanvasToneMapped(rc, 1.08, 0)
}
