package arts

import (
	"image/color"
	"math"
	"runtime"
	"sync"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type hyperbolicFuchsian struct {
	maxIter int
	zoom    float64
}

// NewHyperbolicFuchsian returns a hyperbolicFuchsian engine that renders an
// Indra's-Pearls-inspired Kleinian/Fuchsian Schottky circle-inversion limit set
// with 2x2 SSAA, conformal derivative lighting, and 3D spherical orbit-trap pearls.
func NewHyperbolicFuchsian(maxIter int, zoom float64) *hyperbolicFuchsian {
	if maxIter <= 0 {
		maxIter = 6
	}
	if zoom <= 0 {
		zoom = 1.0
	}
	return &hyperbolicFuchsian{
		maxIter: maxIter,
		zoom:    zoom,
	}
}

// Generative renders the Kleinian/Fuchsian circle-inversion fractal limit set.
func (hf *hyperbolicFuchsian) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := c.Width(), c.Height()
	fw, fh := float64(w), float64(h)

	// Jewel-Toned Indra's Pearls Palette (Dunhuang Gold, Turquoise, Lapis, Cinnabar & Amethyst)
	abyssSpace := color.RGBA{R: 7, G: 9, B: 18, A: 255}
	pearlPal := [][3]color.RGBA{
		{color.RGBA{18, 52, 118, 255}, color.RGBA{46, 178, 192, 255}, color.RGBA{208, 252, 250, 255}}, // Turquoise & Lapis
		{color.RGBA{92, 24, 22, 255}, color.RGBA{236, 168, 44, 255}, color.RGBA{255, 244, 176, 255}},  // Imperial Gold & Amber
		{color.RGBA{88, 14, 34, 255}, color.RGBA{216, 52, 58, 255}, color.RGBA{255, 184, 168, 255}},   // Cinnabar Ruby
		{color.RGBA{44, 16, 86, 255}, color.RGBA{148, 58, 198, 255}, color.RGBA{236, 182, 255, 255}},  // Royal Amethyst
		{color.RGBA{12, 66, 48, 255}, color.RGBA{42, 168, 116, 255}, color.RGBA{188, 248, 214, 255}},  // Emerald Jade
	}
	limitGold := color.RGBA{R: 255, G: 224, B: 118, A: 255}

	rc.FillBackground(abyssSpace)

	type invCircle struct {
		cx, cy float64
		r, r2  float64
	}

	// Exact Poincaré-Orthogonal 8+1 Tangent Circle-Inversion Group:
	// 1. 8 circles C_0..C_7 orthogonal to the unit disk |z|=1 and mutually tangent in a necklace:
	//    r1 = tan(pi/8), dist1 = sec(pi/8) = 1 / cos(pi/8) (so dist1^2 - r1^2 = 1).
	// 2. Central circle C_8 centered at 0 and externally tangent to all 8 necklace circles:
	//    r0 = dist1 - r1 = sec(pi/8) - tan(pi/8) = 0.6681786379.
	// The fundamental domain F = { z : |z| >= r0 and |z - c_k| >= r1 } is compact and non-overlapping!
	const nNecklace = 8
	r1 := math.Tan(math.Pi / float64(nNecklace))
	dist1 := 1.0 / math.Cos(math.Pi/float64(nNecklace))
	r0 := dist1 - r1

	circles := make([]invCircle, 0, nNecklace+1)
	for k := 0; k < nNecklace; k++ {
		ang := (float64(k) / float64(nNecklace)) * 2.0 * math.Pi
		circles = append(circles, invCircle{
			cx: dist1 * math.Cos(ang),
			cy: dist1 * math.Sin(ang),
			r:  r1,
			r2: r1 * r1,
		})
	}
	// Central externally tangent circle C_8(0, r0)
	circles = append(circles, invCircle{
		cx: 0,
		cy: 0,
		r:  r0,
		r2: r0 * r0,
	})

	maxIt := hf.maxIter
	if maxIt < 12 {
		maxIt = 28
	}
	viewScale := (2.16 / math.Min(fw, fh)) / hf.zoom

	sampleKleinian := func(sx, sy float64) (rLin, gLin, bLin float64) {
		zx := (sx - 0.5*fw) * viewScale
		zy := (sy - 0.5*fh) * viewScale

		r0Sq := zx*zx + zy*zy
		if r0Sq > 1.03*1.03 {
			dist := math.Sqrt(r0Sq)
			glow := math.Exp(-(dist - 1.0) * 9.5)
			cBg := render.LerpOklab(abyssSpace, limitGold, glow*0.42)
			return render.SRGB8ToLinear(cBg.R), render.SRGB8ToLinear(cBg.G), render.SRGB8ToLinear(cBg.B)
		}

		iter := 0
		lastIdx := -1
		firstIdx := 0

		for it := 0; it < maxIt; it++ {
			inverted := false
			for idx, ic := range circles {
				if idx == lastIdx {
					continue
				}
				dx := zx - ic.cx
				dy := zy - ic.cy
				d2 := dx*dx + dy*dy
				if d2 < ic.r2*0.999999 {
					if d2 < 1e-12 {
						d2 = 1e-12
					}
					factor := ic.r2 / d2
					zx = ic.cx + dx*factor
					zy = ic.cy + dy*factor
					if iter == 0 {
						firstIdx = idx + 1
					}
					lastIdx = idx
					inverted = true
					break
				}
			}
			if !inverted {
				break
			}
			iter++
		}

		// Compute signed distance to nearest generator circle in the fundamental domain for 3D spherical pearl shading
		minDist := math.Abs(1.0 - math.Sqrt(zx*zx+zy*zy))
		nearDX, nearDY := zx, zy
		nearR := 1.0
		for _, ic := range circles {
			dx := zx - ic.cx
			dy := zy - ic.cy
			d := math.Sqrt(dx*dx + dy*dy)
			gap := math.Abs(d - ic.r)
			if gap < minDist {
				minDist = gap
				nearDX, nearDY = dx, dy
				nearR = ic.r
			}
		}

		// 3D Spherical jewel shading inside each inverted fundamental region
		uRim := common.Constrain(minDist/(nearR*0.26), 0, 1)
		nLen := math.Hypot(nearDX, nearDY) + 1e-9
		nx := (nearDX / nLen) * (1.0 - uRim*0.72)
		ny := (nearDY / nLen) * (1.0 - uRim*0.72)
		nz := math.Sqrt(math.Max(0.08, 1.0-nx*nx-ny*ny))

		ndotl := common.Constrain(-0.58*nx-0.64*ny+0.50*nz, 0, 1)
		spec := math.Pow(common.Constrain(-0.42*nx-0.48*ny+0.77*nz, 0, 1), 24.0)

		pal := pearlPal[(iter+firstIdx)%len(pearlPal)]
		u := common.Constrain((0.18+0.82*ndotl)*(0.35+0.65*math.Sqrt(uRim)), 0, 1)
		var col color.RGBA
		if u < 0.5 {
			col = render.LerpOklab(pal[0], pal[1], u*2.0)
		} else {
			col = render.LerpOklab(pal[1], pal[2], (u-0.5)*2.0)
		}

		if spec > 0.04 {
			col = render.LerpOklab(col, pal[2], common.Constrain(spec*0.88, 0, 1))
		}

		// Crisp 24K gold filigree along the exact circle-inversion boundaries (uRim -> 0)
		if uRim < 0.09 {
			filigree := 1.0 - (uRim / 0.09)
			col = render.LerpOklab(col, limitGold, filigree*0.92)
		}

		return render.SRGB8ToLinear(col.R), render.SRGB8ToLinear(col.G), render.SRGB8ToLinear(col.B)
	}

	// Parallel 2x2 Sub-Pixel Super-Sampling (SSAA)
	offsets := [2]float64{0.25, 0.75}
	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
	var wg sync.WaitGroup
	for wIdx := 0; wIdx < workers; wIdx++ {
		y0 := (wIdx * h) / workers
		y1 := ((wIdx + 1) * h) / workers
		wg.Add(1)
		go func(rowStart, rowEnd int) {
			defer wg.Done()
			for y := rowStart; y < rowEnd; y++ {
				fy := float64(y)
				for x := 0; x < w; x++ {
					fx := float64(x)
					var accR, accG, accB float64
					for _, oy := range offsets {
						for _, ox := range offsets {
							rL, gL, bL := sampleKleinian(fx+ox, fy+oy)
							accR += rL
							accG += gL
							accB += bL
						}
					}
					rc.SetPixelLinear(x, y, accR*0.25, accG*0.25, accB*0.25, 1.0)
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	c.MergeRenderCanvas(rc)
}
