package generativeart

import (
	"github.com/fogleman/gg"
	"math"
	"math/rand"
)

type colorCircle2 struct {
	circleNum int
}

// NewColorCircle2 returns a colorCircle2 object.
func NewColorCircle2(circleNum int) *colorCircle2 {
	return &colorCircle2{
		circleNum: circleNum,
	}
}

// Generative draws a color circle image.
func (cc *colorCircle2) Generative(c *canva) {
	ctex := gg.NewContextForRGBA(c.img)

	for i := 0; i < cc.circleNum; i++ {
		x := RandomRangeFloat64(0, float64(c.width))
		y := RandomRangeFloat64(0, float64(c.height))

		r1 := RandomRangeFloat64(50.0, float64(c.width)/4)
		r2 := RandomRangeFloat64(10.0, float64(c.width)/3)

		cc.circle(ctex, c, x, y, r1, r2)
		if rand.Float64() < 0.3 {
			col := c.opts.colorSchema[rand.Intn(len(c.opts.colorSchema))]
			ctex.SetColor(col)
			ctex.DrawCircle(x, y, r1/2.0)
			ctex.Fill()
		}
	}
}

func (cc *colorCircle2) circle(ctex *gg.Context, c *canva, x, y, d, dx float64) {
	col := c.opts.colorSchema[rand.Intn(len(c.opts.colorSchema))]

	for j := 0.0; j < dx; j += 1.0 {
		dd := d + j*2.0
		alpha := int((dx / j) * 255.0)
		if alpha > 255 {
			alpha = 255
		}
		col.A = uint8(alpha)

		ctex.SetColor(col)
		for i := 0; i < 150; i++ {
			theta := RandomRangeFloat64(0, math.Pi*2)
			xx := x + dd*0.5*math.Cos(theta)
			yy := y + dd*0.5*math.Sin(theta)
			ctex.DrawPoint(xx, yy, 0.51)
			ctex.Fill()
		}
	}
}
