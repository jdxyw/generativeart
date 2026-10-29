package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type botanicalWhorls struct {
	whorlCount  int
	strokeCount int
	phylloAngle float64
}

// NewBotanicalWhorls returns a botanicalWhorls engine that renders multi-scale Fibonacci
// golden-angle lotus whorls, medium succulent rosettes, and micro-buds in Monet Water-Lilies impasto.
func NewBotanicalWhorls(whorlCount, strokeCount int, phylloAngle float64) *botanicalWhorls {
	if whorlCount <= 0 {
		whorlCount = 64
	}
	if strokeCount <= 0 {
		strokeCount = 22000
	}
	if phylloAngle <= 0 {
		phylloAngle = 137.507764
	}
	return &botanicalWhorls{
		whorlCount:  whorlCount,
		strokeCount: strokeCount,
		phylloAngle: phylloAngle,
	}
}

type phytoWhorl struct {
	x, y       float64
	radius     float64
	spin       float64
	strength   float64
	tier       int // 0 = giant lotus whorl (54..66), 1 = medium succulent (32..44), 2 = micro-bud (14..24)
	fibIdx     int
	hasDewdrop bool
}

// Generative renders the static BotanicalWhorls composition.
func (bw *botanicalWhorls) Generative(c *generativeart.Canva) {
	bw.GenerativeFrame(c, 0, 1, 0)
}

// GenerativeFrame renders frame frameIdx of a zero-flicker seamless looping animation.
func (bw *botanicalWhorls) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
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
		seed = 20261031
	}
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	isAnim := totalFrames > 1
	phi := 2.0 * math.Pi * phase
	loopX := 0.07 * math.Sin(phi)
	loopY := 0.07 * (1.0 - math.Cos(phi))

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	// Warm sage-ivory museum paper (#eeebe2)
	sagePaper := color.RGBA{R: 238, G: 235, B: 226, A: 255}
	sageShadow := color.RGBA{R: 225, G: 221, B: 210, A: 255}
	rc.FillBackground(sagePaper)

	// Monet Water-Lilies Palette: Celadon Jade, Sakura Peony-Rose, Apricot-Gold & Twilight Pond Indigo
	celadonJade := color.RGBA{R: 56, G: 176, B: 142, A: 255}    // #38b08e
	viridianEmerald := color.RGBA{R: 22, G: 110, B: 92, A: 255} // #166e5c
	mintTurquoise := color.RGBA{R: 114, G: 228, B: 196, A: 255} // #72e4c4
	deepPine := color.RGBA{R: 14, G: 50, B: 40, A: 255}         // #0e3228

	sakuraCoral := color.RGBA{R: 244, G: 104, B: 124, A: 255} // #f4687c
	peonyMagenta := color.RGBA{R: 200, G: 40, B: 100, A: 255} // #c82864
	apricotGold := color.RGBA{R: 248, G: 172, B: 66, A: 255}  // #f8ac42
	lotusIvory := color.RGBA{R: 252, G: 244, B: 220, A: 255}  // #fcf4dc

	pondIndigo := color.RGBA{R: 24, G: 42, B: 88, A: 255}  // #182a58
	plumViolet := color.RGBA{R: 56, G: 30, B: 84, A: 255}  // #381e54

	// Ragged deckle-edged inner canvas [0.068*w, 0.932*w] x [0.060*h, 0.940*h]
	leftEdge := func(ny float64) float64 {
		return (0.068 + 0.015*noise.FBM2D(ny*5.4+0.9, 1.3, 3, 2.0, 0.5)) * fw
	}
	rightEdge := func(ny float64) float64 {
		return (0.932 + 0.015*noise.FBM2D(ny*5.4+4.1, 2.9, 3, 2.0, 0.5)) * fw
	}
	topEdge := func(nx float64) float64 {
		return (0.060 + 0.015*noise.FBM2D(nx*5.4+2.3, 3.7, 3, 2.0, 0.5)) * fh
	}
	bottomEdge := func(nx float64) float64 {
		return (0.940 + 0.015*noise.FBM2D(nx*5.4+6.1, 5.1, 3, 2.0, 0.5)) * fh
	}

	// Build Multi-Scale Whorl Hierarchy (64 whorls: 6 giant lotus [54..66*scale], 18 medium succulent [32..44*scale], 40 micro-buds [14..24*scale])
	whorls := make([]phytoWhorl, 0, bw.whorlCount)
	goldenRad := bw.phylloAngle * math.Pi / 180.0
	focalX, focalY := 0.56*fw, 0.54*fh
	maxReach := 0.46 * math.Min(fw, fh)

	// Corner fill anchors so the entire rectangle is lushly covered
	cornerOffsets := [][2]float64{
		{0.15, 0.16}, {0.85, 0.15}, {0.15, 0.85}, {0.86, 0.85},
		{0.14, 0.48}, {0.87, 0.50}, {0.48, 0.14}, {0.48, 0.87},
	}

	giantQuota := 6
	mediumQuota := 18
	if bw.whorlCount < 24 {
		giantQuota = int(math.Max(1, float64(bw.whorlCount)*0.12))
		mediumQuota = int(math.Max(2, float64(bw.whorlCount)*0.28))
	}

	for i := 0; i < bw.whorlCount; i++ {
		var wx, wy float64
		if i < bw.whorlCount-len(cornerOffsets) {
			nFrac := (float64(i) + 0.5) / float64(bw.whorlCount-len(cornerOffsets))
			r := math.Sqrt(nFrac) * maxReach
			theta := float64(i) * goldenRad
			// Slight organic FBM warp on Fermat spiral coordinates
			warpX := noise.Simplex2D(math.Cos(theta)*1.7, math.Sin(theta)*1.7) * 18.0 * scale
			warpY := noise.Simplex2D(math.Sin(theta)*1.7+3.1, math.Cos(theta)*1.7-2.4) * 18.0 * scale
			wx = focalX + math.Cos(theta)*r*1.08 + warpX
			wy = focalY + math.Sin(theta)*r*1.04 + warpY
		} else {
			cIdx := i - (bw.whorlCount - len(cornerOffsets))
			co := cornerOffsets[cIdx%len(cornerOffsets)]
			wx = (co[0] + (rng.Float64()-0.5)*0.08) * fw
			wy = (co[1] + (rng.Float64()-0.5)*0.08) * fh
		}

		// Clamp cleanly inside the inner canvas rectangle
		wx = common.Constrain(wx, 0.095*fw, 0.905*fw)
		wy = common.Constrain(wy, 0.090*fh, 0.910*fh)

		var rad float64
		var tier int
		var strength float64
		// Distribute the 6 giant lotus whorls, 18 medium succulent rosettes, and 40 micro-buds
		if i < giantQuota {
			tier = 0
			rad = (54.0 + 12.0*rng.Float64()) * scale
			strength = 1.15 + 0.25*rng.Float64()
		} else if i < giantQuota+mediumQuota {
			tier = 1
			rad = (32.0 + 12.0*rng.Float64()) * scale
			strength = 0.92 + 0.25*rng.Float64()
		} else {
			tier = 2
			rad = (14.0 + 10.0*rng.Float64()) * scale
			strength = 0.76 + 0.24*rng.Float64()
		}

		spin := 1.0
		if i%2 == 1 {
			spin = -1.0
		}
		whorls = append(whorls, phytoWhorl{
			x:          wx,
			y:          wy,
			radius:     rad,
			spin:       spin,
			strength:   strength,
			tier:       tier,
			fibIdx:     i,
			hasDewdrop: tier <= 1 || rng.Float64() < 0.72,
		})
	}

	// Evaluate Monet Water-Lilies macro color + whorl blossom/succulent tinting
	evalBotanicalColor := func(x, y float64) color.RGBA {
		nx := common.Clamp01(x / fw)
		ny := common.Clamp01(y / fh)

		w1 := noise.FBM2D(nx*2.5+1.1+loopX, ny*2.5+0.6+loopY, 3, 2.0, 0.5) * 0.11
		w2 := noise.FBM2D(nx*4.8-1.4-loopX*0.7, ny*4.8+2.3+loopY*0.7, 2, 2.0, 0.5) * 0.05
		fineN := 0.5 + 0.5*noise.Simplex2D(nx*7.0+2.9+loopX, ny*7.0-1.5+loopY)

		// 1. Upper-Right & Center Jade Zone: Lush Celadon Jade, Viridian Emerald, Mint-Turquoise, Deep Pine
		jadeBase := render.LerpOklab(celadonJade, viridianEmerald, common.Smoothstep(0.16, 0.74, ny-0.15*nx+w1))
		mintCrown := common.Smoothstep(0.45, 0.85, nx-0.35*ny+w2)
		jadeBase = render.LerpOklab(jadeBase, mintTurquoise, mintCrown*0.48*fineN)
		pineDepth := common.Smoothstep(0.55, 0.92, noise.FBM2D(nx*3.8+0.4, ny*3.8-0.8, 2, 2.0, 0.5))
		baseCol := render.LerpOklab(jadeBase, deepPine, pineDepth*0.36)

		// 2. Lower-Left & Blossom Whorl Zone: Sakura Coral-Rose, Peony Magenta, Warm Apricot-Gold, Creamy Lotus Ivory
		blossomDiag := (ny - 0.34) - 0.88*nx + w1 + w2
		if blossomDiag > -0.10 {
			blossomMask := common.Smoothstep(-0.10, 0.08, blossomDiag)
			roseCol := render.LerpOklab(peonyMagenta, sakuraCoral, common.Smoothstep(0.15, 0.75, fineN+w2))
			goldAccent := common.Smoothstep(0.48, 0.86, noise.FBM2D(nx*4.2+1.7, ny*4.2+0.5, 2, 2.0, 0.5))
			roseCol = render.LerpOklab(roseCol, apricotGold, goldAccent*0.62)
			ivoryTip := common.Smoothstep(0.72, 0.96, fineN) * 0.45
			roseCol = render.LerpOklab(roseCol, lotusIvory, ivoryTip)
			baseCol = render.LerpOklab(baseCol, roseCol, blossomMask*0.92)
		}

		// 3. Diagonal Twilight Pond Shadow Band: Prussian Pond Indigo (#182a58) & Deep Plum-Violet (#381e54)
		bandY := 0.22 + 0.52*nx + w1*0.85
		dBand := math.Abs(ny-bandY) / (0.12 + 0.04*math.Sin(nx*math.Pi))
		if dBand < 1.25 {
			pondMask := common.Smoothstep(1.25, 0.20, dBand) * common.Smoothstep(0.04, 0.16, nx) * common.Smoothstep(0.95, 0.78, nx)
			pondCol := render.LerpOklab(pondIndigo, plumViolet, common.Smoothstep(0.20, 0.80, fineN+w2))
			pondCol = render.LerpOklab(pondCol, deepPine, 0.22*(1.0-fineN))
			baseCol = render.LerpOklab(baseCol, pondCol, pondMask*0.88)
		}

		// 4. Local Whorl Petal Heart Tinting (Giant Lotus Whorls glow with Creamy Lotus Ivory, Apricot-Gold & Sakura Rose)
		for i := range whorls {
			wh := &whorls[i]
			dx := x - wh.x
			dy := y - wh.y
			if math.Abs(dx) >= wh.radius || math.Abs(dy) >= wh.radius {
				continue
			}
			d := math.Hypot(dx, dy)
			if d < wh.radius {
				u := 1.0 - d/wh.radius
				if wh.tier == 0 {
					// Giant lotus whorl heart: sakura coral -> apricot gold -> creamy lotus ivory core
					petalCol := render.LerpOklab(sakuraCoral, apricotGold, common.Smoothstep(0.20, 0.68, u))
					petalCol = render.LerpOklab(petalCol, lotusIvory, common.Smoothstep(0.58, 0.92, u)*0.75)
					baseCol = render.LerpOklab(baseCol, petalCol, u*u*0.64)
				} else if wh.tier == 1 && i%3 == 0 {
					// Select medium succulent rosettes tipped with mint-turquoise & apricot-gold
					succCol := render.LerpOklab(mintTurquoise, apricotGold, common.Smoothstep(0.35, 0.85, u))
					baseCol = render.LerpOklab(baseCol, succCol, u*u*0.45)
				}
			}
		}

		return baseCol
	}

	// 1. Base Underpainting + Sage-Ivory Paper Grain
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

			grain := 0.5 + 0.5*noise.Simplex2D(fx*0.35, fy*0.35)
			paperCol := render.LerpOklab(sagePaper, sageShadow, grain*0.18)

			if fx >= xMin && fx <= xMax && fy >= yMin && fy <= yMax {
				dEdge := math.Min(math.Min(fx-xMin, xMax-fx), math.Min(fy-yMin, yMax-fy))
				edgeCov := common.Smoothstep(0.0, 3.5*scale, dEdge)

				macro := evalBotanicalColor(fx, fy)
				lab := render.RGBAToOklab(macro)
				underCol := render.OklabToRGBA(render.Oklab{
					L: common.Constrain(lab.L*0.66, 0.04, 0.90),
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

	// 2. Multi-Scale Fibonacci Whorl Velocity Field (locked across frames)
	evalVelocityAngle := func(x, y float64) float64 {
		nx, ny := x/fw, y/fh
		curlVec := noise.Curl2D(nx*2.2+0.6, ny*2.2-0.4, 0.01)
		vxSum := curlVec.X * 0.22
		vySum := curlVec.Y * 0.22

		for i := range whorls {
			wh := &whorls[i]
			dx := x - wh.x
			dy := y - wh.y
			reach := 1.42 * wh.radius
			if math.Abs(dx) >= reach || math.Abs(dy) >= reach {
				continue
			}
			d := math.Hypot(dx, dy)
			if d < reach {
				u := 1.0 - d/reach
				wv := u * u * wh.strength * 2.4
				// Petal whorl angle: spiral inward with a gentle 3-lobed botanical phyllotaxis modulation
				theta := math.Atan2(dy, dx)
				petalMod := 0.08 * math.Sin(3.0*theta+float64(wh.fibIdx)*0.5)
				ang := theta + wh.spin*(math.Pi*0.43+petalMod)
				vxSum += math.Cos(ang) * wv
				vySum += math.Sin(ang) * wv
			}
		}
		return math.Atan2(vySum, vxSum)
	}

	// Coherent zero-flicker Fibonacci blooming wave + rotating 3-petal lotus luminescence wave
	evalLotusBloomWave := func(x, y float64) (dLum, dChroma float64) {
		if !isAnim {
			return 0, 0
		}
		// Radial Fibonacci wave rippling outward from the off-center golden-angle origin
		distFocal := math.Hypot(x-focalX, y-focalY) / maxReach
		fibWave := math.Sin(distFocal*5.5 - phi)
		dLum += 0.04 * fibWave

		for i := range whorls {
			wh := &whorls[i]
			dx := x - wh.x
			dy := y - wh.y
			reach := 1.45 * wh.radius
			if math.Abs(dx) >= reach || math.Abs(dy) >= reach {
				continue
			}
			d := math.Hypot(dx, dy)
			if d < reach {
				u := 1.0 - d/reach
				env := u * u * (3.0 - 2.0*u)
				theta := math.Atan2(dy, dx)
				// Rotating 3-petal lotus luminescence wave inside each whorl synchronized with its Fibonacci index
				petalPhase := 3.0*theta - wh.spin*phi - (d/wh.radius)*3.6 + float64(wh.fibIdx)*0.38
				dLum += env * math.Sin(petalPhase) * 0.11
				dChroma += env * math.Cos(petalPhase) * 0.038
			}
		}
		return common.Constrain(dLum, -0.17, 0.17), common.Constrain(dChroma, -0.05, 0.05)
	}

	// 3. Trace 22,000+ Velvety Petal-Impasto Bristles (coreHalfW = 0.52..0.96*scale, flankHalfW = 0.95*scale)
	const traceSteps = 12
	stepLen := 2.2 * scale
	var pts [traceSteps + 1]render.Point

	for s := 0; s < bw.strokeCount; s++ {
		u := rng.Float64()
		v := rng.Float64()
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

		pts[0] = render.Pt(x0, y0)
		cx, cy := x0, y0
		strokeBend := (rng.Float64() - 0.5) * 0.058
		for k := 1; k <= traceSteps; k++ {
			ang := evalVelocityAngle(cx, cy) + strokeBend*float64(k)
			cx += math.Cos(ang) * stepLen
			cy += math.Sin(ang) * stepLen
			pts[k] = render.Pt(cx, cy)
		}

		rootMacro := evalBotanicalColor(x0, y0)
		lab := render.RGBAToOklab(rootMacro)
		midPt := pts[traceSteps/2]
		waveDL, waveDC := evalLotusBloomWave(midPt.X, midPt.Y)

		dL := -0.09 + 0.21*rng.Float64() + waveDL
		dA := (rng.Float64()-0.5)*0.030 + waveDC
		dB := (rng.Float64()-0.5)*0.030 - waveDC*0.5

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

		flankHalfW := 0.95 * scale
		coreHalfW := (0.52 + 0.44*rng.Float64()) * scale

		if flankPath := buildBristleRibbonPath(pts[:], flankHalfW, 0.36*scale, 0.46*scale); flankPath != nil {
			rc.RasterizePath(flankPath, flankCol, 0.75)
		}
		if corePath := buildBristleRibbonPath(pts[:], coreHalfW, 0, 0); corePath != nil {
			rc.RasterizePath(corePath, litCol, 0.89)
		}
	}

	// 4. Pearl Dewdrop Pinpoints at Whorl Centers
	pearlHalo := color.RGBA{R: 224, G: 248, B: 238, A: 255}
	pearlCore := lotusIvory

	for i := range whorls {
		wh := &whorls[i]
		if !wh.hasDewdrop {
			continue
		}
		nDrops := 1
		if wh.tier == 0 {
			nDrops = 3
		} else if wh.tier == 1 && rng.Float64() < 0.45 {
			nDrops = 2
		}
		for k := 0; k < nDrops; k++ {
			offAng := rng.Float64() * 2.0 * math.Pi
			if isAnim {
				offAng += wh.spin * phi
			}
			offDist := (1.4 + 4.2*rng.Float64()) * scale
			if wh.tier == 0 {
				offDist = (2.0 + 6.5*rng.Float64()) * scale
			}
			px := wh.x + math.Cos(offAng)*offDist
			py := wh.y + math.Sin(offAng)*offDist
			rad := (1.3 + 1.1*rng.Float64()) * scale
			if wh.tier == 0 {
				rad *= 1.18
			}

			cAng := offAng + wh.spin*math.Pi*0.5
			rc.DrawGlowDot(px+math.Cos(cAng)*0.85*scale, py+math.Sin(cAng)*0.85*scale, rad*1.35, deepPine, 0.78, 0.82)
			rc.DrawGlowDot(px, py, rad, pearlHalo, 0.94, 0.86)
			rc.DrawGlowDot(px-0.25*scale, py-0.25*scale, rad*0.58, pearlCore, 0.97, 0.93)
		}
	}

	c.MergeRenderCanvas(rc)
}
