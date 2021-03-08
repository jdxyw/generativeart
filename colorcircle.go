package generativeart

import (
	"github.com/fogleman/gg"
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
func (cc *colorCircle) Generative(c *canva) {
	ctex := gg.NewContextForRGBA(c.img)

	for i := 0; i < cc.circleNum; i++ {
		rnd := rand.Intn(3)
		x := RandomRangeFloat64(-0.1, 1.1) * float64(c.width)
		y := RandomRangeFloat64(-0.1, 1.1) * float64(c.height)
		s := RandomRangeFloat64(0, RandomRangeFloat64(0, float64(c.width/2))) + 10
		if rnd == 2 {
			rnd = rand.Intn(3)
		}
		switch rnd {
		case 0:
			cc.drawCircleV1(ctex, c, x, y, s)
		case 1:
			ctex.SetLineWidth(RandomRangeFloat64(0, 1))
			ctex.SetColor(c.opts.colorSchema[rand.Intn(len(c.opts.colorSchema))])
			ctex.DrawCircle(x, y, RandomRangeFloat64(0, s)/2)
			ctex.Stroke()
		case 2:
			cc.drawCircleV2(ctex, c, x, y, s)
		}
	}
}

func (cc *colorCircle) drawCircleV1(ctex *gg.Context, c *canva, x, y, s float64) {
	n := RandomRangeInt(4, 30)
	cs := RandomRangeFloat64(2, 8)
	ctex.SetColor(c.opts.colorSchema[rand.Intn(len(c.opts.colorSchema))])
	ctex.Push()
	ctex.Translate(x, y)
	for a := 0.0; a < math.Pi*2.0; a += math.Pi * 2.0 / float64(n) {
		ctex.DrawEllipse(s*0.5*math.Cos(a), s*0.5*math.Sin(a), cs/2, cs/2)
		ctex.Fill()
	}
	ctex.Pop()
}

func (cc *colorCircle) drawCircleV2(ctex *gg.Context, c *canva, x, y, s float64) {
	cl := c.opts.colorSchema[rand.Intn(len(c.opts.colorSchema))]
	ctex.SetLineWidth(1.0)
	sx := s * RandomRangeFloat64(0.1, 0.55)
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
			theta := RandomRangeFloat64(0, math.Pi*2)
			xx := x + dd*0.3*math.Cos(theta)
			yy := y + dd*0.3*math.Sin(theta)
			//ctex.DrawLine(xx, yy, xx, yy)
			ctex.DrawPoint(xx, yy, 0.6)
			ctex.Stroke()
		}
	}
}
