package arts

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"math"
)

type blackHole struct {
	circleN   int
	density   float64
	circleGap float64
}

// NewBlackHole returns a blackhole object.
func NewBlackHole(circleN int, density, circleGap float64) *blackHole {
	return &blackHole{
		circleN:   circleN,
		density:   density,
		circleGap: circleGap,
	}
}

// Generative draws a black hole image.
func (b *blackHole) Generative(c *generativeart.Canva) {
	ctex := gg.NewContextForRGBA(c.Img())
	noise := common.NewPerlinNoise()
	kMax := common.RandomRangeFloat64(0.5, 1)
	ctex.SetLineWidth(0.4)
	ctex.SetColor(c.Opts().LineColor())

	for i := 0; i < b.circleN; i++ {
		radius := float64(c.Width()/10) + float64(i)*0.05
		k := kMax * math.Sqrt(float64(i)/float64(b.circleN))
		noisiness := b.density * math.Pow(float64(i)/float64(b.circleN), 2)

		for theta := 0.0; theta < 361; theta += 1.0 {
			r1 := math.Cos(gg.Radians(theta)) + 1
			r2 := math.Sin(gg.Radians(theta)) + 1
			r := radius + noise.Noise3D(k*r1, k*r2, float64(i)*b.circleGap)*noisiness

			x := float64(c.Width())/2 + r*math.Cos(gg.Radians(theta))
			y := float64(c.Height()/2) + r*math.Sin(gg.Radians(theta))
			ctex.LineTo(x, y)
		}
		ctex.Stroke()
		ctex.ClearPath()
	}
}
