package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261027)
	c.SetBackground(color.RGBA{R: 6, G: 9, B: 22, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewWireframeTerrain(76, 1.0))
	c.ToPNG("wireframeterrain.png")
}
