package main

import (
	"image/color"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(20261023)
	c.SetBackground(color.RGBA{R: 6, G: 8, B: 16, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Cyberpunk)
	c.Draw(arts.NewSoapIridescence(42, 480.0))
	c.ToPNG("soapiridescence.png")
}
