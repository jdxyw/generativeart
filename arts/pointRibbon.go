package arts

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart"
	"math"
)

type pointRibbon struct {
	r float64
}

// NewPointRibbon returns a pointRibbon object.
func NewPointRibbon(r float64) *pointRibbon {
	return &pointRibbon{
		r: r,
	}
}

// Generative draws a point ribbon image.
// TODO: make the point as parameters.
func (s *pointRibbon) Generative(c *generativeart.Canva) {
	ctex := gg.NewContextForRGBA(c.Img())
	ctex.SetLineWidth(c.Opts().LineWidth())

	var t float64
	var dt = 0.0001
	for i := 0; i < c.Opts().NIters(); i++ {
		delta := 2.0*s.r*math.Cos(4.0*dt*t) + s.r*math.Cos(t)
		ctex.SetRGBA255(int(delta), int(2*s.r*math.Sin(t)-s.r*math.Cos(3*dt*t)), 100, 10)
		ctex.DrawPoint(2*s.r*math.Sin(2*t*dt)+s.r*math.Cos(t*dt)+float64(c.Width()/2),
			2*s.r*math.Sin(t*dt)-s.r*math.Sin(5*t)+float64(c.Height()/2), 1.0)
		ctex.Stroke()
		t += 0.01
		dt += 0.1
	}
}
