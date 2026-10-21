package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type cloisonneEnamel struct {
	rings  int
	petals int
}

// NewCloisonneEnamel returns a cloisonneEnamel engine that renders an Imperial
// Jingtai-Blue (景泰蓝) cloisonné enamel medallion with gilded brass wire filigree,
// vitreous enamel meniscus shading, auspicious Ruyi cloud scales, and lotus petals.
func NewCloisonneEnamel(rings, petals int) *cloisonneEnamel {
	if rings <= 0 {
		rings = 6
	}
	if petals <= 0 {
		petals = 12
	}
	return &cloisonneEnamel{
		rings:  rings,
		petals: petals,
	}
}

// Generative renders the Imperial cloisonné enamel vessel medallion and Ruyi cloud lattice.
func (ce *cloisonneEnamel) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0
	cx, cy := 0.5*w, 0.5*h

	// Authentic Imperial Jingtai-Blue (景泰蓝) vitreous enamel & gilded bronze palette
	abyssBronze := color.RGBA{R: 18, G: 14, B: 22, A: 255}
	wireShadow := color.RGBA{R: 68, G: 44, B: 18, A: 255}
	wireGold := color.RGBA{R: 226, G: 178, B: 68, A: 255}
	wireSpec := color.RGBA{R: 255, G: 238, B: 162, A: 255}

	type enamelSwatch struct {
		base, rim, highlight color.RGBA
	}
	jingtaiTurquoise := enamelSwatch{
		base:      color.RGBA{R: 28, G: 158, B: 176, A: 255},
		rim:       color.RGBA{R: 12, G: 78, B: 102, A: 255},
		highlight: color.RGBA{R: 164, G: 238, B: 242, A: 255},
	}
	imperialLapis := enamelSwatch{
		base:      color.RGBA{R: 32, G: 74, B: 158, A: 255},
		rim:       color.RGBA{R: 14, G: 32, B: 84, A: 255},
		highlight: color.RGBA{R: 122, G: 176, B: 244, A: 255},
	}
	cinnabarCoral := enamelSwatch{
		base:      color.RGBA{R: 208, G: 58, B: 46, A: 255},
		rim:       color.RGBA{R: 112, G: 22, B: 24, A: 255},
		highlight: color.RGBA{R: 252, G: 158, B: 132, A: 255},
	}
	celadonJade := enamelSwatch{
		base:      color.RGBA{R: 48, G: 154, B: 116, A: 255},
		rim:       color.RGBA{R: 18, G: 76, B: 58, A: 255},
		highlight: color.RGBA{R: 172, G: 238, B: 204, A: 255},
	}
	imperialAmber := enamelSwatch{
		base:      color.RGBA{R: 236, G: 168, B: 44, A: 255},
		rim:       color.RGBA{R: 138, G: 78, B: 16, A: 255},
		highlight: color.RGBA{R: 255, G: 236, B: 148, A: 255},
	}
	peonyRose := enamelSwatch{
		base:      color.RGBA{R: 194, G: 48, B: 108, A: 255},
		rim:       color.RGBA{R: 96, G: 18, B: 54, A: 255},
		highlight: color.RGBA{R: 250, G: 164, B: 198, A: 255},
	}
	ivoryEnamel := enamelSwatch{
		base:      color.RGBA{R: 238, G: 232, B: 210, A: 255},
		rim:       color.RGBA{R: 156, G: 144, B: 118, A: 255},
		highlight: color.RGBA{R: 255, G: 252, B: 242, A: 255},
	}

	rc.FillBackground(abyssBronze)
	rc.SetBlendMode(render.BlendOver)

	wireW := 1.85 * scale

	// 1. Background Concentric Scalloped Ruyi Cloud-Scale Cloisonné Ground (covering the entire canvas)
	maxGroundR := math.Hypot(w*0.5, h*0.5) * 1.06
	rMedallion := 0.44 * w
	const groundRings = 8
	for gr := groundRings - 1; gr >= 0; gr-- {
		rInner := rMedallion*0.92 + float64(gr)*((maxGroundR-rMedallion*0.92)/float64(groundRings))
		rOuter := rInner + ((maxGroundR-rMedallion*0.92)/float64(groundRings))*1.08
		sectors := 28 + gr*4
		dAng := (2.0 * math.Pi) / float64(sectors)
		phase := float64(gr%2) * dAng * 0.5

		for s := 0; s < sectors; s++ {
			th0 := float64(s)*dAng + phase
			th1 := th0 + dAng
			thMid := 0.5 * (th0 + th1)

			// Build a scalloped cloud-scale polygon
			pts := make([]render.Point, 0, 10)
			pts = append(pts,
				render.Pt(cx+rInner*math.Cos(th0), cy+rInner*math.Sin(th0)),
				render.Pt(cx+rInner*math.Cos(thMid), cy+rInner*math.Sin(thMid)),
				render.Pt(cx+rInner*math.Cos(th1), cy+rInner*math.Sin(th1)),
			)
			for k := 4; k >= 0; k-- {
				u := float64(k) / 4.0
				ang := th0 + u*dAng
				arch := math.Sin(u*math.Pi) * (rOuter - rInner) * 0.38
				rCur := rOuter + arch
				pts = append(pts, render.Pt(cx+rCur*math.Cos(ang), cy+rCur*math.Sin(ang)))
			}

			sw := imperialLapis
			if (gr+s)%3 == 0 {
				sw = jingtaiTurquoise
			} else if (gr+s)%7 == 0 {
				sw = celadonJade
			}
			rc.DrawCloisonneCell(pts, sw.base, sw.rim, sw.highlight, wireShadow, wireGold, wireSpec, wireW*0.85)
		}
	}

	// 2. Corner Auspicious Ruyi-Cloud & Lotus Spandrel Medallions
	corners := []render.Point{
		render.Pt(0, 0),
		render.Pt(w, 0),
		render.Pt(w, h),
		render.Pt(0, h),
	}
	for cIdx, corner := range corners {
		baseAng := float64(cIdx) * 0.5 * math.Pi
		const cornerPetals = 5
		for p := 0; p < cornerPetals; p++ {
			thCenter := baseAng + (float64(p)+0.5)*(0.5*math.Pi/float64(cornerPetals))
			dTh := (0.5 * math.Pi / float64(cornerPetals)) * 0.46
			r0 := 18.0 * scale
			r1 := 105.0 * scale
			rTip := 138.0 * scale

			pts := []render.Point{
				render.Pt(corner.X+r0*math.Cos(thCenter-dTh), corner.Y+r0*math.Sin(thCenter-dTh)),
				render.Pt(corner.X+r1*math.Cos(thCenter-dTh*0.85), corner.Y+r1*math.Sin(thCenter-dTh*0.85)),
				render.Pt(corner.X+rTip*math.Cos(thCenter), corner.Y+rTip*math.Sin(thCenter)),
				render.Pt(corner.X+r1*math.Cos(thCenter+dTh*0.85), corner.Y+r1*math.Sin(thCenter+dTh*0.85)),
				render.Pt(corner.X+r0*math.Cos(thCenter+dTh), corner.Y+r0*math.Sin(thCenter+dTh)),
			}
			sw := cinnabarCoral
			if p%2 == 1 {
				sw = imperialAmber
			}
			rc.DrawCloisonneCell(pts, sw.base, sw.rim, sw.highlight, wireShadow, wireGold, wireSpec, wireW)
		}
	}

	// 3. Outer Gilded Rim & Turquoise/Coral Fret Border Ring of the Central Medallion
	numPetals := ce.petals
	if numPetals < 8 {
		numPetals = 12
	}
	outerSegs := numPetals * 2
	rBorderIn := 0.395 * w
	rBorderOut := 0.442 * w
	for i := 0; i < outerSegs; i++ {
		th0 := (float64(i) / float64(outerSegs)) * 2.0 * math.Pi
		th1 := (float64(i+1) / float64(outerSegs)) * 2.0 * math.Pi
		thMid := 0.5 * (th0 + th1)

		pts := []render.Point{
			render.Pt(cx+rBorderIn*math.Cos(th0), cy+rBorderIn*math.Sin(th0)),
			render.Pt(cx+rBorderOut*math.Cos(th0), cy+rBorderOut*math.Sin(th0)),
			render.Pt(cx+(rBorderOut+6.0*scale)*math.Cos(thMid), cy+(rBorderOut+6.0*scale)*math.Sin(thMid)),
			render.Pt(cx+rBorderOut*math.Cos(th1), cy+rBorderOut*math.Sin(th1)),
			render.Pt(cx+rBorderIn*math.Cos(th1), cy+rBorderIn*math.Sin(th1)),
		}
		sw := jingtaiTurquoise
		switch i % 4 {
		case 1:
			sw = cinnabarCoral
		case 2:
			sw = imperialAmber
		case 3:
			sw = celadonJade
		}
		rc.DrawCloisonneCell(pts, sw.base, sw.rim, sw.highlight, wireShadow, wireGold, wireSpec, wireW)
	}

	// 4. Outer Lotus Lappet Ring & Interstitial Lapis Diamonds (r in [0.275w, 0.395w])
	rOutIn := 0.275 * w
	rOutTip := 0.392 * w
	dAng := (2.0 * math.Pi) / float64(numPetals)

	// First draw the interstitial recessed Lapis & Jade diamond compartments between outer petals
	for p := 0; p < numPetals; p++ {
		thValley := (float64(p) + 0.5) * dAng
		pts := []render.Point{
			render.Pt(cx+rOutIn*math.Cos(thValley), cy+rOutIn*math.Sin(thValley)),
			render.Pt(cx+0.34*w*math.Cos(thValley-dAng*0.32), cy+0.34*w*math.Sin(thValley-dAng*0.32)),
			render.Pt(cx+rBorderIn*math.Cos(thValley-dAng*0.22), cy+rBorderIn*math.Sin(thValley-dAng*0.22)),
			render.Pt(cx+rBorderIn*math.Cos(thValley), cy+rBorderIn*math.Sin(thValley)),
			render.Pt(cx+rBorderIn*math.Cos(thValley+dAng*0.22), cy+rBorderIn*math.Sin(thValley+dAng*0.22)),
			render.Pt(cx+0.34*w*math.Cos(thValley+dAng*0.32), cy+0.34*w*math.Sin(thValley+dAng*0.32)),
		}
		sw := imperialLapis
		if p%2 == 1 {
			sw = jingtaiTurquoise
		}
		rc.DrawCloisonneCell(pts, sw.base, sw.rim, sw.highlight, wireShadow, wireGold, wireSpec, wireW)

		// Add a curled golden brass filigree spiral inside each valley compartment
		spiral := &canvas.Path{}
		vCenter := render.Pt(cx+0.352*w*math.Cos(thValley), cy+0.352*w*math.Sin(thValley))
		for k := 0; k <= 20; k++ {
			u := float64(k) / 20.0
			ang := thValley + u*2.8*math.Pi
			rad := (1.0 - u*0.82) * 11.0 * scale
			pt := render.Pt(vCenter.X+rad*math.Cos(ang), vCenter.Y+rad*math.Sin(ang))
			if k == 0 {
				spiral.MoveTo(pt.X, pt.Y)
			} else {
				spiral.LineTo(pt.X, pt.Y)
			}
		}
		rc.StrokePath(spiral, wireW*0.75, canvas.RoundCap, canvas.RoundJoin, wireGold, 0.95)
	}

	// Helper to build a sculptured ogee lotus petal polygon and inner core cell
	drawCloisonnePetalRing := func(rBase, rTip, phase float64, outerSw, innerSw, pearlSw enamelSwatch) {
		for p := 0; p < numPetals; p++ {
			thC := float64(p)*dAng + phase
			span := dAng * 0.44

			// Sculpted ogee lotus petal outline
			pts := []render.Point{
				render.Pt(cx+rBase*math.Cos(thC-span*0.65), cy+rBase*math.Sin(thC-span*0.65)),
				render.Pt(cx+(rBase+(rTip-rBase)*0.35)*math.Cos(thC-span), cy+(rBase+(rTip-rBase)*0.35)*math.Sin(thC-span)),
				render.Pt(cx+(rBase+(rTip-rBase)*0.72)*math.Cos(thC-span*0.78), cy+(rBase+(rTip-rBase)*0.72)*math.Sin(thC-span*0.78)),
				render.Pt(cx+rTip*math.Cos(thC), cy+rTip*math.Sin(thC)),
				render.Pt(cx+(rBase+(rTip-rBase)*0.72)*math.Cos(thC+span*0.78), cy+(rBase+(rTip-rBase)*0.72)*math.Sin(thC+span*0.78)),
				render.Pt(cx+(rBase+(rTip-rBase)*0.35)*math.Cos(thC+span), cy+(rBase+(rTip-rBase)*0.35)*math.Sin(thC+span)),
				render.Pt(cx+rBase*math.Cos(thC+span*0.65), cy+rBase*math.Sin(thC+span*0.65)),
			}
			sw := outerSw
			if p%2 == 1 {
				sw = peonyRose
			}
			rc.DrawCloisonneCell(pts, sw.base, sw.rim, sw.highlight, wireShadow, wireGold, wireSpec, wireW)

			// Inner nested Ruyi heart compartment inside the petal
			rIn0 := rBase + (rTip-rBase)*0.14
			rIn1 := rBase + (rTip-rBase)*0.70
			inSpan := span * 0.52
			inPts := []render.Point{
				render.Pt(cx+rIn0*math.Cos(thC-inSpan*0.6), cy+rIn0*math.Sin(thC-inSpan*0.6)),
				render.Pt(cx+(rIn0+(rIn1-rIn0)*0.55)*math.Cos(thC-inSpan), cy+(rIn0+(rIn1-rIn0)*0.55)*math.Sin(thC-inSpan)),
				render.Pt(cx+rIn1*math.Cos(thC), cy+rIn1*math.Sin(thC)),
				render.Pt(cx+(rIn0+(rIn1-rIn0)*0.55)*math.Cos(thC+inSpan), cy+(rIn0+(rIn1-rIn0)*0.55)*math.Sin(thC+inSpan)),
				render.Pt(cx+rIn0*math.Cos(thC+inSpan*0.6), cy+rIn0*math.Sin(thC+inSpan*0.6)),
			}
			inSw := innerSw
			if p%2 == 1 {
				inSw = ivoryEnamel
			}
			rc.DrawCloisonneCell(inPts, inSw.base, inSw.rim, inSw.highlight, wireShadow, wireGold, wireSpec, wireW*0.88)

			// Gilded enamel pearl cabochon near the petal tip
			rPearl := rBase + (rTip-rBase)*0.82
			pCenter := render.Pt(cx+rPearl*math.Cos(thC), cy+rPearl*math.Sin(thC))
			pearlPts := make([]render.Point, 10)
			for k := 0; k < 10; k++ {
				a := (float64(k) / 10.0) * 2.0 * math.Pi
				pearlPts[k] = render.Pt(pCenter.X+4.8*scale*math.Cos(a), pCenter.Y+4.8*scale*math.Sin(a))
			}
			rc.DrawCloisonneCell(pearlPts, pearlSw.base, pearlSw.rim, pearlSw.highlight, wireShadow, wireGold, wireSpec, wireW*0.75)
		}
	}

	// Render Outer Lotus Ring
	drawCloisonnePetalRing(rOutIn, rOutTip, 0.0, cinnabarCoral, imperialAmber, jingtaiTurquoise)

	// 5. Middle Jade & Turquoise Ruyi Band (r in [0.165w, 0.280w])
	drawCloisonnePetalRing(0.165*w, 0.282*w, dAng*0.5, jingtaiTurquoise, celadonJade, imperialAmber)

	// 6. Inner Imperial Gold & Lapis Star Medallion (r in [0.068w, 0.170w])
	rCoreIn := 0.068 * w
	rCoreOut := 0.170 * w
	for p := 0; p < numPetals; p++ {
		th0 := float64(p) * dAng
		th1 := th0 + dAng
		thMid := 0.5 * (th0 + th1)

		pts := []render.Point{
			render.Pt(cx+rCoreIn*math.Cos(th0), cy+rCoreIn*math.Sin(th0)),
			render.Pt(cx+rCoreOut*0.78*math.Cos(th0), cy+rCoreOut*0.78*math.Sin(th0)),
			render.Pt(cx+rCoreOut*math.Cos(thMid), cy+rCoreOut*math.Sin(thMid)),
			render.Pt(cx+rCoreOut*0.78*math.Cos(th1), cy+rCoreOut*0.78*math.Sin(th1)),
			render.Pt(cx+rCoreIn*math.Cos(th1), cy+rCoreIn*math.Sin(th1)),
		}
		sw := imperialAmber
		if p%2 == 1 {
			sw = imperialLapis
		}
		rc.DrawCloisonneCell(pts, sw.base, sw.rim, sw.highlight, wireShadow, wireGold, wireSpec, wireW*0.92)
	}

	// 7. Central Cabochon Jewel & Gilded Sunburst Heart (r < 0.068w)
	centerPts := make([]render.Point, 24)
	for i := 0; i < 24; i++ {
		a := (float64(i) / 24.0) * 2.0 * math.Pi
		centerPts[i] = render.Pt(cx+rCoreIn*math.Cos(a), cy+rCoreIn*math.Sin(a))
	}
	rc.DrawCloisonneCell(centerPts, cinnabarCoral.base, cinnabarCoral.rim, cinnabarCoral.highlight, wireShadow, wireGold, wireSpec, wireW*1.1)

	innerGemPts := make([]render.Point, 16)
	for i := 0; i < 16; i++ {
		a := (float64(i) / 16.0) * 2.0 * math.Pi
		rG := rCoreIn * 0.52
		if i%2 == 1 {
			rG = rCoreIn * 0.30
		}
		innerGemPts[i] = render.Pt(cx+rG*math.Cos(a), cy+rG*math.Sin(a))
	}
	rc.DrawCloisonneCell(innerGemPts, imperialAmber.base, imperialAmber.rim, imperialAmber.highlight, wireShadow, wireGold, wireSpec, wireW*0.85)

	c.MergeRenderCanvas(rc)
}
