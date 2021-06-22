package common

import (
	"image/color"
	"math"
	"math/rand"
)

type HSV struct {
	H, S, V int
}

// ToRGB converts a HSV color mode to RGB mode
// mh, ms, mv are used to set the maximum number for HSV.
func (hs HSV) ToRGB(mh, ms, mv int) color.RGBA {
	if hs.H > mh {
		hs.H = mh
	}

	if hs.S > ms {
		hs.S = ms
	}

	if hs.V > mv {
		hs.V = mv
	}

	h, s, v := float64(hs.H)/float64(mh), float64(hs.S)/float64(ms), float64(hs.V)/float64(mv)

	var r, g, b float64
	if s == 0 { //HSV from 0 to 1
		r = v * 255
		g = v * 255
		b = v * 255
	} else {
		h = h * 6
		if h == 6 {
			h = 0
		} //H must be < 1
		i := math.Floor(h) //Or ... var_i = floor( var_h )
		v1 := v * (1 - s)
		v2 := v * (1 - s*(h-i))
		v3 := v * (1 - s*(1-(h-i)))

		if i == 0 {
			r = v
			g = v3
			b = v1
		} else if i == 1 {
			r = v2
			g = v
			b = v1
		} else if i == 2 {
			r = v1
			g = v
			b = v3
		} else if i == 3 {
			r = v1
			g = v2
			b = v
		} else if i == 4 {
			r = v3
			g = v1
			b = v
		} else {
			r = v
			g = v1
			b = v2
		}

		r = r * 255 //RGB results from 0 to 255
		g = g * 255
		b = b * 255
	}
	rgb := color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: 0,
	}
	return rgb
}

// ConvertCartesianToPixel converts cartesian coordinates to actual pixels in an image.
func ConvertCartesianToPixel(x, y, xaixs, yaixs float64, h, w int) (int, int) {
	xr, yr := x/xaixs, y/yaixs
	var i, j int
	i = w/2 + int(float64(w)/2*xr)
	j = h/2 + int(float64(h)/2*yr)

	return i, j
}

func ConvertCartesianToPolarPixel(x, y, xaixs, yaixs float64, h, w int) (int, int) {
	r, theta := ConvertCartesianToPolar(x, y)
	return ConvertPolarToPixel(r, theta, xaixs, yaixs, h, w)
}

// ConvertCartesianToPolar converts points from cartesian coordinates to polar coordinates.
func ConvertCartesianToPolar(x, y float64) (float64, float64) {
	r := math.Sqrt(x*x + y*y)
	theta := math.Atanh(y / x)

	return r, theta
}

// ConvertPolarToPixel converts polar coordinates to actual pixels in an image.
func ConvertPolarToPixel(r, theta, xaixs, yaixs float64, h, w int) (int, int) {
	x, y := r*math.Cos(theta), r*math.Sin(theta)

	xr, yr := x/xaixs, y/yaixs
	var i, j int
	i = w/2 + int(float64(w)/2*xr)
	j = h/2 + int(float64(h)/2*yr)

	return i, j
}

// Distance returns the Euclidean distance between two point on 2D dimension.
func Distance(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt(math.Pow(x1-x2, 2.0) + math.Pow(y1-y2, 2.0))
}

// RandomRangeInt returns a integer number between min and max.
func RandomRangeInt(min, max int) int {
	return rand.Intn(max-min) + min
}

// RandomRangeFloat64 returns a float64 number between min and max.
func RandomRangeFloat64(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// RandomGaussian returns a gaussian random float64 number.
func RandomGaussian(mean, std float64) float64 {
	return rand.NormFloat64()*std + mean
}
