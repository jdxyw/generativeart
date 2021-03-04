package generativeart

import (
	"github.com/fogleman/gg"
	"math/rand"
)

type girdSquares struct {
	step, rectSize int
	decay          float64
}

// NewGirdSquares returns a grid squares generator.
func NewGirdSquares(step, rectSize int, decay float64) *girdSquares {
	return &girdSquares{
		step:     step,
		rectSize: rectSize,
		decay:    decay,
	}
}

// Generative draws a grid squares image.
func (g *girdSquares) Generative(c *canva) {
	ctex := gg.NewContextForRGBA(c.img)

	for x := 0; x < c.width; x += g.step {
		for y := 0; y < c.height; y += g.step {
			cl := c.opts.colorSchema[rand.Intn(len(c.opts.colorSchema))]

			x0 := float64(x)
			y0 := float64(y)
			s := float64(g.rectSize)

			theta := rand.Intn(360) + 1
			for i := 0; i < c.opts.nIters; i++ {
				ctex.Push()

				ctex.Translate(x0+float64(g.step/2), y0+float64(g.step/2))
				ctex.Rotate(gg.Radians(float64(theta * i)))

				ctex.Scale(s, s)

				ctex.LineTo(-0.5, 0.5)
				ctex.LineTo(0.5, 0.5)
				ctex.LineTo(0.5, -0.5)
				ctex.LineTo(-0.5, -0.5)
				ctex.LineTo(-0.5, 0.5)

				ctex.SetLineWidth(c.opts.lineWidth)
				ctex.SetColor(c.opts.lineColor)
				ctex.StrokePreserve()
				ctex.SetRGBA255(int(cl.R), int(cl.G), int(cl.B), c.opts.alpha)
				ctex.Fill()
				ctex.Pop()
				s = s - g.decay*float64(g.rectSize)
			}
		}
	}
}
