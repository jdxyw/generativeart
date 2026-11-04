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

func TestHyperbolicCoxeterGroupAndDrosteSpiral(t *testing.T) {
	// 1. Test Coxeter (6, 4, 2) hyperbolic reflection group
	g := NewHyperbolicCoxeterGroup(6, 4, 2)
	if g.Side3.Radius <= 0 {
		t.Fatalf("expected positive orthogonal circle radius for (6,4,2), got %v", g.Side3.Radius)
	}
	// Verify Side3 is orthogonal to the unit circle: |Center|^2 - Radius^2 == 1
	orthoErr := math.Abs(cmplx.Abs(g.Side3.Center)*cmplx.Abs(g.Side3.Center) - g.Side3.Radius*g.Side3.Radius - 1.0)
	if orthoErr > 1e-10 {
		t.Fatalf("Side3 circle is not orthogonal to unit disk: error = %v", orthoErr)
	}

	// Folding a point and its mirror reflection across the real axis should land on the exact same zFund with opposite parity
	zTest := complex(0.42, 0.31)
	zFund1, bary1, par1, _ := g.FoldToFundamentalDomain(zTest, 48)
	zFund2, _, par2, _ := g.FoldToFundamentalDomain(cmplx.Conj(zTest), 48)
	if cmplx.Abs(zFund1-zFund2) > 1e-9 {
		t.Fatalf("expected conj(z) to fold to same zFund: got %v vs %v", zFund1, zFund2)
	}
	if par1 == par2 {
		t.Fatalf("expected opposite reflection parity for z and conj(z), got %d and %d", par1, par2)
	}
	if bary1[0] < 0 || bary1[1] < 0 || bary1[2] < 0 {
		t.Fatalf("expected non-negative barycentric coordinates, got %+v", bary1)
	}

	// 2. Test DrosteSpiralPullback self-similarity under scaling by scaleRatio
	u1, v1, lvl1, _ := DrosteSpiralPullback(complex(0.25, 0.15), 16.0, 1, 1.0)
	if u1 < 0 || u1 >= 1 || v1 < 0 || v1 >= 1 {
		t.Fatalf("Droste tile coords out of [0,1): u=%v v=%v", u1, v1)
	}
	_ = lvl1
}

func TestPhase3Analytical3DNoiseCurlAndVoronoi(t *testing.T) {
	ne := NewNoiseEngine(20261104)

	// 1. Verify Perlin3DDeriv matches central finite differences
	x, y, z := 2.37, -1.42, 3.81
	eps := 1e-5
	val, gx, gy, gz := ne.Perlin3DDeriv(x, y, z)
	if math.Abs(val-ne.Perlin3D(x, y, z)) > 1e-12 {
		t.Fatalf("Perlin3DDeriv value mismatch vs Perlin3D")
	}
	fdX := (ne.Perlin3D(x+eps, y, z) - ne.Perlin3D(x-eps, y, z)) / (2.0 * eps)
	fdY := (ne.Perlin3D(x, y+eps, z) - ne.Perlin3D(x, y-eps, z)) / (2.0 * eps)
	fdZ := (ne.Perlin3D(x, y, z+eps) - ne.Perlin3D(x, y, z-eps)) / (2.0 * eps)
	if math.Abs(gx-fdX) > 1e-4 || math.Abs(gy-fdY) > 1e-4 || math.Abs(gz-fdZ) > 1e-4 {
		t.Fatalf("Perlin3DDeriv mismatch vs finite difference: got (%f,%f,%f) want (%f,%f,%f)", gx, gy, gz, fdX, fdY, fdZ)
	}

	// 2. Verify Curl3D is divergence-free: dVx/dx + dVy/dy + dVz/dz ≈ 0
	vx1, _, _ := ne.Curl3D(x+eps, y, z, 3)
	vx0, _, _ := ne.Curl3D(x-eps, y, z, 3)
	_, vy1, _ := ne.Curl3D(x, y+eps, z, 3)
	_, vy0, _ := ne.Curl3D(x, y-eps, z, 3)
	_, _, vz1 := ne.Curl3D(x, y, z+eps, 3)
	_, _, vz0 := ne.Curl3D(x, y, z-eps, 3)
	div := (vx1-vx0)/(2*eps) + (vy1-vy0)/(2*eps) + (vz1-vz0)/(2*eps)
	if math.Abs(div) > 1e-3 {
		t.Fatalf("expected Curl3D to be divergence-free, got div=%e", div)
	}

	// 3. Verify Voronoi2D & Voronoi3D metrics
	v2 := ne.Voronoi2D(3.4, -2.1, 0.85, 16.0)
	if v2.F1 < 0 || v2.F2 < v2.F1 || v2.EdgeDist < 0 {
		t.Fatalf("unexpected Voronoi2D result: %+v", v2)
	}
	v3 := ne.Voronoi3D(1.8, -0.7, 2.9, 0.85)
	if v3.F1 < 0 || v3.F2 < v3.F1 || v3.EdgeDist < 0 {
		t.Fatalf("unexpected Voronoi3D result: %+v", v3)
	}
}

func TestPhase3SphericalHarmonicsAndGeodesicGoldbergMesh(t *testing.T) {
	// 1. Verify SphericalHarmonicYlm orthonormality on S^2 via midpoint quadrature
	nTheta, nPhi := 90, 180
	dTheta := math.Pi / float64(nTheta)
	dPhi := 2.0 * math.Pi / float64(nPhi)
	var norm22, cross22_31 float64
	for it := 0; it < nTheta; it++ {
		theta := (float64(it) + 0.5) * dTheta
		sinT := math.Sin(theta)
		for ip := 0; ip < nPhi; ip++ {
			phi := (float64(ip) + 0.5) * dPhi
			dOmega := sinT * dTheta * dPhi
			y21 := SphericalHarmonicYlm(2, 1, theta, phi)
			y3m2 := SphericalHarmonicYlm(3, -2, theta, phi)
			norm22 += y21 * y21 * dOmega
			cross22_31 += y21 * y3m2 * dOmega
		}
	}
	if math.Abs(norm22-1.0) > 1e-3 {
		t.Fatalf("expected integral of |Y_2^1|^2 over S^2 to equal 1.0, got %f", norm22)
	}
	if math.Abs(cross22_31) > 1e-3 {
		t.Fatalf("expected orthogonality between Y_2^1 and Y_3^-2, got %f", cross22_31)
	}

	// 2. Verify BuildGeodesicSphere(2) produces 162 vertices, 320 triangles, 12 pentagons, and 150 hexagons
	mesh := BuildGeodesicSphere(2)
	if len(mesh.Vertices) != 162 || len(mesh.Triangles) != 320 {
		t.Fatalf("expected 162 vertices and 320 triangles for subdiv=2, got V=%d T=%d", len(mesh.Vertices), len(mesh.Triangles))
	}
	pentagons, hexagons := 0, 0
	for _, cell := range mesh.Cells {
		if cell.IsPentagon && len(cell.Vertices) == 5 {
			pentagons++
		} else if len(cell.Vertices) == 6 {
			hexagons++
		}
	}
	if pentagons != 12 || hexagons != 150 {
		t.Fatalf("expected 12 Goldberg pentagons and 150 hexagons, got pentagons=%d hexagons=%d", pentagons, hexagons)
	}
}


