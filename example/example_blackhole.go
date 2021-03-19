package main

import (
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
	"image/color"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(color.RGBA{R: 30, G: 30, B: 30, A: 255})
	c.FillBackground()
	c.SetLineWidth(1.0)
	c.SetLineColor(common.Tomato)
	c.Draw(arts.NewBlackHole(200, 400, 0.01))
	c.ToPNG("blackhole.png")
}
