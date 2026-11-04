package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20262204)
	c.SetBackground(color.RGBA{R: 244, G: 234, B: 210, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewMineralShanShui(5, 4, 1.0))
	c.ToPNG("mineralshanshui.png")
}
