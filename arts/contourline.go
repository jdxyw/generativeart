package arts

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"math"
)

type contourLine struct {
	lineNum int
}

// NewContourLine returns a contourLine object.
func NewContourLine(lineNum int) *contourLine {
	return &contourLine{
		lineNum: lineNum,
	}
}

// Generative draws a contour line image.
func (cl *contourLine) Generative(c *generativeart.Canva) {
	ctex := gg.NewContextForRGBA(c.Img())
	noise := common.NewPerlinNoiseWithSeed(c.Seed())
	rng := c.Rand()
	for i := 0; i < cl.lineNum; i++ {
		cls := c.Opts().ColorSchema()[rng.Intn(len(c.Opts().ColorSchema()))]
		x := rng.Float64() * float64(c.Width())
		y := rng.Float64() * float64(c.Height())

		for j := 0; j < 1500; j++ {

			theta := noise.Noise2D(x/800.0, y/800.0) * math.Pi * 2 * 800
			x += math.Cos(theta) * 0.4
			y += math.Sin(theta) * 0.4

			ctex.SetColor(cls)
			ctex.DrawEllipse(x, y, 2, 2)
			ctex.Fill()

			if x > float64(c.Width()) || x < 0 || y > float64(c.Height()) || y < 0 || rng.Float64() < 0.001 {
				x = rng.Float64() * float64(c.Width())
				y = rng.Float64() * float64(c.Height())
			}
		}
	}
}
