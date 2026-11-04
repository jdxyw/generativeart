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
	c.SetBackground(color.RGBA{R: 250, G: 243, B: 227, A: 255})
	c.FillBackground()
	c.Draw(arts.NewEscherMetamorphosis(10, 10, 1.0))
	c.ToPNG("eschermetamorphosis.png")
}
