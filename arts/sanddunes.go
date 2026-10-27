package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type sandDunes struct {
	dunesN    int
	windShear float64
}

// NewSandDunes returns a sandDunes engine that renders a 2x2 SSAA 3D desert
// landscape of Aeolian barchan and star sand dunes (巴丹吉林新月形沙丘链与大漠风波纹)
// with knife-edge sinuous ridge crests, golden-hour windward ripples, and deep
// ultramarine leeward slip-face shadows.
func NewSandDunes(dunesN int, windShear float64) *sandDunes {
	if dunesN <= 0 {
		dunesN = 7
	}
	if windShear <= 0 {
		windShear = 1.0
	}
	return &sandDunes{
		dunesN:    dunesN,
		windShear: windShear,
	}
}

// Generative renders the 3D Aeolian barchan sand dunes with 2x2 SSAA.
func (sd *sandDunes) Generative(c *generativeart.Canva) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	ss := 2
	sw, sh := w*ss, h*ss
	fsw, fsh := float64(sw), float64(sh)
	scale := math.Min(fsw, fsh) / 1200.0

	seed := c.Seed()
	if seed == 0 {
		seed = 20261027
	}
	noise := common.NewNoiseEngine(seed + 404)

	// Golden-Hour Sahara / Badain Jaran Desert Chiaroscuro Palette
	nightSky := color.RGBA{R: 12, G: 14, B: 32, A: 255}
	sunsetViolet := color.RGBA{R: 78, G: 38, B: 96, A: 255}
	sunsetAmber := color.RGBA{R: 242, G: 142, B: 62, A: 255}

	crestGold := color.RGBA{R: 255, G: 234, B: 158, A: 255}
	windwardAmber := color.RGBA{R: 244, G: 172, B: 74, A: 255}
	terracottaSand := color.RGBA{R: 212, G: 98, B: 56, A: 255}
	roseDune := color.RGBA{R: 186, G: 76, B: 92, A: 255}

	slipShadowViolet := color.RGBA{R: 54, G: 34, B: 92, A: 255}
	slipShadowNavy := color.RGBA{R: 18, G: 24, B: 64, A: 255}

	numDunes := float64(sd.dunesN)
	if numDunes < 4 {
		numDunes = 7
	}
	shear := sd.windShear
	if shear <= 0 {
		shear = 1.0
	}

	// Evaluate Aeolian Sand Dune Heightfield:
	// True C^0 cusp knife-edge ridge at frac == 0.62 (concave windward slope + steep leeward slip face)
	evalDunes := func(u, v float64) (hMacro, hRipple, fracOut float64, isWindward bool) {
		warpU := noise.FBM2D(u*0.85+1.4, v*0.85-0.8, 3, 2.0, 0.5) * 0.24 * shear
		warpV := noise.FBM2D(u*0.85-2.1, v*0.85+1.7, 3, 2.0, 0.5) * 0.20 * shear

		// S-curved barchan dune crest coordinate running gracefully across the desert
		duneCoord := (u*1.32 + warpU + 0.36*math.Sin(v*1.9+warpV)) * (numDunes * 0.35)
		frac := duneCoord - math.Floor(duneCoord)

		var profile float64
		if frac < 0.62 {
			isWindward = true
			t := frac / 0.62
			fracOut = t
			// Concave upward windward slope meeting the crest at a sharp non-zero angle
			profile = 0.25*t + 0.75*math.Pow(t, 1.55)
		} else {
			isWindward = false
			t := (1.0 - frac) / 0.38
			fracOut = t
			// Steep leeward slip face (34-deg angle of repose)
			profile = 0.30*t + 0.70*math.Pow(t, 1.35)
		}

		// Secondary gentle transverse dune swell along v
		crossSwell := 0.24 * (0.5 + 0.5*math.Sin((v*1.75-0.42*u+warpV)*math.Pi))
		env := 0.70 + 0.30*math.Cos(u*0.55)*math.Cos(v*0.45)
		hMacro = (profile*0.78 + crossSwell) * env

		// Fine Aeolian wind ripples across the windward slope
		rippleAmp := 0.0
		if isWindward {
			rippleAmp = 0.015 * math.Sin((u*36.0+v*62.0+warpU*12.0)*math.Pi) * math.Sin(fracOut*math.Pi)
		}
		hRipple = hMacro + rippleAmp
		return
	}

	ssBufR := make([]float64, sw*sh)
	ssBufG := make([]float64, sw*sh)
	ssBufB := make([]float64, sw*sh)

	// Twilight desert sky gradient + glowing setting sun disc in the upper sky
	sunScreenX, sunScreenY := 0.28*fsw, 0.11*fsh
	for py := 0; py < sh; py++ {
		tSky := common.Clamp01(float64(py) / (0.32 * fsh))
		var rowCol color.RGBA
		if tSky < 0.55 {
			rowCol = render.LerpOklab(nightSky, sunsetViolet, tSky/0.55)
		} else {
			rowCol = render.LerpOklab(sunsetViolet, sunsetAmber, (tSky-0.55)/0.45)
		}
		rowOff := py * sw
		for px := 0; px < sw; px++ {
			col := rowCol
			dSun := math.Hypot(float64(px)-sunScreenX, float64(py)-sunScreenY) / (76.0 * scale)
			if dSun < 2.2 {
				glow := math.Exp(-dSun * dSun * 1.8)
				col = render.LerpOklab(col, crestGold, common.Clamp01(glow*0.92))
			}
			rL, gL, bL, _ := render.ColorToLinearRGB(col)
			ssBufR[rowOff+px], ssBufG[rowOff+px], ssBufB[rowOff+px] = rL, gL, bL
		}
	}

	yHorizon := make([]int, sw)
	for px := 0; px < sw; px++ {
		yHorizon[px] = sh
	}

	const zSteps = 980
	// Low-angle golden-hour sunlight from the left (-0.82, -0.22, 0.52) casting sharp shadows across slip faces
	sunX, sunZ, sunY := -0.82, -0.22, 0.52
	sLen := math.Sqrt(sunX*sunX + sunZ*sunZ + sunY*sunY)
	sunX, sunZ, sunY = sunX/sLen, sunZ/sLen, sunY/sLen
	eps := 0.0035

	for iz := zSteps - 1; iz >= 0; iz-- {
		tz := float64(iz) / float64(zSteps-1)
		v := (tz*2.0 - 1.0) * 1.08
		persp := 0.85 + 0.35*tz
		invPersp := 1.0 / (0.58 * persp)

		syBase := 0.24*fsh + tz*fsh*0.98
		zScale := 148.0 * scale * (0.76 + 0.32*tz)

		for px := 0; px < sw; px++ {
			if yHorizon[px] <= 0 {
				continue
			}
			tx := ((float64(px)+0.5)/fsw - 0.50) * invPersp
			u := tx * 1.18

			_, hRip, slopeFrac, isWindward := evalDunes(u, v)
			syTop := syBase - hRip*zScale
			yTop := int(math.Floor(syTop))
			if yTop >= yHorizon[px] {
				continue
			}

			_, hR, _, _ := evalDunes(u+eps, v)
			_, hL, _, _ := evalDunes(u-eps, v)
			_, hF, _, _ := evalDunes(u, v+eps)
			_, hB, _, _ := evalDunes(u, v-eps)

			dhdu := (hR - hL) / (2.0 * eps)
			dhdv := (hF - hB) / (2.0 * eps)

			nx, nz, ny := -dhdu*0.55, -dhdv*0.55, 1.0
			nLen := math.Sqrt(nx*nx + nz*nz + ny*ny)
			nx, nz, ny = nx/nLen, nz/nLen, ny/nLen
			ndotl := common.Constrain(nx*(-sunX)+nz*(-sunZ)+ny*sunY, 0, 1)

			var col color.RGBA
			if isWindward {
				// Rich elevation + ripple gradient from inter-dune rose-terracotta up to blazing Sahara gold at the crest
				tGrad := common.Clamp01(0.58*slopeFrac + 0.42*ndotl)
				if tGrad < 0.42 {
					col = render.LerpOklab(roseDune, terracottaSand, tGrad/0.42)
				} else if tGrad < 0.78 {
					col = render.LerpOklab(terracottaSand, windwardAmber, (tGrad-0.42)/0.36)
				} else {
					col = render.LerpOklab(windwardAmber, crestGold, (tGrad-0.78)/0.22)
				}
				// Highlight the exact knife-edge ridge crest
				if slopeFrac > 0.96 {
					col = render.LerpOklab(col, crestGold, (slopeFrac-0.96)/0.04*0.92)
				}
			} else {
				// Shadowed leeward slip face: deep indigo under the crest -> royal violet -> warm dusk rose in the trough
				tShadow := common.Clamp01(1.0 - slopeFrac*0.82 + 0.18*ndotl)
				if tShadow < 0.52 {
					col = render.LerpOklab(slipShadowNavy, slipShadowViolet, tShadow/0.52)
				} else {
					col = render.LerpOklab(slipShadowViolet, roseDune, (tShadow-0.52)/0.48)
				}
			}

			// Atmospheric sunset haze on distant dunes
			farT := common.Smoothstep(0.32, 0.0, tz)
			if farT > 0 {
				col = render.LerpOklab(col, sunsetViolet, farT*0.42)
			}

			litR, litG, litB, _ := render.ColorToLinearRGB(col)
			dpR, dpG, dpB, _ := render.ColorToLinearRGB(render.LerpOklab(col, slipShadowNavy, 0.35))

			yStart := common.MaxInt(0, yTop)
			yEnd := yHorizon[px] - 1
			cliffSpan := math.Max(18.0*scale, syBase-syTop)

			for py := yStart; py <= yEnd; py++ {
				tV := common.Clamp01((float64(py) - syTop) / cliffSpan)
				gCurve := math.Pow(tV, 0.85)
				idx := py*sw + px
				ssBufR[idx] = litR*(1.0-gCurve) + dpR*gCurve
				ssBufG[idx] = litG*(1.0-gCurve) + dpG*gCurve
				ssBufB[idx] = litB*(1.0-gCurve) + dpB*gCurve
			}
			yHorizon[px] = yStart
		}
	}

	rc := c.NewRenderCanvas()
	for y := 0; y < h; y++ {
		y0, y1 := y*2, y*2+1
		for x := 0; x < w; x++ {
			x0, x1 := x*2, x*2+1
			i00, i10 := y0*sw+x0, y0*sw+x1
			i01, i11 := y1*sw+x0, y1*sw+x1
			rc.SetPixelLinear(x, y,
				0.25*(ssBufR[i00]+ssBufR[i10]+ssBufR[i01]+ssBufR[i11]),
				0.25*(ssBufG[i00]+ssBufG[i10]+ssBufG[i01]+ssBufG[i11]),
				0.25*(ssBufB[i00]+ssBufB[i10]+ssBufB[i01]+ssBufB[i11]),
				1.0)
		}
	}
	c.MergeRenderCanvasToneMapped(rc, 1.06, 0)
}
