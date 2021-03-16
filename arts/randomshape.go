package arts

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"math"
	"math/rand"
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

	ctex.Translate(float64(c.Width()/2), float64(c.Height()/2))
	ctex.Rotate(common.RandomRangeFloat64(-1, 1) * math.Pi * 0.25)
	ctex.Translate(-float64(c.Width()/2), -float64(c.Height()/2))

	for i := 0; i < r.shapeNum; i++ {
		x := common.RandomGaussian(0.5, 0.2) * float64(c.Width())
		y := common.RandomGaussian(0.5, 0.2) * float64(c.Height())

		w := common.RandomRangeFloat64(0, float64(c.Width())/3)*common.RandomRangeFloat64(0, rand.Float64()) + 5.0
		h := w + common.RandomRangeFloat64(-1, 1)*3.0

		rnd := rand.Intn(4)
		theta := math.Pi * 2.0 * float64(rand.Intn(4)) / 4

		ctex.Push()
		ctex.Translate(x, y)
		ctex.Rotate(theta)
		ctex.SetColor(c.Opts().ColorSchema()[rand.Intn(len(c.Opts().ColorSchema()))])
		switch rnd {
		case 0:
			ctex.DrawCircle(0, 0, w/2)
		case 1:
			ctex.DrawRectangle(0, 0, w/2, w/2)
		case 2:
			if rand.Float64() < 0.5 {
				ctex.DrawEllipse(0, 0, w/2, h/2)
			} else {
				ctex.DrawRectangle(0, 0, w, h)
			}
		case 3:
			ctex.DrawRectangle(0, 0, w*2, common.RandomRangeFloat64(2, 10))
		}
		ctex.Fill()
		ctex.Pop()
	}
}
