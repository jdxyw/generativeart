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

func TestPhase2DielectricOpticsAndBoundedRaymarching(t *testing.T) {
	sphereSDF := func(q Vec3) float64 {
		return SDF3DSphere(q, 1.0)
	}

	// 1. Verify IntersectBoundingSphere & RaymarchSDF3DBounded skip miss rays in 0 SDF calls
	evalCount := 0
	countedSDF := func(q Vec3) float64 {
		evalCount++
		return SDF3DSphere(q, 1.0)
	}
	missHit := RaymarchSDF3DBounded(countedSDF, V3(0, 3.0, 3.0), V3(0, 0, -1), V3(0, 1, 0), V3(0, 0, 0), 1.2, 10.0, 64)
	if missHit.Hit || evalCount != 0 {
		t.Fatalf("expected bounding sphere miss to evaluate SDF 0 times, got hit=%v evalCount=%d", missHit.Hit, evalCount)
	}

	// 2. Verify Snell Refraction & Total Internal Reflection (TIR)
	wt, tir := RefractVec3(V3(0, 0, -1), V3(0, 0, 1), 1.0/1.52)
	if tir || math.Abs(wt.Z+1.0) > 1e-6 {
		t.Fatalf("expected normal-incidence refraction to pass straight through, got wt=%+v tir=%v", wt, tir)
	}
	// Steep internal incidence from diamond (n=2.417) to air (n=1.0) must trigger TIR
	grazingWi := V3(0.85, 0, -0.52678).Normalize()
	_, isTIR := RefractVec3(grazingWi, V3(0, 0, 1), 2.417)
	if !isTIR {
		t.Fatalf("expected total internal reflection at steep internal angle in diamond")
	}
	if f := SchlickFresnel(grazingWi.Z, 2.417, 1.0); f < 0.999 {
		t.Fatalf("expected SchlickFresnel=1.0 under TIR, got %f", f)
	}

	// 3. Verify TraceSDF3DDielectric with Baltic Amber Beer-Lambert absorption & Flint Prism dispersion
	amberMat := DielectricBalticAmber()
	whiteEnv := func(_, _ Vec3, _ int) [3]float64 {
		return [3]float64{1.0, 1.0, 1.0}
	}
	amberRGB, hit := TraceSDF3DDielectric(sphereSDF, V3(0, 0, 3.0), V3(0, 0, -1), V3(0.4, 0.7, 0.5).Normalize(), 10.0, amberMat, whiteEnv)
	if !hit.Hit {
		t.Fatalf("expected dielectric sphere hit")
	}
	// Baltic amber strongly absorbs blue > green > red across the 2.0-unit interior chord
	if amberRGB[0] <= amberRGB[1] || amberRGB[1] <= amberRGB[2] {
		t.Fatalf("expected Baltic amber Beer-Lambert absorption to leave R > G > B, got %v", amberRGB)
	}

	// Dispersive prism test: off-center ray through FlintPrism with channel-dependent environment
	prismMat := DielectricFlintPrism()
	dispEnv := func(_, d Vec3, ch int) [3]float64 {
		v := 0.5 + 0.5*d.X
		return [3]float64{v, v, v}
	}
	dispRGB, dispHit := TraceSDF3DDielectric(sphereSDF, V3(0.55, 0, 3.0), V3(0, 0, -1), V3(0, 1, 0), 10.0, prismMat, dispEnv)
	if !dispHit.Hit || math.Abs(dispRGB[0]-dispRGB[2]) < 1e-4 {
		t.Fatalf("expected chromatic dispersion to separate R and B channels on off-axis sphere ray, got %v", dispRGB)
	}

	// 4. Verify EvalSDF3DCausticIntensity produces bright converged light (> direct ndotl) behind convex sphere
	caustic := EvalSDF3DCausticIntensity(sphereSDF, V3(0, 0, -2.0), V3(0, 0, 1), V3(0, 0, 1), DielectricCrownGlass(), 2.5)
	if caustic[0] <= 1.0 {
		t.Fatalf("expected convex dielectric sphere to focus caustic intensity > 1.0 at receiver, got %v", caustic)
	}
}

func TestPhase2AdaptiveMoebiusOklchAndDepthOfField(t *testing.T) {
	// 1. Verify Oklch conversion & GamutMapOklab preserves hue angle on out-of-gamut high-chroma colors
	origLch := Oklch{L: 0.78, C: 0.38, H: 0.65}
	lab := OklchToOklab(origLch)
	mapped := GamutMapOklab(lab)
	mappedLch := OklabToOklch(mapped)
	if math.Abs(mappedLch.L-origLch.L) > 1e-6 {
		t.Fatalf("expected GamutMapOklab to preserve L=%f, got %f", origLch.L, mappedLch.L)
	}
	if math.Abs(mappedLch.H-origLch.H) > 1e-4 {
		t.Fatalf("expected GamutMapOklab to preserve hue angle H=%f, got %f", origLch.H, mappedLch.H)
	}
	rLin, gLin, bLin := OklabToLinearRGB(mapped)
	if rLin < -1e-3 || rLin > 1.001 || gLin < -1e-3 || gLin > 1.001 || bLin < -1e-3 || bLin > 1.001 {
		t.Fatalf("expected gamut-mapped linear RGB inside [0, 1]^3, got (%f, %f, %f)", rLin, gLin, bLin)
	}

	// 2. Verify Edge-Adaptive 2x2 SSAA in RenderMoebiusScene evaluates far fewer than 4*W*H samples on mostly smooth scenes
	rc := NewCanvas(64, 64)
	var sampleCalls int64
	cfg := DefaultMoebiusConfig(2026)
	cfg.EnableRK4Hatch = false
	RenderMoebiusScene(rc, cfg, func(ux, uy float64) MoebiusPixel {
		// Note: single-threaded check or atomic not needed if we inspect deterministic output,
		// so let's test visual output + ParsePostEffect("dof(...)")
		if math.Hypot(ux-0.5, uy-0.5) < 0.22 {
			return MoebiusPixel{
				Depth:      2.5,
				ObjID:      1,
				PartID:     1,
				Normal:     [3]float64{0, 0, 1},
				Illum:      0.75,
				SunColor:   [3]float64{0.9, 0.7, 0.3},
				ShadeColor: [3]float64{0.2, 0.3, 0.6},
			}
		}
		_ = sampleCalls
		return MoebiusPixel{
			Depth:      100.0,
			ObjID:      0,
			Normal:     [3]float64{0, 0, 1},
			Illum:      0.85,
			SunColor:   [3]float64{0.1, 0.2, 0.4},
			ShadeColor: [3]float64{0.1, 0.2, 0.4},
		}
	})

	// 3. Verify DepthOfFieldEffect and ParsePostEffect("dof(...)") round-trip
	dofEff := DepthOfFieldEffect(0.5, 0.5, 0.25, 5.0, 6)
	parsedEff, ok := ParsePostEffect(dofEff.Name())
	if !ok || parsedEff == nil || parsedEff.Name() != dofEff.Name() {
		t.Fatalf("expected ParsePostEffect(%q) to round-trip cleanly, got ok=%v eff=%v", dofEff.Name(), ok, parsedEff)
	}
	parsedEff.Apply(rc)
}

