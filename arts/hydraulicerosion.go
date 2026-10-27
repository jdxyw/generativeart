package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type hydraulicErosion struct {
	dropletsN       int
	erosionStrength float64
}

// NewHydraulicErosion returns a hydraulicErosion engine that simulates particle-based
// physical hydraulic erosion (35,000 rain droplets carving dendritic gullies, alluvial
// fans, and braided glacial river valleys) and renders the eroded heightfield with
// a 2x2 SSAA floating-horizon 3D renderer.
func NewHydraulicErosion(dropletsN int, erosionStrength float64) *hydraulicErosion {
	if dropletsN <= 0 {
		dropletsN = 32000
	}
	if erosionStrength <= 0 {
		erosionStrength = 1.0
	}
	return &hydraulicErosion{
		dropletsN:       dropletsN,
		erosionStrength: erosionStrength,
	}
}

// Generative runs the hydraulic droplet erosion simulation and renders the 3D alpine terrain.
func (he *hydraulicErosion) Generative(c *generativeart.Canva) {
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
	rng := rand.New(rand.NewSource(seed + 101))
	noise := common.NewNoiseEngine(seed + 202)

	// 1. Construct Initial Alpine Mountain Range on a 256 x 256 Simulation Grid
	const gridN = 256
	hMap := make([]float64, gridN*gridN)
	waterFlux := make([]float64, gridN*gridN)
	deposition := make([]float64, gridN*gridN)

	for gy := 0; gy < gridN; gy++ {
		v := (float64(gy)/float64(gridN-1))*2.0 - 1.0
		for gx := 0; gx < gridN; gx++ {
			u := (float64(gx)/float64(gridN-1))*2.0 - 1.0

			// Central diagonal alpine massif + flanking peaks + central glacial valley trough
			dPeak1 := math.Hypot(u+0.34, v+0.18)
			dPeak2 := math.Hypot(u-0.38, v-0.12)
			dPeak3 := math.Hypot(u+0.04, v-0.46)
			dPeak4 := math.Hypot(u-0.42, v+0.42)
			peaks := 0.78*math.Exp(-dPeak1*dPeak1*3.2) +
				0.74*math.Exp(-dPeak2*dPeak2*3.4) +
				0.64*math.Exp(-dPeak3*dPeak3*3.8) +
				0.52*math.Exp(-dPeak4*dPeak4*4.0)

			ridged := noise.Ridged2D(u*2.4+1.7, v*2.4-1.3, 5, 2.1, 0.52)
			fbm := 0.5 + 0.5*noise.FBM2D(u*3.6-2.1, v*3.6+1.4, 4, 2.0, 0.5)

			// Broad S-curved valley channel winding from top-center to bottom-center
			valleyCenter := 0.22 * math.Sin(v*2.4+0.4)
			valleyCut := 1.0 - 0.56*math.Exp(-math.Pow((u-valleyCenter)/0.25, 2.0))

			elev := (peaks*0.62 + ridged*0.46*(0.35+0.65*peaks) + fbm*0.12) * valleyCut
			hMap[gy*gridN+gx] = common.Constrain(elev, 0.02, 1.35)
		}
	}

	// 2. Particle-Based Hydraulic Erosion Simulation (Lague / Beyer Bilinear Droplet Model)
	nDrops := he.dropletsN
	if nDrops < 8000 {
		nDrops = 32000
	}
	eStr := common.Constrain(he.erosionStrength, 0.4, 2.0)

	const (
		maxDropSteps = 54
		inertia      = 0.14
		sedCapacity  = 4.2
		minSlope     = 0.008
		erodeSpeed   = 0.34
		depositSpeed = 0.32
		evapSpeed    = 0.022
		gravity      = 4.0
	)

	for d := 0; d < nDrops; d++ {
		px := 4.0 + rng.Float64()*float64(gridN-9)
		py := 4.0 + rng.Float64()*float64(gridN-9)
		dirX, dirY := 0.0, 0.0
		speed := 1.0
		water := 1.0
		sediment := 0.0

		for step := 0; step < maxDropSteps; step++ {
			ix := int(px)
			iy := int(py)
			if ix < 2 || ix >= gridN-3 || iy < 2 || iy >= gridN-3 {
				break
			}
			fx := px - float64(ix)
			fy := py - float64(iy)
			idx := iy*gridN + ix

			// Bilinear height & gradient
			h00 := hMap[idx]
			h10 := hMap[idx+1]
			h01 := hMap[idx+gridN]
			h11 := hMap[idx+gridN+1]

			gx := (h10-h00)*(1.0-fy) + (h11-h01)*fy
			gy := (h01-h00)*(1.0-fx) + (h11-h10)*fx
			hOld := h00*(1.0-fx)*(1.0-fy) + h10*fx*(1.0-fy) + h01*(1.0-fx)*fy + h11*fx*fy

			dirX = dirX*inertia - gx*(1.0-inertia)
			dirY = dirY*inertia - gy*(1.0-inertia)
			dLen := math.Hypot(dirX, dirY)
			if dLen < 1e-6 {
				a := rng.Float64() * 2.0 * math.Pi
				dirX, dirY = math.Cos(a), math.Sin(a)
			} else {
				dirX /= dLen
				dirY /= dLen
			}

			px += dirX
			py += dirY
			ixNew, iyNew := int(px), int(py)
			if ixNew < 2 || ixNew >= gridN-3 || iyNew < 2 || iyNew >= gridN-3 {
				break
			}
			fxNew, fyNew := px-float64(ixNew), py-float64(iyNew)
			idxNew := iyNew*gridN + ixNew
			hNew := hMap[idxNew]*(1.0-fxNew)*(1.0-fyNew) +
				hMap[idxNew+1]*fxNew*(1.0-fyNew) +
				hMap[idxNew+gridN]*(1.0-fxNew)*fyNew +
				hMap[idxNew+gridN+1]*fxNew*fyNew

			dH := hNew - hOld
			cap := math.Max(-dH, minSlope) * speed * water * sedCapacity * eStr

			w00 := (1.0 - fx) * (1.0 - fy)
			w10 := fx * (1.0 - fy)
			w01 := (1.0 - fx) * fy
			w11 := fx * fy

			if sediment > cap || dH > 0 {
				amt := (sediment - cap) * depositSpeed
				if dH > 0 {
					amt = math.Min(dH, sediment)
				}
				sediment -= amt
				hMap[idx] += amt * w00
				hMap[idx+1] += amt * w10
				hMap[idx+gridN] += amt * w01
				hMap[idx+gridN+1] += amt * w11
				deposition[idx] += amt
			} else {
				amt := math.Min((cap-sediment)*erodeSpeed*eStr, -dH)
				sediment += amt
				hMap[idx] -= amt * w00
				hMap[idx+1] -= amt * w10
				hMap[idx+gridN] -= amt * w01
				hMap[idx+gridN+1] -= amt * w11
			}

			// Record stream water flux for glacial river rendering
			if step > 4 {
				waterFlux[idx] += water * w00
				waterFlux[idx+1] += water * w10
				waterFlux[idx+gridN] += water * w01
				waterFlux[idx+gridN+1] += water * w11
			}

			speed = math.Sqrt(math.Max(0.05, speed*speed-dH*gravity))
			water *= (1.0 - evapSpeed)
		}
	}

	// Bilinear sampler for (u, v) in [-1, 1]
	sampleGrid := func(u, v float64) (elev, flux, dep float64) {
		gx := (u*0.5 + 0.5) * float64(gridN-3)
		gy := (v*0.5 + 0.5) * float64(gridN-3)
		ix := common.ConstrainInt(int(math.Floor(gx)), 1, gridN-3)
		iy := common.ConstrainInt(int(math.Floor(gy)), 1, gridN-3)
		fx := common.Clamp01(gx - float64(ix))
		fy := common.Clamp01(gy - float64(iy))
		idx := iy*gridN + ix

		w00 := (1.0 - fx) * (1.0 - fy)
		w10 := fx * (1.0 - fy)
		w01 := (1.0 - fx) * fy
		w11 := fx * fy

		elev = hMap[idx]*w00 + hMap[idx+1]*w10 + hMap[idx+gridN]*w01 + hMap[idx+gridN+1]*w11
		flux = waterFlux[idx]*w00 + waterFlux[idx+1]*w10 + waterFlux[idx+gridN]*w01 + waterFlux[idx+gridN+1]*w11
		dep = deposition[idx]*w00 + deposition[idx+1]*w10 + deposition[idx+gridN]*w01 + deposition[idx+gridN+1]*w11
		return
	}

	// Alpine Glacier, Eroded Granite, Terracotta Scree, Pine Alluvial Fan & Turquoise River Palette
	skyAbyss := color.RGBA{R: 10, G: 16, B: 34, A: 255}
	skyHorizon := color.RGBA{R: 32, G: 68, B: 112, A: 255}
	snowCrest := color.RGBA{R: 252, G: 248, B: 238, A: 255}
	glacierIce := color.RGBA{R: 158, G: 224, B: 238, A: 255}
	graniteSlate := color.RGBA{R: 76, G: 94, B: 122, A: 255}
	ochreScree := color.RGBA{R: 218, G: 138, B: 68, A: 255}
	terracottaGully := color.RGBA{R: 178, G: 68, B: 48, A: 255}
	pineAlluvial := color.RGBA{R: 38, G: 138, B: 112, A: 255}
	riverTurquoise := color.RGBA{R: 64, G: 228, B: 226, A: 255}

	ssBufR := make([]float64, sw*sh)
	ssBufG := make([]float64, sw*sh)
	ssBufB := make([]float64, sw*sh)

	for py := 0; py < sh; py++ {
		tSky := float64(py) / fsh
		cSky := render.LerpOklab(skyAbyss, skyHorizon, math.Pow(common.Clamp01(tSky*1.8), 1.2))
		rL, gL, bL, _ := render.ColorToLinearRGB(cSky)
		rowOff := py * sw
		for px := 0; px < sw; px++ {
			ssBufR[rowOff+px], ssBufG[rowOff+px], ssBufB[rowOff+px] = rL, gL, bL
		}
	}

	yHorizon := make([]int, sw)
	for px := 0; px < sw; px++ {
		yHorizon[px] = sh
	}

	// 3. 2x2 SSAA Floating-Horizon 3D Rendering of the Eroded Heightfield
	const zSteps = 920
	sunX, sunZ, sunY := -0.66, -0.34, 0.67
	sLen := math.Sqrt(sunX*sunX + sunZ*sunZ + sunY*sunY)
	sunX, sunZ, sunY = sunX/sLen, sunZ/sLen, sunY/sLen
	eps := 0.008

	for iz := zSteps - 1; iz >= 0; iz-- {
		tz := float64(iz) / float64(zSteps-1)
		v := (tz*2.0 - 1.0) * 0.96
		persp := 0.88 + 0.30*tz
		invPersp := 1.0 / (0.64 * persp)

		syBase := 0.22*fsh + tz*fsh*1.04
		zScale := 395.0 * scale * (0.76 + 0.32*tz)

		for px := 0; px < sw; px++ {
			if yHorizon[px] <= 0 {
				continue
			}
			tx := ((float64(px)+0.5)/fsw - 0.50) * invPersp
			u := common.Constrain(tx*1.04, -0.98, 0.98)

			elev, flux, dep := sampleGrid(u, v)
			syTop := syBase - elev*zScale
			yTop := int(math.Floor(syTop))
			if yTop >= yHorizon[px] {
				continue
			}

			hR, _, _ := sampleGrid(u+eps, v)
			hL, _, _ := sampleGrid(u-eps, v)
			hF, _, _ := sampleGrid(u, v+eps)
			hB, _, _ := sampleGrid(u, v-eps)

			dhdu := (hR - hL) / (2.0 * eps)
			dhdv := (hF - hB) / (2.0 * eps)
			slope := math.Hypot(dhdu, dhdv)

			nx, nz, ny := -dhdu*0.58, -dhdv*0.58, 1.0
			nLen := math.Sqrt(nx*nx + nz*nz + ny*ny)
			nx, nz, ny = nx/nLen, nz/nLen, ny/nLen
			ndotl := common.Constrain(nx*(-sunX)+nz*(-sunZ)+ny*sunY, 0, 1)

			// Cast sun shadow check
			shadow := 1.0
			for s := 1; s <= 5; s++ {
				dist := float64(s) * 0.032
				hOcc, _, _ := sampleGrid(u-dist*1.1, v-dist*0.6)
				if hOcc > elev+dist*0.42 {
					shadow = math.Min(shadow, math.Exp(-(hOcc-(elev+dist*0.42))*9.5))
				}
			}

			// Geomorphological material classification from elevation, slope, deposition & water flux
			var matCol color.RGBA
			switch {
			case elev > 0.48 && slope < 1.55:
				// High alpine snow & glacier ice
				snowT := common.Clamp01((elev - 0.48) / 0.26)
				matCol = render.LerpOklab(glacierIce, snowCrest, snowT)
			case slope > 0.72:
				// Steep eroded granite & terracotta gully walls
				gullyT := common.Clamp01((slope - 0.72) / 1.1)
				matCol = render.LerpOklab(ochreScree, graniteSlate, gullyT)
			default:
				// Mid-elevation alluvial fans & pine valleys
				alluvT := common.Clamp01(dep * 14.0)
				matCol = render.LerpOklab(pineAlluvial, ochreScree, common.Clamp01(0.28+0.62*alluvT+elev*0.35))
				if dep > 0.012 {
					matCol = render.LerpOklab(matCol, terracottaGully, common.Clamp01(dep*10.0)*0.42)
				}
			}

			// Braided glacial streams in high-flux channels (only inside carved valley bottoms)
			riverW := common.Clamp01((flux - 7.5) / 28.0)
			if riverW > 0.08 && elev < 0.36 {
				matCol = render.LerpOklab(matCol, riverTurquoise, riverW*0.85)
			}

			// Atmospheric depth perspective toward the back peaks
			farT := common.Smoothstep(0.35, 0.0, tz)
			if farT > 0 {
				matCol = render.LerpOklab(matCol, skyHorizon, farT*0.48)
			}

			light := (0.28 + 0.72*ndotl) * (0.38 + 0.62*shadow)
			litCol := render.LerpOklab(skyAbyss, matCol, common.Constrain(0.22+0.88*light, 0.08, 1.0))
			if ndotl > 0.68 && shadow > 0.7 {
				litCol = render.LerpOklab(litCol, snowCrest, (ndotl-0.68)*0.55)
			}
			cliffDark := render.LerpOklab(skyAbyss, graniteSlate, 0.35*light)

			litR, litG, litB, _ := render.ColorToLinearRGB(litCol)
			dpR, dpG, dpB, _ := render.ColorToLinearRGB(cliffDark)

			yStart := common.MaxInt(0, yTop)
			yEnd := yHorizon[px] - 1
			cliffSpan := math.Max(12.0*scale, syBase-syTop)

			for py := yStart; py <= yEnd; py++ {
				tV := common.Clamp01((float64(py) - syTop) / cliffSpan)
				gCurve := math.Pow(tV, 0.72)
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
