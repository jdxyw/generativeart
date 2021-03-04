package main

import (
	"github.com/jdxyw/generativeart"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(generativeart.Lavender)
	c.SetLineWidth(2)
	c.SetIterations(150000)
	c.FillBackground()
	c.Draw(generativeart.NewPointRibbon(50))
	c.ToPNG("pointribbon.png")
}
