package arts

import (
	"image/color"
	"math"
	"runtime"
	"sync"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type volcanicAtollTerraces struct {
	tiers           int
	islands         int
	erosionStrength float64
}

// NewVolcanicAtollTerraces returns a volcanicAtollTerraces engine that synthesizes a fierce,
// high-relief 3D laser-cut topographic relief of an Erupting Volcanic Caldera, Radial Molten
// Lava Rift Channels, Charred Obsidian-Basalt Crags & Boiling Turquoise Trench, with 3D drop
// shadows, incandescent magma bevels, and 24K gilded Marching-Squares contour rims.
func NewVolcanicAtollTerraces(tiers, islands int, erosionStrength float64) *volcanicAtollTerraces {
	if tiers <= 0 {
		tiers = 16
	}
	if islands <= 0 {
		islands = 4
	}
	if erosionStrength <= 0 {
		erosionStrength = 1.75
	}
	return &volcanicAtollTerraces{
		tiers:           common.ConstrainInt(tiers, 8, 22),
		islands:         common.ConstrainInt(islands, 2, 8),
		erosionStrength: common.Constrain(erosionStrength, 0.4, 3.2),
	}
}

// Generative renders a static frame of the Erupting Volcanic Caldera Relief.
func (vat *volcanicAtollTerraces) Generative(c *generativeart.Canva) {
	vat.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360-degree phase loop (phase in [0, 1)).
func (vat *volcanicAtollTerraces) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	minDim := math.Min(fw, fh)
	scale := minDim / 600.0
	animAngle := 2.0 * math.Pi * phase

	rc := c.NewRenderCanvas()
	obsidianFrameCol := color.RGBA{R: 0x10, G: 0x0d, B: 0x14, A: 255}
	rc.FillBackground(obsidianFrameCol)
	rc.SetBlendMode(render.BlendOver)

	seed := c.Seed()
	if seed == 0 {
		seed = 20262502
	}
	ne := common.NewNoiseEngine(seed + 2502)

	gw, gh := w*2+1, h*2+1
	rawGrid := make([]float64, gw*gh)

	driftX := 0.055 * math.Cos(animAngle)
	driftY := 0.055 * math.Sin(animAngle)
	eStr := vat.erosionStrength * 1.25
	islScale := float64(vat.islands) / 4.0

	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}

	var wg sync.WaitGroup
	for wIdx := 0; wIdx < workers; wIdx++ {
		gy0 := (wIdx * gh) / workers
		gy1 := ((wIdx + 1) * gh) / workers
		wg.Add(1)
		go func(rowStart, rowEnd int) {
			defer wg.Done()
			for gy := rowStart; gy < rowEnd; gy++ {
				py := float64(gy) * 0.5
				ny := (py/fh)*2.0 - 1.0
				rowOff := gy * gw
				for gx := 0; gx < gw; gx++ {
					px := float64(gx) * 0.5
					nx := (px/fw)*2.0 - 1.0

					// Asymmetric tectonic domain warp (NO radial symmetry!)
					w1x := ne.Perlin2D(nx*1.65+1.7, ny*1.65-2.1)
					w1y := ne.Perlin2D(nx*1.65-2.8, ny*1.65+1.9)
					wx := nx + driftX + 0.22*w1x
					wy := ny + driftY + 0.22*w1y

					// Multi-scale gradient-damped ErosionFBM2D + sharp Ridged tectonic crags
					eroded, sx, sy := ne.ErosionFBM2D(wx*1.95+1.4, wy*1.95-0.8, 5, 2.05, 0.50, eStr)
					ridged := ne.Ridged2D(wx*2.35-0.7, wy*2.35+1.5, 4, 2.1, 0.50)

					// 1. Diagonal Subduction Ocean Trench & Boiling Bay entering from NW/SW
					// Runs along a meandering S-curve fault from top-left (-0.45, -1.0) to bottom-center (0.05, +1.0)
					alongTrench := -0.42*nx + 0.91*ny
					crossTrench := 0.91*nx + 0.42*ny + 0.28*math.Sin(alongTrench*2.1+0.4) + 0.16*w1x
					// West/Left side of fault drops into the deep ocean trench; East/Right rises into the volcanic arc
					continentalShelf := common.Smoothstep(-0.52, 0.18, crossTrench)

					// Deep fjord & steam-explosion bay cutting into the massif from SW toward the main crater
					gorgeAxis := 0.64*(nx-0.28) + 0.77*(ny+0.18)
					gorgeCross := -0.77*(nx-0.28) + 0.64*(ny+0.18) + 0.18*math.Sin(gorgeAxis*3.4) + 0.10*w1y
					swBreachGorge := 1.0 - 0.62*math.Exp(-math.Pow(gorgeCross/0.16, 2.0))*common.Smoothstep(0.15, -0.65, gorgeAxis)

					// 2. Primary Asymmetric Breached Supervolcano Massif in the Upper-Right / East (cx=0.30, cy=-0.22)
					// Elliptically sheared along a NE-SW tectonic fault line
					dx1 := nx - 0.30 + 0.08*sx
					dy1 := ny + 0.22 + 0.08*sy
					u1 := 0.78*dx1 - 0.62*dy1
					v1 := 0.62*dx1 + 0.78*dy1
					r1 := math.Hypot(u1*0.88, v1*1.22)
					mainMassif := 0.96 * math.Exp(-(r1*r1)/0.34) * (0.72 + 0.28*swBreachGorge)

					// Searing off-center magma fissure inside the breached upper-right caldera + flowing lava tongue
					lavaTongueDist := math.Hypot(gorgeCross*1.35, math.Max(0.0, gorgeAxis-0.08)*0.85)
					moltenStream := 0.58 * math.Exp(-math.Pow(lavaTongueDist/0.14, 2.0)) * common.Smoothstep(-0.75, 0.05, gorgeAxis)

					// 3. Secondary Jagged Rift Ridge in the Lower-Right / South-East (cx=0.22, cy=0.52)
					dx2 := nx - 0.22 + 0.06*sx
					dy2 := ny - 0.52 + 0.06*sy
					u2 := 0.85*dx2 + 0.52*dy2
					v2 := -0.52*dx2 + 0.85*dy2
					r2 := math.Hypot(u2*0.75, v2*1.35)
					southMassif := 0.78 * math.Exp(-(r2*r2)/0.22)

					// 4. Isolated Offshore Volcanic Sea-Stack Island in the NW Boiling Trench (cx=-0.58, cy=-0.32)
					dx3 := nx + 0.58 - 0.05*sx
					dy3 := ny + 0.32 + 0.05*sy
					r3 := math.Hypot(dx3*1.15, dy3*0.85)
					nwSeaStack := (0.62 * islScale) * math.Exp(-(r3*r3)/0.085)

					// Combine continental fault, asymmetric volcanic massifs, molten lava stream, and eroded/ridged relief
					hVal := continentalShelf*(0.24+0.44*eroded+0.24*ridged) +
						mainMassif + moltenStream + southMassif + nwSeaStack +
						0.12*eroded*(1.0-continentalShelf)

					rawGrid[rowOff+gx] = hVal
				}
			}
		}(gy0, gy1)
	}
	wg.Wait()

	minH, maxH := rawGrid[0], rawGrid[0]
	for _, v := range rawGrid {
		if v < minH {
			minH = v
		}
		if v > maxH {
			maxH = v
		}
	}
	spanH := math.Max(1e-6, maxH-minH)
	invSpanH := 1.0 / spanH
	for i, v := range rawGrid {
		norm := common.Clamp01((v - minH) * invSpanH)
		shaped := 0.78*norm + 0.22*common.Smoothstep(0.0, 1.0, norm)
		rawGrid[i] = 0.004 + 0.991*common.Clamp01(shaped)
	}

	sampleH := func(px, py float64) float64 {
		gx := common.Constrain(px*2.0, 0.0, float64(gw-1))
		gy := common.Constrain(py*2.0, 0.0, float64(gh-1))
		ix := int(gx)
		iy := int(gy)
		fx := gx - float64(ix)
		fy := gy - float64(iy)
		ix1 := common.MinInt(ix+1, gw-1)
		iy1 := common.MinInt(iy+1, gh-1)

		v00 := rawGrid[iy*gw+ix]
		v10 := rawGrid[iy*gw+ix1]
		v01 := rawGrid[iy1*gw+ix]
		v11 := rawGrid[iy1*gw+ix1]

		top := v00*(1.0-fx) + v10*fx
		bot := v01*(1.0-fx) + v11*fx
		return top*(1.0-fy) + bot*fy
	}

	numTiers := vat.tiers
	fTiers := float64(numTiers)

	// Fierce 16-Stop Volcanic Eruption Palette:
	// Deep Boiling Trench -> Steaming Turquoise Thermal Shock -> Charred Obsidian Basalt -> Molten Crimson -> Incandescent White-Hot Magma!
	baseStops := []color.RGBA{
		{R: 0x05, G: 0x0d, B: 0x22, A: 255}, // 0: Abyssal Ocean Trench (#050d22)
		{R: 0x0a, G: 0x24, B: 0x4e, A: 255}, // 1: Deep Prussian Sea (#0a244e)
		{R: 0x12, G: 0x56, B: 0x7e, A: 255}, // 2: Boiling Cobalt Shelf (#12567e)
		{R: 0x1c, G: 0xa4, B: 0xa8, A: 255}, // 3: Steaming Thermal Turquoise (#1ca4a8)
		{R: 0x52, G: 0xdc, B: 0xc8, A: 255}, // 4: Vaporous Emerald-Cyan Shock Rim (#52dcc8)
		{R: 0x14, G: 0x11, B: 0x1a, A: 255}, // 5: Quenched Black Basalt Shore (#14111a)
		{R: 0x22, G: 0x16, B: 0x24, A: 255}, // 6: Charred Obsidian Crag (#221624)
		{R: 0x38, G: 0x1a, B: 0x2a, A: 255}, // 7: Scorched Porphyry Talus (#381a2a)
		{R: 0x5a, G: 0x1c, B: 0x28, A: 255}, // 8: Smoldering Cinder Flank (#5a1c28)
		{R: 0x8e, G: 0x1e, B: 0x22, A: 255}, // 9: Deep Magma Crust (#8e1e22)
		{R: 0xc6, G: 0x26, B: 0x1c, A: 255}, // 10: Molten Cinnabar Lava (#c6261c)
		{R: 0xe8, G: 0x48, B: 0x16, A: 255}, // 11: Erupting Vermilion Flow (#e84816)
		{R: 0xf8, G: 0x78, B: 0x18, A: 255}, // 12: Incandescent Solar Orange (#f87818)
		{R: 0xfe, G: 0xb2, B: 0x2c, A: 255}, // 13: Liquid Gold-Amber Magma (#feb22c)
		{R: 0xff, G: 0xe6, B: 0x68, A: 255}, // 14: Blazing Sulfur-Gold Vent (#ffe668)
		{R: 0xff, G: 0xfc, B: 0xeb, A: 255}, // 15: White-Hot Eruption Heart (#fffceb)
	}
	baseGrad := render.NewMultiStopGradient(baseStops...)

	schema := c.Opts().ColorSchema()
	var userGrad *render.MultiStopGradient
	if len(schema) > 0 {
		userGrad = render.NewMultiStopGradient(schema...)
	}

	palLin := make([][3]float64, numTiers)
	for i := 0; i < numTiers; i++ {
		t := float64(i) / math.Max(1.0, float64(numTiers-1))
		col := baseGrad.SampleOklab(t)
		if userGrad != nil {
			col = render.LerpOklab(col, userGrad.SampleOklab(t), 0.14)
		}
		r, g, b, _ := render.ColorToLinearRGB(col)
		palLin[i] = [3]float64{r, g, b}
	}

	moltenGoldColor := color.RGBA{R: 0xff, G: 0xc4, B: 0x42, A: 255}
	lavaFireColor := color.RGBA{R: 0xff, G: 0x58, B: 0x18, A: 255}
	charcoalInkColor := color.RGBA{R: 0x0d, G: 0x0b, B: 0x14, A: 255}
	goldLinR, goldLinG, goldLinB, _ := render.ColorToLinearRGB(moltenGoldColor)
	fireLinR, fireLinG, fireLinB, _ := render.ColorToLinearRGB(lavaFireColor)
	inkLinR, inkLinG, inkLinB, _ := render.ColorToLinearRGB(charcoalInkColor)

	sunOffX := -5.0 * scale
	sunOffY := -6.0 * scale
	sunDistFull := math.Hypot(5.0, 6.0)

	subOffsets := [2]float64{0.25, 0.75}
	for wIdx := 0; wIdx < workers; wIdx++ {
		y0 := (wIdx * h) / workers
		y1 := ((wIdx + 1) * h) / workers
		wg.Add(1)
		go func(rowStart, rowEnd int) {
			defer wg.Done()
			for y := rowStart; y < rowEnd; y++ {
				fyBase := float64(y)
				for x := 0; x < w; x++ {
					fxBase := float64(x)

					var accR, accG, accB float64
					for syIdx := 0; syIdx < 2; syIdx++ {
						py := fyBase + subOffsets[syIdx]
						for sxIdx := 0; sxIdx < 2; sxIdx++ {
							px := fxBase + subOffsets[sxIdx]

							hVal := sampleH(px, py)
							scaledH := hVal * fTiers
							k := int(math.Floor(scaledH))
							if k < 0 {
								k = 0
							} else if k >= numTiers {
								k = numTiers - 1
							}
							frac := scaledH - float64(k)

							baseCol := palLin[k]
							r, g, b := baseCol[0], baseCol[1], baseCol[2]

							dhdx := (sampleH(px+1.0*scale, py) - sampleH(px-1.0*scale, py)) * fTiers
							dhdy := (sampleH(px, py+1.0*scale) - sampleH(px, py-1.0*scale)) * fTiers
							sunSlope := -0.63*dhdx - 0.77*dhdy
							plateShade := 0.94 + 0.10*frac + common.Constrain(sunSlope*0.045, -0.08, 0.08)
							r *= plateShade
							g *= plateShade
							b *= plateShade

							// Boiling thermal shock waves in the ocean tiers (k <= 4)
							if k <= 4 {
								waterWave := math.Sin((px*0.22-py*0.18)/math.Max(0.25, scale) - animAngle*2.0 + frac*14.0)
								shimmer := 1.0 + 0.065*waterWave
								r *= shimmer
								g *= shimmer
								b *= shimmer
							}

							// Self-emissive thermal glow inside high magma tiers (k >= 11)
							if k >= (numTiers*11)/16 {
								magmaBoost := 1.0 + 0.28*frac
								r *= magmaBoost
								g *= magmaBoost
								b *= magmaBoost
							}

							hitHigher := false
							minHitDist := sunDistFull
							maxStepDiff := 0
							for tap := 1; tap <= 6; tap++ {
								t := float64(tap) / 6.0
								spx := px + sunOffX*t
								spy := py + sunOffY*t
								hSun := sampleH(spx, spy)
								kSun := int(math.Floor(hSun * fTiers))
								if kSun >= numTiers {
									kSun = numTiers - 1
								}
								if kSun > k {
									if !hitHigher {
										hitHigher = true
										minHitDist = t * sunDistFull
									}
									if kSun-k > maxStepDiff {
										maxStepDiff = kSun - k
									}
								}
							}

							if hitHigher && k < numTiers-2 {
								shadowDarken := 0.58 * math.Exp(-minHitDist/5.8)
								if maxStepDiff > 1 {
									shadowDarken = math.Min(0.72, shadowDarken*(1.0+0.12*float64(maxStepDiff-1)))
								}
								keep := 1.0 - shadowDarken
								r = r*keep + 0.006*shadowDarken
								g = g*keep + 0.004*shadowDarken
								b = b*keep + 0.012*shadowDarken
							}

							// Incandescent magma bevel highlight along step lips
							hNearSun := sampleH(px-1.4*scale, py-1.7*scale)
							kNearSun := int(math.Floor(hNearSun * fTiers))
							if kNearSun < k && k > 0 {
								bevelStrength := 0.52 * (1.0 - common.Smoothstep(0.0, 0.18, frac))
								if k >= numTiers/3 {
									r = r*(1.0-bevelStrength) + (goldLinR*1.45)*bevelStrength
									g = g*(1.0-bevelStrength) + (goldLinG*1.32)*bevelStrength
									b = b*(1.0-bevelStrength) + (goldLinB*1.10)*bevelStrength
								} else {
									r = r*(1.0-bevelStrength) + (fireLinR*1.25)*bevelStrength
									g = g*(1.0-bevelStrength) + (fireLinG*1.20)*bevelStrength
									b = b*(1.0-bevelStrength) + (fireLinB*1.10)*bevelStrength
								}
							}

							if (frac < 0.045 && k > 0) || (frac > 0.955 && k < numTiers-1) {
								distToEdge := frac
								boundaryK := k
								if frac > 0.955 {
									distToEdge = 1.0 - frac
									boundaryK = k + 1
								}
								edgeWeight := 0.72 * (1.0 - common.Smoothstep(0.0, 0.045, distToEdge))
								if boundaryK >= 5 {
									r = r*(1.0-edgeWeight) + goldLinR*1.2*edgeWeight
									g = g*(1.0-edgeWeight) + goldLinG*1.1*edgeWeight
									b = b*(1.0-edgeWeight) + goldLinB*edgeWeight
								} else {
									r = r*(1.0-edgeWeight) + inkLinR*edgeWeight
									g = g*(1.0-edgeWeight) + inkLinG*edgeWeight
									b = b*(1.0-edgeWeight) + inkLinB*edgeWeight
								}
							}

							accR += r
							accG += g
							accB += b
						}
					}

					rc.SetPixelLinear(x, y, accR*0.25, accG*0.25, accB*0.25, 1.0)
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	// 4. Overlay Crisp Vector Marching-Squares Molten Gold & Basalt Contour Lines
	fieldFn := func(x, y float64) float64 {
		return sampleH(x, y)
	}

	lineWidth := 1.20 * scale
	for k := 1; k < numTiers; k++ {
		isoVal := float64(k) / fTiers
		contours := render.ExtractMarchingSquares(130, 130, fw, fh, isoVal, fieldFn)

		isMagmaOrBasalt := k >= 5
		strokeCol := charcoalInkColor
		underCol := moltenGoldColor
		strokeAlpha := 0.88
		if isMagmaOrBasalt {
			strokeCol = moltenGoldColor
			if k >= 9 && k <= 12 {
				strokeCol = lavaFireColor
			}
			underCol = charcoalInkColor
			strokeAlpha = 0.94
		}

		for _, poly := range contours {
			if len(poly) < 2 {
				continue
			}
			smoothPoly := poly
			if len(poly) >= 4 {
				smoothPoly = render.CatmullRomSpline(poly, 2)
			}
			rc.StrokePolyline(smoothPoly, lineWidth*1.55, underCol, 0.48)
			rc.StrokePolyline(smoothPoly, lineWidth, strokeCol, strokeAlpha)
		}
	}

	// 5. Ejecting Pyroclastic Lava Bombs & Glowing Embers blasting SW from the Upper-Right Crater
	rc.SetBlendMode(render.BlendAdd)
	emberCore := color.RGBA{R: 255, G: 250, B: 210, A: 255}
	emberHalo := color.RGBA{R: 255, G: 110, B: 24, A: 255}
	ventX, ventY := 0.64*fw, 0.39*fh
	for i := 0; i < 36; i++ {
		fi := float64(i)
		// Directed SW pyroclastic blast fan (angles around 2.35 rad = SW) with organic scatter
		ang := 2.32 + 0.85*math.Sin(fi*1.73) + 0.12*math.Sin(animAngle+fi)
		rad := (0.03 + 0.42*math.Mod(fi*0.618034+phase*0.5, 1.0)) * minDim
		ex := ventX + math.Cos(ang)*rad
		ey := ventY + math.Sin(ang)*rad
		if ex > 24*scale && ex < fw-24*scale && ey > 24*scale && ey < fh-24*scale {
			eRad := (2.0 + 2.4*math.Sin(fi*1.7+animAngle)) * scale
			if eRad > 1.1*scale {
				rc.DrawGlowDot(ex, ey, eRad*2.3, emberHalo, 0.40, 0.25)
				rc.DrawGlowDot(ex, ey, eRad*0.85, emberCore, 0.86, 0.75)
			}
		}
	}
	rc.SetBlendMode(render.BlendOver)

	// 6. Dark Obsidian-Gold Cartographic Frame
	borderW := 16.0 * scale
	if borderW >= 2.0 {
		obsLinR, obsLinG, obsLinB, _ := render.ColorToLinearRGB(obsidianFrameCol)
		ib := int(math.Ceil(borderW))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if x < ib || x >= w-ib || y < ib || y >= h-ib {
					rc.SetPixelLinear(x, y, obsLinR, obsLinG, obsLinB, 1.0)
				}
			}
		}

		outerPad := 4.5 * scale
		innerPad := borderW
		outerRect := []render.Point{
			{X: outerPad, Y: outerPad},
			{X: fw - outerPad, Y: outerPad},
			{X: fw - outerPad, Y: fh - outerPad},
			{X: outerPad, Y: fh - outerPad},
			{X: outerPad, Y: outerPad},
		}
		innerRect := []render.Point{
			{X: innerPad, Y: innerPad},
			{X: fw - innerPad, Y: innerPad},
			{X: fw - innerPad, Y: fh - innerPad},
			{X: innerPad, Y: fh - innerPad},
			{X: innerPad, Y: innerPad},
		}
		rc.StrokePolyline(outerRect, 0.9*scale, moltenGoldColor, 0.65)
		rc.StrokePolyline(innerRect, 1.8*scale, moltenGoldColor, 0.92)
	}

	rc.ApplyBloom(0.78, 4.0*scale, 0.18)
	c.MergeRenderCanvas(rc)
}
