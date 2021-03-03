package generativeart

import (
	"math/rand"

	"github.com/fogleman/gg"
)

type maze struct{}

// NewMaze returns a maze generator.
func NewMaze() *maze {
	return &maze{}
}

// Generative draws a random maze image.
func (m *maze) Generative(c *canva) {
	ctex := gg.NewContextForRGBA(c.img)
	ctex.SetColor(c.opts.lineColor)
	ctex.SetLineWidth(c.opts.lineWidth)
	step := c.opts.step
	for x := 0; x < c.width; x += step {
		for y := 0; y < c.height; y += step {
			if rand.Float32() > 0.5 {
				ctex.DrawLine(float64(x), float64(y), float64(x+step), float64(y+step))
			} else {
				ctex.DrawLine(float64(x+step), float64(y), float64(x), float64(y+step))
			}
			ctex.Stroke()
		}
	}
}
