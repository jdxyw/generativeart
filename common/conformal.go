package common

import (
	"math"
	"math/cmplx"
)

// Mobius2D represents a 2D complex Möbius (fractional linear) transformation:
//
//	f(z) = (A*z + B) / (C*z + D)
type Mobius2D struct {
	A, B, C, D complex128
}

// NewMobius2D constructs a Mobius2D transformation with coefficients a, b, c, d.
func NewMobius2D(a, b, c, d complex128) Mobius2D {
	return Mobius2D{A: a, B: b, C: c, D: d}
}

// IdentityMobius2D returns the identity Möbius transformation f(z) = z.
func IdentityMobius2D() Mobius2D {
	return Mobius2D{A: 1, B: 0, C: 0, D: 1}
}

// Apply evaluates the Möbius transformation at z, guarding against division by zero
// when |C*z + D| < 1e-15.
func (m Mobius2D) Apply(z complex128) complex128 {
	if cmplx.IsInf(z) {
		if cmplx.Abs(m.C) < 1e-15 {
			return complex(1e15, 0)
		}
		return m.A / m.C
	}
	num := m.A*z + m.B
	den := m.C*z + m.D
	if cmplx.Abs(den) < 1e-15 {
		if cmplx.Abs(num) < 1e-15 {
			return 0
		}
		return num / complex(1e-15, 0)
	}
	return num / den
}

// Compose returns the composition m(other(z)) via 2x2 complex matrix multiplication.
func (m Mobius2D) Compose(other Mobius2D) Mobius2D {
	return Mobius2D{
		A: m.A*other.A + m.B*other.C,
		B: m.A*other.B + m.B*other.D,
		C: m.C*other.A + m.D*other.C,
		D: m.C*other.B + m.D*other.D,
	}
}

// Inverse returns the adjugate inverse Möbius transformation {D, -B, -C, A}.
func (m Mobius2D) Inverse() Mobius2D {
	return Mobius2D{
		A: m.D,
		B: -m.B,
		C: -m.C,
		D: m.A,
	}
}

// Det returns the complex determinant A*D - B*C.
func (m Mobius2D) Det() complex128 {
	return m.A*m.D - m.B*m.C
}

// Normalized scales A, B, C, D by 1/sqrt(Det()) so that Det() == 1 when |Det()| > 1e-15.
func (m Mobius2D) Normalized() Mobius2D {
	det := m.Det()
	if cmplx.Abs(det) <= 1e-15 {
		return m
	}
	s := cmplx.Sqrt(det)
	return Mobius2D{
		A: m.A / s,
		B: m.B / s,
		C: m.C / s,
		D: m.D / s,
	}
}

// IsometricCircle returns the center -D/C and radius 1/|C| of the isometric circle
// where |f'(z)| == 1 (for normalized transformations) when |C| > 1e-12.
func (m Mobius2D) IsometricCircle() (center complex128, radius float64, ok bool) {
	absC := cmplx.Abs(m.C)
	if absC <= 1e-12 {
		return 0, 0, false
	}
	return -m.D / m.C, 1.0 / absC, true
}

// DiskAutomorphism returns a Möbius transformation mapping the unit disk onto itself:
//
//	f(z) = exp(i*thetaRad) * (z - a) / (1 - conj(a)*z)
func DiskAutomorphism(a complex128, thetaRad float64) Mobius2D {
	rot := cmplx.Rect(1.0, thetaRad)
	return Mobius2D{
		A: rot,
		B: -rot * a,
		C: -cmplx.Conj(a),
		D: 1,
	}
}

// CayleyUpperHalfToDisk returns the Cayley transform mapping the upper half-plane Im(z) > 0
// onto the unit disk |w| < 1 via (z - i) / (z + i).
func CayleyUpperHalfToDisk() Mobius2D {
	return Mobius2D{
		A: 1,
		B: -1i,
		C: 1,
		D: 1i,
	}
}

// CayleyDiskToUpperHalf returns the inverse Cayley transform mapping the unit disk |z| < 1
// onto the upper half-plane Im(w) > 0 via i*(1 + z) / (1 - z).
func CayleyDiskToUpperHalf() Mobius2D {
	return Mobius2D{
		A: 1i,
		B: 1i,
		C: -1,
		D: 1,
	}
}

// Circle2D represents a circle in the complex plane.
type Circle2D struct {
	Center complex128
	Radius float64
}

// InvertPoint returns the inversion of z with respect to c:
//
//	c.Center + c.Radius^2 / conj(z - c.Center)
func (c Circle2D) InvertPoint(z complex128) complex128 {
	diff := z - c.Center
	if cmplx.Abs(diff) < 1e-15 {
		return c.Center + complex(1e15, 0)
	}
	r2 := complex(c.Radius*c.Radius, 0)
	return c.Center + r2/cmplx.Conj(diff)
}

// InvertCircle returns the exact circle inversion of other with respect to c.
func (c Circle2D) InvertCircle(other Circle2D) Circle2D {
	d := other.Center - c.Center
	dist2 := real(d)*real(d) + imag(d)*imag(d)
	r2 := other.Radius * other.Radius
	denom := dist2 - r2
	if math.Abs(denom) < 1e-12 {
		if denom >= 0 {
			denom = 1e-12
		} else {
			denom = -1e-12
		}
	}
	k := (c.Radius * c.Radius) / denom
	return Circle2D{
		Center: c.Center + d*complex(k, 0),
		Radius: math.Abs(k) * other.Radius,
	}
}

// SignedDist returns the signed Euclidean distance cmplx.Abs(z - c.Center) - c.Radius.
func (c Circle2D) SignedDist(z complex128) float64 {
	return cmplx.Abs(z-c.Center) - c.Radius
}

// PoincareGeodesicArc computes the exact hyperbolic geodesic arc inside the unit Poincaré disk
// between z1 and z2, sampled at steps+1 points.
func PoincareGeodesicArc(z1, z2 complex128, steps int) []Vector {
	if steps < 1 {
		steps = 1
	}
	pts := make([]Vector, steps+1)
	v1 := Vector{X: real(z1), Y: imag(z1)}
	v2 := Vector{X: real(z2), Y: imag(z2)}

	cross := v1.Cross(v2)
	if math.Abs(cross) < 1e-7 || cmplx.Abs(z1) < 1e-7 || cmplx.Abs(z2) < 1e-7 || cmplx.Abs(z1-z2) < 1e-9 {
		for i := 0; i <= steps; i++ {
			t := float64(i) / float64(steps)
			pts[i] = v1.Lerp(v2, t)
		}
		return pts
	}

	// Use the inverted point z1* = 1 / conj(z1) so the circle through z1, z2, z1*
	// is orthogonal to the unit circle.
	z1Star := 1.0 / cmplx.Conj(z1)
	ax, ay := real(z1), imag(z1)
	bx, by := real(z2), imag(z2)
	cx, cy := real(z1Star), imag(z1Star)

	d := 2.0 * (ax*(by-cy) + bx*(cy-ay) + cx*(ay-by))
	if math.Abs(d) < 1e-12 {
		for i := 0; i <= steps; i++ {
			t := float64(i) / float64(steps)
			pts[i] = v1.Lerp(v2, t)
		}
		return pts
	}

	a2 := ax*ax + ay*ay
	b2 := bx*bx + by*by
	c2 := cx*cx + cy*cy

	ux := (a2*(by-cy) + b2*(cy-ay) + c2*(ay-by)) / d
	uy := (a2*(cx-bx) + b2*(ax-cx) + c2*(bx-ax)) / d
	radius := math.Hypot(ax-ux, ay-uy)

	theta1 := math.Atan2(ay-uy, ax-ux)
	theta2 := math.Atan2(by-uy, bx-ux)
	dTheta := theta2 - theta1
	for dTheta > math.Pi {
		dTheta -= 2.0 * math.Pi
	}
	for dTheta < -math.Pi {
		dTheta += 2.0 * math.Pi
	}

	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		ang := theta1 + t*dTheta
		pts[i] = Vector{
			X: ux + radius*math.Cos(ang),
			Y: uy + radius*math.Sin(ang),
		}
	}
	return pts
}

// EvaluateDomainColoring computes domain-coloring channels for a complex value fz:
//   - argNorm in [0, 1): normalized phase (cmplx.Phase(fz) + pi) / (2*pi)
//   - logModIsoline in [0, 1]: smooth anti-aliased contour strength near integer multiples of logPeriod in ln(|fz|)
//   - phaseRayIsoline in [0, 1]: smooth anti-aliased contour strength along phaseRays equally spaced phase rays
//   - brightness in [0, 1]: stereographic Riemann-sphere lightness r^2/(r^2+1) modulated by log-modulus sawtooth shading
func EvaluateDomainColoring(fz complex128, logPeriod, phaseRays float64) (argNorm, logModIsoline, phaseRayIsoline, brightness float64) {
	phase := cmplx.Phase(fz)
	if math.IsNaN(phase) {
		phase = 0
	}
	argNorm = (phase + math.Pi) / (2.0 * math.Pi)
	argNorm = argNorm - math.Floor(argNorm)
	if argNorm >= 1.0 || argNorm < 0.0 {
		argNorm = 0.0
	}

	r := cmplx.Abs(fz)
	if math.IsNaN(r) {
		r = 0
	}

	if logPeriod <= 0 {
		logPeriod = 1.0
	}
	logR := math.Log(math.Max(r, 1e-30))
	u := logR / logPeriod
	fracLog := u - math.Floor(u)
	distLog := math.Min(fracLog, 1.0-fracLog)
	logModIsoline = 1.0 - Smoothstep(0.0, 0.08, distLog)

	if phaseRays > 0 {
		v := argNorm * phaseRays
		fracPhase := v - math.Floor(v)
		distPhase := math.Min(fracPhase, 1.0-fracPhase)
		phaseRayIsoline = 1.0 - Smoothstep(0.0, 0.08, distPhase)
	}

	var stereo float64
	if math.IsInf(r, 1) {
		stereo = 1.0
	} else {
		r2 := r * r
		stereo = r2 / (r2 + 1.0)
	}
	sawtooth := 0.7 + 0.3*fracLog
	brightness = Clamp01(stereo*sawtooth + 0.12*(1.0-stereo)*fracLog)
	return argNorm, logModIsoline, phaseRayIsoline, brightness
}
