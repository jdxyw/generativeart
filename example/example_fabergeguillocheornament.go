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
	c.SetSeed(20262703)
	c.SetBackground(color.RGBA{R: 6, G: 8, B: 20, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewFabergeGuillocheOrnament(12, 28, 1.0))
	c.ToPNG("fabergeguillocheornament.png")
}
