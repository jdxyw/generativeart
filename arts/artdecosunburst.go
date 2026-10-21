package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type artDecoSunburst struct {
	rayCount  int
	archTiers int
}

// NewArtDecoSunburst returns an artDecoSunburst engine that renders a 1920s
// Chrysler-and-Metropolis-inspired Art Deco gilded brass sunburst, stepped ziggurat
// skyscrapers, scalloped fountain fans, and geometric chevron inlays.
func NewArtDecoSunburst(rayCount, archTiers int) *artDecoSunburst {
	if rayCount <= 0 {
		rayCount = 36
	}
	if archTiers <= 0 {
		archTiers = 7
	}
	return &artDecoSunburst{
		rayCount:  rayCount,
		archTiers: archTiers,
	}
}

// Generative renders the 1920s Art Deco gilded brass architectural poster.
func (ad *artDecoSunburst) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0

	// 1920s Art Deco Obsidian, Emerald, Peacock Teal & Brushed Brass Palette
	obsidianBlack := color.RGBA{R: 14, G: 16, B: 24, A: 255}
	midnightNavy := color.RGBA{R: 20, G: 30, B: 48, A: 255}
	imperialEmerald := color.RGBA{R: 18, G: 78, B: 74, A: 255}
	peacockTeal := color.RGBA{R: 38, G: 132, B: 128, A: 255}
	jadeMint := color.RGBA{R: 118, G: 196, B: 176, A: 255}

	antiqueBronze := color.RGBA{R: 112, G: 74, B: 28, A: 255}
	copperRose := color.RGBA{R: 188, G: 96, B: 56, A: 255}
	brushedBrass := color.RGBA{R: 222, G: 168, B: 62, A: 255}
	champagneGold := color.RGBA{R: 248, G: 214, B: 112, A: 255}
	ivoryPearl := color.RGBA{R: 252, G: 244, B: 214, A: 255}

	rc.FillBackground(obsidianBlack)
	rc.SetBlendMode(render.BlendOver)

	sunCX, sunCY := 0.50*w, 0.34*h
	maxR := math.Hypot(w, h) * 0.95
	pinstripeW := 1.45 * scale

	// Helper to draw a brushed Art Deco metallic polygon with internal bevel & gold pinstripe
	drawDecoTile := func(pts []render.Point, edgeCol, bodyCol, specCol color.RGBA) {
		if len(pts) < 3 {
			return
		}
		var cx, cy float64
		for _, p := range pts {
			cx += p.X
			cy += p.Y
		}
		center := render.Pt(cx/float64(len(pts)), cy/float64(len(pts)))

		rc.FillPolygon(pts, edgeCol, 1.0)
		in1 := make([]render.Point, len(pts))
		in2 := make([]render.Point, len(pts))
		for i, p := range pts {
			in1[i] = p.Lerp(center, 0.14)
			in2[i] = p.Lerp(center, 0.48)
		}
		rc.FillPolygon(in1, bodyCol, 1.0)
		rc.FillPolygon(in2, specCol, 0.88)

		outline := &canvas.Path{}
		outline.MoveTo(pts[0].X, pts[0].Y)
		for i := 1; i < len(pts); i++ {
			outline.LineTo(pts[i].X, pts[i].Y)
		}
		outline.Close()
		rc.StrokePath(outline, pinstripeW*1.35, canvas.RoundCap, canvas.RoundJoin, obsidianBlack, 1.0)
		rc.StrokePath(outline, pinstripeW*0.72, canvas.RoundCap, canvas.RoundJoin, champagneGold, 0.95)
	}

	// 1. Radiating Art Deco Sunburst Sectors & Stepped Chevron Rings from (sunCX, sunCY)
	rays := ad.rayCount
	if rays < 20 {
		rays = 36
	}
	dAng := (2.0 * math.Pi) / float64(rays)
	const radialBands = 7
	for rb := radialBands - 1; rb >= 0; rb-- {
		r0 := 0.08*w + float64(rb)*((maxR-0.08*w)/float64(radialBands))
		r1 := r0 + ((maxR-0.08*w)/float64(radialBands))*1.02

		for k := 0; k < rays; k++ {
			th0 := float64(k) * dAng
			th1 := th0 + dAng
			thMid := 0.5 * (th0 + th1)

			// Stepped chevron tip along each ray sector
			chevronLift := 14.0 * scale
			if k%2 == 1 {
				chevronLift = -10.0 * scale
			}
			pts := []render.Point{
				render.Pt(sunCX+r0*math.Cos(th0), sunCY+r0*math.Sin(th0)),
				render.Pt(sunCX+r1*math.Cos(th0), sunCY+r1*math.Sin(th0)),
				render.Pt(sunCX+(r1+chevronLift)*math.Cos(thMid), sunCY+(r1+chevronLift)*math.Sin(thMid)),
				render.Pt(sunCX+r1*math.Cos(th1), sunCY+r1*math.Sin(th1)),
				render.Pt(sunCX+r0*math.Cos(th1), sunCY+r0*math.Sin(th1)),
			}

			switch (k + rb) % 4 {
			case 0:
				drawDecoTile(pts, antiqueBronze, brushedBrass, champagneGold)
			case 1:
				drawDecoTile(pts, midnightNavy, imperialEmerald, peacockTeal)
			case 2:
				drawDecoTile(pts, antiqueBronze, copperRose, brushedBrass)
			default:
				drawDecoTile(pts, obsidianBlack, midnightNavy, imperialEmerald)
			}
		}
	}

	// 2. Scalloped Chrysler Sunburst Crown Arches Around the Solar Core
	tiers := ad.archTiers
	if tiers < 4 {
		tiers = 6
	}
	for t := tiers; t >= 1; t-- {
		rArch := (0.055 + float64(t)*0.036) * w
		scallops := 16 + t*2
		dSc := (2.0 * math.Pi) / float64(scallops)
		phase := float64(t%2) * dSc * 0.5

		for s := 0; s < scallops; s++ {
			th0 := float64(s)*dSc + phase
			th1 := th0 + dSc
			thMid := 0.5 * (th0 + th1)

			rIn := rArch * 0.72
			rTip := rArch * 1.10
			pts := []render.Point{
				render.Pt(sunCX+rIn*math.Cos(th0), sunCY+rIn*math.Sin(th0)),
				render.Pt(sunCX+rArch*math.Cos(th0), sunCY+rArch*math.Sin(th0)),
				render.Pt(sunCX+rTip*math.Cos(thMid), sunCY+rTip*math.Sin(thMid)),
				render.Pt(sunCX+rArch*math.Cos(th1), sunCY+rArch*math.Sin(th1)),
				render.Pt(sunCX+rIn*math.Cos(th1), sunCY+rIn*math.Sin(th1)),
			}
			if (t+s)%2 == 0 {
				drawDecoTile(pts, antiqueBronze, brushedBrass, ivoryPearl)
			} else {
				drawDecoTile(pts, imperialEmerald, peacockTeal, jadeMint)
			}

			// Classic Chrysler triangular window cutout inside each arch segment
			winIn := rArch * 0.80
			winOut := rArch * 0.98
			winTri := []render.Point{
				render.Pt(sunCX+winIn*math.Cos(thMid-dSc*0.20), sunCY+winIn*math.Sin(thMid-dSc*0.20)),
				render.Pt(sunCX+winOut*math.Cos(thMid), sunCY+winOut*math.Sin(thMid)),
				render.Pt(sunCX+winIn*math.Cos(thMid+dSc*0.20), sunCY+winIn*math.Sin(thMid+dSc*0.20)),
			}
			rc.FillPolygon(winTri, obsidianBlack, 0.92)
		}
	}

	// Central Solar Medallion
	sunDiamond := make([]render.Point, 16)
	for i := 0; i < 16; i++ {
		ang := (float64(i) / 16.0) * 2.0 * math.Pi
		r := 0.068 * w
		if i%2 == 1 {
			r = 0.042 * w
		}
		sunDiamond[i] = render.Pt(sunCX+r*math.Cos(ang), sunCY+r*math.Sin(ang))
	}
	drawDecoTile(sunDiamond, copperRose, champagneGold, ivoryPearl)

	// 3. Bilateral Stepped Ziggurat Skyscrapers (Left & Right Framing Towers)
	drawZigguratTower := func(xOuter, xDir float64) {
		const steps = 5
		for st := 0; st < steps; st++ {
			tFrac := float64(st) / float64(steps-1)
			towerW := (0.25 - tFrac*0.16) * w
			yTop := (0.16 + (1.0-tFrac)*0.48) * h
			if st == steps-1 {
				yTop = 0.14 * h
			}
			x0 := xOuter
			x1 := xOuter + xDir*towerW

			rect := []render.Point{
				render.Pt(x0, yTop),
				render.Pt(x1, yTop),
				render.Pt(x1, h),
				render.Pt(x0, h),
			}
			if st%2 == 0 {
				drawDecoTile(rect, obsidianBlack, midnightNavy, imperialEmerald)
			} else {
				drawDecoTile(rect, antiqueBronze, copperRose, brushedBrass)
			}

			// Vertical Art Deco brass fluted pilasters along the tower face
			flutes := 4
			for f := 1; f < flutes; f++ {
				u := float64(f) / float64(flutes)
				fx := x0 + (x1-x0)*u
				p := &canvas.Path{}
				p.MoveTo(fx, yTop+6*scale)
				p.LineTo(fx, h)
				rc.StrokePath(p, pinstripeW*0.85, canvas.RoundCap, canvas.RoundJoin, champagneGold, 0.85)
			}
		}
	}
	drawZigguratTower(0.0, 1.0)
	drawZigguratTower(w, -1.0)

	// 4. Central Metropolis Spire & Stepped Chevron Tower (Rising from Bottom Center to 0.34h)
	const spireTiers = 7
	for st := 0; st < spireTiers; st++ {
		u := float64(st) / float64(spireTiers-1) // 0 = base, 1 = needle crown
		halfW := (0.24 - u*0.205) * w
		yBot := h - u*0.42*h
		yTop := h - (u+0.16)*0.44*h
		yPeak := yTop - (18.0+u*16.0)*scale

		// Left and right symmetric chevron wings of the central spire
		leftWing := []render.Point{
			render.Pt(sunCX-halfW, yBot),
			render.Pt(sunCX-halfW*0.88, yTop),
			render.Pt(sunCX, yPeak),
			render.Pt(sunCX, yBot-10*scale),
		}
		rightWing := []render.Point{
			render.Pt(sunCX, yBot-10*scale),
			render.Pt(sunCX, yPeak),
			render.Pt(sunCX+halfW*0.88, yTop),
			render.Pt(sunCX+halfW, yBot),
		}
		if st%2 == 0 {
			drawDecoTile(leftWing, antiqueBronze, brushedBrass, champagneGold)
			drawDecoTile(rightWing, antiqueBronze, copperRose, brushedBrass)
		} else {
			drawDecoTile(leftWing, midnightNavy, imperialEmerald, jadeMint)
			drawDecoTile(rightWing, obsidianBlack, midnightNavy, peacockTeal)
		}
	}

	// 5. Foreground Scalloped Art Deco Fountain Fans along the Bottom Base
	fanCenters := []struct {
		x, y, r float64
	}{
		{0.18 * w, 1.02 * h, 0.22 * w},
		{0.82 * w, 1.02 * h, 0.22 * w},
		{0.50 * w, 1.04 * h, 0.26 * w},
	}
	for _, fc := range fanCenters {
		const blades = 11
		for b := 0; b < blades; b++ {
			th0 := -math.Pi + (float64(b)/float64(blades))*math.Pi
			th1 := -math.Pi + (float64(b+1)/float64(blades))*math.Pi
			thMid := 0.5 * (th0 + th1)

			pts := []render.Point{
				render.Pt(fc.x, fc.y),
				render.Pt(fc.x+fc.r*0.92*math.Cos(th0), fc.y+fc.r*0.92*math.Sin(th0)),
				render.Pt(fc.x+fc.r*1.06*math.Cos(thMid), fc.y+fc.r*1.06*math.Sin(thMid)),
				render.Pt(fc.x+fc.r*0.92*math.Cos(th1), fc.y+fc.r*0.92*math.Sin(th1)),
			}
			if b%2 == 0 {
				drawDecoTile(pts, antiqueBronze, brushedBrass, ivoryPearl)
			} else {
				drawDecoTile(pts, midnightNavy, peacockTeal, jadeMint)
			}
		}
	}

	// 6. Outer Double-Line Gilded Art Deco Poster Border Frame with Corner Stepped Motifs
	m := 14.0 * scale
	framePath := &canvas.Path{}
	framePath.MoveTo(m, m)
	framePath.LineTo(w-m, m)
	framePath.LineTo(w-m, h-m)
	framePath.LineTo(m, h-m)
	framePath.Close()

	m2 := 21.0 * scale
	framePath.MoveTo(m2, m2)
	framePath.LineTo(w-m2, m2)
	framePath.LineTo(w-m2, h-m2)
	framePath.LineTo(m2, h-m2)
	framePath.Close()

	rc.StrokePath(framePath, pinstripeW*1.8, canvas.RoundCap, canvas.RoundJoin, obsidianBlack, 1.0)
	rc.StrokePath(framePath, pinstripeW*0.95, canvas.RoundCap, canvas.RoundJoin, champagneGold, 1.0)

	c.MergeRenderCanvas(rc)
}
