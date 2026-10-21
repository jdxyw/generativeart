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

