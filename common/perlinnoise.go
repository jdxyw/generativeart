package common

import (
	"math"
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

// PerlinNoise is a perlin noise struct to generate perlin noise.
type PerlinNoise struct {
	perlin []float64
}

// NewPerlinNoise returns a PerlinNoise object.
func NewPerlinNoise() *PerlinNoise {
	perlin := &PerlinNoise{perlin: nil}
	perlin.perlin = make([]float64, perlinSize+1)

	for i, _ := range perlin.perlin {
		perlin.perlin[i] = rand.Float64()
	}
	return perlin
}

// Noise1D returns a float noise number on one dimension.
func (p *PerlinNoise) Noise1D(x float64) float64 {
	return p.noise(x, 0, 0)
}

// Noise2D returns a float noise number on two dimensions.
func (p *PerlinNoise) Noise2D(x, y float64) float64 {
	return p.noise(x, y, 0)
}

// Noise3D returns a float noise number on three dimensions.
func (p *PerlinNoise) Noise3D(x, y, z float64) float64 {
	return p.noise(x, y, z)
}

func (p *PerlinNoise) noise(x, y, z float64) float64 {
	if x < 0 {
		x = -x
	}

	if y < 0 {
		y = -y
	}

	if z < 0 {
		z = -z
	}

	xi, yi, zi := int(math.Floor(x)), int(math.Floor(y)), int(math.Floor(z))
	xf, yf, zf := x-float64(xi), y-float64(yi), z-float64(zi)

	var rxf, ryf, n1, n2, n3 float64
	var r float64
	var ampl float64 = 0.5

	for o := 0; o < perlinOctaves; o++ {
		of := xi + (yi << perlinYwrapb) + (zi << perlinZwrapb)

		rxf = scaledCosin(xf)
		ryf = scaledCosin(yf)

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

		n1 += scaledCosin(zf) * (n2 - n1)

		r += n1 * ampl
		ampl *= perlinAmpFalloff
		xi <<= 1
		xf *= 2
		yi <<= 1
		yf *= 2
		zi <<= 1
		zf *= 2

		if xf >= 1.0 {
			xi += 1
			xf -= 1
		}
		if yf >= 1.0 {
			yi += 1
			yf -= 1
		}
		if zf >= 1.0 {
			zi += 1
			zf -= 1
		}
	}
	return r
}

func scaledCosin(x float64) float64 {
	return 0.5 * (1.0 - math.Cos(x*math.Pi))
}
