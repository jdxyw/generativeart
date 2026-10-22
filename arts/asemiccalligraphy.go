package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type asemicCalligraphy struct {
	cols         int
	glyphsPerCol int
}

// NewAsemicCalligraphy returns an asemicCalligraphy engine that renders an
// Eastern cursive ink scroll (狂草书法长卷) on ruled Xuan paper with spring-mass
// brush dynamics, flying-white (Feibai) bristle splits, and vermilion artist seals.
func NewAsemicCalligraphy(cols, glyphsPerCol int) *asemicCalligraphy {
	if cols <= 0 {
		cols = 6
	}
	if glyphsPerCol <= 0 {
		glyphsPerCol = 5
	}
	return &asemicCalligraphy{
		cols:         cols,
		glyphsPerCol: glyphsPerCol,
	}
}

// Generative renders the cursive ink scroll on ruled Xuan paper with vermilion seals.
func (ac *asemicCalligraphy) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0

	// Aged Xuan Paper, Pine-Soot Sumi Ink, Silk Brocade & Cinnabar Seal Palette
	xuanPaper := color.RGBA{R: 244, G: 236, B: 214, A: 255}
	brocadeGold := color.RGBA{R: 218, G: 194, B: 144, A: 255}
	zhisiRed := color.RGBA{R: 198, G: 96, B: 78, A: 255}

	sumiDeep := color.RGBA{R: 20, G: 18, B: 22, A: 255}
	sumiMid := color.RGBA{R: 48, G: 44, B: 50, A: 255}
	sumiWash := color.RGBA{R: 112, G: 108, B: 110, A: 255}
	cinnabarSeal := color.RGBA{R: 196, G: 42, B: 36, A: 255}

	rc.FillBackground(xuanPaper)
	rc.SetBlendMode(render.BlendOver)

	// 1. Top & Bottom Silk Brocade Scroll Borders + Ruled Zhisi (朱丝栏) Columns
	borderH := 28.0 * scale
	rc.FillPolygon([]render.Point{
		render.Pt(0, 0), render.Pt(w, 0), render.Pt(w, borderH), render.Pt(0, borderH),
	}, brocadeGold, 0.55)
	rc.FillPolygon([]render.Point{
		render.Pt(0, h-borderH), render.Pt(w, h-borderH), render.Pt(w, h), render.Pt(0, h),
	}, brocadeGold, 0.55)

	numCols := ac.cols
	if numCols < 3 {
		numCols = 6
	}
	marginX := 44.0 * scale
	marginY := 46.0 * scale
	innerW := w - 2.0*marginX
	innerH := h - 2.0*marginY
	colW := innerW / float64(numCols)

	rulePath := &canvas.Path{}
	// Outer double red scroll frame
	rulePath.MoveTo(marginX, marginY)
	rulePath.LineTo(w-marginX, marginY)
	rulePath.LineTo(w-marginX, h-marginY)
	rulePath.LineTo(marginX, h-marginY)
	rulePath.Close()

	for i := 1; i < numCols; i++ {
		x := marginX + float64(i)*colW
		rulePath.MoveTo(x, marginY)
		rulePath.LineTo(x, h-marginY)
	}
	rc.StrokePath(rulePath, 1.05*scale, canvas.RoundCap, canvas.RoundJoin, zhisiRed, 0.42)

	// 2. Spring-Mass Cursive Brush Dynamics Along Each Vertical Column (Right-to-Left Traditional Order)
	rng := rand.New(rand.NewSource(202620))
	glyphs := ac.glyphsPerCol
	if glyphs < 3 {
		glyphs = 5
	}
	cellH := innerH / float64(glyphs)

	for colIdx := numCols - 1; colIdx >= 0; colIdx-- {
		colCenterX := marginX + (float64(colIdx)+0.5)*colW
		var prevEnd render.Point
		hasPrev := false

		for gIdx := 0; gIdx < glyphs; gIdx++ {
			cellTopY := marginY + float64(gIdx)*cellH
			cellCenterY := cellTopY + 0.5*cellH

			// Construct a 5-to-7 control-point cursive glyph trajectory with loops, hooks, and sweeps
			nPts := 5 + rng.Intn(3)
			pts := make([]render.Point, 0, nPts)
			for p := 0; p < nPts; p++ {
				u := float64(p) / float64(nPts-1)
				// Horizontal swing (left-right cursive rhythm) + vertical progression
				swingX := (math.Sin(u*math.Pi*2.2+float64(gIdx*2+colIdx)*1.3) * 0.34 * colW) +
					(rng.Float64()-0.5)*0.16*colW
				progY := cellTopY + (0.14+0.72*u)*cellH + (rng.Float64()-0.5)*0.12*cellH
				pts = append(pts, render.Pt(colCenterX+swingX, progY))
			}

			// Connect consecutive characters in a column with a delicate cursive ligature (牵丝遊丝)
			if hasPrev && rng.Float64() < 0.72 {
				midLig := prevEnd.Lerp(pts[0], 0.5).Add(render.Pt((rng.Float64()-0.5)*12*scale, 0))
				rc.DrawTaperedSpline(
					[]render.Point{prevEnd, midLig, pts[0]},
					2.2*scale,
					render.TaperBothEnds,
					func(t float64) color.RGBA { return sumiMid },
					0.68,
				)
			}

			// Subtle diluted ink halo at heavy press-down start
			rc.DrawGlowDot(pts[0].X, pts[0].Y, (7.5+rng.Float64()*4.0)*scale, sumiWash, 0.16, 0.25)

			// Main calligraphic brush stroke (heavy press-down & modulated pressure)
			strokeW := (9.5 + rng.Float64()*5.5) * scale
			rc.DrawTaperedSpline(
				pts,
				strokeW,
				render.TaperCalligraphic,
				func(t float64) color.RGBA {
					if t < 0.65 {
						return sumiDeep
					}
					return render.LerpOklab(sumiDeep, sumiMid, (t-0.65)/0.35)
				},
				0.94,
			)

			// Flying-white (飞白) dry-brush bristle filaments along the curve
			rc.DrawBristleStroke(pts, strokeW*0.92, 11, sumiDeep, 0.58, rng)

			// Secondary crossing slash / dot accent (点画 / 撇捺)
			slashStart := render.Pt(
				colCenterX-(0.22+rng.Float64()*0.15)*colW,
				cellCenterY+(rng.Float64()-0.5)*0.22*cellH,
			)
			slashMid := render.Pt(
				colCenterX+(rng.Float64()-0.5)*0.12*colW,
				cellCenterY+0.08*cellH,
			)
			slashEnd := render.Pt(
				colCenterX+(0.24+rng.Float64()*0.16)*colW,
				cellCenterY+(0.16+rng.Float64()*0.18)*cellH,
			)
			rc.DrawTaperedSpline(
				[]render.Point{slashStart, slashMid, slashEnd},
				strokeW*0.78,
				render.TaperCalligraphic,
				func(t float64) color.RGBA { return sumiDeep },
				0.92,
			)

			// Expressive ink dot (侧点)
			if rng.Float64() < 0.78 {
				dotX := colCenterX + (0.26*(float64((gIdx+colIdx)%2)*2-1))*colW
				dotY := cellTopY + (0.18+rng.Float64()*0.62)*cellH
				dotEnd := render.Pt(dotX+5.5*scale, dotY+7.5*scale)
				rc.DrawTaperedSpline(
					[]render.Point{render.Pt(dotX, dotY), dotEnd},
					strokeW*0.68,
					render.TaperTail,
					func(t float64) color.RGBA { return sumiDeep },
					0.95,
				)
			}

			prevEnd = pts[len(pts)-1]
			hasPrev = true
		}
	}

	// 3. Carved Vermilion Cinnabar Seals (落款朱文印 + 引首章)
	drawCinnabarSeal := func(sx, sy, sw, sh float64, inverted bool) {
		rect := []render.Point{
			render.Pt(sx, sy),
			render.Pt(sx+sw, sy),
			render.Pt(sx+sw, sy+sh),
			render.Pt(sx, sy+sh),
		}
		if inverted {
			rc.FillPolygon(rect, cinnabarSeal, 0.92)
			glyph := &canvas.Path{}
			p := 3.0 * scale
			glyph.MoveTo(sx+p, sy+p)
			glyph.LineTo(sx+sw-p, sy+p)
			glyph.LineTo(sx+sw-p, sy+sh-p)
			glyph.LineTo(sx+p, sy+sh-p)
			glyph.Close()
			for r := 0; r < 2; r++ {
				ry := sy + (0.28+float64(r)*0.44)*sh
				glyph.MoveTo(sx+5*scale, ry)
				glyph.LineTo(sx+sw-5*scale, ry)
				glyph.MoveTo(sx+0.5*sw, ry-5*scale)
				glyph.LineTo(sx+0.5*sw, ry+5*scale)
			}
			rc.StrokePath(glyph, 1.5*scale, canvas.RoundCap, canvas.RoundJoin, xuanPaper, 0.92)
		} else {
			glyph := &canvas.Path{}
			glyph.MoveTo(sx, sy)
			glyph.LineTo(sx+sw, sy)
			glyph.LineTo(sx+sw, sy+sh)
			glyph.LineTo(sx, sy+sh)
			glyph.Close()
			glyph.MoveTo(sx+0.5*sw, sy+3*scale)
			glyph.LineTo(sx+0.5*sw, sy+sh-3*scale)
			glyph.MoveTo(sx+4*scale, sy+0.5*sh)
			glyph.LineTo(sx+sw-4*scale, sy+0.5*sh)
			rc.StrokePath(glyph, 2.2*scale, canvas.RoundCap, canvas.RoundJoin, cinnabarSeal, 0.92)
		}
	}

	// Lower-left artist signature seals & upper-right brow seal
	drawCinnabarSeal(marginX+10*scale, h-marginY-78*scale, 28*scale, 34*scale, true)
	drawCinnabarSeal(marginX+10*scale, h-marginY-36*scale, 28*scale, 28*scale, false)
	drawCinnabarSeal(w-marginX-34*scale, marginY+12*scale, 24*scale, 44*scale, true)

	rc.ApplyPaperTexture(0.11, 202620)

	c.MergeRenderCanvas(rc)
}
