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
	c.SetBackground(color.RGBA{R: 16, G: 24, B: 42, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Botanical)
	c.Draw(arts.NewMosaicLotus(12, 1.0))
	c.ToPNG("mosaiclotus.png")
}
