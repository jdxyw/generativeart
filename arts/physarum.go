package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type physarum struct {
	agentNum    int
	steps       int
	sensorAngle float64
	sensorDist  float64
}

type physarumAgent struct {
	x, y    float64
	heading float64
	species int
}

// NewPhysarum returns a physarum engine that simulates a two-species interacting
// slime mold (Physarum polycephalum) transport network with nutrient attractors.
func NewPhysarum(agentNum, steps int, sensorAngle, sensorDist float64) *physarum {
	if agentNum <= 0 {
		agentNum = 45000
	}
	if steps <= 0 {
		steps = 110
	}
	if sensorAngle <= 0 {
		sensorAngle = math.Pi / 4.5
	}
	if sensorDist <= 0 {
		sensorDist = 9.0
	}
	return &physarum{
		agentNum:    agentNum,
		steps:       steps,
		sensorAngle: sensorAngle,
		sensorDist:  sensorDist,
	}
}

// Generative runs the Physarum slime mold simulation and renders its vascular network.
func (ph *physarum) Generative(c *generativeart.Canva) {
	if ph.agentNum <= 0 {
		ph.agentNum = 45000
	}
	if ph.steps <= 0 {
		ph.steps = 110
	}
	if ph.sensorAngle <= 0 {
		ph.sensorAngle = math.Pi / 4.5
	}
	if ph.sensorDist <= 0 {
		ph.sensorDist = 9.0
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

	rng := rand.New(rand.NewSource(rand.Int63()))
	nPix := w * h
	trail0 := make([]float64, nPix)
	trail1 := make([]float64, nPix)
	tmpBuf := make([]float64, nPix)
	nextTrail := make([]float64, nPix)

	cx, cy := fw*0.5, fh*0.5
	maxRadius := minDim * 0.43

	// 6 nutrient attractor nodes arranged on a circle around the center.
	type attractor struct {
		x, y    float64
		species int
	}
	attractors := make([]attractor, 6)
	attractorRingR := minDim * 0.28
	for k := 0; k < 6; k++ {
		ang := float64(k) * (2.0 * math.Pi / 6.0)
		attractors[k] = attractor{
			x:       cx + attractorRingR*math.Cos(ang),
			y:       cy + attractorRingR*math.Sin(ang),
			species: k % 2,
		}
	}

	// Initialize agents across a full-disk mandala + radial filament distribution.
	agents := make([]physarumAgent, ph.agentNum)
	for i := 0; i < ph.agentNum; i++ {
		species := i & 1
		u := float64(i) / float64(ph.agentNum)
		// Mix concentric spiral arms and uniform disk coverage
		var rad, armAngle float64
		if i%3 == 0 {
			rad = math.Sqrt(rng.Float64()) * maxRadius
			armAngle = rng.Float64() * 2.0 * math.Pi
		} else {
			arms := 6.0
			rad = (0.08 + 0.90*math.Sqrt(u)) * maxRadius
			armAngle = float64(i%6)*(2.0*math.Pi/arms) + (rad/maxRadius)*1.85 + rng.NormFloat64()*0.26
		}
		ax := cx + rad*math.Cos(armAngle)
		ay := cy + rad*math.Sin(armAngle)

		inwardAngle := math.Atan2(cy-ay, cx-ax)
		spiralTwist := 0.65
		if species == 1 {
			spiralTwist = -0.65
		}
		heading := inwardAngle + spiralTwist + (rng.Float64()-0.5)*0.85

		agents[i] = physarumAgent{
			x:       common.Constrain(ax, 2, fw-3),
			y:       common.Constrain(ay, 2, fh-3),
			heading: heading,
			species: species,
		}
	}

	sensorDist := ph.sensorDist * scale
	sensorAngle := ph.sensorAngle
	turnSpeed := ph.sensorAngle * 0.78
	stepSize := 1.35 * scale
	decayFactor := 0.92
	diffuseWeight := 0.26

	sampleTrail := func(own, other []float64, sx, sy float64) float64 {
		ix := int(sx)
		iy := int(sy)
		if ix < 0 {
			ix = 0
		} else if ix >= w {
			ix = w - 1
		}
		if iy < 0 {
			iy = 0
		} else if iy >= h {
			iy = h - 1
		}
		idx := iy*w + ix
		return own[idx] - 0.28*other[idx]
	}

	depositBilinear := func(grid []float64, px, py, weight float64) {
		x0 := int(math.Floor(px))
		y0 := int(math.Floor(py))
		if x0 < 0 || x0 >= w-1 || y0 < 0 || y0 >= h-1 {
			return
		}
		fx := px - float64(x0)
		fy := py - float64(y0)
		idx := y0*w + x0
		grid[idx] += (1.0 - fx) * (1.0 - fy) * weight
		grid[idx+1] += fx * (1.0 - fy) * weight
		grid[idx+w] += (1.0 - fx) * fy * weight
		grid[idx+w+1] += fx * fy * weight
	}

	diffuseAndDecay := func(grid []float64) {
		// Horizontal 1D [1, 2, 1] / 4 pass into tmpBuf
		for y := 0; y < h; y++ {
			row := y * w
			tmpBuf[row] = (grid[row]*3.0 + grid[row+1]) * 0.25
			for x := 1; x < w-1; x++ {
				idx := row + x
				tmpBuf[idx] = (grid[idx-1] + 2.0*grid[idx] + grid[idx+1]) * 0.25
			}
			last := row + w - 1
			tmpBuf[last] = (grid[last-1] + grid[last]*3.0) * 0.25
		}
		// Vertical 1D [1, 2, 1] / 4 pass into nextTrail, blended with sharp original trail
		keepWeight := 1.0 - diffuseWeight
		for x := 0; x < w; x++ {
			blur := (tmpBuf[x]*3.0 + tmpBuf[x+w]) * 0.25
			nextTrail[x] = (keepWeight*grid[x] + diffuseWeight*blur) * decayFactor
		}
		for y := 1; y < h-1; y++ {
			row := y * w
			prevRow := row - w
			nextRow := row + w
			for x := 0; x < w; x++ {
				idx := row + x
				blur := (tmpBuf[prevRow+x] + 2.0*tmpBuf[idx] + tmpBuf[nextRow+x]) * 0.25
				nextTrail[idx] = (keepWeight*grid[idx] + diffuseWeight*blur) * decayFactor
			}
		}
		lastRow := (h - 1) * w
		prevLast := lastRow - w
		for x := 0; x < w; x++ {
			idx := lastRow + x
			blur := (tmpBuf[prevLast+x] + tmpBuf[idx]*3.0) * 0.25
			nextTrail[idx] = (keepWeight*grid[idx] + diffuseWeight*blur) * decayFactor
		}
		copy(grid, nextTrail)
	}

	nutrientRadius := int(math.Max(2.0, 5.0*scale))
	for step := 0; step < ph.steps; step++ {
		// Gently emit trail scent at the 6 nutrient attractor nodes each step.
		for _, att := range attractors {
			ax, ay := int(att.x), int(att.y)
			for dy := -nutrientRadius; dy <= nutrientRadius; dy++ {
				py := ay + dy
				if py < 0 || py >= h {
					continue
				}
				for dx := -nutrientRadius; dx <= nutrientRadius; dx++ {
					px := ax + dx
					if px < 0 || px >= w {
						continue
					}
					d2 := float64(dx*dx+dy*dy) / float64(nutrientRadius*nutrientRadius)
					if d2 <= 1.0 {
						scent := 1.35 * (1.0 - d2)
						idx := py*w + px
						if att.species == 0 {
							trail0[idx] += scent
							trail1[idx] += scent * 0.35
						} else {
							trail1[idx] += scent
							trail0[idx] += scent * 0.35
						}
					}
				}
			}
		}

		// Sense, steer, move, and deposit for each agent.
		for i := range agents {
			ag := &agents[i]
			var own, other []float64
			if ag.species == 0 {
				own, other = trail0, trail1
			} else {
				own, other = trail1, trail0
			}

			hFront := ag.heading
			hLeft := ag.heading - sensorAngle
			hRight := ag.heading + sensorAngle

			fVal := sampleTrail(own, other, ag.x+math.Cos(hFront)*sensorDist, ag.y+math.Sin(hFront)*sensorDist)
			lVal := sampleTrail(own, other, ag.x+math.Cos(hLeft)*sensorDist, ag.y+math.Sin(hLeft)*sensorDist)
			rVal := sampleTrail(own, other, ag.x+math.Cos(hRight)*sensorDist, ag.y+math.Sin(hRight)*sensorDist)

			jitter := (rng.Float64() - 0.5) * 0.16
			if fVal > lVal && fVal > rVal {
				ag.heading += jitter * 0.4
			} else if fVal < lVal && fVal < rVal {
				if rng.Float64() < 0.5 {
					ag.heading += turnSpeed + jitter
				} else {
					ag.heading -= turnSpeed + jitter
				}
			} else if lVal > rVal {
				ag.heading -= turnSpeed + jitter
			} else if rVal > lVal {
				ag.heading += turnSpeed + jitter
			} else {
				ag.heading += jitter
			}

			nx := ag.x + math.Cos(ag.heading)*stepSize
			ny := ag.y + math.Sin(ag.heading)*stepSize

			if nx < 2.0 || nx >= fw-2.0 {
				ag.heading = math.Pi - ag.heading
				nx = common.Constrain(nx, 2.0, fw-2.001)
			}
			if ny < 2.0 || ny >= fh-2.0 {
				ag.heading = -ag.heading
				ny = common.Constrain(ny, 2.0, fh-2.001)
			}

			ag.x = nx
			ag.y = ny
			depositBilinear(own, nx, ny, 1.0)
		}

		diffuseAndDecay(trail0)
		diffuseAndDecay(trail1)
	}

	// Determine max log-compressed trail value for normalization.
	maxLog := 0.0
	for i := 0; i < nPix; i++ {
		v := math.Log1p((trail0[i] + trail1[i]) * 6.0)
		if v > maxLog {
			maxLog = v
		}
	}
	if maxLog <= 1e-9 {
		maxLog = 1.0
	}

	schema := c.Opts().ColorSchema()
	if len(schema) == 0 {
		schema = common.Cyberpunk
	}
	grad0 := render.NewMultiStopGradient(schema...)
	grad1 := render.NewMultiStopGradient(common.Viridis...)
	highlightCol := color.RGBA{R: 250, G: 252, B: 235, A: 255}

	rc := c.NewRenderCanvas()
	rc.SetBlendMode(render.BlendAdd)

	for py := 0; py < h; py++ {
		row := py * w
		for px := 0; px < w; px++ {
			idx := row + px
			v0 := math.Log1p(trail0[idx] * 6.0)
			v1 := math.Log1p(trail1[idx] * 6.0)
			vTot := math.Log1p((trail0[idx] + trail1[idx]) * 6.0)
			if vTot <= 1e-4 {
				continue
			}

			normTot := common.Clamp01(vTot / (maxLog * 0.85))
			intensity := math.Pow(normTot, 1.15)
			speciesRatio := v1 / (v0 + v1 + 1e-9)

			col0 := grad0.SampleOklab(common.Clamp01(intensity*0.85 + 0.12))
			col1 := grad1.SampleOklab(common.Clamp01(intensity*0.80 + 0.15))
			col := render.LerpOklab(col0, col1, speciesRatio*0.80)

			if intensity > 0.78 {
				coreT := common.Smoothstep(0.78, 1.0, intensity)
				col = render.LerpOklab(col, highlightCol, coreT*0.55)
			}

			rLin, gLin, bLin, _ := render.ColorToLinearRGB(col)
			hdrBoost := 1.0 + 1.15*math.Pow(intensity, 2.0)
			rc.BlendPixelLinear(px, py, rLin*hdrBoost, gLin*hdrBoost, bLin*hdrBoost, intensity, render.BlendAdd)
		}
	}

	// Subtle luminous nodes at the 6 nutrient attractors.
	for k, att := range attractors {
		col := grad0.SampleOklab(0.65 + 0.05*float64(k))
		rc.DrawGlowDot(att.x, att.y, 8.0*scale, col, 0.22, 0.12)
	}

	rc.ApplyBloom(0.55, 5.0, 0.42)
	rc.ApplyVignette(0.3)
	c.MergeRenderCanvasToneMapped(rc, 1.15, 0)
}
