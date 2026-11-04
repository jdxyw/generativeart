package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type mineralShanShui struct {
	peaks           int
	mistLayers      int
	mineralRichness float64
}

// NewMineralShanShui returns a MineralShanShui generator that fuses Song-Dynasty
// Blue-and-Green Mineral Landscape painting (Qinglü Shanshui — Wang Ximeng's
// A Thousand Li of Rivers and Mountains) with Jean "Mœbius" Giraud's Ligne Claire
// rock relief: 4 receding karst mountain ranges glazed with physical Kubelka-Munk
// Azurite (Shiqing), Malachite (Shilv), Yellow Ochre (Zheshi), Cinnabar (Zhusha),
// and 24k Mud-Gold (Nijin), complete with cold-pressed mineral granulation,
// axe-cut Cunfa fissures, soft valley mist, and an emerald river skiff.
func NewMineralShanShui(peaks, mistLayers int, mineralRichness float64) *mineralShanShui {
	if peaks <= 0 {
		peaks = 5
	}
	if mistLayers <= 0 {
		mistLayers = 4
	}
	if mineralRichness <= 0 {
		mineralRichness = 1.0
	}
	return &mineralShanShui{
		peaks:           peaks,
		mistLayers:      mistLayers,
		mineralRichness: mineralRichness,
	}
}

// Generative renders the static MineralShanShui mineral scroll print.
func (ms *mineralShanShui) Generative(c *generativeart.Canva) {
	ms.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless looping MineralShanShui animation.
func (ms *mineralShanShui) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	seed := c.Seed()
	if seed == 0 {
		seed = 20262204
	}
	ne := common.NewNoiseEngine(seed)
	phi := 2.0 * math.Pi * phase

	// Authentic Wang Ximeng Qianli Jiangshan (千里江山图) Kubelka-Munk Mineral Glazes
	agedSilk := moebiusLinRGB(0xf4, 0xe6, 0xc8)
	zheshiFoot := render.MixKubelkaMunk(render.PW(render.PigmentYellowOchre, 0.72), render.PW(render.PigmentBurntSienna, 0.14), render.PW(render.PigmentTitaniumWhite, 0.14))
	shilvBright := render.MixKubelkaMunk(render.PW(render.PigmentMalachite, 0.62), render.PW(render.PigmentCobaltTeal, 0.24), render.PW(render.PigmentCadmiumYellow, 0.14))
	shilvDeep := render.MixKubelkaMunk(render.PW(render.PigmentMalachite, 0.78), render.PW(render.PigmentAzurite, 0.22))
	shiqingBright := render.MixKubelkaMunk(render.PW(render.PigmentAzurite, 0.65), render.PW(render.PigmentCerulean, 0.35))
	shiqingSummit := render.MixKubelkaMunk(render.PW(render.PigmentUltramarine, 0.58), render.PW(render.PigmentAzurite, 0.42))
	nijinGold := render.MixKubelkaMunk(render.PW(render.PigmentGoldLeaf, 0.85), render.PW(render.PigmentCadmiumYellow, 0.15))
	cinnabarSun := render.MixKubelkaMunk(render.PW(render.PigmentVermilion, 0.82), render.PW(render.PigmentYellowOchre, 0.18))

	// Shadow colors stay rich mineral indigo-emerald (never muddy grey-purple!)
	mineralShadow := render.MixKubelkaMunk(render.PW(render.PigmentAzurite, 0.55), render.PW(render.PigmentMalachite, 0.30), render.PW(render.PigmentPineSootInk, 0.15))
	mineralDeep := render.MixKubelkaMunk(render.PW(render.PigmentPineSootInk, 0.55), render.PW(render.PigmentUltramarine, 0.45))

	waterJade := render.GlazeKubelkaMunk(agedSilk, render.PigmentCobaltTeal, 0.48)
	waterMalachite := render.GlazeKubelkaMunk(waterJade, render.PigmentMalachite, 0.62)

	lightDir := moebiusNorm3(-0.58, -0.48, 0.66)

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)
	cfg := render.DefaultMoebiusConfig(seed)
	cfg.PaperMatColor = color.RGBA{R: 0xf4, G: 0xea, B: 0xd2, A: 255}
	cfg.InnerRuleColor = color.RGBA{R: 0xc8, G: 0x42, B: 0x2c, A: 255}
	cfg.WarmBounceColor = color.RGBA{R: 0xe8, G: 0xae, B: 0x48, A: 255}

	granStrength := 0.28 * common.Constrain(ms.mineralRichness, 0.5, 1.5)

	render.RenderMoebiusScene(rc, cfg, func(ux, uy float64) render.MoebiusPixel {
		x := ux*2.0 - 1.0
		y := uy*2.0 - 1.0

		// =====================================================================
		// 1. 4 RECEDING KARST MOUNTAIN RANGES WITH KUBELKA-MUNK MINERAL GLAZES
		// =====================================================================
		type shanLayer struct {
			id     int
			depth  float64
			baseY  float64
			peak1X float64
			peak1W float64
			peak1H float64
			peak2X float64
			peak2W float64
			peak2H float64
		}
		ranges := [4]shanLayer{
			// Range 0: Foreground Left & Right Promontory Cliffs framing the Emerald River
			{id: 10, depth: 2.0, baseY: 0.68, peak1X: -0.68, peak1W: 0.30, peak1H: 0.34, peak2X: 0.72, peak2W: 0.28, peak2H: 0.40},
			// Range 1: Midground Towering Sovereign Peak (Main Host Mountain 主峰 at x = 0.24)
			{id: 11, depth: 3.6, baseY: 0.48, peak1X: 0.24, peak1W: 0.32, peak1H: 0.82, peak2X: -0.32, peak2W: 0.24, peak2H: 0.48},
			// Range 2: Secondary Escort Peaks (客山)
			{id: 12, depth: 5.4, baseY: 0.30, peak1X: -0.58, peak1W: 0.26, peak1H: 0.58, peak2X: 0.66, peak2W: 0.25, peak2H: 0.54},
			// Range 3: Far Horizon Azurite Peaks
			{id: 13, depth: 7.4, baseY: 0.14, peak1X: -0.06, peak1W: 0.30, peak1H: 0.44, peak2X: -0.84, peak2W: 0.22, peak2H: 0.34},
		}

		for rIdx := range ranges {
			rg := &ranges[rIdx]
			fr := float64(rIdx)

			fbm, dfbmDx, _ := ne.FBM2DDeriv(x*(2.4+0.4*fr)+fr*3.1, fr*2.7, 4, 2.0, 0.5)
			ridgeY := rg.baseY - 0.08*fbm
			slopeDx := -0.08 * dfbmDx

			d1 := (x - rg.peak1X) / rg.peak1W
			if math.Abs(d1) < 1.55 {
				shoulder := 0.28 * math.Exp(-math.Pow((d1+0.48)/0.32, 2.0))
				mainPeak := math.Exp(-math.Pow(d1/0.58, 2.0)) * (1.0 + 0.12*math.Cos(d1*5.0))
				ridgeY -= rg.peak1H * (mainPeak + shoulder)
				slopeDx += rg.peak1H * (2.0 * d1 / (0.58 * 0.58 * rg.peak1W)) * mainPeak
			}

			d2 := (x - rg.peak2X) / rg.peak2W
			if math.Abs(d2) < 1.55 {
				mainPeak2 := math.Exp(-math.Pow(d2/0.60, 2.0)) * (1.0 + 0.10*math.Sin(d2*4.5))
				ridgeY -= rg.peak2H * mainPeak2
				slopeDx += rg.peak2H * (2.0 * d2 / (0.60 * 0.60 * rg.peak2W)) * mainPeak2
			}

			if y >= ridgeY {
				dyBelow := y - ridgeY
				maxH := rg.peak1H + 0.12
				// altT: 1.0 at the summit ridge crest, 0.0 at the foot of the mountain
				altT := common.Clamp01(1.0 - dyBelow/math.Max(maxH*0.65, 0.18))

				// Soft, outline-free Valley Mist Veil (*Yunqi* 云气) at the base of each range!
				mistBandY := rg.baseY - 0.03 + 0.025*math.Sin(x*3.8+fr*1.7-phi)
				mistDist := math.Abs(y - mistBandY)
				mistVeil := common.Smoothstep(0.14, 0.0, mistDist) * common.Smoothstep(0.0, 0.05, dyBelow)
				atmos := common.Clamp01(0.78*mistVeil + 0.12*fr)

				// Axe-Cut *Cunfa* (斧劈皴) vertical/diagonal rock relief
				stone := render.EvaluateMoebiusStoneRelief(ne, x*1.3+0.25*y, y*0.8, 8.5, 6.5, 0.38)
				gullyWave := 0.36 * math.Sin(x*16.0+y*5.5+1.5*fbm)
				nx := common.Constrain(0.42*slopeDx+gullyWave+0.24*stone.NormalMod[0], -0.92, 0.92)
				ny := -0.26 + 0.22*stone.NormalMod[1]
				if dyBelow < 0.026 {
					ny = -0.84
				}
				nz := math.Sqrt(math.Max(0.12, 1.0-common.Clamp01(nx*nx+ny*ny)*0.85))
				n := moebiusNorm3(nx, ny, nz)

				lit := common.Clamp01(0.22 + 0.82*math.Max(0.0, moebiusDot3(n, lightDir)))

				// Authentic 4-Stage Wang Ximeng Mineral Glaze Progression (Foot -> Mid -> Upper -> Ridge):
				//   - Foot (altT < 0.26): Warm *Zheshi* Ochre (#dca258) -> *Shilv* Malachite
				//   - Lower-Mid (0.26..0.56): Luminous *Shilv* Malachite Green (#26a878)
				//   - Upper Slope (0.56..0.84): Rich *Shiqing* Azurite Blue (#1b56a8)
				//   - Summit Crest (altT >= 0.84 or dyBelow < 0.024): Ultramarine + 24k *Nijin* Mud-Gold rim!
				paperGrain := (ne.FBM2D(ux*52.0, uy*52.0, 3, 2.0, 0.5) + 0.15*stone.ColorMod) * granStrength
				effAlt := common.Clamp01(altT + paperGrain)

				var sunC, midC [3]float64
				switch {
				case effAlt < 0.26:
					sunC = moebiusLerp3(zheshiFoot, shilvBright, effAlt/0.26)
					midC = shilvDeep
				case effAlt < 0.58:
					sunC = moebiusLerp3(shilvBright, shiqingBright, (effAlt-0.26)/0.32)
					midC = shilvDeep
				case effAlt < 0.86:
					sunC = moebiusLerp3(shiqingBright, shiqingSummit, (effAlt-0.58)/0.28)
					midC = shiqingBright
				default:
					sunC = moebiusLerp3(shiqingSummit, nijinGold, (effAlt-0.86)/0.14)
					midC = shiqingSummit
				}
				if dyBelow < 0.020 {
					// Crisp 24k *Nijin* (泥金) gold-leaf contour rim along the summit crest!
					sunC = moebiusLerp3(sunC, nijinGold, 0.78)
				}

				organicPart := int(math.Floor(dyBelow*3.8 + 0.85*stone.ColorMod))
				return render.MoebiusPixel{
					Depth:           rg.depth + 0.1*dyBelow,
					ObjID:           rg.id,
					PartID:          rg.id*100 + (organicPart & 15),
					DetailID:        stone.BlockID,
					Normal:          n,
					FlowTangent:     [2]float64{0.38, 0.92},
					Illum:           lit,
					Spec:            common.Smoothstep(0.03, 0.0, dyBelow) * 0.75,
					Bounce:          0.48,
					Atmos:           atmos,
					SunColor:        sunC,
					MidColor:        midC,
					ShadeColor:      mineralShadow,
					DeepShadowColor: mineralDeep,
					AtmosColor:      agedSilk,
					CrackInk:        stone.CrackInk * (1.0 - 0.5*atmos),
					TickMask:        stone.PockInk * (1.0 - 0.5*atmos),
				}
			}
		}

		// =====================================================================
		// 2. EMERALD-JADE RIVER LAGOON & SCHOLAR'S SKIFF (y >= 0.56)
		// =====================================================================
		if y >= 0.56 {
			ty := (y - 0.56) / 0.44
			// Miniature scholar's wooden skiff gliding on the emerald river at (-0.08, 0.76)
			bx := (x - (-0.08 + 0.06*math.Sin(phi))) / 0.12
			by := (y - 0.76) / 0.025
			if math.Abs(bx) <= 1.0 && by >= -0.7 && by <= 1.0 && math.Abs(bx) <= 1.0-0.38*by {
				return render.MoebiusPixel{
					Depth:           2.6,
					ObjID:           60,
					PartID:          600,
					DetailID:        int((bx + 1.0) * 6.0),
					Normal:          [3]float64{0, -0.5, 0.86},
					FlowTangent:     [2]float64{1, 0},
					Illum:           0.86,
					SunColor:        nijinGold,
					MidColor:        cinnabarSun,
					ShadeColor:      mineralShadow,
					DeepShadowColor: mineralDeep,
				}
			}

			ripple := 0.5 + 0.5*math.Sin((ty*18.0+x*6.0-phase)*2.0*math.Pi)
			wCol := moebiusLerp3(waterJade, waterMalachite, 0.25+0.55*ty+0.20*ripple)
			return render.MoebiusPixel{
				Depth:           9.0 - 2.0*ty,
				ObjID:           18,
				PartID:          180,
				DetailID:        180,
				Normal:          [3]float64{0, -0.1, 0.99},
				FlowTangent:     [2]float64{1, 0},
				Illum:           0.86,
				Spec:            common.Smoothstep(0.82, 0.98, ripple) * 0.55,
				SunColor:        wCol,
				MidColor:        waterMalachite,
				ShadeColor:      mineralShadow,
				TickMask:        common.Smoothstep(0.92, 0.98, ripple) * 0.55,
			}
		}

		// =====================================================================
		// 3. IMPERIAL CINNABAR SUN, GOLDEN CLOUDS & AGED SILK SKY
		// =====================================================================
		rSun := math.Hypot(x-(-0.56), y-(-0.54))
		if rSun <= 0.080 {
			return render.MoebiusPixel{
				Depth:      60.0,
				ObjID:      70,
				PartID:     700,
				DetailID:   700,
				Normal:     [3]float64{0, 0, 1},
				Illum:      0.94,
				Spec:       0.45,
				SunColor:   cinnabarSun,
				MidColor:   cinnabarSun,
				ShadeColor: zheshiFoot,
			}
		}

		skyT := common.Clamp01((y + 1.0) / 1.56)
		skyCol := moebiusLerp3(zheshiFoot, agedSilk, common.Smoothstep(0.0, 0.72, skyT))
		return render.MoebiusPixel{
			Depth:       100.0,
			ObjID:       1,
			PartID:      1,
			DetailID:    1,
			Normal:      [3]float64{0, 0, 1},
			FlowTangent: [2]float64{1, 0},
			Illum:       0.88,
			SunColor:    skyCol,
			MidColor:    zheshiFoot,
			ShadeColor:  zheshiFoot,
			Stipple:     (1.0 - skyT) * 0.26,
		}
	})

	c.MergeRenderCanvas(rc)
}
