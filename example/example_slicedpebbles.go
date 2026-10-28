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
	c.SetSeed(2026102803)
	c.SetBackground(color.RGBA{R: 15, G: 18, B: 24, A: 255})
	c.SetColorSchema(common.Sunset)
	c.FillBackground()
	c.Draw(arts.NewSlicedPebbles(6, 3, 18.0))
	if err := c.ToPNG("slicedpebbles.png"); err != nil {
		log.Fatal(err)
	}
}
