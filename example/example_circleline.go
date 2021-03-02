package main

import (
	"generativeart"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(800, 800, 2, 2)
	c.SetBackground(generativeart.Tan)
	c.SetLineWidth(1.0)
	c.SetLineColor(generativeart.Lavender)
	c.FillBackground()
	c.Draw(generativeart.NewCircleLine(0.02, 600, 1.5))
	c.ToPNG("circleline.png")
}
