package arts

import (
	"image/color"
	"math"
	"runtime"
	"sync"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type faradayWaves struct {
	symmetry      int
	harmonicScale float64
}

// NewFaradayWaves returns a faradayWaves engine that simulates 12-fold dodecagonal
// quasicrystalline Faraday capillary standing waves with resonant triad coupling,
// analytical 3D normal Blinn-Phong liquid-metal shading, and Marching Squares contours.
func NewFaradayWaves(symmetry int, harmonicScale float64) *faradayWaves {
	if symmetry <= 0 {
		symmetry = 12
	}
	if harmonicScale <= 0 {
		harmonicScale = 1.0
	}
	return &faradayWaves{
		symmetry:      symmetry,
		harmonicScale: harmonicScale,
	}
}

// Generative renders the 12-fold quasicrystalline Faraday standing-wave relief.
func (fwv *faradayWaves) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := c.Width(), c.Height()
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0
	cx, cy := 0.5*fw, 0.5*fh

	// Liquid-Metal, Porcelain Turquoise, Lapis & Imperial Gold Palette
	abyssTrough := color.RGBA{R: 7, G: 11, B: 24, A: 255}
	deepLapis := color.RGBA{R: 18, G: 48, B: 118, A: 255}
	ceruleanWave := color.RGBA{R: 36, G: 124, B: 176, A: 255}
	turquoiseCrest := color.RGBA{R: 54, G: 196, B: 188, A: 255}
	amberRidge := color.RGBA{R: 246, G: 184, B: 54, A: 255}
	coralPeak := color.RGBA{R: 224, G: 66, B: 46, A: 255}
	ivorySpec := color.RGBA{R: 255, G: 252, B: 234, A: 255}

	rc.FillBackground(abyssTrough)
	rc.SetBlendMode(render.BlendOver)

	sym := fwv.symmetry
	if sym < 6 {
		sym = 12
	}
	k1 := (0.048 * fwv.harmonicScale) / scale
	// Second incommensurate wavenumber k2 = k1 * 2*cos(pi/12) = k1 * sqrt(2 + sqrt(3)) for resonant triad locking
	k2 := k1 * math.Sqrt(2.0+math.Sqrt(3.0)) * 0.52

	dirs := make([][2]float64, sym)
	for m := 0; m < sym; m++ {
		ang := (float64(m) / float64(sym)) * 2.0 * math.Pi
		dirs[m] = [2]float64{math.Cos(ang), math.Sin(ang)}
	}

	normFac := 1.0 / float64(sym)

	// Evaluate Faraday quasicrystal wave height W(x, y) and exact analytical gradient (dW/dx, dW/dy)
	evalFaraday := func(x, y float64) (val, gx, gy float64) {
		dx := x - cx
		dy := y - cy
		for m, d := range dirs {
			u1 := (dx*d[0] + dy*d[1]) * k1
			u2 := (dx*d[0] + dy*d[1]) * k2
			ph := float64(m%2) * 0.35

			c1, s1 := math.Cos(u1+ph), math.Sin(u1+ph)
			c2, s2 := math.Cos(u2-ph), math.Sin(u2-ph)
			// Second harmonic capillary sharpening
			cHarm, sHarm := math.Cos(2.0*u1), math.Sin(2.0*u1)

			val += c1 + 0.58*c2 + 0.22*cHarm
			dVal1 := -s1*k1 - 0.58*s2*k2 - 0.44*sHarm*k1
			gx += dVal1 * d[0]
			gy += dVal1 * d[1]
		}
		val *= normFac * 1.45
		gx *= normFac * 1.45
		gy *= normFac * 1.45
		return
	}

	grad := render.NewMultiStopGradient(
		abyssTrough,
		deepLapis,
		ceruleanWave,
		turquoiseCrest,
		amberRidge,
		coralPeak,
	)

	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
	heightGrid := make([]float64, w*h)

	var wg sync.WaitGroup
	for wIdx := 0; wIdx < workers; wIdx++ {
		y0 := (wIdx * h) / workers
		y1 := ((wIdx + 1) * h) / workers
		wg.Add(1)
		go func(rowStart, rowEnd int) {
			defer wg.Done()
			for y := rowStart; y < rowEnd; y++ {
				fy := float64(y) + 0.5
				rowOff := y * w
				for x := 0; x < w; x++ {
					fx := float64(x) + 0.5
					val, gx, gy := evalFaraday(fx, fy)

					// Normalized wave elevation u in [0, 1]
					u := common.Constrain(0.5+0.52*val, 0, 1)
					heightGrid[rowOff+x] = u

					nx, ny, nz := -gx*16.0*scale, -gy*16.0*scale, 1.0
					nLen := math.Sqrt(nx*nx + ny*ny + nz*nz)
					nx, ny, nz = nx/nLen, ny/nLen, nz/nLen

					ndotl := common.Constrain(-0.58*nx-0.64*ny+0.50*nz, 0, 1)
					ndotFill := common.Constrain(0.55*nx+0.52*ny+0.65*nz, 0, 1)
					spec := math.Pow(common.Constrain(-0.40*nx-0.46*ny+0.79*nz, 0, 1), 28.0)
					fresnel := math.Pow(1.0-nz, 2.2)

					baseCol := grad.SampleOklab(u)
					illum := common.Constrain(0.18+0.72*ndotl+0.22*ndotFill, 0.06, 1.0)
					col := render.LerpOklab(abyssTrough, baseCol, illum)
					if fresnel > 0.12 {
						col = render.LerpOklab(col, amberRidge, fresnel*0.35)
					}
					if spec > 0.03 {
						col = render.LerpOklab(col, ivorySpec, common.Constrain(spec*0.88, 0, 1))
					}

					rc.SetPixelLinear(x, y, render.SRGB8ToLinear(col.R), render.SRGB8ToLinear(col.G), render.SRGB8ToLinear(col.B), 1.0)
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	// Extract Marching Squares Quasicrystalline Crest Contours
	gridLookup := func(x, y float64) float64 {
		ix := common.ConstrainInt(int(x), 0, w-1)
		iy := common.ConstrainInt(int(y), 0, h-1)
		return heightGrid[iy*w+ix]
	}

	for _, iso := range []float64{0.34, 0.52, 0.68, 0.82} {
		contours := render.ExtractMarchingSquares(115, 115, fw, fh, iso, gridLookup)
		cPath := &canvas.Path{}
		for _, poly := range contours {
			if len(poly) < 3 {
				continue
			}
			smooth := render.CatmullRomSpline(poly, 3)
			cPath.MoveTo(smooth[0].X, smooth[0].Y)
			for i := 1; i < len(smooth); i++ {
				cPath.LineTo(smooth[i].X, smooth[i].Y)
			}
		}
		strokeCol := turquoiseCrest
		alpha := 0.48
		if iso > 0.60 {
			strokeCol = ivorySpec
			alpha = 0.72
		}
		rc.StrokePath(cPath, 0.90*scale, canvas.RoundCap, canvas.RoundJoin, strokeCol, alpha)
	}

	rc.ApplyBloom(0.82, 6, 0.24)
	c.MergeRenderCanvas(rc)
}
