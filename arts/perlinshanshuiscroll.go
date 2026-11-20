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

type perlinShanshuiScroll struct {
	peaks       int
	riverSteps  int
	canyonScale float64
}

// NewPerlinShanshuiScroll returns a PerlinShanshuiScroll generator that synthesizes
// a full-bleed 2x2-supersampled 3D shadow-box & gilded silk-brocade Qianli Jiangshan
// (千里江山) Winding River Canyon Scroll: 10 interlocking left-bank and right-bank
// karst mountain headlands in Azurite Blue, Malachite Green, Ochre Gold, and Cinnabar
// plunging into a meandering S-curve golden-and-turquoise river valley beneath arching
// sunset cloud scrolls and a 3-tier recessed solar sanctuary.
func NewPerlinShanshuiScroll(peaks, riverSteps int, canyonScale float64) *perlinShanshuiScroll {
	if peaks <= 0 {
		peaks = 10
	}
	if riverSteps <= 0 {
		riverSteps = 12
	}
	if canyonScale <= 0 {
		canyonScale = 1.35
	}
	return &perlinShanshuiScroll{
		peaks:       common.ConstrainInt(peaks, 6, 18),
		riverSteps:  common.ConstrainInt(riverSteps, 6, 24),
		canyonScale: common.Constrain(canyonScale, 0.5, 2.5),
	}
}

// Generative renders the static PerlinShanshuiScroll composition.
func (pss *perlinShanshuiScroll) Generative(c *generativeart.Canva) {
	pss.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360° looping PerlinShanshuiScroll animation.
func (pss *perlinShanshuiScroll) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	phi := 2.0 * math.Pi * phase

	seed := c.Seed()
	if seed == 0 {
		seed = 20263203
	}
	rng := rand.New(rand.NewSource(seed))
	ne := common.NewNoiseEngine(seed)

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	cScale := pss.canyonScale
	loopX := 0.28 * (math.Cos(phi) - 1.0)
	loopY := 0.28 * math.Sin(phi)
	goldRim := color.RGBA{R: 255, G: 234, B: 164, A: 255}

	gw, gh := w*2+1, h*2+1

	// =========================================================================
	// 1. PRECOMPUTE WINDING S-RIVER SPINE xRiver(gy) FROM HORIZON TO FOREGROUND
	// =========================================================================
	riverCenter := make([]float64, gh)
	riverHalfW := make([]float64, gh)
	for gy := 0; gy < gh; gy++ {
		uy := (float64(gy) * 0.5) / fh
		// tRiver goes from 0 at horizon pass (uy = 0.36) to 1 at foreground bottom (uy = 1.0)
		tRiver := common.Clamp01((uy - 0.34) / 0.66)
		// Grand 1.5-cycle meandering S-curve river spine!
		meander := 0.20 * math.Sin(tRiver*math.Pi*2.15-0.35+0.35*math.Sin(phi)) * (0.45 + 0.55*tRiver)
		riverCenter[gy] = 0.50 + meander
		// Perspective river width: narrow gorge (0.045) at horizon -> wide estuary (0.24) at foreground
		riverHalfW[gy] = 0.042 + 0.20*math.Pow(tRiver, 1.15)
	}

	// =========================================================================
	// 2. DEFINE UPPER SKY CLOUD DRAPERIES, RIVER TERRACES & KARST HEADLANDS
	// =========================================================================
	type cloudLayer struct {
		id       int
		zHeight  float64
		yBase    float64
		freqX    float64
		ampY     float64
		archLift float64
		phaseOff float64
		seedOff  float64
		dir      float64
		colA     color.RGBA
		colB     color.RGBA
	}

	schema := c.Opts().ColorSchema()
	cloudCols := render.ResolvePalettePairs(schema, common.Dunhuang, [][2]color.RGBA{
		{{R: 252, G: 210, B: 112, A: 255}, {R: 244, G: 158, B: 62, A: 255}},
		{{R: 244, G: 142, B: 58, A: 255}, {R: 218, G: 76, B: 52, A: 255}},
		{{R: 204, G: 52, B: 58, A: 255}, {R: 148, G: 28, B: 72, A: 255}},
		{{R: 112, G: 26, B: 84, A: 255}, {R: 48, G: 22, B: 78, A: 255}},
	})
	clouds := make([]cloudLayer, 4)
	cloudY := make([][]float64, 4)
	for i := 0; i < 4; i++ {
		fi := float64(i) / 3.0
		dir := 1.0
		if i%2 == 1 {
			dir = -1.0
		}
		clouds[i] = cloudLayer{
			id:       10 + i,
			zHeight:  5.0 + float64(i)*1.8,
			yBase:    0.27 - 0.19*fi,
			freqX:    1.15 + 0.20*float64(i),
			ampY:     0.042 * cScale,
			archLift: -0.11 * (1.0 - 0.25*fi),
			phaseOff: float64(i)*1.4 + 0.3*rng.Float64(),
			seedOff:  17.3*float64(i+1) + rng.Float64()*30.0,
			dir:      dir,
			colA:     cloudCols[i][0],
			colB:     cloudCols[i][1],
		}
		cloudY[i] = make([]float64, gw)
		cl := &clouds[i]
		for gx := 0; gx < gw; gx++ {
			ux := (float64(gx) * 0.5) / fw
			dxCen := ux - 0.50
			pVal, _, _ := ne.FBM2DDeriv(ux*cl.freqX+cl.seedOff*0.1+loopX*0.3*cl.dir, cl.yBase*2.0+loopY*0.3, 2, 2.0, 0.45)
			w1 := 0.032 * math.Sin(ux*math.Pi*2.0*cl.freqX*0.65+cl.phaseOff+cl.dir*phi)
			archEnv := math.Cos(dxCen * math.Pi * 0.85)
			cloudY[i][gx] = cl.yBase + cl.archLift*archEnv*archEnv + cl.ampY*pVal + w1
		}
	}

	// 8 Stepped River Rapid & Golden Reflection Terraces in the valley floor (z = 3.0 .. 10.0)
	type riverStep struct {
		id       int
		zHeight  float64
		yBase    float64
		phaseOff float64
		seedOff  float64
		dir      float64
		colA     color.RGBA
		colB     color.RGBA
	}
	riverCols := render.ResolvePalettePairs(schema, common.Dunhuang, [][2]color.RGBA{
		{{R: 252, G: 214, B: 108, A: 255}, {R: 242, G: 162, B: 64, A: 255}}, // Distant Sunlit Golden Shallows
		{{R: 88, G: 186, B: 156, A: 255}, {R: 42, G: 152, B: 142, A: 255}},  // Mist Celadon Rapid
		{{R: 248, G: 196, B: 78, A: 255}, {R: 228, G: 138, B: 52, A: 255}},  // Golden Sun-Reflection Cascade
		{{R: 34, G: 164, B: 162, A: 255}, {R: 20, G: 126, B: 146, A: 255}},  // Kingfisher Turquoise Pool
		{{R: 24, G: 142, B: 160, A: 255}, {R: 18, G: 102, B: 138, A: 255}},  // Peacock Cerulean Current
		{{R: 244, G: 184, B: 64, A: 255}, {R: 218, G: 118, B: 46, A: 255}},  // Shimmering Gold Ripple Step
		{{R: 18, G: 108, B: 148, A: 255}, {R: 14, G: 68, B: 118, A: 255}},   // Deep Azure Estuary
		{{R: 14, G: 72, B: 124, A: 255}, {R: 10, G: 38, B: 84, A: 255}},     // Foreground Sapphire Deep
	})
	rSteps := make([]riverStep, 8)
	for j := 0; j < 8; j++ {
		fj := float64(j) / 7.0
		dir := 1.0
		if j%2 == 1 {
			dir = -1.0
		}
		rSteps[j] = riverStep{
			id:       20 + j,
			zHeight:  3.5 + float64(j)*1.1,
			yBase:    0.37 + 0.56*fj,
			phaseOff: float64(j)*1.5 + 0.3*rng.Float64(),
			seedOff:  29.3*float64(j+1) + rng.Float64()*30.0,
			dir:      dir,
			colA:     riverCols[j][0],
			colB:     riverCols[j][1],
		}
	}

	// 10 Interlocking Left-Bank & Right-Bank Song-Dynasty Karst Mountain Headlands (z = 11.5 .. 30.0)
	// Alternating Left Bank (side = -1) and Right Bank (side = +1) from distant horizon to foreground!
	type karstHeadland struct {
		id       int
		side     int // -1 = Left Bank, +1 = Right Bank
		zHeight  float64
		yRidge   float64 // Vertical base of the mountain peak silhouette
		yFoot    float64 // Where the headland base sits along the river valley
		peakX    float64 // Horizontal position of main karst summit
		seedOff  float64
		phaseOff float64
		dir      float64
		colPeak  color.RGBA // Summit Azurite / Cinnabar / Gold glaze
		colMid   color.RGBA // Mid-slope Malachite / Turquoise glaze
		colBase  color.RGBA // Deep indigo/umber cliff shadow base
	}

	// Song-Dynasty Qinglü Shanshui (千里江山) Mineral Palette:
	// Distant peaks: Warm Ochre-Gold & Cinnabar Coral -> Midground: Malachite Green & Kingfisher Teal -> Foreground: Azurite Sapphire & Gold
	headlandCols := render.ResolvePaletteTriples(schema, common.Dunhuang, [][3]color.RGBA{
		{{R: 248, G: 186, B: 68, A: 255}, {R: 224, G: 106, B: 52, A: 255}, {R: 138, G: 38, B: 64, A: 255}},  // 0 (Left distant): Ochre Gold -> Terracotta
		{{R: 244, G: 162, B: 62, A: 255}, {R: 198, G: 68, B: 56, A: 255}, {R: 112, G: 32, B: 74, A: 255}},   // 1 (Right distant): Amber -> Cinnabar Ruby
		{{R: 246, G: 202, B: 84, A: 255}, {R: 54, G: 168, B: 128, A: 255}, {R: 18, G: 84, B: 102, A: 255}},  // 2 (Left mid-far): Gold-Tipped Malachite
		{{R: 242, G: 192, B: 76, A: 255}, {R: 32, G: 154, B: 146, A: 255}, {R: 16, G: 74, B: 112, A: 255}},  // 3 (Right mid-far): Gold-Tipped Kingfisher
		{{R: 232, G: 88, B: 52, A: 255}, {R: 176, G: 36, B: 64, A: 255}, {R: 78, G: 24, B: 82, A: 255}},     // 4 (Left mid): Sunset Vermilion Escarpment
		{{R: 62, G: 182, B: 136, A: 255}, {R: 24, G: 128, B: 138, A: 255}, {R: 14, G: 54, B: 96, A: 255}},   // 5 (Right mid): Mineral Malachite & Teal
		{{R: 248, G: 196, B: 68, A: 255}, {R: 26, G: 136, B: 158, A: 255}, {R: 14, G: 48, B: 108, A: 255}},  // 6 (Left near): Gilded Azurite-Cyan Peak
		{{R: 244, G: 178, B: 58, A: 255}, {R: 38, G: 156, B: 124, A: 255}, {R: 12, G: 44, B: 88, A: 255}},   // 7 (Right near): Gilded Celadon-Jade Cliff
		{{R: 32, G: 118, B: 164, A: 255}, {R: 18, G: 64, B: 128, A: 255}, {R: 10, G: 26, B: 68, A: 255}},    // 8 (Left foreground): Deep Azurite Blue
		{{R: 28, G: 96, B: 152, A: 255}, {R: 16, G: 48, B: 108, A: 255}, {R: 8, G: 20, B: 54, A: 255}},      // 9 (Right foreground): Nocturne Lapis Cliff
	})

	const nHeadlands = 10
	headlands := make([]karstHeadland, nHeadlands)
	for m := 0; m < nHeadlands; m++ {
		pairIdx := m / 2 // 0 .. 4 from horizon to foreground
		fPair := float64(pairIdx) / 4.0
		side := -1
		peakX := 0.16 + 0.08*math.Sin(float64(m)*1.7)
		if m%2 == 1 {
			side = 1
			peakX = 0.84 - 0.08*math.Cos(float64(m)*1.7)
		}
		dir := float64(side)
		headlands[m] = karstHeadland{
			id:       40 + m,
			side:     side,
			zHeight:  12.0 + float64(m)*2.0,
			yRidge:   0.30 + 0.44*fPair,
			yFoot:    0.48 + 0.56*fPair,
			peakX:    peakX,
			seedOff:  37.1*float64(m+1) + rng.Float64()*40.0,
			phaseOff: float64(m)*1.3 + 0.4*rng.Float64(),
			dir:      dir,
			colPeak:  headlandCols[m][0],
			colMid:   headlandCols[m][1],
			colBase:  headlandCols[m][2],
		}
	}

	// =========================================================================
	// 3. RASTERIZE 2X SUPERSAMPLED Z-BUFFER, SHANSHUI CLIFFS & WINDING RIVER
	// =========================================================================
	zGrid := make([]float64, gw*gh)
	idGrid := make([]int, gw*gh)
	rimGrid := make([]float64, gw*gh)
	albRGrid := make([]float64, gw*gh)
	albGGrid := make([]float64, gw*gh)
	albBGrid := make([]float64, gw*gh)

	sunX, sunY := 0.50, 0.25

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
				rCen := riverCenter[gy]
				rHalf := riverHalfW[gy]

				for gx := 0; gx < gw; gx++ {
					ux := (float64(gx) * 0.5) / fw
					idx := rowOff + gx

					// 1) Base Sky Sanctuary & 3-Tier Recessed Solar Halo at (0.50, 0.25)
					dSun := math.Hypot(ux-sunX, (uy-sunY)*(fh/fw))
					curZ := 0.5
					curID := 1
					curRim := 0.0

					skyT := common.Clamp01(dSun / 0.46)
					curCol := render.LerpOklab(
						color.RGBA{R: 255, G: 246, B: 208, A: 255},
						color.RGBA{R: 250, G: 192, B: 92, A: 255},
						skyT,
					)

					haloRadii := [3]float64{0.165, 0.118, 0.076}
					haloCols := [3]color.RGBA{
						{R: 252, G: 214, B: 112, A: 255},
						{R: 255, G: 234, B: 152, A: 255},
						{R: 255, G: 250, B: 218, A: 255},
					}
					for hIdx := 0; hIdx < 3; hIdx++ {
						rHalo := haloRadii[hIdx]
						if dSun <= rHalo {
							curZ = 1.2 + float64(hIdx)*0.8
							curID = 2 + hIdx
							curCol = haloCols[hIdx]
							dEdgePx := (rHalo - dSun) * fw
							if dEdgePx < 1.8 {
								curRim = math.Max(curRim, common.Smoothstep(1.8, 0.3, dEdgePx))
							}
						}
					}

					// 2) Upper Sky Arching Cloud-Scroll Draperies
					for i := 0; i < 4; i++ {
						cl := &clouds[i]
						if cl.zHeight < curZ {
							continue
						}
						signedY := cloudY[i][gx] - uy
						if signedY < 0 {
							continue
						}
						distPx := signedY * fh * 0.92
						spanT := common.Clamp01(signedY / 0.14)
						pCol := render.LerpOklab(cl.colA, cl.colB, spanT)
						rib := math.Sin(distPx * 0.65)
						if rib > 0 {
							pCol = render.LerpOklab(pCol, color.RGBA{R: 255, G: 246, B: 224, A: 255}, rib*0.09)
						} else {
							pCol = render.LerpOklab(pCol, color.RGBA{R: 12, G: 8, B: 22, A: 255}, -rib*0.08)
						}
						curZ = cl.zHeight + 0.35*math.Exp(-distPx/6.0)
						curID = cl.id
						curCol = pCol
						if distPx < 1.65 {
							curRim = common.Smoothstep(1.65, 0.35, distPx)
						} else {
							curRim = 0.0
						}
					}

					// 3) Winding S-River Terraced Rapids & Golden Reflections
					for j := 0; j < 8; j++ {
						rs := &rSteps[j]
						if rs.zHeight < curZ {
							continue
						}
						// River waves ripple across the S-channel centered at rCen
						dxRiv := ux - rCen
						pRiv := ne.Perlin2D(ux*2.4+rs.seedOff*0.1+loopX*0.35*rs.dir, rs.yBase*2.0+loopY*0.35)
						yEdge := rs.yBase + 0.024*math.Cos(dxRiv*math.Pi*4.5+rs.phaseOff+rs.dir*phi) + 0.022*pRiv
						signedY := uy - yEdge
						if signedY < 0 {
							continue
						}
						distPx := signedY * fh * 0.90
						spanT := common.Clamp01(signedY / 0.12)
						// Brighten golden sun reflection along the center of the winding river (|dxRiv| < rHalf*0.45)!
						sunRefl := math.Exp(-math.Pow(dxRiv/math.Max(0.04, rHalf*0.55), 2.0))
						pCol := render.LerpOklab(rs.colA, rs.colB, spanT*(1.0-0.45*sunRefl))
						if sunRefl > 0.25 {
							pCol = render.LerpOklab(pCol, color.RGBA{R: 255, G: 226, B: 118, A: 255}, (sunRefl-0.25)*0.55)
						}
						rib := math.Sin(distPx * 0.72)
						if rib > 0 {
							pCol = render.LerpOklab(pCol, color.RGBA{R: 255, G: 248, B: 224, A: 255}, rib*0.11)
						} else {
							pCol = render.LerpOklab(pCol, color.RGBA{R: 8, G: 14, B: 32, A: 255}, -rib*0.10)
						}
						curZ = rs.zHeight + 0.32*math.Exp(-distPx/5.5)
						curID = rs.id
						curCol = pCol
						if distPx < 1.60 {
							curRim = common.Smoothstep(1.60, 0.35, distPx)
						} else {
							curRim = 0.0
						}
					}

					// 4) 10 Interlocking Left-Bank & Right-Bank Karst Mountain Headlands!
					// Each headland is defined by a single C^infinity smooth ridge-to-river curve yTop(ux)
					// that rises high on its bank and plunges smoothly down into the winding S-river channel!
					for m := 0; m < nHeadlands; m++ {
						hd := &headlands[m]
						if hd.zHeight < curZ {
							continue
						}

						dxPeak := ux - hd.peakX
						peakBell := 0.155 * math.Exp(-(dxPeak*dxPeak)/0.038)
						dPeakBellDx := peakBell * (-2.0 * dxPeak / 0.038)

						// Smooth logistic/tanh river-cliff plunge as ux crosses from the mountain bank into the river channel!
						var bankCoord float64
						var dBankDx float64
						bankEdge := rCen + float64(hd.side)*rHalf*0.48 + 0.028*math.Sin(uy*math.Pi*2.2+hd.phaseOff+hd.dir*phi*0.5)
						if hd.side < 0 {
							// Left-bank mountain: high for ux < bankEdge, plunges downward (larger yTop) for ux > bankEdge
							bankCoord = (ux - bankEdge) / 0.14
							dBankDx = 1.0 / 0.14
						} else {
							// Right-bank mountain: high for ux > bankEdge, plunges downward (larger yTop) for ux < bankEdge
							bankCoord = (bankEdge - ux) / 0.14
							dBankDx = -1.0 / 0.14
						}
						if bankCoord > 1.85 {
							continue
						}
						// Smooth S-plunge: 0 on mountain side (bankCoord < -1), rises smoothly to +0.62 into the river!
						tanhB := math.Tanh(bankCoord * 1.35)
						sech2B := 1.0 - tanhB*tanhB
						cliffPlunge := 0.31 * (1.0 + tanhB)
						dCliffDx := 0.31 * 1.35 * sech2B * dBankDx

						pKarst, dKdx, _ := ne.FBM2DDeriv(ux*1.65+hd.seedOff*0.1+loopX*0.25*hd.dir, hd.yRidge*2.0+loopY*0.25, 2, 2.0, 0.45)
						yTop := hd.yRidge - peakBell + cliffPlunge + 0.042*cScale*pKarst

						signedY := uy - yTop
						if signedY < 0 {
							continue
						}

						dydx := -dPeakBellDx + dCliffDx + 0.042*cScale*dKdx*1.65
						invArc := 1.0 / math.Sqrt(1.0+dydx*dydx)
						distPx := signedY * fh * invArc

						// Song-Dynasty 3-stage vertical mineral glaze (Gold/Ochre Summit -> Malachite/Azurite Mid -> Indigo Base)
						elevT := common.Clamp01(signedY / 0.22)
						var pCol color.RGBA
						if elevT < 0.38 {
							pCol = render.LerpOklab(hd.colPeak, hd.colMid, common.Smoothstep(0.0, 0.38, elevT))
						} else {
							pCol = render.LerpOklab(hd.colMid, hd.colBase, common.Smoothstep(0.38, 1.0, elevT))
						}

						// Combed silk & axe-cut cliff striations following the continuous mountain ridge contour!
						rib := math.Sin(distPx * 0.68)
						if rib > 0 {
							pCol = render.LerpOklab(pCol, color.RGBA{R: 255, G: 246, B: 218, A: 255}, rib*0.13*(1.0-0.4*elevT))
						} else {
							pCol = render.LerpOklab(pCol, color.RGBA{R: 8, G: 10, B: 24, A: 255}, -rib*0.12)
						}

						// Inner 24K gold pinstripe at distPx ≈ 9.5px
						innerPin := math.Exp(-math.Pow((distPx-9.5)/1.3, 2.0))
						if innerPin > 0.02 {
							pCol = render.LerpOklab(pCol, color.RGBA{R: 244, G: 188, B: 58, A: 255}, innerPin*0.75)
						}

						curZ = hd.zHeight + 0.40*math.Exp(-distPx/6.0)
						curID = hd.id
						curCol = pCol
						if distPx < 1.70 {
							curRim = common.Smoothstep(1.70, 0.35, distPx)
						} else {
							curRim = 0.0
						}
					}

					grain := ne.Perlin2D(float64(gx)*0.35, float64(gy)*0.35) * 0.022
					rL, gL, bL, _ := render.ColorToLinearRGB(curCol)
					if curRim > 0 {
						rr, rg, rb, _ := render.ColorToLinearRGB(goldRim)
						rL = rL*(1.0-curRim*0.80) + rr*curRim*0.80
						gL = gL*(1.0-curRim*0.80) + rg*curRim*0.80
						bL = bL*(1.0-curRim*0.80) + rb*curRim*0.80
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
						aoShade := 1.0 - (aoSum/12.0)*0.60

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
