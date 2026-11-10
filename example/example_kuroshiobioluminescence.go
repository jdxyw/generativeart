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
	c.SetSeed(20262602)
	c.SetBackground(color.RGBA{R: 4, G: 10, B: 28, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewKuroshioBioluminescence(14, 18, 24.0))
	c.ToPNG("kuroshiobioluminescence.png")
}
