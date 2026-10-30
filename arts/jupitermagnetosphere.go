package arts

import (
	"image/color"
	"math"
	"math/rand"
	"runtime"
	"sync"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type jupiterMagnetosphere struct {
	bands     int
	vortices  int
	licLength float64
}

// NewJupiterMagnetosphere returns a jupiterMagnetosphere engine that renders a Juno-inspired
// Jovian perijove atmospheric & magnetospheric flow field combining alternating zonal shear jets,
// a Great Red Spot primary super-anticyclone, Kármán vortex street ovals, Fast Line Integral
// Convolution (LIC/OFLIC), physical thin-film eyewall iridescence, Jobard-Lefer magnetic filaments,
// and 2.5D luminance relief shading.
func NewJupiterMagnetosphere(bands, vortices int, licLength float64) *jupiterMagnetosphere {
	if bands <= 0 {
		bands = 9
	}
	if vortices <= 0 {
		vortices = 16
	}
	if licLength <= 0 {
		licLength = 22.0
	}
	return &jupiterMagnetosphere{
		bands:     bands,
		vortices:  vortices,
		licLength: licLength,
	}
}

type jupiterVortex struct {
	x, y        float64
	rx, ry      float64
	cosA, sinA  float64
	spin        float64 // +1 counter-clockwise, -1 clockwise
	strength    float64
	isPrimary   bool
	isWarmStorm bool
	phaseOffset float64
}

// Generative renders a static JupiterMagnetosphere composition.
func (jm *jupiterMagnetosphere) Generative(c *generativeart.Canva) {
	jm.GenerativeFrame(c, 0, 1, 0)
}

// GenerativeFrame renders frame frameIdx of a seamless looping animation over phase in [0, 1).
func (jm *jupiterMagnetosphere) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
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
		seed = 20261030
	}
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	isAnim := totalFrames > 1
	twoPiPhase := 2.0 * math.Pi * phase
	loopX := 0.085 * math.Sin(twoPiPhase)
	loopY := 0.085 * (1.0 - math.Cos(twoPiPhase))

	// Juno Perijove Palette:
	// Deep abyssal Prussian/cobalt blues, turquoise-teal belts, creamy alabaster zones,
	// burnt sienna/terracotta storms, and luminous amber-gold vortex eyewalls.
	abyssalPrussian := color.RGBA{R: 7, G: 19, B: 44, A: 255}     // #07132c
	cobaltMidnight := color.RGBA{R: 16, G: 44, B: 96, A: 255}     // #102c60
	ceruleanBelt := color.RGBA{R: 34, G: 102, B: 158, A: 255}     // #22669e
	turquoiseTeal := color.RGBA{R: 28, G: 176, B: 188, A: 255}    // #1cb0bc
	creamyAlabaster := color.RGBA{R: 246, G: 238, B: 222, A: 255} // #f6eede
	warmIvory := color.RGBA{R: 232, G: 210, B: 175, A: 255}       // #e8d2af
	amberGold := color.RGBA{R: 248, G: 182, B: 46, A: 255}        // #f8b62e
	terracottaStorm := color.RGBA{R: 224, G: 96, B: 42, A: 255}   // #e0602a
	burntSienna := color.RGBA{R: 168, G: 42, B: 24, A: 255}      // #a82a18
	crimsonCore := color.RGBA{R: 112, G: 20, B: 26, A: 255}      // #70141a
	auroralCyan := color.RGBA{R: 150, G: 248, B: 255, A: 255}     // #96f8ff
	gildedChampagne := color.RGBA{R: 255, G: 236, B: 178, A: 255} // #ffecb2

	macroPalette := render.NewMultiStopGradient(
		abyssalPrussian,
		cobaltMidnight,
		ceruleanBelt,
		turquoiseTeal,
		creamyAlabaster,
		warmIvory,
		amberGold,
		terracottaStorm,
		burntSienna,
	)

	// Tilted Jovian perijove coordinate frame (-15 deg inclination with spherical horizon curvature)
	const tiltAngle = -15.0 * math.Pi / 180.0
	cosTilt := math.Cos(tiltAngle)
	sinTilt := math.Sin(tiltAngle)

	// Transform pixel (x, y) into normalized Jovian (uLon, uLat) in [-0.6, 0.6]^2 with Rossby wave meander
	jovianCoords := func(x, y float64) (uLon, uLat, meander, dMeanderDuLon float64) {
		dx := (x - 0.5*fw) / minDim
		dy := (y - 0.5*fh) / minDim
		uLon = dx*cosTilt - dy*sinTilt
		rawLat := dx*sinTilt + dy*cosTilt
		// Subtle planetary horizon curvature across longitude
		curveOffset := 0.085 * (uLon*uLon - 0.12)

		// Multi-mode Rossby planetary waves along longitude
		k1 := 2.0 * math.Pi * 1.45
		k2 := 2.0 * math.Pi * 2.85
		wPhase1 := k1*uLon - twoPiPhase + 0.35
		wPhase2 := k2*uLon + twoPiPhase*0.5 - 0.70

		latEnv := math.Exp(-rawLat * rawLat * 2.2)
		amp1 := 0.036 * latEnv
		amp2 := 0.015 * latEnv

		meander = curveOffset + amp1*math.Sin(wPhase1) + amp2*math.Cos(wPhase2)
		dMeanderDuLon = 0.17*uLon + amp1*k1*math.Cos(wPhase1) - amp2*k2*math.Sin(wPhase2)
		uLat = rawLat - meander
		return uLon, uLat, meander, dMeanderDuLon
	}

	// Invert (uLon, uLat) back to pixel coordinates (x, y) for placing vortices along shear bands
	jovianToPixel := func(uLon, uLat float64) (float64, float64) {
		k1 := 2.0 * math.Pi * 1.45
		k2 := 2.0 * math.Pi * 2.85
		wPhase1 := k1*uLon - twoPiPhase + 0.35
		wPhase2 := k2*uLon + twoPiPhase*0.5 - 0.70
		latEnv := math.Exp(-uLat * uLat * 2.2)
		meander := 0.085*(uLon*uLon-0.12) + 0.036*latEnv*math.Sin(wPhase1) + 0.015*latEnv*math.Cos(wPhase2)
		rawLat := uLat + meander

		dx := uLon*cosTilt + rawLat*sinTilt
		dy := -uLon*sinTilt + rawLat*cosTilt
		return 0.5*fw + dx*minDim, 0.5*fh + dy*minDim
	}

	// Build vortex list: 1 primary Great Red Spot super-anticyclone + jm.vortices Kármán street ovals
	allVortices := make([]jupiterVortex, 0, jm.vortices+1)

	// Primary Super-Vortex (Great Red Spot / Juno Perijove Anticyclone)
	primULon := -0.11
	primULat := 0.05
	if isAnim {
		primULon += 0.018 * math.Sin(twoPiPhase)
		primULat += 0.012 * (1.0 - math.Cos(twoPiPhase))
	}
	primX, primY := jovianToPixel(primULon, primULat)
	primAngle := tiltAngle + 0.08*math.Sin(twoPiPhase)
	allVortices = append(allVortices, jupiterVortex{
		x:           primX,
		y:           primY,
		rx:          104.0 * scale,
		ry:          72.0 * scale,
		cosA:        math.Cos(primAngle),
		sinA:        math.Sin(primAngle),
		spin:        1.0,
		strength:    2.35,
		isPrimary:   true,
		isWarmStorm: true,
		phaseOffset: 0.0,
	})

	// Place jm.vortices counter-rotating Kármán vortex street ovals ("String of Pearls" & Barges)
	effectiveBands := jm.bands
	if effectiveBands < 2 {
		effectiveBands = 2
	}
	nBands := float64(effectiveBands)
	for i := 0; i < jm.vortices; i++ {
		// Distribute across shear interfaces between zonal bands
		bandSlot := (i % (effectiveBands - 1)) + 1
		uLatBase := -0.44 + 0.88*(float64(bandSlot)/nBands)
		// Stagger Kármán street slightly north/south of the shear line
		stagger := 0.022
		if i%2 == 1 {
			stagger = -0.022
		}
		uLatV := uLatBase + stagger + (rng.Float64()-0.5)*0.025

		// Stratified golden-ratio / jittered longitude placement
		uLonV := -0.45 + 0.90*math.Mod(float64(i)*0.61803398875+0.19*float64(i/effectiveBands)+0.12*rng.Float64(), 1.0)

		// Keep satellite ovals clear of the primary Great Red Spot core
		dLon := uLonV - primULon
		dLat := uLatV - primULat
		if math.Hypot(dLon, dLat*1.35) < 0.23 {
			angleAway := math.Atan2(dLat, dLon)
			uLonV = primULon + 0.25*math.Cos(angleAway)
			uLatV = primULat + 0.20*math.Sin(angleAway)
		}

		vPhase := rng.Float64() * 2.0 * math.Pi
		if isAnim {
			driftDir := 1.0
			if bandSlot%2 == 1 {
				driftDir = -1.0
			}
			uLonV += driftDir * 0.024 * math.Sin(twoPiPhase+vPhase)
			uLatV += 0.014 * (1.0 - math.Cos(twoPiPhase+vPhase))
		}

		vx, vy := jovianToPixel(uLonV, uLatV)
		rx := (32.0 + 26.0*rng.Float64()) * scale
		ry := rx * (0.68 + 0.24*rng.Float64())
		vAng := tiltAngle + (rng.Float64()-0.5)*0.28

		spin := 1.0
		if (i+bandSlot)%2 == 1 {
			spin = -1.0
		}
		isWarm := (i%3 == 0) || (spin > 0 && i%2 == 0)

		allVortices = append(allVortices, jupiterVortex{
			x:           vx,
			y:           vy,
			rx:          rx,
			ry:          ry,
			cosA:        math.Cos(vAng),
			sinA:        math.Sin(vAng),
			spin:        spin,
			strength:    1.05 + 0.55*rng.Float64(),
			isPrimary:   false,
			isWarmStorm: isWarm,
			phaseOffset: vPhase,
		})
	}

	// 1. Construct divergence-free + swirling 2D velocity field VectorField2D
	fieldFn := func(x, y float64) render.Point {
		nx, ny := x/fw, y/fh
		_, uLat, _, dMeanderDuLon := jovianCoords(x, y)

		// Alternating zonal shear jet streams across latitude bands with meandering Rossby wave tangent
		bandArg := (uLat + 0.5) * nBands * math.Pi
		zonalSpeed := 0.85 * math.Cos(bandArg)

		// Tangent vector of the meandering latitude line in Jovian (uLon, rawLat) space is (1, dMeanderDuLon)
		tLon := 1.0
		tLat := dMeanderDuLon
		tNorm := math.Hypot(tLon, tLat)
		if tNorm > 1e-9 {
			tLon /= tNorm
			tLat /= tNorm
		}
		// Rotate from Jovian frame back to screen coordinates
		vx := zonalSpeed * (tLon*cosTilt + tLat*sinTilt)
		vy := zonalSpeed * (-tLon*sinTilt + tLat*cosTilt)

		// Multi-octave curl noise perturbation from common.NewNoiseEngine(seed)
		c1 := noise.Curl2D(nx*2.6+0.7+loopX, ny*2.6-0.5+loopY, 0.008)
		c2 := noise.Curl2D(nx*6.2-1.3-loopX*0.8, ny*6.2+1.1+loopY*0.8, 0.008)
		vx += 0.36*c1.X + 0.16*c2.X
		vy += 0.36*c1.Y + 0.16*c2.Y

		// Superpose primary super-vortex + Kármán vortex street ovals
		for i := range allVortices {
			v := &allVortices[i]
			dx := x - v.x
			dy := y - v.y
			// Quick bounding-box reject outside 2.6 * rx
			maxR := 2.6 * v.rx
			if math.Abs(dx) > maxR || math.Abs(dy) > maxR {
				continue
			}
			// Transform into elliptical vortex coordinates
			lx := (dx*v.cosA + dy*v.sinA) / v.rx
			ly := (-dx*v.sinA + dy*v.cosA) / v.ry
			r2 := lx*lx + ly*ly
			if r2 > 6.25 || r2 < 1e-9 {
				continue
			}
			r := math.Sqrt(r2)

			// Lamb-Oseen / Jovian collar tangential swirl profile peaking near r ~ 0.72
			decay := math.Exp(-r2 * 1.18)
			swirlMag := v.strength * (r / (0.18 + r2)) * decay

			// Elliptical tangent in local frame is (-spin*ly, spin*lx) + subtle spiral filament entrainment
			inflow := -0.14 * math.Sin(r*math.Pi*1.5-twoPiPhase)
			tlx := (-v.spin*ly + inflow*lx) / r
			tly := (v.spin*lx + inflow*ly) / r

			// Map local velocity back to screen space (scaled by semi-axes to preserve divergence-free elliptical streamlines)
			svx := (tlx*v.rx*v.cosA - tly*v.ry*v.sinA) / v.rx
			svy := (tlx*v.rx*v.sinA + tly*v.ry*v.cosA) / v.rx

			// Suppress background zonal flow inside the primary vortex core so its oval collar stays pristine
			if v.isPrimary && r < 1.15 {
				shield := common.Smoothstep(0.15, 1.15, r)
				vx *= shield
				vy *= shield
			}

			vx += swirlMag * svx
			vy += swirlMag * svy
		}

		return render.Pt(vx, vy)
	}

	// 2. Compute Fast Line Integral Convolution (LIC / OFLIC) scalar field
	licCfg := render.DefaultLICConfig(jm.licLength * scale)
	licCfg.StreamLength = jm.licLength * scale
	licCfg.Contrast = 1.55
	licCfg.NoiseScale = 0.46
	if isAnim {
		licCfg.AnimatedFreq = 0.38
		licCfg.AnimatedPhase = phase
	}
	licField := render.ComputeLICField(w, h, licCfg, fieldFn, seed)

	// 3. Shade every pixel in parallel into rc = c.NewRenderCanvas()
	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
	var wg sync.WaitGroup
	chunk := (h + workers - 1) / workers

	for wk := 0; wk < workers; wk++ {
		y0 := wk * chunk
		y1 := y0 + chunk
		if y0 >= h {
			break
		}
		if y1 > h {
			y1 = h
		}
		wg.Add(1)
		go func(startY, endY int) {
			defer wg.Done()
			for y := startY; y < endY; y++ {
				fy := float64(y) + 0.5
				rowOff := y * w
				for x := 0; x < w; x++ {
					fx := float64(x) + 0.5
					nx, ny := fx/fw, fy/fh
					licVal := 0.5
					if len(licField) == w*h {
						licVal = licField[rowOff+x]
					}

					// Domain-warped multi-octave FBM & ridged turbulence for folded filamentary clouds
					warpX := noise.FBM2D(nx*3.0+1.1+loopX, ny*3.0-0.8+loopY, 3, 2.0, 0.5)
					warpY := noise.FBM2D(nx*3.0-2.4-loopX, ny*3.0+1.9+loopY, 3, 2.0, 0.5)
					fbmDetail := noise.FBM2D(nx*6.5+warpX*0.45+loopX, ny*6.5+warpY*0.45+loopY, 4, 2.05, 0.5)
					ridged := noise.Ridged2D(nx*5.2+warpX*0.35-loopX, ny*5.2+warpY*0.35+loopY, 3, 2.1, 0.5)

					uLon, uLat, _, _ := jovianCoords(fx+warpX*16.0*scale, fy+warpY*16.0*scale)

					// Zonal band scalar index & polar darkening gradient
					bandPhase := (uLat + 0.5) * nBands * math.Pi
					bandWave := math.Sin(bandPhase) // [-1, 1]: -1 = deep cobalt/teal belt, +1 = alabaster/amber/sienna zone
					shearZone := 1.0 - math.Abs(bandWave)
					polarFactor := common.Smoothstep(0.18, 0.52, math.Abs(uLat))

					// Macro temperature scalar in [0, 1] mapped through the Juno palette
					macroT := common.Clamp01(0.46 + 0.34*bandWave + 0.14*fbmDetail - 0.24*polarFactor + 0.08*(ridged-0.5))
					baseCol := macroPalette.SampleOklab(macroT)

					// Enrich high-latitude polar regions with deep Prussian/cobalt blues and turquoise festoons
					if polarFactor > 0.01 {
						festoon := common.Smoothstep(0.48, 0.85, ridged+0.25*math.Sin(uLon*18.0+warpX*2.5))
						polarCol := render.LerpOklab(abyssalPrussian, cobaltMidnight, 0.45+0.55*warpY)
						polarCol = render.LerpOklab(polarCol, turquoiseTeal, festoon*0.72)
						baseCol = render.LerpOklab(baseCol, polarCol, polarFactor*0.62)
					}

					// Evaluate vortex proximity, spiral arms, eyewall rings, and thin-film iridescence weight
					iridescenceWeight := 0.0
					iridescenceCosTheta := 0.65
					iridescenceThickness := 380.0
					eyewallGlow := 0.0

					for i := range allVortices {
						v := &allVortices[i]
						dx := fx - v.x
						dy := fy - v.y
						maxR := 1.85 * v.rx
						if math.Abs(dx) > maxR || math.Abs(dy) > maxR {
							continue
						}
						lx := (dx*v.cosA + dy*v.sinA) / v.rx
						ly := (-dx*v.sinA + dy*v.cosA) / v.ry
						r := math.Hypot(lx, ly)
						if r >= 1.75 {
							continue
						}

						theta := math.Atan2(ly, lx)
						// Logarithmic/Archimedean spiral arms winding into the vortex core
						spiralPhase := 2.0*theta - v.spin*(r*4.8+twoPiPhase) + v.phaseOffset + warpX*1.4
						spiralArm := 0.5 + 0.5*math.Sin(spiralPhase)
						spiralFine := 0.5 + 0.5*math.Cos(3.0*theta-v.spin*(r*7.5+twoPiPhase*1.0)+fbmDetail*2.2)

						// Eyewall collar ring peaks sharply around r in [0.58, 0.86]
						ringDist := (r - 0.72) / 0.22
						eyewall := math.Exp(-ringDist * ringDist * 2.2)
						coreMask := common.Smoothstep(0.55, 0.0, r)
						outerBlend := common.Smoothstep(1.72, 0.25, r)

						if v.isPrimary {
							// Great Red Spot: crimson-sienna inner eye, terracotta spiral bands, luminous amber-gold & alabaster eyewall collar
							spotCol := render.LerpOklab(burntSienna, terracottaStorm, 0.35+0.65*spiralArm)
							spotCol = render.LerpOklab(spotCol, crimsonCore, coreMask*(0.65+0.35*(1.0-spiralFine)))
							collarCol := render.LerpOklab(amberGold, creamyAlabaster, 0.45*spiralFine+0.35*licVal)
							spotCol = render.LerpOklab(spotCol, collarCol, eyewall*(0.68+0.32*spiralArm))
							// Dark cobalt-teal peripheral wake trench around r ~ 1.08
							wakeRing := math.Exp(-math.Pow((r-1.08)/0.18, 2.0))
							spotCol = render.LerpOklab(spotCol, cobaltMidnight, wakeRing*0.58)

							baseCol = render.LerpOklab(baseCol, spotCol, outerBlend*0.94)
							eyewallGlow = math.Max(eyewallGlow, eyewall*0.85)
							if eyewall > 0.08 {
								wIr := eyewall * (0.55 + 0.45*spiralFine)
								if wIr > iridescenceWeight {
									iridescenceWeight = wIr
									iridescenceCosTheta = common.Clamp01(0.35 + 0.55*math.Abs(r-0.72)/0.25 + 0.15*licVal)
									iridescenceThickness = 340.0 + 210.0*r + 75.0*spiralArm + 45.0*licVal
								}
							}
						} else {
							// Satellite Kármán vortex ovals ("String of Pearls" & warm terracotta storms)
							var vCol color.RGBA
							if v.isWarmStorm {
								vCol = render.LerpOklab(burntSienna, terracottaStorm, spiralArm)
								vCol = render.LerpOklab(vCol, amberGold, eyewall*(0.62+0.38*spiralFine))
								vCol = render.LerpOklab(vCol, gildedChampagne, coreMask*0.48)
							} else {
								vCol = render.LerpOklab(cobaltMidnight, turquoiseTeal, 0.30+0.70*spiralArm)
								vCol = render.LerpOklab(vCol, creamyAlabaster, eyewall*(0.72+0.28*spiralFine)+coreMask*0.65)
							}
							baseCol = render.LerpOklab(baseCol, vCol, outerBlend*0.86)
							eyewallGlow = math.Max(eyewallGlow, eyewall*0.65)

							if eyewall > 0.10 {
								wIr := eyewall * 0.82 * (0.5 + 0.5*spiralArm)
								if wIr > iridescenceWeight {
									iridescenceWeight = wIr
									iridescenceCosTheta = common.Clamp01(0.40 + 0.50*math.Abs(r-0.72)/0.25 + 0.15*licVal)
									iridescenceThickness = 310.0 + 240.0*r + 80.0*spiralFine + 40.0*licVal
								}
							}
						}
					}

					// Auroral & high-shear folded filamentary ribbons also carry subtle thin-film iridescence
					auroralRidge := common.Smoothstep(0.66, 0.92, ridged) * (0.45 + 0.55*shearZone)
					if auroralRidge > 0.05 {
						irWeight := auroralRidge * 0.62
						if irWeight > iridescenceWeight {
							iridescenceWeight = irWeight
							iridescenceCosTheta = common.Clamp01(0.45 + 0.45*licVal)
							iridescenceThickness = 290.0 + 260.0*ridged + 65.0*licVal
						}
					}

					// Combine macro Oklab color with the high-frequency fibrous LIC scalar field
					lab := render.RGBAToOklab(baseCol)
					// Contrast-shaped LIC modulation: crisp dark troughs and silky luminous cloud crests
					licCentered := licVal - 0.5
					licMod := 1.0 + 0.46*licCentered + 0.12*eyewallGlow*(licVal-0.35)
					lab.L = common.Constrain(lab.L*licMod+0.06*licCentered, 0.03, 0.97)
					// Boost chroma slightly along sunlit LIC crests
					chromaBoost := 1.0 + 0.22*licCentered
					lab.A *= chromaBoost
					lab.B *= chromaBoost

					rLin, gLin, bLin := render.OklabToLinearRGB(lab)
					rLin = common.Constrain(rLin, 0.0, 1.0)
					gLin = common.Constrain(gLin, 0.0, 1.0)
					bLin = common.Constrain(bLin, 0.0, 1.0)

					// Blend physical Thin-Film Iridescence (render.SampleThinFilmLinearRGB) along eyewall rings & auroral filaments
					if iridescenceWeight > 0.01 {
						tfR, tfG, tfB := render.SampleThinFilmLinearRGB(iridescenceCosTheta, iridescenceThickness, 1.38)
						// Modulate interference sheen by LIC fiber crest so individual storm filaments shimmer
						sheen := common.Clamp01(iridescenceWeight * (0.32 + 0.68*licVal) * 0.44)
						rLin = rLin*(1.0-sheen) + tfR*sheen*1.12
						gLin = gLin*(1.0-sheen) + tfG*sheen*1.12
						bLin = bLin*(1.0-sheen) + tfB*sheen*1.15
					}

					rc.SetPixelLinear(x, y, rLin, gLin, bLin, 1.0)
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	// 4. Overlay crisp, anti-aliased Jobard-Lefer streamlines stroked as tapered gilded/ivory magnetic field filaments
	slCfg := render.DefaultStreamlineConfig(fw, fh, 19.0*scale)
	slCfg.Margin = 10.0 * scale
	slCfg.DSep = 19.0 * scale
	slCfg.DTest = 9.8 * scale
	slCfg.StepSize = 2.4 * scale
	slCfg.MaxSteps = 135
	slCfg.MinPoints = 10

	streamlines := render.TraceEvenStreamlines(slCfg, fieldFn)

	rc.SetBlendMode(render.BlendScreen)
	for idx, line := range streamlines {
		if len(line) < 6 {
			continue
		}
		// Subsample vertices for smooth Catmull-Rom spline fitting
		var ctrl []render.Point
		step := 3
		if len(line) < 15 {
			step = 2
		}
		for i := 0; i < len(line); i += step {
			ctrl = append(ctrl, line[i])
		}
		if ctrl[len(ctrl)-1].Dist(line[len(line)-1]) > 1.0 {
			ctrl = append(ctrl, line[len(line)-1])
		}
		if len(ctrl) < 3 {
			continue
		}

		// Sample midpoint to determine whether this filament threads a warm vortex collar or cool auroral belt
		midPt := ctrl[len(ctrl)/2]
		_, midLat, _, _ := jovianCoords(midPt.X, midPt.Y)
		isAuroral := math.Abs(midLat) > 0.28 || idx%4 == 0

		strokeW := (1.15 + 0.75*math.Sin(float64(idx)*1.7+1.1)*0.5 + 0.35) * scale
		alpha := 0.42 + 0.22*(0.5+0.5*math.Cos(float64(idx)*2.3))

		var cStart, cMid, cEnd color.RGBA
		if isAuroral {
			cStart = turquoiseTeal
			cMid = auroralCyan
			cEnd = creamyAlabaster
		} else if idx%2 == 0 {
			cStart = terracottaStorm
			cMid = gildedChampagne
			cEnd = amberGold
		} else {
			cStart = amberGold
			cMid = creamyAlabaster
			cEnd = warmIvory
		}

		colorProf := func(t float64) color.RGBA {
			if t < 0.5 {
				return render.LerpOklab(cStart, cMid, t*2.0)
			}
			return render.LerpOklab(cMid, cEnd, (t-0.5)*2.0)
		}

		rc.DrawTaperedSpline(ctrl, strokeW, render.TaperBothEnds, colorProf, alpha)
	}

	// Glowing auroral pinpoint pearls at vortex cores & along eyewall collars
	for i := range allVortices {
		v := &allVortices[i]
		coreCol := gildedChampagne
		haloCol := amberGold
		if !v.isWarmStorm && !v.isPrimary {
			coreCol = creamyAlabaster
			haloCol = auroralCyan
		}

		// Soft atmospheric eyewall halo + luminous pinpoint pearl at vortex center
		ringRad := 0.70 * math.Sqrt(v.rx*v.ry)
		rc.DrawSoftRing(v.x, v.y, ringRad, 2.6*scale, haloCol, 0.28)
		rc.DrawGlowDot(v.x, v.y, 7.5*scale, haloCol, 0.42, 0.15)
		rc.DrawGlowDot(v.x, v.y, 2.8*scale, coreCol, 0.85, 0.65)

		// Satellite auroral pearls orbiting the eyewall collar
		pearlCount := 3
		if v.isPrimary {
			pearlCount = 7
		}
		for p := 0; p < pearlCount; p++ {
			ang := v.phaseOffset + float64(p)*(2.0*math.Pi/float64(pearlCount)) + v.spin*twoPiPhase
			lx := 0.72 * math.Cos(ang) * v.rx
			ly := 0.72 * math.Sin(ang) * v.ry
			px := v.x + lx*v.cosA - ly*v.sinA
			py := v.y + lx*v.sinA + ly*v.cosA
			pRad := (1.8 + 0.7*float64(p%2)) * scale
			if v.isPrimary {
				pRad *= 1.25
			}
			rc.DrawGlowDot(px, py, pRad*2.4, haloCol, 0.36, 0.1)
			rc.DrawGlowDot(px, py, pRad, coreCol, 0.78, 0.6)
		}
	}
	rc.SetBlendMode(render.BlendOver)

	// 5. Composable HDR post-processing on rc
	rc.ApplyLuminanceRelief(0.52, -135.0, 42.0, 0.35)
	rc.ApplyBloom(0.75, 4.0*scale, 0.18)
	rc.ApplyVignette(0.22)

	c.MergeRenderCanvas(rc)
}
