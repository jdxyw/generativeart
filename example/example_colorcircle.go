package main

import (
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
	"image/color"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	colors := []color.RGBA{
		{0xFF, 0xC6, 0x18, 0xFF},
		{0xF4, 0x25, 0x39, 0xFF},
		{0x41, 0x78, 0xF4, 0xFF},
		{0xFE, 0x84, 0xFE, 0xFF},
		{0xFF, 0x81, 0x19, 0xFF},
		{0x56, 0xAC, 0x51, 0xFF},
		{0x98, 0x19, 0xFA, 0xFF},
		{0xFF, 0xFF, 0xFF, 0xFF},
	}
	c := generativeart.NewCanva(1000, 1000)
	c.SetBackground(common.White)
	c.FillBackground()
	c.SetColorSchema(colors)
	c.Draw(arts.NewColorCircle(500))
	c.ToPNG("colorcircle.png")
}
