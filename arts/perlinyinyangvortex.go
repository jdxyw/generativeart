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

type perlinYinYangVortex struct {
	arms      int
	grooves   int
	spiralAmp float64
}

// NewPerlinYinYangVortex returns a PerlinYinYangVortex generator that synthesizes
// a full-bleed 2x2-supersampled 3D Shadow-Box & Gilded Silk Taiji Dual-Sanctuary
// Scroll (太极日月双曜织金云海画卷): 8 concentric Solar-Gold/Cinnabar arch rings around
// a recessed upper-left Solar Sanctuary and 8 concentric Lunar-Sapphire/Turquoise
// arch rings around a recessed lower-right Lunar Sanctuary, interlocked across the
// center by 10 sweeping Yin-Yang S-bridge silk-brocade ribbons with 24K gold rims.
func NewPerlinYinYangVortex(arms, grooves int, spiralAmp float64) *perlinYinYangVortex {
	if arms <= 0 {
		arms = 12
	}
	if grooves <= 0 {
		grooves = 54
	}
	if spiralAmp <= 0 {
		spiralAmp = 1.35
	}
	return &perlinYinYangVortex{
		arms:      common.ConstrainInt(arms, 6, 20),
		grooves:   common.ConstrainInt(grooves, 28, 96),
		spiralAmp: common.Constrain(spiralAmp, 0.5, 2.5),
	}
}

// Generative renders the static PerlinYinYangVortex composition.
func (pyy *perlinYinYangVortex) Generative(c *generativeart.Canva) {
	pyy.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360° looping PerlinYinYangVortex animation.
func (pyy *perlinYinYangVortex) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	phi := 2.0 * math.Pi * phase

	seed := c.Seed()
	if seed == 0 {
		seed = 20263201
	}
	rng := rand.New(rand.NewSource(seed))
	ne := common.NewNoiseEngine(seed)

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	sAmp := pyy.spiralAmp
	loopX := 0.28 * (math.Cos(phi) - 1.0)
	loopY := 0.28 * math.Sin(phi)
	goldRim := color.RGBA{R: 255, G: 234, B: 164, A: 255}

	// =========================================================================
	// 1. SOLAR & LUNAR CONCENTRIC VORTEX RINGS + INTERLOCKING TAIJI S-RIBBONS
	// =========================================================================
	// Solar Eye at (0.23, 0.27), Lunar Eye at (0.77, 0.73) — nestled squarely inside the two Yin-Yang S-bays!
	const sunX, sunY = 0.23, 0.27
	const moonX, moonY = 0.77, 0.73

	schema := c.Opts().ColorSchema()
	solarRingCols := render.ResolvePalettePairs(schema, common.Dunhuang, [][2]color.RGBA{
		{{R: 255, G: 222, B: 108, A: 255}, {R: 248, G: 176, B: 58, A: 255}}, // 0: Inner Sunburst Gold
		{{R: 248, G: 166, B: 54, A: 255}, {R: 238, G: 112, B: 48, A: 255}},  // 1: Sunset Tangerine
		{{R: 236, G: 102, B: 48, A: 255}, {R: 212, G: 52, B: 52, A: 255}},   // 2: Fiery Coral-Vermilion
		{{R: 204, G: 44, B: 56, A: 255}, {R: 164, G: 28, B: 66, A: 255}},    // 3: Cinnabar Ruby
		{{R: 154, G: 26, B: 72, A: 255}, {R: 108, G: 24, B: 84, A: 255}},    // 4: Bordeaux Carmine
		{{R: 96, G: 26, B: 92, A: 255}, {R: 54, G: 24, B: 88, A: 255}},      // 5: Twilight Amethyst
	})

	lunarRingCols := render.ResolvePalettePairs(schema, common.Dunhuang, [][2]color.RGBA{
		{{R: 178, G: 246, B: 238, A: 255}, {R: 58, G: 196, B: 186, A: 255}}, // 0: Inner Seafoam Moon-Jade
		{{R: 44, G: 178, B: 164, A: 255}, {R: 24, G: 144, B: 158, A: 255}},  // 1: Celadon Turquoise
		{{R: 22, G: 138, B: 162, A: 255}, {R: 18, G: 102, B: 148, A: 255}},  // 2: Kingfisher Cyan
		{{R: 18, G: 94, B: 146, A: 255}, {R: 16, G: 62, B: 124, A: 255}},    // 3: Peacock Cerulean
		{{R: 16, G: 56, B: 118, A: 255}, {R: 14, G: 36, B: 92, A: 255}},     // 4: Royal Ultramarine
		{{R: 14, G: 32, B: 84, A: 255}, {R: 10, G: 18, B: 54, A: 255}},      // 5: Midnight Prussian Abyss
	})

	// 9 Sweeping Taiji S-Wave Ribbons that flow diagonally across the canvas,
	// curving around the Solar Eye on the upper-left and cradling the Lunar Eye on the lower-right!
	type taijiRibbon struct {
		id       int
		zHeight  float64
		qBase    float64
		halfW    float64
		phaseOff float64
		seedOff  float64
		dir      float64
		colFlank color.RGBA
		colCore  color.RGBA
	}

	sRibbonCols := render.ResolvePalettePairs(schema, common.Dunhuang, [][2]color.RGBA{
		{{R: 118, G: 26, B: 82, A: 255}, {R: 212, G: 58, B: 56, A: 255}},   // 0: Bordeaux & Vermilion
		{{R: 188, G: 38, B: 58, A: 255}, {R: 246, G: 162, B: 54, A: 255}},  // 1: Cinnabar & Amber Gold
		{{R: 234, G: 96, B: 48, A: 255}, {R: 255, G: 214, B: 92, A: 255}},   // 2: Sunset Coral & Sunburst Gold
		{{R: 248, G: 182, B: 54, A: 255}, {R: 48, G: 172, B: 136, A: 255}},  // 3: Imperial Gold & Celadon Jade (Center Yang Ribbon!)
		{{R: 42, G: 168, B: 132, A: 255}, {R: 252, G: 218, B: 104, A: 255}}, // 4: Malachite Jade & Champagne Gold (Center Bridge!)
		{{R: 22, G: 142, B: 158, A: 255}, {R: 246, G: 184, B: 64, A: 255}},  // 5: Kingfisher Teal & Gold (Center Yin Ribbon!)
		{{R: 18, G: 104, B: 148, A: 255}, {R: 54, G: 188, B: 176, A: 255}},  // 6: Cerulean & Seafoam Aqua
		{{R: 16, G: 64, B: 126, A: 255}, {R: 24, G: 138, B: 158, A: 255}},   // 7: Ultramarine & Kingfisher
		{{R: 48, G: 28, B: 98, A: 255}, {R: 18, G: 84, B: 142, A: 255}},     // 8: Royal Indigo & Cobalt
	})

	const nRibbons = 9
	ribbons := make([]taijiRibbon, nRibbons)
	for i := 0; i < nRibbons; i++ {
		fi := float64(i) / float64(nRibbons-1)
		dir := 1.0
		if i%2 == 1 {
			dir = -1.0
		}
		// Pyramid Z-ordering so the central Gold/Jade/Turquoise Taiji S-ribbons (i=3,4,5)
		// arch proudly in the foreground over the outer ribbons!
		distFromCenter := math.Abs(float64(i) - 4.0)
		zH := 26.0 - distFromCenter*1.8
		ribbons[i] = taijiRibbon{
			id:       50 + i,
			zHeight:  zH,
			qBase:    -0.28 + 0.56*fi,
			halfW:    0.036 + 0.005*math.Cos(float64(i)*1.1),
			phaseOff: float64(i)*0.55 + 0.25*rng.Float64(),
			seedOff:  27.3*float64(i+1) + rng.Float64()*35.0,
			dir:      dir,
			colFlank: sRibbonCols[i][0],
			colCore:  sRibbonCols[i][1],
		}
	}

	// =========================================================================
	// 2. RASTERIZE 2X SUPERSAMPLED Z-BUFFER, DUAL SANCTUARIES & S-BRIDGE
	// =========================================================================
	gw, gh := w*2+1, h*2+1
	zGrid := make([]float64, gw*gh)
	idGrid := make([]int, gw*gh)
	rimGrid := make([]float64, gw*gh)
	albRGrid := make([]float64, gw*gh)
	albGGrid := make([]float64, gw*gh)
	albBGrid := make([]float64, gw*gh)

	// Diagonal Taiji axis: es = (+1/sqrt(2), +1/sqrt(2)) from Sun (0.30,0.30) to Moon (0.70,0.70)
	// Orthogonal S-axis: et = (+1/sqrt(2), -1/sqrt(2)) from bottom-left to top-right!
	const invSqrt2 = 0.7071067811865475

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

					// 1) Dual Solar (Upper-Left) & Lunar (Lower-Right) Sanctuaries & Concentric Perlin Rings
					dSun := math.Hypot(ux-sunX, uy-sunY)
					dMoon := math.Hypot(ux-moonX, uy-moonY)

					curZ := 1.0
					curID := 1
					curRim := 0.0
					// Warm golden-amber base in Solar hemisphere, luminous aqua-teal base in Lunar hemisphere (zero dark gaps!)
					curCol := color.RGBA{R: 252, G: 208, B: 98, A: 255}
					if dMoon < dSun {
						curCol = color.RGBA{R: 28, G: 156, B: 166, A: 255}
					}

					// Determine whether background belongs to the Solar hemisphere (dSun <= dMoon) or Lunar hemisphere
					for side := 0; side < 2; side++ {
						isSolar := side == 0
						var rPole, thPole float64
						if isSolar {
							rPole = dSun
							thPole = math.Atan2(uy-sunY, ux-sunX)
						} else {
							rPole = dMoon
							thPole = math.Atan2(uy-moonY, ux-moonX)
						}
						cosTh, sinTh := math.Cos(thPole), math.Sin(thPole)

						// 3-Tier Recessed Sanctuary Core (rPole <= 0.145)
						haloRadii := [3]float64{0.142, 0.098, 0.058}
						for hIdx := 0; hIdx < 3; hIdx++ {
							rh := haloRadii[hIdx]
							if rPole <= rh {
								dPx := (rh - rPole) * fw
								curZ = 2.0 + float64(hIdx)*1.1
								curID = 10 + side*10 + hIdx
								if isSolar {
									cols := [3]color.RGBA{
										{R: 252, G: 208, B: 98, A: 255},
										{R: 255, G: 232, B: 142, A: 255},
										{R: 255, G: 250, B: 214, A: 255},
									}
									curCol = cols[hIdx]
								} else {
									cols := [3]color.RGBA{
										{R: 28, G: 156, B: 166, A: 255},
										{R: 112, G: 224, B: 218, A: 255},
										{R: 236, G: 254, B: 252, A: 255},
									}
									curCol = cols[hIdx]
								}
								if dPx < 1.8 {
									curRim = common.Smoothstep(1.8, 0.35, dPx)
								} else {
									curRim = 0.0
								}
							}
						}

						// 6 Concentric Perlin-Warped Paper-Cut Vortex Rings expanding around each Sanctuary
						for k := 0; k < 6; k++ {
							rOpen := 0.175 + float64(k)*0.068
							dir := 1.0
							if k%2 == 1 {
								dir = -1.0
							}
							// Seamless periodic Perlin on the unit circle (cosTh, sinTh) — ZERO branch cut!
							pVal, dEdx, dEdy := ne.FBM2DDeriv(
								cosTh*1.15+float64(k*13+side*31)+loopX*0.3*dir,
								sinTh*1.15-float64(k*9)+loopY*0.3,
								2, 2.0, 0.45,
							)
							dPdTh := 1.15 * (-sinTh*dEdx + cosTh*dEdy)
							wTh := 0.022 * math.Sin(3.0*thPole+float64(k)+dir*phi)
							dwTh := 0.022 * 3.0 * math.Cos(3.0*thPole+float64(k)+dir*phi)

							rBound := rOpen + 0.028*sAmp*pVal + wTh
							signedR := rPole - rBound
							if signedR < 0 {
								continue
							}
							// Only draw Solar rings in the upper-left dominance zone and Lunar rings in the lower-right
							if isSolar && dSun > dMoon+0.12 {
								continue
							}
							if !isSolar && dMoon > dSun+0.12 {
								continue
							}

							slopeAng := (0.028*sAmp*dPdTh + dwTh) / math.Max(0.10, rPole)
							invArc := 1.0 / math.Sqrt(1.0+slopeAng*slopeAng)
							distPx := signedR * fw * invArc

							spanT := common.Clamp01(signedR / 0.085)
							var pCol color.RGBA
							if isSolar {
								pCol = render.LerpOklab(solarRingCols[k][0], solarRingCols[k][1], spanT)
							} else {
								pCol = render.LerpOklab(lunarRingCols[k][0], lunarRingCols[k][1], spanT)
							}
							rib := math.Sin(distPx * 0.68)
							if rib > 0 {
								pCol = render.LerpOklab(pCol, color.RGBA{R: 255, G: 246, B: 220, A: 255}, rib*0.10)
							} else {
								pCol = render.LerpOklab(pCol, color.RGBA{R: 10, G: 8, B: 22, A: 255}, -rib*0.09)
							}

							zRing := 6.0 + float64(k)*1.8 + 0.35*math.Exp(-distPx/6.0)
							if zRing >= curZ {
								curZ = zRing
								curID = 30 + side*10 + k
								curCol = pCol
								if distPx < 1.65 {
									curRim = common.Smoothstep(1.65, 0.35, distPx)
								} else {
									curRim = 0.0
								}
							}
						}
					}

					// 2) 9 Sweeping Taiji S-Curve Silk-Brocade Bridge Ribbons!
					// Coordinate s runs from bottom-left to top-right: s = (ux-0.5)*invSqrt2 - (uy-0.5)*invSqrt2
					// Coordinate q runs from top-left (Sun) to bottom-right (Moon): q = (ux-0.5)*invSqrt2 + (uy-0.5)*invSqrt2
					dxC, dyC := ux-0.50, uy-0.50
					s := (dxC - dyC) * invSqrt2
					q := (dxC + dyC) * invSqrt2

					for i := 0; i < nRibbons; i++ {
						rb := &ribbons[i]
						if rb.zHeight < curZ {
							continue
						}

						// Grand Yin-Yang S-curve that arches over the Solar Eye (s > 0 -> q < 0)
						// and cradles under the Lunar Eye (s < 0 -> q > 0)!
						sArch := -0.24 * math.Sin(s*math.Pi*2.15+0.25*rb.dir*math.Sin(phi))
						dsArch := -0.24 * math.Pi * 2.15 * math.Cos(s*math.Pi*2.15+0.25*rb.dir*math.Sin(phi))

						pVal, dPds, _ := ne.FBM2DDeriv((s+0.7)*1.35+rb.seedOff*0.1+loopX*0.3*rb.dir, rb.qBase*2.0+loopY*0.3, 2, 2.0, 0.45)
						wHarm := 0.028 * math.Cos(s*math.Pi*3.2+rb.phaseOff+rb.dir*phi)
						dwHarm := -0.028 * math.Pi * 3.2 * math.Sin(s*math.Pi*3.2+rb.phaseOff+rb.dir*phi)

						qCurve := rb.qBase + sArch + 0.042*sAmp*pVal + wHarm
						dqds := dsArch + 0.042*sAmp*dPds*1.35 + dwHarm

						// Exact Euclidean normal distance to the S-bridge ribbon spine!
						invArc := 1.0 / math.Sqrt(1.0+dqds*dqds)
						dPerp := (q - qCurve) * invArc
						v := dPerp / rb.halfW
						absV := math.Abs(v)
						if absV > 1.0 {
							continue
						}

						distEdgePx := (1.0 - absV) * rb.halfW * fw

						// 3-Band Inlaid Polychrome Silk-Brocade + Inner Gold Pinstripe + Arched Silk Grooves
						centerBand := common.Smoothstep(0.44, 0.34, absV)
						pCol := render.LerpOklab(rb.colFlank, rb.colCore, centerBand*0.86)

						rib := math.Sin(v * 14.0 * math.Pi)
						if rib > 0 {
							pCol = render.LerpOklab(pCol, color.RGBA{R: 255, G: 248, B: 224, A: 255}, rib*0.15)
						} else {
							pCol = render.LerpOklab(pCol, color.RGBA{R: 10, G: 8, B: 22, A: 255}, -rib*0.14)
						}

						pinstripe := math.Exp(-math.Pow((absV-0.40)/0.032, 2.0))
						if pinstripe > 0.02 {
							pCol = render.LerpOklab(pCol, color.RGBA{R: 246, G: 192, B: 58, A: 255}, pinstripe*0.82)
						}

						curZ = rb.zHeight + 0.45*(1.0-math.Pow(absV, 3.0))
						curID = rb.id
						curCol = pCol
						if distEdgePx < 1.75 {
							curRim = common.Smoothstep(1.75, 0.35, distEdgePx)
						} else {
							curRim = 0.0
						}
					}

					grain := ne.Perlin2D(float64(gx)*0.35, float64(gy)*0.35) * 0.022
					rL, gL, bL, _ := render.ColorToLinearRGB(curCol)
					if curRim > 0 {
						rr, rg, rb, _ := render.ColorToLinearRGB(goldRim)
						rL = rL*(1.0-curRim*0.82) + rr*curRim*0.82
						gL = gL*(1.0-curRim*0.82) + rg*curRim*0.82
						bL = bL*(1.0-curRim*0.82) + rb*curRim*0.82
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
								aoSum += common.Constrain(dz/3.2, 0.0, 1.0)
							}
						}
						aoShade := 1.0 - (aoSum/12.0)*0.60

						castOcc := 0.0
						for step := 1; step <= 14; step++ {
							dist := float64(step) * 1.60
							sx := common.ConstrainInt(int(math.Round(float64(gx)-0.62*dist)), 0, gw-1)
							sy := common.ConstrainInt(int(math.Round(float64(gy)-0.78*dist)), 0, gh-1)
							sIdx := sy*gw + sx
							if idGrid[sIdx] != id0 && zGrid[sIdx] > z0 {
								falloff := 1.0 - float64(step-1)/14.0
								castOcc = math.Max(castOcc, falloff*0.52)
							}
						}
						castShade := 1.0 - castOcc

						ux := (float64(gx) * 0.5) / fw
						uy := (float64(gy) * 0.5) / fh
						dSun := math.Hypot(ux-sunX, uy-sunY)
						dMoon := math.Hypot(ux-moonX, uy-moonY)
						sunGlow := 0.14 * math.Exp(-(dSun*dSun)/0.075)
						moonGlow := 0.12 * math.Exp(-(dMoon*dMoon)/0.075)

						shade := aoShade * castShade
						rOut := albRGrid[idx]*shade + sunGlow*1.05 + moonGlow*0.45 + rimGrid[idx]*0.15
						gOut := albGGrid[idx]*shade + sunGlow*0.82 + moonGlow*0.92 + rimGrid[idx]*0.12
						bOut := albBGrid[idx]*shade + sunGlow*0.40 + moonGlow*1.05 + rimGrid[idx]*0.06

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
