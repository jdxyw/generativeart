# Table of Contents

- [Table of Contents](#table-of-contents)
	- [Color Circle 2](#color-circle-2)
		- [parameters](#parameters)
	- [Dot Line](#dot-line)
		- [parameters](#parameters-1)
	- [Random Shape](#random-shape)
		- [parameters](#parameters-2)
	- [Janus](#janus)
		- [TODO](#todo)
		- [parameters](#parameters-3)
	- [Contour Line](#contour-line)
		- [parameters](#parameters-4)
	- [Silk Sky](#silk-sky)
		- [parameters](#parameters-5)
	- [Julia](#julia)
		- [parameters](#parameters-6)
	- [Noise Line](#noise-line)
		- [parameters](#parameters-7)
	- [Ocean Fish](#ocean-fish)
		- [parameters](#parameters-8)
	- [Circle Loop2](#circle-loop2)
		- [parameters](#parameters-9)
	- [Pixel Hole](#pixel-hole)
		- [parameters](#parameters-10)
	- [Dots Wave](#dots-wave)
		- [parameters](#parameters-11)
	- [Circle Move](#circle-move)
		- [parameters](#parameters-12)
	- [Circle Noise](#circle-noise)
		- [parameters](#parameters-13)
	- [Yarn](#yarn)
		- [parameters](#parameters-14)
	- [Black Hole](#black-hole)
		- [parameters](#parameters-15)
	- [Color Canva](#color-canva)
		- [parameters](#parameters-16)
	- [Domain Warping](#domain-warping)
		- [parameters](#parameters-17)
	- [Perlin Perls](#perlin-perls)
		- [parameters](#parameters-18)
	- [Curl Flow](#curl-flow)
		- [parameters](#parameters-19)
	- [Sand Ring](#sand-ring)
		- [parameters](#parameters-20)
	- [Watercolor](#watercolor)
		- [parameters](#parameters-21)
	- [Guilloche](#guilloche)
		- [parameters](#parameters-22)
	- [Clifford Nebula](#clifford-nebula)
		- [parameters](#parameters-23)
	- [Physarum](#physarum)
		- [parameters](#parameters-24)
	- [Differential Growth](#differential-growth)
		- [parameters](#parameters-25)
	- [Chladni](#chladni)
		- [parameters](#parameters-26)
	- [Voronoi Crystal](#voronoi-crystal)
		- [parameters](#parameters-27)
	- [Ink Landscape](#ink-landscape)
		- [parameters](#parameters-28)
	- [Gray Scott](#gray-scott)
		- [parameters](#parameters-29)
	- [Lissajous Ribbon](#lissajous-ribbon)
		- [parameters](#parameters-30)
	- [Apollonian Gasket](#apollonian-gasket)
		- [parameters](#parameters-31)
	- [Space Colonization](#space-colonization)
		- [parameters](#parameters-32)
	- [Truchet Flow](#truchet-flow)
		- [parameters](#parameters-33)
	- [Aurora Wave](#aurora-wave)
		- [parameters](#parameters-34)
	- [Fractal Flame](#fractal-flame)
		- [parameters](#parameters-35)
	- [Joy Division](#joy-division)
		- [parameters](#parameters-36)
	- [Boids Flock](#boids-flock)
		- [parameters](#parameters-37)
	- [Islamic Star](#islamic-star)
		- [parameters](#parameters-38)
	- [Delaunay Stained](#delaunay-stained)
		- [parameters](#parameters-39)
	- [Phyllotaxis](#phyllotaxis)
		- [parameters](#parameters-40)
	- [Poincare Disk](#poincare-disk)
		- [parameters](#parameters-41)
	- [Suminagashi](#suminagashi)
		- [parameters](#parameters-42)
	- [Aizawa Attractor](#aizawa-attractor)
		- [parameters](#parameters-43)
	- [Penrose Tiling](#penrose-tiling)
		- [parameters](#parameters-44)
	- [Isometric City](#isometric-city)
		- [parameters](#parameters-45)
	- [DLA Frost](#dla-frost)
		- [parameters](#parameters-46)
	- [Impasto Flow](#impasto-flow)
		- [parameters](#parameters-47)
	- [Voxel Terrain](#voxel-terrain)
		- [parameters](#parameters-48)
	- [Packed Capsules](#packed-capsules)
		- [parameters](#parameters-49)
	- [Orphic Discs](#orphic-discs)
		- [parameters](#parameters-50)
	- [Nautilus Spiral](#nautilus-spiral)
		- [parameters](#parameters-51)
	- [Mosaic Sun](#mosaic-sun)
		- [parameters](#parameters-52)
	- [Tangent Cones](#tangent-cones)
		- [parameters](#parameters-53)

## Color Circle 2

`Color Circle2` is version 2 of `Color Circle`. It still draws the circle and point cloud.

### parameters

- circleNum: The number of the circle in this drawing.

```go
cc := arts.NewColorCircle2(30)
```

![](../images/colorcircle2.png)

## Dot Line

`dot line` would draw images with the short dot and short. The short lines would compose as a big circle.

### parameters

- n: The number of elements in this image.
- ras/canv: Control the appearance of this image.
- randColor: Use the specified color or random colors.

```go
dl := arts.NewDotLine(100, 20, 50, false)
```

![](../images/dotline.png)

## Random Shape

`random shape` would draw images with random shapes. The whole image would rotate with some degree.

### parameters

- shapeNum: It indicates how many shapes you want to draw.

```go
rs := NewRandomShape(150)
```

![](../images/randomshape.png)

## Janus

`Janus` would draw an image with multiple circles split at its center with random noise in the horizontal direction.

### TODO

### parameters

## Contour Line

`Contour Line` uses the `perlin noise` to do some flow field.

### parameters

- lineNum: It indicates how many lines.

![](../images/contourline.png)

## Silk Sky

`Silk Sky` would draw an image with multiple circles converge to one point or one circle.

### parameters

- circleNum: The number of the circles in this drawing.
- sunRadius: The radius of the sun. The sun is a point/circle where other circles meet.

```go
silkSky := NewSilkSky(circleNum int, sunRadius float64)
```

![](../images/silksky.png)

## Julia

`Julia` is to draw a `Julia Set`. [Julia Set](https://en.wikipedia.org/wiki/Julia_set) is a math concept. You can define your own formula in this package.

### parameters

- fn: The custom julia set function.
- maxz: The maximum modulus length of a complex number.
- xaixs, yaixs: The range for the X-Y coordination used to mapping the julia set number to the real pixel of the image. These should be positive. It only indicates the first quadrant range.

```go
func julia1(z complex128) complex128 {
 c := complex(-0.1, 0.651)
 z = z*z + c
 return z
}
julia := arts.NewJulia(julia1, 40, 1.5, 1.5)
```

![](../images/julia.png)

## Noise Line

`Noise Line` draws some random line and circles based on `perlin noise`.

### parameters

- n: The number of random line.

```go
nl := arts.NewNoiseLine(1000)
```

![](../images/noiseline.png)

## Ocean Fish

`Ocean Fish` draws an ocean and some fishes in the center.

### parameters

- lineNum: The number of the line used to simulate the ocean wave.
- fishNum: The number of fish.

```go
o := arts.NewOceanFish(100, 8)
```

![](../images/oceanfish.png)

## Circle Loop2

`Circle loop2` draws a circle composed by colored circles.

### parameters

- depth: Control the number of circles.

```go
cl := arts.NewCircleLoop2(7)
```

![](../images/colorloop2.png)

## Pixel Hole

`Pixel Hole` draws a hole with colored dots.

### parameters

- dotN: The number of point in each iteration.

```go
ph := arts.NewPixelHole(60)
```

![](../images/pixelhole.png)

## Dots Wave

### parameters

- dotsN: The number of dots wave in the image.

```go
d := arts.NewDotsWave(300)
```

![](../images/dotswave.png)

## Circle Move

### parameters

- circleNum: The number of the circle in the image.

```go
cm := arts.NewCircleMove(1000)
```

![](../images/circlemove.png)

## Circle Noise

### parameters

- dotsN: The number of dot.
- colorMin: The minimum color.
- colorMax: The maximum color.

![](../images/circlenoise.png)

## Yarn

### parameters

- n: The number of the curve.

```go
y := arts.NewYarn(2000)
```

![](../images/yarn.png)

## Black Hole

### parameters

- circleN: The number of the circle.
- density: Control the density of the circle.
- circleGap: Identify the gap between two circles.

```go
b := arts.NewBlackHole(200, 400, 0.01)
```

![](../images/blackhole.png)

## Color Canva

### parameters

- seg:

```go
cc := arts.NewColorCanve(5)
```

![](../images/colorcanva.png)

## Domain Warping

Warping, or domain distortion is a very common technique in computer graphics for generating procedural textures and geometry. It's often used to pinch an object, stretch it, twist it, bend it, make it thicker or apply any deformation you want.

### parameters

- scale: Control the noise generator.
- xOffset: Control the noise generator.
- yOffset: Control the noise generator.
- fn: A function to mapping the `noise` to color.

```go
d := arts.NewDomainWrap(0.01, 4, 8, cmap)
```

![](../images/domainwarp.png)

## Perlin Perls

### parameters

- circleN: The number of circle on this image.
- dotsN: The number of dots in each circle. 
- colorMin: The minimum color.
- colorMax: The maximum color.

```go
pp := arts.NewPerlinPerls(10, 200, 40, 80
```

![](../images/perlinperls.png)

## Curl Flow

`Curl Flow` traces divergence-free curl-noise streamlines across the canvas with smooth organic ribbons and variable stroke widths.

### parameters

- lineNum: The number of streamlines to trace.
- stepNum: The number of integration steps along each streamline.
- noiseScale: The spatial frequency scale of the multi-octave curl noise field.

```go
cf := arts.NewCurlFlow(110, 36, 0.0045)
```

![](../images/curlflow.png)

## Sand Ring

`Sand Ring` renders concentric noisy rings using stochastic sand-spline grain deposition and additive luminous blending.

### parameters

- ringNum: The number of concentric deformed rings.
- grainsPerStep: The density of sand grains deposited per interpolation step.
- noiseScale: The amplitude of the radial Perlin noise deformation.

```go
sr := arts.NewSandRing(22, 22, 1.3)
```

![](../images/sandring.png)

## Watercolor

`Watercolor` simulates organic translucent watercolor washes by recursively subdividing and jittering polygons across multiple pigment layers.

### parameters

- shapeNum: The number of watercolor blotch shapes on the canvas.
- layers: The number of translucent pigment layers stacked per shape.
- subdivisions: The number of recursive edge subdivisions per polygon layer.
- roughness: Controls the edge displacement variance during subdivision.

```go
wc := arts.NewWatercolor(9, 22, 5, 0.32)
```

![](../images/watercolor.png)

## Guilloche

`Guilloche` generates intricate spirographic hypotrochoid and epitrochoid rosette patterns with phase-shifted harmonic envelopes and luminous interference.

### parameters

- curveNum: The number of phase-shifted guilloche curves in the family.
- majorR: The radius of the fixed outer circle.
- minorR: The radius of the rolling circle.
- penOffset: The distance of the drawing pen from the rolling circle center.

```go
g := arts.NewGuilloche(28, 175, 65, 115)
```

![](../images/guilloche.png)

## Clifford Nebula

`Clifford Nebula` renders a strange attractor density cloud (nebula/flam3 histogram rendering) using the Clifford strange attractor equations with log-density tone mapping and velocity-dependent chromaticity.

### parameters

- iterations: The number of strange attractor trajectory steps to accumulate.
- a, b, c, d: The four Clifford attractor control coefficients.

```go
cn := arts.NewCliffordNebula(1500000, -1.7, 1.8, -1.9, -0.4)
```

![](../images/cliffordnebula.png)

## Physarum

`Physarum` simulates a bioluminescent slime mold transport network with chemotaxis-driven agents, trail deposition, and 3x3 Gaussian diffusion and decay.

### parameters

- agentNum: The number of chemotaxis agents in the simulation.
- steps: The number of simulation steps.
- sensorAngle: The angle offset (in radians) of the left/right chemical sensors.
- turnAngle: The steering angle (in radians) when turning toward higher trail concentrations.

```go
p := arts.NewPhysarum(45000, 110, 0, 0)
```

![](../images/physarum.png)

## Differential Growth

`Differential Growth` simulates organic coral and lichen-like folding loops via spring attraction, local repulsion, and dynamic edge subdivision, rendering history trails as luminous strata.

### parameters

- steps: The number of growth simulation iterations.
- maxNodes: The maximum number of vertices allowed in the closed loop.
- minDist: The target equilibrium spring length between adjacent nodes.
- maxDist: The edge length threshold above which a new midpoint node is inserted.
- repulsionRadius: The spatial radius within which non-adjacent nodes repel each other.

```go
dg := arts.NewDifferentialGrowth(220, 1600, 3.5, 7.5, 18.0)
```

![](../images/differentialgrowth.png)

## Chladni

`Chladni` simulates cymatic standing-wave resonance plates where stochastic particles migrate toward zero-amplitude nodal lines of 2D Chladni eigenmodes.

### parameters

- particleNum: The number of acoustic sand particles to simulate.
- m, n: The harmonic mode numbers of the standing wave.
- a, b: The superposition weights of the two orthogonal eigenmodes.

```go
ch := arts.NewChladni(90000, 7, 3, 1.0, -1.0)
```

![](../images/chladni.png)

## Voronoi Crystal

`Voronoi Crystal` generates Lloyd-relaxed Voronoi cellular stained-glass crystals with directional facet shading, inner sub-facets, and glowing mineral veins.

### parameters

- cellNum: The number of Voronoi seed sites.
- relaxSteps: The number of Lloyd centroidal relaxation iterations.

```go
vc := arts.NewVoronoiCrystal(85, 4)
```

![](../images/voronoicrystal.png)

## Ink Landscape

`Ink Landscape` generates traditional East Asian *Shan Shui* ink-wash mountain landscapes with atmospheric fog perspective, FBM ridgelines, vertical *cunfa* texture strokes, and a glowing sun/moon disc.

### parameters

- layerNum: The number of receding mountain ridgeline layers.
- roughness: Controls the fractal Brownian motion terrain relief amplitude.

```go
il := arts.NewInkLandscape(8, 0.45)
```

![](../images/inklandscape.png)

## Gray Scott

`Gray Scott` simulates reaction-diffusion morphogenesis (Turing patterns) on a 2D chemical grid with spatial parameter gradients and 3D normal-mapped specular lighting.

### parameters

- steps: The number of reaction-diffusion PDE integration steps.
- feed: The base chemical feed rate `F`.
- kill: The base chemical kill rate `k`.

```go
gs := arts.NewGrayScott(3200, 0.037, 0.060)
```

![](../images/grayscott.png)

## Lissajous Ribbon

`Lissajous Ribbon` renders 3D phase-evolving harmonic Lissajous silk ribbons with perspective projection, depth-modulated stroke weights, and additive interference glow.

### parameters

- curves: The number of phase-shifted harmonic ribbon layers.
- freqX, freqY, freqZ: The 3D harmonic oscillation frequencies along X, Y, and Z axes.

```go
lr := arts.NewLissajousRibbon(140, 3.0, 4.0, 5.0)
```

![](../images/lissajousribbon.png)

## Apollonian Gasket

`Apollonian Gasket` generates recursive fractal circle packings using Descartes' Circle Theorem and complex Descartes centers with glowing nested rings and jewel-like shading.

### parameters

- maxDepth: The maximum recursion depth for Soddy circle subdivision.
- minRadius: The minimum pixel radius threshold for packed circles.

```go
ag := arts.NewApollonianGasket(7, 2.5)
```

![](../images/apolloniangasket.png)

## Space Colonization

`Space Colonization` grows organic bioluminescent mycelial and vascular branching networks toward nutrient attractors, with Strahler-like branch thickness tapering and glowing spore tips.

### parameters

- attractorNum: The number of nutrient attractor points distributed across the canvas.
- steps: The number of network growth iterations.
- stepDist: The branch extension step length per iteration.
- killDist: The attraction radius within which attractors influence growing nodes.

```go
sc := arts.NewSpaceColonization(1800, 160, 6.0, 55.0)
```

![](../images/spacecolonization.png)

## Truchet Flow

`Truchet Flow` constructs multi-scale quadtree Smith-Truchet quarter-circle tilings with interlocking multi-band ribbons and under/over weave shadows.

### parameters

- baseGrid: The initial coarse grid resolution along each axis.
- maxDepth: The maximum quadtree subdivision depth.

```go
tf := arts.NewTruchetFlow(6, 3)
```

![](../images/truchetflow.png)

## Aurora Wave

`Aurora Wave` renders boreal geomagnetic light curtains and starfields using FBM-warped vertical ray deposition, altitude-based spectral transitions, and water reflections.

### parameters

- curtainNum: The number of overlapping aurora ribbon curtains.
- ribbonSteps: The horizontal sampling density along each aurora curtain.
- turbulence: Controls the FBM fold amplitude and curtain undulation.

```go
aw := arts.NewAuroraWave(4, 220, 1.0)
```

![](../images/aurorawave.png)

## Fractal Flame

`Fractal Flame` implements Scott Draves' Iterated Function System (IFS) cosmic flame algorithm with non-linear variations (sinusoidal, spherical, swirl, horseshoe, handkerchief), dihedral rotational symmetry, and log-density tone mapping.

### parameters

- iterations: The number of chaos-game IFS trajectory iterations to accumulate.
- symmetry: The dihedral rotational symmetry order.

```go
ff := arts.NewFractalFlame(1800000, 5)
```

![](../images/fractalflame.png)

## Joy Division

`Joy Division` renders pulsar elevation ridgelines inspired by CP 1919 (*Unknown Pleasures*) with painter's algorithm horizon occlusion, FBM ridge synthesis, and luminous crest strokes.

### parameters

- lineNum: The number of horizontal pulsar scanlines stacked vertically.
- amplitude: Controls the vertical ridge peak elevation multiplier.
- frequency: Controls the horizontal FBM spatial frequency of the ridges.

```go
jd := arts.NewJoyDivision(72, 1.0, 1.0)
```

![](../images/joydivision.png)

## Boids Flock

`Boids Flock` simulates Craig Reynolds' emergent flocking choreography (separation, alignment, cohesion, and vortex attractors) and renders tapered calligraphic silk ribbon trails with glowing heads.

### parameters

- boidNum: The number of autonomous boid agents in the flock.
- steps: The number of simulation steps recorded along each trajectory.
- perceptionRadius: The neighborhood radius within which boids align, cohere, and separate.

```go
bf := arts.NewBoidsFlock(240, 140, 52.0)
```

![](../images/boidsflock.png)

## Islamic Star

`Islamic Star` generates Moroccan *Zellige* and Persian *Girih* rosettes using Hankin's polygons-in-contact method on an octagonal/square tessellation with over-under woven gold-leaf strapwork.

### parameters

- gridSize: The number of tiling cells along each axis of the tessellation.
- contactAngle: The Hankin ray contact angle (in degrees) controlling star point geometry.
- strapWidth: The pixel width of the woven interlacing ribbon straps.

```go
is := arts.NewIslamicStar(4, 67.5, 4.5)
```

![](../images/islamicstar.png)

## Delaunay Stained

`Delaunay Stained` constructs a Bowyer-Watson Delaunay triangulation stained-glass mosaic with 3D normal-mapped facet illumination, inner jewel sub-facets, and leaded gold-vein borders.

### parameters

- siteNum: The number of seed vertices in the triangulation.
- jitter: Controls vertex displacement from a stratified radial/grid distribution.

```go
ds := arts.NewDelaunayStained(220, 0.65)
```

![](../images/delaunaystained.png)

## Phyllotaxis

`Phyllotaxis` generates golden-angle botanical spiral primordia (Vogel's model) with Fibonacci parastichy veins, shaded 3D botanical scales/florets, and bioluminescent pollen highlights.

### parameters

- seedCount: The number of botanical primordia florets arranged along the spiral.
- divergenceAngle: The divergence angle in degrees between successive florets (typically the golden angle `137.50776405`).
- scaleFactor: The radial spacing constant scaling the square-root spiral radius.

```go
pt := arts.NewPhyllotaxis(420, 137.50776405, 11.5)
```

![](../images/phyllotaxis.png)

## Poincare Disk

`Poincare Disk` generates Escher-inspired hyperbolic `{p, q}` tessellations inside the Poincaré disk using Möbius gyrovector transformations, geodesic circular arcs, and jewel-like cell shading.

### parameters

- p: The number of sides of each hyperbolic polygon (`(p-2)*(q-2) > 4`).
- q: The number of polygons meeting at each vertex.
- depth: The recursion depth of Möbius edge reflections across geodesic arcs.

```go
pd := arts.NewPoincareDisk(7, 3, 5)
```

![](../images/poincaredisk.png)

## Suminagashi

`Suminagashi` simulates Japanese floating-ink marbling using area-preserving radial fluid lens deformations, comb-tine rake advection, and curl-noise micro-turbulence on Washi paper.

### parameters

- dropCount: The number of concentric ink drops deposited into the water bath.
- combPasses: The number of directional comb-rake advection passes across the surface.
- swirlStrength: Controls the curl-noise vortex and tine displacement intensity.

```go
sm := arts.NewSuminagashi(48, 7, 1.0)
```

![](../images/suminagashi.png)

## Aizawa Attractor

`Aizawa Attractor` integrates the 3D Aizawa chaotic strange attractor and projects millions of trajectory samples through a thin-lens camera with Depth-of-Field (DoF) circle-of-confusion bokeh splatting.

### parameters

- iterations: The number of 4th-order Runge-Kutta trajectory steps to accumulate.
- dofStrength: Controls the camera aperture and Depth-of-Field bokeh blur radius.

```go
aa := arts.NewAizawaAttractor(900000, 1.0)
```

![](../images/aizawaattractor.png)

## Penrose Tiling

`Penrose Tiling` constructs 5-fold aperiodic quasicrystal P2/P3 tilings via recursive Robinson golden-triangle deflation, decorated with De Bruijn Ammann bars, shaded facets, and golden vertices.

### parameters

- subdivisions: The number of golden-ratio Robinson triangle deflation generations.

```go
pt := arts.NewPenroseTiling(6)
```

![](../images/penrosetiling.png)

## Isometric City

`Isometric City` generates a *Monument Valley*-inspired 2.5D axonometric architectural citadel with tiered towers, courtyards, arched colonnades, sky-bridges, staircases, and directional sun/shadow illumination.

### parameters

- gridSize: The number of axonometric city blocks along each horizontal axis.
- density: Controls structural height variation, bridge connectivity, and architectural density.

```go
ic := arts.NewIsometricCity(11, 1.0)
```

![](../images/isometriccity.png)

## DLA Frost

`DLA Frost` simulates anisotropic 6-fold Diffusion-Limited Aggregation (DLA) crystal dendrite growth with age-stratified Oklab ice gradients, specular crystal facets, and bioluminescent frost tips.

### parameters

- particleCount: The target number of aggregated crystal particles in the cluster.
- stickiness: Controls the 6-fold hexagonal dendritic anisotropy and sticking probability.

```go
df := arts.NewDLAFrost(18000, 0.55)
```

![](../images/dlafrost.png)

## Impasto Flow

`Impasto Flow` renders multi-vortex fluid dynamics with dense fibrous impasto brushstrokes, 3D specular ridge highlights, and dark ambient troughs.

### parameters

- strokeCount: The number of layered fibrous impasto brush ribbons to trace.
- steps: The number of advection integration steps per stroke.
- fieldScale: Controls the spatial frequency of the curl-noise and vortex velocity field.

```go
iflow := arts.NewImpastoFlow(3200, 520, 0.0032)
```

![](../images/impastoflow.png)

## Voxel Terrain

`Voxel Terrain` generates an isometric micro-voxel stepped topographic canyon with stratified thermal/oceanic contour bands and directional sun-cast shadows.

### parameters

- gridRes: The horizontal resolution of the isometric voxel heightfield grid.
- heightScale: Controls the vertical terrace extrusion multiplier and cliff relief.
- contrast: Controls the shadow depth and top/side facet illumination contrast.

```go
vt := arts.NewVoxelTerrain(340, 2.6, 1.0)
```

![](../images/voxelterrain.png)

## Packed Capsules

`Packed Capsules` generates a multi-scale collision-free packing of oriented flow-aligned capsules, target rings, and micro-dots with crisp ink outlines and elevation drop shadows.

### parameters

- capsuleCount: The number of candidate capsules and circles to pack into the flow field.
- maxScale: The maximum size multiplier for primary focal capsules.
- minScale: The minimum size multiplier for interstitial micro-capsules and dots.

```go
pc := arts.NewPackedCapsules(2200, 7.2, 3.6)
```

![](../images/packedcapsules.png)

## Orphic Discs

`Orphic Discs` creates Robert & Sonia Delaunay-inspired Orphic Cubism compositions with interlocking eccentric target discs, quadrant chord splits, and stippled gouache paper grain.

### parameters

- discCount: The number of nested eccentric disc assemblies distributed across the canvas.

```go
od := arts.NewOrphicDiscs(34)
```

![](../images/orphicdiscs.png)

## Nautilus Spiral

`Nautilus Spiral` constructs a golden logarithmic nautilus shell cross-section with curved hydrostatic septa, stippled biomineral chambers, and a warm Voronoi mosaic backdrop.

### parameters

- turns: The number of full logarithmic spiral revolutions.
- chambersPerTurn: The number of subdivided shell chambers per revolution.

```go
ns := arts.NewNautilusSpiral(5, 68)
```

![](../images/nautilusspiral.png)

## Mosaic Sun

`Mosaic Sun` renders a stained-glass mosaic solar mandala with radiating alternating warm/cool rays, concentric ring tesserae, and swirling starry sky corona arcs.

### parameters

- ringCount: The number of concentric mosaic tile rings radiating from the solar core.
- rayDensity: Controls the angular subdivision density of the radiating solar beams.

```go
ms := arts.NewMosaicSun(32, 1.0)
```

![](../images/mosaicsun.png)

## Tangent Cones

`Tangent Cones` renders 3D conical relief discs with eccentric apexes, tangent-wrap angular shading, cast drop shadows, and fine radial/concentric engraving rings.

### parameters

- coneCount: The number of non-overlapping and nested 3D conical discs.

```go
tc := arts.NewTangentCones(26)
```

![](../images/tangentcones.png)
