package arts

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"math"
)

type randomShape struct {
	shapeNum int
}

func NewRandomShape(shapeNum int) *randomShape {
	return &randomShape{
		shapeNum: shapeNum,
	}
}

// Generative draws a random shape image.
func (r *randomShape) Generative(c *generativeart.Canva) {
	ctex := gg.NewContextForRGBA(c.Img())
	rng := c.Rand()

	ctex.Translate(float64(c.Width()/2), float64(c.Height()/2))
	ctex.Rotate(common.RandomRangeFloat64From(rng, -1, 1) * math.Pi * 0.25)
	ctex.Translate(-float64(c.Width()/2), -float64(c.Height()/2))

	for i := 0; i < r.shapeNum; i++ {
		x := common.RandomGaussianFrom(rng, 0.5, 0.2) * float64(c.Width())
		y := common.RandomGaussianFrom(rng, 0.5, 0.2) * float64(c.Height())

		w := common.RandomRangeFloat64From(rng, 0, float64(c.Width())/3)*common.RandomRangeFloat64From(rng, 0, rng.Float64()) + 5.0
		h := w + common.RandomRangeFloat64From(rng, -1, 1)*3.0

		rnd := rng.Intn(4)
		theta := math.Pi * 2.0 * float64(rng.Intn(4)) / 4

		ctex.Push()
		ctex.Translate(x, y)
		ctex.Rotate(theta)
		ctex.SetColor(c.Opts().ColorSchema()[rng.Intn(len(c.Opts().ColorSchema()))])
		switch rnd {
		case 0:
			ctex.DrawCircle(0, 0, w/2)
		case 1:
			ctex.DrawRectangle(0, 0, w/2, w/2)
		case 2:
			if rng.Float64() < 0.5 {
				ctex.DrawEllipse(0, 0, w/2, h/2)
			} else {
				ctex.DrawRectangle(0, 0, w, h)
			}
		case 3:
			ctex.DrawRectangle(0, 0, w*2, common.RandomRangeFloat64From(rng, 2, 10))
		}
		ctex.Fill()
		ctex.Pop()
	}
}
