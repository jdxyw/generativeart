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
	c.SetBackground(common.Lavender)
	c.SetLineWidth(2)
	c.SetIterations(150000)
	c.FillBackground()
	c.Draw(arts.NewPointRibbon(50))
	c.ToPNG("pointribbon.png")
}
