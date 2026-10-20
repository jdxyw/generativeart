package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type mosaicSun struct {
	ringSteps  int
	brickScale float64
}

// NewMosaicSun returns a mosaicSun engine that renders a flow-aligned anisotropic
// faceted mosaic sun over swirling turquoise, cobalt, and magenta waves.
func NewMosaicSun(ringSteps int, brickScale float64) *mosaicSun {
	if ringSteps <= 0 {
		ringSteps = 32
	}
	if brickScale <= 0 {
		brickScale = 1.0
	}
	return &mosaicSun{
		ringSteps:  ringSteps,
		brickScale: brickScale,
	}
}

// Generative renders the bold, flat-shaded polygonal/diamond mosaic refraction of a glowing sun and waves.
func (ms *mosaicSun) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := (math.Min(w, h) / 600.0) * ms.brickScale
	if scale <= 0 {
		scale = 1.0
	}

	// Crisp poster-like palette
	cadmiumOrange := color.RGBA{R: 232, G: 126, B: 44, A: 255}
	goldenAmber := color.RGBA{R: 238, G: 174, B: 52, A: 255}
	sunYellow := color.RGBA{R: 248, G: 210, B: 82, A: 255}
	crimsonTerracotta := color.RGBA{R: 188, G: 66, B: 38, A: 255}
	espressoDark := color.RGBA{R: 48, G: 30, B: 32, A: 255}

	mintAqua := color.RGBA{R: 154, G: 212, B: 188, A: 255}
	paleIce := color.RGBA{R: 214, G: 240, B: 226, A: 255}
	seafoamTeal := color.RGBA{R: 88, G: 172, B: 168, A: 255}
	deepNavy := color.RGBA{R: 22, G: 26, B: 48, A: 255}

	cobaltBlue := color.RGBA{R: 42, G: 112, B: 168, A: 255}
	deepIndigo := color.RGBA{R: 28, G: 38, B: 82, A: 255}
	vibrantMagenta := color.RGBA{R: 212, G: 28, B: 118, A: 255}
	deepPlum := color.RGBA{R: 118, G: 22, B: 74, A: 255}

	// Fill deep grout backdrop so micro-gaps between faceted bricks enhance the crisp stained-mosaic look
	rc.FillBackground(deepNavy)
	rc.SetBlendMode(render.BlendOver)

	sunCX := 0.48 * w
	sunCY := 0.28 * h
	rSun := 0.27 * w

	rng := rand.New(rand.NewSource(202606))

	// Quantized color field sampler for outside the sun core (sky halo + tangential sunbeams + rolling waves)
	sampleSkyColor := func(x, y, jitter float64) color.RGBA {
		dx := x - sunCX
		dy := y - sunCY
		r := math.Hypot(dx, dy)
		ang := math.Atan2(dy, dx)

		// Normalize radial distance outside sun
		u := (r - rSun) / math.Max(1.0, 0.55*w)
		// Diagonal/swirling wave modulation
		diag := ((x - sunCX) * 0.65 - (y - sunCY) * 0.76) / math.Max(1.0, rSun)
		swirl := math.Sin(u*14.0 - ang*2.0 + 0.6*math.Sin(diag*2.2))

		// Warm tangential sunbeam streaks wrapping around the sun halo
		beamWave := math.Cos(ang*5.0 + u*6.5)
		if u < 0.42 && beamWave > 0.48+0.35*u {
			if jitter > 0.55 {
				return goldenAmber
			}
			return cadmiumOrange
		}

		// occasional dark navy structural accent brick
		if jitter < 0.085 || (swirl < -0.72 && jitter < 0.35) {
			return deepNavy
		}
		if jitter > 0.93 && u < 0.65 {
			return goldenAmber
		}

		// Quantized concentric aqua / seafoam / cobalt bands
		val := swirl + (jitter-0.5)*0.45
		switch {
		case val > 0.52:
			return paleIce
		case val > 0.08:
			return mintAqua
		case val > -0.42:
			return seafoamTeal
		default:
			return cobaltBlue
		}
	}

	sampleWaveColor := func(x, y, bandIdxFrac, jitter float64) color.RGBA {
		nx := x / math.Max(1.0, w)
		ny := (y - 0.58*h) / math.Max(1.0, 0.42*h)

		// Lower-left vibrant magenta/fuchsia wave crest region
		magentaMask := (1.0 - common.Constrain(nx*1.55, 0, 1)) * common.Constrain(ny*1.35, 0, 1)
		wavePhase := math.Sin(nx*9.5 + bandIdxFrac*11.0 + 0.8*math.Cos(nx*4.2))

		if magentaMask > 0.32 && (wavePhase > -0.15 || jitter < magentaMask*0.75) {
			if jitter < 0.18 {
				return deepPlum
			}
			if jitter > 0.86 {
				return cadmiumOrange
			}
			return vibrantMagenta
		}

		// Warm reflected sun glint on wave crests near center
		if math.Abs(nx-0.50) < 0.20 && wavePhase > 0.58 && jitter > 0.55 {
			return goldenAmber
		}

		if jitter < 0.12 {
			return deepNavy
		}

		switch {
		case wavePhase > 0.45:
			return mintAqua
		case wavePhase > 0.02:
			return seafoamTeal
		case wavePhase > -0.48:
			return cobaltBlue
		default:
			return deepIndigo
		}
	}

	// Helper to draw a crisp sheared parallelogram / tilted mosaic brick along a flow curve
	drawFlowBrick := func(p0, p1 render.Point, halfThick, shear float64, col color.RGBA) {
		dir := p1.Sub(p0)
		length := dir.Len()
		if length < 1e-4 {
			return
		}
		tVec := dir.Mul(1.0 / length)
		nVec := tVec.Perp()

		// Slight inset for crisp graphic facet separation
		insetL := math.Min(0.45*scale, length*0.04)
		insetT := math.Min(0.35*scale, halfThick*0.06)
		a := p0.Add(tVec.Mul(insetL))
		b := p1.Sub(tVec.Mul(insetL))
		ht := math.Max(1.2, halfThick-insetT)

		shearVec := tVec.Mul(shear * ht)
		v0 := a.Add(nVec.Mul(ht)).Add(shearVec)
		v1 := b.Add(nVec.Mul(ht)).Add(shearVec)
		v2 := b.Sub(nVec.Mul(ht)).Sub(shearVec)
		v3 := a.Sub(nVec.Mul(ht)).Sub(shearVec)

		rc.FillPolygon([]render.Point{v0, v1, v2, v3}, col, 1.0)
	}

	// 1. Concentric & Swirling Anisotropic Mosaic Rings Outside the Sun Core
	totalRings := ms.ringSteps
	if totalRings < 16 {
		totalRings = 32
	}
	brickThick := 9.5 * scale
	if ms.ringSteps != 32 {
		brickThick = (300.0 / float64(totalRings)) * scale
	}

	maxRadius := math.Hypot(w, h) * 0.95
	for ring := 0; ring < totalRings+14; ring++ {
		rBase := rSun + (float64(ring)+0.5)*brickThick
		if rBase-brickThick > maxRadius {
			break
		}
		circum := 2.0 * math.Pi * rBase
		brickLen := common.Constrain((21.0+4.5*math.Sin(float64(ring)*0.7))*scale, 16.0*scale, 34.0*scale)
		numBricks := int(math.Max(18, math.Round(circum/brickLen)))
		dTheta := (2.0 * math.Pi) / float64(numBricks)
		ringPhase := float64(ring) * 0.37

		for k := 0; k < numBricks; k++ {
			th0 := float64(k)*dTheta + ringPhase
			th1 := th0 + dTheta*0.96
			thMid := 0.5 * (th0 + th1)

			// Anisotropic radial flow warp so rings undulate like refractive sky currents
			warp0 := 8.5 * scale * math.Sin(3.0*th0+float64(ring)*0.32)
			warp1 := 8.5 * scale * math.Sin(3.0*th1+float64(ring)*0.32)
			r0 := rBase + warp0
			r1 := rBase + warp1

			p0 := render.Pt(sunCX+r0*math.Cos(th0), sunCY+r0*math.Sin(th0))
			p1 := render.Pt(sunCX+r1*math.Cos(th1), sunCY+r1*math.Sin(th1))
			mid := p0.Lerp(p1, 0.5)

			// Skip bricks well outside canvas bounds
			if mid.X < -40 || mid.X > w+40 || mid.Y < -40 || mid.Y > h+40 {
				continue
			}

			// In the lower sea region, let the rolling wave bands take over cleanly
			waveHorizon := 0.60*h + 24.0*scale*math.Sin(mid.X/(0.18*w)+0.4) - 18.0*scale*math.Cos(mid.X/(0.31*w))
			if mid.Y > waveHorizon+brickThick*0.6 {
				continue
			}

			jitter := rng.Float64()
			col := sampleSkyColor(mid.X, mid.Y, jitter)
			shear := 0.32 * math.Cos(thMid*2.0+float64(ring)*0.25)
			drawFlowBrick(p0, p1, brickThick*0.48, shear, col)
		}
	}

	// 2. Lower Region (y > 0.60 * h): Rolling Wave Hills of Cobalt, Deep Indigo, and Vibrant Magenta
	waveBands := int(math.Ceil((0.46 * h) / (brickThick * 0.92))) + 4
	for wb := 0; wb < waveBands; wb++ {
		bandFrac := float64(wb) / math.Max(1.0, float64(waveBands-1))
		baseY := 0.58*h + float64(wb)*brickThick*0.90

		// Trace flow-aligned wave streamline across the width of the canvas
		brickLen := (22.0 + 6.0*math.Sin(float64(wb)*0.6)) * scale
		stepsX := int(math.Ceil((w + 60.0*scale) / brickLen))
		dx := (w + 60.0*scale) / float64(stepsX)
		xOffset := -30.0*scale + float64(wb%2)*brickLen*0.45

		waveEvalY := func(x float64) float64 {
			nx := x / math.Max(1.0, w)
			crest1 := 28.0 * scale * math.Sin(nx*2.0*math.Pi*1.35+float64(wb)*0.22)
			crest2 := 16.0 * scale * math.Cos(nx*2.0*math.Pi*2.4-float64(wb)*0.35+0.8)
			swell := -22.0 * scale * math.Exp(-math.Pow((nx-0.24)/0.22, 2.0)) * math.Sin(bandFrac*math.Pi)
			return baseY + crest1 + crest2 + swell
		}

		for sx := 0; sx < stepsX; sx++ {
			x0 := xOffset + float64(sx)*dx
			x1 := x0 + dx*0.96
			p0 := render.Pt(x0, waveEvalY(x0))
			p1 := render.Pt(x1, waveEvalY(x1))
			mid := p0.Lerp(p1, 0.5)
			if mid.Y < 0.52*h || mid.Y > h+30*scale {
				continue
			}

			jitter := rng.Float64()
			col := sampleWaveColor(mid.X, mid.Y, bandFrac, jitter)
			shear := 0.38
			if wb%2 == 1 {
				shear = -0.28
			}
			drawFlowBrick(p0, p1, brickThick*0.47, shear, col)
		}
	}

	// 3. Inside the Sun Core: Bold Low-Poly / Sheared Polygonal & Diamond Shards
	// First lay down a dark espresso rim backing around R_sun
	sunRimPts := make([]render.Point, 36)
	for i := 0; i < 36; i++ {
		ang := (float64(i) / 36.0) * 2.0 * math.Pi
		sunRimPts[i] = render.Pt(sunCX+rSun*1.01*math.Cos(ang), sunCY+rSun*1.01*math.Sin(ang))
	}
	rc.FillPolygon(sunRimPts, espressoDark, 1.0)

	// Construct sheared diagonal-concentric low-poly grid clipped to the sun disc (35-45 bold angular facets)
	// Using tilted u-v axes aligned at -32 degrees (sweeping diagonally across the sun disc)
	diagAngle := -0.56
	uAxis := render.Pt(math.Cos(diagAngle), math.Sin(diagAngle))
	vAxis := uAxis.Perp()

	const uSlices = 7
	const vSlices = 7
	grid := make([][]render.Point, uSlices+1)
	for iu := 0; iu <= uSlices; iu++ {
		grid[iu] = make([]render.Point, vSlices+1)
		uNorm := (float64(iu)/float64(uSlices))*2.0 - 1.0
		for iv := 0; iv <= vSlices; iv++ {
			vNorm := (float64(iv)/float64(vSlices))*2.0 - 1.0

			// Diagonal shear and organic vertex jitter for bold low-poly crystal facets
			jx := 0.0
			jy := 0.0
			if iu > 0 && iu < uSlices && iv > 0 && iv < vSlices {
				jx = (rng.Float64() - 0.5) * 0.24 * rSun
				jy = (rng.Float64() - 0.5) * 0.24 * rSun
			}
			shearU := uNorm + 0.22*math.Sin(vNorm*math.Pi*0.8)
			px := sunCX + (shearU*uAxis.X+vNorm*vAxis.X)*rSun*1.02 + jx
			py := sunCY + (shearU*uAxis.Y+vNorm*vAxis.Y)*rSun*1.02 + jy

			// Clamp outer vertices onto the sun disc circumference so facets form a crisp faceted circle
			dx := px - sunCX
			dy := py - sunCY
			dist := math.Hypot(dx, dy)
			if dist > rSun*0.985 || iu == 0 || iu == uSlices || iv == 0 || iv == vSlices {
				if dist > rSun*0.985 {
					px = sunCX + (dx/dist)*rSun*0.985
					py = sunCY + (dy/dist)*rSun*0.985
				}
			}
			grid[iu][iv] = render.Pt(px, py)
		}
	}

	sampleSunFacetColor := func(centroid render.Point, shardIdx int) color.RGBA {
		dx := (centroid.X - sunCX) / rSun
		dy := (centroid.Y - sunCY) / rSun
		rNorm := math.Hypot(dx, dy)
		ang := math.Atan2(dy, dx)

		// Dark espresso crescent shadow in the lower-right of the sun disc
		if rNorm > 0.54 && ang > -0.25 && ang < 1.85 {
			if rNorm > 0.68 || shardIdx%3 == 0 {
				return espressoDark
			}
			return crimsonTerracotta
		}

		// Diagonal-concentric warm bands across the sun disc
		diagBand := dx*0.72 + dy*0.68
		switch {
		case rNorm < 0.34 && diagBand < 0.12:
			return sunYellow
		case diagBand < -0.22:
			if shardIdx%2 == 0 {
				return goldenAmber
			}
			return sunYellow
		case diagBand < 0.18:
			if shardIdx%3 == 0 {
				return goldenAmber
			}
			return cadmiumOrange
		case diagBand < 0.52:
			if shardIdx%2 == 0 {
				return cadmiumOrange
			}
			return crimsonTerracotta
		default:
			return crimsonTerracotta
		}
	}

	insetPolygon := func(pts []render.Point, shrink float64) []render.Point {
		var cx, cy float64
		for _, p := range pts {
			cx += p.X
			cy += p.Y
		}
		center := render.Pt(cx/float64(len(pts)), cy/float64(len(pts)))
		out := make([]render.Point, len(pts))
		for i, p := range pts {
			d := p.Sub(center)
			l := d.Len()
			if l > shrink*2.5 {
				out[i] = p.Sub(d.Mul(shrink / l))
			} else {
				out[i] = p.Lerp(center, 0.08)
			}
		}
		return out
	}

	shardIdx := 0
	for iu := 0; iu < uSlices; iu++ {
		for iv := 0; iv < vSlices; iv++ {
			p00 := grid[iu][iv]
			p10 := grid[iu+1][iv]
			p11 := grid[iu+1][iv+1]
			p01 := grid[iu][iv+1]

			centroid := render.Pt(
				0.25*(p00.X+p10.X+p11.X+p01.X),
				0.25*(p00.Y+p10.Y+p11.Y+p01.Y),
			)
			if math.Hypot(centroid.X-sunCX, centroid.Y-sunCY) > rSun*0.97 {
				continue
			}

			// Split some cells into bold triangles and keep others as sheared diamond quads
			if (iu+iv)%2 == 0 {
				quad := insetPolygon([]render.Point{p00, p10, p11, p01}, 0.55*scale)
				col := sampleSunFacetColor(centroid, shardIdx)
				rc.FillPolygon(quad, col, 1.0)
				shardIdx++
			} else {
				tri1 := insetPolygon([]render.Point{p00, p10, p11}, 0.55*scale)
				c1 := render.Pt((p00.X+p10.X+p11.X)/3.0, (p00.Y+p10.Y+p11.Y)/3.0)
				rc.FillPolygon(tri1, sampleSunFacetColor(c1, shardIdx), 1.0)
				shardIdx++

				tri2 := insetPolygon([]render.Point{p00, p11, p01}, 0.55*scale)
				c2 := render.Pt((p00.X+p11.X+p01.X)/3.0, (p00.Y+p11.Y+p01.Y)/3.0)
				rc.FillPolygon(tri2, sampleSunFacetColor(c2, shardIdx), 1.0)
				shardIdx++
			}
		}
	}

	c.MergeRenderCanvas(rc)
}
