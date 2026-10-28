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

type squeegeeScrape struct {
	passes         int
	dropletCount   int
	scrapePressure float64
}

// NewSqueegeeScrape returns a squeegeeScrape engine reproducing a Gerhard Richter-style
// horizontal squeegee oil-paint reflection with diagonal rain/bristle striations and paint-droplet bokeh.
func NewSqueegeeScrape(passes, dropletCount int, scrapePressure float64) *squeegeeScrape {
	if passes <= 0 {
		passes = 8
	}
	if dropletCount <= 0 {
		dropletCount = 460
	}
	if scrapePressure <= 0 {
		scrapePressure = 1.0
	}
	return &squeegeeScrape{
		passes:         passes,
		dropletCount:   dropletCount,
		scrapePressure: scrapePressure,
	}
}

func adjustOklabLightness(col color.RGBA, dL float64) color.RGBA {
	lab := render.RGBAToOklab(col)
	lab.L = common.Constrain(lab.L+dL, 0.02, 0.99)
	return render.OklabToRGBA(lab, col.A)
}

func buildHorizontalKnifeStrip(x, y, width, height, taperRight float64) *canvas.Path {
	if width <= 0 || height <= 0 {
		return nil
	}
	hw := width * 0.5
	hh := height * 0.5
	hhRight := hh * (1.0 - 0.65*common.Clamp01(taperRight))
	p := &canvas.Path{}
	p.MoveTo(x-hw, y-hh)
	p.LineTo(x+hw, y-hhRight)
	p.LineTo(x+hw+height*0.6, y)
	p.LineTo(x+hw, y+hhRight)
	p.LineTo(x-hw, y+hh)
	p.LineTo(x-hw-height*0.35, y)
	p.Close()
	return p
}

func buildOrganicDropletPath(cx, cy, rx, ry, angle float64) *canvas.Path {
	const segs = 14
	cosA := math.Cos(angle)
	sinA := math.Sin(angle)
	p := &canvas.Path{}
	for i := 0; i < segs; i++ {
		theta := float64(i) * 2.0 * math.Pi / float64(segs)
		lx := math.Cos(theta) * rx
		ly := math.Sin(theta) * ry
		wx := cx + lx*cosA - ly*sinA
		wy := cy + lx*sinA + ly*cosA
		if i == 0 {
			p.MoveTo(wx, wy)
		} else {
			p.LineTo(wx, wy)
		}
	}
	p.Close()
	return p
}

// Generative renders the SqueegeeScrape composition.
func (ss *squeegeeScrape) Generative(c *generativeart.Canva) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	seed := c.Seed()
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	// Palette colors matching Reference Image 5
	asphaltCharcoal := color.RGBA{R: 28, G: 34, B: 38, A: 255}  // #1c2226
	slateGrey := color.RGBA{R: 52, G: 63, B: 70, A: 255}        // #343f46
	deepObsidian := color.RGBA{R: 16, G: 20, B: 23, A: 255}     // #101417
	coolSteel := color.RGBA{R: 88, G: 100, B: 108, A: 255}      // #58646c

	chartreuseGold := color.RGBA{R: 200, G: 212, B: 70, A: 255} // #c8d446
	lemonChartreuse := color.RGBA{R: 232, G: 240, B: 120, A: 255} // #e8f078
	darkOlive := color.RGBA{R: 66, G: 74, B: 36, A: 255}        // #424a24

	crimsonRed := color.RGBA{R: 184, G: 29, B: 24, A: 255}      // #b81d18
	brightVermilion := color.RGBA{R: 230, G: 59, B: 36, A: 255} // #e63b24
	deepMaroon := color.RGBA{R: 74, G: 16, B: 20, A: 255}       // #4a1014
	cadmiumGold := color.RGBA{R: 245, G: 199, B: 50, A: 255}    // #f5c732
	burntOrange := color.RGBA{R: 218, G: 104, B: 28, A: 255}    // #da681c

	silverWhite := color.RGBA{R: 237, G: 242, B: 240, A: 255}   // #edf2f0
	celadonGrey := color.RGBA{R: 124, G: 138, B: 128, A: 255}   // #7c8a80
	sageTaupe := color.RGBA{R: 92, G: 102, B: 88, A: 255}       // #5c6658
	turquoiseMint := color.RGBA{R: 50, G: 204, B: 184, A: 255}  // #32ccb8
	deepEmerald := color.RGBA{R: 22, G: 98, B: 84, A: 255}     // #166254

	rc.FillBackground(asphaltCharcoal)

	useCustom := false
	var customGrad *render.MultiStopGradient
	if useCustom {
		customGrad = render.NewMultiStopGradient(c.Opts().ColorSchema()...)
	}

	// Angle -36° diagonal rain/bristle vector: direction (cos(-36°), sin(-36°)) or (+cos 36°, +sin 36°)
	// Wait: in screen coordinates (Y down), top-left to bottom-right or top-right to bottom-left?
	// Looking at Image 5: the diagonal bristle striations run from upper-left toward lower-right / upper-right to lower-left.
	// Using perpendicular coordinate uPerp = nx*sin(36°) + ny*cos(36°) and along-coordinate uAlong = nx*cos(36°) - ny*sin(36°).
	cos36 := math.Cos(36.0 * math.Pi / 180.0)
	sin36 := math.Sin(36.0 * math.Pi / 180.0)

	// 1. Macro Underpainting & Diagonal Bristle Striations
	evalUnderpainting := func(nx, ny float64) color.RGBA {
		w1 := noise.FBM2D(nx*2.2+0.7, ny*2.2+1.3, 3, 2.0, 0.5) * 0.12
		w2 := noise.FBM2D(nx*5.0-1.4, ny*5.0+0.8, 2, 2.0, 0.5) * 0.05

		if useCustom && customGrad != nil {
			return customGrad.SampleOklab(common.Clamp01(0.5*nx + 0.5*ny + w1))
		}

		// Start with dark asphalt charcoal / slate-grey / obsidian base (upper-right & top-center dominant)
		skyMix := 0.5 + 0.5*noise.Simplex2D(nx*3.2+0.4, ny*3.2-0.9)
		col := render.LerpOklab(deepObsidian, asphaltCharcoal, common.Smoothstep(0.15, 0.75, skyMix))
		col = render.LerpOklab(col, slateGrey, common.Smoothstep(0.50, 0.92, skyMix)*0.65)

		// Upper-left & upper-mid patches (nx < 0.52, ny < 0.38): glowing chartreuse-yellow / olive-gold openings
		ulMask := common.Smoothstep(0.56, 0.16, nx+w1*0.8) * common.Smoothstep(0.42, 0.08, ny+w2)
		if ulMask > 0 {
			openRib := 0.5 + 0.5*math.Sin(ny*38.0+nx*4.0+w1*12.0)
			oliveBase := render.LerpOklab(darkOlive, chartreuseGold, common.Smoothstep(0.32, 0.78, openRib))
			oliveBase = render.LerpOklab(oliveBase, lemonChartreuse, common.Smoothstep(0.75, 0.96, openRib)*0.7)
			col = render.LerpOklab(col, oliveBase, ulMask*0.88)
		}

		// Center & Lower-Right Reflection Zone (0.30 < nx < 0.92, 0.42 < ny < 0.88):
		// metallic silver-white, cool celadon-olive grey, sage-taupe
		reflX := common.Smoothstep(0.24, 0.44, nx+w1*0.6) * common.Smoothstep(0.96, 0.74, nx+w1*0.5)
		reflY := common.Smoothstep(0.38, 0.50, ny+w2) * common.Smoothstep(0.92, 0.76, ny+w2)
		reflMask := reflX * reflY
		if reflMask > 0 {
			vertColumn := math.Exp(-math.Pow((nx-0.52)/0.16, 2.0))
			reflCol := render.LerpOklab(sageTaupe, celadonGrey, common.Smoothstep(0.15, 0.65, vertColumn+w2))
			reflCol = render.LerpOklab(reflCol, silverWhite, common.Smoothstep(0.45, 0.92, vertColumn+0.25*noise.Simplex2D(nx*4.0, ny*18.0))*0.82)
			col = render.LerpOklab(col, reflCol, reflMask*0.90)
		}

		// Bright turquoise-mint streaks around (0.15..0.48, 0.40..0.46)
		turqX := common.Smoothstep(0.10, 0.22, nx+w2) * common.Smoothstep(0.56, 0.38, nx+w2)
		turqY := math.Exp(-math.Pow((ny-0.435+w2*0.3)/0.042, 2.0))
		turqMask := turqX * turqY
		if turqMask > 0 {
			tCol := render.LerpOklab(deepEmerald, turquoiseMint, common.Smoothstep(0.2, 0.85, turqMask))
			col = render.LerpOklab(col, tCol, turqMask*0.92)
		}

		// Lower-left & bottom (nx < 0.45, ny > 0.35 and across bottom ny > 0.80):
		// intense crimson red, bright vermilion, deep maroon, and bursts of cadmium yellow-gold
		llLeft := common.Smoothstep(0.48, 0.16, nx+w1*0.9) * common.Smoothstep(0.30, 0.48, ny+w2)
		llBottom := common.Smoothstep(0.74, 0.88, ny+w1*0.5) * common.Smoothstep(0.96, 0.35, nx+w2)
		llMask := math.Max(llLeft, llBottom*0.88)
		if llMask > 0 {
			fireWave := 0.5 + 0.5*noise.FBM2D(nx*4.2+1.1, ny*14.0-0.7, 3, 2.0, 0.5)
			redCol := render.LerpOklab(deepMaroon, crimsonRed, common.Smoothstep(0.12, 0.52, fireWave))
			redCol = render.LerpOklab(redCol, brightVermilion, common.Smoothstep(0.48, 0.82, fireWave))
			// Cadmium yellow-gold pockets inside lower-left/mid-left
			goldPocket := common.Smoothstep(0.68, 0.92, fireWave) * common.Smoothstep(0.40, 0.10, nx) * common.Smoothstep(0.88, 0.45, ny)
			if goldPocket > 0 {
				redCol = render.LerpOklab(redCol, render.LerpOklab(burntOrange, cadmiumGold, 0.75), goldPocket*0.85)
			}
			col = render.LerpOklab(col, redCol, llMask*0.95)
		}

		// Fine diagonal bristle/rain striations at -36° across the entire underpainting
		uPerp := (nx*sin36 + ny*cos36) * 145.0
		uAlong := (nx*cos36 - ny*sin36) * 6.5
		striation := noise.FBM2D(uPerp, uAlong, 3, 2.1, 0.55)*0.65 +
			0.35*math.Sin(uPerp*2.6+noise.Simplex2D(uPerp*0.4, uAlong)*1.8)
		// Stronger visibility in the upper-right dark sky and mid-tones
		striationAmp := 0.048 + 0.035*common.Smoothstep(0.55, 0.15, ny)*common.Smoothstep(0.20, 0.65, nx)
		col = adjustOklabLightness(col, striation*striationAmp)

		return col
	}

	for py := 0; py < h; py++ {
		ny := float64(py) / fh
		for px := 0; px < w; px++ {
			nx := float64(px) / fw
			col := evalUnderpainting(nx, ny)
			rL, gL, bL, _ := render.ColorToLinearRGB(col)
			rc.SetPixelLinear(px, py, rL, gL, bL, 1.0)
		}
	}

	// Render 1,200 fine semi-transparent diagonal bristle strokes at -36° in dark slate, silver-grey, and olive
	diagDx := cos36
	diagDy := sin36
	for i := 0; i < 1200; i++ {
		x0 := (rng.Float64()*1.16 - 0.08) * fw
		y0 := (rng.Float64()*1.16 - 0.08) * fh
		nx := common.Clamp01(x0 / fw)
		ny := common.Clamp01(y0 / fh)

		sLen := (28.0 + 68.0*rng.Float64()) * scale
		x1 := x0 + diagDx*sLen
		y1 := y0 + diagDy*sLen

		var sCol color.RGBA
		pick := rng.Float64()
		switch {
		case ny < 0.48 && nx > 0.28:
			if pick < 0.45 {
				sCol = coolSteel
			} else if pick < 0.80 {
				sCol = deepObsidian
			} else {
				sCol = slateGrey
			}
		case nx < 0.42 && ny > 0.40:
			if pick < 0.5 {
				sCol = deepMaroon
			} else {
				sCol = crimsonRed
			}
		default:
			if pick < 0.45 {
				sCol = celadonGrey
			} else if pick < 0.78 {
				sCol = darkOlive
			} else {
				sCol = silverWhite
			}
		}

		hw := (0.32 + 0.42*rng.Float64()) * scale
		pts := []render.Point{render.Pt(x0, y0), render.Pt(x1, y1)}
		if p := buildBristleRibbonPath(pts, hw, 0, 0); p != nil {
			rc.RasterizePath(p, sCol, 0.14+0.18*rng.Float64())
		}
	}

	// 2. Horizontal Squeegee Blade Drag & Skip-Chatter Texture (Pixel + Vector Hybrid)
	// Evaluate pass-specific wet paint color and regional presence mask
	evalPassLayer := func(passIdx int, nx, ny, hNoise, comb float64) (col color.RGBA, regionBoost float64, threshold float64, blendAlpha float64) {
		pMod := passIdx % 8
		fineComb := 0.5 + 0.5*comb

		if useCustom && customGrad != nil {
			t := common.Clamp01(float64(pMod)/7.0*0.65 + 0.35*(0.5+0.5*hNoise))
			col = customGrad.SampleOklab(t)
			regionBoost = 0.25
			threshold = 0.05 + 0.04*float64(pMod)
			blendAlpha = 0.82
			return
		}

		switch pMod {
		case 0:
			// Pass 1: Deep maroon & dark asphalt charcoal horizontal scrape foundation
			cLeft := render.LerpOklab(deepMaroon, crimsonRed, common.Smoothstep(0.35, 0.75, ny)*0.65)
			cRight := render.LerpOklab(deepObsidian, asphaltCharcoal, fineComb*0.75)
			col = render.LerpOklab(cLeft, cRight, common.Smoothstep(0.22, 0.58, nx))
			regionBoost = 0.28 * common.Smoothstep(0.10, 0.85, ny)
			threshold = 0.02
			blendAlpha = 0.84

		case 1:
			// Pass 2: Fiery vermilion & cadmium yellow on the left/bottom, olive-grey & silver in center
			fireMix := common.Smoothstep(0.20, 0.85, 0.5+0.5*noise.Simplex2D(nx*4.5+2.1, ny*18.0))
			leftCol := render.LerpOklab(crimsonRed, brightVermilion, fireMix)
			goldBand := common.Smoothstep(0.62, 0.90, fireMix) * common.Smoothstep(0.78, 0.36, ny)
			leftCol = render.LerpOklab(leftCol, cadmiumGold, goldBand*0.85)

			midCol := render.LerpOklab(sageTaupe, celadonGrey, fineComb)
			col = render.LerpOklab(leftCol, midCol, common.Smoothstep(0.30, 0.55, nx))

			leftPresence := common.Smoothstep(0.52, 0.12, nx) * common.Smoothstep(0.28, 0.46, ny)
			botPresence := common.Smoothstep(0.76, 0.90, ny) * common.Smoothstep(0.92, 0.25, nx)
			regionBoost = 0.52 * math.Max(leftPresence, botPresence)
			threshold = 0.14
			blendAlpha = 0.90

		case 2:
			// Pass 3: Turquoise-mint & emerald horizontal ribbon around ny in [0.38, 0.48] + mid-left accents
			bandEnv := math.Exp(-math.Pow((ny-0.43)/0.048, 2.0)) * common.Smoothstep(0.08, 0.20, nx) * common.Smoothstep(0.66, 0.38, nx)
			secondaryTurq := math.Exp(-math.Pow((ny-0.55)/0.035, 2.0)) * common.Smoothstep(0.20, 0.35, nx) * common.Smoothstep(0.55, 0.38, nx) * 0.65
			presence := math.Max(bandEnv, secondaryTurq)

			col = render.LerpOklab(deepEmerald, turquoiseMint, common.Smoothstep(0.25, 0.82, fineComb))
			col = render.LerpOklab(col, silverWhite, common.Smoothstep(0.78, 0.98, fineComb)*0.45)
			regionBoost = 0.72 * presence
			threshold = 0.34
			blendAlpha = 0.88

		case 3:
			// Pass 4: Bright silver-white & pale celadon horizontal reflection column around nx in [0.32, 0.72], ny in [0.43, 0.86]
			colX := common.Smoothstep(0.26, 0.42, nx) * common.Smoothstep(0.84, 0.58, nx)
			colY := common.Smoothstep(0.40, 0.49, ny) * common.Smoothstep(0.90, 0.76, ny)
			presence := colX * colY

			// Bright vertical reflection core around nx ~ 0.48..0.58
			coreX := math.Exp(-math.Pow((nx-0.52)/0.13, 2.0))
			col = render.LerpOklab(celadonGrey, silverWhite, common.Clamp01(coreX*0.85+fineComb*0.35))
			regionBoost = 0.62 * presence
			threshold = 0.22
			blendAlpha = 0.92

		case 4:
			// Pass 5: Chartreuse-yellow & olive-gold horizontal glints in upper-left ny in [0.08, 0.38], nx in [0.04, 0.54]
			ulX := common.Smoothstep(0.02, 0.12, nx) * common.Smoothstep(0.56, 0.28, nx)
			ulY := common.Smoothstep(0.06, 0.14, ny) * common.Smoothstep(0.40, 0.22, ny)
			// Also a warm yellow-gold glint band in lower-left/mid-left
			llGold := common.Smoothstep(0.42, 0.10, nx) * math.Exp(-math.Pow((ny-0.52)/0.08, 2.0)) * 0.75
			presence := math.Max(ulX*ulY, llGold)

			if ny < 0.40 {
				col = render.LerpOklab(darkOlive, chartreuseGold, common.Smoothstep(0.18, 0.70, fineComb))
				col = render.LerpOklab(col, lemonChartreuse, common.Smoothstep(0.70, 0.96, fineComb)*0.65)
			} else {
				col = render.LerpOklab(burntOrange, cadmiumGold, common.Smoothstep(0.20, 0.85, fineComb))
			}
			regionBoost = 0.66 * presence
			threshold = 0.28
			blendAlpha = 0.88

		case 5:
			// Pass 6: Dark charcoal-olive & slate horizontal scrape bands (especially upper-right & mid-right)
			urPresence := common.Smoothstep(0.22, 0.52, nx) * common.Smoothstep(0.62, 0.12, ny)
			midBand := math.Exp(-math.Pow((ny-0.38)/0.07, 2.0)) * common.Smoothstep(0.35, 0.75, nx)
			presence := math.Max(urPresence, midBand)

			col = render.LerpOklab(deepObsidian, asphaltCharcoal, fineComb*0.7)
			col = render.LerpOklab(col, darkOlive, common.Smoothstep(0.65, 0.25, nx)*0.45)
			regionBoost = 0.45 * presence
			threshold = 0.18
			blendAlpha = 0.82

		case 6:
			// Pass 7: Crisp vermilion & crimson horizontal drag ribbons across lower-left and bottom edge
			botSweep := common.Smoothstep(0.72, 0.88, ny) * common.Smoothstep(0.95, 0.15, nx)
			leftSweep := common.Smoothstep(0.46, 0.08, nx) * common.Smoothstep(0.36, 0.65, ny)
			presence := math.Max(botSweep, leftSweep)

			col = render.LerpOklab(crimsonRed, brightVermilion, common.Smoothstep(0.25, 0.85, fineComb))
			if nx < 0.28 && fineComb > 0.74 {
				col = render.LerpOklab(col, cadmiumGold, (fineComb-0.74)*2.8)
			}
			regionBoost = 0.58 * presence
			threshold = 0.25
			blendAlpha = 0.90

		default:
			// Pass 8: Crisp horizontal palette-knife specular reflection streaks in silver-white, turquoise & gold
			reflSweep := common.Smoothstep(0.30, 0.48, nx) * common.Smoothstep(0.86, 0.56, nx) *
				common.Smoothstep(0.42, 0.52, ny) * common.Smoothstep(0.86, 0.68, ny)
			col = render.LerpOklab(celadonGrey, silverWhite, common.Smoothstep(0.20, 0.80, fineComb))
			regionBoost = 0.54 * reflSweep
			threshold = 0.30
			blendAlpha = 0.92
		}
		return
	}

	pressureScale := common.Constrain(ss.scrapePressure, 0.35, 2.5)
	for pIdx := 0; pIdx < ss.passes; pIdx++ {
		pf := float64(pIdx)
		seedOffX := pf*3.7 + 1.1
		seedOffY := pf*7.3 + 2.9

		for py := 0; py < h; py++ {
			ny := float64(py) / fh
			nyPrev := float64(py-1) / fh
			nyNext := float64(py+1) / fh

			for px := 0; px < w; px++ {
				nx := float64(px) / fw

				// Horizontally-stretched anisotropic FBM + fine horizontal comb striations
				hNoise := noise.FBM2D(nx*1.8+seedOffX, ny*28.0+seedOffY, 4, 2.1, 0.55)
				comb := math.Sin(float64(py)*1.7 + pf*1.9 + noise.Simplex2D(nx*6.0+seedOffX, ny*40.0+seedOffY)*2.5)

				passCol, regionBoost, thresh, bAlpha := evalPassLayer(pIdx, nx, ny, hNoise, comb)
				if regionBoost <= 0.01 {
					continue
				}

				effThresh := thresh - (regionBoost-0.25)*0.65*pressureScale
				contact := hNoise + 0.18*comb
				margin := contact - effThresh
				if margin <= 0 {
					continue
				}

				// Evaluate vertical neighbors to detect scraped horizontal palette-knife ridge edges
				hPrev := noise.FBM2D(nx*1.8+seedOffX, nyPrev*28.0+seedOffY, 4, 2.1, 0.55)
				combPrev := math.Sin(float64(py-1)*1.7 + pf*1.9 + noise.Simplex2D(nx*6.0+seedOffX, nyPrev*40.0+seedOffY)*2.5)
				contactPrev := hPrev + 0.18*combPrev

				hNext := noise.FBM2D(nx*1.8+seedOffX, nyNext*28.0+seedOffY, 4, 2.1, 0.55)
				combNext := math.Sin(float64(py+1)*1.7 + pf*1.9 + noise.Simplex2D(nx*6.0+seedOffX, nyNext*40.0+seedOffY)*2.5)
				contactNext := hNext + 0.18*combNext

				outCol := passCol
				// Subtle internal horizontal bristle modulation
				outCol = adjustOklabLightness(outCol, comb*0.035)

				if contactPrev <= effThresh {
					// Upper transition boundary: specular palette-knife impasto ridge (+0.18 L)
					outCol = adjustOklabLightness(outCol, 0.18)
				} else if contactNext <= effThresh {
					// Lower transition boundary: dark horizontal crevice shadow (-0.22 L)
					outCol = adjustOklabLightness(outCol, -0.22)
				}

				cov := common.Smoothstep(0.0, 0.08, margin) * bAlpha
				rc.BlendPixel(px, py, outCol, cov)
			}
		}
	}

	// Render 750 crisp horizontal palette-knife streaks (width = 18..95 px, height = 1.2..4.5 px)
	for i := 0; i < 750; i++ {
		nx := rng.Float64()
		ny := rng.Float64()
		x := nx * fw
		y := ny * fh

		stripW := (18.0 + 77.0*math.Pow(rng.Float64(), 1.3)) * scale
		stripH := (1.2 + 3.3*math.Pow(rng.Float64(), 1.6)) * scale
		taperR := 0.3 + 0.7*rng.Float64()

		var knifeCol color.RGBA
		rPick := rng.Float64()
		if useCustom && customGrad != nil {
			knifeCol = customGrad.SampleOklab(common.Clamp01(0.5*nx + 0.5*ny + (rPick-0.5)*0.25))
		} else {
			switch {
			case ny > 0.38 && ny < 0.48 && nx > 0.12 && nx < 0.58 && rPick < 0.68:
				// Turquoise-mint & emerald knife streaks
				knifeCol = render.LerpOklab(deepEmerald, turquoiseMint, 0.35+0.65*rng.Float64())
			case nx > 0.30 && nx < 0.85 && ny > 0.42 && ny < 0.88 && rPick < 0.75:
				// Pure white / silver-white / celadon reflection streaks
				if rPick < 0.42 {
					knifeCol = silverWhite
				} else {
					knifeCol = render.LerpOklab(celadonGrey, silverWhite, rng.Float64()*0.7)
				}
			case ny < 0.38 && nx < 0.52 && rPick < 0.62:
				// Chartreuse-yellow / olive-gold upper-left horizontal glints
				knifeCol = render.LerpOklab(chartreuseGold, lemonChartreuse, rng.Float64()*0.75)
			case (nx < 0.46 && ny > 0.34) || ny > 0.78:
				// Vermilion, crimson, and cadmium yellow-gold knife streaks
				if rPick < 0.28 {
					knifeCol = cadmiumGold
				} else if rPick < 0.72 {
					knifeCol = brightVermilion
				} else {
					knifeCol = crimsonRed
				}
			default:
				// Dark asphalt charcoal / slate / olive knife streaks
				if rPick < 0.55 {
					knifeCol = asphaltCharcoal
				} else {
					knifeCol = slateGrey
				}
			}
		}

		// Draw dark bottom crevice shadow just under the knife streak, then the crisp lit knife ridge
		shadowCol := adjustOklabLightness(knifeCol, -0.24)
		if sp := buildHorizontalKnifeStrip(x, y+0.85*scale, stripW, stripH*0.85, taperR); sp != nil {
			rc.RasterizePath(sp, shadowCol, 0.52)
		}
		if kp := buildHorizontalKnifeStrip(x, y, stripW, stripH, taperR); kp != nil {
			rc.RasterizePath(kp, knifeCol, 0.78+0.18*rng.Float64())
		}
		// Razor-thin top specular highlight edge on thicker knife ridges
		if stripH > 2.2*scale {
			hiCol := adjustOklabLightness(knifeCol, 0.16)
			if hp := buildHorizontalKnifeStrip(x, y-stripH*0.35, stripW*0.92, 0.85*scale, taperR); hp != nil {
				rc.RasterizePath(hp, hiCol, 0.68)
			}
		}
	}

	// 3. Paint-Droplet / Water-Bead Bokeh Splatter (dropletCount = 460)
	ivoryWhite := color.RGBA{R: 246, G: 244, B: 236, A: 255}  // #f6f4ec (68%)
	rosePink := color.RGBA{R: 242, G: 182, B: 176, A: 255}    // #f2b6b0 (16%)
	paleLemon := color.RGBA{R: 238, G: 240, B: 152, A: 255}   // #eef098 (9%)
	paleAqua := color.RGBA{R: 174, G: 236, B: 228, A: 255}    // #aeece4 (7%)
	pureWhite := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	shadowRim := color.RGBA{R: 14, G: 18, B: 20, A: 255}

	for i := 0; i < ss.dropletCount; i++ {
		var nx, ny float64
		branch := rng.Float64()
		switch {
		case branch < 0.34:
			// Swath 1: Top-center (0.35, 0.02) through mid-left (0.15, 0.35)
			t := rng.Float64()
			nx = common.Lerp(0.36, 0.14, t) + rng.NormFloat64()*0.11
			ny = common.Lerp(0.03, 0.38, t) + rng.NormFloat64()*0.09
		case branch < 0.72:
			// Swath 2: Mid-left (0.18, 0.36) across center/bottom-right (0.58, 0.78)
			t := rng.Float64()
			nx = common.Lerp(0.18, 0.66, t) + rng.NormFloat64()*0.14
			ny = common.Lerp(0.36, 0.82, t) + rng.NormFloat64()*0.12
		default:
			// Scattered across the entire canvas
			nx = 0.02 + 0.96*rng.Float64()
			ny = 0.02 + 0.96*rng.Float64()
		}
		nx = common.Constrain(nx, 0.015, 0.985)
		ny = common.Constrain(ny, 0.015, 0.985)
		dx := nx * fw
		dy := ny * fh

		r := (1.1 + 3.6*math.Pow(rng.Float64(), 1.8)) * scale

		var dropCol color.RGBA
		cPick := rng.Float64()
		switch {
		case cPick < 0.68:
			dropCol = ivoryWhite
		case cPick < 0.84:
			dropCol = rosePink
		case cPick < 0.93:
			dropCol = paleLemon
		default:
			dropCol = paleAqua
		}

		// Slightly organic/elliptical shape (or short horizontal/diagonal dash for larger drops)
		aspect := 1.0 + 0.35*rng.Float64()
		if r > 2.6*scale && rng.Float64() < 0.42 {
			aspect = 1.45 + 0.65*rng.Float64()
		}
		rx := r * aspect
		ry := r / math.Sqrt(aspect)
		ang := (rng.Float64() - 0.5) * 0.45

		// Subtle dark contact shadow rim (r * 1.18, alpha 0.45) offset by (0.4*scale, 0.6*scale)
		if sp := buildOrganicDropletPath(dx+0.4*scale, dy+0.6*scale, rx*1.18, ry*1.18, ang); sp != nil {
			rc.RasterizePath(sp, shadowRim, 0.45)
		}
		// Bright opaque droplet body
		if bp := buildOrganicDropletPath(dx, dy, rx, ry, ang); bp != nil {
			rc.RasterizePath(bp, dropCol, 0.94)
		}
		// Pure-white specular highlight core
		rc.DrawGlowDot(dx-0.22*r, dy-0.22*r, math.Max(0.75*scale, r*0.48), pureWhite, 0.96, 0.88)
	}

	c.MergeRenderCanvas(rc)
}
