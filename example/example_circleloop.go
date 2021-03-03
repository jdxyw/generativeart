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
	c.SetLineWidth(1)
	c.SetLineColor(generativeart.Orange)
	c.SetAlpha(30)
	c.SetIterations(1000)
	c.FillBackground()
	c.Draw(generativeart.NewCircleLoop(100))
	c.ToPNG("circleloop.png")
}
