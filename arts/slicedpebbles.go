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

type slicedPebbles struct {
	gridSize     int
	maxCuts      int
	cornerRadius float64
}

// NewSlicedPebbles creates a mid-century modern geometric grid engine where
// square tiles are sliced by random straight chords with uniform mortar gaps
// and every resulting sub-polygon is rounded with exact tangential circular arcs.
func NewSlicedPebbles(gridSize, maxCuts int, cornerRadius float64) *slicedPebbles {
	if gridSize <= 0 {
		gridSize = 6
	}
	if maxCuts <= 0 {
		maxCuts = 3
	}
	if cornerRadius <= 0 {
		cornerRadius = 18.0
	}
	return &slicedPebbles{
		gridSize:     gridSize,
		maxCuts:      maxCuts,
		cornerRadius: cornerRadius,
	}
}

func spPolygonSignedArea(poly []render.Point) float64 {
	n := len(poly)
	if n < 3 {
		return 0
	}
	sum := 0.0
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		sum += poly[i].X*poly[j].Y - poly[j].X*poly[i].Y
	}
	return 0.5 * sum
}

func spPolygonArea(poly []render.Point) float64 {
	return math.Abs(spPolygonSignedArea(poly))
}

// spClipHalfPlane clips a convex polygon against the half-plane (p - linePt) · normal >= offset.
func spClipHalfPlane(poly []render.Point, linePt, normal render.Point, offset float64) []render.Point {
	n := len(poly)
	if n < 3 {
		return nil
	}
	eval := func(p render.Point) float64 {
		return (p.X-linePt.X)*normal.X + (p.Y-linePt.Y)*normal.Y - offset
	}
	out := make([]render.Point, 0, n+2)
	for i := 0; i < n; i++ {
		cur := poly[i]
		nxt := poly[(i+1)%n]
		dCur := eval(cur)
		dNxt := eval(nxt)

		if dCur >= 0 {
			out = append(out, cur)
			if dNxt < 0 {
				t := dCur / (dCur - dNxt)
				out = append(out, cur.Lerp(nxt, t))
			}
		} else if dNxt >= 0 {
			t := dCur / (dCur - dNxt)
			out = append(out, cur.Lerp(nxt, t))
		}
	}
	// Deduplicate nearly-coincident vertices
	cleaned := make([]render.Point, 0, len(out))
	for _, p := range out {
		if len(cleaned) == 0 || p.Dist(cleaned[len(cleaned)-1]) > 0.5 {
			cleaned = append(cleaned, p)
		}
	}
	if len(cleaned) >= 2 && cleaned[0].Dist(cleaned[len(cleaned)-1]) <= 0.5 {
		cleaned = cleaned[:len(cleaned)-1]
	}
	if len(cleaned) < 3 {
		return nil
	}
	return cleaned
}

// spMinimumThickness estimates the minimum width (minimum altitude) of a convex polygon.
func spMinimumThickness(poly []render.Point) float64 {
	n := len(poly)
	if n < 3 {
		return 0
	}
	minSpan := math.MaxFloat64
	for i := 0; i < n; i++ {
		a := poly[i]
		b := poly[(i+1)%n]
		edge := b.Sub(a)
		el := edge.Len()
		if el < 1.5 {
			return 0
		}
		norm := edge.Mul(1.0 / el).Perp()
		lo, hi := math.MaxFloat64, -math.MaxFloat64
		for _, p := range poly {
			proj := p.X*norm.X + p.Y*norm.Y
			if proj < lo {
				lo = proj
			}
			if proj > hi {
				hi = proj
			}
		}
		span := hi - lo
		if span < minSpan {
			minSpan = span
		}
	}
	return minSpan
}

// buildFilletedPath constructs a smoothly rounded closed path for a convex polygon
// using exact tangential circular arcs at every vertex so sharp acute tips become
// organic mid-century rounded pebbles without self-intersection.
func buildFilletedPath(poly []render.Point, cellX0, cellY0, cellX1, cellY1, cellSize, baseRadius float64, bigCorners [4]bool) *canvas.Path {
	n := len(poly)
	if n < 3 {
		return nil
	}
	// Ensure consistent orientation
	pts := make([]render.Point, n)
	copy(pts, poly)

	type cornerArc struct {
		pStart render.Point
		pEnd   render.Point
		vtx    render.Point
		center render.Point
		radius float64
		ang0   float64
		ang1   float64
		isArc  bool
	}

	// Compute desired trim distance dTrim at each vertex first, then clamp so
	// dTrim[i] + dTrim[i+1] <= 0.92 * edgeLen[i].
	L := make([]float64, n)
	uIn := make([]render.Point, n)
	uOut := make([]render.Point, n)
	halfAng := make([]float64, n)
	dTrim := make([]float64, n)

	origCorners := [4]render.Point{
		render.Pt(cellX0, cellY0),
		render.Pt(cellX1, cellY0),
		render.Pt(cellX1, cellY1),
		render.Pt(cellX0, cellY1),
	}

	for i := 0; i < n; i++ {
		prev := pts[(i+n-1)%n]
		cur := pts[i]
		next := pts[(i+1)%n]

		vPrev := prev.Sub(cur)
		vNext := next.Sub(cur)
		lPrev := vPrev.Len()
		lNext := vNext.Len()
		L[i] = lNext
		if lPrev < 1e-4 || lNext < 1e-4 {
			continue
		}
		uIn[i] = vPrev.Mul(1.0 / lPrev)
		uOut[i] = vNext.Mul(1.0 / lNext)

		dot := common.Constrain(uIn[i].Dot(uOut[i]), -0.9995, 0.9995)
		ha := 0.5 * math.Acos(dot)
		halfAng[i] = ha

		// Check if cur is one of the 4 outer corners of the original grid square
		rWant := baseRadius
		for cIdx, oc := range origCorners {
			if cur.Dist(oc) < 2.5 {
				if bigCorners[cIdx] {
					rWant = 0.34 * cellSize
				} else {
					rWant = 0.11 * cellSize
				}
				break
			}
		}

		tanHA := math.Max(0.12, math.Tan(ha))
		dt := rWant / tanHA
		maxAllowed := 0.45 * math.Min(lPrev, lNext)
		if dt > maxAllowed {
			dt = maxAllowed
		}
		dTrim[i] = dt
	}

	// Second pass: ensure adjacent trims along edge i -> (i+1) never exceed 92% of that edge
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		edgeLen := L[i]
		if dTrim[i]+dTrim[j] > 0.92*edgeLen && (dTrim[i]+dTrim[j]) > 1e-6 {
			factor := (0.92 * edgeLen) / (dTrim[i] + dTrim[j])
			dTrim[i] *= factor
			dTrim[j] *= factor
		}
	}

	p := &canvas.Path{}
	const arcSteps = 18
	for i := 0; i < n; i++ {
		cur := pts[i]
		dt := dTrim[i]
		ha := halfAng[i]
		if dt < 0.35 || ha < 0.04 || ha > math.Pi*0.495 {
			if i == 0 {
				p.MoveTo(cur.X, cur.Y)
			} else {
				p.LineTo(cur.X, cur.Y)
			}
			continue
		}

		pStart := cur.Add(uIn[i].Mul(dt))
		pEnd := cur.Add(uOut[i].Mul(dt))
		rActual := dt * math.Tan(ha)
		bisector := uIn[i].Add(uOut[i])
		bLen := bisector.Len()
		if bLen < 1e-5 || rActual < 0.25 {
			if i == 0 {
				p.MoveTo(pStart.X, pStart.Y)
			} else {
				p.LineTo(pStart.X, pStart.Y)
			}
			p.LineTo(pEnd.X, pEnd.Y)
			continue
		}
		bisector = bisector.Mul(1.0 / bLen)
		centerDist := rActual / math.Sin(ha)
		arcCenter := cur.Add(bisector.Mul(centerDist))

		ang0 := math.Atan2(pStart.Y-arcCenter.Y, pStart.X-arcCenter.X)
		ang1 := math.Atan2(pEnd.Y-arcCenter.Y, pEnd.X-arcCenter.X)
		// Normalize shortest angular sweep
		dAng := ang1 - ang0
		for dAng > math.Pi {
			dAng -= 2.0 * math.Pi
		}
		for dAng < -math.Pi {
			dAng += 2.0 * math.Pi
		}

		for s := 0; s <= arcSteps; s++ {
			u := float64(s) / float64(arcSteps)
			ang := ang0 + dAng*u
			ax := arcCenter.X + rActual*math.Cos(ang)
			ay := arcCenter.Y + rActual*math.Sin(ang)
			if i == 0 && s == 0 {
				p.MoveTo(ax, ay)
			} else {
				p.LineTo(ax, ay)
			}
		}
	}
	p.Close()
	return p
}

// Generative renders the SlicedPebbles composition.
func (sp *slicedPebbles) Generative(c *generativeart.Canva) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	seed := c.Seed()
	if seed == 0 {
		seed = 2026102803
	}
	rng := rand.New(rand.NewSource(seed))

	rc := c.NewRenderCanvas()
	obsidianBg := color.RGBA{R: 15, G: 18, B: 24, A: 255} // #0f1218
	rc.FillBackground(obsidianBg)
	rc.SetBlendMode(render.BlendOver)

	// Curated 8-color Mid-Century Modern Palette matching Reference Image 3
	palette := []color.RGBA{
		{R: 212, G: 36, B: 42, A: 255},   // 0: Crimson Red (#d4242a)
		{R: 222, G: 109, B: 46, A: 255},  // 1: Burnt Tangerine (#de6d2e)
		{R: 238, G: 226, B: 202, A: 255}, // 2: Warm Cream (#eee2ca)
		{R: 179, G: 130, B: 123, A: 255}, // 3: Dusty Mauve Rose (#b3827b)
		{R: 122, G: 46, B: 56, A: 255},   // 4: Deep Burgundy (#7a2e38)
		{R: 46, G: 66, B: 87, A: 255},    // 5: Prussian Slate (#2e4257)
		{R: 98, G: 123, B: 150, A: 255},  // 6: Steel Periwinkle (#627b96)
		{R: 138, G: 107, B: 82, A: 255},  // 7: Warm Taupe Umber (#8a6b52)
	}

	n := sp.gridSize
	cellGap := 7.2 * scale
	outerPad := 2.0 * scale
	cellW := (fw - 2.0*outerPad - float64(n-1)*cellGap) / float64(n)
	cellH := (fh - 2.0*outerPad - float64(n-1)*cellGap) / float64(n)
	cellSize := math.Min(cellW, cellH)
	cutHalfGap := 3.1 * scale

	// Assign colors to the n x n grid so no orthogonal neighbor shares the same color
	colorGrid := make([][]int, n)
	for r := 0; r < n; r++ {
		colorGrid[r] = make([]int, n)
		for col := 0; col < n; col++ {
			for attempt := 0; attempt < 30; attempt++ {
				cand := rng.Intn(len(palette))
				if r > 0 && colorGrid[r-1][col] == cand {
					continue
				}
				if col > 0 && colorGrid[r][col-1] == cand {
					continue
				}
				if r > 0 && col > 0 && colorGrid[r-1][col-1] == cand && attempt < 15 {
					continue
				}
				colorGrid[r][col] = cand
				break
			}
		}
	}

	for r := 0; r < n; r++ {
		y0 := outerPad + float64(r)*(cellH+cellGap)
		y1 := y0 + cellH
		cy := 0.5 * (y0 + y1)

		for col := 0; col < n; col++ {
			x0 := outerPad + float64(col)*(cellW+cellGap)
			x1 := x0 + cellW
			cx := 0.5 * (x0 + x1)

			baseSquare := []render.Point{
				render.Pt(x0, y0),
				render.Pt(x1, y0),
				render.Pt(x1, y1),
				render.Pt(x0, y1),
			}

			// Select 1 or 2 outer corners of this cell to have a bold mid-century leaf curve
			var bigCorners [4]bool
			diagChoice := rng.Intn(4)
			bigCorners[diagChoice] = true
			if rng.Float64() < 0.65 {
				bigCorners[(diagChoice+2)%4] = true
			}

			polys := [][]render.Point{baseSquare}

			// Decide slicing pattern for this cell matching Image 3:
			// Mode A (~38% of cells): 1 main diagonal/oblique chord through the cell (2 bold pieces)
			// Mode B (~42% of cells): 2 intersecting/flank chords (3-4 rounded pebbles)
			// Mode C (~20% of cells): 3 chords (4-6 organic pebbles)
			rMode := rng.Float64()
			numCuts := 2
			if rMode < 0.38 {
				numCuts = 1
			} else if rMode < 0.80 {
				numCuts = 2
			} else {
				numCuts = 3
			}
			if sp.maxCuts > 0 && numCuts > sp.maxCuts {
				numCuts = sp.maxCuts
			}

			baseAng := rng.Float64() * math.Pi
			for cut := 0; cut < numCuts; cut++ {
				var ang, offsetFromCenter float64
				if cut == 0 {
					// Primary chord passing near the cell center at an expressive oblique angle
					ang = baseAng
					offsetFromCenter = (rng.Float64() - 0.5) * 0.24 * cellSize
				} else if cut == 1 {
					// Secondary chord either roughly orthogonal or slicing one flank
					ang = baseAng + (0.38+0.35*rng.Float64())*math.Pi
					offsetFromCenter = (rng.Float64() - 0.5) * 0.48 * cellSize
				} else {
					// Third chord slicing an outer lobe into smaller pebbles
					ang = baseAng + (0.72+0.45*rng.Float64())*math.Pi
					offsetFromCenter = (0.18 + 0.26*rng.Float64()) * cellSize
					if rng.Float64() < 0.5 {
						offsetFromCenter = -offsetFromCenter
					}
				}

				normal := render.Pt(math.Cos(ang), math.Sin(ang))
				linePt := render.Pt(cx+normal.X*offsetFromCenter, cy+normal.Y*offsetFromCenter)

				nextPolys := make([][]render.Point, 0, len(polys)*2)
				for _, p := range polys {
					// For the 3rd cut, only slice larger pieces so we don't over-fragment small pebbles
					if cut >= 2 && spPolygonArea(p) < 0.26*cellSize*cellSize {
						nextPolys = append(nextPolys, p)
						continue
					}
					halfA := spClipHalfPlane(p, linePt, normal, cutHalfGap)
					halfB := spClipHalfPlane(p, linePt, normal.Mul(-1), cutHalfGap)

					validA := halfA != nil && spPolygonArea(halfA) >= 48.0*scale*scale && spMinimumThickness(halfA) >= 9.0*scale
					validB := halfB != nil && spPolygonArea(halfB) >= 48.0*scale*scale && spMinimumThickness(halfB) >= 9.0*scale

					if validA && validB {
						nextPolys = append(nextPolys, halfA, halfB)
					} else {
						// Keep unsplit polygon if the cut would only shave off a tiny sliver
						nextPolys = append(nextPolys, p)
					}
				}
				polys = nextPolys
			}

			cellCol := palette[colorGrid[r][col]]
			for _, poly := range polys {
				path := buildFilletedPath(poly, x0, y0, x1, y1, cellSize, sp.cornerRadius*scale, bigCorners)
				if path != nil {
					rc.RasterizePath(path, cellCol, 1.0)
				}
			}
		}
	}

	c.MergeRenderCanvas(rc)
}
