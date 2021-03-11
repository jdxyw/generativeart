package common

import (
	"image/color"
	"math"
)

// Constrain returns a value between a minimum and maximum value.
func Constrain(val, low, high float64) float64 {
	return math.Max(math.Min(val, high), low)
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
	r := uint8(float64(c1.R)*ratio + float64(c2.R)*(1-ratio))
	g := uint8(float64(c1.G)*ratio + float64(c2.G)*(1-ratio))
	b := uint8(float64(c1.B)*ratio + float64(c2.B)*(1-ratio))
	a := uint8(float64(c1.A)*ratio + float64(c2.A)*(1-ratio))

	return color.RGBA{r, g, b, a}
}
