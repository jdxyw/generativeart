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
	c.SetSeed(20262501)
	c.SetBackground(color.RGBA{R: 6, G: 9, B: 19, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewHopfTorusSilk(6, 22, 0.72))
	c.ToPNG("hopftorussilk.png")
}
