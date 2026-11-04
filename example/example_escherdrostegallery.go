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
	c.SetSeed(20261101)
	c.SetBackground(color.RGBA{R: 22, G: 32, B: 56, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewEscherDrosteGallery(1, 8, 16.0))
	c.ToPNG("escherdrostegallery.png")
}
