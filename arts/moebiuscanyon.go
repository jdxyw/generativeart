package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

func moebiusLinRGB(r, g, b uint8) [3]float64 {
	lr, lg, lb, _ := render.ColorToLinearRGB(color.RGBA{R: r, G: g, B: b, A: 255})
	return [3]float64{lr, lg, lb}
}

func moebiusLerp3(a, b [3]float64, t float64) [3]float64 {
	t = common.Clamp01(t)
	return [3]float64{
		a[0] + (b[0]-a[0])*t,
		a[1] + (b[1]-a[1])*t,
		a[2] + (b[2]-a[2])*t,
	}
}

func moebiusNorm3(x, y, z float64) [3]float64 {
	l := math.Sqrt(x*x + y*x*0 + y*y + z*z)
	if l < 1e-9 {
		return [3]float64{0, 0, 1}
	}
	inv := 1.0 / l
	return [3]float64{x * inv, y * inv, z * inv}
}

func moebiusDot3(a, b [3]float64) float64 {
	return a[0]*b[0] + a[1]*b[1] + a[2]*b[2]
}

func moebiusBoolFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0.0
}

func moebiusRamp8(stops [8][3]float64, u float64) [3]float64 {
	t := common.Clamp01(u) * 7.0
	idx := int(math.Floor(t))
	if idx >= 7 {
		return stops[7]
	}
	frac := t - float64(idx)
	return moebiusLerp3(stops[idx], stops[idx+1], common.Smoothstep(0.0, 1.0, frac))
}

type moebiusCanyon struct {
	spires       int
	strataLevels int
	monolithTilt float64
}

// NewMoebiusCanyon returns a MoebiusCanyon generator directly inspired by Jean
// "Mœbius" Giraud's 1975 masterpiece Arzach: an asymmetric, craggy petrified rock
// citadel with crenellated summit turrets, a diagonal flying-buttress stone bridge,
// organic cellular boulder fissures, a continuous lemon-gold to vermilion-red to
// indigo-black gouache gradient, and the iconic white Ptéroïde glider soaring
// across the chasm.
func NewMoebiusCanyon(spires, strataLevels int, monolithTilt float64) *moebiusCanyon {
	if spires <= 0 {
		spires = 7
	}
	if strataLevels <= 0 {
		strataLevels = 18
	}
	if monolithTilt == 0 {
		monolithTilt = 0.16
	}
	return &moebiusCanyon{
		spires:       spires,
		strataLevels: strataLevels,
		monolithTilt: monolithTilt,
	}
}

// Generative renders the static MoebiusCanyon (Arzach homage) print.
func (mc *moebiusCanyon) Generative(c *generativeart.Canva) {
	mc.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless looping MoebiusCanyon animation.
func (mc *moebiusCanyon) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	seed := c.Seed()
	if seed == 0 {
		seed = 20262101
	}
	ne := common.NewNoiseEngine(seed)
	phi := 2.0 * math.Pi * phase

	// Authentic Arzach (1975) Gouache & India-Ink Palette
	skyNightTop := moebiusLinRGB(0x0a, 0x16, 0x2a)   // #0a162a Deep Nocturnal Prussian Sky
	skyMistBand := moebiusLinRGB(0x1e, 0x46, 0x78)   // #1e4678 Cobalt-Cerulean Mist Band
	skyAbyss := moebiusLinRGB(0x0c, 0x12, 0x22)      // #0c1222 Abyssal Chasm Night

	peakGold := moebiusLinRGB(0xfc, 0xe4, 0x4a)      // #fce44a Blazing Lemon-Gold Turret Peak
	upperAmber := moebiusLinRGB(0xf6, 0x98, 0x28)    // #f69828 Fiery Tangerine-Amber Upper Cliff
	midVermilion := moebiusLinRGB(0xe0, 0x42, 0x24)  // #e04224 Intense Arzach Vermilion-Red
	lowerSienna := moebiusLinRGB(0x9a, 0x2c, 0x22)   // #9a2c22 Deep Burnt Sienna-Crimson
	baseUmber := moebiusLinRGB(0x48, 0x22, 0x26)     // #482226 Dark Earthy Umber-garnet
	shadowIndigo := moebiusLinRGB(0x24, 0x28, 0x48)  // #242848 Indigo-Slate Penumbra
	deepChiaroscuro := moebiusLinRGB(0x0e, 0x11, 0x1e)// #0e111e Velvety Indigo-Black Shadow Core

	pteroWhite := moebiusLinRGB(0xfc, 0xfa, 0xf2)    // #fcfaf2 Sunlit Ivory-White Ptéroïde Wing
	pteroCyan := moebiusLinRGB(0x68, 0xb8, 0xd8)     // #68b8d8 Cerulean Wing Underside Shadow
	capeGreen := moebiusLinRGB(0x32, 0xa8, 0x74)     // #32a874 Arzach Emerald-Green Cloak

	lightDir := moebiusNorm3(-0.68, -0.42, 0.60)

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)
	cfg := render.DefaultMoebiusConfig(seed)
	cfg.WarmBounceColor = color.RGBA{R: 0xd8, G: 0x42, B: 0x24, A: 255}

	// Helper to shade a craggy Arzach rock structure with continuous vertical
	// fire-to-shadow gouache gradient + domain-warped cellular boulder fissures!
	shadeArzachRock := func(
		x, y, centerX, halfW, topY, depth float64,
		objID int,
		stoneFreqX, stoneFreqY float64,
	) render.MoebiusPixel {
		uHoriz := (x - centerX) / math.Max(halfW, 1e-4) // in [-1, 1] left-to-right across tower
		distFromTop := math.Max(0.0, y-topY)

		// Evaluate weathered Voronoi boulder blocks, organic rock fissures & 3D pillow normals
		stone := render.EvaluateMoebiusStoneRelief(ne, x, y, stoneFreqX, stoneFreqY, 0.42)

		// Macro cylindrical/buttress 3D normal + individual boulder pillow normal
		macroNx := common.Constrain(uHoriz*0.82+0.16*stone.NormalMod[0], -0.95, 0.95)
		macroNy := common.Constrain(-0.28+0.24*stone.NormalMod[1], -0.85, 0.85)
		if distFromTop < 0.045 {
			// Sunlit rounded turret crown at the top of each pinnacle
			macroNy = -0.78
			macroNx *= 0.55
		}
		macroNz := math.Sqrt(math.Max(0.10, 1.0-common.Clamp01(macroNx*macroNx+macroNy*macroNy)*0.88))
		n := moebiusNorm3(macroNx, macroNy, macroNz)

		// Illumination: strong top-left key sun, plunging into deep chiaroscuro on the right (uHoriz > 0.18)
		// and toward the bottom of the canyon (y > 0.55)
		ndotl := math.Max(0.0, moebiusDot3(n, lightDir))
		heightLight := common.Smoothstep(0.92, -0.45, y) // brighter near the summit, darker in the chasm
		lit := common.Clamp01((0.12 + 0.88*ndotl) * (0.42 + 0.58*heightLight))
		if uHoriz > 0.25 {
			lit *= common.Smoothstep(0.95, 0.15, uHoriz)
		}

		// Continuous Arzach Vertical Fire-to-Umber Gouache Gradient (NEVER striped!)
		// Summit (distFromTop ~ 0): Lemon-Gold (#fce44a) -> Upper: Tangerine (#f69828) ->
		// Middle: Vermilion-Red (#e04224) -> Lower: Burnt Sienna (#9a2c22) -> Base: Umber (#482226)
		vGrad := common.Clamp01((y - (-0.58)) / 1.48 + 0.05*stone.ColorMod)
		var sunC, midC [3]float64
		switch {
		case vGrad < 0.22:
			sunC = moebiusLerp3(peakGold, upperAmber, vGrad/0.22)
			midC = upperAmber
		case vGrad < 0.52:
			sunC = moebiusLerp3(upperAmber, midVermilion, (vGrad-0.22)/0.30)
			midC = moebiusLerp3(midVermilion, lowerSienna, 0.45)
		case vGrad < 0.80:
			sunC = moebiusLerp3(midVermilion, lowerSienna, (vGrad-0.52)/0.28)
			midC = lowerSienna
		default:
			sunC = moebiusLerp3(lowerSienna, baseUmber, (vGrad-0.80)/0.20)
			midC = baseUmber
		}

		// Horizontal Mœbius contour hachures wrap around the cylindrical rock face
		// (Use organic warped PartID so there are never straight ruler lines!)
		organicBand := int(math.Floor(distFromTop*3.2 + 0.85*stone.ColorMod))
		return render.MoebiusPixel{
			Depth:           depth - 0.25*macroNz,
			ObjID:           objID,
			PartID:          objID*100 + (organicBand & 15),
			DetailID:        stone.BlockID,
			Normal:          n,
			FlowTangent:     [2]float64{0.96, 0.24 * uHoriz},
			Illum:           lit,
			Spec:            common.Smoothstep(0.05, 0.0, distFromTop) * common.Smoothstep(0.2, -0.5, uHoriz) * 0.65,
			Bounce:          common.Clamp01(0.35 * (1.0 - math.Abs(uHoriz))),
			SunColor:        sunC,
			MidColor:        midC,
			ShadeColor:      shadowIndigo,
			DeepShadowColor: deepChiaroscuro,
			CrackInk:        stone.CrackInk,
			TickMask:        stone.PockInk,
			Stipple:         common.Smoothstep(0.20, 0.0, distFromTop) * 0.42,
		}
	}

	render.RenderMoebiusScene(rc, cfg, func(ux, uy float64) render.MoebiusPixel {
		x := ux*2.0 - 1.0
		y := uy*2.0 - 1.0

		// Organic FBM edge perturbation so every rock silhouette is craggy and hand-drawn
		crag1 := ne.FBM2D(x*4.5+1.3, y*4.5-0.8, 4, 2.0, 0.5)
		crag2 := ne.FBM2D(x*9.0-2.7, y*9.0+1.9, 3, 2.0, 0.5)
		edgeWobble := 0.026*crag1 + 0.010*crag2

		// =====================================================================
		// 1. TWO SOARING WHITE PTÉROÏDE GLIDERS (Articulated Wing Flap + 3D Banking)
		// =====================================================================
		// Helper to evaluate a banking, articulated flapping Ptéroïde glider at (gx, gy),
		// scaleX, scaleY, banking angle bankRad, and wing-flap phase flapPhi.
		evalPteroide := func(
			gx, gy, scaleX, scaleY, bankRad, flapPhi, depth float64,
			objID int,
			cowlCol [3]float64,
		) (render.MoebiusPixel, bool) {
			dx := x - gx
			dy := y - gy
			cosB, sinB := math.Cos(bankRad), math.Sin(bankRad)
			px := (dx*cosB + dy*sinB) / scaleX
			py := (-dx*sinB + dy*cosB) / scaleY

			if math.Abs(px) > 1.08 || math.Abs(py) > 1.15 {
				return render.MoebiusPixel{}, false
			}
			absPx := math.Abs(px)

			// Articulated pterosaur wing kinematics:
			// Shoulder joint leads with sin(flapPhi), outer wingtip whips with phase lag sin(flapPhi - 0.95)!
			shoulderFlap := 0.24 * math.Sin(flapPhi) * absPx
			tipWhip := 0.38 * math.Sin(flapPhi-0.95) * math.Pow(absPx, 1.85)
			wingArch := -0.18*math.Sin(absPx*math.Pi*1.12) + 0.20*absPx*absPx - (shoulderFlap + tipWhip)
			wingHalfThick := 0.135 * (1.0 - 0.82*math.Pow(absPx, 1.3)) * (1.0 + 0.24*math.Cos(absPx*math.Pi*2.0))

			// Body bobs slightly in counter-reaction to the wing downstroke
			bodyBob := 0.08 * math.Sin(flapPhi)
			pyBody := py - bodyBob

			// Tapered teardrop fuselage, pointed head crest, and trailing tail streamer
			bodyHalfW := 0.060 * (1.0 - 0.84*math.Pow(math.Abs(pyBody-0.04)/0.48, 1.35))
			inBody := pyBody > -0.38 && pyBody < 0.46 && absPx < math.Max(0.012, bodyHalfW)
			inRiderCape := absPx < 0.16 && pyBody >= -0.46 && pyBody <= -0.10 &&
				math.Abs(pyBody-(-0.25+0.42*absPx)) < 0.10*(1.0-absPx/0.16)
			inWing := absPx <= 1.0 && math.Abs(py-wingArch) <= wingHalfThick

			if inRiderCape {
				lit := 0.88
				if px > 0.04 {
					lit = 0.42
				}
				return render.MoebiusPixel{
					Depth:           depth - 0.05,
					ObjID:           objID,
					PartID:          objID*10 + 1,
					DetailID:        int(absPx * 18.0),
					Normal:          moebiusNorm3(px*2.0, -0.5, 0.8),
					FlowTangent:     [2]float64{cosB, -sinB},
					Illum:           lit,
					Spec:            0.50,
					SunColor:        cowlCol,
					MidColor:        skyMistBand,
					ShadeColor:      shadowIndigo,
					DeepShadowColor: deepChiaroscuro,
				}, true
			}

			if inBody || inWing {
				uWing := (py - wingArch) / math.Max(wingHalfThick, 1e-3)
				// Wing surface normal tilts with the flap angle so sunlit sheen shifts dynamically!
				flapSlope := 0.25 * math.Cos(flapPhi)
				lit := common.Clamp01(0.86 - 0.38*common.Smoothstep(0.15, 0.65, uWing) + flapSlope)
				if inBody && px > 0.02 {
					lit = 0.42
				}
				part := objID * 10
				if absPx > 0.08 {
					part = objID*10 + 2 + int(moebiusBoolFloat(px > 0))
				}
				return render.MoebiusPixel{
					Depth:           depth,
					ObjID:           objID,
					PartID:          part,
					DetailID:        int(absPx * 12.0),
					Normal:          moebiusNorm3(0.25*px, uWing*0.65-flapSlope, 0.72),
					FlowTangent:     [2]float64{cosB, -sinB},
					Illum:           lit,
					Spec:            0.88,
					SunColor:        pteroWhite,
					MidColor:        pteroCyan,
					ShadeColor:      pteroCyan,
					DeepShadowColor: shadowIndigo,
				}, true
			}
			return render.MoebiusPixel{}, false
		}

		// Primary Foreground Ptéroïde #1: sweeping wide across the lower-middle chasm & cliff
		glide1X := -0.16 + 0.16*math.Sin(phi)
		glide1Y := 0.26 - 0.085*math.Cos(phi) + 0.032*math.Sin(2.0*phi)
		bank1 := -0.22 * math.Cos(phi)
		flap1 := 2.0*phi + 0.35 // 2 full wing-flap cycles per orbital loop!
		if smp, ok := evalPteroide(glide1X, glide1Y, 0.29, 0.145, bank1, flap1, 1.25, 60, capeGreen); ok {
			return smp
		}

		// Secondary Companion Ptéroïde #2: soaring higher near the stone bridge & right bastion
		glide2X := 0.22 - 0.15*math.Cos(phi)
		glide2Y := -0.05 + 0.075*math.Sin(phi) - 0.028*math.Cos(2.0*phi)
		bank2 := 0.20 * math.Sin(phi)
		flap2 := 2.0*phi + 2.45
		if smp, ok := evalPteroide(glide2X, glide2Y, 0.21, 0.105, bank2, flap2, 1.45, 61, upperAmber); ok {
			return smp
		}

		// =====================================================================
		// 2. FOREGROUND LOWER-LEFT & LOWER-RIGHT CRAGGY BOULDER OUTCROPS
		// =====================================================================
		// Lower-center-right foreground diagonal rock shoulder (like bottom of Arzach!)
		fgShoulderTop := 0.56 - 0.24*(x-0.12) + 0.08*math.Sin((x-0.1)*6.0) + edgeWobble*1.2
		if x >= -0.26+edgeWobble && x <= 0.62+edgeWobble && y >= fgShoulderTop {
			smp := shadeArzachRock(x, y, 0.22, 0.38, fgShoulderTop, 1.8, 15, 7.5, 7.5)
			// Foreground boulder is in deeper twilight umber/sienna contrast
			smp.Illum *= 0.68
			return smp
		}

		// Lower-left dark craggy sentinel rock pinnacle (x < -0.56, y > 0.28)
		fgLeftTop := 0.34 + 0.75*(x+0.88)*(x+0.88) + edgeWobble*1.3
		if x <= -0.52+edgeWobble && y >= fgLeftTop {
			smp := shadeArzachRock(x, y, -0.78, 0.24, fgLeftTop, 1.9, 16, 8.0, 8.0)
			smp.Illum *= 0.55
			return smp
		}

		// Lower-right stepped pyramid ruin corner (x > 0.68, y > 0.64)
		fgRightTop := 0.66 + 0.55*math.Abs(x-0.86) + edgeWobble*0.8
		if x >= 0.64+edgeWobble && y >= fgRightTop {
			smp := shadeArzachRock(x, y, 0.84, 0.18, fgRightTop, 1.7, 17, 9.0, 12.0)
			smp.Illum *= 0.62
			return smp
		}

		// =====================================================================
		// 3. RIGHT ROUNDED BASTION ROCK TOWER (Centered at x = 0.34, top at y = -0.28)
		// =====================================================================
		rCenter := 0.34 - 0.04*(y-0.1)
		rHalfW := 0.165 + 0.055*(y+0.28) + edgeWobble*0.85
		rTopY := -0.28 + 1.8*(x-0.34)*(x-0.34) + edgeWobble
		if y >= rTopY && math.Abs(x-rCenter) <= rHalfW {
			return shadeArzachRock(x, y, rCenter, rHalfW, rTopY, 2.4, 22, 7.0, 7.0)
		}

		// =====================================================================
		// 4. DIAGONAL STONE FLYING-BUTTRESS BRIDGE (Connecting Main & Right Towers)
		// =====================================================================
		if x >= -0.06 && x <= 0.32 {
			bridgeT := (x - (-0.06)) / 0.38 // 0 at main tower, 1 at right bastion
			bridgeTopY := -0.04 - 0.19*bridgeT + 0.015*math.Sin(bridgeT*math.Pi*3.0) + edgeWobble*0.6
			// Underside of the bridge arches deeply into a shadowy stone vault
			bridgeBotY := 0.24 - 0.10*bridgeT - 0.09*math.Sin(bridgeT*math.Pi) + edgeWobble*0.6
			if y >= bridgeTopY && y <= bridgeBotY {
				smp := shadeArzachRock(x, y, 0.13, 0.22, bridgeTopY, 2.8, 21, 9.0, 9.0)
				// Cast shadow under the lower vault of the bridge
				uBridgeV := (y - bridgeTopY) / math.Max(bridgeBotY-bridgeTopY, 1e-3)
				if uBridgeV > 0.55 {
					smp.Illum *= 0.38
				}
				return smp
			}
		}

		// =====================================================================
		// 5. COLOSSAL MAIN ARZACH CITADEL ROCK TOWER (Centered at x = -0.16)
		// =====================================================================
		mCenter := -0.15 - 0.03*y
		mHalfW := 0.28 + 0.065*(y+0.4) + edgeWobble
		if math.Abs(x-mCenter) <= mHalfW {
			uTower := (x - mCenter) / mHalfW // in [-1, 1] across main tower width
			// Asymmetric 3-turret crenellated summit silhouette (matching Arzach!):
			//   - Left narrow sentinel finger (uTower in [-1.0, -0.68], top ~ -0.46)
			//   - Central plateau & towering right-center horn pinnacle (uTower in [-0.62, 0.58], peak ~ -0.66 at uTower=0.22)
			//   - Outer right turret projection (uTower in [0.64, 1.0], top ~ -0.32)
			var mTopY float64
			inNotch := false
			switch {
			case uTower < -0.68:
				// Left vertical rock finger
				uFinger := (uTower - (-0.84)) / 0.16
				mTopY = -0.46 + 0.14*uFinger*uFinger
			case uTower < -0.56:
				// Narrow cleft notch between left finger and main citadel crown
				mTopY = -0.22
				inNotch = y < -0.22+edgeWobble*0.5
			case uTower < 0.62:
				// Sloping citadel crown rising to the tall central horn at uTower = 0.24
				hornPeak := 0.22 * math.Exp(-math.Pow((uTower-0.24)/0.22, 2.0))
				mTopY = -0.45 - hornPeak + 0.05*uTower
			case uTower < 0.72:
				// Cleft notch before outer right turret
				mTopY = -0.18
				inNotch = y < -0.18+edgeWobble*0.5
			default:
				// Outer right turret projection (y in [-0.34, -0.06])
				mTopY = -0.33 + 0.12*math.Pow((uTower-0.84)/0.16, 2.0)
				if y > -0.04+edgeWobble {
					inNotch = true // overhanging turret corbel drops off below y = -0.04!
				}
			}
			mTopY += edgeWobble * 0.85

			if !inNotch && y >= mTopY {
				// Recessed dark vertical window/fissure slot on the sunlit left-center face!
				inWindowSlot := math.Abs(uTower-(-0.24)) < 0.085 && y >= -0.26 && y <= -0.04
				if inWindowSlot {
					// Ancient cross-bars inside the window slot
					barWave := math.Sin(y * 85.0)
					isBar := barWave > 0.55
					return render.MoebiusPixel{
						Depth:           3.3,
						ObjID:           20,
						PartID:          2099,
						DetailID:        int(y * 40.0),
						Normal:          [3]float64{0, 0, 1},
						FlowTangent:     [2]float64{1, 0},
						Illum:           moebiusBoolFloat(isBar) * 0.65,
						SunColor:        upperAmber,
						MidColor:        lowerSienna,
						ShadeColor:      deepChiaroscuro,
						DeepShadowColor: deepChiaroscuro,
					}
				}

				smp := shadeArzachRock(x, y, mCenter, mHalfW, mTopY, 3.0, 20, 6.8, 6.8)
				// Smoothly cast shadow from the tall central horn onto the right slope below it
				hornShadow := common.Smoothstep(0.02, 0.36, uTower) * common.Smoothstep(0.04, -0.22, y)
				smp.Illum *= 1.0 - 0.54*hornShadow
				return smp
			}
		}

		// =====================================================================
		// 6. BROODING NOCTURNAL ARZACH SKY, COBALT MIST BAND & PHANTOM MOON
		// =====================================================================
		// Dim cratered phantom moon in the upper-left mist ((-0.66, -0.32), radius 0.062)
		dxMoon := x - (-0.66)
		dyMoon := y - (-0.32)
		rMoon := math.Hypot(dxMoon, dyMoon)
		if rMoon <= 0.062 {
			stoneM := render.EvaluateMoebiusStoneRelief(ne, x*2.0, y*2.0, 12.0, 12.0, 0.3)
			return render.MoebiusPixel{
				Depth:           75.0,
				ObjID:           40,
				PartID:          400,
				DetailID:        stoneM.BlockID,
				Normal:          moebiusNorm3(dxMoon/0.062, dyMoon/0.062, 0.5),
				FlowTangent:     [2]float64{1, 0},
				Illum:           0.22 + 0.18*common.Smoothstep(0.06, -0.06, dxMoon),
				SunColor:        skyMistBand,
				MidColor:        skyMistBand,
				ShadeColor:      skyNightTop,
				DeepShadowColor: skyAbyss,
				CrackInk:        stoneM.CrackInk * 0.5,
			}
		}

		// Glowing nocturnal cobalt-cerulean mist band across y in [-0.46, -0.10]
		mistCenter := -0.28 + 0.06*crag1 + 0.025*math.Sin(x*3.2-phi)
		mistDist := math.Abs(y - mistCenter)
		mistWave := 0.55 + 0.45*math.Sin(x*4.0+2.5*crag1-phi)
		mistGlow := common.Smoothstep(0.32, 0.0, mistDist) * mistWave

		skyCol := moebiusLerp3(skyNightTop, skyMistBand, mistGlow)
		if y > 0.10 {
			skyCol = moebiusLerp3(skyCol, skyAbyss, common.Smoothstep(0.10, 0.75, y))
		}

		return render.MoebiusPixel{
			Depth:           100.0,
			ObjID:           1,
			PartID:          1,
			DetailID:        1,
			Normal:          [3]float64{0, 0, 1},
			FlowTangent:     [2]float64{1, 0},
			Illum:           0.75,
			SunColor:        skyCol,
			MidColor:        skyNightTop,
			ShadeColor:      skyAbyss,
			DeepShadowColor: skyAbyss,
			Stipple:         mistGlow * 0.36,
		}
	})

	c.MergeRenderCanvas(rc)
}
