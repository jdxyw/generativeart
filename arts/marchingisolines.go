package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type marchingIsolines struct {
	levels     int
	noiseScale float64
}

// NewMarchingIsolines returns a marchingIsolines engine that renders stepped
// Huanglong-inspired karst travertine pools and topographic terraces using
// Marching Squares contour extraction, 3D elevation drop shadows, and gilded rims.
func NewMarchingIsolines(levels int, noiseScale float64) *marchingIsolines {
	if levels <= 0 {
		levels = 16
	}
	if noiseScale <= 0 {
		noiseScale = 1.0
	}
	return &marchingIsolines{
		levels:     levels,
		noiseScale: noiseScale,
	}
}

// Generative renders the stepped Marching Squares topographic travertine terraces.
func (mi *marchingIsolines) Generative(c *generativeart.Canva) {
	mi.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360-degree phase loop (phase in [0, 1)).
func (mi *marchingIsolines) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, loopPhase float64) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0
	animPhase := loopPhase * 2.0 * math.Pi

	// Curated Huanglong Karst Travertine & Dunhuang Mineral Palette
	abyssalNavy := color.RGBA{R: 12, G: 22, B: 42, A: 255}
	prussianBlue := color.RGBA{R: 22, G: 58, B: 108, A: 255}
	cobaltLagoon := color.RGBA{R: 36, G: 112, B: 162, A: 255}
	jadeTeal := color.RGBA{R: 54, G: 162, B: 158, A: 255}
	mintTurquoise := color.RGBA{R: 124, G: 212, B: 192, A: 255}
	paleTravertine := color.RGBA{R: 212, G: 242, B: 228, A: 255}

	warmSand := color.RGBA{R: 244, G: 222, B: 168, A: 255}
	goldenOchre := color.RGBA{R: 238, G: 174, B: 64, A: 255}
	cadmiumCoral := color.RGBA{R: 226, G: 108, B: 52, A: 255}
	crimsonTerracotta := color.RGBA{R: 184, G: 54, B: 44, A: 255}

	espressoShadow := color.RGBA{R: 16, G: 14, B: 24, A: 255}
	travertineGold := color.RGBA{R: 248, G: 216, B: 132, A: 255}

	rc.FillBackground(abyssalNavy)
	rc.SetBlendMode(render.BlendOver)

	ne := common.NewNoiseEngine(202613)
	freq := 2.4 * mi.noiseScale / math.Max(1.0, math.Min(w, h))

	// Domain-warped scalar elevation field with 2 gentle caldera peaks for dramatic terraced pools
	ox := 0.15 * (1.0 - math.Cos(animPhase))
	oy := 0.15 * math.Sin(animPhase)
	evalScalarField := func(x, y float64) float64 {
		u := x * freq
		v := y * freq
		wx := ne.FBM2D(u+1.7+ox, v+9.2+oy, 3, 2.0, 0.5) * 0.55
		wy := ne.FBM2D(u+8.3-oy, v+2.8+ox, 3, 2.0, 0.5) * 0.55
		fbm := 0.5 + 0.5*ne.FBM2D(u+wx+ox*0.25, v+wy+oy*0.25, 4, 2.05, 0.5)

		// Add two organic mountain/pool domes that gently breathe and orbit
		d1 := math.Hypot((x-(0.34+0.014*math.Sin(animPhase))*w)/w, (y-(0.36+0.014*(1.0-math.Cos(animPhase)))*h)/h)
		d2 := math.Hypot((x-(0.68-0.014*math.Sin(animPhase))*w)/w, (y-(0.66-0.014*(1.0-math.Cos(animPhase)))*h)/h)
		dome := 0.32*(1.0+0.07*math.Sin(animPhase))*math.Exp(-d1*d1*6.5) + 0.28*(1.0-0.07*math.Sin(animPhase))*math.Exp(-d2*d2*7.5)

		return common.Constrain(fbm*0.76+dome-0.06, 0, 1)
	}

	// Precompute scalar field on a 1px grid (plus shadow offset margin) for fast parallel terrace shading
	gw, gh := c.Width(), c.Height()
	fieldGrid := make([]float64, gw*gh)
	for y := 0; y < gh; y++ {
		fy := float64(y) + 0.5
		rowOff := y * gw
		for x := 0; x < gw; x++ {
			fx := float64(x) + 0.5
			fieldGrid[rowOff+x] = evalScalarField(fx, fy)
		}
	}

	grad := render.NewMultiStopGradient(
		abyssalNavy,
		prussianBlue,
		cobaltLagoon,
		jadeTeal,
		mintTurquoise,
		paleTravertine,
		warmSand,
		goldenOchre,
		cadmiumCoral,
		crimsonTerracotta,
	)

	numLevels := mi.levels
	if numLevels < 8 {
		numLevels = 16
	}
	thresholds := make([]float64, numLevels)
	levelColors := make([]color.RGBA, numLevels)
	for i := 0; i < numLevels; i++ {
		t := float64(i) / float64(numLevels-1)
		thresholds[i] = 0.14 + t*0.74
		levelColors[i] = grad.SampleOklab(t)
	}

	// 1. Render Stepped 3D Terrace Plateaus, Micro-Ripple Striations & Cast Cliff Shadows
	shadowDX := int(math.Round(4.5 * scale))
	shadowDY := int(math.Round(5.5 * scale))
	if shadowDX < 1 {
		shadowDX = 1
	}
	if shadowDY < 1 {
		shadowDY = 1
	}

	for y := 0; y < gh; y++ {
		rowOff := y * gw
		for x := 0; x < gw; x++ {
			v := fieldGrid[rowOff+x]

			// Find highest terrace index k where v >= thresholds[k]
			lvl := -1
			for k := numLevels - 1; k >= 0; k-- {
				if v >= thresholds[k] {
					lvl = k
					break
				}
			}

			var col color.RGBA
			if lvl < 0 {
				col = abyssalNavy
			} else {
				col = levelColors[lvl]
				// Subtle concentric pool ripple modulation inside each terrace band
				nextTh := 1.0
				if lvl+1 < numLevels {
					nextTh = thresholds[lvl+1]
				}
				bandFrac := common.Constrain((v-thresholds[lvl])/math.Max(1e-4, nextTh-thresholds[lvl]), 0, 1)
				ripple := 0.5 + 0.5*math.Sin(bandFrac*math.Pi*4.0)
				if bandFrac > 0.78 {
					// Bright sunlit rim bevel near the upper step
					col = render.LerpOklab(col, paleTravertine, (bandFrac-0.78)*1.4)
				} else {
					col = render.LerpOklab(col, travertineGold, ripple*0.08)
				}
			}

			// Check if a higher terrace at (x - shadowDX, y - shadowDY) casts a drop shadow onto (x, y)
			sx := x - shadowDX
			sy := y - shadowDY
			if sx >= 0 && sx < gw && sy >= 0 && sy < gh {
				vCaster := fieldGrid[sy*gw+sx]
				casterLvl := -1
				for k := numLevels - 1; k >= 0; k-- {
					if vCaster >= thresholds[k] {
						casterLvl = k
						break
					}
				}
				if casterLvl > lvl {
					col = render.LerpOklab(col, espressoShadow, 0.48)
				}
			}

			rc.BlendPixel(x, y, col, 1.0)
		}
	}

	// 2. Extract Exact Sub-Pixel Marching Squares Contours & Stroke Travertine Rim Dykes
	gridRes := 110
	for k := 0; k < numLevels; k++ {
		iso := thresholds[k]
		contours := render.ExtractMarchingSquares(gridRes, gridRes, w, h, iso, evalScalarField)

		darkPath := &canvas.Path{}
		goldPath := &canvas.Path{}
		for _, poly := range contours {
			if len(poly) < 3 {
				continue
			}
			smooth := render.CatmullRomSpline(poly, 4)
			darkPath.MoveTo(smooth[0].X+1.0*scale, smooth[0].Y+1.2*scale)
			goldPath.MoveTo(smooth[0].X, smooth[0].Y)
			for i := 1; i < len(smooth); i++ {
				darkPath.LineTo(smooth[i].X+1.0*scale, smooth[i].Y+1.2*scale)
				goldPath.LineTo(smooth[i].X, smooth[i].Y)
			}
		}

		rimW := 1.45 * scale
		if k%3 == 0 {
			rimW = 2.05 * scale
		}
		rc.StrokePath(darkPath, rimW*1.35, canvas.RoundCap, canvas.RoundJoin, espressoShadow, 0.82)

		rimCol := travertineGold
		if k < numLevels/2 {
			rimCol = render.LerpOklab(paleTravertine, travertineGold, 0.55)
		}
		rc.StrokePath(goldPath, rimW*0.85, canvas.RoundCap, canvas.RoundJoin, rimCol, 0.94)
	}

	c.MergeRenderCanvas(rc)
}
