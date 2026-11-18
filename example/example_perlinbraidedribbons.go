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
	c.SetSeed(20263103)
	c.SetBackground(color.RGBA{R: 12, G: 10, B: 24, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewPerlinBraidedRibbons(28, 56, 1.35))
	c.ToPNG("perlinbraidedribbons.png")
}
