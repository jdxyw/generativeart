package main

import (
	"github.com/jdxyw/generativeart"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(generativeart.Tan)
	c.SetLineWidth(1.0)
	c.SetLineColor(generativeart.LightPink)
	c.FillBackground()
	c.Draw(generativeart.NewCircleLine(0.02, 600, 1.5, 2, 2))
	c.ToPNG("circleline.png")
}
