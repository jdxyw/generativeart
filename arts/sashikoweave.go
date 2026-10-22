package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type sashikoWeave struct {
	gridSize int
}

// NewSashikoWeave returns a sashikoWeave engine that renders a Japanese Boro
// indigo textile patchwork quilted with 3D twisted Sashiko running-stitch
// embroidery (Asanoha hemp leaf, Seigaiha ocean waves, Shippo seven treasures,
// and Kikko hexagonal lattices).
func NewSashikoWeave(gridSize int) *sashikoWeave {
	if gridSize <= 0 {
		gridSize = 3
	}
	return &sashikoWeave{
		gridSize: gridSize,
	}
}

// Generative renders the Japanese Boro indigo patchwork and Sashiko geometric embroidery.
func (sw *sashikoWeave) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0

	// Authentic Japanese Ai-zome Indigo, Boro Linen & Madder Rust Palette
	deepIndigo := color.RGBA{R: 16, G: 28, B: 58, A: 255}
	kachiNavy := color.RGBA{R: 22, G: 44, B: 84, A: 255}
	asagiSlate := color.RGBA{R: 42, G: 88, B: 132, A: 255}
	tealIndigo := color.RGBA{R: 34, G: 104, B: 118, A: 255}
	akaneMadder := color.RGBA{R: 158, G: 48, B: 42, A: 255}
	kitsuneOchre := color.RGBA{R: 178, G: 122, B: 48, A: 255}

	stitchShadow := color.RGBA{R: 8, G: 14, B: 32, A: 255}
	ivoryCotton := color.RGBA{R: 248, G: 242, B: 226, A: 255}
	goldThread := color.RGBA{R: 246, G: 196, B: 88, A: 255}
	coralThread := color.RGBA{R: 236, G: 108, B: 84, A: 255}

	rc.FillBackground(deepIndigo)
	rc.SetBlendMode(render.BlendOver)

	// Helper to draw a dashed Sashiko running-stitch polyline with 3D thread shadow & cotton sheen
	drawSashikoLine := func(p0, p1 render.Point, threadCol color.RGBA, stitchLen, gapLen float64) {
		dist := p0.Dist(p1)
		if dist < 2.0*scale {
			return
		}
		dir := p1.Sub(p0).Mul(1.0 / dist)
		period := stitchLen + gapLen
		steps := int(math.Floor((dist - 1.5*scale) / period))
		if steps < 1 {
			steps = 1
		}

		shPath := &canvas.Path{}
		thPath := &canvas.Path{}
		hiPath := &canvas.Path{}

		for k := 0; k < steps; k++ {
			s0 := 1.5*scale + float64(k)*period
			s1 := math.Min(dist-1.5*scale, s0+stitchLen)
			if s1 <= s0+0.8*scale {
				continue
			}
			a := p0.Add(dir.Mul(s0))
			b := p0.Add(dir.Mul(s1))

			shPath.MoveTo(a.X+0.8*scale, a.Y+1.1*scale)
			shPath.LineTo(b.X+0.8*scale, b.Y+1.1*scale)

			thPath.MoveTo(a.X, a.Y)
			thPath.LineTo(b.X, b.Y)

			midA := a.Lerp(b, 0.2)
			midB := a.Lerp(b, 0.8)
			hiPath.MoveTo(midA.X-0.25*scale, midA.Y-0.25*scale)
			hiPath.LineTo(midB.X-0.25*scale, midB.Y-0.25*scale)
		}

		rc.StrokePath(shPath, 2.2*scale, canvas.RoundCap, canvas.RoundJoin, stitchShadow, 0.82)
		rc.StrokePath(thPath, 1.65*scale, canvas.RoundCap, canvas.RoundJoin, threadCol, 0.98)
		rc.StrokePath(hiPath, 0.72*scale, canvas.RoundCap, canvas.RoundJoin, ivoryCotton, 0.75)
	}

	drawSashikoArc := func(cx, cy, r, thStart, thEnd float64, threadCol color.RGBA) {
		arcLen := r * math.Abs(thEnd-thStart)
		stitchPeriod := 8.5 * scale
		numStitches := int(math.Max(3, math.Round(arcLen/stitchPeriod)))
		dTh := (thEnd - thStart) / float64(numStitches)

		shPath := &canvas.Path{}
		thPath := &canvas.Path{}
		for k := 0; k < numStitches; k++ {
			a0 := thStart + (float64(k)+0.16)*dTh
			a1 := thStart + (float64(k)+0.84)*dTh
			aMid := 0.5 * (a0 + a1)

			p0 := render.Pt(cx+r*math.Cos(a0), cy+r*math.Sin(a0))
			pMid := render.Pt(cx+r*math.Cos(aMid), cy+r*math.Sin(aMid))
			p1 := render.Pt(cx+r*math.Cos(a1), cy+r*math.Sin(a1))

			shPath.MoveTo(p0.X+0.8*scale, p0.Y+1.1*scale)
			shPath.QuadTo(pMid.X+0.8*scale, pMid.Y+1.1*scale, p1.X+0.8*scale, p1.Y+1.1*scale)

			thPath.MoveTo(p0.X, p0.Y)
			thPath.QuadTo(pMid.X, pMid.Y, p1.X, p1.Y)
		}
		rc.StrokePath(shPath, 2.15*scale, canvas.RoundCap, canvas.RoundJoin, stitchShadow, 0.82)
		rc.StrokePath(thPath, 1.60*scale, canvas.RoundCap, canvas.RoundJoin, threadCol, 0.98)
	}

	// Build Boro Patchwork Layout (asymmetric overlapping kimono/linen fabric rectangles)
	type boroPatch struct {
		x0, y0, x1, y1 float64
		fabricCol      color.RGBA
		threadCol      color.RGBA
		patternType    int
	}

	n := sw.gridSize
	if n < 2 {
		n = 3
	}
	rng := rand.New(rand.NewSource(202616))
	fabrics := []color.RGBA{deepIndigo, kachiNavy, asagiSlate, tealIndigo, akaneMadder, kachiNavy, kitsuneOchre}
	threads := []color.RGBA{ivoryCotton, ivoryCotton, goldThread, ivoryCotton, coralThread}

	var patches []boroPatch
	cellW := w / float64(n)
	cellH := h / float64(n)

	for ry := 0; ry < n; ry++ {
		for rx := 0; rx < n; rx++ {
			jx0 := (rng.Float64() - 0.5) * 18.0 * scale
			jy0 := (rng.Float64() - 0.5) * 18.0 * scale
			jx1 := (rng.Float64() - 0.5) * 18.0 * scale
			jy1 := (rng.Float64() - 0.5) * 18.0 * scale
			if rx == 0 {
				jx0 = -6.0 * scale
			}
			if ry == 0 {
				jy0 = -6.0 * scale
			}
			if rx == n-1 {
				jx1 = 6.0 * scale
			}
			if ry == n-1 {
				jy1 = 6.0 * scale
			}

			idx := ry*n + rx
			patches = append(patches, boroPatch{
				x0:          float64(rx)*cellW + jx0,
				y0:          float64(ry)*cellH + jy0,
				x1:          float64(rx+1)*cellW + jx1,
				y1:          float64(ry+1)*cellH + jy1,
				fabricCol:   fabrics[idx%len(fabrics)],
				threadCol:   threads[idx%len(threads)],
				patternType: idx % 4,
			})
		}
	}

	// Render each Boro fabric patch with woven warp/weft texture, SDF drop shadow, and clipped Sashiko stitches
	for _, pt := range patches {
		rectPts := []render.Point{
			render.Pt(pt.x0, pt.y0),
			render.Pt(pt.x1, pt.y0),
			render.Pt(pt.x1, pt.y1),
			render.Pt(pt.x0, pt.y1),
		}
		patchSDF := render.SDFPolygon(rectPts)
		rc.RenderSDFDropShadow(patchSDF, render.Pt(4*scale, 5*scale), 9*scale, 1*scale, stitchShadow, 0.72)
		rc.FillPolygon(rectPts, pt.fabricCol, 1.0)

		rc.PushClipPolygon(rectPts)

		// Subtle woven linen warp & weft threads inside the fabric patch
		weaveLight := render.LerpOklab(pt.fabricCol, ivoryCotton, 0.09)
		weaveDark := render.LerpOklab(pt.fabricCol, stitchShadow, 0.16)
		pw := pt.x1 - pt.x0
		ph := pt.y1 - pt.y0

		rc.DrawHatchPatch(
			[]render.Point{render.Pt(pt.x0, pt.y0), render.Pt(pt.x1, pt.y0)},
			[]render.Point{render.Pt(pt.x0, pt.y1), render.Pt(pt.x1, pt.y1)},
			int(math.Max(14, ph/(4.5*scale))), 0.8*scale, 0.0,
			func(s, t float64) color.RGBA { return weaveLight }, 0.42,
		)
		rc.DrawHatchPatch(
			[]render.Point{render.Pt(pt.x0, pt.y0), render.Pt(pt.x0, pt.y1)},
			[]render.Point{render.Pt(pt.x1, pt.y0), render.Pt(pt.x1, pt.y1)},
			int(math.Max(14, pw/(4.5*scale))), 0.8*scale, 0.0,
			func(s, t float64) color.RGBA { return weaveDark }, 0.38,
		)

		// Draw one of the 4 traditional Japanese Sashiko geometric patterns
		switch pt.patternType {
		case 0:
			// Pattern 0: Asanoha (麻叶纹 — Hemp Leaf 6-Fold Star Lattice)
			triS := 42.0 * scale
			triH := triS * math.Sqrt(3.0) * 0.5
			cols := int(math.Ceil(pw/triS)) + 2
			rows := int(math.Ceil(ph/triH)) + 2

			for r := -1; r <= rows; r++ {
				y := pt.y0 + float64(r)*triH
				xOff := 0.0
				if r%2 != 0 {
					xOff = 0.5 * triS
				}
				for cIdx := -1; cIdx <= cols; cIdx++ {
					x := pt.x0 + float64(cIdx)*triS + xOff
					v0 := render.Pt(x, y)
					vRight := render.Pt(x+triS, y)
					vDownR := render.Pt(x+0.5*triS, y+triH)
					vDownL := render.Pt(x-0.5*triS, y+triH)

					// Main triangular lattice edges
					drawSashikoLine(v0, vRight, pt.threadCol, 6.2*scale, 3.2*scale)
					drawSashikoLine(v0, vDownR, pt.threadCol, 6.2*scale, 3.2*scale)
					drawSashikoLine(v0, vDownL, pt.threadCol, 6.2*scale, 3.2*scale)

					// Upward & downward triangle centroid star spokes
					c1 := render.Pt((v0.X+vRight.X+vDownR.X)/3.0, (v0.Y+vRight.Y+vDownR.Y)/3.0)
					for _, corner := range []render.Point{v0, vRight, vDownR} {
						drawSashikoLine(c1, corner, pt.threadCol, 5.0*scale, 2.8*scale)
					}
					c2 := render.Pt((v0.X+vDownR.X+vDownL.X)/3.0, (v0.Y+vDownR.Y+vDownL.Y)/3.0)
					for _, corner := range []render.Point{v0, vDownR, vDownL} {
						drawSashikoLine(c2, corner, pt.threadCol, 5.0*scale, 2.8*scale)
					}
				}
			}

		case 1:
			// Pattern 1: Seigaiha (青海波 — Concentric Scalloped Ocean Waves)
			rWave := 34.0 * scale
			dx := rWave * 1.72
			dy := rWave * 0.52
			rows := int(math.Ceil(ph/dy)) + 4
			cols := int(math.Ceil(pw/dx)) + 3

			for r := -2; r <= rows; r++ {
				cy := pt.y0 + float64(r)*dy
				xShift := 0.0
				if r%2 != 0 {
					xShift = 0.5 * dx
				}
				for cIdx := -1; cIdx <= cols; cIdx++ {
					cx := pt.x0 + float64(cIdx)*dx + xShift
					// Fill dark indigo backing disc so overlapping wave scales occlude upper rows
					disc := make([]render.Point, 24)
					for k := 0; k <= 23; k++ {
						a := -math.Pi + (float64(k)/23.0)*math.Pi
						disc[k] = render.Pt(cx+rWave*math.Cos(a), cy+rWave*math.Sin(a))
					}
					rc.FillPolygon(disc, pt.fabricCol, 1.0)

					for ring := 1; ring <= 4; ring++ {
						rad := rWave * (float64(ring) / 4.0)
						col := pt.threadCol
						if ring == 4 && r%3 == 0 {
							col = goldThread
						}
						drawSashikoArc(cx, cy, rad, -math.Pi*0.92, -math.Pi*0.08, col)
					}
				}
			}

		case 2:
			// Pattern 2: Shippo-Tsunagi (七宝纹 — Interlocking Seven Treasures Circles)
			rShippo := 26.0 * scale
			step := rShippo
			cols := int(math.Ceil(pw/step)) + 2
			rows := int(math.Ceil(ph/step)) + 2
			for r := 0; r <= rows; r++ {
				cy := pt.y0 + float64(r)*step
				for cIdx := 0; cIdx <= cols; cIdx++ {
					if (r+cIdx)%2 != 0 {
						continue
					}
					cx := pt.x0 + float64(cIdx)*step
					drawSashikoArc(cx, cy, rShippo, 0, 2.0*math.Pi, pt.threadCol)
				}
			}

		default:
			// Pattern 3: Hitomezashi / Kikko Persimmon & Diamond Stitch Weave
			step := 24.0 * scale
			cols := int(math.Ceil(pw/step)) + 2
			rows := int(math.Ceil(ph/step)) + 2
			for r := 0; r <= rows; r++ {
				y := pt.y0 + float64(r)*step
				for cIdx := 0; cIdx <= cols; cIdx++ {
					x := pt.x0 + float64(cIdx)*step
					p00 := render.Pt(x, y)
					p10 := render.Pt(x+step, y)
					p01 := render.Pt(x, y+step)
					p11 := render.Pt(x+step, y+step)
					mid := render.Pt(x+0.5*step, y+0.5*step)

					drawSashikoLine(p00, p10, pt.threadCol, 6.0*scale, 3.0*scale)
					drawSashikoLine(p00, p01, pt.threadCol, 6.0*scale, 3.0*scale)
					if (r+cIdx)%2 == 0 {
						drawSashikoLine(p00, p11, goldThread, 5.5*scale, 2.8*scale)
						drawSashikoLine(p10, p01, goldThread, 5.5*scale, 2.8*scale)
					} else {
						rc.DrawGlowDot(mid.X, mid.Y, 2.2*scale, coralThread, 0.95, 0.85)
					}
				}
			}
		}

		rc.PopClip()

		// Double Boro basting border stitches around the perimeter of each fabric patch
		inset := 5.5 * scale
		b0 := render.Pt(pt.x0+inset, pt.y0+inset)
		b1 := render.Pt(pt.x1-inset, pt.y0+inset)
		b2 := render.Pt(pt.x1-inset, pt.y1-inset)
		b3 := render.Pt(pt.x0+inset, pt.y1-inset)
		drawSashikoLine(b0, b1, goldThread, 7.0*scale, 3.5*scale)
		drawSashikoLine(b1, b2, goldThread, 7.0*scale, 3.5*scale)
		drawSashikoLine(b2, b3, goldThread, 7.0*scale, 3.5*scale)
		drawSashikoLine(b3, b0, goldThread, 7.0*scale, 3.5*scale)
	}

	c.MergeRenderCanvas(rc)
}
