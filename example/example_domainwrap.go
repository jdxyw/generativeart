package main

import (
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
	"image/color"
	"math/rand"
	"time"
)

func cmap(r, m1, m2 float64) color.RGBA {
	rgb := color.RGBA{
		R: uint8(common.Constrain(m1*200*r, 0, 255)),
		G: uint8(common.Constrain(r*200, 0, 255)),
		B: uint8(common.Constrain(m2*255*r, 70, 255)),
		A: 255,
	}
	return rgb
}

func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(500, 500)
	c.SetBackground(common.Black)
	c.FillBackground()
	c.Draw(arts.NewDomainWrap(0.01, 4,4, 20, cmap))
	c.ToPNG("domainwarp.png")
}
