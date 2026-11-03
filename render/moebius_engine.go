package render

import (
	"image/color"
	"math"
	"math/rand"
	"runtime"
	"sync"

	"github.com/jdxyw/generativeart/common"
)

// MoebiusPixel stores the 3D/2.5D G-Buffer geometry, 3-hierarchy Ligne Claire IDs,
// 3D surface normal, screen-projected curvature tangent, and multi-stop gouache /
// airbrush / chiaroscuro shadow-core shading for a single sub-pixel.
type MoebiusPixel struct {
	Depth           float64    // 3D depth (smaller = closer; >= 90 = sky background)
	ObjID           int        // Hierarchy 1: Primary silhouette object ID (bold hand-inked contour)
	PartID          int        // Hierarchy 2: Major rock fissure / structural module ID (medium crease line)
	DetailID        int        // Hierarchy 3: Weathered stone block / panel / rib ID (delicate hairline ink)
	Normal          [3]float64 // 3D unit surface normal (Nx, Ny, Nz)
	FlowTangent     [2]float64 // 2D screen-space contour tangent (Tx, Ty) for short Rapidograph hachures
	Illum           float64    // Direct illumination in [0, 1] (<0.20 deep gouache core, 0.20..0.55 hachures)
	Spec            float64    // Grazing rim / sunlit crest highlight intensity in [0, 1]
	Bounce          float64    // Warm upward canyon/desert reflected bounce light in [0, 1]
	Atmos           float64    // Atmospheric aerial perspective haze factor in [0, 1]
	SunColor        [3]float64 // Linear RGB color in direct sunlight
	MidColor        [3]float64 // Linear RGB warm terracotta/sienna/ochre midtone transition
	ShadeColor      [3]float64 // Linear RGB cool cerulean/indigo/violet penumbra wash
	DeepShadowColor [3]float64 // Linear RGB deep Prussian-indigo/umber chiaroscuro shadow core
	AtmosColor      [3]float64 // Linear RGB atmospheric aerial-perspective veil color
	CrackInk        float64    // Continuous organic cellular rock fissure / glyph ink intensity [0, 1]
	TickMask        float64    // Fine Mœbius horizontal erosion dash / pebble loop ink [0, 1]
	Stipple         float64    // Pointillist ink stipple density in [0, 1]
}

// MoebiusStoneRelief holds the result of evaluating a domain-warped cellular/Voronoi
// weathered Mœbius rock & ancient masonry surface (as seen in Arzach and Panzer Dragoon).
type MoebiusStoneRelief struct {
	BlockID   int        // Unique weathered boulder / masonry block ID
	CrackInk  float64    // Crisp organic ink fissure along boulder boundaries [0, 1]
	PockInk   float64    // Weathered limestone pockmarks, pebbles & short horizontal erosion ticks [0, 1]
	NormalMod [3]float64 // 3D boulder pillow normal perturbation (dNx, dNy, dNz)
	ColorMod  float64    // Per-boulder subtle mineral tone variation in [-1, 1]
}

// EvaluateMoebiusStoneRelief computes an organic, domain-warped Voronoi boulder &
// horizontal erosion-fissure relief field that transforms smooth mathematical shapes
// into craggy, hand-inked Mœbius rock towers, arches, and ancient carved ruins.
func EvaluateMoebiusStoneRelief(ne *common.NoiseEngine, u, v, freqX, freqY, warpStrength float64) MoebiusStoneRelief {
	if ne == nil {
		return MoebiusStoneRelief{}
	}
	// Multi-octave organic domain warp so stone joints are craggy and hand-drawn
	wx, dWxDu, dWxDv := ne.FBM2DDeriv(u*freqX*0.55+1.7, v*freqY*0.55-2.3, 3, 2.0, 0.5)
	wy, _, _ := ne.FBM2DDeriv(u*freqX*0.55-4.1, v*freqY*0.55+3.9, 3, 2.0, 0.5)

	su := u*freqX + warpStrength*wx
	sv := v*freqY + warpStrength*wy

	// Evaluate 3x3 jittered cellular lattice for irregular boulder blocks
	iu := int(math.Floor(su))
	iv := int(math.Floor(sv))
	fu := su - float64(iu)
	fv := sv - float64(iv)

	minD1 := 1e9
	minD2 := 1e9
	var bestDX, bestDY float64
	var bestID int
	var bestHash float64

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			cx := iu + dx
			cy := iv + dy
			// Deterministic pseudo-random cell feature point
			h1 := math.Sin(float64(cx)*127.1 + float64(cy)*311.7)
			h1 = h1*43758.5453 - math.Floor(h1*43758.5453)
			h2 := math.Sin(float64(cx)*269.5 + float64(cy)*183.3)
			h2 = h2*43758.5453 - math.Floor(h2*43758.5453)

			px := float64(dx) + 0.16 + 0.68*h1
			py := float64(dy) + 0.16 + 0.68*h2
			ddx := fu - px
			ddy := (fv - py) * 1.25 // slightly horizontally elongated boulders
			dist := math.Hypot(ddx, ddy)

			if dist < minD1 {
				minD2 = minD1
				minD1 = dist
				bestDX = ddx
				bestDY = ddy
				bestID = ((cx & 255) << 8) ^ (cy & 255)
				bestHash = h1*2.0 - 1.0
			} else if dist < minD2 {
				minD2 = dist
			}
		}
	}

	// F2 - F1 distance to the nearest boulder fissure edge
	edgeDist := minD2 - minD1
	// Modulate crack width with FBM so fissures pinch and swell like hand-inked rock cracks
	crackWidth := 0.065 * (0.65 + 0.55*math.Max(0.0, wx))
	crackInk := 1.0 - common.Smoothstep(crackWidth*0.35, crackWidth, edgeDist)

	// Add horizontal sedimentary erosion grooves & short Mœbius pebble loops
	horizWave := math.Sin(sv*math.Pi*2.2 + 1.4*wx)
	horizCrack := (1.0 - common.Smoothstep(0.04, 0.14, math.Abs(horizWave))) *
		common.Smoothstep(0.1, 0.6, wy+0.3) * 0.72

	// Small circular pebble/pockmark rings inside boulders
	pebbleR := math.Hypot(bestDX*2.6, bestDY*2.6)
	pockInk := 0.0
	if bestHash > 0.35 && pebbleR > 0.28 && pebbleR < 0.48 {
		pockInk = (1.0 - math.Abs(pebbleR-0.38)/0.10) * 0.58
	}
	pockInk = math.Max(pockInk, horizCrack)

	// 3D rounded boulder pillow normal perturbation (each stone block has its own 3D dome!)
	nxMod := common.Constrain(bestDX*1.15+0.18*dWxDu, -0.75, 0.75)
	nyMod := common.Constrain(bestDY*1.25+0.18*dWxDv, -0.75, 0.75)
	// Near a horizontal fissure, normal tilts downward (overhanging rock lip above, sunlit shelf below)
	if math.Abs(horizWave) < 0.22 {
		nyMod += 0.35 * math.Copysign(1.0, horizWave)
	}

	return MoebiusStoneRelief{
		BlockID:   bestID,
		CrackInk:  crackInk,
		PockInk:   pockInk,
		NormalMod: [3]float64{nxMod, nyMod, 0.0},
		ColorMod:  bestHash,
	}
}

// MoebiusConfig controls the Ligne Claire ink weights, short terminator hachures,
// chiaroscuro gouache shadow cores, and archival framing for RenderMoebiusScene.
type MoebiusConfig struct {
	Seed            int64
	InkColor        color.RGBA
	PaperMatColor   color.RGBA
	InnerRuleColor  color.RGBA
	WarmBounceColor color.RGBA
	BorderMargin    float64
	EnableRK4Hatch  bool
	HatchSpacingPx  float64
}

// DefaultMoebiusConfig returns an authentic Jean "Mœbius" Giraud configuration.
func DefaultMoebiusConfig(seed int64) MoebiusConfig {
	return MoebiusConfig{
		Seed:            seed,
		InkColor:        color.RGBA{R: 0x11, G: 0x0e, B: 0x16, A: 255}, // #110e16 Deep India Ink
		PaperMatColor:   color.RGBA{R: 0xf7, G: 0xf0, B: 0xde, A: 255}, // #f7f0de Warm Archival Cotton Margin
		InnerRuleColor:  color.RGBA{R: 0xd4, G: 0x96, B: 0x48, A: 255}, // #d49648 Golden-Ochre Fillet
		WarmBounceColor: color.RGBA{R: 0xec, G: 0x6e, B: 0x3e, A: 255}, // #ec6e3e Warm Vermilion-Amber Bounce
		BorderMargin:    0.028,
		EnableRK4Hatch:  true,
		HatchSpacingPx:  3.6,
	}
}

func moebLerp3(a, b [3]float64, t float64) [3]float64 {
	t = common.Clamp01(t)
	return [3]float64{
		a[0] + (b[0]-a[0])*t,
		a[1] + (b[1]-a[1])*t,
		a[2] + (b[2]-a[2])*t,
	}
}

// RenderMoebiusScene executes the 4-stage Mœbius Gouache + Ligne Claire + Hachure pipeline:
//  1. 2x2 Supersampled (2W x 2H) 3D G-Buffer evaluation with 5-stage gouache/watercolor
//     tonal ramp (Deep Shadow Core -> Cool Penumbra -> Warm Midtone -> Sunlit -> Rim Specular),
//     upward bounce light, and atmospheric aerial perspective.
//  2. Hand-trembled 3-Hierarchy Ligne Claire ink edge detection + organic rock fissures
//     (CrackInk) and horizontal erosion ticks (TickMask).
//  3. Terminator-clustered short Rapidograph hachures traced via RK4 along each boulder
//     and structure's 3D contour tangent.
//  4. Archival Métal Hurlant serigraph border and cotton paper texture.
func RenderMoebiusScene(rc *Canvas, cfg MoebiusConfig, evalFn func(ux, uy float64) MoebiusPixel) {
	if rc == nil || evalFn == nil {
		return
	}
	w, h := rc.Width, rc.Height
	if w <= 0 || h <= 0 {
		return
	}

	if cfg.BorderMargin <= 0 {
		cfg.BorderMargin = 0.028
	}
	if cfg.HatchSpacingPx <= 1.5 {
		cfg.HatchSpacingPx = 3.6
	}

	sw, sh := w*2, h*2
	fsw, fsh := float64(sw), float64(sh)
	gbuf := make([]MoebiusPixel, sw*sh)

	numWorkers := runtime.NumCPU()
	if numWorkers < 1 {
		numWorkers = 1
	}
	var wg sync.WaitGroup
	rowsPerWorker := (sh + numWorkers - 1) / numWorkers

	// Stage 1: Evaluate 2W x 2H supersampled G-Buffer in parallel
	for worker := 0; worker < numWorkers; worker++ {
		y0 := worker * rowsPerWorker
		y1 := y0 + rowsPerWorker
		if y1 > sh {
			y1 = sh
		}
		if y0 >= y1 {
			continue
		}
		wg.Add(1)
		go func(startY, endY int) {
			defer wg.Done()
			for sy := startY; sy < endY; sy++ {
				uy := (float64(sy) + 0.5) / fsh
				rowOff := sy * sw
				for sx := 0; sx < sw; sx++ {
					ux := (float64(sx) + 0.5) / fsw
					p := evalFn(ux, uy)
					tx, ty := p.FlowTangent[0], p.FlowTangent[1]
					tLen := math.Hypot(tx, ty)
					if tLen < 1e-5 {
						// Default to near-horizontal Mœbius contour hachure angle slightly tilted by normal
						tx = 0.96
						ty = 0.22 * p.Normal[0]
						tLen = math.Hypot(tx, ty)
					}
					p.FlowTangent = [2]float64{tx / tLen, ty / tLen}
					gbuf[rowOff+sx] = p
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	inkR, inkG, inkB, _ := ColorToLinearRGB(cfg.InkColor)
	inkLin := [3]float64{inkR, inkG, inkB}
	matR, matG, matB, _ := ColorToLinearRGB(cfg.PaperMatColor)
	matLin := [3]float64{matR, matG, matB}
	BounceR, BounceG, BounceB, _ := ColorToLinearRGB(cfg.WarmBounceColor)
	bounceLin := [3]float64{BounceR, BounceG, BounceB}
	defaultDeepShadow := [3]float64{0.008, 0.012, 0.028} // #12182c Deep Arzach Indigo-Black

	// Stage 2: Resolve 5-stage gouache color, organic rock cracks & Ligne Claire contours
	pxRowsPerWorker := (h + numWorkers - 1) / numWorkers
	for worker := 0; worker < numWorkers; worker++ {
		py0 := worker * pxRowsPerWorker
		py1 := py0 + pxRowsPerWorker
		if py1 > h {
			py1 = h
		}
		if py0 >= py1 {
			continue
		}
		wg.Add(1)
		go func(startY, endY int) {
			defer wg.Done()
			for py := startY; py < endY; py++ {
				for px := 0; px < w; px++ {
					var accR, accG, accB float64

					for dy := 0; dy < 2; dy++ {
						sy := py*2 + dy
						fsy := float64(sy)
						for dx := 0; dx < 2; dx++ {
							sx := px*2 + dx
							fsx := float64(sx)
							center := gbuf[sy*sw+sx]

							// 2a. 5-Stage Mœbius Gouache / Watercolor Shading (inspired by Arzach & Panzer Dragoon)
							illum := common.Clamp01(center.Illum)
							midC := center.MidColor
							if midC[0] == 0 && midC[1] == 0 && midC[2] == 0 {
								midC = moebLerp3(center.ShadeColor, center.SunColor, 0.52)
							}
							deepC := center.DeepShadowColor
							if deepC[0] == 0 && deepC[1] == 0 && deepC[2] == 0 {
								deepC = moebLerp3(defaultDeepShadow, center.ShadeColor, 0.35)
							}

							var col [3]float64
							if illum < 0.22 {
								// Stage 1: Velvety Chiaroscuro Deep Shadow Core (like the base/right of Arzach's rock tower)
								u := common.Smoothstep(0.05, 0.22, illum)
								col = moebLerp3(deepC, center.ShadeColor, u)
							} else if illum < 0.46 {
								// Stage 2: Cool Indigo/Cerulean Penumbra to Warm Vermilion/Terracotta Terminator
								u := common.Smoothstep(0.22, 0.46, illum)
								col = moebLerp3(center.ShadeColor, midC, u)
								bounceW := common.Clamp01(center.Bounce + 0.25*math.Max(0.0, center.Normal[1]))
								col = moebLerp3(col, bounceLin, 0.28*bounceW*(1.0-u*0.5))
							} else {
								// Stage 3: Warm Midtone to Sunlit Ochre/Gold/Cream Plateau
								u := common.Smoothstep(0.46, 0.72, illum)
								col = moebLerp3(midC, center.SunColor, u)
							}

							// Stage 4: Grazing sunlit crest highlight
							specW := math.Max(center.Spec, common.Smoothstep(0.82, 0.98, illum)*0.45)
							if specW > 0 {
								col = moebLerp3(col, matLin, common.Clamp01(specW*0.65))
							}

							// Stage 5: Atmospheric aerial perspective haze
							if center.Atmos > 0 && center.Depth < 85.0 {
								col = moebLerp3(col, center.AtmosColor, common.Clamp01(center.Atmos*0.78))
							}

							// 2b. Fine terminator contour hachures (concentrated in 0.16 < illum < 0.54!)
							var baseHatchInk float64
							if illum > 0.12 && illum < 0.54 && center.Depth < 80.0 {
								tx, ty := center.FlowTangent[0], center.FlowTangent[1]
								// Slight organic hand-wave along the hachure stroke
								handWave := 0.85 * math.Sin(fsx*0.045+fsy*0.035)
								uProj := (-fsx*ty+fsy*tx)*0.82 + handWave
								w1 := 0.5 + 0.5*math.Cos(uProj)
								// Break long lines into short Mœbius hachure segments along (fsx*tx + fsy*ty)
								alongStroke := (fsx*tx + fsy*ty) * 0.14
								segBreak := common.Smoothstep(-0.35, 0.15, math.Sin(alongStroke+float64(center.PartID)*1.7))
								termBell := common.Smoothstep(0.10, 0.22, illum) * common.Smoothstep(0.54, 0.34, illum)
								line1 := (1.0 - common.Smoothstep(0.08, 0.22, w1)) * segBreak * termBell * 0.68
								baseHatchInk = line1
							}

							// 2c. Organic pointillist ink stipple & weathered stone pockmarks
							var stippleInk float64
							if center.Stipple > 0.015 {
								su := fsx*0.74 + fsy*0.46 + 1.2*math.Sin(fsy*0.08)
								sv := -fsx*0.46 + fsy*0.74 + 1.2*math.Cos(fsx*0.08)
								dotField := (0.5 + 0.5*math.Cos(su*1.42)) * (0.5 + 0.5*math.Cos(sv*1.42))
								thresh := 1.0 - 0.54*common.Clamp01(center.Stipple)
								stippleInk = common.Smoothstep(thresh-0.07, thresh+0.05, dotField) *
									common.Clamp01(center.Stipple*1.15) * 0.76
							}

							// 2d. 3-Hierarchy Ligne Claire Neighborhood Edge Detection
							xL := sx - 1
							if xL < 0 {
								xL = 0
							}
							xR := sx + 1
							if xR >= sw {
								xR = sw - 1
							}
							yU := sy - 1
							if yU < 0 {
								yU = 0
							}
							yD := sy + 1
							if yD >= sh {
								yD = sh - 1
							}

							sL := gbuf[sy*sw+xL]
							sR := gbuf[sy*sw+xR]
							sU := gbuf[yU*sw+sx]
							sD := gbuf[yD*sw+sx]
							sUL := gbuf[yU*sw+xL]
							sUR := gbuf[yU*sw+xR]
							sDL := gbuf[yD*sw+xL]
							sDR := gbuf[yD*sw+xR]

							minDepth := math.Min(center.Depth, math.Min(math.Min(sL.Depth, sR.Depth), math.Min(sU.Depth, sD.Depth)))
							depthWeight := common.Clamp01(1.14 - 0.045*minDepth)
							if depthWeight < 0.50 {
								depthWeight = 0.50
							}
							shadowBoost := 0.90 + 0.20*(1.0-illum)

							// Hierarchy 1: Primary Silhouette Contour
							objDiff := 0.0
							if sL.ObjID != center.ObjID || sR.ObjID != center.ObjID ||
								sU.ObjID != center.ObjID || sD.ObjID != center.ObjID {
								objDiff = 1.0
							} else if minDepth < 5.5 && (sUL.ObjID != center.ObjID || sUR.ObjID != center.ObjID ||
								sDL.ObjID != center.ObjID || sDR.ObjID != center.ObjID) {
								objDiff = 0.70 * depthWeight
							}

							dGradX := math.Abs(sR.Depth - sL.Depth)
							dGradY := math.Abs(sD.Depth - sU.Depth)
							relDepthJump := math.Hypot(dGradX, dGradY) / (0.34 + 0.18*minDepth)
							depthEdge := common.Smoothstep(0.28, 0.62, relDepthJump) * depthWeight
							silhouetteInk := math.Max(objDiff*depthWeight, depthEdge) * shadowBoost

							// Hierarchy 2: Secondary Major Crease / Fissure
							partDiff := 0.0
							if sL.PartID != center.PartID || sR.PartID != center.PartID ||
								sU.PartID != center.PartID || sD.PartID != center.PartID {
								partDiff = 0.82 * depthWeight
							}
							nDiffX := math.Hypot(sR.Normal[0]-sL.Normal[0], math.Hypot(sR.Normal[1]-sL.Normal[1], sR.Normal[2]-sL.Normal[2]))
							nDiffY := math.Hypot(sD.Normal[0]-sU.Normal[0], math.Hypot(sD.Normal[1]-sU.Normal[1], sD.Normal[2]-sU.Normal[2]))
							normEdge := common.Smoothstep(0.52, 1.05, math.Hypot(nDiffX, nDiffY)) * 0.72 * depthWeight
							creaseInk := math.Max(partDiff, math.Max(normEdge, center.CrackInk*0.86*depthWeight))

							// Hierarchy 3: Tertiary Boulder / Glyph / Tick Hairline
							detailDiff := 0.0
							if sL.DetailID != center.DetailID || sU.DetailID != center.DetailID {
								detailDiff = 0.48 * depthWeight
							}
							microInk := math.Max(detailDiff, center.TickMask*0.76*depthWeight)

							inkFade := 1.0 - 0.35*common.Clamp01(center.Atmos)
							totalInk := common.Clamp01(math.Max(
								math.Max(silhouetteInk, creaseInk),
								math.Max(microInk, math.Max(baseHatchInk, stippleInk)),
							) * inkFade)

							col = moebLerp3(col, inkLin, totalInk)
							accR += col[0]
							accG += col[1]
							accB += col[2]
						}
					}

					rc.SetPixelLinear(px, py, accR*0.25, accG*0.25, accB*0.25, 1.0)
				}
			}
		}(py0, py1)
	}
	wg.Wait()

	// Stage 3: Short, Organic Terminator-Clustered RK4 Rapidograph Hachure Strokes
	if cfg.EnableRK4Hatch {
		traceMoebiusCurvatureHatching(rc, gbuf, sw, sh, cfg)
	}

	// Stage 4: Archival Métal Hurlant Double-Ruled Serigraph Frame & Cotton Paper Grain
	drawMoebiusSerigraphBorder(rc, cfg)
	rc.ApplyPaperTexture(0.068, cfg.Seed)
}

func traceMoebiusCurvatureHatching(rc *Canvas, gbuf []MoebiusPixel, sw, sh int, cfg MoebiusConfig) {
	w, h := rc.Width, rc.Height
	fw, fh := float64(w), float64(h)
	marginPx := cfg.BorderMargin*math.Min(fw, fh) + 3.0

	rng := rand.New(rand.NewSource(cfg.Seed + 7919))
	spacing := cfg.HatchSpacingPx

	sampleG := func(px, py float64) MoebiusPixel {
		sx := int(px * 2.0)
		sy := int(py * 2.0)
		if sx < 0 {
			sx = 0
		} else if sx >= sw {
			sx = sw - 1
		}
		if sy < 0 {
			sy = 0
		} else if sy >= sh {
			sy = sh - 1
		}
		return gbuf[sy*sw+sx]
	}

	// Trace short, contour-wrapping Mœbius hachure strokes concentrated on terminators & undersides
	traceStroke := func(seedX, seedY float64, crossPass bool) ([]Point, float64, float64) {
		s0 := sampleG(seedX, seedY)
		if s0.Depth >= 80.0 {
			return nil, 0, 0
		}
		if !crossPass && (s0.Illum >= 0.54 || s0.Illum < 0.08) {
			return nil, 0, 0
		}
		if crossPass && s0.Illum >= 0.24 {
			return nil, 0, 0
		}

		getDir := func(x, y float64, prevDx, prevDy float64) (float64, float64, bool) {
			if x < marginPx || x >= fw-marginPx || y < marginPx || y >= fh-marginPx {
				return 0, 0, false
			}
			smp := sampleG(x, y)
			if smp.ObjID != s0.ObjID || math.Abs(smp.Depth-s0.Depth) > 0.36 {
				return 0, 0, false
			}
			if !crossPass && smp.Illum > 0.56 {
				return 0, 0, false
			}
			if crossPass && smp.Illum > 0.28 {
				return 0, 0, false
			}
			tx, ty := smp.FlowTangent[0], smp.FlowTangent[1]
			if crossPass {
				tx, ty = -ty, tx
			}
			if prevDx*tx+prevDy*ty < 0 {
				tx, ty = -tx, -ty
			}
			return tx, ty, true
		}

		const stepLen = 1.35
		const maxHalfSteps = 5 // Short, crisp Mœbius hachure ticks (~12px total length)

		var backPts [maxHalfSteps]Point
		nBack := 0
		cx, cy := seedX, seedY
		pdx, pdy := -s0.FlowTangent[0], -s0.FlowTangent[1]
		if crossPass {
			pdx, pdy = s0.FlowTangent[1], -s0.FlowTangent[0]
		}
		for s := 0; s < maxHalfSteps; s++ {
			k1x, k1y, ok := getDir(cx, cy, pdx, pdy)
			if !ok {
				break
			}
			k2x, k2y, ok := getDir(cx+0.5*stepLen*k1x, cy+0.5*stepLen*k1y, k1x, k1y)
			if !ok {
				break
			}
			cx += stepLen * k2x
			cy += stepLen * k2y
			pdx, pdy = k2x, k2y
			backPts[nBack] = Pt(cx, cy)
			nBack++
		}

		pts := make([]Point, 0, nBack+1+maxHalfSteps)
		for i := nBack - 1; i >= 0; i-- {
			pts = append(pts, backPts[i])
		}
		pts = append(pts, Pt(seedX, seedY))

		cx, cy = seedX, seedY
		pdx, pdy = s0.FlowTangent[0], s0.FlowTangent[1]
		if crossPass {
			pdx, pdy = -s0.FlowTangent[1], s0.FlowTangent[0]
		}
		for s := 0; s < maxHalfSteps; s++ {
			k1x, k1y, ok := getDir(cx, cy, pdx, pdy)
			if !ok {
				break
			}
			k2x, k2y, ok := getDir(cx+0.5*stepLen*k1x, cy+0.5*stepLen*k1y, k1x, k1y)
			if !ok {
				break
			}
			cx += stepLen * k2x
			cy += stepLen * k2y
			pdx, pdy = k2x, k2y
			pts = append(pts, Pt(cx, cy))
		}

		if len(pts) < 3 {
			return nil, 0, 0
		}
		depthScale := common.Clamp01(1.10 - 0.05*s0.Depth)
		if depthScale < 0.50 {
			depthScale = 0.50
		}
		shadowDarkness := common.Smoothstep(0.54, 0.14, s0.Illum)
		width := (0.55 + 0.35*shadowDarkness) * depthScale
		alpha := (0.46 + 0.42*shadowDarkness) * (1.0 - 0.35*common.Clamp01(s0.Atmos))
		return pts, width, alpha
	}

	rc.SetBlendMode(BlendOver)
	inkFn := func(t float64) color.RGBA {
		return cfg.InkColor
	}

	for y := marginPx; y < fh-marginPx; y += spacing {
		for x := marginPx; x < fw-marginPx; x += spacing {
			jx := x + (rng.Float64()-0.5)*spacing*0.65
			jy := y + (rng.Float64()-0.5)*spacing*0.65
			pts, strokeW, alpha := traceStroke(jx, jy, false)
			if len(pts) >= 3 {
				rc.DrawTaperedSpline(pts, strokeW, TaperBothEnds, inkFn, alpha)
			}
		}
	}

	crossSpacing := spacing * 1.25
	for y := marginPx; y < fh-marginPx; y += crossSpacing {
		for x := marginPx; x < fw-marginPx; x += crossSpacing {
			jx := x + (rng.Float64()-0.5)*crossSpacing*0.65
			jy := y + (rng.Float64()-0.5)*crossSpacing*0.65
			pts, strokeW, alpha := traceStroke(jx, jy, true)
			if len(pts) >= 3 {
				rc.DrawTaperedSpline(pts, strokeW*0.85, TaperBothEnds, inkFn, alpha*0.82)
			}
		}
	}
}

func drawMoebiusSerigraphBorder(rc *Canvas, cfg MoebiusConfig) {
	w, h := rc.Width, rc.Height
	fw, fh := float64(w), float64(h)

	inkR, inkG, inkB, _ := ColorToLinearRGB(cfg.InkColor)
	matR, matG, matB, _ := ColorToLinearRGB(cfg.PaperMatColor)
	ruleR, ruleG, ruleB, _ := ColorToLinearRGB(cfg.InnerRuleColor)
	filletR := matR*0.65 + ruleR*0.35
	filletG := matG*0.65 + ruleG*0.35
	filletB := matB*0.65 + ruleB*0.35

	mOuter := cfg.BorderMargin * 0.78
	mRule1 := cfg.BorderMargin * 0.88
	mFillet := cfg.BorderMargin * 0.96
	mRule2 := cfg.BorderMargin * 1.04

	for y := 0; y < h; y++ {
		uy := (float64(y) + 0.5) / fh
		dy := math.Min(uy, 1.0-uy)
		for x := 0; x < w; x++ {
			ux := (float64(x) + 0.5) / fw
			dx := math.Min(ux, 1.0-ux)
			distEdge := math.Min(dx, dy)
			if distEdge >= mRule2 {
				continue
			}
			if distEdge < mOuter {
				rc.SetPixelLinear(x, y, matR, matG, matB, 1.0)
			} else if distEdge < mRule1 {
				rc.SetPixelLinear(x, y, inkR, inkG, inkB, 1.0)
			} else if distEdge < mFillet {
				rc.SetPixelLinear(x, y, filletR, filletG, filletB, 1.0)
			} else {
				rc.SetPixelLinear(x, y, inkR, inkG, inkB, 1.0)
			}
		}
	}
}
