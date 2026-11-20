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

type perlinImpastoSwirl struct {
	swaths       int
	bristles     int
	curlStrength float64
}

// NewPerlinImpastoSwirl returns a PerlinImpastoSwirl generator that synthesizes a
// full-bleed 3D diagonal S-wave & vortex-braided combed-impasto silk relief driven by
// exact Euclidean normal-distance curvilinear coordinates: 24 continuous, edge-to-edge
// S-bending polychrome oil & silk streams with multi-pigment streaks, crisp hog-bristle
// micro-ridges, and deep inter-ribbon cast shadows.
func NewPerlinImpastoSwirl(swaths, bristles int, curlStrength float64) *perlinImpastoSwirl {
	if swaths <= 0 {
		swaths = 24
	}
	if bristles <= 0 {
		bristles = 58
	}
	if curlStrength <= 0 {
		curlStrength = 1.35
	}
	return &perlinImpastoSwirl{
		swaths:       common.ConstrainInt(swaths, 12, 42),
		bristles:     common.ConstrainInt(bristles, 28, 96),
		curlStrength: common.Constrain(curlStrength, 0.5, 2.5),
	}
}

// Generative renders the static PerlinImpastoSwirl composition.
func (pis *perlinImpastoSwirl) Generative(c *generativeart.Canva) {
	pis.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360° looping PerlinImpastoSwirl animation.
func (pis *perlinImpastoSwirl) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
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
		seed = 20263011
	}
	rng := rand.New(rand.NewSource(seed))
	ne := common.NewNoiseEngine(seed)

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	// =========================================================================
	// 1. HIGH-CHROMA STARRY SAPPHIRE, CYAN, CHAMPAGNE GOLD, CORAL & ROSE PALETTE
	// =========================================================================
	streamGrad := render.ResolveCyclicGradient(c.Opts().ColorSchema(), common.Dunhuang,
		color.RGBA{R: 16, G: 128, B: 162, A: 255}, // Cerulean Kingfisher Cyan
		color.RGBA{R: 38, G: 172, B: 132, A: 255}, // Luminous Seafoam Emerald
		color.RGBA{R: 248, G: 192, B: 52, A: 255}, // Imperial Sunburst Gold
		color.RGBA{R: 242, G: 104, B: 44, A: 255}, // Fiery Sunset Tangerine-Coral
		color.RGBA{R: 206, G: 36, B: 58, A: 255},  // Alizarin Cinnabar Crimson
		color.RGBA{R: 136, G: 24, B: 88, A: 255},  // Bordeaux Rose-Plum
		color.RGBA{R: 36, G: 58, B: 142, A: 255},  // Royal Ultramarine Sapphire
		color.RGBA{R: 16, G: 128, B: 162, A: 255}, // Loop back to Kingfisher Cyan
	)

	nPix := w * h
	zBuf := make([]float64, nPix)
	swathIDBuf := make([]int, nPix)
	nxBuf := make([]float64, nPix)
	nyBuf := make([]float64, nPix)
	nzBuf := make([]float64, nPix)
	albR := make([]float64, nPix)
	albG := make([]float64, nPix)
	albB := make([]float64, nPix)
	grooveBuf := make([]float64, nPix)

	// =========================================================================
	// 2. PRECOMPUTE 24 SWEEPING DIAGONAL S-CURVE & VORTEX-BRAIDED RIBBONS
	// =========================================================================
	numSwaths := pis.swaths
	if numSwaths > 32 {
		numSwaths = 24
	}
	cStr := pis.curlStrength
	loopX := 0.14 * math.Cos(phi)
	loopY := 0.14 * math.Sin(phi)

	type sStreamRibbon struct {
		id        int
		baseZ     float64
		theta     float64
		cosT      float64
		sinT      float64
		qCenter   float64
		bandThick float64
		freqS     float64
		sBendAmp  float64
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

	numSwaths = 28
	ribbonList := make([]sStreamRibbon, numSwaths)
	gFreq := float64(pis.bristles) * 0.58

	for i := 0; i < numSwaths; i++ {
		fi := float64(i) / float64(numSwaths-1)
		sOff := float64(i)*19.1 + rng.Float64()*45.0

		// Progress across the canvas from -0.78 to +0.78 with generous overlap (zero gaps!)
		qCen := -0.78 + 1.56*fi
		thick := 0.128 + 0.028*math.Sin(float64(i)*1.7)
		// Gently varying sweep angle per swath so ribbons interweave organically!
		th := -0.52 + 0.34*math.Sin(float64(i)*0.78+0.4)

		hue := math.Mod(fi*1.06+0.05*math.Sin(float64(i)*1.2)+0.16+10.0, 1.0)
		bCol := streamGrad.SampleOklab(hue)
		stCol := streamGrad.SampleOklab(math.Mod(hue+0.09, 1.0))
		acCol := streamGrad.SampleOklab(math.Mod(hue-0.08+1.0, 1.0))
		dpCol := render.LerpOklab(bCol, color.RGBA{R: 8, G: 5, B: 16, A: 255}, 0.80)
		glCol := render.LerpOklab(bCol, color.RGBA{R: 255, G: 248, B: 228, A: 255}, 0.74)

		rb := sStreamRibbon{
			id:        i + 1,
			baseZ:     4.0 + float64(i)*3.6,
			theta:     th,
			cosT:      math.Cos(th),
			sinT:      math.Sin(th),
			qCenter:   qCen,
			bandThick: thick,
			freqS:     0.88 + 0.18*math.Sin(float64(i)*0.9),
			sBendAmp:  (0.16 + 0.05*math.Cos(float64(i)*0.7)) * (0.75 + 0.22*cStr),
			phaseOff:  float64(i)*0.52 + 0.40*math.Sin(float64(i)*0.85),
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

			rb.ridgeLUT[k] = 0.28*pBand + 0.28*pMed + 0.20*pFine + 0.24*wave
			rb.slopeLUT[k] = common.Constrain(
				(0.28*gFreq*0.28*dBand+0.28*gFreq*0.75*dMed+0.20*gFreq*1.60*dFine+0.24*dWave)/(gFreq*1.15),
				-1.45, 1.45,
			)
			rb.streakLUT[k] = common.Clamp01(0.65*(0.5+0.5*ne.Perlin2D(v*4.8+sOff, 2.7)) + 0.35*(0.5+0.5*pMed))
			rb.accentLUT[k] = common.Smoothstep(0.25, 0.78, 0.5+0.5*ne.Perlin2D(v*3.4-sOff*1.2, 6.3))
			rb.threadLUT[k] = math.Pow(0.5+0.5*math.Sin(v*gFreq*math.Pi*0.5+2.0*pBand), 6.0) *
				common.Smoothstep(0.0, 0.55, 0.5+0.5*ne.Perlin2D(v*4.5+sOff, 5.7))
		}
		ribbonList[i] = rb
	}

	// =========================================================================
	// 3. PARALLEL RASTERIZATION WITH EXACT EUCLIDEAN NORMAL DISTANCE
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
				dy := uy - 0.5
				for px := 0; px < w; px++ {
					ux := (float64(px) + 0.5) / fw
					dx := ux - 0.5
					idx := py*w + px

					// Dark sapphire-amethyst fallback underpainting so zero black pixels can ever appear
					nzBuf[idx] = 1.0
					albR[idx], albG[idx], albB[idx] = 0.015, 0.012, 0.038

					for rIdx := 0; rIdx < numSwaths; rIdx++ {
						rb := &ribbonList[rIdx]

						// Per-swath rotated stream coordinates (s, q)
						s := dx*rb.cosT + dy*rb.sinT
						q := -dx*rb.sinT + dy*rb.cosT

						// Smooth Perlin FBM + grand sweeping S-bend along s (zero sharp kinks!)
						sArg := (s + 0.75) * rb.freqS
						pVal, dPds, _ := ne.FBM2DDeriv(sArg+rb.seedOff*0.1+loopX*0.35, rb.qCenter*1.8+loopY*0.35, 2, 2.0, 0.45)

						sWave1 := rb.sBendAmp * math.Sin(s*math.Pi*1.45+rb.phaseOff+phi)
						dsWave1 := rb.sBendAmp * math.Pi * 1.45 * math.Cos(s*math.Pi*1.45+rb.phaseOff+phi)

						sWave2 := 0.038 * math.Cos(s*math.Pi*2.10-rb.phaseOff*0.7-phi)
						dsWave2 := -0.038 * math.Pi * 2.10 * math.Sin(s*math.Pi*2.10-rb.phaseOff*0.7-phi)

						qCurve := rb.qCenter + pVal*0.11*cStr + sWave1 + sWave2
						dqds := 0.11*cStr*dPds*rb.freqS + dsWave1 + dsWave2

						// Organic width modulation along s so ribbons swell and taper like calligrapher strokes
						widthMod := 0.88 + 0.28*(0.5+0.5*math.Sin(s*math.Pi*1.35+rb.phaseOff*1.2+0.8))
						locThick := rb.bandThick * widthMod

						// EXACT Euclidean perpendicular distance to the curved spine!
						invArc := 1.0 / math.Sqrt(1.0+dqds*dqds)
						edgeScallop := 0.005 * math.Tanh(2.6*ne.Perlin2D(s*2.8+rb.seedOff, 3.1))
						dPerp := (q - (qCurve + edgeScallop)) * invArc
						v := dPerp / locThick
						if v < -1.0 || v > 1.0 {
							continue
						}

						// Unit normal in (s, q) space is (-dqds, 1) * invArc; rotate back to screen (x, y)
						ns := -dqds * invArc
						nq := invArc
						nx := ns*rb.cosT - nq*rb.sinT
						ny := ns*rb.sinT + nq*rb.cosT

						lutPos := (v + 1.0) * 0.5 * 511.0
						k0 := common.ConstrainInt(int(math.Floor(lutPos)), 0, 510)
						frac := lutPos - float64(k0)

						ridgeH := rb.ridgeLUT[k0]*(1.0-frac) + rb.ridgeLUT[k0+1]*frac
						dRidge := rb.slopeLUT[k0]*(1.0-frac) + rb.slopeLUT[k0+1]*frac
						streakMix := rb.streakLUT[k0]*(1.0-frac) + rb.streakLUT[k0+1]*frac
						accentMix := rb.accentLUT[k0]*(1.0-frac) + rb.accentLUT[k0+1]*frac
						glintThread := rb.threadLUT[k0]*(1.0-frac) + rb.threadLUT[k0+1]*frac

						windRipple := ne.Perlin2D(s*4.0+rb.seedOff, v*12.0)
						ridgeH += 0.10 * windRipple

						edgeCushion := 1.0 - math.Pow(math.Abs(v), 4.2)
						crestLip := 0.54 * math.Exp(-math.Pow((v+0.88)/0.10, 2.0))
						zSurf := rb.baseZ + 2.8*edgeCushion + 1.2*crestLip + 0.88*ridgeH

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

						// 3-Band Inlaid Polychrome Silk-Brocade + 24K Gold Selvedge & Pinstripe Cross-Section!
						absV := math.Abs(v)
						goldSpec := color.RGBA{R: 255, G: 246, B: 204, A: 255}
						goldBody := color.RGBA{R: 238, G: 182, B: 52, A: 255}

						// Inner contrasting silk-satin band (|v| < 0.36) vs outer primary silk flank (|v| > 0.40)
						centerBand := common.Smoothstep(0.41, 0.35, absV)
						colFlank := render.LerpOklab(rb.baseCol, rb.accentCol, accentMix*0.38)
						colCore := render.LerpOklab(rb.streakCol, rb.glintCol, 0.20+0.35*streakMix)
						colMid := render.LerpOklab(colFlank, colCore, centerBand*0.82)

						var pixCol color.RGBA
						if ridgeH < 0 {
							gT := common.Clamp01(-ridgeH * 1.30)
							pixCol = render.LerpOklab(colMid, rb.deepCol, gT*0.82)
						} else {
							cT := common.Clamp01(ridgeH * 1.15)
							crestCurve := 0.5 + 0.5*math.Cos(s*math.Pi*2.2+rb.phaseOff)
							glintW := common.Clamp01(cT*0.34 + 0.62*glintThread*(0.4+0.6*crestCurve))
							pixCol = render.LerpOklab(colMid, rb.glintCol, glintW)
						}

						// Gleaming 24K gold inner pinstripe divider at |v| ≈ 0.38 and outer selvedge rim at |v| > 0.90!
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
						swathIDBuf[idx] = rb.id
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
					id0 := swathIDBuf[idx]

					contactOcc := 0.0
					for _, off := range aoRing {
						sx := common.ConstrainInt(x+off[0], 0, w-1)
						sy := common.ConstrainInt(y+off[1], 0, h-1)
						sIdx := sy*w + sx
						dz := zBuf[sIdx] - z0
						if dz > 0.25 {
							weight := 0.85
							if swathIDBuf[sIdx] != id0 {
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
						if swathIDBuf[sy*w+sx] != id0 && rawDz > 0 {
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
	// 5. SUBTLE 24K GOLD-LEAF STARLIGHT MOTES (ZERO DARK PEPPER DOTS!)
	// =========================================================================
	spatterRng := rand.New(rand.NewSource(seed + 7719))
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
