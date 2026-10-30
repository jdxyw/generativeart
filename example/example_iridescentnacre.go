package main

import (
	"image/color"
	"math/rand"
	"time"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
)

func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 14, G: 22, B: 30, A: 255})
	c.FillBackground()
	c.Draw(arts.NewIridescentNacre(7, 18000, 410.0))
	c.ToPNG("iridescentnacre.png")
}
