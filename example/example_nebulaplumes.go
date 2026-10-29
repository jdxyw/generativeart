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
	c.SetBackground(color.RGBA{R: 243, G: 236, B: 225, A: 255})
	c.FillBackground()
	c.Draw(arts.NewNebulaPlumes(3, 23000, 1.0))
	c.ToPNG("nebulaplumes.png")
}
