package common

import (
	"math"
	"sort"
)

// NormalizedAssociatedLegendre evaluates the fully orthonormalized associated Legendre
// polynomial \bar{P}_l^m(x) for l >= 0, 0 <= m <= l, and x = cos(theta) in [-1, 1]
// such that \int_{-1}^{1} [\bar{P}_l^m(x)]^2 dx = 1 / (2*pi).
func NormalizedAssociatedLegendre(l, m int, x float64) float64 {
	if l < 0 || m < 0 || m > l {
		return 0
	}
	x = Constrain(x, -1.0, 1.0)
	sinThetaSq := math.Max(0.0, 1.0-x*x)

	// 1. Compute \bar{P}_m^m(x) via stable product recurrence
	pmm := 1.0 / math.Sqrt(4.0*math.Pi)
	if m > 0 {
		sinTheta := math.Sqrt(sinThetaSq)
		for i := 1; i <= m; i++ {
			fi := float64(i)
			pmm *= -math.Sqrt((2.0*fi+1.0)/(2.0*fi)) * sinTheta
		}
	}
	if l == m {
		return pmm
	}

	// 2. Compute \bar{P}_{m+1}^m(x)
	fm := float64(m)
	pmmp1 := x * math.Sqrt(2.0*fm+3.0) * pmm
	if l == m+1 {
		return pmmp1
	}

	// 3. Upward degree recurrence in l
	var pll float64
	for ll := m + 2; ll <= l; ll++ {
		fl := float64(ll)
		a := math.Sqrt((4.0*fl*fl - 1.0) / (fl*fl - fm*fm))
		b := math.Sqrt(((2.0*fl + 1.0) * ((fl - 1.0)*(fl - 1.0) - fm*fm)) / ((2.0*fl - 3.0) * (fl*fl - fm*fm)))
		pll = a*x*pmmp1 - b*pmm
		pmm = pmmp1
		pmmp1 = pll
	}
	return pll
}

// SphericalHarmonicYlm evaluates the real orthonormal spherical harmonic Y_l^m(theta, phi)
// for polar colatitude theta in [0, pi] and azimuth phi in [0, 2*pi), where l >= 0 and -l <= m <= l.
// Satisfies \int_{S^2} Y_l^m(\omega) Y_{l'}^{m'}(\omega) d\omega = \delta_{ll'} \delta_{mm'}.
func SphericalHarmonicYlm(l, m int, theta, phi float64) float64 {
	if l < 0 || m < -l || m > l {
		return 0
	}
	x := math.Cos(theta)
	if m == 0 {
		return NormalizedAssociatedLegendre(l, 0, x)
	}
	absM := m
	if absM < 0 {
		absM = -absM
	}
	plm := NormalizedAssociatedLegendre(l, absM, x)
	const sqrt2 = 1.4142135623730951
	if m > 0 {
		return sqrt2 * plm * math.Cos(float64(absM)*phi)
	}
	return sqrt2 * plm * math.Sin(float64(absM)*phi)
}

// SphericalHarmonicCartesian evaluates the real orthonormal spherical harmonic Y_l^m
// directly at a 3D unit direction vector (nx, ny, nz) on S^2.
func SphericalHarmonicCartesian(l, m int, nx, ny, nz float64) float64 {
	r := math.Sqrt(nx*nx + ny*ny + nz*nz)
	if r < 1e-12 {
		return SphericalHarmonicYlm(l, m, 0, 0)
	}
	ux, uy, uz := nx/r, ny/r, Constrain(nz/r, -1.0, 1.0)
	theta := math.Acos(uz)
	phi := math.Atan2(uy, ux)
	return SphericalHarmonicYlm(l, m, theta, phi)
}

// GeodesicTriangle3D represents a spherical triangle on a subdivided GeodesicMesh3D.
type GeodesicTriangle3D struct {
	A, B, C int
	Center  [3]float64
}

// GoldbergCell3D represents a dual hexagonal (or 12 pentagonal) face on a Goldberg polyhedron.
type GoldbergCell3D struct {
	VertexIndex int
	Center      [3]float64
	Vertices    [][3]float64
	IsPentagon  bool
}

// GeodesicMesh3D holds a subdivided icosahedral geodesic sphere and its dual Goldberg
// hexagonal/pentagonal polyhedron mesh on S^2.
type GeodesicMesh3D struct {
	Vertices  [][3]float64
	Triangles []GeodesicTriangle3D
	Cells     []GoldbergCell3D
}

func norm3D(v [3]float64) [3]float64 {
	l := math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
	if l < 1e-12 {
		return [3]float64{0, 0, 1}
	}
	return [3]float64{v[0] / l, v[1] / l, v[2] / l}
}

// BuildGeodesicSphere constructs a Class-I subdivided icosahedral geodesic sphere
// (subdivisions in [0..5]) and its exact dual Goldberg polyhedron (12 regular pentagons
// at the icosahedral vertices + hexagonal tiles across all subdivided faces).
func BuildGeodesicSphere(subdivisions int) *GeodesicMesh3D {
	if subdivisions < 0 {
		subdivisions = 0
	}
	if subdivisions > 5 {
		subdivisions = 5
	}

	const phi = 1.618033988749895
	rawVerts := [12][3]float64{
		{-1, phi, 0}, {1, phi, 0}, {-1, -phi, 0}, {1, -phi, 0},
		{0, -1, phi}, {0, 1, phi}, {0, -1, -phi}, {0, 1, -phi},
		{phi, 0, -1}, {phi, 0, 1}, {-phi, 0, -1}, {-phi, 0, 1},
	}
	verts := make([][3]float64, 12)
	for i, rv := range rawVerts {
		verts[i] = norm3D(rv)
	}

	tris := [][3]int{
		{0, 11, 5}, {0, 5, 1}, {0, 1, 7}, {0, 7, 10}, {0, 10, 11},
		{1, 5, 9}, {5, 11, 4}, {11, 10, 2}, {10, 7, 6}, {7, 1, 8},
		{3, 9, 4}, {3, 4, 2}, {3, 2, 6}, {3, 6, 8}, {3, 8, 9},
		{4, 9, 5}, {2, 4, 11}, {6, 2, 10}, {8, 6, 7}, {9, 8, 1},
	}

	type edgePair struct{ u, v int }
	for s := 0; s < subdivisions; s++ {
		midCache := make(map[edgePair]int, len(tris)*2)
		getMid := func(u, v int) int {
			a, b := u, v
			if a > b {
				a, b = b, a
			}
			key := edgePair{a, b}
			if idx, ok := midCache[key]; ok {
				return idx
			}
			va, vb := verts[u], verts[v]
			m := norm3D([3]float64{va[0] + vb[0], va[1] + vb[1], va[2] + vb[2]})
			idx := len(verts)
			verts = append(verts, m)
			midCache[key] = idx
			return idx
		}

		nextTris := make([][3]int, 0, len(tris)*4)
		for _, t := range tris {
			a, b, c := t[0], t[1], t[2]
			ab := getMid(a, b)
			bc := getMid(b, c)
			ca := getMid(c, a)
			nextTris = append(nextTris,
				[3]int{a, ab, ca},
				[3]int{b, bc, ab},
				[3]int{c, ca, bc},
				[3]int{ab, bc, ca},
			)
		}
		tris = nextTris
	}

	mesh := &GeodesicMesh3D{
		Vertices:  verts,
		Triangles: make([]GeodesicTriangle3D, len(tris)),
		Cells:     make([]GoldbergCell3D, len(verts)),
	}

	incTris := make([][]int, len(verts))
	for tIdx, t := range tris {
		va, vb, vc := verts[t[0]], verts[t[1]], verts[t[2]]
		cent := norm3D([3]float64{va[0] + vb[0] + vc[0], va[1] + vb[1] + vc[1], va[2] + vb[2] + vc[2]})
		mesh.Triangles[tIdx] = GeodesicTriangle3D{
			A:      t[0],
			B:      t[1],
			C:      t[2],
			Center: cent,
		}
		incTris[t[0]] = append(incTris[t[0]], tIdx)
		incTris[t[1]] = append(incTris[t[1]], tIdx)
		incTris[t[2]] = append(incTris[t[2]], tIdx)
	}

	// Build dual Goldberg cells by sorting incident triangle centroids counter-clockwise around each vertex normal
	for vIdx, n := range verts {
		inc := incTris[vIdx]
		// Orthonormal tangent basis (t1, t2) on the tangent plane at n
		ref := [3]float64{0, 1, 0}
		if math.Abs(n[1]) > 0.9 {
			ref = [3]float64{1, 0, 0}
		}
		// t1 = normalize(ref x n)
		t1 := norm3D([3]float64{
			ref[1]*n[2] - ref[2]*n[1],
			ref[2]*n[0] - ref[0]*n[2],
			ref[0]*n[1] - ref[1]*n[0],
		})
		// t2 = n x t1
		t2 := [3]float64{
			n[1]*t1[2] - n[2]*t1[1],
			n[2]*t1[0] - n[0]*t1[2],
			n[0]*t1[1] - n[1]*t1[0],
		}

		type angleVert struct {
			ang float64
			pt  [3]float64
		}
		avs := make([]angleVert, len(inc))
		for i, tIdx := range inc {
			c := mesh.Triangles[tIdx].Center
			dx, dy, dz := c[0]-n[0], c[1]-n[1], c[2]-n[2]
			u := dx*t1[0] + dy*t1[1] + dz*t1[2]
			v := dx*t2[0] + dy*t2[1] + dz*t2[2]
			avs[i] = angleVert{ang: math.Atan2(v, u), pt: c}
		}
		sort.Slice(avs, func(i, j int) bool {
			return avs[i].ang < avs[j].ang
		})
		poly := make([][3]float64, len(avs))
		for i, av := range avs {
			poly[i] = av.pt
		}
		mesh.Cells[vIdx] = GoldbergCell3D{
			VertexIndex: vIdx,
			Center:      n,
			Vertices:    poly,
			IsPentagon:  len(poly) == 5,
		}
	}

	return mesh
}
