package arts

import (
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type muranoGlassNautilus struct {
	chambers   int
	facets     int
	dispersion float64
}

// NewMuranoGlassNautilus returns a MuranoGlassNautilus generator that raymarches a
// 3D Venetian Murano blown-glass logarithmic-spiral Nautilus chambered vessel and
// an interior rotating flint-diamond icosahedral core cradled inside a non-intersecting
// graduated brass meridian ring above an engraved obsidian-and-24k-gold celestial
// astrolabe altar, using Phase 2–5 3D dielectric Snell refraction, Beer-Lambert
// volume absorption, Abbe RGB chromatic dispersion, and refractive caustics.
func NewMuranoGlassNautilus(chambers, facets int, dispersion float64) *muranoGlassNautilus {
	if chambers <= 0 {
		chambers = 14
	}
	if facets <= 0 {
		facets = 5
	}
	if dispersion <= 0 {
		dispersion = 0.048
	}
	return &muranoGlassNautilus{
		chambers:   chambers,
		facets:     facets,
		dispersion: dispersion,
	}
}

// Generative renders the static MuranoGlassNautilus sculpture & caustic astrolabe print.
func (mg *muranoGlassNautilus) Generative(c *generativeart.Canva) {
	mg.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360° looping MuranoGlassNautilus animation.
func (mg *muranoGlassNautilus) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	seed := c.Seed()
	if seed == 0 {
		seed = 20262401
	}
	ne := common.NewNoiseEngine(seed)
	phi := 2.0 * math.Pi * phase

	// 8-Band Spectral Kubelka-Munk Venetian Dichroic Glass & Renaissance Astrolabe Palette
	goldLeaf := render.MixSpectralKubelkaMunk(render.PW(render.PigmentGoldLeaf, 0.76), render.PW(render.PigmentTitaniumWhite, 0.24))
	amberBronze := render.MixSpectralKubelkaMunk(render.PW(render.PigmentGoldLeaf, 0.66), render.PW(render.PigmentBurntSienna, 0.34))
	emeraldGlass := render.MixSpectralKubelkaMunk(render.PW(render.PigmentCobaltTeal, 0.65), render.PW(render.PigmentMalachite, 0.35))
	cobaltGlass := render.MixSpectralKubelkaMunk(render.PW(render.PigmentAzurite, 0.68), render.PW(render.PigmentCobaltTeal, 0.32))
	rubyAventurine := render.MixSpectralKubelkaMunk(render.PW(render.PigmentVermilion, 0.72), render.PW(render.PigmentQuinacridoneRose, 0.28))
	sapphireDeep := render.MixSpectralKubelkaMunk(render.PW(render.PigmentUltramarine, 0.72), render.PW(render.PigmentPineSootInk, 0.28))
	obsidianMarble := render.MixSpectralKubelkaMunk(render.PW(render.PigmentPineSootInk, 0.82), render.PW(render.PigmentAzurite, 0.18))
	warmParchment := render.MixSpectralKubelkaMunk(render.PW(render.PigmentYellowOchre, 0.28), render.PW(render.PigmentTitaniumWhite, 0.72))
	venetianStone := render.MixSpectralKubelkaMunk(render.PW(render.PigmentYellowOchre, 0.52), render.PW(render.PigmentRawUmber, 0.48))
	deepIntaglio := render.MixSpectralKubelkaMunk(render.PW(render.PigmentPineSootInk, 0.86), render.PW(render.PigmentUltramarine, 0.14))

	lightDir3D := render.V3(-0.48, 0.66, 0.58).Normalize()

	// Seamless periodic rotations for 20-frame loop:
	//  - Blown-glass Nautilus gently rocks by sin(phi)/cos(phi) while chamber flutes advance by +phi.
	//  - Inner 5-fold Icosahedral Diamond Core rotates around Y by +(2*pi/5)*phase (exact 5-fold invariance).
	//  - Outer Graduated Brass Meridian Ring (at R = 0.455, completely outside the R <= 0.375 glass shell)
	//    rotates around its axis by -(2*pi/24)*phase.
	rockZ := 0.12 + 0.05*math.Sin(phi)
	rockY := 0.20 + 0.10*math.Cos(phi)
	coreRotY := (2.0 * math.Pi / 5.0) * phase
	nChambers := float64(common.MaxInt(10, mg.chambers))
	shellCenter := render.V3(0.0, 0.13, 0.0)

	// 1. Symmetrical Full-Whorl Blown-Glass Nautilus / Ammonite + Inner 5-Fold Diamond Core (R_max <= 0.37)
	glassSDF := func(p render.Vec3) float64 {
		q := p.Sub(shellCenter).RotateZ(rockZ).RotateY(rockY)
		rXY := math.Hypot(q.X, q.Y)
		theta := math.Atan2(q.Y, q.X) // in [-pi, pi]

		// Continuous closed fluted Murano blown-glass chambered ring-nautilus with logarithmic spiral sutures:
		//  - Outer chambered ring (spiralR varies smoothly with cos(theta) while tubeR tapers smoothly around the circle,
		//    eliminating any harsh cut edge while showcasing 18 blown-glass chambers around the entire 360°!)
		uAngle := 0.5 + 0.5*math.Sin(theta-0.35) // smooth 360-degree periodic modulation in [0, 1]
		spiralR := 0.21 + 0.035*uAngle
		tubeR := 0.095 + 0.032*uAngle

		fluteWave := math.Cos(theta*nChambers - phi)
		spiralTwist := math.Cos(math.Atan2(q.Z, rXY-spiralR)*8.0 + theta*4.0 - phi)
		dTube := math.Hypot(rXY-spiralR, q.Z) - tubeR*(1.0+0.11*fluteWave+0.035*spiralTwist)

		// Hollow blown-glass wall + internal curved chamber septa
		dWall := math.Abs(dTube+0.014) - 0.018
		septa := math.Abs(math.Sin(theta*nChambers*0.5-0.5*phi))*tubeR - 0.010
		dSepta := math.Max(dTube+0.006, septa)
		dNautilus := render.SDF3DSmoothUnion(dWall, dSepta, 0.014)

		// Inner rotating 5-fold Icosahedral Flint-Diamond Jewel Core at the umbilical center (R = 0.125)
		qCore := render.AlignIcosahedral5FoldY(q.RotateY(coreRotY))
		fIco := render.FoldIcosahedral(qCore)
		dCore := render.SDF3DOctahedron(fIco, 0.125) - 0.006

		return render.SDF3DSmoothUnion(dNautilus, dCore, 0.020)
	}

	// 2. Non-Intersecting Outer Brass Meridian Ring (R = 0.455 > 0.38), Chalice Base & Astrolabe Altar
	opaqueScene := func(p render.Vec3) (float64, int, int, render.Vec3) {
		// Recessed Obsidian & 24k Gold Astrolabe Altar Disc at Y = -0.39
		pAltar := p.Sub(render.V3(0.0, -0.39, 0.0))
		dDisc := render.SDF3DCylinderY(pAltar, 0.60, 0.032)
		rAltar := math.Hypot(pAltar.X, pAltar.Z)
		ringGrooves := 0.004 * math.Cos(rAltar*42.0)
		dAltar := dDisc + ringGrooves

		// Outer Graduated Brass Meridian Ring encircling the Nautilus with a clean air gap (R = 0.455)
		qRing := p.Sub(shellCenter).RotateY(0.18).RotateX(0.10)
		dRing := math.Hypot(math.Hypot(qRing.X, qRing.Y)-0.455, qRing.Z) - 0.017

		// Compact Brass Pedestal Socket strictly at Y in [-0.39, -0.325] (connecting bottom of qRing to Altar with no floating disc!)
		pStem := p.Sub(render.V3(0.0, -0.355, 0.0))
		dStem := render.SDF3DCylinderY(pStem, 0.052, 0.032) - 0.004
		dMount := render.SDF3DSmoothUnion(dRing, dStem, 0.012)

		minD := dAltar
		objID := 20
		angAltar := math.Atan2(pAltar.Z, pAltar.X)
		sector := int(math.Floor((angAltar + math.Pi) / (2.0 * math.Pi) * 16.0))
		partID := 200 + int(rAltar*14.0)*16 + (sector & 15)
		loc := pAltar

		if dMount < minD {
			minD = dMount
			objID = 21
			angR := math.Atan2(qRing.Y, qRing.X) - phi/24.0
			gSec := int(math.Floor((angR + math.Pi) / (2.0 * math.Pi) * 24.0))
			partID = 300 + (gSec & 31)
			loc = qRing
		}
		return minD, objID, partID, loc
	}

	opaqueSDF := func(p render.Vec3) float64 {
		d, _, _, _ := opaqueScene(p)
		return d
	}

	// Clean, luminous Venetian studio optical environment seen through refracted/reflected Murano crystal rays
	envRadiance := func(ro, rd render.Vec3, ch int) [3]float64 {
		sunDot := math.Max(0.0, rd.Dot(lightDir3D))
		sunHighlight := math.Pow(sunDot, 22.0) * 1.45
		// Dichroic internal Venetian Murano glass bands (Cobalt-Teal <-> Emerald <-> 24k Aventurine Gold <-> Ruby)
		bandWave := 0.5 + 0.5*math.Sin(rd.X*6.5+rd.Y*4.5-phi)
		swirlWave := 0.5 + 0.5*math.Cos(rd.Y*8.0-rd.Z*5.0+phi)

		var env [3]float64
		for i := 0; i < 3; i++ {
			dichroic := cobaltGlass[i]*(1.0-bandWave) + emeraldGlass[i]*bandWave
			warmGold := goldLeaf[i]*0.72 + rubyAventurine[i]*0.28
			env[i] = dichroic*(0.55+0.45*(1.0-swirlWave)) + warmGold*(0.42*swirlWave) + warmParchment[i]*sunHighlight
		}
		// Chromatic prism separation for dispersed R/G/B rays
		if ch == 0 {
			env[0] = env[0]*1.24 + 0.08*rubyAventurine[0]
		} else if ch == 1 {
			env[1] = env[1]*1.18 + 0.06*emeraldGlass[1]
		} else if ch == 2 {
			env[2] = env[2]*1.26 + 0.08*cobaltGlass[2]
		}
		return env
	}

	dielectricMat := render.DielectricMuranoEmerald()
	dielectricMat.DispersionDelta = mg.dispersion
	dielectricMat.AbsorbSigma = [3]float64{0.65, 0.14, 0.28}
	dielectricMat.MaxInternalBounces = 3

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)
	cfg := render.DefaultMoebiusConfig(seed)
	cfg.HatchSpacingPx = 3.5

	camPos := render.V3(0.0, 0.15, 1.54)

	render.RenderMoebiusScene(rc, cfg, func(ux, uy float64) render.MoebiusPixel {
		vx := (ux - 0.5) * 2.0
		vy := (0.5 - uy) * 2.0

		// 1. Stationary Venetian Gothic Carved Spandrel Arch in the upper corners
		archCurve := vy - (0.64 + 0.40*vx*vx)
		if math.Abs(vx) > 0.80 || archCurve > 0.14 {
			stoneRel := render.EvaluateMoebiusStoneRelief(ne, ux, uy, 8.5, 8.5, 0.35)
			nFrame := render.V3(-vx*0.55+stoneRel.NormalMod[0]*0.3, -0.35+stoneRel.NormalMod[1]*0.3, 0.78).Normalize()
			illum := common.Clamp01((0.35 + 0.65*math.Max(0.0, nFrame.Dot(lightDir3D))) * (0.85 + 0.15*stoneRel.ColorMod))
			rosetteR := math.Hypot(math.Abs(vx)-0.83, vy-0.78)
			sunCol := venetianStone
			if rosetteR < 0.10 {
				sunCol = goldLeaf
				illum = math.Min(1.0, illum+0.18)
			}
			return render.MoebiusPixel{
				Depth:           0.85,
				ObjID:           4,
				PartID:          40 + (stoneRel.BlockID & 15),
				DetailID:        stoneRel.BlockID,
				Normal:          [3]float64{nFrame.X, nFrame.Y, nFrame.Z},
				FlowTangent:     [2]float64{-nFrame.Y, nFrame.X},
				Illum:           illum,
				Bounce:          0.25,
				SunColor:        sunCol,
				MidColor:        amberBronze,
				ShadeColor:      sapphireDeep,
				DeepShadowColor: deepIntaglio,
				CrackInk:        stoneRel.CrackInk * 0.62,
				TickMask:        stoneRel.PockInk * 0.42,
				Stipple:         0.15 * (1.0 - illum),
			}
		}

		rd := render.V3(vx*0.58, vy*0.58-0.05, -1.0).Normalize()

		// 2. Raymarch both the Dielectric Murano Nautilus Volume and the Non-Intersecting Brass Ring / Altar
		glassHit := render.RaymarchSDF3DBounded(glassSDF, camPos, rd, lightDir3D, shellCenter, 0.44, 3.0, 68)
		opaqueHit := render.RaymarchSDF3DBounded(opaqueSDF, camPos, rd, lightDir3D, render.V3(0, -0.05, 0), 0.78, 3.4, 68)

		if glassHit.Hit && (!opaqueHit.Hit || glassHit.T < opaqueHit.T) {
			refrRGB, _ := render.TraceSDF3DDielectric(glassSDF, camPos, rd, lightDir3D, 3.0, dielectricMat, envRadiance)
			pc := render.CalcSDF3DWeingartenCurvature(glassSDF, glassHit.Pos, 0.004)

			qLoc := glassHit.Pos.Sub(shellCenter).RotateZ(rockZ).RotateY(rockY)
			rLoc := qLoc.Len()
			cosView := common.Clamp01(-rd.Dot(glassHit.Normal))
			fresnelRim := math.Pow(1.0-cosView, 2.4)

			// Rich Murano Dichroic Cobalt-Emerald-Ruby-Gold chamber color bands inside the blown glass!
			angLoc := math.Atan2(qLoc.Y, qLoc.X)
			fluteCrest := 0.5 + 0.5*math.Cos(angLoc*nChambers*0.5-phi)
			rubyRibbon := (0.5 + 0.5*math.Sin(angLoc*nChambers*0.5-phi)) * common.Smoothstep(0.10, 0.30, rLoc)

			var sunCol, midCol, shadeCol [3]float64
			for cIdx := 0; cIdx < 3; cIdx++ {
				chamberTint := emeraldGlass[cIdx]*(1.0-rubyRibbon*0.65) + rubyAventurine[cIdx]*(rubyRibbon*0.65)
				crestGold := chamberTint*(1.0-0.58*fluteCrest) + goldLeaf[cIdx]*(0.58*fluteCrest)
				base := common.Clamp01(0.48*refrRGB[cIdx] + 0.52*crestGold)
				sunCol[cIdx] = common.Clamp01(base*1.12 + warmParchment[cIdx]*0.32*fresnelRim)
				midCol[cIdx] = base
				shadeCol[cIdx] = cobaltGlass[cIdx]*0.62 + base*0.38
			}
			if rLoc < 0.135 {
				// Inner 5-fold Icosahedral Diamond-Ruby Core
				for cIdx := 0; cIdx < 3; cIdx++ {
					sunCol[cIdx] = warmParchment[cIdx]*0.65 + goldLeaf[cIdx]*0.35
					midCol[cIdx] = rubyAventurine[cIdx]*0.65 + goldLeaf[cIdx]*0.35
					shadeCol[cIdx] = sapphireDeep[cIdx]
				}
			}
			illum := common.Clamp01(0.32 + 0.42*fluteCrest + 0.25*fresnelRim + 0.18*glassHit.Shadow)
			chIdx := int(math.Floor((angLoc - phi/nChambers + 4.0*math.Pi) / (2.0 * math.Pi) * nChambers)) % int(nChambers)
			if rLoc < 0.135 {
				chIdx = 60 + int(math.Floor((qLoc.X+qLoc.Y+qLoc.Z)*18.0))
			}

			return render.MoebiusPixel{
				Depth:           glassHit.T,
				ObjID:           10,
				PartID:          100 + (chIdx & 63),
				DetailID:        int(math.Floor(rLoc * 28.0)),
				Normal:          [3]float64{glassHit.Normal.X, glassHit.Normal.Y, glassHit.Normal.Z},
				FlowTangent:     [2]float64{pc.E2.X, pc.E2.Y},
				Illum:           illum,
				Spec:            0.48 + 0.48*fresnelRim,
				Bounce:          0.42,
				SunColor:        sunCol,
				MidColor:        midCol,
				ShadeColor:      shadeCol,
				DeepShadowColor: sapphireDeep,
				CrackInk:        common.Smoothstep(5.0, 13.0, math.Abs(pc.MeanH)) * 0.42,
				Stipple:         0.06 * (1.0 - illum),
			}
		}

		if opaqueHit.Hit {
			_, opID, partID, opLoc := opaqueScene(opaqueHit.Pos)
			ndotl := math.Max(0.0, opaqueHit.Normal.Dot(lightDir3D))
			illum := common.Clamp01((0.18 + 0.82*ndotl*(0.38+0.62*opaqueHit.Shadow)) * opaqueHit.AO)
			inkW, chalkW := render.Evaluate3DBurinEngraving(opLoc, opaqueHit.Normal, illum, 68.0)

			if opID == 20 {
				// Recessed Obsidian & 24k Gold Astrolabe Altar illuminated by Focused Refractive Caustics!
				caustic := render.EvalSDF3DCausticIntensity(glassSDF, opaqueHit.Pos, opaqueHit.Normal, lightDir3D, dielectricMat, 3.4)
				rA := math.Hypot(opLoc.X, opLoc.Z)
				angA := math.Atan2(opLoc.Z, opLoc.X)
				ringInlay := 1.0 - common.Smoothstep(0.0, 0.14, math.Abs(math.Sin(rA*42.0)))
				spokeInlay := 1.0 - common.Smoothstep(0.0, 0.10, math.Abs(math.Sin(angA*16.0)))
				inlay := math.Max(ringInlay, spokeInlay*0.75)

				// Concentric emerald/gold/ruby refractive caustic arcs projected under the glass Nautilus
				causticWave := math.Pow(0.5+0.5*math.Cos(rA*34.0-angA*2.0-phi), 3.0) * math.Exp(-rA*3.2)

				var sunCol, midCol, shadeCol [3]float64
				for cIdx := 0; cIdx < 3; cIdx++ {
					cBoost := common.Clamp01((caustic[cIdx]-0.25)*0.85 + causticWave*0.75)
					base := obsidianMarble[cIdx]*(1.0-inlay) + goldLeaf[cIdx]*inlay
					sunCol[cIdx] = common.Clamp01(base + cBoost*emeraldGlass[cIdx]*0.65 + cBoost*goldLeaf[cIdx]*0.55)
					midCol[cIdx] = common.Clamp01(base*0.80 + cBoost*emeraldGlass[cIdx]*0.50)
					shadeCol[cIdx] = common.Clamp01(obsidianMarble[cIdx]*0.68 + cBoost*0.25)
				}
				illum = common.Clamp01(illum + 0.35*inlay + 0.35*causticWave + 0.25*(caustic[0]+caustic[1]+caustic[2])/3.0)
				return render.MoebiusPixel{
					Depth:           opaqueHit.T,
					ObjID:           opID,
					PartID:          partID,
					DetailID:        int(rA * 36.0),
					Normal:          [3]float64{opaqueHit.Normal.X, opaqueHit.Normal.Y, opaqueHit.Normal.Z},
					FlowTangent:     [2]float64{-math.Sin(angA), math.Cos(angA)},
					Illum:           illum,
					Spec:            inlay*0.45 + causticWave*0.55,
					Bounce:          0.30,
					SunColor:        sunCol,
					MidColor:        midCol,
					ShadeColor:      shadeCol,
					DeepShadowColor: deepIntaglio,
					CrackInk:        math.Max(ringInlay*0.45, inkW*0.50),
					Stipple:         0.16 * (1.0 - illum),
				}
			}

			// Outer Graduated Brass Meridian Ring & Turned Chalice Pedestal Stem
			return render.MoebiusPixel{
				Depth:           opaqueHit.T,
				ObjID:           opID,
				PartID:          partID,
				DetailID:        partID,
				Normal:          [3]float64{opaqueHit.Normal.X, opaqueHit.Normal.Y, opaqueHit.Normal.Z},
				FlowTangent:     [2]float64{-opaqueHit.Normal.Y, opaqueHit.Normal.X},
				Illum:           common.Clamp01(illum + 0.24*chalkW),
				Spec:            0.52 + 0.35*chalkW,
				Bounce:          0.36,
				SunColor:        goldLeaf,
				MidColor:        amberBronze,
				ShadeColor:      sapphireDeep,
				DeepShadowColor: deepIntaglio,
				CrackInk:        inkW * 0.58,
				Stipple:         0.12 * (1.0 - illum),
			}
		}

		// 3. Background Venetian Sanctuary Cupola & Subtle Golden Astrolabe Filigree
		rCenter := math.Hypot(vx, vy-0.08)
		angCenter := math.Atan2(vy-0.08, vx)
		rayBurst := math.Pow(0.5+0.5*math.Cos(angCenter*16.0), 6.0) * math.Exp(-rCenter*2.4) * 0.32
		halo := math.Exp(-rCenter * 2.2)
		skyCol := render.LerpSpectralKubelkaMunk(render.PigmentUltramarine, render.PigmentAzurite, common.Clamp01(halo*0.65))
		skySun := render.LerpSpectralKubelkaMunk(render.PigmentAzurite, render.PigmentGoldLeaf, common.Clamp01(rayBurst+halo*0.28))

		vaultRing := 1.0 - common.Smoothstep(0.0, 0.006, math.Abs(rCenter-0.58))
		vaultRing2 := 1.0 - common.Smoothstep(0.0, 0.006, math.Abs(rCenter-0.72))
		vaultSpoke := (1.0 - common.Smoothstep(0.0, 0.008, math.Abs(math.Sin(angCenter*8.0))*rCenter)) * common.Smoothstep(0.48, 0.58, rCenter) * (1.0 - common.Smoothstep(0.72, 0.78, rCenter))
		filigree := math.Max(math.Max(vaultRing, vaultRing2), vaultSpoke*0.75)

		return render.MoebiusPixel{
			Depth:           95.0,
			ObjID:           1,
			PartID:          1 + int(rCenter*6.0),
			DetailID:        int((angCenter + math.Pi) * 6.0),
			Normal:          [3]float64{0, 0, 1},
			FlowTangent:     [2]float64{-math.Sin(angCenter), math.Cos(angCenter)},
			Illum:           common.Clamp01(0.18 + 0.36*halo + 0.25*rayBurst + 0.52*filigree),
			SunColor:        skySun,
			MidColor:        skyCol,
			ShadeColor:      sapphireDeep,
			DeepShadowColor: deepIntaglio,
			CrackInk:        filigree * 0.55,
			Stipple:         0.16 * (1.0 - halo),
		}
	})
	c.MergeRenderCanvas(rc)
}
