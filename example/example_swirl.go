package main

import (
	"github.com/jdxyw/generativeart"
	"image/color"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(1600, 1600)
	c.SetBackground(generativeart.Azure)
	c.FillBackground()
	c.SetForeground(color.RGBA{113, 3, 0, 140})
	c.SetIterations(4000000)
	c.Draw(generativeart.NewSwirl(0.970, -1.899, 1.381, -1.506, 2.4, 2.4))
	c.ToPNG("swirl.png")
}
