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
	c.SetBackground(color.RGBA{R: 28, G: 34, B: 38, A: 255})
	c.FillBackground()
	c.Draw(arts.NewSqueegeeScrape(8, 460, 1.0))
	c.ToPNG("squeegeescrape.png")
}
