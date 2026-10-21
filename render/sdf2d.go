package render

import (
	"image/color"
	"math"

	"github.com/jdxyw/generativeart/common"
)

// SDF2D represents an analytical 2D Signed Distance Field: negative inside the shape,
// zero on the boundary, and positive outside.
type SDF2D func(p Point) float64

// SDFCircle returns the exact 2D SDF of a circle centered at center with given radius.
func SDFCircle(center Point, radius float64) SDF2D {
	return func(p Point) float64 {
		return p.Dist(center) - radius
	}
}

// SDFBox returns the exact 2D SDF of an axis-aligned rectangle centered at center with halfExtents.
func SDFBox(center, halfExtents Point) SDF2D {
	return func(p Point) float64 {
		dx := math.Abs(p.X-center.X) - halfExtents.X
		dy := math.Abs(p.Y-center.Y) - halfExtents.Y
		out := math.Hypot(math.Max(dx, 0), math.Max(dy, 0))
		in := math.Min(math.Max(dx, dy), 0)
		return out + in
	}
}

// SDFRoundedBox returns the exact 2D SDF of a rounded axis-aligned box.
func SDFRoundedBox(center, halfExtents Point, cornerRadius float64) SDF2D {
	hx := math.Max(0, halfExtents.X-cornerRadius)
	hy := math.Max(0, halfExtents.Y-cornerRadius)
	base := SDFBox(center, Pt(hx, hy))
	return func(p Point) float64 {
		return base(p) - cornerRadius
	}
}

// SDFSegment returns the exact 2D SDF of a round-capped capsule segment from a to b with radius r.
func SDFSegment(a, b Point, r float64) SDF2D {
	ba := b.Sub(a)
	denom := ba.Dot(ba)
	return func(p Point) float64 {
		pa := p.Sub(a)
		h := 0.0
		if denom > 1e-12 {
			h = common.Constrain(pa.Dot(ba)/denom, 0, 1)
		}
		return pa.Sub(ba.Mul(h)).Len() - r
	}
}

// SDFPolygon returns the exact 2D Signed Distance Field for an arbitrary closed polygon pts
// using edge projection and winding-number parity.
func SDFPolygon(pts []Point) SDF2D {
	n := len(pts)
	if n < 3 {
		return func(Point) float64 { return 1e9 }
	}
	cp := make([]Point, n)
	copy(cp, pts)

	return func(p Point) float64 {
		dSq := (p.X-cp[0].X)*(p.X-cp[0].X) + (p.Y-cp[0].Y)*(p.Y-cp[0].Y)
		sign := 1.0
		j := n - 1
		for i := 0; i < n; i++ {
			vi := cp[i]
			vj := cp[j]
			e := vj.Sub(vi)
			w := p.Sub(vi)
			ee := e.Dot(e)
			h := 0.0
			if ee > 1e-12 {
				h = common.Constrain(w.Dot(e)/ee, 0, 1)
			}
			b := w.Sub(e.Mul(h))
			bSq := b.Dot(b)
			if bSq < dSq {
				dSq = bSq
			}
			c1 := p.Y >= vi.Y
			c2 := p.Y < vj.Y
			c3 := e.X*w.Y > e.Y*w.X
			if (c1 && c2 && c3) || (!c1 && !c2 && !c3) {
				sign = -sign
			}
			j = i
		}
		return sign * math.Sqrt(dSq)
	}
}

// SDFUnion returns the exact union min(a(p), b(p)) of two 2D SDFs.
func SDFUnion(a, b SDF2D) SDF2D {
	return func(p Point) float64 {
		return math.Min(a(p), b(p))
	}
}

// SDFIntersect returns the exact intersection max(a(p), b(p)) of two 2D SDFs.
func SDFIntersect(a, b SDF2D) SDF2D {
	return func(p Point) float64 {
		return math.Max(a(p), b(p))
	}
}

// SDFSubtract subtracts cut from base: max(base(p), -cut(p)).
func SDFSubtract(base, cut SDF2D) SDF2D {
	return func(p Point) float64 {
		return math.Max(base(p), -cut(p))
	}
}

// SDFSmoothUnion returns a C1 polynomial smooth union of a and b with blend radius k > 0.
func SDFSmoothUnion(a, b SDF2D, k float64) SDF2D {
	if k <= 1e-6 {
		return SDFUnion(a, b)
	}
	return func(p Point) float64 {
		d1 := a(p)
		d2 := b(p)
		h := common.Constrain(0.5+0.5*(d2-d1)/k, 0, 1)
		return d2*(1.0-h) + d1*h - k*h*(1.0-h)
	}
}

// SDFSmoothSubtract returns a polynomial smooth subtraction of cut from base with radius k > 0.
func SDFSmoothSubtract(base, cut SDF2D, k float64) SDF2D {
	if k <= 1e-6 {
		return SDFSubtract(base, cut)
	}
	return func(p Point) float64 {
		d1 := cut(p)
		d2 := base(p)
		h := common.Constrain(0.5-0.5*(d2+d1)/k, 0, 1)
		return d2*(1.0-h) + (-d1)*h + k*h*(1.0-h)
	}
}

// SDFOnion converts a solid 2D SDF into an annular shell of half-thickness thickness.
func SDFOnion(sdf SDF2D, thickness float64) SDF2D {
	return func(p Point) float64 {
		return math.Abs(sdf(p)) - thickness
	}
}

// SDFGradient computes the unit outward gradient vector (normal) of sdf at p.
func SDFGradient(sdf SDF2D, p Point, eps float64) Point {
	if eps <= 0 {
		eps = 0.5
	}
	dx := sdf(Pt(p.X+eps, p.Y)) - sdf(Pt(p.X-eps, p.Y))
	dy := sdf(Pt(p.X, p.Y+eps)) - sdf(Pt(p.X, p.Y-eps))
	return Pt(dx, dy).Norm()
}

// RenderSDFDropShadow renders an analytical soft penumbra drop shadow for sdf shifted by offset.
func (c *Canvas) RenderSDFDropShadow(
	sdf SDF2D,
	offset Point,
	blurRadius, spread float64,
	shadowCol color.RGBA,
	alpha float64,
) {
	if c == nil || sdf == nil || alpha <= 0 || shadowCol.A == 0 {
		return
	}
	if blurRadius < 0.5 {
		blurRadius = 0.5
	}
	rLin := srgb8ToLinearLUT[shadowCol.R]
	gLin := srgb8ToLinearLUT[shadowCol.G]
	bLin := srgb8ToLinearLUT[shadowCol.B]
	baseAlpha := (float64(shadowCol.A) / 255.0) * common.Constrain(alpha, 0, 1)

	parallelRows(c.Width, c.Height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			fy := float64(y) + 0.5 - offset.Y
			for x := 0; x < c.Width; x++ {
				fx := float64(x) + 0.5 - offset.X
				d := sdf(Pt(fx, fy)) - spread
				if d >= blurRadius {
					continue
				}
				u := common.Constrain((blurRadius-d)/(2.0*blurRadius), 0, 1)
				w := u * u * (3.0 - 2.0*u)
				if w > 1e-4 {
					c.BlendPixelLinear(x, y, rLin, gLin, bLin, baseAlpha*w, c.blendMode)
				}
			}
		}
	})
}

// RenderSDFBevel renders a 2D SDF with 3D beveled edge lighting, interior meniscus depth,
// and directional specular highlights in Oklab space.
func (c *Canvas) RenderSDFBevel(
	sdf SDF2D,
	baseCol, rimCol, specCol color.RGBA,
	bevelWidth float64,
	lightDir Point,
	alpha float64,
) {
	if c == nil || sdf == nil || alpha <= 0 {
		return
	}
	if bevelWidth <= 0.5 {
		bevelWidth = 6.0
	}
	if lightDir.Len() < 1e-6 {
		lightDir = Pt(-0.707, -0.707)
	} else {
		lightDir = lightDir.Norm()
	}

	parallelRows(c.Width, c.Height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			fy := float64(y) + 0.5
			for x := 0; x < c.Width; x++ {
				fx := float64(x) + 0.5
				p := Pt(fx, fy)
				d := sdf(p)
				if d >= 0.75 {
					continue
				}
				// Anti-aliased boundary coverage
				cov := 1.0
				if d > -0.75 {
					cov = common.Constrain((0.75-d)/1.5, 0, 1)
				}

				depth := common.Constrain(-d/bevelWidth, 0, 1)
				bevelFactor := 1.0 - depth*depth*(3.0-2.0*depth)
				col := LerpOklab(baseCol, rimCol, bevelFactor*0.75)

				if bevelFactor > 0.02 {
					n := SDFGradient(sdf, p, 0.75)
					ndotl := -(n.X*lightDir.X + n.Y*lightDir.Y)
					if ndotl > 0 {
						spec := math.Pow(ndotl, 2.2) * bevelFactor
						col = LerpOklab(col, specCol, common.Constrain(spec*0.85, 0, 1))
					} else {
						shade := (-ndotl) * bevelFactor * 0.55
						col = LerpOklab(col, rimCol, common.Constrain(shade, 0, 1))
					}
				}
				c.BlendPixel(x, y, col, alpha*cov)
			}
		}
	})
}
