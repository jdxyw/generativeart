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

type thomasAttractor struct {
	pointsN  int
	dampingB float64
}

// NewThomasAttractor returns a thomasAttractor engine that integrates René Thomas's
// 3D cyclically symmetric labyrinth chaotic strange attractor (dx/dt = sin(y) - b*x)
// with 6-fold octahedral symmetry permutations, thin-lens DoF bokeh, and spectral density mapping.
func NewThomasAttractor(pointsN int, dampingB float64) *thomasAttractor {
	if pointsN <= 0 {
		pointsN = 900000
	}
	if dampingB <= 0 {
		dampingB = 0.19
	}
	return &thomasAttractor{
		pointsN:  pointsN,
		dampingB: dampingB,
	}
}

// Generative renders the 3D Thomas cyclically symmetric labyrinth strange attractor.
func (ta *thomasAttractor) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := c.Width(), c.Height()
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0
	cx, cy := 0.5*fw, 0.5*fh

	abyssBg := color.RGBA{R: 4, G: 7, B: 16, A: 255}
	rc.FillBackground(abyssBg)

	b := ta.dampingB
	if b <= 0 || math.Abs(b-0.19) < 0.005 {
		b = 0.162 // Chaotic labyrinth strange-attractor regime (avoiding the b=0.19 periodic limit cycle)
	}

	deriv := func(x, y, z float64) (dx, dy, dz float64) {
		dx = math.Sin(y) - b*x
		dy = math.Sin(z) - b*y
		dz = math.Sin(x) - b*z
		return
	}

	// Orthonormal basis aligned with the (1, 1, 1) 3-fold symmetry axis
	invSqrt2 := 1.0 / math.Sqrt2
	invSqrt6 := 1.0 / math.Sqrt(6.0)
	invSqrt3 := 1.0 / math.Sqrt(3.0)

	rBuf := make([]float64, w*h)
	gBuf := make([]float64, w*h)
	bBuf := make([]float64, w*h)

	splatBilinear := func(px, py, wR, wG, wB float64) {
		if px < 1.0 || px >= fw-2.0 || py < 1.0 || py >= fh-2.0 {
			return
		}
		x0, y0 := int(px), int(py)
		fx, fy := px-float64(x0), py-float64(y0)
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

	rng := rand.New(rand.NewSource(c.Seed() + 1999))
	steps := ta.pointsN
	if steps < 300000 {
		steps = 1100000
	}
	dt := 0.038
	projScale := 56.0 * scale

	// Integrate a continuous 72-sheet bifurcation manifold family across b in [0.152, 0.196]
	// so the 3D Thomas labyrinth forms a luminous, volumetric 6-fold silk caustic surface
	const numSheets = 72
	stepsPerSheet := steps / numSheets
	for sRun := 0; sRun < numSheets; sRun++ {
		uSheet := float64(sRun) / float64(numSheets-1)
		bLocal := (b - 0.022) + uSheet*0.044
		if bLocal < 0.145 {
			bLocal = 0.145 + uSheet*0.045
		}

		ang0 := uSheet * 2.0 * math.Pi
		x := 1.45*math.Cos(ang0) + 0.25*math.Sin(3.0*ang0)
		y := 1.45*math.Sin(ang0) - 0.25*math.Cos(3.0*ang0)
		z := 1.15 * math.Cos(2.0*ang0+0.4)

		derivLocal := func(px, py, pz float64) (dx, dy, dz float64) {
			return math.Sin(py) - bLocal*px, math.Sin(pz) - bLocal*py, math.Sin(px) - bLocal*pz
		}

		// Short burn-in onto the invariant manifold sheet
		for i := 0; i < 180; i++ {
			k1x, k1y, k1z := derivLocal(x, y, z)
			k2x, k2y, k2z := derivLocal(x+0.5*dt*k1x, y+0.5*dt*k1y, z+0.5*dt*k1z)
			k3x, k3y, k3z := derivLocal(x+0.5*dt*k2x, y+0.5*dt*k2y, z+0.5*dt*k2z)
			k4x, k4y, k4z := derivLocal(x+dt*k3x, y+dt*k3y, z+dt*k3z)
			x += (dt / 6.0) * (k1x + 2.0*k2x + 2.0*k3x + k4x)
			y += (dt / 6.0) * (k1y + 2.0*k2y + 2.0*k3y + k4y)
			z += (dt / 6.0) * (k1z + 2.0*k2z + 2.0*k3z + k4z)
		}

		for i := 0; i < stepsPerSheet; i++ {
			k1x, k1y, k1z := derivLocal(x, y, z)
			k2x, k2y, k2z := derivLocal(x+0.5*dt*k1x, y+0.5*dt*k1y, z+0.5*dt*k1z)
			k3x, k3y, k3z := derivLocal(x+0.5*dt*k2x, y+0.5*dt*k2y, z+0.5*dt*k2z)
			k4x, k4y, k4z := derivLocal(x+dt*k3x, y+dt*k3y, z+dt*k3z)
			x += (dt / 6.0) * (k1x + 2.0*k2x + 2.0*k3x + k4x)
			y += (dt / 6.0) * (k1y + 2.0*k2y + 2.0*k3y + k4y)
			z += (dt / 6.0) * (k1z + 2.0*k2z + 2.0*k3z + k4z)

			speed := math.Sqrt(k1x*k1x + k1y*k1y + k1z*k1z)
			vNorm := common.Constrain(speed/1.45, 0, 1)

			symPts := [6][3]float64{
				{x, y, z},
				{y, z, x},
				{z, x, y},
				{-x, -y, -z},
				{-y, -z, -x},
				{-z, -x, -y},
			}

			for sIdx, p := range symPts {
				px3, py3, pz3 := p[0], p[1], p[2]
				u := (px3 - py3) * invSqrt2
				v := (px3 + py3 - 2.0*pz3) * invSqrt6
				depth := (px3 + py3 + pz3) * invSqrt3

				persp := 12.0 / math.Max(4.5, 12.0-depth*0.65)
				sx := cx + u*projScale*persp
				sy := cy + v*projScale*persp

				coc := (0.35 + math.Abs(depth)*0.22) * scale
				aJ := rng.Float64() * 2.0 * math.Pi
				rad := math.Sqrt(rng.Float64()) * coc
				sx += rad * math.Cos(aJ)
				sy += rad * math.Sin(aJ)

				angP := math.Atan2(v, u) + float64(sIdx)*0.25 + uSheet*1.8
				wR := (0.5 + 0.5*math.Sin(angP*2.0)) * (0.35 + 0.65*vNorm)
				wG := (0.5 + 0.5*math.Cos(uSheet*math.Pi*2.0+depth*0.8)) * 0.90
				wB := (0.5 - 0.5*math.Sin(angP*2.0)) * (1.0 - 0.35*vNorm)
				splatBilinear(sx, sy, wR, wG, wB)
			}
		}
	}
	_ = deriv

	nonZero := make([]float64, 0, w*h/4)
	for i := 0; i < w*h; i += 2 {
		tot := rBuf[i] + gBuf[i] + bBuf[i]
		if tot > 0 {
			nonZero = append(nonZero, tot)
		}
	}
	sort.Float64s(nonZero)
	p96 := 50.0
	if len(nonZero) > 10 {
		p96 = nonZero[int(float64(len(nonZero))*0.96)]
	}
	logNorm := 1.0 / math.Log1p(p96*0.32)

	deepSapphire := color.RGBA{R: 16, G: 42, B: 118, A: 255}
	turquoiseGlow := color.RGBA{R: 38, G: 204, B: 214, A: 255}
	coralFlame := color.RGBA{R: 232, G: 68, B: 54, A: 255}
	imperialGold := color.RGBA{R: 252, G: 196, B: 58, A: 255}
	ivoryCore := color.RGBA{R: 255, G: 252, B: 236, A: 255}

	for py := 0; py < h; py++ {
		rowOff := py * w
		for px := 0; px < w; px++ {
			idx := rowOff + px
			rD := math.Log1p(rBuf[idx]*0.45) * logNorm
			gD := math.Log1p(gBuf[idx]*0.45) * logNorm
			bD := math.Log1p(bBuf[idx]*0.45) * logNorm

			lum := common.Constrain(0.36*rD+0.36*gD+0.28*bD, 0, 1.25)
			if lum < 0.005 {
				dist := math.Hypot(float64(px)-cx, float64(py)-cy) / (0.55 * fw)
				bgCol := render.LerpOklab(abyssBg, deepSapphire, common.Constrain(1.0-dist, 0, 1)*0.16)
				rc.BlendPixel(px, py, bgCol, 1.0)
				continue
			}

			ratio := common.Constrain(rD/math.Max(1e-5, rD+bD), 0, 1)
			tint := render.LerpOklab(turquoiseGlow, coralFlame, ratio)
			tint = render.LerpOklab(tint, imperialGold, common.Constrain(gD*0.55, 0, 1))

			var col color.RGBA
			tLum := common.Constrain(lum, 0, 1)
			if tLum < 0.36 {
				col = render.LerpOklab(abyssBg, deepSapphire, tLum/0.36)
				col = render.LerpOklab(col, tint, tLum*1.85)
			} else if tLum < 0.78 {
				col = render.LerpOklab(tint, imperialGold, (tLum-0.36)/0.42*0.60)
			} else {
				col = render.LerpOklab(imperialGold, ivoryCore, common.Constrain((tLum-0.78)/0.30, 0, 1))
			}
			rc.BlendPixel(px, py, col, 1.0)
		}
	}

	rc.ApplyBloom(0.78, 7, 0.32)
	c.MergeRenderCanvas(rc)
}
