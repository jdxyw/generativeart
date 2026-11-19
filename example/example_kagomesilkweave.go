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
	c.SetSeed(20263302)
	c.SetBackground(color.RGBA{R: 10, G: 18, B: 46, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewKagomeSilkWeave(7, 48, 1.25))
	c.ToPNG("kagomesilkweave.png")
}
