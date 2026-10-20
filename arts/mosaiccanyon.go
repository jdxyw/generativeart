package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type mosaicCanyon struct {
	strataSteps int
	brickScale  float64
}

// NewMosaicCanyon returns a mosaicCanyon engine that renders stratified red-rock
// canyon cliffs, a winding turquoise and gold river, and a blazing low-poly eclipse
// sun in the crisp MosaicSun aesthetic.
func NewMosaicCanyon(strataSteps int, brickScale float64) *mosaicCanyon {
	if strataSteps <= 0 {
		strataSteps = 32
	}
	if brickScale <= 0 {
		brickScale = 1.0
	}
	return &mosaicCanyon{
		strataSteps: strataSteps,
		brickScale:  brickScale,
	}
}

// Generative renders the stratified canyon walls, winding turquoise river, and eclipse sky mosaic.
func (mc *mosaicCanyon) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := (math.Min(w, h) / 600.0) * mc.brickScale
	if scale <= 0 {
		scale = 1.0
	}

	// Curated Desert Canyon, Turquoise River & Eclipse palette
	deepObsidian := color.RGBA{R: 22, G: 18, B: 32, A: 255}
	espressoDark := color.RGBA{R: 46, G: 26, B: 28, A: 255}
	terracottaRed := color.RGBA{R: 196, G: 68, B: 42, A: 255}
	burntSienna := color.RGBA{R: 158, G: 52, B: 38, A: 255}
	cadmiumOrange := color.RGBA{R: 234, G: 124, B: 44, A: 255}
	goldenOchre := color.RGBA{R: 242, G: 178, B: 56, A: 255}
	sunYellow := color.RGBA{R: 250, G: 218, B: 88, A: 255}

	canyonPlum := color.RGBA{R: 98, G: 32, B: 68, A: 255}
	twilightViolet := color.RGBA{R: 62, G: 36, B: 92, A: 255}
	cobaltBlue := color.RGBA{R: 42, G: 108, B: 166, A: 255}
	seafoamTeal := color.RGBA{R: 78, G: 168, B: 164, A: 255}
	mintAqua := color.RGBA{R: 152, G: 216, B: 192, A: 255}
	paleIce := color.RGBA{R: 218, G: 242, B: 232, A: 255}

	rc.FillBackground(deepObsidian)
	rc.SetBlendMode(render.BlendOver)

	rng := rand.New(rand.NewSource(202609))

	sunCX, sunCY := 0.50*w, 0.24*h
	rSun := 0.175 * w

	brickThick := 9.2 * scale
	if mc.strataSteps != 32 {
		brickThick = (295.0 / float64(mc.strataSteps)) * scale
	}

	// Canyon Rim Profile: high mesas on left and right, dipping into a central V-gorge at x = 0.50w
	canyonRimY := func(x float64) float64 {
		nx := x / math.Max(1.0, w)
		distCenter := math.Abs(nx - 0.50)
		// Stepped mesa profile rising toward left and right edges
		mesa := 0.58*h - 0.26*h*math.Pow(common.Constrain(distCenter*2.0, 0, 1), 0.75)
		stepMod := 10.0 * scale * math.Sin(distCenter*24.0)
		return mesa + stepMod
	}

	// S-curving River Centerline & Width from the canyon gap (y = 0.56h) down to bottom (y = h)
	riverBounds := func(y float64) (float64, float64) {
		v := common.Constrain((y-0.55*h)/(0.45*h), 0, 1)
		centerX := 0.50*w + 0.14*w*math.Sin(v*math.Pi*1.55-0.25)*math.Pow(v, 0.65)
		halfW := (0.045 + 0.24*math.Pow(v, 1.15)) * w
		return centerX - halfW, centerX + halfW
	}

	// 1. Concentric Warped Eclipse Corona & Twilight Sky Courses (above canyonRimY)
	maxSkyR := math.Hypot(w, h) * 0.85
	for ring := 0; ring < 34; ring++ {
		rBase := rSun + (float64(ring)+0.5)*brickThick
		if rBase > maxSkyR {
			break
		}
		circum := 2.0 * math.Pi * rBase
		brickLen := common.Constrain((20.0+4.0*math.Sin(float64(ring)*0.6))*scale, 15.0*scale, 32.0*scale)
		numBricks := int(math.Max(18, math.Round(circum/brickLen)))
		dTheta := (2.0 * math.Pi) / float64(numBricks)
		ringPhase := float64(ring) * 0.39

		for k := 0; k < numBricks; k++ {
			th0 := float64(k)*dTheta + ringPhase
			th1 := th0 + dTheta*0.96
			thMid := 0.5 * (th0 + th1)

			warp0 := 7.5 * scale * math.Cos(3.0*th0-float64(ring)*0.28)
			warp1 := 7.5 * scale * math.Cos(3.0*th1-float64(ring)*0.28)
			p0 := render.Pt(sunCX+(rBase+warp0)*math.Cos(th0), sunCY+(rBase+warp0)*math.Sin(th0))
			p1 := render.Pt(sunCX+(rBase+warp1)*math.Cos(th1), sunCY+(rBase+warp1)*math.Sin(th1))
			mid := p0.Lerp(p1, 0.5)

			if mid.X < -30 || mid.X > w+30 || mid.Y < -30 || mid.Y > canyonRimY(mid.X)+brickThick*0.35 {
				continue
			}

			jitter := rng.Float64()
			u := float64(ring) / 18.0
			coronaFlare := math.Cos(thMid*4.0 + float64(ring)*0.3)

			var col color.RGBA
			switch {
			case ring < 2:
				if jitter > 0.35 {
					col = sunYellow
				} else {
					col = cadmiumOrange
				}
			case u < 0.42 && coronaFlare > 0.32:
				if jitter > 0.52 {
					col = goldenOchre
				} else {
					col = terracottaRed
				}
			case jitter < 0.08:
				col = deepObsidian
			default:
				val := math.Sin(u*8.0+thMid*2.0) + (jitter-0.5)*0.4
				switch {
				case val > 0.48:
					col = paleIce
				case val > 0.06:
					col = mintAqua
				case val > -0.38:
					col = seafoamTeal
				default:
					col = cobaltBlue
				}
			}

			shear := 0.30 * math.Sin(thMid*2.0+float64(ring)*0.25)
			rc.DrawMosaicBrick(p0, p1, brickThick*0.48, shear, scale, col, 1.0)
		}
	}

	// 2. Inside the Eclipse Sun Core: Low-Poly Blazing Corona Rim & Obsidian Eclipse Disc
	sunRim := make([]render.Point, 36)
	for i := 0; i < 36; i++ {
		ang := (float64(i) / 36.0) * 2.0 * math.Pi
		sunRim[i] = render.Pt(sunCX+rSun*1.01*math.Cos(ang), sunCY+rSun*1.01*math.Sin(ang))
	}
	rc.FillPolygon(sunRim, espressoDark, 1.0)

	const eSlices = 6
	eGrid := make([][]render.Point, eSlices+1)
	uAx := render.Pt(math.Cos(-0.42), math.Sin(-0.42))
	vAx := uAx.Perp()
	for iu := 0; iu <= eSlices; iu++ {
		eGrid[iu] = make([]render.Point, eSlices+1)
		uN := (float64(iu)/float64(eSlices))*2.0 - 1.0
		for iv := 0; iv <= eSlices; iv++ {
			vN := (float64(iv)/float64(eSlices))*2.0 - 1.0
			jx, jy := 0.0, 0.0
			if iu > 0 && iu < eSlices && iv > 0 && iv < eSlices {
				jx = (rng.Float64() - 0.5) * 0.22 * rSun
				jy = (rng.Float64() - 0.5) * 0.22 * rSun
			}
			shU := uN + 0.18*math.Sin(vN*math.Pi*0.8)
			px := sunCX + (shU*uAx.X+vN*vAx.X)*rSun*1.02 + jx
			py := sunCY + (shU*uAx.Y+vN*vAx.Y)*rSun*1.02 + jy
			dx, dy := px-sunCX, py-sunCY
			dist := math.Hypot(dx, dy)
			if dist > rSun*0.98 || iu == 0 || iu == eSlices || iv == 0 || iv == eSlices {
				if dist > rSun*0.98 {
					px = sunCX + (dx/dist)*rSun*0.98
					py = sunCY + (dy/dist)*rSun*0.98
				}
			}
			eGrid[iu][iv] = render.Pt(px, py)
		}
	}

	fIdx := 0
	for iu := 0; iu < eSlices; iu++ {
		for iv := 0; iv < eSlices; iv++ {
			p00, p10, p11, p01 := eGrid[iu][iv], eGrid[iu+1][iv], eGrid[iu+1][iv+1], eGrid[iu][iv+1]
			cent := render.Pt(0.25*(p00.X+p10.X+p11.X+p01.X), 0.25*(p00.Y+p10.Y+p11.Y+p01.Y))
			if math.Hypot(cent.X-sunCX, cent.Y-sunCY) > rSun*0.97 {
				continue
			}
			pickEclipseCol := func(pt render.Point, idx int) color.RGBA {
				// Offset inner eclipse shadow disc
				dEcl := math.Hypot(pt.X-(sunCX+0.14*rSun), pt.Y-(sunCY+0.10*rSun))
				if dEcl < 0.54*rSun {
					if idx%3 == 0 {
						return twilightViolet
					}
					return deepObsidian
				}
				if dEcl < 0.72*rSun {
					if idx%2 == 0 {
						return terracottaRed
					}
					return cadmiumOrange
				}
				if idx%2 == 0 {
					return sunYellow
				}
				return goldenOchre
			}
			if (iu+iv)%2 == 0 {
				quad := render.InsetPolygon([]render.Point{p00, p10, p11, p01}, 0.55*scale)
				rc.FillPolygon(quad, pickEclipseCol(cent, fIdx), 1.0)
				fIdx++
			} else {
				tri1 := render.InsetPolygon([]render.Point{p00, p10, p11}, 0.55*scale)
				c1 := render.Pt((p00.X+p10.X+p11.X)/3.0, (p00.Y+p10.Y+p11.Y)/3.0)
				rc.FillPolygon(tri1, pickEclipseCol(c1, fIdx), 1.0)
				fIdx++

				tri2 := render.InsetPolygon([]render.Point{p00, p11, p01}, 0.55*scale)
				c2 := render.Pt((p00.X+p11.X+p01.X)/3.0, (p00.Y+p11.Y+p01.Y)/3.0)
				rc.FillPolygon(tri2, pickEclipseCol(c2, fIdx), 1.0)
				fIdx++
			}
		}
	}

	// 3. Stratified Sandstone Canyon Walls (Left and Right of the River Gorge)
	strataCount := int(math.Ceil((0.72 * h) / (brickThick * 0.90))) + 4
	strataPalette := []color.RGBA{
		terracottaRed, cadmiumOrange, goldenOchre, burntSienna,
		canyonPlum, terracottaRed, cadmiumOrange, twilightViolet,
	}

	for sb := 0; sb < strataCount; sb++ {
		baseY := 0.30*h + float64(sb)*brickThick*0.90
		brickLen := (21.0 + 4.5*math.Cos(float64(sb)*0.65)) * scale
		stepsX := int(math.Ceil((w + 60.0*scale) / brickLen))
		dx := (w + 60.0*scale) / float64(stepsX)
		xOffset := -30.0*scale + float64(sb%2)*brickLen*0.45

		evalStrataY := func(x float64) float64 {
			nx := x / math.Max(1.0, w)
			// Geological dip toward the canyon center + wave folding
			dip := 36.0 * scale * (1.0 - math.Abs(nx-0.50)*2.0)
			fold := 14.0 * scale * math.Sin(nx*2.0*math.Pi*2.2+float64(sb)*0.22)
			return baseY + dip + fold
		}

		for sx := 0; sx < stepsX; sx++ {
			x0 := xOffset + float64(sx)*dx
			x1 := x0 + dx*0.96
			p0 := render.Pt(x0, evalStrataY(x0))
			p1 := render.Pt(x1, evalStrataY(x1))
			mid := p0.Lerp(p1, 0.5)

			if mid.Y < canyonRimY(mid.X)-brickThick*0.2 || mid.Y > h+30*scale {
				continue
			}
			// Leave the central canyon floor open for the winding turquoise river
			rLeft, rRight := riverBounds(mid.Y)
			if mid.Y > 0.55*h && mid.X > rLeft-brickThick*0.3 && mid.X < rRight+brickThick*0.3 {
				continue
			}

			jitter := rng.Float64()
			col := strataPalette[(sb+int(jitter*2.5))%len(strataPalette)]
			// Sunlit inner cliff edges glow with golden ochre / sun yellow
			if mid.Y > 0.52*h && (math.Abs(mid.X-rLeft) < 28.0*scale || math.Abs(mid.X-rRight) < 28.0*scale) {
				if jitter > 0.38 {
					col = sunYellow
				} else {
					col = goldenOchre
				}
			} else if jitter < 0.11 {
				col = espressoDark
			}

			shear := 0.34
			if mid.X > 0.50*w {
				shear = -0.34
			}
			rc.DrawMosaicBrick(p0, p1, brickThick*0.47, shear, scale, col, 1.0)
		}
	}

	// 4. Winding Turquoise & Gold River Courses in the Canyon Floor (y in [0.55h, 1.0h])
	riverBands := int(math.Ceil((0.46 * h) / (brickThick * 0.88))) + 2
	for rb := 0; rb < riverBands; rb++ {
		y := 0.56*h + float64(rb)*brickThick*0.88
		if y > h+20*scale {
			break
		}
		rLeft, rRight := riverBounds(y)
		spanW := rRight - rLeft
		if spanW < 12.0*scale {
			continue
		}
		numBricks := int(math.Max(3, math.Round(spanW/(19.0*scale))))
		dx := spanW / float64(numBricks)

		for k := 0; k < numBricks; k++ {
			x0 := rLeft + float64(k)*dx
			x1 := x0 + dx*0.95
			u0 := (x0 - rLeft) / spanW
			u1 := (x1 - rLeft) / spanW
			// Ripple arch across the river width
			y0 := y + 6.5*scale*math.Sin(u0*math.Pi*2.0+float64(rb)*0.45)
			y1 := y + 6.5*scale*math.Sin(u1*math.Pi*2.0+float64(rb)*0.45)

			uMid := 0.5 * (u0 + u1)
			jitter := rng.Float64()

			var col color.RGBA
			switch {
			case math.Abs(uMid-0.50) < 0.18 && jitter > 0.42:
				// Shimmering eclipse reflection down the center of the river
				if jitter > 0.72 {
					col = sunYellow
				} else {
					col = goldenOchre
				}
			case uMid < 0.18 || uMid > 0.82:
				col = cobaltBlue
			case (rb+k)%3 == 0:
				col = paleIce
			case (rb+k)%2 == 0:
				col = mintAqua
			default:
				col = seafoamTeal
			}

			shear := 0.28 * math.Cos(uMid*math.Pi)
			rc.DrawMosaicBrick(render.Pt(x0, y0), render.Pt(x1, y1), brickThick*0.46, shear, scale, col, 1.0)
		}
	}

	c.MergeRenderCanvas(rc)
}
