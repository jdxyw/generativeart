package render

import (
	"image/color"
	"math"
	"math/rand"

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

// ============================================================================
// Part VII Section 7.1 — Fast Line Integral Convolution (LIC & OFLIC)
// ============================================================================

// LICConfig configures Fast Line Integral Convolution (LIC) and Oriented Fast LIC (OFLIC).
type LICConfig struct {
	StreamLength  float64 // integration half-length in pixels (default 18.0)
	StepSize      float64 // RK4 step size in pixels (default 1.0)
	Contrast      float64 // contrast stretch around 0.5 (default 1.35)
	AnimatedPhase float64 // phase in [0, 1) for seamless periodic OFLIC animation
	AnimatedFreq  float64 // spatial wave frequency along streamline (0 = static Hann LIC, >0 = periodic OFLIC wave)
	NoiseScale    float64 // multi-scale noise frequency (default 0.45)
}

// DefaultLICConfig returns a balanced LICConfig for the given integration half-length.
func DefaultLICConfig(streamLen float64) LICConfig {
	if streamLen <= 1.0 {
		streamLen = 18.0
	}
	return LICConfig{
		StreamLength:  streamLen,
		StepSize:      1.0,
		Contrast:      1.35,
		AnimatedPhase: 0.0,
		AnimatedFreq:  0.0,
		NoiseScale:    0.45,
	}
}

func licHash2D(ix, iy int, seed uint64) float64 {
	h := uint64(uint32(ix))*0x9E3779B185EBCA87 ^ uint64(uint32(iy))*0xC2B2AE3D27D4EB4F ^ seed
	h ^= h >> 33
	h *= 0xFF51AFD7ED558CCD
	h ^= h >> 33
	h *= 0xC4CEB9FE1A85EC53
	h ^= h >> 33
	return float64(h&0x00FFFFFFFFFFFFFF) / float64(0x0100000000000000)
}

func licSmoothNoise2D(x, y float64, seed uint64) float64 {
	ix := int(math.Floor(x))
	iy := int(math.Floor(y))
	fx := x - float64(ix)
	fy := y - float64(iy)
	ux := fx * fx * (3.0 - 2.0*fx)
	uy := fy * fy * (3.0 - 2.0*fy)

	v00 := licHash2D(ix, iy, seed)
	v10 := licHash2D(ix+1, iy, seed)
	v01 := licHash2D(ix, iy+1, seed)
	v11 := licHash2D(ix+1, iy+1, seed)

	top := v00 + ux*(v10-v00)
	bot := v01 + ux*(v11-v01)
	return top + uy*(bot-top)
}

// ComputeLICField computes a w x h Line Integral Convolution (or Oriented Fast LIC) scalar field in [0, 1].
func ComputeLICField(w, h int, cfg LICConfig, fieldFn VectorField2D, seed int64) []float64 {
	if w <= 0 || h <= 0 {
		return nil
	}
	streamLen := cfg.StreamLength
	if streamLen <= 1.0 {
		streamLen = 18.0
	}
	stepSize := cfg.StepSize
	if stepSize <= 0.1 {
		stepSize = 1.0
	}
	contrast := cfg.Contrast
	if contrast <= 0 {
		contrast = 1.35
	}
	noiseScale := cfg.NoiseScale
	if noiseScale <= 0 {
		noiseScale = 0.45
	}

	uSeed := uint64(seed)*0x9E3779B97F4A7C15 + 0xBF58476D1CE4E5B9
	noiseTex := make([]float64, w*h)

	// Precompute multi-scale white + band-passed noise texture in [0, 1]
	parallelRows(w, h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			fy := float64(y) + 0.5
			rowOff := y * w
			for x := 0; x < w; x++ {
				fx := float64(x) + 0.5
				white := licHash2D(x, y, uSeed)
				band1 := licSmoothNoise2D(fx*noiseScale, fy*noiseScale, uSeed^0x5555555555555555)
				band2 := licSmoothNoise2D(fx*noiseScale*2.15, fy*noiseScale*2.15, uSeed^0xAAAAAAAAAAAAAAAA)
				v := 0.44*white + 0.36*band1 + 0.20*band2
				// Stretch input noise slightly around 0.5 so integration preserves crisp variance
				v = common.Constrain(0.5+(v-0.5)*1.45, 0.0, 1.0)
				noiseTex[rowOff+x] = v
			}
		}
	})

	if fieldFn == nil {
		return noiseTex
	}

	fw := float64(w)
	fh := float64(h)

	sampleBilinear := func(px, py float64) float64 {
		gx := common.Constrain(px-0.5, 0.0, fw-1.000001)
		gy := common.Constrain(py-0.5, 0.0, fh-1.000001)
		ix := int(gx)
		iy := int(gy)
		if ix >= w-1 {
			ix = w - 2
		}
		if ix < 0 {
			ix = 0
		}
		if iy >= h-1 {
			iy = h - 2
		}
		if iy < 0 {
			iy = 0
		}
		tx := gx - float64(ix)
		ty := gy - float64(iy)
		ix1 := ix + 1
		iy1 := iy + 1
		if ix1 >= w {
			ix1 = w - 1
		}
		if iy1 >= h {
			iy1 = h - 1
		}
		r0 := iy * w
		r1 := iy1 * w
		v00 := noiseTex[r0+ix]
		v10 := noiseTex[r0+ix1]
		v01 := noiseTex[r1+ix]
		v11 := noiseTex[r1+ix1]
		return (v00*(1.0-tx)+v10*tx)*(1.0-ty) + (v01*(1.0-tx)+v11*tx)*ty
	}

	out := make([]float64, w*h)
	twoPiPhase := 2.0 * math.Pi * cfg.AnimatedPhase
	useOFLIC := cfg.AnimatedFreq > 0

	evalKernel := func(signedS float64) float64 {
		hann := 0.5 * (1.0 + math.Cos(math.Pi*signedS/streamLen))
		if useOFLIC {
			hann *= 0.65 + 0.35*math.Cos(cfg.AnimatedFreq*signedS-twoPiPhase)
		}
		return hann
	}

	parallelRows(w, h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			fy := float64(y) + 0.5
			rowOff := y * w
			for x := 0; x < w; x++ {
				fx := float64(x) + 0.5
				w0 := evalKernel(0.0)
				sumVal := w0 * noiseTex[rowOff+x]
				sumWgt := w0

				// Trace forward (+1) and backward (-1) along fieldFn
				for dirIdx := 0; dirIdx < 2; dirIdx++ {
					sign := 1.0
					if dirIdx == 1 {
						sign = -1.0
					}
					cur := Pt(fx, fy)
					prevDir := fieldFn(cur.X, cur.Y).Norm().Mul(sign)
					if prevDir.Len() < 1e-6 {
						continue
					}
					for s := stepSize; s <= streamLen; s += stepSize {
						next, ok := rk4Step(cur, stepSize*sign, fieldFn)
						if !ok || next.X < 0 || next.X >= fw || next.Y < 0 || next.Y >= fh {
							break
						}
						stepDir := next.Sub(cur).Norm()
						if stepDir.Dot(prevDir) < -0.25 {
							break
						}
						signedS := sign * s
						wgt := evalKernel(signedS)
						sumVal += wgt * sampleBilinear(next.X, next.Y)
						sumWgt += wgt
						prevDir = stepDir
						cur = next
					}
				}

				lic := 0.5
				if sumWgt > 1e-9 {
					lic = sumVal / sumWgt
				}
				// Centered contrast stretch around 0.5
				lic = common.Constrain(0.5+(lic-0.5)*contrast*2.1, 0.0, 1.0)
				out[rowOff+x] = lic
			}
		}
	})

	return out
}

// RenderLICField computes the LIC/OFLIC scalar field and composites colorFn(x+0.5, y+0.5, licVal)
// into the HDR Canvas in parallel.
func (c *Canvas) RenderLICField(
	cfg LICConfig,
	fieldFn VectorField2D,
	colorFn func(x, y, licVal float64) color.RGBA,
	alpha float64,
	seed int64,
) {
	if c == nil || c.Width <= 0 || c.Height <= 0 || fieldFn == nil || colorFn == nil || alpha <= 0 {
		return
	}
	licField := ComputeLICField(c.Width, c.Height, cfg, fieldFn, seed)
	if len(licField) != c.Width*c.Height {
		return
	}
	baseAlpha := common.Constrain(alpha, 0.0, 1.0)

	parallelRows(c.Width, c.Height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			fy := float64(y) + 0.5
			rowOff := y * c.Width
			for x := 0; x < c.Width; x++ {
				fx := float64(x) + 0.5
				col := colorFn(fx, fy, licField[rowOff+x])
				if col.A == 0 {
					continue
				}
				rLin := srgb8ToLinearLUT[col.R]
				gLin := srgb8ToLinearLUT[col.G]
				bLin := srgb8ToLinearLUT[col.B]
				a := (float64(col.A) / 255.0) * baseAlpha
				c.BlendPixelLinear(x, y, rLin, gLin, bLin, a, c.blendMode)
			}
		}
	})
}

// ============================================================================
// Part VII Section 7.4 — Weighted Blue-Noise Poisson-Disk & Centroidal Voronoi
// ============================================================================

// SamplePoissonDisk2D generates a blue-noise Poisson-disk point distribution over [0, w) x [0, h).
// If densityFn != nil, local spacing at (x, y) is minRadius / (0.4 + 0.6*common.Constrain(densityFn(x, y), 0, 1))
// so high-density regions pack tighter and low-density regions pack wider.
func SamplePoissonDisk2D(
	w, h, minRadius float64,
	densityFn ScalarField2D,
	rng *rand.Rand,
) []Point {
	if w <= 0 || h <= 0 {
		return nil
	}
	if minRadius <= 0.5 {
		minRadius = 4.0
	}
	if rng == nil {
		rng = rand.New(rand.NewSource(1337))
	}

	evalRadius := func(x, y float64) float64 {
		if densityFn == nil {
			return minRadius
		}
		d := common.Constrain(densityFn(x, y), 0.0, 1.0)
		return minRadius / (0.4 + 0.6*d)
	}

	maxRadius := minRadius
	if densityFn != nil {
		maxRadius = minRadius / 0.4
	}

	cellSize := minRadius / math.Sqrt2
	if cellSize < 0.5 {
		cellSize = 0.5
	}
	invCell := 1.0 / cellSize
	gw := int(math.Ceil(w*invCell)) + 1
	gh := int(math.Ceil(h*invCell)) + 1

	head := make([]int, gw*gh)
	for i := range head {
		head[i] = -1
	}

	var pts []Point
	var radii []float64
	var next []int
	var active []int

	canPlace := func(cand Point, rCand float64) bool {
		if cand.X < 0 || cand.X >= w || cand.Y < 0 || cand.Y >= h {
			return false
		}
		cx := int(cand.X * invCell)
		cy := int(cand.Y * invCell)
		searchR := math.Max(rCand, maxRadius)
		radCells := int(math.Ceil(searchR*invCell)) + 1

		x0 := cx - radCells
		if x0 < 0 {
			x0 = 0
		}
		x1 := cx + radCells
		if x1 >= gw {
			x1 = gw - 1
		}
		y0 := cy - radCells
		if y0 < 0 {
			y0 = 0
		}
		y1 := cy + radCells
		if y1 >= gh {
			y1 = gh - 1
		}

		for ny := y0; ny <= y1; ny++ {
			rowOff := ny * gw
			for nx := x0; nx <= x1; nx++ {
				for idx := head[rowOff+nx]; idx != -1; idx = next[idx] {
					p := pts[idx]
					dx := cand.X - p.X
					dy := cand.Y - p.Y
					req := 0.5 * (rCand + radii[idx])
					if dx*dx+dy*dy < req*req {
						return false
					}
				}
			}
		}
		return true
	}

	addPoint := func(p Point, r float64) {
		idx := len(pts)
		pts = append(pts, p)
		radii = append(radii, r)
		cx := int(p.X * invCell)
		cy := int(p.Y * invCell)
		if cx < 0 {
			cx = 0
		} else if cx >= gw {
			cx = gw - 1
		}
		if cy < 0 {
			cy = 0
		} else if cy >= gh {
			cy = gh - 1
		}
		cellIdx := cy*gw + cx
		next = append(next, head[cellIdx])
		head[cellIdx] = idx
		active = append(active, idx)
	}

	// Initial seed near center
	p0 := Pt(w*(0.35+0.3*rng.Float64()), h*(0.35+0.3*rng.Float64()))
	addPoint(p0, evalRadius(p0.X, p0.Y))

	const kCandidates = 28
	for len(active) > 0 {
		aIdx := rng.Intn(len(active))
		parentIdx := active[aIdx]
		parent := pts[parentIdx]
		rParent := radii[parentIdx]

		found := false
		baseAngle := rng.Float64() * 2.0 * math.Pi
		for k := 0; k < kCandidates; k++ {
			angle := baseAngle + (float64(k)+rng.Float64())*(2.0*math.Pi/float64(kCandidates))
			dist := rParent * (1.0 + rng.Float64())
			cand := Pt(parent.X+math.Cos(angle)*dist, parent.Y+math.Sin(angle)*dist)
			if cand.X < 0 || cand.X >= w || cand.Y < 0 || cand.Y >= h {
				continue
			}
			rCand := evalRadius(cand.X, cand.Y)
			if canPlace(cand, rCand) {
				addPoint(cand, rCand)
				found = true
				break
			}
		}
		if !found {
			last := len(active) - 1
			active[aIdx] = active[last]
			active = active[:last]
		}
	}

	return pts
}

// RelaxCentroidalVoronoi performs iterations passes of weighted Lloyd Centroidal Voronoi
// Tessellation (CVT) relaxation on a copy of pts over a gridRes x gridRes discretization
// of [0, w) x [0, h), weighting each grid cell by densityFn(x, y) (or 1.0 if nil).
func RelaxCentroidalVoronoi(
	pts []Point,
	w, h float64,
	gridRes int,
	densityFn ScalarField2D,
	iterations int,
) []Point {
	n := len(pts)
	if n == 0 || w <= 0 || h <= 0 {
		return nil
	}
	sites := make([]Point, n)
	copy(sites, pts)
	if iterations <= 0 || n == 1 {
		return sites
	}
	if gridRes < 16 {
		gridRes = 128
	}

	dx := w / float64(gridRes)
	dy := h / float64(gridRes)

	// Precompute density weights across the discretization grid
	weights := make([]float64, gridRes*gridRes)
	if densityFn == nil {
		for i := range weights {
			weights[i] = 1.0
		}
	} else {
		for iy := 0; iy < gridRes; iy++ {
			y := (float64(iy) + 0.5) * dy
			rowOff := iy * gridRes
			for ix := 0; ix < gridRes; ix++ {
				x := (float64(ix) + 0.5) * dx
				d := common.Constrain(densityFn(x, y), 0.0, 1.0)
				weights[rowOff+ix] = 0.05 + 0.95*d
			}
		}
	}

	// Spatial acceleration bucket grid for fast O(1) nearest-site lookup
	aspect := w / h
	if aspect <= 0 {
		aspect = 1.0
	}
	bw := int(math.Ceil(math.Sqrt(float64(n) * aspect)))
	bh := int(math.Ceil(math.Sqrt(float64(n) / aspect)))
	if bw < 4 {
		bw = 4
	}
	if bh < 4 {
		bh = 4
	}
	invBW := float64(bw) / w
	invBH := float64(bh) / h
	bucketW := w / float64(bw)
	bucketH := h / float64(bh)
	minBucketDim := math.Min(bucketW, bucketH)

	head := make([]int, bw*bh)
	next := make([]int, n)
	sumX := make([]float64, n)
	sumY := make([]float64, n)
	sumW := make([]float64, n)

	maxRing := bw
	if bh > maxRing {
		maxRing = bh
	}

	for iter := 0; iter < iterations; iter++ {
		for i := range head {
			head[i] = -1
		}
		for i := 0; i < n; i++ {
			bx := int(sites[i].X * invBW)
			by := int(sites[i].Y * invBH)
			if bx < 0 {
				bx = 0
			} else if bx >= bw {
				bx = bw - 1
			}
			if by < 0 {
				by = 0
			} else if by >= bh {
				by = bh - 1
			}
			bIdx := by*bw + bx
			next[i] = head[bIdx]
			head[bIdx] = i
			sumX[i] = 0
			sumY[i] = 0
			sumW[i] = 0
		}

		for iy := 0; iy < gridRes; iy++ {
			y := (float64(iy) + 0.5) * dy
			by := int(y * invBH)
			if by >= bh {
				by = bh - 1
			}
			rowOff := iy * gridRes

			for ix := 0; ix < gridRes; ix++ {
				x := (float64(ix) + 0.5) * dx
				bx := int(x * invBW)
				if bx >= bw {
					bx = bw - 1
				}

				bestIdx := 0
				bestDistSq := math.MaxFloat64

				// Expand concentric bucket rings until lower-bound distance exceeds bestDistSq
				for ring := 0; ring <= maxRing; ring++ {
					if ring > 1 {
						minRingDist := float64(ring-1) * minBucketDim
						if minRingDist*minRingDist > bestDistSq {
							break
						}
					}

					ry0 := by - ring
					if ry0 < 0 {
						ry0 = 0
					}
					ry1 := by + ring
					if ry1 >= bh {
						ry1 = bh - 1
					}
					rx0 := bx - ring
					if rx0 < 0 {
						rx0 = 0
					}
					rx1 := bx + ring
					if rx1 >= bw {
						rx1 = bw - 1
					}

					for cy := ry0; cy <= ry1; cy++ {
						bRow := cy * bw
						isTopOrBot := cy == by-ring || cy == by+ring
						for cx := rx0; cx <= rx1; cx++ {
							if ring > 0 && !isTopOrBot && cx != bx-ring && cx != bx+ring {
								continue
							}
							for sIdx := head[bRow+cx]; sIdx != -1; sIdx = next[sIdx] {
								s := sites[sIdx]
								ddx := x - s.X
								ddy := y - s.Y
								dSq := ddx*ddx + ddy*ddy
								if dSq < bestDistSq {
									bestDistSq = dSq
									bestIdx = sIdx
								}
							}
						}
					}
				}

				wgt := weights[rowOff+ix]
				sumX[bestIdx] += x * wgt
				sumY[bestIdx] += y * wgt
				sumW[bestIdx] += wgt
			}
		}

		for i := 0; i < n; i++ {
			if sumW[i] > 1e-12 {
				sites[i].X = common.Constrain(sumX[i]/sumW[i], 0.0, w)
				sites[i].Y = common.Constrain(sumY[i]/sumW[i], 0.0, h)
			}
		}
	}

	return sites
}
