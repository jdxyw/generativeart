package main

import (
	"generativeart"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600, 1, 1)
	c.SetBackground(generativeart.Viridis[rand.Intn(255)])
	c.SetColorSchema(generativeart.Viridis)
	c.Draw(generativeart.NewGirdSquares())
	c.ToPNG("gsquare.png")
}
