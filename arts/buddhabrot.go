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

type buddhabrot struct {
	samplesN int
	maxIter  int
}

// NewBuddhabrot returns a buddhabrot engine that renders a 3-channel spectral
// Nebulabrot (Buddhabrot) probability density cloud using boundary importance
// sampling, bilateral symmetry, and Oklab false-color HDR tone mapping.
func NewBuddhabrot(samplesN, maxIter int) *buddhabrot {
	if samplesN <= 0 {
		samplesN = 450000
	}
	if maxIter <= 0 {
		maxIter = 360
	}
	return &buddhabrot{
		samplesN: samplesN,
		maxIter:  maxIter,
	}
}

// Generative renders the 3-channel spectral Nebulabrot / Buddhabrot nebula.
func (bb *buddhabrot) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := c.Width(), c.Height()
	fw, fh := float64(w), float64(h)

	abyssSpace := color.RGBA{R: 6, G: 8, B: 18, A: 255}
	deepIndigo := color.RGBA{R: 18, G: 28, B: 74, A: 255}
	cobaltBlue := color.RGBA{R: 34, G: 96, B: 184, A: 255}
	electricCyan := color.RGBA{R: 58, G: 198, B: 214, A: 255}
	vibrantMagenta := color.RGBA{R: 204, G: 44, B: 124, A: 255}
	goldenAmber := color.RGBA{R: 248, G: 182, B: 56, A: 255}
	warmIvory := color.RGBA{R: 255, G: 244, B: 212, A: 255}

	rc.FillBackground(abyssSpace)

	nPix := w * h
	histR := make([]float64, nPix)
	histG := make([]float64, nPix)
	histB := make([]float64, nPix)

	// Complex plane viewport rotated so real axis [-1.95, 0.85] maps vertically top-to-bottom
	// and imaginary axis [-1.40, 1.40] maps horizontally left-to-right
	const (
		rMin = -1.92
		rMax = 0.82
		iMin = -1.37
		iMax = 1.37
	)
	invRSpan := fh / (rMax - rMin)
	invISpan := fw / (iMax - iMin)

	splatHist := func(buf []float64, fx, fy float64) {
		x0 := int(math.Floor(fx))
		y0 := int(math.Floor(fy))
		if x0 < 0 || x0+1 >= w || y0 < 0 || y0+1 >= h {
			return
		}
		tx := fx - float64(x0)
		ty := fy - float64(y0)
		idx00 := y0*w + x0
		buf[idx00] += (1.0 - tx) * (1.0 - ty)
		buf[idx00+1] += tx * (1.0 - ty)
		buf[idx00+w] += (1.0 - tx) * ty
		buf[idx00+w+1] += tx * ty
	}

	maxIt := bb.maxIter
	if maxIt < 120 {
		maxIt = 240
	}
	orbitR := make([]float64, maxIt)
	orbitI := make([]float64, maxIt)

	rng := rand.New(rand.NewSource(202621))
	samples := bb.samplesN
	if samples < 50000 {
		samples = 180000
	}

	// Importance-sampled boundary generator focused strictly on the Mandelbrot boundary
	// (main cardioid rim, period-2 bulb rim, period-3 bulbs, and antenna mini-brots)
	sampleCandidateC := func(idx int) (cr, ci float64) {
		mode := idx % 8
		switch {
		case mode < 5:
			// Tightly hug the main cardioid boundary: r in [1.0002, 1.035]
			t := rng.Float64() * 2.0 * math.Pi
			r := 1.0002 + math.Pow(rng.Float64(), 2.4)*0.045
			cr = 0.5*r*math.Cos(t) - 0.25*r*r*math.Cos(2.0*t)
			ci = 0.5*r*math.Sin(t) - 0.25*r*r*math.Sin(2.0*t)
		case mode < 7:
			// Tightly hug the period-2 bulb boundary centered at (-1, 0) with radius 0.25
			t := rng.Float64() * 2.0 * math.Pi
			r := 0.2501 + math.Pow(rng.Float64(), 2.4)*0.032
			cr = -1.0 + r*math.Cos(t)
			ci = r * math.Sin(t)
		default:
			// Top/Bottom period-3 bulbs around (-0.125, +/- 0.744) and needle around (-1.75, 0)
			if rng.Float64() < 0.65 {
				t := rng.Float64() * 2.0 * math.Pi
				r := 0.092 + math.Pow(rng.Float64(), 2.0)*0.025
				cr = -0.122 + r*math.Cos(t)
				ci = 0.745 + r*math.Sin(t)
			} else {
				cr = -1.78 + (rng.Float64()-0.5)*0.12
				ci = (rng.Float64() - 0.5) * 0.04
			}
		}
		return cr, ci
	}

	for s := 0; s < samples; s++ {
		cr, ci := sampleCandidateC(s)

		// Fast analytical rejection of points strictly inside the main cardioid or period-2 bulb
		q := (cr-0.25)*(cr-0.25) + ci*ci
		if q*(q+(cr-0.25)) <= 0.25*ci*ci*0.998 {
			continue
		}
		if (cr+1.0)*(cr+1.0)+ci*ci <= 0.0625*0.998 {
			continue
		}

		zr, zi := 0.0, 0.0
		escIter := -1
		for it := 0; it < maxIt; it++ {
			zr2 := zr * zr
			zi2 := zi * zi
			if zr2+zi2 > 16.0 {
				escIter = it
				break
			}
			zi = 2.0*zr*zi + ci
			zr = zr2 - zi2 + cr
			orbitR[it] = zr
			orbitI[it] = zi
		}

		// Only keep high-structure escaping orbits (skip short noisy background trajectories)
		if escIter < 24 {
			continue
		}

		// Start from k = 3 so the initial seed ring is not directly imprinted
		for k := 3; k < escIter; k++ {
			py := (orbitR[k] - rMin) * invRSpan
			px1 := (orbitI[k] - iMin) * invISpan
			px2 := (-orbitI[k] - iMin) * invISpan

			if escIter >= 110 {
				splatHist(histR, px1, py)
				splatHist(histR, px2, py)
			}
			if escIter >= 45 && escIter < 160 {
				splatHist(histG, px1, py)
				splatHist(histG, px2, py)
			}
			if escIter >= 24 && escIter < 75 {
				splatHist(histB, px1, py)
				splatHist(histB, px2, py)
			}
		}
	}

	// Smooth 3x3 box/tent filter on histograms to eliminate single-pixel Monte Carlo speckle
	smoothChannel := func(src []float64) []float64 {
		dst := make([]float64, len(src))
		for y := 1; y < h-1; y++ {
			row := y * w
			for x := 1; x < w-1; x++ {
				i := row + x
				dst[i] = 0.25*src[i] +
					0.125*(src[i-1]+src[i+1]+src[i-w]+src[i+w]) +
					0.0625*(src[i-w-1]+src[i-w+1]+src[i+w-1]+src[i+w+1])
			}
		}
		return dst
	}
	histR = smoothChannel(histR)
	histG = smoothChannel(histG)
	histB = smoothChannel(histB)

	findQuantile := func(buf []float64, q float64) float64 {
		nonZero := make([]float64, 0, len(buf)/4)
		for _, v := range buf {
			if v > 1e-4 {
				nonZero = append(nonZero, v)
			}
		}
		if len(nonZero) == 0 {
			return 1.0
		}
		sort.Float64s(nonZero)
		idx := int(float64(len(nonZero)-1) * q)
		return math.Max(1.0, nonZero[idx])
	}

	normR := math.Log1p(findQuantile(histR, 0.995))
	normG := math.Log1p(findQuantile(histG, 0.995))
	normB := math.Log1p(findQuantile(histB, 0.995))

	grad := render.NewMultiStopGradient(
		abyssSpace,
		deepIndigo,
		cobaltBlue,
		electricCyan,
		vibrantMagenta,
		goldenAmber,
		warmIvory,
	)

	// Map 3 spectral channels into a high-contrast Oklab Nebulabrot without whiteout clipping
	for y := 0; y < h; y++ {
		rowOff := y * w
		for x := 0; x < w; x++ {
			idx := rowOff + x
			uR := common.Constrain(math.Log1p(histR[idx])/normR, 0, 1.0)
			uG := common.Constrain(math.Log1p(histG[idx])/normG, 0, 1.0)
			uB := common.Constrain(math.Log1p(histB[idx])/normB, 0, 1.0)

			lum := common.Constrain(0.46*uR+0.34*uG+0.20*uB, 0, 1.0)
			// S-curve contrast shaping so background stays deep obsidian and filaments pop
			lumShaped := math.Pow(lum, 1.25)
			baseCol := grad.SampleOklab(lumShaped)

			// Tint by spectral channel dominance (R -> gold/magenta, G -> cyan, B -> cobalt)
			if uR > uG {
				baseCol = render.LerpOklab(baseCol, goldenAmber, (uR-uG)*0.45)
			} else if uG > uB {
				baseCol = render.LerpOklab(baseCol, electricCyan, (uG-uB)*0.40)
			}

			rc.SetPixelLinear(
				x, y,
				render.SRGB8ToLinear(baseCol.R),
				render.SRGB8ToLinear(baseCol.G),
				render.SRGB8ToLinear(baseCol.B),
				1.0,
			)
		}
	}

	rc.ApplyBloom(0.75, 4.0*(math.Min(fw, fh)/600.0), 0.28)
	c.MergeRenderCanvas(rc)
}
