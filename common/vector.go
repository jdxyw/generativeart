package common

import "math"

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

func (v *Vector) Normalize() {
	l := Magnitude(v.X, v.Y)
	if l == 0 {
		return
	}
	v.Multiple(1 / l)
}

func (v *Vector) SetMag(m float64) {
	v.Normalize()
	v.Multiple(m)
}

// Add returns the vector sum v + o.
func (v Vector) Add(o Vector) Vector {
	return Vector{X: v.X + o.X, Y: v.Y + o.Y}
}

// Sub returns the vector difference v - o.
func (v Vector) Sub(o Vector) Vector {
	return Vector{X: v.X - o.X, Y: v.Y - o.Y}
}

// Scale returns the vector scaled by scalar s.
func (v Vector) Scale(s float64) Vector {
	return Vector{X: v.X * s, Y: v.Y * s}
}

// Dot returns the 2D dot product of v and o.
func (v Vector) Dot(o Vector) float64 {
	return v.X*o.X + v.Y*o.Y
}

// Cross returns the 2D scalar cross product (z-component of 3D cross product).
func (v Vector) Cross(o Vector) float64 {
	return v.X*o.Y - v.Y*o.X
}

// Len returns the Euclidean length of v.
func (v Vector) Len() float64 {
	return math.Hypot(v.X, v.Y)
}

// Dist returns the Euclidean distance between v and o.
func (v Vector) Dist(o Vector) float64 {
	return math.Hypot(v.X-o.X, v.Y-o.Y)
}

// Heading returns the angle of v in radians in [-pi, pi].
func (v Vector) Heading() float64 {
	return math.Atan2(v.Y, v.X)
}

// Rotate returns v rotated counter-clockwise by angle (in radians).
func (v Vector) Rotate(angle float64) Vector {
	s, c := math.Sincos(angle)
	return Vector{
		X: v.X*c - v.Y*s,
		Y: v.X*s + v.Y*c,
	}
}

// Perp returns a perpendicular vector (-Y, X) rotated 90 degrees counter-clockwise.
func (v Vector) Perp() Vector {
	return Vector{X: -v.Y, Y: v.X}
}

// Normalized returns a unit-length copy of v, or the zero vector if v has zero length.
func (v Vector) Normalized() Vector {
	l := v.Len()
	if l == 0 {
		return Vector{}
	}
	return Vector{X: v.X / l, Y: v.Y / l}
}

// Lerp linearly interpolates between v and o by t.
func (v Vector) Lerp(o Vector, t float64) Vector {
	return Vector{
		X: v.X + (o.X-v.X)*t,
		Y: v.Y + (o.Y-v.Y)*t,
	}
}

// CatmullRomPoint evaluates a uniform Catmull-Rom spline segment between p1 and p2 at parameter t in [0, 1].
func CatmullRomPoint(p0, p1, p2, p3 Vector, t float64) Vector {
	t2 := t * t
	t3 := t2 * t
	x := 0.5 * ((2.0 * p1.X) +
		(-p0.X+p2.X)*t +
		(2.0*p0.X-5.0*p1.X+4.0*p2.X-p3.X)*t2 +
		(-p0.X+3.0*p1.X-3.0*p2.X+p3.X)*t3)
	y := 0.5 * ((2.0 * p1.Y) +
		(-p0.Y+p2.Y)*t +
		(2.0*p0.Y-5.0*p1.Y+4.0*p2.Y-p3.Y)*t2 +
		(-p0.Y+3.0*p1.Y-3.0*p2.Y+p3.Y)*t3)
	return Vector{X: x, Y: y}
}

// CatmullRomSpline generates a smooth spline passing through all pts with the given subdivisions per segment.
func CatmullRomSpline(pts []Vector, subdivisions int) []Vector {
	n := len(pts)
	if n < 2 {
		cp := make([]Vector, n)
		copy(cp, pts)
		return cp
	}
	if subdivisions < 1 {
		subdivisions = 1
	}
	out := make([]Vector, 0, (n-1)*subdivisions+1)
	for i := 0; i < n-1; i++ {
		p0 := pts[MaxInt(i-1, 0)]
		p1 := pts[i]
		p2 := pts[i+1]
		p3 := pts[MinInt(i+2, n-1)]
		for s := 0; s < subdivisions; s++ {
			t := float64(s) / float64(subdivisions)
			out = append(out, CatmullRomPoint(p0, p1, p2, p3, t))
		}
	}
	out = append(out, pts[n-1])
	return out
}

// ChaikinSmooth applies Chaikin's corner-cutting subdivision algorithm for the specified number of iterations.
func ChaikinSmooth(pts []Vector, iterations int, closed bool) []Vector {
	if len(pts) < 3 || iterations <= 0 {
		cp := make([]Vector, len(pts))
		copy(cp, pts)
		return cp
	}
	curr := make([]Vector, len(pts))
	copy(curr, pts)

	for iter := 0; iter < iterations; iter++ {
		n := len(curr)
		if closed {
			next := make([]Vector, 0, n*2)
			for i := 0; i < n; i++ {
				p0 := curr[i]
				p1 := curr[(i+1)%n]
				next = append(next, p0.Lerp(p1, 0.25), p0.Lerp(p1, 0.75))
			}
			curr = next
		} else {
			next := make([]Vector, 0, (n-1)*2+2)
			next = append(next, curr[0])
			for i := 0; i < n-1; i++ {
				p0 := curr[i]
				p1 := curr[i+1]
				next = append(next, p0.Lerp(p1, 0.25), p0.Lerp(p1, 0.75))
			}
			next = append(next, curr[n-1])
			curr = next
		}
	}
	return curr
}
