package arts

import (
	"image/color"
	"math"
	"runtime"
	"sync"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type bathymetricChart struct {
	isobaths  int
	seamounts int
}

// NewBathymetricChart returns a bathymetricChart engine that renders an oceanographic
// bathymetric relief chart of an abyssal subduction trench, volcanic island arc, and
// coral atoll shelf (深海海沟、火山岛弧与环礁测深海图) with 22 stepped depth shelves,
// 3D underwater cast shadows, Marching Squares isobaths, and a nautical degree frame.
func NewBathymetricChart(isobaths, seamounts int) *bathymetricChart {
	if isobaths <= 0 {
		isobaths = 22
	}
	if seamounts <= 0 {
		seamounts = 7
	}
	return &bathymetricChart{
		isobaths:  isobaths,
		seamounts: seamounts,
	}
}

// Generative renders the oceanographic bathymetric trench and volcanic island arc chart.
func (bc *bathymetricChart) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := c.Width(), c.Height()
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0

	// Oceanographic Bathymetry & Volcanic Island Arc Palette
	trenchAbyss := color.RGBA{R: 5, G: 10, B: 28, A: 255}
	hadalIndigo := color.RGBA{R: 12, G: 28, B: 74, A: 255}
	pelagicCobalt := color.RGBA{R: 22, G: 68, B: 138, A: 255}
	mesopelagicBlue := color.RGBA{R: 34, G: 118, B: 176, A: 255}
	shelfTurquoise := color.RGBA{R: 48, G: 178, B: 186, A: 255}
	lagoonAqua := color.RGBA{R: 122, G: 228, B: 212, A: 255}
	coralSand := color.RGBA{R: 244, G: 224, B: 168, A: 255}
	islandOchre := color.RGBA{R: 228, G: 148, B: 64, A: 255}
	volcanoCoral := color.RGBA{R: 196, G: 62, B: 46, A: 255}
	craterSnow := color.RGBA{R: 252, G: 246, B: 232, A: 255}
	brassRule := color.RGBA{R: 238, G: 192, B: 88, A: 255}

	rc.FillBackground(trenchAbyss)
	rc.SetBlendMode(render.BlendOver)

	ne := common.NewNoiseEngine(c.Seed() + 505)

	// Volcanic island arc centers along a curved tectonic subduction arc
	arcIslands := [][3]float64{
		{-0.38, -0.34, 0.24},
		{-0.08, -0.14, 0.32},
		{0.22, 0.12, 0.28},
		{0.40, 0.38, 0.22},
		{-0.36, 0.28, 0.19},
		{0.34, -0.36, 0.21},
	}

	// Bathymetric elevation field E(x, y) in [0, 1]:
	// 0.0..0.18 = Hadal Subduction Trench, 0.18..0.64 = Continental/Atoll Slope, 0.68..1.0 = Emergent Volcanic Islands
	evalBathymetry := func(x, y float64) float64 {
		u := (x/fw)*2.0 - 1.0
		v := (y/fh)*2.0 - 1.0

		wx := ne.FBM2D(u*2.2+1.4, v*2.2-1.8, 3, 2.0, 0.5) * 0.18
		wy := ne.FBM2D(u*2.2-2.3, v*2.2+1.1, 3, 2.0, 0.5) * 0.18
		pu, pv := u+wx, v+wy

		// Curved Mariana-style subduction trench sweeping across the lower-right side of the island arc
		trenchCurve := pu - 0.85*pv - 0.24*math.Sin(pv*2.4) - 0.18
		trenchCarve := 0.36 * math.Exp(-math.Pow(trenchCurve/0.22, 2.0))

		// Volcanic island arc seamount chain with ring-atoll caldera depressions
		var seamountSum float64
		for _, isl := range arcIslands {
			d := math.Hypot(pu-isl[0], pv-isl[1]) / isl[2]
			cone := math.Exp(-d * d * 1.85)
			// Subtle volcanic crater depression at the very center of the largest island
			crater := 0.14 * math.Exp(-d*d*18.0)
			seamountSum += (cone - crater) * 0.58
		}

		ridged := ne.Ridged2D(pu*3.2+0.7, pv*3.2-1.5, 4, 2.05, 0.52) * 0.28
		fbm := (0.5 + 0.5*ne.FBM2D(pu*4.5, pv*4.5, 3, 2.0, 0.5)) * 0.16

		elev := 0.34 + seamountSum + ridged + fbm - trenchCarve
		return common.Constrain(elev, 0.01, 0.99)
	}

	numLevels := bc.isobaths
	if numLevels < 12 {
		numLevels = 22
	}

	grad := render.NewMultiStopGradient(
		trenchAbyss,
		hadalIndigo,
		pelagicCobalt,
		mesopelagicBlue,
		shelfTurquoise,
		lagoonAqua,
		coralSand,
		islandOchre,
		volcanoCoral,
		craterSnow,
	)

	thresholds := make([]float64, numLevels)
	levelCols := make([]color.RGBA, numLevels)
	for i := 0; i < numLevels; i++ {
		t := float64(i) / float64(numLevels-1)
		thresholds[i] = 0.06 + t*0.88
		levelCols[i] = grad.SampleOklab(t)
	}

	// 1. Precompute Bathymetric Field & Render Stepped Depth Shelves with Underwater Cast Shadows
	fieldGrid := make([]float64, w*h)
	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
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
					fieldGrid[rowOff+x] = evalBathymetry(float64(x)+0.5, fy)
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	shDX := common.MaxInt(1, int(math.Round(4.0*scale)))
	shDY := common.MaxInt(1, int(math.Round(5.0*scale)))

	for wIdx := 0; wIdx < workers; wIdx++ {
		y0 := (wIdx * h) / workers
		y1 := ((wIdx + 1) * h) / workers
		wg.Add(1)
		go func(rowStart, rowEnd int) {
			defer wg.Done()
			for y := rowStart; y < rowEnd; y++ {
				rowOff := y * w
				yP := common.MinInt(h-1, y+1)
				for x := 0; x < w; x++ {
					xP := common.MinInt(w-1, x+1)
					v := fieldGrid[rowOff+x]

					lvl := 0
					for k := numLevels - 1; k >= 0; k-- {
						if v >= thresholds[k] {
							lvl = k
							break
						}
					}
					col := levelCols[lvl]

					// Continuous 3D hillshade modulation
					gx := (fieldGrid[rowOff+xP] - v) * 28.0
					gy := (fieldGrid[yP*w+x] - v) * 28.0
					shade := -0.65*gx - 0.75*gy
					if shade > 0 {
						col = render.LerpOklab(col, craterSnow, common.Constrain(shade*0.32, 0, 0.45))
					} else {
						col = render.LerpOklab(col, trenchAbyss, common.Constrain((-shade)*0.32, 0, 0.45))
					}

					// Stepped terrace drop shadow
					if x >= shDX && y >= shDY {
						vCast := fieldGrid[(y-shDY)*w+(x-shDX)]
						castLvl := 0
						for k := numLevels - 1; k >= 0; k-- {
							if vCast >= thresholds[k] {
								castLvl = k
								break
							}
						}
						if castLvl > lvl {
							col = render.LerpOklab(col, trenchAbyss, 0.42)
						}
					}

					rc.SetPixelLinear(x, y, render.SRGB8ToLinear(col.R), render.SRGB8ToLinear(col.G), render.SRGB8ToLinear(col.B), 1.0)
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	// 2. Extract Sub-Pixel Marching Squares Isobath Contours (Major Index Contours Every 3rd Level)
	gridLookup := func(x, y float64) float64 {
		ix := common.ConstrainInt(int(x), 0, w-1)
		iy := common.ConstrainInt(int(y), 0, h-1)
		return fieldGrid[iy*w+ix]
	}

	for k := 0; k < numLevels; k++ {
		iso := thresholds[k]
		contours := render.ExtractMarchingSquares(115, 115, fw, fh, iso, gridLookup)
		cPath := &canvas.Path{}
		for _, poly := range contours {
			if len(poly) < 3 {
				continue
			}
			smooth := render.CatmullRomSpline(poly, 3)
			cPath.MoveTo(smooth[0].X, smooth[0].Y)
			for i := 1; i < len(smooth); i++ {
				cPath.LineTo(smooth[i].X, smooth[i].Y)
			}
		}

		isCoastline := k == int(float64(numLevels)*0.68)
		isIndexContour := k%3 == 0
		if isCoastline {
			rc.StrokePath(cPath, 2.1*scale, canvas.RoundCap, canvas.RoundJoin, trenchAbyss, 0.85)
			rc.StrokePath(cPath, 1.25*scale, canvas.RoundCap, canvas.RoundJoin, brassRule, 0.96)
		} else if isIndexContour {
			rc.StrokePath(cPath, 1.2*scale, canvas.RoundCap, canvas.RoundJoin, brassRule, 0.78)
		} else {
			rc.StrokePath(cPath, 0.75*scale, canvas.RoundCap, canvas.RoundJoin, craterSnow, 0.42)
		}
	}

	// 3. Nautical Chart Graticule Lines & Graduated Degree Border Frame
	graticule := &canvas.Path{}
	for g := 1; g <= 5; g++ {
		pos := (float64(g) / 6.0) * fw
		graticule.MoveTo(pos, 14*scale)
		graticule.LineTo(pos, fh-14*scale)
		graticule.MoveTo(14*scale, pos)
		graticule.LineTo(fw-14*scale, pos)
	}
	rc.StrokePath(graticule, 0.65*scale, canvas.RoundCap, canvas.RoundJoin, brassRule, 0.26)

	// Outer double-line nautical chart neatline border
	bPad := 12.0 * scale
	border := &canvas.Path{}
	border.MoveTo(bPad, bPad)
	border.LineTo(fw-bPad, bPad)
	border.LineTo(fw-bPad, fh-bPad)
	border.LineTo(bPad, fh-bPad)
	border.Close()
	rc.StrokePath(border, 1.8*scale, canvas.SquareCap, canvas.MiterJoin, brassRule, 0.90)

	c.MergeRenderCanvas(rc)
}
