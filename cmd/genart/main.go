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
		runList(os.Args[2:])
	case "info", "describe":
		runInfo(os.Args[2:])
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
	fmt.Println("  genart list [-v]")
	fmt.Println("  genart info <art>")
	fmt.Println("  genart render -art <name> [-set k1=v1,k2=v2] [-palette <name>] [-w 600] [-h 600] [-seed 20261021] [-frames 1] [-delay 8] [-shutter 0] [-subsamples 1] [-bloom 0.35] [-chromatic 0.8] [-relief 0.6] [-kuwahara 3] [-paper 0.3] [-vignette 0.25] [-deckle 0.04] [-tonemap 1.1] -out <file.(png|svg|pdf|gif|jpg)>")
	fmt.Println("  genart poster [-art <name> | -arts <a,b,c,d>] [-set k1=v1,k2=v2] [-sweep-x name:min:max] [-sweep-y name:min:max] [-grid 2x2] [-panel 420] [-frames 1] [-delay 8] [-theme light|dark] [-title \"...\"] -out <file.(png|svg|pdf|gif|jpg)>")
	fmt.Println("  genart inspect <file.png>")
	fmt.Println("  genart replay <input.png> [-w 2400] [-h 2400] [-set k1=v1,k2=v2] [-palette <name>] [-seed 0] [-frames 1] [-delay 8] [-bloom ...] [-relief ...] -out <file.(png|svg|pdf|gif|jpg)>")
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
	dof        *float64
	plotter    *bool
	pens       *int
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
		dof:        fs.Float64("dof", 0, "Thin-lens polygonal-aperture bokeh Depth-of-Field max CoC radius in pixels (e.g. 6.0)"),
		plotter:    fs.Bool("plotter", false, "Export pen-plotter SVG with <g> color layers and 2-Opt TSP path optimization"),
		pens:       fs.Int("pens", 0, "Maximum number of pen color layers for -plotter SVG (0 = exact colors)"),
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
	if pf.dof != nil && *pf.dof > 0 {
		effects = append(effects, render.DepthOfFieldEffect(0.5, 0.5, 0.28, *pf.dof, 6))
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

func runList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	verbose := fs.Bool("v", false, "Show tunable parameters and animation support for each art generator")
	_ = fs.Parse(args)

	descs := arts.ListDescriptors()
	sort.Slice(descs, func(i, j int) bool { return descs[i].ID < descs[j].ID })
	fmt.Printf("Available Art Generators (%d):\n", len(descs))
	for _, d := range descs {
		if !*verbose {
			fmt.Printf("  - %-22s [%-6s] (default palette: %s)\n", d.ID, d.Category, d.DefaultPal)
			continue
		}
		animTag := "static"
		if d.SupportsAnimation() {
			animTag = "animated"
		}
		var pList []string
		for _, sp := range d.ParamSpecs() {
			pList = append(pList, fmt.Sprintf("%s:%s=%v", sp.Name, sp.Type, sp.Default))
		}
		fmt.Printf("  - %-22s [%-6s | %-8s] (palette: %-11s) params: {%s}\n",
			d.ID, d.Category, animTag, d.DefaultPal, strings.Join(pList, ", "))
	}

	pNames := arts.PaletteNames()
	fmt.Printf("\nAvailable Palettes (%d): %s\n", len(pNames), strings.Join(pNames, ", "))
}

func runInfo(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: genart info <art>")
		os.Exit(1)
	}
	desc, ok := arts.Lookup(args[0])
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown art generator %q (run 'genart list' to see supported generators)\n", args[0])
		os.Exit(1)
	}
	specs := desc.ParamSpecs()
	infoObj := struct {
		ID                string           `json:"id"`
		StructName        string           `json:"struct_name"`
		Category          string           `json:"category"`
		DefaultPalette    string           `json:"default_palette"`
		DefaultBackground [4]uint8         `json:"default_background"`
		SupportsAnimation bool             `json:"supports_animation"`
		Parameters        []arts.ParamSpec `json:"parameters"`
	}{
		ID:                desc.ID,
		StructName:        desc.StructName,
		Category:          desc.Category,
		DefaultPalette:    desc.DefaultPal,
		DefaultBackground: [4]uint8{desc.DefaultBG.R, desc.DefaultBG.G, desc.DefaultBG.B, desc.DefaultBG.A},
		SupportsAnimation: desc.SupportsAnimation(),
		Parameters:        specs,
	}
	pretty, _ := json.MarshalIndent(infoObj, "", "  ")
	fmt.Println(string(pretty))
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
	setExpr := fs.String("set", "", "Override engine parameters (e.g. 'galleries=8,tiers=6,shaftScale=1.25' or JSON)")
	palName := fs.String("palette", "", "Color palette name (optional override)")
	w := fs.Int("w", 600, "Canvas width in pixels")
	h := fs.Int("h", 600, "Canvas height in pixels")
	seed := fs.Int64("seed", 20261021, "Deterministic random seed")
	frames := fs.Int("frames", 1, "Number of frames (for .gif timelapse)")
	delay := fs.Int("delay", 8, "Frame delay in centiseconds (for .gif)")
	shutter := fs.Float64("shutter", 0, "Shutter-angle fraction in [0, 1] for temporal motion blur in .gif animations")
	subsamples := fs.Int("subsamples", 1, "Number of temporal sub-samples per frame for motion blur (e.g. 4)")
	colors := fs.Int("colors", 31, "Maximum opaque colors in shared GIF palette (2..255, default 31)")
	out := fs.String("out", "output.png", "Output file path (.png, .svg, .pdf, .gif, .jpg)")
	pf := registerPostFlags(fs)
	_ = fs.Parse(args)

	desc, ok := arts.Lookup(*artName)
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown art generator %q (run 'genart list' to see supported generators)\n", *artName)
		os.Exit(1)
	}

	var params map[string]interface{}
	if strings.TrimSpace(*setExpr) != "" {
		rawOverrides, err := arts.ParseParamOverrides(*setExpr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid -set expression: %v\n", err)
			os.Exit(1)
		}
		resolved, err := desc.ResolveParams(rawOverrides)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		params = resolved
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

	eng := desc.Build(params)
	if strings.EqualFold(filepath.Ext(*out), ".svg") && ((pf.plotter != nil && *pf.plotter) || (pf.pens != nil && *pf.pens > 0)) {
		c.Draw(eng)
		pcfg := render.DefaultPlotterConfig()
		if pf.pens != nil && *pf.pens > 0 {
			pcfg.MaxPens = *pf.pens
		}
		stats, err := c.ToPlotterSVG(*out, pcfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "plotter svg failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Rendered Pen-Plotter SVG %s (%dx%d, layers=%d, paths=%d, pen-up: %.0fpx -> %.0fpx [-%.1f%%]) -> %s\n",
			desc.ID, *w, *h, stats.LayerCount, stats.TotalPaths, stats.PenUpBeforePx, stats.PenUpAfterPx, stats.TravelReductionPct, *out)
		return
	}
	if strings.EqualFold(filepath.Ext(*out), ".gif") {
		nFrames := *frames
		if nFrames < 2 {
			nFrames = 8
		}
		if anim, isAnim := eng.(generativeart.AnimatedEngine); isAnim && *subsamples > 1 && *shutter > 0 {
			c.DrawTimelapseFramesTemporalAA(anim, nFrames, *subsamples, *shutter)
		} else {
			c.DrawTimelapseFrames(eng, nFrames)
		}
		if err := c.ToGIFWithColors(*out, *delay, *colors); err != nil {
			fmt.Fprintf(os.Stderr, "gif encode failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Rendered %s (%dx%d, frames=%d, seed=%d) -> %s\n", desc.ID, *w, *h, nFrames, *seed, *out)
		return
	}
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
	setExpr := fs.String("set", "", "Override recipe parameters (e.g. 'galleries=8,shaftScale=1.2')")
	palName := fs.String("palette", "", "Override color palette name")
	seedOverride := fs.Int64("seed", 0, "Override random seed (0 keeps original recipe seed)")
	frames := fs.Int("frames", 1, "Number of frames (for replaying a recipe as a .gif animation)")
	delay := fs.Int("delay", 8, "Frame delay in centiseconds (for .gif)")
	out := fs.String("out", "replayed.png", "Output file path (.png, .svg, .pdf, .gif, .jpg)")
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
		fmt.Fprintln(os.Stderr, "usage: genart replay <input.png> [-w 2400] [-h 2400] [-set k=v] [-frames 20] -out <output.(png|svg|pdf|gif|jpg)>")
		os.Exit(1)
	}

	meta, err := generativeart.ReadPNGRecipe(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "replay failed to read PNG recipe: %v\n", err)
		os.Exit(1)
	}
	var overrides map[string]interface{}
	if strings.TrimSpace(*setExpr) != "" {
		overrides, err = arts.ParseParamOverrides(*setExpr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid -set expression: %v\n", err)
			os.Exit(1)
		}
	}
	c, eng, err := arts.PrepareReplayCanva(meta, *w, *h, overrides)
	if err != nil {
		fmt.Fprintf(os.Stderr, "replay failed: %v\n", err)
		os.Exit(1)
	}
	if *seedOverride != 0 {
		c.SetSeed(*seedOverride)
	}
	if *palName != "" {
		if schema, exists := arts.PaletteByName(*palName); exists {
			c.SetColorSchema(schema)
		}
	}
	if effects := pf.buildEffects(c.Seed()); len(effects) > 0 {
		c.AddPostEffect(effects...)
	}

	if strings.EqualFold(filepath.Ext(*out), ".svg") && ((pf.plotter != nil && *pf.plotter) || (pf.pens != nil && *pf.pens > 0)) {
		c.Draw(eng)
		pcfg := render.DefaultPlotterConfig()
		if pf.pens != nil && *pf.pens > 0 {
			pcfg.MaxPens = *pf.pens
		}
		stats, err := c.ToPlotterSVG(*out, pcfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "plotter svg failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Replayed Pen-Plotter SVG %s (layers=%d, paths=%d, pen-up: %.0fpx -> %.0fpx [-%.1f%%]) -> %s\n",
			meta.Engine, stats.LayerCount, stats.TotalPaths, stats.PenUpBeforePx, stats.PenUpAfterPx, stats.TravelReductionPct, *out)
		return
	}
	if err := saveCanvaByExt(c, eng, *frames, *delay, *out, false); err != nil {
		fmt.Fprintf(os.Stderr, "save failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Replayed %s (seed=%d) at %dx%d -> %s\n", meta.Engine, c.Seed(), c.Width(), c.Height(), *out)
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

func parseSweepAxis(expr string, desc arts.ArtDescriptor) (generativeart.ParameterSweepAxis, error) {
	s := strings.TrimSpace(expr)
	if s == "" {
		return generativeart.ParameterSweepAxis{}, nil
	}
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return generativeart.ParameterSweepAxis{}, fmt.Errorf("expected sweep axis format 'paramName:min:max', got %q", expr)
	}
	rawName := strings.TrimSpace(parts[0])
	minV, err1 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	maxV, err2 := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
	if err1 != nil || err2 != nil {
		return generativeart.ParameterSweepAxis{}, fmt.Errorf("invalid numeric range in sweep axis %q", expr)
	}
	specs := desc.ParamSpecs()
	canonName := rawName
	isInt := false
	found := false
	for _, sp := range specs {
		if strings.EqualFold(sp.Name, rawName) {
			canonName = sp.Name
			isInt = (sp.Type == "int" || sp.Type == "uint")
			found = true
			break
		}
	}
	if !found && len(specs) > 0 {
		var valid []string
		for _, sp := range specs {
			valid = append(valid, sp.Name)
		}
		return generativeart.ParameterSweepAxis{}, fmt.Errorf("unknown parameter %q for %s (valid: %s)", rawName, desc.ID, strings.Join(valid, ", "))
	}
	return generativeart.ParameterSweepAxis{
		Name:  canonName,
		Min:   minV,
		Max:   maxV,
		IsInt: isInt,
	}, nil
}

func runPoster(args []string) {
	fs := flag.NewFlagSet("poster", flag.ExitOnError)
	artName := fs.String("art", "dunhuangapsara", "Single art generator name for variation or parameter-sweep contact sheet")
	artsList := fs.String("arts", "", "Comma-separated list of art generator names for curated multi-engine poster")
	setExpr := fs.String("set", "", "Override base engine parameters (e.g. 'galleries=8,tiers=6')")
	sweepX := fs.String("sweep-x", "", "Sweep parameter across columns: 'paramName:min:max' (e.g. 'ribs:12:36')")
	sweepY := fs.String("sweep-y", "", "Sweep parameter across rows: 'paramName:min:max' (e.g. 'ringTilt:0.15:0.55')")
	palName := fs.String("palette", "", "Optional palette override")
	gridStr := fs.String("grid", "2x2", "Grid layout RowsxCols (e.g. 1x1, 2x2, 3x3)")
	panelSize := fs.Int("panel", 420, "Width/Height in pixels of each artwork panel")
	frames := fs.Int("frames", 1, "Number of animation frames (when >1 or .gif output, renders an animated exhibition poster)")
	delay := fs.Int("delay", 8, "Frame delay in centiseconds (for .gif poster)")
	theme := fs.String("theme", "light", "Poster mat theme: 'light' (Museum Ivory Parchment) or 'dark' (Obsidian Gallery Gold)")
	title := fs.String("title", "", "Custom poster main title")
	subtitle := fs.String("subtitle", "", "Custom poster subtitle")
	seed := fs.Int64("seed", 20261022, "Base deterministic random seed")
	out := fs.String("out", "poster.png", "Output file path (.png, .svg, .pdf, .gif, .jpg)")
	pf := registerPostFlags(fs)
	_ = fs.Parse(args)

	rows, cols := parseGridDim(*gridStr)
	cfg := generativeart.DefaultPosterConfig()
	cfg.Rows = rows
	cfg.Cols = cols
	cfg.PanelSize = *panelSize
	cfg.SeedBase = *seed
	cfg.DarkTheme = strings.EqualFold(strings.TrimSpace(*theme), "dark")
	cfg.PostEffects = pf.buildEffects(*seed)

	isGIF := strings.EqualFold(filepath.Ext(*out), ".gif")
	nFrames := *frames
	if isGIF && nFrames < 2 {
		nFrames = 8
	}

	var rawOverrides map[string]interface{}
	if strings.TrimSpace(*setExpr) != "" {
		var err error
		rawOverrides, err = arts.ParseParamOverrides(*setExpr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid -set expression: %v\n", err)
			os.Exit(1)
		}
	}

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
				Engine:     desc.Build(rawOverrides),
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
		if nFrames > 1 {
			posterCanva = generativeart.RenderAnimatedGalleryPoster(panels, cfg, nFrames)
		} else {
			posterCanva = generativeart.RenderGalleryPoster(panels, cfg)
		}
	} else {
		desc, ok := arts.Lookup(*artName)
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown art generator %q\n", *artName)
			os.Exit(1)
		}
		var baseParams map[string]interface{}
		if len(rawOverrides) > 0 {
			var err error
			baseParams, err = desc.ResolveParams(rawOverrides)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		} else {
			baseParams = desc.DefaultParams()
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
		}
		if *subtitle != "" {
			cfg.Subtitle = *subtitle
		}

		if strings.TrimSpace(*sweepX) != "" || strings.TrimSpace(*sweepY) != "" {
			axX, err := parseSweepAxis(*sweepX, desc)
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid -sweep-x: %v\n", err)
				os.Exit(1)
			}
			axY, err := parseSweepAxis(*sweepY, desc)
			if err != nil {
				fmt.Fprintf(os.Stderr, "invalid -sweep-y: %v\n", err)
				os.Exit(1)
			}
			posterCanva = generativeart.RenderParameterSweepPoster(generativeart.ParameterSweepConfig{
				EngineName: desc.ID,
				Build:      desc.Build,
				BaseParams: baseParams,
				AxisX:      axX,
				AxisY:      axY,
			}, cfg)
		} else {
			if cfg.Title == "" {
				cfg.Title = strings.ToUpper(desc.ID) + "  ·  VARIATION STUDY"
			}
			if cfg.Subtitle == "" {
				cfg.Subtitle = fmt.Sprintf("CATEGORY: %s  ·  PALETTE: %s  ·  64-BIT LINEAR HDR", strings.ToUpper(desc.Category), strings.ToUpper(chosenPalName))
			}
			cfg.RecipeLabel = fmt.Sprintf("ENGINE: arts.New%s  ·  GRID: %dx%d  ·  PANEL: %dx%d  ·  SEED: %d",
				strings.Title(desc.ID), rows, cols, *panelSize, *panelSize, *seed)
			if nFrames > 1 {
				posterCanva = generativeart.RenderAnimatedPoster(desc.Build(baseParams), cfg, nFrames)
			} else {
				posterCanva = generativeart.RenderPoster(desc.Build(baseParams), cfg)
			}
		}
	}

	if err := saveCanvaByExt(posterCanva, nil, nFrames, *delay, *out, true); err != nil {
		fmt.Fprintf(os.Stderr, "poster save failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Rendered exhibition poster (%dx%d px, grid=%dx%d, theme=%s) -> %s\n",
		posterCanva.Width(), posterCanva.Height(), rows, cols, *theme, *out)
}

