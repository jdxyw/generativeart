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

type truchetFlow struct {
	baseGrid int
	maxDepth int
}

type truchetCell struct {
	x, y  float64
	size  float64
	depth int
	gx    int
	gy    int
}

// NewTruchetFlow returns a truchetFlow engine that renders multi-scale quadtree
// Winged/Smith Truchet ribbon weaves with 2.5D under-shadows and Oklab gradients.
func NewTruchetFlow(baseGrid int, maxDepth int) *truchetFlow {
	if baseGrid <= 0 {
		baseGrid = 6
	}
	if maxDepth <= 0 {
		maxDepth = 3
	}
	return &truchetFlow{
		baseGrid: baseGrid,
		maxDepth: maxDepth,
	}
}

// buildArcPath constructs an exact circular arc on a tdewolff/canvas Path centered at (cx, cy)
// with the given radius from angle theta0 to theta1 (in radians) using two cubic Bézier arc segments.
func buildArcPath(cx, cy, radius, theta0, theta1 float64) *canvas.Path {
	p := &canvas.Path{}
	if radius <= 0 {
		return p
	}
	const segments = 2
	dTheta := (theta1 - theta0) / float64(segments)
	// Exact cubic Bézier control point distance for circular arc of angle dTheta
	k := (4.0 / 3.0) * math.Tan(dTheta*0.25) * radius

	startX := cx + radius*math.Cos(theta0)
	startY := cy + radius*math.Sin(theta0)
	p.MoveTo(startX, startY)

	for s := 0; s < segments; s++ {
		a0 := theta0 + float64(s)*dTheta
		a1 := a0 + dTheta
		c0, s0 := math.Cos(a0), math.Sin(a0)
		c1, s1 := math.Cos(a1), math.Sin(a1)

		x0, y0 := cx+radius*c0, cy+radius*s0
		x3, y3 := cx+radius*c1, cy+radius*s1

		cp1x, cp1y := x0-k*s0, y0+k*c0
		cp2x, cp2y := x3+k*s1, y3-k*c1
		p.CubeTo(cp1x, cp1y, cp2x, cp2y, x3, y3)
	}
	return p
}

// buildLinePath constructs a straight segment Path from (x0, y0) to (x1, y1).
func buildLinePath(x0, y0, x1, y1 float64) *canvas.Path {
	p := &canvas.Path{}
	p.MoveTo(x0, y0)
	p.LineTo(x1, y1)
	return p
}

// Generative subdivides the canvas into a multi-scale quadtree of Winged/Smith Truchet tiles
// and renders 5-band concentric weaving ribbons with 2.5D crossing shadows and paper texture.
func (tf *truchetFlow) Generative(c *generativeart.Canva) {
	baseGrid := tf.baseGrid
	if baseGrid <= 0 {
		baseGrid = 6
	}
	maxDepth := tf.maxDepth
	if maxDepth <= 0 {
		maxDepth = 3
	}

	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	seed := rand.Int63()
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	schema := c.Opts().ColorSchema()
	if len(schema) < 2 {
		schema = common.UkiyoE
	}
	grad := render.NewMultiStopGradient(schema...)
	dunhuangGrad := render.NewMultiStopGradient(common.Dunhuang...)

	rc := c.NewRenderCanvas()

	// 1. Rich warm indigo-parchment backdrop with subtle radial & FBM modulation
	bgDark := color.RGBA{R: 20, G: 28, B: 38, A: 255}
	bgMid := render.LerpOklab(bgDark, grad.SampleOklab(0.12), 0.35)
	cx, cy := fw*0.5, fh*0.5
	maxR := math.Hypot(cx, cy)

	for py := 0; py < h; py++ {
		fy := float64(py)
		for px := 0; px < w; px++ {
			fx := float64(px)
			radT := common.Clamp01(math.Hypot(fx-cx, fy-cy) / math.Max(1.0, maxR))
			nVal := 0.5 + 0.5*noise.Simplex2D(fx*0.004/scale, fy*0.004/scale)
			col := render.LerpOklab(bgMid, bgDark, 0.65*radT+0.35*(1.0-nVal))
			rLin, gLin, bLin, _ := render.ColorToLinearRGB(col)
			rc.SetPixelLinear(px, py, rLin, gLin, bLin, 1.0)
		}
	}

	// 2. Build baseGrid x baseGrid square cells and recursively quadtree-subdivide up to maxDepth
	// driven by a smooth FBM + radial focal field.
	cellSizeX := fw / float64(baseGrid)
	cellSizeY := fh / float64(baseGrid)
	var cells []truchetCell

	focalX := fw * (0.42 + 0.16*rng.Float64())
	focalY := fh * (0.42 + 0.16*rng.Float64())

	var subdivide func(x, y, sx, sy float64, depth, gx, gy int)
	subdivide = func(x, y, sx, sy float64, depth, gx, gy int) {
		midX := x + sx*0.5
		midY := y + sy*0.5
		if depth < maxDepth {
			fbm := 0.5 + 0.5*noise.FBM2D(midX*0.006/scale, midY*0.006/scale, 4, 2.0, 0.5)
			distFocal := math.Hypot(midX-focalX, midY-focalY) / math.Max(1.0, math.Min(fw, fh)*0.65)
			radialScore := 1.0 - common.Clamp01(distFocal*0.75)
			detailScore := 0.55*fbm + 0.45*radialScore
			threshold := 0.34 + 0.15*float64(depth)

			if detailScore > threshold {
				halfX := sx * 0.5
				halfY := sy * 0.5
				subdivide(x, y, halfX, halfY, depth+1, gx*2, gy*2)
				subdivide(x+halfX, y, halfX, halfY, depth+1, gx*2+1, gy*2)
				subdivide(x, y+halfY, halfX, halfY, depth+1, gx*2, gy*2+1)
				subdivide(x+halfX, y+halfY, halfX, halfY, depth+1, gx*2+1, gy*2+1)
				return
			}
		}
		cells = append(cells, truchetCell{
			x:     x,
			y:     y,
			size:  math.Min(sx, sy),
			depth: depth,
			gx:    gx,
			gy:    gy,
		})
	}

	for gy := 0; gy < baseGrid; gy++ {
		for gx := 0; gx < baseGrid; gx++ {
			subdivide(float64(gx)*cellSizeX, float64(gy)*cellSizeY, cellSizeX, cellSizeY, 0, gx, gy)
		}
	}

	const numBands = 5
	shadowCol := color.RGBA{R: 8, G: 11, B: 16, A: 255}
	frameCol := render.LerpOklab(grad.SampleOklab(0.5), color.RGBA{R: 220, G: 210, B: 190, A: 255}, 0.4)

	// Helper that renders a 5-band concentric quarter-circle arc with a 2.5D dark under-shadow stroke
	drawArcRibbon := func(cornerX, cornerY, baseRadius, theta0, theta1, bandSpacing, bandStrokeW, colorBaseT, colorShift float64) {
		// 1. Subtle dark 2.5D under-shadow casing behind the entire 5-band ribbon bundle
		totalSpan := float64(numBands-1)*bandSpacing + bandStrokeW
		shadowPath := buildArcPath(cornerX, cornerY, baseRadius, theta0, theta1)
		rc.SetBlendMode(render.BlendMultiply)
		rc.StrokePath(shadowPath, totalSpan+bandStrokeW*2.2, canvas.ButtCap, canvas.RoundJoin, shadowCol, 0.52)
		rc.SetBlendMode(render.BlendOver)
		rc.StrokePath(shadowPath, totalSpan+bandStrokeW*1.1, canvas.ButtCap, canvas.RoundJoin, shadowCol, 0.68)

		// 2. Five concentric parallel ribbon bands in Oklab color space
		for b := 0; b < numBands; b++ {
			offsetIdx := float64(b) - float64(numBands-1)*0.5
			rad := baseRadius + offsetIdx*bandSpacing
			if rad <= 0.5 {
				continue
			}
			arcPath := buildArcPath(cornerX, cornerY, rad, theta0, theta1)

			// Symmetric or progressive color ramp across the 5 concentric bands
			bandU := float64(b) / float64(numBands-1)
			symU := 1.0 - math.Abs(2.0*bandU-1.0) // 0 at outer bands, 1 at center band
			sampleT := common.Clamp01(colorBaseT*0.65 + bandU*0.35 + colorShift)

			colPrimary := grad.SampleOklab(sampleT)
			colAccent := dunhuangGrad.SampleOklab(common.Clamp01(1.0 - sampleT))
			bandCol := render.LerpOklab(colPrimary, colAccent, 0.28*symU)
			if b == 2 {
				// Highlight central spine band for crisp visual definition
				bandCol = render.LerpOklab(bandCol, color.RGBA{R: 250, G: 242, B: 222, A: 255}, 0.32)
			}

			rc.SetBlendMode(render.BlendOver)
			rc.StrokePath(arcPath, bandStrokeW, canvas.RoundCap, canvas.RoundJoin, bandCol, 0.94)
		}
	}

	// Helper that renders a 5-band straight ribbon (for cross-weave Truchet variant) with 2.5D under-shadow
	drawStraightRibbon := func(x0, y0, x1, y1, nx, ny, bandSpacing, bandStrokeW, colorBaseT, colorShift float64) {
		totalSpan := float64(numBands-1)*bandSpacing + bandStrokeW
		centerPath := buildLinePath(x0, y0, x1, y1)
		rc.SetBlendMode(render.BlendMultiply)
		rc.StrokePath(centerPath, totalSpan+bandStrokeW*2.2, canvas.ButtCap, canvas.RoundJoin, shadowCol, 0.55)
		rc.SetBlendMode(render.BlendOver)
		rc.StrokePath(centerPath, totalSpan+bandStrokeW*1.1, canvas.ButtCap, canvas.RoundJoin, shadowCol, 0.72)

		for b := 0; b < numBands; b++ {
			offsetIdx := float64(b) - float64(numBands-1)*0.5
			ox := nx * offsetIdx * bandSpacing
			oy := ny * offsetIdx * bandSpacing
			linePath := buildLinePath(x0+ox, y0+oy, x1+ox, y1+oy)

			bandU := float64(b) / float64(numBands-1)
			symU := 1.0 - math.Abs(2.0*bandU-1.0)
			sampleT := common.Clamp01(colorBaseT*0.65 + bandU*0.35 + colorShift)
			colPrimary := grad.SampleOklab(sampleT)
			colAccent := dunhuangGrad.SampleOklab(common.Clamp01(1.0 - sampleT))
			bandCol := render.LerpOklab(colPrimary, colAccent, 0.28*symU)
			if b == 2 {
				bandCol = render.LerpOklab(bandCol, color.RGBA{R: 250, G: 242, B: 222, A: 255}, 0.32)
			}

			rc.SetBlendMode(render.BlendOver)
			rc.StrokePath(linePath, bandStrokeW, canvas.RoundCap, canvas.RoundJoin, bandCol, 0.94)
		}
	}

	// 3. Render every quadtree Truchet tile
	for _, cell := range cells {
		x, y, s := cell.x, cell.y, cell.size
		midX, midY := x+s*0.5, y+s*0.5
		rBase := s * 0.5

		// Subtle quadtree tile border and recessed background wash
		tileNoise := 0.5 + 0.5*noise.FBM2D(midX*0.005/scale, midY*0.005/scale, 3, 2.0, 0.5)
		tileTone := render.LerpOklab(bgDark, grad.SampleOklab(tileNoise), 0.14)
		tilePoly := []render.Point{
			{X: x, Y: y},
			{X: x + s, Y: y},
			{X: x + s, Y: y + s},
			{X: x, Y: y + s},
		}
		rc.SetBlendMode(render.BlendOver)
		rc.FillPolygon(tilePoly, tileTone, 0.42)
		rc.StrokePolyline(append(tilePoly, tilePoly[0]), math.Max(0.5, 0.65*scale), frameCol, 0.14)

		// Proportioned ribbon spacing & width so ribbons visually weave and kiss across the tile
		bandSpacing := s * 0.064
		bandStrokeW := math.Max(0.75*scale, bandSpacing*0.64)

		// Choose Truchet orientation & variant deterministically from spatial hash + flow noise
		flowAngle := noise.Simplex2D(midX*0.004/scale+11.3, midY*0.004/scale+7.9)
		variantRoll := rng.Float64()

		if variantRoll < 0.40 || (flowAngle >= 0 && variantRoll < 0.72) {
			// Variant A: Quarter-circle pair at Top-Left & Bottom-Right, plus an under-weaving counter-arc
			// when depth <= 2 so ribbons visibly weave over and under in 2.5D
			if cell.depth <= 2 && rng.Float64() < 0.65 {
				drawArcRibbon(x+s, y, rBase, math.Pi*0.5, math.Pi, bandSpacing, bandStrokeW, tileNoise, -0.12)
			}
			drawArcRibbon(x, y, rBase, 0, math.Pi*0.5, bandSpacing, bandStrokeW, tileNoise, 0.0)
			drawArcRibbon(x+s, y+s, rBase, math.Pi, math.Pi*1.5, bandSpacing, bandStrokeW, tileNoise, 0.08)
		} else if variantRoll < 0.84 {
			// Variant B: Quarter-circle pair at Top-Right & Bottom-Left, with optional under-weaving arc
			if cell.depth <= 2 && rng.Float64() < 0.65 {
				drawArcRibbon(x, y, rBase, 0, math.Pi*0.5, bandSpacing, bandStrokeW, tileNoise, -0.12)
			}
			drawArcRibbon(x+s, y, rBase, math.Pi*0.5, math.Pi, bandSpacing, bandStrokeW, tileNoise, 0.04)
			drawArcRibbon(x, y+s, rBase, math.Pi*1.5, math.Pi*2.0, bandSpacing, bandStrokeW, tileNoise, 0.12)
		} else {
			// Variant C: 2.5D Over-Under Orthogonal Cross Weave connecting opposite edge midpoints
			if rng.Float64() < 0.5 {
				drawStraightRibbon(x, midY, x+s, midY, 0, 1, bandSpacing, bandStrokeW, tileNoise, -0.08)
				drawStraightRibbon(midX, y, midX, y+s, 1, 0, bandSpacing, bandStrokeW, tileNoise, 0.10)
			} else {
				drawStraightRibbon(midX, y, midX, y+s, 1, 0, bandSpacing, bandStrokeW, tileNoise, -0.08)
				drawStraightRibbon(x, midY, x+s, midY, 0, 1, bandSpacing, bandStrokeW, tileNoise, 0.10)
			}
		}

		// Winged multi-scale decorative concentric dots at tile center and corner/edge midpoints
		dotOuterCol := grad.SampleOklab(common.Clamp01(tileNoise + 0.18))
		dotInnerCol := render.LerpOklab(dunhuangGrad.SampleOklab(0.35), color.RGBA{R: 255, G: 245, B: 225, A: 255}, 0.45)
		centerDotR := math.Max(1.4*scale, s*0.068)

		rc.SetBlendMode(render.BlendOver)
		rc.DrawGlowDot(midX, midY, centerDotR*1.35, shadowCol, 0.55, 0.75)
		rc.DrawGlowDot(midX, midY, centerDotR, dotOuterCol, 0.90, 0.82)
		rc.DrawGlowDot(midX, midY, centerDotR*0.46, dotInnerCol, 0.95, 0.88)
		rc.DrawSoftRing(midX, midY, centerDotR*1.45, math.Max(0.7*scale, s*0.015), dotInnerCol, 0.55)

		// Edge-port Winged Truchet node pins at the 4 edge midpoints
		portDotR := math.Max(1.0*scale, s*0.036)
		edgePorts := [4]render.Point{
			{X: midX, Y: y},
			{X: x + s, Y: midY},
			{X: midX, Y: y + s},
			{X: x, Y: midY},
		}
		for _, ep := range edgePorts {
			rc.DrawGlowDot(ep.X, ep.Y, portDotR*1.2, shadowCol, 0.45, 0.8)
			rc.DrawGlowDot(ep.X, ep.Y, portDotR*0.75, dotInnerCol, 0.82, 0.85)
		}
	}

	// 4. Finish with cotton paper fiber texture, soft vignette, and sRGB resolve
	rc.ApplyPaperTexture(0.42, 777)
	rc.ApplyVignette(0.22)
	c.MergeRenderCanvas(rc)
}
