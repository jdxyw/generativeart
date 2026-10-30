package render

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart/common"
)

// PostEffect represents a composable 64-bit linear HDR post-processing pass on a Canvas.
type PostEffect interface {
	Name() string
	Apply(c *Canvas)
}

type funcPostEffect struct {
	name string
	fn   func(c *Canvas)
}

// NewPostEffect wraps a custom HDR canvas transformation callback into a named PostEffect.
func NewPostEffect(name string, fn func(c *Canvas)) PostEffect {
	return &funcPostEffect{name: name, fn: fn}
}

func (f *funcPostEffect) Name() string {
	if f == nil || f.name == "" {
		return "custom"
	}
	return f.name
}

func (f *funcPostEffect) Apply(c *Canvas) {
	if f != nil && f.fn != nil && c != nil {
		f.fn(c)
	}
}

// ApplyPostEffects executes a sequence of PostEffect passes in-place on the linear HDR Canvas.
func (c *Canvas) ApplyPostEffects(effects ...PostEffect) {
	if c == nil {
		return
	}
	for _, eff := range effects {
		if eff != nil {
			eff.Apply(c)
		}
	}
}

// BloomEffect returns a composable PostEffect that applies HDR Gaussian bloom.
func BloomEffect(threshold, sigma, strength float64) PostEffect {
	return NewPostEffect(
		fmt.Sprintf("bloom(%.2f,%.2f,%.2f)", threshold, sigma, strength),
		func(c *Canvas) { c.ApplyBloom(threshold, sigma, strength) },
	)
}

// PaperTextureEffect returns a composable PostEffect that modulates the canvas with procedural cotton paper grain.
func PaperTextureEffect(intensity float64, seed int64) PostEffect {
	return NewPostEffect(
		fmt.Sprintf("paper(%.2f,%d)", intensity, seed),
		func(c *Canvas) { c.ApplyPaperTexture(intensity, seed) },
	)
}

// VignetteEffect returns a composable PostEffect that applies smooth radial edge darkening.
func VignetteEffect(strength float64) PostEffect {
	return NewPostEffect(
		fmt.Sprintf("vignette(%.2f)", strength),
		func(c *Canvas) { c.ApplyVignette(strength) },
	)
}

// ToneMapEffect returns a composable PostEffect that applies ACES filmic tone mapping in-place.
func ToneMapEffect(exposure, gamma float64) PostEffect {
	return NewPostEffect(
		fmt.Sprintf("tonemap(%.2f,%.2f)", exposure, gamma),
		func(c *Canvas) { c.ApplyToneMap(exposure, gamma) },
	)
}

// ChromaticAberrationEffect returns a composable PostEffect that applies radial R/G/B lens dispersion.
func ChromaticAberrationEffect(strength float64) PostEffect {
	return NewPostEffect(
		fmt.Sprintf("chromatic(%.2f)", strength),
		func(c *Canvas) { c.ApplyChromaticAberration(strength) },
	)
}

// LuminanceReliefEffect returns a composable PostEffect that shades perceptual luminance as a 2.5D relief surface with micro-SSAO.
func LuminanceReliefEffect(strength, azimuthDeg, elevationDeg, specStrength float64) PostEffect {
	return NewPostEffect(
		fmt.Sprintf("relief(%.2f,%.0f,%.0f,%.2f)", strength, azimuthDeg, elevationDeg, specStrength),
		func(c *Canvas) { c.ApplyLuminanceRelief(strength, azimuthDeg, elevationDeg, specStrength) },
	)
}

// AnisotropicKuwaharaEffect returns a composable PostEffect that applies structure-tensor flow-aligned painterly filtering.
func AnisotropicKuwaharaEffect(radius int, sharpness, anisotropy float64) PostEffect {
	return NewPostEffect(
		fmt.Sprintf("kuwahara(%d,%.1f,%.2f)", radius, sharpness, anisotropy),
		func(c *Canvas) { c.ApplyAnisotropicKuwahara(radius, sharpness, anisotropy) },
	)
}

// ColorGradeEffect returns a composable PostEffect that performs Oklab contrast, chroma scaling, and shadow/highlight split-toning.
func ColorGradeEffect(contrast, saturation, shadowHueDeg, highlightHueDeg, splitStrength float64) PostEffect {
	return NewPostEffect(
		fmt.Sprintf("grade(%.2f,%.2f,%.0f,%.0f,%.2f)", contrast, saturation, shadowHueDeg, highlightHueDeg, splitStrength),
		func(c *Canvas) {
			c.ApplyColorGrade(contrast, saturation, shadowHueDeg, highlightHueDeg, splitStrength)
		},
	)
}

// DeckleBorderEffect returns a composable PostEffect that frames the canvas with a torn deckle-edged museum mat.
func DeckleBorderEffect(marginFrac, roughness float64, matCol color.RGBA, seed int64) PostEffect {
	return NewPostEffect(
		fmt.Sprintf("deckle(%.3f,%.2f,%d)", marginFrac, roughness, seed),
		func(c *Canvas) { c.ApplyDeckleBorder(marginFrac, roughness, matCol, seed) },
	)
}


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

// ApplyToneMap applies ACES filmic tone mapping in-place on the linear HDR buffer.
// If gamma > 0, an additional power adjustment relative to standard sRGB (2.2 / gamma) is applied.
func (c *Canvas) ApplyToneMap(exposure, gamma float64) {
	if c == nil || c.Width <= 0 || c.Height <= 0 {
		return
	}
	if exposure <= 0 {
		exposure = 1.0
	}
	customGamma := gamma > 0 && math.Abs(gamma-2.2) > 1e-3
	gammaExp := 1.0
	if customGamma {
		gammaExp = 2.2 / gamma
	}

	parallelRows(c.Width, c.Height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := 0; x < c.Width; x++ {
				idx := (y*c.Width + x) * 4
				r := acesFilmic(c.pix[idx+0] * exposure)
				g := acesFilmic(c.pix[idx+1] * exposure)
				b := acesFilmic(c.pix[idx+2] * exposure)
				if customGamma {
					r = math.Pow(r, gammaExp)
					g = math.Pow(g, gammaExp)
					b = math.Pow(b, gammaExp)
				}
				c.pix[idx+0] = r
				c.pix[idx+1] = g
				c.pix[idx+2] = b
			}
		}
	})
}

// sampleLinearRGBBilinear performs sub-pixel bilinear interpolation on a 3-channel (R,G,B) buffer.
func sampleLinearRGBBilinear(src []float64, w, h int, fx, fy float64) (r, g, b float64) {
	fx = common.Constrain(fx, 0, float64(w-1))
	fy = common.Constrain(fy, 0, float64(h-1))
	x0 := int(math.Floor(fx))
	y0 := int(math.Floor(fy))
	x1 := x0 + 1
	y1 := y0 + 1
	if x1 >= w {
		x1 = w - 1
	}
	if y1 >= h {
		y1 = h - 1
	}
	tx := fx - float64(x0)
	ty := fy - float64(y0)

	i00 := (y0*w + x0) * 3
	i10 := (y0*w + x1) * 3
	i01 := (y1*w + x0) * 3
	i11 := (y1*w + x1) * 3

	w00 := (1.0 - tx) * (1.0 - ty)
	w10 := tx * (1.0 - ty)
	w01 := (1.0 - tx) * ty
	w11 := tx * ty

	r = src[i00+0]*w00 + src[i10+0]*w10 + src[i01+0]*w01 + src[i11+0]*w11
	g = src[i00+1]*w00 + src[i10+1]*w10 + src[i01+1]*w01 + src[i11+1]*w11
	b = src[i00+2]*w00 + src[i10+2]*w10 + src[i01+2]*w01 + src[i11+2]*w11
	return
}

// ApplyChromaticAberration applies radial optical lens dispersion in 64-bit linear HDR space,
// shifting Red outward and Blue inward with cubic radial falloff and sub-pixel bilinear sampling.
func (c *Canvas) ApplyChromaticAberration(strength float64) {
	if c == nil || c.Width <= 0 || c.Height <= 0 || math.Abs(strength) <= 1e-5 {
		return
	}
	w, h := c.Width, c.Height
	nPix := w * h
	src := make([]float64, nPix*3)
	for i := 0; i < nPix; i++ {
		sIdx := i * 4
		dIdx := i * 3
		src[dIdx+0] = c.pix[sIdx+0]
		src[dIdx+1] = c.pix[sIdx+1]
		src[dIdx+2] = c.pix[sIdx+2]
	}

	cx := float64(w) * 0.5
	cy := float64(h) * 0.5
	halfDiag := math.Hypot(cx, cy)
	if halfDiag < 1.0 {
		halfDiag = 1.0
	}
	// Maximum pixel shift at corners (~2.8% of half-diagonal at strength=1.0)
	maxShift := strength * halfDiag * 0.028

	parallelRows(w, h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			fy := float64(y)
			dy := (fy + 0.5 - cy) / halfDiag
			for x := 0; x < w; x++ {
				fx := float64(x)
				dx := (fx + 0.5 - cx) / halfDiag
				r2 := dx*dx + dy*dy
				// Smooth linear + cubic dispersion profile
				shift := maxShift * (0.30 + 0.70*r2)
				sx := dx * shift
				sy := dy * shift

				rChan, _, _ := sampleLinearRGBBilinear(src, w, h, fx+sx, fy+sy)
				gChan := src[(y*w+x)*3+1]
				_, _, bChan := sampleLinearRGBBilinear(src, w, h, fx-sx, fy-sy)

				idx := (y*w + x) * 4
				c.pix[idx+0] = rChan
				c.pix[idx+1] = gChan
				c.pix[idx+2] = bChan
			}
		}
	})
}

// ApplyLuminanceRelief treats perceptual Oklab lightness L(x,y) as a continuous 2.5D heightfield,
// computing 3x3 Scharr surface normals, 8-tap micro-cavity ambient occlusion, and Blinn-Phong
// directional grazing light + specular highlights in linear HDR space.
func (c *Canvas) ApplyLuminanceRelief(strength, azimuthDeg, elevationDeg, specStrength float64) {
	if c == nil || c.Width <= 0 || c.Height <= 0 || strength <= 0 {
		return
	}
	w, h := c.Width, c.Height
	nPix := w * h
	lum := make([]float64, nPix)

	parallelRows(w, h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			rowOff := y * w
			for x := 0; x < w; x++ {
				idx := (rowOff + x) * 4
				lab := LinearRGBAToOklab(
					math.Max(0, c.pix[idx+0]),
					math.Max(0, c.pix[idx+1]),
					math.Max(0, c.pix[idx+2]),
				)
				lum[rowOff+x] = lab.L
			}
		}
	})

	az := azimuthDeg * math.Pi / 180.0
	el := common.Constrain(elevationDeg, 10.0, 85.0) * math.Pi / 180.0
	lx := math.Cos(az) * math.Cos(el)
	ly := math.Sin(az) * math.Cos(el)
	lz := math.Sin(el)

	// Blinn-Phong half-vector H = norm(L + V) with V = (0, 0, 1)
	hx, hy, hz := lx, ly, lz+1.0
	hLen := math.Sqrt(hx*hx + hy*hy + hz*hz)
	if hLen > 1e-9 {
		hx, hy, hz = hx/hLen, hy/hLen, hz/hLen
	}

	baseFlatDiff := 0.35 + 0.65*lz
	if baseFlatDiff < 0.25 {
		baseFlatDiff = 0.25
	}

	aoTaps := [8][2]int{
		{-1, 0}, {1, 0}, {0, -1}, {0, 1},
		{-2, -2}, {2, -2}, {-2, 2}, {2, 2},
	}

	bumpScale := strength * 6.5
	specScale := math.Max(0, specStrength) * 0.45

	parallelRows(w, h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			ym1 := common.MaxInt(0, y-1)
			yp1 := common.MinInt(h-1, y+1)
			for x := 0; x < w; x++ {
				xm1 := common.MaxInt(0, x-1)
				xp1 := common.MinInt(w-1, x+1)

				// 3x3 Scharr gradient filter
				tl := lum[ym1*w+xm1]
				tc := lum[ym1*w+x]
				tr := lum[ym1*w+xp1]
				ml := lum[y*w+xm1]
				mc := lum[y*w+x]
				mr := lum[y*w+xp1]
				bl := lum[yp1*w+xm1]
				bc := lum[yp1*w+x]
				br := lum[yp1*w+xp1]

				gx := ((3.0*tr + 10.0*mr + 3.0*br) - (3.0*tl + 10.0*ml + 3.0*bl)) / 32.0
				gy := ((3.0*bl + 10.0*bc + 3.0*br) - (3.0*tl + 10.0*tc + 3.0*tr)) / 32.0

				nx := -gx * bumpScale
				ny := -gy * bumpScale
				nz := 1.0
				invLen := 1.0 / math.Sqrt(nx*nx+ny*ny+nz*nz)
				nx, ny, nz = nx*invLen, ny*invLen, nz*invLen

				// 8-tap micro-cavity ambient occlusion
				occ := 0.0
				for _, tap := range aoTaps {
					sx := common.ConstrainInt(x+tap[0], 0, w-1)
					sy := common.ConstrainInt(y+tap[1], 0, h-1)
					dh := lum[sy*w+sx] - mc
					if dh > 0.005 {
						occ += common.Constrain(dh*8.0, 0, 1)
					}
				}
				ao := common.Constrain(1.0-(occ/8.0)*0.45*common.Constrain(strength, 0, 1.5), 0.45, 1.0)

				ndotl := common.Constrain(nx*lx+ny*ly+nz*lz, -0.25, 1.0)
				diffRatio := ((0.35 + 0.65*math.Max(0, ndotl)) / baseFlatDiff) * ao

				ndoth := math.Max(0, nx*hx+ny*hy+nz*hz)
				spec := specScale * math.Pow(ndoth, 28.0) * ao

				idx := (y*w + x) * 4
				c.pix[idx+0] = math.Max(0, c.pix[idx+0]*diffRatio+spec*1.04)
				c.pix[idx+1] = math.Max(0, c.pix[idx+1]*diffRatio+spec*0.98)
				c.pix[idx+2] = math.Max(0, c.pix[idx+2]*diffRatio+spec*0.90)
			}
		}
	})
}

// ApplyAnisotropicKuwahara applies a structure-tensor-driven Anisotropic Kuwahara filter
// in linear HDR space. Local smoothed gradient tensors orient elliptical neighborhood
// sectors along dominant flow tangents, homogenizing brush planes while keeping edges crisp.
func (c *Canvas) ApplyAnisotropicKuwahara(radius int, sharpness, anisotropy float64) {
	if c == nil || c.Width <= 0 || c.Height <= 0 || radius <= 0 {
		return
	}
	if radius > 12 {
		radius = 12
	}
	if sharpness <= 0 {
		sharpness = 8.0
	}
	anisotropy = common.Constrain(anisotropy, 0.0, 2.5)

	w, h := c.Width, c.Height
	nPix := w * h
	src := make([]float64, nPix*3)
	lum := make([]float64, nPix)

	for i := 0; i < nPix; i++ {
		idx := i * 4
		r := math.Max(0, c.pix[idx+0])
		g := math.Max(0, c.pix[idx+1])
		b := math.Max(0, c.pix[idx+2])
		src[i*3+0] = r
		src[i*3+1] = g
		src[i*3+2] = b
		lum[i] = 0.2126*r + 0.7152*g + 0.0722*b
	}

	// 1. Compute raw structure tensor J = [gx^2, gx*gy, gy^2]
	jxx := make([]float64, nPix)
	jxy := make([]float64, nPix)
	jyy := make([]float64, nPix)

	parallelRows(w, h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			ym1 := common.MaxInt(0, y-1)
			yp1 := common.MinInt(h-1, y+1)
			for x := 0; x < w; x++ {
				xm1 := common.MaxInt(0, x-1)
				xp1 := common.MinInt(w-1, x+1)

				gx := (lum[ym1*w+xp1] + 2.0*lum[y*w+xp1] + lum[yp1*w+xp1] -
					(lum[ym1*w+xm1] + 2.0*lum[y*w+xm1] + lum[yp1*w+xm1])) * 0.125
				gy := (lum[yp1*w+xm1] + 2.0*lum[yp1*w+x] + lum[yp1*w+xp1] -
					(lum[ym1*w+xm1] + 2.0*lum[ym1*w+x] + lum[ym1*w+xp1])) * 0.125

				i := y*w + x
				jxx[i] = gx * gx
				jxy[i] = gx * gy
				jyy[i] = gy * gy
			}
		}
	})

	// 2. Smooth structure tensor with 3x3 binomial kernel and evaluate 4 oriented Kuwahara quadrants
	rf := float64(radius)
	rSqMax := (rf + 0.5) * (rf + 0.5)

	parallelRows(w, h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := 0; x < w; x++ {
				var sxx, sxy, syy float64
				for dy := -1; dy <= 1; dy++ {
					sy := common.ConstrainInt(y+dy, 0, h-1)
					wy := 1.0
					if dy == 0 {
						wy = 2.0
					}
					for dx := -1; dx <= 1; dx++ {
						sx := common.ConstrainInt(x+dx, 0, w-1)
						wx := 1.0
						if dx == 0 {
							wx = 2.0
						}
						wt := wx * wy
						ti := sy*w + sx
						sxx += jxx[ti] * wt
						sxy += jxy[ti] * wt
						syy += jyy[ti] * wt
					}
				}

				// Dominant orientation & local anisotropy A in [0, 1]
				diff := sxx - syy
				disc := math.Hypot(diff, 2.0*sxy)
				lam1 := 0.5 * (sxx + syy + disc)
				lam2 := 0.5 * (sxx + syy - disc)
				anis := 0.0
				if lam1+lam2 > 1e-9 {
					anis = (lam1 - lam2) / (lam1 + lam2)
				}
				// Flow tangent angle is perpendicular to gradient eigenvector
				gradAngle := 0.5 * math.Atan2(2.0*sxy, diff)
				flowAngle := gradAngle + math.Pi*0.5
				cosA := math.Cos(flowAngle)
				sinA := math.Sin(flowAngle)

				// Stretch along flow tangent, compress across normal
				scaleAlong := 1.0 + anisotropy*anis
				scaleCross := 1.0 / scaleAlong

				var mR, mG, mB [4]float64
				var sR, sG, sB [4]float64
				var wSum [4]float64

				for dy := -radius; dy <= radius; dy++ {
					sy := common.ConstrainInt(y+dy, 0, h-1)
					fdy := float64(dy)
					for dx := -radius; dx <= radius; dx++ {
						fdx := float64(dx)
						// Project (dx, dy) into flow-aligned frame (u along tangent, v across normal)
						u := (fdx*cosA + fdy*sinA) / scaleAlong
						v := (-fdx*sinA + fdy*cosA) / scaleCross
						distSq := u*u + v*v
						if distSq > rSqMax {
							continue
						}

						sx := common.ConstrainInt(x+dx, 0, w-1)
						pIdx := (sy*w + sx) * 3
						pr, pg, pb := src[pIdx+0], src[pIdx+1], src[pIdx+2]

						// Assign to overlapping quadrants in (u, v) plane
						spatialW := math.Exp(-distSq / (0.5 * rf * rf))
						for q := 0; q < 4; q++ {
							inQ := false
							switch q {
							case 0:
								inQ = u >= -0.25 && v >= -0.25
							case 1:
								inQ = u <= 0.25 && v >= -0.25
							case 2:
								inQ = u <= 0.25 && v <= 0.25
							case 3:
								inQ = u >= -0.25 && v <= 0.25
							}
							if inQ {
								wSum[q] += spatialW
								mR[q] += pr * spatialW
								mG[q] += pg * spatialW
								mB[q] += pb * spatialW
								sR[q] += pr * pr * spatialW
								sG[q] += pg * pg * spatialW
								sB[q] += pb * pb * spatialW
							}
						}
					}
				}

				var outR, outG, outB, totalWeight float64
				for q := 0; q < 4; q++ {
					if wSum[q] <= 1e-9 {
						continue
					}
					invW := 1.0 / wSum[q]
					mr := mR[q] * invW
					mg := mG[q] * invW
					mb := mB[q] * invW
					vr := math.Max(0, sR[q]*invW-mr*mr)
					vg := math.Max(0, sG[q]*invW-mg*mg)
					vb := math.Max(0, sB[q]*invW-mb*mb)
					sigma2 := vr + vg + vb
					qw := 1.0 / (1.0 + math.Pow(256.0*sigma2, 0.5*sharpness))
					outR += mr * qw
					outG += mg * qw
					outB += mb * qw
					totalWeight += qw
				}

				if totalWeight > 1e-12 {
					idx := (y*w + x) * 4
					c.pix[idx+0] = outR / totalWeight
					c.pix[idx+1] = outG / totalWeight
					c.pix[idx+2] = outB / totalWeight
				}
			}
		}
	})
}

// ApplyColorGrade performs perceptual Oklab color grading: S-curve lightness contrast,
// chroma saturation scaling, and shadow/highlight split-toning along shadowHueDeg & highlightHueDeg.
func (c *Canvas) ApplyColorGrade(contrast, saturation, shadowHueDeg, highlightHueDeg, splitStrength float64) {
	if c == nil || c.Width <= 0 || c.Height <= 0 {
		return
	}
	if contrast <= 0 {
		contrast = 1.0
	}
	if saturation < 0 {
		saturation = 1.0
	}
	splitStrength = common.Constrain(splitStrength, 0, 1)

	shRad := shadowHueDeg * math.Pi / 180.0
	hiRad := highlightHueDeg * math.Pi / 180.0
	shA := math.Cos(shRad) * splitStrength * 0.065
	shB := math.Sin(shRad) * splitStrength * 0.065
	hiA := math.Cos(hiRad) * splitStrength * 0.065
	hiB := math.Sin(hiRad) * splitStrength * 0.065

	parallelRows(c.Width, c.Height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := 0; x < c.Width; x++ {
				idx := (y*c.Width + x) * 4
				lab := LinearRGBAToOklab(
					math.Max(0, c.pix[idx+0]),
					math.Max(0, c.pix[idx+1]),
					math.Max(0, c.pix[idx+2]),
				)

				// Centered sigmoidal/power contrast around L = 0.5
				L := common.Constrain(lab.L, 0, 1)
				if math.Abs(contrast-1.0) > 1e-3 {
					if L < 0.5 {
						L = 0.5 * math.Pow(2.0*L, contrast)
					} else {
						L = 1.0 - 0.5*math.Pow(2.0*(1.0-L), contrast)
					}
				}

				A := lab.A * saturation
				B := lab.B * saturation

				if splitStrength > 0 {
					wShadow := (1.0 - L) * (1.0 - L)
					wHigh := L * L
					A += shA*wShadow + hiA*wHigh
					B += shB*wShadow + hiB*wHigh
				}

				rLin, gLin, bLin := OklabToLinearRGB(Oklab{L: L, A: A, B: B})
				c.pix[idx+0] = math.Max(0, rLin)
				c.pix[idx+1] = math.Max(0, gLin)
				c.pix[idx+2] = math.Max(0, bLin)
			}
		}
	})
}

// ApplyDeckleBorder frames the canvas with a ragged multi-octave FBM torn cotton-paper
// deckle edge and a recessed archival mat border of color matCol.
func (c *Canvas) ApplyDeckleBorder(marginFrac, roughness float64, matCol color.RGBA, seed int64) {
	if c == nil || c.Width <= 0 || c.Height <= 0 || marginFrac <= 0 {
		return
	}
	marginFrac = common.Constrain(marginFrac, 0.01, 0.25)
	roughness = common.Constrain(roughness, 0.0, 2.5)

	w, h := c.Width, c.Height
	minDim := math.Min(float64(w), float64(h))
	marginPx := marginFrac * minDim
	roughPx := roughness * marginPx * 0.28

	rng := rand.New(rand.NewSource(seed ^ 0x5DEECE66D))
	s1 := rng.Uint64()
	s2 := rng.Uint64()
	s3 := rng.Uint64()

	matR, matG, matB, _ := ColorToLinearRGB(matCol)

	parallelRows(w, h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			fy := float64(y) + 0.5
			for x := 0; x < w; x++ {
				fx := float64(x) + 0.5
				// Distance from nearest canvas edge
				dLeft := fx
				dRight := float64(w) - fx
				dTop := fy
				dBot := float64(h) - fy
				dEdge := math.Min(math.Min(dLeft, dRight), math.Min(dTop, dBot))

				if dEdge > marginPx+roughPx+6.0 {
					continue
				}

				// Multi-octave torn deckle displacement
				n1 := valueNoise2D(fx*0.04, fy*0.04, s1)
				n2 := valueNoise2D(fx*0.14, fy*0.14, s2)
				n3 := valueNoise2D(fx*0.42, fy*0.42, s3)
				tear := (0.55*n1 + 0.30*n2 + 0.15*n3) * roughPx
				signedDist := dEdge - (marginPx + tear)

				idx := (y*w + x) * 4
				artR, artG, artB := c.pix[idx+0], c.pix[idx+1], c.pix[idx+2]

				// Subtle paper grain on the outer museum mat
				grain := 1.0 + (0.5*n2+0.5*n3)*0.035
				mR := matR * grain
				mG := matG * grain
				mB := matB * grain

				if signedDist < 0 {
					// Outside the torn sheet: show recessed mat with soft cast shadow just outside the deckle tear
					shadow := 1.0 - 0.26*math.Exp(-(-signedDist)/3.2)
					cov := common.Smoothstep(-1.2, 0.0, signedDist)
					c.pix[idx+0] = mR*shadow*(1.0-cov) + artR*cov
					c.pix[idx+1] = mG*shadow*(1.0-cov) + artG*cov
					c.pix[idx+2] = mB*shadow*(1.0-cov) + artB*cov
				} else if signedDist < 2.2 {
					// Fibrous cotton rim highlight along the torn bevel edge
					rim := (1.0 - signedDist/2.2) * 0.06
					c.pix[idx+0] = artR + rim
					c.pix[idx+1] = artG + rim
					c.pix[idx+2] = artB + rim
				}
			}
		}
	})
}

