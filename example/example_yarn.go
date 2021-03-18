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
	c.SetBackground(common.Orange)
	c.FillBackground()
	c.SetLineWidth(0.3)
	c.SetLineColor(color.RGBA{A: 60})
	c.Draw(arts.NewYarn(2000))
	c.ToPNG("yarn.png")
}
