package main

import (
	"github.com/jdxyw/generativeart"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500, 2, 2)
	c.SetBackground(generativeart.Black)
	c.SetLineWidth(1.0)
	c.SetLineColor(generativeart.MediumAquamarine)
	c.SetAlpha(30)
	c.SetColorSchema(generativeart.Plasma)
	c.SetIterations(4)
	c.FillBackground()
	c.Draw(generativeart.NewSilkSmoke(400, 20, 0.2, 2, 10, 30, false))
	c.ToPNG("silksmoke.png")
}
