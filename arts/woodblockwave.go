package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type woodblockWave struct {
	ridgeLayers int
	waveLines   int
}

// NewWoodblockWave returns a woodblockWave engine that renders a polychrome
// Ukiyo-e woodblock print (锦绘版画) with bokashi ink gradation, carved woodgrain
// hatch lines, a vermilion sun with stylized clouds, Mount Fuji, a wooden sail ship,
// and rolling engraved sea swells.
func NewWoodblockWave(ridgeLayers, waveLines int) *woodblockWave {
	if ridgeLayers <= 0 {
		ridgeLayers = 5
	}
	if waveLines <= 0 {
		waveLines = 42
	}
	return &woodblockWave{
		ridgeLayers: ridgeLayers,
		waveLines:   waveLines,
	}
}

// Generative renders the Ukiyo-e polychrome woodblock print and sumi-e keyblock outlines.
func (ww *woodblockWave) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0

	// Traditional Nishiki-e mineral & vegetable pigment palette
	washiCream := color.RGBA{R: 244, G: 234, B: 208, A: 255}
	apricotDawn := color.RGBA{R: 238, G: 176, B: 114, A: 255}
	vermilionSun := color.RGBA{R: 214, G: 64, B: 44, A: 255}
	crimsonSeal := color.RGBA{R: 182, G: 38, B: 34, A: 255}

	beroPrussian := color.RGBA{R: 20, G: 52, B: 102, A: 255}
	indigoBlue := color.RGBA{R: 36, G: 92, B: 146, A: 255}
	mineralTeal := color.RGBA{R: 72, G: 148, B: 156, A: 255}
	celadonMist := color.RGBA{R: 168, G: 208, B: 192, A: 255}

	ochreWood := color.RGBA{R: 192, G: 134, B: 72, A: 255}
	siennaBark := color.RGBA{R: 124, G: 74, B: 44, A: 255}
	sumiInk := color.RGBA{R: 26, G: 24, B: 28, A: 255}

	rc.FillBackground(washiCream)
	rc.SetBlendMode(render.BlendOver)

	// 1. Seamless Full-Width Sky Bokashi & Woodgrain Gradation (top Prussian Blue zenith -> apricot horizon)
	skyRows := int(165 * scale)
	skyMaxY := 0.52 * h
	for r := 0; r < skyRows; r++ {
		s := float64(r) / float64(skyRows-1)
		yBase := s * skyMaxY
		grain := 0.035 * math.Sin(s*110.0)
		u := common.Constrain(s+grain, 0, 1)

		var col color.RGBA
		switch {
		case u < 0.38:
			col = render.LerpOklab(beroPrussian, indigoBlue, u/0.38)
		case u < 0.64:
			col = render.LerpOklab(indigoBlue, celadonMist, (u-0.38)/0.26)
		case u < 0.82:
			col = render.LerpOklab(celadonMist, washiCream, (u-0.64)/0.18)
		default:
			col = render.LerpOklab(washiCream, apricotDawn, (u-0.82)/0.18)
		}

		p := &canvas.Path{}
		p.MoveTo(0, yBase)
		p.LineTo(w, yBase)
		rc.StrokePath(p, (skyMaxY/float64(skyRows))*1.25, canvas.ButtCap, canvas.MiterJoin, col, 0.94)
	}

	// 2. Carved Woodgrain Vermilion Sun Disc at Upper-Left (0.28w, 0.25h)
	sunCX, sunCY, rSun := 0.28*w, 0.25*h, 0.155*w
	sunRim := make([]render.Point, 56)
	for i := 0; i < 56; i++ {
		ang := (float64(i) / 56.0) * 2.0 * math.Pi
		sunRim[i] = render.Pt(sunCX+rSun*math.Cos(ang), sunCY+rSun*math.Sin(ang))
	}
	rc.FillPolygon(sunRim, vermilionSun, 0.97)

	// Horizontal mokume woodblock grain lines across the sun disc
	const sunGrainLines = 36
	for i := 1; i < sunGrainLines; i++ {
		u := (float64(i)/float64(sunGrainLines))*2.0 - 1.0
		dy := u * rSun * 0.96
		halfSpan := math.Sqrt(math.Max(0, rSun*rSun-dy*dy))
		p := &canvas.Path{}
		yBase := sunCY + dy
		p.MoveTo(sunCX-halfSpan, yBase)
		for k := 1; k <= 16; k++ {
			t := float64(k) / 16.0
			x := sunCX - halfSpan + t*2.0*halfSpan
			y := yBase + 1.8*scale*math.Sin(t*math.Pi*3.0+float64(i)*0.7)
			p.LineTo(x, y)
		}
		grainCol := apricotDawn
		if i%2 == 0 {
			grainCol = crimsonSeal
		}
		rc.StrokePath(p, 1.15*scale, canvas.RoundCap, canvas.RoundJoin, grainCol, 0.54)
	}

	// 3. Stylized Ukiyo-e Horizontal Cloud Bands Drifting Across the Sun & Sky
	drawUkiyoeCloudBand := func(xStart, xEnd, yCenter, halfH float64) {
		const samples = 44
		upper := make([]render.Point, samples)
		lower := make([]render.Point, samples)
		for i := 0; i < samples; i++ {
			t := float64(i) / float64(samples-1)
			x := xStart + t*(xEnd-xStart)
			env := math.Sin(t * math.Pi)
			billow := (0.65 + 0.35*math.Sin(t*math.Pi*5.0)) * halfH * env
			upper[i] = render.Pt(x, yCenter-billow)
			lower[i] = render.Pt(x, yCenter+billow*0.42)
		}
		poly := make([]render.Point, 0, samples*2)
		poly = append(poly, upper...)
		for i := samples - 1; i >= 0; i-- {
			poly = append(poly, lower[i])
		}
		rc.FillPolygon(poly, washiCream, 0.96)
		rc.DrawHatchPatch(upper, lower, 7, 1.0*scale, 0.0, func(s, t float64) color.RGBA {
			return render.LerpOklab(apricotDawn, washiCream, s)
		}, 0.72)

		outline := &canvas.Path{}
		outline.MoveTo(poly[0].X, poly[0].Y)
		for i := 1; i < len(poly); i++ {
			outline.LineTo(poly[i].X, poly[i].Y)
		}
		outline.Close()
		rc.StrokePath(outline, 1.35*scale, canvas.RoundCap, canvas.RoundJoin, sumiInk, 0.88)
	}
	drawUkiyoeCloudBand(0.05*w, 0.43*w, 0.20*h, 12.5*scale)
	drawUkiyoeCloudBand(0.20*w, 0.58*w, 0.30*h, 13.5*scale)

	// 4. Distant Mount Fuji with Engraved Snow Ridges & Indigo Bokashi Slopes (Right Horizon)
	const fujiSamples = 64
	fujiTop := make([]render.Point, fujiSamples)
	fujiSnowLine := make([]render.Point, fujiSamples)
	fujiBase := make([]render.Point, fujiSamples)
	for i := 0; i < fujiSamples; i++ {
		t := float64(i) / float64(fujiSamples-1)
		x := 0.36*w + t*0.64*w
		distPeak := math.Abs(x - 0.73*w)
		var yTop float64
		if distPeak < 0.036*w {
			yTop = 0.30*h + 3.0*scale*math.Cos((x-0.73*w)/(0.012*w))
		} else {
			yTop = 0.30*h + 0.34*h*math.Pow((distPeak-0.036*w)/(0.36*w), 1.28)
		}
		yTop = math.Min(yTop, 0.62*h)
		fujiTop[i] = render.Pt(x, yTop)

		snowDrop := 0.088*h*math.Exp(-math.Pow((x-0.73*w)/(0.12*w), 2.0)) +
			14.0*scale*math.Abs(math.Sin(t*math.Pi*9.0))*math.Exp(-math.Pow((x-0.73*w)/(0.15*w), 2.0))
		fujiSnowLine[i] = render.Pt(x, math.Min(0.62*h, yTop+snowDrop))
		fujiBase[i] = render.Pt(x, 0.62*h)
	}

	mtnPoly := make([]render.Point, 0, fujiSamples*2)
	mtnPoly = append(mtnPoly, fujiTop...)
	for i := fujiSamples - 1; i >= 0; i-- {
		mtnPoly = append(mtnPoly, fujiBase[i])
	}
	rc.FillPolygon(mtnPoly, indigoBlue, 1.0)
	rc.DrawHatchPatch(fujiSnowLine, fujiBase, 58, 1.75*scale, 0.0, func(s, t float64) color.RGBA {
		if s < 0.45 {
			return render.LerpOklab(beroPrussian, indigoBlue, s/0.45)
		}
		return render.LerpOklab(indigoBlue, celadonMist, (s-0.45)/0.55)
	}, 0.96)

	snowPoly := make([]render.Point, 0, fujiSamples*2)
	snowPoly = append(snowPoly, fujiTop...)
	for i := fujiSamples - 1; i >= 0; i-- {
		snowPoly = append(snowPoly, fujiSnowLine[i])
	}
	rc.FillPolygon(snowPoly, washiCream, 1.0)

	fujiKey := &canvas.Path{}
	fujiKey.MoveTo(fujiTop[0].X, fujiTop[0].Y)
	for i := 1; i < fujiSamples; i++ {
		fujiKey.LineTo(fujiTop[i].X, fujiTop[i].Y)
	}
	for i := 8; i < fujiSamples-8; i++ {
		if i == 8 {
			fujiKey.MoveTo(fujiSnowLine[i].X, fujiSnowLine[i].Y)
		} else {
			fujiKey.LineTo(fujiSnowLine[i].X, fujiSnowLine[i].Y)
		}
	}
	rc.StrokePath(fujiKey, 1.85*scale, canvas.RoundCap, canvas.RoundJoin, sumiInk, 0.95)

	// 5. Traditional Japanese Wooden Sail Ship (Bezaisen) Riding the Midground Swell (x ~ 0.28w, y ~ 0.55h)
	shipX, shipY := 0.28*w, 0.55*h
	hullPts := []render.Point{
		render.Pt(shipX-56*scale, shipY-6*scale),
		render.Pt(shipX-38*scale, shipY+14*scale),
		render.Pt(shipX+40*scale, shipY+12*scale),
		render.Pt(shipX+66*scale, shipY-18*scale),
		render.Pt(shipX+30*scale, shipY-3*scale),
		render.Pt(shipX-28*scale, shipY-1*scale),
	}
	rc.FillPolygon(hullPts, ochreWood, 1.0)
	rc.DrawHatchPatch(
		[]render.Point{hullPts[0], hullPts[5], hullPts[4], hullPts[3]},
		[]render.Point{hullPts[0], hullPts[1], hullPts[2], hullPts[3]},
		8, 1.15*scale, 0.0,
		func(s, t float64) color.RGBA {
			return render.LerpOklab(ochreWood, siennaBark, s)
		},
		0.92,
	)

	// Billowing square cloth sail with horizontal bamboo battens (bow = 0.0 so battens stay strictly inside the sail)
	sailLeft := []render.Point{
		render.Pt(shipX-26*scale, shipY-96*scale),
		render.Pt(shipX-38*scale, shipY-52*scale),
		render.Pt(shipX-24*scale, shipY-8*scale),
	}
	sailRight := []render.Point{
		render.Pt(shipX+24*scale, shipY-90*scale),
		render.Pt(shipX+14*scale, shipY-46*scale),
		render.Pt(shipX+22*scale, shipY-6*scale),
	}
	sailL := render.CatmullRomSpline(sailLeft, 12)
	sailR := render.CatmullRomSpline(sailRight, 12)
	sailPoly := make([]render.Point, 0, len(sailL)+len(sailR))
	sailPoly = append(sailPoly, sailL...)
	for i := len(sailR) - 1; i >= 0; i-- {
		sailPoly = append(sailPoly, sailR[i])
	}
	rc.FillPolygon(sailPoly, washiCream, 1.0)

	// Draw horizontal bamboo battens across the billowing sail
	const battens = 7
	battenPath := &canvas.Path{}
	for b := 1; b < battens; b++ {
		idx := (b * (len(sailL) - 1)) / battens
		pL := sailL[idx]
		pR := sailR[idx]
		pMid := pL.Lerp(pR, 0.5).Add(render.Pt(-3.0*scale, 2.5*scale))
		battenPath.MoveTo(pL.X, pL.Y)
		battenPath.QuadTo(pMid.X, pMid.Y, pR.X, pR.Y)
	}
	rc.StrokePath(battenPath, 1.05*scale, canvas.RoundCap, canvas.RoundJoin, siennaBark, 0.78)

	shipKey := &canvas.Path{}
	shipKey.MoveTo(hullPts[0].X, hullPts[0].Y)
	for i := 1; i < len(hullPts); i++ {
		shipKey.LineTo(hullPts[i].X, hullPts[i].Y)
	}
	shipKey.Close()
	shipKey.MoveTo(shipX-2*scale, shipY-106*scale)
	shipKey.LineTo(shipX-2*scale, shipY-2*scale)
	shipKey.MoveTo(sailL[0].X, sailL[0].Y)
	for i := 1; i < len(sailL); i++ {
		shipKey.LineTo(sailL[i].X, sailL[i].Y)
	}
	shipKey.LineTo(sailR[len(sailR)-1].X, sailR[len(sailR)-1].Y)
	for i := len(sailR) - 2; i >= 0; i-- {
		shipKey.LineTo(sailR[i].X, sailR[i].Y)
	}
	shipKey.Close()
	rc.StrokePath(shipKey, 1.75*scale, canvas.RoundCap, canvas.RoundJoin, sumiInk, 0.95)

	// 6. Layered Wood-Engraved Sea Swells & Stylized Foam Scrolls (y in [0.52h, 1.0h])
	numSwells := ww.ridgeLayers
	if numSwells < 4 {
		numSwells = 5
	}
	linesPerSwell := ww.waveLines / numSwells
	if linesPerSwell < 12 {
		linesPerSwell = 14
	}

	for sw := 0; sw < numSwells; sw++ {
		swFrac := float64(sw) / float64(numSwells-1)
		baseY := 0.55*h + swFrac*0.38*h
		const wavePts = 76
		crestCurve := make([]render.Point, wavePts)
		troughCurve := make([]render.Point, wavePts)

		for i := 0; i < wavePts; i++ {
			t := float64(i) / float64(wavePts-1)
			x := -24.0*scale + t*(w+48.0*scale)
			phase := t*2.0*math.Pi*(1.45+0.15*float64(sw%2)) + float64(sw)*1.18
			yCrest := baseY - (30.0+12.0*swFrac)*scale*(math.Sin(phase)+0.34*math.Sin(2.0*phase+0.4))
			yTrough := yCrest + (78.0+20.0*swFrac)*scale
			crestCurve[i] = render.Pt(x, yCrest)
			troughCurve[i] = render.Pt(x, yTrough)
		}

		// Fill entire swell body with deep indigo/teal under-block so woodcut lines sit on rich pigment
		foamBot := make([]render.Point, wavePts)
		for i := 0; i < wavePts; i++ {
			foamBot[i] = crestCurve[i].Lerp(troughCurve[i], 0.18)
		}

		// White-and-celadon foam cap band
		foamPoly := make([]render.Point, 0, wavePts*2)
		foamPoly = append(foamPoly, crestCurve...)
		for i := wavePts - 1; i >= 0; i-- {
			foamPoly = append(foamPoly, foamBot[i])
		}
		rc.FillPolygon(foamPoly, washiCream, 1.0)

		// Deep indigo water body polygon
		waterPoly := make([]render.Point, 0, wavePts*2)
		waterPoly = append(waterPoly, foamBot...)
		for i := wavePts - 1; i >= 0; i-- {
			waterPoly = append(waterPoly, troughCurve[i])
		}
		rc.FillPolygon(waterPoly, beroPrussian, 1.0)

		// Carved light celadon/teal/washi wood-engraving stripes across the Prussian blue wave body
		rc.DrawHatchPatch(foamBot, troughCurve, linesPerSwell, 2.0*scale, 0.0, func(s, t float64) color.RGBA {
			if s < 0.32 {
				return render.LerpOklab(washiCream, celadonMist, s/0.32)
			}
			if s < 0.68 {
				return render.LerpOklab(celadonMist, mineralTeal, (s-0.32)/0.36)
			}
			return render.LerpOklab(mineralTeal, indigoBlue, (s-0.68)/0.32)
		}, 0.92)

		// Sumi-e keyblock crest outline & foam division line
		crestPath := &canvas.Path{}
		crestPath.MoveTo(crestCurve[0].X, crestCurve[0].Y)
		for i := 1; i < wavePts; i++ {
			crestPath.LineTo(crestCurve[i].X, crestCurve[i].Y)
		}
		rc.StrokePath(crestPath, 1.95*scale, canvas.RoundCap, canvas.RoundJoin, sumiInk, 0.96)

		// Carved scalloped foam curls along the crest
		for i := 4; i < wavePts-4; i += 5 {
			p0 := crestCurve[i-2]
			p1 := crestCurve[i]
			p2 := crestCurve[i+2]
			if p1.Y < baseY-4.0*scale {
				curl := &canvas.Path{}
				curl.MoveTo(p0.X, p0.Y+4.5*scale)
				curl.QuadTo(p1.X, p1.Y-8.5*scale, p2.X, p2.Y+4.5*scale)
				rc.StrokePath(curl, 1.45*scale, canvas.RoundCap, canvas.RoundJoin, sumiInk, 0.88)
			}
		}
	}

	// 7. Traditional Vermilion Artist Hanko Stamp (落款朱印) in Upper-Right Margin
	sealX, sealY, sealW, sealH := 0.89*w, 0.09*h, 25.0*scale, 54.0*scale
	sealRect := []render.Point{
		render.Pt(sealX, sealY),
		render.Pt(sealX+sealW, sealY),
		render.Pt(sealX+sealW, sealY+sealH),
		render.Pt(sealX, sealY+sealH),
	}
	rc.FillPolygon(sealRect, crimsonSeal, 0.94)
	sealGlyph := &canvas.Path{}
	sealGlyph.MoveTo(sealX+3*scale, sealY+3*scale)
	sealGlyph.LineTo(sealX+sealW-3*scale, sealY+3*scale)
	sealGlyph.LineTo(sealX+sealW-3*scale, sealY+sealH-3*scale)
	sealGlyph.LineTo(sealX+3*scale, sealY+sealH-3*scale)
	sealGlyph.Close()
	for row := 0; row < 3; row++ {
		ry := sealY + (11.5+float64(row)*14.5)*scale
		sealGlyph.MoveTo(sealX+6*scale, ry)
		sealGlyph.LineTo(sealX+sealW-6*scale, ry)
		sealGlyph.MoveTo(sealX+0.5*sealW, ry-4*scale)
		sealGlyph.LineTo(sealX+0.5*sealW, ry+5*scale)
	}
	rc.StrokePath(sealGlyph, 1.4*scale, canvas.RoundCap, canvas.RoundJoin, washiCream, 0.92)

	rc.ApplyPaperTexture(0.08, 202611)

	c.MergeRenderCanvas(rc)
}
