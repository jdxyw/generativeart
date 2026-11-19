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

type kagomeSilkWeave struct {
	bands     int
	filaments int
	waveAmp   float64
}

// NewKagomeSilkWeave returns a KagomeSilkWeave generator that synthesizes a
// full-bleed 3D triaxial Kagome (笼目) over-under interlaced bamboo & gilded
// silk-brocade lattice over recessed hexagonal star-windows: 3 hexagonal axes
// (0°, +60°, -60°) of Perlin-undulating 3-band silk-brocade ribbons that weave
// cyclically over one axis and under the other in true 3D depth, framing
// recessed golden-and-turquoise cabochon jewels in the hexagonal openings.
func NewKagomeSilkWeave(bands, filaments int, waveAmp float64) *kagomeSilkWeave {
	if bands <= 0 {
		bands = 7
	}
	if filaments <= 0 {
		filaments = 48
	}
	if waveAmp <= 0 {
		waveAmp = 1.25
	}
	return &kagomeSilkWeave{
		bands:     common.ConstrainInt(bands, 4, 12),
		filaments: common.ConstrainInt(filaments, 24, 84),
		waveAmp:   common.Constrain(waveAmp, 0.5, 2.2),
	}
}

// Generative renders the static KagomeSilkWeave composition.
func (ksw *kagomeSilkWeave) Generative(c *generativeart.Canva) {
	ksw.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360° looping KagomeSilkWeave animation.
func (ksw *kagomeSilkWeave) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
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
		seed = 20263302
	}
	rng := rand.New(rand.NewSource(seed))
	ne := common.NewNoiseEngine(seed)

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	// =========================================================================
	// 1. TRIAXIAL SILK-BROCADE PALETTES FOR THE 3 HEXAGONAL AXES (0°, +60°, -60°)
	// =========================================================================
	// Axis 0 (0°): Imperial Sunburst Gold, Amber Ochre & Sunset Tangerine
	// Axis 1 (+60°): Kingfisher Turquoise, Peacock Emerald & Royal Sapphire
	// Axis 2 (-60°): Cinnabar Vermilion, Alizarin Rose-Ruby & Bordeaux Amethyst
	axisGrads := [3]*render.MultiStopGradient{
		render.NewMultiStopGradient(
			color.RGBA{R: 250, G: 196, B: 54, A: 255},
			color.RGBA{R: 242, G: 132, B: 46, A: 255},
			color.RGBA{R: 248, G: 214, B: 96, A: 255},
			color.RGBA{R: 250, G: 196, B: 54, A: 255},
		),
		render.NewMultiStopGradient(
			color.RGBA{R: 20, G: 144, B: 162, A: 255},
			color.RGBA{R: 38, G: 172, B: 128, A: 255},
			color.RGBA{R: 22, G: 78, B: 152, A: 255},
			color.RGBA{R: 20, G: 144, B: 162, A: 255},
		),
		render.NewMultiStopGradient(
			color.RGBA{R: 214, G: 44, B: 56, A: 255},
			color.RGBA{R: 156, G: 28, B: 82, A: 255},
			color.RGBA{R: 238, G: 96, B: 52, A: 255},
			color.RGBA{R: 214, G: 44, B: 56, A: 255},
		),
	}

	goldSpec := color.RGBA{R: 255, G: 246, B: 204, A: 255}
	goldBody := color.RGBA{R: 238, G: 182, B: 52, A: 255}

	// Precompute 512-entry silk groove LUTs for the 3 axes
	var ridgeLUT [3][512]float64
	var slopeLUT [3][512]float64
	var threadLUT [3][512]float64
	gFreq := float64(ksw.filaments) * 0.52

	for a := 0; a < 3; a++ {
		sOff := float64(a)*29.3 + rng.Float64()*20.0
		for k := 0; k < 512; k++ {
			v := (float64(k)/511.0)*2.0 - 1.0
			pBand, dBand, _ := ne.Perlin2DDeriv(v*gFreq*0.28+sOff, 1.9)
			pFine, dFine, _ := ne.Perlin2DDeriv(v*gFreq*1.45-sOff*1.3, 5.1)
			wave := math.Sin(v*gFreq*math.Pi + 1.15*pBand)
			dWave := (gFreq*math.Pi + 1.15*gFreq*0.28*dBand) * math.Cos(v*gFreq*math.Pi+1.15*pBand)

			ridgeLUT[a][k] = 0.34*pBand + 0.32*pFine + 0.34*wave
			slopeLUT[a][k] = common.Constrain((0.34*gFreq*0.28*dBand+0.32*gFreq*1.45*dFine+0.34*dWave)/(gFreq*1.15), -1.35, 1.35)
			threadLUT[a][k] = math.Pow(0.5+0.5*math.Sin(v*gFreq*math.Pi*0.5+1.8*pBand), 5.5)
		}
	}

	// =========================================================================
	// 2. PARALLEL RASTERIZATION OF TRUE 3D OVER-UNDER TRIAXIAL KAGOME WEAVE
	// =========================================================================
	nPix := w * h
	zBuf := make([]float64, nPix)
	idBuf := make([]int, nPix)
	nxBuf := make([]float64, nPix)
	nyBuf := make([]float64, nPix)
	nzBuf := make([]float64, nPix)
	albR := make([]float64, nPix)
	albG := make([]float64, nPix)
	albB := make([]float64, nPix)
	grooveBuf := make([]float64, nPix)

	// 3 hexagonal normal vectors N_a = (-sin(theta_a), cos(theta_a)) such that N_0 + N_1 + N_2 = 0!
	// theta_0 = 0, theta_1 = 2*pi/3 (120°), theta_2 = 4*pi/3 (240°)
	angles := [3]float64{0.12, 0.12 + 2.0*math.Pi/3.0, 0.12 + 4.0*math.Pi/3.0}
	var cosA, sinA [3]float64
	for a := 0; a < 3; a++ {
		cosA[a] = math.Cos(angles[a])
		sinA[a] = math.Sin(angles[a])
	}

	freqLattice := float64(ksw.bands) * 0.78
	if freqLattice > 6.2 {
		freqLattice = 5.4
	}
	wAmp := ksw.waveAmp
	loopX := 0.48 * (math.Cos(phi) - 1.0)
	loopY := 0.48 * math.Sin(phi)

	// In a Kagome lattice where N_0 + N_1 + N_2 = 0:
	// Shift each coordinate by +1/6 so q_0 + q_1 + q_2 = 1/2!
	// When ribbon centers sit at integers (dq_a = q_a - round(q_a) == 0), a crossing
	// between Axis 0 (q_0 = 0) and Axis 1 (q_1 = 0) has q_2 = 1/2 - 0 - 0 = 1/2
	// (maximally far from Axis 2's center!), and cos(2*pi*q_1) - cos(2*pi*q_2) = +1 - (-1) = +2
	// is at a flat stationary maximum across the entire crossing diamond!
	const ribbonHalfW = 0.215 // Ribbon occupies |v| <= 1 where v = dq / ribbonHalfW

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
				dy := uy - 0.50
				for px := 0; px < w; px++ {
					ux := (float64(px) + 0.5) / fw
					dx := ux - 0.50
					idx := py*w + px

					// Smooth Perlin organic breathing + zero-sum 3-phase orbital Kagome lattice glide!
					// Notice sum_{a=0..2} sin(phi + 2*pi*a/3) - sin(2*pi*a/3) == 0 identically,
					// so q_0 + q_1 + q_2 == 1/2 stays 100% exact while the entire triaxial weave glides & undulates!
					w1 := 0.052 * wAmp * ne.Perlin2D(ux*1.45+1.3+loopX*0.45, uy*1.45-0.9+loopY*0.45)
					w2 := 0.052 * wAmp * ne.Perlin2D(ux*1.45-2.1-loopY*0.45, uy*1.45+1.7+loopX*0.45)
					warps := [3]float64{w1, w2, -(w1 + w2)}

					// Transverse S-wave ripple along the perpendicular axes that also sums to zero at every pixel!
					waveR := 0.14 * math.Sin((dx*2.2+dy*1.8)*math.Pi+phi) - 0.14*math.Sin((dx*2.2+dy*1.8)*math.Pi)
					waveG := 0.14 * math.Sin((-dx*1.9+dy*2.1)*math.Pi-phi) - 0.14*math.Sin((-dx*1.9+dy*2.1)*math.Pi)
					dynWaves := [3]float64{waveR, waveG, -(waveR + waveG)}

					var q, s [3]float64
					for a := 0; a < 3; a++ {
						nxA := -sinA[a]
						nyA := cosA[a]
						txA := cosA[a]
						tyA := sinA[a]
						phaseA := float64(a) * (2.0 * math.Pi / 3.0)
						// Zero-sum 3-phase orbital translation (0 at phi=0, sweeps 0.56 lattice periods across loop!)
						orbitalGlide := 0.28 * (math.Sin(phi+phaseA) - math.Sin(phaseA))
						s[a] = (dx*txA + dy*tyA) * freqLattice
						// Shift by +1/6 so q_0 + q_1 + q_2 = 1/2; ribbon centers sit at integers!
						q[a] = (dx*nxA+dy*nyA)*freqLattice + warps[a] + orbitalGlide + dynWaves[a] + (1.0 / 6.0)
					}

					// Base Layer: Recessed Hexagonal Star-Window Sanctuary with Concentric Gold & Jewel Cabochons!
					// Since q_0 + q_1 + q_2 = 1/2 (mod 1) and ribbons sit at q_a = integer,
					// the exact center of every open hexagonal window is where all three q_a == integer + 0.5!
					var hexDistSq float64
					for a := 0; a < 3; a++ {
						pHex := q[a] - 0.5
						dp := pHex - math.Round(pHex)
						hexDistSq += dp * dp
					}
					rHex := math.Sqrt(hexDistSq / 1.5)

					bgCol := color.RGBA{R: 10, G: 18, B: 46, A: 255}
					curZ := 1.0
					curID := 1
					curGroove := 0.0
					jewelPulse := 1.0 + 0.14*math.Sin(phi)
					rHexScaled := rHex / jewelPulse
					if rHexScaled < 0.25 {
						if rHexScaled < 0.095 {
							bgCol = color.RGBA{R: 255, G: 222, B: 98, A: 255}
							curZ = 3.8
							curID = 4
						} else if rHexScaled < 0.175 {
							bgCol = color.RGBA{R: 22, G: 152, B: 168, A: 255}
							curZ = 2.8
							curID = 3
						} else {
							bgCol = color.RGBA{R: 196, G: 44, B: 62, A: 255}
							curZ = 1.9
							curID = 2
						}
						curGroove = 0.38 * math.Sin(rHex*52.0-phi)
					}
					rL, gL, bL, _ := render.ColorToLinearRGB(bgCol)
					zBuf[idx] = curZ
					idBuf[idx] = curID
					nzBuf[idx] = 1.0
					albR[idx], albG[idx], albB[idx] = rL, gL, bL
					grooveBuf[idx] = curGroove

					// Test all 3 triaxial ribbon families (a = 0, 1, 2)
					for a := 0; a < 3; a++ {
						kBand := int(math.Round(q[a]))
						dq := q[a] - float64(kBand) // [-0.5, +0.5]; ribbon center is at dq == 0!
						v := dq / ribbonHalfW
						if math.Abs(v) > 1.0 {
							continue
						}

						// TRUE TRIAXIAL OVER-UNDER 3D ARCHING:
						// At crossing with aOver (dq_{aOver} ≈ 0, dq_{aUnder} ≈ 1/2):
						// cos(2*pi*q_{aOver}) ≈ +1 and cos(2*pi*q_{aUnder}) ≈ -1, so overWave - underWave ≈ +2.0!
						aOver := (a + 1) % 3
						aUnder := (a + 2) % 3
						overWave := math.Cos(2.0 * math.Pi * q[aOver])
						underWave := math.Cos(2.0 * math.Pi * q[aUnder])
						zWeave := 3.4 * (overWave - underWave)
						dzds := (3.4 * 2.0 * math.Pi / math.Sqrt(3.0)) * (-math.Sin(2.0*math.Pi*q[aOver]) - math.Sin(2.0*math.Pi*q[aUnder])) * 0.16

						lutPos := (v + 1.0) * 0.5 * 511.0
						k0 := common.ConstrainInt(int(math.Floor(lutPos)), 0, 510)
						fLut := lutPos - float64(k0)

						ridgeH := ridgeLUT[a][k0]*(1.0-fLut) + ridgeLUT[a][k0+1]*fLut
						dRidge := slopeLUT[a][k0]*(1.0-fLut) + slopeLUT[a][k0+1]*fLut
						glintThread := threadLUT[a][k0]*(1.0-fLut) + threadLUT[a][k0+1]*fLut

						edgeCushion := 1.0 - math.Pow(math.Abs(v), 3.6)
						zSurf := 8.5 + zWeave + 2.6*edgeCushion + 0.78*ridgeH
						if zSurf < zBuf[idx] {
							continue
						}

						nxA := -sinA[a]
						nyA := cosA[a]
						txA := cosA[a]
						tyA := sinA[a]

						edgeBevel := 0.65 * v * math.Pow(math.Abs(v), 2.2)
						transSlope := edgeBevel - 0.95*dRidge
						normX := transSlope*nxA - dzds*txA
						normY := transSlope*nyA - dzds*tyA
						normZ := 1.0
						invN := 1.0 / math.Sqrt(normX*normX+normY*normY+normZ*normZ)
						normX, normY, normZ = normX*invN, normY*invN, normZ*invN

						// 3-Band Inlaid Silk-Brocade + 24K Gold Selvedge & Pinstripe along each Kagome ribbon!
						hue := math.Mod(0.18*float64(kBand+10)+0.22*math.Sin(s[a]*0.95-phi*float64(1-2*(a%2))), 1.0)
						colFlank := axisGrads[a].SampleOklab(hue)
						colCore := axisGrads[(a+1)%3].SampleOklab(math.Mod(hue+0.25, 1.0))
						deepCol := render.LerpOklab(colFlank, color.RGBA{R: 8, G: 5, B: 18, A: 255}, 0.80)
						glintCol := render.LerpOklab(colFlank, color.RGBA{R: 255, G: 248, B: 224, A: 255}, 0.72)

						absV := math.Abs(v)
						centerBand := common.Smoothstep(0.42, 0.34, absV)
						colMid := render.LerpOklab(colFlank, colCore, centerBand*0.84)

						var pixCol color.RGBA
						if ridgeH < 0 {
							gT := common.Clamp01(-ridgeH * 1.30)
							pixCol = render.LerpOklab(colMid, deepCol, gT*0.82)
						} else {
							cT := common.Clamp01(ridgeH * 1.15)
							archBoost := common.Clamp01(0.5 + 0.15*zWeave)
							glintW := common.Clamp01(cT*0.34 + 0.62*glintThread*archBoost)
							pixCol = render.LerpOklab(colMid, glintCol, glintW)
						}

						pinstripe := math.Exp(-math.Pow((absV-0.38)/0.026, 2.0))
						selvedge := common.Smoothstep(0.87, 0.94, absV)
						goldMask := common.Clamp01(math.Max(pinstripe*0.86, selvedge))
						if goldMask > 0.01 {
							gCol := render.LerpOklab(goldBody, goldSpec, common.Clamp01(0.5+0.5*ridgeH))
							pixCol = render.LerpOklab(pixCol, gCol, goldMask)
							ridgeH = ridgeH*(1.0-goldMask) + 0.55*goldMask
						}

						prL, pgL, pbL, _ := render.ColorToLinearRGB(pixCol)
						zBuf[idx] = zSurf
						idBuf[idx] = (a+1)*1000 + (kBand + 50)
						nxBuf[idx] = normX
						nyBuf[idx] = normY
						nzBuf[idx] = normZ
						albR[idx], albG[idx], albB[idx] = prL, pgL, pbL
						grooveBuf[idx] = ridgeH
					}
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	// =========================================================================
	// 3. DEFERRED 3D SHADING: OVER-UNDER CROSSING SHADOWS + CONTACT AO
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
					id0 := idBuf[idx]

					contactOcc := 0.0
					for _, off := range aoRing {
						sx := common.ConstrainInt(x+off[0], 0, w-1)
						sy := common.ConstrainInt(y+off[1], 0, h-1)
						sIdx := sy*w + sx
						dz := zBuf[sIdx] - z0
						if dz > 0.25 {
							weight := 0.85
							if idBuf[sIdx] != id0 {
								weight = 2.55
								dz = math.Min(dz, 4.5)
							}
							contactOcc += weight * common.Constrain(dz/3.8, 0.0, 1.0)
						}
					}
					contactAO := common.Constrain(1.0-(contactOcc/12.0)*0.82, 0.08, 1.0)

					castShadow := 1.0
					for step := 1; step <= 16; step++ {
						dist := float64(step) * scale
						sx := int(math.Round(float64(x) + lx*dist*1.35))
						sy := int(math.Round(float64(y) + ly*dist*1.35))
						if sx < 0 || sx >= w || sy < 0 || sy >= h {
							break
						}
						rawDz := zBuf[sy*w+sx] - z0
						if idBuf[sy*w+sx] != id0 && rawDz > 0 {
							rawDz = math.Min(rawDz, 5.2)
						}
						dz := rawDz - dist*0.26
						if dz > 0.15 {
							pen := common.Constrain(dz/2.6, 0.0, 0.78)
							if 1.0-pen < castShadow {
								castShadow = 1.0 - pen
							}
						}
					}

					macroLight := 0.30 + 0.82*common.Smoothstep(1.10, -0.05, uy*0.60-ux*0.28)

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

	c.MergeRenderCanvas(rc)
}
