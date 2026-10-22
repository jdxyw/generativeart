package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type champleveEnamel struct {
	symmetry  int
	ringCount int
}

// NewChampleveEnamel returns a champleveEnamel engine that renders a Warring States
// (战国错金银嵌绿松石青铜鉴) gold, silver, and turquoise inlaid bronze mirror with
// cloud-and-thunder spirals (云雷纹), interlocking dragon scrolls, and verdigris patina.
func NewChampleveEnamel(symmetry, ringCount int) *champleveEnamel {
	if symmetry <= 0 {
		symmetry = 8
	}
	if ringCount <= 0 {
		ringCount = 4
	}
	return &champleveEnamel{
		symmetry:  symmetry,
		ringCount: ringCount,
	}
}

// Generative renders the Warring States gold, silver & turquoise inlaid bronze mirror.
func (ce *champleveEnamel) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0
	cx, cy := 0.5*w, 0.5*h

	// Warring States Black-Lacquer Bronze, 24K Gold Inlay, Sterling Silver & Turquoise Cabochon Palette
	silkBacking := color.RGBA{R: 12, G: 14, B: 18, A: 255}
	darkBronze := color.RGBA{R: 26, G: 28, B: 26, A: 255}
	agedBronze := color.RGBA{R: 44, G: 52, B: 46, A: 255}
	verdigrisPatina := color.RGBA{R: 46, G: 132, B: 112, A: 255}
	turquoiseGem := color.RGBA{R: 54, G: 196, B: 184, A: 255}
	turquoiseRim := color.RGBA{R: 18, G: 88, B: 86, A: 255}
	paleTurquoise := color.RGBA{R: 186, G: 248, B: 242, A: 255}

	inlayGold := color.RGBA{R: 244, G: 188, B: 62, A: 255}
	goldShadow := color.RGBA{R: 118, G: 76, B: 20, A: 255}
	goldSpec := color.RGBA{R: 255, G: 244, B: 178, A: 255}
	inlaySilver := color.RGBA{R: 214, G: 224, B: 232, A: 255}
	cinnabarLacquer := color.RGBA{R: 186, G: 52, B: 42, A: 255}

	rc.FillBackground(silkBacking)
	rc.SetBlendMode(render.BlendOver)

	sym := ce.symmetry
	if sym < 4 {
		sym = 8
	}

	// 1. Corner Warring States Cloud-Scroll Spandrels Outside the Mirror Disc
	for cIdx := 0; cIdx < 4; cIdx++ {
		ang := float64(cIdx)*0.5*math.Pi + 0.25*math.Pi
		dist := 0.56 * w
		scX := cx + dist*math.Cos(ang)
		scY := cy + dist*math.Sin(ang)
		rc.DrawSoftRing(scX, scY, 0.14*w, 2.2*scale, goldShadow, 0.75)
		rc.DrawSoftRing(scX, scY, 0.11*w, 1.6*scale, verdigrisPatina, 0.65)
	}

	// 2. Main Circular Stepped Bronze Mirror Rim (r = 0.465w) with SDF Drop Shadow & Bevel
	rMirror := 0.462 * w
	mirrorSDF := render.SDFCircle(render.Pt(cx, cy), rMirror)
	rc.RenderSDFDropShadow(mirrorSDF, render.Pt(6*scale, 8*scale), 16*scale, 3*scale, silkBacking, 0.92)
	rc.RenderSDFBevel(mirrorSDF, agedBronze, darkBronze, inlayGold, 22.0*scale, render.Pt(-0.65, -0.75), 1.0)

	// Recessed dark bronze mirror field (r = 0.415w)
	rField := 0.415 * w
	fieldSDF := render.SDFCircle(render.Pt(cx, cy), rField)
	rc.RenderSDFBevel(fieldSDF, darkBronze, silkBacking, verdigrisPatina, 12.0*scale, render.Pt(0.65, 0.75), 1.0)

	// 3. Outer Rim Inlaid Gold & Silver Geometric Cloud-Hook Band (r in [0.418w, 0.452w])
	rimHooks := sym * 3
	dRimAng := (2.0 * math.Pi) / float64(rimHooks)
	for i := 0; i < rimHooks; i++ {
		th0 := float64(i) * dRimAng
		th1 := th0 + dRimAng*0.86
		thMid := 0.5 * (th0 + th1)

		rIn := 0.420 * w
		rOut := 0.448 * w
		pts := []render.Point{
			render.Pt(cx+rIn*math.Cos(th0), cy+rIn*math.Sin(th0)),
			render.Pt(cx+rOut*math.Cos(thMid), cy+rOut*math.Sin(thMid)),
			render.Pt(cx+rIn*math.Cos(th1), cy+rIn*math.Sin(th1)),
		}
		col := inlayGold
		if i%2 == 1 {
			col = inlaySilver
		}
		rc.FillPolygon(pts, col, 0.95)
	}

	// 4. Main Warring States Pan-Chi & Cloud-Thunder Inlaid Scroll Band (r in [0.165w, 0.405w])
	dAng := (2.0 * math.Pi) / float64(sym)
	rBandIn := 0.168 * w
	rBandOut := 0.398 * w
	rBandMid := 0.5 * (rBandIn + rBandOut)

	// Background concentric bronze & verdigris striation rings inside the mirror field
	for ring := 0; ring <= 14; ring++ {
		u := float64(ring) / 14.0
		r := rBandIn + u*(rBandOut-rBandIn)
		col := agedBronze
		if ring%3 == 0 {
			col = verdigrisPatina
		}
		rc.DrawSoftRing(cx, cy, r, 1.2*scale, col, 0.42)
	}

	for s := 0; s < sym; s++ {
		baseTh := float64(s) * dAng
		thCenter := baseTh + 0.5*dAng

		// Interlocking S-curved Gold Inlay Dragon-Cloud Scroll (错金蟠螭主纹)
		sPts := make([]render.Point, 0, 28)
		for k := 0; k <= 26; k++ {
			u := float64(k) / 26.0
			r := rBandIn + (rBandOut-rBandIn)*(0.10+0.80*u+0.12*math.Sin(u*math.Pi*2.0))
			th := baseTh + dAng*(0.06+0.88*u+0.18*math.Sin(u*math.Pi*2.0))
			sPts = append(sPts, render.Pt(cx+r*math.Cos(th), cy+r*math.Sin(th)))
		}
		rc.DrawCapsule(sPts, 7.2*scale, goldShadow, 0.95)
		rc.DrawTaperedSpline(sPts, 5.6*scale, render.TaperCalligraphic, func(t float64) color.RGBA {
			return render.LerpOklab(inlayGold, goldSpec, 0.35*math.Sin(t*math.Pi))
		}, 1.0)

		// Secondary outer-wing Gold Inlay C-scroll (错金羽翼卷云纹)
		wingPts := make([]render.Point, 0, 20)
		for k := 0; k <= 18; k++ {
			u := float64(k) / 18.0
			r := rBandMid + (rBandOut-rBandMid)*(0.25+0.62*math.Sin(u*math.Pi))
			th := baseTh + dAng*(0.18+0.64*u)
			wingPts = append(wingPts, render.Pt(cx+r*math.Cos(th), cy+r*math.Sin(th)))
		}
		rc.DrawCapsule(wingPts, 5.2*scale, goldShadow, 0.92)
		rc.DrawTaperedSpline(wingPts, 3.8*scale, render.TaperBothEnds, func(t float64) color.RGBA {
			return inlayGold
		}, 0.96)

		// Counter-scrolling Sterling Silver Inlay Hook (错银几何钩连纹)
		agPts := make([]render.Point, 0, 22)
		for k := 0; k <= 20; k++ {
			u := float64(k) / 20.0
			r := rBandOut - (rBandOut-rBandIn)*(0.12+0.76*u)
			th := baseTh + dAng*(0.12+0.76*u-0.16*math.Sin(u*math.Pi*2.0))
			agPts = append(agPts, render.Pt(cx+r*math.Cos(th), cy+r*math.Sin(th)))
		}
		rc.DrawTaperedSpline(agPts, 3.8*scale, render.TaperBothEnds, func(t float64) color.RGBA {
			return inlaySilver
		}, 0.95)

		// Fine Square-Spiral Yunlei (云雷纹) Ground Filling inside the Sector
		for _, rSpiral := range []float64{rBandIn + 0.24*(rBandOut-rBandIn), rBandIn + 0.76*(rBandOut-rBandIn)} {
			spCenter := render.Pt(cx+rSpiral*math.Cos(thCenter), cy+rSpiral*math.Sin(thCenter))
			yunlei := &canvas.Path{}
			for k := 0; k <= 24; k++ {
				u := float64(k) / 24.0
				a := baseTh + u*3.5*math.Pi
				rad := (1.0 - u*0.80) * 16.5 * scale
				pt := render.Pt(spCenter.X+rad*math.Cos(a), spCenter.Y+rad*math.Sin(a))
				if k == 0 {
					yunlei.MoveTo(pt.X, pt.Y)
				} else {
					yunlei.LineTo(pt.X, pt.Y)
				}
			}
			rc.StrokePath(yunlei, 1.45*scale, canvas.RoundCap, canvas.RoundJoin, inlayGold, 0.88)
		}

		// Inlaid Cabochon Turquoise & Cinnabar Lacquer Lozenge Medallions (嵌绿松石与朱漆菱形块)
		lozCenter := render.Pt(cx+rBandMid*math.Cos(baseTh), cy+rBandMid*math.Sin(baseTh))
		radVec := render.Pt(math.Cos(baseTh), math.Sin(baseTh))
		tanVec := radVec.Perp()
		lozPts := []render.Point{
			lozCenter.Add(radVec.Mul(30.0 * scale)),
			lozCenter.Add(tanVec.Mul(16.0 * scale)),
			lozCenter.Sub(radVec.Mul(30.0 * scale)),
			lozCenter.Sub(tanVec.Mul(16.0 * scale)),
		}
		if s%2 == 0 {
			rc.DrawCloisonneCell(lozPts, turquoiseGem, turquoiseRim, paleTurquoise, goldShadow, inlayGold, goldSpec, 1.75*scale)
		} else {
			rc.DrawCloisonneCell(lozPts, cinnabarLacquer, silkBacking, goldSpec, goldShadow, inlayGold, goldSpec, 1.75*scale)
		}

		// Outer Ring Turquoise Cabochon Studs (乳丁纹嵌绿松石)
		studCenter := render.Pt(cx+0.365*w*math.Cos(thCenter), cy+0.365*w*math.Sin(thCenter))
		studSDF := render.SDFCircle(studCenter, 8.5*scale)
		rc.RenderSDFBevel(studSDF, turquoiseGem, turquoiseRim, paleTurquoise, 4.5*scale, render.Pt(-0.7, -0.7), 1.0)
		rc.DrawSoftRing(studCenter.X, studCenter.Y, 9.0*scale, 1.6*scale, inlayGold, 0.95)
	}

	// 5. Inner Quatrefoil Persimmon-Calyx Seat (四叶柿蒂纹钮座) & Central Perforated Bronze Knob (弦纹桥钮)
	rSeat := 0.155 * w
	rc.DrawSoftRing(cx, cy, rSeat*1.05, 2.8*scale, inlayGold, 0.96)
	rc.DrawSoftRing(cx, cy, rSeat*0.98, 1.6*scale, inlaySilver, 0.90)

	const calyxPetals = 4
	for p := 0; p < calyxPetals; p++ {
		thC := float64(p) * 0.5 * math.Pi
		pts := []render.Point{
			render.Pt(cx+0.045*w*math.Cos(thC-0.42), cy+0.045*w*math.Sin(thC-0.42)),
			render.Pt(cx+0.115*w*math.Cos(thC-0.36), cy+0.115*w*math.Sin(thC-0.36)),
			render.Pt(cx+rSeat*0.92*math.Cos(thC), cy+rSeat*0.92*math.Sin(thC)),
			render.Pt(cx+0.115*w*math.Cos(thC+0.36), cy+0.115*w*math.Sin(thC+0.36)),
			render.Pt(cx+0.045*w*math.Cos(thC+0.42), cy+0.045*w*math.Sin(thC+0.42)),
		}
		rc.DrawCloisonneCell(pts, turquoiseGem, turquoiseRim, paleTurquoise, goldShadow, inlayGold, goldSpec, 1.7*scale)
	}

	// Central Hemispherical Bronze Knob (镜钮)
	knobSDF := render.SDFCircle(render.Pt(cx, cy), 0.052*w)
	rc.RenderSDFDropShadow(knobSDF, render.Pt(3*scale, 4*scale), 8*scale, 1*scale, silkBacking, 0.9)
	rc.RenderSDFBevel(knobSDF, inlayGold, goldShadow, goldSpec, 0.028*w, render.Pt(-0.7, -0.7), 1.0)

	c.MergeRenderCanvas(rc)
}
