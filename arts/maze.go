package arts

import (
	"github.com/jdxyw/generativeart"

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
func (m *maze) Generative(c *generativeart.Canva) {
	ctex := gg.NewContextForRGBA(c.Img())
	ctex.SetColor(c.Opts().LineColor())
	ctex.SetLineWidth(c.Opts().LineWidth())
	rng := c.Rand()
	for x := 0; x < c.Width(); x += m.step {
		for y := 0; y < c.Height(); y += m.step {
			if rng.Float32() > 0.5 {
				ctex.DrawLine(float64(x), float64(y), float64(x+m.step), float64(y+m.step))
			} else {
				ctex.DrawLine(float64(x+m.step), float64(y), float64(x), float64(y+m.step))
			}
			ctex.Stroke()
		}
	}
}
