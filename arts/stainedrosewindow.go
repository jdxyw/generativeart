package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type stainedRoseWindow struct {
	petals int
	rings  int
}

// NewStainedRoseWindow returns a stainedRoseWindow engine that renders a Chartres-
// and-Notre-Dame-inspired Gothic Cathedral Rose Window with backlit jewel-toned
// stained-glass compartments, Gothic pointed-arch tracery, and lead-came ironwork.
func NewStainedRoseWindow(petals, rings int) *stainedRoseWindow {
	if petals <= 0 {
		petals = 12
	}
	if rings <= 0 {
		rings = 5
	}
	return &stainedRoseWindow{
		petals: petals,
		rings:  rings,
	}
}

// Generative renders the backlit Gothic cathedral rose window and corner spandrel trefoils.
func (sr *stainedRoseWindow) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0
	cx, cy := 0.5*w, 0.5*h

	// Cathedral Lead-Came & Backlit Gothic Stained-Glass Palette
	cathedralStone := color.RGBA{R: 12, G: 11, B: 18, A: 255}
	leadCame := color.RGBA{R: 18, G: 16, B: 24, A: 255}
	stoneRib := color.RGBA{R: 34, G: 30, B: 42, A: 255}

	type glassSwatch struct {
		edge, core, glow color.RGBA
	}
	chartresSapphire := glassSwatch{
		edge: color.RGBA{R: 12, G: 28, B: 96, A: 255},
		core: color.RGBA{R: 28, G: 88, B: 214, A: 255},
		glow: color.RGBA{R: 118, G: 196, B: 255, A: 255},
	}
	rubyCrimson := glassSwatch{
		edge: color.RGBA{R: 92, G: 12, B: 28, A: 255},
		core: color.RGBA{R: 214, G: 32, B: 56, A: 255},
		glow: color.RGBA{R: 255, G: 132, B: 142, A: 255},
	}
	imperialAmethyst := glassSwatch{
		edge: color.RGBA{R: 58, G: 16, B: 88, A: 255},
		core: color.RGBA{R: 148, G: 44, B: 196, A: 255},
		glow: color.RGBA{R: 232, G: 142, B: 255, A: 255},
	}
	cathedralAmber := glassSwatch{
		edge: color.RGBA{R: 118, G: 58, B: 12, A: 255},
		core: color.RGBA{R: 238, G: 158, B: 28, A: 255},
		glow: color.RGBA{R: 255, G: 238, B: 128, A: 255},
	}
	emeraldGlass := glassSwatch{
		edge: color.RGBA{R: 12, G: 68, B: 48, A: 255},
		core: color.RGBA{R: 28, G: 164, B: 112, A: 255},
		glow: color.RGBA{R: 138, G: 244, B: 192, A: 255},
	}
	turquoiseGlass := glassSwatch{
		edge: color.RGBA{R: 14, G: 72, B: 88, A: 255},
		core: color.RGBA{R: 36, G: 172, B: 194, A: 255},
		glow: color.RGBA{R: 162, G: 246, B: 252, A: 255},
	}
	roseMagenta := glassSwatch{
		edge: color.RGBA{R: 88, G: 14, B: 56, A: 255},
		core: color.RGBA{R: 212, G: 42, B: 132, A: 255},
		glow: color.RGBA{R: 255, G: 158, B: 212, A: 255},
	}

	rc.FillBackground(cathedralStone)
	rc.SetBlendMode(render.BlendOver)

	leadW := 2.1 * scale
	numPetals := sr.petals
	if numPetals < 8 {
		numPetals = 12
	}

	// 1. Corner Gothic Spandrel Quatrefoil Windows (filling the 4 corners outside the Rose circle)
	spandrelCenters := []render.Point{
		render.Pt(0.12*w, 0.12*h),
		render.Pt(0.88*w, 0.12*h),
		render.Pt(0.88*w, 0.88*h),
		render.Pt(0.12*w, 0.88*h),
	}
	for _, sc := range spandrelCenters {
		rSp := 0.088 * w
		// 4-lobed quatrefoil petals in each corner
		for lobe := 0; lobe < 4; lobe++ {
			thC := float64(lobe)*0.5*math.Pi + 0.25*math.Pi
			pts := make([]render.Point, 0, 12)
			pts = append(pts, sc)
			lobeCenter := render.Pt(sc.X+rSp*0.52*math.Cos(thC), sc.Y+rSp*0.52*math.Sin(thC))
			for k := -4; k <= 4; k++ {
				ang := thC + (float64(k)/4.0)*0.65*math.Pi
				pts = append(pts, render.Pt(lobeCenter.X+rSp*0.52*math.Cos(ang), lobeCenter.Y+rSp*0.52*math.Sin(ang)))
			}
			sw := chartresSapphire
			if lobe%2 == 1 {
				sw = rubyCrimson
			}
			rc.DrawStainedGlassCell(pts, sw.edge, sw.core, sw.glow, leadCame, leadW*0.9)
		}
		// Central golden diamond in each corner spandrel
		dPts := make([]render.Point, 4)
		for k := 0; k < 4; k++ {
			a := float64(k) * 0.5 * math.Pi
			dPts[k] = render.Pt(sc.X+rSp*0.32*math.Cos(a), sc.Y+rSp*0.32*math.Sin(a))
		}
		rc.DrawStainedGlassCell(dPts, cathedralAmber.edge, cathedralAmber.core, cathedralAmber.glow, leadCame, leadW*0.85)
	}

	// 2. Outer Cathedral Stonework Molding & Cusp Ring (r in [0.385w, 0.468w])
	rOuterStone := 0.472 * w
	stoneBacking := make([]render.Point, 72)
	for i := 0; i < 72; i++ {
		a := (float64(i) / 72.0) * 2.0 * math.Pi
		stoneBacking[i] = render.Pt(cx+rOuterStone*math.Cos(a), cy+rOuterStone*math.Sin(a))
	}
	rc.FillPolygon(stoneBacking, stoneRib, 1.0)

	outerCount := numPetals * 2
	dOutAng := (2.0 * math.Pi) / float64(outerCount)
	rRing4In := 0.365 * w
	rRing4Out := 0.452 * w

	for i := 0; i < outerCount; i++ {
		th0 := float64(i) * dOutAng
		th1 := th0 + dOutAng
		thMid := 0.5 * (th0 + th1)

		// Pointed Gothic lancet arch compartment
		pts := []render.Point{
			render.Pt(cx+rRing4In*math.Cos(th0), cy+rRing4In*math.Sin(th0)),
			render.Pt(cx+(rRing4In+(rRing4Out-rRing4In)*0.58)*math.Cos(th0), cy+(rRing4In+(rRing4Out-rRing4In)*0.58)*math.Sin(th0)),
			render.Pt(cx+rRing4Out*math.Cos(thMid), cy+rRing4Out*math.Sin(thMid)),
			render.Pt(cx+(rRing4In+(rRing4Out-rRing4In)*0.58)*math.Cos(th1), cy+(rRing4In+(rRing4Out-rRing4In)*0.58)*math.Sin(th1)),
			render.Pt(cx+rRing4In*math.Cos(th1), cy+rRing4In*math.Sin(th1)),
		}
		var sw glassSwatch
		switch i % 4 {
		case 0:
			sw = chartresSapphire
		case 1:
			sw = rubyCrimson
		case 2:
			sw = emeraldGlass
		default:
			sw = cathedralAmber
		}
		rc.DrawStainedGlassCell(pts, sw.edge, sw.core, sw.glow, leadCame, leadW)

		// Outer interstitial triangle spandrel between pointed arches
		thNext := th1 + dOutAng*0.5
		spPts := []render.Point{
			render.Pt(cx+rRing4Out*math.Cos(thMid), cy+rRing4Out*math.Sin(thMid)),
			render.Pt(cx+rRing4Out*1.02*math.Cos(th1), cy+rRing4Out*1.02*math.Sin(th1)),
			render.Pt(cx+rRing4Out*math.Cos(thNext), cy+rRing4Out*math.Sin(thNext)),
			render.Pt(cx+(rRing4In+(rRing4Out-rRing4In)*0.58)*math.Cos(th1), cy+(rRing4In+(rRing4Out-rRing4In)*0.58)*math.Sin(th1)),
		}
		spSw := imperialAmethyst
		if i%2 == 1 {
			spSw = cathedralAmber
		}
		rc.DrawStainedGlassCell(spPts, spSw.edge, spSw.core, spSw.glow, leadCame, leadW*0.9)
	}

	// 3. Third Ring: Oculus Rosettes & Interlocking Diamonds (r in [0.258w, 0.365w])
	rRing3In := 0.258 * w
	rRing3Out := 0.365 * w
	dAng := (2.0 * math.Pi) / float64(numPetals)

	for p := 0; p < numPetals; p++ {
		th0 := float64(p) * dAng
		th1 := th0 + dAng
		thMid := 0.5 * (th0 + th1)

		// Left and right flanking cobalt/turquoise spandrels around the oculus
		leftFlank := []render.Point{
			render.Pt(cx+rRing3In*math.Cos(th0), cy+rRing3In*math.Sin(th0)),
			render.Pt(cx+rRing3Out*math.Cos(th0), cy+rRing3Out*math.Sin(th0)),
			render.Pt(cx+rRing3Out*math.Cos(thMid), cy+rRing3Out*math.Sin(thMid)),
			render.Pt(cx+rRing3In*math.Cos(thMid), cy+rRing3In*math.Sin(thMid)),
		}
		rightFlank := []render.Point{
			render.Pt(cx+rRing3In*math.Cos(thMid), cy+rRing3In*math.Sin(thMid)),
			render.Pt(cx+rRing3Out*math.Cos(thMid), cy+rRing3Out*math.Sin(thMid)),
			render.Pt(cx+rRing3Out*math.Cos(th1), cy+rRing3Out*math.Sin(th1)),
			render.Pt(cx+rRing3In*math.Cos(th1), cy+rRing3In*math.Sin(th1)),
		}
		rc.DrawStainedGlassCell(leftFlank, chartresSapphire.edge, chartresSapphire.core, chartresSapphire.glow, leadCame, leadW)
		rc.DrawStainedGlassCell(rightFlank, imperialAmethyst.edge, imperialAmethyst.core, imperialAmethyst.glow, leadCame, leadW)

		// Circular 6-lobed Oculus Medallion centered in the sector
		rOcCenter := 0.5 * (rRing3In + rRing3Out)
		ocCenter := render.Pt(cx+rOcCenter*math.Cos(thMid), cy+rOcCenter*math.Sin(thMid))
		rOc := (rRing3Out - rRing3In) * 0.44

		const ocPetals = 6
		for op := 0; op < ocPetals; op++ {
			a0 := (float64(op) / float64(ocPetals)) * 2.0 * math.Pi
			a1 := (float64(op+1) / float64(ocPetals)) * 2.0 * math.Pi
			aMid := 0.5 * (a0 + a1)
			tri := []render.Point{
				ocCenter,
				render.Pt(ocCenter.X+rOc*math.Cos(a0), ocCenter.Y+rOc*math.Sin(a0)),
				render.Pt(ocCenter.X+rOc*1.08*math.Cos(aMid), ocCenter.Y+rOc*1.08*math.Sin(aMid)),
				render.Pt(ocCenter.X+rOc*math.Cos(a1), ocCenter.Y+rOc*math.Sin(a1)),
			}
			sw := rubyCrimson
			if (op+p)%2 == 1 {
				sw = cathedralAmber
			}
			rc.DrawStainedGlassCell(tri, sw.edge, sw.core, sw.glow, leadCame, leadW*0.82)
		}
	}

	// 4. Second Ring: Radiating Gothic Pointed Colonnette Lancets (r in [0.125w, 0.258w])
	rRing2In := 0.125 * w
	rRing2Out := 0.258 * w
	for p := 0; p < numPetals; p++ {
		th0 := float64(p) * dAng
		th1 := th0 + dAng
		thMid := 0.5 * (th0 + th1)

		// Outer pointed arch spandrel
		spandrel := []render.Point{
			render.Pt(cx+rRing2Out*0.76*math.Cos(thMid), cy+rRing2Out*0.76*math.Sin(thMid)),
			render.Pt(cx+rRing2Out*math.Cos(thMid), cy+rRing2Out*math.Sin(thMid)),
			render.Pt(cx+rRing2Out*math.Cos(th1), cy+rRing2Out*math.Sin(th1)),
			render.Pt(cx+rRing2Out*0.76*math.Cos(th0+dAng*1.5), cy+rRing2Out*0.76*math.Sin(th0+dAng*1.5)),
		}
		spSw := turquoiseGlass
		if p%2 == 1 {
			spSw = emeraldGlass
		}
		rc.DrawStainedGlassCell(spandrel, spSw.edge, spSw.core, spSw.glow, leadCame, leadW*0.9)

		// Main Gothic Lancet Petal subdivided into lower shaft and upper trefoil head
		rSplit := rRing2In + (rRing2Out-rRing2In)*0.56
		shaftL := []render.Point{
			render.Pt(cx+rRing2In*math.Cos(th0), cy+rRing2In*math.Sin(th0)),
			render.Pt(cx+rSplit*math.Cos(th0+dAng*0.08), cy+rSplit*math.Sin(th0+dAng*0.08)),
			render.Pt(cx+rSplit*math.Cos(thMid), cy+rSplit*math.Sin(thMid)),
			render.Pt(cx+rRing2In*math.Cos(thMid), cy+rRing2In*math.Sin(thMid)),
		}
		shaftR := []render.Point{
			render.Pt(cx+rRing2In*math.Cos(thMid), cy+rRing2In*math.Sin(thMid)),
			render.Pt(cx+rSplit*math.Cos(thMid), cy+rSplit*math.Sin(thMid)),
			render.Pt(cx+rSplit*math.Cos(th1-dAng*0.08), cy+rSplit*math.Sin(th1-dAng*0.08)),
			render.Pt(cx+rRing2In*math.Cos(th1), cy+rRing2In*math.Sin(th1)),
		}
		archHead := []render.Point{
			render.Pt(cx+rSplit*math.Cos(th0+dAng*0.08), cy+rSplit*math.Sin(th0+dAng*0.08)),
			render.Pt(cx+rRing2Out*0.86*math.Cos(th0+dAng*0.22), cy+rRing2Out*0.86*math.Sin(th0+dAng*0.22)),
			render.Pt(cx+rRing2Out*0.96*math.Cos(thMid), cy+rRing2Out*0.96*math.Sin(thMid)),
			render.Pt(cx+rRing2Out*0.86*math.Cos(th1-dAng*0.22), cy+rRing2Out*0.86*math.Sin(th1-dAng*0.22)),
			render.Pt(cx+rSplit*math.Cos(th1-dAng*0.08), cy+rSplit*math.Sin(th1-dAng*0.08)),
			render.Pt(cx+rSplit*math.Cos(thMid), cy+rSplit*math.Sin(thMid)),
		}

		shSw := rubyCrimson
		headSw := cathedralAmber
		if p%2 == 1 {
			shSw = roseMagenta
			headSw = chartresSapphire
		}
		rc.DrawStainedGlassCell(shaftL, shSw.edge, shSw.core, shSw.glow, leadCame, leadW*0.9)
		rc.DrawStainedGlassCell(shaftR, shSw.edge, shSw.core, shSw.glow, leadCame, leadW*0.9)
		rc.DrawStainedGlassCell(archHead, headSw.edge, headSw.core, headSw.glow, leadCame, leadW*0.9)
	}

	// 5. Inner Star Rosette & Luminous Heart (r < 0.125w)
	rRing1In := 0.048 * w
	for p := 0; p < numPetals; p++ {
		th0 := float64(p) * dAng
		th1 := th0 + dAng
		thMid := 0.5 * (th0 + th1)

		pts := []render.Point{
			render.Pt(cx+rRing1In*math.Cos(th0), cy+rRing1In*math.Sin(th0)),
			render.Pt(cx+rRing2In*math.Cos(thMid), cy+rRing2In*math.Sin(thMid)),
			render.Pt(cx+rRing1In*math.Cos(th1), cy+rRing1In*math.Sin(th1)),
			render.Pt(cx, cy),
		}
		sw := cathedralAmber
		if p%2 == 1 {
			sw = rubyCrimson
		}
		rc.DrawStainedGlassCell(pts, sw.edge, sw.core, sw.glow, leadCame, leadW*0.88)
	}

	// 6. Heavy Gothic Concentric Stone Mullion Rings
	mullionRadii := []float64{rRing1In, rRing2In, rRing3In, rRing4In, rOuterStone}
	for _, rM := range mullionRadii {
		ringPath := &canvas.Path{}
		const steps = 96
		for i := 0; i <= steps; i++ {
			a := (float64(i) / float64(steps)) * 2.0 * math.Pi
			px, py := cx+rM*math.Cos(a), cy+rM*math.Sin(a)
			if i == 0 {
				ringPath.MoveTo(px, py)
			} else {
				ringPath.LineTo(px, py)
			}
		}
		ringPath.Close()
		rc.StrokePath(ringPath, leadW*1.65, canvas.RoundCap, canvas.RoundJoin, leadCame, 1.0)
	}

	c.MergeRenderCanvas(rc)
}
