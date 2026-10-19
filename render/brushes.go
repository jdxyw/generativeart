package render

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart/common"
)

func ensureRNG(rng *rand.Rand) *rand.Rand {
	if rng != nil {
		return rng
	}
	return rand.New(rand.NewSource(42))
}

// DrawSandLine renders an Inconvergent-style stochastic sand-painted line segment between p1 and p2
// by splatting sub-pixel grains with Gaussian perpendicular/longitudinal jitter.
func (c *Canvas) DrawSandLine(p1, p2 Point, grains int, jitter float64, col color.RGBA, alpha float64, rng *rand.Rand) {
	if c == nil || grains <= 0 || alpha <= 0 || col.A == 0 {
		return
	}
	rng = ensureRNG(rng)
	dir := p2.Sub(p1)
	length := dir.Len()
	var tan, norm Point
	if length > 1e-9 {
		tan = dir.Mul(1.0 / length)
		norm = tan.Perp()
	} else {
		tan = Point{X: 1, Y: 0}
		norm = Point{X: 0, Y: 1}
	}

	// Evolving sinusoidal phase gives organic density modulation along the sand line.
	phase := rng.Float64() * 2.0 * math.Pi
	freq := 1.5 + rng.Float64()*2.5

	for i := 0; i < grains; i++ {
		// Beta-like / sinusoidal biased sampling produces characteristic sand-painter silkiness.
		u := rng.Float64()
		t := 0.5 * (1.0 - math.Cos(u*math.Pi))
		base := p1.Lerp(p2, t)

		envelope := 0.35 + 0.65*math.Sin(t*math.Pi*freq+phase)
		latOffset := rng.NormFloat64() * jitter * envelope
		lonOffset := rng.NormFloat64() * (jitter * 0.35)

		pt := base.Add(norm.Mul(latOffset)).Add(tan.Mul(lonOffset))
		c.SplatPoint(pt.X, pt.Y, col, alpha)
	}
}

// DrawSandSpline renders a stochastic sand-painter stroke along a smooth Catmull-Rom spline through pts
// with a progressive sinusoidal/noise wave envelope and colorFn(t) evaluation.
func (c *Canvas) DrawSandSpline(
	pts []Point,
	grainsPerStep int,
	jitter float64,
	colorFn func(t float64) color.RGBA,
	alpha float64,
	rng *rand.Rand,
) {
	if c == nil || len(pts) < 2 || grainsPerStep <= 0 || alpha <= 0 {
		return
	}
	rng = ensureRNG(rng)
	if colorFn == nil {
		colorFn = func(float64) color.RGBA { return common.White }
	}

	curve := CatmullRomSpline(pts, 28)
	m := len(curve)
	if m < 2 {
		return
	}

	arcLen := make([]float64, m)
	totalLen := 0.0
	for i := 1; i < m; i++ {
		totalLen += curve[i].Dist(curve[i-1])
		arcLen[i] = totalLen
	}
	if totalLen < 1e-6 {
		return
	}

	// Progressive wave envelope state across the spline
	phase1 := rng.Float64() * 2.0 * math.Pi
	phase2 := rng.Float64() * 2.0 * math.Pi
	sweepAmp := jitter

	for i := 0; i < m-1; i++ {
		p0 := curve[i]
		p1 := curve[i+1]
		segDir := p1.Sub(p0)
		segLen := segDir.Len()
		if segLen < 1e-9 {
			continue
		}
		tan := segDir.Mul(1.0 / segLen)
		norm := tan.Perp()

		t0 := arcLen[i] / totalLen
		t1 := arcLen[i+1] / totalLen

		// Scale steps by segment pixel length so grain density is uniform per unit arc length.
		subSteps := int(math.Max(1, math.Ceil(segLen*1.5)))
		for s := 0; s < subSteps; s++ {
			frac := (float64(s) + 0.5) / float64(subSteps)
			t := t0 + (t1-t0)*frac
			center := p0.Lerp(p1, frac)
			col := colorFn(t)

			// Progressive harmonic wave envelope
			wave := math.Sin(t*math.Pi*4.0+phase1)*0.6 + math.Cos(t*math.Pi*9.0+phase2)*0.4
			maxSpread := sweepAmp * (0.25 + 0.85*math.Abs(wave)) * math.Sin(math.Pi*common.Constrain(t, 0.02, 0.98))

			for g := 0; g < grainsPerStep; g++ {
				// Inconvergent-style progressive sweep + Gaussian scatter
				r := float64(g+1) / float64(grainsPerStep)
				sweep := math.Sin(r*math.Pi*0.5) * wave * maxSpread
				scatter := rng.NormFloat64() * (0.35 * maxSpread) * r
				lonJitter := rng.NormFloat64() * 0.35

				pos := center.Add(norm.Mul(sweep + scatter)).Add(tan.Mul(lonJitter))
				grainAlpha := alpha * (1.0 - 0.45*r)
				c.SplatPoint(pos.X, pos.Y, col, grainAlpha)
			}
		}
	}
}

// DrawBristleStroke simulates a multi-filament natural bristle brush along pts.
// Each bristle has a lateral offset across width, independent harmonic drift,
// slight pigment variation in Oklab space, and individual filament thickness/opacity.
func (c *Canvas) DrawBristleStroke(
	pts []Point,
	width float64,
	bristles int,
	col color.RGBA,
	alpha float64,
	rng *rand.Rand,
) {
	if c == nil || len(pts) < 2 || width <= 0 || bristles <= 0 || alpha <= 0 || col.A == 0 {
		return
	}
	rng = ensureRNG(rng)

	curve := CatmullRomSpline(pts, 20)
	m := len(curve)
	if m < 2 {
		return
	}

	arcLen := make([]float64, m)
	totalLen := 0.0
	for i := 1; i < m; i++ {
		totalLen += curve[i].Dist(curve[i-1])
		arcLen[i] = totalLen
	}
	if totalLen < 1e-6 {
		return
	}

	normals := make([]Point, m)
	for i := 0; i < m; i++ {
		var tan Point
		switch {
		case i == 0:
			tan = curve[1].Sub(curve[0]).Norm()
		case i == m-1:
			tan = curve[m-1].Sub(curve[m-2]).Norm()
		default:
			tan = curve[i+1].Sub(curve[i-1]).Norm()
		}
		normals[i] = tan.Perp()
	}

	baseLab := RGBAToOklab(col)
	halfW := width * 0.5

	for b := 0; b < bristles; b++ {
		// Distribute bristles across [-halfW, +halfW] with slight jitter
		u := 0.0
		if bristles > 1 {
			u = (float64(b)/float64(bristles-1))*2.0 - 1.0
		}
		baseOffset := u*halfW + rng.NormFloat64()*(width*0.04)
		driftAmp := width * (0.02 + 0.06*rng.Float64())
		driftFreq := 2.0 + rng.Float64()*4.0
		driftPhase := rng.Float64() * 2.0 * math.Pi

		// Subtle per-filament lightness/chroma variation for authentic brush texture
		dL := rng.NormFloat64() * 0.035
		filamentCol := OklabToRGBA(Oklab{
			L: common.Constrain(baseLab.L+dL, 0.02, 0.99),
			A: baseLab.A,
			B: baseLab.B,
		}, col.A)

		edgeAtten := 1.0 - 0.35*math.Abs(u)
		filamentWidth := math.Max(0.45, (width/float64(bristles))*1.65*(0.65+0.7*rng.Float64()))
		filamentAlpha := common.Constrain(alpha*edgeAtten*(0.65+0.35*rng.Float64()), 0, 1)

		filamentPts := make([]Point, m)
		for i := 0; i < m; i++ {
			t := arcLen[i] / totalLen
			// Taper lateral spread slightly at brush tip and tail
			pinch := 0.65 + 0.35*math.Sin(math.Pi*t)
			drift := driftAmp * math.Sin(2.0*math.Pi*driftFreq*t+driftPhase)
			offset := baseOffset*pinch + drift
			filamentPts[i] = curve[i].Add(normals[i].Mul(offset))
		}

		c.DrawTaperedSpline(
			filamentPts,
			filamentWidth,
			TaperCalligraphic,
			func(float64) color.RGBA { return filamentCol },
			filamentAlpha,
		)
	}
}

// subdividePolygonOnce subdivides each polygon edge at its midpoint and displaces it
// along the edge normal (and slightly along the edge) scaled by edge length and roughness.
func subdividePolygonOnce(pts []Point, edgeRoughness []float64, rng *rand.Rand) ([]Point, []float64) {
	n := len(pts)
	if n < 3 {
		return pts, edgeRoughness
	}
	outPts := make([]Point, 0, n*2)
	outRough := make([]float64, 0, n*2)

	for i := 0; i < n; i++ {
		p1 := pts[i]
		p2 := pts[(i+1)%n]
		r := edgeRoughness[i]

		mid := p1.Lerp(p2, 0.5)
		edge := p2.Sub(p1)
		edgeLen := edge.Len()
		if edgeLen > 1e-6 {
			norm := edge.Mul(1.0 / edgeLen).Perp()
			tan := edge.Mul(1.0 / edgeLen)
			dispN := rng.NormFloat64() * edgeLen * r * 0.28
			dispT := rng.NormFloat64() * edgeLen * r * 0.08
			mid = mid.Add(norm.Mul(dispN)).Add(tan.Mul(dispT))
		}

		outPts = append(outPts, p1, mid)
		// Slightly vary roughness on child edges so some edges stay crisp while others bleed more.
		r1 := math.Max(0.01, r*(0.85+0.3*rng.Float64()))
		r2 := math.Max(0.01, r*(0.85+0.3*rng.Float64()))
		outRough = append(outRough, r1, r2)
	}
	return outPts, outRough
}

// DrawWatercolorPolygon renders a Tyler Hobbs-style wet-on-wet watercolor polygon:
// recursively subdivides polygon edges with Gaussian midpoint displacement (roughness per edge),
// and composites deformed translucent layers for organic pigment bleeding and edge pooling.
func (c *Canvas) DrawWatercolorPolygon(
	pts []Point,
	col color.RGBA,
	layers int,
	subdivisions int,
	roughness float64,
	alpha float64,
	rng *rand.Rand,
) {
	if c == nil || len(pts) < 3 || layers <= 0 || alpha <= 0 || col.A == 0 {
		return
	}
	rng = ensureRNG(rng)
	if subdivisions < 1 {
		subdivisions = 1
	}
	if subdivisions > 7 {
		subdivisions = 7
	}
	if roughness < 0 {
		roughness = 0
	}

	// Assign per-edge initial roughness so certain edges bleed heavily and others stay sharper.
	basePts := make([]Point, len(pts))
	copy(basePts, pts)
	baseRough := make([]float64, len(pts))
	for i := range baseRough {
		baseRough[i] = roughness * (0.45 + 1.1*rng.Float64())
	}

	// Pre-subdivide 2 base levels so all layers share the macroscopic organic shape
	preLevels := 2
	if subdivisions < preLevels {
		preLevels = subdivisions
	}
	for s := 0; s < preLevels; s++ {
		basePts, baseRough = subdividePolygonOnce(basePts, baseRough, rng)
	}
	remLevels := subdivisions - preLevels

	baseLab := RGBAToOklab(col)
	for l := 0; l < layers; l++ {
		layerPts := make([]Point, len(basePts))
		copy(layerPts, basePts)
		layerRough := make([]float64, len(baseRough))
		for i := range baseRough {
			// Later layers have slightly tighter spread for darker edge/center pooling
			layerRough[i] = baseRough[i] * (0.65 + 0.7*rng.Float64())
		}

		// Apply a subtle layer-level jitter before fine subdivisions
		for i := range layerPts {
			prev := basePts[(i+len(basePts)-1)%len(basePts)]
			next := basePts[(i+1)%len(basePts)]
			localSpan := prev.Dist(next) * 0.5
			jit := localSpan * layerRough[i] * 0.14
			layerPts[i].X += rng.NormFloat64() * jit
			layerPts[i].Y += rng.NormFloat64() * jit
		}

		for s := 0; s < remLevels; s++ {
			layerPts, layerRough = subdividePolygonOnce(layerPts, layerRough, rng)
		}

		// Subtle warm/cool and lightness shift per wash layer
		dL := rng.NormFloat64() * 0.015
		layerCol := OklabToRGBA(Oklab{
			L: common.Constrain(baseLab.L+dL, 0.01, 0.99),
			A: baseLab.A,
			B: baseLab.B,
		}, col.A)

		c.FillPolygon(layerPts, layerCol, alpha)
	}
}
