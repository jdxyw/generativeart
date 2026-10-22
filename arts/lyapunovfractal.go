package arts

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type lyapunovFractal struct {
	seq     string
	maxIter int
}

// NewLyapunovFractal returns a lyapunovFractal engine that renders a Markus-Lyapunov
// logistic stability exponent landscape with 3D metallic gold and turquoise bevel
// shading over deep cobalt and crimson chaos regions.
func NewLyapunovFractal(seq string, maxIter int) *lyapunovFractal {
	if len(seq) == 0 {
		seq = "AABAB"
	}
	if maxIter <= 0 {
		maxIter = 84
	}
	return &lyapunovFractal{
		seq:     seq,
		maxIter: maxIter,
	}
}

// Generative renders the Markus-Lyapunov stability exponent landscape.
func (lf *lyapunovFractal) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := c.Width(), c.Height()
	fw, fh := float64(w), float64(h)

	// Gilded Brass, Turquoise Enamel & Deep Cobalt Chaos Palette
	chaosAbyss := color.RGBA{R: 8, G: 12, B: 26, A: 255}
	chaosCobalt := color.RGBA{R: 22, G: 56, B: 128, A: 255}
	chaosCrimson := color.RGBA{R: 158, G: 34, B: 62, A: 255}

	stableBronze := color.RGBA{R: 68, G: 44, B: 18, A: 255}
	stableTurquoise := color.RGBA{R: 34, G: 158, B: 168, A: 255}
	stableGold := color.RGBA{R: 238, G: 178, B: 54, A: 255}
	ivorySpec := color.RGBA{R: 255, G: 244, B: 192, A: 255}

	rc.FillBackground(chaosAbyss)

	seq := lf.seq
	if len(seq) < 2 {
		seq = "AABAB"
	}
	seqLen := len(seq)
	isB := make([]bool, seqLen)
	for i := 0; i < seqLen; i++ {
		isB[i] = (seq[i] == 'B' || seq[i] == 'b' || seq[i] == '1')
	}

	// Classic "Zircon Zity" / Golden Swallowtail viewport [2.52, 3.96] x [2.52, 3.96]
	const (
		aMin = 2.52
		aMax = 3.96
		bMin = 2.52
		bMax = 3.96
	)

	iters := lf.maxIter
	if iters < 36 {
		iters = 72
	}
	warmup := 18

	evalLyapunov := func(a, b float64) float64 {
		x := 0.5
		for i := 0; i < warmup; i++ {
			r := a
			if isB[i%seqLen] {
				r = b
			}
			x = r * x * (1.0 - x)
			if x <= 1e-9 || x >= 1.0-1e-9 {
				x = 0.5
			}
		}

		var sumLog float64
		valid := 0
		for i := 0; i < iters; i++ {
			r := a
			if isB[i%seqLen] {
				r = b
			}
			x = r * x * (1.0 - x)
			deriv := math.Abs(r * (1.0 - 2.0*x))
			if deriv < 1e-12 {
				deriv = 1e-12
			}
			sumLog += math.Log(deriv)
			valid++
		}
		if valid == 0 {
			return 0
		}
		return sumLog / float64(valid)
	}

	lyapGrid := make([]float64, w*h)
	for y := 0; y < h; y++ {
		bVal := bMax - (float64(y)+0.5)/fh*(bMax-bMin)
		rowOff := y * w
		for x := 0; x < w; x++ {
			aVal := aMin + (float64(x)+0.5)/fw*(aMax-aMin)
			lyapGrid[rowOff+x] = evalLyapunov(aVal, bVal)
		}
	}

	stableGrad := render.NewMultiStopGradient(
		ivorySpec,
		stableGold,
		stableTurquoise,
		stableBronze,
		chaosAbyss,
	)
	chaosGrad := render.NewMultiStopGradient(
		chaosAbyss,
		chaosCobalt,
		chaosCrimson,
		stableGold,
	)

	// Shade the Lyapunov landscape with 3D surface normal specular highlights on the swallowtail spears
	for y := 0; y < h; y++ {
		rowOff := y * w
		yPrev := y - 1
		if yPrev < 0 {
			yPrev = 0
		}
		yNext := y + 1
		if yNext >= h {
			yNext = h - 1
		}
		for x := 0; x < w; x++ {
			xPrev := x - 1
			if xPrev < 0 {
				xPrev = 0
			}
			xNext := x + 1
			if xNext >= w {
				xNext = w - 1
			}

			lam := lyapGrid[rowOff+x]
			dlx := (lyapGrid[rowOff+xNext] - lyapGrid[rowOff+xPrev]) * 2.5
			dly := (lyapGrid[yNext*w+x] - lyapGrid[yPrev*w+x]) * 2.5

			var col color.RGBA
			if lam < 0 {
				// Stable periodic attractors (gilded swallowtail spears)
				u := common.Constrain(1.0-math.Exp(lam*1.35), 0, 1)
				col = stableGrad.SampleOklab(u)

				// 3D ridge lighting along stability exponent gradient
				nx, ny, nz := -dlx, -dly, 1.0
				invL := 1.0 / math.Sqrt(nx*nx+ny*ny+nz*nz)
				ndotl := common.Constrain((nx*(-0.6)+ny*(-0.65)+nz*0.48)*invL, 0, 1)
				spec := math.Pow(ndotl, 10.0)
				col = render.LerpOklab(stableBronze, col, 0.35+0.65*ndotl)
				if spec > 0.1 {
					col = render.LerpOklab(col, ivorySpec, common.Constrain(spec*0.75, 0, 1))
				}
			} else {
				// Chaotic sea (lambda >= 0)
				u := common.Constrain(math.Tanh(lam*1.6), 0, 1)
				col = chaosGrad.SampleOklab(u * 0.78)
			}

			rc.SetPixelLinear(
				x, y,
				render.SRGB8ToLinear(col.R),
				render.SRGB8ToLinear(col.G),
				render.SRGB8ToLinear(col.B),
				1.0,
			)
		}
	}

	c.MergeRenderCanvas(rc)
}
