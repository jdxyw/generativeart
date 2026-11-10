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
	c.SetSeed(20262701)
	c.SetBackground(color.RGBA{R: 7, G: 9, B: 18, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewAgateGeodeCrystal(20, 3, 1.0))
	c.ToPNG("agategeodecrystal.png")
}
