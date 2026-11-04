//go:build ignore

package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 243, G: 236, B: 224, A: 255})
	c.FillBackground()
	c.Draw(arts.NewEscherCircleLimit(6, 4, 2))
	c.ToPNG("eschercirclelimit.png")
}
