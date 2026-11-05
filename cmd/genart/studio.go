package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"image/png"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/render"
)

type curatedCandidate struct {
	canva   *generativeart.Canva
	artID   string
	palette string
	seed    int64
	params  map[string]interface{}
	metrics render.AestheticMetrics
}

func runCurate(args []string) {
	fs := flag.NewFlagSet("curate", flag.ExitOnError)
	artName := fs.String("art", "", "Art generator name (empty = sample across top modern engines)")
	count := fs.Int("count", 12, "Number of randomized candidates to generate and evaluate")
	keep := fs.Int("keep", 4, "Number of highest-scoring works to save")
	w := fs.Int("w", 480, "Canvas width in pixels")
	h := fs.Int("h", 480, "Canvas height in pixels")
	seedBase := fs.Int64("seed", 20261105, "Base deterministic seed for candidate exploration")
	outDir := fs.String("dir", "curated_out", "Output directory for curated top works")
	pf := registerPostFlags(fs)
	_ = fs.Parse(args)

	if *count < 1 {
		*count = 1
	}
	if *keep < 1 {
		*keep = 1
	}
	if *keep > *count {
		*keep = *count
	}
	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create output directory %q: %v\n", *outDir, err)
		os.Exit(1)
	}

	allDescs := arts.ListDescriptors()
	var pool []arts.ArtDescriptor
	if strings.TrimSpace(*artName) != "" {
		desc, ok := arts.Lookup(*artName)
		if !ok {
			fmt.Fprintf(os.Stderr, "unknown art generator %q\n", *artName)
			os.Exit(1)
		}
		pool = []arts.ArtDescriptor{desc}
	} else {
		// Use top modern descriptors (first 80 in registry)
		limit := len(allDescs)
		if limit > 80 {
			limit = 80
		}
		pool = allDescs[:limit]
	}

	candidates := make([]curatedCandidate, *count)
	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
	if workers > *count {
		workers = *count
	}

	var wg sync.WaitGroup
	for wIdx := 0; wIdx < workers; wIdx++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := workerID; i < *count; i += workers {
				itemSeed := *seedBase + int64(i)*1009 + 137
				rng := rand.New(rand.NewSource(itemSeed))
				desc := pool[rng.Intn(len(pool))]
				palName := desc.RandomizePalette(rng)
				params := desc.RandomizeParams(rng)

				c := generativeart.NewCanva(*w, *h)
				c.SetSeedLocal(itemSeed)
				c.SetBackground(desc.DefaultBG)
				if pal, exists := arts.PaletteByName(palName); exists {
					c.SetColorSchema(pal)
				}
				if effects := pf.buildEffects(itemSeed); len(effects) > 0 {
					c.AddPostEffect(effects...)
				}
				c.FillBackground()
				c.Draw(desc.Build(params))
				metrics := c.EvaluateAestheticMetrics()

				candidates[i] = curatedCandidate{
					canva:   c,
					artID:   desc.ID,
					palette: palName,
					seed:    itemSeed,
					params:  params,
					metrics: metrics,
				}
			}
		}(wIdx)
	}
	wg.Wait()

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].metrics.TotalScore > candidates[j].metrics.TotalScore
	})

	fmt.Printf("Evaluated %d candidates — saving Top %d to %s/:\n", *count, *keep, *outDir)
	for rank := 0; rank < *keep; rank++ {
		cand := candidates[rank]
		fname := fmt.Sprintf("rank%02d_score%.1f_%s_s%d.png", rank+1, cand.metrics.TotalScore, cand.artID, cand.seed)
		fpath := filepath.Join(*outDir, fname)
		if err := cand.canva.ToPNG(fpath); err != nil {
			fmt.Fprintf(os.Stderr, "failed to save %s: %v\n", fpath, err)
			continue
		}
		fmt.Printf("  #%02d  [Score: %5.1f/100 | DR:%.2f Con:%.2f Har:%.2f Str:%.2f Foc:%.2f]  %-22s  palette=%-10s  seed=%d -> %s\n",
			rank+1,
			cand.metrics.TotalScore,
			cand.metrics.DynamicRange,
			cand.metrics.ContrastStd,
			cand.metrics.ColorHarmony,
			cand.metrics.StructuralEntropy,
			cand.metrics.FocalSaliency,
			cand.artID,
			cand.palette,
			cand.seed,
			fpath,
		)
	}
}

// buildStudioMux returns the HTTP handler for `genart serve`, exposing `/api/engines`,
// `/api/render`, `/api/inspect`, and the interactive Studio UI.
func buildStudioMux() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/engines", func(w http.ResponseWriter, r *http.Request) {
		descs := arts.ListDescriptors()
		type engineSummary struct {
			ID                string           `json:"id"`
			Category          string           `json:"category"`
			DefaultPalette    string           `json:"default_palette"`
			SupportsAnimation bool             `json:"supports_animation"`
			Parameters        []arts.ParamSpec `json:"parameters"`
		}
		list := make([]engineSummary, len(descs))
		for i, d := range descs {
			list[i] = engineSummary{
				ID:                d.ID,
				Category:          d.Category,
				DefaultPalette:    d.DefaultPal,
				SupportsAnimation: d.SupportsAnimation(),
				Parameters:        d.ParamSpecs(),
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"engines":  list,
			"palettes": arts.PaletteNames(),
		})
	})

	mux.HandleFunc("/api/render", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		artName := q.Get("art")
		if artName == "" {
			artName = "muranoglassnautilus"
		}
		desc, ok := arts.Lookup(artName)
		if !ok {
			http.Error(w, fmt.Sprintf("unknown art generator %q", artName), http.StatusBadRequest)
			return
		}

		width, _ := strconv.Atoi(q.Get("w"))
		height, _ := strconv.Atoi(q.Get("h"))
		if width <= 0 || width > 1600 {
			width = 480
		}
		if height <= 0 || height > 1600 {
			height = 480
		}
		seed, err := strconv.ParseInt(q.Get("seed"), 10, 64)
		if err != nil || seed == 0 {
			seed = 20261105
		}

		palName := q.Get("palette")
		if palName == "" {
			palName = desc.DefaultPal
		}

		var params map[string]interface{}
		if q.Get("random") == "1" || q.Get("random") == "true" {
			rng := rand.New(rand.NewSource(seed))
			palName = desc.RandomizePalette(rng)
			params = desc.RandomizeParams(rng)
		} else if setStr := q.Get("set"); setStr != "" {
			if raw, err := arts.ParseParamOverrides(setStr); err == nil {
				if resolved, err := desc.ResolveParams(raw); err == nil {
					params = resolved
				}
			}
		}
		if params == nil {
			params = desc.DefaultParams()
		}

		c := generativeart.NewCanva(width, height)
		c.SetSeedLocal(seed)
		c.SetBackground(desc.DefaultBG)
		if pal, exists := arts.PaletteByName(palName); exists {
			c.SetColorSchema(pal)
		}
		if bVal, err := strconv.ParseFloat(q.Get("bloom"), 64); err == nil && bVal > 0 {
			c.AddPostEffect(render.BloomEffect(0.62, 6.0, bVal))
		}
		if rVal, err := strconv.ParseFloat(q.Get("relief"), 64); err == nil && rVal > 0 {
			c.AddPostEffect(render.LuminanceReliefEffect(rVal, -135, 38, 0.35))
		}
		if chVal, err := strconv.ParseFloat(q.Get("chromatic"), 64); err == nil && chVal != 0 {
			c.AddPostEffect(render.ChromaticAberrationEffect(chVal))
		}
		if dofVal, err := strconv.ParseFloat(q.Get("dof"), 64); err == nil && dofVal > 0 {
			c.AddPostEffect(render.DepthOfFieldEffect(0.5, 0.5, 0.28, dofVal, 6))
		}

		c.FillBackground()
		c.Draw(desc.Build(params))
		metrics := c.EvaluateAestheticMetrics()

		var pngBuf bytes.Buffer
		_ = png.Encode(&pngBuf, c.Img())
		pngBytes := pngBuf.Bytes()
		if withMeta, err := generativeart.InjectPNGRecipe(pngBytes, c.Recipe()); err == nil {
			pngBytes = withMeta
		}

		var setParts []string
		for _, sp := range desc.ParamSpecs() {
			if v, exists := params[sp.Name]; exists {
				setParts = append(setParts, fmt.Sprintf("%s=%v", sp.Name, v))
			}
		}
		cliCmd := fmt.Sprintf("go run ./cmd/genart render -art %s -palette %s -seed %d -w %d -h %d", desc.ID, palName, seed, width, height)
		if len(setParts) > 0 {
			cliCmd += fmt.Sprintf(" -set %q", strings.Join(setParts, ","))
		}
		cliCmd += fmt.Sprintf(" -out %s.png", desc.ID)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"art":          desc.ID,
			"palette":      palName,
			"seed":         seed,
			"width":        width,
			"height":       height,
			"params":       params,
			"aesthetic":    metrics,
			"cli_command":  cliCmd,
			"image_base64": "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes),
		})
	})

	mux.HandleFunc("/api/inspect", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(io.LimitReader(r.Body, 32*1024*1024))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		meta, err := generativeart.ReadPNGRecipeFromBytes(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(meta)
	})

	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			http.ServeFile(w, r, "index.html")
			return
		}
		http.NotFound(w, r)
	})

	return mux
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", ":8080", "HTTP listen address (e.g. :8080 or 127.0.0.1:8080)")
	_ = fs.Parse(args)

	fmt.Printf("GenerativeArt Live Studio listening on http://localhost%s\n", *addr)
	fmt.Println("  - GET  /             -> Interactive Exhibition Gallery (index.html)")
	fmt.Println("  - GET  /api/engines  -> Full catalog of 148 engines, palettes & parameter schemas")
	fmt.Println("  - GET  /api/render   -> Real-time multi-core HDR render + Oklab AestheticMetrics score")
	fmt.Println("  - POST /api/inspect  -> Extract embedded GenerativeArt-Recipe tEXt metadata from any PNG")
	if err := http.ListenAndServe(*addr, buildStudioMux()); err != nil {
		fmt.Fprintf(os.Stderr, "serve failed: %v\n", err)
		os.Exit(1)
	}
}
