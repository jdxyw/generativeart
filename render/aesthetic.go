package render

import (
	"math"

	"github.com/jdxyw/generativeart/common"
)

// AestheticMetrics holds perceptual quality metrics evaluated in Oklab/Oklch space on a Canvas.
type AestheticMetrics struct {
	TotalScore        float64 `json:"total_score"`        // Composite aesthetic quality score in [0, 100]
	DynamicRange      float64 `json:"dynamic_range"`      // 5th-to-95th percentile Oklab lightness span in [0, 1]
	ContrastStd       float64 `json:"contrast_std"`       // Normalized Oklab lightness standard deviation in [0, 1]
	ColorHarmony      float64 `json:"color_harmony"`      // Oklch chroma richness and hue-wheel harmonic balance in [0, 1]
	StructuralEntropy float64 `json:"structural_entropy"` // Multi-scale edge gradient richness vs noise penalty in [0, 1]
	FocalSaliency     float64 `json:"focal_saliency"`     // Central / rule-of-thirds focal composition strength in [0, 1]
	ClippingPenalty   float64 `json:"clipping_penalty"`   // Penalty for blank/monotone or over-exposed canvases in [0, 1]
}

// EvaluateAestheticMetrics analyzes the Canvas in perceptual Oklab/Oklch space and returns
// an AestheticMetrics breakdown and composite TotalScore in [0, 100].
func (c *Canvas) EvaluateAestheticMetrics() AestheticMetrics {
	if c == nil || c.Width < 4 || c.Height < 4 {
		return AestheticMetrics{}
	}
	w, h := c.Width, c.Height

	// Sample on at most a 160x160 stride grid so evaluation takes < 0.5ms even on 4K canvases
	stepX := common.MaxInt(1, w/160)
	stepY := common.MaxInt(1, h/160)
	gw := w / stepX
	gh := h / stepY
	if gw < 4 || gh < 4 {
		gw, gh = w, h
		stepX, stepY = 1, 1
	}

	nSamples := gw * gh
	lumGrid := make([]float64, nSamples)
	chromaGrid := make([]float64, nSamples)

	var lumHist [256]int
	var hueHist [24]float64
	var sumL, sumL2, sumC, totalHueWeight float64
	overExposed := 0

	for gy := 0; gy < gh; gy++ {
		py := common.ConstrainInt(gy*stepY, 0, h-1)
		for gx := 0; gx < gw; gx++ {
			px := common.ConstrainInt(gx*stepX, 0, w-1)
			pIdx := (py*w + px) * 4
			rLin := common.Constrain(c.pix[pIdx+0], 0, 1)
			gLin := common.Constrain(c.pix[pIdx+1], 0, 1)
			bLin := common.Constrain(c.pix[pIdx+2], 0, 1)

			lab := LinearRGBAToOklab(rLin, gLin, bLin)
			L := common.Clamp01(lab.L)
			C := math.Hypot(lab.A, lab.B)

			sIdx := gy*gw + gx
			lumGrid[sIdx] = L
			chromaGrid[sIdx] = C

			sumL += L
			sumL2 += L * L
			sumC += C

			bin := common.ConstrainInt(int(math.Round(L*255.0)), 0, 255)
			lumHist[bin]++
			if L > 0.992 {
				overExposed++
			}
			if C > 0.022 {
				ang := math.Atan2(lab.B, lab.A) + math.Pi // in [0, 2*pi]
				hBin := int(math.Floor(ang/(2.0*math.Pi)*24.0)) % 24
				if hBin < 0 {
					hBin += 24
				}
				hueHist[hBin] += C
				totalHueWeight += C
			}
		}
	}

	fn := float64(nSamples)
	meanL := sumL / fn
	varL := math.Max(0, sumL2/fn-meanL*meanL)
	stdL := math.Sqrt(varL)
	meanC := sumC / fn

	// 1. Dynamic Range (5th to 95th percentile lightness span) & Contrast Std
	p05Target := int(0.05 * fn)
	p95Target := int(0.95 * fn)
	cum := 0
	l05, l95 := 0.0, 1.0
	found05 := false
	maxSingleBin := 0
	for i := 0; i < 256; i++ {
		if lumHist[i] > maxSingleBin {
			maxSingleBin = lumHist[i]
		}
		cum += lumHist[i]
		if !found05 && cum >= p05Target {
			l05 = float64(i) / 255.0
			found05 = true
		}
		if cum >= p95Target {
			l95 = float64(i) / 255.0
			break
		}
	}
	dynRange := common.Clamp01((l95 - l05) / 0.68)
	contrastScore := common.Clamp01(stdL / 0.22)

	// 2. Color Harmony (Chroma richness + Hue wheel Shannon entropy)
	chromaScore := common.Smoothstep(0.015, 0.11, meanC) * (1.0 - 0.35*common.Smoothstep(0.26, 0.38, meanC))
	hueEntropy := 0.0
	if totalHueWeight > 1e-6 {
		for b := 0; b < 24; b++ {
			p := hueHist[b] / totalHueWeight
			if p > 1e-7 {
				hueEntropy -= p * math.Log2(p)
			}
		}
	}
	// Optimal hue entropy is 1.4..3.6 bits (structured multi-hue palette without chaotic rainbow static)
	hueHarmony := common.Smoothstep(0.4, 1.8, hueEntropy) * (1.0 - 0.25*common.Smoothstep(3.8, 4.5, hueEntropy))
	colorHarmony := common.Clamp01(0.55*chromaScore + 0.45*hueHarmony)
	if meanC < 0.025 && contrastScore > 0.65 {
		// High-contrast monochromatic copperplate engraving / ink print
		colorHarmony = math.Max(colorHarmony, 0.72*contrastScore)
	}

	// 3. Structural Entropy & Focal Saliency (Gradient distribution across center vs periphery)
	var sumGrad, centerEnergy, outerEnergy float64
	var centerCount, outerCount int
	var tileGrads [16]float64
	var tileCounts [16]int

	for gy := 1; gy < gh-1; gy++ {
		uy := (float64(gy) + 0.5) / float64(gh)
		ty := common.ConstrainInt(int(uy*4.0), 0, 3)
		for gx := 1; gx < gw-1; gx++ {
			ux := (float64(gx) + 0.5) / float64(gw)
			tx := common.ConstrainInt(int(ux*4.0), 0, 3)

			idx := gy*gw + gx
			gxVal := lumGrid[idx+1] - lumGrid[idx-1]
			gyVal := lumGrid[idx+gw] - lumGrid[idx-gw]
			grad := math.Hypot(gxVal, gyVal)
			sumGrad += grad

			tIdx := ty*4 + tx
			tileGrads[tIdx] += grad
			tileCounts[tIdx]++

			distCenter := math.Hypot(ux-0.5, uy-0.5)
			sal := grad + 0.65*chromaGrid[idx]
			if distCenter < 0.36 {
				centerEnergy += sal
				centerCount++
			} else {
				outerEnergy += sal
				outerCount++
			}
		}
	}

	nInner := float64(common.MaxInt(1, (gw-2)*(gh-2)))
	meanGrad := sumGrad / nInner
	// Sweet spot for structural edge density: meanGrad in [0.035, 0.22] (penalize blank <0.005 or noisy >0.38)
	structScore := common.Smoothstep(0.008, 0.055, meanGrad) * (1.0 - 0.55*common.Smoothstep(0.28, 0.48, meanGrad))

	// Check macro-tile structural diversity (so the image isn't 95% empty with 1 tiny dot)
	activeTiles := 0
	for t := 0; t < 16; t++ {
		if tileCounts[t] > 0 && (tileGrads[t]/float64(tileCounts[t])) > meanGrad*0.25 {
			activeTiles++
		}
	}
	tileCoverage := common.Smoothstep(3.0, 10.0, float64(activeTiles))
	structuralEntropy := common.Clamp01(structScore * (0.45 + 0.55*tileCoverage))

	focalSaliency := 0.5
	if centerCount > 0 && outerCount > 0 {
		cAvg := centerEnergy / float64(centerCount)
		oAvg := outerEnergy / float64(outerCount)
		ratio := cAvg / math.Max(1e-4, oAvg)
		focalSaliency = common.Clamp01(common.Smoothstep(0.65, 1.45, ratio))
	}

	// 4. Clipping / Blank Canvas Penalty
	monoFrac := float64(maxSingleBin) / fn
	overFrac := float64(overExposed) / fn
	clipPenalty := common.Clamp01(
		common.Smoothstep(0.55, 0.90, monoFrac) +
			0.75*common.Smoothstep(0.35, 0.75, overFrac) +
			common.Smoothstep(0.015, 0.002, l95-l05),
	)

	rawScore := 25.0*dynRange + 20.0*contrastScore + 20.0*colorHarmony + 20.0*structuralEntropy + 15.0*focalSaliency - 45.0*clipPenalty
	total := common.Constrain(math.Round(rawScore*10.0)/10.0, 0.0, 100.0)

	return AestheticMetrics{
		TotalScore:        total,
		DynamicRange:      math.Round(dynRange*1000) / 1000,
		ContrastStd:       math.Round(contrastScore*1000) / 1000,
		ColorHarmony:      math.Round(colorHarmony*1000) / 1000,
		StructuralEntropy: math.Round(structuralEntropy*1000) / 1000,
		FocalSaliency:     math.Round(focalSaliency*1000) / 1000,
		ClippingPenalty:   math.Round(clipPenalty*1000) / 1000,
	}
}
