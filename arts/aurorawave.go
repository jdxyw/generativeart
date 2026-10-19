package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type auroraWave struct {
	curtainNum int
	rayNum     int
	turbulence float64
}

// NewAuroraWave returns an auroraWave engine that renders a magnetospheric Aurora Borealis
// nightscape with folding 3D geomagnetic ray curtains, ridged alpine peaks, and a rippled lake reflection.
func NewAuroraWave(curtainNum, rayNum int, turbulence float64) *auroraWave {
	if curtainNum <= 0 {
		curtainNum = 4
	}
	if rayNum <= 0 {
		rayNum = 220
	}
	if turbulence <= 0 {
		turbulence = 1.0
	}
	return &auroraWave{
		curtainNum: curtainNum,
		rayNum:     rayNum,
		turbulence: turbulence,
	}
}

// Generative renders the Aurora Borealis nightscape, alpine mountain silhouettes, and lake reflection.
func (aw *auroraWave) Generative(c *generativeart.Canva) {
	curtainNum := aw.curtainNum
	if curtainNum <= 0 {
		curtainNum = 4
	}
	rayNum := aw.rayNum
	if rayNum <= 0 {
		rayNum = 220
	}
	turbulence := aw.turbulence
	if turbulence <= 0 {
		turbulence = 1.0
	}

	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	seed := rand.Int63()
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	rc := c.NewRenderCanvas()
	yHorizon := 0.68 * fh
	horizonRow := common.ConstrainInt(int(math.Round(yHorizon)), 1, h-1)

	// 1. Deep twilight/midnight sky gradient from zenith to horizon (y_horizon = 0.68 * h)
	zenithCol := color.RGBA{R: 3, G: 6, B: 18, A: 255}
	midSkyCol := color.RGBA{R: 7, G: 18, B: 38, A: 255}
	horizonGlowCol := color.RGBA{R: 14, G: 38, B: 56, A: 255}

	schema := c.Opts().ColorSchema()
	if len(schema) >= 2 {
		userGrad := render.NewMultiStopGradient(schema...)
		horizonGlowCol = render.LerpOklab(horizonGlowCol, userGrad.SampleOklab(0.25), 0.18)
	}
	skyGrad := render.NewMultiStopGradient(zenithCol, midSkyCol, horizonGlowCol)

	for y := 0; y < h; y++ {
		var skyT float64
		if y <= horizonRow {
			skyT = float64(y) / math.Max(1.0, float64(horizonRow))
		} else {
			skyT = 1.0 - (float64(y-horizonRow) / math.Max(1.0, float64(h-horizonRow)))
		}
		col := skyGrad.SampleOklab(math.Pow(common.Clamp01(skyT), 1.15))
		rLin, gLin, bLin, _ := render.ColorToLinearRGB(col)
		for x := 0; x < w; x++ {
			rc.SetPixelLinear(x, y, rLin, gLin, bLin, 1.0)
		}
	}

	// 2. Crisp starfield + subtle circumpolar star arcs in the upper sky using SplatPoint and DrawGlowDot
	rc.SetBlendMode(render.BlendAdd)
	poleX := fw * (0.22 + 0.12*rng.Float64())
	poleY := fh * (0.12 + 0.06*rng.Float64())

	starCount := 320
	for i := 0; i < starCount; i++ {
		sx := rng.Float64() * fw
		sy := math.Pow(rng.Float64(), 1.25) * (yHorizon * 0.92)
		altFade := 1.0 - 0.65*(sy/yHorizon)

		warmth := rng.Float64()
		starCol := render.LerpOklab(
			color.RGBA{R: 195, G: 222, B: 255, A: 255},
			color.RGBA{R: 255, G: 238, B: 210, A: 255},
			warmth,
		)
		if i%16 == 0 {
			rad := (1.1 + 1.5*rng.Float64()) * scale
			rc.DrawGlowDot(sx, sy, rad, starCol, (0.38+0.45*rng.Float64())*altFade, 0.45)
		} else {
			rc.SplatPoint(sx, sy, starCol, (0.22+0.55*rng.Float64())*altFade)
		}
	}

	// Subtle circumpolar star trail arcs around the celestial pole
	arcCount := 26
	for i := 0; i < arcCount; i++ {
		rad := (28.0 + float64(i)*14.5 + rng.Float64()*8.0) * scale
		theta0 := rng.Float64() * 2.0 * math.Pi
		span := 0.08 + 0.14*rng.Float64()
		steps := common.MaxInt(12, int(rad*span*0.8))
		trailCol := color.RGBA{R: 185, G: 215, B: 250, A: 255}
		for s := 0; s <= steps; s++ {
			u := float64(s) / float64(steps)
			ang := theta0 + u*span
			px := poleX + rad*math.Cos(ang)
			py := poleY + rad*math.Sin(ang)
			if px < 0 || px >= fw || py < 0 || py >= yHorizon*0.86 {
				continue
			}
			env := math.Sin(u * math.Pi)
			rc.SplatPoint(px, py, trailCol, 0.08*env)
		}
	}

	// 3. Above the horizon, drape curtainNum folding 3D aurora ribbons across the sky
	// Color transitions in Oklab space along altitude u in [0, 1]:
	// - u = 0..0.30: bright luminous emerald / cyan at the lower ionization edge
	// - u = 0.30..0.65: jade / teal in the middle
	// - u = 0.65..1.00: magenta / violet / crimson at the high-altitude fringe
	lowEmerald := color.RGBA{R: 45, G: 255, B: 165, A: 255}
	lowCyan := color.RGBA{R: 55, G: 245, B: 215, A: 255}
	midJade := color.RGBA{R: 24, G: 195, B: 155, A: 255}
	midTeal := color.RGBA{R: 38, G: 145, B: 185, A: 255}
	highViolet := color.RGBA{R: 168, G: 58, B: 215, A: 255}
	highCrimson := color.RGBA{R: 225, G: 48, B: 135, A: 255}

	sampleAuroraOklab := func(u, curtainShift float64) color.RGBA {
		u = common.Clamp01(u)
		lowerEdge := render.LerpOklab(lowEmerald, lowCyan, common.Clamp01(curtainShift))
		middle := render.LerpOklab(midJade, midTeal, common.Clamp01(curtainShift))
		upperFringe := render.LerpOklab(highViolet, highCrimson, common.Clamp01(curtainShift))

		if u <= 0.30 {
			return render.LerpOklab(lowerEdge, middle, u/0.30)
		}
		if u <= 0.65 {
			return render.LerpOklab(middle, upperFringe, (u-0.30)/0.35)
		}
		return render.LerpOklab(upperFringe, highCrimson, (u-0.65)/0.35)
	}

	rc.SetBlendMode(render.BlendAdd)
	// Magnetic field inclination angle (slightly tilted toward the celestial pole)
	magTiltX := -0.11

	for cIdx := 0; cIdx < curtainNum; cIdx++ {
		cFrac := float64(cIdx) / float64(common.MaxInt(curtainNum-1, 1))
		baseY := fh * (0.28 + 0.24*cFrac)
		curtainLen := fh * (0.22 + 0.06*(1.0-cFrac))
		phase := float64(cIdx)*1.75 + rng.Float64()*0.8
		curtainShift := common.Clamp01(0.15 + 0.75*cFrac)

		// Soft volumetric background glow along the curtain path
		for g := 0; g <= 28; g++ {
			gt := float64(g) / 28.0
			foldX := math.Sin(gt*math.Pi*2.4+phase) * 42.0 * turbulence * scale
			gx := common.Lerp(-0.04*fw, 1.04*fw, gt) + foldX
			gy := baseY + math.Sin(gt*math.Pi*3.2+phase)*28.0*turbulence*scale
			rc.DrawGlowDot(gx, gy-curtainLen*0.32, 34.0*scale, sampleAuroraOklab(0.22, curtainShift), 0.042, 0.0)
		}

		// Cast rayNum vertical/slightly-tilted magnetic field rays upward along the folding 3D S-curve baseline
		for rIdx := 0; rIdx < rayNum; rIdx++ {
			t := float64(rIdx) / float64(common.MaxInt(rayNum-1, 1))

			// 3D folding ribbon parameterization: x(t) folds back on itself ligeramente with FBM/Curl noise
			curl := noise.Curl2D(t*2.8+float64(cIdx)*3.1, float64(cIdx)*1.9+0.5, 0.05)
			fbmX := noise.FBM2D(t*3.5+float64(cIdx)*5.3, 1.7, 4, 2.0, 0.5)
			fbmY := noise.FBM2D(t*4.2+float64(cIdx)*2.9, 8.3, 4, 2.0, 0.5)

			sCurveX := math.Sin(t*math.Pi*2.6+phase)*52.0*turbulence*scale + fbmX*36.0*turbulence*scale + curl.X*18.0*turbulence*scale
			sCurveY := math.Sin(t*math.Pi*3.4+phase*0.8)*26.0*turbulence*scale +
				math.Cos(t*math.Pi*6.2+phase)*12.0*turbulence*scale +
				fbmY*22.0*turbulence*scale

			bx := common.Lerp(-0.05*fw, 1.05*fw, t) + sCurveX
			by := common.Constrain(baseY+sCurveY, fh*0.14, yHorizon-12.0*scale)

			// Modulate ray height and intensity with fine magnetic striation noise
			striation := 0.5 + 0.5*noise.Simplex2D(t*24.0+float64(cIdx)*9.1, float64(cIdx)*4.3)
			rayHeight := curtainLen * (0.62 + 0.58*striation)
			rayAlpha := (0.042 + 0.065*math.Pow(striation, 1.35)) * math.Sin(common.Constrain(t, 0.02, 0.98)*math.Pi)

			// Intense lower-edge ionization glow bead at u = 0
			edgeCol := sampleAuroraOklab(0.05, curtainShift)
			rc.DrawGlowDot(bx, by, (2.2+2.0*striation)*scale, edgeCol, rayAlpha*0.55, 0.25)

			// Sample vertical sub-pixel sand/glow streaks along altitude u in [0, 1]
			verticalSamples := common.MaxInt(48, int(rayHeight/(1.4*scale)))
			for s := 0; s < verticalSamples; s++ {
				u := (float64(s) + rng.Float64()) / float64(verticalSamples)
				// Altitude intensity profile: sharp bright lower edge (u ~ 0.08), exponential tail toward u = 1
				altProfile := math.Pow(common.Clamp01(u/0.10), 0.75) * math.Pow(1.0-u, 1.35)
				if altProfile <= 1e-4 {
					continue
				}

				px := bx + u*rayHeight*magTiltX + rng.NormFloat64()*0.95*scale
				py := by - u*rayHeight + rng.NormFloat64()*0.75*scale
				if py < 0 || py >= yHorizon {
					continue
				}

				col := sampleAuroraOklab(u, curtainShift)
				rc.SplatPoint(px, py, col, rayAlpha*altProfile)
				if s%6 == 0 && u < 0.42 {
					rc.SplatPoint(px+rng.NormFloat64()*1.6*scale, py, col, rayAlpha*altProfile*0.55)
				}
			}
		}
	}

	// Snapshot the sky + starfield + aurora linear RGB buffer BEFORE drawing foreground mountains
	// so the lake reflection can mirror both the unobstructed aurora sky and the mountain silhouette!
	skyAuroraBuf := make([]float64, w*(horizonRow+1)*3)
	for y := 0; y <= horizonRow; y++ {
		for x := 0; x < w; x++ {
			rL, gL, bL, _ := rc.PixelLinearAt(x, y)
			idx := (y*w + x) * 3
			skyAuroraBuf[idx+0] = rL
			skyAuroraBuf[idx+1] = gL
			skyAuroraBuf[idx+2] = bL
		}
	}

	// 4. At y_horizon = 0.68 * h, draw 2 layers of dark jagged mountain silhouettes (Ridged2D noise)
	// with subtle snow-crest moonlight/aurora rim.
	rc.SetBlendMode(render.BlendOver)
	for layer := 0; layer < 2; layer++ {
		isForeground := layer == 1
		peakAmp := fh * 0.19
		freq := 2.4 / fw
		seedOff := 14.7
		baseCol := color.RGBA{R: 11, G: 20, B: 32, A: 255}
		rimCol := color.RGBA{R: 95, G: 215, B: 190, A: 255}

		if isForeground {
			peakAmp = fh * 0.135
			freq = 3.2 / fw
			seedOff = 43.2
			baseCol = color.RGBA{R: 6, G: 11, B: 20, A: 255}
			rimCol = color.RGBA{R: 130, G: 235, B: 205, A: 255}
		}

		ridgeY := make([]float64, w)
		crestPts := make([]render.Point, 0, w/4+2)
		for x := 0; x < w; x++ {
			fx := float64(x)
			rNoise := noise.Ridged2D(fx*freq+seedOff, seedOff*0.5, 5, 2.1, 0.52)
			detail := noise.FBM2D(fx*freq*2.5+seedOff, 9.1, 3, 2.0, 0.5)
			// Valley dip near center-right or center-left so aurora reflections shine prominently on the lake
			valleyEnv := 0.52 + 0.48*math.Cos((fx/fw-0.18*float64(layer))*2.0*math.Pi)
			if !isForeground {
				valleyEnv = 0.62 + 0.38*math.Sin((fx/fw)*math.Pi*1.3+0.4)
			}
			hOffset := common.Clamp01(0.76*rNoise+0.24*(0.5*(detail+1.0))) * peakAmp * (0.45 + 0.55*valleyEnv)
			ry := common.Constrain(yHorizon-hOffset, fh*0.32, yHorizon)
			ridgeY[x] = ry
			if x%4 == 0 || x == w-1 {
				crestPts = append(crestPts, render.Pt(fx, ry))
			}
		}

		// Fill mountain silhouette down to y_horizon with subtle alpine snow-slope shading
		for x := 0; x < w; x++ {
			topY := int(math.Floor(ridgeY[x]))
			slope := 0.0
			if x > 0 && x < w-1 {
				slope = (ridgeY[x+1] - ridgeY[x-1]) * 0.5
			}
			for y := topY; y <= horizonRow; y++ {
				distFromCrest := float64(y) - ridgeY[x]
				if distFromCrest < -1.0 {
					continue
				}
				cov := common.Clamp01((distFromCrest + 1.0) / 1.5)
				snowFactor := math.Exp(-math.Max(0, distFromCrest)/(14.0*scale)) * common.Clamp01(0.45-slope*0.25)
				pxCol := render.LerpOklab(baseCol, rimCol, snowFactor*0.28)
				rL, gL, bL, _ := render.ColorToLinearRGB(pxCol)
				rc.BlendPixelLinear(x, y, rL, gL, bL, cov, render.BlendOver)
			}
		}

		// Subtle snow-crest moonlight/aurora rim highlight along the ridgeline
		rc.SetBlendMode(render.BlendAdd)
		rc.StrokePolyline(crestPts, (1.15-0.25*float64(layer))*scale, rimCol, 0.34-0.08*float64(layer))
		rc.SetBlendMode(render.BlendOver)
	}

	// 5. Below y_horizon (y in [0.68*h, h]), render the calm alpine lake reflection:
	// mirror the sky + aurora curtains + mountains vertically with horizontal sine-wave water ripple
	// displacement and Fresnel darkening, plus fine horizontal specular water glints.
	waterTintR, waterTintG, waterTintB, _ := render.ColorToLinearRGB(color.RGBA{R: 5, G: 16, B: 26, A: 255})
	lakeHeight := math.Max(1.0, float64(h-horizonRow))

	for y := horizonRow + 1; y < h; y++ {
		// Normalized distance from shoreline (0 at horizon, 1 at bottom of canvas)
		v := float64(y-horizonRow) / lakeHeight

		// Perspective ripple frequency (higher near horizon, broader near foreground)
		rippleFreq := (0.55 / math.Max(0.08, v)) / scale
		ripplePhase := math.Sin(float64(y)*rippleFreq) + 0.5*math.Cos(float64(y)*rippleFreq*1.7+1.3)
		rippleAmp := (1.5 + 11.5*math.Pow(v, 0.85)) * scale

		// Fresnel reflection attenuation: stronger reflection near grazing angles, darker looking straight down
		fresnel := 0.72 - 0.32*math.Pow(v, 0.75)

		for x := 0; x < w; x++ {
			fx := float64(x)
			nWave := noise.Simplex2D(fx*0.025/scale, float64(y)*0.12/scale)
			dx := (ripplePhase*0.65 + nWave*0.35) * rippleAmp
			dy := nWave * 3.5 * v * scale

			sx := common.ConstrainInt(int(math.Round(fx+dx)), 0, w-1)
			syMirror := common.ConstrainInt(int(math.Round(float64(horizonRow)-(float64(y-horizonRow)*0.95+dy))), 0, horizonRow)
			sySky := common.ConstrainInt(int(math.Round(float64(horizonRow)-(float64(y-horizonRow)*1.18+dy))), 0, horizonRow)

			// Blend reflected mountain/sky scene with unobstructed aurora sky column so aurora light shimmers across the water
			mR, mG, mB, _ := rc.PixelLinearAt(sx, syMirror)
			sIdx := (sySky*w + sx) * 3
			aR := skyAuroraBuf[sIdx+0]
			aG := skyAuroraBuf[sIdx+1]
			aB := skyAuroraBuf[sIdx+2]

			mixAurora := 0.55 * common.Smoothstep(0.02, 0.28, v)
			refR := (mR*(1.0-mixAurora) + aR*mixAurora) * fresnel
			refG := (mG*(1.0-mixAurora) + aG*mixAurora) * fresnel
			refB := (mB*(1.0-mixAurora) + aB*mixAurora) * fresnel

			outR := refR*0.86 + waterTintR*0.14
			outG := refG*0.88 + waterTintG*0.12
			outB := refB*0.90 + waterTintB*0.10

			rc.SetPixelLinear(x, y, outR, outG, outB, 1.0)
		}
	}

	// Crisp shoreline mist & horizontal specular water glints in BlendAdd mode
	rc.SetBlendMode(render.BlendAdd)
	shoreMistCol := color.RGBA{R: 65, G: 210, B: 185, A: 255}
	for x := 0; x < w; x += 3 {
		rc.DrawGlowDot(float64(x), yHorizon, 2.2*scale, shoreMistCol, 0.08, 0.2)
	}

	glintCount := 420
	for i := 0; i < glintCount; i++ {
		gx := rng.Float64() * fw
		v := math.Pow(rng.Float64(), 0.9)
		gy := yHorizon + 2.0 + v*(lakeHeight-3.0)

		// Sample reflected brightness at (gx, gy) so glints concentrate where aurora reflections are brightest
		ix := common.ConstrainInt(int(gx), 0, w-1)
		iy := common.ConstrainInt(int(gy), horizonRow+1, h-1)
		rL, gL, bL, _ := rc.PixelLinearAt(ix, iy)
		lum := 0.2126*rL + 0.7152*gL + 0.0722*bL
		if lum < 0.015 && rng.Float64() < 0.7 {
			continue
		}

		glintHalfW := (2.0 + 7.5*v*rng.Float64()) * scale
		glintCol := render.OklabToRGBA(render.LinearRGBAToOklab(
			common.Constrain(rL*2.4+0.08, 0, 1),
			common.Constrain(gL*2.4+0.12, 0, 1),
			common.Constrain(bL*2.4+0.10, 0, 1),
		), 255)

		steps := common.MaxInt(5, int(glintHalfW*2.2))
		for s := 0; s <= steps; s++ {
			u := (float64(s)/float64(steps))*2.0 - 1.0
			env := 1.0 - u*u
			rc.SplatPoint(gx+u*glintHalfW, gy, glintCol, (0.14+0.32*math.Min(1.0, lum*3.0))*env)
		}
	}

	// 6. Apply HDR bloom, vignette, and ACES filmic tone mapping
	rc.ApplyBloom(0.42, 10.0, 0.65)
	rc.ApplyVignette(0.28)
	c.MergeRenderCanvasToneMapped(rc, 1.12, 0)
}
