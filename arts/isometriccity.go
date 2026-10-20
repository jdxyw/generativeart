package arts

import (
	"image/color"
	"math"
	"math/rand"
	"sort"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type isometricCity struct {
	gridN       int
	heightScale float64
}

type isoCell struct {
	gx, gy    int
	height    float64
	tiers     int
	roofStyle int // 0: stepped flat terrace, 1: monument dome, 2: obelisk spire, 3: rooftop garden tree
	colorT    float64
	isCanal   bool
}

// NewIsometricCity returns an isometricCity engine that renders a Monument-Valley / Brutalist
// 2.5D axonometric architectural citadel with stepped ziggurats, sky-bridges, colonnade windows,
// and warm twilight Oklab lighting.
func NewIsometricCity(gridN int, heightScale float64) *isometricCity {
	if gridN <= 0 {
		gridN = 11
	}
	if heightScale <= 0 {
		heightScale = 1.0
	}
	return &isometricCity{
		gridN:       gridN,
		heightScale: heightScale,
	}
}

// drawIsoBox renders a 3D isometric rectangular prism spanning [x0, x1] x [y0, y1] x [z0, z1]
// with Oklab-shaded top, left (front-left: x=x1), and right (front-right: y=y1) faces,
// and appends crisp edge segments to edgePath and rimPath.
func drawIsoBox(
	rc *render.Canvas,
	iso func(gx, gy, gz float64) render.Point,
	x0, y0, z0, x1, y1, z1 float64,
	baseCol color.RGBA,
	edgePath, rimPath *canvas.Path,
) {
	p001 := iso(x0, y0, z1) // top-back
	p101 := iso(x1, y0, z1) // top-left
	p111 := iso(x1, y1, z1) // top-front
	p011 := iso(x0, y1, z1) // top-right

	p100 := iso(x1, y0, z0) // bottom-left
	p110 := iso(x1, y1, z0) // bottom-front
	p010 := iso(x0, y1, z0) // bottom-right

	lab := render.RGBAToOklab(baseCol)

	// Sunlit top diamond face: warm lifted lightness
	topCol := render.OklabToRGBA(render.Oklab{
		L: common.Constrain(lab.L+0.14, 0.18, 0.96),
		A: lab.A + 0.012,
		B: lab.B + 0.024,
	}, 255)

	// Warm mid-tone left face (x = x1, facing sun at azimuth -left)
	leftCol := render.OklabToRGBA(render.Oklab{
		L: common.Constrain(lab.L+0.01, 0.12, 0.88),
		A: lab.A + 0.018,
		B: lab.B + 0.014,
	}, 255)

	// Cool shadowed right face (y = y1, facing twilight shadow)
	rightCol := render.OklabToRGBA(render.Oklab{
		L: common.Constrain(lab.L-0.16, 0.07, 0.72),
		A: lab.A - 0.006,
		B: lab.B - 0.032,
	}, 255)

	rc.SetBlendMode(render.BlendOver)
	rc.FillPolygon([]render.Point{p101, p111, p110, p100}, leftCol, 1.0)
	rc.FillPolygon([]render.Point{p111, p011, p010, p110}, rightCol, 1.0)
	rc.FillPolygon([]render.Point{p001, p101, p111, p011}, topCol, 1.0)

	if edgePath != nil {
		// Top diamond contour
		edgePath.MoveTo(p001.X, p001.Y)
		edgePath.LineTo(p101.X, p101.Y)
		edgePath.LineTo(p111.X, p111.Y)
		edgePath.LineTo(p011.X, p011.Y)
		edgePath.Close()

		// Left & right silhouettes + central vertical spine
		edgePath.MoveTo(p101.X, p101.Y)
		edgePath.LineTo(p100.X, p100.Y)
		edgePath.LineTo(p110.X, p110.Y)
		edgePath.LineTo(p010.X, p010.Y)
		edgePath.LineTo(p011.X, p011.Y)

		edgePath.MoveTo(p111.X, p111.Y)
		edgePath.LineTo(p110.X, p110.Y)
	}

	if rimPath != nil {
		// Sunlit top-front bevel highlight
		rimPath.MoveTo(p101.X, p101.Y)
		rimPath.LineTo(p111.X, p111.Y)
		rimPath.LineTo(p011.X, p011.Y)
	}
}

// Generative renders the Monument-Valley / Brutalist 2.5D axonometric architectural utopia.
func (ic *isometricCity) Generative(c *generativeart.Canva) {
	gridN := ic.gridN
	if gridN <= 0 {
		gridN = 11
	}
	gridN = common.ConstrainInt(gridN, 5, 18)

	hScale := ic.heightScale
	if hScale <= 0 {
		hScale = 1.0
	}
	hScale = common.Constrain(hScale, 0.4, 2.2)

	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	seed := rand.Int63()
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	schema := c.Opts().ColorSchema()
	if len(schema) < 2 {
		schema = common.Sunset
	}
	grad := render.NewMultiStopGradient(schema...)
	ukiyoGrad := render.NewMultiStopGradient(common.UkiyoE...)

	rc := c.NewRenderCanvas()

	// 1. Warm dusk / twilight sky gradient backdrop
	skyZenith := render.LerpOklab(color.RGBA{R: 22, G: 28, B: 52, A: 255}, grad.SampleOklab(0.06), 0.35)
	skyMid := render.LerpOklab(color.RGBA{R: 68, G: 52, B: 88, A: 255}, grad.SampleOklab(0.45), 0.45)
	skyHorizon := render.LerpOklab(color.RGBA{R: 244, G: 162, B: 108, A: 255}, grad.SampleOklab(0.78), 0.45)
	skyLow := render.LerpOklab(color.RGBA{R: 36, G: 44, B: 62, A: 255}, ukiyoGrad.SampleOklab(0.15), 0.5)

	moonX := fw * 0.24
	moonY := fh * 0.20
	moonR := 52.0 * scale

	for py := 0; py < h; py++ {
		vy := float64(py) / math.Max(1.0, fh-1.0)
		var rowCol color.RGBA
		if vy < 0.46 {
			u := vy / 0.46
			rowCol = render.LerpOklab(skyZenith, skyMid, math.Pow(u, 0.85))
		} else if vy < 0.72 {
			u := (vy - 0.46) / 0.26
			rowCol = render.LerpOklab(skyMid, skyHorizon, common.Smoothstep(0, 1, u))
		} else {
			u := (vy - 0.72) / 0.28
			rowCol = render.LerpOklab(skyHorizon, skyLow, common.Smoothstep(0, 1, u))
		}

		for px := 0; px < w; px++ {
			fx := float64(px)
			fy := float64(py)
			dMoon := math.Hypot(fx-moonX, fy-moonY) / math.Max(1.0, fw*0.55)
			glowT := math.Exp(-3.2 * dMoon * dMoon)
			col := render.LerpOklab(rowCol, color.RGBA{R: 255, G: 222, B: 168, A: 255}, 0.26*glowT)
			rLin, gLin, bLin, _ := render.ColorToLinearRGB(col)
			rc.SetPixelLinear(px, py, rLin, gLin, bLin, 1.0)
		}
	}

	// Fine twilight stars in upper sky + giant low-hanging moon with halo rings
	rc.SetBlendMode(render.BlendAdd)
	starCol := color.RGBA{R: 255, G: 246, B: 228, A: 255}
	for s := 0; s < 95; s++ {
		sx := rng.Float64() * fw
		sy := rng.Float64() * fh * 0.48
		if math.Hypot(sx-moonX, sy-moonY) < moonR*1.35 {
			continue
		}
		rad := (0.65 + 1.35*math.Pow(rng.Float64(), 2.2)) * scale
		alpha := 0.22 + 0.55*rng.Float64()
		rc.DrawGlowDot(sx, sy, rad, starCol, alpha, 0.35)
	}

	moonWarm := color.RGBA{R: 255, G: 236, B: 192, A: 255}
	moonCore := color.RGBA{R: 255, G: 250, B: 235, A: 255}
	rc.DrawGlowDot(moonX, moonY, moonR*2.1, moonWarm, 0.24, 0.05)
	rc.DrawSoftRing(moonX, moonY, moonR*1.34, 3.2*scale, moonWarm, 0.32)
	rc.DrawSoftRing(moonX, moonY, moonR*1.04, 2.0*scale, moonCore, 0.55)
	rc.SetBlendMode(render.BlendOver)
	rc.DrawGlowDot(moonX, moonY, moonR, moonCore, 0.92, 0.88)
	// Subtle crescent shadow bite on the upper-right of the moon
	crescentShade := render.LerpOklab(skyZenith, skyMid, 0.45)
	rc.DrawGlowDot(moonX+moonR*0.34, moonY-moonR*0.18, moonR*0.82, crescentShade, 0.58, 0.72)

	// 2. Define isometric projection Iso(gx, gy, gz)
	fn := float64(gridN)
	tileW := (fw * 0.92) / fn
	tileH := tileW * 0.54
	zScale := tileW * 0.92 * hScale
	cx := fw * 0.50
	baseY := fh*0.55 - fn*tileH*0.12

	iso := func(gx, gy, gz float64) render.Point {
		sx := cx + (gx-gy)*tileW*0.5
		sy := baseY + (gx+gy)*tileH*0.5 - gz*zScale
		return render.Pt(sx, sy)
	}

	// 3. Build city grid cells using terraced Ridged2D + radial citadel profile
	grid := make([][]isoCell, gridN)
	cells := make([]isoCell, 0, gridN*gridN)
	mid := float64(gridN-1) * 0.5
	canalIndex := gridN / 2

	for gx := 0; gx < gridN; gx++ {
		grid[gx] = make([]isoCell, gridN)
		for gy := 0; gy < gridN; gy++ {
			dx := (float64(gx) - mid) / math.Max(1.0, mid)
			dy := (float64(gy) - mid) / math.Max(1.0, mid)
			rad := math.Hypot(dx, dy)

			// Cross-canal avenue passing near the citadel court, with central bridge crossing
			isCanal := (gx == canalIndex && math.Abs(float64(gy)-mid) > 1.2) ||
				(gy == canalIndex && math.Abs(float64(gx)-mid) > 1.8)

			ridged := common.Clamp01(noise.Ridged2D(float64(gx)*0.36+1.7, float64(gy)*0.36+3.1, 3, 2.0, 0.5))
			fbm := common.Clamp01(0.5 + 0.65*noise.FBM2D(float64(gx)*0.32+5.3, float64(gy)*0.32+2.9, 3, 2.0, 0.5))
			hash := float64((gx*17+gy*31)%19) / 18.0

			// Radial citadel profile: grand central acropolis spires tapering to outer terraces
			citadelEnvelope := math.Max(0.22, 1.35-0.88*math.Pow(rad, 1.05))
			rawH := (0.45 + 3.65*(0.45*ridged+0.35*fbm+0.20*hash)) * citadelEnvelope

			// Quantize into stepped architectural terraces
			quantH := math.Round(rawH*4.0) / 4.0
			if isCanal {
				quantH = 0
			} else if math.Abs(float64(gx)-mid) <= 1.1 && math.Abs(float64(gy)-mid) <= 1.1 {
				// Central acropolis keep & surrounding high towers
				if gx == canalIndex && gy == canalIndex {
					quantH = 4.6
				} else {
					quantH = math.Max(quantH, 2.8+0.9*hash)
				}
			} else if rad < 0.62 {
				quantH = math.Max(quantH, 1.6+1.4*hash)
			}

			tiers := 2
			if quantH >= 2.6 {
				tiers = 4
			} else if quantH >= 1.35 {
				tiers = 3
			}

			roofStyle := 0
			if quantH >= 2.1 {
				if (gx+gy)%2 == 0 {
					roofStyle = 1 // Monument dome
				} else {
					roofStyle = 2 // Obelisk spire
				}
			} else if quantH >= 0.7 && (gx*3+gy*5)%3 == 0 {
				roofStyle = 3 // Rooftop miniature cypress/tree
			}

			// Spread colorT across [0.05, 0.95] so districts showcase rich indigo, crimson, terracotta, and gold
			cT := common.Clamp01(0.06 + 0.52*hash + 0.28*fbm + 0.14*(1.0-common.Clamp01(rad)))
			cell := isoCell{
				gx:        gx,
				gy:        gy,
				height:    quantH,
				tiers:     tiers,
				roofStyle: roofStyle,
				colorT:    cT,
				isCanal:   isCanal,
			}
			grid[gx][gy] = cell
			cells = append(cells, cell)
		}
	}

	// Sort cells in strict back-to-front painter's order (gx + gy ascending)
	sort.Slice(cells, func(i, j int) bool {
		si := cells[i].gx + cells[i].gy
		sj := cells[j].gx + cells[j].gy
		if si != sj {
			return si < sj
		}
		return cells[i].gx < cells[j].gx
	})

	// 4. Draw the grand foundational citadel plinth and canal basin under the entire N x N grid
	plinthPath := &canvas.Path{}
	plinthRim := &canvas.Path{}
	plinthCol := render.LerpOklab(ukiyoGrad.SampleOklab(0.25), grad.SampleOklab(0.22), 0.45)
	drawIsoBox(rc, iso, 0, 0, -0.55, fn, fn, 0, plinthCol, plinthPath, plinthRim)

	// Draw plaza pavement tiles and glowing turquoise canal waterways at z = 0
	canalCol := render.LerpOklab(color.RGBA{R: 38, G: 158, B: 172, A: 255}, grad.SampleOklab(0.32), 0.30)
	plazaCol := render.LerpOklab(grad.SampleOklab(0.58), color.RGBA{R: 228, G: 196, B: 168, A: 255}, 0.45)

	for _, cell := range cells {
		fgx := float64(cell.gx)
		fgy := float64(cell.gy)
		if cell.isCanal {
			drawIsoBox(rc, iso, fgx, fgy, -0.14, fgx+1.0, fgy+1.0, -0.02, canalCol, plinthPath, nil)
			// Shimmering water reflection line
			w0 := iso(fgx+0.25, fgy+0.5, -0.01)
			w1 := iso(fgx+0.75, fgy+0.5, -0.01)
			rc.SetBlendMode(render.BlendScreen)
			rc.StrokePolyline([]render.Point{w0, w1}, 1.2*scale, color.RGBA{R: 185, G: 245, B: 255, A: 255}, 0.55)
			rc.SetBlendMode(render.BlendOver)
		} else {
			tileTone := render.LerpOklab(plazaCol, grad.SampleOklab(cell.colorT), 0.25)
			p0 := iso(fgx, fgy, 0)
			p1 := iso(fgx+1, fgy, 0)
			p2 := iso(fgx+1, fgy+1, 0)
			p3 := iso(fgx, fgy+1, 0)
			rc.FillPolygon([]render.Point{p0, p1, p2, p3}, tileTone, 0.95)
			plinthPath.MoveTo(p0.X, p0.Y)
			plinthPath.LineTo(p1.X, p1.Y)
			plinthPath.LineTo(p2.X, p2.Y)
			plinthPath.LineTo(p3.X, p3.Y)
			plinthPath.Close()
		}
	}

	inkCol := color.RGBA{R: 20, G: 22, B: 34, A: 255}
	rimHighlightCol := color.RGBA{R: 255, G: 232, B: 195, A: 255}
	shadowTone := color.RGBA{R: 22, G: 20, B: 42, A: 255}
	windowDarkCol := color.RGBA{R: 18, G: 20, B: 34, A: 255}
	windowAmberCol := color.RGBA{R: 255, G: 202, B: 96, A: 255}

	rc.StrokePath(plinthPath, 0.75*scale, canvas.RoundCap, canvas.RoundJoin, inkCol, 0.45)

	// 5. Render each building column in strict back-to-front painter's order
	for _, cell := range cells {
		if cell.isCanal || cell.height <= 0.05 {
			continue
		}

		fgx := float64(cell.gx)
		fgy := float64(cell.gy)
		totalH := cell.height
		margin := 0.07
		bx0, by0 := fgx+margin, fgy+margin
		bx1, by1 := fgx+1.0-margin, fgy+1.0-margin

		// 5.1 Soft cast shadow polygon onto the ground/plaza (z = 0) in the direction of the low sun
		shLenX := math.Min(1.65, totalH*0.58)
		shLenY := math.Min(1.15, totalH*0.32)
		sp0 := iso(bx1, by0, 0)
		sp1 := iso(bx1+shLenX, by0+shLenY, 0)
		sp2 := iso(bx1+shLenX, by1+shLenY, 0)
		sp3 := iso(bx0+shLenX, by1+shLenY, 0)
		sp4 := iso(bx0, by1, 0)
		sp5 := iso(bx1, by1, 0)

		rc.SetBlendMode(render.BlendMultiply)
		rc.FillPolygon([]render.Point{sp0, sp1, sp2, sp3, sp4, sp5}, shadowTone, 0.36)
		rc.SetBlendMode(render.BlendOver)

		// 5.2 Draw base block and 2-4 stepped upper ziggurat tiers
		bldgEdgePath := &canvas.Path{}
		bldgRimPath := &canvas.Path{}

		basePaletteCol := render.LerpOklab(
			grad.SampleOklab(cell.colorT),
			ukiyoGrad.SampleOklab(common.Clamp01(0.20+0.65*cell.colorT)),
			0.22,
		)

		curZ := 0.0
		tierWeights := make([]float64, cell.tiers)
		weightSum := 0.0
		for t := 0; t < cell.tiers; t++ {
			wT := math.Pow(0.76, float64(t))
			tierWeights[t] = wT
			weightSum += wT
		}

		var topX0, topY0, topX1, topY1, topZ float64

		for t := 0; t < cell.tiers; t++ {
			tierH := totalH * (tierWeights[t] / weightSum)
			nextZ := curZ + tierH

			// Inset each upper tier for stepped Brutalist / Monument-Valley ziggurat silhouette
			inset := float64(t) * 0.095
			tx0, ty0 := bx0+inset, by0+inset
			tx1, ty1 := bx1-inset, by1-inset
			if tx1-tx0 < 0.22 {
				midX := 0.5 * (bx0 + bx1)
				midY := 0.5 * (by0 + by1)
				tx0, tx1 = midX-0.11, midX+0.11
				ty0, ty1 = midY-0.11, midY+0.11
			}

			tierCol := render.LerpOklab(
				basePaletteCol,
				grad.SampleOklab(common.Clamp01(cell.colorT+float64(t)*0.14)),
				0.42,
			)

			drawIsoBox(rc, iso, tx0, ty0, curZ, tx1, ty1, nextZ, tierCol, bldgEdgePath, bldgRimPath)

			// 5.3 Architectural details on vertical faces: recessed vertical colonnade slits & glowing windows
			if tierH >= 0.28 && tx1-tx0 >= 0.28 {
				slitCount := 3
				if tx1-tx0 < 0.52 {
					slitCount = 2
				}
				zBot := curZ + tierH*0.18
				zTop := curZ + tierH*0.80

				// Left face colonnade (x = tx1, varying y in [ty0, ty1])
				for s := 0; s < slitCount; s++ {
					u0 := (float64(s) + 0.28) / float64(slitCount)
					u1 := (float64(s) + 0.68) / float64(slitCount)
					wy0 := common.Lerp(ty0+0.04, ty1-0.04, u0)
					wy1 := common.Lerp(ty0+0.04, ty1-0.04, u1)

					wP0 := iso(tx1, wy0, zTop)
					wP1 := iso(tx1, wy1, zTop)
					wP2 := iso(tx1, wy1, zBot)
					wP3 := iso(tx1, wy0, zBot)

					lit := ((cell.gx*7 + cell.gy*13 + t*5 + s*3) % 5) == 0
					if lit {
						rc.SetBlendMode(render.BlendOver)
						rc.FillPolygon([]render.Point{wP0, wP1, wP2, wP3}, windowAmberCol, 0.92)
						midW := wP0.Lerp(wP2, 0.5)
						rc.SetBlendMode(render.BlendAdd)
						rc.DrawGlowDot(midW.X, midW.Y, 3.8*scale, windowAmberCol, 0.42, 0.25)
						rc.SetBlendMode(render.BlendOver)
					} else {
						rc.FillPolygon([]render.Point{wP0, wP1, wP2, wP3}, windowDarkCol, 0.68)
					}
				}

				// Right face colonnade (y = ty1, varying x in [tx0, tx1])
				for s := 0; s < slitCount; s++ {
					u0 := (float64(s) + 0.28) / float64(slitCount)
					u1 := (float64(s) + 0.68) / float64(slitCount)
					wx0 := common.Lerp(tx1-0.04, tx0+0.04, u0)
					wx1 := common.Lerp(tx1-0.04, tx0+0.04, u1)

					wP0 := iso(wx0, ty1, zTop)
					wP1 := iso(wx1, ty1, zTop)
					wP2 := iso(wx1, ty1, zBot)
					wP3 := iso(wx0, ty1, zBot)

					lit := ((cell.gx*11 + cell.gy*5 + t*7 + s*2) % 6) == 0
					if lit {
						rc.SetBlendMode(render.BlendOver)
						rc.FillPolygon([]render.Point{wP0, wP1, wP2, wP3}, windowAmberCol, 0.85)
						midW := wP0.Lerp(wP2, 0.5)
						rc.SetBlendMode(render.BlendAdd)
						rc.DrawGlowDot(midW.X, midW.Y, 3.5*scale, windowAmberCol, 0.36, 0.25)
						rc.SetBlendMode(render.BlendOver)
					} else {
						rc.FillPolygon([]render.Point{wP0, wP1, wP2, wP3}, windowDarkCol, 0.76)
					}
				}
			}

			topX0, topY0, topX1, topY1, topZ = tx0, ty0, tx1, ty1, nextZ
			curZ = nextZ
		}

		// 5.4 Elevated sky-bridges connecting to adjacent earlier-drawn towers (gx-1 or gy-1)
		if cell.gx > 0 {
			prev := grid[cell.gx-1][cell.gy]
			if !prev.isCanal && prev.height >= 1.15 && cell.height >= 1.15 && (cell.gx+cell.gy)%2 == 0 {
				zBridge := math.Min(prev.height, cell.height) * 0.56
				brY0 := fgy + 0.36
				brY1 := fgy + 0.64
				brCol := render.LerpOklab(basePaletteCol, color.RGBA{R: 235, G: 185, B: 145, A: 255}, 0.35)
				drawIsoBox(rc, iso, fgx-0.25, brY0, zBridge, fgx+0.15, brY1, zBridge+0.14, brCol, bldgEdgePath, bldgRimPath)
			}
		}
		if cell.gy > 0 {
			prev := grid[cell.gx][cell.gy-1]
			if !prev.isCanal && prev.height >= 1.15 && cell.height >= 1.15 && (cell.gx*3+cell.gy)%3 == 0 {
				zBridge := math.Min(prev.height, cell.height) * 0.62
				brX0 := fgx + 0.36
				brX1 := fgx + 0.64
				brCol := render.LerpOklab(basePaletteCol, color.RGBA{R: 235, G: 185, B: 145, A: 255}, 0.35)
				drawIsoBox(rc, iso, brX0, fgy-0.25, zBridge, brX1, fgy+0.15, zBridge+0.14, brCol, bldgEdgePath, bldgRimPath)
			}
		}

		// 5.5 Rooftop structures: Monument-Valley domes, obelisk spires, or miniature sanctuary trees
		midX := 0.5 * (topX0 + topX1)
		midY := 0.5 * (topY0 + topY1)
		spanW := topX1 - topX0

		switch cell.roofStyle {
		case 1:
			// Stepped octagonal/hemispherical Monument-Valley dome + lantern finial
			domeBase := iso(midX, midY, topZ+0.08)
			domeR := spanW * tileW * 0.28
			domeCol := render.LerpOklab(grad.SampleOklab(0.82), color.RGBA{R: 255, G: 215, B: 155, A: 255}, 0.55)
			domeShade := render.LerpOklab(domeCol, inkCol, 0.38)
			rc.SetBlendMode(render.BlendOver)
			rc.DrawGlowDot(domeBase.X, domeBase.Y-domeR*0.25, domeR, domeShade, 0.95, 0.85)
			rc.DrawGlowDot(domeBase.X-domeR*0.18, domeBase.Y-domeR*0.38, domeR*0.76, domeCol, 0.92, 0.72)

			apex := iso(midX, midY, topZ+0.52)
			bldgEdgePath.MoveTo(domeBase.X, domeBase.Y-domeR*0.85)
			bldgEdgePath.LineTo(apex.X, apex.Y)
			rc.SetBlendMode(render.BlendAdd)
			rc.DrawGlowDot(apex.X, apex.Y, 3.6*scale, windowAmberCol, 0.68, 0.35)
			rc.SetBlendMode(render.BlendOver)

		case 2:
			// Pyramidal obelisk spire with glowing celestial beacon
			obX0, obY0 := midX-spanW*0.24, midY-spanW*0.24
			obX1, obY1 := midX+spanW*0.24, midY+spanW*0.24
			spireH := 0.65
			apex := iso(midX, midY, topZ+spireH)
			b10 := iso(obX1, obY0, topZ)
			b11 := iso(obX1, obY1, topZ)
			b01 := iso(obX0, obY1, topZ)

			leftSpire := render.LerpOklab(basePaletteCol, color.RGBA{R: 255, G: 230, B: 190, A: 255}, 0.45)
			rightSpire := render.LerpOklab(basePaletteCol, inkCol, 0.35)
			rc.FillPolygon([]render.Point{apex, b10, b11}, leftSpire, 0.98)
			rc.FillPolygon([]render.Point{apex, b11, b01}, rightSpire, 0.98)

			bldgEdgePath.MoveTo(apex.X, apex.Y)
			bldgEdgePath.LineTo(b10.X, b10.Y)
			bldgEdgePath.LineTo(b11.X, b11.Y)
			bldgEdgePath.LineTo(b01.X, b01.Y)
			bldgEdgePath.Close()
			bldgEdgePath.MoveTo(apex.X, apex.Y)
			bldgEdgePath.LineTo(b11.X, b11.Y)

			rc.SetBlendMode(render.BlendAdd)
			rc.DrawGlowDot(apex.X, apex.Y, 4.2*scale, windowAmberCol, 0.72, 0.35)
			rc.SetBlendMode(render.BlendOver)

		case 3:
			// Miniature sanctuary cypress / courtyard tree on the terrace
			trunkBase := iso(midX, midY, topZ)
			trunkTop := iso(midX, midY, topZ+0.22)
			rc.StrokePolyline([]render.Point{trunkBase, trunkTop}, 1.4*scale, inkCol, 0.85)
			foliageCol := color.RGBA{R: 56, G: 138, B: 122, A: 255}
			foliageHighlight := color.RGBA{R: 135, G: 205, B: 168, A: 255}
			rc.DrawGlowDot(trunkTop.X, trunkTop.Y-2.0*scale, 4.6*scale, foliageCol, 0.92, 0.78)
			rc.DrawGlowDot(trunkTop.X-1.4*scale, trunkTop.Y-3.2*scale, 2.8*scale, foliageHighlight, 0.85, 0.65)
		}

		// 5.6 Stroke all isometric edges of this tower with crisp fine ink lines & warm sunlit rim
		rc.SetBlendMode(render.BlendOver)
		rc.StrokePath(bldgEdgePath, 1.05*scale, canvas.RoundCap, canvas.RoundJoin, inkCol, 0.84)
		rc.SetBlendMode(render.BlendScreen)
		rc.StrokePath(bldgRimPath, 0.75*scale, canvas.RoundCap, canvas.RoundJoin, rimHighlightCol, 0.42)
		rc.SetBlendMode(render.BlendOver)
	}

	// 6. Soft atmospheric mist wash near the bottom of the canvas
	mistCol := render.LerpOklab(skyHorizon, color.RGBA{R: 255, G: 235, B: 215, A: 255}, 0.45)
	mistStartY := int(fh * 0.64)
	for py := mistStartY; py < h; py++ {
		u := float64(py-mistStartY) / math.Max(1.0, float64(h-1-mistStartY))
		smoothU := common.Smoothstep(0, 1, u)
		fy := float64(py)
		for px := 0; px < w; px++ {
			fx := float64(px)
			nMist := 0.5 + 0.5*noise.Simplex2D(fx*0.006/scale, fy*0.015/scale)
			alpha := smoothU * (0.18 + 0.22*nMist)
			rc.BlendPixel(px, py, mistCol, alpha)
		}
	}

	// 7. Paper texture, vignette, and ACES filmic tone-mapped resolve
	rc.ApplyPaperTexture(0.28, 303)
	rc.ApplyVignette(0.24)
	c.MergeRenderCanvasToneMapped(rc, 1.06, 0)
}
