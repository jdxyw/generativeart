package arts

import (
	"image/color"
	"math"
	"runtime"
	"sync"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type gyroidSurface struct {
	scaleFreq float64
	thickness float64
}

// NewGyroidSurface returns a gyroidSurface engine that raymarches a 3D Schoen
// Gyroid triply periodic minimal surface (TPMS) bounded inside a sculpted sphere
// with two-sided Oklab iridescence, ambient occlusion, and soft shadows.
func NewGyroidSurface(scaleFreq, thickness float64) *gyroidSurface {
	if scaleFreq <= 0 {
		scaleFreq = 4.2
	}
	if thickness <= 0 {
		thickness = 0.085
	}
	return &gyroidSurface{
		scaleFreq: scaleFreq,
		thickness: thickness,
	}
}

// Generative raymarches the 3D Gyroid minimal surface sculpture.
func (gs *gyroidSurface) Generative(c *generativeart.Canva) {
	gs.GenerativeFrame(c, 0, 1, 0.0)
}

// GenerativeFrame raymarches frame frameIdx of a seamless 360-degree phase loop (phase in [0, 1)).
func (gs *gyroidSurface) GenerativeFrame(c *generativeart.Canva, frameIdx, totalFrames int, loopPhase float64) {
	rc := c.NewRenderCanvas()
	w, h := c.Width(), c.Height()
	fw, fh := float64(w), float64(h)
	phaseShift := loopPhase * 2.0 * math.Pi

	abyssBg := color.RGBA{R: 10, G: 14, B: 26, A: 255}
	// Side A (+ gradient): Warm Imperial Gold, Amber & Cinnabar Coral
	goldCore := color.RGBA{R: 246, G: 184, B: 56, A: 255}
	coralRim := color.RGBA{R: 214, G: 68, B: 48, A: 255}
	// Side B (- gradient): Cool Turquoise, Cobalt & Lapis
	tealCore := color.RGBA{R: 48, G: 188, B: 186, A: 255}
	cobaltRim := color.RGBA{R: 26, G: 68, B: 148, A: 255}
	ivorySpec := color.RGBA{R: 255, G: 248, B: 222, A: 255}

	rc.FillBackground(abyssBg)

	freq := gs.scaleFreq
	if freq < 1.5 {
		freq = 4.2
	}
	shellThick := gs.thickness
	if shellThick <= 0.01 {
		shellThick = 0.085
	}

	// Rotate 3D coordinates slightly for an isometric sculptural view of the Gyroid tunnels
	cosY, sinY := math.Cos(0.62), math.Sin(0.62)
	cosX, sinX := math.Cos(0.48), math.Sin(0.48)

	rotPoint := func(p vec3) vec3 {
		x1 := p.x*cosY + p.z*sinY
		z1 := -p.x*sinY + p.z*cosY
		y2 := p.y*cosX - z1*sinX
		z2 := p.y*sinX + z1*cosX
		return vec3{x1, y2, z2}
	}

	// Raw signed gyroid scalar function g(p) in [-1.5, 1.5], advancing smoothly by phaseShift
	evalGyroidRaw := func(p vec3) float64 {
		q := rotPoint(p)
		sx, sy, sz := q.x*freq+phaseShift, q.y*freq+phaseShift, q.z*freq+phaseShift
		return math.Sin(sx)*math.Cos(sy) + math.Sin(sy)*math.Cos(sz) + math.Sin(sz)*math.Cos(sx)
	}

	// 3D Signed Distance Field of the Gyroid shell intersected with a radius-1.22 sphere
	sdfScene := func(p vec3) float64 {
		dSphere := math.Sqrt(p.x*p.x+p.y*p.y+p.z*p.z) - 1.22
		if dSphere > 0.15 {
			return dSphere
		}
		g := evalGyroidRaw(p)
		dGyroid := (math.Abs(g) / freq) - shellThick
		// Smooth intersection with outer bounding sphere
		k := 0.045
		h := common.Constrain(0.5-0.5*(dGyroid-dSphere)/k, 0, 1)
		return dGyroid*(1.0-h) + dSphere*h + k*h*(1.0-h)
	}

	calcNormal := func(p vec3) vec3 {
		const eps = 0.0025
		dx := sdfScene(vec3{p.x + eps, p.y, p.z}) - sdfScene(vec3{p.x - eps, p.y, p.z})
		dy := sdfScene(vec3{p.x, p.y + eps, p.z}) - sdfScene(vec3{p.x, p.y - eps, p.z})
		dz := sdfScene(vec3{p.x, p.y, p.z + eps}) - sdfScene(vec3{p.x, p.y, p.z - eps})
		return (vec3{dx, dy, dz}).norm()
	}

	// 3-step Ambient Occlusion along surface normal
	calcAO := func(p, n vec3) float64 {
		occ := 0.0
		sca := 1.0
		for i := 1; i <= 4; i++ {
			hr := 0.035 * float64(i)
			d := sdfScene(vec3{p.x + n.x*hr, p.y + n.y*hr, p.z + n.z*hr})
			occ += (hr - d) * sca
			sca *= 0.65
		}
		return common.Constrain(1.0-2.6*occ, 0.08, 1.0)
	}

	// Camera at (0, 0, 3.15) looking toward origin (0, 0, 0)
	camPos := vec3{0, 0, 3.15}
	keyLight := (vec3{-0.58, -0.64, 0.50}).norm()
	fillLight := (vec3{0.62, 0.48, 0.62}).norm()
	viewVec := vec3{0, 0, 1}
	halfVec := (vec3{keyLight.x + viewVec.x, keyLight.y + viewVec.y, keyLight.z + viewVec.z}).norm()
	fovScale := 1.42 / math.Min(fw, fh)

	shadeRay := func(sx, sy float64) color.RGBA {
		uvX := (sx - 0.5*fw) * fovScale
		uvY := (sy - 0.5*fh) * fovScale
		rayDir := (vec3{uvX, uvY, -1.55}).norm()

		// Analytical ray-sphere bounding check (r = 1.28) to skip background rays immediately
		b := camPos.dot(rayDir)
		cVal := camPos.dot(camPos) - 1.28*1.28
		disc := b*b - cVal
		if disc < 0 {
			rBg := math.Hypot(uvX, uvY)
			bgT := common.Constrain(1.0-rBg*0.85, 0, 1)
			return render.LerpOklab(abyssBg, cobaltRim, bgT*0.22)
		}

		sqrtD := math.Sqrt(disc)
		tMin := math.Max(0.1, -b-sqrtD)
		tMax := -b + sqrtD

		t := tMin
		hit := false
		for step := 0; step < 68 && t <= tMax; step++ {
			p := vec3{
				camPos.x + rayDir.x*t,
				camPos.y + rayDir.y*t,
				camPos.z + rayDir.z*t,
			}
			d := sdfScene(p)
			if d < 0.0020 {
				hit = true
				break
			}
			t += math.Max(0.0055, d*0.82)
		}

		if !hit {
			rBg := math.Hypot(uvX, uvY)
			bgT := common.Constrain(1.0-rBg*0.85, 0, 1)
			return render.LerpOklab(abyssBg, cobaltRim, bgT*0.22)
		}

		hitPos := vec3{
			camPos.x + rayDir.x*t,
			camPos.y + rayDir.y*t,
			camPos.z + rayDir.z*t,
		}
		n := calcNormal(hitPos)
		ao := calcAO(hitPos, n)
		gSign := evalGyroidRaw(hitPos)

		// Two-sided labyrinth coloring: positive gyroid channel = Gold/Coral, negative = Teal/Cobalt
		var baseCol color.RGBA
		rNorm := common.Constrain(math.Sqrt(hitPos.x*hitPos.x+hitPos.y*hitPos.y+hitPos.z*hitPos.z)/1.22, 0, 1)
		if gSign >= 0 {
			baseCol = render.LerpOklab(goldCore, coralRim, rNorm*0.75)
		} else {
			baseCol = render.LerpOklab(tealCore, cobaltRim, rNorm*0.75)
		}

		ndotl := common.Constrain(n.dot(keyLight), 0, 1)
		ndotFill := common.Constrain(n.dot(fillLight), 0, 1)
		fresnel := math.Pow(common.Constrain(1.0-math.Abs(n.dot(rayDir)), 0, 1), 2.5)
		spec := math.Pow(common.Constrain(n.dot(halfVec), 0, 1), 20.0)

		illum := (0.18 + 0.72*ndotl + 0.24*ndotFill) * ao
		col := render.LerpOklab(abyssBg, baseCol, common.Constrain(illum, 0.06, 1.0))
		if fresnel > 0.1 {
			col = render.LerpOklab(col, goldCore, common.Constrain(fresnel*0.42*ao, 0, 1))
		}
		if spec > 0.05 {
			col = render.LerpOklab(col, ivorySpec, common.Constrain(spec*0.85*ao, 0, 1))
		}
		return col
	}

	// Pass 1: Parallel primary center-ray evaluation
	primCols := make([]color.RGBA, w*h)
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
				rowOff := y * w
				fy := float64(y) + 0.5
				for x := 0; x < w; x++ {
					primCols[rowOff+x] = shadeRay(float64(x)+0.5, fy)
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	// Pass 2: Parallel 2x2 Sub-Pixel Edge Anti-Aliasing (SSAA) on silhouette/ridge discontinuity pixels
	subOffsets := [4][2]float64{{0.25, 0.25}, {0.75, 0.25}, {0.25, 0.75}, {0.75, 0.75}}
	for wIdx := 0; wIdx < workers; wIdx++ {
		y0 := (wIdx * h) / workers
		y1 := ((wIdx + 1) * h) / workers
		wg.Add(1)
		go func(rowStart, rowEnd int) {
			defer wg.Done()
			for y := rowStart; y < rowEnd; y++ {
				rowOff := y * w
				for x := 0; x < w; x++ {
					c0 := primCols[rowOff+x]
					needAA := false
					if x > 0 && x+1 < w && y > 0 && y+1 < h {
						cR := primCols[rowOff+x+1]
						cD := primCols[(y+1)*w+x]
						dr1 := int(c0.R) - int(cR.R)
						dg1 := int(c0.G) - int(cR.G)
						db1 := int(c0.B) - int(cR.B)
						dr2 := int(c0.R) - int(cD.R)
						dg2 := int(c0.G) - int(cD.G)
						db2 := int(c0.B) - int(cD.B)
						if dr1*dr1+dg1*dg1+db1*db1 > 480 || dr2*dr2+dg2*dg2+db2*db2 > 480 {
							needAA = true
						}
					}
					if !needAA {
						rc.SetPixelLinear(x, y, render.SRGB8ToLinear(c0.R), render.SRGB8ToLinear(c0.G), render.SRGB8ToLinear(c0.B), 1.0)
						continue
					}
					var accR, accG, accB float64
					fx, fy := float64(x), float64(y)
					for _, off := range subOffsets {
						sc := shadeRay(fx+off[0], fy+off[1])
						accR += render.SRGB8ToLinear(sc.R)
						accG += render.SRGB8ToLinear(sc.G)
						accB += render.SRGB8ToLinear(sc.B)
					}
					rc.SetPixelLinear(x, y, accR*0.25, accG*0.25, accB*0.25, 1.0)
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	c.MergeRenderCanvas(rc)
}
