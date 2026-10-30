package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/render"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "list":
		runList()
	case "render":
		runRender(os.Args[2:])
	case "poster":
		runPoster(os.Args[2:])
	case "inspect":
		runInspect(os.Args[2:])
	case "replay":
		runReplay(os.Args[2:])
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  genart list")
	fmt.Println("  genart render -art <name> [-palette <name>] [-w 600] [-h 600] [-seed 20261021] [-frames 1] [-bloom 0.35] [-chromatic 0.8] [-relief 0.6] [-kuwahara 3] [-paper 0.3] [-vignette 0.25] [-deckle 0.04] [-tonemap 1.1] -out <file.(png|svg|pdf|gif|jpg)>")
	fmt.Println("  genart poster [-art <name> | -arts <a,b,c,d>] [-grid 2x2] [-panel 420] [-theme light|dark] [-title \"...\"] -out <file.(png|svg|pdf|jpg)>")
	fmt.Println("  genart inspect <file.png>")
	fmt.Println("  genart replay <input.png> [-w 2400] [-h 2400] [-bloom ...] [-relief ...] -out <file.(png|svg|pdf|jpg)>")
}

type postFlags struct {
	bloom      *float64
	chromatic  *float64
	relief     *float64
	kuwahara   *int
	contrast   *float64
	saturation *float64
	paper      *float64
	vignette   *float64
	deckle     *float64
	tonemap    *float64
}

func registerPostFlags(fs *flag.FlagSet) postFlags {
	return postFlags{
		bloom:      fs.Float64("bloom", 0, "HDR Gaussian bloom strength (e.g. 0.35)"),
		chromatic:  fs.Float64("chromatic", 0, "Radial optical chromatic aberration strength (e.g. 0.8)"),
		relief:     fs.Float64("relief", 0, "Luminance 2.5D relief & micro-SSAO shading strength (e.g. 0.6)"),
		kuwahara:   fs.Int("kuwahara", 0, "Anisotropic Kuwahara painterly filter radius in pixels (e.g. 3)"),
		contrast:   fs.Float64("contrast", 1.0, "Oklab perceptual lightness contrast (default 1.0)"),
		saturation: fs.Float64("saturation", 1.0, "Oklab perceptual chroma saturation multiplier (default 1.0)"),
		paper:      fs.Float64("paper", 0, "Procedural cotton paper texture intensity (e.g. 0.3)"),
		vignette:   fs.Float64("vignette", 0, "Smooth radial vignette strength (e.g. 0.25)"),
		deckle:     fs.Float64("deckle", 0, "Torn deckle-edged museum mat border fraction (e.g. 0.045)"),
		tonemap:    fs.Float64("tonemap", 0, "ACES filmic tone-mapping exposure (e.g. 1.05)"),
	}
}

func (pf postFlags) buildEffects(seed int64) []render.PostEffect {
	var effects []render.PostEffect
	if pf.kuwahara != nil && *pf.kuwahara > 0 {
		effects = append(effects, render.AnisotropicKuwaharaEffect(*pf.kuwahara, 8.0, 1.25))
	}
	if pf.relief != nil && *pf.relief > 0 {
		effects = append(effects, render.LuminanceReliefEffect(*pf.relief, -135.0, 38.0, 0.38))
	}
	if pf.bloom != nil && *pf.bloom > 0 {
		effects = append(effects, render.BloomEffect(0.62, 6.0, *pf.bloom))
	}
	if pf.chromatic != nil && math.Abs(*pf.chromatic) > 1e-5 {
		effects = append(effects, render.ChromaticAberrationEffect(*pf.chromatic))
	}
	cVal := 1.0
	sVal := 1.0
	if pf.contrast != nil {
		cVal = *pf.contrast
	}
	if pf.saturation != nil {
		sVal = *pf.saturation
	}
	if math.Abs(cVal-1.0) > 1e-3 || math.Abs(sVal-1.0) > 1e-3 {
		effects = append(effects, render.ColorGradeEffect(cVal, sVal, 225.0, 42.0, 0.0))
	}
	if pf.tonemap != nil && *pf.tonemap > 0 {
		effects = append(effects, render.ToneMapEffect(*pf.tonemap, 2.2))
	}
	if pf.paper != nil && *pf.paper > 0 {
		effects = append(effects, render.PaperTextureEffect(*pf.paper, seed))
	}
	if pf.vignette != nil && *pf.vignette > 0 {
		effects = append(effects, render.VignetteEffect(*pf.vignette))
	}
	if pf.deckle != nil && *pf.deckle > 0 {
		effects = append(effects, render.DeckleBorderEffect(*pf.deckle, 1.0, color.RGBA{R: 242, G: 236, B: 224, A: 255}, seed))
	}
	return effects
}

func runList() {
	descs := arts.ListDescriptors()
	sort.Slice(descs, func(i, j int) bool { return descs[i].ID < descs[j].ID })
	fmt.Printf("Available Art Generators (%d):\n", len(descs))
	for _, d := range descs {
		fmt.Printf("  - %-20s [%-6s] (default palette: %s)\n", d.ID, d.Category, d.DefaultPal)
	}

	pNames := arts.PaletteNames()
	fmt.Printf("\nAvailable Palettes (%d): %s\n", len(pNames), strings.Join(pNames, ", "))
}

func saveCanvaByExt(c *generativeart.Canva, eng generativeart.Engine, frames, delay int, outPath string, alreadyDrawn bool) error {
	ext := strings.ToLower(filepath.Ext(outPath))
	switch ext {
	case ".gif":
		nFrames := frames
		if nFrames < 2 {
			nFrames = 8
		}
		if eng != nil {
			c.DrawTimelapseFrames(eng, nFrames)
		}
		return c.ToGIF(outPath, delay)
	case ".svg":
		if !alreadyDrawn && eng != nil {
			c.Draw(eng)
		}
		return c.ToSVG(outPath)
	case ".pdf":
		if !alreadyDrawn && eng != nil {
			c.Draw(eng)
		}
		return c.ToPDF(outPath)
	case ".jpg", ".jpeg":
		if !alreadyDrawn && eng != nil {
			c.Draw(eng)
		}
		return c.ToJPEG(outPath)
	default:
		if !alreadyDrawn && eng != nil {
			c.Draw(eng)
		}
		return c.ToPNG(outPath)
	}
}

func runRender(args []string) {
	fs := flag.NewFlagSet("render", flag.ExitOnError)
	artName := fs.String("art", "mosaicsun", "Art generator name")
	palName := fs.String("palette", "", "Color palette name (optional override)")
	w := fs.Int("w", 600, "Canvas width in pixels")
	h := fs.Int("h", 600, "Canvas height in pixels")
	seed := fs.Int64("seed", 20261021, "Deterministic random seed")
	frames := fs.Int("frames", 1, "Number of frames (for .gif timelapse)")
	delay := fs.Int("delay", 8, "Frame delay in centiseconds (for .gif)")
	out := fs.String("out", "output.png", "Output file path (.png, .svg, .pdf, .gif, .jpg)")
	pf := registerPostFlags(fs)
	_ = fs.Parse(args)

	desc, ok := arts.Lookup(*artName)
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown art generator %q (run 'genart list' to see supported generators)\n", *artName)
		os.Exit(1)
	}

	c := generativeart.NewCanva(*w, *h)
	c.SetSeed(*seed)
	c.SetBackground(desc.DefaultBG)
	c.FillBackground()

	chosenPal := desc.DefaultPal
	if *palName != "" {
		chosenPal = *palName
	}
	if schema, exists := arts.PaletteByName(chosenPal); exists {
		c.SetColorSchema(schema)
	}
	if effects := pf.buildEffects(*seed); len(effects) > 0 {
		c.AddPostEffect(effects...)
	}

	eng := desc.Build(nil)
	if err := saveCanvaByExt(c, eng, *frames, *delay, *out, false); err != nil {
		fmt.Fprintf(os.Stderr, "render failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Rendered %s (%dx%d, seed=%d) -> %s\n", desc.ID, *w, *h, *seed, *out)
}

func runInspect(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: genart inspect <file.png>")
		os.Exit(1)
	}
	meta, err := generativeart.ReadPNGRecipe(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "inspect failed: %v\n", err)
		os.Exit(1)
	}
	pretty, _ := json.MarshalIndent(meta, "", "  ")
	fmt.Println(string(pretty))
}

func runReplay(args []string) {
	fs := flag.NewFlagSet("replay", flag.ExitOnError)
	w := fs.Int("w", 0, "Override width in pixels (0 keeps original width)")
	h := fs.Int("h", 0, "Override height in pixels (0 keeps original height)")
	out := fs.String("out", "replayed.png", "Output file path (.png, .svg, .pdf, .jpg)")
	pf := registerPostFlags(fs)
	// Allow input file before or after flags
	var inputPath string
	var flagArgs []string
	for i := 0; i < len(args); i++ {
		if !strings.HasPrefix(args[i], "-") && inputPath == "" {
			inputPath = args[i]
		} else {
			flagArgs = append(flagArgs, args[i])
		}
	}
	_ = fs.Parse(flagArgs)
	if inputPath == "" && fs.NArg() > 0 {
		inputPath = fs.Arg(0)
	}
	if inputPath == "" {
		fmt.Fprintln(os.Stderr, "usage: genart replay <input.png> [-w 2400] [-h 2400] -out <output.(png|svg|pdf|jpg)>")
		os.Exit(1)
	}

	c, meta, err := arts.ReplayPNG(inputPath, *w, *h)
	if err != nil {
		fmt.Fprintf(os.Stderr, "replay failed: %v\n", err)
		os.Exit(1)
	}
	if effects := pf.buildEffects(meta.Seed); len(effects) > 0 {
		c.ApplyPostEffects(effects...)
	}
	if err := saveCanvaByExt(c, nil, 1, 8, *out, true); err != nil {
		fmt.Fprintf(os.Stderr, "save failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Replayed %s (seed=%d) at %dx%d -> %s\n", meta.Engine, meta.Seed, c.Width(), c.Height(), *out)
}

func parseGridDim(gridStr string) (int, int) {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(gridStr)), "x")
	if len(parts) == 2 {
		r, err1 := strconv.Atoi(parts[0])
		c, err2 := strconv.Atoi(parts[1])
		if err1 == nil && err2 == nil && r > 0 && c > 0 {
			return r, c
		}
	}
	return 2, 2
}

func runPoster(args []string) {
	fs := flag.NewFlagSet("poster", flag.ExitOnError)
	artName := fs.String("art", "dunhuangapsara", "Single art generator name for multi-seed variation contact sheet")
	artsList := fs.String("arts", "", "Comma-separated list of art generator names for curated multi-engine poster")
	palName := fs.String("palette", "", "Optional palette override")
	gridStr := fs.String("grid", "2x2", "Grid layout RowsxCols (e.g. 1x1, 2x2, 3x3)")
	panelSize := fs.Int("panel", 420, "Width/Height in pixels of each artwork panel")
	theme := fs.String("theme", "light", "Poster mat theme: 'light' (Museum Ivory Parchment) or 'dark' (Obsidian Gallery Gold)")
	title := fs.String("title", "", "Custom poster main title")
	subtitle := fs.String("subtitle", "", "Custom poster subtitle")
	seed := fs.Int64("seed", 20261022, "Base deterministic random seed")
	out := fs.String("out", "poster.png", "Output file path (.png, .svg, .pdf, .jpg)")
	_ = fs.Parse(args)

	rows, cols := parseGridDim(*gridStr)
	cfg := generativeart.DefaultPosterConfig()
	cfg.Rows = rows
	cfg.Cols = cols
	cfg.PanelSize = *panelSize
	cfg.SeedBase = *seed
	cfg.DarkTheme = strings.EqualFold(strings.TrimSpace(*theme), "dark")

	var posterCanva *generativeart.Canva

	if strings.TrimSpace(*artsList) != "" {
		rawNames := strings.Split(*artsList, ",")
		total := rows * cols
		panels := make([]generativeart.PosterPanelSpec, total)
		upperNames := make([]string, 0, total)
		for i := 0; i < total; i++ {
			name := strings.TrimSpace(rawNames[i%len(rawNames)])
			desc, ok := arts.Lookup(name)
			if !ok {
				fmt.Fprintf(os.Stderr, "unknown art generator %q in -arts\n", name)
				os.Exit(1)
			}
			chosenPalName := desc.DefaultPal
			if *palName != "" {
				chosenPalName = *palName
			}
			pal, _ := arts.PaletteByName(chosenPalName)
			if i == 0 && len(pal) > 0 {
				cfg.Palette = pal
			}
			pSeed := *seed + int64(i)*137
			panels[i] = generativeart.PosterPanelSpec{
				Engine:     desc.Build(nil),
				Caption:    fmt.Sprintf("PLATE %02d  ·  %s  ·  SEED %d", i+1, strings.ToUpper(desc.ID), pSeed),
				Seed:       pSeed,
				Background: desc.DefaultBG,
				Palette:    pal,
			}
			upperNames = append(upperNames, strings.ToUpper(desc.ID))
		}
		if *title != "" {
			cfg.Title = *title
		} else {
			cfg.Title = "GENERATIVE ART CURATED EXHIBITION"
		}
		if *subtitle != "" {
			cfg.Subtitle = *subtitle
		} else {
			cfg.Subtitle = strings.Join(upperNames, "  ·  ")
		}
		cfg.RecipeLabel = fmt.Sprintf("CURATED %dx%d EXHIBITION SHEET  ·  64-BIT LINEAR HDR  ·  BASE SEED %d", rows, cols, *seed)
		posterCanva = generativeart.RenderGalleryPoster(panels, cfg)
	} else {
		desc, ok := arts.Lookup(*artName)
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown art generator %q\n", *artName)
			os.Exit(1)
		}
		chosenPalName := desc.DefaultPal
		if *palName != "" {
			chosenPalName = *palName
		}
		if pal, exists := arts.PaletteByName(chosenPalName); exists {
			cfg.Palette = pal
		}
		cfg.Background = desc.DefaultBG
		if *title != "" {
			cfg.Title = *title
		} else {
			cfg.Title = strings.ToUpper(desc.ID) + "  ·  VARIATION STUDY"
		}
		if *subtitle != "" {
			cfg.Subtitle = *subtitle
		} else {
			cfg.Subtitle = fmt.Sprintf("CATEGORY: %s  ·  PALETTE: %s  ·  64-BIT LINEAR HDR", strings.ToUpper(desc.Category), strings.ToUpper(chosenPalName))
		}
		cfg.RecipeLabel = fmt.Sprintf("ENGINE: arts.New%s  ·  GRID: %dx%d  ·  PANEL: %dx%d  ·  SEED: %d",
			strings.Title(desc.ID), rows, cols, *panelSize, *panelSize, *seed)
		posterCanva = generativeart.RenderPoster(desc.Build(nil), cfg)
	}

	if err := saveCanvaByExt(posterCanva, nil, 1, 8, *out, true); err != nil {
		fmt.Fprintf(os.Stderr, "poster save failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Rendered exhibition poster (%dx%d px, grid=%dx%d, theme=%s) -> %s\n",
		posterCanva.Width(), posterCanva.Height(), rows, cols, *theme, *out)
}

