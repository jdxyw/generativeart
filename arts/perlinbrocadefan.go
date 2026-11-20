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

type perlinBrocadeFan struct {
	plumes    int
	filaments int
	flareCurl float64
}

// NewPerlinBrocadeFan returns a PerlinBrocadeFan generator that synthesizes a
// full-bleed 3D scalloped lotus-fan & gilded silk-brocade canopy driven by
// exact polar Euclidean normal distance and Perlin FBM2DDeriv: 24 staggered
// arched fan-wave strata radiating from a deep horizon hearth with 24K gold
// selvedge rims, inner gold pinstripes, 3-band inlaid polychrome silk-brocade
// cross-sections, fine silk warp filaments, and deep inter-arch cast shadows.
func NewPerlinBrocadeFan(plumes, filaments int, flareCurl float64) *perlinBrocadeFan {
	if plumes <= 0 {
		plumes = 24
	}
	if filaments <= 0 {
		filaments = 56
	}
	if flareCurl <= 0 {
		flareCurl = 1.35
	}
	return &perlinBrocadeFan{
		plumes:    common.ConstrainInt(plumes, 12, 40),
		filaments: common.ConstrainInt(filaments, 28, 96),
		flareCurl: common.Constrain(flareCurl, 0.5, 2.5),
	}
}

// Generative renders the static PerlinBrocadeFan composition.
func (pbf *perlinBrocadeFan) Generative(c *generativeart.Canva) {
	pbf.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360° looping PerlinBrocadeFan animation.
func (pbf *perlinBrocadeFan) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
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
		seed = 20263102
	}
	rng := rand.New(rand.NewSource(seed))
	ne := common.NewNoiseEngine(seed)

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	// =========================================================================
	// 1. IMPERIAL DUNHUANG & ART-DECO PEACOCK-LOTUS BROCADE PALETTE
	// =========================================================================
	fanGrad := render.ResolveCyclicGradient(c.Opts().ColorSchema(), common.Dunhuang,
		color.RGBA{R: 16, G: 44, B: 122, A: 255},  // Nocturne Lapis Sapphire
		color.RGBA{R: 18, G: 134, B: 158, A: 255}, // Kingfisher Cyan-Turquoise
		color.RGBA{R: 38, G: 170, B: 126, A: 255}, // Peacock Viridian Jade
		color.RGBA{R: 248, G: 190, B: 52, A: 255}, // Imperial Sunburst Gold
		color.RGBA{R: 242, G: 102, B: 46, A: 255}, // Fiery Sunset Tangerine-Coral
		color.RGBA{R: 204, G: 36, B: 58, A: 255},  // Alizarin Cinnabar Ruby
		color.RGBA{R: 128, G: 24, B: 84, A: 255},  // Bordeaux Rose-Plum
		color.RGBA{R: 52, G: 28, B: 112, A: 255},  // Royal Amethyst-Indigo
		color.RGBA{R: 16, G: 44, B: 122, A: 255},  // Loop back to Sapphire
	)

	nPix := w * h
	zBuf := make([]float64, nPix)
	fanIDBuf := make([]int, nPix)
	nxBuf := make([]float64, nPix)
	nyBuf := make([]float64, nPix)
	nzBuf := make([]float64, nPix)
	albR := make([]float64, nPix)
	albG := make([]float64, nPix)
	albB := make([]float64, nPix)
	grooveBuf := make([]float64, nPix)

	numTiers := pbf.plumes
	if numTiers > 32 {
		numTiers = 24
	}
	fCurl := pbf.flareCurl
	loopX := 0.32 * (math.Cos(phi) - 1.0)
	loopY := 0.32 * math.Sin(phi)

	type fanStratum struct {
		id        int
		baseZ     float64
		rCenter   float64
		bandThick float64
		lobeFreq  float64
		stagger   float64
		dir       float64
		phaseOff  float64
		seedOff   float64
		baseCol   color.RGBA
		streakCol color.RGBA
		accentCol color.RGBA
		deepCol   color.RGBA
		glintCol  color.RGBA
		ridgeLUT  [512]float64
		slopeLUT  [512]float64
		streakLUT [512]float64
		accentLUT [512]float64
		threadLUT [512]float64
	}

	strataList := make([]fanStratum, numTiers)
	gFreq := float64(pbf.filaments) * 0.58

	// Polar origin sits far below the bottom center at (0.50, 2.35) so inside [0, 1]^2,
	// r ranges smoothly from 2.40 (top corners) down to 1.35 (bottom center) — every
	// scalloped fan arch stays broad, calm, and majestic from top to bottom!
	for i := 0; i < numTiers; i++ {
		fi := float64(i) / float64(numTiers-1)
		sOff := float64(i)*19.3 + rng.Float64()*45.0

		// Progress from top outer arches (r = 2.42) down to bottom foreground arches (r = 1.32)
		rCen := 2.42 - 1.10*fi
		thick := 0.088 + 0.020*math.Sin(float64(i)*1.7)
		dir := 1.0
		if i%2 == 1 {
			dir = -1.0
		}

		hue := math.Mod(fi*1.08+0.06*math.Sin(float64(i)*1.3)+0.12+10.0, 1.0)
		bCol := fanGrad.SampleOklab(hue)
		stCol := fanGrad.SampleOklab(math.Mod(hue+0.10, 1.0))
		acCol := fanGrad.SampleOklab(math.Mod(hue-0.09+1.0, 1.0))
		dpCol := render.LerpOklab(bCol, color.RGBA{R: 8, G: 5, B: 16, A: 255}, 0.80)
		glCol := render.LerpOklab(bCol, color.RGBA{R: 255, G: 248, B: 228, A: 255}, 0.74)

		st := fanStratum{
			id:        i + 1,
			baseZ:     4.0 + float64(i)*3.6,
			rCenter:   rCen,
			bandThick: thick,
			// Scale angular frequency with rCen so arch width in pixels stays ~185px everywhere!
			lobeFreq:  3.6 * (rCen / 1.85),
			stagger:   float64(i%2) * math.Pi,
			dir:       dir,
			phaseOff:  float64(i)*0.45 + 0.25*rng.Float64(),
			seedOff:   sOff,
			baseCol:   bCol,
			streakCol: stCol,
			accentCol: acCol,
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
			st.streakLUT[k] = common.Clamp01(0.65*(0.5+0.5*ne.Perlin2D(v*4.8+sOff, 2.7)) + 0.35*(0.5+0.5*pMed))
			st.accentLUT[k] = common.Smoothstep(0.25, 0.78, 0.5+0.5*ne.Perlin2D(v*3.4-sOff*1.2, 6.3))
			st.threadLUT[k] = math.Pow(0.5+0.5*math.Sin(v*gFreq*math.Pi*0.5+2.0*pBand), 6.0) *
				common.Smoothstep(0.0, 0.55, 0.5+0.5*ne.Perlin2D(v*4.5+sOff, 5.7))
		}
		strataList[i] = st
	}

	// =========================================================================
	// 2. PARALLEL RASTERIZATION OF SCALLOPED LOTUS-FAN BROCADE STRATA
	// =========================================================================
	const fanCX = 0.50
	const fanCY = 2.35

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
				dy := uy - fanCY
				for px := 0; px < w; px++ {
					ux := (float64(px) + 0.5) / fw
					dx := ux - fanCX
					idx := py*w + px

					nzBuf[idx] = 1.0
					albR[idx], albG[idx], albB[idx] = 0.015, 0.014, 0.038

					r := math.Hypot(dx, dy)
					// Fan angle alpha in [-0.65, +0.65] measured from upward vertical (0, -1)
					alpha := math.Atan2(dx, -dy)
					urx, ury := dx/r, dy/r
					utx, uty := -ury, urx

					for sIdx := 0; sIdx < numTiers; sIdx++ {
						st := &strataList[sIdx]

						pVal, dPdAlpha, _ := ne.FBM2DDeriv(alpha*2.4+st.seedOff*0.1+loopX*0.35*st.dir, st.rCenter*1.6+loopY*0.35, 2, 2.0, 0.42)

						// Broad, single-domed staggered lotus/shell scallop arches along alpha (no double-bump!)
						lobeArg := alpha*st.lobeFreq*math.Pi + st.stagger + st.phaseOff + st.dir*phi
						scallop := 0.048 * math.Cos(lobeArg)
						dScallop := -0.048 * st.lobeFreq * math.Pi * math.Sin(lobeArg)

						rCurve := st.rCenter + pVal*0.055*fCurl + scallop
						drdAlpha := 0.055*fCurl*dPdAlpha*2.4 + dScallop

						// Exact polar Euclidean perpendicular distance!
						slopeAng := drdAlpha / math.Max(0.35, r)
						invArc := 1.0 / math.Sqrt(1.0+slopeAng*slopeAng)
						dPerp := (r - rCurve) * invArc
						v := dPerp / st.bandThick
						if v < -1.0 || v > 1.0 {
							continue
						}

						nx := (urx - slopeAng*utx) * invArc
						ny := (ury - slopeAng*uty) * invArc

						lutPos := (v + 1.0) * 0.5 * 511.0
						k0 := common.ConstrainInt(int(math.Floor(lutPos)), 0, 510)
						frac := lutPos - float64(k0)

						ridgeH := st.ridgeLUT[k0]*(1.0-frac) + st.ridgeLUT[k0+1]*frac
						dRidge := st.slopeLUT[k0]*(1.0-frac) + st.slopeLUT[k0+1]*frac
						streakMix := st.streakLUT[k0]*(1.0-frac) + st.streakLUT[k0+1]*frac
						accentMix := st.accentLUT[k0]*(1.0-frac) + st.accentLUT[k0+1]*frac
						glintThread := st.threadLUT[k0]*(1.0-frac) + st.threadLUT[k0+1]*frac

						windRipple := ne.Perlin2D(alpha*4.5+st.seedOff, v*12.0)
						ridgeH += 0.09 * windRipple

						// Raised lip along the upper/outer overlapping arch edge (v -> +1)
						edgeCushion := 1.0 - math.Pow(math.Abs(v), 4.2)
						crestLip := 0.54 * math.Exp(-math.Pow((v-0.88)/0.10, 2.0))
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

						// 3-Band Inlaid Polychrome Silk-Brocade + 24K Gold Selvedge & Pinstripe!
						absV := math.Abs(v)
						goldSpec := color.RGBA{R: 255, G: 246, B: 204, A: 255}
						goldBody := color.RGBA{R: 238, G: 182, B: 52, A: 255}

						centerBand := common.Smoothstep(0.41, 0.35, absV)
						colFlank := render.LerpOklab(st.baseCol, st.accentCol, accentMix*0.38)
						colCore := render.LerpOklab(st.streakCol, st.glintCol, 0.20+0.35*streakMix)
						colMid := render.LerpOklab(colFlank, colCore, centerBand*0.82)

						var pixCol color.RGBA
						if ridgeH < 0 {
							gT := common.Clamp01(-ridgeH * 1.30)
							pixCol = render.LerpOklab(colMid, st.deepCol, gT*0.82)
						} else {
							cT := common.Clamp01(ridgeH * 1.15)
							crestCurve := 0.5 + 0.5*math.Cos(lobeArg)
							glintW := common.Clamp01(cT*0.34 + 0.62*glintThread*(0.4+0.6*crestCurve))
							pixCol = render.LerpOklab(colMid, st.glintCol, glintW)
						}

						pinstripe := math.Exp(-math.Pow((absV-0.38)/0.022, 2.0))
						selvedge := common.Smoothstep(0.89, 0.95, absV)
						goldMask := common.Clamp01(math.Max(pinstripe*0.88, selvedge))
						if goldMask > 0.01 {
							gCol := render.LerpOklab(goldBody, goldSpec, common.Clamp01(0.5+0.5*ridgeH))
							pixCol = render.LerpOklab(pixCol, gCol, goldMask)
							ridgeH = ridgeH*(1.0-goldMask) + 0.55*goldMask
						}

						rL, gL, bL, _ := render.ColorToLinearRGB(pixCol)
						zBuf[idx] = zSurf
						fanIDBuf[idx] = st.id
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
	// 3. DEFERRED 3D SHADING: DEEP CONTACT AO + HORIZON SHADOWS + SPECULAR
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
					id0 := fanIDBuf[idx]

					contactOcc := 0.0
					for _, off := range aoRing {
						sx := common.ConstrainInt(x+off[0], 0, w-1)
						sy := common.ConstrainInt(y+off[1], 0, h-1)
						sIdx := sy*w + sx
						dz := zBuf[sIdx] - z0
						if dz > 0.25 {
							weight := 0.85
							if fanIDBuf[sIdx] != id0 {
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
						if fanIDBuf[sy*w+sx] != id0 && rawDz > 0 {
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

					macroLight := 0.30 + 0.82*common.Smoothstep(1.10, -0.05, uy*0.60-ux*0.25)

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
	// 4. SUBTLE 24K GOLD-LEAF STARLIGHT MOTES
	// =========================================================================
	spatterRng := rand.New(rand.NewSource(seed + 6193))
	goldFleck := color.RGBA{R: 255, G: 236, B: 172, A: 255}

	for i := 0; i < 110; i++ {
		fx := spatterRng.Float64() * fw
		fy := spatterRng.Float64() * fh
		rad := (0.65 + 1.25*math.Pow(spatterRng.Float64(), 2.4)) * scale
		alpha := 0.48 + 0.28*spatterRng.Float64()

		minX := common.ConstrainInt(int(math.Floor(fx-rad-1)), 0, w-1)
		maxX := common.ConstrainInt(int(math.Ceil(fx+rad+1)), 0, w-1)
		minY := common.ConstrainInt(int(math.Floor(fy-rad-1)), 0, h-1)
		maxY := common.ConstrainInt(int(math.Ceil(fy+rad+1)), 0, h-1)
		for py := minY; py <= maxY; py++ {
			for px := minX; px <= maxX; px++ {
				d := math.Hypot(float64(px)+0.5-fx, float64(py)+0.5-fy)
				if d <= rad+0.6 {
					cov := common.Clamp01(rad + 0.5 - d)
					rc.BlendPixel(px, py, goldFleck, alpha*cov)
				}
			}
		}
	}

	c.MergeRenderCanvas(rc)
}
