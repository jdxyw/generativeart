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

type perlinPlumeTapestry struct {
	plumes    int
	barbs     int
	warpScale float64
}

// NewPerlinPlumeTapestry returns a PerlinPlumeTapestry generator that synthesizes
// a full-bleed 3D shadow-box paper-cut & gold-leaf Perlin cloud, alpine ridge, and
// ocean-wave scroll: 22 overlapping upper-sky cloud draperies and lower-horizon
// terraced wave strata framing a recessed 3D golden sun sanctuary, rendered with
// 2x2 supersampling, multi-step directional drop shadows, archival cotton-paper
// micro-grain, and burnished 24K gold-leaf cut rims.
func NewPerlinPlumeTapestry(plumes, barbs int, warpScale float64) *perlinPlumeTapestry {
	if plumes <= 0 {
		plumes = 22
	}
	if barbs <= 0 {
		barbs = 12
	}
	if warpScale <= 0 {
		warpScale = 1.35
	}
	return &perlinPlumeTapestry{
		plumes:    common.ConstrainInt(plumes, 12, 36),
		barbs:     common.ConstrainInt(barbs, 6, 32),
		warpScale: common.Constrain(warpScale, 0.5, 2.5),
	}
}

// Generative renders the static PerlinPlumeTapestry composition.
func (ppt *perlinPlumeTapestry) Generative(c *generativeart.Canva) {
	ppt.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360° looping PerlinPlumeTapestry animation.
func (ppt *perlinPlumeTapestry) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	phi := 2.0 * math.Pi * phase

	seed := c.Seed()
	if seed == 0 {
		seed = 20263012
	}
	rng := rand.New(rand.NewSource(seed))
	ne := common.NewNoiseEngine(seed)

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	wScale := ppt.warpScale
	loopX := 0.12 * math.Cos(phi)
	loopY := 0.12 * math.Sin(phi)

	// =========================================================================
	// 1. DEFINE 3D SHADOW-BOX PAPER-CUT LAYERS (BACK-TO-FRONT DEPTH Z)
	// =========================================================================
	// Layer 0 (deepest background, z = 0): Warm Sunset Ivory & Golden Sky Sanctuary
	// Layer 1..3 (z = 1..3): Recessed 3D Concentric Sun Halo Rings & Blazing Solar Disk
	// Layer 4..10 (z = 4..10): Upper Arching Perlin Cloud-Plume Draperies (hanging from top)
	// Layer 11..23 (z = 11..23): Lower Rolling Perlin Mountain & Ocean-Wave Strata (rising to foreground)
	type paperLayer struct {
		id       int
		zHeight  float64
		isUpper  bool // true = hangs down from top (uy <= yEdge), false = rises from bottom (uy >= yEdge)
		yBase    float64
		freqX    float64
		ampY     float64
		archLift float64
		phaseOff float64
		seedOff  float64
		colA     color.RGBA
		colB     color.RGBA
		rimCol   color.RGBA
	}

	// Curated Eiko-Ojala / Ukiyo-e / Dunhuang Shadow-Box Palette:
	// Upper cloud layers progress from warm Apricot-Gold & Coral-Vermilion (deep near the sun)
	// to Cinnabar-Red & Twilight Bordeaux-Plum at the top canopy.
	upperCols := [][2]color.RGBA{
		{{R: 252, G: 212, B: 118, A: 255}, {R: 246, G: 168, B: 68, A: 255}}, // 0: Warm Champagne Apricot
		{{R: 248, G: 162, B: 64, A: 255}, {R: 238, G: 112, B: 52, A: 255}},  // 1: Sunset Marigold-Tangerine
		{{R: 238, G: 106, B: 54, A: 255}, {R: 214, G: 58, B: 52, A: 255}},   // 2: Terracotta Vermilion
		{{R: 208, G: 48, B: 56, A: 255}, {R: 168, G: 30, B: 64, A: 255}},    // 3: Cinnabar Crimson
		{{R: 158, G: 28, B: 68, A: 255}, {R: 112, G: 24, B: 78, A: 255}},    // 4: Bordeaux Carmine-Wine
		{{R: 96, G: 26, B: 88, A: 255}, {R: 54, G: 24, B: 82, A: 255}},      // 5: Twilight Imperial Plum
	}

	// Lower mountain & wave layers progress from distant Golden-Ochre & Coral ridges
	// through Celadon-Jade & Kingfisher-Turquoise to deep Cobalt & Prussian Sapphire waves!
	lowerCols := [][2]color.RGBA{
		{{R: 248, G: 192, B: 78, A: 255}, {R: 236, G: 138, B: 56, A: 255}},  // 0: Distant Sunlit Amber Ridge
		{{R: 236, G: 126, B: 54, A: 255}, {R: 204, G: 68, B: 54, A: 255}},   // 1: Warm Terracotta Foothills
		{{R: 192, G: 58, B: 66, A: 255}, {R: 138, G: 38, B: 82, A: 255}},    // 2: Rose-Cinnabar Escarpment
		{{R: 78, G: 166, B: 142, A: 255}, {R: 38, G: 132, B: 126, A: 255}},  // 3: Mist Celadon-Malachite
		{{R: 36, G: 156, B: 158, A: 255}, {R: 20, G: 116, B: 138, A: 255}},  // 4: Kingfisher Turquoise Lagoon
		{{R: 22, G: 132, B: 156, A: 255}, {R: 16, G: 92, B: 134, A: 255}},   // 5: Peacock Cerulean Wave
		{{R: 18, G: 104, B: 148, A: 255}, {R: 14, G: 68, B: 118, A: 255}},   // 6: Deep Azure Current
		{{R: 244, G: 184, B: 62, A: 255}, {R: 226, G: 118, B: 46, A: 255}},  // 7: Gilded Foam-Crest Accent Wave!
		{{R: 16, G: 82, B: 132, A: 255}, {R: 12, G: 48, B: 96, A: 255}},     // 8: Ultramarine Swell
		{{R: 14, G: 56, B: 108, A: 255}, {R: 10, G: 32, B: 74, A: 255}},     // 9: Prussian Sapphire Trench
		{{R: 10, G: 38, B: 82, A: 255}, {R: 7, G: 20, B: 52, A: 255}},      // 10: Midnight Abyssal Foreground
	}

	goldRim := color.RGBA{R: 255, G: 232, B: 164, A: 255}
	layers := make([]paperLayer, 0, 17)

	// 6 Upper arching Perlin cloud-plume layers (hanging from top, receding toward the central sun window)
	for i := 0; i < 6; i++ {
		fi := float64(i) / 5.0
		// i=0 is deepest (lowest edge near y=0.31); i=5 is highest/closest to viewer near top edge (y=0.07)
		yB := 0.31 - 0.24*fi
		layers = append(layers, paperLayer{
			id:       10 + i,
			zHeight:  4.0 + float64(i)*2.4,
			isUpper:  true,
			yBase:    yB,
			freqX:    1.15 + 0.22*float64(i%3),
			ampY:     (0.048 + 0.014*math.Sin(float64(i)*1.3)) * wScale,
			// Arch upward in the center (ux ≈ 0.5) so the full round solar disk is framed cleanly!
			archLift: -0.13 * (1.0 - 0.30*fi),
			phaseOff: float64(i)*1.45 + 0.4*rng.Float64(),
			seedOff:  17.3*float64(i+1) + rng.Float64()*30.0,
			colA:     upperCols[i][0],
			colB:     upperCols[i][1],
			rimCol:   goldRim,
		})
	}

	// 11 Lower rolling Perlin mountain & ocean-wave strata (rising from y=0.48 down to y=0.95)
	for j := 0; j < len(lowerCols); j++ {
		fj := float64(j) / float64(len(lowerCols)-1)
		yB := 0.48 + 0.47*fj
		// Dip gently in the center for the upper valley layers to cradle the full circular sun
		cradleDip := 0.095 * math.Exp(-fj*2.0)
		layers = append(layers, paperLayer{
			id:       20 + j,
			zHeight:  5.5 + float64(j)*2.2,
			isUpper:  false,
			yBase:    yB,
			freqX:    1.10 + 0.25*math.Sin(float64(j)*1.1),
			ampY:     (0.056 + 0.016*math.Cos(float64(j)*1.4)) * wScale,
			archLift: cradleDip,
			phaseOff: float64(j)*1.55 + 0.5*rng.Float64(),
			seedOff:  31.7*float64(j+1) + rng.Float64()*40.0,
			colA:     lowerCols[j][0],
			colB:     lowerCols[j][1],
			rimCol:   goldRim,
		})
	}

	// Precompute 2x horizontal edge curves yEdge[layerIdx][gx] for fast 2x2 supersampling!
	gw, gh := w*2+1, h*2+1
	nLayers := len(layers)
	yEdges := make([][]float64, nLayers)
	dyDxEdges := make([][]float64, nLayers)
	for li := 0; li < nLayers; li++ {
		yEdges[li] = make([]float64, gw)
		dyDxEdges[li] = make([]float64, gw)
		ly := &layers[li]
		for gx := 0; gx < gw; gx++ {
			ux := (float64(gx) * 0.5) / fw
			dxCen := ux - 0.50

			pVal, dPdx, _ := ne.FBM2DDeriv(ux*ly.freqX+ly.seedOff*0.1+loopX*0.35, ly.yBase*2.0+loopY*0.35, 3, 2.0, 0.46)
			w1 := 0.042 * math.Sin(ux*math.Pi*2.0*ly.freqX*0.65+ly.phaseOff+phi)
			dw1 := 0.042 * math.Pi * 2.0 * ly.freqX * 0.65 * math.Cos(ux*math.Pi*2.0*ly.freqX*0.65+ly.phaseOff+phi)

			w2 := 0.022 * math.Cos(ux*math.Pi*3.1-ly.phaseOff*0.8-phi)
			dw2 := -0.022 * math.Pi * 3.1 * math.Sin(ux*math.Pi*3.1-ly.phaseOff*0.8-phi)

			// Smooth central arch/cradle envelope cos^2(dxCen * pi)
			archEnv := math.Cos(dxCen * math.Pi * 0.85)
			archTerm := ly.archLift * archEnv * archEnv
			dArchDx := -ly.archLift * 2.0 * archEnv * math.Sin(dxCen*math.Pi*0.85) * math.Pi * 0.85

			yEdges[li][gx] = ly.yBase + archTerm + ly.ampY*pVal + w1 + w2
			dyDxEdges[li][gx] = dArchDx + ly.ampY*dPdx*ly.freqX + dw1 + dw2
		}
	}

	// =========================================================================
	// 2. EVALUATE 2X SUPERSAMPLED Z-BUFFER, LAYER ID & CUT-EDGE DISTANCE
	// =========================================================================
	zGrid := make([]float64, gw*gh)
	idGrid := make([]int, gw*gh)
	rimGrid := make([]float64, gw*gh)
	albRGrid := make([]float64, gw*gh)
	albGGrid := make([]float64, gw*gh)
	albBGrid := make([]float64, gw*gh)

	// Solar sanctuary center at (0.50, 0.36)
	sunX, sunY := 0.50, 0.36

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

					// Base Layer 0: Warm Golden-Ivory Sky Sanctuary + Recessed Solar Halo Rings
					dSun := math.Hypot(ux-sunX, (uy-sunY)*(fh/fw))
					curZ := 0.5
					curID := 1
					curRim := 0.0

					// Sky background gradient + 3 recessed concentric paper-cut sun rings
					skyT := common.Clamp01(dSun / 0.42)
					curCol := render.LerpOklab(
						color.RGBA{R: 255, G: 244, B: 204, A: 255},
						color.RGBA{R: 250, G: 196, B: 96, A: 255},
						skyT,
					)

					// 3 concentric recessed paper-cut sun halo steps (dSun < 0.22, 0.16, 0.105)
					haloRadii := [3]float64{0.215, 0.155, 0.102}
					haloCols := [3]color.RGBA{
						{R: 252, G: 214, B: 112, A: 255},
						{R: 255, G: 232, B: 148, A: 255},
						{R: 255, G: 248, B: 212, A: 255}, // Blazing Ivory-Gold Solar Core
					}
					for hIdx := 0; hIdx < 3; hIdx++ {
						rHalo := haloRadii[hIdx]
						if dSun <= rHalo {
							curZ = 1.2 + float64(hIdx)*0.85
							curID = 2 + hIdx
							curCol = haloCols[hIdx]
							dEdgePx := (rHalo - dSun) * fw
							if dEdgePx < 1.8 {
								curRim = math.Max(curRim, common.Smoothstep(1.8, 0.3, dEdgePx))
							}
						}
					}

					// Test all 17 Perlin paper-cut cloud and mountain/wave layers
					for li := 0; li < nLayers; li++ {
						ly := &layers[li]
						if ly.zHeight < curZ {
							continue
						}
						yEdge := yEdges[li][gx]
						dydx := dyDxEdges[li][gx]

						var signedDistY float64
						if ly.isUpper {
							signedDistY = yEdge - uy // > 0 inside upper hanging cloud layer
						} else {
							signedDistY = uy - yEdge // > 0 inside lower rising wave layer
						}
						if signedDistY < 0 {
							continue
						}

						// Exact Euclidean pixel distance from the cut paper edge!
						invArc := 1.0 / math.Sqrt(1.0+dydx*dydx)
						distPx := signedDistY * fh * invArc

						// Smooth vertical & Perlin paper gradient within this cut layer
						layerSpan := common.Clamp01(signedDistY / 0.16)
						pShade := 0.5 + 0.5*ne.Perlin2D(ux*3.2+ly.seedOff, uy*3.2)
						pCol := render.LerpOklab(ly.colA, ly.colB, common.Clamp01(0.72*layerSpan+0.28*pShade))

						// Subtle internal combed silk-paper contour ribs parallel to the cut edge!
						ribWave := math.Sin(distPx * 0.65)
						if ribWave > 0 {
							pCol = render.LerpOklab(pCol, color.RGBA{R: 255, G: 246, B: 224, A: 255}, ribWave*0.09*(1.0-0.5*layerSpan))
						} else {
							pCol = render.LerpOklab(pCol, color.RGBA{R: 12, G: 10, B: 24, A: 255}, -ribWave*0.08)
						}

						// 3D beveled cushion lift right along the cut paper edge (distPx < 8px)
						edgeLift := 0.35 * math.Exp(-distPx/6.0)
						curZ = ly.zHeight + edgeLift
						curID = ly.id
						curCol = pCol
						if distPx < 1.65 {
							curRim = common.Smoothstep(1.65, 0.35, distPx)
						} else {
							curRim = 0.0
						}
					}

					// Fine archival cotton-paper micro-grain
					grain := ne.Perlin2D(float64(gx)*0.35, float64(gy)*0.35) * 0.025
					rL, gL, bL, _ := render.ColorToLinearRGB(curCol)
					if curRim > 0 {
						rr, rg, rb, _ := render.ColorToLinearRGB(goldRim)
						rL = rL*(1.0-curRim*0.78) + rr*curRim*0.78
						gL = gL*(1.0-curRim*0.78) + rg*curRim*0.78
						bL = bL*(1.0-curRim*0.78) + rb*curRim*0.78
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
	// 3. PARALLEL 2X2 SUPERSAMPLED SHADOW-BOX SOFT DROP SHADOWS & AO
	// =========================================================================
	subOffs := [4][2]int{{0, 0}, {1, 0}, {0, 1}, {1, 1}}
	// Directional gallery light from upper-center-left + soft ambient shadow-box occlusion ring
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

						// 1) Soft omnidirectional shadow-box ambient occlusion under overlapping paper edges
						aoSum := 0.0
						for _, off := range aoOffsets {
							sx := common.ConstrainInt(gx+off[0], 0, gw-1)
							sy := common.ConstrainInt(gy+off[1], 0, gh-1)
							sIdx := sy*gw + sx
							if idGrid[sIdx] != id0 && zGrid[sIdx] > z0 {
								dz := math.Min(zGrid[sIdx]-z0, 4.5)
								aoSum += common.Constrain(dz/3.2, 0.0, 1.0)
							}
						}
						aoShade := 1.0 - (aoSum/12.0)*0.58

						// 2) Multi-step directional shadow-box cast shadow (cast away from the central sun / top light!)
						ux := (float64(gx) * 0.5) / fw
						uy := (float64(gy) * 0.5) / fh
						// Light vector points toward the solar core (0.50, 0.36) blended with top-left key light
						ldx := 0.45*(sunX-ux) - 0.35
						ldy := 0.45*(sunY-uy) - 0.55
						lNorm := math.Hypot(ldx, ldy)
						if lNorm > 1e-6 {
							ldx, ldy = ldx/lNorm, ldy/lNorm
						}

						castOcc := 0.0
						for step := 1; step <= 14; step++ {
							dist := float64(step) * 1.65
							sx := common.ConstrainInt(int(math.Round(float64(gx)+ldx*dist)), 0, gw-1)
							sy := common.ConstrainInt(int(math.Round(float64(gy)+ldy*dist)), 0, gh-1)
							sIdx := sy*gw + sx
							if idGrid[sIdx] != id0 && zGrid[sIdx] > z0 {
								falloff := 1.0 - float64(step-1)/14.0
								castOcc = math.Max(castOcc, falloff*0.54)
							}
						}
						castShade := 1.0 - castOcc

						// Warm solar radiance glow near the sun window
						dSun := math.Hypot(ux-sunX, (uy-sunY)*(fh/fw))
						sunBloom := 0.14 * math.Exp(-(dSun*dSun)/0.085)

						shade := aoShade * castShade
						rOut := albRGrid[idx]*shade + sunBloom*1.05 + rimGrid[idx]*0.14
						gOut := albGGrid[idx]*shade + sunBloom*0.82 + rimGrid[idx]*0.11
						bOut := albBGrid[idx]*shade + sunBloom*0.42 + rimGrid[idx]*0.06

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
