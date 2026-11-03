package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type moebiusEdena struct {
	domes      int
	terraces   int
	floraScale float64
}

// NewMoebiusEdena returns a MoebiusEdena generator inspired by Jean "Mœbius"
// Giraud's Le Monde d'Edena and 40 Days in the Desert B: an asymmetric, gnarled
// sandstone canyon overhang and natural rock arch framing the left foreground,
// billowing golden-peach cumulus cloud banks, a stepped ancient pyramid & oasis
// sanctuary domes on the right, and a winding turquoise-jade desert lagoon.
func NewMoebiusEdena(domes, terraces int, floraScale float64) *moebiusEdena {
	if domes <= 0 {
		domes = 5
	}
	if terraces <= 0 {
		terraces = 16
	}
	if floraScale <= 0 {
		floraScale = 1.0
	}
	return &moebiusEdena{
		domes:      domes,
		terraces:   terraces,
		floraScale: floraScale,
	}
}

// Generative renders the static MoebiusEdena Ligne Claire print.
func (me *moebiusEdena) Generative(c *generativeart.Canva) {
	me.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless looping MoebiusEdena animation.
func (me *moebiusEdena) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	seed := c.Seed()
	if seed == 0 {
		seed = 20262103
	}
	ne := common.NewNoiseEngine(seed)
	phi := 2.0 * math.Pi * phase

	// Authentic Edena / Mexican Desert Watercolor & Gouache Palette
	skyTop := moebiusLinRGB(0x2c, 0x5c, 0x9c)        // #2c5c9c Mexican Desert Cerulean Sky
	skyMid := moebiusLinRGB(0x68, 0xc2, 0xc2)        // #68c2c2 Luminous Edena Aquamarine
	skyLow := moebiusLinRGB(0xfc, 0xe2, 0xa8)        // #fce2a8 Warm Apricot-Cream Horizon
	atmosHaze := moebiusLinRGB(0xf8, 0xd0, 0x9e)     // #f8d09e Sunlit Desert Dust Veil

	cloudSun := moebiusLinRGB(0xff, 0xf8, 0xe6)      // #fff8e6 Sunlit Cumulus Crest
	cloudMid := moebiusLinRGB(0xf6, 0x9e, 0x72)      // #f69e72 Warm Peach-Coral Cloud Billow
	cloudShade := moebiusLinRGB(0x68, 0x62, 0x9e)    // #68629e Lavender-Violet Cloud Underside

	archSun := moebiusLinRGB(0xf8, 0xc8, 0x74)       // #f8c874 Sunlit Sandstone Arch Crest
	archMid := moebiusLinRGB(0xe2, 0x6c, 0x46)       // #e26c46 Warm Terracotta-Sienna Rock
	archShade := moebiusLinRGB(0x42, 0x46, 0x80)     // #424680 Cool Violet-Indigo Cliff Shadow
	archDeep := moebiusLinRGB(0x18, 0x1a, 0x34)      // #181a34 Deep Overhang Chiaroscuro

	duneSun := moebiusLinRGB(0xfc, 0xe4, 0x9e)       // #fce49e Sunlit Golden Desert Sand
	duneMid := moebiusLinRGB(0xee, 0xa2, 0x62)       // #eea262 Warm Apricot Dune Slope
	duneShade := moebiusLinRGB(0x64, 0x68, 0xa2)     // #6468a2 Lavender-Blue Dune Slip-Face

	lagoonBright := moebiusLinRGB(0x58, 0xec, 0xd8)  // #58ecd8 Sparkling Turquoise Oasis Water
	lagoonMid := moebiusLinRGB(0x24, 0xa8, 0xa2)     // #24a8a2 Emerald-Jade Lagoon
	lagoonDeep := moebiusLinRGB(0x18, 0x5e, 0x7a)    // #185e7a Deep Oasis Basin Shadow

	domeIvory := moebiusLinRGB(0xfd, 0xf8, 0xea)     // #fdf8ea Alabaster Sanctuary Dome
	domeGold := moebiusLinRGB(0xf8, 0xbe, 0x48)      // #f8be48 Gilded Spire Crown
	domeCyan := moebiusLinRGB(0x42, 0xdc, 0xd0)      // #42dcd0 Turquoise Sanctuary Portal

	lightDir := moebiusNorm3(0.62, -0.46, 0.63)

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)
	cfg := render.DefaultMoebiusConfig(seed)
	cfg.WarmBounceColor = color.RGBA{R: 0xf4, G: 0x98, B: 0x58, A: 255}

	render.RenderMoebiusScene(rc, cfg, func(ux, uy float64) render.MoebiusPixel {
		x := ux*2.0 - 1.0
		y := uy*2.0 - 1.0

		crag := ne.FBM2D(x*4.2+0.9, y*4.2-1.3, 4, 2.0, 0.5)
		wobble := 0.024 * crag

		// =====================================================================
		// 1. ASYMMETRIC FOREGROUND GNARLED SANDSTONE ARCH & LEFT CLIFF BASTION
		// =====================================================================
		cliffRightEdge := -0.54 - 0.08*math.Sin(y*3.2) + wobble*1.2
		inLeftCliff := x <= cliffRightEdge

		archBottomCurve := -0.44 - 0.52*(x+0.54) - 0.16*math.Sin((x+0.54)*4.5) + wobble*1.1
		archTopCurve := -0.78 - 0.35*(x+0.54) + wobble
		inUpperArch := x >= cliffRightEdge-0.05 && x <= 0.28 && y <= archBottomCurve && y >= archTopCurve-0.35

		talusTop := 0.52 + 0.62*(x+0.54) + wobble*1.3
		inTalus := x <= -0.12 && y >= talusTop

		if inLeftCliff || inUpperArch || inTalus {
			stone := render.EvaluateMoebiusStoneRelief(ne, x, y, 7.2, 8.5, 0.45)

			var nx, ny float64
			if inUpperArch && !inLeftCliff {
				uArch := (y - archTopCurve) / math.Max(archBottomCurve-archTopCurve, 1e-3)
				nx = 0.25 + 0.22*stone.NormalMod[0]
				ny = (uArch*2.0 - 0.85) + 0.22*stone.NormalMod[1]
			} else if inTalus && !inLeftCliff {
				nx = 0.48 + 0.24*stone.NormalMod[0]
				ny = -0.52 + 0.24*stone.NormalMod[1]
			} else {
				uCliff := (x - (-1.0)) / math.Max(cliffRightEdge-(-1.0), 1e-3)
				nx = (uCliff*1.5 - 0.45) + 0.26*stone.NormalMod[0]
				ny = -0.18 + 0.26*stone.NormalMod[1]
			}
			nz := math.Sqrt(math.Max(0.12, 1.0-common.Clamp01(nx*nx+ny*ny)*0.85))
			n := moebiusNorm3(nx, ny, nz)

			lit := common.Clamp01(0.14 + 0.88*math.Max(0.0, moebiusDot3(n, lightDir)))
			if inUpperArch && y > archBottomCurve-0.12 {
				lit *= 0.34
			}

			sunC := moebiusLerp3(archMid, archSun, common.Smoothstep(0.5, -0.7, y)+0.1*stone.ColorMod)
			organicPart := int(math.Floor((y+1.0)*2.8 + 0.85*stone.ColorMod))
			return render.MoebiusPixel{
				Depth:           1.5 - 0.2*nz,
				ObjID:           40,
				PartID:          400 + (organicPart & 15),
				DetailID:        stone.BlockID,
				Normal:          n,
				FlowTangent:     [2]float64{0.96, -0.24 * nx},
				Illum:           lit,
				Spec:            common.Smoothstep(0.78, 0.98, lit) * 0.60,
				Bounce:          0.42,
				SunColor:        sunC,
				MidColor:        archMid,
				ShadeColor:      archShade,
				DeepShadowColor: archDeep,
				CrackInk:        stone.CrackInk,
				TickMask:        stone.PockInk,
			}
		}

		// =====================================================================
		// 2. MIDGROUND BIOMORPHIC EDENA SANCTUARY DOMES & FLUTED MINARETS
		// =====================================================================
		type sanctuarySpire struct {
			cx, baseY  float64
			radX, radY float64
			spireH     float64
			id         int
		}
		sanctuaries := [3]sanctuarySpire{
			{cx: 0.34, baseY: 0.36, radX: 0.145, radY: 0.125, spireH: 0.22, id: 30},
			{cx: 0.58, baseY: 0.38, radX: 0.088, radY: 0.078, spireH: 0.16, id: 31},
			{cx: 0.12, baseY: 0.33, radX: 0.070, radY: 0.062, spireH: 0.14, id: 32},
		}
		for _, sc := range sanctuaries {
			dx := x - sc.cx
			dyBase := sc.baseY - y
			if dyBase >= 0 && dyBase <= sc.radY*1.6+sc.spireH {
				domeCenterY := sc.baseY - sc.radY*0.75
				if y < domeCenterY-sc.radY*0.72 {
					tSpire := (domeCenterY - sc.radY*0.72 - y) / sc.spireH
					halfW := 0.018 * (1.0 - 0.85*tSpire)
					if tSpire <= 1.0 && math.Abs(dx) <= halfW {
						return render.MoebiusPixel{
							Depth:       3.2,
							ObjID:       sc.id,
							PartID:      sc.id*100 + 90,
							DetailID:    int(tSpire * 6.0),
							Normal:      moebiusNorm3(dx/math.Max(halfW, 1e-4), 0, 0.7),
							FlowTangent: [2]float64{0, 1},
							Illum:       0.88,
							Spec:        0.75,
							Atmos:       0.12,
							SunColor:    domeGold,
							MidColor:    archMid,
							ShadeColor:  archShade,
							AtmosColor:  atmosHaze,
						}
					}
				}
				ex := dx / sc.radX
				ey := (y - domeCenterY) / sc.radY
				effRx := 1.0
				if ey < 0 {
					effRx = math.Sqrt(math.Max(0.0, 1.0-ey*ey)) * (1.0 + 0.18*ey)
				} else {
					effRx = 0.86
				}
				if math.Abs(ex) <= effRx && effRx > 1e-3 {
					uDome := ex / effRx
					nz := math.Sqrt(math.Max(0.10, 1.0-uDome*uDome))
					ribIdx := int(math.Floor((math.Atan2(uDome, nz) + math.Pi) / (2.0 * math.Pi) * 18.0))
					n := moebiusNorm3(uDome*0.85, ey*0.45, nz)
					lit := common.Clamp01(0.20 + 0.84*math.Max(0.0, moebiusDot3(n, lightDir)))

					sCol := domeIvory
					if math.Abs(ey) < 0.14 {
						sCol = domeGold
					} else if ey > 0.25 && ribIdx%2 == 0 && math.Abs(uDome) < 0.65 {
						sCol = domeCyan
					}
					return render.MoebiusPixel{
						Depth:           3.3 - 0.15*nz,
						ObjID:           sc.id,
						PartID:          sc.id*100 + int((ey+1.0)*2.5)*4 + (ribIdx/2)&3,
						DetailID:        ribIdx,
						Normal:          n,
						FlowTangent:     [2]float64{0.96, 0.22 * uDome},
						Illum:           lit,
						Spec:            common.Smoothstep(0.75, 0.96, lit) * 0.70,
						Bounce:          0.40,
						Atmos:           0.12,
						SunColor:        sCol,
						MidColor:        duneMid,
						ShadeColor:      duneShade,
						DeepShadowColor: archDeep,
						AtmosColor:      atmosHaze,
					}
				}
			}
		}

		// =====================================================================
		// 3. DISTANT DESERT MESAS, WIND-RIPPLED DUNES & WINDING OASIS LAGOON
		// =====================================================================
		mesaHorizon := 0.20 - 0.04*math.Sin(x*4.2) + 0.4*wobble
		if x > -0.35 && x < 0.15 {
			mesaHorizon -= 0.08 * (1.0 - math.Abs(x+0.10)/0.25)
		}
		if y >= mesaHorizon && y < 0.28 {
			stone := render.EvaluateMoebiusStoneRelief(ne, x, y, 12.0, 14.0, 0.35)
			lit := 0.80
			if x < -0.10 {
				lit = 0.38
			}
			return render.MoebiusPixel{
				Depth:       8.5,
				ObjID:       12,
				PartID:      120 + int(moebiusBoolFloat(x < -0.10)),
				DetailID:    stone.BlockID,
				Normal:      [3]float64{0.3, -0.4, 0.86},
				FlowTangent: [2]float64{1, 0},
				Illum:       lit,
				Atmos:       0.46,
				SunColor:    archSun,
				MidColor:    archMid,
				ShadeColor:  duneShade,
				AtmosColor:  atmosHaze,
				CrackInk:    stone.CrackInk * 0.55,
			}
		}

		if y >= 0.26 {
			ty := (y - 0.26) / 0.74
			riverCenter := 0.12 + 0.26*math.Sin(ty*math.Pi*1.65-0.4)
			riverHalfW := (0.14 + 0.32*ty) * (0.78 + 0.22*math.Cos(ty*math.Pi*3.0))
			distRiver := (x - riverCenter) / math.Max(riverHalfW, 1e-3)

			if math.Abs(distRiver) <= 1.0 {
				poolCoord := math.Pow(ty, 0.78)*10.0 + 0.28*math.Cos(distRiver*math.Pi) + 0.15*crag
				poolIdx := int(math.Floor(poolCoord))
				poolFrac := poolCoord - float64(poolIdx)

				if poolFrac < 0.20 {
					return render.MoebiusPixel{
						Depth:       5.5 - 3.0*ty,
						ObjID:       15 + (poolIdx & 3),
						PartID:      1500 + poolIdx*2,
						DetailID:    poolIdx,
						Normal:      [3]float64{0, -0.7, 0.7},
						FlowTangent: [2]float64{1, 0},
						Illum:       0.88,
						Spec:        0.65,
						Atmos:       0.24 * (1.0 - ty),
						SunColor:    domeIvory,
						MidColor:    archSun,
						ShadeColor:  duneShade,
						AtmosColor:  atmosHaze,
					}
				}

				uWater := (poolFrac - 0.20) / 0.80
				ripple := 0.5 + 0.5*math.Sin((uWater*3.0+0.2*distRiver-phase)*2.0*math.Pi)
				wCol := moebiusLerp3(lagoonBright, lagoonMid, 0.35*math.Abs(distRiver)+0.45*uWater)
				lit := 0.86
				if distRiver > 0.68 || uWater < 0.15 {
					lit = 0.40
				}
				return render.MoebiusPixel{
					Depth:       5.6 - 3.0*ty,
					ObjID:       15 + (poolIdx & 3),
					PartID:      1500 + poolIdx*2 + 1,
					DetailID:    poolIdx,
					Normal:      [3]float64{0, -0.1, 0.99},
					FlowTangent: [2]float64{1, 0},
					Illum:       lit,
					Spec:        common.Smoothstep(0.78, 0.98, ripple) * 0.65,
					Atmos:       0.24 * (1.0 - ty),
					SunColor:    wCol,
					MidColor:    lagoonMid,
					ShadeColor:  lagoonDeep,
					AtmosColor:  atmosHaze,
					TickMask:    common.Smoothstep(0.91, 0.98, ripple) * 0.60,
				}
			}

			// Sweeping Aeolian Desert Sand Dunes & Weathered Boulders (NO vertical grid!)
			duneCoord := ty*5.5 + 0.45*math.Sin(x*3.8) + 0.22*crag
			duneIdx := int(math.Floor(duneCoord))
			duneFrac := duneCoord - float64(duneIdx)
			isSlipFace := duneFrac < 0.36 && (distRiver < 0 || duneIdx&1 == 0)
			lit := 0.86
			if isSlipFace {
				lit = 0.36
			}
			windRipple := common.Smoothstep(0.92, 0.98, math.Sin((ty*42.0+x*14.0+0.6*crag)*math.Pi)) * 0.48
			return render.MoebiusPixel{
				Depth:       5.4 - 3.0*ty,
				ObjID:       14,
				PartID:      1400 + duneIdx*2 + int(moebiusBoolFloat(isSlipFace)),
				DetailID:    duneIdx, // No x-grid quantization!
				Normal:      moebiusNorm3(-0.3*moebiusBoolFloat(isSlipFace), -0.6, 0.75),
				FlowTangent: [2]float64{0.94, 0.34},
				Illum:       lit,
				Bounce:      0.45,
				Atmos:       0.25 * (1.0 - ty),
				SunColor:    duneSun,
				MidColor:    duneMid,
				ShadeColor:  duneShade,
				AtmosColor:  atmosHaze,
				TickMask:    windRipple,
				Stipple:     common.Smoothstep(0.30, 0.52, duneFrac) * 0.42,
			}
		}

		// =====================================================================
		// 4. GIANT CRESCENT PLANET, BILLOWING PEACH CUMULUS CLOUDS & SKY
		// =====================================================================
		pxP, pyP := x-0.46, y-(-0.46)
		rP := math.Hypot(pxP, pyP)
		if rP <= 0.21 || math.Abs(rP-0.265) <= 0.012 {
			if rP > 0.22 {
				return render.MoebiusPixel{
					Depth:       60.0,
					ObjID:       51,
					PartID:      510,
					DetailID:    510,
					Normal:      [3]float64{0, 0, 1},
					FlowTangent: [2]float64{-pyP, pxP},
					Illum:       0.92,
					SunColor:    domeGold,
					MidColor:    archSun,
					ShadeColor:  skyMid,
				}
			}
			nxP, nyP := pxP/0.21, pyP/0.21
			nzP := math.Sqrt(math.Max(0.05, 1.0-nxP*nxP-nyP*nyP))
			stoneP := render.EvaluateMoebiusStoneRelief(ne, x*2.0, y*2.0, 9.0, 9.0, 0.3)
			lit := common.Clamp01(0.26 + 0.76*math.Max(0.0, moebiusDot3(moebiusNorm3(nxP, nyP, nzP), lightDir)))
			return render.MoebiusPixel{
				Depth:       62.0,
				ObjID:       50,
				PartID:      500,
				DetailID:    stoneP.BlockID,
				Normal:      moebiusNorm3(nxP, nyP, nzP),
				FlowTangent: [2]float64{0.95, -0.3 * nxP},
				Illum:       lit,
				Spec:        common.Smoothstep(0.35, 0.05, nzP) * 0.55,
				SunColor:    domeIvory,
				MidColor:    skyLow,
				ShadeColor:  skyMid,
				CrackInk:    stoneP.CrackInk * 0.38,
				Stipple:     (1.0 - lit) * 0.42,
			}
		}

		// Billowing Mœbius Golden-Peach & Lavender Cumulus Cloud Bank above the Horizon
		cloudCenter := 0.02 - 0.08*x + 0.06*math.Sin(x*3.5+phi)
		cloudDist := y - cloudCenter
		cloudHalfW := 0.14 + 0.06*crag
		if math.Abs(cloudDist) <= cloudHalfW && y < 0.20 {
			uCloud := (cloudDist/cloudHalfW + 1.0) * 0.5
			lobeIdx := int(math.Floor(x*5.0 + 1.2*crag))
			lit := common.Clamp01(0.88 - 0.62*common.Smoothstep(0.40, 0.92, uCloud))
			return render.MoebiusPixel{
				Depth:       70.0,
				ObjID:       45,
				PartID:      450 + (lobeIdx & 15),
				DetailID:    int(uCloud * 2.5),
				Normal:      moebiusNorm3(0, uCloud-0.5, 0.8),
				FlowTangent: [2]float64{1, 0},
				Illum:       lit,
				Spec:        common.Smoothstep(0.3, 0.0, uCloud) * 0.65,
				SunColor:    cloudSun,
				MidColor:    cloudMid,
				ShadeColor:  cloudShade,
				Stipple:     common.Smoothstep(0.45, 0.95, uCloud) * 0.45,
			}
		}

		skyT := common.Clamp01((y + 1.0) / 1.28)
		var sCol [3]float64
		if skyT < 0.48 {
			sCol = moebiusLerp3(skyTop, skyMid, skyT/0.48)
		} else {
			sCol = moebiusLerp3(skyMid, skyLow, (skyT-0.48)/0.52)
		}

		return render.MoebiusPixel{
			Depth:       100.0,
			ObjID:       1,
			PartID:      1,
			DetailID:    1,
			Normal:      [3]float64{0, 0, 1},
			FlowTangent: [2]float64{1, 0},
			Illum:       0.86,
			SunColor:    sCol,
			MidColor:    skyMid,
			ShadeColor:  skyTop,
			Stipple:     common.Smoothstep(0.14, 0.0, math.Abs(skyT-0.45)) * 0.36,
		}
	})

	c.MergeRenderCanvas(rc)
}
