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
	c.SetBackground(color.RGBA{R: 4, G: 7, B: 16, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewThomasAttractor(900000, 0.19))
	c.ToPNG("thomasattractor.png")
}
