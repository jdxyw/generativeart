package arts

import (
	"image/color"
	"math"
	"math/rand"
	"runtime"
	"sync"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type kintsugiGeode struct {
	plates    int
	tiers     int
	warpScale float64
}

// NewKintsugiGeode returns a KintsugiGeode generator that synthesizes a full-bleed
// 2x2-supersampled 3D Voronoi Kintsugi Gold-River & Banded Fortification-Agate
// Tectonic Relief: 24 Lloyd-relaxed Perlin-warped tectonic plates that step upward
// through 4 concentric combed-agate & silk-brocade terraces, separated by deep
// obsidian canyon fissures filled with gleaming 3D 24K Kintsugi gold veins.
func NewKintsugiGeode(plates, tiers int, warpScale float64) *kintsugiGeode {
	if plates <= 0 {
		plates = 24
	}
	if tiers <= 0 {
		tiers = 4
	}
	if warpScale <= 0 {
		warpScale = 1.35
	}
	return &kintsugiGeode{
		plates:    common.ConstrainInt(plates, 12, 42),
		tiers:     common.ConstrainInt(tiers, 3, 7),
		warpScale: common.Constrain(warpScale, 0.5, 2.5),
	}
}

// Generative renders the static KintsugiGeode composition.
func (kg *kintsugiGeode) Generative(c *generativeart.Canva) {
	kg.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360° looping KintsugiGeode animation.
func (kg *kintsugiGeode) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	phi := 2.0 * math.Pi * phase

	seed := c.Seed()
	if seed == 0 {
		seed = 20263301
	}
	rng := rand.New(rand.NewSource(seed))
	ne := common.NewNoiseEngine(seed)

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	wScale := kg.warpScale
	loopX := 0.52 * (math.Cos(phi) - 1.0)
	loopY := 0.52 * math.Sin(phi)

	// =========================================================================
	// 1. GENERATE & LLOYD-RELAX 24 VORONOI TECTONIC AGATE SEEDS
	// =========================================================================
	nSeeds := kg.plates
	if nSeeds > 32 {
		nSeeds = 24
	}
	type plateSeed struct {
		x0, y0   float64
		x, y     float64
		hueBase  float64
		phaseOff float64
		dir      float64
	}

	seeds := make([]plateSeed, nSeeds)
	gridN := int(math.Ceil(math.Sqrt(float64(nSeeds))))
	for i := 0; i < nSeeds; i++ {
		gx := i % gridN
		gy := i / gridN
		u := (float64(gx) + 0.20 + 0.60*rng.Float64()) / float64(gridN)
		v := (float64(gy) + 0.20 + 0.60*rng.Float64()) / float64(gridN)
		seeds[i] = plateSeed{
			x0:       u,
			y0:       v,
			phaseOff: rng.Float64() * 2.0 * math.Pi,
			dir:      1.0 - 2.0*float64(i%2),
		}
	}

	// 3 steps of fast grid-based Lloyd relaxation on the base positions (x0, y0)
	const lloydRes = 80
	for iter := 0; iter < 3; iter++ {
		sumX := make([]float64, nSeeds)
		sumY := make([]float64, nSeeds)
		cnt := make([]int, nSeeds)
		for py := 0; py < lloydRes; py++ {
			uy := (float64(py) + 0.5) / float64(lloydRes)
			for px := 0; px < lloydRes; px++ {
				ux := (float64(px) + 0.5) / float64(lloydRes)
				bestD := 1e9
				bestIdx := 0
				for i := 0; i < nSeeds; i++ {
					d := (ux-seeds[i].x0)*(ux-seeds[i].x0) + (uy-seeds[i].y0)*(uy-seeds[i].y0)
					if d < bestD {
						bestD = d
						bestIdx = i
					}
				}
				sumX[bestIdx] += ux
				sumY[bestIdx] += uy
				cnt[bestIdx]++
			}
		}
		for i := 0; i < nSeeds; i++ {
			if cnt[i] > 0 {
				seeds[i].x0 = sumX[i] / float64(cnt[i])
				seeds[i].y0 = sumY[i] / float64(cnt[i])
			}
		}
	}

	for i := 0; i < nSeeds; i++ {
		// Pronounced orbital drift of each Voronoi tectonic plate over the 360-degree GIF loop (0 at phi=0!)
		orbR := 0.048
		seeds[i].x = seeds[i].x0 + orbR*(math.Cos(phi*seeds[i].dir+seeds[i].phaseOff)-math.Cos(seeds[i].phaseOff))
		seeds[i].y = seeds[i].y0 + orbR*(math.Sin(phi*seeds[i].dir+seeds[i].phaseOff)-math.Sin(seeds[i].phaseOff))
		seeds[i].hueBase = math.Mod(0.58*seeds[i].x0+0.48*(1.0-seeds[i].y0)+0.12*float64(i%3)+10.0, 1.0)
	}

	// Curated Dunhuang & Brazilian Fortification-Agate Palette (overridable via -palette / c.SetColorSchema)
	agateGrad := render.ResolveCyclicGradient(c.Opts().ColorSchema(), common.Dunhuang,
		color.RGBA{R: 16, G: 44, B: 118, A: 255},  // Deep Lapis Sapphire
		color.RGBA{R: 20, G: 132, B: 158, A: 255}, // Kingfisher Cyan-Turquoise
		color.RGBA{R: 42, G: 172, B: 128, A: 255}, // Celadon Malachite Jade
		color.RGBA{R: 248, G: 194, B: 56, A: 255}, // Imperial Sunburst Gold
		color.RGBA{R: 240, G: 106, B: 48, A: 255}, // Sunset Tangerine Coral
		color.RGBA{R: 198, G: 38, B: 56, A: 255},  // Cinnabar Vermilion Ruby
		color.RGBA{R: 118, G: 26, B: 88, A: 255},  // Bordeaux Rose-Plum
		color.RGBA{R: 16, G: 44, B: 118, A: 255},  // Loop back to Sapphire
	)

	goldRim := color.RGBA{R: 255, G: 236, B: 168, A: 255}
	goldBody := color.RGBA{R: 238, G: 180, B: 48, A: 255}
	goldShadow := color.RGBA{R: 92, G: 58, B: 16, A: 255}
	canyonBed := color.RGBA{R: 10, G: 14, B: 32, A: 255}

	// =========================================================================
	// 2. RASTERIZE 2X SUPERSAMPLED Z-BUFFER, STEPPED AGATE TIERS & KINTSUGI GOLD
	// =========================================================================
	gw, gh := w*2+1, h*2+1
	zGrid := make([]float64, gw*gh)
	idGrid := make([]int, gw*gh)
	rimGrid := make([]float64, gw*gh)
	albRGrid := make([]float64, gw*gh)
	albGGrid := make([]float64, gw*gh)
	albBGrid := make([]float64, gw*gh)

	nTiers := kg.tiers
	if nTiers < 3 {
		nTiers = 4
	}

	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}

	var wgRaster sync.WaitGroup
	for wk := 0; wk < workers; wk++ {
		gy0 := (wk * gh) / workers
		gy1 := ((wk + 1) * gh) / workers
		wgRaster.Add(1)
		go func(r0, r1 int) {
			defer wgRaster.Done()
			for gy := r0; gy < r1; gy++ {
				uy := (float64(gy) * 0.5) / fh
				rowOff := gy * gw
				for gx := 0; gx < gw; gx++ {
					ux := (float64(gx) * 0.5) / fw
					idx := rowOff + gx

					// Smooth two-octave Perlin domain warp so Voronoi boundaries curve like natural agate & canyon fissures!
					wx, _, _ := ne.FBM2DDeriv(ux*2.3+1.4+loopX*0.35, uy*2.3-0.9+loopY*0.35, 2, 2.0, 0.45)
					wy, _, _ := ne.FBM2DDeriv(ux*2.3-2.1-loopY*0.35, uy*2.3+1.8+loopX*0.35, 2, 2.0, 0.45)
					uxW := ux + 0.032*wScale*wx
					uyW := uy + 0.032*wScale*wy

					// Find nearest Voronoi seed c1, then compute the exact minimum bisector distance
					// across ALL other seeds j != c1 so every fortification-agate polygon is 100% closed & seamless!
					d1 := 1e9
					c1 := 0
					for i := 0; i < nSeeds; i++ {
						dx := uxW - seeds[i].x
						dy := uyW - seeds[i].y
						d := dx*dx + dy*dy
						if d < d1 {
							d1 = d
							c1 = i
						}
					}

					sx1, sy1 := seeds[c1].x, seeds[c1].y
					dBorder := 1e9
					for j := 0; j < nSeeds; j++ {
						if j == c1 {
							continue
						}
						sx2, sy2 := seeds[j].x, seeds[j].y
						sep := math.Max(1e-4, math.Hypot(sx2-sx1, sy2-sy1))
						mx, my := 0.5*(sx1+sx2), 0.5*(sy1+sy2)
						distJ := ((mx-uxW)*(sx2-sx1) + (my-uyW)*(sy2-sy1)) / sep
						if distJ < dBorder {
							dBorder = distJ
						}
					}
					if dBorder < 0 {
						dBorder = 0
					}
					dBorderPx := dBorder * fw

					// Canyon trench width = 5.5px; Kintsugi gold vein runs down the center (dBorderPx < 2.2px)!
					const trenchPx = 5.5
					const goldVeinPx = 2.15

					if dBorderPx < trenchPx {
						// Recessed Kintsugi Fissure Canyon Floor + 3D Cylindrical 24K Gold River Vein!
						curCol := canyonBed
						curZ := 1.0
						curID := 1
						curRim := 0.0

						if dBorderPx < goldVeinPx {
							uGold := dBorderPx / goldVeinPx
							if uGold < 0.42 {
								curCol = render.LerpOklab(goldRim, goldBody, uGold/0.42)
							} else {
								curCol = render.LerpOklab(goldBody, goldShadow, (uGold-0.42)/0.58)
							}
							curZ = 3.2 + 1.4*(1.0-uGold*uGold)
							curID = 2
							curRim = 1.0 - uGold*0.5
						}

						rL, gL, bL, _ := render.ColorToLinearRGB(curCol)
						zGrid[idx] = curZ
						idGrid[idx] = curID
						rimGrid[idx] = curRim
						albRGrid[idx] = rL
						albGGrid[idx] = gL
						albBGrid[idx] = bL
						continue
					}

					// Inside Voronoi Agate Plate c1: step upward through nTiers concentric fortification tiers!
					// Add continuous radial tier-wave propagation (+0.55*sin(phi)) and full-cycle rib ripple (-dir*phi)!
					inwardPx := dBorderPx - trenchPx
					pTier := ne.Perlin2D(ux*4.2+float64(c1)*5.7+0.45*loopX, uy*4.2-float64(c1)*3.1+0.45*loopY)
					tierShift := 6.8 * math.Sin(phi*seeds[c1].dir+seeds[c1].phaseOff) - 6.8*math.Sin(seeds[c1].phaseOff)
					tierWidthPx := 16.5
					qTier := (inwardPx + 2.8*pTier + tierShift) / tierWidthPx
					if qTier < 0 {
						qTier = 0
					}
					kTier := int(math.Floor(qTier))
					if kTier >= nTiers {
						kTier = nTiers - 1
					}
					fTier := qTier - float64(kTier)
					if kTier == nTiers-1 && qTier >= float64(nTiers-1) {
						fTier = common.Clamp01((qTier - float64(nTiers-1)) * 0.5)
					}

					// Distance in pixels from the current tier's outer cliff step
					distStepPx := (qTier - float64(kTier)) * tierWidthPx
					if kTier == 0 {
						distStepPx = inwardPx
					}

					// Sample harmonious tier colors for plate c1 and tier kTier (with subtle cyclic hue breathing)
					hueShift := 0.08 * (math.Sin(phi*seeds[c1].dir+seeds[c1].phaseOff) - math.Sin(seeds[c1].phaseOff))
					hue0 := math.Mod(seeds[c1].hueBase+float64(kTier)*0.11+hueShift+10.0, 1.0)
					colA := agateGrad.SampleOklab(hue0)
					colB := agateGrad.SampleOklab(math.Mod(hue0+0.08, 1.0))
					pCol := render.LerpOklab(colA, colB, common.Clamp01(fTier))

					// Internal concentric liesegang silk-agate striation ribs travelling continuously over 2*pi!
					ribWave := math.Sin(distStepPx*0.82 - seeds[c1].dir*phi)
					if ribWave > 0 {
						pCol = render.LerpOklab(pCol, color.RGBA{R: 255, G: 248, B: 224, A: 255}, ribWave*0.14)
					} else {
						pCol = render.LerpOklab(pCol, color.RGBA{R: 10, G: 8, B: 22, A: 255}, -ribWave*0.13)
					}

					// Inner 24K gold pinstripe at distStepPx ≈ 7.5px
					innerPin := math.Exp(-math.Pow((distStepPx-7.5)/1.2, 2.0))
					if innerPin > 0.02 {
						pCol = render.LerpOklab(pCol, goldBody, innerPin*0.78)
					}

					curZ := 6.0 + float64(kTier)*3.2 + 0.40*math.Exp(-distStepPx/5.5)
					curID := (c1+1)*100 + kTier
					curRim := 0.0
					if distStepPx < 1.75 {
						curRim = common.Smoothstep(1.75, 0.35, distStepPx)
					}

					grain := ne.Perlin2D(float64(gx)*0.35, float64(gy)*0.35) * 0.022
					rL, gL, bL, _ := render.ColorToLinearRGB(pCol)
					if curRim > 0 {
						rr, rg, rb, _ := render.ColorToLinearRGB(goldRim)
						rL = rL*(1.0-curRim*0.84) + rr*curRim*0.84
						gL = gL*(1.0-curRim*0.84) + rg*curRim*0.84
						bL = bL*(1.0-curRim*0.84) + rb*curRim*0.84
					}

					zGrid[idx] = curZ
					idGrid[idx] = curID
					rimGrid[idx] = curRim
					albRGrid[idx] = common.Clamp01(rL + grain)
					albGGrid[idx] = common.Clamp01(gL + grain)
					albBGrid[idx] = common.Clamp01(bL + grain)
				}
			}
		}(gy0, gy1)
	}
	wgRaster.Wait()

	// =========================================================================
	// 3. PARALLEL 2X2 SUPERSAMPLED SHADOW-BOX DROP SHADOWS & AO
	// =========================================================================
	subOffs := [4][2]int{{0, 0}, {1, 0}, {0, 1}, {1, 1}}
	aoOffsets := [12][2]int{
		{-3, -3}, {0, -4}, {3, -3}, {-4, 0},
		{4, 0}, {-3, 3}, {0, 4}, {3, 3},
		{-8, -6}, {8, -6}, {-6, 8}, {6, 8},
	}

	var wgShade sync.WaitGroup
	chunk := (h + workers - 1) / workers
	for wk := 0; wk < workers; wk++ {
		y0 := wk * chunk
		y1 := y0 + chunk
		if y0 >= h {
			break
		}
		if y1 > h {
			y1 = h
		}
		wgShade.Add(1)
		go func(startY, endY int) {
			defer wgShade.Done()
			for y := startY; y < endY; y++ {
				for x := 0; x < w; x++ {
					var sumR, sumG, sumB float64

					for _, so := range subOffs {
						gx := x*2 + so[0]
						gy := y*2 + so[1]
						idx := gy*gw + gx
						z0 := zGrid[idx]
						id0 := idGrid[idx]

						aoSum := 0.0
						for _, off := range aoOffsets {
							sx := common.ConstrainInt(gx+off[0], 0, gw-1)
							sy := common.ConstrainInt(gy+off[1], 0, gh-1)
							sIdx := sy*gw + sx
							if idGrid[sIdx] != id0 && zGrid[sIdx] > z0 {
								dz := math.Min(zGrid[sIdx]-z0, 4.5)
								aoSum += common.Constrain(dz/3.0, 0.0, 1.0)
							}
						}
						aoShade := 1.0 - (aoSum/12.0)*0.60

						castOcc := 0.0
						for step := 1; step <= 12; step++ {
							dist := float64(step) * 1.55
							sx := common.ConstrainInt(int(math.Round(float64(gx)-0.62*dist)), 0, gw-1)
							sy := common.ConstrainInt(int(math.Round(float64(gy)-0.78*dist)), 0, gh-1)
							sIdx := sy*gw + sx
							if idGrid[sIdx] != id0 && zGrid[sIdx] > z0 {
								falloff := 1.0 - float64(step-1)/12.0
								castOcc = math.Max(castOcc, falloff*0.52)
							}
						}
						castShade := 1.0 - castOcc

						ux := (float64(gx) * 0.5) / fw
						uy := (float64(gy) * 0.5) / fh
						macroLight := 0.82 + 0.26*common.Smoothstep(1.10, -0.05, uy*0.60-ux*0.28)

						shade := aoShade * castShade * macroLight
						rOut := albRGrid[idx]*shade + rimGrid[idx]*0.16
						gOut := albGGrid[idx]*shade + rimGrid[idx]*0.13
						bOut := albBGrid[idx]*shade + rimGrid[idx]*0.06

						sumR += common.Clamp01(rOut)
						sumG += common.Clamp01(gOut)
						sumB += common.Clamp01(bOut)
					}

					rc.SetPixelLinear(x, y, sumR*0.25, sumG*0.25, sumB*0.25, 1.0)
				}
			}
		}(y0, y1)
	}
	wgShade.Wait()

	c.MergeRenderCanvas(rc)
}
