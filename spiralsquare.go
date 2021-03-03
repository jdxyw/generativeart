package generativeart

import (
	"github.com/fogleman/gg"
	"math/rand"
)

type spiralSquare struct {
	squareNum int
	decay     float64
	rectSide  float64
	randColor bool
}

// NewSpiralSquare returns a spiralSquare object.
func NewSpiralSquare(squareNum int, rectSide, decay float64, randColor bool) *spiralSquare {
	return &spiralSquare{
		squareNum: squareNum,
		decay:     decay,
		rectSide:  rectSide,
		randColor: randColor,
	}
}

// Generative draws a spiral square images.
func (s *spiralSquare) Generative(c *canva) {
	ctex := gg.NewContextForRGBA(c.img)

	sl := s.rectSide
	theta := rand.Intn(360) + 1
	for i := 0; i < s.squareNum; i++ {
		ctex.Push()
		ctex.Translate(float64(c.width/2), float64(c.height/2))
		ctex.Rotate(gg.Radians(float64(theta * (i + 1))))

		ctex.Scale(sl, sl)

		ctex.LineTo(-0.5, 0.5)
		ctex.LineTo(0.5, 0.5)
		ctex.LineTo(0.5, -0.5)
		ctex.LineTo(-0.5, -0.5)
		ctex.LineTo(-0.5, 0.5)

		ctex.SetLineWidth(c.opts.lineWidth)
		ctex.SetColor(c.opts.lineColor)
		ctex.StrokePreserve()

		if s.randColor {
			ctex.SetColor(c.opts.colorSchema[rand.Intn(len(c.opts.colorSchema))])
		} else {
			ctex.SetColor(c.opts.foreground)
		}
		ctex.Fill()
		ctex.Pop()
		sl = sl - s.decay*s.rectSide

		if sl < 0 {
			return
		}
	}
}
