package arts

import (
	"image/color"
	"math"
	"math/rand"
	"sort"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type iridescentNacre struct {
	whorls          int
	fiberCount      int
	filmThicknessNm float64
}

// NewIridescentNacre returns an IridescentNacre engine that sculpts a 3D terraced abalone /
// mother-of-pearl logarithmic-spiral shell heightfield and aragonite silk-bristle plumes
// shaded with physical Airy thin-film interference, horizon-mapped soft shadows, and jade SSS.
func NewIridescentNacre(whorls, fiberCount int, filmThicknessNm float64) *iridescentNacre {
	if whorls <= 0 {
		whorls = 7
	}
	if fiberCount <= 0 {
		fiberCount = 18000
	}
	if filmThicknessNm <= 0 {
		filmThicknessNm = 410.0
	}
	return &iridescentNacre{
		whorls:          whorls,
		fiberCount:      fiberCount,
		filmThicknessNm: filmThicknessNm,
	}
}

type nacreWhorlCenter struct {
	x, y     float64
	radius   float64
	spin     float64
	strength float64
	arms     float64
	phaseOff float64
}

type nacreFiberSeed struct {
	x, y       float64
	elev       float64
	thickness  float64
	cosView    float64
	whorlDist  float64
	strokeSeed float64
	lenMod     float64
	radMod     float64
	bendMod    float64
}

// Generative renders the static IridescentNacre composition.
func (in *iridescentNacre) Generative(c *generativeart.Canva) {
	in.GenerativeFrame(c, 0, 1, 0)
}

// GenerativeFrame renders frame frameIdx of a seamless looping IridescentNacre animation.
func (in *iridescentNacre) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	minDim := math.Min(fw, fh)
	scale := minDim / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	seed := c.Seed()
	if seed == 0 {
		seed = 20261802
	}
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	isAnim := totalFrames > 1
	phi := 2.0 * math.Pi * phase
	loopX := 0.08 * math.Sin(phi)
	loopY := 0.08 * (1.0 - math.Cos(phi))

	// 1. Seed multi-center nacreous whorl centers via SamplePoissonDisk2D + RelaxCentroidalVoronoi
	densityFn := func(x, y float64) float64 {
		nx := (x/fw - 0.52) * 2.0
		ny := (y/fh - 0.48) * 2.0
		r := math.Hypot(nx, ny)
		spiral := 0.5 + 0.5*math.Cos(math.Atan2(ny, nx)*2.0-3.4*math.Log(r+0.18))
		return common.Constrain(0.35+0.45*(1.0-0.55*r)+0.20*spiral, 0.05, 1.0)
	}

	approxSpacing := minDim / math.Max(2.4, math.Sqrt(float64(in.whorls))*1.22)
	rawPts := render.SamplePoissonDisk2D(fw*0.76, fh*0.76, approxSpacing, densityFn, rng)
	for i := range rawPts {
		rawPts[i].X += fw * 0.12
		rawPts[i].Y += fh * 0.12
	}
	relaxedPts := render.RelaxCentroidalVoronoi(rawPts, fw, fh, 96, densityFn, 6)

	// Sort candidate sites by distance to the golden-ratio focal center so the primary chamber leads
	focalX, focalY := 0.53*fw, 0.48*fh
	sort.Slice(relaxedPts, func(i, j int) bool {
		di := math.Hypot(relaxedPts[i].X-focalX, relaxedPts[i].Y-focalY)
		dj := math.Hypot(relaxedPts[j].X-focalX, relaxedPts[j].Y-focalY)
		return di < dj
	})

	centers := make([]nacreWhorlCenter, 0, in.whorls)
	goldenAngle := math.Pi * (3.0 - math.Sqrt(5.0))
	ptCursor := 0
	for i := 0; i < in.whorls; i++ {
		var cx, cy float64
		if i == 0 {
			cx, cy = focalX, focalY
		} else {
			found := false
			for ptCursor < len(relaxedPts) {
				p := relaxedPts[ptCursor]
				ptCursor++
				candX := common.Constrain(p.X, 0.15*fw, 0.85*fw)
				candY := common.Constrain(p.Y, 0.15*fh, 0.85*fh)
				tooClose := false
				for _, prev := range centers {
					if math.Hypot(candX-prev.x, candY-prev.y) < 0.22*minDim {
						tooClose = true
						break
					}
				}
				if !tooClose {
					cx, cy = candX, candY
					found = true
					break
				}
			}
			if !found {
				ang := float64(i) * goldenAngle
				rad := (0.22 + 0.16*math.Sqrt(float64(i)/float64(in.whorls))) * minDim
				cx = common.Constrain(focalX+math.Cos(ang)*rad, 0.15*fw, 0.85*fw)
				cy = common.Constrain(focalY+math.Sin(ang)*rad, 0.15*fh, 0.85*fh)
			}
		}

		spin := 1.0
		if i%2 == 1 {
			spin = -1.0
		}
		rad := (155.0 - 11.0*math.Min(float64(i), 7.0)) * scale
		if i == 0 {
			rad = 225.0 * scale
		}
		arms := 3.0 + float64(i%3)
		strength := 1.0 / (1.0 + 0.18*float64(i))
		if i == 0 {
			strength = 1.45
		}

		if isAnim {
			orbitAng := phi + float64(i)*0.85
			cx += 3.5 * scale * math.Sin(orbitAng)
			cy += 3.5 * scale * (1.0 - math.Cos(orbitAng))
		}

		centers = append(centers, nacreWhorlCenter{
			x:        cx,
			y:        cy,
			radius:   rad,
			spin:     spin,
			strength: strength,
			arms:     arms,
			phaseOff: float64(i) * 1.17,
		})
	}

	// 2. Allocate 3D G-Buffer over deep obsidian-peacock shell matrix
	darkAbaloneBase := color.RGBA{R: 14, G: 22, B: 30, A: 255}
	gb := render.NewFiberGBuffer(w, h, darkAbaloneBase)

	deepPeacock := color.RGBA{R: 12, G: 44, B: 58, A: 255}
	abyssalIndigo := color.RGBA{R: 22, G: 26, B: 64, A: 255}
	emeraldNacre := color.RGBA{R: 18, G: 112, B: 96, A: 255}
	warmNacreCream := color.RGBA{R: 246, G: 236, B: 214, A: 255}
	champagneGold := color.RGBA{R: 244, G: 204, B: 142, A: 255}

	const goldenB = 0.3063489 // Golden logarithmic spiral pitch ln(phi)/(pi/2)

	// Evaluate terraced 3D nacreous shell height, local thin-film thickness, and spiral phase at (x, y)
	evalTerraceField := func(x, y float64) (zHeight, filmNm, spiralPhase, nearestU float64) {
		nx, ny := x/fw, y/fh
		warpX := noise.FBM2D(nx*2.8+0.7+loopX, ny*2.8-0.4+loopY, 3, 2.0, 0.5) * 16.0 * scale
		warpY := noise.FBM2D(nx*2.8-1.9-loopX, ny*2.8+1.3+loopY, 3, 2.0, 0.5) * 16.0 * scale
		wx, wy := x+warpX, y+warpY

		// Primary golden-ratio logarithmic spiral ammonite chamber
		c0 := centers[0]
		dx0, dy0 := wx-c0.x, wy-c0.y
		r0 := math.Hypot(dx0, dy0) + 3.0*scale
		theta0 := math.Atan2(dy0, dx0)
		logR0 := math.Log(r0 / (12.0 * scale))
		logSpiral := c0.arms*theta0 - (logR0/goldenB)*2.4 + phi*c0.spin

		// Superpose multi-center nacreous whorls
		whorlSum := 0.0
		wgtSum := 0.0
		bestU := 0.0
		for i := range centers {
			wc := &centers[i]
			dx := wx - wc.x
			dy := wy - wc.y
			d := math.Hypot(dx, dy)
			u := math.Exp(-d / (wc.radius * 0.85))
			if i > 0 && u > bestU {
				bestU = u
			} else if i == 0 && u*0.85 > bestU {
				bestU = u * 0.85
			}
			ang := math.Atan2(dy, dx)
			localLog := math.Log((d+4.0*scale)/(10.0*scale)) / goldenB
			wavePhase := wc.arms*ang*wc.spin - localLog*2.15 + wc.phaseOff - phi*wc.spin*0.5
			flute := math.Sin(wavePhase) + 0.32*math.Sin(2.0*wavePhase+0.6)
			wgt := wc.strength * (0.18 + 0.82*u)
			whorlSum += wgt * flute
			wgtSum += wgt
		}
		if wgtSum > 1e-6 {
			whorlSum /= wgtSum
		}

		// Concentric terraced growth rings with smooth C1 ridges
		ringArg := logSpiral*0.65 + whorlSum*2.2 + (r0/(34.0*scale))*2.1
		sinRing := math.Sin(ringArg)
		cosRing := math.Cos(ringArg)
		// Smooth C1 asymmetric terrace profile (steep nacreous lamella step + broad pearly shelf)
		terrace := 0.5 + 0.5*sinRing
		terraceC1 := terrace*terrace*(3.0-2.0*terrace) + 0.18*math.Sin(2.0*ringArg-0.45)
		microLamella := 0.5 + 0.5*math.Sin(ringArg*3.0+cosRing*0.8)

		zBase := 0.85 + 2.85*terraceC1 + 0.55*microLamella + 1.15*bestU
		// Vary thin-film thickness from 0.65x to 1.45x across growth bands and spiral angle
		bandMod := 0.5 + 0.5*math.Sin(ringArg+0.35*theta0+whorlSum*1.4)
		angleMod := 0.5 + 0.5*math.Cos(theta0*2.0-logR0*1.3+phi)
		filmFactor := 0.65 + 0.80*(0.68*bandMod+0.32*angleMod)
		filmNm = in.filmThicknessNm * common.Constrain(filmFactor, 0.65, 1.45)

		return zBase, filmNm, ringArg, bestU
	}

	// 3. Sculpt underlying 3D terraced nacreous growth rings directly into gb
	eps := 1.25 * scale
	if eps < 0.8 {
		eps = 0.8
	}
	for y := 0; y < h; y++ {
		fy := float64(y) + 0.5
		rowOff := y * w
		for x := 0; x < w; x++ {
			fx := float64(x) + 0.5
			idx := rowOff + x

			zC, filmNm, ringArg, nearU := evalTerraceField(fx, fy)
			zR, _, _, _ := evalTerraceField(fx+eps, fy)
			zL, _, _, _ := evalTerraceField(fx-eps, fy)
			zD, _, _, _ := evalTerraceField(fx, fy+eps)
			zU, _, _, _ := evalTerraceField(fx, fy-eps)

			dzdx := (zR - zL) / (2.0 * eps)
			dzdy := (zD - zU) / (2.0 * eps)
			nx := -dzdx * 2.15
			ny := -dzdy * 2.15
			nz := 1.0
			nLen := math.Sqrt(nx*nx + ny*ny + nz*nz)
			if nLen > 1e-9 {
				nx, ny, nz = nx/nLen, ny/nLen, nz/nLen
			}

			// Sample physical Airy thin-film linear RGB across terraced shell chambers
			cosTheta := common.Constrain(nz*(0.78+0.22*math.Cos(ringArg*0.5)), 0.18, 1.0)
			tfR, tfG, tfB := render.SampleThinFilmLinearRGB(cosTheta, filmNm, 1.52)

			ridgeLift := common.Smoothstep(1.2, 4.2, zC)
			baseMatrix := render.LerpOklab(deepPeacock, abyssalIndigo, 0.5+0.5*math.Sin(ringArg*0.4))
			baseMatrix = render.LerpOklab(baseMatrix, emeraldNacre, ridgeLift*0.55)
			mR, mG, mB, _ := render.ColorToLinearRGB(baseMatrix)

			// Blend deep peacock shell matrix with luminous pearlescent thin-film interference
			iridMix := 0.38 + 0.46*ridgeLift + 0.12*nearU
			albR := mR*(1.0-iridMix) + tfR*0.88*iridMix
			albG := mG*(1.0-iridMix) + tfG*0.92*iridMix
			albB := mB*(1.0-iridMix) + tfB*0.95*iridMix

			gb.Z[idx] = zC
			gb.Nx[idx] = nx
			gb.Ny[idx] = ny
			gb.Nz[idx] = nz
			gb.AlbR[idx] = common.Constrain(albR, 0.01, 0.98)
			gb.AlbG[idx] = common.Constrain(albG, 0.01, 0.98)
			gb.AlbB[idx] = common.Constrain(albB, 0.01, 0.98)
			gb.Roughness[idx] = 0.24 + 0.26*(1.0-ridgeLift)
		}
	}

	// Swirling nacreous tangent vector field along the logarithmic spiral chambers and Voronoi whorls
	evalTangentAngle := func(x, y float64) float64 {
		nx, ny := x/fw, y/fh
		curl := noise.Curl2D(nx*2.1+0.4+loopX*0.6, ny*2.1-0.5+loopY*0.6, 0.01)
		vx := curl.X * 0.24
		vy := curl.Y * 0.24

		for i := range centers {
			wc := &centers[i]
			dx := x - wc.x
			dy := y - wc.y
			d := math.Hypot(dx, dy) + 1.5*scale
			u := math.Exp(-d / (wc.radius * 1.15))
			wgt := wc.strength * (0.14 + 1.65*u)

			theta := math.Atan2(dy, dx)
			// Golden logarithmic spiral tangent angle: theta + spin*(pi/2 - atan(goldenB))
			pitchAngle := math.Pi*0.415 + 0.085*math.Sin(wc.arms*theta-math.Log(d/(9.0*scale))*1.8+wc.phaseOff)
			if isAnim {
				pitchAngle += 0.14 * math.Sin(wc.arms*theta-wc.spin*phi-d/(28.0*scale))
			}
			tang := theta + wc.spin*pitchAngle
			vx += math.Cos(tang) * wgt
			vy += math.Sin(tang) * wgt
		}
		return math.Atan2(vy, vx)
	}

	// 4. Deterministically seed fiber roots and sort from low elevation (outer valleys) to high elevation (inner spiral crests)
	scaledFibers := int(math.Round(float64(in.fiberCount) * scale))
	if scaledFibers < 2500 {
		scaledFibers = 2500
	}

	seeds := make([]nacreFiberSeed, scaledFibers)
	goldenPhi2X := 0.7548776662466927
	goldenPhi2Y := 0.5698402909980532

	for i := 0; i < scaledFibers; i++ {
		fi := float64(i) + 0.5
		// Quasi-random R2 low-discrepancy sequence with slight jitter for organic coverage
		u := math.Mod(fi*goldenPhi2X+0.13*rng.Float64(), 1.0)
		v := math.Mod(fi*goldenPhi2Y+0.13*rng.Float64(), 1.0)

		// Bias 22% of fibers into spiral whorl hearts for lush nacreous plume density
		var x0, y0 float64
		if i%9 < 2 {
			wc := centers[i%len(centers)]
			r := wc.radius * 0.88 * math.Sqrt(rng.Float64())
			ang := rng.Float64() * 2.0 * math.Pi
			x0 = common.Constrain(wc.x+math.Cos(ang)*r, 0.02*fw, 0.98*fw)
			y0 = common.Constrain(wc.y+math.Sin(ang)*r, 0.02*fh, 0.98*fh)
		} else {
			x0 = (0.02 + 0.96*u) * fw
			y0 = (0.02 + 0.96*v) * fh
		}

		zElev, filmNm, ringArg, nearU := evalTerraceField(x0, y0)
		cosV := common.Constrain(0.35+0.55*(0.5+0.5*math.Cos(ringArg*0.7+nearU*1.8)), 0.15, 0.98)

		seeds[i] = nacreFiberSeed{
			x:          x0,
			y:          y0,
			elev:       zElev + 1.35*nearU,
			thickness:  filmNm,
			cosView:    cosV,
			whorlDist:  nearU,
			strokeSeed: rng.Float64() * 2.0 * math.Pi,
			lenMod:     0.78 + 0.54*rng.Float64(),
			radMod:     0.72 + 0.58*rng.Float64(),
			bendMod:    (rng.Float64() - 0.5) * 0.055,
		}
	}

	sort.Slice(seeds, func(i, j int) bool {
		return seeds[i].elev < seeds[j].elev
	})

	// Rasterize overlapping 3D aragonite silk-bristle plumes into gb
	const numCtrlPts = 7
	var ctrlPts [numCtrlPts]render.Point
	stepDist := 4.8 * scale

	peacockCyan := color.RGBA{R: 16, G: 224, B: 208, A: 255}
	abaloneRose := color.RGBA{R: 232, G: 68, B: 148, A: 255}
	imperialViolet := color.RGBA{R: 96, G: 58, B: 214, A: 255}
	obsidianRoot := color.RGBA{R: 8, G: 14, B: 24, A: 255}

	for i := range seeds {
		sd := &seeds[i]
		cx, cy := sd.x, sd.y
		if isAnim {
			ang0 := evalTangentAngle(cx, cy)
			adv := 4.5 * scale * math.Sin(phi-sd.strokeSeed)
			cx += math.Cos(ang0) * adv
			cy += math.Sin(ang0) * adv
		}

		ctrlPts[0] = render.Pt(cx, cy)
		segStep := stepDist * sd.lenMod
		for k := 1; k < numCtrlPts; k++ {
			tk := float64(k) / float64(numCtrlPts-1)
			ang := evalTangentAngle(cx, cy) + sd.bendMod*float64(k)
			// Coherent wave undulation along each aragonite plume
			undul := 0.19 * math.Sin(tk*math.Pi*2.2+sd.strokeSeed-phi)
			ang += undul
			cx += math.Cos(ang) * segStep
			cy += math.Sin(ang) * segStep
			ctrlPts[k] = render.Pt(cx, cy)
		}

		// Macro spiral growth band determines the dominant structural interference family:
		// peacock-emerald -> abalone-rose/violet -> champagne-gold
		bandPhase := 0.5 + 0.5*math.Sin(sd.elev*1.35+sd.whorlDist*3.2)
		var rootCol, midCol, tipCol color.RGBA

		// Deep dark velvet-abalone root so spiral valleys have strong 3D shadow contrast
		rootBase := render.LerpOklab(abyssalIndigo, deepPeacock, bandPhase)
		rootCol = render.LerpOklab(obsidianRoot, rootBase, 0.52)

		// Sample physical thin-film interference and blend with structured abalone bands
		tfCol := render.SampleThinFilmIridescence(sd.cosView, sd.thickness, 1.52)
		var bandCol color.RGBA
		if bandPhase < 0.42 {
			bandCol = render.LerpOklab(emeraldNacre, peacockCyan, bandPhase/0.42)
		} else if bandPhase < 0.76 {
			bandCol = render.LerpOklab(peacockCyan, abaloneRose, (bandPhase-0.42)/0.34)
		} else {
			bandCol = render.LerpOklab(abaloneRose, imperialViolet, (bandPhase-0.76)/0.24)
		}
		midCol = render.LerpOklab(bandCol, tfCol, 0.45)

		// Luminous crest tips on elevated spiral ridges; deep jewel tips in valleys
		ridgeHighlight := common.Smoothstep(2.1, 4.4, sd.elev)
		crestTip := render.LerpOklab(champagneGold, warmNacreCream, 0.45*sd.whorlDist)
		tipCol = render.LerpOklab(midCol, crestTip, 0.25+0.62*ridgeHighlight)

		zRoot := sd.elev - 0.35
		if zRoot < 0.15 {
			zRoot = 0.15
		}
		zTip := sd.elev + 2.15 + 1.45*ridgeHighlight
		maxRadius := (0.72 + 0.48*sd.radMod) * scale
		roughness := 0.18 + 0.16*(1.0-ridgeHighlight)

		gb.RasterizeFiber3D(
			ctrlPts[:],
			zRoot,
			zTip,
			maxRadius,
			rootCol,
			midCol,
			tipCol,
			0.58,
			roughness,
		)
	}

	// 5. Resolve G-Buffer into linear HDR Canvas using ResolveToCanvasAdvanced
	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	horizonRad := int(math.Round(16.0 * scale))
	if horizonRad < 4 {
		horizonRad = 4
	}
	matCfg := render.DefaultGBufferMaterialConfig([3]float64{-0.48, -0.58, 0.66})
	matCfg.HorizonShadowStrength = 0.72
	matCfg.HorizonRadius = horizonRad
	matCfg.SSSStrength = 0.45
	matCfg.SSSTint = color.RGBA{R: 90, G: 245, B: 210, A: 255}
	matCfg.IridescenceStrength = 0.52
	matCfg.FilmThicknessNm = in.filmThicknessNm
	matCfg.FilmIOR = 1.52

	gb.ResolveToCanvasAdvanced(rc, matCfg)

	// 6. Add luminous specular pearl cabochons at the Voronoi-relaxed whorl centers
	pearlShadow := color.RGBA{R: 10, G: 20, B: 28, A: 255}
	pearlSpecular := color.RGBA{R: 255, G: 252, B: 244, A: 255}

	for i := range centers {
		wc := &centers[i]
		cabRad := (7.5 + 3.5*wc.strength) * scale
		if i == 0 {
			cabRad = 13.5 * scale
		}

		// Soft recessed nacreous socket shadow
		rc.DrawGlowDot(wc.x+1.3*scale, wc.y+1.6*scale, cabRad*1.45, pearlShadow, 0.72, 0.75)

		// Concentric thin-film iridescent nacre cabochon layers
		const cabRings = 7
		for rIdx := cabRings; rIdx >= 1; rIdx-- {
			tRing := float64(rIdx) / float64(cabRings)
			nzSphere := math.Sqrt(math.Max(0.0, 1.0-tRing*tRing*0.82))
			localNm := in.filmThicknessNm * (0.78 + 0.52*nzSphere + 0.08*float64(i))
			ringCol := render.SampleThinFilmIridescence(nzSphere, localNm, 1.52)
			ringCol = render.LerpOklab(ringCol, warmNacreCream, (1.0-tRing)*0.55)
			rc.DrawGlowDot(wc.x, wc.y, cabRad*tRing, ringCol, 0.85, 0.82)
		}

		// Glossy off-center key specular highlight + secondary rim pinpoint
		hlX := wc.x - cabRad*0.30
		hlY := wc.y - cabRad*0.34
		rc.DrawGlowDot(hlX, hlY, cabRad*0.42, pearlSpecular, 0.96, 0.90)
		rc.DrawGlowDot(wc.x+cabRad*0.26, wc.y+cabRad*0.28, cabRad*0.24, matCfg.SSSTint, 0.68, 0.78)
	}

	// Frame with HDR Gaussian bloom and recessed museum deckle border
	rc.ApplyBloom(0.78, 4.0*scale, 0.16)
	rc.ApplyDeckleBorder(0.042, 0.95, color.RGBA{R: 242, G: 236, B: 225, A: 255}, seed)

	c.MergeRenderCanvas(rc)
}
