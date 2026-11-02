package arts

import (
	"image/color"
	"math"
	"runtime"
	"sync"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type gaborWaveInterference struct {
	vortexPoles int
	gaborFreq   float64
	bandwidth   float64
}

// NewGaborWaveInterference returns a GaborWaveInterference generator that synthesizes
// a luminous Celestial Astrolabe & Perlin-Gabor Caustic Wave Medallion with dihedral
// Chladni-Gabor standing waves, Airy thin-film structural color ridges, additive curl
// caustic veils, and a double-ringed 24k burnished gold astrolabe bezel.
func NewGaborWaveInterference(vortexPoles int, gaborFreq, bandwidth float64) *gaborWaveInterference {
	if vortexPoles <= 0 {
		vortexPoles = 6
	}
	if gaborFreq <= 0 {
		gaborFreq = 3.2
	}
	if bandwidth <= 0 {
		bandwidth = 1.45
	}
	return &gaborWaveInterference{
		vortexPoles: vortexPoles,
		gaborFreq:   gaborFreq,
		bandwidth:   bandwidth,
	}
}

// Generative renders the static GaborWaveInterference medallion composition.
func (g *gaborWaveInterference) Generative(c *generativeart.Canva) {
	g.GenerativeFrame(c, 0, 1, 0)
}

// GenerativeFrame renders frame frameIdx of a seamless looping GaborWaveInterference animation.
func (g *gaborWaveInterference) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
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
		seed = 20261904
	}
	ne := common.NewNoiseEngine(seed + 1904)

	isAnim := totalFrames > 1
	phi := 2.0 * math.Pi * phase
	loopX, loopY := 0.0, 0.0
	if isAnim {
		loopX = 0.08 * math.Cos(phi)
		loopY = 0.08 * math.Sin(phi)
	}

	cx, cy := 0.5*fw, 0.5*fh
	medallionR := 0.42 * minDim
	invR := 1.0 / math.Max(medallionR, 1.0)

	numPoles := g.vortexPoles
	if numPoles < 3 {
		numPoles = 6
	}
	mFold := float64(numPoles)

	// Palette definitions
	obsidianMidnight := color.RGBA{R: 5, G: 8, B: 20, A: 255}     // #050814
	abyssalSapphire := color.RGBA{R: 7, G: 13, B: 36, A: 255}     // #070d24
	electricCyan := color.RGBA{R: 40, G: 232, B: 224, A: 255}     // #28e8e0
	solarGold := color.RGBA{R: 255, G: 190, B: 56, A: 255}        // #ffbe38
	neonRoseMagenta := color.RGBA{R: 240, G: 56, B: 120, A: 255}  // #f03878
	antiqueBrass := color.RGBA{R: 184, G: 132, B: 44, A: 255}     // #b8842c
	ivoryPearl := color.RGBA{R: 255, G: 248, B: 232, A: 255}      // #fff8e8

	rc := c.NewRenderCanvas()
	rc.FillBackground(obsidianMidnight)
	rc.SetBlendMode(render.BlendOver)

	obsR, obsG, obsB, _ := render.ColorToLinearRGB(obsidianMidnight)
	sapR, sapG, sapB, _ := render.ColorToLinearRGB(abyssalSapphire)
	cyaR, cyaG, cyaB, _ := render.ColorToLinearRGB(electricCyan)
	gldR, gldG, gldB, _ := render.ColorToLinearRGB(solarGold)
	magR, magG, magB, _ := render.ColorToLinearRGB(neonRoseMagenta)
	ivoR, ivoG, ivoB, _ := render.ColorToLinearRGB(ivoryPearl)

	// 1. Inside the Medallion — Multi-Harmonic Gabor + Perlin Caustic Interference Field
	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}

	outerBoundR := medallionR + 2.0*scale
	var wg sync.WaitGroup
	for wIdx := 0; wIdx < workers; wIdx++ {
		y0 := (wIdx * h) / workers
		y1 := ((wIdx + 1) * h) / workers
		wg.Add(1)
		go func(rowStart, rowEnd int) {
			defer wg.Done()
			for y := rowStart; y < rowEnd; y++ {
				dy := (float64(y) + 0.5) - cy
				if math.Abs(dy) > outerBoundR {
					continue
				}
				for x := 0; x < w; x++ {
					dx := (float64(x) + 0.5) - cx
					if math.Abs(dx) > outerBoundR {
						continue
					}
					r := math.Hypot(dx, dy)
					if r > outerBoundR {
						continue
					}

					// Smooth anti-aliased medallion boundary mask
					edgeCov := common.Clamp01((medallionR + 0.75*scale - r) / (1.5 * scale))
					if edgeCov <= 0 {
						continue
					}

					// Normalized polar & Cartesian medallion coordinates in [-1, 1]
					ux := dx * invR
					uy := dy * invR
					rn := r * invR
					theta := math.Atan2(uy, ux)

					// Analytical Perlin FBM value & spatial gradient for organic phase warping
					fbmVal, dfdx, dfdy := ne.FBM2DDeriv(ux*2.15+loopX, uy*2.15+loopY, 4, 2.0, 0.5)
					gradMag := math.Hypot(dfdx, dfdy)
					flowOrient := theta + 0.5*math.Pi + 0.38*fbmVal + 0.16*math.Atan2(dfdy, dfdx)

					// Symmetric vortexPoles-fold folded domain so Gabor wave packets marry
					// dihedral mandala symmetry with organic Perlin caustic turbulence
					foldAng := math.Cos(mFold*theta + 0.45*fbmVal)
					foldRadial := math.Sin(mFold*0.5*theta - 0.35*fbmVal)
					gxCoord := (rn*math.Cos(theta)*0.55 + 0.45*rn*foldAng) * 4.2
					gyCoord := (rn*math.Sin(theta)*0.55 + 0.45*rn*foldRadial) * 4.2

					g1 := ne.GaborNoise2D(
						gxCoord+loopX*1.6,
						gyCoord+loopY*1.6,
						g.gaborFreq,
						flowOrient,
						g.bandwidth,
					)
					g2 := ne.GaborNoise2D(
						ux*5.4-loopY*1.4+1.7,
						uy*5.4+loopX*1.4-1.3,
						g.gaborFreq*1.25,
						theta+0.25*math.Pi*foldAng,
						g.bandwidth*1.15,
					)

					// Dihedral rosette Chladni-Gabor standing wave modes
					radialFreq := (3.2 + 0.45*g.gaborFreq) * math.Pi
					chladniA := math.Cos(mFold*theta+1.15*g1+0.55*fbmVal) * math.Cos(radialFreq*rn-0.85*g2+phi)
					chladniB := math.Sin(2.0*mFold*theta-1.05*g2+0.40*fbmVal) * math.Sin(radialFreq*0.68*rn+0.95*g1-phi)
					ringHarmonic := math.Cos((radialFreq*1.45)*rn + 1.35*g1 + 0.65*fbmVal - phi)

					waveField := 0.54*chladniA + 0.30*chladniB + 0.16*ringHarmonic

					// Razor-sharp caustic ridge extraction from wave zero-crossings and envelopes
					primaryRidge := 1.0 / (1.0 + 38.0*waveField*waveField)
					primaryRidge = math.Pow(primaryRidge, 1.55)

					secondaryCarrier := math.Sin((mFold*theta + radialFreq*1.15*rn) + 1.85*g1 + 1.15*g2 - phi)
					secondaryRidge := 1.0 / (1.0 + 52.0*secondaryCarrier*secondaryCarrier)
					secondaryRidge = math.Pow(secondaryRidge, 1.75)

					// Nodal intersection hotspots where Chladni modes cross
					nodalCross := math.Exp(-24.0 * (chladniA*chladniA + chladniB*chladniB))

					// Radial framing modulation (luminous core + crisp inner bezel shadow)
					coreGlow := math.Exp(-11.5 * rn * rn)
					rimShadow := common.Smoothstep(0.0, 0.08, 1.0-rn)
					causticEnergy := common.Clamp01((0.76*primaryRidge + 0.42*secondaryRidge + 0.55*nodalCross + 0.32*coreGlow) * rimShadow)

					// Airy thin-film structural color along caustic ridges
					cosTheta := common.Clamp01(0.88 - 0.34*rn + 0.14*primaryRidge)
					filmNm := 165.0 + 235.0*(0.5+0.5*math.Sin(mFold*theta*0.5+2.6*rn+1.1*g1+0.7*fbmVal+phi)) + 45.0*gradMag
					tfR, tfG, tfB := render.SampleThinFilmLinearRGB(cosTheta, filmNm, 1.92)
					tfR = math.Pow(common.Clamp01(tfR), 1.45)
					tfG = math.Pow(common.Clamp01(tfG), 1.45)
					tfB = math.Pow(common.Clamp01(tfB), 1.45)

					// Palette harmonic blending between electric cyan, solar gold, and neon rose-magenta
					palPhase := 0.5 + 0.5*math.Sin(mFold*theta+3.8*rn+1.4*fbmVal+0.8*g2-phi)
					var palR, palG, palB float64
					if palPhase < 0.5 {
						u := palPhase * 2.0
						palR = cyaR + (gldR-cyaR)*u
						palG = cyaG + (gldG-cyaG)*u
						palB = cyaB + (gldB-cyaB)*u
					} else {
						u := (palPhase - 0.5) * 2.0
						palR = gldR + (magR-gldR)*u
						palG = gldG + (magG-gldG)*u
						palB = gldB + (magB-gldB)*u
					}

					ridgeR := 0.58*palR + 0.42*tfR
					ridgeG := 0.58*palG + 0.42*tfG
					ridgeB := 0.58*palB + 0.42*tfB

					// Deep abyssal sapphire troughs with high-contrast caustic ridge illumination
					troughDepth := 0.35 + 0.65*(0.5+0.5*math.Cos(mFold*theta))*rimShadow
					baseR := obsR + (sapR-obsR)*troughDepth
					baseG := obsG + (sapG-obsG)*troughDepth
					baseB := obsB + (sapB-obsB)*troughDepth

					hotCore := math.Pow(common.Clamp01(primaryRidge*0.85+nodalCross*0.65+coreGlow*0.5), 3.4)
					colR := baseR*(1.0-0.60*causticEnergy) + ridgeR*(0.68*causticEnergy) + ivoR*(0.24*hotCore)
					colG := baseG*(1.0-0.60*causticEnergy) + ridgeG*(0.68*causticEnergy) + ivoG*(0.24*hotCore)
					colB := baseB*(1.0-0.60*causticEnergy) + ridgeB*(0.68*causticEnergy) + ivoB*(0.24*hotCore)

					// Blend smoothly against obsidian-midnight exterior at the medallion perimeter
					finalR := obsR + (common.Clamp01(colR)-obsR)*edgeCov
					finalG := obsG + (common.Clamp01(colG)-obsG)*edgeCov
					finalB := obsB + (common.Clamp01(colB)-obsB)*edgeCov

					rc.SetPixelLinear(x, y, finalR, finalG, finalB, 1.0)
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	// 2. Additive Calligraphic Curl Caustic Filaments inside the Medallion
	rc.PushClipCircle(cx, cy, medallionR-1.5*scale)
	rc.SetBlendMode(render.BlendAdd)

	const numFilaments = 900
	const stepsPerFilament = 54
	stepLen := 2.35 * scale
	goldenAngle := math.Pi * (3.0 - math.Sqrt(5.0))

	for i := 0; i < numFilaments; i++ {
		fi := float64(i)
		// Seed particles across rosette rings and phyllotaxis spiral inside the medallion
		u := (fi + 0.5) / float64(numFilaments)
		seedR := (0.06 + 0.88*math.Sqrt(u)) * medallionR
		seedAng := fi*goldenAngle + 0.28*math.Sin(fi*0.17)

		px := cx + seedR*math.Cos(seedAng)
		py := cy + seedR*math.Sin(seedAng)

		var filCol color.RGBA
		switch i % 4 {
		case 0:
			filCol = solarGold
		case 1:
			filCol = electricCyan
		case 2:
			filCol = ivoryPearl
		default:
			filCol = neonRoseMagenta
		}

		baseAlpha := 0.032 + 0.034*(0.5+0.5*math.Sin(fi*0.73))
		pts := make([]render.Point, 0, stepsPerFilament)

		for s := 0; s < stepsPerFilament; s++ {
			dx := px - cx
			dy := py - cy
			r := math.Hypot(dx, dy)
			if r >= medallionR-2.0*scale {
				break
			}
			pts = append(pts, render.Pt(px, py))

			ux := dx * invR
			uy := dy * invR
			rn := r * invR
			theta := math.Atan2(uy, ux)

			// Analytical divergence-free curl velocity + dihedral radial harmonic guidance
			curlVec := ne.Curl2DAnalytical(ux*2.6+loopX, uy*2.6+loopY, 3)
			cvx, cvy := curlVec.X, curlVec.Y
			curlNorm := math.Hypot(cvx, cvy) + 1e-6
			cvx /= curlNorm
			cvy /= curlNorm

			// Tangential swirl + m-fold rosette caustic attractor
			rosettePull := math.Sin(mFold*theta+phi) * math.Cos(4.0*math.Pi*rn)
			tx := -math.Sin(theta)*0.65 + math.Cos(theta)*0.35*rosettePull
			ty := math.Cos(theta)*0.65 + math.Sin(theta)*0.35*rosettePull

			vx := 0.58*cvx + 0.42*tx
			vy := 0.58*cvy + 0.42*ty
			vNorm := math.Hypot(vx, vy) + 1e-6
			px += (vx / vNorm) * stepLen
			py += (vy / vNorm) * stepLen

			if s%6 == 0 {
				env := math.Sin(math.Pi * float64(s+1) / float64(stepsPerFilament))
				rc.SplatPoint(px, py, filCol, baseAlpha*0.75*env)
			}
		}

		if len(pts) >= 3 {
			rc.StrokePolyline(pts, 0.68*scale, filCol, baseAlpha)
		}
	}

	// Subtle inner harmonic astrolabe latitude circles & meridian arcs inside the medallion
	for _, frac := range []float64{0.28, 0.54, 0.78} {
		rc.DrawSoftRing(cx, cy, medallionR*frac, 0.85*scale, solarGold, 0.20)
	}
	for p := 0; p < numPoles*2; p++ {
		ang := float64(p) * math.Pi / float64(numPoles)
		r0 := medallionR * 0.14
		r1 := medallionR * 0.96
		rc.StrokePolyline(
			[]render.Point{
				render.Pt(cx+r0*math.Cos(ang), cy+r0*math.Sin(ang)),
				render.Pt(cx+r1*math.Cos(ang), cy+r1*math.Sin(ang)),
			},
			0.75*scale, solarGold, 0.16,
		)
	}
	// Central celestial astrolabe sun-core
	rc.DrawSoftRing(cx, cy, medallionR*0.14, 1.3*scale, electricCyan, 0.52)
	rc.DrawGlowDot(cx, cy, 4.5*scale, ivoryPearl, 0.95, 0.75)

	rc.PopClip()

	// 3. Double-Ringed 24k Burnished Gold Astrolabe Bezel, Degree Ticks, Cardinal Stars & Orbital Jewels
	rc.SetBlendMode(render.BlendOver)
	innerRingR := medallionR
	outerRingR := medallionR * 1.075
	midRingR := 0.5 * (innerRingR + outerRingR)
	outerFiligreeR := medallionR * 1.115

	// Darkened metallic channel between the double astrolabe rings for crisp vernacular contrast
	rc.DrawSoftRing(cx, cy, midRingR, (outerRingR-innerRingR)*0.95, obsidianMidnight, 0.88)

	// Degree ticks & Vernier graduations around the astrolabe bezel
	const totalTicks = 144
	for t := 0; t < totalTicks; t++ {
		ang := (float64(t) + 12.0*phase) * 2.0 * math.Pi / float64(totalTicks)
		cosA, sinA := math.Cos(ang), math.Sin(ang)

		isMajor := t%12 == 0
		isMedium := t%4 == 0
		var rStart, rEnd, tickW, tickAlpha float64
		var tickCol color.RGBA

		if isMajor {
			rStart = innerRingR
			rEnd = outerFiligreeR
			tickW = 1.45 * scale
			tickAlpha = 0.88
			tickCol = ivoryPearl
		} else if isMedium {
			rStart = innerRingR + (outerRingR-innerRingR)*0.15
			rEnd = outerRingR
			tickW = 1.0 * scale
			tickAlpha = 0.68
			tickCol = solarGold
		} else {
			rStart = innerRingR + (outerRingR-innerRingR)*0.32
			rEnd = outerRingR - (outerRingR-innerRingR)*0.12
			tickW = 0.65 * scale
			tickAlpha = 0.44
			tickCol = antiqueBrass
		}

		rc.StrokePolyline(
			[]render.Point{
				render.Pt(cx+rStart*cosA, cy+rStart*sinA),
				render.Pt(cx+rEnd*cosA, cy+rEnd*sinA),
			},
			tickW, tickCol, tickAlpha,
		)
	}

	// Crisp double-ringed 24k burnished gold bezel rims
	rc.DrawSoftRing(cx, cy, innerRingR, 2.2*scale, antiqueBrass, 0.92)
	rc.DrawSoftRing(cx, cy, innerRingR, 1.1*scale, solarGold, 0.96)
	rc.DrawSoftRing(cx, cy, outerRingR, 2.0*scale, antiqueBrass, 0.90)
	rc.DrawSoftRing(cx, cy, outerRingR, 1.0*scale, solarGold, 0.94)
	rc.DrawSoftRing(cx, cy, outerFiligreeR, 0.75*scale, antiqueBrass, 0.52)

	// 4 Cardinal Compass Stars at N, E, S, W on the astrolabe bezel
	for k := 0; k < 4; k++ {
		ang := float64(k) * 0.5 * math.Pi
		cosA, sinA := math.Cos(ang), math.Sin(ang)
		px, py := -sinA, cosA

		starCx := cx + outerRingR*cosA
		starCy := cy + outerRingR*sinA
		tipOuter := 11.5 * scale
		tipInner := 8.5 * scale
		halfBase := 3.2 * scale

		starPoly := []render.Point{
			render.Pt(starCx+tipOuter*cosA, starCy+tipOuter*sinA),
			render.Pt(starCx+halfBase*px, starCy+halfBase*py),
			render.Pt(starCx-tipInner*cosA, starCy-tipInner*sinA),
			render.Pt(starCx-halfBase*px, starCy-halfBase*py),
		}
		rc.FillPolygon(starPoly, solarGold, 0.94)
		rc.StrokePolyline(append(starPoly, starPoly[0]), 0.85*scale, ivoryPearl, 0.88)
		rc.DrawGlowDot(starCx, starCy, 2.8*scale, ivoryPearl, 0.98, 0.75)
	}

	// vortexPoles Orbital Jewel Nodes around the perimeter
	rc.SetBlendMode(render.BlendAdd)
	for p := 0; p < numPoles; p++ {
		ang := (float64(p)+0.5+2.0*phase)*2.0*math.Pi/float64(numPoles) + 0.04*math.Sin(phi)
		jx := cx + midRingR*math.Cos(ang)
		jy := cy + midRingR*math.Sin(ang)

		jewelCol := electricCyan
		if p%2 == 1 {
			jewelCol = neonRoseMagenta
		}
		rc.DrawSoftRing(jx, jy, 6.2*scale, 1.2*scale, solarGold, 0.78)
		rc.DrawSoftRing(jx, jy, 3.8*scale, 1.0*scale, jewelCol, 0.85)
		rc.DrawGlowDot(jx, jy, 2.9*scale, ivoryPearl, 0.96, 0.82)
	}

	// 4. HDR Post-Processing & Merge
	rc.ApplyBloom(0.78, 4.0*scale, 0.18)
	rc.ApplyVignette(0.24)
	c.MergeRenderCanvas(rc)
}
