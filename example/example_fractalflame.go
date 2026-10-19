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
	c.SetBackground(color.RGBA{R: 4, G: 4, B: 10, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Magma)
	c.Draw(arts.NewFractalFlame(1800000, 5))
	c.ToPNG("fractalflame.png")
}
