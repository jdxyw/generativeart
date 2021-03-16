package common

type Vector struct {
	X, Y float64
}

func NewVector(x, y float64) *Vector {
	return &Vector{
		X: x,
		Y: y,
	}
}

func (v *Vector) Multiple(z float64) {
	v.X = v.X * z
	v.Y = v.Y * z
}
