package common

import (
	"math/rand"
)

const (
	perlinYwrapb     = 4
	perlinYwrap      = 1 << perlinYwrapb
	perlinZwrapb     = 8
	perlinZwrap      = 1 << perlinZwrapb
	perlinSize       = 4095
	perlinOctaves    = 4
	perlinAmpFalloff = 0.5
)

// PerlinNoise generates multi-octave C2-continuous value/Perlin noise.
type PerlinNoise struct {
	perlin  []float64
	octaves int
	falloff float64
}

// NewPerlinNoise returns a PerlinNoise object initialized from package math/rand.
func NewPerlinNoise() *PerlinNoise {
	perlin := &PerlinNoise{
		perlin:  make([]float64, perlinSize+1),
		octaves: perlinOctaves,
		falloff: perlinAmpFalloff,
	}
	for i := range perlin.perlin {
		perlin.perlin[i] = rand.Float64()
	}
	return perlin
}

// NewPerlinNoiseWithSeed returns a deterministic PerlinNoise initialized with the given seed.
func NewPerlinNoiseWithSeed(seed int64) *PerlinNoise {
	rng := rand.New(rand.NewSource(seed))
	perlin := &PerlinNoise{
		perlin:  make([]float64, perlinSize+1),
		octaves: perlinOctaves,
		falloff: perlinAmpFalloff,
	}
	for i := range perlin.perlin {
		perlin.perlin[i] = rng.Float64()
	}
	return perlin
}

// SetDetail configures the number of octaves and amplitude falloff factor per octave.
func (p *PerlinNoise) SetDetail(octaves int, falloff float64) {
	if p == nil {
		return
	}
	if octaves > 0 {
		p.octaves = octaves
	}
	if falloff > 0 && falloff < 1.0 {
		p.falloff = falloff
	}
}

// Noise1D returns a float noise number on one dimension in [0, 1].
func (p *PerlinNoise) Noise1D(x float64) float64 {
	return p.noise(x, 0, 0)
}

// Noise2D returns a float noise number on two dimensions in [0, 1].
func (p *PerlinNoise) Noise2D(x, y float64) float64 {
	return p.noise(x, y, 0)
}

// Noise3D returns a float noise number on three dimensions in [0, 1].
func (p *PerlinNoise) Noise3D(x, y, z float64) float64 {
	return p.noise(x, y, z)
}

func (p *PerlinNoise) noise(x, y, z float64) float64 {
	if p == nil || len(p.perlin) <= perlinSize {
		return 0
	}
	octaves := p.octaves
	if octaves <= 0 {
		octaves = perlinOctaves
	}
	falloff := p.falloff
	if falloff <= 0 {
		falloff = perlinAmpFalloff
	}

	// True mathematical floor preserves C2 continuity across negative coordinates (no mirroring at 0)
	xi, yi, zi := fastFloor(x), fastFloor(y), fastFloor(z)
	xf, yf, zf := x-float64(xi), y-float64(yi), z-float64(zi)

	var rxf, ryf, rzf, n1, n2, n3 float64
	var r, norm float64
	ampl := 0.5

	for o := 0; o < octaves; o++ {
		of := (xi & perlinSize) + ((yi & perlinSize) << perlinYwrapb) + ((zi & perlinSize) << perlinZwrapb)

		rxf = quinticFade(xf)
		ryf = quinticFade(yf)
		rzf = quinticFade(zf)

		n1 = p.perlin[of&perlinSize]
		n1 += rxf * (p.perlin[(of+1)&perlinSize] - n1)
		n2 = p.perlin[(of+perlinYwrap)&perlinSize]
		n2 += rxf * (p.perlin[(of+perlinYwrap+1)&perlinSize] - n2)
		n1 += ryf * (n2 - n1)

		of += perlinZwrap
		n2 = p.perlin[of&perlinSize]
		n2 += rxf * (p.perlin[(of+1)&perlinSize] - n2)
		n3 = p.perlin[(of+perlinYwrap)&perlinSize]
		n3 += rxf * (p.perlin[(of+perlinYwrap+1)&perlinSize] - n3)
		n2 += ryf * (n3 - n2)

		n1 += rzf * (n2 - n1)

		r += n1 * ampl
		norm += ampl
		ampl *= falloff

		xi <<= 1
		xf *= 2.0
		yi <<= 1
		yf *= 2.0
		zi <<= 1
		zf *= 2.0

		if xf >= 1.0 {
			xi++
			xf -= 1.0
		}
		if yf >= 1.0 {
			yi++
			yf -= 1.0
		}
		if zf >= 1.0 {
			zi++
			zf -= 1.0
		}
	}
	if norm > 0 {
		return r / norm
	}
	return r
}

// quinticFade evaluates Ken Perlin's C2-continuous quintic interpolation polynomial:
// 6*t^5 - 15*t^4 + 10*t^3, which has zero 1st and 2nd derivatives at t=0 and t=1.
func quinticFade(t float64) float64 {
	return t * t * t * (t*(t*6.0-15.0) + 10.0)
}

// quinticFadeDeriv evaluates the exact analytical derivative of quinticFade(t):
// 30*t^4 - 60*t^3 + 30*t^2 = 30*t^2*(t-1)^2.
func quinticFadeDeriv(t float64) float64 {
	tm1 := t - 1.0
	return 30.0 * t * t * tm1 * tm1
}

func scaledCosin(x float64) float64 {
	return quinticFade(x)
}
