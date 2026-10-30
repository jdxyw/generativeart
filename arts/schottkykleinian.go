package arts

import (
	"image/color"
	"math"
	"math/cmplx"
	"runtime"
	"sort"
	"sync"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type schottkyKleinian struct {
	maxDepth        int
	loxodromicTwist float64
}

// NewSchottkyKleinian creates a SchottkyKleinian generator that renders an "Indra's Pearls"
// Kissing-Schottky Kleinian limit set, 3D iridescent orbit-trap pearls, a holomorphic
// domain-coloring rose window, and a Poincaré geodesic astrolabe filigree.
func NewSchottkyKleinian(maxDepth int, loxodromicTwist float64) *schottkyKleinian {
	if maxDepth <= 0 {
		maxDepth = 7
	}
	if loxodromicTwist <= 0 {
		loxodromicTwist = 1.0
	}
	return &schottkyKleinian{
		maxDepth:        maxDepth,
		loxodromicTwist: loxodromicTwist,
	}
}

type schottkyGenCircle struct {
	circle common.Circle2D
	idx    int
	depth  int
}

// Generative renders the static SchottkyKleinian composition.
func (sk *schottkyKleinian) Generative(c *generativeart.Canva) {
	sk.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360-degree phase loop (phase in [0, 1)).
func (sk *schottkyKleinian) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
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
	cx, cy := 0.5*fw, 0.5*fh
	worldExtent := 1.25

	toScreen := func(z complex128) render.Point {
		return render.Pt(
			cx+(real(z)/worldExtent)*(0.5*minDim),
			cy+(imag(z)/worldExtent)*(0.5*minDim),
		)
	}
	toScreenRadius := func(r float64) float64 {
		return (r / worldExtent) * (0.5 * minDim)
	}

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	phi := 2.0 * math.Pi * phase
	twist := sk.loxodromicTwist

	// 1. Construct the global Mobius2D DiskAutomorphism + loxodromic spiral twist.
	// When static (totalFrames <= 1 and phase == 0), apply a subtle off-center jewel tilt;
	// when animated, orbit the automorphism pole smoothly around a closed circle for zero-flicker looping.
	autoMag := 0.11 * math.Tanh(0.65*twist)
	autoAngle := 0.38*twist + phi
	autoA := cmplx.Rect(autoMag, autoAngle)
	rotTheta := 0.16*twist + phi/8.0
	diskMob := common.DiskAutomorphism(autoA, rotTheta).Normalized()

	// Loxodromic scaling/spiral modulation around the origin composed with the disk automorphism
	loxoSpiral := cmplx.Rect(1.0, 0.08*twist*math.Sin(phi))
	loxoMob := common.NewMobius2D(loxoSpiral, 0, 0, cmplx.Conj(loxoSpiral)).Normalized()
	globalMob := diskMob.Compose(loxoMob).Normalized()
	invGlobalMob := globalMob.Inverse().Normalized()

	// 2. Construct the symmetric necklace of N=8 kissing generator circles orthogonal to |z|=1,
	// plus an inner/outer concentric inversion pair and the conjugate dual orthogonal necklace.
	const numNecklace = 8
	dTheta := (2.0 * math.Pi) / float64(numNecklace)
	halfTheta := math.Pi / float64(numNecklace)

	// Exact kissing circles orthogonal to the unit circle:
	// R_0 = sec(pi/N), r_0 = tan(pi/N), satisfying R_0^2 - r_0^2 = 1 and r_0 = R_0 * sin(pi/N).
	ringDist := 1.0 / math.Cos(halfTheta)
	ringRad := math.Tan(halfTheta)

	// Inner and outer capping circles tangent to the necklace circles:
	// Inner circle radius r_in = R_0 - r_0, outer circle radius r_out = R_0 + r_0 (inverted inside out via a Kelvin sphere).
	rInner := ringDist - ringRad
	rOuter := ringDist + ringRad

	generators := make([]common.Circle2D, 0, numNecklace+2)
	for k := 0; k < numNecklace; k++ {
		ang := float64(k) * dTheta
		generators = append(generators, common.Circle2D{
			Center: cmplx.Rect(ringDist, ang),
			Radius: ringRad,
		})
	}
	// Inner central generator circle
	generators = append(generators, common.Circle2D{
		Center: 0,
		Radius: rInner * 0.985,
	})

	// Conjugate dual necklace circles shifted by pi/N (centered at the kissing points of the primary necklace)
	// plus the unit Poincaré horizon circle |z| = 1 and harmonic inner/outer pearls for 3D orbit-trap shading.
	dualCircles := make([]common.Circle2D, 0, numNecklace+3)
	for k := 0; k < numNecklace; k++ {
		ang := (float64(k) + 0.5) * dTheta
		dualCircles = append(dualCircles, common.Circle2D{
			Center: cmplx.Rect(ringDist, ang),
			Radius: ringRad * 0.92,
		})
	}
	dualCircles = append(dualCircles,
		common.Circle2D{Center: 0, Radius: 1.0},
		common.Circle2D{Center: 0, Radius: rInner},
		common.Circle2D{Center: 0, Radius: rOuter},
	)

	// Color palette for domain coloring and jewel shading (Oklab / Linear RGB)
	domainGrad := render.NewMultiStopGradient(
		color.RGBA{R: 12, G: 28, B: 68, A: 255},   // Abyssal Sapphire
		color.RGBA{R: 22, G: 138, B: 168, A: 255}, // Imperial Cyan
		color.RGBA{R: 42, G: 176, B: 136, A: 255}, // Celadon Turquoise
		color.RGBA{R: 238, G: 184, B: 58, A: 255}, // Burnished Gold
		color.RGBA{R: 198, G: 52, B: 68, A: 255},  // Cinnabar Ruby
		color.RGBA{R: 112, G: 32, B: 118, A: 255}, // Royal Amethyst
		color.RGBA{R: 12, G: 28, B: 68, A: 255},   // Seamless wrap to Abyssal Sapphire
	)

	maxSteps := sk.maxDepth * 3
	if maxSteps < 12 {
		maxSteps = 12
	}

	// 3D directional lighting vectors for Blinn-Phong pearl shading
	lx, ly, lz := -0.58, -0.62, 0.52839
	hx, hy, hz := lx, ly, lz+1.0
	hNorm := math.Sqrt(hx*hx + hy*hy + hz*hz)
	hx, hy, hz = hx/hNorm, hy/hNorm, hz/hNorm

	// 3. Pixel Shader Pass (2x2 SSAA per pixel, parallelized across rows)
	subOffsets := [4][2]float64{
		{0.25, 0.25},
		{0.75, 0.25},
		{0.25, 0.75},
		{0.75, 0.75},
	}

	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
	if workers > h {
		workers = h
	}
	var wg sync.WaitGroup
	chunk := (h + workers - 1) / workers

	for workerID := 0; workerID < workers; workerID++ {
		y0 := workerID * chunk
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
				for x := 0; x < w; x++ {
					var accR, accG, accB float64

					for s := 0; s < 4; s++ {
						sx := (float64(x)+subOffsets[s][0]-cx)/(0.5*minDim) * worldExtent
						sy := (float64(y)+subOffsets[s][1]-cy)/(0.5*minDim) * worldExtent
						z0 := complex(sx, sy)
						r0 := cmplx.Abs(z0)

						// Apply global Mobius2D DiskAutomorphism & loxodromic twist
						z := globalMob.Apply(z0)

						// Kelvin inversion outside the outer bounding sphere so exterior space also packs with pearls
						invertedOuter := 0
						if cmplx.Abs(z) > rOuter {
							outerC := common.Circle2D{Center: 0, Radius: rOuter}
							z = outerC.InvertPoint(z)
							invertedOuter = 1
						}

						iter := 0
						lastGen := -1
						derivScale := 1.0

						// Track minimum orbit-trap distance to the dual orthogonal necklace circles
						minTrapDist := 1e9
						trapNormalZ := complex(0, 0)
						trapIndex := 0

						evalOrbitTrap := func(zp complex128, scaleF float64, stepIdx int) {
							for dIdx := 0; dIdx < numNecklace; dIdx++ {
								dc := dualCircles[dIdx]
								diff := zp - dc.Center
								dist := cmplx.Abs(diff) - dc.Radius
								// Signed/unsigned distance normalized by circle radius for 3D pearl hemisphere profile
								uDist := dist / dc.Radius
								if uDist < minTrapDist {
									minTrapDist = uDist
									trapNormalZ = diff / complex(dc.Radius, 0)
									trapIndex = dIdx + stepIdx*3
								}
							}
						}

						evalOrbitTrap(z, derivScale, 0)

						for step := 0; step < maxSteps; step++ {
							bestIdx := -1
							bestSD := 0.0
							for gIdx, gen := range generators {
								if gIdx == lastGen {
									continue
								}
								sd := gen.SignedDist(z)
								if sd < bestSD {
									bestSD = sd
									bestIdx = gIdx
								}
							}
							if bestIdx < 0 {
								break
							}
							gen := generators[bestIdx]
							diff := z - gen.Center
							distSq := real(diff)*real(diff) + imag(diff)*imag(diff)
							if distSq < 1e-18 {
								distSq = 1e-18
							}
							r2 := gen.Radius * gen.Radius
							derivScale *= r2 / distSq

							z = gen.InvertPoint(z)
							lastGen = bestIdx
							iter++

							evalOrbitTrap(z, derivScale, iter)
						}

						// Nearest boundary distance in the final fundamental domain for limit-set filament glow
						minBoundaryDist := 1e9
						for _, gen := range generators {
							bd := math.Abs(gen.SignedDist(z)) / math.Sqrt(math.Max(derivScale, 1e-12))
							if bd < minBoundaryDist {
								minBoundaryDist = bd
							}
						}

						// Evaluate holomorphic modular rational rose function f(z_final)
						// Invariant 8-fold dihedral rational function with zeros & poles at necklace symmetries
						z4 := z * z * z * z
						z8 := z4 * z4
						twistPhase := cmplx.Rect(1.0, 0.35*twist+phi)
						num := z * (z8 - twistPhase*complex(0.82, 0))
						den := complex(0.82, 0)*z8 - cmplx.Conj(twistPhase)
						if cmplx.Abs(den) < 1e-12 {
							den = complex(1e-12, 0)
						}
						fz := num / den

						argNorm, logModIso, phaseRayIso, domBright := common.EvaluateDomainColoring(fz, 0.45, 12.0)

						// Map domain coloring to deep stained-glass jewel tones
						hueT := math.Mod(argNorm+0.07*float64((iter+invertedOuter)%2), 1.0)
						if hueT < 0 {
							hueT += 1.0
						}
						domCol := domainGrad.SampleOklab(hueT)
						dR, dG, dB, _ := render.ColorToLinearRGB(domCol)

						// Modulate fundamental domain by stereographic brightness & parity
						parityBoost := 0.82 + 0.24*float64((iter+invertedOuter)%2)
						depthAtten := 1.0 / (1.0 + 0.065*float64(iter))
						baseShade := (0.20 + 0.72*domBright) * parityBoost * depthAtten

						rLin := dR * baseShade
						gLin := dG * baseShade
						bLin := dB * baseShade

						// Crisp anti-aliased log-modulus & phase-ray conformal grid isolines
						isoStrength := math.Max(logModIso, phaseRayIso*0.85)
						if isoStrength > 0.001 {
							// Fine golden-cyan conformal mesh lines with dark cloisonné border
							isoCore := common.Smoothstep(0.35, 0.95, isoStrength)
							isoEdge := isoStrength * (1.0 - isoCore)
							rLin *= 1.0 - 0.55*isoEdge
							gLin *= 1.0 - 0.55*isoEdge
							bLin *= 1.0 - 0.55*isoEdge

							goldR, goldG, goldB := 0.88, 0.64, 0.18
							if (iter+invertedOuter)%2 == 1 {
								goldR, goldG, goldB = 0.22, 0.78, 0.92
							}
							rLin += goldR * isoCore * 0.48 * depthAtten
							gLin += goldG * isoCore * 0.48 * depthAtten
							bLin += goldB * isoCore * 0.48 * depthAtten
						}

						// 3D Glassy/Metallic "Indra's Pearls" shading when orbit-trap falls inside a dual pearl
						if minTrapDist < 0.0 {
							// Reconstruct 3D sphere normal (nx, ny, nz) from normalized complex offset inside the trapped circle
							nx := real(trapNormalZ)
							ny := imag(trapNormalZ)
							rSq := nx*nx + ny*ny
							if rSq < 1.0 {
								nz := math.Sqrt(math.Max(0.0, 1.0-rSq))
								// Smooth anti-aliased coverage at pearl edge
								pearlCov := common.Smoothstep(0.0, 0.06, -minTrapDist)

								// Cosine viewing angle cosTheta = N . V = nz
								cosTheta := common.Constrain(nz, 0.0, 1.0)
								fresnel := math.Pow(1.0-cosTheta, 3.2)

								// Thin-film optical path thickness modulated by iteration depth, trap index, and normal curvature
								filmNm := 280.0 + 195.0*cosTheta + 52.0*float64((iter+trapIndex)%7) + 35.0*math.Sin(phi+float64(iter)*0.7)
								tfR, tfG, tfB := render.SampleThinFilmLinearRGB(cosTheta, filmNm, 1.42)

								// Blinn-Phong diffuse & specular highlights
								ndotl := math.Max(0.0, nx*lx+ny*ly+nz*lz)
								ndoth := math.Max(0.0, nx*hx+ny*hy+nz*hz)
								spec := math.Pow(ndoth, 48.0)*1.35 + math.Pow(ndoth, 12.0)*0.28

								// Secondary warm rim light from bottom-right
								ndotl2 := math.Max(0.0, nx*0.65+ny*0.58+nz*0.48)

								// Combine deep metallic pearl body, thin-film iridescence, Fresnel rim glow, and glassy refraction of domain coloring
								diff := 0.16 + 0.78*ndotl + 0.25*ndotl2
								iridMix := 0.45 + 0.50*fresnel
								bodyR := dR*0.42 + tfR*iridMix
								bodyG := dG*0.42 + tfG*iridMix
								bodyB := dB*0.42 + tfB*iridMix

								pR := bodyR*diff + (0.95*tfR+0.25)*fresnel*0.62 + spec*1.08 + rLin*0.22*(1.0-fresnel)
								pG := bodyG*diff + (0.95*tfG+0.22)*fresnel*0.62 + spec*1.02 + gLin*0.22*(1.0-fresnel)
								pB := bodyB*diff + (0.95*tfB+0.30)*fresnel*0.62 + spec*0.94 + bLin*0.22*(1.0-fresnel)

								// Ambient occlusion collar right around the pearl rim
								ao := 0.55 + 0.45*common.Smoothstep(0.0, 0.22, nz)
								pR *= ao
								pG *= ao
								pB *= ao

								rLin = rLin*(1.0-pearlCov) + pR*pearlCov
								gLin = gLin*(1.0-pearlCov) + pG*pearlCov
								bLin = bLin*(1.0-pearlCov) + pB*pearlCov
							}
						} else if minTrapDist < 0.08 {
							// Subtle drop-shadow halo immediately outside each pearl
							halo := 1.0 - 0.28*(1.0-minTrapDist/0.08)
							rLin *= halo
							gLin *= halo
							bLin *= halo
						}

						// Luminous Kleinian limit-set dust & isometric sphere boundary filaments
						limitGlow := math.Exp(-minBoundaryDist * 115.0)
						if iter >= 3 {
							limitGlow += 0.38 * math.Exp(-minBoundaryDist*42.0) * math.Min(1.0, float64(iter)/float64(sk.maxDepth))
						}
						if limitGlow > 1e-3 {
							tfR, tfG, tfB := render.SampleThinFilmLinearRGB(
								0.65,
								340.0+65.0*float64(iter)+40.0*math.Cos(phi+real(z)*3.0),
								1.38,
							)
							glowR := 0.55*tfR + 0.45*0.98
							glowG := 0.55*tfG + 0.45*0.76
							glowB := 0.55*tfB + 0.45*0.32
							rLin += glowR * limitGlow * 0.65
							gLin += glowG * limitGlow * 0.65
							bLin += glowB * limitGlow * 0.65
						}

						// Subtle exterior astrolabe darkening outside |z0| > 1.14
						if r0 > 1.12 {
							extFade := common.Smoothstep(1.12, 1.32, r0)
							rLin = rLin*(1.0-0.55*extFade) + 0.012*extFade
							gLin = gLin*(1.0-0.55*extFade) + 0.018*extFade
							bLin = bLin*(1.0-0.55*extFade) + 0.038*extFade
						}

						accR += rLin
						accG += gLin
						accB += bLin
					}

					rc.SetPixelLinear(x, y, accR*0.25, accG*0.25, accB*0.25, 1.0)
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	// 4. Vector / Geometric Filigree Pass on rc
	// Transform a Circle2D through the inverse global Mobius transformation via 3 boundary points
	transformCircleToWorld := func(c2d common.Circle2D) (common.Circle2D, bool) {
		p0 := invGlobalMob.Apply(c2d.Center + complex(c2d.Radius, 0))
		p1 := invGlobalMob.Apply(c2d.Center + complex(-0.5*c2d.Radius, 0.86602540378*c2d.Radius))
		p2 := invGlobalMob.Apply(c2d.Center + complex(-0.5*c2d.Radius, -0.86602540378*c2d.Radius))
		scx, scy, sr, ok := circleFrom3Complex(p0, p1, p2)
		if !ok {
			return common.Circle2D{}, false
		}
		return common.Circle2D{Center: complex(scx, scy), Radius: sr}, true
	}

	// 4a. Recursively generate the exact tree of nested Schottky isometric spheres up to depth 4
	// using common.Circle2D.InvertCircle
	const maxTreeDepth = 4
	nestedCircles := make([]schottkyGenCircle, 0, 512)

	var buildCircleTree func(cur common.Circle2D, lastIdx, depth int)
	buildCircleTree = func(cur common.Circle2D, lastIdx, depth int) {
		mapped, ok := transformCircleToWorld(cur)
		if !ok {
			return
		}
		scrR := toScreenRadius(mapped.Radius)
		if scrR < 1.15*scale || scrR > 1.8*minDim {
			return
		}
		nestedCircles = append(nestedCircles, schottkyGenCircle{
			circle: mapped,
			idx:    lastIdx,
			depth:  depth,
		})
		if depth >= maxTreeDepth {
			return
		}
		for gIdx, gen := range generators {
			if gIdx == lastIdx {
				continue
			}
			invC := gen.InvertCircle(cur)
			buildCircleTree(invC, gIdx, depth+1)
		}
	}

	for gIdx, gen := range generators {
		buildCircleTree(gen, gIdx, 1)
	}

	// Sort circles from largest to smallest so finer nested rings layer cleanly on top
	sort.Slice(nestedCircles, func(i, j int) bool {
		return nestedCircles[i].circle.Radius > nestedCircles[j].circle.Radius
	})

	brassBase := color.RGBA{R: 168, G: 118, B: 34, A: 255}
	goldRim := color.RGBA{R: 246, G: 204, B: 88, A: 255}
	ivorySpec := color.RGBA{R: 255, G: 250, B: 230, A: 255}
	cyanFiligree := color.RGBA{R: 86, G: 224, B: 242, A: 255}
	shadowCol := color.RGBA{R: 4, G: 8, B: 20, A: 255}
	lightDir2D := render.Pt(-0.707, -0.707)

	// Build a composite SDF for the primary depth-1 Schottky generator rings and unit astrolabe ring
	// so we can render an analytical 2.5D gilded bevel & drop shadow with RenderSDFDropShadow + RenderSDFBevel
	var primaryRingsSDF render.SDF2D
	unitMapped, okUnit := transformCircleToWorld(common.Circle2D{Center: 0, Radius: 1.0})
	if okUnit {
		up := toScreen(unitMapped.Center)
		ur := toScreenRadius(unitMapped.Radius)
		primaryRingsSDF = render.SDFOnion(render.SDFCircle(up, ur), 2.1*scale)
	}
	for _, nc := range nestedCircles {
		if nc.depth != 1 {
			continue
		}
		cp := toScreen(nc.circle.Center)
		cr := toScreenRadius(nc.circle.Radius)
		ringSDF := render.SDFOnion(render.SDFCircle(cp, cr), 1.45*scale)
		if primaryRingsSDF == nil {
			primaryRingsSDF = ringSDF
		} else {
			primaryRingsSDF = render.SDFUnion(primaryRingsSDF, ringSDF)
		}
	}

	if primaryRingsSDF != nil {
		rc.RenderSDFDropShadow(primaryRingsSDF, render.Pt(2.2*scale, 2.6*scale), 4.5*scale, 0.5*scale, shadowCol, 0.52)
		rc.RenderSDFBevel(primaryRingsSDF, brassBase, goldRim, ivorySpec, 2.8*scale, lightDir2D, 0.92)
	}

	// Stroke nested depth 2..4 Schottky isometric spheres with fine gilded brass/gold & cyan bevel rings
	for d := 2; d <= maxTreeDepth; d++ {
		pathGold := &canvas.Path{}
		pathHighlight := &canvas.Path{}
		count := 0
		for _, nc := range nestedCircles {
			if nc.depth != d {
				continue
			}
			cp := toScreen(nc.circle.Center)
			cr := toScreenRadius(nc.circle.Radius)
			if cr < 1.2*scale {
				continue
			}
			subPath := canvas.Circle(cr)
			pathGold.Append(subPath.Translate(cp.X, cp.Y))
			if cr >= 4.0*scale {
				hlPath := canvas.Circle(cr * 0.965)
				pathHighlight.Append(hlPath.Translate(cp.X-0.35*scale, cp.Y-0.35*scale))
			}
			count++
		}
		if count > 0 {
			strokeW := math.Max(0.55, 1.55-0.28*float64(d)) * scale
			alpha := math.Max(0.38, 0.85-0.12*float64(d))
			ringCol := render.LerpOklab(goldRim, cyanFiligree, float64(d-2)/float64(maxTreeDepth-1)*0.45)
			rc.StrokePath(pathGold, strokeW, canvas.RoundCap, canvas.RoundJoin, ringCol, alpha)
			rc.StrokePath(pathHighlight, strokeW*0.45, canvas.RoundCap, canvas.RoundJoin, ivorySpec, alpha*0.55)
		}
	}

	// 4b. Hyperbolic Poincaré Geodesic Astrolabe Arcs connecting kissing points around the unit disk
	// Kissing points of the N=8 necklace lie on the unit circle at angles (k + 0.5) * dTheta
	kissingPts := make([]complex128, numNecklace)
	for k := 0; k < numNecklace; k++ {
		ang := (float64(k) + 0.5) * dTheta
		kissingPts[k] = cmplx.Rect(0.996, ang)
	}

	goldCyanGrad := func(t float64) color.RGBA {
		wave := 0.5 - 0.5*math.Cos(2.0*math.Pi*t)
		col := render.LerpOklab(goldRim, cyanFiligree, wave*0.65)
		return render.LerpOklab(col, ivorySpec, math.Pow(math.Sin(math.Pi*t), 3.0)*0.45)
	}
	cyanGoldGrad := func(t float64) color.RGBA {
		wave := 0.5 - 0.5*math.Cos(2.0*math.Pi*t)
		col := render.LerpOklab(cyanFiligree, goldRim, wave*0.75)
		return render.LerpOklab(col, ivorySpec, math.Pow(math.Sin(math.Pi*t), 2.0)*0.40)
	}

	drawGeodesicSpline := func(z1, z2 complex128, maxW float64, colFn func(float64) color.RGBA, alpha float64) {
		arcVecs := common.PoincareGeodesicArc(z1, z2, 40)
		pts := make([]render.Point, len(arcVecs))
		for i, v := range arcVecs {
			zw := invGlobalMob.Apply(complex(v.X, v.Y))
			pts[i] = toScreen(zw)
		}
		rc.DrawTaperedSpline(pts, maxW, render.TaperBothEnds, colFn, alpha)
	}

	// Star-octagram and adjacent hyperbolic geodesics across the Poincaré disk
	for k := 0; k < numNecklace; k++ {
		// Step-1 adjacent kissing-point geodesics (forming the hyperbolic ideal octagon)
		zA := kissingPts[k]
		zB := kissingPts[(k+1)%numNecklace]
		drawGeodesicSpline(zA, zB, 2.5*scale, goldCyanGrad, 0.88)

		// Step-2 octagram chords
		zC := kissingPts[(k+2)%numNecklace]
		drawGeodesicSpline(zA, zC, 1.85*scale, cyanGoldGrad, 0.76)

		// Step-3 deep star-astrolabe hyperbolic geodesics
		zD := kissingPts[(k+3)%numNecklace]
		drawGeodesicSpline(zA, zD, 1.45*scale, goldCyanGrad, 0.68)
	}

	// Second-generation inner kissing points on the inner necklace ring
	innerKissPts := make([]complex128, numNecklace)
	for k := 0; k < numNecklace; k++ {
		ang := float64(k) * dTheta
		innerKissPts[k] = cmplx.Rect(rInner, ang)
	}
	for k := 0; k < numNecklace; k++ {
		drawGeodesicSpline(innerKissPts[k], innerKissPts[(k+3)%numNecklace], 1.25*scale, cyanGoldGrad, 0.64)
		drawGeodesicSpline(innerKissPts[k], kissingPts[k], 1.35*scale, goldCyanGrad, 0.72)
		drawGeodesicSpline(innerKissPts[(k+1)%numNecklace], kissingPts[k], 1.35*scale, goldCyanGrad, 0.72)
	}

	// Luminous jewel nodes at the primary kissing points and inner tangent vertices
	for k := 0; k < numNecklace; k++ {
		kp := toScreen(invGlobalMob.Apply(kissingPts[k]))
		rc.DrawGlowDot(kp.X, kp.Y, 3.8*scale, ivorySpec, 0.95, 0.85)
		rc.DrawSoftRing(kp.X, kp.Y, 5.6*scale, 1.1*scale, goldRim, 0.88)

		ip := toScreen(invGlobalMob.Apply(innerKissPts[k]))
		rc.DrawGlowDot(ip.X, ip.Y, 2.8*scale, cyanFiligree, 0.90, 0.80)
	}

	// 5. Composable HDR Post-Processing Pass
	rc.ApplyLuminanceRelief(0.42, -135.0, 45.0, 0.32)
	rc.ApplyChromaticAberration(0.55)
	rc.ApplyBloom(0.86, 3.5*scale, 0.10)
	rc.ApplyVignette(0.24)
	c.MergeRenderCanvas(rc)
}
