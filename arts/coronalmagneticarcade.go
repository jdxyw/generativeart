package arts

import (
	"image/color"
	"math"
	"math/rand"
	"sort"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type coronalMagneticArcade struct {
	dipoles     int
	strokeCount int
	flareTwist  float64
}

// NewCoronalMagneticArcade returns a coronalMagneticArcade engine that sculpts an asymmetric
// 3D G-Buffer Total Solar Eclipse limb, bipolar magnetic coronal arcades, and horizon-shadowed
// plasma prominence plumes erupting from an occulted obsidian lunar disc.
func NewCoronalMagneticArcade(dipoles, strokeCount int, flareTwist float64) *coronalMagneticArcade {
	if dipoles <= 0 {
		dipoles = 9
	}
	if strokeCount <= 0 {
		strokeCount = 19000
	}
	if flareTwist <= 0 {
		flareTwist = 1.15
	}
	return &coronalMagneticArcade{
		dipoles:     common.ConstrainInt(dipoles, 4, 18),
		strokeCount: common.ConstrainInt(strokeCount, 4000, 32000),
		flareTwist:  common.Constrain(flareTwist, 0.35, 2.40),
	}
}

type eclipseMagneticPole struct {
	x, y     float64
	radius   float64
	charge   float64
	spin     float64
	strength float64
	phaseOff float64
}

type coronalPlumeSeed struct {
	x, y       float64
	elev       float64
	thickness  float64
	cosView    float64
	limbDist   float64
	strokeSeed float64
	lenMod     float64
	radMod     float64
	bendMod    float64
}

// Generative renders the static CoronalMagneticArcade composition.
func (cma *coronalMagneticArcade) Generative(c *generativeart.Canva) {
	cma.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless looping CoronalMagneticArcade animation.
func (cma *coronalMagneticArcade) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
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
		seed = 20262604
	}
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed + 2604)

	isAnim := totalFrames > 1
	phi := 2.0 * math.Pi * phase
	loopX := 0.08 * math.Sin(phi)
	loopY := 0.08 * (1.0 - math.Cos(phi))

	// Off-center Occulted Lunar Disc at (0.38*fw, 0.58*fh) so massive coronal prominence loops
	// erupt asymmetrically toward the upper-right & center-right!
	moonX := 0.36 * fw
	moonY := 0.60 * fh
	moonR := 0.235 * minDim

	// Place bipolar magnetic active regions along the northeast/east solar limb and outer corona
	poles := make([]eclipseMagneticPole, 0, cma.dipoles)
	for i := 0; i < cma.dipoles; i++ {
		fi := float64(i)
		frac := fi / math.Max(1.0, float64(cma.dipoles-1))
		// Concentrate active regions along the NE/E/NW limb arc (-1.35 rad to +0.55 rad) + outer coronalhelmet streamers
		ang := -1.45 + frac*2.15 + (rng.Float64()-0.5)*0.18
		rMult := 1.04 + 0.12*float64(i%2)
		if i >= (cma.dipoles*2)/3 {
			// Outer coronal vortex anchors
			rMult = 1.55 + 0.45*rng.Float64()
			ang = -1.20 + (fi/float64(cma.dipoles))*1.65
		}
		px := moonX + math.Cos(ang)*moonR*rMult
		py := moonY + math.Sin(ang)*moonR*rMult
		charge := 1.0
		if i%2 == 1 {
			charge = -1.0
		}
		spin := charge * cma.flareTwist
		poles = append(poles, eclipseMagneticPole{
			x:        px,
			y:        py,
			radius:   (42.0 + 28.0*rng.Float64()) * scale,
			charge:   charge,
			spin:     spin,
			strength: 1.05 + 0.45*rng.Float64(),
			phaseOff: fi * 1.15,
		})
	}

	evalCoronalField := func(x, y float64) (zHeight, filmNm, archPhase, limbProximity float64) {
		nx, ny := x/fw, y/fh
		warpX := noise.FBM2D(nx*2.8+1.1+loopX, ny*2.8-0.7+loopY, 3, 2.0, 0.5)
		warpY := noise.FBM2D(nx*2.8-2.1-loopX, ny*2.8+1.6-loopY, 3, 2.0, 0.5)

		dxM := x - moonX
		dyM := y - moonY
		rMoon := math.Hypot(dxM, dyM)
		limbRatio := rMoon / moonR
		limbProximity = math.Exp(-math.Max(0.0, limbRatio-0.95) * 1.35)

		angM := math.Atan2(dyM, dxM)
		archWave := math.Sin(angM*5.0 + warpX*2.2 - phi)
		archPhase = limbRatio*6.2 + archWave*1.4
		zHeight = (0.85 + 2.85*(0.5+0.5*math.Cos(archPhase)) + 1.85*limbProximity)
		filmNm = 340.0 + 210.0*(0.5+0.5*math.Sin(archPhase*0.8+warpY))
		return
	}

	gb := render.NewFiberGBuffer(w, h, color.RGBA{R: 5, G: 7, B: 16, A: 255})
	obsidianSpace := color.RGBA{R: 5, G: 7, B: 16, A: 255}
	coronalIndigo := color.RGBA{R: 16, G: 20, B: 56, A: 255}
	crimsonProminence := color.RGBA{R: 218, G: 42, B: 58, A: 255}
	solarOrange := color.RGBA{R: 248, G: 118, B: 28, A: 255}
	diamondGold := color.RGBA{R: 255, G: 214, B: 82, A: 255}
	plasmaWhite := color.RGBA{R: 255, G: 250, B: 235, A: 255}
	coronalCyan := color.RGBA{R: 48, G: 218, B: 232, A: 255}

	for y := 0; y < h; y++ {
		fy := float64(y) + 0.5
		rowOff := y * w
		for x := 0; x < w; x++ {
			fx := float64(x) + 0.5
			idx := rowOff + x

			zC, _, archPhase, limbProx := evalCoronalField(fx, fy)
			baseCol := render.LerpOklab(obsidianSpace, coronalIndigo, common.Clamp01(limbProx*0.85))
			baseCol = render.LerpOklab(baseCol, crimsonProminence, common.Clamp01(limbProx*(0.25+0.25*math.Sin(archPhase))))
			mR, mG, mB, _ := render.ColorToLinearRGB(baseCol)

			gb.Z[idx] = zC * 0.45
			gb.Nx[idx] = 0
			gb.Ny[idx] = 0
			gb.Nz[idx] = 1
			gb.AlbR[idx] = mR
			gb.AlbG[idx] = mG
			gb.AlbB[idx] = mB
			gb.Roughness[idx] = 0.35
		}
	}

	// Bipolar magnetic field + radial helmet streamer vector field outside the lunar disc
	evalMagneticAngle := func(x, y float64) float64 {
		nx, ny := x/fw, y/fh
		curl := noise.Curl2D(nx*2.2+0.5+loopX*0.6, ny*2.2-0.4+loopY*0.6, 0.01)
		vx := curl.X * 0.25
		vy := curl.Y * 0.25

		// Radial solar wind helmet streamer component streaming toward NE
		dxM := x - moonX
		dyM := y - moonY
		rM := math.Hypot(dxM, dyM) + 1.0
		vx += (dxM / rM) * 0.38
		vy += (dyM / rM) * 0.38

		for i := range poles {
			p := &poles[i]
			dx := x - p.x
			dy := y - p.y
			r2 := dx*dx + dy*dy + 180.0*scale*scale
			r := math.Sqrt(r2)
			wgt := p.strength * (p.radius * p.radius) / r2

			// Superpose radial dipole source/sink + helical flux-rope twist
			rx, ry := dx/r, dy/r
			tx, ty := -ry, rx
			vx += wgt * (p.charge*rx*0.62 + p.spin*tx*0.78)
			vy += wgt * (p.charge*ry*0.62 + p.spin*ty*0.78)
		}
		return math.Atan2(vy, vx)
	}

	scaledStrokes := int(math.Round(float64(cma.strokeCount) * scale))
	if scaledStrokes < 3000 {
		scaledStrokes = 3000
	}

	seeds := make([]coronalPlumeSeed, 0, scaledStrokes)
	goldenPhi2X := 0.7548776662466927
	goldenPhi2Y := 0.5698402909980532

	for i := 0; i < scaledStrokes; i++ {
		fi := float64(i) + 0.5
		var x0, y0 float64
		if i%5 < 2 {
			// Bias 40% of strokes into the erupting NE solar limb & active magnetic dipoles
			p := poles[i%len(poles)]
			r := p.radius * (0.25 + 1.35*rng.Float64())
			ang := rng.Float64() * 2.0 * math.Pi
			x0 = common.Constrain(p.x+math.Cos(ang)*r, 0.02*fw, 0.98*fw)
			y0 = common.Constrain(p.y+math.Sin(ang)*r, 0.02*fh, 0.98*fh)
		} else {
			u := math.Mod(fi*goldenPhi2X+0.13*rng.Float64(), 1.0)
			v := math.Mod(fi*goldenPhi2Y+0.13*rng.Float64(), 1.0)
			x0 = (0.02 + 0.96*u) * fw
			y0 = (0.02 + 0.96*v) * fh
		}

		// Keep seed roots outside moonR*1.06 so 7-step fibers arch outward from the chromosphere rim without entering the dark moon disc
		if math.Hypot(x0-moonX, y0-moonY) < moonR*1.06 {
			continue
		}

		zElev, filmNm, archPhase, limbProx := evalCoronalField(x0, y0)
		cosV := common.Constrain(0.35+0.55*(0.5+0.5*math.Cos(archPhase*0.7+limbProx*1.8)), 0.15, 0.98)

		seeds = append(seeds, coronalPlumeSeed{
			x:          x0,
			y:          y0,
			elev:       zElev + 1.65*limbProx,
			thickness:  filmNm,
			cosView:    cosV,
			limbDist:   limbProx,
			strokeSeed: rng.Float64() * 2.0 * math.Pi,
			lenMod:     0.82 + 0.58*rng.Float64(),
			radMod:     0.72 + 0.58*rng.Float64(),
			bendMod:    (rng.Float64() - 0.5) * 0.055,
		})
	}

	sort.Slice(seeds, func(i, j int) bool {
		return seeds[i].elev < seeds[j].elev
	})

	schema := c.Opts().ColorSchema()
	if len(schema) == 0 {
		schema = common.Inferno
	}
	palGrad := render.NewMultiStopGradient(schema...)

	const numCtrlPts = 7
	var ctrlPts [numCtrlPts]render.Point
	stepDist := 5.2 * scale

	for i := range seeds {
		sd := &seeds[i]
		cx, cy := sd.x, sd.y
		if isAnim {
			ang0 := evalMagneticAngle(cx, cy)
			adv := 4.4 * scale * math.Sin(phi-sd.strokeSeed)
			cx += math.Cos(ang0) * adv
			cy += math.Sin(ang0) * adv
		}

		ctrlPts[0] = render.Pt(cx, cy)
		segStep := stepDist * sd.lenMod
		for k := 1; k < numCtrlPts; k++ {
			tk := float64(k) / float64(numCtrlPts-1)
			ang := evalMagneticAngle(cx, cy) + sd.bendMod*float64(k)
			undul := 0.16 * math.Sin(tk*math.Pi*2.2+sd.strokeSeed-phi)
			ang += undul
			cx += math.Cos(ang) * segStep
			cy += math.Sin(ang) * segStep
			ctrlPts[k] = render.Pt(cx, cy)
		}

		bandPhase := common.Clamp01(0.25 + 0.55*sd.limbDist + 0.22*math.Sin(sd.elev*1.4))
		rootCol := render.LerpOklab(obsidianSpace, coronalIndigo, sd.limbDist)

		var midCol color.RGBA
		if bandPhase < 0.38 {
			midCol = render.LerpOklab(coronalIndigo, coronalCyan, bandPhase/0.38)
		} else if bandPhase < 0.72 {
			midCol = render.LerpOklab(crimsonProminence, solarOrange, (bandPhase-0.38)/0.34)
		} else {
			midCol = render.LerpOklab(solarOrange, diamondGold, (bandPhase-0.72)/0.28)
		}
		midCol = render.LerpOklab(midCol, palGrad.SampleOklab(bandPhase), 0.24)

		ridgeHighlight := common.Smoothstep(2.0, 4.5, sd.elev)
		tipCol := render.LerpOklab(midCol, plasmaWhite, 0.28+0.62*ridgeHighlight*sd.limbDist)

		zRoot := math.Max(0.15, sd.elev-0.35)
		zTip := sd.elev + 2.25 + 1.55*ridgeHighlight
		maxRadius := (0.72 + 0.48*sd.radMod) * scale
		roughness := 0.18 + 0.16*(1.0-ridgeHighlight)

		gb.RasterizeFiber3D(
			ctrlPts[:],
			zRoot,
			zTip,
			maxRadius,
			rootCol,
			midCol,
			tipCol,
			0.58,
			roughness,
		)
	}

	rc := c.NewRenderCanvas()
	rc.SetRecordVector(false)

	horizonRad := common.MaxInt(4, int(math.Round(16.0*scale)))
	matCfg := render.DefaultGBufferMaterialConfig([3]float64{-0.48, -0.58, 0.66})
	matCfg.HorizonShadowStrength = 0.72
	matCfg.HorizonRadius = horizonRad
	matCfg.SSSStrength = 0.42
	matCfg.SSSTint = solarOrange
	matCfg.IridescenceStrength = 0.36
	matCfg.FilmThicknessNm = 410.0
	matCfg.FilmIOR = 1.52

	gb.ResolveToCanvasAdvanced(rc, matCfg)

	// Carve the 100% opaque Occulted Velvet-Obsidian Lunar Disc & Blazing Diamond-Ring Chromosphere Rim!
	moonDiscPts := make([]render.Point, 96)
	for i := 0; i < 96; i++ {
		a := (float64(i) / 96.0) * 2.0 * math.Pi
		moonDiscPts[i] = render.Pt(moonX+moonR*1.01*math.Cos(a), moonY+moonR*1.01*math.Sin(a))
	}
	rc.FillPolygon(moonDiscPts, obsidianSpace, 1.0)
	rc.FillPolygon(moonDiscPts, obsidianSpace, 1.0)
	rc.DrawSoftRing(moonX, moonY, moonR*1.005, 4.2*scale, crimsonProminence, 0.85)
	rc.DrawSoftRing(moonX, moonY, moonR*1.012, 2.2*scale, diamondGold, 0.92)

	// Blazing Diamond-Ring Baily's Beads along the NE solar limb (-0.72 rad)
	rc.SetBlendMode(render.BlendAdd)
	diamondAng := -0.68
	dxRing := moonX + math.Cos(diamondAng)*moonR*1.01
	dyRing := moonY + math.Sin(diamondAng)*moonR*1.01
	rc.DrawGlowDot(dxRing, dyRing, 28.0*scale, solarOrange, 0.62, 0.25)
	rc.DrawGlowDot(dxRing, dyRing, 14.0*scale, diamondGold, 0.88, 0.55)
	rc.DrawGlowDot(dxRing, dyRing, 6.5*scale, plasmaWhite, 0.98, 0.85)

	for i := range poles {
		p := &poles[i]
		if math.Hypot(p.x-moonX, p.y-moonY) > moonR*1.22 {
			continue
		}
		rc.DrawGlowDot(p.x, p.y, 8.5*scale, solarOrange, 0.55, 0.25)
		rc.DrawGlowDot(p.x, p.y, 3.4*scale, plasmaWhite, 0.92, 0.78)
	}
	rc.SetBlendMode(render.BlendOver)

	rc.ApplyBloom(0.76, 4.2*scale, 0.20)
	rc.ApplyVignette(0.22)
	c.MergeRenderCanvas(rc)
}
