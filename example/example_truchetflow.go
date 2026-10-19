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
	c.SetBackground(color.RGBA{R: 245, G: 240, B: 230, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.UkiyoE)
	c.Draw(arts.NewTruchetFlow(6, 3))
	c.ToPNG("truchetflow.png")
}
