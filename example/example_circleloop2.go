package main

import (
	"github.com/jdxyw/generativeart"
	"image/color"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	colors := []color.RGBA{
		{0xF9, 0xC8, 0x0E, 0xFF},
		{0xF8, 0x66, 0x24, 0xFF},
		{0xEA, 0x35, 0x46, 0xFF},
		{0x66, 0x2E, 0x9B, 0xFF},
		{0x43, 0xBC, 0xCD, 0xFF},
	}
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(color.RGBA{8, 10, 20, 255})
	c.FillBackground()
	c.SetColorSchema(colors)
	c.Draw(generativeart.NewCircleLoop2(7))
	c.ToPNG("colorloop2.png")
}
