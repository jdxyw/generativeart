package render

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart/common"
)

// Vec3 is a 3D Euclidean vector used by the 3D Signed Distance Field (SDF3D)
// raymarching, symmetry-folding, and curvature engine.
type Vec3 struct {
	X, Y, Z float64
}

// V3 constructs a Vec3(x, y, z).
func V3(x, y, z float64) Vec3 {
	return Vec3{X: x, Y: y, Z: z}
}

func (a Vec3) Add(b Vec3) Vec3      { return Vec3{a.X + b.X, a.Y + b.Y, a.Z + b.Z} }
func (a Vec3) Sub(b Vec3) Vec3      { return Vec3{a.X - b.X, a.Y - b.Y, a.Z - b.Z} }
func (a Vec3) Scale(s float64) Vec3 { return Vec3{a.X * s, a.Y * s, a.Z * s} }
func (a Vec3) Dot(b Vec3) float64   { return a.X*b.X + a.Y*b.Y + a.Z*b.Z }
func (a Vec3) Len() float64         { return math.Sqrt(a.X*a.X + a.Y*a.Y + a.Z*a.Z) }

func (a Vec3) Cross(b Vec3) Vec3 {
	return Vec3{
		X: a.Y*b.Z - a.Z*b.Y,
		Y: a.Z*b.X - a.X*b.Z,
		Z: a.X*b.Y - a.Y*b.X,
	}
}

func (a Vec3) Normalize() Vec3 {
	l := a.Len()
	if l < 1e-12 {
		return Vec3{0, 0, 1}
	}
	inv := 1.0 / l
	return Vec3{a.X * inv, a.Y * inv, a.Z * inv}
}

func (a Vec3) Abs() Vec3 {
	return Vec3{math.Abs(a.X), math.Abs(a.Y), math.Abs(a.Z)}
}

// RotateX rotates v around the X axis by angleRad.
func (a Vec3) RotateX(angleRad float64) Vec3 {
	c, s := math.Cos(angleRad), math.Sin(angleRad)
	return Vec3{a.X, a.Y*c - a.Z*s, a.Y*s + a.Z*c}
}

// RotateY rotates v around the Y axis by angleRad.
func (a Vec3) RotateY(angleRad float64) Vec3 {
	c, s := math.Cos(angleRad), math.Sin(angleRad)
	return Vec3{a.X*c + a.Z*s, a.Y, -a.X*s + a.Z*c}
}

// RotateZ rotates v around the Z axis by angleRad.
func (a Vec3) RotateZ(angleRad float64) Vec3 {
	c, s := math.Cos(angleRad), math.Sin(angleRad)
	return Vec3{a.X*c - a.Y*s, a.X*s + a.Y*c, a.Z}
}

// SDF3D is a 3D Signed Distance Field function returning signed distance from p.
type SDF3D func(p Vec3) float64

// SDF3DHit records the result of raymarching an SDF3D scene.
type SDF3DHit struct {
	Hit       bool
	T         float64
	Pos       Vec3
	Normal    Vec3
	Curvature float64 // Mean surface curvature in [-1, 1] (>0 convex ridge, <0 concave valley)
	AO        float64 // 5-tap 3D ambient occlusion in [0, 1]
	Shadow    float64 // Soft penumbra direct shadow factor in [0, 1]
}

// =============================================================================
// 1. 3D SDF Analytical Primitives & Architectural Building Blocks
// =============================================================================

func SDF3DSphere(p Vec3, radius float64) float64 {
	return p.Len() - radius
}

func SDF3DBox(p, halfExtents Vec3) float64 {
	q := p.Abs().Sub(halfExtents)
	out := Vec3{math.Max(q.X, 0), math.Max(q.Y, 0), math.Max(q.Z, 0)}.Len()
	in := math.Min(math.Max(q.X, math.Max(q.Y, q.Z)), 0.0)
	return out + in
}

func SDF3DRoundedBox(p, halfExtents Vec3, bevelRadius float64) float64 {
	return SDF3DBox(p, halfExtents) - bevelRadius
}

func SDF3DTorus(p Vec3, majorR, minorR float64) float64 {
	qx := math.Hypot(p.X, p.Z) - majorR
	return math.Hypot(qx, p.Y) - minorR
}

func SDF3DCylinderY(p Vec3, radius, halfHeight float64) float64 {
	dx := math.Hypot(p.X, p.Z) - radius
	dy := math.Abs(p.Y) - halfHeight
	return math.Min(math.Max(dx, dy), 0.0) + math.Hypot(math.Max(dx, 0.0), math.Max(dy, 0.0))
}

func SDF3DCapsule(p, a, b Vec3, radius float64) float64 {
	pa := p.Sub(a)
	ba := b.Sub(a)
	h := common.Clamp01(pa.Dot(ba) / math.Max(ba.Dot(ba), 1e-12))
	return pa.Sub(ba.Scale(h)).Len() - radius
}

func SDF3DOctahedron(p Vec3, size float64) float64 {
	q := p.Abs()
	m := q.X + q.Y + q.Z - size
	var r Vec3
	if 3.0*q.X < m {
		r = q
	} else if 3.0*q.Y < m {
		r = Vec3{q.Y, q.Z, q.X}
	} else if 3.0*q.Z < m {
		r = Vec3{q.Z, q.X, q.Y}
	} else {
		return m * 0.57735027
	}
	k := common.Clamp01(0.5 * (r.Z - r.Y + size))
	return Vec3{r.X, r.Y - size + k, r.Z - k}.Len()
}

// SDF3DFlutedColumnY evaluates a classical Doric/Corinthian fluted column along Y
// with entasis taper, nFlutes concave vertical flutes, and top/bottom molded plinths.
func SDF3DFlutedColumnY(p Vec3, radius, halfHeight float64, nFlutes int) float64 {
	ang := math.Atan2(p.Z, p.X)
	flute := 0.0
	if nFlutes > 0 && math.Abs(p.Y) < halfHeight*0.86 {
		flute = 0.055 * radius * (0.5 + 0.5*math.Cos(float64(nFlutes)*ang))
	}
	shaft := SDF3DCylinderY(p, radius-flute, halfHeight)
	// Capital & base plinth moldings
	capY := math.Abs(p.Y) - halfHeight*0.90
	capital := SDF3DRoundedBox(Vec3{p.X, capY, p.Z}, Vec3{radius * 1.24, halfHeight * 0.10, radius * 1.24}, radius*0.04)
	return math.Min(shaft, capital)
}

// SDF3DHexPrismY evaluates an exact 3D regular hexagonal prism centered on the Y axis
// with flat-to-flat inradius and vertical half-height halfH.
func SDF3DHexPrismY(p Vec3, inradius, halfH float64) float64 {
	kx := -0.8660254037844386
	kz := 0.5
	kt := 0.5773502691896258
	qx := math.Abs(p.X)
	qz := math.Abs(p.Z)
	dot := kx*qx + kz*qz
	if dot < 0 {
		qx -= 2.0 * dot * kx
		qz -= 2.0 * dot * kz
	}
	clampedX := common.Constrain(qx, -kt*inradius, kt*inradius)
	dHex := math.Hypot(qx-clampedX, qz-inradius)
	if qz-inradius < 0 {
		dHex = -dHex
	}
	dy := math.Abs(p.Y) - halfH
	return math.Min(math.Max(dHex, dy), 0.0) + math.Hypot(math.Max(dHex, 0.0), math.Max(dy, 0.0))
}

// SDF3DPlatonicCage evaluates a hollow Renaissance polyhedral strut cage (Leonardo /
// Jamnitzer / Kepler style) of radius r and strut thickness strutR:
//   - kind == 0: Hollow Dodecahedron Cage (12 pentagonal windows, 30 struts, 5-fold Y axis)
//   - kind == 1: Hollow Icosahedron Cage (20 triangular windows, 30 struts, 5-fold Y axis)
//   - kind == 2: Hollow Octahedron Cage (8 triangular windows, 12 struts, 4-fold Y axis)
//   - kind == 3: Hollow Cube / Hexahedron Cage (6 square windows, 12 struts, 4-fold Y axis)
//   - kind == 4: Hollow Star-Tetrahedron / Stella Octangula Cage (4-fold Y axis)
func SDF3DPlatonicCage(p Vec3, kind int, r, strutR float64) float64 {
	wallT := strutR * 0.90
	switch kind {
	case 0:
		// Hollow Dodecahedron / Icosahedron Cage (5-fold Y axis):
		// In FoldIcosahedral, +Y is the face normal, and (X, Z) are tangential on the face!
		q := FoldIcosahedral(AlignIcosahedral5FoldY(p))
		dSolid := SDF3DRoundedBox(q, Vec3{r * 0.65, r, r * 0.65}, strutR*0.45)
		dHollow := math.Max(dSolid, -SDF3DSphere(p, r*0.80))
		dWindow := math.Hypot(q.X, q.Z) - r*0.25
		return SDF3DSmoothSubtract(dWindow, dHollow, strutR*0.30)
	case 1:
		// Dual Slotted Icosahedral Cage (5-fold Y axis) with triangular aperture cutouts
		q := FoldIcosahedral(AlignIcosahedral5FoldY(p))
		dSolid := SDF3DRoundedBox(q, Vec3{r * 0.62, r, r * 0.62}, strutR*0.40)
		dHollow := math.Max(dSolid, -SDF3DSphere(p, r*0.84))
		dWindow := math.Max(q.X-strutR*1.1, q.Z-strutR*1.1)
		return math.Max(dHollow, -dWindow)
	case 2:
		// Hollow Octahedron Cage (4-fold Y axis):
		// Bound by SDF3DOctahedron(q, r), hollowed inside, with triangular window cutouts on each of the 8 faces!
		q := FoldOctahedral(p)
		dOct := SDF3DOctahedron(q, r)
		dShell := math.Max(dOct, -(dOct + wallT*1.8))
		// In sorted FoldOctahedral (x >= y >= z >= 0), the 3 edges of the octahedral face are z=0 and x+y+z=r;
		// cutting where z > strutR*1.25 carves a clean triangular window on every octahedral face!
		dWindow := q.Z - strutR*1.25
		return math.Max(dShell, -dWindow)
	case 3:
		// Hollow Cube / Hexahedron Cage (4-fold Y axis):
		// Bound by SDF3DRoundedBox(q, (r,r,r)), hollowed inside, with open square windows on all 6 faces!
		q := FoldOctahedral(p)
		dCube := SDF3DRoundedBox(q, Vec3{r, r, r}, strutR*0.45)
		dShell := math.Max(dCube, -SDF3DBox(q, Vec3{r - wallT*1.6, r - wallT*1.6, r - wallT*1.6}))
		// In sorted FoldOctahedral (x >= y >= z >= 0), the face is x = r and the face coordinates are (y, z);
		// dWindow < 0 where max(y, z) < r - strutR*2.2 carves open square windows on all 6 cube faces!
		dWindow := math.Max(q.Y, q.Z) - (r - strutR*2.2)
		return SDF3DSmoothSubtract(dWindow, dShell, strutR*0.25)
	default:
		// Star-Octahedron / Keplerian Spiked Core (4-fold Y axis)
		q := FoldOctahedral(p)
		dStar := SDF3DOctahedron(q, r)
		dSpikes := SDF3DCapsule(q, Vec3{0, 0, 0}, Vec3{r * 1.25, 0, 0}, strutR*1.15)
		return SDF3DSmoothUnion(dStar, dSpikes, strutR*0.6)
	}
}

// SDF3DClebschCalabiYau evaluates a bounded, smooth 3D algebraic-trigonometric
// Calabi-Yau / Clebsch diagonal surface manifold with octahedral + 5-fold harmonic
// saddles, hollowed genus-tunnel handles, and thin shell thickness shellT.
func SDF3DClebschCalabiYau(p Vec3, radius, morphPhase, shellT float64) float64 {
	if radius <= 1e-4 {
		radius = 1.0
	}
	u := p.Scale(1.0 / radius)
	r := u.Len()
	if r > 1.45 {
		return (r - 1.15) * radius
	}
	// Smooth octahedral + quintic Calabi-Yau trigonometric cross-section implicit field:
	// Combines Clebsch cubic symmetry (x^3 + y^3 + z^3 - 3xyz) with 5-fold harmonic saddles
	k := 3.141592653589793 * 1.35
	cx, sx := math.Cos(u.X*k+morphPhase), math.Sin(u.X*k+morphPhase)
	cy, sy := math.Cos(u.Y*k+morphPhase), math.Sin(u.Y*k+morphPhase)
	cz, sz := math.Cos(u.Z*k+morphPhase), math.Sin(u.Z*k+morphPhase)

	gyroid := sx*cy + sy*cz + sz*cx
	schwarzP := cx + cy + cz
	clebsch := (u.X*u.X*u.X + u.Y*u.Y*u.Y + u.Z*u.Z*u.Z - 3.0*u.X*u.Y*u.Z) * 1.45

	iso := 0.52*gyroid + 0.34*schwarzP + 0.42*clebsch
	// Lipschitz-bounded shell distance intersected smoothly with a bounding sphere
	dSurface := (math.Abs(iso) / (k * 1.45)) * radius - shellT
	dSphere := (r - 1.0) * radius
	return SDF3DSmoothIntersect(dSurface, dSphere, 0.08*radius)
}

// SDF3DRomanArchZ evaluates a 3D vaulted Romanesque/Piranesi stone arch spanning X
// with thickness along Z, opening width archRadius, and pier height pierH.
func SDF3DRomanArchZ(p Vec3, outerW, totalH, depthZ, archRadius, pierH float64) float64 {
	block := SDF3DBox(p, Vec3{outerW, totalH, depthZ})
	// Semicircular vault + rectangular lower opening cutout
	archCenterY := pierH
	var cutout float64
	if p.Y >= archCenterY {
		cutout = math.Hypot(p.X, p.Y-archCenterY) - archRadius
	} else {
		cutout = math.Abs(p.X) - archRadius
	}
	return math.Max(block, -cutout)
}

// =============================================================================
// 2. Smooth Boolean CSG & Polyhedral Symmetry Folding Operators
// =============================================================================

func SDF3DSmoothUnion(d1, d2, k float64) float64 {
	if k <= 1e-6 {
		return math.Min(d1, d2)
	}
	h := common.Clamp01(0.5 + 0.5*(d2-d1)/k)
	return d2*(1.0-h) + d1*h - k*h*(1.0-h)
}

func SDF3DSmoothSubtract(d1, d2, k float64) float64 {
	if k <= 1e-6 {
		return math.Max(-d1, d2)
	}
	h := common.Clamp01(0.5 - 0.5*(d2+d1)/k)
	return d2*(1.0-h) + (-d1)*h + k*h*(1.0-h)
}

func SDF3DSmoothIntersect(d1, d2, k float64) float64 {
	if k <= 1e-6 {
		return math.Max(d1, d2)
	}
	h := common.Clamp01(0.5 - 0.5*(d2-d1)/k)
	return d2*(1.0-h) + d1*h + k*h*(1.0-h)
}

// FoldPolarModXZ folds p around the Y axis into the fundamental sector of an N-fold
// dihedral rosette in O(1) time, returning the folded point and sector index.
func FoldPolarModXZ(p Vec3, sectors int) (Vec3, int) {
	if sectors <= 1 {
		return p, 0
	}
	r := math.Hypot(p.X, p.Z)
	ang := math.Atan2(p.Z, p.X)
	sectorAngle := 2.0 * math.Pi / float64(sectors)
	idx := int(math.Floor((ang + math.Pi) / sectorAngle))
	foldedAng := math.Mod(ang+math.Pi, sectorAngle)
	if foldedAng < 0 {
		foldedAng += sectorAngle
	}
	foldedAng -= 0.5 * sectorAngle
	return Vec3{r * math.Cos(foldedAng), p.Y, r * math.Sin(foldedAng)}, idx
}

// FoldOctahedral folds p into the fundamental domain of the 48-element octahedral
// symmetry group Oh (sorting |x| >= |y| >= |z|).
func FoldOctahedral(p Vec3) Vec3 {
	q := p.Abs()
	if q.X < q.Y {
		q.X, q.Y = q.Y, q.X
	}
	if q.Y < q.Z {
		q.Y, q.Z = q.Z, q.Y
	}
	if q.X < q.Y {
		q.X, q.Y = q.Y, q.X
	}
	return q
}

// FoldIcosahedral folds p into the fundamental domain of the 120-element icosahedral /
// dodecahedral Coxeter reflection group H3 (mirrors x=0, z=0, and nc=(-1/2, 1/(2*phi), -phi/2),
// meeting at dihedral angles pi/2, pi/3, and pi/5) in 5 reflections.
func FoldIcosahedral(p Vec3) Vec3 {
	const phi = 1.618033988749895
	nc := Vec3{-0.5, 0.5 / phi, -0.5 * phi}
	q := p
	for i := 0; i < 6; i++ {
		q.X = math.Abs(q.X)
		q.Z = math.Abs(q.Z)
		d := q.Dot(nc)
		if d < 0 {
			q = q.Sub(nc.Scale(2.0 * d))
		}
	}
	q.X = math.Abs(q.X)
	q.Z = math.Abs(q.Z)
	return q
}

// AlignIcosahedral5FoldY rotates p around the Z axis by -atan(1/phi) (~-31.717 deg)
// so that the vertical Y axis coincides with the exact 5-fold rotational symmetry axis
// (1, phi, 0) of FoldIcosahedral. Rotating p around Y by multiples of 2*pi/5 before calling
// AlignIcosahedral5FoldY + FoldIcosahedral is 100% invariant (zero loop seam).
func AlignIcosahedral5FoldY(p Vec3) Vec3 {
	const alpha5 = 0.5535743588970452 // math.Atan(1.0 / 1.618033988749895)
	return p.RotateZ(-alpha5)
}

// EvaluateMoebiusStoneRelief3D evaluates object-locked 3D triplanar Voronoi stone &
// metal relief in the rotating object's local coordinate frame (pLocal, nLocal),
// eliminating 2D "shower-door" texture swimming during 3D animations.
// Uses smooth power-weighted triplanar blending across 45-degree bevel transitions
// while retaining a fast 1-plane path on cardinal-aligned faces.
func EvaluateMoebiusStoneRelief3D(ne *common.NoiseEngine, pLocal, nLocal Vec3, freq, warpStrength float64) MoebiusStoneRelief {
	if ne == nil {
		return MoebiusStoneRelief{}
	}
	ax := math.Abs(nLocal.X)
	ay := math.Abs(nLocal.Y)
	az := math.Abs(nLocal.Z)

	// Fast path when one normal component strongly dominates (e.g. > 0.78)
	if ax >= ay && ax >= az && ax > 0.78 && ax-ay > 0.22 && ax-az > 0.22 {
		return EvaluateMoebiusStoneRelief(ne, pLocal.Y+1.7, pLocal.Z-1.7, freq, freq, warpStrength)
	}
	if ay >= ax && ay >= az && ay > 0.78 && ay-ax > 0.22 && ay-az > 0.22 {
		return EvaluateMoebiusStoneRelief(ne, pLocal.X+3.4, pLocal.Z-3.4, freq, freq, warpStrength)
	}
	if az >= ax && az >= ay && az > 0.78 && az-ax > 0.22 && az-ay > 0.22 {
		return EvaluateMoebiusStoneRelief(ne, pLocal.X, pLocal.Y, freq, freq, warpStrength)
	}

	// Smooth 4th-power triplanar weights across diagonal/bevel transitions
	wx := math.Max(0.0, ax-0.12)
	wy := math.Max(0.0, ay-0.12)
	wz := math.Max(0.0, az-0.12)
	wx = wx * wx * wx * wx
	wy = wy * wy * wy * wy
	wz = wz * wz * wz * wz
	wSum := wx + wy + wz
	if wSum < 1e-8 {
		return EvaluateMoebiusStoneRelief(ne, pLocal.X, pLocal.Y, freq, freq, warpStrength)
	}
	wx /= wSum
	wy /= wSum
	wz /= wSum

	var out MoebiusStoneRelief
	bestW := -1.0
	if wx > 0.05 {
		rx := EvaluateMoebiusStoneRelief(ne, pLocal.Y+1.7, pLocal.Z-1.7, freq, freq, warpStrength)
		out.CrackInk += wx * rx.CrackInk
		out.PockInk += wx * rx.PockInk
		out.ColorMod += wx * rx.ColorMod
		out.NormalMod[0] += wx * rx.NormalMod[0]
		out.NormalMod[1] += wx * rx.NormalMod[1]
		if wx > bestW {
			bestW = wx
			out.BlockID = rx.BlockID
		}
	}
	if wy > 0.05 {
		ry := EvaluateMoebiusStoneRelief(ne, pLocal.X+3.4, pLocal.Z-3.4, freq, freq, warpStrength)
		out.CrackInk += wy * ry.CrackInk
		out.PockInk += wy * ry.PockInk
		out.ColorMod += wy * ry.ColorMod
		out.NormalMod[0] += wy * ry.NormalMod[0]
		out.NormalMod[1] += wy * ry.NormalMod[1]
		if wy > bestW {
			bestW = wy
			out.BlockID = ry.BlockID
		}
	}
	if wz > 0.05 {
		rz := EvaluateMoebiusStoneRelief(ne, pLocal.X, pLocal.Y, freq, freq, warpStrength)
		out.CrackInk += wz * rz.CrackInk
		out.PockInk += wz * rz.PockInk
		out.ColorMod += wz * rz.ColorMod
		out.NormalMod[0] += wz * rz.NormalMod[0]
		out.NormalMod[1] += wz * rz.NormalMod[1]
		if wz > bestW {
			out.BlockID = rz.BlockID
		}
	}
	return out
}

// =============================================================================
// 3. Tetrahedral Normal, 3D Mean Curvature, Soft Shadow & 5-Tap 3D AO
// =============================================================================

// CalcSDF3DNormalAndCurvature evaluates the unit surface normal N and normalized
// 3D mean curvature kappa in [-1, 1] (>0 convex ridge, <0 concave valley) using a
// 4-point tetrahedral finite-difference stencil plus center evaluation.
func CalcSDF3DNormalAndCurvature(sdf SDF3D, p Vec3, eps float64) (Vec3, float64) {
	return CalcSDF3DNormalAndCurvatureAtHit(sdf, p, sdf(p), eps)
}

// CalcSDF3DNormalAndCurvatureAtHit evaluates the unit surface normal N and normalized
// 3D mean curvature kappa in [-1, 1] using 4 tetrahedral samples and a known center distance dCenter.
func CalcSDF3DNormalAndCurvatureAtHit(sdf SDF3D, p Vec3, dCenter, eps float64) (Vec3, float64) {
	if eps <= 0 {
		eps = 0.002
	}
	k0 := Vec3{1, -1, -1}
	k1 := Vec3{-1, -1, 1}
	k2 := Vec3{-1, 1, -1}
	k3 := Vec3{1, 1, 1}

	d0 := sdf(p.Add(k0.Scale(eps)))
	d1 := sdf(p.Add(k1.Scale(eps)))
	d2 := sdf(p.Add(k2.Scale(eps)))
	d3 := sdf(p.Add(k3.Scale(eps)))

	grad := k0.Scale(d0).Add(k1.Scale(d1)).Add(k2.Scale(d2)).Add(k3.Scale(d3))
	norm := grad.Normalize()

	// Laplacian estimate from tetrahedral shell minus center
	lap := (d0 + d1 + d2 + d3 - 4.0*dCenter) / (eps * 2.5)
	curv := common.Constrain(lap, -1.0, 1.0)
	return norm, curv
}

// IntersectBoundingSphere analytically intersects ray (ro, rd) with a sphere at center of radius r.
// Returns [tMin, tMax] along the ray and hit=true if the ray enters the sphere ahead of ro.
func IntersectBoundingSphere(ro, rd, center Vec3, radius float64) (tMin, tMax float64, hit bool) {
	oc := ro.Sub(center)
	b := oc.Dot(rd)
	c := oc.Dot(oc) - radius*radius
	disc := b*b - c
	if disc < 0.0 {
		return 0, 0, false
	}
	h := math.Sqrt(disc)
	t0 := -b - h
	t1 := -b + h
	if t1 < 0.0 {
		return 0, 0, false
	}
	if t0 < 0.0 {
		t0 = 0.0
	}
	return t0, t1, true
}

// CalcSDF3DSoftShadow marches a shadow ray from ro along rd and computes a smooth
// penumbra shadow factor in [0, 1] (0 = full umbra, 1 = unoccluded).
func CalcSDF3DSoftShadow(sdf SDF3D, ro, rd Vec3, mint, maxt, k float64) float64 {
	res := 1.0
	t := mint
	prevH := 1e10
	for i := 0; i < 28 && t < maxt; i++ {
		h := sdf(ro.Add(rd.Scale(t)))
		if h < 0.0005 {
			return 0.0
		}
		y := h * h / (2.0 * prevH)
		d := math.Sqrt(math.Max(0.0, h*h-y*y))
		res = math.Min(res, k*d/math.Max(0.0001, t-y))
		prevH = h
		t += common.Constrain(h, 0.015, 0.22)
	}
	return common.Clamp01(res)
}

// CalcSDF3DAO evaluates 5-step exponential 3D ambient occlusion along normal n.
func CalcSDF3DAO(sdf SDF3D, p, n Vec3) float64 {
	occ := 0.0
	weight := 1.0
	for i := 1; i <= 5; i++ {
		dist := 0.025 * float64(i*i)
		d := sdf(p.Add(n.Scale(dist)))
		occ += (dist - d) * weight
		weight *= 0.55
	}
	return common.Clamp01(1.0 - 2.4*occ)
}

// RaymarchSDF3D marches a ray from ro along unit direction rd up to maxDist using
// Keinert over-relaxation sphere tracing with automatic step fallback and secant root
// refinement, returning a complete SDF3DHit with normal, 3D mean curvature, AO, and soft shadow.
func RaymarchSDF3D(sdf SDF3D, ro, rd, lightDir Vec3, maxDist float64, maxSteps int) SDF3DHit {
	return raymarchSDF3DInterval(sdf, ro, rd, lightDir, 0.02, maxDist, maxSteps)
}

// RaymarchSDF3DBounded performs bounding-sphere pre-rejection before sphere tracing sdf
// across the intersected [tMin, tMax] interval, skipping empty sky/background rays in 0 SDF evaluations.
func RaymarchSDF3DBounded(sdf SDF3D, ro, rd, lightDir, boundCenter Vec3, boundRadius, maxDist float64, maxSteps int) SDF3DHit {
	if boundRadius > 0 {
		t0, t1, ok := IntersectBoundingSphere(ro, rd, boundCenter, boundRadius)
		if !ok || t0 >= maxDist {
			return SDF3DHit{Hit: false, T: maxDist}
		}
		tStart := math.Max(0.02, t0)
		tEnd := math.Min(maxDist, t1)
		return raymarchSDF3DInterval(sdf, ro, rd, lightDir, tStart, tEnd, maxSteps)
	}
	return raymarchSDF3DInterval(sdf, ro, rd, lightDir, 0.02, maxDist, maxSteps)
}

func raymarchSDF3DInterval(sdf SDF3D, ro, rd, lightDir Vec3, tStart, tEnd float64, maxSteps int) SDF3DHit {
	if maxSteps <= 0 {
		maxSteps = 72
	}
	t := tStart
	omega := 1.25
	prevRadius := 0.0
	stepLength := 0.0

	for i := 0; i < maxSteps && t < tEnd; i++ {
		pos := ro.Add(rd.Scale(t))
		d := sdf(pos)

		// Detect over-relaxation step overshoot and revert to standard sphere tracing
		if omega > 1.0 && prevRadius > 0 && (prevRadius+math.Abs(d) < stepLength) {
			stepLength -= omega * prevRadius
			t += stepLength
			omega = 1.0
			prevRadius = 0.0
			continue
		}

		if d < 0.0012 {
			if d < 0 && prevRadius > 0 && stepLength > 0 {
				// Secant root interpolation between previous positive sample and current negative sample
				frac := prevRadius / (prevRadius - d)
				t = (t - stepLength) + stepLength*common.Clamp01(frac)
				pos = ro.Add(rd.Scale(t))
				d = 0.0
			}
			norm, curv := CalcSDF3DNormalAndCurvatureAtHit(sdf, pos, d, 0.0025)
			ao := CalcSDF3DAO(sdf, pos, norm)
			sh := CalcSDF3DSoftShadow(sdf, pos.Add(norm.Scale(0.006)), lightDir, 0.015, tEnd*0.65, 10.0)
			return SDF3DHit{
				Hit:       true,
				T:         t,
				Pos:       pos,
				Normal:    norm,
				Curvature: curv,
				AO:        ao,
				Shadow:    sh,
			}
		}
		stepLength = d * omega
		prevRadius = math.Abs(d)
		t += stepLength
	}
	return SDF3DHit{Hit: false, T: tEnd}
}

// =============================================================================
// 4. Phase 3: 3D-to-2D Vector Hidden-Line Removal (HLR) & Contour Extractor
// =============================================================================

// SDF3DVectorHLRConfig configures 3D SDF silhouette, crease, and topographic cross-section
// vector contour extraction with raymarched hidden-line removal.
type SDF3DVectorHLRConfig struct {
	GridRes          int     // Screen-space Marching-Squares evaluation grid resolution (default 140)
	MaxDist          float64 // Maximum raymarching distance (default 18.0)
	FOV              float64 // Camera focal length multiplier (default 1.75)
	DepthJumpThresh  float64 // Depth discontinuity threshold for internal occlusion contours (default 0.22)
	NormalCreaseCos  float64 // Normal dot-product threshold for sharp CSG crease extraction (default 0.80)
	SliceLevels      int     // Number of 3D object-space topographic cross-section iso-contours (default 10)
	SliceAxis        Vec3    // 3D axis along which topographic cross-section contours are sliced (default Y=(0,1,0))
}

// DefaultSDF3DVectorHLRConfig returns a balanced configuration for 3D SDF vector contour extraction.
func DefaultSDF3DVectorHLRConfig() SDF3DVectorHLRConfig {
	return SDF3DVectorHLRConfig{
		GridRes:         140,
		MaxDist:         18.0,
		FOV:             1.75,
		DepthJumpThresh: 0.22,
		NormalCreaseCos: 0.80,
		SliceLevels:     10,
		SliceAxis:       V3(0, 1, 0),
	}
}

// ExtractSDF3DVectorContours extracts visible 2D vector polylines (outer silhouettes, occlusion
// boundaries, CSG normal creases, and 3D topographic cross-section contours) from a 3D SDF scene
// viewed from camPos toward camTarget, performing hidden-line removal automatically via primary raymarching.
func ExtractSDF3DVectorContours(
	sdf SDF3D,
	camPos, camTarget Vec3,
	w, h int,
	cfg SDF3DVectorHLRConfig,
) (silhouettes, creases, slices [][]Point) {
	if sdf == nil || w <= 0 || h <= 0 {
		return nil, nil, nil
	}
	nx := cfg.GridRes
	if nx < 24 {
		nx = 120
	}
	ny := int(math.Round(float64(nx) * float64(h) / float64(w)))
	if ny < 24 {
		ny = 24
	}
	maxDist := cfg.MaxDist
	if maxDist <= 0 {
		maxDist = 18.0
	}
	fov := cfg.FOV
	if fov <= 0 {
		fov = 1.75
	}
	sliceAxis := cfg.SliceAxis.Normalize()
	if sliceAxis.Len() < 1e-6 {
		sliceAxis = V3(0, 1, 0)
	}

	// Build orthonormal camera basis (right, up, forward)
	forward := camTarget.Sub(camPos).Normalize()
	worldUp := V3(0, 1, 0)
	if math.Abs(forward.Y) > 0.96 {
		worldUp = V3(1, 0, 0)
	}
	right := forward.Cross(worldUp).Normalize()
	up := right.Cross(forward).Normalize()
	aspect := float64(w) / float64(h)

	gw, gh := nx+1, ny+1
	hits := make([]SDF3DHit, gw*gh)
	minSlice := 1e9
	maxSlice := -1e9

	parallelRows(gw, gh, func(y0, y1 int) {
		for iy := y0; iy < y1; iy++ {
			vy := (0.5 - float64(iy)/float64(ny)) * 2.0
			rowOff := iy * gw
			for ix := 0; ix < gw; ix++ {
				vx := (float64(ix)/float64(nx) - 0.5) * 2.0 * aspect
				rd := forward.Scale(fov).Add(right.Scale(vx)).Add(up.Scale(vy)).Normalize()
				t := 0.02
				var hHit SDF3DHit
				for step := 0; step < 64 && t < maxDist; step++ {
					pos := camPos.Add(rd.Scale(t))
					d := sdf(pos)
					if d < 0.0015 {
						norm, _ := CalcSDF3DNormalAndCurvatureAtHit(sdf, pos, d, 0.0025)
						hHit = SDF3DHit{
							Hit:    true,
							T:      t,
							Pos:    pos,
							Normal: norm,
						}
						break
					}
					t += d * 0.95
				}
				hits[rowOff+ix] = hHit
			}
		}
	})

	for _, ht := range hits {
		if ht.Hit {
			sVal := ht.Pos.Dot(sliceAxis)
			if sVal < minSlice {
				minSlice = sVal
			}
			if sVal > maxSlice {
				maxSlice = sVal
			}
		}
	}

	fw, fh := float64(w), float64(h)
	sampleBilinearHit := func(x, y float64) SDF3DHit {
		u := common.Clamp01(x/fw) * float64(nx)
		v := common.Clamp01(y/fh) * float64(ny)
		ix := common.ConstrainInt(int(math.Round(u)), 0, nx)
		iy := common.ConstrainInt(int(math.Round(v)), 0, ny)
		return hits[iy*gw+ix]
	}

	// 1. Extract outer silhouette contour (hit vs miss boundary)
	silhouettes = ExtractMarchingSquares(nx, ny, fw, fh, 0.0, func(x, y float64) float64 {
		ht := sampleBilinearHit(x, y)
		if ht.Hit {
			return 1.0
		}
		return -1.0
	})

	// 2. Extract internal occlusion depth-jump & sharp normal-crease contours
	creases = ExtractMarchingSquares(nx, ny, fw, fh, 0.0, func(x, y float64) float64 {
		u := common.ConstrainInt(int(math.Round((x/fw)*float64(nx))), 1, nx-1)
		v := common.ConstrainInt(int(math.Round((y/fh)*float64(ny))), 1, ny-1)
		c := hits[v*gw+u]
		if !c.Hit {
			return -1.0
		}
		r := hits[v*gw+u+1]
		d := hits[(v+1)*gw+u]
		if !r.Hit || !d.Hit {
			return -1.0
		}
		dJump := math.Max(math.Abs(c.T-r.T), math.Abs(c.T-d.T))
		if dJump > cfg.DepthJumpThresh {
			return dJump - cfg.DepthJumpThresh
		}
		minCos := math.Min(c.Normal.Dot(r.Normal), c.Normal.Dot(d.Normal))
		if minCos < cfg.NormalCreaseCos {
			return cfg.NormalCreaseCos - minCos
		}
		return -1.0
	})

	// 3. Extract visible 3D object-space topographic cross-section iso-contours
	if cfg.SliceLevels > 0 && maxSlice > minSlice+1e-4 {
		for lvl := 1; lvl <= cfg.SliceLevels; lvl++ {
			isoVal := minSlice + (float64(lvl)/float64(cfg.SliceLevels+1))*(maxSlice-minSlice)
			rawLines := ExtractMarchingSquares(nx, ny, fw, fh, isoVal, func(x, y float64) float64 {
				ht := sampleBilinearHit(x, y)
				if !ht.Hit {
					return minSlice - 100.0
				}
				return ht.Pos.Dot(sliceAxis)
			})
			// Filter segments so iso-contours never jump across occluded background gaps or sky
			maxSegLen := math.Hypot(fw/float64(nx), fh/float64(ny)) * 2.2
			for _, poly := range rawLines {
				var cur []Point
				for _, pt := range poly {
					ht := sampleBilinearHit(pt.X, pt.Y)
					if !ht.Hit || math.Abs(ht.Pos.Dot(sliceAxis)-isoVal) > (maxSlice-minSlice)*0.18 {
						if len(cur) >= 2 {
							slices = append(slices, cur)
						}
						cur = nil
						continue
					}
					if len(cur) > 0 && cur[len(cur)-1].Dist(pt) > maxSegLen {
						if len(cur) >= 2 {
							slices = append(slices, cur)
						}
						cur = []Point{pt}
						continue
					}
					cur = append(cur, pt)
				}
				if len(cur) >= 2 {
					slices = append(slices, cur)
				}
			}
		}
	}

	return silhouettes, creases, slices
}

// RenderSDF3DVectorContours extracts visible 3D SDF silhouettes, creases, and cross-section
// iso-contours with Hidden-Line Removal and strokes them as vector polylines onto c (ready for
// both HDR raster and infinite-resolution SVG/PDF/Pen-Plotter export).
func (c *Canvas) RenderSDF3DVectorContours(
	sdf SDF3D,
	camPos, camTarget Vec3,
	cfg SDF3DVectorHLRConfig,
	silhouetteCol, creaseCol, sliceCol color.RGBA,
	baseWidth float64,
) {
	if c == nil || sdf == nil {
		return
	}
	if baseWidth <= 0 {
		baseWidth = 1.6
	}
	sil, creases, slices := ExtractSDF3DVectorContours(sdf, camPos, camTarget, c.Width, c.Height, cfg)
	for _, poly := range slices {
		c.StrokePolyline(poly, baseWidth*0.55, sliceCol, 0.78)
	}
	for _, poly := range creases {
		c.StrokePolyline(poly, baseWidth*0.82, creaseCol, 0.90)
	}
	for _, poly := range sil {
		c.StrokePolyline(poly, baseWidth*1.25, silhouetteCol, 0.98)
	}
}
