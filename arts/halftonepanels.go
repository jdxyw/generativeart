package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type halftonePanels struct {
	panelCount int
	dotSpacing float64
}

// NewHalftonePanels returns a halftonePanels engine that renders a retro sci-fi
// comic-book panel layout with mint-green and paper-white Ben-Day halftone alien
// mountain horizons, ringed Saturns, crescent moons, and starry skies.
func NewHalftonePanels(panelCount int, dotSpacing float64) *halftonePanels {
	if panelCount <= 0 {
		panelCount = 22
	}
	if dotSpacing <= 0 {
		dotSpacing = 5.5
	}
	return &halftonePanels{
		panelCount: panelCount,
		dotSpacing: dotSpacing,
	}
}

type comicPanelSpec struct {
	x0, y0, x1, y1 float64
	celestialKind  int // 0: hero Saturn+Moon, 1: corner Saturn, 2: ring moon, 3: glowing full moon, 4: mini Saturn, 5: crescent orb, 6: stars only
	ridgeLayers    int
	yTopFrac       float64
	yBotFrac       float64
	dotScale       float64
}

// Generative renders the retro sci-fi Ben-Day halftone comic panels.
func (hp *halftonePanels) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	fw, fh := float64(c.Width()), float64(c.Height())
	scale := math.Min(fw, fh) / 600.0

	// Strict 3-Tone Retro Sci-Fi Print Palette
	paperWhite := color.RGBA{R: 244, G: 247, B: 244, A: 255} // #f4f7f4
	inkBlack := color.RGBA{R: 11, G: 14, B: 13, A: 255}      // #0b0e0d
	retroMint := color.RGBA{R: 78, G: 232, B: 164, A: 255}   // #4ee8a4
	midMint := color.RGBA{R: 54, G: 184, B: 130, A: 255}     // #36b882
	paleMint := color.RGBA{R: 164, G: 245, B: 207, A: 255}   // #a4f5cf

	rc.FillBackground(paperWhite)
	rc.SetBlendMode(render.BlendOver)

	g := 6.5 * scale
	b := 7.0 * scale
	hg := 0.5 * g

	// Build the iconic multi-panel comic page layout matching Reference Image 6
	panels := make([]comicPanelSpec, 0, 24)

	xSplitMain := 0.525 * fw
	xTopMid := 0.718 * fw
	yTopSplit := 0.308 * fh
	yMidSplit := 0.575 * fh

	// 1. Top-Left Hero Panel (Ringed Saturn + Dark Crescent Moon + 4 bold halftone mountain ridges)
	panels = append(panels, comicPanelSpec{
		x0: b, y0: b, x1: xSplitMain - hg, y1: yMidSplit - hg,
		celestialKind: 0, ridgeLayers: 4, yTopFrac: 0.44, yBotFrac: 0.88, dotScale: 1.0,
	})
	// 2. Top-Middle Portrait Panel (Partial ringed planet at top-right + 6 dense stacked halftone mountain ridges)
	panels = append(panels, comicPanelSpec{
		x0: xSplitMain + hg, y0: b, x1: xTopMid - hg, y1: yTopSplit - hg,
		celestialKind: 1, ridgeLayers: 6, yTopFrac: 0.30, yBotFrac: 0.90, dotScale: 0.82,
	})
	// 3. Top-Right Portrait Panel (Starry sky + 4 mountain ridges)
	panels = append(panels, comicPanelSpec{
		x0: xTopMid + hg, y0: b, x1: fw - b, y1: yTopSplit - hg,
		celestialKind: 6, ridgeLayers: 4, yTopFrac: 0.42, yBotFrac: 0.88, dotScale: 0.85,
	})
	// 4. Mid-Right Wide Panel (Deep starry sky, white ring moon at left, low halftone sea/island horizon)
	panels = append(panels, comicPanelSpec{
		x0: xSplitMain + hg, y0: yTopSplit + hg, x1: fw - b, y1: yMidSplit - hg,
		celestialKind: 2, ridgeLayers: 3, yTopFrac: 0.68, yBotFrac: 0.92, dotScale: 0.86,
	})

	// Bottom Half Columns (y: 0.575*h + g/2 .. h - b)
	c1Right := 0.245 * fw
	c2Right := 0.525 * fw
	c3Right := 0.745 * fw
	yBot0 := yMidSplit + hg
	yBot1 := fh - b

	// Col 1: top landscape panel + 2 bottom vertical strips
	yCol1Mid := 0.772 * fh
	xCol1Sub := b + 0.5*(c1Right-hg-b)
	panels = append(panels,
		comicPanelSpec{
			x0: b, y0: yBot0, x1: c1Right - hg, y1: yCol1Mid - hg,
			celestialKind: 6, ridgeLayers: 4, yTopFrac: 0.36, yBotFrac: 0.88, dotScale: 0.82,
		},
		comicPanelSpec{
			x0: b, y0: yCol1Mid + hg, x1: xCol1Sub - hg*0.7, y1: yBot1,
			celestialKind: 3, ridgeLayers: 4, yTopFrac: 0.34, yBotFrac: 0.88, dotScale: 0.74,
		},
		comicPanelSpec{
			x0: xCol1Sub + hg*0.7, y0: yCol1Mid + hg, x1: c1Right - hg, y1: yBot1,
			celestialKind: 5, ridgeLayers: 4, yTopFrac: 0.38, yBotFrac: 0.88, dotScale: 0.74,
		},
	)

	// Col 2: 2x4 sub-grid of 8 small cinematic panels
	col2X0 := c1Right + hg
	col2X1 := c2Right - hg
	col2MidX := 0.5 * (col2X0 + col2X1)
	col2RowCuts := [5]float64{
		yBot0,
		yBot0 + 0.26*(yBot1-yBot0),
		yBot0 + 0.51*(yBot1-yBot0),
		yBot0 + 0.76*(yBot1-yBot0),
		yBot1,
	}
	subKinds := [8]int{3, 5, 4, 3, 2, 6, 4, 3}
	for r := 0; r < 4; r++ {
		ry0 := col2RowCuts[r]
		if r > 0 {
			ry0 += hg * 0.75
		}
		ry1 := col2RowCuts[r+1]
		if r < 3 {
			ry1 -= hg * 0.75
		}
		for sc := 0; sc < 2; sc++ {
			sx0 := col2X0
			sx1 := col2MidX - hg*0.75
			if sc == 1 {
				sx0 = col2MidX + hg*0.75
				sx1 = col2X1
			}
			idx := r*2 + sc
			panels = append(panels, comicPanelSpec{
				x0: sx0, y0: ry0, x1: sx1, y1: ry1,
				celestialKind: subKinds[idx], ridgeLayers: 3, yTopFrac: 0.44, yBotFrac: 0.88, dotScale: 0.72,
			})
		}
	}

	// Col 3: top narrow strip, mid landscape, bottom split into 2 vertical strips
	col3X0 := c2Right + hg
	col3X1 := c3Right - hg
	yCol3Cut1 := 0.672 * fh
	yCol3Cut2 := 0.788 * fh
	col3MidX := 0.5 * (col3X0 + col3X1)
	panels = append(panels,
		comicPanelSpec{
			x0: col3X0, y0: yBot0, x1: col3X1, y1: yCol3Cut1 - hg*0.8,
			celestialKind: 6, ridgeLayers: 2, yTopFrac: 0.42, yBotFrac: 0.86, dotScale: 0.76,
		},
		comicPanelSpec{
			x0: col3X0, y0: yCol3Cut1 + hg*0.8, x1: col3X1, y1: yCol3Cut2 - hg*0.8,
			celestialKind: 3, ridgeLayers: 3, yTopFrac: 0.40, yBotFrac: 0.88, dotScale: 0.78,
		},
		comicPanelSpec{
			x0: col3X0, y0: yCol3Cut2 + hg*0.8, x1: col3MidX - hg*0.7, y1: yBot1,
			celestialKind: 5, ridgeLayers: 4, yTopFrac: 0.36, yBotFrac: 0.88, dotScale: 0.74,
		},
		comicPanelSpec{
			x0: col3MidX + hg*0.7, y0: yCol3Cut2 + hg*0.8, x1: col3X1, y1: yBot1,
			celestialKind: 4, ridgeLayers: 4, yTopFrac: 0.40, yBotFrac: 0.88, dotScale: 0.74,
		},
	)

	// Col 4: 4 stacked horizontal panoramic panels
	col4X0 := c3Right + hg
	col4X1 := fw - b
	col4Cuts := [5]float64{yBot0, 0.702 * fh, 0.788 * fh, 0.875 * fh, yBot1}
	col4Kinds := [4]int{6, 3, 2, 5}
	for r := 0; r < 4; r++ {
		ry0 := col4Cuts[r]
		if r > 0 {
			ry0 += hg * 0.8
		}
		ry1 := col4Cuts[r+1]
		if r < 3 {
			ry1 -= hg * 0.8
		}
		panels = append(panels, comicPanelSpec{
			x0: col4X0, y0: ry0, x1: col4X1, y1: ry1,
			celestialKind: col4Kinds[r], ridgeLayers: 3, yTopFrac: 0.38, yBotFrac: 0.86, dotScale: 0.76,
		})
	}

	if hp.panelCount > 0 && hp.panelCount < 22 && hp.panelCount < len(panels) {
		panels = panels[:hp.panelCount]
	}

	noise := common.NewNoiseEngine(1984)

	// Helper to draw a tilted Ringed Saturn with back rings, shadowed planet disc, crescent rim, and front rings
	drawRingedSaturn := func(cx, cy, planetR, tiltRad float64, ringCount int) {
		cosT := math.Cos(tiltRad)
		sinT := math.Sin(tiltRad)

		// Ring radii from 1.42*R to 2.28*R
		drawRingHalf := func(front bool) {
			for rIdx := 0; rIdx < ringCount; rIdx++ {
				u := float64(rIdx) / float64(common.MaxInt(1, ringCount-1))
				rx := planetR * (1.42 + 0.86*u)
				ry := 0.29 * rx
				col := retroMint
				if rIdx%2 == 1 {
					col = paperWhite
				}
				sw := (1.15 + 0.85*(1.0-math.Abs(2.0*u-1.0))) * scale

				const steps = 44
				pts := make([]render.Point, 0, steps+1)
				tStart := math.Pi
				tEnd := 2.0 * math.Pi
				if front {
					tStart = 0.0
					tEnd = math.Pi
				}
				for s := 0; s <= steps; s++ {
					th := tStart + (tEnd-tStart)*(float64(s)/float64(steps))
					lx := rx * math.Cos(th)
					ly := ry * math.Sin(th)
					wx := cx + lx*cosT - ly*sinT
					wy := cy + lx*sinT + ly*cosT
					pts = append(pts, render.Pt(wx, wy))
				}
				rc.StrokePolyline(pts, sw, col, 0.96)
			}
		}

		// 1. Back half of concentric elliptical rings
		drawRingHalf(false)

		// 2. Solid ink-black planet sphere + crisp mint crescent rim on upper-right limb
		rc.DrawGlowDot(cx, cy, planetR+1.2*scale, inkBlack, 1.0, 0.98)

		// Draw mint crescent highlight along the right limb using offset circle Difference
		cresPts := make([]render.Point, 0, 64)
		for i := 0; i <= 30; i++ {
			ang := -0.45*math.Pi + (float64(i)/30.0)*1.05*math.Pi
			cresPts = append(cresPts, render.Pt(cx+planetR*0.95*math.Cos(ang), cy+planetR*0.95*math.Sin(ang)))
		}
		for i := 30; i >= 0; i-- {
			ang := -0.45*math.Pi + (float64(i)/30.0)*1.05*math.Pi
			cresPts = append(cresPts, render.Pt(cx-planetR*0.10+planetR*0.88*math.Cos(ang), cy+planetR*0.05+planetR*0.88*math.Sin(ang)))
		}
		rc.FillPolygon(cresPts, retroMint, 0.96)
		rc.DrawSoftRing(cx, cy, planetR, 1.4*scale, inkBlack, 1.0)

		// 3. Front half of concentric elliptical rings with thin black separator shadow across planet face
		for rIdx := 0; rIdx < ringCount; rIdx++ {
			u := float64(rIdx) / float64(common.MaxInt(1, ringCount-1))
			rx := planetR * (1.42 + 0.86*u)
			ry := 0.29 * rx
			pts := make([]render.Point, 0, 33)
			for s := 0; s <= 32; s++ {
				th := (float64(s) / 32.0) * math.Pi
				lx := rx * math.Cos(th)
				ly := ry * math.Sin(th)
				pts = append(pts, render.Pt(cx+lx*cosT-ly*sinT, cy+lx*sinT+ly*cosT))
			}
			rc.StrokePolyline(pts, 2.8*scale, inkBlack, 0.92)
		}
		drawRingHalf(true)
	}

	// Helper to draw a dark sphere with a crisp retro-mint crescent rim and halftone terminator dots
	drawCrescentMoon := func(cx, cy, r, rimAngle float64) {
		rc.DrawGlowDot(cx, cy, r+1.0*scale, inkBlack, 1.0, 0.98)
		rc.DrawSoftRing(cx, cy, r*0.96, 1.4*scale, retroMint, 0.95)

		// Crescent limb polygon centered around rimAngle
		cres := make([]render.Point, 0, 64)
		span := 0.72 * math.Pi
		for i := 0; i <= 28; i++ {
			ang := rimAngle - span + (float64(i)/28.0)*(2.0*span)
			cres = append(cres, render.Pt(cx+r*0.96*math.Cos(ang), cy+r*0.96*math.Sin(ang)))
		}
		offX := -r * 0.16 * math.Cos(rimAngle)
		offY := -r * 0.16 * math.Sin(rimAngle)
		for i := 28; i >= 0; i-- {
			ang := rimAngle - span + (float64(i)/28.0)*(2.0*span)
			cres = append(cres, render.Pt(cx+offX+r*0.86*math.Cos(ang), cy+offY+r*0.86*math.Sin(ang)))
		}
		rc.FillPolygon(cres, retroMint, 0.96)
	}

	// Render each panel
	for pIdx, p := range panels {
		pw := p.x1 - p.x0
		ph := p.y1 - p.y0
		if pw <= 4 || ph <= 4 {
			continue
		}

		rectPoly := []render.Point{
			render.Pt(p.x0, p.y0),
			render.Pt(p.x1, p.y0),
			render.Pt(p.x1, p.y1),
			render.Pt(p.x0, p.y1),
		}

		rc.PushClipPolygon(rectPoly)
		rc.FillPolygon(rectPoly, inkBlack, 1.0)

		rng := rand.New(rand.NewSource(int64(pIdx*918273 + 1337)))

		// 1. Starfield & Lower Sky Stipple Screen
		skyH := ph * (p.yTopFrac + 0.12)
		numStars := common.MaxInt(8, int((pw*skyH)/(210.0*scale*scale)))
		for s := 0; s < numStars; s++ {
			sx := p.x0 + rng.Float64()*pw
			sy := p.y0 + rng.Float64()*skyH
			sr := (0.55 + 0.85*math.Pow(rng.Float64(), 1.6)) * scale
			sCol := paperWhite
			if rng.Float64() < 0.28 {
				sCol = retroMint
			}
			rc.DrawGlowDot(sx, sy, sr, sCol, 0.95, 0.92)

			// Occasional 4-point comic star twinkle in larger panels
			if pw > 110*scale && rng.Float64() < 0.045 {
				arm := sr * 3.2
				rc.StrokePolyline([]render.Point{render.Pt(sx-arm, sy), render.Pt(sx+arm, sy)}, 0.7*scale, sCol, 0.9)
				rc.StrokePolyline([]render.Point{render.Pt(sx, sy-arm), render.Pt(sx, sy+arm)}, 0.7*scale, sCol, 0.9)
			}
		}

		// Regular hexagonal stipple dot band above the highest mountain horizon
		stippleTop := p.y0 + ph*math.Max(0.14, p.yTopFrac-0.18)
		stippleBot := p.y0 + ph*math.Min(0.85, p.yTopFrac+0.12)
		stippleSp := 3.8 * scale
		stippleRowH := stippleSp * 0.8660254
		rowIdx := 0
		for gy := stippleTop; gy <= stippleBot; gy += stippleRowH {
			v := (gy - stippleTop) / math.Max(1.0, stippleBot-stippleTop)
			dotR := (0.28 + 0.58*v) * scale
			rowOff := 0.0
			if rowIdx&1 == 1 {
				rowOff = 0.5 * stippleSp
			}
			col := midMint
			if v > 0.62 {
				col = paleMint
			}
			for gx := p.x0 + rowOff; gx <= p.x1; gx += stippleSp {
				rc.DrawGlowDot(gx, gy, dotR, col, 0.85, 0.95)
			}
			rowIdx++
		}

		// 2. Celestial Bodies per Panel Type
		switch p.celestialKind {
		case 0:
			// Hero Panel: Left Crescent-Rimmed Moon + Right Tilted Ringed Saturn
			drawCrescentMoon(p.x0+0.235*pw, p.y0+0.425*ph, 0.168*pw, -0.18*math.Pi)
			drawRingedSaturn(p.x0+0.795*pw, p.y0+0.285*ph, 0.182*pw, -54.0*math.Pi/180.0, 6)

		case 1:
			// Top-Middle Portrait Panel: Partial Ringed Planet at top-right corner
			drawRingedSaturn(p.x0+0.92*pw, p.y0+0.12*ph, 0.24*pw, -38.0*math.Pi/180.0, 4)

		case 2:
			// White/Mint Ring Moon in sky
			mx := p.x0 + (0.18+0.12*rng.Float64())*pw
			my := p.y0 + 0.34*ph
			mr := math.Min(pw, ph) * 0.16
			rc.DrawGlowDot(mx, my, mr, inkBlack, 1.0, 0.98)
			rc.DrawSoftRing(mx, my, mr, 1.6*scale, paperWhite, 0.98)
			rc.DrawSoftRing(mx, my, mr*0.82, 0.9*scale, retroMint, 0.88)

		case 3:
			// Glowing Halftone Full Moon near the mountain crest
			mx := p.x0 + (0.28+0.44*rng.Float64())*pw
			my := p.y0 + (p.yTopFrac-0.06)*ph
			mr := math.Min(pw, ph) * 0.22
			rc.DrawGlowDot(mx, my, mr, paperWhite, 0.98, 0.98)
			rc.DrawGlowDot(mx, my, mr*0.86, paleMint, 0.75, 0.65)
			rc.DrawSoftRing(mx, my, mr, 1.4*scale, inkBlack, 0.98)

		case 4:
			// Mini Tilted Ringed Saturn
			mx := p.x0 + (0.35+0.30*rng.Float64())*pw
			my := p.y0 + 0.27*ph
			mr := math.Min(pw, ph) * 0.17
			tilt := (-35.0 + 70.0*rng.Float64()) * math.Pi / 180.0
			drawRingedSaturn(mx, my, mr, tilt, 3)

		case 5:
			// Dark Crescent-Rimmed Planet
			mx := p.x0 + (0.32+0.36*rng.Float64())*pw
			my := p.y0 + 0.28*ph
			mr := math.Min(pw, ph) * 0.18
			drawCrescentMoon(mx, my, mr, -0.22*math.Pi)
		}

		// 3. Layered Ben-Day Halftone Mountain Ridges (back-to-front)
		nLayers := p.ridgeLayers
		const xSamples = 48
		crestX := make([]float64, xSamples)
		for i := 0; i < xSamples; i++ {
			u := float64(i) / float64(xSamples-1)
			crestX[i] = p.x0 - 2.0 + u*(pw+4.0)
		}

		for l := 0; l < nLayers; l++ {
			layerT := 0.0
			if nLayers > 1 {
				layerT = float64(l) / float64(nLayers-1)
			}
			yBase := p.y0 + ph*(p.yTopFrac+layerT*(p.yBotFrac-p.yTopFrac))
			amp := ph * (0.15 - 0.04*layerT)
			if p.celestialKind == 2 {
				// Low sea/island horizon in mid-right panel
				amp = ph * (0.07 - 0.02*layerT)
			}

			// Piecewise-angular + ridged FBM mountain crest silhouette
			peakU1 := 0.18 + 0.28*rng.Float64()
			peakU2 := 0.58 + 0.26*rng.Float64()
			if pIdx == 0 {
				// Match Hero Panel mountain peaks from Reference Image 6
				switch l {
				case 0:
					peakU1, peakU2 = 0.62, 0.88
				case 1:
					peakU1, peakU2 = 0.26, 0.72
				case 2:
					peakU1, peakU2 = 0.44, 0.82
				default:
					peakU1, peakU2 = 0.18, 0.58
				}
			}

			crestY := make([]float64, xSamples)
			crestPts := make([]render.Point, xSamples)
			minCrestY := p.y1
			for i := 0; i < xSamples; i++ {
				u := float64(i) / float64(xSamples-1)
				tent1 := math.Max(0.0, 1.0-math.Abs(u-peakU1)*2.6)
				tent2 := math.Max(0.0, 1.0-math.Abs(u-peakU2)*3.0) * 0.78
				tent3 := math.Max(0.0, 1.0-math.Abs(u-0.5*(peakU1+peakU2)+0.18)*3.8) * 0.52
				ridgeN := noise.Ridged2D(u*2.2+float64(pIdx)*1.7+float64(l)*2.3, float64(l)*1.9+float64(pIdx)*0.7, 2, 2.0, 0.45)

				elev := (0.78*math.Max(tent1, math.Max(tent2, tent3)) + 0.22*ridgeN) * amp
				cy := yBase - elev
				cy = common.Constrain(cy, p.y0+3.0*scale, p.y1-2.0*scale)
				crestY[i] = cy
				crestPts[i] = render.Pt(crestX[i], cy)
				if cy < minCrestY {
					minCrestY = cy
				}
			}

			// Thick black separator halo behind the crest to separate cleanly from background ridges
			rc.StrokePolyline(crestPts, 4.0*scale, inkBlack, 1.0)

			// Build closed ridge polygon down to panel bottom
			ridgePoly := make([]render.Point, 0, xSamples+2)
			ridgePoly = append(ridgePoly, crestPts...)
			ridgePoly = append(ridgePoly, render.Pt(p.x1+2.0, p.y1+2.0), render.Pt(p.x0-2.0, p.y1+2.0))

			rc.PushClipPolygon(ridgePoly)

			// Fill ridge with smooth vertical gradient bands from Mint Green (near crest) to Paper White (below)
			bandEnd := p.y1 + 2.0
			ridgeSpan := math.Max(8.0*scale, (bandEnd-minCrestY)*0.78)
			const gradBands = 14
			for gb := 0; gb < gradBands; gb++ {
				u0 := float64(gb) / float64(gradBands)
				u1 := float64(gb+1) / float64(gradBands)
				um := 0.5 * (u0 + u1)

				// Alternate subtle mint/white richness across successive ridges
				var bandCol color.RGBA
				if (l+pIdx)%3 == 1 {
					bandCol = render.LerpOklab(paleMint, paperWhite, math.Pow(um, 0.75))
				} else {
					bandCol = render.LerpOklab(retroMint, paperWhite, math.Pow(um, 0.82))
				}

				polyBand := make([]render.Point, 0, xSamples*2)
				for i := 0; i < xSamples; i++ {
					y0 := crestY[i] + u0*ridgeSpan
					if gb == gradBands-1 {
						y0 = crestY[i] + u0*ridgeSpan
					}
					polyBand = append(polyBand, render.Pt(crestX[i], y0))
				}
				for i := xSamples - 1; i >= 0; i-- {
					y1 := crestY[i] + u1*ridgeSpan
					if gb == gradBands-1 {
						y1 = bandEnd
					}
					polyBand = append(polyBand, render.Pt(crestX[i], y1))
				}
				rc.FillPolygon(polyBand, bandCol, 1.0)
			}

			// Overlay Hexagonal Ben-Day Halftone Dot Screen inside the clipped mountain ridge
			sp := hp.dotSpacing * p.dotScale * scale
			if sp < 2.4*scale {
				sp = 2.4 * scale
			}
			rowStep := sp * 0.8660254
			invertHalftone := (l%2 == 1 && nLayers >= 4)

			rIdx := 0
			for gy := minCrestY - sp; gy <= p.y1+sp; gy += rowStep {
				xOff := 0.0
				if rIdx&1 == 1 {
					xOff = 0.5 * sp
				}
				for gx := p.x0 - sp + xOff; gx <= p.x1+sp; gx += sp {
					// Interpolate crestY at gx
					u := common.Clamp01((gx - crestX[0]) / math.Max(1e-5, crestX[xSamples-1]-crestX[0]))
					fIdx := u * float64(xSamples-1)
					i0 := common.ConstrainInt(int(math.Floor(fIdx)), 0, xSamples-2)
					frac := fIdx - float64(i0)
					yCrestAtGx := crestY[i0]*(1.0-frac) + crestY[i0+1]*frac

					if gy < yCrestAtGx-sp {
						continue
					}
					dY := common.Clamp01((gy - yCrestAtGx) / math.Max(6.0*scale, ridgeSpan*0.85))
					var dotFrac float64
					if invertHalftone {
						dotFrac = common.Constrain(0.14+0.34*dY, 0.10, 0.46)
					} else {
						dotFrac = common.Constrain(0.48-0.36*dY, 0.09, 0.47)
					}
					rDot := sp * dotFrac
					rc.DrawGlowDot(gx, gy, rDot, inkBlack, 1.0, 0.98)
				}
				rIdx++
			}

			rc.PopClip()

			// Crisp comic ink crest outline
			rc.StrokePolyline(crestPts, 2.2*scale, inkBlack, 1.0)
		}

		rc.PopClip()

		// Crisp black inner panel border around every comic panel rectangle
		borderLoop := []render.Point{
			render.Pt(p.x0, p.y0),
			render.Pt(p.x1, p.y0),
			render.Pt(p.x1, p.y1),
			render.Pt(p.x0, p.y1),
			render.Pt(p.x0, p.y0),
		}
		rc.StrokePolyline(borderLoop, 1.85*scale, inkBlack, 1.0)
	}

	// Subtle vintage comic newsprint grain
	rc.ApplyPaperTexture(0.18, 1977)
	c.MergeRenderCanvas(rc)
}
