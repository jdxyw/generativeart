# generativeart

[![Go](https://github.com/jdxyw/generativeart/actions/workflows/go.yml/badge.svg)](https://github.com/jdxyw/generativeart/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jdxyw/generativeart)](https://goreportcard.com/report/github.com/jdxyw/generativeart)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/jdxyw/generativeart/master/LICENSE)

`generativeart` is a pure-`Go` generative art engine and algorithm collection supporting **82 generative art types** across flow fields, strange attractors, biological morphogenesis, non-Euclidean geometry, quasicrystals, stained-glass mosaics, cloisonné enamel, 3D origami, and traditional printmaking.

- 🌐 **Live Interactive Web Gallery (GitHub Pages)**: **[https://jdxyw.github.io/generativeart/](https://jdxyw.github.io/generativeart/)**
- 🖼️ **Full Code & Image Gallery (`gallery.md`)**: **[./gallery.md](./gallery.md)**
- 📚 **Parameter & API Reference (`docs/doc.md`)**: **[./docs/doc.md](./docs/doc.md)**
- 🗺️ **Engine Architecture & Roadmap (`docs/roadmap.md`)**: **[./docs/roadmap.md](./docs/roadmap.md)**

---

## Featured Gallery Highlights

> Browse all **82 artworks**, full `Go` source snippets, and rendered previews in **[`gallery.md`](./gallery.md)** or on the **[Interactive GitHub Pages Gallery](https://jdxyw.github.io/generativeart/)**.

| [Ammann Beenker](./gallery.md#ammann-beenker) | [Sashiko Weave](./gallery.md#sashiko-weave) | [Magnetic Dipole](./gallery.md#magnetic-dipole) | [Newton Basin](./gallery.md#newton-basin) |
| :---: | :---: | :---: | :---: |
| ![Ammann Beenker](images/ammannbeenker.png) | ![Sashiko Weave](images/sashikoweave.png) | ![Magnetic Dipole](images/magneticdipole.png) | ![Newton Basin](images/newtonbasin.png) |
| **[Even Streamlines](./gallery.md#even-streamlines)** | **[Marching Isolines](./gallery.md#marching-isolines)** | **[Origami Tessellation](./gallery.md#origami-tessellation)** | **[Mondrian Subdivision](./gallery.md#mondrian-subdivision)** |
| ![Even Streamlines](images/evenstreamlines.png) | ![Marching Isolines](images/marchingisolines.png) | ![Origami Tessellation](images/origamitessellation.png) | ![Mondrian Subdivision](images/mondriansubdivision.png) |
| **[Cloisonné Enamel](./gallery.md#cloisonne-enamel)** | **[Woodblock Wave](./gallery.md#woodblock-wave)** | **[Art Deco Sunburst](./gallery.md#art-deco-sunburst)** | **[Stained Rose Window](./gallery.md#stained-rose-window)** |
| ![Cloisonne Enamel](images/cloisonneenamel.png) | ![Woodblock Wave](images/woodblockwave.png) | ![Art Deco Sunburst](images/artdecosunburst.png) | ![Stained Rose Window](images/stainedrosewindow.png) |

👉 **[View All 82 Artworks & Code Examples in `gallery.md` →](./gallery.md)**

---

## Core Engine Capabilities

1. **64-Bit Linear HDR Rendering Engine (`render/`)**:
   - Floating-point linear-RGB accumulation buffer (`render.Canvas`) with `O(1)` 256-entry sRGB lookup (`SRGB8ToLinear`), 9 Porter-Duff & separable blend modes (`BlendOver`, `BlendAdd`, `BlendMultiply`, `BlendScreen`, `BlendSoftLight`, etc.), ACES filmic tone mapping, and multi-core row parallelization (`parallelRows`).
2. **Perceptual Color & Multi-Dimensional Noise (`common/`)**:
   - Exact Oklab, CIE-L\*C\*h° (HCL), and CIELAB interpolation (`MultiStopGradient` with pre-cached stops), Inigo Quilez cosine palettes, 12+ curated palettes (`Cyberpunk`, `Dunhuang`, `UkiyoE`, `Inferno`, `Oceanic`, `Sunset`, `Botanical`, etc.), and seeded 2D/3D/4D Simplex, Ridged multifractal, Worley cellular, and seamless periodic `FBM4DLoop` / `Curl4DLoop` fields.
3. **Computational Geometry, Flow Fields & 2D SDFs (`render/field_engine.go`, `render/sdf2d.go`)**:
   - Jobard-Lefer RK4 evenly-spaced streamline integrator (`TraceEvenStreamlines`), sub-pixel stitched Marching Squares contour extractor (`ExtractMarchingSquares`), analytical 2D Signed Distance Fields (`SDFCircle`, `SDFBox`, `SDFPolygon`, `SDFSmoothUnion`), soft drop shadows (`RenderSDFDropShadow`), and 3D beveled enamel/glass primitives (`DrawCloisonneCell`, `DrawStainedGlassCell`, `DrawMosaicBrick`, `FiberGBuffer`).
4. **Dual Vector (`SVG` / `PDF`), Animated `GIF` & Reproducible PNG Recipes (`platform.go`)**:
   - Export any canvas to lossless **PNG** (with embedded `GenerativeArt-Recipe` `tEXt` JSON metadata), infinite-resolution **SVG** (`c.ToSVG`) and **PDF** (`c.ToPDF`), or looping **GIF** animations (`c.DrawTimelapse` + `c.ToGIF`).

---

## Install

### Library

```bash
go get github.com/jdxyw/generativeart
```

### Standalone CLI (`cmd/genart`)

```bash
go install github.com/jdxyw/generativeart/cmd/genart@latest
```

---

## CLI Tool (`cmd/genart`)

The `cmd/genart` command-line tool lets you list available generators/palettes, render artworks to `.png`, `.svg`, `.pdf`, `.gif`, or `.jpg` at custom resolutions and seeds, and inspect embedded PNG recipe metadata.

### 1. List Supported Generators & Palettes

```bash
go run ./cmd/genart list
```

### 2. Render to PNG, Vector SVG/PDF, or Animated GIF

```bash
# Render a 1200x1200 PNG with a deterministic seed (automatically embeds recipe metadata)
go run ./cmd/genart render -art cloisonneenamel -palette dunhuang -w 1200 -h 1200 -seed 20261021 -out cloisonne.png

# Export infinite-resolution Vector SVG or archival PDF
go run ./cmd/genart render -art stainedrosewindow -w 800 -h 800 -out rose.svg
go run ./cmd/genart render -art artdecosunburst -w 800 -h 800 -out artdeco.pdf

# Render an 8-frame animated GIF timelapse
go run ./cmd/genart render -art mosaicvortex -w 400 -h 400 -frames 8 -delay 8 -out vortex.gif
```

### 3. Inspect & Replay Embedded PNG Recipe Metadata

Every PNG saved via `c.ToPNG()` or `genart render` embeds a standard PNG `tEXt` chunk (`GenerativeArt-Recipe`) recording the engine name, dimensions, seed, background color, and struct parameters:

```bash
# Inspect the embedded JSON recipe inside any generated PNG
go run ./cmd/genart inspect cloisonne.png

# Replay a saved PNG recipe at 4K resolution or convert it directly to Vector SVG / PDF
go run ./cmd/genart replay cloisonne.png -w 2400 -h 2400 -out cloisonne_4k.png
go run ./cmd/genart replay cloisonne.png -out cloisonne.svg
```

```json
{
  "engine": "cloisonneEnamel",
  "width": 1200,
  "height": 1200,
  "seed": 20261021,
  "background": [18, 14, 22, 255],
  "params": {
    "petals": 12,
    "rings": 6
  }
}
```

---

## Programmatic Usage

### Basic Rendering, Vector SVG/PDF Export & Recipe Inspection

```go
package main

import (
	"fmt"
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261021)
	c.SetBackground(color.RGBA{R: 18, G: 14, B: 22, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)

	c.Draw(arts.NewCloisonneEnamel(6, 12))

	// Save PNG (with embedded GenerativeArt-Recipe tEXt metadata), Vector SVG, and PDF
	_ = c.ToPNG("cloisonne.png")
	_ = c.ToSVG("cloisonne.svg")
	_ = c.ToPDF("cloisonne.pdf")

	// Read back the embedded recipe from the saved PNG
	recipe, _ := generativeart.ReadPNGRecipe("cloisonne.png")
	fmt.Printf("Engine=%s Seed=%d Params=%v\n", recipe.Engine, recipe.Seed, recipe.Params)
}
```

### Seamless Loop Animation (`DrawTimelapse` + `ToGIF`)

```go
anim := generativeart.NewAnimatedEngine(func(c *generativeart.Canva, frameIdx, totalFrames int, phase float64) {
	rc := c.NewRenderCanvas()
	ne := c.Noise()
	// Sample 4D curl noise along a closed periodic loop for phase in [0, 1)
	v := ne.Curl4DLoop(2.0, 2.0, phase, 1.2, 1e-3)
	rc.DrawGlowDot(300+v.X*40, 300+v.Y*40, 24, common.Tomato, 0.9, 0.5)
	c.MergeRenderCanvas(rc)
})

c.DrawTimelapseFrames(anim, 24)
_ = c.ToGIF("loop.gif", 6)
```

---

## Supported Art Types & Constructors

All 82 generators implement `generativeart.Engine` under package `github.com/jdxyw/generativeart/arts` (see [gallery.md](./gallery.md) for visual examples and [docs/doc.md](./docs/doc.md) for parameter details):

```go
// Wave 10: 8-Fold Quasicrystals, Sashiko Embroidery, Electromagnetic Dipoles & Complex Newton Fractals
NewAmmannBeenker(depth int)
NewSashikoWeave(gridSize int)
NewMagneticDipole(polesN, linesPerPole int)
NewNewtonBasin(degree, maxIter int)

// Wave 9: Topological Contours, Streamline Engraving, 3D Origami & Bauhaus SDF Relief
NewEvenStreamlines(dSep, dTest float64)
NewMarchingIsolines(levels int, noiseScale float64)
NewOrigamiTessellation(cols, rows int, foldDepth float64)
NewMondrianSubdivision(depth int, pad float64)

// Wave 8: Decorative Arts, Cloisonné Enamel & Printmaking
NewCloisonneEnamel(rings, petals int)
NewWoodblockWave(ridgeLayers, waveLines int)
NewArtDecoSunburst(rayCount, archTiers int)
NewStainedRoseWindow(petals, rings int)

// Wave 7: Anisotropic Flow Mosaic & Low-Poly Crystal Series
NewMosaicVortex(ringSteps int, brickScale float64)
NewMosaicWave(waveCourses int, brickScale float64)
NewMosaicCanyon(strataSteps int, brickScale float64)
NewMosaicLotus(petals int, brickScale float64)
NewMosaicSun(ringSteps int, brickScale float64)

// Wave 6: G-Buffer Impasto, Voxel Terrain & Geometric Replications
NewImpastoFlow(strokeCount, steps int, fieldScale float64)
NewVoxelTerrain(gridRes int, heightScale, contrast float64)
NewPackedCapsules(capsuleCount int, maxScale, minScale float64)
NewOrphicDiscs(discCount int)
NewNautilusSpiral(turns int, chambersPerTurn int)
NewTangentCones(coneCount int)

// Wave 5: Hyperbolic Geometry, Quasicrystals & 3D DoF Chaos
NewPoincareDisk(p, q, depth int)
NewSuminagashi(dropCount, combPasses int, swirlStrength float64)
NewAizawaAttractor(iterations int, dofStrength float64)
NewPenroseTiling(subdivisions int)
NewIsometricCity(gridSize int, density float64)
NewDLAFrost(particleCount int, stickiness float64)

// Wave 1–4: HDR Flow, Strange Attractors, Morphogenesis & Sacred Geometry
NewCurlFlow(lineNum, stepNum int, noiseScale float64)
NewSandRing(ringNum, grainsPerStep int, noiseScale float64)
NewWatercolor(shapeNum, layers, subdivisions int, roughness float64)
NewGuilloche(curveNum int, majorR, minorR, penOffset float64)
NewCliffordNebula(iterations int, a, b, c, d float64)
NewPhysarum(agentNum, steps int, sensorAngle, turnAngle float64)
NewDifferentialGrowth(steps, maxNodes int, minDist, maxDist, repulsionRadius float64)
NewChladni(particleNum int, m, n, a, b float64)
NewVoronoiCrystal(cellNum, relaxSteps int)
NewInkLandscape(layerNum int, roughness float64)
NewGrayScott(steps int, feed, kill float64)
NewLissajousRibbon(curves int, freqX, freqY, freqZ float64)
NewApollonianGasket(maxDepth int, minRadius float64)
NewSpaceColonization(attractorNum, steps int, stepDist, killDist float64)
NewTruchetFlow(baseGrid, maxDepth int)
NewAuroraWave(curtainNum, ribbonSteps int, turbulence float64)
NewFractalFlame(iterations, symmetry int)
NewJoyDivision(lineNum int, amplitude, frequency float64)
NewBoidsFlock(boidNum, steps int, perceptionRadius float64)
NewIslamicStar(gridSize int, contactAngle, strapWidth float64)
NewDelaunayStained(siteNum int, jitter float64)
NewPhyllotaxis(seedCount int, divergenceAngle, scaleFactor float64)

// Classic Sketchbook Generators
NewCircleLine(step float64, lineNum int, radius, xaixs, yaixs float64)
NewCircleLoop(radius float64)
NewMaze(step int)
NewRandCicle(mc, msp int, minStep, maxStep, minr, maxr float64, isRandColor bool)
NewSilkSky(circleNum int, sunRadius float64)
NewSilkSmoke(mc, msp int, minStep, maxStep, minRadius, maxRadius float64, isRandColor bool)
NewSpiralSquare(squareNum int, rectSide, decay float64, randColor bool)
NewSwirl(a, b, c, d, xaixs, yaixs float64)
NewDotLine(n int, ras, canv float64, randColor bool)
NewPointRibbon(r float64)
NewJanus(n int, decay float64)
NewRandomShape(shapeNum int)
NewColorCircle(circleNum int)
NewColorCircle2(circleNum int)
NewCircleGrid(circleNumMin, circleNumMax int)
NewContourLine(lineNum int)
NewNoiseLine(n int)
NewCircleLoop2(depth int)
NewPixelHole(dotN int)
NewDotsWave(dotsN int)
NewCircleMove(circleNum int)
NewCircleNoise(dotsN, colorMin, colorMax int)
NewYarn(n int)
NewBlackHole(circleN int, density, circleGap float64)
NewColorCanve(seg float64)
NewDomainWrap(scale, scale2, xOffset, yOffset float64, cmap ColorMapping)
NewPerlinPerls(circleN, dotsN, colorMin, colorMax int)
NewOceanFish(fishN, seg int)
NewJulia(fn EvalFunc, maxz float64, x, y float64)
```

---

## Contributing & Acknowledgements

Contributions, issues, and PRs are welcome! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) and the full architectural roadmap in [docs/roadmap.md](./docs/roadmap.md). If you find this package useful, you can also support it via [PayPal](https://www.paypal.me/jdxyw).

Inspired by works and articles from:
- <https://inconvergent.net/>
- <https://fronkonstin.com/>
- <https://www.iquilezles.org/www/articles/warp/warp.htm>
- <https://github.com/aschinchon/cyclic-cellular-automata>
- <https://github.com/armdz/ProcessingSketchs>
- <https://github.com/Mr-Slesser/Generative-Art-And-Fractals>
- <https://github.com/cdr6934/Generative-Processing-Experiments>
- <https://github.com/pkd2512/inktober2017>
- <http://paulbourke.net/fractals/peterdejong/>
