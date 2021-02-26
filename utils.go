package generativeart

import "math"

func ConvertCartesianToPixel(x, y, xaixs, yaixs float64, h, w int) (int, int) {
	xr, yr := x/xaixs, y/yaixs
	var i, j int
	if xr > 0 {
		i = w/2 + int(float64(w)/2*xr)
	} else {
		i = w/2 + int(float64(w)/2*xr)
	}

	if yr > 0 {
		j = h/2 + int(float64(h)/2*yr)
	} else {
		j = h/2 + int(float64(h)/2*yr)
	}

	return i, j
}

func ConvertCartesianToPolarPixel(x, y, xaixs, yaixs float64, h, w int) (int, int) {
	r, theta := ConvertCartesianToPolar(x, y)
	return ConvertPolarToPixel(r, theta, xaixs, yaixs, h, w)
}

func ConvertCartesianToPolar(x, y float64) (float64, float64) {
	r := math.Sqrt(x*x + y*y)
	theta := math.Atanh(y / x)

	return r, theta
}

func ConvertPolarToPixel(r, theta, xaixs, yaixs float64, h, w int) (int, int) {
	x, y := r*math.Cos(theta), r*math.Sin(theta)

	xr, yr := x/xaixs, y/yaixs
	var i, j int
	if xr > 0 {
		i = w/2 + int(float64(w)/2*xr)
	} else {
		i = w/2 + int(float64(w)/2*xr)
	}

	if yr > 0 {
		j = h/2 + int(float64(h)/2*yr)
	} else {
		j = h/2 + int(float64(h)/2*yr)
	}

	return i, j
}
