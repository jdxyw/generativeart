package arts

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"math"
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
func (cc *colorCircle2) Generative(c *generativeart.Canva) {
	ctex := gg.NewContextForRGBA(c.Img())
	rng := c.Rand()

	for i := 0; i < cc.circleNum; i++ {
		x := common.RandomRangeFloat64From(rng, 0, float64(c.Width()))
		y := common.RandomRangeFloat64From(rng, 0, float64(c.Height()))

		r1 := common.RandomRangeFloat64From(rng, 50.0, float64(c.Width())/4)
		r2 := common.RandomRangeFloat64From(rng, 10.0, float64(c.Width())/3)

		cc.circle(ctex, c, x, y, r1, r2)
		if rng.Float64() < 0.3 {
			col := c.Opts().ColorSchema()[rng.Intn(len(c.Opts().ColorSchema()))]
			ctex.SetColor(col)
			ctex.DrawCircle(x, y, r1/2.0)
			ctex.Fill()
		}
	}
}

func (cc *colorCircle2) circle(ctex *gg.Context, c *generativeart.Canva, x, y, d, dx float64) {
	rng := c.Rand()
	col := c.Opts().ColorSchema()[rng.Intn(len(c.Opts().ColorSchema()))]

	for j := 0.0; j < dx; j += 1.0 {
		dd := d + j*2.0
		alpha := int((dx / j) * 255.0)
		if alpha > 255 || j == 0.0 {
			alpha = 255
		}
		if alpha < 0 {
			alpha = 0
		}

		ctex.SetColor(common.WithAlpha(col, uint8(alpha)))
		for i := 0; i < 150; i++ {
			theta := common.RandomRangeFloat64From(rng, 0, math.Pi*2)
			xx := x + dd*0.5*math.Cos(theta)
			yy := y + dd*0.5*math.Sin(theta)
			ctex.DrawPoint(xx, yy, 0.51)
			ctex.Fill()
		}
	}
}
