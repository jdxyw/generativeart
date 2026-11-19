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
	c.SetSeed(20263303)
	c.SetBackground(color.RGBA{R: 12, G: 18, B: 42, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.Draw(arts.NewNautilusChamberScroll(14, 4, 1.0))
	c.ToPNG("nautiluschamberscroll.png")
}
