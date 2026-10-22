package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type persianTilework struct {
	symmetry int
	rings    int
}

// NewPersianTilework returns a persianTilework engine that renders an Isfahan-inspired
// Persian Muqarnas (stalactite vault) dome and Haft-Rangi glazed tile mandala with
// 16-fold Shamsa starbursts, pointed lancet niches, and golden Islimi arabesques.
func NewPersianTilework(symmetry, rings int) *persianTilework {
	if symmetry <= 0 {
		symmetry = 16
	}
	if rings <= 0 {
		rings = 6
	}
	return &persianTilework{
		symmetry: symmetry,
		rings:    rings,
	}
}

// Generative renders the Isfahan Muqarnas dome and Haft-Rangi glazed tilework.
func (pt *persianTilework) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0
	cx, cy := 0.5*w, 0.5*h

	// Authentic Persian Haft-Rangi (7-Color Glazed Tile) Palette
	aubergineDark := color.RGBA{R: 16, G: 12, B: 24, A: 255}
	lajvardLapis := color.RGBA{R: 22, G: 58, B: 138, A: 255}
	cobaltGlaze := color.RGBA{R: 36, G: 98, B: 184, A: 255}
	firoozehTeal := color.RGBA{R: 38, G: 174, B: 182, A: 255}
	paleTurquoise := color.RGBA{R: 164, G: 238, B: 234, A: 255}

	saffronGold := color.RGBA{R: 242, G: 178, B: 52, A: 255}
	amberOchre := color.RGBA{R: 196, G: 112, B: 32, A: 255}
	persianCoral := color.RGBA{R: 208, G: 64, B: 52, A: 255}
	ivoryGrout := color.RGBA{R: 252, G: 244, B: 222, A: 255}

	rc.FillBackground(aubergineDark)
	rc.SetBlendMode(render.BlendOver)

	sym := pt.symmetry
	if sym < 8 {
		sym = 16
	}
	numRings := pt.rings
	if numRings < 4 {
		numRings = 6
	}

	dAng := (2.0 * math.Pi) / float64(sym)
	maxR := math.Hypot(w*0.5, h*0.5) * 1.04

	// 1. Render Concentric Muqarnas Vault Tiers from Outer Rim Inward to Central Oculus
	for tier := numRings; tier >= 1; tier-- {
		u0 := float64(tier-1) / float64(numRings)
		u1 := float64(tier) / float64(numRings)
		rIn := (0.06 + 0.88*math.Pow(u0, 1.08)) * maxR * 0.74
		rOut := (0.06 + 0.88*math.Pow(u1, 1.08)) * maxR * 0.74
		rMid := 0.5 * (rIn + rOut)
		phase := float64(tier%2) * dAng * 0.5

		for s := 0; s < sym; s++ {
			th0 := float64(s)*dAng + phase
			th1 := th0 + dAng
			thMid := 0.5 * (th0 + th1)

			// Interstitial recessed Lajvard diamond spandrel between Muqarnas niches
			thValley := th1
			spandrelPts := []render.Point{
				render.Pt(cx+rMid*math.Cos(thMid), cy+rMid*math.Sin(thMid)),
				render.Pt(cx+rOut*0.96*math.Cos(thValley-dAng*0.18), cy+rOut*0.96*math.Sin(thValley-dAng*0.18)),
				render.Pt(cx+rOut*1.08*math.Cos(thValley), cy+rOut*1.08*math.Sin(thValley)),
				render.Pt(cx+rOut*0.96*math.Cos(thValley+dAng*0.18), cy+rOut*0.96*math.Sin(thValley+dAng*0.18)),
				render.Pt(cx+rMid*math.Cos(thMid+dAng), cy+rMid*math.Sin(thMid+dAng)),
				render.Pt(cx+rIn*1.05*math.Cos(thValley), cy+rIn*1.05*math.Sin(thValley)),
			}
			rc.DrawCloisonneCell(
				spandrelPts,
				lajvardLapis, aubergineDark, cobaltGlaze,
				aubergineDark, saffronGold, ivoryGrout, 1.35*scale,
			)

			// Sculpted Muqarnas Pointed Lancet Niche (divided into left/right 3D vault facets)
			pBaseL := render.Pt(cx+rIn*math.Cos(th0+dAng*0.06), cy+rIn*math.Sin(th0+dAng*0.06))
			pBaseC := render.Pt(cx+rIn*0.96*math.Cos(thMid), cy+rIn*0.96*math.Sin(thMid))
			pBaseR := render.Pt(cx+rIn*math.Cos(th1-dAng*0.06), cy+rIn*math.Sin(th1-dAng*0.06))
			pShL := render.Pt(cx+rMid*1.04*math.Cos(th0+dAng*0.12), cy+rMid*1.04*math.Sin(th0+dAng*0.12))
			pShR := render.Pt(cx+rMid*1.04*math.Cos(th1-dAng*0.12), cy+rMid*1.04*math.Sin(th1-dAng*0.12))
			pApex := render.Pt(cx+rOut*math.Cos(thMid), cy+rOut*math.Sin(thMid))
			pVaultCenter := render.Pt(cx+rMid*0.96*math.Cos(thMid), cy+rMid*0.96*math.Sin(thMid))

			// Left sunlit Muqarnas vault facet & right shaded vault facet
			leftVault := []render.Point{pBaseC, pBaseL, pShL, pApex, pVaultCenter}
			rightVault := []render.Point{pBaseC, pVaultCenter, pApex, pShR, pBaseR}

			var lBase, rBase, specCol color.RGBA
			switch (tier + s) % 3 {
			case 0:
				lBase, rBase, specCol = firoozehTeal, cobaltGlaze, paleTurquoise
			case 1:
				lBase, rBase, specCol = saffronGold, amberOchre, ivoryGrout
			default:
				lBase, rBase, specCol = cobaltGlaze, lajvardLapis, paleTurquoise
			}
			if tier == 2 && s%2 == 0 {
				lBase, rBase, specCol = persianCoral, amberOchre, saffronGold
			}

			rc.DrawCloisonneCell(leftVault, lBase, aubergineDark, specCol, aubergineDark, ivoryGrout, ivoryGrout, 1.3*scale)
			rc.DrawCloisonneCell(rightVault, rBase, aubergineDark, lBase, aubergineDark, ivoryGrout, ivoryGrout, 1.3*scale)

			// Golden Islimi arabesque vine inside outer tiers
			if tier >= 3 {
				vine := &canvas.Path{}
				for k := 0; k <= 16; k++ {
					t := float64(k) / 16.0
					rV := rIn + (rOut-rIn)*(0.20+0.62*t)
					angV := thMid + math.Sin(t*math.Pi*2.0)*dAng*0.18
					px, py := cx+rV*math.Cos(angV), cy+rV*math.Sin(angV)
					if k == 0 {
						vine.MoveTo(px, py)
					} else {
						vine.LineTo(px, py)
					}
				}
				rc.StrokePath(vine, 1.0*scale, canvas.RoundCap, canvas.RoundJoin, saffronGold, 0.88)
			}
		}
	}

	// 2. Central 16-Pointed Shamsa Sunburst Star Oculus
	rCore := 0.088 * w
	starPts := make([]render.Point, sym*2)
	for i := 0; i < sym*2; i++ {
		a := (float64(i) / float64(sym*2)) * 2.0 * math.Pi
		r := rCore
		if i%2 == 1 {
			r = rCore * 0.56
		}
		starPts[i] = render.Pt(cx+r*math.Cos(a), cy+r*math.Sin(a))
	}
	rc.DrawCloisonneCell(starPts, saffronGold, amberOchre, ivoryGrout, aubergineDark, ivoryGrout, ivoryGrout, 1.65*scale)

	innerStar := make([]render.Point, sym*2)
	for i := 0; i < sym*2; i++ {
		a := (float64(i)/float64(sym*2))*2.0*math.Pi + dAng*0.5
		r := rCore * 0.46
		if i%2 == 1 {
			r = rCore * 0.24
		}
		innerStar[i] = render.Pt(cx+r*math.Cos(a), cy+r*math.Sin(a))
	}
	rc.DrawCloisonneCell(innerStar, firoozehTeal, lajvardLapis, paleTurquoise, aubergineDark, saffronGold, ivoryGrout, 1.4*scale)

	c.MergeRenderCanvas(rc)
}
