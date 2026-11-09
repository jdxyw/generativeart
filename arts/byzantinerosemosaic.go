package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type byzantineRoseMosaic struct {
	petals       int
	rings        int
	ribbonSpiral float64
}

// NewByzantineRoseMosaic returns a byzantineRoseMosaic engine that synthesizes a 3D beveled
// Cloisonné & Stained-Glass Celestial Star-Rose Medallion intertwined with sweeping 3D gilded silk ribbons.
func NewByzantineRoseMosaic(petals, rings int, ribbonSpiral float64) *byzantineRoseMosaic {
	if petals <= 0 {
		petals = 12
	}
	if rings <= 0 {
		rings = 4
	}
	if ribbonSpiral <= 0 {
		ribbonSpiral = 1.15
	}
	return &byzantineRoseMosaic{
		petals:       common.ConstrainInt(petals, 8, 20),
		rings:        common.ConstrainInt(rings, 3, 6),
		ribbonSpiral: common.Constrain(ribbonSpiral, 0.5, 2.2),
	}
}

// Generative renders the static ByzantineRoseMosaic composition.
func (brm *byzantineRoseMosaic) Generative(c *generativeart.Canva) {
	brm.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360-degree phase loop (phase in [0, 1)).
func (brm *byzantineRoseMosaic) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, loopPhase float64) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0
	cx, cy := 0.5*w, 0.5*h
	animPhase := loopPhase * 2.0 * math.Pi

	schema := c.Opts().ColorSchema()
	if len(schema) == 0 {
		schema = common.Dunhuang
	}
	palGrad := render.NewMultiStopGradient(schema...)

	obsidianVault := color.RGBA{R: 12, G: 14, B: 24, A: 255}
	deepSapphire := render.LerpOklab(color.RGBA{R: 14, G: 32, B: 82, A: 255}, palGrad.SampleOklab(0.12), 0.25)
	royalLapis := render.LerpOklab(color.RGBA{R: 28, G: 78, B: 162, A: 255}, palGrad.SampleOklab(0.28), 0.28)
	emeraldJade := render.LerpOklab(color.RGBA{R: 32, G: 158, B: 124, A: 255}, palGrad.SampleOklab(0.45), 0.28)
	turquoiseGlass := render.LerpOklab(color.RGBA{R: 52, G: 196, B: 192, A: 255}, palGrad.SampleOklab(0.58), 0.25)
	cinnabarRuby := render.LerpOklab(color.RGBA{R: 208, G: 48, B: 52, A: 255}, palGrad.SampleOklab(0.75), 0.26)
	imperialAmber := render.LerpOklab(color.RGBA{R: 224, G: 126, B: 38, A: 255}, palGrad.SampleOklab(0.86), 0.25)
	burnishedGold := color.RGBA{R: 244, G: 194, B: 72, A: 255}
	pearlIvory := color.RGBA{R: 252, G: 244, B: 224, A: 255}
	wireShadow := color.RGBA{R: 32, G: 20, B: 14, A: 255}

	rc.FillBackground(obsidianVault)
	rc.SetBlendMode(render.BlendOver)

	// 1. Recessed 8-Pointed Star-Octagon Coffered Sanctuary Frames (Rub el Hizb / Tang Caisson geometry)
	drawOctagonTier := func(halfSide, rotRad float64, fillCol, borderCol, accentCol color.RGBA) {
		for k := 0; k < 2; k++ {
			a := rotRad + float64(k)*(math.Pi/4.0)
			cosA, sinA := math.Cos(a), math.Sin(a)
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
			rc.RenderSDFDropShadow(cofferSDF, render.Pt(4.5*scale, 5.5*scale), 13*scale, 2*scale, obsidianVault, 0.86)
			rc.FillPolygon(outerPts, borderCol, 1.0)

			inS := halfSide * 0.87
			innerPts := []render.Point{
				rotPt(-inS, -inS),
				rotPt(inS, -inS),
				rotPt(inS, inS),
				rotPt(-inS, inS),
			}
			rc.FillPolygon(innerPts, fillCol, 1.0)

			// Cloisonné mosaic border triangles along all 4 sides
			const teeth = 12
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
						col = burnishedGold
					}
					rc.FillPolygon(render.InsetPolygon(tri, 0.5*scale), col, 0.95)
				}
			}

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
			rc.StrokePath(frame, 1.75*scale, canvas.RoundCap, canvas.RoundJoin, burnishedGold, 0.95)
		}
	}

	drawOctagonTier(0.455*w, 0.0, deepSapphire, cinnabarRuby, emeraldJade)
	drawOctagonTier(0.355*w, math.Pi/8.0, obsidianVault, royalLapis, turquoiseGlass)

	// 2. Corner Beveled Stained-Glass Quatrefoil Jewels in the 4 Outer Spandrels
	numPetals := brm.petals
	if numPetals%2 != 0 {
		numPetals++
	}

	// 3. Sweeping 3D Celestial Silk Ribbons Weaving Between the Star Coffer and the Rose Window
	ribbonColors := []color.RGBA{cinnabarRuby, burnishedGold, turquoiseGlass, emeraldJade}
	numRibbons := common.ConstrainInt(numPetals*2/3, 6, 12)
	for rb := 0; rb < numRibbons; rb++ {
		baseAng := (float64(rb) / float64(numRibbons)) * 2.0 * math.Pi
		pts := make([]render.Point, 0, 32)
		for k := 0; k <= 30; k++ {
			u := float64(k) / 30.0
			r := (0.21 + 0.23*u + 0.015*math.Sin(u*math.Pi*3.0-animPhase)) * w
			theta := baseAng + u*brm.ribbonSpiral*math.Pi + 0.13*math.Sin(u*math.Pi*4.0-animPhase*2.0)
			pts = append(pts, render.Pt(cx+r*math.Cos(theta), cy+r*math.Sin(theta)))
		}
		col := ribbonColors[rb%len(ribbonColors)]
		rc.DrawImpastoStroke(pts, 9.5*scale, col, render.Pt(-0.65, -0.75), 0.94)
		rc.DrawTaperedSpline(pts, 2.0*scale, render.TaperBothEnds, func(t float64) color.RGBA {
			return burnishedGold
		}, 0.92)
	}

	// 4. Central Backlit Gothic & Cloisonné Rose Window Medallion
	rRose := 0.258 * w
	discPts := make([]render.Point, 72)
	for i := 0; i < 72; i++ {
		a := (float64(i) / 72.0) * 2.0 * math.Pi
		discPts[i] = render.Pt(cx+rRose*1.04*math.Cos(a), cy+rRose*1.04*math.Sin(a))
	}
	discSDF := render.SDFCircle(render.Pt(cx, cy), rRose*1.04)
	rc.RenderSDFDropShadow(discSDF, render.Pt(4*scale, 5*scale), 12*scale, 2*scale, obsidianVault, 0.90)
	rc.FillPolygon(discPts, obsidianVault, 1.0)
	rc.DrawSoftRing(cx, cy, rRose*1.01, 4.2*scale, burnishedGold, 0.96)

	// Luminous Pearl Halo Chain around the Rose Medallion
	const pearls = 56
	pearlRot := animPhase / float64(pearls)
	for i := 0; i < pearls; i++ {
		a := (float64(i)/float64(pearls))*2.0*math.Pi + pearlRot
		px := cx + rRose*0.955*math.Cos(a)
		py := cy + rRose*0.955*math.Sin(a)
		rc.DrawGlowDot(px, py, 2.6*scale, pearlIvory, 0.95, 0.88)
	}

	dAng := (2.0 * math.Pi) / float64(numPetals)
	stepRot := animPhase / float64(numPetals/2)

	drawCloisonnePetalRing := func(rIn, rOut, phaseOff float64, colEven, colOdd, rimCol color.RGBA) {
		for p := 0; p < numPetals; p++ {
			thC := float64(p)*dAng + phaseOff
			span := dAng * 0.46
			pts := []render.Point{
				render.Pt(cx+rIn*math.Cos(thC-span*0.62), cy+rIn*math.Sin(thC-span*0.62)),
				render.Pt(cx+(rIn+(rOut-rIn)*0.44)*math.Cos(thC-span), cy+(rIn+(rOut-rIn)*0.44)*math.Sin(thC-span)),
				render.Pt(cx+(rIn+(rOut-rIn)*0.78)*math.Cos(thC-span*0.70), cy+(rIn+(rOut-rIn)*0.78)*math.Sin(thC-span*0.70)),
				render.Pt(cx+rOut*math.Cos(thC), cy+rOut*math.Sin(thC)),
				render.Pt(cx+(rIn+(rOut-rIn)*0.78)*math.Cos(thC+span*0.70), cy+(rIn+(rOut-rIn)*0.78)*math.Sin(thC+span*0.70)),
				render.Pt(cx+(rIn+(rOut-rIn)*0.44)*math.Cos(thC+span), cy+(rIn+(rOut-rIn)*0.44)*math.Sin(thC+span)),
				render.Pt(cx+rIn*math.Cos(thC+span*0.62), cy+rIn*math.Sin(thC+span*0.62)),
			}
			bCol := colEven
			if p%2 == 1 {
				bCol = colOdd
			}
			rc.DrawCloisonneCell(pts, bCol, rimCol, pearlIvory, wireShadow, burnishedGold, pearlIvory, 1.50*scale)
		}
	}

	numRings := brm.rings
	rMin := 0.038 * w
	rMax := rRose * 0.90
	for rIdx := numRings - 1; rIdx >= 0; rIdx-- {
		t0 := float64(rIdx) / float64(numRings)
		t1 := float64(rIdx+1) / float64(numRings)
		rIn := rMin + (rMax-rMin)*math.Pow(t0, 0.92)
		rOut := rMin + (rMax-rMin)*math.Pow(t1, 0.92)*1.06
		rot := stepRot
		if rIdx%2 == 1 {
			rot = dAng*0.5 - stepRot
		}
		switch rIdx % 4 {
		case 0:
			drawCloisonnePetalRing(rIn, rOut, rot, burnishedGold, imperialAmber, cinnabarRuby)
		case 1:
			drawCloisonnePetalRing(rIn, rOut, rot, turquoiseGlass, emeraldJade, deepSapphire)
		case 2:
			drawCloisonnePetalRing(rIn, rOut, rot, cinnabarRuby, imperialAmber, deepSapphire)
		default:
			drawCloisonnePetalRing(rIn, rOut, rot, royalLapis, turquoiseGlass, obsidianVault)
		}
	}

	// Central 3D Beveled Solar Cabochon Jewel
	coreOrb := render.SDFCircle(render.Pt(cx, cy), 0.038*w)
	lightDir := render.Pt(-0.7*math.Cos(animPhase)+0.7*math.Sin(animPhase), -0.7*math.Sin(animPhase)-0.7*math.Cos(animPhase))
	rc.RenderSDFBevel(coreOrb, burnishedGold, imperialAmber, pearlIvory, 0.019*w, lightDir, 1.0)

	rc.ApplyPaperTexture(0.06, 20262504)
	c.MergeRenderCanvas(rc)
}
