package generativeart

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart/common"
	"math"
	"math/rand"
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
func (cl *contourLine) Generative(c *canva) {
	ctex := gg.NewContextForRGBA(c.img)
	//noise := perlin.NewPerlin(2, 2, 3, rand.Int63())
	noise := common.NewPerlinNoise()
	for i := 0; i < cl.lineNum; i++ {
		cl := c.opts.colorSchema[rand.Intn(len(c.opts.colorSchema))]
		x := rand.Float64() * float64(c.width)
		y := rand.Float64() * float64(c.height)

		for j := 0; j < 1500; j++ {

			theta := noise.Noise(x/800.0, y/800.0, 0) * math.Pi * 2 * 800
			x += math.Cos(theta) * 0.4
			y += math.Sin(theta) * 0.4

			ctex.SetColor(cl)
			ctex.DrawEllipse(x, y, 2, 2)
			ctex.Fill()

			if x > float64(c.width) || x < 0 || y > float64(c.height) || y < 0 || rand.Float64() < 0.001 {
				x = rand.Float64() * float64(c.width)
				y = rand.Float64() * float64(c.height)
			}
		}
	}
}
