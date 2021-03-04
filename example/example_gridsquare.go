package main

import (
	"github.com/jdxyw/generativeart"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600)
	c.SetBackground(generativeart.DarkPink[rand.Intn(5)])
	c.SetColorSchema(generativeart.DarkPink)
	c.Draw(generativeart.NewGirdSquares(24, 10, 0.2))
	c.ToPNG("gsquare.png")
}
