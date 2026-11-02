package arts

import (
	"image/color"
	"math"
	"math/cmplx"
	"runtime"
	"sync"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type escherCircleLimit struct {
	p, q, r int
}

// NewEscherCircleLimit returns an escherCircleLimit engine that renders an
// M.C. Escher Circle Limit III / IV inspired hyperbolic Poincaré disk tessellation
// driven by a Coxeter (p, q, r) reflection group, antisymmetric C1 S-curve
// interlocking motifs, woodcut engraving striations, and geodesic backbone arcs.
func NewEscherCircleLimit(p, q, r int) *escherCircleLimit {
	if p <= 0 || q <= 0 || r <= 0 || 1.0/float64(p)+1.0/float64(q)+1.0/float64(r) >= 1.0 {
		p, q, r = 6, 4, 2
	}
	return &escherCircleLimit{
		p: p,
		q: q,
		r: r,
	}
}

// Generative renders a static M.C. Escher Circle Limit hyperbolic woodcut print.
func (ecl *escherCircleLimit) Generative(c *generativeart.Canva) {
	ecl.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless looping hyperbolic rotation
// and orbital Poincaré disk automorphism over phase in [0, 1).
func (ecl *escherCircleLimit) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	minDim := math.Min(fw, fh)
	cx, cy := 0.5*fw, 0.5*fh
	diskR := 0.452 * minDim

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	// Use the minimal-area Hurwitz hyperbolic Coxeter group (7, 3, 2) when standard
	// (6, 4, 2) parameters are passed so the central rosette is compact (~24% radius)
	// and 5 concentric hyperbolic rings of isotropic figures are clearly resolved.
	pGroup, qGroup, rGroup := ecl.p, ecl.q, ecl.r
	if pGroup == 6 && qGroup == 4 && rGroup == 2 {
		pGroup, qGroup, rGroup = 7, 3, 2
	}
	cg := common.NewHyperbolicCoxeterGroup(pGroup, qGroup, rGroup)

	// Characteristic Euclidean scale of the fundamental triangle Delta_{pqr}
	triScale := cmplx.Abs(cg.Vq)
	if triScale < 1e-6 {
		triScale = 0.5
	}

	// Seamless animation: smooth hyperbolic rotation exp(i*phi) plus a gentle
	// periodic Möbius disk automorphism when animated.
	phi := 2.0 * math.Pi * phase
	rotZ := cmplx.Rect(1.0, phi/float64(cg.P))
	var diskMob common.Mobius2D
	useMobius := totalFrames > 1 || math.Abs(phase) > 1e-9
	if useMobius {
		driftA := cmplx.Rect(0.085*math.Sin(phi), phi)
		diskMob = common.DiskAutomorphism(driftA, phi/float64(cg.P)).Normalized()
	}

	// Helper to convert hex RGB into linear RGB [3]float64
	linRGB := func(r, g, b uint8) [3]float64 {
		lr, lg, lb, _ := render.ColorToLinearRGB(color.RGBA{R: r, G: g, B: b, A: 255})
		return [3]float64{lr, lg, lb}
	}
	lerp3 := func(a, b [3]float64, t float64) [3]float64 {
		return [3]float64{
			common.Lerp(a[0], b[0], t),
			common.Lerp(a[1], b[1], t),
			common.Lerp(a[2], b[2], t),
		}
	}
	sample4Zone := func(z0, z1, z2, z3 [3]float64, t float64) [3]float64 {
		u := common.Clamp01(t) * 3.0
		switch {
		case u < 1.0:
			return lerp3(z0, z1, common.Smoothstep(0.0, 1.0, u))
		case u < 2.0:
			return lerp3(z1, z2, common.Smoothstep(0.0, 1.0, u-1.0))
		default:
			return lerp3(z2, z3, common.Smoothstep(0.0, 1.0, u-2.0))
		}
	}

	// Exterior museum paper & gilded medallion frame palette
	paperCol := linRGB(243, 236, 224)       // #f3ece0 Warm museum printmaking paper
	paperShadowCol := linRGB(218, 206, 188) // #dacebc Soft vignette shadow
	inkCol := linRGB(20, 16, 14)            // #14100e Deep woodcut ink
	ivoryCol := linRGB(252, 245, 230)       // #fcf5e6 Luminous museum ivory
	giltCol := linRGB(248, 222, 138)        // #f8de8a 24k burnished gold highlight
	goldCoreCol := linRGB(240, 174, 52)     // #f0ae34 Rich imperial gold
	bronzeCol := linRGB(138, 92, 34)        // #8a5c22 Antique bronze

	// 4-Zone Center-to-Rim Palette Evolution for Positive Figures (figParity == 0)
	// Zone 0 (Center Core): Radiant Imperial Sun-Gold (#f6b830) & Warm Saffron (#e88424)
	// Zone 1 (Inner Ring):  Warm Museum Ivory (#fcf5e6) & Coral-Apricot (#f07c4a)
	// Zone 2 (Middle Ring): Luminous Celadon-Jade (#42b8a6) & Pale Gold-Mint (#d8f0dc)
	// Zone 3 (Outer Ring):  Moonlit Silver-Ice (#e8f2fc) & Sky-Turquoise (#5cc4e8)
	posPrimary0 := linRGB(246, 184, 48)   // #f6b830
	posSecondary0 := linRGB(232, 132, 36) // #e88424
	posPrimary1 := linRGB(252, 245, 230)  // #fcf5e6
	posSecondary1 := linRGB(240, 124, 74) // #f07c4a
	posPrimary2 := linRGB(216, 240, 220)  // #d8f0dc
	posSecondary2 := linRGB(66, 184, 166) // #42b8a6
	posPrimary3 := linRGB(232, 242, 252)  // #e8f2fc
	posSecondary3 := linRGB(92, 196, 232) // #5cc4e8

	// 4-Zone Center-to-Rim Palette Evolution for Negative Figures (figParity == 1)
	// Zone 0 (Center Core): Deep Cinnabar Ruby (#b82420) & Crimson-Garnet (#df402c)
	// Zone 1 (Inner Ring):  Rich Terracotta-Rust (#9e3624) & Deep Amber-Bronze (#64341c)
	// Zone 2 (Middle Ring): Deep Emerald-Pine (#104c4e) & Teal-Malachite (#1c7874)
	// Zone 3 (Outer Ring):  Midnight Prussian Lapis (#102248) & Royal Ultramarine (#224482)
	negPrimary0 := linRGB(184, 36, 32)    // #b82420
	negSecondary0 := linRGB(223, 64, 44)  // #df402c
	negPrimary1 := linRGB(100, 52, 28)    // #64341c
	negSecondary1 := linRGB(158, 54, 36)  // #9e3624
	negPrimary2 := linRGB(16, 76, 78)     // #104c4e
	negSecondary2 := linRGB(28, 120, 116) // #1c7874
	negPrimary3 := linRGB(16, 34, 72)     // #102248
	negSecondary3 := linRGB(34, 68, 130)  // #224482

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
			for py := startY; py < endY; py++ {
				for px := 0; px < w; px++ {
					var accR, accG, accB float64

					for s := 0; s < 4; s++ {
						sx := float64(px) + subOffsets[s][0]
						sy := float64(py) + subOffsets[s][1]
						dx := sx - cx
						dy := sy - cy
						rScreen := math.Hypot(dx, dy)
						rNorm := rScreen / diskR

						// ----------------------------------------------------------------
						// 1. Exterior Museum Paper & Double-Ringed Engraved Gold Medallion
						// ----------------------------------------------------------------
						if rNorm > 1.0 {
							vig := common.Clamp01((rScreen / (0.72 * minDim)) * (rScreen / (0.72 * minDim)))
							col := lerp3(paperCol, paperShadowCol, 0.38*vig)

							// Double-ringed engraved dark ink & burnished gold frame (1.0 <= rNorm <= 1.058)
							if rNorm <= 1.058 {
								tRim := (rNorm - 1.0) / 0.058 // in [0, 1]
								thetaRim := math.Atan2(dy, dx)

								bevel := 0.5 + 0.5*math.Cos(thetaRim+0.78)
								goldBand := lerp3(bronzeCol, giltCol, 0.35+0.55*bevel*math.Sin(tRim*math.Pi))

								// Fine astrolabe degree graduations inside the gold bezel
								ticks := 0.5 + 0.5*math.Cos(float64(cg.P*24)*thetaRim)
								tickMask := common.Smoothstep(0.75, 0.95, ticks) *
									common.Smoothstep(0.18, 0.26, tRim) *
									common.Smoothstep(0.52, 0.42, tRim)
								goldBand = lerp3(goldBand, inkCol, 0.55*tickMask)

								innerInk := common.Smoothstep(0.18, 0.11, tRim)
								outerInk := common.Smoothstep(0.72, 0.78, tRim) * common.Smoothstep(0.92, 0.86, tRim)
								inkWeight := math.Max(innerInk, outerInk)

								rimCol := lerp3(goldBand, inkCol, inkWeight)
								fadeOut := common.Smoothstep(1.0, 0.88, tRim)
								col = lerp3(col, rimCol, fadeOut)
							} else if rNorm <= 1.088 {
								// Subtle intaglio plate-mark crease in the warm paper outside the medallion
								tCrease := (rNorm - 1.072) / 0.008
								crease := math.Exp(-tCrease * tCrease)
								thetaRim := math.Atan2(dy, dx)
								emboss := math.Cos(thetaRim + 0.78)
								col[0] = common.Clamp01(col[0] + 0.045*crease*emboss)
								col[1] = common.Clamp01(col[1] + 0.042*crease*emboss)
								col[2] = common.Clamp01(col[2] + 0.038*crease*emboss)
							}

							accR += col[0]
							accG += col[1]
							accB += col[2]
							continue
						}

						// ----------------------------------------------------------------
						// 2. Conformal Hyperbolic Radial Mapping & Coxeter Group Fold
						// ----------------------------------------------------------------
						// With the compact (7, 3, 2) Hurwitz fundamental domain (|Vq| = 0.301, |Vr| = 0.264),
						// capping rPoincare at rMax = 0.835 via rPoincare = 0.835 * rNorm^0.82 places
						// 4 well-proportioned concentric rings across [0, 1] (at rNorm ~ 0.29, 0.59, 0.84, 1.00)
						// so the center is compact and the outer edge remains large, crisp, and never crowded!
						rPoincare := 0.835 * math.Pow(rNorm, 0.82)
						zDisk := cmplx.Rect(rPoincare, math.Atan2(dy, dx))

						var z complex128
						if useMobius {
							z = diskMob.Apply(zDisk)
						} else {
							z = zDisk * rotZ
						}

						rDisk := cmplx.Abs(z)
						if rDisk >= 0.9992 {
							z *= complex(0.9992/rDisk, 0)
							rDisk = 0.9992
						}

						zFund, bary, parity, _ := cg.FoldToFundamentalDomain(z, 96)
						bDist := cg.BoundaryDist(zFund)

						// Count exact hyperbolic circle inversions across Side 3 (C3) while folding
						// angle in O(1) via D_p dihedral symmetry so ring generation c3Steps is 100%
						// rotationally symmetric around all 2*p sectors!
						c3Steps := 0
						{
							cur := z
							c3x, c3y := real(cg.Side3.Center), imag(cg.Side3.Center)
							r3Sq := cg.Side3.Radius * cg.Side3.Radius
							twoAlpha := 2.0 * cg.Alpha
							for step := 0; step < 32; step++ {
								// Fold angle into [0, alpha] without changing |cur|
								ang := math.Atan2(imag(cur), real(cur))
								if ang < 0 {
									ang = -ang
								}
								rem := math.Mod(ang, twoAlpha)
								if rem > cg.Alpha {
									rem = twoAlpha - rem
								}
								rad := cmplx.Abs(cur)
								cur = cmplx.Rect(rad, rem)

								dx3 := real(cur) - c3x
								dy3 := imag(cur) - c3y
								dSq := dx3*dx3 + dy3*dy3
								if dSq < r3Sq-1e-11 && dSq > 1e-18 {
									scale := r3Sq / dSq
									cur = complex(c3x+dx3*scale, c3y+dy3*scale)
									c3Steps++
								} else {
									break
								}
							}
						}

						u, v, wBary := bary[0], bary[1], bary[2]

						// Exact normalized distances to each of the 3 geodesic sides of Delta_{pqr}
						zfX, zfY := real(zFund), imag(zFund)
						e1 := math.Abs(zfY) / triScale
						e2 := math.Abs(zfX*cg.Line2NormX+zfY*cg.Line2NormY) / triScale
						e3 := math.Abs(cg.Side3.SignedDist(zFund)) / triScale
						dTri := bDist / triScale

						// ----------------------------------------------------------------
						// 3. Antisymmetric C1 S-Curve Interlock Across Triangle Boundaries
						// ----------------------------------------------------------------
						// Parameter along each edge (invariant under reflection across that edge)
						t1 := common.Clamp01(v / (u + v + 1e-9))
						t2 := common.Clamp01(wBary / (u + wBary + 1e-9))
						t3 := common.Clamp01(wBary / (v + wBary + 1e-9))

						// Smooth C1 Bell envelopes that vanish at all triangle vertices
						env1 := 4.0 * t1 * (1.0 - t1) * math.Exp(-e1/0.155) *
							common.Smoothstep(0.0, 0.042, e2) * common.Smoothstep(0.0, 0.042, e3)
						env2 := 4.0 * t2 * (1.0 - t2) * math.Exp(-e2/0.155) *
							common.Smoothstep(0.0, 0.042, e1) * common.Smoothstep(0.0, 0.042, e3)
						env3 := 4.0 * t3 * (1.0 - t3) * math.Exp(-e3/0.155) *
							common.Smoothstep(0.0, 0.042, e1) * common.Smoothstep(0.0, 0.042, e2)

						// Bold S-wave wing, head, and fin interlock lobes (~35% higher amplitude)
						p1 := env1 * (0.118*math.Sin(2.0*math.Pi*t1) + 0.035*math.Sin(4.0*math.Pi*t1-0.45))
						p2 := env2 * (-0.114*math.Sin(2.0*math.Pi*t2) + 0.032*math.Sin(4.0*math.Pi*t2+0.35))
						p3 := env3 * (0.104*math.Sin(2.0*math.Pi*t3) - 0.030*math.Sin(4.0*math.Pi*t3))
						perturb := p1 + p2 + p3

						// Signed interlocking figure field:
						// sigma flips sign (+1 vs -1) across every geodesic reflection edge while
						// dTri == 0 on the edge and perturb is continuous across the edge, producing
						// a seamless C1 interlocking boundary between positive and negative figures.
						sigma := 1.0 - 2.0*float64(parity)
						phiInterlock := sigma*dTri - perturb

						figParity := 0
						if phiInterlock < 0 {
							figParity = 1
						}

						// Exact D_p-symmetric ring generation index from C3 hyperbolic inversions
						effRing := c3Steps
						if figParity != parity && effRing > 0 && e3 < e1 && e3 < e2 {
							effRing--
						}
						tileVariant := effRing & 1

						// Continuous 100% radially symmetric center-to-rim progression combining
						// radial position rNorm and C3 hyperbolic ring generation effRing
						ringT := common.Clamp01(0.58*rNorm + 0.42*common.Clamp01(float64(effRing)/3.2))

						// ----------------------------------------------------------------
						// 4. Multi-Ring Chromatic Progression & Anatomical Shading
						// ----------------------------------------------------------------
						wingSpan := common.Clamp01(0.58*v + 0.42*wBary)
						headCrown := common.Smoothstep(0.42, 0.86, u)
						wingTip := common.Smoothstep(0.32, 0.78, wBary)
						absPhi := math.Abs(phiInterlock)

						var baseCol [3]float64

						if figParity == 0 {
							// Positive Figures (warm/luminous motifs evolving from Sun-Gold -> Ivory/Coral -> Jade/Mint -> Silver-Ice/Sky)
							pri := sample4Zone(posPrimary0, posPrimary1, posPrimary2, posPrimary3, ringT)
							sec := sample4Zone(posSecondary0, posSecondary1, posSecondary2, posSecondary3, ringT)

							if tileVariant == 1 {
								// Alternate generation tiles swap primary/secondary weight for strong ring-to-ring rhythm
								pri, sec = lerp3(pri, sec, 0.65), lerp3(sec, pri, 0.65)
							}

							// Sculpted wing-span and crown anatomical gradient
							gradMix := common.Clamp01(0.48*wingSpan + 0.34*wingTip - 0.18*headCrown)
							baseCol = lerp3(pri, sec, gradMix)

							// Luminous ivory-gold crown highlight near the apex u -> 1
							crownGlow := headCrown * (0.42 + 0.25*(1.0-ringT))
							baseCol = lerp3(baseCol, ivoryCol, crownGlow)

							// Warm bronze depth in the inner body core
							coreDepth := common.Smoothstep(0.02, 0.22, absPhi) * 0.16
							baseCol = lerp3(baseCol, bronzeCol, coreDepth)
						} else {
							// Negative Figures (deep/rich motifs evolving from Cinnabar Ruby -> Terracotta/Bronze -> Emerald-Pine -> Prussian Lapis)
							pri := sample4Zone(negPrimary0, negPrimary1, negPrimary2, negPrimary3, ringT)
							sec := sample4Zone(negSecondary0, negSecondary1, negSecondary2, negSecondary3, ringT)

							if tileVariant == 1 {
								pri, sec = lerp3(pri, sec, 0.68), lerp3(sec, pri, 0.68)
							}

							// Wing membrane iridescence & anatomical gradient
							gradMix := common.Clamp01(0.52*wingSpan + 0.32*headCrown)
							baseCol = lerp3(pri, sec, gradMix)

							// Iridescent wing-tip highlight contrasting with the deep body
							tipAccent := lerp3(goldCoreCol, posSecondary3, ringT)
							baseCol = lerp3(baseCol, tipAccent, 0.28*wingTip)
						}

						// ----------------------------------------------------------------
						// 5. Sculpted Wing-Feather Chevrons, Scalloped Scales & Ocellus Eye
						// ----------------------------------------------------------------
						// Keep detail razor-sharp all the way to rNorm = 0.98!
						aaFade := common.Smoothstep(0.995, 0.92, rNorm)
						lineScale := 1.0 - 0.25*rNorm

						// Concentric wing-feather chevrons & scalloped body scales
						featherPhase := 30.0*(v-wBary) + 12.0*u + 5.0*math.Sin(8.5*u)
						chevronPhase := 22.0*v + 22.0*wBary - 14.0*u + 3.5*math.Cos(10.0*(v-wBary))
						striation1 := 0.5 + 0.5*math.Sin(featherPhase)
						striation2 := 0.5 + 0.5*math.Sin(chevronPhase)

						// Scalloped body scale cells in the central torso of each motif
						scaleGrid := (0.5 + 0.5*math.Sin(20.0*u)) * (0.5 + 0.5*math.Cos(18.0*(v-wBary)))
						scaleRim := common.Smoothstep(0.58, 0.82, scaleGrid) * common.Smoothstep(0.94, 0.82, scaleGrid) *
							common.Smoothstep(0.18, 0.45, u) * common.Smoothstep(0.82, 0.55, u)

						hatch1 := common.Smoothstep(0.72, 0.93, striation1)
						hatch2 := common.Smoothstep(0.78, 0.95, striation2) * common.Smoothstep(0.15, 0.60, wingSpan)
						hatchMask := aaFade * common.Clamp01(0.62*hatch1+0.42*hatch2+0.48*scaleRim) *
							common.Smoothstep(0.012, 0.048, absPhi)

						// Luminous feather ridge highlight between woodcut grooves
						featherRidge := aaFade * common.Smoothstep(0.82, 0.96, 1.0-striation1) *
							common.Smoothstep(0.018, 0.055, absPhi)

						if figParity == 0 {
							// Dark woodcut engraving grooves & gilt feather sheen on positive figures
							baseCol = lerp3(baseCol, inkCol, 0.34*hatchMask)
							baseCol = lerp3(baseCol, giltCol, 0.24*featherRidge)
						} else {
							// Luminous gold/ivory gouge highlights & deep ink grooves on negative figures
							baseCol = lerp3(baseCol, inkCol, 0.32*hatchMask)
							gougeCol := lerp3(giltCol, ivoryCol, 0.55*ringT)
							baseCol = lerp3(baseCol, gougeCol, 0.34*featherRidge)
						}

						// Luminous golden/ivory ocellus eye + crown crest near the head of each figure
						eyeDist := math.Hypot(u-0.55, (v-wBary)-0.085*sigma)
						eyeOuterGlow := common.Smoothstep(0.082, 0.052, eyeDist) * common.Smoothstep(0.020, 0.040, eyeDist)
						eyeRing := common.Smoothstep(0.062, 0.045, eyeDist) * common.Smoothstep(0.022, 0.035, eyeDist)
						eyePupil := common.Smoothstep(0.024, 0.011, eyeDist)

						// Crown crest arc just ahead of the eye toward vertex Vp (u -> 0.74)
						crestDist := math.Abs(u-0.70) + 0.55*math.Abs((v-wBary)-0.04*sigma)
						crestMask := common.Smoothstep(0.055, 0.022, crestDist) * common.Smoothstep(0.008, 0.020, crestDist)

						baseCol = lerp3(baseCol, goldCoreCol, 0.58*aaFade*math.Max(eyeOuterGlow, crestMask))
						baseCol = lerp3(baseCol, ivoryCol, 0.78*aaFade*eyeRing)
						baseCol = lerp3(baseCol, inkCol, 0.90*aaFade*eyePupil)

						// 3D carved relief bevel & dark woodcut outline along the interlocking boundary phiInterlock = 0
						bevelBand := math.Exp(-absPhi/0.040) * common.Smoothstep(0.003, 0.020, absPhi)
						if phiInterlock >= 0 {
							baseCol = lerp3(baseCol, ivoryCol, 0.36*bevelBand)
						} else {
							baseCol = lerp3(baseCol, inkCol, 0.38*bevelBand)
						}

						// Crisp woodcut contour line along phiInterlock == 0 (slightly narrowed near rNorm -> 1)
						contourWidth := 0.0095 * lineScale
						contourInk := common.Smoothstep(contourWidth, 0.0020*lineScale, absPhi)
						baseCol = lerp3(baseCol, inkCol, 0.88*contourInk)

						// ----------------------------------------------------------------
						// 6. Circle Limit III Double-Ruled 24k Gold & Ivory Geodesic Highways
						// ----------------------------------------------------------------
						// Along the main symmetry axes (e3 and e1), draw crisp double-ruled
						// 24k gold & ivory splines weaving through all rings from center to rim.
						backboneDist := math.Min(e3, 0.85*e1)
						highwayOuter := 0.021 * lineScale
						highwayMid := 0.0125 * lineScale
						highwayInner := 0.0055 * lineScale

						highwayCasing := common.Smoothstep(highwayOuter, highwayMid, backboneDist)
						highwayGoldRule := common.Smoothstep(highwayMid, highwayInner, backboneDist)
						highwayIvoryCore := common.Smoothstep(highwayInner, 0.0012*lineScale, backboneDist)

						// Secondary delicate gold filigree along fundamental geodesics (dTri near 0)
						filigreeMask := common.Smoothstep(0.0075*lineScale, 0.0014*lineScale, dTri) * 0.52

						if highwayCasing > 1e-4 {
							// Outer dark woodcut casing
							baseCol = lerp3(baseCol, inkCol, 0.76*highwayCasing)
							// Double-ruled 24k burnished gold spline shoulders
							baseCol = lerp3(baseCol, giltCol, 0.92*highwayGoldRule)
							// Thin dark pin-line separating gold shoulders from the gleaming ivory core
							pinLine := common.Smoothstep(0.0072*lineScale, 0.0055*lineScale, backboneDist) *
								common.Smoothstep(0.0035*lineScale, 0.0050*lineScale, backboneDist)
							baseCol = lerp3(baseCol, inkCol, 0.55*pinLine)
							// Luminous museum ivory center spline
							baseCol = lerp3(baseCol, ivoryCol, 0.95*highwayIvoryCore)
						}

						if filigreeMask > 1e-4 {
							baseCol = lerp3(baseCol, giltCol, filigreeMask)
						}

						// Subtle radial horizon framing near rNorm -> 1
						rimShadow := common.Smoothstep(0.965, 1.0, rNorm)
						baseCol = lerp3(baseCol, inkCol, 0.38*rimShadow)

						accR += common.Clamp01(baseCol[0])
						accG += common.Clamp01(baseCol[1])
						accB += common.Clamp01(baseCol[2])
					}

					rc.SetPixelLinear(px, py, accR*0.25, accG*0.25, accB*0.25, 1.0)
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	// Apply tactile cotton printmaking paper grain and merge into the target Canva
	seed := int64(ecl.p*1000 + ecl.q*100 + ecl.r*10 + 1959)
	rc.ApplyPaperTexture(0.11, seed)
	c.MergeRenderCanvas(rc)
}
