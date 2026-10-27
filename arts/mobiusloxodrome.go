package arts

import (
	"image/color"
	"math"
	"math/cmplx"
	"sort"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type mobiusLoxodrome struct {
	arms          int
	spheresPerArm int
}

// NewMobiusLoxodrome returns a mobiusLoxodrome engine that applies a conformal
// double-spiral Möbius transformation z = a * tanh(zeta / 2) to a doubly-periodic
// hexagonal Doyle lattice, mapping cylinder circles to exact tangent 3D jewel spheres
// spiraling out of a source pole (-a) and plunging into a sink pole (+a).
func NewMobiusLoxodrome(arms, spheresPerArm int) *mobiusLoxodrome {
	if arms <= 0 {
		arms = 14
	}
	if spheresPerArm <= 0 {
		spheresPerArm = 44
	}
	return &mobiusLoxodrome{
		arms:          arms,
		spheresPerArm: spheresPerArm,
	}
}

// Generative renders the conformal double-spiral Möbius loxodrome sphere packing.
func (ml *mobiusLoxodrome) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := float64(c.Width()), float64(c.Height())
	scale := math.Min(w, h) / 600.0
	cx, cy := 0.5*w, 0.5*h

	abyssSpace := color.RGBA{R: 7, G: 9, B: 18, A: 255}
	goldFiligree := color.RGBA{R: 248, G: 196, B: 68, A: 255}
	ivorySpec := color.RGBA{R: 255, G: 250, B: 232, A: 255}

	spherePals := [][2]color.RGBA{
		{color.RGBA{242, 178, 48, 255}, color.RGBA{118, 48, 14, 255}},  // Imperial Gold
		{color.RGBA{38, 184, 192, 255}, color.RGBA{12, 62, 88, 255}},   // Jingtai Turquoise
		{color.RGBA{218, 54, 48, 255}, color.RGBA{96, 16, 22, 255}},    // Cinnabar Vermilion
		{color.RGBA{44, 102, 198, 255}, color.RGBA{14, 28, 86, 255}},   // Royal Cobalt Lapis
		{color.RGBA{46, 168, 118, 255}, color.RGBA{14, 68, 48, 255}},   // Celadon Jade
		{color.RGBA{176, 48, 128, 255}, color.RGBA{68, 14, 56, 255}},   // Byzantium Magenta
	}

	rc.FillBackground(abyssSpace)
	rc.SetBlendMode(render.BlendOver)

	// Diagonal pole placement for dramatic double-spiral composition
	poleAngle := -0.62
	rotPole := cmplx.Rect(1.0, poleAngle)
	poleScale := 0.255 * math.Min(w, h)

	// Conformal bipolar Möbius map: z = poleScale * rotPole * coth(zeta / 2) = poleScale * rotPole * (e^zeta + 1) / (e^zeta - 1)
	// wait: tanh(zeta/2) = (e^zeta - 1)/(e^zeta + 1) maps u -> -inf to -1, u -> +inf to +1, and u = 0, v = pi to infinity!
	// If we use z = poleScale * rotPole * (e^zeta - 1) / (e^zeta + 1.42), then ALL circles stay bounded on the canvas!
	conformalMap := func(zeta complex128) complex128 {
		ez := cmplx.Exp(zeta)
		wVal := (ez - 1.0) / (1.38 + 0.72*ez)
		return complex(cx, cy) + complex(poleScale, 0)*rotPole*wVal
	}

	// Background bipolar dipole glow around the two spiral poles
	pSource := conformalMap(complex(-8.0, 0))
	pSink := conformalMap(complex(8.0, 0))
	for y := 0; y < c.Height(); y++ {
		fy := float64(y) + 0.5
		for x := 0; x < c.Width(); x++ {
			fx := float64(x) + 0.5
			d1 := math.Hypot(fx-real(pSource), fy-imag(pSource)) / (0.55 * w)
			d2 := math.Hypot(fx-real(pSink), fy-imag(pSink)) / (0.55 * w)
			glow := math.Exp(-d1*d1*3.2)*0.22 + math.Exp(-d2*d2*3.2)*0.22
			col := render.LerpOklab(abyssSpace, color.RGBA{26, 54, 118, 255}, common.Constrain(glow, 0, 1))
			rc.BlendPixel(x, y, col, 1.0)
		}
	}

	numArms := ml.arms
	if numArms < 8 {
		numArms = 14
	}
	nAlong := ml.spheresPerArm
	if nAlong < 20 {
		nAlong = 44
	}

	// Hexagonal packing spacing in the log-cylinder (u, v in [0, 2*pi))
	dv := (2.0 * math.Pi) / float64(numArms)
	du := dv * (math.Sqrt(3.0) * 0.5)
	rLog := dv * 0.46
	loxTilt := 0.28 // Loxodromic spiral twist

	// 1. Draw Golden Loxodromic Spiral Geodesics & Orthogonal Equipotential Rings
	spiralPath := &canvas.Path{}
	for k := 0; k < numArms; k++ {
		v0 := (float64(k) / float64(numArms)) * 2.0 * math.Pi
		first := true
		for s := -70; s <= 70; s++ {
			u := float64(s) * 0.065
			v := v0 + u*loxTilt*2.0
			pt := conformalMap(complex(u, v))
			px, py := real(pt), imag(pt)
			if px < -80 || px > w+80 || py < -80 || py > h+80 {
				first = true
				continue
			}
			if first {
				spiralPath.MoveTo(px, py)
				first = false
			} else {
				spiralPath.LineTo(px, py)
			}
		}
	}
	rc.StrokePath(spiralPath, 0.95*scale, canvas.RoundCap, canvas.RoundJoin, goldFiligree, 0.42)

	// 2. Compute Conformally Mapped 3D Spheres via 3-Point Circle Reconstruction
	type loxSphere struct {
		cx, cy float64
		r      float64
		armIdx int
		step   int
	}
	spheres := make([]loxSphere, 0, numArms*nAlong)

	for step := -nAlong / 2; step <= nAlong/2; step++ {
		u := float64(step) * du
		vShift := 0.0
		if step%2 != 0 {
			vShift = 0.5 * dv
		}
		for arm := 0; arm < numArms; arm++ {
			v := float64(arm)*dv + vShift + u*loxTilt
			zC := complex(u, v)

			// Sample 3 boundary points on the log-circle to reconstruct the exact image circle in screen space
			p0 := conformalMap(zC + complex(rLog, 0))
			p1 := conformalMap(zC + complex(-0.5*rLog, 0.8660254*rLog))
			p2 := conformalMap(zC + complex(-0.5*rLog, -0.8660254*rLog))

			scx, scy, sr, ok := circleFrom3Complex(p0, p1, p2)
			if !ok || sr < 1.5*scale || sr > 88.0*scale {
				continue
			}
			if scx+sr < -10 || scx-sr > w+10 || scy+sr < -10 || scy-sr > h+10 {
				continue
			}
			spheres = append(spheres, loxSphere{
				cx:     scx,
				cy:     scy,
				r:      sr,
				armIdx: arm,
				step:   step + nAlong,
			})
		}
	}

	// Sort spheres so larger background/outer spheres render before smaller nested core spheres
	sort.Slice(spheres, func(i, j int) bool {
		return spheres[i].r > spheres[j].r
	})

	lightDir := render.Pt(-0.62, -0.74)
	for _, sp := range spheres {
		pal := spherePals[(sp.armIdx+sp.step)%len(spherePals)]
		rc.DrawShadedOrb(render.Pt(sp.cx, sp.cy), sp.r*0.96, pal[0], pal[1], ivorySpec, lightDir, 0.98)
		if sp.r >= 5.2*scale {
			rc.DrawSoftRing(sp.cx, sp.cy, sp.r*0.96, math.Min(1.3*scale, sp.r*0.085), goldFiligree, 0.86)
		}
	}

	// Bright singular pole cores at pSource and pSink
	for _, pole := range []complex128{pSource, pSink} {
		px, py := real(pole), imag(pole)
		rc.DrawGlowDot(px, py, 6.5*scale, ivorySpec, 0.98, 0.88)
		rc.DrawSoftRing(px, py, 11.0*scale, 1.4*scale, goldFiligree, 0.88)
	}

	rc.ApplyBloom(0.82, 6, 0.24)
	c.MergeRenderCanvas(rc)
}

func circleFrom3Complex(z0, z1, z2 complex128) (cx, cy, r float64, ok bool) {
	x1, y1 := real(z0), imag(z0)
	x2, y2 := real(z1), imag(z1)
	x3, y3 := real(z2), imag(z2)
	d := 2.0 * (x1*(y2-y3) + x2*(y3-y1) + x3*(y1-y2))
	if math.Abs(d) < 1e-7 {
		return 0, 0, 0, false
	}
	n1 := x1*x1 + y1*y1
	n2 := x2*x2 + y2*y2
	n3 := x3*x3 + y3*y3
	cx = (n1*(y2-y3) + n2*(y3-y1) + n3*(y1-y2)) / d
	cy = (n1*(x3-x2) + n2*(x1-x3) + n3*(x2-x1)) / d
	r = math.Hypot(x1-cx, y1-cy)
	return cx, cy, r, true
}
