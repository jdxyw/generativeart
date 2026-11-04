package render

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart/common"
)

// PrincipalCurvature2D stores the orthogonal principal curvature unit vectors
// (E1, E2), eigenvalues (K1, K2), and surface gradient for a 2.5D heightfield.
type PrincipalCurvature2D struct {
	E1   Point   // Primary curvature unit direction (direction of maximum bending)
	E2   Point   // Secondary curvature unit direction (orthogonal contour/minimum bending)
	K1   float64 // Maximum principal curvature eigenvalue
	K2   float64 // Minimum principal curvature eigenvalue
	Grad Point   // First-order height gradient (dH/dx, dH/dy)
}

// ComputePrincipalCurvature2D analytically diagonalizes the 2x2 Hessian tensor of
// heightfieldFn(x, y) at (x, y) using central finite differences of step eps,
// returning the orthogonal principal curvature frame (E1, E2) and curvatures (K1, K2).
func ComputePrincipalCurvature2D(heightfieldFn ScalarField2D, x, y, eps float64) PrincipalCurvature2D {
	if eps <= 0 {
		eps = 1.5
	}
	inv2Eps := 1.0 / (2.0 * eps)
	invEpsSq := 1.0 / (eps * eps)

	hC := heightfieldFn(x, y)
	hR := heightfieldFn(x+eps, y)
	hL := heightfieldFn(x-eps, y)
	hD := heightfieldFn(x, y+eps)
	hU := heightfieldFn(x, y-eps)
	hDR := heightfieldFn(x+eps, y+eps)
	hDL := heightfieldFn(x-eps, y+eps)
	hUR := heightfieldFn(x+eps, y-eps)
	hUL := heightfieldFn(x-eps, y-eps)

	hx := (hR - hL) * inv2Eps
	hy := (hD - hU) * inv2Eps
	hxx := (hR - 2.0*hC + hL) * invEpsSq
	hyy := (hD - 2.0*hC + hU) * invEpsSq
	hxy := (hDR - hUR - hDL + hUL) * (0.25 * invEpsSq)

	// Diagonalize symmetric 2x2 Hessian [[hxx, hxy], [hxy, hyy]]
	trace := hxx + hyy
	diff := hxx - hyy
	disc := math.Hypot(diff, 2.0*hxy)
	k1 := 0.5 * (trace + disc)
	k2 := 0.5 * (trace - disc)

	// Principal eigenvector angle theta = 0.5 * atan2(2*hxy, hxx - hyy)
	var e1, e2 Point
	if disc > 1e-9 {
		theta := 0.5 * math.Atan2(2.0*hxy, diff)
		c, s := math.Cos(theta), math.Sin(theta)
		e1 = Point{X: c, Y: s}
		e2 = Point{X: -s, Y: c}
	} else {
		// Fallback to isophote contour / gradient frame when surface is umbilical/flat
		gLen := math.Hypot(hx, hy)
		if gLen > 1e-9 {
			e1 = Point{X: hx / gLen, Y: hy / gLen}
			e2 = Point{X: -hy / gLen, Y: hx / gLen}
		} else {
			e1 = Point{X: 1, Y: 0}
			e2 = Point{X: 0, Y: 1}
		}
	}

	return PrincipalCurvature2D{
		E1:   e1,
		E2:   e2,
		K1:   k1,
		K2:   k2,
		Grad: Point{X: hx, Y: hy},
	}
}

// EngravingConfig configures Albrecht Dürer / Gustave Doré / Piranesi style
// principal-curvature intaglio copperplate burin engraving.
type EngravingConfig struct {
	PrimaryDSep       float64    // Separation between primary curvature streamlines in pixels (default 5.2)
	SecondaryDSep     float64    // Separation between orthogonal cross-hatch streamlines in pixels (default 6.4)
	MaxStrokeWidth    float64    // Maximum burin line width in deep shadow (default 1.55)
	MinStrokeWidth    float64    // Minimum hairline width in midtone (default 0.32)
	LightCutoff       float64    // Illumination threshold above which primary burin lifts off (default 0.78)
	CrossHatchCutoff  float64    // Illumination threshold below which orthogonal cross-hatching engages (default 0.44)
	InkColor          color.RGBA // Copperplate intaglio ink color
	HighlightChalkCol color.RGBA // Optional white/gold chalk highlight stroke color in chiaroscuro prints
	EnableChalk       bool       // Draw delicate white/gold chalk highlight strokes where illum > 0.82
}

// DefaultEngravingConfig returns a museum-grade Renaissance copperplate burin configuration.
func DefaultEngravingConfig() EngravingConfig {
	return EngravingConfig{
		PrimaryDSep:       5.0,
		SecondaryDSep:     6.2,
		MaxStrokeWidth:    1.50,
		MinStrokeWidth:    0.32,
		LightCutoff:       0.78,
		CrossHatchCutoff:  0.44,
		InkColor:          color.RGBA{R: 0x14, G: 0x10, B: 0x16, A: 255},
		HighlightChalkCol: color.RGBA{R: 0xfc, G: 0xf4, B: 0xdc, A: 255},
		EnableChalk:       true,
	}
}

// RenderCurvatureEngraving traces evenly-spaced Jobard-Lefer streamlines along both
// orthogonal principal-curvature vector fields (E2 contour flow and E1 cross-curvature flow)
// of heightFn(x, y), and renders each streamline with variable-width burin pressure
// driven by illumFn(x, y) in [0, 1] (swelling in shadow, tapering to hairlines in light).
func RenderCurvatureEngraving(
	rc *Canvas,
	heightFn ScalarField2D,
	illumFn ScalarField2D,
	maskFn func(x, y float64) bool,
	cfg EngravingConfig,
) {
	if rc == nil || heightFn == nil || illumFn == nil {
		return
	}
	w, h := float64(rc.Width), float64(rc.Height)
	if w <= 0 || h <= 0 {
		return
	}
	if cfg.PrimaryDSep <= 1.5 {
		cfg.PrimaryDSep = 5.0
	}
	if cfg.SecondaryDSep <= 1.5 {
		cfg.SecondaryDSep = 6.2
	}

	// Precompute principal curvature directions on a coarse grid (cell = 3px) for fast RK4 lookup
	gw := int(math.Ceil(w/3.0)) + 2
	gh := int(math.Ceil(h/3.0)) + 2
	e1Grid := make([]Point, gw*gh)
	e2Grid := make([]Point, gw*gh)

	for gy := 0; gy < gh; gy++ {
		py := float64(gy) * 3.0
		for gx := 0; gx < gw; gx++ {
			px := float64(gx) * 3.0
			pc := ComputePrincipalCurvature2D(heightFn, px, py, 2.5)
			e1Grid[gy*gw+gx] = pc.E1
			e2Grid[gy*gw+gx] = pc.E2
		}
	}

	sampleField := func(grid []Point, x, y float64) Point {
		gx := int(x / 3.0)
		gy := int(y / 3.0)
		if gx < 0 {
			gx = 0
		} else if gx >= gw {
			gx = gw - 1
		}
		if gy < 0 {
			gy = 0
		} else if gy >= gh {
			gy = gh - 1
		}
		return grid[gy*gw+gx]
	}

	// Helper to stroke a streamline with dynamic burin width modulated per-segment by illumFn
	strokeVariableBurin := func(pts []Point, cutoff float64, maxW, minW float64, col color.RGBA, invertIllum bool) {
		if len(pts) < 2 {
			return
		}
		// Break polyline into short segments with per-segment illumination-modulated width
		const chunkStep = 3
		for i := 0; i < len(pts)-1; i += chunkStep {
			end := i + chunkStep
			if end >= len(pts) {
				end = len(pts) - 1
			}
			mid := pts[(i+end)/2]
			if maskFn != nil && !maskFn(mid.X, mid.Y) {
				continue
			}
			illum := common.Clamp01(illumFn(mid.X, mid.Y))
			var strength float64
			if invertIllum {
				// White/gold chiaroscuro chalk highlight strokes in bright zones (illum > cutoff)
				if illum <= cutoff {
					continue
				}
				strength = common.Smoothstep(cutoff, 0.98, illum)
			} else {
				// Intaglio dark ink burin strokes in shadow/midtone zones (illum < cutoff)
				if illum >= cutoff {
					continue
				}
				strength = math.Pow(common.Clamp01((cutoff-illum)/math.Max(cutoff, 1e-3)), 0.85)
			}

			segW := minW + (maxW-minW)*strength
			segAlpha := common.Clamp01(0.28 + 0.68*strength)
			rc.StrokePolyline(pts[i:end+1], segW, col, segAlpha)
		}
	}

	// 1. Primary Burin Engraving Pass along Secondary Principal Curvature E2 (contour-wrapping direction)
	primaryLines := TraceEvenStreamlines(StreamlineConfig{
		Width:    w,
		Height:   h,
		DSep:     cfg.PrimaryDSep,
		DTest:    cfg.PrimaryDSep * 0.52,
		StepSize: 1.8,
		MaxSteps: 90,
	}, func(x, y float64) Point {
		return sampleField(e2Grid, x, y)
	})
	for _, line := range primaryLines {
		strokeVariableBurin(line, cfg.LightCutoff, cfg.MaxStrokeWidth, cfg.MinStrokeWidth, cfg.InkColor, false)
		if cfg.EnableChalk {
			strokeVariableBurin(line, 0.82, cfg.MaxStrokeWidth*0.78, cfg.MinStrokeWidth, cfg.HighlightChalkCol, true)
		}
	}

	// 2. Orthogonal Cross-Hatching Pass along Primary Principal Curvature E1 in Shadow Zones
	secondaryLines := TraceEvenStreamlines(StreamlineConfig{
		Width:    w,
		Height:   h,
		DSep:     cfg.SecondaryDSep,
		DTest:    cfg.SecondaryDSep * 0.52,
		StepSize: 1.8,
		MaxSteps: 75,
	}, func(x, y float64) Point {
		return sampleField(e1Grid, x, y)
	})
	for _, line := range secondaryLines {
		strokeVariableBurin(line, cfg.CrossHatchCutoff, cfg.MaxStrokeWidth*0.88, cfg.MinStrokeWidth, cfg.InkColor, false)
	}
}

// Evaluate3DBurinEngraving computes temporally coherent, object-locked 3D copperplate
// burin cross-hatching and chiaroscuro white-chalk highlights directly in a rotating
// 3D object's local coordinate frame (pLocal, nLocal). Because the engraving coordinate
// frame rotates rigidly with each 3D facet while line width dynamically swells in shadow
// and tapers in sunlight, rotating 3D polyhedra and armillary rings exhibit zero 2D
// streamline popping across animation frames.
func Evaluate3DBurinEngraving(pLocal, nLocal Vec3, illum, freq float64) (inkWeight, chalkWeight float64) {
	if freq <= 1.0 {
		freq = 72.0
	}
	// Construct a smooth tangent frame (t1, t2) on the 3D facet from nLocal
	up := Vec3{0, 1, 0}
	if math.Abs(nLocal.Y) > 0.85 {
		up = Vec3{1, 0, 0}
	}
	t1 := nLocal.Cross(up).Normalize()
	t2 := nLocal.Cross(t1).Normalize()

	u := pLocal.Dot(t1) * freq
	v := pLocal.Dot(t2) * freq

	// Primary burin contour lines along u (engages when illum < 0.58)
	if illum < 0.58 {
		shadowStrength := common.Smoothstep(0.58, 0.16, illum)
		w1 := 0.5 + 0.5*math.Cos(u)
		// Line width swells smoothly as shadowStrength increases
		thresh1 := 0.08 + 0.24*shadowStrength
		line1 := (1.0 - common.Smoothstep(thresh1-0.06, thresh1+0.06, w1)) * (0.45 + 0.48*shadowStrength)
		inkWeight = line1

		// Secondary orthogonal burin cross-hatching along v in deep shadow (illum < 0.34)
		if illum < 0.34 {
			deepStrength := common.Smoothstep(0.34, 0.08, illum)
			w2 := 0.5 + 0.5*math.Cos(v)
			thresh2 := 0.07 + 0.20*deepStrength
			line2 := (1.0 - common.Smoothstep(thresh2-0.06, thresh2+0.06, w2)) * 0.78 * deepStrength
			inkWeight = math.Max(inkWeight, line2)
		}
	}

	// Renaissance chiaroscuro white-chalk / gold highlight strokes on sunlit facets (illum > 0.74)
	if illum > 0.74 {
		hiStrength := common.Smoothstep(0.74, 0.96, illum)
		wHi := 0.5 + 0.5*math.Cos(u*0.85+0.4*v)
		threshHi := 0.06 + 0.14*hiStrength
		chalkWeight = (1.0 - common.Smoothstep(threshHi-0.05, threshHi+0.05, wHi)) * 0.68 * hiStrength
	}

	return common.Clamp01(inkWeight), common.Clamp01(chalkWeight)
}
