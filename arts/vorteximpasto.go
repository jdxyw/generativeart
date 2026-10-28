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

type vortexImpasto struct {
	vortexCount  int
	strokeCount  int
	curlStrength float64
}

// NewVortexImpasto returns a vortexImpasto engine that renders a ragged cream-bordered
// swirling micro-vortex fur/impasto field in golden-yellow, turquoise, cerulean, and dark forest charcoal.
func NewVortexImpasto(vortexCount, strokeCount int, curlStrength float64) *vortexImpasto {
	if vortexCount <= 0 {
		vortexCount = 46
	}
	if strokeCount <= 0 {
		strokeCount = 22000
	}
	if curlStrength <= 0 {
		curlStrength = 1.0
	}
	return &vortexImpasto{
		vortexCount:  vortexCount,
		strokeCount:  strokeCount,
		curlStrength: curlStrength,
	}
}

type microVortex struct {
	x, y       float64
	radius     float64
	spin       float64
	strength   float64
	hasEyeSpot bool
}

func buildBristleRibbonPath(pts []render.Point, halfWidth, dxOff, dyOff float64) *canvas.Path {
	n := len(pts)
	if n < 2 || halfWidth <= 0 {
		return nil
	}
	p := &canvas.Path{}
	var rightX, rightY [16]float64
	useStack := n <= 16
	var rightSliceX, rightSliceY []float64
	if !useStack {
		rightSliceX = make([]float64, n)
		rightSliceY = make([]float64, n)
	}

	invN := 1.0 / float64(n-1)
	for i := 0; i < n; i++ {
		var tx, ty float64
		if i == 0 {
			tx = pts[1].X - pts[0].X
			ty = pts[1].Y - pts[0].Y
		} else if i == n-1 {
			tx = pts[n-1].X - pts[n-2].X
			ty = pts[n-1].Y - pts[n-2].Y
		} else {
			tx = pts[i+1].X - pts[i-1].X
			ty = pts[i+1].Y - pts[i-1].Y
		}
		l := math.Hypot(tx, ty)
		var nx, ny float64
		if l > 1e-9 {
			nx = -ty / l
			ny = tx / l
		} else {
			nx, ny = 0, 1
		}

		t := float64(i) * invN
		// Taper root slightly and bristle tip smoothly to a fine point
		env := 1.0
		if t < 0.18 {
			env = 0.45 + 0.55*(t/0.18)
		} else if t > 0.72 {
			u := (t - 0.72) / 0.28
			env = math.Max(0.08, 1.0-u*u)
		}
		w := halfWidth * env
		cx := pts[i].X + dxOff
		cy := pts[i].Y + dyOff

		lx := cx + nx*w
		ly := cy + ny*w
		rx := cx - nx*w
		ry := cy - ny*w

		if i == 0 {
			p.MoveTo(lx, ly)
		} else {
			p.LineTo(lx, ly)
		}
		if useStack {
			rightX[i] = rx
			rightY[i] = ry
		} else {
			rightSliceX[i] = rx
			rightSliceY[i] = ry
		}
	}

	if useStack {
		for i := n - 1; i >= 0; i-- {
			p.LineTo(rightX[i], rightY[i])
		}
	} else {
		for i := n - 1; i >= 0; i-- {
			p.LineTo(rightSliceX[i], rightSliceY[i])
		}
	}
	p.Close()
	return p
}

// Generative renders the VortexImpasto composition.
func (vi *vortexImpasto) Generative(c *generativeart.Canva) {
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
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	rc := c.NewRenderCanvas()
	// Disable vector op recording so 22,000+ bristle paths rasterize directly into HDR buffer with zero memory overhead.
	rc.SetRecordVector(false)

	// Warm museum cream paper (#efe8de)
	creamPaper := color.RGBA{R: 239, G: 232, B: 222, A: 255}
	creamShadow := color.RGBA{R: 229, G: 220, B: 208, A: 255}
	rc.FillBackground(creamPaper)

	// Palette colors matching Reference Image 2
	tangerineOrange := color.RGBA{R: 232, G: 125, B: 40, A: 255} // #e87d28
	coralAmber := color.RGBA{R: 240, G: 149, B: 38, A: 255}      // #f09526
	cadmiumYellow := color.RGBA{R: 247, G: 202, B: 36, A: 255}   // #f7ca24
	sunflowerGold := color.RGBA{R: 244, G: 181, B: 34, A: 255}   // #f4b522
	lemonHighlight := color.RGBA{R: 250, G: 224, B: 86, A: 255}  // #fae056

	chartreuseGreen := color.RGBA{R: 110, G: 196, B: 104, A: 255} // #6ec468
	seafoamTurq := color.RGBA{R: 32, G: 181, B: 170, A: 255}      // #20b5aa
	brightCyan := color.RGBA{R: 28, G: 163, B: 176, A: 255}       // #1ca3b0

	forestCharcoal := color.RGBA{R: 30, G: 41, B: 29, A: 255}  // #1e291d
	oliveUmber := color.RGBA{R: 43, G: 56, B: 38, A: 255}      // #2b3826
	charcoalBlack := color.RGBA{R: 21, G: 29, B: 21, A: 255}   // #151d15
	deepTealForest := color.RGBA{R: 20, G: 72, B: 68, A: 255}  // transition around dark band

	aquaTurquoise := color.RGBA{R: 38, G: 191, B: 181, A: 255} // #26bfb5
	richTurquoise := color.RGBA{R: 24, G: 156, B: 166, A: 255} // #189ca6
	deepCerulean := color.RGBA{R: 18, G: 107, B: 140, A: 255}  // #126b8c
	prussianBlue := color.RGBA{R: 13, G: 74, B: 102, A: 255}   // #0d4a66

	useCustom := false
	var customGrad *render.MultiStopGradient
	if useCustom {
		customGrad = render.NewMultiStopGradient(c.Opts().ColorSchema()...)
	}

	// Macro color zone evaluation matching Reference Image 2
	evalMacroColor := func(nx, ny float64) color.RGBA {
		nx = common.Clamp01(nx)
		ny = common.Clamp01(ny)

		w1 := noise.FBM2D(nx*2.4+1.3, ny*2.4+0.7, 3, 2.0, 0.5) * 0.11
		w2 := noise.FBM2D(nx*4.5-0.9, ny*4.5+2.1, 2, 2.0, 0.5) * 0.05
		fineN := 0.5 + 0.5*noise.Simplex2D(nx*6.5+3.1, ny*6.5-1.7)

		if useCustom && customGrad != nil {
			t := common.Clamp01(0.55*nx + 0.45*(1.0-ny) + w1 + w2)
			return customGrad.SampleOklab(t)
		}

		// Base blue-turquoise canvas field (Zone 4: Turquoise, Cerulean & Deep Prussian Blue)
		baseCol := render.LerpOklab(seafoamTurq, richTurquoise, common.Smoothstep(0.18, 0.62, ny+w2))

		// Upper-right corner (nx > 0.62, ny < 0.22): bright aqua-turquoise #26bfb5
		urWeight := common.Smoothstep(0.54, 0.76, nx+w2) * common.Smoothstep(0.32, 0.10, ny+w2)
		if urWeight > 0 {
			urCol := render.LerpOklab(aquaTurquoise, seafoamTurq, fineN*0.35)
			baseCol = render.LerpOklab(baseCol, urCol, urWeight)
		}

		// Lower-center & mid-right (ny > 0.52, nx > 0.42): rich turquoise -> deep cerulean -> Prussian blue
		blueDrive := common.Smoothstep(0.36, 0.85, 0.58*nx+0.48*ny+w1*0.8+w2)
		if blueDrive > 0 {
			ceruleanMix := render.LerpOklab(richTurquoise, deepCerulean, common.Smoothstep(0.0, 0.62, blueDrive))
			prussianMix := render.LerpOklab(ceruleanMix, prussianBlue, common.Smoothstep(0.45, 0.96, blueDrive)*common.Smoothstep(0.42, 0.88, nx+0.35*ny))
			baseCol = render.LerpOklab(baseCol, prussianMix, common.Smoothstep(0.38, 0.68, nx+0.4*ny+w1))
		}

		// Also lower-right orange/gold corner accent peek if near bottom-center-left
		diagLL := (ny - 0.36) - 0.95*nx + w1 + w2

		// Zone 2: Green-Turquoise Transition Band (-0.09 < diagLL <= 0.04)
		if diagLL > -0.12 && diagLL < 0.08 {
			tBand := common.Smoothstep(-0.11, 0.05, diagLL)
			transCol := render.LerpOklab(brightCyan, seafoamTurq, common.Smoothstep(0.0, 0.45, tBand))
			transCol = render.LerpOklab(transCol, chartreuseGreen, common.Smoothstep(0.35, 0.88, tBand))
			bandMask := common.Smoothstep(-0.12, -0.02, diagLL) * common.Smoothstep(0.08, -0.01, diagLL)
			baseCol = render.LerpOklab(baseCol, transCol, bandMask*0.92)
		}

		// Zone 1: Lower-Left Golden-Yellow & Warm Orange Zone (diagLL > 0)
		if diagLL > -0.02 {
			yellowStrength := common.Smoothstep(-0.02, 0.07, diagLL)
			goldCol := render.LerpOklab(cadmiumYellow, sunflowerGold, fineN*0.65)
			goldCol = render.LerpOklab(goldCol, lemonHighlight, common.Smoothstep(0.62, 0.95, fineN)*0.55)

			// Near the lower-left and bottom edge (nx < 0.25 or ny > 0.85): warm tangerine-orange & coral-amber
			edgeWarmth := math.Max(
				common.Smoothstep(0.28, 0.06, nx+w2*0.8)*common.Smoothstep(0.42, 0.72, ny),
				common.Smoothstep(0.80, 0.93, ny+w2*0.6)*common.Smoothstep(0.55, 0.18, nx),
			)
			// Secondary warm orange eddy pockets inside the golden zone
			orangePocket := common.Smoothstep(0.18, 0.68, noise.FBM2D(nx*5.2+0.4, ny*5.2+1.9, 2, 2.0, 0.5)) *
				common.Smoothstep(0.38, 0.08, nx) * 0.65
			warmT := common.Clamp01(edgeWarmth + orangePocket)
			warmCol := render.LerpOklab(coralAmber, tangerineOrange, common.Smoothstep(0.25, 0.85, warmT))
			zone1Col := render.LerpOklab(goldCol, warmCol, warmT*0.88)

			baseCol = render.LerpOklab(baseCol, zone1Col, yellowStrength)
		}

		// Zone 3: Upper-Left to Center-Right Dark Forest-Olive Charcoal Band
		// Centered along the line from (0.10, 0.16) to (0.78, 0.36)
		bandY := 0.14 + 0.28*nx + w1*0.9
		halfBandW := 0.14 + 0.05*math.Sin(nx*math.Pi)
		dDark := math.Abs(ny-bandY) / halfBandW

		// Include upper-left dark lobe extension around (nx < 0.24, ny < 0.34)
		ulLobe := math.Hypot((nx-0.14)/0.13, (ny-0.20)/0.14) + w2*1.4
		dEffective := math.Min(dDark, ulLobe)

		xGate := common.Smoothstep(0.86, 0.66, nx+w2*0.6) * common.Smoothstep(0.01, 0.08, nx+w2*0.4)
		if dEffective < 1.35 && xGate > 0 {
			// Outer olive/forest halo transitioning to deep charcoal core
			outerRim := common.Smoothstep(1.35, 0.72, dEffective) * xGate
			coreStrength := common.Smoothstep(0.95, 0.22, dEffective) * xGate

			rimCol := render.LerpOklab(deepTealForest, oliveUmber, common.Smoothstep(0.0, 0.8, outerRim))
			darkCol := render.LerpOklab(oliveUmber, forestCharcoal, common.Smoothstep(0.15, 0.65, coreStrength))
			darkCol = render.LerpOklab(darkCol, charcoalBlack, common.Smoothstep(0.55, 0.95, coreStrength)*fineN*0.75)

			baseCol = render.LerpOklab(baseCol, rimCol, outerRim*0.75)
			baseCol = render.LerpOklab(baseCol, darkCol, coreStrength*0.96)
		}

		return baseCol
	}

	// Inner painted canvas rectangle bounds with 1D FBM edge perturbation (±0.015*w)
	leftEdge := func(ny float64) float64 {
		return (0.068 + 0.015*noise.FBM2D(ny*5.5+0.7, 1.1, 3, 2.0, 0.5)) * fw
	}
	rightEdge := func(ny float64) float64 {
		return (0.932 + 0.015*noise.FBM2D(ny*5.5+4.3, 2.7, 3, 2.0, 0.5)) * fw
	}
	topEdge := func(nx float64) float64 {
		return (0.060 + 0.015*noise.FBM2D(nx*5.5+2.1, 3.9, 3, 2.0, 0.5)) * fh
	}
	bottomEdge := func(nx float64) float64 {
		return (0.940 + 0.015*noise.FBM2D(nx*5.5+6.4, 5.2, 3, 2.0, 0.5)) * fh
	}

	// 1. Base Underpainting + Subtle Paper Grain
	for py := 0; py < h; py++ {
		fy := float64(py)
		ny := fy / fh
		xMin := leftEdge(ny)
		xMax := rightEdge(ny)
		for px := 0; px < w; px++ {
			fx := float64(px)
			nx := fx / fw
			yMin := topEdge(nx)
			yMax := bottomEdge(nx)

			// Paper grain outside and inside
			grain := 0.5 + 0.5*noise.Simplex2D(fx*0.35, fy*0.35)
			paperCol := render.LerpOklab(creamPaper, creamShadow, grain*0.18)

			if fx >= xMin && fx <= xMax && fy >= yMin && fy <= yMax {
				// Distance inside ragged border for soft underpainting feathering over 3px
				dEdge := math.Min(math.Min(fx-xMin, xMax-fx), math.Min(fy-yMin, yMax-fy))
				edgeCov := common.Smoothstep(0.0, 3.5*scale, dEdge)

				macro := evalMacroColor(nx, ny)
				lab := render.RGBAToOklab(macro)
				// Darken by 0.68 in Oklab so no unintended white gaps show between bristles
				underCol := render.OklabToRGBA(render.Oklab{
					L: common.Constrain(lab.L*0.68, 0.04, 0.92),
					A: lab.A * 0.92,
					B: lab.B * 0.92,
				}, 255)
				finalCol := render.LerpOklab(paperCol, underCol, edgeCov)
				rL, gL, bL, _ := render.ColorToLinearRGB(finalCol)
				rc.SetPixelLinear(px, py, rL, gL, bL, 1.0)
			} else {
				rL, gL, bL, _ := render.ColorToLinearRGB(paperCol)
				rc.SetPixelLinear(px, py, rL, gL, bL, 1.0)
			}
		}
	}

	// 2. Swirling Micro-Vortex Velocity Field
	// Generate vortexCount well-spaced vortex centers across the inner rectangle
	vortices := make([]microVortex, 0, vi.vortexCount)
	cols := int(math.Max(3, math.Round(math.Sqrt(float64(vi.vortexCount)*fw/fh))))
	rows := (vi.vortexCount + cols - 1) / cols
	innerX0, innerX1 := 0.085*fw, 0.915*fw
	innerY0, innerY1 := 0.080*fh, 0.920*fh
	cellW := (innerX1 - innerX0) / float64(cols)
	cellH := (innerY1 - innerY0) / float64(rows)

	for r := 0; r < rows && len(vortices) < vi.vortexCount; r++ {
		for cl := 0; cl < cols && len(vortices) < vi.vortexCount; cl++ {
			jx := (0.18 + 0.64*rng.Float64()) * cellW
			jy := (0.18 + 0.64*rng.Float64()) * cellH
			vx := innerX0 + float64(cl)*cellW + jx
			vy := innerY0 + float64(r)*cellH + jy
			rad := (32.0 + 28.0*rng.Float64()) * scale
			spin := 1.0
			if rng.Float64() < 0.5 {
				spin = -1.0
			}
			vortices = append(vortices, microVortex{
				x:          vx,
				y:          vy,
				radius:     rad,
				spin:       spin,
				strength:   0.85 + 0.35*rng.Float64(),
				hasEyeSpot: rng.Float64() < 0.76,
			})
		}
	}

	curlScale := vi.curlStrength
	evalVelocityAngle := func(x, y float64) float64 {
		nx := x / fw
		ny := y / fh
		// Smooth curl-noise base angle
		curlVec := noise.Curl2D(nx*2.6+0.5, ny*2.6-0.3, 0.01)
		curlAng := math.Atan2(curlVec.Y, curlVec.X) * curlScale

		// Accumulate weighted vortex tangent + inward spiral pull
		vxSum := math.Cos(curlAng) * 0.28
		vySum := math.Sin(curlAng) * 0.28

		for i := range vortices {
			v := &vortices[i]
			dx := x - v.x
			dy := y - v.y
			reach := 1.35 * v.radius
			if math.Abs(dx) >= reach || math.Abs(dy) >= reach {
				continue
			}
			d := math.Hypot(dx, dy)
			if d < reach {
				u := 1.0 - d/reach
				wv := u * u * v.strength * 2.2
				ang := math.Atan2(dy, dx) + v.spin*(math.Pi*0.44)
				vxSum += math.Cos(ang) * wv
				vySum += math.Sin(ang) * wv
			}
		}
		return math.Atan2(vySum, vxSum)
	}

	// 3. 22,000+ Fine Swirling Fur / Bristle Strokes & Ragged Edge Bleed
	const traceSteps = 12
	stepLen := 2.2 * scale
	var pts [traceSteps + 1]render.Point

	for s := 0; s < vi.strokeCount; s++ {
		// Spawn uniformly across the inner rectangle (right up to the ragged margin so bristles splay out 8..18px over the cream border)
		u := rng.Float64()
		v := rng.Float64()
		// Bias a small fraction of strokes toward the perimeter to guarantee rich furry deckle edges
		if s%9 == 0 {
			side := s % 4
			switch side {
			case 0:
				u = rng.Float64() * 0.055
			case 1:
				u = 1.0 - rng.Float64()*0.055
			case 2:
				v = rng.Float64() * 0.055
			case 3:
				v = 1.0 - rng.Float64()*0.055
			}
		}

		xMin := leftEdge(v) - 3.0*scale
		xMax := rightEdge(v) + 3.0*scale
		x0 := xMin + u*(xMax-xMin)

		nx0 := common.Clamp01(x0 / fw)
		yMin := topEdge(nx0) - 3.0*scale
		yMax := bottomEdge(nx0) + 3.0*scale
		y0 := yMin + v*(yMax-yMin)
		ny0 := common.Clamp01(y0 / fh)

		// Trace 12 steps of 2.2*scale px along the vortex velocity field
		pts[0] = render.Pt(x0, y0)
		cx, cy := x0, y0
		strokeBend := (rng.Float64() - 0.5) * 0.055
		for k := 1; k <= traceSteps; k++ {
			ang := evalVelocityAngle(cx, cy) + strokeBend*float64(k)
			cx += math.Cos(ang) * stepLen
			cy += math.Sin(ang) * stepLen
			pts[k] = render.Pt(cx, cy)
		}

		// Sample macro color at bristle root with Oklab lightness & subtle chroma/hue variation
		rootMacro := evalMacroColor(nx0, ny0)
		lab := render.RGBAToOklab(rootMacro)

		dL := -0.10 + 0.21*rng.Float64()
		dA := (rng.Float64() - 0.5) * 0.032
		dB := (rng.Float64() - 0.5) * 0.032

		litCol := render.OklabToRGBA(render.Oklab{
			L: common.Constrain(lab.L+dL, 0.05, 0.97),
			A: lab.A + dA,
			B: lab.B + dB,
		}, 255)

		flankCol := render.OklabToRGBA(render.Oklab{
			L: common.Constrain(lab.L*0.62, 0.02, 0.85),
			A: lab.A * 0.90,
			B: lab.B * 0.90,
		}, 255)

		flankHalfW := 0.80 * scale
		coreHalfW := (0.425 + 0.375*rng.Float64()) * scale

		if flankPath := buildBristleRibbonPath(pts[:], flankHalfW, 0.35*scale, 0.45*scale); flankPath != nil {
			rc.RasterizePath(flankPath, flankCol, 0.75)
		}
		if corePath := buildBristleRibbonPath(pts[:], coreHalfW, 0, 0); corePath != nil {
			rc.RasterizePath(corePath, litCol, 0.88)
		}
	}

	// 4. Vortex Eye Specular Pinpoints
	// Inside ~75% of the swirling vortex centers: tiny bright pale-cyan / off-white specular pinpoint highlights
	// (#d8f0ec, radius 1.3..2.3 * scale) cupped by a tiny dark crescent.
	pinpointCyan := color.RGBA{R: 216, G: 240, B: 236, A: 255} // #d8f0ec
	pinpointCore := color.RGBA{R: 248, G: 254, B: 252, A: 255}
	darkCrescent := color.RGBA{R: 12, G: 20, B: 18, A: 255}

	for i := range vortices {
		v := &vortices[i]
		if !v.hasEyeSpot {
			continue
		}
		nSpots := 1
		if rng.Float64() < 0.38 {
			nSpots = 2
		}
		for k := 0; k < nSpots; k++ {
			offAng := rng.Float64() * 2.0 * math.Pi
			offDist := (1.5 + 4.5*rng.Float64()) * scale
			px := v.x + math.Cos(offAng)*offDist
			py := v.y + math.Sin(offAng)*offDist
			rad := (1.3 + 1.0*rng.Float64()) * scale

			// Cupped dark crescent offset slightly along the vortex spin direction
			cAng := offAng + v.spin*math.Pi*0.5
			rc.DrawGlowDot(px+math.Cos(cAng)*0.85*scale, py+math.Sin(cAng)*0.85*scale, rad*1.35, darkCrescent, 0.78, 0.82)
			// Crisp pale-cyan specular pinpoint highlight
			rc.DrawGlowDot(px, py, rad, pinpointCyan, 0.94, 0.86)
			rc.DrawGlowDot(px-0.25*scale, py-0.25*scale, rad*0.55, pinpointCore, 0.96, 0.92)
		}
	}

	c.MergeRenderCanvas(rc)
}
