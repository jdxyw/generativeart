package common

import (
	"image/color"
	"math"
)

// Constrain returns a value between a minimum and maximum value.
func Constrain(val, low, high float64) float64 {
	return math.Max(math.Min(val, high), low)
}

// ConstrainInt returns a int value between a minimum and maximum value.
func ConstrainInt(val, low, high int) int {
	return MaxInt(MinInt(val, high), low)
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Remap re-maps a number from one range to another.
func Remap(n, start1, stop1, start2, stop2 float64) float64 {
	newval := (n-start1)/(stop1-start1)*(stop2-start2) + start2

	if start2 < stop2 {
		return Constrain(newval, start2, stop2)
	} else {
		return Constrain(newval, stop2, start2)
	}
}

// LerpColor blends two colors to find a third color somewhere between them.
func LerpColor(c1, c2 color.RGBA, ratio float64) color.RGBA {
	ratio = Constrain(ratio, 0.0, 1.0)
	r := uint8(float64(c1.R)*ratio + float64(c2.R)*(1-ratio))
	g := uint8(float64(c1.G)*ratio + float64(c2.G)*(1-ratio))
	b := uint8(float64(c1.B)*ratio + float64(c2.B)*(1-ratio))
	a := uint8(float64(c1.A)*ratio + float64(c2.A)*(1-ratio))

	return color.RGBA{r, g, b, a}
}

func Magnitude(x, y float64) float64 {
	return math.Sqrt(x*x + y*y)
}

// Clamp01 clamps x to the [0, 1] interval.
func Clamp01(x float64) float64 {
	return Constrain(x, 0.0, 1.0)
}

// Lerp performs linear interpolation between a and b by t.
func Lerp(a, b, t float64) float64 {
	return a + t*(b-a)
}

// Smoothstep performs Hermite interpolation between 0 and 1 when edge0 < x < edge1.
func Smoothstep(edge0, edge1, x float64) float64 {
	if edge0 == edge1 {
		if x < edge0 {
			return 0.0
		}
		return 1.0
	}
	t := Clamp01((x - edge0) / (edge1 - edge0))
	return t * t * (3.0 - 2.0*t)
}

// Smootherstep performs Ken Perlin's quintic interpolation between 0 and 1 when edge0 < x < edge1.
func Smootherstep(edge0, edge1, x float64) float64 {
	if edge0 == edge1 {
		if x < edge0 {
			return 0.0
		}
		return 1.0
	}
	t := Clamp01((x - edge0) / (edge1 - edge0))
	return t * t * t * (t*(t*6.0-15.0) + 10.0)
}

// Gaussian evaluates the unnormalized Gaussian function exp(-0.5 * ((x - mean) / sigma)^2).
func Gaussian(x, mean, sigma float64) float64 {
	if sigma == 0 {
		if x == mean {
			return 1.0
		}
		return 0.0
	}
	d := (x - mean) / sigma
	return math.Exp(-0.5 * d * d)
}
