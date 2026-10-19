package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type boidsFlock struct {
	boidNum     int
	steps       int
	visualRange float64
}

// NewBoidsFlock returns a boidsFlock engine that simulates Craig Reynolds' Boids murmuration
// in a continuous curl-noise and dual-vortex wind field, rendering gossamer neighbor webs,
// tapered calligraphic trajectory ribbons, and crisp origami chevron silhouettes.
func NewBoidsFlock(boidNum, steps int, visualRange float64) *boidsFlock {
	if boidNum <= 0 {
		boidNum = 240
	}
	if steps <= 0 {
		steps = 140
	}
	if visualRange <= 0 {
		visualRange = 52.0
	}
	return &boidsFlock{
		boidNum:     boidNum,
		steps:       steps,
		visualRange: visualRange,
	}
}

type boidSample struct {
	pos       render.Point
	vel       render.Point
	speed     float64
	neighbors int
}

// Generative simulates the Boids murmuration and renders the flock onto the canvas.
func (bf *boidsFlock) Generative(c *generativeart.Canva) {
	boidNum := bf.boidNum
	if boidNum <= 0 {
		boidNum = 240
	}
	steps := bf.steps
	if steps <= 0 {
		steps = 140
	}
	baseVisRange := bf.visualRange
	if baseVisRange <= 0 {
		baseVisRange = 52.0
	}

	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}
	fw, fh := float64(w), float64(h)
	scale := fw / 600.0
	if scale <= 0 {
		scale = 1.0
	}

	visRange := baseVisRange * scale
	sepRange := 14.0 * scale
	minSpeed := 2.2 * scale
	maxSpeed := 5.2 * scale

	seed := rand.Int63()
	rng := rand.New(rand.NewSource(seed))
	noise := common.NewNoiseEngine(seed)

	schema := c.Opts().ColorSchema()
	if len(schema) == 0 {
		schema = common.Sunset
	}
	grad := render.NewMultiStopGradient(schema...)

	// Dual vortex centers in the sky that anchor the sweeping murmuration figure-eight.
	vortex1 := render.Pt(fw*0.35, fh*0.38)
	vortex2 := render.Pt(fw*0.65, fh*0.62)

	// Initialize boidNum boids in 3 balanced sweeping streams around the canvas center.
	streamAnchors := [3]render.Point{
		render.Pt(fw*0.26, fh*0.62),
		render.Pt(fw*0.74, fh*0.38),
		render.Pt(fw*0.52, fh*0.26),
	}
	streamDirs := [3]float64{
		-0.52,
		2.62,
		-2.25,
	}

	pos := make([]render.Point, boidNum)
	vel := make([]render.Point, boidNum)
	group := make([]int, boidNum)
	history := make([][]boidSample, boidNum)

	for i := 0; i < boidNum; i++ {
		g := i % 3
		group[i] = g
		history[i] = make([]boidSample, 0, steps)

		u := float64(i/3) / float64(common.MaxInt(boidNum/3, 1))
		ang := streamDirs[g] + (rng.Float64()-0.5)*0.45
		tan := render.Pt(math.Cos(ang), math.Sin(ang))
		norm := tan.Perp()

		along := (u - 0.5) * 140.0 * scale
		across := rng.NormFloat64() * 26.0 * scale
		p := streamAnchors[g].Add(tan.Mul(along)).Add(norm.Mul(across))
		p.X = common.Constrain(p.X, 45.0*scale, fw-45.0*scale)
		p.Y = common.Constrain(p.Y, 45.0*scale, fh-45.0*scale)

		spd := common.Lerp(minSpeed, maxSpeed, 0.35+0.45*rng.Float64())
		pos[i] = p
		vel[i] = tan.Mul(spd)
	}

	// Simulate bf.steps steps of Craig Reynolds' Boids in a Curl2D + dual-vortex wind field.
	nextPos := make([]render.Point, boidNum)
	nextVel := make([]render.Point, boidNum)
	margin := 82.0 * scale
	noiseScale := 0.0032 / scale

	for step := 0; step < steps; step++ {
		for i := 0; i < boidNum; i++ {
			pi := pos[i]
			vi := vel[i]

			var sepX, sepY float64
			var alignX, alignY float64
			var cohX, cohY float64
			neighborCount := 0

			for j := 0; j < boidNum; j++ {
				if i == j {
					continue
				}
				dx := pos[j].X - pi.X
				dy := pos[j].Y - pi.Y
				dist := math.Hypot(dx, dy)
				if dist <= 1e-6 || dist > visRange {
					continue
				}
				neighborCount++

				// Alignment & Cohesion accumulate over neighbors within visualRange.
				alignX += vel[j].X
				alignY += vel[j].Y
				cohX += pos[j].X
				cohY += pos[j].Y

				// Separation avoids crowding within sepRange (14px scaled).
				if dist < sepRange {
					push := (sepRange - dist) / (sepRange * dist)
					sepX -= dx * push
					sepY -= dy * push
				}
			}

			var steerX, steerY float64
			steerX += sepX * 0.42 * scale
			steerY += sepY * 0.42 * scale

			if neighborCount > 0 {
				invN := 1.0 / float64(neighborCount)
				avgVX := alignX * invN
				avgVY := alignY * invN
				steerX += (avgVX - vi.X) * 0.065
				steerY += (avgVY - vi.Y) * 0.065

				centerX := cohX * invN
				centerY := cohY * invN
				steerX += (centerX - pi.X) * 0.0055
				steerY += (centerY - pi.Y) * 0.0055
			}

			// Continuous Curl2D wind field + dual-vortex orbital swirl.
			curl := noise.Curl2D(pi.X*noiseScale+float64(step)*0.008, pi.Y*noiseScale, 0.02)
			steerX += curl.X * 0.26 * scale
			steerY += curl.Y * 0.26 * scale

			// Dual-vortex swirl attractor so the flock weaves sweeping aerial ribbons.
			targetVortex := vortex1
			swirlSign := 1.0
			if (i+step/45)%2 == 1 {
				targetVortex = vortex2
				swirlSign = -1.0
			}
			vdx := targetVortex.X - pi.X
			vdy := targetVortex.Y - pi.Y
			vDist := math.Hypot(vdx, vdy) + 1e-6
			// Combine radial pull with tangential swirl around the vortex core so streams stay centered.
			ux, uy := vdx/vDist, vdy/vDist
			tx, ty := -uy*swirlSign, ux*swirlSign
			radialWeight := 0.48 + 0.25*common.Clamp01(vDist/(220.0*scale))
			steerX += (ux*radialWeight + tx*0.62) * 0.26 * scale
			steerY += (uy*radialWeight + ty*0.62) * 0.26 * scale

			// Soft boundary containment: steer smoothly away from canvas margins.
			if pi.X < margin {
				t := (margin - pi.X) / margin
				steerX += t * t * 1.35 * scale
			} else if pi.X > fw-margin {
				t := (pi.X - (fw - margin)) / margin
				steerX -= t * t * 1.35 * scale
			}
			if pi.Y < margin {
				t := (margin - pi.Y) / margin
				steerY += t * t * 1.35 * scale
			} else if pi.Y > fh-margin {
				t := (pi.Y - (fh - margin)) / margin
				steerY -= t * t * 1.35 * scale
			}

			nvx := vi.X + steerX
			nvy := vi.Y + steerY
			spd := math.Hypot(nvx, nvy)
			if spd < 1e-6 {
				nvx, nvy = minSpeed, 0
				spd = minSpeed
			} else if spd < minSpeed {
				nvx = nvx / spd * minSpeed
				nvy = nvy / spd * minSpeed
				spd = minSpeed
			} else if spd > maxSpeed {
				nvx = nvx / spd * maxSpeed
				nvy = nvy / spd * maxSpeed
				spd = maxSpeed
			}

			history[i] = append(history[i], boidSample{
				pos:       pi,
				vel:       render.Pt(nvx, nvy),
				speed:     spd,
				neighbors: neighborCount,
			})

			nextVel[i] = render.Pt(nvx, nvy)
			nextPos[i] = render.Pt(pi.X+nvx, pi.Y+nvy)
		}
		copy(pos, nextPos)
		copy(vel, nextVel)
	}

	// Render on rc := c.NewRenderCanvas()
	rc := c.NewRenderCanvas()

	// 1. Atmospheric twilight/nebula backdrop with soft glow at vortex centers.
	zenithCol := color.RGBA{R: 7, G: 10, B: 24, A: 255}
	midTwilight := render.LerpOklab(color.RGBA{R: 16, G: 22, B: 44, A: 255}, grad.SampleOklab(0.25), 0.22)
	horizonWarm := render.LerpOklab(color.RGBA{R: 28, G: 24, B: 48, A: 255}, grad.SampleOklab(0.68), 0.25)
	bgGrad := render.NewMultiStopGradient(zenithCol, midTwilight, horizonWarm)

	for y := 0; y < h; y++ {
		ty := float64(y) / math.Max(1.0, fh-1.0)
		col := bgGrad.SampleOklab(ty)
		rLin, gLin, bLin, _ := render.ColorToLinearRGB(col)
		for x := 0; x < w; x++ {
			rc.SetPixelLinear(x, y, rLin, gLin, bLin, 1.0)
		}
	}

	rc.SetBlendMode(render.BlendAdd)
	rc.DrawGlowDot(vortex1.X, vortex1.Y, 145.0*scale, grad.SampleOklab(0.28), 0.14, 0.0)
	rc.DrawGlowDot(vortex2.X, vortex2.Y, 155.0*scale, grad.SampleOklab(0.76), 0.14, 0.0)
	rc.DrawSoftRing(vortex1.X, vortex1.Y, 68.0*scale, 18.0*scale, grad.SampleOklab(0.35), 0.06)
	rc.DrawSoftRing(vortex2.X, vortex2.Y, 74.0*scale, 18.0*scale, grad.SampleOklab(0.72), 0.06)

	// Fine ambient twilight dust particles in the background.
	for i := 0; i < 160; i++ {
		px := rng.Float64() * fw
		py := rng.Float64() * fh
		dustCol := render.LerpOklab(grad.SampleOklab(rng.Float64()), common.White, 0.45)
		rc.SplatPoint(px, py, dustCol, 0.14+0.24*rng.Float64())
	}

	// 2. First, draw faint gossamer flock-connectivity web lines between close neighbors
	// at selected keyframes in render.BlendAdd mode (using rc.DrawSandLine with low alpha).
	rc.SetBlendMode(render.BlendAdd)
	keyframeStride := common.MaxInt(18, steps/5)
	webThresh := visRange * 0.62
	for kf := keyframeStride; kf < steps; kf += keyframeStride {
		kfFrac := float64(kf) / float64(steps)
		for i := 0; i < boidNum; i++ {
			pi := history[i][kf].pos
			connected := 0
			for j := i + 1; j < boidNum; j++ {
				pj := history[j][kf].pos
				d := pi.Dist(pj)
				if d < webThresh {
					webCol := grad.SampleOklab(common.Clamp01(0.28*float64(group[i]) + 0.5*kfFrac))
					alpha := (1.0 - d/webThresh) * (0.010 + 0.014*kfFrac)
					rc.DrawSandLine(pi, pj, 24, 0.55*scale, webCol, alpha, rng)
					connected++
					if connected >= 3 {
						break
					}
				}
			}
		}
	}

	// 3. Second, render each boid's trajectory as a silky calligraphic ribbon using rc.DrawTaperedSpline
	// (width tapering from 0 at tail to thick at head) colored via Oklab gradient over c.Opts().ColorSchema()
	// based on boid flock group + instantaneous speed!
	rc.SetBlendMode(render.BlendOver)
	stride := common.MaxInt(1, steps/28)
	speedSpan := math.Max(1e-6, maxSpeed-minSpeed)

	for i := 0; i < boidNum; i++ {
		samples := history[i]
		if len(samples) < 4 {
			continue
		}
		pts := make([]render.Point, 0, len(samples)/stride+2)
		spds := make([]float64, 0, len(samples)/stride+2)
		for k := 0; k < len(samples); k += stride {
			pts = append(pts, samples[k].pos)
			spds = append(spds, samples[k].speed)
		}
		last := samples[len(samples)-1]
		if pts[len(pts)-1].Dist(last.pos) > 0.5 {
			pts = append(pts, last.pos)
			spds = append(spds, last.speed)
		}
		if len(pts) < 3 {
			continue
		}

		grpBase := float64(group[i]) * 0.28
		maxRibbonW := (1.35 + 1.25*common.Clamp01(float64(last.neighbors)/10.0)) * scale

		rc.DrawTaperedSpline(
			pts,
			maxRibbonW,
			func(t float64) float64 {
				// Silky calligraphic ribbon tapering from 0 at tail (t=0) to thick near head (t=1).
				t = common.Clamp01(t)
				return math.Pow(t, 1.35) * (0.85 + 0.15*math.Sin(t*math.Pi))
			},
			func(t float64) color.RGBA {
				fIdx := common.Clamp01(t) * float64(len(spds)-1)
				idx0 := int(math.Floor(fIdx))
				idx1 := common.MinInt(idx0+1, len(spds)-1)
				spd := common.Lerp(spds[idx0], spds[idx1], fIdx-float64(idx0))
				spdNorm := common.Clamp01((spd - minSpeed) / speedSpan)

				gradT := common.Clamp01(grpBase + 0.44*spdNorm + 0.20*t)
				col := grad.SampleOklab(gradT)
				if t > 0.78 {
					headLift := common.Smoothstep(0.78, 1.0, t) * 0.38
					col = render.LerpOklab(col, common.White, headLift)
				}
				return col
			},
			0.62,
		)
	}

	// 4. Third, at each boid's final head position, draw a crisp origami chevron / dart silhouette
	// and a luminous head spark (rc.DrawGlowDot).
	sparkWhite := color.RGBA{R: 255, G: 250, B: 238, A: 255}
	for i := 0; i < boidNum; i++ {
		samples := history[i]
		if len(samples) == 0 {
			continue
		}
		head := samples[len(samples)-1]
		dir := head.vel.Norm()
		if dir.Len() < 1e-6 {
			dir = render.Pt(1, 0)
		}
		perp := dir.Perp()

		spdNorm := common.Clamp01((head.speed - minSpeed) / speedSpan)
		baseCol := grad.SampleOklab(common.Clamp01(float64(group[i])*0.28 + 0.55*spdNorm + 0.15))
		dartCol := render.LerpOklab(baseCol, sparkWhite, 0.48)

		tipLen := (5.8 + 2.2*spdNorm) * scale
		wingBack := (4.2 + 1.4*spdNorm) * scale
		wingSpan := (3.4 + 1.1*spdNorm) * scale
		notchBack := 1.6 * scale

		tip := head.pos.Add(dir.Mul(tipLen))
		leftWing := head.pos.Sub(dir.Mul(wingBack)).Add(perp.Mul(wingSpan))
		notch := head.pos.Sub(dir.Mul(notchBack))
		rightWing := head.pos.Sub(dir.Mul(wingBack)).Sub(perp.Mul(wingSpan))

		rc.SetBlendMode(render.BlendOver)
		rc.FillPolygon([]render.Point{tip, leftWing, notch, rightWing}, dartCol, 0.92)

		rc.SetBlendMode(render.BlendAdd)
		rc.DrawGlowDot(tip.X, tip.Y, (1.8+1.2*spdNorm)*scale, render.LerpOklab(dartCol, sparkWhite, 0.55), 0.24, 0.28)
	}

	rc.ApplyBloom(0.66, 7.0, 0.28)
	rc.ApplyVignette(0.28)
	c.MergeRenderCanvasToneMapped(rc, 1.06, 0)
}
