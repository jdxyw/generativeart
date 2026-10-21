package render

import (
	"image/color"
	"math"
	"math/rand"
	"testing"

	"github.com/jdxyw/generativeart/common"
	"github.com/tdewolff/canvas"
)

func TestHDRLinearBlendingAndToneMapping(t *testing.T) {
	c := NewCanvas(16, 16)
	c.FillBackground(common.Black)

	// Accumulate HDR linear glow well above 1.0 using BlendAdd
	c.SetBlendMode(BlendAdd)
	brightOrange := color.RGBA{R: 255, G: 140, B: 40, A: 255}
	for i := 0; i < 10; i++ {
		c.BlendPixel(8, 8, brightOrange, 0.4)
	}

	rLin, gLin, bLin, a := c.PixelLinearAt(8, 8)
	if rLin <= 1.0 {
		t.Fatalf("expected HDR linear R > 1.0 after additive accumulation, got %f", rLin)
	}
	if gLin <= 0.1 || bLin <= 0.0 {
		t.Fatalf("expected positive linear G and B channels, got G=%f B=%f", gLin, bLin)
	}
	if math.Abs(a-1.0) > 1e-6 {
		t.Fatalf("expected alpha 1.0, got %f", a)
	}

	// Verify tone-mapped resolve produces valid clamped 8-bit sRGB output
	tm := c.ResolveToneMapped(nil, 1.0, 2.2)
	px := tm.RGBAAt(8, 8)
	if px.R < 200 || px.G < 100 || px.A != 255 {
		t.Fatalf("unexpected tone-mapped pixel at (8,8): %+v", px)
	}

	// Verify round-trip via NewCanvasFromRGBA
	c2 := NewCanvasFromRGBA(tm)
	r2, _, _, a2 := c2.PixelLinearAt(8, 8)
	if r2 <= 0.5 || math.Abs(a2-1.0) > 1e-6 {
		t.Fatalf("unexpected round-trip linear pixel from RGBA: r=%f a=%f", r2, a2)
	}
}

func TestOklabAndHCLGradients(t *testing.T) {
	c1 := color.RGBA{R: 20, G: 60, B: 220, A: 255}
	c2 := color.RGBA{R: 250, G: 200, B: 30, A: 255}

	midOklab := LerpOklab(c1, c2, 0.5)
	midHCL := LerpHCL(c1, c2, 0.5)
	if midOklab.A != 255 || midHCL.A != 255 {
		t.Fatalf("expected alpha 255 for interpolated colors")
	}

	grad := NewMultiStopGradient(c1, midOklab, c2)
	s0 := grad.SampleOklab(0.0)
	s1 := grad.SampleOklab(1.0)
	sHCL := grad.SampleHCL(0.5)
	if math.Abs(float64(s0.R)-float64(c1.R)) > 1 || math.Abs(float64(s1.R)-float64(c2.R)) > 1 {
		t.Fatalf("endpoint mismatch in MultiStopGradient: s0=%+v s1=%+v", s0, s1)
	}
	if sHCL.A != 255 {
		t.Fatalf("expected valid HCL gradient sample, got %+v", sHCL)
	}
}

func TestSplatPointEnergyConservation(t *testing.T) {
	c := NewCanvas(10, 10)
	c.FillBackground(common.Black)
	c.SetBlendMode(BlendAdd)

	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	alpha := 0.75
	c.SplatPoint(4.37, 5.62, white, alpha)

	var totalR float64
	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			r, _, _, _ := c.PixelLinearAt(x, y)
			totalR += r
		}
	}

	if math.Abs(totalR-alpha) > 1e-6 {
		t.Fatalf("SplatPoint did not conserve linear energy: expected %f, got %f", alpha, totalR)
	}
}

func TestDrawTaperedSplineAndVectorBridge(t *testing.T) {
	c := NewCanvas(64, 64)
	c.FillBackground(common.Black)

	pts := []Point{
		{X: 8, Y: 32},
		{X: 24, Y: 16},
		{X: 40, Y: 48},
		{X: 56, Y: 32},
	}
	grad := NewMultiStopGradient(
		color.RGBA{R: 255, G: 80, B: 120, A: 255},
		color.RGBA{R: 60, G: 200, B: 255, A: 255},
	)

	c.DrawTaperedSpline(pts, 8.0, TaperBothEnds, grad.SampleOklab, 0.9)

	p := &canvas.Path{}
	p.MoveTo(10, 10)
	p.QuadTo(32, 4, 54, 10)
	c.StrokePath(p, 2.5, canvas.RoundCap, canvas.RoundJoin, common.White, 0.8)

	nonBlack := 0
	img := c.ResolveToRGBA(nil)
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			px := img.RGBAAt(x, y)
			if px.R > 0 || px.G > 0 || px.B > 0 {
				nonBlack++
			}
		}
	}
	if nonBlack < 100 {
		t.Fatalf("expected DrawTaperedSpline and StrokePath to rasterize >100 pixels, got %d", nonBlack)
	}
}

func TestDrawWatercolorPolygon(t *testing.T) {
	c := NewCanvas(64, 64)
	c.FillBackground(common.White)
	c.SetBlendMode(BlendMultiply)

	rng := rand.New(rand.NewSource(99))
	poly := []Point{
		{X: 16, Y: 16},
		{X: 48, Y: 16},
		{X: 48, Y: 48},
		{X: 16, Y: 48},
	}
	pigment := color.RGBA{R: 40, G: 110, B: 180, A: 255}
	c.DrawWatercolorPolygon(poly, pigment, 12, 4, 0.18, 0.12, rng)

	rCenter, gCenter, bCenter, _ := c.PixelLinearAt(32, 32)
	if rCenter >= 0.95 || gCenter >= 0.95 || bCenter >= 0.95 {
		t.Fatalf("expected watercolor multiply layers to darken center pixel, got (%f, %f, %f)", rCenter, gCenter, bCenter)
	}
}

func TestDrawSandSplineAndBristleStroke(t *testing.T) {
	c := NewCanvas(64, 64)
	c.FillBackground(common.Black)
	c.SetBlendMode(BlendAdd)

	rng := rand.New(rand.NewSource(7))
	pts := []Point{
		{X: 8, Y: 20},
		{X: 25, Y: 44},
		{X: 56, Y: 24},
	}
	c.DrawSandSpline(pts, 12, 3.0, func(t float64) color.RGBA {
		return color.RGBA{R: 220, G: 180, B: 100, A: 255}
	}, 0.08, rng)

	c.SetBlendMode(BlendOver)
	c.DrawBristleStroke(pts, 6.0, 8, color.RGBA{R: 100, G: 180, B: 240, A: 255}, 0.5, rng)

	var totalEnergy float64
	for y := 0; y < c.Height; y++ {
		for x := 0; x < c.Width; x++ {
			r, g, b, _ := c.PixelLinearAt(x, y)
			totalEnergy += r + g + b
		}
	}
	if totalEnergy <= 1.0 {
		t.Fatalf("expected positive accumulated energy from sand spline and bristle stroke, got %f", totalEnergy)
	}
}

func TestApplyBloomAndPostProcess(t *testing.T) {
	c := NewCanvas(32, 32)
	c.FillBackground(common.Black)

	// Place an intense HDR impulse at center (16, 16)
	c.SetPixelLinear(16, 16, 8.0, 6.0, 4.0, 1.0)

	rBeforeNeighbor, _, _, _ := c.PixelLinearAt(18, 16)
	if rBeforeNeighbor != 0 {
		t.Fatalf("expected neighbor pixel to be 0 before bloom, got %f", rBeforeNeighbor)
	}

	c.ApplyBloom(0.8, 2.5, 1.2)

	rAfterNeighbor, _, _, _ := c.PixelLinearAt(18, 16)
	if rAfterNeighbor <= 0.01 {
		t.Fatalf("expected bloom to spread HDR light to neighboring pixel (18,16), got %f", rAfterNeighbor)
	}

	c.ApplyPaperTexture(0.25, 12345)
	c.ApplyVignette(0.4)

	rCorner, _, _, _ := c.PixelLinearAt(0, 0)
	if math.IsNaN(rCorner) || rCorner < 0 {
		t.Fatalf("expected valid non-negative corner pixel after paper texture + vignette, got %f", rCorner)
	}
}

func TestMosaicPrimitivesAndInsetPolygon(t *testing.T) {
	c := NewCanvas(64, 64)
	c.FillBackground(common.Black)
	c.SetBlendMode(BlendOver)

	amber := color.RGBA{R: 242, G: 172, B: 50, A: 255}
	teal := color.RGBA{R: 72, G: 168, B: 178, A: 255}

	c.DrawMosaicBrick(Pt(10, 20), Pt(40, 22), 5.0, 0.3, 1.0, amber, 1.0)
	c.DrawCurvedTessera([]Point{Pt(12, 42), Pt(28, 38), Pt(48, 44)}, 4.5, -0.25, 1.0, teal, 1.0)

	inset := InsetPolygon([]Point{Pt(10, 10), Pt(30, 10), Pt(20, 28)}, 1.5)
	if len(inset) != 3 {
		t.Fatalf("expected 3 vertices from InsetPolygon, got %d", len(inset))
	}

	img := c.ResolveToRGBA(nil)
	lit := 0
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			px := img.RGBAAt(x, y)
			if px.R > 0 || px.G > 0 || px.B > 0 {
				lit++
			}
		}
	}
	if lit < 120 {
		t.Fatalf("expected DrawMosaicBrick and DrawCurvedTessera to rasterize >120 pixels, got %d", lit)
	}

	c.DrawCloisonneCell(
		[]Point{Pt(8, 8), Pt(28, 8), Pt(28, 28), Pt(8, 28)},
		teal, common.Black, amber, common.Black, amber, common.White, 1.5,
	)
	c.DrawStainedGlassCell(
		[]Point{Pt(34, 8), Pt(56, 8), Pt(56, 28), Pt(34, 28)},
		common.Black, teal, amber, common.Black, 1.8,
	)
}

func TestFieldEngineAndSDF2D(t *testing.T) {
	// 1. Verify SRGB8ToLinear LUT matches SRGBToLinear exactly
	for i := 0; i < 256; i++ {
		got := SRGB8ToLinear(uint8(i))
		want := SRGBToLinear(float64(i) / 255.0)
		if math.Abs(got-want) > 1e-15 {
			t.Fatalf("SRGB8ToLinear(%d) = %v, want %v", i, got, want)
		}
	}

	// 2. Verify Jobard-Lefer Evenly-Spaced Streamlines
	cfg := DefaultStreamlineConfig(120, 120, 10.0)
	lines := TraceEvenStreamlines(cfg, func(x, y float64) Point {
		// Swirling vortex field around (60, 60)
		dx := x - 60.0
		dy := y - 60.0
		return Pt(-dy-0.2*dx, dx-0.2*dy)
	})
	if len(lines) < 5 {
		t.Fatalf("expected at least 5 evenly-spaced streamlines, got %d", len(lines))
	}

	// 3. Verify Marching Squares iso-contour extraction on a circle scalar field
	contours := ExtractMarchingSquares(32, 32, 100.0, 100.0, 30.0, func(x, y float64) float64 {
		return math.Hypot(x-50.0, y-50.0)
	})
	if len(contours) == 0 || len(contours[0]) < 16 {
		t.Fatalf("expected closed circular contour from ExtractMarchingSquares, got %d contours", len(contours))
	}

	// 4. Verify 2D SDF Boolean & Bevel/DropShadow rendering on a 160x160 canvas (triggers parallelRows)
	c := NewCanvas(160, 160)
	c.FillBackground(common.Black)
	c1 := SDFCircle(Pt(65, 80), 32)
	c2 := SDFPolygon([]Point{Pt(75, 50), Pt(125, 80), Pt(75, 110)})
	shape := SDFSmoothUnion(c1, c2, 8.0)

	c.RenderSDFDropShadow(shape, Pt(4, 5), 6.0, 1.0, color.RGBA{R: 20, G: 20, B: 30, A: 255}, 0.8)
	c.RenderSDFBevel(
		shape,
		color.RGBA{R: 40, G: 160, B: 180, A: 255},
		color.RGBA{R: 15, G: 60, B: 90, A: 255},
		color.RGBA{R: 220, G: 250, B: 255, A: 255},
		8.0,
		Pt(-0.7, -0.7),
		1.0,
	)
	c.RenderIsoBand(42.0, 48.0, 1.5, color.RGBA{R: 250, G: 200, B: 60, A: 255}, 0.9, func(x, y float64) float64 {
		return math.Hypot(x-80.0, y-80.0)
	})

	img := c.ResolveToRGBA(nil)
	if img.RGBAAt(65, 80).G < 80 {
		t.Fatalf("expected beveled SDF interior at (65,80) to be lit, got %+v", img.RGBAAt(65, 80))
	}
}


