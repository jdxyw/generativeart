package common

type Rect struct {
	x, y, w, h float64
}

func (r Rect) X() float64 {
	return r.x
}

func (r Rect) Y() float64 {
	return r.y
}

func (r Rect) W() float64 {
	return r.w
}

func (r Rect) H() float64 {
	return r.h
}

func NewRect(x, y, w, h float64) *Rect {
	return &Rect{
		x: x,
		y: y,
		w: w,
		h: h,
	}
}
