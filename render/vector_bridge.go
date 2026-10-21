package render

import (
	"image"
	"image/color"
	"image/draw"
	"io"
	"math"
	"os"

	"github.com/jdxyw/generativeart/common"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/pdf"
	"github.com/tdewolff/canvas/renderers/rasterizer"
	"github.com/tdewolff/canvas/renderers/svg"
	"golang.org/x/image/vector"
)

type vectorOp struct {
	path  *canvas.Path
	col   color.RGBA
	alpha float64
}

// SetRecordVector enables or disables recording of vector paths for SVG/PDF export.
func (c *Canvas) SetRecordVector(enabled bool) {
	if c != nil {
		c.recordVector = enabled
	}
}

// VectorOpCount returns the number of vector path operations recorded on this Canvas.
func (c *Canvas) VectorOpCount() int {
	if c == nil {
		return 0
	}
	return len(c.vecOps)
}

// RasterizePath takes a tdewolff/canvas Path (in pixel coordinates), converts its segments
// (MoveTo, LineTo, QuadTo, CubeTo, ArcTo, Close) to golang.org/x/image/vector.Rasterizer,
// and composites the anti-aliased alpha coverage mask directly into the HDR linear-RGB Canvas
// using the active BlendMode.
func (c *Canvas) RasterizePath(p *canvas.Path, col color.RGBA, alpha float64) {
	if c == nil || p == nil || p.Empty() || alpha <= 0 || col.A == 0 || c.Width <= 0 || c.Height <= 0 {
		return
	}

	if c.recordVector {
		c.vecOps = append(c.vecOps, vectorOp{
			path:  p.Copy(),
			col:   col,
			alpha: common.Constrain(alpha, 0, 1),
		})
	}

	// Replace elliptical arcs with exact cubic Bézier approximations so all segments
	// map cleanly to MoveTo, LineTo, QuadTo, CubeTo, and Close.
	pathNoArcs := p.ReplaceArcs()
	bounds := pathNoArcs.FastBounds()

	minX := int(math.Floor(bounds.X0)) - 2
	minY := int(math.Floor(bounds.Y0)) - 2
	maxX := int(math.Ceil(bounds.X1)) + 2
	maxY := int(math.Ceil(bounds.Y1)) + 2

	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX > c.Width {
		maxX = c.Width
	}
	if maxY > c.Height {
		maxY = c.Height
	}
	if minX >= maxX || minY >= maxY {
		return
	}

	bw := maxX - minX
	bh := maxY - minY

	if c.ras == nil {
		c.ras = vector.NewRasterizer(bw, bh)
	} else {
		c.ras.Reset(bw, bh)
	}
	c.ras.DrawOp = draw.Src

	ox := float32(minX)
	oy := float32(minY)

	d := pathNoArcs.Data()
	subpathOpen := false
	for i := 0; i < len(d); {
		cmd := d[i]
		switch cmd {
		case canvas.MoveToCmd:
			if subpathOpen {
				c.ras.ClosePath()
			}
			c.ras.MoveTo(float32(d[i+1])-ox, float32(d[i+2])-oy)
			subpathOpen = true
			i += 4
		case canvas.LineToCmd:
			c.ras.LineTo(float32(d[i+1])-ox, float32(d[i+2])-oy)
			subpathOpen = true
			i += 4
		case canvas.QuadToCmd:
			c.ras.QuadTo(
				float32(d[i+1])-ox, float32(d[i+2])-oy,
				float32(d[i+3])-ox, float32(d[i+4])-oy,
			)
			subpathOpen = true
			i += 6
		case canvas.CubeToCmd:
			c.ras.CubeTo(
				float32(d[i+1])-ox, float32(d[i+2])-oy,
				float32(d[i+3])-ox, float32(d[i+4])-oy,
				float32(d[i+5])-ox, float32(d[i+6])-oy,
			)
			subpathOpen = true
			i += 8
		case canvas.CloseCmd:
			c.ras.LineTo(float32(d[i+1])-ox, float32(d[i+2])-oy)
			c.ras.ClosePath()
			subpathOpen = false
			i += 4
		default:
			// Fallback should never be reached because ReplaceArcs replaced ArcToCmd.
			i += 8
		}
	}
	if subpathOpen {
		c.ras.ClosePath()
	}

	needed := bw * bh
	if c.maskBuf == nil || cap(c.maskBuf.Pix) < needed {
		c.maskBuf = &image.Alpha{
			Pix:    make([]uint8, needed),
			Stride: bw,
			Rect:   image.Rect(0, 0, bw, bh),
		}
	} else {
		c.maskBuf.Pix = c.maskBuf.Pix[:needed]
		c.maskBuf.Stride = bw
		c.maskBuf.Rect = image.Rect(0, 0, bw, bh)
	}

	c.ras.Draw(c.maskBuf, c.maskBuf.Rect, image.Opaque, image.Point{})

	rLin := srgb8ToLinearLUT[col.R]
	gLin := srgb8ToLinearLUT[col.G]
	bLin := srgb8ToLinearLUT[col.B]
	baseAlpha := (float64(col.A) / 255.0) * common.Constrain(alpha, 0, 1)

	pix := c.maskBuf.Pix
	for y := 0; y < bh; y++ {
		rowOff := y * bw
		dstY := minY + y
		for x := 0; x < bw; x++ {
			m := pix[rowOff+x]
			if m == 0 {
				continue
			}
			cov := (float64(m) / 255.0) * baseAlpha
			c.BlendPixelLinear(minX+x, dstY, rLin, gLin, bLin, cov, c.blendMode)
		}
	}
}

func strokeSingleSubpathSafe(sp *canvas.Path, width float64, capper canvas.Capper, joiner canvas.Joiner) (res *canvas.Path) {
	defer func() {
		if r := recover(); r != nil {
			prev := canvas.FastStroke
			canvas.FastStroke = true
			res = sp.Stroke(width, capper, joiner, canvas.PixelTolerance)
			canvas.FastStroke = prev
		}
	}()
	return sp.Stroke(width, capper, joiner, canvas.PixelTolerance)
}

// StrokePath uses tdewolff/canvas's exact geometric stroke offsetting p.Stroke(width, capper, joiner, tolerance)
// and rasterizes the resulting contour with sub-pixel anti-aliasing into the HDR Canvas.
func (c *Canvas) StrokePath(p *canvas.Path, width float64, capper canvas.Capper, joiner canvas.Joiner, col color.RGBA, alpha float64) {
	if c == nil || p == nil || p.Empty() || width <= 0 || alpha <= 0 || col.A == 0 {
		return
	}
	if capper == nil {
		capper = canvas.RoundCap
	}
	if joiner == nil {
		joiner = canvas.RoundJoin
	}
	subpaths := p.Split()
	if len(subpaths) <= 1 {
		stroked := strokeSingleSubpathSafe(p, width, capper, joiner)
		c.RasterizePath(stroked, col, alpha)
		return
	}
	combined := &canvas.Path{}
	for _, sp := range subpaths {
		if sp.Empty() {
			continue
		}
		stroked := strokeSingleSubpathSafe(sp, width, capper, joiner)
		if stroked != nil && !stroked.Empty() {
			combined = combined.Append(stroked)
		}
	}
	c.RasterizePath(combined, col, alpha)
}

// FillPolygon fills a closed polygon specified by pts with sub-pixel anti-aliasing.
func (c *Canvas) FillPolygon(pts []Point, col color.RGBA, alpha float64) {
	if len(pts) < 3 || alpha <= 0 || col.A == 0 {
		return
	}
	p := &canvas.Path{}
	p.MoveTo(pts[0].X, pts[0].Y)
	for i := 1; i < len(pts); i++ {
		p.LineTo(pts[i].X, pts[i].Y)
	}
	p.Close()
	c.RasterizePath(p, col, alpha)
}

// StrokePolyline strokes a polyline specified by pts with round caps and round joins.
func (c *Canvas) StrokePolyline(pts []Point, width float64, col color.RGBA, alpha float64) {
	if len(pts) < 2 || width <= 0 || alpha <= 0 || col.A == 0 {
		return
	}
	p := &canvas.Path{}
	p.MoveTo(pts[0].X, pts[0].Y)
	for i := 1; i < len(pts); i++ {
		p.LineTo(pts[i].X, pts[i].Y)
	}
	c.StrokePath(p, width, canvas.RoundCap, canvas.RoundJoin, col, alpha)
}

// CompositeTdewolffCanvas renders a tdewolff/canvas.Canvas via its rasterizer package
// and composites it onto the HDR Canvas using the active BlendMode.
func (c *Canvas) CompositeTdewolffCanvas(tc *canvas.Canvas, resolution canvas.Resolution, alpha float64) {
	if c == nil || tc == nil || alpha <= 0 {
		return
	}
	img := rasterizer.Draw(tc, resolution, canvas.DefaultColorSpace)
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	for y := 0; y < h && y < c.Height; y++ {
		for x := 0; x < w && x < c.Width; x++ {
			rgba := img.RGBAAt(b.Min.X+x, b.Min.Y+y)
			if rgba.A == 0 {
				continue
			}
			a := (float64(rgba.A) / 255.0) * common.Constrain(alpha, 0, 1)
			rS := common.Constrain(float64(rgba.R)/float64(rgba.A), 0, 1)
			gS := common.Constrain(float64(rgba.G)/float64(rgba.A), 0, 1)
			bS := common.Constrain(float64(rgba.B)/float64(rgba.A), 0, 1)
			c.BlendPixelLinear(x, y, SRGBToLinear(rS), SRGBToLinear(gS), SRGBToLinear(bS), a, c.blendMode)
		}
	}
}

// BuildVectorCanvas constructs an infinite-resolution tdewolff/canvas.Canvas scene graph
// from the recorded vector operations. If no vector paths were recorded (e.g. pure pixel-splat
// attractors), it embeds the resolved HDR raster image so SVG/PDF export always succeeds.
func (c *Canvas) BuildVectorCanvas() *canvas.Canvas {
	if c == nil || c.Width <= 0 || c.Height <= 0 {
		return canvas.New(1, 1)
	}
	w := float64(c.Width)
	h := float64(c.Height)
	tc := canvas.New(w, h)
	ctx := canvas.NewContext(tc)

	if len(c.vecOps) == 0 {
		img := c.ResolveToRGBA(nil)
		ctx.DrawImage(0, 0, img, canvas.DPMM(1.0))
		return tc
	}

	if c.hasBg {
		ctx.SetFillColor(c.bgCol)
		ctx.SetStrokeColor(color.Transparent)
		ctx.DrawPath(0, 0, canvas.Rectangle(w, h))
	}

	// Flip Y from top-left screen coordinates to bottom-left tdewolff/canvas coordinates: y' = h - y
	flipY := canvas.Identity.Translate(0, h).Scale(1, -1)
	for _, op := range c.vecOps {
		if op.path == nil || op.path.Empty() {
			continue
		}
		effA := uint8(math.Round(float64(op.col.A) * common.Constrain(op.alpha, 0, 1)))
		if effA == 0 {
			continue
		}
		nrgba := color.NRGBA{R: op.col.R, G: op.col.G, B: op.col.B, A: effA}
		ctx.SetFillColor(nrgba)
		ctx.SetStrokeColor(color.Transparent)
		ctx.DrawPath(0, 0, op.path.Transform(flipY))
	}
	return tc
}

// WriteSVG writes the recorded vector scene graph (or embedded raster fallback) as SVG to w.
func (c *Canvas) WriteSVG(w io.Writer) error {
	tc := c.BuildVectorCanvas()
	s := svg.New(w, tc.W, tc.H, nil)
	tc.RenderTo(s)
	return s.Close()
}

// SaveSVG writes the recorded vector scene graph as an SVG file at filename.
func (c *Canvas) SaveSVG(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	if err := c.WriteSVG(f); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// WritePDF writes the recorded vector scene graph (or embedded raster fallback) as PDF to w.
func (c *Canvas) WritePDF(w io.Writer) error {
	tc := c.BuildVectorCanvas()
	p := pdf.New(w, tc.W, tc.H, nil)
	tc.RenderTo(p)
	return p.Close()
}

// SavePDF writes the recorded vector scene graph as a PDF file at filename.
func (c *Canvas) SavePDF(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	if err := c.WritePDF(f); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

