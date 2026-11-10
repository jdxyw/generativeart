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
	c.SetSeed(20262604)
	c.SetBackground(color.RGBA{R: 5, G: 7, B: 16, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Inferno)
	c.Draw(arts.NewCoronalMagneticArcade(9, 19000, 1.15))
	c.ToPNG("coronalmagneticarcade.png")
}
