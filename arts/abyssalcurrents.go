package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type abyssalCurrents struct {
	jetBands    int
	strokeCount int
	meanderAmp  float64
}

// NewAbyssalCurrents returns an abyssalCurrents engine that renders a dark obsidian-bordered
// meandering Kuroshio S-curve jet stream and 18 flanking Kármán rollup vortices in bioluminescent cyan, neon rose-magenta, and tangerine-gold.
func NewAbyssalCurrents(jetBands, strokeCount int, meanderAmp float64) *abyssalCurrents {
	if jetBands <= 0 {
		jetBands = 18
	}
	if strokeCount <= 0 {
		strokeCount = 23000
	}
	if meanderAmp <= 0 {
		meanderAmp = 1.0
	}
	return &abyssalCurrents{
		jetBands:    jetBands,
		strokeCount: strokeCount,
		meanderAmp:  meanderAmp,
	}
}

type karmanEddy struct {
	x, y       float64
	radius     float64
	spin       float64
	strength   float64
	upperFlank bool
}

// Generative renders the static AbyssalCurrents composition.
func (ac *abyssalCurrents) Generative(c *generativeart.Canva) {
	ac.GenerativeFrame(c, 0, 1, 0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360-degree looping animation over phase in [0, 1).
func (ac *abyssalCurrents) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
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
		seed = 20261030
	}
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	phi := 2.0 * math.Pi * phase
	loopX := 0.07 * math.Sin(phi)
	loopY := 0.07 * (1.0 - math.Cos(phi))

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	// Deep Obsidian-Slate matte gallery border (#0d1018) with subtle slate paper grain
	obsidianBorder := color.RGBA{R: 13, G: 16, B: 24, A: 255}
	obsidianGrain := color.RGBA{R: 20, G: 25, B: 38, A: 255}
	rc.FillBackground(obsidianBorder)

	// Palette: Bioluminescent Abyssal Cyan, Neon Rose-Magenta, Sunset Tangerine & Trench Navy
	// Central Jet Stream Core
	iceAqua := color.RGBA{R: 168, G: 252, B: 255, A: 255}     // #a8fcff
	electricCyan := color.RGBA{R: 0, G: 238, B: 216, A: 255}  // #00eed8
	deepCerulean := color.RGBA{R: 20, G: 120, B: 200, A: 255} // #1478c8

	// Upper-Left Shear Eddies
	neonRose := color.RGBA{R: 255, G: 42, B: 120, A: 255}     // #ff2a78
	coralPink := color.RGBA{R: 255, G: 112, B: 150, A: 255}   // #ff7096
	royalAmethyst := color.RGBA{R: 120, G: 36, B: 184, A: 255} // #7824b8

	// Lower-Right Shear Eddies & Abyssal Trench
	moltenTangerine := color.RGBA{R: 255, G: 158, B: 27, A: 255} // #ff9e1b
	amberLime := color.RGBA{R: 184, G: 230, B: 46, A: 255}       // #b8e62e
	trenchNavy := color.RGBA{R: 7, G: 20, B: 46, A: 255}         // #07142e
	midnightVoid := color.RGBA{R: 5, G: 10, B: 24, A: 255}       // #050a18

	// Diagonal jet axis from bottom-left (0.08, 0.92) to top-right (0.92, 0.08)
	// Unit axis uJet = (1/sqrt(2), -1/sqrt(2)), unit normal nJet = (-1/sqrt(2), -1/sqrt(2)) (points toward upper-left)
	const invSqrt2 = 0.7071067811865475
	const axisSpan = 0.84 * 1.4142135623730951 // length between (0.08,0.92) and (0.92,0.08)

	jetCoords := func(nx, ny float64) (sJet, dJet, tangentAngle float64) {
		dx := nx - 0.08
		dy := ny - 0.92
		along := dx*invSqrt2 - dy*invSqrt2
		sJet = along / axisSpan
		// Transverse distance toward upper-left is negative of (dx*invSqrt2 + dy*invSqrt2)
		trans := -(dx*invSqrt2 + dy*invSqrt2)
		dCenter := 0.11 * ac.meanderAmp * math.Sin(sJet*math.Pi*2.2-0.4)
		dJet = trans - dCenter

		// Derivative ddCenter/dsJet in world normal vs axis
		slope := (0.11 * ac.meanderAmp * math.Pi * 2.2 / axisSpan) * math.Cos(sJet*math.Pi*2.2-0.4)
		// Base jet direction is (-45 deg = -pi/4), deflected toward normal (-3*pi/4) by slope
		tx := invSqrt2 - slope*invSqrt2
		ty := -invSqrt2 - slope*invSqrt2
		tangentAngle = math.Atan2(ty, tx)
		return sJet, dJet, tangentAngle
	}

	// Place jetBands (18) alternating Kármán vortex street eddies along both shear flanks (dJet ≈ ±0.14 and ±0.28)
	vortices := make([]karmanEddy, 0, ac.jetBands)
	for i := 0; i < ac.jetBands; i++ {
		t := (float64(i) + 0.5) / float64(ac.jetBands)
		sPos := 0.05 + 0.90*t + (rng.Float64()-0.5)*0.025
		dCenter := 0.11 * ac.meanderAmp * math.Sin(sPos*math.Pi*2.2-0.4)

		upperFlank := i%2 == 0
		tierOuter := (i/2)%3 == 1
		magD := 0.145 + 0.025*rng.Float64()
		if tierOuter {
			magD = 0.265 + 0.030*rng.Float64()
		}
		signedD := magD
		spin := 1.0 // counter-clockwise rollup on upper-left shear flank
		if !upperFlank {
			signedD = -magD
			spin = -1.0 // clockwise rollup on lower-right shear flank
		}

		trans := dCenter + signedD
		along := sPos * axisSpan
		nx := 0.08 + along*invSqrt2 - trans*invSqrt2
		ny := 0.92 - along*invSqrt2 - trans*invSqrt2

		rad := (44.0 + 24.0*rng.Float64()) * scale
		vortices = append(vortices, karmanEddy{
			x:          common.Constrain(nx, 0.11, 0.89) * fw,
			y:          common.Constrain(ny, 0.11, 0.89) * fh,
			radius:     rad,
			spin:       spin,
			strength:   0.95 + 0.35*rng.Float64(),
			upperFlank: upperFlank,
		})
	}

	evalMacroColor := func(nx, ny float64) color.RGBA {
		nx = common.Clamp01(nx)
		ny = common.Clamp01(ny)

		w1 := noise.FBM2D(nx*2.6+0.9+loopX, ny*2.6-0.8+loopY, 3, 2.0, 0.5) * 0.055
		w2 := noise.FBM2D(nx*5.0-1.4-loopX*0.7, ny*5.0+1.6+loopY*0.7, 2, 2.0, 0.5) * 0.028
		fineN := 0.5 + 0.5*noise.Simplex2D(nx*6.4+2.1+loopX, ny*6.4-1.9+loopY)

		sJet, dJet, _ := jetCoords(nx+w1*0.6, ny+w2*0.6)
		absD := math.Abs(dJet)

		// Deep abyssal trench navy & midnight void base
		trenchMix := common.Smoothstep(-0.35, 0.35, dJet+w1)
		baseCol := render.LerpOklab(midnightVoid, trenchNavy, 0.35+0.55*fineN)
		baseCol = render.LerpOklab(baseCol, royalAmethyst, common.Smoothstep(0.08, 0.36, dJet)*0.48)
		_ = trenchMix

		// Upper-Left Shear Flank (dJet > 0.05): Royal Amethyst -> Neon Rose-Magenta -> Coral-Pink
		if dJet > 0.03 {
			ulBand := common.Smoothstep(0.03, 0.16, dJet) * common.Smoothstep(0.42, 0.15, dJet)
			ulCol := render.LerpOklab(royalAmethyst, neonRose, common.Smoothstep(0.15, 0.72, ulBand+fineN*0.25))
			ulCol = render.LerpOklab(ulCol, coralPink, common.Smoothstep(0.68, 0.98, ulBand+fineN*0.30))
			baseCol = render.LerpOklab(baseCol, ulCol, ulBand*0.92)
		}

		// Lower-Right Shear Flank (dJet < -0.03): Deep Trench Navy -> Molten Tangerine-Gold -> Amber-Lime
		if dJet < -0.03 {
			lrBand := common.Smoothstep(-0.03, -0.16, dJet) * common.Smoothstep(-0.42, -0.15, dJet)
			lrCol := render.LerpOklab(trenchNavy, moltenTangerine, common.Smoothstep(0.12, 0.70, lrBand+fineN*0.25))
			lrCol = render.LerpOklab(lrCol, amberLime, common.Smoothstep(0.66, 0.96, lrBand+fineN*0.28)*0.75)
			baseCol = render.LerpOklab(baseCol, lrCol, lrBand*0.92)
		}

		// Kármán vortex ammonite spiral color enrichment
		px, py := nx*fw, ny*fh
		for i := range vortices {
			v := &vortices[i]
			dx := px - v.x
			dy := py - v.y
			reach := 1.35 * v.radius
			if math.Abs(dx) >= reach || math.Abs(dy) >= reach {
				continue
			}
			d := math.Hypot(dx, dy)
			if d < reach {
				u := 1.0 - d/reach
				theta := math.Atan2(dy, dx)
				spiral := 0.5 + 0.5*math.Sin(2.0*theta-v.spin*phi-(d/v.radius)*4.6+float64(i)*0.7)
				if v.upperFlank {
					eddyCol := render.LerpOklab(royalAmethyst, neonRose, spiral)
					eddyCol = render.LerpOklab(eddyCol, coralPink, u*spiral*0.75)
					baseCol = render.LerpOklab(baseCol, eddyCol, u*u*0.82)
				} else {
					eddyCol := render.LerpOklab(moltenTangerine, amberLime, spiral*0.68)
					baseCol = render.LerpOklab(baseCol, eddyCol, u*u*0.82)
				}
			}
		}

		// Central Meandering S-Curve Jet Stream Core (|dJet| < 0.13)
		if absD < 0.145 {
			jetCore := common.Smoothstep(0.145, 0.0, absD)
			longWave := 0.5 + 0.5*math.Sin(sJet*math.Pi*6.0-2.0*phi)
			jetCol := render.LerpOklab(deepCerulean, electricCyan, common.Smoothstep(0.12, 0.65, jetCore+longWave*0.18))
			jetCol = render.LerpOklab(jetCol, iceAqua, common.Smoothstep(0.62, 0.96, jetCore+longWave*0.22+fineN*0.12))
			baseCol = render.LerpOklab(baseCol, jetCol, jetCore*0.95)
		}

		return baseCol
	}

	// Ragged deckle-edged inner canvas [0.068*w, 0.932*w] x [0.060*h, 0.940*h]
	leftEdge := func(ny float64) float64 {
		return (0.068 + 0.016*noise.FBM2D(ny*5.5+1.2, 1.5, 3, 2.0, 0.5)) * fw
	}
	rightEdge := func(ny float64) float64 {
		return (0.932 + 0.016*noise.FBM2D(ny*5.5+4.6, 2.4, 3, 2.0, 0.5)) * fw
	}
	topEdge := func(nx float64) float64 {
		return (0.060 + 0.016*noise.FBM2D(nx*5.5+2.5, 3.8, 3, 2.0, 0.5)) * fh
	}
	bottomEdge := func(nx float64) float64 {
		return (0.940 + 0.016*noise.FBM2D(nx*5.5+6.3, 5.1, 3, 2.0, 0.5)) * fh
	}

	// 1. Base Underpainting + Subtle Dark Obsidian Slate Paper Grain
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

			grain := 0.5 + 0.5*noise.Simplex2D(fx*0.36, fy*0.36)
			borderCol := render.LerpOklab(obsidianBorder, obsidianGrain, grain*0.32)

			if fx >= xMin && fx <= xMax && fy >= yMin && fy <= yMax {
				dEdge := math.Min(math.Min(fx-xMin, xMax-fx), math.Min(fy-yMin, yMax-fy))
				edgeCov := common.Smoothstep(0.0, 3.5*scale, dEdge)

				macro := evalMacroColor(nx, ny)
				lab := render.RGBAToOklab(macro)
				underCol := render.OklabToRGBA(render.Oklab{
					L: common.Constrain(lab.L*0.64, 0.03, 0.88),
					A: lab.A * 0.90,
					B: lab.B * 0.90,
				}, 255)
				finalCol := render.LerpOklab(borderCol, underCol, edgeCov)
				rL, gL, bL, _ := render.ColorToLinearRGB(finalCol)
				rc.SetPixelLinear(px, py, rL, gL, bL, 1.0)
			} else {
				rL, gL, bL, _ := render.ColorToLinearRGB(borderCol)
				rc.SetPixelLinear(px, py, rL, gL, bL, 1.0)
			}
		}
	}

	// 2. Locked Meandering Kuroshio Jet + 18 Kármán Vortex Street Velocity Field
	evalVelocityAngle := func(x, y float64) float64 {
		nx := x / fw
		ny := y / fh
		curlVec := noise.Curl2D(nx*2.5+0.4, ny*2.5-0.6, 0.01)
		curlAng := math.Atan2(curlVec.Y, curlVec.X)

		vxSum := math.Cos(curlAng) * 0.22
		vySum := math.Sin(curlAng) * 0.22

		// Swift tangent flow along the meandering S-curve jet core (|dJet| < 0.11)
		_, dJet, jetAng := jetCoords(nx, ny)
		jetWeight := 2.45 * math.Exp(-(dJet*dJet)/(2.0*0.095*0.095))
		vxSum += math.Cos(jetAng) * jetWeight
		vySum += math.Sin(jetAng) * jetWeight

		// 18 flanking Kármán rollup vortices curling streamlines into tight ammonite spirals
		for i := range vortices {
			v := &vortices[i]
			dx := x - v.x
			dy := y - v.y
			reach := 1.42 * v.radius
			if math.Abs(dx) >= reach || math.Abs(dy) >= reach {
				continue
			}
			d := math.Hypot(dx, dy)
			if d < reach {
				u := 1.0 - d/reach
				wv := u * u * v.strength * 2.55
				ang := math.Atan2(dy, dx) + v.spin*(math.Pi*0.44)
				vxSum += math.Cos(ang) * wv
				vySum += math.Sin(ang) * wv
			}
		}
		return math.Atan2(vySum, vxSum)
	}

	// Coherent travelling jet wave packet + rotating 2-arm spiral waves inside all 18 Kármán vortices
	evalAbyssalWave := func(x, y float64) (dLum, dChroma float64) {
		nx, ny := x/fw, y/fh
		sJet, dJet, _ := jetCoords(nx, ny)

		// Travelling longitudinal wave packet racing along the jet stream
		jetEnv := math.Exp(-(dJet * dJet) / (2.0 * 0.12 * 0.12))
		jetPacket := math.Sin(sJet*math.Pi*6.0 - 2.0*phi)
		dLum += jetEnv * jetPacket * 0.12
		dChroma += jetEnv * math.Cos(sJet*math.Pi*6.0-2.0*phi) * 0.038

		for i := range vortices {
			v := &vortices[i]
			dx := x - v.x
			dy := y - v.y
			reach := 1.48 * v.radius
			if math.Abs(dx) >= reach || math.Abs(dy) >= reach {
				continue
			}
			d := math.Hypot(dx, dy)
			if d < reach {
				u := 1.0 - d/reach
				env := u * u * (3.0 - 2.0*u)
				theta := math.Atan2(dy, dx)
				spiralPhase := 2.0*theta - v.spin*phi - (d/v.radius)*4.5 + float64(i)*0.63
				w := math.Sin(spiralPhase)
				dLum += env * w * 0.105
				dChroma += env * math.Cos(spiralPhase) * 0.035
			}
		}
		return common.Constrain(dLum, -0.18, 0.18), common.Constrain(dChroma, -0.055, 0.055)
	}

	// 3. 23,000+ Glowing Bioluminescent Bristle Ribbons Splaying Over the Dark Obsidian Border
	const traceSteps = 14
	stepLen := 2.35 * scale
	var pts [traceSteps + 1]render.Point

	for s := 0; s < ac.strokeCount; s++ {
		u := rng.Float64()
		v := rng.Float64()
		// Bias perimeter strokes so glowing neon bristles splay out over the dark obsidian border
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

		pts[0] = render.Pt(x0, y0)
		cx, cy := x0, y0
		strokeBend := (rng.Float64() - 0.5) * 0.050
		for k := 1; k <= traceSteps; k++ {
			ang := evalVelocityAngle(cx, cy) + strokeBend*float64(k)
			cx += math.Cos(ang) * stepLen
			cy += math.Sin(ang) * stepLen
			pts[k] = render.Pt(cx, cy)
		}

		rootMacro := evalMacroColor(nx0, ny0)
		lab := render.RGBAToOklab(rootMacro)
		waveDL, waveDC := evalAbyssalWave(pts[traceSteps/2].X, pts[traceSteps/2].Y)

		dL := -0.09 + 0.21*rng.Float64() + waveDL
		dA := (rng.Float64()-0.5)*0.032 + waveDC
		dB := (rng.Float64()-0.5)*0.032 - waveDC*0.55

		litCol := render.OklabToRGBA(render.Oklab{
			L: common.Constrain(lab.L+dL, 0.04, 0.98),
			A: lab.A + dA,
			B: lab.B + dB,
		}, 255)

		flankCol := render.OklabToRGBA(render.Oklab{
			L: common.Constrain(lab.L*0.58, 0.02, 0.84),
			A: lab.A * 0.90,
			B: lab.B * 0.90,
		}, 255)

		flankHalfW := 0.82 * scale
		coreHalfW := (0.42 + 0.38*rng.Float64()) * scale

		if flankPath := buildBristleRibbonPath(pts[:], flankHalfW, 0.36*scale, 0.46*scale); flankPath != nil {
			rc.RasterizePath(flankPath, flankCol, 0.76)
		}
		if corePath := buildBristleRibbonPath(pts[:], coreHalfW, 0, 0); corePath != nil {
			rc.RasterizePath(corePath, litCol, 0.90)
		}
	}

	// 4. Orbiting Bioluminescent Plankton Pinpoints inside all 18 Kármán Vortices & Along the Jet Stream
	darkCrescent := color.RGBA{R: 5, G: 8, B: 18, A: 255}
	planktonCore := color.RGBA{R: 246, G: 255, B: 254, A: 255}

	for i := range vortices {
		v := &vortices[i]
		nSpots := 3
		for k := 0; k < nSpots; k++ {
			offAng := rng.Float64()*2.0*math.Pi + v.spin*phi
			offDist := (1.8 + 14.0*rng.Float64()) * scale
			px := v.x + math.Cos(offAng)*offDist
			py := v.y + math.Sin(offAng)*offDist
			rad := (1.3 + 1.1*rng.Float64()) * scale

			glowCol := amberLime
			if v.upperFlank {
				glowCol = coralPink
			}
			cAng := offAng + v.spin*math.Pi*0.5
			rc.DrawGlowDot(px+math.Cos(cAng)*0.85*scale, py+math.Sin(cAng)*0.85*scale, rad*1.35, darkCrescent, 0.82, 0.82)
			rc.DrawGlowDot(px, py, rad*1.12, glowCol, 0.94, 0.85)
			rc.DrawGlowDot(px-0.25*scale, py-0.25*scale, rad*0.55, planktonCore, 0.97, 0.92)
		}
	}

	// Travelling bioluminescent plankton sparks drifting along the central Kuroshio S-curve jet core
	for j := 0; j < 20; j++ {
		sBase := rng.Float64()
		dOff := (rng.Float64() - 0.5) * 0.08
		sAnim := math.Mod(sBase+phase, 1.0)
		if sAnim < 0 {
			sAnim += 1.0
		}
		dCenter := 0.11 * ac.meanderAmp * math.Sin(sAnim*math.Pi*2.2-0.4)
		trans := dCenter + dOff
		along := sAnim * axisSpan
		nx := 0.08 + along*invSqrt2 - trans*invSqrt2
		ny := 0.92 - along*invSqrt2 - trans*invSqrt2
		px, py := nx*fw, ny*fh
		rad := (1.25 + 0.95*rng.Float64()) * scale
		rc.DrawGlowDot(px, py, rad*1.15, iceAqua, 0.92, 0.84)
		rc.DrawGlowDot(px-0.2*scale, py-0.2*scale, rad*0.55, planktonCore, 0.96, 0.92)
	}

	c.MergeRenderCanvas(rc)
}
