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
	if mh <= 0 {
		mh = 360
	}
	if ms <= 0 {
		ms = 100
	}
	if mv <= 0 {
		mv = 100
	}

	hVal := hs.H % mh
	if hVal < 0 {
		hVal += mh
	}
	sVal := ConstrainInt(hs.S, 0, ms)
	vVal := ConstrainInt(hs.V, 0, mv)

	h, s, v := float64(hVal)/float64(mh), float64(sVal)/float64(ms), float64(vVal)/float64(mv)

	var r, g, b float64
	if s == 0 { //HSV from 0 to 1
		r = v * 255
		g = v * 255
		b = v * 255
	} else {
		h = h * 6
		if h >= 6 {
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
		R: uint8(ConstrainInt(int(math.Round(r)), 0, 255)),
		G: uint8(ConstrainInt(int(math.Round(g)), 0, 255)),
		B: uint8(ConstrainInt(int(math.Round(b)), 0, 255)),
		A: 255,
	}
	return rgb
}

// HSVFromFloat converts floating-point HSV (h in [0, 360] or [0, 1], s in [0, 1], v in [0, 1]) to color.RGBA.
func HSVFromFloat(h, s, v float64) color.RGBA {
	if h > 1.0 || h < 0.0 {
		h = math.Mod(h, 360.0)
		if h < 0 {
			h += 360.0
		}
		h /= 360.0
	}
	s = Clamp01(s)
	v = Clamp01(v)

	var r, g, b float64
	if s == 0 {
		r, g, b = v, v, v
	} else {
		h6 := h * 6.0
		if h6 >= 6.0 {
			h6 = 0.0
		}
		i := math.Floor(h6)
		f := h6 - i
		p := v * (1.0 - s)
		q := v * (1.0 - s*f)
		t := v * (1.0 - s*(1.0-f))
		switch int(i) {
		case 0:
			r, g, b = v, t, p
		case 1:
			r, g, b = q, v, p
		case 2:
			r, g, b = p, v, t
		case 3:
			r, g, b = p, q, v
		case 4:
			r, g, b = t, p, v
		default:
			r, g, b = v, p, q
		}
	}
	return color.RGBA{
		R: uint8(ConstrainInt(int(math.Round(r*255.0)), 0, 255)),
		G: uint8(ConstrainInt(int(math.Round(g*255.0)), 0, 255)),
		B: uint8(ConstrainInt(int(math.Round(b*255.0)), 0, 255)),
		A: 255,
	}
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
	theta := math.Atan2(y, x)

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
