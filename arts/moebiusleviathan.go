package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type moebiusLeviathan struct {
	ribs         int
	crystalCount int
	bankAngle    float64
}

// NewMoebiusLeviathan returns a MoebiusLeviathan generator inspired by Jean
// "Mœbius" Giraud's Starwatcher, Les Maîtres du Temps, and Dune: a 3D SDF
// raymarched biomechanical Nautilus-whale airship with fluted bone-ivory carapace
// segments, vermilion-gold solar dorsal fins, and escort gliders soaring above
// Kubelka-Munk watercolor cloud banks and craggy sandstone spires.
func NewMoebiusLeviathan(ribs, crystalCount int, bankAngle float64) *moebiusLeviathan {
	if ribs <= 0 {
		ribs = 14
	}
	if crystalCount <= 0 {
		crystalCount = 5
	}
	if bankAngle == 0 {
		bankAngle = -0.18
	}
	return &moebiusLeviathan{
		ribs:         ribs,
		crystalCount: crystalCount,
		bankAngle:    bankAngle,
	}
}

// Generative renders the static MoebiusLeviathan print.
func (ml *moebiusLeviathan) Generative(c *generativeart.Canva) {
	ml.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless looping MoebiusLeviathan animation.
func (ml *moebiusLeviathan) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	seed := c.Seed()
	if seed == 0 {
		seed = 20262202
	}
	ne := common.NewNoiseEngine(seed)
	phi := 2.0 * math.Pi * phase

	// Kubelka-Munk Physical Mineral & Watercolor Palette
	skyTop := render.MixKubelkaMunk(render.PW(render.PigmentUltramarine, 0.65), render.PW(render.PigmentCerulean, 0.35))
	skyMid := render.MixKubelkaMunk(render.PW(render.PigmentCerulean, 0.48), render.PW(render.PigmentTitaniumWhite, 0.52))
	skyLow := render.MixKubelkaMunk(render.PW(render.PigmentCadmiumYellow, 0.28), render.PW(render.PigmentTitaniumWhite, 0.72))

	carapaceIvory := render.MixKubelkaMunk(render.PW(render.PigmentTitaniumWhite, 0.85), render.PW(render.PigmentYellowOchre, 0.15))
	carapaceOchre := render.MixKubelkaMunk(render.PW(render.PigmentYellowOchre, 0.72), render.PW(render.PigmentTitaniumWhite, 0.28))
	carapaceTeal := render.MixKubelkaMunk(render.PW(render.PigmentCobaltTeal, 0.70), render.PW(render.PigmentTitaniumWhite, 0.30))
	finVermilion := render.MixKubelkaMunk(render.PW(render.PigmentVermilion, 0.75), render.PW(render.PigmentCadmiumYellow, 0.25))
	finGold := render.MixKubelkaMunk(render.PW(render.PigmentGoldLeaf, 0.80), render.PW(render.PigmentCadmiumYellow, 0.20))
	hullMid := render.MixKubelkaMunk(render.PW(render.PigmentBurntSienna, 0.45), render.PW(render.PigmentImperialPurple, 0.55))
	hullShade := render.MixKubelkaMunk(render.PW(render.PigmentUltramarine, 0.62), render.PW(render.PigmentRawUmber, 0.38))
	hullDeep := render.MixKubelkaMunk(render.PW(render.PigmentPineSootInk, 0.75), render.PW(render.PigmentUltramarine, 0.25))

	cloudSun := render.MixKubelkaMunk(render.PW(render.PigmentTitaniumWhite, 0.88), render.PW(render.PigmentCadmiumYellow, 0.12))
	cloudPeach := render.MixKubelkaMunk(render.PW(render.PigmentQuinacridoneRose, 0.35), render.PW(render.PigmentYellowOchre, 0.45), render.PW(render.PigmentTitaniumWhite, 0.20))
	cloudViolet := render.MixKubelkaMunk(render.PW(render.PigmentImperialPurple, 0.52), render.PW(render.PigmentUltramarine, 0.48))

	rockSun := render.MixKubelkaMunk(render.PW(render.PigmentYellowOchre, 0.65), render.PW(render.PigmentVermilion, 0.35))
	rockMid := render.MixKubelkaMunk(render.PW(render.PigmentBurntSienna, 0.75), render.PW(render.PigmentRawUmber, 0.25))

	lightDir3D := render.V3(-0.58, 0.52, 0.62).Normalize() // Y is up in 3D SDF space!
	lightDirScreen := moebiusNorm3(-0.58, -0.52, 0.62)

	// Define the 3D SDF for the Biomechanical Nautilus-Whale Leviathan Airship!
	// Centered at (0, 0.14, 0) in 3D world coordinates (X right, Y up, Z toward camera)
	shipBobY := 0.035 * math.Sin(phi)
	shipPitch := ml.bankAngle + 0.04*math.Cos(phi)
	shipYaw := -0.32 + 0.05*math.Sin(phi)

	transformToShipLocal := func(p render.Vec3) render.Vec3 {
		q := p.Sub(render.V3(-0.04, 0.16+shipBobY, 0.0))
		q = q.RotateZ(-shipPitch).RotateY(-shipYaw)
		// Gentle swimming S-undulation along the hull spine X in [-0.75, 0.75]
		swimWave := 0.045 * math.Sin(q.X*3.2-phi)
		q.Y -= swimWave
		q.Z -= 0.025 * math.Cos(q.X*2.8-phi)
		return q
	}

	leviathanSDF := func(p render.Vec3) float64 {
		q := transformToShipLocal(p)
		// 1. Main Spindle Hull along X in [-0.68, 0.68] (thick armored prow at X < 0, tapering tail at X > 0)
		dMain := render.SDF3DCapsule(q, render.V3(-0.52, 0.0, 0.0), render.V3(0.62, 0.02, 0.0), 0.21*(1.0-0.38*q.X))
		// Armored Cephalopod Prow Hood at X in [-0.66, -0.18]
		dHood := render.SDF3DSphere(render.V3((q.X+0.34)*1.15, (q.Y-0.03)*1.25, q.Z*1.35), 0.26)
		dHull := render.SDF3DSmoothUnion(dMain, dHood, 0.08)

		// Fluted biomechanical carapace ribs carved along X
		ribWave := math.Sin(q.X * float64(ml.ribs) * 3.2)
		dHull -= 0.011 * ribWave * common.Smoothstep(-0.55, 0.45, q.X)

		// 2. Ventral Brass Gondola Keel beneath the hull (Y in [-0.30, -0.12])
		dGondola := render.SDF3DRoundedBox(q.Sub(render.V3(-0.12, -0.21, 0.0)), render.V3(0.26, 0.055, 0.065), 0.025)
		dHull = render.SDF3DSmoothUnion(dHull, dGondola, 0.04)

		// 3. Swept Dorsal Sail Fin (Y > 0.14) & Twin Lateral Diving Fins
		finX := q.X - 0.08 - 0.55*math.Max(0.0, q.Y-0.12)
		dDorsal := render.SDF3DBox(render.V3(finX, q.Y-0.28, q.Z), render.V3(0.18*(1.0-(q.Y-0.12)*1.2), 0.16, 0.014))

		latAbsZ := math.Abs(q.Z)
		pecX := q.X + 0.08 - 0.48*latAbsZ
		dPectoral := render.SDF3DBox(render.V3(pecX, q.Y+0.05+0.18*latAbsZ, latAbsZ-0.28), render.V3(0.16*(1.0-latAbsZ*0.9), 0.014, 0.22))

		return render.SDF3DSmoothUnion(dHull, math.Min(dDorsal, dPectoral), 0.025)
	}

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)
	cfg := render.DefaultMoebiusConfig(seed)
	cfg.WarmBounceColor = color.RGBA{R: 0xf4, G: 0x92, B: 0x48, A: 255}

	render.RenderMoebiusScene(rc, cfg, func(ux, uy float64) render.MoebiusPixel {
		x := ux*2.0 - 1.0
		y := uy*2.0 - 1.0
		crag := ne.FBM2D(x*4.2+1.1, y*4.2-0.8, 4, 2.0, 0.5)

		// =====================================================================
		// 1. FOREGROUND LOWER-LEFT & LOWER-RIGHT CRAGGY SANDSTONE CANYON SPIRES
		// =====================================================================
		fbmMtn, dMtnDx, _ := ne.FBM2DDeriv(x*3.8+0.7, y*3.8, 4, 2.0, 0.5)
		spireY := 0.72 + 0.12*math.Cos(x*3.0) + 0.03*fbmMtn
		dSpireDx := -0.36*math.Sin(x*3.0) + 0.03*dMtnDx

		// Left towering sandstone hoodoo spire (x ~ -0.72) and right canyon buttress (x ~ 0.68)
		dL := (x - (-0.72)) / 0.22
		if math.Abs(dL) < 1.5 {
			eL := math.Exp(-dL * dL * 1.9)
			spireY -= 0.54 * eL
			dSpireDx -= 0.54 * (-3.8 * dL / 0.22) * eL
		}
		dR := (x - 0.70) / 0.25
		if math.Abs(dR) < 1.5 {
			eR := math.Exp(-dR * dR * 1.8)
			spireY -= 0.42 * eR
			dSpireDx -= 0.42 * (-3.6 * dR / 0.25) * eR
		}

		if y >= spireY {
			dyBelow := y - spireY
			stone := render.EvaluateMoebiusStoneRelief(ne, x, y, 7.8, 8.8, 0.42)
			nx := common.Constrain(0.48*dSpireDx+0.26*stone.NormalMod[0], -0.92, 0.92)
			ny := -0.25 + 0.24*stone.NormalMod[1]
			if dyBelow < 0.030 {
				ny = -0.78
			}
			nz := math.Sqrt(math.Max(0.12, 1.0-common.Clamp01(nx*nx+ny*ny)*0.85))
			n := moebiusNorm3(nx, ny, nz)
			lit := common.Clamp01(0.14 + 0.88*math.Max(0.0, moebiusDot3(n, lightDirScreen)))

			return render.MoebiusPixel{
				Depth:           1.6 + 0.2*dyBelow,
				ObjID:           10,
				PartID:          100 + int(math.Floor(dyBelow*4.0+0.8*stone.ColorMod))&15,
				DetailID:        stone.BlockID,
				Normal:          n,
				FlowTangent:     [2]float64{0.96, 0.24 * nx},
				Illum:           lit,
				Spec:            common.Smoothstep(0.04, 0.0, dyBelow) * 0.60,
				Bounce:          0.42,
				SunColor:        rockSun,
				MidColor:        rockMid,
				ShadeColor:      hullShade,
				DeepShadowColor: hullDeep,
				CrackInk:        stone.CrackInk,
				TickMask:        stone.PockInk,
			}
		}

		// =====================================================================
		// 2. 3D SDF RAYMARCHED BIOMECHANICAL NAUTILUS-WHALE LEVIATHAN AIRSHIP
		// =====================================================================
		if x >= -0.92 && x <= 0.92 && y >= -0.68 && y <= 0.36 {
			ro := render.V3(x*0.96, -y*0.96, 2.2)
			rd := render.V3(0.0, 0.0, -1.0)
			hit := render.RaymarchSDF3D(leviathanSDF, ro, rd, lightDir3D, 3.2, 56)
			if hit.Hit {
				qLoc := transformToShipLocal(hit.Pos)
				stone := render.EvaluateMoebiusStoneRelief(ne, qLoc.X*1.8, qLoc.Y*1.8, 7.0, 7.0, 0.25)

				nScreen := moebiusNorm3(
					hit.Normal.X+0.10*stone.NormalMod[0],
					-hit.Normal.Y+0.10*stone.NormalMod[1],
					hit.Normal.Z,
				)
				lit := common.Clamp01((0.16 + 0.84*math.Max(0.0, hit.Normal.Dot(lightDir3D))) *
					(0.45 + 0.55*hit.Shadow) * hit.AO)

				// Exact hull radius at qLoc.X so only the protruding dorsal & lateral fins get striped!
				hullR := 0.22 * (1.0 - 0.35*qLoc.X)
				if qLoc.X < -0.15 {
					hullR = 0.25
				}
				ribIdx := int(math.Floor((qLoc.X + 1.0) * float64(ml.ribs) * 1.4))
				isDorsalFin := qLoc.Y > hullR+0.015
				isPectoralFin := math.Abs(qLoc.Z) > hullR+0.025 && qLoc.Y < 0.04
				isGondola := qLoc.Y < -0.16 && math.Abs(qLoc.Z) <= 0.11
				isProwEye := qLoc.X < -0.34 && math.Abs(qLoc.Y-0.02) < 0.052

				sCol := carapaceIvory
				part := 300 + (ribIdx & 31)
				if isProwEye {
					sCol = carapaceTeal
					lit = 0.94
					part = 390
				} else if isDorsalFin || isPectoralFin {
					stripe := int(math.Floor((qLoc.X + qLoc.Y) * 18.0))
					if stripe&1 == 0 {
						sCol = finGold
					} else {
						sCol = finVermilion
					}
					part = 350 + (stripe & 15)
				} else if isGondola {
					sCol = carapaceOchre
					part = 380 + (ribIdx & 7)
				} else if math.Abs(qLoc.Y) < 0.035 {
					// Lateral brass armor belt along the hull equator
					sCol = carapaceOchre
					part = 340 + (ribIdx & 15)
				}

				return render.MoebiusPixel{
					Depth:           hit.T,
					ObjID:           30,
					PartID:          part,
					DetailID:        stone.BlockID,
					Normal:          nScreen,
					FlowTangent:     [2]float64{math.Cos(shipPitch), math.Sin(shipPitch)},
					Illum:           lit,
					Spec:            common.Clamp01(0.45*math.Max(0.0, hit.Curvature) + common.Smoothstep(0.78, 0.98, lit)*0.65),
					Bounce:          0.45,
					SunColor:        sCol,
					MidColor:        hullMid,
					ShadeColor:      hullShade,
					DeepShadowColor: hullDeep,
					CrackInk:        math.Max(stone.CrackInk*0.48, common.Smoothstep(0.28, 0.68, -hit.Curvature)*0.72),
					TickMask:        stone.PockInk * 0.60,
				}
			}
		}

		// =====================================================================
		// 3. ESCORT SOLAR GLIDERS FLANKING THE LEVIATHAN
		// =====================================================================
		escorts := [2][3]float64{
			{-0.52 + 0.08*math.Cos(phi), -0.34 + 0.04*math.Sin(phi), 0.14},
			{0.48 - 0.08*math.Sin(phi), 0.06 + 0.04*math.Cos(phi), 0.16},
		}
		for eIdx, esc := range escorts {
			ex := (x - esc[0]) / esc[2]
			ey := (y - esc[1]) / (esc[2] * 0.48)
			absEx := math.Abs(ex)
			flap := 0.22 * math.Sin(2.0*phi+float64(eIdx)*2.1) * absEx * absEx
			wingY := -0.16*math.Sin(absEx*math.Pi) + 0.20*absEx*absEx - flap
			if absEx <= 1.0 && math.Abs(ey-wingY) <= 0.14*(1.0-0.78*absEx) {
				return render.MoebiusPixel{
					Depth:           2.8,
					ObjID:           40 + eIdx,
					PartID:          (40+eIdx)*10 + int(moebiusBoolFloat(ex > 0)),
					DetailID:        int(absEx * 10.0),
					Normal:          [3]float64{0, -0.5, 0.86},
					FlowTangent:     [2]float64{1, 0},
					Illum:           0.90,
					Spec:            0.85,
					SunColor:        carapaceIvory,
					MidColor:        carapaceTeal,
					ShadeColor:      hullShade,
					DeepShadowColor: hullDeep,
				}
			}
		}

		// =====================================================================
		// 4. KUBELKA-MUNK WATERCOLOR CUMULUS CLOUD BANKS & CELESTIAL SKY
		// =====================================================================
		// Organic billowing cloud banks with smooth scalloped crests (zero vertical cell lines!)
		for cIdx := 0; cIdx < 2; cIdx++ {
			fc := float64(cIdx)
			cloudCenter := (0.14 + 0.24*fc) - 0.06*x + 0.07*math.Sin(x*3.6+fc*2.1+phi) + 0.03*math.Cos(x*7.2-fc)
			cloudHalfW := 0.13 + 0.065*crag
			cloudDist := y - cloudCenter
			if math.Abs(cloudDist) <= cloudHalfW {
				uC := (cloudDist/cloudHalfW + 1.0) * 0.5 // 0 at sunlit upper crest, 1 at violet underside
				lobeSlope := 0.25 * math.Cos(x*3.6+fc*2.1+phi)
				lit := common.Clamp01(0.90 - 0.64*common.Smoothstep(0.38, 0.92, uC) - 0.12*lobeSlope)
				return render.MoebiusPixel{
					Depth:           6.5 + 2.0*fc,
					ObjID:           50 + cIdx,
					PartID:          500 + cIdx*10 + int(uC*2.5),
					DetailID:        int(uC * 3.0),
					Normal:          moebiusNorm3(lobeSlope, uC-0.5, 0.8),
					FlowTangent:     [2]float64{0.96, -0.25 * lobeSlope},
					Illum:           lit,
					Spec:            common.Smoothstep(0.3, 0.0, uC) * 0.70,
					SunColor:        cloudSun,
					MidColor:        cloudPeach,
					ShadeColor:      cloudViolet,
					DeepShadowColor: hullDeep,
					Stipple:         common.Smoothstep(0.42, 0.95, uC) * 0.46,
				}
			}
		}

		// Kubelka-Munk Watercolor Sky Gradient with Mineral Granulation
		skyT := common.Clamp01((y + 1.0) / 1.65)
		var skyCol [3]float64
		if skyT < 0.48 {
			skyCol = moebiusLerp3(skyTop, skyMid, skyT/0.48)
		} else {
			skyCol = moebiusLerp3(skyMid, skyLow, (skyT-0.48)/0.52)
		}
		// Subtle Azurite watercolor granulation in upper sky
		skyCol = render.EvaluateWatercolorWash(ne, ux, uy, skyT-0.65, render.PigmentCerulean, skyCol, render.WatercolorWashConfig{
			BaseThickness:     0.22,
			EdgeDarkening:     0.35,
			EdgeWidth:         0.08,
			GranulationAmount: 0.75,
			PaperFreq:         52.0,
			WetFeather:        0.15,
		})

		return render.MoebiusPixel{
			Depth:       100.0,
			ObjID:       1,
			PartID:      1,
			DetailID:    1,
			Normal:      [3]float64{0, 0, 1},
			FlowTangent: [2]float64{1, 0},
			Illum:       0.86,
			SunColor:    skyCol,
			MidColor:    skyMid,
			ShadeColor:  skyTop,
			Stipple:     common.Smoothstep(0.14, 0.0, math.Abs(skyT-0.46)) * 0.36,
		}
	})

	c.MergeRenderCanvas(rc)
}
