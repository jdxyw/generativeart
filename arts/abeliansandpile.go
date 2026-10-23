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

type abelianSandpile struct {
	grains int
}

// NewAbelianSandpile returns an abelianSandpile engine that computes the exact
// Bak-Tang-Wiesenfeld Abelian Sandpile self-organized criticality fractal and
// toppling odometer field with 3D surface normal shading and crystal facets.
func NewAbelianSandpile(grains int) *abelianSandpile {
	if grains <= 0 {
		grains = 260000
	}
	return &abelianSandpile{
		grains: grains,
	}
}

// Generative relaxes the Abelian sandpile to its unique fixed point and renders the fractal crystal.
func (as *abelianSandpile) Generative(c *generativeart.Canva) {
	rc := c.NewRenderCanvas()
	w, h := c.Width(), c.Height()
	fw, fh := float64(w), float64(h)

	// High-Contrast Imperial Mineral Palette for the 4 recurrent sandpile states {0, 1, 2, 3}
	abyssLapis := color.RGBA{R: 8, G: 14, B: 32, A: 255}      // State 0: Deep Obsidian-Lapis
	cobaltBlue := color.RGBA{R: 26, G: 84, B: 172, A: 255}     // State 1: Royal Cobalt-Turquoise
	turquoiseJade := color.RGBA{R: 48, G: 192, B: 184, A: 255} // State 2: Luminous Jingtai Jade
	imperialGold := color.RGBA{R: 248, G: 192, B: 58, A: 255}  // State 3: 24K Imperial Gold
	cinnabarAccent := color.RGBA{R: 218, G: 58, B: 44, A: 255} // High-odometer ridge accent
	ivorySpec := color.RGBA{R: 255, G: 250, B: 230, A: 255}

	rc.FillBackground(abyssLapis)

	// Simulate on an odd-sized square lattice (301 x 301) with active bounding-box acceleration
	const gridN = 301
	const mid = gridN / 2
	grid := make([]int, gridN*gridN)
	nextGrid := make([]int, gridN*gridN)
	odometer := make([]int, gridN*gridN)

	numGrains := as.grains
	if numGrains < 40000 {
		numGrains = 260000
	}
	// Seed a central primary pile plus a 8-fold harmonic ring seed so the sandpile radius fills 92% of the canvas
	grid[mid*gridN+mid] = numGrains
	ringGrains := numGrains / 18
	for k := 0; k < 8; k++ {
		ang := (float64(k) / 8.0) * 2.0 * math.Pi
		rx := mid + int(math.Round(46.0*math.Cos(ang)))
		ry := mid + int(math.Round(46.0*math.Sin(ang)))
		grid[ry*gridN+rx] += ringGrains
	}

	minR, maxR := mid-50, mid+50
	minC, maxC := mid-50, mid+50

	// Fast synchronous batch relaxation: if cell has z >= 4, topple k = z >> 2 grains to each of its 4 neighbors
	for iter := 0; iter < 35000; iter++ {
		changed := false
		copy(nextGrid, grid)
		r0, r1 := minR, maxR
		c0, c1 := minC, maxC
		for r := r0; r <= r1; r++ {
			rowOff := r * gridN
			for col := c0; col <= c1; col++ {
				z := grid[rowOff+col]
				if z >= 4 {
					changed = true
					spill := z >> 2
					nextGrid[rowOff+col] -= spill << 2
					odometer[rowOff+col] += spill
					if r > 1 {
						nextGrid[(r-1)*gridN+col] += spill
						if r-1 < minR {
							minR = r - 1
						}
					}
					if r+1 < gridN-1 {
						nextGrid[(r+1)*gridN+col] += spill
						if r+1 > maxR {
							maxR = r + 1
						}
					}
					if col > 1 {
						nextGrid[rowOff+col-1] += spill
						if col-1 < minC {
							minC = col - 1
						}
					}
					if col+1 < gridN-1 {
						nextGrid[rowOff+col+1] += spill
						if col+1 > maxC {
							maxC = col + 1
						}
					}
				}
			}
		}
		grid, nextGrid = nextGrid, grid
		if !changed {
			break
		}
	}

	maxOdom := float64(common.MaxInt(1, odometer[mid*gridN+mid]))
	activeSpan := float64(common.MaxInt(80, (maxR-minR)+14))
	stateColors := [4]color.RGBA{abyssLapis, cobaltBlue, turquoiseJade, imperialGold}

	// Render at full 600x600 resolution with smooth sub-cell bevels and 3D odometer relief normals
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
			for py := rowStart; py < rowEnd; py++ {
				fy := (float64(py)+0.5)/fh - 0.5
				gy := float64(mid) + fy*activeSpan
				iy := common.ConstrainInt(int(math.Floor(gy)), 1, gridN-2)
				fracY := gy - float64(iy)

				for px := 0; px < w; px++ {
					fx := (float64(px)+0.5)/fw - 0.5
					gx := float64(mid) + fx*activeSpan
					ix := common.ConstrainInt(int(math.Floor(gx)), 1, gridN-2)
					fracX := gx - float64(ix)

					idx := iy*gridN + ix
					st := grid[idx] & 3
					od := float64(odometer[idx])

					if od <= 0 && st == 0 {
						rNorm := math.Hypot(fx, fy) * 1.8
						bg := render.LerpOklab(abyssLapis, cobaltBlue, common.Constrain(1.0-rNorm, 0, 1)*0.18)
						rc.SetPixelLinear(px, py, render.SRGB8ToLinear(bg.R), render.SRGB8ToLinear(bg.G), render.SRGB8ToLinear(bg.B), 1.0)
						continue
					}

					baseCol := stateColors[st]
					// Modulate state 2 and state 3 with odometer potential for fiery cinnabar-gold fractal rings
					odNorm := math.Pow(common.Constrain(od/maxOdom, 0, 1), 0.38)
					if st == 3 {
						ringWave := 0.5 + 0.5*math.Sin(odNorm*math.Pi*18.0)
						baseCol = render.LerpOklab(imperialGold, cinnabarAccent, ringWave*0.52)
					} else if st == 1 {
						baseCol = render.LerpOklab(cobaltBlue, turquoiseJade, odNorm*0.42)
					}

					// Compute local 3D surface normal from discrete height differences + odometer gradient
					zL := float64(grid[idx-1]&3) + 0.002*float64(odometer[idx-1])
					zR := float64(grid[idx+1]&3) + 0.002*float64(odometer[idx+1])
					zU := float64(grid[idx-gridN]&3) + 0.002*float64(odometer[idx-gridN])
					zD := float64(grid[idx+gridN]&3) + 0.002*float64(odometer[idx+gridN])

					dzdx := (zR - zL) * 0.45
					dzdy := (zD - zU) * 0.45
					// Subtle micro-facet bevel inside each sandpile lattice cell
					bevelX := (fracX - 0.5) * 0.35
					bevelY := (fracY - 0.5) * 0.35
					nx := -dzdx - bevelX
					ny := -dzdy - bevelY
					nz := 1.0
					nLen := math.Sqrt(nx*nx + ny*ny + nz*nz)
					nx, ny, nz = nx/nLen, ny/nLen, nz/nLen

					ndotl := common.Constrain(-0.62*nx-0.66*ny+0.42*nz, 0, 1)
					spec := math.Pow(common.Constrain(-0.45*nx-0.48*ny+0.75*nz, 0, 1), 18.0)

					var col color.RGBA
					if ndotl < 0.45 {
						col = render.LerpOklab(abyssLapis, baseCol, 0.35+0.65*(ndotl/0.45))
					} else {
						col = render.LerpOklab(baseCol, ivorySpec, (ndotl-0.45)*0.48)
					}
					if spec > 0.08 && st >= 2 {
						col = render.LerpOklab(col, ivorySpec, common.Constrain(spec*0.75, 0, 1))
					}

					rc.SetPixelLinear(px, py, render.SRGB8ToLinear(col.R), render.SRGB8ToLinear(col.G), render.SRGB8ToLinear(col.B), 1.0)
				}
			}
		}(y0, y1)
	}
	wg.Wait()

	c.MergeRenderCanvas(rc)
}
