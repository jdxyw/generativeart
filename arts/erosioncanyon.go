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

type erosionCanyon struct {
	octaves         int
	erosionStrength float64
	strataFreq      float64
}

// NewErosionCanyon returns an erosionCanyon engine that synthesizes a crisp
// architectural 3D laser-cut Perlin topographic paper/lacquer relief with
// 16 Ukiyo-e archival terrace plates, 3D paper-cut drop shadows, beveled lips,
// and 24k gilded Marching Squares contour rims.
func NewErosionCanyon(octaves int, erosionStrength, strataFreq float64) *erosionCanyon {
	if octaves <= 0 {
		octaves = 6
	}
	if erosionStrength <= 0 {
		erosionStrength = 1.8
	}
	if strataFreq <= 0 {
		strataFreq = 16.0
	}
	return &erosionCanyon{
		octaves:         octaves,
		erosionStrength: erosionStrength,
		strataFreq:      strataFreq,
	}
}

// Generative renders a static frame of the 3D Laser-Cut Topographic Paper/Lacquer Relief.
func (ec *erosionCanyon) Generative(c *generativeart.Canva) {
	ec.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360-degree phase loop (phase in [0, 1)).
func (ec *erosionCanyon) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	minDim := math.Min(fw, fh)
	scale := minDim / 600.0
	animAngle := 2.0 * math.Pi * phase

	rc := c.NewRenderCanvas()
	archivalParchment := color.RGBA{R: 0xf4, G: 0xec, B: 0xd6, A: 255}
	rc.FillBackground(archivalParchment)
	rc.SetBlendMode(render.BlendOver)

	eStrength := ec.erosionStrength
	if eStrength < 0.2 {
		eStrength = 1.8
	}

	seed := c.Seed()
	if seed == 0 {
		seed = 20261101
	}
	ne := common.NewNoiseEngine(seed + 1902)

	// 1. Precompute Smooth Perlin + ErosionFBM2D Scalar Elevation Field H(x, y) in [0, 1]
	// Evaluated on a 2x supersampled grid (gw x gh) for fast, exact 2x2 SSAA and ray-cast drop shadows.
	gw, gh := w*2+1, h*2+1
	rawGrid := make([]float64, gw*gh)

	driftX := 0.06 * math.Cos(animAngle)
	driftY := 0.06 * math.Sin(animAngle)

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

					// Subtle organic domain warp so terrace contours flow like carved woodblock/lacquer plates
					wx := nx + driftX + 0.14*ne.Perlin2D(nx*1.35+2.1, ny*1.35-1.4)
					wy := ny + driftY + 0.14*ne.Perlin2D(nx*1.35-3.2, ny*1.35+2.7)

					// Analytical gradient-damped ErosionFBM2D at frequency 1.65, 4 octaves
					eroded, sx, sy := ne.ErosionFBM2D(wx*1.65+1.8, wy*1.65-1.2, 4, 2.0, 0.46, eStrength)
					fineDetail := ne.Perlin2D(wx*3.1-0.7, wy*3.1+1.9)

					// Graceful diagonal meandering canyon valley (running SW <-> NE with S-curve bend)
					alongCanyon := 0.72*nx + 0.69*ny
					crossCanyon := 0.69*nx - 0.72*ny + 0.22*math.Sin(alongCanyon*2.15+0.35) + 0.08*ne.Perlin2D(nx*1.2, ny*1.2)
					canyonDist := math.Abs(crossCanyon)

					// Secondary branching tributary gorge entering from NW
					tribCross := (0.55*nx + 0.83*ny + 0.24) + 0.12*math.Sin(nx*3.2)
					tribDist := math.Abs(tribCross)
					if crossCanyon > 0.08 {
						tribDist += (crossCanyon - 0.08) * 2.5
					}

					// Smooth canyon basin profile: 0 in abyssal lake center -> 1 on high canyon rims
					mainValley := common.Smoothstep(0.02, 0.54, canyonDist)
					innerAbyss := common.Smoothstep(0.00, 0.22, canyonDist)
					tribValley := 0.35 + 0.65*common.Smoothstep(0.03, 0.28, tribDist)
					valleyEnvelope := mainValley * (0.28 + 0.72*innerAbyss) * tribValley

					// Twin mountain peaks flanking the diagonal canyon + secondary ridge shoulders
					dPeak1 := math.Hypot(nx-0.46+0.05*sx, ny+0.40+0.05*sy)
					peak1 := 0.58 * math.Exp(-(dPeak1*dPeak1)/0.30)

					dPeak2 := math.Hypot(nx+0.46+0.05*sx, ny-0.42+0.05*sy)
					peak2 := 0.56 * math.Exp(-(dPeak2*dPeak2)/0.32)

					dShoulder1 := math.Hypot(nx-0.52, ny-0.38)
					shoulder1 := 0.28 * math.Exp(-(dShoulder1*dShoulder1)/0.36)

					dShoulder2 := math.Hypot(nx+0.50, ny+0.44)
					shoulder2 := 0.26 * math.Exp(-(dShoulder2*dShoulder2)/0.36)

					// Combine valley structure, twin alpine peaks, and eroded FBM relief
					hVal := valleyEnvelope*(0.26+0.52*eroded+0.06*fineDetail) +
						valleyEnvelope*(peak1+peak2+shoulder1+shoulder2) +
						0.06*eroded*innerAbyss

					rawGrid[rowOff+gx] = hVal
				}
			}
		}(gy0, gy1)
	}
	wg.Wait()

	// Normalize rawGrid cleanly to [0.005, 0.995] so all 16 terrace tiers (k = 0..15) are guaranteed present
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
		// Subtle S-curve contrast shaping to balance abyssal lake, foothills, cliffs, and snow summits
		shaped := 0.82*norm + 0.18*common.Smoothstep(0.0, 1.0, norm)
		rawGrid[i] = 0.004 + 0.991*common.Clamp01(shaped)
	}

	// Bilinear sampler on the 2x supersampled elevation grid for continuous (px, py) queries
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

	// 2. Curated 16-Stop Ukiyo-e / Archival Map Palette (k = 0..15)
	const numTiers = 16
	const fTiers = 16.0
	paletteSRGB := [numTiers]color.RGBA{
		// Tiers 0..3: Deep Abyssal Canyon Lake (#081834 -> #0f3664 -> #18688c -> #2ca8a8)
		{R: 0x08, G: 0x18, B: 0x34, A: 255},
		{R: 0x0f, G: 0x36, B: 0x64, A: 255},
		{R: 0x18, G: 0x68, B: 0x8c, A: 255},
		{R: 0x2c, G: 0xa8, B: 0xa8, A: 255},
		// Tiers 4..7: Shoreline & Forested Foothills (#52c8b4 -> #98dec0 -> #f4ecd6 -> #eed8b4)
		{R: 0x52, G: 0xc8, B: 0xb4, A: 255},
		{R: 0x98, G: 0xde, B: 0xc0, A: 255},
		{R: 0xf4, G: 0xec, B: 0xd6, A: 255},
		{R: 0xee, G: 0xd8, B: 0xb4, A: 255},
		// Tiers 8..11: Sunlit Ochre & Terracotta Cliffs (#f2b852 -> #ec8c38 -> #dc582e -> #be3428)
		{R: 0xf2, G: 0xb8, B: 0x52, A: 255},
		{R: 0xec, G: 0x8c, B: 0x38, A: 255},
		{R: 0xdc, G: 0x58, B: 0x2e, A: 255},
		{R: 0xbe, G: 0x34, B: 0x28, A: 255},
		// Tiers 12..15: High Alpine Ridges & Snow Caps (#8c2638 -> #582442 -> #dcdce6 -> #ffffff)
		{R: 0x8c, G: 0x26, B: 0x38, A: 255},
		{R: 0x58, G: 0x24, B: 0x42, A: 255},
		{R: 0xdc, G: 0xdc, B: 0xe6, A: 255},
		{R: 0xff, G: 0xff, B: 0xff, A: 255},
	}

	var palLin [numTiers][3]float64
	for i := 0; i < numTiers; i++ {
		r, g, b, _ := render.ColorToLinearRGB(paletteSRGB[i])
		palLin[i] = [3]float64{r, g, b}
	}

	goldRimColor := color.RGBA{R: 0xf8, G: 0xcf, B: 0x6a, A: 255}     // #f8cf6a burnished 24k gold
	charcoalInkColor := color.RGBA{R: 0x18, G: 0x1c, B: 0x28, A: 255} // #181c28 deep charcoal-ink
	goldLinR, goldLinG, goldLinB, _ := render.ColorToLinearRGB(goldRimColor)
	inkLinR, inkLinG, inkLinB, _ := render.ColorToLinearRGB(charcoalInkColor)

	// Upper-left sun shadow offset vector: (-4.5*scale, -5.5*scale)
	sunOffX := -4.5 * scale
	sunOffY := -5.5 * scale
	sunDistFull := math.Hypot(4.5, 5.5) // ~7.106 in unscaled pixel units

	// 3. 3D Drop-Shadowed Paper-Cut / Lacquer Relief Shader with 2x2 SSAA
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

							// Subtle satin-lacquer plate tonal gradation across the terrace width
							// plus gentle directional paper relief from local slope
							dhdx := (sampleH(px+1.0*scale, py) - sampleH(px-1.0*scale, py)) * fTiers
							dhdy := (sampleH(px, py+1.0*scale) - sampleH(px, py-1.0*scale)) * fTiers
							sunSlope := (-0.63*dhdx - 0.77*dhdy)
							plateShade := 0.96 + 0.06*frac + common.Constrain(sunSlope*0.035, -0.06, 0.06)
							r *= plateShade
							g *= plateShade
							b *= plateShade

							// Subtle water ripple lines on the lowest abyssal lake tiers (k <= 3)
							if k <= 3 {
								waterWave := math.Sin((px+py)*0.18/math.Max(0.25, scale) - animAngle*2.0 + frac*12.0)
								shimmer := 1.0 + 0.045*waterWave
								r *= shimmer
								g *= shimmer
								b *= shimmer
							}

							// Check along the upper-left sun vector (px - 4.5*scale, py - 5.5*scale)
							// using 6 ray-march taps so we capture both exact full offset and the closest casting edge distance
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

							// Cast smooth, realistic 3D paper-cut drop shadow onto the lower terrace
							if hitHigher {
								shadowDarken := 0.52 * math.Exp(-minHitDist/6.0)
								if maxStepDiff > 1 {
									shadowDarken = math.Min(0.64, shadowDarken*(1.0+0.10*float64(maxStepDiff-1)))
								}
								keep := 1.0 - shadowDarken
								// Cool deep indigo-umber ambient shadow tint so orange/ochre cliffs never turn muddy
								r = r*keep + 0.004*shadowDarken
								g = g*keep + 0.008*shadowDarken
								b = b*keep*1.06 + 0.022*shadowDarken
							}

							// 3D Bevel Highlight right along the sunlit upper-left lip of each contour step
							// Check if a point slightly sunward (1.35*scale toward upper-left) drops to a lower terrace k_near < k
							hNearSun := sampleH(px-1.35*scale, py-1.65*scale)
							kNearSun := int(math.Floor(hNearSun * fTiers))
							if kNearSun < k && k > 0 {
								// Sunlit upper-left beveled edge of this laser-cut plate
								bevelStrength := 0.42 * (1.0 - common.Smoothstep(0.0, 0.16, frac))
								r = r*(1.0-bevelStrength) + (goldLinR*1.25)*bevelStrength
								g = g*(1.0-bevelStrength) + (goldLinG*1.22)*bevelStrength
								b = b*(1.0-bevelStrength) + (goldLinB*1.15)*bevelStrength
							}

							// Fine recessed dark/gold ink line right on the contour boundary (where frac(H*N) is within 0.04 of 0 or 1)
							if (frac < 0.04 && k > 0) || (frac > 0.96 && k < numTiers-1) {
								distToEdge := frac
								boundaryK := k
								if frac > 0.96 {
									distToEdge = 1.0 - frac
									boundaryK = k + 1
								}
								edgeWeight := 0.65 * (1.0 - common.Smoothstep(0.0, 0.04, distToEdge))
								isDarkTier := boundaryK <= 5 || (boundaryK >= 11 && boundaryK <= 14)
								if isDarkTier {
									r = r*(1.0-edgeWeight) + goldLinR*edgeWeight
									g = g*(1.0-edgeWeight) + goldLinG*edgeWeight
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

	// 4. Overlay Crisp Vector Marching-Squares Gilded Contour Lines (k = 1..N-1)
	fieldFn := func(x, y float64) float64 {
		return sampleH(x, y)
	}

	lineWidth := 1.15 * scale
	for k := 1; k < numTiers; k++ {
		isoVal := float64(k) / fTiers
		contours := render.ExtractMarchingSquares(120, 120, fw, fh, isoVal, fieldFn)

		// Burnished gold #f8cf6a on dark tiers and deep charcoal-ink #181c28 on light tiers
		isDarkTier := k <= 5 || (k >= 11 && k <= 14)
		strokeCol := charcoalInkColor
		strokeAlpha := 0.86
		if isDarkTier {
			strokeCol = goldRimColor
			strokeAlpha = 0.92
		}

		for _, poly := range contours {
			if len(poly) < 2 {
				continue
			}
			smoothPoly := poly
			if len(poly) >= 4 {
				smoothPoly = render.CatmullRomSpline(poly, 2)
			}
			// Subtle recessed charcoal under-shadow for gold rims to make the 24k gold pop crisply
			if isDarkTier {
				rc.StrokePolyline(smoothPoly, lineWidth*1.45, charcoalInkColor, 0.42)
			} else {
				rc.StrokePolyline(smoothPoly, lineWidth*1.45, goldRimColor, 0.45)
			}
			rc.StrokePolyline(smoothPoly, lineWidth, strokeCol, strokeAlpha)
		}
	}

	// 5. Clean 18px*scale Archival Map Border with Subtle Topographic Grid Tick Marks
	borderW := 18.0 * scale
	if borderW >= 2.0 {
		parchmentLinR, parchmentLinG, parchmentLinB, _ := render.ColorToLinearRGB(archivalParchment)
		ib := int(math.Ceil(borderW))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if x < ib || x >= w-ib || y < ib || y >= h-ib {
					rc.SetPixelLinear(x, y, parchmentLinR, parchmentLinG, parchmentLinB, 1.0)
				}
			}
		}

		// Outer and inner architectural frame rules
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
		rc.StrokePolyline(outerRect, 0.85*scale, charcoalInkColor, 0.72)
		rc.StrokePolyline(innerRect, 1.85*scale, goldRimColor, 0.90)
		rc.StrokePolyline(innerRect, 1.05*scale, charcoalInkColor, 0.88)

		// Topographic coordinate graduation ticks along the archival map border
		tickStep := math.Max(12.0, 28.0*scale)
		tickLenMajor := 5.5 * scale
		tickLenMinor := 3.0 * scale
		idxTick := 0
		for tx := innerPad + tickStep*0.5; tx < fw-innerPad-2.0*scale; tx += tickStep * 0.5 {
			tLen := tickLenMinor
			alpha := 0.45
			if idxTick%2 == 0 {
				tLen = tickLenMajor
				alpha = 0.72
			}
			// Top & bottom border ticks
			rc.StrokePolyline([]render.Point{{X: tx, Y: innerPad}, {X: tx, Y: innerPad - tLen}}, 0.75*scale, charcoalInkColor, alpha)
			rc.StrokePolyline([]render.Point{{X: tx, Y: fh - innerPad}, {X: tx, Y: fh - innerPad + tLen}}, 0.75*scale, charcoalInkColor, alpha)
			idxTick++
		}
		idxTick = 0
		for ty := innerPad + tickStep*0.5; ty < fh-innerPad-2.0*scale; ty += tickStep * 0.5 {
			tLen := tickLenMinor
			alpha := 0.45
			if idxTick%2 == 0 {
				tLen = tickLenMajor
				alpha = 0.72
			}
			// Left & right border ticks
			rc.StrokePolyline([]render.Point{{X: innerPad, Y: ty}, {X: innerPad - tLen, Y: ty}}, 0.75*scale, charcoalInkColor, alpha)
			rc.StrokePolyline([]render.Point{{X: fw - innerPad, Y: ty}, {X: fw - innerPad + tLen, Y: ty}}, 0.75*scale, charcoalInkColor, alpha)
			idxTick++
		}
	}

	// 6. Subtle tactile archival paper grain & final merge
	rc.ApplyPaperTexture(0.10, seed)
	c.MergeRenderCanvas(rc)
}
