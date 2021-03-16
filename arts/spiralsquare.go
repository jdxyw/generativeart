package arts

import (
	"github.com/fogleman/gg"
	"github.com/jdxyw/generativeart"
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
func (s *spiralSquare) Generative(c *generativeart.Canva) {
	ctex := gg.NewContextForRGBA(c.Img())

	sl := s.rectSide
	theta := rand.Intn(360) + 1
	for i := 0; i < s.squareNum; i++ {
		ctex.Push()
		ctex.Translate(float64(c.Width()/2), float64(c.Height()/2))
		ctex.Rotate(gg.Radians(float64(theta * (i + 1))))

		ctex.Scale(sl, sl)

		ctex.LineTo(-0.5, 0.5)
		ctex.LineTo(0.5, 0.5)
		ctex.LineTo(0.5, -0.5)
		ctex.LineTo(-0.5, -0.5)
		ctex.LineTo(-0.5, 0.5)

		ctex.SetLineWidth(c.Opts().LineWidth())
		ctex.SetColor(c.Opts().LineColor())
		ctex.StrokePreserve()

		if s.randColor {
			ctex.SetColor(c.Opts().ColorSchema()[rand.Intn(len(c.Opts().ColorSchema()))])
		} else {
			ctex.SetColor(c.Opts().Foreground())
		}
		ctex.Fill()
		ctex.Pop()
		sl = sl - s.decay*s.rectSide

		if sl < 0 {
			return
		}
	}
}
