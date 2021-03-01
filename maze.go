package generativeart

import (
	"image/color"
	"math/rand"

	"github.com/fogleman/gg"
)

type maze struct{}

// NewMaze returns a maze generator.
func NewMaze() *maze {
	return &maze{}
}

// Generative draws a random maze image.
func (m *maze) Generative(c *canva, rgba color.RGBA) {
	ctex := gg.NewContextForRGBA(c.img)
	ctex.SetColor(rgba)
	ctex.SetLineWidth(3.0)
	step := 20
	for x := 0; x < c.width; x += step {
		for y := 0; y < c.height; y += step {
			if rand.Float32() > 0.5 {
				ctex.DrawLine(float64(x), float64(y), float64(x+step), float64(y+step))

				//ctex.Stroke()
			} else {
				ctex.DrawLine(float64(x+step), float64(y), float64(x), float64(y+step))
				ctex.SetLineWidth(3.0)
				//ctex.Stroke()
			}
			ctex.Stroke()
		}
	}
}
