package arts

import (
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type calabiYauManifold struct {
	foldOrder  int
	rings      int
	phaseMorph float64
}

// NewCalabiYauManifold returns a CalabiYauManifold generator that raymarches a
// 3D string-theory Calabi-Yau / Clebsch diagonal algebraic surface manifold with
// hollowed genus-tunnel handles, nested inside twin orthogonal gilded geodesic
// astrolabe rings and orbited by 4 topological monopole spheres, shaded with
// two-sided Kubelka-Munk Lapis/Malachite/Cinnabar/Gold-Leaf glazes and 3D
// object-locked copperplate burin cross-hatching.
func NewCalabiYauManifold(foldOrder, rings int, phaseMorph float64) *calabiYauManifold {
	if foldOrder <= 0 {
		foldOrder = 5
	}
	if rings <= 0 {
		rings = 2
	}
	if phaseMorph == 0 {
		phaseMorph = 0.35
	}
	return &calabiYauManifold{
		foldOrder:  foldOrder,
		rings:      rings,
		phaseMorph: phaseMorph,
	}
}

// Generative renders the static CalabiYauManifold sculpture & celestial print.
func (cy *calabiYauManifold) Generative(c *generativeart.Canva) {
	cy.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360° looping CalabiYauManifold animation.
func (cy *calabiYauManifold) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	seed := c.Seed()
	if seed == 0 {
		seed = 20262304
	}
	ne := common.NewNoiseEngine(seed)
	phi := 2.0 * math.Pi * phase

	// Kubelka-Munk Multi-Sheet Algebraic Manifold & Renaissance Celestial Observatory Palette
	goldLeaf := render.MixKubelkaMunk(render.PW(render.PigmentGoldLeaf, 0.68), render.PW(render.PigmentTitaniumWhite, 0.32))
	amberBronze := render.MixKubelkaMunk(render.PW(render.PigmentGoldLeaf, 0.68), render.PW(render.PigmentBurntSienna, 0.32))
	cinnabarRidge := render.MixKubelkaMunk(render.PW(render.PigmentVermilion, 0.74), render.PW(render.PigmentCadmiumYellow, 0.26))
	roseCoral := render.MixKubelkaMunk(render.PW(render.PigmentQuinacridoneRose, 0.68), render.PW(render.PigmentTitaniumWhite, 0.32))
	cobaltTeal := render.MixKubelkaMunk(render.PW(render.PigmentCobaltTeal, 0.72), render.PW(render.PigmentTitaniumWhite, 0.28))
	azuriteCavity := render.MixKubelkaMunk(render.PW(render.PigmentAzurite, 0.75), render.PW(render.PigmentCobaltTeal, 0.25))
	imperialViolet := render.MixKubelkaMunk(render.PW(render.PigmentImperialPurple, 0.68), render.PW(render.PigmentUltramarine, 0.32))
	malachiteSaddle := render.MixKubelkaMunk(render.PW(render.PigmentMalachite, 0.72), render.PW(render.PigmentTitaniumWhite, 0.28))
	warmParchment := render.MixKubelkaMunk(render.PW(render.PigmentYellowOchre, 0.38), render.PW(render.PigmentTitaniumWhite, 0.62))
	umberShade := render.MixKubelkaMunk(render.PW(render.PigmentRawUmber, 0.56), render.PW(render.PigmentUltramarine, 0.44))
	deepIntaglio := render.MixKubelkaMunk(render.PW(render.PigmentPineSootInk, 0.78), render.PW(render.PigmentUltramarine, 0.22))

	lightDir3D := render.V3(-0.55, 0.58, 0.60).Normalize()

	// Exact symmetry-matched rotation angles for seamless 20-frame loop:
	//  - SDF3DClebschCalabiYau evaluated on FoldOctahedral(q) has exact 4-fold symmetry around Y,
	//    so rotating around Y by +(2*pi/4)*phase (90 deg per cycle) is 100% periodic!
	//  - Morph phase oscillates smoothly by sin(2*pi*phase) (periodic over [0, 1))!
	//  - 4 Orbiting Monopole Spheres counter-rotate around Y by -(2*pi/4)*phase!
	//  - Inner Geodesic Astrolabe Ring tumbles by +pi*phase!
	rotManifoldY := 0.30 + (2.0*math.Pi/4.0)*phase
	rotMonopolesY := -0.25 - (2.0*math.Pi/4.0)*phase
	rotGimbalY := 0.40 + math.Pi*phase
	morph := cy.phaseMorph + 0.26*math.Sin(phi)
	tiltX := 0.32

	evalManifoldScene := func(p render.Vec3) (float64, int, int, render.Vec3) {
		// 1. Central Calabi-Yau / Clebsch Algebraic Surface Manifold (Octahedral 4-Fold Symmetric)
		qMan := p.RotateX(tiltX).RotateY(rotManifoldY)
		fOct := render.FoldOctahedral(qMan)
		dCalabi := render.SDF3DClebschCalabiYau(fOct, 0.44, morph, 0.036)

		// 2. Outer Equatorial Geodesic Astrolabe Ring
		qRing1 := p.RotateZ(0.24).RotateX(0.36)
		dRing1 := render.SDF3DTorus(qRing1, 0.58, 0.022)

		// 3. Inner Tumbling Meridian Geodesic Ring
		qRing2 := p.RotateZ(0.24).RotateY(rotGimbalY).RotateX(math.Pi * 0.5)
		dRing2 := render.SDF3DTorus(qRing2, 0.51, 0.018)

		// 4. Four Orbiting Topological Monopole Spheres
		qMono := p.RotateX(tiltX).RotateY(rotMonopolesY)
		fMono, monoIdx := render.FoldPolarModXZ(qMono, 4)
		pSphere := fMono.Sub(render.V3(0.45, 0.0, 0.0))
		dMonopole := render.SDF3DSphere(pSphere, 0.052)

		minD := dCalabi
		objID := 30
		partID := 300 + int(math.Floor((fOct.X+fOct.Y)*9.0))
		loc := fOct

		if dRing1 < minD {
			minD = dRing1
			objID = 31
			ang := math.Atan2(qRing1.Z, qRing1.X) + phi/24.0
			seg := int(math.Floor((ang + math.Pi) / (2.0 * math.Pi) * 24.0))
			partID = 340 + (seg & 15)
			loc = qRing1
		}
		if dRing2 < minD {
			minD = dRing2
			objID = 32
			ang := math.Atan2(qRing2.Z, qRing2.X)
			seg := int(math.Floor((ang + math.Pi) / (2.0 * math.Pi) * 20.0))
			partID = 360 + (seg & 15)
			loc = qRing2
		}
		if dMonopole < minD {
			minD = dMonopole
			objID = 33
			partID = 380 + monoIdx
			loc = pSphere
		}
		return minD, objID, partID, loc
	}

	sceneSDF := func(p render.Vec3) float64 {
		d, _, _, _ := evalManifoldScene(p)
		return d
	}

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)
	cfg := render.DefaultMoebiusConfig(seed)
	cfg.HatchSpacingPx = 3.5

	render.RenderMoebiusScene(rc, cfg, func(ux, uy float64) render.MoebiusPixel {
		x := ux*2.0 - 1.0
		y := uy*2.0 - 1.0

		// =====================================================================
		// 1. RAYMARCH 3D CALABI-YAU ALGEBRAIC MANIFOLD & GEODESIC RINGS
		// =====================================================================
		rScreen := math.Hypot(x, y)
		if rScreen <= 0.78 {
			ro := render.V3(x*0.86, -y*0.86, 2.2)
			rd := render.V3(0, 0, -1)
			hit := render.RaymarchSDF3D(sceneSDF, ro, rd, lightDir3D, 3.4, 68)
			if hit.Hit {
				_, objID, partID, pLoc := evalManifoldScene(hit.Pos)
				qMan := hit.Pos.RotateX(tiltX).RotateY(rotManifoldY)
				nLoc := hit.Normal.RotateX(tiltX).RotateY(rotManifoldY)
				stone := render.EvaluateMoebiusStoneRelief3D(ne, pLoc.Scale(2.5), nLoc, 7.5, 0.25)
				// Add warm upward bounce illumination so interior genus tunnels glow with rich mineral chroma!
				upBounce := math.Max(0.0, -hit.Normal.Y) * 0.36
				lit := common.Clamp01((0.20 + 0.78*math.Max(0.0, hit.Normal.Dot(lightDir3D)) + upBounce) *
					(0.50 + 0.50*hit.Shadow) * (0.62 + 0.38*hit.AO))
				burinInk, burinChalk := render.Evaluate3DBurinEngraving(pLoc, nLoc, lit, 86.0)

				var sCol, mCol [3]float64
				switch objID {
				case 30:
					// Multi-Branch Calabi-Yau Riemann Sheet Coloring in 3D Object-Locked Coordinates:
					//  - Interior genus tunnels & deep negative-curvature saddles: Royal Azurite, Cobalt Teal & Imperial Violet
					//  - 4-fold symmetric harmonic sheets (cos(4*theta) & sin(6*phi)): Malachite Emerald, Cinnabar Vermilion, Rose-Coral & 24k Gold Leaf
					rMan := pLoc.Len()
					theta4 := math.Atan2(qMan.Z, qMan.X) * 4.0 // 4-fold invariant under 90-deg Y rotation!
					harmonicSheet := 0.5 + 0.5*math.Sin(theta4+qMan.Y*7.5)
					branchBand := int(math.Floor((pLoc.X + pLoc.Y + pLoc.Z) * 9.0))
					partID = 300 + (branchBand & 31)

					if rMan < 0.34 || hit.Curvature < -0.12 {
						// Interior genus handle tunnels & concave saddle portals
						uCav := common.Clamp01((0.36-rMan)*4.5 + math.Max(0.0, -hit.Curvature))
						sCol = moebiusLerp3(cobaltTeal, azuriteCavity, uCav)
						mCol = imperialViolet
					} else if harmonicSheet < 0.34 {
						// Sheet 1: Malachite Emerald & Cobalt Turquoise harmonic lobe
						u := harmonicSheet / 0.34
						sCol = moebiusLerp3(cobaltTeal, malachiteSaddle, u)
						mCol = azuriteCavity
					} else if harmonicSheet < 0.68 {
						// Sheet 2: 24k Burnished Gold Leaf & Warm Parchment crest
						u := (harmonicSheet - 0.34) / 0.34
						sCol = moebiusLerp3(goldLeaf, warmParchment, u*0.55)
						mCol = amberBronze
					} else {
						// Sheet 3: Cinnabar Vermilion & Quinacridone Rose-Coral lobe
						u := (harmonicSheet - 0.68) / 0.32
						sCol = moebiusLerp3(cinnabarRidge, roseCoral, u*0.65)
						mCol = amberBronze
					}
				case 31:
					// Outer Gilded Geodesic Astrolabe Ring (Alternating Gold & Azurite Enamel Ticks)
					sCol = goldLeaf
					if (partID & 3) == 0 {
						sCol = cinnabarRidge
					}
					mCol = amberBronze
				case 32:
					// Inner Tumbling Meridian Ring
					sCol = moebiusLerp3(goldLeaf, cobaltTeal, float64(partID&1)*0.55)
					mCol = azuriteCavity
				default:
					// Orbiting Topological Monopole Spheres (Cinnabar Ruby & Cobalt Turquoise)
					if (partID & 1) == 0 {
						sCol = cinnabarRidge
						mCol = roseCoral
					} else {
						sCol = cobaltTeal
						mCol = azuriteCavity
					}
				}

				nScreen := moebiusNorm3(hit.Normal.X, -hit.Normal.Y, hit.Normal.Z)
				return render.MoebiusPixel{
					Depth:           hit.T,
					ObjID:           objID,
					PartID:          partID,
					DetailID:        stone.BlockID,
					Normal:          nScreen,
					FlowTangent:     [2]float64{-nScreen[1], nScreen[0]},
					Illum:           lit,
					Spec:            common.Clamp01(0.60*math.Max(0.0, hit.Curvature) + burinChalk + common.Smoothstep(0.74, 0.98, lit)*0.65),
					Bounce:          0.46,
					SunColor:        sCol,
					MidColor:        mCol,
					ShadeColor:      umberShade,
					DeepShadowColor: deepIntaglio,
					CrackInk:        math.Max(stone.CrackInk*0.42, common.Smoothstep(0.20, 0.62, -hit.Curvature)*0.84),
					TickMask:        math.Max(stone.PockInk*0.38, burinInk*0.86),
				}
			}
		}

		// =====================================================================
		// 2. POLYCHROME RENAISSANCE CELESTIAL ASTROLABE, HARMONIC ROSE & CORNER ROSETTES
		// =====================================================================
		ang := math.Atan2(y, x)
		stone := render.EvaluateMoebiusStoneRelief(ne, x, y, 5.5, 5.5, 0.25)

		// 2a. Engraved Polychrome Astrolabe Bezel Ring (0.74 <= rScreen <= 0.88)
		if rScreen >= 0.74 && rScreen <= 0.88 {
			// 8 Beveled Jewel Cabochons at the 8 compass points around the bezel (r = 0.81)
			compassAng := math.Mod(ang+2.0*math.Pi+math.Pi/8.0, math.Pi/4.0) - math.Pi/8.0
			compassIdx := int(math.Floor((ang + math.Pi + math.Pi/8.0) / (math.Pi / 4.0)))
			dJewel := math.Hypot(rScreen-0.81, compassAng*0.81)
			if dJewel < 0.036 {
				uJ := 1.0 - dJewel/0.036
				jCol := cinnabarRidge
				if compassIdx&1 == 1 {
					jCol = cobaltTeal
				}
				return render.MoebiusPixel{
					Depth:           6.1,
					ObjID:           16,
					PartID:          160 + (compassIdx & 7),
					DetailID:        int(uJ * 8.0),
					Normal:          moebiusNorm3(-x*0.5, -y*0.5, 0.85),
					FlowTangent:     [2]float64{-math.Sin(ang), math.Cos(ang)},
					Illum:           0.65 + 0.35*uJ,
					Spec:            common.Smoothstep(0.5, 0.95, uJ) * 0.85,
					SunColor:        jCol,
					MidColor:        goldLeaf,
					ShadeColor:      umberShade,
					DeepShadowColor: deepIntaglio,
				}
			}

			degIdx := int(math.Floor((ang + math.Pi) / (2.0 * math.Pi) * 72.0))
			zodiacIdx := int(math.Floor((ang + math.Pi) / (2.0 * math.Pi) * 12.0))
			inOuterGoldRim := rScreen < 0.768 || rScreen > 0.852
			inMidBand := rScreen >= 0.795 && rScreen <= 0.825

			sCol := goldLeaf
			mCol := amberBronze
			if !inOuterGoldRim {
				if inMidBand {
					// Polychrome cloisonné zodiac segments in Azurite, Cinnabar, Malachite & Gold!
					switch zodiacIdx & 3 {
					case 0:
						sCol = azuriteCavity
						mCol = imperialViolet
					case 1:
						sCol = cinnabarRidge
						mCol = amberBronze
					case 2:
						sCol = malachiteSaddle
						mCol = azuriteCavity
					default:
						sCol = goldLeaf
						mCol = amberBronze
					}
				} else if degIdx&1 == 1 {
					sCol = warmParchment
				}
			}

			lit := common.Clamp01(0.74 - 0.24*(x*0.5+y*0.5))
			return render.MoebiusPixel{
				Depth:           6.5,
				ObjID:           15,
				PartID:          150 + (degIdx & 31) + int(moebiusBoolFloat(inOuterGoldRim))*64 + int(moebiusBoolFloat(inMidBand))*32,
				DetailID:        degIdx,
				Normal:          moebiusNorm3(-x*0.6, -y*0.6, 0.78),
				FlowTangent:     [2]float64{-math.Sin(ang), math.Cos(ang)},
				Illum:           lit,
				Spec:            0.58,
				SunColor:        sCol,
				MidColor:        mCol,
				ShadeColor:      umberShade,
				DeepShadowColor: deepIntaglio,
				CrackInk:        stone.CrackInk * 0.38,
			}
		}

		// 2b. Four Ornate Corner Planisphere Rosette Medallions (rScreen > 0.88)
		if rScreen > 0.88 {
			cx := 0.76
			if x < 0 {
				cx = -0.76
			}
			cyCorner := 0.76
			if y < 0 {
				cyCorner = -0.76
			}
			rCorner := math.Hypot(x-cx, y-cyCorner)
			angCorner := math.Atan2(y-cyCorner, x-cx)
			spokeCorner := int(math.Floor((angCorner + math.Pi) / (2.0 * math.Pi) * 16.0))

			if rCorner < 0.26 && rScreen > 0.91 {
				ringIdx := int(math.Floor(rCorner * 18.0))
				sCol := goldLeaf
				mCol := amberBronze
				lit := 0.76
				if ringIdx == 1 || ringIdx == 3 {
					sCol = cinnabarRidge
					if (spokeCorner & 1) == 1 {
						sCol = cobaltTeal
					}
				} else if (spokeCorner & 1) == 1 {
					sCol = warmParchment
				}
				return render.MoebiusPixel{
					Depth:           9.0,
					ObjID:           7,
					PartID:          70 + ringIdx*16 + (spokeCorner & 15),
					DetailID:        spokeCorner,
					Normal:          moebiusNorm3(-(x - cx), -(y - cyCorner), 0.82),
					FlowTangent:     [2]float64{-math.Sin(angCorner), math.Cos(angCorner)},
					Illum:           lit,
					Spec:            0.52,
					SunColor:        sCol,
					MidColor:        mCol,
					ShadeColor:      umberShade,
					DeepShadowColor: deepIntaglio,
					CrackInk:        stone.CrackInk * 0.35,
				}
			}

			// Deep Lapis-Ultramarine & Engraved Star-Grid Spandrel Field
			spandrelRing := int(math.Floor(rScreen * 16.0))
			return render.MoebiusPixel{
				Depth:           12.0,
				ObjID:           6,
				PartID:          60 + (spandrelRing & 7) + (spokeCorner & 3)*8,
				DetailID:        stone.BlockID,
				Normal:          [3]float64{-x * 0.3, -y * 0.3, 0.9},
				FlowTangent:     [2]float64{-math.Sin(ang), math.Cos(ang)},
				Illum:           0.46,
				SunColor:        azuriteCavity,
				MidColor:        imperialViolet,
				ShadeColor:      umberShade,
				DeepShadowColor: deepIntaglio,
				CrackInk:        stone.CrackInk * 0.45,
				Stipple:         0.38,
			}
		}

		// 2c. Inner Celestial Harmonic Rose Mandala & Conformal Astrolabe Chart (rScreen < 0.74)
		// Combines 8-fold & 12-fold polar harmonic petals with concentric degree rings!
		rose8 := 0.5 + 0.5*math.Cos(8.0*ang)
		rose4 := 0.5 + 0.5*math.Cos(4.0*ang+math.Pi*0.25)
		petalR := 0.46 + 0.16*rose8 + 0.08*rose4
		sec12 := int(math.Floor((ang + math.Pi) / (2.0 * math.Pi) * 24.0))
		ringBand := int(math.Floor(rScreen * 10.0))

		var sCol, mCol [3]float64
		var lit float64
		part := 50 + ringBand*4 + (sec12 & 3)

		if rScreen > petalR {
			// Outer Star-Chart Annulus inside the Bezel: Deep Lapis & Cobalt Turquoise with Golden Meridians!
			uOuter := common.Clamp01((rScreen - petalR) / math.Max(0.74-petalR, 0.05))
			sCol = moebiusLerp3(cobaltTeal, azuriteCavity, uOuter)
			if (sec12 & 1) == 0 {
				sCol = moebiusLerp3(sCol, goldLeaf, 0.25)
			}
			mCol = azuriteCavity
			lit = common.Clamp01(0.62 - 0.18*uOuter + 0.10*stone.ColorMod)
			part = 80 + (sec12 & 15) + ringBand*16
		} else {
			// Inner Harmonic Rose Mandala: Warm Parchment, Gold Leaf, Soft Coral & Celadon-Malachite Petals!
			petalMix := common.Smoothstep(0.28, 0.64, rScreen) * rose8
			if (ringBand & 1) == 0 {
				sCol = moebiusLerp3(warmParchment, goldLeaf, 0.45*rose4)
				mCol = amberBronze
			} else {
				sCol = moebiusLerp3(warmParchment, roseCoral, 0.38*petalMix)
				if (sec12 & 2) == 0 {
					sCol = moebiusLerp3(warmParchment, malachiteSaddle, 0.35*petalMix)
				}
				mCol = amberBronze
			}
			lit = common.Clamp01(0.78 - 0.20*rScreen + 0.10*(1.0-uy))
		}

		return render.MoebiusPixel{
			Depth:           11.0,
			ObjID:           5,
			PartID:          part,
			DetailID:        stone.BlockID,
			Normal:          [3]float64{-x * 0.28, -y * 0.28, 0.92},
			FlowTangent:     [2]float64{-math.Sin(ang), math.Cos(ang)},
			Illum:           lit,
			Spec:            common.Smoothstep(0.76, 0.96, lit) * 0.42,
			SunColor:        sCol,
			MidColor:        mCol,
			ShadeColor:      umberShade,
			DeepShadowColor: deepIntaglio,
			CrackInk:        stone.CrackInk * 0.40,
			TickMask:        stone.PockInk * 0.34,
			Stipple:         (1.0 - lit) * 0.32,
		}
	})

	c.MergeRenderCanvas(rc)
}
