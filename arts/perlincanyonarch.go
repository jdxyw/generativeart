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

type perlinCanyonArch struct {
	rings     int
	ripples   int
	warpScale float64
}

// NewPerlinCanyonArch returns a PerlinCanyonArch generator that synthesizes a
// full-bleed 2x2-supersampled 3D shadow-box paper-cut & gold-leaf Perlin Canyon
// Portal, Solar Sanctuary & Turquoise Lagoon Scroll: 6 receding concentric
// Perlin-warped sandstone canyon arch frames opening onto a 3-tier golden sun
// halo and 9 terraced emerald/turquoise/sapphire lagoon waves, with multi-step
// directional drop shadows, cotton-paper ribs, and 24K gold-leaf cut rims.
func NewPerlinCanyonArch(rings, ripples int, warpScale float64) *perlinCanyonArch {
	if rings <= 0 {
		rings = 18
	}
	if ripples <= 0 {
		ripples = 14
	}
	if warpScale <= 0 {
		warpScale = 1.35
	}
	return &perlinCanyonArch{
		rings:     common.ConstrainInt(rings, 10, 32),
		ripples:   common.ConstrainInt(ripples, 6, 28),
		warpScale: common.Constrain(warpScale, 0.5, 2.5),
	}
}

// Generative renders the static PerlinCanyonArch composition.
func (pca *perlinCanyonArch) Generative(c *generativeart.Canva) {
	pca.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360° looping PerlinCanyonArch animation.
func (pca *perlinCanyonArch) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	phi := 2.0 * math.Pi * phase

	seed := c.Seed()
	if seed == 0 {
		seed = 20263101
	}
	rng := rand.New(rand.NewSource(seed))
	ne := common.NewNoiseEngine(seed)

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	wScale := pca.warpScale
	loopX := 0.28 * (math.Cos(phi) - 1.0)
	loopY := 0.28 * math.Sin(phi)

	goldRim := color.RGBA{R: 255, G: 234, B: 168, A: 255}

	// =========================================================================
	// 1. DEFINE LAGOON WAVE STRATA (Z = 4.0 .. 18.0, PLUS 2 FOREGROUND WAVES)
	// =========================================================================
	type waveLayer struct {
		id       int
		zHeight  float64
		yBase    float64
		freqX    float64
		ampY     float64
		cradle   float64
		phaseOff float64
		seedOff  float64
		dir      float64
		colA     color.RGBA
		colB     color.RGBA
	}

	waveCols := [][2]color.RGBA{
		{{R: 248, G: 188, B: 74, A: 255}, {R: 234, G: 126, B: 52, A: 255}}, // 0: Distant Sunlit Amber Foothill
		{{R: 228, G: 102, B: 54, A: 255}, {R: 186, G: 52, B: 62, A: 255}},  // 1: Terracotta Coral Ridge
		{{R: 72, G: 172, B: 142, A: 255}, {R: 34, G: 136, B: 126, A: 255}}, // 2: Celadon Jade Shallows
		{{R: 32, G: 158, B: 158, A: 255}, {R: 18, G: 118, B: 138, A: 255}}, // 3: Kingfisher Turquoise Lagoon
		{{R: 22, G: 134, B: 156, A: 255}, {R: 16, G: 94, B: 136, A: 255}},  // 4: Peacock Cerulean Wave
		{{R: 246, G: 188, B: 64, A: 255}, {R: 228, G: 122, B: 46, A: 255}}, // 5: Gilded Sun-Reflection Crest!
		{{R: 18, G: 106, B: 148, A: 255}, {R: 14, G: 68, B: 118, A: 255}},  // 6: Deep Azure Current
		{{R: 14, G: 72, B: 124, A: 255}, {R: 10, G: 42, B: 88, A: 255}},    // 7: Foreground Ultramarine Swell
		{{R: 10, G: 44, B: 92, A: 255}, {R: 6, G: 22, B: 56, A: 255}},      // 8: Foreground Prussian Abyss
	}

	nWaves := len(waveCols)
	waves := make([]waveLayer, nWaves)
	for j := 0; j < nWaves; j++ {
		fj := float64(j) / float64(nWaves-1)
		yB := 0.50 + 0.44*fj
		zH := 4.5 + float64(j)*1.6
		if j >= 7 {
			// Last 2 waves sit in the very foreground in front of the canyon arch base!
			zH = 26.5 + float64(j-7)*2.2
		}
		dir := 1.0
		if j%2 == 1 {
			dir = -1.0
		}
		waves[j] = waveLayer{
			id:       10 + j,
			zHeight:  zH,
			yBase:    yB,
			freqX:    1.15 + 0.24*math.Sin(float64(j)*1.2),
			ampY:     (0.048 + 0.015*math.Cos(float64(j)*1.3)) * wScale,
			cradle:   0.065 * math.Exp(-fj*1.8),
			phaseOff: float64(j)*1.52 + 0.4*rng.Float64(),
			seedOff:  23.7*float64(j+1) + rng.Float64()*35.0,
			dir:      dir,
			colA:     waveCols[j][0],
			colB:     waveCols[j][1],
		}
	}

	gw, gh := w*2+1, h*2+1
	waveY := make([][]float64, nWaves)
	waveDyDx := make([][]float64, nWaves)
	for j := 0; j < nWaves; j++ {
		waveY[j] = make([]float64, gw)
		waveDyDx[j] = make([]float64, gw)
		wv := &waves[j]
		for gx := 0; gx < gw; gx++ {
			ux := (float64(gx) * 0.5) / fw
			dxCen := ux - 0.50

			pVal, dPdx, _ := ne.FBM2DDeriv(ux*wv.freqX+wv.seedOff*0.1+loopX*0.35*wv.dir, wv.yBase*2.0+loopY*0.35, 3, 2.0, 0.46)
			w1 := 0.038 * math.Sin(ux*math.Pi*2.0*wv.freqX*0.65+wv.phaseOff+wv.dir*phi)
			dw1 := 0.038 * math.Pi * 2.0 * wv.freqX * 0.65 * math.Cos(ux*math.Pi*2.0*wv.freqX*0.65+wv.phaseOff+wv.dir*phi)

			w2 := 0.020 * math.Cos(ux*math.Pi*3.2-wv.phaseOff*0.8-wv.dir*phi)
			dw2 := -0.020 * math.Pi * 3.2 * math.Sin(ux*math.Pi*3.2-wv.phaseOff*0.8-wv.dir*phi)

			archEnv := math.Cos(dxCen * math.Pi * 0.85)
			cradleTerm := wv.cradle * archEnv * archEnv
			dCradleDx := -wv.cradle * 2.0 * archEnv * math.Sin(dxCen*math.Pi*0.85) * math.Pi * 0.85

			waveY[j][gx] = wv.yBase + cradleTerm + wv.ampY*pVal + w1 + w2
			waveDyDx[j][gx] = dCradleDx + wv.ampY*dPdx*wv.freqX + dw1 + dw2
		}
	}

	// =========================================================================
	// 2. DEFINE 6 CONCENTRIC 3D PAPER-CUT CANYON ARCH FRAMES (Z = 12.0 .. 24.5)
	// =========================================================================
	type canyonArch struct {
		id       int
		zHeight  float64
		rOpen    float64
		phaseOff float64
		seedOff  float64
		dir      float64
		colA     color.RGBA
		colB     color.RGBA
	}

	archCols := [6][2]color.RGBA{
		{{R: 250, G: 196, B: 88, A: 255}, {R: 242, G: 142, B: 58, A: 255}}, // 0 (innermost arch): Sunlit Apricot-Gold
		{{R: 242, G: 136, B: 56, A: 255}, {R: 218, G: 76, B: 52, A: 255}},  // 1: Sunset Tangerine-Coral
		{{R: 214, G: 64, B: 54, A: 255}, {R: 176, G: 34, B: 60, A: 255}},   // 2: Terracotta Cinnabar
		{{R: 168, G: 30, B: 64, A: 255}, {R: 122, G: 24, B: 76, A: 255}},   // 3: Alizarin Bordeaux Wine
		{{R: 112, G: 26, B: 84, A: 255}, {R: 68, G: 24, B: 92, A: 255}},    // 4: Imperial Amethyst Plum
		{{R: 56, G: 24, B: 86, A: 255}, {R: 24, G: 18, B: 58, A: 255}},     // 5 (outermost foreground): Nocturne Indigo
	}

	const nArches = 6
	arches := make([]canyonArch, nArches)
	for k := 0; k < nArches; k++ {
		fk := float64(k) / float64(nArches-1)
		dir := 1.0
		if k%2 == 1 {
			dir = -1.0
		}
		arches[k] = canyonArch{
			id:       30 + k,
			zHeight:  12.5 + float64(k)*2.3,
			// Opening radius expands from 0.34 (innermost) to 0.64 (outermost)
			rOpen:    0.34 + 0.30*fk,
			phaseOff: float64(k)*1.15 + 0.35*rng.Float64(),
			seedOff:  41.3*float64(k+1) + rng.Float64()*40.0,
			dir:      dir,
			colA:     archCols[k][0],
			colB:     archCols[k][1],
		}
	}

	// =========================================================================
	// 3. RASTERIZE 2X SUPERSAMPLED Z-BUFFER, LAYER ID & GOLD CUT RIMS
	// =========================================================================
	zGrid := make([]float64, gw*gh)
	idGrid := make([]int, gw*gh)
	rimGrid := make([]float64, gw*gh)
	albRGrid := make([]float64, gw*gh)
	albGGrid := make([]float64, gw*gh)
	albBGrid := make([]float64, gw*gh)

	sunX, sunY := 0.50, 0.34
	archCX, archCY := 0.50, 0.49

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

					// 1) Base Sky Sanctuary & 3-Tier Recessed Solar Halo
					dSun := math.Hypot(ux-sunX, (uy-sunY)*(fh/fw))
					curZ := 0.5
					curID := 1
					curRim := 0.0

					skyT := common.Clamp01(dSun / 0.44)
					curCol := render.LerpOklab(
						color.RGBA{R: 255, G: 246, B: 208, A: 255},
						color.RGBA{R: 250, G: 192, B: 92, A: 255},
						skyT,
					)

					haloRadii := [3]float64{0.195, 0.140, 0.092}
					haloCols := [3]color.RGBA{
						{R: 252, G: 214, B: 112, A: 255},
						{R: 255, G: 234, B: 152, A: 255},
						{R: 255, G: 250, B: 218, A: 255},
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

					// 2) Evaluate 6 Concentric 3D Paper-Cut Canyon Arch Portal Frames
					adx := (ux - archCX) / 0.84
					ady := (uy - archCY) / 0.96
					rArch := math.Hypot(adx, ady)
					theta := math.Atan2(ady, adx)
					cosTh, sinTh := math.Cos(theta), math.Sin(theta)

					for k := 0; k < nArches; k++ {
						ar := &arches[k]
						if ar.zHeight < curZ {
							continue
						}
						// Seamless 360-degree periodic Perlin evaluation on the unit circle!
						pVal, dEdx, dEdy := ne.FBM2DDeriv(
							cosTh*1.10+ar.seedOff*0.07+loopX*0.35*ar.dir,
							sinTh*1.10-ar.seedOff*0.05+loopY*0.35,
							2, 2.0, 0.45,
						)
						dPdTheta := 1.10 * (-sinTh*dEdx + cosTh*dEdy)

						// Natural arch profile: slightly wider at bottom-sides and arched at crown
						wTheta := 0.032*math.Sin(3.0*theta+ar.phaseOff+ar.dir*phi) +
							0.022*math.Cos(2.0*theta-0.5)
						dwTheta := 0.032*3.0*math.Cos(3.0*theta+ar.phaseOff+ar.dir*phi) -
							0.022*2.0*math.Sin(2.0*theta-0.5)

						rBoundary := ar.rOpen + 0.042*wScale*pVal + wTheta
						drdTheta := 0.042*wScale*dPdTheta + dwTheta

						signedR := rArch - rBoundary // > 0 inside the solid canyon arch frame!
						if signedR < 0 {
							continue
						}

						slopeAng := 0.0
						if rArch > 0.05 {
							slopeAng = common.Constrain(drdTheta/rArch, -1.2, 1.2)
						}
						invArc := 1.0 / math.Sqrt(1.0+slopeAng*slopeAng)
						distPx := signedR * 0.88 * fw * invArc

						spanT := common.Clamp01(signedR / 0.16)
						pShade := 0.5 + 0.5*ne.Perlin2D(ux*3.0+ar.seedOff, uy*3.0)
						pCol := render.LerpOklab(ar.colA, ar.colB, common.Clamp01(0.72*spanT+0.28*pShade))

						ribWave := math.Sin(distPx * 0.65)
						if ribWave > 0 {
							pCol = render.LerpOklab(pCol, color.RGBA{R: 255, G: 246, B: 224, A: 255}, ribWave*0.09*(1.0-0.5*spanT))
						} else {
							pCol = render.LerpOklab(pCol, color.RGBA{R: 12, G: 8, B: 22, A: 255}, -ribWave*0.08)
						}

						edgeLift := 0.35 * math.Exp(-distPx/6.0)
						curZ = ar.zHeight + edgeLift
						curID = ar.id
						curCol = pCol
						if distPx < 1.65 {
							curRim = common.Smoothstep(1.65, 0.35, distPx)
						} else {
							curRim = 0.0
						}
					}

					// 3) Evaluate Lagoon & Mountain Wave Layers (midground inside arch + 2 foreground waves)
					for j := 0; j < nWaves; j++ {
						wv := &waves[j]
						if wv.zHeight < curZ {
							continue
						}
						yEdge := waveY[j][gx]
						dydx := waveDyDx[j][gx]
						signedY := uy - yEdge
						if signedY < 0 {
							continue
						}

						invArc := 1.0 / math.Sqrt(1.0+dydx*dydx)
						distPx := signedY * fh * invArc

						spanT := common.Clamp01(signedY / 0.14)
						pShade := 0.5 + 0.5*ne.Perlin2D(ux*3.2+wv.seedOff, uy*3.2)
						pCol := render.LerpOklab(wv.colA, wv.colB, common.Clamp01(0.72*spanT+0.28*pShade))

						ribWave := math.Sin(distPx * 0.65)
						if ribWave > 0 {
							pCol = render.LerpOklab(pCol, color.RGBA{R: 255, G: 246, B: 224, A: 255}, ribWave*0.09*(1.0-0.5*spanT))
						} else {
							pCol = render.LerpOklab(pCol, color.RGBA{R: 10, G: 12, B: 26, A: 255}, -ribWave*0.08)
						}

						edgeLift := 0.35 * math.Exp(-distPx/6.0)
						curZ = wv.zHeight + edgeLift
						curID = wv.id
						curCol = pCol
						if distPx < 1.65 {
							curRim = common.Smoothstep(1.65, 0.35, distPx)
						} else {
							curRim = 0.0
						}
					}

					grain := ne.Perlin2D(float64(gx)*0.35, float64(gy)*0.35) * 0.024
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
	// 4. PARALLEL 2X2 SUPERSAMPLED SHADOW-BOX DROP SHADOWS & AO
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
								aoSum += common.Constrain(dz/3.2, 0.0, 1.0)
							}
						}
						aoShade := 1.0 - (aoSum/12.0)*0.58

						ux := (float64(gx) * 0.5) / fw
						uy := (float64(gy) * 0.5) / fh
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
