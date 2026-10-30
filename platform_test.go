package generativeart_test

import (
	"bytes"
	"image/color"
	"image/gif"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

func TestPNGRecipeMetadataRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	pngPath := filepath.Join(tmpDir, "mosaic_recipe.png")

	c := generativeart.NewCanva(160, 160)
	c.SetSeed(20261021)
	c.SetBackground(color.RGBA{R: 22, G: 26, B: 48, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewMosaicSun(20, 1.0))

	if err := c.ToPNG(pngPath); err != nil {
		t.Fatalf("ToPNG failed: %v", err)
	}

	// Verify standard image/png decoder still decodes the file cleanly
	f, err := os.Open(pngPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("standard png.Decode failed on PNG with embedded recipe: %v", err)
	}
	if img.Bounds().Dx() != 160 || img.Bounds().Dy() != 160 {
		t.Fatalf("unexpected image dimensions: %v", img.Bounds())
	}

	// Verify ReadPNGRecipe extracts the exact recipe metadata
	meta, err := generativeart.ReadPNGRecipe(pngPath)
	if err != nil {
		t.Fatalf("ReadPNGRecipe failed: %v", err)
	}
	if meta.Engine != "mosaicSun" || meta.Seed != 20261021 || meta.Width != 160 || meta.Height != 160 {
		t.Fatalf("unexpected recipe metadata: %+v", meta)
	}
	if ringSteps, ok := meta.Params["ringSteps"].(float64); !ok || int(ringSteps) != 20 {
		t.Fatalf("expected ringSteps=20 in recipe params, got %+v", meta.Params)
	}
}

func TestDualVectorSVGPDFExport(t *testing.T) {
	c := generativeart.NewCanva(160, 160)
	c.SetSeed(99)
	c.SetBackground(color.RGBA{R: 18, G: 14, B: 22, A: 255})
	c.FillBackground()
	c.Draw(arts.NewCloisonneEnamel(4, 8))

	rc := c.LastRenderCanvas()
	if rc == nil || rc.VectorOpCount() < 20 {
		t.Fatalf("expected recorded vector operations from CloisonneEnamel, got %v", rc)
	}

	var svgBuf bytes.Buffer
	if err := c.WriteSVG(&svgBuf); err != nil {
		t.Fatalf("WriteSVG failed: %v", err)
	}
	svgStr := svgBuf.String()
	if !strings.Contains(svgStr, "<svg") || !strings.Contains(svgStr, "<path") {
		t.Fatalf("expected valid SVG with <path> elements, got %d bytes", len(svgStr))
	}

	var pdfBuf bytes.Buffer
	if err := c.WritePDF(&pdfBuf); err != nil {
		t.Fatalf("WritePDF failed: %v", err)
	}
	if !bytes.HasPrefix(pdfBuf.Bytes(), []byte("%PDF-")) {
		t.Fatalf("expected valid PDF header, got %d bytes", pdfBuf.Len())
	}
}

func TestDrawTimelapseAndAnimatedGIF(t *testing.T) {
	c := generativeart.NewCanva(96, 96)
	c.SetSeed(777)
	c.SetBackground(color.RGBA{R: 12, G: 16, B: 28, A: 255})

	// Create a seamless 4-frame looping animation using Curl4DLoop
	anim := generativeart.NewAnimatedEngine(func(cv *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
		rc := cv.NewRenderCanvas()
		ne := cv.Noise()
		v := ne.Curl4DLoop(1.5, 2.5, phase, 1.0, 1e-3)
		cx := 48.0 + 18.0*math.Cos(phase*2.0*math.Pi) + v.X*2.0
		cy := 48.0 + 18.0*math.Sin(phase*2.0*math.Pi) + v.Y*2.0
		rc.FillPolygon([]render.Point{
			render.Pt(cx-12, cy-12),
			render.Pt(cx+12, cy-12),
			render.Pt(cx+12, cy+12),
			render.Pt(cx-12, cy+12),
		}, color.RGBA{R: 240, G: 180, B: 60, A: 255}, 1.0)
		cv.MergeRenderCanvas(rc)
	})

	c.DrawTimelapseFrames(anim, 4)
	if len(c.Frames()) != 4 {
		t.Fatalf("expected 4 recorded frames, got %d", len(c.Frames()))
	}

	var gifBuf bytes.Buffer
	if err := c.WriteGIF(&gifBuf, 6); err != nil {
		t.Fatalf("WriteGIF failed: %v", err)
	}

	decodedGIF, err := gif.DecodeAll(&gifBuf)
	if err != nil {
		t.Fatalf("gif.DecodeAll failed: %v", err)
	}
	if len(decodedGIF.Image) != 4 {
		t.Fatalf("expected 4 frames in decoded GIF, got %d", len(decodedGIF.Image))
	}
}

func TestPersistentHDRCanvaAndReplayPNG(t *testing.T) {
	// 1. Verify persistent HDR buffer across multiple Engine draws without 8-bit clamping
	c := generativeart.NewCanva(64, 64)
	c.SetBackground(common.Black)
	c.FillBackground()

	hdrPass := generativeart.NewAnimatedEngine(func(cv *generativeart.Canva, _, _ int, _ float64) {
		rc := cv.NewRenderCanvas()
		rc.SetBlendMode(render.BlendAdd)
		rc.SetPixelLinear(32, 32, 0.8, 0.5, 0.2, 1.0)
		// Add 0.8 again in linear light -> 1.6 (above 1.0!)
		rc.BlendPixelLinear(32, 32, 0.8, 0.5, 0.2, 1.0, render.BlendAdd)
		cv.MergeRenderCanvas(rc)
	})
	c.Draw(hdrPass)

	// Calling NewRenderCanvas() again before c.Img() must preserve the un-clamped 1.6 HDR value
	rc2 := c.NewRenderCanvas()
	rLin, _, _, _ := rc2.PixelLinearAt(32, 32)
	if rLin < 1.5 {
		t.Fatalf("expected persistent HDR buffer to retain linear R > 1.5 across passes, got %f", rLin)
	}

	// 2. Verify arts.ReplayPNG reconstructs and re-renders a saved PNG at a new resolution
	tmpDir := t.TempDir()
	srcPNG := filepath.Join(tmpDir, "source_rose.png")

	cRose := generativeart.NewCanva(120, 120)
	cRose.SetSeed(2026)
	cRose.SetBackground(color.RGBA{R: 12, G: 11, B: 18, A: 255})
	cRose.FillBackground()
	cRose.Draw(arts.NewStainedRoseWindow(8, 4))
	if err := cRose.ToPNG(srcPNG); err != nil {
		t.Fatalf("ToPNG failed: %v", err)
	}

	replayed, meta, err := arts.ReplayPNG(srcPNG, 240, 240)
	if err != nil {
		t.Fatalf("ReplayPNG failed: %v", err)
	}
	if meta.Engine != "stainedRoseWindow" || replayed.Width() != 240 || replayed.Height() != 240 {
		t.Fatalf("unexpected replayed canvas or metadata: w=%d h=%d meta=%+v", replayed.Width(), replayed.Height(), meta)
	}
}

func TestCanvaPostEffectPipelineAndRecipe(t *testing.T) {
	tmpDir := t.TempDir()
	outPNG := filepath.Join(tmpDir, "post_pipeline.png")

	c := generativeart.NewCanva(120, 120)
	c.SetSeed(20261030)
	c.SetBackground(color.RGBA{R: 18, G: 22, B: 34, A: 255})
	c.FillBackground()
	c.AddPostEffect(
		render.LuminanceReliefEffect(0.6, -135, 38, 0.35),
		render.ChromaticAberrationEffect(0.8),
		render.VignetteEffect(0.25),
	)
	if len(c.PostEffects()) != 3 {
		t.Fatalf("expected 3 registered post-effects, got %d", len(c.PostEffects()))
	}

	c.Draw(arts.NewMosaicSun(14, 1.0))
	if err := c.ToPNG(outPNG); err != nil {
		t.Fatalf("ToPNG failed: %v", err)
	}

	meta, err := generativeart.ReadPNGRecipe(outPNG)
	if err != nil {
		t.Fatalf("ReadPNGRecipe failed: %v", err)
	}
	if len(meta.PostEffects) != 3 {
		t.Fatalf("expected 3 post_effects recorded in PNG recipe metadata, got %+v", meta.PostEffects)
	}

	c.ClearPostEffects()
	if len(c.PostEffects()) != 0 {
		t.Fatalf("expected 0 post-effects after ClearPostEffects")
	}
}


