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

type escherDrosteGallery struct {
	branches      int
	archesPerRing int
	scaleRatio    float64
}

// NewEscherDrosteGallery creates an EscherDrosteGallery generator that renders
// a 100% conformal M.C. Escher / Alhambra logarithmic-spiral ornamental tapestry
// featuring interlocking over-under girih strapwork, alternating 8-fold star-polyhedron
// medallions and 4-fold curvilinear Escher pinwheel rosettes, golden guilloché
// spiral dividers, and a central 16-pointed compass-rose seal.
func NewEscherDrosteGallery(branches, archesPerRing int, scaleRatio float64) *escherDrosteGallery {
	if branches <= 0 {
		branches = 1
	}
	if archesPerRing <= 0 {
		archesPerRing = 8
	}
	if scaleRatio <= 1.1 {
		scaleRatio = 16.0
	}
	return &escherDrosteGallery{
		branches:      branches,
		archesPerRing: archesPerRing,
		scaleRatio:    scaleRatio,
	}
}

// Generative renders the static EscherDrosteGallery composition.
func (edg *escherDrosteGallery) Generative(c *generativeart.Canva) {
	edg.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless infinite Droste zoom loop (phase in [0, 1)).
func (edg *escherDrosteGallery) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	minDim := math.Min(fw, fh)
	invNorm := 1.0 / (0.48 * minDim)

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	// Convert museum-grade woodcut & cloisonné palette colors to linear RGB
	ivoryR, ivoryG, ivoryB, _ := render.ColorToLinearRGB(color.RGBA{R: 0xf8, G: 0xf1, B: 0xe2, A: 255})       // #f8f1e2 Warm Ivory Parchment
	indigoR, indigoG, indigoB, _ := render.ColorToLinearRGB(color.RGBA{R: 0x12, G: 0x24, B: 0x44, A: 255})    // #122444 Deep Prussian Indigo
	celadonR, celadonG, celadonB, _ := render.ColorToLinearRGB(color.RGBA{R: 0x22, G: 0x8e, B: 0x84, A: 255}) // #228e84 Imperial Celadon
	cinnabarR, cinnabarG, cinnabarB, _ := render.ColorToLinearRGB(color.RGBA{R: 0xd8, G: 0x48, B: 0x30, A: 255}) // #d84830 Cinnabar Vermilion
	goldR, goldG, goldB, _ := render.ColorToLinearRGB(color.RGBA{R: 0xec, G: 0xc0, B: 0x52, A: 255})          // #ecc052 Burnished 24k Gold
	sandR, sandG, sandB, _ := render.ColorToLinearRGB(color.RGBA{R: 0xe8, G: 0xb8, B: 0x72, A: 255})          // #e8b872 Golden Sandstone
	inkR, inkG, inkB, _ := render.ColorToLinearRGB(color.RGBA{R: 0x0b, G: 0x10, B: 0x1c, A: 255})             // #0b101c Deep Woodcut Ink

	// Exact holomorphic conformal spiral lattice parameters:
	// Choose even integer angular cell count N and even spiral shift B = 2 * branches so
	// W = (kR + i*kI) * ln(z) is 100% conformal (zero shear, 1:1 unit squares everywhere)
	// and a 2x2 supercell pattern has zero discontinuity across theta = +-pi.
	numCellsAngular := edg.archesPerRing + 4
	if numCellsAngular < 12 {
		numCellsAngular = 12
	}
	if numCellsAngular%2 != 0 {
		numCellsAngular++
	}
	spiralShiftB := 2 * edg.branches

	kR := float64(numCellsAngular) / (2.0 * math.Pi)
	kI := float64(spiralShiftB) / (2.0 * math.Pi)
	phi := 2.0 * math.Pi * phase

	// Sub-pixel 2x2 offsets for supersampled anti-aliasing
	subOffsets := [4][2]float64{
		{0.25, 0.25},
		{0.75, 0.25},
		{0.25, 0.75},
		{0.75, 0.75},
	}

	// Evaluate the conformal Escher ornamental tapestry at continuous conformal coordinates (uCont, vCont)
	evalOrnament := func(uCont, vCont, confScale float64) (float64, float64, float64) {
		flU := math.Floor(uCont)
		flV := math.Floor(vCont)
		cellU := int(flU)
		cellV := int(flV)

		fu := uCont - flU // in [0, 1)
		fv := vCont - flV // in [0, 1)

		// Symmetric conformal unit-cell coordinates in [-1, 1]^2
		fx := 2.0*fu - 1.0
		fy := 2.0*fv - 1.0
		absX := math.Abs(fx)
		absY := math.Abs(fy)
		rCell := math.Hypot(fx, fy)
		angCell := math.Atan2(fy, fx)

		// Even-period 2x2 checkerboard and ring parities (100% seamless across theta = +-pi)
		parity := (cellU + cellV) & 1
		ringParity := cellU & 1

		// Sub-pixel detail attenuation near the central singularity
		detailVis := common.Smoothstep(0.008, 0.045, confScale)

		// Helper for fine woodcut engraving striations
		woodcutEngrave := func(coord1, coord2, weight float64) float64 {
			if detailVis <= 1e-3 || weight <= 1e-3 {
				return 1.0
			}
			w1 := 0.5 + 0.5*math.Sin(coord1)
			w2 := 0.5 + 0.5*math.Sin(coord2)
			l1 := 1.0 - common.Smoothstep(0.20, 0.46, w1)
			l2 := (1.0 - common.Smoothstep(0.16, 0.42, w2)) * 0.60
			return 1.0 - 0.46*math.Max(l1, l2)*weight*detailVis
		}

		var rOut, gOut, bOut float64

		if parity == 0 {
			// =========================================================================
			// MOTIF A (parity == 0): 8-Fold Alhambra Star-Polyhedron & Scalloped Lotus Compass
			// =========================================================================
			// 1. Cloisonné enamel background with radial woodcut guilloché waves
			radialWave := 0.5 + 0.5*math.Sin(rCell*18.0-4.0*math.Cos(8.0*angCell)-phi)
			bgMix := 0.28 + 0.42*common.Smoothstep(0.25, 0.85, rCell)*radialWave
			if ringParity != 0 {
				bgMix = 0.35 + 0.40*(1.0-radialWave*0.5)
			}
			engraveBg := woodcutEngrave(rCell*42.0, 16.0*angCell, 0.55)
			rOut = (indigoR*(1.0-bgMix) + celadonR*bgMix) * engraveBg
			gOut = (indigoG*(1.0-bgMix) + celadonG*bgMix) * engraveBg
			bOut = (indigoB*(1.0-bgMix) + celadonB*bgMix) * engraveBg

			// 2. Corner quarter-rosettes at (|fx|, |fy|) near (1, 1) that form complete 8-petal
			// floral medallions where 4 cells meet at every lattice vertex!
			cornerDist := math.Hypot(1.0-absX, 1.0-absY)
			if cornerDist < 0.44 {
				cAng := math.Atan2(1.0-absY, 1.0-absX)
				cPetal := 0.34 + 0.05*math.Cos(8.0*cAng)
				if cornerDist < cPetal+0.02 {
					cCov := 1.0 - common.Smoothstep(cPetal-0.015, cPetal+0.015, cornerDist)
					cNorm := cornerDist / cPetal
					cBevel := 0.65 + 0.35*math.Sin(cNorm*math.Pi)
					isRim := cNorm > 0.76
					var cR, cG, cB float64
					if isRim {
						cR, cG, cB = goldR*cBevel, goldG*cBevel, goldB*cBevel
					} else if cNorm > 0.42 {
						cR = (ivoryR*0.80 + sandR*0.20) * cBevel
						cG = (ivoryG*0.80 + sandG*0.20) * cBevel
						cB = (ivoryB*0.80 + sandB*0.20) * cBevel
					} else {
						cR, cG, cB = cinnabarR*cBevel, cinnabarG*cBevel, cinnabarB*cBevel
					}
					rOut = rOut*(1.0-cCov) + cR*cCov
					gOut = gOut*(1.0-cCov) + cG*cCov
					bOut = bOut*(1.0-cCov) + cB*cCov
				}
			}

			// 3. Outer 16-scallop lotus aureole ring (rCell in [0.50, 0.78])
			scallopR := 0.72 + 0.045*math.Cos(16.0*angCell)
			if rCell < scallopR+0.04 {
				// Soft drop shadow cast by the scalloped aureole onto the indigo background
				if rCell >= scallopR-0.01 {
					sh := 1.0 - common.Smoothstep(scallopR-0.01, scallopR+0.04, rCell)
					rOut *= 1.0 - 0.48*sh
					gOut *= 1.0 - 0.48*sh
					bOut *= 1.0 - 0.48*sh
				}
				if rCell < scallopR+0.01 {
					scCov := 1.0 - common.Smoothstep(scallopR-0.015, scallopR+0.01, rCell)
					scT := common.Clamp01((rCell - 0.46) / (scallopR - 0.46))
					// Radial ribbed flute shading across the 16 lotus petals
					flute := 0.5 + 0.5*math.Cos(16.0*angCell)
					petalLight := (0.62 + 0.38*math.Sin(scT*math.Pi)*(0.6+0.4*flute)) *
						woodcutEngrave(32.0*angCell, rCell*36.0, 0.35)

					var scR, scG, scB float64
					if scT > 0.84 {
						// Burnished 24k gold scalloped rim with dark ink engraving line
						scR, scG, scB = goldR*petalLight*1.08, goldG*petalLight*1.08, goldB*petalLight*1.08
					} else {
						// Alternating warm ivory and celadon-inlaid lotus petals
						petalIdx := int(math.Floor((angCell + math.Pi) / (2.0 * math.Pi) * 16.0))
						if petalIdx%2 == 0 {
							scR = (ivoryR*0.86 + sandR*0.14) * petalLight
							scG = (ivoryG*0.86 + sandG*0.14) * petalLight
							scB = (ivoryB*0.86 + sandB*0.14) * petalLight
						} else {
							scR = (celadonR*0.72 + ivoryR*0.28) * petalLight
							scG = (celadonG*0.72 + ivoryG*0.28) * petalLight
							scB = (celadonB*0.72 + ivoryB*0.28) * petalLight
						}
					}
					rOut = rOut*(1.0-scCov) + scR*scCov
					gOut = gOut*(1.0-scCov) + scG*scCov
					bOut = bOut*(1.0-scCov) + scB*scCov
				}
			}

			// 4. Interlocking 8-Pointed Alhambra Star Polyhedron ({8/2} Rub el Hizb)
			// Constructed from two orthogonal squares rotated by 0 and 45 degrees:
			s1 := math.Max(absX, absY)
			s2 := math.Max(math.Abs(fx+fy), math.Abs(fx-fy)) * 0.70710678
			star8 := math.Min(s1, s2)
			starRadius := 0.52

			if star8 < starRadius+0.035 {
				if star8 >= starRadius-0.01 {
					sh := 1.0 - common.Smoothstep(starRadius-0.01, starRadius+0.035, star8)
					rOut *= 1.0 - 0.52*sh
					gOut *= 1.0 - 0.52*sh
					bOut *= 1.0 - 0.52*sh
				}
				if star8 < starRadius+0.01 {
					stCov := 1.0 - common.Smoothstep(starRadius-0.015, starRadius+0.01, star8)
					// Over-under weave between Square 1 (axis-aligned) and Square 2 (45-deg rotated)
					border1 := math.Abs(s1 - (starRadius - 0.055))
					border2 := math.Abs(s2 - (starRadius - 0.055))
					inRibbon1 := s1 <= starRadius && border1 < 0.055
					inRibbon2 := s2 <= starRadius && border2 < 0.055

					// 3D faceted polyhedral ridge lighting inside the 8-pointed star
					ridge := math.Sin(8.0 * angCell)
					facetLight := 0.68 + 0.28*ridge*(1.0-0.5*star8/starRadius) + 0.15*math.Cos(star8/starRadius*math.Pi)
					facetLight *= woodcutEngrave(fx*36.0, fy*36.0, 0.30)

					var stR, stG, stB float64
					if inRibbon1 || inRibbon2 {
						// Interlocking 24k gold & ivory border straps of the two squares
						ribT := border1 / 0.055
						if inRibbon2 && (!inRibbon1 || (fx*fy > 0 && border2 < border1)) {
							ribT = border2 / 0.055
						}
						ribBevel := 0.65 + 0.42*math.Cos(ribT*math.Pi*0.5)
						edgeInk := common.Smoothstep(0.75, 0.96, ribT)
						stR = (goldR*0.82+ivoryR*0.18)*ribBevel*(1.0-0.65*edgeInk) + inkR*0.65*edgeInk
						stG = (goldG*0.82+ivoryG*0.18)*ribBevel*(1.0-0.65*edgeInk) + inkG*0.65*edgeInk
						stB = (goldB*0.82+ivoryB*0.18)*ribBevel*(1.0-0.65*edgeInk) + inkB*0.65*edgeInk
					} else {
						// Inner faceted star field alternating Cinnabar Vermilion and Warm Ivory
						if ridge > 0 {
							stR = (cinnabarR*0.85 + goldR*0.15) * facetLight
							stG = (cinnabarG*0.85 + goldG*0.15) * facetLight
							stB = (cinnabarB*0.85 + goldB*0.15) * facetLight
						} else {
							stR = (ivoryR*0.88 + sandR*0.12) * facetLight
							stG = (ivoryG*0.88 + sandG*0.12) * facetLight
							stB = (ivoryB*0.88 + sandB*0.12) * facetLight
						}
					}

					rOut = rOut*(1.0-stCov) + stR*stCov
					gOut = gOut*(1.0-stCov) + stG*stCov
					bOut = bOut*(1.0-stCov) + stB*stCov
				}
			}

			// 5. Central Cabochon Jewel & Gold Filigree Rosette (rCell < 0.22)
			if rCell < 0.23 {
				cabCov := 1.0 - common.Smoothstep(0.20, 0.23, rCell)
				nr := rCell / 0.21
				var cabR, cabG, cabB float64
				if nr > 0.72 {
					// Beaded 24k gold bezel setting
					beads := 0.5 + 0.5*math.Cos(16.0*angCell)
					bLight := 0.68 + 0.38*beads
					cabR, cabG, cabB = goldR*bLight, goldG*bLight, goldB*bLight
				} else {
					// Domed specular cabochon gem (alternating Celadon and Cinnabar by ringParity)
					domeZ := math.Sqrt(math.Max(0.0, 1.0-nr*nr))
					spec := math.Pow(math.Max(0.0, domeZ+0.35*(-fx-fy)/0.21), 5.0) * 0.55
					if ringParity == 0 {
						cabR = celadonR*(0.45+0.65*domeZ) + ivoryR*spec
						cabG = celadonG*(0.45+0.65*domeZ) + ivoryG*spec
						cabB = celadonB*(0.45+0.65*domeZ) + ivoryB*spec
					} else {
						cabR = cinnabarR*(0.45+0.65*domeZ) + ivoryR*spec
						cabG = cinnabarG*(0.45+0.65*domeZ) + ivoryG*spec
						cabB = cinnabarB*(0.45+0.65*domeZ) + ivoryB*spec
					}
				}
				rOut = rOut*(1.0-cabCov) + cabR*cabCov
				gOut = gOut*(1.0-cabCov) + cabG*cabCov
				bOut = bOut*(1.0-cabCov) + cabB*cabCov
			}
		} else {
			// =========================================================================
			// MOTIF B (parity == 1): Escher Interlocking 4-Fold Pinwheel & Quatrefoil Knot
			// =========================================================================
			// 1. Warm Ivory Parchment & Golden Sandstone ground with concentric woodcut hatching
			concentric := 0.5 + 0.5*math.Cos(rCell*22.0+4.0*angCell)
			sandMix := 0.25 + 0.42*common.Smoothstep(0.15, 0.90, rCell)*concentric
			engraveGround := woodcutEngrave((fx+fy)*28.0, (fx-fy)*28.0, 0.38)
			rOut = (ivoryR*(1.0-sandMix) + sandR*sandMix) * engraveGround
			gOut = (ivoryG*(1.0-sandMix) + sandG*sandMix) * engraveGround
			bOut = (ivoryB*(1.0-sandMix) + sandB*sandMix) * engraveGround

			// 2. Corner quarter-rosettes at (|fx|, |fy|) near (1, 1) matching Motif A's corners
			// so every 4-cell vertex forms a complete unbroken circular floral boss!
			cornerDist := math.Hypot(1.0-absX, 1.0-absY)
			if cornerDist < 0.44 {
				cAng := math.Atan2(1.0-absY, 1.0-absX)
				cPetal := 0.34 + 0.05*math.Cos(8.0*cAng)
				if cornerDist < cPetal+0.02 {
					cCov := 1.0 - common.Smoothstep(cPetal-0.015, cPetal+0.015, cornerDist)
					cNorm := cornerDist / cPetal
					cBevel := 0.65 + 0.35*math.Sin(cNorm*math.Pi)
					var cR, cG, cB float64
					if cNorm > 0.76 {
						cR, cG, cB = goldR*cBevel, goldG*cBevel, goldB*cBevel
					} else if cNorm > 0.42 {
						cR = (ivoryR*0.80 + sandR*0.20) * cBevel
						cG = (ivoryG*0.80 + sandG*0.20) * cBevel
						cB = (ivoryB*0.80 + sandB*0.20) * cBevel
					} else {
						cR, cG, cB = cinnabarR*cBevel, cinnabarG*cBevel, cinnabarB*cBevel
					}
					rOut = rOut*(1.0-cCov) + cR*cCov
					gOut = gOut*(1.0-cCov) + cG*cCov
					bOut = bOut*(1.0-cCov) + cB*cCov
				}
			}

			// 3. Recessed Deep Prussian Indigo & Celadon Quatrefoil Cartouche Frame
			quatR := 0.68 + 0.12*math.Cos(4.0*angCell) - 0.04*math.Cos(8.0*angCell)
			if rCell < quatR+0.03 {
				if rCell < quatR+0.01 {
					qCov := 1.0 - common.Smoothstep(quatR-0.015, quatR+0.01, rCell)
					qNorm := rCell / quatR
					var qR, qG, qB float64
					if qNorm > 0.88 {
						// Outer 24k gold quatrefoil rim
						rimLight := 0.72 + 0.32*math.Sin((qNorm-0.88)/0.12*math.Pi)
						qR, qG, qB = goldR*rimLight, goldG*rimLight, goldB*rimLight
					} else {
						// Deep indigo recessed cartouche bed with fine radial woodcut lines
						bedLight := (0.65 + 0.35*(1.0-qNorm)) * woodcutEngrave(rCell*38.0, 12.0*angCell, 0.50)
						qR = (indigoR*0.82 + celadonR*0.18) * bedLight
						qG = (indigoG*0.82 + celadonG*0.18) * bedLight
						qB = (indigoB*0.82 + celadonB*0.18) * bedLight
					}
					rOut = rOut*(1.0-qCov) + qR*qCov
					gOut = gOut*(1.0-qCov) + qG*qCov
					bOut = bOut*(1.0-qCov) + qB*qCov
				}
			}

			// 4. Escher 4-Fold Rotational Pinwheel & Interlocking Curvilinear Trefoil/Quatrefoil Knot
			// Swirl angle twists smoothly with radius to form Escher's signature interlocking blades
			twistDir := 1.0
			if ringParity != 0 {
				twistDir = -1.0
			}
			swirlAng := angCell + twistDir*(1.35*rCell*math.Sin(rCell*math.Pi)+0.5*phi)
			bladeWave := math.Cos(4.0 * swirlAng)
			bladeEnv := 0.62 * (0.68 + 0.32*math.Cos(4.0*angCell))

			if rCell < bladeEnv && bladeWave > -0.25 {
				bladeCov := common.Smoothstep(-0.22, -0.02, bladeWave) *
					(1.0 - common.Smoothstep(bladeEnv-0.03, bladeEnv, rCell))

				// Determine which of the 4 rotational blades we are on for alternating palette & over/under shading
				sectorIdx := int(math.Floor((swirlAng + 2.0*math.Pi) / (0.5 * math.Pi)))
				bladeProfile := common.Clamp01((bladeWave + 0.15) / 1.15)
				bladeBevel := (0.58 + 0.45*math.Sin(bladeProfile*math.Pi)) *
					woodcutEngrave(swirlAng*20.0, rCell*40.0, 0.32)

				// Gold filigree outline along the edge of each curvilinear Escher blade
				goldEdge := 1.0 - common.Smoothstep(0.0, 0.22, bladeProfile)

				var bR, bG, bB float64
				if sectorIdx%2 == 0 {
					// Cinnabar Vermilion & Burnished Gold blade
					bR = cinnabarR*bladeBevel*(1.0-goldEdge) + goldR*goldEdge*1.05
					bG = cinnabarG*bladeBevel*(1.0-goldEdge) + goldG*goldEdge*1.05
					bB = cinnabarB*bladeBevel*(1.0-goldEdge) + goldB*goldEdge*1.05
				} else {
					// Imperial Celadon & Warm Ivory blade
					bR = (celadonR*0.78+ivoryR*0.22)*bladeBevel*(1.0-goldEdge) + goldR*goldEdge*1.05
					bG = (celadonG*0.78+ivoryG*0.22)*bladeBevel*(1.0-goldEdge) + goldG*goldEdge*1.05
					bB = (celadonB*0.78+ivoryB*0.22)*bladeBevel*(1.0-goldEdge) + goldB*goldEdge*1.05
				}

				rOut = rOut*(1.0-bladeCov) + bR*bladeCov
				gOut = gOut*(1.0-bladeCov) + bG*bladeCov
				bOut = bOut*(1.0-bladeCov) + bB*bladeCov
			}

			// 5. Central 8-Petal Ivory & Gold Pinwheel Hub (rCell < 0.19)
			hubR := 0.16 + 0.025*math.Cos(8.0*angCell)
			if rCell < hubR+0.015 {
				hCov := 1.0 - common.Smoothstep(hubR-0.015, hubR+0.015, rCell)
				hNorm := rCell / hubR
				hBevel := 0.65 + 0.40*math.Cos(hNorm*math.Pi*0.5)
				var hR, hG, hB float64
				if hNorm > 0.68 {
					hR, hG, hB = goldR*hBevel, goldG*hBevel, goldB*hBevel
				} else if hNorm > 0.34 {
					hR, hG, hB = ivoryR*hBevel, ivoryG*hBevel, ivoryB*hBevel
				} else {
					hR, hG, hB = indigoR*hBevel, indigoG*hBevel, indigoB*hBevel
				}
				rOut = rOut*(1.0-hCov) + hR*hCov
				gOut = gOut*(1.0-hCov) + hG*hCov
				bOut = bOut*(1.0-hCov) + hB*hCov
			}
		}

		// =========================================================================
		// CONTINUOUS INTERLOCKING OVER-UNDER GIRIH STRAPWORK & SPIRAL GUILLOCHÉ BORDERS
		// =========================================================================
		// 1. Diagonal Girih Straps: |fx| + |fy| = 1.0 connects midpoints of all 4 cell edges
		// continuously across the entire conformal lattice!
		dGirih1 := math.Abs((fx + fy) - 1.0) * 0.70710678
		dGirih2 := math.Abs((fx + fy) + 1.0) * 0.70710678
		dGirih3 := math.Abs((fx - fy) - 1.0) * 0.70710678
		dGirih4 := math.Abs((fx - fy) + 1.0) * 0.70710678
		dDiagA := math.Min(dGirih1, dGirih2) // +45 deg family
		dDiagB := math.Min(dGirih3, dGirih4) // -45 deg family
		dGirih := math.Min(dDiagA, dDiagB)

		strapHalfW := 0.052
		if dGirih < strapHalfW+0.025 {
			// Cast subtle 3D drop shadow from the girih strap onto the underlying ornament
			if dGirih >= strapHalfW-0.005 {
				sh := 1.0 - common.Smoothstep(strapHalfW-0.005, strapHalfW+0.025, dGirih)
				rOut *= 1.0 - 0.42*sh
				gOut *= 1.0 - 0.42*sh
				bOut *= 1.0 - 0.42*sh
			}
			if dGirih < strapHalfW+0.006 {
				strapCov := 1.0 - common.Smoothstep(strapHalfW-0.010, strapHalfW+0.006, dGirih)
				// Determine over/under crossing shadow where +45 and -45 straps intersect near edge midpoints
				activeD := dDiagA
				otherD := dDiagB
				// Alternate which diagonal family passes on top at (fx=+-1, fy=0) vs (fx=0, fy=+-1)
				aOnTop := (absX > absY) == (parity == 0)
				if dDiagB < dDiagA && (!aOnTop || dDiagA > strapHalfW) {
					activeD = dDiagB
					otherD = dDiagA
				}
				underShadow := 1.0
				if otherD < strapHalfW+0.03 && activeD > otherD {
					underShadow = 0.55 + 0.45*common.Smoothstep(strapHalfW-0.01, strapHalfW+0.03, otherD)
				}

				sNorm := common.Clamp01(activeD / strapHalfW)
				sProfile := math.Sqrt(math.Max(0.0, 1.0-sNorm*sNorm))
				// Double gold fillet with warm ivory/celadon core and dark engraved edges
				coreMask := 1.0 - common.Smoothstep(0.32, 0.52, sNorm)
				inkEdge := common.Smoothstep(0.80, 0.96, sNorm)

				sLight := (0.55 + 0.52*sProfile) * underShadow
				var sR, sG, sB float64
				if parity == 0 {
					sR = (goldR*(1.0-coreMask) + ivoryR*coreMask) * sLight
					sG = (goldG*(1.0-coreMask) + ivoryG*coreMask) * sLight
					sB = (goldB*(1.0-coreMask) + ivoryB*coreMask) * sLight
				} else {
					sR = (goldR*(1.0-coreMask) + (celadonR*0.65+ivoryR*0.35)*coreMask) * sLight
					sG = (goldG*(1.0-coreMask) + (celadonG*0.65+ivoryG*0.35)*coreMask) * sLight
					sB = (goldB*(1.0-coreMask) + (celadonB*0.65+ivoryB*0.35)*coreMask) * sLight
				}
				sR = sR*(1.0-0.70*inkEdge) + inkR*0.70*inkEdge
				sG = sG*(1.0-0.70*inkEdge) + inkG*0.70*inkEdge
				sB = sB*(1.0-0.70*inkEdge) + inkB*0.70*inkEdge

				rOut = rOut*(1.0-strapCov) + sR*strapCov
				gOut = gOut*(1.0-strapCov) + sG*strapCov
				bOut = bOut*(1.0-strapCov) + sB*strapCov
			}
		}

		// 2. Spiral Golden Guilloché Divider Ribbons along cell boundaries (|fx| -> 1 and |fy| -> 1)
		edgeDistU := 1.0 - absX
		edgeDistV := 1.0 - absY
		minBorder := math.Min(edgeDistU, edgeDistV)
		borderHalfW := 0.044

		if minBorder < borderHalfW {
			bCov := 1.0 - common.Smoothstep(borderHalfW-0.010, borderHalfW, minBorder)
			bNorm := minBorder / borderHalfW
			// Guilloché cable wave along the spiral boundary
			cableWave := 0.5 + 0.5*math.Sin(math.Pi*12.0*(fu+fv)-phi)
			bLight := 0.62 + 0.38*math.Cos(bNorm*math.Pi*0.5)*(0.75+0.25*cableWave)
			bInk := common.Smoothstep(0.76, 0.95, bNorm)

			bR := (goldR*0.85+sandR*0.15)*bLight*(1.0-0.65*bInk) + inkR*0.65*bInk
			bG := (goldG*0.85+sandG*0.15)*bLight*(1.0-0.65*bInk) + inkG*0.65*bInk
			bB := (goldB*0.85+sandB*0.15)*bLight*(1.0-0.65*bInk) + inkB*0.65*bInk

			rOut = rOut*(1.0-bCov) + bR*bCov
			gOut = gOut*(1.0-bCov) + bG*bCov
			bOut = bOut*(1.0-bCov) + bB*bCov
		}

		// 3. Pearl Studs at Edge Midpoints ((|fx|=1, fy=0) and (fx=0, |fy|=1)) where Girih straps cross the border
		midDist := math.Min(math.Hypot(1.0-absX, fy), math.Hypot(fx, 1.0-absY))
		if midDist < 0.095 {
			pCov := 1.0 - common.Smoothstep(0.075, 0.095, midDist)
			pNorm := midDist / 0.085
			pZ := math.Sqrt(math.Max(0.0, 1.0-pNorm*pNorm))
			pR := ivoryR*(0.55+0.55*pZ) + goldR*0.25*(1.0-pZ)
			pG := ivoryG*(0.55+0.55*pZ) + goldG*0.25*(1.0-pZ)
			pB := ivoryB*(0.55+0.55*pZ) + goldB*0.25*(1.0-pZ)
			rOut = rOut*(1.0-pCov) + pR*pCov
			gOut = gOut*(1.0-pCov) + pG*pCov
			bOut = bOut*(1.0-pCov) + pB*pCov
		}

		return rOut, gOut, bOut
	}

	// Parallel row rendering with 2x2 SSAA
	numWorkers := runtime.NumCPU()
	if numWorkers < 1 {
		numWorkers = 1
	}
	var wg sync.WaitGroup
	rowsPerWorker := (h + numWorkers - 1) / numWorkers

	for worker := 0; worker < numWorkers; worker++ {
		y0 := worker * rowsPerWorker
		y1 := y0 + rowsPerWorker
		if y1 > h {
			y1 = h
		}
		if y0 >= y1 {
			continue
		}

		wg.Add(1)
		go func(startY, endY int) {
			defer wg.Done()
			for y := startY; y < endY; y++ {
				for x := 0; x < w; x++ {
					var accR, accG, accB float64

					for s := 0; s < 4; s++ {
						px := float64(x) + subOffsets[s][0]
						py := float64(y) + subOffsets[s][1]

						// Map screen coordinate (x, y) to normalized complex coordinate zScreen
						zScreen := complex((px-0.5*fw)*invNorm, (py-0.5*fh)*invNorm)
						rScreen := cmplx.Abs(zScreen)
						theta := cmplx.Phase(zScreen)

						// Obtain conformal scale factor for anti-aliased woodcut detail attenuation
						_, _, _, confScale := common.DrosteSpiralPullback(zScreen, edg.scaleRatio, edg.branches, 1.0)

						// 100% Holomorphic Conformal Multiplication:
						// W = (kR + i * kI) * (ln(r) + i * theta) + 2.0 * phase
						lnR := math.Log(math.Max(rScreen, 1e-9))
						uCont := kR*lnR - kI*theta + 2.0*phase
						vCont := kI*lnR + kR*theta

						sR, sG, sB := evalOrnament(uCont, vCont, confScale)

						// Central 16-Pointed Compass-Rose Medallion at rScreen < 0.032
						// Resolves the logarithmic spiral origin with jewel-like precision
						if rScreen < 0.032 {
							medCov := 1.0 - common.Smoothstep(0.026, 0.032, rScreen)
							nr := rScreen / 0.029 // normalized radius in [0, 1.1]
							thetaCompass := theta - phi/8.0

							var medR, medG, medB float64
							if nr > 0.84 {
								// Outer 24k gold bezel ring with 32 engraved compass ticks
								ticks := 0.5 + 0.5*math.Cos(32.0*thetaCompass)
								bezelShade := 0.72 + 0.32*ticks
								medR = goldR * bezelShade
								medG = goldG * bezelShade
								medB = goldB * bezelShade
							} else if nr > 0.74 {
								// Deep Prussian Indigo inner fillet ring
								medR, medG, medB = indigoR, indigoG, indigoB
							} else {
								// 16-pointed gold & ivory compass star over celadon/indigo enamel
								star16 := 0.48 + 0.22*math.Cos(8.0*thetaCompass) + 0.10*math.Cos(16.0*thetaCompass)
								if nr < 0.22 {
									// Central cabochon ruby/cinnabar jewel
									dome := math.Sqrt(math.Max(0.0, 1.0-(nr/0.22)*(nr/0.22)))
									medR = cinnabarR*(0.55+0.55*dome) + goldR*0.20
									medG = cinnabarG*(0.55+0.55*dome) + goldG*0.20
									medB = cinnabarB*(0.55+0.55*dome) + goldB*0.20
								} else if nr < star16 {
									sector := int(math.Floor((thetaCompass + 2.0*math.Pi) / (2.0 * math.Pi) * 16.0))
									if sector%2 == 0 {
										medR, medG, medB = goldR*1.05, goldG*1.05, goldB*1.05
									} else {
										medR, medG, medB = ivoryR*0.95, ivoryG*0.95, ivoryB*0.95
									}
								} else {
									medR = indigoR*0.70 + celadonR*0.30
									medG = indigoG*0.70 + celadonG*0.30
									medB = indigoB*0.70 + celadonB*0.30
								}
							}

							sR = sR*(1.0-medCov) + medR*medCov
							sG = sG*(1.0-medCov) + medG*medCov
							sB = sB*(1.0-medCov) + medB*medCov
						}

						accR += sR
						accG += sG
						accB += sB
					}

					rc.SetPixelLinear(x, y, accR*0.25, accG*0.25, accB*0.25, 1.0)
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	// Post-processing pass: 2.5D luminance relief, museum cotton paper texture, and radial vignette
	const seed int64 = 20261101
	rc.ApplyLuminanceRelief(0.26, -135.0, 48.0, 0.16)
	rc.ApplyPaperTexture(0.09, seed)
	rc.ApplyVignette(0.16)
	c.MergeRenderCanvas(rc)
}
