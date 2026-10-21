package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type mondrianSubdivision struct {
	depth int
	pad   float64
}

// NewMondrianSubdivision returns a mondrianSubdivision engine that renders a
// Bauhaus / Constructivist 2.5D relief composition using golden-ratio KD-tree
// subdivision, sub-pixel clip mask stacks, and 2D SDF drop shadows & bevels.
func NewMondrianSubdivision(depth int, pad float64) *mondrianSubdivision {
	if depth <= 0 {
		depth = 4
	}
	if pad <= 0 {
		pad = 6.0
	}
	return &mondrianSubdivision{
		depth: depth,
		pad:   pad,
	}
}

type bauhausCell struct {
	x0, y0, x1, y1 float64
}

// Generative renders the Bauhaus / Neoplastic 2.5D SDF relief poster.
func (ms *mondrianSubdivision) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0

	// Curated Bauhaus Weimar / Constructivist Palette
	obsidianFrame := color.RGBA{R: 18, G: 18, B: 24, A: 255}
	parchmentCream := color.RGBA{R: 244, G: 236, B: 218, A: 255}
	warmStone := color.RGBA{R: 222, G: 210, B: 188, A: 255}

	kleinCobalt := color.RGBA{R: 28, G: 78, B: 168, A: 255}
	deepPrussian := color.RGBA{R: 16, G: 38, B: 88, A: 255}
	skyCerulean := color.RGBA{R: 112, G: 184, B: 228, A: 255}

	vermilionRed := color.RGBA{R: 214, G: 56, B: 42, A: 255}
	crimsonShadow := color.RGBA{R: 118, G: 24, B: 22, A: 255}
	coralHighlight := color.RGBA{R: 252, G: 148, B: 122, A: 255}

	bauhausYellow := color.RGBA{R: 244, G: 186, B: 46, A: 255}
	amberBronze := color.RGBA{R: 148, G: 96, B: 18, A: 255}
	ivorySpec := color.RGBA{R: 255, G: 248, B: 224, A: 255}

	bauhausTeal := color.RGBA{R: 42, G: 146, B: 138, A: 255}

	rc.FillBackground(obsidianFrame)
	rc.SetBlendMode(render.BlendOver)

	rng := rand.New(rand.NewSource(202615))

	// 1. Golden-Ratio Recursive KD-Subdivision into Architectural Chambers
	var cells []bauhausCell
	var subdivide func(b bauhausCell, d int)
	subdivide = func(b bauhausCell, d int) {
		bw := b.x1 - b.x0
		bh := b.y1 - b.y0
		if d <= 0 || (bw < 115*scale && bh < 115*scale) {
			cells = append(cells, b)
			return
		}
		// Stop early on occasional large focal chambers
		if d <= ms.depth-2 && rng.Float64() < 0.24 {
			cells = append(cells, b)
			return
		}

		// Golden-ratio split ratio (0.382 or 0.618)
		ratios := []float64{0.382, 0.50, 0.618}
		r := ratios[rng.Intn(len(ratios))]

		splitHoriz := bw > bh*1.15 || (bh <= bw*1.15 && rng.Float64() < 0.5)
		if splitHoriz && bw >= 110*scale {
			xm := b.x0 + bw*r
			subdivide(bauhausCell{b.x0, b.y0, xm, b.y1}, d-1)
			subdivide(bauhausCell{xm, b.y0, b.x1, b.y1}, d-1)
		} else if bh >= 110*scale {
			ym := b.y0 + bh*r
			subdivide(bauhausCell{b.x0, b.y0, b.x1, ym}, d-1)
			subdivide(bauhausCell{b.x0, ym, b.x1, b.y1}, d-1)
		} else {
			cells = append(cells, b)
		}
	}

	outerMargin := 10.0 * scale
	subdivide(bauhausCell{outerMargin, outerMargin, w - outerMargin, h - outerMargin}, ms.depth)

	pad := ms.pad * 0.5 * scale
	bgChoices := []color.RGBA{parchmentCream, warmStone, parchmentCream, deepPrussian, kleinCobalt, vermilionRed, bauhausYellow}

	// 2. Render Each Chamber Using Sub-Pixel Clip Mask Stack + 2D SDF Relief Shaders
	for idx, cell := range cells {
		x0 := cell.x0 + pad
		y0 := cell.y0 + pad
		x1 := cell.x1 - pad
		y1 := cell.y1 - pad
		cw := x1 - x0
		ch := y1 - y0
		if cw < 12*scale || ch < 12*scale {
			continue
		}

		rectPts := []render.Point{
			render.Pt(x0, y0),
			render.Pt(x1, y0),
			render.Pt(x1, y1),
			render.Pt(x0, y1),
		}

		cellBG := bgChoices[(idx*3+rng.Intn(2))%len(bgChoices)]
		rc.FillPolygon(rectPts, cellBG, 1.0)

		// Push chamber clip mask so all internal Bauhaus geometry & SDF shadows stay strictly inside [x0, y0, x1, y1]
		rc.PushClipPolygon(rectPts)

		cx := 0.5 * (x0 + x1)
		cy := 0.5 * (y0 + y1)
		minDim := math.Min(cw, ch)

		motif := (idx + rng.Intn(2)) % 4
		switch motif {
		case 0:
			// Motif 0: 2D SDF Beveled Bauhaus Target Sphere + Diagonal Engraving Lines
			lineCol := obsidianFrame
			if cellBG == deepPrussian || cellBG == kleinCobalt {
				lineCol = parchmentCream
			}
			rc.DrawHatchPatch(
				[]render.Point{render.Pt(x0, y0), render.Pt(x1, y0)},
				[]render.Point{render.Pt(x0, y1), render.Pt(x1, y1)},
				int(math.Max(8, ch/(7.5*scale))),
				1.15*scale,
				0.0,
				func(s, t float64) color.RGBA { return lineCol },
				0.28,
			)

			rOrb := minDim * 0.36
			orbSDF := render.SDFCircle(render.Pt(cx, cy), rOrb)
			rc.RenderSDFDropShadow(orbSDF, render.Pt(5*scale, 6*scale), 10*scale, 1*scale, obsidianFrame, 0.65)

			baseCol, rimCol, specCol := vermilionRed, crimsonShadow, coralHighlight
			if cellBG == vermilionRed {
				baseCol, rimCol, specCol = bauhausYellow, amberBronze, ivorySpec
			} else if idx%2 == 1 {
				baseCol, rimCol, specCol = kleinCobalt, deepPrussian, skyCerulean
			}
			rc.RenderSDFBevel(orbSDF, baseCol, rimCol, specCol, rOrb*0.38, render.Pt(-0.7, -0.7), 1.0)

			// Inner concentric Bauhaus ring
			rc.DrawSoftRing(cx, cy, rOrb*0.46, 2.2*scale, ivorySpec, 0.92)

		case 1:
			// Motif 1: Quarter-Circle Corner Arch & Stepped Bauhaus Fan
			cornerX, cornerY := x0, y1
			startAng, endAng := -0.5*math.Pi, 0.0
			if idx%2 == 1 {
				cornerX, cornerY = x1, y0
				startAng, endAng = 0.5*math.Pi, math.Pi
			}
			maxR := math.Min(cw, ch) * 0.92
			archCols := []color.RGBA{bauhausYellow, vermilionRed, bauhausTeal, kleinCobalt, obsidianFrame}
			for ring := 4; ring >= 0; ring-- {
				rOut := maxR * float64(ring+1) / 5.0
				rIn := maxR * float64(ring) / 5.0
				rc.DrawAnnularSector(cornerX, cornerY, rIn, rOut, startAng, endAng, archCols[(idx+ring)%len(archCols)], 1.0)
			}

		case 2:
			// Motif 2: Constructivist Rotated Geometric Pill & Crossbeams (using 2D Transform Stack)
			rc.PushMatrix()
			rc.Translate(cx, cy)
			rc.Rotate(-0.52)

			beamW := minDim * 0.68
			beamH := minDim * 0.22
			bar1 := []render.Point{
				render.Pt(-beamW, -beamH),
				render.Pt(beamW, -beamH),
				render.Pt(beamW, beamH),
				render.Pt(-beamW, beamH),
			}
			bar2 := []render.Point{
				render.Pt(-beamH*0.8, -beamW*0.95),
				render.Pt(beamH*0.8, -beamW*0.95),
				render.Pt(beamH*0.8, beamW*0.95),
				render.Pt(-beamH*0.8, beamW*0.95),
			}
			rc.FillPolygon(bar1, obsidianFrame, 0.92)
			rc.FillPolygon(bar2, bauhausTeal, 0.95)
			rc.PopMatrix()

			// Center beveled diamond jewel
			dR := minDim * 0.26
			diamondPts := []render.Point{
				render.Pt(cx, cy-dR),
				render.Pt(cx+dR, cy),
				render.Pt(cx, cy+dR),
				render.Pt(cx-dR, cy),
			}
			dSDF := render.SDFPolygon(diamondPts)
			rc.RenderSDFDropShadow(dSDF, render.Pt(4*scale, 5*scale), 8*scale, 1*scale, obsidianFrame, 0.65)
			rc.RenderSDFBevel(dSDF, bauhausYellow, amberBronze, ivorySpec, dR*0.35, render.Pt(-0.7, -0.7), 1.0)

		default:
			// Motif 3: Split Dual-Tone Bauhaus Arch & Floating Half-Discs
			rArch := minDim * 0.38
			boxSDF := render.SDFRoundedBox(render.Pt(cx, cy+rArch*0.25), render.Pt(rArch, rArch*0.85), rArch*0.95)
			rc.RenderSDFDropShadow(boxSDF, render.Pt(5*scale, 6*scale), 10*scale, 1*scale, obsidianFrame, 0.6)
			rc.RenderSDFBevel(boxSDF, bauhausTeal, deepPrussian, ivorySpec, rArch*0.32, render.Pt(-0.7, -0.7), 1.0)

			// Upper floating sun disc
			rc.DrawShadedOrb(
				render.Pt(cx, y0+ch*0.26),
				rArch*0.42,
				ivorySpec,
				vermilionRed,
				obsidianFrame,
				render.Pt(-0.4, -0.4),
				1.0,
			)
		}

		rc.PopClip()

		// Crisp inner architectural bevel frame
		framePath := &canvas.Path{}
		framePath.MoveTo(x0, y0)
		framePath.LineTo(x1, y0)
		framePath.LineTo(x1, y1)
		framePath.LineTo(x0, y1)
		framePath.Close()
		rc.StrokePath(framePath, 2.2*scale, canvas.RoundCap, canvas.RoundJoin, obsidianFrame, 1.0)
	}

	rc.ApplyPaperTexture(0.06, 202615)

	c.MergeRenderCanvas(rc)
}
