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
	c.SetBackground(color.RGBA{R: 12, G: 15, B: 26, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Oceanic)
	c.Draw(arts.NewGuilloche(28, 175, 65, 115))
	c.ToPNG("guilloche.png")
}
