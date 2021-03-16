package arts

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"math/rand"
)

type sileSmoke struct {
	maxCircle         int
	maxStepsPerCircle int
	minSteps          float64
	maxSteps          float64
	minRadius         float64
	maxRadius         float64
	isRandColor       bool
}

func NewSilkSmoke(mc, msp int, minStep, maxStep, minRadius, maxRadius float64, isRandColor bool) *sileSmoke {
	return &sileSmoke{
		maxCircle:         mc,
		maxStepsPerCircle: msp,
		minSteps:          minStep,
		maxSteps:          maxStep,
		minRadius:         minRadius,
		maxRadius:         maxRadius,
		isRandColor:       isRandColor,
	}
}

// Generative draws a silk smoke image.
func (s *sileSmoke) Generative(c *generativeart.Canva) {
	ctex := gg.NewContextForRGBA(c.Img())

	cn := rand.Intn(s.maxCircle) + int(s.maxCircle/3)
	circles := newCircleSlice(cn, c.Width(), c.Height(), s.minSteps, s.maxSteps, s.minRadius, s.maxRadius)

	for i := 0; i < s.maxStepsPerCircle; i++ {
		ctex.SetRGBA255(0, 0, 0, 5)
		ctex.DrawRectangle(0, 0, float64(c.Width()), float64(c.Height()))
		ctex.Fill()

		for _, c1 := range circles {
			for _, c2 := range circles {

				cl := c.Opts().LineColor()
				if s.isRandColor {
					cl = c.Opts().ColorSchema()[rand.Intn(len(c.Opts().ColorSchema()))]
				}

				if c1 == c2 {
					continue
				}

				distance := common.Distance(c1.x, c1.y, c2.x, c2.y)

				if distance <= c1.radius+c2.radius {
					cx := (c1.x + c2.x) / 2
					cy := (c1.y + c2.y) / 2

					ctex.SetRGBA255(int(cl.R), int(cl.G), int(cl.B), c.Opts().Alpha())
					ctex.SetLineWidth(c.Opts().LineWidth())

					ctex.LineTo(c1.x, c1.y)
					ctex.LineTo(c2.x, c2.y)
					ctex.LineTo(cx, cy)
					ctex.LineTo(c1.x, c1.y)

					ctex.Stroke()
				}
			}
		}

		circles = circleSliceUpdate(circles, c.Width(), c.Height())
	}
}
