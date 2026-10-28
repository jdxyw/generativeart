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
	c.SetBackground(color.RGBA{R: 239, G: 232, B: 222, A: 255})
	c.FillBackground()
	c.Draw(arts.NewVortexImpasto(46, 22000, 1.0))
	c.ToPNG("vorteximpasto.png")
}
