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

type lichtenbergTree struct {
	branches  int
	roughness float64
}

// NewLichtenbergTree returns a lichtenbergTree engine that renders a high-voltage
// dielectric breakdown Lichtenberg lightning figure captured inside a beveled
// obsidian-sapphire acrylic block with Strahler charge-flux thickness scaling and HDR plasma glow.
func NewLichtenbergTree(branches int, roughness float64) *lichtenbergTree {
	if branches <= 0 {
		branches = 9
	}
	if roughness <= 0 {
		roughness = 1.0
	}
	return &lichtenbergTree{
		branches:  branches,
		roughness: roughness,
	}
}

type lbSegment struct {
	p0, p1   render.Point
	depth    int
	maxDepth int
	flux     float64
}

// Generative renders the high-voltage dielectric breakdown Lichtenberg tree in acrylic.
func (lt *lichtenbergTree) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0

	// High-Voltage Plasma, Electric Cyan, Violet Corona & Incandescent Gold Palette
	abyssAcrylic := color.RGBA{R: 6, G: 8, B: 18, A: 255}
	sapphireBlock := color.RGBA{R: 12, G: 22, B: 48, A: 255}
	violetCorona := color.RGBA{R: 96, G: 38, B: 196, A: 255}
	electricCyan := color.RGBA{R: 48, G: 212, B: 255, A: 255}
	plasmaGold := color.RGBA{R: 255, G: 196, B: 78, A: 255}
	incandescentWhite := color.RGBA{R: 255, G: 250, B: 238, A: 255}

	rc.FillBackground(abyssAcrylic)
	rc.SetBlendMode(render.BlendOver)

	// 1. Render Beveled Acrylic Specimen Block Frame
	margin := 24.0 * scale
	blockSDF := render.SDFRoundedBox(
		render.Pt(0.5*w, 0.5*h),
		render.Pt(0.5*w-margin, 0.5*h-margin),
		12.0*scale,
	)
	rc.RenderSDFBevel(
		blockSDF,
		sapphireBlock,
		abyssAcrylic,
		electricCyan,
		18.0*scale,
		render.Pt(-0.65, -0.75),
		0.85,
	)

	// 2. Generate Recursive Dielectric Breakdown Tree with Jagged Laplace Sub-Stepping
	rng := rand.New(rand.NewSource(202623))
	maxDepth := lt.branches
	if maxDepth < 6 {
		maxDepth = 9
	}
	if maxDepth > 11 {
		maxDepth = 11
	}

	var segments []lbSegment

	// Subdivide a straight chord into a jagged lightning path via midpoint displacement
	jaggedPath := func(a, b render.Point, jScale float64) []render.Point {
		pts := []render.Point{a, b}
		for pass := 0; pass < 3; pass++ {
			next := make([]render.Point, 0, len(pts)*2-1)
			for i := 0; i < len(pts)-1; i++ {
				p0, p1 := pts[i], pts[i+1]
				mid := p0.Lerp(p1, 0.5)
				perp := p1.Sub(p0).Perp()
				disp := (rng.Float64() - 0.5) * 0.28 * jScale
				next = append(next, p0, mid.Add(perp.Mul(disp)))
			}
			next = append(next, pts[len(pts)-1])
			pts = next
		}
		return pts
	}

	var buildBranch func(p render.Point, angle, length float64, depth int) float64
	buildBranch = func(p render.Point, angle, length float64, depth int) float64 {
		if depth > maxDepth || length < 4.0*scale {
			return 1.0
		}
		end := render.Pt(
			p.X+length*math.Cos(angle),
			p.Y+length*math.Sin(angle),
		)
		// Keep inside the acrylic specimen block
		if end.X < margin*1.3 || end.X > w-margin*1.3 || end.Y < margin*1.3 || end.Y > h-margin*1.3 {
			return 0.5
		}

		idx := len(segments)
		segments = append(segments, lbSegment{
			p0:       p,
			p1:       end,
			depth:    depth,
			maxDepth: maxDepth,
		})

		totalFlux := 1.0
		// Primary leader + 1 or 2 lateral dielectric breakdown forks
		spread := (0.34 + 0.16*rng.Float64()) * lt.roughness
		lenDecayA := 0.76 + 0.08*rng.Float64()
		lenDecayB := 0.64 + 0.12*rng.Float64()

		sign := 1.0
		if depth%2 == 1 {
			sign = -1.0
		}
		totalFlux += buildBranch(end, angle-sign*spread*0.42, length*lenDecayA, depth+1)
		totalFlux += buildBranch(end, angle+sign*spread*0.95, length*lenDecayB, depth+1)

		// Occasional tertiary feathery corona twig
		if depth >= 2 && depth <= maxDepth-2 && rng.Float64() < 0.48 {
			totalFlux += buildBranch(end, angle-sign*spread*1.25, length*0.52, depth+2)
		}

		segments[idx].flux = totalFlux
		return totalFlux
	}

	// Discharge cathode entry point at bottom center
	cathode := render.Pt(0.50*w, h-margin*1.35)
	maxFlux := buildBranch(cathode, -0.5*math.Pi, 108.0*scale, 1)
	// Two secondary basal dischargeleaders
	_ = buildBranch(cathode, -0.5*math.Pi-0.52, 74.0*scale, 2)
	_ = buildBranch(cathode, -0.5*math.Pi+0.52, 74.0*scale, 2)

	// 3. Render Multi-Pass HDR Plasma Channels (Outer Violet Corona -> Cyan/Gold Body -> White Core)
	rc.SetBlendMode(render.BlendAdd)

	for _, seg := range segments {
		fluxNorm := common.Constrain(math.Pow(seg.flux/math.Max(1.0, maxFlux), 0.44), 0.05, 1.0)
		jPts := jaggedPath(seg.p0, seg.p1, lt.roughness)

		coreW := (0.65 + 5.2*fluxNorm) * scale
		coronaW := coreW * 2.8

		// Outer violet/cobalt corona sheath
		rc.StrokePolyline(jPts, coronaW, violetCorona, 0.18+0.24*fluxNorm)

		// Mid ionized channel (shifts from electric cyan on fine twigs to plasma gold on main trunk)
		midCol := render.LerpOklab(electricCyan, plasmaGold, math.Pow(fluxNorm, 0.75))
		rc.StrokePolyline(jPts, coreW*1.35, midCol, 0.48+0.38*fluxNorm)

		// Incandescent white-gold core filament
		rc.StrokePolyline(jPts, coreW*0.52, incandescentWhite, 0.55+0.42*fluxNorm)
	}

	// Cathode impact flash at the base of the acrylic block
	rc.DrawGlowDot(cathode.X, cathode.Y, 22.0*scale, plasmaGold, 0.65, 0.25)
	rc.DrawGlowDot(cathode.X, cathode.Y, 9.0*scale, incandescentWhite, 0.95, 0.75)

	// Acrylic specimen double border pinstripe
	rc.SetBlendMode(render.BlendOver)
	bezel := &canvas.Path{}
	bezel.MoveTo(margin, margin)
	bezel.LineTo(w-margin, margin)
	bezel.LineTo(w-margin, h-margin)
	bezel.LineTo(margin, h-margin)
	bezel.Close()
	rc.StrokePath(bezel, 1.6*scale, canvas.RoundCap, canvas.RoundJoin, electricCyan, 0.72)

	rc.ApplyBloom(0.72, 5.5*scale, 0.36)
	c.MergeRenderCanvasToneMapped(rc, 1.08, 0.0)
}
