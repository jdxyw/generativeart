package arts

import (
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type moebiusStarwatcher struct {
	rings     int
	sails     int
	orbitTilt float64
}

// NewMoebiusStarwatcher returns a MoebiusStarwatcher generator inspired by Jean
// "Mœbius" Giraud's Starwatcher, Chroniques Métalliques, and Dune concept art:
// an asymmetric weathered bone-ivory and ochre-brass relic ring-citadel with
// ribbed biomechanical carapace plates (zero rainbow checkerboard!), a sweeping
// organic coral-gold cosmic nebula veil, and jagged boulder-strewn desert badlands
// rising asymmetrically in the foreground.
func NewMoebiusStarwatcher(rings, sails int, orbitTilt float64) *moebiusStarwatcher {
	if rings <= 0 {
		rings = 3
	}
	if sails <= 0 {
		sails = 4
	}
	if orbitTilt == 0 {
		orbitTilt = -0.24
	}
	return &moebiusStarwatcher{
		rings:     rings,
		sails:     sails,
		orbitTilt: orbitTilt,
	}
}

// Generative renders the static MoebiusStarwatcher Ligne Claire print.
func (ms *moebiusStarwatcher) Generative(c *generativeart.Canva) {
	ms.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless looping MoebiusStarwatcher animation.
func (ms *moebiusStarwatcher) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	seed := c.Seed()
	if seed == 0 {
		seed = 20262104
	}
	ne := common.NewNoiseEngine(seed)
	phi := 2.0 * math.Pi * phase

	// Authentic Starwatcher / Chroniques Métalliques Palette (Cohesive Bone-Ivory, Ochre, Coral & Prussian Indigo)
	spaceDeep := moebiusLinRGB(0x0d, 0x1a, 0x3a)     // #0d1a3a Deep Starwatcher Indigo Night
	spaceMid := moebiusLinRGB(0x1e, 0x4e, 0x7a)      // #1e4e7a Cerulean Twilight
	spaceHorizon := moebiusLinRGB(0xdc, 0x98, 0x68)  // #dc9868 Warm Amber-Apricot Badlands Glow

	nebulaGold := moebiusLinRGB(0xfc, 0xdc, 0x78)    // #fcdc78 Sunlit Golden Cloud Core
	nebulaCoral := moebiusLinRGB(0xee, 0x6e, 0x54)   // #ee6e54 Warm Coral-Terracotta Cloud Billow
	nebulaViolet := moebiusLinRGB(0x5e, 0x52, 0x8e)  // #5e528e Lavender-Indigo Cloud Underside

	hullIvory := moebiusLinRGB(0xf8, 0xf0, 0xda)     // #f8f0da Weathered Bone-Ivory Spacecraft Hull
	hullOchre := moebiusLinRGB(0xec, 0xb2, 0x56)     // #ecb256 Warm Brass-Ochre Armor Band
	hullVermilion := moebiusLinRGB(0xdc, 0x48, 0x36) // #dc4836 Vermilion Squadron Marking
	hullCyan := moebiusLinRGB(0x52, 0xe4, 0xd8)      // #52e4d8 Glowing Reactor Slit
	hullMid := moebiusLinRGB(0x9a, 0x76, 0x6e)       // #9a766e Weathered Umber-Rose Hull Midtone
	hullShade := moebiusLinRGB(0x2c, 0x3e, 0x74)     // #2c3e74 Cobalt-Slate Hull Shadow
	hullDeep := moebiusLinRGB(0x12, 0x18, 0x32)      // #121832 Deep Hull Crevice Core

	rockSun := moebiusLinRGB(0xf6, 0xbe, 0x68)       // #f6be68 Sunlit Badlands Sandstone
	rockVermilion := moebiusLinRGB(0xdc, 0x58, 0x3e) // #dc583e Red-Rock Canyon Face
	rockShade := moebiusLinRGB(0x32, 0x3c, 0x6e)     // #323c6e Indigo Badlands Shadow

	lightDir := moebiusNorm3(-0.62, -0.44, 0.65)

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)
	cfg := render.DefaultMoebiusConfig(seed)

	render.RenderMoebiusScene(rc, cfg, func(ux, uy float64) render.MoebiusPixel {
		x := ux*2.0 - 1.0
		y := uy*2.0 - 1.0

		crag := ne.FBM2D(x*4.5+1.7, y*4.5-0.9, 4, 2.0, 0.5)

		// =====================================================================
		// 1. FOREGROUND ASYMMETRIC CRAGGY DESERT BADLANDS & PINNACLES
		//    (Rising high on the right to y = 0.18, dipping into a canyon on the left)
		// =====================================================================
		fbmMtn, dFbmDx, _ := ne.FBM2DDeriv(x*4.5+1.7, y*4.5-0.9, 4, 2.0, 0.5)
		badlandsY := 0.58 - 0.24*x - 0.14*math.Sin(x*3.6-0.5) + 0.022*fbmMtn*1.3
		dBadlandsDx := -0.24 - 0.14*3.6*math.Cos(x*3.6-0.5) + 0.028*dFbmDx

		// Towering jagged rock pinnacle on the far right (x ~ 0.74) and lower-left (x ~ -0.76)
		dRight := (x - 0.74) / 0.24
		if math.Abs(dRight) < 1.6 {
			expR := math.Exp(-dRight * dRight * 1.8)
			badlandsY -= 0.34 * expR
			dBadlandsDx -= 0.34 * (-3.6 * dRight / 0.24) * expR
		}
		dLeft := (x - (-0.76)) / 0.22
		if math.Abs(dLeft) < 1.6 {
			expL := math.Exp(-dLeft * dLeft * 1.8)
			badlandsY -= 0.24 * expL
			dBadlandsDx -= 0.24 * (-3.6 * dLeft / 0.22) * expL
		}

		if y >= badlandsY {
			dyBelow := y - badlandsY
			stone := render.EvaluateMoebiusStoneRelief(ne, x, y, 7.8, 8.8, 0.42)

			// Smooth 3D ridge normal from analytical slope derivative dBadlandsDx + eroded gullies
			// (dBadlandsDx < 0 means ridge rises to the right -> faces left toward the sun!)
			gullySlope := 0.35 * math.Sin(x*14.0+y*6.0+1.8*fbmMtn)
			nx := common.Constrain(0.52*dBadlandsDx+gullySlope+0.26*stone.NormalMod[0], -0.92, 0.92)
			ny := -0.25 + 0.24*stone.NormalMod[1]
			if dyBelow < 0.032 {
				ny = -0.78
			}
			nz := math.Sqrt(math.Max(0.12, 1.0-common.Clamp01(nx*nx+ny*ny)*0.85))
			n := moebiusNorm3(nx, ny, nz)

			lit := common.Clamp01(0.14 + 0.88*math.Max(0.0, moebiusDot3(n, lightDir)))
			sunC := moebiusLerp3(rockVermilion, rockSun, common.Smoothstep(0.85, 0.20, y)+0.08*stone.ColorMod)
			organicLayer := int(math.Floor(dyBelow*4.5 + 0.8*stone.ColorMod))

			return render.MoebiusPixel{
				Depth:           1.8 + 0.2*dyBelow,
				ObjID:           10,
				PartID:          100 + (organicLayer & 15),
				DetailID:        stone.BlockID,
				Normal:          n,
				FlowTangent:     [2]float64{0.95, 0.25 * nx},
				Illum:           lit,
				Spec:            common.Smoothstep(0.04, 0.0, dyBelow) * 0.60,
				Bounce:          0.42,
				SunColor:        sunC,
				MidColor:        rockVermilion,
				ShadeColor:      rockShade,
				DeepShadowColor: hullDeep,
				CrackInk:        stone.CrackInk,
				TickMask:        stone.PockInk,
			}
		}

		// =====================================================================
		// 2. COLOSSAL WEATHERED BONE-IVORY RELIC CITADEL & RIBBED CARAPACE RINGS
		//    (Tilted off-center at (-0.08, -0.14) — cohesive bone/ochre armor, zero disco!)
		// =====================================================================
		sx := x - (-0.08)
		sy := y - (-0.14 + 0.015*math.Sin(phi))
		cosT, sinT := math.Cos(ms.orbitTilt), math.Sin(ms.orbitTilt)
		rx := sx*cosT - sy*sinT
		ry := sx*sinT + sy*cosT

		rCore := math.Hypot(rx, ry*1.15)
		inCore := rCore <= 0.165

		type relicRing struct {
			rad      float64
			halfW    float64
			wallH    float64
			tiltY    float64
			rotSpeed float64
			id       int
		}
		aRings := [2]relicRing{
			{rad: 0.58, halfW: 0.062, wallH: 0.052, tiltY: 0.40, rotSpeed: 1.0, id: 20},
			{rad: 0.36, halfW: 0.048, wallH: 0.042, tiltY: 0.54, rotSpeed: -1.0, id: 21},
		}

		checkRelicRings := func(frontPass bool) (render.MoebiusPixel, bool) {
			for idx := range aRings {
				ar := &aRings[idx]
				inFront := ry >= 0.0
				if inFront != frontPass {
					continue
				}

				// 2a. Curved Ribbed Carapace Top Deck (like Mœbius's weathered bone/ceramic hulls!)
				eyTop := ry / ar.tiltY
				rTop := math.Hypot(rx, eyTop)
				if math.Abs(rTop-ar.rad) <= ar.halfW {
					ang := math.Atan2(eyTop, rx) + ar.rotSpeed*phi/16.0
					uDeck := (rTop - ar.rad) / ar.halfW // in [-1, 1] across ring width

					// Fluted biomechanical ribs along the ring circumference
					ribCoord := (ang + math.Pi) / (2.0 * math.Pi) * 36.0
					ribIdx := int(math.Floor(ribCoord))
					ribFrac := ribCoord - float64(ribIdx)
					ribPillow := 0.22 * math.Cos(ribFrac*math.Pi*2.0)

					stone := render.EvaluateMoebiusStoneRelief(ne, ang*1.5, uDeck, 8.0, 6.0, 0.25)

					nx := rx/ar.rad*0.45 + 0.12*stone.NormalMod[0]
					ny := -0.52 + 0.45*uDeck + ribPillow
					nz := math.Sqrt(math.Max(0.12, 1.0-common.Clamp01(nx*nx+ny*ny)*0.85))
					n := moebiusNorm3(nx, ny, nz)
					lit := common.Clamp01(0.20 + 0.84*math.Max(0.0, moebiusDot3(n, lightDir)))
					if rx > 0.25*ar.rad {
						lit *= 0.58
					}

					// Cohesive weathered bone-ivory & brass-ochre hull (with occasional vermilion squadron band)
					inReactorSlit := math.Abs(uDeck) < 0.18 && ribIdx%3 == 0
					sCol := hullIvory
					if inReactorSlit {
						sCol = hullCyan
						lit = 0.94
					} else if ribIdx%12 == 0 {
						sCol = hullVermilion
					} else if math.Abs(uDeck) > 0.65 {
						sCol = hullOchre
					}

					dVal := 2.4 + 0.3*float64(idx)
					if !inFront {
						dVal += 1.4
					}
					tangX := -math.Sin(ang) * ar.rad
					tangY := math.Cos(ang) * ar.rad * ar.tiltY
					return render.MoebiusPixel{
						Depth:           dVal,
						ObjID:           ar.id,
						PartID:          ar.id*100 + (ribIdx & 31),
						DetailID:        ribIdx*4 + int(moebiusBoolFloat(inReactorSlit)),
						Normal:          n,
						FlowTangent:     [2]float64{tangX*cosT + tangY*sinT, -tangX*sinT + tangY*cosT},
						Illum:           lit,
						Spec:            common.Smoothstep(0.75, 0.96, lit) * 0.70,
						Bounce:          0.42,
						SunColor:        sCol,
						MidColor:        hullMid,
						ShadeColor:      hullShade,
						DeepShadowColor: hullDeep,
						CrackInk:        stone.CrackInk * 0.55,
						TickMask:        stone.PockInk * 0.65,
					}, true
				}

				// 2b. Vertical Beveled Under-Hull Skirt Wall with Circular Docking Portholes
				for step := 1; step <= 4; step++ {
					yShift := float64(step) * 0.25 * ar.wallH
					eySide := (ry - yShift) / ar.tiltY
					rSide := math.Hypot(rx, eySide)
					if math.Abs(rSide-ar.rad) <= ar.halfW {
						ang := math.Atan2(eySide, rx) + ar.rotSpeed*phi/16.0
						ribIdx := int(math.Floor((ang + math.Pi) / (2.0 * math.Pi) * 36.0))
						n := moebiusNorm3(rx/ar.rad, 0.45, 0.58)
						lit := common.Clamp01(0.14 + 0.76*math.Max(0.0, moebiusDot3(n, lightDir)))
						if !inFront || rx > 0.08*ar.rad {
							lit = 0.24
						}
						dVal := 2.5 + 0.3*float64(idx)
						if !inFront {
							dVal += 1.4
						}
						return render.MoebiusPixel{
							Depth:           dVal,
							ObjID:           ar.id,
							PartID:          ar.id*100 + 64 + (ribIdx & 31),
							DetailID:        ribIdx*4 + step,
							Normal:          n,
							FlowTangent:     [2]float64{-sinT, cosT},
							Illum:           lit,
							Bounce:          0.45,
							SunColor:        hullOchre,
							MidColor:        hullMid,
							ShadeColor:      hullShade,
							DeepShadowColor: hullDeep,
						}, true
					}
				}
			}
			return render.MoebiusPixel{}, false
		}

		if smp, ok := checkRelicRings(true); ok {
			return smp
		}

		// 2c. Central Ribbed Nautilus Command Core (Cohesive Bone-Ivory & Brass, NOT a checkerboard!)
		if inCore {
			uCore := rx / 0.165
			vCore := (ry * 1.15) / 0.165
			zCore := math.Sqrt(math.Max(0.05, 1.0-uCore*uCore-vCore*vCore))
			stone := render.EvaluateMoebiusStoneRelief(ne, uCore, vCore, 6.0, 6.0, 0.25)

			// Smooth latitudinal armor bands + equatorial glowing observation visor
			latBand := int(math.Floor((vCore + 1.0) * 4.5))
			n := moebiusNorm3(uCore*cosT+vCore*sinT+0.12*stone.NormalMod[0], -uCore*sinT+vCore*cosT+0.12*stone.NormalMod[1], zCore)
			lit := common.Clamp01(0.16 + 0.86*math.Max(0.0, moebiusDot3(n, lightDir)))

			sCol := hullIvory
			if math.Abs(vCore) < 0.14 {
				sCol = hullCyan
				lit = 0.94
			} else if latBand == 2 || latBand == 6 {
				sCol = hullOchre
			}
			return render.MoebiusPixel{
				Depth:           3.0 - 0.25*zCore,
				ObjID:           25,
				PartID:          2500 + latBand,
				DetailID:        stone.BlockID,
				Normal:          n,
				FlowTangent:     [2]float64{cosT, sinT},
				Illum:           lit,
				Spec:            common.Smoothstep(0.75, 0.96, lit) * 0.78,
				Bounce:          0.42,
				SunColor:        sCol,
				MidColor:        hullMid,
				ShadeColor:      hullShade,
				DeepShadowColor: hullDeep,
				CrackInk:        stone.CrackInk * 0.55,
				TickMask:        stone.PockInk * 0.65,
			}
		}

		// Axial Spire Mast & Antenna Needles
		if math.Abs(ry) <= 0.52 && math.Abs(rx) <= 0.024*(1.0-math.Abs(ry)/0.56) {
			uSpire := rx / 0.024
			n := moebiusNorm3(uSpire*0.8, 0, 0.6)
			lit := common.Clamp01(0.18 + 0.84*math.Max(0.0, moebiusDot3(n, lightDir)))
			tier := int(math.Abs(ry) * 16.0)
			return render.MoebiusPixel{
				Depth:           3.1,
				ObjID:           26,
				PartID:          2600 + tier,
				DetailID:        tier*2 + int(moebiusBoolFloat(rx > 0)),
				Normal:          n,
				FlowTangent:     [2]float64{-sinT, cosT},
				Illum:           lit,
				SunColor:        hullOchre,
				MidColor:        hullMid,
				ShadeColor:      hullShade,
				DeepShadowColor: hullDeep,
			}
		}

		if smp, ok := checkRelicRings(false); ok {
			return smp
		}

		// =====================================================================
		// 3. BILLOWING MŒBIUS NEBULA CLOUD BANK, CRESCENT MOON & DEEP SPACE
		// =====================================================================
		// Crescent Moon in Upper-Right Sky ((0.56, -0.58), radius 0.095)
		pxM, pyM := x-0.56, y-(-0.58)
		rM := math.Hypot(pxM, pyM)
		if rM <= 0.095 {
			nxM, nyM := pxM/0.095, pyM/0.095
			nzM := math.Sqrt(math.Max(0.05, 1.0-nxM*nxM-nyM*nyM))
			lit := common.Clamp01(0.20 + 0.84*math.Max(0.0, moebiusDot3(moebiusNorm3(nxM, nyM, nzM), lightDir)))
			stoneM := render.EvaluateMoebiusStoneRelief(ne, x*2.5, y*2.5, 10.0, 10.0, 0.3)
			return render.MoebiusPixel{
				Depth:           68.0,
				ObjID:           60,
				PartID:          600,
				DetailID:        stoneM.BlockID,
				Normal:          moebiusNorm3(nxM, nyM, nzM),
				FlowTangent:     [2]float64{0.92, -0.38 * nxM},
				Illum:           lit,
				Spec:            common.Smoothstep(0.35, 0.05, nzM) * 0.55,
				SunColor:        hullIvory,
				MidColor:        nebulaGold,
				ShadeColor:      nebulaViolet,
				DeepShadowColor: spaceDeep,
				CrackInk:        stoneM.CrackInk * 0.45,
			}
		}

		// Organic Billowing Cosmic Cloud Bank sweeping diagonally behind the vessel
		// (No grid lines! Pure organic scalloped Mœbius cloud billows with contour hachures!)
		cloudCenter := -0.36*x - 0.14*math.Sin(x*2.4) + 0.04
		cloudDist := y - cloudCenter
		cloudHalfW := 0.22 + 0.08*crag
		if math.Abs(cloudDist) <= cloudHalfW {
			uCloud := (cloudDist/cloudHalfW + 1.0) * 0.5 // 0 at sunlit upper crest, 1 at shadowed lower belly
			lobePhase := x*5.2 + 1.2*crag
			lobeIdx := int(math.Floor(lobePhase))
			lobeFrac := lobePhase - float64(lobeIdx) - 0.5

			nxC := lobeFrac * 0.65
			nyC := uCloud*1.6 - 0.75
			nzC := math.Sqrt(math.Max(0.15, 1.0-common.Clamp01(nxC*nxC+nyC*nyC)*0.8))
			nC := moebiusNorm3(nxC, nyC, nzC)
			lit := common.Clamp01(0.22 + 0.82*math.Max(0.0, moebiusDot3(nC, lightDir)))

			return render.MoebiusPixel{
				Depth:           82.0,
				ObjID:           4,
				PartID:          40 + (lobeIdx & 15),
				DetailID:        int(uCloud * 3.0),
				Normal:          nC,
				FlowTangent:     [2]float64{0.92, -0.36},
				Illum:           lit,
				Spec:            common.Smoothstep(0.25, 0.0, uCloud) * 0.65,
				SunColor:        nebulaGold,
				MidColor:        nebulaCoral,
				ShadeColor:      nebulaViolet,
				DeepShadowColor: spaceDeep,
				Stipple:         common.Smoothstep(0.45, 0.95, uCloud) * 0.48,
			}
		}

		skyT := common.Clamp01((y + 1.0) / 1.55)
		var skyCol [3]float64
		if skyT < 0.62 {
			skyCol = moebiusLerp3(spaceDeep, spaceMid, skyT/0.62)
		} else {
			skyCol = moebiusLerp3(spaceMid, spaceHorizon, (skyT-0.62)/0.38)
		}

		return render.MoebiusPixel{
			Depth:           100.0,
			ObjID:           1,
			PartID:          1,
			DetailID:        1,
			Normal:          [3]float64{0, 0, 1},
			FlowTangent:     [2]float64{1, 0},
			Illum:           0.84,
			SunColor:        skyCol,
			MidColor:        spaceMid,
			ShadeColor:      spaceDeep,
			DeepShadowColor: spaceDeep,
			Stipple:         common.Smoothstep(0.14, 0.0, math.Abs(skyT-0.58)) * 0.38,
		}
	})

	c.MergeRenderCanvas(rc)
}
