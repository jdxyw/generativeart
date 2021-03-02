package generativeart

import (
	"math"

	"github.com/fogleman/gg"
)

type Formula func(x, y float64) (float64, float64)

type wave struct {
	fn Formula
}

// NewWave returns a wave generator.
func NewWave(fn Formula) *wave {
	return &wave{fn: fn}
}

// Generative draws the image.
func (w *wave) Generative(c *canva) {
	ctex := gg.NewContextForRGBA(c.img)
	for x := -math.Pi; x <= math.Pi; x += 0.01 {
		for y := -math.Pi; y <= math.Pi; y += 0.01 {
			xi, yi := w.fn(x, y)
			var i, j int
			if c.opts.isPolarCoodinate {
				i, j = ConvertCartesianToPolarPixel(xi, yi, c.xaixs, c.yaixs, c.height, c.width)
			} else {
				i, j = ConvertCartesianToPixel(xi, yi, c.xaixs, c.yaixs, c.height, c.width)
			}

			if i < 0 || i > c.width-1 || j < 0 || j > c.height-1 {
				continue
			}

			ctex.DrawCircle(float64(i), float64(j), c.opts.radius)
			ctex.SetColor(c.opts.foreground)
			ctex.Fill()
		}
	}
}
