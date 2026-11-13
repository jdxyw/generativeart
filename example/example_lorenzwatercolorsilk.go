package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262804)
	c.SetBackground(color.RGBA{R: 10, G: 18, B: 32, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewLorenzWatercolorSilk(8, 20000, 395.0))
	c.ToPNG("lorenzwatercolorsilk.png")
}
