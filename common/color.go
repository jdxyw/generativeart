package common

import (
	"fmt"
	"image/color"
	"math"
	"sort"
	"strings"

	colorful "github.com/lucasb-eyer/go-colorful"
)

var (
	MistyRose        = color.RGBA{R: 0xFF, G: 0xE4, B: 0xE1, A: 0xFF}
	DarkSalmon       = color.RGBA{R: 0xE9, G: 0x96, B: 0x7A, A: 0xFF}
	Tan              = color.RGBA{R: 0xD2, G: 0xB4, B: 0x8C, A: 0xFF}
	Bisque           = color.RGBA{R: 0xFF, G: 0xE4, B: 0xC4, A: 0xFF}
	Mintcream        = color.RGBA{R: 0xF5, G: 0xFF, B: 0xFA, A: 0xFF}
	Aquamarine       = color.RGBA{R: 0x7F, G: 0xFF, B: 0xD4, A: 0xFF}
	Azure            = color.RGBA{R: 0xF0, G: 0xFF, B: 0xFF, A: 0xFF}
	Lavender         = color.RGBA{R: 0xE6, G: 0xE6, B: 0xFA, A: 0xFF}
	Plum             = color.RGBA{R: 0xDD, G: 0xA0, B: 0xDD, A: 0xFF}
	AntiqueWhite     = color.RGBA{R: 0xFA, G: 0xEB, B: 0xD7, A: 0xFF}
	NavajoWhite      = color.RGBA{R: 0xFF, G: 0xDE, B: 0xAD, A: 0xFF}
	Moccasin         = color.RGBA{R: 0xFF, G: 0xE4, B: 0xB5, A: 0xFF}
	MediumAquamarine = color.RGBA{R: 0x66, G: 0xCD, B: 0xAA, A: 0xFF}
	PaleTurquoise    = color.RGBA{R: 0xAF, G: 0xEE, B: 0xEE, A: 0xFF}
	LightPink        = color.RGBA{R: 0xFF, G: 0xB6, B: 0xC1, A: 0xFF}
	Tomato           = color.RGBA{R: 0xFF, G: 0x63, B: 0x47, A: 0xFF}
	Orange           = color.RGBA{R: 0xFF, G: 0xA5, A: 0xFF}
	Black            = color.RGBA{A: 0xFF}
	White            = color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
	LightGray        = color.RGBA{R: 200, G: 200, B: 200, A: 255}
	Outdoors         = []color.RGBA{
		{67, 110, 165, 255},
		{47, 76, 114, 255},
		{165, 138, 3, 255},
		{242, 182, 4, 255},
		{191, 131, 59, 255},
	}
	Reddery = []color.RGBA{
		{89, 2, 16, 255},
		{37, 2, 5, 255},
		{186, 7, 55, 255},
		{253, 55, 51, 255},
		{245, 169, 167, 255},
	}
	DarkPink = []color.RGBA{
		{2, 64, 89, 255},
		{242, 131, 107, 255},
		{140, 40, 31, 255},
		{191, 69, 57, 255},
		{13, 13, 13, 255},
	}

	DarkRed = []color.RGBA{
		{48, 19, 21, 255},
		{71, 15, 16, 255},
		{92, 10, 12, 255},
		{115, 5, 4, 255},
		{139, 0, 3, 255},
		{163, 32, 1, 255},
		{185, 66, 0, 255},
		{208, 98, 1, 255},
		{231, 133, 0, 255},
		{254, 165, 0, 255},
	}

	Sleek = []color.RGBA{
		{R: 0x2C, G: 0x35, B: 0x31, A: 0xFF},
		{R: 0x11, G: 0x64, B: 0x66, A: 0xFF},
		{R: 0xD9, G: 0xB0, B: 0x8C, A: 0xFF},
		{R: 0xFF, G: 0xCB, B: 0x9A, A: 0xFF},
		{R: 0xD1, G: 0xE8, B: 0xE2, A: 0xFF},
	}

	Energetic = []color.RGBA{
		{R: 0x56, G: 0x80, B: 0xE9, A: 0xFF},
		{R: 0x84, G: 0xCE, B: 0xEB, A: 0xFF},
		{R: 0x5A, G: 0x89, B: 0xEA, A: 0xFF},
		{R: 0xC1, G: 0xC8, B: 0xEA, A: 0xFF},
		{R: 0x88, G: 0x60, B: 0xD0, A: 0xFF},
	}

	Youthful = []color.RGBA{
		{R: 0xA6, G: 0x4A, B: 0xC9, A: 0xFF},
		{R: 0xFC, G: 0xCD, B: 0x04, A: 0xFF},
		{R: 0xFF, G: 0xB4, B: 0x8F, A: 0xFF},
		{R: 0xF5, G: 0xE6, B: 0xCC, A: 0xFF},
		{R: 0x17, G: 0xE9, B: 0xE0, A: 0xFF},
	}

	PinkPastel = []color.RGBA{
		{R: 0xA1, G: 0xC3, B: 0xD1, A: 0xFF},
		{R: 0xB3, G: 0x9B, B: 0xC8, A: 0xFF},
		{R: 0xF0, G: 0xEB, B: 0xF4, A: 0xFF},
		{R: 0xF1, G: 0x72, B: 0xA1, A: 0xFF},
		{R: 0xE6, G: 0x43, B: 0x98, A: 0xFF},
	}

	Citrus = []color.RGBA{
		{R: 0x1F, G: 0x26, B: 0x05, A: 0xFF},
		{R: 0x1F, G: 0x65, B: 0x21, A: 0xFF},
		{R: 0x53, G: 0x90, B: 0x0F, A: 0xFF},
		{R: 0xA4, G: 0xA7, B: 0x1E, A: 0xFF},
		{R: 0xD6, G: 0xCE, B: 0x15, A: 0xFF},
	}

	Cheerful = []color.RGBA{
		{R: 0xFB, G: 0xE8, B: 0xA6, A: 0xFF},
		{R: 0xF4, G: 0x97, B: 0x6C, A: 0xFF},
		{R: 0x30, G: 0x3C, B: 0x6C, A: 0xFF},
		{R: 0xB4, G: 0xDF, B: 0xE5, A: 0xFF},
		{R: 0xD2, G: 0xFD, B: 0xFF, A: 0xFF},
	}

	Earthy = []color.RGBA{
		{R: 0x8D, G: 0x87, B: 0x41, A: 0xFF},
		{R: 0xF4, G: 0x97, B: 0x6C, A: 0xFF},
		{R: 0x30, G: 0x3C, B: 0x6C, A: 0xFF},
		{R: 0xB4, G: 0xDF, B: 0xE5, A: 0xFF},
		{R: 0xD2, G: 0xFD, B: 0xFF, A: 0xFF},
	}

	Plasma = []color.RGBA{
		{R: 0x0D, G: 0x08, B: 0x87, A: 0xFF},
		{R: 0x5B, G: 0x02, B: 0xA3, A: 0xFF},
		{R: 0x9A, G: 0x17, B: 0x9B, A: 0xFF},
		{R: 0xCB, G: 0x46, B: 0x79, A: 0xFF},
		{R: 0xEB, G: 0x78, B: 0x52, A: 0xFF},
		{R: 0xFB, G: 0xB3, B: 0x2F, A: 0xFF},
		{R: 0xF0, G: 0xF9, B: 0x21, A: 0xFF},
	}

	Viridis = []color.RGBA{
		{R: 0x44, G: 0x01, B: 0x54, A: 0xFF},
		{R: 0x46, G: 0x32, B: 0x7E, A: 0xFF},
		{R: 0x36, G: 0x5C, B: 0x8D, A: 0xFF},
		{R: 0x27, G: 0x7F, B: 0x8E, A: 0xFF},
		{R: 0x1F, G: 0xA1, B: 0x87, A: 0xFF},
		{R: 0x4A, G: 0xC1, B: 0x6D, A: 0xFF},
		{R: 0x9F, G: 0xDA, B: 0x3A, A: 0xFF},
		{R: 0xFD, G: 0xE7, B: 0x25, A: 0xFF},
	}

	Inferno = []color.RGBA{
		{R: 0x00, G: 0x00, B: 0x04, A: 0xFF},
		{R: 0x28, G: 0x0B, B: 0x53, A: 0xFF},
		{R: 0x65, G: 0x15, B: 0x6E, A: 0xFF},
		{R: 0x9F, G: 0x2A, B: 0x63, A: 0xFF},
		{R: 0xD4, G: 0x48, B: 0x42, A: 0xFF},
		{R: 0xF5, G: 0x7D, B: 0x15, A: 0xFF},
		{R: 0xFA, G: 0xC2, B: 0x28, A: 0xFF},
		{R: 0xFC, G: 0xFF, B: 0xA4, A: 0xFF},
	}

	Magma = []color.RGBA{
		{R: 0x00, G: 0x00, B: 0x04, A: 0xFF},
		{R: 0x2C, G: 0x11, B: 0x5F, A: 0xFF},
		{R: 0x72, G: 0x1F, B: 0x81, A: 0xFF},
		{R: 0xB6, G: 0x36, B: 0x79, A: 0xFF},
		{R: 0xF1, G: 0x60, B: 0x5D, A: 0xFF},
		{R: 0xFE, G: 0xAF, B: 0x77, A: 0xFF},
		{R: 0xFC, G: 0xFB, B: 0xBF, A: 0xFF},
	}

	Sunset = []color.RGBA{
		{R: 0x26, G: 0x46, B: 0x53, A: 0xFF},
		{R: 0x2A, G: 0x9D, B: 0x8F, A: 0xFF},
		{R: 0xE9, G: 0xC4, B: 0x6A, A: 0xFF},
		{R: 0xF4, G: 0xA2, B: 0x61, A: 0xFF},
		{R: 0xE7, G: 0x6F, B: 0x51, A: 0xFF},
	}

	Oceanic = []color.RGBA{
		{R: 0x03, G: 0x04, B: 0x5E, A: 0xFF},
		{R: 0x02, G: 0x3E, B: 0x8A, A: 0xFF},
		{R: 0x00, G: 0x77, B: 0xB6, A: 0xFF},
		{R: 0x00, G: 0x96, B: 0xC7, A: 0xFF},
		{R: 0x00, G: 0xB4, B: 0xD8, A: 0xFF},
		{R: 0x48, G: 0xCA, B: 0xE4, A: 0xFF},
		{R: 0x90, G: 0xE0, B: 0xEF, A: 0xFF},
		{R: 0xCA, G: 0xF0, B: 0xF8, A: 0xFF},
	}

	Cyberpunk = []color.RGBA{
		{R: 0x0B, G: 0x00, B: 0x1A, A: 0xFF},
		{R: 0x24, G: 0x00, B: 0x46, A: 0xFF},
		{R: 0x72, G: 0x09, B: 0xB7, A: 0xFF},
		{R: 0xF7, G: 0x25, B: 0x85, A: 0xFF},
		{R: 0x43, G: 0x61, B: 0xEE, A: 0xFF},
		{R: 0x4C, G: 0xC9, B: 0xF0, A: 0xFF},
		{R: 0xFF, G: 0xEA, B: 0x00, A: 0xFF},
	}

	UkiyoE = []color.RGBA{
		{R: 0x1B, G: 0x36, B: 0x44, A: 0xFF},
		{R: 0x2C, G: 0x5E, B: 0x7A, A: 0xFF},
		{R: 0x78, G: 0x9D, B: 0xA6, A: 0xFF},
		{R: 0xE8, G: 0xDF, B: 0xCA, A: 0xFF},
		{R: 0xD9, G: 0x8A, B: 0x54, A: 0xFF},
		{R: 0xB8, G: 0x40, B: 0x30, A: 0xFF},
	}

	Dunhuang = []color.RGBA{
		{R: 0x8C, G: 0x2D, B: 0x19, A: 0xFF},
		{R: 0xD9, G: 0x79, B: 0x25, A: 0xFF},
		{R: 0xE8, G: 0xC5, B: 0x87, A: 0xFF},
		{R: 0x3A, G: 0x7D, B: 0x70, A: 0xFF},
		{R: 0x1E, G: 0x47, B: 0x59, A: 0xFF},
		{R: 0x4A, G: 0x6B, B: 0x4E, A: 0xFF},
	}

	WatercolorInk = []color.RGBA{
		{R: 0x16, G: 0x19, B: 0x20, A: 0xFF},
		{R: 0x2F, G: 0x3E, B: 0x46, A: 0xFF},
		{R: 0x52, G: 0x79, B: 0x6F, A: 0xFF},
		{R: 0x84, G: 0xA9, B: 0x8C, A: 0xFF},
		{R: 0xCA, G: 0xD2, B: 0xC5, A: 0xFF},
		{R: 0xF4, G: 0xF1, B: 0xEA, A: 0xFF},
	}

	Botanical = []color.RGBA{
		{R: 0x28, G: 0x36, B: 0x18, A: 0xFF},
		{R: 0x60, G: 0x6C, B: 0x38, A: 0xFF},
		{R: 0xA3, G: 0xB1, B: 0x8A, A: 0xFF},
		{R: 0xFE, G: 0xFA, B: 0xE0, A: 0xFF},
		{R: 0xDD, G: 0xA1, B: 0x5E, A: 0xFF},
		{R: 0xBC, G: 0x6C, B: 0x25, A: 0xFF},
	}

	PastelDream = []color.RGBA{
		{R: 0xFF, G: 0xAD, B: 0xAD, A: 0xFF},
		{R: 0xFF, G: 0xD6, B: 0xA5, A: 0xFF},
		{R: 0xFD, G: 0xFF, B: 0xB6, A: 0xFF},
		{R: 0xCA, G: 0xFF, B: 0xBF, A: 0xFF},
		{R: 0x9B, G: 0xF6, B: 0xFF, A: 0xFF},
		{R: 0xA0, G: 0xC4, B: 0xFF, A: 0xFF},
		{R: 0xBD, G: 0xB2, B: 0xFF, A: 0xFF},
		{R: 0xFF, G: 0xC6, B: 0xFF, A: 0xFF},
	}
)

// ParseHexColor parses color string likes #FFFFFF or #2398EFFF
func ParseHexColor(s string) (color.RGBA, error) {
	s = strings.TrimSpace(strings.ToLower(s))

	if !strings.HasPrefix(s, "#") {
		return Black, fmt.Errorf("invalid hex color string %v", s)
	}

	if c, ok := parseHex(s[1:]); ok {
		return c, nil
	}
	return Black, fmt.Errorf("invalid hex color string %v", s)
}

// parseHex returns a color.RGBA by parsing a hex string
// Reference: https://stackoverflow.com/questions/54197913/parse-hex-string-to-image-color
func parseHex(s string) (color.RGBA, bool) {
	c := color.RGBA{}
	c.A = 255
	ok := true

	hexToByte := func(b byte) byte {
		switch {
		case b >= '0' && b <= '9':
			return b - '0'
		case b >= 'a' && b <= 'f':
			return b - 'a' + 10
		}
		ok = false
		return 0
	}

	n := len(s)
	if n == 6 || n == 8 {
		c.R = hexToByte(s[0])<<4 + hexToByte(s[1])
		c.G = hexToByte(s[2])<<4 + hexToByte(s[3])
		c.B = hexToByte(s[4])<<4 + hexToByte(s[5])
		if n == 8 {
			c.A = hexToByte(s[6])<<4 + hexToByte(s[7])
		}
	} else if n == 3 || n == 4 {
		c.R = hexToByte(s[0]) * 17
		c.G = hexToByte(s[1]) * 17
		c.B = hexToByte(s[2]) * 17
		if n == 4 {
			c.A = hexToByte(s[3]) * 17
		}
	} else {
		ok = false
	}
	return c, ok
}

// WithAlpha returns a non-premultiplied color.NRGBA suitable for gg.SetColor.
func WithAlpha(c color.RGBA, alpha uint8) color.NRGBA {
	return color.NRGBA{R: c.R, G: c.G, B: c.B, A: alpha}
}

// PremultiplyRGBA scales R, G, B by alpha/255 so the premultiplied color.RGBA invariant (R,G,B <= A) holds.
func PremultiplyRGBA(c color.RGBA, alpha uint8) color.RGBA {
	a := float64(alpha) / 255.0
	return color.RGBA{
		R: uint8(math.Round(float64(c.R) * a)),
		G: uint8(math.Round(float64(c.G) * a)),
		B: uint8(math.Round(float64(c.B) * a)),
		A: alpha,
	}
}

func toColorful(c color.RGBA) colorful.Color {
	return colorful.Color{
		R: float64(c.R) / 255.0,
		G: float64(c.G) / 255.0,
		B: float64(c.B) / 255.0,
	}
}

func fromColorful(c colorful.Color, alpha uint8) color.RGBA {
	r, g, b := c.Clamped().RGB255()
	return color.RGBA{R: r, G: g, B: b, A: alpha}
}

func lerpAlpha(a1, a2 uint8, t float64) uint8 {
	return uint8(ConstrainInt(int(math.Round(float64(a1)*(1.0-t)+float64(a2)*t)), 0, 255))
}

// LerpColorOklab blends two colors in the perceptual Oklab color space (t=0 -> c1, t=1 -> c2).
func LerpColorOklab(c1, c2 color.RGBA, t float64) color.RGBA {
	t = Constrain(t, 0.0, 1.0)
	cf := toColorful(c1).BlendOkLab(toColorful(c2), t)
	return fromColorful(cf, lerpAlpha(c1.A, c2.A, t))
}

// LerpColorHCL blends two colors in the perceptual CIE-L*C*h° (HCL) color space (t=0 -> c1, t=1 -> c2).
func LerpColorHCL(c1, c2 color.RGBA, t float64) color.RGBA {
	t = Constrain(t, 0.0, 1.0)
	cf := toColorful(c1).BlendHcl(toColorful(c2), t)
	return fromColorful(cf, lerpAlpha(c1.A, c2.A, t))
}

// LerpColorLab blends two colors in the CIE-L*a*b* color space (t=0 -> c1, t=1 -> c2).
func LerpColorLab(c1, c2 color.RGBA, t float64) color.RGBA {
	t = Constrain(t, 0.0, 1.0)
	cf := toColorful(c1).BlendLab(toColorful(c2), t)
	return fromColorful(cf, lerpAlpha(c1.A, c2.A, t))
}

// LerpColorLinear blends two colors in gamma-correct linear RGB space (t=0 -> c1, t=1 -> c2).
func LerpColorLinear(c1, c2 color.RGBA, t float64) color.RGBA {
	t = Constrain(t, 0.0, 1.0)
	cf := toColorful(c1).BlendLinearRgb(toColorful(c2), t)
	return fromColorful(cf, lerpAlpha(c1.A, c2.A, t))
}

// GradientStop represents a color stop at position Pos in [0, 1].
type GradientStop struct {
	Pos   float64
	Color color.RGBA
}

// Gradient represents a multi-stop color ramp.
type Gradient struct {
	Stops []GradientStop
}

// NewGradient creates a Gradient with evenly spaced stops from the given colors.
func NewGradient(colors ...color.RGBA) *Gradient {
	n := len(colors)
	if n == 0 {
		return &Gradient{Stops: []GradientStop{{Pos: 0, Color: Black}, {Pos: 1, Color: White}}}
	}
	if n == 1 {
		return &Gradient{Stops: []GradientStop{{Pos: 0, Color: colors[0]}, {Pos: 1, Color: colors[0]}}}
	}
	stops := make([]GradientStop, n)
	for i, c := range colors {
		stops[i] = GradientStop{
			Pos:   float64(i) / float64(n-1),
			Color: c,
		}
	}
	return &Gradient{Stops: stops}
}

// NewGradientWithStops creates a Gradient from explicit stops, sorted by position.
func NewGradientWithStops(stops ...GradientStop) *Gradient {
	if len(stops) == 0 {
		return NewGradient(Black, White)
	}
	cp := make([]GradientStop, len(stops))
	copy(cp, stops)
	sort.Slice(cp, func(i, j int) bool {
		return cp[i].Pos < cp[j].Pos
	})
	return &Gradient{Stops: cp}
}

func (g *Gradient) sampleWith(t float64, blendFn func(c1, c2 color.RGBA, u float64) color.RGBA) color.RGBA {
	if len(g.Stops) == 0 {
		return Black
	}
	if len(g.Stops) == 1 {
		return g.Stops[0].Color
	}
	t = Constrain(t, 0.0, 1.0)
	if t <= g.Stops[0].Pos {
		return g.Stops[0].Color
	}
	last := g.Stops[len(g.Stops)-1]
	if t >= last.Pos {
		return last.Color
	}
	for i := 0; i < len(g.Stops)-1; i++ {
		s0, s1 := g.Stops[i], g.Stops[i+1]
		if t >= s0.Pos && t <= s1.Pos {
			span := s1.Pos - s0.Pos
			if span <= 1e-12 {
				return s1.Color
			}
			u := (t - s0.Pos) / span
			return blendFn(s0.Color, s1.Color, u)
		}
	}
	return last.Color
}

// SampleOklab samples the gradient at t in [0, 1] using Oklab interpolation.
func (g *Gradient) SampleOklab(t float64) color.RGBA {
	return g.sampleWith(t, LerpColorOklab)
}

// SampleHCL samples the gradient at t in [0, 1] using HCL interpolation.
func (g *Gradient) SampleHCL(t float64) color.RGBA {
	return g.sampleWith(t, LerpColorHCL)
}

// SampleLinear samples the gradient at t in [0, 1] using linear RGB interpolation.
func (g *Gradient) SampleLinear(t float64) color.RGBA {
	return g.sampleWith(t, LerpColorLinear)
}

// CosinePalette implements Inigo Quilez's procedural cosine color palette:
// color(t) = A + B * cos(2*pi*(C*t + D))
type CosinePalette struct {
	A, B, C, D [3]float64
}

var (
	CosineRainbow = CosinePalette{
		A: [3]float64{0.5, 0.5, 0.5},
		B: [3]float64{0.5, 0.5, 0.5},
		C: [3]float64{1.0, 1.0, 1.0},
		D: [3]float64{0.00, 0.33, 0.67},
	}
	CosineSunset = CosinePalette{
		A: [3]float64{0.5, 0.5, 0.5},
		B: [3]float64{0.5, 0.5, 0.5},
		C: [3]float64{1.0, 0.7, 0.4},
		D: [3]float64{0.00, 0.15, 0.20},
	}
	CosineCyber = CosinePalette{
		A: [3]float64{0.5, 0.5, 0.5},
		B: [3]float64{0.5, 0.5, 0.5},
		C: [3]float64{2.0, 1.0, 0.0},
		D: [3]float64{0.50, 0.20, 0.25},
	}
	CosineIridescent = CosinePalette{
		A: [3]float64{0.8, 0.5, 0.4},
		B: [3]float64{0.2, 0.4, 0.2},
		C: [3]float64{2.0, 1.0, 1.0},
		D: [3]float64{0.00, 0.25, 0.25},
	}
)

// Sample evaluates the cosine palette at parameter t.
func (cp CosinePalette) Sample(t float64) color.RGBA {
	var rgb [3]uint8
	for i := 0; i < 3; i++ {
		v := cp.A[i] + cp.B[i]*math.Cos(2.0*math.Pi*(cp.C[i]*t+cp.D[i]))
		v = Constrain(v, 0.0, 1.0)
		rgb[i] = uint8(math.Round(v * 255.0))
	}
	return color.RGBA{R: rgb[0], G: rgb[1], B: rgb[2], A: 255}
}

// GenerateSoftPalette generates n harmonious soft colors using go-colorful.
func GenerateSoftPalette(n int) []color.RGBA {
	if n <= 0 {
		return nil
	}
	cols, err := colorful.SoftPalette(n)
	if err != nil {
		res := make([]color.RGBA, n)
		for i := 0; i < n; i++ {
			res[i] = fromColorful(colorful.FastHappyColor(), 255)
		}
		return res
	}
	res := make([]color.RGBA, len(cols))
	for i, c := range cols {
		res[i] = fromColorful(c, 255)
	}
	return res
}

// GenerateWarmPalette generates n warm colors using go-colorful.
func GenerateWarmPalette(n int) []color.RGBA {
	if n <= 0 {
		return nil
	}
	cols, err := colorful.WarmPalette(n)
	if err != nil {
		res := make([]color.RGBA, n)
		for i := 0; i < n; i++ {
			res[i] = fromColorful(colorful.FastWarmColor(), 255)
		}
		return res
	}
	res := make([]color.RGBA, len(cols))
	for i, c := range cols {
		res[i] = fromColorful(c, 255)
	}
	return res
}