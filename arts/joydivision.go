package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type joyDivision struct {
	lineNum             int
	ampScale, freqScale float64
}

// NewJoyDivision returns a joyDivision engine that renders an elevated Pulsar
// Horizon-Occluded Ridgeline Wave artwork inspired by CP 1919 (Unknown Pleasures).
func NewJoyDivision(lineNum int, ampScale, freqScale float64) *joyDivision {
	if lineNum <= 0 {
		lineNum = 72
	}
	if ampScale <= 0 {
		ampScale = 1.0
	}
	if freqScale <= 0 {
		freqScale = 1.0
	}
	return &joyDivision{
		lineNum:   lineNum,
		ampScale:  ampScale,
		freqScale: freqScale,
	}
}

// Generative renders the Pulsar Horizon-Occluded Ridgeline Wave artwork onto the canvas.
func (jd *joyDivision) Generative(c *generativeart.Canva) {
	jd.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360-degree phase loop (phase in [0, 1)).
func (jd *joyDivision) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, loopPhase float64) {
	animPhase := loopPhase * 2.0 * math.Pi
	lineNum := jd.lineNum
	if lineNum <= 0 {
		lineNum = 72
	}
	ampScale := jd.ampScale
	if ampScale <= 0 {
		ampScale = 1.0
	}
	freqScale := jd.freqScale
	if freqScale <= 0 {
		freqScale = 1.0
	}

	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	seed := c.Seed()
	if seed == 0 {
		seed = 202604
	}
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	schema := c.Opts().ColorSchema()
	if len(schema) == 0 {
		schema = common.Cyberpunk
	}
	grad := render.NewMultiStopGradient(schema...)

	rc := c.NewRenderCanvas()

	// 1. Deep obsidian-indigo space background with a subtle vertical cosmic glow column in the center.
	bgBase := color.RGBA{R: 5, G: 6, B: 16, A: 255}
	bgCenterTint := render.LerpOklab(bgBase, grad.SampleOklab(0.28), 0.14)
	for y := 0; y < h; y++ {
		vy := float64(y) / math.Max(1.0, fh-1.0)
		vertEnv := math.Sin(vy * math.Pi)
		for x := 0; x < w; x++ {
			ux := (float64(x)/math.Max(1.0, fw-1.0))*2.0 - 1.0
			colGlow := math.Exp(-5.5*ux*ux) * vertEnv
			col := render.LerpOklab(bgBase, bgCenterTint, colGlow*0.75)
			rLin, gLin, bLin, _ := render.ColorToLinearRGB(col)
			rc.SetPixelLinear(x, y, rLin, gLin, bLin, 1.0)
		}
	}

	// Fine starfield in the upper margin and outer flanks.
	rc.SetBlendMode(render.BlendAdd)
	starCount := 180
	for i := 0; i < starCount; i++ {
		sx := rng.Float64() * fw
		sy := math.Pow(rng.Float64(), 1.6) * (fh * 0.42)
		starCol := render.LerpOklab(
			color.RGBA{R: 190, G: 215, B: 255, A: 255},
			grad.SampleOklab(0.75),
			0.35*rng.Float64(),
		)
		fade := 1.0 - common.Clamp01(sy/(fh*0.42))*0.65
		if i%14 == 0 {
			rc.DrawGlowDot(sx, sy, (1.1+1.3*rng.Float64())*scale, starCol, 0.35*fade, 0.4)
		} else {
			rc.SplatPoint(sx, sy, starCol, (0.18+0.42*rng.Float64())*fade)
		}
	}

	// Soft vertical pulsar radio-beam glow column in the center.
	for k := 0; k < 14; k++ {
		kt := float64(k) / 13.0
		gy := common.Lerp(0.14*fh, 0.86*fh, kt)
		beamCol := grad.SampleOklab(0.35 + 0.35*math.Sin(kt*math.Pi))
		rc.DrawGlowDot(fw*0.5, gy, 54.0*scale, beamCol, 0.028, 0.0)
	}

	// 2. Iterate row i from 0 (top/back, yBase = 0.16*h) to lineNum-1 (bottom/front, yBase = 0.88*h).
	const samples = 160
	xLeft := 0.10 * fw
	xRight := 0.90 * fw
	maxPeakHeight := 92.0 * ampScale * scale
	occludeDrop := 45.0 * scale

	peakWhite := color.RGBA{R: 250, G: 248, B: 255, A: 255}
	darkVioletTint := color.RGBA{R: 13, G: 9, B: 28, A: 255}

	for i := 0; i < lineNum; i++ {
		rowFrac := float64(i) / float64(common.MaxInt(lineNum-1, 1))
		yBase := common.Lerp(0.16*fh, 0.88*fh, rowFrac)

		// Per-row pulsar sub-pulse drift so successive radio pulses evolve organically down the stack.
		ox := 0.15 * (1.0 - math.Cos(animPhase))
		oy := 0.15 * math.Sin(animPhase)
		rowSeedY := float64(i) * 0.31 * freqScale
		subShift1 := 0.28 + 0.09*noise.Simplex2D(rowFrac*3.2+ox, 1.7+oy) + 0.035*math.Sin(animPhase-rowFrac*math.Pi*2.0)
		subShift2 := -0.32 + 0.09*noise.Simplex2D(rowFrac*3.2-oy, 8.4+ox) - 0.035*math.Sin(animPhase-rowFrac*math.Pi*2.0)
		rowBurst := (0.72 + 0.42*noise.Simplex2D01(rowFrac*5.5+ox, 3.9+oy)) * (1.0 + 0.08*math.Sin(animPhase-rowFrac*math.Pi*2.0))

		ridgePts := make([]render.Point, samples)
		elevNorms := make([]float64, samples)

		for s := 0; s < samples; s++ {
			t := float64(s) / float64(samples-1)
			x := common.Lerp(xLeft, xRight, t)
			u := t*2.0 - 1.0 // normalized u in [-1, 1]

			// Central multi-modal Gaussian pulsar envelope:
			// exp(-4.5*u^2) + 0.45*exp(-12.0*(u-0.28)^2) + 0.4*exp(-14.0*(u+0.32)^2)
			d1 := u - subShift1
			d2 := u - subShift2
			pulsarEnv := math.Exp(-4.5*u*u) +
				0.45*math.Exp(-12.0*d1*d1) +
				0.40*math.Exp(-14.0*d2*d2)

			// Smooth window taper at outer ends so flanks settle cleanly into the baseline.
			edgeWindow := common.Smoothstep(0.0, 0.12, t) * common.Smoothstep(0.0, 0.12, 1.0-t)

			// Multi-octave Ridged2D + FBM2D noise and high-frequency carrier ripples.
			nx := u * 3.6 * freqScale
			ridged := noise.Ridged2D(nx+1.9+ox, rowSeedY+oy, 4, 2.15, 0.52)
			fbm := 0.5 * (noise.FBM2D(nx*1.6-2.4-oy, rowSeedY*1.2+4.1+ox, 4, 2.0, 0.5) + 1.0)
			carrier := 0.5 + 0.5*math.Cos((u*18.0*freqScale+noise.Simplex2D(nx*2.0, rowSeedY)*1.8)*math.Pi-animPhase)

			// Dramatic sharp pulsar peaks in the center, calm baseline micro-waves on the flanks.
			centerSignal := (0.58*math.Pow(ridged, 1.25) + 0.27*fbm + 0.15*carrier*ridged) * pulsarEnv * rowBurst
			flankWave := (0.018*noise.Simplex2D(u*7.5*freqScale, rowSeedY*1.5) + 0.012*ridged) * edgeWindow

			elev := (centerSignal*edgeWindow + flankWave) * maxPeakHeight
			if elev < 0 {
				elev *= 0.25
			}
			y := yBase - elev

			ridgePts[s] = render.Pt(x, y)
			elevNorms[s] = common.Clamp01(elev / (maxPeakHeight * 0.85))
		}

		// True Horizon Occlusion (Painter's Algorithm):
		// Construct a closed polygon consisting of the ridge polyline from left to right
		// and closing down to yBase + 45px, filled in BlendOver mode with the deep background color.
		poly := make([]render.Point, 0, samples+2)
		poly = append(poly, ridgePts...)
		poly = append(poly,
			render.Pt(xRight, yBase+occludeDrop),
			render.Pt(xLeft, yBase+occludeDrop),
		)

		occludeCol := render.LerpOklab(bgBase, darkVioletTint, 0.18+0.22*rowFrac)
		rc.SetBlendMode(render.BlendOver)
		rc.FillPolygon(poly, occludeCol, 0.98)

		// Helper to sample Oklab color along the ridge at normalized arc parameter t in [0, 1].
		sampleRidgeColor := func(t float64) color.RGBA {
			fIdx := common.Clamp01(t) * float64(samples-1)
			idx0 := int(math.Floor(fIdx))
			idx1 := common.MinInt(idx0+1, samples-1)
			frac := fIdx - float64(idx0)
			eNorm := common.Lerp(elevNorms[idx0], elevNorms[idx1], frac)

			// Color transitions in Oklab space based on local elevation height and row progression.
			gradT := common.Clamp01(0.12 + 0.18*rowFrac + 0.70*math.Pow(eNorm, 0.78))
			col := grad.SampleOklab(gradT)
			if eNorm > 0.55 {
				whiteMix := common.Smoothstep(0.55, 0.98, eNorm) * 0.72
				col = render.LerpOklab(col, peakWhite, whiteMix)
			}
			return col
		}

		// Ridge Crest Rendering:
		// 1) Soft glow along the ridge in BlendAdd mode.
		rc.SetBlendMode(render.BlendAdd)
		rc.DrawTaperedSpline(
			ridgePts,
			3.6*scale,
			func(t float64) float64 {
				fIdx := common.Clamp01(t) * float64(samples-1)
				idx0 := int(math.Floor(fIdx))
				eNorm := elevNorms[common.ConstrainInt(idx0, 0, samples-1)]
				return 0.35 + 0.65*eNorm
			},
			sampleRidgeColor,
			0.18+0.10*rowFrac,
		)

		// 2) Crisp luminous ridge crest stroke using DrawTaperedSpline + subtle core highlight.
		rc.SetBlendMode(render.BlendOver)
		rc.DrawTaperedSpline(
			ridgePts,
			(1.45+0.45*rowFrac)*scale,
			func(t float64) float64 {
				fIdx := common.Clamp01(t) * float64(samples-1)
				idx0 := int(math.Floor(fIdx))
				eNorm := elevNorms[common.ConstrainInt(idx0, 0, samples-1)]
				flankTaper := 0.55 + 0.45*math.Sin(common.Clamp01(t)*math.Pi)
				return flankTaper * (0.72 + 0.48*eNorm)
			},
			sampleRidgeColor,
			0.94,
		)

		// Additive specular crest filament on the highest peaks.
		rc.SetBlendMode(render.BlendAdd)
		rc.StrokePolyline(ridgePts, 0.75*scale, render.LerpOklab(grad.SampleOklab(0.85), peakWhite, 0.55), 0.24)
	}

	rc.ApplyBloom(0.48, 8.0, 0.55)
	rc.ApplyVignette(0.28)
	c.MergeRenderCanvasToneMapped(rc, 1.1, 0)
}
