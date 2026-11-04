package common

import (
	"math"
	"math/rand"
)

// NoiseEngine provides seeded 2D/3D Simplex, FBM, Ridged, Curl, and Worley cellular noise.
type NoiseEngine struct {
	perm   [512]uint8
	perm12 [512]uint8
	seed   int64
}

var grad3 = [12][3]float64{
	{1, 1, 0}, {-1, 1, 0}, {1, -1, 0}, {-1, -1, 0},
	{1, 0, 1}, {-1, 0, 1}, {1, 0, -1}, {-1, 0, -1},
	{0, 1, 1}, {0, -1, 1}, {0, 1, -1}, {0, -1, -1},
}

var grad4 = [32][4]float64{
	{0, 1, 1, 1}, {0, 1, 1, -1}, {0, 1, -1, 1}, {0, 1, -1, -1},
	{0, -1, 1, 1}, {0, -1, 1, -1}, {0, -1, -1, 1}, {0, -1, -1, -1},
	{1, 0, 1, 1}, {1, 0, 1, -1}, {1, 0, -1, 1}, {1, 0, -1, -1},
	{-1, 0, 1, 1}, {-1, 0, 1, -1}, {-1, 0, -1, 1}, {-1, 0, -1, -1},
	{1, 1, 0, 1}, {1, 1, 0, -1}, {1, -1, 0, 1}, {1, -1, 0, -1},
	{-1, 1, 0, 1}, {-1, 1, 0, -1}, {-1, -1, 0, 1}, {-1, -1, 0, -1},
	{1, 1, 1, 0}, {1, 1, -1, 0}, {1, -1, 1, 0}, {1, -1, -1, 0},
	{-1, 1, 1, 0}, {-1, 1, -1, 0}, {-1, -1, 1, 0}, {-1, -1, -1, 0},
}

const (
	f2 = 0.3660254037844386  // 0.5 * (sqrt(3.0) - 1.0)
	g2 = 0.21132486540518713 // (3.0 - sqrt(3.0)) / 6.0
	f3 = 1.0 / 3.0
	g3 = 1.0 / 6.0
	f4 = 0.30901699437494745 // (sqrt(5.0) - 1.0) / 4.0
	g4 = 0.1381966011250105  // (5.0 - sqrt(5.0)) / 20.0
)

// NewNoiseEngine creates a deterministic NoiseEngine seeded with seed.
func NewNoiseEngine(seed int64) *NoiseEngine {
	ne := &NoiseEngine{seed: seed}
	rng := rand.New(rand.NewSource(seed))
	var p [256]uint8
	for i := 0; i < 256; i++ {
		p[i] = uint8(i)
	}
	for i := 255; i > 0; i-- {
		j := rng.Intn(i + 1)
		p[i], p[j] = p[j], p[i]
	}
	for i := 0; i < 512; i++ {
		v := p[i&255]
		ne.perm[i] = v
		ne.perm12[i] = v % 12
	}
	return ne
}

func fastFloor(x float64) int {
	xi := int(x)
	if x < float64(xi) {
		return xi - 1
	}
	return xi
}

func dot2(g [3]float64, x, y float64) float64 {
	return g[0]*x + g[1]*y
}

func dot3(g [3]float64, x, y, z float64) float64 {
	return g[0]*x + g[1]*y + g[2]*z
}

// Simplex2D evaluates 2D Simplex noise at (x, y), returning a value in [-1, 1].
func (ne *NoiseEngine) Simplex2D(x, y float64) float64 {
	s := (x + y) * f2
	i := fastFloor(x + s)
	j := fastFloor(y + s)
	t := float64(i+j) * g2
	X0 := float64(i) - t
	Y0 := float64(j) - t
	x0 := x - X0
	y0 := y - Y0

	var i1, j1 int
	if x0 > y0 {
		i1, j1 = 1, 0
	} else {
		i1, j1 = 0, 1
	}

	x1 := x0 - float64(i1) + g2
	y1 := y0 - float64(j1) + g2
	x2 := x0 - 1.0 + 2.0*g2
	y2 := y0 - 1.0 + 2.0*g2

	ii := i & 255
	jj := j & 255
	gi0 := ne.perm12[ii+int(ne.perm[jj])]
	gi1 := ne.perm12[ii+i1+int(ne.perm[jj+j1])]
	gi2 := ne.perm12[ii+1+int(ne.perm[jj+1])]

	var n0, n1, n2 float64
	t0 := 0.5 - x0*x0 - y0*y0
	if t0 >= 0 {
		t0 *= t0
		n0 = t0 * t0 * dot2(grad3[gi0], x0, y0)
	}
	t1 := 0.5 - x1*x1 - y1*y1
	if t1 >= 0 {
		t1 *= t1
		n1 = t1 * t1 * dot2(grad3[gi1], x1, y1)
	}
	t2 := 0.5 - x2*x2 - y2*y2
	if t2 >= 0 {
		t2 *= t2
		n2 = t2 * t2 * dot2(grad3[gi2], x2, y2)
	}

	return Constrain(70.0*(n0+n1+n2), -1.0, 1.0)
}

// Simplex2D01 evaluates 2D Simplex noise normalized to [0, 1].
func (ne *NoiseEngine) Simplex2D01(x, y float64) float64 {
	return Clamp01(0.5 * (ne.Simplex2D(x, y) + 1.0))
}

// Simplex3D evaluates 3D Simplex noise at (x, y, z), returning a value in [-1, 1].
func (ne *NoiseEngine) Simplex3D(x, y, z float64) float64 {
	s := (x + y + z) * f3
	i := fastFloor(x + s)
	j := fastFloor(y + s)
	k := fastFloor(z + s)
	t := float64(i+j+k) * g3
	X0 := float64(i) - t
	Y0 := float64(j) - t
	Z0 := float64(k) - t
	x0 := x - X0
	y0 := y - Y0
	z0 := z - Z0

	var i1, j1, k1 int
	var i2, j2, k2 int
	if x0 >= y0 {
		if y0 >= z0 {
			i1, j1, k1 = 1, 0, 0
			i2, j2, k2 = 1, 1, 0
		} else if x0 >= z0 {
			i1, j1, k1 = 1, 0, 0
			i2, j2, k2 = 1, 0, 1
		} else {
			i1, j1, k1 = 0, 0, 1
			i2, j2, k2 = 1, 0, 1
		}
	} else {
		if y0 < z0 {
			i1, j1, k1 = 0, 0, 1
			i2, j2, k2 = 0, 1, 1
		} else if x0 < z0 {
			i1, j1, k1 = 0, 1, 0
			i2, j2, k2 = 0, 1, 1
		} else {
			i1, j1, k1 = 0, 1, 0
			i2, j2, k2 = 1, 1, 0
		}
	}

	x1 := x0 - float64(i1) + g3
	y1 := y0 - float64(j1) + g3
	z1 := z0 - float64(k1) + g3
	x2 := x0 - float64(i2) + 2.0*g3
	y2 := y0 - float64(j2) + 2.0*g3
	z2 := z0 - float64(k2) + 2.0*g3
	x3 := x0 - 1.0 + 3.0*g3
	y3 := y0 - 1.0 + 3.0*g3
	z3 := z0 - 1.0 + 3.0*g3

	ii := i & 255
	jj := j & 255
	kk := k & 255
	gi0 := ne.perm12[ii+int(ne.perm[jj+int(ne.perm[kk])])]
	gi1 := ne.perm12[ii+i1+int(ne.perm[jj+j1+int(ne.perm[kk+k1])])]
	gi2 := ne.perm12[ii+i2+int(ne.perm[jj+j2+int(ne.perm[kk+k2])])]
	gi3 := ne.perm12[ii+1+int(ne.perm[jj+1+int(ne.perm[kk+1])])]

	var n0, n1, n2, n3 float64
	t0 := 0.6 - x0*x0 - y0*y0 - z0*z0
	if t0 >= 0 {
		t0 *= t0
		n0 = t0 * t0 * dot3(grad3[gi0], x0, y0, z0)
	}
	t1 := 0.6 - x1*x1 - y1*y1 - z1*z1
	if t1 >= 0 {
		t1 *= t1
		n1 = t1 * t1 * dot3(grad3[gi1], x1, y1, z1)
	}
	t2 := 0.6 - x2*x2 - y2*y2 - z2*z2
	if t2 >= 0 {
		t2 *= t2
		n2 = t2 * t2 * dot3(grad3[gi2], x2, y2, z2)
	}
	t3 := 0.6 - x3*x3 - y3*y3 - z3*z3
	if t3 >= 0 {
		t3 *= t3
		n3 = t3 * t3 * dot3(grad3[gi3], x3, y3, z3)
	}

	return Constrain(32.0*(n0+n1+n2+n3), -1.0, 1.0)
}

// FBM2D evaluates fractal Brownian motion over 2D Simplex noise in [-1, 1].
func (ne *NoiseEngine) FBM2D(x, y float64, octaves int, lacunarity, gain float64) float64 {
	if octaves < 1 {
		octaves = 1
	}
	if lacunarity <= 0 {
		lacunarity = 2.0
	}
	if gain <= 0 {
		gain = 0.5
	}
	var sum, norm float64
	amp := 1.0
	freq := 1.0
	for i := 0; i < octaves; i++ {
		sum += amp * ne.Simplex2D(x*freq, y*freq)
		norm += amp
		amp *= gain
		freq *= lacunarity
	}
	if norm == 0 {
		return 0
	}
	return Constrain(sum/norm, -1.0, 1.0)
}

// Ridged2D evaluates multi-fractal ridged noise in [0, 1].
func (ne *NoiseEngine) Ridged2D(x, y float64, octaves int, lacunarity, gain float64) float64 {
	if octaves < 1 {
		octaves = 1
	}
	if lacunarity <= 0 {
		lacunarity = 2.0
	}
	if gain <= 0 {
		gain = 0.5
	}
	var sum, norm float64
	amp := 0.5
	freq := 1.0
	weight := 1.0
	for i := 0; i < octaves; i++ {
		signal := 1.0 - math.Abs(ne.Simplex2D(x*freq, y*freq))
		signal *= signal
		signal *= weight
		weight = Clamp01(signal * 2.0)
		sum += signal * amp
		norm += amp
		freq *= lacunarity
		amp *= gain
	}
	if norm == 0 {
		return 0
	}
	return Clamp01(sum / norm)
}

// Curl2D computes a divergence-free 2D curl vector field (dPsi/dy, -dPsi/dx) from FBM potential.
func (ne *NoiseEngine) Curl2D(x, y, eps float64) Vector {
	if eps <= 0 {
		eps = 1e-3
	}
	inv2Eps := 1.0 / (2.0 * eps)
	dPsiDy := (ne.FBM2D(x, y+eps, 3, 2.0, 0.5) - ne.FBM2D(x, y-eps, 3, 2.0, 0.5)) * inv2Eps
	dPsiDx := (ne.FBM2D(x+eps, y, 3, 2.0, 0.5) - ne.FBM2D(x-eps, y, 3, 2.0, 0.5)) * inv2Eps
	return Vector{X: dPsiDy, Y: -dPsiDx}
}

// Worley2D computes 2D cellular (Worley/Voronoi) noise, returning the nearest (f1) and second-nearest (f2) feature distances.
func (ne *NoiseEngine) Worley2D(x, y float64) (f1, f2 float64) {
	ix := fastFloor(x)
	iy := fastFloor(y)
	f1 = math.MaxFloat64
	f2 = math.MaxFloat64

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			cx := ix + dx
			cy := iy + dy
			h1 := ne.perm[(cx&255)+int(ne.perm[cy&255])]
			h2 := ne.perm[((cx*13+7)&255)+int(ne.perm[(cy*17+3)&255])]
			px := float64(cx) + float64(h1)/255.0
			py := float64(cy) + float64(h2)/255.0
			d := math.Hypot(x-px, y-py)
			if d < f1 {
				f2 = f1
				f1 = d
			} else if d < f2 {
				f2 = d
			}
		}
	}
	return f1, f2
}

func dot4(g [4]float64, x, y, z, w float64) float64 {
	return g[0]*x + g[1]*y + g[2]*z + g[3]*w
}

// Simplex4D evaluates 4D Simplex noise at (x, y, z, w), returning a value in [-1, 1].
func (ne *NoiseEngine) Simplex4D(x, y, z, w float64) float64 {
	s := (x + y + z + w) * f4
	i := fastFloor(x + s)
	j := fastFloor(y + s)
	k := fastFloor(z + s)
	l := fastFloor(w + s)
	t := float64(i+j+k+l) * g4

	X0 := float64(i) - t
	Y0 := float64(j) - t
	Z0 := float64(k) - t
	W0 := float64(l) - t

	x0 := x - X0
	y0 := y - Y0
	z0 := z - Z0
	w0 := w - W0

	// Rank-order the 4 coordinates to determine which of the 24 4D pentatopes we are in
	var rankx, ranky, rankz, rankw int
	if x0 > y0 {
		rankx++
	} else {
		ranky++
	}
	if x0 > z0 {
		rankx++
	} else {
		rankz++
	}
	if x0 > w0 {
		rankx++
	} else {
		rankw++
	}
	if y0 > z0 {
		ranky++
	} else {
		rankz++
	}
	if y0 > w0 {
		ranky++
	} else {
		rankw++
	}
	if z0 > w0 {
		rankz++
	} else {
		rankw++
	}

	b2i := func(cond bool) int {
		if cond {
			return 1
		}
		return 0
	}

	i1, j1, k1, l1 := b2i(rankx >= 3), b2i(ranky >= 3), b2i(rankz >= 3), b2i(rankw >= 3)
	i2, j2, k2, l2 := b2i(rankx >= 2), b2i(ranky >= 2), b2i(rankz >= 2), b2i(rankw >= 2)
	i3, j3, k3, l3 := b2i(rankx >= 1), b2i(ranky >= 1), b2i(rankz >= 1), b2i(rankw >= 1)

	x1 := x0 - float64(i1) + g4
	y1 := y0 - float64(j1) + g4
	z1 := z0 - float64(k1) + g4
	w1 := w0 - float64(l1) + g4

	x2 := x0 - float64(i2) + 2.0*g4
	y2 := y0 - float64(j2) + 2.0*g4
	z2 := z0 - float64(k2) + 2.0*g4
	w2 := w0 - float64(l2) + 2.0*g4

	x3 := x0 - float64(i3) + 3.0*g4
	y3 := y0 - float64(j3) + 3.0*g4
	z3 := z0 - float64(k3) + 3.0*g4
	w3 := w0 - float64(l3) + 3.0*g4

	x4 := x0 - 1.0 + 4.0*g4
	y4 := y0 - 1.0 + 4.0*g4
	z4 := z0 - 1.0 + 4.0*g4
	w4 := w0 - 1.0 + 4.0*g4

	ii := i & 255
	jj := j & 255
	kk := k & 255
	ll := l & 255

	gi0 := ne.perm[ii+int(ne.perm[jj+int(ne.perm[kk+int(ne.perm[ll])])])] & 31
	gi1 := ne.perm[ii+i1+int(ne.perm[jj+j1+int(ne.perm[kk+k1+int(ne.perm[ll+l1])])])] & 31
	gi2 := ne.perm[ii+i2+int(ne.perm[jj+j2+int(ne.perm[kk+k2+int(ne.perm[ll+l2])])])] & 31
	gi3 := ne.perm[ii+i3+int(ne.perm[jj+j3+int(ne.perm[kk+k3+int(ne.perm[ll+l3])])])] & 31
	gi4 := ne.perm[ii+1+int(ne.perm[jj+1+int(ne.perm[kk+1+int(ne.perm[ll+1])])])] & 31

	var n0, n1, n2, n3, n4 float64
	if t0 := 0.6 - x0*x0 - y0*y0 - z0*z0 - w0*w0; t0 >= 0 {
		t0 *= t0
		n0 = t0 * t0 * dot4(grad4[gi0], x0, y0, z0, w0)
	}
	if t1 := 0.6 - x1*x1 - y1*y1 - z1*z1 - w1*w1; t1 >= 0 {
		t1 *= t1
		n1 = t1 * t1 * dot4(grad4[gi1], x1, y1, z1, w1)
	}
	if t2 := 0.6 - x2*x2 - y2*y2 - z2*z2 - w2*w2; t2 >= 0 {
		t2 *= t2
		n2 = t2 * t2 * dot4(grad4[gi2], x2, y2, z2, w2)
	}
	if t3 := 0.6 - x3*x3 - y3*y3 - z3*z3 - w3*w3; t3 >= 0 {
		t3 *= t3
		n3 = t3 * t3 * dot4(grad4[gi3], x3, y3, z3, w3)
	}
	if t4 := 0.6 - x4*x4 - y4*y4 - z4*z4 - w4*w4; t4 >= 0 {
		t4 *= t4
		n4 = t4 * t4 * dot4(grad4[gi4], x4, y4, z4, w4)
	}

	return Constrain(27.0*(n0+n1+n2+n3+n4), -1.0, 1.0)
}

// FBM4DLoop evaluates fractal Brownian motion over 4D Simplex noise along a closed circle
// of radius loopRadius in the (z, w) plane parameterized by phase in [0, 1), producing a
// seamless periodic noise field where FBM4DLoop(x, y, 0, ...) == FBM4DLoop(x, y, 1, ...).
func (ne *NoiseEngine) FBM4DLoop(x, y, phase, loopRadius float64, octaves int, lacunarity, gain float64) float64 {
	if octaves < 1 {
		octaves = 1
	}
	if lacunarity <= 0 {
		lacunarity = 2.0
	}
	if gain <= 0 {
		gain = 0.5
	}
	if loopRadius <= 0 {
		loopRadius = 1.0
	}
	ang := phase * 2.0 * math.Pi
	z := loopRadius * math.Cos(ang)
	w := loopRadius * math.Sin(ang)

	var sum, norm float64
	amp := 1.0
	freq := 1.0
	for i := 0; i < octaves; i++ {
		sum += amp * ne.Simplex4D(x*freq, y*freq, z*freq, w*freq)
		norm += amp
		amp *= gain
		freq *= lacunarity
	}
	if norm == 0 {
		return 0
	}
	return Constrain(sum/norm, -1.0, 1.0)
}

// Curl4DLoop computes a divergence-free 2D curl vector field evolving smoothly over a
// periodic time loop parameterized by phase in [0, 1).
func (ne *NoiseEngine) Curl4DLoop(x, y, phase, loopRadius, eps float64) Vector {
	if eps <= 0 {
		eps = 1e-3
	}
	inv2Eps := 1.0 / (2.0 * eps)
	dPsiDy := (ne.FBM4DLoop(x, y+eps, phase, loopRadius, 3, 2.0, 0.5) -
		ne.FBM4DLoop(x, y-eps, phase, loopRadius, 3, 2.0, 0.5)) * inv2Eps
	dPsiDx := (ne.FBM4DLoop(x+eps, y, phase, loopRadius, 3, 2.0, 0.5) -
		ne.FBM4DLoop(x-eps, y, phase, loopRadius, 3, 2.0, 0.5)) * inv2Eps
	return Vector{X: dPsiDy, Y: -dPsiDx}
}

// ============================================================================
// Upgraded Gradient Perlin, Analytical Derivatives, Erosion FBM, Tileable,
// Multi-Stage Domain Warping, Flow Noise & Anisotropic Gabor Noise
// ============================================================================

// grad2Unit contains 12 unit-circle 2D gradient vectors for isotropic 2D Gradient Perlin noise.
var grad2Unit = [12][2]float64{
	{1.0, 0.0},
	{0.8660254037844386, 0.5},
	{0.5, 0.8660254037844386},
	{0.0, 1.0},
	{-0.5, 0.8660254037844386},
	{-0.8660254037844386, 0.5},
	{-1.0, 0.0},
	{-0.8660254037844386, -0.5},
	{-0.5, -0.8660254037844386},
	{0.0, -1.0},
	{0.5, -0.8660254037844386},
	{0.8660254037844386, -0.5},
}

// Perlin2D evaluates Ken Perlin's improved C2-continuous 2D gradient noise at (x, y) in [-1, 1].
func (ne *NoiseEngine) Perlin2D(x, y float64) float64 {
	v, _, _ := ne.Perlin2DDeriv(x, y)
	return v
}

// Perlin2DDeriv evaluates 2D C2-continuous gradient Perlin noise at (x, y) in [-1, 1]
// together with its exact analytical partial derivatives (dIdx, dIdy) in a single pass.
func (ne *NoiseEngine) Perlin2DDeriv(x, y float64) (val, dIdx, dIdy float64) {
	ix := fastFloor(x)
	iy := fastFloor(y)
	xf := x - float64(ix)
	yf := y - float64(iy)

	X0 := ix & 255
	Y0 := iy & 255
	X1 := (ix + 1) & 255
	Y1 := (iy + 1) & 255

	g00 := grad2Unit[ne.perm12[X0+int(ne.perm[Y0])]]
	g10 := grad2Unit[ne.perm12[X1+int(ne.perm[Y0])]]
	g01 := grad2Unit[ne.perm12[X0+int(ne.perm[Y1])]]
	g11 := grad2Unit[ne.perm12[X1+int(ne.perm[Y1])]]

	v00 := g00[0]*xf + g00[1]*yf
	v10 := g10[0]*(xf-1.0) + g10[1]*yf
	v01 := g01[0]*xf + g01[1]*(yf-1.0)
	v11 := g11[0]*(xf-1.0) + g11[1]*(yf-1.0)

	u := quinticFade(xf)
	v := quinticFade(yf)
	du := quinticFadeDeriv(xf)
	dv := quinticFadeDeriv(yf)

	k0 := v00
	k1 := v10 - v00
	k2 := v01 - v00
	k3 := v11 - v10 - v01 + v00

	const normScale = 1.4142135623730951
	val = Constrain(normScale*(k0+u*k1+v*k2+u*v*k3), -1.0, 1.0)

	gx := g00[0] + u*(g10[0]-g00[0]) + v*(g01[0]-g00[0]) + u*v*(g11[0]-g10[0]-g01[0]+g00[0])
	gy := g00[1] + u*(g10[1]-g00[1]) + v*(g01[1]-g00[1]) + u*v*(g11[1]-g10[1]-g01[1]+g00[1])

	dIdx = normScale * (gx + du*(k1+v*k3))
	dIdy = normScale * (gy + dv*(k2+u*k3))
	return val, dIdx, dIdy
}

// Perlin3D evaluates Ken Perlin's improved C2-continuous 3D gradient noise at (x, y, z) in [-1, 1].
func (ne *NoiseEngine) Perlin3D(x, y, z float64) float64 {
	ix := fastFloor(x)
	iy := fastFloor(y)
	iz := fastFloor(z)
	xf := x - float64(ix)
	yf := y - float64(iy)
	zf := z - float64(iz)

	X0, X1 := ix&255, (ix+1)&255
	Y0, Y1 := iy&255, (iy+1)&255
	Z0, Z1 := iz&255, (iz+1)&255

	u := quinticFade(xf)
	v := quinticFade(yf)
	w := quinticFade(zf)

	g000 := grad3[ne.perm12[X0+int(ne.perm[Y0+int(ne.perm[Z0])])]]
	g100 := grad3[ne.perm12[X1+int(ne.perm[Y0+int(ne.perm[Z0])])]]
	g010 := grad3[ne.perm12[X0+int(ne.perm[Y1+int(ne.perm[Z0])])]]
	g110 := grad3[ne.perm12[X1+int(ne.perm[Y1+int(ne.perm[Z0])])]]
	g001 := grad3[ne.perm12[X0+int(ne.perm[Y0+int(ne.perm[Z1])])]]
	g101 := grad3[ne.perm12[X1+int(ne.perm[Y0+int(ne.perm[Z1])])]]
	g011 := grad3[ne.perm12[X0+int(ne.perm[Y1+int(ne.perm[Z1])])]]
	g111 := grad3[ne.perm12[X1+int(ne.perm[Y1+int(ne.perm[Z1])])]]

	n000 := dot3(g000, xf, yf, zf)
	n100 := dot3(g100, xf-1.0, yf, zf)
	n010 := dot3(g010, xf, yf-1.0, zf)
	n110 := dot3(g110, xf-1.0, yf-1.0, zf)
	n001 := dot3(g001, xf, yf, zf-1.0)
	n101 := dot3(g101, xf-1.0, yf, zf-1.0)
	n011 := dot3(g011, xf, yf-1.0, zf-1.0)
	n111 := dot3(g111, xf-1.0, yf-1.0, zf-1.0)

	nx00 := n000 + u*(n100-n000)
	nx10 := n010 + u*(n110-n010)
	nx01 := n001 + u*(n101-n001)
	nx11 := n011 + u*(n111-n011)

	nxy0 := nx00 + v*(nx10-nx00)
	nxy1 := nx01 + v*(nx11-nx01)

	return Constrain(1.15470053838*(nxy0+w*(nxy1-nxy0)), -1.0, 1.0)
}

// FBM2DDeriv evaluates multi-octave 2D FBM along with its exact analytical partial derivatives (dIdx, dIdy).
func (ne *NoiseEngine) FBM2DDeriv(x, y float64, octaves int, lacunarity, gain float64) (val, dIdx, dIdy float64) {
	if octaves < 1 {
		octaves = 1
	}
	if lacunarity <= 0 {
		lacunarity = 2.0
	}
	if gain <= 0 {
		gain = 0.5
	}
	var sum, sumDx, sumDy, norm float64
	amp := 1.0
	freq := 1.0
	for i := 0; i < octaves; i++ {
		v, gx, gy := ne.Perlin2DDeriv(x*freq, y*freq)
		sum += amp * v
		sumDx += amp * freq * gx
		sumDy += amp * freq * gy
		norm += amp
		amp *= gain
		freq *= lacunarity
	}
	if norm == 0 {
		return 0, 0, 0
	}
	invNorm := 1.0 / norm
	return Constrain(sum*invNorm, -1.0, 1.0), sumDx * invNorm, sumDy * invNorm
}

// Curl2DAnalytical computes an exact divergence-free 2D velocity vector (dPsi/dy, -dPsi/dx)
// in a single pass using analytical FBM derivatives (3.5x faster than finite differences).
func (ne *NoiseEngine) Curl2DAnalytical(x, y float64, octaves int) Vector {
	if octaves < 1 {
		octaves = 3
	}
	_, dPsiDx, dPsiDy := ne.FBM2DDeriv(x, y, octaves, 2.0, 0.5)
	return Vector{X: dPsiDy, Y: -dPsiDx}
}

// ErosionFBM2D evaluates Inigo Quilez's gradient-damped erosion FBM:
// high-frequency octaves are attenuated by 1 / (1 + erosionStrength * ||sum(grad)||^2),
// carving smooth eroded gullies on steep slopes while retaining fine detail on flat plateaus.
func (ne *NoiseEngine) ErosionFBM2D(x, y float64, octaves int, lacunarity, gain, erosionStrength float64) (height, slopeX, slopeY float64) {
	if octaves < 1 {
		octaves = 4
	}
	if lacunarity <= 0 {
		lacunarity = 2.0
	}
	if gain <= 0 {
		gain = 0.5
	}
	if erosionStrength < 0 {
		erosionStrength = 1.0
	}

	var sum, norm, accDx, accDy float64
	amp := 1.0
	freq := 1.0
	// 2D rotation matrix (cos 37°, -sin 37°; sin 37°, cos 37°) decorrelates octave grid axes
	const rotC, rotS = 0.80, 0.60
	px, py := x, y

	for i := 0; i < octaves; i++ {
		v, gx, gy := ne.Perlin2DDeriv(px, py)
		// Map gy, gx back to world frame accumulate slope
		accDx += gx
		accDy += gy
		damp := 1.0 / (1.0 + erosionStrength*(accDx*accDx+accDy*accDy))
		sum += amp * (0.5 + 0.5*v) * damp
		norm += amp

		slopeX += amp * freq * gx * damp
		slopeY += amp * freq * gy * damp

		amp *= gain
		freq *= lacunarity
		nx := (rotC*px - rotS*py) * lacunarity
		ny := (rotS*px + rotC*py) * lacunarity
		px, py = nx, ny
	}
	if norm > 0 {
		height = Clamp01(sum / norm)
		slopeX /= norm
		slopeY /= norm
	}
	return height, slopeX, slopeY
}

func posModInt(a, m int) int {
	if m <= 0 {
		return a & 255
	}
	r := a % m
	if r < 0 {
		r += m
	}
	return r
}

// TileablePerlin2D evaluates 2D gradient Perlin noise that tiles seamlessly with period (periodX, periodY).
func (ne *NoiseEngine) TileablePerlin2D(x, y, periodX, periodY float64) float64 {
	px := int(math.Round(periodX))
	py := int(math.Round(periodY))
	if px <= 0 || py <= 0 {
		return ne.Perlin2D(x, y)
	}
	ix := fastFloor(x)
	iy := fastFloor(y)
	xf := x - float64(ix)
	yf := y - float64(iy)

	X0 := posModInt(ix, px) & 255
	Y0 := posModInt(iy, py) & 255
	X1 := posModInt(ix+1, px) & 255
	Y1 := posModInt(iy+1, py) & 255

	g00 := grad2Unit[ne.perm12[X0+int(ne.perm[Y0])]]
	g10 := grad2Unit[ne.perm12[X1+int(ne.perm[Y0])]]
	g01 := grad2Unit[ne.perm12[X0+int(ne.perm[Y1])]]
	g11 := grad2Unit[ne.perm12[X1+int(ne.perm[Y1])]]

	v00 := g00[0]*xf + g00[1]*yf
	v10 := g10[0]*(xf-1.0) + g10[1]*yf
	v01 := g01[0]*xf + g01[1]*(yf-1.0)
	v11 := g11[0]*(xf-1.0) + g11[1]*(yf-1.0)

	u := quinticFade(xf)
	v := quinticFade(yf)

	nx0 := v00 + u*(v10-v00)
	nx1 := v01 + u*(v11-v01)
	return Constrain(1.4142135623730951*(nx0+v*(nx1-nx0)), -1.0, 1.0)
}

// TileableFBM2D evaluates multi-octave 2D FBM that tiles seamlessly over [0, periodX) x [0, periodY).
func (ne *NoiseEngine) TileableFBM2D(x, y, periodX, periodY float64, octaves int, lacunarity, gain float64) float64 {
	if octaves < 1 {
		octaves = 1
	}
	if lacunarity <= 0 {
		lacunarity = 2.0
	}
	if gain <= 0 {
		gain = 0.5
	}
	var sum, norm float64
	amp := 1.0
	freq := 1.0
	for i := 0; i < octaves; i++ {
		sum += amp * ne.TileablePerlin2D(x*freq, y*freq, periodX*freq, periodY*freq)
		norm += amp
		amp *= gain
		freq *= lacunarity
	}
	if norm == 0 {
		return 0
	}
	return Constrain(sum/norm, -1.0, 1.0)
}

// DomainWarp2D performs Inigo Quilez's two-stage iterative domain warping:
// f(p + warpStrength * q(p + warpStrength * r(p))), returning both the final warped
// scalar value in [-1, 1] and the intermediate vector warp displacements (qx, qy, rx, ry)
// for rich multi-channel pigment coloring.
func (ne *NoiseEngine) DomainWarp2D(x, y, warpStrength float64, octaves int) (val, qx, qy, rx, ry float64) {
	if octaves < 1 {
		octaves = 4
	}
	qx = ne.FBM2D(x, y, octaves, 2.0, 0.5)
	qy = ne.FBM2D(x+5.2, y+1.3, octaves, 2.0, 0.5)

	rx = ne.FBM2D(x+warpStrength*qx+1.7, y+warpStrength*qy+9.2, octaves, 2.0, 0.5)
	ry = ne.FBM2D(x+warpStrength*qx+8.3, y+warpStrength*qy+2.8, octaves, 2.0, 0.5)

	val = ne.FBM2D(x+warpStrength*rx, y+warpStrength*ry, octaves, 2.0, 0.5)
	return val, qx, qy, rx, ry
}

// FlowNoise2D evaluates Perlin-Neyret rotating-gradient flow noise over spinPhase in [0, 1),
// where each lattice gradient vector rotates in-place at a deterministic integer harmonic rate
// so turbulence boils and swirls organically with exact C-infinity periodicity over spinPhase.
func (ne *NoiseEngine) FlowNoise2D(x, y, spinPhase float64, octaves int) float64 {
	if octaves < 1 {
		octaves = 3
	}
	basePhi := spinPhase * 2.0 * math.Pi
	var sum, norm float64
	amp := 1.0
	freq := 1.0

	evalSingleOctave := func(px, py, phi float64) float64 {
		ix := fastFloor(px)
		iy := fastFloor(py)
		xf := px - float64(ix)
		yf := py - float64(iy)
		X0, X1 := ix&255, (ix+1)&255
		Y0, Y1 := iy&255, (iy+1)&255

		cornerDot := func(cx, cy int, dx, dy float64) float64 {
			h := ne.perm[cx+int(ne.perm[cy])]
			baseAng := float64(h) * (2.0 * math.Pi / 256.0)
			spinRate := float64(1 + int(h&1))
			if (h & 2) != 0 {
				spinRate = -spinRate
			}
			ang := baseAng + spinRate*phi
			return math.Cos(ang)*dx + math.Sin(ang)*dy
		}

		v00 := cornerDot(X0, Y0, xf, yf)
		v10 := cornerDot(X1, Y0, xf-1.0, yf)
		v01 := cornerDot(X0, Y1, xf, yf-1.0)
		v11 := cornerDot(X1, Y1, xf-1.0, yf-1.0)

		u := quinticFade(xf)
		v := quinticFade(yf)
		nx0 := v00 + u*(v10-v00)
		nx1 := v01 + u*(v11-v01)
		return Constrain(1.4142135623730951*(nx0+v*(nx1-nx0)), -1.0, 1.0)
	}

	for i := 0; i < octaves; i++ {
		sum += amp * evalSingleOctave(x*freq, y*freq, basePhi*float64(i+1))
		norm += amp
		amp *= 0.5
		freq *= 2.0
	}
	if norm == 0 {
		return 0
	}
	return Constrain(sum/norm, -1.0, 1.0)
}

// GaborNoise2D evaluates Lagae et al.'s procedural anisotropic Gabor kernel noise at (x, y),
// convolving sparse Poisson-like impulses in a 3x3 cell neighborhood with directional harmonic
// Gaussian wavelets oriented along orientationRad with spatial frequency freq and Gaussian width bandwidth.
func (ne *NoiseEngine) GaborNoise2D(x, y, freq, orientationRad, bandwidth float64) float64 {
	if freq <= 0 {
		freq = 2.0
	}
	if bandwidth <= 0 {
		bandwidth = 1.5
	}
	ix := fastFloor(x)
	iy := fastFloor(y)

	cosO := math.Cos(orientationRad)
	sinO := math.Sin(orientationRad)
	gaussDecay := math.Pi * bandwidth * bandwidth
	omega := 2.0 * math.Pi * freq

	const impulsesPerCell = 4
	var sum float64

	for dy := -1; dy <= 1; dy++ {
		cy := iy + dy
		for dx := -1; dx <= 1; dx++ {
			cx := ix + dx
			cellHash := uint32(ne.perm[(cx&255)+int(ne.perm[cy&255])])
			for k := 0; k < impulsesPerCell; k++ {
				// Fast deterministic LCG per cell impulse
				h := (uint32(cx*73856093) ^ uint32(cy*19349663) ^ uint32(k*83492791) ^ (cellHash * 2654435761))
				h ^= h >> 16
				h *= 0x85ebca6b
				h ^= h >> 13
				h *= 0xc2b2ae35
				h ^= h >> 16

				ux := float64(h&0xFFFF) / 65535.0
				uy := float64((h>>16)&0xFFFF) / 65535.0
				weight := 1.0
				if (h & 0x10000) != 0 {
					weight = -1.0
				}

				rx := x - (float64(cx) + ux)
				ry := y - (float64(cy) + uy)
				r2 := rx*rx + ry*ry
				if r2 > 1.25 {
					continue
				}

				env := math.Exp(-gaussDecay * r2)
				proj := rx*cosO + ry*sinO
				sum += weight * env * math.Cos(omega*proj)
			}
		}
	}

	return Constrain(sum*0.48, -1.0, 1.0)
}

// ============================================================================
// Phase 3: Analytical 3D Noise Derivatives, Curl3D & Rich 2D/3D Voronoi Kernels
// ============================================================================

// Perlin3DDeriv evaluates 3D C2-continuous gradient Perlin noise at (x, y, z) in [-1, 1]
// together with its exact analytical partial derivatives (dIdx, dIdy, dIdz) in a single pass.
func (ne *NoiseEngine) Perlin3DDeriv(x, y, z float64) (val, dIdx, dIdy, dIdz float64) {
	ix := fastFloor(x)
	iy := fastFloor(y)
	iz := fastFloor(z)
	xf := x - float64(ix)
	yf := y - float64(iy)
	zf := z - float64(iz)

	X0, X1 := ix&255, (ix+1)&255
	Y0, Y1 := iy&255, (iy+1)&255
	Z0, Z1 := iz&255, (iz+1)&255

	u := quinticFade(xf)
	v := quinticFade(yf)
	w := quinticFade(zf)
	du := quinticFadeDeriv(xf)
	dv := quinticFadeDeriv(yf)
	dw := quinticFadeDeriv(zf)

	g000 := grad3[ne.perm12[X0+int(ne.perm[Y0+int(ne.perm[Z0])])]]
	g100 := grad3[ne.perm12[X1+int(ne.perm[Y0+int(ne.perm[Z0])])]]
	g010 := grad3[ne.perm12[X0+int(ne.perm[Y1+int(ne.perm[Z0])])]]
	g110 := grad3[ne.perm12[X1+int(ne.perm[Y1+int(ne.perm[Z0])])]]
	g001 := grad3[ne.perm12[X0+int(ne.perm[Y0+int(ne.perm[Z1])])]]
	g101 := grad3[ne.perm12[X1+int(ne.perm[Y0+int(ne.perm[Z1])])]]
	g011 := grad3[ne.perm12[X0+int(ne.perm[Y1+int(ne.perm[Z1])])]]
	g111 := grad3[ne.perm12[X1+int(ne.perm[Y1+int(ne.perm[Z1])])]]

	va := dot3(g000, xf, yf, zf)
	vb := dot3(g100, xf-1.0, yf, zf)
	vc := dot3(g010, xf, yf-1.0, zf)
	vd := dot3(g110, xf-1.0, yf-1.0, zf)
	ve := dot3(g001, xf, yf, zf-1.0)
	vf := dot3(g101, xf-1.0, yf, zf-1.0)
	vg := dot3(g011, xf, yf-1.0, zf-1.0)
	vh := dot3(g111, xf-1.0, yf-1.0, zf-1.0)

	k0 := va
	k1 := vb - va
	k2 := vc - va
	k3 := ve - va
	k4 := va - vb - vc + vd
	k5 := va - vc - ve + vg
	k6 := va - vb - ve + vf
	k7 := -va + vb + vc - vd + ve - vf - vg + vh

	const norm3 = 1.15470053838
	val = Constrain(norm3*(k0+u*k1+v*k2+w*k3+u*v*k4+v*w*k5+w*u*k6+u*v*w*k7), -1.0, 1.0)

	// Interpolate corner gradient vectors + product-rule quintic fade derivative terms
	interpGrad := func(idx int) float64 {
		ga := g000[idx]
		gb := g100[idx]
		gc := g010[idx]
		gd := g110[idx]
		ge := g001[idx]
		gf := g101[idx]
		gg := g011[idx]
		gh := g111[idx]
		return ga + u*(gb-ga) + v*(gc-ga) + w*(ge-ga) +
			u*v*(ga-gb-gc+gd) + v*w*(ga-gc-ge+gg) + w*u*(ga-gb-ge+gf) +
			u*v*w*(-ga+gb+gc-gd+ge-gf-gg+gh)
	}

	dIdx = norm3 * (interpGrad(0) + du*(k1+v*k4+w*k6+v*w*k7))
	dIdy = norm3 * (interpGrad(1) + dv*(k2+u*k4+w*k5+w*u*k7))
	dIdz = norm3 * (interpGrad(2) + dw*(k3+v*k5+u*k6+u*v*k7))
	return val, dIdx, dIdy, dIdz
}

// FBM3D evaluates multi-octave 3D Fractional Brownian Motion at (x, y, z) in [-1, 1].
func (ne *NoiseEngine) FBM3D(x, y, z float64, octaves int, lacunarity, gain float64) float64 {
	v, _, _, _ := ne.FBM3DDeriv(x, y, z, octaves, lacunarity, gain)
	return v
}

// FBM3DDeriv evaluates multi-octave 3D FBM along with its exact analytical partial derivatives (dIdx, dIdy, dIdz).
func (ne *NoiseEngine) FBM3DDeriv(x, y, z float64, octaves int, lacunarity, gain float64) (val, dIdx, dIdy, dIdz float64) {
	if octaves < 1 {
		octaves = 1
	}
	if lacunarity <= 0 {
		lacunarity = 2.0
	}
	if gain <= 0 {
		gain = 0.5
	}
	var sum, sumDx, sumDy, sumDz, norm float64
	amp := 1.0
	freq := 1.0
	for i := 0; i < octaves; i++ {
		v, gx, gy, gz := ne.Perlin3DDeriv(x*freq, y*freq, z*freq)
		sum += amp * v
		sumDx += amp * freq * gx
		sumDy += amp * freq * gy
		sumDz += amp * freq * gz
		norm += amp
		amp *= gain
		freq *= lacunarity
	}
	if norm == 0 {
		return 0, 0, 0, 0
	}
	invNorm := 1.0 / norm
	return Constrain(sum*invNorm, -1.0, 1.0), sumDx * invNorm, sumDy * invNorm, sumDz * invNorm
}

// Ridged3D evaluates multi-octave 3D ridged multifractal noise in [0, 1].
func (ne *NoiseEngine) Ridged3D(x, y, z float64, octaves int, lacunarity, gain float64) float64 {
	if octaves < 1 {
		octaves = 1
	}
	if lacunarity <= 0 {
		lacunarity = 2.0
	}
	if gain <= 0 {
		gain = 0.5
	}
	var sum, norm float64
	amp := 1.0
	freq := 1.0
	weight := 1.0
	for i := 0; i < octaves; i++ {
		n := 1.0 - math.Abs(ne.Perlin3D(x*freq, y*freq, z*freq))
		n = n * n * weight
		weight = Clamp01(n * 2.0)
		sum += n * amp
		norm += amp
		amp *= gain
		freq *= lacunarity
	}
	if norm == 0 {
		return 0
	}
	return Clamp01(sum / norm)
}

// Curl3D computes a divergence-free 3D velocity vector (vx, vy, vz) = curl(Psi)
// analytically from 3 decorrelated FBM3DDeriv vector-potential channels.
func (ne *NoiseEngine) Curl3D(x, y, z float64, octaves int) (vx, vy, vz float64) {
	if octaves < 1 {
		octaves = 3
	}
	_, _, dPsiXDy, dPsiXDz := ne.FBM3DDeriv(x, y, z, octaves, 2.0, 0.5)
	_, dPsiYDx, _, dPsiYDz := ne.FBM3DDeriv(x+31.416, y-47.853, z+12.793, octaves, 2.0, 0.5)
	_, dPsiZDx, dPsiZDy, _ := ne.FBM3DDeriv(x-92.653, y+19.324, z-63.185, octaves, 2.0, 0.5)

	vx = dPsiZDy - dPsiYDz
	vy = dPsiXDz - dPsiZDx
	vz = dPsiYDx - dPsiXDy
	return vx, vy, vz
}

// Voronoi2DResult holds rich 2D cellular Voronoi metrics including F1/F2 distances,
// Inigo Quilez's exact distance to the nearest cell border (EdgeDist), log-sum-exp
// smooth distance (SmoothF1), analytical gradient (GradX, GradY), and deterministic CellID.
type Voronoi2DResult struct {
	F1       float64
	F2       float64
	EdgeDist float64
	SmoothF1 float64
	GradX    float64
	GradY    float64
	DeltaX   float64
	DeltaY   float64
	CellID   int
}

// Voronoi2D evaluates a rich 2D cellular Voronoi field with jitter in [0, 1] and
// smooth-min falloff smoothness (e.g. 16.0; <=0 uses default 16.0).
func (ne *NoiseEngine) Voronoi2D(x, y, jitter, smoothness float64) Voronoi2DResult {
	jitter = Clamp01(jitter)
	if smoothness <= 0 {
		smoothness = 16.0
	}
	ix := fastFloor(x)
	iy := fastFloor(y)
	fx := x - float64(ix)
	fy := y - float64(iy)

	f1Sq := 1e9
	f2Sq := 1e9
	var bestRx, bestRy float64
	var bestID int
	var sumExp float64

	var offsets [9][2]float64
	idx := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			cx := (ix + dx) & 255
			cy := (iy + dy) & 255
			h1 := ne.perm[cx+int(ne.perm[cy])]
			h2 := ne.perm[cy+int(ne.perm[(cx+73)&255])]
			px := float64(dx) + 0.5 + jitter*(float64(h1)/255.0-0.5)
			py := float64(dy) + 0.5 + jitter*(float64(h2)/255.0-0.5)
			rx := px - fx
			ry := py - fy
			offsets[idx] = [2]float64{rx, ry}
			idx++

			dSq := rx*rx + ry*ry
			dist := math.Sqrt(dSq)
			sumExp += math.Exp(-smoothness * dist)

			if dSq < f1Sq {
				f2Sq = f1Sq
				f1Sq = dSq
				bestRx = rx
				bestRy = ry
				bestID = (int(h1) << 8) | int(h2)
			} else if dSq < f2Sq {
				f2Sq = dSq
			}
		}
	}

	// Pass 2: Inigo Quilez's exact distance to the nearest Voronoi bisector edge
	edgeDist := 1e9
	for i := 0; i < 9; i++ {
		rx, ry := offsets[i][0], offsets[i][1]
		diffX := rx - bestRx
		diffY := ry - bestRy
		diffLen := math.Hypot(diffX, diffY)
		if diffLen < 1e-5 {
			continue
		}
		midX := 0.5 * (bestRx + rx)
		midY := 0.5 * (bestRy + ry)
		dBorder := (midX*diffX + midY*diffY) / diffLen
		if dBorder < edgeDist {
			edgeDist = dBorder
		}
	}

	f1 := math.Sqrt(f1Sq)
	f2 := math.Sqrt(f2Sq)
	var gx, gy float64
	if f1 > 1e-7 {
		// Gradient of F1 = ||x - p_1|| points away from the nearest feature point: -(p_1 - x)/F1
		gx = -bestRx / f1
		gy = -bestRy / f1
	}
	smoothF1 := f1
	if sumExp > 0 {
		smoothF1 = -math.Log(sumExp) / smoothness
	}

	return Voronoi2DResult{
		F1:       f1,
		F2:       f2,
		EdgeDist: math.Max(0, edgeDist),
		SmoothF1: math.Max(0, smoothF1),
		GradX:    gx,
		GradY:    gy,
		DeltaX:   -bestRx,
		DeltaY:   -bestRy,
		CellID:   bestID,
	}
}

// Voronoi3DResult holds 3D cellular Voronoi distances (F1, F2), exact distance to the nearest
// 3D Voronoi cell face (EdgeDist), displacement Delta from the nearest feature point, and CellID.
type Voronoi3DResult struct {
	F1       float64
	F2       float64
	EdgeDist float64
	Delta    [3]float64
	CellID   int
}

// Voronoi3D evaluates a 3D cellular Voronoi field over a 3x3x3 lattice neighborhood.
func (ne *NoiseEngine) Voronoi3D(x, y, z, jitter float64) Voronoi3DResult {
	jitter = Clamp01(jitter)
	ix := fastFloor(x)
	iy := fastFloor(y)
	iz := fastFloor(z)
	fx := x - float64(ix)
	fy := y - float64(iy)
	fz := z - float64(iz)

	f1Sq := 1e9
	f2Sq := 1e9
	var bestR [3]float64
	var bestID int
	var offsets [27][3]float64
	idx := 0

	for dz := -1; dz <= 1; dz++ {
		cz := (iz + dz) & 255
		for dy := -1; dy <= 1; dy++ {
			cy := (iy + dy) & 255
			for dx := -1; dx <= 1; dx++ {
				cx := (ix + dx) & 255
				h1 := ne.perm[cx+int(ne.perm[cy+int(ne.perm[cz])])]
				h2 := ne.perm[cy+int(ne.perm[cz+int(ne.perm[(cx+59)&255])])]
				h3 := ne.perm[cz+int(ne.perm[cx+int(ne.perm[(cy+131)&255])])]

				px := float64(dx) + 0.5 + jitter*(float64(h1)/255.0-0.5)
				py := float64(dy) + 0.5 + jitter*(float64(h2)/255.0-0.5)
				pz := float64(dz) + 0.5 + jitter*(float64(h3)/255.0-0.5)

				rx := px - fx
				ry := py - fy
				rz := pz - fz
				offsets[idx] = [3]float64{rx, ry, rz}
				idx++

				dSq := rx*rx + ry*ry + rz*rz
				if dSq < f1Sq {
					f2Sq = f1Sq
					f1Sq = dSq
					bestR = [3]float64{rx, ry, rz}
					bestID = (int(h1) << 16) | (int(h2) << 8) | int(h3)
				} else if dSq < f2Sq {
					f2Sq = dSq
				}
			}
		}
	}

	edgeDist := 1e9
	for i := 0; i < 27; i++ {
		r := offsets[i]
		dx := r[0] - bestR[0]
		dy := r[1] - bestR[1]
		dz := r[2] - bestR[2]
		dLen := math.Sqrt(dx*dx + dy*dy + dz*dz)
		if dLen < 1e-5 {
			continue
		}
		mx := 0.5 * (bestR[0] + r[0])
		my := 0.5 * (bestR[1] + r[1])
		mz := 0.5 * (bestR[2] + r[2])
		dFace := (mx*dx + my*dy + mz*dz) / dLen
		if dFace < edgeDist {
			edgeDist = dFace
		}
	}

	return Voronoi3DResult{
		F1:       math.Sqrt(f1Sq),
		F2:       math.Sqrt(f2Sq),
		EdgeDist: math.Max(0, edgeDist),
		Delta:    [3]float64{-bestR[0], -bestR[1], -bestR[2]},
		CellID:   bestID,
	}
}



