package main

import (
	"github.com/jdxyw/generativeart"
	"github.com/jdxyw/generativeart/arts"
	"github.com/jdxyw/generativeart/common"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(common.DarkPink[rand.Intn(5)])
	c.SetColorSchema(common.DarkPink)
	c.Draw(arts.NewGirdSquares(24, 10, 0.2))
	c.ToPNG("gsquare.png")
}
