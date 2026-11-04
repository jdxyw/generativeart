package arts

import (
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type keplerHarmonices struct {
	shells      int
	rings       int
	inclination float64
}

// NewKeplerHarmonices returns a KeplerHarmonices generator inspired by Johannes
// Kepler's Mysterium Cosmographicum (1596, Tabula III) and Harmonices Mundi (1619):
// a 3D SDF raymarched astronomical monument nesting the Platonic solids (Hollow
// Dodecahedron, Icosahedron, Cube/Octahedron, and Star-Tetrahedron) inside
// zodiacal armillary rings around a central solar sphere, shaded with Kubelka-Munk
// mineral pigments and object-locked 3D copperplate burin cross-hatching.
func NewKeplerHarmonices(shells, rings int, inclination float64) *keplerHarmonices {
	if shells <= 0 {
		shells = 5
	}
	if rings <= 0 {
		rings = 3
	}
	if inclination == 0 {
		inclination = 0.26
	}
	return &keplerHarmonices{
		shells:      shells,
		rings:       rings,
		inclination: inclination,
	}
}

// Generative renders the static KeplerHarmonices copperplate & mineral-glaze print.
func (kh *keplerHarmonices) Generative(c *generativeart.Canva) {
	kh.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360° looping KeplerHarmonices animation.
func (kh *keplerHarmonices) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	seed := c.Seed()
	if seed == 0 {
		seed = 20262302
	}
	ne := common.NewNoiseEngine(seed)
	phi := 2.0 * math.Pi * phase

	// Kubelka-Munk Renaissance Astronomical Brass, Lapis, Malachite & Cinnabar Palette
	solarGold := render.MixKubelkaMunk(render.PW(render.PigmentGoldLeaf, 0.65), render.PW(render.PigmentTitaniumWhite, 0.35))
	burnishedBrass := render.MixKubelkaMunk(render.PW(render.PigmentGoldLeaf, 0.72), render.PW(render.PigmentBurntSienna, 0.28))
	ivoryParchment := render.MixKubelkaMunk(render.PW(render.PigmentYellowOchre, 0.38), render.PW(render.PigmentTitaniumWhite, 0.62))
	cinnabarCoral := render.MixKubelkaMunk(render.PW(render.PigmentVermilion, 0.68), render.PW(render.PigmentYellowOchre, 0.32))
	azuriteLapis := render.MixKubelkaMunk(render.PW(render.PigmentAzurite, 0.74), render.PW(render.PigmentUltramarine, 0.26))
	malachiteJade := render.MixKubelkaMunk(render.PW(render.PigmentMalachite, 0.68), render.PW(render.PigmentCobaltTeal, 0.32))
	umberShadow := render.MixKubelkaMunk(render.PW(render.PigmentRawUmber, 0.60), render.PW(render.PigmentUltramarine, 0.40))
	deepIntaglio := render.MixKubelkaMunk(render.PW(render.PigmentPineSootInk, 0.78), render.PW(render.PigmentUltramarine, 0.22))

	lightDir3D := render.V3(-0.55, 0.58, 0.60).Normalize()

	// Exact symmetry-matched rotation speeds so phase=1 matches phase=0 with zero seam:
	//  - 4-fold Cube/Octahedron outer cage: rotates by +(2*pi/4)*phase (90 deg per cycle)
	//  - 5-fold Dodecahedron/Icosahedron middle cage: counter-rotates by -(2*pi/5)*phase (72 deg per cycle)
	//  - 4-fold Inner Stella Octangula + Solar Core: rotates by +(2*pi/4)*phase
	//  - Zodiacal Ecliptic Gimbal Ring: tumbles by +pi*phase (180 deg per cycle)
	rot4FoldPos := 0.34 + (2.0*math.Pi/4.0)*phase
	rot5FoldNeg := -0.28 - (2.0*math.Pi/5.0)*phase
	rotGimbal := 0.48 + math.Pi*phase
	tiltX := kh.inclination

	centerY := 0.08

	// 1. Renaissance Astronomical Pedestal & Atlas Cradle at bottom (y < -0.44)
	evalPedestal := func(p render.Vec3) (float64, render.Vec3) {
		pBase := p.Sub(render.V3(0.0, -0.60, 0.0))
		qHex, _ := render.FoldPolarModXZ(pBase, 12)
		dPlinth := render.SDF3DRoundedBox(qHex, render.V3(0.46, 0.12, 0.46), 0.018)
		dCap := render.SDF3DCylinderY(pBase.Sub(render.V3(0, 0.12, 0)), 0.40, 0.032)
		dPillar := render.SDF3DFlutedColumnY(pBase.Sub(render.V3(0, 0.20, 0)), 0.14, 0.08, 12)
		// Cradle arc holding the outer armillary sphere
		qCradle := p.Sub(render.V3(0, centerY, 0))
		dCradle := math.Max(render.SDF3DTorus(qCradle.RotateX(math.Pi*0.5), 0.62, 0.024), p.Y-(-0.12))
		return math.Min(math.Min(dPlinth, dCap), math.Min(dPillar, dCradle)), pBase
	}

	// 2. Outer Zodiacal Ecliptic Armillary Band & Tumbling Meridian Ring
	evalZodiacRings := func(p render.Vec3) (float64, int, render.Vec3) {
		qEcl := p.Sub(render.V3(0, centerY, 0)).RotateZ(-0.24).RotateX(0.30)
		// Wide flat zodiacal belt (cylindrical band)
		dBeltOuter := render.SDF3DCylinderY(qEcl, 0.59, 0.042)
		dBeltInner := render.SDF3DCylinderY(qEcl, 0.555, 0.060)
		dZodiacBelt := math.Max(dBeltOuter, -dBeltInner)

		qMer := p.Sub(render.V3(0, centerY, 0)).RotateZ(-0.24).RotateY(rotGimbal).RotateX(math.Pi * 0.5)
		dMeridian := render.SDF3DTorus(qMer, 0.535, 0.019)

		if dZodiacBelt < dMeridian {
			return dZodiacBelt, 31, qEcl
		}
		return dMeridian, 32, qMer
	}

	// 3. Outer Platonic Shell (Hollow Dodecahedron Cage — 5-fold Y axis, radius 0.42)
	evalOuterDodecaShell := func(p render.Vec3) (float64, render.Vec3) {
		q := p.Sub(render.V3(0, centerY, 0)).RotateX(tiltX).RotateY(rot5FoldNeg)
		d := render.SDF3DPlatonicCage(q, 0, 0.41, 0.022)
		return d, render.FoldIcosahedral(render.AlignIcosahedral5FoldY(q))
	}

	// 4. Middle Platonic Shell (Hollow Cube / Hexahedron Cage — 4-fold Y axis, radius 0.25)
	evalMiddleCubeShell := func(p render.Vec3) (float64, render.Vec3) {
		q := p.Sub(render.V3(0, centerY, 0)).RotateX(tiltX).RotateY(rot4FoldPos)
		d := render.SDF3DPlatonicCage(q, 3, 0.235, 0.020)
		return d, render.FoldOctahedral(q)
	}

	// 5. Inner Platonic Shell (Hollow Octahedron & Blazing Solar Core — 4-fold Y axis)
	evalInnerOctaAndSun := func(p render.Vec3) (float64, int, render.Vec3) {
		q := p.Sub(render.V3(0, centerY, 0)).RotateX(tiltX).RotateY(-rot4FoldPos)
		dOcta := render.SDF3DPlatonicCage(q, 2, 0.19, 0.018)
		dSun := render.SDF3DSphere(q, 0.088)
		fOct := render.FoldOctahedral(q)
		if dSun < dOcta {
			return dSun, 36, fOct
		}
		return dOcta, 35, fOct
	}

	keplerSDF := func(p render.Vec3) float64 {
		d0, _ := evalPedestal(p)
		d1, _, _ := evalZodiacRings(p)
		d2, _ := evalOuterDodecaShell(p)
		d3, _ := evalMiddleCubeShell(p)
		d4, _, _ := evalInnerOctaAndSun(p)
		return math.Min(d0, math.Min(math.Min(d1, d2), math.Min(d3, d4)))
	}

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)
	cfg := render.DefaultMoebiusConfig(seed)
	cfg.HatchSpacingPx = 3.5

	render.RenderMoebiusScene(rc, cfg, func(ux, uy float64) render.MoebiusPixel {
		x := ux*2.0 - 1.0
		y := uy*2.0 - 1.0

		// =====================================================================
		// 1. RAYMARCH 3D KEPLER MYSTERIUM COSMOGRAPHICUM NESTED SOLIDS
		// =====================================================================
		if math.Abs(x) <= 0.80 && y >= -0.72 && y <= 0.88 {
			ro := render.V3(x*0.88, -y*0.88, 2.2)
			rd := render.V3(0, 0, -1)
			hit := render.RaymarchSDF3D(keplerSDF, ro, rd, lightDir3D, 3.5, 68)
			if hit.Hit {
				d0, loc0 := evalPedestal(hit.Pos)
				d1, id1, loc1 := evalZodiacRings(hit.Pos)
				d2, loc2 := evalOuterDodecaShell(hit.Pos)
				d3, loc3 := evalMiddleCubeShell(hit.Pos)
				d4, id4, loc4 := evalInnerOctaAndSun(hit.Pos)

				minD := d0
				obj := 20
				pLoc := loc0
				nLoc := hit.Normal

				if d1 < minD {
					minD = d1
					obj = id1
					pLoc = loc1
					nLoc = hit.Normal.RotateZ(-0.24).RotateY(rotGimbal)
				}
				if d2 < minD {
					minD = d2
					obj = 33
					pLoc = loc2
					nLoc = hit.Normal.RotateX(tiltX).RotateY(rot5FoldNeg)
				}
				if d3 < minD {
					minD = d3
					obj = 34
					pLoc = loc3
					nLoc = hit.Normal.RotateX(tiltX).RotateY(rot4FoldPos)
				}
				if d4 < minD {
					minD = d4
					obj = id4
					pLoc = loc4
					nLoc = hit.Normal.RotateX(tiltX).RotateY(-rot4FoldPos)
				}

				stone := render.EvaluateMoebiusStoneRelief3D(ne, pLoc.Scale(2.4), nLoc, 7.2, 0.26)
				nScreen := moebiusNorm3(
					hit.Normal.X+0.07*stone.NormalMod[0],
					-hit.Normal.Y+0.07*stone.NormalMod[1],
					hit.Normal.Z,
				)
				lit := common.Clamp01((0.15 + 0.85*math.Max(0.0, hit.Normal.Dot(lightDir3D))) *
					(0.44 + 0.56*hit.Shadow) * hit.AO)
				burinInk, burinChalk := render.Evaluate3DBurinEngraving(pLoc, nLoc, lit, 84.0)

				var sCol, mCol [3]float64
				var part int
				switch obj {
				case 20:
					// Carved Renaissance Pedestal & Cradle
					sCol = ivoryParchment
					mCol = burnishedBrass
					part = 200 + int(math.Floor((pLoc.Y+0.4)*10.0))
				case 31:
					// Zodiacal Ecliptic Belt with 12 Constellation Divisions (advancing 1 sign per cycle!)
					ang := math.Atan2(pLoc.Z, pLoc.X) + phi/12.0
					signIdx := int(math.Floor((ang + math.Pi) / (2.0 * math.Pi) * 24.0))
					sCol = solarGold
					if signIdx&1 == 1 {
						sCol = ivoryParchment
					}
					mCol = burnishedBrass
					part = 310 + (signIdx & 15)
				case 32:
					// Tumbling Meridian Gimbal Ring
					ang := math.Atan2(pLoc.Z, pLoc.X)
					seg := int(math.Floor((ang + math.Pi) / (2.0 * math.Pi) * 20.0))
					sCol = moebiusLerp3(solarGold, azuriteLapis, float64(seg&1)*0.50)
					mCol = burnishedBrass
					part = 320 + (seg & 15)
				case 33:
					// Outer Hollow Cube / Hexahedron Cage (Warm Ivory & Terracotta-Brass)
					sCol = ivoryParchment
					mCol = cinnabarCoral
					part = 330 + int(math.Floor(pLoc.X*8.0))
				case 34:
					// Middle Counter-Rotating Hollow Dodecahedron Cage (Malachite-Jade & Gold)
					sCol = moebiusLerp3(malachiteJade, solarGold, 0.38)
					mCol = malachiteJade
					part = 340 + int(math.Floor(pLoc.Y*8.0))
				case 35:
					// Inner Hollow Octahedron Cage (Royal Azurite Lapis)
					sCol = azuriteLapis
					mCol = burnishedBrass
					part = 350 + int(math.Floor(pLoc.Z*10.0))
				default:
					// Central Blazing Solar Sphere (Sol)
					sCol = solarGold
					mCol = cinnabarCoral
					lit = math.Max(lit, 0.88)
					part = 360
				}

				return render.MoebiusPixel{
					Depth:           hit.T,
					ObjID:           obj,
					PartID:          part,
					DetailID:        stone.BlockID,
					Normal:          nScreen,
					FlowTangent:     [2]float64{-nScreen[1], nScreen[0]},
					Illum:           lit,
					Spec:            common.Clamp01(0.58*math.Max(0.0, hit.Curvature) + burinChalk + common.Smoothstep(0.75, 0.98, lit)*0.65),
					Bounce:          0.45,
					SunColor:        sCol,
					MidColor:        mCol,
					ShadeColor:      umberShadow,
					DeepShadowColor: deepIntaglio,
					CrackInk:        math.Max(stone.CrackInk*0.44, common.Smoothstep(0.22, 0.65, -hit.Curvature)*0.80),
					TickMask:        math.Max(stone.PockInk*0.40, burinInk*0.85),
				}
			}
		}

		// =====================================================================
		// 2. KEPLERIAN HELIOCENTRIC ORBITAL CHART & RENAISSANCE FLOOR
		// =====================================================================
		if y >= 0.44 {
			ty := (y - 0.44) / 0.56
			px := x / (0.36 + 0.64*ty)
			pz := 1.6 / (0.28 + 0.72*ty)
			rFloor := math.Hypot(px, pz-2.2)
			ringTile := (int(math.Floor(rFloor*3.5)) + int(math.Floor(math.Atan2(pz-2.2, px)*6.0/math.Pi))) & 1
			stone := render.EvaluateMoebiusStoneRelief(ne, px, pz, 4.0, 4.0, 0.24)

			lit := 0.76 - 0.32*(1.0-ty)
			if x > 0.04 && x < 0.68 && y < 0.80 {
				lit *= 0.42
			}
			sCol := ivoryParchment
			if ringTile == 1 {
				sCol = moebiusLerp3(ivoryParchment, burnishedBrass, 0.42)
			}
			return render.MoebiusPixel{
				Depth:           5.0 - 2.4*ty,
				ObjID:           10,
				PartID:          100 + ringTile,
				DetailID:        stone.BlockID,
				Normal:          [3]float64{0, -0.82, 0.57},
				FlowTangent:     [2]float64{1, 0},
				Illum:           lit,
				SunColor:        sCol,
				MidColor:        burnishedBrass,
				ShadeColor:      umberShadow,
				DeepShadowColor: deepIntaglio,
				CrackInk:        stone.CrackInk * 0.52,
			}
		}

		// =====================================================================
		// 3. CELESTIAL KEPLERIAN HARMONIC DIAGRAM BACKDROP
		// =====================================================================
		rChart := math.Hypot(x, y+0.08)
		angChart := math.Atan2(y+0.08, x)
		inCornerSpandrel := y < -0.15 && rChart > 0.86

		stone := render.EvaluateMoebiusStoneRelief(ne, x, y, 5.5, 5.5, 0.26)
		lit := common.Clamp01(0.70 - 0.42*rChart + 0.14*(1.0-uy))
		part := 50
		if inCornerSpandrel {
			lit *= 0.46
			part = 55
		} else {
			// Subtle engraved planetary orbit circles on the parchment chart
			orbitWave := math.Abs(math.Sin(rChart * 18.0))
			if orbitWave > 0.96 {
				part = 52 + int(math.Floor(rChart*6.0))
			}
			// Radial zodiac sector spokes every 30 degrees
			spokeMod := math.Abs(math.Mod(angChart+2.0*math.Pi, math.Pi/6.0) - math.Pi/12.0)
			if rChart > 0.60 && rChart < 0.84 && spokeMod < 0.012 {
				part = 58
			}
		}

		return render.MoebiusPixel{
			Depth:           12.0,
			ObjID:           5,
			PartID:          part,
			DetailID:        stone.BlockID,
			Normal:          [3]float64{-x * 0.35, -y * 0.35, 0.88},
			FlowTangent:     [2]float64{-math.Sin(angChart), math.Cos(angChart)},
			Illum:           lit,
			SunColor:        ivoryParchment,
			MidColor:        burnishedBrass,
			ShadeColor:      umberShadow,
			DeepShadowColor: deepIntaglio,
			CrackInk:        stone.CrackInk * 0.42,
			TickMask:        stone.PockInk * 0.38,
			Stipple:         (1.0 - lit) * 0.36,
		}
	})

	c.MergeRenderCanvas(rc)
}
