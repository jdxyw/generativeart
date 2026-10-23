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

type soapIridescence struct {
	cells  int
	filmNm float64
}

// NewSoapIridescence returns a soapIridescence engine that renders a Plateau's-law
// soap foam lattice with physical thin-film optical interference (纳米薄膜干涉牛顿环与皂膜晶格),
// Marangoni涡流 swirls, and 3D spherical meniscus highlights.
func NewSoapIridescence(cells int, filmNm float64) *soapIridescence {
	if cells <= 0 {
		cells = 42
	}
	if filmNm <= 0 {
		filmNm = 480.0
	}
	return &soapIridescence{
		cells:  cells,
		filmNm: filmNm,
	}
}

// Generative renders the thin-film interference soap bubble foam lattice.
func (si *soapIridescence) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := c.Width(), c.Height()
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0

	abyssBg := color.RGBA{R: 6, G: 8, B: 16, A: 255}
	liquidCyan := color.RGBA{R: 98, G: 238, B: 255, A: 255}
	liquidGold := color.RGBA{R: 255, G: 218, B: 112, A: 255}
	liquidMagenta := color.RGBA{R: 255, G: 96, B: 198, A: 255}
	specWhite := color.RGBA{R: 255, G: 254, B: 248, A: 255}

	rc.FillBackground(abyssBg)
	rc.SetBlendMode(render.BlendOver)

	rng := rand.New(rand.NewSource(c.Seed() + 717))
	ne := common.NewNoiseEngine(c.Seed() + 818)

	numCells := si.cells
	if numCells < 16 {
		numCells = 42
	}

	type bubbleSite struct {
		x, y      float64
		weight    float64
		baseThick float64
		swirlFreq float64
		phaseOff  float64
	}

	// Generate uniformly distributed bubble sites across an overscanned grid [-0.08..1.08]
	// so every foam cell on the canvas is cleanly bounded on all sides
	gridSide := int(math.Ceil(math.Sqrt(float64(numCells))))
	if gridSide < 5 {
		gridSide = 6
	}
	sites := make([]bubbleSite, 0, gridSide*gridSide)
	spanW := fw * 1.16
	spanH := fh * 1.16
	cellW := spanW / float64(gridSide)
	cellH := spanH / float64(gridSide)
	for gy := 0; gy < gridSide; gy++ {
		for gx := 0; gx < gridSide; gx++ {
			sx := -0.08*fw + (float64(gx)+0.22+0.56*rng.Float64())*cellW
			sy := -0.08*fh + (float64(gy)+0.22+0.56*rng.Float64())*cellH
			sites = append(sites, bubbleSite{
				x:         sx,
				y:         sy,
				weight:    (38.0 + 18.0*rng.Float64()) * scale,
				baseThick: si.filmNm * (0.58 + 0.78*rng.Float64()),
				swirlFreq: 2.5 + 3.0*rng.Float64(),
				phaseOff:  rng.Float64() * 2.0 * math.Pi,
			})
		}
	}

	// Physical Thin-Film Interference Reflectance R(lambda) for soap film (n_film = 1.35)
	// Phase difference: delta = (4 * pi * n_film * d * cos(theta_t)) / lambda + pi
	thinFilmColor := func(dNm, cosTheta float64) (rVal, gVal, bVal float64) {
		const nFilm = 1.35
		// Snell's law: sin^2(theta_t) = (1 - cos^2(theta_i)) / nFilm^2
		sin2I := math.Max(0, 1.0-cosTheta*cosTheta)
		cosT := math.Sqrt(math.Max(0.1, 1.0-sin2I/(nFilm*nFilm)))
		opd := 2.0 * nFilm * dNm * cosT // Optical Path Difference in nanometers

		// Evaluate multi-wavelength interference across R (630nm), G (532nm), B (465nm)
		evalLambda := func(lambdaA, lambdaB float64) float64 {
			s1 := math.Sin(math.Pi * opd / lambdaA)
			s2 := math.Sin(math.Pi * opd / lambdaB)
			return 0.5 * (s1*s1 + s2*s2)
		}

		// Below 50nm -> Newton black film attenuation
		blackEnv := common.Constrain((dNm-35.0)/55.0, 0.05, 1.0)
		rVal = evalLambda(615.0, 650.0) * blackEnv
		gVal = evalLambda(520.0, 550.0) * blackEnv
		bVal = evalLambda(445.0, 475.0) * blackEnv
		return
	}

	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
	pixCols := make([]color.RGBA, w*h)

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

					// Find 1st and 2nd nearest weighted bubble sites (Power / Multiplicative Voronoi)
					best1, best2 := 1e18, 1e18
					idx1 := 0
					for i, s := range sites {
						dx := fx - s.x
						dy := fy - s.y
						d := math.Hypot(dx, dy) / s.weight
						if d < best1 {
							best2 = best1
							best1 = d
							idx1 = i
						} else if d < best2 {
							best2 = d
						}
					}

					s := sites[idx1]
					// Normalized cell coordinate uCell in [0, 1] (0 at bubble center, 1 at exact Plateau border)
					uCell := common.Constrain(best1/math.Max(1e-6, best2), 0, 1)
					borderGap := (best2 - best1) * s.weight
					vecX := fx - s.x
					vecY := fy - s.y
					vLen := math.Hypot(vecX, vecY) + 1e-6
					dirX, dirY := vecX/vLen, vecY/vLen

					// 3D Spherical Cap Dome Normal parameterized smoothly by uCell
					rDome := uCell * 0.92
					nz := math.Sqrt(math.Max(0.08, 1.0-rDome*rDome))
					nx, ny := dirX*rDome, dirY*rDome

					// Marangoni convection vortices + gravitational drainage inside the soap dome
					ang := math.Atan2(dirY, dirX)
					curl := ne.FBM2D(fx/(46.0*scale)+float64(idx1)*1.3, fy/(46.0*scale), 3, 2.1, 0.55)
					swirl := math.Sin(uCell*7.5 - ang*2.0 + s.phaseOff + curl*3.6)
					// Film drains thinner near the top of each bubble dome and thickens near the Plateau border
					meniscusBoost := math.Exp(-borderGap / (10.5 * scale)) * 185.0
					dNm := s.baseThick*(0.62+0.32*(0.5+0.5*dirY*uCell)+0.28*swirl+0.24*curl) + meniscusBoost

					irR, irG, irB := thinFilmColor(dNm, nz)

					// Vivid thin-film iridescence across the entire dome + strong grazing Fresnel boost
					fresnel := 0.54 + 0.46*math.Pow(uCell, 1.45)
					// Studio softbox window highlight in upper-left + secondary fill in lower-right
					spec1 := math.Pow(common.Constrain(-0.55*nx-0.60*ny+0.58*nz, 0, 1), 28.0) * (0.35 + 0.65*uCell)
					spec2 := math.Pow(common.Constrain(0.52*nx+0.48*ny+0.70*nz, 0, 1), 18.0) * 0.32

					baseR := uint8(common.Constrain((0.05+irR*fresnel*0.92+spec1*0.88+spec2*0.38)*255.0, 0, 255))
					baseG := uint8(common.Constrain((0.06+irG*fresnel*0.92+spec1*0.88+spec2*0.42)*255.0, 0, 255))
					baseB := uint8(common.Constrain((0.10+irB*fresnel*0.95+spec1*0.92+spec2*0.48)*255.0, 0, 255))
					col := color.RGBA{R: baseR, G: baseG, B: baseB, A: 255}

					// Plateau Border Liquid Channel (120-degree liquid meniscus bridge between cells)
					if borderGap < 4.2*scale {
						tBorder := 1.0 - (borderGap / (4.2 * scale))
						// Dark refractive trough with blazing central liquid specular thread
						col = render.LerpOklab(col, abyssBg, math.Sin(tBorder*math.Pi)*0.65)
						if borderGap < 1.35*scale {
							coreT := 1.0 - borderGap/(1.35*scale)
							rimTint := liquidCyan
							if idx1%3 == 1 {
								rimTint = liquidGold
							} else if idx1%3 == 2 {
								rimTint = liquidMagenta
							}
							col = render.LerpOklab(col, rimTint, coreT*0.75)
							col = render.LerpOklab(col, specWhite, math.Pow(coreT, 2.2)*0.85)
						}
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

	// Crisp curved specular arc reflections inside each bubble dome
	for i, s := range sites {
		arcR := s.weight * 0.56
		if s.x-arcR < 10 || s.x+arcR > fw-10 || s.y-arcR < 10 || s.y+arcR > fh-10 {
			continue
		}
		arcPath := &canvas.Path{}
		for k := 0; k <= 16; k++ {
			a := (-145.0 + float64(k)*3.2) * math.Pi / 180.0
			px := s.x + arcR*math.Cos(a)
			py := s.y + arcR*math.Sin(a)
			if k == 0 {
				arcPath.MoveTo(px, py)
			} else {
				arcPath.LineTo(px, py)
			}
		}
		tint := specWhite
		if i%2 == 1 {
			tint = render.LerpOklab(specWhite, liquidGold, 0.25)
		}
		rc.StrokePath(arcPath, 1.65*scale, canvas.RoundCap, canvas.RoundJoin, tint, 0.62)
	}

	rc.ApplyBloom(0.82, 7, 0.28)
	c.MergeRenderCanvas(rc)
}
