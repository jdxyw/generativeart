package main

import (
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/common"
	"image/color"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(common.Black)
	c.FillBackground()
	c.SetLineColor(color.RGBA{255, 64, 8, 128})
	c.Draw(generativeart.NewSolarFlare())
	c.ToPNG("solarflare.png")
}
