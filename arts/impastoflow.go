package arts

import (
	"image/color"
	"math"
	"math/rand"
	"sort"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
	"github.com/tdewolff/canvas"
)

type impastoFlow struct {
	furCount     int
	bristleCount int
	noiseScale   float64
}

// NewImpastoFlow returns an impastoFlow engine that renders fine, continuously-graded
// 3D vector bristle plumes emerging smoothly from a luminous combed dark indigo/burgundy fur coat.
func NewImpastoFlow(furCount, bristleCount int, noiseScale float64) *impastoFlow {
	if furCount <= 0 {
		furCount = 3200
	}
	if bristleCount <= 0 {
		bristleCount = 520
	}
	if noiseScale <= 0 {
		noiseScale = 0.0032
	}
	return &impastoFlow{
		furCount:     furCount,
		bristleCount: bristleCount,
		noiseScale:   noiseScale,
	}
}

func isCustomSchema(schema []color.RGBA) bool {
	if len(schema) == 0 {
		return false
	}
	if len(schema) == len(common.Youthful) {
		same := true
		for i := range schema {
			if schema[i] != common.Youthful[i] {
				same = false
				break
			}
		}
		if same {
			return false
		}
	}
	return true
}

func taperFinePlume(t float64) float64 {
	t = common.Constrain(t, 0, 1)
	if t < 0.30 {
		return 0.38 + 0.62*common.Smoothstep(0.0, 0.30, t)
	}
	if t > 0.86 {
		u := (t - 0.86) / 0.14
		return math.Sqrt(math.Max(0, 1.0-u*u))
	}
	return 1.0
}

// Generative renders the Impasto Fur & Feathered Bristle Plumes composition.
func (iflow *impastoFlow) Generative(c *generativeart.Canva) {
	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	seed := rand.Int63()
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	rc := c.NewRenderCanvas()
	obsidianBg := color.RGBA{R: 5, G: 5, B: 13, A: 255}
	rc.FillBackground(obsidianBg)

	// Velvet dark fur palette (with rich indigo & burgundy visibility)
	deepVoid := color.RGBA{R: 5, G: 5, B: 14, A: 255}
	navyDark := color.RGBA{R: 18, G: 22, B: 58, A: 255}
	indigoVelvet := color.RGBA{R: 36, G: 40, B: 96, A: 255}
	violetPlum := color.RGBA{R: 68, G: 32, B: 96, A: 255}
	burgundyTop := color.RGBA{R: 136, G: 34, B: 74, A: 255}
	wineGlow := color.RGBA{R: 182, G: 58, B: 96, A: 255}

	// Impasto brush palette (warm ochre/sienna/gold dominant + crimson & cobalt accents)
	warmUmber := color.RGBA{R: 64, G: 36, B: 34, A: 255}
	rawSienna := color.RGBA{R: 156, G: 84, B: 48, A: 255}
	burntSienna := color.RGBA{R: 204, G: 102, B: 52, A: 255}
	warmTan := color.RGBA{R: 196, G: 148, B: 96, A: 255}
	ochreGold := color.RGBA{R: 236, G: 190, B: 90, A: 255}
	creamGold := color.RGBA{R: 254, G: 240, B: 174, A: 255}
	carmineRed := color.RGBA{R: 214, G: 38, B: 74, A: 255}
	cobaltBlue := color.RGBA{R: 48, G: 120, B: 218, A: 255}
	ceruleanTip := color.RGBA{R: 112, G: 188, B: 246, A: 255}
	whiteGold := color.RGBA{R: 255, G: 250, B: 216, A: 255}

	evalFlowAngle := func(x, y float64) float64 {
		nx := x / fw
		ny := y / fh
		sWave := -0.34 + 0.85*math.Sin(nx*math.Pi*2.0-0.45)*math.Cos((ny-0.48)*math.Pi*1.35)
		upPull := math.Exp(-math.Pow((nx-0.53)/0.22, 2.0)) * common.Smoothstep(0.55, 0.08, ny)
		n1 := noise.FBM2D(nx*1.18+1.4, ny*1.18-0.9, 3, 2.0, 0.45) * 0.85
		return sWave*(1.0-0.55*upPull) + (-math.Pi*0.42)*(0.55*upPull) + n1
	}

	traceSmoothLock := func(x0, y0, length, fanOffset, bend float64, nPts int) []render.Point {
		pts := make([]render.Point, nPts)
		cx, cy := x0, y0
		pts[0] = render.Pt(cx, cy)
		stepD := length / float64(nPts-1)
		for s := 1; s < nPts; s++ {
			t := float64(s) / float64(nPts-1)
			ang := evalFlowAngle(cx, cy) + fanOffset*(0.2+0.8*t) + bend*t*t
			cx += math.Cos(ang) * stepD
			cy += math.Sin(ang) * stepD
			pts[s] = render.Pt(cx, cy)
		}
		return pts
	}

	rc.SetBlendMode(render.BlendOver)

	// 1. Base Atmospheric Velvet Underpainting
	for py := 0; py < h; py++ {
		ny := float64(py) / fh
		for px := 0; px < w; px++ {
			nx := float64(px) / fw
			upper := common.Smoothstep(0.35, 0.02, ny+0.08*math.Sin(nx*math.Pi*2.2))
			nVal := 0.5 + 0.5*noise.Simplex2D(nx*2.0+1.1, ny*2.0-0.7)
			base := render.LerpOklab(deepVoid, navyDark, nVal*0.80)
			base = render.LerpOklab(base, burgundyTop, upper*0.52)
			rL, gL, bL, _ := render.ColorToLinearRGB(base)
			rc.SetPixelLinear(px, py, rL, gL, bL, 1.0)
		}
	}

	// 2. Layer 1: Crisp, Fine Vector-Antialiased Dark Velvet Fur (12,000 continuous paths)
	for i := 0; i < 12000; i++ {
		x0 := (rng.Float64()*1.18 - 0.09) * fw
		y0 := (rng.Float64()*1.18 - 0.09) * fh
		nx := common.Clamp01(x0 / fw)
		ny := common.Clamp01(y0 / fh)

		voidDist := math.Hypot((nx-0.13)*1.1, (ny-0.80)*1.3)
		voidMask := common.Smoothstep(0.06, 0.25, voidDist)

		hairLen := (48.0 + rng.Float64()*46.0) * scale
		fan := (rng.Float64() - 0.5) * 0.08
		bend := (rng.Float64() - 0.5) * 0.12
		pts := traceSmoothLock(x0, y0, hairLen, fan, bend, 9)

		upper := common.Smoothstep(0.32, 0.01, ny+0.09*math.Sin(nx*math.Pi*2.3))
		lockShade := 0.5 + 0.5*noise.FBM2D(nx*3.5+2.1, ny*3.5-1.4, 2, 2.0, 0.5)

		col := render.LerpOklab(navyDark, indigoVelvet, lockShade)
		col = render.LerpOklab(col, violetPlum, lockShade*0.55)
		if upper > 0.03 {
			col = render.LerpOklab(col, render.LerpOklab(burgundyTop, wineGlow, lockShade), upper*0.90)
		}
		col = render.LerpOklab(deepVoid, col, voidMask)
		if i%3 == 0 {
			col = render.LerpOklab(deepVoid, col, 0.38)
		}

		p := &canvas.Path{}
		p.MoveTo(pts[0].X, pts[0].Y)
		for k := 1; k < len(pts); k++ {
			p.LineTo(pts[k].X, pts[k].Y)
		}
		lw := (0.78 + rng.Float64()*0.75) * scale
		rc.StrokePath(p, lw, canvas.RoundCap, canvas.RoundJoin, col, 0.65+0.28*lockShade)
	}

	// 3. Layer 2: Fine, Continuously-Graded Impasto Brush Locks (2,600 fine tapered splines, width 1.25..2.35px)
	evalPlumeField := func(x, y float64) (intensity float64, splaySign float64, lockColorT float64) {
		nx := x / fw
		ny := y / fh

		riverY := 0.58 + 0.12*math.Sin(nx*math.Pi*2.2-0.4) - 0.07*math.Cos(nx*math.Pi*1.4)
		dyRiver := ny - riverY
		wRiver := math.Exp(-(dyRiver * dyRiver) / (2.0 * 0.092 * 0.092))

		dxUp, dyUp := (nx-0.54)*1.3, (ny-0.30)*0.85
		wUpper := 0.90 * math.Exp(-(dxUp*dxUp+dyUp*dyUp)/(2.0*0.10*0.10))

		dxLeft, dyLeft := (nx-0.11)*1.1, (ny-0.40)*0.95
		wLeft := 0.92 * math.Exp(-(dxLeft*dxLeft+dyLeft*dyLeft)/(2.0*0.10*0.10))

		dxRight, dyRight := (nx-0.84)*1.1, (ny-0.52)*0.95
		wRight := 0.85 * math.Exp(-(dxRight*dxRight+dyRight*dyRight)/(2.0*0.10*0.10))

		dxTopR, dyTopR := (nx-0.80)*1.3, (ny-0.22)*1.2
		wTopR := 0.78 * math.Exp(-(dxTopR*dxTopR+dyTopR*dyTopR)/(2.0*0.055*0.055))

		best := math.Max(math.Max(wRiver, wUpper), math.Max(wLeft, math.Max(wRight, wTopR)))
		crossCoord := ny*3.5 - nx*1.6 + 0.35*noise.Simplex2D(nx*2.6, ny*2.6)
		lockWave := 0.34 + 0.66*math.Pow(0.5+0.5*math.Sin(crossCoord*math.Pi*3.4), 1.3)
		intensity = common.Clamp01(best * lockWave * 1.26)

		if dyRiver < 0 {
			splaySign = -1.0
		} else {
			splaySign = 1.0
		}
		lockColorT = 0.5 + 0.5*math.Sin(crossCoord*math.Pi*2.4+nx*3.0)
		return
	}

	// Weighted toward warm ochre, tan, raw sienna, and gold (~68%), with crimson and cobalt accent bands
	impastoGradient := render.NewMultiStopGradient(
		cobaltBlue,
		ceruleanTip,
		warmTan,
		ochreGold,
		creamGold,
		rawSienna,
		burntSienna,
		carmineRed,
		ochreGold,
		warmTan,
	)

	type ribbonJob struct {
		pts       []render.Point
		width     float64
		rootCol   color.RGBA
		midCol    color.RGBA
		tipCol    color.RGBA
		hiCol     color.RGBA
		tipStart  float64
		sortOrder float64
	}
	var ribbons []ribbonJob

	attempts := 0
	for len(ribbons) < 2600 && attempts < 22000 {
		attempts++
		x0 := (rng.Float64()*1.04 - 0.02) * fw
		y0 := (rng.Float64()*1.04 - 0.02) * fh

		inten, splaySign, cT := evalPlumeField(x0, y0)
		if inten < 0.15 || rng.Float64() > inten*1.12 {
			continue
		}

		length := (56.0 + rng.Float64()*52.0) * (0.72 + 0.36*inten) * scale
		fan := splaySign*(0.10+0.24*(1.0-inten)) + (rng.Float64()-0.5)*0.09
		bend := splaySign*(0.14+0.18*inten) + (rng.Float64()-0.5)*0.08
		pts := traceSmoothLock(x0, y0, length, fan, bend, 9)

		tCol := common.Clamp01(cT + (rng.Float64()-0.5)*0.05)
		tipCol := impastoGradient.SampleOklab(tCol)
		if inten > 0.72 {
			tipCol = render.LerpOklab(tipCol, creamGold, (inten-0.72)*1.35)
		}
		rootCol := render.LerpOklab(deepVoid, navyDark, 0.5)
		midCol := render.LerpOklab(warmUmber, tipCol, 0.50)
		hiCol := render.LerpOklab(tipCol, whiteGold, 0.45)

		ribbons = append(ribbons, ribbonJob{
			pts:       pts,
			width:     (1.25 + rng.Float64()*1.10) * (0.85 + 0.22*inten) * scale,
			rootCol:   rootCol,
			midCol:    midCol,
			tipCol:    tipCol,
			hiCol:     hiCol,
			tipStart:  0.44 + 0.18*(1.0-inten),
			sortOrder: inten*2.0 + (x0/fw)*0.5,
		})
	}

	sort.Slice(ribbons, func(i, j int) bool {
		return ribbons[i].sortOrder < ribbons[j].sortOrder
	})

	for _, rb := range ribbons {
		shadowPts := make([]render.Point, len(rb.pts))
		for i, p := range rb.pts {
			shadowPts[i] = render.Pt(p.X+0.7*scale, p.Y+1.1*scale)
		}
		rc.DrawTaperedSpline(
			shadowPts,
			rb.width*1.10,
			taperFinePlume,
			func(float64) color.RGBA { return deepVoid },
			0.55,
		)

		ts := rb.tipStart
		rc.DrawTaperedSpline(
			rb.pts,
			rb.width,
			taperFinePlume,
			func(t float64) color.RGBA {
				if t < ts {
					return render.LerpOklab(rb.rootCol, rb.midCol, common.Smoothstep(0.0, ts, t))
				}
				return render.LerpOklab(rb.midCol, rb.tipCol, common.Smoothstep(ts, 0.92, t))
			},
			0.92,
		)

		if rb.width >= 1.65*scale && len(rb.pts) >= 6 {
			tipPts := rb.pts[4:]
			rc.DrawTaperedSpline(
				tipPts,
				rb.width*0.36,
				render.TaperBothEnds,
				func(t float64) color.RGBA {
					return render.LerpOklab(rb.midCol, rb.hiCol, t)
				},
				0.72,
			)
		}
	}

	rc.ApplyVignette(0.16)
	c.MergeRenderCanvasToneMapped(rc, 1.06, 0)
}
