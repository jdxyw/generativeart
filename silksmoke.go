package generativeart

import (
	"github.com/fogleman/gg"
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
func (s *sileSmoke) Generative(c *canva) {
	ctex := gg.NewContextForRGBA(c.img)

	cn := rand.Intn(s.maxCircle) + int(s.maxCircle/3)
	circles := newCircleSlice(cn, c.width, c.height, s.minSteps, s.maxSteps, s.minRadius, s.maxRadius)

	for i := 0; i < s.maxStepsPerCircle; i++ {
		ctex.SetRGBA255(0, 0, 0, 5)
		ctex.DrawRectangle(0, 0, float64(c.width), float64(c.height))
		ctex.Fill()

		for _, c1 := range circles {
			for _, c2 := range circles {

				cl := c.opts.lineColor
				if s.isRandColor {
					cl = c.opts.colorSchema[rand.Intn(len(c.opts.colorSchema))]
				}

				if c1 == c2 {
					continue
				}

				distance := Distance(c1.x, c1.y, c2.x, c2.y)

				if distance <= c1.radius+c2.radius {
					cx := (c1.x + c2.x) / 2
					cy := (c1.y + c2.y) / 2

					ctex.SetRGBA255(int(cl.R), int(cl.G), int(cl.B), 30)
					ctex.SetLineWidth(c.opts.lineWidth)

					ctex.LineTo(c1.x, c1.y)
					ctex.LineTo(c2.x, c2.y)
					ctex.LineTo(cx, cy)
					ctex.LineTo(c1.x, c1.y)

					ctex.Stroke()
				}
			}
		}

		circles = circleSliceUpdate(circles, c.width, c.height)
	}
}
