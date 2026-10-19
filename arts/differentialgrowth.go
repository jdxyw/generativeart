package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type differentialGrowth struct {
	steps                             int
	maxNodes                          int
	minDist, maxDist, repulsionRadius float64
}

// NewDifferentialGrowth returns a differentialGrowth engine that simulates an elastic
// closed contour undergoing differential growth with O(N) spatial hash repulsion.
func NewDifferentialGrowth(steps, maxNodes int, minDist, maxDist, repulsionRadius float64) *differentialGrowth {
	if steps <= 0 {
		steps = 220
	}
	if maxNodes <= 0 {
		maxNodes = 1600
	}
	if minDist <= 0 {
		minDist = 3.5
	}
	if maxDist <= 0 {
		maxDist = 7.5
	}
	if repulsionRadius <= 0 {
		repulsionRadius = 18.0
	}
	return &differentialGrowth{
		steps:           steps,
		maxNodes:        maxNodes,
		minDist:         minDist,
		maxDist:         maxDist,
		repulsionRadius: repulsionRadius,
	}
}

// Generative runs the differential growth curve simulation and renders layered organic coral folds.
func (dg *differentialGrowth) Generative(c *generativeart.Canva) {
	if dg.steps <= 0 {
		dg.steps = 220
	}
	if dg.maxNodes <= 0 {
		dg.maxNodes = 1600
	}
	if dg.minDist <= 0 {
		dg.minDist = 3.5
	}
	if dg.maxDist <= 0 {
		dg.maxDist = 7.5
	}
	if dg.repulsionRadius <= 0 {
		dg.repulsionRadius = 18.0
	}

	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	scale := fw / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	minDist := dg.minDist * scale
	maxDist := dg.maxDist * scale
	repRadius := dg.repulsionRadius * scale
	if repRadius < 2.0 {
		repRadius = 2.0
	}
	if maxDist <= minDist {
		maxDist = minDist * 2.0
	}

	seed := rand.Int63()
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	cx, cy := fw*0.5, fh*0.5
	initNodes := 64
	baseRadius := 68.0 * scale

	nodes := make([]render.Point, initNodes)
	for i := 0; i < initNodes; i++ {
		ang := (float64(i) / float64(initNodes)) * 2.0 * math.Pi
		wave := 1.0 + 0.20*math.Sin(ang*5.0) + 0.15*math.Cos(ang*3.0) + 0.10*math.Sin(ang*7.0+0.9) + (rng.Float64()-0.5)*0.04
		r := baseRadius * wave
		nodes[i] = render.Pt(cx+r*math.Cos(ang), cy+r*math.Sin(ang))
	}

	// Spatial hash grid sizing
	cellSize := repRadius * 1.25
	invCell := 1.0 / cellSize
	gridW := int(math.Ceil(fw*invCell)) + 2
	gridH := int(math.Ceil(fh*invCell)) + 2
	if gridW < 4 {
		gridW = 4
	}
	if gridH < 4 {
		gridH = 4
	}
	head := make([]int, gridW*gridH)
	nextNode := make([]int, dg.maxNodes+initNodes+64)
	disp := make([]render.Point, dg.maxNodes+initNodes+64)

	var snapshots [][]render.Point
	recordSnapshot := func(pts []render.Point) {
		cp := make([]render.Point, len(pts))
		copy(cp, pts)
		snapshots = append(snapshots, cp)
	}
	recordSnapshot(nodes)

	margin := 34.0 * scale
	maxBoundR := math.Min(fw, fh)*0.5 - margin
	restLen := (minDist + maxDist) * 0.54
	effRepRadius := cellSize

	for step := 0; step < dg.steps; step++ {
		n := len(nodes)
		if n < 3 {
			break
		}
		if len(nextNode) < n {
			nextNode = make([]int, n*2)
			disp = make([]render.Point, n*2)
		}

		// Clear spatial hash grid headers
		for i := range head {
			head[i] = -1
		}
		for i := 0; i < n; i++ {
			gx := common.ConstrainInt(int(nodes[i].X*invCell), 0, gridW-1)
			gy := common.ConstrainInt(int(nodes[i].Y*invCell), 0, gridH-1)
			cell := gy*gridW + gx
			nextNode[i] = head[cell]
			head[cell] = i
		}

		repRadiusSq := effRepRadius * effRepRadius
		for i := 0; i < n; i++ {
			prev := nodes[(i+n-1)%n]
			curr := nodes[i]
			next := nodes[(i+1)%n]

			// 1. Elastic fairing attraction toward midpoint of neighbors (prevents jagged kinks)
			mid := prev.Lerp(next, 0.5)
			springForce := mid.Sub(curr).Mul(0.16)

			// Edge rest-length regulation: push apart when compressed below restLen, pull when stretched
			for _, nb := range [2]render.Point{prev, next} {
				dVec := curr.Sub(nb)
				d := dVec.Len()
				if d > 1e-6 {
					diff := (restLen - d) / restLen
					springForce = springForce.Add(dVec.Mul((0.30 * diff) / d))
				}
			}

			// 2. Repulsion away from all non-adjacent nodes within effRepRadius
			var repForce render.Point
			gx := common.ConstrainInt(int(curr.X*invCell), 0, gridW-1)
			gy := common.ConstrainInt(int(curr.Y*invCell), 0, gridH-1)

			for dy := -1; dy <= 1; dy++ {
				ny := gy + dy
				if ny < 0 || ny >= gridH {
					continue
				}
				rowOff := ny * gridW
				for dx := -1; dx <= 1; dx++ {
					nx := gx + dx
					if nx < 0 || nx >= gridW {
						continue
					}
					for j := head[rowOff+nx]; j != -1; j = nextNode[j] {
						// Ignore self and immediate 2-hop neighbors along the contour
						diffIdx := (j - i + n) % n
						if diffIdx <= 2 || diffIdx >= n-2 {
							continue
						}
						diffX := curr.X - nodes[j].X
						diffY := curr.Y - nodes[j].Y
						d2 := diffX*diffX + diffY*diffY
						if d2 > 1e-9 && d2 < repRadiusSq {
							dist := math.Sqrt(d2)
							wRep := 1.0 - dist/effRepRadius
							wRep = wRep * wRep
							mag := (2.35 * scale * wRep) / dist
							repForce.X += diffX * mag
							repForce.Y += diffY * mag
						}
					}
				}
			}

			// 3. Simplex noise + gentle radial expansion + outward normal buckling bias
			tangent := next.Sub(prev)
			var normal render.Point
			if tangent.Len() > 1e-6 {
				normal = tangent.Norm().Perp()
			}
			radVec := render.Pt(curr.X-cx, curr.Y-cy)
			var outward render.Point
			if radVec.Len() > 1e-6 {
				outward = radVec.Norm()
			}
			nVal := noise.Simplex2D(curr.X*(0.008/scale)+float64(step)*0.012, curr.Y*(0.008/scale))
			growthForce := outward.Mul(0.38 * scale).Add(normal.Mul(nVal * 0.42 * scale))

			total := springForce.Add(repForce).Add(growthForce)
			maxStepDisp := 3.2 * scale
			if total.Len() > maxStepDisp {
				total = total.Norm().Mul(maxStepDisp)
			}
			disp[i] = total
		}

		for i := 0; i < n; i++ {
			np := nodes[i].Add(disp[i])
			// Soft radial containment inside canvas margins
			rx, ry := np.X-cx, np.Y-cy
			rDist := math.Hypot(rx, ry)
			if rDist > maxBoundR && rDist > 1e-6 {
				pull := (rDist - maxBoundR) * 0.55
				np.X -= (rx / rDist) * pull
				np.Y -= (ry / rDist) * pull
			}
			np.X = common.Constrain(np.X, margin*0.5, fw-margin*0.5)
			np.Y = common.Constrain(np.Y, margin*0.5, fh-margin*0.5)
			nodes[i] = np
		}

		// Differential growth injection: split edges exceeding maxDist, plus nutrient-driven
		// subdivision along curved/active segments so the perimeter continuously buckles.
		if len(nodes) < dg.maxNodes {
			targetAdd := common.MaxInt(4, (dg.maxNodes-initNodes)/common.MaxInt(dg.steps-20, 1))
			added := 0
			grown := make([]render.Point, 0, common.MinInt(len(nodes)+targetAdd*2, dg.maxNodes))
			nCur := len(nodes)
			for i := 0; i < nCur; i++ {
				p0 := nodes[(i+nCur-1)%nCur]
				p1 := nodes[i]
				p2 := nodes[(i+1)%nCur]
				grown = append(grown, p1)
				if len(grown)+(nCur-1-i) >= dg.maxNodes {
					continue
				}
				edgeLen := p1.Dist(p2)
				// Curvature estimate at p1
				v1 := p1.Sub(p0).Norm()
				v2 := p2.Sub(p1).Norm()
				curv := 1.0 - 0.5*(v1.X*v2.X+v1.Y*v2.Y+1.0) // [0..1]
				nutrient := 0.5 + 0.5*noise.Simplex2D(p1.X*(0.012/scale), p1.Y*(0.012/scale)+float64(step)*0.02)
				splitScore := curv*0.6 + nutrient*0.4
				shouldSplit := edgeLen > maxDist || (edgeLen > minDist*0.85 && added < targetAdd && rng.Float64() < 0.08+0.22*splitScore)
				if shouldSplit {
					mid := p1.Lerp(p2, 0.5)
					edgeNormal := p2.Sub(p1).Norm().Perp()
					jitter := (rng.Float64() - 0.5) * 0.35 * scale
					grown = append(grown, mid.Add(edgeNormal.Mul(jitter)))
					added++
				}
			}
			nodes = grown
		}

		if (step+1)%3 == 0 {
			recordSnapshot(nodes)
		}
	}

	if len(snapshots) == 0 || len(snapshots[len(snapshots)-1]) != len(nodes) {
		recordSnapshot(nodes)
	}

	schema := c.Opts().ColorSchema()
	if len(schema) == 0 {
		schema = common.Sunset
	}
	grad := render.NewMultiStopGradient(schema...)
	ukiyoGrad := render.NewMultiStopGradient(common.UkiyoE...)
	rc := c.NewRenderCanvas()

	// Soft radial glow in background.
	rc.SetBlendMode(render.BlendAdd)
	rc.DrawGlowDot(cx, cy, math.Min(fw, fh)*0.46, grad.SampleOklab(0.18), 0.20, 0.0)
	rc.DrawGlowDot(cx, cy, math.Min(fw, fh)*0.24, ukiyoGrad.SampleOklab(0.55), 0.14, 0.0)

	// Helper to build a closed control polygon suitable for smooth Catmull-Rom evaluation.
	closedSplinePts := func(loop []render.Point) []render.Point {
		m := len(loop)
		if m < 3 {
			return loop
		}
		pts := make([]render.Point, 0, m+3)
		pts = append(pts, loop...)
		pts = append(pts, loop[0], loop[1], loop[2])
		return pts
	}

	numSnaps := len(snapshots)
	denom := float64(common.MaxInt(numSnaps-1, 1))

	for idx, snap := range snapshots {
		u := float64(idx) / denom
		ctrlPts := closedSplinePts(snap)
		curve := render.CatmullRomSpline(ctrlPts, 8)

		baseCol := grad.SampleOklab(u)
		accentCol := ukiyoGrad.SampleOklab(common.Clamp01(0.15 + 0.80*u))
		strokeCol := render.LerpOklab(baseCol, accentCol, 0.25*math.Sin(u*math.Pi))
		strokeCol = render.LerpOklab(color.RGBA{R: 18, G: 14, B: 34, A: 255}, strokeCol, 0.35+0.65*math.Pow(u, 0.75))

		// Translucent dark-tinted filled contour every few snapshots for volumetric depth.
		if idx%6 == 0 {
			rc.SetBlendMode(render.BlendOver)
			fillCol := render.LerpOklab(baseCol, color.RGBA{R: 12, G: 10, B: 24, A: 255}, 0.72)
			rc.FillPolygon(curve, fillCol, 0.14)
		}

		// Crisp contour stroke using BlendOver so inner growth rings stay distinct.
		rc.SetBlendMode(render.BlendOver)
		strokeW := (0.95 + 1.05*u) * scale
		strokeAlpha := 0.32 + 0.50*math.Pow(u, 1.2)
		rc.StrokePolyline(curve, strokeW, strokeCol, strokeAlpha)

		if idx%8 == 0 && u > 0.35 {
			rc.SetBlendMode(render.BlendScreen)
			uCapture := u
			rc.DrawTaperedSpline(
				ctrlPts,
				(1.2+1.2*u)*scale,
				render.TaperUniform,
				func(t float64) color.RGBA {
					return grad.SampleOklab(common.Clamp01(uCapture*0.78 + 0.22*math.Sin(t*2.0*math.Pi)*0.5 + 0.10))
				},
				0.22*u,
			)
		}
	}

	// Draw the final outer ruffled loop with a crisp luminous stroke.
	finalPts := closedSplinePts(snapshots[numSnaps-1])
	finalCurve := render.CatmullRomSpline(finalPts, 12)
	finalBaseCol := grad.SampleOklab(0.95)
	luminousEdge := render.LerpOklab(finalBaseCol, color.RGBA{R: 255, G: 250, B: 235, A: 255}, 0.55)

	rc.SetBlendMode(render.BlendOver)
	rc.StrokePolyline(finalCurve, 2.4*scale, finalBaseCol, 0.92)
	rc.SetBlendMode(render.BlendAdd)
	rc.StrokePolyline(finalCurve, 1.25*scale, luminousEdge, 0.65)

	rc.ApplyBloom(0.62, 6.0, 0.35)
	rc.ApplyVignette(0.3)
	c.MergeRenderCanvasToneMapped(rc, 1.08, 0)
}
