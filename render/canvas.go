package render

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"runtime"
	"sync"

	"github.com/jdxyw/generativeart/common"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/tdewolff/canvas"
	"golang.org/x/image/vector"
)

var srgb8ToLinearLUT [256]float64

func init() {
	for i := 0; i < 256; i++ {
		srgb8ToLinearLUT[i] = SRGBToLinear(float64(i) / 255.0)
	}
}

// SRGB8ToLinear returns the exact linear-light value in [0, 1] for an 8-bit sRGB component
// using an O(1) precomputed lookup table.
func SRGB8ToLinear(b uint8) float64 {
	return srgb8ToLinearLUT[b]
}

// parallelRows partitions [0, height) across runtime.NumCPU() workers when the canvas
// has at least 16,384 pixels, falling back to a single-threaded call for small buffers.
func parallelRows(width, height int, fn func(y0, y1 int)) {
	if height <= 0 {
		return
	}
	workers := runtime.NumCPU()
	if workers <= 1 || width*height < 16384 || height < workers*2 {
		fn(0, height)
		return
	}
	if workers > height {
		workers = height
	}
	var wg sync.WaitGroup
	chunk := (height + workers - 1) / workers
	for w := 0; w < workers; w++ {
		y0 := w * chunk
		y1 := y0 + chunk
		if y0 >= height {
			break
		}
		if y1 > height {
			y1 = height
		}
		wg.Add(1)
		go func(startY, endY int) {
			defer wg.Done()
			fn(startY, endY)
		}(y0, y1)
	}
	wg.Wait()
}

// Point represents a 2D high-precision floating-point coordinate.
type Point struct {
	X, Y float64
}

// Pt creates a Point from x and y coordinates.
func Pt(x, y float64) Point {
	return Point{X: x, Y: y}
}

// Vec converts a common.Vector into a render.Point.
func Vec(v common.Vector) Point {
	return Point{X: v.X, Y: v.Y}
}

// ToVector converts a render.Point into a common.Vector.
func (p Point) ToVector() common.Vector {
	return common.Vector{X: p.X, Y: p.Y}
}

// Add returns the vector sum p + q.
func (p Point) Add(q Point) Point {
	return Point{X: p.X + q.X, Y: p.Y + q.Y}
}

// Sub returns the vector difference p - q.
func (p Point) Sub(q Point) Point {
	return Point{X: p.X - q.X, Y: p.Y - q.Y}
}

// Mul scales p by scalar s.
func (p Point) Mul(s float64) Point {
	return Point{X: p.X * s, Y: p.Y * s}
}

// Len returns the Euclidean length of p.
func (p Point) Len() float64 {
	return math.Hypot(p.X, p.Y)
}

// Dist returns the Euclidean distance between p and q.
func (p Point) Dist(q Point) float64 {
	return common.Distance(p.X, p.Y, q.X, q.Y)
}

// Norm returns the unit vector in the direction of p, or (0, 0) if zero-length.
func (p Point) Norm() Point {
	l := math.Hypot(p.X, p.Y)
	if l < 1e-12 {
		return Point{}
	}
	return Point{X: p.X / l, Y: p.Y / l}
}

// Perp returns the counter-clockwise 90-degree perpendicular vector (-Y, X).
func (p Point) Perp() Point {
	return Point{X: -p.Y, Y: p.X}
}

// Lerp linearly interpolates between p and q by t in [0, 1].
func (p Point) Lerp(q Point, t float64) Point {
	return Point{
		X: p.X + (q.X-p.X)*t,
		Y: p.Y + (q.Y-p.Y)*t,
	}
}

// Dot returns the 2D dot product p.X*q.X + p.Y*q.Y.
func (p Point) Dot(q Point) float64 {
	return p.X*q.X + p.Y*q.Y
}

// Cross returns the 2D scalar cross product p.X*q.Y - p.Y*q.X.
func (p Point) Cross(q Point) float64 {
	return p.X*q.Y - p.Y*q.X
}

// BlendMode defines how source colors are composited into the linear-RGB HDR canvas.
type BlendMode int

const (
	BlendOver BlendMode = iota
	BlendAdd
	BlendMultiply
	BlendScreen
	BlendOverlay
	BlendSoftLight
	BlendDarken
	BlendLighten
	BlendDifference
)

// ColorSpace specifies the perceptual color space used for color interpolation.
type ColorSpace int

const (
	ColorSpaceOklab ColorSpace = iota
	ColorSpaceHCL
	ColorSpaceLab
	ColorSpaceLuv
	ColorSpaceLinearRGB
)

// Oklab represents a color in Björn Ottosson's Oklab perceptual color space.
type Oklab struct {
	L, A, B float64
}

// SRGBToLinear converts a normalized sRGB component [0, 1] to linear light [0, 1].
func SRGBToLinear(c float64) float64 {
	if c <= 0 {
		return 0
	}
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

// LinearToSRGB converts a linear light component [0, 1] to normalized sRGB [0, 1].
func LinearToSRGB(c float64) float64 {
	if c <= 0 {
		return 0
	}
	if c >= 1 {
		return 1
	}
	if c <= 0.0031308 {
		return 12.92 * c
	}
	return 1.055*math.Pow(c, 1.0/2.4) - 0.055
}

// ColorToLinearRGB extracts non-premultiplied linear RGB and normalized alpha from any color.Color.
func ColorToLinearRGB(col color.Color) (rLin, gLin, bLin, a float64) {
	if rgba, ok := col.(color.RGBA); ok {
		rLin = srgb8ToLinearLUT[rgba.R]
		gLin = srgb8ToLinearLUT[rgba.G]
		bLin = srgb8ToLinearLUT[rgba.B]
		a = float64(rgba.A) / 255.0
		return
	}
	if nrgba, ok := col.(color.NRGBA); ok {
		rLin = srgb8ToLinearLUT[nrgba.R]
		gLin = srgb8ToLinearLUT[nrgba.G]
		bLin = srgb8ToLinearLUT[nrgba.B]
		a = float64(nrgba.A) / 255.0
		return
	}
	r, g, b, alpha := col.RGBA()
	if alpha == 0 {
		return 0, 0, 0, 0
	}
	fa := float64(alpha)
	rLin = SRGBToLinear(float64(r) / fa)
	gLin = SRGBToLinear(float64(g) / fa)
	bLin = SRGBToLinear(float64(b) / fa)
	a = fa / 65535.0
	return
}

// LinearRGBAToOklab converts linear RGB [0, 1] to Oklab.
func LinearRGBAToOklab(rLin, gLin, bLin float64) Oklab {
	l := 0.4122214708*rLin + 0.5363325363*gLin + 0.0514459929*bLin
	m := 0.2119034982*rLin + 0.6806995451*gLin + 0.1073969566*bLin
	s := 0.0883024619*rLin + 0.2817188376*gLin + 0.6299787005*bLin

	l_ := math.Cbrt(l)
	m_ := math.Cbrt(m)
	s_ := math.Cbrt(s)

	return Oklab{
		L: 0.2104542553*l_ + 0.7936177850*m_ - 0.0040720468*s_,
		A: 1.9779984951*l_ - 2.4285922050*m_ + 0.4505937099*s_,
		B: 0.0259040371*l_ + 0.7827717662*m_ - 0.8086757660*s_,
	}
}

// OklabToLinearRGB converts Oklab coordinates back to linear RGB.
func OklabToLinearRGB(lab Oklab) (rLin, gLin, bLin float64) {
	l_ := lab.L + 0.3963377774*lab.A + 0.2158037573*lab.B
	m_ := lab.L - 0.1055613458*lab.A - 0.0638541728*lab.B
	s_ := lab.L - 0.0894841775*lab.A - 1.2914855480*lab.B

	l := l_ * l_ * l_
	m := m_ * m_ * m_
	s := s_ * s_ * s_

	rLin = +4.0767416621*l - 3.3077115913*m + 0.2309699292*s
	gLin = -1.2684380046*l + 2.6097574011*m - 0.3413193965*s
	bLin = -0.0041960863*l - 0.7034186147*m + 1.7076147010*s
	return
}

// RGBAToOklab converts an sRGB color.RGBA to Oklab.
func RGBAToOklab(c color.RGBA) Oklab {
	return LinearRGBAToOklab(srgb8ToLinearLUT[c.R], srgb8ToLinearLUT[c.G], srgb8ToLinearLUT[c.B])
}

// OklabToRGBA converts an Oklab color and alpha byte back to sRGB color.RGBA with gamut clamping.
func OklabToRGBA(lab Oklab, alpha uint8) color.RGBA {
	rLin, gLin, bLin := OklabToLinearRGB(lab)
	r := LinearToSRGB(common.Constrain(rLin, 0, 1))
	g := LinearToSRGB(common.Constrain(gLin, 0, 1))
	b := LinearToSRGB(common.Constrain(bLin, 0, 1))
	return color.RGBA{
		R: uint8(math.Round(r * 255.0)),
		G: uint8(math.Round(g * 255.0)),
		B: uint8(math.Round(b * 255.0)),
		A: alpha,
	}
}

// LerpOklab interpolates between c1 and c2 in exact Oklab perceptual space for t in [0, 1].
func LerpOklab(c1, c2 color.RGBA, t float64) color.RGBA {
	t = common.Constrain(t, 0, 1)
	lab1 := RGBAToOklab(c1)
	lab2 := RGBAToOklab(c2)
	lab := Oklab{
		L: lab1.L + (lab2.L-lab1.L)*t,
		A: lab1.A + (lab2.A-lab1.A)*t,
		B: lab1.B + (lab2.B-lab1.B)*t,
	}
	a := uint8(math.Round(float64(c1.A) + (float64(c2.A)-float64(c1.A))*t))
	return OklabToRGBA(lab, a)
}

// LerpHCL interpolates between c1 and c2 in CIE-L*C*h° (HCL) space using go-colorful.
func LerpHCL(c1, c2 color.RGBA, t float64) color.RGBA {
	t = common.Constrain(t, 0, 1)
	col1 := colorful.Color{R: float64(c1.R) / 255.0, G: float64(c1.G) / 255.0, B: float64(c1.B) / 255.0}
	col2 := colorful.Color{R: float64(c2.R) / 255.0, G: float64(c2.G) / 255.0, B: float64(c2.B) / 255.0}
	blended := col1.BlendHcl(col2, t).Clamped()
	r, g, b := blended.RGB255()
	a := uint8(math.Round(float64(c1.A) + (float64(c2.A)-float64(c1.A))*t))
	return color.RGBA{R: r, G: g, B: b, A: a}
}

// LerpLab interpolates between c1 and c2 in CIELAB space using go-colorful.
func LerpLab(c1, c2 color.RGBA, t float64) color.RGBA {
	t = common.Constrain(t, 0, 1)
	col1 := colorful.Color{R: float64(c1.R) / 255.0, G: float64(c1.G) / 255.0, B: float64(c1.B) / 255.0}
	col2 := colorful.Color{R: float64(c2.R) / 255.0, G: float64(c2.G) / 255.0, B: float64(c2.B) / 255.0}
	blended := col1.BlendLab(col2, t).Clamped()
	r, g, b := blended.RGB255()
	a := uint8(math.Round(float64(c1.A) + (float64(c2.A)-float64(c1.A))*t))
	return color.RGBA{R: r, G: g, B: b, A: a}
}

// LerpLuv interpolates between c1 and c2 in CIELUV space using go-colorful.
func LerpLuv(c1, c2 color.RGBA, t float64) color.RGBA {
	t = common.Constrain(t, 0, 1)
	col1 := colorful.Color{R: float64(c1.R) / 255.0, G: float64(c1.G) / 255.0, B: float64(c1.B) / 255.0}
	col2 := colorful.Color{R: float64(c2.R) / 255.0, G: float64(c2.G) / 255.0, B: float64(c2.B) / 255.0}
	blended := col1.BlendLuv(col2, t).Clamped()
	r, g, b := blended.RGB255()
	a := uint8(math.Round(float64(c1.A) + (float64(c2.A)-float64(c1.A))*t))
	return color.RGBA{R: r, G: g, B: b, A: a}
}

// MultiStopGradient represents an ordered multi-stop color gradient with pre-cached
// Oklab and Linear-RGB stop coordinates for fast interpolation.
type MultiStopGradient struct {
	Colors      []color.RGBA
	oklabStops  []Oklab
	linearStops [][4]float64
}

// NewMultiStopGradient creates a MultiStopGradient from a slice or variadic list of colors.
func NewMultiStopGradient(colors ...color.RGBA) *MultiStopGradient {
	n := len(colors)
	cp := make([]color.RGBA, n)
	copy(cp, colors)
	okStops := make([]Oklab, n)
	linStops := make([][4]float64, n)
	for i, c := range cp {
		rL, gL, bL, a := ColorToLinearRGB(c)
		linStops[i] = [4]float64{rL, gL, bL, a}
		okStops[i] = LinearRGBAToOklab(rL, gL, bL)
	}
	return &MultiStopGradient{
		Colors:      cp,
		oklabStops:  okStops,
		linearStops: linStops,
	}
}

// SampleOklab samples the multi-stop gradient at t in [0, 1] using Oklab interpolation.
func (g *MultiStopGradient) SampleOklab(t float64) color.RGBA {
	return g.Sample(t, ColorSpaceOklab)
}

// SampleHCL samples the multi-stop gradient at t in [0, 1] using HCL interpolation.
func (g *MultiStopGradient) SampleHCL(t float64) color.RGBA {
	return g.Sample(t, ColorSpaceHCL)
}

// Sample samples the multi-stop gradient at t in [0, 1] in the requested ColorSpace.
func (g *MultiStopGradient) Sample(t float64, space ColorSpace) color.RGBA {
	n := len(g.Colors)
	if n == 0 {
		return common.Black
	}
	if n == 1 {
		return g.Colors[0]
	}
	t = common.Constrain(t, 0, 1)
	scaled := t * float64(n-1)
	idx := int(math.Floor(scaled))
	if idx >= n-1 {
		return g.Colors[n-1]
	}
	if idx < 0 {
		return g.Colors[0]
	}
	localT := scaled - float64(idx)
	c1, c2 := g.Colors[idx], g.Colors[idx+1]
	switch space {
	case ColorSpaceHCL:
		return LerpHCL(c1, c2, localT)
	case ColorSpaceLab:
		return LerpLab(c1, c2, localT)
	case ColorSpaceLuv:
		return LerpLuv(c1, c2, localT)
	case ColorSpaceLinearRGB:
		var r1, g1, b1, a1, r2, g2, b2, a2 float64
		if len(g.linearStops) == n {
			s1, s2 := g.linearStops[idx], g.linearStops[idx+1]
			r1, g1, b1, a1 = s1[0], s1[1], s1[2], s1[3]
			r2, g2, b2, a2 = s2[0], s2[1], s2[2], s2[3]
		} else {
			r1, g1, b1, a1 = ColorToLinearRGB(c1)
			r2, g2, b2, a2 = ColorToLinearRGB(c2)
		}
		r := LinearToSRGB(common.Constrain(r1+(r2-r1)*localT, 0, 1))
		gr := LinearToSRGB(common.Constrain(g1+(g2-g1)*localT, 0, 1))
		b := LinearToSRGB(common.Constrain(b1+(b2-b1)*localT, 0, 1))
		a := common.Constrain(a1+(a2-a1)*localT, 0, 1)
		return color.RGBA{
			R: uint8(math.Round(r * 255.0)),
			G: uint8(math.Round(gr * 255.0)),
			B: uint8(math.Round(b * 255.0)),
			A: uint8(math.Round(a * 255.0)),
		}
	default:
		if len(g.oklabStops) == n {
			lab1, lab2 := g.oklabStops[idx], g.oklabStops[idx+1]
			lab := Oklab{
				L: lab1.L + (lab2.L-lab1.L)*localT,
				A: lab1.A + (lab2.A-lab1.A)*localT,
				B: lab1.B + (lab2.B-lab1.B)*localT,
			}
			a := uint8(math.Round(float64(c1.A) + (float64(c2.A)-float64(c1.A))*localT))
			return OklabToRGBA(lab, a)
		}
		return LerpOklab(c1, c2, localT)
	}
}

// Canvas is a high-precision 64-bit float HDR Linear-RGB accumulation buffer.
// Pixels are stored in row-major order with 4 float64 channels per pixel:
// non-premultiplied linear R, G, B, and straight alpha A in [0, 1].
type Canvas struct {
	Width     int
	Height    int
	pix       []float64
	blendMode BlendMode

	ras          *vector.Rasterizer
	maskBuf      *image.Alpha
	recordVector bool
	hasBg        bool
	bgCol        color.RGBA
	vecOps       []vectorOp

	hasCustomTransform bool
	curMatrix          canvas.Matrix
	transformStack     []canvas.Matrix
	clipStack          [][]uint8
	clipPolyStack      [][][]Point
}

// NewCanvas allocates a new HDR linear-RGB Canvas with dimensions w x h.
func NewCanvas(w, h int) *Canvas {
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	return &Canvas{
		Width:        w,
		Height:       h,
		pix:          make([]float64, w*h*4),
		blendMode:    BlendOver,
		recordVector: true,
		curMatrix:    canvas.Identity,
	}
}

// NewCanvasFromRGBA initializes an HDR linear-RGB Canvas from an existing *image.RGBA.
func NewCanvasFromRGBA(img *image.RGBA) *Canvas {
	if img == nil {
		return NewCanvas(0, 0)
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	c := NewCanvas(w, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			rgba := img.RGBAAt(b.Min.X+x, b.Min.Y+y)
			var rS, gS, bS, a float64
			if rgba.A > 0 {
				a = float64(rgba.A) / 255.0
				// image.RGBA stores premultiplied bytes; un-premultiply before sRGB->Linear.
				rS = common.Constrain(float64(rgba.R)/float64(rgba.A), 0, 1)
				gS = common.Constrain(float64(rgba.G)/float64(rgba.A), 0, 1)
				bS = common.Constrain(float64(rgba.B)/float64(rgba.A), 0, 1)
			}
			idx := (y*w + x) * 4
			c.pix[idx+0] = SRGBToLinear(rS)
			c.pix[idx+1] = SRGBToLinear(gS)
			c.pix[idx+2] = SRGBToLinear(bS)
			c.pix[idx+3] = a
		}
	}
	return c
}

// SetBlendMode sets the active compositing BlendMode for subsequent drawing operations.
func (c *Canvas) SetBlendMode(mode BlendMode) {
	c.blendMode = mode
}

// GetBlendMode returns the currently active BlendMode.
func (c *Canvas) GetBlendMode() BlendMode {
	return c.blendMode
}

// FillBackground fills the entire canvas with a solid color in Linear RGB.
func (c *Canvas) FillBackground(col color.Color) {
	rLin, gLin, bLin, a := ColorToLinearRGB(col)
	for i := 0; i < len(c.pix); i += 4 {
		c.pix[i+0] = rLin
		c.pix[i+1] = gLin
		c.pix[i+2] = bLin
		c.pix[i+3] = a
	}
	if rgba, ok := col.(color.RGBA); ok {
		c.bgCol = rgba
	} else {
		c.bgCol = color.RGBA{
			R: uint8(math.Round(LinearToSRGB(rLin) * 255.0)),
			G: uint8(math.Round(LinearToSRGB(gLin) * 255.0)),
			B: uint8(math.Round(LinearToSRGB(bLin) * 255.0)),
			A: uint8(math.Round(a * 255.0)),
		}
	}
	c.hasBg = true
	c.vecOps = c.vecOps[:0]
}

// PixelLinearAt returns the linear-RGB and alpha values at pixel (x, y).
func (c *Canvas) PixelLinearAt(x, y int) (r, g, b, a float64) {
	if x < 0 || x >= c.Width || y < 0 || y >= c.Height {
		return 0, 0, 0, 0
	}
	idx := (y*c.Width + x) * 4
	return c.pix[idx+0], c.pix[idx+1], c.pix[idx+2], c.pix[idx+3]
}

// SetPixelLinear directly writes linear-RGB and alpha values at pixel (x, y).
func (c *Canvas) SetPixelLinear(x, y int, r, g, b, a float64) {
	if x < 0 || x >= c.Width || y < 0 || y >= c.Height {
		return
	}
	idx := (y*c.Width + x) * 4
	c.pix[idx+0] = r
	c.pix[idx+1] = g
	c.pix[idx+2] = b
	c.pix[idx+3] = common.Constrain(a, 0, 1)
}

// blendChannel computes the separable blend mode function B(cb, cs) on linear components.
func blendChannel(cb, cs float64, mode BlendMode) float64 {
	switch mode {
	case BlendMultiply:
		return cb * cs
	case BlendScreen:
		b := common.Constrain(cb, 0, 1)
		s := common.Constrain(cs, 0, 1)
		return b + s - b*s
	case BlendOverlay:
		b := common.Constrain(cb, 0, 1)
		s := common.Constrain(cs, 0, 1)
		if b < 0.5 {
			return 2.0 * b * s
		}
		return 1.0 - 2.0*(1.0-b)*(1.0-s)
	case BlendSoftLight:
		b := common.Constrain(cb, 0, 1)
		s := common.Constrain(cs, 0, 1)
		if s <= 0.5 {
			return b - (1.0-2.0*s)*b*(1.0-b)
		}
		var d float64
		if b <= 0.25 {
			d = ((16.0*b-12.0)*b + 4.0) * b
		} else {
			d = math.Sqrt(b)
		}
		return b + (2.0*s-1.0)*(d-b)
	case BlendDarken:
		return math.Min(cb, cs)
	case BlendLighten:
		return math.Max(cb, cs)
	case BlendDifference:
		return math.Abs(cb - cs)
	default:
		return cs
	}
}

// BlendPixelLinear composites a linear-RGB source color with coverage/opacity alpha into (x, y).
func (c *Canvas) BlendPixelLinear(x, y int, rLin, gLin, bLin, alpha float64, mode BlendMode) {
	if x < 0 || x >= c.Width || y < 0 || y >= c.Height || alpha <= 0 {
		return
	}
	if len(c.clipStack) > 0 {
		maskCov := c.clipStack[len(c.clipStack)-1][y*c.Width+x]
		if maskCov == 0 {
			return
		}
		alpha *= float64(maskCov) / 255.0
	}
	sa := common.Constrain(alpha, 0, 1)
	idx := (y*c.Width + x) * 4
	dr := c.pix[idx+0]
	dg := c.pix[idx+1]
	db := c.pix[idx+2]
	da := c.pix[idx+3]

	if mode == BlendAdd {
		// True linear-light additive accumulation (HDR capable, does not clamp above 1.0).
		outA := da + sa*(1.0-da)
		if outA <= 1e-12 {
			return
		}
		c.pix[idx+0] = (dr*da + rLin*sa) / outA
		c.pix[idx+1] = (dg*da + gLin*sa) / outA
		c.pix[idx+2] = (db*da + bLin*sa) / outA
		c.pix[idx+3] = outA
		return
	}

	outA := sa + da*(1.0-sa)
	if outA <= 1e-12 {
		return
	}

	if mode == BlendOver || da <= 1e-9 {
		invSa := 1.0 - sa
		c.pix[idx+0] = (rLin*sa + dr*da*invSa) / outA
		c.pix[idx+1] = (gLin*sa + dg*da*invSa) / outA
		c.pix[idx+2] = (bLin*sa + db*da*invSa) / outA
		c.pix[idx+3] = outA
		return
	}

	// Standard PDF/SVG Porter-Duff separable blend compositing in linear light:
	// C_out * a_out = (1 - a_b) * a_s * C_s + (1 - a_s) * a_b * C_b + a_s * a_b * B(C_b, C_s)
	br := blendChannel(dr, rLin, mode)
	bg := blendChannel(dg, gLin, mode)
	bb := blendChannel(db, bLin, mode)

	wSrc := (1.0 - da) * sa
	wDst := (1.0 - sa) * da
	wBoth := sa * da

	c.pix[idx+0] = (wSrc*rLin + wDst*dr + wBoth*br) / outA
	c.pix[idx+1] = (wSrc*gLin + wDst*dg + wBoth*bg) / outA
	c.pix[idx+2] = (wSrc*bLin + wDst*db + wBoth*bb) / outA
	c.pix[idx+3] = outA
}

// BlendPixel converts sRGB color c to Linear RGB and blends it at (x, y) using the active BlendMode.
// The effective alpha is (c.A / 255.0) * alpha.
func (c *Canvas) BlendPixel(x, y int, col color.RGBA, alpha float64) {
	if x < 0 || x >= c.Width || y < 0 || y >= c.Height || alpha <= 0 || col.A == 0 {
		return
	}
	rLin := srgb8ToLinearLUT[col.R]
	gLin := srgb8ToLinearLUT[col.G]
	bLin := srgb8ToLinearLUT[col.B]
	effAlpha := (float64(col.A) / 255.0) * common.Constrain(alpha, 0, 1)
	c.BlendPixelLinear(x, y, rLin, gLin, bLin, effAlpha, c.blendMode)
}

// ResolveToRGBA converts the HDR linear-RGB buffer to a gamma-correct sRGB *image.RGBA with clamping.
// If dst is nil or does not match the canvas dimensions, a new *image.RGBA is allocated.
func (c *Canvas) ResolveToRGBA(dst *image.RGBA) *image.RGBA {
	if dst == nil || dst.Bounds().Dx() != c.Width || dst.Bounds().Dy() != c.Height {
		dst = image.NewRGBA(image.Rect(0, 0, c.Width, c.Height))
	}
	b := dst.Bounds()
	parallelRows(c.Width, c.Height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := 0; x < c.Width; x++ {
				idx := (y*c.Width + x) * 4
				rLin := common.Constrain(c.pix[idx+0], 0, 1)
				gLin := common.Constrain(c.pix[idx+1], 0, 1)
				bLin := common.Constrain(c.pix[idx+2], 0, 1)
				a := common.Constrain(c.pix[idx+3], 0, 1)

				// Convert linear RGB -> sRGB and premultiply by alpha for image.RGBA.
				rS := LinearToSRGB(rLin) * a
				gS := LinearToSRGB(gLin) * a
				bS := LinearToSRGB(bLin) * a

				off := dst.PixOffset(b.Min.X+x, b.Min.Y+y)
				dst.Pix[off+0] = uint8(math.Round(rS * 255.0))
				dst.Pix[off+1] = uint8(math.Round(gS * 255.0))
				dst.Pix[off+2] = uint8(math.Round(bS * 255.0))
				dst.Pix[off+3] = uint8(math.Round(a * 255.0))
			}
		}
	})
	return dst
}

// acesFilmic applies Narkowicz ACES filmic tone mapping curve to a positive HDR value.
func acesFilmic(x float64) float64 {
	if x <= 0 {
		return 0
	}
	const (
		a = 2.51
		b = 0.03
		c = 2.43
		d = 0.59
		e = 0.14
	)
	return common.Constrain((x*(a*x+b))/(x*(c*x+d)+e), 0, 1)
}

// ResolveToneMapped resolves HDR linear-RGB pixels using exposure scaling and ACES filmic tone mapping.
// If gamma <= 0, standard sRGB transfer function is used; otherwise 1/gamma power curve is applied.
func (c *Canvas) ResolveToneMapped(dst *image.RGBA, exposure, gamma float64) *image.RGBA {
	if dst == nil || dst.Bounds().Dx() != c.Width || dst.Bounds().Dy() != c.Height {
		dst = image.NewRGBA(image.Rect(0, 0, c.Width, c.Height))
	}
	if exposure <= 0 {
		exposure = 1.0
	}
	b := dst.Bounds()
	useSRGB := gamma <= 0
	invGamma := 1.0 / 2.2
	if !useSRGB {
		invGamma = 1.0 / gamma
	}

	parallelRows(c.Width, c.Height, func(y0, y1 int) {
		for y := y0; y < y1; y++ {
			for x := 0; x < c.Width; x++ {
				idx := (y*c.Width + x) * 4
				rMapped := acesFilmic(c.pix[idx+0] * exposure)
				gMapped := acesFilmic(c.pix[idx+1] * exposure)
				bMapped := acesFilmic(c.pix[idx+2] * exposure)
				a := common.Constrain(c.pix[idx+3], 0, 1)

				var rS, gS, bS float64
				if useSRGB {
					rS = LinearToSRGB(rMapped) * a
					gS = LinearToSRGB(gMapped) * a
					bS = LinearToSRGB(bMapped) * a
				} else {
					rS = math.Pow(rMapped, invGamma) * a
					gS = math.Pow(gMapped, invGamma) * a
					bS = math.Pow(bMapped, invGamma) * a
				}

				off := dst.PixOffset(b.Min.X+x, b.Min.Y+y)
				dst.Pix[off+0] = uint8(math.Round(common.Constrain(rS, 0, 1) * 255.0))
				dst.Pix[off+1] = uint8(math.Round(common.Constrain(gS, 0, 1) * 255.0))
				dst.Pix[off+2] = uint8(math.Round(common.Constrain(bS, 0, 1) * 255.0))
				dst.Pix[off+3] = uint8(math.Round(a * 255.0))
			}
		}
	})
	return dst
}

// SavePNG resolves the canvas to sRGB RGBA and writes it to a PNG file.
func (c *Canvas) SavePNG(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	img := c.ResolveToRGBA(nil)
	return png.Encode(f, img)
}
