# generativeart

[![Go](https://github.com/jdxyw/generativeart/actions/workflows/go.yml/badge.svg)](https://github.com/jdxyw/generativeart/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jdxyw/generativeart)](https://goreportcard.com/report/github.com/jdxyw/generativeart)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/jdxyw/generativeart/master/LICENSE)

`generativeart` is a `Go` package to generate many kinds of generative art. The goal is to collect some excellent generative art (implemented in `R` or `Processing`), and rewrite them in `Go` again. I would paste the original link at the end of this README(If I remember, you can also submit a PR if you found I missed something.). Currently, it supports the following type.

# Table of Contents

- [generativeart](#generativeart)
- [Table of Contents](#table-of-contents)
  - [Buy me a coffee](#buy-me-a-coffee)
  - [Contributing](#contributing)
  - [Types supported](#types-supported)
  - [Install](#install)
  - [Art Type](#art-type)
  - [Docs](#docs)
  - [Examples](#examples)
  - [General Options](#general-options)
  - [Usage and example](#usage-and-example)
    - [Junas](#junas)
    - [Random Shapes](#random-shapes)
    - [Color Circle2](#color-circle2)
    - [Circle Grid](#circle-grid)
    - [Circle Composes Circle](#circle-composes-circle)
    - [Pixel Hole](#pixel-hole)
    - [Dots Wave](#dots-wave)
    - [Contour Line](#contour-line)
    - [Noise Line](#noise-line)
    - [Dot Line](#dot-line)
    - [Ocean Fish](#ocean-fish)
    - [Circle Loop](#circle-loop)
    - [Domain Warp](#domain-warp)
    - [Circle Noise](#circle-noise)
    - [Perlin Perls](#perlin-perls)
    - [Color Canva](#color-canva)
    - [Julia Set](#julia-set)
    - [Black Hole](#black-hole)
    - [Silk Sky](#silk-sky)
    - [Circle Move](#circle-move)
    - [Random Circle](#random-circle)
    - [Curl Flow](#curl-flow)
    - [Sand Ring](#sand-ring)
    - [Watercolor](#watercolor)
    - [Guilloche](#guilloche)
    - [Clifford Nebula](#clifford-nebula)
    - [Physarum](#physarum)
    - [Differential Growth](#differential-growth)
    - [Chladni](#chladni)
    - [Voronoi Crystal](#voronoi-crystal)
    - [Ink Landscape](#ink-landscape)
    - [Gray Scott](#gray-scott)
    - [Lissajous Ribbon](#lissajous-ribbon)
    - [Apollonian Gasket](#apollonian-gasket)
    - [Space Colonization](#space-colonization)
    - [Truchet Flow](#truchet-flow)
    - [Aurora Wave](#aurora-wave)
    - [Fractal Flame](#fractal-flame)
    - [Joy Division](#joy-division)
    - [Boids Flock](#boids-flock)
    - [Islamic Star](#islamic-star)
    - [Delaunay Stained](#delaunay-stained)
    - [Phyllotaxis](#phyllotaxis)
    - [Poincare Disk](#poincare-disk)
    - [Suminagashi](#suminagashi)
    - [Aizawa Attractor](#aizawa-attractor)
    - [Penrose Tiling](#penrose-tiling)
    - [Isometric City](#isometric-city)
    - [DLA Frost](#dla-frost)
    - [Impasto Flow](#impasto-flow)
    - [Voxel Terrain](#voxel-terrain)
    - [Packed Capsules](#packed-capsules)
    - [Orphic Discs](#orphic-discs)
    - [Nautilus Spiral](#nautilus-spiral)
    - [Mosaic Sun](#mosaic-sun)
    - [Tangent Cones](#tangent-cones)
  - [TODO](#todo)
  - [Contribution](#contribution)

## Buy me a coffee

If you found this package is useful, maybe you could buy me a coffee from [Paypal](https://www.paypal.me/jdxyw).

## Contributing

I am glad to see someone to contribute to this repo, please follow the [CONTRIBUTING.md](https://github.com/jdxyw/generativeart/blob/master/CONTRIBUTING.md) for more information. Or you can just open an issue or PR.

## Types supported

This package is still working in progress. More types would be added. Welcome anyone interested in this field to submit your PR.

- Maze
- Julia Set
- Random Circle Trails
- Silk Smoke
- Spiral Square
- Square Grid
- Circle Line
- Circle Loop
- Silk Sky
- Dot Line
- Swirl
- Point Ribbon
- Janus
- Random Shapes
- Color Circle
- Color Circle2
- Circle Grid
- Contour Line
- Noise Line
- Ocean Fish
- Circle Loop2
- Pixel Hole
- Dots Wave
- Circle Move
- Circle Noise
- Yarn
- Black Hole
- Color Canva
- Domain Warp
- Perlin Perls
- Curl Flow
- Sand Ring
- Watercolor
- Guilloche
- Clifford Nebula
- Physarum
- Differential Growth
- Chladni
- Voronoi Crystal
- Ink Landscape
- Gray Scott
- Lissajous Ribbon
- Apollonian Gasket
- Space Colonization
- Truchet Flow
- Aurora Wave
- Fractal Flame
- Joy Division
- Boids Flock
- Islamic Star
- Delaunay Stained
- Phyllotaxis
- Poincare Disk
- Suminagashi
- Aizawa Attractor
- Penrose Tiling
- Isometric City
- DLA Frost
- Impasto Flow
- Voxel Terrain
- Packed Capsules
- Orphic Discs
- Nautilus Spiral
- Mosaic Sun
- Tangent Cones

For these kinds of art, the package provides as many parameters to control the appearance.

## Install

The go version I used is `Go 1.16`.

```bash
go get github.com/jdxyw/generativeart
```

## Art Type

```go
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
NewPoincareDisk(p, q, depth int)
NewSuminagashi(dropCount, combPasses int, swirlStrength float64)
NewAizawaAttractor(iterations int, dofStrength float64)
NewPenroseTiling(subdivisions int)
NewIsometricCity(gridSize int, density float64)
NewDLAFrost(particleCount int, stickiness float64)
NewImpastoFlow(strokeCount, steps int, fieldScale float64)
NewVoxelTerrain(gridRes int, heightScale, contrast float64)
NewPackedCapsules(capsuleCount int, maxScale, minScale float64)
NewOrphicDiscs(discCount int)
NewNautilusSpiral(turns int, chambersPerTurn int)
NewMosaicSun(ringCount int, rayDensity float64)
NewTangentCones(coneCount int)
```

## Docs

You could find the docs in the [doc](./docs).

## Examples

You could find examples for all types under [example](./example).

## General Options

```go
type Options struct {
 background  color.RGBA
 foreground  color.RGBA
 lineColor   color.RGBA
 lineWidth   float64
 colorSchema []color.RGBA
 nIters      int
 alpha       int
}
```

The `Options` is a global option for the whole `canva`. It includes those general parameters used by different kinds of types, such as `background`, `lineColor`, and `colorScheme`.

For those parameters specified for different kinds of art types, they have their own `struct`.

## Usage and example

Below are some selective examples. For more examples, you could check the example folder.

### Junas

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(500, 500)
 c.SetBackground(common.Black)
 c.FillBackground()
 c.SetColorSchema(common.DarkRed)
 c.SetForeground(common.LightPink)
 c.Draw(arts.NewJanus(10, 0.2))
 c.ToPNG("janus.png")
}
```

![](images/janus.png)

### Random Shapes

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(500, 500)
 c.SetBackground(common.White)
 c.FillBackground()
 c.SetColorSchema([]color.RGBA{
  {0xCF, 0x2B, 0x34, 0xFF},
  {0xF0, 0x8F, 0x46, 0xFF},
  {0xF0, 0xC1, 0x29, 0xFF},
  {0x19, 0x6E, 0x94, 0xFF},
  {0x35, 0x3A, 0x57, 0xFF},
 })
 c.Draw(arts.NewRandomShape(150))
 c.ToPNG("randomshape.png")
}
```

![](images/randomshape.png)

### Color Circle2

```go
func main() {
 rand.Seed(time.Now().Unix())
 colors := []color.RGBA{
  {0x11, 0x60, 0xC6, 0xFF},
  {0xFD, 0xD9, 0x00, 0xFF},
  {0xF5, 0xB4, 0xF8, 0xFF},
  {0xEF, 0x13, 0x55, 0xFF},
  {0xF4, 0x9F, 0x0A, 0xFF},
 }
 c := generativeart.NewCanva(800, 800)
 c.SetBackground(common.White)
 c.FillBackground()
 c.SetColorSchema(colors)
 c.Draw(arts.NewColorCircle2(30))
 c.ToPNG("colorcircle2.png")
}
```

![](images/colorcircle2.png)

### Circle Grid

```go
func main() {
 rand.Seed(time.Now().Unix())
 colors := []color.RGBA{
  {0xED, 0x34, 0x41, 0xFF},
  {0xFF, 0xD6, 0x30, 0xFF},
  {0x32, 0x9F, 0xE3, 0xFF},
  {0x15, 0x42, 0x96, 0xFF},
  {0x00, 0x00, 0x00, 0xFF},
  {0xFF, 0xFF, 0xFF, 0xFF},
 }
 c := generativeart.NewCanva(500, 500)
 c.SetBackground(color.RGBA{0xDF, 0xEB, 0xF5, 0xFF})
 c.FillBackground()
 c.SetColorSchema(colors)
 c.SetLineWidth(2.0)
 c.Draw(arts.NewCircleGrid(4, 6))
 c.ToPNG("circlegrid.png")
}
```

![](images/circlegrid.png)

### Circle Composes Circle

```go
func main() {
 rand.Seed(time.Now().Unix())
 colors := []color.RGBA{
  {0xF9, 0xC8, 0x0E, 0xFF},
  {0xF8, 0x66, 0x24, 0xFF},
  {0xEA, 0x35, 0x46, 0xFF},
  {0x66, 0x2E, 0x9B, 0xFF},
  {0x43, 0xBC, 0xCD, 0xFF},
 }
 c := generativeart.NewCanva(500, 500)
 c.SetBackground(color.RGBA{8, 10, 20, 255})
 c.FillBackground()
 c.SetColorSchema(colors)
 c.Draw(arts.NewCircleLoop2(7))
 c.ToPNG("colorloop2.png")
}
```

![](images/colorloop2.png)

### Pixel Hole

```go
func main() {
 rand.Seed(time.Now().Unix())
 colors := []color.RGBA{
  {0xF9, 0xC8, 0x0E, 0xFF},
  {0xF8, 0x66, 0x24, 0xFF},
  {0xEA, 0x35, 0x46, 0xFF},
  {0x66, 0x2E, 0x9B, 0xFF},
  {0x43, 0xBC, 0xCD, 0xFF},
 }
 c := generativeart.NewCanva(800, 800)
 c.SetBackground(common.Black)
 c.FillBackground()
 c.SetColorSchema(colors)
 c.SetIterations(1200)
 c.Draw(arts.NewPixelHole(60))
 c.ToPNG("pixelhole.png")

}
```

![](images/pixelhole.png)

### Dots Wave

```go
func main() {
 rand.Seed(time.Now().Unix())
 colors := []color.RGBA{
  {0xFF, 0xBE, 0x0B, 0xFF},
  {0xFB, 0x56, 0x07, 0xFF},
  {0xFF, 0x00, 0x6E, 0xFF},
  {0x83, 0x38, 0xEC, 0xFF},
  {0x3A, 0x86, 0xFF, 0xFF},
 }
 c := generativeart.NewCanva(500, 500)
 c.SetBackground(common.Black)
 c.FillBackground()
 c.SetColorSchema(colors)
 c.Draw(arts.NewDotsWave(300))
 c.ToPNG("dotswave.png")
}
```

![](images/dotswave.png)

### Contour Line

```go
func main() {
 rand.Seed(time.Now().Unix())
 colors := []color.RGBA{
  {0x58, 0x18, 0x45, 0xFF},
  {0x90, 0x0C, 0x3F, 0xFF},
  {0xC7, 0x00, 0x39, 0xFF},
  {0xFF, 0x57, 0x33, 0xFF},
  {0xFF, 0xC3, 0x0F, 0xFF},
 }
 c := generativeart.NewCanva(1600, 1600)
 c.SetBackground(color.RGBA{0x1a, 0x06, 0x33, 0xFF})
 c.FillBackground()
 c.SetColorSchema(colors)
 c.Draw(arts.NewContourLine(500))
 c.ToPNG("contourline.png")
}
```

![](images/contourline.png)

### Noise Line

```go
func main() {
 rand.Seed(time.Now().Unix())
 colors := []color.RGBA{
  {0x06, 0x7B, 0xC2, 0xFF},
  {0x84, 0xBC, 0xDA, 0xFF},
  {0xEC, 0xC3, 0x0B, 0xFF},
  {0xF3, 0x77, 0x48, 0xFF},
  {0xD5, 0x60, 0x62, 0xFF},
 }
 c := generativeart.NewCanva(1000, 1000)
 c.SetBackground(color.RGBA{0xF0, 0xFE, 0xFF, 0xFF})
 c.FillBackground()
 c.SetColorSchema(colors)
 c.Draw(arts.NewNoiseLine(1000))
 c.ToPNG("noiseline.png")
}
```

![](images/noiseline.png)

### Dot Line

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(2080, 2080)
 c.SetBackground(color.RGBA{230, 230, 230, 255})
 c.SetLineWidth(10)
 c.SetIterations(15000)
 c.SetColorSchema(common.Plasma)
 c.FillBackground()
 c.Draw(arts.NewDotLine(100, 20, 50, false))
 c.ToPNG("dotline.png")
}
```

![](images/dotline.png)

### Ocean Fish

```go
func main() {
 rand.Seed(time.Now().Unix())
 colors := []color.RGBA{
  {0xCF, 0x2B, 0x34, 0xFF},
  {0xF0, 0x8F, 0x46, 0xFF},
  {0xF0, 0xC1, 0x29, 0xFF},
  {0x19, 0x6E, 0x94, 0xFF},
  {0x35, 0x3A, 0x57, 0xFF},
 }
 c := generativeart.NewCanva(500, 500)
 c.SetColorSchema(colors)
 c.Draw(arts.NewOceanFish(100, 8))
 c.ToPNG("oceanfish.png")
}
```

![](images/oceanfish.png)

### Circle Loop

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(500, 500)
 c.SetBackground(common.Black)
 c.SetLineWidth(1)
 c.SetLineColor(common.Orange)
 c.SetAlpha(30)
 c.SetIterations(1000)
 c.FillBackground()
 c.Draw(arts.NewCircleLoop(100))
 c.ToPNG("circleloop.png")
}
```

![](images/circleloop.png)

### Domain Warp

```go
func cmap(r, m1, m2 float64) color.RGBA {
 rgb := color.RGBA{
  uint8(common.Constrain(m1*200*r, 0, 255)),
  uint8(common.Constrain(r*200, 0, 255)),
  uint8(common.Constrain(m2*255*r, 70, 255)),
  255,
 }
 return rgb
}

func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(500, 500)
 c.SetBackground(common.Black)
 c.FillBackground()
 c.Draw(arts.NewDomainWrap(0.01, 4,4, 20, cmap))
 c.ToPNG("domainwarp.png")
}
```

![](images/domainwarp.png)

### Circle Noise

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(500, 500)
 c.SetBackground(common.White)
 c.SetAlpha(80)
 c.SetLineWidth(0.3)
 c.FillBackground()
 c.SetIterations(400)
 c.Draw(arts.NewCircleNoise(2000, 60, 80))
 c.ToPNG("circlenoise.png")
}
```

![](images/circlenoise.png)

### Perlin Perls

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(500, 500)
 c.SetBackground(common.White)
 c.SetAlpha(120)
 c.SetLineWidth(0.3)
 c.FillBackground()
 c.SetIterations(200)
 c.Draw(arts.NewPerlinPerls(10, 200, 40, 80))
 c.ToPNG("perlinperls.png")
}
```

![](images/perlinperls.png)

### Color Canva

```go
func main() {
 rand.Seed(time.Now().Unix())
 colors := []color.RGBA{
  {0xF9, 0xC8, 0x0E, 0xFF},
  {0xF8, 0x66, 0x24, 0xFF},
  {0xEA, 0x35, 0x46, 0xFF},
  {0x66, 0x2E, 0x9B, 0xFF},
  {0x43, 0xBC, 0xCD, 0xFF},
 }
 c := generativeart.NewCanva(500, 500)
 c.SetBackground(common.Black)
 c.FillBackground()
 c.SetLineWidth(8)
 c.SetColorSchema(colors)
 c.Draw(arts.NewColorCanve(5))
 c.ToPNG("colorcanva.png")
}
```

![](images/colorcanva.png)

### Julia Set

```go
func julia1(z complex128) complex128 {
 c := complex(-0.1, 0.651)

 z = z*z + c

 return z
}

func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(500, 500)
 c.SetIterations(800)
 c.SetColorSchema(common.Viridis)
 c.FillBackground()
 c.Draw(arts.NewJulia(julia1, 40, 1.5, 1.5))
 c.ToPNG("julia.png")
}
```

![](images/julia.png)

### Black Hole

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(500, 500)
 c.SetBackground(color.RGBA{R: 30, G: 30, B: 30, A: 255})
 c.FillBackground()
 c.SetLineWidth(1.0)
 c.SetLineColor(common.Tomato)
 c.Draw(arts.NewBlackHole(200, 400, 0.01))
 c.ToPNG("blackhole.png")
}
```

![](images/blackhole.png)

### Silk Sky

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetAlpha(10)
 c.Draw(arts.NewSilkSky(15, 5))
 c.ToPNG("silksky.png")
}
```

![](images/silksky.png)

### Circle Move

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(1200, 500)
 c.SetBackground(common.White)
 c.FillBackground()
 c.Draw(arts.NewCircleMove(1000))
 c.ToPNG("circlemove.png")
}
```

![](images/circlemove.png)

### Random Circle

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(500, 500)
 c.SetBackground(common.MistyRose)
 c.SetLineWidth(1.0)
 c.SetLineColor(color.RGBA{
  R: 122,
  G: 122,
  B: 122,
  A: 30,
 })
 c.SetColorSchema(common.Plasma)
 c.SetIterations(4)
 c.FillBackground()
 c.Draw(arts.NewRandCicle(30, 80, 0.2, 2, 10, 30, true))
 c.ToPNG("randcircle.png")
}
```

![](images/randcircle.png)

### Curl Flow

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 14, G: 18, B: 28, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Cyberpunk)
 c.Draw(arts.NewCurlFlow(110, 36, 0.0045))
 c.ToPNG("curlflow.png")
}
```

![Curl Flow](images/curlflow.png)

### Sand Ring

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 8, G: 10, B: 18, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Plasma)
 c.Draw(arts.NewSandRing(22, 22, 1.3))
 c.ToPNG("sandring.png")
}
```

![Sand Ring](images/sandring.png)

### Watercolor

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 247, G: 243, B: 234, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Dunhuang)
 c.SetLineColor(color.RGBA{R: 32, G: 28, B: 38, A: 255})
 c.Draw(arts.NewWatercolor(9, 22, 5, 0.32))
 c.ToPNG("watercolor.png")
}
```

![Watercolor](images/watercolor.png)

### Guilloche

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 12, G: 15, B: 26, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Oceanic)
 c.Draw(arts.NewGuilloche(28, 175, 65, 115))
 c.ToPNG("guilloche.png")
}
```

![Guilloche](images/guilloche.png)

### Clifford Nebula

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 6, G: 6, B: 14, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Inferno)
 c.Draw(arts.NewCliffordNebula(1500000, -1.7, 1.8, -1.9, -0.4))
 c.ToPNG("cliffordnebula.png")
}
```

![Clifford Nebula](images/cliffordnebula.png)

### Physarum

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 6, G: 10, B: 18, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Cyberpunk)
 c.Draw(arts.NewPhysarum(45000, 110, 0, 0))
 c.ToPNG("physarum.png")
}
```

![Physarum Network](images/physarum.png)

### Differential Growth

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 10, G: 12, B: 22, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Sunset)
 c.Draw(arts.NewDifferentialGrowth(220, 1600, 3.5, 7.5, 18.0))
 c.ToPNG("differentialgrowth.png")
}
```

![Differential Growth Coral](images/differentialgrowth.png)

### Chladni

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 8, G: 12, B: 22, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Oceanic)
 c.Draw(arts.NewChladni(90000, 7, 3, 1.0, -1.0))
 c.ToPNG("chladni.png")
}
```

![Chladni Cymatics](images/chladni.png)

### Voronoi Crystal

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 24, G: 22, B: 30, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Dunhuang)
 c.Draw(arts.NewVoronoiCrystal(85, 4))
 c.ToPNG("voronoicrystal.png")
}
```

![Voronoi Agate Crystal](images/voronoicrystal.png)

### Ink Landscape

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 244, G: 239, B: 228, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.UkiyoE)
 c.Draw(arts.NewInkLandscape(8, 0.45))
 c.ToPNG("inklandscape.png")
}
```

![Shan-Shui Ink Landscape](images/inklandscape.png)

### Gray Scott

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 8, G: 12, B: 20, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Viridis)
 c.Draw(arts.NewGrayScott(3200, 0.037, 0.060))
 c.ToPNG("grayscott.png")
}
```

![Gray Scott Turing Coral](images/grayscott.png)

### Lissajous Ribbon

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 6, G: 8, B: 16, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Plasma)
 c.Draw(arts.NewLissajousRibbon(140, 3.0, 4.0, 5.0))
 c.ToPNG("lissajousribbon.png")
}
```

![Lissajous Silk Ribbon](images/lissajousribbon.png)

### Apollonian Gasket

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 10, G: 14, B: 24, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Dunhuang)
 c.Draw(arts.NewApollonianGasket(7, 2.5))
 c.ToPNG("apolloniangasket.png")
}
```

![Apollonian Gasket Astrolabe](images/apolloniangasket.png)

### Space Colonization

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 8, G: 14, B: 18, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Cyberpunk)
 c.Draw(arts.NewSpaceColonization(1800, 160, 6.0, 55.0))
 c.ToPNG("spacecolonization.png")
}
```

![Space Colonization Tree](images/spacecolonization.png)

### Truchet Flow

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 245, G: 240, B: 230, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.UkiyoE)
 c.Draw(arts.NewTruchetFlow(6, 3))
 c.ToPNG("truchetflow.png")
}
```

![Multi-Scale Truchet Flow](images/truchetflow.png)

### Aurora Wave

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 5, G: 8, B: 18, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Oceanic)
 c.Draw(arts.NewAuroraWave(4, 220, 1.0))
 c.ToPNG("aurorawave.png")
}
```

![Aurora Borealis Wave](images/aurorawave.png)

### Fractal Flame

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 4, G: 4, B: 10, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Magma)
 c.Draw(arts.NewFractalFlame(1800000, 5))
 c.ToPNG("fractalflame.png")
}
```

![Fractal Flame IFS](images/fractalflame.png)

### Joy Division

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 8, G: 10, B: 18, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Cyberpunk)
 c.Draw(arts.NewJoyDivision(72, 1.0, 1.0))
 c.ToPNG("joydivision.png")
}
```

![Joy Division Pulsar Waves](images/joydivision.png)

### Boids Flock

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 10, G: 14, B: 24, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Sunset)
 c.Draw(arts.NewBoidsFlock(240, 140, 52.0))
 c.ToPNG("boidsflock.png")
}
```

![Boids Murmuration Flock](images/boidsflock.png)

### Islamic Star

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 18, G: 22, B: 32, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Dunhuang)
 c.Draw(arts.NewIslamicStar(4, 67.5, 4.5))
 c.ToPNG("islamicstar.png")
}
```

![Islamic Girih Star Rosette](images/islamicstar.png)

### Delaunay Stained

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 12, G: 14, B: 22, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Plasma)
 c.Draw(arts.NewDelaunayStained(220, 0.65))
 c.ToPNG("delaunaystained.png")
}
```

![Delaunay Prismatic Stained Glass](images/delaunaystained.png)

### Phyllotaxis

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 12, G: 18, B: 20, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Botanical)
 c.Draw(arts.NewPhyllotaxis(420, 137.50776405, 11.5))
 c.ToPNG("phyllotaxis.png")
}
```

![Golden Angle Phyllotaxis Lotus](images/phyllotaxis.png)

### Poincare Disk

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 8, G: 10, B: 20, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Dunhuang)
 c.Draw(arts.NewPoincareDisk(7, 3, 5))
 c.ToPNG("poincaredisk.png")
}
```

![Poincare Disk](images/poincaredisk.png)

### Suminagashi

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 244, G: 239, B: 229, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.UkiyoE)
 c.Draw(arts.NewSuminagashi(48, 7, 1.0))
 c.ToPNG("suminagashi.png")
}
```

![Suminagashi](images/suminagashi.png)

### Aizawa Attractor

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 5, G: 7, B: 15, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Cyberpunk)
 c.Draw(arts.NewAizawaAttractor(900000, 1.0))
 c.ToPNG("aizawaattractor.png")
}
```

![Aizawa Attractor](images/aizawaattractor.png)

### Penrose Tiling

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 10, G: 14, B: 24, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Oceanic)
 c.Draw(arts.NewPenroseTiling(6))
 c.ToPNG("penrosetiling.png")
}
```

![Penrose Tiling](images/penrosetiling.png)

### Isometric City

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 18, G: 20, B: 34, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Sunset)
 c.Draw(arts.NewIsometricCity(11, 1.0))
 c.ToPNG("isometriccity.png")
}
```

![Isometric City](images/isometriccity.png)

### DLA Frost

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 4, G: 8, B: 18, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Oceanic)
 c.Draw(arts.NewDLAFrost(18000, 0.55))
 c.ToPNG("dlafrost.png")
}
```

![DLA Frost](images/dlafrost.png)

### Impasto Flow

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 8, G: 7, B: 16, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Sunset)
 c.Draw(arts.NewImpastoFlow(3200, 520, 0.0032))
 c.ToPNG("impastoflow.png")
}
```

![Impasto Flow](images/impastoflow.png)

### Voxel Terrain

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 10, G: 12, B: 26, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Oceanic)
 c.Draw(arts.NewVoxelTerrain(340, 2.6, 1.0))
 c.ToPNG("voxelterrain.png")
}
```

![Voxel Terrain](images/voxelterrain.png)

### Packed Capsules

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 238, G: 236, B: 229, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Oceanic)
 c.Draw(arts.NewPackedCapsules(2200, 7.2, 3.6))
 c.ToPNG("packedcapsules.png")
}
```

![Packed Capsules](images/packedcapsules.png)

### Orphic Discs

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 72, G: 178, B: 102, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Youthful)
 c.Draw(arts.NewOrphicDiscs(34))
 c.ToPNG("orphicdiscs.png")
}
```

![Orphic Discs](images/orphicdiscs.png)

### Nautilus Spiral

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 18, G: 24, B: 34, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Dunhuang)
 c.Draw(arts.NewNautilusSpiral(5, 68))
 c.ToPNG("nautilusspiral.png")
}
```

![Nautilus Spiral](images/nautilusspiral.png)

### Mosaic Sun

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 22, G: 26, B: 48, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Sunset)
 c.Draw(arts.NewMosaicSun(32, 1.0))
 c.ToPNG("mosaicsun.png")
}
```

![Mosaic Sun](images/mosaicsun.png)

### Tangent Cones

```go
func main() {
 rand.Seed(time.Now().Unix())
 c := generativeart.NewCanva(600, 600)
 c.SetBackground(color.RGBA{R: 234, G: 88, B: 56, A: 255})
 c.FillBackground()
 c.SetColorSchema(common.Sunset)
 c.Draw(arts.NewTangentCones(26))
 c.ToPNG("tangentcones.png")
}
```

![Tangent Cones](images/tangentcones.png)

## TODO

See the full architectural and feature roadmap in [docs/roadmap.md](./docs/roadmap.md):

1. **Dynamic Evolution & Seamless Loop Animation (`DrawTimelapse`)**: Support frame recording, 4D Simplex/Curl noise seamless loops, and GIF/APNG/MP4 export for time-evolving systems (`Physarum`, `DifferentialGrowth`, `GrayScott`, `BoidsFlock`, `CurlFlow`).
2. **Pure-Go WebAssembly (WASM) Interactive Playground**: Compile `generativeart` to WASM (`GOOS=js GOARCH=wasm`) so users can tweak parameters, seeds, and palettes live in the browser gallery (`index.html`).
3. **CLI Tool (`cmd/genart`) & Reproducible Seed Metadata**: Deterministic per-canvas RNG seeding, PNG `tEXt` recipe metadata embedding, and a CLI for 4K/8K wallpaper generation.
4. **Image-Guided & Typography-Driven Art**: Reference image luminance/gradient fields, Weighted Voronoi Stippling portraits, TSP continuous-line art, and Edge-Tangent Flow painterly brushes.
5. **Vector Export (`SVG` / `PDF`) & Pen Plotter Ecosystem**: Native vector output via `tdewolff/canvas`, Jobard-Lefer evenly-spaced streamlines, and 3D hidden-line removal.
6. **3D SDF Raymarching & Mathematical Sculptures**: Pure-Go parallel Signed Distance Field raymarcher for `Mandelbulb`, `Gyroid` minimal surfaces, and 3D Kleinian/Apollonian sphere packings.
7. **Next-Wave `arts/` Paradigms (20+ Planned Generators)**:
   - *Flow & Topology*: `EvenStreamlines` (Jobard-Lefer), `MarchingIsolines`, `MagneticDipole`
   - *3D Chaos & Complex Dynamics*: `AizawaAttractor` (3D DoF bokeh), `Buddhabrot`, `NewtonBasin`, `LyapunovFractal`
   - *Hyperbolic & Quasicrystals*: `PoincareDisk` (`{p,q}` hyperbolic), `PenroseTiling`, `AmmannBeenker`
   - *Morphogenesis & Physics*: `DLAFrost`, `CyclicAutomata` (BZ spirals), `VortexFluid`, `CircuitWFC`
   - *2.5D Architecture & Origami*: `MondrianSubdivision`, `IsometricCity`, `OrigamiTessellation`
   - *Natural Media & Script*: `Suminagashi` (fluid ink marbling), `AsemicCalligraphy`, `SashikoWeave`

## Contribution

Thanks for the following sites and repos, I got lots of ideas, inspiration, code, and tricks from them. The list would be very long; sorry for forgetting some of them.

- <https://inconvergent.net/>
- <https://fronkonstin.com/>
- <https://github.com/aschinchon/cyclic-cellular-automata>
- <https://github.com/armdz/ProcessingSketchs>
- <https://github.com/Mr-Slesser/Generative-Art-And-Fractals>
- <https://github.com/cdr6934/Generative-Processing-Experiments>
- <https://github.com/pkd2512/inktober2017>
- <http://blog.dragonlab.de/2015/03/generative-art-week-1>
- <https://editor.p5js.org/kenekk1/sketches/Ly-5XYvKX>
- <http://paulbourke.net/fractals/peterdejong/>
- <https://editor.p5js.org/kenekk1/sketches/O44Dln5oo>
- <https://openprocessing.org/sketch/1071233>
- <https://twitter.com/okazz_>
- <https://openprocessing.org/sketch/738638>
- <https://openprocessing.org/sketch/1102157>
- <https://openprocessing.org/sketch/1071233>
- <https://openprocessing.org/user/139364>
- <https://openprocessing.org/sketch/792407>
- <https://www.iquilezles.org/www/articles/warp/warp.htm>
