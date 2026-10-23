package arts

import (
	"image/color"
	"math"
	"math/rand"
	"runtime"
	"sync"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type malachiteBanding struct {
	nodes int
	rings int
}

// NewMalachiteBanding returns a malachiteBanding engine that renders a polished
// Congolese botryoidal malachite and azurite mineral cross-section (孔雀石与蓝铜矿钟乳状同心截面)
// with smooth-union stalactite nuclei and Kajiya-Kay chatoyant velvet sheen.
func NewMalachiteBanding(nodes, rings int) *malachiteBanding {
	if nodes <= 0 {
		nodes = 14
	}
	if rings <= 0 {
		rings = 22
	}
	return &malachiteBanding{
		nodes: nodes,
		rings: rings,
	}
}

// Generative renders the polished botryoidal malachite & azurite mineral slab.
func (mb *malachiteBanding) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := c.Width(), c.Height()
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0

	// Authentic Botryoidal Malachite, Chrysocolla, Azurite & Pyrite Gold Palette
	velvetBlackGreen := color.RGBA{R: 5, G: 28, B: 18, A: 255}
	deepMalachite := color.RGBA{R: 12, G: 76, B: 48, A: 255}
	imperialEmerald := color.RGBA{R: 26, G: 146, B: 88, A: 255}
	malachiteBand := color.RGBA{R: 62, G: 204, B: 132, A: 255}
	paleMintSilk := color.RGBA{R: 164, G: 244, B: 196, A: 255}
	chrysocollaTeal := color.RGBA{R: 32, G: 168, B: 176, A: 255}
	azuriteCobalt := color.RGBA{R: 20, G: 52, B: 142, A: 255}
	pyriteVein := color.RGBA{R: 238, G: 186, B: 64, A: 255}

	rc.FillBackground(velvetBlackGreen)
	rc.SetBlendMode(render.BlendOver)

	rng := rand.New(rand.NewSource(c.Seed() + 606))
	ne := common.NewNoiseEngine(c.Seed() + 919)

	type stalactiteNode struct {
		x, y   float64
		radius float64
		phase  float64
	}

	numNodes := mb.nodes
	if numNodes < 6 {
		numNodes = 14
	}
	nodes := make([]stalactiteNode, numNodes)
	for i := 0; i < numNodes; i++ {
		u := math.Sqrt((float64(i) + 0.4) / float64(numNodes))
		th := float64(i)*2.39996323 + (rng.Float64()-0.5)*0.28
		r := u * 0.38 * math.Min(fw, fh)
		nodes[i] = stalactiteNode{
			x:      0.5*fw + r*math.Cos(th),
			y:      0.5*fh + r*math.Sin(th),
			radius: (34.0 + 36.0*rng.Float64()) * scale,
			phase:  rng.Float64() * 0.35,
		}
	}

	// Polynomial smooth minimum smin(a, b, k) for merging botryoidal stalactite rings
	smin := func(a, b, k float64) float64 {
		h := common.Constrain(0.5+0.5*(b-a)/k, 0.0, 1.0)
		return b*(1.0-h) + a*h - k*h*(1.0-h)
	}

	smoothK := 46.0 * scale
	// Evaluate composite botryoidal mineral growth field d(x, y)
	evalMineralField := func(x, y float64) float64 {
		wx := ne.FBM2D(x/(115.0*scale), y/(115.0*scale), 4, 2.05, 0.52) * 16.0 * scale
		wy := ne.FBM2D((x+400)/(115.0*scale), (y+400)/(115.0*scale), 4, 2.05, 0.52) * 16.0 * scale
		px := x + wx
		py := y + wy

		dMin := 1e9
		for i, nd := range nodes {
			d := math.Hypot(px-nd.x, py-nd.y) - nd.radius*0.35 + nd.phase*18.0*scale
			if i == 0 {
				dMin = d
			} else {
				dMin = smin(dMin, d, smoothK)
			}
		}
		// Fine crystalline micro-undulation along the concentric growth fronts
		micro := ne.FBM2D(px/(28.0*scale), py/(28.0*scale), 3, 2.1, 0.5) * 3.2 * scale
		return (dMin + micro) / (155.0 * scale)
	}

	ringFreq := float64(mb.rings)
	if ringFreq < 10 {
		ringFreq = 22
	}

	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
	pixCols := make([]color.RGBA, w*h)
	eps := 1.2 * scale

	var wg sync.WaitGroup
	for wIdx := 0; wIdx < workers; wIdx++ {
		y0 := (wIdx * h) / workers
		y1 := ((wIdx + 1) * h) / workers
		wg.Add(1)
		go func(rowStart, rowEnd int) {
			defer wg.Done()
			for y := rowStart; y < rowEnd; y++ {
				fy := float64(y) + 0.5
				rowOff := y * w
				for x := 0; x < w; x++ {
					fx := float64(x) + 0.5
					d := evalMineralField(fx, fy)
					dXp := evalMineralField(fx+eps, fy)
					dYp := evalMineralField(fx, fy+eps)

					// Analytical gradient & tangent for Kajiya-Kay chatoyant silk sheen
					gx := (dXp - d) / eps
					gy := (dYp - d) / eps
					gLen := math.Hypot(gx, gy) + 1e-6
					nx, ny := gx/gLen, gy/gLen

					// Multi-harmonic Liesegang banding profile characteristic of natural malachite
					phi := d * ringFreq * math.Pi
					w1 := 0.5 + 0.5*math.Sin(phi)
					w2 := 0.5 + 0.5*math.Cos(phi*2.0+0.6)
					w3 := 0.5 + 0.5*math.Sin(phi*4.0-0.3)
					bandVal := common.Constrain(0.52*w1+0.32*w2+0.16*w3, 0, 1)

					var col color.RGBA
					// Outer interstitial vugs transition into Chrysocolla Teal & Royal Azurite Blue
					if d > 0.72 {
						azT := common.Constrain((d-0.72)/0.45, 0, 1)
						if azT < 0.35 {
							col = render.LerpOklab(deepMalachite, chrysocollaTeal, azT/0.35)
						} else if azT < 0.52 {
							// Thin golden pyrite boundary seam between malachite and azurite
							seam := math.Sin(((azT - 0.35) / 0.17) * math.Pi)
							col = render.LerpOklab(chrysocollaTeal, pyriteVein, seam*0.88)
						} else {
							subWave := 0.5 + 0.5*math.Sin(phi*1.8)
							col = render.LerpOklab(azuriteCobalt, velvetBlackGreen, (1.0-subWave)*0.55)
						}
					} else {
						// Concentric Malachite Bullseye Rings
						switch {
						case bandVal < 0.22:
							col = render.LerpOklab(velvetBlackGreen, deepMalachite, bandVal/0.22)
						case bandVal < 0.52:
							col = render.LerpOklab(deepMalachite, imperialEmerald, (bandVal-0.22)/0.30)
						case bandVal < 0.80:
							col = render.LerpOklab(imperialEmerald, malachiteBand, (bandVal-0.52)/0.28)
						default:
							col = render.LerpOklab(malachiteBand, paleMintSilk, (bandVal-0.80)/0.20)
						}
					}

					// Anisotropic Chatoyant Velvet Highlight (light glinting across radial acicular crystal fibers)
					lightDot := -0.62*nx - 0.78*ny
					chatoyancy := math.Pow(common.Constrain(0.5+0.5*lightDot, 0, 1), 6.0)
					// Secondary cross-fiber silk band
					silkBand := math.Pow(math.Abs(nx*0.78-ny*0.62), 8.0) * 0.38
					sheen := common.Constrain(chatoyancy*0.32+silkBand, 0, 0.55)
					col = render.LerpOklab(col, paleMintSilk, sheen)

					// Sharp dark Liesegang growth-pause rings
					if math.Abs(math.Sin(phi*0.5)) < 0.045 {
						col = render.LerpOklab(col, velvetBlackGreen, 0.52)
					}

					pixCols[rowOff+x] = col
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	for y := 0; y < h; y++ {
		rowOff := y * w
		for x := 0; x < w; x++ {
			rc.BlendPixel(x, y, pixCols[rowOff+x], 1.0)
		}
	}

	c.MergeRenderCanvas(rc)
}
