package render

import (
	"math"

	"github.com/jdxyw/generativeart/common"
	"github.com/tdewolff/canvas"
)

// ============================================================================
// 1. 2D Affine Transform Matrix Stack
// ============================================================================

func isIdentityMatrix(m canvas.Matrix) bool {
	return math.Abs(m[0][0]-1.0) < 1e-9 &&
		math.Abs(m[0][1]) < 1e-9 &&
		math.Abs(m[0][2]) < 1e-9 &&
		math.Abs(m[1][0]) < 1e-9 &&
		math.Abs(m[1][1]-1.0) < 1e-9 &&
		math.Abs(m[1][2]) < 1e-9
}

// ResetTransform resets the active 2D affine transformation to the identity matrix
// and clears the transform stack.
func (c *Canvas) ResetTransform() {
	if c == nil {
		return
	}
	c.curMatrix = canvas.Identity
	c.hasCustomTransform = false
	c.transformStack = c.transformStack[:0]
}

// PushMatrix saves the current 2D affine transformation matrix onto the stack.
func (c *Canvas) PushMatrix() {
	if c == nil {
		return
	}
	if !c.hasCustomTransform && len(c.transformStack) == 0 {
		c.curMatrix = canvas.Identity
	}
	c.transformStack = append(c.transformStack, c.curMatrix)
}

// PopMatrix restores the most recently pushed 2D affine transformation matrix.
func (c *Canvas) PopMatrix() {
	if c == nil || len(c.transformStack) == 0 {
		return
	}
	n := len(c.transformStack) - 1
	c.curMatrix = c.transformStack[n]
	c.transformStack = c.transformStack[:n]
	c.hasCustomTransform = !isIdentityMatrix(c.curMatrix)
}

// CurrentMatrix returns the active 2D affine transformation matrix.
func (c *Canvas) CurrentMatrix() canvas.Matrix {
	if c == nil || (!c.hasCustomTransform && len(c.transformStack) == 0) {
		return canvas.Identity
	}
	return c.curMatrix
}

func (c *Canvas) applyMatrix(m canvas.Matrix) {
	if !c.hasCustomTransform && len(c.transformStack) == 0 {
		c.curMatrix = canvas.Identity
	}
	c.curMatrix = c.curMatrix.Mul(m)
	c.hasCustomTransform = !isIdentityMatrix(c.curMatrix)
}

// Translate applies a translation (tx, ty) to the current transformation matrix.
func (c *Canvas) Translate(tx, ty float64) {
	if c == nil {
		return
	}
	c.applyMatrix(canvas.Identity.Translate(tx, ty))
}

// Rotate applies a rotation by radians around the origin (0, 0).
// Note: tdewolff/canvas Rotate takes degrees.
func (c *Canvas) Rotate(radians float64) {
	if c == nil {
		return
	}
	deg := radians * 180.0 / math.Pi
	c.applyMatrix(canvas.Identity.Rotate(deg))
}

// RotateAbout applies a rotation by radians around pivot (cx, cy).
func (c *Canvas) RotateAbout(radians, cx, cy float64) {
	if c == nil {
		return
	}
	deg := radians * 180.0 / math.Pi
	c.applyMatrix(canvas.Identity.RotateAbout(deg, cx, cy))
}

// Scale applies a scaling (sx, sy) around the origin (0, 0).
func (c *Canvas) Scale(sx, sy float64) {
	if c == nil {
		return
	}
	c.applyMatrix(canvas.Identity.Scale(sx, sy))
}

// ScaleAbout applies a scaling (sx, sy) around pivot (cx, cy).
func (c *Canvas) ScaleAbout(sx, sy, cx, cy float64) {
	if c == nil {
		return
	}
	c.applyMatrix(canvas.Identity.ScaleAbout(sx, sy, cx, cy))
}

// TransformPoint maps point p through the active 2D affine transformation matrix.
func (c *Canvas) TransformPoint(p Point) Point {
	if c == nil || !c.hasCustomTransform {
		return p
	}
	tp := c.curMatrix.Dot(canvas.Point{X: p.X, Y: p.Y})
	return Pt(tp.X, tp.Y)
}

// TransformScale returns the isotropic scale factor sqrt(|det(M)|) of the active matrix.
func (c *Canvas) TransformScale() float64 {
	if c == nil || !c.hasCustomTransform {
		return 1.0
	}
	det := math.Abs(c.curMatrix[0][0]*c.curMatrix[1][1] - c.curMatrix[0][1]*c.curMatrix[1][0])
	if det <= 1e-12 {
		return 1.0
	}
	return math.Sqrt(det)
}

// ============================================================================
// 2. Sub-Pixel Stencil / Clip Mask Stack
// ============================================================================

// ClipDepth returns the current depth of the active stencil/clip mask stack.
func (c *Canvas) ClipDepth() int {
	if c == nil {
		return 0
	}
	return len(c.clipStack)
}

// PopClip removes the top stencil/clip mask from the stack.
func (c *Canvas) PopClip() {
	if c == nil || len(c.clipStack) == 0 {
		return
	}
	c.clipStack = c.clipStack[:len(c.clipStack)-1]
}

// ClearClips removes all stencil/clip masks from the canvas.
func (c *Canvas) ClearClips() {
	if c == nil {
		return
	}
	c.clipStack = c.clipStack[:0]
}

func (c *Canvas) pushRawClipMask(newMask []uint8) {
	if len(c.clipStack) > 0 {
		prev := c.clipStack[len(c.clipStack)-1]
		for i := range newMask {
			newMask[i] = uint8((uint16(newMask[i]) * uint16(prev[i])) / 255)
		}
	}
	c.clipStack = append(c.clipStack, newMask)
}

// PushClipPath rasterizes path p (respecting the active transform matrix) into an
// 8-bit sub-pixel anti-aliased stencil mask and pushes it onto the clip stack.
func (c *Canvas) PushClipPath(p *canvas.Path) {
	if c == nil || c.Width <= 0 || c.Height <= 0 || p == nil {
		return
	}
	if c.hasCustomTransform {
		p = p.Transform(c.curMatrix)
	}
	fullMask := make([]uint8, c.Width*c.Height)
	pix, minX, minY, bw, bh, release := rasterizePathToMask(p, c.Width, c.Height)
	defer release()

	if pix != nil && bw > 0 && bh > 0 {
		for y := 0; y < bh; y++ {
			srcOff := y * bw
			dstOff := (minY+y)*c.Width + minX
			copy(fullMask[dstOff:dstOff+bw], pix[srcOff:srcOff+bw])
		}
	}
	c.pushRawClipMask(fullMask)
}

// PushClipPolygon pushes a closed polygonal clip mask onto the clip stack.
func (c *Canvas) PushClipPolygon(pts []Point) {
	if c == nil || len(pts) < 3 {
		return
	}
	p := &canvas.Path{}
	p.MoveTo(pts[0].X, pts[0].Y)
	for i := 1; i < len(pts); i++ {
		p.LineTo(pts[i].X, pts[i].Y)
	}
	p.Close()
	c.PushClipPath(p)
}

// PushClipCircle pushes an anti-aliased circular clip mask centered at (cx, cy) with radius r.
func (c *Canvas) PushClipCircle(cx, cy, r float64) {
	c.PushClipSDF(SDFCircle(Pt(cx, cy), r), false)
}

// PushClipSDF evaluates a 2D Signed Distance Field across the canvas with 1.5px smoothstep
// anti-aliasing and pushes it onto the clip stack. If invert is true, only pixels outside
// the SDF (d > 0) are kept.
func (c *Canvas) PushClipSDF(sdf SDF2D, invert bool) {
	if c == nil || c.Width <= 0 || c.Height <= 0 || sdf == nil {
		return
	}
	fullMask := make([]uint8, c.Width*c.Height)
	parallelRows(c.Width, c.Height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			fy := float64(y) + 0.5
			rowOff := y * c.Width
			for x := 0; x < c.Width; x++ {
				fx := float64(x) + 0.5
				d := sdf(Pt(fx, fy))
				if invert {
					d = -d
				}
				if d >= 0.75 {
					continue
				}
				cov := 1.0
				if d > -0.75 {
					t := common.Constrain((0.75-d)/1.5, 0, 1)
					cov = t * t * (3.0 - 2.0*t)
				}
				fullMask[rowOff+x] = uint8(math.Round(cov * 255.0))
			}
		}
	})
	c.pushRawClipMask(fullMask)
}

// ============================================================================
// 3. Independent HDR Layer Compositing
// ============================================================================

// NewLayer allocates a transparent child HDR Canvas with the same dimensions as c.
func (c *Canvas) NewLayer() *Canvas {
	if c == nil {
		return NewCanvas(0, 0)
	}
	return NewCanvas(c.Width, c.Height)
}

// CompositeLayer blends a child HDR layer into c using the specified BlendMode and opacity,
// respecting c's active clip mask stack and merging recorded vector operations.
func (c *Canvas) CompositeLayer(layer *Canvas, mode BlendMode, opacity float64) {
	if c == nil || layer == nil || opacity <= 0 {
		return
	}
	opacity = common.Constrain(opacity, 0, 1)
	w := c.Width
	if layer.Width < w {
		w = layer.Width
	}
	h := c.Height
	if layer.Height < h {
		h = layer.Height
	}

	parallelRows(w, h, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := 0; x < w; x++ {
				sIdx := (y*layer.Width + x) * 4
				sr := layer.pix[sIdx+0]
				sg := layer.pix[sIdx+1]
				sb := layer.pix[sIdx+2]
				sa := layer.pix[sIdx+3] * opacity
				if sa <= 1e-9 {
					continue
				}
				c.BlendPixelLinear(x, y, sr, sg, sb, sa, mode)
			}
		}
	})

	if c.recordVector && len(layer.vecOps) > 0 {
		for _, op := range layer.vecOps {
			c.vecOps = append(c.vecOps, vectorOp{
				path:  op.path.Copy(),
				col:   op.col,
				alpha: op.alpha * opacity,
			})
		}
	}
}
