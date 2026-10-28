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
	c.SetBackground(color.RGBA{R: 7, G: 6, B: 5, A: 255})
	c.FillBackground()
	c.Draw(arts.NewHexMandala(13, 16, 1.0))
	c.ToPNG("hexmandala.png")
}
