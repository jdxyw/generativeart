package arts

import (
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type haeckelRadiolaria struct {
	shells     int
	spineOrder int
	poreScale  float64
}

// NewHaeckelRadiolaria returns a HaeckelRadiolaria generator inspired by Ernst Haeckel's
// 1904 Kunstformen der Natur (Phaeodaria / Acantharia / Dorataspis): 3 concentric 3D
// perforated geodesic silica endoskeleton shells where 20 radial 3-bladed crystal spicules
// grow directly from solid raised silica bosses surrounded by symmetric 6-pore and 12-pore
// rosettes (zero floating-in-hole clipping), revealing a glowing cinnabar-gold central
// protoplasmic capsule over an abyssal Prussian-blue chromolithograph plate.
func NewHaeckelRadiolaria(shells, spineOrder int, poreScale float64) *haeckelRadiolaria {
	if shells <= 0 {
		shells = 3
	}
	if spineOrder <= 0 {
		spineOrder = 6
	}
	if poreScale <= 0 {
		poreScale = 1.0
	}
	return &haeckelRadiolaria{
		shells:     shells,
		spineOrder: spineOrder,
		poreScale:  poreScale,
	}
}

// Generative renders the static HaeckelRadiolaria natural-history chromolithograph plate.
func (hr *haeckelRadiolaria) Generative(c *generativeart.Canva) {
	hr.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360° looping HaeckelRadiolaria animation.
func (hr *haeckelRadiolaria) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	seed := c.Seed()
	if seed == 0 {
		seed = 20262402
	}
	phi := 2.0 * math.Pi * phase

	// 19th-Century Haeckel Chromolithograph & Biomineralized Porcelain-Opal Silica Palette
	silicaPorcelain := render.MixSpectralKubelkaMunk(render.PW(render.PigmentTitaniumWhite, 0.88), render.PW(render.PigmentYellowOchre, 0.12))
	opalCeladon := render.MixSpectralKubelkaMunk(render.PW(render.PigmentTitaniumWhite, 0.72), render.PW(render.PigmentCobaltTeal, 0.28))
	amberGold := render.MixSpectralKubelkaMunk(render.PW(render.PigmentGoldLeaf, 0.70), render.PW(render.PigmentYellowOchre, 0.30))
	cinnabarNucleus := render.MixSpectralKubelkaMunk(render.PW(render.PigmentVermilion, 0.76), render.PW(render.PigmentQuinacridoneRose, 0.24))
	coralProtoplasm := render.MixSpectralKubelkaMunk(render.PW(render.PigmentVermilion, 0.46), render.PW(render.PigmentGoldLeaf, 0.54))
	abyssalPrussian := render.MixSpectralKubelkaMunk(render.PW(render.PigmentUltramarine, 0.68), render.PW(render.PigmentPineSootInk, 0.32))
	ceruleanShade := render.MixSpectralKubelkaMunk(render.PW(render.PigmentAzurite, 0.52), render.PW(render.PigmentCobaltTeal, 0.48))
	deepIntaglio := render.MixSpectralKubelkaMunk(render.PW(render.PigmentPineSootInk, 0.86), render.PW(render.PigmentUltramarine, 0.14))

	lightDir3D := render.V3(-0.50, 0.62, 0.60).Normalize()

	// Exact icosahedral 5-fold symmetry around Y via AlignIcosahedral5FoldY:
	//  - Outer Perforated Silica Shell & 20 Radial Spicules rotate around Y by +(2*pi/5)*phase!
	//  - Middle Dodecahedral Medullary Cage counter-rotates around Y by -(2*pi/5)*phase!
	//  - Inner Protoplasmic Capsule pulses gently with sin(phi).
	rotOuterY := 0.24 + (2.0*math.Pi/5.0)*phase
	rotMidY := -0.18 - (2.0*math.Pi/5.0)*phase
	tiltX := 0.28
	pulse := 0.015 * math.Sin(phi)
	lOrder := common.ConstrainInt(hr.spineOrder, 4, 8)

	evalRadiolaria := func(p render.Vec3) (float64, int, int, render.Vec3) {
		pTilt := p.RotateX(tiltX)

		// 1. Outer Perforated Icosahedral Silica Endoskeleton Shell (R = 0.385)
		// In FoldIcosahedral, +Y is the exact center normal of each of the 20 icosahedral faces,
		// and (fOut.X, fOut.Z) are tangential coordinates on the face!
		qOut := render.AlignIcosahedral5FoldY(pTilt.RotateY(rotOuterY))
		fOut := render.FoldIcosahedral(qOut)
		rOut := qOut.Len()

		uDir := qOut.Scale(1.0 / math.Max(1e-4, rOut))
		yHarm := common.SphericalHarmonicCartesian(lOrder, 0, uDir.X, uDir.Y, uDir.Z) +
			0.45*common.SphericalHarmonicCartesian(lOrder, 5, uDir.X, uDir.Y, uDir.Z)
		shellR := 0.385 + 0.012*yHarm
		dShellSphere := math.Abs(rOut-shellR) - 0.017

		// Tangential polar coordinates (rFace, angFace) around the solid spicule boss at (fOut.X=0, fOut.Z=0):
		//  - Central region rFace < 0.048 is a SOLID raised silica pyramidal boss anchoring the radial spicule!
		//  - Ring 1 (rFace = 0.095): 6 circular/hexagonal pores arranged in a 6-fold floral rosette around the boss.
		//  - Ring 2 (rFace = 0.168): 12 circular pores arranged around the outer icosahedral triangle perimeter.
		rFace := math.Hypot(fOut.X, fOut.Z)
		angFace := math.Atan2(fOut.Z, fOut.X)

		// Fold angFace into 6 sectors for Ring 1 and 12 sectors for Ring 2
		sec6 := math.Pi / 3.0
		a6 := math.Mod(angFace+2.0*math.Pi, sec6) - 0.5*sec6
		p1X, p1Z := rFace*math.Cos(a6)-0.095, rFace*math.Sin(a6)
		dPoreRing1 := math.Hypot(p1X, p1Z) - 0.028*hr.poreScale

		sec12 := math.Pi / 6.0
		a12 := math.Mod(angFace+2.0*math.Pi+0.5*sec12, sec12) - 0.5*sec12
		p2X, p2Z := rFace*math.Cos(a12)-0.168, rFace*math.Sin(a12)
		dPoreRing2 := math.Hypot(p2X, p2Z) - 0.024*hr.poreScale

		dPoreCutter := math.Min(dPoreRing1, dPoreRing2)
		dOuterLattice := render.SDF3DSmoothSubtract(dPoreCutter, dShellSphere, 0.006)

		// Raised hexagonal silica rim collars around each pore + raised central pyramidal boss at rFace < 0.048
		poreRim := common.Smoothstep(0.010, 0.0, math.Abs(dPoreCutter)-0.004) * 0.005
		bossCone := math.Max(0.0, 0.052-rFace) * 0.45
		dOuterLattice -= poreRim + bossCone*common.Smoothstep(0.03, 0.0, math.Abs(rOut-shellR))

		// 20 Radial 3-Bladed Crystal Spicules growing directly out of the solid central boss (from fOut.Y = 0.36 to 0.68)
		tSpine := common.Clamp01((fOut.Y - 0.36) / 0.32)
		// Flared conical basal buttress (cortinar collar) smoothly anchors the spicule to the shell wall!
		baseFlare := 0.020 * math.Exp(-tSpine*6.5)
		spineRadius := baseFlare + 0.014*(1.0-0.84*tSpine) + 0.0022
		spineRadius *= 1.0 + 0.18*math.Cos(angFace*3.0)
		dVertexSpine := rFace - spineRadius
		dVertexSpine = math.Max(dVertexSpine, math.Abs(fOut.Y-0.52)-0.16)

		// Two tiers of slender, needle-sharp orthogonal cross-apophyses at fOut.Y = 0.47 and 0.56
		crossAxisDist := math.Min(math.Abs(fOut.X), math.Abs(fOut.Z))
		dBarb1 := math.Max(math.Hypot(crossAxisDist, fOut.Y-0.475)-0.0045, rFace-0.038)
		dBarb2 := math.Max(math.Hypot(crossAxisDist, fOut.Y-0.560)-0.0038, rFace-0.026)
		dSpicule := render.SDF3DSmoothUnion(dVertexSpine, math.Min(dBarb1, dBarb2), 0.004)

		dOuter := render.SDF3DSmoothUnion(dOuterLattice, dSpicule, 0.014)

		// 2. Middle Counter-Rotating Dodecahedral Medullary Cage (R = 0.255, strictly inside R = 0.385)
		qMid := pTilt.RotateY(rotMidY)
		fMid := render.FoldIcosahedral(render.AlignIcosahedral5FoldY(qMid))
		dMiddle := render.SDF3DPlatonicCage(qMid, 0, 0.255, 0.022)

		// 3. Inner Glowing Cinnabar-Gold Protoplasmic Central Capsule (R = 0.155, strictly inside R = 0.255)
		dCore := render.SDF3DSphere(qMid, 0.155+pulse) - 0.004*math.Cos(fMid.X*36.0)*math.Cos(fMid.Z*36.0)

		minD := dOuter
		objID := 10
		partID := 100 + int(math.Floor(rFace*24.0))
		loc := fOut
		if fOut.Y > shellR+0.018 && dSpicule < dOuterLattice+0.008 {
			objID = 11
			partID = 150 + int(math.Floor(fOut.Y*28.0))
		}
		if hr.shells >= 2 && dMiddle < minD {
			minD = dMiddle
			objID = 12
			partID = 200 + int(math.Floor((fMid.X+fMid.Z)*18.0))
			loc = fMid
		}
		if hr.shells >= 3 && dCore < minD {
			minD = dCore
			objID = 13
			partID = 300 + int(math.Floor(fMid.Y*24.0))
			loc = fMid
		}
		return minD, objID, partID, loc
	}

	sceneSDF := func(p render.Vec3) float64 {
		d, _, _, _ := evalRadiolaria(p)
		return d
	}

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)
	cfg := render.DefaultMoebiusConfig(seed)
	cfg.HatchSpacingPx = 3.6

	camPos := render.V3(0.0, 0.0, 1.54)

	render.RenderMoebiusScene(rc, cfg, func(ux, uy float64) render.MoebiusPixel {
		vx := (ux - 0.5) * 2.0
		vy := (0.5 - uy) * 2.0
		rScreen := math.Hypot(vx, vy)
		angScreen := math.Atan2(vy, vx)

		// 1. Stationary 19th-Century Haeckel Brass Specimen Bezel & Corner Diatom Medallions
		if rScreen > 0.90 && rScreen < 0.935 {
			tick := 1.0 - common.Smoothstep(0.0, 0.12, math.Abs(math.Sin(angScreen*36.0)))
			return render.MoebiusPixel{
				Depth:           0.88,
				ObjID:           4,
				PartID:          40 + (int((angScreen+math.Pi)*18.0) & 15),
				DetailID:        int(tick * 4.0),
				Normal:          [3]float64{-vx / rScreen, -vy / rScreen, 0.75},
				FlowTangent:     [2]float64{-vy / rScreen, vx / rScreen},
				Illum:           0.68 + 0.22*tick,
				Spec:            0.42,
				SunColor:        amberGold,
				MidColor:        silicaPorcelain,
				ShadeColor:      abyssalPrussian,
				DeepShadowColor: deepIntaglio,
				CrackInk:        tick * 0.55,
			}
		}

		cornerX := math.Abs(vx) - 0.78
		cornerY := math.Abs(vy) - 0.78
		rCorner := math.Hypot(cornerX, cornerY)
		if rScreen > 0.935 && rCorner < 0.12 {
			aCorner := math.Atan2(cornerY, cornerX)
			spoke := 0.5 + 0.5*math.Cos(aCorner*12.0)
			ring := 0.5 + 0.5*math.Cos(rCorner*110.0)
			illum := common.Clamp01(0.45 + 0.35*spoke + 0.18*ring)
			return render.MoebiusPixel{
				Depth:           0.90,
				ObjID:           5,
				PartID:          50 + int(rCorner*40.0),
				DetailID:        int(spoke * 8.0),
				Normal:          [3]float64{-cornerX * 4.0, -cornerY * 4.0, 0.85},
				FlowTangent:     [2]float64{-math.Sin(aCorner), math.Cos(aCorner)},
				Illum:           illum,
				SunColor:        silicaPorcelain,
				MidColor:        opalCeladon,
				ShadeColor:      abyssalPrussian,
				DeepShadowColor: deepIntaglio,
				CrackInk:        (1.0 - ring) * 0.48,
				Stipple:         0.14 * (1.0 - illum),
			}
		}

		rd := render.V3(vx*0.56, vy*0.56, -1.0).Normalize()
		hit := render.RaymarchSDF3DBounded(sceneSDF, camPos, rd, lightDir3D, render.V3(0, 0, 0), 0.74, 3.0, 80)

		if hit.Hit {
			_, objID, partID, loc := evalRadiolaria(hit.Pos)
			ndotl := math.Max(0.0, hit.Normal.Dot(lightDir3D))
			cosView := common.Clamp01(-rd.Dot(hit.Normal))
			// Translucent opalescent silica rim glow (keep minimum illumination >= 0.30 on silica so it stays luminous ivory-opal!)
			rimSSS := math.Pow(1.0-cosView, 2.0) * 0.52
			illum := common.Clamp01(0.28 + 0.68*ndotl*(0.45+0.55*hit.Shadow)*hit.AO + rimSSS*0.48)

			inkW, chalkW, pc := render.Evaluate3DWeingartenEngraving(sceneSDF, loc, illum, 64.0)

			var sunCol, midCol, shadeCol [3]float64
			switch objID {
			case 10:
				// Outer Perforated Geodesic Silica Shell (Luminous Porcelain-Ivory & Opal-Celadon Sheen)
				tfR, tfG, tfB := render.SampleThinFilmLinearRGB(cosView, 340.0+45.0*pc.MeanH, 1.44)
				sunCol = [3]float64{
					silicaPorcelain[0]*0.86 + tfR*0.14,
					silicaPorcelain[1]*0.86 + tfG*0.14,
					silicaPorcelain[2]*0.86 + tfB*0.14,
				}
				midCol = opalCeladon
				shadeCol = ceruleanShade
			case 11:
				// Radial 3-Bladed Crystal Spicules & Verticil Barbs (Warm Amber-Gold base to Porcelain-Ivory tips)
				tipT := common.Clamp01((hit.Pos.Len() - 0.38) * 3.4)
				for i := 0; i < 3; i++ {
					sunCol[i] = amberGold[i]*(1.0-tipT) + silicaPorcelain[i]*tipT
				}
				midCol = amberGold
				shadeCol = ceruleanShade
			case 12:
				// Middle Medullary Dodecahedral Cage (Burnished 24k Gold)
				sunCol = amberGold
				midCol = coralProtoplasm
				shadeCol = abyssalPrussian
				illum = math.Max(0.36, illum)
			default:
				// Inner Glowing Cinnabar-Vermilion Protoplasmic Capsule
				sunCol = coralProtoplasm
				midCol = cinnabarNucleus
				shadeCol = cinnabarNucleus
				illum = math.Min(1.0, illum+0.28)
			}

			return render.MoebiusPixel{
				Depth:           hit.T,
				ObjID:           objID,
				PartID:          partID,
				DetailID:        int(math.Floor((pc.K1 + 2.0) * 6.0)),
				Normal:          [3]float64{hit.Normal.X, hit.Normal.Y, hit.Normal.Z},
				FlowTangent:     [2]float64{pc.E2.X, pc.E2.Y},
				Illum:           common.Clamp01(illum + 0.22*chalkW),
				Spec:            rimSSS*0.85 + chalkW*0.45,
				Bounce:          0.38,
				SunColor:        sunCol,
				MidColor:        midCol,
				ShadeColor:      shadeCol,
				DeepShadowColor: deepIntaglio,
				CrackInk:        inkW * 0.42,
				Stipple:         0.10 * (1.0 - illum),
			}
		}

		// 2. Abyssal Prussian-Blue Chromolithograph Backdrop with Subtle Bioluminescent Halo
		halo := math.Exp(-rScreen * 2.0)
		rayMod := 0.5 + 0.5*math.Cos(angScreen*20.0)
		bgGlow := halo * (0.78 + 0.22*rayMod)
		return render.MoebiusPixel{
			Depth:           95.0,
			ObjID:           1,
			PartID:          1,
			DetailID:        int(rScreen * 10.0),
			Normal:          [3]float64{0, 0, 1},
			FlowTangent:     [2]float64{-math.Sin(angScreen), math.Cos(angScreen)},
			Illum:           common.Clamp01(0.14 + 0.42*bgGlow),
			SunColor:        opalCeladon,
			MidColor:        ceruleanShade,
			ShadeColor:      abyssalPrussian,
			DeepShadowColor: deepIntaglio,
			Stipple:         0.18 * (1.0 - bgGlow),
		}
	})
	c.MergeRenderCanvas(rc)
}
