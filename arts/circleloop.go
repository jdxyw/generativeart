package arts

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart"
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
func (cl *circleLoop) Generative(c *generativeart.Canva) {
	ctex := gg.NewContextForRGBA(c.Img())

	r := cl.radius
	var theta float64 = 0
	for i := 0; i < c.Opts().NIters(); i++ {
		ctex.Push()
		ctex.Translate(float64(c.Width()/2), float64(c.Height()/2))
		x := cl.radius * math.Cos(gg.Radians(theta))
		y := cl.radius * math.Sin(gg.Radians(theta*2))

		ctex.SetLineWidth(c.Opts().LineWidth())
		ctex.SetColor(c.Opts().LineColor())
		ctex.SetRGBA255(int(c.Opts().LineColor().R), int(c.Opts().LineColor().G), int(c.Opts().LineColor().B), c.Opts().Alpha())
		ctex.DrawEllipse(x, y, r/2, r/2)
		ctex.Stroke()
		ctex.Pop()
		r += math.Cos((theta))*math.Sin((theta/2)) + math.Sin((theta))*math.Cos((theta/2))
		theta += math.Pi / 2
	}
}
