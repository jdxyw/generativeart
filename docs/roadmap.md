# GenerativeArt Future Roadmap

This document outlines the next-stage architectural and artistic directions for `generativeart`, building on the 64-bit Linear HDR rendering engine (`render/`), perceptual color/noise foundation (`common/`), and 53 generative art engines (`arts/`).

---

## 1. Dynamic Evolution & Seamless Loop Animation (`DrawTimelapse`)

### Motivation
`generativeart.go` currently reserves an unimplemented `DrawTimelapse(e Engine)` method. Many algorithms in this repository—such as `Physarum`, `DifferentialGrowth`, `GrayScott`, `BoidsFlock`, `SpaceColonization`, and `CurlFlow`—are inherently time-evolving dynamical systems.

### Proposed Design
- **`AnimatedEngine` / Frame Hook Interface**:
  ```go
  type FrameRecorder func(frameIdx int, img image.Image)

  type AnimatedEngine interface {
      Engine
      GenerativeFrames(c *Canva, totalFrames int, record FrameRecorder)
  }
  ```
- **4D Simplex / Curl Noise Seamless Loops**:
  - Sample 4D noise along a closed circle in $(z, w)$ space:
    $$(z(t), w(t)) = (R \cos(2\pi t), R \sin(2\pi t)), \quad t \in [0, 1)$$
    producing infinitely looping animations where frame $N$ transitions seamlessly back to frame $0$.
- **Export Targets**:
  - Built-in animated GIF (`image/gif` with Oklab median-cut / K-means palette quantization), APNG, and frame sequence / FFmpeg pipe export.

---

## 2. Pure-Go WebAssembly (WASM) Interactive Playground

### Motivation
Because `generativeart` and `render/` are implemented in 100% pure Go (no CGo dependencies), the entire package compiles cleanly to WebAssembly via `GOOS=js GOARCH=wasm`.

### Proposed Design
- **Browser-Side Real-Time Generator**:
  - Compile a lightweight `wasm/main.go` binary exposing a JS function `renderArt(configJSON)` that returns a PNG/RGBA buffer directly to an HTML5 `<canvas>`.
- **Interactive Controls in `index.html`**:
  - Add a **"Live Generate"** tab in the gallery lightbox modal allowing users to tweak parameters (e.g., `lineNum`, `noiseScale`, attractor constants, `seed`), switch palettes (`Cyberpunk`, `Dunhuang`, `Inferno`, `UkiyoE`), render a custom artwork live in the browser, and download high-resolution PNGs.

---

## 3. CLI Tool (`cmd/genart`) & Reproducible Seed Metadata

### Motivation
Users and generative artists often want to batch-generate artworks from the terminal, explore random variations, and reproduce a specific masterpiece at 4K/8K print resolution later.

### Proposed Design
- **Standalone CLI (`cmd/genart`)**:
  ```bash
  go install github.com/jdxyw/generativeart/cmd/genart@latest
  genart list
  genart render --art=cliffordnebula --palette=inferno --size=3840x2160 --seed=20261019 -o=nebula_4k.png
  ```
- **Deterministic RNG & PNG `tEXt` Metadata**:
  - Add `c.SetSeed(seed int64)` to `Canva` so all engines draw from a deterministic `*rand.Rand` instance instead of global `math/rand`.
  - Automatically embed a JSON recipe (`art`, `params`, `seed`, `palette`, `version`) into a PNG `tEXt` chunk (`Software` / `GenerativeArt-Recipe`) on `ToPNG()`, and provide `genart inspect <file.png>` / `genart replay <file.png> --size=7680x4320` to re-render any saved PNG at higher resolution.

---

## 4. Image-Guided & Typography-Driven Generative Art

### Motivation
Combining autonomous generative systems with an input reference image or text mask bridges abstract mathematical art and recognizable portraiture/poster design.

### Proposed Design
- **Reference Field API (`render.ImageField`)**:
  - Extract luminance $L(x,y)$, Oklab color, Sobel/Scharr gradient $\nabla L$, and **Edge-Tangent Flow (ETF)** vectors from any input `image.Image` or rendered typography mask.
- **New Image-Guided Art Engines**:
  - **`WeightedVoronoiStipple`**: Secord's algorithm weighting Lloyd's Voronoi relaxation by image darkness $1 - L(x,y)$ for stippled ink portraits.
  - **`TSPArt`**: Connects weighted stipple points into a single continuous non-crossing tour using 2-opt relaxation.
  - **`ImpressionistBrush` / `ImageFlow`**: Drives `DrawBristleStroke`, `DrawWatercolorPolygon`, and `DrawTaperedSpline` along the image's Edge-Tangent Flow using sampled Oklab colors.

---

## 5. Infinite-Resolution Vector Export (SVG / PDF) & Pen Plotter Utilities

### Motivation
`render/` already integrates `github.com/tdewolff/canvas`. Exposing native vector output enables large-format archival printing and hardware pen plotting (e.g., AxiDraw).

### Proposed Design
- **Vector Output Methods on `Canva` / `render.Canvas`**:
  - `ToSVG(fpath string) error` and `ToPDF(fpath string) error`.
- **Plotter-Oriented Algorithms**:
  - **Jobard-Lefer Evenly-Spaced Streamlines**: Guarantees a minimum separation distance $d_{\text{sep}}$ between streamlines so lines never collide or pool ink.
  - **2.5D / 3D Hidden-Line Removal (HLR)**: Analytic silhouette and polygon occlusion for wireframe terrains, architectural cities, and 3D parametric surfaces.
  - **TSP Path Optimization**: Sorts stroke paths to minimize pen-up travel distance.

---

## 6. 3D Signed Distance Field (SDF) Raymarching & Mathematical Sculptures

### Motivation
The 64-bit Linear HDR buffer (`render.Canvas`) is ideally suited for physically based shading, soft shadows, subsurface scattering, and bloom in 3D mathematical art.

### Proposed Design
- **Pure-Go Parallel SDF Raymarcher (`render/sdf.go`)**:
  - Multi-threaded tile renderer using `runtime.NumCPU()` goroutines.
  - Supports smooth boolean operations (`SmoothUnion`, `SmoothSubtraction`, `SmoothIntersection`), domain repetition, twist/bend operators, ambient occlusion (AO), soft penumbra shadows, and Oklab subsurface scattering.
- **New 3D Generative Art Engines**:
  - **`Mandelbulb` / `JuliaQuaternion`**: 3D fractal sculptures with volumetric glow.
  - **`GyroidMinimalSurface`**: Triply periodic minimal surfaces (Gyroid, Schwarz P) with iridescent thin-film interference.
  - **`Apollonian3D`**: 3D Kleinian / Apollonian sphere packing foams.

---

## 7. New `arts/` Algorithms & Visual Paradigms Roadmap

Beyond engine-level features, the `arts/` package can expand into the following 6 mathematical and artistic families (20+ concrete new generators):

### 7.1 Advanced Flow Fields & Topological Contours
| Proposed Art Type | Constructor Signature | Core Algorithm & Visual Aesthetic |
| :--- | :--- | :--- |
| **`EvenStreamlines`** | `NewEvenStreamlines(dSep, dTest float64)` | **Jobard & Lefer Evenly-Spaced Streamlines** with RK4 integration and spatial distance grid; lines never collide or bunch up, creating currency-grade engraving and topographic flow maps. |
| **`MarchingIsolines`** | `NewMarchingIsolines(levels int, scale float64)` | **Marching Squares** exact topological contour extraction over domain-warped scalar fields with elevation Oklab color bands and shaded relief. |
| **`MagneticDipole`** | `NewMagneticDipole(polesN, linesPerPole int)` | Multi-pole electromagnetic vector field lines and orthogonal equipotential contours rendered with HDR plasma glow. |

### 7.2 Strange Attractors, 3D Chaos & Complex Dynamics
| Proposed Art Type | Constructor Signature | Core Algorithm & Visual Aesthetic |
| :--- | :--- | :--- |
| **`AizawaAttractor` / `LorenzRibbon`** | `NewAizawaAttractor(steps int)` | 3D chaotic attractors (Aizawa, Lorenz, Thomas, Halvorsen) projected with **Depth-of-Field (DoF) Bokeh Splatting** (`DrawGlowDot` radius proportional to $\|z - z_{\text{focal}}\|$). |
| **`Buddhabrot`** | `NewBuddhabrot(samplesN, maxIter int)` | **Nebulabrot / Buddhabrot** probability density accumulation of escaping Mandelbrot trajectories across 3 iteration-length spectral channels. |
| **`NewtonBasin`** | `NewNewtonBasin(degree, maxIter int)` | Complex **Newton-Raphson root-finding fractals** ($z_{n+1} = z_n - p(z_n)/p'(z_n)$) with smooth exponential convergence shading and basin boundary glow. |
| **`LyapunovFractal`** | `NewLyapunov(seq string, rMin, rMax float64)` | **Markus-Lyapunov** logistic stability exponent landscapes rendered in metallic gold and deep cobalt. |

### 7.3 Non-Euclidean Geometry & Aperiodic Quasicrystals
| Proposed Art Type | Constructor Signature | Core Algorithm & Visual Aesthetic |
| :--- | :--- | :--- |
| **`PoincareDisk`** | `NewPoincareDisk(p, q, depth int)` | Escher-inspired **Hyperbolic $\{p, q\}$ Tessellations** inside the Poincaré Disk using exact Möbius transformations and geodesic circular arcs. |
| **`PenroseTiling`** | `NewPenroseTiling(subdivisions int)` | Robinson-triangle deflation **Penrose P2/P3 Quasicrystal Tilings** (Kite-and-Dart / Rhombus) with De Bruijn pentagrid interwoven ribbons. |
| **`AmmannBeenker`** | `NewAmmannBeenker(depth int)` | 8-fold octagonal aperiodic quasicrystal tessellation with cloisonné watercolor washes and golden bevels. |

### 7.4 Morphogenesis, Cellular Automata & Physical Simulations
| Proposed Art Type | Constructor Signature | Core Algorithm & Visual Aesthetic |
| :--- | :--- | :--- |
| **`DLAFrost`** | `NewDLAFrost(particlesN int, anisotropy float64)` | Anisotropic **Diffusion-Limited Aggregation (DLA)** generating crystalline ice dendrites, Lichtenberg lightning figures, and mineral arborizations. |
| **`CyclicAutomata`** | `NewCyclicAutomata(states, threshold, steps int)` | Greenberg-Hastings / **Cyclic Cellular Automata** excitable media producing *Belousov-Zhabotinsky* chemical reaction spiral waves. |
| **`VortexFluid`** | `NewVortexFluid(vorticesN, steps int)` | **Point-Vortex / Kelvin-Helmholtz Fluid Advection** simulating Kármán vortex streets and Jupiter-like atmospheric bands advecting Oklab dye tracers. |
| **`CircuitWFC`** | `NewCircuitWFC(gridW, gridH int)` | Constraint-propagated **Wave Function Collapse (WFC)** synthesizing intricate circuit-board traces, Celtic knots, and architectural pipe labyrinths. |

### 7.5 Architectural, 2.5D Axonometric & Recursive Subdivision
| Proposed Art Type | Constructor Signature | Core Algorithm & Visual Aesthetic |
| :--- | :--- | :--- |
| **`MondrianSubdivision`** | `NewMondrianGrid(depth int, pad float64)` | Recursive KD-tree / Golden-ratio rectangular subdivision with Bauhaus/De Stijl textures, inner drop shadows, and procedural micro-patterns per cell. |
| **`IsometricCity`** | `NewIsometricCity(gridN int, heightScale float64)` | *Monument Valley*-inspired **2.5D Axonometric Architecture** with procedural towers, arches, staircases, and directional sun/shadow casting. |
| **`OrigamiTessellation`** | `NewOrigamiCrease(cols, rows int, fold float64)` | **Miura-ori & Ron Resch 3D Origami** folded paper triangulations with analytic Lambertian shading, valley ambient occlusion, and cotton paper grain. |

### 7.6 Natural Media, Algorithmic Marbling & Calligraphic Script
| Proposed Art Type | Constructor Signature | Core Algorithm & Visual Aesthetic |
| :--- | :--- | :--- |
| **`Suminagashi`** | `NewSuminagashi(dropsN, tinesN int)` | **Japanese Floating-Ink Marbling** using exact area-preserving fluid drop expansion ($p' = c + (p - c)\sqrt{1 + r^2/\|p-c\|^2}$) and comb-tine rake advection. |
| **`AsemicCalligraphy`** | `NewAsemicScript(cols, glyphsPerCol int)` | Generative **Asemic Cursive Ink Scroll** driven by spring-mass brush dynamics (`DrawBristleStroke` + `DrawTaperedSpline`) on aged Xuan paper. |
| **`SashikoWeave`** | `NewSashikoWeave(gridN int)` | Procedural Japanese **Sashiko & Jacquard Textile** simulation rendering individual twisted indigo/ivory threads with micro-shadow depth. |

---

## Suggested Implementation Priority

| Phase | Focus | Key Deliverables |
| :--- | :--- | :--- |
| **Phase 1** | Signature New `arts/` Wave 4 | `PoincareDisk`, `PenroseTiling`, `Suminagashi`, `AizawaAttractor` (3D DoF), `IsometricCity`, `DLAFrost` |
| **Phase 2** | Reproducibility & CLI | Deterministic `SetSeed(int64)`, PNG recipe metadata, `cmd/genart` CLI |
| **Phase 3** | WebAssembly Live Playground | `wasm/` build target + live parameter sliders in `index.html` |
| **Phase 4** | Animation (`DrawTimelapse`) | `AnimatedEngine`, seamless 4D noise loops, GIF/APNG export |
| **Phase 5** | Image-Guided, Plotter & 3D SDF | `WeightedVoronoiStipple`, `EvenStreamlines`, `ToSVG`/`ToPDF`, `Mandelbulb` SDF raymarcher |

