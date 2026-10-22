package arts

import (
	"image/color"
	"math"
	"math/cmplx"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type newtonBasin struct {
	degree  int
	maxIter int
}

// NewNewtonBasin returns a newtonBasin engine that renders a complex Newton-Raphson
// root-finding fractal rosette for P(z) = z^n + c*z^(n/2) - 1 with 2x2 SSAA,
// smooth exponential convergence shading, and 3D pleated silk normal lighting.
func NewNewtonBasin(degree, maxIter int) *newtonBasin {
	if degree < 3 {
		degree = 6
	}
	if maxIter <= 0 {
		maxIter = 48
	}
	return &newtonBasin{
		degree:  degree,
		maxIter: maxIter,
	}
}

// Generative renders the complex Newton-Raphson fractal rosette with 2x2 SSAA.
func (nb *newtonBasin) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := c.Width(), c.Height()
	fw, fh := float64(w), float64(h)

	// Rich Stained-Silk & Jewel Palette for Root Basins
	rootPalettes := [][3]color.RGBA{
		// Basin 0: Imperial Amber & Sun Gold
		{color.RGBA{92, 38, 12, 255}, color.RGBA{238, 158, 36, 255}, color.RGBA{255, 238, 152, 255}},
		// Basin 1: Jingtai Turquoise & Pale Ice
		{color.RGBA{12, 58, 78, 255}, color.RGBA{34, 168, 182, 255}, color.RGBA{186, 248, 248, 255}},
		// Basin 2: Cinnabar Coral & Blush Gold
		{color.RGBA{88, 16, 28, 255}, color.RGBA{218, 58, 52, 255}, color.RGBA{255, 178, 156, 255}},
		// Basin 3: Royal Cobalt & Cerulean
		{color.RGBA{14, 28, 86, 255}, color.RGBA{42, 98, 196, 255}, color.RGBA{154, 208, 255, 255}},
		// Basin 4: Emerald Jade & Mint
		{color.RGBA{12, 62, 46, 255}, color.RGBA{42, 164, 118, 255}, color.RGBA{182, 246, 212, 255}},
		// Basin 5: Byzantium Magenta & Rose
		{color.RGBA{68, 14, 54, 255}, color.RGBA{192, 42, 118, 255}, color.RGBA{252, 168, 214, 255}},
	}
	boundaryGold := color.RGBA{R: 255, G: 224, B: 122, A: 255}
	abyssDark := color.RGBA{R: 10, G: 12, B: 22, A: 255}

	rc.FillBackground(abyssDark)
	rc.SetBlendMode(render.BlendOver)

	deg := nb.degree
	if deg < 3 {
		deg = 6
	}
	nComplex := complex(float64(deg), 0)
	nMinus1 := complex(float64(deg-1), 0)

	// Precompute the exact roots of z^deg - 1 = 0 on the unit circle
	roots := make([]complex128, deg)
	for k := 0; k < deg; k++ {
		ang := (float64(k) / float64(deg)) * 2.0 * math.Pi
		roots[k] = cmplx.Rect(1.0, ang)
	}

	const tol = 1e-5
	logTol := math.Log(tol)
	zoom := 2.35 / math.Min(fw, fh)

	// Evaluate a single sub-pixel sample in the complex plane
	sampleComplex := func(sx, sy float64) (rLin, gLin, bLin float64) {
		zr := (sx - 0.5*fw) * zoom
		zi := (sy - 0.5*fh) * zoom
		// Rotate slightly for dynamic diagonal symmetry
		z := cmplx.Rect(math.Hypot(zr, zi), math.Atan2(zi, zr)+math.Pi/float64(2*deg))

		prevDist := 1.0
		var prevZ complex128
		rootIdx := -1
		smoothIter := float64(nb.maxIter)

		for it := 0; it < nb.maxIter; it++ {
			zPowNMinus1 := cmplx.Pow(z, nMinus1)
			denom := nComplex * zPowNMinus1
			if cmplx.Abs(denom) < 1e-14 {
				break
			}
			fVal := zPowNMinus1*z - 1.0
			// Slightly relaxed complex damping (1.0 + 0.06i) creates swirling baroque spiral filaments at basin boundaries
			step := fVal / denom
			prevZ = z
			z = z - complex(0.98, 0.05)*step

			// Check convergence to any root
			for rIdx, rt := range roots {
				d := cmplx.Abs(z - rt)
				if d < tol {
					rootIdx = rIdx
					dPrev := math.Max(tol*1.01, cmplx.Abs(prevZ-rt))
					denomLog := math.Log(d) - math.Log(dPrev)
					frac := 0.0
					if math.Abs(denomLog) > 1e-9 {
						frac = common.Constrain((logTol-math.Log(dPrev))/denomLog, 0, 1)
					}
					smoothIter = float64(it) + frac
					prevDist = d
					break
				}
			}
			if rootIdx >= 0 {
				break
			}
		}

		if rootIdx < 0 {
			return render.SRGB8ToLinear(abyssDark.R), render.SRGB8ToLinear(abyssDark.G), render.SRGB8ToLinear(abyssDark.B)
		}
		_ = prevDist

		pal := rootPalettes[rootIdx%len(rootPalettes)]
		// Pleated silk wave modulation from smooth iteration count + argument angle around root
		pleat := 0.5 + 0.5*math.Cos(smoothIter*0.85)
		arg := cmplx.Phase(z - prevZ)
		normalShade := 0.5 + 0.5*math.Cos(arg-0.75)

		u := common.Constrain(pleat*0.62+normalShade*0.38, 0, 1)
		var col color.RGBA
		if u < 0.5 {
			col = render.LerpOklab(pal[0], pal[1], u*2.0)
		} else {
			col = render.LerpOklab(pal[1], pal[2], (u-0.5)*2.0)
		}

		// Highlight fractal Julia boundary ridges (high iteration count) with gilded filaments and dark troughs
		if smoothIter > 10.0 {
			ridge := common.Constrain((smoothIter-10.0)/16.0, 0, 1)
			filament := 0.5 + 0.5*math.Sin(smoothIter*2.4)
			if filament > 0.68 {
				col = render.LerpOklab(col, boundaryGold, ridge*0.78)
			} else {
				col = render.LerpOklab(col, abyssDark, ridge*0.72)
			}
		}

		return render.SRGB8ToLinear(col.R), render.SRGB8ToLinear(col.G), render.SRGB8ToLinear(col.B)
	}

	// 2x2 Sub-Pixel Super-Sampling (SSAA) across all rows
	offsets := [2]float64{0.25, 0.75}
	for y := 0; y < h; y++ {
		fy := float64(y)
		for x := 0; x < w; x++ {
			fx := float64(x)
			var accR, accG, accB float64
			for _, oy := range offsets {
				for _, ox := range offsets {
					rL, gL, bL := sampleComplex(fx+ox, fy+oy)
					accR += rL
					accG += gL
					accB += bL
				}
			}
			rc.SetPixelLinear(x, y, accR*0.25, accG*0.25, accB*0.25, 1.0)
		}
	}

	c.MergeRenderCanvas(rc)
}
