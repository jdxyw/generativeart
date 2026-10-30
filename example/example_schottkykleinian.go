package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261030)
	c.SetBackground(color.RGBA{R: 8, G: 12, B: 26, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewSchottkyKleinian(7, 1.0))
	c.ToPNG("schottkykleinian.png")
}
