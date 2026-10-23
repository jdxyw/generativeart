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

type vortexFluid struct {
	vorticesN int
	steps     int
}

// NewVortexFluid returns a vortexFluid engine that simulates Jovian atmospheric
// bands and Kármán vortex streets using backward Lagrangian RK4 advection of
// Rankine vortices, 3D cloud-top relief lighting, and Marching Squares filaments.
func NewVortexFluid(vorticesN, steps int) *vortexFluid {
	if vorticesN <= 0 {
		vorticesN = 9
	}
	if steps <= 0 {
		steps = 18
	}
	return &vortexFluid{
		vorticesN: vorticesN,
		steps:     steps,
	}
}

// Generative renders the Jovian vortex fluid bands and contour filaments.
func (vf *vortexFluid) Generative(c *generativeart.Canva) {
	vf.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame renders frame frameIdx of a seamless 360-degree phase loop (phase in [0, 1)).
func (vf *vortexFluid) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, loopPhase float64) {
	rc := c.NewRenderCanvas()
	w, h := c.Width(), c.Height()
	fw, fh := float64(w), float64(h)
	scale := math.Min(fw, fh) / 600.0
	animPhase := loopPhase * 2.0 * math.Pi

	// Curated NASA Juno Jovian Atmospheric Palette
	deepAbyss := color.RGBA{R: 12, G: 22, B: 46, A: 255}
	prussianBelt := color.RGBA{R: 24, G: 62, B: 118, A: 255}
	cobaltCurrent := color.RGBA{R: 48, G: 118, B: 172, A: 255}
	turquoiseEddy := color.RGBA{R: 88, G: 184, B: 182, A: 255}
	ivoryZone := color.RGBA{R: 248, G: 238, B: 212, A: 255}
	amberCloud := color.RGBA{R: 242, G: 178, B: 68, A: 255}
	ochreStorm := color.RGBA{R: 226, G: 118, B: 46, A: 255}
	greatRedSpot := color.RGBA{R: 192, G: 52, B: 42, A: 255}
	espressoDark := color.RGBA{R: 28, G: 18, B: 24, A: 255}

	rc.FillBackground(deepAbyss)
	rc.SetBlendMode(render.BlendOver)

	type rankineVortex struct {
		x, y   float64
		radius float64
		gamma  float64
	}

	numV := vf.vorticesN
	if numV < 5 {
		numV = 9
	}
	vortices := make([]rankineVortex, 0, numV)
	// Primary Great Red Spot anticyclone at (0.38w, 0.46h)
	vortices = append(vortices, rankineVortex{
		x:      0.38 * fw,
		y:      0.46 * fh,
		radius: 86.0 * scale,
		gamma:  215.0 * scale,
	})
	// Secondary counter-rotating Kármán vortex street eddies
	for i := 1; i < numV; i++ {
		u := float64(i) / float64(numV)
		vx := (0.14 + 0.74*u) * fw
		vy := (0.22 + 0.56*math.Mod(float64(i)*0.37, 1.0)) * fh
		if i == 2 {
			vx, vy = 0.70*fw, 0.32*fh
		} else if i == 3 {
			vx, vy = 0.66*fw, 0.72*fh
		}
		sign := 1.0
		if i%2 == 1 {
			sign = -1.0
		}
		rad := (42.0 + 24.0*math.Sin(float64(i)*1.9)) * scale
		vortices = append(vortices, rankineVortex{
			x:      vx + 3.2*scale*math.Sin(animPhase)*math.Cos(float64(i)*1.1),
			y:      vy + 2.6*scale*(1.0-math.Cos(animPhase))*math.Sin(float64(i)*1.4),
			radius: rad,
			gamma:  sign * (125.0 + 35.0*math.Cos(float64(i)*1.3)) * scale * (1.0 + 0.08*math.Sin(animPhase+float64(i))),
		})
	}

	ne := common.NewNoiseEngine(202617)
	ox := 0.14 * (1.0 - math.Cos(animPhase))
	oy := 0.14 * math.Sin(animPhase)

	// Velocity field combining zonal shear jets + Rankine vortices + curl turbulence
	velocityAt := func(x, y float64) render.Point {
		ny := y / fh
		// Alternating east-west Jovian zonal jet streams
		vx := 28.0 * scale * math.Cos(ny*2.0*math.Pi*3.0)
		vy := 4.5 * scale * math.Sin((x/fw)*2.0*math.Pi*2.0-animPhase)

		for _, v := range vortices {
			dx := x - v.x
			dy := y - v.y
			// Elliptical aspect ratio for Jovian anticyclones
			dyScaled := dy * 1.28
			r2 := dx*dx + dyScaled*dyScaled
			rCore2 := v.radius * v.radius
			// Smooth Rankine-Lamb-Oseen vortex profile
			omega := v.gamma / (r2 + rCore2)
			vx += -dyScaled * omega
			vy += dx * omega * 0.78
		}

		curl := ne.Curl2D(x/(190.0*scale)+ox, y/(190.0*scale)+oy, 1e-3)
		vx += curl.X * 14.0 * scale
		vy += curl.Y * 14.0 * scale
		return render.Pt(vx, vy)
	}

	// Backward Lagrangian RK4 trace from (x, y) to its initial dye coordinate
	advSteps := vf.steps
	if advSteps < 8 {
		advSteps = 16
	}
	if advSteps > 28 {
		advSteps = 28
	}
	dt := 0.22

	traceDyeScalar := func(x, y float64) float64 {
		px, py := x, y
		for s := 0; s < advSteps; s++ {
			v1 := velocityAt(px, py)
			v2 := velocityAt(px-0.5*dt*v1.X, py-0.5*dt*v1.Y)
			v3 := velocityAt(px-0.5*dt*v2.X, py-0.5*dt*v2.Y)
			v4 := velocityAt(px-dt*v3.X, py-dt*v3.Y)
			px -= (dt / 6.0) * (v1.X + 2.0*v2.X + 2.0*v3.X + v4.X)
			py -= (dt / 6.0) * (v1.Y + 2.0*v2.Y + 2.0*v3.Y + v4.Y)
		}
		// Initial stratified latitude dye + gentle Rossby wave breathing
		lat := py / fh
		lon := px / fw
		wave := 0.5 + 0.5*math.Sin(lat*2.0*math.Pi*2.6+0.35*math.Sin(lon*2.0*math.Pi*2.0-0.45* math.Sin(animPhase)))
		// Boost warm core inside primary Great Red Spot
		dRed := math.Hypot((x-vortices[0].x)/(vortices[0].radius*1.15), (y-vortices[0].y)/(vortices[0].radius*0.88))
		if dRed < 1.2 {
			redBoost := math.Exp(-dRed * dRed * 2.2)
			wave = common.Constrain(wave*(1.0-redBoost)+0.94*redBoost, 0, 1)
		}
		return common.Constrain(wave, 0, 1)
	}

	grad := render.NewMultiStopGradient(
		deepAbyss,
		prussianBelt,
		cobaltCurrent,
		turquoiseEddy,
		ivoryZone,
		amberCloud,
		ochreStorm,
		greatRedSpot,
	)

	// 1. Evaluate Advected Dye Field on Full Pixel Grid in Parallel & Compute Cloud-Top 3D Relief Shading
	scalarGrid := make([]float64, w*h)
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
					fx := float64(x) + 0.5
					scalarGrid[rowOff+x] = traceDyeScalar(fx, fy)
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	for y := 0; y < h; y++ {
		rowOff := y * w
		yPrev := y - 1
		if yPrev < 0 {
			yPrev = 0
		}
		yNext := y + 1
		if yNext >= h {
			yNext = h - 1
		}
		for x := 0; x < w; x++ {
			xPrev := x - 1
			if xPrev < 0 {
				xPrev = 0
			}
			xNext := x + 1
			if xNext >= w {
				xNext = w - 1
			}

			val := scalarGrid[rowOff+x]
			dzdx := (scalarGrid[rowOff+xNext] - scalarGrid[rowOff+xPrev]) * 18.0
			dzdy := (scalarGrid[yNext*w+x] - scalarGrid[yPrev*w+x]) * 18.0

			// Quantize slightly into crisp silk-like Jovian cloud bands while keeping smooth gradients
			bandQuant := math.Floor(val*14.0)/14.0*0.45 + val*0.55
			baseCol := grad.SampleOklab(common.Constrain(bandQuant, 0, 1))

			// Directional sunlight from upper-left (-0.65, -0.75)
			shade := (-0.65*dzdx - 0.75*dzdy)
			if shade > 0 {
				baseCol = render.LerpOklab(baseCol, ivoryZone, common.Constrain(shade*0.42, 0, 0.65))
			} else {
				baseCol = render.LerpOklab(baseCol, deepAbyss, common.Constrain((-shade)*0.38, 0, 0.60))
			}
			rc.SetPixelLinear(
				x, y,
				render.SRGB8ToLinear(baseCol.R),
				render.SRGB8ToLinear(baseCol.G),
				render.SRGB8ToLinear(baseCol.B),
				1.0,
			)
		}
	}

	// 2. Extract Marching Squares Cloud Filament Contours & Stroke Crisp Gold/Espresso Streamlines
	gridLookup := func(x, y float64) float64 {
		ix := int(common.Constrain(math.Floor(x), 0, fw-1))
		iy := int(common.Constrain(math.Floor(y), 0, fh-1))
		return scalarGrid[iy*w+ix]
	}

	for lvl := 1; lvl <= 9; lvl++ {
		iso := float64(lvl) / 10.0
		contours := render.ExtractMarchingSquares(100, 100, fw, fh, iso, gridLookup)
		darkPath := &canvas.Path{}
		lightPath := &canvas.Path{}
		for _, poly := range contours {
			if len(poly) < 4 {
				continue
			}
			smooth := render.CatmullRomSpline(poly, 3)
			darkPath.MoveTo(smooth[0].X+0.8*scale, smooth[0].Y+1.0*scale)
			lightPath.MoveTo(smooth[0].X, smooth[0].Y)
			for k := 1; k < len(smooth); k++ {
				darkPath.LineTo(smooth[k].X+0.8*scale, smooth[k].Y+1.0*scale)
				lightPath.LineTo(smooth[k].X, smooth[k].Y)
			}
		}
		rc.StrokePath(darkPath, 1.35*scale, canvas.RoundCap, canvas.RoundJoin, espressoDark, 0.62)
		lineCol := ivoryZone
		if lvl >= 6 {
			lineCol = amberCloud
		}
		rc.StrokePath(lightPath, 0.90*scale, canvas.RoundCap, canvas.RoundJoin, lineCol, 0.78)
	}

	c.MergeRenderCanvas(rc)
}
