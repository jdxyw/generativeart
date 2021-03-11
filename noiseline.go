package generativeart

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart/common"
	"math"
	"math/rand"
)

type noiseLine struct {
	n int
}

// NewNoiseLine returns a noiseLine object.
func NewNoiseLine(n int) *noiseLine {
	return &noiseLine{
		n: n,
	}
}

// Generative draws a noise line image.
func (nl *noiseLine) Generative(c *canva) {
	ctex := gg.NewContextForRGBA(c.img)
	noise := common.NewPerlinNoise()

	ctex.SetColor(Black)
	for i := 0; i < 80; i++ {
		x := rand.Float64() * float64(c.width)
		y := rand.Float64() * float64(c.height)

		s := rand.Float64() * float64(c.width) / 8
		ctex.SetLineWidth(0.5)
		ctex.DrawEllipse(x, y, s, s)
		ctex.Stroke()
	}

	t := rand.Float64() * 10
	for i := 0; i < nl.n; i++ {
		x := common.RandomRangeFloat64(-0.5, 1.5) * float64(c.width)
		y := common.RandomRangeFloat64(-0.5, 1.5) * float64(c.height)
		cl := c.opts.colorSchema[rand.Intn(len(c.opts.colorSchema))]
		cl.A = uint8(c.opts.alpha)

		l := 400
		for j := 0; j < l; j++ {
			var ns = 0.0005
			w := math.Sin(math.Pi*float64(j)/float64(l-1)) * 5
			theta := noise.Noise3D(x*ns, y*ns, t) * 100
			ctex.SetColor(cl)
			ctex.DrawCircle(x, y, w)
			ctex.Fill()
			x += math.Cos(theta)
			y += math.Sin(theta)
			t += 0.0000003
		}
	}
}
