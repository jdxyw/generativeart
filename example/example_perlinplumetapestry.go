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
	c.SetSeed(20263012)
	c.SetBackground(color.RGBA{R: 6, G: 10, B: 22, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewPerlinPlumeTapestry(22, 12, 1.35))
	c.ToPNG("perlinplumetapestry.png")
}
