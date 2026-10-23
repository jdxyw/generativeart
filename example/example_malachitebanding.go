package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261023)
	c.SetBackground(color.RGBA{R: 5, G: 28, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Botanical)
	c.Draw(arts.NewMalachiteBanding(14, 22))
	c.ToPNG("malachitebanding.png")
}
