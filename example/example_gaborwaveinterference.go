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
	c.SetBackground(color.RGBA{R: 6, G: 8, B: 20, A: 255})
	c.FillBackground()
	c.Draw(arts.NewGaborWaveInterference(6, 3.2, 1.45))
	c.ToPNG("gaborwaveinterference.png")
}
