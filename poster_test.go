package generativeart_test

import (
	"image/color"
	"os"
	"path/filepath"
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
