package arts

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"math"
	"math/rand"
)

type circleMove struct {
	circleNum int
}

func NewCircleMove(circleNum int) *circleMove {
	return &circleMove{
		circleNum: circleNum,
	}
}

// Generative draws a sircle moving images.
func (cm *circleMove) Generative(c *generativeart.Canva) {
	ctex := gg.NewContextForRGBA(c.Img())
	ctex.SetLineWidth(0.3)
	noise := common.NewPerlinNoise()
	cl := rand.Intn(255)
	for i := 0; i < cm.circleNum; i++ {
		//var sx, sy float64
		var cxx float64
		var np = 300.0
		for j := 0.0; j < np; j += 1.0 {
			theta := common.Remap(j, 0, np, 0, math.Pi*2)
			cx := float64(i)*3 - 200.0
			cy := float64(c.Height())/2 + math.Sin(float64(i)/50)*float64(c.Height())/12.0
			xx := math.Cos(theta+cx/10) * float64(c.Height()) / 6.0
			yy := math.Sin(theta+cx/10) * float64(c.Height()) / 6.0
			p := common.NewVector(xx, yy)
			xx = (xx + cx) / 150
			yy = (yy + cy) / 150
			p.Multiple(1 + 1.5*noise.Noise2D(xx, yy))
			ctex.LineTo(cx+p.X, cy+p.Y)
			cxx = cx
		}
		hue := int(cxx/4) - cl
		if hue < 0 {
			hue += 255
		}

		h := common.HSV{
			H: hue,
			S: 180,
			V: 120,
		}

		rgba := h.ToRGB(255, 255, 255)
		rgba.A = 255
		ctex.SetColor(rgba)
		ctex.Stroke()
		ctex.ClosePath()
	}
}
