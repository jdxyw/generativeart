package arts

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"math"
	"math/rand"
)

type colorCircle struct {
	circleNum int
}

func NewColorCircle(circleNum int) *colorCircle {
	return &colorCircle{
		circleNum: circleNum,
	}
}

// Generative draws a color circle images.
func (cc *colorCircle) Generative(c *generativeart.Canva) {
	ctex := gg.NewContextForRGBA(c.Img())

	for i := 0; i < cc.circleNum; i++ {
		rnd := rand.Intn(3)
		x := common.RandomRangeFloat64(-0.1, 1.1) * float64(c.Width())
		y := common.RandomRangeFloat64(-0.1, 1.1) * float64(c.Height())
		s := common.RandomRangeFloat64(0, common.RandomRangeFloat64(0, float64(c.Width()/2))) + 10
		if rnd == 2 {
			rnd = rand.Intn(3)
		}
		switch rnd {
		case 0:
			cc.drawCircleV1(ctex, c, x, y, s)
		case 1:
			ctex.SetLineWidth(common.RandomRangeFloat64(0, 1))
			ctex.SetColor(c.Opts().ColorSchema()[rand.Intn(len(c.Opts().ColorSchema()))])
			ctex.DrawCircle(x, y, common.RandomRangeFloat64(0, s)/2)
			ctex.Stroke()
		case 2:
			cc.drawCircleV2(ctex, c, x, y, s)
		}
	}
}

func (cc *colorCircle) drawCircleV1(ctex *gg.Context, c *generativeart.Canva, x, y, s float64) {
	n := common.RandomRangeInt(4, 30)
	cs := common.RandomRangeFloat64(2, 8)
	ctex.SetColor(c.Opts().ColorSchema()[rand.Intn(len(c.Opts().ColorSchema()))])
	ctex.Push()
	ctex.Translate(x, y)
	for a := 0.0; a < math.Pi*2.0; a += math.Pi * 2.0 / float64(n) {
		ctex.DrawEllipse(s*0.5*math.Cos(a), s*0.5*math.Sin(a), cs/2, cs/2)
		ctex.Fill()
	}
	ctex.Pop()
}

func (cc *colorCircle) drawCircleV2(ctex *gg.Context, c *generativeart.Canva, x, y, s float64) {
	cl := c.Opts().ColorSchema()[rand.Intn(len(c.Opts().ColorSchema()))]
	ctex.SetLineWidth(1.0)
	sx := s * common.RandomRangeFloat64(0.1, 0.55)
	for j := 0.0001; j < sx; j++ {
		dd := s + j*2.0
		alpha := int(255 * sx / j)
		if alpha > 255 {
			alpha = 255
		}

		if alpha < 0 {
			alpha = 0
		}

		//alpha := RandomRangeInt(30, 150)
		cl.A = uint8(alpha)
		ctex.SetColor(cl)

		for i := 0; i < 200; i++ {
			theta := common.RandomRangeFloat64(0, math.Pi*2)
			xx := x + dd*0.3*math.Cos(theta)
			yy := y + dd*0.3*math.Sin(theta)
			//ctex.DrawLine(xx, yy, xx, yy)
			ctex.DrawPoint(xx, yy, 0.6)
			ctex.Stroke()
		}
	}
}
