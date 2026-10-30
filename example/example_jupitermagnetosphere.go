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
	c.SetBackground(color.RGBA{R: 7, G: 19, B: 44, A: 255})
	c.FillBackground()
	c.Draw(arts.NewJupiterMagnetosphere(9, 16, 22.0))
	c.ToPNG("jupitermagnetosphere.png")
}
