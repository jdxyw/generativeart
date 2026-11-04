package arts

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
)

// TODO: figure out what the meaning of these parameters.
type dotLine struct {
	n         int
	ras       float64
	canv      float64
	randColor bool
}

func NewDotLine(n int, ras, canv float64, randColor bool) *dotLine {
	return &dotLine{
		n:         n,
		ras:       ras,
		canv:      canv,
		randColor: randColor,
	}
}

func (d *dotLine) Generative(c *generativeart.Canva) {
	ctex := gg.NewContextForRGBA(c.Img())

	rng := c.Rand()
	ctex.SetLineWidth(c.Opts().LineWidth())
	var dir []int = []int{-1, 1}
	for i := 0; i < c.Opts().NIters(); i++ {
		oldx := rng.Intn(d.n - 1)
		oldy := rng.Intn(d.n - 1)

		n := rng.Intn(7)
		if d.randColor {
			ctex.SetColor(c.Opts().ColorSchema()[rng.Intn(len(c.Opts().ColorSchema()))])
		} else {
			ctex.SetRGBA255(common.RandomRangeIntFrom(rng, 222, 255), common.RandomRangeIntFrom(rng, 20, 222), 0, 255)
		}
		for j := 0; j < n; j++ {
			newx := oldx + dir[rng.Intn(2)]
			newy := oldy + dir[rng.Intn(2)]

			if common.Distance(float64(newx), float64(newy), float64(d.n/2), float64(d.n/2)) > float64(d.n/2-10) {
				newx = oldx
				newy = oldy
			}
			if newx == oldx && rng.Intn(6) > 4 {
				ctex.DrawEllipse(float64(oldx)*d.ras+d.canv, float64(oldy)*d.ras+d.canv, c.Opts().LineWidth(), c.Opts().LineWidth())
				ctex.Fill()
				continue
			}
			ctex.DrawLine(float64(oldx)*d.ras+d.canv, float64(oldy)*d.ras+d.canv, float64(newx)*d.ras+d.canv, float64(newy)*d.ras+d.canv)
			oldx = newx
			oldy = newy

			ctex.Stroke()
		}
	}

}
