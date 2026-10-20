package arts

import (
	"image/color"
	"math"
	"math/rand"
	"sort"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type aizawaAttractor struct {
	steps       int
	dofStrength float64
}

// NewAizawaAttractor returns an aizawaAttractor engine that integrates the 3D Aizawa
// strange attractor via 4th-order Runge-Kutta (RK4) and renders it with thin-lens
// depth-of-field (DoF) stochastic aperture splatting and HDR ACES tone mapping.
func NewAizawaAttractor(steps int, dofStrength float64) *aizawaAttractor {
	if steps <= 0 {
		steps = 900000
	}
	if dofStrength <= 0 {
		dofStrength = 1.0
	}
	return &aizawaAttractor{
		steps:       steps,
		dofStrength: dofStrength,
	}
}

// aizawaDeriv computes the canonical 3D Aizawa strange attractor vector field:
//
//	dx = (z - b)*x - d*y
//	dy = d*x + (z - b)*y
//	dz = c + a*z - (z^3)/3 - (x^2 + y^2)*(1 + e*z) + f*z*x^3
func aizawaDeriv(x, y, z float64) (dx, dy, dz float64) {
	const (
		a = 0.95
		b = 0.7
		c = 0.6
		d = 3.5
		e = 0.25
		f = 0.1
	)
	dx = (z-b)*x - d*y
	dy = d*x + (z-b)*y
	dz = c + a*z - (z*z*z)/3.0 - (x*x+y*y)*(1.0+e*z) + f*z*x*x*x
	return dx, dy, dz
}

// aizawaRK4Step advances (x, y, z) by one 4th-order Runge-Kutta step of size dt.
func aizawaRK4Step(x, y, z, dt float64) (nx, ny, nz, vel float64) {
	k1x, k1y, k1z := aizawaDeriv(x, y, z)
	k2x, k2y, k2z := aizawaDeriv(x+0.5*dt*k1x, y+0.5*dt*k1y, z+0.5*dt*k1z)
	k3x, k3y, k3z := aizawaDeriv(x+0.5*dt*k2x, y+0.5*dt*k2y, z+0.5*dt*k2z)
	k4x, k4y, k4z := aizawaDeriv(x+dt*k3x, y+dt*k3y, z+dt*k3z)

	vx := (k1x + 2.0*k2x + 2.0*k3x + k4x) / 6.0
	vy := (k1y + 2.0*k2y + 2.0*k3y + k4y) / 6.0
	vz := (k1z + 2.0*k2z + 2.0*k3z + k4z) / 6.0

	nx = x + dt*vx
	ny = y + dt*vy
	nz = z + dt*vz
	vel = math.Sqrt(vx*vx + vy*vy + vz*vz)
	return nx, ny, nz, vel
}

// Generative renders the 3D Aizawa Strange Attractor with thin-lens DoF bokeh into c.
func (aa *aizawaAttractor) Generative(c *generativeart.Canva) {
	steps := aa.steps
	dofStrength := aa.dofStrength
	if steps <= 0 {
		steps = 900000
	}
	if dofStrength <= 0 {
		dofStrength = 1.0
	}

	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	minDim := math.Min(fw, fh)
	scale := minDim / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	schema := c.Opts().ColorSchema()
	if len(schema) == 0 {
		schema = common.Cyberpunk
	}
	grad := render.NewMultiStopGradient(schema...)
	plasmaGrad := render.NewMultiStopGradient(common.Plasma...)

	// Precompute a 512-entry linear RGB lookup table from the Oklab palette for fast inner-loop splatting
	const lutSize = 512
	var lutR, lutG, lutB [lutSize]float64
	for i := 0; i < lutSize; i++ {
		t := float64(i) / float64(lutSize-1)
		col := render.LerpOklab(grad.SampleOklab(t), plasmaGrad.SampleOklab(t), 0.25)
		rL, gL, bL, _ := render.ColorToLinearRGB(col)
		lutR[i] = rL
		lutG[i] = gL
		lutB[i] = bL
	}

	nPix := w * h
	rAcc := make([]float64, nPix)
	gAcc := make([]float64, nPix)
	bAcc := make([]float64, nPix)
	density := make([]float64, nPix)

	// Camera orientation: pitch = -0.62 rad, yaw = 0.45 rad so both the equatorial torus shell
	// and the central vertical magnetic spindle are prominently visible.
	const (
		pitch  = -0.62
		yaw    = 0.45
		roll   = 0.14
		dt     = 0.0045
		zCenter = 0.65
		zFocus  = -0.22
	)
	sinYaw, cosYaw := math.Sincos(yaw)
	sinPitch, cosPitch := math.Sincos(pitch)
	sinRoll, cosRoll := math.Sincos(roll)

	cx := fw * 0.5
	cy := fh * 0.51
	projScale := minDim * 0.295

	rng := rand.New(rand.NewSource(202609))

	type bokehCandidate struct {
		px, py float64
		coc    float64
		deltaZ float64
		col    color.RGBA
	}
	var bokehCandidates []bokehCandidate

	// Warm up trajectory onto the Aizawa attractor invariant manifold
	x, y, z := 0.1, 0.0, 0.0
	for i := 0; i < 1500; i++ {
		x, y, z, _ = aizawaRK4Step(x, y, z, dt)
	}

	for i := 0; i < steps; i++ {
		nx, ny, nz, vel := aizawaRK4Step(x, y, z, dt)
		if math.IsNaN(nx) || math.IsNaN(ny) || math.IsNaN(nz) || nx*nx+ny*ny+nz*nz > 25.0 {
			// Re-seed slightly perturbed orbit on the attractor if numerical escape ever occurs
			ang := float64(i) * 0.61803398875 * 2.0 * math.Pi
			x, y, z = 0.12*math.Cos(ang), 0.12*math.Sin(ang), 0.1
			continue
		}
		x, y, z = nx, ny, nz

		// 1. Rotate 3D coordinates (x, y, z - zCenter) by yaw, pitch, and subtle roll
		zc := z - zCenter
		x1 := x*cosYaw - y*sinYaw
		y1 := x*sinYaw + y*cosYaw

		// Tilt by pitch so Z spindle projects vertically and equatorial torus tilts into 3D view
		yCam := -zc*cosPitch + y1*sinPitch
		zCam := zc*sinPitch + y1*cosPitch
		xCam := x1

		// Subtle camera roll for dynamic composition
		sx := xCam*cosRoll - yCam*sinRoll
		sy := xCam*sinRoll + yCam*cosRoll

		// Mild perspective foreshortening
		persp := 1.0 / (1.0 + 0.14*zCam)
		basePx := cx + sx*projScale*persp
		basePy := cy + sy*projScale*persp

		// 2. Thin-lens Circle-of-Confusion (CoC) from camera depth relative to focal plane
		deltaZ := math.Abs(zCam - zFocus)
		coc := dofStrength * math.Pow(deltaZ, 1.35) * 8.5 * scale

		// Stochastic thin-lens aperture sampling (slightly ring-weighted for authentic optical bokeh)
		uRand := rng.Float64()
		rad := coc * math.Pow(uRand, 0.42)
		theta := rng.Float64() * 2.0 * math.Pi
		px := basePx + rad*math.Cos(theta)
		py := basePy + rad*math.Sin(theta)

		x0 := int(math.Floor(px))
		y0 := int(math.Floor(py))
		if x0 < 0 || x0 >= w-1 || y0 < 0 || y0 >= h-1 {
			continue
		}

		// 3. Color by 3D altitude z + velocity + radial spindle proximity in Oklab space
		zNorm := common.Clamp01((z + 0.55) / 2.45)
		velNorm := common.Clamp01((vel - 0.5) / 5.2)
		rEq := math.Hypot(x, y)
		spindleT := common.Clamp01(1.0 - rEq/0.45)

		tCol := common.Clamp01(0.54*zNorm + 0.30*velNorm + 0.16*spindleT)
		lutIdx := int(tCol * float64(lutSize-1))
		if lutIdx < 0 {
			lutIdx = 0
		} else if lutIdx >= lutSize {
			lutIdx = lutSize - 1
		}
		cr, cg, cb := lutR[lutIdx], lutG[lutIdx], lutB[lutIdx]

		// Energy conservation weight: sharper focus steps deposit higher peak density
		focusWeight := 1.0 / (1.0 + 0.18*(coc/scale))

		// Bilinear sub-pixel splatting into float64 HDR accumulation buffers
		fx := px - float64(x0)
		fy := py - float64(y0)
		w00 := (1.0 - fx) * (1.0 - fy) * focusWeight
		w10 := fx * (1.0 - fy) * focusWeight
		w01 := (1.0 - fx) * fy * focusWeight
		w11 := fx * fy * focusWeight

		idx00 := y0*w + x0
		idx10 := idx00 + 1
		idx01 := idx00 + w
		idx11 := idx01 + 1

		density[idx00] += w00
		density[idx10] += w10
		density[idx01] += w01
		density[idx11] += w11

		rAcc[idx00] += w00 * cr
		gAcc[idx00] += w00 * cg
		bAcc[idx00] += w00 * cb

		rAcc[idx10] += w10 * cr
		gAcc[idx10] += w10 * cg
		bAcc[idx10] += w10 * cb

		rAcc[idx01] += w01 * cr
		gAcc[idx01] += w01 * cg
		bAcc[idx01] += w01 * cb

		rAcc[idx11] += w11 * cr
		gAcc[idx11] += w11 * cg
		bAcc[idx11] += w11 * cb

		// Sample out-of-focus foreground trajectory points as candidates for optical bokeh rings
		if i%900 == 0 && deltaZ > 0.65 && basePx > 40*scale && basePx < fw-40*scale && basePy > 40*scale && basePy < fh-40*scale {
			bokehCandidates = append(bokehCandidates, bokehCandidate{
				px:     basePx,
				py:     basePy,
				coc:    coc,
				deltaZ: deltaZ,
				col:    grad.SampleOklab(tCol),
			})
		}
	}

	// 4. Compute 99.8th-percentile reference density for logarithmic HDR tone mapping
	nonZero := make([]float64, 0, nPix/4)
	maxDensity := 0.0
	for _, d := range density {
		if d > 0 {
			nonZero = append(nonZero, d)
			if d > maxDensity {
				maxDensity = d
			}
		}
	}
	if len(nonZero) > 100 {
		sort.Float64s(nonZero)
		pIdx := int(float64(len(nonZero)-1) * 0.998)
		refD := nonZero[pIdx]
		if refD > 0 {
			maxDensity = refD
		}
	}
	if maxDensity <= 0 {
		maxDensity = 1.0
	}

	rc := c.NewRenderCanvas()

	// Deep cosmic void backdrop with subtle ambient glow
	bgCol := color.RGBA{R: 6, G: 7, B: 16, A: 255}
	rc.FillBackground(bgCol)
	rc.SetBlendMode(render.BlendAdd)
	rc.DrawGlowDot(cx, cy, minDim*0.45, grad.SampleOklab(0.20), 0.10, 0.0)

	coreWhiteGold := color.RGBA{R: 255, G: 248, B: 228, A: 255}
	cwR, cwG, cwB, _ := render.ColorToLinearRGB(coreWhiteGold)

	logDenom := math.Log(1.0 + maxDensity*22.0)
	if logDenom <= 1e-9 {
		logDenom = 1.0
	}

	for py := 0; py < h; py++ {
		rowOff := py * w
		for px := 0; px < w; px++ {
			idx := rowOff + px
			d := density[idx]
			if d <= 0 {
				continue
			}

			dNorm := common.Clamp01(math.Log(1.0+d*22.0) / logDenom)
			mapped := math.Pow(dNorm, 0.74)

			invD := 1.0 / d
			rLin := rAcc[idx] * invD
			gLin := gAcc[idx] * invD
			bLin := bAcc[idx] * invD

			// Shift ultra-dense focal filaments toward incandescent white-gold
			if mapped > 0.65 {
				coreT := common.Smoothstep(0.65, 1.0, mapped) * 0.72
				rLin = rLin*(1.0-coreT) + cwR*coreT
				gLin = gLin*(1.0-coreT) + cwG*coreT
				bLin = bLin*(1.0-coreT) + cwB*coreT
			}

			hdrBoost := 1.0 + 1.65*math.Pow(mapped, 2.1)
			rc.BlendPixelLinear(px, py, rLin*hdrBoost, gLin*hdrBoost, bLin*hdrBoost, mapped, render.BlendAdd)
		}
	}

	// 5. Render 25 soft optical bokeh rings/discs at well-separated out-of-focus foreground trajectory points
	sort.Slice(bokehCandidates, func(i, j int) bool {
		return bokehCandidates[i].deltaZ > bokehCandidates[j].deltaZ
	})

	selectedBokeh := make([]bokehCandidate, 0, 25)
	minSep := 26.0 * scale
	for _, cand := range bokehCandidates {
		if len(selectedBokeh) >= 25 {
			break
		}
		tooClose := false
		for _, sel := range selectedBokeh {
			if math.Hypot(cand.px-sel.px, cand.py-sel.py) < minSep {
				tooClose = true
				break
			}
		}
		if !tooClose {
			selectedBokeh = append(selectedBokeh, cand)
		}
	}

	rc.SetBlendMode(render.BlendAdd)
	for i, bk := range selectedBokeh {
		u := 1.0 - float64(i)/float64(common.MaxInt(len(selectedBokeh), 1))
		rad := common.Constrain(bk.coc*(0.85+0.45*u), 6.0*scale, 26.0*scale)
		rimCol := render.LerpOklab(bk.col, coreWhiteGold, 0.38)
		// Soft interior aperture disc + luminous spherical-aberration bokeh ring
		rc.DrawGlowDot(bk.px, bk.py, rad*0.95, bk.col, 0.05+0.05*u, 0.72)
		rc.DrawSoftRing(bk.px, bk.py, rad, math.Max(1.2*scale, rad*0.14), rimCol, 0.14+0.12*u)
	}

	rc.ApplyBloom(0.42, 9.0, 0.65)
	rc.ApplyVignette(0.3)
	c.MergeRenderCanvasToneMapped(rc, 1.15, 0)
}
