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
	c.SetBackground(color.RGBA{R: 13, G: 16, B: 24, A: 255})
	c.FillBackground()
	c.Draw(arts.NewAbyssalCurrents(18, 23000, 1.0))
	c.ToPNG("abyssalcurrents.png")
}
