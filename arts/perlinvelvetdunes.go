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

type perlinVelvetDunes struct {
	strata    int
	grooves   int
	windShear float64
}

// NewPerlinVelvetDunes returns a PerlinVelvetDunes generator that synthesizes a
// full-bleed, horizon-to-horizon 3D wind-combed silk & sandstone dune relief
// driven by gradient-damped Perlin erosion (ErosionFBM2D) and FBM2DDeriv:
// overlapping undulating strata with hundreds of continuous wind-carved bristle
// grooves, deep inter-crest cast shadows, multi-pigment mineral streaks, and
// wind-blown desert stardust flecks.
func NewPerlinVelvetDunes(strata, grooves int, windShear float64) *perlinVelvetDunes {
	if strata <= 0 {
		strata = 24
	}
	if grooves <= 0 {
		grooves = 58
	}
	if windShear <= 0 {
		windShear = 1.35
	}
	return &perlinVelvetDunes{
		strata:    common.ConstrainInt(strata, 12, 42),
		grooves:   common.ConstrainInt(grooves, 28, 96),
		windShear: common.Constrain(windShear, 0.5, 2.5),
	}
}

// Generative renders the static PerlinVelvetDunes composition.
func (pvd *perlinVelvetDunes) Generative(c *generativeart.Canva) {
	pvd.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360° looping PerlinVelvetDunes animation.
func (pvd *perlinVelvetDunes) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	minDim := math.Min(fw, fh)
	scale := minDim / 600.0
	phi := 2.0 * math.Pi * phase

	seed := c.Seed()
	if seed == 0 {
		seed = 20263013
	}
	rng := rand.New(rand.NewSource(seed))
	ne := common.NewNoiseEngine(seed)

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	// =========================================================================
	// 1. POLYCHROME SILK-ROAD DUNHUANG & SAHARA MINERAL PALETTE
	// =========================================================================
	// Sweeps across the 24 overlapping strata from Nocturne Ultramarine & Peacock
	// Turquoise in the upper/mid currents to Saffron Gold, Terracotta Coral,
	// Cinnabar Ruby, and Deep Wine-Burgundy!
	duneGrad := render.NewMultiStopGradient(
		color.RGBA{R: 18, G: 48, B: 122, A: 255},  // Nocturne Lapis Ultramarine
		color.RGBA{R: 22, G: 136, B: 148, A: 255}, // Oasis Kingfisher Teal
		color.RGBA{R: 48, G: 164, B: 124, A: 255}, // Celadon Malachite
		color.RGBA{R: 242, G: 176, B: 52, A: 255}, // Sahara Sunburst Gold
		color.RGBA{R: 232, G: 96, B: 42, A: 255},  // Burnt Tangerine Ochre
		color.RGBA{R: 196, G: 38, B: 46, A: 255},  // Cinnabar Vermilion
		color.RGBA{R: 138, G: 22, B: 58, A: 255},  // Bordeaux Wine-Carmine
		color.RGBA{R: 78, G: 24, B: 96, A: 255},   // Imperial Twilight Plum
		color.RGBA{R: 18, G: 48, B: 122, A: 255},  // Loop back to Lapis
	)

	nPix := w * h
	zBuf := make([]float64, nPix)
	stratumIDBuf := make([]int, nPix)
	nxBuf := make([]float64, nPix)
	nyBuf := make([]float64, nPix)
	nzBuf := make([]float64, nPix)
	albR := make([]float64, nPix)
	albG := make([]float64, nPix)
	albB := make([]float64, nPix)
	grooveBuf := make([]float64, nPix)

	// =========================================================================
	// 2. PRECOMPUTE 24 OVERLAPPING FULL-WIDTH EROSION-PERLIN WAVE STRATA
	// =========================================================================
	numStrata := pvd.strata
	shear := pvd.windShear
	loopX := 0.14 * math.Cos(phi)
	loopY := 0.14 * math.Sin(phi)

	type duneStratum struct {
		id         int
		baseZ      float64
		yCenter    float64
		bandThick  float64
		freqX      float64
		phaseOff   float64
		seedOff    float64
		baseCol    color.RGBA
		streakCol  color.RGBA
		deepCol    color.RGBA
		glintCol   color.RGBA
		ridgeLUT   [512]float64
		slopeLUT   [512]float64
		streakLUT  [512]float64
		threadLUT  [512]float64
	}

	strataList := make([]duneStratum, numStrata)
	gFreq := float64(pvd.grooves) * 0.62

	for i := 0; i < numStrata; i++ {
		fi := float64(i) / float64(numStrata-1)
		sOff := float64(i)*17.3 + rng.Float64()*50.0

		// Progress from top (-0.12) to bottom (1.12) with generous shingle overlap!
		yCen := -0.10 + 1.22*fi
		thick := (0.115 + 0.045*math.Sin(float64(i)*1.9)) * (1.0 + 0.15*shear*0.3)

		hue := math.Mod(fi*1.15+0.08*math.Sin(float64(i)*1.3)+10.0, 1.0)
		bCol := duneGrad.SampleOklab(hue)
		stCol := duneGrad.SampleOklab(math.Mod(hue+0.09, 1.0))
		dpCol := render.LerpOklab(bCol, color.RGBA{R: 10, G: 4, B: 14, A: 255}, 0.80)
		glCol := render.LerpOklab(bCol, color.RGBA{R: 255, G: 246, B: 226, A: 255}, 0.72)

		st := duneStratum{
			id:        i + 1,
			baseZ:     4.0 + float64(i)*3.6,
			yCenter:   yCen,
			bandThick: thick,
			freqX:     1.25 + 0.35*math.Sin(float64(i)*1.1),
			phaseOff:  float64(i)*1.45 + 0.5*rng.Float64(),
			seedOff:   sOff,
			baseCol:   bCol,
			streakCol: stCol,
			deepCol:   dpCol,
			glintCol:  glCol,
		}

		for k := 0; k < 512; k++ {
			v := (float64(k)/511.0)*2.0 - 1.0
			pBand, dBand, _ := ne.Perlin2DDeriv(v*gFreq*0.28+sOff, 1.9)
			pMed, dMed, _ := ne.Perlin2DDeriv(v*gFreq*0.75-sOff*1.3, 4.3)
			pFine, dFine, _ := ne.Perlin2DDeriv(v*gFreq*1.60+sOff*2.1, 8.1)
			wave := math.Sin(v*gFreq*math.Pi + 1.25*pBand)
			dWave := (gFreq*math.Pi + 1.25*gFreq*0.28*dBand) * math.Cos(v*gFreq*math.Pi+1.25*pBand)

			st.ridgeLUT[k] = 0.28*pBand + 0.28*pMed + 0.20*pFine + 0.24*wave
			st.slopeLUT[k] = common.Constrain(
				(0.28*gFreq*0.28*dBand+0.28*gFreq*0.75*dMed+0.20*gFreq*1.60*dFine+0.24*dWave)/(gFreq*1.15),
				-1.45, 1.45,
			)
			st.streakLUT[k] = common.Clamp01(0.65*(0.5+0.5*ne.Perlin2D(v*5.0+sOff, 2.7)) + 0.35*(0.5+0.5*pMed))
			st.threadLUT[k] = math.Pow(0.5+0.5*math.Sin(v*gFreq*math.Pi*0.5+2.0*pBand), 6.0) *
				common.Smoothstep(0.0, 0.55, 0.5+0.5*ne.Perlin2D(v*4.5+sOff, 5.7))
		}
		strataList[i] = st
	}

	// =========================================================================
	// 3. PARALLEL RASTERIZATION OF WIND-COMBED EROSION-PERLIN DUNE STRATA
	// =========================================================================
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
			for py := startY; py < endY; py++ {
				uy := (float64(py) + 0.5) / fh
				for px := 0; px < w; px++ {
					ux := (float64(px) + 0.5) / fw
					idx := py*w + px

					// Evaluate strata from bottom to top (foreground to background) or top to bottom
					for sIdx := 0; sIdx < numStrata; sIdx++ {
						st := &strataList[sIdx]

						// ErosionFBM2D + Perlin wave crest displacement yCrest(ux)
						erodedH, sxE, _ := ne.ErosionFBM2D(ux*st.freqX+st.seedOff*0.1+loopX*0.35, st.yCenter*2.2+loopY*0.35, 3, 2.0, 0.5, 1.4*shear)
						waveS := 0.085*math.Sin(ux*math.Pi*2.2*st.freqX*0.6+st.phaseOff) +
							0.045*math.Cos(ux*math.Pi*3.6-st.phaseOff*0.7)
						yCurve := st.yCenter + (erodedH-0.5)*0.18*shear + waveS

						// Scalloped upper and lower edges of the wind-combed dune ribbon
						edgeScallop := 0.012 * math.Tanh(3.5*ne.Perlin2D(ux*5.5+st.seedOff, 3.1))
						dy := uy - (yCurve + edgeScallop)
						v := dy / st.bandThick // [-1, +1] across the stratum ribbon
						if v < -1.0 || v > 1.0 {
							continue
						}

						// Tangent vector (1, dy/dx) and normal vector (-dy/dx, 1) of the wave curve
						dydx := (0.18*shear*sxE*st.freqX + 0.085*math.Pi*2.2*st.freqX*0.6*math.Cos(ux*math.Pi*2.2*st.freqX*0.6+st.phaseOff))
						invLen := 1.0 / math.Sqrt(1.0+dydx*dydx)
						nx, ny := -dydx*invLen, invLen

						lutPos := (v + 1.0) * 0.5 * 511.0
						k0 := common.ConstrainInt(int(math.Floor(lutPos)), 0, 510)
						frac := lutPos - float64(k0)

						ridgeH := st.ridgeLUT[k0]*(1.0-frac) + st.ridgeLUT[k0+1]*frac
						dRidge := st.slopeLUT[k0]*(1.0-frac) + st.slopeLUT[k0+1]*frac
						streakMix := st.streakLUT[k0]*(1.0-frac) + st.streakLUT[k0+1]*frac
						glintThread := st.threadLUT[k0]*(1.0-frac) + st.threadLUT[k0+1]*frac

						// Longitudinal wind-ripple modulation along ux
						windRipple := ne.Perlin2D(ux*4.5+st.seedOff, v*14.0)
						ridgeH += 0.10 * windRipple

						// 3D Elevation Z: raised crest along the upper overlapping lip (v -> -1)!
						edgeCushion := 1.0 - math.Pow(math.Abs(v), 4.2)
						crestLip := 0.52 * math.Exp(-math.Pow((v+0.88)/0.10, 2.0))
						zSurf := st.baseZ + 2.8*edgeCushion + 1.2*crestLip + 0.88*ridgeH

						if zSurf < zBuf[idx] {
							continue
						}

						edgeBevel := 0.65 * v * math.Pow(math.Abs(v), 2.4)
						transSlope := edgeBevel - 1.05*dRidge
						normX := transSlope * nx
						normY := transSlope * ny
						normZ := 1.0
						invN := 1.0 / math.Sqrt(normX*normX+normY*normY+normZ*normZ)
						normX, normY, normZ = normX*invN, normY*invN, normZ*invN

						colMid := render.LerpOklab(st.baseCol, st.streakCol, math.Pow(streakMix, 1.35)*0.68)
						var pixCol color.RGBA
						if ridgeH < 0 {
							gT := common.Clamp01(-ridgeH * 1.35)
							pixCol = render.LerpOklab(colMid, st.deepCol, gT*0.86)
						} else {
							cT := common.Clamp01(ridgeH * 1.15)
							crestCurve := 0.5 + 0.5*math.Sin(ux*math.Pi*2.5+st.phaseOff)
							glintW := common.Clamp01(cT*0.36 + 0.66*glintThread*(0.4+0.6*crestCurve))
							pixCol = render.LerpOklab(colMid, st.glintCol, glintW)
						}

						rL, gL, bL, _ := render.ColorToLinearRGB(pixCol)
						zBuf[idx] = zSurf
						stratumIDBuf[idx] = st.id
						nxBuf[idx] = normX
						nyBuf[idx] = normY
						nzBuf[idx] = normZ
						albR[idx] = rL
						albG[idx] = gL
						albB[idx] = bL
						grooveBuf[idx] = ridgeH
					}
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	// =========================================================================
	// 4. DEFERRED 3D SHADING: DEEP CONTACT AO + HORIZON SHADOWS + SPECULAR
	// =========================================================================
	lx, ly, lz := -0.48+0.10*math.Cos(phi), -0.60+0.10*math.Sin(phi), 0.64
	lLen := math.Sqrt(lx*lx + ly*ly + lz*lz)
	lx, ly, lz = lx/lLen, ly/lLen, lz/lLen

	hx, hy, hz := lx, ly, lz+1.0
	hLen := math.Sqrt(hx*hx + hy*hy + hz*hz)
	hx, hy, hz = hx/hLen, hy/hLen, hz/hLen

	aoRing := [12][2]int{
		{-2, -2}, {0, -3}, {2, -2}, {-3, 0},
		{3, 0}, {-2, 2}, {0, 3}, {2, 2},
		{-6, -4}, {6, -4}, {-4, 6}, {4, 6},
	}

	var wgShade sync.WaitGroup
	for wk := 0; wk < workers; wk++ {
		y0 := wk * chunk
		y1 := y0 + chunk
		if y0 >= h {
			break
		}
		if y1 > h {
			y1 = h
		}
		wgShade.Add(1)
		go func(startY, endY int) {
			defer wgShade.Done()
			for y := startY; y < endY; y++ {
				uy := float64(y) / fh
				for x := 0; x < w; x++ {
					ux := float64(x) / fw
					idx := y*w + x
					z0 := zBuf[idx]
					id0 := stratumIDBuf[idx]

					contactOcc := 0.0
					for _, off := range aoRing {
						sx := common.ConstrainInt(x+off[0], 0, w-1)
						sy := common.ConstrainInt(y+off[1], 0, h-1)
						sIdx := sy*w + sx
						dz := zBuf[sIdx] - z0
						if dz > 0.25 {
							weight := 0.85
							if stratumIDBuf[sIdx] != id0 {
								weight = 2.45
								dz = math.Min(dz, 4.2)
							}
							contactOcc += weight * common.Constrain(dz/3.8, 0.0, 1.0)
						}
					}
					contactAO := common.Constrain(1.0-(contactOcc/12.0)*0.82, 0.08, 1.0)

					castShadow := 1.0
					for step := 1; step <= 18; step++ {
						dist := float64(step) * scale
						sx := int(math.Round(float64(x) + lx*dist*1.35))
						sy := int(math.Round(float64(y) + ly*dist*1.35))
						if sx < 0 || sx >= w || sy < 0 || sy >= h {
							break
						}
						rawDz := zBuf[sy*w+sx] - z0
						if stratumIDBuf[sy*w+sx] != id0 && rawDz > 0 {
							rawDz = math.Min(rawDz, 4.8)
						}
						dz := rawDz - dist*0.26
						if dz > 0.15 {
							pen := common.Constrain(dz/2.6, 0.0, 0.78)
							if 1.0-pen < castShadow {
								castShadow = 1.0 - pen
							}
						}
					}

					macroLight := 0.28 + 0.84*common.Smoothstep(1.10, -0.05, uy*0.62-ux*0.28)

					nx, ny, nz := nxBuf[idx], nyBuf[idx], nzBuf[idx]
					nDotL := nx*lx + ny*ly + nz*lz
					wrapDiff := common.Constrain((nDotL+0.28)/1.28, 0.0, 1.0)

					grooveAO := 0.45 + 0.55*common.Clamp01(0.5+0.65*grooveBuf[idx])

					nDotH := math.Max(0.0, nx*hx+ny*hy+nz*hz)
					ridgeMask := common.Clamp01(0.25 + 0.75*grooveBuf[idx])
					spec := math.Pow(nDotH, 22.0) * 0.56 * ridgeMask * contactAO * castShadow

					totalShade := (0.10 + 1.02*wrapDiff*castShadow) * contactAO * grooveAO * macroLight

					rOut := albR[idx]*totalShade + spec*1.12
					gOut := albG[idx]*totalShade + spec*0.98
					bOut := albB[idx]*totalShade + spec*0.86

					rOut = rOut / (1.0 + 0.16*rOut)
					gOut = gOut / (1.0 + 0.16*gOut)
					bOut = bOut / (1.0 + 0.16*bOut)

					rc.SetPixelLinear(x, y, common.Clamp01(rOut), common.Clamp01(gOut), common.Clamp01(bOut), 1.0)
				}
			}
		}(y0, y1)
	}
	wgShade.Wait()

	// =========================================================================
	// 5. WIND-BLOWN MINERAL & GOLD-DUST FLECKS
	// =========================================================================
	spatterRng := rand.New(rand.NewSource(seed + 8351))
	darkFleck := color.RGBA{R: 10, G: 5, B: 10, A: 255}
	goldFleck := color.RGBA{R: 255, G: 224, B: 152, A: 255}

	for i := 0; i < 480; i++ {
		fx := spatterRng.Float64() * fw
		fy := spatterRng.Float64() * fh
		rad := (0.72 + 1.80*math.Pow(spatterRng.Float64(), 2.2)) * scale

		col := darkFleck
		alpha := 0.75 + 0.22*spatterRng.Float64()
		if i%9 == 0 {
			col = goldFleck
			alpha = 0.60
			rad *= 0.88
		}

		minX := common.ConstrainInt(int(math.Floor(fx-rad-1)), 0, w-1)
		maxX := common.ConstrainInt(int(math.Ceil(fx+rad+1)), 0, w-1)
		minY := common.ConstrainInt(int(math.Floor(fy-rad-1)), 0, h-1)
		maxY := common.ConstrainInt(int(math.Ceil(fy+rad+1)), 0, h-1)
		for py := minY; py <= maxY; py++ {
			for px := minX; px <= maxX; px++ {
				d := math.Hypot(float64(px)+0.5-fx, float64(py)+0.5-fy)
				if d <= rad+0.6 {
					cov := common.Clamp01(rad + 0.5 - d)
					rc.BlendPixel(px, py, col, alpha*cov)
				}
			}
		}
	}

	c.MergeRenderCanvas(rc)
}
