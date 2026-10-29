package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type nebulaPlumes struct {
	spiralArms     int
	strokeCount    int
	swirlTightness float64
}

// NewNebulaPlumes returns a nebulaPlumes engine that renders a ragged parchment-bordered
// colliding binary spiral galaxy and 14 satellite eddies in cosmic vermilion-gold, jade-emerald, and lapis-violet.
func NewNebulaPlumes(spiralArms, strokeCount int, swirlTightness float64) *nebulaPlumes {
	if spiralArms <= 0 {
		spiralArms = 3
	}
	if strokeCount <= 0 {
		strokeCount = 23000
	}
	if swirlTightness <= 0 {
		swirlTightness = 1.0
	}
	return &nebulaPlumes{
		spiralArms:     spiralArms,
		strokeCount:    strokeCount,
		swirlTightness: swirlTightness,
	}
}

type nebulaEddy struct {
	x, y     float64
	radius   float64
	spin     float64
	strength float64
	isMacro  bool
	isWarmA  bool
}

// Generative renders the static NebulaPlumes composition.
func (np *nebulaPlumes) Generative(c *generativeart.Canva) {
	np.GenerativeFrame(c, 0, 1, 0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360-degree looping animation over phase in [0, 1).
func (np *nebulaPlumes) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
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
		seed = 20261029
	}
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	isAnim := totalFrames > 1
	phi := 2.0 * math.Pi * phase
	loopX := 0.07 * math.Sin(phi)
	loopY := 0.07 * (1.0 - math.Cos(phi))

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	// Warm antique museum parchment (#f3ece1)
	parchment := color.RGBA{R: 243, G: 236, B: 225, A: 255}
	parchmentShadow := color.RGBA{R: 231, G: 222, B: 209, A: 255}
	rc.FillBackground(parchment)

	// Cosmic Vermilion-Gold vs Jade-Emerald vs Lapis-Violet palette
	// Galaxy A (upper-left): warm solar nebula
	whiteHotPeach := color.RGBA{R: 255, G: 238, B: 210, A: 255} // #ffeed2
	moltenAmber := color.RGBA{R: 255, G: 170, B: 43, A: 255}    // #ffaa2b
	fieryVermilion := color.RGBA{R: 230, G: 72, B: 46, A: 255}  // #e6482e
	deepCarmine := color.RGBA{R: 158, G: 28, B: 52, A: 255}     // #9e1c34

	// Galaxy B (lower-right): bioluminescent jade-emerald nebula
	iceMint := color.RGBA{R: 184, G: 248, B: 232, A: 255}       // #b8f8e8
	emeraldTeal := color.RGBA{R: 34, G: 200, B: 176, A: 255}    // #22c8b0
	peacockCyan := color.RGBA{R: 14, G: 136, B: 148, A: 255}    // #0e8894
	abyssalTeal := color.RGBA{R: 8, G: 72, B: 92, A: 255}       // #08485c

	// S-Shaped Tidal Bridge & Outer Halo
	orchidViolet := color.RGBA{R: 162, G: 82, B: 224, A: 255}   // #a252e0
	lapisUltra := color.RGBA{R: 34, G: 52, B: 136, A: 255}      // #223488
	aubergineVoid := color.RGBA{R: 17, G: 12, B: 40, A: 255}    // #110c28

	// 2 Colossal Counter-Rotating Super-Vortices + 14 Medium Satellite Eddies
	eddies := make([]nebulaEddy, 0, 16)
	baseGalAX, baseGalAY := 0.33*fw, 0.36*fh
	baseGalBX, baseGalBY := 0.68*fw, 0.65*fh
	galAX, galAY := baseGalAX, baseGalAY
	galBX, galBY := baseGalBX, baseGalBY
	if isAnim {
		galAX += 4.5 * scale * math.Sin(phi)
		galAY += 4.5 * scale * (1.0 - math.Cos(phi))
		galBX -= 4.5 * scale * math.Sin(phi)
		galBY -= 4.5 * scale * (1.0 - math.Cos(phi))
	}
	galARad := 215.0 * scale
	galBRad := 205.0 * scale

	eddies = append(eddies,
		nebulaEddy{
			x:        galAX,
			y:        galAY,
			radius:   galARad,
			spin:     1.0,
			strength: 1.55,
			isMacro:  true,
			isWarmA:  true,
		},
		nebulaEddy{
			x:        galBX,
			y:        galBY,
			radius:   galBRad,
			spin:     -1.0,
			strength: 1.50,
			isMacro:  true,
			isWarmA:  false,
		},
	)

	// 14 Medium Satellite Eddies distributed along the S-shaped tidal bridge and outer spiral arms
	satAnchors := [14][3]float64{
		{0.44, 0.45, 1.0},  // tidal bridge upper
		{0.52, 0.51, -1.0}, // tidal bridge center
		{0.58, 0.57, 1.0},  // tidal bridge lower
		{0.20, 0.24, 1.0},  // Galaxy A outer arm NW
		{0.46, 0.22, -1.0}, // Galaxy A outer arm NE
		{0.18, 0.48, -1.0}, // Galaxy A outer arm W
		{0.28, 0.62, 1.0},  // Galaxy A trailing plume SW
		{0.82, 0.50, -1.0}, // Galaxy B outer arm E
		{0.74, 0.34, 1.0},  // Galaxy B trailing plume NE
		{0.82, 0.77, -1.0}, // Galaxy B outer arm SE
		{0.54, 0.78, 1.0},  // Galaxy B outer arm S
		{0.37, 0.76, -1.0}, // lower tidal tail
		{0.64, 0.20, 1.0},  // upper tidal tail
		{0.16, 0.74, 1.0},  // far SW dust eddy
	}
	for i := 0; i < 14; i++ {
		jx := (rng.Float64() - 0.5) * 0.035 * fw
		jy := (rng.Float64() - 0.5) * 0.035 * fh
		rad := (38.0 + 26.0*rng.Float64()) * scale
		sx0 := satAnchors[i][0]*fw + jx
		sy0 := satAnchors[i][1]*fh + jy
		dA := math.Hypot(sx0-baseGalAX, sy0-baseGalAY)
		dB := math.Hypot(sx0-baseGalBX, sy0-baseGalBY)
		sx, sy := sx0, sy0
		if isAnim {
			ePhase := float64(i) * 1.2
			sx += 7.5 * scale * math.Sin(phi+ePhase)
			sy += 7.5 * scale * (1.0 - math.Cos(phi+ePhase))
		}
		eddies = append(eddies, nebulaEddy{
			x:        sx,
			y:        sy,
			radius:   rad,
			spin:     satAnchors[i][2],
			strength: 0.90 + 0.35*rng.Float64(),
			isMacro:  false,
			isWarmA:  dA <= dB,
		})
	}

	arms := float64(np.spiralArms)

	evalMacroColor := func(nx, ny float64) color.RGBA {
		nx = common.Clamp01(nx)
		ny = common.Clamp01(ny)

		w1 := noise.FBM2D(nx*2.5+1.1+loopX, ny*2.5+0.6+loopY, 3, 2.0, 0.5) * 0.10
		w2 := noise.FBM2D(nx*4.8-0.7-loopX*0.7, ny*4.8+1.9+loopY*0.7, 2, 2.0, 0.5) * 0.05
		fineN := 0.5 + 0.5*noise.Simplex2D(nx*6.2+2.7+loopX, ny*6.2-1.4+loopY)

		// Deep cosmic void and royal lapis background halo
		haloT := common.Smoothstep(0.10, 0.78, noise.FBM2D(nx*2.2+0.4, ny*2.2+1.7, 3, 2.0, 0.5)*0.5+0.5)
		baseCol := render.LerpOklab(aubergineVoid, lapisUltra, haloT*0.78)

		// S-shaped tidal bridge connecting (0.33, 0.36) and (0.68, 0.65)
		// Parameterize along line from A to B
		abX, abY := 0.68-0.33, 0.65-0.36
		abLenSq := abX*abX + abY*abY
		tBridge := ((nx-0.33)*abX + (ny-0.36)*abY) / abLenSq
		tClamped := common.Clamp01(tBridge)
		sCurve := 0.065 * math.Sin(tClamped*math.Pi*2.0)
		perpX, perpY := -abY/math.Sqrt(abLenSq), abX/math.Sqrt(abLenSq)
		bx := 0.33 + tClamped*abX + perpX*sCurve + w1*0.6
		by := 0.36 + tClamped*abY + perpY*sCurve + w2*0.6
		dBridge := math.Hypot(nx-bx, ny-by)
		bridgeMask := common.Smoothstep(0.24, 0.02, dBridge) * common.Smoothstep(-0.15, 0.15, tBridge) * common.Smoothstep(1.15, 0.85, tBridge)
		if bridgeMask > 0 {
			bridgeCol := render.LerpOklab(lapisUltra, orchidViolet, common.Smoothstep(0.15, 0.85, bridgeMask+fineN*0.25))
			baseCol = render.LerpOklab(baseCol, bridgeCol, bridgeMask*0.88)
		}

		// Galaxy A (upper-left): Cosmic Vermilion-Gold with logarithmic spiral arm structure
		dxA := nx - 0.33 + w1*0.45
		dyA := ny - 0.36 + w2*0.45
		rA := math.Hypot(dxA, dyA)
		if rA < 0.52 {
			thetaA := math.Atan2(dyA, dxA)
			armA := 0.5 + 0.5*math.Cos(arms*thetaA-rA*13.5*np.swirlTightness-phi)
			uA := common.Smoothstep(0.50, 0.0, rA)
			warmCol := render.LerpOklab(deepCarmine, fieryVermilion, common.Smoothstep(0.10, 0.52, uA+armA*0.25))
			warmCol = render.LerpOklab(warmCol, moltenAmber, common.Smoothstep(0.42, 0.80, uA+armA*0.22))
			warmCol = render.LerpOklab(warmCol, whiteHotPeach, common.Smoothstep(0.74, 0.96, uA+fineN*0.14))
			wA := common.Clamp01(uA * (0.68 + 0.32*armA))
			baseCol = render.LerpOklab(baseCol, warmCol, wA)
		}

		// Galaxy B (lower-right): Jade-Emerald & Peacock Cyan with counter-rotating spiral arms
		dxB := nx - 0.68 - w1*0.45
		dyB := ny - 0.65 - w2*0.45
		rB := math.Hypot(dxB, dyB)
		if rB < 0.50 {
			thetaB := math.Atan2(dyB, dxB)
			armB := 0.5 + 0.5*math.Cos(arms*thetaB+rB*13.5*np.swirlTightness+phi)
			uB := common.Smoothstep(0.48, 0.0, rB)
			coolCol := render.LerpOklab(abyssalTeal, peacockCyan, common.Smoothstep(0.10, 0.50, uB+armB*0.25))
			coolCol = render.LerpOklab(coolCol, emeraldTeal, common.Smoothstep(0.40, 0.78, uB+armB*0.22))
			coolCol = render.LerpOklab(coolCol, iceMint, common.Smoothstep(0.74, 0.96, uB+fineN*0.14))
			wB := common.Clamp01(uB * (0.68 + 0.32*armB))
			baseCol = render.LerpOklab(baseCol, coolCol, wB)
		}

		return baseCol
	}

	// Ragged deckle-edged inner canvas [0.068*w, 0.932*w] x [0.060*h, 0.940*h] perturbed by 1D FBM (±0.016*w)
	leftEdge := func(ny float64) float64 {
		return (0.068 + 0.016*noise.FBM2D(ny*5.5+0.9, 1.3, 3, 2.0, 0.5)) * fw
	}
	rightEdge := func(ny float64) float64 {
		return (0.932 + 0.016*noise.FBM2D(ny*5.5+4.1, 2.9, 3, 2.0, 0.5)) * fw
	}
	topEdge := func(nx float64) float64 {
		return (0.060 + 0.016*noise.FBM2D(nx*5.5+2.3, 3.7, 3, 2.0, 0.5)) * fh
	}
	bottomEdge := func(nx float64) float64 {
		return (0.940 + 0.016*noise.FBM2D(nx*5.5+6.1, 5.4, 3, 2.0, 0.5)) * fh
	}

	// 1. Base Underpainting + Subtle Museum Parchment Grain
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

			grain := 0.5 + 0.5*noise.Simplex2D(fx*0.34, fy*0.34)
			paperCol := render.LerpOklab(parchment, parchmentShadow, grain*0.18)

			if fx >= xMin && fx <= xMax && fy >= yMin && fy <= yMax {
				dEdge := math.Min(math.Min(fx-xMin, xMax-fx), math.Min(fy-yMin, yMax-fy))
				edgeCov := common.Smoothstep(0.0, 3.5*scale, dEdge)

				macro := evalMacroColor(nx, ny)
				lab := render.RGBAToOklab(macro)
				underCol := render.OklabToRGBA(render.Oklab{
					L: common.Constrain(lab.L*0.66, 0.03, 0.90),
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

	// 2. Binary Galaxy + 14 Satellite Eddy Velocity Field
	pitchBase := math.Pi * (0.43 + 0.03*common.Constrain(np.swirlTightness-1.0, -0.5, 1.0))
	evalVelocityAngle := func(x, y float64) float64 {
		nx := x / fw
		ny := y / fh
		curlVec := noise.Curl2D(nx*2.4+0.7, ny*2.4-0.5, 0.01)
		curlAng := math.Atan2(curlVec.Y, curlVec.X)
		if isAnim {
			curlAng += 0.20 * math.Sin(nx*math.Pi*2.5+ny*math.Pi*2.5-phi)
		}

		vxSum := math.Cos(curlAng) * 0.25
		vySum := math.Sin(curlAng) * 0.25

		for i := range eddies {
			v := &eddies[i]
			dx := x - v.x
			dy := y - v.y
			reach := 1.42 * v.radius
			if v.isMacro {
				reach = 1.65 * v.radius
			}
			if math.Abs(dx) >= reach || math.Abs(dy) >= reach {
				continue
			}
			d := math.Hypot(dx, dy)
			if d < reach {
				u := 1.0 - d/reach
				wv := u * u * v.strength * 2.35
				theta := math.Atan2(dy, dx)
				ang := theta + v.spin*pitchBase
				if isAnim {
					if v.isMacro {
						spiralWave := math.Sin(float64(arms)*theta - v.spin*phi - (d/v.radius)*4.2)
						ang = theta + v.spin*(pitchBase+0.26*spiralWave)
					} else {
						spiralWave := math.Sin(2.0*theta - v.spin*phi - (d/v.radius)*3.5 + float64(i)*0.7)
						ang = theta + v.spin*(pitchBase+0.22*spiralWave)
					}
				}
				vxSum += math.Cos(ang) * wv
				vySum += math.Sin(ang) * wv
			}
		}
		return math.Atan2(vySum, vxSum)
	}

	// Coherent spiral luminescence & chroma wave across the binary galaxies, 14 satellite eddies, and tidal bridge
	evalNebulaSpiralWave := func(x, y float64) (dLum, dChroma float64) {
		for i := range eddies {
			e := &eddies[i]
			dx := x - e.x
			dy := y - e.y
			reach := 1.48 * e.radius
			if e.isMacro {
				reach = 1.68 * e.radius
			}
			if math.Abs(dx) >= reach || math.Abs(dy) >= reach {
				continue
			}
			d := math.Hypot(dx, dy)
			if d < reach {
				u := 1.0 - d/reach
				env := u * u * (3.0 - 2.0*u)
				theta := math.Atan2(dy, dx)
				if e.isMacro {
					spiralPhase := arms*theta - e.spin*phi - (d/e.radius)*4.8
					w := math.Sin(spiralPhase)
					dLum += env * w * 0.125
					dChroma += env * math.Cos(spiralPhase) * 0.042
				} else {
					spiralPhase := 2.0*theta - e.spin*phi - (d/e.radius)*4.2 + float64(i)*0.67
					w := math.Sin(spiralPhase)
					dLum += env * w * 0.095
					dChroma += env * math.Cos(spiralPhase) * 0.032
				}
			}
		}
		// Travelling tidal-stream wave between Galaxy A and B
		nx, ny := x/fw, y/fh
		tidalProj := (nx-0.33)*0.35 + (ny-0.36)*0.29
		tidalWave := math.Sin(tidalProj*math.Pi*14.0 - phi)
		dLum += 0.04 * tidalWave
		return common.Constrain(dLum, -0.18, 0.18), common.Constrain(dChroma, -0.055, 0.055)
	}

	// 3. 23,000+ Sweeping Feathery Bristle Plumes & Deckle Border Splay (10..20 px)
	const traceSteps = 15
	stepLen := 2.45 * scale
	var pts [traceSteps + 1]render.Point

	for s := 0; s < np.strokeCount; s++ {
		u := rng.Float64()
		v := rng.Float64()
		// Bias perimeter strokes so feathery plumes splay 10..20 px over the antique parchment deckle border
		if s%8 == 0 {
			side := s % 4
			switch side {
			case 0:
				u = rng.Float64() * 0.058
			case 1:
				u = 1.0 - rng.Float64()*0.058
			case 2:
				v = rng.Float64() * 0.058
			case 3:
				v = 1.0 - rng.Float64()*0.058
			}
		}

		xMin := leftEdge(v) - 3.5*scale
		xMax := rightEdge(v) + 3.5*scale
		x0 := xMin + u*(xMax-xMin)

		nx0 := common.Clamp01(x0 / fw)
		yMin := topEdge(nx0) - 3.5*scale
		yMax := bottomEdge(nx0) + 3.5*scale
		y0 := yMin + v*(yMax-yMin)
		ny0 := common.Clamp01(y0 / fh)

		cx, cy := x0, y0
		var spatialPsi float64
		if isAnim {
			thetaA := math.Atan2(y0-galAY, x0-galAX)
			rA := math.Hypot(x0-galAX, y0-galAY) / (215.0 * scale)
			spatialPsi = float64(arms)*thetaA - rA*3.8
			ang0 := evalVelocityAngle(x0, y0)
			adv := 9.0 * scale * math.Sin(phi-spatialPsi)
			cx = x0 + math.Cos(ang0)*adv
			cy = y0 + math.Sin(ang0)*adv
		}
		pts[0] = render.Pt(cx, cy)
		strokeBend := (rng.Float64() - 0.5) * 0.048
		for k := 1; k <= traceSteps; k++ {
			var whip float64
			if isAnim {
				tStep := float64(k) / float64(traceSteps)
				whip = 0.30 * tStep * math.Cos(phi-1.9*tStep-spatialPsi)
			}
			ang := evalVelocityAngle(cx, cy) + strokeBend*float64(k) + whip
			cx += math.Cos(ang) * stepLen
			cy += math.Sin(ang) * stepLen
			pts[k] = render.Pt(cx, cy)
		}

		rootMacro := evalMacroColor(nx0, ny0)
		lab := render.RGBAToOklab(rootMacro)
		waveDL, waveDC := evalNebulaSpiralWave(pts[traceSteps/2].X, pts[traceSteps/2].Y)

		dL := -0.09 + 0.20*rng.Float64() + waveDL
		dA := (rng.Float64()-0.5)*0.032 + waveDC
		dB := (rng.Float64()-0.5)*0.032 - waveDC*0.55

		litCol := render.OklabToRGBA(render.Oklab{
			L: common.Constrain(lab.L+dL, 0.04, 0.98),
			A: lab.A + dA,
			B: lab.B + dB,
		}, 255)

		flankCol := render.OklabToRGBA(render.Oklab{
			L: common.Constrain(lab.L*0.60, 0.02, 0.84),
			A: lab.A * 0.90,
			B: lab.B * 0.90,
		}, 255)

		flankHalfW := 0.82 * scale
		coreHalfW := (0.42 + 0.38*rng.Float64()) * scale

		if flankPath := buildBristleRibbonPath(pts[:], flankHalfW, 0.36*scale, 0.46*scale); flankPath != nil {
			rc.RasterizePath(flankPath, flankCol, 0.74)
		}
		if corePath := buildBristleRibbonPath(pts[:], coreHalfW, 0, 0); corePath != nil {
			rc.RasterizePath(corePath, litCol, 0.88)
		}
	}

	// 4. ~65 Glowing Starlight Pinpoint Highlights with Dark Crescents & Orbiting Phase
	darkCrescent := color.RGBA{R: 11, G: 8, B: 24, A: 255}
	starWhite := color.RGBA{R: 255, G: 252, B: 246, A: 255}

	for i := range eddies {
		e := &eddies[i]
		nSpots := 3
		maxDist := 18.0 * scale
		if e.isMacro {
			nSpots = 12
			maxDist = 42.0 * scale
		}
		for k := 0; k < nSpots; k++ {
			offAng := rng.Float64()*2.0*math.Pi + e.spin*phi
			offDist := (1.8 + rng.Float64()*maxDist/scale) * scale
			px := e.x + math.Cos(offAng)*offDist
			py := e.y + math.Sin(offAng)*offDist
			rad := (1.35 + 1.15*rng.Float64()) * scale

			haloCol := iceMint
			if e.isWarmA {
				haloCol = whiteHotPeach
			}
			cAng := offAng + e.spin*math.Pi*0.5
			rc.DrawGlowDot(px+math.Cos(cAng)*0.88*scale, py+math.Sin(cAng)*0.88*scale, rad*1.38, darkCrescent, 0.80, 0.82)
			rc.DrawGlowDot(px, py, rad*1.12, haloCol, 0.94, 0.85)
			rc.DrawGlowDot(px-0.25*scale, py-0.25*scale, rad*0.58, starWhite, 0.97, 0.92)
		}
	}

	c.MergeRenderCanvas(rc)
}
