package render

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"math"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/jdxyw/generativeart/common"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/pdf"
	"github.com/tdewolff/canvas/renderers/rasterizer"
	"github.com/tdewolff/canvas/renderers/svg"
	"golang.org/x/image/vector"
)

type rasterWorkspace struct {
	ras     *vector.Rasterizer
	maskBuf *image.Alpha
}

var rasterWorkspacePool = sync.Pool{
	New: func() interface{} {
		return &rasterWorkspace{
			ras: vector.NewRasterizer(64, 64),
			maskBuf: &image.Alpha{
				Pix:    make([]uint8, 64*64),
				Stride: 64,
				Rect:   image.Rect(0, 0, 64, 64),
			},
		}
	},
}

// rasterizePathToMask rasterizes path p (in pixel coordinates) into a pooled alpha mask
// clipped to [0, canvasW) x [0, canvasH). Call release() when finished reading mask.
func rasterizePathToMask(p *canvas.Path, canvasW, canvasH int) (mask []uint8, minX, minY, bw, bh int, release func()) {
	noop := func() {}
	if p == nil || p.Empty() || canvasW <= 0 || canvasH <= 0 {
		return nil, 0, 0, 0, 0, noop
	}
	pathNoArcs := p.ReplaceArcs()
	bounds := pathNoArcs.FastBounds()

	minX = int(math.Floor(bounds.X0)) - 2
	minY = int(math.Floor(bounds.Y0)) - 2
	maxX := int(math.Ceil(bounds.X1)) + 2
	maxY := int(math.Ceil(bounds.Y1)) + 2

	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX > canvasW {
		maxX = canvasW
	}
	if maxY > canvasH {
		maxY = canvasH
	}
	if minX >= maxX || minY >= maxY {
		return nil, 0, 0, 0, 0, noop
	}

	bw = maxX - minX
	bh = maxY - minY
	ws := rasterWorkspacePool.Get().(*rasterWorkspace)
	ws.ras.Reset(bw, bh)
	ws.ras.DrawOp = draw.Src

	ox := float32(minX)
	oy := float32(minY)

	d := pathNoArcs.Data()
	subpathOpen := false
	for i := 0; i < len(d); {
		cmd := d[i]
		switch cmd {
		case canvas.MoveToCmd:
			if subpathOpen {
				ws.ras.ClosePath()
			}
			ws.ras.MoveTo(float32(d[i+1])-ox, float32(d[i+2])-oy)
			subpathOpen = true
			i += 4
		case canvas.LineToCmd:
			ws.ras.LineTo(float32(d[i+1])-ox, float32(d[i+2])-oy)
			subpathOpen = true
			i += 4
		case canvas.QuadToCmd:
			ws.ras.QuadTo(
				float32(d[i+1])-ox, float32(d[i+2])-oy,
				float32(d[i+3])-ox, float32(d[i+4])-oy,
			)
			subpathOpen = true
			i += 6
		case canvas.CubeToCmd:
			ws.ras.CubeTo(
				float32(d[i+1])-ox, float32(d[i+2])-oy,
				float32(d[i+3])-ox, float32(d[i+4])-oy,
				float32(d[i+5])-ox, float32(d[i+6])-oy,
			)
			subpathOpen = true
			i += 8
		case canvas.CloseCmd:
			ws.ras.LineTo(float32(d[i+1])-ox, float32(d[i+2])-oy)
			ws.ras.ClosePath()
			subpathOpen = false
			i += 4
		default:
			i += 8
		}
	}
	if subpathOpen {
		ws.ras.ClosePath()
	}

	needed := bw * bh
	if cap(ws.maskBuf.Pix) < needed {
		ws.maskBuf.Pix = make([]uint8, needed)
	} else {
		ws.maskBuf.Pix = ws.maskBuf.Pix[:needed]
	}
	ws.maskBuf.Stride = bw
	ws.maskBuf.Rect = image.Rect(0, 0, bw, bh)

	ws.ras.Draw(ws.maskBuf, ws.maskBuf.Rect, image.Opaque, image.Point{})
	return ws.maskBuf.Pix, minX, minY, bw, bh, func() {
		rasterWorkspacePool.Put(ws)
	}
}

type vectorOp struct {
	path       *canvas.Path
	centerline *canvas.Path
	strokeW    float64
	col        color.RGBA
	alpha      float64
}

// pointInPolygonRings evaluates the even-odd ray-casting rule across a set of closed polygon rings.
func pointInPolygonRings(pt Point, rings [][]Point) bool {
	inside := false
	for _, ring := range rings {
		n := len(ring)
		if n < 3 {
			continue
		}
		j := n - 1
		for i := 0; i < n; i++ {
			vi := ring[i]
			vj := ring[j]
			if ((vi.Y > pt.Y) != (vj.Y > pt.Y)) &&
				(pt.X < (vj.X-vi.X)*(pt.Y-vi.Y)/(vj.Y-vi.Y+1e-18)+vi.X) {
				inside = !inside
			}
			j = i
		}
	}
	return inside
}

func pointInPolyStack(pt Point, stack [][][]Point) bool {
	for _, rings := range stack {
		if len(rings) > 0 && !pointInPolygonRings(pt, rings) {
			return false
		}
	}
	return true
}

// segmentIntersectT returns the parameter t in (0, 1) along segment AB where it intersects segment CD.
func segmentIntersectT(a, b, c, d Point) (float64, bool) {
	r := b.Sub(a)
	s := d.Sub(c)
	rxs := r.Cross(s)
	if math.Abs(rxs) < 1e-12 {
		return 0, false
	}
	qp := c.Sub(a)
	t := qp.Cross(s) / rxs
	u := qp.Cross(r) / rxs
	if t > 1e-7 && t < 1.0-1e-7 && u >= -1e-7 && u <= 1.0+1e-7 {
		return t, true
	}
	return 0, false
}

// clipOpenPathAgainstPolyStack trims an open or closed vector path against every active clip
// polygon level in stack by splitting segments at exact boundary intersections and retaining
// sub-segments whose midpoints lie inside all clip levels.
func clipOpenPathAgainstPolyStack(p *canvas.Path, stack [][][]Point) *canvas.Path {
	if p == nil || p.Empty() || len(stack) == 0 {
		return p
	}
	flat := p.ReplaceArcs().Flatten(0.45)
	out := &canvas.Path{}

	for _, sp := range flat.Split() {
		d := sp.Data()
		var pts []Point
		for i := 0; i < len(d); {
			switch d[i] {
			case canvas.MoveToCmd, canvas.LineToCmd, canvas.CloseCmd:
				pts = append(pts, Pt(d[i+1], d[i+2]))
				i += 4
			case canvas.QuadToCmd:
				pts = append(pts, Pt(d[i+3], d[i+4]))
				i += 6
			case canvas.CubeToCmd:
				pts = append(pts, Pt(d[i+5], d[i+6]))
				i += 8
			default:
				i += 4
			}
		}
		if len(pts) < 2 {
			continue
		}

		penDown := false
		var lastPt Point
		for idx := 0; idx < len(pts)-1; idx++ {
			a := pts[idx]
			b := pts[idx+1]
			if a.Dist(b) < 1e-6 {
				continue
			}

			ts := []float64{0.0, 1.0}
			for _, rings := range stack {
				for _, ring := range rings {
					m := len(ring)
					if m < 3 {
						continue
					}
					for e := 0; e < m; e++ {
						cPt := ring[e]
						dPt := ring[(e+1)%m]
						if t, ok := segmentIntersectT(a, b, cPt, dPt); ok {
							ts = append(ts, t)
						}
					}
				}
			}
			sort.Float64s(ts)

			for k := 0; k < len(ts)-1; k++ {
				t0, t1 := ts[k], ts[k+1]
				if t1-t0 < 1e-6 {
					continue
				}
				mid := a.Lerp(b, 0.5*(t0+t1))
				if pointInPolyStack(mid, stack) {
					p0 := a.Lerp(b, t0)
					p1 := a.Lerp(b, t1)
					if !penDown || lastPt.Dist(p0) > 1e-3 {
						out.MoveTo(p0.X, p0.Y)
					}
					out.LineTo(p1.X, p1.Y)
					lastPt = p1
					penDown = true
				} else {
					penDown = false
				}
			}
		}
	}
	return out
}

func ringsToCanvasPath(rings [][]Point) *canvas.Path {
	cp := &canvas.Path{}
	for _, ring := range rings {
		if len(ring) < 3 {
			continue
		}
		cp.MoveTo(ring[0].X, ring[0].Y)
		for i := 1; i < len(ring); i++ {
			cp.LineTo(ring[i].X, ring[i].Y)
		}
		cp.Close()
	}
	return cp
}

func safePathAnd(subj, clip *canvas.Path) (res *canvas.Path, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			res = nil
			ok = false
		}
	}()
	res = subj.And(clip)
	return res, true
}

// clipFilledPathAgainstPolyStack clips a filled closed vector path against every active clip
// level in stack using exact boolean path intersection with polygon-clipping fallback.
func clipFilledPathAgainstPolyStack(p *canvas.Path, stack [][][]Point) *canvas.Path {
	if p == nil || p.Empty() || len(stack) == 0 {
		return p
	}
	cur := p
	for _, rings := range stack {
		if len(rings) == 0 {
			continue
		}
		clipPath := ringsToCanvasPath(rings)
		if clipPath.Empty() {
			continue
		}
		if andPath, ok := safePathAnd(cur, clipPath); ok && andPath != nil {
			cur = andPath
			if cur.Empty() {
				return cur
			}
			continue
		}
		// Fallback to segment-intersection clipping if boolean Path.And recovered from a degenerate curve
		cur = clipOpenPathAgainstPolyStack(cur, [][][]Point{rings})
		if cur == nil || cur.Empty() {
			return cur
		}
	}
	return cur
}

// ClipPolylineToPolygon clips an open polyline pts against a closed polygon clipPoly
// and returns the list of clipped polyline sub-segments that lie inside clipPoly.
func ClipPolylineToPolygon(pts, clipPoly []Point) [][]Point {
	if len(pts) < 2 || len(clipPoly) < 3 {
		return nil
	}
	p := &canvas.Path{}
	p.MoveTo(pts[0].X, pts[0].Y)
	for i := 1; i < len(pts); i++ {
		p.LineTo(pts[i].X, pts[i].Y)
	}
	clipped := clipOpenPathAgainstPolyStack(p, [][][]Point{{clipPoly}})
	if clipped == nil || clipped.Empty() {
		return nil
	}
	var out [][]Point
	for _, sp := range clipped.Split() {
		d := sp.Data()
		var seg []Point
		for i := 0; i < len(d); i += 4 {
			seg = append(seg, Pt(d[i+1], d[i+2]))
		}
		if len(seg) >= 2 {
			out = append(out, seg)
		}
	}
	return out
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
	c.rasterizePathInternal(p, nil, 0, col, alpha)
}

func (c *Canvas) rasterizePathInternal(p, centerline *canvas.Path, strokeW float64, col color.RGBA, alpha float64) {
	if c == nil || p == nil || p.Empty() || alpha <= 0 || col.A == 0 || c.Width <= 0 || c.Height <= 0 {
		return
	}

	if c.hasCustomTransform {
		p = p.Transform(c.curMatrix)
		if centerline != nil {
			centerline = centerline.Transform(c.curMatrix)
			strokeW *= c.TransformScale()
		}
	}

	if c.recordVector {
		vecP := p.Copy()
		var vecCL *canvas.Path
		if centerline != nil && !centerline.Empty() {
			vecCL = centerline.Copy()
		}
		if len(c.clipPolyStack) > 0 {
			vecP = clipFilledPathAgainstPolyStack(vecP, c.clipPolyStack)
			if vecCL != nil {
				vecCL = clipOpenPathAgainstPolyStack(vecCL, c.clipPolyStack)
			}
		}
		if (vecP != nil && !vecP.Empty()) || (vecCL != nil && !vecCL.Empty()) {
			c.vecOps = append(c.vecOps, vectorOp{
				path:       vecP,
				centerline: vecCL,
				strokeW:    strokeW,
				col:        col,
				alpha:      common.Constrain(alpha, 0, 1),
			})
		}
	}

	pix, minX, minY, bw, bh, release := rasterizePathToMask(p, c.Width, c.Height)
	defer release()
	if pix == nil || bw <= 0 || bh <= 0 {
		return
	}

	rLin := srgb8ToLinearLUT[col.R]
	gLin := srgb8ToLinearLUT[col.G]
	bLin := srgb8ToLinearLUT[col.B]
	baseAlpha := (float64(col.A) / 255.0) * common.Constrain(alpha, 0, 1)

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
// It also records the centerline path for pen-plotter SVG export and TSP path optimization.
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
		c.rasterizePathInternal(stroked, p, width, col, alpha)
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
	c.rasterizePathInternal(combined, p, width, col, alpha)
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

// ============================================================================
// Part VI — Direction 5: Pen-Plotter SVG Export & TSP Path Optimization
// ============================================================================

// PlotterConfig configures pen-plotter SVG export, color-layer grouping, and TSP tour optimization.
type PlotterConfig struct {
	MaxPens            int     // Maximum number of physical pen color layers (0 = keep exact stroke colors, >0 = Oklab K-Means palette quantization)
	DefaultStrokeWidth float64 // Fallback pen width in pixels for filled contours when no explicit stroke width was recorded (default 0.9)
	TwoOptPasses       int     // Number of 2-Opt local search refinement passes per pen layer (default 40)
	CenterlineOnly     bool    // If true, only export explicit stroke centerlines and skip filled polygons
	IncludeBackground  bool    // If true, emit a background <rect> element
}

// DefaultPlotterConfig returns a balanced configuration for AxiDraw / mechanical pen-plotter SVG export.
func DefaultPlotterConfig() PlotterConfig {
	return PlotterConfig{
		MaxPens:            0,
		DefaultStrokeWidth: 0.9,
		TwoOptPasses:       40,
		CenterlineOnly:     false,
		IncludeBackground:  false,
	}
}

// PlotterStats reports pen-plotter path count, layer count, drawing distance, and TSP pen-up travel savings.
type PlotterStats struct {
	LayerCount         int     `json:"layer_count"`
	TotalPaths         int     `json:"total_paths"`
	PenDownDistancePx  float64 `json:"pen_down_distance_px"`
	PenUpBeforePx      float64 `json:"pen_up_before_px"`
	PenUpAfterPx       float64 `json:"pen_up_after_px"`
	TravelReductionPct float64 `json:"travel_reduction_pct"`
}

type plotterStroke struct {
	path       *canvas.Path
	start      Point
	end        Point
	strokeW    float64
	col        color.RGBA
	alpha      float64
	reversible bool
}

func extractPlotterSubpaths(op vectorOp, defaultW float64, centerlineOnly bool) []plotterStroke {
	srcPath := op.centerline
	w := op.strokeW
	if srcPath == nil || srcPath.Empty() {
		if centerlineOnly {
			return nil
		}
		srcPath = op.path
		w = defaultW
	}
	if srcPath == nil || srcPath.Empty() {
		return nil
	}
	if w <= 0 {
		w = defaultW
	}

	var out []plotterStroke
	for _, sp := range srcPath.Split() {
		if sp == nil || sp.Empty() {
			continue
		}
		spStart := sp.StartPos()
		spEnd := sp.Pos()
		out = append(out, plotterStroke{
			path:       sp,
			start:      Pt(spStart.X, spStart.Y),
			end:        Pt(spEnd.X, spEnd.Y),
			strokeW:    w,
			col:        op.col,
			alpha:      op.alpha,
			reversible: true,
		})
	}
	return out
}

// quantizePlotterColors clusters the distinct stroke colors into at most maxPens Oklab centroids.
func quantizePlotterColors(strokes []plotterStroke, maxPens int) {
	if maxPens <= 0 || len(strokes) == 0 {
		return
	}
	type colEntry struct {
		col   color.RGBA
		lab   Oklab
		count int
	}
	freqMap := make(map[[3]uint8]*colEntry)
	for _, s := range strokes {
		key := [3]uint8{s.col.R, s.col.G, s.col.B}
		if e, ok := freqMap[key]; ok {
			e.count++
		} else {
			freqMap[key] = &colEntry{
				col:   color.RGBA{R: s.col.R, G: s.col.G, B: s.col.B, A: 255},
				lab:   RGBAToOklab(s.col),
				count: 1,
			}
		}
	}
	if len(freqMap) <= maxPens {
		return
	}

	entries := make([]*colEntry, 0, len(freqMap))
	for _, e := range freqMap {
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].count != entries[j].count {
			return entries[i].count > entries[j].count
		}
		return entries[i].lab.L < entries[j].lab.L
	})

	// Farthest-point diversity seeding in Oklab space
	centers := make([]Oklab, 0, maxPens)
	centers = append(centers, entries[0].lab)
	for len(centers) < maxPens {
		bestIdx := 0
		bestMinDist := -1.0
		for i, e := range entries {
			minD := 1e9
			for _, c := range centers {
				dL := e.lab.L - c.L
				dA := e.lab.A - c.A
				dB := e.lab.B - c.B
				d := dL*dL + 2.5*(dA*dA+dB*dB)
				if d < minD {
					minD = d
				}
			}
			if minD > bestMinDist {
				bestMinDist = minD
				bestIdx = i
			}
		}
		centers = append(centers, entries[bestIdx].lab)
	}

	// 4 iterations of weighted Oklab K-Means
	for iter := 0; iter < 4; iter++ {
		sumL := make([]float64, maxPens)
		sumA := make([]float64, maxPens)
		sumB := make([]float64, maxPens)
		sumW := make([]float64, maxPens)
		for _, e := range entries {
			bestK := 0
			bestD := 1e9
			for k, c := range centers {
				dL := e.lab.L - c.L
				dA := e.lab.A - c.A
				dB := e.lab.B - c.B
				d := dL*dL + 2.5*(dA*dA+dB*dB)
				if d < bestD {
					bestD = d
					bestK = k
				}
			}
			w := float64(e.count)
			sumL[bestK] += e.lab.L * w
			sumA[bestK] += e.lab.A * w
			sumB[bestK] += e.lab.B * w
			sumW[bestK] += w
		}
		for k := range centers {
			if sumW[k] > 0 {
				centers[k] = Oklab{
					L: sumL[k] / sumW[k],
					A: sumA[k] / sumW[k],
					B: sumB[k] / sumW[k],
				}
			}
		}
	}

	centerRGBA := make([]color.RGBA, maxPens)
	for k, c := range centers {
		centerRGBA[k] = OklabToRGBA(c, 255)
	}

	for i := range strokes {
		lab := RGBAToOklab(strokes[i].col)
		bestK := 0
		bestD := 1e9
		for k, c := range centers {
			dL := lab.L - c.L
			dA := lab.A - c.A
			dB := lab.B - c.B
			d := dL*dL + 2.5*(dA*dA+dB*dB)
			if d < bestD {
				bestD = d
				bestK = k
			}
		}
		strokes[i].col = centerRGBA[bestK]
	}
}

func computePenUpDistance(strokes []plotterStroke, startOrigin Point) float64 {
	cur := startOrigin
	total := 0.0
	for _, s := range strokes {
		total += cur.Dist(s.start)
		cur = s.end
	}
	return total
}

func reversePlotterStroke(s *plotterStroke) {
	s.path = s.path.Reverse()
	s.start, s.end = s.end, s.start
}

// optimizeLayerStrokesTSP reorders strokes (and flips reversible stroke directions) using
// Bidirectional Nearest-Neighbor initialization followed by 2-Opt segment-reversal local search.
func optimizeLayerStrokesTSP(strokes []plotterStroke, startOrigin Point, twoOptPasses int) []plotterStroke {
	n := len(strokes)
	if n <= 1 {
		return strokes
	}

	// 1. Bidirectional Nearest-Neighbor greedy tour construction
	used := make([]bool, n)
	tour := make([]plotterStroke, 0, n)
	cur := startOrigin

	for step := 0; step < n; step++ {
		bestIdx := -1
		bestRev := false
		bestDistSq := 1e30

		for i := 0; i < n; i++ {
			if used[i] {
				continue
			}
			dxF := strokes[i].start.X - cur.X
			dyF := strokes[i].start.Y - cur.Y
			dSqF := dxF*dxF + dyF*dyF
			if dSqF < bestDistSq {
				bestDistSq = dSqF
				bestIdx = i
				bestRev = false
			}
			if strokes[i].reversible {
				dxR := strokes[i].end.X - cur.X
				dyR := strokes[i].end.Y - cur.Y
				dSqR := dxR*dxR + dyR*dyR
				if dSqR < bestDistSq {
					bestDistSq = dSqR
					bestIdx = i
					bestRev = true
				}
			}
		}

		used[bestIdx] = true
		chosen := strokes[bestIdx]
		if bestRev {
			reversePlotterStroke(&chosen)
		}
		cur = chosen.end
		tour = append(tour, chosen)
	}

	// 2. 2-Opt Local Search + Endpoint Flip Refinement
	if twoOptPasses <= 0 {
		twoOptPasses = 30
	}
	maxWindow := 96
	if n < maxWindow {
		maxWindow = n
	}

	for pass := 0; pass < twoOptPasses; pass++ {
		improved := false

		// 2a. Single-stroke orientation flip check
		for i := 0; i < n; i++ {
			if !tour[i].reversible {
				continue
			}
			prevEnd := startOrigin
			if i > 0 {
				prevEnd = tour[i-1].end
			}
			if i < n-1 {
				nextStart := tour[i+1].start
				curCost := prevEnd.Dist(tour[i].start) + tour[i].end.Dist(nextStart)
				flipCost := prevEnd.Dist(tour[i].end) + tour[i].start.Dist(nextStart)
				if flipCost+1e-6 < curCost {
					reversePlotterStroke(&tour[i])
					improved = true
				}
			} else {
				curCost := prevEnd.Dist(tour[i].start)
				flipCost := prevEnd.Dist(tour[i].end)
				if flipCost+1e-6 < curCost {
					reversePlotterStroke(&tour[i])
					improved = true
				}
			}
		}

		// 2b. 2-Opt sub-tour reversal (reversing tour[i..k] and flipping each stroke inside [i..k])
		for i := 0; i < n-1; i++ {
			prevEnd := startOrigin
			if i > 0 {
				prevEnd = tour[i-1].end
			}
			kLimit := i + maxWindow
			if kLimit >= n {
				kLimit = n - 1
			}
			for k := i + 1; k <= kLimit; k++ {
				var curBridge, newBridge float64
				if k < n-1 {
					nextStart := tour[k+1].start
					curBridge = prevEnd.Dist(tour[i].start) + tour[k].end.Dist(nextStart)
					newBridge = prevEnd.Dist(tour[k].end) + tour[i].start.Dist(nextStart)
				} else {
					curBridge = prevEnd.Dist(tour[i].start)
					newBridge = prevEnd.Dist(tour[k].end)
				}
				if newBridge+1e-5 < curBridge {
					// Reverse sub-slice tour[i : k+1] and flip each stroke's direction
					for l, r := i, k; l < r; l, r = l+1, r-1 {
						tour[l], tour[r] = tour[r], tour[l]
					}
					for m := i; m <= k; m++ {
						reversePlotterStroke(&tour[m])
					}
					improved = true
				}
			}
		}

		if !improved {
			break
		}
	}

	return tour
}

// OptimizeVectorOpsTSP reorders the Canvas's recorded vector operations in-place by grouping
// into pen color layers and solving a Bidirectional Nearest-Neighbor + 2-Opt TSP tour per layer.
func (c *Canvas) OptimizeVectorOpsTSP(cfg PlotterConfig) PlotterStats {
	if c == nil || len(c.vecOps) == 0 {
		return PlotterStats{}
	}
	if cfg.DefaultStrokeWidth <= 0 {
		cfg.DefaultStrokeWidth = 0.9
	}

	var allStrokes []plotterStroke
	for _, op := range c.vecOps {
		allStrokes = append(allStrokes, extractPlotterSubpaths(op, cfg.DefaultStrokeWidth, cfg.CenterlineOnly)...)
	}
	if len(allStrokes) == 0 {
		return PlotterStats{}
	}

	if cfg.MaxPens > 0 {
		quantizePlotterColors(allStrokes, cfg.MaxPens)
	}

	// Group strokes by RGB color while preserving first-seen layer order
	type penLayer struct {
		key     [3]uint8
		col     color.RGBA
		strokes []plotterStroke
	}
	layerMap := make(map[[3]uint8]int)
	var layers []penLayer

	for _, s := range allStrokes {
		key := [3]uint8{s.col.R, s.col.G, s.col.B}
		if lIdx, exists := layerMap[key]; exists {
			layers[lIdx].strokes = append(layers[lIdx].strokes, s)
		} else {
			layerMap[key] = len(layers)
			layers = append(layers, penLayer{
				key:     key,
				col:     s.col,
				strokes: []plotterStroke{s},
			})
		}
	}

	// Sort layers from lightest/brightest pen to darkest pen (standard pen-plotter convention)
	sort.SliceStable(layers, func(i, j int) bool {
		li := RGBAToOklab(layers[i].col).L
		lj := RGBAToOklab(layers[j].col).L
		return li > lj
	})

	var stats PlotterStats
	stats.LayerCount = len(layers)
	stats.TotalPaths = len(allStrokes)
	stats.PenUpBeforePx = computePenUpDistance(allStrokes, Pt(0, 0))

	optimizedOps := make([]vectorOp, 0, len(allStrokes))
	penPos := Pt(0, 0)

	for i := range layers {
		beforeLayer := computePenUpDistance(layers[i].strokes, penPos)
		optStrokes := optimizeLayerStrokesTSP(layers[i].strokes, penPos, cfg.TwoOptPasses)
		afterLayer := computePenUpDistance(optStrokes, penPos)
		if afterLayer <= beforeLayer {
			layers[i].strokes = optStrokes
			stats.PenUpAfterPx += afterLayer
		} else {
			stats.PenUpAfterPx += beforeLayer
		}

		for _, s := range layers[i].strokes {
			stats.PenDownDistancePx += s.path.Length()
			penPos = s.end
			stroked := strokeSingleSubpathSafe(s.path, s.strokeW, canvas.RoundCap, canvas.RoundJoin)
			optimizedOps = append(optimizedOps, vectorOp{
				path:       stroked,
				centerline: s.path,
				strokeW:    s.strokeW,
				col:        s.col,
				alpha:      s.alpha,
			})
		}
	}

	if stats.PenUpBeforePx > 1e-6 {
		stats.TravelReductionPct = math.Max(0, (1.0-stats.PenUpAfterPx/stats.PenUpBeforePx)*100.0)
	}
	c.vecOps = optimizedOps
	return stats
}

// WritePlotterSVG exports the recorded vector paths as an AxiDraw / Inkscape pen-plotter SVG
// with color-grouped <g inkscape:groupmode="layer"> pen layers and Bidirectional 2-Opt TSP path ordering.
func (c *Canvas) WritePlotterSVG(w io.Writer, cfg PlotterConfig) (PlotterStats, error) {
	if c == nil || c.Width <= 0 || c.Height <= 0 {
		return PlotterStats{}, fmt.Errorf("invalid canvas dimensions")
	}
	if cfg.DefaultStrokeWidth <= 0 {
		cfg.DefaultStrokeWidth = 0.9
	}

	var allStrokes []plotterStroke
	for _, op := range c.vecOps {
		allStrokes = append(allStrokes, extractPlotterSubpaths(op, cfg.DefaultStrokeWidth, cfg.CenterlineOnly)...)
	}
	if len(allStrokes) == 0 {
		return PlotterStats{}, c.WriteSVG(w)
	}

	if cfg.MaxPens > 0 {
		quantizePlotterColors(allStrokes, cfg.MaxPens)
	}

	type penLayer struct {
		col     color.RGBA
		strokes []plotterStroke
	}
	layerMap := make(map[[3]uint8]int)
	var layers []penLayer

	for _, s := range allStrokes {
		key := [3]uint8{s.col.R, s.col.G, s.col.B}
		if lIdx, exists := layerMap[key]; exists {
			layers[lIdx].strokes = append(layers[lIdx].strokes, s)
		} else {
			layerMap[key] = len(layers)
			layers = append(layers, penLayer{
				col:     s.col,
				strokes: []plotterStroke{s},
			})
		}
	}

	sort.SliceStable(layers, func(i, j int) bool {
		return RGBAToOklab(layers[i].col).L > RGBAToOklab(layers[j].col).L
	})

	var stats PlotterStats
	stats.LayerCount = len(layers)
	stats.TotalPaths = len(allStrokes)
	stats.PenUpBeforePx = computePenUpDistance(allStrokes, Pt(0, 0))

	penPos := Pt(0, 0)
	for i := range layers {
		beforeLayer := computePenUpDistance(layers[i].strokes, penPos)
		opt := optimizeLayerStrokesTSP(layers[i].strokes, penPos, cfg.TwoOptPasses)
		afterLayer := computePenUpDistance(opt, penPos)
		if afterLayer <= beforeLayer {
			layers[i].strokes = opt
			stats.PenUpAfterPx += afterLayer
		} else {
			stats.PenUpAfterPx += beforeLayer
		}
		for _, s := range layers[i].strokes {
			stats.PenDownDistancePx += s.path.Length()
			penPos = s.end
		}
	}
	if stats.PenUpBeforePx > 1e-6 {
		stats.TravelReductionPct = math.Max(0, (1.0-stats.PenUpAfterPx/stats.PenUpBeforePx)*100.0)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, `<?xml version="1.0" encoding="UTF-8"?>`+"\n")
	fmt.Fprintf(&sb, `<!-- GenerativeArt Pen-Plotter SVG | Layers: %d | Paths: %d | Pen-Down: %.1fpx | Pen-Up Before: %.1fpx -> After: %.1fpx (-%.1f%%) -->`+"\n",
		stats.LayerCount, stats.TotalPaths, stats.PenDownDistancePx, stats.PenUpBeforePx, stats.PenUpAfterPx, stats.TravelReductionPct)
	fmt.Fprintf(&sb, `<svg xmlns="http://www.w3.org/2000/svg" xmlns:inkscape="http://www.inkscape.org/namespaces/inkscape" width="%d" height="%d" viewBox="0 0 %d %d">`+"\n",
		c.Width, c.Height, c.Width, c.Height)

	if cfg.IncludeBackground && c.hasBg {
		fmt.Fprintf(&sb, `  <rect width="%d" height="%d" fill="#%02x%02x%02x"/>`+"\n",
			c.Width, c.Height, c.bgCol.R, c.bgCol.G, c.bgCol.B)
	}

	for idx, layer := range layers {
		hexCol := fmt.Sprintf("#%02x%02x%02x", layer.col.R, layer.col.G, layer.col.B)
		fmt.Fprintf(&sb, `  <g id="pen-layer-%d-%02x%02x%02x" inkscape:groupmode="layer" inkscape:label="Pen %d (%s)" fill="none" stroke="%s" stroke-linecap="round" stroke-linejoin="round">`+"\n",
			idx+1, layer.col.R, layer.col.G, layer.col.B, idx+1, hexCol, hexCol)
		for _, s := range layer.strokes {
			dStr := s.path.ToSVG()
			if dStr == "" {
				continue
			}
			if s.alpha < 0.995 {
				fmt.Fprintf(&sb, `    <path d="%s" stroke-width="%.2f" stroke-opacity="%.2f"/>`+"\n", dStr, s.strokeW, s.alpha)
			} else {
				fmt.Fprintf(&sb, `    <path d="%s" stroke-width="%.2f"/>`+"\n", dStr, s.strokeW)
			}
		}
		sb.WriteString("  </g>\n")
	}
	sb.WriteString("</svg>\n")

	_, err := io.WriteString(w, sb.String())
	return stats, err
}

// SavePlotterSVG writes a pen-plotter-optimized SVG file at filename and returns PlotterStats.
func (c *Canvas) SavePlotterSVG(filename string, cfg PlotterConfig) (PlotterStats, error) {
	f, err := os.Create(filename)
	if err != nil {
		return PlotterStats{}, err
	}
	stats, wErr := c.WritePlotterSVG(f, cfg)
	if cErr := f.Close(); wErr == nil {
		wErr = cErr
	}
	return stats, wErr
}


