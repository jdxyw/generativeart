package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type celestialArmillary struct {
	rings          int
	constellations int
}

// NewCelestialArmillary returns a celestialArmillary engine that renders a classical
// brass equatorial armillary sphere and astrolabe (浑天仪与二十八宿赤道星盘) with
// 3D tilted ecliptic/horizon rings, toothed clockwork gears, and constellation charts.
func NewCelestialArmillary(rings, constellations int) *celestialArmillary {
	if rings <= 0 {
		rings = 6
	}
	if constellations <= 0 {
		constellations = 12
	}
	return &celestialArmillary{
		rings:          rings,
		constellations: constellations,
	}
}

// Generative renders the brass equatorial armillary sphere and celestial star chart.
func (ca *celestialArmillary) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := c.Width(), c.Height()
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0
	cx, cy := 0.5*fw, 0.5*fh

	// Classical Brass Horology, Lapis Celestial Vault & Vermilion Lacquer Palette
	cosmosAbyss := color.RGBA{R: 8, G: 12, B: 24, A: 255}
	lapisVault := color.RGBA{R: 16, G: 36, B: 78, A: 255}
	ceruleanGrid := color.RGBA{R: 52, G: 138, B: 176, A: 255}
	turquoiseStar := color.RGBA{R: 72, G: 214, B: 212, A: 255}
	vermilionRing := color.RGBA{R: 196, G: 52, B: 38, A: 255}
	bronzeDark := color.RGBA{R: 84, G: 52, B: 22, A: 255}
	brassWarm := color.RGBA{R: 218, G: 158, B: 46, A: 255}
	goldSpec := color.RGBA{R: 252, G: 218, B: 108, A: 255}
	starlightIvory := color.RGBA{R: 255, G: 250, B: 232, A: 255}

	rc.FillBackground(cosmosAbyss)
	rc.SetBlendMode(render.BlendOver)

	rMax := 0.44 * math.Min(fw, fh)
	lightDir := render.Pt(-0.65, -0.75)

	// 1. Deep Lapis Celestial Vault Backdrop & Declination/Right-Ascension Graticule
	for y := 0; y < h; y++ {
		fy := float64(y) + 0.5
		for x := 0; x < w; x++ {
			fx := float64(x) + 0.5
			r := math.Hypot(fx-cx, fy-cy)
			if r <= rMax {
				u := r / rMax
				col := render.LerpOklab(lapisVault, cosmosAbyss, math.Pow(u, 1.35))
				rc.BlendPixel(x, y, col, 1.0)
			}
		}
	}

	// Celestial declination circles & 24 right-ascension hour meridians
	for k := 1; k <= 5; k++ {
		rk := rMax * (float64(k) / 6.0)
		rc.DrawSoftRing(cx, cy, rk, 0.9*scale, ceruleanGrid, 0.32)
	}
	meridians := &canvas.Path{}
	for m := 0; m < 24; m++ {
		a := (float64(m) / 24.0) * 2.0 * math.Pi
		rIn := rMax * 0.16
		if m%2 == 1 {
			rIn = rMax * 0.33
		}
		meridians.MoveTo(cx+rIn*math.Cos(a), cy+rIn*math.Sin(a))
		meridians.LineTo(cx+rMax*0.92*math.Cos(a), cy+rMax*0.92*math.Sin(a))
	}
	rc.StrokePath(meridians, 0.75*scale, canvas.RoundCap, canvas.RoundJoin, ceruleanGrid, 0.28)

	// 4 Corner Spandrel Astrolabe Gear Wheels outside the main Meridian Ring
	corners := []render.Point{
		render.Pt(0.11*fw, 0.11*fh),
		render.Pt(0.89*fw, 0.11*fh),
		render.Pt(0.89*fw, 0.89*fh),
		render.Pt(0.11*fw, 0.89*fh),
	}
	drawBrassGear := func(center render.Point, rOut, rIn float64, teeth int, baseCol color.RGBA) {
		pts := make([]render.Point, 0, teeth*4)
		dAng := (2.0 * math.Pi) / float64(teeth)
		rRoot := rOut * 0.86
		for t := 0; t < teeth; t++ {
			a0 := float64(t) * dAng
			pts = append(pts,
				render.Pt(center.X+rRoot*math.Cos(a0), center.Y+rRoot*math.Sin(a0)),
				render.Pt(center.X+rOut*math.Cos(a0+dAng*0.18), center.Y+rOut*math.Sin(a0+dAng*0.18)),
				render.Pt(center.X+rOut*math.Cos(a0+dAng*0.52), center.Y+rOut*math.Sin(a0+dAng*0.52)),
				render.Pt(center.X+rRoot*math.Cos(a0+dAng*0.70), center.Y+rRoot*math.Sin(a0+dAng*0.70)),
			)
		}
		gearSDF := render.SDFSubtract(render.SDFPolygon(pts), render.SDFCircle(center, rIn))
		rc.RenderSDFDropShadow(gearSDF, render.Pt(2.5*scale, 3.5*scale), 6.0*scale, 1.0*scale, cosmosAbyss, 0.75)
		rc.RenderSDFBevel(gearSDF, baseCol, bronzeDark, goldSpec, 2.8*scale, lightDir, 0.95)
	}
	for _, cp := range corners {
		drawBrassGear(cp, 38.0*scale, 16.0*scale, 16, brassWarm)
		rc.DrawSoftRing(cp.X, cp.Y, 8.0*scale, 2.2*scale, goldSpec, 0.85)
	}

	// 2. Twenty-Eight Lunar Mansions Star Constellations (二十八宿星官连线图)
	rng := rand.New(rand.NewSource(c.Seed() + 618))
	numConst := ca.constellations
	if numConst < 6 {
		numConst = 12
	}
	for cIdx := 0; cIdx < numConst; cIdx++ {
		baseAng := (float64(cIdx)/float64(numConst))*2.0*math.Pi + 0.15
		baseR := (0.28 + 0.46*float64(cIdx%3)/2.0) * rMax
		nStars := 4 + (cIdx % 3)
		stars := make([]render.Point, nStars)
		constPath := &canvas.Path{}
		for s := 0; s < nStars; s++ {
			sa := baseAng + (float64(s)-1.5)*0.11 + (rng.Float64()-0.5)*0.05
			sr := baseR + (rng.Float64()-0.5)*0.18*rMax
			pt := render.Pt(cx+sr*math.Cos(sa), cy+sr*math.Sin(sa))
			stars[s] = pt
			if s == 0 {
				constPath.MoveTo(pt.X, pt.Y)
			} else {
				constPath.LineTo(pt.X, pt.Y)
			}
		}
		rc.StrokePath(constPath, 1.05*scale, canvas.RoundCap, canvas.RoundJoin, goldSpec, 0.65)
		for s, pt := range stars {
			starCol := starlightIvory
			if s == 0 {
				starCol = turquoiseStar
				rc.DrawSoftRing(pt.X, pt.Y, 4.8*scale, 0.9*scale, goldSpec, 0.75)
			}
			rc.DrawGlowDot(pt.X, pt.Y, (2.2+float64(s%2)*1.1)*scale, starCol, 0.96, 0.82)
		}
	}

	// Helper to draw a 3D tilted armillary ring (either back half or front half for true 3D occlusion)
	drawTiltedRingHalf := func(radius, aspect, rotRad, width float64, baseCol, rimCol color.RGBA, frontHalf bool) {
		cosA, sinA := math.Cos(rotRad), math.Sin(rotRad)
		tStart, tEnd := math.Pi, 2.0*math.Pi
		if frontHalf {
			tStart, tEnd = 0.0, math.Pi
		}

		const steps = 64
		pts := make([]render.Point, 0, steps+1)
		tickPath := &canvas.Path{}
		for i := 0; i <= steps; i++ {
			t := tStart + (float64(i)/float64(steps))*(tEnd-tStart)
			lx := radius * math.Cos(t)
			ly := radius * aspect * math.Sin(t)
			px := cx + lx*cosA - ly*sinA
			py := cy + lx*sinA + ly*cosA
			pts = append(pts, render.Pt(px, py))

			if i%2 == 0 {
				nx := math.Cos(t)*cosA - aspect*math.Sin(t)*sinA
				ny := math.Cos(t)*sinA + aspect*math.Sin(t)*cosA
				nLen := math.Hypot(nx, ny) + 1e-6
				nx, ny = nx/nLen, ny/nLen
				tickLen := width * 0.42
				tickPath.MoveTo(px-nx*tickLen, py-ny*tickLen)
				tickPath.LineTo(px+nx*tickLen, py+ny*tickLen)
			}
		}

		// Cast shadow for front half
		if frontHalf {
			shPath := &canvas.Path{}
			for i, pt := range pts {
				if i == 0 {
					shPath.MoveTo(pt.X+3.5*scale, pt.Y+4.5*scale)
				} else {
					shPath.LineTo(pt.X+3.5*scale, pt.Y+4.5*scale)
				}
			}
			rc.StrokePath(shPath, width*1.25, canvas.RoundCap, canvas.RoundJoin, cosmosAbyss, 0.65)
		}

		alpha := 0.86
		if frontHalf {
			alpha = 0.98
		}
		rc.DrawImpastoStroke(pts, width, baseCol, lightDir, alpha)
		rc.StrokePath(tickPath, 0.85*scale, canvas.RoundCap, canvas.RoundJoin, starlightIvory, alpha*0.72)

		// Gilded selvedge edges
		rc.DrawTaperedSpline(pts, 1.35*scale, render.TaperUniform, func(u float64) color.RGBA {
			return rimCol
		}, alpha*0.88)
	}

	type ringSpec struct {
		r, aspect, rot, w float64
		col, rim          color.RGBA
	}
	armillaryRings := []ringSpec{
		{rMax * 0.86, 0.36, -0.42, 11.5 * scale, brassWarm, goldSpec},           // Ecliptic Zodiac Ring (黄道游仪环)
		{rMax * 0.82, 0.42, 0.62, 10.5 * scale, vermilionRing, goldSpec},        // Equatorial Colure Ring (赤道环)
		{rMax * 0.66, 0.34, 1.18, 9.2 * scale, turquoiseStar, starlightIvory},   // Lunar Horizon Ring (白道环)
		{rMax * 0.50, 0.46, -0.95, 8.5 * scale, brassWarm, starlightIvory},      // Inner Solstitial Ring (四游环)
	}

	// 3. Draw Back Halves of the 3D Tilted Armillary Rings
	for _, ar := range armillaryRings {
		drawTiltedRingHalf(ar.r, ar.aspect, ar.rot, ar.w, ar.col, ar.rim, false)
	}

	// 4. Central Clockwork Orrery Gear, Sighting Alidade (窥管横箫) & Blazing Solar Core
	drawBrassGear(render.Pt(cx, cy), rMax*0.28, rMax*0.14, 24, brassWarm)
	drawBrassGear(render.Pt(cx, cy), rMax*0.18, rMax*0.07, 16, vermilionRing)

	// Diagonal Sighting Alidade Rule (窥管横箫) spanning across the Meridian Ring
	alidadeAng := -0.58
	alidadeR := rMax * 0.96
	pA := render.Pt(cx-alidadeR*math.Cos(alidadeAng), cy-alidadeR*math.Sin(alidadeAng))
	pB := render.Pt(cx+alidadeR*math.Cos(alidadeAng), cy+alidadeR*math.Sin(alidadeAng))
	alidadeSDF := render.SDFSegment(pA, pB, 5.5*scale)
	rc.RenderSDFDropShadow(alidadeSDF, render.Pt(3*scale, 4*scale), 8*scale, 1*scale, cosmosAbyss, 0.78)
	rc.RenderSDFBevel(alidadeSDF, brassWarm, bronzeDark, goldSpec, 2.6*scale, lightDir, 0.98)

	// Central 3D Beveled Sun Core & Orbiting Planetary Orbs
	sunSDF := render.SDFCircle(render.Pt(cx, cy), 19.5*scale)
	rc.RenderSDFDropShadow(sunSDF, render.Pt(3*scale, 4*scale), 10*scale, 2*scale, cosmosAbyss, 0.85)
	rc.RenderSDFBevel(sunSDF, goldSpec, vermilionRing, starlightIvory, 8.5*scale, lightDir, 1.0)

	// 5. Draw Front Halves of the 3D Tilted Armillary Rings (Completing Over/Under 3D Interlocking)
	for _, ar := range armillaryRings {
		drawTiltedRingHalf(ar.r, ar.aspect, ar.rot, ar.w, ar.col, ar.rim, true)
	}

	// 6. Grand Outer Graduated Brass Meridian Bezel Ring (外层子午周天360度刻度铜环)
	meridianSDF := render.SDFOnion(render.SDFCircle(render.Pt(cx, cy), rMax), 15.5*scale)
	rc.RenderSDFDropShadow(meridianSDF, render.Pt(5*scale, 6*scale), 14*scale, 2*scale, cosmosAbyss, 0.88)
	rc.RenderSDFBevel(meridianSDF, brassWarm, bronzeDark, goldSpec, 6.2*scale, lightDir, 1.0)

	// 360-degree finely engraved degree ticks and 24 solar-term rivets around the outer bezel
	degTicks := &canvas.Path{}
	for deg := 0; deg < 360; deg += 2 {
		a := (float64(deg) / 360.0) * 2.0 * math.Pi
		tIn := rMax - 14.5*scale
		tOut := rMax - 9.0*scale
		if deg%15 == 0 {
			tOut = rMax - 3.5*scale
		} else if deg%5 == 0 {
			tOut = rMax - 6.5*scale
		}
		degTicks.MoveTo(cx+tIn*math.Cos(a), cy+tIn*math.Sin(a))
		degTicks.LineTo(cx+tOut*math.Cos(a), cy+tOut*math.Sin(a))
	}
	rc.StrokePath(degTicks, 0.95*scale, canvas.RoundCap, canvas.RoundJoin, starlightIvory, 0.82)

	for rv := 0; rv < 24; rv++ {
		a := (float64(rv) / 24.0) * 2.0 * math.Pi
		rx := cx + (rMax+6.5*scale)*math.Cos(a)
		ry := cy + (rMax+6.5*scale)*math.Sin(a)
		rc.DrawGlowDot(rx, ry, 2.2*scale, starlightIvory, 0.92, 0.65)
	}

	rc.ApplyBloom(0.82, 6, 0.22)
	c.MergeRenderCanvas(rc)
}
