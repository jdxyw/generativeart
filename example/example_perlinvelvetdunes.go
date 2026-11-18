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
	c.SetSeed(20263013)
	c.SetBackground(color.RGBA{R: 10, G: 6, B: 14, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewPerlinVelvetDunes(24, 58, 1.35))
	c.ToPNG("perlinvelvetdunes.png")
}
