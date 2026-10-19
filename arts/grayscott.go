package arts

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"github.com/jdxyw/generativeart/render"
)

type grayScott struct {
	steps      int
	feed, kill float64
}

// NewGrayScott returns a grayScott engine that simulates a 2D Gray-Scott reaction-diffusion
// morphogenesis system and renders it as a 3D Blinn-Phong lit bioluminescent coral sculpture.
func NewGrayScott(steps int, feed, kill float64) *grayScott {
	if steps <= 0 {
		steps = 3200
	}
	if feed <= 0 {
		feed = 0.037
	}
	if kill <= 0 {
		kill = 0.060
	}
	return &grayScott{
		steps: steps,
		feed:  feed,
		kill:  kill,
	}
}

// cubicHermite evaluates a 1D Catmull-Rom cubic interpolation at fractional offset t in [0, 1].
func cubicHermite(p0, p1, p2, p3, t float64) float64 {
	a := -0.5*p0 + 1.5*p1 - 1.5*p2 + 0.5*p3
	b := p0 - 2.5*p1 + 2.0*p2 - 0.5*p3
	c := -0.5*p0 + 0.5*p2
	d := p1
	return ((a*t+b)*t+c)*t + d
}

// sampleBicubic evaluates a gridSize x gridSize field at floating-point grid coordinates (gx, gy).
func sampleBicubic(field []float64, gridSize int, gx, gy float64) float64 {
	ix := int(math.Floor(gx))
	iy := int(math.Floor(gy))
	tx := gx - float64(ix)
	ty := gy - float64(iy)

	var row [4]float64
	for j := -1; j <= 2; j++ {
		ry := common.ConstrainInt(iy+j, 0, gridSize-1) * gridSize
		x0 := common.ConstrainInt(ix-1, 0, gridSize-1)
		x1 := common.ConstrainInt(ix, 0, gridSize-1)
		x2 := common.ConstrainInt(ix+1, 0, gridSize-1)
		x3 := common.ConstrainInt(ix+2, 0, gridSize-1)
		row[j+1] = cubicHermite(field[ry+x0], field[ry+x1], field[ry+x2], field[ry+x3], tx)
	}
	return common.Constrain(cubicHermite(row[0], row[1], row[2], row[3], ty), 0, 1)
}

func normalizeVec3(x, y, z float64) (float64, float64, float64) {
	l := math.Sqrt(x*x + y*y + z*z)
	if l < 1e-12 {
		return 0, 0, 1
	}
	inv := 1.0 / l
	return x * inv, y * inv, z * inv
}

// Generative simulates the Gray-Scott reaction-diffusion PDE and renders a 3D normal-mapped
// and Blinn-Phong illuminated morphogenesis surface onto c.
func (gs *grayScott) Generative(c *generativeart.Canva) {
	steps := gs.steps
	feed := gs.feed
	kill := gs.kill
	if steps <= 0 {
		steps = 3200
	}
	if feed <= 0 {
		feed = 0.037
	}
	if kill <= 0 {
		kill = 0.060
	}

	w, h := c.Width(), c.Height()
	if w <= 0 || h <= 0 {
		return
	}

	const (
		gridSize = 260
		du       = 0.2097
		dv       = 0.105
		dt       = 1.0
	)
	nCells := gridSize * gridSize
	u0 := make([]float64, nCells)
	v0 := make([]float64, nCells)
	u1 := make([]float64, nCells)
	v1 := make([]float64, nCells)
	feedMap := make([]float64, nCells)
	killMap := make([]float64, nCells)

	for i := range u0 {
		u0[i] = 1.0
	}

	halfGrid := float64(gridSize) * 0.5
	maxGridR := halfGrid * 0.90

	// Spatially vary feed across radius from (feed - 0.004) at center to (feed + 0.006) near rim
	// so the interior forms intricate labyrinthine coral ridges while the outer halo forms mitosis pearls.
	for y := 0; y < gridSize; y++ {
		dy := float64(y) + 0.5 - halfGrid
		for x := 0; x < gridSize; x++ {
			dx := float64(x) + 0.5 - halfGrid
			rNorm := math.Hypot(dx, dy) / maxGridR
			idx := y*gridSize + x
			tRad := common.Clamp01(rNorm)
			feedMap[idx] = (feed - 0.004) + 0.010*tRad
			kLocal := kill
			// Softly increase kill outside the outer rim so the colony stays cleanly framed.
			if rNorm > 0.96 {
				excess := common.Clamp01((rNorm - 0.96) / 0.12)
				kLocal += 0.018 * excess * excess
			}
			killMap[idx] = kLocal
		}
	}

	// Seed initial V concentration along a multi-ring mandala / spiral of 28 circular seeds.
	rng := rand.New(rand.NewSource(rand.Int63()))
	type seedPt struct {
		x, y, rad float64
	}
	seeds := make([]seedPt, 0, 28)
	// Ring 0: 4 tightly coupled core seeds around center
	for i := 0; i < 4; i++ {
		ang := float64(i)*math.Pi*0.5 + 0.2
		r := halfGrid * 0.08
		seeds = append(seeds, seedPt{
			x:   halfGrid + r*math.Cos(ang),
			y:   halfGrid + r*math.Sin(ang),
			rad: 4.2,
		})
	}
	// Ring 1: 8 inner mandala seeds
	for i := 0; i < 8; i++ {
		ang := float64(i)*(2.0*math.Pi/8.0) + 0.35
		r := halfGrid * (0.28 + 0.03*math.Sin(float64(i)*3.0))
		seeds = append(seeds, seedPt{
			x:   halfGrid + r*math.Cos(ang),
			y:   halfGrid + r*math.Sin(ang),
			rad: 4.0,
		})
	}
	// Ring 2: 16 outer spiral/mandala seeds
	for i := 0; i < 16; i++ {
		ang := float64(i)*(2.0*math.Pi/16.0) + 0.18*math.Cos(float64(i)*2.0)
		r := halfGrid * (0.52 + 0.10*math.Sin(float64(i)*2.0+0.6))
		seeds = append(seeds, seedPt{
			x:   halfGrid + r*math.Cos(ang),
			y:   halfGrid + r*math.Sin(ang),
			rad: 3.6,
		})
	}

	for y := 0; y < gridSize; y++ {
		fy := float64(y) + 0.5
		for x := 0; x < gridSize; x++ {
			fx := float64(x) + 0.5
			idx := y*gridSize + x
			for _, s := range seeds {
				d2 := (fx-s.x)*(fx-s.x) + (fy-s.y)*(fy-s.y)
				if d2 <= s.rad*s.rad {
					u0[idx] = 0.50 + (rng.Float64()-0.5)*0.04
					v0[idx] = 0.25 + (rng.Float64()-0.5)*0.04
					break
				}
			}
		}
	}

	// Run Gray-Scott reaction-diffusion steps using 9-point isotropic Laplacian stencil:
	// center: -1.0, orthogonal neighbors: 0.2, diagonal neighbors: 0.05.
	for step := 0; step < steps; step++ {
		// Fast interior update (1 <= y < gridSize-1, 1 <= x < gridSize-1)
		for y := 1; y < gridSize-1; y++ {
			row := y * gridSize
			rowUp := row - gridSize
			rowDn := row + gridSize
			for x := 1; x < gridSize-1; x++ {
				idx := row + x
				uc := u0[idx]
				vc := v0[idx]

				lapU := -uc +
					0.2*(u0[idx-1]+u0[idx+1]+u0[rowUp+x]+u0[rowDn+x]) +
					0.05*(u0[rowUp+x-1]+u0[rowUp+x+1]+u0[rowDn+x-1]+u0[rowDn+x+1])

				lapV := -vc +
					0.2*(v0[idx-1]+v0[idx+1]+v0[rowUp+x]+v0[rowDn+x]) +
					0.05*(v0[rowUp+x-1]+v0[rowUp+x+1]+v0[rowDn+x-1]+v0[rowDn+x+1])

				uvv := uc * vc * vc
				f := feedMap[idx]
				k := killMap[idx]

				un := uc + (du*lapU-uvv+f*(1.0-uc))*dt
				vn := vc + (dv*lapV+uvv-(f+k)*vc)*dt

				if un < 0 {
					un = 0
				} else if un > 1 {
					un = 1
				}
				if vn < 0 {
					vn = 0
				} else if vn > 1 {
					vn = 1
				}
				u1[idx] = un
				v1[idx] = vn
			}
		}

		// Zero-flux / clamped boundary cells
		for i := 0; i < gridSize; i++ {
			u1[i] = 1.0
			v1[i] = 0.0
			u1[(gridSize-1)*gridSize+i] = 1.0
			v1[(gridSize-1)*gridSize+i] = 0.0
			u1[i*gridSize] = 1.0
			v1[i*gridSize] = 0.0
			u1[i*gridSize+(gridSize-1)] = 1.0
			v1[i*gridSize+(gridSize-1)] = 0.0
		}

		u0, u1 = u1, u0
		v0, v1 = v1, v0
	}

	// Render 3D Blinn-Phong lit reaction-diffusion surface to HDR canvas.
	rc := c.NewRenderCanvas()
	schema := c.Opts().ColorSchema()
	if len(schema) == 0 {
		schema = common.Viridis
	}
	grad := render.NewMultiStopGradient(schema...)
	oceanGrad := render.NewMultiStopGradient(common.Oceanic...)

	// Normalize key light L1, rim light L2, and half-vectors with camera V = (0, 0, 1).
	l1x, l1y, l1z := normalizeVec3(-0.6, -0.7, 0.8)
	l2x, l2y, l2z := normalizeVec3(0.7, 0.5, 0.4)
	h1x, h1y, h1z := normalizeVec3(l1x, l1y, l1z+1.0)
	h2x, h2y, h2z := normalizeVec3(l2x, l2y, l2z+1.0)

	// Light colors in linear RGB
	warmKeyR, warmKeyG, warmKeyB := 1.08, 0.96, 0.84
	coolRimR, coolRimG, coolRimB := 0.45, 0.78, 1.10
	specColR, specColG, specColB := 1.15, 1.08, 0.98

	deepAbyss := color.RGBA{R: 7, G: 11, B: 22, A: 255}
	fw, fh := float64(w), float64(h)
	eps := 0.65

	for py := 0; py < h; py++ {
		vNorm := (float64(py) + 0.5) / fh
		gy := vNorm * float64(gridSize-1)
		dyNorm := vNorm*2.0 - 1.0

		for px := 0; px < w; px++ {
			uNorm := (float64(px) + 0.5) / fw
			gx := uNorm * float64(gridSize-1)
			dxNorm := uNorm*2.0 - 1.0
			radNorm := common.Clamp01(math.Hypot(dxNorm, dyNorm))

			vVal := sampleBicubic(v0, gridSize, gx, gy)
			vxPlus := sampleBicubic(v0, gridSize, gx+eps, gy)
			vxMinus := sampleBicubic(v0, gridSize, gx-eps, gy)
			vyPlus := sampleBicubic(v0, gridSize, gx, gy+eps)
			vyMinus := sampleBicubic(v0, gridSize, gx, gy-eps)

			dvdx := (vxPlus - vxMinus) / (2.0 * eps)
			dvdy := (vyPlus - vyMinus) / (2.0 * eps)

			// 3D surface normal N = normalize(-dV/dx * 14.0, -dV/dy * 14.0, 1.0)
			nx, ny, nz := normalizeVec3(-dvdx*14.0, -dvdy*14.0, 1.0)

			// Normalize V concentration (typical Gray-Scott coral ridges peak around V ~ 0.38 - 0.44)
			ridgeHeight := common.Smoothstep(0.03, 0.38, vVal)

			// Local mean curvature / Laplacian for valley ambient occlusion
			lap := (vxPlus + vxMinus + vyPlus + vyMinus - 4.0*vVal) / (eps * eps)
			valleyAO := common.Clamp01(0.32 + 0.68*common.Smoothstep(-0.04, 0.03, ridgeHeight*0.12-lap*0.25))

			// Base albedo from Oklab MultiStopGradient modulated by V concentration and radial distance
			colorT := common.Clamp01(0.12 + 0.72*ridgeHeight + 0.18*radNorm)
			palCol := grad.SampleOklab(colorT)
			rimTint := oceanGrad.SampleOklab(common.Clamp01(0.25 + 0.65*ridgeHeight))
			ridgeCol := render.LerpOklab(palCol, rimTint, 0.22*radNorm)
			albedoCol := render.LerpOklab(deepAbyss, ridgeCol, common.Smoothstep(0.0, 0.75, ridgeHeight*0.88+0.12))

			albR, albG, albB, _ := render.ColorToLinearRGB(albedoCol)

			// Blinn-Phong diffuse & specular terms
			ndotl1 := math.Max(0.0, nx*l1x+ny*l1y+nz*l1z)
			ndotl2 := math.Max(0.0, nx*l2x+ny*l2y+nz*l2z)
			ndoth1 := math.Max(0.0, nx*h1x+ny*h1y+nz*h1z)
			ndoth2 := math.Max(0.0, nx*h2x+ny*h2y+nz*h2z)

			specMask := common.Smoothstep(0.12, 0.35, vVal)
			spec1 := math.Pow(ndoth1, 48.0) * specMask * 0.78
			spec2 := math.Pow(ndoth2, 32.0) * specMask * 0.28

			// Subsurface scattering / bioluminescent glow along ridge crests
			sss := math.Pow(ridgeHeight, 1.6) * 0.22

			amb := 0.18 * valleyAO
			diffR := amb + valleyAO*(0.82*ndotl1*warmKeyR+0.42*ndotl2*coolRimR) + sss
			diffG := amb + valleyAO*(0.82*ndotl1*warmKeyG+0.42*ndotl2*coolRimG) + sss
			diffB := amb + valleyAO*(0.82*ndotl1*warmKeyB+0.42*ndotl2*coolRimB) + sss

			outR := albR*diffR + spec1*specColR + spec2*coolRimR
			outG := albG*diffG + spec1*specColG + spec2*coolRimG
			outB := albB*diffB + spec1*specColB + spec2*coolRimB

			rc.SetPixelLinear(px, py, outR, outG, outB, 1.0)
		}
	}

	rc.ApplyBloom(0.55, 8.0, 0.45)
	rc.ApplyVignette(0.3)
	c.MergeRenderCanvasToneMapped(rc, 1.1, 0)
}
