package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type solarProminence struct {
	archCount     int
	strokeCount   int
	magneticTwist float64
}

// NewSolarProminence returns a solarProminence engine that renders a circular deckle-edged
// solar chromosphere tondo medallion with bipolar magnetic dipole arcades and coronal prominence loops.
func NewSolarProminence(archCount, strokeCount int, magneticTwist float64) *solarProminence {
	if archCount <= 0 {
		archCount = 14
	}
	if strokeCount <= 0 {
		strokeCount = 24000
	}
	if magneticTwist <= 0 {
		magneticTwist = 1.0
	}
	return &solarProminence{
		archCount:     archCount,
		strokeCount:   strokeCount,
		magneticTwist: magneticTwist,
	}
}

type solarPole struct {
	x, y       float64
	radius     float64
	charge     float64 // +1 for positive magnetic source, -1 for negative sink
	spin       float64
	strength   float64
	isRim      bool
	hasSparkle bool
}

// Generative renders the static SolarProminence composition.
func (sp *solarProminence) Generative(c *generativeart.Canva) {
	sp.GenerativeFrame(c, 0, 1, 0)
}

// GenerativeFrame renders frame frameIdx of a zero-flicker seamless looping animation.
func (sp *solarProminence) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	minDim := math.Min(fw, fh)
	scale := minDim / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	seed := c.Seed()
	if seed == 0 {
		seed = 20261030
	}
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	isAnim := totalFrames > 1
	phi := 2.0 * math.Pi * phase
	loopX := 0.07 * math.Sin(phi)
	loopY := 0.07 * (1.0 - math.Cos(phi))

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	// Warm museum linen paper (#f2ebe0)
	linenPaper := color.RGBA{R: 242, G: 235, B: 224, A: 255}
	linenShadow := color.RGBA{R: 230, G: 221, B: 207, A: 255}
	rc.FillBackground(linenPaper)

	// Solar Chromosphere Palette
	obsidianPlum := color.RGBA{R: 18, G: 8, B: 22, A: 255}       // #120816
	midnightIndigo := color.RGBA{R: 24, G: 24, B: 68, A: 255}    // #181844
	darkBurgundy := color.RGBA{R: 72, G: 12, B: 34, A: 255}      // #480c22
	whiteGold := color.RGBA{R: 255, G: 248, B: 214, A: 255}      // #fff8d6
	solarGold := color.RGBA{R: 255, G: 190, B: 24, A: 255}       // #ffbe18
	fieryVermilion := color.RGBA{R: 242, G: 68, B: 29, A: 255}   // #f2441d
	carmineRose := color.RGBA{R: 200, G: 29, B: 78, A: 255}      // #c81d4e
	electricViolet := color.RGBA{R: 78, G: 84, B: 232, A: 255}   // #4e54e8
	turquoiseTeal := color.RGBA{R: 28, G: 196, B: 184, A: 255}   // #1cc4b8

	cxDisc, cyDisc := 0.5*fw, 0.5*fh
	baseRadius := 0.422 * minDim

	// Circular deckle-edged sun disc radius R_disc(theta) = (0.422 + 0.018*FBM) * min(fw, fh)
	discRadiusAt := func(theta float64) float64 {
		ux := math.Cos(theta)*2.2 + 1.7
		uy := math.Sin(theta)*2.2 - 0.9
		fbm := noise.FBM2D(ux, uy, 3, 2.0, 0.5)
		return (0.422 + 0.018*fbm) * minDim
	}

	// Place archCount (default 14: 8 near outer chromosphere rim r ~ 0.78*R_disc, 6 in interior)
	poles := make([]solarPole, 0, sp.archCount)
	rimCount := (sp.archCount * 4) / 7
	if sp.archCount == 14 {
		rimCount = 8
	}
	if rimCount < 2 {
		rimCount = sp.archCount
	}
	interiorCount := sp.archCount - rimCount

	for i := 0; i < rimCount; i++ {
		ang := (float64(i)/float64(rimCount))*2.0*math.Pi + (rng.Float64()-0.5)*0.22
		rFrac := 0.75 + 0.06*rng.Float64() // ~0.78 * R_disc
		radDisc := baseRadius * rFrac
		px := cxDisc + math.Cos(ang)*radDisc
		py := cyDisc + math.Sin(ang)*radDisc
		charge := 1.0
		if i%2 == 1 {
			charge = -1.0
		}
		spin := 1.0
		if rng.Float64() < 0.5 {
			spin = -1.0
		}
		poles = append(poles, solarPole{
			x:          px,
			y:          py,
			radius:     (36.0 + 18.0*rng.Float64()) * scale,
			charge:     charge,
			spin:       spin,
			strength:   0.95 + 0.35*rng.Float64(),
			isRim:      true,
			hasSparkle: true,
		})
	}

	for i := 0; i < interiorCount; i++ {
		ang := (float64(i)/math.Max(1.0, float64(interiorCount)))*2.0*math.Pi + 0.38 + (rng.Float64()-0.5)*0.32
		rFrac := 0.24 + 0.34*rng.Float64()
		radDisc := baseRadius * rFrac
		px := cxDisc + math.Cos(ang)*radDisc
		py := cyDisc + math.Sin(ang)*radDisc
		charge := 1.0
		if i%2 == 1 {
			charge = -1.0
		}
		spin := 1.0
		if rng.Float64() < 0.5 {
			spin = -1.0
		}
		poles = append(poles, solarPole{
			x:          px,
			y:          py,
			radius:     (32.0 + 20.0*rng.Float64()) * scale,
			charge:     charge,
			spin:       spin,
			strength:   0.88 + 0.32*rng.Float64(),
			isRim:      false,
			hasSparkle: rng.Float64() < 0.85,
		})
	}

	if isAnim {
		for i := range poles {
			orbitPhase := phi + float64(i)*1.3
			poles[i].x += 6.0 * scale * math.Sin(orbitPhase)
			poles[i].y += 6.0 * scale * (1.0 - math.Cos(orbitPhase))
		}
	}

	// Evaluate macro chromosphere color + sunspot umbra cores + flaring arcades + coronal rim shock fronts
	evalSolarColor := func(x, y float64) (col color.RGBA, dipolePot float64) {
		dx := x - cxDisc
		dy := y - cyDisc
		r := math.Hypot(dx, dy)
		rNorm := r / baseRadius
		nx, ny := x/fw, y/fh

		w1 := noise.FBM2D(nx*3.2+1.1+loopX, ny*3.2-0.7+loopY, 3, 2.0, 0.5)
		w2 := noise.FBM2D(nx*6.4-2.3-loopX*0.7, ny*6.4+1.9+loopY*0.7, 2, 2.0, 0.5)
		granulation := 0.5 + 0.5*noise.Simplex2D(nx*9.0+0.5+loopX, ny*9.0-1.3+loopY)

		// Base fiery chromosphere: carmine-rose -> fiery vermilion -> blazing solar gold
		fireT := common.Clamp01(0.52 + 0.38*w1 + 0.18*w2 - 0.16*rNorm)
		baseCol := render.LerpOklab(carmineRose, fieryVermilion, common.Smoothstep(0.12, 0.58, fireT))
		baseCol = render.LerpOklab(baseCol, solarGold, common.Smoothstep(0.42, 0.86, fireT+0.15*granulation))

		// Evaluate magnetic dipole potential and sunspot umbra proximity
		minPoleRatio := 10.0
		potSum := 0.0
		arcadeIntensity := 0.0

		for i := range poles {
			p := &poles[i]
			pdx := x - p.x
			pdy := y - p.y
			d := math.Hypot(pdx, pdy)
			ratio := d / p.radius
			if ratio < minPoleRatio {
				minPoleRatio = ratio
			}
			// Dipole scalar potential phi_m ~ sum(q_i / sqrt(d^2 + eps^2))
			softD := math.Sqrt(d*d + (0.35*p.radius)*(0.35*p.radius))
			potSum += p.charge * p.strength * (p.radius / softD)

			// Flaring magnetic arcade ridge between nearby opposite poles
			if ratio < 2.2 {
				u := 1.0 - ratio/2.2
				arcadeIntensity += u * u * p.strength
			}
		}

		// Bright flaring magnetic arcade ribbons where |potSum| is small (neutral line between + and - poles) and arcadeIntensity is high
		neutralLine := math.Exp(-potSum * potSum * 2.4)
		flareStrength := common.Clamp01(arcadeIntensity * (0.45 + 0.65*neutralLine) * (0.75 + 0.25*granulation))
		if flareStrength > 0.15 {
			flareCol := render.LerpOklab(solarGold, whiteGold, common.Smoothstep(0.32, 0.88, flareStrength))
			baseCol = render.LerpOklab(baseCol, flareCol, common.Smoothstep(0.18, 0.82, flareStrength)*0.86)
		}

		// Coronal rim shock accents near the outer perimeter (rNorm in [0.76, 1.08])
		if rNorm > 0.72 {
			rimBand := common.Smoothstep(0.72, 0.96, rNorm) * common.Smoothstep(1.14, 0.90, rNorm)
			shockWave := 0.5 + 0.5*math.Sin(math.Atan2(dy, dx)*7.0+w1*3.5)
			shockMask := rimBand * common.Smoothstep(0.52, 0.88, shockWave+0.25*w2)
			if shockMask > 0.02 {
				shockCol := render.LerpOklab(electricViolet, turquoiseTeal, common.Smoothstep(0.25, 0.82, granulation+0.2*w1))
				baseCol = render.LerpOklab(baseCol, shockCol, shockMask*0.72)
			}
		}

		// Dark sunspot umbra & penumbra rosettes around pole centers
		if minPoleRatio < 0.92 {
			penumbra := common.Smoothstep(0.92, 0.34, minPoleRatio)
			umbra := common.Smoothstep(0.46, 0.10, minPoleRatio)
			penCol := render.LerpOklab(darkBurgundy, midnightIndigo, common.Smoothstep(0.15, 0.75, penumbra+0.2*w2))
			umbCol := render.LerpOklab(midnightIndigo, obsidianPlum, common.Smoothstep(0.10, 0.85, umbra))
			baseCol = render.LerpOklab(baseCol, penCol, penumbra*0.88)
			baseCol = render.LerpOklab(baseCol, umbCol, umbra*0.95)
		}

		return baseCol, potSum
	}

	// 1. Base Underpainting: Warm Linen Paper outside R_disc(theta) + Dark Solar Chromosphere inside
	for py := 0; py < h; py++ {
		fy := float64(py)
		for px := 0; px < w; px++ {
			fx := float64(px)
			grain := 0.5 + 0.5*noise.Simplex2D(fx*0.35, fy*0.35)
			paperCol := render.LerpOklab(linenPaper, linenShadow, grain*0.18)

			dx := fx - cxDisc
			dy := fy - cyDisc
			r := math.Hypot(dx, dy)
			theta := math.Atan2(dy, dx)
			rDisc := discRadiusAt(theta)

			if r <= rDisc {
				dEdge := rDisc - r
				edgeCov := common.Smoothstep(0.0, 4.0*scale, dEdge)
				macro, _ := evalSolarColor(fx, fy)
				lab := render.RGBAToOklab(macro)
				// Deep chromosphere underpainting so dense plasma bristles glow richly over dark ground
				underCol := render.OklabToRGBA(render.Oklab{
					L: common.Constrain(lab.L*0.58, 0.03, 0.88),
					A: lab.A * 0.94,
					B: lab.B * 0.94,
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

	// 2. Bipolar Magnetic Dipole + Helical Sunspot Rosette Velocity Field (with breathing twist & dancing rim in animation)
	twist := sp.magneticTwist
	evalVelocityAngle := func(x, y float64) float64 {
		nx, ny := x/fw, y/fh
		curlVec := noise.Curl2D(nx*2.5+0.4+loopX*0.8, ny*2.5-0.6+loopY*0.8, 0.01)
		vxSum := curlVec.X * 0.24
		vySum := curlVec.Y * 0.24

		dxC := x - cxDisc
		dyC := y - cyDisc
		rC := math.Hypot(dxC, dyC)

		for i := range poles {
			p := &poles[i]
			dx := x - p.x
			dy := y - p.y
			d2 := dx*dx + dy*dy + (8.0*scale)*(8.0*scale)
			d := math.Sqrt(d2)

			// Magnetic dipole radial source/sink vector: q_i * (r - r_i) / |r - r_i|^2.4
			radialWeight := p.strength * (p.radius * p.radius) / (d2 * math.Pow(d/p.radius+0.35, 0.45))
			rx := (dx / d) * p.charge
			ry := (dy / d) * p.charge

			// Helical vortex twist around each sunspot pole so streamlines spiral into penumbra rosettes
			localTwist := twist
			if isAnim {
				theta := math.Atan2(dy, dx)
				localTwist += 0.26 * math.Sin(2.0*theta-p.spin*phi-(d/p.radius)*3.5)
			}
			tx := -(dy / d) * p.spin * localTwist
			ty := (dx / d) * p.spin * localTwist

			rosetteEnv := math.Exp(-d / (1.35 * p.radius))
			vxSum += radialWeight * (rx*0.72 + tx*0.68*rosetteEnv)
			vySum += radialWeight * (ry*0.72 + ty*0.68*rosetteEnv)
		}

		// Near the circular rim (rC > 0.78*baseRadius), add arching coronal prominence loop curvature
		if rC > 0.76*baseRadius {
			theta := math.Atan2(dyC, dxC)
			rimFactor := common.Smoothstep(0.76*baseRadius, 1.04*baseRadius, rC)
			// Arching prominence loops that erupt outward and curve back toward the chromosphere rim
			var archWave float64
			if isAnim {
				archWave = math.Sin(theta*float64(rimCount)*0.5 - phi + noise.Simplex2D(nx*3.0+loopX, ny*3.0+loopY)*1.2)
			} else {
				archWave = math.Sin(theta*float64(rimCount)*0.5 + noise.Simplex2D(nx*3.0, ny*3.0)*1.2)
			}
			outwardAng := theta + archWave*math.Pi*0.36
			vxSum += math.Cos(outwardAng) * rimFactor * 0.65
			vySum += math.Sin(outwardAng) * rimFactor * 0.65
		}

		return math.Atan2(vySum, vxSum)
	}

	// Coherent zero-flicker plasma arch wave + rotating penumbra spiral wave for animation
	evalProminenceWave := func(x, y, dipolePot float64) (dLum, dChroma float64) {
		if !isAnim {
			return 0, 0
		}
		// 1. Plasma arch wave travelling along dipole potential from positive poles (+pot) to negative poles (-pot)
		dipolePhase := dipolePot * 3.4
		archWave := math.Sin(dipolePhase - 2.0*phi)
		dLum += 0.065 * archWave

		// 2. Rotating penumbra spiral waves around all sunspot poles
		for i := range poles {
			p := &poles[i]
			dx := x - p.x
			dy := y - p.y
			reach := 1.65 * p.radius
			if math.Abs(dx) >= reach || math.Abs(dy) >= reach {
				continue
			}
			d := math.Hypot(dx, dy)
			if d < reach {
				u := 1.0 - d/reach
				env := u * u * (3.0 - 2.0*u)
				theta := math.Atan2(dy, dx)
				spiralPhase := 3.0*theta - p.spin*phi - (d/p.radius)*3.8 + float64(i)*0.67
				dLum += env * math.Sin(spiralPhase) * 0.11
				dChroma += env * math.Cos(spiralPhase) * 0.04
			}
		}
		return common.Constrain(dLum, -0.18, 0.18), common.Constrain(dChroma, -0.05, 0.05)
	}

	// 3. Trace 24,000 fine, dense plasma bristle filaments (traceSteps = 14, stepLen = 2.3 * scale)
	const traceSteps = 14
	stepLen := 2.3 * scale
	var pts [traceSteps + 1]render.Point

	for s := 0; s < sp.strokeCount; s++ {
		u1 := rng.Float64()
		u2 := rng.Float64()
		u3 := rng.Float64()

		theta := u1 * 2.0 * math.Pi
		rDisc := discRadiusAt(theta)

		var rSpawn float64
		// Bias ~24% of strokes to the outer chromosphere rim so curved coronal prominence loops splay 12..26px across the circular deckle rim
		if s%4 == 0 {
			rSpawn = rDisc * (0.86 + 0.17*u2)
		} else {
			rSpawn = rDisc * math.Sqrt(u2) * 1.01
		}

		x0 := cxDisc + math.Cos(theta)*rSpawn
		y0 := cyDisc + math.Sin(theta)*rSpawn

		sx, sy := x0, y0
		var spatialPsi, flareBoost float64
		if isAnim {
			rNorm := math.Hypot(x0-cxDisc, y0-cyDisc) / rDisc
			thetaDisc := math.Atan2(y0-cyDisc, x0-cxDisc)
			spatialPsi = 4.0*thetaDisc - rNorm*4.5
			ang0 := evalVelocityAngle(x0, y0)
			adv := 8.5 * scale * math.Sin(phi-spatialPsi)
			sx += math.Cos(ang0) * adv
			sy += math.Sin(ang0) * adv
			flareBoost = 1.0 + 0.65*common.Smoothstep(0.70, 1.05, rNorm)
		}

		pts[0] = render.Pt(sx, sy)
		cx, cy := sx, sy
		strokeBend := (u3 - 0.5) * 0.065
		for k := 1; k <= traceSteps; k++ {
			ang := evalVelocityAngle(cx, cy) + strokeBend*float64(k)
			if isAnim {
				tStep := float64(k) / float64(traceSteps)
				whip := 0.30 * flareBoost * tStep * math.Sin(phi-1.9*tStep-spatialPsi)
				ang += whip
			}
			cx += math.Cos(ang) * stepLen
			cy += math.Sin(ang) * stepLen
			pts[k] = render.Pt(cx, cy)
		}

		midPt := pts[traceSteps/2]
		rootCol, dipolePot := evalSolarColor(x0, y0)
		lab := render.RGBAToOklab(rootCol)
		waveDL, waveDC := evalProminenceWave(midPt.X, midPt.Y, dipolePot)

		dL := -0.09 + 0.22*rng.Float64() + waveDL
		dA := (rng.Float64()-0.5)*0.032 + waveDC
		dB := (rng.Float64()-0.5)*0.032 - waveDC*0.5

		litCol := render.OklabToRGBA(render.Oklab{
			L: common.Constrain(lab.L+dL, 0.04, 0.98),
			A: lab.A + dA,
			B: lab.B + dB,
		}, 255)

		flankCol := render.OklabToRGBA(render.Oklab{
			L: common.Constrain(lab.L*0.58, 0.02, 0.84),
			A: lab.A * 0.92,
			B: lab.B * 0.92,
		}, 255)

		coreHalfW := (0.36 + 0.36*rng.Float64()) * scale
		flankHalfW := coreHalfW * 1.32

		if flankPath := buildBristleRibbonPath(pts[:], flankHalfW, 0.32*scale, 0.40*scale); flankPath != nil {
			rc.RasterizePath(flankPath, flankCol, 0.74)
		}
		if corePath := buildBristleRibbonPath(pts[:], coreHalfW, 0, 0); corePath != nil {
			rc.RasterizePath(corePath, litCol, 0.90)
		}
	}

	// 4. Orbiting White-Gold Coronal Spark Pinpoints around Sunspot Rosettes & Rim Loops
	for i := range poles {
		p := &poles[i]
		if !p.hasSparkle {
			continue
		}
		nSparks := 2
		if p.isRim {
			nSparks = 3
		}
		for k := 0; k < nSparks; k++ {
			baseAng := rng.Float64() * 2.0 * math.Pi
			orbitR := (2.2 + 7.5*rng.Float64()) * scale
			rad := (1.3 + 1.1*rng.Float64()) * scale

			ang := baseAng
			if isAnim {
				ang += p.spin * phi
			}
			px := p.x + math.Cos(ang)*orbitR
			py := p.y + math.Sin(ang)*orbitR

			// Dark obsidian-plum crescent shadow
			cAng := ang + p.spin*math.Pi*0.5
			rc.DrawGlowDot(px+math.Cos(cAng)*0.85*scale, py+math.Sin(cAng)*0.85*scale, rad*1.35, obsidianPlum, 0.80, 0.84)

			// Incandescent solar-gold & white-gold coronal pinpoint
			sparkHalo := solarGold
			if p.isRim && k == 0 {
				sparkHalo = turquoiseTeal
			}
			rc.DrawGlowDot(px, py, rad*1.15, sparkHalo, 0.92, 0.85)
			rc.DrawGlowDot(px-0.22*scale, py-0.22*scale, rad*0.62, whiteGold, 0.97, 0.94)
		}
	}

	c.MergeRenderCanvas(rc)
}
