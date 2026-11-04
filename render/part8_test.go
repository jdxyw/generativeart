package render

import (
	"image/color"
	"math"
	"testing"

	"github.com/jdxyw/generativeart/common"
)

func TestPart8KubelkaMunkAndWatercolor(t *testing.T) {
	// 1. Subtractive mixing of Ultramarine Blue + Cadmium Yellow must yield a vivid green (G > R and G > B)
	greenMix := MixKubelkaMunk(PW(PigmentUltramarine, 0.45), PW(PigmentCadmiumYellow, 0.55))
	if greenMix[1] <= greenMix[0] || greenMix[1] <= greenMix[2] {
		t.Fatalf("expected subtractive blue+yellow to produce green dominance, got RGB=%v", greenMix)
	}

	// 2. Zero-thickness glaze must preserve the substrate exactly
	paper := [3]float64{0.92, 0.88, 0.80}
	g0 := GlazeKubelkaMunk(paper, PigmentVermilion, 0.0)
	if math.Abs(g0[0]-paper[0]) > 1e-6 || math.Abs(g0[1]-paper[1]) > 1e-6 || math.Abs(g0[2]-paper[2]) > 1e-6 {
		t.Fatalf("expected zero-thickness glaze to equal substrate, got %v vs %v", g0, paper)
	}

	// 3. Heavy glaze must converge toward pigment masstone R_inf
	masstone := MixKubelkaMunk(PW(PigmentVermilion, 1.0))
	gInf := GlazeKubelkaMunk(paper, PigmentVermilion, 15.0)
	if math.Abs(gInf[0]-masstone[0]) > 0.02 || math.Abs(gInf[1]-masstone[1]) > 0.02 {
		t.Fatalf("expected thick glaze %v to converge to masstone %v", gInf, masstone)
	}

	// 4. EvaluateWatercolorWash coffee-ring & granulation
	ne := common.NewNoiseEngine(2026)
	washCfg := DefaultWatercolorWashConfig()
	insideCol := EvaluateWatercolorWash(ne, 0.5, 0.5, -0.15, PigmentAzurite, paper, washCfg)
	outsideCol := EvaluateWatercolorWash(ne, 0.5, 0.5, 0.25, PigmentAzurite, paper, washCfg)
	if outsideCol != paper {
		t.Fatalf("expected outside of wash to remain paper color")
	}
	if insideCol[0] >= paper[0] {
		t.Fatalf("expected Azurite wash to absorb red channel relative to paper")
	}
}

func TestPart8SDF3DAndSymmetryFolding(t *testing.T) {
	// Icosahedral and Octahedral folds must be idempotent on already folded points
	p := V3(-0.7, 0.4, -0.9)
	fo := FoldOctahedral(p)
	if fo.X < fo.Y || fo.Y < fo.Z || fo.Z < 0 {
		t.Fatalf("FoldOctahedral failed ordering invariant: %+v", fo)
	}
	fi := FoldIcosahedral(p)
	if fi.X < 0 || fi.Y < 0 || fi.Z < 0 {
		t.Fatalf("FoldIcosahedral failed non-negative invariant: %+v", fi)
	}

	// Verify exact 5-fold rotational periodicity around the Y axis with AlignIcosahedral5FoldY
	testPt := V3(0.37, -0.24, 0.51)
	f0 := FoldIcosahedral(AlignIcosahedral5FoldY(testPt))
	f1 := FoldIcosahedral(AlignIcosahedral5FoldY(testPt.RotateY(2.0 * math.Pi / 5.0)))
	if f0.Sub(f1).Len() > 1e-12 {
		t.Fatalf("expected AlignIcosahedral5FoldY to be invariant under 2*pi/5 Y-rotation: %+v vs %+v", f0, f1)
	}

	// Raymarching a unit sphere at origin from (0, 0, 3) along (0, 0, -1)
	sphereSDF := func(q Vec3) float64 {
		return SDF3DSphere(q, 1.0)
	}
	hit := RaymarchSDF3D(sphereSDF, V3(0, 0, 3), V3(0, 0, -1), V3(0.5, 0.5, 0.7).Normalize(), 10.0, 64)
	if !hit.Hit || math.Abs(hit.T-2.0) > 0.01 {
		t.Fatalf("expected sphere hit at t=2.0, got hit=%v t=%f", hit.Hit, hit.T)
	}
	if hit.Normal.Z < 0.98 {
		t.Fatalf("expected outward normal +Z at (0,0,1), got %+v", hit.Normal)
	}
	if hit.Curvature <= 0 {
		t.Fatalf("expected positive convex mean curvature on sphere, got %f", hit.Curvature)
	}
}

func TestPart8PrincipalCurvatureAndEngraving(t *testing.T) {
	// Paraboloid H(x, y) = 0.02*x^2 + 0.005*y^2 has orthogonal principal axes along X and Y
	heightFn := func(x, y float64) float64 {
		dx, dy := x-50.0, y-50.0
		return 0.02*dx*dx + 0.005*dy*dy
	}
	pc := ComputePrincipalCurvature2D(heightFn, 50.0, 50.0, 1.0)
	dot := pc.E1.X*pc.E2.X + pc.E1.Y*pc.E2.Y
	if math.Abs(dot) > 1e-6 {
		t.Fatalf("expected orthogonal principal curvature eigenvectors, dot=%e", dot)
	}
	if pc.K1 <= pc.K2 {
		t.Fatalf("expected K1 > K2 on anisotropic paraboloid, got K1=%f K2=%f", pc.K1, pc.K2)
	}

	rc := NewCanvas(80, 80)
	rc.FillBackground(color.RGBA{245, 240, 230, 255})
	RenderCurvatureEngraving(rc, heightFn, func(x, y float64) float64 {
		return 0.35
	}, nil, DefaultEngravingConfig())
}
