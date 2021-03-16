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
		{0xF9, 0xC8, 0x0E, 0xFF},
		{0xF8, 0x66, 0x24, 0xFF},
		{0xEA, 0x35, 0x46, 0xFF},
		{0x66, 0x2E, 0x9B, 0xFF},
		{0x43, 0xBC, 0xCD, 0xFF},
	}
	c := generativeart.NewCanva(800, 800)
	c.SetBackground(common.Black)
	c.FillBackground()
	c.SetColorSchema(colors)
	c.SetIterations(1200)
	c.Draw(arts.NewPixelHole(60))
	c.ToPNG("pixelhole.png")

}
