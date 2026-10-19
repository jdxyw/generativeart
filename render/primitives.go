package render

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart/common"
	"github.com/tdewolff/canvas"
)

// SplatPoint performs bilinear sub-pixel splatting of a point at floating-point coordinates (x, y)
// across the 4 nearest pixels ((x0, y0), (x1, y0), (x0, y1), (x1, y1)) with exact fractional weights
// that sum to 1.0 (conserving total energy in linear light).
func (c *Canvas) SplatPoint(x, y float64, col color.RGBA, alpha float64) {
	if c == nil || alpha <= 0 || col.A == 0 {
		return
	}
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	fx := x - float64(x0)
	fy := y - float64(y0)

	w00 := (1.0 - fx) * (1.0 - fy)
	w10 := fx * (1.0 - fy)
	w01 := (1.0 - fx) * fy
	w11 := fx * fy

	rLin := SRGBToLinear(float64(col.R) / 255.0)
	gLin := SRGBToLinear(float64(col.G) / 255.0)
	bLin := SRGBToLinear(float64(col.B) / 255.0)
	baseAlpha := (float64(col.A) / 255.0) * common.Constrain(alpha, 0, 1)

	c.BlendPixelLinear(x0, y0, rLin, gLin, bLin, baseAlpha*w00, c.blendMode)
	c.BlendPixelLinear(x0+1, y0, rLin, gLin, bLin, baseAlpha*w10, c.blendMode)
	c.BlendPixelLinear(x0, y0+1, rLin, gLin, bLin, baseAlpha*w01, c.blendMode)
	c.BlendPixelLinear(x0+1, y0+1, rLin, gLin, bLin, baseAlpha*w11, c.blendMode)
}

// DrawGlowDot renders an analytical radial Gaussian/smoothstep soft dot centered at sub-pixel (x, y).
// hardness in [0, 1] controls the profile between a soft Gaussian glow (0) and a crisp smoothstep disc (1).
func (c *Canvas) DrawGlowDot(x, y, radius float64, col color.RGBA, alpha, hardness float64) {
	if c == nil || radius <= 0 || alpha <= 0 || col.A == 0 {
		return
	}
	if radius < 0.75 {
		c.SplatPoint(x, y, col, alpha*common.Constrain(radius/0.75, 0.1, 1.0))
		return
	}

	hardness = common.Constrain(hardness, 0, 1)
	rLin := SRGBToLinear(float64(col.R) / 255.0)
	gLin := SRGBToLinear(float64(col.G) / 255.0)
	bLin := SRGBToLinear(float64(col.B) / 255.0)
	baseAlpha := (float64(col.A) / 255.0) * common.Constrain(alpha, 0, 1)

	pad := radius + 1.5
	minX := int(math.Floor(x - pad))
	maxX := int(math.Ceil(x + pad))
	minY := int(math.Floor(y - pad))
	maxY := int(math.Ceil(y + pad))

	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX >= c.Width {
		maxX = c.Width - 1
	}
	if maxY >= c.Height {
		maxY = c.Height - 1
	}

	invR := 1.0 / radius
	for py := minY; py <= maxY; py++ {
		dy := (float64(py) + 0.5) - y
		for px := minX; px <= maxX; px++ {
			dx := (float64(px) + 0.5) - x
			dist := math.Hypot(dx, dy)
			u := dist * invR
			if u >= 1.25 {
				continue
			}

			// Gaussian falloff component
			gauss := math.Exp(-3.5 * u * u)
			if u > 1.0 {
				// Smoothly taper tail to zero by u = 1.25
				tail := common.Constrain(1.0-(u-1.0)/0.25, 0, 1)
				gauss *= tail * tail
			}

			// Smoothstep disc component with 1px anti-aliased edge
			edge0 := math.Max(0, radius-0.75)
			edge1 := radius + 0.75
			t := common.Constrain((dist-edge0)/(edge1-edge0), 0, 1)
			disc := 1.0 - t*t*(3.0-2.0*t)

			w := gauss*(1.0-hardness) + disc*hardness
			if w > 1e-5 {
				c.BlendPixelLinear(px, py, rLin, gLin, bLin, baseAlpha*w, c.blendMode)
			}
		}
	}
}

// DrawSoftRing renders an anti-aliased soft ring/halo centered at sub-pixel (x, y)
// with given radius and radial thickness.
func (c *Canvas) DrawSoftRing(x, y, radius, thickness float64, col color.RGBA, alpha float64) {
	if c == nil || radius <= 0 || thickness <= 0 || alpha <= 0 || col.A == 0 {
		return
	}
	rLin := SRGBToLinear(float64(col.R) / 255.0)
	gLin := SRGBToLinear(float64(col.G) / 255.0)
	bLin := SRGBToLinear(float64(col.B) / 255.0)
	baseAlpha := (float64(col.A) / 255.0) * common.Constrain(alpha, 0, 1)

	halfThick := math.Max(0.5, thickness*0.5)
	outer := radius + halfThick + 1.5
	inner := math.Max(0, radius-halfThick-1.5)
	innerSq := inner * inner

	minX := int(math.Floor(x - outer))
	maxX := int(math.Ceil(x + outer))
	minY := int(math.Floor(y - outer))
	maxY := int(math.Ceil(y + outer))

	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX >= c.Width {
		maxX = c.Width - 1
	}
	if maxY >= c.Height {
		maxY = c.Height - 1
	}

	aaWidth := math.Max(1.0, halfThick)
	for py := minY; py <= maxY; py++ {
		dy := (float64(py) + 0.5) - y
		for px := minX; px <= maxX; px++ {
			dx := (float64(px) + 0.5) - x
			dSq := dx*dx + dy*dy
			if dSq < innerSq {
				continue
			}
			dist := math.Sqrt(dSq)
			radialDist := math.Abs(dist - radius)
			if radialDist >= aaWidth+1.0 {
				continue
			}
			u := common.Constrain(radialDist/aaWidth, 0, 1)
			// Smooth bell profile across the ring thickness
			w := (1.0 - u*u*(3.0-2.0*u)) * math.Exp(-1.5*u*u)
			if w > 1e-5 {
				c.BlendPixelLinear(px, py, rLin, gLin, bLin, baseAlpha*w, c.blendMode)
			}
		}
	}
}

// CatmullRomSpline evaluates a smooth centripetal/uniform Catmull-Rom spline through pts,
// returning a dense polyline with samplesPerSegment samples per control segment.
func CatmullRomSpline(pts []Point, samplesPerSegment int) []Point {
	n := len(pts)
	if n == 0 {
		return nil
	}
	if n == 1 {
		return []Point{pts[0]}
	}
	if samplesPerSegment < 2 {
		samplesPerSegment = 16
	}
	if n == 2 {
		out := make([]Point, 0, samplesPerSegment+1)
		for j := 0; j <= samplesPerSegment; j++ {
			t := float64(j) / float64(samplesPerSegment)
			out = append(out, pts[0].Lerp(pts[1], t))
		}
		return out
	}

	out := make([]Point, 0, (n-1)*samplesPerSegment+1)
	for i := 0; i < n-1; i++ {
		p1 := pts[i]
		p2 := pts[i+1]

		var p0, p3 Point
		if i == 0 {
			// Extrapolate phantom endpoint p0 = 2*p1 - p2
			p0 = p1.Mul(2).Sub(p2)
		} else {
			p0 = pts[i-1]
		}
		if i+2 >= n {
			// Extrapolate phantom endpoint p3 = 2*p2 - p1
			p3 = p2.Mul(2).Sub(p1)
		} else {
			p3 = pts[i+2]
		}

		steps := samplesPerSegment
		for j := 0; j < steps; j++ {
			t := float64(j) / float64(steps)
			out = append(out, evalCatmullRom(p0, p1, p2, p3, t))
		}
	}
	out = append(out, pts[n-1])
	return out
}

func evalCatmullRom(p0, p1, p2, p3 Point, t float64) Point {
	t2 := t * t
	t3 := t2 * t
	x := 0.5 * ((2.0 * p1.X) +
		(-p0.X+p2.X)*t +
		(2.0*p0.X-5.0*p1.X+4.0*p2.X-p3.X)*t2 +
		(-p0.X+3.0*p1.X-3.0*p2.X+p3.X)*t3)
	y := 0.5 * ((2.0 * p1.Y) +
		(-p0.Y+p2.Y)*t +
		(2.0*p0.Y-5.0*p1.Y+4.0*p2.Y-p3.Y)*t2 +
		(-p0.Y+3.0*p1.Y-3.0*p2.Y+p3.Y)*t3)
	return Point{X: x, Y: y}
}

// DrawSmoothSpline draws a uniform-width anti-aliased Catmull-Rom spline through pts.
func (c *Canvas) DrawSmoothSpline(pts []Point, width float64, col color.RGBA, alpha float64) {
	if c == nil || len(pts) < 2 || width <= 0 || alpha <= 0 || col.A == 0 {
		return
	}
	curve := CatmullRomSpline(pts, 20)
	c.StrokePolyline(curve, width, col, alpha)
}

// Preset width profiles for DrawTaperedSpline.

// TaperBothEnds tapers smoothly from 0 at both endpoints to 1.0 in the middle: sin(pi * t).
func TaperBothEnds(t float64) float64 {
	return math.Sin(math.Pi * common.Constrain(t, 0, 1))
}

// TaperTail starts at full width (1.0) at t=0 and tapers linearly to 0 at t=1.
func TaperTail(t float64) float64 {
	return 1.0 - common.Constrain(t, 0, 1)
}

// TaperHead starts at 0 at t=0 and grows linearly to full width (1.0) at t=1.
func TaperHead(t float64) float64 {
	return common.Constrain(t, 0, 1)
}

// TaperCalligraphic simulates a natural calligraphic brush pressure curve:
// quick attack, sustained mid-stroke modulation, and graceful tapered release.
func TaperCalligraphic(t float64) float64 {
	t = common.Constrain(t, 0, 1)
	envelope := math.Sin(math.Pi * math.Pow(t, 0.72))
	modulation := 0.82 + 0.18*math.Cos(2.0*math.Pi*t)
	return common.Constrain(envelope*modulation, 0, 1)
}

// TaperUniform maintains constant width (1.0) across the entire spline.
func TaperUniform(t float64) float64 {
	return 1.0
}

// DrawTaperedSpline constructs a smooth Catmull-Rom spline through pts, computes unit normals
// at each sample t in [0, 1], extrudes left/right ribbon vertices by maxWidth*widthProfile(t)/2,
// and rasterizes the variable-width stroke with smooth colorProfile(t) and active BlendMode.
func (c *Canvas) DrawTaperedSpline(
	pts []Point,
	maxWidth float64,
	widthProfile func(t float64) float64,
	colorProfile func(t float64) color.RGBA,
	alpha float64,
) {
	if c == nil || len(pts) < 2 || maxWidth <= 0 || alpha <= 0 {
		return
	}
	if widthProfile == nil {
		widthProfile = TaperBothEnds
	}
	if colorProfile == nil {
		colorProfile = func(float64) color.RGBA { return common.White }
	}

	curve := CatmullRomSpline(pts, 24)
	m := len(curve)
	if m < 2 {
		return
	}

	// Compute cumulative arc length so parameter t in [0, 1] is uniform in arc length.
	arcLen := make([]float64, m)
	totalLen := 0.0
	for i := 1; i < m; i++ {
		totalLen += curve[i].Dist(curve[i-1])
		arcLen[i] = totalLen
	}
	if totalLen < 1e-6 {
		return
	}

	left := make([]Point, m)
	right := make([]Point, m)
	tangents := make([]Point, m)
	colors := make([]color.RGBA, m)
	uniformColor := true
	firstCol := colorProfile(0)

	for i := 0; i < m; i++ {
		t := arcLen[i] / totalLen
		var tan Point
		switch {
		case i == 0:
			tan = curve[1].Sub(curve[0]).Norm()
		case i == m-1:
			tan = curve[m-1].Sub(curve[m-2]).Norm()
		default:
			tan = curve[i+1].Sub(curve[i-1]).Norm()
		}
		if tan.Len() < 1e-9 && i > 0 {
			tan = tangents[i-1]
		}
		tangents[i] = tan
		normal := tan.Perp()

		w := math.Max(0, maxWidth*widthProfile(t)) * 0.5
		left[i] = curve[i].Add(normal.Mul(w))
		right[i] = curve[i].Sub(normal.Mul(w))

		col := colorProfile(t)
		colors[i] = col
		if col != firstCol {
			uniformColor = false
		}
	}

	// Fast, seam-free single-contour rasterization when the stroke color is uniform.
	if uniformColor {
		path := &canvas.Path{}
		path.MoveTo(left[0].X, left[0].Y)
		for i := 1; i < m; i++ {
			path.LineTo(left[i].X, left[i].Y)
		}
		for i := m - 1; i >= 0; i-- {
			path.LineTo(right[i].X, right[i].Y)
		}
		path.Close()
		c.RasterizePath(path, firstCol, alpha)
		return
	}

	// Variable color along the stroke: rasterize sub-ribbons with tiny tangent overlap to prevent hairline seams.
	const chunkSize = 2
	for i := 0; i < m-1; i += chunkSize {
		end := i + chunkSize
		if end >= m {
			end = m - 1
		}
		midIdx := (i + end) / 2
		segColor := LerpOklab(colors[i], colors[end], 0.5)
		if midIdx > i && midIdx < end {
			segColor = colors[midIdx]
		}

		path := &canvas.Path{}
		path.MoveTo(left[i].X, left[i].Y)
		for k := i + 1; k <= end; k++ {
			if k == end && end < m-1 {
				// Extend slightly along tangent to eliminate sub-pixel gaps between adjacent color quads.
				ext := tangents[k].Mul(0.25)
				path.LineTo(left[k].X+ext.X, left[k].Y+ext.Y)
			} else {
				path.LineTo(left[k].X, left[k].Y)
			}
		}
		for k := end; k >= i; k-- {
			if k == end && end < m-1 {
				ext := tangents[k].Mul(0.25)
				path.LineTo(right[k].X+ext.X, right[k].Y+ext.Y)
			} else {
				path.LineTo(right[k].X, right[k].Y)
			}
		}
		path.Close()
		c.RasterizePath(path, segColor, alpha)
	}
}
