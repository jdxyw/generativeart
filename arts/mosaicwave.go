package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type mosaicWave struct {
	waveCourses int
	brickScale  float64
}

// NewMosaicWave returns a mosaicWave engine that renders a Ukiyo-e-inspired
// Great Cresting Wave of anisotropic mosaic tesserae, low-poly geometric sea-foam
// claws, a faceted crimson sun disc, and a snow-capped volcanic peak.
func NewMosaicWave(waveCourses int, brickScale float64) *mosaicWave {
	if waveCourses <= 0 {
		waveCourses = 32
	}
	if brickScale <= 0 {
		brickScale = 1.0
	}
	return &mosaicWave{
		waveCourses: waveCourses,
		brickScale:  brickScale,
	}
}

// Generative renders the cresting wave mosaic, low-poly foam shards, crimson sun, and peak.
func (mw *mosaicWave) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := (math.Min(w, h) / 600.0) * mw.brickScale
	if scale <= 0 {
		scale = 1.0
	}

	// Curated Ukiyo-e & modernist stained-mosaic palette
	deepNavy := color.RGBA{R: 18, G: 24, B: 44, A: 255}
	espressoDark := color.RGBA{R: 42, G: 26, B: 28, A: 255}
	prussianBlue := color.RGBA{R: 26, G: 58, B: 110, A: 255}
	cobaltBlue := color.RGBA{R: 44, G: 112, B: 172, A: 255}
	seafoamTeal := color.RGBA{R: 84, G: 168, B: 166, A: 255}
	mintAqua := color.RGBA{R: 158, G: 218, B: 196, A: 255}
	ivoryFoam := color.RGBA{R: 242, G: 244, B: 232, A: 255}
	paleIce := color.RGBA{R: 212, G: 238, B: 230, A: 255}

	vermilionRed := color.RGBA{R: 216, G: 62, B: 44, A: 255}
	crimsonCoral := color.RGBA{R: 234, G: 98, B: 56, A: 255}
	cadmiumOrange := color.RGBA{R: 238, G: 142, B: 52, A: 255}
	goldenAmber := color.RGBA{R: 246, G: 192, B: 68, A: 255}
	warmCream := color.RGBA{R: 246, G: 228, B: 182, A: 255}

	rc.FillBackground(deepNavy)
	rc.SetBlendMode(render.BlendOver)

	rng := rand.New(rand.NewSource(202608))

	sunCX, sunCY := 0.72*w, 0.28*h
	rSun := 0.18 * w

	brickThick := 9.4 * scale
	if mw.waveCourses != 32 {
		brickThick = (300.0 / float64(mw.waveCourses)) * scale
	}

	// Great Wave Upper Envelope: rises high on the left/center (crest at x=0.34w, y=0.22h),
	// then plunges into a deep trough on the right (x=0.72w, y=0.70h) to reveal the sun and Mount Fuji.
	greatWaveTopY := func(x float64) float64 {
		nx := x / math.Max(1.0, w)
		arch := 0.56*h - 0.34*h*math.Exp(-math.Pow((nx-0.30)/0.28, 2.0))
		trough := 0.16 * h * math.Exp(-math.Pow((nx-0.74)/0.22, 2.0))
		ripple := 10.0 * scale * math.Sin(nx*2.0*math.Pi*1.6)
		return arch + trough + ripple
	}

	// Mount Fuji Silhouette on the right horizon (x in [0.54w, 0.96w])
	fujiTopY := func(x float64) float64 {
		nx := x / math.Max(1.0, w)
		dist := math.Abs(nx - 0.76)
		if dist < 0.035 {
			// Flat volcanic crater rim
			return 0.52*h + 4.0*scale*math.Cos((nx-0.76)*80.0)
		}
		// Classic concave volcanic slopes
		return 0.52*h + 0.36*h*math.Pow(dist/0.22, 1.35)
	}

	// 1. Background Radiating Sunset Sky Courses around the Rising Sun
	maxSkyRadius := math.Hypot(w, h) * 0.90
	for ring := 0; ring < 42; ring++ {
		rBase := rSun + (float64(ring)+0.5)*brickThick
		if rBase > maxSkyRadius {
			break
		}
		circum := 2.0 * math.Pi * rBase
		brickLen := common.Constrain((20.0+4.0*math.Sin(float64(ring)*0.6))*scale, 15.0*scale, 32.0*scale)
		numBricks := int(math.Max(18, math.Round(circum/brickLen)))
		dTheta := (2.0 * math.Pi) / float64(numBricks)
		ringPhase := float64(ring) * 0.37

		for k := 0; k < numBricks; k++ {
			th0 := float64(k)*dTheta + ringPhase
			th1 := th0 + dTheta*0.96
			thMid := 0.5 * (th0 + th1)

			warp0 := 7.0 * scale * math.Sin(3.0*th0+float64(ring)*0.30)
			warp1 := 7.0 * scale * math.Sin(3.0*th1+float64(ring)*0.30)
			p0 := render.Pt(sunCX+(rBase+warp0)*math.Cos(th0), sunCY+(rBase+warp0)*math.Sin(th0))
			p1 := render.Pt(sunCX+(rBase+warp1)*math.Cos(th1), sunCY+(rBase+warp1)*math.Sin(th1))
			mid := p0.Lerp(p1, 0.5)

			if mid.X < -30 || mid.X > w+30 || mid.Y < -30 || mid.Y > h+30 {
				continue
			}
			// Clip sky bricks behind the Great Wave or Mount Fuji
			if mid.Y > greatWaveTopY(mid.X)-brickThick*0.25 || mid.Y > fujiTopY(mid.X)-brickThick*0.25 {
				continue
			}

			jitter := rng.Float64()
			u := float64(ring) / 22.0
			ray := math.Cos(thMid*5.0 + float64(ring)*0.35)

			var col color.RGBA
			switch {
			case u < 0.38 && ray > 0.35:
				if jitter > 0.5 {
					col = goldenAmber
				} else {
					col = crimsonCoral
				}
			case jitter < 0.08:
				col = deepNavy
			default:
				val := math.Sin(u*9.0-thMid*2.0) + (jitter-0.5)*0.38
				switch {
				case val > 0.45:
					col = warmCream
				case val > 0.02:
					col = goldenAmber
				case val > -0.40:
					col = mintAqua
				default:
					col = seafoamTeal
				}
			}

			shear := 0.28 * math.Cos(thMid*2.0+float64(ring)*0.25)
			rc.DrawMosaicBrick(p0, p1, brickThick*0.48, shear, scale, col, 1.0)
		}
	}

	// 2. Inside the Rising Sun Core: Low-Poly Vermilion, Coral & Gold Shards
	sunRim := make([]render.Point, 36)
	for i := 0; i < 36; i++ {
		ang := (float64(i) / 36.0) * 2.0 * math.Pi
		sunRim[i] = render.Pt(sunCX+rSun*1.01*math.Cos(ang), sunCY+rSun*1.01*math.Sin(ang))
	}
	rc.FillPolygon(sunRim, espressoDark, 1.0)

	const sSlices = 6
	sGrid := make([][]render.Point, sSlices+1)
	uAx := render.Pt(math.Cos(-0.52), math.Sin(-0.52))
	vAx := uAx.Perp()
	for iu := 0; iu <= sSlices; iu++ {
		sGrid[iu] = make([]render.Point, sSlices+1)
		uN := (float64(iu)/float64(sSlices))*2.0 - 1.0
		for iv := 0; iv <= sSlices; iv++ {
			vN := (float64(iv)/float64(sSlices))*2.0 - 1.0
			jx, jy := 0.0, 0.0
			if iu > 0 && iu < sSlices && iv > 0 && iv < sSlices {
				jx = (rng.Float64() - 0.5) * 0.22 * rSun
				jy = (rng.Float64() - 0.5) * 0.22 * rSun
			}
			shU := uN + 0.20*math.Sin(vN*math.Pi*0.8)
			px := sunCX + (shU*uAx.X+vN*vAx.X)*rSun*1.02 + jx
			py := sunCY + (shU*uAx.Y+vN*vAx.Y)*rSun*1.02 + jy
			dx, dy := px-sunCX, py-sunCY
			dist := math.Hypot(dx, dy)
			if dist > rSun*0.98 || iu == 0 || iu == sSlices || iv == 0 || iv == sSlices {
				if dist > rSun*0.98 {
					px = sunCX + (dx/dist)*rSun*0.98
					py = sunCY + (dy/dist)*rSun*0.98
				}
			}
			sGrid[iu][iv] = render.Pt(px, py)
		}
	}

	shardIdx := 0
	for iu := 0; iu < sSlices; iu++ {
		for iv := 0; iv < sSlices; iv++ {
			p00, p10, p11, p01 := sGrid[iu][iv], sGrid[iu+1][iv], sGrid[iu+1][iv+1], sGrid[iu][iv+1]
			cent := render.Pt(0.25*(p00.X+p10.X+p11.X+p01.X), 0.25*(p00.Y+p10.Y+p11.Y+p01.Y))
			if math.Hypot(cent.X-sunCX, cent.Y-sunCY) > rSun*0.97 {
				continue
			}
			pickSunCol := func(pt render.Point, idx int) color.RGBA {
				dx := (pt.X - sunCX) / rSun
				dy := (pt.Y - sunCY) / rSun
				rN := math.Hypot(dx, dy)
				ang := math.Atan2(dy, dx)
				if rN > 0.56 && ang > 0.0 && ang < 1.9 {
					if rN > 0.72 || idx%3 == 0 {
						return espressoDark
					}
					return vermilionRed
				}
				diag := dx*0.65 + dy*0.75
				switch {
				case diag < -0.25:
					return warmCream
				case diag < 0.08:
					if idx%2 == 0 {
						return goldenAmber
					}
					return cadmiumOrange
				case diag < 0.42:
					return crimsonCoral
				default:
					return vermilionRed
				}
			}
			if (iu+iv)%2 == 0 {
				quad := render.InsetPolygon([]render.Point{p00, p10, p11, p01}, 0.55*scale)
				rc.FillPolygon(quad, pickSunCol(cent, shardIdx), 1.0)
				shardIdx++
			} else {
				tri1 := render.InsetPolygon([]render.Point{p00, p10, p11}, 0.55*scale)
				c1 := render.Pt((p00.X+p10.X+p11.X)/3.0, (p00.Y+p10.Y+p11.Y)/3.0)
				rc.FillPolygon(tri1, pickSunCol(c1, shardIdx), 1.0)
				shardIdx++

				tri2 := render.InsetPolygon([]render.Point{p00, p11, p01}, 0.55*scale)
				c2 := render.Pt((p00.X+p11.X+p01.X)/3.0, (p00.Y+p11.Y+p01.Y)/3.0)
				rc.FillPolygon(tri2, pickSunCol(c2, shardIdx), 1.0)
				shardIdx++
			}
		}
	}

	// 3. Low-Poly Faceted Mount Fuji on the Right Horizon
	const fRows = 5
	const fCols = 6
	fGrid := make([][]render.Point, fRows+1)
	for ir := 0; ir <= fRows; ir++ {
		fGrid[ir] = make([]render.Point, fCols+1)
		v := float64(ir) / float64(fRows) // 0 = summit, 1 = base
		halfSpan := (0.035 + v*0.21) * w
		for ic := 0; ic <= fCols; ic++ {
			u := (float64(ic)/float64(fCols))*2.0 - 1.0
			px := 0.76*w + u*halfSpan
			py := fujiTopY(0.76*w+halfSpan) + v*12.0*scale
			if ir == 0 {
				py = fujiTopY(px)
			} else {
				yTop := fujiTopY(0.76 * w)
				yBot := 0.78 * h
				py = yTop + math.Pow(v, 1.15)*(yBot-yTop)
			}
			if ir > 0 && ir < fRows && ic > 0 && ic < fCols {
				px += (rng.Float64() - 0.5) * 14.0 * scale
				py += (rng.Float64() - 0.5) * 9.0 * scale
			}
			fGrid[ir][ic] = render.Pt(px, py)
		}
	}

	fujiOutline := make([]render.Point, 0, (fCols+1)*2)
	for ic := 0; ic <= fCols; ic++ {
		fujiOutline = append(fujiOutline, fGrid[0][ic])
	}
	for ir := 1; ir <= fRows; ir++ {
		fujiOutline = append(fujiOutline, fGrid[ir][fCols])
	}
	for ic := fCols - 1; ic >= 0; ic-- {
		fujiOutline = append(fujiOutline, fGrid[fRows][ic])
	}
	for ir := fRows - 1; ir >= 1; ir-- {
		fujiOutline = append(fujiOutline, fGrid[ir][0])
	}
	rc.FillPolygon(fujiOutline, deepNavy, 1.0)

	for ir := 0; ir < fRows; ir++ {
		for ic := 0; ic < fCols; ic++ {
			p00, p10, p11, p01 := fGrid[ir][ic], fGrid[ir+1][ic], fGrid[ir+1][ic+1], fGrid[ir][ic+1]
			pickFujiCol := func(rIdx, cIdx int) color.RGBA {
				// Snow cap on the upper 2 tiers with jagged ridge fingers
				if rIdx == 0 || (rIdx == 1 && (cIdx+rIdx)%2 == 0) {
					if cIdx < fCols/2 {
						return ivoryFoam
					}
					return paleIce
				}
				if cIdx < 2 {
					return crimsonCoral
				}
				if (rIdx+cIdx)%2 == 0 {
					return cobaltBlue
				}
				return prussianBlue
			}
			tri1 := render.InsetPolygon([]render.Point{p00, p10, p11}, 0.55*scale)
			tri2 := render.InsetPolygon([]render.Point{p00, p11, p01}, 0.55*scale)
			rc.FillPolygon(tri1, pickFujiCol(ir, ic), 1.0)
			rc.FillPolygon(tri2, pickFujiCol(ir, ic+1), 1.0)
		}
	}

	// 4. Sweeping Great Wave & Foreground Rolling Sea Courses
	totalWaveBands := mw.waveCourses + 6
	for wb := 0; wb < totalWaveBands; wb++ {
		bandFrac := float64(wb) / math.Max(1.0, float64(totalWaveBands-1))
		brickLen := (20.0 + 5.5*math.Sin(float64(wb)*0.55)) * scale
		stepsX := int(math.Ceil((w + 60.0*scale) / brickLen))
		dx := (w + 60.0*scale) / float64(stepsX)
		xOffset := -30.0*scale + float64(wb%2)*brickLen*0.45

		evalWaveCourseY := func(x float64) float64 {
			nx := x / math.Max(1.0, w)
			topEnv := greatWaveTopY(x)
			// Blend smoothly from the high curling arch at wb=0 to foreground swells at bottom
			secondarySwell := 0.72*h + 26.0*scale*math.Sin(nx*2.0*math.Pi*1.45+float64(wb)*0.24) -
				15.0*scale*math.Cos(nx*2.0*math.Pi*2.6-float64(wb)*0.32)
			blend := common.Constrain(bandFrac*1.25, 0, 1)
			y := (1.0-blend)*topEnv + blend*secondarySwell + float64(wb)*brickThick*0.88
			// Curl overhang compression near the wave lip (nx in [0.22, 0.52])
			if wb < 8 {
				curlHook := 14.0 * scale * math.Exp(-math.Pow((nx-0.42)/0.14, 2.0)) * math.Sin(float64(wb)*0.45)
				y += curlHook
			}
			return y
		}

		for sx := 0; sx < stepsX; sx++ {
			x0 := xOffset + float64(sx)*dx
			x1 := x0 + dx*0.96
			p0 := render.Pt(x0, evalWaveCourseY(x0))
			p1 := render.Pt(x1, evalWaveCourseY(x1))
			mid := p0.Lerp(p1, 0.5)
			if mid.Y > h+30*scale || mid.X < -30*scale || mid.X > w+30*scale {
				continue
			}

			jitter := rng.Float64()
			nx := mid.X / math.Max(1.0, w)
			phase := math.Sin(nx*9.0 + float64(wb)*0.55)

			var col color.RGBA
			switch {
			case wb == 0 || (wb == 1 && jitter > 0.28) || (wb == 2 && phase > 0.35 && jitter > 0.45):
				// White/Ice foam crest courses along the top of the wave
				if jitter > 0.42 {
					col = ivoryFoam
				} else {
					col = paleIce
				}
			case math.Abs(nx-0.72) < 0.16 && phase > 0.38 && jitter > 0.48:
				// Warm crimson/gold sun reflection on the water under the sun
				if jitter > 0.76 {
					col = vermilionRed
				} else {
					col = goldenAmber
				}
			case jitter < 0.11:
				col = deepNavy
			default:
				switch {
				case phase > 0.48:
					col = mintAqua
				case phase > 0.04:
					col = seafoamTeal
				case phase > -0.45:
					col = cobaltBlue
				default:
					col = prussianBlue
				}
			}

			shear := 0.36
			if wb%2 == 1 {
				shear = -0.28
			}
			rc.DrawMosaicBrick(p0, p1, brickThick*0.47, shear, scale, col, 1.0)
		}
	}

	// 5. Geometric Low-Poly Wave Claws & Foam Spray Shards at the Breaking Crest
	// Place 5 dramatic curling claw clusters along the overhanging lip of the Great Wave
	clawAnchors := []struct {
		x, y, ang, size float64
	}{
		{0.34 * w, 0.21 * h, -0.18, 42.0 * scale},
		{0.41 * w, 0.23 * h, 0.12, 48.0 * scale},
		{0.47 * w, 0.28 * h, 0.38, 44.0 * scale},
		{0.52 * w, 0.35 * h, 0.58, 36.0 * scale},
		{0.26 * w, 0.23 * h, -0.45, 34.0 * scale},
	}

	foamPalette := []color.RGBA{ivoryFoam, paleIce, mintAqua, warmCream}
	for idx, cl := range clawAnchors {
		tDir := render.Pt(math.Cos(cl.ang), math.Sin(cl.ang))
		nDir := tDir.Perp()
		tip := render.Pt(cl.x, cl.y).Add(tDir.Mul(cl.size))
		baseTop := render.Pt(cl.x, cl.y).Add(nDir.Mul(cl.size * 0.36))
		baseBot := render.Pt(cl.x, cl.y).Sub(nDir.Mul(cl.size * 0.36))
		midPt := render.Pt(cl.x, cl.y).Add(tDir.Mul(cl.size * 0.42)).Add(nDir.Mul(cl.size * 0.08))
		hookTip := tip.Sub(nDir.Mul(cl.size * 0.32)).Add(tDir.Mul(cl.size * 0.22))

		// Dark navy backing for crisp grout around the foam claw shards
		backing := render.InsetPolygon([]render.Point{baseTop, tip, hookTip, baseBot}, -2.5*scale)
		rc.FillPolygon(backing, deepNavy, 1.0)

		shards := [][]render.Point{
			{baseTop, midPt, baseBot},
			{baseTop, tip, midPt},
			{midPt, tip, hookTip},
			{baseBot, midPt, hookTip},
		}
		for sIdx, sh := range shards {
			insetSh := render.InsetPolygon(sh, 0.65*scale)
			rc.FillPolygon(insetSh, foamPalette[(idx+sIdx)%len(foamPalette)], 1.0)
		}
	}

	c.MergeRenderCanvas(rc)
}
