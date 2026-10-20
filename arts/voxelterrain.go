package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type voxelTerrain struct {
	gridRes       int
	swirlStrength float64
	heightScale   float64
}

// NewVoxelTerrain returns a voxelTerrain engine that renders a 2x2 Supersampled (SSAA)
// Floating-Horizon 3D swirling canyon with 10-12 sharp craggy spiral ridges, crisp
// vertical micro-columnar fluting, left-sun directional shadows, and zero Moire or blur.
func NewVoxelTerrain(gridRes int, swirlStrength, heightScale float64) *voxelTerrain {
	if gridRes <= 0 {
		gridRes = 340
	}
	if swirlStrength <= 0 {
		swirlStrength = 2.6
	}
	if heightScale <= 0 {
		heightScale = 1.0
	}
	return &voxelTerrain{
		gridRes:       gridRes,
		swirlStrength: swirlStrength,
		heightScale:   heightScale,
	}
}

// Generative renders the 3D Swirling Canyon Terrain onto c.
func (vt *voxelTerrain) Generative(c *generativeart.Canva) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	ss := 2
	sw := w * ss
	sh := h * ss
	fsw, fsh := float64(sw), float64(sh)
	scale := math.Min(fsw, fsh) / 1200.0

	swirl := vt.swirlStrength
	if swirl <= 0 {
		swirl = 2.6
	}
	hScale := vt.heightScale
	if hScale <= 0 {
		hScale = 1.0
	}

	seed := rand.Int63()
	noise := common.NewNoiseEngine(seed)

	// Reference Image 2 Palette
	crestCream := color.RGBA{R: 255, G: 246, B: 226, A: 255}
	crestPeach := color.RGBA{R: 252, G: 178, B: 144, A: 255}
	crestCoral := color.RGBA{R: 234, G: 122, B: 152, A: 255}
	mintLight := color.RGBA{R: 186, G: 244, B: 212, A: 255}
	emeraldMid := color.RGBA{R: 68, G: 194, B: 156, A: 255}
	tealDeep := color.RGBA{R: 44, G: 134, B: 148, A: 255}
	sageCool := color.RGBA{R: 152, G: 188, B: 178, A: 255}
	mauveFold := color.RGBA{R: 156, G: 88, B: 168, A: 255}
	violetTrough := color.RGBA{R: 58, G: 36, B: 108, A: 255}
	sapphireFar := color.RGBA{R: 20, G: 44, B: 112, A: 255}
	midnightBg := color.RGBA{R: 8, G: 12, B: 34, A: 255}

	hookU, hookV := 0.02, -0.04
	evalTerrain := func(u, v float64) (hMacro, hCraggy, coreW float64, col color.RGBA) {
		du := u - hookU
		dv := (v - hookV) * 1.08
		r2 := du*du + dv*dv
		r := math.Sqrt(r2)

		// Non-singular Cartesian vortex rotation: twists parallel ridges into a tight S-spiral hook
		twist := swirl * 1.58 / (1.0 + 0.90*r2)
		cosT, sinT := math.Cos(twist), math.Sin(twist)
		su := du*cosT - dv*sinT
		sv := du*sinT + dv*cosT

		warpU := noise.FBM2D(su*1.45+1.9, sv*1.45-1.2, 3, 2.0, 0.5) * 0.24
		warpV := noise.FBM2D(su*1.45-2.4, sv*1.45+2.1, 3, 2.0, 0.5) * 0.24
		wu := su + warpU
		wv := sv + warpV

		// 1. Central S-curved Hero Wave Ridge (sweeping from bottom-left up into the central curled hook!)
		spineDist := math.Abs(wu + 0.22*wv*wv)
		heroSpine := math.Exp(-spineDist*spineDist/(2.0*0.14*0.14)) * math.Exp(-r2/(2.0*0.58*0.58))

		// 2. 10-12 Sharp, Craggy Nested Spiral Wave Ridges
		waveCoord := wu*3.70 + 0.82*wv + 2.25*r
		phase := math.Mod(waveCoord+20.0, 1.0)
		var asymWave float64
		if phase < 0.65 {
			t := phase / 0.65
			asymWave = t * t * (3.0 - 2.0*t)
		} else {
			t := (1.0 - phase) / 0.35
			asymWave = t * t * (3.0 - 2.0*t)
		}
		sharpWave := math.Pow(asymWave, 1.72)

		ridged := noise.Ridged2D(wu*3.2+2.3, wv*3.2-1.7, 4, 2.05, 0.52)
		cragPeaks := math.Pow(noise.Ridged2D(wu*7.0-1.4, wv*7.0+2.1, 3, 2.0, 0.5), 1.6) * 0.16

		rawMacro := (0.52*heroSpine*(0.65+0.48*ridged) +
			0.48*sharpWave*(0.48+0.62*ridged)) * math.Exp(-0.08*r2)

		hMacro = math.Max(0.02, rawMacro*hScale)

		// Crisp stepped strata shelves + fine columnar fluting along the ridge crests
		const strataSteps = 26.0
		terraced := math.Floor(rawMacro*strataSteps)/strataSteps +
			0.62*(rawMacro-math.Floor(rawMacro*strataSteps)/strataSteps)
		fluteRib := (0.5 + 0.5*math.Cos((wu*52.0+wv*18.0)*math.Pi)) * 0.018 * (0.3 + 0.7*sharpWave)

		hCraggy = math.Max(0.02, (0.48*rawMacro+0.52*terraced+cragPeaks*(0.25+0.75*sharpWave+0.7*heroSpine)+fluteRib)*hScale)

		coreW = common.Clamp01(heroSpine * 1.35 * math.Exp(-r2/(2.0*0.36*0.36)) * 1.45)

		bandT := 0.5 + 0.5*math.Sin(waveCoord*math.Pi*0.65+0.5)
		if bandT < 0.36 {
			col = render.LerpOklab(tealDeep, emeraldMid, bandT/0.36)
		} else if bandT < 0.72 {
			col = render.LerpOklab(emeraldMid, mintLight, (bandT-0.36)/0.36)
		} else {
			col = render.LerpOklab(mintLight, sageCool, (bandT-0.72)/0.28)
		}

		col = render.LerpOklab(violetTrough, col, common.Clamp01(0.20+0.92*sharpWave))
		purpleArc := math.Exp(-math.Pow((v+0.26)/0.26, 2.0)) * math.Exp(-math.Pow((u+0.14)/0.48, 2.0))
		if purpleArc > 0.04 {
			col = render.LerpOklab(col, mauveFold, common.Clamp01(purpleArc*0.88))
		}

		if coreW > 0.02 {
			warm := render.LerpOklab(crestCoral, crestPeach, common.Clamp01(sharpWave*0.75+heroSpine*0.45))
			warm = render.LerpOklab(warm, crestCream, math.Pow(common.Clamp01(heroSpine*0.75+sharpWave*0.35), 1.7)*0.85)
			col = render.LerpOklab(col, warm, common.Clamp01(coreW*1.25))
		}

		farDepth := common.Smoothstep(-0.18, -1.05, v) * common.Smoothstep(0.32, 1.12, r)
		if farDepth > 0 {
			col = render.LerpOklab(col, sapphireFar, farDepth*0.90)
		}
		return
	}

	ssBufR := make([]float64, sw*sh)
	ssBufG := make([]float64, sw*sh)
	ssBufB := make([]float64, sw*sh)

	bgR, bgG, bgB, _ := render.ColorToLinearRGB(midnightBg)
	for i := range ssBufR {
		ssBufR[i], ssBufG[i], ssBufB[i] = bgR, bgG, bgB
	}

	yHorizon := make([]int, sw)
	for px := 0; px < sw; px++ {
		yHorizon[px] = sh
	}

	const zSteps = 1040
	sunX, sunZ, sunY := -0.68, -0.32, 0.66
	sLen := math.Sqrt(sunX*sunX + sunZ*sunZ + sunY*sunY)
	sunX, sunZ, sunY = sunX/sLen, sunZ/sLen, sunY/sLen

	epsU := 0.0028
	epsV := 0.0028

	for iz := zSteps - 1; iz >= 0; iz-- {
		tz := float64(iz) / float64(zSteps-1)
		v := (tz*2.0 - 1.0) * 1.16
		persp := 0.86 + 0.34*tz
		invPersp := 1.0 / (0.58 * persp)

		syBase := 0.04*fsh + tz*fsh*1.29
		zScale := 175.0 * scale * (0.78 + 0.30*tz)

		for px := 0; px < sw; px++ {
			if yHorizon[px] <= 0 {
				continue
			}
			colIdx := px / 2
			tx := ((float64(colIdx*2)+1.0)/fsw - 0.50) * invPersp
			u := tx * 1.24

			hMac, hCrag, coreW, baseCol := evalTerrain(u, v)
			syTop := syBase - (26.0*scale + hCrag*zScale)
			yTop := int(math.Floor(syTop))

			if yTop >= yHorizon[px] {
				continue
			}

			_, hR, _, _ := evalTerrain(u+epsU, v)
			_, hL, _, _ := evalTerrain(u-epsU, v)
			_, hF, _, _ := evalTerrain(u, v+epsV)
			_, hB, _, _ := evalTerrain(u, v-epsV)

			dhdu := (hR - hL) / (2.0 * epsU)
			dhdv := (hF - hB) / (2.0 * epsV)

			nx, nz, ny := -dhdu*0.34, -dhdv*0.34, 1.0
			nLen := math.Sqrt(nx*nx + nz*nz + ny*ny)
			nx, nz, ny = nx/nLen, nz/nLen, ny/nLen

			ndotl := common.Constrain(nx*(-sunX)+nz*(-sunZ)+ny*sunY, 0.0, 1.0)

			// Right-facing shadowed slope of central S-ridge shifts toward deep amethyst-purple!
			if dhdu < -0.15 {
				rightSlopeT := common.Clamp01((-dhdu - 0.15) * 0.65)
				shadowTint := render.LerpOklab(tealDeep, mauveFold, coreW*0.88)
				baseCol = render.LerpOklab(baseCol, shadowTint, rightSlopeT*0.65)
			}

			shadow := 1.0
			for s := 1; s <= 6; s++ {
				dist := float64(s) * 0.030
				hOcc, _, _, _ := evalTerrain(u-dist*1.2, v-dist*0.8)
				diff := hOcc - (hMac + dist*0.36)
				if diff > 0 {
					shadow = math.Min(shadow, math.Exp(-diff*10.0))
				}
			}

			hBroadL, _, _, _ := evalTerrain(u-0.05, v)
			hBroadR, _, _, _ := evalTerrain(u+0.05, v)
			hBroadB, _, _, _ := evalTerrain(u, v-0.05)
			hBroadF, _, _, _ := evalTerrain(u, v+0.05)
			broadAvg := 0.25 * (hBroadL + hBroadR + hBroadB + hBroadF)
			sss := common.Clamp01((hMac-broadAvg)*5.0) * (0.4 + 0.6*shadow)

			localAvg := 0.25 * (hR + hL + hF + hB)
			ao := common.Constrain(0.78+2.8*(hCrag-localAvg), 0.32, 1.18)

			if sss > 0.03 {
				sssTint := render.LerpOklab(mintLight, crestCream, coreW)
				baseCol = render.LerpOklab(baseCol, sssTint, common.Clamp01(sss*0.52))
			}

			lab := render.RGBAToOklab(baseCol)
			light := (0.32 + 0.68*ndotl) * (0.36 + 0.64*shadow) * ao

			litCol := render.OklabToRGBA(render.Oklab{
				L: common.Constrain(lab.L*(0.58+0.44*light)+0.06*shadow+0.05*sss, 0.05, 0.98),
				A: lab.A * 0.94,
				B: lab.B + 0.010*shadow,
			}, 255)
			deepCol := render.OklabToRGBA(render.Oklab{
				L: common.Constrain(lab.L*(0.20+0.16*light), 0.015, 0.42),
				A: lab.A * 0.78,
				B: lab.B - 0.02,
			}, 255)

			litR, litG, litB, _ := render.ColorToLinearRGB(litCol)
			dpR, dpG, dpB, _ := render.ColorToLinearRGB(deepCol)

			colBevel := 1.0
			if px%2 == 1 {
				colBevel = 0.76
			}

			yStart := common.MaxInt(0, yTop)
			yEnd := yHorizon[px] - 1
			cliffSpan := math.Max(14.0*scale, syBase-syTop)

			for py := yStart; py <= yEnd; py++ {
				tVert := common.Clamp01((float64(py) - syTop) / cliffSpan)
				gCurve := math.Pow(tVert, 0.75)
				rLin := (litR*(1.0-gCurve) + dpR*gCurve) * colBevel
				gLin := (litG*(1.0-gCurve) + dpG*gCurve) * colBevel
				bLin := (litB*(1.0-gCurve) + dpB*gCurve) * colBevel

				idx := py*sw + px
				ssBufR[idx] = rLin
				ssBufG[idx] = gLin
				ssBufB[idx] = bLin
			}

			yHorizon[px] = yStart
		}
	}

	rc := c.NewRenderCanvas()
	for y := 0; y < h; y++ {
		y0 := y * 2
		y1 := y0 + 1
		for x := 0; x < w; x++ {
			x0 := x * 2
			x1 := x0 + 1
			i00 := y0*sw + x0
			i10 := y0*sw + x1
			i01 := y1*sw + x0
			i11 := y1*sw + x1

			r := 0.25 * (ssBufR[i00] + ssBufR[i10] + ssBufR[i01] + ssBufR[i11])
			g := 0.25 * (ssBufG[i00] + ssBufG[i10] + ssBufG[i01] + ssBufG[i11])
			b := 0.25 * (ssBufB[i00] + ssBufB[i10] + ssBufB[i01] + ssBufB[i11])
			rc.SetPixelLinear(x, y, r, g, b, 1.0)
		}
	}

	c.MergeRenderCanvasToneMapped(rc, 1.06, 0)
}
