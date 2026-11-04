# GenerativeArt Engine Architecture & Future Roadmap

This document serves as the architectural blueprint for `generativeart`, covering:
1. **Core Engine Refactoring & Performance Upgrades (Package 1 — Implemented)**
2. **Computational Geometry, Topological Fields & 2D SDF Engine (Package 2 — Implemented)**
3. **Platform & Export Foundations (Package 3 — Implemented: Dual Vector/HDR Backend, `DrawTimelapse` HD GIF, PNG `tEXt` Recipes & `cmd/genart` CLI)**
4. **Waves 9–22 `arts/` Visual Paradigms (140 Art Engines + 29 HD 600×600 Loop GIFs — Implemented)**
5. **Exhibition Poster & ContactSheet Engine (`poster.go` — Implemented) & HD GIF Expansion**
6. **Next-Generation Architecture & Platform Roadmap (6 Directions, with Directions 4 & 5 Implemented)**
7. **Art-Empowering Rendering & Mathematical Foundations (Part VII & Part VIII — Implemented)**
8. **Wave 23 Master Architecture, Celestial Mechanics & Sci-Fi Megastructures + Part IX–XI Future Roadmap**
9. **Part X. Underlying Infrastructure Audit & 6-Phase Core Engine Upgrade Roadmap (Phases 1–5 — Implemented)**

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

### 2.1 4D Simplex Noise, Analytical-Derivative Gradient Perlin, Erosion FBM & Gabor Fields (`common/noise.go`, `common/perlinnoise.go`)
- **`PerlinNoise` C² Quintic & Negative-Domain Upgrade (`common/perlinnoise.go`, `c.Perlin()`)**:
  - Replaced $C^1$ `math.Cos` interpolation with Ken Perlin's $C^2$ quintic polynomial $u(t) = 6t^5 - 15t^4 + 10t^3$ (`4×–6×` faster, zero second-derivative Mach bands), fixed negative-coordinate mirroring via mathematical `fastFloor` + bitwise wrapping, added `SetDetail(octaves, falloff)`, and migrated all 12 legacy `PerlinNoise` engines to deterministic `c.Seed()`.
- **Analytical Derivative Gradient Perlin & Curl (`Perlin2D`, `Perlin3D`, `Perlin2DDeriv`, `FBM2DDeriv`, `Curl2DAnalytical`)**:
  - Evaluates $C^2$ 2D gradient Perlin noise and its exact analytical partial derivatives $\left(\frac{\partial N}{\partial x}, \frac{\partial N}{\partial y}\right)$ in a single pass using $u'(t) = 30t^2(t-1)^2$, powering single-pass divergence-free `Curl2DAnalytical`.
- **Inigo Quilez Gradient-Damped Erosion FBM (`ErosionFBM2D`)**:
  - Accumulates analytical octave gradients $\mathbf{d} = \sum \nabla n_i$ and damps high-frequency octaves by $\bigl(1 + \text{erosionStrength}\cdot \|\mathbf{d}\|^2\bigr)^{-1}$, naturally carving eroded alpine gullies on steep slopes while retaining micro-detail in valleys.
- **Seamless Tileable Perlin/FBM, Iterative Domain Warping, Flow Noise & Anisotropic Gabor Noise**:
  - `TileablePerlin2D` & `TileableFBM2D(x, y, periodX, periodY, ...)`: exact toroidal lattice wrapping for seamless patterns.
  - `DomainWarp2D(x, y, warpStrength, octaves)`: two-stage $f(\mathbf{p} + w\,\mathbf{q}(\mathbf{p} + w\,\mathbf{r}(\mathbf{p})))$ returning both warped scalar and intermediate warp vectors $(q_x, q_y, r_x, r_y)$.
  - `FlowNoise2D(x, y, spinPhase, octaves)`: Perlin-Neyret rotating-gradient swirling turbulence with exact $C^\infty$ periodicity over `spinPhase in [0, 1)`.
  - `GaborNoise2D(x, y, freq, orientationRad, bandwidth)`: Lagae et al. sparse-convolution anisotropic harmonic wavelet noise.
- **`Simplex4D`, `FBM4DLoop` & `Curl4DLoop`**:
  - Full 4D Simplex lattice implementation over 5-cell pentatopes ($F_4 = (\sqrt{5}-1)/4$, $G_4 = (5-\sqrt{5})/20$) with 32 4D gradient vectors, sampling closed circles in $(z, w)$ for seamless time loops $\phi \in [0, 1)$.

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
- **Standalone CLI (`cmd/genart`) & Unified Registry (`arts/registry.go`)**:
  - `go run ./cmd/genart list`
  - `go run ./cmd/genart render -art cloisonneenamel -w 1200 -h 1200 -seed 2026 -out cloisonne.svg` (supports `.png`, `.svg`, `.pdf`, `.gif`, `.jpg`)
  - `go run ./cmd/genart inspect images/mosaicsun.png`
  - `go run ./cmd/genart replay images/mosaicsun.png -w 2400 -h 2400 -out mosaicsun_2400.png` (reconstructs exact engine & parameters from PNG `tEXt` recipe and re-renders at arbitrary resolution or vector `.svg`/`.pdf`)

### 3.3 Modern 2D/HDR Pipeline Stack (`render/pipeline.go`, `generativeart.go`)
- **Persistent Primary HDR Pipeline**: `Canva` maintains a persistent 64-bit linear HDR buffer across multi-pass `Draw` calls with lazy 8-bit `RGBA` resolution on `c.Img()`.
- **2D Affine Transform Matrix Stack**: `PushMatrix`, `PopMatrix`, `Translate`, `Rotate`, `RotateAbout`, `Scale`, `ScaleAbout`.
- **Sub-Pixel Stencil / Clip Mask Stack**: `PushClipPath`, `PushClipPolygon`, `PushClipCircle`, `PushClipSDF(sdf, invert)`, `PopClip`.
- **Independent HDR Child Layers & `sync.Pool` Rasterizer**: `c.NewLayer()` and `c.CompositeLayer(layer, mode, opacity)` with pooled `vector.Rasterizer` workspaces.

### 3.4 Future Extension: Pure-Go WebAssembly (WASM) Interactive Gallery Playground
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
12. **`AsemicCalligraphy`** (`NewAsemicCalligraphy(cols, glyphsPerCol int)`): Real Chinese running-cursive scroll (*Lantingji Xu*) with *Yongzi Bafa* brush physics, *Feibai* dry-brush splits, and vermilion seals.

### Wave 12: Cultural Heritage, 3D TPMS Raymarching & High-Voltage Physics
13. **`DunhuangApsara`** (`NewDunhuangApsara(ribbons, haloRings int)`): Tang-Dynasty Mogao caisson ceiling (*Zaojing*) with *Baoxianghua* lotus mandala and flying silk ribbons.
14. **`CelticKnotwork`** (`NewCelticKnotwork(gridN int, ribbonWidth float64)`): *Book of Kells* illuminated manuscript interlaced knotwork with alternating over-under crossings.
15. **`GyroidSurface`** (`NewGyroidSurface(scaleFreq, thickness float64)`): 3D SDF raymarched Schoen Gyroid minimal surface with 2×2 edge SSAA and two-sided ceramic glaze.
16. **`LichtenbergTree`** (`NewLichtenbergTree(generations int, branchSpread float64)`): Dielectric breakdown lightning tree inside a beveled sapphire acrylic block.
17. **`PersianTilework`** (`NewPersianTilework(petals, tiers int)`): 16-fold Isfahan *Muqarnas* stalactite vault dome with *Haft-Rangi* glazed tiles and *Islimi* vines.
18. **`ChampleveEnamel`** (`NewChampleveEnamel(symmetry, rings int)`): Warring States black-bronze mirror inlaid with gold/silver dragon-cloud scrolls and cabochon turquoise.

---

## Part V. Next-Phase Roadmap (Directions 1–4)

### 5.1 Direction 1: Wave 13 — Reaching the 100 Generative Art Engines Milestone (+6 Engines — Implemented)
Expand `arts/` from **94** to **100** generators with pure mathematical attractors, Kleinian limit sets, self-organized criticality, topological phase defects, number-theoretic Ford circles, and quantum chaos:
1. **`HalvorsenAttractor`** (`NewHalvorsenAttractor(pointsN int, alpha float64)`): 3-fold cyclic-symmetric 3D Halvorsen strange attractor nebula projected along $(1,1,1)$ with thin-lens DoF bokeh and 3-channel spectral velocity coloring.
2. **`HyperbolicFuchsian`** (`NewHyperbolicFuchsian(maxIter int, zoom float64)`): *Indra's Pearls* Kleinian/Fuchsian Schottky circle-inversion limit set with 2×2 SSAA and 3D spherical orbit-trap pearls.
3. **`AbelianSandpile`** (`NewAbelianSandpile(grains int)`): Bak-Tang-Wiesenfeld Abelian Sandpile self-organized criticality fractal and toppling odometer relief.
4. **`KuramotoWave`** (`NewKuramotoWave(defectsN, streamlines int)`): Coupled phase-oscillator topological defect field with branch-cut-free cyclic Oklab shading and Jobard-Lefer phase-gradient streamlines.
5. **`FordCircles`** (`NewFordCircles(maxDenom int)`): Conformal modular-group Farey-sequence and Ford-circle tangency gasket ($|ps - qr| = 1$).
6. **`QuantumBilliard`** (`NewQuantumBilliard(modesN int, waveNumber float64)`): Berry-Heller quantum chaos random-wave eigenfunction scars and Bohmian probability-current streamlines ($\mathbf{J} = \text{Im}(\Psi^* \nabla \Psi)$).

### 5.1b Wave 14 — Complex Phase Portraits, Differential Topology & Nonlinear Dynamics (+6 Engines, 106 Total — Implemented)
7. **`RiemannZeta`** (`NewRiemannZeta(termsN, phaseRays int)`): Complex phase portrait of $\zeta(s)$ across the critical strip $\text{Re}(s)=\frac{1}{2}$ with Borwein Chebyshev acceleration, non-trivial zero vortices, and Newton streamlines.
8. **`HopfFibration`** (`NewHopfFibration(toriN, fibersPerTorus int)`): 3D stereographic projection of the $S^3 \to S^2$ Hopf fibration into linked Villarceau circles sweeping nested Clifford tori.
9. **`ThomasAttractor`** (`NewThomasAttractor(pointsN int, dampingB float64)`): 3D cyclically symmetric Thomas labyrinth strange attractor manifold family ($\dot{x} = \sin y - b x$).
10. **`FaradayWaves`** (`NewFaradayWaves(symmetry int, harmonicScale float64)`): 12-fold dodecagonal quasicrystalline Faraday capillary standing waves with analytical 3D normal shading.
11. **`MobiusLoxodrome`** (`NewMobiusLoxodrome(arms, spheresPerArm int)`): Conformal double-spiral Möbius transformation of a hexagonal Doyle sphere lattice.
12. **`LagrangianCoherent`** (`NewLagrangianCoherent(advSteps int, ftlScale float64)`): Forward and backward Finite-Time Lyapunov Exponent (FTLE) ridges from the Cauchy-Green strain tensor $\mathbf{C} = (\nabla \Phi_0^T)^\top (\nabla \Phi_0^T)$.

### 5.1c Wave 15 — 3D Terrain Geomorphology, Hydraulic Erosion & Oceanographic Bathymetry (+5 Engines, 111 Total — Implemented)
13. **`HydraulicErosion`** (`NewHydraulicErosion(dropletsN int, erosionStrength float64)`): 32,000-droplet physical hydraulic erosion simulator carving dendritic alpine gullies, alluvial fans, and glacial rivers rendered with 2×2 SSAA floating-horizon 3D projection.
14. **`BasaltColumns`** (`NewBasaltColumns(hexRings int, pillarRelief float64)`): Giant's Causeway isometric 3D hexagonal basalt columnar jointing and volcanic magma fissure canyon.
15. **`SandDunes`** (`NewSandDunes(dunesN int, windShear float64)`): 2×2 SSAA 3D Aeolian barchan & transverse desert sand dunes with C⁰ cusp knife-edge crests, windward ripples, and leeward slip-face shadows.
16. **`BathymetricChart`** (`NewBathymetricChart(isobaths, seamounts int)`): 22-level abyssal subduction trench & volcanic island arc bathymetric chart with Marching Squares isobaths.
17. **`WireframeTerrain`** (`NewWireframeTerrain(gridRes int, reliefScale float64)`): 3D perspective triangulated alpine glacier & fjord mesh with 3D facet normal lighting and glowing wireframe edges.

### 5.1d Wave 16 — Order-to-Disorder Grids, Swirling Vortex Impasto, Tangential Pebble Filleting, Hexagonal Mandalas, Richter Squeegee & Sci-Fi Halftone Panels (+6 Engines + Nautilus Overhaul, 117 Total — Implemented)
18. **`ConcentricMorph`** (`NewConcentricMorph(gridCols, gridRows int, disorder float64)`): Vera Molnár-inspired order-to-disorder grid morphing terracotta concentric squares into ivory circles and scattered cerulean/teal clouds.
19. **`VortexImpasto`** (`NewVortexImpasto(vortexCount, strokeCount int, curlStrength float64)`): Ragged deckle-edged museum canvas with 22,000+ swirling micro-vortex bristle ribbons and vortex-eye specular pinpoints.
20. **`SlicedPebbles`** (`NewSlicedPebbles(gridSize, maxCuts int, cornerRadius float64)`): Mid-century $6 \times 6$ grid of squares sliced by half-plane chords with uniform mortar gaps and exact tangential circular-arc corner fillets.
21. **`HexMandala`** (`NewHexMandala(cols, rows int, colorScale float64)`): Flat-topped hexagonal honeycomb tessellation of $D_6 / D_{12}$ dihedral rosettes and hexagrams modulated by a cyan-nebula and burnished-gold macro field.
22. **`SqueegeeScrape`** (`NewSqueegeeScrape(passes, dropletCount int, scrapePressure float64)`): Gerhard Richter-style horizontal squeegee oil-paint drag over a diagonally striated underpainting with specular droplet splatter.
23. **`HalftonePanels`** (`NewHalftonePanels(panelCount int, dotSpacing float64)`): 22-panel retro sci-fi comic layout with ringed Saturns, crescent moons, and hexagonal Ben-Day halftone mountain ridges in mint, black, and white.
24. **`NautilusSpiral` Overhaul** (`NewNautilusSpiral(turns, linesPerChamber int)`): Upgraded to pure single-direction 3D fluted radial chambers, overlapping longitudinal copper/gold/steel-blue S-ribbon plumes, and continuous multi-strand spiral cables.

### 5.1e Wave 17 — Tactile Bristle & Vortex Impasto Series (+4 Engines + 4 HD Loop GIFs, 121 Total — Implemented)
25. **`NebulaPlumes`** (`NewNebulaPlumes(spiralArms, strokeCount int, swirlTightness float64)`): Macro-scale binary spiral galaxy collision with 2 counter-rotating super-vortices, 14 satellite eddies, and 23,000 feathery bristle plumes over antique parchment.
26. **`AbyssalCurrents`** (`NewAbyssalCurrents(jetBands, strokeCount int, meanderAmp float64)`): Obsidian-bordered meandering bioluminescent cyan Kuroshio jet stream flanked by 18 alternating neon rose-magenta and tangerine-gold Kármán vortex street eddies.
27. **`SolarProminence`** (`NewSolarProminence(archCount, strokeCount int, magneticTwist float64)`): Circular deckle-edged solar tondo medallion with 14 bipolar sunspot poles, twisted magnetic arcade loops, and coronal prominence plumes.
28. **`BotanicalWhorls`** (`NewBotanicalWhorls(whorlCount, strokeCount int, phylloAngle float64)`): 64 multi-scale Fibonacci golden-angle whorls cascading from giant lotus rosettes to micro-buds in Monet celadon-jade, sakura rose, and twilight pond indigo.

### 5.1f Wave 18 — LIC Jovian Magnetosphere, Thin-Film Iridescent Nacre, Kissing-Schottky Kleinian Pearls & 3D Bismuth Hopper Crystals (+4 Engines, 125 Total — Implemented)
29. **`JupiterMagnetosphere`** (`NewJupiterMagnetosphere(bands, vortices int, licLength float64)`): Juno perijove zonal shear bands, Great Red Spot super-vortex, and Kármán vortex street rendered with Fast Line Integral Convolution (`ComputeLICField`), Jobard-Lefer magnetic filaments, thin-film iridescent eyewalls, and 2.5D luminance relief.
30. **`IridescentNacre`** (`NewIridescentNacre(whorls, fiberCount int, filmThicknessNm float64)`): 3D abalone / mother-of-pearl logarithmic-spiral shell relief and 18,000 aragonite silk-bristle plumes seeded via blue-noise Poisson-disk + Lloyd CVT relaxation and shaded with `ResolveToCanvasAdvanced` (horizon soft shadows, jade SSS, Airy thin-film iridescence, and deckle border).
31. **`SchottkyKleinian`** (`NewSchottkyKleinian(maxDepth int, loxodromicTwist float64)`): 8-fold Kissing-Schottky Kleinian limit set with 3D iridescent orbit-trap pearls, `Circle2D.InvertCircle` gilded sphere trees, `PoincareGeodesicArc` astrolabe chords, `EvaluateDomainColoring` conformal grids, and radial chromatic dispersion.
32. **`BismuthHopper`** (`NewBismuthHopper(terraces, crystals int, oxideNm float64)`): 3D synthetic Bismuth (`[Bi]`) hopper crystal labyrinth with rectilinear stair-step cavities, 45° chamfered metallic bevels, dual-sun horizon soft shadows, and $\text{Bi}_2\text{O}_3$ Airy thin-film interference.

### 5.1g Wave 19 — Analytical Perlin & Gabor Quartet: Fidenza Flow Blocks, 3D Topographic Lacquer Terraces, 3D Perlin Ribbon Sculpture & Celestial Gabor-Curl Astrolabe (+4 Engines, 129 Total — Implemented)
33. **`PerlinMalachite`** (`NewPerlinMalachite(bands, octaves int, warpStrength float64)`): *Fidenza*-inspired Perlin flow-field curved color-block tapestry over warm cotton parchment with collision-free beam packing, 3D drop shadows, and gilded/ink borders.
34. **`ErosionCanyon`** (`NewErosionCanyon(octaves int, erosionStrength, strataFreq float64)`): 16-tier 3D laser-cut Perlin topographic paper/lacquer relief driven by `ErosionFBM2D` with directional paper-cut shadows and 24k gold Marching-Squares contour rims.
35. **`AnalyticalCurlSilk`** (`NewAnalyticalCurlSilk(streamlines, octaves int, curlScale float64)`): 3D depth-sorted Perlin silk and metallic ribbon wave sculpture (`FBM2DDeriv` + `Perlin3D`) floating over midnight-obsidian space.
36. **`GaborWaveInterference`** (`NewGaborWaveInterference(vortexPoles int, gaborFreq, bandwidth float64)`): Double-ringed 24k gold Celestial Astrolabe medallion enclosing a multi-harmonic `GaborNoise2D` + `FBM2DDeriv` Chladni caustic mandala and `Curl2DAnalytical` filaments.

### 5.1h Wave 20 — M.C. Escher Trio: Circle Limit Hyperbolic Tessellation, Print Gallery Droste Spiral & Isohedral Metamorphosis (+3 Engines + 4 Escher Mathematical Kernels, 132 Total — Implemented)
37. **`EscherCircleLimit`** (`NewEscherCircleLimit(p, q, r int)`): *Circle Limit III / IV* hyperbolic Poincaré disk interlocking tessellation powered by `common.NewHyperbolicCoxeterGroup(p, q, r)` and `FoldToFundamentalDomain`.
38. **`EscherDrosteGallery`** (`NewEscherDrosteGallery(branches, archesPerRing int, scaleRatio float64)`): Branch-cut-free Lenstra-de Smit *Print Gallery* conformal log-polar spiral (`common.DrosteSpiralPullback`) with dual Nautilus-Cloisonné and Glide-Reflection Wave-Bird modes.
39. **`EscherMetamorphosis`** (`NewEscherMetamorphosis(cols, rows int, morphStrength float64)`): Zero-gap isohedral tessellation (`render.BuildIsohedralMorphTile`) morphing 3D Necker cubes into interlocking Escher winged tiles.

### 5.1i Wave 21 — Jean "Mœbius" Giraud *Ligne Claire* & Gouache Quartet (+4 Engines + `render/moebius_engine.go`, 136 Total — Implemented)
40. **`MoebiusCanyon`** (`NewMoebiusCanyon(spires, strataLevels int, monolithTilt float64)`): *Arzach* (1975) asymmetric petrified rock citadel, flying-buttress stone bridge, domain-warped Voronoi boulder fissures (`EvaluateMoebiusStoneRelief`), lemon-gold to vermilion-red to indigo-black gouache gradient, and soaring white *Ptéroïde* glider.
41. **`MoebiusIncal`** (`NewMoebiusIncal(facets, ringLayers int, crystalScale float64)`): *Panzer Dragoon* & *L'Incal* asymmetric vista framed by a carved Mayan/Mœbius stone column & coffered soffit, sunlit stepped desert pyramid, spirally-fluted Nautilus airships, and dual-pyramid Incal crystal.
42. **`MoebiusEdena`** (`NewMoebiusEdena(domes, terraces int, floraScale float64)`): *Le Monde d'Edena* & *40 Days in the Desert B* gnarled sandstone canyon arch, wind-rippled desert dunes, winding turquoise oasis lagoon, and biomorphic sanctuary domes.
43. **`MoebiusStarwatcher`** (`NewMoebiusStarwatcher(rings, sails int, orbitTilt float64)`): *Starwatcher* & *Chroniques Métalliques* weathered bone-ivory and brass-ochre relic ring-citadel over jagged desert badlands and a billowing coral-gold cosmic cloud bank.

### 5.1j Wave 22 — Part VIII Flagship Quartet: Piranesi Carceri Vault Labyrinth, 3D SDF Mœbius Leviathan Airship, Dürer/Jamnitzer 3D Polyhedron & Kubelka-Munk Mineral Shan-Shui (+4 Engines + Part VIII Foundations, 140 Total — Implemented)
44. **`PiranesiCarceri`** (`NewPiranesiCarceri(arches, tiers int, vaultDepth float64)`): *Carceri d'Invenzione* 4-tier subterranean vaulted stone labyrinth with suspension catwalks, volumetric sunbeams, Kubelka-Munk sepia/sienna/indigo pigments, and Hessian principal-curvature burin cross-hatching (`RenderCurvatureEngraving`).
45. **`MoebiusLeviathan`** (`NewMoebiusLeviathan(ribs, crystalCount int, bankAngle float64)`): 3D SDF raymarched (`RaymarchSDF3D`) biomechanical Nautilus-whale airship with fluted bone-ivory carapace ribs, striped vermilion-gold dorsal sails, and escort gliders over Kubelka-Munk watercolor cloud banks.
46. **`DurerPolyhedron`** (`NewDurerPolyhedron(symmetry, rings int, twist float64)`): 3D SDF raymarched (`FoldIcosahedral` / `FoldOctahedral`) hollow Renaissance star-polyhedron & armillary monument engraved with principal-curvature copperplate burin cross-hatching and chiaroscuro white-chalk highlights.
47. **`MineralShanShui`** (`NewMineralShanShui(peaks, mistLayers int, mineralRichness float64)`): Song-Dynasty *Qinglü Shanshui* karst landscape glazed with physical Kubelka-Munk Azurite (*Shiqing*), Malachite (*Shilv*), Yellow Ochre (*Zheshi*), and 24k Mud-Gold (*Nijin*) with cold-pressed mineral granulation (`EvaluateWatercolorWash`) and axe-cut *Cunfa* rock fissures.

### 5.2 Direction 2: Pure-Go WebAssembly (WASM) Interactive Gallery Playground
- **Architecture (`wasm/main.go`, `GOOS=js GOARCH=wasm`)**:
  - Compile the pure-Go `generativeart` registry and HDR engine to WebAssembly.
  - Provide an interactive **Live Studio / Playground** inside the web gallery where users can select any generator, tweak parameters via live sliders, switch palettes/seeds, render directly in the browser, and export lossless `.png` or vector `.svg`.

### 5.3 Direction 3: Museum Exhibition Poster & Multi-Panel Contact Sheet Engine (`poster.go`, `cmd/genart poster` — Implemented)
- **Motivation**: Generative artists routinely present works either as single matted museum prints or as $R \times C$ (`2×2`, `3×3`) **Small Multiples / Contact Sheets** comparing seed variations, parameter sweeps, or curated multi-engine series, complete with archival mat framing, palette swatch bars, and engraved technical typography.
- **Architecture (`poster.go` & `cmd/genart`)**:
  - `PosterConfig` & `PosterPanel`: Supports `1×1` single-work museum framing, `R×C` multi-seed variation grids (`RenderPoster`), and `R×C` curated multi-artwork exhibition sheets (`RenderGalleryPoster`).
  - **Archival Mat & Double-Fillet Bevel**: Supports both **Museum Ivory Parchment (`light`)** and **Obsidian Gallery Gold (`dark`)** themes with 2D SDF drop shadows, recessed mat bevels, and double-line gilded fillet frames.
  - **Zero-Dependency Vector Engraver Typography (`drawPosterText`)**: Built-in geometric vector stroke font supporting full ASCII (`A-Z`, `a-z`, `0-9`, technical symbols) rendered via sub-pixel anti-aliased vector paths (compatible with both HDR PNG and infinite-resolution SVG/PDF export).
  - **Palette Swatch Bar & Seed Badges**: Automatically renders beveled circular color swatches, per-panel `#01 · SEED` captions, title/subtitle headers, and constructor recipe signatures.
  - **CLI Subcommand (`genart poster`)**:
    ```bash
    # Render a 2x2 multi-seed variation contact sheet on Museum Ivory Parchment
    go run ./cmd/genart poster -art dunhuangapsara -grid 2x2 -panel 420 -theme light -out images/poster_dunhuang_2x2.png

    # Render a 2x2 curated multi-artwork exhibition poster in Obsidian Dark & Gold
    go run ./cmd/genart poster -arts dunhuangapsara,celticknotwork,gyroidsurface,persiantilework -grid 2x2 -panel 420 -theme dark -title "WAVE XII EXHIBITION" -out images/poster_wave12.png
    ```

### 5.4 Direction 4: High-Definition 600×600 Seamless Loop GIF Expansion (29 HD Loop GIFs — Implemented)
- Expanded the `600×600` 18-bit LUT + K-Means quantized seamless 360° loop GIF collection to **29 HD animations**: `durerpolyhedron_loop.gif`, `piranesicarceri_loop.gif`, `moebiusleviathan_loop.gif`, `moebiuscanyon_loop.gif`, `moebiusincal_loop.gif`, `eschercirclelimit_loop.gif`, `escherdrostegallery_loop.gif`, `escherdrostewave_loop.gif`, `gaborwaveinterference_loop.gif`, `analyticalcurlsilk_loop.gif`, `nebulaplumes_loop.gif`, `abyssalcurrents_loop.gif`, `solarprominence_loop.gif`, `botanicalwhorls_loop.gif`, `vorteximpasto_loop.gif`, `squeegeescrape_loop.gif`, `basaltcolumns_loop.gif`, `wireframeterrain_loop.gif`, `lissajousribbon_loop.gif`, `marchingisolines_loop.gif`, `vortexfluid_loop.gif`, `dunhuangapsara_loop.gif`, `stainedrosewindow_loop.gif`, `newtonbasin_loop.gif`, `origamitessellation_loop.gif`, `cyclicautomata_loop.gif`, `gyroidsurface_loop.gif`, `mosaicsun_loop.gif`, and `magneticdipole_loop.gif`.

---

## Part VI. Next-Generation Architecture & Platform Roadmap (6 Directions)

### 6.1 Direction 1: Lock-Free Parallel Multi-Frame Animation Engine (`platform.go`, `generativeart.go`)
- **Problem**: `DrawTimelapseFrames` currently renders frames sequentially on a single goroutine, and `Canva.SetSeed` touches global `rand.Seed`, creating lock contention during concurrent panel/frame rendering.
- **Architecture**:
  - Add `c.CloneForWorker(seed int64) *Canva` to allocate isolated per-worker `Canva` instances sharing read-only options while keeping separate HDR/RGBA buffers and local `*rand.Rand` / `*common.NoiseEngine`.
  - Parallelize `DrawTimelapseFrames` across `runtime.NumCPU()` workers (`5×–8×` speedup for 20-frame `600×600` HD GIFs).

### 6.2 Direction 2: Full Recipe Introspection & Dynamic CLI Parameter Injection (`arts/registry.go`, `cmd/genart`)
- **Problem**: `RecipeMetadata` currently omits `ColorSchema` and `Options` state (`LineWidth`, `NIters`, `Alpha`), causing `ReplayRecipe` to fall back to `DefaultPal`; `cmd/genart render` cannot override engine-specific parameters from the command line.
- **Architecture**:
  - Extend `RecipeMetadata` to persist exact `ColorSchema` RGBA stops, `PaletteName`, and `PostEffects`.
  - Add `ParamSpec` metadata (`Name`, `Type`, `Default`, `Min`, `Max`, `Description`) to `ArtDescriptor`.
  - Support `genart info <art>` and `genart render -art <name> -set key1=val1,key2=val2`.

### 6.3 Direction 3: 2D Parameter-Sweep Phase Diagrams & Animated Exhibition Posters (`poster.go`)
- **Architecture**:
  - **2D Parameter-Space Contact Sheets (`RenderParameterSweepPoster`)**: Sweep parameter $X$ across columns and parameter $Y$ across rows (or sweep across all 12 curated palettes), automatically engraving exact coordinate captions (`SWIRL=1.25 · ARMS=4`) beneath each panel.
  - **Animated Exhibition Poster GIFs (`RenderAnimatedPoster`)**: Synchronize `AnimatedEngine` frames across all $R \times C$ matted panels and export a museum-framed looping exhibition GIF.

### 6.4 Direction 4: Composable HDR Post-Processing & Optical Shader Pipeline (`render/postprocess.go`, `generativeart.go`, `cmd/genart` — Implemented)
- **Motivation**: Decouple optical, material, and relief post-processing from individual `arts/` generators into a declarative, chainable 64-bit linear HDR post-processing pipeline usable both programmatically (`c.AddPostEffect(...)`, `c.WithPostEffects(...)`) and from the CLI (`cmd/genart render` & `replay`).
- **Architecture & Operators**:
  1. **Unified `render.PostEffect` Interface & Pipeline**:
     - `PostEffect` (`Apply(rc *Canvas)`), `PostEffectFunc`, and automatic execution on `Canva.Draw` / `Canva.DrawTimelapseFrames` when post-effects are attached to `Canva`.
  2. **Radial Chromatic Aberration & Lens Dispersion (`ApplyChromaticAberration(strength float64)`, `ChromaticAberrationEffect`)**:
     - Sub-pixel bilinear radial dispersion of R/G/B channels in 64-bit linear HDR space, adding authentic optical prism fringing to cosmic nebulae, attractors, and interference fields.
  3. **Luminance 2.5D Relief & Micro-SSAO Shading (`ApplyLuminanceRelief(strength, azimuthDeg, elevationDeg, specStrength float64)`, `LuminanceReliefEffect`)**:
     - Treats perceptual Oklab lightness $L(x,y)$ as a continuous 2.5D heightfield, computes $3\times 3$ Scharr surface normals, 8-tap micro-cavity ambient occlusion, and Blinn-Phong grazing illumination + specular highlights.
  4. **Structure-Tensor Anisotropic Kuwahara Filter (`ApplyAnisotropicKuwahara(radius int, sharpness, anisotropy float64)`, `AnisotropicKuwaharaEffect`)**:
     - Evaluates smoothed $2\times 2$ structure tensors $J = G_\sigma * (\nabla I \nabla I^\top)$ to align elliptical Kuwahara sectors along the dominant local flow tangent $\theta(x,y)$, turning noisy pixel fields into crisp, flow-coherent gouache/oil planes without blurring edges.
  5. **Oklab Perceptual Color Grading & Split-Toning (`ApplyColorGrade(contrast, saturation, shadowHueDeg, highlightHueDeg, splitStrength float64)`, `ColorGradeEffect`)**:
     - Performs linear-light filmic S-curve contrast, Oklab chroma scaling, and shadow/highlight split-toning without hue distortion.
  6. **Archival Deckle-Edged Paper Border (`ApplyDeckleBorder(marginFrac, roughness float64, matCol color.RGBA, seed int64)`, `DeckleBorderEffect`)**:
     - Multi-octave FBM torn cotton-rag deckle edge with soft recessed drop shadow, turning any full-bleed canvas into a museum float-mounted sheet.
  7. **Tone-Mapping & Classic Passes as Composable `PostEffect`s**:
     - `BloomEffect`, `PaperTextureEffect`, `VignetteEffect`, and `ToneMapEffect(exposure, gamma)` unified under the same interface and exposed in `cmd/genart` via `-bloom`, `-paper`, `-vignette`, `-chromatic`, `-relief`, `-kuwahara`, `-grade`, `-deckle`, and `-tonemap`.

### 6.5 Direction 5: Pen-Plotter Vector Clipping & TSP Path Optimization (`render/vector_bridge.go`, `render/pipeline.go`, `cmd/genart` — Implemented)
- **Architecture**:
  - **Vector Clip Stack Propagation (`clipPolyStack`)**: Propagates `PushClipPath`, `PushClipPolygon`, `PushClipCircle`, and `PushClipSDF` geometry into vector recording (`clipFilledPathAgainstPolyStack`, `clipOpenPathAgainstPolyStack`, `ClipPolylineToPolygon`) so exported SVG/PDF vector paths respect active stencil masks 1:1 with the HDR raster canvas.
  - **Centerline Recording & Multi-Pen Layer Grouping**: Records stroke centerlines on `StrokePath`, `StrokePolyline`, and `DrawTaperedSpline`, groups paths by color (or quantizes colors into `MaxPens` Oklab K-Means palette slots via `-pens N`), and emits `<g id="pen-layer-..." inkscape:groupmode="layer">` SVG layers ordered from lightest to darkest pen.
  - **Bidirectional Nearest-Neighbor + 2-Opt TSP Tour Optimizer (`OptimizeVectorOpsTSP`, `WritePlotterSVG`, `ToPlotterSVG`)**: Reorders strokes and flips reversible path directions per layer to minimize non-drawing pen-up travel distance (`60%–85%` typical reduction), reporting `PlotterStats` (`LayerCount`, `TotalPaths`, `PenDownDistancePx`, `PenUpBeforePx`, `PenUpAfterPx`, `TravelReductionPct`) via `genart render -plotter -out file.svg`.

### 6.6 Direction 6: Pure-Go WebAssembly (WASM) Interactive Studio (`wasm/main.go`, `index.html`)
- **Architecture**:
  - Compile the engine registry to WebAssembly (`GOOS=js GOARCH=wasm`) and add a Recipe Inspector / Live Parameter & Post-Effect Studio inside `index.html`.

---

## Part VII. Art-Empowering Rendering & Mathematical Foundations (For Future `arts/` — Implemented)

To elevate future generative artworks in `arts/` (especially mathematical topology, fluid dynamics, geomorphology, and tactile physical media), the following 5 foundational subsystems are implemented in `render/` and `common/`:

### 7.1 Fast Line Integral Convolution (LIC / OFLIC) & Anisotropic Texture Advection (`render/field_engine.go` — Implemented)
- **Capability**: Given any `VectorField2D` $\mathbf{v}(x,y)$ and a multi-scale pigment/noise seed buffer (`ComputeLICField`, `RenderLICField`), integrate forward and backward along RK4 streamlines at every pixel with a Hann/Gaussian convolution kernel (plus phase-shifted periodic waves for seamless OFLIC animations):
  $$I_{\text{LIC}}(\mathbf{x}_0) = \int_{-L}^{L} k(s)\, T\bigl(\mathbf{\Phi}_s(\mathbf{x}_0)\bigr)\, ds$$
- **Artistic Unlock**: Produces continuous, micro-fibrous silk, magnetic iron filings, planetary atmospheric bands, and phase-coherent fluid currents without discrete stroke gaps or overlapping alpha artifacts.

### 7.2 Thin-Film Optical Iridescence & Spectral Interference (`render/gbuffer_engine.go` — Implemented)
- **Capability**: Analytical Airy thin-film interference shader `SampleThinFilmLinearRGB` & `SampleThinFilmIridescence(cosTheta, filmThicknessNm, iorFilm float64) color.RGBA` computing optical path difference $\Delta = 2 n_{\text{film}} d \cos\theta_t + \frac{\lambda}{2}$ across 3-tap R/G/B spectral sub-bands (`630–670nm`, `520–550nm`, `450–480nm`) in linear HDR light with natural high-order coherence decay.
- **Artistic Unlock**: Enables physical structural color — iridescent beetle carapaces, bismuth hopper crystals, raku ceramic glazes, mother-of-pearl nacre chambers (`NautilusSpiral`), soap-film minimal surfaces (`GyroidSurface`), and oil-slick vortices.

### 7.3 Heightfield Horizon-Mapped Soft Shadows & Subsurface Scattering (`render/gbuffer_engine.go` — Implemented)
- **Capability**: Upgraded `FiberGBuffer` with multi-ray penumbra screen-space horizon raymarching (`ComputeHorizonShadows(lightDir, maxRadius, softness)`) and `ResolveToCanvasAdvanced(rc, cfg)` combining 12-tap SSAO, horizon soft shadows, translucent subsurface scattering (SSS), and thin-film iridescence.
- **Artistic Unlock**: Casts true long-range self-shadows across alpine ridges, sand dunes, and overlapping impasto bristle bundles while giving jade, wax, marble, and botanical petals translucent internal glow.

### 7.4 Weighted Blue-Noise Poisson-Disk & Centroidal Voronoi Stippling (`render/field_engine.go` — Implemented)
- **Capability**:
  - `SamplePoissonDisk2D(w, h, minRadius float64, densityFn ScalarField2D, rng *rand.Rand) []Point`: Bridson's $O(N)$ spatial-grid blue-noise point generator with adaptive density-driven radius.
  - `RelaxCentroidalVoronoi(pts []Point, w, h float64, gridRes int, densityFn ScalarField2D, iterations int) []Point`: Spatial-bucket accelerated weighted Lloyd's CVT relaxation.
- **Artistic Unlock**: Eliminates random clumping and rigid grid artifacts in bristle root seeding, botanical phyllotaxis, cellular mosaics, stipple engravings, and Delaunay/Voronoi tessellations.

### 7.5 Conformal, Hyperbolic & Modular Domain-Coloring Kernel (`common/conformal.go` — Implemented)
- **Capability**:
  - First-class `Mobius2D` ($\frac{az+b}{cz+d}$, composition, inversion, determinant normalization, isometric circles, Poincaré disk automorphisms, Cayley transforms), `Circle2D` point/circle inversions, `PoincareGeodesicArc`, and `EvaluateDomainColoring(fz, logPeriod, phaseRays)` combining normalized argument phase with anti-aliased log-modulus isolines and stereographic Riemann-sphere shading.
- **Artistic Unlock**: Dramatically simplifies and enriches conformal fractals, Kleinian limit sets, modular forms, elliptic functions, and complex dynamical systems.

---

## Part VIII. Physical Subtractive Pigments, 3D SDF Raymarching & Principal-Curvature Copperplate Engraving (`render/` — Implemented)

### 8.1 Two-Flux Kubelka-Munk Subtractive Pigment Mixing & Watercolor Granulation (`render/pigment_engine.go` — Implemented)
- **Capability**:
  - Models spectral absorption $K(\lambda)$ and scattering $S(\lambda)$ across R/G/B bands with無限光学厚度反射率 $R_\infty = 1 + \frac{K}{S} - \sqrt{\left(\frac{K}{S}\right)^2 + 2\frac{K}{S}}$ and Saunderson surface-reflection correction (`MixKubelkaMunk`, `LerpKubelkaMunk`, `GlazeKubelkaMunk`).
  - Curated 15 historical mineral & traditional pigments (`PigmentAzurite`, `PigmentMalachite`, `PigmentCinnabar`, `PigmentYellowOchre`, `PigmentBurntSienna`, `PigmentRawUmber`, `PigmentUltramarine`, `PigmentIndigo`, `PigmentGoldLeaf`, `PigmentPineSootInk`, `PigmentTitaniumWhite`, etc.) plus `EvaluateWatercolorWash` for capillary coffee-ring edge darkening and cold-pressed paper valley granulation.

### 8.2 3D Signed Distance Field (SDF) Raymarching & Polyhedral Symmetry Folding (`render/sdf3d.go` — Implemented)
- **Capability**:
  - Analytical 3D primitives (`SDF3DSphere`, `SDF3DBox`, `SDF3DTorus`, `SDF3DCylinderY`, `SDF3DOctahedron`, `SDF3DFlutedColumnY`, `SDF3DRomanArchZ`, `SDF3DHexPrismY`, `SDF3DPlatonicCage`, `SDF3DClebschCalabiYau`), smooth CSG booleans (`SDF3DSmoothUnion`, `SDF3DSmoothSubtract`, `SDF3DSmoothIntersect`), and symmetry folding (`FoldPolarModXZ`, `FoldOctahedral`, `FoldIcosahedral`, `AlignIcosahedral5FoldY`).
  - Tetrahedral finite-difference normal & 3D mean curvature solver (`CalcSDF3DNormalAndCurvature`), penumbra soft shadows (`CalcSDF3DSoftShadow`), 5-tap exponential 3D AO (`CalcSDF3DAO`), `RaymarchSDF3D`, and 3D object-locked stone relief (`EvaluateMoebiusStoneRelief3D`).

### 8.3 Hessian Principal-Curvature & 3D Object-Locked Copperplate Burin Engraving (`render/engraving_engine.go` — Implemented)
- **Capability**:
  - 2nd-order Hessian principal curvature eigenvector solver (`ComputePrincipalCurvature2D`), variable-width Jobard-Lefer intaglio burin cross-hatching (`RenderCurvatureEngraving`), and 3D object-locked triplanar/cylindrical facet burin hatching (`Evaluate3DBurinEngraving`) that rotates rigidly with 3D geometry without 2D screen-space swimming.

---

## Part IX. Next-Horizon Roadmap: 4 Strategic Directions (Waves 23–26)

### 9.1 Direction 1: Master Architecture, Celestial Mechanics & Sci-Fi Megastructures (Wave 23 — Implementing)
Leveraging Part VIII's 3D SDF raymarcher, 3D object-locked copperplate burin engraving, Kubelka-Munk mineral pigments, and Mœbius *Ligne Claire* gouache pipeline:
1. **`BabelLibrary`** (`NewBabelLibrary(galleries, tiers int, shaftScale float64)`):
   - Inspired by Jorge Luis Borges's *The Library of Babel*, Piranesi, and Escher: 3D SDF raymarched infinite hexagonal book-alcove galleries, spiral staircases, suspended brass lanterns, and receding hexagonal ventilation shafts with Kubelka-Munk ochre/sienna/umber glazes and 3D object-locked burin cross-hatching (`images/babellibrary.png` + `images/babellibrary_loop.gif`).
2. **`KeplerHarmonices`** (`NewKeplerHarmonices(shells, rings int, inclination float64)`):
   - Homage to Johannes Kepler's 1596 *Mysterium Cosmographicum* & 1619 *Harmonices Mundi*: nested 3D SDF hollow Platonic polyhedron cages (Dodecahedron, Icosahedron, Octahedron, Cube, Star-Tetrahedron) interleaved with zodiacal armillary spheres around a central solar core, counter-rotating seamlessly along their 5-fold and 4-fold symmetry axes (`images/keplerharmonices.png` + `images/keplerharmonices_loop.gif`).
3. **`MoebiusHermeticGarage`** (`NewMoebiusHermeticGarage(ribs, spires int, ringTilt float64)`):
   - Homage to Jean "Mœbius" Giraud's *Le Garage Hermétique* (1976): inside a colossal weathered asteroid cavern, a bone-ivory and vermilion-striped toroidal mothership dock with ribbed mechanical struts and brass conduits hovers above an emerald-turquoise subterranean drydock, flanked by crystalline mineral spires, an observation catwalk, and orbiting shuttle pods (`images/moebiushermeticgarage.png` + `images/moebiushermeticgarage_loop.gif`).
4. **`CalabiYauManifold`** (`NewCalabiYauManifold(foldOrder, rings int, phaseMorph float64)`):
   - String-theory Calabi-Yau / Clebsch algebraic surface manifold raymarched in 3D with hollowed genus-tunnel handles and orbiting golden geodesic rings, combining Kubelka-Munk Lapis Lazuli, Malachite, Vermilion, and 24k Gold Leaf glazes with 3D object-locked principal-curvature burin cross-hatching (`images/calabiyaumanifold.png` + `images/calabiyaumanifold_loop.gif`).

### 9.2 Direction 2: Part IX Optical Upgrade — 3D Dielectrics, Snell Refraction, Beer-Lambert Absorption & Volumetric Caustics (`render/sdf3d_optics.go`)
- **Underlying Engine**:
  - Extend `sdf3d.go` with multi-bounce Snell refraction ($\eta_1 \sin\theta_1 = \eta_2 \sin\theta_2$), Schlick-Fresnel reflectance, internal total reflection, Beer-Lambert spectral volume attenuation $T(\lambda) = e^{-\sigma_a(\lambda)\, d}$, and spectral index-of-refraction dispersion ($\eta_R, \eta_G, \eta_B$).
- **Target Masterpieces**:
  - **`MuranoGlassNautilus`**: Venetian Murano blown-glass nautilus with gold-leaf *Aventurine* inclusions and internal caustic light pipes.
  - **`DispersionPrism`**: Newton optical prism array splitting collimated beams into dispersed spectral caustics.
  - **`AmberInclusion`**: Baltic amber cabochon with Beer-Lambert honey-gold internal absorption enclosing fractal dendritic inclusions.

### 9.3 Direction 3: Part X Biological Upgrade — 3D Morphogenesis, Biomineralization & Ernst Haeckel *Kunstformen der Natur* (`render/morphogenesis3d.go`)
- **Underlying Engine**:
  - 3D spherical-harmonic and polyhedral-lattice reaction-diffusion / Voronoi skeletal operators generating perforated siliceous endoskeletons and calcareous plates in 3D SDF space, paired with 19th-century chromolithograph & copperplate engraving shaders.
- **Target Masterpieces**:
  - **`HaeckelRadiolaria`**: Ernst Haeckel (1904) *Phaeodaria / Acantharia* concentric perforated geodesic silica shells and radial spicules.
  - **`DiatomFrustule`**: Multi-scale hierarchical sieve-pore (*areolae*) microscopic glass frustules.
  - **`BrainCoral3D`**: True 3D surface-extruded gyroid/meandroid brain-coral labyrinth with abyssal bioluminescence.

### 9.4 Direction 4: Interactive Gallery Enhancements (`index.html`)
- Inline **Static PNG / Animated GIF toggle badge** directly on dual-mode gallery cards.
- **Museum Frame Fullscreen Mode** and interactive **Algorithm Lineage Graph** inside the exhibition lightbox.

---

## Part X. Underlying Infrastructure Audit & 3-Phase Core Engine Upgrade Roadmap

A comprehensive technical audit across `generativeart.go`, `platform.go`, `poster.go`, `cmd/genart`, `arts/registry.go`, `render/`, and `common/` identified three major foundational upgrade tracks:

### 10.1 Phase 1: Architecture, Concurrency, Full Recipe Introspection & CLI Toolchain Upgrade (Implemented)

1. **Lock-Free Parallel Multi-Frame Animation Engine & Temporal Motion Blur (`generativeart.go`, `platform.go`)**:
   - **Problem Solved**: `DrawTimelapseFrames` previously rendered frames sequentially on a single goroutine, and `Canva.SetSeed` mutated global `rand.Seed(seed)` causing lock contention during concurrent panel/frame rendering.
   - **Upgrades**:
     - `(*Canva).CloneForWorker(seed int64) *Canva`: Allocates an isolated worker `Canva` sharing read-only options, palette stops, and post-effect pipelines while maintaining independent 64-bit HDR/RGBA buffers and lock-free local `*rand.Rand`, `*common.NoiseEngine`, and `*common.PerlinNoise` instances without touching global `math/rand`.
     - **Multi-Core Parallel `DrawTimelapseFrames`**: Automatically dispatches `AnimatedEngine` frame generation across `runtime.NumCPU()` worker goroutines (`5×–8×` speedup on multi-frame `600×600` HD GIFs).
     - **Shutter-Angle Temporal Supersampling (`DrawTimelapseFramesTemporalAA(e, totalFrames, subSamples, shutterFraction)`)**: Integrates `subSamples` sub-frames across $[\phi_i, \phi_i + \Delta\phi_{\text{shutter}}]$ directly in 64-bit linear HDR space before tone-mapping/quantization, producing physically accurate cinematic motion blur.
     - **Configurable GIF Palette Capacity (`WriteGIFWithColors(w, delayCs, maxColors)`)**: Allows callers to choose between compact 32-color (5-bit LZW root) and high-gamut 64/128/256-color shared K-Means palettes, while sampling up to 8 frames across the animation cycle.

2. **100% Lossless PNG Recipe Introspection & PostEffect Serialization (`platform.go`, `arts/registry.go`, `render/postprocess.go`)**:
   - **Problem Solved**: `RecipeMetadata` previously omitted `ColorSchema` (`[][4]uint8`), `PaletteName`, `Foreground`, `LineColor`, `LineWidth`, `NIters`, `Alpha`, and structured post-effect parameters, causing `arts.ReplayRecipe` to fall back to `desc.DefaultPal` and drop post-processing passes.
   - **Upgrades**:
     - Extended `RecipeMetadata` with `Palette [][4]uint8`, `Foreground [4]uint8`, `LineColor [4]uint8`, `LineWidth float64`, `NIters int`, and `Alpha int`.
     - Added `render.ParsePostEffect(spec string) (render.PostEffect, bool)` to parse serialized post-effect descriptors (`bloom(0.62,6.00,0.35)`, `relief(0.60,-135,38,0.38)`, `chromatic(0.80)`, `kuwahara(3,8.0,1.25)`, `grade(...)`, `tonemap(...)`, `paper(...)`, `vignette(...)`, `deckle(...)`) back into live executable HDR `PostEffect` passes during `arts.ReplayRecipe` / `genart replay`.

3. **Dynamic CLI Parameter Introspection (`genart info`) & Injection (`-set key=val`) (`arts/registry.go`, `cmd/genart/main.go`)**:
   - **Problem Solved**: `cmd/genart` previously hardcoded `desc.Build(nil)` with no way to inspect or override generator parameters from the command line.
   - **Upgrades**:
     - Added `(*ArtDescriptor).DefaultParams() map[string]interface{}` and `(*ArtDescriptor).ParamSpecs() []ParamSpec` using reflection on `desc.Build(nil)` to expose parameter names, Go types (`int`, `float64`, `string`, `bool`), and default values.
     - Added `arts.ParseParamOverrides(spec string) (map[string]interface{}, error)` supporting case-insensitive parameter matching and `key1=val1,key2=val2` syntax.
     - Added `genart info <art>` subcommand and `-set key1=val1,key2=val2` flag across `genart render`, `genart replay`, and `genart poster`, plus `-frames` / `-delay` support on `genart replay` so any static PNG recipe can be replayed directly into a looping `.gif`.

4. **2D Parameter-Sweep Phase Diagrams, Animated Exhibition Posters & Vector Sub-Panel Preservation (`poster.go`, `cmd/genart/main.go`)**:
   - **Problem Solved**: `poster.go` could not sweep 2D parameter spaces, could not render animated multi-panel GIF posters, lacked mathematical/Greek engraver glyphs (`=`, `[`, `]`, `%`, `*`, `π`, `θ`, `φ`, `λ`, `σ`, `α`, `β`, `ω`), and discarded sub-panel vector paths when blitting onto the poster canvas.
   - **Upgrades**:
     - **`ParameterSweepSpec` & `RenderParameterSweepPoster`**: Sweeps parameter $X \in [x_{\min}, x_{\max}]$ across columns and parameter $Y \in [y_{\min}, y_{\max}]$ across rows with a fixed seed, automatically engraving exact coordinate captions (`X_NAME=VAL · Y_NAME=VAL`) beneath each panel (`genart poster -art <name> -sweep-x param:min:max -sweep-y param:min:max`).
     - **`RenderAnimatedPoster` & `RenderAnimatedGalleryPoster`**: Synchronizes `AnimatedEngine` frames across all $R \times C$ matted panels in parallel and encodes a museum-framed looping exhibition GIF (`genart poster ... -frames 20 -delay 12 -out poster.gif`).
     - **Sub-Panel Vector Path Preservation (`CompositeSubCanvasVectors`)**: Translates and clips sub-panel `vectorOp` paths directly into the poster's `render.Canvas` so exporting a poster to `.svg` or `.pdf` retains 100% native infinite-resolution vector paths alongside the museum mat frame.
     - **Expanded Mathematical & Greek Engraver Typography (`appendEngraverGlyph`)**: Added vector stroke glyphs for `=`, `[`, `]`, `{`, `}`, `%`, `*`, `;`, `!`, `?`, `<`, `>`, `~`, and Greek symbols (`α`, `β`, `γ`, `Δ`, `θ`, `λ`, `μ`, `π`, `σ`, `τ`, `φ`, `ω`, `Ω`).

---

### 10.2 Phase 2: Core 3D, Optical & Rendering Pipeline Upgrades (Implemented)

1. **Edge-Adaptive 2×2 SSAA & Fast Integer Voronoi Hash in `render/moebius_engine.go`**:
   - Two-pass edge-adaptive supersampling: evaluates $1\times 1$ primary centers first, and only evaluates the remaining 3 sub-pixel rays where neighbor `ObjID`/`PartID`/`DetailID`, depth $|\Delta Z|$, normal $\|\Delta\mathbf{N}\|$, illumination $|\Delta I|$, crack/tick masks, or terminator hachures require supersampling (`50%–75%` raymarching reduction on sky/flat regions with zero loss of contour or hachure crispness).
   - Replaced 18 trigonometric `math.Sin` calls in `EvaluateMoebiusStoneRelief`'s $3\times 3$ inner Voronoi loop with 64-bit integer bit-mixing (`moebiusHash2D`), and parallelized `drawMoebiusSerigraphBorder`.
2. **Keinert Over-Relaxation Sphere Tracing, Bounding-Sphere Early-Out & Smooth Triplanar Blending (`render/sdf3d.go`)**:
   - Added `IntersectBoundingSphere` and `RaymarchSDF3DBounded` for 0-call sky/background ray rejection, plus Keinert over-relaxation ($\omega = 1.25$) with automatic step fallback and secant root interpolation in `RaymarchSDF3D`, and `CalcSDF3DNormalAndCurvatureAtHit` (eliminating redundant center SDF evaluations).
   - Upgraded `EvaluateMoebiusStoneRelief3D` with 4th-power smooth triplanar normal blending across $45^\circ$ bevel transitions while keeping a fast 1-plane path on cardinal faces.
3. **3D Dielectrics, Snell Refraction, Beer-Lambert Absorption & Spectral Dispersion (`render/sdf3d_optics.go`)**:
   - Implemented `ReflectVec3`, `RefractVec3`, `SchlickFresnel`, `BeerLambertTransmittance`, `MarchSDF3DInterior`, `TraceSDF3DDielectric`, and `EvalSDF3DCausticIntensity`, plus 6 physical material presets (`DielectricCrownGlass`, `DielectricFlintPrism`, `DielectricBalticAmber`, `DielectricMuranoEmerald`, `DielectricCobaltSapphire`, `DielectricDiamond`).
4. **Oklch Chroma-Preserving Gamut Mapping & Thin-Lens Polygonal Bokeh DoF (`render/canvas.go`, `render/postprocess.go`, `cmd/genart`)**:
   - Added `Oklch`, `OklabToOklch`, `OklchToOklab`, and `GamutMapOklab` in `render/canvas.go` so `OklabToRGBA` preserves perceptual lightness $L$ and hue angle $H$ via binary-searched chroma compression instead of per-channel RGB clipping.
   - Added `ApplyDepthOfField`, `ApplyDepthOfFieldBuffer`, and `DepthOfFieldEffect` (`dof(...)` in `ParsePostEffect`, `-dof` in `cmd/genart`) with 48-tap Golden-Angle Vogel spiral sampling shaped by an $N$-blade regular polygon aperture.

---

### 10.3 Phase 3: Computational Geometry, Mathematical PDE Fields & Vector Hidden-Line Removal (Implemented)

1. **Unified Bowyer–Watson Delaunay Triangulation & Dual Clipped Voronoi Mesh (`render/geometry2d.go`)**:
   - Implemented `DelaunayTriangle`, `VoronoiCell2D`, `DelaunayVoronoiMesh2D`, and `BuildDelaunayVoronoi2D(sites, boundsMin, boundsMax)` outputting exact Delaunay triangles, circumcenters, unique undirected Delaunay edges, and Sutherland–Hodgman half-plane clipped dual Voronoi polygons with exact Green's-theorem shoelace area and centroid (`PolygonAreaAndCentroid2D`, `ClipPolygonHalfPlane2D`).
   - Added exact geometric polygon-centroid Lloyd CVT relaxation (`RelaxVoronoiPolygonCVT`), Gabriel Graph extraction (`ExtractGabrielGraph`), and $O(E \log E)$ Kruskal Euclidean Minimum Spanning Tree (`ExtractMinimumSpanningTree`).
2. **Analytical 3D Noise Derivatives, `Curl3D` & Rich 2D/3D Cellular Voronoi (`common/noise.go`)**:
   - Added `Perlin3DDeriv(x, y, z)` (single-pass 3D $C^2$ gradient Perlin noise + exact analytical $\nabla N$), `FBM3D`, `FBM3DDeriv`, `Ridged3D`, and divergence-free `Curl3D(x, y, z, octaves)` ($\nabla \times \vec{\Psi}$).
   - Added `Voronoi2D` and `Voronoi3D` returning `F1`, `F2`, Inigo Quilez's exact bisector distance `EdgeDist`, log-sum-exp `SmoothF1`, analytical gradient `(GradX, GradY)`, and `CellID`.
3. **Spherical Harmonics $Y_l^m(\theta,\phi)$, Geodesic Polyhedra & Anisotropic PDE Solvers (`common/spherical_harmonics.go`, `render/field_engine.go`)**:
   - Added `NormalizedAssociatedLegendre`, real orthonormal `SphericalHarmonicYlm(l, m, theta, phi)`, `SphericalHarmonicCartesian`, and `BuildGeodesicSphere(subdivisions)` generating both subdivided icosahedral geodesic triangles and dual **Goldberg Polyhedron** hexagonal/pentagonal cells on $S^2$.
   - Upgraded `ExtractMarchingSquares` segment stitching to an $O(M)$ spatial endpoint hash map, and added `SolveAnisotropicReactionDiffusion2D` (flow-tensor Gray–Scott / Turing solver) and `SolveScreenedPoisson2D` (Red-Black Gauss–Seidel SOR Helmholtz/Poisson solver).
4. **2D & 3D Vector Hidden-Line Removal (HLR) & Plotter Polygon Hatching (`render/vector_bridge.go`, `render/sdf3d.go`, `cmd/genart`)**:
   - Extended `PlotterConfig` with `HiddenLineRemoval` (`-hlr` in `cmd/genart`, occlusion-culling background strokes against foreground opaque polygons via `clipOpenPathOutsideRings`) and `HatchFills` (`-hatch-fills` in `cmd/genart`, converting filled polygons into Oklab lightness-driven parallel and cross-hatched pen strokes).
   - Added `ExtractSDF3DVectorContours` and `(*Canvas).RenderSDF3DVectorContours` in `render/sdf3d.go` to extract visible 3D silhouettes, occlusion creases, and 3D topographic cross-section iso-contours with raymarched hidden-line removal.

---

### 10.4 Phase 4: Repo-Wide Deterministic Seed Hygiene, `go test ./...` Isolation & 16-Bit PNG / Radiance `.hdr` Export (Implemented)

1. **100% Deterministic Seed Hygiene Across All 144 Engines (`arts/*.go`, `common/utils.go`)**:
   - Migrated all remaining legacy engines that previously called global `rand.Int63()` / `rand.Float64()` / `rand.Intn()` or hardcoded constant seeds (`rand.NewSource(2026xx)`) to `c.Seed()` and `c.Rand()` (`common.RandomRangeIntFrom`, `common.RandomRangeFloat64From`, `common.RandomGaussianFrom`). Every single one of the 144 engines now responds deterministically to `c.SetSeed(seed)` with zero global `math/rand` lock contention.
2. **`example/` Build Isolation for Zero-Error `go test ./...` (`example/*.go`, `.github/workflows/go.yml`)**:
   - Added `//go:build ignore` to all standalone `example/example_*.go` scripts and added `example/doc.go` so `go test ./...` and `go build ./...` pass cleanly across the entire repository while `go run example/example_<name>.go` continues to work unchanged.
3. **16-Bit Deep-Color PNG (`image.RGBA64`) & Radiance RGBE `.hdr` Export (`render/canvas.go`, `platform.go`, `cmd/genart`)**:
   - Added `(*Canvas).ResolveToRGBA64`, `(*Canvas).ResolveToneMappedRGBA64`, `(*Canva).Img64`, `(*Canva).WritePNG16`, and `(*Canva).ToPNG16` (`-depth 16` in `cmd/genart`), exporting 48/64-bit deep-color PNGs (65,536 levels per channel) with embedded `GenerativeArt-Recipe` `tEXt` metadata to eliminate 8-bit Mach banding.
   - Added `FloatToRGBE`, `RGBEToFloat`, `(*Canvas).WriteHDR`, `(*Canvas).SaveHDR`, `(*Canva).WriteHDR`, and `(*Canva).ToHDR` (`.hdr` output in `cmd/genart`) for un-clamped Radiance RGBE high-dynamic-range export.

---

### 10.5 Phase 5: 8-Band Spectral Kubelka-Munk, Wet-on-Wet Fluid Bleed & 3D Weingarten Curvature Tensor (Implemented)

1. **8-Wavelength Spectral Kubelka-Munk & Wet-on-Wet Capillary Fluid Diffusion (`render/pigment_engine.go`)**:
   - Upgraded `PigmentKM` with 8 visible spectral bands (`K8`, `S8` across `405nm..685nm`), `LinearRGBToSpectrum8`, `Spectrum8ToLinearRGB`, `MixSpectralKubelkaMunk`, `LerpSpectralKubelkaMunk`, and `GlazeSpectralKubelkaMunk`.
   - Added a 2D coupled water-pigment capillary advection-diffusion solver (`WetWatercolorDeposit`, `WetWatercolorConfig`, `SimulateWetWatercolorBleed`, `(*Canvas).RenderWetWatercolorBleed`) simulating wet-on-wet multi-pigment cross-bleeding, outward coffee-ring perimeter drying, and cold-pressed paper valley granulation.
2. **3D Surface Weingarten Shape Operator & Principal Curvature Directions (`render/sdf3d.go`, `render/engraving_engine.go`)**:
   - Added `SDF3DPrincipalCurvature` and `CalcSDF3DWeingartenCurvature` to diagonalize the $2\times 2$ tangent-plane Weingarten shape operator $\mathbf{S} = -(\mathbf{I}-\mathbf{n}\mathbf{n}^\top)\nabla^2 f$ directly on any `SDF3D`, extracting exact 3D principal curvature eigenvectors $(\mathbf{E}_1, \mathbf{E}_2)$, principal curvatures $(\kappa_1, \kappa_2)$, mean curvature $H$, and Gaussian curvature $K = \kappa_1\kappa_2$.
   - Added `Evaluate3DWeingartenEngraving` in `render/engraving_engine.go` for intrinsic 3D principal-curvature copperplate burin cross-hatching.

---

### 10.6 Phase 6: Interactive Local Live Studio (`genart serve`) & Gallery Lightbox CLI Builder (Planned)

1. **`genart serve` Interactive HTTP Studio (`cmd/genart/main.go`, `index.html`)**:
   - Add `genart serve [-port 8080]` serving `index.html` alongside `/api/engines`, `/api/render`, `/api/poster`, and `/api/inspect` endpoints for live browser parameter tweaking, post-effect chaining, and drag-and-drop PNG recipe inspection.
2. **Inline PNG/GIF Toggle Badge & Visual CLI Command Builder (`index.html`)**:
   - Add one-click static PNG / animated GIF toggle badges on dual-mode gallery cards and an interactive parameter slider + CLI command generator inside the exhibition lightbox.





