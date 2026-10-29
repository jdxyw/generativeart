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
	c.SetBackground(color.RGBA{R: 238, G: 235, B: 226, A: 255})
	c.FillBackground()
	c.Draw(arts.NewBotanicalWhorls(64, 22000, 137.507764))
	c.ToPNG("botanicalwhorls.png")
}
