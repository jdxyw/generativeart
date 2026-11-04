package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262302)
	c.SetBackground(color.RGBA{R: 248, G: 241, B: 224, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewKeplerHarmonices(5, 3, 0.26))
	c.ToPNG("keplerharmonices.png")
}
