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

type tangentCones struct {
	pairCount int
}

// NewTangentCones returns a tangentCones engine that renders a Bauhaus / Kandinsky
// geometric network of external tangent cones, translucent shaded orbs, and constellation axis rods.
func NewTangentCones(pairCount int) *tangentCones {
	if pairCount <= 0 {
		pairCount = 26
	}
	return &tangentCones{
		pairCount: pairCount,
	}
}

type orbPairSpec struct {
	c1, c2 render.Point
	r1, r2 float64
	style1 int
	style2 int
}

// Generative renders the Bauhaus / Kandinsky cosmic tangent-cone and translucent orb composition.
func (tc *tangentCones) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	// Bauhaus / Kandinsky palette
	coralVermilion := color.RGBA{R: 234, G: 88, B: 56, A: 255}
	amberOrange := color.RGBA{R: 244, G: 152, B: 42, A: 255}
	twilightIndigo := color.RGBA{R: 32, G: 24, B: 38, A: 255}
	lilacViolet := color.RGBA{R: 196, G: 148, B: 208, A: 255}
	periwinkle := color.RGBA{R: 128, G: 138, B: 224, A: 255}
	warmRose := color.RGBA{R: 228, G: 118, B: 136, A: 255}
	iceBlue := color.RGBA{R: 210, G: 228, B: 252, A: 255}
	goldenYellow := color.RGBA{R: 250, G: 206, B: 74, A: 255}
	obsidianBlack := color.RGBA{R: 18, G: 16, B: 24, A: 255}
	charcoalLine := color.RGBA{R: 28, G: 24, B: 32, A: 255}
	creamWhite := color.RGBA{R: 248, G: 242, B: 228, A: 255}

	// 1. Backdrop: rich warm coral-vermilion transitioning smoothly to amber-orange in the upper-right
	// and deep twilight-indigo in the upper-left corner.
	for py := 0; py < c.Height(); py++ {
		ny := (float64(py) + 0.5) / math.Max(1.0, h)
		for px := 0; px < c.Width(); px++ {
			nx := (float64(px) + 0.5) / math.Max(1.0, w)

			// Upper-right warm amber-orange glow
			urDist := math.Hypot(nx-0.88, ny-0.16)
			urWeight := common.Constrain(1.0-urDist*1.35, 0, 1)
			urSmooth := urWeight * urWeight * (3.0 - 2.0*urWeight)
			col := render.LerpOklab(coralVermilion, amberOrange, urSmooth*0.88)

			// Upper-left deep twilight-indigo atmospheric corner
			ulDist := math.Hypot(nx*1.15, ny*1.35)
			ulWeight := common.Constrain(1.0-ulDist*1.42, 0, 1)
			ulSmooth := ulWeight * ulWeight * (3.0 - 2.0*ulWeight)
			col = render.LerpOklab(col, twilightIndigo, ulSmooth*0.94)

			rc.BlendPixel(px, py, col, 1.0)
		}
	}

	rc.SetBlendMode(render.BlendOver)

	// Overlay 4 giant soft translucent background circles in lilac-violet, periwinkle, and warm rose
	bgCircles := []struct {
		cx, cy, r       float64
		coreCol, rimCol color.RGBA
		strokeCol       color.RGBA
		alpha           float64
	}{
		{0.46 * w, 0.44 * h, 195.0 * scale, lilacViolet, warmRose, creamWhite, 0.34},
		{0.64 * w, 0.34 * h, 155.0 * scale, periwinkle, lilacViolet, lilacViolet, 0.30},
		{0.34 * w, 0.66 * h, 175.0 * scale, warmRose, coralVermilion, amberOrange, 0.28},
		{0.72 * w, 0.68 * h, 138.0 * scale, lilacViolet, periwinkle, creamWhite, 0.26},
	}
	for _, bgc := range bgCircles {
		rc.DrawShadedOrb(
			render.Pt(bgc.cx, bgc.cy),
			bgc.r,
			bgc.coreCol,
			bgc.rimCol,
			bgc.strokeCol,
			render.Pt(-0.35, -0.35),
			bgc.alpha,
		)
	}

	rng := rand.New(rand.NewSource(7071926))

	// 2. Upper-left meteor / rain shower: ~45 crisp vertical fine lines with glowing diamond/circle bead nodes
	const meteorCount = 46
	for i := 0; i < meteorCount; i++ {
		mx := (0.02 + 0.44*rng.Float64()) * w
		// Concentrate longer lines toward the upper-left corner
		maxYFrac := 0.14 + 0.34*(1.0-mx/(0.50*w)) + 0.12*rng.Float64()
		y0 := (0.01 + 0.08*rng.Float64()) * h
		y1 := common.Constrain(maxYFrac, 0.10, 0.52) * h
		if y1 <= y0+12.0*scale {
			y1 = y0 + (18.0+35.0*rng.Float64())*scale
		}

		lineCol := render.LerpOklab(creamWhite, periwinkle, rng.Float64()*0.55)
		alpha := 0.35 + 0.42*rng.Float64()
		p := &canvas.Path{}
		p.MoveTo(mx, y0)
		p.LineTo(mx, y1)
		rc.StrokePath(p, (0.65+0.35*rng.Float64())*scale, canvas.RoundCap, canvas.RoundJoin, lineCol, alpha)

		// Tiny glowing diamond / circle bead nodes along the vertical line
		beads := 1 + rng.Intn(3)
		for b := 0; b < beads; b++ {
			by := y0 + (0.2+0.75*rng.Float64())*(y1-y0)
			rad := (1.4 + 1.8*rng.Float64()) * scale
			if (i+b)%2 == 0 {
				// Crisp tiny diamond node
				diamond := []render.Point{
					render.Pt(mx, by-rad*1.35),
					render.Pt(mx+rad, by),
					render.Pt(mx, by+rad*1.35),
					render.Pt(mx-rad, by),
				}
				rc.FillPolygon(diamond, creamWhite, alpha*1.1)
			} else {
				rc.DrawGlowDot(mx, by, rad*1.2, goldenYellow, alpha*1.1, 0.45)
			}
		}
	}

	// 3. Generate pairCount circle pairs ((c1, r1), (c2, r2)) with external tangent cones and shaded orbs
	pairs := make([]orbPairSpec, 0, tc.pairCount)
	for i := 0; i < tc.pairCount; i++ {
		// Distribute pairs across the composition with golden-angle quasi-random coverage
		u := (float64(i) + 0.5) / float64(tc.pairCount)
		phi := float64(i) * 2.399963229728653
		radSpread := (0.10 + 0.36*math.Sqrt(u)) * math.Min(w, h)
		cx := 0.50*w + radSpread*math.Cos(phi)*1.08 + (rng.Float64()-0.5)*36.0*scale
		cy := 0.50*h + radSpread*math.Sin(phi)*0.96 + (rng.Float64()-0.5)*36.0*scale
		cx = common.Constrain(cx, 0.11*w, 0.89*w)
		cy = common.Constrain(cy, 0.12*h, 0.88*h)

		r1 := (18.0 + 30.0*math.Pow(rng.Float64(), 1.15)) * scale
		r2 := (6.0 + 16.0*rng.Float64()) * scale
		if r2 >= r1*0.85 {
			r2 = r1 * 0.65
		}

		dist := (45.0 + 120.0*rng.Float64()) * scale
		if dist < r1+r2+12.0*scale {
			dist = r1 + r2 + 14.0*scale
		}

		// Favor dynamic Bauhaus diagonal angles
		ang := -0.65 + (rng.Float64()-0.5)*2.4
		c1 := render.Pt(cx, cy)
		c2 := render.Pt(cx+dist*math.Cos(ang), cy+dist*math.Sin(ang))

		// Keep c2 comfortably inside canvas margins
		if c2.X < 0.06*w || c2.X > 0.94*w || c2.Y < 0.06*h || c2.Y > 0.94*h {
			c2 = render.Pt(cx-dist*math.Cos(ang), cy-dist*math.Sin(ang))
		}
		c2.X = common.Constrain(c2.X, 0.06*w, 0.94*w)
		c2.Y = common.Constrain(c2.Y, 0.06*h, 0.94*h)

		pairs = append(pairs, orbPairSpec{
			c1:     c1,
			c2:     c2,
			r1:     r1,
			r2:     r2,
			style1: i % 4,
			style2: (i + 1 + rng.Intn(3)) % 4,
		})
	}

	drawStyledOrb := func(center render.Point, radius float64, style int) {
		switch style {
		case 0:
			// Deep obsidian-black sphere with warm amber/vermilion rim glow
			rc.DrawGlowDot(center.X, center.Y, radius*1.28, amberOrange, 0.34, 0.15)
			rc.DrawShadedOrb(
				center,
				radius,
				obsidianBlack,
				color.RGBA{R: 92, G: 32, B: 36, A: 255},
				amberOrange,
				render.Pt(0.32, 0.32),
				0.96,
			)
		case 1:
			// Luminous periwinkle / ice-blue orb with dark indigo core
			rc.DrawShadedOrb(
				center,
				radius,
				twilightIndigo,
				iceBlue,
				creamWhite,
				render.Pt(-0.28, -0.28),
				0.88,
			)
		case 2:
			// Glowing golden-yellow / amber solar orb with crisp dark rim
			rc.DrawShadedOrb(
				center,
				radius,
				creamWhite,
				goldenYellow,
				charcoalLine,
				render.Pt(-0.35, -0.35),
				0.90,
			)
		default:
			// Translucent violet-rose Bauhaus sphere with obsidian center
			rc.DrawShadedOrb(
				center,
				radius,
				obsidianBlack,
				lilacViolet,
				creamWhite,
				render.Pt(0.25, -0.25),
				0.86,
			)
		}
	}

	conePalettes := []struct {
		col1, col2, edgeCol color.RGBA
	}{
		{periwinkle, lilacViolet, creamWhite},
		{obsidianBlack, coralVermilion, charcoalLine},
		{goldenYellow, warmRose, creamWhite},
		{twilightIndigo, periwinkle, charcoalLine},
	}

	// Render each tangent cone sleeve followed by its two endpoint spheres
	for idx, pr := range pairs {
		cp := conePalettes[idx%len(conePalettes)]
		rc.DrawTangentCone(pr.c1, pr.r1, pr.c2, pr.r2, cp.col1, cp.col2, cp.edgeCol, 0.78)
		drawStyledOrb(pr.c1, pr.r1, pr.style1)
		drawStyledOrb(pr.c2, pr.r2, pr.style2)
	}

	// 4. Constellation Axis Rods: thin crisp geometric lines (solid & dashed) connecting paired and neighboring orb centers
	drawDashedSegment := func(p0, p1 render.Point, width, dashLen, gapLen float64, col color.RGBA, alpha float64) {
		dVec := p1.Sub(p0)
		total := dVec.Len()
		if total < 1e-4 {
			return
		}
		dir := dVec.Mul(1.0 / total)
		path := &canvas.Path{}
		for s := 0.0; s < total; s += dashLen + gapLen {
			e := math.Min(total, s+dashLen)
			a := p0.Add(dir.Mul(s))
			b := p0.Add(dir.Mul(e))
			path.MoveTo(a.X, a.Y)
			path.LineTo(b.X, b.Y)
		}
		rc.StrokePath(path, width, canvas.RoundCap, canvas.RoundJoin, col, alpha)
	}

	axisWidth := 0.9 * scale
	for idx, pr := range pairs {
		// Extend the axis rod slightly past both circle centers like a Kandinsky drafting compass axis
		dir := pr.c2.Sub(pr.c1).Norm()
		ext1 := pr.c1.Sub(dir.Mul(pr.r1 * 0.55))
		ext2 := pr.c2.Add(dir.Mul(pr.r2 * 0.75))

		rodCol := creamWhite
		if idx%2 == 1 {
			rodCol = charcoalLine
		}

		if idx%3 == 0 {
			drawDashedSegment(ext1, ext2, axisWidth, 5.5*scale, 4.0*scale, rodCol, 0.82)
		} else {
			p := &canvas.Path{}
			p.MoveTo(ext1.X, ext1.Y)
			p.LineTo(ext2.X, ext2.Y)
			rc.StrokePath(p, axisWidth, canvas.RoundCap, canvas.RoundJoin, rodCol, 0.78)
		}

		// Connect to nearest neighboring pair center to form a constellation network
		bestJ := -1
		bestDist := 185.0 * scale
		for j := idx + 1; j < len(pairs); j++ {
			d := pr.c1.Dist(pairs[j].c1)
			if d > 35.0*scale && d < bestDist {
				bestDist = d
				bestJ = j
			}
		}
		if bestJ >= 0 {
			neighbor := pairs[bestJ].c1
			if (idx+bestJ)%2 == 0 {
				drawDashedSegment(pr.c1, neighbor, axisWidth*0.88, 4.5*scale, 4.5*scale, creamWhite, 0.58)
			} else {
				p := &canvas.Path{}
				p.MoveTo(pr.c1.X, pr.c1.Y)
				p.LineTo(neighbor.X, neighbor.Y)
				rc.StrokePath(p, axisWidth*0.88, canvas.RoundCap, canvas.RoundJoin, charcoalLine, 0.55)
			}
		}

		// Crisp pinhead dots at exact circle centers
		pinCol1 := creamWhite
		if pr.style1 == 2 {
			pinCol1 = obsidianBlack
		}
		pinCol2 := creamWhite
		if pr.style2 == 2 {
			pinCol2 = obsidianBlack
		}
		rc.DrawGlowDot(pr.c1.X, pr.c1.Y, 2.1*scale, pinCol1, 0.95, 0.85)
		rc.DrawGlowDot(pr.c2.X, pr.c2.Y, 1.8*scale, pinCol2, 0.95, 0.85)
	}

	rc.ApplyPaperTexture(0.18, 707)
	c.MergeRenderCanvas(rc)
}
