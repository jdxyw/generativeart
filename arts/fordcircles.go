package arts

import (
	"image/color"
	"math"
	"sort"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type fordCircles struct {
	maxDenom int
}

// NewFordCircles returns a fordCircles engine that renders a conformal modular-group
// Farey-sequence and Ford-circle tangency gasket (法里数列福特相切圆数论分形) where
// circles C(p/q) of radius 1/(2q^2) are mutually tangent iff |p*s - q*r| = 1.
func NewFordCircles(maxDenom int) *fordCircles {
	if maxDenom <= 0 {
		maxDenom = 14
	}
	return &fordCircles{
		maxDenom: maxDenom,
	}
}

func gcdInt(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// Generative renders the conformal Ford circles and Farey geodesic network.
func (fc *fordCircles) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0
	cx, cy := 0.5*w, 0.5*h

	abyssBg := color.RGBA{R: 8, G: 10, B: 20, A: 255}
	goldFiligree := color.RGBA{R: 246, G: 194, B: 68, A: 255}
	ivorySpec := color.RGBA{R: 255, G: 250, B: 232, A: 255}

	spherePals := [][2]color.RGBA{
		{color.RGBA{238, 174, 48, 255}, color.RGBA{118, 48, 14, 255}},  // Imperial Gold & Bronze
		{color.RGBA{38, 178, 186, 255}, color.RGBA{12, 62, 88, 255}},   // Jingtai Turquoise & Deep Teal
		{color.RGBA{214, 54, 46, 255}, color.RGBA{96, 16, 22, 255}},    // Cinnabar Vermilion & Ruby
		{color.RGBA{42, 98, 196, 255}, color.RGBA{14, 28, 86, 255}},    // Royal Lapis & Cobalt
		{color.RGBA{46, 168, 118, 255}, color.RGBA{14, 68, 48, 255}},   // Celadon Emerald
		{color.RGBA{176, 48, 128, 255}, color.RGBA{68, 14, 56, 255}},   // Byzantium Magenta
	}

	rc.FillBackground(abyssBg)
	rc.SetBlendMode(render.BlendOver)

	rDisk := 0.445 * math.Min(w, h)
	lightDir := render.Pt(-0.62, -0.74)

	// 1. Deep Cosmic Disk Backdrop & Astrolabe Modular Degree Ring
	for y := 0; y < c.Height(); y++ {
		fy := float64(y) + 0.5
		for x := 0; x < c.Width(); x++ {
			fx := float64(x) + 0.5
			r := math.Hypot(fx-cx, fy-cy) / rDisk
			if r <= 1.02 {
				col := render.LerpOklab(color.RGBA{18, 28, 58, 255}, abyssBg, math.Pow(common.Constrain(r, 0, 1), 1.25))
				rc.BlendPixel(x, y, col, 1.0)
			}
		}
	}

	// Generate Farey sequence fractions p/q in [0, 1) with gcd(p, q) == 1
	maxQ := fc.maxDenom
	if maxQ < 8 {
		maxQ = 14
	}
	if maxQ > 22 {
		maxQ = 22
	}

	type fareyFrac struct {
		p, q int
	}
	type diskSphere struct {
		cx, cy float64
		r      float64
		q      int
		p      int
	}

	// Conformal map from upper half-plane Ford circle (x = theta, y = h) into an annular ring [rIn, rOut]:
	// In polar coordinates around a base circle of radius R_base, a Ford circle at angle theta = 2*pi*(p/q + sector/S)
	// with Farey denominator q has radius proportional to 1 / (q^2)!
	const symmetrySectors = 6
	tiers := []struct {
		rBase   float64
		scaleC  float64
		inward  float64
		rotOff  float64
		maxQCap int
	}{
		{rDisk * 0.96, rDisk * 0.44, -1.0, 0.0, maxQ},               // Outer inward-growing Ford circle crown
		{rDisk * 0.44, rDisk * 0.24, 1.0, math.Pi / 6.0, maxQ - 3},  // Middle outward-growing Ford circle ring
		{rDisk * 0.42, rDisk * 0.23, -1.0, math.Pi / 6.0, maxQ - 4}, // Middle inward-growing Ford circle ring
	}

	spheres := make([]diskSphere, 0, 1200)
	geodesics := &canvas.Path{}

	for tIdx, tier := range tiers {
		qLim := tier.maxQCap
		if qLim < 6 {
			qLim = 6
		}
		fracs := make([]fareyFrac, 0, 128)
		for q := 1; q <= qLim; q++ {
			for p := 0; p < q; p++ {
				if gcdInt(p, q) == 1 {
					fracs = append(fracs, fareyFrac{p: p, q: q})
				}
			}
		}

		for sec := 0; sec < symmetrySectors; sec++ {
			for _, fr := range fracs {
				u := (float64(sec) + float64(fr.p)/float64(fr.q)) / float64(symmetrySectors)
				ang := u*2.0*math.Pi + tier.rotOff

				// Exact Ford radius scaling r_Ford = scaleC / (2 * q^2) with conformal curvature correction
				rF := tier.scaleC / (1.85 * float64(fr.q*fr.q))
				if rF < 1.6*scale {
					continue
				}
				rCenter := tier.rBase + tier.inward*rF
				sx := cx + rCenter*math.Cos(ang)
				sy := cy + rCenter*math.Sin(ang)

				spheres = append(spheres, diskSphere{
					cx: sx,
					cy: sy,
					r:  rF,
					q:  fr.q + tIdx,
					p:  fr.p + sec,
				})

				// Draw golden Farey chord from base boundary to circle center for principal fractions (q <= 5)
				if fr.q <= 5 {
					bx := cx + tier.rBase*math.Cos(ang)
					by := cy + tier.rBase*math.Sin(ang)
					geodesics.MoveTo(bx, by)
					geodesics.LineTo(sx, sy)
				}
			}
		}
	}

	// Central Apollonian Sun Core Sphere
	spheres = append(spheres, diskSphere{
		cx: cx,
		cy: cy,
		r:  rDisk * 0.145,
		q:  1,
		p:  0,
	})

	// Sort spheres from largest to smallest so smaller tangent spheres nest crisply
	sort.Slice(spheres, func(i, j int) bool {
		return spheres[i].r > spheres[j].r
	})

	// 2. Render Golden Farey Geodesic Chords & Base Modular Rings
	rc.StrokePath(geodesics, 0.85*scale, canvas.RoundCap, canvas.RoundJoin, goldFiligree, 0.38)
	rc.DrawSoftRing(cx, cy, rDisk*0.96, 2.2*scale, goldFiligree, 0.92)
	rc.DrawSoftRing(cx, cy, rDisk*0.43, 1.8*scale, goldFiligree, 0.86)

	// 3. Render 3D Shaded Enamel/Jewel Ford Spheres with Specular Highlights & Gold Rings
	for _, sp := range spheres {
		pal := spherePals[(sp.q+sp.p)%len(spherePals)]
		rc.DrawShadedOrb(render.Pt(sp.cx, sp.cy), sp.r*0.97, pal[0], pal[1], ivorySpec, lightDir, 0.98)
		if sp.r >= 5.5*scale {
			rc.DrawSoftRing(sp.cx, sp.cy, sp.r*0.97, math.Min(1.4*scale, sp.r*0.09), goldFiligree, 0.88)
		}
	}

	// 4. Outer Beveled Astrolabe Modular Bezel Ring
	bezelSDF := render.SDFOnion(render.SDFCircle(render.Pt(cx, cy), rDisk*0.985), 8.5*scale)
	rc.RenderSDFDropShadow(bezelSDF, render.Pt(4*scale, 5*scale), 12*scale, 2*scale, abyssBg, 0.85)
	rc.RenderSDFBevel(bezelSDF, goldFiligree, color.RGBA{92, 54, 18, 255}, ivorySpec, 3.8*scale, lightDir, 1.0)

	rc.ApplyBloom(0.82, 6, 0.22)
	c.MergeRenderCanvas(rc)
}
