package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type piranesiCarceri struct {
	arches     int
	tiers      int
	vaultDepth float64
}

// NewPiranesiCarceri returns a PiranesiCarceri generator inspired by Giovanni
// Battista Piranesi's Carceri d'Invenzione (Imaginary Prisons) and Mœbius's
// La Citadelle Aveugle: a 4-tier subterranean vaulted stone labyrinth of Romanesque
// viaducts, coffered barrel vaults, spiral staircases, suspension catwalks, and
// volumetric sunbeams rendered with Kubelka-Munk sepia/sienna/indigo pigments and
// Hessian principal-curvature burin cross-hatching.
func NewPiranesiCarceri(arches, tiers int, vaultDepth float64) *piranesiCarceri {
	if arches <= 0 {
		arches = 4
	}
	if tiers <= 0 {
		tiers = 4
	}
	if vaultDepth <= 0 {
		vaultDepth = 1.0
	}
	return &piranesiCarceri{
		arches:     arches,
		tiers:      tiers,
		vaultDepth: vaultDepth,
	}
}

// Generative renders the static PiranesiCarceri intaglio & gouache composition.
func (pc *piranesiCarceri) Generative(c *generativeart.Canva) {
	pc.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless looping PiranesiCarceri animation.
func (pc *piranesiCarceri) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	seed := c.Seed()
	if seed == 0 {
		seed = 20262201
	}
	ne := common.NewNoiseEngine(seed)
	phi := 2.0 * math.Pi * phase

	// Kubelka-Munk Physical Mineral & Etching Ink Palette
	sunGold := render.MixKubelkaMunk(render.PW(render.PigmentGoldLeaf, 0.65), render.PW(render.PigmentTitaniumWhite, 0.35))
	sunStone := render.MixKubelkaMunk(render.PW(render.PigmentYellowOchre, 0.55), render.PW(render.PigmentTitaniumWhite, 0.45))
	midSienna := render.MixKubelkaMunk(render.PW(render.PigmentBurntSienna, 0.60), render.PW(render.PigmentYellowOchre, 0.40))
	shadeUmber := render.MixKubelkaMunk(render.PW(render.PigmentRawUmber, 0.58), render.PW(render.PigmentUltramarine, 0.42))
	deepCarceri := render.MixKubelkaMunk(render.PW(render.PigmentPineSootInk, 0.72), render.PW(render.PigmentUltramarine, 0.28))
	sunbeamMist := render.MixKubelkaMunk(render.PW(render.PigmentCadmiumYellow, 0.35), render.PW(render.PigmentTitaniumWhite, 0.65))

	lightDir := moebiusNorm3(-0.55, -0.58, 0.60)

	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	// Store heightfield & illumination buffers for principal-curvature burin pass
	heightBuf := make([]float64, w*h)
	illumBuf := make([]float64, w*h)

	cfg := render.DefaultMoebiusConfig(seed)
	cfg.WarmBounceColor = color.RGBA{R: 0xd8, G: 0x72, B: 0x3c, A: 255}
	cfg.HatchSpacingPx = 3.8

	render.RenderMoebiusScene(rc, cfg, func(ux, uy float64) render.MoebiusPixel {
		x := ux*2.0 - 1.0
		y := uy*2.0 - 1.0

		// Diagonal volumetric sunbeams streaming from top-left oculi across the vault!
		beamCoord := (x + 0.62*y + 0.25) * 3.8
		beamWave := 0.5 + 0.5*math.Cos(beamCoord*math.Pi*2.0-phi)
		sunbeamIntensity := math.Pow(beamWave, 3.2) * common.Smoothstep(0.85, -0.85, x*0.5+y*0.5)

		// =====================================================================
		// 4 RECEDING ARCHITECTURAL VAULT & VIADUCT PLANES (zLayer = 0..3)
		// =====================================================================
		for zLayer := 0; zLayer < 4; zLayer++ {
			fz := float64(zLayer)
			perspScale := 1.0 + 0.58*fz // deeper layers have finer architectural scale
			depthVal := 1.6 + 1.8*fz
			atmos := common.Clamp01(0.14*fz + 0.28*sunbeamIntensity*common.Smoothstep(0.0, 2.0, fz))

			// Transform screen (x, y) into layer-local architectural coordinates (lx, ly).
			// Keep the stone vault architecture fixed in world space so Hessian principal-curvature
			// burin lines stay locked to the stone blocks while sunbeams and chains animate!
			shiftX := 0.24 * math.Sin(fz*1.9+0.4)
			shiftY := 0.16 * math.Cos(fz*2.3)
			lx := (x - shiftX) * perspScale
			ly := (y - shiftY) * perspScale

			//Foreground layer 0 is a massive framing barrel-vault arch around the border!
			if zLayer == 0 {
				archSpanX := x - (-0.06)
				archSpanY := y - 0.18
				rVault := math.Hypot(archSpanX/0.82, archSpanY/0.92)
				inForegroundPier := math.Abs(x-(-0.06)) > 0.74 || (y < 0.18 && rVault > 0.88) || y > 0.86

				// Diagonal suspension staircase & timber/stone catwalk spanning Layer 0 -> Layer 1
				catwalkY := 0.28 - 0.42*x
				inCatwalk := x >= -0.76 && x <= 0.78 && math.Abs(y-catwalkY) < 0.052
				inRailing := x >= -0.76 && x <= 0.78 && y >= catwalkY-0.095 && y < catwalkY-0.052 &&
					(math.Abs(y-(catwalkY-0.090)) < 0.008 || int(math.Floor((x+1.0)*28.0))%2 == 0)

				if !inForegroundPier && !inCatwalk && !inRailing {
					continue
				}

				stone := render.EvaluateMoebiusStoneRelief(ne, x, y, 8.5, 9.5, 0.38)
				var nx, ny float64
				part := 100 + int(math.Floor((y+1.0)*3.5+0.6*stone.ColorMod))

				if inRailing {
					return render.MoebiusPixel{
						Depth:           1.45,
						ObjID:           11,
						PartID:          119,
						DetailID:        int((x + 1.0) * 36.0),
						Normal:          [3]float64{0, -0.5, 0.86},
						FlowTangent:     [2]float64{0.92, -0.39},
						Illum:           0.32 + 0.45*sunbeamIntensity,
						SunColor:        sunStone,
						MidColor:        midSienna,
						ShadeColor:      shadeUmber,
						DeepShadowColor: deepCarceri,
					}
				}
				if inCatwalk && !inForegroundPier {
					// Stepped stone staircase catwalk with underside corbel arch
					stairIdx := int(math.Floor((x + 1.0) * 18.0))
					uCat := (y - (catwalkY - 0.052)) / 0.104
					lit := 0.76
					if uCat > 0.48 {
						lit = 0.24 // shadowed underside of catwalk
					}
					return render.MoebiusPixel{
						Depth:           1.50,
						ObjID:           11,
						PartID:          110 + (stairIdx/2)&15,
						DetailID:        stone.BlockID,
						Normal:          moebiusNorm3(0.3, uCat*1.4-0.6, 0.7),
						FlowTangent:     [2]float64{0.92, -0.39},
						Illum:           common.Clamp01(lit + 0.25*sunbeamIntensity),
						Spec:            common.Smoothstep(0.4, 0.0, uCat) * 0.60,
						Bounce:          0.45,
						SunColor:        sunGold,
						MidColor:        midSienna,
						ShadeColor:      shadeUmber,
						DeepShadowColor: deepCarceri,
						CrackInk:        stone.CrackInk * 0.75,
						TickMask:        stone.PockInk,
					}
				}

				// Massive foreground rusticated stone arch & coffered soffit
				if y < 0.18 && rVault > 0.88 && rVault < 1.04 {
					// Radial keystone voussoirs along the great vault rim
					angV := math.Atan2(archSpanY, archSpanX)
					voussoirIdx := int(math.Floor((angV + math.Pi) / (2.0 * math.Pi) * 32.0))
					part = 140 + voussoirIdx
					nx = -archSpanX * 0.85
					ny = -archSpanY * 0.85
				} else {
					nx = -common.Constrain(archSpanX*0.9, -0.8, 0.8) + 0.25*stone.NormalMod[0]
					ny = -0.20 + 0.25*stone.NormalMod[1]
				}
				nz := math.Sqrt(math.Max(0.12, 1.0-common.Clamp01(nx*nx+ny*ny)*0.85))
				n := moebiusNorm3(nx, ny, nz)
				lit := common.Clamp01(0.12 + 0.78*math.Max(0.0, moebiusDot3(n, lightDir)))
				// Smooth Piranesi chiaroscuro shadow across the left foreground pier (no vertical seam!)
				leftShadow := common.Smoothstep(-0.20, -0.72, x-0.25*y)
				lit *= 1.0 - 0.58*leftShadow

				pxIdx := int(ux * float64(w))
				pyIdx := int(uy * float64(h))
				if pxIdx >= 0 && pxIdx < w && pyIdx >= 0 && pyIdx < h {
					heightBuf[pyIdx*w+pxIdx] = rVault + 0.25*stone.ColorMod
					illumBuf[pyIdx*w+pxIdx] = lit
				}

				return render.MoebiusPixel{
					Depth:           depthVal - 0.2*nz,
					ObjID:           10,
					PartID:          part,
					DetailID:        stone.BlockID,
					Normal:          n,
					FlowTangent:     [2]float64{0.96, 0.24 * nx},
					Illum:           lit,
					Spec:            common.Smoothstep(0.75, 0.96, lit) * 0.55,
					Bounce:          0.42,
					SunColor:        sunStone,
					MidColor:        midSienna,
					ShadeColor:      shadeUmber,
					DeepShadowColor: deepCarceri,
					CrackInk:        stone.CrackInk,
					TickMask:        stone.PockInk,
				}
			}

			// Layers 1..3: Multi-story Romanesque Arched Viaduct Colonnades & Spiral Ramps
			colPeriod := 1.15
			bayX := math.Mod(lx+10.0*colPeriod, colPeriod) - 0.5*colPeriod // in [-0.5*colPeriod, 0.5*colPeriod]
			bayIdxX := int(math.Floor((lx + 10.0*colPeriod) / colPeriod))

			tierPeriod := 1.05
			bayY := math.Mod(ly+10.0*tierPeriod, tierPeriod) - 0.5*tierPeriod
			tierIdxY := int(math.Floor((ly + 10.0*tierPeriod) / tierPeriod))

			archOpenW := 0.36 * colPeriod
			archSpringY := -0.05 * tierPeriod
			entablatureY := -0.38 * tierPeriod

			// Determine if (bayX, bayY) is inside the arched opening or on the stone viaduct
			var rArchOpen float64
			inArchOpening := false
			if bayY >= archSpringY {
				inArchOpening = math.Abs(bayX) < archOpenW
			} else if bayY >= entablatureY {
				rArchOpen = math.Hypot(bayX/archOpenW, (bayY-archSpringY)/(archSpringY-entablatureY))
				inArchOpening = rArchOpen < 1.0
			}

			// Hanging stone/iron chains & swinging lanterns inside the arch openings of Layer 1 & 2!
			if inArchOpening {
				pendulumX := 0.08 + 0.065*math.Sin(phi+fz*1.7+float64(bayIdxX))*common.Clamp01((bayY-entablatureY)/0.75)
				if (zLayer == 1 || zLayer == 2) && math.Abs(bayX-pendulumX) < 0.018 {
					return render.MoebiusPixel{
						Depth:           depthVal + 0.1,
						ObjID:           20 + zLayer,
						PartID:          (20+zLayer)*100 + 99,
						DetailID:        int(ly * 32.0),
						Normal:          [3]float64{0.6, 0, 0.8},
						FlowTangent:     [2]float64{0, 1},
						Illum:           0.28 + 0.45*sunbeamIntensity,
						Atmos:           atmos,
						SunColor:        sunGold,
						MidColor:        midSienna,
						ShadeColor:      shadeUmber,
						DeepShadowColor: deepCarceri,
						AtmosColor:      sunbeamMist,
					}
				}
				continue
			}

			stone := render.EvaluateMoebiusStoneRelief(ne, lx*0.7, ly*0.7, 7.5, 8.5, 0.36)
			isEntablature := bayY < entablatureY
			isVoussoir := !isEntablature && bayY < archSpringY && rArchOpen >= 1.0 && rArchOpen < 1.24

			var nx, ny float64
			part := (20+zLayer)*100 + tierIdxY*8 + (bayIdxX & 7)
			if isEntablature {
				// Molded cornice parapet with balustrade slots
				uEnt := (bayY - (-0.5*tierPeriod)) / (entablatureY - (-0.5*tierPeriod))
				ny = -0.75 + 1.2*uEnt
				nx = 0.15 * stone.NormalMod[0]
				part += 64
			} else if isVoussoir {
				angV := math.Atan2(bayY-archSpringY, bayX)
				vIdx := int(math.Floor((angV + math.Pi) / (2.0 * math.Pi) * 18.0))
				part = (20+zLayer)*100 + 128 + vIdx
				nx = -bayX * 1.4
				ny = -(bayY - archSpringY) * 1.4
			} else {
				// Fluted stone pier column
				uPier := (math.Abs(bayX) - archOpenW) / (0.5*colPeriod - archOpenW)
				nx = math.Copysign(0.75*(1.0-uPier), -bayX) + 0.22*stone.NormalMod[0]
				ny = 0.22 * stone.NormalMod[1]
			}

			nz := math.Sqrt(math.Max(0.12, 1.0-common.Clamp01(nx*nx+ny*ny)*0.85))
			n := moebiusNorm3(nx, ny, nz)

			lit := common.Clamp01(0.16 + 0.80*math.Max(0.0, moebiusDot3(n, lightDir)) + 0.28*sunbeamIntensity)
			sCol := sunStone
			if isEntablature {
				sCol = sunGold
			} else if zLayer >= 2 {
				sCol = moebiusLerp3(sunStone, sunbeamMist, 0.35)
			}

			pxIdx := int(ux * float64(w))
			pyIdx := int(uy * float64(h))
			if pxIdx >= 0 && pxIdx < w && pyIdx >= 0 && pyIdx < h {
				heightBuf[pyIdx*w+pxIdx] = math.Hypot(bayX, bayY) + 0.2*stone.ColorMod
				illumBuf[pyIdx*w+pxIdx] = lit
			}

			return render.MoebiusPixel{
				Depth:           depthVal - 0.18*nz,
				ObjID:           20 + zLayer,
				PartID:          part,
				DetailID:        stone.BlockID,
				Normal:          n,
				FlowTangent:     [2]float64{0.96, 0.24 * nx},
				Illum:           lit,
				Spec:            common.Smoothstep(0.75, 0.96, lit) * 0.60,
				Bounce:          0.38,
				Atmos:           atmos,
				SunColor:        sCol,
				MidColor:        midSienna,
				ShadeColor:      shadeUmber,
				DeepShadowColor: deepCarceri,
				AtmosColor:      sunbeamMist,
				CrackInk:        stone.CrackInk * (1.0 - 0.18*fz),
				TickMask:        stone.PockInk * (1.0 - 0.18*fz),
			}
		}

		// Background Glowing Oculi & Sunbeam Vault Haze
		bgLit := common.Clamp01(0.32 + 0.65*sunbeamIntensity + 0.25*(1.0-uy))
		bgCol := moebiusLerp3(shadeUmber, sunbeamMist, bgLit)
		return render.MoebiusPixel{
			Depth:           20.0,
			ObjID:           2,
			PartID:          2,
			DetailID:        2,
			Normal:          [3]float64{0, 0, 1},
			FlowTangent:     [2]float64{0.62, -0.78},
			Illum:           bgLit,
			SunColor:        bgCol,
			MidColor:        midSienna,
			ShadeColor:      shadeUmber,
			DeepShadowColor: deepCarceri,
			Stipple:         (1.0 - sunbeamIntensity) * 0.42,
		}
	})

	// Overlay Hessian Principal-Curvature Copperplate Burin Cross-Hatching!
	engCfg := render.DefaultEngravingConfig()
	engCfg.PrimaryDSep = 6.5
	engCfg.SecondaryDSep = 8.0
	engCfg.MaxStrokeWidth = 1.15
	engCfg.MinStrokeWidth = 0.28
	engCfg.LightCutoff = 0.52
	engCfg.CrossHatchCutoff = 0.28
	engCfg.EnableChalk = true

	mPx := 0.032 * math.Min(float64(w), float64(h))
	render.RenderCurvatureEngraving(rc, func(x, y float64) float64 {
		ix, iy := int(x), int(y)
		if ix < 0 {
			ix = 0
		} else if ix >= w {
			ix = w - 1
		}
		if iy < 0 {
			iy = 0
		} else if iy >= h {
			iy = h - 1
		}
		return heightBuf[iy*w+ix]
	}, func(x, y float64) float64 {
		ix, iy := int(x), int(y)
		if ix < 0 || ix >= w || iy < 0 || iy >= h {
			return 1.0
		}
		return illumBuf[iy*w+ix]
	}, func(x, y float64) bool {
		return x > mPx && x < float64(w)-mPx && y > mPx && y < float64(h)-mPx
	}, engCfg)

	c.MergeRenderCanvas(rc)
}
