package main

import (
	"image/color"
	"log"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	c := generativeart.NewCanva(600, 600)
	c.SetSeed(2026102801)
	c.SetBackground(color.RGBA{R: 22, G: 27, B: 34, A: 255})
	c.SetColorSchema(common.Dunhuang)
	c.FillBackground()
	c.Draw(arts.NewConcentricMorph(18, 18, 1.0))
	if err := c.ToPNG("concentricmorph.png"); err != nil {
		log.Fatal(err)
	}
}
