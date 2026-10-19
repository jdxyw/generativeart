package main

import (
	"image/color"
	"math/rand"
	"time"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
)

func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 247, G: 243, B: 234, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Dunhuang)
	c.SetLineColor(color.RGBA{R: 32, G: 28, B: 38, A: 255})
	c.Draw(arts.NewWatercolor(9, 22, 5, 0.32))
	c.ToPNG("watercolor.png")
}
