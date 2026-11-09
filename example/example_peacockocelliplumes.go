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
	c.SetSeed(20262503)
	c.SetBackground(color.RGBA{R: 12, G: 20, B: 30, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewPeacockOcelliPlumes(7, 18000, 390.0))
	c.ToPNG("peacockocelliplumes.png")
}
