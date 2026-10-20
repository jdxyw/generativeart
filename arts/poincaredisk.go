package arts

import (
	"image/color"
	"math"
	"math/cmplx"
	"sort"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type poincareDisk struct {
	p, q, maxDepth int
}

// NewPoincareDisk returns a poincareDisk engine that generates an exact
// hyperbolic {p, q} tessellation inside the Poincare unit disk via Mobius
// circle inversions across orthogonal hyperbolic geodesics.
func NewPoincareDisk(p, q, maxDepth int) *poincareDisk {
	if p <= 0 {
		p = 7
	}
	if q <= 0 {
		q = 3
	}
	if maxDepth <= 0 {
		maxDepth = 5
	}
	if (p-2)*(q-2) <= 4 {
		p = 7
		q = 3
	}
	return &poincareDisk{
		p:        p,
		q:        q,
		maxDepth: maxDepth,
	}
}

type hyperbolicTile struct {
	verts  []complex128
	center complex128
	layer  int
	parity int
}

// geodesicCircle computes the Euclidean circle (center C, radius R) orthogonal to
// the unit circle |z| = 1 passing through points z1 and z2 inside the unit disk.
// If z1 and z2 are nearly collinear with the origin, isLine is true.
func geodesicCircle(z1, z2 complex128) (c complex128, r float64, isLine bool) {
	x1, y1 := real(z1), imag(z1)
	x2, y2 := real(z2), imag(z2)
	det := x1*y2 - y1*x2
	if math.Abs(det) < 1e-9 {
		return 0, 0, true
	}
	u1 := 0.5 * (x1*x1 + y1*y1 + 1.0)
	u2 := 0.5 * (x2*x2 + y2*y2 + 1.0)
	cx := (u1*y2 - u2*y1) / det
	cy := (x1*u2 - x2*u1) / det
	rSq := cx*cx + cy*cy - 1.0
	if rSq <= 1e-12 {
		return 0, 0, true
	}
	return complex(cx, cy), math.Sqrt(rSq), false
}

// reflectGeodesic reflects point z across the hyperbolic geodesic passing through (z1, z2).
func reflectGeodesic(z, z1, z2 complex128) complex128 {
	c, r, isLine := geodesicCircle(z1, z2)
	if isLine {
		dir := z2 - z1
		norm := cmplx.Abs(dir)
		if norm < 1e-12 {
			return z
		}
		u := dir / complex(norm, 0)
		return u * u * cmplx.Conj(z)
	}
	diff := cmplx.Conj(z - c)
	if cmplx.Abs(diff) < 1e-12 {
		return z
	}
	return c + complex(r*r, 0)/diff
}

// sampleGeodesicArc samples nSteps points along the hyperbolic geodesic arc from z1 to z2
// (excluding the endpoint z2 so adjacent polygon edges concatenate cleanly).
func sampleGeodesicArc(z1, z2 complex128, nSteps int) []complex128 {
	if nSteps < 1 {
		nSteps = 1
	}
	c, r, isLine := geodesicCircle(z1, z2)
	out := make([]complex128, 0, nSteps)
	if isLine {
		for i := 0; i < nSteps; i++ {
			t := float64(i) / float64(nSteps)
			out = append(out, z1+complex(t, 0)*(z2-z1))
		}
		return out
	}
	ang1 := math.Atan2(imag(z1-c), real(z1-c))
	ang2 := math.Atan2(imag(z2-c), real(z2-c))
	dAng := ang2 - ang1
	for dAng <= -math.Pi {
		dAng += 2.0 * math.Pi
	}
	for dAng > math.Pi {
		dAng -= 2.0 * math.Pi
	}
	for i := 0; i < nSteps; i++ {
		t := float64(i) / float64(nSteps)
		ang := ang1 + t*dAng
		out = append(out, c+complex(r*math.Cos(ang), r*math.Sin(ang)))
	}
	return out
}

// hyperbolicMidpoint returns the exact hyperbolic midpoint between z1 and z2 on their geodesic arc.
func hyperbolicMidpoint(z1, z2 complex128) complex128 {
	c, r, isLine := geodesicCircle(z1, z2)
	if isLine {
		return 0.5 * (z1 + z2)
	}
	ang1 := math.Atan2(imag(z1-c), real(z1-c))
	ang2 := math.Atan2(imag(z2-c), real(z2-c))
	dAng := ang2 - ang1
	for dAng <= -math.Pi {
		dAng += 2.0 * math.Pi
	}
	for dAng > math.Pi {
		dAng -= 2.0 * math.Pi
	}
	midAng := ang1 + 0.5*dAng
	return c + complex(r*math.Cos(midAng), r*math.Sin(midAng))
}

// buildHyperbolicTiling performs breadth-first hyperbolic reflections across polygon edges
// up to maxDepth layers, deduplicating tiles by quantized Euclidean vertex centroid.
func buildHyperbolicTiling(p, q, maxDepth int) []hyperbolicTile {
	r0 := math.Sqrt(math.Cos(math.Pi/float64(p)+math.Pi/float64(q)) / math.Cos(math.Pi/float64(p)-math.Pi/float64(q)))
	centralVerts := make([]complex128, p)
	for k := 0; k < p; k++ {
		theta := 2.0*math.Pi*float64(k)/float64(p) + math.Pi/float64(p)
		centralVerts[k] = complex(r0*math.Cos(theta), r0*math.Sin(theta))
	}

	seen := make(map[[2]int64]bool)
	const quantScale = 8000.0

	polyCentroid := func(verts []complex128) complex128 {
		var sum complex128
		for _, v := range verts {
			sum += v
		}
		return sum / complex(float64(len(verts)), 0)
	}

	isSeenOrMark := func(c complex128) bool {
		kx := int64(math.Round(real(c) * quantScale))
		ky := int64(math.Round(imag(c) * quantScale))
		for dx := int64(-1); dx <= 1; dx++ {
			for dy := int64(-1); dy <= 1; dy++ {
				if seen[[2]int64{kx + dx, ky + dy}] {
					return true
				}
			}
		}
		seen[[2]int64{kx, ky}] = true
		return false
	}

	root := hyperbolicTile{
		verts:  centralVerts,
		center: 0,
		layer:  0,
		parity: 0,
	}
	isSeenOrMark(polyCentroid(centralVerts))

	tiles := []hyperbolicTile{root}
	queue := []hyperbolicTile{root}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		if curr.layer >= maxDepth {
			continue
		}

		for i := 0; i < p; i++ {
			z1 := curr.verts[i]
			z2 := curr.verts[(i+1)%p]

			newVerts := make([]complex128, p)
			maxAbs := 0.0
			valid := true
			for j := 0; j < p; j++ {
				rv := reflectGeodesic(curr.verts[j], z1, z2)
				absV := cmplx.Abs(rv)
				if math.IsNaN(absV) || absV >= 0.9985 {
					valid = false
					break
				}
				if absV > maxAbs {
					maxAbs = absV
				}
				// Reverse order so every tile preserves counter-clockwise orientation
				newVerts[p-1-j] = rv
			}
			if !valid {
				continue
			}

			// Estimate Euclidean chord span so microscopic boundary slivers are skipped
			perim := 0.0
			for j := 0; j < p; j++ {
				perim += cmplx.Abs(newVerts[(j+1)%p] - newVerts[j])
			}
			if perim < 0.012 {
				continue
			}

			vMean := polyCentroid(newVerts)
			if cmplx.Abs(vMean) >= 0.992 || isSeenOrMark(vMean) {
				continue
			}

			newCenter := reflectGeodesic(curr.center, z1, z2)
			if cmplx.Abs(newCenter) >= 0.995 {
				newCenter = vMean
			}

			child := hyperbolicTile{
				verts:  newVerts,
				center: newCenter,
				layer:  curr.layer + 1,
				parity: (curr.parity + 1 + i%2) % 2,
			}
			tiles = append(tiles, child)
			queue = append(queue, child)
		}
	}

	return tiles
}

// Generative renders the Hyperbolic {p, q} Poincare Disk tessellation into c.
func (pd *poincareDisk) Generative(c *generativeart.Canva) {
	p, q, maxDepth := pd.p, pd.q, pd.maxDepth
	if p <= 0 {
		p = 7
	}
	if q <= 0 {
		q = 3
	}
	if maxDepth <= 0 {
		maxDepth = 5
	}
	if (p-2)*(q-2) <= 4 {
		p = 7
		q = 3
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
	cx, cy := fw*0.5, fh*0.5
	diskR := 0.435 * minDim

	schema := c.Opts().ColorSchema()
	if len(schema) == 0 {
		schema = common.Dunhuang
	}
	grad := render.NewMultiStopGradient(schema...)
	infernoGrad := render.NewMultiStopGradient(common.Inferno...)

	rc := c.NewRenderCanvas()

	// 1. Deep celestial obsidian backdrop with subtle radial nebula aura
	bgCol := color.RGBA{R: 8, G: 11, B: 22, A: 255}
	rc.FillBackground(bgCol)

	rc.SetBlendMode(render.BlendAdd)
	rc.DrawGlowDot(cx, cy, diskR*1.12, grad.SampleOklab(0.22), 0.18, 0.0)
	rc.DrawGlowDot(cx, cy, diskR*0.72, infernoGrad.SampleOklab(0.55), 0.14, 0.0)
	rc.SetBlendMode(render.BlendOver)

	// 2. Generate exact Hyperbolic {p, q} tessellation tiles
	tiles := buildHyperbolicTiling(p, q, maxDepth)

	// Sort tiles from outer boundary layers inward so central tiles and their crisp gold borders sit cleanly on top
	sort.Slice(tiles, func(i, j int) bool {
		if tiles[i].layer != tiles[j].layer {
			return tiles[i].layer > tiles[j].layer
		}
		return cmplx.Abs(tiles[i].center) > cmplx.Abs(tiles[j].center)
	})

	goldStroke := color.RGBA{R: 244, G: 206, B: 132, A: 255}
	ivoryStroke := color.RGBA{R: 255, G: 246, B: 224, A: 255}
	darkCasing := color.RGBA{R: 10, G: 13, B: 24, A: 255}

	toScreen := func(z complex128) render.Point {
		return render.Pt(cx+real(z)*diskR, cy+imag(z)*diskR)
	}

	addCleanPolyline := func(dst *canvas.Path, pts []render.Point, closePath bool) {
		if len(pts) < 2 {
			return
		}
		dst.MoveTo(pts[0].X, pts[0].Y)
		last := pts[0]
		count := 1
		for i := 1; i < len(pts); i++ {
			if last.Dist(pts[i]) < 0.35 {
				continue
			}
			dst.LineTo(pts[i].X, pts[i].Y)
			last = pts[i]
			count++
		}
		if closePath && count >= 3 {
			dst.Close()
		}
	}

	// 3. Render each hyperbolic polygon tile
	for _, tile := range tiles {
		euclidDist := common.Clamp01(cmplx.Abs(tile.center))
		// Exact Poincare hyperbolic distance d_H = 2 * atanh(|z|)
		hypDist := 2.0 * math.Atanh(math.Min(0.992, euclidDist))
		hypNorm := common.Clamp01(hypDist / 5.2)
		layerNorm := common.Clamp01(float64(tile.layer) / float64(common.MaxInt(maxDepth, 1)))
		polarAng := (math.Atan2(imag(tile.center), real(tile.center)) + math.Pi) / (2.0 * math.Pi)

		// Sample geodesic arc edges to construct a smooth screen-space polygon
		stepsPerEdge := 16
		if tile.layer <= 2 {
			stepsPerEdge = 24
		} else if tile.layer >= 4 {
			stepsPerEdge = 10
		}

		polyPts := make([]render.Point, 0, p*stepsPerEdge)
		for i := 0; i < p; i++ {
			z1 := tile.verts[i]
			z2 := tile.verts[(i+1)%p]
			arcZ := sampleGeodesicArc(z1, z2, stepsPerEdge)
			for _, z := range arcZ {
				polyPts = append(polyPts, toScreen(z))
			}
		}

		// Compute jewel-like Oklab tile palette modulated by hyperbolic radius, layer parity, and azimuth
		tBase := common.Clamp01(0.52*hypNorm + 0.28*layerNorm + 0.12*math.Sin(polarAng*float64(p)*2.0*math.Pi) + 0.08)
		if tile.parity == 1 {
			tBase = common.Clamp01(1.0 - 0.65*tBase + 0.18*layerNorm)
		}
		tileCol := grad.SampleOklab(tBase)
		accentCol := infernoGrad.SampleOklab(common.Clamp01(0.85 - 0.55*hypNorm))

		lab := render.RGBAToOklab(tileCol)
		outerRimCol := render.OklabToRGBA(render.Oklab{
			L: common.Constrain(lab.L*0.62, 0.08, 0.85),
			A: lab.A * 1.12,
			B: lab.B * 1.12,
		}, 255)
		innerJewelCol := render.LerpOklab(
			render.OklabToRGBA(render.Oklab{
				L: common.Constrain(lab.L+0.18, 0.20, 0.95),
				A: lab.A * 0.90,
				B: lab.B * 0.90,
			}, 255),
			accentCol,
			0.30,
		)

		rc.SetBlendMode(render.BlendOver)
		rc.FillPolygon(polyPts, outerRimCol, 0.94)

		// Construct inner hyperbolic scaled polygon by interpolating vertices along geodesics toward tile.center
		innerVerts := make([]complex128, p)
		for i := 0; i < p; i++ {
			arc := sampleGeodesicArc(tile.center, tile.verts[i], 10)
			innerVerts[i] = arc[7] // 70% along geodesic from center to vertex
		}
		innerPts := make([]render.Point, 0, p*8)
		for i := 0; i < p; i++ {
			arcZ := sampleGeodesicArc(innerVerts[i], innerVerts[(i+1)%p], 8)
			for _, z := range arcZ {
				innerPts = append(innerPts, toScreen(z))
			}
		}
		rc.FillPolygon(innerPts, tileCol, 0.88)

		// Innermost jewel core polygon
		corePts := make([]render.Point, p)
		cScreen := toScreen(tile.center)
		for i := 0; i < p; i++ {
			vp := toScreen(innerVerts[i])
			corePts[i] = cScreen.Lerp(vp, 0.56)
		}
		rc.FillPolygon(corePts, innerJewelCol, 0.52)

		// Nested inner hyperbolic star/rosette geodesic chords for tiles with sufficient screen size
		tileRadiusPx := cScreen.Dist(toScreen(tile.verts[0]))
		if tileRadiusPx > 14.0*scale {
			for i := 0; i < p; i++ {
				// Star chord connecting vertex i to vertex (i+2)%p along exact hyperbolic geodesic
				zA := innerVerts[i]
				zB := innerVerts[(i+2)%p]
				chordArc := sampleGeodesicArc(zA, zB, 10)
				chordPts := make([]render.Point, 0, len(chordArc)+1)
				for _, z := range chordArc {
					chordPts = append(chordPts, toScreen(z))
				}
				chordPts = append(chordPts, toScreen(zB))
				chordPath := &canvas.Path{}
				addCleanPolyline(chordPath, chordPts, false)
				rc.StrokePath(chordPath, 0.72*scale, canvas.RoundCap, canvas.RoundJoin, goldStroke, 0.42)

				// Radial geodesic spoke from tile center to edge hyperbolic midpoint
				midZ := hyperbolicMidpoint(tile.verts[i], tile.verts[(i+1)%p])
				spokeArc := sampleGeodesicArc(tile.center, midZ, 6)
				spokePts := make([]render.Point, 0, len(spokeArc)+1)
				for _, z := range spokeArc {
					spokePts = append(spokePts, toScreen(z))
				}
				spokePts = append(spokePts, toScreen(midZ))
				spokePath := &canvas.Path{}
				addCleanPolyline(spokePath, spokePts, false)
				rc.StrokePath(spokePath, 0.72*scale, canvas.RoundCap, canvas.RoundJoin, goldStroke, 0.42)
			}

			// Luminous jewel glow at tile hyperbolic center
			rc.SetBlendMode(render.BlendAdd)
			glowRad := common.Constrain(tileRadiusPx*0.48, 2.5*scale, 28.0*scale)
			rc.DrawGlowDot(cScreen.X, cScreen.Y, glowRad, innerJewelCol, 0.22*(1.0-0.5*hypNorm), 0.25)
			if tile.layer <= 2 {
				rc.DrawGlowDot(cScreen.X, cScreen.Y, math.Max(2.2*scale, tileRadiusPx*0.10), ivoryStroke, 0.65, 0.55)
			}
			rc.SetBlendMode(render.BlendOver)
		}

		// Stroke primary geodesic tile boundary per tile in painter's order
		tileBoundaryPath := &canvas.Path{}
		addCleanPolyline(tileBoundaryPath, polyPts, true)
		rc.StrokePath(tileBoundaryPath, 2.25*scale, canvas.RoundCap, canvas.RoundJoin, darkCasing, 0.85)
		rc.StrokePath(tileBoundaryPath, 1.18*scale, canvas.RoundCap, canvas.RoundJoin, goldStroke, 0.88)
		if tileRadiusPx > 10.0*scale {
			rc.StrokePath(tileBoundaryPath, 0.50*scale, canvas.RoundCap, canvas.RoundJoin, ivoryStroke, 0.72)
		}
	}

	// 5. Frame the Poincare unit disk with an ornate celestial horizon ring and degree ticks
	bezelR1 := diskR * 1.022
	bezelR2 := diskR * 1.064
	bezelR3 := diskR * 1.082

	// Subtle dark outer mask ring between diskR and bezelR1 for crisp horizon separation
	horizonPath := &canvas.Path{}
	const ringSegs = 180
	for i := 0; i <= ringSegs; i++ {
		ang := (float64(i) / float64(ringSegs)) * 2.0 * math.Pi
		px := cx + diskR*math.Cos(ang)
		py := cy + diskR*math.Sin(ang)
		if i == 0 {
			horizonPath.MoveTo(px, py)
		} else {
			horizonPath.LineTo(px, py)
		}
	}
	horizonPath.Close()
	rc.StrokePath(horizonPath, 3.6*scale, canvas.RoundCap, canvas.RoundJoin, darkCasing, 0.95)
	rc.StrokePath(horizonPath, 1.8*scale, canvas.RoundCap, canvas.RoundJoin, goldStroke, 0.95)
	rc.StrokePath(horizonPath, 0.7*scale, canvas.RoundCap, canvas.RoundJoin, ivoryStroke, 0.85)

	rc.SetBlendMode(render.BlendAdd)
	rc.DrawSoftRing(cx, cy, diskR, 2.4*scale, goldStroke, 0.62)
	rc.DrawSoftRing(cx, cy, bezelR1, 1.2*scale, goldStroke, 0.48)
	rc.DrawSoftRing(cx, cy, bezelR2, 1.6*scale, grad.SampleOklab(0.75), 0.42)
	rc.DrawSoftRing(cx, cy, bezelR3, 0.9*scale, goldStroke, 0.32)

	tickPath := &canvas.Path{}
	const nTicks = 144
	for i := 0; i < nTicks; i++ {
		ang := (float64(i) / float64(nTicks)) * 2.0 * math.Pi
		rIn := bezelR1
		rOut := bezelR1 + (bezelR2-bezelR1)*0.52
		if i%12 == 0 {
			rIn = diskR
			rOut = bezelR3
		} else if i%4 == 0 {
			rIn = diskR + (bezelR1-diskR)*0.3
			rOut = bezelR2
		} else if i%2 == 0 {
			rOut = bezelR1 + (bezelR2-bezelR1)*0.78
		}
		tickPath.MoveTo(cx+rIn*math.Cos(ang), cy+rIn*math.Sin(ang))
		tickPath.LineTo(cx+rOut*math.Cos(ang), cy+rOut*math.Sin(ang))
	}
	rc.SetBlendMode(render.BlendOver)
	rc.StrokePath(tickPath, 0.9*scale, canvas.RoundCap, canvas.RoundJoin, goldStroke, 0.55)

	// Cardinal & p-fold symmetry compass jewels along the outer bezel
	rc.SetBlendMode(render.BlendAdd)
	for k := 0; k < p*2; k++ {
		ang := (float64(k) / float64(p*2)) * 2.0 * math.Pi
		jx := cx + bezelR2*math.Cos(ang)
		jy := cy + bezelR2*math.Sin(ang)
		rc.DrawGlowDot(jx, jy, 3.2*scale, ivoryStroke, 0.68, 0.5)
	}

	rc.ApplyBloom(0.5, 8.0, 0.45)
	rc.ApplyVignette(0.3)
	c.MergeRenderCanvasToneMapped(rc, 1.1, 0)
}
