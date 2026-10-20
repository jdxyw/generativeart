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

type dlaFrost struct {
	particlesN int
	anisotropy float64
}

type dlaNode struct {
	x, y          int
	fx, fy        float64
	parent        int
	depth         int
	arrivalT      float64
	subtreeWeight float64
	hasChild      bool
}

// NewDLAFrost returns a dlaFrost engine that simulates fast Diffusion-Limited Aggregation (DLA)
// with 6-fold hexagonal crystal anisotropy, Strahler/subtree pipe weighting, and Oklab HDR rendering.
func NewDLAFrost(particlesN int, anisotropy float64) *dlaFrost {
	if particlesN <= 0 {
		particlesN = 18000
	}
	if anisotropy <= 0 {
		anisotropy = 0.55
	}
	return &dlaFrost{
		particlesN: particlesN,
		anisotropy: anisotropy,
	}
}

// Generative runs the 2D occupancy-grid DLA simulation and renders a 6-fold dendritic frost crystal.
func (df *dlaFrost) Generative(c *generativeart.Canva) {
	particlesN := df.particlesN
	if particlesN <= 0 {
		particlesN = 18000
	}
	particlesN = common.ConstrainInt(particlesN, 500, 45000)

	aniso := df.anisotropy
	if aniso <= 0 {
		aniso = 0.55
	}
	aniso = common.Constrain(aniso, 0.05, 0.95)

	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	cx, cy := w/2, h/2
	fcx, fcy := float64(cx), float64(cy)
	maxClusterR := math.Min(fw, fh)*0.45 - 4.0
	if maxClusterR < 16.0 {
		maxClusterR = math.Min(fw, fh) * 0.45
	}

	seed := rand.Int63()
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	// 1. 2D occupancy grid (resolution w x h) storing nodeIndex + 1 (0 if empty)
	// plus an 8x8 block proximity acceleration mask so far walkers jump fast.
	occGrid := make([]int32, w*h)
	tileW := (w + 7) >> 3
	tileH := (h + 7) >> 3
	nearCrystalTile := make([]bool, tileW*tileH)

	nodes := make([]dlaNode, 0, particlesN+32)
	rMax := 3.0

	addCrystalNode := func(x, y, parentIdx, depth int, arrivalT float64) int {
		if x < 1 || x >= w-1 || y < 1 || y >= h-1 {
			return -1
		}
		if occGrid[y*w+x] != 0 {
			return -1
		}
		idx := len(nodes)
		fx := float64(x)
		fy := float64(y)
		if parentIdx >= 0 {
			p := nodes[parentIdx]
			// Subtle sub-pixel smoothing toward the parent-child hexagonal ray
			fx = float64(x) + (rng.Float64()-0.5)*0.28
			fy = float64(y) + (rng.Float64()-0.5)*0.28
			_ = p
			nodes[parentIdx].hasChild = true
		}
		nodes = append(nodes, dlaNode{
			x:             x,
			y:             y,
			fx:            fx,
			fy:            fy,
			parent:        parentIdx,
			depth:         depth,
			arrivalT:      arrivalT,
			subtreeWeight: 1.0,
		})
		occGrid[y*w+x] = int32(idx + 1)

		// Mark 3x3 neighborhood in the 8x8 tile acceleration grid
		tx, ty := x>>3, y>>3
		for dy := -1; dy <= 1; dy++ {
			nty := ty + dy
			if nty < 0 || nty >= tileH {
				continue
			}
			for dx := -1; dx <= 1; dx++ {
				ntx := tx + dx
				if ntx >= 0 && ntx < tileW {
					nearCrystalTile[nty*tileW+ntx] = true
				}
			}
		}

		dist := math.Hypot(float64(x)-fcx, float64(y)-fcy)
		if dist > rMax {
			rMax = dist
		}
		return idx
	}

	// 2. Seed the center and a 6-point hexagonal crystal nucleus around the center
	rootIdx := addCrystalNode(cx, cy, -1, 0, 0.0)
	for k := 0; k < 6; k++ {
		ang := float64(k) * (math.Pi / 3.0)
		prev := rootIdx
		for step := 1; step <= 3; step++ {
			nx := cx + int(math.Round(float64(step)*math.Cos(ang)))
			ny := cy + int(math.Round(float64(step)*math.Sin(ang)))
			if id := addCrystalNode(nx, ny, prev, step, 0.0); id >= 0 {
				prev = id
			}
		}
	}

	// Precompute 64 unit direction vectors for zero-trig fast walker jumps
	var dirCos, dirSin [64]float64
	for i := 0; i < 64; i++ {
		a := float64(i) * (2.0 * math.Pi / 64.0)
		dirCos[i] = math.Cos(a)
		dirSin[i] = math.Sin(a)
	}

	// 8-neighborhood offsets
	dx8 := [8]int{1, 1, 0, -1, -1, -1, 0, 1}
	dy8 := [8]int{0, 1, 1, 1, 0, -1, -1, -1}

	// 3. Aggregate DLA particles with adaptive jump steps and 6-fold hexagonal sticking probability
	maxTotalWalkers := particlesN * 4
	walkersLaunched := 0

	for len(nodes) < particlesN && rMax < maxClusterR && walkersLaunched < maxTotalWalkers {
		walkersLaunched++
		progressT := float64(len(nodes)) / float64(particlesN)

		// Spawn walker on a circle just outside the current maximum cluster radius (rMax + 15),
		// with slight 6-fold angular concentration so the 6 primary dendrites grow lush and balanced.
		spawnR := math.Min(maxClusterR+10.0, rMax+15.0)
		sector := float64(walkersLaunched%6) * (math.Pi / 3.0)
		angOffset := (rng.Float64() - 0.5) * (math.Pi / 3.0) * (1.0 - 0.32*aniso)
		spawnAng := sector + angOffset

		wx := fcx + spawnR*math.Cos(spawnAng)
		wy := fcy + spawnR*math.Sin(spawnAng)
		killR := math.Min(math.Hypot(fcx, fcy)+25.0, rMax+45.0)

		for step := 0; step < 450; step++ {
			ix := int(wx + 0.5)
			iy := int(wy + 0.5)
			if ix <= 1 || ix >= w-2 || iy <= 1 || iy >= h-2 {
				break
			}

			dxC := wx - fcx
			dyC := wy - fcy
			rCur := math.Hypot(dxC, dyC)
			if rCur > killR {
				break
			}

			// Adaptive jump when outside rMax + 4
			if rCur > rMax+4.0 {
				jump := math.Max(1.5, (rCur-rMax)*0.58)
				dIdx := rng.Intn(64)
				// Slight inward radial drift keeps walkers from escaping to infinity
				wx += dirCos[dIdx]*jump - (dxC/rCur)*0.22*jump
				wy += dirSin[dIdx]*jump - (dyC/rCur)*0.22*jump
				continue
			}

			// Coarse 8x8 tile check: if not near any crystal node, jump 3.5 pixels at once
			if !nearCrystalTile[(iy>>3)*tileW+(ix>>3)] {
				dIdx := rng.Intn(64)
				wx += dirCos[dIdx]*3.2 - (dxC/math.Max(1.0, rCur))*0.28
				wy += dirSin[dIdx]*3.2 - (dyC/math.Max(1.0, rCur))*0.28
				continue
			}

			// 1px step on the lattice
			dir8 := rng.Intn(8)
			ix += dx8[dir8]
			iy += dy8[dir8]
			wx = float64(ix)
			wy = float64(iy)

			if occGrid[iy*w+ix] != 0 {
				// Step landed on an occupied cell; back off one pixel
				ix -= dx8[dir8]
				iy -= dy8[dir8]
				wx = float64(ix)
				wy = float64(iy)
				if occGrid[iy*w+ix] != 0 {
					break
				}
			}

			// Check 8 neighbors for an existing crystal site
			parentCandidate := -1
			startN := rng.Intn(8)
			for k := 0; k < 8; k++ {
				nd := (startN + k) & 7
				nx := ix + dx8[nd]
				ny := iy + dy8[nd]
				if val := occGrid[ny*w+nx]; val > 0 {
					parentCandidate = int(val) - 1
					break
				}
			}

			if parentCandidate >= 0 {
				pNode := nodes[parentCandidate]
				// Bond angle from parent to walker + radial azimuth alignment
				bondTheta := math.Atan2(float64(iy-pNode.y), float64(ix-pNode.x))
				radTheta := math.Atan2(float64(iy)-fcy, float64(ix)-fcx)

				// Sticking probability modulated by 1.0 + anisotropy*cos(6*theta)
				bondMod := (1.0 + aniso*math.Cos(6.0*bondTheta)) / (1.0 + aniso)
				radMod := (1.0 + aniso*0.65*math.Cos(6.0*radTheta)) / (1.0 + aniso*0.65)
				stickProb := math.Pow(common.Clamp01(bondMod), 1.65) * (0.35 + 0.65*radMod)

				if rng.Float64() <= stickProb {
					addCrystalNode(ix, iy, parentCandidate, pNode.depth+1, progressT)
					break
				}
			}
		}
	}

	// 4. Compute child subtree count (Strahler / pipe weight) from tips back to the center seed
	for i := len(nodes) - 1; i >= 0; i-- {
		p := nodes[i].parent
		if p >= 0 {
			nodes[p].subtreeWeight += nodes[i].subtreeWeight
		}
	}

	// 5. Render on HDR Canvas
	schema := c.Opts().ColorSchema()
	if len(schema) < 2 {
		schema = common.Oceanic
	}
	grad := render.NewMultiStopGradient(schema...)

	rc := c.NewRenderCanvas()

	// Deep midnight-sapphire winter backdrop with subtle radial crystalline aura & 6-fold frost nebula
	bgOuter := color.RGBA{R: 4, G: 8, B: 20, A: 255}
	bgCore := render.LerpOklab(color.RGBA{R: 10, G: 26, B: 58, A: 255}, grad.SampleOklab(0.18), 0.42)
	bgNebula := render.LerpOklab(color.RGBA{R: 24, G: 88, B: 148, A: 255}, grad.SampleOklab(0.52), 0.45)

	invMaxR := 1.0 / math.Max(1.0, rMax*1.12)
	for py := 0; py < h; py++ {
		fy := float64(py) - fcy
		for px := 0; px < w; px++ {
			fx := float64(px) - fcx
			rad := math.Hypot(fx, fy) * invMaxR
			ang := math.Atan2(fy, fx)
			hexHarmonic := 0.5 + 0.5*math.Cos(6.0*ang)
			fbm := 0.5 + 0.5*noise.FBM2D(float64(px)*0.005/scale, float64(py)*0.005/scale, 3, 2.0, 0.5)

			aura := math.Exp(-2.2*rad*rad) * (0.65 + 0.35*hexHarmonic) * (0.7 + 0.3*fbm)
			col := render.LerpOklab(bgOuter, bgCore, common.Clamp01(1.0-0.75*rad))
			col = render.LerpOklab(col, bgNebula, common.Clamp01(aura*0.45))

			rLin, gLin, bLin, _ := render.ColorToLinearRGB(col)
			rc.SetPixelLinear(px, py, rLin, gLin, bLin, 1.0)
		}
	}

	// Faint hexagonal crystal lattice rings in the background
	rc.SetBlendMode(render.BlendScreen)
	for ring := 1; ring <= 3; ring++ {
		rr := rMax * (0.28 * float64(ring))
		rc.DrawSoftRing(fcx, fcy, rr, 1.4*scale, bgNebula, 0.14)
	}

	// 6. Batch parent-to-child crystal bonds into thickness/color tiers from root to tips
	// Thickness is proportional to log(1 + subtreeWeight), color transitions in Oklab space
	// from deep electric blue at the core to icy cyan and diamond-white at the outer frost tips.
	const numBins = 18
	binPaths := make([]*canvas.Path, numBins)
	for b := 0; b < numBins; b++ {
		binPaths[b] = &canvas.Path{}
	}

	maxLogW := math.Log(1.0 + math.Max(2.0, nodes[0].subtreeWeight*0.35))
	var tipIndices []int

	for i := 1; i < len(nodes); i++ {
		nd := nodes[i]
		if nd.parent < 0 {
			continue
		}
		p := nodes[nd.parent]

		if !nd.hasChild {
			tipIndices = append(tipIndices, i)
		}

		logW := math.Log(1.0 + nd.subtreeWeight)
		wNorm := common.Clamp01(logW / math.Max(1e-6, maxLogW))
		radNorm := common.Clamp01(math.Hypot(nd.fx-fcx, nd.fy-fcy) / math.Max(1.0, rMax))

		// Combine structural hierarchy (wNorm) and radial arrival so trunk veins and delicate tips separate cleanly
		tierMetric := common.Clamp01(0.62*(1.0-wNorm) + 0.38*(0.5*radNorm+0.5*nd.arrivalT))
		bIdx := int(tierMetric * float64(numBins-1))
		bIdx = common.ConstrainInt(bIdx, 0, numBins-1)

		binPaths[bIdx].MoveTo(p.fx, p.fy)
		binPaths[bIdx].LineTo(nd.fx, nd.fy)
	}

	coreElectricBlue := render.LerpOklab(color.RGBA{R: 28, G: 112, B: 238, A: 255}, grad.SampleOklab(0.32), 0.45)
	midIcyCyan := render.LerpOklab(color.RGBA{R: 88, G: 218, B: 252, A: 255}, grad.SampleOklab(0.68), 0.42)
	tipDiamondWhite := render.LerpOklab(color.RGBA{R: 238, G: 250, B: 255, A: 255}, grad.SampleOklab(0.95), 0.25)

	// Render from thickest root veins (b = 0) to finest outer frost needles (b = numBins-1)
	for b := 0; b < numBins; b++ {
		if binPaths[b].Empty() {
			continue
		}
		u := float64(b) / float64(numBins-1) // 0 = heavy root trunk, 1 = delicate outer tip
		weightFactor := 1.0 - u               // 1 at root, 0 at tip

		// Thickness proportional to log(1 + subtreeWeight)
		strokeW := (0.58 + 3.15*math.Pow(weightFactor, 1.35)) * scale

		var bondCol color.RGBA
		if u < 0.5 {
			bondCol = render.LerpOklab(coreElectricBlue, midIcyCyan, u*2.0)
		} else {
			bondCol = render.LerpOklab(midIcyCyan, tipDiamondWhite, (u-0.5)*2.0)
		}

		// Soft outer glow pass for major crystalline branches
		if b < numBins/2 {
			rc.SetBlendMode(render.BlendScreen)
			rc.StrokePath(binPaths[b], strokeW*2.2, canvas.RoundCap, canvas.RoundJoin, bondCol, 0.28)
		}

		// Crisp crystalline core pass
		rc.SetBlendMode(render.BlendAdd)
		alpha := 0.52 + 0.36*(1.0-0.4*u)
		rc.StrokePath(binPaths[b], strokeW, canvas.RoundCap, canvas.RoundJoin, bondCol, alpha)
	}

	// 7. Add specular frost glints at branch tips and central hexagonal seed
	rc.SetBlendMode(render.BlendAdd)
	rc.DrawGlowDot(fcx, fcy, 12.0*scale, midIcyCyan, 0.55, 0.20)
	rc.DrawGlowDot(fcx, fcy, 5.0*scale, tipDiamondWhite, 0.85, 0.55)

	maxGlints := 260
	stride := common.MaxInt(1, len(tipIndices)/maxGlints)
	for i := 0; i < len(tipIndices); i += stride {
		nd := nodes[tipIndices[i]]
		radNorm := math.Hypot(nd.fx-fcx, nd.fy-fcy) / math.Max(1.0, rMax)
		if radNorm < 0.22 {
			continue
		}
		glintR := (1.4 + 2.2*math.Pow(radNorm, 1.4)*rng.Float64()) * scale
		glintCol := render.LerpOklab(midIcyCyan, tipDiamondWhite, common.Clamp01(radNorm*1.1))
		rc.DrawGlowDot(nd.fx, nd.fy, glintR, glintCol, 0.42+0.35*radNorm, 0.30)
	}

	// 8. Post-processing & ACES filmic tone-mapped resolve
	rc.ApplyBloom(0.45, 8.0, 0.6)
	rc.ApplyVignette(0.3)
	c.MergeRenderCanvasToneMapped(rc, 1.12, 0)
}
