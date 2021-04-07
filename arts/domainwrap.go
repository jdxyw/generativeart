package arts

import (
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"image/color"
)

// ColorMapping maps some parameters to color space.
type ColorMapping func(float64, float64, float64) color.RGBA

type domainWrap struct {
	noise            *common.PerlinNoise
	scale            float64
	scale2           float64
	xOffset, yOffset float64
	fn               ColorMapping
}

// NewDomainWrap returns a domainWrap object.
func NewDomainWrap(scale, scale2, xOffset, yOffset float64, cmap ColorMapping) *domainWrap {
	return &domainWrap{
		scale:   scale,
		scale2:  scale2,
		xOffset: xOffset,
		yOffset: yOffset,
		noise:   common.NewPerlinNoise(),
		fn:      cmap,
	}
}

// Generative draws a domain warp image.
// Reference: https://www.iquilezles.org/www/articles/warp/warp.htm
func (d *domainWrap) Generative(c *generativeart.Canva) {
	for h := 0.0; h < float64(c.Height()); h += 1.0 {
		for w := 0.0; w < float64(c.Width()); w += 1.0 {
			r, m1, m2 := d.pattern(w*d.scale, h*d.scale, d.xOffset, d.yOffset)
			rgb := d.fn(r, m1, m2)
			c.Img().Set(int(w), int(h), rgb)
		}
	}
}

func (d *domainWrap) pattern(x, y, xOffest, yOffset float64) (float64, float64, float64) {
	qx := d.fbm(x+xOffest, y+yOffset)
	qy := d.fbm(x+xOffest+5.2, y+yOffset+1.3)

	rx := d.fbm(x+d.scale2*qx+1.7, y+d.scale2*qy+9.2)
	ry := d.fbm(x+d.scale2*qx+8.3, y+d.scale2*qy+2.8)

	return d.fbm(qx+d.scale2*rx, qy+d.scale2*ry), common.Magnitude(qx, qy), common.Magnitude(rx, ry)
}

func (d *domainWrap) fbm(x, y float64) float64 {
	return d.noise.Noise2D(x, y)
}
