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
	c.SetSeed(20262803)
	c.SetBackground(color.RGBA{R: 7, G: 9, B: 16, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewPenroseWavefrontLoom(5, 12, 410.0))
	c.ToPNG("penrosewavefrontloom.png")
}
