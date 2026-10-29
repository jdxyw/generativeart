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
	c.SetBackground(color.RGBA{R: 242, G: 235, B: 224, A: 255})
	c.FillBackground()
	c.Draw(arts.NewSolarProminence(14, 24000, 1.0))
	c.ToPNG("solarprominence.png")
}
