package arts

import (
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type durerPolyhedron struct {
	symmetry int
	rings    int
	twist    float64
}

// NewDurerPolyhedron returns a DurerPolyhedron generator inspired by Albrecht
// Dürer's Melencolia I and Wenzel Jamnitzer's Perspectiva Corporum Regularium
// (1568): a 3D SDF raymarched icosahedral/dodecahedral Renaissance polyhedron &
// gimbaled astronomical armillary monument on a carved stone plinth, shaded with
// Kubelka-Munk gold-leaf/raw-umber/azurite pigments and object-locked 3D
// copperplate burin cross-hatching.
func NewDurerPolyhedron(symmetry, rings int, twist float64) *durerPolyhedron {
	if symmetry <= 0 {
		symmetry = 12
	}
	if rings <= 0 {
		rings = 3
	}
	if twist == 0 {
		twist = 0.32
	}
	return &durerPolyhedron{
		symmetry: symmetry,
		rings:    rings,
		twist:    twist,
	}
}

// Generative renders the static DurerPolyhedron copperplate & chiaroscuro print.
func (dp *durerPolyhedron) Generative(c *generativeart.Canva) {
	dp.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 3D rotating DurerPolyhedron animation.
func (dp *durerPolyhedron) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	seed := c.Seed()
	if seed == 0 {
		seed = 20262203
	}
	ne := common.NewNoiseEngine(seed)
	phi := 2.0 * math.Pi * phase

	// Kubelka-Munk Renaissance Chiaroscuro Woodcut/Copperplate & Gold-Leaf Palette
	giltHighlight := render.MixKubelkaMunk(render.PW(render.PigmentGoldLeaf, 0.55), render.PW(render.PigmentTitaniumWhite, 0.45))
	goldBronze := render.MixKubelkaMunk(render.PW(render.PigmentGoldLeaf, 0.72), render.PW(render.PigmentBurntSienna, 0.28))
	warmParchment := render.MixKubelkaMunk(render.PW(render.PigmentYellowOchre, 0.42), render.PW(render.PigmentTitaniumWhite, 0.58))
	terracottaStone := render.MixKubelkaMunk(render.PW(render.PigmentBurntSienna, 0.58), render.PW(render.PigmentYellowOchre, 0.42))
	azuriteEnamel := render.MixKubelkaMunk(render.PW(render.PigmentAzurite, 0.72), render.PW(render.PigmentCobaltTeal, 0.28))
	umberShadow := render.MixKubelkaMunk(render.PW(render.PigmentRawUmber, 0.62), render.PW(render.PigmentUltramarine, 0.38))
	deepIntaglio := render.MixKubelkaMunk(render.PW(render.PigmentPineSootInk, 0.78), render.PW(render.PigmentUltramarine, 0.22))

	lightDir3D := render.V3(-0.56, 0.58, 0.59).Normalize()

	// Exact periodic 3D rotation angles for each sub-assembly:
	//  - Outer Hollow Icosahedral Cage: 5-fold symmetry axis aligned with Y via AlignIcosahedral5FoldY,
	//    so rotating around Y by (2*pi/5)*phase is 100% periodic!
	//  - Inner Azurite Star-Octahedron Core: 4-fold symmetry around Y, counter-rotating by -(2*pi/4)*phase!
	//  - Gimbaled Meridian Ring: symmetric under 180-deg rotation, tumbling by pi*phase!
	rotCageY := dp.twist + (2.0*math.Pi/5.0)*phase
	rotCoreY := -dp.twist - (2.0*math.Pi/4.0)*phase
	rotGimbalY := 0.42 + math.Pi*phase
	cageTiltX := 0.24

	// Sub-assembly distance evaluators returning (dist, subObjID, localPos)
	evalPlinth := func(p render.Vec3) (float64, render.Vec3) {
		pPlinth := p.Sub(render.V3(0.0, -0.56, 0.0))
		qPlinth, _ := render.FoldPolarModXZ(pPlinth, 8)
		dBaseShaft := render.SDF3DRoundedBox(qPlinth, render.V3(0.46, 0.16, 0.46), 0.02)
		dTopMolding := render.SDF3DRoundedBox(qPlinth.Sub(render.V3(0, 0.15, 0)), render.V3(0.52, 0.035, 0.52), 0.015)
		dBotMolding := render.SDF3DRoundedBox(qPlinth.Sub(render.V3(0, -0.15, 0)), render.V3(0.56, 0.040, 0.56), 0.015)
		return math.Min(dBaseShaft, math.Min(dTopMolding, dBotMolding)), pPlinth
	}

	evalOuterCage := func(p render.Vec3) (float64, render.Vec3) {
		// Tilt cage toward camera, then rotate around its exact 5-fold Y axis!
		q := p.Sub(render.V3(0.0, 0.08, 0.0)).RotateX(cageTiltX).RotateY(rotCageY)
		fIco := render.FoldIcosahedral(render.AlignIcosahedral5FoldY(q))

		// In FoldIcosahedral, the icosahedral/dodecahedral face normal is along +Y (fIco.Y)
		dOuterBox := render.SDF3DRoundedBox(fIco, render.V3(0.26, 0.40, 0.26), 0.018)
		dPolyShell := math.Max(dOuterBox, -render.SDF3DSphere(q, 0.33))

		// Carve circular Renaissance oculus windows through the 12 dodecahedral faces!
		dWindowCut := math.Hypot(fIco.X, fIco.Z) - 0.140
		dHollowCage := render.SDF3DSmoothSubtract(dWindowCut, dPolyShell, 0.015)
		return dHollowCage, fIco
	}

	evalInnerCore := func(p render.Vec3) (float64, render.Vec3) {
		q := p.Sub(render.V3(0.0, 0.08, 0.0)).RotateX(cageTiltX).RotateY(rotCoreY)
		fOct := render.FoldOctahedral(q)
		dStar := render.SDF3DOctahedron(fOct, 0.26)
		dSpikes := render.SDF3DCapsule(fOct, render.V3(0, 0, 0), render.V3(0.32, 0, 0), 0.028)
		return render.SDF3DSmoothUnion(dStar, dSpikes, 0.02), fOct
	}

	evalArmillary1 := func(p render.Vec3) (float64, render.Vec3) {
		q := p.Sub(render.V3(0.0, 0.08, 0.0)).RotateZ(-0.26).RotateX(0.32)
		return render.SDF3DTorus(q, 0.54, 0.025), q
	}

	evalArmillary2 := func(p render.Vec3) (float64, render.Vec3) {
		// Tumbling inner gimbal ring!
		q := p.Sub(render.V3(0.0, 0.08, 0.0)).RotateZ(-0.26).RotateY(rotGimbalY).RotateX(math.Pi * 0.5)
		return render.SDF3DTorus(q, 0.47, 0.021), q
	}

	monumentSDF := func(p render.Vec3) float64 {
		d0, _ := evalPlinth(p)
		d1, _ := evalOuterCage(p)
		d2, _ := evalInnerCore(p)
		d3, _ := evalArmillary1(p)
		d4, _ := evalArmillary2(p)
		return math.Min(d0, math.Min(math.Min(d1, d2), math.Min(d3, d4)))
	}

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)
	cfg := render.DefaultMoebiusConfig(seed)
	cfg.HatchSpacingPx = 3.6

	render.RenderMoebiusScene(rc, cfg, func(ux, uy float64) render.MoebiusPixel {
		x := ux*2.0 - 1.0
		y := uy*2.0 - 1.0

		// =====================================================================
		// 1. RAYMARCH 3D JAMNITZER POLYHEDRON, GIMBALED ARMILLARY & PLINTH
		// =====================================================================
		if math.Abs(x) <= 0.78 && y >= -0.68 && y <= 0.86 {
			ro := render.V3(x*0.88, -y*0.88, 2.2)
			rd := render.V3(0.0, 0.0, -1.0)
			hit := render.RaymarchSDF3D(monumentSDF, ro, rd, lightDir3D, 3.4, 64)
			if hit.Hit {
				// Identify which exact 3D sub-assembly was hit and recover its object-local coordinates!
				d0, loc0 := evalPlinth(hit.Pos)
				d1, loc1 := evalOuterCage(hit.Pos)
				d2, loc2 := evalInnerCore(hit.Pos)
				d3, loc3 := evalArmillary1(hit.Pos)
				d4, loc4 := evalArmillary2(hit.Pos)

				minD := d0
				obj := 20
				pLoc := loc0
				nLoc := hit.Normal

				if d1 < minD {
					minD = d1
					obj = 30
					pLoc = loc1
					nLoc = hit.Normal.RotateX(cageTiltX).RotateY(rotCageY)
				}
				if d2 < minD {
					minD = d2
					obj = 32
					pLoc = loc2
					nLoc = hit.Normal.RotateX(cageTiltX).RotateY(rotCoreY)
				}
				if d3 < minD {
					minD = d3
					obj = 31
					pLoc = loc3
					nLoc = hit.Normal.RotateZ(-0.26).RotateX(0.32)
				}
				if d4 < minD {
					minD = d4
					obj = 33
					pLoc = loc4
					nLoc = hit.Normal.RotateZ(-0.26).RotateY(rotGimbalY).RotateX(math.Pi * 0.5)
				}

				// Evaluate 3D object-locked Voronoi stone/metal relief (ZERO 2D shower-door swimming!)
				stone := render.EvaluateMoebiusStoneRelief3D(ne, pLoc.Scale(2.0), nLoc, 7.0, 0.28)
				nScreen := moebiusNorm3(
					hit.Normal.X+0.08*stone.NormalMod[0],
					-hit.Normal.Y+0.08*stone.NormalMod[1],
					hit.Normal.Z,
				)
				lit := common.Clamp01((0.14 + 0.86*math.Max(0.0, hit.Normal.Dot(lightDir3D))) *
					(0.42 + 0.58*hit.Shadow) * hit.AO)

				// Evaluate object-locked 3D copperplate burin cross-hatching & chiaroscuro chalk!
				burinInk, burinChalk := render.Evaluate3DBurinEngraving(pLoc, nLoc, lit, 82.0)

				var sCol, mCol [3]float64
				var part int
				switch obj {
				case 20:
					// Octagonal Stone Plinth
					sCol = warmParchment
					mCol = terracottaStone
					part = 200 + int(math.Floor((pLoc.Y+0.3)*12.0))*2 + int(moebiusBoolFloat(pLoc.X > 0))
				case 32:
					// Inner Counter-Rotating Azurite Star-Octahedron Core
					sCol = azuriteEnamel
					mCol = goldBronze
					part = 320 + int(math.Floor(pLoc.X*10.0))
				case 31:
					// Outer Equatorial Armillary Ring (degree ticks advance by 1 tick per cycle!)
					ang := math.Atan2(pLoc.Z, pLoc.X) + phi/24.0
					seg := int(math.Floor((ang + math.Pi) / (2.0 * math.Pi) * 24.0))
					sCol = giltHighlight
					mCol = goldBronze
					part = 310 + (seg & 15)
				case 33:
					// Tumbling Inner Meridian Gimbal Ring
					ang := math.Atan2(pLoc.Z, pLoc.X)
					seg := int(math.Floor((ang + math.Pi) / (2.0 * math.Pi) * 20.0))
					sCol = moebiusLerp3(giltHighlight, azuriteEnamel, float64(seg&1)*0.45)
					mCol = goldBronze
					part = 330 + (seg & 15)
				default:
					// Outer 5-Fold Rotating Hollow Icosahedral-Dodecahedral Cage
					sCol = moebiusLerp3(warmParchment, giltHighlight, 0.58)
					mCol = goldBronze
					part = 300 + int(math.Floor(pLoc.Y*8.0))*4 + int(math.Floor(pLoc.Z*8.0))
				}

				return render.MoebiusPixel{
					Depth:           hit.T,
					ObjID:           obj,
					PartID:          part,
					DetailID:        stone.BlockID,
					Normal:          nScreen,
					FlowTangent:     [2]float64{-nScreen[1], nScreen[0]},
					Illum:           lit,
					Spec:            common.Clamp01(0.55*math.Max(0.0, hit.Curvature) + burinChalk + common.Smoothstep(0.76, 0.98, lit)*0.65),
					Bounce:          0.42,
					SunColor:        sCol,
					MidColor:        mCol,
					ShadeColor:      umberShadow,
					DeepShadowColor: deepIntaglio,
					CrackInk:        math.Max(stone.CrackInk*0.48, common.Smoothstep(0.22, 0.65, -hit.Curvature)*0.82),
					TickMask:        math.Max(stone.PockInk*0.45, burinInk*0.85),
				}
			}
		}

		// =====================================================================
		// 2. CLASSICAL RENAISSANCE COFFERED ARCHED NICHE & PERSPECTIVE FLOOR
		// =====================================================================
		if y >= 0.42 {
			ty := (y - 0.42) / 0.58
			px := x / (0.38 + 0.62*ty)
			pz := 1.6 / (0.28 + 0.72*ty)
			tileP := (int(math.Floor(px*3.0)) + int(math.Floor(pz*2.5))) & 1
			stone := render.EvaluateMoebiusStoneRelief(ne, px, pz, 4.0, 4.0, 0.25)

			lit := 0.78 - 0.35*(1.0-ty)
			if x > 0.05 && x < 0.68 && y < 0.78 {
				lit *= 0.42
			}
			sCol := warmParchment
			if tileP == 1 {
				sCol = terracottaStone
			}
			return render.MoebiusPixel{
				Depth:           6.0 - 2.0*ty,
				ObjID:           10,
				PartID:          100 + int(math.Floor(px*3.0))*8 + int(math.Floor(pz*2.5)),
				DetailID:        stone.BlockID,
				Normal:          [3]float64{0, -0.8, 0.6},
				FlowTangent:     [2]float64{1, 0},
				Illum:           lit,
				SunColor:        sCol,
				MidColor:        terracottaStone,
				ShadeColor:      umberShadow,
				DeepShadowColor: deepIntaglio,
				CrackInk:        stone.CrackInk * 0.5,
			}
		}

		rRose := math.Hypot(x, y-(-0.08))
		angRose := math.Atan2(y-(-0.08), x)
		inArchWall := math.Abs(x) > 0.76 || (y < -0.10 && rRose > 0.76)

		if inArchWall {
			stone := render.EvaluateMoebiusStoneRelief(ne, x, y, 7.0, 8.0, 0.32)
			lit := common.Clamp01(0.58 - 0.32*x)
			return render.MoebiusPixel{
				Depth:           5.2,
				ObjID:           12,
				PartID:          120 + int(math.Floor((angRose+math.Pi)*5.0)),
				DetailID:        stone.BlockID,
				Normal:          moebiusNorm3(-x*0.5, -0.2, 0.8),
				FlowTangent:     [2]float64{1, 0},
				Illum:           lit,
				SunColor:        warmParchment,
				MidColor:        terracottaStone,
				ShadeColor:      umberShadow,
				DeepShadowColor: deepIntaglio,
				CrackInk:        stone.CrackInk * 0.65,
			}
		}

		// Engraved celestial astrolabe rays inside the recessed niche dome (slowly rotating by 1 ray per cycle!)
		rayWave := 0.5 + 0.5*math.Cos(24.0*angRose-phi)
		ringWave := 0.5 + 0.5*math.Cos(28.0*rRose)
		nicheLit := common.Clamp01(0.32 + 0.42*(1.0-rRose/0.78) + 0.18*common.Smoothstep(-0.5, 0.5, x))
		return render.MoebiusPixel{
			Depth:           8.5,
			ObjID:           2,
			PartID:          20 + int(rRose*8.0),
			DetailID:        int((angRose + math.Pi) * 12.0),
			Normal:          moebiusNorm3(-x*0.6, -(y+0.08)*0.6, 0.7),
			FlowTangent:     [2]float64{-math.Sin(angRose), math.Cos(angRose)},
			Illum:           nicheLit,
			SunColor:        azuriteEnamel,
			MidColor:        umberShadow,
			ShadeColor:      umberShadow,
			DeepShadowColor: deepIntaglio,
			TickMask:        common.Smoothstep(0.88, 0.96, math.Max(rayWave, ringWave)) * 0.55,
			Stipple:         (1.0 - nicheLit) * 0.45,
		}
	})

	c.MergeRenderCanvas(rc)
}
