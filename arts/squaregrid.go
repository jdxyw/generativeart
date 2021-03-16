package arts

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart"
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
func (g *girdSquares) Generative(c *generativeart.Canva) {
	ctex := gg.NewContextForRGBA(c.Img())

	for x := 0; x < c.Width(); x += g.step {
		for y := 0; y < c.Height(); y += g.step {
			cl := c.Opts().ColorSchema()[rand.Intn(len(c.Opts().ColorSchema()))]

			x0 := float64(x)
			y0 := float64(y)
			s := float64(g.rectSize)

			theta := rand.Intn(360) + 1
			for i := 0; i < c.Opts().NIters(); i++ {
				ctex.Push()

				ctex.Translate(x0+float64(g.step/2), y0+float64(g.step/2))
				ctex.Rotate(gg.Radians(float64(theta * i)))

				ctex.Scale(s, s)

				ctex.LineTo(-0.5, 0.5)
				ctex.LineTo(0.5, 0.5)
				ctex.LineTo(0.5, -0.5)
				ctex.LineTo(-0.5, -0.5)
				ctex.LineTo(-0.5, 0.5)

				ctex.SetLineWidth(c.Opts().LineWidth())
				ctex.SetColor(c.Opts().LineColor())
				ctex.StrokePreserve()
				ctex.SetRGBA255(int(cl.R), int(cl.G), int(cl.B), c.Opts().Alpha())
				ctex.Fill()
				ctex.Pop()
				s = s - g.decay*float64(g.rectSize)
			}
		}
	}
}
