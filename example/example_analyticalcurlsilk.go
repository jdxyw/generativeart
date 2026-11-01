package main

import (
	"image/color"
	"math/rand"
	"time"

	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
)

func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(color.RGBA{R: 10, G: 18, B: 40, A: 255})
	c.FillBackground()
	c.Draw(arts.NewAnalyticalCurlSilk(420, 4, 2.2))
	c.ToPNG("analyticalcurlsilk.png")
}
