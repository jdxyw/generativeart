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
	c.SetBackground(color.RGBA{R: 244, G: 247, B: 244, A: 255})
	c.FillBackground()
	c.Draw(arts.NewHalftonePanels(22, 5.5))
	c.ToPNG("halftonepanels.png")
}
