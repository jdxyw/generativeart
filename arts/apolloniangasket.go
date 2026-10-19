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

type apollonianGasket struct {
	maxDepth  int
	minRadius float64
}

// NewApollonianGasket returns an apollonianGasket engine that recursively solves the
// exact Complex Descartes Circle Theorem to generate a fractal celestial astrolabe gasket.
func NewApollonianGasket(maxDepth int, minRadius float64) *apollonianGasket {
	if maxDepth <= 0 {
		maxDepth = 7
	}
	if minRadius <= 0 {
		minRadius = 2.5
	}
	return &apollonianGasket{
		maxDepth:  maxDepth,
		minRadius: minRadius,
	}
}

type descartesCircle struct {
	z     complex128 // center x + i*y relative to gasket center
	k     float64    // signed curvature (k < 0 for enclosing circle, k > 0 for interior circles)
	r     float64    // radius = 1 / |k|
	depth int
}

// tangencyError measures how closely candidate circle c4 is mutually tangent to c1, c2, c3.
func tangencyError(c1, c2, c3, c4 descartesCircle) float64 {
	checkPair := func(a, b descartesCircle) float64 {
		d := cmplx.Abs(a.z - b.z)
		if a.k < 0 {
			// b is internally tangent inside enclosing circle a
			return math.Abs(d - (a.r - b.r))
		}
		if b.k < 0 {
			return math.Abs(d - (b.r - a.r))
		}
		// Both positive curvature: externally tangent
		return math.Abs(d - (a.r + b.r))
	}
	return checkPair(c1, c4) + checkPair(c2, c4) + checkPair(c3, c4)
}

// solveDescartes computes the two candidate fourth circles tangent to (c1, c2, c3)
// using both the curvature Descartes theorem and the Complex Descartes theorem:
//
//	k4 = k1 + k2 + k3 +/- 2*sqrt(k1*k2 + k2*k3 + k3*k1)
//	z4 = (z1*k1 + z2*k2 + z3*k3 +/- 2*sqrt(k1*k2*z1*z2 + k2*k3*z2*z3 + k3*k1*z3*z1)) / k4
func solveDescartes(c1, c2, c3 descartesCircle, nextDepth int) []descartesCircle {
	k1, k2, k3 := c1.k, c2.k, c3.k
	kSum := k1 + k2 + k3
	discK := k1*k2 + k2*k3 + k3*k1
	if discK < 0 {
		discK = 0
	}
	twoSqrtK := 2.0 * math.Sqrt(discK)
	kCandidates := [2]float64{kSum + twoSqrtK, kSum - twoSqrtK}

	kz1 := complex(k1, 0) * c1.z
	kz2 := complex(k2, 0) * c2.z
	kz3 := complex(k3, 0) * c3.z
	kzSum := kz1 + kz2 + kz3

	discZ := kz1*kz2 + kz2*kz3 + kz3*kz1
	twoSqrtZ := 2.0 * cmplx.Sqrt(discZ)

	var out []descartesCircle
	for _, k4 := range kCandidates {
		if k4 <= 1e-6 {
			continue
		}
		r4 := 1.0 / k4
		candPlus := descartesCircle{
			z:     (kzSum + twoSqrtZ) / complex(k4, 0),
			k:     k4,
			r:     r4,
			depth: nextDepth,
		}
		candMinus := descartesCircle{
			z:     (kzSum - twoSqrtZ) / complex(k4, 0),
			k:     k4,
			r:     r4,
			depth: nextDepth,
		}
		errPlus := tangencyError(c1, c2, c3, candPlus)
		errMinus := tangencyError(c1, c2, c3, candMinus)
		best := candPlus
		bestErr := errPlus
		if errMinus < errPlus {
			best = candMinus
			bestErr = errMinus
		}
		// Accept if geometric tangency holds within tolerance
		if bestErr <= math.Max(1.5, r4*0.08) {
			out = append(out, best)
		}
	}
	return out
}

// buildApollonianGasket generates a deduplicated list of circles inside bounding circle R0.
func buildApollonianGasket(r0, minRadius float64, maxDepth int, rotAngle float64) []descartesCircle {
	// Outer bounding circle C0 has negative curvature k0 = -1 / R0.
	c0 := descartesCircle{
		z:     0,
		k:     -1.0 / r0,
		r:     r0,
		depth: 0,
	}

	// Three mutually tangent congruent circles inscribed inside C0:
	// r = R0 / (1 + 2/sqrt(3)), centers at distance (R0 - r) from origin.
	rInit := r0 / (1.0 + 2.0/math.Sqrt(3.0))
	dInit := r0 - rInit
	var initThree [3]descartesCircle
	for i := 0; i < 3; i++ {
		ang := rotAngle + float64(i)*(2.0*math.Pi/3.0)
		initThree[i] = descartesCircle{
			z:     complex(dInit*math.Cos(ang), dInit*math.Sin(ang)),
			k:     1.0 / rInit,
			r:     rInit,
			depth: 1,
		}
	}

	// Central interstitial circle C4 tangent to all three initThree circles:
	rCenter := dInit - rInit
	cCenter := descartesCircle{
		z:     0,
		k:     1.0 / rCenter,
		r:     rCenter,
		depth: 1,
	}

	var circles []descartesCircle
	seen := make(map[[3]int64]bool)

	quantKey := func(c descartesCircle) [3]int64 {
		return [3]int64{
			int64(math.Round(real(c.z) * 20.0)),
			int64(math.Round(imag(c.z) * 20.0)),
			int64(math.Round(c.r * 20.0)),
		}
	}

	isDuplicate := func(c descartesCircle) bool {
		k := quantKey(c)
		for dx := int64(-1); dx <= 1; dx++ {
			for dy := int64(-1); dy <= 1; dy++ {
				for dr := int64(-1); dr <= 1; dr++ {
					if seen[[3]int64{k[0] + dx, k[1] + dy, k[2] + dr}] {
						return true
					}
				}
			}
		}
		return false
	}

	addCircle := func(c descartesCircle) bool {
		if c.r < minRadius || math.IsNaN(real(c.z)) || math.IsNaN(imag(c.z)) || math.IsNaN(c.r) {
			return false
		}
		// Must stay inside outer circle C0 (with small numerical tolerance)
		if cmplx.Abs(c.z)+c.r > r0+1.5 {
			return false
		}
		if isDuplicate(c) {
			return false
		}
		seen[quantKey(c)] = true
		circles = append(circles, c)
		return true
	}

	addCircle(initThree[0])
	addCircle(initThree[1])
	addCircle(initThree[2])
	addCircle(cCenter)

	var recurse func(a, b, c descartesCircle, depth int)
	recurse = func(a, b, c descartesCircle, depth int) {
		if depth > maxDepth {
			return
		}
		cands := solveDescartes(a, b, c, depth)
		for _, c4 := range cands {
			if addCircle(c4) && depth < maxDepth && c4.r >= minRadius*1.05 {
				recurse(a, b, c4, depth+1)
				recurse(b, c, c4, depth+1)
				recurse(c, a, c4, depth+1)
			}
		}
	}

	// Recurse on the 4 mutually tangent triples formed by (C0, C1, C2, C3, C_center)
	recurse(c0, initThree[0], initThree[1], 2)
	recurse(c0, initThree[1], initThree[2], 2)
	recurse(c0, initThree[2], initThree[0], 2)
	recurse(initThree[0], initThree[1], cCenter, 2)
	recurse(initThree[1], initThree[2], cCenter, 2)
	recurse(initThree[2], initThree[0], cCenter, 2)

	return circles
}

// circlePath constructs a closed circular tdewolff/canvas Path centered at (cx, cy) with radius r.
func circlePath(cx, cy, r float64) *canvas.Path {
	const segs = 72
	p := &canvas.Path{}
	for i := 0; i <= segs; i++ {
		ang := (float64(i) / float64(segs)) * 2.0 * math.Pi
		px := cx + r*math.Cos(ang)
		py := cy + r*math.Sin(ang)
		if i == 0 {
			p.MoveTo(px, py)
		} else {
			p.LineTo(px, py)
		}
	}
	p.Close()
	return p
}

// Generative renders the Apollonian Gasket celestial astrolabe into c.
func (ag *apollonianGasket) Generative(c *generativeart.Canva) {
	maxDepth := ag.maxDepth
	minRadius := ag.minRadius
	if maxDepth <= 0 {
		maxDepth = 7
	}
	if minRadius <= 0 {
		minRadius = 2.5
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
	r0 := 0.44 * minDim
	scaledMinR := math.Max(1.5, minRadius*scale)

	schema := c.Opts().ColorSchema()
	if len(schema) == 0 {
		schema = common.Dunhuang
	}
	grad := render.NewMultiStopGradient(schema...)
	sunsetGrad := render.NewMultiStopGradient(common.Sunset...)

	rc := c.NewRenderCanvas()

	// 1. Deep celestial astrolabe background with radial glow and concentric degree ticks.
	bgCol := color.RGBA{R: 7, G: 10, B: 22, A: 255}
	rc.FillBackground(bgCol)

	rc.SetBlendMode(render.BlendAdd)
	rc.DrawGlowDot(cx, cy, r0*1.12, grad.SampleOklab(0.25), 0.18, 0.0)
	rc.DrawGlowDot(cx, cy, r0*0.65, sunsetGrad.SampleOklab(0.65), 0.12, 0.0)

	goldRim := color.RGBA{R: 245, G: 205, B: 125, A: 255}
	celestialCyan := render.LerpOklab(color.RGBA{R: 70, G: 165, B: 195, A: 255}, grad.SampleOklab(0.75), 0.5)

	// Outer astrolabe bezel rings and 72 degree ticks
	bezelR1 := r0 * 1.025
	bezelR2 := r0 * 1.065
	rc.DrawSoftRing(cx, cy, r0, 2.2*scale, goldRim, 0.72)
	rc.DrawSoftRing(cx, cy, bezelR1, 1.1*scale, goldRim, 0.42)
	rc.DrawSoftRing(cx, cy, bezelR2, 1.4*scale, celestialCyan, 0.35)

	tickPath := &canvas.Path{}
	for i := 0; i < 144; i++ {
		ang := (float64(i) / 144.0) * 2.0 * math.Pi
		rInner := bezelR1
		rOuter := bezelR1 + (bezelR2-bezelR1)*0.55
		if i%6 == 0 {
			rInner = r0
			rOuter = bezelR2
		} else if i%2 == 0 {
			rOuter = bezelR1 + (bezelR2-bezelR1)*0.82
		}
		tickPath.MoveTo(cx+rInner*math.Cos(ang), cy+rInner*math.Sin(ang))
		tickPath.LineTo(cx+rOuter*math.Cos(ang), cy+rOuter*math.Sin(ang))
	}
	rc.StrokePath(tickPath, 0.85*scale, canvas.RoundCap, canvas.RoundJoin, goldRim, 0.35)

	// 2. Solve exact Descartes Circle Theorem for the primary Apollonian Gasket,
	// plus recursively nested sub-gaskets inside the largest circles.
	primary := buildApollonianGasket(r0, scaledMinR, maxDepth, math.Pi/6.0)
	allCircles := make([]descartesCircle, 0, len(primary)*2)
	allCircles = append(allCircles, primary...)

	for _, pc := range primary {
		if pc.depth == 1 && pc.r > 42.0*scale {
			sub := buildApollonianGasket(pc.r*0.985, scaledMinR*1.1, common.MaxInt(maxDepth-2, 4), -math.Pi/6.0+float64(pc.depth)*0.4)
			for _, sc := range sub {
				sc.z += pc.z
				sc.depth += pc.depth
				allCircles = append(allCircles, sc)
			}
		}
	}

	// Sort circles from largest radius to smallest so fine nested circles layer cleanly.
	sort.Slice(allCircles, func(i, j int) bool {
		return allCircles[i].r > allCircles[j].r
	})

	logMin := math.Log(math.Max(1.0, scaledMinR))
	logMax := math.Log(math.Max(2.0, r0*0.5))

	// 3. Render each packed circle with Oklab gradient glow, luminous rim, and astrolabe rosette geometry.
	for _, cir := range allCircles {
		px := cx + real(cir.z)
		py := cy + imag(cir.z)
		r := cir.r

		polarAng := math.Atan2(imag(cir.z), real(cir.z))
		angNorm := (polarAng + math.Pi) / (2.0 * math.Pi)
		logNorm := common.Clamp01((math.Log(math.Max(1.0, r)) - logMin) / math.Max(1e-6, logMax-logMin))
		depthNorm := common.Clamp01(float64(cir.depth-1) / float64(common.MaxInt(maxDepth, 1)))

		tCol := common.Clamp01(0.48*(1.0-logNorm) + 0.32*depthNorm + 0.20*math.Sin(angNorm*2.0*math.Pi)*0.5 + 0.10)
		baseCol := grad.SampleOklab(tCol)
		accentCol := sunsetGrad.SampleOklab(common.Clamp01(1.0 - tCol))
		rimCol := render.LerpOklab(baseCol, goldRim, 0.42+0.25*(1.0-logNorm))

		// Soft inner glow (BlendAdd)
		rc.SetBlendMode(render.BlendAdd)
		glowAlpha := common.Clamp01(0.08 + 0.16*(1.0-logNorm))
		rc.DrawGlowDot(px, py, r*0.92, baseCol, glowAlpha, 0.18)

		// Crisp outer luminous rim
		rimThick := common.Constrain(r*0.045, 0.75*scale, 2.2*scale)
		rimAlpha := common.Clamp01(0.42 + 0.38*logNorm)
		if r > 12.0*scale {
			rc.DrawSoftRing(px, py, r, rimThick*1.25, rimCol, rimAlpha*0.65)
			cPath := circlePath(px, py, r)
			rc.StrokePath(cPath, rimThick*0.65, canvas.RoundCap, canvas.RoundJoin, goldRim, rimAlpha*0.75)
		} else {
			rc.DrawSoftRing(px, py, r, rimThick, rimCol, rimAlpha)
		}

		// Intricate nested astrolabe / rosette chord geometry for larger circles (r > 18)
		if r > 18.0*scale {
			rosettePath := &canvas.Path{}
			nPetals := 8
			if r > 45.0*scale {
				nPetals = 12
			}
			steps := 96
			rInner := r * 0.86
			for s := 0; s <= steps; s++ {
				theta := (float64(s) / float64(steps)) * 2.0 * math.Pi
				// Hypotrochoid / astrolabe star chord modulation inside the circle
				mod := 0.58 + 0.42*math.Cos(float64(nPetals)*theta*0.5)
				rx := px + rInner*mod*math.Cos(theta+polarAng)
				ry := py + rInner*mod*math.Sin(theta+polarAng)
				if s == 0 {
					rosettePath.MoveTo(rx, ry)
				} else {
					rosettePath.LineTo(rx, ry)
				}
			}
			rosettePath.Close()

			// Inscribed polygonal chords connecting cardinal & ordinal astrolabe points
			for pIdx := 0; pIdx < nPetals; pIdx++ {
				a1 := polarAng + float64(pIdx)*(2.0*math.Pi/float64(nPetals))
				a2 := polarAng + float64((pIdx+3)%nPetals)*(2.0*math.Pi/float64(nPetals))
				rosettePath.MoveTo(px+rInner*math.Cos(a1), py+rInner*math.Sin(a1))
				rosettePath.LineTo(px+rInner*math.Cos(a2), py+rInner*math.Sin(a2))
			}

			rc.StrokePath(rosettePath, 0.72*scale, canvas.RoundCap, canvas.RoundJoin, accentCol, 0.28)
			rc.DrawSoftRing(px, py, r*0.52, 0.9*scale, goldRim, 0.24)
			rc.DrawGlowDot(px, py, math.Max(1.8*scale, r*0.08), goldRim, 0.48, 0.4)
		}
	}

	rc.ApplyBloom(0.5, 8.0, 0.45)
	rc.ApplyVignette(0.3)
	c.MergeRenderCanvasToneMapped(rc, 1.1, 0)
}
