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

type kuroshioBioluminescence struct {
	eddies      int
	streamlines int
	licLength   float64
}

// NewKuroshioBioluminescence returns a kuroshioBioluminescence engine that synthesizes a 64-bit HDR
// 3D Translucent Bioluminescent Deep-Sea Medusa (Pelagia noctiluca) swimming through the Kuroshio
// abyss with a ribbed thin-film umbrella bell, frilled helical oral arms, and trailing nematocyst filaments.
func NewKuroshioBioluminescence(eddies, streamlines int, licLength float64) *kuroshioBioluminescence {
	if eddies <= 0 {
		eddies = 14
	}
	if streamlines <= 0 {
		streamlines = 18
	}
	if licLength <= 0 {
		licLength = 24.0
	}
	return &kuroshioBioluminescence{
		eddies:      common.ConstrainInt(eddies, 8, 24),
		streamlines: common.ConstrainInt(streamlines, 10, 32),
		licLength:   common.Constrain(licLength, 10.0, 42.0),
	}
}

// Generative renders a static KuroshioBioluminescence composition.
func (kb *kuroshioBioluminescence) Generative(c *generativeart.Canva) {
	kb.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless looping animation over phase in [0, 1).
func (kb *kuroshioBioluminescence) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
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
		seed = 20262602
	}
	rng := rand.New(rand.NewSource(seed))
	ne := common.NewNoiseEngine(seed + 2602)

	phi := 2.0 * math.Pi * phase
	pulse := math.Sin(phi)

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	// Palette blended with c.Opts().ColorSchema()
	abyssMidnight := color.RGBA{R: 3, G: 7, B: 20, A: 255}
	trenchSapphire := color.RGBA{R: 8, G: 24, B: 62, A: 255}
	electricCyan := color.RGBA{R: 20, G: 240, B: 232, A: 255}
	iceAqua := color.RGBA{R: 185, G: 254, B: 255, A: 255}
	neonMagenta := color.RGBA{R: 246, G: 48, B: 142, A: 255}
	orchidRose := color.RGBA{R: 255, G: 136, B: 196, A: 255}
	solarAmber := color.RGBA{R: 255, G: 188, B: 48, A: 255}
	warmCoral := color.RGBA{R: 250, G: 96, B: 52, A: 255}
	royalViolet := color.RGBA{R: 96, G: 36, B: 196, A: 255}

	schema := c.Opts().ColorSchema()
	if len(schema) > 0 {
		uGrad := render.NewMultiStopGradient(schema...)
		electricCyan = render.LerpOklab(electricCyan, uGrad.SampleOklab(0.25), 0.18)
		neonMagenta = render.LerpOklab(neonMagenta, uGrad.SampleOklab(0.60), 0.18)
		solarAmber = render.LerpOklab(solarAmber, uGrad.SampleOklab(0.85), 0.18)
	}

	// Bell geometry: positioned in upper-right quadrant (0.61*fw, 0.35*fh), tilted at -38 deg (swimming toward NE)
	bellCX := 0.60 * fw
	bellCY := 0.36 * fh
	swimAngle := -0.66 // direction of bell apex (toward upper-right)
	cosSwim := math.Cos(swimAngle)
	sinSwim := math.Sin(swimAngle)
	// Unit axis pointing from apex toward trailing tentacles (lower-left)
	trailDirX := -cosSwim
	trailDirY := -sinSwim
	// Unit lateral axis perpendicular to swimming direction
	latDirX := -sinSwim
	latDirY := cosSwim

	bellRad := (148.0 + 6.0*pulse) * scale
	bellHeight := (118.0 - 5.0*pulse) * scale
	numLobes := kb.eddies

	// 1. Parallel Per-Pixel Pass: Deep Kuroshio Abyss + Volumetric Caustic Veil + 3D Translucent Medusa Bell
	abyssR, abyssG, abyssB, _ := render.ColorToLinearRGB(abyssMidnight)
	sapR, sapG, sapB, _ := render.ColorToLinearRGB(trenchSapphire)
	cyaR, cyaG, cyaB, _ := render.ColorToLinearRGB(electricCyan)
	magR, magG, magB, _ := render.ColorToLinearRGB(neonMagenta)
	ambR, ambG, ambB, _ := render.ColorToLinearRGB(solarAmber)
	ivoR, ivoG, ivoB, _ := render.ColorToLinearRGB(iceAqua)

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
				for x := 0; x < w; x++ {
					fx := float64(x) + 0.5
					nx, ny := fx/fw, fy/fh

					// Transform pixel into local Medusa Bell frame (uLat across width, vAxial along swim axis)
					dx := fx - bellCX
					dy := fy - bellCY
					uLat := (dx*latDirX + dy*latDirY) / bellRad
					vAxial := (dx*cosSwim + dy*sinSwim) / bellHeight

					// Deep abyssal background + diagonal Kuroshio current veil
					neb := 0.5 + 0.5*ne.FBM2D(nx*2.4+0.7, ny*2.4-0.5, 3, 2.0, 0.5)
					streamWave := 0.5 + 0.5*math.Sin((uLat*3.2+vAxial*1.4)*math.Pi+neb*2.2-phi)
					wakeEnv := math.Exp(-uLat*uLat*1.35) * common.Smoothstep(1.2, -2.2, vAxial)

					rLin := abyssR + (sapR-abyssR)*(0.35*neb+0.55*wakeEnv)
					gLin := abyssG + (sapG-abyssG)*(0.35*neb+0.55*wakeEnv)
					bLin := abyssB + (sapB-abyssB)*(0.35*neb+0.55*wakeEnv)

					// Subtle bioluminescent wake glow behind the bell
					wakeGlow := wakeEnv * math.Pow(streamWave, 3.0) * 0.14
					rLin += (0.5*cyaR + 0.5*magR) * wakeGlow
					gLin += (0.7*cyaG + 0.3*magG) * wakeGlow
					bLin += (0.7*cyaB + 0.3*magB) * wakeGlow

					// 3D Translucent Ribbed Umbrella Bell Shader
					// In local frame: apex is at vAxial = +1.0, scalloped rim is near vAxial = -0.22 +skirtWave
					if vAxial > -0.42 && vAxial < 1.08 && math.Abs(uLat) < 1.18 {
						// Normalize axial height hBell in [0 (apex) .. 1 (margin rim)]
						hBell := (0.96 - vAxial) / 1.18
						if hBell > 0.0 && hBell < 1.12 {
							// Bell profile radius R(hBell): hemispherical near apex, flaring slightly at the scalloped skirt
							profileR := math.Sqrt(math.Max(0.0, 1.0-(1.0-hBell)*(1.0-hBell)))
							if hBell > 0.78 {
								flare := (hBell - 0.78) / 0.34
								profileR += 0.08 * math.Sin(flare*math.Pi)
							}

							if profileR > 1e-4 {
								uCyl := uLat / profileR // [-1, 1] across the projected bell width at this axial slice
								if math.Abs(uCyl) < 1.04 {
									// Azimuthal angle around the 3D bell axis in [-pi/2, +pi/2] on front hemisphere
									clampedU := common.Constrain(uCyl, -0.999, 0.999)
									azim := math.Asin(clampedU)
									nzFront := math.Sqrt(math.Max(0.02, 1.0-clampedU*clampedU))

									// Scalloped lappet rim modulation along the lower bell margin
									lobeWave := math.Cos(float64(numLobes) * azim)
									rimLimit := 0.96 + 0.055*lobeWave + 0.018*math.Sin(float64(numLobes*2)*azim-phi)

									if hBell <= rimLimit+0.03 {
										rimCov := common.Smoothstep(rimLimit+0.03, rimLimit-0.02, hBell)
										latCov := common.Smoothstep(1.03, 0.96, math.Abs(uCyl))
										bellCov := rimCov * latCov

										// 3D normal on the ribbed umbrella dome
										ribSlope := -0.22 * math.Sin(float64(numLobes)*azim)
										nxLocal := clampedU + ribSlope*nzFront
										nyLocal := (1.0 - hBell) * 0.68
										nzLocal := nzFront * (1.0 - 0.25*math.Abs(1.0-hBell))
										nLen := math.Sqrt(nxLocal*nxLocal + nyLocal*nyLocal + nzLocal*nzLocal)
										nxLocal, nyLocal, nzLocal = nxLocal/nLen, nyLocal/nLen, nzLocal/nLen

										// Schlick-Fresnel glass rim glow + Airy thin-film structural iridescence
										fresnel := math.Pow(1.0-nzLocal, 2.2)
										filmNm := 320.0 + 195.0*(0.5+0.5*math.Sin(hBell*4.5+lobeWave*1.2-phi))
										tfR, tfG, tfB := render.SampleThinFilmLinearRGB(nzLocal, filmNm, 1.48)

										// Radial gastrovascular canals (sharp glowing ribs where lobeWave -> +1)
										canalGlow := math.Pow(0.5+0.5*lobeWave, 8.0)
										// Scalloped margin rhopalia glow right along hBell ~ rimLimit
										marginGlow := math.Exp(-math.Pow((hBell-rimLimit)/0.045, 2.0))

										// Inner warm amber-magenta manubrium core visible through the translucent dome
										coreDist := math.Hypot(uLat*1.35, (hBell-0.52)*1.6)
										innerCore := math.Exp(-coreDist * coreDist * 2.8)

										// Specular highlight on the upper-left dome curvature
										spec := math.Pow(math.Max(0.0, -0.45*nxLocal+0.52*nyLocal+0.72*nzLocal), 38.0)

										// Synthesize front + back translucent jelly shell radiance
										jellyR := (0.28*cyaR+0.72*tfR)*fresnel*1.35 +
											(0.65*ivoR+0.35*cyaR)*canalGlow*0.72 +
											(0.75*ambR+0.25*magR)*innerCore*0.85 +
											(0.55*magR+0.45*ivoR)*marginGlow*0.92 +
											ivoR*spec*0.95
										jellyG := (0.35*cyaG+0.65*tfG)*fresnel*1.35 +
											(0.65*ivoG+0.35*cyaG)*canalGlow*0.72 +
											(0.60*ambG+0.40*magG)*innerCore*0.75 +
											(0.45*cyaG+0.55*ivoG)*marginGlow*0.92 +
											ivoG*spec*0.95
										jellyB := (0.35*cyaB+0.65*tfB)*fresnel*1.35 +
											(0.65*ivoB+0.35*cyaB)*canalGlow*0.75 +
											(0.40*ambB+0.60*magB)*innerCore*0.70 +
											(0.55*magB+0.45*ivoB)*marginGlow*0.95 +
											ivoB*spec*0.95

										// Translucent additive/screen composite over the abyssal background
										rLin = rLin*(1.0-0.45*bellCov) + jellyR*bellCov
										gLin = gLin*(1.0-0.45*bellCov) + jellyG*bellCov
										bLin = bLin*(1.0-0.45*bellCov) + jellyB*bellCov
									}
								}
							}
						}

						// Concave subumbrella opening shadow & glowing inner ring at the base of the bell (vAxial ~ -0.22)
						subDist := math.Hypot(uLat*1.02, (vAxial+0.22)*3.8)
						if subDist < 1.0 {
							subRing := math.Exp(-math.Pow((subDist-0.82)/0.14, 2.0))
							rLin += (0.6*cyaR + 0.4*magR) * subRing * 0.45
							gLin += (0.7*cyaG + 0.3*magG) * subRing * 0.45
							bLin += (0.7*cyaB + 0.3*magB) * subRing * 0.50
						}
					}

					rc.SetPixelLinear(x, y, rLin, gLin, bLin, 1.0)
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	// 2. Trailing Fine Nematocyst Silk Tentacles (48+ long catenary/vortex threads in BlendScreen)
	rc.SetBlendMode(render.BlendScreen)
	numTentacles := kb.streamlines * 3
	const tentSteps = 44

	for tIdx := 0; tIdx < numTentacles; tIdx++ {
		ft := float64(tIdx) / float64(numTentacles-1)
		latStart := (ft*2.0 - 1.0) * 0.88 * bellRad
		// Start along the scalloped subumbrella margin
		startX := bellCX + cosSwim*(-0.20*bellHeight) + latDirX*latStart
		startY := bellCY + sinSwim*(-0.20*bellHeight) + latDirY*latStart

		tentLen := (240.0 + 170.0*math.Sin(ft*math.Pi) + 65.0*rng.Float64()) * scale * (kb.licLength / 24.0)
		waveFreq := 2.2 + 1.6*rng.Float64()
		wavePhase := rng.Float64()*2.0*math.Pi - phi

		pts := make([]render.Point, tentSteps+1)
		for k := 0; k <= tentSteps; k++ {
			u := float64(k) / float64(tentSteps)
			sDist := u * tentLen
			// Graceful S-curve Kármán vortex street deflection in the wake
			sway := (18.0*u + 26.0*u*u) * scale * math.Sin(u*math.Pi*waveFreq+wavePhase+ft*3.5)
			spread := latStart * (1.0 + 0.35*u)

			px := startX + trailDirX*sDist + latDirX*(spread-latStart+sway)
			py := startY + trailDirY*sDist + latDirY*(spread-latStart+sway)
			pts[k] = render.Pt(px, py)
		}

		var cStart, cMid, cEnd color.RGBA
		switch tIdx % 3 {
		case 0:
			cStart, cMid, cEnd = iceAqua, electricCyan, royalViolet
		case 1:
			cStart, cMid, cEnd = solarAmber, neonMagenta, royalViolet
		default:
			cStart, cMid, cEnd = orchidRose, electricCyan, trenchSapphire
		}

		strokeW := (0.85 + 0.95*math.Sin(ft*math.Pi)) * scale
		alpha := 0.38 + 0.34*math.Sin(ft*math.Pi)
		rc.DrawTaperedSpline(pts, strokeW, render.TaperBothEnds, func(t float64) color.RGBA {
			if t < 0.45 {
				return render.LerpOklab(cStart, cMid, t/0.45)
			}
			return render.LerpOklab(cMid, cEnd, (t-0.45)/0.55)
		}, alpha)

		// Glowing nematocyst bead clusters along select tentacles
		if tIdx%2 == 0 {
			for k := 4; k < tentSteps; k += 6 {
				u := float64(k) / float64(tentSteps)
				pt := pts[k]
				bRad := (1.8 - 1.0*u) * scale
				rc.DrawGlowDot(pt.X, pt.Y, bRad*2.2, cMid, 0.36*(1.0-u), 0.20)
				rc.DrawGlowDot(pt.X, pt.Y, bRad*0.85, iceAqua, 0.78*(1.0-u), 0.75)
			}
		}
	}

	// 3. Four Sculpted 3D Helical Frilled Oral Arms (Multi-Strand Organza Ribbon Bundles)
	const numArms = 4
	const armStrands = 14
	const armSteps = 64

	for aIdx := 0; aIdx < numArms; aIdx++ {
		fa := (float64(aIdx)/float64(numArms-1))*2.0 - 1.0 // [-1, 1]
		baseLat := fa * 0.38 * bellRad
		armLen := (265.0 + 35.0*math.Cos(fa*1.2)) * scale

		for sIdx := 0; sIdx < armStrands; sIdx++ {
			fs := (float64(sIdx)/float64(armStrands-1))*2.0 - 1.0 // [-1, 1] across the frilled arm width
			pts := make([]render.Point, armSteps+1)

			for k := 0; k <= armSteps; k++ {
				u := float64(k) / float64(armSteps)
				sDist := u * armLen
				taper := math.Sin(math.Pi * common.Constrain(u*0.92+0.05, 0.02, 0.98))

				// Helical twist of the oral arm ribbon + ruffled edge frill
				helixAng := u*math.Pi*3.6 + float64(aIdx)*1.4 - phi
				armSway := (14.0*u + 24.0*u*u) * scale * math.Sin(u*math.Pi*2.4+float64(aIdx)*1.1-phi)
				frillOffset := fs * (18.0 * scale * taper) * math.Cos(helixAng)
				ruffle := 4.5 * scale * taper * math.Sin(u*math.Pi*14.0+fs*2.5-phi)

				latPos := baseLat*(1.0+0.25*u) + armSway + frillOffset + ruffle
				axPos := (-0.12 * bellHeight) + sDist

				px := bellCX + trailDirX*axPos + latDirX*latPos
				py := bellCY + trailDirY*axPos + latDirY*latPos
				pts[k] = render.Pt(px, py)
			}

			isSelvedge := sIdx == 0 || sIdx == armStrands-1
			wgt := 0.24
			strokeW := 1.15 * scale
			if isSelvedge {
				wgt = 0.62
				strokeW = 1.65 * scale
			}

			armWarm := solarAmber
			armMid := neonMagenta
			armTip := electricCyan
			if aIdx%2 == 1 {
				armWarm = orchidRose
				armMid = warmCoral
				armTip = iceAqua
			}

			rc.DrawTaperedSpline(pts, strokeW, render.TaperBothEnds, func(t float64) color.RGBA {
				var col color.RGBA
				if t < 0.5 {
					col = render.LerpOklab(armWarm, armMid, t*2.0)
				} else {
					col = render.LerpOklab(armMid, armTip, (t-0.5)*2.0)
				}
				if isSelvedge {
					col = render.LerpOklab(col, iceAqua, 0.45)
				}
				return col
			}, wgt*(1.0-0.25*math.Abs(fs)))
		}
	}

	// 4. Drifting Bioluminescent Noctiluca Plankton & Vortex Bokeh in the Surrounding Kuroshio Abyss
	rc.SetBlendMode(render.BlendAdd)
	for i := 0; i < 68; i++ {
		fi := float64(i)
		px := math.Mod(fi*0.754877+0.13*rng.Float64(), 1.0) * fw
		py := math.Mod(fi*0.569840+0.13*rng.Float64(), 1.0) * fh
		pRad := (1.2 + 2.8*math.Pow(rng.Float64(), 2.0)) * scale
		pCol := electricCyan
		if i%3 == 1 {
			pCol = neonMagenta
		} else if i%4 == 0 {
			pCol = solarAmber
		}
		rc.DrawGlowDot(px, py, pRad*2.6, pCol, 0.34, 0.18)
		rc.DrawGlowDot(px, py, pRad*0.8, iceAqua, 0.85, 0.78)
	}
	rc.SetBlendMode(render.BlendOver)

	rc.ApplyBloom(0.76, 4.2*scale, 0.20)
	rc.ApplyVignette(0.22)
	c.MergeRenderCanvas(rc)
}
