package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type suminagashi struct {
	dropNum, tineCount int
	swirlStrength      float64
}

// NewSuminagashi returns a suminagashi engine that synthesizes Japanese Suminagashi / Ebru
// floating-ink marbling using exact closed-form area-preserving fluid operations evaluated
// via inverse backward mapping with 2x2 sub-pixel supersampling.
func NewSuminagashi(dropNum, tineCount int, swirlStrength float64) *suminagashi {
	if dropNum <= 0 {
		dropNum = 48
	}
	if tineCount <= 0 {
		tineCount = 7
	}
	if swirlStrength <= 0 {
		swirlStrength = 1.0
	}
	return &suminagashi{
		dropNum:       dropNum,
		tineCount:     tineCount,
		swirlStrength: swirlStrength,
	}
}

type marblingDrop struct {
	cx, cy     float64
	r, rSq     float64
	nRings     float64
	colorIndex int
	phase      float64
	hasGold    bool
}

type marblingRake struct {
	p0x, p0y   float64
	dx, dy     float64 // unit stroke direction D
	nx, ny     float64 // unit perpendicular N = (-dy, dx)
	amp        float64
	invWidth   float64
	freq       float64
	phase      float64
}

type marblingSwirl struct {
	vx, vy   float64
	invRvSq  float64
	strength float64
}

// Generative renders Suminagashi floating-ink marbling onto c.
func (sm *suminagashi) Generative(c *generativeart.Canva) {
	dropNum := sm.dropNum
	tineCount := sm.tineCount
	swirlStrength := sm.swirlStrength
	if dropNum <= 0 {
		dropNum = 48
	}
	if tineCount <= 0 {
		tineCount = 7
	}
	if swirlStrength <= 0 {
		swirlStrength = 1.0
	}

	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	minDim := math.Min(fw, fh)

	schema := c.Opts().ColorSchema()
	if len(schema) == 0 {
		schema = common.UkiyoE
	}
	grad := render.NewMultiStopGradient(schema...)
	dunhuangGrad := render.NewMultiStopGradient(common.Dunhuang...)

	// Precompute Oklab palette colors in linear RGB for fast sub-pixel shading
	const numPaletteStops = 16
	type linRGB struct {
		r, g, b float64
	}
	pigmentTable := make([]linRGB, numPaletteStops)
	darkRimTable := make([]linRGB, numPaletteStops)

	sumiDark := color.RGBA{R: 18, G: 26, B: 38, A: 255}
	xuanCream := color.RGBA{R: 246, G: 239, B: 225, A: 255}
	kintsugiGold := color.RGBA{R: 238, G: 196, B: 104, A: 255}

	creamR, creamG, creamB, _ := render.ColorToLinearRGB(xuanCream)
	sumiR, sumiG, sumiB, _ := render.ColorToLinearRGB(sumiDark)
	goldR, goldG, goldB, _ := render.ColorToLinearRGB(kintsugiGold)

	for i := 0; i < numPaletteStops; i++ {
		t := float64(i) / float64(numPaletteStops-1)
		col := render.LerpOklab(grad.SampleOklab(t), dunhuangGrad.SampleOklab(1.0-t), 0.24)
		if i%4 == 0 {
			// Periodically interleave deep sumi-ink indigo/charcoal rings
			col = render.LerpOklab(col, sumiDark, 0.72)
		}
		rL, gL, bL, _ := render.ColorToLinearRGB(col)
		pigmentTable[i] = linRGB{r: rL, g: gL, b: bL}

		rimCol := render.LerpOklab(col, sumiDark, 0.65)
		rrL, rgL, rbL, _ := render.ColorToLinearRGB(rimCol)
		darkRimTable[i] = linRGB{r: rrL, g: rgL, b: rbL}
	}

	rng := rand.New(rand.NewSource(202605))

	// 1. Construct concentric multi-ring ink drops using phyllotaxis + jittered organic centers
	drops := make([]marblingDrop, dropNum)
	goldenAngle := math.Pi * (3.0 - math.Sqrt(5.0))
	for k := 0; k < dropNum; k++ {
		u := (float64(k) + 0.5) / float64(dropNum)
		radFrac := 0.44 * math.Sqrt(u)
		theta := float64(k)*goldenAngle + (rng.Float64()-0.5)*0.45

		cx := fw * (0.5 + radFrac*math.Cos(theta) + (rng.Float64()-0.5)*0.12)
		cy := fh * (0.5 + radFrac*math.Sin(theta) + (rng.Float64()-0.5)*0.12)

		// Earlier drops are larger to cover the entire bath; later drops form delicate focal eyes
		sizeDecay := 1.0 - 0.52*math.Pow(u, 0.85)
		r := minDim * (0.14 + 0.12*rng.Float64()) * sizeDecay
		nRings := float64(5 + (k*3)%6)

		drops[k] = marblingDrop{
			cx:         cx,
			cy:         cy,
			r:          r,
			rSq:        r * r,
			nRings:     nRings,
			colorIndex: (k * 3) % numPaletteStops,
			phase:      rng.Float64() * 2.0 * math.Pi,
			hasGold:    k%3 == 0,
		}
	}

	// 2. Construct area-preserving comb-rake / tine strokes
	rakes := make([]marblingRake, tineCount)
	for i := 0; i < tineCount; i++ {
		t := float64(i) / float64(common.MaxInt(tineCount-1, 1))
		// Alternate between sweeping horizontal-diagonal and vertical-diagonal comb passes
		baseAng := 0.22*math.Pi + float64(i%2)*0.48*math.Pi + (rng.Float64()-0.5)*0.38
		dx := math.Cos(baseAng)
		dy := math.Sin(baseAng)
		nx := -dy
		ny := dx

		p0x := fw * (0.18 + 0.64*t + (rng.Float64()-0.5)*0.08)
		p0y := fh * (0.18 + 0.64*(1.0-t) + (rng.Float64()-0.5)*0.08)
		amp := minDim * (0.095 + 0.055*math.Sin(float64(i)*1.7)) * (0.75 + 0.25*swirlStrength)
		width := minDim * (0.38 + 0.18*rng.Float64())
		// Comb tine spacing across perpendicular axis N
		tineSpacing := minDim / (5.5 + float64(i%3)*2.0)
		freq := 2.0 * math.Pi / tineSpacing

		rakes[i] = marblingRake{
			p0x:      p0x,
			p0y:      p0y,
			dx:       dx,
			dy:       dy,
			nx:       nx,
			ny:       ny,
			amp:      amp,
			invWidth: 1.0 / math.Max(1.0, width),
			freq:     freq,
			phase:    float64(i)*1.15 + rng.Float64()*math.Pi,
		}
	}

	// 3. Construct area-preserving Gaussian vortex swirls
	const numSwirls = 5
	swirls := make([]marblingSwirl, numSwirls)
	for i := 0; i < numSwirls; i++ {
		ang := float64(i)*(2.0*math.Pi/float64(numSwirls)) + 0.35
		dist := minDim * (0.18 + 0.14*float64(i%2))
		vx := fw*0.5 + dist*math.Cos(ang)
		vy := fh*0.5 + dist*math.Sin(ang)
		rv := minDim * (0.24 + 0.07*float64(i%3))
		sign := 1.0
		if i%2 == 1 {
			sign = -1.0
		}
		swirls[i] = marblingSwirl{
			vx:       vx,
			vy:       vy,
			invRvSq:  1.0 / (rv * rv),
			strength: sign * swirlStrength * (1.85 + 0.35*float64(i%3)),
		}
	}

	// evalSample traces sub-pixel (px, py) backward through swirls, comb-rakes, and ink drops.
	evalSample := func(px, py float64) (rOut, gOut, bOut float64) {
		x, y := px, py

		// Step A: Reverse of vortex swirls in reverse order
		for i := len(swirls) - 1; i >= 0; i-- {
			sw := swirls[i]
			rx := x - sw.vx
			ry := y - sw.vy
			rSq := rx*rx + ry*ry
			if rSq < 9.0/sw.invRvSq {
				ang := -sw.strength * math.Exp(-rSq*sw.invRvSq)
				sinA, cosA := math.Sincos(ang)
				x = sw.vx + rx*cosA - ry*sinA
				y = sw.vy + rx*sinA + ry*cosA
			}
		}

		// Step B: Reverse of comb-rake / tine strokes in reverse order
		for i := len(rakes) - 1; i >= 0; i-- {
			rk := rakes[i]
			relX := x - rk.p0x
			relY := y - rk.p0y
			dotN := relX*rk.nx + relY*rk.ny
			env := math.Exp(-math.Abs(dotN) * rk.invWidth)
			shift := rk.amp * env * math.Sin(rk.freq*dotN+rk.phase)
			x -= rk.dx * shift
			y -= rk.dy * shift
		}

		// Step C: Reverse of concentric ink drops in reverse order
		for k := len(drops) - 1; k >= 0; k-- {
			dr := drops[k]
			dx := x - dr.cx
			dy := y - dr.cy
			dSq := dx*dx + dy*dy
			if dSq < dr.rSq {
				d := math.Sqrt(dSq)
				u := d / dr.r
				theta := math.Atan2(dy, dx)

				// Subtle organic radial feathering inside the concentric drop target
				feather := 0.035 * math.Sin(6.0*theta+dr.phase) * (1.0 - u) * u
				uMod := common.Clamp01(u + feather)

				ringPos := uMod * dr.nRings
				ringIdx := int(math.Floor(ringPos))
				ringFrac := ringPos - float64(ringIdx)

				// Golden kintsugi vein along outer drop boundary or primary inner ring seam
				if dr.hasGold {
					outerDist := math.Abs(uMod - 0.965)
					innerSeam := math.Abs(ringFrac - 0.5)
					if outerDist < 0.022 || (ringIdx == 2 && innerSeam < 0.045) {
						var goldW float64
						if outerDist < 0.022 {
							goldW = 1.0 - outerDist/0.022
						} else {
							goldW = 1.0 - innerSeam/0.045
						}
						goldW = goldW * goldW * (3.0 - 2.0*goldW)
						pCol := pigmentTable[(dr.colorIndex+ringIdx)%numPaletteStops]
						return pCol.r*(1.0-goldW) + goldR*goldW*1.12,
							pCol.g*(1.0-goldW) + goldG*goldW*1.12,
							pCol.b*(1.0-goldW) + goldB*goldW*1.12
					}
				}

				if ringIdx%2 == 0 {
					// Pigment ring with concentrated sumi-ink meniscus at ring edges
					cIdx := (dr.colorIndex + ringIdx/2) % numPaletteStops
					pCol := pigmentTable[cIdx]
					rimCol := darkRimTable[cIdx]

					edgeDist := math.Min(ringFrac, 1.0-ringFrac) * 2.0 // 0 at boundary, 1 at band center
					centerProfile := common.Smoothstep(0.0, 0.65, edgeDist)
					rOut = rimCol.r*(1.0-centerProfile) + pCol.r*centerProfile
					gOut = rimCol.g*(1.0-centerProfile) + pCol.g*centerProfile
					bOut = rimCol.b*(1.0-centerProfile) + pCol.b*centerProfile
					return rOut, gOut, bOut
				}

				// Alternating Xuan-paper cream surfactant band with delicate pigment halo at edges
				cIdx := (dr.colorIndex + ringIdx/2) % numPaletteStops
				pCol := pigmentTable[cIdx]
				edgeDist := math.Min(ringFrac, 1.0-ringFrac) * 2.0
				creamWeight := 0.78 + 0.18*common.Smoothstep(0.0, 0.5, edgeDist)
				rOut = pCol.r*(1.0-creamWeight) + creamR*creamWeight
				gOut = pCol.g*(1.0-creamWeight) + creamG*creamWeight
				bOut = pCol.b*(1.0-creamWeight) + creamB*creamWeight
				return rOut, gOut, bOut
			}

			// Outside drop #k: invert area-preserving radial displacement
			scale := math.Sqrt(1.0 - dr.rSq/dSq)
			x = dr.cx + dx*scale
			y = dr.cy + dy*scale
		}

		// Step D: Undeformed background water bath ripples (Xuan paper + faint sumi ink veins)
		rx := (x - fw*0.5) / minDim
		ry := (y - fh*0.5) / minDim
		rad := math.Hypot(rx, ry)
		wave := 0.5 + 0.5*math.Sin(rad*28.0*math.Pi+2.5*math.Atan2(ry, rx))
		vein := math.Pow(wave, 8.0)

		kintWave := math.Abs(math.Sin(rad*14.0*math.Pi - 1.5))
		if kintWave < 0.035 {
			gw := 1.0 - kintWave/0.035
			return creamR*(1.0-gw) + goldR*gw,
				creamG*(1.0-gw) + goldG*gw,
				creamB*(1.0-gw) + goldB*gw
		}

		tBg := 0.14 * vein
		return creamR*(1.0-tBg) + sumiR*tBg,
			creamG*(1.0-tBg) + sumiG*tBg,
			creamB*(1.0-tBg) + sumiB*tBg
	}

	rc := c.NewRenderCanvas()

	// 4. Evaluate backward mapping across every pixel with 2x2 Rotated-Grid / Sub-Pixel Supersampling
	subOffsets := [4][2]float64{
		{0.25, 0.25},
		{0.75, 0.25},
		{0.25, 0.75},
		{0.75, 0.75},
	}

	for py := 0; py < h; py++ {
		fpy := float64(py)
		for px := 0; px < w; px++ {
			fpx := float64(px)
			var accR, accG, accB float64
			for s := 0; s < 4; s++ {
				sr, sg, sb := evalSample(fpx+subOffsets[s][0], fpy+subOffsets[s][1])
				accR += sr
				accG += sg
				accB += sb
			}
			rc.SetPixelLinear(px, py, accR*0.25, accG*0.25, accB*0.25, 1.0)
		}
	}

	rc.ApplyPaperTexture(0.38, 2026)
	rc.ApplyVignette(0.2)
	c.MergeRenderCanvas(rc)
}
