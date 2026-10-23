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

type waterCaustics struct {
	stonesN int
	waveAmp float64
}

// NewWaterCaustics returns a waterCaustics engine that renders a crystal-clear sunlit
// turquoise pond with submerged agate river pebbles, swimming koi fish, and refractive
// Snell's-law Jacobian optical caustic networks with chromatic dispersion.
func NewWaterCaustics(stonesN int, waveAmp float64) *waterCaustics {
	if stonesN <= 0 {
		stonesN = 28
	}
	if waveAmp <= 0 {
		waveAmp = 1.0
	}
	return &waterCaustics{
		stonesN: stonesN,
		waveAmp: waveAmp,
	}
}

// Generative renders the refractive underwater caustic network over river stones and koi.
func (wc *waterCaustics) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := c.Width(), c.Height()
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0

	// Crystal Pond, Submerged Agate Stones & Chromatic Caustic Palette
	pondAbyss := color.RGBA{R: 8, G: 28, B: 44, A: 255}
	pondTeal := color.RGBA{R: 18, G: 84, B: 102, A: 255}
	pondEmerald := color.RGBA{R: 32, G: 138, B: 136, A: 255}
	causticCyan := color.RGBA{R: 124, G: 242, B: 246, A: 255}
	causticGold := color.RGBA{R: 255, G: 232, B: 148, A: 255}
	causticCoral := color.RGBA{R: 255, G: 146, B: 112, A: 255}
	sunWhite := color.RGBA{R: 255, G: 254, B: 244, A: 255}

	rc.FillBackground(pondAbyss)
	rc.SetBlendMode(render.BlendOver)

	// 1. Sandy/Jade Pond Bed Depth Gradient
	for y := 0; y < h; y++ {
		fy := float64(y) / fh
		for x := 0; x < w; x++ {
			fx := float64(x) / fw
			rCenter := math.Hypot(fx-0.5, fy-0.5)
			t := common.Constrain(1.0-rCenter*1.15, 0, 1)
			col := render.LerpOklab(pondAbyss, pondTeal, t*0.85)
			if t > 0.45 {
				col = render.LerpOklab(col, pondEmerald, (t-0.45)*0.65)
			}
			rc.BlendPixel(x, y, col, 1.0)
		}
	}

	// 2. Submerged Polished River Pebbles & Banded Yuhua Agate Stones (水底五彩雨花玛瑙卵石)
	rng := rand.New(rand.NewSource(c.Seed() + 404))
	stoneColors := [][2]color.RGBA{
		{color.RGBA{R: 168, G: 62, B: 46, A: 255}, color.RGBA{R: 232, G: 164, B: 96, A: 255}},   // Terracotta & Amber Agate
		{color.RGBA{R: 42, G: 112, B: 104, A: 255}, color.RGBA{R: 134, G: 204, B: 182, A: 255}}, // Celadon Jade
		{color.RGBA{R: 196, G: 146, B: 64, A: 255}, color.RGBA{R: 244, G: 222, B: 168, A: 255}}, // Honey Quartz
		{color.RGBA{R: 38, G: 58, B: 88, A: 255}, color.RGBA{R: 98, G: 142, B: 178, A: 255}},    // Slate Lapis
		{color.RGBA{R: 128, G: 54, B: 68, A: 255}, color.RGBA{R: 214, G: 142, B: 148, A: 255}},  // Rose Porphyry
	}

	type pebble struct {
		cx, cy, rx, ry, rot float64
		c0, c1              color.RGBA
	}
	numStones := wc.stonesN
	if numStones < 12 {
		numStones = 28
	}
	pebbles := make([]pebble, 0, numStones)
	for attempt := 0; attempt < numStones*25 && len(pebbles) < numStones; attempt++ {
		px := (0.08 + 0.84*rng.Float64()) * fw
		py := (0.08 + 0.84*rng.Float64()) * fh
		rx := (18.0 + 30.0*rng.Float64()) * scale
		ry := rx * (0.65 + 0.30*rng.Float64())
		overlap := false
		for _, ex := range pebbles {
			if math.Hypot(px-ex.cx, py-ex.cy) < (rx+ex.rx)*0.86 {
				overlap = true
				break
			}
		}
		if overlap {
			continue
		}
		pair := stoneColors[len(pebbles)%len(stoneColors)]
		pebbles = append(pebbles, pebble{
			cx: px, cy: py, rx: rx, ry: ry,
			rot: rng.Float64() * math.Pi,
			c0:  pair[0], c1: pair[1],
		})
	}

	lightDir := render.Pt(-0.62, -0.68)
	for _, pb := range pebbles {
		cosA, sinA := math.Cos(pb.rot), math.Sin(pb.rot)
		buildStonePoly := func(scaleR, offX, offY float64) []render.Point {
			const nPts = 36
			pts := make([]render.Point, nPts)
			for i := 0; i < nPts; i++ {
				a := (float64(i) / float64(nPts)) * 2.0 * math.Pi
				// Subtle organic egg-stone asymmetry
				mod := 1.0 + 0.06*math.Cos(2.0*a+0.5) + 0.03*math.Sin(3.0*a)
				lx := pb.rx * scaleR * mod * math.Cos(a)
				ly := pb.ry * scaleR * mod * math.Sin(a)
				pts[i] = render.Pt(pb.cx+offX+lx*cosA-ly*sinA, pb.cy+offY+lx*sinA+ly*cosA)
			}
			return pts
		}

		// Underwater cast shadow on the pond bed
		rc.FillPolygon(buildStonePoly(1.04, 6.0*scale, 8.0*scale), pondAbyss, 0.58)

		// Concentric banded agate rings inside the pebble
		const bands = 6
		for b := bands; b >= 1; b-- {
			frac := float64(b) / float64(bands)
			shift := (1.0 - frac) * 0.18
			col := render.LerpOklab(pb.c0, pb.c1, 1.0-frac)
			// Blend with underwater teal attenuation
			col = render.LerpOklab(col, pondTeal, 0.28)
			if b%2 == 0 {
				col = render.LerpOklab(col, sunWhite, 0.14)
			}
			rc.FillPolygon(buildStonePoly(frac, -pb.rx*shift*0.4, -pb.ry*shift*0.4), col, 0.92)
		}
	}

	// 3. Two Graceful Submerged Koi Fish (水底双锦鲤游弋)
	drawSubmergedKoi := func(kx, ky, bodyLen, heading, tailCurve float64, primaryCol, spotCol color.RGBA) {
		cosH, sinH := math.Cos(heading), math.Sin(heading)
		spine := make([]render.Point, 24)
		shadowSpine := make([]render.Point, 24)
		for i := 0; i < 24; i++ {
			t := float64(i) / 23.0 // 0 = head, 1 = tail tip
			s := (0.5 - t) * bodyLen
			wiggle := math.Sin(t*math.Pi*2.2+tailCurve) * bodyLen * 0.11 * t * t
			wx := kx + s*cosH - wiggle*sinH
			wy := ky + s*sinH + wiggle*cosH
			spine[i] = render.Pt(wx, wy)
			shadowSpine[i] = render.Pt(wx+8.0*scale, wy+10.0*scale)
		}
		// Underwater koi shadow on the bottom stones
		rc.DrawTaperedSpline(shadowSpine, bodyLen*0.24, render.TaperBothEnds, func(t float64) color.RGBA {
			return pondAbyss
		}, 0.52)

		// Pectoral & tail fins in translucent ivory silk
		headP := spine[4]
		for _, side := range []float64{-1.0, 1.0} {
			finAng := heading + side*2.15
			finPts := []render.Point{
				headP,
				render.Pt(headP.X+bodyLen*0.22*math.Cos(finAng-0.25*side), headP.Y+bodyLen*0.22*math.Sin(finAng-0.25*side)),
				render.Pt(headP.X+bodyLen*0.18*math.Cos(finAng+0.35*side), headP.Y+bodyLen*0.18*math.Sin(finAng+0.35*side)),
			}
			rc.FillPolygon(finPts, sunWhite, 0.55)
		}

		// Sculpted koi body with Kohaku vermilion-and-ivory pattern
		rc.DrawImpastoStroke(spine[:19], bodyLen*0.22, primaryCol, lightDir, 0.92)
		rc.DrawTaperedSpline(spine[3:14], bodyLen*0.14, render.TaperBothEnds, func(t float64) color.RGBA {
			if t > 0.25 && t < 0.72 {
				return spotCol
			}
			return primaryCol
		}, 0.90)
		// Flowing bifurcated tail fin
		tailBase := spine[18]
		tailEnd := spine[23]
		for _, lobe := range []float64{-0.28, 0.28} {
			tPath := []render.Point{
				tailBase,
				tailBase.Lerp(tailEnd, 0.55).Add(render.Pt(-sinH*bodyLen*lobe*0.4, cosH*bodyLen*lobe*0.4)),
				tailEnd.Add(render.Pt(-sinH*bodyLen*lobe, cosH*bodyLen*lobe)),
			}
			rc.DrawTaperedSpline(tPath, bodyLen*0.09, render.TaperBothEnds, func(t float64) color.RGBA {
				return render.LerpOklab(primaryCol, sunWhite, t*0.7)
			}, 0.75)
		}
	}

	drawSubmergedKoi(0.36*fw, 0.42*fh, 138.0*scale, -0.55, 0.4,
		color.RGBA{R: 228, G: 76, B: 42, A: 255}, sunWhite)
	drawSubmergedKoi(0.65*fw, 0.60*fh, 122.0*scale, 2.45, -0.8,
		color.RGBA{R: 246, G: 182, B: 48, A: 255}, color.RGBA{R: 212, G: 52, B: 38, A: 255})

	// 4. Smooth Analytical Water Wave Hessian & Cellular Optical Caustic Network
	ne := common.NewNoiseEngine(c.Seed() + 808)
	// Smooth long-wavelength gravity-capillary wave modes (analytical 2nd derivatives -> zero noise speckle)
	waveModes := [][4]float64{
		{0.048 / scale, 0.022 / scale, 0.6, 1.00},
		{-0.034 / scale, 0.044 / scale, 1.9, 0.92},
		{0.018 / scale, -0.052 / scale, 3.1, 0.88},
		{0.056 / scale, -0.016 / scale, 4.4, 0.78},
		{-0.042 / scale, -0.038 / scale, 5.2, 0.82},
		{0.029 / scale, 0.051 / scale, 2.4, 0.74},
	}
	rippleCenters := [][2]float64{
		{0.36 * fw, 0.42 * fh},
		{0.65 * fw, 0.60 * fh},
	}

	// Evaluate analytical wave deflection gradient (gx, gy) and smooth caustic focusing at (x, y)
	evalCausticIntensity := func(x, y, chromaShift float64) (intensity, glint float64) {
		// Low-frequency domain warp from smooth curl/simplex swell
		wx := ne.Simplex2D(x/(145.0*scale), y/(145.0*scale)) * 22.0 * scale
		wy := ne.Simplex2D((x+300)/(145.0*scale), (y+300)/(145.0*scale)) * 22.0 * scale
		px := x + wx + chromaShift*3.2*scale
		py := y + wy - chromaShift*2.4*scale

		var hxx, hyy, hxy, gx, gy float64
		for _, wm := range waveModes {
			kx, ky, ph, amp := wm[0], wm[1], wm[2], wm[3]*wc.waveAmp
			theta := px*kx + py*ky + ph
			s := math.Sin(theta) * amp
			cVal := math.Cos(theta) * amp
			gx += cVal * kx
			gy += cVal * ky
			hxx -= s * kx * kx
			hyy -= s * ky * ky
			hxy -= s * kx * ky
		}

		// Add smooth Worley cellular ridge network (dual-scale refractive cusp filaments)
		f1A, f2A := ne.Worley2D(px/(96.0*scale), py/(96.0*scale))
		f1B, f2B := ne.Worley2D((px+180)/(68.0*scale)+0.35*gx*scale, (py+180)/(68.0*scale)+0.35*gy*scale)
		edgeA := math.Exp(-math.Abs(f2A-f1A) * 12.5)
		edgeB := math.Exp(-math.Abs(f2B-f1B) * 14.0)

		// Analytical Jacobian convergence factor from smooth wave Hessian
		focal := (220.0 + 28.0*chromaShift) * scale
		j11 := 1.0 - focal*hxx
		j22 := 1.0 - focal*hyy
		j12 := -focal * hxy
		det := math.Abs(j11*j22 - j12*j12)
		hessianFocus := math.Pow(common.Constrain(1.0/(0.55+det*1.15), 0, 1), 2.8)

		// Combine Worley cusp network with analytical Hessian wave focusing
		web := math.Max(edgeA*0.95, edgeB*0.68) * (0.48 + 0.72*hessianFocus) + hessianFocus*0.14
		intensity = math.Pow(common.Constrain(web, 0, 1), 1.45)

		// Specular surface sun-glint on wave crests
		if edgeA > 0.86 && edgeB > 0.76 {
			glint = math.Pow((edgeA-0.86)/0.14, 2.2)
		}
		return intensity, glint
	}

	workers := runtime.NumCPU()
	if workers < 1 {
		workers = 1
	}
	type causticPix struct {
		cR, cG, cB float64
		glint      float64
	}
	causticBuf := make([]causticPix, w*h)

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
					iR, _ := evalCausticIntensity(fx, fy, -1.0)
					iG, gl := evalCausticIntensity(fx, fy, 0.0)
					iB, _ := evalCausticIntensity(fx, fy, 1.0)
					causticBuf[rowOff+x] = causticPix{cR: iR, cG: iG, cB: iB, glint: gl}
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	// Composite chromatic caustics with additive linear-light blending
	rc.SetBlendMode(render.BlendAdd)
	for y := 0; y < h; y++ {
		rowOff := y * w
		for x := 0; x < w; x++ {
			cp := causticBuf[rowOff+x]
			if cp.cR > 0.01 {
				rc.BlendPixel(x, y, causticCoral, cp.cR*0.28)
			}
			if cp.cG > 0.01 {
				rc.BlendPixel(x, y, causticGold, cp.cG*0.36)
			}
			if cp.cB > 0.01 {
				rc.BlendPixel(x, y, causticCyan, cp.cB*0.44)
			}
			if cp.glint > 0.05 {
				rc.BlendPixel(x, y, sunWhite, cp.glint*0.75)
			}
		}
	}

	// 5. Delicate Floating Water-Surface Ripple Rings
	rc.SetBlendMode(render.BlendOver)
	for _, rcPt := range rippleCenters {
		for k := 1; k <= 4; k++ {
			r := float64(k) * 34.0 * scale
			ring := &canvas.Path{}
			for s := 0; s <= 64; s++ {
				a := (float64(s) / 64.0) * 2.0 * math.Pi
				px := rcPt[0] + r*math.Cos(a)
				py := rcPt[1] + r*math.Sin(a)
				if s == 0 {
					ring.MoveTo(px, py)
				} else {
					ring.LineTo(px, py)
				}
			}
			ring.Close()
			rc.StrokePath(ring, (1.5-0.25*float64(k))*scale, canvas.RoundCap, canvas.RoundJoin, sunWhite, 0.28/float64(k))
		}
	}

	rc.ApplyBloom(0.76, 8, 0.32)
	rc.ApplyVignette(0.32)
	c.MergeRenderCanvas(rc)
}
