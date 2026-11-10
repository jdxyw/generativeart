package arts

import (
	"image/color"
	"math"
	"math/rand"
	"runtime"
	"sort"
	"sync"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type auroraCurtainFold struct {
	curtains int
	rays     int
	foldCurl float64
}

// NewAuroraCurtainFold returns an auroraCurtainFold engine that synthesizes ultra-fine
// 3D Kelvin-Helmholtz folded Polar Aurora Borealis draperies, Lofoten snow-carved alpine
// fjord peaks, and a continuous per-pixel capillary water reflection.
func NewAuroraCurtainFold(curtains, rays int, foldCurl float64) *auroraCurtainFold {
	if curtains <= 0 {
		curtains = 4
	}
	if rays <= 0 {
		rays = 180
	}
	if foldCurl <= 0 {
		foldCurl = 1.15
	}
	return &auroraCurtainFold{
		curtains: common.ConstrainInt(curtains, 2, 7),
		rays:     common.ConstrainInt(rays, 80, 320),
		foldCurl: common.Constrain(foldCurl, 0.4, 2.4),
	}
}

type auroraMicroFilament struct {
	xBot, yBot float64
	xTop, yTop float64
	zDepth     float64
	width      float64
	energy     float64
	curtainID  int
	uParam     float64
}

// Generative renders a static frame of the 3D Polar Aurora Borealis & Lofoten Fjord.
func (acf *auroraCurtainFold) Generative(c *generativeart.Canva) {
	acf.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360-degree phase loop (phase in [0, 1)).
func (acf *auroraCurtainFold) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	minDim := math.Min(fw, fh)
	scale := minDim / 600.0
	phi := 2.0 * math.Pi * phase

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	seed := c.Seed()
	if seed == 0 {
		seed = 20262704
	}
	rng := rand.New(rand.NewSource(seed))
	ne := common.NewNoiseEngine(seed + 2704)

	// Palette: 100km Nitrogen Crimson-Magenta, 150km Oxygen Emerald-Green, 220km Electric Turquoise, 300km Ionospheric Violet
	zenithSpace := color.RGBA{R: 2, G: 5, B: 15, A: 255}
	horizonIndigo := color.RGBA{R: 8, G: 22, B: 48, A: 255}
	nitrogenMagenta := color.RGBA{R: 248, G: 46, B: 138, A: 255}
	oxygenEmerald := color.RGBA{R: 22, G: 250, B: 146, A: 255}
	auroralCyan := color.RGBA{R: 38, G: 220, B: 248, A: 255}
	highViolet := color.RGBA{R: 122, G: 48, B: 242, A: 255}
	auroralWhite := color.RGBA{R: 238, G: 255, B: 246, A: 255}

	schema := c.Opts().ColorSchema()
	if len(schema) > 0 {
		uGrad := render.NewMultiStopGradient(schema...)
		nitrogenMagenta = render.LerpOklab(nitrogenMagenta, uGrad.SampleOklab(0.20), 0.16)
		oxygenEmerald = render.LerpOklab(oxygenEmerald, uGrad.SampleOklab(0.55), 0.16)
		highViolet = render.LerpOklab(highViolet, uGrad.SampleOklab(0.85), 0.16)
	}

	waterlineY := 0.72 * fh
	yWaterInt := int(math.Floor(waterlineY))

	// 1. Arctic Night Sky & Milky Way Starfield above waterlineY
	zenR, zenG, zenB, _ := render.ColorToLinearRGB(zenithSpace)
	horR, horG, horB, _ := render.ColorToLinearRGB(horizonIndigo)
	emR, emG, emB, _ := render.ColorToLinearRGB(oxygenEmerald)
	cyaR, _, cyaB, _ := render.ColorToLinearRGB(auroralCyan)

	for y := 0; y < h; y++ {
		fy := float64(y) + 0.5
		var tSky float64
		if fy <= waterlineY {
			tSky = math.Pow(fy/waterlineY, 1.45)
		} else {
			tSky = math.Pow(1.0-(fy-waterlineY)/(fh-waterlineY), 1.3) * 0.75
		}
		for x := 0; x < w; x++ {
			fx := float64(x) + 0.5
			ux := fx / fw
			// Subtle high-altitude ionospheric airglow arch
			airglow := math.Exp(-math.Pow((fy/waterlineY-0.58-0.12*math.Sin(ux*math.Pi))/0.28, 2.0)) * 0.08
			rL := zenR + (horR-zenR)*tSky + cyaR*airglow*0.4
			gL := zenG + (horG-zenG)*tSky + emG*airglow*0.8
			bL := zenB + (horB-zenB)*tSky + cyaB*airglow*0.7
			rc.SetPixelLinear(x, y, rL, gL, bL, 1.0)
		}
	}

	// Crisp multi-magnitude Arctic stars & constellations
	rc.SetBlendMode(render.BlendAdd)
	for s := 0; s < 160; s++ {
		fs := float64(s)
		sx := math.Mod(fs*0.754877+0.09*rng.Float64(), 1.0) * fw
		sy := math.Mod(fs*0.569840+0.09*rng.Float64(), 1.0) * (waterlineY * 0.84)
		sRad := (0.45 + 1.15*math.Pow(rng.Float64(), 3.0)) * scale
		twinkle := 0.55 + 0.45*math.Sin(phi+fs*1.9)
		sCol := auroralWhite
		if s%5 == 0 {
			sCol = auroralCyan
		}
		rc.DrawGlowDot(sx, sy, sRad, sCol, 0.78*twinkle, 0.85)
	}

	// 2. Continuous Sub-Pixel 3D Volumetric Aurora Borealis Draperies
	// Each curtain is integrated across 3,000 sub-pixel arc slices u in [0, 1] with a 5-tap
	// horizontal Gaussian kernel and continuous vertical scale-height emission decay I(v),
	// completely eliminating discrete polygon comb lines while naturally forming bright
	// volumetric caustic pillars wherever the 3D Kelvin-Helmholtz ribbon folds tangent to the view!
	numCurtains := acf.curtains
	slicesPerCurtain := common.MaxInt(2800, acf.rays*16)
	foldAmp := acf.foldCurl
	zenithVanishX := 0.52 * fw
	zenithVanishY := -0.48 * fh

	filaments := make([]auroraMicroFilament, 0, numCurtains*slicesPerCurtain)

	for cIdx := 0; cIdx < numCurtains; cIdx++ {
		fc := float64(cIdx) / math.Max(1.0, float64(numCurtains-1))
		// Stagger curtains from high sweeping background arches down to foreground S-ribbons
		baseYBot := (0.39 + 0.11*fc) * fh
		baseSpanH := (0.34 - 0.03*fc) * fh
		cSeed := float64(cIdx)*2.09 + 0.4

		var prevXBot float64
		for rIdx := 0; rIdx < slicesPerCurtain; rIdx++ {
			u := float64(rIdx) / float64(slicesPerCurtain-1)

			// Smooth 3D Kelvin-Helmholtz serpentine spiral fold in (X, Z) space
			arg1 := u*math.Pi*2.4 + cSeed - phi*0.5
			arg2 := u*math.Pi*5.2 - cSeed*0.7 + phi
			arg3 := u*math.Pi*9.6 + cSeed*1.2 - phi

			foldX := (u-0.49)*1.12 +
				0.21*foldAmp*math.Sin(arg1) +
				0.09*foldAmp*math.Sin(arg2) +
				0.035*foldAmp*math.Cos(arg3)
			zDepth := float64(cIdx)*0.60 + 0.36*math.Cos(arg1) + 0.16*math.Cos(arg2)

			xBot := (0.50 + foldX*0.78) * fw
			yWave := (26.0*math.Sin(arg1+0.45) + 12.0*math.Cos(arg2-0.35) - 20.0*(u-0.5)) * scale
			yBot := baseYBot + yWave + zDepth*10.0*scale

			dxSlice := 1.0
			if rIdx > 0 {
				dxSlice = math.Abs(xBot-prevXBot) / scale
			}
			prevXBot = xBot
			// Regularized Jacobian weight: keeps broad sweeping sheets luminous while fold turnbacks stack ~3x brighter
			jacobianBoost := math.Pow(math.Sqrt(dxSlice*dxSlice+0.20*0.20), 0.75)

			// Continuous field-aligned magnetic ray intensity modulation along the curtain arc u
			fineRay := 0.5 + 0.5*math.Sin(u*math.Pi*52.0+cSeed*3.7-phi*2.0)
			medRay := 0.5 + 0.5*math.Cos(u*math.Pi*16.0-cSeed*2.1+phi)
			coarseArch := 0.65 + 0.35*math.Sin(u*math.Pi*3.0+cSeed)
			rayMod := 0.22 + 0.52*math.Pow(fineRay, 2.1) + 0.26*medRay

			// Smooth crown height envelope
			rayH := baseSpanH * (0.72 + 0.24*coarseArch + 0.15*medRay + 0.08*fineRay)
			dxVan := zenithVanishX - xBot
			dyVan := zenithVanishY - yBot
			vanLen := math.Hypot(dxVan, dyVan) + 1e-6
			xTop := xBot + (dxVan/vanLen)*rayH*0.25
			yTop := math.Max(14.0*scale, yBot-rayH)

			endTaper := math.Pow(math.Sin(math.Pi*common.Constrain(u, 0.005, 0.995)), 0.75)

			filaments = append(filaments, auroraMicroFilament{
				xBot:      xBot,
				yBot:      yBot,
				xTop:      xTop,
				yTop:      yTop,
				zDepth:    zDepth,
				width:     (1.45 + 0.70*medRay) * scale,
				energy:    (0.078 + 0.168*rayMod) * jacobianBoost * endTaper,
				curtainID: cIdx,
				uParam:    u,
			})
		}
	}

	sort.Slice(filaments, func(i, j int) bool {
		return filaments[i].zDepth < filaments[j].zDepth
	})

	magR, magG, magB, _ := render.ColorToLinearRGB(nitrogenMagenta)
	whR, whG, whB, _ := render.ColorToLinearRGB(auroralWhite)
	_, cyaG, _, _ := render.ColorToLinearRGB(auroralCyan)
	vioR, vioG, vioB, _ := render.ColorToLinearRGB(highViolet)

	// Exact scanline-continuous sub-pixel Gaussian sheet accumulation into linear HDR sky buffer
	for i := range filaments {
		f := &filaments[i]
		if f.xBot < -20*scale || f.xBot > fw+20*scale {
			continue
		}
		spanY := f.yBot - f.yTop
		if spanY < 2.0 {
			continue
		}
		invSpanY := 1.0 / spanY
		dx := f.xTop - f.xBot
		sigmaX := math.Max(0.85*scale, f.width*0.65)
		invTwoSigmaSq := 1.0 / (2.0 * sigmaX * sigmaX)
		normFactor := 1.0 / (2.506628 * sigmaX)

		pyStart := common.MaxInt(0, int(math.Floor(f.yTop)))
		pyEnd := common.MinInt(yWaterInt-1, int(math.Ceil(f.yBot)))

		for py := pyStart; py <= pyEnd; py++ {
			fy := float64(py) + 0.5
			v := (f.yBot - fy) * invSpanY // exact continuous v in [0, 1] at scanline center fy
			if v <= 0.0 || v >= 1.0 {
				continue
			}
			pxCenter := f.xBot + dx*v

			// Physical ionospheric vertical luminosity profile:
			// Smooth lower attack -> luminous 100-120km lower border -> smooth (1-v)^1.55 scale-height fade
			botAttack := common.Smoothstep(0.0, 0.045, v)
			lowerCore := 0.78 + 1.15*math.Exp(-math.Pow((v-0.080)/0.080, 2.0))
			topFade := math.Pow(1.0-v, 1.55)
			vIntensity := f.energy * botAttack * lowerCore * topFade
			if vIntensity < 1e-5 {
				continue
			}

			// Altitude-dependent auroral emission spectrum in linear RGB
			var rCol, gCol, bCol float64
			if v < 0.095 {
				// Excited N2 lower crimson-pink fringe transitioning into brilliant white-green core
				uBot := v / 0.095
				rCol = magR*1.25*(1.0-uBot) + (emR*0.65+whR*0.35)*uBot
				gCol = magG*1.10*(1.0-uBot) + (emG*0.78+whG*0.22)*uBot
				bCol = magB*1.25*(1.0-uBot) + (emB*0.65+whB*0.35)*uBot
			} else if v < 0.48 {
				// Main 557.7nm atomic oxygen emerald curtain transitioning to aquamarine cyan
				uMid := (v - 0.095) / 0.385
				rCol = emR*(1.0-uMid*0.68) + cyaR*(uMid*0.68)
				gCol = emG*(1.0-uMid*0.45) + cyaG*(uMid*0.45)
				bCol = emB*(1.0-uMid*0.68) + cyaB*(uMid*0.68)
			} else {
				// High-altitude 630nm/427.8nm violet-magenta crown
				uTop := (v - 0.48) / 0.52
				topR, topG, topB := vioR*1.35, vioG*1.20, vioB*1.35
				if f.curtainID%2 == 1 {
					topR = (vioR*0.48 + magR*0.52) * 1.35
					topG = (vioG*0.48 + magG*0.52) * 1.15
					topB = (vioB*0.48 + magB*0.52) * 1.35
				}
				rCol = (emR*0.32+cyaR*0.68)*(1.0-uTop) + topR*uTop
				gCol = (emG*0.55+cyaG*0.45)*(1.0-uTop) + topG*uTop
				bCol = (emB*0.32+cyaB*0.68)*(1.0-uTop) + topB*uTop
			}

			pxBase := int(math.Floor(pxCenter))
			for tap := -2; tap <= 3; tap++ {
				px := pxBase + tap
				if px < 0 || px >= w {
					continue
				}
				distX := (float64(px) + 0.5) - pxCenter
				wTap := vIntensity * normFactor * math.Exp(-distX*distX*invTwoSigmaSq)
				rc.BlendPixelLinear(px, py, rCol, gCol, bCol, wTap, render.BlendAdd)
			}
		}
	}
	rc.SetBlendMode(render.BlendOver)

	// 3. Three Receding Layers of Lofoten Snow-Sculpted Alpine Fjord Peaks Along the Horizon!
	// Uses 2D diagonal pyramid ridge-spur geometry so snow couloirs & granite facets sweep
	// diagonally down the flanks of each peak rather than forming vertical stripes!
	distPeakCol := linRGB3(16, 36, 70)
	midPeakDark := linRGB3(7, 14, 30)
	midPeakLit := linRGB3(24, 54, 90)
	snowShadow := linRGB3(68, 124, 174)
	snowAuroraLit := linRGB3(152, 240, 214)
	foreHeadland := linRGB3(4, 7, 15)

	evalMountainRange := func(layer int, ux float64) float64 {
		switch layer {
		case 0:
			env := 0.25 + 0.65*math.Exp(-math.Pow((ux-0.48)/0.32, 2.0)) +
				0.55*math.Exp(-math.Pow((ux-0.15)/0.22, 2.0))
			rdg := ne.Ridged2D(ux*2.8+1.1, 1.4, 4, 2.1, 0.50)
			return waterlineY - (0.03+0.13*env*(0.40+0.60*rdg))*fh
		case 1:
			env := 0.14 + 0.90*math.Exp(-math.Pow((ux-0.19)/0.20, 2.0)) +
				0.98*math.Exp(-math.Pow((ux-0.81)/0.22, 2.0)) +
				0.36*math.Exp(-math.Pow((ux-0.51)/0.14, 2.0))
			rdg := ne.Ridged2D(ux*3.4+2.7, 3.2, 4, 2.10, 0.50)
			return waterlineY - (0.02+0.20*env*(0.36+0.64*rdg))*fh
		default:
			leftCape := common.Smoothstep(0.30, 0.0, ux)
			rightCape := common.Smoothstep(0.70, 1.0, ux)
			cape := math.Max(leftCape, rightCape)
			rdg := 0.5 + 0.5*ne.Perlin2D(ux*8.0+4.1, 5.3)
			return waterlineY + 0.01*fh - (0.042*cape*(0.5+0.5*rdg))*fh
		}
	}

	for layer := 0; layer < 3; layer++ {
		for x := 0; x < w; x++ {
			ux := float64(x) / fw
			cY := evalMountainRange(layer, ux)
			// Macro-scale flank orientation (smooth over 3.5% canvas width so it never chatters pixel-to-pixel)
			cYLeftMacro := evalMountainRange(layer, ux-0.018)
			cYRightMacro := evalMountainRange(layer, ux+0.018)
			macroSlope := (cYRightMacro - cYLeftMacro) / (0.036 * fw)

			yStart := common.MaxInt(0, int(math.Floor(cY)))
			yEnd := common.MinInt(yWaterInt, h-1)
			for y := yStart; y <= yEnd; y++ {
				fy := float64(y) + 0.5
				if fy < cY {
					continue
				}
				vMtn := (fy - cY) / math.Max(1.0, waterlineY-cY)
				dyNorm := (fy - cY) / fh

				if layer == 0 {
					haze := 0.52 + 0.48*vMtn
					rM := distPeakCol[0]*(1.0-0.3*haze) + cyaR*0.18*haze
					gM := distPeakCol[1]*(1.0-0.3*haze) + emG*0.22*haze
					bM := distPeakCol[2]*(1.0-0.3*haze) + cyaB*0.22*haze
					rc.SetPixelLinear(x, y, rM, gM, bM, 1.0)
				} else if layer == 1 {
					// 2D diagonal alpine ridge spurs & snow couloirs descending outward from each summit
					spurShear := math.Tanh(macroSlope*0.65) * dyNorm * 0.42
					uxSpur := ux - spurShear
					nTex := ne.FBM2D(uxSpur*11.0+1.7, dyNorm*11.0+2.3, 4, 2.05, 0.52)
					spurWave := math.Sin((uxSpur*10.5+nTex*0.55)*math.Pi)

					// Combine macro peak flank lighting (left-facing snow slope vs right-facing shadow wall) with 2D couloir relief
					flankLight := common.Smoothstep(-0.95, 0.85, macroSlope*0.75+spurWave*0.55+nTex*0.35)
					altSnow := common.Smoothstep(0.95, 0.05, vMtn)
					snowMask := common.Constrain(altSnow*(0.28+0.72*math.Pow(flankLight, 1.25)), 0.0, 1.0)

					rockR := midPeakDark[0]*(1.0-flankLight*0.55) + midPeakLit[0]*flankLight*0.55
					rockG := midPeakDark[1]*(1.0-flankLight*0.55) + midPeakLit[1]*flankLight*0.55
					rockB := midPeakDark[2]*(1.0-flankLight*0.55) + midPeakLit[2]*flankLight*0.55

					snR := snowShadow[0]*(1.0-flankLight*0.68) + snowAuroraLit[0]*flankLight*0.68
					snG := snowShadow[1]*(1.0-flankLight*0.68) + snowAuroraLit[1]*flankLight*0.68
					snB := snowShadow[2]*(1.0-flankLight*0.68) + snowAuroraLit[2]*flankLight*0.68

					rimGlow := math.Exp(-(fy - cY) / (2.2 * scale))
					rM := rockR*(1.0-snowMask*0.80) + snR*snowMask*0.80 + emR*0.26*rimGlow
					gM := rockG*(1.0-snowMask*0.80) + snG*snowMask*0.80 + emG*0.34*rimGlow
					bM := rockB*(1.0-snowMask*0.80) + snB*snowMask*0.80 + emB*0.28*rimGlow
					rc.SetPixelLinear(x, y, rM, gM, bM, 1.0)
				} else {
					rc.SetPixelLinear(x, y, foreHeadland[0], foreHeadland[1], foreHeadland[2], 1.0)
				}
			}
		}
	}

	// 4. True Per-Pixel Capillary Water Reflection of the Sky, Aurora & Lofoten Peaks (y in [yWaterInt, h))!
	// Snapshot the upper sky+mountain image via rc.ResolveToneMapped into a scratch buffer or read directly:
	// Wait: we can resolve the current HDR canvas to an RGBA snapshot and sample it with linear conversion
	// so every single pixel of the aurora and snow peaks reflects in the rippling fjord water!
	skySnap := c.Img()
	rc.ResolveToneMapped(skySnap, 1.0, 2.2)

	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
	var wg sync.WaitGroup
	waterRows := h - yWaterInt
	chunkW := (waterRows + workers - 1) / workers

	for wk := 0; wk < workers; wk++ {
		ry0 := yWaterInt + wk*chunkW
		ry1 := ry0 + chunkW
		if ry0 >= h {
			break
		}
		if ry1 > h {
			ry1 = h
		}
		wg.Add(1)
		go func(startY, endY int) {
			defer wg.Done()
			for y := startY; y < endY; y++ {
				fy := float64(y) + 0.5
				wDepth := (fy - waterlineY) / math.Max(1.0, fh-waterlineY) // 0 at shoreline -> 1 at bottom foreground
				// Perspective-foreshortened capillary wave frequency (finer near shoreline, broader in foreground)
				freqY := 24.0 / (0.14 + 0.86*wDepth)

				for x := 0; x < w; x++ {
					fx := float64(x) + 0.5
					ux := fx / fw

					// Multi-harmonic horizontal water wave normal
					w1 := math.Sin(fy*freqY/scale + ux*14.0 - phi*2.0)
					w2 := math.Cos(fy*freqY*1.65/scale - ux*22.0 + phi*1.0)
					nWave := ne.Perlin2D(ux*12.0+0.3*math.Cos(phi), (fy-waterlineY)*0.22/scale)
					wave := 0.52*w1 + 0.28*w2 + 0.20*nWave

					// Horizontal & vertical reflection refraction offsets
					dxRefl := wave * (2.5 + 14.0*wDepth) * scale
					dyRefl := (fy - waterlineY) * (1.28 + 0.14*wave)

					// 3-tap vertical blur across the water ripple for authentic stretched column reflections!
					var sumR, sumG, sumB float64
					for tap := -1; tap <= 1; tap++ {
						sx := common.ConstrainInt(int(math.Round(fx+dxRefl+float64(tap)*1.2*scale)), 0, w-1)
						sy := common.ConstrainInt(int(math.Round(waterlineY-dyRefl+float64(tap)*3.5*scale)), 0, yWaterInt-1)
						cRGBA := skySnap.RGBAAt(sx, sy)
						lr, lg, lb, _ := render.ColorToLinearRGB(cRGBA)
						sumR += lr
						sumG += lg
						sumB += lb
					}
					sr, sg, sb := sumR/3.0, sumG/3.0, sumB/3.0

					// Schlick water Fresnel reflectance + ripple crest specular glint
					fresnel := 0.36 + 0.48*math.Pow(wDepth, 0.65)
					crestGlint := 1.0 + 0.35*math.Pow(math.Max(0.0, wave), 4.0)
					shoreShadow := common.Smoothstep(0.0, 0.06, wDepth)

					rW := (sr*fresnel*crestGlint + zenR*0.22) * (0.35 + 0.65*shoreShadow)
					gW := (sg*fresnel*crestGlint + zenG*0.28) * (0.35 + 0.65*shoreShadow)
					bW := (sb*fresnel*crestGlint + zenB*0.36) * (0.35 + 0.65*shoreShadow)
					rc.SetPixelLinear(x, y, rW, gW, bW, 1.0)
				}
			}
		}(ry0, ry1)
	}
	wg.Wait()

	// Subtle shoreline ice/foam rim along waterlineY
	rc.SetBlendMode(render.BlendAdd)
	rc.DrawTaperedSpline([]render.Point{
		{X: 0, Y: waterlineY},
		{X: 0.5 * fw, Y: waterlineY},
		{X: fw, Y: waterlineY},
	}, 1.2*scale, render.TaperBothEnds, func(t float64) color.RGBA { return auroralCyan }, 0.22)
	rc.SetBlendMode(render.BlendOver)

	rc.ApplyBloom(0.76, 4.0*scale, 0.18)
	rc.ApplyVignette(0.20)
	c.MergeRenderCanvas(rc)
}

func linRGB3(r, g, b uint8) [3]float64 {
	lr, lg, lb, _ := render.ColorToLinearRGB(color.RGBA{R: r, G: g, B: b, A: 255})
	return [3]float64{lr, lg, lb}
}
