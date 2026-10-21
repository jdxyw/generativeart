# GenerativeArt Engine Architecture & Future Roadmap

This document serves as the architectural blueprint for `generativeart`, covering:
1. **Core Engine Refactoring & Performance Upgrades (Package 1 — Implemented)**
2. **Computational Geometry, Topological Fields & 2D SDF Engine (Package 2 — Implemented)**
3. **Future Platform Foundations (Dual Vector/HDR Backend, `DrawTimelapse`, WASM & CLI)**
4. **Next-Wave `arts/` Visual Paradigms (Directions 2–5)**

---

## Part I. Core Engine Architecture & Performance Upgrades (Package 1)

### 1.1 Deterministic Per-Canvas RNG & Reproducibility (`generativeart.go`)
- **Problem**: Relying solely on global `math/rand` prevents deterministic seed replay across concurrent goroutines and causes lock contention when rendering multi-panel galleries.
- **Architecture**:
  - `Canva` and `Options` maintain a dedicated deterministic seed (`seed int64`) and lock-free local generator (`rng *rand.Rand`).
  - APIs:
    - `c.SetSeed(seed int64)`: Re-seeds the canvas RNG deterministically.
    - `c.Seed() int64`: Returns the active seed for recipe inspection.
    - `c.Rand() *rand.Rand`: Returns the canvas-scoped `*rand.Rand` instance.
    - `c.Noise() *common.NoiseEngine`: Lazily constructs or returns a `common.NoiseEngine` seeded from `c.Seed()`.

### 1.2 64-Bit Linear HDR Hot-Path Acceleration (`render/canvas.go`, `render/postprocess.go`)
- **O(1) sRGB-to-Linear Lookup Table (`srgb8ToLinearLUT`)**:
  - Replaces millions of per-pixel `math.Pow((c+0.055)/1.055, 2.4)` calls with an exact 256-entry `float64` lookup table initialized at startup (`SRGB8ToLinear(b uint8) float64`).
- **Pre-Cached Perceptual Stops in `MultiStopGradient`**:
  - `NewMultiStopGradient` pre-computes and caches both `Oklab` (`oklabStops []Oklab`) and `LinearRGB` (`linearStops [][4]float64`) coordinates for every stop.
  - `SampleOklab(t)` performs direct linear interpolation in `(L, a, b)` followed by a single `OklabToRGBA` conversion, eliminating redundant `SRGBToLinear` and `math.Cbrt` evaluations on the endpoints.
- **Multi-Core Parallel Row Workers (`parallelRows`)**:
  - Full-canvas pixel passes (`ResolveToRGBA`, `ResolveToneMapped`, `ApplyBloom`, `ApplyPaperTexture`, `ApplyVignette`) automatically partition rows across `runtime.NumCPU()` worker goroutines when canvas area exceeds 16,384 pixels.

### 1.3 Seam-Free `DrawHatchPatch` & Normalized Bow Geometry (`render/advanced_primitives.go`)
- **Problem**: Splitting hatched engraving curves into 4 sub-paths with `RoundCap` produced visible semi-circular alpha-overlap seams at $t \in \{0.25, 0.50, 0.75\}$; multiplying `bow` directly by the unnormalized chord perpendicular caused disproportionate bowing on wide spans.
- **Fix**:
  - Detects when `colorFn(s, t)` is invariant along $t$ and strokes a single continuous path with zero intermediate joints.
  - When color varies along $t$, uses `ButtCap` with a `0.18px` tangent overlap so transitions are invisible.
  - Supports both fractional chord bowing ($|\text{bow}| \le 0.5$) and explicit pixel-space normal bowing ($|\text{bow}| > 0.5$).

---

## Part II. Computational Geometry, Topological Fields & 2D SDF Engine (Package 2)

### 2.1 4D Simplex Noise & Seamless Time-Loop Fields (`common/noise.go`)
- **`Simplex4D(x, y, z, w float64) float64`**:
  - Full 4D Simplex lattice implementation over 5-cell pentatopes ($F_4 = (\sqrt{5}-1)/4$, $G_4 = (5-\sqrt{5})/20$) with 32 4D gradient vectors.
- **`FBM4DLoop(x, y, phase, loopRadius float64, octaves int, lacunarity, gain float64) float64`**:
  - Samples 4D FBM along a closed circle in the $(z, w)$ plane:
    $$(z(\phi), w(\phi)) = (R \cos(2\pi \phi),\, R \sin(2\pi \phi)), \quad \phi \in [0, 1)$$
    guaranteeing $C^\infty$ periodicity at $\phi = 0$ and $\phi = 1$ for seamless animations.
- **`Curl4DLoop(x, y, phase, loopRadius, eps float64) Vector`**:
  - Divergence-free 2D velocity field evolving smoothly over a periodic time loop $\phi \in [0, 1)$.

### 2.2 Jobard-Lefer Evenly-Spaced Streamlines & Marching Squares (`render/field_engine.go`)
- **`TraceEvenStreamlines(cfg StreamlineConfig, fieldFn VectorField2D) [][]Point`**:
  - Implements Jobard & Lefer's algorithm for evenly-spaced streamlines:
    1. Integrates bidirectional streamlines using 4th-order Runge-Kutta (RK4).
    2. Enforces minimum separation $d_{\text{sep}}$ for seeding and $d_{\text{test}} = \alpha d_{\text{sep}}$ during integration via `SpatialGrid`.
    3. Generates candidate seeds at $\pm d_{\text{sep}}$ along the local normal of accepted streamlines until the domain is uniformly saturated without collisions.
- **`ExtractMarchingSquares(nx, ny int, w, h, isoVal float64, fieldFn ScalarField2D) [][]Point`**:
  - Evaluates `fieldFn` on an $(n_x+1) \times (n_y+1)$ grid, performs exact sub-cell linear root interpolation along cell edges, resolves saddle ambiguities via center-point sampling, and stitches segment fragments into continuous ordered polylines (`[][]Point`).
- **`ExtractIsoBandPolygons` & `RenderTopographicTerraces`**:
  - Builds layered contour elevation bands with optional cast drop shadows and Oklab rim highlights.

### 2.3 2D Signed Distance Field (SDF) & Smooth Boolean Engine (`render/sdf2d.go`)
- **Analytical Primitives (`SDF2D = func(p Point) float64`)**:
  - `SDFCircle`, `SDFBox`, `SDFRoundedBox`, `SDFSegment`, `SDFPolygon` (exact signed distance to arbitrary closed polygon), `SDFStar`, `SDFHeart`.
- **CSG & Smooth Boolean Operators**:
  - `SDFUnion`, `SDFIntersect`, `SDFSubtract`, `SDFSmoothUnion(a, b, k)`, `SDFSmoothSubtract(a, b, k)`, `SDFOnion(sdf, thickness)`.
- **Physical & Graphic SDF Shaders on `*Canvas`**:
  - `SDFGradient(sdf, p, eps)`: Computes unit surface normal $\nabla d / \|\nabla d\|$.
  - `RenderSDFDropShadow(sdf, offset, blurRadius, spread, col, alpha)`: Renders analytical soft penumbra drop shadows.
  - `RenderSDFBevel(sdf, baseCol, rimCol, specCol, bevelWidth, lightDir, alpha)`: Renders 3D beveled enamel/relief shapes with directional specular highlights directly from any composite 2D SDF.

---

## Part III. Platform & Export Foundations (Package 3 — Implemented)

### 3.1 Dual Vector (`SVG` / `PDF`) & HDR Raster Backend (`render/vector_bridge.go`, `platform.go`)
- Automatically records vector path operations (`RasterizePath`, `StrokePath`, `FillPolygon`, `DrawMosaicBrick`, `DrawCloisonneCell`, `DrawStainedGlassCell`, `DrawTaperedSpline`, `DrawHatchPatch`, etc.) alongside the 64-bit HDR pixel buffer.
- Exposes `c.ToSVG(path)`, `c.WriteSVG(w)`, `c.ToPDF(path)`, and `c.WritePDF(w)` on `*Canva` (and `SaveSVG` / `SavePDF` on `*render.Canvas`) for infinite-resolution archival printing and pen plotters, with automatic raster embedding fallback for pure pixel-splat engines.

### 3.2 `DrawTimelapse` Animation & PNG `tEXt` Recipe Metadata (`platform.go`, `cmd/genart`)
- **`AnimatedEngine` & `DrawTimelapse` / `ToGIF`**:
  - Supports phase-driven seamless loop animations (`phase in [0, 1)`) via `AnimatedEngine` and `NewAnimatedEngine(...)`, as well as multi-seed timelapse sequences for any standard `Engine`, exporting adaptive-palette Floyd-Steinberg dithered looping GIFs via `c.ToGIF(path, delayCs)`.
- **PNG `tEXt` Recipe Metadata (`GenerativeArt-Recipe`)**:
  - `c.ToPNG(path)` automatically embeds a standard PNG `tEXt` chunk containing the JSON recipe (`engine`, `width`, `height`, `seed`, `background`, `params`), inspectable via `ReadPNGRecipe(path)` or `genart inspect <file.png>`.
- **Standalone CLI (`cmd/genart`)**:
  - `go run ./cmd/genart list`
  - `go run ./cmd/genart render -art cloisonneenamel -w 1200 -h 1200 -seed 2026 -out cloisonne.svg` (supports `.png`, `.svg`, `.pdf`, `.gif`, `.jpg`)
  - `go run ./cmd/genart inspect images/mosaicsun.png`

### 3.3 Future Extension: Pure-Go WebAssembly (WASM) Interactive Gallery Playground
- Compile `wasm/main.go` (`GOOS=js GOARCH=wasm`) so users can tweak seeds, palettes, and parameters live inside `index.html`.

---

## Part IV. Next-Wave `arts/` Visual Paradigms (Directions 2–5)

With Packages 1 & 2 in place, the following 12 generators are unlocked:

### Direction 2: Topological Contours, Even Streamlines & Field Engraving
1. **`EvenStreamlines`** (`NewEvenStreamlines(dSep, dTest float64)`): Jobard-Lefer collision-free currency engraving & flow maps.
2. **`MarchingIsolines`** (`NewMarchingIsolines(levels int, noiseScale float64)`): Marching Squares domain-warped karst terraces with 3D cast shadows.
3. **`MagneticDipole`** (`NewMagneticDipole(polesN, linesPerPole int)`): Multi-pole electromagnetic field lines and orthogonal equipotential rings.

### Direction 3: 2.5D Origami, Quasicrystals & Neoplasticism
4. **`OrigamiTessellation`** (`NewOrigamiTessellation(cols, rows int, foldDepth float64)`): Miura-ori / Ron Resch 3D folded paper with Lambertian shading & valley AO.
5. **`MondrianSubdivision`** (`NewMondrianSubdivision(depth int, pad float64)`): Bauhaus / De Stijl recursive golden-ratio subdivision with SDF drop shadows.
6. **`AmmannBeenker`** (`NewAmmannBeenker(depth int)`): 8-fold octagonal quasicrystal tiling with beveled metallic/enamel facets.

### Direction 4: Complex Dynamics, Vortex Fluids & Chemical Waves
7. **`VortexFluid`** (`NewVortexFluid(vorticesN, steps int)`): Point-vortex Jupiter atmospheric bands and Kármán vortex streets.
8. **`NewtonBasin`** (`NewNewtonBasin(degree, maxIter int)`): Complex Newton-Raphson root-finding fractal rosettes.
9. **`Buddhabrot`** (`NewBuddhabrot(samplesN, maxIter int)`): 3-channel spectral Nebulabrot escaping trajectory density cloud.
10. **`CyclicAutomata`** (`NewCyclicAutomata(states, steps int)`): Belousov-Zhabotinsky spiral chemical reaction waves.

### Direction 5: Traditional Textile Weave & Asemic Calligraphy
11. **`SashikoWeave`** (`NewSashikoWeave(gridN int)`): Indigo-dyed Japanese Sashiko geometric embroidery with 3D twisted thread shading.
12. **`AsemicCalligraphy`** (`NewAsemicCalligraphy(cols, glyphsPerCol int)`): Spring-mass cursive brush scroll with flying-white (`Feibai`) dry-brush physics and vermilion seal.
