package common

import "math"

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
