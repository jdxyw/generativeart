package generativeart

import (
	"math/rand"

	"github.com/fogleman/gg"
)

type maze struct {
	step int
}

// NewMaze returns a maze generator.
func NewMaze(step int) *maze {
	return &maze{
		step: step,
	}
}

// Generative draws a random maze image.
func (m *maze) Generative(c *canva) {
	ctex := gg.NewContextForRGBA(c.img)
	ctex.SetColor(c.opts.lineColor)
	ctex.SetLineWidth(c.opts.lineWidth)
	for x := 0; x < c.width; x += m.step {
		for y := 0; y < c.height; y += m.step {
			if rand.Float32() > 0.5 {
				ctex.DrawLine(float64(x), float64(y), float64(x+m.step), float64(y+m.step))
			} else {
				ctex.DrawLine(float64(x+m.step), float64(y), float64(x), float64(y+m.step))
			}
			ctex.Stroke()
		}
	}
}
