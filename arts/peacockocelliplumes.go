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

type peacockOcelliPlumes struct {
	ocelli    int
	barbCount int
	filmNm    float64
}

// NewPeacockOcelliPlumes returns a peacockOcelliPlumes engine that sculpts a 3D G-Buffer
// photonic peacock-feather ocelli array and horizon-shadowed aragonite/keratin barbules
// with Airy thin-film structural color and sapphire-gold cabochon eyes.
func NewPeacockOcelliPlumes(ocelli, barbCount int, filmNm float64) *peacockOcelliPlumes {
	if ocelli <= 0 {
		ocelli = 7
	}
	if barbCount <= 0 {
		barbCount = 18000
	}
	if filmNm <= 0 {
		filmNm = 390.0
	}
	return &peacockOcelliPlumes{
		ocelli:    common.ConstrainInt(ocelli, 3, 14),
		barbCount: common.ConstrainInt(barbCount, 4000, 32000),
		filmNm:    common.Constrain(filmNm, 220.0, 620.0),
	}
}

type ocellusEyeCenter struct {
	x, y     float64
	radius   float64
	tilt     float64
	strength float64
	phaseOff float64
}

type barbPlumeSeed struct {
	x, y       float64
	elev       float64
	thickness  float64
	cosView    float64
	eyeRing    float64
	strokeSeed float64
	lenMod     float64
	radMod     float64
	bendMod    float64
}

// Generative renders the static PeacockOcelliPlumes composition.
func (pop *peacockOcelliPlumes) Generative(c *generativeart.Canva) {
	pop.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless looping PeacockOcelliPlumes animation.
func (pop *peacockOcelliPlumes) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
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
		seed = 20262503
	}
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	isAnim := totalFrames > 1
	phi := 2.0 * math.Pi * phase
	loopX := 0.07 * math.Sin(phi)
	loopY := 0.07 * (1.0 - math.Cos(phi))

	// 1. Seed peacock ocellus centers via Poisson-Disk + Centroidal Voronoi relaxation
	densityFn := func(x, y float64) float64 {
		nx := (x/fw - 0.50) * 2.0
		ny := (y/fh - 0.50) * 2.0
		r := math.Hypot(nx, ny)
		return common.Constrain(0.45+0.55*(1.0-0.55*r), 0.1, 1.0)
	}

	approxSpacing := minDim / math.Max(2.3, math.Sqrt(float64(pop.ocelli))*1.18)
	rawPts := render.SamplePoissonDisk2D(fw*0.74, fh*0.74, approxSpacing, densityFn, rng)
	for i := range rawPts {
		rawPts[i].X += fw * 0.13
		rawPts[i].Y += fh * 0.13
	}
	relaxedPts := render.RelaxCentroidalVoronoi(rawPts, fw, fh, 96, densityFn, 6)

	focalX, focalY := 0.50*fw, 0.49*fh
	sort.Slice(relaxedPts, func(i, j int) bool {
		di := math.Hypot(relaxedPts[i].X-focalX, relaxedPts[i].Y-focalY)
		dj := math.Hypot(relaxedPts[j].X-focalX, relaxedPts[j].Y-focalY)
		return di < dj
	})

	centers := make([]ocellusEyeCenter, 0, pop.ocelli)
	goldenAngle := math.Pi * (3.0 - math.Sqrt(5.0))
	ptCursor := 0
	for i := 0; i < pop.ocelli; i++ {
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
				fi := float64(i)
				r := (0.24 + 0.08*math.Sqrt(fi/float64(pop.ocelli))) * minDim
				ang := fi*goldenAngle + 0.4
				cx = common.Constrain(focalX+math.Cos(ang)*r, 0.15*fw, 0.85*fw)
				cy = common.Constrain(focalY+math.Sin(ang)*r, 0.15*fh, 0.85*fh)
			}
		}

		rad := (0.21 + 0.06*rng.Float64()) * minDim
		str := 0.90 + 0.30*rng.Float64()
		if i == 0 {
			rad = 0.29 * minDim
			str = 1.38
		}
		dxC := cx - focalX
		dyC := cy - focalY
		tilt := math.Atan2(dyC, dxC)
		if i == 0 {
			tilt = -math.Pi * 0.5
		}

		centers = append(centers, ocellusEyeCenter{
			x:        cx,
			y:        cy,
			radius:   rad,
			tilt:     tilt,
			strength: str,
			phaseOff: float64(i) * goldenAngle,
		})
	}

	// 2. Evaluate continuous 3D peacock ocellus terrace relief and photonic melanin-rod film thickness
	evalOcellusField := func(x, y float64) (zHeight, filmNm, ringPhase, nearU float64) {
		nx, ny := x/fw, y/fh
		warpX := noise.FBM2D(nx*2.6+0.8+loopX, ny*2.6-0.5+loopY, 3, 2.0, 0.5)
		warpY := noise.FBM2D(nx*2.6-1.9-loopX, ny*2.6+1.4-loopY, 3, 2.0, 0.5)
		wx := x + warpX*14.0*scale
		wy := y + warpY*14.0*scale

		bestScore := 1e9
		bestIdx := 0
		for i := range centers {
			ec := &centers[i]
			dx := wx - ec.x
			dy := wy - ec.y
			// Cardioid/heart-shaped peacock ocellus metric
			ca, sa := math.Cos(ec.tilt), math.Sin(ec.tilt)
			u := (dx*ca + dy*sa) / ec.radius
			v := (-dx*sa + dy*ca) / ec.radius
			uCard := u - 0.16*v*v
			dNorm := math.Hypot(uCard, v*1.08) / ec.strength
			if dNorm < bestScore {
				bestScore = dNorm
				bestIdx = i
			}
		}

		ec := &centers[bestIdx]
		dx := wx - ec.x
		dy := wy - ec.y
		ca, sa := math.Cos(ec.tilt), math.Sin(ec.tilt)
		u := (dx*ca + dy*sa) / ec.radius
		v := (-dx*sa + dy*ca) / ec.radius
		rOcellus := math.Hypot(u-0.16*v*v, v*1.08)

		nearU = math.Exp(-rOcellus * 1.15)
		ringArg := rOcellus*7.5 - ec.phaseOff
		if isAnim {
			ringArg -= phi
		}

		s := math.Sin(ringArg)
		terrace := 0.5 + 0.5*s + 0.18*math.Sin(2.0*ringArg-0.4)
		zHeight = (0.75 + 2.65*terrace + 1.55*nearU) * (0.68 + 0.32*ec.strength)

		// Melanin lattice spacing sweeps from royal sapphire (340nm) -> turquoise (410nm) -> emerald/gold (485nm) -> bronze-magenta (540nm)
		filmNm = pop.filmNm*(0.76+0.44*math.Cos(ringArg*0.75)+0.16*warpX) + 55.0*(1.0-nearU)
		return zHeight, filmNm, ringArg, nearU
	}

	// 3. Initialize 3D G-Buffer with iridescent peacock matrix under-layer
	gb := render.NewFiberGBuffer(w, h, c.Opts().Background())
	deepSapphire := color.RGBA{R: 12, G: 28, B: 88, A: 255}
	abyssalTeal := color.RGBA{R: 8, G: 44, B: 58, A: 255}
	emeraldGold := color.RGBA{R: 28, G: 168, B: 118, A: 255}
	champagneGold := color.RGBA{R: 248, G: 204, B: 96, A: 255}
	warmCream := color.RGBA{R: 254, G: 246, B: 230, A: 255}

	eps := 1.5 * scale
	for y := 0; y < h; y++ {
		fy := float64(y) + 0.5
		rowOff := y * w
		for x := 0; x < w; x++ {
			fx := float64(x) + 0.5
			idx := rowOff + x

			zC, filmNm, ringArg, nearU := evalOcellusField(fx, fy)
			zR, _, _, _ := evalOcellusField(fx+eps, fy)
			zD, _, _, _ := evalOcellusField(fx, fy+eps)

			dzdx := (zR - zC) / eps
			dzdy := (zD - zC) / eps
			nx, ny, nz := -dzdx*0.55, -dzdy*0.55, 1.0
			nLen := math.Sqrt(nx*nx + ny*ny + nz*nz)
			if nLen > 1e-9 {
				nx, ny, nz = nx/nLen, ny/nLen, nz/nLen
			}

			cosTheta := common.Constrain(nz*(0.78+0.22*math.Cos(ringArg*0.5)), 0.18, 1.0)
			tfR, tfG, tfB := render.SampleThinFilmLinearRGB(cosTheta, filmNm, 1.56)

			ridgeLift := common.Smoothstep(1.2, 4.2, zC)
			baseMatrix := render.LerpOklab(abyssalTeal, deepSapphire, 0.5+0.5*math.Sin(ringArg*0.4))
			baseMatrix = render.LerpOklab(baseMatrix, emeraldGold, ridgeLift*0.55)
			mR, mG, mB, _ := render.ColorToLinearRGB(baseMatrix)

			iridMix := 0.42 + 0.44*ridgeLift + 0.12*nearU
			albR := mR*(1.0-iridMix) + tfR*0.90*iridMix
			albG := mG*(1.0-iridMix) + tfG*0.92*iridMix
			albB := mB*(1.0-iridMix) + tfB*0.95*iridMix

			gb.Z[idx] = zC
			gb.Nx[idx] = nx
			gb.Ny[idx] = ny
			gb.Nz[idx] = nz
			gb.AlbR[idx] = common.Constrain(albR, 0.01, 0.98)
			gb.AlbG[idx] = common.Constrain(albG, 0.01, 0.98)
			gb.AlbB[idx] = common.Constrain(albB, 0.01, 0.98)
			gb.Roughness[idx] = 0.22 + 0.24*(1.0-ridgeLift)
		}
	}

	// Feather barbule tangent vector field: radiates in graceful dipole/fountain arches from each ocellusrachis
	evalBarbAngle := func(x, y float64) float64 {
		nx, ny := x/fw, y/fh
		curl := noise.Curl2D(nx*2.0+0.3+loopX*0.5, ny*2.0-0.4+loopY*0.5, 0.01)
		vx := curl.X * 0.22
		vy := curl.Y * 0.22

		for i := range centers {
			ec := &centers[i]
			dx := x - ec.x
			dy := y - ec.y
			d := math.Hypot(dx, dy) + 1.5*scale
			u := math.Exp(-d / (ec.radius * 1.18))
			wgt := ec.strength * (0.15 + 1.65*u)

			theta := math.Atan2(dy, dx)
			relAng := theta - ec.tilt
			// Peacock feather barbs fan outward from the central rachis and curve into concentric ocellus arches
			archBend := 0.42 * math.Sin(relAng) * (0.45 + 0.55*math.Cos(d/(ec.radius*0.35)))
			if isAnim {
				archBend += 0.12 * math.Sin(2.0*relAng-phi-d/(28.0*scale))
			}
			tang := theta + archBend + math.Pi*0.32*math.Sin(relAng)
			vx += math.Cos(tang) * wgt
			vy += math.Sin(tang) * wgt
		}
		return math.Atan2(vy, vx)
	}

	scaledBarbs := int(math.Round(float64(pop.barbCount) * scale))
	if scaledBarbs < 2500 {
		scaledBarbs = 2500
	}

	seeds := make([]barbPlumeSeed, scaledBarbs)
	goldenPhi2X := 0.7548776662466927
	goldenPhi2Y := 0.5698402909980532

	for i := 0; i < scaledBarbs; i++ {
		fi := float64(i) + 0.5
		u := math.Mod(fi*goldenPhi2X+0.13*rng.Float64(), 1.0)
		v := math.Mod(fi*goldenPhi2Y+0.13*rng.Float64(), 1.0)

		var x0, y0 float64
		if i%8 < 2 {
			ec := centers[i%len(centers)]
			r := ec.radius * 0.86 * math.Sqrt(rng.Float64())
			ang := rng.Float64() * 2.0 * math.Pi
			x0 = common.Constrain(ec.x+math.Cos(ang)*r, 0.02*fw, 0.98*fw)
			y0 = common.Constrain(ec.y+math.Sin(ang)*r, 0.02*fh, 0.98*fh)
		} else {
			x0 = (0.02 + 0.96*u) * fw
			y0 = (0.02 + 0.96*v) * fh
		}

		zElev, filmNm, ringArg, nearU := evalOcellusField(x0, y0)
		cosV := common.Constrain(0.35+0.55*(0.5+0.5*math.Cos(ringArg*0.7+nearU*1.8)), 0.15, 0.98)

		seeds[i] = barbPlumeSeed{
			x:          x0,
			y:          y0,
			elev:       zElev + 1.35*nearU,
			thickness:  filmNm,
			cosView:    cosV,
			eyeRing:    nearU,
			strokeSeed: rng.Float64() * 2.0 * math.Pi,
			lenMod:     0.78 + 0.54*rng.Float64(),
			radMod:     0.72 + 0.58*rng.Float64(),
			bendMod:    (rng.Float64() - 0.5) * 0.052,
		}
	}

	sort.Slice(seeds, func(i, j int) bool {
		return seeds[i].elev < seeds[j].elev
	})

	schema := c.Opts().ColorSchema()
	if len(schema) == 0 {
		schema = common.Oceanic
	}
	palGrad := render.NewMultiStopGradient(schema...)

	const numCtrlPts = 7
	var ctrlPts [numCtrlPts]render.Point
	stepDist := 4.8 * scale

	peacockCyan := color.RGBA{R: 18, G: 226, B: 214, A: 255}
	bronzeCopper := color.RGBA{R: 224, G: 132, B: 46, A: 255}
	royalViolet := color.RGBA{R: 88, G: 48, B: 212, A: 255}
	obsidianRoot := color.RGBA{R: 7, G: 12, B: 22, A: 255}

	for i := range seeds {
		sd := &seeds[i]
		cx, cy := sd.x, sd.y
		if isAnim {
			ang0 := evalBarbAngle(cx, cy)
			adv := 4.2 * scale * math.Sin(phi-sd.strokeSeed)
			cx += math.Cos(ang0) * adv
			cy += math.Sin(ang0) * adv
		}

		ctrlPts[0] = render.Pt(cx, cy)
		segStep := stepDist * sd.lenMod
		for k := 1; k < numCtrlPts; k++ {
			tk := float64(k) / float64(numCtrlPts-1)
			ang := evalBarbAngle(cx, cy) + sd.bendMod*float64(k)
			undul := 0.18 * math.Sin(tk*math.Pi*2.2+sd.strokeSeed-phi)
			ang += undul
			cx += math.Cos(ang) * segStep
			cy += math.Sin(ang) * segStep
			ctrlPts[k] = render.Pt(cx, cy)
		}

		bandPhase := 0.5 + 0.5*math.Sin(sd.elev*1.35+sd.eyeRing*3.5)
		rootBase := render.LerpOklab(deepSapphire, abyssalTeal, bandPhase)
		rootCol := render.LerpOklab(obsidianRoot, rootBase, 0.50)

		tfCol := render.SampleThinFilmIridescence(sd.cosView, sd.thickness, 1.56)
		var bandCol color.RGBA
		if bandPhase < 0.38 {
			bandCol = render.LerpOklab(royalViolet, peacockCyan, bandPhase/0.38)
		} else if bandPhase < 0.74 {
			bandCol = render.LerpOklab(peacockCyan, emeraldGold, (bandPhase-0.38)/0.36)
		} else {
			bandCol = render.LerpOklab(emeraldGold, bronzeCopper, (bandPhase-0.74)/0.26)
		}
		palCol := palGrad.SampleOklab(bandPhase)
		bandCol = render.LerpOklab(bandCol, palCol, 0.24)
		midCol := render.LerpOklab(bandCol, tfCol, 0.46)

		ridgeHighlight := common.Smoothstep(2.1, 4.4, sd.elev)
		crestTip := render.LerpOklab(champagneGold, warmCream, 0.45*sd.eyeRing)
		tipCol := render.LerpOklab(midCol, crestTip, 0.26+0.62*ridgeHighlight)

		zRoot := math.Max(0.15, sd.elev-0.35)
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

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	horizonRad := common.MaxInt(4, int(math.Round(16.0*scale)))
	matCfg := render.DefaultGBufferMaterialConfig([3]float64{-0.48, -0.58, 0.66})
	matCfg.HorizonShadowStrength = 0.72
	matCfg.HorizonRadius = horizonRad
	matCfg.SSSStrength = 0.45
	matCfg.SSSTint = color.RGBA{R: 68, G: 242, B: 214, A: 255}
	matCfg.IridescenceStrength = 0.54
	matCfg.FilmThicknessNm = pop.filmNm
	matCfg.FilmIOR = 1.56

	gb.ResolveToCanvasAdvanced(rc, matCfg)

	// 6. Add royal sapphire & gold iridescent ocellus cabochon jewels at each eye heart
	eyeShadow := color.RGBA{R: 8, G: 16, B: 28, A: 255}
	eyeSpecular := color.RGBA{R: 255, G: 252, B: 244, A: 255}

	for i := range centers {
		ec := &centers[i]
		cabRad := (8.0 + 3.8*ec.strength) * scale
		if i == 0 {
			cabRad = 14.5 * scale
		}

		rc.DrawGlowDot(ec.x+1.4*scale, ec.y+1.7*scale, cabRad*1.48, eyeShadow, 0.74, 0.76)

		const cabRings = 7
		for rIdx := cabRings; rIdx >= 1; rIdx-- {
			tRing := float64(rIdx) / float64(cabRings)
			nzSphere := math.Sqrt(math.Max(0.0, 1.0-tRing*tRing*0.82))
			localNm := pop.filmNm * (0.76 + 0.54*nzSphere + 0.09*float64(i))
			ringCol := render.SampleThinFilmIridescence(nzSphere, localNm, 1.56)
			ringCol = render.LerpOklab(ringCol, deepSapphire, (1.0-tRing)*0.48)
			rc.DrawGlowDot(ec.x, ec.y, cabRad*tRing, ringCol, 0.88, 0.84)
		}

		hlX := ec.x - cabRad*0.30
		hlY := ec.y - cabRad*0.34
		rc.DrawGlowDot(hlX, hlY, cabRad*0.42, eyeSpecular, 0.96, 0.90)
		rc.DrawGlowDot(ec.x+cabRad*0.26, ec.y+cabRad*0.28, cabRad*0.25, champagneGold, 0.72, 0.80)
	}

	rc.ApplyBloom(0.78, 4.0*scale, 0.16)
	rc.ApplyDeckleBorder(0.042, 0.95, color.RGBA{R: 242, G: 236, B: 225, A: 255}, seed)
	c.MergeRenderCanvas(rc)
}
