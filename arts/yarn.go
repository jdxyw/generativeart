package arts

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
)

type yarn struct {
	n int
}

// NewYarn returns a yarn object.
func NewYarn(n int) *yarn {
	return &yarn{
		n: n,
	}
}

// Generative draws a yarn image.
func (y *yarn) Generative(c *generativeart.Canva) {
	ctex := gg.NewContextForRGBA(c.Img())
	ctex.SetLineWidth(c.Opts().LineWidth())
	ctex.SetColor(c.Opts().LineColor())
	noise := common.NewPerlinNoise()

	var offset = 0.0
	var inc = 0.005
	for i := 0; i < y.n; i++ {
		x0 := float64(c.Width()) * noise.Noise1D(offset+15)
		x1 := float64(c.Width()) * noise.Noise1D(offset+25)
		x2 := float64(c.Width()) * noise.Noise1D(offset+35)
		x3 := float64(c.Width()) * noise.Noise1D(offset+45)
		y0 := float64(c.Height()) * noise.Noise1D(offset+55)
		y1 := float64(c.Height()) * noise.Noise1D(offset+65)
		y2 := float64(c.Height()) * noise.Noise1D(offset+75)
		y3 := float64(c.Height()) * noise.Noise1D(offset+85)
		ctex.MoveTo(x0, y0)
		ctex.CubicTo(x1, y1, x2, y2, x3, y3)
		ctex.Stroke()
		ctex.ClearPath()
		offset += inc
	}
}
