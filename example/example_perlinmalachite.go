//go:build ignore

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
	c.SetBackground(color.RGBA{R: 7, G: 40, B: 30, A: 255})
	c.FillBackground()
	c.Draw(arts.NewPerlinMalachite(16, 5, 2.4))
	c.ToPNG("perlinmalachite.png")
}
