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
	c.SetBackground(color.RGBA{R: 6, G: 6, B: 14, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Inferno)
	c.Draw(arts.NewCliffordNebula(1500000, -1.7, 1.8, -1.9, -0.4))
	c.ToPNG("cliffordnebula.png")
}
