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
	c.SetSeed(20262601)
	c.SetBackground(color.RGBA{R: 7, G: 11, B: 22, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewFidenzaVelvetLoom(18, 5, 2.2))
	c.ToPNG("fidenzavelvetloom.png")
}
