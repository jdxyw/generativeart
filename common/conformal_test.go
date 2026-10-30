package common

import (
	"math"
	"math/cmplx"
	"testing"
)

func TestMobius2DOperations(t *testing.T) {
	id := IdentityMobius2D()
	z := complex(0.4, -0.3)
	if cmplx.Abs(id.Apply(z)-z) > 1e-12 {
		t.Fatalf("IdentityMobius2D.Apply(%v) = %v, want %v", z, id.Apply(z), z)
	}

	m1 := NewMobius2D(2, 1+1i, 1-1i, 3)
	m2 := NewMobius2D(1+0.5i, -0.5, 0.25i, 2)

	// Composition vs sequential application
	comp := m1.Compose(m2)
	got := comp.Apply(z)
	want := m1.Apply(m2.Apply(z))
	if cmplx.Abs(got-want) > 1e-12 {
		t.Fatalf("Compose.Apply(%v) = %v, want %v", z, got, want)
	}

	// Inverse round-trip
	inv := m1.Inverse()
	roundTrip := inv.Apply(m1.Apply(z))
	if cmplx.Abs(roundTrip-z) > 1e-12 {
		t.Fatalf("Inverse round-trip = %v, want %v", roundTrip, z)
	}

	// Normalized determinant == 1 and preserves mapping
	norm := m1.Normalized()
	if cmplx.Abs(norm.Det()-1.0) > 1e-12 {
		t.Fatalf("Normalized().Det() = %v, want 1", norm.Det())
	}
	if cmplx.Abs(norm.Apply(z)-m1.Apply(z)) > 1e-12 {
		t.Fatalf("Normalized().Apply(%v) = %v, want %v", z, norm.Apply(z), m1.Apply(z))
	}

	// Isometric circle of normalized transformation
	center, radius, ok := norm.IsometricCircle()
	if !ok {
		t.Fatalf("expected IsometricCircle ok = true")
	}
	expectedCenter := -norm.D / norm.C
	expectedRadius := 1.0 / cmplx.Abs(norm.C)
	if cmplx.Abs(center-expectedCenter) > 1e-12 || math.Abs(radius-expectedRadius) > 1e-12 {
		t.Fatalf("IsometricCircle = (%v, %v), want (%v, %v)", center, radius, expectedCenter, expectedRadius)
	}

	// Isometric circle when C == 0
	if _, _, ok := id.IsometricCircle(); ok {
		t.Fatalf("expected IsometricCircle ok = false when C == 0")
	}

	// Guard against near-zero denominator
	pole := -m1.D / m1.C
	poleVal := m1.Apply(pole)
	if cmplx.IsNaN(poleVal) {
		t.Fatalf("expected non-NaN value at pole, got %v", poleVal)
	}
}

func TestDiskAutomorphismAndCayley(t *testing.T) {
	a := complex(0.35, -0.25)
	auto := DiskAutomorphism(a, 0.7)

	// Maps a to 0
	if cmplx.Abs(auto.Apply(a)) > 1e-12 {
		t.Fatalf("DiskAutomorphism(a).Apply(a) = %v, want 0", auto.Apply(a))
	}

	// Maps unit circle to unit circle and interior to interior
	for k := 0; k < 16; k++ {
		ang := 2.0 * math.Pi * float64(k) / 16.0
		boundaryPt := cmplx.Rect(1.0, ang)
		mappedBoundary := auto.Apply(boundaryPt)
		if math.Abs(cmplx.Abs(mappedBoundary)-1.0) > 1e-12 {
			t.Fatalf("|auto(exp(i*theta))| = %v, want 1", cmplx.Abs(mappedBoundary))
		}
		interiorPt := cmplx.Rect(0.6, ang)
		if cmplx.Abs(auto.Apply(interiorPt)) >= 1.0 {
			t.Fatalf("expected interior point to remain inside unit disk")
		}
	}

	// Cayley transforms between upper half-plane and unit disk
	toDisk := CayleyUpperHalfToDisk()
	toUHP := CayleyDiskToUpperHalf()
	if cmplx.Abs(toDisk.Apply(1i)) > 1e-12 {
		t.Fatalf("CayleyUpperHalfToDisk(i) = %v, want 0", toDisk.Apply(1i))
	}
	uhpPt := complex(-1.2, 2.5)
	diskPt := toDisk.Apply(uhpPt)
	if cmplx.Abs(diskPt) >= 1.0 {
		t.Fatalf("expected upper half-plane point %v to map inside unit disk, got %v", uhpPt, diskPt)
	}
	recovered := toUHP.Apply(diskPt)
	if cmplx.Abs(recovered-uhpPt) > 1e-12 {
		t.Fatalf("Cayley round-trip = %v, want %v", recovered, uhpPt)
	}
}

func TestCircle2DInversion(t *testing.T) {
	c := Circle2D{Center: complex(1.0, -0.5), Radius: 2.0}
	p := complex(2.2, 0.4)

	pInv := c.InvertPoint(p)
	pBack := c.InvertPoint(pInv)
	if cmplx.Abs(pBack-p) > 1e-12 {
		t.Fatalf("InvertPoint involution failed: got %v, want %v", pBack, p)
	}

	// Point on circle stays fixed
	onCircle := c.Center + cmplx.Rect(c.Radius, 1.1)
	if math.Abs(c.SignedDist(onCircle)) > 1e-12 {
		t.Fatalf("SignedDist on circle = %v, want 0", c.SignedDist(onCircle))
	}
	if cmplx.Abs(c.InvertPoint(onCircle)-onCircle) > 1e-12 {
		t.Fatalf("point on circle should be fixed under InvertPoint")
	}

	// InvertCircle consistency with sampled points on other circle
	other := Circle2D{Center: complex(4.0, 1.0), Radius: 1.2}
	invCircle := c.InvertCircle(other)
	for k := 0; k < 12; k++ {
		ang := 2.0 * math.Pi * float64(k) / 12.0
		sample := other.Center + cmplx.Rect(other.Radius, ang)
		invertedSample := c.InvertPoint(sample)
		if math.Abs(invCircle.SignedDist(invertedSample)) > 1e-10 {
			t.Fatalf("inverted point %v does not lie on inverted circle (dist = %v)", invertedSample, invCircle.SignedDist(invertedSample))
		}
	}
}

func TestPoincareGeodesicArc(t *testing.T) {
	// Non-collinear points inside the unit disk
	z1 := complex(0.5, 0.1)
	z2 := complex(-0.1, 0.6)
	steps := 32
	arc := PoincareGeodesicArc(z1, z2, steps)
	if len(arc) != steps+1 {
		t.Fatalf("len(arc) = %d, want %d", len(arc), steps+1)
	}
	if math.Hypot(arc[0].X-real(z1), arc[0].Y-imag(z1)) > 1e-12 {
		t.Fatalf("arc start = %v, want %v", arc[0], z1)
	}
	if math.Hypot(arc[steps].X-real(z2), arc[steps].Y-imag(z2)) > 1e-12 {
		t.Fatalf("arc end = %v, want %v", arc[steps], z2)
	}
	for i, pt := range arc {
		if pt.Len() >= 1.0 {
			t.Fatalf("arc point %d (%v) escaped unit disk", i, pt)
		}
	}

	// Collinear with origin (diameter segment)
	d1 := complex(-0.6, -0.3)
	d2 := complex(0.4, 0.2)
	diamArc := PoincareGeodesicArc(d1, d2, 10)
	if len(diamArc) != 11 {
		t.Fatalf("len(diamArc) = %d, want 11", len(diamArc))
	}
	mid := diamArc[5]
	wantMid := Vector{X: -0.1, Y: -0.05}
	if mid.Dist(wantMid) > 1e-12 {
		t.Fatalf("collinear midpoint = %v, want %v", mid, wantMid)
	}
}

func TestEvaluateDomainColoring(t *testing.T) {
	// At |fz| == 1 (ln|fz| == 0) and phase == 0 (argNorm == 0.5)
	argNorm, logIso, phaseIso, brightness := EvaluateDomainColoring(complex(1.0, 0.0), 1.0, 8.0)
	if math.Abs(argNorm-0.5) > 1e-12 {
		t.Fatalf("argNorm(1+0i) = %v, want 0.5", argNorm)
	}
	if logIso < 0.99 {
		t.Fatalf("expected strong logModIsoline at |z|=1, got %v", logIso)
	}
	if phaseIso < 0.99 {
		t.Fatalf("expected strong phaseRayIsoline at phase=0 with 8 rays, got %v", phaseIso)
	}
	if brightness <= 0 || brightness > 1 {
		t.Fatalf("brightness out of range: %v", brightness)
	}

	// Mid-period and mid-ray point should have low contour strength
	fzMid := cmplx.Rect(math.Exp(0.5), math.Pi/8.0)
	arg2, logIso2, phaseIso2, bright2 := EvaluateDomainColoring(fzMid, 1.0, 8.0)
	if arg2 < 0 || arg2 >= 1 {
		t.Fatalf("argNorm out of [0, 1): %v", arg2)
	}
	if logIso2 > 0.05 {
		t.Fatalf("expected low logModIsoline at half period, got %v", logIso2)
	}
	if phaseIso2 > 0.05 {
		t.Fatalf("expected low phaseRayIsoline between rays, got %v", phaseIso2)
	}
	if bright2 < 0 || bright2 > 1 {
		t.Fatalf("brightness out of [0, 1]: %v", bright2)
	}
}
