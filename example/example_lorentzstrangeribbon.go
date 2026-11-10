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
	c.SetSeed(20262603)
	c.SetBackground(color.RGBA{R: 6, G: 9, B: 19, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewLorentzStrangeRibbon(26, 28.0, 2.667))
	c.ToPNG("lorentzstrangeribbon.png")
}
