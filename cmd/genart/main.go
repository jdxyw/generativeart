package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
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
	fmt.Println("  genart render -art <name> [-palette <name>] [-w 600] [-h 600] [-seed 20261021] [-frames 1] -out <file.(png|svg|pdf|gif|jpg)>")
	fmt.Println("  genart inspect <file.png>")
	fmt.Println("  genart replay <input.png> [-w 2400] [-h 2400] -out <file.(png|svg|pdf|jpg)>")
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
	if err := saveCanvaByExt(c, nil, 1, 8, *out, true); err != nil {
		fmt.Fprintf(os.Stderr, "save failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Replayed %s (seed=%d) at %dx%d -> %s\n", meta.Engine, meta.Seed, c.Width(), c.Height(), *out)
}
