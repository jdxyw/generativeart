package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type orphicDiscs struct {
	discCount int
}

// NewOrphicDiscs returns an orphicDiscs engine that renders Robert Delaunay Orphism
// concentric radial mosaic discs with angular harmonic color waves and feathered starburst rims.
func NewOrphicDiscs(discCount int) *orphicDiscs {
	if discCount <= 0 {
		discCount = 34
	}
	return &orphicDiscs{
		discCount: discCount,
	}
}

type orphicDiscSpec struct {
	cx, cy    float64
	maxRadius float64
	ringCount int
	tier      int // 0: giant edge-anchored, 1: large hero, 2: medium, 3: satellite
}

// Generative renders the Robert Delaunay Orphism Radial Mosaic Discs composition.
func (od *orphicDiscs) Generative(c *generativeart.Canva) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	discCount := od.discCount
	if discCount <= 0 {
		discCount = 34
	}

	seed := rand.Int63()
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	rc := c.NewRenderCanvas()

	// Rich Orphist palette
	emeraldBg := color.RGBA{R: 72, G: 178, B: 102, A: 255}
	emeraldDeep := color.RGBA{R: 42, G: 142, B: 78, A: 255}
	cadmiumYellow := color.RGBA{R: 252, G: 212, B: 42, A: 255}
	hotMagenta := color.RGBA{R: 234, G: 48, B: 136, A: 255}
	blushPink := color.RGBA{R: 246, G: 178, B: 208, A: 255}
	vermilionRed := color.RGBA{R: 228, G: 62, B: 42, A: 255}
	ultramarine := color.RGBA{R: 24, G: 58, B: 148, A: 255}
	cerulean := color.RGBA{R: 48, G: 148, B: 208, A: 255}
	crispWhite := color.RGBA{R: 248, G: 248, B: 242, A: 255}
	darkSep := color.RGBA{R: 20, G: 28, B: 42, A: 255}

	palette := []color.RGBA{
		cadmiumYellow,
		hotMagenta,
		blushPink,
		vermilionRed,
		ultramarine,
		cerulean,
		emeraldDeep,
		crispWhite,
	}

	schema := c.Opts().ColorSchema()
	if isCustomSchema(schema) {
		userGrad := render.NewMultiStopGradient(schema...)
		for i := range palette {
			t := float64(i) / float64(len(palette)-1)
			palette[i] = render.LerpOklab(palette[i], userGrad.SampleOklab(t), 0.32)
		}
	}

	// Vibrant emerald-meadow green background {72, 178, 102, 255} with soft tonal variation
	for py := 0; py < h; py++ {
		ny := float64(py) / math.Max(1.0, fh-1.0)
		for px := 0; px < w; px++ {
			nx := float64(px) / math.Max(1.0, fw-1.0)
			fbm := 0.5 + 0.5*noise.FBM2D(nx*2.6+1.4, ny*2.6-0.9, 3, 2.0, 0.5)
			col := render.LerpOklab(emeraldDeep, emeraldBg, 0.35+0.65*fbm)
			warmLift := common.Smoothstep(0.2, 0.85, 1.0-math.Hypot(nx-0.48, ny-0.46))
			col = render.LerpOklab(col, cadmiumYellow, warmLift*0.09)
			rLin, gLin, bLin, _ := render.ColorToLinearRGB(col)
			rc.SetPixelLinear(px, py, rLin, gLin, bLin, 1.0)
		}
	}

	// Build multi-scale Orphic discs:
	// - 2 giant partial discs anchored on the bottom/top edges
	// - 3 large hero discs of radius 110..165px
	// - 10 medium discs of radius 50..95px
	// - remaining smaller satellite discs of radius 18..45px
	discs := make([]orphicDiscSpec, 0, discCount+2)

	// 2 giant partial discs anchored on the bottom and top edges
	discs = append(discs,
		orphicDiscSpec{
			cx:        fw * (0.36 + 0.14*rng.Float64()),
			cy:        fh * (1.03 + 0.05*rng.Float64()),
			maxRadius: (175.0 + 35.0*rng.Float64()) * scale,
			ringCount: 8,
			tier:      0,
		},
		orphicDiscSpec{
			cx:        fw * (0.62 + 0.14*rng.Float64()),
			cy:        -fh * (0.04 + 0.04*rng.Float64()),
			maxRadius: (160.0 + 30.0*rng.Float64()) * scale,
			ringCount: 8,
			tier:      0,
		},
	)

	heroCount := 3
	if discCount < 8 {
		heroCount = common.MaxInt(1, discCount/3)
	}
	medCount := 10
	if discCount-heroCount < medCount {
		medCount = common.MaxInt(0, (discCount-heroCount)/2)
	}
	satCount := common.MaxInt(0, discCount-heroCount-medCount)

	// 3 large hero discs (radius 110..165px) placed in balanced triangular composition
	heroAnchors := [][2]float64{
		{0.34, 0.42},
		{0.68, 0.36},
		{0.52, 0.68},
	}
	for i := 0; i < heroCount; i++ {
		anchor := heroAnchors[i%len(heroAnchors)]
		cx := fw * (anchor[0] + (rng.Float64()-0.5)*0.12)
		cy := fh * (anchor[1] + (rng.Float64()-0.5)*0.12)
		rad := (110.0 + rng.Float64()*55.0) * scale
		discs = append(discs, orphicDiscSpec{
			cx:        cx,
			cy:        cy,
			maxRadius: rad,
			ringCount: 7 + rng.Intn(3), // 7..9 rings
			tier:      1,
		})
	}

	// 10 medium discs (radius 50..95px)
	for i := 0; i < medCount; i++ {
		angle := (float64(i)/math.Max(1.0, float64(medCount)))*2.0*math.Pi + rng.Float64()*0.45
		dist := (0.16 + 0.28*rng.Float64()) * math.Min(fw, fh)
		cx := fw*0.5 + math.Cos(angle)*dist + (rng.Float64()-0.5)*45.0*scale
		cy := fh*0.5 + math.Sin(angle)*dist + (rng.Float64()-0.5)*45.0*scale
		rad := (50.0 + rng.Float64()*45.0) * scale
		discs = append(discs, orphicDiscSpec{
			cx:        cx,
			cy:        cy,
			maxRadius: rad,
			ringCount: 5 + rng.Intn(3), // 5..7 rings
			tier:      2,
		})
	}

	// Smaller satellite discs (radius 18..45px) with mild blue-noise separation
	satCenters := make([]render.Point, 0, satCount)
	for i := 0; i < satCount; i++ {
		bestPt := render.Pt(
			(0.08+0.84*rng.Float64())*fw,
			(0.08+0.84*rng.Float64())*fh,
		)
		bestDist := -1.0
		for cand := 0; cand < 8; cand++ {
			pt := render.Pt(
				(0.06+0.88*rng.Float64())*fw,
				(0.06+0.88*rng.Float64())*fh,
			)
			minD := math.MaxFloat64
			for _, prev := range satCenters {
				d := pt.Dist(prev)
				if d < minD {
					minD = d
				}
			}
			if minD > bestDist {
				bestDist = minD
				bestPt = pt
			}
		}
		satCenters = append(satCenters, bestPt)
		rad := (18.0 + rng.Float64()*27.0) * scale
		discs = append(discs, orphicDiscSpec{
			cx:        bestPt.X,
			cy:        bestPt.Y,
			maxRadius: rad,
			ringCount: 4 + rng.Intn(3), // 4..6 rings
			tier:      3,
		})
	}

	coreColors := []color.RGBA{
		cadmiumYellow,
		hotMagenta,
		ultramarine,
		vermilionRed,
		blushPink,
	}

	// Render each disc with concentric mosaic rings, separator circles, and solid center core
	for dIdx, d := range discs {
		coreR := d.maxRadius * (0.15 + 0.07*rng.Float64())
		rings := common.ConstrainInt(d.ringCount, 4, 9)

		// Compute monotonically increasing ring boundary radii from coreR to d.maxRadius
		ringRadii := make([]float64, rings+1)
		ringRadii[0] = coreR
		weights := make([]float64, rings)
		wSum := 0.0
		for r := 0; r < rings; r++ {
			wRing := 0.65 + 0.7*rng.Float64()
			if r == rings-1 {
				wRing *= 1.18 // slightly wider outer starburst ring
			}
			weights[r] = wRing
			wSum += wRing
		}
		acc := coreR
		availSpan := d.maxRadius - coreR
		for r := 0; r < rings; r++ {
			acc += availSpan * (weights[r] / wSum)
			ringRadii[r+1] = acc
		}

		// Base angular sector subdivision (36..96 sectors depending on disc radius)
		baseSectors := int(math.Round(36.0 + 60.0*common.Clamp01((d.maxRadius/scale-18.0)/155.0)))
		baseSectors = (baseSectors / 4) * 4
		baseSectors = common.ConstrainInt(baseSectors, 36, 96)

		discPhase := rng.Float64() * 2.0 * math.Pi
		paletteOffset := rng.Intn(len(palette))

		// 1. Draw 4–9 concentric annular rings around the solid center core
		for r := 0; r < rings; r++ {
			rInner := ringRadii[r]
			rOuter := ringRadii[r+1]
			isOutermost := r == rings-1

			sectors := baseSectors
			if r < rings/2 && sectors > 48 {
				sectors = common.MaxInt(36, (sectors*3)/4)
			}
			dTheta := (2.0 * math.Pi) / float64(sectors)
			checkerShift := float64(r%2) * 0.5 * dTheta
			ringFreq := float64(2 + (r+dIdx)%5)

			// Pick 3 primary harmonic colors for this ring
			cA := palette[(paletteOffset+r*2)%len(palette)]
			cB := palette[(paletteOffset+r*2+3)%len(palette)]
			cC := palette[(paletteOffset+r*3+5)%len(palette)]

			for s := 0; s < sectors; s++ {
				theta1 := float64(s)*dTheta + checkerShift
				theta2 := theta1 + dTheta*1.015 // tiny angular overlap prevents sub-pixel radial hairline gaps
				midAng := 0.5 * (theta1 + theta2)

				// Rhythmic harmonic wave pattern + mosaic checker phase shift between rings
				wave := 0.5 + 0.5*math.Sin(midAng*ringFreq+discPhase+float64(r)*0.95)
				checker := (s + r) % 3

				var secCol color.RGBA
				switch checker {
				case 0:
					secCol = render.LerpOklab(cA, cB, wave)
				case 1:
					secCol = render.LerpOklab(cB, cC, 1.0-wave)
				default:
					secCol = render.LerpOklab(cC, cA, 0.35+0.65*wave)
				}

				// Stained-glass prism intersection blend modes on alternating outer rings and sector beats
				if isOutermost {
					// Outermost ring: feather the radial tips (varying rOuter per sector with radial alpha falloff)
					// to create the radial motion-blur / starburst fringe
					rc.SetBlendMode(render.BlendOver)
					tipFactor := 0.88 + 0.24*(0.5+0.5*math.Sin(float64(s)*1.7+discPhase)) + 0.14*rng.Float64()
					featheredOuter := rInner + (rOuter-rInner)*tipFactor
					midR := rInner + (featheredOuter-rInner)*0.58

					// Solid inner half of outer ring
					rc.DrawAnnularSector(d.cx, d.cy, rInner, midR, theta1, theta2, secCol, 0.90)
					// Translucent feathered outer tip
					if s%4 == 0 {
						rc.SetBlendMode(render.BlendScreen)
					}
					rc.DrawAnnularSector(d.cx, d.cy, midR, featheredOuter, theta1, theta2, secCol, 0.48)
				} else {
					alpha := 0.92
					if r >= rings-2 && (s+dIdx)%5 == 0 {
						rc.SetBlendMode(render.BlendOverlay)
						alpha = 0.78
					} else if r >= rings-3 && (s+r)%7 == 0 {
						rc.SetBlendMode(render.BlendScreen)
						alpha = 0.68
					} else {
						rc.SetBlendMode(render.BlendOver)
					}
					rc.DrawAnnularSector(d.cx, d.cy, rInner, rOuter, theta1, theta2, secCol, alpha)
				}
			}
		}

		// 2. Draw crisp thin dark or white concentric separator circles (rc.StrokePath, width 0.9px)
		rc.SetBlendMode(render.BlendOver)
		sepWidth := 0.9 * scale
		for r := 0; r < rings; r++ {
			if r > 0 && r%2 == 1 && rng.Float64() < 0.35 {
				continue
			}
			rad := ringRadii[r]
			circlePath := canvas.Circle(rad).Translate(d.cx, d.cy)
			sepCol := darkSep
			sepAlpha := 0.65
			if (r+dIdx)%2 == 0 {
				sepCol = crispWhite
				sepAlpha = 0.72
			}
			rc.StrokePath(circlePath, sepWidth, canvas.RoundCap, canvas.RoundJoin, sepCol, sepAlpha)
		}

		// 3. Draw the solid high-saturation inner core disc at the center
		coreCol := coreColors[(dIdx+paletteOffset)%len(coreColors)]
		corePath := canvas.Circle(coreR).Translate(d.cx, d.cy)
		rc.RasterizePath(corePath, coreCol, 0.98)

		// Inner accent jewel ring inside larger discs
		if coreR > 12.0*scale {
			innerCoreCol := coreColors[(dIdx+paletteOffset+2)%len(coreColors)]
			innerPath := canvas.Circle(coreR * 0.46).Translate(d.cx, d.cy)
			rc.RasterizePath(innerPath, innerCoreCol, 0.96)
		}
	}

	rc.ApplyPaperTexture(0.22, 404)
	c.MergeRenderCanvas(rc)
}
