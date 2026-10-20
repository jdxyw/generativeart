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
	c.SetBackground(color.RGBA{R: 72, G: 178, B: 102, A: 255})
	c.FillBackground()
	c.SetColorSchema(common.Youthful)
	c.Draw(arts.NewOrphicDiscs(34))
	c.ToPNG("orphicdiscs.png")
}
