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
	c.SetBackground(color.RGBA{R: 12, G: 14, B: 20, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewKintsugiPorcelain(96, 5))
	c.ToPNG("kintsugiporcelain.png")
}
