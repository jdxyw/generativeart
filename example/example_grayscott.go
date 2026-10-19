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
	c.SetBackground(color.RGBA{R: 8, G: 12, B: 20, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Viridis)
	c.Draw(arts.NewGrayScott(3200, 0.037, 0.060))
	c.ToPNG("grayscott.png")
}
