package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type moebiusIncal struct {
	facets       int
	ringLayers   int
	crystalScale float64
}

// NewMoebiusIncal returns a MoebiusIncal generator inspired by Jean "Mœbius"
// Giraud and Alejandro Jodorowsky's The Incal (L'Incal): a deep 3D perspective
// Techno-City megastructure chasm with weathered stone/metal relief, directional
// sunlit-ochre vs Prussian-indigo chiaroscuro, 3D beveled gyroscopic astrolabe
// rings, and the luminous Light & Dark dual-octahedron Incal crystal at the core.
func NewMoebiusIncal(facets, ringLayers int, crystalScale float64) *moebiusIncal {
	if facets <= 0 {
		facets = 8
	}
	if ringLayers <= 0 {
		ringLayers = 3
	}
	if crystalScale <= 0 {
		crystalScale = 1.0
	}
	return &moebiusIncal{
		facets:       facets,
		ringLayers:   ringLayers,
		crystalScale: crystalScale,
	}
}

// Generative renders the static MoebiusIncal Ligne Claire composition.
func (mi *moebiusIncal) Generative(c *generativeart.Canva) {
	mi.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless looping MoebiusIncal animation.
func (mi *moebiusIncal) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	seed := c.Seed()
	if seed == 0 {
		seed = 20262102
	}
	ne := common.NewNoiseEngine(seed)
	phi := 2.0 * math.Pi * phase
	scale := common.Constrain(mi.crystalScale, 0.75, 1.30)

	// Cohesive Mœbius L'Incal Chiaroscuro & Gouache Palette (Zero Patchwork Quilt!)
	incalGold := moebiusLinRGB(0xff, 0xdc, 0x58)    // #ffdc58 Luminous White-Gold Incal Core
	incalCyan := moebiusLinRGB(0x48, 0xf0, 0xe2)    // #48f0e2 Refractive Turquoise-Cyan Facet
	incalCoral := moebiusLinRGB(0xf2, 0x4c, 0x62)   // #f24c62 Electric Magenta-Coral Lower Incal
	incalViolet := moebiusLinRGB(0x7c, 0x5c, 0xb8)  // #7c5cb8 Amethyst Prism Midtone
	incalShade := moebiusLinRGB(0x22, 0x34, 0x7c)   // #22347c Royal Ultramarine Crystal Shadow
	incalDeep := moebiusLinRGB(0x10, 0x16, 0x38)    // #101638 Deep Abyssal Shadow Core

	ringIvory := moebiusLinRGB(0xfc, 0xf4, 0xde)    // #fcf4de Sunlit Bone-Ivory Gyro-Ring
	ringBrass := moebiusLinRGB(0xf6, 0xbc, 0x56)    // #f6bc56 Warm Saffron-Brass Ring
	ringShade := moebiusLinRGB(0x2c, 0x52, 0x86)    // #2c5286 Cerulean-Slate Ring Shadow

	shaftSunGold := moebiusLinRGB(0xf9, 0xd8, 0x88) // #f9d888 Sunlit Upper-Left Shaft Limestone
	shaftSunCoral := moebiusLinRGB(0xec, 0x78, 0x52)// #ec7852 Warm Terracotta Citadel Wall
	shaftMid := moebiusLinRGB(0x98, 0x68, 0x8e)     // #98688e Rose-Lavender Terminator
	shaftShade := moebiusLinRGB(0x28, 0x3c, 0x78)   // #283c78 Cool Prussian-Cobalt Chasm Shadow
	abyssCore := moebiusLinRGB(0xff, 0xf8, 0xe8)    // #fff8e8 Central Star-Core Glow
	atmosGlow := moebiusLinRGB(0x64, 0xe4, 0xd8)    // #64e4d8 Bioluminescent Cyan Chasm Mist

	lightDir := moebiusNorm3(-0.62, -0.50, 0.60)

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)
	cfg := render.DefaultMoebiusConfig(seed)
	cfg.WarmBounceColor = color.RGBA{R: 0xec, G: 0x78, B: 0x52, A: 255}

	render.RenderMoebiusScene(rc, cfg, func(ux, uy float64) render.MoebiusPixel {
		x := ux*2.0 - 1.0
		y := uy*2.0 - 1.0
		r := math.Hypot(x, y)
		theta := math.Atan2(y, x)

		// =====================================================================
		// 1. CENTRAL LUMINOUS DUAL-PYRAMID CRYSTAL ("THE INCAL": Light & Dark)
		// =====================================================================
		gapHalf := 0.020 + 0.004*math.Sin(phi)
		apexH := 0.35 * scale
		baseW := 0.28 * scale

		absY := math.Abs(y)
		if absY >= gapHalf && absY <= apexH {
			tPyr := (absY - gapHalf) / (apexH - gapHalf)
			halfSpan := baseW * (1.0 - 0.88*tPyr)
			if math.Abs(x) <= halfSpan {
				uPyr := x / math.Max(halfSpan, 1e-4)
				isUpper := y < 0
				rotDir := 1.0
				if !isUpper {
					rotDir = -1.0
				}
				ridgeShift := 0.24 * math.Sin(rotDir*phi)

				var facetIdx int
				var fnx float64
				switch {
				case uPyr < -0.46+ridgeShift:
					facetIdx = 0
					fnx = -0.82
				case uPyr < -0.10+ridgeShift:
					facetIdx = 1
					fnx = -0.36
				case uPyr < 0.26+ridgeShift:
					facetIdx = 2
					fnx = 0.08
				case uPyr < 0.62+ridgeShift:
					facetIdx = 3
					fnx = 0.48
				default:
					facetIdx = 4
					fnx = 0.84
				}
				fny := -0.46
				if !isUpper {
					fny = 0.48
				}
				fnz := math.Sqrt(math.Max(0.14, 1.0-fnx*fnx-fny*fny))
				n := moebiusNorm3(fnx, fny, fnz)
				lit := common.Clamp01(0.14 + 0.88*math.Max(0.0, moebiusDot3(n, lightDir)))

				edgeDist := math.Min(1.0-math.Abs(uPyr), math.Min(tPyr, 1.0-tPyr)*2.4)
				inInnerGem := edgeDist > 0.24
				tierStep := int(math.Floor(tPyr * 5.0))
				stone := render.EvaluateMoebiusStoneRelief(ne, uPyr, tPyr, 5.0, 5.0, 0.2)

				var sunC, midC [3]float64
				if isUpper {
					sunC = moebiusLerp3(incalCyan, incalGold, float64(4-facetIdx)/4.0*0.72)
					if inInnerGem || tierStep == 0 {
						sunC = incalGold
					}
					midC = incalCyan
				} else {
					sunC = moebiusLerp3(incalCoral, incalGold, float64(4-facetIdx)/4.0*0.58)
					if inInnerGem {
						sunC = incalGold
					}
					midC = incalViolet
				}

				obj := 30
				if !isUpper {
					obj = 31
				}
				return render.MoebiusPixel{
					Depth:           1.75 - 0.35*(1.0-tPyr),
					ObjID:           obj,
					PartID:          obj*100 + facetIdx*2 + int(moebiusBoolFloat(inInnerGem)),
					DetailID:        tierStep*8 + facetIdx,
					Normal:          n,
					FlowTangent:     [2]float64{0.94, 0.34 * fnx},
					Illum:           lit,
					Spec:            common.Smoothstep(0.75, 0.96, lit) * 0.82,
					Bounce:          0.45,
					SunColor:        sunC,
					MidColor:        midC,
					ShadeColor:      incalShade,
					DeepShadowColor: incalDeep,
					CrackInk:        stone.CrackInk * 0.45,
				}
			}
		}

		if absY < gapHalf && math.Abs(x) < baseW*0.94 {
			return render.MoebiusPixel{
				Depth:      1.85,
				ObjID:      32,
				PartID:     320,
				DetailID:   int(math.Floor((x + 1.0) * 10.0)),
				Normal:     [3]float64{0, 0, 1},
				Illum:      0.98,
				Spec:       0.92,
				SunColor:   abyssCore,
				MidColor:   incalGold,
				ShadeColor: incalGold,
			}
		}

		// =====================================================================
		// 2. 3D BEVELED GYROSCOPIC ASTROLABE RINGS (Cohesive Bone-Ivory & Brass)
		// =====================================================================
		numRings := mi.ringLayers
		if numRings > 4 {
			numRings = 4
		}
		for k := 0; k < numRings; k++ {
			fk := float64(k)
			tiltAng := 0.38*(fk-1.0) + 0.08*math.Sin(phi+fk*1.7)
			cosA, sinA := math.Cos(tiltAng), math.Sin(tiltAng)
			rx := x*cosA - y*sinA
			ry := (x*sinA + y*cosA) / (0.52 + 0.14*fk)
			rEll := math.Hypot(rx, ry)

			rTarget := (0.41 + 0.145*fk) * scale
			bandHalfW := 0.032 * scale
			if math.Abs(rEll-rTarget) <= bandHalfW {
				ringTheta := math.Atan2(ry, rx) + (1.0-2.0*float64(k&1))*phi/12.0
				panelIdx := int(math.Floor((ringTheta + math.Pi) / (2.0 * math.Pi) * 28.0))
				panelFrac := ((ringTheta+math.Pi)/(2.0*math.Pi)*28.0) - float64(panelIdx)
				if panelIdx%7 == 0 && panelFrac < 0.22 {
					continue
				}
				uTube := (rEll - rTarget) / bandHalfW
				nz := math.Sqrt(math.Max(0.12, 1.0-uTube*uTube*0.85))
				n := moebiusNorm3(rx*0.52, uTube*0.68, nz)
				lit := common.Clamp01(0.18 + 0.86*math.Max(0.0, moebiusDot3(n, lightDir)))
				if rx > 0.24 {
					lit *= 0.58
				}

				sCol := moebiusLerp3(ringIvory, ringBrass, 0.45+0.45*math.Abs(uTube))
				if math.Abs(uTube) < 0.24 && panelIdx%3 == 0 {
					sCol = incalCyan
					lit = math.Max(lit, 0.88)
				}

				tangX := -math.Sin(ringTheta)
				tangY := math.Cos(ringTheta) * (0.52 + 0.14*fk)
				return render.MoebiusPixel{
					Depth:           2.6 + 0.35*fk,
					ObjID:           20 + k,
					PartID:          (20+k)*100 + (panelIdx & 15),
					DetailID:        panelIdx*2 + int(moebiusBoolFloat(math.Abs(uTube) < 0.28)),
					Normal:          n,
					FlowTangent:     [2]float64{tangX*cosA + tangY*sinA, -tangX*sinA + tangY*cosA},
					Illum:           lit,
					Spec:            common.Smoothstep(0.75, 0.96, lit) * 0.68,
					Bounce:          0.38,
					SunColor:        sCol,
					MidColor:        shaftSunCoral,
					ShadeColor:      ringShade,
					DeepShadowColor: incalDeep,
					AtmosColor:      atmosGlow,
				}
			}
		}

		// =====================================================================
		// 3. COHESIVE 3D PERSPECTIVE TECHNO-CITY MEGASTRUCTURE CHASM
		//    (Sunlit Ochre/Terracotta on Upper-Left -> Deep Prussian-Indigo on Lower-Right!)
		// =====================================================================
		const fSec = 16.0
		secCoord := (theta + math.Pi) / (2.0 * math.Pi) * fSec
		secIdx := int(math.Floor(secCoord))
		secFrac := secCoord - float64(secIdx)
		secSym := math.Abs(secFrac*2.0 - 1.0)

		if r > 0.22 {
			logR := math.Log(r/0.22) * 4.2
			tierIdx := int(math.Floor(logR))
			tierFrac := logR - float64(tierIdx)

			isButtress := secSym > 0.72
			isLedge := tierFrac < 0.18

			stone := render.EvaluateMoebiusStoneRelief(ne, theta*2.5, logR*0.6, 6.0, 6.0, 0.3)

			// Cohesive 3D cylindrical chasm normal:
			// The shaft interior at angle theta has inward-pointing normal (-cos(theta), -sin(theta))!
			// Sunlight from upper-left (-0.62, -0.50) illuminates the lower-right interior wall
			// while the upper-left overhanging wall is in dramatic chiaroscuro shadow!
			inwardNx := -math.Cos(theta) + 0.18*stone.NormalMod[0]
			inwardNy := -math.Sin(theta) + 0.18*stone.NormalMod[1]
			tangNx := -math.Sin(theta)
			tangNy := math.Cos(theta)
			if secFrac < 0.5 {
				tangNx = -tangNx
				tangNy = -tangNy
			}

			var n [3]float64
			if isLedge {
				n = moebiusNorm3(inwardNx*0.75, inwardNy*0.75, 0.65)
			} else if isButtress {
				n = moebiusNorm3(tangNx*0.72+inwardNx*0.30, tangNy*0.72+inwardNy*0.30, 0.62)
			} else {
				n = moebiusNorm3(inwardNx*0.65+tangNx*0.25, inwardNy*0.65+tangNy*0.25, 0.72)
			}

			lit := common.Clamp01(0.12 + 0.88*math.Max(0.0, moebiusDot3(n, lightDir)))
			subWindowU := int(math.Floor(secFrac * 4.0))
			subWindowV := int(math.Floor(tierFrac * 3.0))
			inWindowArch := !isButtress && !isLedge && subWindowU >= 1 && subWindowU <= 2 && subWindowV == 1
			if inWindowArch {
				lit *= 0.45
			}

			// Cohesive, continuous Mœbius shaft gradient:
			// Warm sunlit limestone (#f9d888) and terracotta (#ec7852) smoothly varying with radius & angle
			warmMix := 0.5 + 0.5*math.Sin(logR*0.8+0.2*stone.ColorMod)
			sunC := moebiusLerp3(shaftSunCoral, shaftSunGold, warmMix)
			if isButtress || isLedge {
				sunC = shaftSunGold
			} else if inWindowArch && ((tierIdx+secIdx)%6 == 0) {
				windowPulse := 0.55 + 0.45*math.Cos(phi-float64(tierIdx)*0.9)
				sunC = moebiusLerp3(shaftSunGold, incalCyan, windowPulse)
				lit = 0.78 + 0.18*windowPulse
			}

			atmos := common.Smoothstep(0.76, 0.22, r) * 0.58
			subPart := (tierIdx&7)*16 + (secIdx & 15)
			if isButtress {
				subPart += 128
			}
			if isLedge {
				subPart += 256
			}

			return render.MoebiusPixel{
				Depth:           10.0 - 5.8*common.Clamp01(r) - 0.35*moebiusBoolFloat(isButtress),
				ObjID:           10 + (secIdx & 3),
				PartID:          1000 + subPart,
				DetailID:        stone.BlockID,
				Normal:          n,
				FlowTangent:     [2]float64{-math.Sin(theta), math.Cos(theta)},
				Illum:           lit,
				Spec:            moebiusBoolFloat(isLedge) * 0.55,
				Bounce:          common.Smoothstep(0.65, 0.22, r) * 0.55,
				Atmos:           atmos,
				SunColor:        sunC,
				MidColor:        shaftMid,
				ShadeColor:      shaftShade,
				DeepShadowColor: incalDeep,
				AtmosColor:      atmosGlow,
				CrackInk:        stone.CrackInk * 0.55,
				TickMask:        stone.PockInk * 0.5,
				Stipple:         atmos * 0.48,
			}
		}

		// Central luminous abyssal star-core behind the Incal crystal (r <= 0.22)
		coreT := common.Clamp01(r / 0.22)
		return render.MoebiusPixel{
			Depth:       30.0,
			ObjID:       2,
			PartID:      20,
			DetailID:    20,
			Normal:      [3]float64{0, 0, 1},
			FlowTangent: [2]float64{-math.Sin(theta), math.Cos(theta)},
			Illum:       0.94,
			Spec:        0.85 * (1.0 - coreT),
			SunColor:    moebiusLerp3(abyssCore, incalCyan, coreT*0.72),
			MidColor:    incalGold,
			ShadeColor:  shaftShade,
			Stipple:     coreT * 0.52,
		}
	})

	c.MergeRenderCanvas(rc)
}
