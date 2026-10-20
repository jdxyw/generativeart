package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type packedCapsules struct {
	maxCapsules int
	lineWidth   float64
	minGap      float64
}

// NewPackedCapsules returns a packedCapsules engine that renders Tyler Hobbs-style
// tightly-packed non-overlapping cobalt flow capsules over warm alabaster paper.
func NewPackedCapsules(maxCapsules int, lineWidth, minGap float64) *packedCapsules {
	if maxCapsules <= 0 {
		maxCapsules = 2200
	}
	if lineWidth <= 0 {
		lineWidth = 7.2
	}
	if minGap <= 0 {
		minGap = 3.6
	}
	return &packedCapsules{
		maxCapsules: maxCapsules,
		lineWidth:   lineWidth,
		minGap:      minGap,
	}
}

// Generative renders the non-overlapping cobalt flow capsules onto c.
func (pc *packedCapsules) Generative(c *generativeart.Canva) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	scale := fw / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	maxCapsules := pc.maxCapsules
	if maxCapsules <= 0 {
		maxCapsules = 2600
	}
	// Use crisp capsule thickness and tight uniform breathing gap matching Reference Image 3
	lineWidth := 6.0 * scale
	minGap := 2.5 * scale
	minSep := lineWidth + minGap

	seed := rand.Int63()
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	rc := c.NewRenderCanvas()
	alabasterBg := color.RGBA{R: 238, G: 236, B: 229, A: 255}
	rc.FillBackground(alabasterBg)

	cobaltBlue := color.RGBA{R: 28, G: 76, B: 158, A: 255}

	// Ultra-smooth, low-frequency flow angle field so adjacent capsules run in tight parallel rivers
	// with 2 sweeping macro-bends / bifurcations across the canvas (zero high-frequency wiggle!).
	phaseX := rng.Float64() * 10.0
	phaseY := rng.Float64() * 10.0
	evalDir := func(p render.Point) render.Point {
		nx := p.X / fw
		ny := p.Y / fh
		// Very low frequency 1-octave simplex + smooth harmonic bend
		ang := noise.Simplex2D(nx*1.05+phaseX, ny*1.05+phaseY) * math.Pi * 0.95
		ang += 0.65 * math.Sin(nx*math.Pi*1.7-0.4) * math.Cos(ny*math.Pi*1.5+0.3)
		return render.Pt(math.Cos(ang), math.Sin(ang))
	}

	stepRK4 := func(p render.Point, dt float64) render.Point {
		k1 := evalDir(p)
		k2 := evalDir(p.Add(k1.Mul(0.5 * dt)))
		k3 := evalDir(p.Add(k2.Mul(0.5 * dt)))
		k4 := evalDir(p.Add(k3.Mul(dt)))
		dir := k1.Add(k2.Mul(2.0)).Add(k3.Mul(2.0)).Add(k4).Norm()
		return p.Add(dir.Mul(dt))
	}

	sg := render.NewSpatialGrid(fw, fh, minSep*0.5)
	margin := -lineWidth // Allow capsules to bleed slightly to the canvas edge like Reference Image 3
	inBounds := func(p render.Point) bool {
		return p.X >= margin && p.X <= fw-margin && p.Y >= margin && p.Y <= fh-margin
	}

	// Jobard-Lefer style frontier queue: seed new capsules at exact normal offset +/-(minSep + eps)
	// from placed capsules so streamlines pack into dense, parallel, aligned rivers!
	var queue []render.Point

	// Also build a fine stratified background grid of fallback seeds
	stepGrid := minSep * 0.45
	for y := 0.0; y <= fh; y += stepGrid {
		for x := 0.0; x <= fw; x += stepGrid {
			queue = append(queue, render.Pt(x+(rng.Float64()-0.5)*stepGrid, y+(rng.Float64()-0.5)*stepGrid))
		}
	}
	rng.Shuffle(len(queue), func(i, j int) {
		queue[i], queue[j] = queue[j], queue[i]
	})

	// Seed priority frontier list (checked before grid fallbacks)
	var frontier []render.Point
	frontier = append(frontier, render.Pt(fw*0.5, fh*0.5))

	stepSize := 1.6 * scale
	minCapsuleLen := 11.0 * scale
	sampleSpacing := 1.2 * scale

	rc.SetBlendMode(render.BlendOver)
	placed := 0
	qIdx := 0

	for placed < maxCapsules*2 {
		var p0 render.Point
		if len(frontier) > 0 {
			// Pop from frontier
			last := len(frontier) - 1
			p0 = frontier[last]
			frontier = frontier[:last]
		} else if qIdx < len(queue) {
			p0 = queue[qIdx]
			qIdx++
		} else {
			break
		}

		if !inBounds(p0) || !sg.CanPlace(p0, minSep) {
			continue
		}

		// Mix of medium flowing segments (25..95px) and short pills (12..28px) to fill bifurcations
		var targetLen float64
		if rng.Float64() < 0.72 {
			targetLen = (26.0 + rng.Float64()*72.0) * scale
		} else {
			targetLen = (12.0 + rng.Float64()*18.0) * scale
		}
		halfSteps := int(math.Ceil((targetLen * 0.5) / stepSize))

		fwd := make([]render.Point, 0, halfSteps)
		curr := p0
		for s := 0; s < halfSteps; s++ {
			next := stepRK4(curr, stepSize)
			if !inBounds(next) || !sg.CanPlace(next, minSep) {
				break
			}
			fwd = append(fwd, next)
			curr = next
		}

		bwd := make([]render.Point, 0, halfSteps)
		curr = p0
		for s := 0; s < halfSteps; s++ {
			next := stepRK4(curr, -stepSize)
			if !inBounds(next) || !sg.CanPlace(next, minSep) {
				break
			}
			bwd = append(bwd, next)
			curr = next
		}

		totalSteps := len(bwd) + len(fwd)
		if float64(totalSteps)*stepSize < minCapsuleLen {
			continue
		}

		pts := make([]render.Point, 0, totalSteps+1)
		for i := len(bwd) - 1; i >= 0; i-- {
			pts = append(pts, bwd[i])
		}
		pts = append(pts, p0)
		pts = append(pts, fwd...)

		sg.InsertPolyline(pts, sampleSpacing)

		// Enqueue normal-offset neighbor seeds along the newly placed capsule so adjacent capsules pack tightly in parallel!
		dSeed := minSep * 1.03
		for i := 0; i < len(pts)-1; i += 3 {
			tan := pts[i+1].Sub(pts[i]).Norm()
			norm := tan.Perp()
			frontier = append(frontier,
				pts[i].Add(norm.Mul(dSeed)),
				pts[i].Sub(norm.Mul(dSeed)),
			)
		}
		// Also enqueue inline continuation seeds just beyond the two capsule tips!
		if len(pts) >= 2 {
			tailDir := pts[0].Sub(pts[1]).Norm()
			headDir := pts[len(pts)-1].Sub(pts[len(pts)-2]).Norm()
			frontier = append(frontier,
				pts[0].Add(tailDir.Mul(minSep*1.08)),
				pts[len(pts)-1].Add(headDir.Mul(minSep*1.08)),
			)
		}

		lab := render.RGBAToOklab(cobaltBlue)
		dL := (rng.Float64()*2.0 - 1.0) * 0.015
		capCol := render.OklabToRGBA(render.Oklab{
			L: common.Constrain(lab.L+dL, 0.15, 0.90),
			A: lab.A,
			B: lab.B,
		}, 255)

		rc.DrawCapsule(pts, lineWidth, capCol, 0.99)
		placed++
	}

	rc.ApplyPaperTexture(0.14, 103)
	c.MergeRenderCanvas(rc)
}
