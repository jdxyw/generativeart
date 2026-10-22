package generativeart

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"reflect"
	"strings"
	"sync"

	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

func posterEngineName(e Engine) string {
	if e == nil {
		return "Custom"
	}
	t := reflect.TypeOf(e)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Name() != "" {
		return t.Name()
	}
	return "Custom"
}

// PosterPanelSpec describes a single panel in a multi-panel exhibition poster or contact sheet.
type PosterPanelSpec struct {
	Engine     Engine
	Caption    string
	Seed       int64
	Background color.RGBA
	Palette    []color.RGBA
	PreDrawn   *image.RGBA
}

// PosterConfig configures the layout, typography, framing, and theme of an exhibition poster.
type PosterConfig struct {
	Title       string
	Subtitle    string
	RecipeLabel string
	HeaderTag   string
	Rows        int
	Cols        int
	PanelSize   int
	SeedBase    int64
	DarkTheme   bool
	Background  color.RGBA
	Palette     []color.RGBA
}

// DefaultPosterConfig returns a balanced museum poster configuration.
func DefaultPosterConfig() PosterConfig {
	return PosterConfig{
		Title:       "GENERATIVE ART EXHIBITION",
		Subtitle:    "ALGORITHMIC MORPHOGENESIS & COMPUTATIONAL AESTHETICS IN GO",
		RecipeLabel: "64-BIT LINEAR HDR ENGINE  ·  OKLAB PERCEPTUAL COLOR",
		HeaderTag:   "MUSEUM ARCHIVE EDITION  ·  GITHUB.COM/JDXYW/GENERATIVEART",
		Rows:        2,
		Cols:        2,
		PanelSize:   420,
		SeedBase:    20261022,
		DarkTheme:   false,
		Background:  color.RGBA{R: 16, G: 20, B: 34, A: 255},
		Palette:     common.Dunhuang,
	}
}

// RenderPoster renders either a 1x1 framed museum print or an Rows x Cols multi-seed
// variation contact sheet for engine e.
func RenderPoster(e Engine, cfg PosterConfig) *Canva {
	if cfg.Rows <= 0 {
		cfg.Rows = 2
	}
	if cfg.Cols <= 0 {
		cfg.Cols = 2
	}
	if cfg.PanelSize <= 0 {
		cfg.PanelSize = 420
	}
	if cfg.SeedBase == 0 {
		cfg.SeedBase = 20261022
	}
	if cfg.Background.A == 0 {
		cfg.Background = color.RGBA{R: 16, G: 20, B: 34, A: 255}
	}
	if len(cfg.Palette) == 0 {
		cfg.Palette = common.Dunhuang
	}
	if cfg.Title == "" && e != nil {
		cfg.Title = strings.ToUpper(posterEngineName(e))
	}
	if cfg.RecipeLabel == "" && e != nil {
		cfg.RecipeLabel = fmt.Sprintf("ENGINE: arts.New%s  ·  BASE SEED: %d  ·  PANEL: %dx%d",
			strings.Title(posterEngineName(e)), cfg.SeedBase, cfg.PanelSize, cfg.PanelSize)
	}

	total := cfg.Rows * cfg.Cols
	panels := make([]PosterPanelSpec, total)
	for i := 0; i < total; i++ {
		pSeed := cfg.SeedBase + int64(i)*137
		caption := fmt.Sprintf("PLATE %02d  ·  SEED %d", i+1, pSeed)
		if total == 1 {
			caption = fmt.Sprintf("ARCHIVE PLATE  ·  SEED %d", pSeed)
		}
		panels[i] = PosterPanelSpec{
			Engine:     e,
			Caption:    caption,
			Seed:       pSeed,
			Background: cfg.Background,
			Palette:    cfg.Palette,
		}
	}
	return RenderGalleryPoster(panels, cfg)
}

// ToPosterPNG wraps the current canvas in a single-panel museum exhibition frame and saves it to fpath.
func (c *Canva) ToPosterPNG(fpath string, cfg PosterConfig) error {
	cfg.Rows = 1
	cfg.Cols = 1
	if cfg.PanelSize <= 0 {
		cfg.PanelSize = c.Width()
	}
	if cfg.SeedBase == 0 {
		cfg.SeedBase = c.Seed()
	}
	if cfg.Background.A == 0 {
		cfg.Background = c. opts.background
	}
	if len(cfg.Palette) == 0 && len(c.opts.colorSchema) > 0 {
		cfg.Palette = c.opts.colorSchema
	}
	if cfg.Title == "" && c.lastRecipe != nil && c.lastRecipe.Engine != "" {
		cfg.Title = strings.ToUpper(c.lastRecipe.Engine)
	}
	panel := PosterPanelSpec{
		Caption:    fmt.Sprintf("ARCHIVE PLATE  ·  SEED %d", cfg.SeedBase),
		Seed:       cfg.SeedBase,
		Background: cfg.Background,
		Palette:    cfg.Palette,
		PreDrawn:   c.Img(),
	}
	posterCanva := RenderGalleryPoster([]PosterPanelSpec{panel}, cfg)
	return posterCanva.ToPNG(fpath)
}

// RenderGalleryPoster renders a grid of PosterPanelSpec entries onto an archival museum poster Canva.
func RenderGalleryPoster(panels []PosterPanelSpec, cfg PosterConfig) *Canva {
	if cfg.Rows <= 0 {
		cfg.Rows = 1
	}
	if cfg.Cols <= 0 {
		cfg.Cols = 1
	}
	if cfg.PanelSize < 120 {
		cfg.PanelSize = 420
	}
	if cfg.HeaderTag == "" {
		cfg.HeaderTag = "GENERATIVEART EXHIBITION ARCHIVE  ·  64-BIT LINEAR HDR"
	}
	if cfg.Title == "" {
		cfg.Title = "GENERATIVE ART EXHIBITION"
	}
	if cfg.Subtitle == "" {
		cfg.Subtitle = "COMPUTATIONAL GEOMETRY, TOPOLOGICAL FIELDS & PERCEPTUAL COLOR IN GO"
	}
	if cfg.RecipeLabel == "" {
		cfg.RecipeLabel = fmt.Sprintf("GRID %dx%d  ·  PANEL %dx%d PX  ·  SEED %d",
			cfg.Rows, cfg.Cols, cfg.PanelSize, cfg.PanelSize, cfg.SeedBase)
	}
	if len(cfg.Palette) == 0 {
		if len(panels) > 0 && len(panels[0].Palette) > 0 {
			cfg.Palette = panels[0].Palette
		} else {
			cfg.Palette = common.Dunhuang
		}
	}

	ps := float64(cfg.PanelSize)
	scale := ps / 420.0
	if scale < 0.55 {
		scale = 0.55
	}

	outerPad := int(math.Round(56.0 * scale))
	gridGapX := int(math.Round(34.0 * scale))
	captionH := int(math.Round(24.0 * scale))
	gridGapY := gridGapX + captionH
	headerH := int(math.Round(56.0 * scale))
	footerH := int(math.Round(132.0 * scale))

	totalW := outerPad*2 + cfg.Cols*cfg.PanelSize + (cfg.Cols-1)*gridGapX
	totalH := outerPad + headerH + cfg.Rows*cfg.PanelSize + (cfg.Rows-1)*gridGapY + captionH + footerH + outerPad/2

	// Render all panel images concurrently
	totalPanels := cfg.Rows * cfg.Cols
	renderedPanels := make([]*image.RGBA, totalPanels)
	var wg sync.WaitGroup
	for idx := 0; idx < totalPanels && idx < len(panels); idx++ {
		wg.Add(1)
		go func(i int, spec PosterPanelSpec) {
			defer wg.Done()
			if spec.PreDrawn != nil && spec.PreDrawn.Bounds().Dx() == cfg.PanelSize && spec.PreDrawn.Bounds().Dy() == cfg.PanelSize {
				renderedPanels[i] = spec.PreDrawn
				return
			}
			sub := NewCanva(cfg.PanelSize, cfg.PanelSize)
			seed := spec.Seed
			if seed == 0 {
				seed = cfg.SeedBase + int64(i)*137
			}
			sub.SetSeed(seed)
			bg := spec.Background
			if bg.A == 0 {
				bg = cfg.Background
			}
			if bg.A == 0 {
				bg = color.RGBA{R: 14, G: 18, B: 28, A: 255}
			}
			sub.SetBackground(bg)
			sub.FillBackground()
			pal := spec.Palette
			if len(pal) == 0 {
				pal = cfg.Palette
			}
			if len(pal) > 0 {
				sub.SetColorSchema(pal)
			}
			if spec.Engine != nil {
				sub.Draw(spec.Engine)
			}
			renderedPanels[i] = sub.Img()
		}(idx, panels[idx])
	}
	wg.Wait()

	// Theme color palette (Museum Ivory Parchment vs Obsidian Gallery Dark)
	var (
		matBg         color.RGBA
		matRecess     color.RGBA
		bevelLight    color.RGBA
		bevelShadow   color.RGBA
		goldFillet    color.RGBA
		ruleLine      color.RGBA
		inkPrimary    color.RGBA
		inkSecondary  color.RGBA
		inkMuted      color.RGBA
		sealVermilion color.RGBA
		sealInk       color.RGBA
	)

	if cfg.DarkTheme {
		matBg = color.RGBA{R: 12, G: 15, B: 24, A: 255}
		matRecess = color.RGBA{R: 7, G: 9, B: 15, A: 255}
		bevelLight = color.RGBA{R: 52, G: 62, B: 86, A: 255}
		bevelShadow = color.RGBA{R: 3, G: 4, B: 8, A: 255}
		goldFillet = color.RGBA{R: 228, G: 182, B: 78, A: 255}
		ruleLine = color.RGBA{R: 68, G: 78, B: 104, A: 255}
		inkPrimary = color.RGBA{R: 244, G: 240, B: 230, A: 255}
		inkSecondary = color.RGBA{R: 224, G: 182, B: 86, A: 255}
		inkMuted = color.RGBA{R: 146, G: 158, B: 182, A: 255}
		sealVermilion = color.RGBA{R: 198, G: 48, B: 38, A: 255}
		sealInk = color.RGBA{R: 250, G: 242, B: 226, A: 255}
	} else {
		matBg = color.RGBA{R: 245, G: 240, B: 230, A: 255}
		matRecess = color.RGBA{R: 232, G: 225, B: 210, A: 255}
		bevelLight = color.RGBA{R: 255, G: 253, B: 248, A: 255}
		bevelShadow = color.RGBA{R: 186, G: 176, B: 158, A: 255}
		goldFillet = color.RGBA{R: 176, G: 132, B: 54, A: 255}
		ruleLine = color.RGBA{R: 194, G: 184, B: 166, A: 255}
		inkPrimary = color.RGBA{R: 26, G: 28, B: 36, A: 255}
		inkSecondary = color.RGBA{R: 148, G: 52, B: 38, A: 255}
		inkMuted = color.RGBA{R: 98, G: 94, B: 86, A: 255}
		sealVermilion = color.RGBA{R: 192, G: 44, B: 34, A: 255}
		sealInk = color.RGBA{R: 248, G: 242, B: 232, A: 255}
	}

	poster := NewCanva(totalW, totalH)
	poster.SetSeed(cfg.SeedBase)
	poster.SetBackground(matBg)
	poster.FillBackground()

	rc := poster.NewRenderCanvas()
	rc.FillBackground(matBg)
	rc.SetBlendMode(render.BlendOver)

	fw, fh := float64(totalW), float64(totalH)

	// 1. Outer Double-Fillet Passe-Partout Border Frame
	bMargin := float64(outerPad) * 0.44
	drawRectStroke(rc, bMargin, bMargin, fw-2*bMargin, fh-2*bMargin, 1.6*scale, goldFillet, 0.85)
	bMarginIn := bMargin + 5.0*scale
	drawRectStroke(rc, bMarginIn, bMarginIn, fw-2*bMarginIn, fh-2*bMarginIn, 0.85*scale, ruleLine, 0.72)

	// Corner museum registration ticks
	tickL := 10.0 * scale
	corners := [][4]float64{
		{bMargin, bMargin, 1, 1},
		{fw - bMargin, bMargin, -1, 1},
		{fw - bMargin, fh - bMargin, -1, -1},
		{bMargin, fh - bMargin, 1, -1},
	}
	for _, c := range corners {
		cx, cy, dx, dy := c[0], c[1], c[2], c[3]
		p := &canvas.Path{}
		p.MoveTo(cx+dx*tickL, cy+dy*2.5*scale)
		p.LineTo(cx+dx*2.5*scale, cy+dy*2.5*scale)
		p.LineTo(cx+dx*2.5*scale, cy+dy*tickL)
		rc.StrokePath(p, 1.4*scale, canvas.RoundCap, canvas.RoundJoin, goldFillet, 0.92)
	}

	// 2. Top Header Exhibition Strip
	headerY := float64(outerPad) + 14.0*scale
	drawPosterText(rc, float64(outerPad), headerY, 9.5*scale, 1.15*scale, 0.22, cfg.HeaderTag, inkMuted, 0.90)

	editionStr := fmt.Sprintf("SERIES %dX%d", cfg.Rows, cfg.Cols)
	edW := measurePosterText(editionStr, 9.5*scale, 0.22)
	drawPosterText(rc, fw-float64(outerPad)-edW, headerY, 9.5*scale, 1.15*scale, 0.22, editionStr, goldFillet, 0.92)

	ruleY := float64(outerPad) + float64(headerH) - 16.0*scale
	drawLine(rc, float64(outerPad), ruleY, fw-float64(outerPad), ruleY, 1.0*scale, ruleLine, 0.75)

	// 3. Render Each Artwork Panel with Recessed Mat Bevel, SDF Shadow & Plate Caption
	bevelW := 7.0 * scale
	filletPad := 11.0 * scale
	for r := 0; r < cfg.Rows; r++ {
		for c := 0; c < cfg.Cols; c++ {
			idx := r*cfg.Cols + c
			px := outerPad + c*(cfg.PanelSize+gridGapX)
			py := outerPad + headerH + r*(cfg.PanelSize+gridGapY)
			fpx, fpy := float64(px), float64(py)
			fps := float64(cfg.PanelSize)

			// Outer gilded fillet around the mat opening
			drawRectStroke(rc, fpx-filletPad, fpy-filletPad, fps+2*filletPad, fps+2*filletPad, 1.05*scale, goldFillet, 0.78)

			// 45-degree Passe-Partout Mat Bevel (Top & Left shadowed, Bottom & Right illuminated)
			topBevel := []render.Point{
				render.Pt(fpx-bevelW, fpy-bevelW),
				render.Pt(fpx+fps+bevelW, fpy-bevelW),
				render.Pt(fpx+fps, fpy),
				render.Pt(fpx, fpy),
			}
			leftBevel := []render.Point{
				render.Pt(fpx-bevelW, fpy-bevelW),
				render.Pt(fpx, fpy),
				render.Pt(fpx, fpy+fps),
				render.Pt(fpx-bevelW, fpy+fps+bevelW),
			}
			bottomBevel := []render.Point{
				render.Pt(fpx, fpy+fps),
				render.Pt(fpx+fps, fpy+fps),
				render.Pt(fpx+fps+bevelW, fpy+fps+bevelW),
				render.Pt(fpx-bevelW, fpy+fps+bevelW),
			}
			rightBevel := []render.Point{
				render.Pt(fpx+fps, fpy),
				render.Pt(fpx+fps+bevelW, fpy-bevelW),
				render.Pt(fpx+fps+bevelW, fpy+fps+bevelW),
				render.Pt(fpx+fps, fpy+fps),
			}
			rc.FillPolygon(topBevel, bevelShadow, 0.92)
			rc.FillPolygon(leftBevel, render.LerpOklab(bevelShadow, matRecess, 0.25), 0.92)
			rc.FillPolygon(bottomBevel, bevelLight, 0.92)
			rc.FillPolygon(rightBevel, render.LerpOklab(bevelLight, matRecess, 0.25), 0.92)

			// Blit rendered artwork pixels at 1:1 native resolution
			if idx < len(renderedPanels) && renderedPanels[idx] != nil {
				artImg := renderedPanels[idx]
				for y := 0; y < cfg.PanelSize; y++ {
					srcOff := y * artImg.Stride
					for x := 0; x < cfg.PanelSize; x++ {
						cr := artImg.Pix[srcOff+x*4+0]
						cg := artImg.Pix[srcOff+x*4+1]
						cb := artImg.Pix[srcOff+x*4+2]
						rc.SetPixelLinear(px+x, py+y,
							render.SRGB8ToLinear(cr),
							render.SRGB8ToLinear(cg),
							render.SRGB8ToLinear(cb),
							1.0)
					}
				}
			}

			// Inner crisp shadow rim along top/left of artwork opening
			drawRectStroke(rc, fpx, fpy, fps, fps, 1.1*scale, bevelShadow, 0.65)

			// Sub-panel plate caption
			if idx < len(panels) && panels[idx].Caption != "" {
				capY := fpy + fps + bevelW + 9.5*scale
				drawPosterText(rc, fpx, capY, 8.8*scale, 1.08*scale, 0.18, panels[idx].Caption, inkMuted, 0.88)
			}
		}
	}

	// 4. Bottom Museum Nameplate, Typography & Beveled Palette Swatch Bar
	footerTopY := float64(outerPad + headerH + cfg.Rows*cfg.PanelSize + (cfg.Rows-1)*gridGapY + captionH + 10)
	drawLine(rc, float64(outerPad), footerTopY, fw-float64(outerPad), footerTopY, 1.1*scale, ruleLine, 0.80)
	drawLine(rc, float64(outerPad), footerTopY+3.5*scale, float64(outerPad)+96.0*scale, footerTopY+3.5*scale, 2.2*scale, goldFillet, 0.95)

	titleY := footerTopY + 18.0*scale
	drawPosterText(rc, float64(outerPad), titleY, 21.0*scale, 2.35*scale, 0.16, cfg.Title, inkPrimary, 0.98)

	subY := titleY + 31.0*scale
	drawPosterText(rc, float64(outerPad), subY, 10.5*scale, 1.25*scale, 0.18, cfg.Subtitle, inkSecondary, 0.92)

	recipeY := subY + 21.0*scale
	drawPosterText(rc, float64(outerPad), recipeY, 8.8*scale, 1.05*scale, 0.14, cfg.RecipeLabel, inkMuted, 0.88)

	// Right-aligned 3D Beveled Palette Swatches + Vermilion Exhibition Seal
	sealSize := 42.0 * scale
	sealCX := fw - float64(outerPad) - sealSize*0.5
	sealCY := footerTopY + 46.0*scale
	drawVermilionSeal(rc, sealCX, sealCY, sealSize, scale, sealVermilion, sealInk)

	// Draw up to 6 circular 3D beveled palette swatches to the left of the seal
	maxSwatches := len(cfg.Palette)
	if maxSwatches > 6 {
		maxSwatches = 6
	}
	swatchR := 12.5 * scale
	swatchSpacing := 31.0 * scale
	swatchRightX := sealCX - sealSize*0.5 - 24.0*scale - swatchR
	swatchCY := sealCY - 5.0*scale

	for i := 0; i < maxSwatches; i++ {
		col := cfg.Palette[i]
		cx := swatchRightX - float64(maxSwatches-1-i)*swatchSpacing
		orbSDF := render.SDFCircle(render.Pt(cx, swatchCY), swatchR)
		rc.RenderSDFDropShadow(orbSDF, render.Pt(1.5*scale, 2.2*scale), 4.5*scale, 0.5*scale, bevelShadow, 0.55)
		darkRim := render.LerpOklab(col, color.RGBA{8, 10, 16, 255}, 0.42)
		specCol := render.LerpOklab(col, color.RGBA{255, 252, 240, 255}, 0.65)
		rc.RenderSDFBevel(orbSDF, col, darkRim, specCol, swatchR*0.38, render.Pt(-0.65, -0.75), 1.0)

		// Gilded ring around each swatch
		ringPath := &canvas.Path{}
		for s := 0; s <= 32; s++ {
			a := (float64(s) / 32.0) * 2.0 * math.Pi
			rx := cx + swatchR*math.Cos(a)
			ry := swatchCY + swatchR*math.Sin(a)
			if s == 0 {
				ringPath.MoveTo(rx, ry)
			} else {
				ringPath.LineTo(rx, ry)
			}
		}
		ringPath.Close()
		rc.StrokePath(ringPath, 1.0*scale, canvas.RoundCap, canvas.RoundJoin, goldFillet, 0.88)
	}

	if maxSwatches > 0 {
		palLabel := "CURATED OKLAB PALETTE"
		palW := measurePosterText(palLabel, 7.8*scale, 0.18)
		rightEdge := swatchRightX + swatchR
		drawPosterText(rc, rightEdge-palW, swatchCY+swatchR+8.5*scale, 7.8*scale, 0.95*scale, 0.18, palLabel, inkMuted, 0.84)
	}

	// Subtle museum cotton-rag paper texture across the mat
	texStrength := 0.038
	if cfg.DarkTheme {
		texStrength = 0.022
	}
	rc.ApplyPaperTexture(texStrength, cfg.SeedBase+99)

	poster.MergeRenderCanvas(rc)
	return poster
}

func drawRectStroke(rc *render.Canvas, x, y, w, h, strokeW float64, col color.RGBA, alpha float64) {
	p := &canvas.Path{}
	p.MoveTo(x, y)
	p.LineTo(x+w, y)
	p.LineTo(x+w, y+h)
	p.LineTo(x, y+h)
	p.Close()
	rc.StrokePath(p, strokeW, canvas.SquareCap, canvas.MiterJoin, col, alpha)
}

func drawLine(rc *render.Canvas, x0, y0, x1, y1, strokeW float64, col color.RGBA, alpha float64) {
	p := &canvas.Path{}
	p.MoveTo(x0, y0)
	p.LineTo(x1, y1)
	rc.StrokePath(p, strokeW, canvas.RoundCap, canvas.RoundJoin, col, alpha)
}

// drawVermilionSeal renders a traditional square vermilion seal stamp (朱文印章) with seal-script geometry.
func drawVermilionSeal(rc *render.Canvas, cx, cy, size, scale float64, vermilion, ink color.RGBA) {
	half := size * 0.5
	boxSDF := render.SDFRoundedBox(render.Pt(cx, cy), render.Pt(half, half), 3.2*scale)
	rc.RenderSDFDropShadow(boxSDF, render.Pt(1.5*scale, 2.0*scale), 4.0*scale, 0.5*scale, color.RGBA{0, 0, 0, 255}, 0.35)

	pts := []render.Point{
		render.Pt(cx-half, cy-half),
		render.Pt(cx+half, cy-half),
		render.Pt(cx+half, cy+half),
		render.Pt(cx-half, cy+half),
	}
	rc.FillPolygon(pts, vermilion, 0.95)
	drawRectStroke(rc, cx-half+2.2*scale, cy-half+2.2*scale, size-4.4*scale, size-4.4*scale, 1.3*scale, ink, 0.92)

	// Internal symmetrical Seal-Script (篆书) labyrinth motif inside the stamp
	inH := half - 5.5*scale
	sealPath := &canvas.Path{}
	// Vertical center pillar + left/right seal chambers
	sealPath.MoveTo(cx, cy-inH)
	sealPath.LineTo(cx, cy+inH)
	sealPath.MoveTo(cx-inH, cy)
	sealPath.LineTo(cx+inH, cy)

	// Upper-left & upper-right seal-script scrolls
	sealPath.MoveTo(cx-inH*0.78, cy-inH*0.75)
	sealPath.LineTo(cx-inH*0.25, cy-inH*0.75)
	sealPath.LineTo(cx-inH*0.25, cy-inH*0.28)
	sealPath.LineTo(cx-inH*0.78, cy-inH*0.28)

	sealPath.MoveTo(cx+inH*0.25, cy-inH*0.75)
	sealPath.LineTo(cx+inH*0.78, cy-inH*0.75)
	sealPath.LineTo(cx+inH*0.78, cy-inH*0.28)
	sealPath.LineTo(cx+inH*0.25, cy-inH*0.28)

	// Lower-left & lower-right seal-script hooks
	sealPath.MoveTo(cx-inH*0.78, cy+inH*0.28)
	sealPath.LineTo(cx-inH*0.25, cy+inH*0.28)
	sealPath.LineTo(cx-inH*0.25, cy+inH*0.75)

	sealPath.MoveTo(cx+inH*0.78, cy+inH*0.28)
	sealPath.LineTo(cx+inH*0.25, cy+inH*0.28)
	sealPath.LineTo(cx+inH*0.25, cy+inH*0.75)

	rc.StrokePath(sealPath, 1.45*scale, canvas.RoundCap, canvas.RoundJoin, ink, 0.92)
}

// measurePosterText returns the horizontal advance width of text in pixels.
func measurePosterText(text string, fontH, tracking float64) float64 {
	var total float64
	runes := []rune(text)
	for i, r := range runes {
		isLower := r >= 'a' && r <= 'z'
		upperR := r
		if isLower {
			upperR = r - ('a' - 'A')
		}
		wRatio := glyphAdvanceRatio(upperR)
		if isLower {
			wRatio *= 0.82
		}
		total += fontH * wRatio
		if i+1 < len(runes) {
			total += fontH * tracking
		}
	}
	return total
}

// drawPosterText renders clean geometric engraver typography (Full Caps + Small Caps for lowercase)
// at top-left position (x, y) with height fontH and stroke width strokeW.
func drawPosterText(rc *render.Canvas, x, y, fontH, strokeW, tracking float64, text string, col color.RGBA, alpha float64) {
	path := &canvas.Path{}
	curX := x
	for _, r := range text {
		isLower := r >= 'a' && r <= 'z'
		upperR := r
		glyphH := fontH
		glyphY := y
		if isLower {
			upperR = r - ('a' - 'A')
			glyphH = fontH * 0.78
			glyphY = y + (fontH - glyphH)
		}
		adv := glyphAdvanceRatio(upperR) * glyphH
		appendEngraverGlyph(path, upperR, curX, glyphY, adv, glyphH)
		curX += adv + fontH*tracking
	}
	rc.StrokePath(path, strokeW, canvas.RoundCap, canvas.RoundJoin, col, alpha)
}

func glyphAdvanceRatio(r rune) float64 {
	switch r {
	case ' ':
		return 0.42
	case 'I', '1', ':', '.', ',', '\'', '|':
		return 0.28
	case 'J', 'L', 'F', 'T', 'E', 'Z', '7', '(', ')':
		return 0.56
	case 'M', 'W':
		return 0.78
	case 'O', 'Q', 'G', 'C', 'D', 'H', 'N', 'U', 'A', 'V', 'X', 'Y', 'K', 'R', 'B', 'P', 'S':
		return 0.64
	case '·', '-', '+', '=', '/':
		return 0.45
	default:
		return 0.60
	}
}

func addEllipseArc(p *canvas.Path, cx, cy, rx, ry, degStart, degEnd float64, moveFirst bool) {
	const steps = 16
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		rad := (degStart + (degEnd-degStart)*t) * math.Pi / 180.0
		px := cx + rx*math.Cos(rad)
		py := cy + ry*math.Sin(rad)
		if i == 0 && moveFirst {
			p.MoveTo(px, py)
		} else {
			p.LineTo(px, py)
		}
	}
}

// appendEngraverGlyph appends clean geometric strokes for uppercase ASCII, digits, and punctuation
// inside the bounding box [x, x+w] x [y, y+h] (where y is top and y+h is baseline).
func appendEngraverGlyph(p *canvas.Path, r rune, x, y, w, h float64) {
	xm := x + 0.5*w
	ym := y + 0.5*h
	rx := 0.5 * w
	ry := 0.5 * h

	switch r {
	case ' ':
		return
	case '·':
		addEllipseArc(p, xm, ym, w*0.12, w*0.12, 0, 360, true)
	case '.':
		addEllipseArc(p, xm, y+h*0.92, h*0.04, h*0.04, 0, 360, true)
	case ',':
		p.MoveTo(xm, y+h*0.88)
		p.LineTo(xm-w*0.25, y+h*1.08)
	case ':':
		addEllipseArc(p, xm, y+h*0.32, h*0.04, h*0.04, 0, 360, true)
		addEllipseArc(p, xm, y+h*0.78, h*0.04, h*0.04, 0, 360, true)
	case '-':
		p.MoveTo(x+w*0.12, ym)
		p.LineTo(x+w*0.88, ym)
	case '_':
		p.MoveTo(x, y+h)
		p.LineTo(x+w, y+h)
	case '/':
		p.MoveTo(x+w*0.85, y)
		p.LineTo(x+w*0.15, y+h)
	case '(':
		addEllipseArc(p, x+w*0.75, ym, w*0.55, ry, 115, 245, true)
	case ')':
		addEllipseArc(p, x+w*0.25, ym, w*0.55, ry, -65, 65, true)
	case '#':
		p.MoveTo(x+w*0.32, y+h*0.1)
		p.LineTo(x+w*0.22, y+h*0.9)
		p.MoveTo(x+w*0.78, y+h*0.1)
		p.LineTo(x+w*0.68, y+h*0.9)
		p.MoveTo(x+w*0.08, y+h*0.36)
		p.LineTo(x+w*0.92, y+h*0.36)
		p.MoveTo(x+w*0.08, y+h*0.64)
		p.LineTo(x+w*0.92, y+h*0.64)
	case '+':
		p.MoveTo(xm, y+h*0.22)
		p.LineTo(xm, y+h*0.78)
		p.MoveTo(x+w*0.12, ym)
		p.LineTo(x+w*0.88, ym)
	case '&':
		addEllipseArc(p, xm-w*0.06, y+h*0.24, w*0.34, h*0.22, 0, 360, true)
		addEllipseArc(p, xm-w*0.08, y+h*0.72, w*0.40, h*0.28, 30, 310, true)
		p.MoveTo(x+w*0.18, y+h*0.42)
		p.LineTo(x+w*0.95, y+h)
	case 'A':
		p.MoveTo(x, y+h)
		p.LineTo(xm, y)
		p.LineTo(x+w, y+h)
		p.MoveTo(x+w*0.20, y+h*0.64)
		p.LineTo(x+w*0.80, y+h*0.64)
	case 'B':
		p.MoveTo(x, y)
		p.LineTo(x, y+h)
		p.MoveTo(x, y)
		p.LineTo(x+w*0.56, y)
		addEllipseArc(p, x+w*0.56, y+h*0.24, w*0.38, h*0.24, -90, 90, false)
		p.LineTo(x, y+h*0.48)
		p.MoveTo(x, y+h*0.48)
		p.LineTo(x+w*0.60, y+h*0.48)
		addEllipseArc(p, x+w*0.60, y+h*0.74, w*0.40, h*0.26, -90, 90, false)
		p.LineTo(x, y+h)
	case 'C':
		addEllipseArc(p, xm, ym, rx, ry, 42, 318, true)
	case 'D':
		p.MoveTo(x, y)
		p.LineTo(x, y+h)
		p.MoveTo(x, y)
		p.LineTo(x+w*0.45, y)
		addEllipseArc(p, x+w*0.45, ym, w*0.55, ry, -90, 90, false)
		p.LineTo(x, y+h)
	case 'E':
		p.MoveTo(x+w, y)
		p.LineTo(x, y)
		p.LineTo(x, y+h)
		p.LineTo(x+w, y+h)
		p.MoveTo(x, ym)
		p.LineTo(x+w*0.82, ym)
	case 'F':
		p.MoveTo(x+w, y)
		p.LineTo(x, y)
		p.LineTo(x, y+h)
		p.MoveTo(x, ym)
		p.LineTo(x+w*0.80, ym)
	case 'G':
		addEllipseArc(p, xm, ym, rx, ry, 318, 0, true)
		p.LineTo(xm, ym)
	case 'H':
		p.MoveTo(x, y)
		p.LineTo(x, y+h)
		p.MoveTo(x+w, y)
		p.LineTo(x+w, y+h)
		p.MoveTo(x, ym)
		p.LineTo(x+w, ym)
	case 'I':
		p.MoveTo(xm, y)
		p.LineTo(xm, y+h)
		p.MoveTo(x+w*0.15, y)
		p.LineTo(x+w*0.85, y)
		p.MoveTo(x+w*0.15, y+h)
		p.LineTo(x+w*0.85, y+h)
	case 'J':
		p.MoveTo(x+w*0.88, y)
		p.LineTo(x+w*0.88, y+h*0.68)
		addEllipseArc(p, x+w*0.46, y+h*0.68, w*0.42, h*0.32, 0, 165, false)
	case 'K':
		p.MoveTo(x, y)
		p.LineTo(x, y+h)
		p.MoveTo(x+w, y)
		p.LineTo(x, y+h*0.54)
		p.MoveTo(x+w*0.30, y+h*0.38)
		p.LineTo(x+w, y+h)
	case 'L':
		p.MoveTo(x, y)
		p.LineTo(x, y+h)
		p.LineTo(x+w, y+h)
	case 'M':
		p.MoveTo(x, y+h)
		p.LineTo(x, y)
		p.LineTo(xm, y+h*0.62)
		p.LineTo(x+w, y)
		p.LineTo(x+w, y+h)
	case 'N':
		p.MoveTo(x, y+h)
		p.LineTo(x, y)
		p.LineTo(x+w, y+h)
		p.LineTo(x+w, y)
	case 'O':
		addEllipseArc(p, xm, ym, rx, ry, 0, 360, true)
	case 'P':
		p.MoveTo(x, y+h)
		p.LineTo(x, y)
		p.LineTo(x+w*0.56, y)
		addEllipseArc(p, x+w*0.56, y+h*0.26, w*0.42, h*0.26, -90, 90, false)
		p.LineTo(x, y+h*0.52)
	case 'Q':
		addEllipseArc(p, xm, ym, rx, ry, 0, 360, true)
		p.MoveTo(xm+w*0.08, ym+h*0.12)
		p.LineTo(x+w*1.02, y+h*1.04)
	case 'R':
		p.MoveTo(x, y+h)
		p.LineTo(x, y)
		p.LineTo(x+w*0.56, y)
		addEllipseArc(p, x+w*0.56, y+h*0.25, w*0.42, h*0.25, -90, 90, false)
		p.LineTo(x, y+h*0.50)
		p.MoveTo(x+w*0.48, y+h*0.50)
		p.LineTo(x+w, y+h)
	case 'S':
		addEllipseArc(p, xm, y+h*0.25, w*0.45, h*0.25, -30, -270, true)
		addEllipseArc(p, xm, y+h*0.75, w*0.47, h*0.25, -90, 155, false)
	case 'T':
		p.MoveTo(x, y)
		p.LineTo(x+w, y)
		p.MoveTo(xm, y)
		p.LineTo(xm, y+h)
	case 'U':
		p.MoveTo(x, y)
		p.LineTo(x, y+h*0.64)
		addEllipseArc(p, xm, y+h*0.64, rx, h*0.36, 180, 0, false)
		p.LineTo(x+w, y)
	case 'V':
		p.MoveTo(x, y)
		p.LineTo(xm, y+h)
		p.LineTo(x+w, y)
	case 'W':
		p.MoveTo(x, y)
		p.LineTo(x+w*0.24, y+h)
		p.LineTo(xm, y+h*0.38)
		p.LineTo(x+w*0.76, y+h)
		p.LineTo(x+w, y)
	case 'X':
		p.MoveTo(x, y)
		p.LineTo(x+w, y+h)
		p.MoveTo(x+w, y)
		p.LineTo(x, y+h)
	case 'Y':
		p.MoveTo(x, y)
		p.LineTo(xm, ym)
		p.LineTo(x+w, y)
		p.MoveTo(xm, ym)
		p.LineTo(xm, y+h)
	case 'Z':
		p.MoveTo(x, y)
		p.LineTo(x+w, y)
		p.LineTo(x, y+h)
		p.LineTo(x+w, y+h)
	case '0':
		addEllipseArc(p, xm, ym, rx, ry, 0, 360, true)
		p.MoveTo(x+w*0.22, y+h*0.78)
		p.LineTo(x+w*0.78, y+h*0.22)
	case '1':
		p.MoveTo(x+w*0.18, y+h*0.18)
		p.LineTo(xm, y)
		p.LineTo(xm, y+h)
		p.MoveTo(x+w*0.15, y+h)
		p.LineTo(x+w*0.85, y+h)
	case '2':
		addEllipseArc(p, xm, y+h*0.28, w*0.46, h*0.28, -165, 35, true)
		p.LineTo(x, y+h)
		p.LineTo(x+w, y+h)
	case '3':
		addEllipseArc(p, xm, y+h*0.25, w*0.44, h*0.25, -150, 90, true)
		addEllipseArc(p, xm, y+h*0.74, w*0.48, h*0.26, -90, 155, false)
	case '4':
		p.MoveTo(x+w*0.74, y+h)
		p.LineTo(x+w*0.74, y)
		p.LineTo(x, y+h*0.68)
		p.LineTo(x+w, y+h*0.68)
	case '5':
		p.MoveTo(x+w*0.92, y)
		p.LineTo(x+w*0.12, y)
		p.LineTo(x+w*0.08, y+h*0.44)
		addEllipseArc(p, xm, y+h*0.70, w*0.48, h*0.30, -120, 150, false)
	case '6':
		addEllipseArc(p, xm, y+h*0.68, w*0.46, h*0.32, 0, 360, true)
		addEllipseArc(p, xm+w*0.06, ym, w*0.52, ry, 180, 295, true)
	case '7':
		p.MoveTo(x, y)
		p.LineTo(x+w, y)
		p.LineTo(x+w*0.28, y+h)
	case '8':
		addEllipseArc(p, xm, y+h*0.24, w*0.42, h*0.24, 0, 360, true)
		addEllipseArc(p, xm, y+h*0.74, w*0.48, h*0.26, 0, 360, true)
	case '9':
		addEllipseArc(p, xm, y+h*0.32, w*0.46, h*0.32, 0, 360, true)
		addEllipseArc(p, xm-w*0.06, ym, w*0.52, ry, 0, 115, true)
	}
}
