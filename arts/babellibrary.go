package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type babelLibrary struct {
	galleries  int
	tiers      int
	shaftScale float64
}

// NewBabelLibrary returns a BabelLibrary generator inspired by Jorge Luis Borges's
// The Library of Babel, Giovanni Battista Piranesi's Carceri, and M.C. Escher's
// hexagonal galleries: a 3D perspective & SDF raymarched infinite hexagonal
// chasm of arched book-alcove galleries, carved balustrades, spiral staircases,
// and suspended spherical brass lamps, shaded with Kubelka-Munk ochre/sienna/umber
// pigments and object-locked 3D copperplate burin cross-hatching.
func NewBabelLibrary(galleries, tiers int, shaftScale float64) *babelLibrary {
	if galleries <= 0 {
		galleries = 6
	}
	if tiers <= 0 {
		tiers = 5
	}
	if shaftScale <= 0 {
		shaftScale = 1.0
	}
	return &babelLibrary{
		galleries:  galleries,
		tiers:      tiers,
		shaftScale: shaftScale,
	}
}

// Generative renders the static BabelLibrary composition.
func (bl *babelLibrary) Generative(c *generativeart.Canva) {
	bl.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360° looping BabelLibrary animation.
func (bl *babelLibrary) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	seed := c.Seed()
	if seed == 0 {
		seed = 20262301
	}
	ne := common.NewNoiseEngine(seed)
	_ = 2.0 * math.Pi * phase

	// Kubelka-Munk Physical Mineral & Archival Library Palette
	lampGold := render.MixKubelkaMunk(render.PW(render.PigmentGoldLeaf, 0.72), render.PW(render.PigmentTitaniumWhite, 0.28))
	sunLimestone := render.MixKubelkaMunk(render.PW(render.PigmentYellowOchre, 0.50), render.PW(render.PigmentTitaniumWhite, 0.50))
	oakTimber := render.MixKubelkaMunk(render.PW(render.PigmentBurntSienna, 0.62), render.PW(render.PigmentYellowOchre, 0.38))
	crimsonCodex := render.MixKubelkaMunk(render.PW(render.PigmentVermilion, 0.68), render.PW(render.PigmentBurntSienna, 0.32))
	azuriteCodex := render.MixKubelkaMunk(render.PW(render.PigmentAzurite, 0.70), render.PW(render.PigmentUltramarine, 0.30))
	umberVault := render.MixKubelkaMunk(render.PW(render.PigmentRawUmber, 0.60), render.PW(render.PigmentUltramarine, 0.40))
	deepAbyss := render.MixKubelkaMunk(render.PW(render.PigmentPineSootInk, 0.75), render.PW(render.PigmentUltramarine, 0.25))
	shaftMist := render.MixKubelkaMunk(render.PW(render.PigmentCadmiumYellow, 0.36), render.PW(render.PigmentTitaniumWhite, 0.64))

	lightDir3D := render.V3(-0.52, 0.56, 0.64).Normalize()

	// Exact 6-fold hexagonal rotational period: rotating by (2*pi/6)*phase advances
	// every hexagonal gallery wall by exactly 1 sector per loop cycle!
	rotHex := (2.0 * math.Pi / 6.0) * phase
	rotAstrolabe := -(2.0 * math.Pi / 4.0) * phase

	// Central suspended Borges "Spherical Brass Lamp & Octahedral Astrolabe" in the shaft!
	evalCenterAstrolabe := func(p render.Vec3) (float64, int, render.Vec3) {
		qRing := p.RotateX(0.28).RotateY(rotHex * 2.0)
		dRing1 := render.SDF3DTorus(qRing, 0.23, 0.015)

		qMer := p.RotateZ(-0.24).RotateY(rotAstrolabe).RotateX(math.Pi * 0.5)
		dRing2 := render.SDF3DTorus(qMer, 0.185, 0.013)

		qCore := p.RotateX(0.22).RotateY(rotAstrolabe)
		fOct := render.FoldOctahedral(qCore)
		dCore := render.SDF3DPlatonicCage(qCore, 2, 0.135, 0.014)
		dLamp := render.SDF3DSphere(p, 0.068)

		minD := dRing1
		subID := 40
		loc := qRing
		if dRing2 < minD {
			minD = dRing2
			subID = 41
			loc = qMer
		}
		if dCore < minD {
			minD = dCore
			subID = 42
			loc = fOct
		}
		if dLamp < minD {
			minD = dLamp
			subID = 43
			loc = fOct
		}
		return minD, subID, loc
	}

	astrolabeSDF := func(p render.Vec3) float64 {
		d, _, _ := evalCenterAstrolabe(p)
		return d
	}

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)
	cfg := render.DefaultMoebiusConfig(seed)
	cfg.WarmBounceColor = color.RGBA{R: 0xe2, G: 0x84, B: 0x3c, A: 255}
	cfg.HatchSpacingPx = 3.6

	render.RenderMoebiusScene(rc, cfg, func(ux, uy float64) render.MoebiusPixel {
		x := ux*2.0 - 1.0
		y := uy*2.0 - 1.0

		// =====================================================================
		// 1. RAYMARCH CENTRAL SUSPENDED BORGES BRASS ASTROLABE & LAMP ORB
		// =====================================================================
		rScreen := math.Hypot(x, y)
		if rScreen < 0.30 {
			ro := render.V3(x*0.92, -y*0.92, 1.8)
			rd := render.V3(0, 0, -1)
			hit := render.RaymarchSDF3D(astrolabeSDF, ro, rd, lightDir3D, 2.8, 56)
			if hit.Hit {
				_, subID, pLoc := evalCenterAstrolabe(hit.Pos)
				nLoc := hit.Normal.RotateY(rotAstrolabe)
				stone := render.EvaluateMoebiusStoneRelief3D(ne, pLoc.Scale(3.5), nLoc, 7.0, 0.25)
				lit := common.Clamp01((0.18 + 0.82*math.Max(0.0, hit.Normal.Dot(lightDir3D))) *
					(0.45 + 0.55*hit.Shadow) * hit.AO)
				burinInk, burinChalk := render.Evaluate3DBurinEngraving(pLoc, nLoc, lit, 90.0)

				sCol := lampGold
				mCol := oakTimber
				if subID == 42 {
					sCol = azuriteCodex
					mCol = lampGold
				} else if subID == 43 {
					lit = 0.96
				}

				nScreen := moebiusNorm3(hit.Normal.X, -hit.Normal.Y, hit.Normal.Z)
				return render.MoebiusPixel{
					Depth:           1.2 + hit.T*0.4,
					ObjID:           subID,
					PartID:          subID*10 + int(math.Floor((pLoc.X+pLoc.Y)*12.0)),
					DetailID:        stone.BlockID,
					Normal:          nScreen,
					FlowTangent:     [2]float64{-nScreen[1], nScreen[0]},
					Illum:           lit,
					Spec:            common.Clamp01(0.65*math.Max(0.0, hit.Curvature) + burinChalk + 0.55),
					Bounce:          0.48,
					SunColor:        sCol,
					MidColor:        mCol,
					ShadeColor:      umberVault,
					DeepShadowColor: deepAbyss,
					CrackInk:        stone.CrackInk * 0.45,
					TickMask:        burinInk * 0.82,
				}
			}
		}

		// =====================================================================
		// 2. INFINITE RECEDING HEXAGONAL BOOK-ALCOVE GALLERIES (Borges Hexagons)
		// =====================================================================
		theta := math.Atan2(y, x) + rotHex
		// Fold angle into 6 regular hexagonal sectors (each sector spans pi/3 = 60 deg)
		const secSpan = math.Pi / 3.0
		normAng := math.Mod(theta+4.0*math.Pi, 2.0*math.Pi)
		secIdx := int(math.Floor(normAng / secSpan))
		localAng := normAng - (float64(secIdx)+0.5)*secSpan // in [-pi/6, +pi/6]

		// Exact radial distance in regular hexagon metric: rHex = r * cos(localAng) / cos(pi/6)
		rHex := rScreen * math.Cos(localAng) / 0.8660254037844386
		// Tangential coordinate along the hexagonal wall in [-1, 1]
		uWall := math.Tan(localAng) / 0.5773502691896258

		// Volumetric golden light streaming from the central ventilation shaft
		shaftRay := 0.5 + 0.5*math.Cos(localAng*6.0)
		shaftGlow := common.Smoothstep(0.75, 0.16, rHex) * (0.65 + 0.35*shaftRay)

		if rHex >= 0.21 {
			// Logarithmic perspective depth along the hexagonal chasm:
			// Each tier corresponds to one stories-high hexagonal gallery floor!
			logZ := math.Log(rHex/0.21) * 3.85 * bl.shaftScale
			tierIdx := int(math.Floor(logZ))
			tierFrac := logZ - float64(tierIdx) // [0, 1): radial position within one gallery tier

			// Corner stone piers between hexagonal walls (|uWall| > 0.84)
			isCornerPier := math.Abs(uWall) > 0.84
			// Balustrade railing lip at the inner edge of each hexagonal tier (tierFrac < 0.15)
			isBalustrade := tierFrac < 0.15
			// Molded stone cornice at the outer edge of each tier (tierFrac > 0.86)
			isCornice := tierFrac > 0.86

			// Hanging spherical Borges brass lamps suspended at the 6 hexagon corners!
			lampU := 0.0
			lampV := 0.24
			distLamp := math.Hypot((uWall-lampU)*1.8, (tierFrac-lampV)*1.5)
			if distLamp < 0.085 && (tierIdx%2 == 0) {
				lampLit := common.Clamp01(1.0 - distLamp/0.085*0.45)
				return render.MoebiusPixel{
					Depth:           8.0 - 4.5*common.Clamp01(rHex) - 0.3,
					ObjID:           25,
					PartID:          250 + tierIdx*6 + secIdx,
					DetailID:        int(distLamp * 40.0),
					Normal:          moebiusNorm3(uWall-lampU, tierFrac-lampV, 0.8),
					FlowTangent:     [2]float64{-math.Sin(theta), math.Cos(theta)},
					Illum:           0.85 + 0.15*lampLit,
					Spec:            0.88,
					Bounce:          0.60,
					SunColor:        lampGold,
					MidColor:        oakTimber,
					ShadeColor:      umberVault,
					DeepShadowColor: deepAbyss,
				}
			}

			// 3D object-locked stone & timber coordinates (uWall, logZ) rotate rigidly with the hexagon!
			pWall3D := render.V3(uWall*2.2, logZ*1.4, float64(secIdx)*0.7)
			// Inward-pointing normal of this hexagonal wall in world space
			wallAngleWorld := (float64(secIdx)+0.5)*secSpan - rotHex
			inwardNx := -math.Cos(wallAngleWorld)
			inwardNy := -math.Sin(wallAngleWorld)
			tangNx := -math.Sin(wallAngleWorld)
			tangNy := math.Cos(wallAngleWorld)

			stone := render.EvaluateMoebiusStoneRelief3D(ne, pWall3D, render.V3(0, 0, 1), 6.5, 0.28)

			// 5 Borges Book-Alcoves per hexagonal wall (|uWall| <= 0.84)
			alcoveCoord := (uWall + 0.84) / 1.68 * 5.0 // [0, 5]
			alcoveIdx := int(math.Floor(alcoveCoord))
			alcoveU := (alcoveCoord - float64(alcoveIdx))*2.0 - 1.0 // [-1, 1] inside alcove
			alcoveV := (tierFrac - 0.15) / 0.71                      // [0, 1] inside gallery wall

			inBookcase := !isCornerPier && !isBalustrade && !isCornice && math.Abs(alcoveU) < 0.76 && alcoveV > 0.08 && alcoveV < 0.90

			var nScreen [3]float64
			var sCol, mCol [3]float64
			var partID, detailID int
			var extraInk float64

			if isBalustrade || isCornice {
				// Carved Limestone & Oak Hexagonal Balustrade / Cornice
				nScreen = moebiusNorm3(inwardNx*0.72, inwardNy*0.72, 0.68)
				sCol = sunLimestone
				mCol = oakTimber
				balusterPosts := int(math.Floor((uWall + 1.0) * 12.0))
				partID = 100 + tierIdx*16 + secIdx*2 + int(moebiusBoolFloat(isBalustrade))
				detailID = balusterPosts
				if isBalustrade && (balusterPosts&1 == 0) && tierFrac > 0.04 && tierFrac < 0.11 {
					extraInk = 0.45
				}
			} else if isCornerPier {
				// Fluted Hexagonal Corner Pilaster & Spiral Vestibule Column
				nScreen = moebiusNorm3(inwardNx*0.55+tangNx*0.48, inwardNy*0.55+tangNy*0.48, 0.68)
				sCol = sunLimestone
				mCol = oakTimber
				partID = 200 + tierIdx*8 + secIdx
				detailID = stone.BlockID
			} else if inBookcase {
				// Borges 4-Shelf Bookcase Alcove with rows of ancient leather & mineral codices!
				shelfIdx := int(math.Floor(alcoveV * 4.0))
				shelfFrac := alcoveV*4.0 - float64(shelfIdx)
				codexIdx := int(math.Floor((alcoveU + 0.76) * 9.0))

				nScreen = moebiusNorm3(inwardNx*0.82, inwardNy*0.82, 0.56)
				partID = 300 + tierIdx*32 + secIdx*5 + alcoveIdx
				detailID = shelfIdx*16 + codexIdx

				if shelfFrac < 0.18 {
					// Horizontal oak bookshelf plank
					sCol = sunLimestone
					mCol = oakTimber
				} else {
					// Vertical book spines in alternating ochre, crimson-cinnabar, and azurite-lapis bindings!
					spineHash := (tierIdx*7 + secIdx*13 + alcoveIdx*5 + shelfIdx*3 + codexIdx) & 7
					switch spineHash {
					case 0, 3:
						sCol = lampGold
						mCol = oakTimber
					case 1, 5:
						sCol = crimsonCodex
						mCol = umberVault
					case 2, 6:
						sCol = azuriteCodex
						mCol = umberVault
					default:
						sCol = oakTimber
						mCol = umberVault
					}
					extraInk = 0.38
				}
			} else {
				// Coffered Stone Arch Frame between Book-Alcoves
				nScreen = moebiusNorm3(inwardNx*0.65+tangNx*0.30*alcoveU, inwardNy*0.65+tangNy*0.30*alcoveU, 0.70)
				sCol = sunLimestone
				mCol = oakTimber
				partID = 400 + tierIdx*16 + secIdx*5 + alcoveIdx
				detailID = stone.BlockID
			}

			// Directional illumination from upper-left + warm golden shaft lantern glow
			dotL := inwardNx*(-0.58) + inwardNy*(-0.52) + 0.62*nScreen[2]
			lit := common.Clamp01(0.14 + 0.68*math.Max(0.0, dotL) + 0.28*shaftGlow)
			if inBookcase {
				lit *= 0.76 // recessed shadow inside book alcoves
			}

			burinInk, burinChalk := render.Evaluate3DBurinEngraving(
				pWall3D,
				render.V3(0, 0, 1),
				lit,
				54.0,
			)
			atmos := common.Smoothstep(0.88, 0.21, rHex) * 0.46

			return render.MoebiusPixel{
				Depth:           9.5 - 6.2*common.Clamp01(rHex) - 0.25*moebiusBoolFloat(isBalustrade || isCornerPier),
				ObjID:           10 + (tierIdx & 7),
				PartID:          partID,
				DetailID:        detailID,
				Normal:          nScreen,
				FlowTangent:     [2]float64{tangNx, tangNy},
				Illum:           lit,
				Spec:            common.Clamp01(burinChalk*0.85 + moebiusBoolFloat(isBalustrade)*0.45),
				Bounce:          0.45,
				Atmos:           atmos,
				SunColor:        sCol,
				MidColor:        mCol,
				ShadeColor:      umberVault,
				DeepShadowColor: deepAbyss,
				AtmosColor:      shaftMist,
				CrackInk:        math.Max(stone.CrackInk*0.52, extraInk),
				TickMask:        math.Max(stone.PockInk*0.42, burinInk*0.78),
				Stipple:         atmos * 0.40,
			}
		}

		// =====================================================================
		// 3. CENTRAL INFINITE VENTILATION SHAFT ABYSS (rHex < 0.21)
		// =====================================================================
		coreT := common.Clamp01(rHex / 0.21)
		return render.MoebiusPixel{
			Depth:           28.0,
			ObjID:           2,
			PartID:          20,
			DetailID:        20,
			Normal:          [3]float64{0, 0, 1},
			FlowTangent:     [2]float64{-math.Sin(theta), math.Cos(theta)},
			Illum:           0.92 - 0.25*coreT,
			Spec:            0.78 * (1.0 - coreT),
			SunColor:        moebiusLerp3(lampGold, shaftMist, coreT*0.65),
			MidColor:        oakTimber,
			ShadeColor:      umberVault,
			DeepShadowColor: deepAbyss,
			Stipple:         coreT * 0.45,
		}
	})

	c.MergeRenderCanvas(rc)
}
