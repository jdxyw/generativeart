package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type dunhuangApsara struct {
	ribbons   int
	haloRings int
}

// NewDunhuangApsara returns a dunhuangApsara engine that renders a Tang-Dynasty
// Dunhuang Mogao Cave caisson ceiling (敦煌莫高窟藻井) with recessed square-diamond
// coffers, Baoxianghua lotus medallions, pearl chains, and flying celestial silk ribbons.
func NewDunhuangApsara(ribbons, haloRings int) *dunhuangApsara {
	if ribbons <= 0 {
		ribbons = 8
	}
	if haloRings <= 0 {
		haloRings = 5
	}
	return &dunhuangApsara{
		ribbons:   ribbons,
		haloRings: haloRings,
	}
}

// Generative renders the Dunhuang Mogao caisson ceiling and flying silk ribbons.
func (da *dunhuangApsara) Generative(c *generativeart.Canva) {
	da.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360-degree phase loop (phase in [0, 1)).
func (da *dunhuangApsara) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, loopPhase float64) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0
	cx, cy := 0.5*w, 0.5*h
	animPhase := loopPhase * 2.0 * math.Pi

	// Authentic Dunhuang Mineral Pigment Palette (石青, 石绿, 朱砂, 赭石, 泥金, 胡粉)
	grottoDark := color.RGBA{R: 20, G: 14, B: 18, A: 255}
	mineralLapis := color.RGBA{R: 26, G: 66, B: 138, A: 255}
	deepLapis := color.RGBA{R: 14, G: 32, B: 74, A: 255}
	malachiteGreen := color.RGBA{R: 42, G: 146, B: 114, A: 255}
	turquoiseJade := color.RGBA{R: 58, G: 176, B: 168, A: 255}
	cinnabarRed := color.RGBA{R: 198, G: 56, B: 42, A: 255}
	burntOchre := color.RGBA{R: 156, G: 68, B: 34, A: 255}
	mudGold := color.RGBA{R: 238, G: 184, B: 68, A: 255}
	chalkIvory := color.RGBA{R: 248, G: 238, B: 212, A: 255}
	wireShadow := color.RGBA{R: 48, G: 28, B: 16, A: 255}

	rc.FillBackground(grottoDark)
	rc.SetBlendMode(render.BlendOver)

	// Helper to draw a rotated square coffer tier with SDF drop shadow, flame/cloud border, and pearl chain
	drawCofferSquare := func(halfSide, rotRad float64, fillCol, borderCol, accentCol color.RGBA) {
		cosA := math.Cos(rotRad)
		sinA := math.Sin(rotRad)
		rotPt := func(lx, ly float64) render.Point {
			return render.Pt(cx+lx*cosA-ly*sinA, cy+lx*sinA+ly*cosA)
		}

		outerPts := []render.Point{
			rotPt(-halfSide, -halfSide),
			rotPt(halfSide, -halfSide),
			rotPt(halfSide, halfSide),
			rotPt(-halfSide, halfSide),
		}
		cofferSDF := render.SDFPolygon(outerPts)
		rc.RenderSDFDropShadow(cofferSDF, render.Pt(5*scale, 6*scale), 14*scale, 2*scale, grottoDark, 0.88)
		rc.FillPolygon(outerPts, borderCol, 1.0)

		// Inner recessed coffer panel
		inS := halfSide * 0.86
		innerPts := []render.Point{
			rotPt(-inS, -inS),
			rotPt(inS, -inS),
			rotPt(inS, inS),
			rotPt(-inS, inS),
		}
		rc.FillPolygon(innerPts, fillCol, 1.0)

		// Decorative Tang honeysuckle / triangle-flame valence teeth along the 4 edges
		teeth := 14
		for side := 0; side < 4; side++ {
			pA := outerPts[side]
			pB := outerPts[(side+1)%4]
			pInA := innerPts[side]
			pInB := innerPts[(side+1)%4]
			for t := 0; t < teeth; t++ {
				u0 := float64(t) / float64(teeth)
				u1 := float64(t+1) / float64(teeth)
				uMid := 0.5 * (u0 + u1)
				tri := []render.Point{
					pA.Lerp(pB, u0),
					pA.Lerp(pB, u1),
					pInA.Lerp(pInB, uMid),
				}
				col := accentCol
				if t%2 == 1 {
					col = mudGold
				}
				rc.FillPolygon(render.InsetPolygon(tri, 0.5*scale), col, 0.95)
			}
		}

		// Gilded double framing lines + Tang pearl chain (联珠纹)
		frame := &canvas.Path{}
		frame.MoveTo(outerPts[0].X, outerPts[0].Y)
		for i := 1; i < 4; i++ {
			frame.LineTo(outerPts[i].X, outerPts[i].Y)
		}
		frame.Close()
		frame.MoveTo(innerPts[0].X, innerPts[0].Y)
		for i := 1; i < 4; i++ {
			frame.LineTo(innerPts[i].X, innerPts[i].Y)
		}
		frame.Close()
		rc.StrokePath(frame, 1.8*scale, canvas.RoundCap, canvas.RoundJoin, mudGold, 0.95)
	}

	// 1. Recessed Square-and-Diamond Coffer Frames (方井叠梲: 0°, 45°, 0°, 45°)
	drawCofferSquare(0.465*w, 0.0, deepLapis, cinnabarRed, malachiteGreen)
	drawCofferSquare(0.385*w, math.Pi/4.0, burntOchre, mineralLapis, turquoiseJade)
	drawCofferSquare(0.312*w, 0.0, deepLapis, malachiteGreen, cinnabarRed)

	// 2. Sweeping Celestial Flying Silk Ribbons (飞天吴带当风长绸) Weaving Around the Medallion
	numRibbons := da.ribbons
	if numRibbons < 4 {
		numRibbons = 8
	}
	ribbonPal := []color.RGBA{cinnabarRed, mudGold, turquoiseJade, malachiteGreen}
	for rb := 0; rb < numRibbons; rb++ {
		baseAng := (float64(rb) / float64(numRibbons)) * 2.0 * math.Pi
		pts := make([]render.Point, 0, 28)
		for k := 0; k <= 26; k++ {
			u := float64(k) / 26.0
			r := (0.20 + 0.25*u + 0.014*math.Sin(u*math.Pi*2.0-animPhase)) * w
			// Sinuous S-wave oscillation travelling seamlessly along the spiral silk ribbon
			theta := baseAng + u*1.15*math.Pi + 0.14*math.Sin(u*math.Pi*4.0-animPhase*2.0)
			pts = append(pts, render.Pt(cx+r*math.Cos(theta), cy+r*math.Sin(theta)))
		}
		col := ribbonPal[rb%len(ribbonPal)]
		rc.DrawImpastoStroke(pts, 10.5*scale, col, render.Pt(-0.65, -0.75), 0.94)
		// Gilded silk selvedge edge
		rc.DrawTaperedSpline(pts, 2.2*scale, render.TaperBothEnds, func(t float64) color.RGBA {
			return mudGold
		}, 0.92)
	}

	// 3. Central Baoxianghua Lotus Mandala (盛唐八瓣重层宝相花藻井心)
	rMandala := 0.235 * w
	// Pearl-chain halo ring (联珠纹圆环)
	discPts := make([]render.Point, 64)
	for i := 0; i < 64; i++ {
		a := (float64(i) / 64.0) * 2.0 * math.Pi
		discPts[i] = render.Pt(cx+rMandala*1.04*math.Cos(a), cy+rMandala*1.04*math.Sin(a))
	}
	rc.FillPolygon(discPts, grottoDark, 1.0)
	rc.DrawSoftRing(cx, cy, rMandala*1.01, 4.5*scale, mudGold, 0.96)

	const pearls = 48
	pearlRot := animPhase / float64(pearls)
	for i := 0; i < pearls; i++ {
		a := (float64(i)/float64(pearls))*2.0*math.Pi + pearlRot
		px := cx + rMandala*0.95*math.Cos(a)
		py := cy + rMandala*0.95*math.Sin(a)
		rc.DrawGlowDot(px, py, 2.8*scale, chalkIvory, 0.95, 0.88)
	}

	// 3 concentric rings of sculpted Baoxianghua lotus petals (12 petals with p%2 alternation -> period = 2*dAng = 2*pi/6)
	const petals = 12
	dAng := (2.0 * math.Pi) / float64(petals)
	lotusStepRot := animPhase / 6.0

	drawBaoxiangRing := func(rIn, rOut, phase float64, baseCol, rimCol, specCol color.RGBA) {
		for p := 0; p < petals; p++ {
			thC := float64(p)*dAng + phase
			span := dAng * 0.46
			pts := []render.Point{
				render.Pt(cx+rIn*math.Cos(thC-span*0.65), cy+rIn*math.Sin(thC-span*0.65)),
				render.Pt(cx+(rIn+(rOut-rIn)*0.45)*math.Cos(thC-span), cy+(rIn+(rOut-rIn)*0.45)*math.Sin(thC-span)),
				render.Pt(cx+(rIn+(rOut-rIn)*0.78)*math.Cos(thC-span*0.72), cy+(rIn+(rOut-rIn)*0.78)*math.Sin(thC-span*0.72)),
				render.Pt(cx+rOut*math.Cos(thC), cy+rOut*math.Sin(thC)),
				render.Pt(cx+(rIn+(rOut-rIn)*0.78)*math.Cos(thC+span*0.72), cy+(rIn+(rOut-rIn)*0.78)*math.Sin(thC+span*0.72)),
				render.Pt(cx+(rIn+(rOut-rIn)*0.45)*math.Cos(thC+span), cy+(rIn+(rOut-rIn)*0.45)*math.Sin(thC+span)),
				render.Pt(cx+rIn*math.Cos(thC+span*0.65), cy+rIn*math.Sin(thC+span*0.65)),
			}
			bCol := baseCol
			if p%2 == 1 {
				bCol = render.LerpOklab(baseCol, mudGold, 0.28)
			}
			rc.DrawCloisonneCell(pts, bCol, rimCol, specCol, wireShadow, mudGold, chalkIvory, 1.55*scale)
		}
	}

	drawBaoxiangRing(0.135*w, rMandala*0.90, lotusStepRot, cinnabarRed, burntOchre, chalkIvory)
	drawBaoxiangRing(0.078*w, 0.162*w, dAng*0.5-lotusStepRot, turquoiseJade, deepLapis, chalkIvory)
	drawBaoxiangRing(0.032*w, 0.095*w, lotusStepRot, mudGold, burntOchre, chalkIvory)

	// Central Golden Lotus Seed Pod
	coreOrb := render.SDFCircle(render.Pt(cx, cy), 0.036*w)
	lightDir := render.Pt(-0.7*math.Cos(animPhase)+0.7*math.Sin(animPhase), -0.7*math.Sin(animPhase)-0.7*math.Cos(animPhase))
	rc.RenderSDFBevel(coreOrb, mudGold, burntOchre, chalkIvory, 0.018*w, lightDir, 1.0)

	rc.ApplyPaperTexture(0.07, 202622)
	c.MergeRenderCanvas(rc)
}
