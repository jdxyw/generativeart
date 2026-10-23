package arts

import (
	"image/color"
	"math"
	"math/rand"
	"sort"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type halvorsenAttractor struct {
	pointsN int
	alpha   float64
}

// NewHalvorsenAttractor returns a halvorsenAttractor engine that integrates the 3D
// cyclic-symmetric Halvorsen chaotic strange attractor along its (1,1,1) trisector
// axis with thin-lens depth-of-field bokeh and 3-channel spectral velocity density mapping.
func NewHalvorsenAttractor(pointsN int, alpha float64) *halvorsenAttractor {
	if pointsN <= 0 {
		pointsN = 1200000
	}
	if alpha <= 0 {
		alpha = 1.4
	}
	return &halvorsenAttractor{
		pointsN: pointsN,
		alpha:   alpha,
	}
}

// Generative renders the 3D Halvorsen strange attractor nebula.
func (ha *halvorsenAttractor) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := c.Width(), c.Height()
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0
	cx, cy := 0.5*fw, 0.56*fh

	abyssBg := color.RGBA{R: 4, G: 6, B: 14, A: 255}
	rc.FillBackground(abyssBg)

	a := ha.alpha
	if a < 1.2 || a > 1.8 {
		a = 1.4
	}

	// Halvorsen vector field f(x, y, z)
	deriv := func(x, y, z float64) (dx, dy, dz float64) {
		dx = -a*x - 4.0*y - 4.0*z - y*y
		dy = -a*y - 4.0*z - 4.0*x - z*z
		dz = -a*z - 4.0*x - 4.0*y - x*x
		return
	}

	// Orthonormal basis aligned with the (1, 1, 1) 3-fold cyclic symmetry axis:
	// e1 = (1, -1, 0)/sqrt(2), e2 = (1, 1, -2)/sqrt(6), e3 = (1, 1, 1)/sqrt(3)
	invSqrt2 := 1.0 / math.Sqrt2
	invSqrt6 := 1.0 / math.Sqrt(6.0)
	invSqrt3 := 1.0 / math.Sqrt(3.0)

	// Spectral 3-channel HDR density buffer (rBuf = high speed/core, gBuf = mid speed/twist, bBuf = outer veil)
	rBuf := make([]float64, w*h)
	gBuf := make([]float64, w*h)
	bBuf := make([]float64, w*h)

	splatBilinear := func(px, py, wR, wG, wB float64) {
		if px < 1.0 || px >= fw-2.0 || py < 1.0 || py >= fh-2.0 {
			return
		}
		x0 := int(px)
		y0 := int(py)
		fx := px - float64(x0)
		fy := py - float64(y0)

		w00 := (1.0 - fx) * (1.0 - fy)
		w10 := fx * (1.0 - fy)
		w01 := (1.0 - fx) * fy
		w11 := fx * fy

		idx00 := y0*w + x0
		idx10 := idx00 + 1
		idx01 := idx00 + w
		idx11 := idx01 + 1

		rBuf[idx00] += wR * w00
		gBuf[idx00] += wG * w00
		bBuf[idx00] += wB * w00

		rBuf[idx10] += wR * w10
		gBuf[idx10] += wG * w10
		bBuf[idx10] += wB * w10

		rBuf[idx01] += wR * w01
		gBuf[idx01] += wG * w01
		bBuf[idx01] += wB * w01

		rBuf[idx11] += wR * w11
		gBuf[idx11] += wG * w11
		bBuf[idx11] += wB * w11
	}

	rng := rand.New(rand.NewSource(c.Seed() + 1409))
	x, y, z := -1.48, -1.51, 2.04
	dt := 0.0038

	// Burn-in 2,000 steps onto the strange attractor manifold
	for i := 0; i < 2000; i++ {
		k1x, k1y, k1z := deriv(x, y, z)
		k2x, k2y, k2z := deriv(x+0.5*dt*k1x, y+0.5*dt*k1y, z+0.5*dt*k1z)
		k3x, k3y, k3z := deriv(x+0.5*dt*k2x, y+0.5*dt*k2y, z+0.5*dt*k2z)
		k4x, k4y, k4z := deriv(x+dt*k3x, y+dt*k3y, z+dt*k3z)
		x += (dt / 6.0) * (k1x + 2.0*k2x + 2.0*k3x + k4x)
		y += (dt / 6.0) * (k1y + 2.0*k2y + 2.0*k3y + k4y)
		z += (dt / 6.0) * (k1z + 2.0*k2z + 2.0*k3z + k4z)
	}

	steps := ha.pointsN
	if steps < 200000 {
		steps = 1200000
	}
	projScale := 25.2 * scale
	focalDepth := -5.2

	// Integrate RK4 and splat all 3 cyclic symmetry permutations (x,y,z), (y,z,x), (z,x,y)
	for i := 0; i < steps; i++ {
		k1x, k1y, k1z := deriv(x, y, z)
		k2x, k2y, k2z := deriv(x+0.5*dt*k1x, y+0.5*dt*k1y, z+0.5*dt*k1z)
		k3x, k3y, k3z := deriv(x+0.5*dt*k2x, y+0.5*dt*k2y, z+0.5*dt*k2z)
		k4x, k4y, k4z := deriv(x+dt*k3x, y+dt*k3y, z+dt*k3z)
		x += (dt / 6.0) * (k1x + 2.0*k2x + 2.0*k3x + k4x)
		y += (dt / 6.0) * (k1y + 2.0*k2y + 2.0*k3y + k4y)
		z += (dt / 6.0) * (k1z + 2.0*k2z + 2.0*k3z + k4z)

		speed := math.Sqrt(k1x*k1x + k1y*k1y + k1z*k1z)
		vNorm := common.Constrain((speed-4.0)/38.0, 0, 1)

		perms := [3][3]float64{
			{x, y, z},
			{y, z, x},
			{z, x, y},
		}
		for pIdx, p := range perms {
			px3, py3, pz3 := p[0], p[1], p[2]
			u := (px3 - py3) * invSqrt2
			v := (px3 + py3 - 2.0*pz3) * invSqrt6
			depth := (px3 + py3 + pz3) * invSqrt3

			// Distinct spectral channel emphasis per lobe + velocity
			ang := math.Atan2(v, u) + float64(pIdx)*0.0
			wR := (0.5 + 0.5*math.Cos(ang)) * (0.4 + 0.6*vNorm)
			wG := (0.5 + 0.5*math.Cos(ang-2.094)) * 0.9
			wB := (0.5 + 0.5*math.Cos(ang+2.094)) * (1.1 - 0.5*vNorm)

			// Perspective projection + thin-lens Depth-of-Field circle of confusion
			persp := 18.0 / math.Max(6.0, 18.0-depth*0.45)
			sx := cx + u*projScale*persp
			sy := cy + v*projScale*persp

			coc := math.Abs(depth-focalDepth) * 0.16 * scale
			if coc > 0.12 {
				aJ := rng.Float64() * 2.0 * math.Pi
				rad := math.Sqrt(rng.Float64()) * coc
				sx += rad * math.Cos(aJ)
				sy += rad * math.Sin(aJ)
			}
			splatBilinear(sx, sy, wR, wG, wB)
		}
	}

	// Compute 99.6th percentile density for automatic exposure normalization
	nonZero := make([]float64, 0, w*h/4)
	for i := 0; i < w*h; i += 2 {
		tot := rBuf[i] + gBuf[i] + bBuf[i]
		if tot > 0 {
			nonZero = append(nonZero, tot)
		}
	}
	sort.Float64s(nonZero)
	p996 := 100.0
	if len(nonZero) > 10 {
		p996 = nonZero[int(float64(len(nonZero))*0.996)]
	}
	logNorm := 1.0 / math.Log1p(p996*0.42)

	// Map 3-channel log density to luminous Oklab cosmic palette (Sapphire -> Magenta/Violet -> Solar Gold -> Ivory)
	deepCobalt := color.RGBA{R: 18, G: 44, B: 128, A: 255}
	electricCyan := color.RGBA{R: 42, G: 206, B: 238, A: 255}
	nebulaMagenta := color.RGBA{R: 212, G: 42, B: 128, A: 255}
	solarGold := color.RGBA{R: 252, G: 188, B: 54, A: 255}
	coreIvory := color.RGBA{R: 255, G: 252, B: 238, A: 255}

	for py := 0; py < h; py++ {
		rowOff := py * w
		for px := 0; px < w; px++ {
			idx := rowOff + px
			rD := math.Log1p(rBuf[idx]*0.45) * logNorm
			gD := math.Log1p(gBuf[idx]*0.45) * logNorm
			bD := math.Log1p(bBuf[idx]*0.45) * logNorm

			lum := common.Constrain(0.36*rD+0.38*gD+0.26*bD, 0, 1.25)
			if lum < 0.005 {
				// Subtle radial cosmic background glow
				dist := math.Hypot(float64(px)-cx, float64(py)-cy) / (0.55 * fw)
				bgCol := render.LerpOklab(abyssBg, deepCobalt, common.Constrain(1.0-dist, 0, 1)*0.16)
				rc.BlendPixel(px, py, bgCol, 1.0)
				continue
			}

			// Spectral ratio between fast core (rD) and slow outer veil (bD)
			spectralRatio := common.Constrain(rD/math.Max(1e-5, rD+bD), 0, 1)
			hueA := render.LerpOklab(electricCyan, nebulaMagenta, common.Constrain(gD/math.Max(1e-5, bD+gD), 0, 1))
			hueB := render.LerpOklab(nebulaMagenta, solarGold, spectralRatio)
			tint := render.LerpOklab(hueA, hueB, spectralRatio)

			var col color.RGBA
			tLum := common.Constrain(lum, 0, 1)
			if tLum < 0.35 {
				col = render.LerpOklab(abyssBg, deepCobalt, tLum/0.35)
				col = render.LerpOklab(col, tint, tLum*1.8)
			} else if tLum < 0.78 {
				col = render.LerpOklab(tint, solarGold, (tLum-0.35)/0.43*0.55)
			} else {
				col = render.LerpOklab(solarGold, coreIvory, common.Constrain((tLum-0.78)/0.32, 0, 1))
			}
			rc.BlendPixel(px, py, col, 1.0)
		}
	}

	rc.ApplyBloom(0.78, 7, 0.34)
	c.MergeRenderCanvas(rc)
}
