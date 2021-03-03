package generativeart

import (
	"github.com/fogleman/gg"
	"math"
)

type circleLoop struct {
	radius float64
}

func NewCircleLoop(radius float64) *circleLoop {
	return &circleLoop{
		radius: radius,
	}
}

// Generative draws a Circle Loop images.
func (cl *circleLoop) Generative(c *canva) {
	ctex := gg.NewContextForRGBA(c.img)

	r := cl.radius
	var theta float64 = 0
	for i := 0; i < c.opts.nIters; i++ {
		ctex.Push()
		ctex.Translate(float64(c.width/2), float64(c.height/2))
		x := cl.radius * math.Cos(gg.Radians(theta))
		y := cl.radius * math.Sin(gg.Radians(theta*2))

		ctex.SetLineWidth(c.opts.lineWidth)
		ctex.SetColor(c.opts.lineColor)
		ctex.SetRGBA255(int(c.opts.lineColor.R), int(c.opts.lineColor.G), int(c.opts.lineColor.B), c.opts.alpha)
		ctex.DrawEllipse(x, y, r/2, r/2)
		ctex.Stroke()
		ctex.Pop()
		r += math.Cos((theta))*math.Sin((theta/2)) + math.Sin((theta))*math.Cos((theta/2))
		theta += math.Pi / 2
	}
}
