package arts

import (
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type weingartenTorusKnot struct {
	pWind     int
	qWind     int
	tubeScale float64
}

// NewWeingartenTorusKnot returns a WeingartenTorusKnot generator combining M.C. Escher,
// Albrecht Dürer, and Song-Dynasty mineral watercolor washes: a continuous 3D fluted
// (p, q) trefoil torus-knot monument surrounding a faceted 24k-gold core crystal and
// framed by an outer graduated celestial ring with exact topological clearance (zero
// interpenetration), shaded via Phase 5's intrinsic 3D Weingarten shape operator over
// an organic full-bleed 2D Wet-on-Wet Capillary Watercolor Bleed scroll.
func NewWeingartenTorusKnot(pWind, qWind int, tubeScale float64) *weingartenTorusKnot {
	if pWind <= 0 {
		pWind = 2
	}
	if qWind <= 0 {
		qWind = 3
	}
	if tubeScale <= 0 {
		tubeScale = 1.0
	}
	return &weingartenTorusKnot{
		pWind:     pWind,
		qWind:     qWind,
		tubeScale: tubeScale,
	}
}

// Generative renders the static WeingartenTorusKnot sculpture & wet-watercolor scroll.
func (wt *weingartenTorusKnot) Generative(c *generativeart.Canva) {
	wt.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360° looping WeingartenTorusKnot animation.
func (wt *weingartenTorusKnot) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	seed := c.Seed()
	if seed == 0 {
		seed = 20262404
	}
	ne := common.NewNoiseEngine(seed)
	phi := 2.0 * math.Pi * phase

	// 8-Band Spectral Kubelka-Munk Song-Dynasty Mineral & Renaissance Intaglio Palette
	mudGold := render.MixSpectralKubelkaMunk(render.PW(render.PigmentGoldLeaf, 0.76), render.PW(render.PigmentTitaniumWhite, 0.24))
	amberBronze := render.MixSpectralKubelkaMunk(render.PW(render.PigmentGoldLeaf, 0.68), render.PW(render.PigmentBurntSienna, 0.32))
	warmIvory := render.MixSpectralKubelkaMunk(render.PW(render.PigmentTitaniumWhite, 0.82), render.PW(render.PigmentYellowOchre, 0.18))
	warmParchment := render.MixSpectralKubelkaMunk(render.PW(render.PigmentYellowOchre, 0.36), render.PW(render.PigmentTitaniumWhite, 0.64))
	cinnabarCore := render.MixSpectralKubelkaMunk(render.PW(render.PigmentVermilion, 0.72), render.PW(render.PigmentGoldLeaf, 0.28))
	malachitePatina := render.MixSpectralKubelkaMunk(render.PW(render.PigmentMalachite, 0.68), render.PW(render.PigmentCobaltTeal, 0.32))
	azuriteShadow := render.MixSpectralKubelkaMunk(render.PW(render.PigmentAzurite, 0.62), render.PW(render.PigmentRawUmber, 0.38))
	carvedStone := render.MixSpectralKubelkaMunk(render.PW(render.PigmentYellowOchre, 0.54), render.PW(render.PigmentRawUmber, 0.46))
	umberShade := render.MixSpectralKubelkaMunk(render.PW(render.PigmentRawUmber, 0.64), render.PW(render.PigmentUltramarine, 0.36))
	deepIntaglio := render.MixSpectralKubelkaMunk(render.PW(render.PigmentPineSootInk, 0.86), render.PW(render.PigmentRawUmber, 0.14))

	washCfg := render.DefaultWatercolorWashConfig()
	washCfg.BaseThickness = 0.48
	washCfg.EdgeDarkening = 0.65

	lightDir3D := render.V3(-0.50, 0.62, 0.60).Normalize()

	// Exact 3-fold rotational symmetry of the (2, 3) Trefoil Torus Knot around Z:
	//  - Knot centerline radius R(t) = 0.29 + 0.105*cos(3*t), tubeR = 0.072:
	//    -> Inner clearance hole: r >= 0.29 - 0.105 - 0.072 = 0.113.
	//    -> Outer envelope radius: r <= 0.29 + 0.105 + 0.072 = 0.467.
	//  - Central 5-fold Star-Core sits at r <= 0.092 (< 0.113 -> ZERO collision!).
	//  - Outer Graduated Astrolabe Ring sits at R = 0.515 (> 0.467 -> ZERO collision!).
	qSym := common.MaxInt(3, wt.qWind)
	rotKnotZ := 0.22 + (2.0*math.Pi/float64(qSym))*phase
	coreRotY := -(2.0 * math.Pi / 4.0) * phase
	tiltX := 0.30
	tubeR := 0.072 * wt.tubeScale

	const knotSamples = 84
	var knotPts [knotSamples]render.Vec3
	var knotTangents [knotSamples]render.Vec3
	for i := 0; i < knotSamples; i++ {
		t := (float64(i) / float64(knotSamples)) * 2.0 * math.Pi
		rMaj := 0.29 + 0.105*math.Cos(3.0*t)
		x := rMaj * math.Cos(2.0*t)
		y := rMaj * math.Sin(2.0*t)
		z := 0.135 * math.Sin(3.0*t)
		knotPts[i] = render.V3(x, y, z)
	}
	for i := 0; i < knotSamples; i++ {
		prev := knotPts[(i-1+knotSamples)%knotSamples]
		next := knotPts[(i+1)%knotSamples]
		knotTangents[i] = next.Sub(prev).Normalize()
	}

	evalKnotScene := func(p render.Vec3) (float64, int, int, render.Vec3, float64, float64) {
		pTilt := p.RotateX(tiltX)
		q := pTilt.RotateZ(rotKnotZ)

		// 1. Continuous C2 3D Fluted Trefoil Torus Knot Sculpture (0.113 <= r <= 0.467)
		minDistSq := 1e9
		bestIdx := 0
		bestTSeg := 0.0
		var bestClosest render.Vec3
		for i := 0; i < knotSamples; i++ {
			a := knotPts[i]
			b := knotPts[(i+1)%knotSamples]
			ab := b.Sub(a)
			abLenSq := ab.Dot(ab)
			tSeg := 0.0
			if abLenSq > 1e-9 {
				tSeg = common.Clamp01(q.Sub(a).Dot(ab) / abLenSq)
			}
			pt := a.Add(ab.Scale(tSeg))
			dSq := q.Sub(pt).Dot(q.Sub(pt))
			if dSq < minDistSq {
				minDistSq = dSq
				bestIdx = i
				bestTSeg = tSeg
				bestClosest = pt
			}
		}

		radVec := q.Sub(bestClosest)
		rTube := math.Sqrt(minDistSq)
		sParam := ((float64(bestIdx) + bestTSeg) / float64(knotSamples)) * 2.0 * math.Pi

		t0 := knotTangents[bestIdx]
		t1 := knotTangents[(bestIdx+1)%knotSamples]
		tang := t0.Add(t1.Sub(t0).Scale(bestTSeg)).Normalize()
		b1 := tang.Cross(render.V3(0, 0, 1))
		if b1.Len() < 1e-4 {
			b1 = render.V3(1, 0, 0)
		} else {
			b1 = b1.Normalize()
		}
		b2 := tang.Cross(b1).Normalize()
		crossAngle := math.Atan2(radVec.Dot(b2), radVec.Dot(b1))

		// 8 twisting longitudinal flutes + clean architectural collar rings every 15 degrees
		fluteWave := math.Cos(crossAngle*8.0 + sParam*6.0)
		collarRing := math.Exp(-math.Pow(math.Sin(sParam*12.0)/0.14, 2.0))
		dKnot := (rTube - tubeR) - 0.0065*fluteWave - 0.0045*collarRing

		// 2. Non-Intersecting Outer Graduated Astrolabe Ring at R = 0.515 (> 0.467) in the same tilted plane
		dRing := math.Hypot(math.Hypot(pTilt.X, pTilt.Y)-0.515, pTilt.Z) - 0.016

		// 3. Non-Intersecting Central Octahedral Star-Jewel Core at r <= 0.092 (< 0.113)
		qCore := render.FoldOctahedral(pTilt.RotateY(coreRotY))
		dCore := render.SDF3DOctahedron(qCore, 0.090) - 0.005

		minD := dKnot
		objID := 10
		partID := 100 + (int(math.Floor(sParam/(2.0*math.Pi)*24.0)) % 8)
		loc := q
		ridgeFactor := 0.5 + 0.5*fluteWave
		if collarRing > 0.55 {
			ridgeFactor = 1.0
		}

		if dRing < minD {
			minD = dRing
			objID = 11
			angR := math.Atan2(pTilt.Y, pTilt.X) - phi/24.0
			partID = 200 + (int(math.Floor((angR+math.Pi)/(2.0*math.Pi)*24.0)) & 31)
			loc = pTilt
		}
		if dCore < minD {
			minD = dCore
			objID = 12
			partID = 300 + int(math.Floor((qCore.X+qCore.Y)*18.0))
			loc = qCore
		}
		return minD, objID, partID, loc, sParam, ridgeFactor
	}

	sceneSDF := func(p render.Vec3) float64 {
		d, _, _, _, _, _ := evalKnotScene(p)
		return d
	}

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)
	cfg := render.DefaultMoebiusConfig(seed)
	cfg.HatchSpacingPx = 3.5

	camPos := render.V3(0.0, 0.0, 1.54)

	render.RenderMoebiusScene(rc, cfg, func(ux, uy float64) render.MoebiusPixel {
		vx := (ux - 0.5) * 2.0
		vy := (0.5 - uy) * 2.0
		rScreen := math.Hypot(vx, vy)
		angScreen := math.Atan2(vy, vx)
		stoneRel := render.EvaluateMoebiusStoneRelief(ne, ux, uy, 7.5, 7.5, 0.30)

		// 1. Stationary Classical Stone Arch Spandrels & 4 Corner Astrolabe Rosettes (rScreen > 0.85)
		if rScreen > 0.85 {
			cornerX := math.Abs(vx) - 0.78
			cornerY := math.Abs(vy) - 0.78
			rCorner := math.Hypot(cornerX, cornerY)
			if rScreen > 0.89 && rCorner < 0.14 {
				aCorner := math.Atan2(cornerY, cornerX)
				spoke := int(math.Floor((aCorner + math.Pi) / (2.0 * math.Pi) * 16.0))
				ringIdx := int(math.Floor(rCorner * 22.0))
				sCol := mudGold
				if (spoke+ringIdx)&1 == 1 {
					sCol = warmParchment
				}
				return render.MoebiusPixel{
					Depth:           0.82,
					ObjID:           6,
					PartID:          60 + ringIdx*16 + (spoke & 15),
					DetailID:        spoke,
					Normal:          [3]float64{-cornerX * 3.5, -cornerY * 3.5, 0.85},
					FlowTangent:     [2]float64{-math.Sin(aCorner), math.Cos(aCorner)},
					Illum:           0.74,
					Spec:            0.48,
					SunColor:        sCol,
					MidColor:        amberBronze,
					ShadeColor:      umberShade,
					DeepShadowColor: deepIntaglio,
					CrackInk:        stoneRel.CrackInk * 0.35,
				}
			}

			nStone := render.V3(-vx*0.45+stoneRel.NormalMod[0]*0.25, -vy*0.45+stoneRel.NormalMod[1]*0.25, 0.82).Normalize()
			illum := common.Clamp01((0.36 + 0.62*math.Max(0.0, nStone.Dot(lightDir3D))) * (0.86 + 0.14*stoneRel.ColorMod))
			sunCol := carvedStone
			partID := 40 + (stoneRel.BlockID & 15)
			if rScreen < 0.89 {
				// Graduated 24k Gold & Ivory Oculus Bezel Ring
				degSec := int(math.Floor((angScreen + math.Pi) / (2.0 * math.Pi) * 36.0))
				sunCol = mudGold
				if degSec&1 == 1 {
					sunCol = warmIvory
				}
				illum = 0.76
				partID = 50 + (degSec & 31)
			}
			return render.MoebiusPixel{
				Depth:           0.85,
				ObjID:           4,
				PartID:          partID,
				DetailID:        stoneRel.BlockID,
				Normal:          [3]float64{nStone.X, nStone.Y, nStone.Z},
				FlowTangent:     [2]float64{-vy, vx},
				Illum:           illum,
				Bounce:          0.28,
				SunColor:        sunCol,
				MidColor:        amberBronze,
				ShadeColor:      umberShade,
				DeepShadowColor: deepIntaglio,
				CrackInk:        stoneRel.CrackInk * 0.52,
				TickMask:        stoneRel.PockInk * 0.40,
				Stipple:         0.16 * (1.0 - illum),
			}
		}

		rd := render.V3(vx*0.56, vy*0.56, -1.0).Normalize()
		hit := render.RaymarchSDF3DBounded(sceneSDF, camPos, rd, lightDir3D, render.V3(0, 0, 0), 0.66, 3.0, 76)

		if hit.Hit {
			_, objID, partID, loc, sParam, ridgeFactor := evalKnotScene(hit.Pos)
			ndotl := math.Max(0.0, hit.Normal.Dot(lightDir3D))
			illum := common.Clamp01((0.22 + 0.78*ndotl*(0.42+0.58*hit.Shadow)) * hit.AO)

			// Evaluate Phase 5 Intrinsic 3D Weingarten Shape Operator & Principal-Curvature Engraving!
			inkW, chalkW, pc := render.Evaluate3DWeingartenEngraving(sceneSDF, loc, illum, 64.0)

			if objID == 10 {
				// Clean, noble Bronze/Gold/Malachite/Azurite shading driven by smooth flute ridgeFactor & Weingarten curvature:
				//  - Raised flute crests & collar rings (ridgeFactor -> 1): Burnished 24k Mud-Gold & Warm Ivory
				//  - Concave flute channels (ridgeFactor -> 0): Patinated Song Malachite & Deep Azurite
				var sunCol, midCol [3]float64
				for cIdx := 0; cIdx < 3; cIdx++ {
					sunCol[cIdx] = warmIvory[cIdx]*(1.0-ridgeFactor) + mudGold[cIdx]*ridgeFactor
					midCol[cIdx] = malachitePatina[cIdx]*(1.0-ridgeFactor*0.75) + mudGold[cIdx]*(ridgeFactor*0.75)
				}
				return render.MoebiusPixel{
					Depth:           hit.T,
					ObjID:           objID,
					PartID:          partID,
					DetailID:        int(math.Floor(sParam * 12.0)),
					Normal:          [3]float64{hit.Normal.X, hit.Normal.Y, hit.Normal.Z},
					FlowTangent:     [2]float64{pc.E2.X, pc.E2.Y},
					Illum:           common.Clamp01(illum + 0.22*chalkW + 0.12*ridgeFactor),
					Spec:            ridgeFactor*0.52 + chalkW*0.38,
					Bounce:          0.36,
					SunColor:        sunCol,
					MidColor:        midCol,
					ShadeColor:      azuriteShadow,
					DeepShadowColor: deepIntaglio,
					CrackInk:        inkW * 0.52,
					Stipple:         0.14 * (1.0 - illum),
				}
			}

			if objID == 12 {
				// Central Octahedral Star-Jewel Core (Glowing Cinnabar-Vermilion & 24k Gold)
				return render.MoebiusPixel{
					Depth:           hit.T,
					ObjID:           objID,
					PartID:          partID,
					DetailID:        partID,
					Normal:          [3]float64{hit.Normal.X, hit.Normal.Y, hit.Normal.Z},
					FlowTangent:     [2]float64{pc.E2.X, pc.E2.Y},
					Illum:           common.Clamp01(illum + 0.28),
					Spec:            0.65,
					Bounce:          0.40,
					SunColor:        mudGold,
					MidColor:        cinnabarCore,
					ShadeColor:      azuriteShadow,
					DeepShadowColor: deepIntaglio,
					CrackInk:        inkW * 0.48,
				}
			}

			// Outer Non-Intersecting Gilded Astrolabe Ring
			return render.MoebiusPixel{
				Depth:           hit.T,
				ObjID:           objID,
				PartID:          partID,
				DetailID:        partID,
				Normal:          [3]float64{hit.Normal.X, hit.Normal.Y, hit.Normal.Z},
				FlowTangent:     [2]float64{pc.E2.X, pc.E2.Y},
				Illum:           common.Clamp01(illum + 0.25*chalkW),
				Spec:            0.55 + 0.35*chalkW,
				Bounce:          0.35,
				SunColor:        mudGold,
				MidColor:        cinnabarCore,
				ShadeColor:      azuriteShadow,
				DeepShadowColor: deepIntaglio,
				CrackInk:        inkW * 0.52,
				Stipple:         0.12 * (1.0 - illum),
			}
		}

		// 2. Crisp 12-Fold Harmonic Compass Rose & Warm Antique Parchment Watercolor Scroll Backdrop (rScreen <= 0.85)
		rose12 := 0.5 + 0.5*math.Cos(12.0*angScreen)
		rose6 := 0.5 + 0.5*math.Cos(6.0*angScreen+math.Pi/6.0)
		petalR := 0.52 + 0.14*rose12 + 0.08*rose6
		sec24 := int(math.Floor((angScreen + math.Pi) / (2.0 * math.Pi) * 24.0))
		ringBand := int(math.Floor(rScreen * 9.0))

		// Subtle Kubelka-Munk mineral watercolor wash on warm antique parchment
		parchmentWash := render.EvaluateWatercolorWash(ne, ux, uy, rScreen-petalR, render.PigmentYellowOchre, warmParchment, washCfg)
		celadonWash := render.EvaluateWatercolorWash(ne, ux, uy, math.Abs(rScreen-0.64)-0.08, render.PigmentMalachite, parchmentWash, washCfg)

		sCol := parchmentWash
		mCol := amberBronze
		lit := common.Clamp01(0.76 - 0.18*rScreen + 0.08*(1.0-uy))
		part := 10 + ringBand*4 + (sec24 & 3)

		if rScreen > petalR {
			// Outer Compass Rose Ring inside Bezel: Warm Parchment, Burnished Gold & Subtle Celadon-Malachite Sectors
			sCol = celadonWash
			if (sec24 & 1) == 0 {
				for cIdx := 0; cIdx < 3; cIdx++ {
					sCol[cIdx] = parchmentWash[cIdx]*0.65 + mudGold[cIdx]*0.35
				}
			}
			lit = common.Clamp01(0.68 - 0.12*(rScreen-petalR) + 0.08*stoneRel.ColorMod)
			part = 25 + (sec24 & 15)
		}

		return render.MoebiusPixel{
			Depth:           95.0,
			ObjID:           1,
			PartID:          part,
			DetailID:        stoneRel.BlockID,
			Normal:          [3]float64{-vx * 0.25, -vy * 0.25, 0.94},
			FlowTangent:     [2]float64{-math.Sin(angScreen), math.Cos(angScreen)},
			Illum:           lit,
			Spec:            common.Smoothstep(0.76, 0.95, lit) * 0.38,
			SunColor:        sCol,
			MidColor:        mCol,
			ShadeColor:      umberShade,
			DeepShadowColor: deepIntaglio,
			CrackInk:        stoneRel.CrackInk * 0.36,
			TickMask:        stoneRel.PockInk * 0.30,
			Stipple:         0.16 * (1.0 - lit),
		}
	})
	c.MergeRenderCanvas(rc)
}
