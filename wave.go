package generativeart

import (
	"image/color"
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

// GenerativePolar draws the image in polar coordinate system.
func (w *wave) GenerativePolar(c *canva, rgba color.RGBA) {
	ctex := gg.NewContextForRGBA(c.img)
	for x := -math.Pi; x <= math.Pi; x += 0.01 {
		for y := -math.Pi; y <= math.Pi; y += 0.01 {
			xi, yi := w.fn(x, y)
			i, j := ConvertCartesianToPolarPixel(xi, yi, c.xaixs, c.yaixs, c.height, c.width)
			if i < 0 || i > c.width-1 || j < 0 || j > c.height-1 {
				continue
			}
			ctex.DrawCircle(float64(i), float64(j), 1.0)
			ctex.SetColor(rgba)
			ctex.Fill()
			//c.img.Set(i, j, rgba)
		}
	}
}

// Generative draws the image in cartesian coordinate system.
func (w *wave) Generative(c *canva, rgba color.RGBA) {
	for x := -math.Pi; x <= math.Pi; x += 0.002 {
		for y := -math.Pi; y <= math.Pi; y += 0.002 {
			xi, yi := w.fn(x, y)
			i, j := ConvertCartesianToPixel(xi, yi, c.xaixs, c.yaixs, c.height, c.width)
			if i < 0 || i > c.width-1 || j < 0 || j > c.height-1 {
				continue
			}
			c.img.Set(i, j, rgba)
		}
	}
}
