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
	c.SetSeed(20263101)
	c.SetBackground(color.RGBA{R: 252, G: 224, B: 142, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewPerlinCanyonArch(18, 14, 1.35))
	c.ToPNG("perlincanyonarch.png")
}
