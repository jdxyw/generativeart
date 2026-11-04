package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type moebiusHermeticGarage struct {
	ribs     int
	spires   int
	ringTilt float64
}

// NewMoebiusHermeticGarage returns a MoebiusHermeticGarage generator inspired by
// Jean "Mœbius" Giraud's Le Garage Hermétique (1976) and Sur l'Étoile: inside a
// colossal weathered asteroid cavern, a 3D SDF raymarched bone-ivory and
// vermilion-striped toroidal mothership dock with ribbed mechanical struts and
// counter-rotating reactor core hovers above an emerald-turquoise subterranean
// drydock, flanked by crystalline mineral spires, an observation catwalk, and
// orbiting shuttle pods.
func NewMoebiusHermeticGarage(ribs, spires int, ringTilt float64) *moebiusHermeticGarage {
	if ribs <= 0 {
		ribs = 24
	}
	if spires <= 0 {
		spires = 6
	}
	if ringTilt == 0 {
		ringTilt = 0.36
	}
	return &moebiusHermeticGarage{
		ribs:     ribs,
		spires:   spires,
		ringTilt: ringTilt,
	}
}

// Generative renders the static MoebiusHermeticGarage composition.
func (hg *moebiusHermeticGarage) Generative(c *generativeart.Canva) {
	hg.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360° looping MoebiusHermeticGarage animation.
func (hg *moebiusHermeticGarage) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	seed := c.Seed()
	if seed == 0 {
		seed = 20262303
	}
	ne := common.NewNoiseEngine(seed)
	phi := 2.0 * math.Pi * phase

	// Kubelka-Munk Mœbius Hermetic Garage Gouache & Mineral Palette
	hullIvory := render.MixKubelkaMunk(render.PW(render.PigmentTitaniumWhite, 0.80), render.PW(render.PigmentYellowOchre, 0.20))
	hullVermilion := render.MixKubelkaMunk(render.PW(render.PigmentVermilion, 0.76), render.PW(render.PigmentCadmiumYellow, 0.24))
	hullGold := render.MixKubelkaMunk(render.PW(render.PigmentGoldLeaf, 0.68), render.PW(render.PigmentCadmiumYellow, 0.32))
	hullShade := render.MixKubelkaMunk(render.PW(render.PigmentUltramarine, 0.54), render.PW(render.PigmentTitaniumWhite, 0.46))

	cavernSun := render.MixKubelkaMunk(render.PW(render.PigmentYellowOchre, 0.56), render.PW(render.PigmentTitaniumWhite, 0.44))
	cavernSienna := render.MixKubelkaMunk(render.PW(render.PigmentBurntSienna, 0.65), render.PW(render.PigmentYellowOchre, 0.35))
	cavernUmber := render.MixKubelkaMunk(render.PW(render.PigmentRawUmber, 0.58), render.PW(render.PigmentUltramarine, 0.42))
	deepIndigo := render.MixKubelkaMunk(render.PW(render.PigmentUltramarine, 0.52), render.PW(render.PigmentPineSootInk, 0.48))

	crystalCyan := render.MixKubelkaMunk(render.PW(render.PigmentCobaltTeal, 0.68), render.PW(render.PigmentTitaniumWhite, 0.32))
	crystalJade := render.MixKubelkaMunk(render.PW(render.PigmentMalachite, 0.65), render.PW(render.PigmentAzurite, 0.35))

	lightDir3D := render.V3(-0.56, 0.58, 0.59).Normalize()

	// Exact symmetry-matched rotation speeds for seamless 20-frame loop:
	//  - 24-ribbed Toroidal Mothership (6-fold spoke symmetry): rotates by +(2*pi/6)*phase
	//  - Central 4-fold Octahedral Reactor Core: counter-rotates by -(2*pi/4)*phase
	//  - 3 Orbiting Shuttle Pods: advance by +(2*pi/3)*phase (1 pod spacing per cycle!)
	rotMothership := 0.22 + (2.0*math.Pi/6.0)*phase
	rotCore := -0.35 - (2.0*math.Pi/4.0)*phase
	rotPods := 0.15 + (2.0*math.Pi/3.0)*phase

	shipCenter := render.V3(0.04, 0.12, 0.0)
	tiltX := 0.92 // View toroidal mothership at an expressive Mœbius perspective angle
	tiltZ := -0.22

	evalGarage3D := func(p render.Vec3) (float64, int, int, render.Vec3) {
		qShip := p.Sub(shipCenter).RotateZ(tiltZ).RotateX(tiltX).RotateY(rotMothership)

		// 1. Ribbed Toroidal Mothership Ring (24 radial carapace ribs + 6 docking spokes)
		fRib, ribIdx := render.FoldPolarModXZ(qShip, 24)
		ribFlute := 0.014 * math.Cos(float64(ribIdx)*math.Pi)
		dMainTorus := render.SDF3DTorus(qShip, 0.44, 0.092+ribFlute)
		// Armored rib collar rings along the torus
		dRibCollar := render.SDF3DBox(fRib.Sub(render.V3(0.44, 0, 0)), render.V3(0.105, 0.105, 0.018)) - 0.008
		dOuterRing := render.SDF3DSmoothUnion(dMainTorus, dRibCollar, 0.015)

		// 6 Radial Docking Spokes connecting the outer torus to the central hub
		fSpoke, spokeIdx := render.FoldPolarModXZ(qShip, 6)
		dSpoke := render.SDF3DCapsule(fSpoke, render.V3(0.14, 0, 0), render.V3(0.36, 0, 0), 0.024)
		dHubRing := render.SDF3DTorus(qShip, 0.16, 0.032)

		// 2. Counter-Rotating Central Octahedral Crystal Reactor Core
		qCore := p.Sub(shipCenter).RotateZ(tiltZ).RotateX(tiltX).RotateY(rotCore)
		fOct := render.FoldOctahedral(qCore)
		dCore := render.SDF3DOctahedron(fOct, 0.135)

		// 3. Three Orbiting Reconnaissance Shuttle Pods on an inclined outer track
		qPodTrack := p.Sub(shipCenter).RotateZ(0.28).RotateX(1.05).RotateY(rotPods)
		fPod, podIdx := render.FoldPolarModXZ(qPodTrack, 3)
		pPod := fPod.Sub(render.V3(0.62, 0.0, 0.0))
		dPodBody := render.SDF3DCapsule(pPod, render.V3(0, 0, -0.045), render.V3(0, 0, 0.045), 0.026)
		dPodWing := render.SDF3DRoundedBox(pPod, render.V3(0.068, 0.006, 0.028), 0.004)
		dPods := render.SDF3DSmoothUnion(dPodBody, dPodWing, 0.008)

		minD := dOuterRing
		objID := 30
		partID := 300 + (ribIdx & 31)
		loc := fRib

		if dSpoke < minD {
			minD = dSpoke
			objID = 31
			partID = 340 + spokeIdx
			loc = fSpoke
		}
		if dHubRing < minD {
			minD = dHubRing
			objID = 31
			partID = 350 + (ribIdx/2)&15
			loc = qShip
		}
		if dCore < minD {
			minD = dCore
			objID = 32
			partID = 360 + int(math.Floor(fOct.X*12.0))
			loc = fOct
		}
		if dPods < minD {
			minD = dPods
			objID = 33
			partID = 370 + podIdx
			loc = pPod
		}
		return minD, objID, partID, loc
	}

	garageSDF := func(p render.Vec3) float64 {
		d, _, _, _ := evalGarage3D(p)
		return d
	}

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)
	cfg := render.DefaultMoebiusConfig(seed)
	cfg.WarmBounceColor = color.RGBA{R: 0x38, G: 0xd8, B: 0xc4, A: 255}
	cfg.HatchSpacingPx = 3.6

	render.RenderMoebiusScene(rc, cfg, func(ux, uy float64) render.MoebiusPixel {
		x := ux*2.0 - 1.0
		y := uy*2.0 - 1.0

		// =====================================================================
		// 1. FOREGROUND SUNLIT STRATIFIED SANDSTONE ARCH, GIRDERS & CATWALK
		// =====================================================================
		// Asymmetric Mœbius canyon/cavern arch framing top-left & left cliff wall
		cavernNoise := 0.065*ne.Simplex2D(x*2.4, y*2.4) + 0.032*ne.Simplex2D(x*5.2, y*5.2)
		leftWallBoundary := -0.68 + 0.10*y + cavernNoise
		roofArchBoundary := -0.64 + 0.24*(x*x) + cavernNoise
		inLeftCliff := x < leftWallBoundary
		inRoofArch := y < roofArchBoundary

		// Stepped mechanical observation balcony & railing in lower-right foreground
		catwalkTopY := 0.56 + 0.14*(x-0.18)
		inCatwalk := x > 0.14 && y > catwalkTopY
		inCatwalkRail := x > 0.16 && y >= catwalkTopY-0.065 && y <= catwalkTopY &&
			(math.Abs(y-(catwalkTopY-0.060)) < 0.008 || int(math.Floor((x+1.0)*26.0))%2 == 0)

		if inLeftCliff || inRoofArch || inCatwalk || inCatwalkRail {
			stone := render.EvaluateMoebiusStoneRelief(ne, x, y, 7.5, 8.5, 0.36)
			if inCatwalkRail {
				return render.MoebiusPixel{
					Depth:           1.15,
					ObjID:           12,
					PartID:          129,
					DetailID:        int((x + 1.0) * 32.0),
					Normal:          [3]float64{-0.2, -0.7, 0.68},
					FlowTangent:     [2]float64{0.98, 0.14},
					Illum:           0.84,
					SunColor:        hullGold,
					MidColor:        cavernSienna,
					ShadeColor:      cavernUmber,
					DeepShadowColor: deepIndigo,
				}
			}
			if inCatwalk && !inLeftCliff {
				panelIdx := int(math.Floor((x - 0.14) * 14.0))
				uCat := (y - catwalkTopY) / 0.44
				lit := 0.80 - 0.36*common.Clamp01(uCat)
				return render.MoebiusPixel{
					Depth:           1.20,
					ObjID:           12,
					PartID:          120 + (panelIdx & 15),
					DetailID:        stone.BlockID,
					Normal:          moebiusNorm3(-0.25, -0.65+0.4*uCat, 0.72),
					FlowTangent:     [2]float64{0.98, 0.14},
					Illum:           lit,
					Spec:            common.Smoothstep(0.25, 0.0, uCat) * 0.55,
					Bounce:          0.48,
					SunColor:        hullIvory,
					MidColor:        cavernSienna,
					ShadeColor:      cavernUmber,
					DeepShadowColor: deepIndigo,
					CrackInk:        stone.CrackInk * 0.65,
					TickMask:        stone.PockInk * 0.50,
				}
			}

			// True distance to the inner opening of the sandstone canyon arch (zero internal seam!):
			archOpenDist := math.Min(x-leftWallBoundary, y-roofArchBoundary)
			rimDist := math.Max(0.0, -archOpenDist)

			strataCoord := (y+1.0)*5.8 + 0.25*x + 0.38*stone.ColorMod
			strataIdx := int(math.Floor(strataCoord))
			strataFrac := strataCoord - float64(strataIdx)
			isStrataLedge := strataFrac < 0.22

			var nx, ny float64
			if (y - roofArchBoundary) < (x - leftWallBoundary) {
				nx = -0.30*x + 0.22*stone.NormalMod[0]
				ny = -0.55 + 0.22*stone.NormalMod[1]
			} else {
				nx = -0.52 + 0.24*stone.NormalMod[0]
				ny = -0.38 + 0.24*stone.NormalMod[1]
			}
			n := moebiusNorm3(nx, ny, 0.72)

			// Sunlit sandstone & terracotta chiaroscuro: bright warm ochre along the inner arch rim,
			// transitioning smoothly to rich burnt-sienna and umber toward the outer rock mass!
			rimGlow := common.Smoothstep(0.32, 0.0, rimDist) * 0.44
			ledgeBoost := moebiusBoolFloat(isStrataLedge) * 0.16
			lit := common.Clamp01(0.34 + rimGlow + ledgeBoost + 0.14*stone.ColorMod)

			sCol := cavernSun
			mCol := cavernSienna
			part := 100 + (strataIdx & 15)

			if (strataIdx % 3) == 0 {
				sCol = moebiusLerp3(cavernSun, hullGold, 0.45)
				mCol = cavernSienna
			} else if (strataIdx % 3) == 1 {
				sCol = moebiusLerp3(cavernSun, hullVermilion, 0.36)
				mCol = cavernSienna
			}

			return render.MoebiusPixel{
				Depth:           1.35,
				ObjID:           10,
				PartID:          part,
				DetailID:        stone.BlockID,
				Normal:          n,
				FlowTangent:     [2]float64{0.92, -0.38},
				Illum:           lit,
				Spec:            common.Smoothstep(0.76, 0.96, lit) * 0.48,
				Bounce:          0.48,
				SunColor:        sCol,
				MidColor:        mCol,
				ShadeColor:      cavernUmber,
				DeepShadowColor: deepIndigo,
				CrackInk:        stone.CrackInk * 0.78,
				TickMask:        stone.PockInk * 0.55,
			}
		}

		// =====================================================================
		// 2. RAYMARCH 3D TOROIDAL MOTHERSHIP, REACTOR CORE & ORBITING PODS
		// =====================================================================
		if x >= -0.72 && x <= 0.82 && y >= -0.68 && y <= 0.48 {
			ro := render.V3(x*0.90, -y*0.90, 2.2)
			rd := render.V3(0, 0, -1)
			hit := render.RaymarchSDF3D(garageSDF, ro, rd, lightDir3D, 3.4, 60)
			if hit.Hit {
				_, objID, partID, pLoc := evalGarage3D(hit.Pos)
				nLoc := hit.Normal.RotateZ(tiltZ).RotateX(tiltX).RotateY(rotMothership)
				stone := render.EvaluateMoebiusStoneRelief3D(ne, pLoc.Scale(2.8), nLoc, 7.0, 0.24)
				// Include upward turquoise-and-gold lagoon bounce light so the lower ribs glow warmly!
				upBounce := math.Max(0.0, -hit.Normal.Y) * 0.46
				lit := common.Clamp01((0.24 + 0.72*math.Max(0.0, hit.Normal.Dot(lightDir3D)) + upBounce) *
					(0.56 + 0.44*hit.Shadow) * (0.65 + 0.35*hit.AO))
				burinInk, burinChalk := render.Evaluate3DBurinEngraving(pLoc, nLoc, lit, 78.0)

				sCol := hullIvory
				mCol := hullShade
				switch objID {
				case 30:
					// Toroidal Mothership Carapace: Bone-Ivory with Mœbius Vermilion & Gold Signal Rib Bands!
					if (partID % 6) == 0 {
						sCol = hullVermilion
						mCol = cavernSienna
					} else if (partID % 6) == 3 {
						sCol = hullGold
						mCol = cavernSienna
					}
				case 31:
					// Brass Docking Spokes & Inner Hub Ring
					sCol = hullGold
					mCol = cavernSienna
				case 32:
					// Bioluminescent Turquoise & Gold Octahedral Reactor Core
					sCol = crystalCyan
					mCol = crystalJade
					lit = math.Max(lit, 0.78)
				case 33:
					// Orbiting Reconnaissance Shuttle Pods
					sCol = hullIvory
					if pLoc.X > 0.025 {
						sCol = hullVermilion
					}
					mCol = hullGold
				}

				nScreen := moebiusNorm3(hit.Normal.X, -hit.Normal.Y, hit.Normal.Z)
				return render.MoebiusPixel{
					Depth:           2.0 + hit.T*0.45,
					ObjID:           objID,
					PartID:          partID,
					DetailID:        stone.BlockID,
					Normal:          nScreen,
					FlowTangent:     [2]float64{-nScreen[1], nScreen[0]},
					Illum:           lit,
					Spec:            common.Clamp01(0.55*math.Max(0.0, hit.Curvature) + burinChalk + common.Smoothstep(0.75, 0.98, lit)*0.60),
					Bounce:          0.52,
					SunColor:        sCol,
					MidColor:        mCol,
					ShadeColor:      cavernUmber,
					DeepShadowColor: deepIndigo,
					CrackInk:        math.Max(stone.CrackInk*0.42, common.Smoothstep(0.25, 0.65, -hit.Curvature)*0.75),
					TickMask:        math.Max(stone.PockInk*0.38, burinInk*0.80),
				}
			}
		}

		// =====================================================================
		// 3. SUBTERRANEAN CRYSTAL SPIRES & TURQUOISE DRYDOCK LAGOON (y > 0.20)
		// =====================================================================
		if y > 0.16 {
			// Faceted hexagonal crystal spires rising from the subterranean lagoon
			spireFreq := 7.5
			spireCell := int(math.Floor((x + 1.2) * spireFreq))
			spireFrac := (x+1.2)*spireFreq - float64(spireCell) - 0.5
			spireH := 0.22 + 0.14*math.Sin(float64(spireCell)*2.1+0.6)
			spireTopY := 0.64 - spireH + math.Abs(spireFrac)*0.22

			if y >= spireTopY && y < 0.74 {
				facet := 0
				fnx := -0.68
				if spireFrac > -0.12 && spireFrac < 0.18 {
					facet = 1
					fnx = 0.10
				} else if spireFrac >= 0.18 {
					facet = 2
					fnx = 0.72
				}
				n := moebiusNorm3(fnx, -0.35, 0.68)
				lit := common.Clamp01(0.22 + 0.78*math.Max(0.0, n[0]*(-0.55)+n[1]*(-0.55)+n[2]*0.62))
				return render.MoebiusPixel{
					Depth:           5.2,
					ObjID:           22,
					PartID:          220 + spireCell*3 + facet,
					DetailID:        spireCell*4 + facet,
					Normal:          n,
					FlowTangent:     [2]float64{0, 1},
					Illum:           lit,
					Spec:            common.Smoothstep(0.68, 0.94, lit) * 0.78,
					Bounce:          0.55,
					SunColor:        crystalCyan,
					MidColor:        crystalJade,
					ShadeColor:      cavernUmber,
					DeepShadowColor: deepIndigo,
				}
			}

			if y >= 0.68 {
				// Shimmering Turquoise Subterranean Drydock Lagoon with concentric ripples
				ty := (y - 0.68) / 0.32
				ripple := 0.5 + 0.5*math.Cos((y-0.68)*72.0-phi+0.4*math.Sin(x*14.0))
				lit := common.Clamp01(0.58 + 0.34*ripple*(1.0-ty*0.4))
				return render.MoebiusPixel{
					Depth:           6.8 - 1.5*ty,
					ObjID:           24,
					PartID:          240 + int(math.Floor(ty*6.0)),
					DetailID:        int(ripple * 8.0),
					Normal:          [3]float64{0, -0.85, 0.52},
					FlowTangent:     [2]float64{1, 0},
					Illum:           lit,
					Spec:            common.Smoothstep(0.78, 0.96, ripple) * 0.72,
					SunColor:        moebiusLerp3(crystalCyan, hullGold, common.Smoothstep(0.75, 0.98, ripple)*0.55),
					MidColor:        crystalJade,
					ShadeColor:      cavernUmber,
					DeepShadowColor: deepIndigo,
				}
			}
		}

		// =====================================================================
		// 4. DISTANT HERMETIC CAVERN DOMES & WATERCOLOR CLOUD BANK
		// =====================================================================
		cloudH := -0.06 + 0.08*math.Sin(x*3.6+0.5) + 0.04*ne.Simplex2D(x*4.2, y*4.2)
		if y > cloudH {
			stone := render.EvaluateMoebiusStoneRelief(ne, x*0.8, y*0.8, 5.0, 5.0, 0.28)
			uCloud := common.Clamp01((y - cloudH) / 0.35)
			lit := common.Clamp01(0.82 - 0.38*uCloud + 0.12*stone.ColorMod)
			return render.MoebiusPixel{
				Depth:           14.0,
				ObjID:           6,
				PartID:          60 + int(math.Floor(uCloud*4.0)),
				DetailID:        stone.BlockID,
				Normal:          [3]float64{0, -0.6, 0.8},
				FlowTangent:     [2]float64{1, 0},
				Illum:           lit,
				Atmos:           0.35,
				SunColor:        cavernSun,
				MidColor:        cavernSienna,
				ShadeColor:      cavernUmber,
				DeepShadowColor: deepIndigo,
				AtmosColor:      crystalCyan,
				CrackInk:        stone.CrackInk * 0.35,
				Stipple:         uCloud * 0.38,
			}
		}

		// Deep Prussian-Indigo Hermetic Vault Sky with Celestial Grid Arcs
		skyT := common.Clamp01((y + 1.0) / 0.95)
		return render.MoebiusPixel{
			Depth:           30.0,
			ObjID:           1,
			PartID:          10,
			DetailID:        10,
			Normal:          [3]float64{0, 0, 1},
			FlowTangent:     [2]float64{1, 0},
			Illum:           0.22 + 0.56*skyT,
			SunColor:        cavernSun,
			MidColor:        hullShade,
			ShadeColor:      cavernUmber,
			DeepShadowColor: deepIndigo,
			Stipple:         (1.0 - skyT) * 0.46,
		}
	})

	c.MergeRenderCanvas(rc)
}
