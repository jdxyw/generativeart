package render

import (
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart/common"
)

// ApplyBloom extracts bright linear-RGB pixels with luminance above threshold,
// applies a separable 1D Gaussian blur (horizontal + vertical) with standard deviation sigma,
// and adds the blurred HDR glow back into the canvas scaled by strength.
func (c *Canvas) ApplyBloom(threshold, sigma, strength float64) {
	if c == nil || c.Width <= 0 || c.Height <= 0 || sigma <= 0 || strength <= 0 {
		return
	}
	w, h := c.Width, c.Height
	nPix := w * h

	// 1. Extract bright pass in linear RGB
	bright := make([]float64, nPix*3)
	softKnee := math.Max(0.05, threshold*0.25)

	for i := 0; i < nPix; i++ {
		idx := i * 4
		r := c.pix[idx+0]
		g := c.pix[idx+1]
		b := c.pix[idx+2]
		a := c.pix[idx+3]
		if a <= 0 {
			continue
		}
		// Rec. 709 linear luminance
		lum := 0.2126*r + 0.7152*g + 0.0722*b
		if lum <= threshold-softKnee {
			continue
		}
		// Soft-knee quadratic extraction so bloom transitions smoothly without hard banding
		knee := lum - threshold + softKnee
		if knee < 2.0*softKnee {
			knee = (knee * knee) / (4.0 * softKnee)
		} else {
			knee = lum - threshold
		}
		scale := (knee / math.Max(lum, 1e-6)) * a
		if scale > 0 {
			bIdx := i * 3
			bright[bIdx+0] = r * scale
			bright[bIdx+1] = g * scale
			bright[bIdx+2] = b * scale
		}
	}

	// 2. Build 1D Gaussian kernel
	radius := int(math.Ceil(sigma * 3.0))
	if radius < 1 {
		radius = 1
	}
	if radius > 96 {
		radius = 96
	}
	kernel := make([]float64, 2*radius+1)
	sum := 0.0
	invTwoSigmaSq := 1.0 / (2.0 * sigma * sigma)
	for k := -radius; k <= radius; k++ {
		val := math.Exp(-float64(k*k) * invTwoSigmaSq)
		kernel[k+radius] = val
		sum += val
	}
	for k := range kernel {
		kernel[k] /= sum
	}

	// 3. Horizontal separable pass (parallelized across rows)
	temp := make([]float64, nPix*3)
	parallelRows(w, h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			rowOff := y * w
			for x := 0; x < w; x++ {
				var accR, accG, accB float64
				for k := -radius; k <= radius; k++ {
					sx := x + k
					if sx < 0 {
						sx = 0
					} else if sx >= w {
						sx = w - 1
					}
					wk := kernel[k+radius]
					sIdx := (rowOff + sx) * 3
					accR += bright[sIdx+0] * wk
					accG += bright[sIdx+1] * wk
					accB += bright[sIdx+2] * wk
				}
				dIdx := (rowOff + x) * 3
				temp[dIdx+0] = accR
				temp[dIdx+1] = accG
				temp[dIdx+2] = accB
			}
		}
	})

	// 4. Vertical separable pass + additive HDR composite (parallelized across rows)
	parallelRows(w, h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := 0; x < w; x++ {
				var accR, accG, accB float64
				for k := -radius; k <= radius; k++ {
					sy := y + k
					if sy < 0 {
						sy = 0
					} else if sy >= h {
						sy = h - 1
					}
					wk := kernel[k+radius]
					sIdx := (sy*w + x) * 3
					accR += temp[sIdx+0] * wk
					accG += temp[sIdx+1] * wk
					accB += temp[sIdx+2] * wk
				}
				cIdx := (y*w + x) * 4
				c.pix[cIdx+0] += accR * strength
				c.pix[cIdx+1] += accG * strength
				c.pix[cIdx+2] += accB * strength
				if c.pix[cIdx+3] < 1.0 {
					glowA := common.Constrain((accR+accG+accB)*strength*0.333, 0, 1)
					c.pix[cIdx+3] = c.pix[cIdx+3] + glowA*(1.0-c.pix[cIdx+3])
				}
			}
		}
	})
}

// hash2D is a fast deterministic 2D spatial hash returning a value in [-1, 1].
func hash2D(ix, iy int, seed uint64) float64 {
	h := uint64(uint32(ix))*0x9E3779B185EBCA87 + uint64(uint32(iy))*0xC2B2AE3D27D4EB4F + seed*0x165667B19E3779F9
	h ^= h >> 33
	h *= 0xFF51AFD7ED558CCD
	h ^= h >> 33
	h *= 0xC4CEB9FE1A85EC53
	h ^= h >> 33
	return (float64(h&0xFFFFFFFF)/float64(0xFFFFFFFF))*2.0 - 1.0
}

// valueNoise2D evaluates smooth bicubic-interpolated 2D value noise in [-1, 1].
func valueNoise2D(x, y float64, seed uint64) float64 {
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	fx := x - float64(x0)
	fy := y - float64(y0)

	u := fx * fx * (3.0 - 2.0*fx)
	v := fy * fy * (3.0 - 2.0*fy)

	n00 := hash2D(x0, y0, seed)
	n10 := hash2D(x0+1, y0, seed)
	n01 := hash2D(x0, y0+1, seed)
	n11 := hash2D(x0+1, y0+1, seed)

	nx0 := n00 + u*(n10-n00)
	nx1 := n01 + u*(n11-n01)
	return nx0 + v*(nx1-nx0)
}

// ApplyPaperTexture applies a multi-scale procedural cotton paper fiber & micro-grain modulation
// to the canvas in linear light.
func (c *Canvas) ApplyPaperTexture(intensity float64, seed int64) {
	if c == nil || c.Width <= 0 || c.Height <= 0 || intensity <= 0 {
		return
	}
	intensity = common.Constrain(intensity, 0, 1)
	rng := rand.New(rand.NewSource(seed))
	s1 := rng.Uint64()
	s2 := rng.Uint64()
	s3 := rng.Uint64()

	w, h := c.Width, c.Height
	parallelRows(w, h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			fy := float64(y)
			for x := 0; x < w; x++ {
				fx := float64(x)

				// Coarse paper clouding + medium directional weave + fine micro-tooth
				coarse := valueNoise2D(fx*0.035, fy*0.035, s1)
				fiberX := valueNoise2D(fx*0.25, fy*0.06, s2)
				fiberY := valueNoise2D(fx*0.06, fy*0.25, s2^0xA5A5A5A5)
				fine := hash2D(x, y, s3)

				paper := 0.35*coarse + 0.25*fiberX + 0.25*fiberY + 0.15*fine
				mod := 1.0 + paper*intensity*0.14
				lift := paper * intensity * 0.008

				idx := (y*w + x) * 4
				c.pix[idx+0] = math.Max(0, c.pix[idx+0]*mod+lift)
				c.pix[idx+1] = math.Max(0, c.pix[idx+1]*mod+lift)
				c.pix[idx+2] = math.Max(0, c.pix[idx+2]*mod+lift)
			}
		}
	})
}

// ApplyVignette applies a smooth cosine/polynomial radial darkening near the canvas edges.
func (c *Canvas) ApplyVignette(strength float64) {
	if c == nil || c.Width <= 0 || c.Height <= 0 || strength <= 0 {
		return
	}
	strength = common.Constrain(strength, 0, 1)
	cx := float64(c.Width) * 0.5
	cy := float64(c.Height) * 0.5
	invRx := 1.0 / math.Max(1.0, cx)
	invRy := 1.0 / math.Max(1.0, cy)

	parallelRows(c.Width, c.Height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			dy := (float64(y) + 0.5 - cy) * invRy
			for x := 0; x < c.Width; x++ {
				dx := (float64(x) + 0.5 - cx) * invRx
				// r2 is 0 at center, 1 at midpoint of edges, 2 at corners
				r2 := (dx*dx + dy*dy) * 0.5
				t := common.Constrain(r2, 0, 1)
				smooth := t * t * (3.0 - 2.0*t)
				factor := 1.0 - strength*smooth*0.65

				idx := (y*c.Width + x) * 4
				c.pix[idx+0] *= factor
				c.pix[idx+1] *= factor
				c.pix[idx+2] *= factor
			}
		}
	})
}
