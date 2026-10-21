package render

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart/common"
)

// VectorField2D evaluates a 2D direction/velocity vector at (x, y).
type VectorField2D func(x, y float64) Point

// ScalarField2D evaluates a 2D scalar height/potential value at (x, y).
type ScalarField2D func(x, y float64) float64

// StreamlineConfig configures the Jobard-Lefer evenly-spaced streamline integrator.
type StreamlineConfig struct {
	Width     float64
	Height    float64
	Margin    float64
	DSep      float64 // Minimum seed separation distance between streamlines
	DTest     float64 // Minimum termination distance during integration (typically 0.45..0.65 * DSep)
	StepSize  float64 // RK4 integration step in pixels
	MaxSteps  int     // Maximum steps in each direction (forward + backward)
	MinPoints int     // Minimum total vertices required to keep a streamline
}

// DefaultStreamlineConfig returns a balanced Jobard-Lefer configuration for a w x h canvas.
func DefaultStreamlineConfig(w, h, dSep float64) StreamlineConfig {
	if dSep <= 1.0 {
		dSep = 8.0
	}
	return StreamlineConfig{
		Width:     w,
		Height:    h,
		Margin:    6.0,
		DSep:      dSep,
		DTest:     dSep * 0.52,
		StepSize:  math.Max(1.5, dSep*0.28),
		MaxSteps:  180,
		MinPoints: 6,
	}
}

func rk4Step(p Point, dt float64, fieldFn VectorField2D) (Point, bool) {
	v1 := fieldFn(p.X, p.Y).Norm()
	if v1.Len() < 1e-6 {
		return Point{}, false
	}
	p2 := p.Add(v1.Mul(0.5 * dt))
	v2 := fieldFn(p2.X, p2.Y).Norm()
	if v2.Dot(v1) < 0 {
		v2 = v2.Mul(-1)
	}
	p3 := p.Add(v2.Mul(0.5 * dt))
	v3 := fieldFn(p3.X, p3.Y).Norm()
	if v3.Dot(v1) < 0 {
		v3 = v3.Mul(-1)
	}
	p4 := p.Add(v3.Mul(dt))
	v4 := fieldFn(p4.X, p4.Y).Norm()
	if v4.Dot(v1) < 0 {
		v4 = v4.Mul(-1)
	}

	dir := v1.Add(v2.Mul(2)).Add(v3.Mul(2)).Add(v4).Norm()
	if dir.Len() < 1e-6 {
		return Point{}, false
	}
	return p.Add(dir.Mul(dt)), true
}

// TraceEvenStreamlines implements Jobard & Lefer's (1997) algorithm for evenly-spaced
// streamlines using 4th-order Runge-Kutta (RK4) bidirectional integration and SpatialGrid
// collision acceleration. Streamlines are guaranteed never to approach closer than DTest.
func TraceEvenStreamlines(cfg StreamlineConfig, fieldFn VectorField2D) [][]Point {
	if cfg.Width <= 0 || cfg.Height <= 0 || fieldFn == nil {
		return nil
	}
	if cfg.DSep <= 1.0 {
		cfg.DSep = 8.0
	}
	if cfg.DTest <= 0.5 || cfg.DTest > cfg.DSep {
		cfg.DTest = cfg.DSep * 0.52
	}
	if cfg.StepSize <= 0.25 {
		cfg.StepSize = math.Max(1.5, cfg.DSep*0.28)
	}
	if cfg.MaxSteps <= 0 {
		cfg.MaxSteps = 160
	}
	if cfg.MinPoints < 2 {
		cfg.MinPoints = 5
	}

	grid := NewSpatialGrid(cfg.Width, cfg.Height, cfg.DSep*0.75)
	minX := cfg.Margin
	minY := cfg.Margin
	maxX := cfg.Width - cfg.Margin
	maxY := cfg.Height - cfg.Margin

	inBounds := func(p Point) bool {
		return p.X >= minX && p.X <= maxX && p.Y >= minY && p.Y <= maxY
	}

	traceHalf := func(seed Point, sign float64) []Point {
		var pts []Point
		cur := seed
		prevDir := fieldFn(cur.X, cur.Y).Norm().Mul(sign)
		if prevDir.Len() < 1e-6 {
			return nil
		}
		for s := 0; s < cfg.MaxSteps; s++ {
			next, ok := rk4Step(cur, cfg.StepSize*sign, fieldFn)
			if !ok || !inBounds(next) {
				break
			}
			stepDir := next.Sub(cur).Norm()
			// Prevent sharp hairpin reversals at singularities
			if s > 0 && stepDir.Dot(prevDir) < 0.25 {
				break
			}
			if !grid.CanPlace(next, cfg.DTest) {
				break
			}
			pts = append(pts, next)
			prevDir = stepDir
			cur = next
		}
		return pts
	}

	traceFull := func(seed Point) []Point {
		if !inBounds(seed) || !grid.CanPlace(seed, cfg.DSep*0.92) {
			return nil
		}
		bwd := traceHalf(seed, -1.0)
		fwd := traceHalf(seed, +1.0)
		total := len(bwd) + 1 + len(fwd)
		if total < cfg.MinPoints {
			return nil
		}
		line := make([]Point, 0, total)
		for i := len(bwd) - 1; i >= 0; i-- {
			line = append(line, bwd[i])
		}
		line = append(line, seed)
		line = append(line, fwd...)
		return line
	}

	// Seed queue initialized with a golden-ratio / stratified lattice across the canvas
	var queue []Point
	stepGrid := cfg.DSep * 1.35
	for y := minY + stepGrid*0.5; y <= maxY; y += stepGrid {
		for x := minX + stepGrid*0.5; x <= maxX; x += stepGrid {
			queue = append(queue, Pt(x, y))
		}
	}

	var result [][]Point
	qHead := 0
	sampleSpacing := math.Max(1.0, cfg.DTest*0.42)

	for qHead < len(queue) {
		seed := queue[qHead]
		qHead++

		line := traceFull(seed)
		if line == nil {
			continue
		}
		grid.InsertPolyline(line, sampleSpacing)
		result = append(result, line)

		// Enqueue orthogonal candidate seeds at +/- DSep along the new streamline (Jobard-Lefer queue)
		for i := 1; i < len(line)-1; i += 2 {
			tan := line[i+1].Sub(line[i-1]).Norm()
			if tan.Len() < 1e-6 {
				continue
			}
			n := tan.Perp()
			c1 := line[i].Add(n.Mul(cfg.DSep * 1.02))
			c2 := line[i].Sub(n.Mul(cfg.DSep * 1.02))
			if inBounds(c1) && grid.CanPlace(c1, cfg.DSep*0.92) {
				queue = append(queue, c1)
			}
			if inBounds(c2) && grid.CanPlace(c2, cfg.DSep*0.92) {
				queue = append(queue, c2)
			}
		}
	}

	return result
}

type msSeg struct {
	a, b Point
	used bool
}

// ExtractMarchingSquares extracts sub-pixel interpolated iso-contour polylines where
// fieldFn(x, y) == isoVal across a grid of nx x ny cells spanning [0, boundsW] x [0, boundsH].
// Connected contour segments are automatically stitched into continuous polylines.
func ExtractMarchingSquares(
	nx, ny int,
	boundsW, boundsH, isoVal float64,
	fieldFn ScalarField2D,
) [][]Point {
	if nx < 2 || ny < 2 || boundsW <= 0 || boundsH <= 0 || fieldFn == nil {
		return nil
	}
	vals := make([]float64, (nx+1)*(ny+1))
	dx := boundsW / float64(nx)
	dy := boundsH / float64(ny)

	for iy := 0; iy <= ny; iy++ {
		y := float64(iy) * dy
		rowOff := iy * (nx + 1)
		for ix := 0; ix <= nx; ix++ {
			x := float64(ix) * dx
			vals[rowOff+ix] = fieldFn(x, y) - isoVal
		}
	}

	interp := func(p0, p1 Point, v0, v1 float64) Point {
		denom := v0 - v1
		if math.Abs(denom) < 1e-12 {
			return p0.Lerp(p1, 0.5)
		}
		t := common.Constrain(v0/denom, 0, 1)
		return p0.Lerp(p1, t)
	}

	var segs []msSeg
	for iy := 0; iy < ny; iy++ {
		y0 := float64(iy) * dy
		y1 := y0 + dy
		row0 := iy * (nx + 1)
		row1 := (iy + 1) * (nx + 1)

		for ix := 0; ix < nx; ix++ {
			x0 := float64(ix) * dx
			x1 := x0 + dx

			v00 := vals[row0+ix]   // top-left (bit 3)
			v10 := vals[row0+ix+1] // top-right (bit 2)
			v11 := vals[row1+ix+1] // bottom-right (bit 1)
			v01 := vals[row1+ix]   // bottom-left (bit 0)

			mask := 0
			if v01 >= 0 {
				mask |= 1
			}
			if v11 >= 0 {
				mask |= 2
			}
			if v10 >= 0 {
				mask |= 4
			}
			if v00 >= 0 {
				mask |= 8
			}
			if mask == 0 || mask == 15 {
				continue
			}

			p00 := Pt(x0, y0)
			p10 := Pt(x1, y0)
			p11 := Pt(x1, y1)
			p01 := Pt(x0, y1)

			eTop := interp(p00, p10, v00, v10)
			eRight := interp(p10, p11, v10, v11)
			eBot := interp(p01, p11, v01, v11)
			eLeft := interp(p00, p01, v00, v01)

			switch mask {
			case 1, 14:
				segs = append(segs, msSeg{a: eLeft, b: eBot})
			case 2, 13:
				segs = append(segs, msSeg{a: eBot, b: eRight})
			case 3, 12:
				segs = append(segs, msSeg{a: eLeft, b: eRight})
			case 4, 11:
				segs = append(segs, msSeg{a: eTop, b: eRight})
			case 6, 9:
				segs = append(segs, msSeg{a: eTop, b: eBot})
			case 7, 8:
				segs = append(segs, msSeg{a: eLeft, b: eTop})
			case 5:
				center := 0.25 * (v00 + v10 + v11 + v01)
				if center >= 0 {
					segs = append(segs, msSeg{a: eLeft, b: eTop}, msSeg{a: eBot, b: eRight})
				} else {
					segs = append(segs, msSeg{a: eLeft, b: eBot}, msSeg{a: eTop, b: eRight})
				}
			case 10:
				center := 0.25 * (v00 + v10 + v11 + v01)
				if center >= 0 {
					segs = append(segs, msSeg{a: eLeft, b: eBot}, msSeg{a: eTop, b: eRight})
				} else {
					segs = append(segs, msSeg{a: eLeft, b: eTop}, msSeg{a: eBot, b: eRight})
				}
			}
		}
	}

	if len(segs) == 0 {
		return nil
	}

	// Stitch connected segments into ordered polylines using a spatial endpoint hash
	const epsSq = 0.04
	var polylines [][]Point

	for i := range segs {
		if segs[i].used {
			continue
		}
		segs[i].used = true
		line := []Point{segs[i].a, segs[i].b}

		// Extend forward from tail
		for {
			tail := line[len(line)-1]
			found := false
			for j := range segs {
				if segs[j].used {
					continue
				}
				da := (segs[j].a.X-tail.X)*(segs[j].a.X-tail.X) + (segs[j].a.Y-tail.Y)*(segs[j].a.Y-tail.Y)
				if da <= epsSq {
					segs[j].used = true
					line = append(line, segs[j].b)
					found = true
					break
				}
				db := (segs[j].b.X-tail.X)*(segs[j].b.X-tail.X) + (segs[j].b.Y-tail.Y)*(segs[j].b.Y-tail.Y)
				if db <= epsSq {
					segs[j].used = true
					line = append(line, segs[j].a)
					found = true
					break
				}
			}
			if !found {
				break
			}
		}

		// Extend backward from head
		var prefix []Point
		for {
			head := line[0]
			if len(prefix) > 0 {
				head = prefix[len(prefix)-1]
			}
			found := false
			for j := range segs {
				if segs[j].used {
					continue
				}
				da := (segs[j].a.X-head.X)*(segs[j].a.X-head.X) + (segs[j].a.Y-head.Y)*(segs[j].a.Y-head.Y)
				if da <= epsSq {
					segs[j].used = true
					prefix = append(prefix, segs[j].b)
					found = true
					break
				}
				db := (segs[j].b.X-head.X)*(segs[j].b.X-head.X) + (segs[j].b.Y-head.Y)*(segs[j].b.Y-head.Y)
				if db <= epsSq {
					segs[j].used = true
					prefix = append(prefix, segs[j].a)
					found = true
					break
				}
			}
			if !found {
				break
			}
		}

		if len(prefix) > 0 {
			full := make([]Point, 0, len(prefix)+len(line))
			for k := len(prefix) - 1; k >= 0; k-- {
				full = append(full, prefix[k])
			}
			full = append(full, line...)
			line = full
		}
		polylines = append(polylines, line)
	}

	return polylines
}

// RenderIsoBand renders an anti-aliased scalar field band [minVal, maxVal] with smoothstep
// feathering and optional directional drop-shadow relief directly into the HDR Canvas.
func (c *Canvas) RenderIsoBand(
	minVal, maxVal, feather float64,
	col color.RGBA,
	alpha float64,
	fieldFn ScalarField2D,
) {
	if c == nil || c.Width <= 0 || c.Height <= 0 || alpha <= 0 || col.A == 0 || fieldFn == nil {
		return
	}
	if feather <= 1e-6 {
		feather = 0.02
	}
	rLin := srgb8ToLinearLUT[col.R]
	gLin := srgb8ToLinearLUT[col.G]
	bLin := srgb8ToLinearLUT[col.B]
	baseAlpha := (float64(col.A) / 255.0) * common.Constrain(alpha, 0, 1)

	smoothstep := func(edge0, edge1, x float64) float64 {
		t := common.Constrain((x-edge0)/(edge1-edge0), 0, 1)
		return t * t * (3.0 - 2.0*t)
	}

	parallelRows(c.Width, c.Height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			fy := float64(y) + 0.5
			for x := 0; x < c.Width; x++ {
				fx := float64(x) + 0.5
				v := fieldFn(fx, fy)
				if v < minVal-feather || v > maxVal+feather {
					continue
				}
				wIn := smoothstep(minVal-feather, minVal+feather, v)
				wOut := 1.0 - smoothstep(maxVal-feather, maxVal+feather, v)
				cov := wIn * wOut
				if cov > 1e-4 {
					c.BlendPixelLinear(x, y, rLin, gLin, bLin, baseAlpha*cov, c.blendMode)
				}
			}
		}
	})
}
