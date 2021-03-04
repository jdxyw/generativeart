package main

import (
	"github.com/jdxyw/generativeart"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	c := generativeart.NewCanva(600, 600, 2, 2)
	c.SetBackground(generativeart.Azure)
	c.SetLineWidth(3)
	c.SetLineColor(generativeart.Orange)
	c.FillBackground()
	c.Draw(generativeart.NewMaze())
	c.ToPNG("maze.png")
}
