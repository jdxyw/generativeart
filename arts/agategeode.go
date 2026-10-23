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
	"github.com/tdewolff/canvas"
)

type agateGeode struct {
	bandCount       int
	crystalClusters int
}

// NewAgateGeode returns an agateGeode engine that renders a cross-sectioned Brazilian
// amethyst and fortification agate geode (紫水晶洞与缠丝玛瑙晶簇切面) with rhythmic
// chalcedony mineral bands, gold-leaf veins, and 3D inward-pointing hexagonal quartz crystals.
func NewAgateGeode(bandCount, crystalClusters int) *agateGeode {
	if bandCount <= 0 {
		bandCount = 26
	}
	if crystalClusters <= 0 {
		crystalClusters = 36
	}
	return &agateGeode{
		bandCount:       bandCount,
		crystalClusters: crystalClusters,
	}
}

// Generative renders the cross-sectioned fortification agate geode and amethyst crystal cavity.
func (ag *agateGeode) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := c.Width(), c.Height()
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0
	cx, cy := 0.5*fw, 0.5*fh

	// Brazilian Amethyst, Fortification Agate & Pyrite Gold Palette
	basaltMatrix := color.RGBA{R: 12, G: 11, B: 18, A: 255}
	druzyAbyss := color.RGBA{R: 16, G: 8, B: 28, A: 255}
	deepAmethyst := color.RGBA{R: 54, G: 18, B: 96, A: 255}
	royalViolet := color.RGBA{R: 124, G: 42, B: 178, A: 255}
	orchidLavender := color.RGBA{R: 204, G: 138, B: 238, A: 255}
	roseQuartz := color.RGBA{R: 226, G: 96, B: 142, A: 255}
	chalcedonyTeal := color.RGBA{R: 44, G: 166, B: 168, A: 255}
	 milkyQuartz := color.RGBA{R: 244, G: 238, B: 248, A: 255}
	pyriteGold := color.RGBA{R: 242, G: 188, B: 62, A: 255}
	diamondGlint := color.RGBA{R: 255, G: 252, B: 242, A: 255}

	rc.FillBackground(basaltMatrix)
	rc.SetBlendMode(render.BlendOver)

	ne := common.NewNoiseEngine(c.Seed() + 505)
	baseR := 0.44 * math.Min(fw, fh)

	// Organic multi-lobed geode nodule distance field:
	// u = 0 at outer basalt rind, u = 1 at inner crystal cavity boundary
	geodeField := func(x, y float64) float64 {
		dx := x - cx
		dy := y - cy
		r := math.Hypot(dx, dy)
		th := math.Atan2(dy, dx)

		// 5-lobed fortification geometry + angular crystal-habit faceting
		lobe := 1.0 + 0.09*math.Cos(3.0*th-0.4) + 0.06*math.Sin(5.0*th+0.8) + 0.035*math.Cos(9.0*th)
		// Fortification sharp corner creases (缠丝玛瑙折角堡垒纹)
		fort := 0.028 * math.Abs(math.Sin(6.0*th+0.3))
		n := ne.FBM2D(x/(95.0*scale), y/(95.0*scale), 4, 2.05, 0.52) * 0.065

		rOuter := baseR * (lobe - fort + n)
		return 1.0 - (r / math.Max(1.0, rOuter))
	}

	bands := float64(ag.bandCount)
	if bands < 12 {
		bands = 26
	}

	// Rich cyclic mineral band palette
	bandStops := []color.RGBA{
		deepAmethyst,
		royalViolet,
		milkyQuartz,
		roseQuartz,
		deepAmethyst,
		pyriteGold,
		chalcedonyTeal,
		milkyQuartz,
		royalViolet,
		orchidLavender,
	}

	// 1. Parallel Evaluation of Polished Fortification Agate Slab & Sub-Surface Specular
	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
	slabCols := make([]color.RGBA, w*h)
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
					u := geodeField(fx, fy)

					// Outside the geode nodule: rough dark volcanic basalt matrix
					if u < -0.02 {
						grain := ne.FBM2D(fx/(18.0*scale), fy/(18.0*scale), 3, 2.2, 0.55)
						cBg := render.LerpOklab(basaltMatrix, color.RGBA{26, 24, 36, 255}, common.Constrain(0.5+grain*0.6, 0, 1))
						slabCols[rowOff+x] = cBg
						continue
					}

					// Outer golden pyrite & celadonite crust (u in [-0.02, 0.04])
					if u < 0.04 {
						tCrust := (u + 0.02) / 0.06
						col := render.LerpOklab(basaltMatrix, pyriteGold, math.Sin(tCrust*math.Pi))
						slabCols[rowOff+x] = col
						continue
					}

					// Inner druzy cavity hollow (u > 0.54)
					if u > 0.54 {
						depthT := common.Constrain((u-0.54)/0.46, 0, 1)
						col := render.LerpOklab(deepAmethyst, druzyAbyss, math.Pow(depthT, 0.75))
						slabCols[rowOff+x] = col
						continue
					}

					// Fortification Agate Banding Zone (u in [0.04, 0.54])
					uNorm := (u - 0.04) / 0.50
					phase := uNorm * bands
					bandIdx := int(math.Floor(phase))
					frac := phase - math.Floor(phase)

					cA := bandStops[bandIdx%len(bandStops)]
					cB := bandStops[(bandIdx+1)%len(bandStops)]
					smoothFrac := frac * frac * (3.0 - 2.0*frac)
					col := render.LerpOklab(cA, cB, smoothFrac)

					// Crisp dark/white Liesegang micro-rings at band boundaries
					ringEdge := math.Sin(frac * math.Pi)
					if ringEdge < 0.22 {
						col = render.LerpOklab(col, milkyQuartz, (0.22-ringEdge)*2.4)
					}

					// Analytical gradient normal of the mineral bands for 3D cabochon sheen
					uXp := geodeField(fx+eps, fy)
					uYp := geodeField(fx, fy+eps)
					gx := (uXp - u) / eps
					gy := (uYp - u) / eps
					gLen := math.Hypot(gx, gy) + 1e-6
					nx, ny := gx/gLen, gy/gLen
					lightDot := common.Constrain(0.5+0.5*(-0.65*nx-0.75*ny), 0, 1)
					spec := math.Pow(lightDot, 12.0) * (0.35 + 0.65*ringEdge)
					col = render.LerpOklab(col, diamondGlint, common.Constrain(spec*0.42, 0, 1))

					slabCols[rowOff+x] = col
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	for y := 0; y < h; y++ {
		rowOff := y * w
		for x := 0; x < w; x++ {
			rc.BlendPixel(x, y, slabCols[rowOff+x], 1.0)
		}
	}

	// 2. Inward-Growing 3D Faceted Hexagonal Amethyst Crystal Prisms (内腔立体六棱紫水晶簇)
	rng := rand.New(rand.NewSource(c.Seed() + 909))
	numCrystals := ag.crystalClusters
	if numCrystals < 20 {
		numCrystals = 36
	}

	// Find points along the inner cavity wall (u == 0.52) and grow 3D terminated quartz prisms inward
	for i := 0; i < numCrystals; i++ {
		th := (float64(i)/float64(numCrystals))*2.0*math.Pi + (rng.Float64()-0.5)*0.06
		// Binary search along ray th for u == 0.51
		rLow, rHigh := 0.05*baseR, 0.95*baseR
		for iter := 0; iter < 14; iter++ {
			rMid := 0.5 * (rLow + rHigh)
			px := cx + rMid*math.Cos(th)
			py := cy + rMid*math.Sin(th)
			if geodeField(px, py) > 0.51 {
				rLow = rMid
			} else {
				rHigh = rMid
			}
		}
		rWall := 0.5 * (rLow + rHigh)
		basePt := render.Pt(cx+rWall*math.Cos(th), cy+rWall*math.Sin(th))

		// Point crystal inward toward cavity center with slight natural cluster splay
		inwardAng := th + math.Pi + (rng.Float64()-0.5)*0.32
		dir := render.Pt(math.Cos(inwardAng), math.Sin(inwardAng))
		perp := render.Pt(-dir.Y, dir.X)

		cLen := (26.0 + 24.0*rng.Float64()) * scale
		cHalfW := (7.5 + 5.5*rng.Float64()) * scale
		shoulderDist := cLen * 0.64

		tip := basePt.Add(dir.Mul(cLen))
		shL := basePt.Add(dir.Mul(shoulderDist)).Add(perp.Mul(cHalfW))
		shC := basePt.Add(dir.Mul(shoulderDist * 1.08))
		shR := basePt.Add(dir.Mul(shoulderDist)).Sub(perp.Mul(cHalfW))
		bL := basePt.Add(perp.Mul(cHalfW * 0.92))
		bR := basePt.Sub(perp.Mul(cHalfW * 0.92))

		// Lighting depends on crystal angle relative to upper-left key light
		litVal := 0.5 + 0.5*math.Cos(inwardAng+0.78)
		faceLeftCol := render.LerpOklab(royalViolet, orchidLavender, common.Constrain(0.35+litVal*0.65, 0, 1))
		faceRightCol := render.LerpOklab(druzyAbyss, deepAmethyst, common.Constrain(0.25+(1.0-litVal)*0.65, 0, 1))
		tipLeftCol := render.LerpOklab(orchidLavender, milkyQuartz, common.Constrain(0.45+litVal*0.55, 0, 1))
		tipRightCol := render.LerpOklab(deepAmethyst, royalViolet, 0.68)
		if i%5 == 0 {
			// Occasional golden citrine / rose amethyst crystal twin
			faceLeftCol = render.LerpOklab(faceLeftCol, pyriteGold, 0.45)
			tipLeftCol = render.LerpOklab(tipLeftCol, diamondGlint, 0.55)
		}

		// Left & right prism shaft facets
		rc.FillPolygon([]render.Point{bL, shL, shC, basePt}, faceLeftCol, 0.96)
		rc.FillPolygon([]render.Point{basePt, shC, shR, bR}, faceRightCol, 0.96)
		// Rhombohedral termination facets at the crystal apex
		rc.FillPolygon([]render.Point{shL, tip, shC}, tipLeftCol, 0.98)
		rc.FillPolygon([]render.Point{shC, tip, shR}, tipRightCol, 0.98)

		// Crisp specular ridge lines along crystal edges
		edges := &canvas.Path{}
		edges.MoveTo(bL.X, bL.Y)
		edges.LineTo(shL.X, shL.Y)
		edges.LineTo(tip.X, tip.Y)
		edges.LineTo(shR.X, shR.Y)
		edges.LineTo(bR.X, bR.Y)
		edges.MoveTo(basePt.X, basePt.Y)
		edges.LineTo(shC.X, shC.Y)
		edges.LineTo(tip.X, tip.Y)
		edges.MoveTo(shL.X, shL.Y)
		edges.LineTo(shC.X, shC.Y)
		edges.LineTo(shR.X, shR.Y)
		rc.StrokePath(edges, 0.95*scale, canvas.RoundCap, canvas.RoundJoin, milkyQuartz, 0.72)

		// Sparkling diamond-white star glint at crystal termination apex
		if i%2 == 0 {
			rc.DrawGlowDot(tip.X, tip.Y, 2.4*scale, diamondGlint, 0.95, 0.75)
		}
	}

	// 3. Fine Druzy Micro-Crystal Sparkles inside the Deep Geode Core
	for s := 0; s < 85; s++ {
		r := math.Sqrt(rng.Float64()) * 0.22 * baseR
		a := rng.Float64() * 2.0 * math.Pi
		px := cx + r*math.Cos(a)
		py := cy + r*math.Sin(a)
		col := orchidLavender
		if s%3 == 0 {
			col = pyriteGold
		} else if s%5 == 0 {
			col = diamondGlint
		}
		rc.DrawGlowDot(px, py, (1.2+1.6*rng.Float64())*scale, col, 0.88, 0.65)
	}

	rc.ApplyBloom(0.80, 6, 0.26)
	c.MergeRenderCanvas(rc)
}
