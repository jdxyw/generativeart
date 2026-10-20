package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type mosaicVortex struct {
	ringSteps  int
	brickScale float64
}

// NewMosaicVortex returns a mosaicVortex engine that renders a Starry-Night-inspired
// swirling sky vortex, low-poly crystal crescent moon, radiant star halos, rolling
// twilight hills, and a faceted cypress spire in the crisp MosaicSun aesthetic.
func NewMosaicVortex(ringSteps int, brickScale float64) *mosaicVortex {
	if ringSteps <= 0 {
		ringSteps = 34
	}
	if brickScale <= 0 {
		brickScale = 1.0
	}
	return &mosaicVortex{
		ringSteps:  ringSteps,
		brickScale: brickScale,
	}
}

// Generative renders the anisotropic flow mosaic vortex, low-poly moon, and cypress spire.
func (mv *mosaicVortex) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := (math.Min(w, h) / 600.0) * mv.brickScale
	if scale <= 0 {
		scale = 1.0
	}

	// Quantized stained-mosaic & poster palette
	deepMidnight := color.RGBA{R: 16, G: 20, B: 38, A: 255}
	obsidianDark := color.RGBA{R: 28, G: 22, B: 34, A: 255}
	prussianBlue := color.RGBA{R: 28, G: 56, B: 112, A: 255}
	cobaltBlue := color.RGBA{R: 46, G: 112, B: 176, A: 255}
	ceruleanTeal := color.RGBA{R: 72, G: 168, B: 178, A: 255}
	mintAqua := color.RGBA{R: 152, G: 218, B: 196, A: 255}
	paleIce := color.RGBA{R: 220, G: 244, B: 234, A: 255}

	sunYellow := color.RGBA{R: 250, G: 216, B: 84, A: 255}
	goldenAmber := color.RGBA{R: 242, G: 172, B: 50, A: 255}
	cadmiumOrange := color.RGBA{R: 232, G: 118, B: 42, A: 255}
	coralFlame := color.RGBA{R: 204, G: 68, B: 48, A: 255}

	royalViolet := color.RGBA{R: 92, G: 42, B: 118, A: 255}
	vibrantMagenta := color.RGBA{R: 198, G: 36, B: 114, A: 255}
	forestEmerald := color.RGBA{R: 26, G: 96, B: 82, A: 255}
	deepCypress := color.RGBA{R: 18, G: 52, B: 50, A: 255}
	bronzeOchre := color.RGBA{R: 168, G: 112, B: 44, A: 255}

	rc.FillBackground(deepMidnight)
	rc.SetBlendMode(render.BlendOver)

	rng := rand.New(rand.NewSource(202607))

	// Key focal anchors
	moonCX, moonCY := 0.75*w, 0.24*h
	rMoon := 0.185 * w

	vortexCX, vortexCY := 0.35*w, 0.36*h
	rVortexCore := 0.085 * w

	brickThick := 9.2 * scale
	if mv.ringSteps != 34 {
		brickThick = (312.0 / float64(mv.ringSteps)) * scale
	}

	// Horizon curve separating the sky vortex from the rolling twilight hills
	hillHorizon := func(x float64) float64 {
		nx := x / math.Max(1.0, w)
		return 0.66*h + 26.0*scale*math.Sin(nx*2.0*math.Pi*1.15-0.3) - 18.0*scale*math.Cos(nx*2.0*math.Pi*2.1+0.5)
	}

	// Domain partition between Moon Halo rings and Primary Galaxy Vortex rings
	isMoonDomain := func(x, y float64) bool {
		dMoon := math.Hypot(x-moonCX, y-moonCY) - rMoon
		dVortex := math.Hypot(x-vortexCX, y-vortexCY) - rVortexCore
		return dMoon < dVortex*0.62+14.0*scale
	}

	// 1. Moon Halo Concentric Mosaic Courses (around upper-right Moon)
	for ring := 0; ring < 24; ring++ {
		rBase := rMoon + (float64(ring)+0.5)*brickThick
		circum := 2.0 * math.Pi * rBase
		brickLen := common.Constrain((20.0+4.0*math.Sin(float64(ring)*0.6))*scale, 15.0*scale, 32.0*scale)
		numBricks := int(math.Max(18, math.Round(circum/brickLen)))
		dTheta := (2.0 * math.Pi) / float64(numBricks)
		ringPhase := float64(ring) * 0.41

		for k := 0; k < numBricks; k++ {
			th0 := float64(k)*dTheta + ringPhase
			th1 := th0 + dTheta*0.96
			thMid := 0.5 * (th0 + th1)

			warp0 := 6.5 * scale * math.Sin(4.0*th0+float64(ring)*0.35)
			warp1 := 6.5 * scale * math.Sin(4.0*th1+float64(ring)*0.35)
			p0 := render.Pt(moonCX+(rBase+warp0)*math.Cos(th0), moonCY+(rBase+warp0)*math.Sin(th0))
			p1 := render.Pt(moonCX+(rBase+warp1)*math.Cos(th1), moonCY+(rBase+warp1)*math.Sin(th1))
			mid := p0.Lerp(p1, 0.5)

			if mid.X < -30 || mid.X > w+30 || mid.Y < -30 || mid.Y > hillHorizon(mid.X)+brickThick*0.4 {
				continue
			}
			if !isMoonDomain(mid.X, mid.Y) {
				continue
			}

			jitter := rng.Float64()
			u := float64(ring) / 14.0
			beam := math.Cos(thMid*6.0 - float64(ring)*0.45)

			var col color.RGBA
			switch {
			case ring < 2 && beam > -0.15:
				if jitter > 0.45 {
					col = sunYellow
				} else {
					col = goldenAmber
				}
			case u < 0.45 && beam > 0.38:
				if jitter > 0.55 {
					col = goldenAmber
				} else {
					col = cadmiumOrange
				}
			case jitter < 0.08:
				col = prussianBlue
			default:
				val := math.Sin(u*8.5-thMid*2.0) + (jitter-0.5)*0.4
				switch {
				case val > 0.48:
					col = paleIce
				case val > 0.05:
					col = mintAqua
				case val > -0.42:
					col = ceruleanTeal
				default:
					col = cobaltBlue
				}
			}

			shear := 0.30 * math.Sin(thMid*2.0+float64(ring)*0.3)
			rc.DrawMosaicBrick(p0, p1, brickThick*0.48, shear, scale, col, 1.0)
		}
	}

	// 2. Primary Swirling Sky Vortex Courses (around vortexCX, vortexCY)
	maxRadius := math.Hypot(w, h) * 0.95
	for ring := 0; ring < mv.ringSteps+16; ring++ {
		rBase := rVortexCore + (float64(ring)+0.5)*brickThick
		if rBase-brickThick > maxRadius {
			break
		}
		circum := 2.0 * math.Pi * rBase
		brickLen := common.Constrain((20.0+4.5*math.Cos(float64(ring)*0.65))*scale, 15.0*scale, 33.0*scale)
		numBricks := int(math.Max(16, math.Round(circum/brickLen)))
		dTheta := (2.0 * math.Pi) / float64(numBricks)
		ringPhase := float64(ring) * 0.33

		for k := 0; k < numBricks; k++ {
			th0 := float64(k)*dTheta + ringPhase
			th1 := th0 + dTheta*0.96
			thMid := 0.5 * (th0 + th1)

			// Spiral S-wave radial modulation for dynamic Starry-Night turbulence
			warpFn := func(th float64) float64 {
				return 11.0*scale*math.Sin(2.0*th-float64(ring)*0.28) +
					5.5*scale*math.Cos(3.0*th+float64(ring)*0.22)
			}
			r0 := rBase + warpFn(th0)
			r1 := rBase + warpFn(th1)

			p0 := render.Pt(vortexCX+r0*math.Cos(th0), vortexCY+r0*math.Sin(th0))
			p1 := render.Pt(vortexCX+r1*math.Cos(th1), vortexCY+r1*math.Sin(th1))
			mid := p0.Lerp(p1, 0.5)

			if mid.X < -35 || mid.X > w+35 || mid.Y < -35 || mid.Y > hillHorizon(mid.X)+brickThick*0.4 {
				continue
			}
			if isMoonDomain(mid.X, mid.Y) {
				continue
			}

			jitter := rng.Float64()
			spiralArm := math.Sin(2.0*thMid - float64(ring)*0.38)
			var col color.RGBA
			switch {
			case spiralArm > 0.62 && ring < 18:
				if jitter > 0.62 {
					col = sunYellow
				} else if jitter > 0.25 {
					col = paleIce
				} else {
					col = mintAqua
				}
			case spiralArm > 0.18:
				if jitter > 0.86 {
					col = goldenAmber
				} else if jitter > 0.42 {
					col = mintAqua
				} else {
					col = ceruleanTeal
				}
			case spiralArm > -0.35:
				if jitter < 0.12 {
					col = deepMidnight
				} else if jitter > 0.88 {
					col = cadmiumOrange
				} else {
					col = cobaltBlue
				}
			default:
				if jitter < 0.18 {
					col = deepMidnight
				} else if jitter < 0.52 {
					col = prussianBlue
				} else {
					col = cobaltBlue
				}
			}

			shear := 0.34 * math.Cos(thMid*2.0-float64(ring)*0.25)
			rc.DrawMosaicBrick(p0, p1, brickThick*0.48, shear, scale, col, 1.0)
		}
	}

	// 3. Low-Poly Vortex Eye Core at (vortexCX, vortexCY)
	vortexRim := make([]render.Point, 28)
	for i := 0; i < 28; i++ {
		ang := (float64(i) / 28.0) * 2.0 * math.Pi
		vortexRim[i] = render.Pt(vortexCX+rVortexCore*1.02*math.Cos(ang), vortexCY+rVortexCore*1.02*math.Sin(ang))
	}
	rc.FillPolygon(vortexRim, obsidianDark, 1.0)

	const vCoreRings = 3
	const vCoreSectors = 10
	vGrid := make([][]render.Point, vCoreRings+1)
	for ir := 0; ir <= vCoreRings; ir++ {
		vGrid[ir] = make([]render.Point, vCoreSectors)
		rFrac := float64(ir) / float64(vCoreRings)
		rCur := rFrac * rVortexCore * 0.96
		for is := 0; is < vCoreSectors; is++ {
			ang := (float64(is)/float64(vCoreSectors))*2.0*math.Pi + float64(ir)*0.32
			if ir > 0 && ir < vCoreRings {
				ang += (rng.Float64() - 0.5) * 0.18
			}
			vGrid[ir][is] = render.Pt(vortexCX+rCur*math.Cos(ang), vortexCY+rCur*math.Sin(ang))
		}
	}
	vCoreColors := []color.RGBA{sunYellow, paleIce, mintAqua, goldenAmber, ceruleanTeal, cadmiumOrange}
	for ir := 0; ir < vCoreRings; ir++ {
		for is := 0; is < vCoreSectors; is++ {
			nextS := (is + 1) % vCoreSectors
			if ir == 0 {
				tri := render.InsetPolygon([]render.Point{
					render.Pt(vortexCX, vortexCY),
					vGrid[1][is],
					vGrid[1][nextS],
				}, 0.5*scale)
				rc.FillPolygon(tri, vCoreColors[(is)%2], 1.0)
			} else {
				q00 := vGrid[ir][is]
				q10 := vGrid[ir+1][is]
				q11 := vGrid[ir+1][nextS]
				q01 := vGrid[ir][nextS]
				tri1 := render.InsetPolygon([]render.Point{q00, q10, q11}, 0.5*scale)
				tri2 := render.InsetPolygon([]render.Point{q00, q11, q01}, 0.5*scale)
				rc.FillPolygon(tri1, vCoreColors[(ir+is)%len(vCoreColors)], 1.0)
				rc.FillPolygon(tri2, vCoreColors[(ir+is+2)%len(vCoreColors)], 1.0)
			}
		}
	}

	// 4. Inside the Moon Core: Low-Poly Golden Crescent & Obsidian Shadow Facets
	moonRim := make([]render.Point, 36)
	for i := 0; i < 36; i++ {
		ang := (float64(i) / 36.0) * 2.0 * math.Pi
		moonRim[i] = render.Pt(moonCX+rMoon*1.01*math.Cos(ang), moonCY+rMoon*1.01*math.Sin(ang))
	}
	rc.FillPolygon(moonRim, obsidianDark, 1.0)

	const mSlices = 6
	mGrid := make([][]render.Point, mSlices+1)
	diagAng := -0.48
	uAx := render.Pt(math.Cos(diagAng), math.Sin(diagAng))
	vAx := uAx.Perp()
	for iu := 0; iu <= mSlices; iu++ {
		mGrid[iu] = make([]render.Point, mSlices+1)
		uN := (float64(iu)/float64(mSlices))*2.0 - 1.0
		for iv := 0; iv <= mSlices; iv++ {
			vN := (float64(iv)/float64(mSlices))*2.0 - 1.0
			jx, jy := 0.0, 0.0
			if iu > 0 && iu < mSlices && iv > 0 && iv < mSlices {
				jx = (rng.Float64() - 0.5) * 0.22 * rMoon
				jy = (rng.Float64() - 0.5) * 0.22 * rMoon
			}
			shU := uN + 0.18*math.Sin(vN*math.Pi*0.8)
			px := moonCX + (shU*uAx.X+vN*vAx.X)*rMoon*1.02 + jx
			py := moonCY + (shU*uAx.Y+vN*vAx.Y)*rMoon*1.02 + jy
			dx, dy := px-moonCX, py-moonCY
			dist := math.Hypot(dx, dy)
			if dist > rMoon*0.98 || iu == 0 || iu == mSlices || iv == 0 || iv == mSlices {
				if dist > rMoon*0.98 {
					px = moonCX + (dx/dist)*rMoon*0.98
					py = moonCY + (dy/dist)*rMoon*0.98
				}
			}
			mGrid[iu][iv] = render.Pt(px, py)
		}
	}

	// Crescent shadow circle offset to upper-left of moon center
	shadCX := moonCX - 0.26*rMoon
	shadCY := moonCY - 0.16*rMoon
	rShad := 0.68 * rMoon

	sampleMoonFacet := func(cPt render.Point, idx int) color.RGBA {
		dShad := math.Hypot(cPt.X-shadCX, cPt.Y-shadCY)
		if dShad < rShad*0.82 {
			if idx%3 == 0 {
				return prussianBlue
			}
			return obsidianDark
		}
		if dShad < rShad {
			return coralFlame
		}
		dx := (cPt.X - moonCX) / rMoon
		dy := (cPt.Y - moonCY) / rMoon
		band := dx*0.6 + dy*0.8
		switch {
		case band < -0.15:
			return paleIce
		case band < 0.22:
			if idx%2 == 0 {
				return sunYellow
			}
			return goldenAmber
		default:
			if idx%2 == 0 {
				return goldenAmber
			}
			return cadmiumOrange
		}
	}

	fIdx := 0
	for iu := 0; iu < mSlices; iu++ {
		for iv := 0; iv < mSlices; iv++ {
			p00, p10, p11, p01 := mGrid[iu][iv], mGrid[iu+1][iv], mGrid[iu+1][iv+1], mGrid[iu][iv+1]
			cent := render.Pt(0.25*(p00.X+p10.X+p11.X+p01.X), 0.25*(p00.Y+p10.Y+p11.Y+p01.Y))
			if math.Hypot(cent.X-moonCX, cent.Y-moonCY) > rMoon*0.97 {
				continue
			}
			if (iu+iv)%2 == 0 {
				quad := render.InsetPolygon([]render.Point{p00, p10, p11, p01}, 0.55*scale)
				rc.FillPolygon(quad, sampleMoonFacet(cent, fIdx), 1.0)
				fIdx++
			} else {
				tri1 := render.InsetPolygon([]render.Point{p00, p10, p11}, 0.55*scale)
				c1 := render.Pt((p00.X+p10.X+p11.X)/3.0, (p00.Y+p10.Y+p11.Y)/3.0)
				rc.FillPolygon(tri1, sampleMoonFacet(c1, fIdx), 1.0)
				fIdx++

				tri2 := render.InsetPolygon([]render.Point{p00, p11, p01}, 0.55*scale)
				c2 := render.Pt((p00.X+p11.X+p01.X)/3.0, (p00.Y+p11.Y+p01.Y)/3.0)
				rc.FillPolygon(tri2, sampleMoonFacet(c2, fIdx), 1.0)
				fIdx++
			}
		}
	}

	// 5. Rolling Twilight Hills (Lower Region y > hillHorizon(x))
	hillBands := int(math.Ceil((0.40 * h) / (brickThick * 0.90))) + 4
	for hb := 0; hb < hillBands; hb++ {
		bandFrac := float64(hb) / math.Max(1.0, float64(hillBands-1))
		baseY := 0.64*h + float64(hb)*brickThick*0.90

		brickLen := (21.0 + 5.0*math.Sin(float64(hb)*0.7)) * scale
		stepsX := int(math.Ceil((w + 60.0*scale) / brickLen))
		dx := (w + 60.0*scale) / float64(stepsX)
		xOffset := -30.0*scale + float64(hb%2)*brickLen*0.45

		evalHillY := func(x float64) float64 {
			nx := x / math.Max(1.0, w)
			h1 := 24.0 * scale * math.Sin(nx*2.0*math.Pi*1.15-0.3+float64(hb)*0.18)
			h2 := -16.0 * scale * math.Cos(nx*2.0*math.Pi*2.1+0.5-float64(hb)*0.28)
			return baseY + h1 + h2
		}

		for sx := 0; sx < stepsX; sx++ {
			x0 := xOffset + float64(sx)*dx
			x1 := x0 + dx*0.96
			p0 := render.Pt(x0, evalHillY(x0))
			p1 := render.Pt(x1, evalHillY(x1))
			mid := p0.Lerp(p1, 0.5)
			if mid.Y < 0.58*h || mid.Y > h+30*scale {
				continue
			}

			jitter := rng.Float64()
			nx := mid.X / math.Max(1.0, w)
			wavePhase := math.Sin(nx*8.5 + bandFrac*10.0)

			var col color.RGBA
			switch {
			case hb < 2 && jitter > 0.62:
				col = goldenAmber
			case nx > 0.55 && wavePhase > 0.15:
				if jitter > 0.78 {
					col = cadmiumOrange
				} else {
					col = vibrantMagenta
				}
			case wavePhase > 0.42:
				col = ceruleanTeal
			case wavePhase > -0.05:
				col = cobaltBlue
			case wavePhase > -0.48:
				col = royalViolet
			default:
				col = prussianBlue
			}

			shear := 0.35
			if hb%2 == 1 {
				shear = -0.28
			}
			rc.DrawMosaicBrick(p0, p1, brickThick*0.47, shear, scale, col, 1.0)
		}
	}

	// 6. Foreground Left Low-Poly Cypress Flame Spire
	// Sculpted flame silhouette rising from y = h at x in [0.06w, 0.25w] up to tip at (0.15w, 0.20h)
	const cRows = 11
	const cCols = 4
	cGrid := make([][]render.Point, cRows+1)
	for ir := 0; ir <= cRows; ir++ {
		cGrid[ir] = make([]render.Point, cCols+1)
		v := float64(ir) / float64(cRows) // 0 = bottom, 1 = flame tip
		yBase := h - v*(0.80*h)
		// Sinuous flame centerline and tapering half-width
		centerX := 0.155*w + 16.0*scale*math.Sin(v*math.Pi*1.65)
		halfW := (0.095 * w) * math.Pow(1.0-v*0.92, 0.78) * (1.0 + 0.14*math.Sin(v*math.Pi*3.0))

		for ic := 0; ic <= cCols; ic++ {
			u := (float64(ic)/float64(cCols))*2.0 - 1.0
			px := centerX + u*halfW
			py := yBase + 6.0*scale*math.Cos(u*math.Pi*0.5)*math.Sin(v*math.Pi)
			if ir > 0 && ir < cRows && ic > 0 && ic < cCols {
				px += (rng.Float64() - 0.5) * 0.28 * halfW
				py += (rng.Float64() - 0.5) * 14.0 * scale
			}
			cGrid[ir][ic] = render.Pt(px, py)
		}
	}

	// Backing silhouette for dark grout inside the cypress flame
	cypressOutline := make([]render.Point, 0, (cRows+1)*2)
	for ir := 0; ir <= cRows; ir++ {
		cypressOutline = append(cypressOutline, cGrid[ir][0])
	}
	for ir := cRows; ir >= 0; ir-- {
		cypressOutline = append(cypressOutline, cGrid[ir][cCols])
	}
	rc.FillPolygon(cypressOutline, obsidianDark, 1.0)

	for ir := 0; ir < cRows; ir++ {
		for ic := 0; ic < cCols; ic++ {
			p00 := cGrid[ir][ic]
			p10 := cGrid[ir+1][ic]
			p11 := cGrid[ir+1][ic+1]
			p01 := cGrid[ir][ic+1]

			pickCypressCol := func(uFrac, vFrac float64, idx int) color.RGBA {
				// Right flank catches warm moonlight; left flank stays deep cypress/emerald
				if uFrac > 0.65 && (idx%2 == 0 || vFrac > 0.5) {
					if idx%3 == 0 {
						return goldenAmber
					}
					return bronzeOchre
				}
				if uFrac > 0.35 {
					if idx%2 == 0 {
						return forestEmerald
					}
					return ceruleanTeal
				}
				if idx%3 == 0 {
					return obsidianDark
				}
				return deepCypress
			}

			uFrac := (float64(ic) + 0.5) / float64(cCols)
			vFrac := (float64(ir) + 0.5) / float64(cRows)
			tri1 := render.InsetPolygon([]render.Point{p00, p10, p11}, 0.55*scale)
			tri2 := render.InsetPolygon([]render.Point{p00, p11, p01}, 0.55*scale)
			rc.FillPolygon(tri1, pickCypressCol(uFrac, vFrac, ir+ic), 1.0)
			rc.FillPolygon(tri2, pickCypressCol(uFrac+0.15, vFrac, ir+ic+1), 1.0)
		}
	}

	c.MergeRenderCanvas(rc)
}
