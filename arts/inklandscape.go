package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type inkLandscape struct {
	mountainLayers int
	roughness      float64
}

// NewInkLandscape returns an inkLandscape engine that renders an East-Asian Shan-Shui
// ink-and-mineral wash landscape with misty ridged mountains, dry-brush calligraphy crests,
// recursive plum/pine branches, and rice-paper texture.
func NewInkLandscape(mountainLayers int, roughness float64) *inkLandscape {
	if mountainLayers <= 0 {
		mountainLayers = 8
	}
	if roughness <= 0 {
		roughness = 0.45
	}
	return &inkLandscape{
		mountainLayers: mountainLayers,
		roughness:      roughness,
	}
}

// drawBotanicalBranch recursively grows a calligraphic plum/pine silhouette branch
// using multi-filament bristle strokes and glowing crimson/blush blossoms.
func (il *inkLandscape) drawBotanicalBranch(
	rc *render.Canvas,
	start render.Point,
	angle, length, thickness float64,
	depth int,
	inkCol, blossomOuter, blossomInner color.RGBA,
	scale float64,
	rng *rand.Rand,
) {
	if depth <= 0 || length < 3.0*scale || thickness < 0.35*scale {
		return
	}

	// Build a slightly gnarled organic control polyline for this branch segment
	const segs = 5
	pts := make([]render.Point, segs+1)
	pts[0] = start
	cur := start
	curAngle := angle
	stepLen := length / float64(segs)

	for i := 1; i <= segs; i++ {
		curAngle += (rng.Float64() - 0.48) * 0.18
		cur = render.Pt(
			cur.X+math.Cos(curAngle)*stepLen,
			cur.Y+math.Sin(curAngle)*stepLen,
		)
		pts[i] = cur
	}

	rc.SetBlendMode(render.BlendOver)
	bristles := int(math.Max(5, math.Min(16, math.Round(thickness*2.2))))
	rc.DrawBristleStroke(pts, thickness, bristles, inkCol, 0.86, rng)

	// Reinforce core ink spine with calligraphic taper
	rc.DrawTaperedSpline(
		pts,
		thickness*0.65,
		render.TaperTail,
		func(float64) color.RGBA { return inkCol },
		0.72,
	)

	// Spawn blossoms near outer twigs
	if depth <= 3 {
		blossomCount := 2 + rng.Intn(3)
		for b := 0; b < blossomCount; b++ {
			anchor := pts[1+rng.Intn(len(pts)-1)]
			bx := anchor.X + rng.NormFloat64()*5.5*scale
			by := anchor.Y + rng.NormFloat64()*5.5*scale
			rad := (2.4 + rng.Float64()*4.2) * scale
			rc.DrawGlowDot(bx, by, rad, blossomOuter, 0.68, 0.55)
			rc.DrawGlowDot(bx, by, rad*0.42, blossomInner, 0.82, 0.75)
		}
	}

	// Main continuation + lateral sub-branches
	end := pts[len(pts)-1]
	spread := 0.34 + 0.22*rng.Float64()
	il.drawBotanicalBranch(
		rc, end,
		curAngle-spread*(0.65+0.5*rng.Float64()),
		length*(0.68+0.14*rng.Float64()),
		thickness*0.68,
		depth-1,
		inkCol, blossomOuter, blossomInner, scale, rng,
	)
	il.drawBotanicalBranch(
		rc, end,
		curAngle+spread*(0.65+0.5*rng.Float64()),
		length*(0.64+0.15*rng.Float64()),
		thickness*0.64,
		depth-1,
		inkCol, blossomOuter, blossomInner, scale, rng,
	)
	if depth >= 3 && rng.Float64() < 0.45 {
		mid := pts[len(pts)/2]
		sideSign := 1.0
		if rng.Float64() < 0.5 {
			sideSign = -1.0
		}
		il.drawBotanicalBranch(
			rc, mid,
			curAngle+sideSign*(0.45+0.25*rng.Float64()),
			length*0.52,
			thickness*0.52,
			depth-2,
			inkCol, blossomOuter, blossomInner, scale, rng,
		)
	}
}

// Generative renders an East-Asian Shan-Shui ink-and-mineral wash landscape into c.
func (il *inkLandscape) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	width, height := c.Width(), c.Height()
	if width <= 0 || height <= 0 {
		return
	}
	fw, fh := float64(width), float64(height)
	scale := math.Min(fw, fh) / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	seed := rand.Int63()
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	// Traditional Shan-Shui rice-paper and mineral-ink palette
	skyTop := color.RGBA{R: 244, G: 236, B: 220, A: 255}
	skyWarm := color.RGBA{R: 236, G: 222, B: 198, A: 255}
	mistCol := color.RGBA{R: 240, G: 235, B: 225, A: 255}
	distantMineral := color.RGBA{R: 92, G: 118, B: 126, A: 255}
	midJade := color.RGBA{R: 54, G: 82, B: 88, A: 255}
	deepInk := color.RGBA{R: 22, G: 26, B: 32, A: 255}

	schema := c.Opts().ColorSchema()
	if len(schema) >= 2 {
		grad := render.NewMultiStopGradient(schema...)
		distantMineral = render.LerpOklab(distantMineral, grad.SampleOklab(0.35), 0.35)
		midJade = render.LerpOklab(midJade, grad.SampleOklab(0.65), 0.28)
	}

	// 1. Warm Xuan rice-paper sky gradient backdrop
	for y := 0; y < height; y++ {
		ty := float64(y) / math.Max(1.0, fh-1.0)
		col := render.LerpOklab(skyTop, skyWarm, math.Pow(ty, 0.85))
		rLin, gLin, bLin, _ := render.ColorToLinearRGB(col)
		for x := 0; x < width; x++ {
			rc.SetPixelLinear(x, y, rLin, gLin, bLin, 1.0)
		}
	}

	// 2. Soft glowing crimson/amber sun disc rendered with DrawWatercolorPolygon + DrawGlowDot
	sunX := fw * (0.62 + (rng.Float64()-0.5)*0.12)
	sunY := fh * (0.23 + (rng.Float64()-0.5)*0.05)
	sunR := 46.0 * scale

	crimsonSun := color.RGBA{R: 198, G: 58, B: 42, A: 255}
	amberGlow := color.RGBA{R: 238, G: 138, B: 68, A: 255}

	rc.SetBlendMode(render.BlendOver)
	rc.DrawGlowDot(sunX, sunY, sunR*2.4, amberGlow, 0.24, 0.0)

	sunPoly := make([]render.Point, 20)
	for i := 0; i < 20; i++ {
		ang := (float64(i) / 20.0) * 2.0 * math.Pi
		r := sunR * (0.94 + 0.1*rng.Float64())
		sunPoly[i] = render.Pt(sunX+r*math.Cos(ang), sunY+r*math.Sin(ang))
	}
	rc.SetBlendMode(render.BlendMultiply)
	rc.DrawWatercolorPolygon(sunPoly, crimsonSun, 14, 4, 0.18, 0.065, rng)
	rc.SetBlendMode(render.BlendOver)
	rc.DrawGlowDot(sunX, sunY, sunR*0.92, crimsonSun, 0.45, 0.65)

	// 3. Distant to foreground mountain ridgelines with vertical ink-wash dilution & atmospheric mist
	inkMountainGrad := render.NewMultiStopGradient(distantMineral, midJade, deepInk)
	mistLinR, mistLinG, mistLinB, _ := render.ColorToLinearRGB(mistCol)

	denom := float64(il.mountainLayers - 1)
	if denom < 1 {
		denom = 1
	}

	for layer := 0; layer < il.mountainLayers; layer++ {
		depthT := float64(layer) / denom // 0 = farthest peak, 1 = closest foreground ridge
		baseY := fh * (0.28 + 0.58*math.Pow(depthT, 0.92))
		amp := fh * (0.19 - 0.08*depthT) * (0.75 + 0.6*il.roughness)
		freq := (1.6 + depthT*2.2) / fw
		layerSeedX := float64(layer)*37.3 + 11.7
		layerSeedY := float64(layer)*19.1 + 5.3

		ridgeY := make([]float64, width)
		for x := 0; x < width; x++ {
			fx := float64(x)
			ridged := noise.Ridged2D(fx*freq+layerSeedX, layerSeedY, 5, 2.05, 0.52)
			fbm := noise.FBM2D(fx*freq*1.8+layerSeedX*0.7, layerSeedY+3.1, 4, 2.0, 0.5)
			envelope := 0.72 + 0.28*math.Sin((fx/fw)*math.Pi*(1.2+0.5*float64(layer%3))+float64(layer)*1.1)
			offset := (0.68*ridged + 0.32*(0.5*(fbm+1.0))) * amp * envelope
			ridgeY[x] = common.Constrain(baseY-offset, 4.0, fh-2.0)
		}

		// Layer mineral/ink crest color darkened progressively toward foreground
		layerCol := inkMountainGrad.SampleOklab(math.Pow(depthT, 0.85))
		layerLab := render.RGBAToOklab(layerCol)
		crestAlpha := common.Constrain(0.38+0.56*math.Pow(depthT, 0.8), 0, 0.96)
		fadeSpan := fh * (0.16 + 0.22*(1.0-depthT*0.45))

		// Vertical ink-wash body fill with atmospheric mist dilution at the base of the ridge
		for x := 0; x < width; x++ {
			fx := float64(x)
			topY := int(math.Floor(ridgeY[x]))
			if topY < 0 {
				topY = 0
			}
			// Internal vertical texture streak modulation (Simplex noise)
			tex := noise.FBM2D(fx*0.012+layerSeedX, float64(topY)*0.008, 3, 2.0, 0.5)

			for y := topY; y < height; y++ {
				distDown := float64(y) - ridgeY[x]
				if distDown < 0 {
					continue
				}
				// Antialiased 1.5px top edge feather
				edgeCov := common.Constrain((distDown+1.0)/1.5, 0, 1)

				// Normalized vertical dilution u in [0, 1] from crest to misty valley base
				u := common.Constrain(distDown/math.Max(12.0, fadeSpan), 0, 1)
				washCurve := math.Pow(1.0-u, 1.45)

				// Modulate ink lightness slightly for Shan-Shui mineral wash texture
				localL := common.Constrain(layerLab.L+(1.0-washCurve)*0.32+tex*0.035, 0.04, 0.96)
				washCol := render.OklabToRGBA(render.Oklab{
					L: localL,
					A: layerLab.A * (0.4 + 0.6*washCurve),
					B: layerLab.B * (0.4 + 0.6*washCurve),
				}, 255)
				rLin, gLin, bLin, _ := render.ColorToLinearRGB(washCol)

				// Blend ink wash near crest and misty fog near base so successive layers separate cleanly
				mistWeight := math.Pow(u, 0.85) * (0.85 - 0.35*depthT)
				outR := rLin*(1.0-mistWeight) + mistLinR*mistWeight
				outG := gLin*(1.0-mistWeight) + mistLinG*mistWeight
				outB := bLin*(1.0-mistWeight) + mistLinB*mistWeight

				alpha := edgeCov * (crestAlpha*washCurve + 0.72*mistWeight)
				rc.BlendPixelLinear(x, y, outR, outG, outB, common.Constrain(alpha, 0, 0.95), render.BlendOver)
			}
		}

		// Stroke the ridgeline crest using DrawBristleStroke + DrawTaperedSpline for dry-brush Cufa texture
		const ctrlStep = 10
		crestPts := make([]render.Point, 0, width/ctrlStep+2)
		for x := 0; x < width; x += ctrlStep {
			crestPts = append(crestPts, render.Pt(float64(x), ridgeY[x]))
		}
		if crestPts[len(crestPts)-1].X < fw-1 {
			crestPts = append(crestPts, render.Pt(fw-1, ridgeY[width-1]))
		}

		crestInk := render.OklabToRGBA(render.Oklab{
			L: common.Constrain(layerLab.L*0.62, 0.04, 0.55),
			A: layerLab.A * 0.8,
			B: layerLab.B * 0.8,
		}, 255)

		rc.SetBlendMode(render.BlendOver)
		strokeW := (1.6 + 3.2*depthT) * scale
		rc.DrawBristleStroke(crestPts, strokeW*1.35, 10, crestInk, 0.32+0.42*depthT, rng)
		rc.DrawTaperedSpline(
			crestPts,
			strokeW*0.75,
			render.TaperCalligraphic,
			func(float64) color.RGBA { return crestInk },
			0.35+0.45*depthT,
		)
	}

	// 4. Distant calligraphic flying birds near the sun
	birdInk := color.RGBA{R: 38, G: 42, B: 48, A: 255}
	birdCount := 5
	for i := 0; i < birdCount; i++ {
		bx := sunX + (float64(i)-2.0)*22.0*scale + rng.NormFloat64()*6.0*scale
		by := sunY - 18.0*scale + float64(i%3)*12.0*scale + rng.NormFloat64()*4.0*scale
		span := (6.5 + rng.Float64()*5.5) * scale
		lift := (2.5 + rng.Float64()*2.5) * scale
		wingPts := []render.Point{
			{X: bx - span, Y: by - lift*0.6},
			{X: bx - span*0.45, Y: by - lift},
			{X: bx, Y: by + lift*0.35},
			{X: bx + span*0.45, Y: by - lift},
			{X: bx + span, Y: by - lift*0.5},
		}
		rc.DrawTaperedSpline(
			wingPts,
			1.45*scale,
			render.TaperBothEnds,
			func(float64) color.RGBA { return birdInk },
			0.68,
		)
	}

	// 5. Foreground recursive botanical plum/pine branch from the lower edge
	branchInk := color.RGBA{R: 18, G: 20, B: 25, A: 255}
	blossomOuter := color.RGBA{R: 214, G: 68, B: 72, A: 255}
	blossomInner := color.RGBA{R: 252, G: 224, B: 210, A: 255}

	rootX := 0.0
	rootY := fh * (0.76 + rng.Float64()*0.1)
	rootAngle := -0.32 + rng.Float64()*0.12
	if rng.Float64() < 0.35 {
		rootX = fw
		rootAngle = -math.Pi + 0.32 - rng.Float64()*0.12
	}
	il.drawBotanicalBranch(
		rc,
		render.Pt(rootX, rootY),
		rootAngle,
		105.0*scale,
		7.2*scale,
		5,
		branchInk,
		blossomOuter,
		blossomInner,
		scale,
		rng,
	)

	// 6. Finish with Xuan paper fiber texture and subtle vignette
	rc.ApplyPaperTexture(0.5, 888)
	rc.ApplyVignette(0.2)
	c.MergeRenderCanvas(rc)
}
