//go:build ignore

package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262502)
	c.SetBackground(color.RGBA{R: 244, G: 236, B: 214, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Sunset)
	c.Draw(arts.NewVolcanicAtollTerraces(16, 4, 1.75))
	c.ToPNG("volcanicatollterraces.png")
}
