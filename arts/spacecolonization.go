package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type spaceColonization struct {
	attractorNum  int
	maxSteps      int
	killDist      float64
	influenceDist float64
}

type scNode struct {
	pos      render.Point
	parent   int
	children []int
	depth    int
	weight   float64
	rootID   int
}

type scAttractor struct {
	pos    render.Point
	active bool
}

// NewSpaceColonization returns a spaceColonization engine that grows organic vascular
// and botanical branching networks using Runions' Space Colonization algorithm and Murray's law.
func NewSpaceColonization(attractorNum, maxSteps int, killDist, influenceDist float64) *spaceColonization {
	if attractorNum <= 0 {
		attractorNum = 1800
	}
	if maxSteps <= 0 {
		maxSteps = 160
	}
	if killDist <= 0 {
		killDist = 6.0
	}
	if influenceDist <= 0 {
		influenceDist = 55.0
	}
	return &spaceColonization{
		attractorNum:  attractorNum,
		maxSteps:      maxSteps,
		killDist:      killDist,
		influenceDist: influenceDist,
	}
}

// Generative runs the Space Colonization vascular growth simulation and renders a
// bioluminescent botanical/neural canopy with pipe-model tapering and glowing synaptic tips.
func (sc *spaceColonization) Generative(c *generativeart.Canva) {
	attractorNum := sc.attractorNum
	if attractorNum <= 0 {
		attractorNum = 1800
	}
	maxSteps := sc.maxSteps
	if maxSteps <= 0 {
		maxSteps = 160
	}
	baseKill := sc.killDist
	if baseKill <= 0 {
		baseKill = 6.0
	}
	baseInfluence := sc.influenceDist
	if baseInfluence <= 0 {
		baseInfluence = 55.0
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

	killDist := baseKill * scale
	influenceDist := baseInfluence * scale
	stepDist := 4.2 * scale
	if influenceDist <= killDist*1.5 {
		influenceDist = killDist * 3.0
	}

	seed := rand.Int63()
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	cx, cy := fw*0.5, fh*0.44
	canopyRx := fw * 0.41
	canopyRy := fh * 0.36

	// 1. Scatter auxin attractors inside a rich dual-lobe botanical corona with spiral arm modulation,
	// plus a vascular trunk corridor from the base root cluster up into the canopy heart.
	attractors := make([]scAttractor, 0, attractorNum)
	trunkCount := common.MaxInt(90, attractorNum/14)
	rootBaseY := fh * 0.87

	for i := 0; i < trunkCount; i++ {
		u := float64(i) / float64(common.MaxInt(trunkCount-1, 1))
		ty := common.Lerp(rootBaseY-8.0*scale, cy+canopyRy*0.18, u)
		// Gently braided triple-stem corridor that widens as it enters the lower corona
		spread := (10.0 + 48.0*math.Pow(u, 1.35)) * scale
		swirl := math.Sin(u*math.Pi*2.6+float64(i%3)*2.094) * spread * 0.55
		tx := cx + swirl + rng.NormFloat64()*spread*0.32
		attractors = append(attractors, scAttractor{
			pos:    render.Pt(common.Constrain(tx, 20*scale, fw-20*scale), common.Constrain(ty, 20*scale, fh-20*scale)),
			active: true,
		})
	}

	// Dual-lobe + spiral-arm botanical corona attractors
	attempts := 0
	maxAttempts := attractorNum * 12
	for len(attractors) < attractorNum && attempts < maxAttempts {
		attempts++
		// Sample in normalized [-1, 1]^2 and shape into a dual-lobe heart/corona canopy
		ang := rng.Float64() * 2.0 * math.Pi
		rad := math.Sqrt(rng.Float64())

		// Dual-lobe botanical envelope + 5-fold floral/spiral harmonic modulation
		lobeEnvelope := 0.82 + 0.22*math.Cos(2.0*(ang-math.Pi*0.5)) - 0.10*math.Sin(ang)
		floralLobes := 1.0 + 0.14*math.Cos(5.0*ang+rad*2.8)
		rEff := rad * lobeEnvelope * floralLobes

		px := cx + rEff*canopyRx*math.Cos(ang)
		py := cy + rEff*canopyRy*math.Sin(ang)

		// Spiral arm + FBM density field for organic clustering and negative-space lacunae
		spiralPhase := 3.0*ang + 4.8*rad + 0.65*noise.FBM2D(px*0.008/scale, py*0.008/scale, 3, 2.0, 0.5)
		spiralDensity := 0.55 + 0.45*math.Cos(spiralPhase)
		fbmDensity := 0.5 + 0.5*noise.FBM2D(px*0.012/scale+3.7, py*0.012/scale+7.1, 3, 2.0, 0.5)
		acceptProb := (0.35 + 0.65*spiralDensity) * (0.45 + 0.55*fbmDensity)

		if rng.Float64() <= acceptProb {
			margin := 22.0 * scale
			if px >= margin && px <= fw-margin && py >= margin && py <= fh-margin {
				attractors = append(attractors, scAttractor{
					pos:    render.Pt(px, py),
					active: true,
				})
			}
		}
	}

	// Fill any remaining slots if rejection sampling stopped early
	for len(attractors) < attractorNum {
		ang := rng.Float64() * 2.0 * math.Pi
		rad := math.Sqrt(rng.Float64()) * 0.85
		px := cx + rad*canopyRx*math.Cos(ang)
		py := cy + rad*canopyRy*math.Sin(ang)
		attractors = append(attractors, scAttractor{
			pos:    render.Pt(px, py),
			active: true,
		})
	}

	// 2. Seed 3 root nodes at the base/center
	nodes := make([]scNode, 0, 4096)
	rootOffsets := [3]render.Point{
		{X: cx, Y: rootBaseY},
		{X: cx - 18.0*scale, Y: rootBaseY + 10.0*scale},
		{X: cx + 18.0*scale, Y: rootBaseY + 10.0*scale},
	}
	for rIdx, rPos := range rootOffsets {
		nodes = append(nodes, scNode{
			pos:    rPos,
			parent: -1,
			depth:  0,
			weight: 1.0,
			rootID: rIdx,
		})
	}

	// 3. Run Space Colonization iterative growth
	infDistSq := influenceDist * influenceDist
	killDistSq := killDist * killDist
	minChildSepSq := (stepDist * 0.55) * (stepDist * 0.55)

	type accumDir struct {
		dir   render.Point
		count int
	}

	for step := 0; step < maxSteps; step++ {
		nNodes := len(nodes)
		acc := make([]accumDir, nNodes)
		anyAssociated := false

		for i := range attractors {
			if !attractors[i].active {
				continue
			}
			ap := attractors[i].pos
			bestNode := -1
			bestDistSq := infDistSq

			for j := 0; j < nNodes; j++ {
				dx := ap.X - nodes[j].X()
				dy := ap.Y - nodes[j].Y()
				d2 := dx*dx + dy*dy
				if d2 < bestDistSq {
					bestDistSq = d2
					bestNode = j
				}
			}

			if bestNode >= 0 {
				dir := ap.Sub(nodes[bestNode].pos).Norm()
				acc[bestNode].dir = acc[bestNode].dir.Add(dir)
				acc[bestNode].count++
				anyAssociated = true
			}
		}

		if !anyAssociated {
			break
		}

		spawnedThisStep := 0
		for j := 0; j < nNodes; j++ {
			if acc[j].count == 0 {
				continue
			}
			avgDir := acc[j].dir.Norm()
			if avgDir.Len() < 1e-8 {
				continue
			}

			// Add subtle curl noise perturbation for natural botanical/synaptic curvature
			p := nodes[j].pos
			curl := noise.Curl2D(p.X*0.009/scale+float64(step)*0.01, p.Y*0.009/scale, 0.05)
			perturbed := render.Pt(
				avgDir.X+curl.X*0.18,
				avgDir.Y+curl.Y*0.18,
			).Norm()
			if perturbed.Len() < 1e-8 {
				perturbed = avgDir
			}

			childPos := p.Add(perturbed.Mul(stepDist))

			// Prevent degenerate overlapping children on the same parent
			tooClose := false
			for _, chIdx := range nodes[j].children {
				dx := childPos.X - nodes[chIdx].pos.X
				dy := childPos.Y - nodes[chIdx].pos.Y
				if dx*dx+dy*dy < minChildSepSq {
					tooClose = true
					break
				}
			}
			if tooClose {
				continue
			}

			newIdx := len(nodes)
			nodes[j].children = append(nodes[j].children, newIdx)
			nodes = append(nodes, scNode{
				pos:    childPos,
				parent: j,
				depth:  nodes[j].depth + 1,
				weight: 1.0,
				rootID: nodes[j].rootID,
			})
			spawnedThisStep++
		}

		if spawnedThisStep == 0 {
			break
		}

		// Remove attractors within killDist of any newly spawned or existing tree node
		activeRemaining := 0
		for i := range attractors {
			if !attractors[i].active {
				continue
			}
			ap := attractors[i].pos
			killed := false
			// Check newest nodes first since they are most likely to kill attractors
			for j := len(nodes) - 1; j >= 0; j-- {
				dx := ap.X - nodes[j].pos.X
				dy := ap.Y - nodes[j].pos.Y
				if dx*dx+dy*dy <= killDistSq {
					killed = true
					break
				}
			}
			if killed {
				attractors[i].active = false
			} else {
				activeRemaining++
			}
		}

		if activeRemaining == 0 {
			break
		}
	}

	// 4. Compute exact pipe-model vascular thickness (Murray's law: parentWeight = sum(childWeight^2.4)^(1/2.4))
	// Since children always have an index strictly greater than their parent, reverse iteration is an exact post-order traversal.
	const murrayExp = 2.4
	const invMurrayExp = 1.0 / murrayExp
	maxDepth := 1
	for i := len(nodes) - 1; i >= 0; i-- {
		if nodes[i].depth > maxDepth {
			maxDepth = nodes[i].depth
		}
		if len(nodes[i].children) == 0 {
			nodes[i].weight = 1.0
			continue
		}
		sumPow := 0.0
		for _, ch := range nodes[i].children {
			sumPow += math.Pow(nodes[ch].weight, murrayExp)
		}
		nodes[i].weight = math.Pow(sumPow, invMurrayExp)
	}

	maxWeight := 1.0
	for i := range nodes {
		if nodes[i].weight > maxWeight {
			maxWeight = nodes[i].weight
		}
	}

	nodeThickness := func(idx int) float64 {
		wNorm := nodes[idx].weight / math.Max(1.0, maxWeight)
		return (0.65 + 9.2*math.Pow(wNorm, 0.78)) * scale
	}

	// 5. Set up palette & HDR render canvas
	schema := c.Opts().ColorSchema()
	if len(schema) < 2 {
		schema = common.Botanical
	}
	grad := render.NewMultiStopGradient(schema...)
	cyberGrad := render.NewMultiStopGradient(common.Cyberpunk...)

	rc := c.NewRenderCanvas()

	// Deep bioluminescent abyssal backdrop with subtle vertical & radial gradient
	bgTop := color.RGBA{R: 5, G: 9, B: 18, A: 255}
	bgBottom := color.RGBA{R: 10, G: 18, B: 28, A: 255}
	bgAccent := render.LerpOklab(bgBottom, grad.SampleOklab(0.15), 0.14)

	for y := 0; y < h; y++ {
		ty := float64(y) / math.Max(1.0, fh-1.0)
		rowCol := render.LerpOklab(bgTop, bgAccent, math.Pow(ty, 0.9))
		rLin, gLin, bLin, _ := render.ColorToLinearRGB(rowCol)
		for x := 0; x < w; x++ {
			rc.SetPixelLinear(x, y, rLin, gLin, bLin, 1.0)
		}
	}

	// Soft nebula glow behind the canopy
	rc.SetBlendMode(render.BlendAdd)
	rc.DrawGlowDot(cx, cy, math.Min(fw, fh)*0.48, grad.SampleOklab(0.22), 0.18, 0.0)
	rc.DrawGlowDot(cx-canopyRx*0.34, cy-canopyRy*0.12, math.Min(fw, fh)*0.30, grad.SampleOklab(0.48), 0.15, 0.0)
	rc.DrawGlowDot(cx+canopyRx*0.34, cy-canopyRy*0.12, math.Min(fw, fh)*0.30, cyberGrad.SampleOklab(0.72), 0.14, 0.0)
	rc.DrawGlowDot(cx, rootBaseY-35.0*scale, math.Min(fw, fh)*0.24, grad.SampleOklab(0.12), 0.16, 0.0)

	// Fine atmospheric bioluminescent spore dust in the background
	spores := 380
	for i := 0; i < spores; i++ {
		ang := rng.Float64() * 2.0 * math.Pi
		rad := math.Sqrt(rng.Float64()) * 1.08
		sx := cx + rad*canopyRx*math.Cos(ang)
		sy := cy + rad*canopyRy*math.Sin(ang)
		u := common.Clamp01(0.2 + 0.75*rad)
		sporeCol := render.LerpOklab(grad.SampleOklab(u), cyberGrad.SampleOklab(0.8), 0.35)
		if i%9 == 0 {
			rc.DrawGlowDot(sx, sy, (1.2+rng.Float64()*2.4)*scale, sporeCol, 0.18+0.20*rng.Float64(), 0.2)
		} else {
			rc.SplatPoint(sx, sy, sporeCol, 0.16+0.25*rng.Float64())
		}
	}

	// 6. Extract smooth branch chains from roots to tips using heavy-child continuation
	// so main vascular trunks form long continuous splines and lateral branches fork smoothly.
	visited := make([]bool, len(nodes))
	var chains [][]int

	var extractChains func(startIdx int)
	extractChains = func(startIdx int) {
		chain := make([]int, 0, 32)
		if nodes[startIdx].parent >= 0 {
			chain = append(chain, nodes[startIdx].parent)
		}
		curr := startIdx
		for curr >= 0 && !visited[curr] {
			visited[curr] = true
			chain = append(chain, curr)
			if len(nodes[curr].children) == 0 {
				break
			}
			// Pick heaviest child to continue the primary spline chain
			heavyChild := nodes[curr].children[0]
			bestW := nodes[heavyChild].weight
			for _, ch := range nodes[curr].children[1:] {
				if nodes[ch].weight > bestW {
					bestW = nodes[ch].weight
					heavyChild = ch
				}
			}
			// Recursively extract side branches after building this chain
			for _, ch := range nodes[curr].children {
				if ch != heavyChild {
					defer extractChains(ch)
				}
			}
			curr = heavyChild
		}
		if len(chain) >= 2 {
			chains = append(chains, chain)
		}
	}

	for i := 0; i < len(nodes); i++ {
		if nodes[i].parent == -1 && !visited[i] {
			extractChains(i)
		}
	}

	// Render chains from thickest trunks to finest capillaries
	fMaxDepth := float64(common.MaxInt(maxDepth, 1))
	for _, chain := range chains {
		pts := make([]render.Point, len(chain))
		weights := make([]float64, len(chain))
		depths := make([]float64, len(chain))
		chainMaxW := 0.1

		for i, idx := range chain {
			pts[i] = nodes[idx].pos
			wVal := nodeThickness(idx)
			weights[i] = wVal
			if wVal > chainMaxW {
				chainMaxW = wVal
			}
			depths[i] = common.Clamp01(float64(nodes[idx].depth) / fMaxDepth)
		}

		nChain := len(chain)
		widthProf := func(t float64) float64 {
			if nChain < 2 || chainMaxW <= 1e-6 {
				return 1.0
			}
			pos := common.Clamp01(t) * float64(nChain-1)
			i0 := int(math.Floor(pos))
			if i0 >= nChain-1 {
				return weights[nChain-1] / chainMaxW
			}
			frac := pos - float64(i0)
			wLerp := weights[i0] + (weights[i0+1]-weights[i0])*frac
			return common.Clamp01(wLerp / chainMaxW)
		}

		colorProf := func(t float64) color.RGBA {
			pos := common.Clamp01(t) * float64(nChain-1)
			i0 := int(math.Floor(pos))
			if i0 >= nChain-1 {
				i0 = nChain - 1
			}
			dNorm := depths[i0]
			base := grad.SampleOklab(common.Clamp01(0.08 + 0.88*math.Pow(dNorm, 0.85)))
			luminous := cyberGrad.SampleOklab(common.Clamp01(0.45 + 0.50*dNorm))
			return render.LerpOklab(base, luminous, 0.28*dNorm)
		}

		// Soft outer bioluminescent vascular aura
		rc.SetBlendMode(render.BlendAdd)
		midCol := colorProf(0.5)
		if len(pts) >= 3 {
			smoothCurve := render.CatmullRomSpline(pts, 10)
			rc.StrokePolyline(smoothCurve, chainMaxW*1.45, midCol, 0.14)
		} else {
			rc.StrokePolyline(pts, chainMaxW*1.35, midCol, 0.14)
		}

		// Core Murray's law tapered vascular branch
		rc.SetBlendMode(render.BlendOver)
		rc.DrawTaperedSpline(pts, chainMaxW, widthProf, colorProf, 0.90)

		// Bright inner filament highlight
		rc.SetBlendMode(render.BlendAdd)
		rc.DrawTaperedSpline(pts, math.Max(0.55*scale, chainMaxW*0.38), widthProf, func(t float64) color.RGBA {
			c0 := colorProf(t)
			return render.LerpOklab(c0, color.RGBA{R: 245, G: 255, B: 245, A: 255}, 0.42)
		}, 0.36)
	}

	// 7. At terminal leaf/synapse tips, render glowing blossom/synaptic nodes using DrawGlowDot and DrawSoftRing in BlendAdd mode
	rc.SetBlendMode(render.BlendAdd)
	whiteCore := color.RGBA{R: 252, G: 255, B: 248, A: 255}
	for i := range nodes {
		if len(nodes[i].children) != 0 {
			continue
		}
		p := nodes[i].pos
		dNorm := common.Clamp01(float64(nodes[i].depth) / fMaxDepth)
		tipCol := grad.SampleOklab(common.Clamp01(0.62 + 0.36*dNorm))
		accentCol := cyberGrad.SampleOklab(common.Clamp01(0.65 + 0.30*math.Sin(float64(i)*0.7)))
		blossomCol := render.LerpOklab(tipCol, accentCol, 0.38)

		rad := (2.4 + 2.8*rng.Float64()) * scale
		rc.DrawGlowDot(p.X, p.Y, rad*1.85, blossomCol, 0.42, 0.15)
		rc.DrawGlowDot(p.X, p.Y, rad*0.72, whiteCore, 0.78, 0.65)

		if i%3 == 0 {
			ringRad := rad * (1.55 + 0.95*rng.Float64())
			rc.DrawSoftRing(p.X, p.Y, ringRad, 0.95*scale, blossomCol, 0.44)
		}
	}

	// 8. Post-processing & ACES filmic tone mapping
	rc.ApplyBloom(0.45, 9.0, 0.6)
	rc.ApplyVignette(0.3)
	c.MergeRenderCanvasToneMapped(rc, 1.12, 0)
}

func (n scNode) X() float64 { return n.pos.X }
func (n scNode) Y() float64 { return n.pos.Y }
