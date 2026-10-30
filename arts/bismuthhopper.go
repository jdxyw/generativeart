package arts

import (
	"image/color"
	"math"
	"math/rand"
	"sort"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type bismuthHopper struct {
	terraces int
	crystals int
	oxideNm  float64
}

// NewBismuthHopper returns a bismuthHopper engine that synthesizes a 3D architectural
// Synthetic Bismuth (Bi) Hopper Crystal labyrinth (合成铋晶体阶梯漏斗迷宫与氧化薄膜干涉)
// with nested Greek-key rectangular spiral stair-step cavities, 45-degree chamfered metallic
// bevel ledges, multi-directional horizon soft shadows, and Airy thin-film interference
// across a Bi2O3 oxide layer (IOR ~ 2.05-2.15).
func NewBismuthHopper(terraces, crystals int, oxideNm float64) *bismuthHopper {
	if terraces <= 0 {
		terraces = 14
	}
	if crystals <= 0 {
		crystals = 9
	}
	if oxideNm <= 0 {
		oxideNm = 420.0
	}
	return &bismuthHopper{
		terraces: terraces,
		crystals: crystals,
		oxideNm:  oxideNm,
	}
}

type hopperCrystalTower struct {
	cx, cy      float64
	halfSize    float64
	cosA, sinA  float64
	terraces    float64
	maxHeight   float64
	spiralHand  float64
	oxideOffset float64
	subFreq     float64
	isMacro     bool
}

// Generative renders a static frame of the 3D Synthetic Bismuth Hopper Crystal labyrinth.
func (bh *bismuthHopper) Generative(c *generativeart.Canva) {
	bh.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360-degree phase loop (phase in [0, 1)).
func (bh *bismuthHopper) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	minDim := math.Min(fw, fh)
	scale := minDim / 600.0
	animAngle := 2.0 * math.Pi * phase

	rc := c.NewRenderCanvas()
	rc.FillBackground(color.RGBA{R: 7, G: 9, B: 16, A: 255})

	numTerraces := bh.terraces
	if numTerraces < 6 {
		numTerraces = 6
	} else if numTerraces > 28 {
		numTerraces = 28
	}
	numCrystals := bh.crystals
	if numCrystals < 3 {
		numCrystals = 3
	} else if numCrystals > 24 {
		numCrystals = 24
	}
	baseOxide := bh.oxideNm
	if baseOxide < 180.0 {
		baseOxide = 420.0
	}

	// 1. Seed hierarchical hopper crystal centers via Poisson-disk sampling + Centroidal Voronoi relaxation
	rng := rand.New(rand.NewSource(c.Seed() + 1804))
	densityFn := func(x, y float64) float64 {
		dx := (x - 0.5*fw) / (0.5 * fw)
		dy := (y - 0.5*fh) / (0.5 * fh)
		r := math.Hypot(dx, dy)
		// Encourage a ring of interlocking satellite hoppers around the central macro hopper
		ring := math.Exp(-math.Pow((r-0.56)/0.28, 2.0))
		return common.Constrain(0.35+0.65*ring, 0.0, 1.0)
	}

	poissonRadius := minDim / (math.Sqrt(float64(numCrystals)) * 1.45)
	rawPts := render.SamplePoissonDisk2D(fw, fh, poissonRadius, densityFn, rng)
	relaxedPts := render.RelaxCentroidalVoronoi(rawPts, fw, fh, 96, densityFn, 10)

	// Sort candidate sites by distance to the ideal satellite ring around the center
	cx0, cy0 := 0.50*fw, 0.50*fh
	sort.Slice(relaxedPts, func(i, j int) bool {
		di := math.Abs(math.Hypot(relaxedPts[i].X-cx0, relaxedPts[i].Y-cy0) - 0.31*minDim)
		dj := math.Abs(math.Hypot(relaxedPts[j].X-cx0, relaxedPts[j].Y-cy0) - 0.31*minDim)
		return di < dj
	})

	towers := make([]hopperCrystalTower, 0, numCrystals+4)

	// Dominant central macro hopper crystal (axis-aligned monumental skeletal hopper pyramid)
	towers = append(towers, hopperCrystalTower{
		cx:          cx0,
		cy:          cy0,
		halfSize:    0.34 * minDim,
		cosA:        1.0,
		sinA:        0.0,
		terraces:    float64(numTerraces),
		maxHeight:   30.0 * scale,
		spiralHand:  1.0,
		oxideOffset: 0.0,
		subFreq:     3.0,
		isMacro:     true,
	})

	// Four symmetric corner satellite hopper towers interlocking at 45-degree and axis-aligned angles
	satCoords := [8][3]float64{
		{0.15 * fw, 0.15 * fh, 0.25 * math.Pi},
		{0.85 * fw, 0.15 * fh, -0.25 * math.Pi},
		{0.15 * fw, 0.85 * fh, -0.25 * math.Pi},
		{0.85 * fw, 0.85 * fh, 0.25 * math.Pi},
		{0.50 * fw, 0.10 * fh, 0.0},
		{0.50 * fw, 0.90 * fh, 0.50 * math.Pi},
		{0.10 * fw, 0.50 * fh, 0.50 * math.Pi},
		{0.90 * fw, 0.50 * fh, 0.0},
	}
	for k, sc := range satCoords {
		if len(towers) >= numCrystals+3 {
			break
		}
		hand := 1.0
		if k%2 == 1 {
			hand = -1.0
		}
		sz := 0.23 * minDim
		if k >= 4 {
			sz = 0.19 * minDim
		}
		towers = append(towers, hopperCrystalTower{
			cx:          sc[0],
			cy:          sc[1],
			halfSize:    sz,
			cosA:        math.Cos(sc[2]),
			sinA:        math.Sin(sc[2]),
			terraces:    float64(common.MaxInt(7, numTerraces-3)),
			maxHeight:   (20.0 - 1.5*float64(k/4)) * scale,
			spiralHand:  hand,
			oxideOffset: float64(k*55 - 140),
			subFreq:     2.0,
			isMacro:     false,
		})
	}

	// Additional Voronoi-relaxed peripheral crystal towers
	quantAngles := [4]float64{0.0, 0.25 * math.Pi, 0.50 * math.Pi, -0.25 * math.Pi}
	for i := 0; i < len(relaxedPts) && len(towers) < numCrystals+5; i++ {
		pt := relaxedPts[i]
		distCenter := math.Hypot(pt.X-cx0, pt.Y-cy0)
		if distCenter < 0.36*minDim {
			continue
		}
		idx := len(towers)
		ang := quantAngles[idx%4]
		hand := 1.0
		if idx%2 == 1 {
			hand = -1.0
		}
		towers = append(towers, hopperCrystalTower{
			cx:          pt.X,
			cy:          pt.Y,
			halfSize:    0.17 * minDim,
			cosA:        math.Cos(ang),
			sinA:        math.Sin(ang),
			terraces:    float64(common.MaxInt(6, numTerraces-4)),
			maxHeight:   17.0 * scale,
			spiralHand:  hand,
			oxideOffset: float64((idx*67)%220) - 90.0,
			subFreq:     2.0,
			isMacro:     false,
		})
	}

	// Evaluate a single Bismuth Hopper Crystal tower at world position (px, py)
	evalTower := func(t hopperCrystalTower, px, py float64) (z, ledgeDist, oxideNm, terraceIdx, rimFactor float64) {
		dx := px - t.cx
		dy := py - t.cy
		u := (dx*t.cosA + dy*t.sinA) / t.halfSize
		v := (-dx*t.sinA + dy*t.cosA) / t.halfSize

		ax := math.Abs(u)
		ay := math.Abs(v)
		rInf := math.Max(ax, ay)
		if rInf >= 1.18 {
			return -1.0, 1.0, baseOxide, 0.0, 0.0
		}

		// Razor-straight crystallographic rectangular hopper walls (C1 continuous around all 4 walls)
		spiralCoord := rInf * t.terraces

		// Smooth stair-step terrace quantization with crisp risers and flat horizontal ledges
		kStep := math.Floor(spiralCoord)
		frac := spiralCoord - kStep

		// Riser occupies [0, 0.16] and [0.88, 1.0], leaving [0.16, 0.88] as a flat metallic terrace ledge
		riserUp := common.Smoothstep(0.0, 0.15, frac)
		riserNext := common.Smoothstep(0.87, 1.0, frac)
		quantStep := (kStep + riserUp + riserNext - 0.5) / t.terraces

		// Hopper envelope: outer crystal perimeter is elevated, inner cavity sinks into a deep pyramidal well,
		// while each concentric wall forms a raised rim ridge (true hopper skeletal growth habit).
		pedestal := 0.20
		if t.isMacro {
			pedestal = 0.42
		}
		hopperWell := pedestal + (1.0-pedestal)*common.Constrain(quantStep, 0.0, 1.08)

		// Raised skeletal rim lip on the outer edge of each terrace step
		ledgeCenterFrac := math.Abs(frac - 0.51)
		edgeDist := 0.36 - ledgeCenterFrac // positive on flat terrace, near 0 at chamfered bevel edge
		chamferLip := 0.056 * math.Exp(-math.Pow((frac-0.18)/0.062, 2.0))

		// Diagonal crystallographic growth striation groove where L_inf faces meet (|u| == |v|)
		diagDist := math.Abs(ax-ay) / math.Max(0.08, rInf)
		miterRidge := 0.032 * math.Exp(-math.Pow(diagDist/0.035, 2.0))

		// Outer crystal cliff drop-off beyond rInf = 1.0
		outerCliff := 1.0 - common.Smoothstep(0.97, 1.03, rInf)
		outerBevel := common.Smoothstep(0.93, 0.99, rInf) * (1.0 - common.Smoothstep(0.99, 1.03, rInf))

		elevNorm := (hopperWell + chamferLip + miterRidge) * outerCliff
		z = elevNorm * t.maxHeight

		// Distance to nearest step/miter edge in [0, 1] (0 = right on chamfered bevel rim)
		stepEdgeProx := math.Exp(-math.Pow((frac-0.15)/0.050, 2.0)) +
			math.Exp(-math.Pow((frac-0.88)/0.050, 2.0)) +
			0.75*math.Exp(-math.Pow(diagDist/0.035, 2.0)) +
			0.90*outerBevel
		rimFactor = common.Constrain(stepEdgeProx, 0.0, 1.0)
		ledgeDist = common.Constrain(edgeDist/0.36, 0.0, 1.0)

		// First- and second-order Newton thin-film oxide thickness d_nm(x, y) in [62 nm, 248 nm]
		// at IOR = 2.05: sweeps cleanly through Gold/Amber -> Crimson/Magenta -> Royal Violet -> Electric Cyan -> Emerald/Gold!
		wellToRim := common.Constrain(0.74*rInf+0.26*elevNorm, 0.0, 1.05)
		azimWave := 14.0 * math.Cos(2.0*math.Atan2(v, u)*t.spiralHand+0.25*kStep+animAngle)
		oxideNm = 62.0 + 178.0*wellToRim + (baseOxide-420.0)*0.12 + t.oxideOffset*0.18 + azimWave

		return z, ledgeDist, oxideNm, kStep, rimFactor
	}

	// 2. Populate G-Buffer heightfield Z[w*h] and auxiliary per-pixel crystal material channels
	gb := render.NewFiberGBuffer(w, h, c.Opts().Background())
	oxideBuf := make([]float64, w*h)
	rimBuf := make([]float64, w*h)
	ledgeBuf := make([]float64, w*h)
	stepIdxBuf := make([]float64, w*h)

	ne := common.NewNoiseEngine(c.Seed() + 909)

	for y := 0; y < h; y++ {
		fy := float64(y) + 0.5
		rowOff := y * w
		for x := 0; x < w; x++ {
			fx := float64(x) + 0.5

			bestZ := 0.25 * scale
			secondZ := 0.0
			bestOxide := baseOxide
			bestRim := 0.0
			bestLedge := 0.5
			bestStep := 0.0

			for i := range towers {
				tz, tLedge, tOxide, tStep, tRim := evalTower(towers[i], fx, fy)
				if tz <= 0 {
					continue
				}
				if tz > bestZ {
					secondZ = bestZ
					bestZ = tz
					bestOxide = tOxide
					bestRim = tRim
					bestLedge = tLedge
					bestStep = tStep
				} else if tz > secondZ {
					secondZ = tz
				}
			}

			// Inter-crystal grain boundary seam where two hopper towers interlock
			if secondZ > 0.5*scale {
				seamDiff := (bestZ - secondZ) / math.Max(1.0, scale)
				if seamDiff < 1.35 {
					seamGroove := (1.0 - seamDiff/1.35)
					bestZ -= 0.85 * scale * seamGroove
					bestRim = math.Max(bestRim, 0.88*seamGroove)
					bestOxide += 38.0 * seamGroove
				}
			}

			// Subtle micro-crystalline metallic growth terrace striation
			microNoise := ne.FBM2D(fx/(18.0*scale), fy/(18.0*scale), 2, 2.0, 0.5)
			bestZ = math.Max(0.15*scale, bestZ+0.04*scale*microNoise)
			bestOxide = common.Constrain(bestOxide+4.0*microNoise, 52.0, 265.0)

			idx := rowOff + x
			gb.Z[idx] = bestZ
			oxideBuf[idx] = bestOxide
			rimBuf[idx] = bestRim
			ledgeBuf[idx] = bestLedge
			stepIdxBuf[idx] = bestStep
		}
	}

	// 3. Multi-directional Horizon Soft Shadows (Key Sun + Opposite Fill Sun)
	keyLight := [3]float64{-0.52, -0.58, 0.63}
	fillLight := [3]float64{0.56, 0.42, 0.71}
	horizonRad := common.MaxInt(10, int(math.Round(24.0*scale)))
	keyVis := gb.ComputeHorizonShadows(keyLight, horizonRad, 1.45*scale)
	fillVis := gb.ComputeHorizonShadows(fillLight, common.MaxInt(8, horizonRad/2), 1.85*scale)

	// Normalize key and fill light vectors & half-vectors for Blinn-Phong metallic specular
	normVec := func(v [3]float64) (float64, float64, float64) {
		l := math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
		if l < 1e-9 {
			return 0, 0, 1
		}
		return v[0] / l, v[1] / l, v[2] / l
	}
	klx, kly, klz := normVec(keyLight)
	khx, khy, khz := normVec([3]float64{klx, kly, klz + 1.0})

	flx, fly, flz := normVec(fillLight)
	fhx, fhy, fhz := normVec([3]float64{flx, fly, flz + 1.0})

	// 12-tap multi-scale ambient occlusion offsets for deep hopper wells
	r1 := common.MaxInt(1, int(math.Round(2.0*scale)))
	r2 := common.MaxInt(2, int(math.Round(5.0*scale)))
	r3 := common.MaxInt(4, int(math.Round(10.0*scale)))
	aoOffsets := [12][2]int{
		{-r1, 0}, {r1, 0}, {0, -r1}, {0, r1},
		{-r2, -r1}, {r2, -r1}, {-r2, r1}, {r2, r1},
		{-r3, -r2}, {r3, -r2}, {-r2, r3}, {r2, r3},
	}

	maxZ := 28.0 * scale
	invScale := 1.0 / math.Max(0.25, scale)

	// 4. Deferred 3D Normal, Cavity AO & Physical Bi2O3 Airy Thin-Film Metallic Shader
	for y := 0; y < h; y++ {
		ym1 := common.MaxInt(0, y-1)
		yp1 := common.MinInt(h-1, y+1)
		rowOff := y * w
		for x := 0; x < w; x++ {
			xm1 := common.MaxInt(0, x-1)
			xp1 := common.MinInt(w-1, x+1)
			idx := rowOff + x
			z0 := gb.Z[idx]

			// Central-difference 3D surface normal N = (nx, ny, nz)
			dzdx := (gb.Z[y*w+xp1] - gb.Z[y*w+xm1]) * 0.5 * invScale
			dzdy := (gb.Z[yp1*w+x] - gb.Z[ym1*w+x]) * 0.5 * invScale
			slopeMag := math.Hypot(dzdx, dzdy)

			nx := -dzdx * 1.35
			ny := -dzdy * 1.35
			nz := 1.0
			nLen := math.Sqrt(nx*nx + ny*ny + nz*nz)
			nx, ny, nz = nx/nLen, ny/nLen, nz/nLen

			// Classify surface geometry: flat horizontal terrace vs 45-degree chamfer bevel vs vertical riser cliff
			isRiserCliff := common.Smoothstep(0.65, 2.20, slopeMag)
			chamferBevel := math.Max(
				rimBuf[idx],
				math.Exp(-math.Pow((slopeMag-0.85)/0.45, 2.0)),
			)

			// 12-tap Ambient Occlusion in sunken hopper wells
			occ := 0.0
			ridgeLift := 0.0
			for _, off := range aoOffsets {
				sx := common.ConstrainInt(x+off[0], 0, w-1)
				sy := common.ConstrainInt(y+off[1], 0, h-1)
				dz := (gb.Z[sy*w+sx] - z0) * invScale
				if dz > 0.12 {
					occ += common.Constrain(dz/4.2, 0.0, 1.0)
				} else if dz < -0.12 {
					ridgeLift += common.Constrain(-dz/3.6, 0.0, 1.0)
				}
			}
			cavityAO := common.Constrain(1.0-(occ/12.0)*0.86, 0.08, 1.0)
			elevFactor := common.Constrain(z0/maxZ, 0.0, 1.0)
			ao := cavityAO * (0.14 + 0.86*math.Pow(elevFactor, 1.15))

			// Multi-directional horizon shadow visibility
			kShadow := 1.0
			if keyVis != nil {
				kShadow = 0.12 + 0.88*keyVis[idx]
			}
			fShadow := 1.0
			if fillVis != nil {
				fShadow = 0.18 + 0.82*fillVis[idx]
			}

			// Viewing cosine angle and effective oxide film thickness on risers vs terraces
			cosTheta := common.Constrain(nz, 0.08, 1.0)
			localOxideNm := oxideBuf[idx] + 18.0*isRiserCliff - 10.0*chamferBevel

			// Evaluate physical Bi2O3 thin-film structural color (IOR = 2.05)
			iridR, iridG, iridB := render.SampleThinFilmLinearRGB(cosTheta, localOxideNm, 2.05)
			irid2R, irid2G, irid2B := render.SampleThinFilmLinearRGB(cosTheta*0.94, localOxideNm*0.88+14.0, 2.08)
			// Non-linear chroma contrast expansion so bismuth oxide displays jewel-grade cyan, gold, magenta & violet
			tfR := math.Pow(common.Constrain(0.70*iridR+0.30*irid2R, 0.0, 1.0), 2.35)
			tfG := math.Pow(common.Constrain(0.70*iridG+0.30*irid2G, 0.0, 1.0), 2.35)
			tfB := math.Pow(common.Constrain(0.70*iridB+0.30*irid2B, 0.0, 1.0), 2.35)

			// Underlying raw dark gunmetal/pewter Bismuth base reflectance
			biMetalR, biMetalG, biMetalB := 0.14, 0.13, 0.18

			// Schlick metallic Fresnel reflectance F(cosTheta)
			oneMinusCos := 1.0 - cosTheta
			fresnel := 0.42 + 0.58*math.Pow(oneMinusCos, 2.6)

			// Blend dark gunmetal base with vivid thin-film oxide interference
			mR := (0.12*biMetalR + 1.25*tfR) * fresnel
			mG := (0.12*biMetalG + 1.25*tfG) * fresnel
			mB := (0.12*biMetalB + 1.25*tfB) * fresnel

			// Directional key & fill illumination
			ndotlKey := math.Max(0.0, nx*klx+ny*kly+nz*klz)
			ndotlFill := math.Max(0.0, nx*flx+ny*fly+nz*flz)

			keyDiff := (0.08 + 0.92*ndotlKey) * kShadow
			fillDiff := (0.05 + 0.65*ndotlFill) * fShadow
			skyDome := (0.08 + 0.18*nz) * ao

			illum := (0.72*keyDiff + 0.28*fillDiff + skyDome) * ao

			// Razor-sharp metallic specular highlights along 45-degree chamfered bevel edges & terraces
			ndothKey := math.Max(0.0, nx*khx+ny*khy+nz*khz)
			ndothFill := math.Max(0.0, nx*fhx+ny*fhy+nz*fhz)

			specTerrace := math.Pow(ndothKey, 52.0)*0.55*kShadow + math.Pow(ndothFill, 32.0)*0.22*fShadow
			// Chamfered step edges catch intense linear metallic glints
			edgeGlint := chamferBevel * (0.22 + 0.95*math.Pow(ndothKey, 14.0)*kShadow + 0.40*math.Pow(ndothFill, 10.0)*fShadow) *
				(0.45 + 0.55*(ridgeLift/12.0))

			// Tint specular highlights with shifted thin-film gold/cyan/rose reflections
			specTintR := 0.35 + 0.85*tfR
			specTintG := 0.33 + 0.85*tfG
			specTintB := 0.38 + 0.85*tfB

			// Subtle step-to-step terrace sheen modulation
			stepMod := 0.88 + 0.12*math.Sin(stepIdxBuf[idx]*1.7+ledgeBuf[idx]*math.Pi)

			rOut := mR*illum*stepMod + (specTerrace+edgeGlint)*specTintR
			gOut := mG*illum*stepMod + (specTerrace+edgeGlint)*specTintG
			bOut := mB*illum*stepMod + (specTerrace+edgeGlint)*specTintB

			rc.SetPixelLinear(x, y, rOut, gOut, bOut, 1.0)
		}
	}

	// 5. Composable HDR Post-Processing Pass
	rc.ApplyLuminanceRelief(0.38, -135.0, 42.0, 0.28)
	rc.ApplyBloom(0.82, 3.5*scale, 0.14)
	rc.ApplyColorGrade(1.22, 1.24, 235.0, 42.0, 0.16)
	rc.ApplyVignette(0.28)
	c.MergeRenderCanvas(rc)
}
