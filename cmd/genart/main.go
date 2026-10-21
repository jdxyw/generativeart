package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

type artPreset struct {
	bg      color.RGBA
	palette string
	build   func() generativeart.Engine
}

var palettes = map[string][]color.RGBA{
	"cyberpunk":   common.Cyberpunk,
	"dunhuang":    common.Dunhuang,
	"ukiyoe":      common.UkiyoE,
	"inferno":     common.Inferno,
	"magma":       common.Magma,
	"viridis":     common.Viridis,
	"plasma":      common.Plasma,
	"oceanic":     common.Oceanic,
	"sunset":      common.Sunset,
	"botanical":   common.Botanical,
	"pasteldream": common.PastelDream,
	"youthful":    common.Youthful,
}

var presets = map[string]artPreset{
	"cloisonneenamel":   {color.RGBA{18, 14, 22, 255}, "dunhuang", func() generativeart.Engine { return arts.NewCloisonneEnamel(6, 12) }},
	"woodblockwave":     {color.RGBA{242, 232, 206, 255}, "ukiyoe", func() generativeart.Engine { return arts.NewWoodblockWave(5, 42) }},
	"artdecosunburst":   {color.RGBA{14, 16, 24, 255}, "sunset", func() generativeart.Engine { return arts.NewArtDecoSunburst(36, 7) }},
	"stainedrosewindow": {color.RGBA{12, 11, 18, 255}, "cyberpunk", func() generativeart.Engine { return arts.NewStainedRoseWindow(12, 5) }},
	"mosaicvortex":      {color.RGBA{16, 20, 38, 255}, "oceanic", func() generativeart.Engine { return arts.NewMosaicVortex(34, 1.0) }},
	"mosaicwave":        {color.RGBA{18, 24, 44, 255}, "ukiyoe", func() generativeart.Engine { return arts.NewMosaicWave(32, 1.0) }},
	"mosaiccanyon":      {color.RGBA{22, 18, 32, 255}, "dunhuang", func() generativeart.Engine { return arts.NewMosaicCanyon(32, 1.0) }},
	"mosaiclotus":       {color.RGBA{16, 24, 42, 255}, "botanical", func() generativeart.Engine { return arts.NewMosaicLotus(12, 1.0) }},
	"mosaicsun":         {color.RGBA{22, 26, 48, 255}, "sunset", func() generativeart.Engine { return arts.NewMosaicSun(32, 1.0) }},
	"impastoflow":       {color.RGBA{8, 7, 16, 255}, "sunset", func() generativeart.Engine { return arts.NewImpastoFlow(3200, 520, 0.0032) }},
	"voxelterrain":      {color.RGBA{10, 12, 26, 255}, "oceanic", func() generativeart.Engine { return arts.NewVoxelTerrain(340, 2.6, 1.0) }},
	"packedcapsules":    {color.RGBA{238, 236, 229, 255}, "oceanic", func() generativeart.Engine { return arts.NewPackedCapsules(2200, 7.2, 3.6) }},
	"orphicdiscs":       {color.RGBA{72, 178, 102, 255}, "youthful", func() generativeart.Engine { return arts.NewOrphicDiscs(34) }},
	"nautilusspiral":    {color.RGBA{18, 24, 34, 255}, "dunhuang", func() generativeart.Engine { return arts.NewNautilusSpiral(5, 68) }},
	"tangentcones":      {color.RGBA{234, 88, 56, 255}, "sunset", func() generativeart.Engine { return arts.NewTangentCones(26) }},
	"poincaredisk":      {color.RGBA{8, 10, 20, 255}, "dunhuang", func() generativeart.Engine { return arts.NewPoincareDisk(7, 3, 5) }},
	"suminagashi":       {color.RGBA{244, 239, 229, 255}, "ukiyoe", func() generativeart.Engine { return arts.NewSuminagashi(48, 7, 1.0) }},
	"aizawaattractor":   {color.RGBA{5, 7, 15, 255}, "cyberpunk", func() generativeart.Engine { return arts.NewAizawaAttractor(900000, 1.0) }},
	"penrosetiling":     {color.RGBA{10, 14, 24, 255}, "oceanic", func() generativeart.Engine { return arts.NewPenroseTiling(6) }},
	"isometriccity":     {color.RGBA{18, 20, 34, 255}, "sunset", func() generativeart.Engine { return arts.NewIsometricCity(11, 1.0) }},
	"dlafrost":          {color.RGBA{4, 8, 18, 255}, "oceanic", func() generativeart.Engine { return arts.NewDLAFrost(18000, 0.55) }},
	"curlflow":          {color.RGBA{14, 18, 28, 255}, "cyberpunk", func() generativeart.Engine { return arts.NewCurlFlow(110, 36, 0.0045) }},
	"sandring":          {color.RGBA{10, 10, 18, 255}, "dunhuang", func() generativeart.Engine { return arts.NewSandRing(22, 22, 1.3) }},
	"watercolor":        {color.RGBA{246, 242, 234, 255}, "botanical", func() generativeart.Engine { return arts.NewWatercolor(9, 22, 5, 0.32) }},
	"guilloche":         {color.RGBA{8, 12, 22, 255}, "cyberpunk", func() generativeart.Engine { return arts.NewGuilloche(28, 175, 65, 115) }},
	"cliffordnebula":    {color.RGBA{6, 6, 14, 255}, "inferno", func() generativeart.Engine { return arts.NewCliffordNebula(1500000, -1.7, 1.8, -1.9, -0.4) }},
	"physarum":          {color.RGBA{6, 10, 18, 255}, "cyberpunk", func() generativeart.Engine { return arts.NewPhysarum(45000, 110, 0, 0) }},
	"differentialgrowth": {color.RGBA{10, 12, 22, 255}, "sunset", func() generativeart.Engine {
		return arts.NewDifferentialGrowth(220, 1600, 3.5, 7.5, 18.0)
	}},
	"inklandscape":      {color.RGBA{244, 239, 228, 255}, "ukiyoe", func() generativeart.Engine { return arts.NewInkLandscape(8, 0.45) }},
	"voronoicrystal":    {color.RGBA{24, 22, 30, 255}, "dunhuang", func() generativeart.Engine { return arts.NewVoronoiCrystal(85, 4) }},
	"grayscott":         {color.RGBA{8, 12, 20, 255}, "viridis", func() generativeart.Engine { return arts.NewGrayScott(3200, 0.037, 0.060) }},
	"lissajousribbon":   {color.RGBA{6, 8, 16, 255}, "plasma", func() generativeart.Engine { return arts.NewLissajousRibbon(140, 3.0, 4.0, 5.0) }},
	"apolloniangasket":  {color.RGBA{10, 14, 24, 255}, "dunhuang", func() generativeart.Engine { return arts.NewApollonianGasket(7, 2.5) }},
	"spacecolonization": {color.RGBA{8, 14, 18, 255}, "cyberpunk", func() generativeart.Engine { return arts.NewSpaceColonization(1800, 160, 6.0, 55.0) }},
	"truchetflow":       {color.RGBA{245, 240, 230, 255}, "ukiyoe", func() generativeart.Engine { return arts.NewTruchetFlow(6, 3) }},
	"aurorawave":        {color.RGBA{6, 10, 22, 255}, "oceanic", func() generativeart.Engine { return arts.NewAuroraWave(7, 160, 1.0) }},
	"fractalflame":      {color.RGBA{4, 4, 10, 255}, "magma", func() generativeart.Engine { return arts.NewFractalFlame(1200000, 3) }},
	"joydivision":       {color.RGBA{10, 12, 20, 255}, "cyberpunk", func() generativeart.Engine { return arts.NewJoyDivision(58, 46.0, 0.022) }},
	"boidsflock":        {color.RGBA{10, 14, 24, 255}, "sunset", func() generativeart.Engine { return arts.NewBoidsFlock(260, 140, 48.0) }},
	"islamicstar":       {color.RGBA{12, 18, 32, 255}, "dunhuang", func() generativeart.Engine { return arts.NewIslamicStar(6, 60.0, 4.5) }},
	"delaunaystained":   {color.RGBA{16, 14, 24, 255}, "sunset", func() generativeart.Engine { return arts.NewDelaunayStained(130, 0.65) }},
	"phyllotaxis":       {color.RGBA{10, 14, 20, 255}, "botanical", func() generativeart.Engine { return arts.NewPhyllotaxis(950, 137.507764, 8.2) }},
}

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
	case "inspect":
		runInspect(os.Args[2:])
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  genart list")
	fmt.Println("  genart render -art <name> [-palette <name>] [-w 600] [-h 600] [-seed 20261021] [-frames 1] -out <file.(png|svg|pdf|gif|jpg)>")
	fmt.Println("  genart inspect <file.png>")
}

func runList() {
	names := make([]string, 0, len(presets))
	for k := range presets {
		names = append(names, k)
	}
	sort.Strings(names)
	fmt.Printf("Available Art Generators (%d):\n", len(names))
	for _, n := range names {
		fmt.Printf("  - %-20s (default palette: %s)\n", n, presets[n].palette)
	}

	pNames := make([]string, 0, len(palettes))
	for k := range palettes {
		pNames = append(pNames, k)
	}
	sort.Strings(pNames)
	fmt.Printf("\nAvailable Palettes (%d): %s\n", len(pNames), strings.Join(pNames, ", "))
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
	_ = fs.Parse(args)

	key := strings.ToLower(strings.TrimSpace(*artName))
	preset, ok := presets[key]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown art generator %q (run 'genart list' to see supported generators)\n", *artName)
		os.Exit(1)
	}

	c := generativeart.NewCanva(*w, *h)
	c.SetSeed(*seed)
	c.SetBackground(preset.bg)
	c.FillBackground()

	chosenPal := preset.palette
	if *palName != "" {
		chosenPal = strings.ToLower(strings.TrimSpace(*palName))
	}
	if schema, exists := palettes[chosenPal]; exists {
		c.SetColorSchema(schema)
	}

	eng := preset.build()
	ext := strings.ToLower(filepath.Ext(*out))

	var err error
	switch ext {
	case ".gif":
		nFrames := *frames
		if nFrames < 2 {
			nFrames = 8
		}
		c.DrawTimelapseFrames(eng, nFrames)
		err = c.ToGIF(*out, *delay)
	case ".svg":
		c.Draw(eng)
		err = c.ToSVG(*out)
	case ".pdf":
		c.Draw(eng)
		err = c.ToPDF(*out)
	case ".jpg", ".jpeg":
		c.Draw(eng)
		err = c.ToJPEG(*out)
	default:
		c.Draw(eng)
		err = c.ToPNG(*out)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "render failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Rendered %s (%dx%d, seed=%d) -> %s\n", key, *w, *h, *seed, *out)
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
