package arts

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
)

type domainWrap struct {
	noise            *common.PerlinNoise
	scale            float64
	xOffset, yOffset float64
}

// NewDomainWrap returns a domainWrap object.
func NewDomainWrap(scale, xOffset, yOffset float64) *domainWrap {
	return &domainWrap{
		scale:   scale,
		xOffset: xOffset,
		yOffset: yOffset,
		noise:   common.NewPerlinNoise(),
	}
}

// Generative draws a domain warp image.
// Reference: https://www.iquilezles.org/www/articles/warp/warp.htm
func (d *domainWrap) Generative(c *generativeart.Canva) {
	_ = gg.NewContextForRGBA(c.Img())

	for h := 0.0; h < float64(c.Height()); h += 1.0 {
		for w := 0.0; h < float64(c.Width()); h += 1.0 {
			_, _, _ = d.pattern(w*d.scale, h*d.scale, d.xOffset, d.yOffset)
		}
	}
}

func (d *domainWrap) pattern(x, y, xOffest, yOffset float64) (float64, float64, float64) {
	qx := d.fbm(x+xOffest, y+yOffset)
	qy := d.fbm(x+xOffest+5.2, y+yOffset+1.3)

	rx := d.fbm(x+4.0*qx+1.7, y+4.0*qy+9.2)
	ry := d.fbm(x+4.0*qx+8.3, y+4.0*qy+2.8)

	return d.fbm(qx+4*rx, qy+4*ry), common.Magnitude(qx, qy), common.Magnitude(rx, ry)
}

func (d *domainWrap) fbm(x, y float64) float64 {
	return d.noise.Noise2D(x, y)
}
