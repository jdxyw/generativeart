package generativeart_test

import (
	"bytes"
	"image/color"
	"image/gif"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func TestRenderPosterAndToPosterPNG(t *testing.T) {
	cfg := generativeart.DefaultPosterConfig()
	cfg.Rows = 2
	cfg.Cols = 2
	cfg.PanelSize = 140
	cfg.Title = "MAZE VARIATIONS"
	cfg.DarkTheme = false
	cfg.Background = common.Azure
	cfg.Palette = common.Youthful

	poster := generativeart.RenderPoster(arts.NewMaze(16), cfg)
	if poster.Width() <= 280 || poster.Height() <= 280 {
		t.Fatalf("unexpected poster dimensions: %dx%d", poster.Width(), poster.Height())
	}

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "poster_test.png")
	c := generativeart.NewCanva(160, 160)
	c.SetSeed(2026)
	c.SetBackground(color.RGBA{R: 16, G: 20, B: 34, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewMaze(16))

	darkCfg := generativeart.DefaultPosterConfig()
	darkCfg.DarkTheme = true
	darkCfg.Title = "SINGLE MUSEUM PLATE"
	if err := c.ToPosterPNG(outPath, darkCfg); err != nil {
		t.Fatalf("ToPosterPNG failed: %v", err)
	}
	info, err := os.Stat(outPath)
	if err != nil || info.Size() == 0 {
		t.Fatalf("expected non-empty poster PNG, err=%v", err)
	}
}

func TestParameterSweepAndAnimatedPosterAndVectorPreservation(t *testing.T) {
	desc, ok := arts.Lookup("cloisonneenamel")
	if !ok {
		t.Fatalf("expected cloisonneenamel in registry")
	}

	// 1. 2D Parameter-Sweep Phase Diagram Poster + Vector SVG Preservation + Math/Greek Glyphs
	cfg := generativeart.DefaultPosterConfig()
	cfg.Rows = 2
	cfg.Cols = 2
	cfg.PanelSize = 120
	cfg.SeedBase = 20261104
	cfg.Title = "PHASE DIAGRAM [α, β, θ, π, φ, λ, σ, Ω = 100%]"

	sweepPoster := generativeart.RenderParameterSweepPoster(generativeart.ParameterSweepConfig{
		EngineName: desc.ID,
		Build:      desc.Build,
		BaseParams: desc.DefaultParams(),
		AxisX:      generativeart.ParameterSweepAxis{Name: "rings", Min: 3, Max: 5, IsInt: true},
		AxisY:      generativeart.ParameterSweepAxis{Name: "petals", Min: 6, Max: 10, IsInt: true},
	}, cfg)

	if sweepPoster.Width() <= 240 || sweepPoster.Height() <= 240 {
		t.Fatalf("unexpected sweep poster dimensions: %dx%d", sweepPoster.Width(), sweepPoster.Height())
	}
	// Verify sub-panel vector paths were preserved into the poster's RenderCanvas
	rc := sweepPoster.LastRenderCanvas()
	if rc == nil || rc.VectorOpCount() < 100 {
		t.Fatalf("expected sub-panel vector paths preserved in poster RenderCanvas, got %v", rc.VectorOpCount())
	}
	var svgBuf bytes.Buffer
	if err := sweepPoster.WriteSVG(&svgBuf); err != nil {
		t.Fatalf("WriteSVG on sweepPoster failed: %v", err)
	}
	if !strings.Contains(svgBuf.String(), "<path") {
		t.Fatalf("expected vector <path> elements in exported poster SVG")
	}

	// 2. Animated Multi-Panel Exhibition Poster GIF
	animCfg := generativeart.DefaultPosterConfig()
	animCfg.Rows = 1
	animCfg.Cols = 2
	animCfg.PanelSize = 120
	animCfg.DarkTheme = true
	animPoster := generativeart.RenderAnimatedPoster(arts.NewCloisonneEnamel(3, 6), animCfg, 3)
	if len(animPoster.Frames()) != 3 {
		t.Fatalf("expected 3 animated poster frames, got %d", len(animPoster.Frames()))
	}
	var gifBuf bytes.Buffer
	if err := animPoster.WriteGIF(&gifBuf, 10); err != nil {
		t.Fatalf("WriteGIF on animated poster failed: %v", err)
	}
	decoded, err := gif.DecodeAll(&gifBuf)
	if err != nil || len(decoded.Image) != 3 {
		t.Fatalf("expected 3 frames in decoded animated poster GIF, err=%v", err)
	}
}

