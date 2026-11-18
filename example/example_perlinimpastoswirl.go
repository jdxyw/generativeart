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
	c.SetSeed(20263011)
	c.SetBackground(color.RGBA{R: 18, G: 8, B: 14, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewPerlinImpastoSwirl(28, 58, 1.35))
	c.ToPNG("perlinimpastoswirl.png")
}
