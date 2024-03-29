package common

import (
	"fmt"
	"image/color"
	"strings"
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
		c.R = hexToByte(s[0])<<4+hexToByte(s[1])
		c.G = hexToByte(s[2])<<4+hexToByte(s[3])
		c.B = hexToByte(s[4])<<4+hexToByte(s[5])
		if n == 8 {
			c.A = hexToByte(s[6])<<4+hexToByte(s[7])
		}
	} else if n == 3 || n == 4 {
		c.R = hexToByte(s[0])*17
		c.G = hexToByte(s[1])*17
		c.B = hexToByte(s[2])*17
		if n == 4 {
			c.A = hexToByte(s[3])*17
		}
	} else {
		ok = false
	}
	return c, ok
}