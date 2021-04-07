package main

import (
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(common.White)
	c.SetAlpha(120)
	c.SetLineWidth(0.3)
	c.FillBackground()
	c.SetIterations(200)
	c.Draw(arts.NewPerlinPerls(10, 200, 40, 80))
	c.ToPNG("perlinperls.png")
}
